## Context

Rendering changes apply across every existing markdown surface introduced in `add-studio-foundation` and `add-project-spec-management`. Project context is new persisted data. See proposal.md for motivation.

## Goals / Non-Goals

**Goals:**
- One shared markdown-rendering component used everywhere, so diagram/table/code support is consistent rather than reimplemented per view.
- Reviews and tech debt stored as plain files under `openspec/context/`, keeping the filesystem-is-source-of-truth convention from `add-studio-foundation`.

**Non-Goals:**
- A full issue-tracker feature set (assignees, workflows, notifications) for tech debt - this is lightweight tracking co-located with specs, not a project-management replacement.
- Diagram authoring/editing UI - rendering only; diagrams are authored as Mermaid text in markdown.

## Decisions

- **Mermaid rendering client-side** in the React frontend (standard approach for desktop/web apps), not server-rendered by the Go backend, since it needs to react live to theme toggles without a round trip.
- **Review notes and debt items as individual markdown files with frontmatter** (id, status, links, timestamps) under `openspec/context/reviews/` and `openspec/context/debts/`, mirroring how specs/changes are already files-on-disk - keeps the "no separate database" convention from project.md.
- **Linking to a spec/change reuses the same id-reference mechanism** as `add-project-spec-management`'s spec-dependency links, rather than inventing a second linking format.

## Risks / Trade-offs

- Client-side Mermaid rendering can be slow for very large/complex diagrams → acceptable for typical architecture/flow diagrams in specs; not optimizing for pathological cases.
- Storing context files inside `openspec/` means non-Studio tools that walk that directory will see unfamiliar `context/` content → kept additive and namespaced so it doesn't collide with `specs/` or `changes/`.
