package domain

import "time"

type ProviderKind string

const (
	ProviderKindHosted ProviderKind = "hosted"
	ProviderKindCLI    ProviderKind = "cli"
)

// Provider is a configured AI provider (hosted API or local CLI agent).
// The API key itself is never held here - it lives in OS keychain / an
// encrypted local file, referenced by ID.
type Provider struct {
	ID        int64        `json:"id"`
	Name      string       `json:"name"`
	Kind      ProviderKind `json:"kind"`
	Model     string       `json:"model,omitempty"`
	Endpoint  string       `json:"endpoint,omitempty"`
	CLIPath   string       `json:"cliPath,omitempty"`
	CLIArgs   []string     `json:"cliArgs,omitempty"`
	IsDefault bool         `json:"isDefault"`
	CreatedAt time.Time    `json:"createdAt"`
	HasAPIKey bool         `json:"hasApiKey"`
}

// HealthResult is the outcome of a provider connectivity check.
type HealthResult struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}
