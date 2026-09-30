## 1. Dependencies and generation wiring

- [x] 1.1 Add `github.com/swaggo/swag`, `github.com/swaggo/gin-swagger`, `github.com/swaggo/files` to `backend/go.mod` and verify `go build ./...` succeeds
- [x] 1.2 Add a `make swagger` target that runs `swag init` from `backend/`, outputting to `backend/docs/`, and verify running it produces `backend/docs/docs.go`, `swagger.json`, `swagger.yaml`
- [x] 1.3 Wire `make swagger` into `make build` and `make test` (run before the Go build/test steps) and verify `make build` regenerates `backend/docs/` from a clean checkout

## 2. Annotate existing handlers

- [x] 2.1 Add top-level API metadata annotations (title, version, base path `/api`) in `backend/cmd/server/main.go` and verify `swag init` picks them up in the generated `swagger.json`'s `info`/`basePath`
- [x] 2.2 Annotate `HealthzHandler` (`GET /api/healthz`) with method, path, and 200 response shape, and verify the generated spec includes this route with its response schema
- [x] 2.3 Annotate `ProjectHandler.List` (`GET /api/projects`) with method, path, and response shape (array of project objects), and verify the generated spec includes this route with its response schema
- [x] 2.4 Annotate `ProjectHandler.Open` (`POST /api/projects`) with method, path, request body shape, success response, and 4xx error response, and verify the generated spec includes this route with request/response schemas
- [x] 2.5 Add a discoverability-only annotation to `EventsHandler` (`GET /api/events`) noting it is an SSE stream, without a JSON response schema, and verify it appears in the generated spec without causing a `swag init` error

## 3. Serve the Swagger UI

- [x] 3.1 Wire `gin-swagger` + `swaggo/files` into `handler.New` to serve the UI at `/swagger/*any`, reading the generated `backend/docs` package, and verify `go build ./...` succeeds
- [x] 3.2 Start the backend and verify `GET /swagger/index.html` returns 200 and the page lists `/api/healthz`, `/api/projects` (GET and POST)
- [x] 3.3 Verify calling `/api/healthz`, `GET /api/projects`, and `POST /api/projects` directly (outside the Swagger UI) behaves identically to before this change (same status codes and response bodies)

## 4. Convention documentation

- [x] 4.1 Add a short "API documentation" convention note to `openspec/project.md` stating new handlers must carry swag annotations, and verify the note is present and consistent with the existing Conventions section style
