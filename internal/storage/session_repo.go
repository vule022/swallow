package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/vule022/swallow/internal/model"
)

type SessionRepo struct {
	db *sql.DB
}

func newSessionRepo(db *sql.DB) *SessionRepo {
	return &SessionRepo{db: db}
}

func (r *SessionRepo) Save(ctx context.Context, e *model.SessionEntry) error {
	relFiles, _ := json.Marshal(e.RelatedFiles)
	decisions, _ := json.Marshal(e.Decisions)
	openQs, _ := json.Marshal(e.OpenQuestions)

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO session_entries
		 (id, project_id, type, summary, related_files, decisions, open_questions, next_action, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.ID, e.ProjectID, e.Type, e.Summary,
		string(relFiles), string(decisions), string(openQs),
		e.NextAction, e.CreatedAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("session_repo.Save: %w", err)
	}
	return nil
}

func (r *SessionRepo) ListByProject(ctx context.Context, projectID string) ([]*model.SessionEntry, error) {
	return r.getRecent(ctx, projectID, 100)
}

func (r *SessionRepo) GetRecent(ctx context.Context, projectID string, limit int) ([]*model.SessionEntry, error) {
	return r.getRecent(ctx, projectID, limit)
}

func (r *SessionRepo) getRecent(ctx context.Context, projectID string, limit int) ([]*model.SessionEntry, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, project_id, type, summary, related_files, decisions, open_questions, next_action, created_at
		 FROM session_entries WHERE project_id = ? ORDER BY created_at DESC LIMIT ?`, projectID, limit)
	if err != nil {
		return nil, fmt.Errorf("session_repo.getRecent: %w", err)
	}
	defer rows.Close()

	var entries []*model.SessionEntry
	for rows.Next() {
		e := &model.SessionEntry{}
		var relFiles, decisions, openQs string
		if err := rows.Scan(&e.ID, &e.ProjectID, &e.Type, &e.Summary,
			&relFiles, &decisions, &openQs, &e.NextAction, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("session_repo: scan: %w", err)
		}
		_ = json.Unmarshal([]byte(relFiles), &e.RelatedFiles)
		_ = json.Unmarshal([]byte(decisions), &e.Decisions)
		_ = json.Unmarshal([]byte(openQs), &e.OpenQuestions)
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
