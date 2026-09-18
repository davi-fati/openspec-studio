## Why

OpenSpec Studio has no code yet. Every other planned feature (Kanban + spec management, AI-assisted authoring, diagram rendering, overview, specflow) needs a running shell to live in: a Tauri 2 desktop app, a Go backend in MVC shape, an SSE channel for push updates, and a React + TypeScript frontend that can talk to both. Building this foundation first avoids each feature change re-deciding process boundaries, API conventions, and project-loading behavior.

## What Changes

- Scaffold a Tauri 2 application shell (`src-tauri/`) that launches and supervises a local Go backend process.
- Scaffold the Go backend (`backend/`) using MVC layering: models for openspec artifacts (project, spec, change), controllers for HTTP handlers, views for JSON/SSE payload shaping.
- Establish the HTTP + SSE contract between backend and frontend: a base REST API under `/api` and a single SSE stream at `/api/events` for real-time updates (file watcher driven).
- Scaffold the React + TypeScript frontend (`frontend/`) with the top-level tab shell (Overview, Kanban, Specs, Specflow) as empty/placeholder views, theme (light/dark) support, and the base layout look borrowed from `kanri`'s platform aesthetic.
- Implement "open a project": user points the Studio at a directory; backend validates/creates an `openspec/` structure there and registers it as an active project for the session.
- Implement multi-project registry: the Studio can track more than one opened project at once (name, path, last-opened), persisted locally, surfaced in a project switcher used by later features.

## Capabilities

### New Capabilities
- `desktop-shell`: Tauri 2 shell lifecycle - launching, supervising, and shutting down the embedded Go backend, and exposing the app window.
- `backend-api-contract`: The Go backend's MVC structure, its base REST conventions, and the SSE event stream contract consumed by the frontend.
- `project-registry`: Opening, validating, and tracking multiple OpenSpec projects (directories) within a single Studio session.

## Impact

- New directories: `src-tauri/`, `backend/`, `frontend/`.
- New local persistence: project registry (opened projects list) stored on disk outside any single project's `openspec/` folder.
- No existing code affected (greenfield).
