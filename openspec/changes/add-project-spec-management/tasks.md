## 1. Backend: writable spec/change model

- [ ] 1.1 Extend model layer with write operations (create/update/delete) for spec and change artifacts and verify unit tests cover each operation
- [ ] 1.2 Implement provenance metadata storage (id, created timestamp, author, project) alongside each spec/change and verify it round-trips through save/reload
- [ ] 1.3 Implement dependency metadata storage on specs (list of referenced spec ids, possibly cross-project) and verify cross-project links resolve correctly
- [ ] 1.4 Add validation preventing self-referential dependencies and verify a rejection response is returned

## 2. Backend: API endpoints

- [ ] 2.1 Implement `POST /api/projects/new` (scaffold new project) and verify it creates a valid `openspec/` structure
- [ ] 2.2 Implement `POST /api/specs` (create), `PATCH /api/specs/:id` (edit), `DELETE /api/specs/:id` and verify each against a test project
- [ ] 2.3 Implement dependency-aware delete guard returning dependents list and verify deletion is blocked until confirmed
- [ ] 2.4 Implement `POST /api/specs/:id/dependencies` to add/remove links and verify persisted metadata matches request
- [ ] 2.5 Implement spec-generation endpoint that drafts a spec from a text description without writing to disk and verify the draft response is well-formed OpenSpec markdown

## 3. Frontend: Kanban tab

- [ ] 3.1 Build Kanban board component with lifecycle columns styled per `kanri` reference and verify it renders sample data
- [ ] 3.2 Wire project filter (single/all projects) and verify switching updates visible cards
- [ ] 3.3 Wire card click to open detail view (Proposal/Specs/Tasks/Design tabs) and verify each tab renders its artifact
- [ ] 3.4 Subscribe to SSE updates and verify a card moves column after an external file edit without manual refresh

## 4. Frontend: Specs tab and editor

- [ ] 4.1 Build Specs list grouped by project/capability and verify it lists specs from all registered projects
- [ ] 4.2 Build spec creation dialog including the generate-from-description flow and verify a generated draft is editable before save
- [ ] 4.3 Build spec editor (requirements/scenarios) with save persisting via `PATCH /api/specs/:id` and verify saved content matches the file on disk
- [ ] 4.4 Build delete flow with confirmation and dependents warning and verify a blocked delete shows dependent spec names
- [ ] 4.5 Display provenance metadata (id, created date/time, author, project) in spec detail view and verify it matches backend data

## 5. Frontend: dependency visualization

- [ ] 5.1 Build dependency picker for adding/removing links, including cross-project selection, and verify links persist after reload
- [ ] 5.2 Build dependency graph display in spec detail view (depends-on / depended-by) and verify both directions render correctly
- [ ] 5.3 Flag broken dependency links visually and verify a deleted dependency shows as broken rather than disappearing silently
