package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"anirank/api/internal/domain"
)

const amqSaveEligibleSongFrom = `
FROM songs s
JOIN animes a ON s.anime_id = a.id
JOIN song_types st ON s.type_id = st.id
JOIN song_variants sv ON s.id = sv.song_id
JOIN videos v ON sv.id = v.song_variant_id
`

const amqSaveEligibleSongWhere = `
WHERE s.status = true
  AND a.status = true
  AND sv.status = true
  AND v.video_src IS NOT NULL
  AND v.video_src <> ''
  AND v.video_src NOT LIKE 'http%%'
`

const amqSaveEligibleSongJoin = amqSaveEligibleSongFrom + amqSaveEligibleSongWhere

func amqSaveThemeTypeClause(themeTypes []string, argIdx int) (string, []interface{}, int) {
	if len(themeTypes) == 0 {
		return "", nil, argIdx
	}
	placeholders := make([]string, len(themeTypes))
	args := make([]interface{}, len(themeTypes))
	for i, t := range themeTypes {
		placeholders[i] = fmt.Sprintf("$%d", argIdx+i)
		args[i] = t
	}
	clause := fmt.Sprintf(" AND st.slug IN (%s)", joinWS(placeholders, ", "))
	return clause, args, argIdx + len(themeTypes)
}

func buildFindRandomArtistForAMQSaveQuery(themeTypes []string, minSongs int) (string, []interface{}) {
	typeClause, typeArgs, nextIdx := amqSaveThemeTypeClause(themeTypes, 1)
	query := fmt.Sprintf(`
		SELECT ar.id, ar.uuid, ar.name
		FROM artists ar
		WHERE ar.status = true
		  AND (
		    SELECT COUNT(DISTINCT s.id)
		    %s
		    JOIN artist_song asng ON s.id = asng.song_id AND asng.artist_id = ar.id
		    %s
		    %s
		  ) >= $%d
		ORDER BY RANDOM()
		LIMIT 1
	`, amqSaveEligibleSongFrom, amqSaveEligibleSongWhere, typeClause, nextIdx)
	return query, append(typeArgs, minSongs)
}

func (r *songRepository) FindRandomArtistForAMQSave(ctx context.Context, themeTypes []string, minSongs int) (*domain.AMQSaveThemeAnchor, error) {
	query, args := buildFindRandomArtistForAMQSaveQuery(themeTypes, minSongs)
	var row struct {
		ID   uint64 `db:"id"`
		UUID string `db:"uuid"`
		Name string `db:"name"`
	}
	if err := r.db.GetContext(ctx, &row, query, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	id := row.ID
	return &domain.AMQSaveThemeAnchor{
		Kind:       "artist",
		ArtistID:   &id,
		ArtistUUID: row.UUID,
		ArtistName: row.Name,
	}, nil
}

func (r *songRepository) FindRandomYearForAMQSave(ctx context.Context, themeTypes []string, minSongs int) (*domain.AMQSaveThemeAnchor, error) {
	typeClause, typeArgs, nextIdx := amqSaveThemeTypeClause(themeTypes, 1)
	query := fmt.Sprintf(`
		SELECT y.id, y.name
		FROM years y
		WHERE (
		  SELECT COUNT(DISTINCT s.id)
		  %s
		  AND s.year_id = y.id
		  %s
		) >= $%d
		ORDER BY RANDOM()
		LIMIT 1
	`, amqSaveEligibleSongJoin, typeClause, nextIdx)

	args := append(typeArgs, minSongs)
	var row struct {
		ID   uint64 `db:"id"`
		Name string `db:"name"`
	}
	if err := r.db.GetContext(ctx, &row, query, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	id := row.ID
	return &domain.AMQSaveThemeAnchor{
		Kind:     "year",
		YearID:   &id,
		YearName: row.Name,
	}, nil
}

func (r *songRepository) FindRandomSeasonYearForAMQSave(ctx context.Context, themeTypes []string, minSongs int) (*domain.AMQSaveThemeAnchor, error) {
	typeClause, typeArgs, nextIdx := amqSaveThemeTypeClause(themeTypes, 1)
	query := fmt.Sprintf(`
		SELECT s.season_id, se.name AS season_name, s.year_id, y.name AS year_name
		FROM songs s
		JOIN seasons se ON s.season_id = se.id
		JOIN years y ON s.year_id = y.id
		JOIN animes a ON s.anime_id = a.id
		JOIN song_types st ON s.type_id = st.id
		JOIN song_variants sv ON s.id = sv.song_id
		JOIN videos v ON sv.id = v.song_variant_id
		WHERE s.status = true
		  AND a.status = true
		  AND sv.status = true
		  AND v.video_src IS NOT NULL
		  AND v.video_src <> ''
		  AND v.video_src NOT LIKE 'http%%'
		  %s
		GROUP BY s.season_id, se.name, s.year_id, y.name
		HAVING COUNT(DISTINCT s.id) >= $%d
		ORDER BY RANDOM()
		LIMIT 1
	`, typeClause, nextIdx)

	args := append(typeArgs, minSongs)
	var row struct {
		SeasonID   uint64 `db:"season_id"`
		SeasonName string `db:"season_name"`
		YearID     uint64 `db:"year_id"`
		YearName   string `db:"year_name"`
	}
	if err := r.db.GetContext(ctx, &row, query, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	seasonID := row.SeasonID
	yearID := row.YearID
	return &domain.AMQSaveThemeAnchor{
		Kind:       "season",
		SeasonID:   &seasonID,
		SeasonName: row.SeasonName,
		YearID:     &yearID,
		YearName:   row.YearName,
	}, nil
}

func (r *songRepository) FindRandomAnimeForAMQSave(ctx context.Context, themeTypes []string, minThemes int) (*domain.AMQSaveThemeAnchor, error) {
	typeClause, typeArgs, nextIdx := amqSaveThemeTypeClause(themeTypes, 1)
	query := fmt.Sprintf(`
		SELECT a.id, a.uuid, a.title
		FROM (
		  SELECT s.anime_id,
		         COUNT(DISTINCT (st.slug || ':' || s.theme_num)) AS theme_count
		  %s
		  %s
		  GROUP BY s.anime_id
		  HAVING COUNT(DISTINCT (st.slug || ':' || s.theme_num)) >= $%d
		) q
		JOIN animes a ON q.anime_id = a.id
		ORDER BY RANDOM()
		LIMIT 1
	`, amqSaveEligibleSongJoin, typeClause, nextIdx)

	args := append(typeArgs, minThemes)
	var row struct {
		ID    uint64 `db:"id"`
		UUID  string `db:"uuid"`
		Title string `db:"title"`
	}
	if err := r.db.GetContext(ctx, &row, query, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	id := row.ID
	return &domain.AMQSaveThemeAnchor{
		Kind:       "anime",
		AnimeID:    &id,
		AnimeUUID:  row.UUID,
		AnimeTitle: row.Title,
	}, nil
}

func buildFindRandomGenreForAMQSaveQuery(themeTypes []string, minSongs int) (string, []interface{}) {
	typeClause, typeArgs, nextIdx := amqSaveThemeTypeClause(themeTypes, 1)
	query := fmt.Sprintf(`
		SELECT g.id, g.name
		FROM genres g
		WHERE (
		  SELECT COUNT(DISTINCT s.id)
		  %s
		  JOIN anime_genre ag ON ag.anime_id = s.anime_id AND ag.genre_id = g.id
		  %s
		  %s
		) >= $%d
		ORDER BY RANDOM()
		LIMIT 1
	`, amqSaveEligibleSongFrom, amqSaveEligibleSongWhere, typeClause, nextIdx)
	return query, append(typeArgs, minSongs)
}

func (r *songRepository) FindRandomGenreForAMQSave(ctx context.Context, themeTypes []string, minSongs int) (*domain.AMQSaveThemeAnchor, error) {
	query, args := buildFindRandomGenreForAMQSaveQuery(themeTypes, minSongs)
	var row struct {
		ID   uint64 `db:"id"`
		Name string `db:"name"`
	}
	if err := r.db.GetContext(ctx, &row, query, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	id := row.ID
	return &domain.AMQSaveThemeAnchor{
		Kind:      "genre",
		GenreID:   &id,
		GenreName: row.Name,
	}, nil
}

func (r *songRepository) GetRandomSongIDsForAMQSave(ctx context.Context, anchor domain.AMQSaveThemeAnchor, themeTypes []string, count int) ([]uint64, error) {
	typeClause, typeArgs, nextIdx := amqSaveThemeTypeClause(themeTypes, 1)
	var filterClause string
	var filterArgs []interface{}

	switch anchor.Kind {
	case "artist":
		if anchor.ArtistID == nil {
			return nil, fmt.Errorf("artist anchor missing artist_id")
		}
		filterClause = fmt.Sprintf(" AND s.id IN (SELECT song_id FROM artist_song WHERE artist_id = $%d)", nextIdx)
		filterArgs = append(filterArgs, *anchor.ArtistID)
		nextIdx++
	case "year":
		if anchor.YearID == nil {
			return nil, fmt.Errorf("year anchor missing year_id")
		}
		filterClause = fmt.Sprintf(" AND s.year_id = $%d", nextIdx)
		filterArgs = append(filterArgs, *anchor.YearID)
		nextIdx++
	case "season":
		if anchor.SeasonID == nil || anchor.YearID == nil {
			return nil, fmt.Errorf("season anchor missing season_id or year_id")
		}
		filterClause = fmt.Sprintf(" AND s.season_id = $%d AND s.year_id = $%d", nextIdx, nextIdx+1)
		filterArgs = append(filterArgs, *anchor.SeasonID, *anchor.YearID)
		nextIdx += 2
	case "anime":
		if anchor.AnimeID == nil {
			return nil, fmt.Errorf("anime anchor missing anime_id")
		}
		filterClause = fmt.Sprintf(" AND s.anime_id = $%d", nextIdx)
		filterArgs = append(filterArgs, *anchor.AnimeID)
		nextIdx++
	case "genre":
		if anchor.GenreID == nil {
			return nil, fmt.Errorf("genre anchor missing genre_id")
		}
		filterClause = fmt.Sprintf(" AND s.anime_id IN (SELECT anime_id FROM anime_genre WHERE genre_id = $%d)", nextIdx)
		filterArgs = append(filterArgs, *anchor.GenreID)
		nextIdx++
	case "fallback":
		filterClause = ""
	default:
		return nil, fmt.Errorf("unsupported anchor kind: %s", anchor.Kind)
	}

	query := fmt.Sprintf(`
		SELECT id FROM (
		  SELECT DISTINCT s.id
		  %s
		  %s
		  %s
		) q
		ORDER BY RANDOM()
		LIMIT $%d
	`, amqSaveEligibleSongJoin, typeClause, filterClause, nextIdx)

	args := append(typeArgs, filterArgs...)
	args = append(args, count)

	var ids []uint64
	if err := r.db.SelectContext(ctx, &ids, query, args...); err != nil {
		return nil, err
	}
	if ids == nil {
		return []uint64{}, nil
	}
	return ids, nil
}
