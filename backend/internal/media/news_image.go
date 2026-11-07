package media

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
)

const newsImagePrefix = "news/images/"

// NewsImageStorage stores optimized news article images in COS under
// content-addressed keys, so repeated imports of the same image are idempotent.
type NewsImageStorage struct {
	client        *cos.Client
	publicBaseURL string
	prefix        string // object key prefix; empty means newsImagePrefix
}

// WithPrefix returns storage that writes under another key prefix (for
// example "map/images/") with the same bucket and credentials.
func (s *NewsImageStorage) WithPrefix(prefix string) *NewsImageStorage {
	if s == nil {
		return nil
	}
	copy := *s
	copy.prefix = prefix
	return &copy
}

func (s *NewsImageStorage) keyPrefix() string {
	if s.prefix != "" {
		return s.prefix
	}
	return newsImagePrefix
}

// NewCOSNewsImageStorage returns nil when COS is not configured.
func NewCOSNewsImageStorage(bucketURLRaw, publicBaseURLRaw, secretID, secretKey string) (*NewsImageStorage, error) {
	bucketURLRaw = strings.TrimSpace(bucketURLRaw)
	publicBaseURLRaw = strings.TrimRight(strings.TrimSpace(publicBaseURLRaw), "/")
	secretID, secretKey = strings.TrimSpace(secretID), strings.TrimSpace(secretKey)
	if bucketURLRaw == "" || publicBaseURLRaw == "" || secretID == "" || secretKey == "" {
		return nil, nil
	}
	bucketURL, err := url.Parse(bucketURLRaw)
	if err != nil || bucketURL.Host == "" || bucketURL.Scheme != "https" {
		return nil, fmt.Errorf("COS_BUCKET_URL is invalid")
	}
	publicURL, err := url.Parse(publicBaseURLRaw)
	if err != nil || publicURL.Host == "" || publicURL.Scheme != "https" || publicURL.RawQuery != "" || publicURL.Fragment != "" {
		return nil, fmt.Errorf("news image public URL is invalid")
	}
	client := cos.NewClient(&cos.BaseURL{BucketURL: bucketURL}, &http.Client{
		Timeout:   30 * time.Second,
		Transport: &cos.AuthorizationTransport{SecretID: secretID, SecretKey: secretKey},
	})
	return &NewsImageStorage{client: client, publicBaseURL: publicBaseURLRaw}, nil
}

func (s *NewsImageStorage) Put(ctx context.Context, jpeg []byte) (string, error) {
	if s == nil || s.client == nil {
		return "", fmt.Errorf("news image storage is not configured")
	}
	sum := sha256.Sum256(jpeg)
	key := s.keyPrefix() + hex.EncodeToString(sum[:]) + ".jpg"
	options := &cos.ObjectPutOptions{
		ACLHeaderOptions: &cos.ACLHeaderOptions{XCosACL: "public-read"},
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType:  "image/jpeg",
			CacheControl: "public, max-age=31536000, immutable",
		},
	}
	if _, err := s.client.Object.Put(ctx, key, bytes.NewReader(jpeg), options); err != nil {
		return "", fmt.Errorf("upload image to COS: %w", err)
	}
	return s.publicBaseURL + "/" + key, nil
}

func (s *NewsImageStorage) Owns(imageURL string) bool {
	return s != nil && strings.HasPrefix(imageURL, s.publicBaseURL+"/"+s.keyPrefix())
}
