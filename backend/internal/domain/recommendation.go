package domain

import (
	"context"
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
)

// Vector representa un embedding vectorial (tipo vector de pgvector)
type Vector []float32

// Scan implementa sql.Scanner para leer el formato de pgvector: "[0.12,0.34,...]"
func (v *Vector) Scan(src interface{}) error {
	if src == nil {
		*v = nil
		return nil
	}

	var source string
	switch s := src.(type) {
	case string:
		source = s
	case []byte:
		source = string(s)
	default:
		return fmt.Errorf("incompatible type for Vector: %T", src)
	}

	source = strings.Trim(source, "[]")
	if source == "" {
		*v = Vector{}
		return nil
	}

	parts := strings.Split(source, ",")
	res := make(Vector, len(parts))
	for i, part := range parts {
		val, err := strconv.ParseFloat(strings.TrimSpace(part), 32)
		if err != nil {
			return fmt.Errorf("failed to parse vector element: %w", err)
		}
		res[i] = float32(val)
	}

	*v = res
	return nil
}

// Value implementa driver.Valuer para escribir en la DB en formato pgvector: "[0.1,0.2,...]"
func (v Vector) Value() (driver.Value, error) {
	if v == nil {
		return nil, nil
	}
	strs := make([]string, len(v))
	for i, val := range v {
		strs[i] = strconv.FormatFloat(float64(val), 'f', -1, 32)
	}
	return "[" + strings.Join(strs, ",") + "]", nil
}

// Recommendation representa un item recomendado con su score de similitud
type Recommendation struct {
	SongID uint64  `db:"song_id"`
	Score  float64 `db:"score"`
}

// RecommendationRepository define las consultas de persistencia vectorial en PostgreSQL
type RecommendationRepository interface {
	GetSimilarSongsByVector(ctx context.Context, embedding Vector, excludeSongID, excludeAnimeID uint64, limit int) ([]Song, error)
	UpdateSongEmbedding(ctx context.Context, songID uint64, embedding Vector) error
	GetSongsWithoutEmbeddings(ctx context.Context, limit int) ([]Song, error)
	GetUserPreferencesVector(ctx context.Context, userID uint64) (Vector, error)
}

// RecommendationUsecase define la lógica de negocio de recomendación y personalización
type RecommendationUsecase interface {
	GetSimilarSongs(ctx context.Context, userID *uint64, songUUID string, limit int) ([]Song, error)
	GetPersonalizedRecommendations(ctx context.Context, userID uint64, limit int) ([]Song, error)
}
