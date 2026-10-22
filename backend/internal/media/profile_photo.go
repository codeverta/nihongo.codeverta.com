package media

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tencentyun/cos-go-sdk-v5"
	xdraw "golang.org/x/image/draw"
	"nihongo-master-backend/internal/imageopt"
)

const (
	MaxProfilePhotoUploadBytes = 8 << 20
	maxProfilePhotoPixels      = 24_000_000
	profilePhotoMaxSide        = 512
)

var ErrInvalidProfilePhoto = errors.New("invalid profile photo")

// ProfilePhotoStorage stores user-owned, public profile photos in COS.
type ProfilePhotoStorage struct {
	client        *cos.Client
	publicBaseURL string
}

// NewCOSProfilePhotoStorage returns nil when COS credentials are not set, so
// the API can start while exposing the upload endpoint as unavailable.
func NewCOSProfilePhotoStorage(bucketURLRaw, publicBaseURLRaw, secretID, secretKey string) (*ProfilePhotoStorage, error) {
	bucketURLRaw = strings.TrimSpace(bucketURLRaw)
	publicBaseURLRaw = strings.TrimRight(strings.TrimSpace(publicBaseURLRaw), "/")
	secretID = strings.TrimSpace(secretID)
	secretKey = strings.TrimSpace(secretKey)
	if bucketURLRaw == "" || publicBaseURLRaw == "" || secretID == "" || secretKey == "" {
		return nil, nil
	}
	bucketURL, err := url.Parse(bucketURLRaw)
	if err != nil || bucketURL.Host == "" || bucketURL.Scheme != "https" {
		return nil, fmt.Errorf("COS_BUCKET_URL is invalid")
	}
	publicURL, err := url.Parse(publicBaseURLRaw)
	if err != nil || publicURL.Host == "" || publicURL.RawQuery != "" || publicURL.Fragment != "" || publicURL.Scheme != "https" {
		return nil, fmt.Errorf("profile photo public URL is invalid")
	}
	client := cos.NewClient(&cos.BaseURL{BucketURL: bucketURL}, &http.Client{
		Timeout: 30 * time.Second,
		Transport: &cos.AuthorizationTransport{
			SecretID:  secretID,
			SecretKey: secretKey,
		},
	})
	return &ProfilePhotoStorage{client: client, publicBaseURL: publicBaseURLRaw}, nil
}

// Upload decodes and bounds the image, resizes it, strips metadata, and
// re-encodes it before sending the bytes to COS.
func (s *ProfilePhotoStorage) Upload(ctx context.Context, userID string, input []byte) (string, error) {
	if s == nil || s.client == nil {
		return "", fmt.Errorf("profile photo storage is not configured")
	}
	optimized, err := OptimizeProfilePhoto(input)
	if err != nil {
		return "", err
	}
	objectKey := "profile/avatars/" + userID + "/" + uuid.NewString() + ".jpg"
	options := &cos.ObjectPutOptions{
		ACLHeaderOptions: &cos.ACLHeaderOptions{XCosACL: "public-read"},
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType:  "image/jpeg",
			CacheControl: "public, max-age=86400",
		},
	}
	if _, err := s.client.Object.Put(ctx, objectKey, bytes.NewReader(optimized), options); err != nil {
		return "", fmt.Errorf("upload profile photo to COS: %w", err)
	}
	return s.publicBaseURL + "/" + objectKey, nil
}

// Delete removes a previously uploaded photo only when its URL and object key
// belong to this user's avatar directory on the configured public origin.
func (s *ProfilePhotoStorage) Delete(ctx context.Context, userID, photoURL string) error {
	if s == nil || s.client == nil || userID == "" || photoURL == "" {
		return nil
	}
	parsed, err := url.Parse(photoURL)
	if err != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil
	}
	base, _ := url.Parse(s.publicBaseURL)
	if parsed.Scheme != base.Scheme || parsed.Host != base.Host {
		return nil
	}
	basePath := strings.TrimRight(base.Path, "/") + "/"
	if !strings.HasPrefix(parsed.Path, basePath) {
		return nil
	}
	objectKey, err := url.PathUnescape(strings.TrimPrefix(parsed.Path, basePath))
	if err != nil || !strings.HasPrefix(objectKey, "profile/avatars/"+userID+"/") || strings.Contains(objectKey, "..") {
		return nil
	}
	if _, err := s.client.Object.Delete(ctx, objectKey); err != nil {
		return fmt.Errorf("delete profile photo from COS: %w", err)
	}
	return nil
}

// OptimizeProfilePhoto accepts JPEG, PNG, and GIF, then stores an opaque JPEG
// bounded to 512 pixels. The mobile client sends JPEG on both Android and iOS.
func OptimizeProfilePhoto(input []byte) ([]byte, error) {
	if len(input) == 0 || len(input) > MaxProfilePhotoUploadBytes {
		return nil, fmt.Errorf("%w: foto harus berukuran maksimal 8 MB", ErrInvalidProfilePhoto)
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(input))
	if err != nil || (format != "jpeg" && format != "png" && format != "gif") {
		return nil, fmt.Errorf("%w: format foto tidak didukung; pilih JPEG, PNG, atau GIF", ErrInvalidProfilePhoto)
	}
	configWidth, configHeight := int64(config.Width), int64(config.Height)
	if configWidth <= 0 || configHeight <= 0 || configWidth > maxProfilePhotoPixels/configHeight {
		return nil, fmt.Errorf("%w: ukuran gambar terlalu besar", ErrInvalidProfilePhoto)
	}
	source, decodedFormat, err := image.Decode(bytes.NewReader(input))
	if err != nil || decodedFormat != format {
		return nil, fmt.Errorf("%w: foto tidak dapat dibaca", ErrInvalidProfilePhoto)
	}
	width, height := source.Bounds().Dx(), source.Bounds().Dy()
	if width > profilePhotoMaxSide || height > profilePhotoMaxSide {
		scale := min(float64(profilePhotoMaxSide)/float64(width), float64(profilePhotoMaxSide)/float64(height))
		width = max(1, int(float64(width)*scale))
		height = max(1, int(float64(height)*scale))
	}
	output := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(output, output.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	xdraw.CatmullRom.Scale(output, output.Bounds(), source, source.Bounds(), draw.Over, nil)
	result, err := imageopt.JPEG(output)
	if err != nil {
		return nil, fmt.Errorf("gagal mengompres foto")
	}
	return result, nil
}
