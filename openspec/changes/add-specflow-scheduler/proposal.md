## Why

Planning specs is only half the strategy shift the Studio aims for: once specs are fully planned, implementation should be automatable rather than manually triggered change by change. Users want to schedule a sequence of planned specs/changes to be implemented automatically - ideally during off-peak windows (e.g., overnight) to make efficient use of AI token/rate-limit windows - across one or more projects, without babysitting each agent run.

## What Changes

- Add the **Specflow** tab (after Specs in the top-level navigation) for scheduling implementation of planned specs/changes.
- Add a scheduling model: an ordered flow of changes (each already fully planned per OpenSpec, i.e. proposal+specs+tasks present) to be implemented in sequence, with a schedule (e.g., a time window such as an overnight period) and the AI provider (from `add-ai-assistant-config`) that will drive implementation.
- Add a project filter on the Specflow view: "all projects" or a specific project; when multiple projects have scheduled flows, the view SHALL visually separate flows by project so the user can tell which flow belongs to which project.
- Add execution: at the scheduled time, the Studio triggers implementation of each change in order via the configured AI provider, respecting each change's declared spec dependencies (a change SHALL NOT start before its dependencies are implemented, whether those dependencies are earlier in the same flow or already-applied changes).
- Add run status tracking: per scheduled flow item, track pending/running/succeeded/failed, with logs/output visible in the Studio, and surface this on Kanban and Overview once a run starts (a change moving through implementation should be visible without opening Specflow).
- Add safe-stop behavior: a running flow can be paused/cancelled by the user, and a failed item halts only that item's flow (not unrelated flows for other projects) with the failure surfaced for review.

## Capabilities

### New Capabilities
- `specflow-scheduling`: Building an ordered, dependency-aware flow of planned changes to implement, per project or across projects.
- `specflow-execution`: Triggering, tracking, and controlling scheduled implementation runs via a configured AI provider.

## Impact

- Depends on `add-studio-foundation` (backend/SSE), `add-project-spec-management` (change/spec/dependency data, Kanban status), and `add-ai-assistant-config` (provider that performs implementation).
- New backend responsibility: a scheduler/runner that can invoke a CLI agent or hosted-API-driven agent loop against real project code, not just generate text - this is a meaningfully larger trust boundary than spec generation and should default to conservative safeguards (explicit opt-in per flow, visible logs, easy cancel).
