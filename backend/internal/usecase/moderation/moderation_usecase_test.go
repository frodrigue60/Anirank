package moderation_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"anirank/api/internal/domain"
	"anirank/api/internal/testutil"
	"anirank/api/internal/usecase/moderation"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockModerationRepository implements domain.ModerationRepository
type MockModerationRepository struct {
	mock.Mock
}

func (m *MockModerationRepository) CreateSongReport(ctx context.Context, report *domain.SongReport) error {
	args := m.Called(ctx, report)
	return args.Error(0)
}
func (m *MockModerationRepository) IsSongReportedByUser(ctx context.Context, userID, songID uint64) (bool, error) {
	args := m.Called(ctx, userID, songID)
	return args.Bool(0), args.Error(1)
}
func (m *MockModerationRepository) GetSongReportsByUserAndSongIDs(ctx context.Context, userID uint64, songIDs []uint64) (map[uint64]bool, error) {
	args := m.Called(ctx, userID, songIDs)
	return args.Get(0).(map[uint64]bool), args.Error(1)
}
func (m *MockModerationRepository) CreateCommentReport(ctx context.Context, report *domain.CommentReport) error {
	args := m.Called(ctx, report)
	return args.Error(0)
}
func (m *MockModerationRepository) CreateUserRequest(ctx context.Context, request *domain.UserRequest) error {
	args := m.Called(ctx, request)
	return args.Error(0)
}
func (m *MockModerationRepository) GetSongReports(ctx context.Context, status *bool, limit, offset int) ([]domain.SongReport, error) {
	args := m.Called(ctx, status, limit, offset)
	return args.Get(0).([]domain.SongReport), args.Error(1)
}
func (m *MockModerationRepository) GetSongReport(ctx context.Context, reportID uint64) (*domain.SongReport, error) {
	args := m.Called(ctx, reportID)
	return args.Get(0).(*domain.SongReport), args.Error(1)
}
func (m *MockModerationRepository) ResolveSongReport(ctx context.Context, reportID uint64, isAccepted bool) error {
	args := m.Called(ctx, reportID, isAccepted)
	return args.Error(0)
}
func (m *MockModerationRepository) DeleteSongReport(ctx context.Context, reportID uint64) error {
	args := m.Called(ctx, reportID)
	return args.Error(0)
}
func (m *MockModerationRepository) GetCommentReports(ctx context.Context, status *bool, limit, offset int) ([]domain.CommentReport, error) {
	args := m.Called(ctx, status, limit, offset)
	return args.Get(0).([]domain.CommentReport), args.Error(1)
}
func (m *MockModerationRepository) GetCommentReport(ctx context.Context, reportID uint64) (*domain.CommentReport, error) {
	args := m.Called(ctx, reportID)
	return args.Get(0).(*domain.CommentReport), args.Error(1)
}
func (m *MockModerationRepository) ResolveCommentReport(ctx context.Context, reportID uint64, isAccepted bool) error {
	args := m.Called(ctx, reportID, isAccepted)
	return args.Error(0)
}
func (m *MockModerationRepository) DeleteCommentReport(ctx context.Context, reportID uint64) error {
	args := m.Called(ctx, reportID)
	return args.Error(0)
}
func (m *MockModerationRepository) GetUserRequests(ctx context.Context, status *bool, limit, offset int) ([]domain.UserRequest, error) {
	args := m.Called(ctx, status, limit, offset)
	return args.Get(0).([]domain.UserRequest), args.Error(1)
}
func (m *MockModerationRepository) GetUserRequest(ctx context.Context, requestID uint64) (*domain.UserRequest, error) {
	args := m.Called(ctx, requestID)
	return args.Get(0).(*domain.UserRequest), args.Error(1)
}
func (m *MockModerationRepository) UpdateUserRequestStatus(ctx context.Context, requestID uint64, status bool, adminID uint64) error {
	args := m.Called(ctx, requestID, status, adminID)
	return args.Error(0)
}
func (m *MockModerationRepository) DeleteUserRequest(ctx context.Context, requestID uint64) error {
	args := m.Called(ctx, requestID)
	return args.Error(0)
}
func (m *MockModerationRepository) CreateUserReport(ctx context.Context, report *domain.UserReport) error {
	args := m.Called(ctx, report)
	return args.Error(0)
}
func (m *MockModerationRepository) IsUserReportedByReporter(ctx context.Context, reporterID, reportedID uint64) (bool, error) {
	args := m.Called(ctx, reporterID, reportedID)
	return args.Bool(0), args.Error(1)
}
func (m *MockModerationRepository) GetUserReports(ctx context.Context, status *bool, limit, offset int) ([]domain.UserReport, error) {
	args := m.Called(ctx, status, limit, offset)
	return args.Get(0).([]domain.UserReport), args.Error(1)
}
func (m *MockModerationRepository) GetUserReport(ctx context.Context, reportID uint64) (*domain.UserReport, error) {
	args := m.Called(ctx, reportID)
	return args.Get(0).(*domain.UserReport), args.Error(1)
}
func (m *MockModerationRepository) ResolveUserReport(ctx context.Context, reportID uint64, isAccepted bool) error {
	args := m.Called(ctx, reportID, isAccepted)
	return args.Error(0)
}
func (m *MockModerationRepository) DeleteUserReport(ctx context.Context, reportID uint64) error {
	args := m.Called(ctx, reportID)
	return args.Error(0)
}
func (m *MockModerationRepository) ShadowbanUser(ctx context.Context, userID uint64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}
func (m *MockModerationRepository) UnshadowbanUser(ctx context.Context, userID uint64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}
func (m *MockModerationRepository) SetCommentShadowban(ctx context.Context, commentID uint64, isShadowbanned bool) error {
	args := m.Called(ctx, commentID, isShadowbanned)
	return args.Error(0)
}
func (m *MockModerationRepository) SetRatingShadowban(ctx context.Context, ratingID uint64, isShadowbanned bool) error {
	args := m.Called(ctx, ratingID, isShadowbanned)
	return args.Error(0)
}
func (m *MockModerationRepository) UpdateUserTruthScore(ctx context.Context, userID uint64, delta int) error {
	args := m.Called(ctx, userID, delta)
	return args.Error(0)
}
func (m *MockModerationRepository) GetPendingReportsCount(ctx context.Context, userID uint64) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}
func (m *MockModerationRepository) GetCommentReportsCountByTrustedUsers(ctx context.Context, commentID uint64, minScore int) (int, error) {
	args := m.Called(ctx, commentID, minScore)
	return args.Int(0), args.Error(1)
}
func (m *MockModerationRepository) GetUserReportsCountByTrustedUsers(ctx context.Context, reportedUserID uint64, minScore int) (int, error) {
	args := m.Called(ctx, reportedUserID, minScore)
	return args.Int(0), args.Error(1)
}

// MockNotificationUsecase implements domain.NotificationUsecase
type MockNotificationUsecase struct {
	mock.Mock
}

func (m *MockNotificationUsecase) Create(ctx context.Context, n *domain.Notification) error {
	args := m.Called(ctx, n)
	return args.Error(0)
}
func (m *MockNotificationUsecase) GetNotifications(ctx context.Context, userID uint64, nType string, limit, offset int) ([]domain.Notification, int, int, error) {
	return nil, 0, 0, nil
}
func (m *MockNotificationUsecase) MarkAsRead(ctx context.Context, id string, userID uint64) error { return nil }
func (m *MockNotificationUsecase) MarkAllAsRead(ctx context.Context, userID uint64) error         { return nil }
func (m *MockNotificationUsecase) Delete(ctx context.Context, id string, userID uint64) error    { return nil }
func (m *MockNotificationUsecase) GetUnreadCount(ctx context.Context, userID uint64) (int, error) {
	return 0, nil
}
func (m *MockNotificationUsecase) SubscribeToStream(ctx context.Context, userID uint64) (domain.Subscriber, error) {
	return nil, nil
}
func (m *MockNotificationUsecase) GetSettings(ctx context.Context, userID uint64) (*domain.NotificationSettings, error) {
	return nil, nil
}
func (m *MockNotificationUsecase) UpdateSettings(ctx context.Context, userID uint64, settings json.RawMessage) error {
	return nil
}

func TestModerationUsecase_ValidateInteraction(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockModerationRepository)
	mockUserRepo := new(testutil.MockUserRepository)
	mockSongRepo := new(testutil.MockSongRepository)
	mockCommentRepo := new(testutil.MockCommentRepository)
	mockNotif := new(MockNotificationUsecase)
	mockMedia := new(testutil.MockMediaService)

	uc := moderation.NewModerationUsecase(mockRepo, mockUserRepo, mockSongRepo, mockCommentRepo, mockNotif, mockMedia)

	t.Run("Blocked by Softban", func(t *testing.T) {
		userID := uint64(1)
		mockUserRepo.GetByIDFunc = func(id uint64) (*domain.User, error) {
			return &domain.User{ID: id, IsSoftbanned: true}, nil
		}

		isShadow, err := uc.ValidateInteraction(ctx, userID, "hello")
		assert.Error(t, err)
		assert.False(t, isShadow)
		assert.Contains(t, err.Error(), "restricted")
	})

	t.Run("Rate Limit - Level 1", func(t *testing.T) {
		userID := uint64(2)
		mockUserRepo.GetByIDFunc = func(id uint64) (*domain.User, error) {
			return &domain.User{ID: id, Level: 1}, nil
		}
		// Last interaction 30 seconds ago (limit is 120)
		mockUserRepo.GetLastInteractionTimeFunc = func(uID uint64) (time.Time, error) {
			return time.Now().Add(-30 * time.Second), nil
		}

		isShadow, err := uc.ValidateInteraction(ctx, userID, "hello")
		assert.Error(t, err)
		assert.False(t, isShadow)
		assert.Contains(t, err.Error(), "too fast")
	})

	t.Run("Rate Limit Pass - Level 1", func(t *testing.T) {
		userID := uint64(3)
		mockUserRepo.GetByIDFunc = func(id uint64) (*domain.User, error) {
			return &domain.User{ID: id, Level: 1}, nil
		}
		// Last interaction 150 seconds ago
		mockUserRepo.GetLastInteractionTimeFunc = func(uID uint64) (time.Time, error) {
			return time.Now().Add(-150 * time.Second), nil
		}

		isShadow, err := uc.ValidateInteraction(ctx, userID, "hello")
		assert.NoError(t, err)
		assert.False(t, isShadow)
	})

	t.Run("Link Blocked - Level 1", func(t *testing.T) {
		userID := uint64(4)
		mockUserRepo.GetByIDFunc = func(id uint64) (*domain.User, error) {
			return &domain.User{ID: id, Level: 1}, nil
		}
		mockUserRepo.GetLastInteractionTimeFunc = func(uID uint64) (time.Time, error) {
			return time.Time{}, nil
		}

		isShadow, err := uc.ValidateInteraction(ctx, userID, "check this http://malicious.com")
		assert.Error(t, err)
		assert.False(t, isShadow)
		assert.Contains(t, err.Error(), "not allowed to post links")
	})

	t.Run("Link Shadowbanned - Level 6", func(t *testing.T) {
		userID := uint64(5)
		mockUserRepo.GetByIDFunc = func(id uint64) (*domain.User, error) {
			return &domain.User{ID: id, Level: 6}, nil
		}
		mockUserRepo.GetLastInteractionTimeFunc = func(uID uint64) (time.Time, error) {
			return time.Time{}, nil
		}

		isShadow, err := uc.ValidateInteraction(ctx, userID, "check this https://coolsite.com")
		assert.NoError(t, err)
		assert.True(t, isShadow)
	})

	t.Run("Link Allowed - Level 12", func(t *testing.T) {
		userID := uint64(6)
		mockUserRepo.GetByIDFunc = func(id uint64) (*domain.User, error) {
			return &domain.User{ID: id, Level: 12}, nil
		}
		mockUserRepo.GetLastInteractionTimeFunc = func(uID uint64) (time.Time, error) {
			return time.Time{}, nil
		}

		isShadow, err := uc.ValidateInteraction(ctx, userID, "check this www.google.com")
		assert.NoError(t, err)
		assert.False(t, isShadow)
	})
}

func TestModerationUsecase_ShadowbanTriggers(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockModerationRepository)
	mockUserRepo := new(testutil.MockUserRepository)
	mockSongRepo := new(testutil.MockSongRepository)
	mockCommentRepo := new(testutil.MockCommentRepository)
	mockNotif := new(MockNotificationUsecase)
	mockMedia := new(testutil.MockMediaService)

	uc := moderation.NewModerationUsecase(mockRepo, mockUserRepo, mockSongRepo, mockCommentRepo, mockNotif, mockMedia)

	t.Run("Trigger Shadowban on TruthScore < 50", func(t *testing.T) {
		reportID := uint64(10)
		reporterID := uint64(1)
		reportedID := uint64(2)

		report := &domain.UserReport{
			ID:             reportID,
			ReporterUserID: reporterID,
			ReportedUserID: reportedID,
		}

		mockRepo.On("GetUserReport", ctx, reportID).Return(report, nil)
		mockRepo.On("ResolveUserReport", ctx, reportID, true).Return(nil)
		mockRepo.On("UpdateUserTruthScore", ctx, reporterID, 5).Return(nil)
		mockRepo.On("UpdateUserTruthScore", ctx, reportedID, -10).Return(nil)
		mockRepo.On("GetPendingReportsCount", ctx, reportedID).Return(0, nil)

		// User starts with 55, minus 10 is 45 (should shadowban)
		mockUserRepo.GetByIDFunc = func(id uint64) (*domain.User, error) {
			return &domain.User{ID: id, TruthScore: 45, IsShadowbanned: false}, nil
		}

		mockRepo.On("ShadowbanUser", ctx, reportedID).Return(nil)

		err := uc.ResolveUserReport(ctx, reportID, true)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Trigger Softban on low TruthScore and high reports", func(t *testing.T) {
		reportID := uint64(11)
		reportedID := uint64(3)

		report := &domain.UserReport{ID: reportID, ReportedUserID: reportedID}

		mockRepo.On("GetUserReport", ctx, reportID).Return(report, nil)
		mockRepo.On("ResolveUserReport", ctx, reportID, true).Return(nil)
		mockRepo.On("UpdateUserTruthScore", ctx, mock.Anything, mock.Anything).Return(nil)

		// User with 25 truth score
		mockUserRepo.GetByIDFunc = func(id uint64) (*domain.User, error) {
			return &domain.User{ID: id, TruthScore: 25, IsSoftbanned: false, IsShadowbanned: true}, nil
		}
		// 5 pending reports (threshold is > 3)
		mockRepo.On("GetPendingReportsCount", ctx, reportedID).Return(5, nil)
		mockUserRepo.UpdateSoftbanStatusFunc = func(id uint64, status bool) error {
			assert.Equal(t, reportedID, id)
			assert.True(t, status)
			return nil
		}

		err := uc.ResolveUserReport(ctx, reportID, true)
		assert.NoError(t, err)
	})
}
