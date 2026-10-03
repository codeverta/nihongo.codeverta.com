package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"nihongo-master-backend/internal/api"
	"nihongo-master-backend/internal/auth"
	"nihongo-master-backend/internal/config"
	"nihongo-master-backend/internal/store"
)

func main() {
	if err := config.LoadEnvFile(".env", "backend/.env"); err != nil {
		log.Fatal(err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	memoryStore := store.NewMemoryStore()
	memoryStore.SetAssetBaseURL(os.Getenv("COS_BUCKET_URL"))
	apiServer := api.NewServer(
		memoryStore,
		auth.NewGoogleVerifier(os.Getenv("GOOGLE_CLIENT_ID")),
	)
	brandingPath := os.Getenv("BRANDING_CONFIG_PATH")
	if brandingPath == "" {
		brandingPath = "data/branding.json"
	}
	if err := apiServer.ConfigureBranding(brandingPath, os.Getenv("BRANDING_ADMIN_TOKEN")); err != nil {
		log.Fatal(err)
	}
	handler := apiServer.Handler()
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("nihongo master API listening on http://localhost:%s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
