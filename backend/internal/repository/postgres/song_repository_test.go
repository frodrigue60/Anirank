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
	mock.ExpectQuery("SELECT s.id, s.uuid").
		WithArgs(10, 0).
		WillReturnRows(rows)

	filters := domain.SongFilters{IsAdmin: true}
	songs, err := repo.GetPaginated(context.Background(), 10, 0, filters)
	assert.NoError(t, err)
	assert.Len(t, songs, 2)

	// 2. Test Cursor Pagination
	rows2 := sqlmock.NewRows([]string{"id", "uuid", "song_romaji"}).
		AddRow(3, "uuid-3", "Song 3")

	mock.ExpectQuery("SELECT s.id, s.uuid").
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

func TestSongRepository_GetGeneratedLikedSongs(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &songRepository{db: sqlx.NewDb(db, "postgres").Unsafe()}
	rows := sqlmock.NewRows([]string{"id", "uuid", "song_romaji"}).
		AddRow(9, "song-uuid", "Song")
	mock.ExpectQuery(`(?s)JOIN song_reactions i ON i.song_id = s.id.*i.user_id = \$1 AND i.type = 1.*s.status = true AND a.status = true.*ORDER BY i.updated_at DESC`).
		WithArgs(uint64(7), 10, 0).
		WillReturnRows(rows)

	songs, err := repo.GetGeneratedPersonalSongs(context.Background(), 7, "liked", 10, 0)
	assert.NoError(t, err)
	assert.Len(t, songs, 1)
	assert.Equal(t, "song-uuid", songs[0].UUID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSongRepository_GetGeneratedFavoritedSongs(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &songRepository{db: sqlx.NewDb(db, "postgres").Unsafe()}
	rows := sqlmock.NewRows([]string{"id", "uuid", "song_romaji"}).
		AddRow(9, "song-uuid", "Song")
	mock.ExpectQuery(`(?s)JOIN song_user i ON i.song_id = s.id.*i.user_id = \$1.*s.status = true AND a.status = true.*ORDER BY i.updated_at DESC`).
		WithArgs(uint64(7), 10, 0).
		WillReturnRows(rows)

	songs, err := repo.GetGeneratedPersonalSongs(context.Background(), 7, "favorited", 10, 0)
	assert.NoError(t, err)
	assert.Len(t, songs, 1)
	assert.Equal(t, "song-uuid", songs[0].UUID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSongRepository_GetGeneratedPlaylistDescriptors(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &songRepository{db: sqlx.NewDb(db, "postgres").Unsafe()}
	season := "summer"
	rows := sqlmock.NewRows([]string{
		"key", "kind", "name", "description", "year", "season", "href", "song_count", "latest_banner",
	}).AddRow(
		"generated-season-2026-summer", "season", "Summer 2026", "Top rated songs from Summer 2026",
		2026, season, "/playlists/generated/season/2026/summer", 12, "banners/summer.webp",
	)
	mock.ExpectQuery(`(?s)WITH active_songs AS.*s.status = true AND a.status = true.*UNION ALL`).
		WillReturnRows(rows)

	descriptors, err := repo.GetGeneratedPlaylistDescriptors(context.Background())
	assert.NoError(t, err)
	assert.Len(t, descriptors, 1)
	assert.Equal(t, 12, descriptors[0].SongCount)
	assert.Equal(t, season, *descriptors[0].Season)
	assert.NoError(t, mock.ExpectationsWereMet())
}
