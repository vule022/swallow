package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/vule022/swallow/internal/model"
)

// ProjectRepo handles project persistence.
type ProjectRepo struct {
	db *sql.DB
}

func newProjectRepo(db *sql.DB) *ProjectRepo {
	return &ProjectRepo{db: db}
}

func (r *ProjectRepo) Create(ctx context.Context, p *model.Project) error {
	tags, _ := json.Marshal(p.Tags)
	goals, _ := json.Marshal(p.ActiveGoals)

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO projects (id, name, root_path, summary, tags, active_goals, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.Name, p.RootPath, p.Summary, string(tags), string(goals),
		p.CreatedAt.UTC(), p.UpdatedAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("project_repo.Create: %w", err)
	}
	return nil
}

func (r *ProjectRepo) GetByID(ctx context.Context, id string) (*model.Project, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, root_path, summary, tags, active_goals, created_at, updated_at
		 FROM projects WHERE id = ?`, id)
	return scanProject(row)
}

func (r *ProjectRepo) GetByName(ctx context.Context, name string) (*model.Project, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, root_path, summary, tags, active_goals, created_at, updated_at
		 FROM projects WHERE name = ?`, name)
	return scanProject(row)
}

func (r *ProjectRepo) List(ctx context.Context) ([]*model.Project, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, root_path, summary, tags, active_goals, created_at, updated_at
		 FROM projects ORDER BY created_at ASC`)
	if err != nil {
		return nil, fmt.Errorf("project_repo.List: %w", err)
	}
	defer rows.Close()

	var projects []*model.Project
	for rows.Next() {
		p, err := scanProjectRow(rows)
		if err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func (r *ProjectRepo) Update(ctx context.Context, p *model.Project) error {
	tags, _ := json.Marshal(p.Tags)
	goals, _ := json.Marshal(p.ActiveGoals)
	p.UpdatedAt = time.Now().UTC()

	res, err := r.db.ExecContext(ctx,
		`UPDATE projects SET name=?, root_path=?, summary=?, tags=?, active_goals=?, updated_at=?
		 WHERE id=?`,
		p.Name, p.RootPath, p.Summary, string(tags), string(goals), p.UpdatedAt, p.ID,
	)
	if err != nil {
		return fmt.Errorf("project_repo.Update: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ProjectRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM projects WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("project_repo.Delete: %w", err)
	}
	return nil
}

func (r *ProjectRepo) SetActive(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO project_settings (key, value) VALUES ('active_project_id', ?)`, id)
	if err != nil {
		return fmt.Errorf("project_repo.SetActive: %w", err)
	}
	return nil
}

func (r *ProjectRepo) GetActive(ctx context.Context) (*model.Project, error) {
	var id string
	err := r.db.QueryRowContext(ctx,
		`SELECT value FROM project_settings WHERE key = 'active_project_id'`).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) || id == "" {
		return nil, ErrNoActiveProject
	}
	if err != nil {
		return nil, fmt.Errorf("project_repo.GetActive: %w", err)
	}
	return r.GetByID(ctx, id)
}

func scanProject(row *sql.Row) (*model.Project, error) {
	var p model.Project
	var tagsJSON, goalsJSON string
	err := row.Scan(&p.ID, &p.Name, &p.RootPath, &p.Summary, &tagsJSON, &goalsJSON, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("project_repo: scan: %w", err)
	}
	_ = json.Unmarshal([]byte(tagsJSON), &p.Tags)
	_ = json.Unmarshal([]byte(goalsJSON), &p.ActiveGoals)
	if p.Tags == nil {
		p.Tags = []string{}
	}
	if p.ActiveGoals == nil {
		p.ActiveGoals = []string{}
	}
	return &p, nil
}

func scanProjectRow(rows *sql.Rows) (*model.Project, error) {
	var p model.Project
	var tagsJSON, goalsJSON string
	err := rows.Scan(&p.ID, &p.Name, &p.RootPath, &p.Summary, &tagsJSON, &goalsJSON, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("project_repo: scan row: %w", err)
	}
	_ = json.Unmarshal([]byte(tagsJSON), &p.Tags)
	_ = json.Unmarshal([]byte(goalsJSON), &p.ActiveGoals)
	if p.Tags == nil {
		p.Tags = []string{}
	}
	if p.ActiveGoals == nil {
		p.ActiveGoals = []string{}
	}
	return &p, nil
}
