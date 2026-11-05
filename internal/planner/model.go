package planner

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"time"
)

type Window struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}
type Department struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
	Goal  string `json:"goal"`
}
type Qualification struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type Person struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	DepartmentID   string   `json:"department_id"`
	Qualifications []string `json:"qualifications"`
	Availability   []Window `json:"availability"`
}
type Duty struct {
	ID       string    `json:"id"`
	PersonID string    `json:"person_id"`
	Name     string    `json:"name"`
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	Fatigue  bool      `json:"fatigue"`
}
type Asset struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Capacity     int      `json:"capacity"`
	Availability []Window `json:"availability"`
}
type Rule struct {
	SourceRef       *SourceReference `json:"source_ref,omitempty"`
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	Type            string           `json:"type"`
	Scope           string           `json:"scope"`
	Confirmed       bool             `json:"confirmed"`
	QualificationID string           `json:"qualification_id,omitempty"`
	PersonID        string           `json:"person_id,omitempty"`
	AssetID         string           `json:"asset_id,omitempty"`
	Count           int              `json:"count,omitempty"`
	Start           time.Time        `json:"start,omitempty"`
	End             time.Time        `json:"end,omitempty"`
	TargetTag       string           `json:"target_tag,omitempty"`
	TargetEventID   string           `json:"target_event_id,omitempty"`
	Minutes         int              `json:"minutes,omitempty"`
}

// EventTemplate contains reusable defaults, independent of any calendar date.
type EventTemplate struct {
	SourceRef     *SourceReference `json:"source_ref,omitempty"`
	Description   string           `json:"description,omitempty"`
	ID            string           `json:"id"`
	Name          string           `json:"name"`
	DepartmentID  string           `json:"department_id"`
	Tier          int              `json:"tier"`
	Contact       string           `json:"contact"`
	Type          string           `json:"type"`
	Tags          []string         `json:"tags"`
	Duration      int              `json:"duration"`
	ConstraintIDs []string         `json:"constraint_ids"`
	Fatigue       bool             `json:"fatigue"`
}

// UnmarshalJSON supplies a tier for templates saved before tiers were introduced.
func (t *EventTemplate) UnmarshalJSON(data []byte) error {
	type plain EventTemplate
	var v struct {
		plain
		Tier *int `json:"tier"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*t = EventTemplate(v.plain)
	t.Tier = SuggestedTier(t.Type)
	if v.Tier != nil {
		t.Tier = *v.Tier
	}
	return nil
}

func (t EventTemplate) Occurrence(start, windowStart, windowEnd time.Time) Event {
	return Event{ID: ID(), TemplateID: t.ID, SourceRef: t.SourceRef, Description: t.Description, Name: t.Name, DepartmentID: t.DepartmentID,
		Tier: t.Tier, Contact: t.Contact, Type: t.Type, Tags: append([]string{}, t.Tags...), Duration: t.Duration,
		ConstraintIDs: append([]string{}, t.ConstraintIDs...), Fatigue: t.Fatigue,
		Start: start, RequestedStart: start, WindowStart: windowStart, WindowEnd: windowEnd, Status: "draft"}
}

type Event struct {
	IgnoreConflicts            bool             `json:"ignore_conflicts,omitempty"`
	IgnoreConflictsDescription string           `json:"ignore_conflicts_description,omitempty"`
	SourceRef                  *SourceReference `json:"source_ref,omitempty"`
	Description                string           `json:"description,omitempty"`
	ID                         string           `json:"id"`
	TemplateID                 string           `json:"template_id,omitempty"`
	Name                       string           `json:"name"`
	DepartmentID               string           `json:"department_id"`
	Tier                       int              `json:"tier"`
	Contact                    string           `json:"contact"`
	Type                       string           `json:"type"`
	Tags                       []string         `json:"tags"`
	Start                      time.Time        `json:"start"`
	RequestedStart             time.Time        `json:"requested_start"`
	Duration                   int              `json:"duration"`
	WindowStart                time.Time        `json:"window_start"`
	WindowEnd                  time.Time        `json:"window_end"`
	Fixed                      bool             `json:"fixed"`
	Status                     string           `json:"status"`
	ConstraintIDs              []string         `json:"constraint_ids"`
	Fatigue                    bool             `json:"fatigue"`
}

// UnmarshalJSON also defaults tiers in legacy events and saved proposal snapshots.
func (e *Event) UnmarshalJSON(data []byte) error {
	type plain Event
	var v struct {
		plain
		Tier *int `json:"tier"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*e = Event(v.plain)
	e.Tier = SuggestedTier(e.Type)
	if v.Tier != nil {
		e.Tier = *v.Tier
	}
	return nil
}

func (e Event) End() time.Time { return e.Start.Add(time.Duration(e.Duration) * time.Minute) }

type Requirement struct {
	SourceRef    *SourceReference `json:"source_ref,omitempty"`
	ID           string           `json:"id"`
	Name         string           `json:"name"`
	DepartmentID string           `json:"department_id"`
	Type         string           `json:"type"`
	Start        time.Time        `json:"start"`
	End          time.Time        `json:"end"`
	Count        int              `json:"count"`
	Source       string           `json:"source"`
	RepeatDays   int              `json:"repeat_days"`
}
type Completion struct {
	ID           string    `json:"id"`
	DepartmentID string    `json:"department_id"`
	Type         string    `json:"type"`
	At           time.Time `json:"at"`
	EventID      string    `json:"event_id,omitempty"`
}
type Assignment struct {
	EventID  string `json:"event_id"`
	PersonID string `json:"person_id"`
	Role     string `json:"role"`
}
type Data struct {
	Departments    []Department    `json:"departments"`
	Qualifications []Qualification `json:"qualifications"`
	People         []Person        `json:"people"`
	Duties         []Duty          `json:"duties"`
	Assets         []Asset         `json:"assets"`
	Constraints    []Rule          `json:"constraints"`
	Events         []Event         `json:"events"`
	Templates      []EventTemplate `json:"templates"`
	Requirements   []Requirement   `json:"requirements"`
	Completions    []Completion    `json:"completions"`
	Assignments    []Assignment    `json:"assignments"`
}
type Conflict struct {
	Type        string    `json:"type"`
	RuleID      string    `json:"rule_id,omitempty"`
	EventIDs    []string  `json:"event_ids"`
	ResourceIDs []string  `json:"resource_ids"`
	Start       time.Time `json:"start"`
	End         time.Time `json:"end"`
	Message     string    `json:"message"`
}
type Coverage struct {
	Requirement Requirement `json:"requirement"`
	Completed   int         `json:"completed"`
	Planned     int         `json:"planned"`
	Gap         int         `json:"gap"`
}
type Evaluation struct {
	Feasible    bool         `json:"feasible"`
	Conflicts   []Conflict   `json:"conflicts"`
	Warnings    []Conflict   `json:"warnings"`
	Assignments []Assignment `json:"assignments"`
	Coverage    []Coverage   `json:"coverage"`
}
type Change struct {
	EventID string     `json:"event_id"`
	Name    string     `json:"name"`
	Before  *time.Time `json:"before"`
	After   time.Time  `json:"after"`
	Reason  string     `json:"reason"`
}
type Preference struct {
	EventID string    `json:"event_id"`
	Start   time.Time `json:"start"`
	Reason  string    `json:"reason"`
}
type DepartmentProposal struct {
	ActionableConflicts []Conflict   `json:"actionable_conflicts,omitempty"`
	DepartmentID        string       `json:"department_id"`
	Round               int          `json:"round"`
	Summary             string       `json:"summary"`
	Preferences         []Preference `json:"preferences"`
	Evaluation          Evaluation   `json:"evaluation"`
}
type Proposal struct {
	TargetEventID       string               `json:"target_event_id,omitempty"`
	ID                  string               `json:"id"`
	Week                time.Time            `json:"week"`
	Status              string               `json:"status"`
	Progress            string               `json:"progress"`
	Mode                string               `json:"mode"`
	Snapshot            Data                 `json:"snapshot"`
	Events              []Event              `json:"events"`
	Evaluation          Evaluation           `json:"evaluation"`
	Changes             []Change             `json:"changes"`
	PreviousAssignments []Assignment         `json:"previous_assignments"`
	Negotiation         []DepartmentProposal `json:"negotiation"`
	Unscheduled         []Event              `json:"unscheduled"`
	Summary             string               `json:"summary"`
	Score               [3]int               `json:"score"`
	Alternate           bool                 `json:"alternate"`
	Error               string               `json:"error,omitempty"`
}

func ID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
func Week(t time.Time, loc *time.Location) time.Time {
	t = t.In(loc)
	day := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
	return day.AddDate(0, 0, -(int(day.Weekday())+6)%7)
}
func Overlap(a, b, c, d time.Time) bool { return a.Before(d) && c.Before(b) }
