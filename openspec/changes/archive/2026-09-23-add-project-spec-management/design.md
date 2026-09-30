## Context

Builds directly on `add-studio-foundation`'s backend model layer, project registry, and SSE contract. See proposal.md for motivation.

## Goals / Non-Goals

**Goals:**
- Full CRUD on real `openspec/` artifacts through the UI, kept byte-compatible with the OpenSpec CLI's own conventions so a project remains editable by either tool.
- Provenance and dependency metadata that survives round-trips through the OpenSpec CLI (i.e., doesn't corrupt files the CLI also touches).

**Non-Goals:**
- Replacing `openspec validate`/`openspec archive` semantics - the Studio writes conformant files, it does not reimplement the CLI's lifecycle commands.
- AI-driven generation quality - that's `add-ai-assistant-config`'s concern; this change only defines the generation UI/endpoint shape.

## Decisions

- **Provenance and dependency metadata storage**: stored as YAML frontmatter or a sibling `.meta.json` per spec, not inline in the requirement prose, so it doesn't pollute the human-readable spec content or break `openspec validate`'s parsing of `### Requirement:` / `#### Scenario:` blocks. Alternative considered: encoding metadata as an HTML comment inside `spec.md` - rejected as fragile to strip/preserve correctly on edit.
- **Cross-project dependency links** store the target project's registry id + spec path rather than an absolute path, so links survive a project being moved and re-registered at a new path (matched by id where possible, otherwise flagged broken per the spec's "broken dependency" scenario).
- **Kanban status computation reuses `openspec-ui`'s logic** (draft/todo/in-progress/done based on presence of proposal/tasks and checkbox state) rather than inventing a new status model, since it's already proven. **Done and Archived are kept as separate columns** (unlike an earlier draft of this change that merged them) - archived status is read from whether the change directory lives under `openspec/changes/archive/`, independent of task-completion state, matching `openspec-ui`'s own column split.
- **Column layout**: five equal-width flexible columns (`flex-1 min-w-0`, not fixed pixel widths), matching `openspec-ui`'s `KanbanBoard` component - avoids the too-wide/cramped columns the Studio's earlier placeholder grid had.

## Risks / Trade-offs

- Writing to files a user might also be editing by hand or via the OpenSpec CLI concurrently risks conflicting writes → mitigated by re-reading from disk before every write and surfacing a conflict rather than silently overwriting (exact conflict UX is an implementation detail, not a spec-level behavior change here).
- Metadata sidecar files add a new file type to `openspec/` that other OpenSpec tooling doesn't know about → kept additive and ignorable (CLI tools that don't read it simply don't see it, and it's not required for `openspec validate` to pass).
