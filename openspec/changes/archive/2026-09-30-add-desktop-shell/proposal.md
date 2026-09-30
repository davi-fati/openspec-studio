## Why

Phase 1 closed out the core feature set as a web app that needs two terminals (`make dev`) and a browser tab. Specflow in particular wants a long-lived, always-available process (overnight runs), which a browser tab plus a manually started backend handles poorly. Packaging the Studio as a Tauri desktop app gives it a real app lifecycle (launch, background, quit) while the same backend and frontend keep working as a plain web app.

## What Changes

- Add a **Tauri 2 desktop shell** (`src-tauri/`) that bundles the Go backend as a sidecar, starts it on launch, waits for it to become healthy, and loads the Studio UI in its webview. Quitting the app stops the backend.
- Keep **web mode** as a first-class target: the backend binary serves the built frontend itself, so `openspec-studio-server` alone is enough to use the Studio from a browser. The current dev workflow (`make dev`, Vite proxy) is unchanged.
- Enforce **one backend per Studio data directory**: a second backend refuses to start against the same SQLite database, and the desktop app attaches to an already-running backend instead of spawning a competing one. This prevents two schedulers from running the same Specflow flow twice.
- In the desktop app, use the **native OS folder picker** for Open/New project; the in-app directory browser remains the web-mode picker.
- In the desktop app, closing the window while flows are scheduled or running **keeps the Studio running in the background** (tray/menu bar) instead of quitting, so scheduled runs are not silently missed.
- Make CLI providers resolve from the **user's login-shell `PATH`** when launched from the desktop, since GUI apps on macOS do not inherit the terminal environment (otherwise `claude`, `agent`, `openspec` are "not found").
- Add `make desktop-dev` / `make desktop-build`; `make build` produces a single self-contained web binary.

## Capabilities

### New Capabilities
- `desktop-shell`: Desktop app lifecycle - launching and supervising the backend, native folder picker, background running while flows are pending, and login-shell environment for CLI providers.

### Modified Capabilities
- `backend-api-contract`: Backend serves the frontend UI itself (web mode without a dev server) and enforces a single running backend per Studio data directory.

## Impact

- New `src-tauri/` (Rust, Tauri 2 + `tauri-plugin-shell`, `tauri-plugin-dialog`); Rust toolchain becomes a build dependency for desktop builds only.
- Backend: `cmd/server` gains frontend embedding, ephemeral-port mode with a machine-readable ready line, and a data-dir lock.
- Frontend: runtime detection (desktop vs web) for the folder picker; API calls stay relative (`/api`), no base-URL plumbing.
- `openspec/project.md`: Phase 2 moves from "deferred" to "in progress"; Running and Project Structure sections updated.
- Initial target is macOS (unsigned local build). Windows/Linux, code signing/notarization, and auto-update are out of scope.
