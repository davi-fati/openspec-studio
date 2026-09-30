package service

import (
	"context"
	"testing"
	"time"

	"github.com/davidlima/openspec-studio/backend/internal/domain"
)

func TestSpecflowServiceMarksMissedOverdueFlows(t *testing.T) {
	ctx := context.Background()
	proj := newTestProject(t, 1, "demo")
	writeFullyPlannedChange(t, proj.Path, "add-a", "cap-a")

	lister := &stubProjectLister{projects: []domain.Project{proj}}
	providerRepo := newStubProviderRepository()
	providerSvc := NewProviderService(providerRepo)
	p, _ := providerSvc.CreateCLI(ctx, "noop", "true", nil, false)

	repo := newStubSpecflowRepo()
	svc := NewSpecflowService(repo, lister, providerSvc, nil)

	// Simulates "the Studio was not running when this flow's scheduled time
	// arrived": the flow's start time is already in the past when the
	// backend (re)starts.
	overdue, err := svc.Create(ctx, 1, p.ID, time.Now().Add(-time.Hour), []string{"add-a"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := svc.MarkMissedOverdueFlows(ctx); err != nil {
		t.Fatalf("MarkMissedOverdueFlows: %v", err)
	}

	got, err := svc.Get(ctx, overdue.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !got.Missed {
		t.Fatal("expected the overdue flow to be flagged missed")
	}
	if got.Status != domain.FlowStatusPending {
		t.Fatalf("expected status to remain pending (not silently run), got %q", got.Status)
	}

	// A missed flow must not be picked up by the regular scheduler poll.
	due, err := repo.DueFlows(ctx, time.Now())
	if err != nil {
		t.Fatalf("DueFlows: %v", err)
	}
	for _, f := range due {
		if f.ID == overdue.ID {
			t.Fatal("a missed flow must not appear as due for the scheduler to auto-run")
		}
	}
}
