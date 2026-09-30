package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/davidlima/openspec-studio/backend/internal/service"
)

type OverviewHandler struct {
	overview *service.OverviewService
}

func NewOverviewHandler(overview *service.OverviewService) *OverviewHandler {
	return &OverviewHandler{overview: overview}
}

// Get godoc
//
//	@Summary		Get cross-project overview
//	@Description	Aggregates spec/change progress across every registered project (or a single project via ?projectId=): per-project spec/change-stage counts, a progress ratio, and a broken-cross-project-dependency count, plus aggregate totals and a most-recently-updated list of {projectId, projectName, lastOpenedAt} entries.
//	@Tags			overview
//	@Produce		json
//	@Param			projectId	query		int	false	"Scope the overview to a single registered project"
//	@Success		200			{object}	domain.Overview
//	@Failure		400			{object}	errorResponse
//	@Failure		500			{object}	errorResponse
//	@Router			/overview [get]
func (h *OverviewHandler) Get(c *gin.Context) {
	var projectID *int64
	if q := c.Query("projectId"); q != "" {
		id, err := strconv.ParseInt(q, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid projectId"})
			return
		}
		projectID = &id
	}

	overview, err := h.overview.GetOverview(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, overview)
}
