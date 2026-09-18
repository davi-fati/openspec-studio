## Purpose

Owns the desktop application lifecycle: launching the Tauri 2 window and supervising the embedded Go backend process it depends on.

## ADDED Requirements

### Requirement: Backend process supervision
The desktop shell SHALL start the Go backend as a child process on app launch and terminate it when the app window closes.

#### Scenario: App launch starts backend
- **WHEN** the user launches OpenSpec Studio
- **THEN** the Tauri shell starts the Go backend process and waits for it to report ready before showing the main window

#### Scenario: App close stops backend
- **WHEN** the user closes the OpenSpec Studio window
- **THEN** the Tauri shell terminates the Go backend process and does not leave it running

#### Scenario: Backend crash is surfaced
- **WHEN** the Go backend process exits unexpectedly while the app is open
- **THEN** the shell shows an error state to the user instead of a silently frozen UI

### Requirement: Local-only backend binding
The Go backend SHALL bind only to a local loopback address and SHALL NOT be reachable from outside the machine.

#### Scenario: Backend port is loopback-only
- **WHEN** the backend starts
- **THEN** it listens on `127.0.0.1` on a port chosen at startup, not `0.0.0.0`
