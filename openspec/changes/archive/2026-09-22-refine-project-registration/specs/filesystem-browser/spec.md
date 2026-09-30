## Purpose

Lets the Studio's Open and New project flows browse the user's real local filesystem from inside the app - a Phase-1 (web) stand-in for a native OS folder picker, backed by the Go backend's own local filesystem access.

## ADDED Requirements

### Requirement: Browse directories
The system SHALL let the user navigate the local filesystem's real directory tree from within the app: starting from a sensible default location, drilling into a subdirectory, and navigating back up to a parent directory.

#### Scenario: Start browsing
- **WHEN** the user opens the directory browser
- **THEN** it starts at a sensible default location (e.g. the user's home directory) and lists that directory's subdirectories

#### Scenario: Drill into a subdirectory
- **WHEN** the user selects a listed subdirectory to browse into
- **THEN** the browser navigates into it and lists its subdirectories

#### Scenario: Navigate to a parent directory
- **WHEN** the user navigates up from the current directory
- **THEN** the browser shows the parent directory's contents

### Requirement: Flag OpenSpec projects and empty directories
For each listed directory, the system SHALL indicate whether it already contains a valid `openspec/` structure and whether it is empty, so the picker can present this to the user before a directory is selected.

#### Scenario: Existing project flagged
- **WHEN** a listed directory already contains `openspec/specs` and `openspec/changes`
- **THEN** it is flagged as an existing OpenSpec project in the listing

#### Scenario: Empty directory flagged
- **WHEN** a listed directory has no entries (aside from dotfiles)
- **THEN** it is flagged as empty in the listing

### Requirement: Read-only, local-only browsing
The system SHALL only read directory listings from the local filesystem the backend runs on, and SHALL NOT create, modify, or delete anything while browsing - only the subsequent Open/New confirmation performs any write (New's scaffold, or Open's registration).

#### Scenario: Browsing performs no writes
- **WHEN** the user browses through several directories without confirming a selection
- **THEN** no files or directories are created, modified, or deleted anywhere on disk
