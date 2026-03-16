package project

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/vule022/swallow/internal/model"
	"github.com/vule022/swallow/internal/storage"
)

// Manager handles project lifecycle operations.
type Manager struct {
	repos *storage.Repos
}

// New creates a new project Manager.
func New(repos *storage.Repos) *Manager {
	return &Manager{repos: repos}
}

// Init creates a new project rooted at rootPath.
func (m *Manager) Init(ctx context.Context, rootPath, name, summary string) (*model.Project, error) {
	if name == "" {
		return nil, fmt.Errorf("project: name is required")
	}

	now := time.Now().UTC()
	p := &model.Project{
		ID:          uuid.New().String(),
		Name:        name,
		RootPath:    rootPath,
		Summary:     summary,
		Tags:        []string{},
		ActiveGoals: []string{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := m.repos.Projects.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("project.Init: %w", err)
	}
	return p, nil
}

// List returns all projects.
func (m *Manager) List(ctx context.Context) ([]*model.Project, error) {
	return m.repos.Projects.List(ctx)
}

// Use sets the active project by name or ID.
func (m *Manager) Use(ctx context.Context, nameOrID string) (*model.Project, error) {
	// Try by ID first.
	p, err := m.repos.Projects.GetByID(ctx, nameOrID)
	if err != nil && err != storage.ErrNotFound {
		return nil, fmt.Errorf("project.Use: %w", err)
	}
	if p == nil {
		// Try by name.
		p, err = m.repos.Projects.GetByName(ctx, nameOrID)
		if err != nil {
			return nil, fmt.Errorf("project.Use: project %q not found", nameOrID)
		}
	}

	if err := m.repos.Projects.SetActive(ctx, p.ID); err != nil {
		return nil, fmt.Errorf("project.Use: %w", err)
	}
	return p, nil
}

// GetActive returns the currently active project.
func (m *Manager) GetActive(ctx context.Context) (*model.Project, error) {
	return m.repos.Projects.GetActive(ctx)
}

// AutoDetect finds a project whose root path is a prefix of cwd.
// If exactly one project exists, it is returned regardless of cwd.
func (m *Manager) AutoDetect(ctx context.Context, cwd string) (*model.Project, error) {
	projects, err := m.repos.Projects.List(ctx)
	if err != nil {
		return nil, err
	}
	if len(projects) == 0 {
		return nil, storage.ErrNotFound
	}
	if len(projects) == 1 {
		return projects[0], nil
	}

	// Find longest matching root path.
	var best *model.Project
	var bestLen int
	for _, p := range projects {
		if p.RootPath == "" {
			continue
		}
		root := p.RootPath
		if !strings.HasSuffix(root, "/") {
			root += "/"
		}
		if strings.HasPrefix(cwd+"/", root) && len(root) > bestLen {
			best = p
			bestLen = len(root)
		}
	}
	if best != nil {
		return best, nil
	}
	return nil, storage.ErrNotFound
}

// ResolveActive returns the active project, auto-detecting from cwd if needed.
func (m *Manager) ResolveActive(ctx context.Context, cwd string) (*model.Project, error) {
	// Try explicitly set active project first.
	p, err := m.repos.Projects.GetActive(ctx)
	if err == nil {
		return p, nil
	}
	if err != storage.ErrNoActiveProject {
		return nil, err
	}

	// Fall back to auto-detect by cwd.
	return m.AutoDetect(ctx, cwd)
}
