package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/vule022/swallow/internal/model"
)

// DocumentSummary is a lightweight doc reference for context building.
type DocumentSummary struct {
	RelativePath string
	Summary      string
	Kind         string
}

// RecentContext is all recent context gathered for spit/export.
type RecentContext struct {
	Project           *model.Project
	RecentOutputs     []*model.CodingOutput
	RecentSessions    []*model.SessionEntry
	DocumentSummaries []DocumentSummary
}

// ContextOptions controls how much context is gathered.
type ContextOptions struct {
	MaxOutputs      int
	MaxSessions     int
	MaxDocSummaries int
}

// DefaultContextOptions returns sensible defaults for standard detail level.
func DefaultContextOptions() ContextOptions {
	return ContextOptions{
		MaxOutputs:      5,
		MaxSessions:     5,
		MaxDocSummaries: 8,
	}
}

type ContextRepo struct {
	db       *sql.DB
	projects *ProjectRepo
	outputs  *OutputRepo
	sessions *SessionRepo
}

func newContextRepo(db *sql.DB, projects *ProjectRepo, outputs *OutputRepo, sessions *SessionRepo) *ContextRepo {
	return &ContextRepo{db: db, projects: projects, outputs: outputs, sessions: sessions}
}

func (r *ContextRepo) FetchRecentContext(ctx context.Context, projectID string, opts ContextOptions) (*RecentContext, error) {
	project, err := r.projects.GetByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("context_repo: get project: %w", err)
	}

	outputs, err := r.outputs.GetRecent(ctx, projectID, opts.MaxOutputs)
	if err != nil {
		return nil, fmt.Errorf("context_repo: get outputs: %w", err)
	}

	sessions, err := r.sessions.GetRecent(ctx, projectID, opts.MaxSessions)
	if err != nil {
		return nil, fmt.Errorf("context_repo: get sessions: %w", err)
	}

	docSummaries, err := r.fetchDocSummaries(ctx, projectID, opts.MaxDocSummaries)
	if err != nil {
		return nil, fmt.Errorf("context_repo: get doc summaries: %w", err)
	}

	return &RecentContext{
		Project:           project,
		RecentOutputs:     outputs,
		RecentSessions:    sessions,
		DocumentSummaries: docSummaries,
	}, nil
}

func (r *ContextRepo) fetchDocSummaries(ctx context.Context, projectID string, limit int) ([]DocumentSummary, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT relative_path, summary, kind FROM documents
		 WHERE project_id = ? AND summary != ''
		 ORDER BY modified_at DESC LIMIT ?`, projectID, limit)
	if err != nil {
		return nil, fmt.Errorf("context_repo.fetchDocSummaries: %w", err)
	}
	defer rows.Close()

	var summaries []DocumentSummary
	for rows.Next() {
		var s DocumentSummary
		if err := rows.Scan(&s.RelativePath, &s.Summary, &s.Kind); err != nil {
			return nil, fmt.Errorf("context_repo: scan summary: %w", err)
		}
		summaries = append(summaries, s)
	}
	return summaries, rows.Err()
}
