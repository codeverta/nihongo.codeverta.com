package ankitransfer

import (
	"bytes"
	"fmt"
	_ "golang.org/x/image/webp"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
)

// DetectMediaMIME ignores file names and user-supplied MIME declarations.
func DetectMediaMIME(data []byte) (string, error) {
	if len(data) == 0 || len(data) > 10<<20 {
		return "", fmt.Errorf("media must be between 1 byte and 10 MiB")
	}
	kind := http.DetectContentType(data)
	switch kind {
	case "image/png", "image/jpeg", "image/gif", "image/webp":
		config, _, err := image.DecodeConfig(bytes.NewReader(data))
		if err != nil || config.Width <= 0 || config.Height <= 0 || int64(config.Width)*int64(config.Height) > 20_000_000 {
			return "", fmt.Errorf("invalid image or image exceeds 20 megapixels")
		}
	case "audio/mpeg", "audio/ogg", "audio/wave", "audio/wav":
	case "application/ogg":
		kind = "audio/ogg"
	case "video/mp4":
		kind = "audio/mp4"
	default:
		return "", fmt.Errorf("unsupported image/audio content")
	}
	return kind, nil
}
