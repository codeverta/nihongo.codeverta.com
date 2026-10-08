package main

import (
	"log"
	"net/http"
	"os"
	"strings"
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
	assetBaseURL := strings.TrimSpace(os.Getenv("ASSET_BASE_URL"))
	if assetBaseURL == "" {
		assetBaseURL = os.Getenv("COS_BUCKET_URL")
	}
	memoryStore.SetAssetBaseURL(assetBaseURL)
	apiServer := api.NewServer(
		memoryStore,
		auth.NewGoogleVerifier(os.Getenv("GOOGLE_CLIENT_ID")),
	)
	assetDirectory := os.Getenv("ASSET_DIRECTORY")
	if assetDirectory == "" {
		assetDirectory = "assets"
		if _, err := os.Stat("backend/assets"); err == nil {
			assetDirectory = "backend/assets"
		}
	}
	if err := apiServer.ConfigureAssets(assetDirectory); err != nil {
		if assetBaseURL == "" {
			log.Fatalf("local learning assets are required without ASSET_BASE_URL/COS_BUCKET_URL: %v", err)
		}
		log.Printf("local assets unavailable; using remote asset base URL")
	}
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
