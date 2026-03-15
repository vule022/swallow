package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/vule022/swallow/internal/model"
)

type OutputRepo struct {
	db *sql.DB
}

func newOutputRepo(db *sql.DB) *OutputRepo {
	return &OutputRepo{db: db}
}

func (r *OutputRepo) Save(ctx context.Context, o *model.CodingOutput) error {
	actions, _ := json.Marshal(o.Actions)
	files, _ := json.Marshal(o.Files)
	decisions, _ := json.Marshal(o.Decisions)
	blockers, _ := json.Marshal(o.Blockers)
	nextActions, _ := json.Marshal(o.NextActions)
	validationNotes, _ := json.Marshal(o.ValidationNotes)
	commitRecs, _ := json.Marshal(o.CommitRecommendations)

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO coding_outputs
		 (id, project_id, source, raw_text, goal, actions, files, decisions, blockers, next_actions, validation_notes, commit_recommendations, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		o.ID, o.ProjectID, o.Source, o.RawText, o.Goal,
		string(actions), string(files), string(decisions), string(blockers),
		string(nextActions), string(validationNotes), string(commitRecs),
		o.CreatedAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("output_repo.Save: %w", err)
	}
	return nil
}

func (r *OutputRepo) ListByProject(ctx context.Context, projectID string) ([]*model.CodingOutput, error) {
	return r.getRecent(ctx, projectID, 100)
}

func (r *OutputRepo) GetRecent(ctx context.Context, projectID string, limit int) ([]*model.CodingOutput, error) {
	return r.getRecent(ctx, projectID, limit)
}

func (r *OutputRepo) getRecent(ctx context.Context, projectID string, limit int) ([]*model.CodingOutput, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, project_id, source, raw_text, goal, actions, files, decisions, blockers, next_actions, validation_notes, commit_recommendations, created_at
		 FROM coding_outputs WHERE project_id = ? ORDER BY created_at DESC LIMIT ?`, projectID, limit)
	if err != nil {
		return nil, fmt.Errorf("output_repo.getRecent: %w", err)
	}
	defer rows.Close()

	var outputs []*model.CodingOutput
	for rows.Next() {
		o := &model.CodingOutput{}
		var actions, files, decisions, blockers, nextActions, validationNotes, commitRecs string
		if err := rows.Scan(&o.ID, &o.ProjectID, &o.Source, &o.RawText, &o.Goal,
			&actions, &files, &decisions, &blockers, &nextActions, &validationNotes, &commitRecs, &o.CreatedAt); err != nil {
			return nil, fmt.Errorf("output_repo: scan: %w", err)
		}
		_ = json.Unmarshal([]byte(actions), &o.Actions)
		_ = json.Unmarshal([]byte(files), &o.Files)
		_ = json.Unmarshal([]byte(decisions), &o.Decisions)
		_ = json.Unmarshal([]byte(blockers), &o.Blockers)
		_ = json.Unmarshal([]byte(nextActions), &o.NextActions)
		_ = json.Unmarshal([]byte(validationNotes), &o.ValidationNotes)
		_ = json.Unmarshal([]byte(commitRecs), &o.CommitRecommendations)
		outputs = append(outputs, o)
	}
	return outputs, rows.Err()
}
