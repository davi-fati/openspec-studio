# specflow-scheduling Specification

## Purpose
Lets the user build an ordered, dependency-aware flow of fully-planned changes to be implemented automatically, scoped to one project or viewed across all projects.

## Requirements

### Requirement: Specflow tab and project filter
The system SHALL provide a Specflow tab, positioned after Specs in the top-level navigation, with a filter for "all projects" or a single selected project.

#### Scenario: Filter to single project
- **WHEN** the user selects a specific project in the Specflow filter
- **THEN** only that project's scheduled flow items are shown

#### Scenario: All-projects view separates flows visually
- **WHEN** the user selects "all projects" and more than one project has scheduled flow items
- **THEN** the Specflow view groups and visually separates items by project so flow boundaries between projects are clear

### Requirement: Only fully-planned changes can be scheduled
The system SHALL only allow a change to be added to a Specflow flow if it has a complete planning set (proposal, specs, tasks present per OpenSpec conventions).

#### Scenario: Attempt to schedule an incomplete change
- **WHEN** the user attempts to add a change that lacks `tasks.md` to a flow
- **THEN** the Studio rejects the addition and explains what is missing

### Requirement: Ordered flow with explicit sequence
The system SHALL let the user define an ordered sequence of changes within a flow and SHALL let the user reorder items before the flow runs.

#### Scenario: Reorder before execution
- **WHEN** the user drags a pending flow item to a new position before the flow has started
- **THEN** the new order is saved and used at execution time

### Requirement: Dependency ordering is enforced
The system SHALL prevent a flow from being scheduled to run a change before its declared spec dependencies are either earlier in the same flow or already implemented/applied.

#### Scenario: Dependency missing from flow and not yet applied
- **WHEN** the user attempts to schedule a change whose dependency is neither earlier in the flow nor already applied
- **THEN** the Studio blocks scheduling and indicates the unmet dependency

### Requirement: Schedule a time window and provider
The system SHALL let the user assign a flow a scheduled time window (e.g., a start time, optionally recurring, such as an overnight window) and select the AI provider (from configured providers) that will drive implementation.

#### Scenario: Schedule overnight run
- **WHEN** the user sets a flow's schedule to a specific start time and selects a configured provider
- **THEN** the flow is saved as pending and will begin execution at that time
