## 1. Backend: directory browsing

- [x] 1.1 Implement a shared `hasOpenSpecStructure(dir string) bool` helper (presence of `openspec/specs` and `openspec/changes`) reused by both the browse endpoint and `OpenProject`'s validation, and verify a unit test covers both a valid and an invalid directory
- [x] 1.2 Implement `service.BrowseDirectory(path string) (BrowseResult, error)`: defaults to the OS home directory when path is empty, lists real subdirectories with `{name, path, hasOpenspec, isEmpty}`, skips entries that error on stat (permission-denied, broken symlinks) instead of failing the whole call, and verify a unit test against a temp directory tree with a mix of valid-project, empty, and non-empty-non-project subdirectories
- [x] 1.3 Implement `GET /api/fs/browse?path=` returning the resolved path, parent path (empty at a filesystem root), and the entry list, and verify it against a real directory via curl
- [x] 1.4 Verify the endpoint defaults to the home directory when `path` is omitted, and rejects a `path` that does not exist or is not a directory with a 400

## 2. Backend: Open becomes read-only-if-valid

- [x] 2.1 Add `ErrNoOpenSpecProject` sentinel error and change `ProjectService.OpenProject` to return it (instead of scaffolding) when the target directory lacks a valid `openspec/` structure, and verify a unit test covers both the now-rejected case and the still-accepted valid-project case
- [x] 2.2 Wire `ProjectHandler.Open` to return 400 with a clear message on `ErrNoOpenSpecProject`, distinct from the existing "path does not exist" message, and verify via curl that opening a non-OpenSpec directory is rejected without creating any files
- [x] 2.3 Verify (regression) opening a directory that already has a valid `openspec/` structure still registers it exactly as before, via curl

## 3. Frontend: directory browser dialog

- [x] 3.1 Add `browseDirectory(path?: string)` to the API client calling `GET /api/fs/browse` and add `BrowseEntry`/`BrowseResult` types, and verify a manual fetch against the running backend returns the expected shape
- [x] 3.2 Build `DirectoryBrowserDialog` component: breadcrumb/current-path display, drill into a subdirectory, navigate to parent, list entries with an "OpenSpec project" or "empty" badge per the design's flags, and verify it renders and navigates using real backend data
- [x] 3.3 Add an `'open' | 'new'` mode prop: in `'open'` mode only `hasOpenspec` entries are selectable/confirmable (others show a disabled state with a hint), in `'new'` mode only `isEmpty` entries are, and verify both modes correctly gate the confirm action against real directory data
- [x] 3.4 Wire the confirm action to call `openProjectByPath`/`createProjectByPath` depending on mode, closing the dialog and switching the active project on success, and verify end-to-end that confirming a valid selection registers the project and populates Kanban/Specs/Overview

## 4. Frontend: wire into ProjectSwitcher

- [x] 4.1 Replace `ProjectSwitcher`'s free-text Open input with a button that opens `DirectoryBrowserDialog` in `'open'` mode, and verify the old text input is gone
- [x] 4.2 Replace the New button's direct text-path flow with a button that opens `DirectoryBrowserDialog` in `'new'` mode, and verify it
- [x] 4.3 Verify end-to-end: browsing to and selecting a directory with no `openspec/` structure via Open shows the rejection message directing to New, without creating any files, and browsing to and selecting an empty directory via New scaffolds and registers it as before
