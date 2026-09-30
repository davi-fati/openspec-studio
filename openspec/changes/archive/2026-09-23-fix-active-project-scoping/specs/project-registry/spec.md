## MODIFIED Requirements

### Requirement: Active project selection scopes the UI
The system SHALL let the user switch which registered project is active, and SHALL scope the Kanban, Specs, and Overview views to that selection (or an "all projects" view where supported).

#### Scenario: Switching active project changes visible data
- **WHEN** the user selects a different project in the project switcher
- **THEN** the Kanban and Specs views reload to show only that project's changes and specs, and the Overview view reloads to show only that project's summary card and totals

#### Scenario: Selecting "all projects" restores the unscoped view
- **WHEN** the user selects "all projects" in the project switcher
- **THEN** the Kanban, Specs, and Overview views reload to show data across every registered project, as before a project was selected
