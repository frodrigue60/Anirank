package v1

import (
	"bufio"
	"context"
	"time"

	"anirank/api/internal/domain"

	"github.com/gofiber/fiber/v2"
)

// StartAnimeThemesImport godoc
// @Summary     Start bulk AnimeThemes import
// @Description Launches an async background job that hydrates animes, songs, variants, and artists from AnimeThemes, then enriches with AniList metadata. Returns immediately with the job ID.
// @Tags        admin,import
// @Produce     json
// @Success     202 {object} fiber.Map{"job_id": "string"}
// @Failure     409 {object} domain.AppError
// @Router      /admin/import/animethemes/start [post]
func (h *AdminHandler) StartAnimeThemesImport(c *fiber.Ctx) error {
	if h.importUsecase == nil {
		return domain.NewAppError(503, "Import service not available", nil)
	}

	jobID, err := h.importUsecase.StartAnimeThemesImport(c.Context())
	if err != nil {
		return domain.NewAppError(409, err.Error(), err)
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"job_id":  jobID,
		"message": "Import job started. Use /import/:jobID/stream or /import/:jobID/status to monitor progress.",
	})
}

// GetLatestAnimeThemesImportStatus godoc
// @Summary     Get latest animethemes import job status
// @Description Returns the most recent import job for animethemes source.
// @Tags        admin,import
// @Produce     json
// @Success     200 {object} domain.ImportJob
// @Failure     404 {object} domain.AppError
// @Router      /admin/import/animethemes/status [get]
func (h *AdminHandler) GetLatestAnimeThemesImportStatus(c *fiber.Ctx) error {
	if h.importUsecase == nil {
		return domain.NewAppError(503, "Import service not available", nil)
	}

	job, err := h.importUsecase.GetLatestJobStatus(c.Context(), "animethemes")
	if err != nil {
		return domain.NewAppError(404, "Job not found", err)
	}

	return c.JSON(job)
}

// GetImportJobStatus godoc
// @Summary     Get import job status
// @Description Returns the current progress of an import job by its ID.
// @Tags        admin,import
// @Produce     json
// @Param       jobID path string true "Import Job UUID"
// @Success     200 {object} domain.ImportJob
// @Failure     404 {object} domain.AppError
// @Router      /admin/import/{jobID}/status [get]
func (h *AdminHandler) GetImportJobStatus(c *fiber.Ctx) error {
	if h.importUsecase == nil {
		return domain.NewAppError(503, "Import service not available", nil)
	}

	jobID := c.Params("jobID")
	if jobID == "" {
		return domain.NewAppError(400, "jobID is required", nil)
	}

	job, err := h.importUsecase.GetJobStatus(c.Context(), jobID)
	if err != nil {
		return domain.NewAppError(404, "Job not found", err)
	}

	return c.JSON(fiber.Map{"data": job})
}

// CancelImportJob godoc
// @Summary     Cancel a running import job
// @Description Marks a pending or running import job as canceled.
// @Tags        admin,import
// @Produce     json
// @Param       jobID path string true "Import Job UUID"
// @Success     200 {object} fiber.Map
// @Router      /admin/import/{jobID}/cancel [post]
func (h *AdminHandler) CancelImportJob(c *fiber.Ctx) error {
	if h.importUsecase == nil {
		return domain.NewAppError(503, "Import service not available", nil)
	}

	jobID := c.Params("jobID")
	if jobID == "" {
		return domain.NewAppError(400, "jobID is required", nil)
	}

	if err := h.importUsecase.CancelJob(c.Context(), jobID); err != nil {
		return domain.NewAppError(500, "Failed to cancel job", err)
	}

	return c.JSON(fiber.Map{"message": "Job cancellation requested"})
}

// StreamImportProgress godoc
// @Summary     Stream import job progress via SSE
// @Description Server-Sent Events stream that polls the import job status and emits progress updates every 2 seconds until the job finishes.
// @Tags        admin,import
// @Produce     text/event-stream
// @Param       jobID path string true "Import Job UUID"
// @Router      /admin/import/{jobID}/stream [get]
func (h *AdminHandler) StreamImportProgress(c *fiber.Ctx) error {
	if h.importUsecase == nil {
		return domain.NewAppError(503, "Import service not available", nil)
	}

	jobID := c.Params("jobID")
	if jobID == "" {
		return domain.NewAppError(400, "jobID is required", nil)
	}

	setSSEHeaders(c)

	// Capture usecase reference so the goroutine outlives the handler scope
	importUC := h.importUsecase

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		bgCtx := context.Background()
		streamImportJobSSE(w, 2*time.Second, func() (*domain.ImportJob, error) {
			return importUC.GetJobStatus(bgCtx, jobID)
		})
	})

	return nil
}

// StartTitleBackfill godoc
// @Summary     Start title-variants backfill job
// @Description Launches an async background job that iterates over all animes with an anilist_id but missing title_english/title_native and hydrates only those title fields from AniList. Returns immediately with the job ID.
// @Tags        admin,import
// @Produce     json
// @Success     202 {object} fiber.Map{"job_id": "string"}
// @Failure     409 {object} domain.AppError
// @Router      /admin/import/backfill-titles/start [post]
func (h *AdminHandler) StartTitleBackfill(c *fiber.Ctx) error {
	if h.importUsecase == nil {
		return domain.NewAppError(503, "Import service not available", nil)
	}

	jobID, err := h.importUsecase.StartTitleBackfill(c.Context())
	if err != nil {
		return domain.NewAppError(409, err.Error(), err)
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"job_id":  jobID,
		"message": "Title backfill job started. Use /import/:jobID/stream or /import/:jobID/status to monitor progress.",
	})
}

// GetLatestTitleBackfillStatus godoc
// @Summary     Get latest title backfill job status
// @Description Returns the most recent title-backfill import job.
// @Tags        admin,import
// @Produce     json
// @Success     200 {object} domain.ImportJob
// @Failure     404 {object} domain.AppError
// @Router      /admin/import/backfill-titles/status [get]
func (h *AdminHandler) GetLatestTitleBackfillStatus(c *fiber.Ctx) error {
	if h.importUsecase == nil {
		return domain.NewAppError(503, "Import service not available", nil)
	}

	job, err := h.importUsecase.GetLatestJobStatus(c.Context(), "backfill_titles")
	if err != nil {
		return domain.NewAppError(404, "No backfill job found", err)
	}

	return c.JSON(job)
}

// StartIncrementalSongSync godoc
// @Summary     Start incremental AnimeThemes song sync
// @Description Imports only songs with AnimeThemes IDs greater than the local watermark (MAX songs.anime_themes_id), then upserts parent anime and variants.
// @Tags        admin,import
// @Produce     json
// @Success     202 {object} fiber.Map{"job_id": "string"}
// @Failure     409 {object} domain.AppError
// @Router      /admin/import/animethemes/incremental/start [post]
func (h *AdminHandler) StartIncrementalSongSync(c *fiber.Ctx) error {
	if h.importUsecase == nil {
		return domain.NewAppError(503, "Import service not available", nil)
	}

	jobID, err := h.importUsecase.StartIncrementalSongSync(c.Context())
	if err != nil {
		return domain.NewAppError(409, err.Error(), err)
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"job_id":  jobID,
		"message": "Incremental song sync started. Use /import/:jobID/stream or /import/:jobID/status to monitor progress.",
	})
}

// GetLatestIncrementalSongSyncStatus godoc
// @Summary     Get latest incremental song sync status
// @Description Returns the most recent at_song_incremental import job.
// @Tags        admin,import
// @Produce     json
// @Success     200 {object} domain.ImportJob
// @Failure     404 {object} domain.AppError
// @Router      /admin/import/animethemes/incremental/status [get]
func (h *AdminHandler) GetLatestIncrementalSongSyncStatus(c *fiber.Ctx) error {
	if h.importUsecase == nil {
		return domain.NewAppError(503, "Import service not available", nil)
	}

	job, err := h.importUsecase.GetLatestJobStatus(c.Context(), "at_song_incremental")
	if err != nil {
		return domain.NewAppError(404, "No incremental sync job found", err)
	}

	return c.JSON(job)
}
