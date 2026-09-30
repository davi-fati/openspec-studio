package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/davidlima/openspec-studio/backend/internal/domain"
	"github.com/davidlima/openspec-studio/backend/internal/service"
)

type SpecHandler struct {
	specs    *service.SpecService
	projects *service.ProjectService
	context  *service.ContextService
}

func NewSpecHandler(specs *service.SpecService, projects *service.ProjectService, context *service.ContextService) *SpecHandler {
	return &SpecHandler{specs: specs, projects: projects, context: context}
}

type specRequirementDTO struct {
	Name string `json:"name" binding:"required"`
	Body string `json:"body"`
}

type specDependencyDTO struct {
	ProjectPath string `json:"projectPath"`
	ProjectName string `json:"projectName"`
	Capability  string `json:"capability"`
	Available   bool   `json:"available"`
}

type specResponse struct {
	ID           string               `json:"id"`
	Capability   string               `json:"capability"`
	ProjectID    int64                `json:"projectId"`
	ProjectName  string               `json:"projectName"`
	Purpose      string               `json:"purpose"`
	Requirements []specRequirementDTO `json:"requirements"`
	Provenance   domain.SpecProvenance `json:"provenance"`
	DependsOn    []specDependencyDTO  `json:"dependsOn"`
	DependedBy   []specDependencyDTO  `json:"dependedBy"`
	LinkedDebts  []debtResponse       `json:"linkedDebts,omitempty"`
	LinkedReviews []reviewResponse    `json:"linkedReviews,omitempty"`
}

func toSpecResponse(projectID int64, s domain.Spec) specResponse {
	reqs := make([]specRequirementDTO, 0, len(s.Requirements))
	for _, r := range s.Requirements {
		reqs = append(reqs, specRequirementDTO{Name: r.Name, Body: r.Body})
	}
	deps := make([]specDependencyDTO, 0, len(s.DependsOn))
	for _, d := range s.DependsOn {
		deps = append(deps, specDependencyDTO{ProjectPath: d.ProjectPath, ProjectName: d.ProjectName, Capability: d.Capability, Available: d.Available})
	}
	dependedBy := make([]specDependencyDTO, 0, len(s.DependedBy))
	for _, d := range s.DependedBy {
		dependedBy = append(dependedBy, specDependencyDTO{ProjectPath: d.ProjectPath, ProjectName: d.ProjectName, Capability: d.Capability, Available: d.Available})
	}
	return specResponse{
		ID:           domain.EncodeSpecID(projectID, s.Capability),
		Capability:   s.Capability,
		ProjectID:    projectID,
		ProjectName:  s.ProjectName,
		Purpose:      s.Purpose,
		Requirements: reqs,
		Provenance:   s.Meta.Provenance,
		DependsOn:    deps,
		DependedBy:   dependedBy,
	}
}

// projectIDForPath is resolved via a second pass over the registry inside
// ListSpecs' caller; handlers instead carry the id explicitly since routes
// are always addressed by id (see decodeSpecID).

// List godoc
//
//	@Summary		List specs
//	@Description	Lists every spec across every registered, available project (or a single project via ?projectId=), with resolved dependency links.
//	@Tags			specs
//	@Produce		json
//	@Param			projectId	query		int	false	"Filter to a single registered project"
//	@Success		200			{array}		specResponse
//	@Failure		400			{object}	errorResponse
//	@Failure		500			{object}	errorResponse
//	@Router			/specs [get]
func (h *SpecHandler) List(c *gin.Context) {
	specs, err := h.specs.ListSpecs(c.Request.Context())
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

	resp := make([]specResponse, 0, len(specs))
	for _, s := range specs {
		id, ok := byPath[s.ProjectPath]
		if !ok {
			continue
		}
		if filterProjectID != -1 && id != filterProjectID {
			continue
		}
		resp = append(resp, toSpecResponse(id, s))
	}
	c.JSON(http.StatusOK, resp)
}

// Get godoc
//
//	@Summary		Get a spec
//	@Tags			specs
//	@Produce		json
//	@Param			id	path		string	true	"Spec id"
//	@Success		200	{object}	specResponse
//	@Failure		404	{object}	errorResponse
//	@Router			/specs/{id} [get]
func (h *SpecHandler) Get(c *gin.Context) {
	projectID, capability, err := domain.DecodeSpecID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid spec id"})
		return
	}
	spec, err := h.specs.GetSpec(c.Request.Context(), projectID, capability)
	if err != nil {
		respondSpecError(c, err)
		return
	}
	resp := toSpecResponse(projectID, spec)

	if debts, err := h.context.ListDebts(c.Request.Context(), projectID); err == nil {
		for _, d := range debts {
			if d.Link.Kind == "spec" && d.Link.Capability == capability {
				resp.LinkedDebts = append(resp.LinkedDebts, toDebtResponse(projectID, d))
			}
		}
	}
	if reviews, err := h.context.ListReviews(c.Request.Context(), projectID); err == nil {
		for _, r := range reviews {
			if r.Link.Kind == "spec" && r.Link.Capability == capability {
				resp.LinkedReviews = append(resp.LinkedReviews, toReviewResponse(projectID, r))
			}
		}
	}

	c.JSON(http.StatusOK, resp)
}

type createSpecRequest struct {
	ProjectID    int64                `json:"projectId" binding:"required"`
	Capability   string               `json:"capability" binding:"required"`
	Purpose      string               `json:"purpose"`
	Requirements []specRequirementDTO `json:"requirements"`
}

// Create godoc
//
//	@Summary		Create a spec
//	@Tags			specs
//	@Accept			json
//	@Produce		json
//	@Param			body	body		createSpecRequest	true	"New spec"
//	@Success		201		{object}	specResponse
//	@Failure		400		{object}	errorResponse
//	@Failure		409		{object}	errorResponse	"spec already exists"
//	@Router			/specs [post]
func (h *SpecHandler) Create(c *gin.Context) {
	var req createSpecRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	spec, err := h.specs.CreateSpec(c.Request.Context(), req.ProjectID, req.Capability, req.Purpose, toDomainRequirements(req.Requirements))
	if err != nil {
		respondSpecError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toSpecResponse(req.ProjectID, spec))
}

type updateSpecRequest struct {
	Purpose      string               `json:"purpose"`
	Requirements []specRequirementDTO `json:"requirements"`
}

// Update godoc
//
//	@Summary		Update a spec
//	@Tags			specs
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string				true	"Spec id"
//	@Param			body	body		updateSpecRequest	true	"Updated purpose/requirements"
//	@Success		200		{object}	specResponse
//	@Failure		404		{object}	errorResponse
//	@Router			/specs/{id} [patch]
func (h *SpecHandler) Update(c *gin.Context) {
	projectID, capability, err := domain.DecodeSpecID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid spec id"})
		return
	}
	var req updateSpecRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	spec, err := h.specs.UpdateSpec(c.Request.Context(), projectID, capability, req.Purpose, toDomainRequirements(req.Requirements))
	if err != nil {
		respondSpecError(c, err)
		return
	}
	c.JSON(http.StatusOK, toSpecResponse(projectID, spec))
}

// Delete godoc
//
//	@Summary		Delete a spec
//	@Description	Deletes a spec. If other specs depend on it and force=true is not set, returns 409 with the list of dependent specs instead of deleting.
//	@Tags			specs
//	@Produce		json
//	@Param			id		path	string	true	"Spec id"
//	@Param			force	query	bool	false	"Delete despite dependents"
//	@Success		204
//	@Failure		404	{object}	errorResponse
//	@Failure		409	{object}	[]specResponse	"dependent specs"
//	@Router			/specs/{id} [delete]
func (h *SpecHandler) Delete(c *gin.Context) {
	projectID, capability, err := domain.DecodeSpecID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid spec id"})
		return
	}
	force := c.Query("force") == "true"

	dependents, err := h.specs.DeleteSpec(c.Request.Context(), projectID, capability, force)
	if err != nil {
		if errors.Is(err, service.ErrSpecHasDependents) {
			byPath, idErr := projectIDsByPath(c, h.projects)
			if idErr != nil {
				c.JSON(http.StatusInternalServerError, errorResponse{Error: idErr.Error()})
				return
			}
			resp := make([]specResponse, 0, len(dependents))
			for _, d := range dependents {
				id, ok := byPath[d.ProjectPath]
				if !ok {
					continue
				}
				resp = append(resp, toSpecResponse(id, d))
			}
			c.JSON(http.StatusConflict, resp)
			return
		}
		respondSpecError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

type dependencyRequest struct {
	TargetProjectID int64  `json:"targetProjectId" binding:"required"`
	TargetCapability string `json:"targetCapability" binding:"required"`
}

// AddDependency godoc
//
//	@Summary		Add a spec dependency
//	@Tags			specs
//	@Accept			json
//	@Param			id		path	string				true	"Spec id"
//	@Param			body	body	dependencyRequest	true	"Dependency target"
//	@Success		204
//	@Failure		400	{object}	errorResponse	"self-dependency rejected"
//	@Router			/specs/{id}/dependencies [post]
func (h *SpecHandler) AddDependency(c *gin.Context) {
	projectID, capability, err := domain.DecodeSpecID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid spec id"})
		return
	}
	var req dependencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	if err := h.specs.AddDependency(c.Request.Context(), projectID, capability, req.TargetProjectID, req.TargetCapability); err != nil {
		respondSpecError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RemoveDependency godoc
//
//	@Summary		Remove a spec dependency
//	@Tags			specs
//	@Accept			json
//	@Param			id		path	string				true	"Spec id"
//	@Param			body	body	dependencyRequest	true	"Dependency target"
//	@Success		204
//	@Router			/specs/{id}/dependencies [delete]
func (h *SpecHandler) RemoveDependency(c *gin.Context) {
	projectID, capability, err := domain.DecodeSpecID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid spec id"})
		return
	}
	var req dependencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	if err := h.specs.RemoveDependency(c.Request.Context(), projectID, capability, req.TargetProjectID, req.TargetCapability); err != nil {
		respondSpecError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

type generateSpecRequest struct {
	Description string `json:"description" binding:"required"`
	ProviderID  int64  `json:"providerId,omitempty"` // 0 = use the configured default provider
}

// Generate godoc
//
//	@Summary		Generate a draft spec
//	@Description	Drafts a spec from a short description without writing anything to disk. Uses the default AI provider, or an explicit providerId override, falling back to a deterministic template if none is configured or the provider call fails.
//	@Tags			specs
//	@Accept			json
//	@Produce		json
//	@Param			body	body		generateSpecRequest	true	"Description (and optional provider override)"
//	@Success		200		{object}	specResponse
//	@Router			/specs/generate [post]
func (h *SpecHandler) Generate(c *gin.Context) {
	var req generateSpecRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	draft := h.specs.GenerateDraft(c.Request.Context(), req.Description, req.ProviderID)
	c.JSON(http.StatusOK, toSpecResponse(0, draft))
}

func toDomainRequirements(dtos []specRequirementDTO) []domain.SpecRequirement {
	out := make([]domain.SpecRequirement, 0, len(dtos))
	for _, d := range dtos {
		out = append(out, domain.SpecRequirement{Name: d.Name, Body: d.Body})
	}
	return out
}

func respondSpecError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrSpecNotFound), errors.Is(err, service.ErrProjectNotFound):
		c.JSON(http.StatusNotFound, errorResponse{Error: err.Error()})
	case errors.Is(err, service.ErrSpecExists):
		c.JSON(http.StatusConflict, errorResponse{Error: err.Error()})
	case errors.Is(err, service.ErrSelfDependency):
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
	}
}

// projectIDsByPath resolves every registered project's registry id from its
// path. Used when a spec/change was discovered via filesystem walk (which
// only knows the path) but the response needs the opaque id.
func projectIDsByPath(c *gin.Context, projects *service.ProjectService) (map[string]int64, error) {
	all, err := projects.ListProjects(c.Request.Context())
	if err != nil {
		return nil, err
	}
	byPath := make(map[string]int64, len(all))
	for _, p := range all {
		byPath[p.Path] = p.ID
	}
	return byPath, nil
}
