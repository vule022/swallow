package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/vule022/swallow/internal/config"
	"github.com/vule022/swallow/internal/storage"
)

func newInitCmd(c *Container) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialise swallow configuration and storage",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := config.Dir()
			if err != nil {
				return err
			}
			if err := os.MkdirAll(dir, 0o700); err != nil {
				return fmt.Errorf("cannot create config dir: %w", err)
			}

			cfgPath, _ := config.Path()
			if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
				if err := config.Save(config.Defaults()); err != nil {
					return err
				}
				fmt.Printf("Created config: %s\n", cfgPath)
			} else {
				fmt.Printf("Config already exists: %s\n", cfgPath)
			}

			dbPath := filepath.Join(dir, config.DBFileName)
			db, err := storage.Open(dbPath)
			if err != nil {
				return fmt.Errorf("cannot initialise database: %w", err)
			}
			db.Close()
			fmt.Printf("Database ready: %s\n", dbPath)

			fmt.Println("\nSwallow is ready. Next:")
			fmt.Println("  cd /path/to/your/project")
			fmt.Println("  swallow project init --name <name>")
			fmt.Printf("\nSet your API key:  export %s=sk-...\n", config.EnvAPIKey)
			return nil
		},
	}
}
