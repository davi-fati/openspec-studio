## Purpose

Provides consistent, feature-rich Markdown rendering (beyond plain paragraphs) across every surface in the Studio that displays OpenSpec artifact content.

## ADDED Requirements

### Requirement: Consistent rich Markdown rendering
The system SHALL render tables, syntax-highlighted fenced code blocks, task lists (`- [ ]` / `- [x]`), and footnotes consistently across every view that displays markdown content (spec detail, proposal, design, tasks, review notes, tech-debt notes).

#### Scenario: Table renders as a table
- **WHEN** a markdown document contains a GitHub-flavored table
- **THEN** the Studio renders it as an HTML table, not raw pipe-delimited text

#### Scenario: Code block gets syntax highlighting
- **WHEN** a fenced code block specifies a known language (e.g. `go`, `typescript`)
- **THEN** the Studio applies syntax highlighting for that language

#### Scenario: Task list checkboxes render visually
- **WHEN** a `tasks.md` file contains `- [ ]` and `- [x]` items
- **THEN** the Studio renders them as visually distinct unchecked/checked items, consistent with how the Kanban board interprets task completion
