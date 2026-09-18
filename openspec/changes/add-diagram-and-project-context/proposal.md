## Why

Specs and design docs often reference diagrams (architecture, flows) and richer markdown than plain text; rendering these as plain text loses information the author intended. Separately, the day-to-day of managing several projects needs more than specs: reviews and technical debt need a home too, and that home should live alongside a project's specs in the Studio rather than in a separate tool.

## What Changes

- Add Mermaid diagram rendering anywhere Studio renders markdown content (spec detail, design.md, proposal.md).
- Add enriched Markdown rendering support (tables, code blocks with syntax highlighting, task lists, footnotes) consistently across all markdown views in the Studio.
- Add a **Project Context** section (within a project's view, alongside its OpenSpec sections) that covers, at minimum: **Reviews** (notes/findings tied to a project, optionally to a specific spec/change) and **Tech Debt** (tracked items describing known debt, status, and optional links to the spec/change that would resolve them).

## Capabilities

### New Capabilities
- `diagram-rendering`: Mermaid diagram rendering embedded in markdown views across the Studio.
- `enriched-markdown`: Consistent rich Markdown rendering (tables, syntax-highlighted code, task lists, footnotes) across all markdown surfaces.
- `project-context`: Per-project Reviews and Tech Debt tracking, stored alongside the project's OpenSpec artifacts.

## Impact

- Depends on `add-studio-foundation` (markdown views already exist for proposal/design/spec content) and benefits from `add-project-spec-management` (linking debt/review items to specs uses the same dependency-link mechanism).
- New persisted artifact types: review notes and tech-debt items, stored under the project's `openspec/` tree (e.g., `openspec/context/reviews/`, `openspec/context/debts/`) so they travel with the project like specs do.
