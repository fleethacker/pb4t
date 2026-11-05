package agents

import (
	"context"
	"fmt"
	"log"
	"scheduler/internal/planner"
	"slices"
	"strings"
	"sync"
	"time"
)

// Negotiate uses a barrier between rounds. Only this goroutine publishes progress
// or mutates history; department calls receive the same read-only round snapshot.
func (c *Client) Negotiate(ctx context.Context, runID string, d planner.Data, w time.Time, progress func(string, []planner.DepartmentProposal)) ([]planner.DepartmentProposal, []planner.Preference, error) {
	history := []planner.DepartmentProposal{}
	candidate := d
	feedback := actionable(d, planner.WeekEvents(d, w))
	for round := 1; round <= 2; round++ {
		departments := []planner.Department{}
		for _, dep := range d.Departments {
			if !slices.ContainsFunc(planner.WeekEvents(candidate, w), func(e planner.Event) bool { return e.DepartmentID == dep.ID && !e.Fixed && e.Status != "completed" }) {
				continue
			}
			if round == 2 && len(feedback[dep.ID]) == 0 {
				continue
			}
			departments = append(departments, dep)
		}
		if len(departments) == 0 {
			break
		}
		progress(fmt.Sprintf("Department proposals · round %d of 2 · 0/%d complete", round, len(departments)), history)
		type result struct {
			index    int
			proposal planner.DepartmentProposal
			err      error
		}
		results := make(chan result, len(departments))
		roundCtx, cancel := context.WithCancel(ctx)
		// Bound concurrency while supporting additional departments in larger inputs.
		slots := make(chan struct{}, 4)
		var workers sync.WaitGroup
		for i, dep := range departments {
			workers.Add(1)
			go func(i int, dep planner.Department) {
				defer workers.Done()
				select {
				case slots <- struct{}{}:
				case <-roundCtx.Done():
					results <- result{index: i, err: roundCtx.Err()}
					return
				}
				defer func() { <-slots }()
				started := time.Now()
				trace := fmt.Sprintf("%s/%s/%d", runID, dep.ID, round)
				log.Printf("department trace=%s phase=start", trace)
				p, err := c.Department(WithTrace(roundCtx, trace), candidate, w, dep, round, history, feedback[dep.ID])
				log.Printf("department trace=%s phase=finish duration=%s failed=%t", trace, time.Since(started).Round(time.Millisecond), err != nil)
				results <- result{i, p, err}
			}(i, dep)
		}
		// Collect every result before changing the snapshot or history.
		ordered := make([]planner.DepartmentProposal, len(departments))
		var firstErr error
		for completed := 1; completed <= len(departments); completed++ {
			r := <-results
			if r.err != nil && firstErr == nil {
				firstErr = fmt.Errorf("%s proposal: %w", departments[r.index].Name, r.err)
				cancel()
			}
			ordered[r.index] = r.proposal
			progress(fmt.Sprintf("Department proposals · round %d of 2 · %d/%d complete", round, completed, len(departments)), history)
		}
		workers.Wait()
		cancel()
		if firstErr != nil {
			return history, nil, firstErr
		}
		combined := mergePlacements(candidate.Events, ordered)
		evaluation := planner.Evaluate(candidate, weekEvents(combined, w))
		feedback = actionable(candidate, weekEvents(combined, w))
		for i := range ordered {
			ordered[i].Evaluation = evaluation
			ordered[i].ActionableConflicts = feedback[ordered[i].DepartmentID]
		}
		history = append(history, ordered...)
		candidate.Events = combined
		progress(fmt.Sprintf("Round %d validated · %d shared conflicts", round, len(evaluation.Conflicts)), history)
	}
	// Preserve earlier placements when a revision submits no change for an event.
	latest := map[string]planner.Preference{}
	for _, p := range history {
		for _, pref := range p.Preferences {
			latest[pref.EventID] = pref
		}
	}
	preferences := []planner.Preference{}
	for _, e := range d.Events {
		if pref, ok := latest[e.ID]; ok {
			preferences = append(preferences, pref)
		}
	}
	return history, preferences, nil
}

func mergePlacements(events []planner.Event, proposals []planner.DepartmentProposal) []planner.Event {
	merged := slices.Clone(events)
	for _, p := range proposals {
		for _, pref := range p.Preferences {
			for i := range merged {
				if merged[i].ID == pref.EventID {
					merged[i].Start = pref.Start
				}
			}
		}
	}
	return merged
}
func weekEvents(events []planner.Event, w time.Time) []planner.Event {
	return planner.WeekEvents(planner.Data{Events: events}, w)
}

// actionable attributes direct violations to departments that can change the
// affected events. Minimize within each reported staffing group so independent
// personnel conflicts keep their own department routing.
func actionable(d planner.Data, events []planner.Event) map[string][]planner.Conflict {
	out := map[string][]planner.Conflict{}
	for _, conflict := range planner.Evaluate(d, events).Conflicts {
		if conflict.Type == "personnel" {
			core := []planner.Event{}
			for _, e := range events {
				if slices.Contains(conflict.EventIDs, e.ID) {
					core = append(core, e)
				}
			}
			for i := 0; i < len(core); {
				reduced := slices.Delete(slices.Clone(core), i, i+1)
				stillBlocked := slices.ContainsFunc(planner.Evaluate(d, reduced).Conflicts, func(c planner.Conflict) bool { return c.Type == "personnel" })
				if stillBlocked {
					core = reduced
				} else {
					i++
				}
			}
			conflict.EventIDs = []string{}
			names := []string{}
			for _, e := range core {
				conflict.EventIDs = append(conflict.EventIDs, e.ID)
				names = append(names, e.Name)
			}
			conflict.Message = "No joint staffing assignment satisfies qualifications, availability, and configured rest for: " + strings.Join(names, ", ") + "."
		}
		owners := map[string]bool{}
		for _, e := range events {
			if !e.Fixed && e.Status != "completed" && slices.Contains(conflict.EventIDs, e.ID) {
				owners[e.DepartmentID] = true
			}
		}
		for id := range owners {
			out[id] = append(out[id], conflict)
		}
	}
	return out
}
