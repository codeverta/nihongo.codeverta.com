package manga

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
	_ "golang.org/x/image/webp"
	"nihongo-master-backend/internal/imageopt"
)

const MaxImageBytes = 12 << 20

type ImageStore interface {
	Put(context.Context, []byte) (Asset, error)
}

type COSStore struct {
	client     *cos.Client
	publicBase string
}

func NewCOSStore(bucket, publicBase, secretID, secretKey string) (*COSStore, error) {
	if bucket == "" || publicBase == "" || secretID == "" || secretKey == "" {
		return nil, nil
	}
	u, err := url.Parse(bucket)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return nil, errors.New("invalid manga COS bucket URL")
	}
	p, err := url.Parse(publicBase)
	if err != nil || p.Scheme != "https" || p.Host == "" || p.User != nil || p.RawQuery != "" || p.Fragment != "" {
		return nil, errors.New("invalid manga COS public URL")
	}
	client := cos.NewClient(&cos.BaseURL{BucketURL: u}, &http.Client{Timeout: 60 * time.Second, Transport: &cos.AuthorizationTransport{SecretID: secretID, SecretKey: secretKey}})
	return &COSStore{client: client, publicBase: strings.TrimRight(publicBase, "/")}, nil
}

func inspectImage(body []byte) (image.Config, string, error) {
	if len(body) == 0 || len(body) > MaxImageBytes {
		return image.Config{}, "", errors.New("manga image exceeds byte limit")
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(body))
	if err != nil || (format != "jpeg" && format != "png" && format != "webp") || cfg.Width < 64 || cfg.Height < 64 || cfg.Width > 20000 || cfg.Height > 20000 || int64(cfg.Width)*int64(cfg.Height) > 60_000_000 {
		return image.Config{}, "", errors.New("invalid or oversized manga image")
	}
	return cfg, format, nil
}

func (s *COSStore) Put(ctx context.Context, body []byte) (Asset, error) {
	_, _, err := inspectImage(body)
	if err != nil {
		return Asset{}, err
	}
	decoded, _, err := image.Decode(bytes.NewReader(body))
	if err != nil {
		return Asset{}, errors.New("invalid manga image")
	}
	optimized, err := imageopt.JPEG(decoded)
	if err != nil {
		return Asset{}, fmt.Errorf("optimize manga image: %w", err)
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(optimized))
	if err != nil {
		return Asset{}, fmt.Errorf("read optimized manga image: %w", err)
	}
	h := sha256.Sum256(optimized)
	key := "manga/pages/" + hex.EncodeToString(h[:]) + ".jpg"
	_, err = s.client.Object.Put(ctx, key, bytes.NewReader(optimized), &cos.ObjectPutOptions{
		ACLHeaderOptions:       &cos.ACLHeaderOptions{XCosACL: "public-read"},
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{ContentType: "image/jpeg", CacheControl: "public, max-age=31536000, immutable"},
	})
	if err != nil {
		return Asset{}, fmt.Errorf("manga COS upload: %w", err)
	}
	return Asset{URL: s.publicBase + "/" + key, ObjectKey: key, Width: cfg.Width, Height: cfg.Height, Bytes: len(optimized), CreatedAt: time.Now().UTC()}, nil
}
