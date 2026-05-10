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
	
	subject := "Verifica tu cuenta en AniRank"
	html := fmt.Sprintf(`
		<div style="font-family: sans-serif; max-width: 600px; margin: auto; padding: 20px; border: 1px solid #eee; border-radius: 10px;">
			<h2 style="color: #7f13ec;">¡Hola, %s!</h2>
			<p>Gracias por unirte a AniRank. Para completar tu registro, por favor verifica tu dirección de correo electrónico haciendo clic en el siguiente botón:</p>
			<div style="text-align: center; margin: 30px 0;">
				<a href="%s" style="background-color: #7f13ec; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; font-weight: bold;">Verificar mi cuenta</a>
			</div>
			<p style="color: #666; font-size: 0.9em;">Si no puedes hacer clic en el botón, copia y pega este enlace en tu navegador:</p>
			<p style="color: #666; font-size: 0.8em; word-break: break-all;">%s</p>
			<hr style="border: 0; border-top: 1px solid #eee; margin: 20px 0;">
			<p style="font-size: 0.8em; color: #999; text-align: center;">Este enlace expirará en 24 horas.</p>
		</div>
	`, name, verifyURL, verifyURL)

	text := fmt.Sprintf("¡Hola, %s!\n\nGracias por unirte a AniRank. Para completar tu registro, por favor verifica tu cuenta en el siguiente enlace:\n\n%s\n\nEste enlace expirará en 24 horas.\n\n© AniRank", name, verifyURL)

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
	
	subject := "Restablecer tu contraseña en AniRank"
	html := fmt.Sprintf(`
		<div style="font-family: sans-serif; max-width: 600px; margin: auto; padding: 20px; border: 1px solid #eee; border-radius: 10px;">
			<h2 style="color: #7f13ec;">Hola, %s</h2>
			<p>Recibimos una solicitud para restablecer la contraseña de tu cuenta en AniRank. Si no realizaste esta solicitud, puedes ignorar este correo.</p>
			<p>Para elegir una nueva contraseña, haz clic en el siguiente botón:</p>
			<div style="text-align: center; margin: 30px 0;">
				<a href="%s" style="background-color: #7f13ec; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; font-weight: bold;">Restablecer contraseña</a>
			</div>
			<p style="color: #666; font-size: 0.9em;">Si no puedes hacer clic en el botón, copia y pega este enlace en tu navegador:</p>
			<p style="color: #666; font-size: 0.8em; word-break: break-all;">%s</p>
			<hr style="border: 0; border-top: 1px solid #eee; margin: 20px 0;">
			<p style="font-size: 0.8em; color: #999; text-align: center;">Este enlace expirará en 1 hora por seguridad.</p>
		</div>
	`, name, resetURL, resetURL)

	text := fmt.Sprintf("Hola, %s\n\nRecibimos una solicitud para restablecer tu contraseña en AniRank. Si no realizaste esta solicitud, puedes ignorar este correo.\n\nPara elegir una nueva contraseña, visita el siguiente enlace:\n\n%s\n\nEste enlace expirará en 1 hora por seguridad.\n\n© AniRank", name, resetURL)

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
