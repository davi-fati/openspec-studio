## 1. Backend: flow model and validation

- [x] 1.1 Define flow and flow-item models (project, ordered change references, schedule window, provider, status) and verify unit tests cover creation and reordering
- [x] 1.2 Implement "fully-planned" validation (proposal+specs+tasks present) before a change can be added to a flow and verify an incomplete change is rejected
- [x] 1.3 Implement dependency-ordering validation against `add-project-spec-management`'s spec-dependency data and verify a flow with an unmet dependency is rejected with a clear error

## 2. Backend: scheduler and execution engine

- [x] 2.1 Implement a persistent scheduler (goroutine/timer-based) that triggers flow execution at the configured start time and verify a flow scheduled a few seconds out actually starts
- [x] 2.2 Implement missed-schedule detection on backend startup and surface it via the flow's status/API rather than silently running or dropping it, verified by simulating a missed window
- [x] 2.3 Implement sequential per-flow execution invoking the configured AI provider (via `add-ai-assistant-config`'s invocation abstraction) per item and verify a two-item flow implements items strictly in order
- [x] 2.4 Implement per-item log capture and streaming and verify logs are retrievable both live (during run) and after completion
- [x] 2.5 Implement failure isolation: a failed item halts only its own flow and verify a concurrent second project's flow is unaffected by the first's failure
- [x] 2.6 Implement pause/cancel, including terminating an in-progress provider process, and verify a cancelled run leaves prior succeeded items intact and stops further items

## 3. Backend: API and SSE events

- [x] 3.1 Implement `GET/POST/PATCH/DELETE /api/specflow/flows` and reorder endpoint and verify against a test project
- [x] 3.2 Implement `POST /api/specflow/flows/:id/cancel` and verify it stops a running flow (found and fixed two real bugs in the process: (1) UTC vs local-time mismatch made SQLite's lexical timestamp comparison never consider a flow due; (2) cancelling only killed the direct CLI subprocess, leaving orphaned grandchildren (e.g. a script's own `sleep`) holding the log pipe open forever - fixed via process-group kill)
- [x] 3.3 Emit SSE events for flow/item status transitions and verify the frontend receives a running→succeeded/failed transition in real time

## 4. Frontend: Specflow tab

- [x] 4.1 Add Specflow tab after Specs in navigation with project filter (all/specific) and verify filter changes visible flows
- [x] 4.2 Build all-projects view with visual grouping/separation by project and verify two projects' flows are visually distinguishable
- [x] 4.3 Build flow builder: add fully-planned changes, reorder, set schedule window and provider, and verify an attempt to add an incomplete change is rejected in the UI
- [x] 4.4 Build flow item status display (pending/running/succeeded/failed) with log viewer and verify live logs render during a running item
- [x] 4.5 Build pause/cancel controls and verify cancelling a running flow updates its status immediately

## 5. Frontend: cross-feature status surfacing

- [x] 5.1 Surface running/succeeded/failed implementation status on Kanban cards for changes affected by an active Specflow run and verify it appears without opening Specflow
- [x] 5.2 Surface flow status in the Overview dashboard per project and verify it reflects the same status as Specflow and Kanban
