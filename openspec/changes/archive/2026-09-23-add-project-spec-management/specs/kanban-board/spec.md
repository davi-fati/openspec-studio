## Purpose

Provides a Kanban view of change lifecycle across the active project (or all projects), matching `openspec-ui`'s behavior with the Studio's own visual language.

## ADDED Requirements

### Requirement: Kanban columns reflect change lifecycle
The system SHALL show changes as cards in five separate columns representing their lifecycle stage, in this order: Draft (proposal only), Todo (tasks exist, none complete), In Progress (some tasks complete), Done (all tasks complete, not yet archived), Archived (moved to `openspec/changes/archive/`). Done and Archived SHALL be distinct columns, not merged.

#### Scenario: Change with only a proposal is Draft
- **WHEN** a change has `proposal.md` but no `tasks.md`
- **THEN** it appears in the Draft column

#### Scenario: Change with all tasks checked is Done, not Archived
- **WHEN** a change's `tasks.md` has all checkboxes marked complete but the change has not been archived
- **THEN** it appears in the Done column, separate from the Archived column

#### Scenario: Archived change appears in its own column
- **WHEN** a change has been moved to `openspec/changes/archive/`
- **THEN** it appears in the Archived column, not in Done

### Requirement: Kanban scoped by project filter
The system SHALL filter the Kanban board by the active project selection, including an "all projects" option that visually groups cards by project.

#### Scenario: Single project selected
- **WHEN** the user has one project selected in the project switcher
- **THEN** the Kanban board shows only that project's changes

#### Scenario: All projects selected
- **WHEN** the user selects "all projects"
- **THEN** the Kanban board shows changes from every registered project, visually grouped or badged by project

### Requirement: Card click opens detail view
The system SHALL open a detail view showing proposal, specs, tasks, and design content when a Kanban card is clicked.

#### Scenario: Open change detail
- **WHEN** the user clicks a Kanban card
- **THEN** the Studio opens a detail view with tabs for Proposal, Specs, Tasks, and Design (when present)

### Requirement: Real-time board updates
The system SHALL update the Kanban board when underlying change files are modified, using the SSE event stream from the backend.

#### Scenario: External edit moves a card
- **WHEN** a `tasks.md` file is edited outside the Studio and a task is checked off
- **THEN** the corresponding card moves column without requiring a manual refresh
