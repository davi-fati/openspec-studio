package service

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/davidlima/openspec-studio/backend/internal/domain"
)

// stubSpecflowRepo is an in-memory SpecflowRepo for tests.
type stubSpecflowRepo struct {
	mu     sync.Mutex
	flows  map[int64]domain.Flow
	nextID int64
}

func newStubSpecflowRepo() *stubSpecflowRepo {
	return &stubSpecflowRepo{flows: map[int64]domain.Flow{}}
}

func (r *stubSpecflowRepo) List(ctx context.Context) ([]domain.Flow, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []domain.Flow
	for _, f := range r.flows {
		out = append(out, f)
	}
	return out, nil
}

func (r *stubSpecflowRepo) Get(ctx context.Context, id int64) (domain.Flow, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	f, ok := r.flows[id]
	return f, ok, nil
}

func (r *stubSpecflowRepo) DueFlows(ctx context.Context, now time.Time) ([]domain.Flow, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []domain.Flow
	for _, f := range r.flows {
		if f.Status == domain.FlowStatusPending && !f.Missed && !f.ScheduledAt.After(now) {
			out = append(out, f)
		}
	}
	return out, nil
}

func (r *stubSpecflowRepo) Create(ctx context.Context, projectID, providerID int64, scheduledAt time.Time, changeNames []string) (domain.Flow, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nextID++
	items := make([]domain.FlowItem, len(changeNames))
	for i, name := range changeNames {
		items[i] = domain.FlowItem{ID: r.nextID*1000 + int64(i), Position: i, ChangeName: name, Status: domain.FlowItemPending}
	}
	f := domain.Flow{
		ID: r.nextID, ProjectID: projectID, ProviderID: providerID, ScheduledAt: scheduledAt,
		Status: domain.FlowStatusPending, CreatedAt: time.Now().UTC(), Items: items,
	}
	r.flows[f.ID] = f
	return f, nil
}

func (r *stubSpecflowRepo) Reorder(ctx context.Context, flowID int64, orderedItemIDs []int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	f := r.flows[flowID]
	byID := map[int64]domain.FlowItem{}
	for _, it := range f.Items {
		byID[it.ID] = it
	}
	newItems := make([]domain.FlowItem, 0, len(orderedItemIDs))
	for i, id := range orderedItemIDs {
		it := byID[id]
		it.Position = i
		newItems = append(newItems, it)
	}
	f.Items = newItems
	r.flows[flowID] = f
	return nil
}

func (r *stubSpecflowRepo) Delete(ctx context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.flows, id)
	return nil
}

func (r *stubSpecflowRepo) SetStatus(ctx context.Context, id int64, status domain.FlowStatus) error {
	if err := ctx.Err(); err != nil {
		// Real SQLite writes fail the same way: a cancelled ctx must never
		// silently prevent persisting a flow's final status.
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	f := r.flows[id]
	f.Status = status
	r.flows[id] = f
	return nil
}

func (r *stubSpecflowRepo) MarkMissed(ctx context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	f := r.flows[id]
	f.Missed = true
	r.flows[id] = f
	return nil
}

func (r *stubSpecflowRepo) SetItemStatus(ctx context.Context, itemID int64, status domain.FlowItemStatus, log string, startedAt, finishedAt *time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for fid, f := range r.flows {
		for i, it := range f.Items {
			if it.ID == itemID {
				it.Status = status
				it.Log = log
				it.StartedAt = startedAt
				it.FinishedAt = finishedAt
				f.Items[i] = it
				r.flows[fid] = f
				return nil
			}
		}
	}
	return nil
}

// writeFullyPlannedChange scaffolds a change dir with proposal.md, tasks.md,
// and one delta spec under specs/<capability>/spec.md.
func writeFullyPlannedChange(t *testing.T, projectPath, changeName, capability string) {
	t.Helper()
	dir := filepath.Join(projectPath, "openspec", "changes", changeName)
	if err := os.MkdirAll(filepath.Join(dir, "specs", capability), 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(dir, "proposal.md"), []byte("## Why\ntest\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "tasks.md"), []byte("## 1\n- [ ] 1.1 x\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "specs", capability, "spec.md"), []byte("## ADDED Requirements\n\n### Requirement: X\nThe system SHALL x.\n\n#### Scenario: S\n- **WHEN** a\n- **THEN** b\n"), 0o644)
}

func newTestSpecflowService(t *testing.T, lister *stubProjectLister) (*SpecflowService, *stubSpecflowRepo) {
	t.Helper()
	repo := newStubSpecflowRepo()
	providerRepo := newStubProviderRepository()
	providerSvc := NewProviderService(providerRepo)
	svc := NewSpecflowService(repo, lister, providerSvc, nil)
	return svc, repo
}

func TestSpecflowServiceCreateAndReorder(t *testing.T) {
	ctx := context.Background()
	proj := newTestProject(t, 1, "demo")
	writeFullyPlannedChange(t, proj.Path, "add-a", "cap-a")
	writeFullyPlannedChange(t, proj.Path, "add-b", "cap-b")
	lister := &stubProjectLister{projects: []domain.Project{proj}}
	svc, _ := newTestSpecflowService(t, lister)

	flow, err := svc.Create(ctx, 1, 1, time.Now().Add(time.Hour), []string{"add-a", "add-b"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if len(flow.Items) != 2 || flow.Items[0].ChangeName != "add-a" {
		t.Fatalf("unexpected items: %+v", flow.Items)
	}

	reversed := []int64{flow.Items[1].ID, flow.Items[0].ID}
	if err := svc.Reorder(ctx, flow.ID, reversed); err != nil {
		t.Fatalf("Reorder: %v", err)
	}
	reordered, err := svc.Get(ctx, flow.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if reordered.Items[0].ChangeName != "add-b" {
		t.Fatalf("expected add-b first after reorder, got %+v", reordered.Items)
	}
}

func TestSpecflowServiceRejectsIncompleteChange(t *testing.T) {
	ctx := context.Background()
	proj := newTestProject(t, 1, "demo")
	// add-incomplete has a proposal but no tasks.md / specs.
	dir := filepath.Join(proj.Path, "openspec", "changes", "add-incomplete")
	os.MkdirAll(dir, 0o755)
	os.WriteFile(filepath.Join(dir, "proposal.md"), []byte("## Why\nx\n"), 0o644)

	lister := &stubProjectLister{projects: []domain.Project{proj}}
	svc, _ := newTestSpecflowService(t, lister)

	if _, err := svc.Create(ctx, 1, 1, time.Now().Add(time.Hour), []string{"add-incomplete"}); err == nil {
		t.Fatal("expected an error for an incompletely-planned change")
	} else if !strings.Contains(err.Error(), "not fully planned") {
		t.Fatalf("expected 'not fully planned' error, got %v", err)
	}
}

func TestSpecflowServiceRejectsUnmetDependency(t *testing.T) {
	ctx := context.Background()
	proj := newTestProject(t, 1, "demo")
	writeFullyPlannedChange(t, proj.Path, "add-consumer", "consumer-cap")

	// Declare (via the same sidecar format SpecService uses) that
	// consumer-cap depends on base-cap, which is neither applied nor
	// earlier in the flow.
	consumerSpecDir := filepath.Join(proj.Path, "openspec", "changes", "add-consumer", "specs", "consumer-cap")
	meta := domain.SpecMeta{DependsOn: []domain.SpecDependencyRef{{ProjectPath: proj.Path, Capability: "base-cap"}}}
	if err := writeSpecMeta(consumerSpecDir, meta); err != nil {
		t.Fatal(err)
	}
	// capabilityDependencies reads from the PROJECT's main specs tree by
	// capability name, so mirror the same dependency declaration there too
	// (this is where a capability's dependsOn actually lives once authored).
	mainCapDir := filepath.Join(proj.Path, "openspec", "specs", "consumer-cap")
	if err := os.MkdirAll(mainCapDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := writeSpecMeta(mainCapDir, meta); err != nil {
		t.Fatal(err)
	}

	lister := &stubProjectLister{projects: []domain.Project{proj}}
	svc, _ := newTestSpecflowService(t, lister)

	_, err := svc.Create(ctx, 1, 1, time.Now().Add(time.Hour), []string{"add-consumer"})
	if err == nil {
		t.Fatal("expected an unmet-dependency error")
	}
	if !strings.Contains(err.Error(), "base-cap") {
		t.Fatalf("expected error naming the unmet capability, got %v", err)
	}
}

func TestSpecflowServiceDependencySatisfiedEarlierInFlow(t *testing.T) {
	ctx := context.Background()
	proj := newTestProject(t, 1, "demo")
	writeFullyPlannedChange(t, proj.Path, "add-base", "base-cap")
	writeFullyPlannedChange(t, proj.Path, "add-consumer", "consumer-cap")

	consumerSpecDir := filepath.Join(proj.Path, "openspec", "changes", "add-consumer", "specs", "consumer-cap")
	meta := domain.SpecMeta{DependsOn: []domain.SpecDependencyRef{{ProjectPath: proj.Path, Capability: "base-cap"}}}
	writeSpecMeta(consumerSpecDir, meta)
	mainCapDir := filepath.Join(proj.Path, "openspec", "specs", "consumer-cap")
	os.MkdirAll(mainCapDir, 0o755)
	writeSpecMeta(mainCapDir, meta)

	lister := &stubProjectLister{projects: []domain.Project{proj}}
	svc, _ := newTestSpecflowService(t, lister)

	// base-cap's implementing change (add-base) comes first in the flow, so scheduling should succeed.
	if _, err := svc.Create(ctx, 1, 1, time.Now().Add(time.Hour), []string{"add-base", "add-consumer"}); err != nil {
		t.Fatalf("expected scheduling to succeed with dependency earlier in flow, got %v", err)
	}
}
