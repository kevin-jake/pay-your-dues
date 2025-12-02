package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"pay-your-dues/internal/domain/interfaces"
	"pay-your-dues/internal/models"
)

// userSettingsService implements the UserSettingsService interface
type userSettingsService struct {
	settingsRepo interfaces.UserSettingsRepository
	userRepo     interfaces.UserRepository
	logger       zerolog.Logger
}

// NewUserSettingsService creates a new user settings service
func NewUserSettingsService(
	settingsRepo interfaces.UserSettingsRepository,
	userRepo interfaces.UserRepository,
	logger zerolog.Logger,
) interfaces.UserSettingsService {
	return &userSettingsService{
		settingsRepo: settingsRepo,
		userRepo:     userRepo,
		logger:       logger,
	}
}

// GetUserSettings retrieves user settings for a given user ID
func (s *userSettingsService) GetUserSettings(ctx context.Context, userID uuid.UUID) (*models.UserSettings, error) {
	// Verify user exists
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Get or create settings (creates defaults if not found)
	settings, err := s.settingsRepo.GetOrCreate(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user settings: %w", err)
	}

	return settings, nil
}

// UpdateUserSettings updates user settings for a given user ID
func (s *userSettingsService) UpdateUserSettings(ctx context.Context, userID uuid.UUID, req *models.UpdateUserSettingsRequest) (*models.UserSettings, error) {
	// Verify user exists
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Get or create existing settings
	settings, err := s.settingsRepo.GetOrCreate(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user settings: %w", err)
	}

	// Update fields if provided
	if req.NotificationEmail != nil {
		settings.NotificationEmail = *req.NotificationEmail
	}
	if req.NotificationSMS != nil {
		settings.NotificationSMS = *req.NotificationSMS
	}
	if req.NotificationWebhook != nil {
		settings.NotificationWebhook = *req.NotificationWebhook
	}
	if req.NotificationReminderDays != nil {
		settings.NotificationReminderDays = *req.NotificationReminderDays
	}
	if req.NotificationTime != nil {
		settings.NotificationTime = *req.NotificationTime
	}
	if req.OverdueReminderFrequency != nil {
		settings.OverdueReminderFrequency = *req.OverdueReminderFrequency
	}
	if req.CustomEmailMessage != nil {
		settings.CustomEmailMessage = req.CustomEmailMessage
	}
	if req.CustomSMSMessage != nil {
		settings.CustomSMSMessage = req.CustomSMSMessage
	}
	if req.SlackWebhookURL != nil {
		settings.SlackWebhookURL = req.SlackWebhookURL
	}
	if req.TelegramBotToken != nil {
		settings.TelegramBotToken = req.TelegramBotToken
	}
	if req.TelegramChatID != nil {
		settings.TelegramChatID = req.TelegramChatID
	}
	if req.DiscordWebhookURL != nil {
		settings.DiscordWebhookURL = req.DiscordWebhookURL
	}
	if req.EventNotificationsEnabled != nil {
		settings.EventNotificationsEnabled = *req.EventNotificationsEnabled
	}
	if req.NotifyContactOnPayment != nil {
		settings.NotifyContactOnPayment = *req.NotifyContactOnPayment
	}
	if req.DefaultCurrency != nil {
		settings.DefaultCurrency = *req.DefaultCurrency
	}
	if req.Timezone != nil {
		settings.Timezone = *req.Timezone
	}

	// Update timestamp
	settings.UpdatedAt = time.Now()

	// Save updated settings
	if err := s.settingsRepo.Update(ctx, settings); err != nil {
		return nil, fmt.Errorf("failed to update user settings: %w", err)
	}

	s.logger.Info().
		Str("user_id", userID.String()).
		Msg("User settings updated successfully")

	return settings, nil
}

// GetOrCreateUserSettings gets or creates user settings (used internally)
func (s *userSettingsService) GetOrCreateUserSettings(ctx context.Context, userID uuid.UUID) (*models.UserSettings, error) {
	return s.settingsRepo.GetOrCreate(ctx, userID)
}

