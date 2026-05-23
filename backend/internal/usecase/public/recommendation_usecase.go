package public

import (
	"context"
	"fmt"
	"log"
	"time"

	"anirank/api/internal/domain"
	"anirank/api/internal/infrastructure"
)

type recommendationUsecase struct {
	recommendationRepo domain.RecommendationRepository
	songRepo           domain.SongRepository
	animeRepo          domain.AnimeRepository
	interactionRepo    domain.InteractionRepository
	moderationRepo     domain.ModerationRepository
	mediaService       infrastructure.MediaService
	cache              domain.Cache
}

// NewRecommendationUsecase crea un nuevo caso de uso para el sistema de recomendaciones
func NewRecommendationUsecase(
	rr domain.RecommendationRepository,
	sr domain.SongRepository,
	ar domain.AnimeRepository,
	ir domain.InteractionRepository,
	mr domain.ModerationRepository,
	media infrastructure.MediaService,
	appCache domain.Cache,
) domain.RecommendationUsecase {
	return &recommendationUsecase{
		recommendationRepo: rr,
		songRepo:           sr,
		animeRepo:          ar,
		interactionRepo:    ir,
		moderationRepo:     mr,
		mediaService:       media,
		cache:              appCache,
	}
}

// GetSimilarSongs obtiene canciones similares basadas en similitud coseno del embedding
func (u *recommendationUsecase) GetSimilarSongs(ctx context.Context, userID *uint64, songUUID string, limit int) ([]domain.Song, error) {
	// 1. Obtener la canción base por su UUID
	baseSong, err := u.songRepo.GetByUUID(ctx, songUUID)
	if err != nil {
		return nil, domain.NewAppError(404, "Base song not found", err)
	}

	// 2. Intentar buscar en caché
	cacheKey := fmt.Sprintf("similar:songs:%d:%d", baseSong.ID, limit)
	var cachedSongs []domain.Song
	if err := u.safeCacheGet(ctx, cacheKey, &cachedSongs); err == nil {
		if userID != nil {
			u.enrichSongsBulk(ctx, userID, cachedSongs)
		}
		return cachedSongs, nil
	}

	// 3. Si la canción no tiene embedding, caemos a una lista segura del mismo anime
	if len(baseSong.Embedding) == 0 {
		fallbackSongs, err := u.songRepo.GetByAnimeID(ctx, baseSong.AnimeID, false)
		if err != nil {
			return []domain.Song{}, nil
		}
		// Filtrar la canción base
		var filtered []domain.Song
		for _, s := range fallbackSongs {
			if s.ID != baseSong.ID {
				filtered = append(filtered, s)
			}
		}
		if len(filtered) > limit {
			filtered = filtered[:limit]
		}
		u.enrichSongsBulk(ctx, userID, filtered)
		return filtered, nil
	}

	// 4. Buscar canciones similares en DB
	songs, err := u.recommendationRepo.GetSimilarSongsByVector(ctx, baseSong.Embedding, baseSong.ID, limit)
	if err != nil {
		return nil, domain.NewAppError(500, "Failed to load similar songs", err)
	}

	// 5. Guardar en caché (sin la personalización del usuario)
	u.safeCacheSet(ctx, cacheKey, songs, 24*time.Hour)

	// 6. Hidratar relaciones (anime, artista, calificación, etc.) para el usuario solicitante
	u.enrichSongsBulk(ctx, userID, songs)
	return songs, nil
}

// GetPersonalizedRecommendations obtiene recomendaciones basadas en el perfil de interacciones del usuario
func (u *recommendationUsecase) GetPersonalizedRecommendations(ctx context.Context, userID uint64, limit int) ([]domain.Song, error) {
	// 1. Intentar buscar en la caché del usuario
	cacheKey := fmt.Sprintf("recommendations:user:%d:%d", userID, limit)
	var cachedSongs []domain.Song
	if err := u.safeCacheGet(ctx, cacheKey, &cachedSongs); err == nil {
		u.enrichSongsBulk(ctx, &userID, cachedSongs)
		return cachedSongs, nil
	}

	// 2. Obtener el vector promedio del perfil del usuario
	userVector, err := u.recommendationRepo.GetUserPreferencesVector(ctx, userID)
	if err != nil {
		return nil, domain.NewAppError(500, "Failed to load user preferences", err)
	}

	var recommendedSongs []domain.Song

	// 3. Si el vector es nulo (Cold Start - nuevo usuario), caemos a tendencias globales
	if len(userVector) == 0 {
		log.Printf("[Recommendations] Cold start triggered for user %d. Falling back to trends.", userID)
		// Traer canciones populares ordenadas por visitas y favoritos
		songs, err := u.songRepo.GetPaginated(ctx, limit, 0, domain.SongFilters{Sort: "favorites"})
		if err != nil {
			return nil, domain.NewAppError(500, "Failed to load popular songs fallback", err)
		}
		recommendedSongs = songs
	} else {
		// 4. Buscar canciones similares al vector del usuario (excluimos id=0 porque no representa canción real)
		songs, err := u.recommendationRepo.GetSimilarSongsByVector(ctx, userVector, 0, limit)
		if err != nil {
			return nil, domain.NewAppError(500, "Failed to calculate recommendations", err)
		}
		recommendedSongs = songs
	}

	// 5. Guardar en caché por 10 minutos
	u.safeCacheSet(ctx, cacheKey, recommendedSongs, 10*time.Minute)

	// 6. Hidratar relaciones en lote
	u.enrichSongsBulk(ctx, &userID, recommendedSongs)
	return recommendedSongs, nil
}

// ─── Métodos Auxiliares de Caché Resiliente ───

func (u *recommendationUsecase) safeCacheGet(ctx context.Context, key string, dest interface{}) error {
	if u.cache == nil || !u.cache.IsAvailable() {
		return fmt.Errorf("cache unavailable")
	}
	return u.cache.Get(ctx, key, dest)
}

func (u *recommendationUsecase) safeCacheSet(ctx context.Context, key string, value interface{}, exp time.Duration) {
	if u.cache != nil && u.cache.IsAvailable() {
		_ = u.cache.Set(ctx, key, value, exp)
	}
}

// ─── Batch Enrichment Pattern (N+1 Prevention) ───

func (u *recommendationUsecase) enrichSongsBulk(ctx context.Context, userID *uint64, songs []domain.Song) {
	if len(songs) == 0 {
		return
	}

	songIDs := make([]uint64, len(songs))
	animeIDsMap := make(map[uint64]bool)
	for i, s := range songs {
		songIDs[i] = s.ID
		if s.Anime == nil && s.AnimeID > 0 {
			animeIDsMap[s.AnimeID] = true
		}
	}

	// 1. Cargar relaciones en lote
	var artistsMap map[uint64][]domain.Artist
	if u.songRepo != nil {
		artistsMap, _ = u.songRepo.GetArtistsBySongIDs(ctx, songIDs)
	}
	var ratingsMap map[uint64]float64
	if u.interactionRepo != nil {
		ratingsMap, _ = u.interactionRepo.GetAverageRatingsBySongIDs(ctx, songIDs)
	}

	// 2. Cargar animes en lote
	var animeMap map[uint64]domain.Anime
	if len(animeIDsMap) > 0 && u.animeRepo != nil {
		var animeIDs []uint64
		for id := range animeIDsMap {
			animeIDs = append(animeIDs, id)
		}
		animes, _ := u.animeRepo.GetMany(ctx, animeIDs)
		animeMap = make(map[uint64]domain.Anime)
		for _, a := range animes {
			animeMap[a.ID] = a
		}
	}

	// 3. Interacciones y Moderaciones específicas de usuario
	var userInteractions map[uint64]domain.UserSongInteraction
	var reportedMap map[uint64]bool
	if userID != nil {
		if u.interactionRepo != nil {
			userInteractions, _ = u.interactionRepo.GetUserInteractionsBySongIDs(ctx, *userID, songIDs)
		}
		if u.moderationRepo != nil {
			reportedMap, _ = u.moderationRepo.GetSongReportsByUserAndSongIDs(ctx, *userID, songIDs)
		}
	}

	// 4. Mapear de regreso a la lista original
	for i := range songs {
		s := &songs[i]

		// Anime
		if animeMap != nil {
			if a, ok := animeMap[s.AnimeID]; ok {
				if s.Anime == nil || s.Anime.Banner == nil {
					animeCopy := a
					s.Anime = &animeCopy
				}
			}
		}
		if s.Anime != nil {
			s.Anime.CoverUrl = u.mediaService.Resolve(s.Anime.Cover)
			s.Anime.BannerUrl = u.mediaService.Resolve(s.Anime.Banner)
		}

		// Artistas
		if len(s.Artists) == 0 {
			if artists, ok := artistsMap[s.ID]; ok {
				for j := range artists {
					artists[j].AvatarUrl = u.mediaService.Resolve(artists[j].Avatar)
				}
				s.Artists = artists
			} else {
				s.Artists = []domain.Artist{}
			}
		}

		// Ratings
		if avg, ok := ratingsMap[s.ID]; ok {
			s.AverageRating = avg
		}

		// Interacciones de usuario
		if userInteractions != nil {
			if inter, ok := userInteractions[s.ID]; ok {
				s.IsFavorited = inter.IsFavorited
				s.IsLiked = inter.Reaction == 1
				s.IsDisliked = inter.Reaction == -1
				s.UserRating = inter.Rating
			}
		}

		// Moderación
		if reportedMap != nil {
			s.IsReported = reportedMap[s.ID]
		}

		// Computed names
		if s.SongRomaji != nil && *s.SongRomaji != "" {
			s.Name = *s.SongRomaji
		} else if s.SongEN != nil && *s.SongEN != "" {
			s.Name = *s.SongEN
		} else if s.SongJP != nil && *s.SongJP != "" {
			s.Name = *s.SongJP
		} else if s.Anime != nil && s.Anime.Title != "" {
			s.Name = s.Anime.Title
		} else {
			s.Name = "N/A"
		}

		// Tipo
		switch s.Type {
		case "OP":
			s.TypeName = "Opening"
		case "ED":
			s.TypeName = "Ending"
		case "INS":
			s.TypeName = "Insert"
		default:
			s.TypeName = "Other"
		}
	}
}
