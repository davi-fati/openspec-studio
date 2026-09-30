package repository

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/zalando/go-keyring"
)

const keyringService = "openspec-studio-ai-provider"

var ErrCredentialNotFound = errors.New("credential not found")

// CredentialStore persists AI-provider API keys in the OS keychain where
// available, falling back to an AES-GCM encrypted local file (key material
// generated once and kept 0600 in the app-data dir) when no keychain
// backend exists - e.g. a bare Linux CI/container environment. Either way,
// the plaintext key never touches SQLite or any request other than the
// provider it belongs to.
type CredentialStore struct {
	appDataDir  string
	useKeyring  bool
}

func NewCredentialStore(appDataDir string) *CredentialStore {
	useKeyring := true
	// Probe once: some environments (headless Linux, CI) have no keychain
	// backend at all, in which case every keyring call errors out.
	if err := keyring.Set(keyringService, "__probe__", "probe"); err != nil {
		useKeyring = false
	} else {
		_ = keyring.Delete(keyringService, "__probe__")
	}
	return &CredentialStore{appDataDir: appDataDir, useKeyring: useKeyring}
}

func (s *CredentialStore) key(providerID int64) string {
	return strconv.FormatInt(providerID, 10)
}

func (s *CredentialStore) Set(providerID int64, secret string) error {
	if s.useKeyring {
		return keyring.Set(keyringService, s.key(providerID), secret)
	}
	return s.setEncryptedFile(providerID, secret)
}

func (s *CredentialStore) Get(providerID int64) (string, error) {
	if s.useKeyring {
		secret, err := keyring.Get(keyringService, s.key(providerID))
		if errors.Is(err, keyring.ErrNotFound) {
			return "", ErrCredentialNotFound
		}
		return secret, err
	}
	return s.getEncryptedFile(providerID)
}

func (s *CredentialStore) Delete(providerID int64) error {
	if s.useKeyring {
		err := keyring.Delete(keyringService, s.key(providerID))
		if errors.Is(err, keyring.ErrNotFound) {
			return nil
		}
		return err
	}
	path := s.fallbackPath(providerID)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// --- encrypted-file fallback ---

func (s *CredentialStore) fallbackPath(providerID int64) string {
	return filepath.Join(s.appDataDir, "credentials", s.key(providerID)+".enc")
}

func (s *CredentialStore) fallbackKeyPath() string {
	return filepath.Join(s.appDataDir, "credentials", ".key")
}

func (s *CredentialStore) loadOrCreateFallbackKey() ([]byte, error) {
	keyPath := s.fallbackKeyPath()
	if data, err := os.ReadFile(keyPath); err == nil && len(data) == 32 {
		return data, nil
	}
	if err := os.MkdirAll(filepath.Dir(keyPath), 0o700); err != nil {
		return nil, err
	}
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	if err := os.WriteFile(keyPath, key, 0o600); err != nil {
		return nil, err
	}
	return key, nil
}

func (s *CredentialStore) setEncryptedFile(providerID int64, secret string) error {
	key, err := s.loadOrCreateFallbackKey()
	if err != nil {
		return err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(secret), nil)

	path := s.fallbackPath(providerID)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(base64.StdEncoding.EncodeToString(ciphertext)), 0o600)
}

func (s *CredentialStore) getEncryptedFile(providerID int64) (string, error) {
	path := s.fallbackPath(providerID)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrCredentialNotFound
		}
		return "", err
	}
	ciphertext, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return "", fmt.Errorf("decode credential: %w", err)
	}

	key, err := s.loadOrCreateFallbackKey()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(ciphertext) < gcm.NonceSize() {
		return "", fmt.Errorf("corrupt credential file")
	}
	nonce, ct := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt credential: %w", err)
	}
	return string(plaintext), nil
}
