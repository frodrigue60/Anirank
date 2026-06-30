package v1

import (
	"bufio"
	"context"
	"fmt"
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

func (h *AdminHandler) StreamVideoAuditProgress(c *fiber.Ctx) error {
	if h.videoAuditUsecase == nil {
		return domain.NewAppError(503, "Video audit service not available", nil)
	}

	jobID := c.Params("jobID")
	if jobID == "" {
		return domain.NewAppError(400, "jobID is required", nil)
	}

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	auditUC := h.videoAuditUsecase

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		bgCtx := context.Background()

		for range ticker.C {
			job, err := auditUC.GetJobStatus(bgCtx, jobID)
			if err != nil {
				fmt.Fprintf(w, "data: {\"error\": \"job not found\"}\n\n")
				w.Flush()
				return
			}

			payload := adminUC.MarshalJob(job)
			fmt.Fprintf(w, "data: %s\n\n", payload)
			w.Flush()

			if job.Status == domain.ImportJobDone ||
				job.Status == domain.ImportJobFailed ||
				job.Status == domain.ImportJobCanceled {
				return
			}
		}
	})

	return nil
}
