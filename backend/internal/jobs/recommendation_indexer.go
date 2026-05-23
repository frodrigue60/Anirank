package jobs

import (
	"context"
	"fmt"
	"log"
	"strings"

	"anirank/api/internal/domain"
)

// TopGenres es el catálogo estático de los 40 géneros principales de anime para mapear en las primeras 40 dimensiones del vector.
var TopGenres = []string{
	"Action", "Adventure", "Comedy", "Drama", "Fantasy",
	"Romance", "Sci-Fi", "Slice of Life", "Supernatural", "Mystery",
	"Sports", "Music", "Historical", "Horror", "Psychological",
	"Thriller", "Mecha", "Super Power", "Martial Arts", "Demons",
	"Magic", "Military", "Harem", "School", "Seinen",
	"Shounen", "Shoujo", "Josei", "Kids", "Parody",
	"Samurai", "Space", "Vampire", "Ecchi", "Isekai",
	"Game", "Yaoi", "Yuri", "Cyberpunk", "Suspense",
}

// ProcessPendingEmbeddings busca canciones que no tengan vector asignado y genera sus embeddings en segundo plano.
func ProcessPendingEmbeddings(ctx context.Context, rr domain.RecommendationRepository, sr domain.SongRepository, ar domain.AnimeRepository, tr domain.TaxonomyRepository) error {
	// 1. Obtener canciones que no tienen embeddings asignados (lote de 100)
	songs, err := rr.GetSongsWithoutEmbeddings(ctx, 100)
	if err != nil {
		return err
	}

	if len(songs) == 0 {
		return nil
	}

	log.Printf("[RECS-INDEXER] Found %d songs without embeddings. Calculating...", len(songs))

	// 2. Pre-cargar mapeo de géneros para velocidad de O(1)
	genreMap := make(map[string]int)
	for idx, g := range TopGenres {
		genreMap[strings.ToLower(g)] = idx
	}

	for _, s := range songs {
		// Inicializar vector de 64 dimensiones vacías (0.0)
		vec := make(domain.Vector, 64)

		// 3. Obtener el anime y cargar sus relaciones para extraer los géneros y el formato
		anime, err := ar.GetByID(ctx, s.AnimeID)
		if err != nil {
			log.Printf("[RECS-INDEXER-ERR] Failed to load anime %d for song %d: %v", s.AnimeID, s.ID, err)
			continue
		}

		err = ar.LoadRelations(ctx, anime, false)
		if err != nil {
			log.Printf("[RECS-INDEXER-ERR] Failed to load relations for anime %d: %v", anime.ID, err)
		}

		// 4. Mapear Géneros (Dimensiones 0-39)
		for _, genre := range anime.Genres {
			if idx, found := genreMap[strings.ToLower(genre.Name)]; found {
				vec[idx] = 1.0
			}
		}

		// 5. Mapear Formato de Anime (Dimensiones 40-45)
		if anime.Format != nil {
			switch strings.ToUpper(anime.Format.Name) {
			case "TV":
				vec[40] = 1.0
			case "MOVIE":
				vec[41] = 1.0
			case "OVA":
				vec[42] = 1.0
			case "ONA":
				vec[43] = 1.0
			case "SPECIAL":
				vec[44] = 1.0
			case "MUSIC":
				vec[45] = 1.0
			}
		}

		// 6. Mapear Tipo de Canción (Dimensiones 50-52)
		switch strings.ToUpper(s.Type) {
		case "OP":
			vec[50] = 1.0
		case "ED":
			vec[51] = 1.0
		case "INS":
			vec[52] = 1.0
		}

		// 7. Mapear Época / Década (Dimensiones 55-58)
		// Cargar el año para saber la época
		year, err := tr.GetYearByID(ctx, s.YearID)
		if err == nil && year != nil {
			var yearVal int
			_, _ = fmt.Sscanf(year.Name, "%d", &yearVal)
			if yearVal > 0 {
				if yearVal < 2000 {
					vec[55] = 1.0
				} else if yearVal < 2010 {
					vec[56] = 1.0
				} else if yearVal < 2020 {
					vec[57] = 1.0
				} else {
					vec[58] = 1.0
				}
			}
		}

		// 8. Mapear Popularidad / Rating Promedio (Dimensión 60)
		// Normalizamos el promedio de calificaciones de 0-10 a un rango de 0.0 - 1.0
		vec[60] = float32(s.AverageRating / 10.0)

		// 9. Guardar el vector resultante en la base de datos
		err = rr.UpdateSongEmbedding(ctx, s.ID, vec)
		if err != nil {
			log.Printf("[RECS-INDEXER-ERR] Failed to save embedding for song %d: %v", s.ID, err)
		}
	}

	log.Printf("[RECS-INDEXER] Successfully processed %d embeddings.", len(songs))
	return nil
}
