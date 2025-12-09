package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"pay-your-dues/internal/domain/interfaces"
)

// TelegramSubscriptionService handles Telegram bot subscriptions
type TelegramSubscriptionService struct {
	userSettingsRepo interfaces.UserSettingsRepository
	userRepo         interfaces.UserRepository
	botToken         string
	logger           zerolog.Logger

	// In-memory store for pending link codes (code -> pendingLink)
	// In production, consider using Redis or database with TTL
	pendingLinks   map[string]pendingLink
	pendingLinksMu sync.RWMutex

	// Long polling support
	pollingStopChan chan struct{}
	pollingRunning  bool
	pollingMu       sync.Mutex
}

type pendingLink struct {
	UserID    uuid.UUID
	ExpiresAt time.Time
}

// TelegramUpdate represents an incoming Telegram update
type TelegramUpdate struct {
	UpdateID int64            `json:"update_id"`
	Message  *TelegramMessage `json:"message"`
}

// TelegramMessage represents a Telegram message
type TelegramMessage struct {
	MessageID int64         `json:"message_id"`
	From      *TelegramUser `json:"from"`
	Chat      *TelegramChat `json:"chat"`
	Date      int64         `json:"date"`
	Text      string        `json:"text"`
}

// TelegramUser represents a Telegram user
type TelegramUser struct {
	ID        int64  `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
}

// TelegramChat represents a Telegram chat
type TelegramChat struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title,omitempty"`
	Username  string `json:"username,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// NewTelegramSubscriptionService creates a new Telegram subscription service
func NewTelegramSubscriptionService(
	userSettingsRepo interfaces.UserSettingsRepository,
	userRepo interfaces.UserRepository,
	botToken string,
	logger zerolog.Logger,
) *TelegramSubscriptionService {
	return &TelegramSubscriptionService{
		userSettingsRepo: userSettingsRepo,
		userRepo:         userRepo,
		botToken:         botToken,
		logger:           logger.With().Str("service", "telegram_subscription").Logger(),
		pendingLinks:     make(map[string]pendingLink),
	}
}

// GenerateLinkCode creates a unique code for a user to link their Telegram account
func (s *TelegramSubscriptionService) GenerateLinkCode(ctx context.Context, userID uuid.UUID) (string, error) {
	// Verify user exists
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("user not found: %w", err)
	}

	// Generate a random 6-character code
	bytes := make([]byte, 3)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate code: %w", err)
	}
	code := hex.EncodeToString(bytes)

	// Store the pending link (expires in 10 minutes)
	s.pendingLinksMu.Lock()
	s.pendingLinks[code] = pendingLink{
		UserID:    userID,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	s.pendingLinksMu.Unlock()

	s.logger.Info().
		Str("user_id", userID.String()).
		Str("code", code).
		Msg("Generated Telegram link code")

	return code, nil
}

// HandleWebhookUpdate processes an incoming Telegram update
func (s *TelegramSubscriptionService) HandleWebhookUpdate(ctx context.Context, update *TelegramUpdate) error {
	if update.Message == nil {
		return nil // Ignore non-message updates
	}

	chatID := update.Message.Chat.ID
	text := update.Message.Text
	username := ""
	if update.Message.From != nil {
		username = update.Message.From.Username
	}

	s.logger.Info().
		Int64("chat_id", chatID).
		Str("text", text).
		Str("username", username).
		Msg("Received Telegram message")

	// Handle /start command
	if text == "/start" {
		return s.sendMessage(chatID,
			"👋 Welcome to Pay Your Dues notifications!\n\n"+
				"To link your account, please send the 6-character code "+
				"shown in your app settings.\n\n"+
				"Example: Send `abc123` to link your account.")
	}

	// Handle /status command
	if text == "/status" {
		return s.sendMessage(chatID,
			"ℹ️ To check your subscription status, please visit the app settings page.")
	}

	// Handle /help command
	if text == "/help" {
		return s.sendMessage(chatID,
			"📚 *Pay Your Dues Bot Help*\n\n"+
				"Commands:\n"+
				"• `/start` - Get started with linking your account\n"+
				"• `/status` - Check subscription status\n"+
				"• `/help` - Show this help message\n\n"+
				"To receive notifications:\n"+
				"1. Go to your app settings\n"+
				"2. Click 'Connect Telegram'\n"+
				"3. Send the code here to link your account")
	}

	// Try to match link code (6 hex characters)
	if len(text) == 6 {
		s.pendingLinksMu.Lock()
		pending, exists := s.pendingLinks[text]
		if exists && time.Now().Before(pending.ExpiresAt) {
			delete(s.pendingLinks, text) // Use the code
			s.pendingLinksMu.Unlock()

			// Link the account
			if err := s.linkAccount(ctx, pending.UserID, chatID, username); err != nil {
				s.logger.Error().Err(err).Msg("Failed to link Telegram account")
				return s.sendMessage(chatID, "❌ Failed to link your account. Please try again.")
			}

			return s.sendMessage(chatID,
				"✅ *Account linked successfully!*\n\n"+
					"You will now receive payment notifications here.\n\n"+
					"To unlink your account, visit the app settings.")
		}
		s.pendingLinksMu.Unlock()
	}

	// Unknown message or expired/invalid code
	return s.sendMessage(chatID,
		"❓ Unknown command or invalid/expired code.\n\n"+
			"Send /start to see instructions, or /help for more info.")
}

// linkAccount saves the Telegram chat ID to user settings
func (s *TelegramSubscriptionService) linkAccount(ctx context.Context, userID uuid.UUID, chatID int64, username string) error {
	settings, err := s.userSettingsRepo.GetOrCreate(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user settings: %w", err)
	}

	// Store the chat ID (as string for consistency with existing schema)
	chatIDStr := fmt.Sprintf("%d", chatID)
	settings.TelegramChatID = &chatIDStr

	if err := s.userSettingsRepo.Update(ctx, settings); err != nil {
		return fmt.Errorf("failed to update user settings: %w", err)
	}

	s.logger.Info().
		Str("user_id", userID.String()).
		Int64("chat_id", chatID).
		Str("username", username).
		Msg("Telegram account linked successfully")

	return nil
}

// sendMessage sends a message to a Telegram chat
func (s *TelegramSubscriptionService) sendMessage(chatID int64, text string) error {
	if s.botToken == "" {
		return fmt.Errorf("Telegram bot token not configured")
	}

	bot, err := tgbotapi.NewBotAPI(s.botToken)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"

	_, err = bot.Send(msg)
	return err
}

// UnlinkAccount removes Telegram subscription for a user
func (s *TelegramSubscriptionService) UnlinkAccount(ctx context.Context, userID uuid.UUID) error {
	settings, err := s.userSettingsRepo.GetOrCreate(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user settings: %w", err)
	}

	// Send farewell message if chat ID exists
	if settings.TelegramChatID != nil && *settings.TelegramChatID != "" {
		var chatID int64
		if _, err := fmt.Sscanf(*settings.TelegramChatID, "%d", &chatID); err == nil {
			_ = s.sendMessage(chatID,
				"👋 Your Telegram account has been unlinked from Pay Your Dues.\n\n"+
					"You will no longer receive notifications here.\n"+
					"Send /start to re-link your account anytime.")
		}
	}

	settings.TelegramChatID = nil

	return s.userSettingsRepo.Update(ctx, settings)
}

// GetBotToken returns the shared bot token
func (s *TelegramSubscriptionService) GetBotToken() string {
	return s.botToken
}

// GetBotInfo returns information about the bot
func (s *TelegramSubscriptionService) GetBotInfo() (*tgbotapi.User, error) {
	if s.botToken == "" {
		return nil, fmt.Errorf("Telegram bot token not configured")
	}

	bot, err := tgbotapi.NewBotAPI(s.botToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	return &bot.Self, nil
}

// IsConfigured returns true if the Telegram bot is configured
func (s *TelegramSubscriptionService) IsConfigured() bool {
	return s.botToken != ""
}

// CleanupExpiredCodes removes expired link codes from memory
// Call this periodically (e.g., every hour) to prevent memory leaks
func (s *TelegramSubscriptionService) CleanupExpiredCodes() {
	s.pendingLinksMu.Lock()
	defer s.pendingLinksMu.Unlock()

	now := time.Now()
	for code, link := range s.pendingLinks {
		if now.After(link.ExpiresAt) {
			delete(s.pendingLinks, code)
		}
	}
}

// ============================================================================
// Long Polling Support (for development without a webhook/domain)
// ============================================================================

// StartLongPolling starts polling Telegram for updates
// Use this for local development when you don't have a public domain for webhooks
func (s *TelegramSubscriptionService) StartLongPolling() error {
	if s.botToken == "" {
		return fmt.Errorf("Telegram bot token not configured")
	}

	s.pollingMu.Lock()
	if s.pollingRunning {
		s.pollingMu.Unlock()
		return fmt.Errorf("long polling is already running")
	}

	bot, err := tgbotapi.NewBotAPI(s.botToken)
	if err != nil {
		s.pollingMu.Unlock()
		return fmt.Errorf("failed to create bot: %w", err)
	}

	// Remove any existing webhook (required for long polling)
	_, err = bot.Request(tgbotapi.DeleteWebhookConfig{})
	if err != nil {
		s.logger.Warn().Err(err).Msg("Failed to delete existing webhook")
	}

	s.pollingStopChan = make(chan struct{})
	s.pollingRunning = true
	s.pollingMu.Unlock()

	s.logger.Info().
		Str("bot_username", bot.Self.UserName).
		Msg("Starting Telegram long polling mode")

	go s.pollUpdates(bot)

	return nil
}

// StopLongPolling stops the long polling loop
func (s *TelegramSubscriptionService) StopLongPolling() {
	s.pollingMu.Lock()
	defer s.pollingMu.Unlock()

	if !s.pollingRunning {
		return
	}

	close(s.pollingStopChan)
	s.pollingRunning = false
	s.logger.Info().Msg("Stopped Telegram long polling")
}

// IsPollingRunning returns true if long polling is active
func (s *TelegramSubscriptionService) IsPollingRunning() bool {
	s.pollingMu.Lock()
	defer s.pollingMu.Unlock()
	return s.pollingRunning
}

// pollUpdates is the main long polling loop
func (s *TelegramSubscriptionService) pollUpdates(bot *tgbotapi.BotAPI) {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30 // Long polling timeout in seconds

	updates := bot.GetUpdatesChan(updateConfig)

	for {
		select {
		case <-s.pollingStopChan:
			s.logger.Info().Msg("Long polling stopped")
			return
		case update := <-updates:
			if update.Message == nil {
				continue
			}

			// Convert to our TelegramUpdate type
			telegramUpdate := &TelegramUpdate{
				UpdateID: int64(update.UpdateID),
				Message: &TelegramMessage{
					MessageID: int64(update.Message.MessageID),
					Text:      update.Message.Text,
					Date:      int64(update.Message.Date),
					Chat: &TelegramChat{
						ID:        update.Message.Chat.ID,
						Type:      update.Message.Chat.Type,
						Title:     update.Message.Chat.Title,
						Username:  update.Message.Chat.UserName,
						FirstName: update.Message.Chat.FirstName,
						LastName:  update.Message.Chat.LastName,
					},
				},
			}

			if update.Message.From != nil {
				telegramUpdate.Message.From = &TelegramUser{
					ID:        update.Message.From.ID,
					IsBot:     update.Message.From.IsBot,
					FirstName: update.Message.From.FirstName,
					LastName:  update.Message.From.LastName,
					Username:  update.Message.From.UserName,
				}
			}

			// Process the update
			ctx := context.Background()
			if err := s.HandleWebhookUpdate(ctx, telegramUpdate); err != nil {
				s.logger.Error().Err(err).Msg("Failed to handle Telegram update")
			}
		}
	}
}

