package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"anirank/api/internal/domain"

	"github.com/jmoiron/sqlx"
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

func (r *animeRepository) GetPaginated(ctx context.Context, limit, offset int, filters domain.AnimeFilters) ([]domain.Anime, error) {
	animes := []domain.Anime{}
	query := "SELECT * FROM animes"
	var args []interface{}

	if !filters.IsAdmin {
		query += " WHERE status = true"
	} else {
		query += " WHERE true"
	}

	i := 1
	if filters.IsAdmin && filters.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", i)
		args = append(args, *filters.Status)
		i++
	}


	if filters.YearID != nil && *filters.YearID > 0 {
		query += fmt.Sprintf(" AND year_id = $%d", i)
		args = append(args, *filters.YearID)
		i++
	}
	if filters.SeasonID != nil && *filters.SeasonID > 0 {
		query += fmt.Sprintf(" AND season_id = $%d", i)
		args = append(args, *filters.SeasonID)
		i++
	}
	if filters.FormatID != nil && *filters.FormatID > 0 {
		query += fmt.Sprintf(" AND format_id = $%d", i)
		args = append(args, *filters.FormatID)
		i++
	}
	if filters.Search != "" {
		query += fmt.Sprintf(" AND title ILIKE $%d", i)
		args = append(args, "%"+filters.Search+"%")
		i++
	}

	// Sorting
	switch filters.Sort {
	case "title":
		query += " ORDER BY title ASC"
	case "latest":
		query += " ORDER BY created_at DESC, id DESC"
	case "most_themes":
		query += " ORDER BY songs_count DESC"
	case "least_themes":
		query += " ORDER BY songs_count ASC"
	default:
		query += " ORDER BY created_at DESC, id DESC"
	}

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", i, i+1)
	args = append(args, limit, offset)

	err := r.db.SelectContext(ctx, &animes, query, args...)
	return animes, err
}

func (r *animeRepository) Count(ctx context.Context, filters domain.AnimeFilters) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM animes"
	var args []interface{}

	if !filters.IsAdmin {
		query += " WHERE status = true"
	} else {
		query += " WHERE true"
	}

	i := 1
	if filters.IsAdmin && filters.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", i)
		args = append(args, *filters.Status)
		i++
	}


	if filters.YearID != nil && *filters.YearID > 0 {
		query += fmt.Sprintf(" AND year_id = $%d", i)
		args = append(args, *filters.YearID)
		i++
	}
	if filters.SeasonID != nil && *filters.SeasonID > 0 {
		query += fmt.Sprintf(" AND season_id = $%d", i)
		args = append(args, *filters.SeasonID)
		i++
	}
	if filters.FormatID != nil && *filters.FormatID > 0 {
		query += fmt.Sprintf(" AND format_id = $%d", i)
		args = append(args, *filters.FormatID)
		i++
	}
	if filters.Search != "" {
		query += fmt.Sprintf(" AND title ILIKE $%d", i)
		args = append(args, "%"+filters.Search+"%")
		i++
	}

	err := r.db.GetContext(ctx, &count, query, args...)
	return count, err
}

// Write Operations
func (r *animeRepository) Create(ctx context.Context, anime *domain.Anime) error {
	query := `
		INSERT INTO animes (title, slug, description, anilist_id, status, year_id, season_id, format_id, cover, banner, created_at, updated_at) 
		VALUES (:title, :slug, :description, :anilist_id, :status, :year_id, :season_id, :format_id, :cover, :banner, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
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
		    status = :status, year_id = :year_id, 
		    season_id = :season_id, format_id = :format_id, 
		    cover = :cover, banner = :banner, updated_at = CURRENT_TIMESTAMP
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
	externalLinksQuery := `SELECT el.id, el.icon, el.type, el.name, el.url FROM external_links el JOIN anime_external_link ael ON el.id = ael.external_link_id WHERE ael.anime_id = $1`
	_ = r.db.SelectContext(ctx, &anime.ExternalLinks, externalLinksQuery, anime.ID)

	// 4. Songs belong to this anime
	songsQuery := `SELECT id, song_romaji, song_jp, song_en, theme_num, type, slug, anime_id, season_id, year_id, views, created_at, updated_at, status,
		(SELECT COALESCE(AVG(rating), 0) FROM song_ratings WHERE song_id = songs.id) as average_score
		FROM songs WHERE anime_id = $1`
	
	if !isAdmin {
		songsQuery += " AND status = true"
	}
	
	songsQuery += " ORDER BY type, theme_num"

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
		domain.Genre
		AnimeID uint64 `db:"anime_id"`
	}
	gQ, gArgs, _ := sqlx.In(genresQuery, ids)
	if err := r.db.SelectContext(ctx, &genreRows, r.db.Rebind(gQ), gArgs...); err == nil {
		gMap := make(map[uint64][]domain.Genre)
		for _, row := range genreRows {
			gMap[row.AnimeID] = append(gMap[row.AnimeID], row.Genre)
		}
		for i := range animes {
			animes[i].Genres = gMap[animes[i].ID]
		}
	}

	studiosQuery := `SELECT s.id, s.name, s.slug, s.logo, s.anime_count, ast.anime_id FROM studios s JOIN anime_studio ast ON s.id = ast.studio_id WHERE ast.anime_id IN (?)`
	var studioRows []struct {
		domain.Studio
		AnimeID uint64 `db:"anime_id"`
	}
	sQ, sArgs, _ := sqlx.In(studiosQuery, ids)
	if err := r.db.SelectContext(ctx, &studioRows, r.db.Rebind(sQ), sArgs...); err == nil {
		sMap := make(map[uint64][]domain.Studio)
		for _, row := range studioRows {
			sMap[row.AnimeID] = append(sMap[row.AnimeID], row.Studio)
		}
		for i := range animes {
			animes[i].Studios = sMap[animes[i].ID]
		}
	}

	// 3. Load Songs Count
	songsCountQuery := `SELECT anime_id, COUNT(*) as count FROM songs WHERE anime_id IN (?)`
	if !isAdmin {
		songsCountQuery += " AND status = true"
	}
	songsCountQuery += " GROUP BY anime_id"

	var countRows []struct {
		AnimeID uint64 `db:"anime_id"`
		Count   int    `db:"count"`
	}
	cQ, cArgs, _ := sqlx.In(songsCountQuery, ids)
	if err := r.db.SelectContext(ctx, &countRows, r.db.Rebind(cQ), cArgs...); err == nil {
		cMap := make(map[uint64]int)
		for _, row := range countRows {
			cMap[row.AnimeID] = row.Count
		}
		for i := range animes {
			animes[i].SongsCount = cMap[animes[i].ID]
		}
	}

	return nil
}

func (r *animeRepository) Search(ctx context.Context, term string, limit int) ([]domain.Anime, error) {
	var animes []domain.Anime
	query := `
		SELECT id, title, slug, cover, banner, year_id, season_id, format_id 
		FROM animes 
		WHERE status = true AND title ILIKE $1
		ORDER BY year_id DESC, season_id DESC 
		LIMIT $2
	`
	err := r.db.SelectContext(ctx, &animes, query, term, limit)
	return animes, err
}

func (r *animeRepository) ToggleStatus(ctx context.Context, id uint64) error {
	query := "UPDATE animes SET status = NOT status WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *animeRepository) UpdateStudios(ctx context.Context, animeID uint64, studioIDs []uint64) error {
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
	_, err := r.db.ExecContext(ctx, "DELETE FROM anime_producer WHERE anime_id = $1", animeID)
	if err != nil {
		return err
	}

	for _, pID := range producerIDs {
		_, err = r.db.ExecContext(ctx, "INSERT INTO anime_producer (anime_id, producer_id) VALUES ($1, $2)", animeID, pID)
		if err != nil {
			return err
		}
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

		query := "INSERT INTO external_links (name, url, type, created_at, updated_at) VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) RETURNING id"
		var linkID uint64
		err := tx.QueryRowContext(ctx, query, link.Name, link.URL, linkType).Scan(&linkID)
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

func (r *animeRepository) GetPublicSlugs(ctx context.Context) ([]domain.SitemapItem, error) {
	var items []domain.SitemapItem
	query := `SELECT slug as loc, updated_at as lastmod FROM animes WHERE status = true`
	err := r.db.SelectContext(ctx, &items, query)
	return items, err
}
