package public

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"time"

	"anirank/api/internal/domain"
	"anirank/api/internal/infrastructure"
	"anirank/api/internal/infrastructure/anilist"
)

var iframeSrcRegex = regexp.MustCompile(`src="([^"]+)"`)

type CatalogUsecase struct {
	animeRepo       domain.AnimeRepository
	songRepo        domain.SongRepository
	artistRepo      domain.ArtistRepository
	taxonomyRepo    domain.TaxonomyRepository
	userRepo        domain.UserRepository
	playlistRepo    domain.PlaylistRepository
	interactionRepo domain.InteractionRepository
	moderationRepo  domain.ModerationRepository
	anilistClient   anilist.AnilistClient
	mediaService    infrastructure.MediaService
	cache           domain.Cache
	encryptionKey   string
}

type RankingResponse struct {
	Songs         []domain.Song  `json:"songs"`
	Total         int            `json:"total"`
	CurrentSeason *domain.Season `json:"current_season,omitempty"`
	CurrentYear   *domain.Year   `json:"current_year,omitempty"`
}

func NewCatalogUsecase(
	ar domain.AnimeRepository,
	sr domain.SongRepository,
	artistR domain.ArtistRepository,
	tr domain.TaxonomyRepository,
	ur domain.UserRepository,
	plr domain.PlaylistRepository,
	ir domain.InteractionRepository,
	mr domain.ModerationRepository,
	anilist anilist.AnilistClient,
	media infrastructure.MediaService,
	appCache domain.Cache,
	encKey string,
) *CatalogUsecase {
	return &CatalogUsecase{
		animeRepo:       ar,
		songRepo:        sr,
		artistRepo:      artistR,
		taxonomyRepo:    tr,
		userRepo:        ur,
		playlistRepo:    plr,
		interactionRepo: ir,
		moderationRepo:  mr,
		anilistClient:   anilist,
		mediaService:    media,
		cache:           appCache,
		encryptionKey:   encKey,
	}
}

// ─── Songs ───

func (u *CatalogUsecase) enrichSongsBulk(ctx context.Context, userID *uint64, songs []domain.Song) {
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

	// 1. Bulk fetch Relations
	artistsMap, _ := u.songRepo.GetArtistsBySongIDs(ctx, songIDs)
	variantsMap, _ := u.songRepo.GetVariantsBySongIDs(ctx, songIDs)
	ratingsMap, _ := u.interactionRepo.GetAverageRatingsBySongIDs(ctx, songIDs)

	// 2. Bulk fetch Animes if needed
	var animeMap map[uint64]domain.Anime
	if len(animeIDsMap) > 0 {
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

	// 3. User interactions & Moderation
	var userInteractions map[uint64]domain.UserSongInteraction
	var reportedMap map[uint64]bool
	if userID != nil {
		userInteractions, _ = u.interactionRepo.GetUserInteractionsBySongIDs(ctx, *userID, songIDs)
		reportedMap, _ = u.moderationRepo.GetSongReportsByUserAndSongIDs(ctx, *userID, songIDs)
	}

	// 4. Map everything back
	for i := range songs {
		s := &songs[i]

		// Anime
		if s.Anime == nil && animeMap != nil {
			if a, ok := animeMap[s.AnimeID]; ok {
				s.Anime = &a
			}
		}
		if s.Anime != nil {
			s.Anime.CoverUrl = u.mediaService.Resolve(s.Anime.Cover)
			s.Anime.BannerUrl = u.mediaService.Resolve(s.Anime.Banner)
		}

		// Artists
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

		// Variants & Embed cleanup
		if len(s.Variants) == 0 {
			if variants, ok := variantsMap[s.ID]; ok {
				for j := range variants {
					if variants[j].Video != nil && variants[j].Video.EmbedUrl != nil {
						matches := iframeSrcRegex.FindStringSubmatch(*variants[j].Video.EmbedUrl)
						if len(matches) > 1 {
							variants[j].Video.EmbedUrl = &matches[1]
						}
					}
				}
				s.Variants = variants
			} else {
				s.Variants = []domain.SongVariant{}
			}
		}

		// Ratings
		if avg, ok := ratingsMap[s.ID]; ok {
			s.AverageRating = avg
		}

		// User specific
		if userInteractions != nil {
			if inter, ok := userInteractions[s.ID]; ok {
				s.IsFavorited = inter.IsFavorited
				s.IsLiked = inter.Reaction == 1
				s.IsDisliked = inter.Reaction == -1
				s.UserRating = inter.Rating
			}
		}

		// Moderation (Bulk)
		if reportedMap != nil {
			s.IsReported = reportedMap[s.ID]
		}

		// Computed names (Consistent with enrichSong)
		if s.SongRomaji != nil && *s.SongRomaji != "" {
			s.Name = *s.SongRomaji
		} else if s.SongEN != nil && *s.SongEN != "" {
			s.Name = *s.SongEN
		} else if s.SongJP != nil && *s.SongJP != "" {
			s.Name = *s.SongJP
		} else {
			s.Name = "N/A"
		}

		// Type Name
		switch s.Type {
		case "OP":
			s.TypeName = "Opening"
		case "ED":
			s.TypeName = "Ending"
		case "INS":
			s.TypeName = "Insert"
		default:
			if s.SongType != nil && s.SongType.Name != nil {
				s.TypeName = *s.SongType.Name
			} else {
				s.TypeName = "Other"
			}
		}
	}
}

// ─── Songs ───

func (u *CatalogUsecase) GetPaginatedSongs(ctx context.Context, userID *uint64, limit, offset int, filters domain.SongFilters) ([]domain.Song, int, error) {
	songs, err := u.songRepo.GetPaginated(ctx, limit, offset, filters)
	if err != nil {
		return nil, 0, domain.NewAppError(500, "Failed to load songs list", err)
	}

	total, _ := u.songRepo.Count(ctx, filters)

	u.enrichSongsBulk(ctx, userID, songs)
	return songs, total, nil
}

func (u *CatalogUsecase) GetSongByAnimeSongSlug(ctx context.Context, userID *uint64, animeSlug, songSlug string) (*domain.Song, []domain.Song, error) {
	anime, err := u.animeRepo.GetBySlug(ctx, animeSlug)
	if err != nil {
		return nil, nil, domain.NewAppError(404, "Anime not found", err)
	}

	song, err := u.songRepo.GetByAnimeIDAndSlug(ctx, anime.ID, songSlug)
	if err != nil {
		return nil, nil, domain.NewAppError(404, "Song not found", err)
	}

	song.Anime = anime
	u.enrichSong(ctx, userID, song)

	// Increment views with cache protection
	if clientIP := ctx.Value("client_ip"); clientIP != nil {
		ipStr := clientIP.(string)
		cacheKey := fmt.Sprintf("view:%s:%d", ipStr, song.ID)

		// Use the new cache interface instead of viewCache
		var lastView time.Time
		err := u.safeCacheGet(ctx, cacheKey, &lastView)
		
		if err != nil { // Cache miss or error
			_ = u.songRepo.IncrementViews(ctx, song.ID)
			u.safeCacheSet(ctx, cacheKey, time.Now(), 24*time.Hour)
		}
	}

	// enrichment now handles variants and iframe cleanup

	related, _ := u.songRepo.GetByAnimeID(ctx, anime.ID)
	var filtered []domain.Song
	for _, s := range related {
		if s.ID != song.ID {
			s.Anime = anime
			filtered = append(filtered, s)
		}
	}
	
	if len(filtered) > 0 {
		u.enrichSongsBulk(ctx, userID, filtered)
	} else {
		filtered = []domain.Song{}
	}

	return song, filtered, nil
}

func (u *CatalogUsecase) GetSongRanking(ctx context.Context, userID *uint64, rankingType, songType string, limit, offset int) (*RankingResponse, error) {
	// v2: cache only user-agnostic enrichment; per-request user fields (rating, likes, etc.) are applied after Get.
	cacheKey := fmt.Sprintf("ranking:v2:%s:%s:%d:%d", rankingType, songType, limit, offset)

	var cachedResponse RankingResponse
	if err := u.safeCacheGet(ctx, cacheKey, &cachedResponse); err == nil {
		u.enrichSongsBulk(ctx, userID, cachedResponse.Songs)
		return &cachedResponse, nil
	}

	songs, err := u.songRepo.GetRanking(ctx, rankingType, songType, limit, offset)
	if err != nil {
		return nil, domain.NewAppError(500, "Failed to load ranking", err)
	}

	u.enrichSongsBulk(ctx, nil, songs)

	total, _ := u.songRepo.CountRanking(ctx, rankingType, songType)

	response := &RankingResponse{
		Songs: songs,
		Total: total,
	}

	if rankingType == "seasonal" {
		response.CurrentSeason, _ = u.taxonomyRepo.GetCurrentSeason(ctx)
		response.CurrentYear, _ = u.taxonomyRepo.GetCurrentYear(ctx)
	}

	u.safeCacheSet(ctx, cacheKey, response, 5*time.Minute)

	u.enrichSongsBulk(ctx, userID, response.Songs)

	return response, nil
}

// ─── Artists ───

func (u *CatalogUsecase) GetPaginatedArtists(ctx context.Context, limit, offset int, filters domain.ArtistFilters) ([]domain.Artist, int, error) {
	artists, err := u.artistRepo.GetPaginated(ctx, limit, offset, filters)
	if err != nil {
		return nil, 0, err
	}
	for i := range artists {
		u.enrichArtist(ctx, nil, &artists[i])
	}
	total, _ := u.artistRepo.Count(ctx, filters)
	return artists, total, nil
}

func (u *CatalogUsecase) GetSongsByArtistSlug(ctx context.Context, userID *uint64, slug string, limit, offset int, filters domain.SongFilters) (*domain.Artist, []domain.Song, int, error) {
	artist, err := u.artistRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, nil, 0, domain.NewAppError(404, "Artist not found", err)
	}

	if !artist.Status {
		return nil, nil, 0, domain.NewAppError(404, "Artist not found", nil)
	}

	u.enrichArtist(ctx, userID, artist)

	songs, err := u.songRepo.GetByArtistID(ctx, artist.ID, limit, offset, filters)
	if err != nil {
		return nil, nil, 0, domain.NewAppError(500, "Failed to load artist songs", err)
	}

	total, _ := u.songRepo.CountByArtistID(ctx, artist.ID, filters)

	u.enrichSongsBulk(ctx, userID, songs)

	return artist, songs, total, nil
}

// ─── Studios ───

func (u *CatalogUsecase) GetPaginatedStudios(ctx context.Context, limit, offset int, filters domain.StudioFilters) ([]domain.Studio, int, error) {
	studios, err := u.taxonomyRepo.GetPaginatedStudios(ctx, limit, offset, filters)
	if err != nil {
		return nil, 0, err
	}
	total, _ := u.taxonomyRepo.CountStudios(ctx, filters)

	for i := range studios {
		u.enrichStudio(&studios[i])
	}

	return studios, total, nil
}

func (u *CatalogUsecase) GetAnimesByStudioSlug(ctx context.Context, slug string, limit, offset int) (*domain.Studio, []domain.Anime, int, error) {
	os.Stderr.WriteString(fmt.Sprintf("[DEBUG-Usecase] Starting for slug: %s\n", slug))
	studio, err := u.taxonomyRepo.GetStudioBySlug(ctx, slug)
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("[DEBUG-Usecase] GetStudioBySlug failed: %v\n", err))
		return nil, nil, 0, domain.NewAppError(404, "Studio not found", err)
	}
	os.Stderr.WriteString(fmt.Sprintf("[DEBUG-Usecase] Found Studio ID: %d\n", studio.ID))

	animes, err := u.taxonomyRepo.GetAnimesByStudioID(ctx, studio.ID, limit, offset)
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("[CRITICAL-DEBUG] GetAnimesByStudioID failed: %v\n", err))
		return nil, nil, 0, domain.NewAppError(500, "Failed to load studio animes", err)
	}
	os.Stderr.WriteString(fmt.Sprintf("[DEBUG-Usecase] Found Animes count: %d\n", len(animes)))

	total, err := u.taxonomyRepo.CountAnimesByStudioID(ctx, studio.ID)
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("[DEBUG-Usecase] CountAnimesByStudioID failed: %v\n", err))
	}
	os.Stderr.WriteString(fmt.Sprintf("[DEBUG-Usecase] Total count: %d\n", total))

	for i := range animes {
		u.enrichAnime(ctx, &animes[i])
	}

	u.enrichStudio(studio)
	return studio, animes, total, nil
}

// ─── Producers ───

func (u *CatalogUsecase) GetPaginatedProducers(ctx context.Context, limit, offset int, filters domain.ProducerFilters) ([]domain.Producer, int, error) {
	producers, err := u.taxonomyRepo.GetPaginatedProducers(ctx, limit, offset, filters)
	if err != nil {
		return nil, 0, err
	}
	total, _ := u.taxonomyRepo.CountProducers(ctx, filters)

	for i := range producers {
		u.enrichProducer(&producers[i])
	}

	return producers, total, nil
}

func (u *CatalogUsecase) GetAnimesByProducerSlug(ctx context.Context, slug string, limit, offset int) (*domain.Producer, []domain.Anime, int, error) {
	producer, err := u.taxonomyRepo.GetProducerBySlug(ctx, slug)
	if err != nil {
		return nil, nil, 0, domain.NewAppError(404, "Producer not found", err)
	}

	animes, err := u.taxonomyRepo.GetAnimesByProducerID(ctx, producer.ID, limit, offset)
	if err != nil {
		return nil, nil, 0, domain.NewAppError(500, "Failed to load producer animes", err)
	}

	total, _ := u.taxonomyRepo.CountAnimesByProducerID(ctx, producer.ID)

	for i := range animes {
		u.enrichAnime(ctx, &animes[i])
	}

	u.enrichProducer(producer)
	return producer, animes, total, nil
}

// ─── Playlists ───

func (u *CatalogUsecase) GetPaginatedPlaylists(ctx context.Context, limit, offset int, filters domain.PlaylistFilters) ([]domain.Playlist, int, error) {
	playlists, err := u.playlistRepo.GetPaginatedPublicPlaylists(ctx, limit, offset, filters)
	if err != nil {
		return nil, 0, err
	}
	total, _ := u.playlistRepo.CountPublicPlaylists(ctx, filters)

	for i := range playlists {
		u.enrichPlaylist(&playlists[i])
	}

	return playlists, total, nil
}

// ─── Users ───

func (u *CatalogUsecase) GetUserBySlug(ctx context.Context, requestingUserID *uint64, slug string) (*domain.User, error) {
	user, err := u.userRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, domain.NewAppError(404, "User not found", err)
	}
	user.Password = ""

	u.enrichUserProfile(ctx, user)

	if requestingUserID != nil {
		isFollowing, _ := u.userRepo.IsFollowing(ctx, *requestingUserID, user.ID)
		user.IsFollowing = isFollowing
	}

	return user, nil
}

func (u *CatalogUsecase) GetUserPlaylists(ctx context.Context, requestingUserID *uint64, slug string, limit, offset int) ([]domain.Playlist, int, error) {
	user, err := u.userRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, 0, domain.NewAppError(404, "User not found", err)
	}

	includePrivate := false
	if requestingUserID != nil {
		if *requestingUserID == user.ID {
			includePrivate = true
		}
	}

	playlists, err := u.playlistRepo.GetByUserID(ctx, user.ID, includePrivate, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, _ := u.playlistRepo.CountByUserID(ctx, user.ID, includePrivate)

	for i := range playlists {
		u.enrichPlaylist(&playlists[i])
	}

	return playlists, total, nil
}

func (u *CatalogUsecase) GetUserFavorites(ctx context.Context, userID string, limit, offset int) ([]domain.Song, int, error) {
	if userID == "" {
		return []domain.Song{}, 0, nil
	}
	user, err := u.userRepo.GetByUUID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	internalID := user.ID

	total, err := u.songRepo.CountFavoritesByUserID(ctx, internalID)
	if err != nil {
		return nil, 0, domain.NewAppError(500, "Could not count user favorites", err)
	}

	if total == 0 {
		return []domain.Song{}, 0, nil
	}

	songs, err := u.songRepo.GetFavoritesByUserID(ctx, internalID, limit, offset)
	if err != nil {
		return nil, 0, domain.NewAppError(500, "Could not load user favorites", err)
	}

	uid := &internalID
	u.enrichSongsBulk(ctx, uid, songs)

	return songs, total, nil
}

func (u *CatalogUsecase) GetUserFavoriteArtists(ctx context.Context, userID string, limit, offset int) ([]domain.Artist, int, error) {
	if userID == "" {
		return []domain.Artist{}, 0, nil
	}
	user, err := u.userRepo.GetByUUID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	internalID := user.ID
	total, err := u.artistRepo.CountFavoritesByUserID(ctx, internalID)
	if err != nil {
		return nil, 0, domain.NewAppError(500, "Could not count user favorite artists", err)
	}

	if total == 0 {
		return []domain.Artist{}, 0, nil
	}

	artists, err := u.artistRepo.GetFavoritesByUserID(ctx, internalID, limit, offset)
	if err != nil {
		return nil, 0, domain.NewAppError(500, "Could not load user favorite artists", err)
	}

	for i := range artists {
		u.enrichArtist(ctx, nil, &artists[i])
	}

	return artists, total, nil
}

func (u *CatalogUsecase) GetUserFollowers(ctx context.Context, slug string, limit, offset int) ([]domain.User, int, error) {
	user, err := u.userRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, 0, domain.NewAppError(404, "User not found", err)
	}

	followers, err := u.userRepo.GetFollowers(ctx, user.ID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	for i := range followers {
		u.enrichUserProfile(ctx, &followers[i])
	}

	total, _ := u.userRepo.GetFollowersCount(ctx, user.ID)
	return followers, total, nil
}

func (u *CatalogUsecase) GetUserFollowing(ctx context.Context, slug string, limit, offset int) ([]domain.User, int, error) {
	user, err := u.userRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, 0, domain.NewAppError(404, "User not found", err)
	}

	following, err := u.userRepo.GetFollowing(ctx, user.ID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	for i := range following {
		u.enrichUserProfile(ctx, &following[i])
	}

	total, _ := u.userRepo.GetFollowingCount(ctx, user.ID)
	return following, total, nil
}

func (u *CatalogUsecase) GetUserRanking(ctx context.Context, sortBy string, limit, offset int) ([]domain.RankingUser, int, error) {
	users, total, err := u.userRepo.GetRanking(ctx, sortBy, limit, offset)
	if err != nil {
		return nil, 0, domain.NewAppError(500, "Failed to load user ranking", err)
	}

	for i := range users {
		u.enrichUserProfile(ctx, &users[i].User)
	}

	return users, total, nil
}

type AnilistHybridItem struct {
	AnilistID  int    `json:"anilist_id"`
	Status     string `json:"status"`
	Title      string `json:"title"`
	CoverImage string `json:"cover_image"`
	IsInDB     bool   `json:"is_in_db"`
	AnimeSlug  string `json:"anime_slug,omitempty"`
	Season     string `json:"season,omitempty"`
	SeasonYear int    `json:"season_year,omitempty"`
	Format     string `json:"format,omitempty"`
}

type AnilistListCacheResponse struct {
	Items []AnilistHybridItem `json:"items"`
	Total int                 `json:"total"`
}

func (u *CatalogUsecase) GetUserAnilistList(ctx context.Context, slug string, status string, page, limit int) ([]AnilistHybridItem, int, error) {
	// First check cache
	cacheKey := fmt.Sprintf("anilist:list:%s:%s:%d:%d", slug, status, page, limit)
	var cached AnilistListCacheResponse
	if err := u.safeCacheGet(ctx, cacheKey, &cached); err == nil {
		return cached.Items, cached.Total, nil
	}

	user, err := u.userRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, 0, domain.NewAppError(404, "User not found", err)
	}

	anilistID := user.GetSocialID("anilist")
	if anilistID == nil {
		return nil, 0, domain.NewAppError(400, "User has not linked AniList", nil)
	}

	// Fetch from AniList
	var alID int64
	fmt.Sscanf(*anilistID, "%d", &alID)
	anilistResp, err := u.anilistClient.GetUserMediaList(ctx, alID, status, page, limit)
	if err != nil {
		return nil, 0, domain.NewAppError(500, "Failed to fetch AniList data", err)
	}

	total := anilistResp.Data.Page.PageInfo.Total
	items := make([]AnilistHybridItem, 0, len(anilistResp.Data.Page.MediaList))

	// Collect candidate AniList IDs to check in local DB
	anilistIDs := make([]int, 0, len(anilistResp.Data.Page.MediaList))
	for _, entry := range anilistResp.Data.Page.MediaList {
		anilistIDs = append(anilistIDs, entry.Media.ID)
	}

	// Bulk check local DB
	localAnimes, _ := u.animeRepo.GetByAnilistIDs(ctx, anilistIDs)
	localMap := make(map[int]domain.Anime)
	for _, a := range localAnimes {
		if a.AnilistID != nil {
			localMap[int(*a.AnilistID)] = a
		}
	}

	for _, entry := range anilistResp.Data.Page.MediaList {
		item := AnilistHybridItem{
			AnilistID:  entry.Media.ID,
			Status:     entry.Status,
			Title:      entry.Media.Title.Romaji,
			CoverImage: entry.Media.CoverImage.ExtraLarge,
			Season:     entry.Media.Season,
			SeasonYear: entry.Media.SeasonYear,
			Format:     entry.Media.Format,
		}
		if item.Title == "" {
			item.Title = entry.Media.Title.English
		}

		if local, ok := localMap[entry.Media.ID]; ok {
			item.IsInDB = true
			item.AnimeSlug = local.Slug
		}

		items = append(items, item)
	}

	// Store in cache for 5 minutes
	u.safeCacheSet(ctx, cacheKey, AnilistListCacheResponse{
		Items: items,
		Total: total,
	}, 5*time.Minute)

	return items, total, nil
}

func (u *CatalogUsecase) enrichUserProfile(ctx context.Context, user *domain.User) {
	user.AvatarUrl = u.mediaService.Resolve(user.Avatar)
	user.BannerUrl = u.mediaService.Resolve(user.Banner)

	// Load Badges
	badges, err := u.userRepo.GetBadgesByUserID(ctx, user.ID)
	if err == nil {
		for i := range badges {
			badges[i].IconUrl = u.mediaService.Resolve(badges[i].Icon)
		}
		user.Badges = badges
	}
}

func (u *CatalogUsecase) enrichArtist(ctx context.Context, userID *uint64, artist *domain.Artist) {
	artist.AvatarUrl = u.mediaService.Resolve(artist.Avatar)
	artist.LatestBannerUrl = u.mediaService.Resolve(artist.LatestBanner)

	if userID != nil && u.interactionRepo != nil {
		isFav, _ := u.interactionRepo.IsFavoritedByUser(ctx, *userID, artist.ID, "artist")
		artist.IsFavorited = isFav
	}
}

// ─── Home ───

type HomeData struct {
	FeaturedSong    *domain.Song    `json:"featured_song"`
	WeaklyRanking   WeaklyRanking   `json:"weakly_ranking"`
	RecentlyAdded   []domain.Song   `json:"recently_added"`
	MostPopular     []domain.Song   `json:"most_popular"`
	MostViewed      []domain.Song   `json:"most_viewed"`
	FeaturedArtists []domain.Artist `json:"featured_artists"`
}

type WeaklyRanking struct {
	OP []domain.Song `json:"op"`
	ED []domain.Song `json:"ed"`
}

func (u *CatalogUsecase) GetHomeData(ctx context.Context, userID *uint64) (*HomeData, error) {
	cacheKey := "home_data"
	var data HomeData

	// 1. Try global cache first
	if err := u.safeCacheGet(ctx, cacheKey, &data); err == nil {
		// Cache hit! Inject personalized state for the requester if logged in
		if userID != nil {
			u.injectUserInteractions(ctx, *userID, &data)
		}
		return &data, nil
	}

	// 2. Cache miss -> Fetch all candidate songs (without enrichment yet)
	recent, _ := u.songRepo.GetPaginated(ctx, 10, 0, domain.SongFilters{})
	popular, _ := u.songRepo.GetPaginated(ctx, 10, 0, domain.SongFilters{Sort: "favorites"})
	viewed, _ := u.songRepo.GetPaginated(ctx, 10, 0, domain.SongFilters{Sort: "views"})
	ranking, _ := u.songRepo.GetRanking(ctx, "global", "all", 20, 0)

	// 3. Combine all for single bulk enrichment (Excluding user-specific data for caching)
	allSongs := make([]domain.Song, 0, len(recent)+len(popular)+len(viewed)+len(ranking))
	allSongs = append(allSongs, recent...)
	allSongs = append(allSongs, popular...)
	allSongs = append(allSongs, viewed...)
	allSongs = append(allSongs, ranking...)

	// 4. Perform bulk enrichment (Global data only)
	u.enrichSongsBulk(ctx, nil, allSongs)

	// 5. Map enriched songs back to sections using slice bounds
	cursor := 0
	data.RecentlyAdded = allSongs[cursor : cursor+len(recent)]
	cursor += len(recent)
	data.MostPopular = allSongs[cursor : cursor+len(popular)]
	cursor += len(popular)
	data.MostViewed = allSongs[cursor : cursor+len(viewed)]
	cursor += len(viewed)
	enrichedRanking := allSongs[cursor : cursor+len(ranking)]

	// Featured Song (First from popular)
	if len(data.MostPopular) > 0 {
		data.FeaturedSong = &data.MostPopular[0]
	}

	// Weakly Ranking (OP/ED split)
	data.WeaklyRanking.OP = []domain.Song{}
	data.WeaklyRanking.ED = []domain.Song{}
	for _, s := range enrichedRanking {
		if s.Type == "OP" && len(data.WeaklyRanking.OP) < 3 {
			data.WeaklyRanking.OP = append(data.WeaklyRanking.OP, s)
		} else if s.Type == "ED" && len(data.WeaklyRanking.ED) < 3 {
			data.WeaklyRanking.ED = append(data.WeaklyRanking.ED, s)
		}
	}

	// Featured Artists
	artists, err := u.artistRepo.GetFeatured(ctx, 5)
	if err == nil {
		for i := range artists {
			artists[i].AvatarUrl = u.mediaService.Resolve(artists[i].Avatar)
		}
		data.FeaturedArtists = artists
	}

	// 6. Save to cache for 10 minutes (Shared across all guests/users)
	u.safeCacheSet(ctx, cacheKey, &data, 10*time.Minute)

	// 7. Finally, inject user specific state for the requester
	if userID != nil {
		u.injectUserInteractions(ctx, *userID, &data)
	}

	return &data, nil
}

// ─── Helpers ───

func (u *CatalogUsecase) injectUserInteractions(ctx context.Context, userID uint64, data *HomeData) {
	// Collect all song pointers from the data structure
	allPointers := make([]*domain.Song, 0, 60)

	for i := range data.RecentlyAdded {
		allPointers = append(allPointers, &data.RecentlyAdded[i])
	}
	for i := range data.MostPopular {
		allPointers = append(allPointers, &data.MostPopular[i])
	}
	for i := range data.MostViewed {
		allPointers = append(allPointers, &data.MostViewed[i])
	}
	for i := range data.WeaklyRanking.OP {
		allPointers = append(allPointers, &data.WeaklyRanking.OP[i])
	}
	for i := range data.WeaklyRanking.ED {
		allPointers = append(allPointers, &data.WeaklyRanking.ED[i])
	}
	if data.FeaturedSong != nil {
		allPointers = append(allPointers, data.FeaturedSong)
	}

	if len(allPointers) == 0 {
		return
	}

	// Deduplicate IDs for query efficiency
	uniqueIDs := make(map[uint64]bool)
	idList := make([]uint64, 0)
	for _, s := range allPointers {
		if !uniqueIDs[s.ID] {
			uniqueIDs[s.ID] = true
			idList = append(idList, s.ID)
		}
	}

	// Bulk fetch user interactions
	interactions, err := u.interactionRepo.GetUserInteractionsBySongIDs(ctx, userID, idList)
	if err != nil {
		return
	}

	// Apply interactions to all instances of the songs
	for _, s := range allPointers {
		if inter, ok := interactions[s.ID]; ok {
			s.IsFavorited = inter.IsFavorited
			s.IsLiked = inter.Reaction == 1
			s.IsDisliked = inter.Reaction == -1
			s.UserRating = inter.Rating
		}
		// Individual moderation check (can be bulked if repository adds support)
		s.IsReported, _ = u.moderationRepo.IsSongReportedByUser(ctx, userID, s.ID)
	}
}

func (u *CatalogUsecase) enrichSong(ctx context.Context, userID *uint64, s *domain.Song) {
	if s.Anime == nil {
		anime, _ := u.animeRepo.GetByID(ctx, s.AnimeID)
		if anime != nil {
			u.animeRepo.LoadRelations(ctx, anime, false)
		}
		s.Anime = anime
	}
	if len(s.Artists) == 0 {
		artists, _ := u.songRepo.GetArtistsBySongID(ctx, s.ID, false)
		for i := range artists {
			artists[i].AvatarUrl = u.mediaService.Resolve(artists[i].Avatar)
		}
		s.Artists = artists
	}

	// Load variants to populate thumbnail_url in DTOs
	if len(s.Variants) == 0 {
		variants, _ := u.songRepo.GetVariantsBySongID(ctx, s.ID)
		// Clean up iframe tags to get just the URL for each variant
		for i := range variants {
			if variants[i].Video != nil && variants[i].Video.EmbedUrl != nil {
				matches := iframeSrcRegex.FindStringSubmatch(*variants[i].Video.EmbedUrl)
				if len(matches) > 1 {
					variants[i].Video.EmbedUrl = &matches[1]
				}
			}
		}
		s.Variants = variants
	}

	// Set computed fields
	if s.SongRomaji != nil && *s.SongRomaji != "" {
		s.Name = *s.SongRomaji
	} else if s.SongEN != nil && *s.SongEN != "" {
		s.Name = *s.SongEN
	} else if s.SongJP != nil && *s.SongJP != "" {
		s.Name = *s.SongJP
	} else {
		s.Name = "N/A"
	}

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

	if s.Anime != nil {
		s.Anime.CoverUrl = u.mediaService.Resolve(s.Anime.Cover)
		s.Anime.BannerUrl = u.mediaService.Resolve(s.Anime.Banner)
	}

	// Fetch Average Rating
	if u.interactionRepo != nil {
		avg, _ := u.interactionRepo.GetAverageRating(ctx, s.ID)
		s.AverageRating = avg

		if userID != nil {
			isFav, _ := u.interactionRepo.IsFavoritedByUser(ctx, *userID, s.ID, "song")
			s.IsFavorited = isFav

			// Reaction (Like/Dislike)
			react, err := u.interactionRepo.GetReactionByUser(ctx, *userID, s.ID, domain.TypeSong)
			if err == nil && react != nil {
				s.IsLiked = react.Type == 1
				s.IsDisliked = react.Type == -1
			}
			// Check if reported
			if u.moderationRepo != nil {
				reported, _ := u.moderationRepo.IsSongReportedByUser(ctx, *userID, s.ID)
				s.IsReported = reported
			}

			// Fetch user specific rating
			userRating, err := u.interactionRepo.GetRatingByUser(ctx, *userID, s.ID)
			if err == nil && userRating != nil {
				s.UserRating = &userRating.Rating
			}
		}
	}
}

func (u *CatalogUsecase) enrichAnime(ctx context.Context, anime *domain.Anime) {
	anime.CoverUrl = u.mediaService.Resolve(anime.Cover)
	anime.BannerUrl = u.mediaService.Resolve(anime.Banner)

	for i := range anime.Studios {
		u.enrichStudio(&anime.Studios[i])
	}
	for i := range anime.Producers {
		u.enrichProducer(&anime.Producers[i])
	}
}

func (u *CatalogUsecase) enrichStudio(s *domain.Studio) {
	s.LogoUrl = u.mediaService.Resolve(s.Logo)
	s.BannerUrl = u.mediaService.Resolve(s.LatestBanner)
	// Fallback for logo if not present
	if s.LogoUrl == nil {
		s.LogoUrl = s.BannerUrl
	}
}

func (u *CatalogUsecase) enrichProducer(p *domain.Producer) {
	p.LogoUrl = u.mediaService.Resolve(p.Logo)
	p.BannerUrl = u.mediaService.Resolve(p.LatestBanner)
	// Fallback for logo if not present
	if p.LogoUrl == nil {
		p.LogoUrl = p.BannerUrl
	}
}

func (u *CatalogUsecase) enrichPlaylist(p *domain.Playlist) {
	p.BannerUrl = u.mediaService.Resolve(p.LatestBanner)
}

func (u *CatalogUsecase) GetSitemapData(ctx context.Context) ([]domain.SitemapItem, error) {
	allItems := []domain.SitemapItem{}

	// 1. Animes (High Priority)
	animes, err := u.animeRepo.GetPublicSlugs(ctx)
	if err == nil {
		for i := range animes {
			animes[i].Loc = "/animes/" + animes[i].Loc
			animes[i].Priority = 0.8
			animes[i].ChangeFreq = "weekly"
		}
		allItems = append(allItems, animes...)
	}

	// 2. Songs (Medium Priority)
	songs, err := u.songRepo.GetPublicSlugs(ctx)
	if err == nil {
		for i := range songs {
			songs[i].Loc = "/songs/" + songs[i].Loc
			songs[i].Priority = 0.7
			songs[i].ChangeFreq = "monthly"
		}
		allItems = append(allItems, songs...)
	}

	// 3. Artists (Medium Priority)
	artists, err := u.artistRepo.GetPublicSlugs(ctx)
	if err == nil {
		for i := range artists {
			artists[i].Loc = "/artists/" + artists[i].Loc
			artists[i].Priority = 0.6
			artists[i].ChangeFreq = "monthly"
		}
		allItems = append(allItems, artists...)
	}

	return allItems, nil
}

func (u *CatalogUsecase) safeCacheGet(ctx context.Context, key string, dest interface{}) error {
	if !u.cache.IsAvailable() {
		return fmt.Errorf("cache unavailable")
	}
	cacheCtx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()
	return u.cache.Get(cacheCtx, key, dest)
}

func (u *CatalogUsecase) safeCacheSet(ctx context.Context, key string, val interface{}, exp time.Duration) {
	if !u.cache.IsAvailable() {
		return
	}
	cacheCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	_ = u.cache.Set(cacheCtx, key, val, exp)
}
