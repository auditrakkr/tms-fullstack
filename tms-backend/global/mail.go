package global

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"
)

// ── SMTP runtime config ───────────────────────────────────────────────────────

var (
	smtpHost     string
	smtpPort     string
	smtpUsername string
	smtpPassword string
)

// ── Mail option settings ──────────────────────────────────────────────────────

// MailOptions defines the parameters for sending an email.
type MailOptions struct {
	To          string
	From        string
	Subject     string
	Text        string
	Html        string
	ReplyTo     string
	Cc          []string
	Bcc         []string
	Attachments []Attachment
}

// Attachment represents an email attachment.
type Attachment struct {
	Filename    string
	Content     []byte
	ContentType string
}

// MailOptionSettings holds templates and defaults for a notification email type.
type MailOptionSettings struct {
	From         string
	Subject      string
	TextTemplate string
	HtmlTemplate string
}

// ConfirmEmailMailOptionSettings and ResetPasswordMailOptionSettings are the
// defaults used by UserService. They are overridden by InitMailTemplates.
var (
	ConfirmEmailMailOptionSettings = MailOptionSettings{
		From:         "noreply@auditrakkr.com",
		Subject:      "Email Verification",
		TextTemplate: "Please verify your email by clicking this link: {url}",
		HtmlTemplate: "<p>Please verify your email by clicking this link: <a href=\"{url}\">Verify Email</a></p>",
	}

	ResetPasswordMailOptionSettings = MailOptionSettings{
		From:         "noreply@auditrakkr.com",
		Subject:      "Password Reset Request",
		TextTemplate: "Reset your password by clicking this link: {url}",
		HtmlTemplate: "<p>Reset your password by clicking this link: <a href=\"{url}\">Reset Password</a></p>",
	}
)

// InitMailConfig sets SMTP credentials. Call once from main() after LoadConfig.
// port should be a string (e.g. "587").
func InitMailConfig(host, port, user, password string) {
	smtpHost = host
	smtpPort = port
	smtpUsername = user
	smtpPassword = password

	if smtpUsername == "" || smtpPassword == "" {
		slog.Warn("SMTP credentials not configured — email will be disabled")
	}
}

// InitMailTemplates overrides the default from-address and email templates.
// Call once from main() after LoadConfig, after InitMailConfig.
func InitMailTemplates(
	fromAddress,
	confirmSubject, confirmTemplate,
	resetSubject, resetTemplate string,
) {
	if fromAddress != "" {
		ConfirmEmailMailOptionSettings.From = fromAddress
		ResetPasswordMailOptionSettings.From = fromAddress
	}
	if confirmSubject != "" {
		ConfirmEmailMailOptionSettings.Subject = confirmSubject
	}
	if confirmTemplate != "" {
		ConfirmEmailMailOptionSettings.TextTemplate = confirmTemplate
	}
	if resetSubject != "" {
		ResetPasswordMailOptionSettings.Subject = resetSubject
	}
	if resetTemplate != "" {
		ResetPasswordMailOptionSettings.TextTemplate = resetTemplate
	}
}

// ── Sending ───────────────────────────────────────────────────────────────────

// SendMail sends a synchronous email using the configured SMTP server.
func SendMail(options MailOptions) error {
	if smtpUsername == "" || smtpPassword == "" {
		return fmt.Errorf("SMTP not configured: set SMTP_USER and SMTP_PASSWORD")
	}

	auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost)

	headers := map[string]string{
		"From":         options.From,
		"To":           options.To,
		"Subject":      options.Subject,
		"MIME-Version": "1.0",
	}
	if len(options.Cc) > 0 {
		headers["Cc"] = strings.Join(options.Cc, ",")
	}
	if options.ReplyTo != "" {
		headers["Reply-To"] = options.ReplyTo
	}

	var contentType, body string
	if options.Html != "" {
		contentType = "text/html; charset=UTF-8"
		body = options.Html
	} else {
		contentType = "text/plain; charset=UTF-8"
		body = options.Text
	}
	headers["Content-Type"] = contentType

	var message string
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	var recipients []string
	recipients = append(recipients, strings.Split(options.To, ",")...)
	recipients = append(recipients, options.Cc...)
	recipients = append(recipients, options.Bcc...)

	client, err := smtp.Dial(fmt.Sprintf("%s:%s", smtpHost, smtpPort))
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP: %w", err)
	}
	defer client.Close()

	if err := client.StartTLS(&tls.Config{ServerName: smtpHost}); err != nil {
		return fmt.Errorf("SMTP StartTLS failed: %w", err)
	}
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP auth failed: %w", err)
	}
	if err := client.Mail(options.From); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}
	for _, r := range recipients {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}
		if err := client.Rcpt(r); err != nil {
			return fmt.Errorf("failed to add recipient %s: %w", r, err)
		}
	}
	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to open data session: %w", err)
	}
	if _, err := wc.Write([]byte(message)); err != nil {
		return fmt.Errorf("failed to write email body: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}
	return client.Quit()
}

// SendMailAsync fires an email in a background goroutine.
// Errors are logged; the caller is never notified of failure.
func SendMailAsync(options MailOptions) {
	go func() {
		if err := SendMail(options); err != nil {
			slog.Error("failed to send email", "to", options.To, "error", err)
		}
	}()
}
