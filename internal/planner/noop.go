package planner

import (
	"context"

	"github.com/vule022/swallow/internal/model"
)

// NoopPlanner returns empty results. Used when no API key is configured.
type NoopPlanner struct{}

func (n *NoopPlanner) Plan(_ context.Context, req PlanRequest) (*model.PlanResult, error) {
	return &model.PlanResult{
		Title:           "Session Brief",
		Goal:            req.Goal,
		WhyNow:          "",
		CurrentContext:  []string{},
		RelevantFiles:   []model.FileReference{},
		ExecutionPlan:   []string{},
		Constraints:     []string{},
		Validation:      []string{},
		ExpectedOutput:  "",
		CopyReadyPrompt: req.Goal,
	}, nil
}

func (n *NoopPlanner) Compress(_ context.Context, _ string) (*CompressResult, error) {
	return &CompressResult{
		Actions:        []string{},
		Decisions:      []string{},
		Blockers:       []string{},
		NextActions:    []string{},
		FilesMentioned: []string{},
	}, nil
}
