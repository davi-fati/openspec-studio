package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/davidlima/openspec-studio/backend/internal/service"
)

type browseEntryDTO struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	HasOpenspec bool   `json:"hasOpenspec"`
	IsEmpty     bool   `json:"isEmpty"`
}

type browseResponse struct {
	Path        string           `json:"path"`
	Parent      string           `json:"parent,omitempty"`
	HasOpenspec bool             `json:"hasOpenspec"`
	IsEmpty     bool             `json:"isEmpty"`
	Entries     []browseEntryDTO `json:"entries"`
}

// Browse godoc
//
//	@Summary		Browse local directories
//	@Description	Lists a directory's real subdirectories on the machine the backend runs on, flagging each as an existing OpenSpec project and/or empty - powers the in-app Open/New directory picker. Read-only: never creates, modifies, or deletes anything. Defaults to the user's home directory when path is omitted.
//	@Tags			filesystem
//	@Produce		json
//	@Param			path	query		string	false	"Absolute directory path to list (defaults to home directory)"
//	@Success		200		{object}	browseResponse
//	@Failure		400		{object}	errorResponse	"path does not exist or is not a directory"
//	@Router			/fs/browse [get]
func Browse(projects *service.ProjectService) gin.HandlerFunc {
	return func(c *gin.Context) {
		result, err := projects.BrowseDirectory(c.Query("path"))
		if err != nil {
			if errors.Is(err, service.ErrBrowsePathNotFound) {
				c.JSON(http.StatusBadRequest, errorResponse{Error: "path does not exist or is not a directory"})
				return
			}
			c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
			return
		}

		entries := make([]browseEntryDTO, 0, len(result.Entries))
		for _, e := range result.Entries {
			entries = append(entries, browseEntryDTO{Name: e.Name, Path: e.Path, HasOpenspec: e.HasOpenspec, IsEmpty: e.IsEmpty})
		}
		c.JSON(http.StatusOK, browseResponse{
			Path:        result.Path,
			Parent:      result.Parent,
			HasOpenspec: result.HasOpenspec,
			IsEmpty:     result.IsEmpty,
			Entries:     entries,
		})
	}
}
