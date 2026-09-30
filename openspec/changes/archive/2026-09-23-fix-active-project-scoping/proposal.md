## Why

`project-registry`'s "Active project selection scopes the UI" requirement already states the project switcher (next to the Open button) should scope Kanban, Specs, and Overview to the selected project. In practice, only Kanban and Specflow ever received `activeProjectId` - `listSpecs()` and `getOverview()` take no project filter at all, so picking a project in the switcher leaves Specs and Overview showing every registered project's data unfiltered. This closes that gap: selecting a project in the switcher SHALL filter every page, not just two of them.

## What Changes

- Add an optional `projectId` query filter to `GET /api/specs`, following the same pattern already used by `GET /api/changes` (filter the already-fetched list by the registered project's id).
- Add an optional `projectId` query filter to `GET /api/overview`: when present, the response scopes to that single project (its own card only, aggregate totals computed from just that project, "recently updated" limited to it).
- Wire `Specs.tsx` and `Overview.tsx` to pass the existing `activeProjectId` from `ProjectsContext` through to `listSpecs`/`getOverview`, matching how `Kanban.tsx`/`Specflow.tsx` already pass it to `listChanges`/`listFlows`.
- Update each page's descriptive subtext to reflect scope, same pattern Kanban already uses (`"Changes for the selected project"` vs `"Changes across all N registered project(s)"`).
- Complete `project-registry`'s existing "Active project selection scopes the UI" scenario, which today only asserts Kanban and Specs reload - add the missing Overview assertion so the spec matches the behavior actually being delivered.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `project-registry`: the "Active project selection scopes the UI" requirement's scenario is completed to explicitly cover Overview reloading to the selected project's data, not just Kanban and Specs.

## Impact

- Backend: `backend/internal/handler/specs.go` (`List`), `backend/internal/handler/overview.go` (`Get`), `backend/internal/service/overview.go` (`GetOverview`).
- Frontend: `frontend/src/lib/api.ts` (`listSpecs`, `getOverview` signatures), `frontend/src/pages/Specs.tsx`, `frontend/src/pages/Overview.tsx`.
- No new dependencies. No breaking change to existing callers: the `projectId` filter is optional on both endpoints, an omitted value preserves today's "all projects" behavior exactly.
