package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/vule022/swallow/internal/model"
	"github.com/vule022/swallow/internal/planner"
	"github.com/vule022/swallow/internal/prompt"
	"github.com/vule022/swallow/internal/render"
	"github.com/vule022/swallow/internal/storage"
)

func newSpitCmd(c *Container) *cobra.Command {
	var compactOnly, full bool

	cmd := &cobra.Command{
		Use:   "spit \"<goal>\"",
		Short: "Generate a next coding session brief",
		Long: `Generate a structured next-session plan based on your project context.

The goal can be vague or in any language:
  swallow spit "fix auth flow"
  swallow spit "cleanup ingest and storage"
  swallow spit "add better export"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireDB(c); err != nil {
				return err
			}

			p, err := requireActiveProject(cmd, c)
			if err != nil {
				return err
			}

			goal := args[0]

			// Resolve detail level.
			levelStr := GetDetailLevelStr(cmd)
			if compactOnly && levelStr == "" {
				levelStr = "compact"
			} else if full && levelStr == "" {
				levelStr = "detailed"
			}
			level, err := prompt.Parse(levelStr)
			if err != nil {
				return err
			}

			// Fetch context from DB.
			maxOutputs, maxSessions, maxDocs := level.ContextLimits()
			ctx := cmd.Context()
			recentCtx, err := c.Repos.Context.FetchRecentContext(ctx, p.ID, storage.ContextOptions{
				MaxOutputs:      maxOutputs,
				MaxSessions:     maxSessions,
				MaxDocSummaries: maxDocs,
			})
			if err != nil {
				return fmt.Errorf("fetch context: %w", err)
			}

			modelOverride := GetModelOverride(cmd)

			fmt.Fprintf(os.Stderr, "Generating session brief for project '%s'...\n", p.Name)

			result, err := c.Planner.Plan(ctx, planner.PlanRequest{
				Goal:          goal,
				Context:       recentCtx,
				DetailLevel:   level,
				ModelOverride: modelOverride,
			})
			if err != nil {
				return err
			}

			r := render.New()
			r.PlanResult(result, compactOnly)

			// Record session.
			_ = recordSession(ctx, c, p.ID, model.SessionTypeSpit,
				fmt.Sprintf("spit: %s → %s", goal, result.Title),
				"",
			)

			// Save the spit result as a session entry with the plan details.
			entry := &model.SessionEntry{
				ID:            uuid.New().String(),
				ProjectID:     p.ID,
				Type:          model.SessionTypeSpit,
				Summary:       fmt.Sprintf("%s: %s", result.Title, result.Goal),
				RelatedFiles:  extractFilePaths(result.RelevantFiles),
				Decisions:     []string{},
				OpenQuestions: []string{},
				NextAction:    firstOrEmpty(result.ExecutionPlan),
				CreatedAt:     time.Now().UTC(),
			}
			_ = c.Repos.Sessions.Save(ctx, entry)

			return nil
		},
	}

	AddModelFlag(cmd)
	AddDetailLevelFlag(cmd)
	cmd.Flags().BoolVar(&compactOnly, "compact-only", false, "output compact prompt only (no structured sections)")
	cmd.Flags().BoolVar(&full, "full", false, "output full detailed plan")
	return cmd
}

func extractFilePaths(refs []model.FileReference) []string {
	paths := make([]string, 0, len(refs))
	for _, r := range refs {
		paths = append(paths, r.Path)
	}
	return paths
}

func firstOrEmpty(items []string) string {
	if len(items) > 0 {
		return items[0]
	}
	return ""
}
