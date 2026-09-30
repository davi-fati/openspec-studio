package main

import "testing"

func TestExtractMarkedPathIgnoresProfileNoise(t *testing.T) {
	out := "Welcome!\nnvm: using node 22\n" + pathMarker + "/Users/me/.local/bin:/usr/bin" + pathMarker + "\n"
	got, err := extractMarkedPath(out)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if got != "/Users/me/.local/bin:/usr/bin" {
		t.Fatalf("got %q", got)
	}
}

func TestExtractMarkedPathMissing(t *testing.T) {
	for _, out := range []string{"", "noise only", pathMarker + "/usr/bin", pathMarker + pathMarker} {
		if _, err := extractMarkedPath(out); err == nil {
			t.Fatalf("extract(%q): want error", out)
		}
	}
}
