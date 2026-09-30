package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/davidlima/openspec-studio/backend/internal/domain"
)

var (
	ErrFlowNotFound        = errors.New("flow not found")
	ErrChangeNotFullyPlanned = errors.New("change is not fully planned (needs proposal.md, tasks.md, and at least one spec)")
	ErrUnmetDependency     = errors.New("unmet spec dependency")
	ErrFlowNotRunning      = errors.New("flow is not running")
)

// SpecflowRepo is the persistence contract SpecflowService depends on.
type SpecflowRepo interface {
	List(ctx context.Context) ([]domain.Flow, error)
	Get(ctx context.Context, id int64) (domain.Flow, bool, error)
	DueFlows(ctx context.Context, now time.Time) ([]domain.Flow, error)
	Create(ctx context.Context, projectID, providerID int64, scheduledAt time.Time, changeNames []string) (domain.Flow, error)
	Reorder(ctx context.Context, flowID int64, orderedItemIDs []int64) error
	Delete(ctx context.Context, id int64) error
	SetStatus(ctx context.Context, id int64, status domain.FlowStatus) error
	MarkMissed(ctx context.Context, id int64) error
	SetItemStatus(ctx context.Context, itemID int64, status domain.FlowItemStatus, log string, startedAt, finishedAt *time.Time) error
}

type SpecflowService struct {
	repo      SpecflowRepo
	projects  SpecProjectLister
	providers *ProviderService
	events    *EventBroadcaster

	mu      sync.Mutex
	cancels map[int64]context.CancelFunc
	// running tracks scheduler-started executions so Shutdown can wait for
	// them to record their final status; closing stops new ones starting.
	running sync.WaitGroup
	closing bool
}

func NewSpecflowService(repo SpecflowRepo, projects SpecProjectLister, providers *ProviderService, events *EventBroadcaster) *SpecflowService {
	return &SpecflowService{repo: repo, projects: projects, providers: providers, events: events, cancels: make(map[int64]context.CancelFunc)}
}

func (s *SpecflowService) List(ctx context.Context, projectID int64) ([]domain.Flow, error) {
	flows, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	if projectID == 0 {
		s.attachProjectNames(ctx, flows)
		return flows, nil
	}
	var filtered []domain.Flow
	for _, f := range flows {
		if f.ProjectID == projectID {
			filtered = append(filtered, f)
		}
	}
	s.attachProjectNames(ctx, filtered)
	return filtered, nil
}

func (s *SpecflowService) Get(ctx context.Context, id int64) (domain.Flow, error) {
	f, ok, err := s.repo.Get(ctx, id)
	if err != nil {
		return domain.Flow{}, err
	}
	if !ok {
		return domain.Flow{}, ErrFlowNotFound
	}
	s.attachProjectNames(ctx, []domain.Flow{f})
	return f, nil
}

func (s *SpecflowService) attachProjectNames(ctx context.Context, flows []domain.Flow) {
	cache := map[int64]string{}
	for i := range flows {
		id := flows[i].ProjectID
		if name, ok := cache[id]; ok {
			flows[i].ProjectName = name
			continue
		}
		if p, err := s.projects.GetProject(ctx, id); err == nil {
			cache[id] = p.Name
			flows[i].ProjectName = p.Name
		}
	}
}

// Create validates every change is fully planned and that dependency order
// is satisfiable (earlier in the flow, or already applied) before
// persisting anything - both checks give early feedback at schedule time;
// they're re-checked at execution time too, since project state can change
// between scheduling and the run actually starting.
func (s *SpecflowService) Create(ctx context.Context, projectID, providerID int64, scheduledAt time.Time, changeNames []string) (domain.Flow, error) {
	p, err := s.projects.GetProject(ctx, projectID)
	if err != nil {
		return domain.Flow{}, err
	}

	if err := s.validateChangesSchedulable(p, changeNames); err != nil {
		return domain.Flow{}, err
	}

	return s.repo.Create(ctx, projectID, providerID, scheduledAt, changeNames)
}

func (s *SpecflowService) validateChangesSchedulable(p domain.Project, changeNames []string) error {
	seenCapabilities := map[string]bool{} // capabilities implemented earlier in this same flow

	for _, name := range changeNames {
		dir := filepath.Join(p.Path, "openspec", "changes", name)
		if !changeIsFullyPlanned(dir) {
			return fmt.Errorf("%w: %q", ErrChangeNotFullyPlanned, name)
		}

		capabilities := changeCapabilities(dir)
		for _, dep := range capabilityDependencies(p.Path, capabilities) {
			if seenCapabilities[dep.Capability] {
				continue
			}
			if capabilityAlreadyApplied(p.Path, dep.Capability) {
				continue
			}
			return fmt.Errorf("%w: change %q depends on capability %q, which is neither earlier in this flow nor already applied", ErrUnmetDependency, name, dep.Capability)
		}
		for _, cap := range capabilities {
			seenCapabilities[cap] = true
		}
	}
	return nil
}

func changeIsFullyPlanned(dir string) bool {
	if !fileExists(filepath.Join(dir, "proposal.md")) || !fileExists(filepath.Join(dir, "tasks.md")) {
		return false
	}
	entries, err := os.ReadDir(filepath.Join(dir, "specs"))
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() && fileExists(filepath.Join(dir, "specs", e.Name(), "spec.md")) {
			return true
		}
	}
	return false
}

func changeCapabilities(dir string) []string {
	entries, err := os.ReadDir(filepath.Join(dir, "specs"))
	if err != nil {
		return nil
	}
	var caps []string
	for _, e := range entries {
		if e.IsDir() {
			caps = append(caps, e.Name())
		}
	}
	return caps
}

// capabilityDependencies reads each capability's dependsOn from its
// .meta.json in the project's own (already-applied) spec tree - the only
// place dependency metadata currently lives (see spec-dependencies).
func capabilityDependencies(projectPath string, capabilities []string) []domain.SpecDependencyRef {
	var deps []domain.SpecDependencyRef
	for _, cap := range capabilities {
		meta, err := readSpecMeta(filepath.Join(projectPath, "openspec", "specs", cap))
		if err != nil {
			continue
		}
		deps = append(deps, meta.DependsOn...)
	}
	return deps
}

func capabilityAlreadyApplied(projectPath, capability string) bool {
	return fileExists(filepath.Join(projectPath, "openspec", "specs", capability, "spec.md"))
}

func (s *SpecflowService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

// ErrFlowAlreadyStarted is returned when reordering is attempted on a flow
// that is no longer pending - the spec only allows reordering before execution.
var ErrFlowAlreadyStarted = errors.New("flow has already started")

func (s *SpecflowService) Reorder(ctx context.Context, flowID int64, orderedItemIDs []int64) error {
	flow, err := s.Get(ctx, flowID)
	if err != nil {
		return err
	}
	if flow.Status != domain.FlowStatusPending {
		return ErrFlowAlreadyStarted
	}
	return s.repo.Reorder(ctx, flowID, orderedItemIDs)
}

// Cancel stops a running flow: its context is cancelled, which terminates
// any in-progress provider subprocess, and the flow is marked cancelled -
// already-succeeded items are left as they are.
func (s *SpecflowService) Cancel(ctx context.Context, id int64) error {
	s.mu.Lock()
	cancel, ok := s.cancels[id]
	s.mu.Unlock()
	if !ok {
		return ErrFlowNotRunning
	}
	cancel()
	return nil
}

// Shutdown stops the scheduler from starting new flows, cancels every
// running flow exactly as a user cancel would (killing provider process
// groups and recording "cancelled"), and waits for those executions to
// finish or ctx to expire.
func (s *SpecflowService) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	s.closing = true
	for _, cancel := range s.cancels {
		cancel()
	}
	s.mu.Unlock()

	done := make(chan struct{})
	go func() {
		s.running.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// MarkMissedOverdueFlows runs once at startup: any pending flow whose
// scheduled time already passed is flagged missed (never silently run or
// dropped) rather than auto-executed on launch.
func (s *SpecflowService) MarkMissedOverdueFlows(ctx context.Context) error {
	due, err := s.repo.DueFlows(ctx, time.Now().UTC())
	if err != nil {
		return err
	}
	for _, f := range due {
		if err := s.repo.MarkMissed(ctx, f.ID); err != nil {
			return err
		}
		log.Printf("specflow: flow %d missed its scheduled window (%s) and was not auto-run", f.ID, f.ScheduledAt)
	}
	return nil
}

// RunScheduler polls for due flows and executes each in its own goroutine,
// so a failure or long run in one project's flow never blocks another's.
// Call once; it blocks until ctx is cancelled.
func (s *SpecflowService) RunScheduler(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			due, err := s.repo.DueFlows(ctx, time.Now().UTC())
			if err != nil {
				log.Printf("specflow: poll for due flows: %v", err)
				continue
			}
			for _, f := range due {
				s.mu.Lock()
				if s.closing {
					s.mu.Unlock()
					return
				}
				s.running.Add(1)
				s.mu.Unlock()
				go func(id int64) {
					defer s.running.Done()
					s.execute(id)
				}(f.ID)
			}
		}
	}
}

func (s *SpecflowService) execute(flowID int64) {
	ctx, cancel := context.WithCancel(context.Background())
	s.mu.Lock()
	if s.closing {
		// Shutdown already ran its cancel sweep; leave the flow pending so
		// the next startup flags it missed.
		s.mu.Unlock()
		cancel()
		return
	}
	s.cancels[flowID] = cancel
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.cancels, flowID)
		s.mu.Unlock()
		cancel()
	}()

	flow, ok, err := s.repo.Get(ctx, flowID)
	if err != nil || !ok {
		log.Printf("specflow: flow %d disappeared before execution", flowID)
		return
	}

	p, err := s.projects.GetProject(ctx, flow.ProjectID)
	if err != nil {
		log.Printf("specflow: flow %d project unavailable: %v", flowID, err)
		s.finishFlow(ctx, flowID, domain.FlowStatusFailed)
		return
	}

	s.setFlowStatus(ctx, flowID, domain.FlowStatusRunning)

	seenCapabilities := map[string]bool{}
	for _, item := range flow.Items {
		dir := filepath.Join(p.Path, "openspec", "changes", item.ChangeName)

		// Re-validate at execution time: project state may have changed
		// since scheduling.
		if !changeIsFullyPlanned(dir) {
			s.failItem(ctx, item.ID, "change is no longer fully planned")
			s.finishFlow(ctx, flowID, domain.FlowStatusFailed)
			return
		}
		caps := changeCapabilities(dir)
		blocked := false
		for _, dep := range capabilityDependencies(p.Path, caps) {
			if seenCapabilities[dep.Capability] || capabilityAlreadyApplied(p.Path, dep.Capability) {
				continue
			}
			s.failItem(ctx, item.ID, fmt.Sprintf("unmet dependency: %s", dep.Capability))
			s.finishFlow(ctx, flowID, domain.FlowStatusFailed)
			blocked = true
			break
		}
		if blocked {
			return
		}

		if err := s.runItem(ctx, flow, item); err != nil {
			if ctx.Err() != nil {
				s.finishFlow(ctx, flowID, domain.FlowStatusCancelled)
				return
			}
			s.finishFlow(ctx, flowID, domain.FlowStatusFailed)
			return
		}
		for _, cap := range caps {
			seenCapabilities[cap] = true
		}
	}

	s.finishFlow(ctx, flowID, domain.FlowStatusSucceeded)
}

func (s *SpecflowService) runItem(ctx context.Context, flow domain.Flow, item domain.FlowItem) error {
	started := time.Now().UTC()
	s.setItemStatus(ctx, item.ID, domain.FlowItemRunning, "", &started, nil)
	s.publish(flow.ID, item.ID, "item_status", string(domain.FlowItemRunning))

	var logBuf sliceBuffer
	var mu sync.Mutex
	prompt := fmt.Sprintf("Implement OpenSpec change %q per its tasks.md.", item.ChangeName)

	err := s.providers.InvokeStreaming(ctx, flow.ProviderID, prompt, func(chunk string) {
		mu.Lock()
		logBuf.Append(chunk)
		snapshot := logBuf.String()
		mu.Unlock()
		s.setItemStatus(ctx, item.ID, domain.FlowItemRunning, snapshot, &started, nil)
		s.publish(flow.ID, item.ID, "item_log", chunk)
	})

	finished := time.Now().UTC()
	status := domain.FlowItemSucceeded
	if err != nil {
		status = domain.FlowItemFailed
		mu.Lock()
		logBuf.Append("\n[error] " + err.Error())
		mu.Unlock()
	}
	mu.Lock()
	finalLog := logBuf.String()
	mu.Unlock()

	s.setItemStatus(ctx, item.ID, status, finalLog, &started, &finished)
	s.publish(flow.ID, item.ID, "item_status", string(status))
	return err
}

func (s *SpecflowService) failItem(ctx context.Context, itemID int64, message string) {
	now := time.Now().UTC()
	s.setItemStatus(ctx, itemID, domain.FlowItemFailed, message, &now, &now)
}

// setItemStatus and setFlowStatus persist status changes on a fresh,
// independent context rather than the flow's own execution ctx: on
// cancel, that ctx is the one we just cancelled to kill the subprocess, so
// reusing it here would cancel the very DB write that's supposed to record
// "cancelled" - the persistence must outlive the ctx that triggered it.
func (s *SpecflowService) setItemStatus(ctx context.Context, itemID int64, status domain.FlowItemStatus, log string, started, finished *time.Time) {
	writeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.repo.SetItemStatus(writeCtx, itemID, status, log, started, finished); err != nil {
		fmt.Printf("specflow: persist item %d status: %v\n", itemID, err)
	}
}

func (s *SpecflowService) setFlowStatus(ctx context.Context, flowID int64, status domain.FlowStatus) {
	writeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.repo.SetStatus(writeCtx, flowID, status); err != nil {
		fmt.Printf("specflow: persist flow %d status: %v\n", flowID, err)
	}
	s.publish(flowID, 0, "flow_status", string(status))
}

func (s *SpecflowService) finishFlow(ctx context.Context, flowID int64, status domain.FlowStatus) {
	s.setFlowStatus(ctx, flowID, status)
}

func (s *SpecflowService) publish(flowID, itemID int64, name, data string) {
	if s.events == nil {
		return
	}
	s.events.Publish(Event{Name: "specflow", Data: fmt.Sprintf(`{"flowId":%d,"itemId":%d,"type":%q,"data":%q}`, flowID, itemID, name, data)})
}

// sliceBuffer is a tiny append-only string buffer (avoids pulling in
// strings.Builder's write-once semantics when we need the running snapshot
// repeatedly).
type sliceBuffer struct {
	parts []string
}

func (b *sliceBuffer) Append(s string) {
	b.parts = append(b.parts, s)
}

func (b *sliceBuffer) String() string {
	total := 0
	for _, p := range b.parts {
		total += len(p)
	}
	out := make([]byte, 0, total)
	for _, p := range b.parts {
		out = append(out, p...)
	}
	return string(out)
}
