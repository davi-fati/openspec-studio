## 1. Frontend: markdown rendering pipeline

- [x] 1.1 Integrate a Markdown renderer with GitHub-flavored extensions (tables, task lists, footnotes) shared by all artifact views and verify a sample doc with each feature renders correctly
- [x] 1.2 Integrate syntax highlighting for fenced code blocks and verify highlighting applies for at least Go, TypeScript, and JSON samples
- [x] 1.3 Integrate Mermaid rendering for ` ```mermaid ` fences with light/dark theme awareness and verify a flowchart renders and recolors on theme toggle
- [x] 1.4 Add graceful inline error handling for invalid mermaid syntax and verify the rest of the document still renders

## 2. Backend: project context model and API

- [x] 2.1 Define review-note and tech-debt-item models (id, description, status/link fields, timestamps, optional linked spec/change id) and verify unit tests cover creation and linking
- [x] 2.2 Implement storage under the project's `openspec/context/reviews/` and `openspec/context/debts/` directories and verify files are created there, not in Studio-only state
- [x] 2.3 Implement `GET/POST/PATCH/DELETE /api/projects/:id/reviews` and `.../debts` endpoints and verify each against a test project
- [x] 2.4 Implement linking a debt item or review to a spec/change id and verify the link is retrievable from both sides (spec detail shows linked debt, debt shows linked spec)

## 3. Frontend: project context UI

- [x] 3.1 Build Reviews list and creation form (with optional spec/change link) under the project's context section and verify created reviews appear immediately
- [x] 3.2 Build Tech Debt list, creation form, and status update control and verify status changes persist and reflect in the list
- [x] 3.3 Surface linked debt/review items on the relevant spec or change detail view and verify the link is visible from that side too
