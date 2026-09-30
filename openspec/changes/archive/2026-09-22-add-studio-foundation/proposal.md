## Why

OpenSpec Studio has no code yet. Every other planned feature (Kanban + spec management, AI-assisted authoring, diagram rendering, overview, specflow) needs a running shell to live in: a Go backend layered as handler → service → repository (the conventional Gin community pattern), an SSE channel for push updates, and a React + TypeScript frontend that can talk to both. Building this foundation first avoids each feature change re-deciding process boundaries, API conventions, and project-loading behavior.

Phase 1 targets the web: the Go backend runs as a plain local server and the React frontend runs in a regular browser tab. Desktop packaging (Tauri 2) is deferred to a later change, once the core features (Kanban, spec authoring, dependencies, overview, diagrams, AI config, specflow) are closed out - the REST + SSE contract established here is designed to carry over unchanged when that happens.

## What Changes

- Scaffold the Go backend (`backend/`) using handler → service → repository layering: handlers for HTTP/SSE request-response shaping, services for business logic (including openspec artifact orchestration), repository for SQLite persistence via sqlc, domain for shared structs and filesystem parsing. Runs as a standalone local process (`go run` / built binary) bound to `127.0.0.1:<port>`.
- Establish the HTTP + SSE contract between backend and frontend: a base REST API under `/api` and a single SSE stream at `/api/events` for real-time updates (file watcher driven).
- Scaffold the React + TypeScript frontend (`frontend/`) with the top-level tab shell (Overview, Kanban, Specs, Specflow) as empty/placeholder views, theme (light/dark) support, and Tailwind CSS v4 + shadcn/ui-style primitives matching `openspec-ui`'s visual language (superseding the original `kanri`-inspired look - see `project.md`). Runs via a dev server, reachable in a browser.
- Implement "open a project": user points the Studio at a directory; backend validates/creates an `openspec/` structure there and registers it as an active project for the session.
- Implement multi-project registry: the Studio can track more than one opened project at once (name, path, last-opened), persisted locally, surfaced in a project switcher used by later features.

## Capabilities

### New Capabilities
- `backend-api-contract`: The Go backend's handler/service/repository structure, its base REST conventions, and the SSE event stream contract consumed by the frontend.
- `project-registry`: Opening, validating, and tracking multiple OpenSpec projects (directories) within a single Studio session.

### Deferred
- `desktop-shell` (Tauri 2 window lifecycle, supervising the Go backend as a subprocess) moves to a future change, proposed once the core feature set is complete. It reuses the same REST + SSE contract this change establishes.

## Impact

- New directories: `backend/`, `frontend/`.
- New local persistence: project registry (opened projects list) stored in a SQLite database file on disk, outside any single project's `openspec/` folder.
- No existing code affected (greenfield).
