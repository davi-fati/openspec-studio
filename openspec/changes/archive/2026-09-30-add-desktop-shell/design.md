## Context

The Go backend (`backend/cmd/server`) binds `127.0.0.1:4173`, stores Studio state in `os.UserConfigDir()/openspec-studio/studio.db`, and runs the Specflow scheduler in-process. The React frontend calls relative `/api/...` URLs and opens `EventSource('/api/events')`; in dev, Vite proxies `/api` to the backend. The backend does not serve the frontend today, so web use requires the Vite dev server. See proposal.md for motivation.

## Goals / Non-Goals

**Goals:**
- One frontend build and one backend binary shared by desktop and web; no mode-specific API client.
- The desktop app never runs a second scheduler against the same database.
- `make dev` stays exactly as it is.

**Non-Goals:**
- Windows/Linux packaging, code signing/notarization, auto-update (later changes).
- Authentication on the local API. It is already reachable by any local process in web mode; this change does not widen that.
- Rewriting any backend logic in Rust or moving features to Tauri commands.

## Decisions

### 1. The webview loads the UI from the backend's HTTP origin
The desktop app navigates its webview to `http://127.0.0.1:<port>/`, served by the sidecar backend, instead of loading the bundled frontend from Tauri's `tauri://` origin.

- Keeps relative `/api` and same-origin SSE working with zero frontend changes; no CORS, no base-URL injection.
- The same embedded UI is what web mode serves, so desktop and web cannot drift.
- Alternative considered: bundle `frontend/dist` in Tauri and inject an API base URL. Rejected: cross-origin `EventSource`, CORS config on the backend, and a second code path for every request.
- Consequence: Tauri IPC (needed only for the native dialog and the title-bar drag region, Decision 8) must be granted to the remote origin `http://127.0.0.1:*` via a narrowly scoped capability (dialog open, window start-dragging and toggle-maximize).

### 2. Frontend embedded in the Go binary
`frontend/dist` is embedded with `go:embed` behind a build step (`make build` builds the frontend first, then the backend). Gin serves static files and falls back to `index.html` for non-`/api`, non-`/swagger` paths (`NoRoute`). A dev build without an embedded UI still works (API only), so `go run` in `make dev` is unaffected.

### 3. Sidecar lifecycle
- Tauri bundles the backend as `externalBin` (`openspec-studio-server-<target-triple>`), built by `make desktop-build`.
- On launch, the shell first checks for a live backend via the data-dir lock file (Decision 4). If one is healthy, it attaches (loads its URL, does not own it).
- Otherwise it spawns the sidecar with `--port 0`, reads stdout for the readiness line `OPENSPEC_STUDIO_READY addr=127.0.0.1:NNNNN`, confirms with `GET /api/healthz`, then navigates the webview. A 15s timeout shows an error page with captured stderr.
- On quit (owned backend only) the shell sends SIGTERM and the backend shuts down gracefully: cancels running flows via the existing cancel path (process-group kill), closes the DB. SIGKILL after a grace period.
- Alternative considered: run the backend in-process via a Rust FFI build of Go. Rejected: brittle toolchain and no benefit over a sidecar.

### 4. Data-dir lock
The backend takes an OS advisory lock (`flock`) on `<data-dir>/studio.lock` and writes its bound address into it after binding. Advisory locks are released by the OS when the process dies, so crashes leave no stale lock. A second backend fails to acquire it, reads the address, and exits with that address in the error. The desktop shell uses the same file to discover an existing backend.
- Alternative considered: PID file. Rejected: stale-PID detection is racy and PID reuse makes it unreliable.

### 5. Login-shell `PATH`
When the backend starts, if it detects it was not launched from a terminal (desktop sidecar: an explicit `--inherit-login-path` flag passed by the shell), it runs `$SHELL -l -i -c 'printf %s "$PATH"'` once with a short timeout and uses that `PATH` for `exec.LookPath` and child processes. On failure it keeps the inherited `PATH` and logs a warning. This lives in the backend, not Tauri, so every provider call path gets it.

### 6. Background mode
The Tauri shell intercepts window close. It asks the backend (`GET /api/specflow/flows`) whether any flow is `pending`/`running`; if so it hides the window and shows a tray icon (Open Studio / Quit). Otherwise it quits. Tray Quit with a running flow shows a native confirm dialog, then quits (graceful shutdown cancels the flow per Decision 3).

### 7. Folder picker selection
The frontend checks `'__TAURI_INTERNALS__' in window`. In desktop mode `DirectoryBrowserDialog` is replaced by `@tauri-apps/plugin-dialog`'s `open({ directory: true })`; the chosen path then goes through the existing Open/New API calls. Web mode is unchanged.

### 8. Unified title bar on macOS
The window uses `titleBarStyle: "Overlay"` with `hiddenTitle: true`, and `trafficLightPosition` centers the window controls in the Studio's 56px header, so the header is the title bar, as in Finder, Xcode or Linear. The default separate title bar showed a centered title above an app header that already names the Studio.

- In desktop mode only, the header reserves room on the left for the window controls, except in fullscreen where macOS hides them. Fullscreen is detected from the viewport filling the screen, so no window-state API is granted to the webview.
- The header carries `data-tauri-drag-region="deep"`: empty areas and the logo drag the window and double-click zooms. Tauri's drag script already skips buttons, selects, links and tab roles, so the header's controls keep working.
- This needs `core:window:allow-start-dragging` and `core:window:allow-internal-toggle-maximize` on the remote-origin capability. Neither reaches files, processes or other windows. The bundled splash/error page gets the same two permissions (its own local-origin capability) for a drag strip at the top, so the window can be moved while the backend starts or after a startup error.
- Alternative considered: `titleBarStyle: "Transparent"`. Rejected: it only tints the bar, which still takes its own row.

## Risks / Trade-offs

- [Remote-origin IPC grants native APIs to whatever is served on 127.0.0.1] → capability limited to `dialog:allow-open` plus window drag/zoom for the title bar; the tray is driven from Rust with no webview command; window only ever navigates to the address the shell itself verified.
- [Attaching to a web-mode backend of a different Studio version] → `/api/healthz` returns a version; on mismatch the shell shows a warning with the running backend's address instead of attaching silently.
- [Login shell with slow or interactive profile hangs startup] → 3s timeout, fall back to inherited `PATH`, log it; provider health check surfaces the "not found" as before.
- [Unsigned macOS app is blocked by Gatekeeper] → acceptable for local use; documented (`xattr -dr com.apple.quarantine` or right-click Open). Signing is a later change.
- [Rust toolchain now needed] → only for `desktop-*` targets; web build and `make dev` do not require it.

## Migration Plan

No data migration: the desktop app uses the same data directory as web mode. Existing users keep running `make dev` or the web binary. Rollback is deleting the app bundle.
