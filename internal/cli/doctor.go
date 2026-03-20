package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vule022/swallow/internal/config"
	"github.com/vule022/swallow/internal/render"
)

func newDoctorCmd(c *Container) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Verify swallow configuration and connectivity",
		RunE: func(cmd *cobra.Command, args []string) error {
			r := render.New()
			r.Header("SWALLOW DOCTOR")

			allOK := true
			check := func(label string, ok bool, detail string) {
				if ok {
					r.Success(label)
				} else {
					r.Error(label + ": " + detail)
					allOK = false
				}
			}

			// Config dir.
			dir, err := config.Dir()
			check("Config directory", err == nil, fmt.Sprintf("%v", err))
			if err == nil {
				_, statErr := os.Stat(dir)
				check("Config dir exists", statErr == nil, dir+" not found — run 'swallow init'")
			}

			// Config file.
			cfgPath, _ := config.Path()
			_, cfgErr := os.Stat(cfgPath)
			check("Config file", cfgErr == nil, cfgPath+" not found — run 'swallow init'")

			// Config values.
			check("Provider set", c.Config.Provider != "", "set provider in "+cfgPath)
			check("Model set", c.Config.Model != "", "set model in "+cfgPath)
			check("Base URL set", c.Config.BaseURL != "", "set base_url in "+cfgPath)

			// API key.
			apiKey := config.APIKey()
			check("API key ("+config.EnvAPIKey+")", apiKey != "",
				"export "+config.EnvAPIKey+"=sk-... to enable LLM features")

			// Database.
			dbPath, _ := config.DBPath()
			_, dbErr := os.Stat(dbPath)
			check("Database", dbErr == nil, dbPath+" not found — run 'swallow init'")

			// Active project.
			if c.Projects != nil {
				cwd, _ := os.Getwd()
				p, projErr := c.Projects.ResolveActive(cmd.Context(), cwd)
				if projErr == nil {
					check("Active project", true, "")
					r.Dim(fmt.Sprintf("  → %s (%s)", p.Name, p.RootPath))
				} else {
					r.Dim("  No active project — run 'swallow project use <name>'")
				}
			} else {
				r.Dim("  Database not initialised — run 'swallow init'")
			}

			// Config details.
			fmt.Println()
			r.Label("Provider", c.Config.Provider)
			r.Label("Model", c.Config.Model)
			r.Label("Base URL", c.Config.BaseURL)
			r.Label("Max tokens", fmt.Sprintf("%d", c.Config.MaxTokens))

			fmt.Println()
			if allOK {
				r.Success("All checks passed. Swallow is ready.")
			} else {
				r.Error("Some checks failed. See above for details.")
				return fmt.Errorf("doctor: one or more checks failed")
			}
			return nil
		},
	}
}
