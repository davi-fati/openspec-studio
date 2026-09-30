package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/davidlima/openspec-studio/backend/internal/domain"
	"github.com/davidlima/openspec-studio/backend/internal/service"
)

type ContextHandler struct {
	context *service.ContextService
}

func NewContextHandler(context *service.ContextService) *ContextHandler {
	return &ContextHandler{context: context}
}

type artifactRefDTO struct {
	Kind       string `json:"kind,omitempty"`
	Capability string `json:"capability,omitempty"`
	ChangeName string `json:"changeName,omitempty"`
}

func toRefDTO(r domain.ArtifactRef) artifactRefDTO {
	return artifactRefDTO{Kind: r.Kind, Capability: r.Capability, ChangeName: r.ChangeName}
}

func (d artifactRefDTO) toDomain() domain.ArtifactRef {
	return domain.ArtifactRef{Kind: d.Kind, Capability: d.Capability, ChangeName: d.ChangeName}
}

type reviewResponse struct {
	ID        string         `json:"id"`
	ProjectID int64          `json:"projectId"`
	Title     string         `json:"title"`
	Body      string         `json:"body"`
	Link      artifactRefDTO `json:"link"`
	CreatedAt string         `json:"createdAt"`
}

func toReviewResponse(projectID int64, r domain.Review) reviewResponse {
	return reviewResponse{
		ID:        r.ID,
		ProjectID: projectID,
		Title:     r.Title,
		Body:      r.Body,
		Link:      toRefDTO(r.Link),
		CreatedAt: r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

type debtResponse struct {
	ID        string         `json:"id"`
	ProjectID int64          `json:"projectId"`
	Title     string         `json:"title"`
	Body      string         `json:"body"`
	Status    string         `json:"status"`
	Link      artifactRefDTO `json:"link"`
	CreatedAt string         `json:"createdAt"`
}

func toDebtResponse(projectID int64, d domain.DebtItem) debtResponse {
	return debtResponse{
		ID:        d.ID,
		ProjectID: projectID,
		Title:     d.Title,
		Body:      d.Body,
		Status:    string(d.Status),
		Link:      toRefDTO(d.Link),
		CreatedAt: d.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func projectIDParam(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid project id"})
		return 0, false
	}
	return id, true
}

// ListReviews godoc
//
//	@Summary		List a project's reviews
//	@Tags			context
//	@Produce		json
//	@Param			id	path		int	true	"Project id"
//	@Success		200	{array}		reviewResponse
//	@Router			/projects/{id}/reviews [get]
func (h *ContextHandler) ListReviews(c *gin.Context) {
	projectID, ok := projectIDParam(c)
	if !ok {
		return
	}
	reviews, err := h.context.ListReviews(c.Request.Context(), projectID)
	if err != nil {
		respondContextError(c, err)
		return
	}
	resp := make([]reviewResponse, 0, len(reviews))
	for _, r := range reviews {
		resp = append(resp, toReviewResponse(projectID, r))
	}
	c.JSON(http.StatusOK, resp)
}

type createReviewRequest struct {
	Title string         `json:"title" binding:"required"`
	Body  string         `json:"body"`
	Link  artifactRefDTO `json:"link"`
}

// CreateReview godoc
//
//	@Summary		Create a review note
//	@Tags			context
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"Project id"
//	@Param			body	body		createReviewRequest	true	"Review"
//	@Success		201		{object}	reviewResponse
//	@Router			/projects/{id}/reviews [post]
func (h *ContextHandler) CreateReview(c *gin.Context) {
	projectID, ok := projectIDParam(c)
	if !ok {
		return
	}
	var req createReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	review, err := h.context.CreateReview(c.Request.Context(), projectID, req.Title, req.Body, req.Link.toDomain())
	if err != nil {
		respondContextError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toReviewResponse(projectID, review))
}

// DeleteReview godoc
//
//	@Summary		Delete a review note
//	@Tags			context
//	@Param			id			path	int		true	"Project id"
//	@Param			reviewId	path	string	true	"Review id"
//	@Success		204
//	@Router			/projects/{id}/reviews/{reviewId} [delete]
func (h *ContextHandler) DeleteReview(c *gin.Context) {
	projectID, ok := projectIDParam(c)
	if !ok {
		return
	}
	if err := h.context.DeleteReview(c.Request.Context(), projectID, c.Param("reviewId")); err != nil {
		respondContextError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListDebts godoc
//
//	@Summary		List a project's tech-debt items
//	@Tags			context
//	@Produce		json
//	@Param			id	path		int	true	"Project id"
//	@Success		200	{array}		debtResponse
//	@Router			/projects/{id}/debts [get]
func (h *ContextHandler) ListDebts(c *gin.Context) {
	projectID, ok := projectIDParam(c)
	if !ok {
		return
	}
	debts, err := h.context.ListDebts(c.Request.Context(), projectID)
	if err != nil {
		respondContextError(c, err)
		return
	}
	resp := make([]debtResponse, 0, len(debts))
	for _, d := range debts {
		resp = append(resp, toDebtResponse(projectID, d))
	}
	c.JSON(http.StatusOK, resp)
}

type createDebtRequest struct {
	Title string         `json:"title" binding:"required"`
	Body  string         `json:"body"`
	Link  artifactRefDTO `json:"link"`
}

// CreateDebt godoc
//
//	@Summary		Create a tech-debt item
//	@Tags			context
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"Project id"
//	@Param			body	body		createDebtRequest	true	"Debt item"
//	@Success		201		{object}	debtResponse
//	@Router			/projects/{id}/debts [post]
func (h *ContextHandler) CreateDebt(c *gin.Context) {
	projectID, ok := projectIDParam(c)
	if !ok {
		return
	}
	var req createDebtRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	debt, err := h.context.CreateDebt(c.Request.Context(), projectID, req.Title, req.Body, req.Link.toDomain())
	if err != nil {
		respondContextError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toDebtResponse(projectID, debt))
}

type updateDebtRequest struct {
	Status string `json:"status" binding:"required"`
}

// UpdateDebt godoc
//
//	@Summary		Update a tech-debt item's status
//	@Tags			context
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"Project id"
//	@Param			debtId	path		string				true	"Debt id"
//	@Param			body	body		updateDebtRequest	true	"New status"
//	@Success		200		{object}	debtResponse
//	@Router			/projects/{id}/debts/{debtId} [patch]
func (h *ContextHandler) UpdateDebt(c *gin.Context) {
	projectID, ok := projectIDParam(c)
	if !ok {
		return
	}
	var req updateDebtRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	status := domain.DebtStatus(req.Status)
	switch status {
	case domain.DebtStatusOpen, domain.DebtStatusInProgress, domain.DebtStatusResolved:
	default:
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid status"})
		return
	}
	debt, err := h.context.UpdateDebtStatus(c.Request.Context(), projectID, c.Param("debtId"), status)
	if err != nil {
		respondContextError(c, err)
		return
	}
	c.JSON(http.StatusOK, toDebtResponse(projectID, debt))
}

// DeleteDebt godoc
//
//	@Summary		Delete a tech-debt item
//	@Tags			context
//	@Param			id		path	int		true	"Project id"
//	@Param			debtId	path	string	true	"Debt id"
//	@Success		204
//	@Router			/projects/{id}/debts/{debtId} [delete]
func (h *ContextHandler) DeleteDebt(c *gin.Context) {
	projectID, ok := projectIDParam(c)
	if !ok {
		return
	}
	if err := h.context.DeleteDebt(c.Request.Context(), projectID, c.Param("debtId")); err != nil {
		respondContextError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func respondContextError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrContextItemNotFound), errors.Is(err, service.ErrProjectNotFound):
		c.JSON(http.StatusNotFound, errorResponse{Error: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
	}
}
