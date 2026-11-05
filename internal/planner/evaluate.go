package planner

import (
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"
)

func Rules(d Data, e Event) []Rule {
	out := []Rule{}
	for _, r := range d.Constraints {
		if r.Scope == "global" || slices.Contains(e.ConstraintIDs, r.ID) {
			out = append(out, r)
		}
	}
	return out
}
func inside(start, end time.Time, windows []Window) bool {
	for _, w := range windows {
		if !start.Before(w.Start) && !end.After(w.End) {
			return true
		}
	}
	return false
}
func restFor(d Data, e Event) int {
	n := 0
	for _, r := range Rules(d, e) {
		if r.Type == "rest" && r.Confirmed {
			n = max(n, r.Minutes)
		}
	}
	return n
}
func tooClose(a, b, c, d time.Time, minutes int) bool {
	gap := time.Duration(minutes) * time.Minute
	return Overlap(a, b, c.Add(-gap), d.Add(gap))
}

// FeasibleFor checks one event's conflicts, or the whole plan when eventID is empty.
func (e Evaluation) FeasibleFor(eventID string) bool {
	if eventID == "" {
		return e.Feasible
	}
	return !slices.ContainsFunc(e.Conflicts, func(c Conflict) bool {
		return slices.Contains(c.EventIDs, eventID)
	})
}

// Evaluate performs the same joint resource and staffing check for every caller.
func Evaluate(d Data, events []Event) Evaluation {
	out := Evaluation{Conflicts: []Conflict{}, Warnings: []Conflict{}, Assignments: []Assignment{}, Coverage: []Coverage{}}
	active := []Event{}
	for _, e := range events {
		if e.Status != "completed" && !e.IgnoreConflicts {
			active = append(active, e)
		}
	}
	add := func(kind, rule, msg string, ids, resources []string, start, end time.Time) {
		out.Conflicts = append(out.Conflicts, Conflict{kind, rule, ids, resources, start, end, msg})
	}
	for _, e := range active {
		if e.Duration <= 0 || e.Start.IsZero() {
			add("time", "", e.Name+": a start time and positive duration are required.", []string{e.ID}, nil, e.Start, e.End())
		}
		if (!e.WindowStart.IsZero() && e.Start.Before(e.WindowStart)) || (!e.WindowEnd.IsZero() && e.End().After(e.WindowEnd)) {
			add("window", "", e.Name+" falls outside its allowed scheduling window.", []string{e.ID}, nil, e.Start, e.End())
		}
		for _, id := range e.ConstraintIDs {
			found := false
			for _, r := range d.Constraints {
				if r.ID == id {
					found = true
				}
			}
			if !found {
				add("constraint", id, e.Name+": attached constraint is missing.", []string{e.ID}, nil, e.Start, e.End())
			}
		}
		for _, r := range Rules(d, e) {
			if !r.Confirmed || r.Type == "note" {
				add("unenforced", r.ID, e.Name+": "+r.Name+" is not an enforced structured rule.", []string{e.ID}, nil, e.Start, e.End())
				continue
			}
			switch r.Type {
			case "allowed":
				if e.Start.Before(r.Start) || e.End().After(r.End) {
					add("window", r.ID, e.Name+" must fit "+r.Name+".", []string{e.ID}, nil, e.Start, e.End())
				}
			case "excluded":
				if Overlap(e.Start, e.End(), r.Start, r.End) {
					add("window", r.ID, e.Name+" overlaps "+r.Name+".", []string{e.ID}, nil, e.Start, e.End())
				}
			case "incompatibility":
				for _, other := range active {
					if other.ID != e.ID && (other.ID == r.TargetEventID || (r.TargetTag != "" && slices.Contains(other.Tags, r.TargetTag))) && Overlap(e.Start, e.End(), other.Start, other.End()) {
						add("incompatibility", r.ID, e.Name+" cannot overlap "+other.Name+" ("+r.Name+").", []string{e.ID, other.ID}, nil, e.Start, e.End())
					}
				}
			case "qualification", "person", "asset", "rest":
			default:
				add("unenforced", r.ID, e.Name+": unsupported constraint type.", []string{e.ID}, nil, e.Start, e.End())
			}
		}
		if !e.RequestedStart.IsZero() && !e.Start.Equal(e.RequestedStart) {
			out.Warnings = append(out.Warnings, Conflict{Type: "preference", EventIDs: []string{e.ID}, Start: e.Start, End: e.End(), Message: e.Name + " differs from its requested time."})
		}
	}
	for _, a := range d.Assets {
		type demand struct {
			e    Event
			n    int
			rule string
		}
		demands := []demand{}
		for _, e := range active {
			for _, r := range Rules(d, e) {
				if r.Type == "asset" && r.AssetID == a.ID && r.Confirmed {
					demands = append(demands, demand{e, r.Count, r.ID})
					if !inside(e.Start, e.End(), a.Availability) {
						add("asset", r.ID, a.Name+" is unavailable for "+e.Name+".", []string{e.ID}, []string{a.ID}, e.Start, e.End())
					}
				}
			}
		}
		reported := [][]string{}
		for _, point := range demands {
			sum := 0
			ids := []string{}
			for _, x := range demands {
				if !point.e.Start.Before(x.e.Start) && point.e.Start.Before(x.e.End()) {
					sum += x.n
					ids = append(ids, x.e.ID)
				}
			}
			if sum > a.Capacity && !slices.ContainsFunc(reported, func(previous []string) bool { return slices.Equal(previous, ids) }) {
				add("asset", point.rule, fmt.Sprintf("%s: concurrent demand %d exceeds capacity %d.", a.Name, sum, a.Capacity), ids, []string{a.ID}, point.e.Start, point.e.End())
				reported = append(reported, ids)
			}
		}
	}
	slots := []personnelSlot{}
	for _, e := range active {
		for _, r := range Rules(d, e) {
			if !r.Confirmed || (r.Type != "qualification" && r.Type != "person") {
				continue
			}
			count := r.Count
			if r.Type == "person" {
				count = 1
			}
			s := personnelSlot{e: e, rule: r, candidates: []Person{}, blocked: []string{}}
			for _, p := range d.People {
				if (r.Type == "person" && r.PersonID != p.ID) || (r.Type == "qualification" && !slices.Contains(p.Qualifications, r.QualificationID)) {
					continue
				}
				ok := inside(e.Start, e.End(), p.Availability)
				if !ok {
					s.blocked = append(s.blocked, p.Name+": unavailable")
				}
				for _, duty := range d.Duties {
					if duty.PersonID != p.ID {
						continue
					}
					if Overlap(e.Start, e.End(), duty.Start, duty.End) {
						ok = false
						s.blocked = append(s.blocked, p.Name+": "+duty.Name)
					} else if duty.Fatigue && e.Fatigue && tooClose(e.Start, e.End(), duty.Start, duty.End, restFor(d, e)) {
						ok = false
						s.blocked = append(s.blocked, fmt.Sprintf("%s: %s ends %s; needs %d minutes rest before/after duty", p.Name, duty.Name, duty.End.Format(time.RFC3339), restFor(d, e)))
					}
				}
				for _, a := range d.Assignments {
					if a.PersonID != p.ID {
						continue
					}
					for _, history := range d.Events {
						if history.ID == a.EventID && history.Status == "completed" && !history.IgnoreConflicts && history.Fatigue && e.Fatigue && tooClose(e.Start, e.End(), history.Start, history.End(), restFor(d, e)) {
							ok = false
							s.blocked = append(s.blocked, p.Name+": rest after "+history.Name)
						}
					}
				}
				if ok {
					s.candidates = append(s.candidates, p)
				}
			}
			for i := 0; i < count; i++ {
				slots = append(slots, s)
			}
		}
	}
	for _, slots := range personnelGroups(d, slots) {
		sort.SliceStable(slots, func(i, j int) bool { return len(slots[i].candidates) < len(slots[j].candidates) })
		chosen := make([]Assignment, len(slots))
		nodes := 0
		var assign func(int) bool
		assign = func(i int) bool {
			nodes++
			if nodes > 100000 {
				return false
			}
			if i == len(slots) {
				return true
			}
			s := slots[i]
			for _, p := range s.candidates {
				ok := true
				for j := 0; j < i; j++ {
					if chosen[j].PersonID != p.ID {
						continue
					}
					other := slots[j].e
					if competingPersonnel(d, s.e, other) {
						ok = false
						break
					}
				}
				if ok {
					role := s.rule.QualificationID
					if s.rule.Type == "person" {
						role = "required participant"
					}
					chosen[i] = Assignment{s.e.ID, p.ID, role}
					if assign(i + 1) {
						return true
					}
				}
			}
			return false
		}
		if assign(0) {
			out.Assignments = append(out.Assignments, chosen...)
		} else {
			ids, resources := []string{}, []string{}
			details := []string{}
			for _, s := range slots {
				if !slices.Contains(ids, s.e.ID) {
					ids = append(ids, s.e.ID)
				}
				names := []string{}
				for _, p := range s.candidates {
					names = append(names, p.Name)
					if !slices.Contains(resources, p.ID) {
						resources = append(resources, p.ID)
					}
				}
				role := s.rule.QualificationID
				for _, q := range d.Qualifications {
					if q.ID == role {
						role = q.Name
					}
				}
				if s.rule.Type == "person" {
					role = "named participant " + s.rule.PersonID
				}
				detail := fmt.Sprintf("%s requires %d × %s; available: %s", s.e.Name, max(1, s.rule.Count), role, strings.Join(names, ", "))
				if len(s.blocked) > 0 {
					detail += ". Restricted: " + strings.Join(s.blocked, "; ")
				}
				if !slices.Contains(details, detail) {
					details = append(details, detail)
				}
			}
			start, end := time.Time{}, time.Time{}
			if len(slots) > 0 {
				start = slots[0].e.Start
				end = slots[0].e.End()
				for _, s := range slots[1:] {
					start = minTime(start, s.e.Start)
					end = maxTime(end, s.e.End())
				}
			}
			add("personnel", "", "No joint personnel assignment satisfies availability, collisions, and configured rest. "+strings.Join(details, " | "), ids, resources, start, end)
		}
	}
	out.Feasible = len(out.Conflicts) == 0
	out.Coverage = Training(d, events, out.Conflicts)
	return out
}

func Training(d Data, events []Event, conflicts []Conflict) []Coverage {
	out := []Coverage{}
	bad := map[string]bool{}
	for _, c := range conflicts {
		for _, id := range c.EventIDs {
			bad[id] = true
		}
	}
	for _, r := range d.Requirements {
		c := Coverage{Requirement: r}
		linked := map[string]bool{}
		for _, x := range d.Completions {
			if x.DepartmentID == r.DepartmentID && x.Type == r.Type && !x.At.Before(r.Start) && x.At.Before(r.End) {
				c.Completed++
				if x.EventID != "" {
					linked[x.EventID] = true
				}
			}
		}
		for _, e := range events {
			if e.DepartmentID == r.DepartmentID && e.Type == r.Type && e.Status == "scheduled" && !linked[e.ID] && !bad[e.ID] && !e.Start.Before(r.Start) && !e.End().After(r.End) {
				c.Planned++
			}
		}
		c.Gap = max(0, r.Count-c.Completed-c.Planned)
		out = append(out, c)
	}
	return out
}

// MergeAssignments retains historical and out-of-week staffing records.
func MergeAssignments(d Data, events []Event, assignments []Assignment) []Assignment {
	out := slices.DeleteFunc(slices.Clone(d.Assignments), func(a Assignment) bool {
		return slices.ContainsFunc(events, func(e Event) bool { return e.ID == a.EventID && e.Status != "completed" })
	})
	return append(out, assignments...)
}
