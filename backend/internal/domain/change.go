package domain

// ChangeStatus is a change's position on the Kanban board. Done and Archived
// are deliberately distinct - archived is a separate lifecycle event (moved
// under openspec/changes/archive/), not implied by task completion alone.
type ChangeStatus string

const (
	ChangeStatusDraft      ChangeStatus = "draft"
	ChangeStatusTodo       ChangeStatus = "todo"
	ChangeStatusInProgress ChangeStatus = "in_progress"
	ChangeStatusDone       ChangeStatus = "done"
	ChangeStatusArchived   ChangeStatus = "archived"
)

// Change is a single OpenSpec change proposal as tracked by the Studio.
type Change struct {
	Name        string       `json:"name"`
	ProjectPath string       `json:"projectPath"`
	ProjectName string       `json:"projectName"`
	Status      ChangeStatus `json:"status"`
	HasProposal bool         `json:"hasProposal"`
	HasDesign   bool         `json:"hasDesign"`
	HasSpecs    bool         `json:"hasSpecs"`
	HasTasks    bool         `json:"hasTasks"`
	TasksTotal  int          `json:"tasksTotal"`
	TasksDone   int          `json:"tasksDone"`
}

func (c Change) ID(projectID int64) string {
	return EncodeChangeID(projectID, c.Name)
}

// ChangeDetail carries a change's full artifact content, for the Kanban card
// detail view (Proposal/Specs/Tasks/Design tabs).
type ChangeDetail struct {
	Change
	ProposalMarkdown string            `json:"proposalMarkdown,omitempty"`
	DesignMarkdown   string            `json:"designMarkdown,omitempty"`
	TasksMarkdown    string            `json:"tasksMarkdown,omitempty"`
	Specs            map[string]string `json:"specs,omitempty"` // capability -> spec.md content
}

// ComputeStatus derives Draft/Todo/In Progress/Done from proposal/tasks
// presence and checkbox state, matching openspec-ui's status model. Archived
// is set by the caller (based on directory location), not here.
func ComputeStatus(hasProposal, hasTasks bool, tasksTotal, tasksDone int) ChangeStatus {
	switch {
	case !hasTasks:
		if hasProposal {
			return ChangeStatusDraft
		}
		return ChangeStatusDraft
	case tasksTotal > 0 && tasksDone == tasksTotal:
		return ChangeStatusDone
	case tasksDone > 0:
		return ChangeStatusInProgress
	default:
		return ChangeStatusTodo
	}
}
