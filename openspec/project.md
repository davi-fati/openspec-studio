# OpenSpec Studio

A desktop development and project-management studio built around the OpenSpec framework. Where openspec-ui is a read-only monitoring dashboard, OpenSpec Studio is a full authoring environment: open real projects, create/edit/delete specs and changes, link dependencies between specs, drive AI-assisted spec generation, render diagrams, track project-level context (reviews, tech debt), see cross-project progress, and schedule automated implementation runs (specflow).

## Tech Stack
- **Desktop shell**: Tauri 2
- **Backend**: Go, MVC architecture, embedded in the Tauri app as a sidecar/local server
- **Frontend**: React + TypeScript
- **Real-time**: Server-Sent Events (SSE) from the Go backend to the frontend
- **Diagrams**: Mermaid rendering + enriched Markdown
- **Design reference**: visual/interaction language inspired by `kanri` (platform feel); feature parity baseline is `openspec-ui` (Kanban + Specs views)

## Conventions
- MVC on the Go backend: Models (spec/change/project parsing), Views (JSON/SSE payloads), Controllers (HTTP handlers)
- SSE is the only push mechanism; no WebSockets
- Filesystem (each project's `openspec/` folder) remains the source of truth; the app reads and writes real OpenSpec artifacts, it does not maintain a separate database of spec content
- Every spec/change created through the Studio carries provenance metadata: id, created date/time, author, and owning project
- Mobile-first is not a goal (desktop app), but layouts should be responsive to window resizing
- Light/dark theme support

## Project Structure
```
openspec-studio/
├── src-tauri/         # Tauri 2 shell + Go backend integration
├── backend/           # Go backend (MVC)
├── frontend/          # React + TypeScript app
└── openspec/          # This project's own specs
```

## Navigation / Top-level Tabs
Order: **Overview → Kanban → Specs → Specflow**

## Running
Tauri dev/build commands TBD once scaffolding lands (tracked as part of the `add-studio-foundation` change).
