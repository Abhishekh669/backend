package lib

import (
	"fmt"
	"net/smtp"

	"github.com/Abhishekh669/backend/internals/config"
)

// Use this method - Simple and reliable
func SendForgetPasswordTokenToEmail(email string, pin string) error {
	// Check if email exists

	// Send plain text email
	senderEmail := config.AppConfig.SMTPEmail
	senderPassword := config.AppConfig.SMTPPassword
	smtpHost := config.AppConfig.SMTPHost
	smtpPort := config.AppConfig.SMTPPort

	// Better formatted email content
	subject := "Password Reset Request"
	body := fmt.Sprintf(`
Dear User,

We received a request to reset the password for your account.

Your password reset verification PIN is:

    %s

This PIN is valid for 15 minutes.

If you didn't request this password reset, please ignore this email.

For security reasons, never share this PIN with anyone.

Best regards,
Support Team
`, pin)

	// Proper email message format with headers
	message := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/plain; charset=utf-8\r\n"+
		"\r\n"+
		"%s\r\n", senderEmail, email, subject, body)

	auth := smtp.PlainAuth("", senderEmail, senderPassword, smtpHost)
	smtpAddr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)

	return smtp.SendMail(smtpAddr, auth, senderEmail, []string{email}, []byte(message))
}
