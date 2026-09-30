# OpenSpec Studio

A desktop development and project-management studio built around the OpenSpec framework. Where openspec-ui is a read-only monitoring dashboard, OpenSpec Studio is a full authoring environment: open real projects, create/edit/delete specs and changes, link dependencies between specs, drive AI-assisted spec generation, render diagrams, track project-level context (reviews, tech debt), see cross-project progress, and schedule automated implementation runs (specflow).

## Tech Stack
- **Phasing**: Phase 1 (web) closed out the core feature set. Phase 2 (in progress, `add-desktop-shell`) adds the Tauri desktop app while web mode stays a first-class target: one backend binary and one frontend build serve both.
- **Desktop shell (Phase 2)**: Tauri 2 in `src-tauri/`, a thin shell that supervises the Go backend as a sidecar (`--port 0`, waits for its `OPENSPEC_STUDIO_READY` line) or attaches to one already running, then points its webview at the UI the backend serves. It owns the app lifecycle: tray, staying in the background while Specflow flows are pending, graceful backend shutdown on quit. macOS only for now, unsigned.
- **Backend**: Go, layered as handler → service → repository (the Gin community's conventional pattern), HTTP via **Gin**. Runs as a local server bound to `127.0.0.1:<port>`, standalone (web) or as the desktop app's sidecar. A production build embeds the frontend and serves it itself. One backend per Studio data directory, enforced by an OS lock on `studio.lock` (see `backend-api-contract`).
- **Frontend**: React + TypeScript. Served by Vite in development; in builds it is embedded in and served by the backend, loaded either in a browser tab or in the Tauri webview (same origin, same relative `/api` calls). Desktop-only integrations (native folder picker) branch on `isDesktop()` - not a Nuxt/Vue app.
- **Frontend styling**: Tailwind CSS v4 (via `@tailwindcss/vite`) + shadcn/ui-style primitives (`class-variance-authority`, `clsx`/`tailwind-merge`, `lucide-react` icons), same approach as `openspec-ui`'s frontend - oklch color tokens, `@theme inline`, no separate `tailwind.config.js`. Flat, neutral surfaces (solid `card`/`background`, thin `border`, soft shadow); no glassmorphism/blur - tried and explicitly rejected for readability on a text-dense dev tool.
- **Real-time**: Server-Sent Events (SSE) from the Go backend to the frontend
- **Studio-level storage**: SQLite (for now) via **sqlc** (hand-written SQL, generated type-safe Go, no ORM) - holds Studio state that isn't OpenSpec content itself (project registry today; likely AI provider config and specflow schedules later)
- **Diagrams**: Mermaid rendering + enriched Markdown
- **Design reference**: visual language matches `openspec-ui` (same Tailwind/shadcn tokens, fonts: Plus Jakarta Sans + JetBrains Mono) rather than `kanri` - kanri was the original reference but its glassmorphism/Nuxt-flavored look didn't fit; feature parity baseline remains `openspec-ui` (Kanban + Specs views)

## Conventions
- Handler → service → repository on the Go backend: Handlers (Gin, HTTP/SSE request-response shaping), Services (business logic, openspec artifact orchestration), Repository (all SQLite persistence via sqlc), Domain (shared structs + filesystem parsing for openspec artifacts). No handler touches the repository or the filesystem directly.
- SSE is the only push mechanism; no WebSockets
- Filesystem (each project's `openspec/` folder) remains the source of truth for spec/change content; the app reads and writes real OpenSpec artifacts and does not mirror spec content into the SQLite database. SQLite only holds Studio-level state (project registry, and later config/schedules) that has no home in any single project's `openspec/` folder.
- Every spec/change created through the Studio carries provenance metadata: id, created date/time, author, and owning project
- Mobile-first is not a goal (desktop app), but layouts should be responsive to window resizing
- Light/dark theme support
- API documentation: every JSON handler under `/api` carries `swaggo/swag` doc-comment annotations (method, path, request/response shapes, status codes); a streaming endpoint like `/api/events` gets a discoverability-only annotation (no response schema). The OpenAPI spec is generated (`make swagger`), never hand-edited, and served as an interactive UI at `/swagger/*any`. New handlers follow this convention from the start.

## Project Structure
```
openspec-studio/
├── backend/           # Go backend (handler / service / repository / domain)
├── frontend/          # React + TypeScript app
├── openspec/          # This project's own specs
└── src-tauri/         # Tauri desktop shell (Rust); binaries/ holds the built sidecar
```
`backend/internal/webui/dist` is filled by `make build` with the frontend bundle the backend embeds.

## Navigation / Top-level Tabs
Order: **Overview → Kanban → Specs → Specflow**

## Running
Development: `make dev` (backend on `127.0.0.1:4173` + Vite on `localhost:5173` with an `/api` proxy).
Web build: `make build`, then run `backend/bin/openspec-studio-server` and open its address - one self-contained binary.
Desktop: `make desktop-dev` (runs the app against a fresh full build) or `make desktop-build` (`.app`/`.dmg` under `src-tauri/target/release/bundle/`). Needs the Rust toolchain; web targets do not.
