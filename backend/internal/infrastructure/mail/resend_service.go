package mail

import (
	"context"
	"fmt"
	"anirank/api/internal/domain"
	"github.com/resend/resend-go/v2"
)

type resendService struct {
	client    *resend.Client
	fromEmail string
	appURL    string
}

func NewResendService(apiKey string, fromEmail string, appURL string) domain.MailService {
	client := resend.NewClient(apiKey)
	return &resendService{
		client:    client,
		fromEmail: fromEmail,
		appURL:    appURL,
	}
}

func (s *resendService) SendVerificationEmail(ctx context.Context, to string, name string, token string) error {
	verifyURL := fmt.Sprintf("%s/verify-email?token=%s", s.appURL, token)
	
	subject := "Verify your AniRank account"
	html := fmt.Sprintf(`
		<div style="font-family: sans-serif; max-width: 600px; margin: auto; padding: 20px; border: 1px solid #eee; border-radius: 10px;">
			<h2 style="color: #7f13ec;">Hello, %s!</h2>
			<p>Thanks for joining AniRank. To complete your registration, please verify your email address by clicking the button below:</p>
			<div style="text-align: center; margin: 30px 0;">
				<a href="%s" style="background-color: #7f13ec; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; font-weight: bold;">Verify my account</a>
			</div>
			<p style="color: #666; font-size: 0.9em;">If you cannot click the button, copy and paste this link into your browser:</p>
			<p style="color: #666; font-size: 0.8em; word-break: break-all;">%s</p>
			<hr style="border: 0; border-top: 1px solid #eee; margin: 20px 0;">
			<p style="font-size: 0.8em; color: #999; text-align: center;">This link will expire in 24 hours.</p>
		</div>
	`, name, verifyURL, verifyURL)

	text := fmt.Sprintf("Hello, %s!\n\nThanks for joining AniRank. To complete your registration, please verify your account at the following link:\n\n%s\n\nThis link will expire in 24 hours.\n\n© AniRank", name, verifyURL)

	params := &resend.SendEmailRequest{
		From:    s.fromEmail,
		To:      []string{to},
		Subject: subject,
		Html:    html,
		Text:    text,
	}

	_, err := s.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return fmt.Errorf("resend: failed to send verification email: %w", err)
	}

	return nil
}

func (s *resendService) SendPasswordResetEmail(ctx context.Context, to string, name string, token string) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.appURL, token)
	
	subject := "Reset your AniRank password"
	html := fmt.Sprintf(`
		<div style="font-family: sans-serif; max-width: 600px; margin: auto; padding: 20px; border: 1px solid #eee; border-radius: 10px;">
			<h2 style="color: #7f13ec;">Hello, %s</h2>
			<p>We received a request to reset the password for your AniRank account. If you did not make this request, you can safely ignore this email.</p>
			<p>To choose a new password, click the button below:</p>
			<div style="text-align: center; margin: 30px 0;">
				<a href="%s" style="background-color: #7f13ec; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; font-weight: bold;">Reset password</a>
			</div>
			<p style="color: #666; font-size: 0.9em;">If you cannot click the button, copy and paste this link into your browser:</p>
			<p style="color: #666; font-size: 0.8em; word-break: break-all;">%s</p>
			<hr style="border: 0; border-top: 1px solid #eee; margin: 20px 0;">
			<p style="font-size: 0.8em; color: #999; text-align: center;">This link will expire in 1 hour for security reasons.</p>
		</div>
	`, name, resetURL, resetURL)

	text := fmt.Sprintf("Hello, %s\n\nWe received a request to reset your password on AniRank. If you did not make this request, you can safely ignore this email.\n\nTo choose a new password, visit the following link:\n\n%s\n\nThis link will expire in 1 hour for security reasons.\n\n© AniRank", name, resetURL)

	params := &resend.SendEmailRequest{
		From:    s.fromEmail,
		To:      []string{to},
		Subject: subject,
		Html:    html,
		Text:    text,
	}

	_, err := s.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return fmt.Errorf("resend: failed to send reset password email: %w", err)
	}

	return nil
}
