## Purpose

Lets a user open one or more real project directories in the Studio and keeps track of which ones are active across sessions.

## Requirements

### Requirement: Open a project directory
The system SHALL allow the user to browse and select a directory to open as a project, using the in-app filesystem browser. The system SHALL only register a directory as a project if it already contains a valid `openspec/` structure - it SHALL NOT scaffold one on the user's behalf via Open. If the selected directory has no valid `openspec/` structure, the system SHALL reject the selection with a clear message directing the user to use "New" instead.

#### Scenario: Open existing OpenSpec project
- **WHEN** the user browses to and selects a directory that already contains `openspec/specs` and `openspec/changes`
- **THEN** the Studio registers it as an active project and loads its specs and changes

#### Scenario: Open directory without OpenSpec structure
- **WHEN** the user selects a directory with no `openspec/` folder via Open
- **THEN** the Studio does not create any files, rejects the selection, and tells the user no valid OpenSpec project exists there - directing them to "New" to scaffold one from scratch instead

#### Scenario: Directory flagged while browsing
- **WHEN** the user is browsing directories in the picker
- **THEN** each listed directory is visually flagged as already containing a valid OpenSpec project or not, before the user commits to selecting it

### Requirement: Multi-project registry persists across sessions
The system SHALL persist the list of opened projects (name, absolute path, last-opened timestamp) locally so they remain available after restarting the Studio.

#### Scenario: Reopen app shows previously opened projects
- **WHEN** the user restarts OpenSpec Studio
- **THEN** all previously opened projects appear in the project switcher without needing to be re-selected

#### Scenario: Registered project path no longer exists
- **WHEN** a registered project's path is missing or inaccessible on disk
- **THEN** the Studio marks that project as unavailable in the switcher instead of crashing or silently dropping it

### Requirement: Active project selection scopes the UI
The system SHALL let the user switch which registered project is active, and SHALL scope the Kanban, Specs, and Overview views to that selection (or an "all projects" view where supported).

#### Scenario: Switching active project changes visible data
- **WHEN** the user selects a different project in the project switcher
- **THEN** the Kanban and Specs views reload to show only that project's changes and specs, and the Overview view reloads to show only that project's summary card and totals

#### Scenario: Selecting "all projects" restores the unscoped view
- **WHEN** the user selects "all projects" in the project switcher
- **THEN** the Kanban, Specs, and Overview views reload to show data across every registered project, as before a project was selected
