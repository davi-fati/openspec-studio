## Purpose

Gives the user a single landing page summarizing spec/change progress across all registered projects, as the first tab in the Studio.

## ADDED Requirements

### Requirement: Overview is the default landing tab
The system SHALL show the Overview tab first in the top-level navigation (Overview, Kanban, Specs, Specflow) and SHALL make it the tab shown on app launch.

#### Scenario: App launch shows Overview
- **WHEN** the user launches the Studio
- **THEN** the Overview tab is active by default

### Requirement: Per-project progress card
The system SHALL show one card per registered project summarizing: count of specs, count of changes by Kanban stage (Draft/Todo/In Progress/Done), and an overall applied-vs-planned progress indicator.

#### Scenario: Project with changes in multiple stages
- **WHEN** a project has changes in Draft, In Progress, and Done
- **THEN** its Overview card shows counts for each stage and a combined progress indicator

#### Scenario: Unavailable project is marked
- **WHEN** a registered project's path is currently inaccessible
- **THEN** its Overview card shows an unavailable state rather than stale or zeroed data

### Requirement: Aggregate multi-project summary
The system SHALL show an aggregate summary across all registered projects: total open changes, total in-progress changes, and most-recently-updated projects.

#### Scenario: Aggregate reflects all projects
- **WHEN** the user has three registered projects with open changes
- **THEN** the aggregate summary's total open-changes count equals the sum across all three

### Requirement: Drill-through navigation
The system SHALL let the user click a project's Overview card to navigate to that project's Kanban or Specs tab, pre-filtered to that project.

#### Scenario: Click card opens filtered Kanban
- **WHEN** the user clicks a project's Overview card
- **THEN** the Studio switches to the Kanban tab with that project already selected in the project filter

### Requirement: Real-time overview updates
The system SHALL update Overview data when underlying project files change, using the SSE event stream.

#### Scenario: Task completion updates progress indicator
- **WHEN** a task is checked off in a project's `tasks.md` outside the Studio
- **THEN** that project's Overview progress indicator updates without a manual refresh
