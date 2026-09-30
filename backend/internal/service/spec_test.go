package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/davidlima/openspec-studio/backend/internal/domain"
)

// stubProjectLister is a minimal in-memory SpecProjectLister for tests, so
// spec/change service tests don't need a real SQLite-backed registry.
type stubProjectLister struct {
	projects []domain.Project
}

func (s *stubProjectLister) ListProjects(ctx context.Context) ([]domain.Project, error) {
	return s.projects, nil
}

func (s *stubProjectLister) GetProject(ctx context.Context, id int64) (domain.Project, error) {
	for _, p := range s.projects {
		if p.ID == id {
			return p, nil
		}
	}
	return domain.Project{}, ErrProjectNotFound
}

func newTestProject(t *testing.T, id int64, name string) domain.Project {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "openspec", "specs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "openspec", "changes", "archive"), 0o755); err != nil {
		t.Fatal(err)
	}
	return domain.Project{ID: id, Name: name, Path: dir, Available: true}
}

func TestSpecServiceCreateUpdateDelete(t *testing.T) {
	ctx := context.Background()
	proj := newTestProject(t, 1, "demo")
	lister := &stubProjectLister{projects: []domain.Project{proj}}
	svc := NewSpecService(lister, "test-author")

	reqs := []domain.SpecRequirement{{
		Name: "Do the thing",
		Body: "The system SHALL do the thing.\n\n#### Scenario: It does the thing\n- **WHEN** triggered\n- **THEN** it happens",
	}}

	created, err := svc.CreateSpec(ctx, 1, "widgets", "Widgets capability for testing.", reqs)
	if err != nil {
		t.Fatalf("CreateSpec: %v", err)
	}
	if created.Capability != "widgets" || len(created.Requirements) != 1 {
		t.Fatalf("unexpected created spec: %+v", created)
	}

	// Creating again at the same capability must fail.
	if _, err := svc.CreateSpec(ctx, 1, "widgets", "dup", reqs); err != ErrSpecExists {
		t.Fatalf("expected ErrSpecExists, got %v", err)
	}

	updated, err := svc.UpdateSpec(ctx, 1, "widgets", "Updated purpose.", []domain.SpecRequirement{{
		Name: "Do the thing better",
		Body: "The system SHALL do the thing better.\n\n#### Scenario: Better\n- **WHEN** triggered\n- **THEN** it happens better",
	}})
	if err != nil {
		t.Fatalf("UpdateSpec: %v", err)
	}
	if updated.Purpose != "Updated purpose." || updated.Requirements[0].Name != "Do the thing better" {
		t.Fatalf("update did not persist: %+v", updated)
	}

	if _, err := svc.DeleteSpec(ctx, 1, "widgets", false); err != nil {
		t.Fatalf("DeleteSpec: %v", err)
	}
	if _, err := svc.GetSpec(ctx, 1, "widgets"); err != ErrSpecNotFound {
		t.Fatalf("expected spec to be gone, got %v", err)
	}
}

func TestSpecServiceProvenanceRoundTrip(t *testing.T) {
	ctx := context.Background()
	proj := newTestProject(t, 1, "demo")
	lister := &stubProjectLister{projects: []domain.Project{proj}}
	svc := NewSpecService(lister, "alice")

	created, err := svc.CreateSpec(ctx, 1, "auth", "Handles auth.", []domain.SpecRequirement{{
		Name: "Login", Body: "The system SHALL allow login.\n\n#### Scenario: Login works\n- **WHEN** valid creds\n- **THEN** session starts",
	}})
	if err != nil {
		t.Fatalf("CreateSpec: %v", err)
	}
	if created.Meta.Provenance.ID == "" {
		t.Fatal("expected a generated provenance id")
	}
	if created.Meta.Provenance.Author != "alice" {
		t.Fatalf("expected author alice, got %q", created.Meta.Provenance.Author)
	}
	if created.Meta.Provenance.Project != "demo" {
		t.Fatalf("expected project demo, got %q", created.Meta.Provenance.Project)
	}

	// Reload from disk (fresh ListSpecs call) and verify provenance survived.
	reloaded, err := svc.GetSpec(ctx, 1, "auth")
	if err != nil {
		t.Fatalf("GetSpec: %v", err)
	}
	if reloaded.Meta.Provenance.ID != created.Meta.Provenance.ID {
		t.Fatalf("provenance id did not round-trip: got %q want %q", reloaded.Meta.Provenance.ID, created.Meta.Provenance.ID)
	}
	if reloaded.Meta.Provenance.CreatedAt.IsZero() {
		t.Fatal("expected a non-zero created timestamp after reload")
	}
}

func TestSpecServiceCrossProjectDependencies(t *testing.T) {
	ctx := context.Background()
	projA := newTestProject(t, 1, "project-a")
	projB := newTestProject(t, 2, "project-b")
	lister := &stubProjectLister{projects: []domain.Project{projA, projB}}
	svc := NewSpecService(lister, "author")

	if _, err := svc.CreateSpec(ctx, 1, "base", "Base capability.", []domain.SpecRequirement{{
		Name: "Base req", Body: "The system SHALL provide the base.\n\n#### Scenario: Base works\n- **WHEN** used\n- **THEN** it works",
	}}); err != nil {
		t.Fatalf("CreateSpec project A: %v", err)
	}
	if _, err := svc.CreateSpec(ctx, 2, "consumer", "Consumer capability.", []domain.SpecRequirement{{
		Name: "Consumer req", Body: "The system SHALL consume the base.\n\n#### Scenario: Consumes\n- **WHEN** used\n- **THEN** it consumes",
	}}); err != nil {
		t.Fatalf("CreateSpec project B: %v", err)
	}

	if err := svc.AddDependency(ctx, 2, "consumer", 1, "base"); err != nil {
		t.Fatalf("AddDependency: %v", err)
	}

	consumer, err := svc.GetSpec(ctx, 2, "consumer")
	if err != nil {
		t.Fatalf("GetSpec consumer: %v", err)
	}
	if len(consumer.DependsOn) != 1 || !consumer.DependsOn[0].Available {
		t.Fatalf("expected one available cross-project dependency, got %+v", consumer.DependsOn)
	}
	if consumer.DependsOn[0].ProjectName != "project-a" {
		t.Fatalf("expected dependency on project-a, got %q", consumer.DependsOn[0].ProjectName)
	}

	base, err := svc.GetSpec(ctx, 1, "base")
	if err != nil {
		t.Fatalf("GetSpec base: %v", err)
	}
	if len(base.DependedBy) != 1 || base.DependedBy[0].Capability != "consumer" {
		t.Fatalf("expected base to be depended on by consumer, got %+v", base.DependedBy)
	}

	// Deleting a spec with dependents must be blocked without force.
	if _, err := svc.DeleteSpec(ctx, 1, "base", false); err != ErrSpecHasDependents {
		t.Fatalf("expected ErrSpecHasDependents, got %v", err)
	}
}

func TestSpecServiceRejectsSelfDependency(t *testing.T) {
	ctx := context.Background()
	proj := newTestProject(t, 1, "demo")
	lister := &stubProjectLister{projects: []domain.Project{proj}}
	svc := NewSpecService(lister, "author")

	if _, err := svc.CreateSpec(ctx, 1, "widgets", "Widgets.", []domain.SpecRequirement{{
		Name: "Req", Body: "The system SHALL work.\n\n#### Scenario: Works\n- **WHEN** used\n- **THEN** it works",
	}}); err != nil {
		t.Fatalf("CreateSpec: %v", err)
	}

	if err := svc.AddDependency(ctx, 1, "widgets", 1, "widgets"); err != ErrSelfDependency {
		t.Fatalf("expected ErrSelfDependency, got %v", err)
	}
}
