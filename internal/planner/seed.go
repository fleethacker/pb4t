package planner

import "time"

func Seed(loc *time.Location) Data {
	w := Week(time.Now(), loc)
	at := func(day, hour int) time.Time { return w.AddDate(0, 0, day).Add(time.Duration(hour) * time.Hour) }
	all := []Window{{at(-1, 0), at(8, 0)}}
	d := Data{Departments: []Department{
		{"engineering", "Engineering", "#bd743e", "Protect engineering watch coverage; prefer morning casualty-control drills."},
		{"combat", "Combat Systems", "#5479ab", "Preserve TAO coverage and simulator access; afternoon exercises are acceptable."},
		{"weapons", "Weapons", "#9771aa", "Complete weapons training before the end of the week; share evaluator time."},
		{"dc", "Damage Control", "#3c9182", "Complete damage-control training with qualified Fire Party Leaders."},
		{"operations", "Ship Operations", "#6f7d87", "Preserve fixed ship operational commitments."}},
		Qualifications: []Qualification{{"fpl", "Fire Party Leader"}, {"eoow", "EOOW"}, {"tao", "TAO"}, {"evaluator", "Drill Evaluator"}},
		People: []Person{
			{"avery", "LT Avery Chen", "engineering", []string{"fpl", "eoow"}, all},
			{"morgan", "CPO Morgan Reed", "dc", []string{"fpl", "evaluator"}, all},
			{"jordan", "LT Jordan Brooks", "combat", []string{"tao", "evaluator"}, all},
			{"riley", "CPO Riley Hayes", "weapons", []string{"evaluator"}, all}},
		Assets: []Asset{{"sim", "Tactical simulator", 1, all}, {"dc-space", "Damage control trainer", 1, all}, {"radios", "Training radios", 4, all}},
		Duties: []Duty{{"overnight", "avery", "Engineering overnight watch", at(-1, 20), at(0, 6), true}},
		Constraints: []Rule{
			{ID: "two-fpl", Name: "2 × Fire Party Leaders", Type: "qualification", Scope: "event", Confirmed: true, QualificationID: "fpl", Count: 2},
			{ID: "one-eoow", Name: "1 × EOOW", Type: "qualification", Scope: "event", Confirmed: true, QualificationID: "eoow", Count: 1},
			{ID: "one-tao", Name: "1 × TAO", Type: "qualification", Scope: "event", Confirmed: true, QualificationID: "tao", Count: 1},
			{ID: "one-evaluator", Name: "1 × Drill Evaluator", Type: "qualification", Scope: "event", Confirmed: true, QualificationID: "evaluator", Count: 1},
			{ID: "sim-rule", Name: "1 × Tactical simulator", Type: "asset", Scope: "event", Confirmed: true, AssetID: "sim", Count: 1},
			{ID: "dc-rule", Name: "1 × Damage control trainer", Type: "asset", Scope: "event", Confirmed: true, AssetID: "dc-space", Count: 1},
			{ID: "no-flight", Name: "No overlap with flight quarters", Type: "incompatibility", Scope: "event", Confirmed: true, TargetTag: "flight"},
			{ID: "rest", Name: "Illustrative 120-minute minimum rest", Type: "rest", Scope: "global", Confirmed: true, Minutes: 120}},
		Events: []Event{}, Templates: []EventTemplate{}, Requirements: []Requirement{}, Completions: []Completion{}, Assignments: []Assignment{}}
	event := func(id, name, dep, kind string, day, hour, duration int, rules []string, fixed bool) Event {
		return Event{ID: id, Name: name, DepartmentID: dep, Contact: "Department training officer", Type: kind, Tags: []string{kind}, Start: at(day, hour), RequestedStart: at(day, hour), Duration: duration, WindowStart: at(0, 8), WindowEnd: at(4, 17), Fixed: fixed, Status: "scheduled", ConstraintIDs: rules, Fatigue: true}
	}
	d.Events = []Event{
		event("flight-mon", "Flight quarters", "operations", "flight", 0, 9, 120, []string{}, true),
		event("dc-mon", "Damage control drill", "dc", "damage-control", 0, 10, 60, []string{"two-fpl", "dc-rule", "no-flight"}, false),
		event("eng-mon", "Engineering casualty control", "engineering", "engineering-casualty", 0, 10, 60, []string{"one-eoow"}, false),
		event("combat-tue", "Combat systems exercise", "combat", "combat-exercise", 1, 10, 90, []string{"one-tao", "sim-rule"}, false),
		event("weapons-tue", "Weapons systems drill", "weapons", "weapons-drill", 1, 10, 60, []string{"one-evaluator", "sim-rule"}, false),
		event("flight-wed", "Flight quarters", "operations", "flight", 2, 9, 120, []string{}, true),
		event("eng-thu", "Engineering readiness review", "engineering", "readiness", 3, 13, 60, []string{"one-eoow"}, false)}
	for i := range d.Events {
		if !d.Events[i].Fixed {
			d.Events[i].Status = "draft"
		}
	}
	seenTypes := map[string]bool{}
	for _, e := range d.Events {
		if seenTypes[e.Type] {
			continue
		}
		seenTypes[e.Type] = true
		d.Templates = append(d.Templates, EventTemplate{ID: "template-" + e.Type, Name: e.Name, DepartmentID: e.DepartmentID, Contact: e.Contact, Type: e.Type, Tags: e.Tags, Duration: e.Duration, ConstraintIDs: e.ConstraintIDs, Fatigue: e.Fatigue})
	}
	d.Requirements = []Requirement{
		{ID: "req-dc", Name: "Damage control periodicity", DepartmentID: "dc", Type: "damage-control", Start: at(0, 0), End: at(5, 0), Count: 2, Source: "Illustrative demo requirement — not official SFTRM policy", RepeatDays: 7},
		{ID: "req-eng", Name: "Engineering casualty training", DepartmentID: "engineering", Type: "engineering-casualty", Start: at(0, 0), End: at(5, 0), Count: 1, Source: "Illustrative demo requirement", RepeatDays: 7},
		{ID: "req-combat", Name: "Combat systems readiness", DepartmentID: "combat", Type: "combat-exercise", Start: at(0, 0), End: at(5, 0), Count: 1, Source: "Illustrative demo requirement", RepeatDays: 7},
	}

	return d
}
