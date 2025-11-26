package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SlackSender handles sending notifications to Slack via webhooks
type SlackSender struct {
	client *http.Client
}

// NewSlackSender creates a new Slack sender instance
func NewSlackSender() *SlackSender {
	return &SlackSender{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SlackMessage represents a Slack message with blocks
type SlackMessage struct {
	Text   string       `json:"text"`
	Blocks []SlackBlock `json:"blocks"`
}

// SlackBlock represents a Slack message block
type SlackBlock struct {
	Type   string             `json:"type"`
	Text   *SlackTextObject   `json:"text,omitempty"`
	Fields []SlackTextObject  `json:"fields,omitempty"`
}

// SlackTextObject represents a Slack text object
type SlackTextObject struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// SendNotification sends a notification to Slack
func (s *SlackSender) SendNotification(webhookURL string, data TemplateData) error {
	if webhookURL == "" {
		return fmt.Errorf("Slack webhook URL is required")
	}

	message := s.formatMessage(data)

	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack payload: %w", err)
	}

	resp, err := s.client.Post(webhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to send Slack webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Slack webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// formatMessage formats notification data into a Slack message with blocks
func (s *SlackSender) formatMessage(data TemplateData) SlackMessage {
	// Determine message type based on available data
	if data.PaymentStatus != "" {
		// This is a payment verification/confirmation notification
		if data.PaymentStatus == "completed" {
			return s.formatPaymentVerified(data)
		} else if data.PaymentStatus == "rejected" {
			return s.formatPaymentRejected(data)
		}
		return s.formatPaymentConfirmation(data)
	}
	
	// Default to payment reminder
	return s.formatPaymentReminder(data)
}

// formatPaymentReminder formats a payment reminder message
func (s *SlackSender) formatPaymentReminder(data TemplateData) SlackMessage {
	blocks := []SlackBlock{
		{
			Type: "header",
			Text: &SlackTextObject{
				Type: "plain_text",
				Text: "💰 Payment Reminder",
			},
		},
		{
			Type: "section",
			Text: &SlackTextObject{
				Type: "mrkdwn",
				Text: fmt.Sprintf("*Payment due to %s*", data.ContactName),
			},
		},
		{
			Type: "section",
			Fields: []SlackTextObject{
			{
				Type: "mrkdwn",
				Text: fmt.Sprintf("*Amount:*\n%s %s", data.Currency, data.Amount.StringFixed(2)),
			},
			{
				Type: "mrkdwn",
				Text: fmt.Sprintf("*Due Date:*\n%s", data.DueDate.Format("January 02, 2006")),
			},
			{
				Type: "mrkdwn",
				Text: fmt.Sprintf("*Days Until Due:*\n%d days", data.DaysUntilDue),
			},
			},
		},
	}

	// Add installment info if present
	if data.InstallmentNumber > 0 && data.InstallmentTotal > 0 {
		blocks = append(blocks, SlackBlock{
			Type: "section",
			Fields: []SlackTextObject{
				{
					Type: "mrkdwn",
					Text: fmt.Sprintf("*Installment:*\n#%d of %d", data.InstallmentNumber, data.InstallmentTotal),
				},
				{
					Type: "mrkdwn",
					Text: fmt.Sprintf("*Remaining Debt:*\n%s %s", data.Currency, data.RemainingDebt.StringFixed(2)),
				},
			},
		})
	}

	return SlackMessage{
		Text:   fmt.Sprintf("Payment Reminder: %s %s due to %s", data.Currency, data.Amount.StringFixed(2), data.ContactName),
		Blocks: blocks,
	}
}

// formatPaymentConfirmation formats a payment confirmation message
func (s *SlackSender) formatPaymentConfirmation(data TemplateData) SlackMessage {
	return SlackMessage{
		Text: "Payment Received",
		Blocks: []SlackBlock{
			{
				Type: "header",
				Text: &SlackTextObject{
					Type: "plain_text",
					Text: "✓ Payment Received",
				},
			},
			{
				Type: "section",
				Text: &SlackTextObject{
					Type: "mrkdwn",
					Text: fmt.Sprintf("*A payment has been recorded for %s*", data.ContactName),
				},
			},
			{
				Type: "section",
				Fields: []SlackTextObject{
					{
						Type: "mrkdwn",
						Text: fmt.Sprintf("*Amount:*\n%s %s", data.Currency, data.Amount),
					},
					{
						Type: "mrkdwn",
						Text: fmt.Sprintf("*Status:*\n%s", data.PaymentStatus),
					},
				},
			},
		},
	}
}

// formatPaymentVerified formats a payment verified message
func (s *SlackSender) formatPaymentVerified(data TemplateData) SlackMessage {
	return SlackMessage{
		Text: "Payment Verified",
		Blocks: []SlackBlock{
			{
				Type: "header",
				Text: &SlackTextObject{
					Type: "plain_text",
					Text: "✅ Payment Verified",
				},
			},
			{
				Type: "section",
				Text: &SlackTextObject{
					Type: "mrkdwn",
					Text: fmt.Sprintf("*Your payment to %s has been verified and accepted!*", data.ContactName),
				},
			},
			{
				Type: "section",
				Fields: []SlackTextObject{
					{
						Type: "mrkdwn",
						Text: fmt.Sprintf("*Amount:*\n%s %s", data.Currency, data.Amount),
					},
					{
						Type: "mrkdwn",
						Text: "*Status:*\nVerified & Completed",
					},
				},
			},
		},
	}
}

// formatPaymentRejected formats a payment rejected message
func (s *SlackSender) formatPaymentRejected(data TemplateData) SlackMessage {
	return SlackMessage{
		Text: "Payment Rejected",
		Blocks: []SlackBlock{
			{
				Type: "header",
				Text: &SlackTextObject{
					Type: "plain_text",
					Text: "❌ Payment Rejected",
				},
			},
			{
				Type: "section",
				Text: &SlackTextObject{
					Type: "mrkdwn",
					Text: fmt.Sprintf("*Your payment to %s has been rejected.*", data.ContactName),
				},
			},
			{
				Type: "section",
				Fields: []SlackTextObject{
					{
						Type: "mrkdwn",
						Text: fmt.Sprintf("*Amount:*\n%s %s", data.Currency, data.Amount),
					},
					{
						Type: "mrkdwn",
						Text: "*Status:*\nRejected",
					},
				},
			},
			{
				Type: "section",
				Text: &SlackTextObject{
					Type: "mrkdwn",
					Text: fmt.Sprintf("*Reason:*\n%s", data.RejectionReason),
				},
			},
		},
	}
}

// formatDefaultMessage formats a default message
func (s *SlackSender) formatDefaultMessage(data TemplateData) SlackMessage {
	message := data.CustomMessage
	if message == "" {
		message = fmt.Sprintf("Payment notification for %s: %s %s", data.ContactName, data.Currency, data.Amount.StringFixed(2))
	}
	
	return SlackMessage{
		Text: "Payment Notification",
		Blocks: []SlackBlock{
			{
				Type: "section",
				Text: &SlackTextObject{
					Type: "mrkdwn",
					Text: message,
				},
			},
		},
	}
}

