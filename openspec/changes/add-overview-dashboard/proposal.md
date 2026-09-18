## Why

An engineer working across several projects at once needs a single place to see where each project stands - spec progress, what's in flight, what's stuck - without clicking into each project's Kanban board individually. This is the day-to-day landing page of the Studio.

## What Changes

- Add the **Overview** tab as the first tab in the top-level navigation (Overview → Kanban → Specs → Specflow).
- Show, per registered project, spec/change progress: counts of specs by status, changes by Kanban stage, and an overall "applied vs planned" progress indicator per project.
- Show an aggregate multi-project summary (e.g., total open changes, total in-progress, recently updated projects) for users tracking many projects at once.
- Support drilling from an Overview project card into that project's Kanban or Specs tab, pre-filtered to that project.

## Capabilities

### New Capabilities
- `overview-dashboard`: Cross-project progress summary landing page, with per-project cards and drill-through navigation.

## Impact

- Depends on `add-studio-foundation` (project registry, SSE) and `add-project-spec-management` (Kanban stage/status data it summarizes).
- Read-only feature: it aggregates existing data, no new writable artifacts.
