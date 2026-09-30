// Package instancelock guarantees a single running backend per Studio data
// directory. Two backends sharing one studio.db would each run their own
// Specflow scheduler and execute the same due flow twice.
//
// The lock is an OS advisory lock on <data-dir>/studio.lock, so it is
// released by the kernel when the holder dies (crash, SIGKILL) and never
// goes stale. The holder writes its bound address into the file, which lets
// a second backend report where the Studio is already running and lets the
// desktop shell attach to it.
package instancelock

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const fileName = "studio.lock"

// HeldError is returned by Acquire when another live process holds the lock.
type HeldError struct {
	// Addr is the address the holder recorded, or "" if it has not bound yet.
	Addr string
}

func (e *HeldError) Error() string {
	if e.Addr == "" {
		return "another OpenSpec Studio backend is already running for this data directory"
	}
	return fmt.Sprintf("another OpenSpec Studio backend is already running at %s", e.Addr)
}

// Lock is a held instance lock. Keep it alive for the life of the process.
type Lock struct {
	f *os.File
}

// Acquire takes the exclusive lock for dir without blocking.
func Acquire(dir string) (*Lock, error) {
	path := filepath.Join(dir, fileName)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open instance lock: %w", err)
	}
	if err := tryLock(f); err != nil {
		f.Close()
		if errors.Is(err, errWouldBlock) {
			return nil, &HeldError{Addr: ReadAddr(dir)}
		}
		return nil, fmt.Errorf("lock %s: %w", path, err)
	}
	return &Lock{f: f}, nil
}

// SetAddr records the address this backend is serving on.
func (l *Lock) SetAddr(addr string) error {
	if err := l.f.Truncate(0); err != nil {
		return err
	}
	if _, err := l.f.WriteAt([]byte(addr+"\n"), 0); err != nil {
		return err
	}
	return l.f.Sync()
}

// Release clears the recorded address and drops the lock.
func (l *Lock) Release() {
	_ = l.f.Truncate(0)
	_ = unlock(l.f)
	_ = l.f.Close()
}

// ReadAddr returns the address recorded in dir's lock file, or "" when
// there is none. It does not tell whether the recorder is still alive.
func ReadAddr(dir string) string {
	b, err := os.ReadFile(filepath.Join(dir, fileName))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}
