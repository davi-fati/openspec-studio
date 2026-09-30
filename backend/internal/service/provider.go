package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"sync"
	"time"

	"github.com/davidlima/openspec-studio/backend/internal/domain"
)

var (
	ErrProviderNotFound  = errors.New("provider not found")
	ErrProviderExecutableNotFound = errors.New("executable not found")
)

// ProviderRepository is the persistence contract this service depends on.
type ProviderRepository interface {
	List(ctx context.Context) ([]domain.Provider, error)
	Get(ctx context.Context, id int64) (domain.Provider, bool, error)
	Create(ctx context.Context, p domain.Provider, apiKey string) (domain.Provider, error)
	Update(ctx context.Context, id int64, p domain.Provider, apiKey *string) (domain.Provider, error)
	Delete(ctx context.Context, id int64) error
	SetDefault(ctx context.Context, id int64) error
	GetAPIKey(id int64) (string, error)
}

type ProviderService struct {
	repo ProviderRepository
}

func NewProviderService(repo ProviderRepository) *ProviderService {
	return &ProviderService{repo: repo}
}

func (s *ProviderService) List(ctx context.Context) ([]domain.Provider, error) {
	return s.repo.List(ctx)
}

func (s *ProviderService) Get(ctx context.Context, id int64) (domain.Provider, error) {
	p, ok, err := s.repo.Get(ctx, id)
	if err != nil {
		return domain.Provider{}, err
	}
	if !ok {
		return domain.Provider{}, ErrProviderNotFound
	}
	return p, nil
}

func (s *ProviderService) CreateHosted(ctx context.Context, name, model, endpoint, apiKey string, isDefault bool) (domain.Provider, error) {
	p := domain.Provider{Name: name, Kind: domain.ProviderKindHosted, Model: model, Endpoint: endpoint, IsDefault: isDefault}
	return s.repo.Create(ctx, p, apiKey)
}

// CreateCLI validates the executable resolves before persisting - a
// non-resolvable path must never be saved silently.
func (s *ProviderService) CreateCLI(ctx context.Context, name, path string, args []string, isDefault bool) (domain.Provider, error) {
	if _, err := exec.LookPath(path); err != nil {
		return domain.Provider{}, ErrProviderExecutableNotFound
	}
	p := domain.Provider{Name: name, Kind: domain.ProviderKindCLI, CLIPath: path, CLIArgs: args, IsDefault: isDefault}
	return s.repo.Create(ctx, p, "")
}

func (s *ProviderService) Update(ctx context.Context, id int64, name, model, endpoint, cliPath string, cliArgs []string, apiKey *string) (domain.Provider, error) {
	existing, err := s.Get(ctx, id)
	if err != nil {
		return domain.Provider{}, err
	}
	if existing.Kind == domain.ProviderKindCLI && cliPath != "" {
		if _, err := exec.LookPath(cliPath); err != nil {
			return domain.Provider{}, ErrProviderExecutableNotFound
		}
	}
	p := domain.Provider{Name: name, Model: model, Endpoint: endpoint, CLIPath: cliPath, CLIArgs: cliArgs}
	return s.repo.Update(ctx, id, p, apiKey)
}

func (s *ProviderService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *ProviderService) SetDefault(ctx context.Context, id int64) error {
	return s.repo.SetDefault(ctx, id)
}

// HealthCheck runs a lightweight, explicit connectivity check - never
// triggered automatically, so it never causes surprise API usage or spawns
// a CLI process without the user asking for it right now.
func (s *ProviderService) HealthCheck(ctx context.Context, id int64) (domain.HealthResult, error) {
	p, err := s.Get(ctx, id)
	if err != nil {
		return domain.HealthResult{}, err
	}

	switch p.Kind {
	case domain.ProviderKindHosted:
		return s.healthCheckHosted(ctx, p), nil
	case domain.ProviderKindCLI:
		return s.healthCheckCLI(ctx, p), nil
	default:
		return domain.HealthResult{OK: false, Message: "unknown provider kind"}, nil
	}
}

func (s *ProviderService) healthCheckHosted(ctx context.Context, p domain.Provider) domain.HealthResult {
	if p.Endpoint == "" {
		return domain.HealthResult{OK: false, Message: "no endpoint configured"}
	}
	apiKey, err := s.repo.GetAPIKey(p.ID)
	if err != nil {
		return domain.HealthResult{OK: false, Message: "no API key configured"}
	}

	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.Endpoint, nil)
	if err != nil {
		return domain.HealthResult{OK: false, Message: fmt.Sprintf("invalid endpoint: %v", err)}
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return domain.HealthResult{OK: false, Message: fmt.Sprintf("unreachable: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return domain.HealthResult{OK: false, Message: fmt.Sprintf("server error: HTTP %d", resp.StatusCode)}
	}
	return domain.HealthResult{OK: true, Message: fmt.Sprintf("reachable (HTTP %d)", resp.StatusCode)}
}

func (s *ProviderService) healthCheckCLI(ctx context.Context, p domain.Provider) domain.HealthResult {
	resolved, err := exec.LookPath(p.CLIPath)
	if err != nil {
		return domain.HealthResult{OK: false, Message: "executable not found"}
	}

	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, resolved, "--version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return domain.HealthResult{OK: false, Message: fmt.Sprintf("invocation failed: %v", err)}
	}
	return domain.HealthResult{OK: true, Message: fmt.Sprintf("ok: %s", firstLine(string(out)))}
}

// Invoke runs a prompt against a provider (or the configured default when
// providerID is 0), returning its raw text output. Backs spec-generation
// today and specflow implementation runs later - callers don't need to know
// whether they're talking to a hosted API or a local CLI agent.
func (s *ProviderService) Invoke(ctx context.Context, providerID int64, prompt string) (string, error) {
	p, err := s.resolveProvider(ctx, providerID)
	if err != nil {
		return "", err
	}

	switch p.Kind {
	case domain.ProviderKindHosted:
		return s.invokeHosted(ctx, p, prompt)
	case domain.ProviderKindCLI:
		return s.invokeCLI(ctx, p, prompt)
	default:
		return "", fmt.Errorf("unknown provider kind %q", p.Kind)
	}
}

func (s *ProviderService) resolveProvider(ctx context.Context, providerID int64) (domain.Provider, error) {
	if providerID != 0 {
		return s.Get(ctx, providerID)
	}
	providers, err := s.repo.List(ctx)
	if err != nil {
		return domain.Provider{}, err
	}
	for _, p := range providers {
		if p.IsDefault {
			return p, nil
		}
	}
	return domain.Provider{}, ErrProviderNotFound
}

func (s *ProviderService) invokeHosted(ctx context.Context, p domain.Provider, prompt string) (string, error) {
	apiKey, err := s.repo.GetAPIKey(p.ID)
	if err != nil {
		return "", fmt.Errorf("no API key configured for provider %q", p.Name)
	}

	body, err := json.Marshal(map[string]any{"model": p.Model, "prompt": prompt})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.Endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("provider request failed: %w", err)
	}
	defer resp.Body.Close()

	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("provider returned HTTP %d: %s", resp.StatusCode, string(out))
	}
	return string(out), nil
}

func (s *ProviderService) invokeCLI(ctx context.Context, p domain.Provider, prompt string) (string, error) {
	var out bytes.Buffer
	err := s.invokeCLIStreaming(ctx, p, prompt, func(chunk string) { out.WriteString(chunk) })
	return out.String(), err
}

// InvokeStreaming runs prompt against providerID, calling onChunk as output
// becomes available instead of waiting for completion - used by Specflow so
// a running item's log is near-live rather than appearing all at once when
// the process exits. Hosted providers have no incremental response in this
// abstraction (a single HTTP round trip), so onChunk fires once with the
// full body; CLI providers stream their stdout/stderr line by line as the
// process runs, and ctx cancellation kills the underlying process (Cancel).
func (s *ProviderService) InvokeStreaming(ctx context.Context, providerID int64, prompt string, onChunk func(string)) error {
	p, err := s.resolveProvider(ctx, providerID)
	if err != nil {
		return err
	}
	switch p.Kind {
	case domain.ProviderKindHosted:
		out, err := s.invokeHosted(ctx, p, prompt)
		if err != nil {
			return err
		}
		onChunk(out)
		return nil
	case domain.ProviderKindCLI:
		return s.invokeCLIStreaming(ctx, p, prompt, onChunk)
	default:
		return fmt.Errorf("unknown provider kind %q", p.Kind)
	}
}

// invokeCLIStreaming runs the CLI provider, calling onChunk as stdout/stderr
// lines arrive. onChunk may be called concurrently (stdout and stderr are
// drained on separate goroutines) - callers must synchronize it themselves.
func (s *ProviderService) invokeCLIStreaming(ctx context.Context, p domain.Provider, prompt string, onChunk func(string)) error {
	resolved, err := exec.LookPath(p.CLIPath)
	if err != nil {
		return ErrProviderExecutableNotFound
	}
	args := append([]string{}, p.CLIArgs...)
	args = append(args, prompt)
	cmd := exec.CommandContext(ctx, resolved, args...)
	setNewProcessGroup(cmd)
	// Override CommandContext's default cancel (kills only cmd.Process) with
	// a process-group kill, so a script's own child processes die too -
	// otherwise an orphaned child keeps the stdout/stderr pipes open and
	// the drain goroutines below never see EOF.
	cmd.Cancel = func() error { return killProcessGroup(cmd) }

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start CLI provider: %w", err)
	}

	var wg sync.WaitGroup
	drain := func(r io.Reader) {
		defer wg.Done()
		scanner := bufio.NewScanner(r)
		scanner.Buffer(make([]byte, 64*1024), 1024*1024)
		for scanner.Scan() {
			onChunk(scanner.Text() + "\n")
		}
	}
	wg.Add(2)
	go drain(stdout)
	go drain(stderr)
	wg.Wait()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("CLI provider failed: %w", err)
	}
	return nil
}

func firstLine(s string) string {
	for i, r := range s {
		if r == '\n' {
			return s[:i]
		}
	}
	return s
}
