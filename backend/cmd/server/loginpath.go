package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"time"
)

const pathMarker = "__OPENSPEC_STUDIO_PATH__"

// loginShellPath asks the user's login shell for its PATH. Apps started
// from the macOS Dock/Finder get a minimal launchd PATH, so CLI providers
// installed under ~/.local/bin, Homebrew, or a Node version manager would
// otherwise be "not found" even though they work in the terminal.
func loginShellPath(timeout time.Duration) (string, error) {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/zsh"
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// -i as well as -l: many setups (nvm, asdf) only extend PATH in the
	// interactive rc file. Markers fence off anything the profile prints.
	cmd := exec.CommandContext(ctx, shell, "-l", "-i", "-c", `printf '`+pathMarker+`%s`+pathMarker+`' "$PATH"`)
	cmd.Stdin = nil
	cmd.WaitDelay = time.Second
	out, err := cmd.Output()
	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	if err != nil {
		return "", err
	}
	return extractMarkedPath(string(out))
}

func extractMarkedPath(out string) (string, error) {
	start := strings.Index(out, pathMarker)
	if start < 0 {
		return "", errors.New("login shell did not print PATH")
	}
	rest := out[start+len(pathMarker):]
	end := strings.Index(rest, pathMarker)
	if end < 0 {
		return "", errors.New("login shell PATH output truncated")
	}
	path := strings.TrimSpace(rest[:end])
	if path == "" {
		return "", errors.New("login shell PATH is empty")
	}
	return path, nil
}
