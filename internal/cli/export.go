package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newExportCmd(c *Container) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export a handoff document for another coding agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("export: not yet implemented")
			return nil
		},
	}
	AddModelFlag(cmd)
	AddDetailLevelFlag(cmd)
	cmd.Flags().Bool("compact", false, "compact export")
	cmd.Flags().Bool("full", false, "full export")
	return cmd
}
