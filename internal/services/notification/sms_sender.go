package notification

import (
	"fmt"

	"pay-your-dues/internal/config"
	
	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

// SMSSender handles sending SMS notifications via Twilio
type SMSSender struct {
	config *config.NotificationConfig
	client *twilio.RestClient
}

// NewSMSSender creates a new SMS sender instance
func NewSMSSender(cfg *config.NotificationConfig) *SMSSender {
	var client *twilio.RestClient
	if cfg.TwilioAccountSID != "" && cfg.TwilioAuthToken != "" {
		client = twilio.NewRestClientWithParams(twilio.ClientParams{
			Username: cfg.TwilioAccountSID,
			Password: cfg.TwilioAuthToken,
		})
	}

	return &SMSSender{
		config: cfg,
		client: client,
	}
}

// SendSMS sends an SMS message to the specified phone number
func (ss *SMSSender) SendSMS(to, message string) error {
	if !ss.IsConfigured() {
		return fmt.Errorf("Twilio SMS is not configured")
	}

	if to == "" {
		return fmt.Errorf("recipient phone number is required")
	}

	if message == "" {
		return fmt.Errorf("SMS message is required")
	}

	// Ensure phone number is in E.164 format
	if to[0] != '+' {
		return fmt.Errorf("phone number must be in E.164 format (starting with +)")
	}

	params := &twilioApi.CreateMessageParams{}
	params.SetTo(to)
	params.SetFrom(ss.config.TwilioPhoneNumber)
	params.SetBody(message)

	resp, err := ss.client.Api.CreateMessage(params)
	if err != nil {
		return fmt.Errorf("failed to send SMS to %s: %w", to, err)
	}

	// Check if message was queued/sent successfully
	if resp.Status == nil {
		return fmt.Errorf("unexpected response from Twilio")
	}

	// Status can be "queued", "sending", "sent", "failed", "delivered", etc.
	if *resp.Status == "failed" {
		return fmt.Errorf("Twilio failed to send SMS to %s", to)
	}

	return nil
}

// SendBulkSMS sends multiple SMS messages
func (ss *SMSSender) SendBulkSMS(messages []SMSMessage) error {
	if !ss.IsConfigured() {
		return fmt.Errorf("Twilio SMS is not configured")
	}

	if len(messages) == 0 {
		return nil
	}

	// Send each SMS
	for _, msg := range messages {
		if err := ss.SendSMS(msg.To, msg.Message); err != nil {
			// Log error but continue with other messages
			fmt.Printf("Failed to send SMS to %s: %v\n", msg.To, err)
			continue
		}
	}

	return nil
}

// SMSMessage represents a single SMS message
type SMSMessage struct {
	To      string
	Message string
}

// IsConfigured checks if Twilio SMS is properly configured
func (ss *SMSSender) IsConfigured() bool {
	return ss.config.TwilioAccountSID != "" &&
		ss.config.TwilioAuthToken != "" &&
		ss.config.TwilioPhoneNumber != "" &&
		ss.client != nil
}

// ValidatePhoneNumber validates a phone number format (E.164)
func ValidatePhoneNumber(phone string) bool {
	if len(phone) == 0 {
		return false
	}
	
	// Must start with +
	if phone[0] != '+' {
		return false
	}
	
	// Must be between 8 and 15 digits (excluding the +)
	digits := phone[1:]
	if len(digits) < 8 || len(digits) > 15 {
		return false
	}
	
	// All characters after + must be digits
	for _, c := range digits {
		if c < '0' || c > '9' {
			return false
		}
	}
	
	return true
}

