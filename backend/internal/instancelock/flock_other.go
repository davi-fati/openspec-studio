//go:build !unix

package instancelock

import (
	"errors"
	"os"
)

// Only macOS is packaged today; other platforms run without the guard
// until they get a native implementation.
var errWouldBlock = errors.New("would block")

func tryLock(*os.File) error { return nil }

func unlock(*os.File) error { return nil }
