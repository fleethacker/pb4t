package planner

import (
	"encoding/json"
	"fmt"
	"slices"
	"time"
)

func update[T any](list *[]T, raw json.RawMessage, id string, remove bool) error {
	var item T
	if !remove {
		if err := json.Unmarshal(raw, &item); err != nil {
			return err
		}
	}
	out := []T{}
	for _, x := range *list {
		b, _ := json.Marshal(x)
		var key struct {
			ID string `json:"id"`
		}
		_ = json.Unmarshal(b, &key)
		if key.ID != id {
			out = append(out, x)
		}
	}
	if !remove {
		out = append(out, item)
	}
	*list = out
	return nil
}
func Edit(d *Data, kind, id string, raw json.RawMessage, remove bool) error {
	var err error
	switch kind {
	case "departments":
		err = update(&d.Departments, raw, id, remove)
	case "qualifications":
		err = update(&d.Qualifications, raw, id, remove)
	case "people":
		err = update(&d.People, raw, id, remove)
	case "duties":
		err = update(&d.Duties, raw, id, remove)
	case "assets":
		err = update(&d.Assets, raw, id, remove)
	case "constraints":
		err = update(&d.Constraints, raw, id, remove)
	case "events":
		err = update(&d.Events, raw, id, remove)
	case "templates":
		err = update(&d.Templates, raw, id, remove)
	case "requirements":
		err = update(&d.Requirements, raw, id, remove)
	case "completions":
		err = update(&d.Completions, raw, id, remove)
	default:
		return fmt.Errorf("unknown planning collection")
	}
	if err != nil {
		return err
	}
	if kind == "events" && (remove || !slices.ContainsFunc(d.Events, func(e Event) bool { return e.ID == id && e.Status == "completed" })) {
		d.Assignments = slices.DeleteFunc(d.Assignments, func(a Assignment) bool { return a.EventID == id })
	}
	return ValidateData(*d)
}
func ValidateEvent(e Event) error {
	if e.ID == "" || e.Name == "" || e.DepartmentID == "" {
		return fmt.Errorf("event requires an ID, name, and department")
	}
	if e.Duration < 15 || e.Duration%15 != 0 {
		return fmt.Errorf("duration must be a positive multiple of 15 minutes")
	}
	if e.Start.IsZero() {
		return fmt.Errorf("event requires a start time")
	}
	if !e.WindowStart.IsZero() && !e.WindowEnd.IsZero() && !e.WindowEnd.After(e.WindowStart) {
		return fmt.Errorf("latest allowed end must be after earliest allowed start")
	}
	if e.Status != "draft" && e.Status != "scheduled" && e.Status != "completed" {
		return fmt.Errorf("unknown event status")
	}
	return nil
}
func ValidateRule(d Data, r Rule) error {
	if r.ID == "" || r.Name == "" {
		return fmt.Errorf("constraint requires a name and ID")
	}
	if r.Scope != "event" && r.Scope != "global" {
		return fmt.Errorf("constraint scope must be event or global")
	}
	switch r.Type {
	case "qualification":
		if r.Count < 1 || !slices.ContainsFunc(d.Qualifications, func(q Qualification) bool { return q.ID == r.QualificationID }) {
			return fmt.Errorf("select a qualification and positive count")
		}
	case "person":
		if !slices.ContainsFunc(d.People, func(p Person) bool { return p.ID == r.PersonID }) {
			return fmt.Errorf("select a named participant")
		}
	case "asset":
		if r.Count < 1 || !slices.ContainsFunc(d.Assets, func(a Asset) bool { return a.ID == r.AssetID }) {
			return fmt.Errorf("select an asset and positive quantity")
		}
	case "allowed", "excluded":
		if r.Start.IsZero() || !r.End.After(r.Start) {
			return fmt.Errorf("select a valid time window")
		}
	case "incompatibility":
		if r.TargetTag == "" && r.TargetEventID == "" {
			return fmt.Errorf("select an incompatible event or tag")
		}
		if r.TargetEventID != "" && !slices.ContainsFunc(d.Events, func(e Event) bool { return e.ID == r.TargetEventID }) {
			return fmt.Errorf("incompatible event does not exist")
		}
	case "rest":
		if r.Minutes < 0 {
			return fmt.Errorf("minimum rest cannot be negative")
		}
	case "note":
	default:
		return fmt.Errorf("unsupported constraint type")
	}
	return nil
}
func ValidateData(d Data) error {
	dep := func(id string) bool {
		return slices.ContainsFunc(d.Departments, func(x Department) bool { return x.ID == id })
	}
	person := func(id string) bool { return slices.ContainsFunc(d.People, func(x Person) bool { return x.ID == id }) }
	windows := func(ws []Window) bool {
		for _, w := range ws {
			if w.Start.IsZero() || !w.End.After(w.Start) {
				return false
			}
		}
		return true
	}
	for _, p := range d.People {
		if p.ID == "" || p.Name == "" || !dep(p.DepartmentID) || !windows(p.Availability) {
			return fmt.Errorf("person requires a name, department, and valid availability")
		}
		for _, id := range p.Qualifications {
			if !slices.ContainsFunc(d.Qualifications, func(q Qualification) bool { return q.ID == id }) {
				return fmt.Errorf("person references missing qualification")
			}
		}
	}
	for _, a := range d.Assets {
		if a.ID == "" || a.Name == "" || a.Capacity < 1 || !windows(a.Availability) {
			return fmt.Errorf("asset requires a name, positive capacity, and valid availability")
		}
	}
	for _, x := range d.Duties {
		if !person(x.PersonID) || !x.End.After(x.Start) {
			return fmt.Errorf("duty requires an existing person and valid interval")
		}
	}
	for _, r := range d.Constraints {
		if err := ValidateRule(d, r); err != nil {
			return err
		}
	}
	for _, e := range d.Events {
		if err := ValidateEvent(e); err != nil {
			return err
		}
		if !dep(e.DepartmentID) {
			return fmt.Errorf("event department does not exist")
		}
		for _, id := range e.ConstraintIDs {
			if !slices.ContainsFunc(d.Constraints, func(r Rule) bool { return r.ID == id }) {
				return fmt.Errorf("event references missing constraint; detach it first")
			}
		}
	}
	for _, t := range d.Templates {
		if t.ID == "" || t.Name == "" || !dep(t.DepartmentID) || t.Duration < 15 || t.Duration%15 != 0 {
			return fmt.Errorf("template requires a name, department, and duration in 15-minute increments")
		}
		for _, id := range t.ConstraintIDs {
			if !slices.ContainsFunc(d.Constraints, func(r Rule) bool { return r.ID == id }) {
				return fmt.Errorf("template references missing constraint; detach it first")
			}
		}
	}
	for _, r := range d.Requirements {
		if r.ID == "" || r.Name == "" || !dep(r.DepartmentID) || r.Count < 1 || !r.End.After(r.Start) || r.RepeatDays < 0 {
			return fmt.Errorf("requirement needs a name, department, count, and valid coverage window")
		}
	}
	for _, c := range d.Completions {
		if !dep(c.DepartmentID) || c.Type == "" || c.At.IsZero() {
			return fmt.Errorf("completion needs department, drill type, and completion time")
		}
		if c.EventID != "" && !slices.ContainsFunc(d.Events, func(e Event) bool { return e.ID == c.EventID }) {
			return fmt.Errorf("completion references missing event")
		}
	}
	for _, x := range d.Departments {
		if x.ID == "" || x.Name == "" {
			return fmt.Errorf("department needs an ID and name")
		}
	}
	for _, x := range d.Qualifications {
		if x.ID == "" || x.Name == "" {
			return fmt.Errorf("qualification needs an ID and name")
		}
	}
	return nil
}
func Relevant(d Data, w time.Time) Data { d = ExpandRequirements(d, w); return d }
