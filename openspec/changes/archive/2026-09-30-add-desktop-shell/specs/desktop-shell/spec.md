## Purpose

Runs OpenSpec Studio as a desktop application that owns the backend's lifecycle, integrates with native OS facilities, and stays available for scheduled Specflow runs.

## ADDED Requirements

### Requirement: Desktop app launches and supervises the backend
The desktop app SHALL start the bundled Studio backend on launch, SHALL show the Studio UI only after the backend reports healthy, and SHALL stop the backend it started when the app quits.

#### Scenario: Cold launch
- **WHEN** the user opens the desktop app and no Studio backend is running
- **THEN** the app starts its bundled backend, waits until it is healthy, and then shows the Studio UI

#### Scenario: Backend fails to start
- **WHEN** the bundled backend exits or does not become healthy within a bounded startup timeout
- **THEN** the app shows an error screen with the failure reason and backend output instead of a blank or broken UI

#### Scenario: Quit stops the backend
- **WHEN** the user quits the desktop app that started the backend
- **THEN** the backend process and any provider processes it spawned are terminated

### Requirement: Desktop app attaches to an existing backend
The desktop app SHALL NOT start a second backend when a healthy Studio backend is already running against the same data directory; it SHALL connect to that backend instead and SHALL NOT stop it on quit.

#### Scenario: Web backend already running
- **WHEN** the user opens the desktop app while `openspec-studio-server` is already running for the same user
- **THEN** the app shows the UI served by the running backend, and quitting the app leaves that backend running

### Requirement: Web mode remains available
The Studio SHALL remain fully usable from a regular browser without the desktop app, with the same features and data, except where a desktop-only integration has a documented web fallback.

#### Scenario: Same data in both modes
- **WHEN** the user registers a project in the desktop app, quits it, and then opens the Studio in a browser via the standalone backend
- **THEN** the project and its Specflow flows are present in the browser

### Requirement: Native folder picker in desktop mode
In the desktop app, the Open and New project flows SHALL use the operating system's native folder picker; in web mode they SHALL keep using the in-app directory browser.

#### Scenario: Open project in desktop
- **WHEN** the user chooses Open project in the desktop app
- **THEN** the native OS folder dialog appears and the selected folder is registered through the same validation as the web flow

#### Scenario: Picker cancelled
- **WHEN** the user dismisses the native folder dialog without selecting a folder
- **THEN** no project is opened or created and no error is shown

### Requirement: Background running while flows are pending
When the user closes the desktop app's window while any Specflow flow is scheduled or running, the app SHALL keep running in the background with a tray/menu-bar presence from which the window can be reopened or the app quit. With no pending flows, closing the window SHALL quit the app.

#### Scenario: Close window with a scheduled overnight flow
- **WHEN** the user closes the window while a flow is scheduled for later
- **THEN** the app keeps running in the background and the flow starts at its scheduled time

#### Scenario: Close window with nothing pending
- **WHEN** the user closes the window and no flow is scheduled or running
- **THEN** the app quits and the backend it started is stopped

#### Scenario: Quit from tray with a running flow
- **WHEN** the user chooses Quit from the tray menu while a flow is running
- **THEN** the app asks for confirmation, and on confirm cancels the running flow the same way a user cancel does before exiting

### Requirement: CLI providers see the user's shell environment
When launched from the desktop app, CLI providers, provider health checks, and Specflow runs SHALL resolve executables using the user's login-shell `PATH`, so a CLI that works in the user's terminal also works in the desktop app.

#### Scenario: CLI installed via a shell-managed path
- **WHEN** a CLI provider's executable is only on the `PATH` set by the user's shell profile (e.g. under `~/.local/bin` or a Node version manager) and the app is launched from the Dock or Finder
- **THEN** the provider health check succeeds, just as it does when the backend is started from a terminal

### Requirement: Native macOS window chrome
On macOS the desktop app's window SHALL integrate its title bar into the Studio header, with the window controls inside the header and no separate title row, and the header SHALL behave like a native title bar for moving and zooming the window. The web mode header SHALL be unaffected.

#### Scenario: Window controls in the header
- **WHEN** the user opens the desktop app on macOS
- **THEN** the close, minimize and zoom controls appear inside the Studio header without overlapping the logo, and no separate title bar row is shown

#### Scenario: Move and zoom from the header
- **WHEN** the user drags an empty area of the header, or double-clicks it
- **THEN** the window moves, or zooms, as with a native title bar, while the header's tabs, buttons and project selector keep working as before

#### Scenario: Fullscreen
- **WHEN** the desktop window enters fullscreen
- **THEN** the header no longer reserves space for the hidden window controls

#### Scenario: Web mode
- **WHEN** the Studio is opened in a regular browser
- **THEN** the header layout is the same as before this requirement

