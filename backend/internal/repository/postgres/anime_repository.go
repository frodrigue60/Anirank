package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"anirank/api/internal/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type animeRepository struct {
	db *sqlx.DB
}

func NewAnimeRepository(db *sqlx.DB) domain.AnimeRepository {
	return &animeRepository{db: db.Unsafe()}
}

func (r *animeRepository) GetByID(ctx context.Context, id uint64) (*domain.Anime, error) {
	var anime domain.Anime
	query := "SELECT * FROM animes WHERE id = $1"
	err := r.db.GetContext(ctx, &anime, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &anime, err
}

func (r *animeRepository) GetByUUID(ctx context.Context, uuid string) (*domain.Anime, error) {
	var anime domain.Anime
	query := "SELECT * FROM animes WHERE uuid = $1"
	err := r.db.GetContext(ctx, &anime, query, uuid)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &anime, err
}

func (r *animeRepository) GetMany(ctx context.Context, ids []uint64) ([]domain.Anime, error) {
	if len(ids) == 0 {
		return []domain.Anime{}, nil
	}
	query, args, err := sqlx.In("SELECT * FROM animes WHERE id IN (?)", ids)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)
	var animes []domain.Anime
	err = r.db.SelectContext(ctx, &animes, query, args...)
	if animes == nil {
		animes = []domain.Anime{}
	}
	return animes, err
}

func (r *animeRepository) GetBySlug(ctx context.Context, slug string) (*domain.Anime, error) {
	var anime domain.Anime
	// status = true (active) is a boolean flag
	query := "SELECT * FROM animes WHERE slug = $1 AND status = true"
	err := r.db.GetContext(ctx, &anime, query, slug)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &anime, err
}

func (r *animeRepository) GetByAnilistID(ctx context.Context, anilistID int64) (*domain.Anime, error) {
	var anime domain.Anime
	query := "SELECT * FROM animes WHERE anilist_id = $1"
	err := r.db.GetContext(ctx, &anime, query, anilistID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &anime, err
}

func (r *animeRepository) GetByAnilistIDs(ctx context.Context, anilistIDs []int) ([]domain.Anime, error) {
	if len(anilistIDs) == 0 {
		return []domain.Anime{}, nil
	}
	query, args, err := sqlx.In("SELECT * FROM animes WHERE anilist_id IN (?)", anilistIDs)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)
	var animes []domain.Anime
	err = r.db.SelectContext(ctx, &animes, query, args...)
	if animes == nil {
		animes = []domain.Anime{}
	}
	return animes, err
}

func (r *animeRepository) GetPaginated(ctx context.Context, limit, offset int, filters domain.AnimeFilters) ([]domain.Anime, error) {
	animes := []domain.Anime{}
	query := "SELECT animes.* FROM animes"
	var args []interface{}
	i := 1

	// Joins (Only for public slug-based filtering)
	if !filters.IsAdmin {
		if filters.Year != "" {
			query += " JOIN years y ON animes.year_id = y.id"
		}
		if filters.Season != "" {
			query += " JOIN seasons s ON animes.season_id = s.id"
		}
		if filters.Format != "" {
			query += " JOIN formats f ON animes.format_id = f.id"
		}
		if filters.Genre != "" {
			query += " JOIN anime_genre ag ON animes.id = ag.anime_id"
			query += " JOIN genres g ON ag.genre_id = g.id"
		}
	} else if filters.Genre != "" {
		// Admin still needs join for genres but filters by genre_id
		query += " JOIN anime_genre ag ON animes.id = ag.anime_id"
	}

	// Where Clause
	if !filters.IsAdmin {
		query += " WHERE animes.status = true"
	} else {
		query += " WHERE true"
	}

	if filters.IsAdmin && filters.Status != nil {
		query += fmt.Sprintf(" AND animes.status = $%d", i)
		args = append(args, *filters.Status)
		i++
	}

	if filters.Year != "" {
		if filters.IsAdmin {
			query += fmt.Sprintf(" AND animes.year_id = $%d", i)
		} else {
			query += fmt.Sprintf(" AND y.slug = $%d", i)
		}
		args = append(args, filters.Year)
		i++
	}
	if filters.Season != "" {
		if filters.IsAdmin {
			query += fmt.Sprintf(" AND animes.season_id = $%d", i)
		} else {
			query += fmt.Sprintf(" AND s.slug = $%d", i)
		}
		args = append(args, filters.Season)
		i++
	}
	if filters.Format != "" {
		if filters.IsAdmin {
			query += fmt.Sprintf(" AND animes.format_id = $%d", i)
		} else {
			query += fmt.Sprintf(" AND f.slug = $%d", i)
		}
		args = append(args, filters.Format)
		i++
	}
	if filters.Genre != "" {
		if filters.IsAdmin {
			query += fmt.Sprintf(" AND ag.genre_id = $%d", i)
		} else {
			query += fmt.Sprintf(" AND g.slug = $%d", i)
		}
		args = append(args, filters.Genre)
		i++
	}
	if filters.Search != "" {
		query += fmt.Sprintf(" AND (animes.title ILIKE $%d OR animes.title_english ILIKE $%d OR animes.title_native ILIKE $%d OR $%d = ANY(animes.synonyms))", i, i, i, i+1)
		args = append(args, "%"+filters.Search+"%", filters.Search)
		i += 2
	}

	// Sorting
	switch filters.Sort {
	case "title":
		query += " ORDER BY animes.title ASC"
	case "latest":
		query += " ORDER BY animes.created_at DESC, animes.id DESC"
	case "most_themes":
		query += " ORDER BY animes.songs_count DESC"
	case "least_themes":
		query += " ORDER BY animes.songs_count ASC"
	default:
		query += " ORDER BY animes.created_at DESC, animes.id DESC"
	}

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", i, i+1)
	args = append(args, limit, offset)

	err := r.db.SelectContext(ctx, &animes, query, args...)
	return animes, err
}

func (r *animeRepository) Count(ctx context.Context, filters domain.AnimeFilters) (int, error) {
	var count int
	query := "SELECT COUNT(DISTINCT animes.id) FROM animes"
	var args []interface{}
	i := 1

	// Joins (Only for public slug-based filtering)
	if !filters.IsAdmin {
		if filters.Year != "" {
			query += " JOIN years y ON animes.year_id = y.id"
		}
		if filters.Season != "" {
			query += " JOIN seasons s ON animes.season_id = s.id"
		}
		if filters.Format != "" {
			query += " JOIN formats f ON animes.format_id = f.id"
		}
		if filters.Genre != "" {
			query += " JOIN anime_genre ag ON animes.id = ag.anime_id"
			query += " JOIN genres g ON ag.genre_id = g.id"
		}
	} else if filters.Genre != "" {
		// Admin still needs join for genres but filters by genre_id
		query += " JOIN anime_genre ag ON animes.id = ag.anime_id"
	}

	// Where Clause
	if !filters.IsAdmin {
		query += " WHERE animes.status = true"
	} else {
		query += " WHERE true"
	}

	if filters.IsAdmin && filters.Status != nil {
		query += fmt.Sprintf(" AND animes.status = $%d", i)
		args = append(args, *filters.Status)
		i++
	}

	if filters.Year != "" {
		if filters.IsAdmin {
			query += fmt.Sprintf(" AND animes.year_id = $%d", i)
		} else {
			query += fmt.Sprintf(" AND y.slug = $%d", i)
		}
		args = append(args, filters.Year)
		i++
	}
	if filters.Season != "" {
		if filters.IsAdmin {
			query += fmt.Sprintf(" AND animes.season_id = $%d", i)
		} else {
			query += fmt.Sprintf(" AND s.slug = $%d", i)
		}
		args = append(args, filters.Season)
		i++
	}
	if filters.Format != "" {
		if filters.IsAdmin {
			query += fmt.Sprintf(" AND animes.format_id = $%d", i)
		} else {
			query += fmt.Sprintf(" AND f.slug = $%d", i)
		}
		args = append(args, filters.Format)
		i++
	}
	if filters.Genre != "" {
		if filters.IsAdmin {
			query += fmt.Sprintf(" AND ag.genre_id = $%d", i)
		} else {
			query += fmt.Sprintf(" AND g.slug = $%d", i)
		}
		args = append(args, filters.Genre)
		i++
	}
	if filters.Search != "" {
		query += fmt.Sprintf(" AND (animes.title ILIKE $%d OR animes.title_english ILIKE $%d OR animes.title_native ILIKE $%d OR $%d = ANY(animes.synonyms))", i, i, i, i+1)
		args = append(args, "%"+filters.Search+"%", filters.Search)
		i += 2
	}

	err := r.db.GetContext(ctx, &count, query, args...)
	return count, err
}

// Write Operations
func (r *animeRepository) Create(ctx context.Context, anime *domain.Anime) error {
	query := `
		INSERT INTO animes (uuid, title, slug, description, anilist_id, anime_themes_id, status, year_id, season_id, format_id, cover, banner, title_english, title_native, synonyms, created_at, updated_at) 
		VALUES (:uuid, :title, :slug, :description, :anilist_id, :anime_themes_id, :status, :year_id, :season_id, :format_id, :cover, :banner, :title_english, :title_native, :synonyms, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id
	`
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return err
	}
	err = stmt.QueryRowContext(ctx, anime).Scan(&anime.ID)
	return err
}

func (r *animeRepository) Update(ctx context.Context, anime *domain.Anime) error {
	query := `
		UPDATE animes 
		SET title = :title, slug = :slug, description = :description, anilist_id = :anilist_id, 
		    anime_themes_id = :anime_themes_id, status = :status, year_id = :year_id, 
		    season_id = :season_id, format_id = :format_id, 
		    cover = :cover, banner = :banner, title_english = :title_english,
		    title_native = :title_native, synonyms = :synonyms, updated_at = CURRENT_TIMESTAMP
		WHERE id = :id
	`
	res, err := r.db.NamedExecContext(ctx, query, anime)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err == nil && affected == 0 {
		return nil
	}
	return err
}

func (r *animeRepository) Delete(ctx context.Context, id uint64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Delete junction table entries
	_, _ = tx.ExecContext(ctx, "DELETE FROM anime_genre WHERE anime_id = $1", id)
	_, _ = tx.ExecContext(ctx, "DELETE FROM anime_studio WHERE anime_id = $1", id)
	_, _ = tx.ExecContext(ctx, "DELETE FROM anime_producer WHERE anime_id = $1", id)

	// 2. Delete external links
	_, _ = tx.ExecContext(ctx, "DELETE FROM external_links WHERE id IN (SELECT external_link_id FROM anime_external_link WHERE anime_id = $1)", id)
	_, _ = tx.ExecContext(ctx, "DELETE FROM anime_external_link WHERE anime_id = $1", id)

	// 3. Delete songs
	_, _ = tx.ExecContext(ctx, "DELETE FROM songs WHERE anime_id = $1", id)

	// 4. Delete the anime record itself
	query := "DELETE FROM animes WHERE id = $1"
	res, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err == nil && affected == 0 {
		return errors.New("anime not found")
	}

	return tx.Commit()
}

func (r *animeRepository) BatchDelete(ctx context.Context, ids []uint64) error {
	if len(ids) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Helper to handle sqlx.In and rebind
	executeIn := func(query string, ids []uint64) error {
		q, args, err := sqlx.In(query, ids)
		if err != nil {
			return err
		}
		q = tx.Rebind(q)
		_, err = tx.ExecContext(ctx, q, args...)
		return err
	}

	_ = executeIn("DELETE FROM anime_genre WHERE anime_id IN (?)", ids)
	_ = executeIn("DELETE FROM anime_studio WHERE anime_id IN (?)", ids)
	_ = executeIn("DELETE FROM anime_producer WHERE anime_id IN (?)", ids)
	_ = executeIn("DELETE FROM external_links WHERE id IN (SELECT external_link_id FROM anime_external_link WHERE anime_id IN (?))", ids)
	_ = executeIn("DELETE FROM anime_external_link WHERE anime_id IN (?)", ids)
	_ = executeIn("DELETE FROM songs WHERE anime_id IN (?)", ids)

	if err := executeIn("DELETE FROM animes WHERE id IN (?)", ids); err != nil {
		return err
	}

	return tx.Commit()
}

// Relationships loaders
func (r *animeRepository) LoadRelations(ctx context.Context, anime *domain.Anime, isAdmin bool) error {
	// 1. Year & Season & Format
	var year domain.Year
	if err := r.db.GetContext(ctx, &year, "SELECT id, name FROM years WHERE id = $1", anime.YearID); err == nil {
		anime.Year = &year
	}
	var season domain.Season
	if err := r.db.GetContext(ctx, &season, "SELECT id, name FROM seasons WHERE id = $1", anime.SeasonID); err == nil {
		anime.Season = &season
	}
	var format domain.Format
	if err := r.db.GetContext(ctx, &format, "SELECT id, name, slug FROM formats WHERE id = $1", anime.FormatID); err == nil {
		anime.Format = &format
	}

	// 2. Many-To-Many Joins
	studiosQuery := `SELECT s.id, s.name, s.slug, s.logo, s.anime_count FROM studios s JOIN anime_studio ast ON s.id = ast.studio_id WHERE ast.anime_id = $1`
	_ = r.db.SelectContext(ctx, &anime.Studios, studiosQuery, anime.ID)

	genresQuery := `SELECT g.id, g.name, g.slug FROM genres g JOIN anime_genre ag ON g.id = ag.genre_id WHERE ag.anime_id = $1`
	_ = r.db.SelectContext(ctx, &anime.Genres, genresQuery, anime.ID)

	producersQuery := `SELECT p.id, p.name, p.slug, p.logo, p.anime_count FROM producers p JOIN anime_producer ap ON p.id = ap.producer_id WHERE ap.anime_id = $1`
	_ = r.db.SelectContext(ctx, &anime.Producers, producersQuery, anime.ID)

	// 3. External Links
	externalLinksQuery := `SELECT el.id, el.uuid, el.icon, el.type, el.name, el.url FROM external_links el JOIN anime_external_link ael ON el.id = ael.external_link_id WHERE ael.anime_id = $1`
	_ = r.db.SelectContext(ctx, &anime.ExternalLinks, externalLinksQuery, anime.ID)

	// 4. Songs belong to this anime
	songsQuery := `SELECT s.id, s.uuid, s.song_romaji, s.song_jp, s.song_en, s.theme_num, st.slug AS type, (st.slug || s.theme_num) AS slug, s.anime_id, s.season_id, s.year_id, s.views, s.created_at, s.updated_at, s.status,
		(SELECT COALESCE(AVG(rating), 0) FROM song_ratings WHERE song_id = s.id) as average_score
		FROM songs s
		LEFT JOIN song_types st ON s.type_id = st.id
		WHERE s.anime_id = $1`
	
	if !isAdmin {
		songsQuery += " AND s.status = true"
	}
	
	songsQuery += " ORDER BY s.type_id ASC, LENGTH(s.theme_num) ASC, s.theme_num ASC"

	if err := r.db.SelectContext(ctx, &anime.Songs, songsQuery, anime.ID); err != nil {
		return err
	}
	anime.SongsCount = len(anime.Songs)

	return nil
}


func (r *animeRepository) LoadManyRelations(ctx context.Context, animes []domain.Anime, isAdmin bool) error {
	if len(animes) == 0 {
		return nil
	}

	ids := make([]uint64, len(animes))
	yearIDsMap := make(map[uint64]bool)
	seasonIDsMap := make(map[uint64]bool)
	formatIDsMap := make(map[uint64]bool)

	for i, a := range animes {
		ids[i] = a.ID
		if a.YearID > 0 {
			yearIDsMap[a.YearID] = true
		}
		if a.SeasonID > 0 {
			seasonIDsMap[a.SeasonID] = true
		}
		if a.FormatID > 0 {
			formatIDsMap[a.FormatID] = true
		}
	}

	// 1. Load Year, Season & Format
	if len(yearIDsMap) > 0 {
		var yearIDs []uint64
		for id := range yearIDsMap {
			yearIDs = append(yearIDs, id)
		}
		var years []domain.Year
		q, args, _ := sqlx.In("SELECT id, name FROM years WHERE id IN (?)", yearIDs)
		if err := r.db.SelectContext(ctx, &years, r.db.Rebind(q), args...); err == nil {
			yMap := make(map[uint64]domain.Year)
			for _, y := range years {
				yMap[y.ID] = y
			}
			for i := range animes {
				if y, ok := yMap[animes[i].YearID]; ok {
					animes[i].Year = &y
				}
			}
		}
	}

	if len(seasonIDsMap) > 0 {
		var seasonIDs []uint64
		for id := range seasonIDsMap {
			seasonIDs = append(seasonIDs, id)
		}
		var seasons []domain.Season
		q, args, _ := sqlx.In("SELECT id, name FROM seasons WHERE id IN (?)", seasonIDs)
		if err := r.db.SelectContext(ctx, &seasons, r.db.Rebind(q), args...); err == nil {
			sMap := make(map[uint64]domain.Season)
			for _, s := range seasons {
				sMap[s.ID] = s
			}
			for i := range animes {
				if s, ok := sMap[animes[i].SeasonID]; ok {
					animes[i].Season = &s
				}
			}
		}
	}

	if len(formatIDsMap) > 0 {
		var formatIDs []uint64
		for id := range formatIDsMap {
			formatIDs = append(formatIDs, id)
		}
		var formats []domain.Format
		q, args, _ := sqlx.In("SELECT id, name, slug FROM formats WHERE id IN (?)", formatIDs)
		if err := r.db.SelectContext(ctx, &formats, r.db.Rebind(q), args...); err == nil {
			fMap := make(map[uint64]domain.Format)
			for _, f := range formats {
				fMap[f.ID] = f
			}
			for i := range animes {
				if f, ok := fMap[animes[i].FormatID]; ok {
					animes[i].Format = &f
				}
			}
		}
	}

	// 2. Load Many-To-Many (Genres, Studios, Producers)
	genresQuery := `SELECT g.id, g.name, g.slug, ag.anime_id FROM genres g JOIN anime_genre ag ON g.id = ag.genre_id WHERE ag.anime_id IN (?)`
	var genreRows []struct {
		ID      uint64 `db:"id"`
		Name    string `db:"name"`
		Slug    string `db:"slug"`
		AnimeID uint64 `db:"anime_id"`
	}
	gQ, gArgs, err := sqlx.In(genresQuery, ids)
	if err != nil {
		return err
	}
	if err := r.db.SelectContext(ctx, &genreRows, r.db.Rebind(gQ), gArgs...); err != nil {
		return err
	}

	gMap := make(map[uint64][]domain.Genre)
	for _, row := range genreRows {
		gMap[row.AnimeID] = append(gMap[row.AnimeID], domain.Genre{
			ID:   row.ID,
			Name: row.Name,
			Slug: row.Slug,
		})
	}

	// 2. Studios
	studiosQuery := `SELECT s.id, s.uuid, s.name, s.slug, s.logo, s.anime_count, ast.anime_id FROM studios s JOIN anime_studio ast ON s.id = ast.studio_id WHERE ast.anime_id IN (?)`
	sQ, sArgs, err := sqlx.In(studiosQuery, ids)
	if err != nil {
		return err
	}
	type studioRow struct {
		ID         uint64  `db:"id"`
		UUID       string  `db:"uuid"`
		Name       string  `db:"name"`
		Slug       string  `db:"slug"`
		Logo       *string `db:"logo"`
		AnimeCount int     `db:"anime_count"`
		AnimeID    uint64  `db:"anime_id"`
	}
	var studioRows []studioRow
	if err := r.db.SelectContext(ctx, &studioRows, r.db.Rebind(sQ), sArgs...); err != nil {
		return err
	}

	sMap := make(map[uint64][]domain.Studio)
	for _, row := range studioRows {
		sMap[row.AnimeID] = append(sMap[row.AnimeID], domain.Studio{
			ID:         row.ID,
			UUID:       row.UUID,
			Name:       row.Name,
			Slug:       row.Slug,
			LogoUrl:    row.Logo,
			AnimeCount: row.AnimeCount,
		})
	}

	// 3. Songs Count
	songsCountQuery := `SELECT anime_id, COUNT(*) as songs_count FROM songs WHERE anime_id IN (?) GROUP BY anime_id`
	scQ, scArgs, err := sqlx.In(songsCountQuery, ids)
	if err != nil {
		return err
	}
	type countRow struct {
		AnimeID    uint64 `db:"anime_id"`
		SongsCount int    `db:"songs_count"`
	}
	var countRows []countRow
	if err := r.db.SelectContext(ctx, &countRows, r.db.Rebind(scQ), scArgs...); err != nil {
		return err
	}

	scMap := make(map[uint64]int)
	for _, row := range countRows {
		scMap[row.AnimeID] = row.SongsCount
	}

	// 4. Producers
	producersQuery := `SELECT p.id, p.uuid, p.name, p.slug, p.logo, p.anime_count, apu.anime_id FROM producers p JOIN anime_producer apu ON p.id = apu.producer_id WHERE apu.anime_id IN (?)`
	pQ, pArgs, err := sqlx.In(producersQuery, ids)
	if err == nil {
		type producerRow struct {
			ID         uint64  `db:"id"`
			UUID       string  `db:"uuid"`
			Name       string  `db:"name"`
			Slug       string  `db:"slug"`
			Logo       *string `db:"logo"`
			AnimeCount int     `db:"anime_count"`
			AnimeID    uint64  `db:"anime_id"`
		}
		var producerRows []producerRow
		if err := r.db.SelectContext(ctx, &producerRows, r.db.Rebind(pQ), pArgs...); err == nil {
			pMap := make(map[uint64][]domain.Producer)
			for _, row := range producerRows {
				pMap[row.AnimeID] = append(pMap[row.AnimeID], domain.Producer{
					ID:         row.ID,
					UUID:       row.UUID,
					Name:       row.Name,
					Slug:       row.Slug,
					LogoUrl:    row.Logo,
					AnimeCount: row.AnimeCount,
				})
			}
			for i := range animes {
				animes[i].Producers = pMap[animes[i].ID]
			}
		}
	}

	// 5. Map back everything else
	for i := range animes {
		animes[i].Genres = gMap[animes[i].ID]
		animes[i].Studios = sMap[animes[i].ID]
		animes[i].SongsCount = scMap[animes[i].ID]
	}

	// 6. External Links
	linksQuery := `SELECT el.id, el.uuid, el.icon, el.type, el.name, el.url, ael.anime_id 
		FROM external_links el 
		JOIN anime_external_link ael ON el.id = ael.external_link_id 
		WHERE ael.anime_id IN (?)`
	var linkRows []struct {
		domain.ExternalLink
		AnimeID uint64 `db:"anime_id"`
	}
	lQ, lArgs, err := sqlx.In(linksQuery, ids)
	if err == nil {
		if err := r.db.SelectContext(ctx, &linkRows, r.db.Rebind(lQ), lArgs...); err == nil {
			lMap := make(map[uint64][]domain.ExternalLink)
			for _, row := range linkRows {
				lMap[row.AnimeID] = append(lMap[row.AnimeID], row.ExternalLink)
			}
			for i := range animes {
				animes[i].ExternalLinks = lMap[animes[i].ID]
			}
		}
	}

	return nil
}

func (r *animeRepository) Search(ctx context.Context, term string, limit int) ([]domain.Anime, error) {
	var animes []domain.Anime
	query := `
		SELECT id, uuid, title, slug, cover, banner, year_id, season_id, format_id 
		FROM animes 
		WHERE status = true AND (
			title ILIKE $1 OR 
			title_english ILIKE $1 OR 
			title_native ILIKE $1 OR 
			$2 = ANY(synonyms)
		)
		ORDER BY year_id DESC, season_id DESC 
		LIMIT $3
	`
	err := r.db.SelectContext(ctx, &animes, query, "%"+term+"%", term, limit)
	return animes, err
}

func (r *animeRepository) ToggleStatus(ctx context.Context, id uint64) error {
	query := "UPDATE animes SET status = NOT status WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *animeRepository) UpdateStudios(ctx context.Context, animeID uint64, studioIDs []uint64) error {
	// 0. Get old studio IDs to recount later
	var oldIDs []uint64
	_ = r.db.SelectContext(ctx, &oldIDs, "SELECT studio_id FROM anime_studio WHERE anime_id = $1", animeID)

	// 1. Delete existing
	_, err := r.db.ExecContext(ctx, "DELETE FROM anime_studio WHERE anime_id = $1", animeID)
	if err != nil {
		return err
	}

	// 2. Insert new
	for _, sID := range studioIDs {
		_, err = r.db.ExecContext(ctx, "INSERT INTO anime_studio (anime_id, studio_id) VALUES ($1, $2)", animeID, sID)
		if err != nil {
			return err
		}
	}

	// 3. Recount for affected IDs (Union of old and new)
	affectedMap := make(map[uint64]bool)
	for _, id := range oldIDs {
		affectedMap[id] = true
	}
	for _, id := range studioIDs {
		affectedMap[id] = true
	}

	for id := range affectedMap {
		_, _ = r.db.ExecContext(ctx, "UPDATE studios SET anime_count = (SELECT count(*) FROM anime_studio WHERE studio_id = $1) WHERE id = $1", id)
	}

	return nil
}

func (r *animeRepository) UpdateGenres(ctx context.Context, animeID uint64, genreIDs []uint64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM anime_genre WHERE anime_id = $1", animeID)
	if err != nil {
		return err
	}

	for _, gID := range genreIDs {
		_, err = r.db.ExecContext(ctx, "INSERT INTO anime_genre (anime_id, genre_id) VALUES ($1, $2)", animeID, gID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *animeRepository) UpdateProducers(ctx context.Context, animeID uint64, producerIDs []uint64) error {
	// 0. Get old producer IDs to recount later
	var oldIDs []uint64
	_ = r.db.SelectContext(ctx, &oldIDs, "SELECT producer_id FROM anime_producer WHERE anime_id = $1", animeID)

	// 1. Delete existing
	_, err := r.db.ExecContext(ctx, "DELETE FROM anime_producer WHERE anime_id = $1", animeID)
	if err != nil {
		return err
	}

	// 2. Insert new
	for _, pID := range producerIDs {
		_, err = r.db.ExecContext(ctx, "INSERT INTO anime_producer (anime_id, producer_id) VALUES ($1, $2)", animeID, pID)
		if err != nil {
			return err
		}
	}

	// 3. Recount for affected IDs (Union of old and new)
	affectedMap := make(map[uint64]bool)
	for _, id := range oldIDs {
		affectedMap[id] = true
	}
	for _, id := range producerIDs {
		affectedMap[id] = true
	}

	for id := range affectedMap {
		_, _ = r.db.ExecContext(ctx, "UPDATE producers SET anime_count = (SELECT count(*) FROM anime_producer WHERE producer_id = $1) WHERE id = $1", id)
	}

	return nil
}

func (r *animeRepository) UpdateExternalLinks(ctx context.Context, animeID uint64, links []domain.ExternalLink) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Delete associated `external_links` via subquery
	_, err = tx.ExecContext(ctx, "DELETE FROM external_links WHERE id IN (SELECT external_link_id FROM anime_external_link WHERE anime_id = $1)", animeID)
	if err != nil {
		return err
	}

	// 2. Delete pivot table entries
	_, err = tx.ExecContext(ctx, "DELETE FROM anime_external_link WHERE anime_id = $1", animeID)
	if err != nil {
		return err
	}

	for _, link := range links {
		linkType := link.Type
		if linkType == "" {
			linkType = "info"
		}

		query := "INSERT INTO external_links (uuid, name, url, type, created_at, updated_at) VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) RETURNING id"
		var linkID uint64
		err := tx.QueryRowContext(ctx, query, uuid.New().String(), link.Name, link.URL, linkType).Scan(&linkID)
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, "INSERT INTO anime_external_link (anime_id, external_link_id, created_at, updated_at) VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", animeID, linkID)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *animeRepository) RecountAnimeStats(ctx context.Context, ids []uint64) error {
	query := `
		UPDATE animes 
		SET 
			enabled_songs = (SELECT COUNT(*) FROM songs s WHERE s.anime_id = animes.id AND s.status = true),
			disabled_songs = (SELECT COUNT(*) FROM songs s WHERE s.anime_id = animes.id AND s.status = false),
			songs_count = (SELECT COUNT(*) FROM songs s WHERE s.anime_id = animes.id)
		WHERE ($1::bigint[] IS NULL OR id = ANY($1))
	`
	_, err := r.db.ExecContext(ctx, query, ids)
	return err
}

func (r *animeRepository) GetPublicSlugs(ctx context.Context) ([]domain.SitemapItem, error) {
	var items []domain.SitemapItem
	query := `SELECT slug as loc, updated_at as lastmod FROM animes WHERE status = true`
	err := r.db.SelectContext(ctx, &items, query)
	return items, err
}

// UpsertFromAnimeThemes inserts an anime by its anime_themes_id.
// If a record with the same anime_themes_id already exists, it is a no-op (idempotent).
// Returns (true, nil) if the record was created, (false, nil) if it already existed.
func (r *animeRepository) UpsertFromAnimeThemes(ctx context.Context, anime *domain.Anime) (bool, error) {
	anime.UUID = uuid.New().String()
	now := time.Now().UTC()

	var returnedID uint64
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO animes (uuid, title, slug, description, anilist_id, anime_themes_id, status, year_id, season_id, format_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, false, $7, $8, $9, $10, $10)
		ON CONFLICT (anime_themes_id) DO NOTHING
		RETURNING id
	`,
		anime.UUID, anime.Title, anime.Slug, anime.Description,
		anime.AnilistID, anime.AnimeThemesID,
		anime.YearID, anime.SeasonID, anime.FormatID,
		now,
	).Scan(&returnedID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// ON CONFLICT DO NOTHING — row already existed, fetch its ID
			err2 := r.db.QueryRowContext(ctx,
				`SELECT id FROM animes WHERE anime_themes_id = $1`, anime.AnimeThemesID,
			).Scan(&anime.ID)
			return false, err2
		}
		return false, err
	}
	anime.ID = returnedID
	return true, nil
}

// EnrichFromAniList updates cover, banner and description from AniList data
// only when those fields are currently empty — never overwrites existing content.
func (r *animeRepository) EnrichFromAniList(ctx context.Context, anilistID int64, cover, banner, description, titleEnglish, titleNative *string, synonyms []string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE animes SET
			cover         = COALESCE(cover, $2),
			banner        = COALESCE(banner, $3),
			description   = COALESCE(description, $4),
			title_english = COALESCE(title_english, $5),
			title_native  = COALESCE(title_native, $6),
			synonyms      = CASE WHEN synonyms = '{}' OR synonyms IS NULL THEN $7 ELSE synonyms END,
			updated_at    = CURRENT_TIMESTAMP
		WHERE anilist_id = $1
	`, anilistID, cover, banner, description, titleEnglish, titleNative, pq.Array(synonyms))
	return err
}

// buildUniqueAnimeSlug generates a slug and, on collision, appends the anime_themes_id.
func buildUniqueAnimeSlug(base string, animeThemesID uint64) string {
	s := strings.ToLower(base)
	// Replace non-alphanumeric characters with hyphens
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteRune('-')
		}
	}
	clean := strings.Trim(b.String(), "-")
	// Collapse consecutive hyphens
	for strings.Contains(clean, "--") {
		clean = strings.ReplaceAll(clean, "--", "-")
	}
	if len(clean) == 0 {
		clean = fmt.Sprintf("anime-%d", animeThemesID)
	}
	return clean
}

// GetAnimesWithMissingTitleVariants returns animes that have an anilist_id but are still
// missing title_english or title_native. Used exclusively by the backfill pipeline.
func (r *animeRepository) GetAnimesWithMissingTitleVariants(ctx context.Context, limit int, lastID uint64) ([]domain.Anime, error) {
	var animes []domain.Anime
	query := `
		SELECT * FROM animes
		WHERE anilist_id IS NOT NULL
		  AND (title_english IS NULL OR title_native IS NULL)
		  AND id > $1
		ORDER BY id ASC
		LIMIT $2`
	err := r.db.SelectContext(ctx, &animes, query, lastID, limit)
	if animes == nil {
		animes = []domain.Anime{}
	}
	return animes, err
}

// CountAnimesWithMissingTitleVariants returns the total count for GetAnimesWithMissingTitleVariants.
func (r *animeRepository) CountAnimesWithMissingTitleVariants(ctx context.Context) (int, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM animes
		WHERE anilist_id IS NOT NULL
		  AND (title_english IS NULL OR title_native IS NULL)`
	err := r.db.GetContext(ctx, &count, query)
	return count, err
}

