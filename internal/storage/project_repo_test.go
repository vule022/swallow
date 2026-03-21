package storage

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/vule022/swallow/internal/model"
)

func testDB(t *testing.T) *DB {
	t.Helper()
	// Use a temp file so we can open it (modernc.org/sqlite requires a path or :memory:).
	f, err := os.CreateTemp(t.TempDir(), "test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	db, err := Open(f.Name())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestProjectCRUD(t *testing.T) {
	db := testDB(t)
	repo := newProjectRepo(db.db)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Second)
	p := &model.Project{
		ID:          "test-id-1",
		Name:        "myproject",
		RootPath:    "/tmp/myproject",
		Summary:     "A test project",
		Tags:        []string{"go", "cli"},
		ActiveGoals: []string{"finish v1"},
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Create.
	if err := repo.Create(ctx, p); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// GetByID.
	got, err := repo.GetByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != p.Name {
		t.Errorf("Name = %q, want %q", got.Name, p.Name)
	}
	if len(got.Tags) != 2 {
		t.Errorf("Tags len = %d, want 2", len(got.Tags))
	}

	// GetByName.
	got2, err := repo.GetByName(ctx, "myproject")
	if err != nil {
		t.Fatalf("GetByName: %v", err)
	}
	if got2.ID != p.ID {
		t.Errorf("ID = %q, want %q", got2.ID, p.ID)
	}

	// List.
	projects, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(projects) != 1 {
		t.Errorf("List len = %d, want 1", len(projects))
	}

	// NotFound.
	_, err = repo.GetByID(ctx, "nonexistent")
	if err != ErrNotFound {
		t.Errorf("GetByID nonexistent: expected ErrNotFound, got %v", err)
	}
}

func TestActiveProject(t *testing.T) {
	db := testDB(t)
	repo := newProjectRepo(db.db)
	ctx := context.Background()

	// No active project yet.
	_, err := repo.GetActive(ctx)
	if err != ErrNoActiveProject {
		t.Errorf("GetActive (none): expected ErrNoActiveProject, got %v", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	p := &model.Project{
		ID: "proj-1", Name: "p1", RootPath: "/p1",
		Tags: []string{}, ActiveGoals: []string{},
		CreatedAt: now, UpdatedAt: now,
	}
	if err := repo.Create(ctx, p); err != nil {
		t.Fatal(err)
	}

	if err := repo.SetActive(ctx, p.ID); err != nil {
		t.Fatalf("SetActive: %v", err)
	}

	active, err := repo.GetActive(ctx)
	if err != nil {
		t.Fatalf("GetActive: %v", err)
	}
	if active.ID != p.ID {
		t.Errorf("active.ID = %q, want %q", active.ID, p.ID)
	}
}
