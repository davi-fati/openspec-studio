## Purpose

Generates and serves interactive OpenAPI/Swagger documentation for the Go backend's REST API, kept in sync with the handlers via source annotations rather than hand-maintained separately.

## ADDED Requirements

### Requirement: OpenAPI spec generated from handler annotations
The system SHALL generate an OpenAPI spec from doc-comment annotations on Gin JSON handlers, and the generated spec SHALL describe every JSON request/response endpoint under `/api` (method, path, request body where applicable, response shape, status codes).

#### Scenario: Spec reflects a JSON endpoint
- **WHEN** the OpenAPI spec is generated
- **THEN** it includes an entry for `GET /api/healthz` and for `GET /api/projects` and `POST /api/projects`, each with their documented response shape and status codes

#### Scenario: New handler without annotations is caught
- **WHEN** a JSON handler under `/api` has no swag annotations
- **THEN** the generation step completes but the resulting spec omits that endpoint, making the gap visible in the generated docs rather than failing silently unnoticed

### Requirement: Streaming endpoint is documented without a response schema
The system SHALL list the `/api/events` SSE endpoint in the generated documentation with its path and a description identifying it as a Server-Sent Events stream, without requiring a JSON response schema for it.

#### Scenario: SSE endpoint appears in docs
- **WHEN** the OpenAPI spec is generated
- **THEN** `/api/events` appears in it with a description noting it is an SSE stream, and generation does not fail or require a response body schema for that entry

### Requirement: Interactive documentation UI is served
The system SHALL serve an interactive Swagger UI backed by the generated OpenAPI spec at a dedicated route, without altering any existing `/api` route's behavior.

#### Scenario: Swagger UI is reachable
- **WHEN** the backend is running and a browser requests the Swagger UI route
- **THEN** the interactive documentation page loads and lists the documented `/api` endpoints

#### Scenario: Existing API behavior is unchanged
- **WHEN** any existing `/api` endpoint is called the same way as before this change
- **THEN** its request handling, response, and status codes are identical to before
