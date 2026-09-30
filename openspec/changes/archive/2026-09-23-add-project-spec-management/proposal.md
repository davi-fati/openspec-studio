## Why

`openspec-ui` proved the Kanban + Specs read-only view. OpenSpec Studio's core differentiator is that specs stop being read-only: a user must be able to start a new project, author new specs, edit and delete them, and express dependencies between specs, all through the Studio rather than by hand-editing markdown. Every generated spec also needs provenance (who/what project/when) so multi-project, multi-author work stays auditable.

## What Changes

- Add the **Kanban** tab, matching `openspec-ui`'s change-lifecycle board layout (five separate, equal-width columns: Draft, Todo, In Progress, Done, Archived - Done and Archived kept distinct, not merged) but scoped to the active project (or all projects), styled per the Studio's Tailwind/shadcn visual language (see `project.md`; superseded the original `kanri` reference).
- Add the **Specs** tab: browse specs grouped by project/capability, with full CRUD - create a new spec/capability, edit its requirements, delete it - persisted back to the project's `openspec/specs/` files on disk.
- Add "new project" flow: scaffold a fresh OpenSpec-conformant project (via the foundation's project-registry init) directly from the Studio.
- Add spec generation: create a new spec (or change proposal) from a short description, producing a properly-formatted OpenSpec artifact the user then edits.
- Add spec provenance metadata: every spec/change created via the Studio SHALL record an id, created date/time, author, and owning project, and the Studio SHALL display this metadata wherever the spec is shown.
- Add spec dependency linking: a spec can declare that it depends on one or more other specs (same or different project), and the Studio SHALL visualize these links (e.g., in spec detail view and/or Kanban card).

## Capabilities

### New Capabilities
- `spec-authoring`: Create, edit, delete specs/capabilities and change proposals through the Studio UI, persisted to `openspec/` files.
- `spec-dependencies`: Declaring and visualizing dependency links between specs, including across projects.
- `kanban-board`: Project-scoped (or all-projects) Kanban view of change lifecycle, with Done and Archived as distinct columns, styled per the Studio's Tailwind/shadcn visual language.

## Impact

- Depends on `add-studio-foundation` (desktop shell, backend API/SSE contract, project registry).
- New backend model responsibilities: writing (not just reading) `openspec/` artifacts, and a small dependency-graph representation stored as spec metadata.
- New frontend surfaces: Kanban tab, Specs tab (list + editor), spec creation dialog, dependency picker/visualizer.
