## Context

Greenfield project. No frontend, backend, or shell code exists yet. See proposal.md for motivation.

## Goals / Non-Goals

**Goals:**
- A minimal but correctly-layered vertical slice: Tauri window → Go backend (MVC) → REST + SSE → React shell → "open a project" working end to end.
- Establish conventions (API shape, MVC package layout, SSE event naming) that later feature changes build on without renegotiating them.

**Non-Goals:**
- Building Kanban, spec editing, AI assistance, diagrams, overview metrics, or specflow scheduling - those are separate changes layered on this foundation.
- Packaging/signing/distribution of the desktop app.
- Multi-user or remote/cloud sync of the project registry.

## Decisions

- **Go backend as a subprocess, not a Tauri Rust plugin**: the user specified Go for the backend with MVC architecture. Tauri 2 supports sidecar binaries; the Rust side of Tauri only supervises the process and can proxy webview requests to it (or the webview talks to `127.0.0.1:<port>` directly). Alternative considered: reimplement backend logic in Rust to stay inside Tauri's native model - rejected because the user explicitly wants Go.
- **SSE over WebSockets**: user specified SSE explicitly. SSE is simpler for the one-directional "something changed, re-fetch" pattern this app needs and avoids bidirectional connection management.
- **MVC package layout in Go**: `backend/internal/model`, `backend/internal/controller`, `backend/internal/view` (or equivalent), with `model` owning all `openspec/` filesystem parsing. Alternative considered: a flatter handler-per-file layout - rejected per explicit MVC requirement.
- **Project registry storage**: a small JSON file in the OS app-data directory (outside any project's own `openspec/` folder), since it's Studio-level state, not project-level OpenSpec content.

## Risks / Trade-offs

- Running a full Go HTTP server as a subprocess adds startup latency and a second binary to manage → mitigated by a readiness handshake (backend signals "listening" before Tauri shows the window) and clear crash surfacing.
- Loopback-only binding is still locally reachable by other processes on the same machine → acceptable for a local dev tool; revisit if the threat model changes.

## Open Questions

- Exact Tauri↔Go process communication: sidecar binary bundled by Tauri vs. developer-run Go process during `dev`. Resolve during task 1 implementation; does not change the spec or later features either way.
