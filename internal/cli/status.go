package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newStatusCmd(c *Container) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show active project status",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireDB(c); err != nil {
				return err
			}
			cwd, _ := os.Getwd()
			p, err := c.Projects.ResolveActive(cmd.Context(), cwd)
			if err != nil {
				return fmt.Errorf("no active project. Run 'swallow project use <name>' or cd into a project directory")
			}

			fmt.Printf("Project:  %s\n", p.Name)
			fmt.Printf("ID:       %s\n", p.ID)
			fmt.Printf("Root:     %s\n", p.RootPath)
			if p.Summary != "" {
				fmt.Printf("Summary:  %s\n", p.Summary)
			}

			if len(p.ActiveGoals) > 0 {
				fmt.Println("\nActive goals:")
				for _, g := range p.ActiveGoals {
					fmt.Printf("  - %s\n", g)
				}
			}

			docCount, _ := c.Repos.Documents.CountByProject(cmd.Context(), p.ID)
			fmt.Printf("\nIndexed files: %d\n", docCount)

			sessions, _ := c.Repos.Sessions.GetRecent(cmd.Context(), p.ID, 3)
			if len(sessions) > 0 {
				fmt.Println("\nRecent sessions:")
				for _, s := range sessions {
					fmt.Printf("  [%s] %s\n", s.Type, s.Summary)
				}
			}
			return nil
		},
	}
}
