## 1. Backend: provider model and storage

- [ ] 1.1 Define provider model (type, config fields, credential reference) supporting both hosted-API and CLI-agent shapes and verify unit tests cover both
- [ ] 1.2 Implement OS keychain-backed credential storage with encrypted-file fallback and verify a stored key can be retrieved only by the owning provider
- [ ] 1.3 Implement CRUD endpoints `GET/POST/PATCH/DELETE /api/ai-providers` and verify each against a test config

## 2. Backend: health checks and invocation

- [ ] 2.1 Implement hosted-provider health check (minimal API call) and verify success/failure reporting for a reachable and an unreachable provider
- [ ] 2.2 Implement CLI-provider health check (invoke with version/no-op flag, resolve executable path) and verify failure is reported for a nonexistent path
- [ ] 2.3 Implement provider invocation abstraction (single interface for "generate from prompt") backing both hosted and CLI providers, verified by a unit test against a mocked hosted provider and a mocked CLI provider

## 3. Frontend: settings UI

- [ ] 3.1 Build AI provider settings section listing configured providers and verify it reflects backend state
- [ ] 3.2 Build "add hosted provider" form (type, API key, model) and verify a saved provider appears in the list
- [ ] 3.3 Build "add CLI provider" form (executable path, args) with path validation and verify an invalid path shows an inline error
- [ ] 3.4 Build health-check trigger with pass/fail indicator per provider and verify it reflects backend health-check results
- [ ] 3.5 Build default-provider selection and per-request provider override control, and verify a spec-generation request honors an explicit override

## 4. Integration

- [ ] 4.1 Wire `add-project-spec-management`'s spec-generation endpoint to call the selected provider via the invocation abstraction and verify generation succeeds end to end with at least one hosted and one CLI provider configured
