// news-image-migrate uploads news images still stored on local disk
// (/news-images/<hash>.jpg) to COS and rewrites news_articles.image_url.
// It is safe to run repeatedly; already migrated rows are skipped.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"nihongo-master-backend/internal/config"
	"nihongo-master-backend/internal/database"
	"nihongo-master-backend/internal/media"
	"nihongo-master-backend/internal/news"
)

func main() {
	dryRun := flag.Bool("dry-run", false, "List what would be migrated without uploading")
	flag.Parse()
	if err := config.LoadEnvFile(".env", "backend/.env"); err != nil {
		log.Fatal(err)
	}
	dsn, err := config.MariaDBDSN()
	if err != nil {
		log.Fatal(err)
	}
	db, err := database.Open(dsn)
	if err != nil {
		log.Fatal(err)
	}
	conn, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	assetBaseURL := strings.TrimSpace(os.Getenv("ASSET_BASE_URL"))
	if assetBaseURL == "" {
		assetBaseURL = os.Getenv("COS_BUCKET_URL")
	}
	store, err := media.NewCOSNewsImageStorage(os.Getenv("COS_BUCKET_URL"), assetBaseURL, os.Getenv("COS_SECRET_ID"), os.Getenv("COS_SECRET_KEY"))
	if err != nil {
		log.Fatal(err)
	}
	if store == nil && !*dryRun {
		log.Fatal("COS_BUCKET_URL, COS_SECRET_ID and COS_SECRET_KEY are required")
	}
	brandingPath := os.Getenv("BRANDING_CONFIG_PATH")
	if brandingPath == "" {
		brandingPath = "data/branding.json"
	}
	directory := filepath.Join(filepath.Dir(brandingPath), "news-images")
	var articles []news.Article
	if err := db.Where("image_url LIKE ?", "/news-images/%").Find(&articles).Error; err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	migrated, failed := 0, 0
	for _, a := range articles {
		name := filepath.Base(a.ImageURL)
		if *dryRun {
			log.Printf("would migrate %s (%s)", a.ImageURL, a.URL)
			continue
		}
		data, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil {
			log.Printf("skip %s: %v", a.URL, err)
			failed++
			continue
		}
		url, err := store.Put(ctx, data)
		if err != nil {
			log.Printf("upload %s: %v", a.URL, err)
			failed++
			continue
		}
		if err := db.Model(&news.Article{}).Where("id = ?", a.ID).Update("image_url", url).Error; err != nil {
			log.Printf("update %s: %v", a.URL, err)
			failed++
			continue
		}
		migrated++
	}
	log.Printf("news images: %d candidates, %d migrated, %d failed", len(articles), migrated, failed)
	if failed > 0 {
		os.Exit(1)
	}
}
