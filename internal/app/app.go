package app

import (
	"context"
	"fmt"
	"os"

	"github.com/vule022/swallow/internal/cli"
	"github.com/vule022/swallow/internal/config"
	"github.com/vule022/swallow/internal/project"
	"github.com/vule022/swallow/internal/storage"
)

// Run is the main entry point. Returns exit code.
func Run() int {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	dbPath, err := config.DBPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	container := &cli.Container{
		Config: cfg,
	}

	// DB may not exist yet (before `swallow init`).
	if _, statErr := os.Stat(dbPath); statErr == nil {
		db, openErr := storage.Open(dbPath)
		if openErr != nil {
			fmt.Fprintf(os.Stderr, "error opening database: %v\n", openErr)
			return 1
		}
		defer db.Close()
		repos := storage.NewRepos(db)
		container.DB = db
		container.Repos = repos
		container.Projects = project.New(repos)
	}

	rootCmd := cli.NewRootCmd(container)
	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		return 1
	}
	return 0
}
