package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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
	backfillSource   = "backfill_titles"
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

// StartTitleBackfill creates a new backfill job and launches it in the background.
// The job iterates over all animes that have an anilist_id but are missing title_english
// or title_native, and hydrates only those title fields (no cover/banner downloads).
// Returns the job ID immediately so the caller can stream progress via /import/:jobID/stream.
func (u *ImportUsecase) StartTitleBackfill(ctx context.Context) (string, error) {
	latest, err := u.jobRepo.GetLatest(ctx, backfillSource)
	if err == nil && latest != nil && latest.Status == domain.ImportJobRunning {
		return "", fmt.Errorf("a backfill job is already running (id=%s)", latest.ID)
	}

	now := time.Now().UTC()
	job := &domain.ImportJob{
		ID:         uuid.New().String(),
		Source:     backfillSource,
		Status:     domain.ImportJobRunning,
		StartedAt:  &now,
		UpdatedAt:  now,
		ErrorsJSON: "[]",
	}

	if err := u.jobRepo.Create(ctx, job); err != nil {
		return "", fmt.Errorf("failed to create backfill job: %w", err)
	}

	bgCtx, cancel := context.WithCancel(context.Background())
	u.mu.Lock()
	u.cancelMap[job.ID] = cancel
	u.mu.Unlock()

	go u.runBackfill(bgCtx, job)

	return job.ID, nil
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

// ─── Backfill Worker ──────────────────────────────────────────────────────────

func (u *ImportUsecase) runBackfill(ctx context.Context, job *domain.ImportJob) {
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

	if err := u.alClient.Ping(ctx); err != nil {
		job.Status = domain.ImportJobFailed
		job.Errors = append(job.Errors, fmt.Sprintf("AniList API is offline/unresponsive: %v", err))
		_ = u.jobRepo.UpdateProgress(context.Background(), job)
		return
	}

	if err := u.phaseBackfillTitles(ctx, job); err != nil {
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

	now := time.Now().UTC()
	job.Status = domain.ImportJobDone
	job.FinishedAt = &now
	_ = u.jobRepo.UpdateProgress(context.Background(), job)
}

// phaseBackfillTitles iterates animes with missing title_english/title_native and
// enriches only the title fields and synonyms from AniList. Cover/banner downloads
// are intentionally skipped to keep the job fast.
func (u *ImportUsecase) phaseBackfillTitles(ctx context.Context, job *domain.ImportJob) error {
	const dbBatchSize = 200

	// Determine total so the dashboard can show progress
	total, err := u.animeRepo.CountAnimesWithMissingTitleVariants(ctx)
	if err != nil {
		return fmt.Errorf("backfill: count animes: %w", err)
	}
	if total == 0 {
		return nil // nothing to do
	}

	// Use TotalPages to carry the total record count (repurposed for this job type)
	job.TotalPages = (total + alChunkSize - 1) / alChunkSize
	_ = u.jobRepo.UpdateProgress(ctx, job)

	var lastID uint64 = 0
	totalProcessed := 0

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		batch, err := u.animeRepo.GetAnimesWithMissingTitleVariants(ctx, dbBatchSize, lastID)
		if err != nil {
			return fmt.Errorf("backfill: fetch batch after ID %d: %w", lastID, err)
		}
		if len(batch) == 0 {
			break
		}

		// Build slice of AniList IDs and lookup map
		var anilistIDs []int
		animeMap := make(map[int64]*domain.Anime)
		for i := range batch {
			if batch[i].AnilistID != nil {
				anilistIDs = append(anilistIDs, int(*batch[i].AnilistID))
				animeMap[*batch[i].AnilistID] = &batch[i]
			}
		}

		// Process in sub-chunks of 50 (AniList limit)
		for idx := 0; idx < len(anilistIDs); idx += alChunkSize {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			subEnd := idx + alChunkSize
			if subEnd > len(anilistIDs) {
				subEnd = len(anilistIDs)
			}
			subChunk := anilistIDs[idx:subEnd]

			var medias []anilist.Media
			var lastErr error
			for attempt := 0; attempt < alMaxRetries; attempt++ {
				medias, lastErr = u.alClient.GetMediaByIDs(ctx, subChunk)
				if lastErr == nil {
					break
				}
				if strings.Contains(lastErr.Error(), "429") || strings.Contains(lastErr.Error(), "rate") {
					time.Sleep(65 * time.Second)
				} else {
					break
				}
			}
			if lastErr != nil {
				errStr := fmt.Sprintf("backfill chunk %d-%d failed after %d retries: %v", totalProcessed+idx, totalProcessed+subEnd, alMaxRetries, lastErr)
				job.Errors = append(job.Errors, errStr)
				_ = u.jobRepo.UpdateProgress(ctx, job)
				return fmt.Errorf("%s", errStr)
			}

			for _, media := range medias {
				anilistID := int64(media.ID)
				if _, exists := animeMap[anilistID]; !exists {
					continue
				}

				var titleEnglishPtr *string
				if media.Title.English != "" {
					e := media.Title.English
					titleEnglishPtr = &e
				}
				var titleNativePtr *string
				if media.Title.Native != "" {
					n := media.Title.Native
					titleNativePtr = &n
				}
				synonyms := make([]string, 0, len(media.Synonyms))
				for _, s := range media.Synonyms {
					if s != "" {
						synonyms = append(synonyms, s)
					}
				}

				// EnrichFromAniList with nil cover/banner/description to leave them untouched
				if err := u.animeRepo.EnrichFromAniList(ctx, anilistID, nil, nil, nil, titleEnglishPtr, titleNativePtr, synonyms); err != nil {
					job.Errors = append(job.Errors, fmt.Sprintf("backfill enrich anilist %d: %v", media.ID, err))
				}

				job.Processed++
			}

			totalProcessed += len(subChunk)
			job.CurrentPage = totalProcessed/alChunkSize + 1
			_ = u.jobRepo.UpdateProgress(ctx, job)

			time.Sleep(time.Duration(alInterChunkSec) * time.Second)
		}

		lastID = batch[len(batch)-1].ID
	}

	return nil
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

	upsertResult, err := u.animeRepo.UpsertFromAnimeThemes(ctx, anime)
	if err != nil {
		return fmt.Errorf("upsert anime: %w", err)
	}
	if upsertResult.DuplicateAnilist && anilistID != nil {
		job.Errors = append(job.Errors, fmt.Sprintf(
			"anime %d: duplicate anilist_id %d — songs linked to existing catalog entry",
			atAnime.ID, *anilistID,
		))
	}
	if upsertResult.Created {
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

		// Build slug: "V1", "V2" version style to align with local architecture
		variantSlug := fmt.Sprintf("V%d", version)

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
		videoInputs := make([]ATVideoInput, 0, len(entry.Videos))
		for _, entryVideo := range entry.Videos {
			videoInputs = append(videoInputs, ATVideoInput{
				Path: entryVideo.Path,
				Tags: entryVideo.Tags,
			})
		}
		videos := buildVideosFromATInputs(videoInputs)

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
	offset := 0
	const dbBatchSize = 200
	totalProcessed := 0

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		animesBatch, err := u.animeRepo.GetPaginated(ctx, dbBatchSize, offset, domain.AnimeFilters{IsAdmin: true})
		if err != nil {
			return fmt.Errorf("anilist enrichment: fetch animes: %w", err)
		}
		if len(animesBatch) == 0 {
			break
		}

		var chunkAnilistIDs []int
		chunkAnimeMap := make(map[int64]*domain.Anime)
		for i := range animesBatch {
			if animesBatch[i].AnilistID != nil {
				chunkAnilistIDs = append(chunkAnilistIDs, int(*animesBatch[i].AnilistID))
				chunkAnimeMap[*animesBatch[i].AnilistID] = &animesBatch[i]
			}
		}

		if len(chunkAnilistIDs) > 0 {
			for idx := 0; idx < len(chunkAnilistIDs); idx += alChunkSize {
				subEnd := idx + alChunkSize
				if subEnd > len(chunkAnilistIDs) {
					subEnd = len(chunkAnilistIDs)
				}
				subChunk := chunkAnilistIDs[idx:subEnd]

				var medias []anilist.Media
				var lastErr error
				for attempt := 0; attempt < alMaxRetries; attempt++ {
					medias, lastErr = u.alClient.GetMediaByIDs(ctx, subChunk)
					if lastErr == nil {
						break
					}
					if strings.Contains(lastErr.Error(), "429") || strings.Contains(lastErr.Error(), "rate") {
						time.Sleep(65 * time.Second)
					} else {
						break
					}
				}
				if lastErr != nil {
					errStr := fmt.Sprintf("anilist chunk %d-%d failed after %d retries: %v", totalProcessed+idx, totalProcessed+subEnd, alMaxRetries, lastErr)
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
					anime, exists := chunkAnimeMap[anilistID]
					if !exists {
						continue
					}

					var finalCover, finalBanner *string
					if cover != "" {
						if path, err := u.downloadAndStore(ctx, cover, "animes/covers", anime.ID, infrastructure.PresetPoster); err == nil {
							finalCover = &path
							time.Sleep(200 * time.Millisecond)
						} else {
							job.Errors = append(job.Errors, fmt.Sprintf("failed to download cover for anime %d: %v", anime.ID, err))
						}
					}
					if banner != "" {
						if path, err := u.downloadAndStore(ctx, banner, "animes/banners", anime.ID, infrastructure.PresetLandscape); err == nil {
							finalBanner = &path
							time.Sleep(200 * time.Millisecond)
						} else {
							job.Errors = append(job.Errors, fmt.Sprintf("failed to download banner for anime %d: %v", anime.ID, err))
						}
					}

					coverToSave := coverPtr
					if finalCover != nil {
						coverToSave = finalCover
					}
					bannerToSave := bannerPtr
					if finalBanner != nil {
						bannerToSave = finalBanner
					}

					var titleEnglishPtr *string
					if media.Title.English != "" {
						e := media.Title.English
						titleEnglishPtr = &e
					}
					var titleNativePtr *string
					if media.Title.Native != "" {
						n := media.Title.Native
						titleNativePtr = &n
					}
					synonyms := make([]string, 0)
					for _, s := range media.Synonyms {
						if s != "" {
							synonyms = append(synonyms, s)
						}
					}

					if err := u.animeRepo.EnrichFromAniList(ctx, anilistID, coverToSave, bannerToSave, descPtr, titleEnglishPtr, titleNativePtr, synonyms); err != nil {
						job.Errors = append(job.Errors, fmt.Sprintf("enrich anilist %d: %v", media.ID, err))
					}

					var genreIDs []uint64
					for _, g := range media.Genres {
						if obj, err := u.taxonomyRepo.GetOrCreateGenre(ctx, g); err == nil && obj != nil {
							genreIDs = append(genreIDs, obj.ID)
						} else if err != nil {
							job.Errors = append(job.Errors, fmt.Sprintf("genre %s for anime %d: %v", g, anime.ID, err))
						}
					}
					if err := u.animeRepo.UpdateGenres(ctx, anime.ID, genreIDs); err != nil {
						job.Errors = append(job.Errors, fmt.Sprintf("update genres for anime %d: %v", anime.ID, err))
					}

					var studioIDs []uint64
					var producerIDs []uint64
					for _, edge := range media.Studios.Edges {
						if edge.Node.Name == "" {
							continue
						}
						if edge.IsMain {
							if obj, err := u.taxonomyRepo.GetOrCreateStudio(ctx, edge.Node.Name); err == nil && obj != nil {
								studioIDs = append(studioIDs, obj.ID)
							} else if err != nil {
								job.Errors = append(job.Errors, fmt.Sprintf("studio %s for anime %d: %v", edge.Node.Name, anime.ID, err))
							}
						} else {
							if obj, err := u.taxonomyRepo.GetOrCreateProducer(ctx, edge.Node.Name); err == nil && obj != nil {
								producerIDs = append(producerIDs, obj.ID)
							} else if err != nil {
								job.Errors = append(job.Errors, fmt.Sprintf("producer %s for anime %d: %v", edge.Node.Name, anime.ID, err))
							}
						}
					}
					if err := u.animeRepo.UpdateStudios(ctx, anime.ID, studioIDs); err != nil {
						job.Errors = append(job.Errors, fmt.Sprintf("update studios for anime %d: %v", anime.ID, err))
					}
					if err := u.animeRepo.UpdateProducers(ctx, anime.ID, producerIDs); err != nil {
						job.Errors = append(job.Errors, fmt.Sprintf("update producers for anime %d: %v", anime.ID, err))
					}

					if len(media.ExternalLinks) > 0 {
						if err := u.animeRepo.LoadRelations(ctx, anime, true); err != nil {
							job.Errors = append(job.Errors, fmt.Sprintf("load relations for anime %d: %v", anime.ID, err))
						}

						linksMap := make(map[string]domain.ExternalLink)
						for _, l := range anime.ExternalLinks {
							linksMap[l.URL] = l
						}

						for _, l := range media.ExternalLinks {
							if l.URL == "" {
								continue
							}
							linksMap[l.URL] = domain.ExternalLink{
								Name: l.Site,
								URL:  l.URL,
								Type: strings.ToLower(l.Site),
							}
						}

						finalLinks := make([]domain.ExternalLink, 0, len(linksMap))
						for _, l := range linksMap {
							finalLinks = append(finalLinks, l)
						}

						if err := u.animeRepo.UpdateExternalLinks(ctx, anime.ID, finalLinks); err != nil {
							job.Errors = append(job.Errors, fmt.Sprintf("update external links for anime %d: %v", anime.ID, err))
						}
					}
				}

				totalProcessed += len(subChunk)
				job.CurrentPage = totalProcessed/alChunkSize + 1
				_ = u.jobRepo.UpdateProgress(ctx, job)

				time.Sleep(time.Duration(alInterChunkSec) * time.Second)
			}
		}

		offset += len(animesBatch)
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

func (u *ImportUsecase) downloadAndStore(ctx context.Context, url string, prefix string, id uint64, preset infrastructure.ResolutionPreset) (string, error) {
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

	path, _, err := u.mediaService.UploadWithResolutions(ctx, prefix, id, bytes.NewReader(buf), preset)
	return path, err
}
