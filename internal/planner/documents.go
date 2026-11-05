package planner

import "time"

// SourceReference keeps accepted records tied to the retained original document.
type SourceReference struct {
	DocumentID string `json:"document_id"`
	Page       int    `json:"page"`
	Excerpt    string `json:"excerpt"`
}

type TrainingSuggestion struct {
	TrainingActivity string    `json:"training_activity"`
	ID               string    `json:"id"`
	Kind             string    `json:"kind"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	Basis            string    `json:"basis"`
	DepartmentID     string    `json:"department_id"`
	Contact          string    `json:"contact"`
	Type             string    `json:"type"`
	Duration         int       `json:"duration"`
	DurationEvidence string    `json:"duration_evidence,omitempty"`
	Tags             []string  `json:"tags"`
	ConstraintIDs    []string  `json:"constraint_ids"`
	Fatigue          bool      `json:"fatigue"`
	Count            int       `json:"count"`
	RepeatDays       int       `json:"repeat_days"`
	Start            time.Time `json:"start"`
	End              time.Time `json:"end"`
	Page             int       `json:"page"`
	Excerpt          string    `json:"excerpt"`
	CitationStatus   string    `json:"citation_status,omitempty"`
	ReviewNotes      string    `json:"review_notes"`
	Missing          []string  `json:"missing"`
	Rule             *Rule     `json:"rule,omitempty"`
	CreatedID        string    `json:"created_id,omitempty"`
	Deleted          bool      `json:"deleted,omitempty"`
}

type TrainingDocument struct {
	ID          string               `json:"id"`
	Filename    string               `json:"filename"`
	Size        int                  `json:"size"`
	SHA256      string               `json:"sha256"`
	ImportedAt  time.Time            `json:"imported_at"`
	Status      string               `json:"status"`
	Error       string               `json:"error,omitempty"`
	Title       string               `json:"title"`
	Summary     string               `json:"summary"`
	Suggestions []TrainingSuggestion `json:"suggestions"`
}
