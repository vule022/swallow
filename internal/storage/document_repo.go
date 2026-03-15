package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/vule022/swallow/internal/model"
)

type DocumentRepo struct {
	db *sql.DB
}

func newDocumentRepo(db *sql.DB) *DocumentRepo {
	return &DocumentRepo{db: db}
}

func (r *DocumentRepo) Save(ctx context.Context, d *model.Document) error {
	rawStored := 0
	if d.RawStored {
		rawStored = 1
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO documents
		 (id, project_id, path, relative_path, kind, hash, size, modified_at, summary, raw_stored, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		d.ID, d.ProjectID, d.Path, d.RelativePath, d.Kind, d.Hash,
		d.Size, d.ModifiedAt.UTC(), d.Summary, rawStored, d.CreatedAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("document_repo.Save: %w", err)
	}
	return nil
}

func (r *DocumentRepo) GetByPath(ctx context.Context, projectID, relativePath string) (*model.Document, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, project_id, path, relative_path, kind, hash, size, modified_at, summary, raw_stored, created_at
		 FROM documents WHERE project_id = ? AND relative_path = ?`, projectID, relativePath)
	return scanDocument(row)
}

func (r *DocumentRepo) ListByProject(ctx context.Context, projectID string) ([]*model.Document, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, project_id, path, relative_path, kind, hash, size, modified_at, summary, raw_stored, created_at
		 FROM documents WHERE project_id = ? ORDER BY relative_path ASC`, projectID)
	if err != nil {
		return nil, fmt.Errorf("document_repo.ListByProject: %w", err)
	}
	defer rows.Close()

	var docs []*model.Document
	for rows.Next() {
		d, err := scanDocumentRow(rows)
		if err != nil {
			return nil, err
		}
		docs = append(docs, d)
	}
	return docs, rows.Err()
}

func (r *DocumentRepo) CountByProject(ctx context.Context, projectID string) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM documents WHERE project_id = ?`, projectID).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("document_repo.CountByProject: %w", err)
	}
	return n, nil
}

func scanDocument(row *sql.Row) (*model.Document, error) {
	var d model.Document
	var rawStored int
	err := row.Scan(&d.ID, &d.ProjectID, &d.Path, &d.RelativePath, &d.Kind, &d.Hash,
		&d.Size, &d.ModifiedAt, &d.Summary, &rawStored, &d.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("document_repo: scan: %w", err)
	}
	d.RawStored = rawStored != 0
	return &d, nil
}

func scanDocumentRow(rows *sql.Rows) (*model.Document, error) {
	var d model.Document
	var rawStored int
	err := rows.Scan(&d.ID, &d.ProjectID, &d.Path, &d.RelativePath, &d.Kind, &d.Hash,
		&d.Size, &d.ModifiedAt, &d.Summary, &rawStored, &d.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("document_repo: scan row: %w", err)
	}
	d.RawStored = rawStored != 0
	return &d, nil
}
