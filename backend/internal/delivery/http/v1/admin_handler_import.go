package v1

import (
	"bufio"
	"context"
	"fmt"
	"time"

	"anirank/api/internal/domain"
	adminUC "anirank/api/internal/usecase/admin"

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

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	// Capture usecase reference so the goroutine outlives the handler scope
	importUC := h.importUsecase

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		// Use a detached context because c.Context() gets recycled when the handler returns
		bgCtx := context.Background()

		for range ticker.C {
			job, err := importUC.GetJobStatus(bgCtx, jobID)
			if err != nil {
				fmt.Fprintf(w, "data: {\"error\": \"job not found\"}\n\n")
				w.Flush()
				return
			}

			payload := adminUC.MarshalJob(job)
			fmt.Fprintf(w, "data: %s\n\n", payload)
			w.Flush()

			// Stop streaming when job has reached a terminal state
			if job.Status == domain.ImportJobDone ||
				job.Status == domain.ImportJobFailed ||
				job.Status == domain.ImportJobCanceled {
				return
			}
		}
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
