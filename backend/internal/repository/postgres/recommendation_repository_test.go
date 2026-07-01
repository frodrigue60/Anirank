package postgres

import (
	"context"
	"regexp"
	"testing"

	"anirank/api/internal/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestRecommendationRepository_GetSimilarSongsByVector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewRecommendationRepository(sqlxDB)

	vector := domain.Vector{0.1, 0.2, 0.3}
	excludeSongID := uint64(5)
	excludeAnimeID := uint64(42)
	limit := 10

	rows := sqlmock.NewRows([]string{"id", "uuid", "song_romaji"}).
		AddRow(1, "uuid-1", "Similar Song 1").
		AddRow(2, "uuid-2", "Similar Song 2")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
		WithArgs(excludeSongID, vector, limit, excludeAnimeID).
		WillReturnRows(rows)

	songs, err := repo.GetSimilarSongsByVector(context.Background(), vector, excludeSongID, excludeAnimeID, limit)
	assert.NoError(t, err)
	assert.Len(t, songs, 2)
	assert.Equal(t, uint64(1), songs[0].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRecommendationRepository_UpdateSongEmbedding(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewRecommendationRepository(sqlxDB)

	vector := domain.Vector{0.5, 0.6, 0.7}
	songID := uint64(12)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE songs SET embedding = $1 WHERE id = $2")).
		WithArgs(vector, songID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.UpdateSongEmbedding(context.Background(), songID, vector)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRecommendationRepository_GetSongsWithoutEmbeddings(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewRecommendationRepository(sqlxDB)

	limit := 5

	rows := sqlmock.NewRows([]string{"id", "uuid", "song_romaji"}).
		AddRow(101, "uuid-101", "Song without Vector")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
		WithArgs(limit).
		WillReturnRows(rows)

	songs, err := repo.GetSongsWithoutEmbeddings(context.Background(), limit)
	assert.NoError(t, err)
	assert.Len(t, songs, 1)
	assert.Equal(t, uint64(101), songs[0].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRecommendationRepository_GetUserPreferencesVector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewRecommendationRepository(sqlxDB)

	userID := uint64(99)

	// Simular retorno de dos vectores de canciones gustadas
	rows := sqlmock.NewRows([]string{"embedding"}).
		AddRow("[1.0, 0.0, 0.5]").
		AddRow("[0.0, 1.0, 0.5]")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
		WithArgs(userID).
		WillReturnRows(rows)

	avgVector, err := repo.GetUserPreferencesVector(context.Background(), userID)
	assert.NoError(t, err)
	assert.NotNil(t, avgVector)

	// El promedio de [1, 0, 0.5] y [0, 1, 0.5] debe ser [0.5, 0.5, 0.5]
	expected := domain.Vector{0.5, 0.5, 0.5}
	assert.Equal(t, expected, avgVector)
	assert.NoError(t, mock.ExpectationsWereMet())
}
