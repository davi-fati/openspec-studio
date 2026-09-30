package domain

import "time"

type FlowStatus string

const (
	FlowStatusPending   FlowStatus = "pending"
	FlowStatusRunning   FlowStatus = "running"
	FlowStatusSucceeded FlowStatus = "succeeded"
	FlowStatusFailed    FlowStatus = "failed"
	FlowStatusCancelled FlowStatus = "cancelled"
)

type FlowItemStatus string

const (
	FlowItemPending   FlowItemStatus = "pending"
	FlowItemRunning   FlowItemStatus = "running"
	FlowItemSucceeded FlowItemStatus = "succeeded"
	FlowItemFailed    FlowItemStatus = "failed"
)

// FlowItem is one change's slot within a Specflow's ordered sequence.
type FlowItem struct {
	ID         int64          `json:"id"`
	Position   int            `json:"position"`
	ChangeName string         `json:"changeName"`
	Status     FlowItemStatus `json:"status"`
	Log        string         `json:"log"`
	StartedAt  *time.Time     `json:"startedAt,omitempty"`
	FinishedAt *time.Time     `json:"finishedAt,omitempty"`
}

// Flow is a scheduled, ordered sequence of fully-planned changes to
// implement automatically in one project via a configured AI provider.
type Flow struct {
	ID          int64      `json:"id"`
	ProjectID   int64      `json:"projectId"`
	ProjectName string     `json:"projectName"`
	ProviderID  int64      `json:"providerId"`
	ScheduledAt time.Time  `json:"scheduledAt"`
	Status      FlowStatus `json:"status"`
	Missed      bool       `json:"missed"`
	CreatedAt   time.Time  `json:"createdAt"`
	Items       []FlowItem `json:"items"`
}
