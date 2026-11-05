package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"scheduler/internal/agents"
	"scheduler/internal/planner"
	"scheduler/internal/store"
	"slices"
	"strings"
	"sync"
	"syscall"
	"time"
	_ "time/tzdata"
)

type app struct {
	db  *store.Store
	ai  *agents.Client
	loc *time.Location
	mu  sync.Mutex
}

func env(key, fallback string) string {
	if x := os.Getenv(key); x != "" {
		return x
	}
	return fallback
}
func main() {
	reset := flag.Bool("reset-demo", false, "replace local data with synthetic demo data")
	flag.Parse()
	loc, err := time.LoadLocation(env("SHIP_TIMEZONE", "America/New_York"))
	if err != nil {
		log.Fatal(err)
	}
	db, err := store.Open(env("DATABASE_PATH", "./data/scheduler.db"), loc, *reset)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if *reset {
		fmt.Println("Demo restored. Local planning data was replaced.")
		return
	}
	if err = db.RecoverDocumentAnalyses(context.Background()); err != nil {
		log.Fatal(err)
	}
	a := &app{db: db, ai: agents.New(os.Getenv("OPENAI_API_KEY"), env("OPENAI_MODEL", "gpt-5-mini")), loc: loc}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/context", a.context)
	mux.HandleFunc("PUT /api/data/{kind}/{id}", a.edit)
	mux.HandleFunc("DELETE /api/data/{kind}/{id}", a.edit)
	mux.HandleFunc("POST /api/validate", a.validate)
	mux.HandleFunc("POST /api/runs", a.startRun)
	mux.HandleFunc("GET /api/runs/{id}", a.getRun)
	mux.HandleFunc("POST /api/runs/{id}/apply", a.apply)
	mux.HandleFunc("POST /api/constraints/interpret", a.interpret)
	mux.HandleFunc("POST /api/events/{id}/complete", a.complete)
	mux.HandleFunc("GET /api/documents", a.listDocuments)
	mux.HandleFunc("POST /api/documents", a.importDocument)
	mux.HandleFunc("GET /api/documents/{id}", a.getDocument)
	mux.HandleFunc("GET /api/documents/{id}/file", a.documentFile)
	mux.HandleFunc("POST /api/documents/{id}/analyze", a.startDocumentAnalysis)
	mux.HandleFunc("POST /api/documents/{id}/citations/check", a.checkDocumentCitations)
	mux.HandleFunc("PUT /api/documents/{id}/suggestions/{suggestion}", a.reviewSuggestion)
	mux.HandleFunc("DELETE /api/documents/{id}/suggestions/{suggestion}", a.reviewSuggestion)
	mux.HandleFunc("POST /api/documents/{id}/suggestions/{suggestion}/create", a.reviewSuggestion)
	files := http.FileServer(http.Dir("web/dist"))
	mux.Handle("/", files)
	srv := &http.Server{Addr: net.JoinHostPort(env("LISTEN_HOST", "127.0.0.1"), env("PORT", "8080")), Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	listener, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nPB4T planner ready → http://localhost:%s\nDatabase: %s\n", env("PORT", "8080"), env("DATABASE_PATH", "./data/scheduler.db"))
	if a.ai.Key == "" {
		fmt.Println("Core planning enabled. Live AI unavailable until OPENAI_API_KEY is configured.")
	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stop
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()
	if err = srv.Serve(listener); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
func respond(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func fail(w http.ResponseWriter, status int, err error) {
	respond(w, status, map[string]string{"error": err.Error()})
}
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 2<<20)
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		fail(w, 400, err)
		return false
	}
	return true
}
func (a *app) week(raw string) (time.Time, error) {
	if raw == "" {
		return planner.Week(time.Now(), a.loc), nil
	}
	t, err := time.ParseInLocation("2006-01-02", raw, a.loc)
	if err != nil {
		t, err = time.Parse(time.RFC3339, raw)
	}
	return planner.Week(t, a.loc), err
}
func (a *app) context(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	week, err := a.week(r.URL.Query().Get("week"))
	if err != nil {
		fail(w, 400, err)
		return
	}
	d, err := a.db.Load(r.Context())
	if err != nil {
		fail(w, 500, err)
		return
	}
	runs, err := a.db.Proposals(r.Context())
	if err != nil {
		fail(w, 500, err)
		return
	}
	summaries := []map[string]any{}
	for _, p := range runs {
		if p.Week.Equal(week) {
			summaries = append(summaries, map[string]any{"id": p.ID, "status": p.Status, "mode": p.Mode, "summary": p.Summary})
		}
	}
	respond(w, 200, map[string]any{"data": d, "week": week, "timezone": a.loc.String(), "ai_available": a.ai.Key != "", "evaluation": planner.Evaluate(planner.Relevant(d, week), planner.WeekEvents(d, week)), "runs": summaries})
}
func (a *app) edit(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	d, err := a.db.Load(r.Context())
	if err != nil {
		fail(w, 500, err)
		return
	}
	var raw json.RawMessage
	remove := r.Method == "DELETE"
	kind, id := r.PathValue("kind"), r.PathValue("id")
	if !remove {
		if !decode(w, r, &raw) {
			return
		}
		var key struct {
			ID string `json:"id"`
		}
		if json.Unmarshal(raw, &key) != nil || key.ID != id {
			fail(w, 400, fmt.Errorf("record ID must match URL"))
			return
		}
	}
	old := slices.Clone(d.Events)
	if err = planner.Edit(&d, kind, id, raw, remove); err != nil {
		fail(w, 400, err)
		return
	}
	if kind == "events" && !remove {
		var e planner.Event
		_ = json.Unmarshal(raw, &e)
		for _, x := range old {
			if x.ID == e.ID && x.Fixed && (!e.Start.Equal(x.Start) || e.Duration != x.Duration) {
				fail(w, 400, fmt.Errorf("unlock the event before changing its fixed placement"))
				return
			}
		}
		if e.Status == "scheduled" {
			ev := planner.Evaluate(planner.Relevant(d, planner.Week(e.Start, a.loc)), planner.WeekEvents(d, planner.Week(e.Start, a.loc)))
			if !ev.Feasible {
				respond(w, 409, map[string]any{"error": "Placement conflicts with the working plan. Resolve it or save as draft.", "evaluation": ev})
				return
			}
			d.Assignments = planner.MergeAssignments(d, planner.WeekEvents(d, planner.Week(e.Start, a.loc)), ev.Assignments)
		}
	}
	if err = a.db.Save(r.Context(), d); err != nil {
		fail(w, 500, err)
		return
	}
	respond(w, 200, map[string]bool{"saved": true})
}

type candidateRequest struct {
	Week   string          `json:"week"`
	Event  *planner.Event  `json:"event"`
	Events []planner.Event `json:"events"`
	Agents bool            `json:"agents"`
}

func candidate(d planner.Data, week time.Time, req candidateRequest) ([]planner.Event, error) {
	events := planner.WeekEvents(d, week)
	if req.Events != nil {
		events = req.Events
	}
	if req.Event != nil {
		if err := planner.ValidateEvent(*req.Event); err != nil {
			return nil, err
		}
		i := slices.IndexFunc(events, func(e planner.Event) bool { return e.ID == req.Event.ID })
		if i >= 0 {
			if events[i].Fixed && (!events[i].Start.Equal(req.Event.Start) || events[i].Duration != req.Event.Duration) {
				return nil, fmt.Errorf("fixed event cannot be moved")
			}
			events[i] = *req.Event
		} else {
			events = append(events, *req.Event)
		}
	}
	return events, nil
}
func (a *app) validate(w http.ResponseWriter, r *http.Request) {
	var req candidateRequest
	if !decode(w, r, &req) {
		return
	}
	d, err := a.db.Load(r.Context())
	if err != nil {
		fail(w, 500, err)
		return
	}
	week, err := a.week(req.Week)
	if err != nil {
		fail(w, 400, err)
		return
	}
	events, err := candidate(d, week, req)
	if err != nil {
		fail(w, 400, err)
		return
	}
	copyData := d
	copyData.Events = append(slices.Clone(d.Events), events...)
	if req.Event != nil {
		copyData.Events = slices.DeleteFunc(slices.Clone(d.Events), func(e planner.Event) bool { return e.ID == req.Event.ID })
		copyData.Events = append(copyData.Events, *req.Event)
		if err = planner.ValidateData(copyData); err != nil {
			fail(w, 400, err)
			return
		}
	}
	respond(w, 200, planner.Evaluate(planner.Relevant(d, week), events))
}
func (a *app) startRun(w http.ResponseWriter, r *http.Request) {
	var req candidateRequest
	if !decode(w, r, &req) {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	d, err := a.db.Load(r.Context())
	if err != nil {
		fail(w, 500, err)
		return
	}
	week, err := a.week(req.Week)
	if err != nil {
		fail(w, 400, err)
		return
	}
	if _, err = candidate(d, week, req); err != nil {
		fail(w, 400, err)
		return
	}
	if req.Event != nil {
		copyData := d
		copyData.Events = slices.DeleteFunc(slices.Clone(d.Events), func(e planner.Event) bool { return e.ID == req.Event.ID })
		copyData.Events = append(copyData.Events, *req.Event)
		if err = planner.ValidateData(copyData); err != nil {
			fail(w, 400, err)
			return
		}
	}
	if req.Agents && a.ai.Key == "" {
		fail(w, 400, fmt.Errorf("configure an OpenAI key to enable live negotiation"))
		return
	}
	mode := "deterministic"
	if req.Agents {
		mode = "live-agents"
	}
	p := planner.Proposal{ID: planner.ID(), Week: week, Status: "running", Progress: "Preparing shared planning context", Snapshot: d, Mode: mode, Negotiation: []planner.DepartmentProposal{}}
	if err = a.db.SaveProposal(r.Context(), p); err != nil {
		fail(w, 500, err)
		return
	}
	go a.resolve(p, req)
	respond(w, 202, p)
}
func (a *app) resolve(p planner.Proposal, req candidateRequest) {
	started := time.Now()
	log.Printf("resolution run=%s mode=%s phase=start", p.ID, p.Mode)
	defer func() {
		log.Printf("resolution run=%s status=%s phase=finish duration=%s", p.ID, p.Status, time.Since(started).Round(time.Millisecond))
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	save := func() {
		persistCtx, persistCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer persistCancel()
		if err := a.db.SaveProposal(persistCtx, p); err != nil {
			log.Printf("save resolution: %v", err)
		}
	}
	d := planner.Relevant(p.Snapshot, p.Week)
	if req.Event != nil {
		events, _ := candidate(d, p.Week, req)
		outside := slices.DeleteFunc(slices.Clone(d.Events), func(e planner.Event) bool { return planner.Overlap(e.Start, e.End(), p.Week, p.Week.AddDate(0, 0, 7)) })
		d.Events = append(outside, events...)
	}
	weekEvents := planner.AddMissing(d, planner.WeekEvents(d, p.Week), p.Week)
	outside := slices.DeleteFunc(slices.Clone(d.Events), func(e planner.Event) bool { return planner.Overlap(e.Start, e.End(), p.Week, p.Week.AddDate(0, 0, 7)) })
	d.Events = append(outside, weekEvents...)
	prefs := []planner.Preference{}
	if req.Agents {
		negotiationData := d
		negotiationData.Events = slices.Clone(d.Events)
		if req.Event != nil {
			for i := range negotiationData.Events {
				if negotiationData.Events[i].ID == req.Event.ID {
					negotiationData.Events[i].Fixed = true
				}
			}
		}
		history, preferences, err := a.ai.Negotiate(ctx, p.ID, negotiationData, p.Week, func(stage string, history []planner.DepartmentProposal) {
			p.Progress = stage
			p.Negotiation = history
			save()
		})
		p.Negotiation = history
		if err != nil {
			p.Status = "failed"
			p.Error = err.Error()
			p.Progress = "Live negotiation stopped"
			save()
			return
		}
		prefs = preferences
	}

	p.Progress = "Searching placements and joint personnel assignments"
	save()
	solverStarted := time.Now()
	log.Printf("solver run=%s phase=start", p.ID)
	solved := planner.Solve(d, p.Week, req.Event, prefs, planner.WeekEvents(p.Snapshot, p.Week))
	log.Printf("solver run=%s phase=finish duration=%s status=%s", p.ID, time.Since(solverStarted).Round(time.Millisecond), solved.Status)
	solved.ID = p.ID
	solved.Snapshot = p.Snapshot
	solved.Mode = p.Mode
	solved.Negotiation = p.Negotiation
	solved.Progress = "Ready for review"
	solved.PreviousAssignments = p.Snapshot.Assignments
	p = solved
	save()
}
func (a *app) getRun(w http.ResponseWriter, r *http.Request) {
	p, err := a.db.Proposal(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, 404, err)
		return
	}
	respond(w, 200, p)
}
func (a *app) apply(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	p, err := a.db.Proposal(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, 404, err)
		return
	}
	if p.Status != "feasible" {
		fail(w, 400, fmt.Errorf("only a complete feasible proposal can be applied"))
		return
	}
	d, err := a.db.Load(r.Context())
	if err != nil {
		fail(w, 500, err)
		return
	}
	if !reflect.DeepEqual(d, p.Snapshot) {
		fail(w, 409, fmt.Errorf("planning inputs changed; generate a fresh proposal"))
		return
	}
	for _, old := range planner.WeekEvents(d, p.Week) {
		if old.Fixed && !slices.ContainsFunc(p.Events, func(e planner.Event) bool {
			return e.ID == old.ID && e.Start.Equal(old.Start) && e.Duration == old.Duration
		}) {
			fail(w, 409, fmt.Errorf("proposal changes a fixed event"))
			return
		}
	}
	evaluation := planner.Evaluate(planner.Relevant(d, p.Week), p.Events)
	if !evaluation.Feasible {
		respond(w, 409, map[string]any{"error": "proposal no longer validates", "evaluation": evaluation})
		return
	}
	for _, c := range evaluation.Coverage {
		if !c.Requirement.End.After(p.Week.AddDate(0, 0, 7)) && c.Gap > 0 {
			fail(w, 409, fmt.Errorf("required training remains uncovered"))
			return
		}
	}
	d.Events = slices.DeleteFunc(d.Events, func(e planner.Event) bool { return planner.Overlap(e.Start, e.End(), p.Week, p.Week.AddDate(0, 0, 7)) })
	d.Events = append(d.Events, p.Events...)
	d.Assignments = slices.DeleteFunc(d.Assignments, func(x planner.Assignment) bool {
		return slices.ContainsFunc(p.Events, func(e planner.Event) bool { return e.ID == x.EventID && e.Status != "completed" })
	})
	d.Assignments = append(d.Assignments, evaluation.Assignments...)
	if err = a.db.Save(r.Context(), d); err != nil {
		fail(w, 500, err)
		return
	}
	p.Status = "applied"
	p.Evaluation = evaluation
	if err = a.db.SaveProposal(r.Context(), p); err != nil {
		fail(w, 500, err)
		return
	}
	respond(w, 200, p)
}
func (a *app) interpret(w http.ResponseWriter, r *http.Request) {
	if a.ai.Key == "" {
		fail(w, 400, fmt.Errorf("live constraint assistance requires an OpenAI key"))
		return
	}
	var req struct {
		Text  string         `json:"text"`
		Event *planner.Event `json:"event"`
	}
	if !decode(w, r, &req) {
		return
	}
	if strings.TrimSpace(req.Text) == "" {
		fail(w, 400, fmt.Errorf("enter a constraint description"))
		return
	}
	d, err := a.db.Load(r.Context())
	if err != nil {
		fail(w, 500, err)
		return
	}
	drafts, err := a.ai.Constraints(r.Context(), d, req.Text, req.Event)
	if err != nil {
		fail(w, 502, err)
		return
	}
	respond(w, 200, drafts)
}
func (a *app) complete(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	d, err := a.db.Load(r.Context())
	if err != nil {
		fail(w, 500, err)
		return
	}
	i := slices.IndexFunc(d.Events, func(e planner.Event) bool { return e.ID == r.PathValue("id") })
	if i < 0 {
		fail(w, 404, fmt.Errorf("event not found"))
		return
	}
	e := &d.Events[i]
	if e.Status == "completed" {
		fail(w, 400, fmt.Errorf("already completed"))
		return
	}
	e.Status = "completed"
	d.Completions = append(d.Completions, planner.Completion{ID: planner.ID(), DepartmentID: e.DepartmentID, Type: e.Type, At: e.End(), EventID: e.ID})
	if err = a.db.Save(r.Context(), d); err != nil {
		fail(w, 500, err)
		return
	}
	respond(w, 200, map[string]bool{"completed": true})
}
