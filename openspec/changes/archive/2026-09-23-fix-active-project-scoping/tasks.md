## 1. Backend: Specs endpoint filter

- [x] 1.1 Add an optional `?projectId=` query param to `SpecHandler.List`, parsed the same way `ChangeHandler.List` parses it, skipping specs whose resolved project id doesn't match when present, and verify via curl that `GET /api/specs?projectId=<id>` returns only that project's specs while `GET /api/specs` (no param) is unchanged
- [x] 1.2 Verify an invalid (non-numeric) `projectId` value returns 400, matching `ChangeHandler.List`'s existing behavior

## 2. Backend: Overview endpoint filter

- [x] 2.1 Change `OverviewService.GetOverview` to accept an optional `*int64` project filter, restricting the `projects` slice to that one project before the existing aggregation loop, and verify a unit test covers both the unfiltered (nil) case and a filtered case where totals/recentlyUpdated reflect only the selected project
- [x] 2.2 Wire `OverviewHandler.Get` to parse `?projectId=` and pass it through, returning 400 on an invalid value, and verify via curl that `GET /api/overview?projectId=<id>` returns a `projects` array with just that one project and totals computed only from it, while `GET /api/overview` (no param) is unchanged

## 3. Frontend: wire activeProjectId through

- [x] 3.1 Update `listSpecs` in `frontend/src/lib/api.ts` to accept `projectId: number | null` and encode it as `?projectId=` when non-null, matching `listChanges`'s existing signature, and verify the updated type-checks and a manual fetch against the running backend returns the filtered shape
- [x] 3.2 Update `getOverview` in `frontend/src/lib/api.ts` to accept `projectId: number | null` the same way, and verify likewise
- [x] 3.3 In `Specs.tsx`, read `activeProjectId` from `useProjects()`, pass it to `listSpecs`, and add it to `refresh`'s dependency array so switching projects re-fetches, and verify switching the project switcher reloads the Specs list to just that project's specs
- [x] 3.4 In `Overview.tsx`, read `activeProjectId` from `useProjects()`, pass it to `getOverview`, and add it to `refresh`'s dependency array, and verify switching the project switcher reloads Overview to just that project's card and totals

## 4. Frontend: subtext parity with Kanban

- [x] 4.1 Update `Specs.tsx`'s subtitle to distinguish "all projects" vs a selected project, matching `Kanban.tsx`'s existing `activeProjectId == null ? ... : ...` pattern, and verify both states render the expected text
- [x] 4.2 Update `Overview.tsx`'s subtitle the same way, and verify both states render the expected text

## 5. End-to-end verification

- [x] 5.1 Run the full frontend build (`npm run build`) and the backend build (`go build ./...`), and verify both complete with no new errors
- [x] 5.2 Run `go test ./internal/service/...` and verify the new Overview filter test and all existing tests pass
- [x] 5.3 Start the backend and frontend dev servers against real registered project data, and verify end-to-end: selecting a project in the switcher scopes Kanban (already worked), Specs, and Overview to that project simultaneously, and selecting "all projects" restores every page to the unscoped view
- [x] 5.4 Run `openspec validate --all` and verify it passes
