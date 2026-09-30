## 1. Backend: web-mode self-sufficiency

- [x] 1.1 Embed `frontend/dist` into the backend binary and serve it with an `index.html` fallback for non-`/api`, non-`/swagger` paths; verify a built binary alone serves the UI and a reload on `/specflow` returns the UI
- [x] 1.2 Keep unknown `/api/*` paths returning JSON 404 and verify with a handler test
- [x] 1.3 Allow building/running without an embedded UI (API-only) so `make dev` is unchanged; verify `make dev` still works via the Vite proxy
- [x] 1.4 Update `make build` to build the frontend before the backend and produce one self-contained binary; verify `make clean build` then running the binary serves the full Studio

## 2. Backend: process contract for a supervisor

- [x] 2.1 Support `--port 0` and print one `OPENSPEC_STUDIO_READY addr=<host:port>` line on stdout once listening; verify with a test that parses the line and hits `/api/healthz`
- [x] 2.2 Add the Studio version to `/api/healthz`; verify the response includes it
- [x] 2.3 Take an advisory lock on `<data-dir>/studio.lock` and record the bound address; verify a second backend exits non-zero naming the first one's address and does not start its scheduler
- [x] 2.4 Verify the lock is released after the backend is killed with SIGKILL (a new backend starts normally)
- [x] 2.5 Handle SIGTERM with graceful shutdown that cancels running flows through the existing cancel path and closes the DB; verify a running flow ends as cancelled and no provider process is left behind
- [x] 2.6 Add `--inherit-login-path` that resolves `PATH` from the user's login shell with a timeout and fallback; verify a CLI provider that is only on the shell-profile `PATH` passes its health check

## 3. Desktop shell (Tauri)

- [x] 3.1 Scaffold `src-tauri/` (Tauri 2) with the backend as `externalBin` and `tauri-plugin-shell`/`tauri-plugin-dialog`; verify `cargo tauri dev` opens a window
- [x] 3.2 Implement launch: attach to a live backend via the lock file, otherwise spawn the sidecar with `--port 0 --inherit-login-path`, wait for the ready line and healthz, then navigate the webview; verify both the cold-launch and attach paths
- [x] 3.3 Implement the startup error screen (timeout, early exit, version mismatch) showing captured backend output; verify by pointing the sidecar at a binary that exits immediately
- [x] 3.4 Stop the owned sidecar on quit (SIGTERM, then SIGKILL after a grace period) and never stop an attached backend; verify no `openspec-studio-server` process remains after quitting a cold-launched app
- [x] 3.5 Implement close-to-tray when flows are pending/running, tray menu (Open Studio / Quit), and confirm-on-quit with a running flow; verify a flow scheduled a minute out starts after the window is closed
- [x] 3.6 Scope the remote-origin capability to `http://127.0.0.1:*` with dialog-open plus window start-dragging/toggle-maximize only; verify other plugin APIs are rejected from the webview
- [x] 3.7 Use an overlay title bar on macOS (hidden title, window controls inside the Studio header): the header drags the window and double-click zooms while its controls stay clickable, and space for the window controls is reserved only in the desktop app and not in fullscreen; verify the web header is unchanged

## 4. Frontend: desktop integrations

- [x] 4.1 Add runtime detection (desktop vs web) and use the native folder dialog for Open/New project in desktop mode; verify the in-app directory browser is still used in a browser
- [x] 4.2 Verify cancelling the native dialog leaves state unchanged and shows no error

## 5. Build, docs, and verification

- [x] 5.1 Add `make desktop-dev` and `make desktop-build` (builds frontend, backend for the host target triple, then `cargo tauri build`); verify a `.app`/`.dmg` is produced on macOS
- [x] 5.2 Update `openspec/project.md` (Phase 2 in progress, Running, Project Structure with `src-tauri/`) and remove the Phase-2 note in `DirectoryBrowserDialog.tsx`; verify `openspec validate --strict` passes
- [x] 5.3 End-to-end check: register a project in the desktop app, quit, start the web binary, and confirm the project and flows are present in the browser
- [x] 5.4 Add the ⌘ brand icon: app icon set, monochrome template tray icon, web favicon/touch icon and header logo, all regenerated from SVG masters by `make icons`; verify the built `.app`, the tray and the browser tab show it
- [x] 5.5 Add `make desktop-stop` (quit the app, SIGTERM then SIGKILL its sidecar, leave a web-mode backend alone) and `make desktop-clean` (plus bundles, `src-tauri/target`, installed copy and webview data, keeping `studio.db`); verify no desktop process or artifact remains and `studio.db` is kept
