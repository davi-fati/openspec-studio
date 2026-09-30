package service

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// ProjectWatcher watches every directory under each registered project's
// openspec/ tree (recursively - fsnotify has no native recursive mode) and
// publishes an "update" event on any file change within it.
type ProjectWatcher struct {
	watcher *fsnotify.Watcher
	events  *EventBroadcaster

	mu      sync.Mutex
	watched map[string]struct{}
}

func NewProjectWatcher(events *EventBroadcaster) (*ProjectWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &ProjectWatcher{watcher: w, events: events, watched: make(map[string]struct{})}, nil
}

// WatchProject recursively adds every directory under projectPath/openspec
// to the watch list, so changes deep inside a change or spec directory
// (e.g. editing that change's tasks.md) are seen, not just top-level
// specs/changes folders. Safe to call repeatedly, including after the
// directory was removed and recreated.
func (pw *ProjectWatcher) WatchProject(projectPath string) error {
	root := filepath.Join(projectPath, "openspec")

	pw.mu.Lock()
	defer pw.mu.Unlock()

	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return nil // best-effort: skip unreadable entries rather than aborting the whole walk
		}
		if !d.IsDir() {
			return nil
		}
		pw.addLocked(path)
		return nil
	})
}

func (pw *ProjectWatcher) addLocked(dir string) {
	if _, already := pw.watched[dir]; already {
		return
	}
	if err := pw.watcher.Add(dir); err != nil {
		return
	}
	pw.watched[dir] = struct{}{}
}

// Run consumes filesystem events until the watcher is closed, publishing an
// "update" event for every change. A newly created directory is added to the
// watch set on the fly, so a fresh change/spec directory starts being
// observed without a full project re-scan. A removed watched directory is
// evicted so a later WatchProject call can re-arm it once it reappears.
func (pw *ProjectWatcher) Run() {
	for {
		select {
		case ev, ok := <-pw.watcher.Events:
			if !ok {
				return
			}
			log.Printf("openspec change detected: %s (%s)", ev.Name, ev.Op)

			if ev.Op.Has(fsnotify.Create) {
				if info, err := os.Stat(ev.Name); err == nil && info.IsDir() {
					pw.mu.Lock()
					pw.addLocked(ev.Name)
					pw.mu.Unlock()
				}
			}
			if ev.Op.Has(fsnotify.Remove) || ev.Op.Has(fsnotify.Rename) {
				pw.mu.Lock()
				for watched := range pw.watched {
					if watched == ev.Name || strings.HasPrefix(watched, ev.Name+string(filepath.Separator)) {
						delete(pw.watched, watched)
					}
				}
				pw.mu.Unlock()
			}

			pw.events.Publish(Event{Name: "update", Data: ev.Name})
		case err, ok := <-pw.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("watcher error: %v", err)
		}
	}
}

func (pw *ProjectWatcher) Close() error {
	return pw.watcher.Close()
}
