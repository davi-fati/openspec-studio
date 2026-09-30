// Package webui embeds the built frontend so the backend binary alone can
// serve the Studio. `make build` copies frontend/dist into ./dist before
// compiling; a plain `go run`/`go build` embeds only the placeholder and the
// backend runs API-only (the `make dev` setup, where Vite serves the UI).
package webui

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var embedded embed.FS

// FS returns the embedded UI rooted at its dist directory, and false when
// no UI was built into this binary.
func FS() (fs.FS, bool) {
	sub, err := fs.Sub(embedded, "dist")
	if err != nil {
		return nil, false
	}
	if _, err := fs.Stat(sub, "index.html"); err != nil {
		return nil, false
	}
	return sub, true
}
