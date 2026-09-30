package domain

import "time"

// DebtStatus is a tech-debt item's lifecycle state.
type DebtStatus string

const (
	DebtStatusOpen       DebtStatus = "open"
	DebtStatusInProgress DebtStatus = "in_progress"
	DebtStatusResolved   DebtStatus = "resolved"
)

// ArtifactRef optionally links a review or debt item to a spec or change.
type ArtifactRef struct {
	Kind       string `json:"kind"` // "spec" or "change"
	Capability string `json:"capability,omitempty"`
	ChangeName string `json:"changeName,omitempty"`
}

// Review is a review note tied to a project, optionally to a specific spec/change.
type Review struct {
	ID          string      `json:"id"`
	ProjectPath string      `json:"projectPath"`
	ProjectName string      `json:"projectName"`
	Title       string      `json:"title"`
	Body        string      `json:"body"`
	Link        ArtifactRef `json:"link"`
	CreatedAt   time.Time   `json:"createdAt"`
}

// DebtItem is a tracked piece of tech debt, optionally linked to the
// spec/change that would resolve it.
type DebtItem struct {
	ID          string      `json:"id"`
	ProjectPath string      `json:"projectPath"`
	ProjectName string      `json:"projectName"`
	Title       string      `json:"title"`
	Body        string      `json:"body"`
	Status      DebtStatus  `json:"status"`
	Link        ArtifactRef `json:"link"`
	CreatedAt   time.Time   `json:"createdAt"`
}
