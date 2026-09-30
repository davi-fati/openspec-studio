## Why

The Overview page (`frontend/src/pages/Overview.tsx`) works but has six concrete gaps against what a landing dashboard should do: "recently updated" is a truncated, unclickable string; there's no loading state on first load; project cards use a non-focusable `<div onClick>` instead of a real interactive element; broken cross-project spec dependencies (already computed and shown on the Specs page) are invisible here; a card click always jumps to Kanban even when the user wants Specs or Specflow; and the empty state is a passive sentence instead of a direct call to action.

## What Changes

- **Recently updated becomes a real list**: each entry shows the project name and its last-opened timestamp, and is clickable (jumps to that project's Kanban, same as a card click) - not a truncated comma-joined string.
- **Loading state**: while the initial `GET /api/overview` request is in flight, show a loading indicator instead of a blank page.
- **Accessible project cards**: each card becomes a real focusable, keyboard-activatable control (`<button>`/`role="button"` with `tabIndex`), not a plain `<div onClick>`.
- **Broken-dependency indicator**: a project card shows a warning badge when any of its specs has a broken cross-project dependency (same "broken" computation the Specs page already uses), with a count.
- **Quick multi-page navigation per card**: instead of always jumping to Kanban, a card offers three destinations (Kanban / Specs / Specflow), each pre-filtered to that project - **BREAKING**: the existing single-click-goes-to-Kanban behavior from `overview-dashboard`'s "Drill-through navigation" requirement is replaced by this three-way choice.
- **Actionable empty state**: when no projects are registered, the empty state offers Open/New buttons directly (reusing the existing in-app directory browser), instead of only a sentence pointing at the header.
- **Progress indicator reflects real completion and is color-coded**: the ratio now counts Archived changes as completed work alongside Done (today only Done counts, so a project that's entirely archived shows a 0% bar even though nothing is left to do). The bar is shown in a distinct "complete" color when nothing remains in Todo/In Progress, and a distinct "in progress" color otherwise - matching the color convention `ChangeCard` already uses for its own task-completion bar (`emerald` when complete, the default color otherwise).

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `overview-dashboard`: "Aggregate multi-project summary" gains richer recently-updated entries; "Per-project progress card" gains a broken-dependency indicator and a completion-aware, color-coded progress ratio (counting Archived as completed); "Drill-through navigation" changes from a single Kanban-only destination to a three-way per-card choice (**BREAKING**); adds a loading-state requirement and an actionable-empty-state requirement.

## Impact

- Backend: `backend/internal/domain/overview.go` (`ProjectOverview`/`Overview` shape changes), `backend/internal/service/overview.go` (`GetOverview` computes broken-dependency counts and richer recently-updated entries).
- Frontend: `frontend/src/pages/Overview.tsx` (loading state, accessible cards, badge, multi-destination menu, empty-state CTA), `frontend/src/lib/types.ts` (`Overview`/`ProjectOverview` types).
- **BREAKING**: `Overview.recentlyUpdated` changes from `string[]` (project names) to a structured list with project id, name, and timestamp. `ProjectOverview` gains a `brokenSpecCount` field. Any external consumer of `GET /api/overview` depending on the old shape must adapt - none exist today besides this frontend.
