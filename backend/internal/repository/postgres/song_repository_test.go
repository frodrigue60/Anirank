package postgres

import (
	"context"
	"testing"

	"anirank/api/internal/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestSongRepository_GetPaginated_Cursor(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewSongRepository(sqlxDB)

	rows := sqlmock.NewRows([]string{"id", "uuid", "song_romaji"}).
		AddRow(1, "uuid-1", "Song 1").
		AddRow(2, "uuid-2", "Song 2")

	// 1. Test Offset Pagination (Fallback)
	mock.ExpectQuery("SELECT s.\\*").
		WithArgs(10, 0).
		WillReturnRows(rows)

	filters := domain.SongFilters{IsAdmin: true}
	songs, err := repo.GetPaginated(context.Background(), 10, 0, filters)
	assert.NoError(t, err)
	assert.Len(t, songs, 2)

	// 2. Test Cursor Pagination
	rows2 := sqlmock.NewRows([]string{"id", "uuid", "song_romaji"}).
		AddRow(3, "uuid-3", "Song 3")

	mock.ExpectQuery("SELECT s.\\*").
		WithArgs(500, 10). // 500 is the cursor ID from "500"
		WillReturnRows(rows2)

	filtersWithCursor := domain.SongFilters{
		IsAdmin: true,
		Cursor:  "500",
	}
	songs, err = repo.GetPaginated(context.Background(), 10, 0, filtersWithCursor)
	assert.NoError(t, err)
	assert.Len(t, songs, 1)
	assert.Equal(t, uint64(3), songs[0].ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}
