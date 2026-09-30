## Purpose

Defines the Go backend's handler/service/repository layering and the REST + SSE contract the frontend relies on for all data and real-time updates.

## ADDED Requirements

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
