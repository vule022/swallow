package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vule022/swallow/internal/config"
	"github.com/vule022/swallow/internal/planner"
	"github.com/vule022/swallow/internal/project"
	"github.com/vule022/swallow/internal/storage"
)

type contextKey string

const containerKey contextKey = "container"

// Container holds all wired dependencies accessible to CLI commands.
type Container struct {
	Config   *config.Config
	Repos    *storage.Repos
	Projects *project.Manager
	DB       *storage.DB
	Planner  planner.Planner
}

// NewRootCmd builds the root cobra command tree.
func NewRootCmd(c *Container) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "swallow",
		Short: "Plan your next coding session with context-aware briefs",
		Long: `Swallow is a project-based CLI tool that helps you continue software
development sessions more effectively.

It ingests your codebase and coding agent outputs, then uses an LLM to
generate a structured next-session brief ready to paste into another coding agent.

Run 'swallow init' to get started.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDefault(cmd, c)
		},
	}

	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		cmd.SetContext(context.WithValue(cmd.Context(), containerKey, c))
		return nil
	}

	cmd.AddCommand(
		newInitCmd(c),
		newProjectCmd(c),
		newStatusCmd(c),
		newIngestCmd(c),
		newIngestOutputCmd(c),
		newSpitCmd(c),
		newExportCmd(c),
		newDoctorCmd(c),
		newWatchCmd(c),
		newHooksCmd(c),
	)

	return cmd
}

func runDefault(cmd *cobra.Command, c *Container) error {
	if c.Projects == nil {
		fmt.Println("Swallow is not initialised. Run 'swallow init' to get started.")
		return nil
	}

	cwd, _ := os.Getwd()
	p, err := c.Projects.ResolveActive(cmd.Context(), cwd)
	if err != nil {
		cmd.Help()
		fmt.Println("\nTip: run 'swallow project init' to create a project in the current directory.")
		return nil
	}

	fmt.Printf("Active project: %s\n", p.Name)
	fmt.Println("Run 'swallow status' for details, or 'swallow spit \"<goal>\"' to plan your next session.")
	return nil
}
