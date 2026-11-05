package planner

import (
	"fmt"
	"slices"
	"sort"
	"time"
)

func WeekEvents(d Data, w time.Time) []Event {
	out := []Event{}
	end := w.AddDate(0, 0, 7)
	for _, e := range d.Events {
		if Overlap(e.Start, e.End(), w, end) {
			out = append(out, e)
		}
	}
	return out
}
func ExpandRequirements(d Data, w time.Time) Data {
	out := d
	out.Requirements = []Requirement{}
	for _, r := range d.Requirements {
		if r.RepeatDays > 0 && !r.End.After(w) {
			for !r.End.After(w) {
				r.Start = r.Start.AddDate(0, 0, r.RepeatDays)
				r.End = r.End.AddDate(0, 0, r.RepeatDays)
			}
		}
		if Overlap(r.Start, r.End, w, w.AddDate(0, 0, 7)) {
			out.Requirements = append(out.Requirements, r)
		}
	}
	return out
}
func AddMissing(d Data, events []Event, w time.Time) []Event {
	intended := slices.Clone(events)
	for i := range intended {
		if intended[i].Status != "completed" && !intended[i].IgnoreConflicts {
			intended[i].Status = "scheduled"
		}
	}
	for _, c := range Training(d, intended, nil) {
		r := c.Requirement
		if r.End.After(w.AddDate(0, 0, 7)) {
			continue
		}
		for n := 0; n < c.Gap; n++ {
			for _, t := range d.Templates {
				if t.Type == r.Type && t.DepartmentID == r.DepartmentID {
					windowStart := maxTime(w, r.Start)
					windowEnd := minTime(w.AddDate(0, 0, 7), r.End)
					occurrence := t.Occurrence(windowStart.Add(10*time.Hour), windowStart, windowEnd)
					occurrence.Name = fmt.Sprintf("%s · required occurrence", t.Name)
					occurrence.Status = "scheduled"
					events = append(events, occurrence)
					break
				}
			}
		}
	}
	return events
}
func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}
func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}
func score(events, original []Event, prefs []Preference) [3]int {
	s := [3]int{}
	for _, e := range events {
		existed := false
		for _, old := range original {
			if e.ID == old.ID {
				existed = true
			}
			if e.ID == old.ID && !e.Start.Equal(old.Start) {
				s[0]++
				s[1] += int(e.Start.Sub(old.Start).Abs().Minutes())
			}
		}
		if !existed && !e.RequestedStart.IsZero() {
			s[1] += int(e.Start.Sub(e.RequestedStart).Abs().Minutes())
		}
		if !e.RequestedStart.IsZero() && !e.Start.Equal(e.RequestedStart) {
			s[2]++
		}
		for _, p := range prefs {
			if p.EventID == e.ID && !e.Start.Equal(p.Start) {
				s[2]++
			}
		}
	}
	return s
}
func less(a, b [3]int) bool {
	for i := range a {
		if a[i] != b[i] {
			return a[i] < b[i]
		}
	}
	return false
}

// Solve searches the weekly plan, or only target's placement when target is set.
func Solve(d Data, w time.Time, target *Event, prefs []Preference, baseline []Event) Proposal {
	original := baseline
	events := WeekEvents(d, w)
	if target != nil {
		found := false
		for i := range events {
			if events[i].ID == target.ID {
				events[i] = *target
				found = true
			}
		}
		if !found {
			events = append(events, *target)
		}
	}
	if target == nil {
		events = AddMissing(d, events, w)
	}
	for i := range events {
		if events[i].Status != "completed" && !events[i].IgnoreConflicts && (target == nil || events[i].ID == target.ID) {
			events[i].Status = "scheduled"
		}
	}
	p := Proposal{Week: w, Status: "no_solution_found", Events: []Event{}, Changes: []Change{}, Unscheduled: []Event{}, Negotiation: []DepartmentProposal{}, PreviousAssignments: d.Assignments, Summary: "No complete plan found within the bounded search."}
	if target != nil {
		p.TargetEventID = target.ID
	}
	type variable struct {
		e     Event
		slots []time.Time
	}
	vars := []variable{}
	placed := []Event{}
	for _, e := range events {
		if e.Fixed || e.IgnoreConflicts || e.Status == "completed" || (target != nil && e.ID != target.ID) {
			placed = append(placed, e)
			continue
		}
		starts := []time.Time{e.Start}
		lo := maxTime(w, e.WindowStart)
		hi := minTime(w.AddDate(0, 0, 7), e.WindowEnd)
		if e.WindowEnd.IsZero() {
			hi = w.AddDate(0, 0, 7)
		}
		for t := lo; t.Add(time.Duration(e.Duration)*time.Minute).Compare(hi) <= 0; t = t.Add(15 * time.Minute) {
			local := t.In(w.Location())
			if local.Hour() < 8 || local.Hour() >= 17 || local.Weekday() == time.Saturday || local.Weekday() == time.Sunday {
				continue
			}
			if t.Add(time.Duration(e.Duration)*time.Minute).In(w.Location()).Hour() > 17 {
				continue
			}
			if !t.Equal(e.Start) {
				starts = append(starts, t)
			}
		}
		sort.SliceStable(starts, func(i, j int) bool {
			ei, ej := e, e
			ei.Start = starts[i]
			ej.Start = starts[j]
			return less(score([]Event{ei}, original, prefs), score([]Event{ej}, original, prefs))
		})
		// Remove individually invalid positions before ordering the search variables.
		valid := []time.Time{}
		for _, t := range starts {
			x := e
			x.Start = t
			if Evaluate(d, []Event{x}).Feasible {
				valid = append(valid, t)
			}
		}
		vars = append(vars, variable{e, valid})
	}
	sort.SliceStable(vars, func(i, j int) bool { return len(vars[i].slots) < len(vars[j].slots) })
	deadline := time.Now().Add(3 * time.Second)
	nodes := 0
	best := [3]int{1 << 30, 1 << 30, 1 << 30}
	partial := slices.Clone(placed)
	var search func(int)
	search = func(i int) {
		nodes++
		if nodes > 30000 || time.Now().After(deadline) {
			return
		}
		if len(placed) > len(partial) {
			partial = slices.Clone(placed)
		}
		sc := score(placed, original, prefs)
		if !less(sc, best) {
			return
		}
		if i == len(vars) {
			evaluation := Evaluate(d, placed)
			for _, c := range evaluation.Coverage {
				if target == nil && !c.Requirement.End.After(w.AddDate(0, 0, 7)) && c.Gap > 0 {
					return
				}
			}
			if evaluation.FeasibleFor(p.TargetEventID) {
				best = sc
				p.Events = slices.Clone(placed)
				p.Evaluation = evaluation
				p.Score = sc
				p.Status = "feasible"
			}
			return
		}
		v := vars[i]
		for _, t := range v.slots {
			e := v.e
			e.Start = t
			placed = append(placed, e)
			if Evaluate(d, placed).FeasibleFor(p.TargetEventID) {
				search(i + 1)
			}
			placed = placed[:len(placed)-1]
			if nodes > 30000 || time.Now().After(deadline) {
				break
			}
		}
	}
	if Evaluate(d, placed).FeasibleFor(p.TargetEventID) {
		search(0)
	}
	if p.Status != "feasible" {
		if target != nil {
			p.Events = events
			p.Evaluation = Evaluate(d, events)
			p.Summary = "No feasible placement found for this event while keeping all other events unchanged. Review the reported conflicts or resolve the weekly plan."
			return p
		}
		p.Events = partial
		p.Evaluation = Evaluate(d, events)
		if len(partial) > 0 {
			p.Status = "partial"
		}
		for _, e := range events {
			found := false
			for _, x := range partial {
				if x.ID == e.ID {
					found = true
				}
			}
			if !found {
				p.Unscheduled = append(p.Unscheduled, e)
			}
		}
		return p
	}
	p.Summary = "Recommended feasible plan. All included events are staffed and supplied hard constraints are satisfied. Review changes before applying."
	if slices.ContainsFunc(p.Events, func(e Event) bool { return e.IgnoreConflicts }) {
		p.Summary = "Recommended plan. Conflicts on ignored events were waived and their saved times were retained. All other hard constraints are satisfied. Review changes before applying."
	}
	if target != nil {
		p.Summary = "Recommended placement for this event. All other events are unchanged. Review the suggested time before applying."
		if !p.Evaluation.Feasible {
			p.Summary += " Unrelated conflicts remain elsewhere in the weekly plan."
		}
	}
	for _, e := range p.Events {
		if e.IgnoreConflicts {
			continue
		}
		var before *time.Time
		reason := "Added to close a training gap"
		if target != nil {
			reason = "Placed this event while keeping all other events unchanged"
		}
		for _, old := range original {
			if e.ID == old.ID {
				b := old.Start
				before = &b
				reason = "Rescheduled to satisfy shared constraints"
				if old.Start.Equal(e.Start) {
					reason = "Placement retained; staffing validated"
				}
			}
		}
		if before == nil || !before.Equal(e.Start) {
			p.Changes = append(p.Changes, Change{e.ID, e.Name, before, e.Start, reason})
		}
	}
	return p
}
