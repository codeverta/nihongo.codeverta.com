package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"nihongo-master-backend/internal/api"
	"nihongo-master-backend/internal/auth"
	"nihongo-master-backend/internal/billing"
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
	databasePath := strings.TrimSpace(os.Getenv("GAMIFICATION_DB_PATH"))
	if databasePath == "" {
		databasePath = "data/gamification.db"
	}
	if err := memoryStore.EnableGamificationPersistence(databasePath); err != nil {
		log.Fatal(err)
	}
	defer memoryStore.Close()
	if err := memoryStore.EnableBillingPersistence(); err != nil {
		log.Fatal(err)
	}
	verifier, err := billing.NewStoreVerifier(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	billingService := &billing.Service{Store: memoryStore, Verifier: verifier, ProductID: verifier.ProductID}
	assetBaseURL := strings.TrimSpace(os.Getenv("ASSET_BASE_URL"))
	if assetBaseURL == "" {
		assetBaseURL = os.Getenv("COS_BUCKET_URL")
	}
	memoryStore.SetAssetBaseURL(assetBaseURL)
	apiServer := api.NewServer(
		memoryStore,
		auth.NewGoogleVerifier(os.Getenv("GOOGLE_CLIENT_ID")),
	)
	apiServer.ConfigureBilling(billingService, os.Getenv("GOOGLE_PLAY_PUSH_AUDIENCE"), os.Getenv("GOOGLE_PLAY_PUSH_EMAIL"))
	dictionaryCache, err := memoryStore.DictionaryCache()
	if err != nil {
		log.Fatal(err)
	}
	apiServer.ConfigureDictionaryCache(dictionaryCache)
	billingContext, stopBilling := context.WithCancel(context.Background())
	defer stopBilling()
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			ctx, cancel := context.WithTimeout(billingContext, 4*time.Minute)
			if err := billingService.Reconcile(ctx); err != nil {
				log.Printf("billing reconciliation: %v", err)
			}
			cancel()
			select {
			case <-billingContext.Done():
				return
			case <-ticker.C:
			}
		}
	}()
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
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("nihongo master API listening on http://localhost:%s", port)
	shutdownSignal, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	shutdownDone := make(chan struct{})
	go func() {
		<-shutdownSignal.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("HTTP shutdown: %v", err)
		}
		close(shutdownDone)
	}()
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
	<-shutdownDone
}
