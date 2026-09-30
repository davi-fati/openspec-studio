package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/gin-gonic/gin"
)

func newUIRouter(ui fstest.MapFS) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/healthz", HealthzHandler)
	if ui == nil {
		ServeUI(r, nil)
	} else {
		ServeUI(r, ui)
	}
	return r
}

func get(r *gin.Engine, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	return w
}

var testUI = fstest.MapFS{
	"index.html":         {Data: []byte("<!doctype html><div id=root></div>")},
	"assets/app-abc.js":  {Data: []byte("console.log(1)")},
	"assets/app-abc.css": {Data: []byte("body{}")},
}

func TestServeUIClientRouteReturnsIndex(t *testing.T) {
	r := newUIRouter(testUI)
	for _, path := range []string{"/", "/specflow", "/kanban/some-change"} {
		w := get(r, path)
		if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `id=root`) {
			t.Fatalf("GET %s: code=%d body=%q, want index page", path, w.Code, w.Body.String())
		}
	}
}

func TestServeUIStaticAsset(t *testing.T) {
	w := get(newUIRouter(testUI), "/assets/app-abc.js")
	if w.Code != http.StatusOK || w.Body.String() != "console.log(1)" {
		t.Fatalf("asset: code=%d body=%q", w.Code, w.Body.String())
	}
}

func TestServeUIUnknownAPIPathIsJSON404(t *testing.T) {
	r := newUIRouter(testUI)
	for _, path := range []string{"/api/nope", "/api", "/swagger/nope"} {
		w := get(r, path)
		if w.Code != http.StatusNotFound {
			t.Fatalf("GET %s: code=%d, want 404", path, w.Code)
		}
		var body errorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil || body.Error == "" {
			t.Fatalf("GET %s: body %q is not a JSON error", path, w.Body.String())
		}
	}
}

func TestServeUIKnownAPIRouteUnaffected(t *testing.T) {
	w := get(newUIRouter(testUI), "/api/healthz")
	var body healthzResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil || body.Status != "ok" || body.Version == "" {
		t.Fatalf("healthz: code=%d body=%q", w.Code, w.Body.String())
	}
}

func TestServeUIWithoutEmbeddedUI(t *testing.T) {
	w := get(newUIRouter(nil), "/specflow")
	if w.Code != http.StatusNotFound {
		t.Fatalf("no UI: code=%d, want 404", w.Code)
	}
}
