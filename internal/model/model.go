package model

import "time"

// Project represents a tracked software project.
type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	RootPath    string    `json:"root_path"`
	Summary     string    `json:"summary"`
	Tags        []string  `json:"tags"`
	ActiveGoals []string  `json:"active_goals"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Document represents an ingested file or document.
type Document struct {
	ID           string    `json:"id"`
	ProjectID    string    `json:"project_id"`
	Path         string    `json:"path"`
	RelativePath string    `json:"relative_path"`
	Kind         string    `json:"kind"`
	Hash         string    `json:"hash"`
	Size         int64     `json:"size"`
	ModifiedAt   time.Time `json:"modified_at"`
	Summary      string    `json:"summary"`
	RawStored    bool      `json:"raw_stored"`
	CreatedAt    time.Time `json:"created_at"`
}

// CodingOutput represents an ingested coding agent output.
type CodingOutput struct {
	ID                    string    `json:"id"`
	ProjectID             string    `json:"project_id"`
	Source                string    `json:"source"`
	RawText               string    `json:"raw_text"`
	Goal                  string    `json:"goal"`
	Actions               []string  `json:"actions"`
	Files                 []string  `json:"files"`
	Decisions             []string  `json:"decisions"`
	Blockers              []string  `json:"blockers"`
	NextActions           []string  `json:"next_actions"`
	ValidationNotes       []string  `json:"validation_notes"`
	CommitRecommendations []string  `json:"commit_recommendations"`
	CreatedAt             time.Time `json:"created_at"`
}

// SessionEntry records a history entry for a project.
type SessionEntry struct {
	ID            string    `json:"id"`
	ProjectID     string    `json:"project_id"`
	Type          string    `json:"type"`
	Summary       string    `json:"summary"`
	RelatedFiles  []string  `json:"related_files"`
	Decisions     []string  `json:"decisions"`
	OpenQuestions []string  `json:"open_questions"`
	NextAction    string    `json:"next_action"`
	CreatedAt     time.Time `json:"created_at"`
}

// SessionType constants.
const (
	SessionTypeFolderIngest       = "folder_ingest"
	SessionTypeFileIngest         = "file_ingest"
	SessionTypeCodingOutputIngest = "coding_output_ingest"
	SessionTypeSpit               = "spit"
	SessionTypeExport             = "export"
)

// PlanResult is the structured output from the spit/export planning step.
type PlanResult struct {
	Title           string          `json:"title"`
	Goal            string          `json:"goal"`
	WhyNow          string          `json:"why_now"`
	CurrentContext  []string        `json:"current_context"`
	RelevantFiles   []FileReference `json:"relevant_files"`
	ExecutionPlan   []string        `json:"execution_plan"`
	Constraints     []string        `json:"constraints"`
	Validation      []string        `json:"validation"`
	ExpectedOutput  string          `json:"expected_output"`
	CopyReadyPrompt string          `json:"copy_ready_prompt"`
}

// FileReference is a file path with a reason for its relevance.
type FileReference struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
}
