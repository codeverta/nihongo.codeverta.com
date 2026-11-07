// news-sync runs due imports once; ordinary application reads never run this path.
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
	force := flag.Bool("force", false, "Refresh enabled sources now, including article images")
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
	repo, err := news.New(db)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 18*time.Minute)
	defer cancel()
	sources, err := repo.Sources(ctx)
	if err != nil {
		log.Fatal(err)
	}
	failed := false
	brandingPath := os.Getenv("BRANDING_CONFIG_PATH")
	if brandingPath == "" {
		brandingPath = "data/branding.json"
	}
	fetcher := news.NewFetcher(filepath.Join(filepath.Dir(brandingPath), "news-images"))
	assetBaseURL := strings.TrimSpace(os.Getenv("ASSET_BASE_URL"))
	if assetBaseURL == "" {
		assetBaseURL = os.Getenv("COS_BUCKET_URL")
	}
	images, err := media.NewCOSNewsImageStorage(os.Getenv("COS_BUCKET_URL"), assetBaseURL, os.Getenv("COS_SECRET_ID"), os.Getenv("COS_SECRET_KEY"))
	if err != nil {
		log.Fatal(err)
	}
	if images != nil {
		fetcher.SetImageStore(images)
	} else {
		log.Print("COS is not configured; news images stay on local disk")
	}
	fetcher.SetImageLookup(repo.HasImage)
	for _, source := range sources {
		if !source.Enabled || (!*force && source.NextFetchAt.After(time.Now())) {
			continue
		}
		count, err := repo.Sync(ctx, fetcher, source.ID, *force)
		if err != nil {
			log.Printf("%s: %v", source.ID, err)
			failed = true
		} else {
			log.Printf("%s: processed %d articles", source.ID, count)
		}
	}
	if failed {
		log.Fatal("one or more imports failed; previous stored articles preserved")
	}
}
