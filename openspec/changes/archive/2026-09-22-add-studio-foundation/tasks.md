## 1. Go backend scaffold

- [x] 1.1 Initialize Go module in `backend/` with `internal/handler`, `internal/service`, `internal/repository`, `internal/domain` packages, add Gin, the SQLite driver, and sqlc tooling (`sqlc.yaml`, `internal/repository/sql/`) as dependencies, and verify `go build ./...` and `sqlc generate` succeed
- [x] 1.2 Implement Gin HTTP server bound to `127.0.0.1:<port>` (port from flag/env, default fallback), routes grouped under `/api`, and verify `curl` against a `/api/healthz` endpoint returns 200
- [x] 1.3 Implement `/api/events` SSE endpoint (manual flush via `c.Writer`/`http.Flusher`) with a heartbeat and verify a manual SSE client receives periodic events
- [x] 1.4 Add filesystem watcher (fsnotify or equivalent) wired to registered project `openspec/` directories and verify an SSE `update` event fires on file change

## 2. Project registry (backend)

- [x] 2.1 Write the project registry SQL schema (path, name, last-opened) and queries under `internal/repository/sql/`, generate the sqlc query code, apply the schema to a SQLite database file in the OS app-data dir on startup, wrap it in a `repository` package function set, and verify round-trip read/write via unit test
- [x] 2.2 Implement `POST /api/projects` (open/register a project path) with validation of `openspec/` presence and verify it rejects nonexistent paths with a 4xx error
- [x] 2.3 Implement `openspec/` scaffold-on-init for a directory lacking OpenSpec structure and verify the resulting folder matches `openspec init` conventions (specs/, changes/, config.yaml)
- [x] 2.4 Implement `GET /api/projects` (list registered projects, flag unavailable paths) and verify a removed-path project is marked unavailable, not omitted

## 3. React + TypeScript frontend shell

- [x] 3.1 Scaffold `frontend/` with Vite + React + TypeScript and verify `npm run dev` serves a page in a browser
- [x] 3.2 Implement top-level tab shell (Overview, Kanban, Specs, Specflow) as placeholder routes and verify all four tabs render and switch without errors
- [x] 3.3 Implement light/dark theme toggle persisted client-side, styled per `openspec-ui`'s Tailwind/shadcn visual language (see `project.md`), and verify toggling persists across reload
- [x] 3.4 Implement project switcher UI backed by `GET /api/projects` and `POST /api/projects`, and verify opening a project updates the switcher list
- [x] 3.5 Implement SSE client hook that reconnects on drop and verify via dev tools that a forced connection drop reconnects within a few seconds

## 4. Integration verification

- [x] 4.1 End-to-end manual test: start the Go backend, open the frontend in a browser, open a real directory without `openspec/`, confirm scaffold + registration + switcher entry all work
- [x] 4.2 End-to-end manual test: edit a file inside an opened project's `openspec/` folder externally and confirm the frontend receives an SSE-triggered refresh
