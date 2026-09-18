## Purpose

Lets a user open one or more real project directories in the Studio and keeps track of which ones are active across sessions.

## ADDED Requirements

### Requirement: Open a project directory
The system SHALL allow the user to select a directory to open as a project. If the directory has no `openspec/` structure, the system SHALL offer to initialize one following the OpenSpec framework conventions.

#### Scenario: Open existing OpenSpec project
- **WHEN** the user selects a directory that already contains `openspec/specs` and `openspec/changes`
- **THEN** the Studio registers it as an active project and loads its specs and changes

#### Scenario: Open directory without OpenSpec structure
- **WHEN** the user selects a directory with no `openspec/` folder
- **THEN** the Studio prompts to initialize OpenSpec structure in that directory before registering it as a project

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
- **THEN** the Kanban and Specs views reload to show only that project's changes and specs
