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

	// 2. Query against search_index
	// We use 'simple' configuration as defined in the migration.
	query := `
		SELECT 
			item_type, 
			item_id, 
			title, 
			subtitle, 
			slug, 
			image_url,
			ts_rank(search_vector, to_tsquery('simple', :query)) as rank
		FROM search_index
		WHERE search_vector @@ to_tsquery('simple', :query)
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
