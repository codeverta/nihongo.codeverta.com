package api

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"image"
	"os"
	"path/filepath"

	"nihongo-master-backend/internal/imageopt"
)

// optimizeStoredImage upgrades an existing PNG or oversized JPEG to the
// backend's bounded JPEG format and returns its content-addressed filename.
func optimizeStoredImage(directory, name string) (string, bool, error) {
	path := filepath.Join(directory, name)
	data, err := os.ReadFile(path)
	if err != nil {
		return name, false, err
	}
	if filepath.Ext(name) == ".jpg" && len(data) <= 200*1024 {
		return name, false, nil
	}
	decoded, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return name, false, err
	}
	optimized, err := imageopt.JPEG(decoded)
	if err != nil {
		return name, false, err
	}
	sum := sha256.Sum256(optimized)
	name = hex.EncodeToString(sum[:]) + ".jpg"
	target := filepath.Join(directory, name)
	if _, err := os.Stat(target); errors.Is(err, os.ErrNotExist) {
		if err = atomicBannerWrite(target, optimized, 0644); err != nil {
			if _, statErr := os.Stat(target); statErr != nil {
				return name, false, err
			}
		}
	} else if err != nil {
		return name, false, err
	}
	return name, true, nil
}
