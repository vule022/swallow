package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func newProjectCmd(c *Container) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage projects",
	}
	cmd.AddCommand(
		newProjectInitCmd(c),
		newProjectListCmd(c),
		newProjectUseCmd(c),
	)
	return cmd
}

func newProjectInitCmd(c *Container) *cobra.Command {
	var name, summary string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create a new project from the current directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireDB(c); err != nil {
				return err
			}
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("cannot determine current directory: %w", err)
			}
			if name == "" {
				// Default name from directory basename.
				name = strings.ToLower(strings.ReplaceAll(filepath.Base(cwd), " ", "-"))
			}

			p, err := c.Projects.Init(cmd.Context(), cwd, name, summary)
			if err != nil {
				return err
			}
			if err := c.Repos.Projects.SetActive(cmd.Context(), p.ID); err != nil {
				return err
			}
			fmt.Printf("Created project: %s\n", p.Name)
			fmt.Printf("  ID:       %s\n", p.ID)
			fmt.Printf("  Root:     %s\n", p.RootPath)
			fmt.Println("\nNext: run 'swallow ingest .' to index your codebase.")
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "project name (defaults to directory name)")
	cmd.Flags().StringVar(&summary, "summary", "", "short project description")
	return cmd
}

func newProjectListCmd(c *Container) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireDB(c); err != nil {
				return err
			}
			projects, err := c.Projects.List(cmd.Context())
			if err != nil {
				return err
			}
			if len(projects) == 0 {
				fmt.Println("No projects yet. Run 'swallow project init' to create one.")
				return nil
			}

			activeProject, _ := c.Projects.GetActive(cmd.Context())
			for _, p := range projects {
				marker := "  "
				if activeProject != nil && p.ID == activeProject.ID {
					marker = "* "
				}
				fmt.Printf("%s%s\n", marker, p.Name)
				if p.RootPath != "" {
					fmt.Printf("    root: %s\n", p.RootPath)
				}
				if p.Summary != "" {
					fmt.Printf("    %s\n", p.Summary)
				}
			}
			return nil
		},
	}
}

func newProjectUseCmd(c *Container) *cobra.Command {
	return &cobra.Command{
		Use:   "use <name-or-id>",
		Short: "Set the active project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireDB(c); err != nil {
				return err
			}
			p, err := c.Projects.Use(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			fmt.Printf("Active project set to: %s\n", p.Name)
			return nil
		},
	}
}

func requireDB(c *Container) error {
	if c.Repos == nil {
		return fmt.Errorf("swallow is not initialised. Run 'swallow init' first.")
	}
	return nil
}
