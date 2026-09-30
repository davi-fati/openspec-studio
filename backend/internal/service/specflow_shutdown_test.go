package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/davidlima/openspec-studio/backend/internal/domain"
)

// TestSpecflowServiceShutdownCancelsRunningFlow drives a flow through the
// real scheduler, then shuts the service down mid-item: the flow must end
// cancelled (as with a user cancel) and the provider's process group must
// be gone by the time Shutdown returns.
func TestSpecflowServiceShutdownCancelsRunningFlow(t *testing.T) {
	ctx := context.Background()
	proj := newTestProject(t, 1, "demo")
	writeFullyPlannedChange(t, proj.Path, "add-slow", "cap-slow")

	pidFile := filepath.Join(t.TempDir(), "sleep.pid")
	slowScript := filepath.Join(t.TempDir(), "slow.sh")
	writeFakeAgentScript(t, slowScript, fmt.Sprintf("sleep 30 & echo $! > %q\nwait\n", pidFile))

	lister := &stubProjectLister{projects: []domain.Project{proj}}
	providerSvc := NewProviderService(newStubProviderRepository())
	p, err := providerSvc.CreateCLI(ctx, "slow-agent", slowScript, nil, true)
	if err != nil {
		t.Fatalf("CreateCLI: %v", err)
	}

	repo := newStubSpecflowRepo()
	svc := NewSpecflowService(repo, lister, providerSvc, nil)
	flow, err := svc.Create(ctx, 1, p.ID, time.Now(), []string{"add-slow"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	schedCtx, stopScheduler := context.WithCancel(ctx)
	defer stopScheduler()
	go svc.RunScheduler(schedCtx)

	var sleepPID int
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if b, err := os.ReadFile(pidFile); err == nil && len(strings.TrimSpace(string(b))) > 0 {
			sleepPID, _ = strconv.Atoi(strings.TrimSpace(string(b)))
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if sleepPID == 0 {
		t.Fatal("flow never started its provider process")
	}

	stopScheduler()
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := svc.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}

	final, _, _ := repo.Get(ctx, flow.ID)
	if final.Status != domain.FlowStatusCancelled {
		t.Fatalf("flow status after shutdown = %q, want cancelled", final.Status)
	}
	if err := syscall.Kill(sleepPID, 0); err == nil {
		t.Fatalf("provider grandchild %d still alive after shutdown", sleepPID)
	}
}

// TestSpecflowServiceShutdownBlocksNewExecutions checks that a flow the
// scheduler hands off after Shutdown is left untouched rather than run.
func TestSpecflowServiceShutdownBlocksNewExecutions(t *testing.T) {
	ctx := context.Background()
	proj := newTestProject(t, 1, "demo")
	writeFullyPlannedChange(t, proj.Path, "add-a", "cap-a")

	marker := filepath.Join(t.TempDir(), "ran.marker")
	script := filepath.Join(t.TempDir(), "agent.sh")
	writeFakeAgentScript(t, script, "touch "+marker+"\n")

	lister := &stubProjectLister{projects: []domain.Project{proj}}
	providerSvc := NewProviderService(newStubProviderRepository())
	p, _ := providerSvc.CreateCLI(ctx, "agent", script, nil, true)

	repo := newStubSpecflowRepo()
	svc := NewSpecflowService(repo, lister, providerSvc, nil)
	flow, err := svc.Create(ctx, 1, p.ID, time.Now(), []string{"add-a"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := svc.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
	svc.execute(flow.ID)

	if _, err := os.Stat(marker); err == nil {
		t.Fatal("flow executed after Shutdown")
	}
	final, _, _ := repo.Get(ctx, flow.ID)
	if final.Status != domain.FlowStatusPending {
		t.Fatalf("flow status = %q, want pending", final.Status)
	}
}
