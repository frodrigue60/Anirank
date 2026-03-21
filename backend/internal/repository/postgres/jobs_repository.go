package postgres

import (
	"context"

	"anirank/api/internal/domain"

	"github.com/jmoiron/sqlx"
)

type jobsRepository struct {
	db *sqlx.DB
}

func NewJobsRepository(db *sqlx.DB) domain.JobsRepository {
	return &jobsRepository{db: db}
}

func (r *jobsRepository) CalculateDailyRankings(ctx context.Context) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Clear today's previous run
	deleteQuery := `DELETE FROM ranking_histories WHERE date = CURRENT_DATE`
	if _, err := tx.ExecContext(ctx, deleteQuery); err != nil {
		return err
	}

	// 2. Insert computed scores
	insertQuery := `
		INSERT INTO ranking_histories (song_id, date, score, rank, seasonal_rank, created_at, updated_at)
		SELECT 
			s.id, 
			CURRENT_DATE,
			(
				COALESCE(s.views, 0) * 1.0 + 
				COALESCE(rc.likes_count, 0) * 10.0 - 
				COALESCE(rc.dislikes_count, 0) * 5.0 + 
				COALESCE((SELECT AVG(rating) FROM song_ratings WHERE song_id = s.id), 0) * 50.0
			) as calculated_score,
			0, 
			0,
			CURRENT_TIMESTAMP, 
			CURRENT_TIMESTAMP
		FROM songs s
		LEFT JOIN (
			SELECT id as reactable_id, likes_count, dislikes_count FROM songs
		) rc ON rc.reactable_id = s.id
	`
	// Note: The original query joined reaction_counters which might be legacy or part of a different model.
	// Since I refactored songs table to have counts, I'm using those.
	if _, err := tx.ExecContext(ctx, insertQuery); err != nil {
		return err
	}

	// 3. Update the 'rank' column based on the computed score (global rank for today)
	// PostgreSQL UPDATE FROM syntax
	rankQuery := `
		UPDATE ranking_histories rh
		SET rank = ranked.new_rank
		FROM (
			SELECT id, ROW_NUMBER() OVER (PARTITION BY date ORDER BY score DESC) as new_rank
			FROM ranking_histories
			WHERE date = CURRENT_DATE
		) ranked
		WHERE rh.id = ranked.id
	`
	if _, err := tx.ExecContext(ctx, rankQuery); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *jobsRepository) SnapshotRankingPositions(ctx context.Context) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Snapshot Seasonal Rank
	seasonalQuery := `
		UPDATE songs s
		SET prev_seasonal_rank = ranked_list.current_rank
		FROM (
			SELECT ss.id, ROW_NUMBER() OVER (
				ORDER BY (SELECT COALESCE(AVG(rating), 0) FROM song_ratings WHERE song_id = ss.id) DESC, ss.views DESC
			) as current_rank
			FROM songs ss
			JOIN seasons sea ON ss.season_id = sea.id
			JOIN years y ON ss.year_id = y.id
			WHERE sea.current = true AND y.current = true
		) as ranked_list
		WHERE s.id = ranked_list.id
	`
	if _, err := tx.ExecContext(ctx, seasonalQuery); err != nil {
		return err
	}

	// 2. Snapshot Global Rank
	globalQuery := `
		UPDATE songs s
		SET prev_main_rank = ranked_list.current_rank
		FROM (
			SELECT ss.id, ROW_NUMBER() OVER (
				ORDER BY (SELECT COALESCE(AVG(rating), 0) FROM song_ratings WHERE song_id = ss.id) DESC, ss.views DESC
			) as current_rank
			FROM songs ss
		) as ranked_list
		WHERE s.id = ranked_list.id
	`
	if _, err := tx.ExecContext(ctx, globalQuery); err != nil {
		return err
	}

	return tx.Commit()
}
