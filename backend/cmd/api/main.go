package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"nihongo-master-backend/internal/api"
	"nihongo-master-backend/internal/auth"
	"nihongo-master-backend/internal/billing"
	"nihongo-master-backend/internal/config"
	"nihongo-master-backend/internal/diagnostics"
	"nihongo-master-backend/internal/jobfetch"
	"nihongo-master-backend/internal/jobs"
	"nihongo-master-backend/internal/manga"
	"nihongo-master-backend/internal/media"
	"nihongo-master-backend/internal/news"
	"nihongo-master-backend/internal/places"
	"nihongo-master-backend/internal/notifications"
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
	dsn, err := config.MariaDBDSN()
	if err != nil {
		log.Fatal(err)
	}
	if err := memoryStore.EnableGamificationPersistence(dsn); err != nil {
		log.Fatal(err)
	}
	defer memoryStore.Close()
	if err := memoryStore.EnableBillingPersistence(); err != nil {
		log.Fatal(err)
	}
	// Keep Play reviewer credentials in the server environment, outside the repository.
	if err := memoryStore.ConfigurePlayReviewAccount(os.Getenv("PLAY_REVIEW_EMAIL"), os.Getenv("PLAY_REVIEW_PASSWORD")); err != nil {
		log.Fatalf("configure Google Play review account: %v", err)
	}
	verifier, err := billing.NewStoreVerifier(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	offlineSigner, err := billing.NewOfflineLeaseSignerFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	billingService := &billing.Service{Store: memoryStore, Verifier: verifier, ProductID: verifier.ProductID, OfflineSigner: offlineSigner}
	if offlineSigner == nil {
		log.Print("offline premium access disabled: OFFLINE_ACCESS_PRIVATE_KEY is not configured")
	} else {
		log.Printf("offline premium access signing enabled (public key for mobile build: %s)", offlineSigner.PublicKeyBase64())
	}
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
	dictionaryToken := os.Getenv("DICTIONARY_ADMIN_TOKEN")
	if dictionaryToken == "" {
		dictionaryToken = os.Getenv("BRANDING_ADMIN_TOKEN")
	}
	apiServer.ConfigureDictionaryAdmin(dictionaryToken)
	jobDB, err := memoryStore.JobsDatabase()
	if err != nil {
		log.Fatal(err)
	}
	diagnosticRepository, err := diagnostics.New(jobDB)
	if err != nil {
		log.Fatalf("configure error diagnostics: %v", err)
	}
	diagnosticToken := os.Getenv("DIAGNOSTICS_ADMIN_TOKEN")
	if diagnosticToken == "" {
		diagnosticToken = os.Getenv("BRANDING_ADMIN_TOKEN")
	}
	apiServer.ConfigureDiagnostics(diagnosticRepository, diagnosticToken)
	jobInterval, err := jobs.Interval(os.Getenv("JOBS_SYNC_INTERVAL_HOURS"))
	if err != nil {
		log.Fatal(err)
	}
	jobRepository, err := jobs.New(jobDB, jobInterval)
	if err != nil {
		log.Fatalf("configure Japan jobs: %v", err)
	}
	jobRepository.SetErrorHandler(func(err error) { apiServer.ReportBackendWorkerError("jobs-sync", err) })
	apiServer.ConfigureJobs(jobRepository)
	jobsContext, stopJobs := context.WithCancel(context.Background())
	defer stopJobs()
	if os.Getenv("JOBS_SYNC_ENABLED") != "false" {
		go jobRepository.Run(jobsContext, jobfetch.NewClient())
	}
	billingContext, stopBilling := context.WithCancel(context.Background())
	defer stopBilling()
	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				apiServer.ReportBackendWorkerError("billing-reconciliation", fmt.Errorf("panic: %v\n%s", recovered, debug.Stack()))
			}
		}()
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			ctx, cancel := context.WithTimeout(billingContext, 4*time.Minute)
			if err := billingService.Reconcile(ctx); err != nil {
				log.Printf("billing reconciliation: %v", err)
				apiServer.ReportBackendWorkerError("billing-reconciliation", err)
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
	newsRepository, err := news.New(jobDB)
	if err != nil {
		log.Fatalf("configure news: %v", err)
	}
	newsFetcher := news.NewFetcher(filepath.Join(filepath.Dir(brandingPath), "news-images"))
	newsImages, err := media.NewCOSNewsImageStorage(os.Getenv("COS_BUCKET_URL"), assetBaseURL, os.Getenv("COS_SECRET_ID"), os.Getenv("COS_SECRET_KEY"))
	if err != nil {
		log.Fatalf("configure news image COS: %v", err)
	}
	if newsImages != nil {
		newsFetcher.SetImageStore(newsImages)
		apiServer.ConfigureNewsImageStore(newsImages)
	} else {
		log.Printf("news images stay on local disk: COS is not configured")
	}
	newsFetcher.SetImageLookup(newsRepository.HasImage)
	newsToken := os.Getenv("NEWS_ADMIN_TOKEN")
	if newsToken == "" {
		newsToken = os.Getenv("BRANDING_ADMIN_TOKEN")
	}
	if newsToken == "" {
		newsToken = os.Getenv("DICTIONARY_ADMIN_TOKEN")
	}
	if err := apiServer.ConfigureNews(newsRepository, newsFetcher, newsToken, filepath.Join(filepath.Dir(brandingPath), "news-images"), jobsContext); err != nil {
		log.Fatal(err)
	}
	newsRepository.SetErrorHandler(func(err error) { apiServer.ReportBackendWorkerError("news-sync", err) })
	if os.Getenv("NEWS_SYNC_ENABLED") != "false" {
		go newsRepository.Run(jobsContext, newsFetcher)
	}
	mapRepository, err := places.New(jobDB)
	if err != nil {
		log.Fatalf("configure map spots: %v", err)
	}
	mapService := &places.Service{
		Store:   mapRepository,
		Fetcher: places.NewFetcher(os.Getenv("MAP_SOURCE_USER_AGENT")),
		Reading: news.Reading,
	}
	var mapLocalImages *places.LocalImages
	if newsImages != nil {
		mapService.Images = newsImages.WithPrefix("map/images/")
	} else {
		mapLocalImages = &places.LocalImages{Directory: filepath.Join(filepath.Dir(brandingPath), "map-images"), PublicPath: "/map-images/"}
		mapService.Images = mapLocalImages
	}
	mapToken := os.Getenv("MAP_ADMIN_TOKEN")
	if mapToken == "" {
		mapToken = os.Getenv("BRANDING_ADMIN_TOKEN")
	}
	apiServer.ConfigureMap(mapService, mapToken, mapLocalImages)
	mangaImages, err := manga.NewCOSStore(os.Getenv("COS_BUCKET_URL"), assetBaseURL, os.Getenv("COS_SECRET_ID"), os.Getenv("COS_SECRET_KEY"))
	if err != nil {
		log.Fatalf("configure manga COS: %v", err)
	}
	var mangaStorage manga.ImageStore
	if mangaImages != nil {
		mangaStorage = mangaImages
	}
	mangaRepository, err := manga.New(jobDB, mangaStorage)
	if err != nil {
		log.Fatalf("configure manga: %v", err)
	}
	mangaToken := os.Getenv("MANGA_ADMIN_TOKEN")
	if mangaToken == "" {
		mangaToken = os.Getenv("BRANDING_ADMIN_TOKEN")
	}
	apiServer.ConfigureManga(mangaRepository, mangaToken)
	mangaRepository.SetErrorHandler(func(err error) { apiServer.ReportBackendWorkerError("manga-sync", err) })
	if os.Getenv("MANGA_SYNC_ENABLED") != "false" && mangaStorage != nil {
		go mangaRepository.Run(jobsContext, manga.NewFetcher())
	} else if mangaStorage == nil {
		log.Print("manga import disabled: COS storage is not configured")
	}
	bannerPath := os.Getenv("HOME_BANNERS_CONFIG_PATH")
	if bannerPath == "" {
		bannerPath = "data/home-banners.json"
	}
	bannerToken := os.Getenv("HOME_BANNERS_ADMIN_TOKEN")
	if bannerToken == "" {
		bannerToken = os.Getenv("BRANDING_ADMIN_TOKEN")
	}
	if err := apiServer.ConfigureHomeBanners(bannerPath, bannerToken); err != nil {
		log.Fatalf("configure home banners: %v", err)
	}
	pricingPath := os.Getenv("PRICING_CONFIG_PATH")
	if pricingPath == "" {
		pricingPath = "data/pricing.json"
	}
	pricingToken := os.Getenv("PRICING_ADMIN_TOKEN")
	if pricingToken == "" {
		pricingToken = os.Getenv("BRANDING_ADMIN_TOKEN")
	}
	if err := apiServer.ConfigurePricing(pricingPath, pricingToken); err != nil {
		log.Fatalf("configure app pricing: %v", err)
	}
	updatePath := os.Getenv("APP_UPDATE_CONFIG_PATH")
	if updatePath == "" {
		updatePath = "data/app-updates.json"
	}
	updateToken := os.Getenv("APP_UPDATE_ADMIN_TOKEN")
	if updateToken == "" {
		updateToken = os.Getenv("BRANDING_ADMIN_TOKEN")
	}
	if err := apiServer.ConfigureAppUpdates(updatePath, updateToken); err != nil {
		log.Fatalf("configure app update policy: %v", err)
	}
	adsPath := os.Getenv("ADS_CONFIG_PATH")
	if adsPath == "" {
		adsPath = "data/ads.json"
	}
	adsToken := os.Getenv("ADS_ADMIN_TOKEN")
	if adsToken == "" {
		adsToken = os.Getenv("BRANDING_ADMIN_TOKEN")
	}
	if err := apiServer.ConfigureAds(adsPath, adsToken); err != nil {
		log.Fatalf("configure ads: %v", err)
	}
	notificationPath := os.Getenv("NOTIFICATIONS_CONFIG_PATH")
	if notificationPath == "" {
		notificationPath = "data/notifications.json"
	}
	notificationToken := os.Getenv("NOTIFICATIONS_ADMIN_TOKEN")
	if notificationToken == "" {
		notificationToken = os.Getenv("BRANDING_ADMIN_TOKEN")
	}
	fcm, err := notifications.NewFCM(context.Background(), os.Getenv("FCM_PROJECT_ID"), os.Getenv("FCM_CREDENTIALS_FILE"))
	if err != nil {
		log.Fatalf("configure Firebase messaging: %v", err)
	}
	var notificationSender notifications.Sender
	if fcm != nil {
		notificationSender = fcm
	}
	if err := apiServer.ConfigureNotifications(notificationPath, notificationToken, notificationSender); err != nil {
		log.Fatalf("configure notifications: %v", err)
	}
	profilePhotos, err := media.NewCOSProfilePhotoStorage(
		os.Getenv("COS_BUCKET_URL"), assetBaseURL,
		os.Getenv("COS_SECRET_ID"), os.Getenv("COS_SECRET_KEY"),
	)
	if err != nil {
		log.Printf("profile photo uploads disabled: %v", err)
	} else if profilePhotos == nil {
		log.Printf("profile photo uploads disabled: COS_BUCKET_URL, public asset URL, and COS credentials are required")
	}
	apiServer.ConfigureProfilePhotoStorage(profilePhotos)
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
		stopJobs()
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
