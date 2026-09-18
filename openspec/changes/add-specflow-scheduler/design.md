## Context

Builds on `add-project-spec-management` (change/spec/dependency data), `add-ai-assistant-config` (provider invocation abstraction), and `add-studio-foundation` (SSE, backend). See proposal.md for motivation.

## Goals / Non-Goals

**Goals:**
- Reliable unattended execution of a pre-approved sequence of implementation runs, resilient to the Studio not running exactly at the scheduled moment.
- Clear isolation between projects' flows so a failure in one never blocks or confuses another.
- Visibility: nothing about an automated run should be a black box - logs, status, and Kanban/Overview surfacing are first-class.

**Non-Goals:**
- Fully unattended conflict resolution or code review - a failed item stops for human review, it does not retry indefinitely or auto-merge.
- Distributed/remote execution - runs happen on the same machine as the Studio backend, driven by the locally configured provider (CLI or hosted API).
- Cross-machine schedule sync - schedules are local to the Studio installation that created them.

## Decisions

- **Scheduler lives in the Go backend as a persistent process/goroutine**, not in the Tauri/Rust layer or frontend, so scheduled runs are not tied to the window being focused (though they do require the app process to be running, per the "missed schedule" scenario in the execution spec). Alternative considered: OS-level cron - rejected because it can't share the backend's in-memory provider invocation abstraction and project state easily.
- **Missed-schedule handling is "surface and let user decide," not silent auto-run on next launch** - resuming an overnight run at 9am with no user awareness is a worse failure mode than a visible "missed" flag.
- **Provider invocation reuses `add-ai-assistant-config`'s abstraction** (hosted API vs CLI agent) rather than a separate implementation-specific integration, keeping one place that knows how to talk to each provider type.
- **Dependency enforcement checked at both schedule-time and execution-time**: schedule-time to give early feedback, execution-time because project state may have changed between scheduling and the run actually starting.

## Risks / Trade-offs

- Unattended agent runs modifying real project code is the highest-trust-boundary feature in the Studio → mitigated by: explicit opt-in scheduling (nothing runs without a user-built flow), visible logs, pause/cancel, and failure isolation per flow.
- Long-running overnight processes risk resource contention if many flows are scheduled concurrently → sequential execution within a flow is required by spec; cross-flow concurrency limits are an implementation detail to tune, not a spec-level behavior change.

## Open Questions

- Whether cross-flow (different projects) execution should run concurrently or be serialized Studio-wide by default. Does not change any spec behavior above (both are compatible with "failure halts only the affected flow"); can be decided as a configuration detail during implementation.
