package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/davidlima/openspec-studio/backend/internal/domain"
	"github.com/davidlima/openspec-studio/backend/internal/service"
)

// ProjectWatcher is the subset of *service.ProjectWatcher this handler needs,
// kept as an interface so it can be swapped/stubbed in tests.
type ProjectWatcher interface {
	WatchProject(path string) error
}

type ProjectHandler struct {
	projects *service.ProjectService
	watcher  ProjectWatcher
}

func NewProjectHandler(projects *service.ProjectService, watcher ProjectWatcher) *ProjectHandler {
	return &ProjectHandler{projects: projects, watcher: watcher}
}

type openProjectRequest struct {
	Path string `json:"path" binding:"required"`
}

type projectResponse struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Path         string `json:"path"`
	LastOpenedAt string `json:"lastOpenedAt"`
	Available    bool   `json:"available"`
}

func toResponse(p domain.Project) projectResponse {
	return projectResponse{
		ID:           p.ID,
		Name:         p.Name,
		Path:         p.Path,
		LastOpenedAt: p.LastOpenedAt.Format("2006-01-02T15:04:05Z07:00"),
		Available:    p.Available,
	}
}

// Open godoc
//
//	@Summary		Open (register) an existing project
//	@Description	Registers a directory as an active project. Read-only discovery: the directory must already contain a valid openspec/ structure - Open never scaffolds one. Use POST /projects/new for that.
//	@Tags			projects
//	@Accept			json
//	@Produce		json
//	@Param			body	body		openProjectRequest	true	"Directory to open"
//	@Success		200		{object}	projectResponse
//	@Failure		400		{object}	errorResponse	"path missing, does not exist, or has no valid openspec/ structure"
//	@Failure		500		{object}	errorResponse
//	@Router			/projects [post]
func (h *ProjectHandler) Open(c *gin.Context) {
	var req openProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "path is required"})
		return
	}

	project, err := h.projects.OpenProject(c.Request.Context(), req.Path)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPathNotFound):
			c.JSON(http.StatusBadRequest, errorResponse{Error: "path does not exist"})
		case errors.Is(err, service.ErrNoOpenSpecProject):
			c.JSON(http.StatusBadRequest, errorResponse{Error: "no valid OpenSpec project at this path - use \"New\" to scaffold one from scratch instead"})
		default:
			c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		}
		return
	}

	if h.watcher != nil {
		_ = h.watcher.WatchProject(project.Path)
	}

	c.JSON(http.StatusOK, toResponse(project))
}

// New godoc
//
//	@Summary		Create a new project
//	@Description	Scaffolds a fresh OpenSpec structure in an empty (or nonexistent-content) directory and registers it, distinct from opening an existing project.
//	@Tags			projects
//	@Accept			json
//	@Produce		json
//	@Param			body	body		openProjectRequest	true	"Directory to create the project in"
//	@Success		201		{object}	projectResponse
//	@Failure		400		{object}	errorResponse	"path missing, does not exist, or is not empty"
//	@Failure		500		{object}	errorResponse
//	@Router			/projects/new [post]
func (h *ProjectHandler) New(c *gin.Context) {
	var req openProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "path is required"})
		return
	}

	project, err := h.projects.CreateProject(c.Request.Context(), req.Path)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPathNotFound):
			c.JSON(http.StatusBadRequest, errorResponse{Error: "path does not exist"})
		case errors.Is(err, service.ErrDirNotEmpty):
			c.JSON(http.StatusBadRequest, errorResponse{Error: "directory is not empty"})
		default:
			c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		}
		return
	}

	if h.watcher != nil {
		_ = h.watcher.WatchProject(project.Path)
	}

	c.JSON(http.StatusCreated, toResponse(project))
}

// List godoc
//
//	@Summary		List registered projects
//	@Description	Lists every registered project, most recently opened first. A project whose path no longer exists on disk is included with available=false rather than omitted.
//	@Tags			projects
//	@Produce		json
//	@Success		200	{array}		projectResponse
//	@Failure		500	{object}	errorResponse
//	@Router			/projects [get]
func (h *ProjectHandler) List(c *gin.Context) {
	projects, err := h.projects.ListProjects(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}

	resp := make([]projectResponse, 0, len(projects))
	for _, p := range projects {
		resp = append(resp, toResponse(p))
	}
	c.JSON(http.StatusOK, resp)
}
