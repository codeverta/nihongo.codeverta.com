// jobs-sync performs a single due-only pass; it cannot bypass weekly deadlines.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"nihongo-master-backend/internal/config"
	"nihongo-master-backend/internal/database"
	"nihongo-master-backend/internal/jobfetch"
	"nihongo-master-backend/internal/jobs"
	"os"
	"os/signal"
	"time"
)

type snapshotSource struct {
	Source    jobfetch.Source `json:"source"`
	FetchedAt time.Time       `json:"fetchedAt"`
	Result    jobfetch.Result `json:"result"`
}
type snapshotFetcher struct {
	Sources []snapshotSource `json:"sources"`
}

func (s snapshotFetcher) Fetch(ctx context.Context, id string, v jobfetch.Validators) (jobfetch.Result, error) {
	for _, entry := range s.Sources {
		if entry.Source.ID == id {
			return entry.Result, nil
		}
	}
	return jobfetch.Result{}, fmt.Errorf("snapshot missing configured source")
}
func (s snapshotFetcher) ObservationTime(id string) time.Time {
	for _, entry := range s.Sources {
		if entry.Source.ID == id {
			return entry.FetchedAt
		}
	}
	return time.Time{}
}
func main() {
	snapshot := flag.String("snapshot", "", "Reuse a recent collector snapshot without re-requesting websites")
	status := flag.Bool("status", false, "Show stored catalog size and source deadlines without fetching websites")
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
		log.Fatal("connect jobs MariaDB failed")
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()
	interval, err := jobs.Interval(os.Getenv("JOBS_SYNC_INTERVAL_HOURS"))
	if err != nil {
		log.Fatal(err)
	}
	repository, err := jobs.New(db, interval)
	if err != nil {
		log.Fatal(err)
	}
	if *status {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		page, err := repository.List(ctx, jobs.Filter{Limit: 1})
		if err != nil {
			log.Fatal("read stored jobs failed")
		}
		sources, err := repository.Sources(ctx)
		if err != nil {
			log.Fatal("read sources failed")
		}
		counts := map[string]int64{}
		for _, option := range jobs.Categories {
			p, err := repository.List(ctx, jobs.Filter{Limit: 1, Category: option.ID})
			if err != nil {
				log.Fatal("read category counts failed")
			}
			counts[option.ID] = p.Total
		}
		json.NewEncoder(os.Stdout).Encode(map[string]any{"activeJobs": page.Total, "categories": counts, "sources": sources})
		return
	}
	var fetcher jobs.Fetcher = jobfetch.NewClient()
	var captured []snapshotSource
	if *snapshot != "" {
		f, err := os.Open(*snapshot)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		var snap snapshotFetcher
		decoder := json.NewDecoder(f)
		if err = decoder.Decode(&snap); err != nil {
			log.Fatal(err)
		}
		seen := map[string]bool{}
		for _, entry := range snap.Sources {
			if seen[entry.Source.ID] || entry.FetchedAt.Before(time.Now().Add(-24*time.Hour)) || entry.FetchedAt.After(time.Now().Add(5*time.Minute)) {
				log.Fatal("snapshot requires unique actual collector results from the last 24 hours")
			}
			seen[entry.Source.ID] = true
		}
		fetcher = snap
		captured = snap.Sources
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	var result jobs.SyncReport
	if *snapshot != "" {
		result.Errors = []string{}
		for _, entry := range captured {
			var imported int
			imported, err = repository.ImportCaptured(ctx, entry.Source.ID, entry.Result, entry.FetchedAt)
			if err != nil {
				break
			}
			result.JobsSeen += imported
		}
	} else {
		result, err = repository.SyncDue(ctx, fetcher)
	}
	if err != nil {
		log.Fatal(err)
	}
	json.NewEncoder(os.Stdout).Encode(result)
	if len(result.Errors) > 0 {
		os.Exit(1)
	}
}
