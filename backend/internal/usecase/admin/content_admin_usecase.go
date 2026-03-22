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
	"net/http"
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
	}

	if err := u.animeRepo.Create(ctx, &newAnime); err != nil {
		return nil, err
	}

	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "created_from_anilist", newAnime.ID, "anime", nil, newAnime, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	_ = u.SyncAnimeWithAnilist(ctx, &newAnime, media)

	go func(m anilist.Media, a domain.Anime) {
		updated := false
		if m.CoverImage.ExtraLarge != "" {
			if imgUrl, err := u.downloadAndStore(ctx, m.CoverImage.ExtraLarge, "animes/covers", uint64(m.ID)); err == nil {
				a.Cover = &imgUrl
				updated = true
			}
		}
		if m.BannerImage != "" {
			if imgUrl, err := u.downloadAndStore(ctx, m.BannerImage, "animes/banners", uint64(m.ID)); err == nil {
				a.Banner = &imgUrl
				updated = true
			}
		}
		if updated {
			u.animeRepo.Update(context.Background(), &a)
		}
	}(*media, newAnime)

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

	return u.SyncAnimeWithAnilist(ctx, anime, &media)
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
	_ = u.animeRepo.UpdateStudios(ctx, anime.ID, studioIDs)
	_ = u.animeRepo.UpdateProducers(ctx, anime.ID, producerIDs)

	var genreIDs []uint64
	for _, gName := range media.Genres {
		obj, err := u.taxonomyRepo.GetOrCreateGenre(ctx, gName)
		if err == nil {
			genreIDs = append(genreIDs, obj.ID)
		}
	}
	_ = u.animeRepo.UpdateGenres(ctx, anime.ID, genreIDs)

	var links []domain.ExternalLink
	for _, l := range media.ExternalLinks {
		links = append(links, domain.ExternalLink{
			Name: l.Site,
			URL:  l.URL,
		})
	}
	_ = u.animeRepo.UpdateExternalLinks(ctx, anime.ID, links)

	return nil
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


func (u *ContentAdminUsecase) downloadAndStore(ctx context.Context, url string, prefix string, id uint64) (string, error) {
	if u.mediaService == nil {
		return "", fmt.Errorf("no media service configured")
	}

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch image: %s", resp.Status)
	}

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	path, url, err := u.mediaService.UploadImage(ctx, prefix, id, bytes.NewReader(buf), int64(len(buf)), resp.Header.Get("Content-Type"))
	_ = path // avoid unused
	return url, err
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
	return u.songRepo.SyncArtists(ctx, songID, artistIDs)
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

	// Role-based status control
	u.validateStatusPermissions(meta.Role, &a.Status, true)

	err := u.artistRepo.Create(ctx, a)
	if err != nil {
		return err
	}

	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "created", a.ID, "artist", nil, a, &meta.URL, &meta.IPAddress, &meta.UserAgent)

	go func() {
		_ = u.GenerateArtistAvatar(context.Background(), a.ID)
	}()

	return nil
}

func (u *ContentAdminUsecase) GenerateArtistAvatar(ctx context.Context, artistID uint64) error {
	artist, err := u.artistRepo.GetByID(ctx, artistID)
	if err != nil {
		return err
	}

	res, err := avatar.Generate(ctx, artist.Name)
	if err != nil {
		return err
	}

	_, url, err := u.mediaService.UploadImage(ctx, "artists/avatars", artistID, bytes.NewReader(res.Data), res.Size, res.ContentType)
	if err != nil {
		return err
	}

	return u.artistRepo.UpdateAvatar(ctx, artistID, url)
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

