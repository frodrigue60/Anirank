package notification

import (
	"context"

	"anirank/api/internal/domain"
)

type notificationUsecase struct {
	repo domain.NotificationRepository
}

func NewNotificationUsecase(repo domain.NotificationRepository) domain.NotificationUsecase {
	return &notificationUsecase{repo: repo}
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
