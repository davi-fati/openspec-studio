## Context

Greenfield project. No frontend or backend code exists yet. See proposal.md for motivation. Phase 1 is web-only; the desktop (Tauri) shell is deferred to a future change - see "Future Phase: Desktop Shell" below.

## Goals / Non-Goals

**Goals:**
- A minimal but correctly-layered vertical slice: Go backend (handler → service → repository) → REST + SSE → React frontend in a browser → "open a project" working end to end.
- Establish conventions (API shape, package layout, SSE event naming) that later feature changes - and the future desktop shell - build on without renegotiating them.

**Non-Goals:**
- Building Kanban, spec editing, AI assistance, diagrams, overview metrics, or specflow scheduling - those are separate changes layered on this foundation.
- Desktop packaging (Tauri 2), signing, or distribution - deferred to a future change once core features are complete.
- Multi-user or remote/cloud sync of the project registry.

## Decisions

- **Go backend as a standalone local server (v1: web)**: the user specified Go for the backend. For phase 1 it runs as a plain local process (`go run` in dev, a built binary later) bound to `127.0.0.1:<port>`, with no supervising shell - the frontend dev server/browser talks to it directly over HTTP/SSE. No Tauri involvement in this phase.
- **SSE over WebSockets**: user specified SSE explicitly. SSE is simpler for the one-directional "something changed, re-fetch" pattern this app needs and avoids bidirectional connection management.
- **Package layout in Go: handler → service → repository**, the conventional layering in the Gin community, replacing the earlier MVC naming. `backend/internal/handler` (Gin handlers, `gin.Context`, request/response shaping - what a "Controller" and "View" would have been), `backend/internal/service` (business logic: opening a project, scaffolding `openspec/`, orchestrating the file watcher), `backend/internal/repository` (all persistence - the sqlc-generated SQLite access), `backend/internal/domain` (plain structs shared across layers: `Project`, `Spec`, `Change`, etc., plus the filesystem parsing for `openspec/` artifacts). A handler never touches the repository or SQL directly, and never parses the filesystem itself - it only calls a service.
- **HTTP framework: Gin**: user's choice. Handlers are Gin handler functions (`gin.Context`), grouped under an `/api` router group. SSE at `/api/events` is implemented via `c.Writer` with manual flush (`http.Flusher`), since Gin's own SSE helpers are limited - this stays an implementation detail inside the handler, not a contract change.
- **Project registry storage: SQLite (for now)**: the registry is persisted in a SQLite database file in a local app-data directory (outside any project's own `openspec/` folder), since it's Studio-level state, not project-level OpenSpec content. Chosen over a flat JSON file for simple querying/future growth (e.g. later Studio-level state like AI provider config, specflow schedules) without a schema migration to a real database later. Explicitly "for now": may be revisited if it proves to be overkill for a single-table registry.
- **Data access: sqlc (no ORM)**: user's choice, over GORM. Hand-written SQL (schema + queries) lives under `backend/internal/repository/sql/`, and `sqlc generate` produces type-safe Go query structs/functions consumed only by the `repository` package - services call the repository, never sqlc-generated code or raw SQL directly.

## Future Phase: Desktop Shell

Once the core feature set (Kanban, spec authoring, dependencies, overview, diagrams, AI config, specflow) is closed out, a separate change will add a Tauri 2 desktop shell that:
- Launches a window and supervises the same Go backend as a subprocess (sidecar or dev-run), rather than the browser talking to it directly.
- Adds a readiness handshake (backend signals "listening" before the window shows) and crash surfacing in the shell.
- Reuses the REST + SSE contract from this change unchanged - the frontend and backend should not need to know whether they're running under a browser tab or a Tauri webview.

This section exists so that decision doesn't need to be re-litigated later; it is not part of this change's scope or tasks.

## Risks / Trade-offs

- Running the Go HTTP server as a separate, unsupervised process in dev means it can be left running or forgotten between sessions → acceptable for local dev; revisit when the desktop shell adds process supervision.
- Loopback-only binding is still locally reachable by other processes on the same machine → acceptable for a local dev tool; revisit if the threat model changes.

## Open Questions

- None currently blocking phase 1 implementation.
