## Context

Purely aggregates data already produced by `add-project-spec-management` (Kanban status) and `add-studio-foundation` (project registry, SSE). See proposal.md for motivation.

## Goals / Non-Goals

**Goals:**
- A fast, read-only aggregation view; no new writable state.
- Reuse existing per-project status computation rather than duplicating Kanban-stage logic.

**Non-Goals:**
- New analytics/history (trends over time) - this is a current-state snapshot, not a reporting system.

## Decisions

- **Single `GET /api/overview` endpoint** that aggregates across all registered projects server-side, rather than the frontend fetching each project's data separately and combining client-side - keeps the "recently updated" and aggregate totals consistent and avoids N round trips.
- **Reuses the Kanban stage computation** from `add-project-spec-management`'s model layer as a shared function, not a re-derived one, so Overview and Kanban can never disagree about a change's stage.

## Risks / Trade-offs

- Aggregating many projects on every request could get slow as project count grows → acceptable at expected scale (a handful to dozens of projects for one engineer); revisit with caching if needed.
