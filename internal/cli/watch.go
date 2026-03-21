package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/vule022/swallow/internal/config"
	"github.com/vule022/swallow/internal/watch"
)

func newWatchCmd(c *Container) *cobra.Command {
	var (
		dir     string
		project string
	)

	cmd := &cobra.Command{
		Use:   "watch [--dir <path>]",
		Short: "Watch the inbox and auto-ingest new files from any coding agent",
		Long: `Watch a directory for new files and automatically ingest them.

Drop any coding agent output (markdown, text, JSON) into the inbox:

  ~/.swallow/inbox/

Swallow will ingest it immediately, extract structure via LLM, and make
it available for the next 'swallow spit' call.

Set up agent integrations automatically:
  swallow hooks install --agent claude-code`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireDB(c); err != nil {
				return err
			}

			// Resolve watch directory.
			if dir == "" {
				cfgDir, err := config.Dir()
				if err != nil {
					return fmt.Errorf("resolve config dir: %w", err)
				}
				dir = filepath.Join(cfgDir, config.InboxDirName)
			}
			if err := os.MkdirAll(filepath.Join(dir, config.ProcessedDirName), 0o700); err != nil {
				return fmt.Errorf("create inbox dirs: %w", err)
			}

			// Resolve project.
			projectID, projectName, err := resolveWatchProject(cmd.Context(), c, project)
			if err != nil {
				return err
			}

			// Signal-aware context so Ctrl+C stops the watcher cleanly.
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			w, err := watch.New(watch.Options{
				Dir: dir,
				OnIngest: func(watchCtx context.Context, filePath string) error {
					return watchIngestFile(watchCtx, c, projectID, projectName, filePath)
				},
				OnError: func(e error) {
					fmt.Fprintf(os.Stderr, "watch error: %v\n", e)
				},
			})
			if err != nil {
				return err
			}

			fmt.Printf("Watching %s for project '%s'\n", dir, projectName)
			fmt.Println("Drop any agent output file into the inbox. Press Ctrl+C to stop.")
			return w.Run(ctx)
		},
	}

	cmd.Flags().StringVar(&dir, "dir", "", "directory to watch (default: ~/.swallow/inbox/)")
	cmd.Flags().StringVar(&project, "project", "", "project name or ID (default: auto-detect)")
	return cmd
}

func resolveWatchProject(ctx context.Context, c *Container, nameOrID string) (id, name string, err error) {
	if nameOrID != "" {
		p, e := c.Projects.Use(ctx, nameOrID)
		if e != nil {
			return "", "", e
		}
		return p.ID, p.Name, nil
	}

	cwd, _ := os.Getwd()
	p, e := c.Projects.AutoDetect(ctx, cwd)
	if e != nil {
		p, e = c.Projects.GetActive(ctx)
		if e != nil {
			return "", "", fmt.Errorf("no project resolved — pass --project <name> or cd into a project directory")
		}
	}
	return p.ID, p.Name, nil
}

func watchIngestFile(ctx context.Context, c *Container, projectID, projectName, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read %s: %w", filePath, err)
	}
	output, _, err := ingestRawText(ctx, c, projectID, filePath, string(data))
	if err != nil {
		return err
	}
	msg := fmt.Sprintf("ingested: %s → project %s", filepath.Base(filePath), projectName)
	if output.Goal != "" {
		msg += fmt.Sprintf(" (%s)", output.Goal)
	}
	fmt.Println(msg)
	return nil
}
