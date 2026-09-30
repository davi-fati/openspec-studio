//go:build !windows

package service

import (
	"os/exec"
	"syscall"
)

// setNewProcessGroup puts cmd in its own process group before Start(), so
// killProcessGroup can later terminate it AND every child process it spawns
// (e.g. a wrapper script's own subprocesses) - killing only cmd.Process
// directly leaves orphaned children holding the stdout/stderr pipes open,
// which hangs our log-draining goroutines indefinitely.
func setNewProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// killProcessGroup sends SIGKILL to cmd's entire process group.
func killProcessGroup(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}

// processAlive reports whether pid still exists (signal 0 probes without
// actually sending a signal). Test-only helper, kept here instead of in a
// _test.go file so it stays behind the same unix build tag as syscall.Kill.
func processAlive(pid int) bool {
	return syscall.Kill(pid, 0) == nil
}
