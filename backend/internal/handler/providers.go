package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/davidlima/openspec-studio/backend/internal/domain"
	"github.com/davidlima/openspec-studio/backend/internal/service"
)

type ProviderHandler struct {
	providers *service.ProviderService
}

func NewProviderHandler(providers *service.ProviderService) *ProviderHandler {
	return &ProviderHandler{providers: providers}
}

type providerResponse struct {
	ID        int64    `json:"id"`
	Name      string   `json:"name"`
	Kind      string   `json:"kind"`
	Model     string   `json:"model,omitempty"`
	Endpoint  string   `json:"endpoint,omitempty"`
	CLIPath   string   `json:"cliPath,omitempty"`
	CLIArgs   []string `json:"cliArgs,omitempty"`
	IsDefault bool     `json:"isDefault"`
	HasAPIKey bool     `json:"hasApiKey"`
	CreatedAt string   `json:"createdAt"`
}

func toProviderResponse(p domain.Provider) providerResponse {
	return providerResponse{
		ID:        p.ID,
		Name:      p.Name,
		Kind:      string(p.Kind),
		Model:     p.Model,
		Endpoint:  p.Endpoint,
		CLIPath:   p.CLIPath,
		CLIArgs:   p.CLIArgs,
		IsDefault: p.IsDefault,
		HasAPIKey: p.HasAPIKey,
		CreatedAt: p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// List godoc
//
//	@Summary		List configured AI providers
//	@Tags			ai-providers
//	@Produce		json
//	@Success		200	{array}	providerResponse
//	@Router			/ai-providers [get]
func (h *ProviderHandler) List(c *gin.Context) {
	providers, err := h.providers.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}
	resp := make([]providerResponse, 0, len(providers))
	for _, p := range providers {
		resp = append(resp, toProviderResponse(p))
	}
	c.JSON(http.StatusOK, resp)
}

type createProviderRequest struct {
	Name      string   `json:"name" binding:"required"`
	Kind      string   `json:"kind" binding:"required"` // "hosted" or "cli"
	Model     string   `json:"model"`
	Endpoint  string   `json:"endpoint"`
	APIKey    string   `json:"apiKey"`
	CLIPath   string   `json:"cliPath"`
	CLIArgs   []string `json:"cliArgs"`
	IsDefault bool     `json:"isDefault"`
}

// Create godoc
//
//	@Summary		Configure a new AI provider
//	@Tags			ai-providers
//	@Accept			json
//	@Produce		json
//	@Param			body	body		createProviderRequest	true	"Provider config"
//	@Success		201		{object}	providerResponse
//	@Failure		400		{object}	errorResponse	"invalid kind, or CLI executable not found"
//	@Router			/ai-providers [post]
func (h *ProviderHandler) Create(c *gin.Context) {
	var req createProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	var (
		p   domain.Provider
		err error
	)
	switch domain.ProviderKind(req.Kind) {
	case domain.ProviderKindHosted:
		p, err = h.providers.CreateHosted(c.Request.Context(), req.Name, req.Model, req.Endpoint, req.APIKey, req.IsDefault)
	case domain.ProviderKindCLI:
		p, err = h.providers.CreateCLI(c.Request.Context(), req.Name, req.CLIPath, req.CLIArgs, req.IsDefault)
	default:
		c.JSON(http.StatusBadRequest, errorResponse{Error: "kind must be 'hosted' or 'cli'"})
		return
	}
	if err != nil {
		respondProviderError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toProviderResponse(p))
}

type updateProviderRequest struct {
	Name     string   `json:"name" binding:"required"`
	Model    string   `json:"model"`
	Endpoint string   `json:"endpoint"`
	APIKey   *string  `json:"apiKey,omitempty"`
	CLIPath  string   `json:"cliPath"`
	CLIArgs  []string `json:"cliArgs"`
}

// Update godoc
//
//	@Summary		Update an AI provider
//	@Tags			ai-providers
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int						true	"Provider id"
//	@Param			body	body		updateProviderRequest	true	"Updated config"
//	@Success		200		{object}	providerResponse
//	@Router			/ai-providers/{id} [patch]
func (h *ProviderHandler) Update(c *gin.Context) {
	id, ok := parseInt64Param(c, "id")
	if !ok {
		return
	}
	var req updateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	p, err := h.providers.Update(c.Request.Context(), id, req.Name, req.Model, req.Endpoint, req.CLIPath, req.CLIArgs, req.APIKey)
	if err != nil {
		respondProviderError(c, err)
		return
	}
	c.JSON(http.StatusOK, toProviderResponse(p))
}

// Delete godoc
//
//	@Summary		Delete an AI provider
//	@Tags			ai-providers
//	@Param			id	path	int	true	"Provider id"
//	@Success		204
//	@Router			/ai-providers/{id} [delete]
func (h *ProviderHandler) Delete(c *gin.Context) {
	id, ok := parseInt64Param(c, "id")
	if !ok {
		return
	}
	if err := h.providers.Delete(c.Request.Context(), id); err != nil {
		respondProviderError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// SetDefault godoc
//
//	@Summary		Set the default AI provider
//	@Tags			ai-providers
//	@Param			id	path	int	true	"Provider id"
//	@Success		204
//	@Router			/ai-providers/{id}/default [post]
func (h *ProviderHandler) SetDefault(c *gin.Context) {
	id, ok := parseInt64Param(c, "id")
	if !ok {
		return
	}
	if err := h.providers.SetDefault(c.Request.Context(), id); err != nil {
		respondProviderError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// HealthCheck godoc
//
//	@Summary		Run a connectivity health check against a provider
//	@Tags			ai-providers
//	@Produce		json
//	@Param			id	path		int	true	"Provider id"
//	@Success		200	{object}	domain.HealthResult
//	@Router			/ai-providers/{id}/health [post]
func (h *ProviderHandler) HealthCheck(c *gin.Context) {
	id, ok := parseInt64Param(c, "id")
	if !ok {
		return
	}
	result, err := h.providers.HealthCheck(c.Request.Context(), id)
	if err != nil {
		respondProviderError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func parseInt64Param(c *gin.Context, name string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid " + name})
		return 0, false
	}
	return id, true
}

func respondProviderError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrProviderNotFound):
		c.JSON(http.StatusNotFound, errorResponse{Error: err.Error()})
	case errors.Is(err, service.ErrProviderExecutableNotFound):
		c.JSON(http.StatusBadRequest, errorResponse{Error: "executable not found"})
	default:
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
	}
}
