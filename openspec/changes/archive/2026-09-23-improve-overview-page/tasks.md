## 1. Backend: broken-dependency count

- [x] 1.1 Add `BrokenSpecCount int` to `domain.ProjectOverview` (`backend/internal/domain/overview.go`) with a `json:"brokenSpecCount"` tag, and verify the struct still compiles and JSON-encodes the new field
- [x] 1.2 In `OverviewService.GetOverview`, build a `brokenByPath map[string]int` from `allSpecs` alongside the existing `specCountByPath`, counting a spec as broken when any of its `DependsOn`/`DependedBy` entries has `Available == false`, and set `po.BrokenSpecCount` from it, and verify a unit test with a spec that has a broken cross-project dependency reports the correct count while an unaffected project reports zero

## 2. Backend: richer recently-updated entries

- [x] 2.1 Add a `RecentProject{ProjectID int64; ProjectName string; LastOpenedAt time.Time}` type to `backend/internal/domain/overview.go` and change `Overview.RecentlyUpdated` from `[]string` to `[]RecentProject`, and verify the struct compiles
- [x] 2.2 Update `GetOverview`'s existing sort-by-`LastOpenedAt` loop to append a `RecentProject` (id, name, timestamp) per project instead of just the name, and verify a unit test confirms entries are still ordered most-recently-opened first and each carries the correct id/name/timestamp
- [x] 2.3 Update the `OverviewHandler` swagger doc comment to reflect the new `recentlyUpdated` shape, and regenerate swagger docs (`make swagger` or equivalent), and verify the generated docs reflect the new shape

## 3. Frontend: type and API updates

- [x] 3.1 Update `ProjectOverview`/`Overview`/add `RecentProject` in `frontend/src/lib/types.ts` to match the new backend shapes, and verify the frontend build's type-check step (`tsc -b`) passes
- [x] 3.2 Verify a manual fetch of `GET /api/overview` against the running backend returns `brokenSpecCount` per project and `recentlyUpdated` as an array of `{projectId, projectName, lastOpenedAt}` objects

## 4. Frontend: loading state

- [x] 4.1 In `Overview.tsx`, render a centered `Loader2` spinner in place of the summary/grid while `overview == null && !error`, and verify it appears on first render before data arrives and disappears once `overview` is set

## 5. Frontend: accessible, multi-destination cards

- [x] 5.1 Replace the per-project `<Card>`'s `<div onClick={() => goToProject(p.projectId)}>` wrapper with a non-interactive `<Card>` containing three real `<button>` destination controls (Kanban/Specs/Specflow), each calling `setActiveProjectId(p.projectId)` then `setActiveTab(<tab>)`, and verify each button is independently reachable via Tab key and activatable via Enter/Space
- [x] 5.2 Only render the three destination buttons when the project `available` is true, and verify an unavailable project's card shows no destination buttons
- [x] 5.3 Verify clicking each of the three buttons on a real project card navigates to the corresponding tab with that project selected in the switcher

## 6. Frontend: broken-dependency badge

- [x] 6.1 Render a warning badge on a project card showing `brokenSpecCount` when it is greater than zero, styled consistently with the existing "unavailable" badge, and verify it appears only when the count is nonzero and shows the correct count
- [x] 6.2 Verify a project with zero broken specs shows no badge

## 7. Frontend: recently-updated list

- [x] 7.1 Replace the truncated `recentlyUpdated.slice(0,3).join(', ')` summary card content with a real list of entries showing each project's name and a formatted last-opened timestamp, and verify it renders every entry (not just 3) or a scrollable/reasonable subset with clear indication if truncated
- [x] 7.2 Make each entry clickable, calling the same navigation used by a card's Kanban button, and verify clicking an entry navigates to that project's Kanban tab with it selected

## 8. Frontend: actionable empty state

- [x] 8.1 Add local `browserMode: 'open' | 'new' | null` state and a `DirectoryBrowserDialog` instance to `Overview.tsx`, mirroring `ProjectSwitcher.tsx`'s existing pattern, wired to `openProjectByPath`/`createProjectByPath` from `useProjects()`, and verify the dialog opens correctly from Overview
- [x] 8.2 Add Open and New buttons next to the existing "no projects registered yet" empty-state text, and verify they launch the directory browser in the corresponding mode and a successful confirm registers a project and the empty state disappears once it does

## 9. End-to-end verification

- [x] 9.1 Run the full frontend build (`npm run build`) and the backend build (`go build ./...`), and verify both complete with no new errors
- [x] 9.2 Run `go test ./internal/service/...` and verify the new/updated Overview tests and all existing tests pass
- [x] 9.3 Start the backend and frontend dev servers against real registered project data, and verify end-to-end: loading state appears briefly on first load, a project with a real broken dependency shows the warning badge with the correct count, each card's three destination buttons navigate correctly, the recently-updated list shows real timestamps and is clickable, and keyboard-only navigation (Tab + Enter) can activate a card's destination buttons
- [x] 9.4 Verify the empty-state Open/New buttons work end-to-end against a real directory (temporarily viewable by unregistering all projects, or by inspecting the empty-state render path directly)
- [x] 9.5 Run `openspec validate --all` and verify it passes

## 10. Completion-aware, color-coded progress indicator

- [x] 10.1 Change `OverviewService.GetOverview`'s `ProgressRatio` formula from `done/(todo+in_progress+done)` to `(done+archived)/(todo+in_progress+done+archived)`, and verify a unit test with a project that has Todo:1, In Progress:2, Done:0, Archived:6 reports a ratio of 6/9 (not 0)
- [x] 10.2 In `Overview.tsx`, compute `isComplete = changeCounts.todo === 0 && changeCounts.in_progress === 0` per project card and color the progress bar `bg-emerald-500` when complete, `bg-amber-500` otherwise - matching `ChangeCard`'s existing complete/incomplete color convention, and verify the color switches correctly for a project with pending Todo/In Progress work versus one with none
- [x] 10.3 Run the backend and frontend builds and the backend test suite, and verify both pass with no new errors
- [x] 10.4 Verify end-to-end against a real registered project with archived changes: the progress bar now shows nonzero width reflecting archived work, and is amber while any Todo/In Progress changes remain, turning emerald only once none remain
- [x] 10.5 Run `openspec validate --all` and verify it passes
