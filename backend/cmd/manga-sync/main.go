// manga-sync processes due series/queued chapters; durable deadlines prevent
// cron, process restarts, or manual invocations from bypassing the weekly budget.
package main

import (
	"context"
	"log"
	"nihongo-master-backend/internal/config"
	"nihongo-master-backend/internal/database"
	"nihongo-master-backend/internal/manga"
	"os"
	"time"
)

func main() {
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
	base := os.Getenv("ASSET_BASE_URL")
	if base == "" {
		base = os.Getenv("COS_BUCKET_URL")
	}
	images, err := manga.NewCOSStore(os.Getenv("COS_BUCKET_URL"), base, os.Getenv("COS_SECRET_ID"), os.Getenv("COS_SECRET_KEY"))
	if err != nil {
		log.Fatal(err)
	}
	if images == nil {
		log.Fatal("manga sync requires configured COS storage")
	}
	repo, err := manga.New(db, images)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	if err = repo.Sync(ctx, manga.NewFetcher()); err != nil {
		log.Fatal(err)
	}
	log.Print("manga sync finished; existing cached chapters preserved")
}
