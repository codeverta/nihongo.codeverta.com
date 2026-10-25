package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"nihongo-master-backend/internal/billing"
)

func main() {
	backendRoot, err := findBackendRoot()
	if err != nil {
		fatal(err)
	}
	secretsDir := filepath.Join(backendRoot, "secrets")
	if err := os.MkdirAll(secretsDir, 0700); err != nil {
		fatal(fmt.Errorf("create private secrets directory: %w", err))
	}
	privateKeyPath := filepath.Join(secretsDir, "offline-access-private-key")
	privateKey, created, err := loadOrCreatePrivateKey(privateKeyPath)
	if err != nil {
		fatal(err)
	}
	_, practiceCreated, err := loadOrCreateEncodedSecret(
		filepath.Join(secretsDir, "offline-practice-signing-key"),
		32,
	)
	if err != nil {
		fatal(err)
	}
	signer, err := billing.NewOfflineLeaseSigner(privateKey)
	if err != nil {
		fatal(fmt.Errorf("validate offline signing key: %w", err))
	}
	if created {
		fmt.Fprintln(os.Stderr, "Created a persistent Ed25519 signing key in the ignored backend/secrets directory.")
	} else {
		fmt.Fprintln(os.Stderr, "Reusing the existing Ed25519 signing key; it was not rotated.")
	}
	if practiceCreated {
		fmt.Fprintln(os.Stderr, "Created a distinct offline-practice HMAC secret in the ignored backend/secrets directory.")
	} else {
		fmt.Fprintln(os.Stderr, "Reusing the existing offline-practice HMAC secret; it was not rotated.")
	}
	fmt.Printf("OFFLINE_ACCESS_PUBLIC_KEY=%s\n", signer.PublicKeyBase64())
	fmt.Println()
	fmt.Println("Set these backend/.env entries to enable server signing:")
	fmt.Println("OFFLINE_ACCESS_PRIVATE_KEY_FILE=secrets/offline-access-private-key")
	fmt.Println("OFFLINE_PRACTICE_SIGNING_KEY_FILE=secrets/offline-practice-signing-key")
}

func findBackendRoot() (string, error) {
	current, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("find backend directory: %w", err)
	}
	for {
		if info, err := os.Stat(filepath.Join(current, "go.mod")); err == nil && !info.IsDir() {
			if info, err := os.Stat(filepath.Join(current, "internal", "billing")); err == nil && info.IsDir() {
				return current, nil
			}
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", errors.New("run this command inside the backend workspace")
		}
		current = parent
	}
}

func loadOrCreatePrivateKey(path string) ([]byte, bool, error) {
	if existing, err := os.ReadFile(path); err == nil {
		if err := restrictPrivateFile(path); err != nil {
			return nil, false, err
		}
		key, err := decodePrivateKey(existing)
		if err != nil {
			return nil, false, fmt.Errorf("existing offline access key is invalid: %w", err)
		}
		return key, false, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, false, fmt.Errorf("read existing offline access key: %w", err)
	}

	seed := make([]byte, ed25519.SeedSize)
	if _, err := rand.Read(seed); err != nil {
		return nil, false, fmt.Errorf("generate offline access key: %w", err)
	}
	contents := []byte(base64.StdEncoding.EncodeToString(seed) + "\n")
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if errors.Is(err, os.ErrExist) {
		existing, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil, false, fmt.Errorf("read concurrently created offline access key: %w", readErr)
		}
		if err := restrictPrivateFile(path); err != nil {
			return nil, false, err
		}
		key, decodeErr := decodePrivateKey(existing)
		if decodeErr != nil {
			return nil, false, fmt.Errorf("concurrently created offline access key is invalid: %w", decodeErr)
		}
		return key, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("save offline access key: %w", err)
	}
	if _, err := file.Write(contents); err != nil {
		file.Close()
		os.Remove(path)
		return nil, false, fmt.Errorf("write offline access key: %w", err)
	}
	if err := file.Close(); err != nil {
		os.Remove(path)
		return nil, false, fmt.Errorf("close offline access key: %w", err)
	}
	if err := os.Chmod(path, 0600); err != nil {
		return nil, false, fmt.Errorf("restrict offline access key permissions: %w", err)
	}
	return seed, true, nil
}

func loadOrCreateEncodedSecret(path string, size int) ([]byte, bool, error) {
	decode := func(contents []byte) ([]byte, error) {
		encoded := strings.TrimSpace(string(contents))
		key, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			key, err = base64.RawStdEncoding.DecodeString(encoded)
		}
		if err != nil {
			return nil, err
		}
		if len(key) != size {
			return nil, fmt.Errorf("expected a %d-byte secret", size)
		}
		return key, nil
	}
	readExisting := func() ([]byte, error) {
		contents, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		return decode(contents)
	}
	if _, err := os.Stat(path); err == nil {
		if err := restrictPrivateFile(path); err != nil {
			return nil, false, err
		}
		key, err := readExisting()
		if err != nil {
			return nil, false, fmt.Errorf("existing secret is invalid: %w", err)
		}
		return key, false, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, false, fmt.Errorf("inspect existing secret: %w", err)
	}

	secret := make([]byte, size)
	if _, err := rand.Read(secret); err != nil {
		return nil, false, fmt.Errorf("generate secret: %w", err)
	}
	contents := []byte(base64.StdEncoding.EncodeToString(secret) + "\n")
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if errors.Is(err, os.ErrExist) {
		key, readErr := readExisting()
		if readErr != nil {
			return nil, false, fmt.Errorf("read concurrently created secret: %w", readErr)
		}
		if err := restrictPrivateFile(path); err != nil {
			return nil, false, err
		}
		return key, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("save secret: %w", err)
	}
	if _, err := file.Write(contents); err != nil {
		file.Close()
		os.Remove(path)
		return nil, false, fmt.Errorf("write secret: %w", err)
	}
	if err := file.Close(); err != nil {
		os.Remove(path)
		return nil, false, fmt.Errorf("close secret: %w", err)
	}
	if err := os.Chmod(path, 0600); err != nil {
		return nil, false, fmt.Errorf("restrict secret permissions: %w", err)
	}
	return secret, true, nil
}

func restrictPrivateFile(path string) error {
	if err := os.Chmod(path, 0600); err != nil {
		return fmt.Errorf("restrict private file permissions: %w", err)
	}
	return nil
}

func decodePrivateKey(contents []byte) ([]byte, error) {
	encoded := strings.TrimSpace(string(contents))
	key, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		key, err = base64.RawStdEncoding.DecodeString(encoded)
	}
	if err != nil {
		return nil, err
	}
	if len(key) != ed25519.SeedSize && len(key) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("expected a %d-byte seed or %d-byte private key", ed25519.SeedSize, ed25519.PrivateKeySize)
	}
	if len(key) == ed25519.PrivateKeySize {
		if _, err := billing.NewOfflineLeaseSigner(key); err != nil {
			return nil, err
		}
	}
	return key, nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
