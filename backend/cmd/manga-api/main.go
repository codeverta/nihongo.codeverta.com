// manga-api is an optional standalone Manga service. It shares the same tables,
// scheduler leases, and handlers as cmd/api, without starting an account writer.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"nihongo-master-backend/internal/api"
	"nihongo-master-backend/internal/config"
	"nihongo-master-backend/internal/database"
	"nihongo-master-backend/internal/manga"
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
	cos, err := manga.NewCOSStore(os.Getenv("COS_BUCKET_URL"), base, os.Getenv("COS_SECRET_ID"), os.Getenv("COS_SECRET_KEY"))
	if err != nil {
		log.Fatal(err)
	}
	var storage manga.ImageStore
	if cos != nil {
		storage = cos
	}
	repo, err := manga.New(db, storage)
	if err != nil {
		log.Fatal(err)
	}
	token := os.Getenv("MANGA_ADMIN_TOKEN")
	if token == "" {
		token = os.Getenv("BRANDING_ADMIN_TOKEN")
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if storage != nil && os.Getenv("MANGA_SYNC_ENABLED") != "false" {
		go repo.Run(ctx, manga.NewFetcher())
	}
	port := os.Getenv("MANGA_PORT")
	if port == "" {
		port = "8081"
	}
	authBase := os.Getenv("MANGA_AUTH_API_BASE_URL")
	if authBase == "" {
		authBase = "http://127.0.0.1:8080/v1"
	}
	server := &http.Server{Addr: ":" + port, Handler: api.NewMangaHandler(repo, token, api.NewMangaBillingChecker(authBase)), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	go func() {
		<-ctx.Done()
		work, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(work)
	}()
	log.Printf("manga API listening on http://localhost:%s", port)
	if err = server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
