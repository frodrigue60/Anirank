package postgres

import (
	"context"
	"fmt"

	"anirank/api/internal/domain"

	"github.com/jmoiron/sqlx"
)

type recommendationRepository struct {
	db *sqlx.DB
}

// NewRecommendationRepository crea una nueva instancia del repositorio de recomendaciones
func NewRecommendationRepository(db *sqlx.DB) domain.RecommendationRepository {
	return &recommendationRepository{db: db.Unsafe()}
}

// GetSimilarSongsByVector realiza una búsqueda de vecinos más cercanos utilizando similitud de coseno
func (r *recommendationRepository) GetSimilarSongsByVector(ctx context.Context, embedding domain.Vector, excludeSongID uint64, limit int) ([]domain.Song, error) {
	var songs []domain.Song
	query := fmt.Sprintf(`
		SELECT %s
		FROM songs s
		LEFT JOIN song_types st ON s.type_id = st.id
		WHERE s.status = true AND s.id != $1 AND s.embedding IS NOT NULL
		ORDER BY s.embedding <=> $2
		LIMIT $3
	`, songColumns)

	err := r.db.SelectContext(ctx, &songs, query, excludeSongID, embedding, limit)
	if err != nil {
		return nil, err
	}
	return songs, nil
}

// UpdateSongEmbedding actualiza el vector de embedding de una canción
func (r *recommendationRepository) UpdateSongEmbedding(ctx context.Context, songID uint64, embedding domain.Vector) error {
	query := `UPDATE songs SET embedding = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, embedding, songID)
	return err
}

// GetSongsWithoutEmbeddings obtiene canciones que aún no tienen vector asignado
func (r *recommendationRepository) GetSongsWithoutEmbeddings(ctx context.Context, limit int) ([]domain.Song, error) {
	var songs []domain.Song
	query := fmt.Sprintf(`
		SELECT %s
		FROM songs s
		LEFT JOIN song_types st ON s.type_id = st.id
		WHERE s.embedding IS NULL
		LIMIT $1
	`, songColumns)

	err := r.db.SelectContext(ctx, &songs, query, limit)
	if err != nil {
		return nil, err
	}
	return songs, nil
}

// GetUserPreferencesVector calcula el vector de preferencias promediando las canciones gustadas por el usuario
func (r *recommendationRepository) GetUserPreferencesVector(ctx context.Context, userID uint64) (domain.Vector, error) {
	// Obtenemos los vectores de todas las canciones que al usuario le gustan:
	// 1. Añadidas a favoritos
	// 2. Calificadas con 7.0 o más
	query := `
		SELECT DISTINCT s.embedding
		FROM songs s
		JOIN favorites f ON f.favoritable_id = s.id AND f.favoritable_type = 'song'
		WHERE f.user_id = $1 AND s.embedding IS NOT NULL
		UNION
		SELECT DISTINCT s.embedding
		FROM songs s
		JOIN ratings r ON r.song_id = s.id
		WHERE r.user_id = $1 AND r.rating >= 7.0 AND s.embedding IS NOT NULL
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vectors []domain.Vector
	for rows.Next() {
		var v domain.Vector
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		if len(v) > 0 {
			vectors = append(vectors, v)
		}
	}

	if len(vectors) == 0 {
		return nil, nil // No hay preferencias, activará Cold Start
	}

	// Promediar los vectores de forma matemática en memoria de Go
	dimensions := len(vectors[0])
	avgVector := make(domain.Vector, dimensions)
	for _, v := range vectors {
		for i := 0; i < dimensions; i++ {
			if i < len(v) {
				avgVector[i] += v[i]
			}
		}
	}

	count := float32(len(vectors))
	for i := 0; i < dimensions; i++ {
		avgVector[i] = avgVector[i] / count
	}

	return avgVector, nil
}
