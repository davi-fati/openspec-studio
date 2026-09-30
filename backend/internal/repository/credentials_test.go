package repository

import (
	"testing"
)

// TestCredentialStoreEncryptedFileFallback exercises the encrypted-file path
// directly (bypassing the keyring probe), since CI/sandboxed environments
// usually have no OS keychain backend available.
func TestCredentialStoreEncryptedFileFallback(t *testing.T) {
	store := &CredentialStore{appDataDir: t.TempDir(), useKeyring: false}

	if err := store.Set(1, "sk-secret-for-provider-1"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := store.Set(2, "sk-secret-for-provider-2"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	got, err := store.Get(1)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != "sk-secret-for-provider-1" {
		t.Fatalf("Get returned %q, want the provider-1 secret", got)
	}

	// A different provider's key must only ever return its own secret.
	got2, err := store.Get(2)
	if err != nil {
		t.Fatalf("Get (provider 2): %v", err)
	}
	if got2 != "sk-secret-for-provider-2" {
		t.Fatalf("Get (provider 2) returned %q, want its own secret", got2)
	}

	if err := store.Delete(1); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := store.Get(1); err != ErrCredentialNotFound {
		t.Fatalf("expected ErrCredentialNotFound after delete, got %v", err)
	}
}
