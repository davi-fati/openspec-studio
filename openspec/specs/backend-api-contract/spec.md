## Purpose

Defines the Go backend's handler/service/repository layering and the REST + SSE contract the frontend relies on for all data and real-time updates.

## Requirements

### Requirement: Handler/service/repository layering
The Go backend SHALL be organized as Handlers (Gin HTTP/SSE request-response shaping), Services (business logic and openspec artifact orchestration), and a Repository (all SQLite persistence via sqlc), with no handler performing raw filesystem parsing or persistence directly.

#### Scenario: Handler delegates to service
- **WHEN** a handler handles a request that needs spec, change, or project data
- **THEN** it calls into a service to obtain that data rather than reading/parsing files or querying the database itself

#### Scenario: Service delegates persistence to repository
- **WHEN** a service needs to read or write Studio-level state (e.g. the project registry)
- **THEN** it calls into the repository package rather than using sqlc-generated code or SQL directly

### Requirement: REST API namespace
The backend SHALL expose all HTTP endpoints under an `/api` prefix and return JSON.

#### Scenario: API request returns JSON
- **WHEN** the frontend calls any `/api/*` endpoint
- **THEN** the response is valid JSON with an appropriate HTTP status code

### Requirement: Single SSE event stream
The backend SHALL expose one Server-Sent Events endpoint at `/api/events` that pushes update notifications when watched project files change.

#### Scenario: File change triggers SSE event
- **WHEN** a file inside an opened project's `openspec/` directory changes on disk
- **THEN** the backend emits an SSE event on `/api/events` that the frontend uses to re-fetch affected data

#### Scenario: SSE reconnect
- **WHEN** the SSE connection drops
- **THEN** the frontend reconnects automatically without requiring a full app reload

### Requirement: Backend serves the frontend UI
The backend SHALL serve the built Studio frontend from its own HTTP origin, with `/api/*` and `/swagger/*` keeping their current behavior and any other unknown path returning the UI's entry page, so the Studio is usable from a browser with only the backend running.

#### Scenario: Standalone web mode
- **WHEN** the user starts only the built backend binary and opens its address in a browser
- **THEN** the Studio UI loads and all features work without a separate frontend dev server

#### Scenario: Deep link reload
- **WHEN** the user reloads the browser on a client-side route such as the Specflow tab
- **THEN** the backend returns the UI entry page and the UI restores that route, rather than a 404

#### Scenario: Unknown API path is not masked by the UI
- **WHEN** a client requests a non-existent `/api/*` path
- **THEN** the backend returns a JSON 404, not the UI entry page

### Requirement: Single backend per data directory
The backend SHALL hold an exclusive lock on its Studio data directory while running. A second backend started against the same data directory SHALL exit with a clear error that identifies the address of the backend already running, and SHALL NOT start its scheduler or modify the database.

#### Scenario: Second backend started
- **WHEN** a backend is running and the user starts another backend with the same data directory
- **THEN** the second backend exits non-zero with a message naming the running backend's address, and no Specflow flow is executed by it

#### Scenario: Stale lock after crash
- **WHEN** the previous backend crashed without releasing the lock and a new backend starts
- **THEN** the new backend detects the lock is no longer held by a live process and starts normally

### Requirement: Ephemeral port with readiness signal
The backend SHALL support being started on an OS-assigned free port and SHALL announce the bound address in a single machine-readable line on standard output once it is ready to serve requests.

#### Scenario: Started by a supervisor on port 0
- **WHEN** the backend is started with port `0`
- **THEN** it binds a free local port and prints one readiness line containing the bound address before or as it begins serving
