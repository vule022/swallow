package ingest

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/vule022/swallow/internal/fsutil"
	"github.com/vule022/swallow/internal/model"
	"github.com/vule022/swallow/internal/storage"
)

func (i *Ingester) ingestFile(ctx context.Context, projectID, root, path string, size int64, opts Options, result *Result) error {
	hash, err := fsutil.HashFile(path)
	if err != nil {
		return fmt.Errorf("ingest: hash %s: %w", path, err)
	}

	relPath := fsutil.RelPath(root, path)
	if relPath == "." || relPath == path {
		relPath = filepath.Base(path)
	}

	kind := fsutil.DetectKind(path)

	existing, err := i.repos.Documents.GetByPath(ctx, projectID, relPath)
	if err != nil && err != storage.ErrNotFound {
		return fmt.Errorf("ingest: lookup %s: %w", relPath, err)
	}

	if existing != nil && existing.Hash == hash {
		result.Skipped++
		return nil
	}

	stat, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("ingest: stat %s: %w", path, err)
	}

	doc := &model.Document{
		ProjectID:    projectID,
		Path:         path,
		RelativePath: relPath,
		Kind:         kind,
		Hash:         hash,
		Size:         size,
		ModifiedAt:   stat.ModTime().UTC(),
		CreatedAt:    time.Now().UTC(),
	}

	if existing != nil {
		doc.ID = existing.ID
		doc.CreatedAt = existing.CreatedAt
		result.Updated++
	} else {
		doc.ID = uuid.New().String()
		result.New++
	}

	return i.repos.Documents.Save(ctx, doc)
}
