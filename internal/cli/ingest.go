package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/vule022/swallow/internal/ingest"
	"github.com/vule022/swallow/internal/model"
	"github.com/vule022/swallow/internal/session"
	"github.com/vule022/swallow/internal/storage"
)

func newIngestCmd(c *Container) *cobra.Command {
	return &cobra.Command{
		Use:   "ingest <path>",
		Short: "Ingest a file or directory into the active project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireDB(c); err != nil {
				return err
			}
			cwd, _ := os.Getwd()
			p, err := c.Projects.ResolveActive(cmd.Context(), cwd)
			if err != nil {
				return fmt.Errorf("no active project: run 'swallow project use <name>'")
			}

			ing := ingest.New(c.Repos)
			result, err := ing.IngestPath(cmd.Context(), p.ID, args[0], ingest.Options{})
			if err != nil {
				return fmt.Errorf("ingest failed: %w", err)
			}

			fmt.Printf("Ingested into project '%s':\n", p.Name)
			fmt.Printf("  new: %d  updated: %d  skipped: %d\n", result.New, result.Updated, result.Skipped)
			if len(result.Errors) > 0 {
				fmt.Printf("  errors: %d\n", len(result.Errors))
				for _, e := range result.Errors {
					fmt.Printf("    - %s\n", e)
				}
			}

			sessionType := model.SessionTypeFolderIngest
			if result.New+result.Updated == 1 {
				sessionType = model.SessionTypeFileIngest
			}
			_ = recordSession(cmd.Context(), c, p.ID, sessionType,
				fmt.Sprintf("Ingested %s: %d new, %d updated, %d skipped", args[0], result.New, result.Updated, result.Skipped),
				"")
			return nil
		},
	}
}

func newIngestOutputCmd(c *Container) *cobra.Command {
	var fromStdin bool

	cmd := &cobra.Command{
		Use:   "ingest-output [path]",
		Short: "Ingest a coding agent output as structured context",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireDB(c); err != nil {
				return err
			}
			p, err := requireActiveProject(cmd, c)
			if err != nil {
				return err
			}

			var rawText string
			var source string

			if fromStdin {
				data, err := io.ReadAll(bufio.NewReader(os.Stdin))
				if err != nil {
					return fmt.Errorf("reading stdin: %w", err)
				}
				rawText = string(data)
				source = "stdin"
			} else {
				if len(args) == 0 {
					return fmt.Errorf("provide a path or use --stdin")
				}
				data, err := os.ReadFile(args[0])
				if err != nil {
					return fmt.Errorf("read file: %w", err)
				}
				rawText = string(data)
				source = args[0]
			}

			if rawText == "" {
				return fmt.Errorf("input is empty")
			}

			ctx := cmd.Context()
			output, entry, err := ingestRawText(ctx, c, p.ID, source, rawText)
			if err != nil {
				return err
			}

			fmt.Printf("Saved coding output from '%s' to project '%s'\n", source, p.Name)
			fmt.Printf("  ID: %s\n", output.ID)
			if output.Goal != "" {
				fmt.Printf("  Goal: %s\n", output.Goal)
			}
			if len(output.NextActions) > 0 {
				fmt.Printf("  Next actions: %d extracted\n", len(output.NextActions))
			}
			if entry != nil && entry.NextAction != "" {
				fmt.Printf("  Suggested next: %s\n", entry.NextAction)
			}
			fmt.Println("\n  Tip: run 'swallow spit' to incorporate this into your next session brief.")

			return nil
		},
	}

	cmd.Flags().BoolVar(&fromStdin, "stdin", false, "read from stdin")
	return cmd
}

// ingestRawText runs the full ingest-output pipeline on rawText.
// It is the single canonical place for building CodingOutput, normalizing,
// saving, and recording a session. Called by ingest-output, watch, and hooks run.
func ingestRawText(ctx context.Context, c *Container, projectID, source, rawText string) (*model.CodingOutput, *model.SessionEntry, error) {
	output := &model.CodingOutput{
		ID:                    uuid.New().String(),
		ProjectID:             projectID,
		Source:                source,
		RawText:               rawText,
		Actions:               []string{},
		Files:                 []string{},
		Decisions:             []string{},
		Blockers:              []string{},
		NextActions:           []string{},
		ValidationNotes:       []string{},
		CommitRecommendations: []string{},
		CreatedAt:             time.Now().UTC(),
	}

	sessMgr := session.New(c.Repos, c.Planner)
	if err := sessMgr.NormalizeOutput(ctx, output); err != nil {
		fmt.Fprintf(os.Stderr, "warning: normalization failed: %v\n", err)
	}

	if err := c.Repos.Outputs.Save(ctx, output); err != nil {
		return nil, nil, fmt.Errorf("save output: %w", err)
	}

	entry, err := sessMgr.RecordFromOutput(ctx, output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: record session: %v\n", err)
	}

	return output, entry, nil
}

func recordSession(ctx context.Context, c *Container, projectID, sessionType, summary, nextAction string) error {
	entry := &model.SessionEntry{
		ID:            uuid.New().String(),
		ProjectID:     projectID,
		Type:          sessionType,
		Summary:       summary,
		RelatedFiles:  []string{},
		Decisions:     []string{},
		OpenQuestions: []string{},
		NextAction:    nextAction,
		CreatedAt:     time.Now().UTC(),
	}
	return c.Repos.Sessions.Save(ctx, entry)
}

func requireActiveProject(cmd *cobra.Command, c *Container) (*model.Project, error) {
	cwd, _ := os.Getwd()
	p, err := c.Projects.ResolveActive(cmd.Context(), cwd)
	if err != nil {
		if err == storage.ErrNoActiveProject {
			return nil, fmt.Errorf("no active project. Run 'swallow project use <name>'")
		}
		return nil, err
	}
	return p, nil
}
