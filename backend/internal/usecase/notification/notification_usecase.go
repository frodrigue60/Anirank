package notification

import (
	"context"
	"encoding/json"
	"fmt"

	"anirank/api/internal/domain"
)

type notificationUsecase struct {
	repo  domain.NotificationRepository
	cache domain.Cache
}

func NewNotificationUsecase(repo domain.NotificationRepository, cache domain.Cache) domain.NotificationUsecase {
	return &notificationUsecase{repo: repo, cache: cache}
}

func (u *notificationUsecase) Create(ctx context.Context, notification *domain.Notification) error {
	// Check user settings before creating
	settings, err := u.repo.GetSettings(ctx, notification.UserID)
	if err == nil && settings != nil {
		var prefs map[string]bool
		if err := json.Unmarshal(settings.Settings, &prefs); err == nil {
			// Check if the specific notification type is disabled
			// If not found in map, we assume it's enabled (default)
			if enabled, exists := prefs[notification.Type]; exists && !enabled {
				return nil // Silently skip
			}
		}
	}

	err = u.repo.Create(ctx, notification)
	if err != nil {
		return err
	}

	// Notificamos vía Redis Pub/Sub al canal del usuario
	if u.cache.IsAvailable() {
		channel := fmt.Sprintf("notifications:%d", notification.UserID)
		_ = u.cache.Publish(ctx, channel, notification)
	}
	return nil
}

func (u *notificationUsecase) GetNotifications(ctx context.Context, userID uint64, notificationType string, limit, offset int) ([]domain.Notification, int, int, error) {
	notifications, total, err := u.repo.GetByUserID(ctx, userID, notificationType, limit, offset)
	if err != nil {
		return nil, 0, 0, err
	}

	unreadCount, err := u.repo.GetUnreadCount(ctx, userID)
	if err != nil {
		return nil, 0, 0, err
	}

	return notifications, total, unreadCount, nil
}

func (u *notificationUsecase) MarkAsRead(ctx context.Context, id string, userID uint64) error {
	return u.repo.MarkAsRead(ctx, id, userID)
}

func (u *notificationUsecase) MarkAllAsRead(ctx context.Context, userID uint64) error {
	return u.repo.MarkAllAsRead(ctx, userID)
}

func (u *notificationUsecase) Delete(ctx context.Context, id string, userID uint64) error {
	return u.repo.Delete(ctx, id, userID)
}

func (u *notificationUsecase) GetUnreadCount(ctx context.Context, userID uint64) (int, error) {
	return u.repo.GetUnreadCount(ctx, userID)
}

func (u *notificationUsecase) SubscribeToStream(ctx context.Context, userID uint64) (domain.Subscriber, error) {
	channel := fmt.Sprintf("notifications:%d", userID)
	return u.cache.Subscribe(ctx, channel)
}

func (u *notificationUsecase) GetSettings(ctx context.Context, userID uint64) (*domain.NotificationSettings, error) {
	return u.repo.GetSettings(ctx, userID)
}

func (u *notificationUsecase) UpdateSettings(ctx context.Context, userID uint64, settings json.RawMessage) error {
	return u.repo.UpdateSettings(ctx, userID, settings)
}
