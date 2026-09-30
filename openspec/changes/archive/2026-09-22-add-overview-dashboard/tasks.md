## 1. Backend: aggregation endpoint

- [x] 1.1 Implement `GET /api/overview` aggregating spec counts, per-stage change counts, and progress ratio per registered project and verify output against a multi-project test fixture
- [x] 1.2 Implement aggregate totals (open, in-progress, recently updated) across all projects and verify sums match per-project data
- [x] 1.3 Mark unavailable projects distinctly in the aggregation response and verify a project with a missing path is flagged, not omitted or zeroed

## 2. Frontend: Overview tab

- [x] 2.1 Add Overview as the first tab and default active tab on launch, verified by a fresh-launch manual check
- [x] 2.2 Build per-project progress card component and verify it renders stage counts and a progress indicator from `GET /api/overview`
- [x] 2.3 Build aggregate summary panel and verify totals match the sum of visible project cards
- [x] 2.4 Wire drill-through click to navigate to Kanban/Specs pre-filtered to the clicked project and verify the filter is applied on arrival
- [x] 2.5 Subscribe Overview to SSE updates and verify a progress indicator updates after an external task-completion edit
