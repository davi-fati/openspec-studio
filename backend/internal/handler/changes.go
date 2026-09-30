package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/davidlima/openspec-studio/backend/internal/domain"
	"github.com/davidlima/openspec-studio/backend/internal/service"
)

type ChangeHandler struct {
	changes  *service.ChangeService
	projects *service.ProjectService
}

func NewChangeHandler(changes *service.ChangeService, projects *service.ProjectService) *ChangeHandler {
	return &ChangeHandler{changes: changes, projects: projects}
}

type changeResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ProjectID   int64  `json:"projectId"`
	ProjectName string `json:"projectName"`
	Status      string `json:"status"`
	HasProposal bool   `json:"hasProposal"`
	HasDesign   bool   `json:"hasDesign"`
	HasSpecs    bool   `json:"hasSpecs"`
	HasTasks    bool   `json:"hasTasks"`
	TasksTotal  int    `json:"tasksTotal"`
	TasksDone   int    `json:"tasksDone"`
}

func toChangeResponse(projectID int64, ch domain.Change) changeResponse {
	return changeResponse{
		ID:          domain.EncodeChangeID(projectID, ch.Name),
		Name:        ch.Name,
		ProjectID:   projectID,
		ProjectName: ch.ProjectName,
		Status:      string(ch.Status),
		HasProposal: ch.HasProposal,
		HasDesign:   ch.HasDesign,
		HasSpecs:    ch.HasSpecs,
		HasTasks:    ch.HasTasks,
		TasksTotal:  ch.TasksTotal,
		TasksDone:   ch.TasksDone,
	}
}

// List godoc
//
//	@Summary		List changes
//	@Description	Lists every change across every registered, available project (or a single project via ?projectId=), including archived changes, with Kanban lifecycle status computed.
//	@Tags			changes
//	@Produce		json
//	@Param			projectId	query		int	false	"Filter to a single registered project"
//	@Success		200			{array}		changeResponse
//	@Failure		500			{object}	errorResponse
//	@Router			/changes [get]
func (h *ChangeHandler) List(c *gin.Context) {
	changes, err := h.changes.ListChanges(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}
	byPath, err := projectIDsByPath(c, h.projects)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}

	var filterProjectID int64 = -1
	if q := c.Query("projectId"); q != "" {
		id, err := strconv.ParseInt(q, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid projectId"})
			return
		}
		filterProjectID = id
	}

	resp := make([]changeResponse, 0, len(changes))
	for _, ch := range changes {
		id, ok := byPath[ch.ProjectPath]
		if !ok {
			continue
		}
		if filterProjectID != -1 && id != filterProjectID {
			continue
		}
		resp = append(resp, toChangeResponse(id, ch))
	}
	c.JSON(http.StatusOK, resp)
}

type changeDetailResponse struct {
	changeResponse
	ProposalMarkdown string            `json:"proposalMarkdown,omitempty"`
	DesignMarkdown   string            `json:"designMarkdown,omitempty"`
	TasksMarkdown    string            `json:"tasksMarkdown,omitempty"`
	Specs            map[string]string `json:"specs,omitempty"`
}

// Get godoc
//
//	@Summary		Get change detail
//	@Description	Returns a change's full artifact content (proposal/design/tasks/specs) for the Kanban card detail view.
//	@Tags			changes
//	@Produce		json
//	@Param			id	path		string	true	"Change id"
//	@Success		200	{object}	changeDetailResponse
//	@Failure		404	{object}	errorResponse
//	@Router			/changes/{id} [get]
func (h *ChangeHandler) Get(c *gin.Context) {
	projectID, name, err := domain.DecodeChangeID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid change id"})
		return
	}
	detail, err := h.changes.GetChangeDetail(c.Request.Context(), projectID, name)
	if err != nil {
		if errors.Is(err, service.ErrChangeNotFound) || errors.Is(err, service.ErrProjectNotFound) {
			c.JSON(http.StatusNotFound, errorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, changeDetailResponse{
		changeResponse:   toChangeResponse(projectID, detail.Change),
		ProposalMarkdown: detail.ProposalMarkdown,
		DesignMarkdown:   detail.DesignMarkdown,
		TasksMarkdown:    detail.TasksMarkdown,
		Specs:            detail.Specs,
	})
}
