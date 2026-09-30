package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Version is the Studio version reported by /api/healthz, set by main at
// startup. The desktop shell compares it to its own before attaching to an
// already-running backend.
var Version = "dev"

// HealthzHandler godoc
//
//	@Summary		Health check
//	@Description	Reports that the backend process is up, and its version.
//	@Tags			healthz
//	@Produce		json
//	@Success		200	{object}	healthzResponse
//	@Router			/healthz [get]
func HealthzHandler(c *gin.Context) {
	c.JSON(http.StatusOK, healthzResponse{Status: "ok", Version: Version})
}
