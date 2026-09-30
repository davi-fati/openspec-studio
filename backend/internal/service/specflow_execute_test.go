package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/davidlima/openspec-studio/backend/internal/domain"
)

func writeFakeAgentScript(t *testing.T, path, body string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fake shell script CLI provider not supported on windows")
	}
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body), 0o755); err != nil {
		t.Fatalf("write fake agent: %v", err)
	}
}

func TestSpecflowServiceExecutesItemsInOrder(t *testing.T) {
	ctx := context.Background()
	proj := newTestProject(t, 1, "demo")
	writeFullyPlannedChange(t, proj.Path, "add-a", "cap-a")
	writeFullyPlannedChange(t, proj.Path, "add-b", "cap-b")

	orderLog := filepath.Join(t.TempDir(), "order.log")
	scriptPath := filepath.Join(t.TempDir(), "agent.sh")
	writeFakeAgentScript(t, scriptPath, fmt.Sprintf("echo \"$*\" >> %q\necho done\n", orderLog))

	lister := &stubProjectLister{projects: []domain.Project{proj}}
	providerRepo := newStubProviderRepository()
	providerSvc := NewProviderService(providerRepo)
	p, err := providerSvc.CreateCLI(ctx, "fake-agent", scriptPath, nil, true)
	if err != nil {
		t.Fatalf("CreateCLI: %v", err)
	}

	repo := newStubSpecflowRepo()
	svc := NewSpecflowService(repo, lister, providerSvc, nil)

	flow, err := svc.Create(ctx, 1, p.ID, time.Now(), []string{"add-a", "add-b"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	svc.execute(flow.ID)

	final, _, err := repo.Get(ctx, flow.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if final.Status != domain.FlowStatusSucceeded {
		t.Fatalf("expected flow to succeed, got %q (items: %+v)", final.Status, final.Items)
	}
	for _, item := range final.Items {
		if item.Status != domain.FlowItemSucceeded {
			t.Fatalf("expected item %q to succeed, got %q: %s", item.ChangeName, item.Status, item.Log)
		}
	}

	order, err := os.ReadFile(orderLog)
	if err != nil {
		t.Fatalf("read order log: %v", err)
	}
	got := string(order)
	idxA := indexOf(got, "add-a")
	idxB := indexOf(got, "add-b")
	if idxA == -1 || idxB == -1 || idxA > idxB {
		t.Fatalf("expected add-a to run strictly before add-b, got order log: %q", got)
	}
}

func TestSpecflowServiceFailureIsolatedPerFlow(t *testing.T) {
	ctx := context.Background()
	projA := newTestProject(t, 1, "project-a")
	projB := newTestProject(t, 2, "project-b")
	writeFullyPlannedChange(t, projA.Path, "add-fails", "cap-fail")
	writeFullyPlannedChange(t, projB.Path, "add-succeeds", "cap-ok")

	lister := &stubProjectLister{projects: []domain.Project{projA, projB}}
	providerRepo := newStubProviderRepository()
	providerSvc := NewProviderService(providerRepo)

	failScript := filepath.Join(t.TempDir(), "fail.sh")
	writeFakeAgentScript(t, failScript, "echo boom 1>&2\nexit 1\n")
	okScript := filepath.Join(t.TempDir(), "ok.sh")
	writeFakeAgentScript(t, okScript, "echo ok\n")

	failProvider, _ := providerSvc.CreateCLI(ctx, "fail-agent", failScript, nil, false)
	okProvider, _ := providerSvc.CreateCLI(ctx, "ok-agent", okScript, nil, false)

	repo := newStubSpecflowRepo()
	svc := NewSpecflowService(repo, lister, providerSvc, nil)

	flowFail, err := svc.Create(ctx, 1, failProvider.ID, time.Now(), []string{"add-fails"})
	if err != nil {
		t.Fatalf("Create (fail flow): %v", err)
	}
	flowOK, err := svc.Create(ctx, 2, okProvider.ID, time.Now(), []string{"add-succeeds"})
	if err != nil {
		t.Fatalf("Create (ok flow): %v", err)
	}

	// Run both "concurrently" as the real scheduler would.
	done := make(chan struct{}, 2)
	go func() { svc.execute(flowFail.ID); done <- struct{}{} }()
	go func() { svc.execute(flowOK.ID); done <- struct{}{} }()
	<-done
	<-done

	got1, _, _ := repo.Get(ctx, flowFail.ID)
	got2, _, _ := repo.Get(ctx, flowOK.ID)
	if got1.Status != domain.FlowStatusFailed {
		t.Fatalf("expected project A's flow to fail, got %q", got1.Status)
	}
	if got2.Status != domain.FlowStatusSucceeded {
		t.Fatalf("expected project B's flow to succeed despite A's failure, got %q", got2.Status)
	}
}

func TestSpecflowServiceCancelStopsRunningProcess(t *testing.T) {
	ctx := context.Background()
	proj := newTestProject(t, 1, "demo")
	writeFullyPlannedChange(t, proj.Path, "add-slow", "cap-slow")

	// The grandchild's own PID is recorded so the test can verify, after
	// cancel, that the orphaned "sleep" was actually killed - not just that
	// our Go code stopped waiting on it.
	pidFile := filepath.Join(t.TempDir(), "sleep.pid")
	slowScript := filepath.Join(t.TempDir(), "slow.sh")
	writeFakeAgentScript(t, slowScript, fmt.Sprintf("sleep 30 & echo $! > %q\nwait\n", pidFile))

	lister := &stubProjectLister{projects: []domain.Project{proj}}
	providerRepo := newStubProviderRepository()
	providerSvc := NewProviderService(providerRepo)
	p, _ := providerSvc.CreateCLI(ctx, "slow-agent", slowScript, nil, false)

	repo := newStubSpecflowRepo()
	svc := NewSpecflowService(repo, lister, providerSvc, nil)

	flow, err := svc.Create(ctx, 1, p.ID, time.Now(), []string{"add-slow"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	execDone := make(chan struct{})
	go func() {
		svc.execute(flow.ID)
		close(execDone)
	}()

	// Wait for the "sleep 30" grandchild to actually fork (its pid file to
	// appear) before cancelling. Cancelling too early - before the shell
	// forks sleep - would kill the shell before it has any orphan-able
	// children, masking the process-group-kill behavior this test exists
	// to exercise.
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(pidFile); err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if _, err := os.Stat(pidFile); err != nil {
		t.Fatal("sleep subprocess never started (pid file never appeared)")
	}

	start := time.Now()
	if err := svc.Cancel(ctx, flow.ID); err != nil {
		t.Fatalf("Cancel: %v", err)
	}

	select {
	case <-execDone:
	case <-time.After(5 * time.Second):
		t.Fatal("execute did not return within 5s of Cancel - subprocess likely not terminated")
	}
	elapsed := time.Since(start)
	if elapsed > 5*time.Second {
		t.Fatalf("cancel took too long (%s) - the sleep 30 subprocess was likely not killed", elapsed)
	}

	final, _, _ := repo.Get(ctx, flow.ID)
	if final.Status != domain.FlowStatusCancelled {
		t.Fatalf("expected flow status cancelled, got %q", final.Status)
	}

	// The real assertion: the orphaned grandchild ("sleep 30", backgrounded
	// by the script) must actually be dead, not just detached from our
	// pipe - a plain kill of the direct child would leave it running.
	pidBytes, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatalf("read pid file: %v", err)
	}
	pid := 0
	if _, err := fmt.Sscanf(string(pidBytes), "%d", &pid); err != nil || pid == 0 {
		t.Fatalf("parse pid file %q: %v", string(pidBytes), err)
	}
	if processAlive(pid) {
		t.Fatalf("expected sleep process (pid %d) to be dead after cancel, but it is still alive", pid)
	}
}

func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
