package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/davidlima/openspec-studio/backend/internal/domain"
)

// writeChange creates a change directory with a proposal.md and a tasks.md
// whose checkbox counts drive domain.ComputeStatus, optionally placed under
// changes/archive/ to force the Archived status.
func writeChange(t *testing.T, projectDir, name string, archived bool, tasksTotal, tasksDone int) {
	t.Helper()
	root := filepath.Join(projectDir, "openspec", "changes")
	if archived {
		root = filepath.Join(root, "archive")
	}
	dir := filepath.Join(root, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "proposal.md"), []byte("## Why\n\ntest\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var tasks string
	for i := 0; i < tasksTotal; i++ {
		box := "[ ]"
		if i < tasksDone {
			box = "[x]"
		}
		tasks += fmt.Sprintf("- %s task %d\n", box, i+1)
	}
	if err := os.WriteFile(filepath.Join(dir, "tasks.md"), []byte(tasks), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestOverviewServiceGetOverview(t *testing.T) {
	ctx := context.Background()
	projA := newTestProject(t, 1, "project-a")
	projB := newTestProject(t, 2, "project-b")
	lister := &stubProjectLister{projects: []domain.Project{projA, projB}}

	specs := NewSpecService(lister, "author")
	if _, err := specs.CreateSpec(ctx, 1, "widgets", "Widgets capability.", []domain.SpecRequirement{{
		Name: "Req", Body: "The system SHALL work.\n\n#### Scenario: Works\n- **WHEN** used\n- **THEN** it works",
	}}); err != nil {
		t.Fatalf("CreateSpec project-a: %v", err)
	}

	changes := NewChangeService(lister)
	overview := NewOverviewService(lister, specs, changes)

	t.Run("unfiltered includes every project", func(t *testing.T) {
		result, err := overview.GetOverview(ctx, nil)
		if err != nil {
			t.Fatalf("GetOverview: %v", err)
		}
		if len(result.Projects) != 2 {
			t.Fatalf("expected 2 projects, got %d: %+v", len(result.Projects), result.Projects)
		}
		if len(result.RecentlyUpdated) != 2 {
			t.Fatalf("expected 2 recently-updated entries, got %+v", result.RecentlyUpdated)
		}
	})

	t.Run("filtered scopes to a single project", func(t *testing.T) {
		id := int64(1)
		result, err := overview.GetOverview(ctx, &id)
		if err != nil {
			t.Fatalf("GetOverview: %v", err)
		}
		if len(result.Projects) != 1 {
			t.Fatalf("expected 1 project, got %d: %+v", len(result.Projects), result.Projects)
		}
		if result.Projects[0].ProjectName != "project-a" {
			t.Fatalf("expected project-a, got %q", result.Projects[0].ProjectName)
		}
		if result.Projects[0].SpecCount != 1 {
			t.Fatalf("expected spec count 1 for project-a, got %d", result.Projects[0].SpecCount)
		}
		if len(result.RecentlyUpdated) != 1 || result.RecentlyUpdated[0].ProjectName != "project-a" {
			t.Fatalf("expected recentlyUpdated scoped to project-a, got %+v", result.RecentlyUpdated)
		}
		if result.RecentlyUpdated[0].ProjectID != 1 {
			t.Fatalf("expected recentlyUpdated entry to carry project id 1, got %+v", result.RecentlyUpdated[0])
		}
	})

	t.Run("filtering by an unregistered id yields an empty result", func(t *testing.T) {
		id := int64(999)
		result, err := overview.GetOverview(ctx, &id)
		if err != nil {
			t.Fatalf("GetOverview: %v", err)
		}
		if len(result.Projects) != 0 {
			t.Fatalf("expected 0 projects, got %+v", result.Projects)
		}
	})
}

func TestOverviewServiceBrokenSpecCount(t *testing.T) {
	ctx := context.Background()
	projA := newTestProject(t, 1, "project-a")
	projB := newTestProject(t, 2, "project-b")
	lister := &stubProjectLister{projects: []domain.Project{projA, projB}}

	specs := NewSpecService(lister, "author")
	if _, err := specs.CreateSpec(ctx, 1, "base", "Base capability.", []domain.SpecRequirement{{
		Name: "Base req", Body: "The system SHALL provide the base.\n\n#### Scenario: Base works\n- **WHEN** used\n- **THEN** it works",
	}}); err != nil {
		t.Fatalf("CreateSpec base: %v", err)
	}
	if _, err := specs.CreateSpec(ctx, 2, "consumer", "Consumer capability.", []domain.SpecRequirement{{
		Name: "Consumer req", Body: "The system SHALL consume the base.\n\n#### Scenario: Consumes\n- **WHEN** used\n- **THEN** it consumes",
	}}); err != nil {
		t.Fatalf("CreateSpec consumer: %v", err)
	}
	if err := specs.AddDependency(ctx, 2, "consumer", 1, "base"); err != nil {
		t.Fatalf("AddDependency: %v", err)
	}

	changes := NewChangeService(lister)
	overview := NewOverviewService(lister, specs, changes)

	t.Run("no broken dependencies yet", func(t *testing.T) {
		result, err := overview.GetOverview(ctx, nil)
		if err != nil {
			t.Fatalf("GetOverview: %v", err)
		}
		for _, p := range result.Projects {
			if p.BrokenSpecCount != 0 {
				t.Fatalf("expected no broken specs before deletion, got %+v", p)
			}
		}
	})

	// Force-delete "base" so consumer's dependency becomes unavailable.
	if _, err := specs.DeleteSpec(ctx, 1, "base", true); err != nil {
		t.Fatalf("DeleteSpec base: %v", err)
	}

	t.Run("broken dependency is counted on the dependent project only", func(t *testing.T) {
		result, err := overview.GetOverview(ctx, nil)
		if err != nil {
			t.Fatalf("GetOverview: %v", err)
		}
		byName := map[string]domain.ProjectOverview{}
		for _, p := range result.Projects {
			byName[p.ProjectName] = p
		}
		if byName["project-b"].BrokenSpecCount != 1 {
			t.Fatalf("expected project-b to have 1 broken spec, got %+v", byName["project-b"])
		}
		if byName["project-a"].BrokenSpecCount != 0 {
			t.Fatalf("expected project-a to have 0 broken specs, got %+v", byName["project-a"])
		}
	})
}

// TestOverviewServiceProgressRatioCountsArchived reproduces the reported bug:
// a project with Todo:1, In Progress:2, Done:0, Archived:6 previously showed
// a 0% progress ratio because Archived wasn't counted as completed work.
func TestOverviewServiceProgressRatioCountsArchived(t *testing.T) {
	ctx := context.Background()
	proj := newTestProject(t, 1, "judit-monitoramento")
	lister := &stubProjectLister{projects: []domain.Project{proj}}

	writeChange(t, proj.Path, "todo-change", false, 2, 0)
	writeChange(t, proj.Path, "in-progress-1", false, 2, 1)
	writeChange(t, proj.Path, "in-progress-2", false, 4, 1)
	for i := 1; i <= 6; i++ {
		writeChange(t, proj.Path, fmt.Sprintf("archived-%d", i), true, 1, 1)
	}

	specs := NewSpecService(lister, "author")
	changes := NewChangeService(lister)
	overview := NewOverviewService(lister, specs, changes)

	result, err := overview.GetOverview(ctx, nil)
	if err != nil {
		t.Fatalf("GetOverview: %v", err)
	}
	if len(result.Projects) != 1 {
		t.Fatalf("expected 1 project, got %+v", result.Projects)
	}
	p := result.Projects[0]
	if got := p.ChangeCounts[string(domain.ChangeStatusTodo)]; got != 1 {
		t.Fatalf("expected 1 todo change, got %d (%+v)", got, p.ChangeCounts)
	}
	if got := p.ChangeCounts[string(domain.ChangeStatusInProgress)]; got != 2 {
		t.Fatalf("expected 2 in-progress changes, got %d (%+v)", got, p.ChangeCounts)
	}
	if got := p.ChangeCounts[string(domain.ChangeStatusDone)]; got != 0 {
		t.Fatalf("expected 0 done changes, got %d (%+v)", got, p.ChangeCounts)
	}
	if got := p.ChangeCounts[string(domain.ChangeStatusArchived)]; got != 6 {
		t.Fatalf("expected 6 archived changes, got %d (%+v)", got, p.ChangeCounts)
	}
	wantRatio := 6.0 / 9.0
	if diff := p.ProgressRatio - wantRatio; diff > 1e-9 || diff < -1e-9 {
		t.Fatalf("expected progress ratio %.4f (archived counted as completed), got %.4f", wantRatio, p.ProgressRatio)
	}
}
