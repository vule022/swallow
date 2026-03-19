package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vule022/swallow/internal/export"
	"github.com/vule022/swallow/internal/model"
	"github.com/vule022/swallow/internal/render"
)

func newExportCmd(c *Container) *cobra.Command {
	var compact, full bool

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export a handoff document for another coding agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireDB(c); err != nil {
				return err
			}

			p, err := requireActiveProject(cmd, c)
			if err != nil {
				return err
			}

			exp := export.New(c.Repos)
			_, result, err := exp.Export(cmd.Context(), p.ID, export.Options{
				Compact: compact,
				Full:    full,
			})
			if err != nil {
				return err
			}

			r := render.New()
			r.PlanResult(result, compact)

			_ = recordSession(cmd.Context(), c, p.ID, model.SessionTypeExport,
				fmt.Sprintf("export for project '%s'", p.Name), "")

			return nil
		},
	}

	AddModelFlag(cmd)
	AddDetailLevelFlag(cmd)
	cmd.Flags().BoolVar(&compact, "compact", false, "compact export")
	cmd.Flags().BoolVar(&full, "full", false, "full export with all context")
	return cmd
}
