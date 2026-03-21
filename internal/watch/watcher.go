package watch

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// IngestFunc is called for each new file. filePath is the absolute path.
type IngestFunc func(ctx context.Context, filePath string) error

// Options configures a Watcher.
type Options struct {
	Dir      string   // directory to watch (required)
	Exts     []string // file extensions to accept, e.g. []string{".md", ".txt"}
	OnIngest IngestFunc
	OnError  func(err error)
}

// DefaultExts are the file extensions accepted by default.
var DefaultExts = []string{".md", ".txt", ".json", ".log"}

// Watcher watches a directory and ingests new files.
type Watcher struct {
	opts Options
	seen map[string]struct{}
	mu   sync.Mutex
	fsw  *fsnotify.Watcher
}

// New creates a Watcher and ensures processed/ subdir exists.
func New(opts Options) (*Watcher, error) {
	if opts.Dir == "" {
		return nil, fmt.Errorf("watch: Dir is required")
	}
	if opts.OnIngest == nil {
		return nil, fmt.Errorf("watch: OnIngest is required")
	}
	if len(opts.Exts) == 0 {
		opts.Exts = DefaultExts
	}
	if opts.OnError == nil {
		opts.OnError = func(err error) { fmt.Fprintf(os.Stderr, "watch error: %v\n", err) }
	}

	if err := os.MkdirAll(processedDir(opts.Dir), 0o700); err != nil {
		return nil, fmt.Errorf("watch: create processed dir: %w", err)
	}

	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("watch: create fsnotify watcher: %w", err)
	}

	return &Watcher{opts: opts, seen: make(map[string]struct{}), fsw: fsw}, nil
}

// Run starts watching and blocks until ctx is cancelled.
func (w *Watcher) Run(ctx context.Context) error {
	defer w.fsw.Close()

	if err := w.fsw.Add(w.opts.Dir); err != nil {
		return fmt.Errorf("watch: add dir %s: %w", w.opts.Dir, err)
	}

	// Process files already in the inbox before we started.
	if err := w.scanExisting(ctx); err != nil {
		w.opts.OnError(fmt.Errorf("initial scan: %w", err))
	}

	for {
		select {
		case <-ctx.Done():
			return nil

		case event, ok := <-w.fsw.Events:
			if !ok {
				return nil
			}
			if event.Has(fsnotify.Create) {
				// Settle delay: give the writer time to finish flushing.
				time.Sleep(100 * time.Millisecond)
				if err := w.processFile(ctx, event.Name); err != nil {
					w.opts.OnError(err)
				}
			}

		case err, ok := <-w.fsw.Errors:
			if !ok {
				return nil
			}
			w.opts.OnError(err)
		}
	}
}

func (w *Watcher) scanExisting(ctx context.Context) error {
	return filepath.WalkDir(w.opts.Dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if d.IsDir() {
			if path == w.opts.Dir {
				return nil
			}
			return fs.SkipDir // don't descend into processed/ or any subdir
		}
		return w.processFile(ctx, path)
	})
}

func (w *Watcher) processFile(ctx context.Context, path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	// Skip files inside processed/.
	if filepath.Dir(abs) == processedDir(w.opts.Dir) {
		return nil
	}

	if !w.isSupported(abs) {
		return nil
	}

	if !w.markSeen(abs) {
		return nil // already processed
	}

	if err := w.opts.OnIngest(ctx, abs); err != nil {
		return fmt.Errorf("ingest %s: %w", filepath.Base(abs), err)
	}

	if err := moveToProcessed(abs, w.opts.Dir); err != nil {
		// Non-fatal: file is already ingested, just couldn't move it.
		w.opts.OnError(fmt.Errorf("move to processed: %w", err))
	}
	return nil
}

func (w *Watcher) isSupported(path string) bool {
	ext := filepath.Ext(path)
	for _, e := range w.opts.Exts {
		if ext == e {
			return true
		}
	}
	return false
}

// markSeen records path. Returns false if already seen.
func (w *Watcher) markSeen(path string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	if _, ok := w.seen[path]; ok {
		return false
	}
	w.seen[path] = struct{}{}
	return true
}

// processedDir returns the processed/ subdirectory path.
func processedDir(inboxDir string) string {
	return filepath.Join(inboxDir, "processed")
}

// moveToProcessed renames src into processedDir with a timestamp suffix.
func moveToProcessed(src, inboxDir string) error {
	base := filepath.Base(src)
	ext := filepath.Ext(base)
	stem := base[:len(base)-len(ext)]
	dst := filepath.Join(processedDir(inboxDir),
		fmt.Sprintf("%s-%d%s", stem, time.Now().UnixNano(), ext))
	return os.Rename(src, dst)
}
