package places

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// LocalImages stores photos on disk when COS is not configured; the API serves
// them from PublicPath (for example "/map-images/").
type LocalImages struct {
	Directory  string
	PublicPath string
}

var localImageName = regexp.MustCompile(`^[0-9a-f]{64}\.jpg$`)

func (l LocalImages) Put(_ context.Context, jpeg []byte) (string, error) {
	if err := os.MkdirAll(l.Directory, 0o755); err != nil {
		return "", fmt.Errorf("create map image directory: %w", err)
	}
	sum := sha256.Sum256(jpeg)
	name := hex.EncodeToString(sum[:]) + ".jpg"
	path := filepath.Join(l.Directory, name)
	if _, err := os.Stat(path); err == nil {
		return l.PublicPath + name, nil
	}
	temp, err := os.CreateTemp(l.Directory, ".upload-*")
	if err != nil {
		return "", err
	}
	defer os.Remove(temp.Name())
	if _, err := temp.Write(jpeg); err != nil {
		temp.Close()
		return "", err
	}
	if err := temp.Close(); err != nil {
		return "", err
	}
	if err := os.Rename(temp.Name(), path); err != nil {
		return "", err
	}
	return l.PublicPath + name, nil
}

// Open returns the stored file for a validated name.
func (l LocalImages) Open(name string) (*os.File, error) {
	if !localImageName.MatchString(name) {
		return nil, os.ErrNotExist
	}
	return os.Open(filepath.Join(l.Directory, name))
}
