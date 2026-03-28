package v1

import (
	"strconv"

	"anirank/api/internal/domain"
	"anirank/api/internal/usecase/moderation"

	"github.com/gofiber/fiber/v2"
)

type ModerationHandler struct {
	usecase     *moderation.ModerationUsecase
	songRepo    domain.SongRepository
	commentRepo domain.CommentRepository
}

func NewModerationHandler(
	usecase *moderation.ModerationUsecase,
	songRepo domain.SongRepository,
	commentRepo domain.CommentRepository,
) *ModerationHandler {
	return &ModerationHandler{
		usecase:     usecase,
		songRepo:    songRepo,
		commentRepo: commentRepo,
	}
}

// ==== USER FACING ENDPOINTS ====

// CreateSongReport
// @Summary Create a Song Report
// @Description Submit a report for a specific song as a user.
// @Tags Moderation
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body domain.SongReport true "Song Report Data"
// @Success 201 {object} object{message=string}
// @Failure 400 {object} domain.AppError
// @Failure 401 {object} domain.AppError
// @Router /songs/reports [post]
func (h *ModerationHandler) CreateSongReport(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint64)
	if !ok {
		return domain.NewAppError(401, "Unauthorized", nil)
	}

	type songReportReq struct {
		SongID  string `json:"song_id"`
		Title   string `json:"title"`
		Content string `json:"content"`
		Source  string `json:"source"`
	}

	var body songReportReq
	if err := c.BodyParser(&body); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}

	// Resolve Song ID
	songID, err := strconv.ParseUint(body.SongID, 10, 64)
	if err != nil {
		song, err := h.songRepo.GetByUUID(c.Context(), body.SongID)
		if err != nil {
			return domain.NewAppError(404, "Song not found", err)
		}
		songID = song.ID
	}

	req := domain.SongReport{
		SongID:  songID,
		UserID:  userID,
		Title:   body.Title,
		Content: body.Content,
		Source:  body.Source,
	}

	if err := h.usecase.CreateSongReport(c.Context(), userID, &req); err != nil {
		return err
	}

	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"message": "Report submitted successfully",
		},
	})
}

// CreateUserRequest
// @Summary Create a User Request
// @Description Submit a general support request (E.g request an Artist to be added).
// @Tags Moderation
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body domain.UserRequest true "User Request Data"
// @Success 201 {object} object{message=string}
// @Failure 400 {object} domain.AppError
// @Failure 401 {object} domain.AppError
// @Router /user-requests [post]
func (h *ModerationHandler) CreateUserRequest(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint64)
	if !ok {
		return domain.NewAppError(401, "Unauthorized", nil)
	}

	var req domain.UserRequest
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}

	if err := h.usecase.CreateUserRequest(c.Context(), userID, &req); err != nil {
		return err
	}

	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"message": "Request submitted successfully",
		},
	})
}

// CreateCommentReport
// @Summary Create a Comment Report
// @Description Submit a report for a specific comment as a user.
// @Tags Moderation
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body domain.CommentReport true "Comment Report Data"
// @Success 201 {object} object{message=string}
// @Failure 400 {object} domain.AppError
// @Failure 401 {object} domain.AppError
// @Router /comments/reports [post]
func (h *ModerationHandler) CreateCommentReport(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint64)
	if !ok {
		return domain.NewAppError(401, "Unauthorized", nil)
	}

	type commentReportReq struct {
		CommentID string `json:"comment_id"`
		Title     string `json:"title"`
		Content   string `json:"content"`
		Source    string `json:"source"`
	}

	var body commentReportReq
	if err := c.BodyParser(&body); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}

	// Resolve Comment ID
	commentID, err := strconv.ParseUint(body.CommentID, 10, 64)
	if err != nil {
		comment, err := h.commentRepo.GetByUUID(c.Context(), body.CommentID)
		if err != nil {
			return domain.NewAppError(404, "Comment not found", err)
		}
		commentID = comment.ID
	}

	req := domain.CommentReport{
		CommentID: commentID,
		UserID:    userID,
		Title:     body.Title,
		Content:   body.Content,
		Source:    body.Source,
	}

	if err := h.usecase.CreateCommentReport(c.Context(), userID, &req); err != nil {
		return err
	}

	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"message": "Report submitted successfully",
		},
	})
}

// ==== ADMIN FACING ENDPOINTS ====

// GetSongReports
// @Summary List Song Reports
// @Description Admin endpoint to list song reports, filtered by status.
// @Tags Admin Moderation
// @Security BearerAuth
// @Produce json
// @Param status query string false "Status (pending, fixed)" default(pending)
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} domain.SongReport
// @Router /admin/songs/reports [get]
func (h *ModerationHandler) GetSongReports(c *fiber.Ctx) error {
	status := c.Query("status", "pending")
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	var appStatus *bool
	if status == "fixed" || status == "resolved" {
		t := true
		appStatus = &t
	} else if status == "pending" {
		f := false
		appStatus = &f
	}

	reports, err := h.usecase.GetSongReports(c.Context(), appStatus, limit, offset)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"data": reports})
}

// GetSongReport
// @Summary Get Song Report Details
// @Description Admin endpoint to get a specific song report.
// @Tags Admin Moderation
// @Security BearerAuth
// @Produce json
// @Param id path int true "Report ID"
// @Success 200 {object} domain.SongReport
// @Router /admin/songs/reports/{id} [get]
func (h *ModerationHandler) GetSongReport(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return domain.NewAppError(400, "Invalid ID format", err)
	}

	report, err := h.usecase.GetSongReport(c.Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"data": report})
}

// ResolveSongReport
// @Summary Resolve Song Report
// @Description Admin endpoint to mark a song report as resolved.
// @Tags Admin Moderation
// @Security BearerAuth
// @Produce json
// @Param id path int true "Report ID"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} domain.AppError
// @Failure 401 {object} domain.AppError
// @Failure 403 {object} domain.AppError
// @Router /admin/songs/reports/{id}/resolve [put]
func (h *ModerationHandler) ResolveSongReport(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return domain.NewAppError(400, "Invalid ID", nil)
	}

	if err := h.usecase.ResolveSongReport(c.Context(), id); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"message": "Report resolved successfully",
		},
	})
}

// DeleteSongReport
// @Summary Delete Song Report
// @Description Admin endpoint to completely delete a song report.
// @Tags Admin Moderation
// @Security BearerAuth
// @Produce json
// @Param id path int true "Report ID"
// @Success 204
// @Failure 400 {object} domain.AppError
// @Failure 401 {object} domain.AppError
// @Router /admin/songs/reports/{id} [delete]
func (h *ModerationHandler) DeleteSongReport(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return domain.NewAppError(400, "Invalid ID", nil)
	}

	if err := h.usecase.DeleteSongReport(c.Context(), id); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Song report deleted successfully",
		"data":    nil,
	})
}

// GetCommentReports
// @Summary List Comment Reports
// @Description Admin endpoint to list comment reports, filtered by status.
// @Tags Admin Moderation
// @Security BearerAuth
// @Produce json
// @Param status query string false "Status (pending, fixed)" default(pending)
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} domain.CommentReport
// @Router /admin/comments/reports [get]
func (h *ModerationHandler) GetCommentReports(c *fiber.Ctx) error {
	status := c.Query("status", "pending")
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	var appStatus *bool
	if status == "fixed" || status == "resolved" {
		t := true
		appStatus = &t
	} else if status == "pending" {
		f := false
		appStatus = &f
	}

	reports, err := h.usecase.GetCommentReports(c.Context(), appStatus, limit, offset)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"data": reports})
}

// GetCommentReport
// @Summary Get Comment Report Details
// @Description Admin endpoint to get a specific comment report.
// @Tags Admin Moderation
// @Security BearerAuth
// @Produce json
// @Param id path int true "Report ID"
// @Success 200 {object} domain.CommentReport
// @Router /admin/comments/reports/{id} [get]
func (h *ModerationHandler) GetCommentReport(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return domain.NewAppError(400, "Invalid ID format", err)
	}

	report, err := h.usecase.GetCommentReport(c.Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"data": report})
}

// ResolveCommentReport
// @Summary Resolve Comment Report
// @Description Admin endpoint to mark a comment report as resolved.
// @Tags Admin Moderation
// @Security BearerAuth
// @Produce json
// @Param id path int true "Report ID"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} domain.AppError
// @Failure 401 {object} domain.AppError
// @Failure 403 {object} domain.AppError
// @Router /admin/comments/reports/{id}/resolve [put]
func (h *ModerationHandler) ResolveCommentReport(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return domain.NewAppError(400, "Invalid ID", nil)
	}

	if err := h.usecase.ResolveCommentReport(c.Context(), id); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"message": "Report resolved successfully",
		},
	})
}

// DeleteCommentReport
// @Summary Delete Comment Report
// @Description Admin endpoint to completely delete a comment report.
// @Tags Admin Moderation
// @Security BearerAuth
// @Produce json
// @Param id path int true "Report ID"
// @Success 204
// @Failure 400 {object} domain.AppError
// @Failure 401 {object} domain.AppError
// @Router /admin/comments/reports/{id} [delete]
func (h *ModerationHandler) DeleteCommentReport(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return domain.NewAppError(400, "Invalid ID", nil)
	}

	if err := h.usecase.DeleteCommentReport(c.Context(), id); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Comment report deleted successfully",
		"data":    nil,
	})
}

// GetUserRequests
// @Summary List User Requests
// @Description Admin endpoint to list all user requests, filtered by status.
// @Tags Admin Moderation
// @Security BearerAuth
// @Produce json
// @Param status query string false "Status (pending, attended)" default(pending)
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} domain.UserRequest
// @Router /admin/user-requests [get]
func (h *ModerationHandler) GetUserRequests(c *fiber.Ctx) error {
	status := c.Query("status", "pending")
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	var appStatus *bool
	if status == "attended" {
		t := true
		appStatus = &t
	} else if status == "pending" {
		f := false
		appStatus = &f
	}

	reqs, err := h.usecase.GetUserRequests(c.Context(), appStatus, limit, offset)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"data": reqs})
}

// GetUserRequest
// @Summary Get User Request
// @Description Admin endpoint to get a specific user request.
// @Tags Admin Moderation
// @Security BearerAuth
// @Produce json
// @Param id path int true "Request ID"
// @Success 200 {object} domain.UserRequest
// @Router /admin/user-requests/{id} [get]
func (h *ModerationHandler) GetUserRequest(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return domain.NewAppError(400, "Invalid ID format", err)
	}

	req, err := h.usecase.GetUserRequest(c.Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"data": req})
}

// UpdateUserRequestStatus
// @Summary Update User Request Status
// @Description Admin endpoint to update user request status (pending or attended)
// @Tags Admin Moderation
// @Security BearerAuth
// @Produce json
// @Param id path int true "Request ID"
// @Param body body object{status=string} true "Status update"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} domain.AppError
// @Failure 401 {object} domain.AppError
// @Router /admin/user-requests/{id}/status [patch]
func (h *ModerationHandler) UpdateUserRequestStatus(c *fiber.Ctx) error {
	adminID, ok := c.Locals("user_id").(uint64)
	if !ok {
		return domain.NewAppError(401, "Unauthorized or could not get Admin ID context", nil)
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return domain.NewAppError(400, "Invalid ID", nil)
	}

	type RequestBody struct {
		Status string `json:"status"`
	}
	var req RequestBody
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}

	isAttended := req.Status == "attended"
	if err := h.usecase.UpdateUserRequestStatus(c.Context(), id, isAttended, adminID); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"message": "Status updated successfully",
		},
	})
}

// DeleteUserRequest
// @Summary Delete User Request
// @Description Admin endpoint to completely delete a user request.
// @Tags Admin Moderation
// @Security BearerAuth
// @Produce json
// @Param id path int true "Request ID"
// @Success 204
// @Router /admin/user-requests/{id} [delete]
func (h *ModerationHandler) DeleteUserRequest(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return domain.NewAppError(400, "Invalid ID format", err)
	}

	if err := h.usecase.DeleteUserRequest(c.Context(), id); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "User request deleted successfully",
		"data":    nil,
	})
}
