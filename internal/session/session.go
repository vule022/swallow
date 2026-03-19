package session

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/vule022/swallow/internal/model"
	"github.com/vule022/swallow/internal/planner"
	"github.com/vule022/swallow/internal/storage"
)

// Manager handles session recording and output normalization.
type Manager struct {
	repos   *storage.Repos
	planner planner.Planner
}

// New creates a new session Manager.
func New(repos *storage.Repos, p planner.Planner) *Manager {
	return &Manager{repos: repos, planner: p}
}

// RecordFromOutput creates a session entry from a CodingOutput.
func (m *Manager) RecordFromOutput(ctx context.Context, output *model.CodingOutput) (*model.SessionEntry, error) {
	summary := output.Goal
	if summary == "" && len(output.Actions) > 0 {
		summary = output.Actions[0]
	}
	if summary == "" {
		summary = fmt.Sprintf("coding output from %s", output.Source)
	}

	nextAction := ""
	if len(output.NextActions) > 0 {
		nextAction = output.NextActions[0]
	}

	entry := &model.SessionEntry{
		ID:            uuid.New().String(),
		ProjectID:     output.ProjectID,
		Type:          model.SessionTypeCodingOutputIngest,
		Summary:       summary,
		RelatedFiles:  output.Files,
		Decisions:     output.Decisions,
		OpenQuestions: output.Blockers,
		NextAction:    nextAction,
		CreatedAt:     time.Now().UTC(),
	}

	if err := m.repos.Sessions.Save(ctx, entry); err != nil {
		return nil, fmt.Errorf("session: save: %w", err)
	}
	return entry, nil
}

// NormalizeOutput uses the planner to extract structured data from raw text.
func (m *Manager) NormalizeOutput(ctx context.Context, output *model.CodingOutput) error {
	result, err := m.planner.Compress(ctx, output.RawText)
	if err != nil {
		// Non-fatal: leave output as raw text.
		return nil
	}

	if output.Goal == "" {
		output.Goal = result.Goal
	}
	if len(result.Actions) > 0 && len(output.Actions) == 0 {
		output.Actions = result.Actions
	}
	if len(result.Decisions) > 0 && len(output.Decisions) == 0 {
		output.Decisions = result.Decisions
	}
	if len(result.Blockers) > 0 && len(output.Blockers) == 0 {
		output.Blockers = result.Blockers
	}
	if len(result.NextActions) > 0 && len(output.NextActions) == 0 {
		output.NextActions = result.NextActions
	}
	if len(result.FilesMentioned) > 0 && len(output.Files) == 0 {
		output.Files = result.FilesMentioned
	}
	return nil
}
