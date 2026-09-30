//go:build unix

package instancelock

import (
	"errors"
	"testing"
)

func TestSecondAcquireReportsHolderAddress(t *testing.T) {
	dir := t.TempDir()

	first, err := Acquire(dir)
	if err != nil {
		t.Fatalf("first acquire: %v", err)
	}
	defer first.Release()
	if err := first.SetAddr("127.0.0.1:4173"); err != nil {
		t.Fatalf("set addr: %v", err)
	}

	_, err = Acquire(dir)
	var held *HeldError
	if !errors.As(err, &held) {
		t.Fatalf("second acquire: want HeldError, got %v", err)
	}
	if held.Addr != "127.0.0.1:4173" {
		t.Fatalf("held addr = %q, want 127.0.0.1:4173", held.Addr)
	}
}

func TestAcquireAfterReleaseSucceeds(t *testing.T) {
	dir := t.TempDir()

	first, err := Acquire(dir)
	if err != nil {
		t.Fatalf("first acquire: %v", err)
	}
	_ = first.SetAddr("127.0.0.1:1")
	first.Release()

	if got := ReadAddr(dir); got != "" {
		t.Fatalf("addr after release = %q, want empty", got)
	}
	second, err := Acquire(dir)
	if err != nil {
		t.Fatalf("acquire after release: %v", err)
	}
	second.Release()
}
