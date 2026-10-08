package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
	"nihongo-master-backend/internal/config"
)

func main() {
	if err := config.LoadEnvFile(".env", "backend/.env"); err != nil {
		log.Fatal(err)
	}
	source := flag.String("source", "assets", "local asset directory (relative to backend)")
	prefix := flag.String("prefix", "assets", "COS object key prefix")
	publicRead := flag.Bool("public-read", true, "make uploaded assets publicly readable")
	flag.Parse()

	secretID := strings.TrimSpace(os.Getenv("COS_SECRET_ID"))
	secretKey := strings.TrimSpace(os.Getenv("COS_SECRET_KEY"))
	bucketURL := strings.TrimRight(strings.TrimSpace(os.Getenv("COS_BUCKET_URL")), "/")
	if secretID == "" || secretKey == "" || bucketURL == "" {
		log.Fatal("COS_SECRET_ID, COS_SECRET_KEY, and COS_BUCKET_URL are required")
	}
	u, err := url.Parse(bucketURL)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		log.Fatal("COS_BUCKET_URL must be a valid HTTPS bucket URL")
	}
	client := cos.NewClient(
		&cos.BaseURL{BucketURL: u},
		&http.Client{
			Timeout: 2 * time.Minute,
			Transport: &cos.AuthorizationTransport{
				SecretID: secretID, SecretKey: secretKey,
			},
		},
	)

	root, err := filepath.Abs(*source)
	if err != nil {
		log.Fatal(err)
	}
	uploaded := 0
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		key := strings.Trim(strings.TrimSpace(*prefix), "/") + "/" + filepath.ToSlash(relative)
		var options *cos.ObjectPutOptions
		if *publicRead {
			options = &cos.ObjectPutOptions{
				ACLHeaderOptions: &cos.ACLHeaderOptions{XCosACL: "public-read"},
			}
		}
		if _, err := client.Object.PutFromFile(context.Background(), key, path, options); err != nil {
			return fmt.Errorf("upload %s: %w", key, err)
		}
		uploaded++
		log.Printf("uploaded %s", key)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("asset sync completed: %d files", uploaded)
}
