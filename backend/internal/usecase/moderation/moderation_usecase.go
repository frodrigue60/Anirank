package moderation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"anirank/api/internal/domain"
	"anirank/api/internal/infrastructure"
)

func strPtr(s string) *string {
	return &s
}

type ModerationUsecase struct {
	repo                domain.ModerationRepository
	userRepo            domain.UserRepository
	songRepo            domain.SongRepository
	commentRepo         domain.CommentRepository
	notificationUsecase domain.NotificationUsecase
	mediaService        infrastructure.MediaService
}

const (
	CommentHideThreshold = 5
	UserSoftbanThreshold = 10
	TrustedReporterScore = 70
)

func NewModerationUsecase(
	repo domain.ModerationRepository,
	userRepo domain.UserRepository,
	songRepo domain.SongRepository,
	commentRepo domain.CommentRepository,
	nu domain.NotificationUsecase,
	ms infrastructure.MediaService,
) *ModerationUsecase {
	return &ModerationUsecase{
		repo:                repo,
		userRepo:            userRepo,
		songRepo:            songRepo,
		commentRepo:         commentRepo,
		notificationUsecase: nu,
		mediaService:        ms,
	}
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

	// Fetch Snapshot
	song, err := u.songRepo.GetByID(ctx, req.SongID)
	if err == nil && song != nil {
		snapshot, _ := json.Marshal(song)
		req.Snapshot = strPtr(string(snapshot))
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

	// Fetch Snapshot
	comment, err := u.commentRepo.GetByID(ctx, req.CommentID)
	if err == nil && comment != nil {
		snapshot, _ := json.Marshal(comment)
		req.Snapshot = strPtr(string(snapshot))
	}

	req.UserID = userID
	err = u.repo.CreateCommentReport(ctx, req)
	if err != nil {
		return err
	}

	// Trigger Auto-Hide Check
	u.checkCommentThreshold(ctx, req.CommentID)

	return nil
}

func (u *ModerationUsecase) checkCommentThreshold(ctx context.Context, commentID uint64) {
	// Count unique reports from trusted users
	reports, err := u.repo.GetCommentReportsCountByTrustedUsers(ctx, commentID, TrustedReporterScore)
	if err == nil && reports >= CommentHideThreshold {
		_ = u.repo.SetCommentShadowban(ctx, commentID, true)
	}
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

func (u *ModerationUsecase) ResolveSongReport(ctx context.Context, reportID uint64, isAccepted bool) error {
	report, err := u.repo.GetSongReport(ctx, reportID)
	if err != nil {
		return err
	}

	if err := u.repo.ResolveSongReport(ctx, reportID, isAccepted); err != nil {
		return err
	}

	if isAccepted {
		// Reward reporter
		u.repo.UpdateUserTruthScore(ctx, report.UserID, 5)
		// Shadowban the rating if it exists
		// Note: Song reports might be for the song itself or a specific interaction.
		// For now, we assume if it's accepted, we might want to shadowban the user's rating on that song.
		u.repo.SetRatingShadowban(ctx, report.ID, true) // This is a bit simplified, usually you'd have a specific rating ID
	} else {
		// Penalty for false report
		u.repo.UpdateUserTruthScore(ctx, report.UserID, -5)
		u.checkAndApplyShadowban(ctx, report.UserID)
	}

	return nil
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

func (u *ModerationUsecase) ResolveCommentReport(ctx context.Context, reportID uint64, isAccepted bool) error {
	report, err := u.repo.GetCommentReport(ctx, reportID)
	if err != nil {
		return err
	}

	if err := u.repo.ResolveCommentReport(ctx, reportID, isAccepted); err != nil {
		return err
	}

	if isAccepted {
		// Reward reporter
		u.repo.UpdateUserTruthScore(ctx, report.UserID, 5)

		// Penalty for comment author
		if report.Comment != nil {
			u.repo.UpdateUserTruthScore(ctx, report.Comment.UserID, -10)
			u.checkAndApplyShadowban(ctx, report.Comment.UserID)
			// Shadowban the comment itself
			u.repo.SetCommentShadowban(ctx, report.CommentID, true)
		}
	} else {
		u.repo.UpdateUserTruthScore(ctx, report.UserID, -5)
		u.checkAndApplyShadowban(ctx, report.UserID)
	}

	return nil
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
		_ = u.notificationUsecase.Create(ctx, notif)
	}

	return nil
}

func (u *ModerationUsecase) DeleteUserRequest(ctx context.Context, requestID uint64) error {
	return u.repo.DeleteUserRequest(ctx, requestID)
}

// ---- User Reports (Create & Manage) ----

func (u *ModerationUsecase) CreateUserReport(ctx context.Context, userID uint64, req *domain.UserReport) error {
	if req.ReportedUserID == 0 {
		return domain.NewAppError(400, "reported_user_id is required", nil)
	}
	if req.ReportedUserID == userID {
		return domain.NewAppError(400, "You cannot report yourself", nil)
	}
	// Check for existing pending report
	reported, err := u.repo.IsUserReportedByReporter(ctx, userID, req.ReportedUserID)
	if err == nil && reported {
		return domain.NewAppError(409, "You have already reported this user", nil)
	}
	if req.Reason == "" {
		return domain.NewAppError(400, "reason is required", nil)
	}
	if req.Source == "" {
		req.Source = "web"
	}
	if len(req.Content) > 1000 {
		return domain.NewAppError(400, "content cannot exceed 1000 characters", nil)
	}

	// Fetch Snapshot
	user, err := u.userRepo.GetByID(ctx, req.ReportedUserID)
	if err == nil && user != nil {
		snapshot, _ := json.Marshal(user)
		req.Snapshot = strPtr(string(snapshot))
	}

	req.ReporterUserID = userID
	err = u.repo.CreateUserReport(ctx, req)
	if err != nil {
		return err
	}

	// Trigger Auto-Hide / Softban Check
	u.checkUserThreshold(ctx, req.ReportedUserID)

	return nil
}

func (u *ModerationUsecase) checkUserThreshold(ctx context.Context, userID uint64) {
	// Count unique reports from trusted users
	reports, err := u.repo.GetUserReportsCountByTrustedUsers(ctx, userID, TrustedReporterScore)
	if err == nil && reports >= UserSoftbanThreshold {
		_ = u.userRepo.UpdateSoftbanStatus(ctx, userID, true)
	}
}

func (u *ModerationUsecase) GetUserReports(ctx context.Context, status *bool, limit, offset int) ([]domain.UserReport, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	reports, err := u.repo.GetUserReports(ctx, status, limit, offset)
	if err != nil {
		return nil, err
	}

	for i := range reports {
		u.resolveUserAvatars(&reports[i])
	}
	return reports, nil
}

func (u *ModerationUsecase) GetUserReport(ctx context.Context, reportID uint64) (*domain.UserReport, error) {
	report, err := u.repo.GetUserReport(ctx, reportID)
	if err != nil {
		return nil, err
	}
	u.resolveUserAvatars(report)
	return report, nil
}

func (u *ModerationUsecase) resolveUserAvatars(report *domain.UserReport) {
	if report.ReportedUser != nil {
		report.ReportedUser.AvatarUrl = u.mediaService.Resolve(report.ReportedUser.Avatar)
	}
	if report.ReporterUser != nil {
		report.ReporterUser.AvatarUrl = u.mediaService.Resolve(report.ReporterUser.Avatar)
	}
}

func (u *ModerationUsecase) ResolveUserReport(ctx context.Context, reportID uint64, isAccepted bool) error {
	report, err := u.repo.GetUserReport(ctx, reportID)
	if err != nil {
		return err
	}

	if err := u.repo.ResolveUserReport(ctx, reportID, isAccepted); err != nil {
		return err
	}

	if isAccepted {
		// Reward reporter
		u.repo.UpdateUserTruthScore(ctx, report.ReporterUserID, 5)
		// Penalty for reported user
		u.repo.UpdateUserTruthScore(ctx, report.ReportedUserID, -10)
		u.checkAndApplyShadowban(ctx, report.ReportedUserID)
	} else {
		// Penalty for reporter (false report)
		u.repo.UpdateUserTruthScore(ctx, report.ReporterUserID, -5)
		u.checkAndApplyShadowban(ctx, report.ReporterUserID)
	}

	return nil
}

func (u *ModerationUsecase) checkAndApplyShadowban(ctx context.Context, userID uint64) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return
	}

	// 1. Global Shadowban logic (TruthScore < 50)
	if user.TruthScore < 50 && !user.IsShadowbanned {
		u.repo.ShadowbanUser(ctx, userID)
	}

	// 2. Softban logic (TruthScore < 30 AND PendingReports > 3)
	pendingCount, err := u.repo.GetPendingReportsCount(ctx, userID)
	if err == nil {
		if user.TruthScore < 30 && pendingCount > 3 && !user.IsSoftbanned {
			u.userRepo.UpdateSoftbanStatus(ctx, userID, true)
		} else if user.TruthScore >= 40 && user.IsSoftbanned {
			// Auto-lift softban if reputation recovers
			u.userRepo.UpdateSoftbanStatus(ctx, userID, false)
		}
	}
}

func (u *ModerationUsecase) ValidateInteraction(ctx context.Context, userID uint64, content string) (bool, error) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false, err
	}

	// 1. Softban Check
	if user.IsSoftbanned {
		return false, domain.NewAppError(403, "Your account is currently restricted due to low reputation or pending reports.", nil)
	}

	// 2. Rate Limit Check
	lastTime, err := u.userRepo.GetLastInteractionTime(ctx, userID)
	if err == nil && !lastTime.IsZero() {
		limitSeconds := 0
		if user.Level < 5 {
			limitSeconds = 120 // 2 mins
		} else if user.Level < 10 {
			limitSeconds = 60 // 1 min
		}

		if limitSeconds > 0 && time.Since(lastTime).Seconds() < float64(limitSeconds) {
			remaining := limitSeconds - int(time.Since(lastTime).Seconds())
			return false, domain.NewAppError(429, fmt.Sprintf("You are posting too fast. Please wait %d seconds.", remaining), nil)
		}
	}

	// 3. Link Detection
	hasLinks := strings.Contains(content, "http://") || strings.Contains(content, "https://") || strings.Contains(content, "www.")
	if hasLinks {
		if user.Level < 5 {
			return false, domain.NewAppError(400, "Users below level 5 are not allowed to post links.", nil)
		} else if user.Level < 10 {
			// Auto-shadowban: Allow post but mark as shadowbanned
			return true, nil
		}
	}

	return false, nil
}

func (u *ModerationUsecase) DeleteUserReport(ctx context.Context, reportID uint64) error {
	return u.repo.DeleteUserReport(ctx, reportID)
}
