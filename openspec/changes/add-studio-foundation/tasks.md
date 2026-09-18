## 1. Go backend scaffold

- [ ] 1.1 Initialize Go module in `backend/` with `internal/model`, `internal/controller`, `internal/view` packages and verify `go build ./...` succeeds
- [ ] 1.2 Implement HTTP server bound to `127.0.0.1:<port>` (port from flag/env, default fallback) and verify `curl` against a `/api/healthz` endpoint returns 200
- [ ] 1.3 Implement `/api/events` SSE endpoint with a heartbeat and verify a manual SSE client receives periodic events
- [ ] 1.4 Add filesystem watcher (fsnotify or equivalent) wired to registered project `openspec/` directories and verify an SSE `update` event fires on file change

## 2. Project registry (backend)

- [ ] 2.1 Implement project registry model (path, name, last-opened) persisted as JSON in the OS app-data dir and verify round-trip read/write via unit test
- [ ] 2.2 Implement `POST /api/projects` (open/register a project path) with validation of `openspec/` presence and verify it rejects nonexistent paths with a 4xx error
- [ ] 2.3 Implement `openspec/` scaffold-on-init for a directory lacking OpenSpec structure and verify the resulting folder matches `openspec init` conventions (specs/, changes/, config.yaml)
- [ ] 2.4 Implement `GET /api/projects` (list registered projects, flag unavailable paths) and verify a removed-path project is marked unavailable, not omitted

## 3. Tauri 2 shell

- [ ] 3.1 Scaffold `src-tauri/` Tauri 2 app and verify `tauri dev` opens an empty window
- [ ] 3.2 Wire Go backend subprocess launch on app start with a readiness handshake before showing the window, verified by manual run showing no blank-window flash
- [ ] 3.3 Wire backend process termination on window close and verify no orphaned `backend` process remains after quitting (process list check)
- [ ] 3.4 Surface a backend-crash error state in the shell and verify it appears when the backend binary is killed externally during a running session

## 4. React + TypeScript frontend shell

- [ ] 4.1 Scaffold `frontend/` with Vite + React + TypeScript and verify `npm run dev` serves a page
- [ ] 4.2 Implement top-level tab shell (Overview, Kanban, Specs, Specflow) as placeholder routes and verify all four tabs render and switch without errors
- [ ] 4.3 Implement light/dark theme toggle persisted client-side, styled per `kanri` reference, and verify toggling persists across reload
- [ ] 4.4 Implement project switcher UI backed by `GET /api/projects` and `POST /api/projects`, and verify opening a project updates the switcher list
- [ ] 4.5 Implement SSE client hook that reconnects on drop and verify via dev tools that a forced connection drop reconnects within a few seconds

## 5. Integration verification

- [ ] 5.1 End-to-end manual test: launch app, open a real directory without `openspec/`, confirm scaffold + registration + switcher entry all work
- [ ] 5.2 End-to-end manual test: edit a file inside an opened project's `openspec/` folder externally and confirm the frontend receives an SSE-triggered refresh
