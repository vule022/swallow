package ingest

import (
	"context"
	"os"
	"path/filepath"

	"github.com/vule022/swallow/internal/storage"
)

// Options controls ingestion behaviour.
type Options struct {
	Summarize bool
	MaxSizeMB int
}

// Result summarises an ingestion run.
type Result struct {
	New     int
	Updated int
	Skipped int
	Errors  []string
}

// Ingester handles file and directory ingestion.
type Ingester struct {
	repos *storage.Repos
}

// New creates a new Ingester.
func New(repos *storage.Repos) *Ingester {
	return &Ingester{repos: repos}
}

// IngestPath ingests a single file or an entire directory.
func (i *Ingester) IngestPath(ctx context.Context, projectID, path string, opts Options) (*Result, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		return i.ingestDir(ctx, projectID, absPath, opts)
	}
	result := &Result{}
	if err := i.ingestFile(ctx, projectID, absPath, absPath, info.Size(), opts, result); err != nil {
		result.Errors = append(result.Errors, err.Error())
	}
	return result, nil
}
