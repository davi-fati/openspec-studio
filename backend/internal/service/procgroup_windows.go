//go:build windows

package service

import "os/exec"

// Windows has no POSIX process groups in the same sense; falling back to
// killing the direct process only. Phase 1 targets macOS/Linux dev
// machines, so this is a documented gap rather than a full implementation.
func setNewProcessGroup(cmd *exec.Cmd) {}

func killProcessGroup(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	return cmd.Process.Kill()
}

func processAlive(pid int) bool {
	return false // unused: the test that calls this skips on windows
}
