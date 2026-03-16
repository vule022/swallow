package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newIngestCmd(c *Container) *cobra.Command {
	return &cobra.Command{
		Use:   "ingest <path>",
		Short: "Ingest a file or directory into the active project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("ingest: not yet implemented")
			return nil
		},
	}
}

func newIngestOutputCmd(c *Container) *cobra.Command {
	var fromStdin bool
	cmd := &cobra.Command{
		Use:   "ingest-output [path]",
		Short: "Ingest a coding agent output file",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("ingest-output: not yet implemented")
			return nil
		},
	}
	cmd.Flags().BoolVar(&fromStdin, "stdin", false, "read from stdin")
	return cmd
}
