package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// DiscordSender handles sending notifications to Discord via webhooks
type DiscordSender struct {
	client *http.Client
}

// NewDiscordSender creates a new Discord sender instance
func NewDiscordSender() *DiscordSender {
	return &DiscordSender{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// DiscordMessage represents a Discord webhook message
type DiscordMessage struct {
	Content string         `json:"content,omitempty"`
	Embeds  []DiscordEmbed `json:"embeds,omitempty"`
}

// DiscordEmbed represents a Discord embed
type DiscordEmbed struct {
	Title       string              `json:"title,omitempty"`
	Description string              `json:"description,omitempty"`
	Color       int                 `json:"color,omitempty"`
	Fields      []DiscordEmbedField `json:"fields,omitempty"`
	Footer      *DiscordEmbedFooter `json:"footer,omitempty"`
	Timestamp   string              `json:"timestamp,omitempty"`
}

// DiscordEmbedField represents a field in a Discord embed
type DiscordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// DiscordEmbedFooter represents a footer in a Discord embed
type DiscordEmbedFooter struct {
	Text string `json:"text"`
}

// SendNotification sends a notification to Discord
func (d *DiscordSender) SendNotification(webhookURL string, data TemplateData) error {
	if webhookURL == "" {
		return fmt.Errorf("Discord webhook URL is required")
	}

	message := d.formatMessage(data)

	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal Discord payload: %w", err)
	}

	resp, err := d.client.Post(webhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to send Discord webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Discord webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// formatMessage formats notification data into a Discord message with embeds
func (d *DiscordSender) formatMessage(data TemplateData) DiscordMessage {
	// Determine message type based on available data
	if data.PaymentStatus != "" {
		// This is a payment verification/confirmation notification
		if data.PaymentStatus == "completed" {
			return d.formatPaymentVerified(data)
		} else if data.PaymentStatus == "rejected" {
			return d.formatPaymentRejected(data)
		}
		return d.formatPaymentConfirmation(data)
	}
	
	// Default to payment reminder
	return d.formatPaymentReminder(data)
}

// formatPaymentReminder formats a payment reminder message
func (d *DiscordSender) formatPaymentReminder(data TemplateData) DiscordMessage {
	fields := []DiscordEmbedField{
		{
			Name:   "Amount",
			Value:  fmt.Sprintf("%s %s", data.Currency, data.Amount),
			Inline: true,
		},
		{
			Name:   "Due Date",
			Value:  data.DueDate.Format("January 02, 2006"),
			Inline: true,
		},
		{
			Name:   "Days Until Due",
			Value:  fmt.Sprintf("%s days", data.DaysUntilDue),
			Inline: true,
		},
	}

	// Add installment info if present
	if data.InstallmentNumber > 0 && data.InstallmentTotal > 0 {
		fields = append(fields, DiscordEmbedField{
			Name:   "Installment",
			Value:  fmt.Sprintf("#%s of %s", data.InstallmentNumber, data.InstallmentTotal),
			Inline: true,
		})
		fields = append(fields, DiscordEmbedField{
			Name:   "Remaining Debt",
			Value:  fmt.Sprintf("%s %s", data.Currency, data.RemainingDebt),
			Inline: true,
		})
	}

	embed := DiscordEmbed{
		Title:       "💰 Payment Reminder",
		Description: fmt.Sprintf("Payment due to **%s**", data.ContactName),
		Color:       3066993, // Green color
		Fields:      fields,
		Footer: &DiscordEmbedFooter{
			Text: "Pay Your Dues",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	return DiscordMessage{
		Embeds: []DiscordEmbed{embed},
	}
}

// formatPaymentConfirmation formats a payment confirmation message
func (d *DiscordSender) formatPaymentConfirmation(data TemplateData) DiscordMessage {
	embed := DiscordEmbed{
		Title:       "✓ Payment Received",
		Description: fmt.Sprintf("A payment has been recorded for **%s**", data.ContactName),
		Color:       3447003, // Blue color
		Fields: []DiscordEmbedField{
			{
				Name:   "Amount",
				Value:  fmt.Sprintf("%s %s", data.Currency, data.Amount),
				Inline: true,
			},
			{
				Name:   "Status",
				Value:  data.PaymentStatus,
				Inline: true,
			},
		},
		Footer: &DiscordEmbedFooter{
			Text: "Pay Your Dues",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	return DiscordMessage{
		Embeds: []DiscordEmbed{embed},
	}
}

// formatPaymentVerified formats a payment verified message
func (d *DiscordSender) formatPaymentVerified(data TemplateData) DiscordMessage {
	embed := DiscordEmbed{
		Title:       "✅ Payment Verified",
		Description: fmt.Sprintf("Your payment to **%s** has been verified and accepted!", data.ContactName),
		Color:       3066993, // Green color
		Fields: []DiscordEmbedField{
			{
				Name:   "Amount",
				Value:  fmt.Sprintf("%s %s", data.Currency, data.Amount),
				Inline: true,
			},
			{
				Name:   "Status",
				Value:  "Verified & Completed",
				Inline: true,
			},
		},
		Footer: &DiscordEmbedFooter{
			Text: "Pay Your Dues",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	return DiscordMessage{
		Embeds: []DiscordEmbed{embed},
	}
}

// formatPaymentRejected formats a payment rejected message
func (d *DiscordSender) formatPaymentRejected(data TemplateData) DiscordMessage {
	embed := DiscordEmbed{
		Title:       "❌ Payment Rejected",
		Description: fmt.Sprintf("Your payment to **%s** has been rejected.", data.ContactName),
		Color:       15158332, // Red color
		Fields: []DiscordEmbedField{
			{
				Name:   "Amount",
				Value:  fmt.Sprintf("%s %s", data.Currency, data.Amount),
				Inline: true,
			},
			{
				Name:   "Status",
				Value:  "Rejected",
				Inline: true,
			},
		{
			Name:   "Reason",
			Value:  data.RejectionReason,
			Inline: false,
		},
		},
		Footer: &DiscordEmbedFooter{
			Text: "Pay Your Dues",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	return DiscordMessage{
		Embeds: []DiscordEmbed{embed},
	}
}

// formatDefaultMessage formats a default message
func (d *DiscordSender) formatDefaultMessage(data TemplateData) DiscordMessage {
	message := data.CustomMessage
	if message == "" {
		message = fmt.Sprintf("Payment notification for **%s**: %s %s", data.ContactName, data.Currency, data.Amount.StringFixed(2))
	}
	
	embed := DiscordEmbed{
		Title:       "Payment Notification",
		Description: message,
		Color:       3447003, // Blue color
		Footer: &DiscordEmbedFooter{
			Text: "Pay Your Dues",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	return DiscordMessage{
		Embeds: []DiscordEmbed{embed},
	}
}

