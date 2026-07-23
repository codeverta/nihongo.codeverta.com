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

	handler := api.NewServer(
		store.NewMemoryStore(),
		auth.NewGoogleVerifier(os.Getenv("GOOGLE_CLIENT_ID")),
	).Handler()
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
