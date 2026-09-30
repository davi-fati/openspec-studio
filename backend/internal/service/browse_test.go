package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/davidlima/openspec-studio/backend/internal/domain"
)

// stubProjectServiceRepo is a minimal in-memory ProjectRepository for
// ProjectService tests.
type stubProjectServiceRepo struct {
	upserted []domain.Project
	nextID   int64
}

func newStubProjectServiceRepo() *stubProjectServiceRepo {
	return &stubProjectServiceRepo{}
}

func (r *stubProjectServiceRepo) Upsert(ctx context.Context, name, path string, openedAt time.Time) (domain.Project, error) {
	r.nextID++
	p := domain.Project{ID: r.nextID, Name: name, Path: path, LastOpenedAt: openedAt, Available: true}
	r.upserted = append(r.upserted, p)
	return p, nil
}

func (r *stubProjectServiceRepo) List(ctx context.Context) ([]domain.Project, error) {
	return r.upserted, nil
}

func (r *stubProjectServiceRepo) GetByPath(ctx context.Context, path string) (domain.Project, bool, error) {
	for _, p := range r.upserted {
		if p.Path == path {
			return p, true, nil
		}
	}
	return domain.Project{}, false, nil
}

func (r *stubProjectServiceRepo) GetByID(ctx context.Context, id int64) (domain.Project, bool, error) {
	for _, p := range r.upserted {
		if p.ID == id {
			return p, true, nil
		}
	}
	return domain.Project{}, false, nil
}

func TestHasOpenSpecStructure(t *testing.T) {
	valid := newTestProject(t, 1, "valid")
	if !hasOpenSpecStructure(valid.Path) {
		t.Fatal("expected a scaffolded project directory to be detected as valid")
	}

	invalid := t.TempDir()
	if hasOpenSpecStructure(invalid) {
		t.Fatal("expected a plain empty directory to not be detected as an OpenSpec project")
	}

	partial := t.TempDir()
	if err := os.MkdirAll(filepath.Join(partial, "openspec", "specs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if hasOpenSpecStructure(partial) {
		t.Fatal("expected a directory with only openspec/specs/ (missing changes/) to not count as valid")
	}
}

func TestProjectServiceBrowseDirectory(t *testing.T) {
	root := t.TempDir()

	// A valid OpenSpec project subdirectory.
	projectDir := filepath.Join(root, "my-project")
	if err := os.MkdirAll(filepath.Join(projectDir, "openspec", "specs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(projectDir, "openspec", "changes"), 0o755); err != nil {
		t.Fatal(err)
	}

	// An empty directory.
	emptyDir := filepath.Join(root, "empty-dir")
	if err := os.MkdirAll(emptyDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// A non-empty directory that is not an OpenSpec project.
	junkDir := filepath.Join(root, "junk-dir")
	if err := os.MkdirAll(junkDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(junkDir, "notes.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}

	svc := &ProjectService{}
	result, err := svc.BrowseDirectory(root)
	if err != nil {
		t.Fatalf("BrowseDirectory: %v", err)
	}
	if result.Path != root {
		t.Fatalf("expected resolved path %q, got %q", root, result.Path)
	}
	if result.Parent != filepath.Dir(root) {
		t.Fatalf("expected parent %q, got %q", filepath.Dir(root), result.Parent)
	}

	byName := map[string]BrowseEntry{}
	for _, e := range result.Entries {
		byName[e.Name] = e
	}
	if len(byName) != 3 {
		t.Fatalf("expected 3 entries, got %d: %+v", len(byName), result.Entries)
	}

	if p := byName["my-project"]; !p.HasOpenspec || p.IsEmpty {
		t.Fatalf("expected my-project flagged hasOpenspec=true isEmpty=false, got %+v", p)
	}
	if e := byName["empty-dir"]; e.HasOpenspec || !e.IsEmpty {
		t.Fatalf("expected empty-dir flagged hasOpenspec=false isEmpty=true, got %+v", e)
	}
	if j := byName["junk-dir"]; j.HasOpenspec || j.IsEmpty {
		t.Fatalf("expected junk-dir flagged hasOpenspec=false isEmpty=false, got %+v", j)
	}

	// The current directory itself (root, which has 3 subdirs and is not an
	// OpenSpec project) must also be flagged, not just its children - the
	// picker needs this to gate "use this folder" on where the user is now.
	if result.HasOpenspec {
		t.Fatal("expected root itself to not be flagged as an OpenSpec project")
	}
	if result.IsEmpty {
		t.Fatal("expected root itself to not be flagged empty (it has 3 subdirs)")
	}

	// Browsing into the project dir itself must flag IT as hasOpenspec, not
	// just report it as an entry of its parent.
	projectResult, err := svc.BrowseDirectory(projectDir)
	if err != nil {
		t.Fatalf("BrowseDirectory(projectDir): %v", err)
	}
	if !projectResult.HasOpenspec {
		t.Fatal("expected browsing into the project dir to flag it hasOpenspec=true")
	}

	// Browsing into the empty dir itself must flag IT as isEmpty.
	emptyResult, err := svc.BrowseDirectory(emptyDir)
	if err != nil {
		t.Fatalf("BrowseDirectory(emptyDir): %v", err)
	}
	if !emptyResult.IsEmpty {
		t.Fatal("expected browsing into the empty dir to flag it isEmpty=true")
	}
}

func TestProjectServiceBrowseDirectoryDefaultsToHome(t *testing.T) {
	svc := &ProjectService{}
	result, err := svc.BrowseDirectory("")
	if err != nil {
		t.Fatalf("BrowseDirectory: %v", err)
	}
	home, _ := os.UserHomeDir()
	if result.Path != home {
		t.Fatalf("expected default browse path to be home dir %q, got %q", home, result.Path)
	}
}

func TestProjectServiceOpenProjectRejectsNonOpenSpecDir(t *testing.T) {
	ctx := context.Background()
	repo := newStubProjectServiceRepo()
	svc := NewProjectService(repo)

	dir := t.TempDir()
	if _, err := svc.OpenProject(ctx, dir); err != ErrNoOpenSpecProject {
		t.Fatalf("expected ErrNoOpenSpecProject, got %v", err)
	}
	if len(repo.upserted) != 0 {
		t.Fatalf("expected no registration on rejection, got %+v", repo.upserted)
	}

	// Still accepts a directory that IS a valid project.
	valid := t.TempDir()
	if err := os.MkdirAll(filepath.Join(valid, "openspec", "specs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(valid, "openspec", "changes"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.OpenProject(ctx, valid); err != nil {
		t.Fatalf("expected a valid OpenSpec project to still be opened, got %v", err)
	}
	if len(repo.upserted) != 1 {
		t.Fatalf("expected the valid project to be registered, got %+v", repo.upserted)
	}
}
