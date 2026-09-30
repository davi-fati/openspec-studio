package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/davidlima/openspec-studio/backend/internal/domain"
)

// TestSpecflowServiceSchedulerTriggersDueFlow exercises RunScheduler itself
// (the real ticker-driven poll loop), not execute() directly - task 2.1
// asks specifically to verify a flow scheduled a few seconds out actually
// starts through the scheduler.
func TestSpecflowServiceSchedulerTriggersDueFlow(t *testing.T) {
	ctx := context.Background()
	proj := newTestProject(t, 1, "demo")
	writeFullyPlannedChange(t, proj.Path, "add-a", "cap-a")

	marker := filepath.Join(t.TempDir(), "ran.marker")
	scriptPath := filepath.Join(t.TempDir(), "agent.sh")
	writeFakeAgentScript(t, scriptPath, "touch "+marker+"\necho ok\n")

	lister := &stubProjectLister{projects: []domain.Project{proj}}
	providerRepo := newStubProviderRepository()
	providerSvc := NewProviderService(providerRepo)
	p, err := providerSvc.CreateCLI(ctx, "fake-agent", scriptPath, nil, true)
	if err != nil {
		t.Fatalf("CreateCLI: %v", err)
	}

	repo := newStubSpecflowRepo()
	svc := NewSpecflowService(repo, lister, providerSvc, nil)

	// Scheduled a couple seconds out, inside the scheduler's 5s poll window.
	flow, err := svc.Create(ctx, 1, p.ID, time.Now().Add(2*time.Second), []string{"add-a"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	schedCtx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()
	go svc.RunScheduler(schedCtx)

	deadline := time.Now().Add(11 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(marker); err == nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatal("scheduler did not trigger the due flow within the expected window")
	}

	final, _, err := repo.Get(ctx, flow.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if final.Status != domain.FlowStatusSucceeded {
		t.Fatalf("expected scheduler-triggered flow to succeed, got %q", final.Status)
	}
}
