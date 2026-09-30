package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/davidlima/openspec-studio/backend/internal/domain"
	"github.com/davidlima/openspec-studio/backend/internal/service"
)

type SpecflowHandler struct {
	flows *service.SpecflowService
}

func NewSpecflowHandler(flows *service.SpecflowService) *SpecflowHandler {
	return &SpecflowHandler{flows: flows}
}

type flowItemResponse struct {
	ID         int64   `json:"id"`
	Position   int     `json:"position"`
	ChangeName string  `json:"changeName"`
	Status     string  `json:"status"`
	Log        string  `json:"log"`
	StartedAt  *string `json:"startedAt,omitempty"`
	FinishedAt *string `json:"finishedAt,omitempty"`
}

type flowResponse struct {
	ID          int64               `json:"id"`
	ProjectID   int64               `json:"projectId"`
	ProjectName string              `json:"projectName"`
	ProviderID  int64               `json:"providerId"`
	ScheduledAt string              `json:"scheduledAt"`
	Status      string              `json:"status"`
	Missed      bool                `json:"missed"`
	CreatedAt   string              `json:"createdAt"`
	Items       []flowItemResponse  `json:"items"`
}

const timeFormat = "2006-01-02T15:04:05Z07:00"

func toFlowResponse(f domain.Flow) flowResponse {
	items := make([]flowItemResponse, 0, len(f.Items))
	for _, it := range f.Items {
		resp := flowItemResponse{ID: it.ID, Position: it.Position, ChangeName: it.ChangeName, Status: string(it.Status), Log: it.Log}
		if it.StartedAt != nil {
			s := it.StartedAt.Format(timeFormat)
			resp.StartedAt = &s
		}
		if it.FinishedAt != nil {
			s := it.FinishedAt.Format(timeFormat)
			resp.FinishedAt = &s
		}
		items = append(items, resp)
	}
	return flowResponse{
		ID: f.ID, ProjectID: f.ProjectID, ProjectName: f.ProjectName, ProviderID: f.ProviderID,
		ScheduledAt: f.ScheduledAt.Format(timeFormat), Status: string(f.Status), Missed: f.Missed,
		CreatedAt: f.CreatedAt.Format(timeFormat), Items: items,
	}
}

// List godoc
//
//	@Summary		List Specflow flows
//	@Description	Lists scheduled flows, optionally filtered to one project.
//	@Tags			specflow
//	@Produce		json
//	@Param			projectId	query		int	false	"Filter to a single registered project"
//	@Success		200			{array}		flowResponse
//	@Router			/specflow/flows [get]
func (h *SpecflowHandler) List(c *gin.Context) {
	var projectID int64
	if q := c.Query("projectId"); q != "" {
		id, err := strconv.ParseInt(q, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid projectId"})
			return
		}
		projectID = id
	}
	flows, err := h.flows.List(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}
	resp := make([]flowResponse, 0, len(flows))
	for _, f := range flows {
		resp = append(resp, toFlowResponse(f))
	}
	c.JSON(http.StatusOK, resp)
}

// Get godoc
//
//	@Summary		Get a Specflow flow
//	@Tags			specflow
//	@Produce		json
//	@Param			id	path		int	true	"Flow id"
//	@Success		200	{object}	flowResponse
//	@Router			/specflow/flows/{id} [get]
func (h *SpecflowHandler) Get(c *gin.Context) {
	id, ok := parseInt64Param(c, "id")
	if !ok {
		return
	}
	f, err := h.flows.Get(c.Request.Context(), id)
	if err != nil {
		respondSpecflowError(c, err)
		return
	}
	c.JSON(http.StatusOK, toFlowResponse(f))
}

type createFlowRequest struct {
	ProjectID   int64    `json:"projectId" binding:"required"`
	ProviderID  int64    `json:"providerId" binding:"required"`
	ScheduledAt string   `json:"scheduledAt" binding:"required"` // RFC3339
	ChangeNames []string `json:"changeNames" binding:"required"`
}

// Create godoc
//
//	@Summary		Schedule a new Specflow flow
//	@Description	Only fully-planned changes can be scheduled, and dependency order is validated before the flow is saved.
//	@Tags			specflow
//	@Accept			json
//	@Produce		json
//	@Param			body	body		createFlowRequest	true	"Flow"
//	@Success		201		{object}	flowResponse
//	@Failure		400		{object}	errorResponse	"incomplete change or unmet dependency"
//	@Router			/specflow/flows [post]
func (h *SpecflowHandler) Create(c *gin.Context) {
	var req createFlowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	scheduledAt, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "scheduledAt must be RFC3339"})
		return
	}
	// Normalize to UTC: SQLite has no native datetime type, so stored
	// timestamps compare lexically - mixed zone offsets would sort wrong
	// against the scheduler's own UTC "now".
	scheduledAt = scheduledAt.UTC()
	flow, err := h.flows.Create(c.Request.Context(), req.ProjectID, req.ProviderID, scheduledAt, req.ChangeNames)
	if err != nil {
		respondSpecflowError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toFlowResponse(flow))
}

type reorderFlowRequest struct {
	ItemIDs []int64 `json:"itemIds" binding:"required"`
}

// Reorder godoc
//
//	@Summary		Reorder a pending flow's items
//	@Tags			specflow
//	@Accept			json
//	@Param			id		path	int					true	"Flow id"
//	@Param			body	body	reorderFlowRequest	true	"New item order"
//	@Success		204
//	@Failure		400	{object}	errorResponse	"flow already started"
//	@Router			/specflow/flows/{id}/reorder [patch]
func (h *SpecflowHandler) Reorder(c *gin.Context) {
	id, ok := parseInt64Param(c, "id")
	if !ok {
		return
	}
	var req reorderFlowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	if err := h.flows.Reorder(c.Request.Context(), id, req.ItemIDs); err != nil {
		respondSpecflowError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Delete godoc
//
//	@Summary		Delete a Specflow flow
//	@Tags			specflow
//	@Param			id	path	int	true	"Flow id"
//	@Success		204
//	@Router			/specflow/flows/{id} [delete]
func (h *SpecflowHandler) Delete(c *gin.Context) {
	id, ok := parseInt64Param(c, "id")
	if !ok {
		return
	}
	if err := h.flows.Delete(c.Request.Context(), id); err != nil {
		respondSpecflowError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Cancel godoc
//
//	@Summary		Cancel a running Specflow flow
//	@Description	Terminates the in-progress provider process and marks the flow cancelled; already-succeeded items are left intact.
//	@Tags			specflow
//	@Param			id	path	int	true	"Flow id"
//	@Success		204
//	@Failure		409	{object}	errorResponse	"flow is not running"
//	@Router			/specflow/flows/{id}/cancel [post]
func (h *SpecflowHandler) Cancel(c *gin.Context) {
	id, ok := parseInt64Param(c, "id")
	if !ok {
		return
	}
	if err := h.flows.Cancel(c.Request.Context(), id); err != nil {
		if errors.Is(err, service.ErrFlowNotRunning) {
			c.JSON(http.StatusConflict, errorResponse{Error: err.Error()})
			return
		}
		respondSpecflowError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func respondSpecflowError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrFlowNotFound), errors.Is(err, service.ErrProjectNotFound):
		c.JSON(http.StatusNotFound, errorResponse{Error: err.Error()})
	case errors.Is(err, service.ErrChangeNotFullyPlanned), errors.Is(err, service.ErrUnmetDependency), errors.Is(err, service.ErrFlowAlreadyStarted):
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
	}
}
