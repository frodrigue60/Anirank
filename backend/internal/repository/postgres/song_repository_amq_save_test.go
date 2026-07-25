package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"anirank/api/internal/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newAMQSaveTestRepo(t *testing.T) (domain.SongRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewSongRepository(sqlx.NewDb(db, "postgres")), mock
}

func TestAmqSaveThemeTypeClause(t *testing.T) {
	t.Run("empty theme types", func(t *testing.T) {
		clause, args, next := amqSaveThemeTypeClause(nil, 1)
		assert.Empty(t, clause)
		assert.Nil(t, args)
		assert.Equal(t, 1, next)
	})

	t.Run("single OP filter", func(t *testing.T) {
		clause, args, next := amqSaveThemeTypeClause([]string{"OP"}, 2)
		assert.Contains(t, clause, "st.slug IN ($2)")
		assert.Equal(t, []interface{}{"OP"}, args)
		assert.Equal(t, 3, next)
	})
}

func TestFindRandomArtistForAMQSave(t *testing.T) {
	repo, mock := newAMQSaveTestRepo(t)

	mock.ExpectQuery("FROM artists ar").
		WithArgs("OP", 4).
		WillReturnRows(sqlmock.NewRows([]string{"id", "uuid", "name"}).
			AddRow(uint64(10), "artist-uuid", "LiSA"))

	anchor, err := repo.FindRandomArtistForAMQSave(context.Background(), []string{"OP"}, 4)
	require.NoError(t, err)
	require.NotNil(t, anchor)
	assert.Equal(t, "artist", anchor.Kind)
	assert.Equal(t, uint64(10), *anchor.ArtistID)
	assert.Equal(t, "artist-uuid", anchor.ArtistUUID)
	assert.Equal(t, "LiSA", anchor.ArtistName)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindRandomArtistForAMQSave_NoRows(t *testing.T) {
	repo, mock := newAMQSaveTestRepo(t)

	mock.ExpectQuery("FROM artists ar").
		WithArgs("ED", 2).
		WillReturnError(sql.ErrNoRows)

	anchor, err := repo.FindRandomArtistForAMQSave(context.Background(), []string{"ED"}, 2)
	require.NoError(t, err)
	assert.Nil(t, anchor)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindRandomYearForAMQSave(t *testing.T) {
	repo, mock := newAMQSaveTestRepo(t)

	mock.ExpectQuery("FROM years y").
		WithArgs("OP", 4).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
			AddRow(uint64(2011), "2011"))

	anchor, err := repo.FindRandomYearForAMQSave(context.Background(), []string{"OP"}, 4)
	require.NoError(t, err)
	require.NotNil(t, anchor)
	assert.Equal(t, "year", anchor.Kind)
	assert.Equal(t, uint64(2011), *anchor.YearID)
	assert.Equal(t, "2011", anchor.YearName)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindRandomSeasonYearForAMQSave(t *testing.T) {
	repo, mock := newAMQSaveTestRepo(t)

	mock.ExpectQuery("GROUP BY s.season_id").
		WithArgs("OP", 4).
		WillReturnRows(sqlmock.NewRows([]string{"season_id", "season_name", "year_id", "year_name"}).
			AddRow(uint64(1), "Winter", uint64(2020), "2020"))

	anchor, err := repo.FindRandomSeasonYearForAMQSave(context.Background(), []string{"OP"}, 4)
	require.NoError(t, err)
	require.NotNil(t, anchor)
	assert.Equal(t, "season", anchor.Kind)
	assert.Equal(t, uint64(1), *anchor.SeasonID)
	assert.Equal(t, "Winter", anchor.SeasonName)
	assert.Equal(t, uint64(2020), *anchor.YearID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindRandomAnimeForAMQSave(t *testing.T) {
	repo, mock := newAMQSaveTestRepo(t)

	mock.ExpectQuery("HAVING COUNT\\(DISTINCT LOWER\\(TRIM\\(COALESCE").
		WithArgs("OP", 4).
		WillReturnRows(sqlmock.NewRows([]string{"id", "uuid", "title"}).
			AddRow(uint64(99), "anime-uuid", "Test Anime"))

	anchor, err := repo.FindRandomAnimeForAMQSave(context.Background(), []string{"OP"}, 4)
	require.NoError(t, err)
	require.NotNil(t, anchor)
	assert.Equal(t, "anime", anchor.Kind)
	assert.Equal(t, uint64(99), *anchor.AnimeID)
	assert.Equal(t, "anime-uuid", anchor.AnimeUUID)
	assert.Equal(t, "Test Anime", anchor.AnimeTitle)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindRandomAnimeForAMQSave_QueryUsesDistinctSongNames(t *testing.T) {
	query, args := buildFindRandomAnimeForAMQSaveQuery([]string{"OP"}, 4)
	assert.Contains(t, query, amqSaveSongNameKey)
	assert.NotContains(t, query, "theme_num")
	assert.Equal(t, []interface{}{"OP", 4}, args)
}

func TestFindRandomGenreForAMQSave(t *testing.T) {
	repo, mock := newAMQSaveTestRepo(t)

	mock.ExpectQuery("FROM genres g").
		WithArgs("ED", 4).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
			AddRow(uint64(3), "Action"))

	anchor, err := repo.FindRandomGenreForAMQSave(context.Background(), []string{"ED"}, 4)
	require.NoError(t, err)
	require.NotNil(t, anchor)
	assert.Equal(t, "genre", anchor.Kind)
	assert.Equal(t, uint64(3), *anchor.GenreID)
	assert.Equal(t, "Action", anchor.GenreName)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRandomSongIDsForAMQSave_ArtistAnchor(t *testing.T) {
	repo, mock := newAMQSaveTestRepo(t)
	artistID := uint64(10)

	mock.ExpectQuery("SELECT id FROM").
		WithArgs("OP", artistID, 4).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).
			AddRow(uint64(101)).
			AddRow(uint64(102)).
			AddRow(uint64(103)).
			AddRow(uint64(104)))

	ids, err := repo.GetRandomSongIDsForAMQSave(context.Background(), domain.AMQSaveThemeAnchor{
		Kind:     "artist",
		ArtistID: &artistID,
	}, []string{"OP"}, 4)
	require.NoError(t, err)
	assert.Equal(t, []uint64{101, 102, 103, 104}, ids)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRandomSongIDsForAMQSave_FallbackPool(t *testing.T) {
	repo, mock := newAMQSaveTestRepo(t)

	mock.ExpectQuery("SELECT id FROM").
		WithArgs("OP", 6).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).
			AddRow(uint64(1)).
			AddRow(uint64(2)))

	ids, err := repo.GetRandomSongIDsForAMQSave(context.Background(), domain.AMQSaveThemeAnchor{
		Kind: "fallback",
	}, []string{"OP"}, 6)
	require.NoError(t, err)
	assert.Equal(t, []uint64{1, 2}, ids)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRandomSongIDsForAMQSave_AnimeAnchor(t *testing.T) {
	repo, mock := newAMQSaveTestRepo(t)
	animeID := uint64(99)

	mock.ExpectQuery("ROW_NUMBER\\(\\) OVER \\(PARTITION BY").
		WithArgs("ED", animeID, 3).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).
			AddRow(uint64(501)).
			AddRow(uint64(502)).
			AddRow(uint64(503)))

	ids, err := repo.GetRandomSongIDsForAMQSave(context.Background(), domain.AMQSaveThemeAnchor{
		Kind:    "anime",
		AnimeID: &animeID,
	}, []string{"ED"}, 3)
	require.NoError(t, err)
	assert.Len(t, ids, 3)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRandomAnimeSongIDsForAMQSave_QueryDedupesBySongName(t *testing.T) {
	query, args := buildGetRandomAnimeSongIDsForAMQSaveQuery([]string{"OP"}, 42, 4)
	assert.Contains(t, query, "ROW_NUMBER() OVER (PARTITION BY")
	assert.Contains(t, query, amqSaveSongNameKey)
	assert.Equal(t, []interface{}{"OP", uint64(42), 4}, args)
}

func TestGetRandomSongIDsForAMQSave_MissingArtistID(t *testing.T) {
	repo, _ := newAMQSaveTestRepo(t)

	_, err := repo.GetRandomSongIDsForAMQSave(context.Background(), domain.AMQSaveThemeAnchor{
		Kind: "artist",
	}, []string{"OP"}, 4)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "artist anchor missing artist_id")
}

func TestGetRandomSongIDsForAMQSave_UnsupportedKind(t *testing.T) {
	repo, _ := newAMQSaveTestRepo(t)

	_, err := repo.GetRandomSongIDsForAMQSave(context.Background(), domain.AMQSaveThemeAnchor{
		Kind: "unknown",
	}, []string{"OP"}, 4)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported anchor kind")
}

func TestAMQSaveQueries_RequireLocalVideoOnly(t *testing.T) {
	repo, mock := newAMQSaveTestRepo(t)

	mock.ExpectQuery("video_src NOT LIKE 'http%").
		WithArgs("OP", 4).
		WillReturnRows(sqlmock.NewRows([]string{"id", "uuid", "name"}).
			AddRow(uint64(1), "a", "Artist"))

	_, err := repo.FindRandomArtistForAMQSave(context.Background(), []string{"OP"}, 4)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindRandomArtistForAMQSave_QueryJoinBeforeWhere(t *testing.T) {
	query, args := buildFindRandomArtistForAMQSaveQuery([]string{"OP"}, 4)
	artistJoin := strings.Index(query, "JOIN artist_song")
	whereClause := strings.Index(query, "WHERE s.status = true")
	require.Greater(t, artistJoin, 0)
	require.Greater(t, whereClause, 0)
	assert.Less(t, artistJoin, whereClause)
	assert.Equal(t, []interface{}{"OP", 4}, args)
}

func TestFindRandomGenreForAMQSave_QueryJoinBeforeWhere(t *testing.T) {
	query, args := buildFindRandomGenreForAMQSaveQuery([]string{"ED"}, 6)
	genreJoin := strings.Index(query, "JOIN anime_genre")
	whereClause := strings.Index(query, "WHERE s.status = true")
	require.Greater(t, genreJoin, 0)
	require.Greater(t, whereClause, 0)
	assert.Less(t, genreJoin, whereClause)
	assert.Equal(t, []interface{}{"ED", 6}, args)
}
