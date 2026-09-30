# specflow-execution Specification

## Purpose
Triggers, tracks, and lets the user control the automated implementation of a scheduled Specflow flow via a configured AI provider.

## Requirements

### Requirement: Execution starts at the scheduled time
The system SHALL trigger a flow's execution automatically when its scheduled time window begins, invoking the selected AI provider on each change in flow order.

#### Scenario: Flow starts unattended
- **WHEN** a flow's scheduled start time is reached and the Studio is running
- **THEN** the Studio begins implementing the first flow item without requiring the user to be present

#### Scenario: Studio not running at scheduled time
- **WHEN** the Studio is not running when a flow's scheduled time arrives
- **THEN** the flow item remains pending and the Studio surfaces that it was missed the next time it runs, rather than silently skipping it

### Requirement: Per-item status tracking
The system SHALL track each flow item's status as pending, running, succeeded, or failed, and SHALL make implementation logs/output for that item viewable in the Studio.

#### Scenario: View logs of a running item
- **WHEN** the user opens a currently-running flow item
- **THEN** the Studio shows live or near-live output from the AI provider's implementation process

### Requirement: Failure halts only the affected flow
The system SHALL halt only the flow containing a failed item upon failure, surfacing the failure for review, and SHALL NOT halt unrelated flows for other projects.

#### Scenario: One project's flow fails, another's continues
- **WHEN** a flow item fails for project A while project B has a separate flow scheduled
- **THEN** project A's flow stops and is marked failed at that item, while project B's flow proceeds unaffected

### Requirement: User can pause or cancel a running flow
The system SHALL let the user pause or cancel a running flow, stopping further items from starting; an in-progress item's provider process SHALL be terminated on cancel.

#### Scenario: Cancel mid-run
- **WHEN** the user cancels a flow while an item is running
- **THEN** the Studio stops the in-progress implementation process and marks the flow cancelled, leaving already-succeeded items intact

### Requirement: Implementation progress visible outside Specflow
The system SHALL surface a change's implementation status (running/succeeded/failed) on the Kanban board and Overview dashboard while or after a Specflow run affects it.

#### Scenario: Kanban reflects in-flight implementation
- **WHEN** a change is currently being implemented by a Specflow run
- **THEN** its Kanban card shows an in-progress/running indicator without the user needing to open Specflow
