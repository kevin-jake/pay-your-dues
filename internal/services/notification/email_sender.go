package notification

import (
	"fmt"

	"pay-your-dues/internal/config"

	"gopkg.in/gomail.v2"
)

// EmailSender handles sending email notifications via SMTP
type EmailSender struct {
	config *config.NotificationConfig
	dialer *gomail.Dialer
}

// NewEmailSender creates a new email sender instance
func NewEmailSender(cfg *config.NotificationConfig) *EmailSender {
	dialer := gomail.NewDialer(
		cfg.SMTPHost,
		cfg.SMTPPort,
		cfg.SMTPUsername,
		cfg.SMTPPassword,
	)

	return &EmailSender{
		config: cfg,
		dialer: dialer,
	}
}

// SendEmail sends an email with the given parameters
func (es *EmailSender) SendEmail(to, subject, body string) error {
	if to == "" {
		return fmt.Errorf("recipient email address is required")
	}

	if subject == "" {
		return fmt.Errorf("email subject is required")
	}

	if body == "" {
		return fmt.Errorf("email body is required")
	}

	m := gomail.NewMessage()
	
	// Set sender
	m.SetHeader("From", m.FormatAddress(es.config.SMTPFromEmail, es.config.SMTPFromName))
	
	// Set recipient
	m.SetHeader("To", to)
	
	// Set subject
	m.SetHeader("Subject", subject)
	
	// Set body as HTML
	m.SetBody("text/html", body)

	// Send the email
	if err := es.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email to %s: %w", to, err)
	}

	return nil
}

// SendEmailWithPlainText sends an email with both HTML and plain text versions
func (es *EmailSender) SendEmailWithPlainText(to, subject, htmlBody, plainTextBody string) error {
	if to == "" {
		return fmt.Errorf("recipient email address is required")
	}

	if subject == "" {
		return fmt.Errorf("email subject is required")
	}

	if htmlBody == "" && plainTextBody == "" {
		return fmt.Errorf("email body is required")
	}

	m := gomail.NewMessage()
	
	// Set sender
	m.SetHeader("From", m.FormatAddress(es.config.SMTPFromEmail, es.config.SMTPFromName))
	
	// Set recipient
	m.SetHeader("To", to)
	
	// Set subject
	m.SetHeader("Subject", subject)
	
	// Set body with alternative
	if plainTextBody != "" {
		m.SetBody("text/plain", plainTextBody)
	}
	if htmlBody != "" {
		m.AddAlternative("text/html", htmlBody)
	}

	// Send the email
	if err := es.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email to %s: %w", to, err)
	}

	return nil
}

// SendBulkEmails sends multiple emails efficiently using a single SMTP connection
func (es *EmailSender) SendBulkEmails(emails []EmailMessage) error {
	if len(emails) == 0 {
		return nil
	}

	// Open a persistent connection
	sender, err := es.dialer.Dial()
	if err != nil {
		return fmt.Errorf("failed to open SMTP connection: %w", err)
	}
	defer sender.Close()

	// Send each email using the same connection
	for _, email := range emails {
		m := gomail.NewMessage()
		m.SetHeader("From", m.FormatAddress(es.config.SMTPFromEmail, es.config.SMTPFromName))
		m.SetHeader("To", email.To)
		m.SetHeader("Subject", email.Subject)
		m.SetBody("text/html", email.Body)

		if err := gomail.Send(sender, m); err != nil {
			// Log error but continue with other emails
			fmt.Printf("Failed to send email to %s: %v\n", email.To, err)
			continue
		}
	}

	return nil
}

// EmailMessage represents a single email message
type EmailMessage struct {
	To      string
	Subject string
	Body    string
}

// IsConfigured checks if SMTP is properly configured
func (es *EmailSender) IsConfigured() bool {
	return es.config.SMTPHost != "" &&
		es.config.SMTPUsername != "" &&
		es.config.SMTPPassword != "" &&
		es.config.SMTPFromEmail != ""
}

