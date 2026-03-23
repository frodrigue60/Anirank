package public

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"time"

	"anirank/api/internal/domain"
	"anirank/api/internal/infrastructure"
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
	mediaService    infrastructure.MediaService
	cache           domain.Cache
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
	media infrastructure.MediaService,
	appCache domain.Cache,
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
		mediaService:    media,
		cache:           appCache,
	}
}

// ─── Songs ───

func (u *CatalogUsecase) GetPaginatedSongs(ctx context.Context, userID *uint64, limit, offset int, filters domain.SongFilters) ([]domain.Song, int, error) {
	songs, err := u.songRepo.GetPaginated(ctx, limit, offset, filters)
	if err != nil {
		return nil, 0, domain.NewAppError(500, "Failed to load songs list", err)
	}

	total, _ := u.songRepo.Count(ctx, filters)

	for i := range songs {
		u.enrichSong(ctx, userID, &songs[i])
	}
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
		err := u.cache.Get(ctx, cacheKey, &lastView)
		
		if err != nil { // Cache miss (ErrCacheMiss) or error
			_ = u.songRepo.IncrementViews(ctx, song.ID)
			_ = u.cache.Set(ctx, cacheKey, time.Now(), 24*time.Hour)
		}
	}

	song.Variants, _ = u.songRepo.GetVariantsBySongID(ctx, song.ID)

	// Clean up iframe tags to get just the URL
	for i := range song.Variants {
		if song.Variants[i].Video != nil && song.Variants[i].Video.EmbedUrl != nil {
			matches := iframeSrcRegex.FindStringSubmatch(*song.Variants[i].Video.EmbedUrl)
			if len(matches) > 1 {
				song.Variants[i].Video.EmbedUrl = &matches[1]
			}
		}
	}

	related, _ := u.songRepo.GetByAnimeID(ctx, anime.ID)
	var filtered []domain.Song
	for _, s := range related {
		if s.ID != song.ID {
			s.Anime = anime
			u.enrichSong(ctx, userID, &s)
			filtered = append(filtered, s)
		}
	}
	if filtered == nil {
		filtered = []domain.Song{}
	}

	return song, filtered, nil
}

func (u *CatalogUsecase) GetSongRanking(ctx context.Context, userID *uint64, rankingType, songType string, limit, offset int) (*RankingResponse, error) {
	// v2: cache only user-agnostic enrichment; per-request user fields (rating, likes, etc.) are applied after Get.
	cacheKey := fmt.Sprintf("ranking:v2:%s:%s:%d:%d", rankingType, songType, limit, offset)

	var cachedResponse RankingResponse
	if err := u.cache.Get(ctx, cacheKey, &cachedResponse); err == nil {
		for i := range cachedResponse.Songs {
			u.enrichSong(ctx, userID, &cachedResponse.Songs[i])
		}
		return &cachedResponse, nil
	}

	songs, err := u.songRepo.GetRanking(ctx, rankingType, songType, limit, offset)
	if err != nil {
		return nil, domain.NewAppError(500, "Failed to load ranking", err)
	}

	for i := range songs {
		u.enrichSong(ctx, nil, &songs[i])
	}

	total, _ := u.songRepo.CountRanking(ctx, rankingType, songType)

	response := &RankingResponse{
		Songs: songs,
		Total: total,
	}

	if rankingType == "seasonal" {
		response.CurrentSeason, _ = u.taxonomyRepo.GetCurrentSeason(ctx)
		response.CurrentYear, _ = u.taxonomyRepo.GetCurrentYear(ctx)
	}

	_ = u.cache.Set(ctx, cacheKey, response, 5*time.Minute)

	for i := range response.Songs {
		u.enrichSong(ctx, userID, &response.Songs[i])
	}

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

	for i := range songs {
		u.enrichSong(ctx, userID, &songs[i])
	}

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

func (u *CatalogUsecase) GetUserPlaylists(ctx context.Context, requestingUserID *uint64, slug string) ([]domain.Playlist, error) {
	user, err := u.userRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, domain.NewAppError(404, "User not found", err)
	}

	includePrivate := false
	if requestingUserID != nil {
		if *requestingUserID == user.ID {
			includePrivate = true
		}
	}

	playlists, err := u.playlistRepo.GetByUserID(ctx, user.ID, includePrivate, 50, 0)
	if err != nil {
		return nil, err
	}

	for i := range playlists {
		u.enrichPlaylist(&playlists[i])
	}

	return playlists, nil
}

func (u *CatalogUsecase) GetUserFavorites(ctx context.Context, userID uint64, limit, offset int) ([]domain.Song, int, error) {
	total, err := u.songRepo.CountFavoritesByUserID(ctx, userID)
	if err != nil {
		return nil, 0, domain.NewAppError(500, "Could not count user favorites", err)
	}

	if total == 0 {
		return []domain.Song{}, 0, nil
	}

	songs, err := u.songRepo.GetFavoritesByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, domain.NewAppError(500, "Could not load user favorites", err)
	}

	uid := userID
	for i := range songs {
		u.enrichSong(ctx, &uid, &songs[i])
	}

	return songs, total, nil
}

func (u *CatalogUsecase) GetUserFavoriteArtists(ctx context.Context, userID uint64, limit, offset int) ([]domain.Artist, int, error) {
	total, err := u.artistRepo.CountFavoritesByUserID(ctx, userID)
	if err != nil {
		return nil, 0, domain.NewAppError(500, "Could not count user favorite artists", err)
	}

	if total == 0 {
		return []domain.Artist{}, 0, nil
	}

	artists, err := u.artistRepo.GetFavoritesByUserID(ctx, userID, limit, offset)
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
	var data HomeData

	// Recently Added
	recent, _ := u.songRepo.GetPaginated(ctx, 10, 0, domain.SongFilters{})
	for i := range recent {
		u.enrichSong(ctx, userID, &recent[i])
	}
	data.RecentlyAdded = recent

	// Most Popular
	popular, _ := u.songRepo.GetPaginated(ctx, 10, 0, domain.SongFilters{Sort: "favorites"})
	for i := range popular {
		u.enrichSong(ctx, userID, &popular[i])
	}
	data.MostPopular = popular

	// Most Viewed
	viewed, _ := u.songRepo.GetPaginated(ctx, 10, 0, domain.SongFilters{Sort: "views"})
	for i := range viewed {
		u.enrichSong(ctx, userID, &viewed[i])
	}
	data.MostViewed = viewed

	// Featured Song (First from popular)
	if len(popular) > 0 {
		data.FeaturedSong = &popular[0]
	}

	// Weakly Ranking (OP/ED split from ranking)
	ranking, _ := u.songRepo.GetRanking(ctx, "global", "all", 20, 0)
	data.WeaklyRanking.OP = []domain.Song{}
	data.WeaklyRanking.ED = []domain.Song{}
	for i := range ranking {
		u.enrichSong(ctx, userID, &ranking[i])
		if ranking[i].Type == "OP" && len(data.WeaklyRanking.OP) < 3 {
			data.WeaklyRanking.OP = append(data.WeaklyRanking.OP, ranking[i])
		} else if ranking[i].Type == "ED" && len(data.WeaklyRanking.ED) < 3 {
			data.WeaklyRanking.ED = append(data.WeaklyRanking.ED, ranking[i])
		}
	}

	// Featured Artists
	artists, err := u.artistRepo.GetFeatured(ctx, 5)
	if err != nil {
		fmt.Printf("[Catalog] Error fetching featured artists: %v\n", err)
	}
	for i := range artists {
		artists[i].AvatarUrl = u.mediaService.Resolve(artists[i].Avatar)
	}
	data.FeaturedArtists = artists

	return &data, nil
}

// ─── Helpers ───

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
	var allItems []domain.SitemapItem

	// 1. Animes (High Priority)
	animes, err := u.animeRepo.GetPublicSlugs(ctx)
	if err == nil {
		for i := range animes {
			animes[i].Loc = "/anime/" + animes[i].Loc
			animes[i].Priority = 0.8
			animes[i].ChangeFreq = "weekly"
		}
		allItems = append(allItems, animes...)
	}

	// 2. Songs (Medium Priority)
	songs, err := u.songRepo.GetPublicSlugs(ctx)
	if err == nil {
		for i := range songs {
			// Loc in DB is slug (e.g. chainsaw-man/kick-back)
			// But wait, song slug in DB might already be full or just song slug.
			// Let's assume catalog logic uses /song/ANIME_SLUG/SONG_SLUG
			// Actually, in GetSongByAnimeSongSlug it uses anime_slug/song_slug
			songs[i].Loc = "/song/" + songs[i].Loc
			songs[i].Priority = 0.7
			songs[i].ChangeFreq = "monthly"
		}
		allItems = append(allItems, songs...)
	}

	// 3. Artists (Medium Priority)
	artists, err := u.artistRepo.GetPublicSlugs(ctx)
	if err == nil {
		for i := range artists {
			artists[i].Loc = "/artist/" + artists[i].Loc
			artists[i].Priority = 0.6
			artists[i].ChangeFreq = "monthly"
		}
		allItems = append(allItems, artists...)
	}

	return allItems, nil
}
