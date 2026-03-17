package ingest

import (
	"context"
	"fmt"

	"github.com/vule022/swallow/internal/fsutil"
)

func (i *Ingester) ingestDir(ctx context.Context, projectID, root string, opts Options) (*Result, error) {
	filter := fsutil.DefaultFilter()
	if opts.MaxSizeMB > 0 {
		filter.MaxSizeBytes = int64(opts.MaxSizeMB) * 1024 * 1024
	}

	files, err := fsutil.WalkFiles(root, filter)
	if err != nil {
		return nil, fmt.Errorf("ingest: walk %s: %w", root, err)
	}

	result := &Result{}
	for _, f := range files {
		if err := i.ingestFile(ctx, projectID, root, f.Path, f.Size, opts, result); err != nil {
			result.Errors = append(result.Errors, err.Error())
		}
	}
	return result, nil
}
