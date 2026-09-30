package service

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/davidlima/openspec-studio/backend/internal/domain"
)

var errTestCredentialNotFound = errors.New("credential not found")

// stubProviderRepository is an in-memory ProviderRepository for tests, so
// invocation/health-check logic can be exercised without SQLite or a real
// keychain.
type stubProviderRepository struct {
	providers map[int64]domain.Provider
	apiKeys   map[int64]string
	nextID    int64
}

func newStubProviderRepository() *stubProviderRepository {
	return &stubProviderRepository{providers: map[int64]domain.Provider{}, apiKeys: map[int64]string{}}
}

func (r *stubProviderRepository) List(ctx context.Context) ([]domain.Provider, error) {
	var out []domain.Provider
	for _, p := range r.providers {
		out = append(out, p)
	}
	return out, nil
}

func (r *stubProviderRepository) Get(ctx context.Context, id int64) (domain.Provider, bool, error) {
	p, ok := r.providers[id]
	return p, ok, nil
}

func (r *stubProviderRepository) Create(ctx context.Context, p domain.Provider, apiKey string) (domain.Provider, error) {
	r.nextID++
	p.ID = r.nextID
	p.CreatedAt = time.Now().UTC()
	r.providers[p.ID] = p
	if apiKey != "" {
		r.apiKeys[p.ID] = apiKey
	}
	return p, nil
}

func (r *stubProviderRepository) Update(ctx context.Context, id int64, p domain.Provider, apiKey *string) (domain.Provider, error) {
	existing := r.providers[id]
	existing.Name = p.Name
	existing.Model = p.Model
	existing.Endpoint = p.Endpoint
	existing.CLIPath = p.CLIPath
	existing.CLIArgs = p.CLIArgs
	r.providers[id] = existing
	if apiKey != nil {
		r.apiKeys[id] = *apiKey
	}
	return existing, nil
}

func (r *stubProviderRepository) Delete(ctx context.Context, id int64) error {
	delete(r.providers, id)
	delete(r.apiKeys, id)
	return nil
}

func (r *stubProviderRepository) SetDefault(ctx context.Context, id int64) error {
	for pid, p := range r.providers {
		p.IsDefault = pid == id
		r.providers[pid] = p
	}
	return nil
}

func (r *stubProviderRepository) GetAPIKey(id int64) (string, error) {
	key, ok := r.apiKeys[id]
	if !ok {
		return "", errTestCredentialNotFound
	}
	return key, nil
}

func TestProviderServiceInvokeHosted(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Write([]byte("## Purpose\nGenerated purpose.\n\n## Requirements\n\n### Requirement: Do thing\nThe system SHALL do the thing.\n\n#### Scenario: It works\n- **WHEN** x\n- **THEN** y\n"))
	}))
	defer server.Close()

	ctx := context.Background()
	repo := newStubProviderRepository()
	svc := NewProviderService(repo)

	p, err := svc.CreateHosted(ctx, "test-hosted", "test-model", server.URL, "test-key", true)
	if err != nil {
		t.Fatalf("CreateHosted: %v", err)
	}

	out, err := svc.Invoke(ctx, p.ID, "draft something")
	if err != nil {
		t.Fatalf("Invoke: %v", err)
	}
	if out == "" {
		t.Fatal("expected non-empty response from mocked hosted provider")
	}

	health, err := svc.HealthCheck(ctx, p.ID)
	if err != nil {
		t.Fatalf("HealthCheck: %v", err)
	}
	if !health.OK {
		t.Fatalf("expected healthy mocked hosted provider, got %+v", health)
	}
}

func TestProviderServiceInvokeCLI(t *testing.T) {
	ctx := context.Background()
	repo := newStubProviderRepository()
	svc := NewProviderService(repo)

	// A tiny real script stands in for a CLI agent: it just echoes its args.
	scriptPath := filepath.Join(t.TempDir(), "fake-agent.sh")
	script := "#!/bin/sh\necho \"received: $*\"\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake CLI: %v", err)
	}
	if runtime.GOOS == "windows" {
		t.Skip("fake shell script CLI provider not supported on windows")
	}

	p, err := svc.CreateCLI(ctx, "test-cli", scriptPath, []string{"--flag"}, false)
	if err != nil {
		t.Fatalf("CreateCLI: %v", err)
	}

	out, err := svc.Invoke(ctx, p.ID, "draft something")
	if err != nil {
		t.Fatalf("Invoke: %v", err)
	}
	if out == "" {
		t.Fatal("expected non-empty output from fake CLI provider")
	}

	health, err := svc.HealthCheck(ctx, p.ID)
	if err != nil {
		t.Fatalf("HealthCheck: %v", err)
	}
	_ = health // the fake script doesn't implement --version; just verifying no panic/error path

	// A nonexistent executable path must be rejected at creation time, not saved silently.
	if _, err := svc.CreateCLI(ctx, "broken-cli", "/nonexistent/path/to/agent", nil, false); err != ErrProviderExecutableNotFound {
		t.Fatalf("expected ErrProviderExecutableNotFound, got %v", err)
	}
}
