package postgres

import (
	"context"
	"fmt"
	"strings"

	"anirank/api/internal/domain"

	"github.com/jmoiron/sqlx"
)

type SearchRepository struct {
	db *sqlx.DB
}

func NewSearchRepository(db *sqlx.DB) *SearchRepository {
	return &SearchRepository{db: db}
}

func (r *SearchRepository) GlobalSearch(ctx context.Context, term string, limit int) ([]domain.SearchIndexItem, error) {
	// 1. Prepare search term for partial match (luffy -> luffy:*)
	words := strings.Fields(term)
	if len(words) == 0 {
		return nil, nil
	}

	for i, word := range words {
		// Escape special characters and add partial match marker
		cleanWord := strings.ReplaceAll(word, "'", "''")
		words[i] = fmt.Sprintf("%s:*", cleanWord)
	}
	tsQuery := strings.Join(words, " & ")

	// 2. Query against search_index, excluding inactive catalog entities.
	// Defense in depth vs stale index rows (trigger sync also filters status).
	query := `
		SELECT 
			si.item_type, 
			si.item_id, 
			si.title, 
			si.subtitle, 
			si.slug, 
			si.image_url,
			ts_rank(si.search_vector, to_tsquery('simple', :query)) as rank
		FROM search_index si
		WHERE si.search_vector @@ to_tsquery('simple', :query)
		  AND (
			(si.item_type = 'anime' AND EXISTS (
				SELECT 1 FROM animes a
				WHERE a.uuid = si.item_id AND a.status = true
			))
			OR (si.item_type = 'song' AND EXISTS (
				SELECT 1 FROM songs s
				JOIN animes a ON a.id = s.anime_id
				WHERE s.uuid = si.item_id
				  AND s.status = true
				  AND a.status = true
			))
			OR (si.item_type = 'artist' AND EXISTS (
				SELECT 1 FROM artists ar
				WHERE ar.uuid = si.item_id AND ar.status = true
			))
			OR si.item_type NOT IN ('anime', 'song', 'artist')
		  )
		ORDER BY rank DESC
		LIMIT :limit
	`

	rows, err := r.db.NamedQueryContext(ctx, query, map[string]interface{}{
		"query": tsQuery,
		"limit": limit,
	})
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.SearchIndexItem
	for rows.Next() {
		var item domain.SearchIndexItem
		if err := rows.StructScan(&item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}
