package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newSpitCmd(c *Container) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spit \"<goal>\"",
		Short: "Generate a next coding session brief",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("spit: not yet implemented")
			return nil
		},
	}
	AddModelFlag(cmd)
	AddDetailLevelFlag(cmd)
	cmd.Flags().Bool("compact-only", false, "output compact prompt only")
	cmd.Flags().Bool("full", false, "output full detailed plan")
	return cmd
}
