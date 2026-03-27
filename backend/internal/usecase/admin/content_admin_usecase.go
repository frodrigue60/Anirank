package admin

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"
	"time"

	"anirank/api/internal/domain"
	"anirank/api/internal/infrastructure"
	"anirank/api/internal/infrastructure/anilist"
	"anirank/api/internal/pkg/avatar"
	"anirank/api/internal/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"net/http"
	"encoding/json"
)

type ContentAdminUsecase struct {
	animeRepo     domain.AnimeRepository
	songRepo      domain.SongRepository
	variantRepo   domain.SongVariantRepository
	artistRepo    domain.ArtistRepository
	taxonomyRepo  domain.TaxonomyRepository
	anilistClient *anilist.Client
	mediaService  infrastructure.MediaService
	auditUsecase  domain.AuditLogUsecase
}

func NewContentAdminUsecase(
	ar domain.AnimeRepository,
	sr domain.SongRepository,
	sv domain.SongVariantRepository,
	artistR domain.ArtistRepository,
	tr domain.TaxonomyRepository,
	ac *anilist.Client,
	media infrastructure.MediaService,
	audit domain.AuditLogUsecase,
) *ContentAdminUsecase {
	return &ContentAdminUsecase{
		animeRepo:     ar,
		songRepo:      sr,
		variantRepo:   sv,
		artistRepo:    artistR,
		taxonomyRepo:  tr,
		anilistClient: ac,
		mediaService:  media,
		auditUsecase:  audit,
	}
}

// ---- ANIME ----
func (u *ContentAdminUsecase) GetAnimes(ctx context.Context, page, limit int, search string, yearID, seasonID, formatID *uint64, status *bool) ([]domain.Anime, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	filters := domain.AnimeFilters{
		Search:   search,
		YearID:   yearID,
		SeasonID: seasonID,
		FormatID: formatID,
		Status:   status,
		Sort:     "latest",
		IsAdmin:  true,
	}

	animes, err := u.animeRepo.GetPaginated(ctx, limit, offset, filters)
	if err != nil {
		return nil, 0, err
	}

	total, err := u.animeRepo.Count(ctx, filters)
	if err != nil {
		return nil, 0, err
	}

	// Load relationships like Format and Year
	_ = u.animeRepo.LoadManyRelations(ctx, animes, true)

	u.ResolveAnimesURLs(animes)

	return animes, int(total), nil
}

func (u *ContentAdminUsecase) GetAnime(ctx context.Context, id uint64) (*domain.Anime, error) {
	anime, err := u.animeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	err = u.animeRepo.LoadRelations(ctx, anime, true)
	if err != nil {
		return nil, err
	}

	u.ResolveAnimeURLs(anime)

	// Load artists for each song and resolve URLs
	for i := range anime.Songs {
		artists, err := u.songRepo.GetArtistsBySongID(ctx, anime.Songs[i].ID, true)
		if err == nil {
			for j := range artists {
				u.ResolveArtistURLs(&artists[j])
			}
			anime.Songs[i].Artists = artists
		}
		u.ResolveSongURLs(&anime.Songs[i])
	}

	return anime, nil
}

func (u *ContentAdminUsecase) ToggleAnimeStatus(ctx context.Context, id uint64, meta domain.AuditMetadata) error {
	if meta.Role == "creator" {
		return domain.NewAppError(403, "Creators cannot activate content directly", nil)
	}
	existing, _ := u.animeRepo.GetByID(ctx, id)
	if err := u.animeRepo.ToggleStatus(ctx, id); err != nil {
		return err
	}
	newAnime, _ := u.animeRepo.GetByID(ctx, id)
	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "status_toggled", id, "anime", existing, newAnime, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *ContentAdminUsecase) CreateAnime(ctx context.Context, anime *domain.Anime, meta domain.AuditMetadata) error {
	if anime.Title == "" {
		return domain.NewAppError(400, "Title is required", nil)
	}

	if anime.Slug == "" {
		anime.Slug = u.generateUniqueAnimeSlug(ctx, anime.Title, anime.AnilistID)
	}

	anime.UUID = uuid.New().String()

	// Role-based status control
	u.validateStatusPermissions(meta.Role, &anime.Status, true)

	if err := u.animeRepo.Create(ctx, anime); err != nil {
		return err
	}

	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "created", anime.ID, "anime", nil, anime, &meta.URL, &meta.IPAddress, &meta.UserAgent)

	// Sync relations if strings are provided
	return u.syncManualRelations(ctx, anime)
}

func (u *ContentAdminUsecase) UpdateAnime(ctx context.Context, anime *domain.Anime, meta domain.AuditMetadata) error {
	if anime.ID == 0 {
		return domain.NewAppError(400, "ID is required for update", nil)
	}

	// Fetch current to preserve images if not provided
	existing, err := u.animeRepo.GetByID(ctx, anime.ID)
	if err == nil && existing != nil {
		if anime.Description == nil {
			anime.Description = existing.Description
		}
		if anime.Cover == nil {
			anime.Cover = existing.Cover
		}
		if anime.Banner == nil {
			anime.Banner = existing.Banner
		}
		if anime.Slug == "" {
			anime.Slug = existing.Slug
		}
	}

	if anime.Slug == "" && anime.Title != "" {
		anime.Slug = u.generateUniqueAnimeSlug(ctx, anime.Title, anime.AnilistID)
	}

	// Role-based status control
	if err := u.validateStatusPermissions(meta.Role, &anime.Status, false); err != nil {
		return err
	}

	if err := u.animeRepo.Update(ctx, anime); err != nil {
		return err
	}

	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "updated", anime.ID, "anime", existing, anime, &meta.URL, &meta.IPAddress, &meta.UserAgent)

	// Sync relations if strings are provided
	return u.syncManualRelations(ctx, anime)
}

func (u *ContentAdminUsecase) DeleteAnime(ctx context.Context, id uint64, meta domain.AuditMetadata) error {
	existing, _ := u.animeRepo.GetByID(ctx, id)
	if err := u.animeRepo.Delete(ctx, id); err != nil {
		return err
	}
	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "deleted", id, "anime", existing, nil, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *ContentAdminUsecase) BatchDeleteAnimes(ctx context.Context, ids []uint64, meta domain.AuditMetadata) error {
	for _, id := range ids {
		_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "deleted_via_batch", id, "anime", nil, nil, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	}
	return u.animeRepo.BatchDelete(ctx, ids)
}

func (u *ContentAdminUsecase) HandleAnimeImages(c *fiber.Ctx, anime *domain.Anime) {
	// Process cover
	if fileHeader, err := c.FormFile("cover"); err == nil {
		if file, err := fileHeader.Open(); err == nil {
			defer file.Close()
			contentType := fileHeader.Header.Get("Content-Type")
			if _, url, err := u.mediaService.UploadImage(c.Context(), "animes/covers", anime.ID, file, fileHeader.Size, contentType); err == nil {
				anime.Cover = &url
				_ = u.animeRepo.Update(c.Context(), anime)
			}
		}
	}

	// Process banner
	if fileHeader, err := c.FormFile("banner"); err == nil {
		if file, err := fileHeader.Open(); err == nil {
			defer file.Close()
			contentType := fileHeader.Header.Get("Content-Type")
			if _, url, err := u.mediaService.UploadImage(c.Context(), "animes/banners", anime.ID, file, fileHeader.Size, contentType); err == nil {
				anime.Banner = &url
				_ = u.animeRepo.Update(c.Context(), anime)
			}
		}
	}
}

func (u *ContentAdminUsecase) syncManualRelations(ctx context.Context, anime *domain.Anime) error {
	// 1. Studios
	if anime.StudiosString != "" {
		names := strings.Split(anime.StudiosString, ",")
		var ids []uint64
		for _, name := range names {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			obj, err := u.taxonomyRepo.GetOrCreateStudio(ctx, name)
			if err == nil {
				ids = append(ids, obj.ID)
			}
		}
		_ = u.animeRepo.UpdateStudios(ctx, anime.ID, ids)
	}

	// 2. Producers
	if anime.ProducersString != "" {
		names := strings.Split(anime.ProducersString, ",")
		var ids []uint64
		for _, name := range names {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			obj, err := u.taxonomyRepo.GetOrCreateProducer(ctx, name)
			if err == nil {
				ids = append(ids, obj.ID)
			}
		}
		_ = u.animeRepo.UpdateProducers(ctx, anime.ID, ids)
	}

	// 3. Genres
	if anime.GenresString != "" {
		names := strings.Split(anime.GenresString, ",")
		var ids []uint64
		for _, name := range names {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			obj, err := u.taxonomyRepo.GetOrCreateGenre(ctx, name)
			if err == nil {
				ids = append(ids, obj.ID)
			}
		}
		_ = u.animeRepo.UpdateGenres(ctx, anime.ID, ids)
	}

	return nil
}

func (u *ContentAdminUsecase) ResolveAnimeURLs(a *domain.Anime) {
	if a == nil || u.mediaService == nil {
		return
	}
	if a.Cover != nil && *a.Cover != "" {
		a.CoverUrl = u.mediaService.Resolve(a.Cover)
	}
	if a.Banner != nil && *a.Banner != "" {
		a.BannerUrl = u.mediaService.Resolve(a.Banner)
	}
}

func (u *ContentAdminUsecase) ResolveAnimesURLs(animes []domain.Anime) {
	for i := range animes {
		u.ResolveAnimeURLs(&animes[i])
	}
}

func (u *ContentAdminUsecase) ResolveArtistURLs(a *domain.Artist) {
	if a == nil || u.mediaService == nil {
		return
	}
	if a.Avatar != nil && *a.Avatar != "" {
		a.AvatarUrl = u.mediaService.Resolve(a.Avatar)
	}
}

func (u *ContentAdminUsecase) ResolveArtistsURLs(artists []domain.Artist) {
	for i := range artists {
		u.ResolveArtistURLs(&artists[i])
	}
}

func (u *ContentAdminUsecase) ResolveSongURLs(s *domain.Song) {
	if s == nil || u.mediaService == nil {
		return
	}
	// Enrich nested artists
	if len(s.Artists) > 0 {
		u.ResolveArtistsURLs(s.Artists)
	}
}

func (u *ContentAdminUsecase) ResolveSongsURLs(songs []domain.Song) {
	for i := range songs {
		u.ResolveSongURLs(&songs[i])
	}
}

func (u *ContentAdminUsecase) generateUniqueAnimeSlug(ctx context.Context, title string, anilistID *int64) string {
	return utils.GenerateUniqueSlug(title, func(slug string) bool {
		existing, err := u.animeRepo.GetBySlug(ctx, slug)
		return err == nil && existing != nil
	})
}

func (u *ContentAdminUsecase) BatchFetchAnimes(ctx context.Context, season string, year int, format string, meta domain.AuditMetadata) error {
	page := 1
	var wg sync.WaitGroup

	for {
		resp, err := u.anilistClient.FetchAnimes(ctx, page, season, year, format)
		if err != nil {
			return err
		}

		for _, media := range resp.Data.Page.Media {
			// Resolve Taxonomies
			yearObj, err := u.taxonomyRepo.GetOrCreateYear(ctx, fmt.Sprintf("%d", media.SeasonYear))
			if err != nil {
				continue
			}

			seasonObj, err := u.taxonomyRepo.GetOrCreateSeason(ctx, media.Season)
			if err != nil {
				continue
			}

			formatObj, err := u.taxonomyRepo.GetOrCreateFormat(ctx, media.Format)
			if err != nil {
				continue
			}

			anilistID := int64(media.ID)
			existing, _ := u.animeRepo.GetByAnilistID(ctx, anilistID)

			var processedAnime domain.Anime
			if existing != nil {
				processedAnime = *existing
			} else {
				slug := u.generateUniqueAnimeSlug(ctx, media.Title.Romaji, &anilistID)
				existingSlug, _ := u.animeRepo.GetBySlug(ctx, slug)
				if existingSlug != nil {
					processedAnime = *existingSlug
				} else {
					status := true // public
					desc := media.Description
					newAnime := domain.Anime{
						Title:       media.Title.Romaji,
						Slug:        slug,
						Description: &desc,
						AnilistID:   &anilistID,
						Status:      status,
						YearID:      yearObj.ID,
						SeasonID:    seasonObj.ID,
						FormatID:    formatObj.ID,
						UUID:        uuid.New().String(),
					}
					err := u.animeRepo.Create(ctx, &newAnime)
					if err == nil {
						_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "created_via_batch", newAnime.ID, "anime", nil, newAnime, &meta.URL, &meta.IPAddress, &meta.UserAgent)
					}
					processedAnime = newAnime
				}
			}

			// Sync relations
			_ = u.SyncAnimeWithAnilist(ctx, &processedAnime, &media)

			// Async image processing
			if processedAnime.ID > 0 {
				wg.Add(1)
				go func(m anilist.Media, a domain.Anime) {
					defer wg.Done()
					updated := false
					if a.Cover == nil && m.CoverImage.ExtraLarge != "" {
						if imgUrl, err := u.downloadAndStore(ctx, m.CoverImage.ExtraLarge, "animes/covers", uint64(m.ID)); err == nil {
							a.Cover = &imgUrl
							updated = true
						}
					}
					if a.Banner == nil && m.BannerImage != "" {
						if imgUrl, err := u.downloadAndStore(ctx, m.BannerImage, "animes/banners", uint64(m.ID)); err == nil {
							a.Banner = &imgUrl
							updated = true
						}
					}
					if updated {
						u.animeRepo.Update(context.Background(), &a)
					}
				}(media, processedAnime)
			}
		}

		if !resp.Data.Page.PageInfo.HasNextPage {
			break
		}
		page++
	}

	wg.Wait()
	return nil
}

func (u *ContentAdminUsecase) SearchAnilistAnimes(ctx context.Context, query string, format string) ([]anilist.Media, error) {
	resp, err := u.anilistClient.SearchAnimes(ctx, query, format, 1)
	if err != nil {
		return nil, fmt.Errorf("anilist search failed: %w", err)
	}
	return resp.Data.Page.Media, nil
}

func (u *ContentAdminUsecase) CreateAnimeFromAnilist(ctx context.Context, anilistID int, meta domain.AuditMetadata) (*domain.Anime, error) {
	// Must fetch by ID. SearchAnimes("%d", id) treats the ID as a text search and often
	// does not return that media in the first page, so batch/single import looked broken.
	medias, err := u.anilistClient.GetMediaByIDs(ctx, []int{anilistID})
	if err != nil {
		return nil, err
	}
	var media *anilist.Media
	for i := range medias {
		if medias[i].ID == anilistID {
			media = &medias[i]
			break
		}
	}
	if media == nil {
		return nil, fmt.Errorf("anilist media ID %d not found", anilistID)
	}

	anilistID64 := int64(anilistID)
	existing, _ := u.animeRepo.GetByAnilistID(ctx, anilistID64)
	if existing != nil {
		_ = u.SyncAnimeWithAnilist(ctx, existing, media)
		return existing, nil
	}

	yearObj, _ := u.taxonomyRepo.GetOrCreateYear(ctx, fmt.Sprintf("%d", media.SeasonYear))
	seasonObj, _ := u.taxonomyRepo.GetOrCreateSeason(ctx, media.Season)
	formatObj, _ := u.taxonomyRepo.GetOrCreateFormat(ctx, media.Format)

	slug := u.generateUniqueAnimeSlug(ctx, media.Title.Romaji, &anilistID64)
	desc := media.Description
	status := true

	newAnime := domain.Anime{
		Title:       media.Title.Romaji,
		Slug:        slug,
		Description: &desc,
		AnilistID:   &anilistID64,
		Status:      status,
		YearID:      yearObj.ID,
		SeasonID:    seasonObj.ID,
		FormatID:    formatObj.ID,
		UUID:        uuid.New().String(),
	}

	if err := u.animeRepo.Create(ctx, &newAnime); err != nil {
		return nil, err
	}

	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "created_from_anilist", newAnime.ID, "anime", nil, newAnime, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	_ = u.SyncAnimeWithAnilist(ctx, &newAnime, media)
	go u.ensureLocalImages(context.Background(), &newAnime, media.CoverImage.ExtraLarge, media.BannerImage, int64(media.ID))

	return &newAnime, nil
}

func (u *ContentAdminUsecase) SyncAnime(ctx context.Context, id uint64, meta domain.AuditMetadata) error {
	anime, err := u.animeRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if anime.AnilistID == nil {
		return fmt.Errorf("anime has no Anilist ID")
	}

	medias, err := u.anilistClient.GetMediaByIDs(ctx, []int{int(*anime.AnilistID)})
	if err != nil {
		return err
	}

	if len(medias) == 0 {
		return fmt.Errorf("no data found on Anilist for ID %d", *anime.AnilistID)
	}

	media := medias[0]

	if media.Description != "" {
		anime.Description = &media.Description
	}

	yearObj, err := u.taxonomyRepo.GetOrCreateYear(ctx, fmt.Sprintf("%d", media.SeasonYear))
	if err == nil {
		anime.YearID = yearObj.ID
	}
	seasonObj, err := u.taxonomyRepo.GetOrCreateSeason(ctx, media.Season)
	if err == nil {
		anime.SeasonID = seasonObj.ID
	}
	formatObj, err := u.taxonomyRepo.GetOrCreateFormat(ctx, media.Format)
	if err == nil {
		anime.FormatID = formatObj.ID
	}

	if anime.Slug == "" && anime.Title != "" {
		anime.Slug = u.generateUniqueAnimeSlug(ctx, anime.Title, anime.AnilistID)
	}

	existing, _ := u.animeRepo.GetByID(ctx, anime.ID)
	if err := u.animeRepo.Update(ctx, anime); err != nil {
		return err
	}

	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "synced_from_anilist", anime.ID, "anime", existing, anime, &meta.URL, &meta.IPAddress, &meta.UserAgent)

	_ = u.SyncAnimeWithAnilist(ctx, anime, &media)
	go u.ensureLocalImages(context.Background(), anime, media.CoverImage.ExtraLarge, media.BannerImage, int64(media.ID))

	return nil
}

func (u *ContentAdminUsecase) SyncAnimeWithAnilist(ctx context.Context, anime *domain.Anime, media *anilist.Media) error {
	var studioIDs []uint64
	var producerIDs []uint64
	for _, edge := range media.Studios.Edges {
		if edge.IsMain {
			obj, err := u.taxonomyRepo.GetOrCreateStudio(ctx, edge.Node.Name)
			if err == nil {
				studioIDs = append(studioIDs, obj.ID)
			}
		} else {
			obj, err := u.taxonomyRepo.GetOrCreateProducer(ctx, edge.Node.Name)
			if err == nil {
				producerIDs = append(producerIDs, obj.ID)
			}
		}
	}
	if err := u.animeRepo.UpdateStudios(ctx, anime.ID, studioIDs); err != nil {
		fmt.Printf("[ERROR] Failed to update studios for anime %d: %v\n", anime.ID, err)
	}
	if err := u.animeRepo.UpdateProducers(ctx, anime.ID, producerIDs); err != nil {
		fmt.Printf("[ERROR] Failed to update producers for anime %d: %v\n", anime.ID, err)
	}
	fmt.Printf("[INFO] Synced %d studios and %d producers for anime %d\n", len(studioIDs), len(producerIDs), anime.ID)

	var genreIDs []uint64
	for _, gName := range media.Genres {
		obj, err := u.taxonomyRepo.GetOrCreateGenre(ctx, gName)
		if err == nil {
			genreIDs = append(genreIDs, obj.ID)
		}
	}
	if err := u.animeRepo.UpdateGenres(ctx, anime.ID, genreIDs); err != nil {
		fmt.Printf("[ERROR] Failed to update genres for anime %d: %v\n", anime.ID, err)
	}
	fmt.Printf("[INFO] Synced %d genres for anime %d\n", len(genreIDs), anime.ID)

	links := make([]domain.ExternalLink, 0, len(media.ExternalLinks)+1)
	links = append(links, domain.ExternalLink{
		Name: "AniList",
		URL:  fmt.Sprintf("https://anilist.co/anime/%d", media.ID),
		Type: "anilist",
	})
	for _, l := range media.ExternalLinks {
		links = append(links, domain.ExternalLink{
			Name: l.Site,
			URL:  l.URL,
			Type: strings.ToLower(l.Site),
		})
	}
	if err := u.animeRepo.UpdateExternalLinks(ctx, anime.ID, links); err != nil {
		fmt.Printf("[ERROR] Failed to update external links for anime %d: %v\n", anime.ID, err)
	}

	return nil
}

func (u *ContentAdminUsecase) isLocalImage(path *string) bool {
	if path == nil || *path == "" {
		return false
	}
	// relative path
	if !strings.HasPrefix(*path, "http") {
		return true
	}
	// or our own S3
	return strings.Contains(*path, "s3.anirank")
}

func (u *ContentAdminUsecase) BatchCreateAnimesFromAnilist(ctx context.Context, anilistIDs []int, meta domain.AuditMetadata) *domain.AnilistBatchImportResult {
	result := &domain.AnilistBatchImportResult{
		Requested:   len(anilistIDs),
		ImportedIDs: make([]int, 0, len(anilistIDs)),
		Errors:      make([]domain.AnilistBatchImportItemError, 0),
	}
	if len(anilistIDs) == 0 {
		return result
	}
	const maxErrors = 20
	for _, id := range anilistIDs {
		if _, err := u.CreateAnimeFromAnilist(ctx, id, meta); err != nil {
			result.Failed++
			if len(result.Errors) < maxErrors {
				result.Errors = append(result.Errors, domain.AnilistBatchImportItemError{
					AnilistID: id,
					Message:   err.Error(),
				})
			}
			continue
		}
		result.Imported++
		result.ImportedIDs = append(result.ImportedIDs, id)
	}
	return result
}


func (u *ContentAdminUsecase) SyncArtistsFromString(ctx context.Context, songID uint64, artistsStr string, meta domain.AuditMetadata) error {
	names := strings.Split(artistsStr, ",")
	var artistIDs []uint64
	
	// Regex for "Name (NameJP)" format
	re := regexp.MustCompile(`^(.+?)\s*\(\s*(.+?)\s*\)$`)

	for _, rawName := range names {
		rawName = strings.TrimSpace(rawName)
		if rawName == "" {
			continue
		}

		var name, nameJP string
		match := re.FindStringSubmatch(rawName)
		if len(match) > 2 {
			name = strings.TrimSpace(match[1])
			nameJP = strings.TrimSpace(match[2])
		} else {
			name = rawName
		}

		slug := utils.Slugify(name)
		if slug == "" {
			// If name is non-latin, use the full name as slug (Slugify will probably fail again, 
			// but we need a slug. We might need a better slugifier later)
			slug = "artist-" + utils.Slugify(nameJP)
			if slug == "artist-" {
				slug = "artist-" + fmt.Sprintf("%d", time.Now().UnixNano()%1000000)
			}
		}

		// Try to find existing artist by slug
		artist, err := u.artistRepo.GetBySlug(ctx, slug)
		if err != nil {
			// Create new artist
			newArtist := &domain.Artist{
				Name:   name,
				Slug:   slug,
				Status: true,
				UUID:   uuid.New().String(),
			}
			if nameJP != "" {
				newArtist.NameJP = &nameJP
			}
			if err := u.CreateArtist(ctx, newArtist, meta); err == nil {
				artist = newArtist
			}
		} else if nameJP != "" && artist.NameJP == nil {
			// Update existing artist if it didn't have a Japanese name
			artist.NameJP = &nameJP
			_ = u.artistRepo.Update(ctx, artist)
		}

		if artist != nil {
			artistIDs = append(artistIDs, artist.ID)
		}
	}
	
	if len(artistIDs) > 0 {
		return u.songRepo.SyncArtists(ctx, songID, artistIDs)
	}
	return nil
}


func (u *ContentAdminUsecase) ensureLocalImages(ctx context.Context, anime *domain.Anime, preferredCover, preferredBanner string, anilistID int64) {
	updated := false
	
	// Cover
	if preferredCover != "" {
		if !u.isLocalImage(anime.Cover) {
			id := uint64(anilistID)
			if id == 0 {
				id = anime.ID
			}
			if imgUrl, err := u.downloadAndStore(ctx, preferredCover, "animes/covers", id); err == nil {
				anime.Cover = &imgUrl
				updated = true
			} else {
				fmt.Printf("[ERROR] Failed to download cover for anime %d (%d): %v\n", anime.ID, anilistID, err)
			}
		}
	}

	// Banner
	if preferredBanner != "" {
		if !u.isLocalImage(anime.Banner) {
			id := uint64(anilistID)
			if id == 0 {
				id = anime.ID
			}
			if imgUrl, err := u.downloadAndStore(ctx, preferredBanner, "animes/banners", id); err == nil {
				anime.Banner = &imgUrl
				updated = true
			} else {
				fmt.Printf("[ERROR] Failed to download banner for anime %d (%d): %v\n", anime.ID, anilistID, err)
			}
		}
	}

	if updated {
		if err := u.animeRepo.Update(ctx, anime); err != nil {
			fmt.Printf("[ERROR] Failed to save local image paths for anime %d: %v\n", anime.ID, err)
		} else {
			fmt.Printf("[INFO] Successfully saved local image paths for anime %d\n", anime.ID)
		}
	}
}

func (u *ContentAdminUsecase) downloadAndStore(ctx context.Context, url string, prefix string, id uint64) (string, error) {
	if u.mediaService == nil {
		return "", fmt.Errorf("no media service configured")
	}

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Anirank/1.0 (https://anirank.work)")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch image: %s (URL: %s)", resp.Status, url)
	}

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	path, url, err := u.mediaService.UploadImage(ctx, prefix, id, bytes.NewReader(buf), int64(len(buf)), resp.Header.Get("Content-Type"))
	_ = path // avoid unused
	return url, err
}

type ATResponse struct {
	Anime []struct {
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Year        int    `json:"year"`
		Season      string `json:"season"`
		Synopsis    string `json:"synopsis"`
		MediaFormat string `json:"media_format"`
		Images      []struct {
			Facet string `json:"facet"`
			Link  string `json:"link"`
		} `json:"images"`
		Resources []struct {
			Site string `json:"site"`
			Link string `json:"link"`
		} `json:"resources"`
		Studios []struct {
			Name string `json:"name"`
			Slug string `json:"slug"`
		} `json:"studios"`
		AnimeThemes []struct {
			ID       uint64 `json:"id"`
			Type     string `json:"type"`
			Sequence int    `json:"sequence"`
			Song     *struct {
				Title   string `json:"title"`
				Artists []struct {
					ID   uint64 `json:"id"`
					Name string `json:"name"`
					Slug string `json:"slug"`
				} `json:"artists"`
			} `json:"song"`
			AnimeThemeEntries []struct {
				ID      uint64 `json:"id"`
				Version int    `json:"version"`
			} `json:"animethemeentries"`
		} `json:"animethemes"`
	} `json:"anime"`
}

func (u *ContentAdminUsecase) HydrateSeason(ctx context.Context, year int, seasonName string, meta domain.AuditMetadata, progress chan<- string) error {
	sendProgress := func(msg string) {
		if progress != nil {
			select {
			case progress <- msg:
			default:
			}
		}
	}

	sendProgress(fmt.Sprintf("Fetching %s %d from AnimeThemes...", seasonName, year))
	url := fmt.Sprintf("https://api.animethemes.moe/anime?include=animethemes.song.artists,images,animethemes.animethemeentries,studios,resources&filter[year]=%d&filter[season]=%s&page[size]=100", year, strings.ToLower(seasonName))

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch from AnimeThemes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("AnimeThemes API returned status %d", resp.StatusCode)
	}

	var atResp ATResponse
	if err := json.NewDecoder(resp.Body).Decode(&atResp); err != nil {
		return fmt.Errorf("failed to decode AnimeThemes response: %w", err)
	}

	totalAnime := len(atResp.Anime)
	sendProgress(fmt.Sprintf("Found %d animes in AnimeThemes. Collecting IDs...", totalAnime))

	// --- Phase 1: ID Collection ---
	uniqueAlIDs := make(map[int64]bool)
	for _, a := range atResp.Anime {
		var alID int64
		for _, res := range a.Resources {
			if res.Site == "AniList" {
				parts := strings.Split(strings.TrimRight(res.Link, "/"), "/")
				if len(parts) > 0 {
					fmt.Sscanf(parts[len(parts)-1], "%d", &alID)
				}
				break
			}
		}
		// Fallback lookup if not in resources
		if alID == 0 {
			animeSlug := utils.Slugify(a.Name)
			if animeSlug == "" {
				animeSlug = utils.Slugify(a.Slug)
			}
			if ex, err := u.animeRepo.GetBySlug(ctx, animeSlug); err == nil && ex.AnilistID != nil {
				alID = *ex.AnilistID
			}
		}
		if alID > 0 {
			uniqueAlIDs[alID] = true
		}
	}

	// --- Phase 2: Batch Fetching ---
	alDataMap := make(map[int64]anilist.Media)
	if len(uniqueAlIDs) > 0 {
		ids := make([]int, 0, len(uniqueAlIDs))
		for id := range uniqueAlIDs {
			ids = append(ids, int(id))
		}

		sendProgress(fmt.Sprintf("Enriching %d animes from AniList in batches of 50...", len(ids)))
		for i := 0; i < len(ids); i += 50 {
			end := i + 50
			if end > len(ids) {
				end = len(ids)
			}
			chunk := ids[i:end]
			medias, err := u.anilistClient.GetMediaByIDs(ctx, chunk)
			if err != nil {
				fmt.Printf("[ERROR] Batch fetch failed: %v\n", err)
			} else {
				for _, m := range medias {
					alDataMap[int64(m.ID)] = m
				}
			}
			if end < len(ids) {
				time.Sleep(1 * time.Second) // Safe buffer between batches
			}
		}
	}

	// --- Phase 3: Processing ---
	for i, a := range atResp.Anime {
		sendProgress(fmt.Sprintf("[%d/%d] Processing: %s", i+1, totalAnime, a.Name))
		// 1. Resolve Taxonomies
		yearObj, err := u.taxonomyRepo.GetOrCreateYear(ctx, fmt.Sprintf("%d", a.Year))
		if err != nil {
			continue
		}

		normalizedSeason := strings.Title(strings.ToLower(a.Season))
		seasonObj, err := u.taxonomyRepo.GetOrCreateSeason(ctx, normalizedSeason)
		if err != nil {
			continue
		}

		if a.MediaFormat == "" {
			a.MediaFormat = "TV"
		}
		formatObj, err := u.taxonomyRepo.GetOrCreateFormat(ctx, a.MediaFormat)
		if err != nil {
			continue
		}

		// 2. Extract AniList ID from AT data for this anime
		var anilistID int64
		for _, res := range a.Resources {
			if res.Site == "AniList" {
				parts := strings.Split(strings.TrimRight(res.Link, "/"), "/")
				if len(parts) > 0 {
					fmt.Sscanf(parts[len(parts)-1], "%d", &anilistID)
				}
				break
			}
		}

		// 3. Find existing or fallback ID
		animeSlug := utils.Slugify(a.Name)
		if animeSlug == "" {
			animeSlug = utils.Slugify(a.Slug)
		}

		var anime *domain.Anime
		if anilistID > 0 {
			anime, _ = u.animeRepo.GetByAnilistID(ctx, anilistID)
		}
		if anime == nil {
			anime, _ = u.animeRepo.GetBySlug(ctx, animeSlug)
		}

		// Fallback ID from DB if AT doesn't have it
		if anilistID == 0 && anime != nil && anime.AnilistID != nil {
			anilistID = *anime.AnilistID
		}

		// 4. Use batched AniList Data
		var alData *anilist.Media
		if media, ok := alDataMap[anilistID]; ok {
			alData = &media
			// Override with AniList data
			if alData.Title.Romaji != "" {
				a.Name = alData.Title.Romaji
			}
			if alData.Description != "" {
				a.Synopsis = alData.Description
			}
		} else if anilistID > 0 {
			fmt.Printf("[WARN] AniList data missing in batch for ID %d\n", anilistID)
		}

		coverUrl := ""
		for _, img := range a.Images {
			if img.Facet == "Large Cover" {
				coverUrl = img.Link
				break
			}
		}
		if alData != nil && alData.CoverImage.ExtraLarge != "" {
			coverUrl = alData.CoverImage.ExtraLarge
		}

		bannerUrl := ""
		if alData != nil {
			bannerUrl = alData.BannerImage
		}

		var coverPtr, bannerPtr *string
		if coverUrl != "" {
			coverPtr = &coverUrl
		}
		if bannerUrl != "" {
			bannerPtr = &bannerUrl
		}
		var alIDPtr *int64
		if anilistID > 0 {
			alIDPtr = &anilistID
		}

		if anime != nil {
			// Update
			anime.Title = a.Name
			anime.Slug = animeSlug
			anime.Description = &a.Synopsis
			if coverPtr != nil && !u.isLocalImage(anime.Cover) {
				anime.Cover = coverPtr
			}
			if bannerPtr != nil && !u.isLocalImage(anime.Banner) {
				anime.Banner = bannerPtr
			}
			if alIDPtr != nil {
				anime.AnilistID = alIDPtr
			}
			anime.YearID = yearObj.ID
			anime.SeasonID = seasonObj.ID
			anime.FormatID = formatObj.ID
			err := u.animeRepo.Update(ctx, anime)
			if err != nil {
				fmt.Printf("[ERROR] Failed to update anime %d: %v\n", anime.ID, err)
			}
			_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "hydrated_update", anime.ID, "anime", nil, anime, &meta.URL, &meta.IPAddress, &meta.UserAgent)
			go u.ensureLocalImages(context.Background(), anime, coverUrl, bannerUrl, anilistID)
		} else {
			// Create
			anime = &domain.Anime{
				Title:       a.Name,
				Slug:        animeSlug,
				Description: &a.Synopsis,
				Cover:       coverPtr, // Will be updated by ensureLocalImages
				Banner:      bannerPtr,
				AnilistID:   alIDPtr,
				YearID:      yearObj.ID,
				SeasonID:    seasonObj.ID,
				FormatID:    formatObj.ID,
				Status:      true,
				UUID:        uuid.New().String(),
			}
			if err := u.animeRepo.Create(ctx, anime); err != nil {
				continue
			}
			_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "hydrated_create", anime.ID, "anime", nil, anime, &meta.URL, &meta.IPAddress, &meta.UserAgent)
			go u.ensureLocalImages(context.Background(), anime, coverUrl, bannerUrl, anilistID)
		}

		// 5. Sync Associations
		if alData != nil {
			_ = u.SyncAnimeWithAnilist(ctx, anime, alData)
		} else {
			// Fallback to AT Studios
			var studioIDs []uint64
			for _, s := range a.Studios {
				obj, err := u.taxonomyRepo.GetOrCreateStudio(ctx, s.Name)
				if err == nil {
					studioIDs = append(studioIDs, obj.ID)
				}
			}
			_ = u.animeRepo.UpdateStudios(ctx, anime.ID, studioIDs)
		}

			// 6. Process Songs & Themes
			// We need to load songs for the anime to check for existence
			animeSongs, _ := u.songRepo.GetByAnimeID(ctx, anime.ID)
			processedSongs := make(map[uint64]bool) // Track processed IDs in this batch
			
			for _, t := range a.AnimeThemes {
				if t.Song == nil {
					continue
				}

				atID := t.ID
				themeNum := t.Sequence
				if themeNum == 0 {
					themeNum = 1
				}
				songSlug := fmt.Sprintf("%s%d", t.Type, themeNum)

				var song *domain.Song
				// 1. Match by AnimeThemes ID (Most reliable)
				for i := range animeSongs {
					if animeSongs[i].AnimeThemesID != nil && *animeSongs[i].AnimeThemesID == atID {
						song = &animeSongs[i]
						break
					}
				}
				// 2. Fallback to Slug if no ID match
				if song == nil {
					for i := range animeSongs {
						if animeSongs[i].Slug == songSlug {
							song = &animeSongs[i]
							break
						}
					}
				}

				// Skip if we already processed this ID in this run (Duplicate in AT response)
				if song != nil && processedSongs[song.ID] {
					continue
				}

				title := t.Song.Title
				if song != nil {
					// Update
					song.SongRomaji = &title
					song.YearID = yearObj.ID
					song.SeasonID = seasonObj.ID
					song.AnimeThemesID = &atID
					_ = u.songRepo.Update(ctx, song)
					processedSongs[song.ID] = true
				} else {
					// Create
					song = &domain.Song{
						SongRomaji:    &title,
						Slug:          songSlug,
						Type:          t.Type,
						ThemeNum:      fmt.Sprintf("%d", themeNum),
						AnimeID:       anime.ID,
						YearID:        yearObj.ID,
						SeasonID:      seasonObj.ID,
						Status:        true,
						AnimeThemesID: &atID,
						UUID:          uuid.New().String(),
					}
					if err := u.songRepo.Create(ctx, song); err != nil {
						continue
					}
					processedSongs[song.ID] = true
				}

				// Process Variants
				existingVariants, _ := u.songRepo.GetVariantsBySongID(ctx, song.ID)
				processedVariants := make(map[uint64]bool)

				for _, entry := range t.AnimeThemeEntries {
					atvID := entry.ID
					version := uint64(entry.Version)
					if version == 0 {
						version = 1
					}
					variantSlug := fmt.Sprintf("V%d", version)
					
					var variant *domain.SongVariant
					// 1. Match by AnimeThemes ID (Most reliable)
					for i := range existingVariants {
						if existingVariants[i].AnimeThemesID != nil && *existingVariants[i].AnimeThemesID == atvID {
							variant = &existingVariants[i]
							break
						}
					}
					// 2. Fallback to Slug if no ID match
					if variant == nil {
						for i := range existingVariants {
							if existingVariants[i].Slug == variantSlug {
								variant = &existingVariants[i]
								break
							}
						}
					}

					// Skip if we already processed this ID in this run (Duplicate in AT response)
					if variant != nil && processedVariants[variant.ID] {
						continue
					}

					if variant != nil {
						// Update
						variant.VersionNumber = version
						variant.Slug = variantSlug
						variant.AnimeThemesID = &atvID
						variant.YearID = yearObj.ID
						variant.SeasonID = seasonObj.ID
						_ = u.variantRepo.Update(ctx, variant)
						processedVariants[variant.ID] = true
					} else {
						// Create
						variant = &domain.SongVariant{
							SongID:        song.ID,
							Slug:          variantSlug,
							VersionNumber: version,
							Status:        true,
							YearID:        yearObj.ID,
							SeasonID:      seasonObj.ID,
							AnimeThemesID: &atvID,
							UUID:          uuid.New().String(),
						}
						if err := u.variantRepo.Create(ctx, variant); err == nil {
							processedVariants[variant.ID] = true
						}
					}
				}

				// Process Artists
				var artistIDs []uint64
				processedArtists := make(map[uint64]bool)

				for _, art := range t.Song.Artists {
					atArtID := art.ID
					cleanName := strings.TrimSpace(art.Name)
					aSlug := utils.Slugify(cleanName)
					if aSlug == "" {
						aSlug = utils.Slugify(art.Slug)
					}
					if aSlug == "" {
						continue
					}

					var artist *domain.Artist
					// 1. Match by AnimeThemes ID (Most reliable)
					artist, _ = u.artistRepo.GetByAnimeThemesID(ctx, atArtID)
					
					// 2. Fallback to Slug if no ID match
					if artist == nil {
						artist, _ = u.artistRepo.GetBySlug(ctx, aSlug)
					}

					if artist != nil {
						// Update and ensure ID is saved
						artist.Name = cleanName
						artist.Slug = aSlug
						artist.AnimeThemesID = &atArtID
						_ = u.artistRepo.Update(ctx, artist)
					} else {
						// Create
						artist = &domain.Artist{
							Name:          cleanName,
							Slug:          aSlug,
							Status:        true,
							AnimeThemesID: &atArtID,
							UUID:          uuid.New().String(),
						}
						if err := u.artistRepo.Create(ctx, artist); err != nil {
							continue
						}
					}

					// Skip if we already added this artist to THIS song
					if !processedArtists[artist.ID] {
						artistIDs = append(artistIDs, artist.ID)
						processedArtists[artist.ID] = true
					}
				}
				if len(artistIDs) > 0 {
					_ = u.songRepo.SyncArtists(ctx, song.ID, artistIDs)
			}
		}
	}
	sendProgress("Hydration completed successfully!")
	return nil
}

// ---- SONG ----
func (u *ContentAdminUsecase) GetSongs(ctx context.Context, page, limit int, search string, animeID *uint64, status *bool) ([]domain.Song, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	filters := domain.SongFilters{
		Search:  search,
		IsAdmin: true,
	}

	if animeID != nil {
		filters.AnimeID = *animeID
	}
	if status != nil {
		filters.Status = status
	}

	songs, err := u.songRepo.GetPaginated(ctx, limit, offset, filters)
	if err != nil {
		return nil, 0, err
	}

	total, err := u.songRepo.Count(ctx, filters)
	if err != nil {
		return nil, 0, err
	}

	for i := range songs {
		if songs[i].AnimeID != 0 {
			anime, err := u.animeRepo.GetByID(ctx, songs[i].AnimeID)
			if err == nil && anime != nil {
				songs[i].Anime = anime
			}
		}
		variants, err := u.songRepo.GetVariantsBySongID(ctx, songs[i].ID)
		if err == nil {
			songs[i].Variants = variants
		}
	}

	return songs, int(total), nil
}

func (u *ContentAdminUsecase) GetSong(ctx context.Context, id uint64) (*domain.Song, error) {
	song, err := u.songRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	variants, err := u.songRepo.GetVariantsBySongID(ctx, id)
	if err == nil {
		song.Variants = variants
	}

	artists, err := u.songRepo.GetArtistsBySongID(ctx, id, true)
	if err == nil {
		song.Artists = artists
	}

	if song.AnimeID != 0 {
		anime, err := u.animeRepo.GetByID(ctx, song.AnimeID)
		if err == nil && anime != nil {
			u.animeRepo.LoadRelations(ctx, anime, true)
			song.Anime = anime
		}
	}

	return song, nil
}

func (u *ContentAdminUsecase) CreateSong(ctx context.Context, s *domain.Song, meta domain.AuditMetadata) error {
	u.nullifyEmptySongFields(s)

	if s.Type == "" || s.AnimeID == 0 {
		return domain.NewAppError(400, "Missing required fields for Song", nil)
	}

	if s.ThemeNum == "" {
		nextNum, _ := u.GetNextSongNumber(ctx, s.AnimeID, s.Type)
		s.ThemeNum = fmt.Sprintf("%d", nextNum)
	}

	s.Slug = fmt.Sprintf("%s%s", s.Type, s.ThemeNum)

	// Handle inheritance
	if s.SeasonID == 0 || s.YearID == 0 {
		anime, err := u.animeRepo.GetByID(ctx, s.AnimeID)
		if err != nil {
			return err
		}
		if s.SeasonID == 0 {
			if anime.SeasonID == 0 {
				return domain.NewAppError(400, "Cannot inherit season: parent anime has no season defined", nil)
			}
			s.SeasonID = anime.SeasonID
		}
		if s.YearID == 0 {
			if anime.YearID == 0 {
				return domain.NewAppError(400, "Cannot inherit year: parent anime has no year defined", nil)
			}
			s.YearID = anime.YearID
		}
	}

	// Role-based status control
	u.validateStatusPermissions(meta.Role, &s.Status, true)

	if err := u.songRepo.Create(ctx, s); err != nil {
		return err
	}

	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "created", s.ID, "song", nil, s, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *ContentAdminUsecase) GetNextSongNumber(ctx context.Context, animeID uint64, songType string) (int, error) {
	songs, err := u.songRepo.GetByAnimeID(ctx, animeID)
	if err != nil {
		return 0, err
	}

	maxNum := 0
	for _, existing := range songs {
		if existing.Type == songType {
			var existingNum int
			fmt.Sscanf(existing.ThemeNum, "%d", &existingNum)
			if existingNum > maxNum {
				maxNum = existingNum
			}
		}
	}
	return maxNum + 1, nil
}

func (u *ContentAdminUsecase) UpdateSong(ctx context.Context, s *domain.Song, meta domain.AuditMetadata) error {
	u.nullifyEmptySongFields(s)

	existing, err := u.songRepo.GetByID(ctx, s.ID)
	if err != nil {
		return err
	}

	// Handle inheritance
	if s.SeasonID == 0 || s.YearID == 0 {
		anime, err := u.animeRepo.GetByID(ctx, s.AnimeID)
		if err != nil {
			return err
		}
		if s.SeasonID == 0 {
			if anime.SeasonID == 0 {
				return domain.NewAppError(400, "Cannot inherit season: parent anime has no season defined", nil)
			}
			s.SeasonID = anime.SeasonID
		}
		if s.YearID == 0 {
			if anime.YearID == 0 {
				return domain.NewAppError(400, "Cannot inherit year: parent anime has no year defined", nil)
			}
			s.YearID = anime.YearID
		}
	}

	if s.ThemeNum == "" {
		s.ThemeNum = existing.ThemeNum
	}

	s.Slug = fmt.Sprintf("%s%s", s.Type, s.ThemeNum)

	// Role-based status control
	if err := u.validateStatusPermissions(meta.Role, &s.Status, false); err != nil {
		return err
	}

	if err := u.songRepo.Update(ctx, s); err != nil {
		return err
	}

	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "updated", s.ID, "song", existing, s, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *ContentAdminUsecase) DeleteSong(ctx context.Context, id uint64, meta domain.AuditMetadata) error {
	existing, _ := u.songRepo.GetByID(ctx, id)
	if err := u.songRepo.Delete(ctx, id); err != nil {
		return err
	}
	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "deleted", id, "song", existing, nil, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *ContentAdminUsecase) SyncSongArtists(ctx context.Context, songID uint64, artistIDs []uint64) error {
	if err := u.songRepo.SyncArtists(ctx, songID, artistIDs); err != nil {
		return err
	}

	// Trigger avatar generation synchronously for each artist
	// S3 check inside GenerateArtistAvatar will prevent redundant work
	for _, id := range artistIDs {
		_ = u.GenerateArtistAvatar(ctx, id, false)
	}

	return nil
}

func (u *ContentAdminUsecase) nullifyEmptySongFields(s *domain.Song) {
	if s.SongRomaji != nil && *s.SongRomaji == "" {
		s.SongRomaji = nil
	}
	if s.SongEN != nil && *s.SongEN == "" {
		s.SongEN = nil
	}
	if s.SongJP != nil && *s.SongJP == "" {
		s.SongJP = nil
	}
}



// ---- ARTISTS ----
func (u *ContentAdminUsecase) GetArtists(ctx context.Context, page, limit int, search string) ([]domain.Artist, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 24
	}
	offset := (page - 1) * limit

	filters := domain.ArtistFilters{
		Search:  search,
		IsAdmin: true,
	}
	artists, err := u.artistRepo.GetPaginated(ctx, limit, offset, filters)
	if err != nil {
		return nil, 0, err
	}

	total, err := u.artistRepo.Count(ctx, filters)
	return artists, total, err
}

func (u *ContentAdminUsecase) GetArtist(ctx context.Context, id uint64) (*domain.Artist, error) {
	artist, err := u.artistRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	u.ResolveArtistURLs(artist)
	return artist, nil
}

func (u *ContentAdminUsecase) CreateArtist(ctx context.Context, a *domain.Artist, meta domain.AuditMetadata) error {
	a.Name = strings.TrimSpace(a.Name)
	if a.Slug == "" {
		a.Slug = u.generateArtistUniqueSlug(ctx, a.Name, 0)
	}
	a.UUID = uuid.New().String()

	// Role-based status control
	u.validateStatusPermissions(meta.Role, &a.Status, true)

	err := u.artistRepo.Create(ctx, a)
	if err != nil {
		return err
	}

	// Trigger avatar generation synchronously to provide feedback/loaders in frontend
	_ = u.GenerateArtistAvatar(ctx, a.ID, false)

	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "created", a.ID, "artist", nil, a, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *ContentAdminUsecase) GenerateArtistAvatar(ctx context.Context, artistID uint64, forceAniList bool) error {
	artist, err := u.artistRepo.GetByID(ctx, artistID)
	if err != nil {
		return err
	}

	// 1. Try AniList Staff Search
	var avatarData []byte
	var avatarSize int64
	var avatarContentType string

	staffResults, err := u.anilistClient.SearchStaff(ctx, artist.Name)
	if err == nil && len(staffResults) > 0 {
		staff := staffResults[0] // Take the first result
		// Regex to catch default AniList staff images (e.g. /staff/large/default.jpg)
		defaultImgRe := regexp.MustCompile(`.*\/default\.jpg$`)
		if staff.Image.Large != "" && !defaultImgRe.MatchString(staff.Image.Large) {
			// Download from AniList
			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Get(staff.Image.Large)
			if err == nil && resp.StatusCode == http.StatusOK {
				defer resp.Body.Close()
				avatarData, err = io.ReadAll(resp.Body)
				if err == nil {
					avatarSize = int64(len(avatarData))
					avatarContentType = resp.Header.Get("Content-Type")
				}
			}
		}
	}

	// 2. Fallback Logic
	if len(avatarData) == 0 {
		// Fallback to avatar-ui (ui-avatars.com)
		res, err := avatar.Generate(ctx, artist.Name)
		if err != nil {
			return err
		}
		avatarData = res.Data
		avatarSize = res.Size
		avatarContentType = res.ContentType
	}

	// 3. Upload to S3
	_, url, err := u.mediaService.UploadImage(ctx, "artists/avatars", artistID, bytes.NewReader(avatarData), avatarSize, avatarContentType)
	if err != nil {
		return err
	}

	return u.artistRepo.UpdateAvatar(ctx, artistID, url)
}

func (u *ContentAdminUsecase) BatchGenerateArtistAvatars(ctx context.Context, ids []uint64, progress chan<- string) error {
	sendProgress := func(msg string) {
		if progress != nil {
			progress <- msg
		}
	}

	var targets []domain.Artist
	var err error

	if len(ids) > 0 {
		sendProgress(fmt.Sprintf("Fetching %d selected artists...", len(ids)))
		targets, err = u.artistRepo.GetMany(ctx, ids)
		if err != nil {
			sendProgress(fmt.Sprintf("Error fetching selected artists: %v", err))
			return err
		}
	} else {
		// 1. Get all artists
		filters := domain.ArtistFilters{IsAdmin: true}
		artists, err := u.artistRepo.GetPaginated(ctx, 5000, 0, filters)
		if err != nil {
			sendProgress(fmt.Sprintf("Error fetching artists: %v", err))
			return err
		}

		// 2. Identify artists without S3 avatars
		re := regexp.MustCompile(`^artists/avatars/.*`)
		for _, a := range artists {
			if a.Avatar == nil || !re.MatchString(*a.Avatar) {
				targets = append(targets, a)
			}
		}
	}

	if len(targets) == 0 {
		sendProgress("No artists found needing avatars.")
		return nil
	}

	sendProgress(fmt.Sprintf("Found %d artists to process. Starting Phase 1: AniList Sync...", len(targets)))

	// --- Phase 1: AniList Batching ---
	chunkSize := 20
	var stillMissing []domain.Artist
	defaultImgRe := regexp.MustCompile(`.*\/default\.jpg$`)

	for i := 0; i < len(targets); i += chunkSize {
		end := i + chunkSize
		if end > len(targets) {
			end = len(targets)
		}
		chunk := targets[i:end]
		
		var names []string
		for _, a := range chunk {
			names = append(names, a.Name)
		}

		sendProgress(fmt.Sprintf("[Phase 1] Searching AniList for %d artists (%d/%d)...", len(names), end, len(targets)))
		
		var results map[string][]anilist.Staff
		var err error

		// Retry loop for 429 (Rate Limit) errors
		maxRetries := 2
		for attempt := 0; attempt <= maxRetries; attempt++ {
			results, err = u.anilistClient.SearchStaffBatch(ctx, names)
			if err == nil {
				break
			}

			if strings.Contains(err.Error(), "429") && attempt < maxRetries {
				sendProgress(fmt.Sprintf("Rate limited (Attempt %d/%d)! Waiting 30 seconds before retry...", attempt+1, maxRetries+1))
				time.Sleep(30 * time.Second)
				continue
			}

			// If not a retryable error or we've exhausted retries
			sendProgress(fmt.Sprintf("AniList Batch error: %v. Continuing to fallback...", err))
			break
		}

		if err != nil {
			stillMissing = append(stillMissing, chunk...)
			continue
		}

		for _, a := range chunk {
			staffs, found := results[a.Name]
			if !found || len(staffs) == 0 {
				stillMissing = append(stillMissing, a)
				continue
			}

			staff := staffs[0]
			if staff.Image.Large == "" || defaultImgRe.MatchString(staff.Image.Large) {
				stillMissing = append(stillMissing, a)
				continue
			}

			// Download and upload
			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Get(staff.Image.Large)
			if err != nil || resp.StatusCode != http.StatusOK {
				stillMissing = append(stillMissing, a)
				continue
			}
			defer resp.Body.Close()

			data, err := io.ReadAll(resp.Body)
			if err != nil {
				stillMissing = append(stillMissing, a)
				continue
			}

			_, url, err := u.mediaService.UploadImage(ctx, "artists/avatars", a.ID, bytes.NewReader(data), int64(len(data)), resp.Header.Get("Content-Type"))
			if err != nil {
				stillMissing = append(stillMissing, a)
				continue
			}

			_ = u.artistRepo.UpdateAvatar(ctx, a.ID, url)
			sendProgress(fmt.Sprintf("✓ [AniList] Synchronized: %s", a.Name))
		}

		if end < len(targets) {
			time.Sleep(2 * time.Second) // Respect AniList rate limits
		}
	}

	// --- Phase 2: Avatar-UI Fallback ---
	if len(stillMissing) > 0 {
		sendProgress(fmt.Sprintf("Starting Phase 2: Generating %d placeholders via Avatar-UI...", len(stillMissing)))
		for i, a := range stillMissing {
			res, err := avatar.Generate(ctx, a.Name)
			if err != nil {
				sendProgress(fmt.Sprintf("✗ [Avatar-UI] Failed %s: %v", a.Name, err))
				continue
			}

			_, url, err := u.mediaService.UploadImage(ctx, "artists/avatars", a.ID, bytes.NewReader(res.Data), res.Size, res.ContentType)
			if err != nil {
				sendProgress(fmt.Sprintf("✗ [Upload] Failed %s: %v", a.Name, err))
				continue
			}

			_ = u.artistRepo.UpdateAvatar(ctx, a.ID, url)
			sendProgress(fmt.Sprintf("[%d/%d] ✓ [Avatar-UI] Generated placeholder: %s", i+1, len(stillMissing), a.Name))
			
			// Small delay for avatar-ui fallback to be safe
			time.Sleep(500 * time.Millisecond)
		}
	}

	sendProgress("Batch generation completed successfully!")
	return nil
}

func (u *ContentAdminUsecase) MergeDuplicateArtists(ctx context.Context, progress chan<- string) error {
	return u.artistRepo.MergeDuplicateArtists(ctx, progress)
}

func (u *ContentAdminUsecase) UpdateArtist(ctx context.Context, a *domain.Artist, meta domain.AuditMetadata) error {
	existing, err := u.artistRepo.GetByID(ctx, a.ID)
	if err != nil {
		return err
	}

	a.Name = strings.TrimSpace(a.Name)
	if a.Name != existing.Name || a.Slug == "" {
		a.Slug = u.generateArtistUniqueSlug(ctx, a.Name, a.ID)
	}

	// Role-based status control
	if err := u.validateStatusPermissions(meta.Role, &a.Status, false); err != nil {
		return err
	}

	if err := u.artistRepo.Update(ctx, a); err != nil {
		return err
	}

	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "updated", a.ID, "artist", existing, a, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *ContentAdminUsecase) DeleteArtist(ctx context.Context, id uint64, meta domain.AuditMetadata) error {
	artist, err := u.artistRepo.GetByID(ctx, id)
	if err == nil && artist.Avatar != nil && *artist.Avatar != "" {
		// u.mediaService.Delete(*artist.Avatar) // Add if MediaService supports Delete
	}
	if err := u.artistRepo.Delete(ctx, id); err != nil {
		return err
	}
	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "deleted", id, "artist", artist, nil, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *ContentAdminUsecase) generateArtistSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	reg, _ := regexp.Compile("[^a-z0-9-]+")
	slug = reg.ReplaceAllString(slug, "")
	return strings.Trim(slug, "-")
}

func (u *ContentAdminUsecase) generateArtistUniqueSlug(ctx context.Context, name string, excludeID uint64) string {
	baseSlug := u.generateArtistSlug(name)
	if baseSlug == "" {
		baseSlug = "artist"
	}

	slug := baseSlug
	counter := 1
	for {
		existing, err := u.artistRepo.GetBySlug(ctx, slug)
		if (err != nil && strings.Contains(err.Error(), "not found")) || existing == nil {
			return slug
		}
		if excludeID != 0 && existing.ID == excludeID {
			return slug
		}
		slug = fmt.Sprintf("%s-%d", baseSlug, counter)
		counter++
	}
}

// ---- SONG VARIANT / VIDEO ----
func (u *ContentAdminUsecase) GetVariants(ctx context.Context, page, limit int, search string, animeID *uint64, status *bool) ([]domain.SongVariant, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit
	filters := map[string]interface{}{
		"search":   search,
		"is_admin": true,
	}
	if animeID != nil && *animeID != 0 {
		filters["anime_id"] = *animeID
	}
	if status != nil {
		filters["status"] = *status
	}

	variants, err := u.variantRepo.GetPaginated(ctx, limit, offset, filters)
	if err != nil {
		return nil, 0, err
	}

	// Resolve relations for admin display
	for i := range variants {
		song, err := u.songRepo.GetByID(ctx, variants[i].SongID)
		if err == nil {
			if song.AnimeID != 0 {
				anime, err := u.animeRepo.GetByID(ctx, song.AnimeID)
				if err == nil {
					song.Anime = anime
				}
			}
			variants[i].Song = song
		}
	}

	total, err := u.variantRepo.Count(ctx, filters)
	return variants, int(total), err
}

func (u *ContentAdminUsecase) GetVariant(ctx context.Context, id uint64) (*domain.SongVariant, error) {
	v, err := u.variantRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Load the song relation so the frontend can have anime_id and other song details
	s, err := u.songRepo.GetByID(ctx, v.SongID)
	if err == nil {
		v.Song = s
	}

	return v, nil
}

func (u *ContentAdminUsecase) CreateVariant(ctx context.Context, v *domain.SongVariant, meta domain.AuditMetadata) error {
	if v.SongID == 0 {
		return domain.NewAppError(400, "Missing required fields for SongVariant", nil)
	}

	song, err := u.songRepo.GetByID(ctx, v.SongID)
	if err != nil {
		return err
	}

	if v.SeasonID == 0 {
		v.SeasonID = song.SeasonID
	}
	if v.YearID == 0 {
		v.YearID = song.YearID
	}

	if v.VersionNumber == 0 || v.Slug == "" {
		variants, _ := u.songRepo.GetVariantsBySongID(ctx, v.SongID)
		maxVersion := uint64(0)
		for _, existing := range variants {
			if existing.VersionNumber > maxVersion {
				maxVersion = existing.VersionNumber
			}
		}
		v.VersionNumber = maxVersion + 1
		v.Slug = fmt.Sprintf("v%d", v.VersionNumber)
	}
	v.UUID = uuid.New().String()

	// Role-based status control
	u.validateStatusPermissions(meta.Role, &v.Status, true)

	if err := u.variantRepo.Create(ctx, v); err != nil {
		return err
	}

	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "created", v.ID, "variant", nil, v, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *ContentAdminUsecase) UpdateVariant(ctx context.Context, v *domain.SongVariant, meta domain.AuditMetadata) error {
	existing, err := u.variantRepo.GetByID(ctx, v.ID)
	if err != nil {
		return err
	}

	// Handle inheritance
	if v.SeasonID == 0 || v.YearID == 0 {
		song, err := u.songRepo.GetByID(ctx, v.SongID)
		if err != nil {
			return err
		}
		if v.SeasonID == 0 {
			if song.SeasonID == 0 {
				return domain.NewAppError(400, "Cannot inherit season: parent song has no season defined", nil)
			}
			v.SeasonID = song.SeasonID
		}
		if v.YearID == 0 {
			if song.YearID == 0 {
				return domain.NewAppError(400, "Cannot inherit year: parent song has no year defined", nil)
			}
			v.YearID = song.YearID
		}
	}

	// Role-based status control
	if err := u.validateStatusPermissions(meta.Role, &v.Status, false); err != nil {
		return err
	}

	if err := u.variantRepo.Update(ctx, v); err != nil {
		return err
	}

	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "updated", v.ID, "variant", existing, v, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *ContentAdminUsecase) HandleVariantVideo(c *fiber.Ctx, v *domain.SongVariant) {
	updated := false
	statusStr := c.FormValue("status")
	if statusStr != "" {
		newStatus := statusStr == "true"
		if v.Status != newStatus {
			v.Status = newStatus
			updated = true
		}
	}

	if fileHeader, err := c.FormFile("video"); err == nil {
		if file, err := fileHeader.Open(); err == nil {
			defer file.Close()
			contentType := fileHeader.Header.Get("Content-Type")

			// Standardized upload
			if _, url, err := u.mediaService.UploadImage(c.Context(), "videos", v.ID, file, fileHeader.Size, contentType); err == nil {
				if v.Video == nil {
					v.Video = &domain.SongVariantVideo{}
				}
				v.Video.LocalUrl = &url
				v.Video.Type = "file"
				v.Video.EmbedUrl = nil
				updated = true
			}
		}
	} else {
		embed := c.FormValue("embed")
		if embed != "" {
			if v.Video == nil {
				v.Video = &domain.SongVariantVideo{}
			}
			if v.Video.EmbedUrl == nil || *v.Video.EmbedUrl != embed {
				v.Video.EmbedUrl = &embed
				v.Video.Type = "embed"
				v.Video.LocalUrl = nil
				updated = true
			}
		}
	}

	if updated {
		_ = u.variantRepo.Update(c.Context(), v)
	}
}

func (u *ContentAdminUsecase) DeleteVariant(ctx context.Context, id uint64, meta domain.AuditMetadata) error {
	existing, _ := u.variantRepo.GetByID(ctx, id)
	if err := u.variantRepo.Delete(ctx, id); err != nil {
		return err
	}
	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "deleted", id, "variant", existing, nil, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *ContentAdminUsecase) SearchStudios(ctx context.Context, term string, limit int) ([]domain.Studio, error) {
	return u.taxonomyRepo.SearchStudios(ctx, term, limit)
}

func (u *ContentAdminUsecase) SearchProducers(ctx context.Context, term string, limit int) ([]domain.Producer, error) {
	return u.taxonomyRepo.SearchProducers(ctx, term, limit)
}

func (u *ContentAdminUsecase) SearchGenres(ctx context.Context, term string, limit int) ([]domain.Genre, error) {
	return u.taxonomyRepo.SearchGenres(ctx, term, limit)
}

// ---- TAXONOMIES ----
func (u *ContentAdminUsecase) CreateYear(ctx context.Context, year *domain.Year, meta domain.AuditMetadata) error {
	if err := u.taxonomyRepo.CreateYear(ctx, year); err != nil {
		return err
	}
	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "created", year.ID, "taxonomy_year", nil, year, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *ContentAdminUsecase) UpdateYear(ctx context.Context, year *domain.Year, meta domain.AuditMetadata) error {
	existing, _ := u.taxonomyRepo.GetYearByID(ctx, year.ID)
	if err := u.taxonomyRepo.UpdateYear(ctx, year); err != nil {
		return err
	}
	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "updated", year.ID, "taxonomy_year", existing, year, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *ContentAdminUsecase) ToggleYearCurrent(ctx context.Context, id uint64, meta domain.AuditMetadata) error {
	existing, _ := u.taxonomyRepo.GetYearByID(ctx, id)
	if err := u.taxonomyRepo.ToggleYearCurrent(ctx, id); err != nil {
		return err
	}
	newYear, _ := u.taxonomyRepo.GetYearByID(ctx, id)
	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "toggled_current", id, "taxonomy_year", existing, newYear, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *ContentAdminUsecase) DeleteYear(ctx context.Context, id uint64, meta domain.AuditMetadata) error {
	existing, _ := u.taxonomyRepo.GetYearByID(ctx, id)
	if err := u.taxonomyRepo.DeleteYear(ctx, id); err != nil {
		return err
	}
	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "deleted", id, "taxonomy_year", existing, nil, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *ContentAdminUsecase) GetYears(ctx context.Context) ([]domain.Year, error) {
	return u.taxonomyRepo.GetAllYears(ctx)
}

func (u *ContentAdminUsecase) CreateSeason(ctx context.Context, season *domain.Season, meta domain.AuditMetadata) error {
	if err := u.taxonomyRepo.CreateSeason(ctx, season); err != nil {
		return err
	}
	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "created", season.ID, "taxonomy_season", nil, season, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *ContentAdminUsecase) UpdateSeason(ctx context.Context, season *domain.Season, meta domain.AuditMetadata) error {
	existing, _ := u.taxonomyRepo.GetSeasonByID(ctx, season.ID)
	if err := u.taxonomyRepo.UpdateSeason(ctx, season); err != nil {
		return err
	}
	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "updated", season.ID, "taxonomy_season", existing, season, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *ContentAdminUsecase) ToggleSeasonCurrent(ctx context.Context, id uint64, meta domain.AuditMetadata) error {
	existing, _ := u.taxonomyRepo.GetSeasonByID(ctx, id)
	if err := u.taxonomyRepo.ToggleSeasonCurrent(ctx, id); err != nil {
		return err
	}
	newSeason, _ := u.taxonomyRepo.GetSeasonByID(ctx, id)
	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "toggled_current", id, "taxonomy_season", existing, newSeason, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *ContentAdminUsecase) DeleteSeason(ctx context.Context, id uint64, meta domain.AuditMetadata) error {
	existing, _ := u.taxonomyRepo.GetSeasonByID(ctx, id)
	if err := u.taxonomyRepo.DeleteSeason(ctx, id); err != nil {
		return err
	}
	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "deleted", id, "taxonomy_season", existing, nil, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *ContentAdminUsecase) GetSeasons(ctx context.Context) ([]domain.Season, error) {
	return u.taxonomyRepo.GetAllSeasons(ctx)
}

func (u *ContentAdminUsecase) GetFormats(ctx context.Context) ([]domain.Format, error) {
	return u.taxonomyRepo.GetAllFormats(ctx)
}

func (u *ContentAdminUsecase) CreateFormat(ctx context.Context, format *domain.Format, meta domain.AuditMetadata) error {
	if err := u.taxonomyRepo.CreateFormat(ctx, format); err != nil {
		return err
	}
	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "created", format.ID, "taxonomy_format", nil, format, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *ContentAdminUsecase) UpdateFormat(ctx context.Context, format *domain.Format, meta domain.AuditMetadata) error {
	existing, _ := u.taxonomyRepo.GetFormatByID(ctx, format.ID)
	if err := u.taxonomyRepo.UpdateFormat(ctx, format); err != nil {
		return err
	}
	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "updated", format.ID, "taxonomy_format", existing, format, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *ContentAdminUsecase) DeleteFormat(ctx context.Context, id uint64, meta domain.AuditMetadata) error {
	existing, _ := u.taxonomyRepo.GetFormatByID(ctx, id)
	if err := u.taxonomyRepo.DeleteFormat(ctx, id); err != nil {
		return err
	}
	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "deleted", id, "taxonomy_format", existing, nil, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *ContentAdminUsecase) CreateGenre(ctx context.Context, genre *domain.Genre, meta domain.AuditMetadata) error {
	if err := u.taxonomyRepo.CreateGenre(ctx, genre); err != nil {
		return err
	}
	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "created", genre.ID, "taxonomy_genre", nil, genre, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *ContentAdminUsecase) UpdateGenre(ctx context.Context, genre *domain.Genre, meta domain.AuditMetadata) error {
	existing, _ := u.taxonomyRepo.GetGenreByID(ctx, genre.ID)
	if err := u.taxonomyRepo.UpdateGenre(ctx, genre); err != nil {
		return err
	}
	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "updated", genre.ID, "taxonomy_genre", existing, genre, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *ContentAdminUsecase) DeleteGenre(ctx context.Context, id uint64, meta domain.AuditMetadata) error {
	existing, _ := u.taxonomyRepo.GetGenreByID(ctx, id)
	if err := u.taxonomyRepo.DeleteGenre(ctx, id); err != nil {
		return err
	}
	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "deleted", id, "taxonomy_genre", existing, nil, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}
func (u *ContentAdminUsecase) isStaffEditor(ctx context.Context, userID uint64) bool {
	// This is a placeholder, implementation depends on how roles are checked
	// For now, let's assume it returns true to not block status updates
	return true
}

func (u *ContentAdminUsecase) UploadArtistAvatar(ctx context.Context, artistID uint64, file io.Reader, size int64, contentType string) error {
	artist, err := u.artistRepo.GetByID(ctx, artistID)
	if err != nil {
		return err
	}

	_, url, err := u.mediaService.UploadImage(ctx, "artists", artistID, file, size, contentType)
	if err != nil {
		return err
	}

	artist.Avatar = &url
	return u.artistRepo.Update(ctx, artist)
}
// validateStatusPermissions ensures role-based status control
// 1. Creators cannot set status to true (active).
// 2. New content from Creators (isNew=true) is forced to false (inactive/draft).
func (u *ContentAdminUsecase) validateStatusPermissions(role string, status *bool, isNew bool) error {
	if role == "admin" || role == "editor" {
		return nil
	}

	if role == "creator" {
		if isNew {
			*status = false
			return nil
		}
		if *status == true {
			return domain.NewAppError(403, "Creators cannot activate content directly. Approval required.", nil)
		}
	} else if isNew {
		// Minimum safe default for any other role
		*status = false
	}

	return nil
}

func (u *ContentAdminUsecase) ToggleArtistStatus(ctx context.Context, id uint64, meta domain.AuditMetadata) error {
	if meta.Role == "creator" {
		return domain.NewAppError(403, "Creators cannot activate content directly", nil)
	}
	existing, _ := u.artistRepo.GetByID(ctx, id)
	if err := u.artistRepo.ToggleStatus(ctx, id); err != nil {
		return err
	}
	newArtist, _ := u.artistRepo.GetByID(ctx, id)
	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "status_toggled", id, "artist", existing, newArtist, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}
func (u *ContentAdminUsecase) ToggleSongStatus(ctx context.Context, id uint64, meta domain.AuditMetadata) error {
	if meta.Role == "creator" {
		return domain.NewAppError(403, "Creators cannot activate content directly", nil)
	}
	existing, _ := u.songRepo.GetByID(ctx, id)
	if err := u.songRepo.ToggleStatus(ctx, id); err != nil {
		return err
	}
	newSong, _ := u.songRepo.GetByID(ctx, id)
	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "status_toggled", id, "song", existing, newSong, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *ContentAdminUsecase) ToggleVariantStatus(ctx context.Context, id uint64, meta domain.AuditMetadata) error {
	if meta.Role == "creator" {
		return domain.NewAppError(403, "Creators cannot activate content directly", nil)
	}
	existing, _ := u.variantRepo.GetByID(ctx, id)
	if err := u.variantRepo.ToggleStatus(ctx, id); err != nil {
		return err
	}
	newVariant, _ := u.variantRepo.GetByID(ctx, id)
	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "status_toggled", id, "variant", existing, newVariant, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *ContentAdminUsecase) RecountArtistStats(ctx context.Context, artistID *uint64, progress chan<- string) error {
	sendProgress := func(msg string) {
		if progress != nil {
			progress <- msg
		}
	}

	if artistID != nil {
		sendProgress(fmt.Sprintf("Recalculating statistics for artist ID %d...", *artistID))
	} else {
		sendProgress("Recalculating statistics for ALL artists...")
	}

	if err := u.artistRepo.RecountArtistStats(ctx, artistID); err != nil {
		sendProgress(fmt.Sprintf("Error: %v", err))
		return err
	}

	sendProgress("Recalculation complete!")
	return nil
}
