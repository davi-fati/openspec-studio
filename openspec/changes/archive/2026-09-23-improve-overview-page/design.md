## Context

`Overview.tsx` currently renders three summary cards (open changes, in progress, a truncated `recentlyUpdated.slice(0,3).join(', ')`) then a grid of per-project `<Card>` elements as plain `<div onClick={() => goToProject(p.projectId)}>`, always jumping to Kanban. The backend's `domain.Overview`/`domain.ProjectOverview` (`backend/internal/domain/overview.go`) and `OverviewService.GetOverview` (`backend/internal/service/overview.go`, already accepting an optional `*int64` project filter from `fix-active-project-scoping`) compute everything from `ListProjects`/`ListSpecs`/`ListChanges` in one pass, with `RecentlyUpdated` built by sorting a copy of `projects` by `LastOpenedAt` and collecting just the names.

The Specs page already computes "broken dependency" per spec: `[...s.dependsOn, ...s.dependedBy].some(d => !d.available)` (`frontend/src/pages/Specs.tsx`), fed by `SpecService.resolveDependencies` (`backend/internal/service/spec.go:394-427`), which marks `DependsOn[i].Available = false` when the target project or capability no longer exists. `OverviewService.GetOverview` already calls `s.specs.ListSpecs(ctx)` to compute `specCountByPath` - the same result has each spec's resolved `DependsOn`, so a broken count needs no extra query.

`ProjectSwitcher.tsx` owns its own `DirectoryBrowserDialog` open/close state locally (`browserMode` useState) rather than through a shared context - the Overview page's new empty-state CTA needs the same dialog, so it needs its own local instance of the same pattern, not a new global toggle.

## Goals / Non-Goals

**Goals:**
- All six gaps closed with minimal, targeted diffs to existing files - no new pages, no new routing.
- Broken-dependency count computed once, in the same aggregation pass `GetOverview` already does, not a second backend round-trip.
- Match existing conventions exactly: `Kanban.tsx`'s `activeProjectId == null ? ... : ...` subtitle pattern, `ProjectSwitcher.tsx`'s local `DirectoryBrowserDialog` pattern, `Loader2`'s existing spinner usage.

**Non-Goals:**
- No new "delta since last visit" or trend indicators - out of scope, not one of the six identified gaps.
- No change to how the project-scoping filter (`?projectId=`, from `fix-active-project-scoping`) works - this change layers on top of it, doesn't touch it.
- No change to `SpecService`/dependency-resolution logic itself - only *reading* its existing `DependsOn[].Available` result.

## Decisions

**`ProjectOverview` gains `brokenSpecCount int`, computed in `GetOverview`'s existing spec-aggregation loop.** `GetOverview` already builds `specCountByPath` by iterating `allSpecs`; add a parallel `brokenByPath map[string]int` incremented when `sp.DependsOn` or `sp.DependedBy` contains an unavailable entry, using the exact same predicate the Specs page uses client-side. No new service call, no new endpoint - one more map built from data already in hand.

**`Overview.RecentlyUpdated` changes from `[]string` to `[]RecentProject{ProjectID int64; ProjectName string; LastOpenedAt time.Time}`.** This is the proposal's flagged **BREAKING** change: the existing sort-by-`LastOpenedAt` loop already has every field needed, it just currently throws away everything but the name. No behavior change to the sort itself, only to what's captured per entry.

**Drill-through becomes a small per-card menu, not a bigger navigation change.** `goToProject(projectId, tab: Tab)` replaces the single-purpose `goToProject(projectId)`, calling `setActiveProjectId(projectId)` then `setActiveTab(tab)` for whichever of `'Kanban' | 'Specs' | 'Specflow'` the user picked. Implemented as three small buttons inside the card (not a dropdown menu component - the project has no existing dropdown/menu primitive, and three inline buttons match the project's plain-Tailwind-button convention better than introducing one for this).

**Card accessibility: swap `<div onClick>` for a real `<button>` wrapping the card content**, or `role="button"` + `tabIndex={0}` + an `onKeyDown` Enter/Space handler if a `<button>` can't cleanly wrap the per-destination buttons (nested interactive elements are invalid HTML). Since the card now contains three separate destination buttons, the *card itself* is no longer a single click target - the outer card becomes a non-interactive `<Card>` wrapper, and each of the three destination buttons is independently focusable and keyboard-activatable by virtue of being real `<button>` elements. This also resolves the "keyboard-operable" spec requirement without a custom `role="button"` div at all, once the drill-through change lands.

**Loading state: a boolean derived from `overview === null` before the first successful fetch**, distinct from `error`. Render a centered `Loader2` (already imported for the "specflow running" badge) in place of the summary/grid while `overview == null && !error`.

**Empty-state Open/New reuses `ProjectSwitcher`'s exact pattern**: `Overview.tsx` gets its own `browserMode: 'open' | 'new' | null` state and a `<DirectoryBrowserDialog>` instance, wired to the same `openProjectByPath`/`createProjectByPath` from `useProjects()` that `ProjectSwitcher` already uses. Two small buttons next to the existing empty-state sentence, not a new shared component - `ProjectSwitcher` doesn't expose its dialog-trigger as reusable, and duplicating ~10 lines is simpler than refactoring dialog ownership into a context for one more caller.

**"Recently updated" list entries are clickable via the same `goToProject(projectId, 'Kanban')` used by cards** - clicking a recently-updated entry always goes to Kanban specifically (no per-entry Specs/Specflow choice), since the recently-updated list is a compact secondary list, not the primary per-project card.

**`ProgressRatio`'s formula changes from `done/(todo+in_progress+done)` to `(done+archived)/(todo+in_progress+done+archived)`**, computed in the same `OverviewService.GetOverview` loop that already builds `po.ChangeCounts`. Draft stays excluded from both numerator and denominator (pre-planning work isn't "not done yet", it's not committed yet). This directly fixes the reported bug: a project that's entirely archived (e.g. Todo:1, In Progress:2, Done:0, Archived:6) previously computed `0/(1+2+0) = 0%`, hiding that 6 of 9 non-draft changes are actually complete: it now computes `6/9 ≈ 67%`.

**Progress bar color follows `ChangeCard`'s existing convention** (`frontend/src/components/kanban/ChangeCard.tsx`: `isComplete ? 'bg-emerald-500' : 'bg-primary'`). `Overview.tsx` computes `isComplete = po.changeCounts.todo === 0 && po.changeCounts.in_progress === 0` per card (true even when the denominator is 0 - a project with only Draft or no changes at all has nothing "in progress" either, so it doesn't need an amber bar) and applies `emerald-500` when complete, `amber-500` otherwise - amber rather than `ChangeCard`'s plain `bg-primary` default, per the explicit ask for an orange/amber "in progress" color distinct from the default. This is computed client-side from `changeCounts`, already present on every `ProjectOverview` - no new backend field.

## Risks / Trade-offs

- **`brokenSpecCount` duplicates the Specs page's client-side broken-check logic on the backend** in a different layer (client array predicate vs. server-side aggregation). Accepted: the predicate is simple (`some dep unavailable`) and unlikely to drift, and keeping it server-side avoids Overview having to separately fetch and cross-reference full spec dependency graphs client-side just to render a count.
- **Breaking `Overview.recentlyUpdated`'s shape** has zero known external consumers (documented in the proposal's Impact section), but is still a real breaking change to a public REST response - acceptable given this is a pre-1.0 internal tool with one frontend consumer, updated in the same change.
- **Three destination buttons per card adds visual density** to an already information-dense card. Mitigated by keeping them small, icon-labeled, and only shown when the project is `available` (an unavailable project has nothing to navigate to).
