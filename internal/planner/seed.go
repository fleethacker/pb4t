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
		Qualifications: []Qualification{{"fpl", "Fire Party Leader"}, {"eoow", "EOOW"}, {"tao", "TAO"}, {"evaluator", "Drill Evaluator"},
			{ID: "observer", Name: "Observer"}, {ID: "medic", Name: "Medical responder"}, {ID: "diver", Name: "Diver"},
			{ID: "aloft-supervisor", Name: "Aloft supervisor"}, {ID: "comms-operator", Name: "Communications operator"}},
		People: []Person{
			{"avery", "LT Avery Chen", "engineering", []string{"fpl", "eoow"}, all},
			{"morgan", "CPO Morgan Reed", "dc", []string{"fpl", "evaluator"}, all},
			{"jordan", "LT Jordan Brooks", "combat", []string{"tao", "evaluator"}, all},
			{"riley", "CPO Riley Hayes", "weapons", []string{"evaluator"}, all},
			{ID: "casey", Name: "PO Casey Nguyen", DepartmentID: "operations", Qualifications: []string{"comms-operator"}, Availability: all},
			{ID: "sam", Name: "PO Sam Patel", DepartmentID: "operations", Qualifications: []string{"aloft-supervisor"}, Availability: all},
			{ID: "alex", Name: "PO Alex Rivera", DepartmentID: "operations", Qualifications: []string{"diver"}, Availability: all},
			{ID: "jamie", Name: "PO Jamie Ellis", DepartmentID: "dc", Qualifications: []string{"medic"}, Availability: all},
			{ID: "taylor", Name: "PO Taylor Kim", DepartmentID: "engineering", Qualifications: []string{"observer"}, Availability: all},
			{ID: "drew", Name: "PO Drew Bennett", DepartmentID: "weapons", Qualifications: []string{"observer"}, Availability: all}},
		Assets: []Asset{{"sim", "Tactical simulator", 1, all}, {"dc-space", "Damage control trainer", 1, all}, {"radios", "Training radios", 4, all},
			{ID: "radar", Name: "Training radar", Capacity: 1, Availability: all},
			{ID: "mess-decks", Name: "Mess decks", Capacity: 1, Availability: all},
			{ID: "dive-kit", Name: "Dive support kit", Capacity: 1, Availability: all}},
		Duties: []Duty{{"overnight", "avery", "Engineering overnight watch", at(-1, 20), at(0, 6), true}},
		Constraints: []Rule{
			{ID: "two-fpl", Name: "2 × Fire Party Leaders", Type: "qualification", Scope: "event", Confirmed: true, QualificationID: "fpl", Count: 2},
			{ID: "one-eoow", Name: "1 × EOOW", Type: "qualification", Scope: "event", Confirmed: true, QualificationID: "eoow", Count: 1},
			{ID: "one-tao", Name: "1 × TAO", Type: "qualification", Scope: "event", Confirmed: true, QualificationID: "tao", Count: 1},
			{ID: "one-evaluator", Name: "1 × Drill Evaluator", Type: "qualification", Scope: "event", Confirmed: true, QualificationID: "evaluator", Count: 1},
			{ID: "sim-rule", Name: "1 × Tactical simulator", Type: "asset", Scope: "event", Confirmed: true, AssetID: "sim", Count: 1},
			{ID: "dc-rule", Name: "1 × Damage control trainer", Type: "asset", Scope: "event", Confirmed: true, AssetID: "dc-space", Count: 1},
			{ID: "no-flight", Name: "No overlap with flight quarters", Type: "incompatibility", Scope: "event", Confirmed: true, TargetTag: "flight"},
			{ID: "one-observer", Name: "1 × Observer", Type: "qualification", Scope: "event", Confirmed: true, QualificationID: "observer", Count: 1},
			{ID: "one-medic", Name: "1 × Medical responder", Type: "qualification", Scope: "event", Confirmed: true, QualificationID: "medic", Count: 1},
			{ID: "one-diver", Name: "1 × Diver", Type: "qualification", Scope: "event", Confirmed: true, QualificationID: "diver", Count: 1},
			{ID: "one-aloft-supervisor", Name: "1 × Aloft supervisor", Type: "qualification", Scope: "event", Confirmed: true, QualificationID: "aloft-supervisor", Count: 1},
			{ID: "one-comms-operator", Name: "1 × Communications operator", Type: "qualification", Scope: "event", Confirmed: true, QualificationID: "comms-operator", Count: 1},
			{ID: "no-radios", Name: "No radios", Type: "incompatibility", Scope: "event", Confirmed: true, TargetTag: "needs-radios"},
			{ID: "needs-radios", Name: "Needs radios", Type: "asset", Scope: "event", Confirmed: true, AssetID: "radios", Count: 1},
			{ID: "no-radars", Name: "No radars", Type: "incompatibility", Scope: "event", Confirmed: true, TargetTag: "needs-radars"},
			{ID: "needs-radars", Name: "Needs radars", Type: "asset", Scope: "event", Confirmed: true, AssetID: "radar", Count: 1},
			{ID: "no-mess-decks", Name: "No mess decks (food)", Type: "incompatibility", Scope: "event", Confirmed: true, TargetTag: "needs-mess-decks"},
			{ID: "needs-mess-decks", Name: "Needs mess decks", Type: "asset", Scope: "event", Confirmed: true, AssetID: "mess-decks", Count: 1},
			{ID: "dive-kit-rule", Name: "1 × Dive support kit", Type: "asset", Scope: "event", Confirmed: true, AssetID: "dive-kit", Count: 1},
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
	// CSV event names and tag pairings inspire these examples. Timing, staffing,
	// resource quantities, and descriptions are illustrative demo choices.
	samples := []struct {
		id, name, department, kind, contact string
		day, hour, duration                 int
		tags, rules                         []string
		fixed                               bool
	}{
		{"aloft-mon", "Aloft", "operations", "aloft", "Topside supervisor", 0, 14, 60, []string{"no-radios"}, []string{"no-radios", "one-aloft-supervisor", "one-observer"}, false},
		{"hf-mon", "HF radio event", "operations", "hf-radio", "Communications training officer", 0, 14, 60, []string{"needs-radios"}, []string{"needs-radios", "one-comms-operator"}, false},
		{"mass-casualty-tue", "Mass casualty", "dc", "mass-casualty", "Medical training coordinator", 1, 12, 60, []string{"no-mess-decks"}, []string{"no-mess-decks", "one-medic", "one-observer"}, false},
		{"chow-tue", "Chow (food)", "operations", "chow", "Food service coordinator", 1, 12, 60, []string{"needs-mess-decks"}, []string{"needs-mess-decks"}, true},
		{"dc-fire-tue", "DC fire drill", "dc", "dc-fire", "Damage control training officer", 1, 15, 60, []string{}, []string{"two-fpl", "dc-rule", "no-flight"}, false},
		{"flight-ops-wed", "Flight operations", "operations", "flight-operations", "Air operations coordinator", 2, 13, 90, []string{"flight", "needs-radios", "no-radars"}, []string{"needs-radios", "no-radars"}, true},
		{"radar-wed", "Radar tracking check", "combat", "radar-tracking", "Combat systems training officer", 2, 13, 60, []string{"needs-radars"}, []string{"needs-radars", "one-tao"}, false},
		{"steering-thu", "Loss of steering drill", "engineering", "loss-of-steering", "Engineering training officer", 3, 9, 60, []string{}, []string{"one-eoow", "one-observer"}, false},
		{"divers-thu", "Divers over the side", "operations", "diving", "Dive supervisor", 3, 14, 90, []string{}, []string{"one-diver", "one-observer", "dive-kit-rule"}, false},
		{"comms-fri", "Comms check", "operations", "comms-check", "Communications training officer", 4, 9, 30, []string{"needs-radios"}, []string{"needs-radios", "one-comms-operator"}, false},
		{"emcon-fri", "EMCON", "operations", "emcon", "Operations training officer", 4, 9, 60, []string{"no-radios"}, []string{"no-radios"}, false},
		{"ammo-fri", "Moving ammo", "weapons", "ammo-movement", "Weapons training officer", 4, 13, 60, []string{"no-radios", "no-radars"}, []string{"no-radios", "no-radars", "one-evaluator"}, false},
		{"cstt-fri", "CSTT training drill", "combat", "cstt-drill", "Combat systems training officer", 4, 13, 60, []string{"needs-radios"}, []string{"needs-radios", "one-tao", "sim-rule"}, false},
		{"comms-sat", "Comms check", "operations", "comms-check", "Communications training officer", 5, 9, 30, []string{"needs-radios"}, []string{"needs-radios", "one-comms-operator"}, false},
		{"eng-sat", "Engineering readiness review", "engineering", "readiness", "Engineering training officer", 5, 11, 60, []string{}, []string{"one-eoow"}, false},
		{"dc-fire-sun", "DC fire drill", "dc", "dc-fire", "Damage control training officer", 6, 10, 60, []string{}, []string{"two-fpl", "dc-rule", "no-flight"}, false},
		{"cstt-sun", "CSTT training drill", "combat", "cstt-drill", "Combat systems training officer", 6, 13, 60, []string{"needs-radios"}, []string{"needs-radios", "one-tao", "sim-rule"}, false},
	}
	for _, sample := range samples {
		e := event(sample.id, sample.name, sample.department, sample.kind, sample.day, sample.hour, sample.duration, sample.rules, sample.fixed)
		e.Tags = append(e.Tags, sample.tags...)
		e.Contact = sample.contact
		e.Description = "Illustrative sample scenario. Duration, staffing, and resource settings are editable demo defaults, not operational policy."
		if sample.day >= 5 {
			e.WindowEnd = at(6, 17)
		}
		d.Events = append(d.Events, e)
	}
	for i := range d.Events {
		d.Events[i].TemplateID = "template-" + d.Events[i].Type
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
		d.Templates = append(d.Templates, EventTemplate{ID: e.TemplateID, Description: e.Description, Name: e.Name, DepartmentID: e.DepartmentID, Contact: e.Contact, Type: e.Type, Tags: e.Tags, Duration: e.Duration, ConstraintIDs: e.ConstraintIDs, Fatigue: e.Fatigue})
	}
	d.Requirements = []Requirement{
		{ID: "req-dc", Name: "Damage control periodicity", DepartmentID: "dc", Type: "damage-control", Start: at(0, 0), End: at(5, 0), Count: 2, Source: "Illustrative demo requirement — not official SFTRM policy", RepeatDays: 7},
		{ID: "req-eng", Name: "Engineering casualty training", DepartmentID: "engineering", Type: "engineering-casualty", Start: at(0, 0), End: at(5, 0), Count: 1, Source: "Illustrative demo requirement", RepeatDays: 7},
		{ID: "req-combat", Name: "Combat systems readiness", DepartmentID: "combat", Type: "combat-exercise", Start: at(0, 0), End: at(5, 0), Count: 1, Source: "Illustrative demo requirement", RepeatDays: 7},
		{ID: "req-steering", Name: "Steering casualty practice", DepartmentID: "engineering", Type: "loss-of-steering", Start: at(0, 0), End: at(5, 0), Count: 1, Source: "Illustrative demo requirement — frequency not supplied by the CSV", RepeatDays: 7},
		{ID: "req-comms", Name: "Communications practice", DepartmentID: "operations", Type: "comms-check", Start: at(0, 0), End: at(5, 0), Count: 1, Source: "Illustrative demo requirement — frequency not supplied by the CSV", RepeatDays: 7},
		{ID: "req-medical", Name: "Medical response practice", DepartmentID: "dc", Type: "mass-casualty", Start: at(0, 0), End: at(5, 0), Count: 1, Source: "Illustrative demo requirement — frequency not supplied by the CSV", RepeatDays: 7},
	}

	return d
}
