package handler

import (
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

// ServeUI answers every route not registered on r. Unknown /api and
// /swagger paths get a JSON 404 so API clients never receive HTML. Other
// GET/HEAD paths serve a file from ui when one exists, and otherwise the
// SPA entry page so client-side routes survive a reload. A nil ui (no
// frontend embedded) leaves only the JSON 404 behavior.
func ServeUI(r *gin.Engine, ui fs.FS) {
	var index []byte
	if ui != nil {
		index, _ = fs.ReadFile(ui, "index.html")
	}
	files := http.FS(ui)

	r.NoRoute(func(c *gin.Context) {
		p := c.Request.URL.Path
		if p == "/api" || strings.HasPrefix(p, "/api/") || strings.HasPrefix(p, "/swagger/") {
			c.JSON(http.StatusNotFound, errorResponse{Error: "not found"})
			return
		}
		if index == nil || (c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead) {
			c.JSON(http.StatusNotFound, errorResponse{Error: "not found"})
			return
		}

		name := strings.TrimPrefix(path.Clean(p), "/")
		if name != "" && name != "index.html" {
			if info, err := fs.Stat(ui, name); err == nil && !info.IsDir() {
				c.FileFromFS(name, files)
				return
			}
		}
		// Served from memory rather than FileFromFS: http.FileServer
		// redirects any request for index.html to "./".
		c.Header("Cache-Control", "no-cache")
		c.Data(http.StatusOK, "text/html; charset=utf-8", index)
	})
}
