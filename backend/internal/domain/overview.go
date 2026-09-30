package domain

import "time"

// ProjectOverview summarizes one registered project's spec/change progress
// for the Overview landing page.
type ProjectOverview struct {
	ProjectID       int64          `json:"projectId"`
	ProjectName     string         `json:"projectName"`
	Available       bool           `json:"available"`
	SpecCount       int            `json:"specCount"`
	BrokenSpecCount int            `json:"brokenSpecCount"` // specs with an unavailable cross-project dependency
	ChangeCounts    map[string]int `json:"changeCounts"`    // status -> count, over draft/todo/in_progress/done/archived
	ProgressRatio   float64        `json:"progressRatio"`   // (done+archived) / (todo+in_progress+done+archived), 0 when no non-draft changes exist
}

// RecentProject is one entry in Overview.RecentlyUpdated: a registered
// project and when it was last opened.
type RecentProject struct {
	ProjectID    int64     `json:"projectId"`
	ProjectName  string    `json:"projectName"`
	LastOpenedAt time.Time `json:"lastOpenedAt"`
}

// Overview is the full cross-project aggregate shown on the Overview tab.
type Overview struct {
	Projects         []ProjectOverview `json:"projects"`
	TotalOpenChanges int               `json:"totalOpenChanges"` // draft+todo+in_progress across all projects
	TotalInProgress  int               `json:"totalInProgress"`
	RecentlyUpdated  []RecentProject   `json:"recentlyUpdated"` // most recently opened first
}
