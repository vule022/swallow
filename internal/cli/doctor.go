package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDoctorCmd(c *Container) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Verify swallow configuration and connectivity",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("doctor: not yet implemented")
			return nil
		},
	}
}
