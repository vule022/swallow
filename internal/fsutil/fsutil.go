package fsutil

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// WalkResult holds information about a file discovered during a walk.
type WalkResult struct {
	Path string
	Size int64
	Kind string
}

// WalkFiles recursively walks root, returning files that pass the filter.
func WalkFiles(root string, filter Filter) ([]WalkResult, error) {
	var results []WalkResult

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			// Skip files/dirs we can't access.
			if os.IsPermission(walkErr) {
				return nil
			}
			return walkErr
		}

		// Skip symlinks to avoid loops.
		if d.Type()&fs.ModeSymlink != 0 {
			return nil
		}

		if d.IsDir() {
			if path == root {
				return nil
			}
			if !filter.ShouldIncludeDir(d.Name()) {
				return fs.SkipDir
			}
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil // skip unreadable files
		}

		if !filter.ShouldIncludeFile(path, info.Size()) {
			return nil
		}

		results = append(results, WalkResult{
			Path: path,
			Size: info.Size(),
			Kind: DetectKind(path),
		})
		return nil
	})

	return results, err
}

// HashFile computes the SHA-256 hash of a file's contents.
func HashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("fsutil.HashFile: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("fsutil.HashFile: %w", err)
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// RelPath returns the relative path of abs from root.
// Falls back to abs if rel computation fails.
func RelPath(root, abs string) string {
	rel, err := filepath.Rel(root, abs)
	if err != nil {
		return abs
	}
	return rel
}
