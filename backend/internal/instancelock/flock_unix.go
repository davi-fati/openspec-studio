//go:build unix

package instancelock

import (
	"os"
	"syscall"
)

var errWouldBlock = syscall.EWOULDBLOCK

func tryLock(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
}

func unlock(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
}
