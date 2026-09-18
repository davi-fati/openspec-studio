## Purpose

Defines the Go backend's MVC layering and the REST + SSE contract the frontend relies on for all data and real-time updates.

## ADDED Requirements

### Requirement: MVC layering
The Go backend SHALL be organized as Models (openspec artifact parsing/representation), Controllers (HTTP request handlers), and Views (response/JSON/SSE payload shaping), with no controller performing raw filesystem parsing directly.

#### Scenario: Controller delegates parsing to model
- **WHEN** a controller handles a request that needs spec or change data
- **THEN** it calls into a model package to obtain that data rather than reading and parsing files itself

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
