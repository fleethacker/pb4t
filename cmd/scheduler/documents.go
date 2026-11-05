package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"scheduler/internal/agents"
	"scheduler/internal/planner"
)

const maxDocumentSize = 20 << 20

func (a *app) listDocuments(w http.ResponseWriter, r *http.Request) {
	docs, err := a.db.Documents(r.Context())
	if err != nil {
		fail(w, 500, err)
		return
	}
	respond(w, 200, docs)
}
func (a *app) getDocument(w http.ResponseWriter, r *http.Request) {
	doc, err := a.db.Document(r.Context(), r.PathValue("id"))
	if err != nil {
		documentError(w, err)
		return
	}
	respond(w, 200, doc)
}
func documentError(w http.ResponseWriter, err error) {
	if errors.Is(err, sql.ErrNoRows) {
		fail(w, 404, fmt.Errorf("document not found"))
	} else {
		fail(w, 500, err)
	}
}
func (a *app) documentFile(w http.ResponseWriter, r *http.Request) {
	doc, err := a.db.Document(r.Context(), r.PathValue("id"))
	if err != nil {
		documentError(w, err)
		return
	}
	data, err := a.db.DocumentFile(r.Context(), doc.ID)
	if err != nil {
		documentError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Disposition", mime.FormatMediaType("inline", map[string]string{"filename": doc.Filename}))
	http.ServeContent(w, r, doc.Filename, doc.ImportedAt, bytes.NewReader(data))
}
func (a *app) importDocument(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxDocumentSize+(1<<20))
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		fail(w, 400, fmt.Errorf("upload a PDF no larger than 20 MB"))
		return
	}
	defer r.MultipartForm.RemoveAll()
	file, header, err := r.FormFile("file")
	if err != nil {
		fail(w, 400, fmt.Errorf("choose a PDF file"))
		return
	}
	defer file.Close()
	original, err := io.ReadAll(io.LimitReader(file, maxDocumentSize+1))
	if err != nil {
		fail(w, 400, err)
		return
	}
	if len(original) > maxDocumentSize || !bytes.HasPrefix(bytes.TrimSpace(original), []byte("%PDF-")) {
		fail(w, 400, fmt.Errorf("only PDF files up to 20 MB are supported"))
		return
	}
	hash := sha256.Sum256(original)
	name := filepath.Base(header.Filename)
	doc := planner.TrainingDocument{ID: planner.ID(), Filename: name, Title: strings.TrimSuffix(name, filepath.Ext(name)), Size: len(original), SHA256: hex.EncodeToString(hash[:]), ImportedAt: time.Now().UTC(), Status: "uploaded", Suggestions: []planner.TrainingSuggestion{}}
	if a.ai.Key != "" {
		doc.Status = "analyzing"
	}
	if err = a.db.AddDocument(r.Context(), doc, original); err != nil {
		fail(w, 500, err)
		return
	}
	if doc.Status == "analyzing" {
		go a.analyzeDocument(doc, original)
	}
	respond(w, 202, doc)
}
func (a *app) startDocumentAnalysis(w http.ResponseWriter, r *http.Request) {
	if a.ai.Key == "" {
		fail(w, 400, fmt.Errorf("configure an OpenAI key to generate document suggestions; the original PDF is already saved"))
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	doc, err := a.db.Document(r.Context(), r.PathValue("id"))
	if err != nil {
		documentError(w, err)
		return
	}
	if doc.Status == "analyzing" {
		fail(w, 409, fmt.Errorf("document analysis is already running"))
		return
	}
	if len(doc.Suggestions) > 0 {
		fail(w, 409, fmt.Errorf("this document already has saved suggestions; import another copy for a fresh analysis"))
		return
	}
	original, err := a.db.DocumentFile(r.Context(), doc.ID)
	if err != nil {
		documentError(w, err)
		return
	}
	doc.Status = "analyzing"
	doc.Error = ""
	if err = a.db.UpdateDocument(r.Context(), doc); err != nil {
		fail(w, 500, err)
		return
	}
	go a.analyzeDocument(doc, original)
	respond(w, 202, doc)
}
func (a *app) analyzeDocument(doc planner.TrainingDocument, original []byte) {
	started := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	log.Printf("document id=%s phase=analyze bytes=%d", doc.ID, len(original))
	d, err := a.db.Load(ctx)
	if err == nil {
		doc, err = a.ai.AnalyzeDocument(ctx, doc, original, d)
	}
	if err != nil {
		doc.Status = "failed"
		doc.Error = err.Error()
	} else {
		doc.Status = "ready"
		doc.Error = ""
	}
	persistCtx, persistCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer persistCancel()
	if saveErr := a.db.UpdateDocument(persistCtx, doc); saveErr != nil {
		log.Printf("document id=%s phase=persist failed=%v", doc.ID, saveErr)
	}
	log.Printf("document id=%s phase=finish status=%s suggestions=%d duration=%s", doc.ID, doc.Status, len(doc.Suggestions), time.Since(started).Round(time.Millisecond))
}
func (a *app) checkDocumentCitations(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	doc, err := a.db.Document(r.Context(), r.PathValue("id"))
	if err != nil {
		documentError(w, err)
		return
	}
	if doc.Status != "ready" {
		fail(w, 409, fmt.Errorf("finish document analysis before checking citations"))
		return
	}
	original, err := a.db.DocumentFile(r.Context(), doc.ID)
	if err != nil {
		documentError(w, err)
		return
	}
	agents.CheckDocumentSources(&doc, original)
	if err = a.db.UpdateDocument(r.Context(), doc); err != nil {
		fail(w, 500, err)
		return
	}
	respond(w, 200, doc)
}

func (a *app) reviewSuggestion(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	doc, err := a.db.Document(r.Context(), r.PathValue("id"))
	if err != nil {
		documentError(w, err)
		return
	}
	i := slices.IndexFunc(doc.Suggestions, func(s planner.TrainingSuggestion) bool { return s.ID == r.PathValue("suggestion") })
	if i < 0 {
		fail(w, 404, fmt.Errorf("suggestion not found"))
		return
	}
	original := doc.Suggestions[i]
	if r.Method == http.MethodDelete {
		if original.CreatedID != "" {
			fail(w, 409, fmt.Errorf("this suggestion was already created; manage the record in Planning inputs"))
			return
		}
		doc.Suggestions[i].Deleted = true
		if err = a.db.UpdateDocument(r.Context(), doc); err != nil {
			fail(w, 500, err)
			return
		}
		respond(w, 200, doc)
		return
	}
	if original.CreatedID != "" {
		respond(w, 200, doc)
		return
	}
	if original.Deleted && r.Method != http.MethodPut {
		fail(w, 409, fmt.Errorf("restore the deleted suggestion before creating it"))
		return
	}
	var s planner.TrainingSuggestion
	if !decode(w, r, &s) {
		return
	}
	// The original document stays fixed; the user can correct the AI citation.
	s.TrainingActivity = original.TrainingActivity
	s.DurationEvidence = original.DurationEvidence
	s.ID = original.ID
	s.Kind = original.Kind
	s.Basis = original.Basis
	s.CreatedID = ""
	// Replacing a draft also restores it when the user selects Undo.
	s.Deleted = false
	if strings.TrimSpace(s.Name) == "" {
		fail(w, 400, fmt.Errorf("enter a name"))
		return
	}
	if s.Page < 1 || strings.TrimSpace(s.Excerpt) == "" {
		fail(w, 400, fmt.Errorf("enter a source page and excerpt"))
		return
	}
	pdf, err := a.db.DocumentFile(r.Context(), doc.ID)
	if err != nil {
		documentError(w, err)
		return
	}
	checked := planner.TrainingDocument{Suggestions: []planner.TrainingSuggestion{s}}
	agents.CheckDocumentSources(&checked, pdf)
	s = checked.Suggestions[0]
	doc.Suggestions[i] = s
	if r.Method == "PUT" {
		if err = a.db.UpdateDocument(r.Context(), doc); err != nil {
			fail(w, 500, err)
			return
		}
		respond(w, 200, doc)
		return
	}
	d, err := a.db.Load(r.Context())
	if err != nil {
		fail(w, 500, err)
		return
	}
	source := &planner.SourceReference{DocumentID: doc.ID, Page: s.Page, Excerpt: s.Excerpt}
	createdID := planner.ID()
	if s.Kind == "template" {
		d.Templates = append(d.Templates, planner.EventTemplate{ID: createdID, Name: s.Name, Description: s.Description, DepartmentID: s.DepartmentID, Contact: s.Contact, Type: s.Type, Duration: s.Duration, Tags: s.Tags, ConstraintIDs: s.ConstraintIDs, Fatigue: s.Fatigue, SourceRef: source})
	} else if s.Kind == "constraint" {
		if s.Rule == nil {
			fail(w, 400, fmt.Errorf("review the structured constraint before creating it"))
			return
		}
		rule := *s.Rule
		rule.ID, rule.Name, rule.SourceRef = createdID, s.Name, source
		if rule.Type == "note" {
			rule.Confirmed = false
		} else if !rule.Confirmed {
			fail(w, 400, fmt.Errorf("review and confirm this rule before creating the constraint"))
			return
		}
		d.Constraints = append(d.Constraints, rule)
	} else {
		if strings.TrimSpace(s.Type) == "" {
			fail(w, 400, fmt.Errorf("enter a drill type to match this requirement to event templates"))
			return
		}
		d.Requirements = append(d.Requirements, planner.Requirement{ID: createdID, Name: s.Name, DepartmentID: s.DepartmentID, Type: s.Type, Start: s.Start, End: s.End, Count: s.Count, RepeatDays: s.RepeatDays, Source: fmt.Sprintf("%s, p. %d (%s)", doc.Title, s.Page, s.Basis), SourceRef: source})
	}
	if err = planner.ValidateData(d); err != nil {
		fail(w, 400, err)
		return
	}
	doc.Suggestions[i].CreatedID = createdID
	if err = a.db.SaveDocumentReview(r.Context(), d, doc); err != nil {
		fail(w, 500, err)
		return
	}
	respond(w, 201, doc)
}
