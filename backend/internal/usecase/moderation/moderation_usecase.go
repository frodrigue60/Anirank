package moderation

import (
	"context"

	"anirank/api/internal/domain"
	"encoding/json"
)

type ModerationUsecase struct {
	repo             domain.ModerationRepository
	notificationRepo domain.NotificationRepository
}

func NewModerationUsecase(repo domain.ModerationRepository, nr domain.NotificationRepository) *ModerationUsecase {
	return &ModerationUsecase{repo: repo, notificationRepo: nr}
}

// ---- User Facing (Create) ----

func (u *ModerationUsecase) CreateSongReport(ctx context.Context, userID uint64, req *domain.SongReport) error {
	if req.SongID == 0 {
		return domain.NewAppError(400, "song_id is required", nil)
	}
	// Check for existing report
	reported, err := u.repo.IsSongReportedByUser(ctx, userID, req.SongID)
	if err == nil && reported {
		return domain.NewAppError(409, "You have already reported this song", nil)
	}
	if req.Title == "" {
		return domain.NewAppError(400, "title is required", nil)
	}
	if req.Source == "" {
		req.Source = "web"
	}
	if len(req.Content) > 1000 {
		return domain.NewAppError(400, "content cannot exceed 1000 characters", nil)
	}

	req.UserID = userID
	return u.repo.CreateSongReport(ctx, req)
}

func (u *ModerationUsecase) CreateUserRequest(ctx context.Context, userID uint64, req *domain.UserRequest) error {
	if req.Content == "" {
		return domain.NewAppError(400, "content is required", nil)
	}
	if len(req.Content) > 2000 {
		return domain.NewAppError(400, "content cannot exceed 2000 characters", nil)
	}

	req.UserID = userID
	return u.repo.CreateUserRequest(ctx, req)
}

func (u *ModerationUsecase) CreateCommentReport(ctx context.Context, userID uint64, req *domain.CommentReport) error {
	if req.CommentID == 0 {
		return domain.NewAppError(400, "comment_id is required", nil)
	}
	if req.Title == "" {
		return domain.NewAppError(400, "title is required", nil)
	}
	if req.Source == "" {
		req.Source = "web"
	}
	if len(req.Content) > 1000 {
		return domain.NewAppError(400, "content cannot exceed 1000 characters", nil)
	}

	req.UserID = userID
	return u.repo.CreateCommentReport(ctx, req)
}

// ---- Admin Facing (Read & Update) ----

func (u *ModerationUsecase) GetSongReports(ctx context.Context, status *bool, limit, offset int) ([]domain.SongReport, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return u.repo.GetSongReports(ctx, status, limit, offset)
}

func (u *ModerationUsecase) GetSongReport(ctx context.Context, reportID uint64) (*domain.SongReport, error) {
	return u.repo.GetSongReport(ctx, reportID)
}

func (u *ModerationUsecase) ResolveSongReport(ctx context.Context, reportID uint64) error {
	return u.repo.ResolveSongReport(ctx, reportID)
}

func (u *ModerationUsecase) DeleteSongReport(ctx context.Context, reportID uint64) error {
	return u.repo.DeleteSongReport(ctx, reportID)
}

func (u *ModerationUsecase) GetCommentReports(ctx context.Context, status *bool, limit, offset int) ([]domain.CommentReport, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return u.repo.GetCommentReports(ctx, status, limit, offset)
}

func (u *ModerationUsecase) GetCommentReport(ctx context.Context, reportID uint64) (*domain.CommentReport, error) {
	return u.repo.GetCommentReport(ctx, reportID)
}

func (u *ModerationUsecase) ResolveCommentReport(ctx context.Context, reportID uint64) error {
	return u.repo.ResolveCommentReport(ctx, reportID)
}

func (u *ModerationUsecase) DeleteCommentReport(ctx context.Context, reportID uint64) error {
	return u.repo.DeleteCommentReport(ctx, reportID)
}

func (u *ModerationUsecase) GetUserRequests(ctx context.Context, status *bool, limit, offset int) ([]domain.UserRequest, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return u.repo.GetUserRequests(ctx, status, limit, offset)
}

func (u *ModerationUsecase) GetUserRequest(ctx context.Context, requestID uint64) (*domain.UserRequest, error) {
	return u.repo.GetUserRequest(ctx, requestID)
}

func (u *ModerationUsecase) UpdateUserRequestStatus(ctx context.Context, requestID uint64, status bool, adminID uint64) error {
	err := u.repo.UpdateUserRequestStatus(ctx, requestID, status, adminID)
	if err != nil {
		return err
	}

	// Trigger Notification
	req, err := u.repo.GetUserRequest(ctx, requestID)
	if err == nil && req != nil {
		subjectType := "user_request"
		dataObj := map[string]interface{}{
			"request_id": requestID,
			"status":     status,
		}
		dataJSON, _ := json.Marshal(dataObj)

		notif := &domain.Notification{
			UserID:      req.UserID,
			Type:        "user_request_feedback",
			SubjectID:   &requestID,
			SubjectType: &subjectType,
			Data:        dataJSON,
		}
		_ = u.notificationRepo.Create(ctx, notif)
	}

	return nil
}

func (u *ModerationUsecase) DeleteUserRequest(ctx context.Context, requestID uint64) error {
	return u.repo.DeleteUserRequest(ctx, requestID)
}
