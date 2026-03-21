package jobs

import (
	"context"
	"log"
	"time"

	"anirank/api/internal/domain"
	"anirank/api/internal/usecase/tournament"

	"github.com/robfig/cron/v3"
)

// StartCronScheduler initializes and runs all background jobs using robfig/cron.
func StartCronScheduler(repo domain.JobsRepository, tournamentUsecase *tournament.TournamentUsecase) *cron.Cron {
	// Setup with standard timezone parsing and seconds precision if needed, but standard is fine.
	c := cron.New(cron.WithLocation(time.UTC))

	// Register TrackDailyRanking
	_, err := c.AddFunc("0 0 * * *", func() {
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

	log.Println("Starting background cron scheduler...")
	c.Start()

	return c
}
