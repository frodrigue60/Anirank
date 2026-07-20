package admin

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"anirank/api/internal/domain"
	"anirank/api/internal/infrastructure"
	"anirank/api/internal/infrastructure/anilist"
	"anirank/api/internal/infrastructure/animethemes"
	"anirank/api/internal/pkg/avatar"

	"github.com/google/uuid"
)

const (
	atThemeFetchDelay = 150 // ms between theme show requests
)

// StartIncrementalSongSync creates a job that imports only AnimeThemes songs
// newer than MAX(songs.anime_themes_id), then ensures anime + variants.
func (u *ImportUsecase) StartIncrementalSongSync(ctx context.Context) (string, error) {
	latest, err := u.jobRepo.GetLatest(ctx, incrementalSongSource)
	if err == nil && latest != nil && latest.Status == domain.ImportJobRunning {
		return "", fmt.Errorf("an incremental song sync is already running (id=%s)", latest.ID)
	}

	now := time.Now().UTC()
	job := &domain.ImportJob{
		ID:         uuid.New().String(),
		Source:     incrementalSongSource,
		Status:     domain.ImportJobRunning,
		StartedAt:  &now,
		UpdatedAt:  now,
		ErrorsJSON: "[]",
	}

	if err := u.jobRepo.Create(ctx, job); err != nil {
		return "", fmt.Errorf("failed to create incremental job: %w", err)
	}

	bgCtx, cancel := context.WithCancel(context.Background())
	u.mu.Lock()
	u.cancelMap[job.ID] = cancel
	u.mu.Unlock()

	go u.runIncrementalSongSync(bgCtx, job)

	return job.ID, nil
}

func (u *ImportUsecase) runIncrementalSongSync(ctx context.Context, job *domain.ImportJob) {
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

	touchedAnilistIDs, err := u.phaseIncrementalSongs(ctx, job)
	if err != nil {
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

	if err := u.phaseIncrementalAniList(ctx, job, touchedAnilistIDs); err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			job.Status = domain.ImportJobCanceled
			_ = u.jobRepo.UpdateProgress(context.Background(), job)
			return
		}
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

// phaseIncrementalSongs pages AT songs after the local watermark and hydrates
// anime → song → variants for each related theme.
func (u *ImportUsecase) phaseIncrementalSongs(ctx context.Context, job *domain.ImportJob) (map[int64]uint64, error) {
	job.Phase = 1

	watermark, err := u.songRepo.GetMaxAnimeThemesID(ctx)
	if err != nil {
		return nil, fmt.Errorf("incremental: read song watermark: %w", err)
	}

	touched := make(map[int64]uint64) // anilistID -> internal anime ID
	page := 1

	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		resp, err := u.atClient.FetchSongPage(ctx, page, atPageSize, watermark)
		if err != nil {
			return nil, fmt.Errorf("incremental: song page %d: %w", page, err)
		}
		if len(resp.Songs) == 0 {
			break
		}

		for _, atSong := range resp.Songs {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}

			if len(atSong.AnimeThemes) == 0 {
				job.Skipped++
				job.Processed++
				continue
			}

			for _, themeStub := range atSong.AnimeThemes {
				if ctx.Err() != nil {
					return nil, ctx.Err()
				}

				theme, err := u.atClient.FetchThemeByID(ctx, themeStub.ID)
				if err != nil {
					job.Errors = append(job.Errors, fmt.Sprintf("theme %d (song %d): %v", themeStub.ID, atSong.ID, err))
					time.Sleep(time.Duration(atThemeFetchDelay) * time.Millisecond)
					continue
				}

				anilistID, animeID, err := u.processIncrementalTheme(ctx, job, theme)
				if err != nil {
					job.Errors = append(job.Errors, fmt.Sprintf("theme %d: %v", theme.ID, err))
				} else if anilistID != nil && animeID > 0 {
					touched[*anilistID] = animeID
				}

				time.Sleep(time.Duration(atThemeFetchDelay) * time.Millisecond)
			}

			job.Processed++
		}

		job.CurrentPage = page
		_ = u.jobRepo.UpdateProgress(ctx, job)

		if resp.Links.Next == nil || *resp.Links.Next == "" {
			break
		}
		page++
		time.Sleep(time.Duration(atInterPageDelay) * time.Millisecond)
	}

	job.TotalPages = page
	return touched, nil
}

func (u *ImportUsecase) processIncrementalTheme(
	ctx context.Context,
	job *domain.ImportJob,
	theme *animethemes.ATTheme,
) (*int64, uint64, error) {
	if theme == nil || theme.Anime == nil || theme.Song == nil {
		return nil, 0, fmt.Errorf("incomplete theme payload")
	}

	atAnime := theme.Anime

	// Resolve AniList id from resources (show-by-slug); skip if anime already linked.
	var anilistID *int64
	if existingID, err := u.animeRepo.GetIDByAnimeThemesID(ctx, atAnime.ID); err != nil {
		return nil, 0, err
	} else if existingID == 0 {
		if full, err := u.atClient.FetchAnimeBySlug(ctx, atAnime.Slug); err == nil && full != nil {
			atAnime.Resources = full.Resources
			atAnime.Synopsis = firstNonNilString(atAnime.Synopsis, full.Synopsis)
		}
		anilistID = extractAniListID(atAnime.Resources)
	}

	year, err := u.taxonomyRepo.GetOrCreateYear(ctx, fmt.Sprintf("%d", atAnime.Year))
	if err != nil {
		return nil, 0, fmt.Errorf("year: %w", err)
	}
	season, err := u.taxonomyRepo.GetOrCreateSeason(ctx, normalizeSeason(atAnime.Season))
	if err != nil {
		return nil, 0, fmt.Errorf("season: %w", err)
	}
	format, err := u.taxonomyRepo.GetOrCreateFormat(ctx, normalizeFormat(atAnime.MediaFormat))
	if err != nil {
		return nil, 0, fmt.Errorf("format: %w", err)
	}

	if anilistID == nil {
		anilistID = extractAniListID(atAnime.Resources)
	}

	anime := &domain.Anime{
		Title:         atAnime.Name,
		Slug:          buildUniqueAnimeSlug(atAnime.Slug, atAnime.ID),
		Description:   atAnime.Synopsis,
		AnilistID:     anilistID,
		AnimeThemesID: &atAnime.ID,
		YearID:        year.ID,
		SeasonID:      season.ID,
		FormatID:      format.ID,
	}

	upsertResult, err := u.animeRepo.UpsertFromAnimeThemes(ctx, anime)
	if err != nil {
		return nil, 0, fmt.Errorf("upsert anime: %w", err)
	}
	if upsertResult.Created {
		job.Created++
	}

	// Load anilist id from DB row when anime already existed without a fresh resource fetch.
	if anilistID == nil && anime.ID > 0 {
		if existing, err := u.animeRepo.GetByID(ctx, anime.ID); err == nil && existing != nil {
			anilistID = existing.AnilistID
		}
	}

	if err := u.upsertIncrementalSongGraph(ctx, job, theme, anime, year.ID, season.ID); err != nil {
		return anilistID, anime.ID, err
	}

	return anilistID, anime.ID, nil
}

func (u *ImportUsecase) upsertIncrementalSongGraph(
	ctx context.Context,
	job *domain.ImportJob,
	theme *animethemes.ATTheme,
	anime *domain.Anime,
	yearID, seasonID uint64,
) error {
	songType := mapSongType(theme.Type)
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
		AnimeThemesID: &theme.Song.ID,
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

	// Artists — always link; generate avatar only when missing.
	for _, atArtist := range theme.Song.Artists {
		artist := &domain.Artist{
			Name:          atArtist.Name,
			Slug:          atArtist.Slug,
			AnimeThemesID: &atArtist.ID,
		}
		artistCreated, err := u.artistRepo.UpsertFromAnimeThemes(ctx, artist)
		if err != nil {
			job.Errors = append(job.Errors, fmt.Sprintf("artist %d: %v", atArtist.ID, err))
			continue
		}

		needsAvatar := artistCreated
		if !artistCreated {
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
		_ = u.songRepo.LinkArtistToSong(ctx, song.ID, artist.ID)
	}

	// Variants/entries — always process, even when the song already existed.
	for entryIdx, entry := range theme.Entries {
		version := 1
		if entry.Version != nil {
			version = *entry.Version
		}
		if version == 0 {
			version = 1
		}

		variant := &domain.SongVariant{
			VersionNumber: uint64(version),
			SongID:        song.ID,
			Slug:          fmt.Sprintf("V%d", version),
			SeasonID:      seasonID,
			YearID:        yearID,
			Spoiler:       entry.Spoiler,
			NSFW:          entry.NSFW,
			AnimeThemesID: &entry.ID,
		}

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

// phaseIncrementalAniList enriches only animes touched by the song delta.
func (u *ImportUsecase) phaseIncrementalAniList(ctx context.Context, job *domain.ImportJob, touched map[int64]uint64) error {
	job.Phase = 2
	job.CurrentPage = 0

	if len(touched) == 0 {
		job.TotalPages = 1
		_ = u.jobRepo.UpdateProgress(ctx, job)
		return nil
	}

	ids := make([]int, 0, len(touched))
	for alID := range touched {
		ids = append(ids, int(alID))
	}

	job.TotalPages = (len(ids) + alChunkSize - 1) / alChunkSize
	if job.TotalPages < 1 {
		job.TotalPages = 1
	}
	_ = u.jobRepo.UpdateProgress(ctx, job)

	for idx := 0; idx < len(ids); idx += alChunkSize {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		subEnd := idx + alChunkSize
		if subEnd > len(ids) {
			subEnd = len(ids)
		}
		chunk := ids[idx:subEnd]

		var needsEnrichment []int
		animeByAL := make(map[int64]*domain.Anime)
		for _, alID := range chunk {
			internalID := touched[int64(alID)]
			anime, err := u.animeRepo.GetByID(ctx, internalID)
			if err != nil || anime == nil {
				continue
			}
			animeByAL[int64(alID)] = anime
			if animeFullyEnrichedFromAniList(anime) {
				job.Skipped++
				continue
			}
			needsEnrichment = append(needsEnrichment, alID)
		}

		job.CurrentPage = idx/alChunkSize + 1

		if len(needsEnrichment) == 0 {
			_ = u.jobRepo.UpdateProgress(ctx, job)
			continue
		}

		var medias []anilist.Media
		var lastErr error
		for attempt := 0; attempt < alMaxRetries; attempt++ {
			medias, lastErr = u.alClient.GetMediaByIDs(ctx, needsEnrichment)
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
			return fmt.Errorf("anilist chunk failed: %w", lastErr)
		}

		for _, media := range medias {
			anime := animeByAL[int64(media.ID)]
			if anime == nil {
				continue
			}
			if err := u.enrichAnimeFromMedia(ctx, job, anime, media); err != nil {
				job.Errors = append(job.Errors, fmt.Sprintf("enrich anilist %d: %v", media.ID, err))
			}
		}

		_ = u.jobRepo.UpdateProgress(ctx, job)
		time.Sleep(time.Duration(alInterChunkSec) * time.Second)
	}

	return nil
}

func (u *ImportUsecase) enrichAnimeFromMedia(ctx context.Context, job *domain.ImportJob, anime *domain.Anime, media anilist.Media) error {
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

	var finalCover, finalBanner *string
	if cover != "" && !hasStoredImage(anime.Cover) {
		if path, err := u.downloadAndStore(ctx, cover, "animes/covers", anime.ID, infrastructure.PresetPoster); err == nil {
			finalCover = &path
			time.Sleep(200 * time.Millisecond)
		} else {
			job.Errors = append(job.Errors, fmt.Sprintf("failed to download cover for anime %d: %v", anime.ID, err))
		}
	}
	if banner != "" && !hasStoredImage(anime.Banner) {
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
	} else if hasStoredImage(anime.Cover) {
		coverToSave = nil
	}
	bannerToSave := bannerPtr
	if finalBanner != nil {
		bannerToSave = finalBanner
	} else if hasStoredImage(anime.Banner) {
		bannerToSave = nil
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

	anilistID := int64(media.ID)
	if err := u.animeRepo.EnrichFromAniList(ctx, anilistID, coverToSave, bannerToSave, descPtr, titleEnglishPtr, titleNativePtr, synonyms); err != nil {
		return err
	}

	var genreIDs []uint64
	for _, g := range media.Genres {
		if obj, err := u.taxonomyRepo.GetOrCreateGenre(ctx, g); err == nil && obj != nil {
			genreIDs = append(genreIDs, obj.ID)
		}
	}
	_ = u.animeRepo.UpdateGenres(ctx, anime.ID, genreIDs)

	var studioIDs []uint64
	var producerIDs []uint64
	for _, edge := range media.Studios.Edges {
		if edge.Node.Name == "" {
			continue
		}
		if edge.IsMain {
			if obj, err := u.taxonomyRepo.GetOrCreateStudio(ctx, edge.Node.Name); err == nil && obj != nil {
				studioIDs = append(studioIDs, obj.ID)
			}
		} else {
			if obj, err := u.taxonomyRepo.GetOrCreateProducer(ctx, edge.Node.Name); err == nil && obj != nil {
				producerIDs = append(producerIDs, obj.ID)
			}
		}
	}
	_ = u.animeRepo.UpdateStudios(ctx, anime.ID, studioIDs)
	_ = u.animeRepo.UpdateProducers(ctx, anime.ID, producerIDs)

	return nil
}

func extractAniListID(resources []animethemes.ATResource) *int64 {
	for _, res := range resources {
		if strings.EqualFold(res.Site, "AniList") && res.ExternalID != nil {
			v := int64(*res.ExternalID)
			return &v
		}
	}
	return nil
}

func firstNonNilString(a, b *string) *string {
	if a != nil && strings.TrimSpace(*a) != "" {
		return a
	}
	return b
}
