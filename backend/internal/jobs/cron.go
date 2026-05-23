package jobs

import (
	"context"
	"log"
	"time"

	"anirank/api/internal/domain"
	"anirank/api/internal/repository/postgres"
	"anirank/api/internal/usecase/tournament"

	"github.com/jmoiron/sqlx"
	"github.com/robfig/cron/v3"
)

// StartCronScheduler initializes and runs all background jobs using robfig/cron.
func StartCronScheduler(db *sqlx.DB, repo domain.JobsRepository, tournamentUsecase *tournament.TournamentUsecase) *cron.Cron {
	// Setup with standard timezone parsing and seconds precision if needed, but standard is fine.
	c := cron.New(cron.WithLocation(time.UTC))

	// Setup repositories for recommendation indexing
	recommendationRepo := postgres.NewRecommendationRepository(db)
	songRepo := postgres.NewSongRepository(db)
	animeRepo := postgres.NewAnimeRepository(db)
	taxonomyRepo := postgres.NewTaxonomyRepository(db)

	// Run initial embedding calculation asynchronously on startup
	go func() {
		log.Println("[CRON] Running initial ProcessPendingEmbeddings on startup...")
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer cancel()
		if err := ProcessPendingEmbeddings(ctx, recommendationRepo, songRepo, animeRepo, taxonomyRepo); err != nil {
			log.Printf("[CRON-ERR] Initial ProcessPendingEmbeddings failed: %v", err)
		}
	}()

	// Register ProcessPendingEmbeddings every 30 minutes
	_, err := c.AddFunc("*/30 * * * *", func() {
		log.Println("[CRON] Running ProcessPendingEmbeddings...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		if err := ProcessPendingEmbeddings(ctx, recommendationRepo, songRepo, animeRepo, taxonomyRepo); err != nil {
			log.Printf("[CRON-ERR] ProcessPendingEmbeddings failed: %v", err)
		}
	})

	if err != nil {
		log.Fatalf("Error registering recommendation indexer cron job: %v", err)
	}

	// Register TrackDailyRanking
	_, err = c.AddFunc("0 0 * * *", func() {
		log.Println("[CRON] Starting TrackDailyRanking job at midnight UTC...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		if err := repo.CalculateDailyRankings(ctx); err != nil {
			log.Printf("[CRON-ERR] TrackDailyRanking failed: %v", err)
		} else {
			log.Println("[CRON] Successfully completed TrackDailyRanking job.")
		}
	})

	if err != nil {
		log.Fatalf("Error registering daily ranking cron job: %v", err)
	}

	// Register ProcessTournaments every 5 minutes
	_, err = c.AddFunc("*/5 * * * *", func() {
		log.Println("[CRON] Running ProcessTournaments to advance brackets...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		if err := tournamentUsecase.ProcessTournaments(ctx); err != nil {
			log.Printf("[CRON-ERR] ProcessTournaments failed: %v", err)
		}
	})

	if err != nil {
		log.Fatalf("Error registering tournament cron job: %v", err)
	}

	// Register SnapshotRankingPositions every 6 hours
	_, err = c.AddFunc("0 */6 * * *", func() {
		log.Println("[CRON] Starting SnapshotRankingPositions job...")
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer cancel()

		if err := repo.SnapshotRankingPositions(ctx); err != nil {
			log.Printf("[CRON-ERR] SnapshotRankingPositions failed: %v", err)
		} else {
			log.Println("[CRON] Successfully completed SnapshotRankingPositions job.")
		}
	})

	if err != nil {
		log.Fatalf("Error registering ranking snapshot cron job: %v", err)
	}

	// Register SynchronizeDailySiteMetrics every midnight UTC
	_, err = c.AddFunc("0 0 * * *", func() {
		log.Println("[CRON] Starting SynchronizeDailySiteMetrics job...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		if err := repo.SynchronizeDailySiteMetrics(ctx); err != nil {
			log.Printf("[CRON-ERR] SynchronizeDailySiteMetrics failed: %v", err)
		} else {
			log.Println("[CRON] Successfully completed SynchronizeDailySiteMetrics job.")
		}
	})

	if err != nil {
		log.Fatalf("Error registering daily metrics cron job: %v", err)
	}

	log.Println("Starting background cron scheduler...")
	c.Start()

	return c
}
