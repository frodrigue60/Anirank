package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"anirank/api/internal/domain"
	"anirank/api/internal/infrastructure"
	"anirank/api/internal/infrastructure/anilist"
	"anirank/api/internal/infrastructure/animethemes"
	"anirank/api/internal/pkg/avatar"

	"github.com/google/uuid"
)

const (
	atPageSize       = 100          // records per AnimeThemes page
	atInterPageDelay = 250          // ms between AT pages
	alChunkSize      = 50           // max IDs per AniList batch request
	alInterChunkSec  = 1            // seconds between AniList chunks
	alMaxRetries     = 3            // retries on AniList 429
	importSource     = "animethemes"
)

// ImportUsecase handles the full AnimeThemes + AniList data hydration pipeline.
type ImportUsecase struct {
	jobRepo      domain.ImportJobRepository
	animeRepo    domain.AnimeRepository
	songRepo     domain.SongRepository
	artistRepo   domain.ArtistRepository
	taxonomyRepo domain.TaxonomyRepository
	atClient     animethemes.AnimeThemesClient
	alClient     anilist.AnilistClient
	mediaService infrastructure.MediaService

	mu        sync.Mutex
	cancelMap map[string]context.CancelFunc
}

// NewImportUsecase creates a new ImportUsecase with all required dependencies.
func NewImportUsecase(
	jobRepo domain.ImportJobRepository,
	animeRepo domain.AnimeRepository,
	songRepo domain.SongRepository,
	artistRepo domain.ArtistRepository,
	taxonomyRepo domain.TaxonomyRepository,
	atClient animethemes.AnimeThemesClient,
	alClient anilist.AnilistClient,
	mediaService infrastructure.MediaService,
) *ImportUsecase {
	return &ImportUsecase{
		jobRepo:      jobRepo,
		animeRepo:    animeRepo,
		songRepo:     songRepo,
		artistRepo:   artistRepo,
		taxonomyRepo: taxonomyRepo,
		atClient:     atClient,
		alClient:     alClient,
		mediaService: mediaService,
		cancelMap:    make(map[string]context.CancelFunc),
	}
}

// ─── Public API ───────────────────────────────────────────────────────────────

// StartAnimeThemesImport creates a new import job and launches it in a background goroutine.
// Returns the job ID immediately so the caller can poll / stream progress.
func (u *ImportUsecase) StartAnimeThemesImport(ctx context.Context) (string, error) {
	// Check if a job is already running
	latest, err := u.jobRepo.GetLatest(ctx, importSource)
	if err == nil && latest != nil && latest.Status == domain.ImportJobRunning {
		return "", fmt.Errorf("an import job is already running (id=%s)", latest.ID)
	}

	now := time.Now().UTC()
	job := &domain.ImportJob{
		ID:         uuid.New().String(),
		Source:     importSource,
		Status:     domain.ImportJobRunning,
		StartedAt:  &now,
		UpdatedAt:  now,
		ErrorsJSON: "[]",
	}

	if err := u.jobRepo.Create(ctx, job); err != nil {
		return "", fmt.Errorf("failed to create job: %w", err)
	}

	// Create a background context for the goroutine that can be canceled
	bgCtx, cancel := context.WithCancel(context.Background())
	u.mu.Lock()
	u.cancelMap[job.ID] = cancel
	u.mu.Unlock()

	go u.runImport(bgCtx, job)

	return job.ID, nil
}

// GetJobStatus returns the current state of an import job.
func (u *ImportUsecase) GetJobStatus(ctx context.Context, jobID string) (*domain.ImportJob, error) {
	return u.jobRepo.GetByID(ctx, jobID)
}

// GetLatestJobStatus returns the latest import job for a given source.
func (u *ImportUsecase) GetLatestJobStatus(ctx context.Context, source string) (*domain.ImportJob, error) {
	return u.jobRepo.GetLatest(ctx, source)
}

// CancelJob marks a pending/running job as canceled.
func (u *ImportUsecase) CancelJob(ctx context.Context, jobID string) error {
	u.mu.Lock()
	if cancel, exists := u.cancelMap[jobID]; exists {
		cancel()
		delete(u.cancelMap, jobID)
	}
	u.mu.Unlock()

	return u.jobRepo.Cancel(ctx, jobID)
}

// ─── Worker ───────────────────────────────────────────────────────────────────

func (u *ImportUsecase) runImport(ctx context.Context, job *domain.ImportJob) {
	defer func() {
		u.mu.Lock()
		delete(u.cancelMap, job.ID)
		u.mu.Unlock()
	}()

	defer func() {
		if r := recover(); r != nil {
			job.Status = domain.ImportJobFailed
			job.Errors = append(job.Errors, fmt.Sprintf("panic: %v", r))
			_ = u.jobRepo.UpdateProgress(context.Background(), job)
		}
	}()

	// Verify AniList availability before executing the import pipeline
	if err := u.alClient.Ping(ctx); err != nil {
		job.Status = domain.ImportJobFailed
		job.Errors = append(job.Errors, fmt.Sprintf("AniList API is offline/unresponsive: %v", err))
		_ = u.jobRepo.UpdateProgress(context.Background(), job)
		return
	}

	if err := u.phaseAnimeThemes(ctx, job); err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			job.Status = domain.ImportJobCanceled
			_ = u.jobRepo.UpdateProgress(context.Background(), job)
			return
		}
		job.Status = domain.ImportJobFailed
		job.Errors = append(job.Errors, err.Error())
		_ = u.jobRepo.UpdateProgress(context.Background(), job)
		return
	}

	// Phase 2: Enrich from AniList
	if err := u.phaseAniList(ctx, job); err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			job.Status = domain.ImportJobCanceled
			_ = u.jobRepo.UpdateProgress(context.Background(), job)
			return
		}
		// Treat AniList failure as fatal per resilience requirements
		job.Status = domain.ImportJobFailed
		job.Errors = append(job.Errors, fmt.Sprintf("AniList enrichment failed: %v", err))
		_ = u.jobRepo.UpdateProgress(context.Background(), job)
		return
	}

	now := time.Now().UTC()
	job.Status = domain.ImportJobDone
	job.FinishedAt = &now
	_ = u.jobRepo.UpdateProgress(context.Background(), job)
}

// ─── Phase 1: AnimeThemes ─────────────────────────────────────────────────────

func (u *ImportUsecase) phaseAnimeThemes(ctx context.Context, job *domain.ImportJob) error {
	page := 1

	for {
		// Check for cancellation
		if ctx.Err() != nil {
			return ctx.Err()
		}

		resp, err := u.atClient.FetchAnimePage(ctx, page, atPageSize)
		if err != nil {
			return fmt.Errorf("animethemes page %d: %w", page, err)
		}

		for _, atAnime := range resp.Anime {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if err := u.processAnime(ctx, job, &atAnime); err != nil {
				job.Errors = append(job.Errors, fmt.Sprintf("anime %d: %v", atAnime.ID, err))
			}
			job.Processed++
		}

		job.CurrentPage = page
		_ = u.jobRepo.UpdateProgress(ctx, job)

		// Stop when there is no next page
		if resp.Links.Next == nil || *resp.Links.Next == "" {
			break
		}

		page++
		time.Sleep(time.Duration(atInterPageDelay) * time.Millisecond)
	}

	job.TotalPages = page
	return nil
}

func (u *ImportUsecase) processAnime(ctx context.Context, job *domain.ImportJob, atAnime *animethemes.ATAnime) error {
	// 1. Resolve taxonomy IDs
	year, err := u.taxonomyRepo.GetOrCreateYear(ctx, fmt.Sprintf("%d", atAnime.Year))
	if err != nil {
		return fmt.Errorf("year: %w", err)
	}
	season, err := u.taxonomyRepo.GetOrCreateSeason(ctx, normalizeSeason(atAnime.Season))
	if err != nil {
		return fmt.Errorf("season: %w", err)
	}
	format, err := u.taxonomyRepo.GetOrCreateFormat(ctx, normalizeFormat(atAnime.MediaFormat))
	if err != nil {
		return fmt.Errorf("format: %w", err)
	}

	// 2. Extract AniList ID from resources
	var anilistID *int64
	for _, res := range atAnime.Resources {
		if strings.EqualFold(res.Site, "AniList") && res.ExternalID != nil {
			v := int64(*res.ExternalID)
			anilistID = &v
			break
		}
	}

	// 3. Build slug — use AT slug, convert to our format
	slug := buildUniqueAnimeSlug(atAnime.Slug, atAnime.ID)

	// 4. Upsert anime
	anime := &domain.Anime{
		Title:         atAnime.Name,
		Slug:          slug,
		Description:   atAnime.Synopsis,
		AnilistID:     anilistID,
		AnimeThemesID: (*uint64)(&atAnime.ID),
		YearID:        year.ID,
		SeasonID:      season.ID,
		FormatID:      format.ID,
	}

	created, err := u.animeRepo.UpsertFromAnimeThemes(ctx, anime)
	if err != nil {
		return fmt.Errorf("upsert anime: %w", err)
	}
	if created {
		job.Created++
	} else {
		job.Skipped++
	}

	// 5. Process themes → songs → variants
	for _, theme := range atAnime.AnimeThemes {
		if theme.Song == nil {
			continue
		}
		if err := u.processSong(ctx, job, &theme, anime, year.ID, season.ID); err != nil {
			job.Errors = append(job.Errors, fmt.Sprintf("theme %d: %v", theme.ID, err))
		}
	}

	return nil
}

func (u *ImportUsecase) processSong(
	ctx context.Context,
	job *domain.ImportJob,
	theme *animethemes.ATTheme,
	anime *domain.Anime,
	yearID, seasonID uint64,
) error {
	// Map AnimeThemes type to our DB enum: OP/ED/IN→INS/OTH
	songType := mapSongType(theme.Type)

	// Build theme_num: sequence nil = "1"
	themeNum := "1"
	if theme.Sequence != nil && *theme.Sequence > 0 {
		themeNum = fmt.Sprintf("%d", *theme.Sequence)
	}

	song := &domain.Song{
		SongRomaji:    &theme.Song.Title,
		ThemeNum:      themeNum,
		Type:          songType,
		AnimeID:       anime.ID,
		YearID:        yearID,
		SeasonID:      seasonID,
		AnimeThemesID: (*uint64)(&theme.Song.ID),
	}

	created, err := u.songRepo.UpsertSongFromAnimeThemes(ctx, song)
	if err != nil {
		return fmt.Errorf("upsert song: %w", err)
	}
	if created {
		job.Created++
	} else {
		job.Skipped++
	}

	// Artists
	for _, atArtist := range theme.Song.Artists {
		artist := &domain.Artist{
			Name:          atArtist.Name,
			Slug:          atArtist.Slug,
			AnimeThemesID: (*uint64)(&atArtist.ID),
		}
		created, err := u.artistRepo.UpsertFromAnimeThemes(ctx, artist)
		if err != nil {
			job.Errors = append(job.Errors, fmt.Sprintf("artist %d: %v", atArtist.ID, err))
			continue
		}

		// Check if we need to generate an avatar
		needsAvatar := created
		if !created {
			if dbArtist, err := u.artistRepo.GetByID(ctx, artist.ID); err == nil && dbArtist.Avatar == nil {
				needsAvatar = true
			}
		}

		if needsAvatar {
			if res, err := avatar.Generate(ctx, artist.Name, 180); err == nil {
				path, _, err := u.mediaService.UploadWithResolutions(ctx, "artists/avatars", artist.ID, bytes.NewReader(res.Data), infrastructure.PresetSquare)
				if err == nil {
					artist.Avatar = &path
					_ = u.artistRepo.Update(ctx, artist)
				}
			}
		}

		// Link pivot
		_ = u.songRepo.LinkArtistToSong(ctx, song.ID, artist.ID)
	}

	// Variants (entries)
	for entryIdx, entry := range theme.Entries {
		version := 1
		if entry.Version != nil {
			version = *entry.Version
		}
		if version == 0 {
			version = 1
		}

		// Build slug: "OP1v2" style
		variantSlug := fmt.Sprintf("%s%s", strings.ToUpper(songType), themeNum)
		if version > 1 {
			variantSlug = fmt.Sprintf("%sv%d", variantSlug, version)
		}

		variant := &domain.SongVariant{
			VersionNumber: uint64(version),
			SongID:        song.ID,
			Slug:          variantSlug,
			SeasonID:      seasonID,
			YearID:        yearID,
			Spoiler:       entry.Spoiler,
			NSFW:          entry.NSFW,
			AnimeThemesID: (*uint64)(&entry.ID),
		}

		// Process all videos for this variant entry
		var videos []domain.SongVariantVideo
		for _, entryVideo := range entry.Videos {
			path := entryVideo.Path
			if path != "" {
				isNC, isBD, resolution, isUncensored, isSubbed, isLyrics, source, overlap := parseVideoTags(entryVideo.Tags)
				vSrc := path
				videos = append(videos, domain.SongVariantVideo{
					VideoSrc:     &vSrc,
					IsNC:         isNC,
					IsBD:         isBD,
					Resolution:   resolution,
					IsUncensored: isUncensored,
					IsSubbed:     isSubbed,
					IsLyrics:     isLyrics,
					Source:       source,
					Overlap:      overlap,
				})
			}
		}

		variantCreated, err := u.songRepo.UpsertVariantFromAnimeThemes(ctx, variant, videos)
		if err != nil {
			job.Errors = append(job.Errors, fmt.Sprintf("variant %d (entry %d): %v", entry.ID, entryIdx, err))
			continue
		}
		if variantCreated {
			job.Created++
		} else {
			job.Skipped++
		}
	}

	return nil
}

// ─── Phase 2: AniList Enrichment ─────────────────────────────────────────────

func (u *ImportUsecase) phaseAniList(ctx context.Context, job *domain.ImportJob) error {
	// Fetch all animes that have an anilist_id but are missing cover or banner
	type animeRow struct {
		ID        uint64 `db:"id"`
		AnilistID *int64 `db:"anilist_id"`
	}

	// We use a raw query via GetByAnilistIDs approach, but simpler: just load all
	// animes with anilist_id set. The EnrichFromAniList query uses COALESCE so it
	// won't overwrite existing data.
	allAnimes, err := u.animeRepo.GetPaginated(ctx, 5000, 0, domain.AnimeFilters{IsAdmin: true})
	if err != nil {
		return fmt.Errorf("anilist enrichment: fetch animes: %w", err)
	}

	// Collect anilist IDs
	var anilistIDs []int
	idMap := make(map[int]int64) // anilist_id int → internal anilist_id int64
	for _, a := range allAnimes {
		if a.AnilistID != nil {
			anilistIDs = append(anilistIDs, int(*a.AnilistID))
			idMap[int(*a.AnilistID)] = *a.AnilistID
		}
	}

	if len(anilistIDs) == 0 {
		return nil
	}

	// Process in chunks of alChunkSize
	for i := 0; i < len(anilistIDs); i += alChunkSize {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		end := i + alChunkSize
		if end > len(anilistIDs) {
			end = len(anilistIDs)
		}
		chunk := anilistIDs[i:end]

		var medias []anilist.Media
		var lastErr error
		for attempt := 0; attempt < alMaxRetries; attempt++ {
			medias, lastErr = u.alClient.GetMediaByIDs(ctx, chunk)
			if lastErr == nil {
				break
			}
			// On rate limit, wait 65 seconds before retrying
			if strings.Contains(lastErr.Error(), "429") || strings.Contains(lastErr.Error(), "rate") {
				time.Sleep(65 * time.Second)
			} else {
				break
			}
		}
		if lastErr != nil {
			errStr := fmt.Sprintf("anilist chunk %d-%d failed after %d retries: %v", i, end, alMaxRetries, lastErr)
			job.Errors = append(job.Errors, errStr)
			_ = u.jobRepo.UpdateProgress(ctx, job)
			return fmt.Errorf("%s", errStr)
		}

		for _, media := range medias {
			cover := media.CoverImage.ExtraLarge
			banner := media.BannerImage
			desc := media.Description

			var coverPtr, bannerPtr, descPtr *string
			if cover != "" {
				coverPtr = &cover
			}
			if banner != "" {
				bannerPtr = &banner
			}
			if desc != "" {
				descPtr = &desc
			}

			anilistID := int64(media.ID)
			if err := u.animeRepo.EnrichFromAniList(ctx, anilistID, coverPtr, bannerPtr, descPtr); err != nil {
				job.Errors = append(job.Errors, fmt.Sprintf("enrich anilist %d: %v", media.ID, err))
			}
		}

		job.CurrentPage = i/alChunkSize + 1
		_ = u.jobRepo.UpdateProgress(ctx, job)

		time.Sleep(time.Duration(alInterChunkSec) * time.Second)
	}

	return nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func (u *ImportUsecase) isCanceled(ctx context.Context, jobID string) bool {
	job, err := u.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return false
	}
	return job.Status == domain.ImportJobCanceled
}

// normalizeSeason maps AnimeThemes season strings to our DB season names.
func normalizeSeason(s string) string {
	switch strings.ToLower(s) {
	case "winter":
		return "Winter"
	case "spring":
		return "Spring"
	case "summer":
		return "Summer"
	case "fall":
		return "Fall"
	default:
		if s == "" {
			return "Winter"
		}
		return s
	}
}

// normalizeFormat maps AnimeThemes media_format to our DB format names.
func normalizeFormat(f string) string {
	switch strings.ToUpper(f) {
	case "TV":
		return "TV"
	case "TV_SHORT":
		return "TV Short"
	case "OVA":
		return "OVA"
	case "ONA":
		return "ONA"
	case "MOVIE":
		return "Movie"
	case "SPECIAL":
		return "Special"
	case "MUSIC":
		return "Music"
	default:
		if f == "" {
			return "TV"
		}
		return f
	}
}

// mapSongType maps AnimeThemes theme type to our DB song type slug.
// DB constraint: 'OP' | 'ED' | 'INS' | 'OTH'
func mapSongType(t string) string {
	switch strings.ToUpper(t) {
	case "OP":
		return "OP"
	case "ED":
		return "ED"
	case "IN":
		return "IN"
	case "INS":
		return "INS"
	default:
		return "OTH"
	}
}

// buildUniqueAnimeSlug is defined in anime_repository.go in the postgres package.
// Here we replicate it as a package-level helper for the usecase.
func buildUniqueAnimeSlug(base string, animeThemesID uint64) string {
	s := strings.ToLower(base)
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteRune('-')
		}
	}
	clean := strings.Trim(b.String(), "-")
	for strings.Contains(clean, "--") {
		clean = strings.ReplaceAll(clean, "--", "-")
	}
	if len(clean) == 0 {
		clean = fmt.Sprintf("anime-%d", animeThemesID)
	}
	return clean
}

// MarshalJob serialises an ImportJob to JSON for SSE events.
func MarshalJob(job *domain.ImportJob) string {
	b, _ := json.Marshal(map[string]interface{}{
		"id":           job.ID,
		"status":       job.Status,
		"current_page": job.CurrentPage,
		"total_pages":  job.TotalPages,
		"processed":    job.Processed,
		"created":      job.Created,
		"skipped":      job.Skipped,
		"errors":       job.Errors,
	})
	return string(b)
}

func parseVideoTags(tags string) (isNC bool, isBD bool, resolution int, isUncensored bool, isSubbed bool, isLyrics bool, source string, overlap string) {
	normalized := strings.ToUpper(tags)
	isNC = strings.Contains(normalized, "NC")
	isBD = strings.Contains(normalized, "BD")
	isUncensored = strings.Contains(normalized, "UNCENSORED")
	isSubbed = strings.Contains(normalized, "SUBBED")
	isLyrics = strings.Contains(normalized, "LYRICS")

	// Determine Source (BD, TV, WEB, DVD, LD)
	if strings.Contains(normalized, "BD") || strings.Contains(normalized, "BLU-RAY") {
		source = "BD"
	} else if strings.Contains(normalized, "WEB") {
		source = "WEB"
	} else if strings.Contains(normalized, "DVD") {
		source = "DVD"
	} else if strings.Contains(normalized, "LD") {
		source = "LD"
	} else {
		source = "TV" // Default to TV
	}

	// Determine Overlap
	if strings.Contains(normalized, "OVERLAP") {
		overlap = "Overlap"
	} else if strings.Contains(normalized, "TRANSITION") {
		overlap = "Transition"
	} else {
		overlap = "None"
	}

	if strings.Contains(normalized, "2160") {
		resolution = 2160
	} else if strings.Contains(normalized, "1440") {
		resolution = 1440
	} else if strings.Contains(normalized, "1080") {
		resolution = 1080
	} else if strings.Contains(normalized, "720") {
		resolution = 720
	} else if strings.Contains(normalized, "576") {
		resolution = 576
	} else if strings.Contains(normalized, "480") {
		resolution = 480
	} else if strings.Contains(normalized, "360") {
		resolution = 360
	}
	return
}
