package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func TestGetUserInteractionsBySongIDsCombinesAllStates(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &interactionRepository{db: sqlx.NewDb(db, "postgres").Unsafe()}
	ids := []uint64{11, 12}

	mock.ExpectQuery(`SELECT song_id FROM song_user`).
		WithArgs(uint64(7), uint64(11), uint64(12)).
		WillReturnRows(sqlmock.NewRows([]string{"song_id"}).AddRow(11))
	mock.ExpectQuery(`SELECT song_id, type FROM song_reactions`).
		WithArgs(uint64(7), uint64(11), uint64(12)).
		WillReturnRows(sqlmock.NewRows([]string{"song_id", "type"}).
			AddRow(11, 1).
			AddRow(12, -1))
	mock.ExpectQuery(`SELECT song_id, rating FROM song_ratings`).
		WithArgs(uint64(7), uint64(11), uint64(12)).
		WillReturnRows(sqlmock.NewRows([]string{"song_id", "rating"}).
			AddRow(12, 8.5))

	result, err := repo.GetUserInteractionsBySongIDs(context.Background(), 7, ids)
	if err != nil {
		t.Fatal(err)
	}
	if !result[11].IsFavorited || result[11].Reaction != 1 {
		t.Fatalf("unexpected interaction for song 11: %+v", result[11])
	}
	if result[12].Reaction != -1 || result[12].Rating == nil || *result[12].Rating != 8.5 {
		t.Fatalf("unexpected interaction for song 12: %+v", result[12])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetUserInteractionsBySongIDsPropagatesQueryErrors(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &interactionRepository{db: sqlx.NewDb(db, "postgres").Unsafe()}
	mock.ExpectQuery(`SELECT song_id FROM song_user`).
		WithArgs(uint64(7), uint64(11)).
		WillReturnError(errors.New("database unavailable"))

	_, err = repo.GetUserInteractionsBySongIDs(
		context.Background(),
		7,
		[]uint64{11},
	)
	if err == nil {
		t.Fatal("expected interaction query error")
	}
}

func TestGetUserRatingInsightsReturnsSummaryAndRecentRatings(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &interactionRepository{db: sqlx.NewDb(db, "postgres").Unsafe()}
	ratedAt := time.Date(2026, time.September, 8, 12, 0, 0, 0, time.UTC)

	mock.ExpectQuery(`AVG\(sr\.rating\)::float8 AS average_score`).
		WithArgs(uint64(7)).
		WillReturnRows(sqlmock.NewRows([]string{
			"average_score",
			"total_ratings",
			"score_90_100",
			"score_75_89",
			"score_50_74",
			"score_below_50",
		}).AddRow(81.25, 4, 1, 2, 1, 0))

	mock.ExpectQuery(`SELECT sr\.song_id, sr\.rating, sr\.updated_at AS rated_at`).
		WithArgs(uint64(7), 6).
		WillReturnRows(sqlmock.NewRows([]string{"song_id", "rating", "rated_at"}).
			AddRow(11, 95, ratedAt).
			AddRow(12, 80, ratedAt.Add(-time.Hour)))

	insights, err := repo.GetUserRatingInsights(context.Background(), 7, 6)
	if err != nil {
		t.Fatal(err)
	}
	if insights.AverageScore == nil || *insights.AverageScore != 81.25 {
		t.Fatalf("unexpected average score: %v", insights.AverageScore)
	}
	if insights.TotalRatings != 4 || len(insights.Distribution) != 4 {
		t.Fatalf("unexpected summary: %+v", insights)
	}
	if insights.Distribution[0].Count != 1 || insights.Distribution[1].Count != 2 {
		t.Fatalf("unexpected distribution: %+v", insights.Distribution)
	}
	if len(insights.RecentRatings) != 2 || insights.RecentRatings[0].SongID != 11 {
		t.Fatalf("unexpected recent ratings: %+v", insights.RecentRatings)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
