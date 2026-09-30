## 1. Backend: writable spec/change model

- [x] 1.1 Extend model layer with write operations (create/update/delete) for spec and change artifacts and verify unit tests cover each operation
- [x] 1.2 Implement provenance metadata storage (id, created timestamp, author, project) alongside each spec/change and verify it round-trips through save/reload
- [x] 1.3 Implement dependency metadata storage on specs (list of referenced spec ids, possibly cross-project) and verify cross-project links resolve correctly
- [x] 1.4 Add validation preventing self-referential dependencies and verify a rejection response is returned

## 2. Backend: API endpoints

- [x] 2.1 Implement `POST /api/projects/new` (scaffold new project) and verify it creates a valid `openspec/` structure
- [x] 2.2 Implement `POST /api/specs` (create), `PATCH /api/specs/:id` (edit), `DELETE /api/specs/:id` and verify each against a test project (also added `GET /api/specs` and `GET /api/specs/:id`, and `GET /api/changes`(`/:id`) - necessary read endpoints implied by tasks 3.x/4.x that weren't separately enumerated)
- [x] 2.3 Implement dependency-aware delete guard returning dependents list and verify deletion is blocked until confirmed
- [x] 2.4 Implement `POST /api/specs/:id/dependencies` to add/remove links and verify persisted metadata matches request
- [x] 2.5 Implement spec-generation endpoint that drafts a spec from a text description without writing to disk and verify the draft response is well-formed OpenSpec markdown

## 3. Frontend: Kanban tab

- [x] 3.1 Build Kanban board component with five separate lifecycle columns (Draft, Todo, In Progress, Done, Archived - equal-width, not fixed-width) styled per `openspec-ui`'s `KanbanBoard` layout, and verify it renders sample data
- [x] 3.2 Wire project filter (single/all projects) and verify switching updates visible cards (filter is the shared project-switcher selection, per the `kanban-board` spec's "active project selection" wording, not a separate control)
- [x] 3.3 Wire card click to open detail view (Proposal/Specs/Tasks/Design tabs) and verify each tab renders its artifact
- [x] 3.4 Subscribe to SSE updates and verify a card moves column after an external file edit without manual refresh (found and fixed a watcher bug in the process: it only watched top-level `specs/`/`changes/` dirs, not each change/spec's own subdirectory, so edits inside them were invisible - now watches recursively and picks up new subdirectories dynamically)

## 4. Frontend: Specs tab and editor

- [x] 4.1 Build Specs list grouped by project/capability and verify it lists specs from all registered projects
- [x] 4.2 Build spec creation dialog including the generate-from-description flow and verify a generated draft is editable before save
- [x] 4.3 Build spec editor (requirements/scenarios) with save persisting via `PATCH /api/specs/:id` and verify saved content matches the file on disk
- [x] 4.4 Build delete flow with confirmation and dependents warning and verify a blocked delete shows dependent spec names
- [x] 4.5 Display provenance metadata (id, created date/time, author, project) in spec detail view and verify it matches backend data

## 5. Frontend: dependency visualization

- [x] 5.1 Build dependency picker for adding/removing links, including cross-project selection, and verify links persist after reload
- [x] 5.2 Build dependency graph display in spec detail view (depends-on / depended-by) and verify both directions render correctly
- [x] 5.3 Flag broken dependency links visually and verify a deleted dependency shows as broken rather than disappearing silently
