package planner

import (
	"context"

	"github.com/vule022/swallow/internal/model"
	"github.com/vule022/swallow/internal/prompt"
	"github.com/vule022/swallow/internal/storage"
)

// Planner generates session plans and document summaries.
type Planner interface {
	Plan(ctx context.Context, req PlanRequest) (*model.PlanResult, error)
	Compress(ctx context.Context, text string) (*CompressResult, error)
}

// PlanRequest is the input to the Plan method.
type PlanRequest struct {
	Goal          string
	Context       *storage.RecentContext
	DetailLevel   prompt.DetailLevel
	ModelOverride string
}

// CompressResult holds the structured output from Compress.
type CompressResult struct {
	Summary        string
	Goal           string
	Actions        []string
	Decisions      []string
	Blockers       []string
	NextActions    []string
	FilesMentioned []string
}
