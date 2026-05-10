package domain

import "context"

type MailService interface {
	SendVerificationEmail(ctx context.Context, to string, name string, token string) error
	SendPasswordResetEmail(ctx context.Context, to string, name string, token string) error
}
