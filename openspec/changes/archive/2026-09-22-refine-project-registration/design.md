## Context

Builds on the existing `project-registry` capability (`add-studio-foundation`, already archived) and its `ProjectService`/`ProjectHandler` (`backend/internal/service/project.go`, `backend/internal/handler/projects.go`) and the frontend `ProjectSwitcher`/`ProjectsContext`. See proposal.md for motivation.

## Goals / Non-Goals

**Goals:**
- Real absolute-path directory browsing that works today in Phase 1 (web), without a native OS dialog.
- Clean separation: Open never writes to disk; only New (already the case) and the browsing-triggered registration write anything.
- Reuse `ProjectService.OpenProject`/`CreateProject` as the single source of truth for what counts as "valid" or "empty" - the browser's flags and the final confirm action must never disagree.

**Non-Goals:**
- A native Tauri OS dialog - that's Phase 2. This browser is explicitly a stand-in with the same backend contract underneath, swapped later.
- General-purpose file browsing (viewing/opening files, non-directory entries) - directories only, scoped to picking a project root.
- Filesystem write operations from the browser itself (create folder, rename, delete) - out of scope; the only writes remain New's scaffold and Open's registration.

## Decisions

- **New endpoint `GET /api/fs/browse?path=<abs path>`** (path omitted defaults to the user's home directory via `os.UserHomeDir()`). Returns the resolved absolute path, its parent (or empty if at a root), and a list of subdirectories, each with `{name, path, hasOpenspec, isEmpty}`. Directories only (no files) - this endpoint exists solely to pick a project root, not to browse file contents.
- **"Has openspec" and "is empty" reuse existing logic, not new checks**: `hasOpenspec` is computed the same way `OpenProject`'s validation will check it (presence of `openspec/specs` and `openspec/changes`, or more precisely a readable `openspec/` dir - kept as one small shared helper so the browser's flag and the actual Open/New validation can never drift apart). `isEmpty` reuses the same "only dotfiles" rule `CreateProject` already applies.
- **`OpenProject` becomes read-only-if-valid**: replace the current `ensureOpenSpecScaffold` call in `OpenProject` with a validation check; return a new `ErrNoOpenSpecProject` sentinel (400) when the directory lacks a valid structure, instead of scaffolding. `CreateProject` is untouched - it already requires an empty directory and already scaffolds; New's contract doesn't change, only how the user reaches it (via the browser instead of a text input).
- **One shared `DirectoryBrowserDialog` component, opened in two modes** (`'open'` | `'new'`) from the existing Open/New buttons: in `'open'` mode, only directories flagged `hasOpenspec` are confirmable (others show why they're disabled, pointing at "New"); in `'new'` mode, only directories flagged `isEmpty` are confirmable. Both modes share the same browsing/navigation code, differing only in which flag gates the confirm action and which API call (`openProject` vs `createProject`) confirming makes.
- **Silent permission errors are skipped, not fatal**: `os.ReadDir` entries the process can't stat (permission-denied subdirectories, broken symlinks) are omitted from the listing rather than failing the whole browse request - a real local filesystem always has a few of these, and one bad entry shouldn't block picking a project three folders up.

## Risks / Trade-offs

- Every browse request does a live `os.ReadDir` + per-entry stat of the real filesystem → acceptable for a local dev tool on typical directory sizes; not optimized for extremely large directories (thousands of siblings), which isn't the expected use case for "where's my project root."
- The endpoint only limits itself to directories the OS user running the backend can already read - no additional sandboxing beyond normal file permissions, consistent with the rest of the Studio's "local, single-user, trusted machine" threat model (`add-studio-foundation`'s loopback-only binding decision).
