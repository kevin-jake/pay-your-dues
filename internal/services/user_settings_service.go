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
	settingsRepo     interfaces.UserSettingsRepository
	userRepo         interfaces.UserRepository
	notificationRepo interfaces.NotificationRepository
	logger           zerolog.Logger
}

// NewUserSettingsService creates a new user settings service
func NewUserSettingsService(
	settingsRepo interfaces.UserSettingsRepository,
	userRepo interfaces.UserRepository,
	notificationRepo interfaces.NotificationRepository,
	logger zerolog.Logger,
) interfaces.UserSettingsService {
	return &userSettingsService{
		settingsRepo:     settingsRepo,
		userRepo:         userRepo,
		notificationRepo: notificationRepo,
		logger:           logger,
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

	// Track if notification-related settings changed
	notificationSettingsChanged := false

	// Update fields if provided
	if req.NotificationEmail != nil {
		settings.NotificationEmail = *req.NotificationEmail
		notificationSettingsChanged = true
	}
	if req.NotificationSMS != nil {
		settings.NotificationSMS = *req.NotificationSMS
		notificationSettingsChanged = true
	}
	if req.NotificationWebhook != nil {
		settings.NotificationWebhook = *req.NotificationWebhook
		notificationSettingsChanged = true
	}
	if req.NotificationReminderDays != nil {
		settings.NotificationReminderDays = *req.NotificationReminderDays
		notificationSettingsChanged = true
	}
	if req.NotificationTime != nil {
		settings.NotificationTime = *req.NotificationTime
		notificationSettingsChanged = true
	}
	if req.OverdueReminderFrequency != nil {
		settings.OverdueReminderFrequency = *req.OverdueReminderFrequency
		notificationSettingsChanged = true
	}
	if req.CustomEmailMessage != nil {
		settings.CustomEmailMessage = req.CustomEmailMessage
	}
	if req.CustomSMSMessage != nil {
		settings.CustomSMSMessage = req.CustomSMSMessage
	}
	if req.SlackWebhookURL != nil {
		settings.SlackWebhookURL = req.SlackWebhookURL
		notificationSettingsChanged = true
	}
	// Note: TelegramChatID is managed via the Telegram bot subscription flow
	if req.DiscordWebhookURL != nil {
		settings.DiscordWebhookURL = req.DiscordWebhookURL
		notificationSettingsChanged = true
	}
	if req.EventNotificationsEnabled != nil {
		settings.EventNotificationsEnabled = *req.EventNotificationsEnabled
		notificationSettingsChanged = true
	}
	if req.NotifyContactOnPayment != nil {
		settings.NotifyContactOnPayment = *req.NotifyContactOnPayment
		notificationSettingsChanged = true
	}
	if req.NotificationRecipient != nil {
		settings.NotificationRecipient = *req.NotificationRecipient
		notificationSettingsChanged = true
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

	// If notification settings changed, update pending notifications
	if notificationSettingsChanged && s.notificationRepo != nil {
		if err := s.updatePendingNotifications(userID, settings); err != nil {
			// Log error but don't fail the settings update
			s.logger.Error().
				Err(err).
				Str("user_id", userID.String()).
				Msg("Failed to update pending notifications after settings change")
		} else {
			s.logger.Info().
				Str("user_id", userID.String()).
				Msg("Pending notifications updated with new settings")
		}
	}

	s.logger.Info().
		Str("user_id", userID.String()).
		Msg("User settings updated successfully")

	return settings, nil
}

// updatePendingNotifications updates pending notifications based on new user settings
func (s *userSettingsService) updatePendingNotifications(userID uuid.UUID, settings *models.UserSettings) error {
	// Get pending notifications for this user
	pendingNotifications, err := s.notificationRepo.GetPendingNotificationsByUserID(userID)
	if err != nil {
		return fmt.Errorf("failed to get pending notifications: %w", err)
	}

	if len(pendingNotifications) == 0 {
		return nil // No pending notifications to update
	}

	// Determine which notification types should be enabled
	enabledTypes := make(map[string]bool)
	enabledTypes["email"] = settings.NotificationEmail
	enabledTypes["sms"] = settings.NotificationSMS
	
	// Webhook types enabled based on webhook setting AND configured URLs/subscriptions
	if settings.NotificationWebhook {
		if settings.SlackWebhookURL != nil && *settings.SlackWebhookURL != "" {
			enabledTypes["slack"] = true
		}
		// Telegram uses app-level bot token, only need chat ID (set via bot subscription)
		if settings.TelegramChatID != nil && *settings.TelegramChatID != "" {
			enabledTypes["telegram"] = true
		}
		if settings.DiscordWebhookURL != nil && *settings.DiscordWebhookURL != "" {
			enabledTypes["discord"] = true
		}
	}

	// Determine recipient filter
	recipientSetting := settings.NotificationRecipient
	if recipientSetting == "" {
		recipientSetting = "both"
	}

	// Update each pending notification
	updatedCount := 0
	disabledCount := 0
	for _, notif := range pendingNotifications {
		shouldEnable := true

		// Check if notification type is still enabled
		if enabled, exists := enabledTypes[notif.NotificationType]; exists {
			if !enabled {
				shouldEnable = false
			}
		}

		// Check if event notifications are enabled (for event-type notifications)
		if notif.ScheduleType == "event" && !settings.EventNotificationsEnabled {
			shouldEnable = false
		}

		// Check recipient type against settings
		if recipientSetting != "both" {
			if notif.RecipientType != recipientSetting {
				shouldEnable = false
			}
		}

		// Check notify contact setting
		if notif.RecipientType == "contact" && !settings.NotifyContactOnPayment {
			shouldEnable = false
		}

		// Update notification enabled status
		if notif.Enabled != shouldEnable {
			if shouldEnable {
				if err := s.notificationRepo.EnableNotification(notif.ID); err != nil {
					s.logger.Error().Err(err).Str("notification_id", notif.ID.String()).Msg("Failed to enable notification")
					continue
				}
			} else {
				if err := s.notificationRepo.DisableNotification(notif.ID); err != nil {
					s.logger.Error().Err(err).Str("notification_id", notif.ID.String()).Msg("Failed to disable notification")
					continue
				}
				disabledCount++
			}
			updatedCount++
		}
	}

	s.logger.Info().
		Int("total_pending", len(pendingNotifications)).
		Int("updated", updatedCount).
		Int("disabled", disabledCount).
		Str("user_id", userID.String()).
		Msg("Pending notifications processed")

	return nil
}

// GetOrCreateUserSettings gets or creates user settings (used internally)
func (s *userSettingsService) GetOrCreateUserSettings(ctx context.Context, userID uuid.UUID) (*models.UserSettings, error) {
	return s.settingsRepo.GetOrCreate(ctx, userID)
}

