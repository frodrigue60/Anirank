package domain

import "context"

type MailService interface {
	SendVerificationEmail(ctx context.Context, to string, name string, token string) error
	SendPasswordResetEmail(ctx context.Context, to string, name string, token string) error
	SendActivityNotificationEmail(ctx context.Context, to string, userName string, notificationType string, data map[string]interface{}) error
}
