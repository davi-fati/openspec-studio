package repository

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestProjectRepositoryRoundTrip(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "studio.db")

	db, err := Open(ctx, dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	repo := NewProjectRepository(db)

	openedAt := time.Now().Truncate(time.Second)
	if _, err := repo.Upsert(ctx, "demo", "/tmp/demo-project", openedAt); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	got, found, err := repo.GetByPath(ctx, "/tmp/demo-project")
	if err != nil {
		t.Fatalf("GetByPath: %v", err)
	}
	if !found {
		t.Fatalf("GetByPath: project not found after upsert")
	}
	if got.Name != "demo" || got.Path != "/tmp/demo-project" {
		t.Fatalf("GetByPath: unexpected project %+v", got)
	}

	// Re-upserting the same path updates in place rather than duplicating.
	laterOpenedAt := openedAt.Add(time.Hour)
	if _, err := repo.Upsert(ctx, "demo-renamed", "/tmp/demo-project", laterOpenedAt); err != nil {
		t.Fatalf("Upsert (update): %v", err)
	}

	list, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("List: expected 1 project, got %d", len(list))
	}
	if list[0].Name != "demo-renamed" {
		t.Fatalf("List: expected updated name, got %q", list[0].Name)
	}

	_, found, err = repo.GetByPath(ctx, "/tmp/does-not-exist")
	if err != nil {
		t.Fatalf("GetByPath (missing): %v", err)
	}
	if found {
		t.Fatalf("GetByPath (missing): expected not found")
	}
}
