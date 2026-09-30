package service

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/davidlima/openspec-studio/backend/internal/domain"
)

func TestSpecflowServiceLogsRetrievableLiveAndAfterCompletion(t *testing.T) {
	ctx := context.Background()
	proj := newTestProject(t, 1, "demo")
	writeFullyPlannedChange(t, proj.Path, "add-a", "cap-a")

	scriptPath := filepath.Join(t.TempDir(), "agent.sh")
	writeFakeAgentScript(t, scriptPath, "echo line-one\nsleep 1\necho line-two\n")

	lister := &stubProjectLister{projects: []domain.Project{proj}}
	providerRepo := newStubProviderRepository()
	providerSvc := NewProviderService(providerRepo)
	p, _ := providerSvc.CreateCLI(ctx, "fake-agent", scriptPath, nil, true)

	repo := newStubSpecflowRepo()
	svc := NewSpecflowService(repo, lister, providerSvc, nil)

	flow, err := svc.Create(ctx, 1, p.ID, time.Now(), []string{"add-a"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	execDone := make(chan struct{})
	go func() {
		svc.execute(flow.ID)
		close(execDone)
	}()

	// While the item is still running (between "line-one" and the 1s sleep),
	// the first line's output must already be retrievable - this is the
	// "live" half of the requirement.
	var sawLiveLine1 bool
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		f, _, _ := repo.Get(ctx, flow.ID)
		if len(f.Items) == 1 && strings.Contains(f.Items[0].Log, "line-one") {
			sawLiveLine1 = true
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if !sawLiveLine1 {
		t.Fatal("expected 'line-one' to be visible in the log before the item finished running")
	}

	<-execDone

	final, _, _ := repo.Get(ctx, flow.ID)
	log := final.Items[0].Log
	if !strings.Contains(log, "line-one") || !strings.Contains(log, "line-two") {
		t.Fatalf("expected both lines in the final persisted log, got %q", log)
	}
	if final.Items[0].Status != domain.FlowItemSucceeded {
		t.Fatalf("expected item to succeed, got %q", final.Items[0].Status)
	}
}
