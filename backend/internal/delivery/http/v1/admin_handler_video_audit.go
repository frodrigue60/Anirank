package v1

import (
	"bufio"
	"context"
	"strconv"
	"time"

	"anirank/api/internal/domain"
	"anirank/api/internal/dto"
	adminUC "anirank/api/internal/usecase/admin"

	"github.com/gofiber/fiber/v2"
)

func (h *AdminHandler) StartVideoAudit(c *fiber.Ctx) error {
	if h.videoAuditUsecase == nil {
		return domain.NewAppError(503, "Video audit service not available", nil)
	}

	var req struct {
		Prefix         string `json:"prefix"`
		IncludeOrphans bool   `json:"include_orphans"`
	}
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}

	jobID, err := h.videoAuditUsecase.StartAudit(c.Context(), adminUC.VideoAuditOptions{
		Prefix:         req.Prefix,
		IncludeOrphans: req.IncludeOrphans,
	})
	if err != nil {
		return domain.NewAppError(409, err.Error(), err)
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"job_id":  jobID,
		"message": "Video audit started. Use /admin/system/video-audit/:jobID/stream to monitor progress.",
	})
}

func (h *AdminHandler) GetLatestVideoAuditStatus(c *fiber.Ctx) error {
	if h.videoAuditUsecase == nil {
		return domain.NewAppError(503, "Video audit service not available", nil)
	}

	job, err := h.videoAuditUsecase.GetLatestJobStatus(c.Context())
	if err != nil {
		return domain.NewAppError(404, "Job not found", err)
	}

	return c.JSON(job)
}

func (h *AdminHandler) GetVideoAuditReport(c *fiber.Ctx) error {
	if h.videoAuditUsecase == nil {
		return domain.NewAppError(503, "Video audit service not available", nil)
	}

	jobID := c.Params("jobID")
	if jobID == "" {
		return domain.NewAppError(400, "jobID is required", nil)
	}

	job, err := h.videoAuditUsecase.GetJobStatus(c.Context(), jobID)
	if err != nil {
		return domain.NewAppError(404, "Job not found", err)
	}
	if !h.videoAuditUsecase.IsVideoAuditJob(job) {
		return domain.NewAppError(400, "Job is not a video audit", nil)
	}

	report, ok := h.videoAuditUsecase.GetReport(jobID)
	if !ok {
		if job.Status == domain.ImportJobRunning || job.Status == domain.ImportJobPending {
			return domain.NewAppError(409, "Audit still running", nil)
		}
		return domain.NewAppError(404, "Audit report not found", nil)
	}

	return c.JSON(fiber.Map{"data": dto.ToVideoAuditReportDTO(report)})
}

func (h *AdminHandler) CancelVideoAudit(c *fiber.Ctx) error {
	if h.videoAuditUsecase == nil {
		return domain.NewAppError(503, "Video audit service not available", nil)
	}

	jobID := c.Params("jobID")
	if jobID == "" {
		return domain.NewAppError(400, "jobID is required", nil)
	}

	job, err := h.videoAuditUsecase.GetJobStatus(c.Context(), jobID)
	if err != nil {
		return domain.NewAppError(404, "Job not found", err)
	}
	if !h.videoAuditUsecase.IsVideoAuditJob(job) {
		return domain.NewAppError(400, "Job is not a video audit", nil)
	}

	if err := h.videoAuditUsecase.CancelJob(c.Context(), jobID); err != nil {
		return domain.NewAppError(500, "Failed to cancel audit", err)
	}

	return c.JSON(fiber.Map{"message": "Video audit cancellation requested"})
}

func (h *AdminHandler) CheckSongVideoStorage(c *fiber.Ctx) error {
	if h.videoAuditUsecase == nil {
		return domain.NewAppError(503, "Video audit service not available", nil)
	}

	id, err := h.resolveID(c, "song")
	if err != nil {
		return err
	}

	results, err := h.videoAuditUsecase.CheckSongVideoStorage(c.Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"data": results})
}

func (h *AdminHandler) CheckVariantVideoStorage(c *fiber.Ctx) error {
	if h.videoAuditUsecase == nil {
		return domain.NewAppError(503, "Video audit service not available", nil)
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil || id == 0 {
		return domain.NewAppError(400, "Invalid variant ID", err)
	}

	results, err := h.videoAuditUsecase.CheckVariantVideoStorage(c.Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"data": results})
}

func (h *AdminHandler) StreamVideoAuditProgress(c *fiber.Ctx) error {
	if h.videoAuditUsecase == nil {
		return domain.NewAppError(503, "Video audit service not available", nil)
	}

	jobID := c.Params("jobID")
	if jobID == "" {
		return domain.NewAppError(400, "jobID is required", nil)
	}

	setSSEHeaders(c)

	auditUC := h.videoAuditUsecase

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		bgCtx := context.Background()
		streamImportJobSSE(w, 2*time.Second, func() (*domain.ImportJob, error) {
			return auditUC.GetJobStatus(bgCtx, jobID)
		})
	})

	return nil
}
