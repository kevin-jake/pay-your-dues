package notification

import (
	"fmt"

	"pay-your-dues/internal/models"
)

// WebhookService provides a unified interface for sending notifications via webhooks
type WebhookService struct {
	slackSender    *SlackSender
	telegramSender *TelegramSender
	discordSender  *DiscordSender
}

// NewWebhookService creates a new webhook service instance
func NewWebhookService() *WebhookService {
	return &WebhookService{
		slackSender:    NewSlackSender(),
		telegramSender: NewTelegramSender(),
		discordSender:  NewDiscordSender(),
	}
}

// SendNotification sends a notification via the specified webhook platform
func (ws *WebhookService) SendNotification(
	webhookType string,
	settings *models.UserSettings,
	data TemplateData,
) error {
	switch webhookType {
	case "slack":
		return ws.sendSlackNotification(settings, data)
	case "telegram":
		return ws.sendTelegramNotification(settings, data)
	case "discord":
		return ws.sendDiscordNotification(settings, data)
	default:
		return fmt.Errorf("unsupported webhook type: %s", webhookType)
	}
}

// sendSlackNotification sends a notification to Slack
func (ws *WebhookService) sendSlackNotification(settings *models.UserSettings, data TemplateData) error {
	if settings.SlackWebhookURL == nil || *settings.SlackWebhookURL == "" {
		return fmt.Errorf("Slack webhook URL not configured")
	}

	return ws.slackSender.SendNotification(*settings.SlackWebhookURL, data)
}

// sendTelegramNotification sends a notification to Telegram
func (ws *WebhookService) sendTelegramNotification(settings *models.UserSettings, data TemplateData) error {
	if settings.TelegramBotToken == nil || *settings.TelegramBotToken == "" {
		return fmt.Errorf("Telegram bot token not configured")
	}

	if settings.TelegramChatID == nil || *settings.TelegramChatID == "" {
		return fmt.Errorf("Telegram chat ID not configured")
	}

	return ws.telegramSender.SendNotification(
		*settings.TelegramBotToken,
		*settings.TelegramChatID,
		data,
	)
}

// sendDiscordNotification sends a notification to Discord
func (ws *WebhookService) sendDiscordNotification(settings *models.UserSettings, data TemplateData) error {
	if settings.DiscordWebhookURL == nil || *settings.DiscordWebhookURL == "" {
		return fmt.Errorf("Discord webhook URL not configured")
	}

	return ws.discordSender.SendNotification(*settings.DiscordWebhookURL, data)
}

// IsWebhookConfigured checks if a specific webhook type is configured for the user
func (ws *WebhookService) IsWebhookConfigured(webhookType string, settings *models.UserSettings) bool {
	switch webhookType {
	case "slack":
		return settings.SlackWebhookURL != nil && *settings.SlackWebhookURL != ""
	case "telegram":
		return settings.TelegramBotToken != nil && *settings.TelegramBotToken != "" &&
			settings.TelegramChatID != nil && *settings.TelegramChatID != ""
	case "discord":
		return settings.DiscordWebhookURL != nil && *settings.DiscordWebhookURL != ""
	default:
		return false
	}
}

// GetConfiguredWebhooks returns a list of configured webhook types for the user
func (ws *WebhookService) GetConfiguredWebhooks(settings *models.UserSettings) []string {
	var configured []string

	if ws.IsWebhookConfigured("slack", settings) {
		configured = append(configured, "slack")
	}

	if ws.IsWebhookConfigured("telegram", settings) {
		configured = append(configured, "telegram")
	}

	if ws.IsWebhookConfigured("discord", settings) {
		configured = append(configured, "discord")
	}

	return configured
}

// SendToAllConfiguredWebhooks sends a notification to all configured webhook platforms
func (ws *WebhookService) SendToAllConfiguredWebhooks(
	settings *models.UserSettings,
	data TemplateData,
) map[string]error {
	results := make(map[string]error)
	configured := ws.GetConfiguredWebhooks(settings)

	for _, webhookType := range configured {
		err := ws.SendNotification(webhookType, settings, data)
		if err != nil {
			results[webhookType] = err
		} else {
			results[webhookType] = nil
		}
	}

	return results
}

