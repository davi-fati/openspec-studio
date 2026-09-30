## Context

`activeProjectId` already lives in `ProjectsContext` (`frontend/src/lib/ProjectsContext.tsx`) as shared, app-wide state set by the switcher next to the Open button (`frontend/src/components/ProjectSwitcher.tsx`). `Kanban.tsx` and `Specflow.tsx` already read it and pass it to `listChanges(activeProjectId)`/`listFlows(activeProjectId)` (`frontend/src/lib/api.ts`), which forward it as `?projectId=` to `GET /api/changes`/`GET /api/specflow/flows`. The `Changes` handler (`backend/internal/handler/changes.go:64-92`) establishes the pattern: fetch the full list, resolve each item's registered project id via `projectIDsByPath`, then optionally filter by a `?projectId=` query param, defaulting to no filter when the param is absent.

`Specs.tsx` and `Overview.tsx` never adopted this - `listSpecs()` (`frontend/src/lib/api.ts`) and `getOverview()` take no arguments, and their backend counterparts (`SpecHandler.List` in `backend/internal/handler/specs.go:88-107`, `OverviewHandler.Get` in `backend/internal/handler/overview.go:28-35` calling `OverviewService.GetOverview` in `backend/internal/service/overview.go:24-95`) always operate over every registered project. This is a gap against `project-registry`'s already-existing "Active project selection scopes the UI" requirement, not new product behavior.

## Goals / Non-Goals

**Goals:**
- Selecting a project in the switcher scopes Specs and Overview exactly the way it already scopes Kanban and Specflow.
- Follow the exact existing `?projectId=` filter convention (`changes.go`), not a new pattern.
- Zero behavior change when no project is selected ("all projects" / `activeProjectId == null`).

**Non-Goals:**
- No changes to Kanban or Specflow - already correct.
- No changes to how the project switcher itself works (covered by the already-archived `refine-project-registration` change).
- No persistence of the selected project across sessions - `activeProjectId` is already in-memory only, unchanged here.

## Decisions

**Specs: filter in the handler, same as Changes.** `SpecHandler.List` already computes `byPath` (project path -> registered id) to build each `specResponse`. Add the same `?projectId=` param read as `ChangeHandler.List` does, and skip a spec whose resolved id doesn't match. `SpecService.ListSpecs` itself is untouched - filtering by registered-project-id, not by filesystem path, belongs in the handler where the id resolution already happens, exactly like Changes.

**Overview: filter in the service, not just the handler**, because `GetOverview` aggregates *across* projects (totals, recently-updated ordering) rather than returning a flat per-item list - a handler-only filter would still compute wrong totals from all projects and then throw away rows. `OverviewService.GetOverview(ctx, projectID *int64)` takes an optional filter: when non-nil, it restricts the `projects` slice to that one project before doing the existing aggregation loop, so `TotalOpenChanges`/`TotalInProgress`/`RecentlyUpdated` naturally end up scoped to just that project with no separate scoped-math branch. `OverviewHandler.Get` parses `?projectId=` (mirroring `ChangeHandler.List`'s `strconv.ParseInt` handling) and passes it through.

**Frontend: pass `activeProjectId` straight through, matching Kanban/Specflow exactly.** `listSpecs(projectId: number | null)` and `getOverview(projectId: number | null)` gain the same signature `listChanges`/`listFlows` already have, encoding `projectId` as `?projectId=` when non-null (reuse the existing helper those two already use in `api.ts`, if any, otherwise mirror their inline logic). `Specs.tsx`'s `refresh` and `Overview.tsx`'s `refresh` add `activeProjectId` to their `useCallback` dependency array so switching projects re-fetches, same as `Kanban.tsx`'s `refresh` does today.

**Overview's per-project card grid still renders through the same JSX** - when scoped, `overview.projects` from the backend now naturally contains one entry, so no frontend conditional rendering logic is needed beyond passing the id through and adjusting the header subtext (matching `Kanban.tsx`'s existing `activeProjectId == null ? ... : ...` ternary for its subtitle).

## Risks / Trade-offs

- **`RecentlyUpdated` becomes a single-item list when scoped** (just the active project's own name) - a bit redundant next to its own summary card, but consistent and not worth a special case to hide it.
- **Overview's card `onClick` (`goToProject`) still sets `activeProjectId` and jumps to Kanban** - already idempotent when the clicked card is the only one shown (scoped view), so no change needed there.
