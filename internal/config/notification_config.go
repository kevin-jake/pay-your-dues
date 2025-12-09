package config

import (
	"strconv"
	"time"
)

type NotificationConfig struct {
	// SMTP Configuration
	SMTPHost      string
	SMTPPort      int
	SMTPUsername  string
	SMTPPassword  string
	SMTPFromEmail string
	SMTPFromName  string

	// Twilio Configuration
	TwilioAccountSID  string
	TwilioAuthToken   string
	TwilioPhoneNumber string

	// Telegram Bot Configuration (app-level, shared by all users)
	TelegramBotToken      string
	TelegramWebhookSecret string
	TelegramPollingMode   bool // Use long polling instead of webhook (for development)

	// Worker Configuration
	WorkerInterval time.Duration
	MaxRetry       int
	BatchSize      int
}

func LoadNotificationConfig() *NotificationConfig {
	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	maxRetry, _ := strconv.Atoi(getEnv("NOTIFICATION_MAX_RETRY", "3"))
	batchSize, _ := strconv.Atoi(getEnv("NOTIFICATION_BATCH_SIZE", "100"))

	workerInterval, _ := time.ParseDuration(getEnv("NOTIFICATION_WORKER_INTERVAL", "1m"))

	// Parse telegram polling mode (default to true for development convenience)
	telegramPollingMode := getEnv("TELEGRAM_POLLING_MODE", "true") == "true"

	return &NotificationConfig{
		SMTPHost:              getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:              smtpPort,
		SMTPUsername:          getEnv("SMTP_USERNAME", ""),
		SMTPPassword:          getEnv("SMTP_PASSWORD", ""),
		SMTPFromEmail:         getEnv("SMTP_FROM_EMAIL", "noreply@payyourdues.com"),
		SMTPFromName:          getEnv("SMTP_FROM_NAME", "Pay Your Dues"),
		TwilioAccountSID:      getEnv("TWILIO_ACCOUNT_SID", ""),
		TwilioAuthToken:       getEnv("TWILIO_AUTH_TOKEN", ""),
		TwilioPhoneNumber:     getEnv("TWILIO_PHONE_NUMBER", ""),
		TelegramBotToken:      getEnv("TELEGRAM_BOT_TOKEN", ""),
		TelegramWebhookSecret: getEnv("TELEGRAM_WEBHOOK_SECRET", ""),
		TelegramPollingMode:   telegramPollingMode,
		WorkerInterval:        workerInterval,
		MaxRetry:              maxRetry,
		BatchSize:             batchSize,
	}
}

