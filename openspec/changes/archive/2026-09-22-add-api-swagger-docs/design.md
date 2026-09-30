## Context

Backend is Go + Gin, layered `handler → service → repository → domain` (established in `add-studio-foundation`). Current handlers: `HealthzHandler`, `EventsHandler` (SSE), `ProjectHandler.List`/`Open` under `/api`. See proposal.md for motivation.

## Goals / Non-Goals

**Goals:**
- Generated docs stay a byproduct of annotated handler code, not a hand-maintained artifact that drifts.
- Zero behavior change to any existing `/api` endpoint.
- Low friction for future handlers to pick up the same convention.

**Non-Goals:**
- Authentication/authorization documentation - no auth exists in the backend yet.
- Client SDK generation from the spec.
- Any special-casing for how this behaves once the desktop (Tauri) shell lands in a later change - the Gin router is unaffected by which shell it runs under, so no extra design is needed here.

## Decisions

- **swaggo/swag + gin-swagger, not a hand-written OpenAPI YAML file**: keeps the spec next to the code it describes (as Go doc comments on each handler), so it's much harder for docs to silently drift from behavior. Alternative considered: hand-written `openapi.yaml` - rejected, becomes stale the moment a handler changes and nobody remembers to update it.
- **Generated output committed to `backend/docs/`, regenerated via a Makefile target**: `swag init` output is deterministic from annotations, so it's safe to commit (lets `gin-swagger` serve it without requiring every dev to run `swag init` before first boot) while still being regenerated (not hand-edited) whenever handlers change. `make build` and `make test` are extended to (re)run `make swagger` so a stale generated spec doesn't silently ship.
- **`/swagger/*any` mounted on the same Gin engine as `/api`**: simplest wiring, no separate server/port; matches how `add-studio-foundation` already mounts everything on one `*gin.Engine`.
- **SSE endpoint documented via a plain `@Description`/`@Router` annotation, no `@Success` response schema**: `swag`'s response-schema annotations assume a single JSON body, which doesn't model a stream of `event:`/`data:` frames. A minimal annotation still makes `/api/events` discoverable in the UI without forcing a misleading schema onto it.
- **No route-level gate for phase 1**: the Studio only runs locally against `127.0.0.1` in phase 1 (per `add-studio-foundation`), so `/swagger/*any` is just another loopback-only route, consistent with the existing security posture. If/when the app is distributed more broadly (desktop phase 2), revisit whether this route should be dev-only - not decided here since it doesn't change this change's scope or tasks.

## Risks / Trade-offs

- Annotations can still drift from actual handler behavior if a dev changes code but not the comment above it → mitigated by code review discipline; no automated drift-check is in scope here.
- Regenerating docs on every `make build`/`make test` adds a small amount of build time → acceptable, `swag init` over a handful of handlers is fast.
