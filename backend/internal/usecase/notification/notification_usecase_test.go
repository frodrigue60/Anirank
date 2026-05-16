package notification_test

import (
	"anirank/api/internal/domain"
	"anirank/api/internal/testutil"
	"anirank/api/internal/usecase/notification"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNotificationUsecase_Create(t *testing.T) {
	ctx := context.Background()
	mockRepo := &testutil.MockNotificationRepository{}
	mockUserRepo := &testutil.MockUserRepository{}
	mockMail := &testutil.MockMailService{}
	mockCache := &testutil.MockCache{}

	user := &domain.User{
		ID:    1,
		Name:  "Test User",
		Email: "test@example.com",
	}

	notifData := map[string]interface{}{
		"follower_name": "Follower",
	}
	dataJSON, _ := json.Marshal(notifData)

	notif := &domain.Notification{
		UserID: 1,
		Type:   "follow",
		Data:   dataJSON,
	}

	t.Run("Create with emails disabled", func(t *testing.T) {
		emailSent := false
		mockMail.SendActivityNotificationFunc = func(to, userName, notificationType string, data map[string]interface{}) error {
			emailSent = true
			return nil
		}
		mockRepo.CreateFunc = func(n *domain.Notification) error {
			return nil
		}

		uc := notification.NewNotificationUsecase(mockRepo, mockUserRepo, mockMail, mockCache, false)
		err := uc.Create(ctx, notif)

		assert.NoError(t, err)
		// Wait a bit for the goroutine (though it shouldn't even start)
		time.Sleep(50 * time.Millisecond)
		assert.False(t, emailSent, "Email should not be sent when enableEmails is false")
	})

	t.Run("Create with emails enabled", func(t *testing.T) {
		emailSent := make(chan bool, 1)
		mockMail.SendActivityNotificationFunc = func(to, userName, notificationType string, data map[string]interface{}) error {
			assert.Equal(t, user.Email, to)
			assert.Equal(t, user.Name, userName)
			assert.Equal(t, "follow", notificationType)
			emailSent <- true
			return nil
		}
		mockUserRepo.GetByIDFunc = func(id uint64) (*domain.User, error) {
			return user, nil
		}
		mockRepo.CreateFunc = func(n *domain.Notification) error {
			return nil
		}

		uc := notification.NewNotificationUsecase(mockRepo, mockUserRepo, mockMail, mockCache, true)
		err := uc.Create(ctx, notif)

		assert.NoError(t, err)

		select {
		case sent := <-emailSent:
			assert.True(t, sent)
		case <-time.After(1 * time.Second):
			t.Fatal("Timeout waiting for email notification")
		}
	})
}
