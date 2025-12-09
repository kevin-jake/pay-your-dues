package notification

import (
	"fmt"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TelegramSender handles sending notifications to Telegram via Bot API
type TelegramSender struct{}

// NewTelegramSender creates a new Telegram sender instance
func NewTelegramSender() *TelegramSender {
	return &TelegramSender{}
}

// SendNotification sends a notification to Telegram
func (t *TelegramSender) SendNotification(botToken, chatID string, data TemplateData) error {
	if botToken == "" {
		return fmt.Errorf("Telegram bot token is required")
	}

	if chatID == "" {
		return fmt.Errorf("Telegram chat ID is required")
	}

	// Create bot instance
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return fmt.Errorf("failed to create Telegram bot: %w", err)
	}

	// Parse chat ID to int64
	chatIDInt, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid Telegram chat ID: %w", err)
	}

	// Format message based on type
	message := t.formatMessage(data)

	// Create message config
	msg := tgbotapi.NewMessage(chatIDInt, message)
	msg.ParseMode = "Markdown"

	// Send message
	_, err = bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send Telegram message: %w", err)
	}

	return nil
}

// formatMessage formats notification data into a Telegram message
func (t *TelegramSender) formatMessage(data TemplateData) string {
	// Determine message type based on available data
	if data.PaymentStatus != "" {
		// This is a payment verification/confirmation notification
		if data.PaymentStatus == "completed" {
			return t.formatPaymentVerified(data)
		} else if data.PaymentStatus == "rejected" {
			return t.formatPaymentRejected(data)
		}
		return t.formatPaymentConfirmation(data)
	}
	
	// Default to payment reminder
	return t.formatPaymentReminder(data)
}

// formatPaymentReminder formats a payment reminder message
func (t *TelegramSender) formatPaymentReminder(data TemplateData) string {
	message := "💰 *Payment Reminder*\n\n"
	message += fmt.Sprintf("Payment due to *%s*\n\n", data.ContactName)
	message += fmt.Sprintf("*Amount:* %s %s\n", data.Currency, data.Amount)
	message += fmt.Sprintf("*Due Date:* %s\n", data.DueDate)
	message += fmt.Sprintf("*Days Until Due:* %d days\n", data.DaysUntilDue)

	if data.InstallmentNumber > 0 && data.InstallmentTotal > 0 {
		message += fmt.Sprintf("\n*Installment:* #%d of %d\n", data.InstallmentNumber, data.InstallmentTotal)
		message += fmt.Sprintf("*Remaining Debt:* %s %s\n", data.Currency, data.RemainingDebt)
	}

	return message
}

// formatPaymentConfirmation formats a payment confirmation message
func (t *TelegramSender) formatPaymentConfirmation(data TemplateData) string {
	message := "✓ *Payment Received*\n\n"
	message += fmt.Sprintf("A payment has been recorded for *%s*\n\n", data.ContactName)
	message += fmt.Sprintf("*Amount:* %s %s\n", data.Currency, data.Amount)
	message += fmt.Sprintf("*Status:* %s\n", data.PaymentStatus)

	return message
}

// formatPaymentVerified formats a payment verified message
func (t *TelegramSender) formatPaymentVerified(data TemplateData) string {
	message := "✅ *Payment Verified*\n\n"
	message += fmt.Sprintf("Your payment to *%s* has been verified and accepted!\n\n", data.ContactName)
	message += fmt.Sprintf("*Amount:* %s %s\n", data.Currency, data.Amount)
	message += "*Status:* Verified & Completed\n"

	return message
}

// formatPaymentRejected formats a payment rejected message
func (t *TelegramSender) formatPaymentRejected(data TemplateData) string {
	message := "❌ *Payment Rejected*\n\n"
	message += fmt.Sprintf("Your payment to *%s* has been rejected.\n\n", data.ContactName)
	message += fmt.Sprintf("*Amount:* %s %s\n", data.Currency, data.Amount)
	message += "*Status:* Rejected\n\n"
	message += fmt.Sprintf("*Reason:* %s\n", data.RejectionReason)

	return message
}

// formatDefaultMessage formats a default message
func (t *TelegramSender) formatDefaultMessage(data TemplateData) string {
	message := data.CustomMessage
	if message == "" {
		message = fmt.Sprintf("Payment notification for *%s*: %s %s", data.ContactName, data.Currency, data.Amount.StringFixed(2))
	}
	return message
}

