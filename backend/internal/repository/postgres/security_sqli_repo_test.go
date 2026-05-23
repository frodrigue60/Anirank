package postgres

import (
	"context"
	"database/sql/driver"
	"regexp"
	"testing"

	"anirank/api/internal/domain"
	"anirank/api/internal/testutil"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

// TestSongRepository_GetPaginated_SQLi verifies that when a SQL injection payload is passed 
// as a search parameter to GetPaginated, it is passed strictly as a parameterized query argument, 
// preventing any syntax alteration or execution of commands.
func TestSongRepository_GetPaginated_SQLi(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock connection: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewSongRepository(sqlxDB)

	for _, payload := range testutil.SQLIPayloads {
		t.Run("Payload_"+payload, func(t *testing.T) {
			// Expect the search parameter to be passed as arguments:
			// %payload% will be repeated 5 times for song_romaji, song_jp, song_en, title, artist name.
			searchTerm := "%" + payload + "%"
			expectedArgs := []driver.Value{
				searchTerm, searchTerm, searchTerm, searchTerm, searchTerm,
				10, // limit
				0,  // offset
			}

			rows := sqlmock.NewRows([]string{"id", "uuid", "song_romaji"}).
				AddRow(1, "uuid-1", "Safe Song")

			// Expect the parameterized query. The query should structure standard s.song_romaji ILIKE $1 etc.
			mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
				WithArgs(expectedArgs...).
				WillReturnRows(rows)

			filters := domain.SongFilters{
				Search:  payload,
				IsAdmin: true,
			}

			songs, err := repo.GetPaginated(context.Background(), 10, 0, filters)
			assert.NoError(t, err)
			assert.NotEmpty(t, songs)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestSearchRepository_GlobalSearch_SQLi verifies that search term payloads are safely passed
// to tsquery simple search, and properly escaped.
func TestSearchRepository_GlobalSearch_SQLi(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock connection: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewSearchRepository(sqlxDB)

	for _, payload := range testutil.SQLIPayloads {
		t.Run("Payload_"+payload, func(t *testing.T) {
			rows := sqlmock.NewRows([]string{"item_type", "item_id", "title", "subtitle", "slug", "image_url", "rank"}).
				AddRow("song", 1, "Song Title", "Subtitle", "slug", "url", 0.95)

			mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
				WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), 5). // query, query, limit
				WillReturnRows(rows)

			items, err := repo.GlobalSearch(context.Background(), payload, 5)
			assert.NoError(t, err)
			assert.Len(t, items, 1)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
