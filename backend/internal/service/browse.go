package service

import (
	"fmt"
	"os"
	"path/filepath"
)

// BrowseEntry is one subdirectory in a browse listing, flagged for the
// picker: whether it's already a valid OpenSpec project, and whether it's
// empty (so New can gate on it separately from Open).
type BrowseEntry struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	HasOpenspec bool   `json:"hasOpenspec"`
	IsEmpty     bool   `json:"isEmpty"`
}

// BrowseResult is one directory's contents for the in-app filesystem
// browser. HasOpenspec/IsEmpty describe the current directory (Path) itself
// - not just its listed children - so the picker can gate "use this folder"
// on the directory the user has actually navigated into.
type BrowseResult struct {
	Path        string        `json:"path"`
	Parent      string        `json:"parent,omitempty"` // empty at a filesystem root
	HasOpenspec bool          `json:"hasOpenspec"`
	IsEmpty     bool          `json:"isEmpty"`
	Entries     []BrowseEntry `json:"entries"`
}

// ErrBrowsePathNotFound is returned when the requested browse path does not
// exist or is not a directory.
var ErrBrowsePathNotFound = fmt.Errorf("path does not exist or is not a directory")

// BrowseDirectory lists path's real subdirectories for the in-app directory
// browser (Open/New pickers), read-only - it never creates, modifies, or
// deletes anything. path defaults to the OS user's home directory when
// empty. Entries that error on stat (permission-denied, broken symlinks)
// are skipped rather than failing the whole listing.
func (s *ProjectService) BrowseDirectory(path string) (BrowseResult, error) {
	resolved := path
	if resolved == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return BrowseResult{}, fmt.Errorf("resolve home directory: %w", err)
		}
		resolved = home
	}

	info, err := os.Stat(resolved)
	if err != nil || !info.IsDir() {
		return BrowseResult{}, ErrBrowsePathNotFound
	}

	dirEntries, err := os.ReadDir(resolved)
	if err != nil {
		return BrowseResult{}, fmt.Errorf("read directory: %w", err)
	}

	entries := make([]BrowseEntry, 0, len(dirEntries))
	for _, e := range dirEntries {
		if !e.IsDir() {
			continue
		}
		if len(e.Name()) > 0 && e.Name()[0] == '.' {
			continue // skip dotfiles/hidden dirs - not useful project-root candidates
		}
		full := filepath.Join(resolved, e.Name())
		if _, statErr := os.Stat(full); statErr != nil {
			continue // permission-denied or broken symlink - skip, don't fail the listing
		}
		empty, emptyErr := isEmptyDir(full)
		if emptyErr != nil {
			continue
		}
		entries = append(entries, BrowseEntry{
			Name:        e.Name(),
			Path:        full,
			HasOpenspec: hasOpenSpecStructure(full),
			IsEmpty:     empty,
		})
	}

	parent := filepath.Dir(resolved)
	if parent == resolved {
		parent = ""
	}

	empty, err := isEmptyDir(resolved)
	if err != nil {
		return BrowseResult{}, fmt.Errorf("read directory: %w", err)
	}

	return BrowseResult{
		Path:        resolved,
		Parent:      parent,
		HasOpenspec: hasOpenSpecStructure(resolved),
		IsEmpty:     empty,
		Entries:     entries,
	}, nil
}
