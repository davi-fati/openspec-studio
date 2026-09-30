## Why

Today "Open" and "New" behave almost identically: both take a raw path typed into a text input, and both silently scaffold an `openspec/` structure if one is missing (`ProjectService.OpenProject` calls the same `ensureOpenSpecScaffold` as `CreateProject`). This blurs a meaningful distinction - Open should be for discovering an *existing* OpenSpec project, New should be for starting a *fresh* one - and typing an absolute path by hand is also just bad UX for browsing a real filesystem. Since the Studio's Go backend already runs locally with real filesystem access, it can power an in-app directory browser today (Phase 1, web) without waiting for the Tauri desktop dialog in Phase 2.

## What Changes

- Add a Go backend endpoint that lists a directory's real subdirectories, flagging per-entry whether it already contains a valid `openspec/` structure and whether it is empty - the data an in-app folder browser needs to render.
- Add a shared frontend directory-browser dialog (breadcrumb + drill in/up) used by both the "Open" and "New" buttons, replacing the current raw path text inputs.
- **BREAKING** (behavioral): change Open's semantics from "open-or-scaffold" to **read-only discovery**. Opening a directory without a valid `openspec/` structure SHALL NOT scaffold one - it SHALL reject with guidance pointing the user to "New" instead. Only "New" scaffolds, and only into an empty directory (already true today).

## Capabilities

### New Capabilities
- `filesystem-browser`: Lists real subdirectories of a path, flagging which contain a valid OpenSpec project and which are empty, powering an in-app folder picker for both Open and New.

### Modified Capabilities
- `project-registry`: "Open a project directory" requirement changes from open-or-scaffold to open-only-if-valid, with clear rejection guidance when the chosen directory has no `openspec/` structure.

## Impact

- Backend: new read-only `GET /api/fs/browse` endpoint (or similar); `ProjectService.OpenProject` loses its scaffold-on-missing behavior (an `ErrNoOpenSpecProject`-style rejection replaces it). `CreateProject`/`POST /api/projects/new` is unchanged - New already requires an empty directory and already scaffolds.
- Frontend: `ProjectSwitcher`'s two text inputs are replaced by a shared `DirectoryBrowserDialog` (or similar) launched by the Open and New buttons; `ProjectsContext`'s `openProjectByPath`/`createProjectByPath` callers change from "type a path, submit" to "browse, pick, confirm."
- Phase-2 note: this in-app browser is a Phase-1 (web) stand-in. When the Tauri desktop shell lands, the picker UI can be swapped for the native OS dialog while the backend `OpenProject`/`CreateProject` validation stays the same.
