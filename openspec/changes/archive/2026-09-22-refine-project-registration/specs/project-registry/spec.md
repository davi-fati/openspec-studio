## MODIFIED Requirements

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
