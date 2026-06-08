package services

import (
	"context"
	"fmt"
	"time"

	"pay-your-dues/internal/domain/entities"
	"pay-your-dues/internal/domain/interfaces"
	"pay-your-dues/internal/messaging"
	"pay-your-dues/internal/models"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// NotificationService implements the NotificationService interface for the API layer.
type NotificationService struct {
	notificationRepo interfaces.NotificationRepository
	templateRepo     interfaces.NotificationTemplateRepository
	debtListRepo     interfaces.DebtListRepository
	userSettingsRepo interfaces.UserSettingsRepository
	publisher        interfaces.NotificationPublisher
	logger           zerolog.Logger
}

// NewNotificationService creates a new notification service instance for the API.
func NewNotificationService(
	notificationRepo interfaces.NotificationRepository,
	templateRepo interfaces.NotificationTemplateRepository,
	debtListRepo interfaces.DebtListRepository,
	userSettingsRepo interfaces.UserSettingsRepository,
	publisher interfaces.NotificationPublisher,
	logger zerolog.Logger,
) interfaces.NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		templateRepo:     templateRepo,
		debtListRepo:     debtListRepo,
		userSettingsRepo: userSettingsRepo,
		publisher:        publisher,
		logger:           logger,
	}
}

// CreateNotification creates a new notification
func (s *NotificationService) CreateNotification(userID uuid.UUID, req *models.CreateNotificationRequest) (*models.Notification, error) {
	ctx := context.Background()
	
	// Verify user owns the debt list
	debtList, err := s.debtListRepo.GetByID(ctx, req.DebtListID)
	if err != nil {
		return nil, fmt.Errorf("debt list not found: %w", err)
	}

	if debtList.UserID != userID {
		return nil, fmt.Errorf("unauthorized: user does not own this debt list")
	}

	// Create notification
	now := time.Now()
	scheduledFor := req.ScheduledFor
	if scheduledFor == nil {
		// Default to now if not specified
		scheduledFor = &now
	}
	
	scheduleType := req.ScheduleType
	if scheduleType == "" {
		scheduleType = "manual"
	}
	
	notification := &models.Notification{
		ID:                 uuid.New(),
		DebtListID:         req.DebtListID,
		InstallmentNumber:  req.InstallmentNumber,
		InstallmentDueDate: req.InstallmentDueDate,
		NotificationType:   req.NotificationType,
		WebhookType:        req.WebhookType,
		RecipientType:      req.RecipientType,
		Message:            req.Message,
		Status:             "pending",
		ScheduleType:       scheduleType,
		ScheduledFor:       scheduledFor,
		NextRunAt:          scheduledFor, // Set next_run_at to scheduled_for for worker processing
		Enabled:            true,
		RecipientEmail:     req.RecipientEmail,
		RecipientPhone:     req.RecipientPhone,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := s.notificationRepo.Create(notification); err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	if scheduledFor != nil && !scheduledFor.After(time.Now()) {
		s.publishImmediate(notification.ID, messaging.JobTypeImmediate)
	}

	return notification, nil
}

// CreateNotificationsForDebtList creates notifications for all installments of a debt list
func (s *NotificationService) CreateNotificationsForDebtList(userID uuid.UUID, debtListID uuid.UUID) error {
	ctx := context.Background()
	
	// Verify user owns the debt list
	belongs, err := s.debtListRepo.BelongsToUser(ctx, debtListID, userID)
	if err != nil {
		return fmt.Errorf("failed to verify ownership: %w", err)
	}
	if !belongs {
		return fmt.Errorf("unauthorized: user does not own this debt list")
	}

	// Get debt list details
	debtList, err := s.debtListRepo.GetByID(ctx, debtListID)
	if err != nil {
		return fmt.Errorf("debt list not found: %w", err)
	}

	// Get user settings for notification preferences
	userSettings, err := s.getUserSettings(userID)
	if err != nil {
		return fmt.Errorf("failed to get user settings: %w", err)
	}

	// If it's a one-time debt, create single notification
	if debtList.InstallmentPlan == "onetime" {
		return s.createOneTimeNotifications(s.entityToModelDebtList(debtList), userSettings)
	}

	// For installment debts, create notifications for each installment
	return s.createInstallmentNotifications(s.entityToModelDebtList(debtList), userSettings)
}

// createOneTimeNotifications creates notifications for one-time debts
func (s *NotificationService) createOneTimeNotifications(debtList *models.DebtList, settings *models.UserSettings) error {
	reminderDays := []int64{7, 3, 1} // Default
	if len(settings.NotificationReminderDays) > 0 {
		reminderDays = settings.NotificationReminderDays
	}
	notificationTime := "09:00:00"
	if settings.NotificationTime != "" {
		notificationTime = settings.NotificationTime
	}

	// Determine which notification types to create based on user settings
	notificationTypes := s.getEnabledNotificationTypes(settings)
	if len(notificationTypes) == 0 {
		s.logger.Info().
			Str("debt_list_id", debtList.ID.String()).
			Msg("No notification types enabled, skipping notification creation")
		return nil
	}

	var notifications []*models.Notification

	for _, daysBefore := range reminderDays {
		scheduledFor := debtList.DueDate.AddDate(0, 0, -int(daysBefore))
		
		// Parse notification time (try with seconds first, then without)
		t, err := time.Parse("15:04:05", notificationTime)
		if err != nil {
			t, _ = time.Parse("15:04", notificationTime)
		}
		scheduledFor = time.Date(
			scheduledFor.Year(), scheduledFor.Month(), scheduledFor.Day(),
			t.Hour(), t.Minute(), 0, 0, scheduledFor.Location(),
		)

		// Only create if scheduled time is in the future
		if scheduledFor.After(time.Now()) {
			daysBefore32 := int(daysBefore)
			
			// Create a notification for each enabled notification type
			for _, notifType := range notificationTypes {
				notification := &models.Notification{
					ID:                 uuid.New(),
					DebtListID:         debtList.ID,
					NotificationType:   notifType,
					RecipientType:      "user",
					Message:            "Payment reminder",
					Status:             "pending",
					ScheduleType:       "reminder",
					ScheduledFor:       &scheduledFor,
					ReminderDaysBefore: &daysBefore32,
					Enabled:            true,
					CreatedAt:          time.Now(),
					UpdatedAt:          time.Now(),
				}

				notifications = append(notifications, notification)
			}
		}
	}

	if len(notifications) > 0 {
		return s.notificationRepo.CreateBatch(notifications)
	}

	return nil
}

// getEnabledNotificationTypes returns a list of enabled notification types based on user settings
func (s *NotificationService) getEnabledNotificationTypes(settings *models.UserSettings) []string {
	var types []string
	
	if settings.NotificationEmail {
		types = append(types, "email")
	}
	if settings.NotificationSMS {
		types = append(types, "sms")
	}
	if settings.NotificationWebhook {
		// Add webhook types based on configured webhooks
		if settings.SlackWebhookURL != nil && *settings.SlackWebhookURL != "" {
			types = append(types, "slack")
		}
		// Telegram uses app-level bot token, only need chat ID (set via bot subscription)
		if settings.TelegramChatID != nil && *settings.TelegramChatID != "" {
			types = append(types, "telegram")
		}
		if settings.DiscordWebhookURL != nil && *settings.DiscordWebhookURL != "" {
			types = append(types, "discord")
		}
	}
	
	return types
}

// createInstallmentNotifications creates notifications for installment debts
func (s *NotificationService) createInstallmentNotifications(debtList *models.DebtList, settings *models.UserSettings) error {
	if debtList.NumberOfPayments == nil || *debtList.NumberOfPayments == 0 {
		return fmt.Errorf("invalid number of payments for installment debt")
	}

	reminderDays := []int64{7, 3, 1} // Default
	if len(settings.NotificationReminderDays) > 0 {
		reminderDays = settings.NotificationReminderDays
	}

	notificationTime := "09:00:00"
	if settings.NotificationTime != "" {
		notificationTime = settings.NotificationTime
	}

	// Determine which notification types to create based on user settings
	notificationTypes := s.getEnabledNotificationTypes(settings)
	if len(notificationTypes) == 0 {
		s.logger.Info().
			Str("debt_list_id", debtList.ID.String()).
			Msg("No notification types enabled, skipping notification creation")
		return nil
	}

	var notifications []*models.Notification
	currentDueDate := debtList.DueDate

	for i := 1; i <= *debtList.NumberOfPayments; i++ {
		installmentNumber := i

		for _, daysBefore := range reminderDays {
			scheduledFor := currentDueDate.AddDate(0, 0, -int(daysBefore))
			
			// Parse notification time (try with seconds first, then without)
			t, err := time.Parse("15:04:05", notificationTime)
			if err != nil {
				t, _ = time.Parse("15:04", notificationTime)
			}
			scheduledFor = time.Date(
				scheduledFor.Year(), scheduledFor.Month(), scheduledFor.Day(),
				t.Hour(), t.Minute(), 0, 0, scheduledFor.Location(),
			)

			// Only create if scheduled time is in the future
			if scheduledFor.After(time.Now()) {
				daysBefore32 := int(daysBefore)
				
				// Create a notification for each enabled notification type
				for _, notifType := range notificationTypes {
					notification := &models.Notification{
						ID:                 uuid.New(),
						DebtListID:         debtList.ID,
						InstallmentNumber:  &installmentNumber,
						InstallmentDueDate: &currentDueDate,
						NotificationType:   notifType,
						RecipientType:      "user",
						Message:            fmt.Sprintf("Installment #%d payment reminder", installmentNumber),
						Status:             "pending",
						ScheduleType:       "reminder",
						ScheduledFor:       &scheduledFor,
						ReminderDaysBefore: &daysBefore32,
						Enabled:            true,
						CreatedAt:          time.Now(),
						UpdatedAt:          time.Now(),
					}

					notifications = append(notifications, notification)
				}
			}
		}

		// Calculate next due date based on installment plan
		currentDueDate = s.calculateNextDueDate(currentDueDate, debtList.InstallmentPlan)
	}

	if *debtList.NumberOfPayments == 1 {

		for _, daysBefore := range reminderDays {
			scheduledFor := currentDueDate.AddDate(0, 0, -int(daysBefore))
			
			// Parse notification time (try with seconds first, then without)
			t, err := time.Parse("15:04:05", notificationTime)
			if err != nil {
				t, _ = time.Parse("15:04", notificationTime)
			}
			scheduledFor = time.Date(
				scheduledFor.Year(), scheduledFor.Month(), scheduledFor.Day(),
				t.Hour(), t.Minute(), 0, 0, scheduledFor.Location(),
			)

			// Only create if scheduled time is in the future
			if scheduledFor.After(time.Now()) {
				daysBefore32 := int(daysBefore)
				
				// Create a notification for each enabled notification type
				for _, notifType := range notificationTypes {
					notification := &models.Notification{
						ID:                 uuid.New(),
						DebtListID:         debtList.ID,
						NotificationType:   notifType,
						RecipientType:      "user",
						Status:             "pending",
						ScheduleType:       "reminder",
						ScheduledFor:       &scheduledFor,
						ReminderDaysBefore: &daysBefore32,
						Enabled:            true,
						CreatedAt:          time.Now(),
						UpdatedAt:          time.Now(),
					}

					notifications = append(notifications, notification)
				}
			}
		}
	}

	if len(notifications) > 0 {
		s.logger.Info().
			Str("debt_list_id", debtList.ID.String()).
			Int("notification_count", len(notifications)).
			Msg("Creating installment notifications")

		return s.notificationRepo.CreateBatch(notifications)
	}

	return nil
}

// calculateNextDueDate calculates the next due date based on installment plan
func (s *NotificationService) calculateNextDueDate(currentDate time.Time, plan string) time.Time {
	switch plan {
	case "weekly":
		return currentDate.AddDate(0, 0, 7)
	case "biweekly":
		return currentDate.AddDate(0, 0, 14)
	case "monthly":
		return currentDate.AddDate(0, 1, 0)
	case "quarterly":
		return currentDate.AddDate(0, 3, 0)
	case "yearly":
		return currentDate.AddDate(1, 0, 0)
	default:
		return currentDate.AddDate(0, 1, 0)
	}
}

// CreateNotificationsForInstallment creates notifications for a specific installment
func (s *NotificationService) CreateNotificationsForInstallment(debtListID uuid.UUID, installmentNumber int, installmentDueDate time.Time) error {
	ctx := context.Background()
	
	debtList, err := s.debtListRepo.GetByID(ctx, debtListID)
	if err != nil {
		return fmt.Errorf("debt list not found: %w", err)
	}

	userSettings, err := s.getUserSettings(debtList.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user settings: %w", err)
	}

	reminderDays := []int64{7, 3, 1}
	if len(userSettings.NotificationReminderDays) > 0 {
		reminderDays = userSettings.NotificationReminderDays
	}

	// Determine which notification types to create based on user settings
	notificationTypes := s.getEnabledNotificationTypes(userSettings)
	if len(notificationTypes) == 0 {
		s.logger.Info().
			Str("debt_list_id", debtListID.String()).
			Int("installment_number", installmentNumber).
			Msg("No notification types enabled, skipping notification creation")
		return nil
	}

	var notifications []*models.Notification

	for _, daysBefore := range reminderDays {
		scheduledFor := installmentDueDate.AddDate(0, 0, -int(daysBefore))
		
		if scheduledFor.After(time.Now()) {
			daysBefore32 := int(daysBefore)
			
			// Create a notification for each enabled notification type
			for _, notifType := range notificationTypes {
				notification := &models.Notification{
					ID:                 uuid.New(),
					DebtListID:         debtListID,
					InstallmentNumber:  &installmentNumber,
					InstallmentDueDate: &installmentDueDate,
					NotificationType:   notifType,
					RecipientType:      "user",
					Message:            fmt.Sprintf("Installment #%d payment reminder", installmentNumber),
					Status:             "pending",
					ScheduleType:       "reminder",
					ScheduledFor:       &scheduledFor,
					ReminderDaysBefore: &daysBefore32,
					Enabled:            true,
					CreatedAt:          time.Now(),
					UpdatedAt:          time.Now(),
				}

				notifications = append(notifications, notification)
			}
		}
	}

	if len(notifications) > 0 {
		return s.notificationRepo.CreateBatch(notifications)
	}

	return nil
}

// GetNotification retrieves a notification by ID
func (s *NotificationService) GetNotification(userID uuid.UUID, notificationID uuid.UUID) (*models.Notification, error) {
	ctx := context.Background()
	
	notification, err := s.notificationRepo.GetByID(notificationID)
	if err != nil {
		return nil, fmt.Errorf("notification not found: %w", err)
	}

	// Verify ownership
	debtList, err := s.debtListRepo.GetByID(ctx, notification.DebtListID)
	if err != nil {
		return nil, fmt.Errorf("debt list not found: %w", err)
	}

	if debtList.UserID != userID {
		return nil, fmt.Errorf("unauthorized: user does not own this notification")
	}

	return notification, nil
}

// GetNotificationsByDebtList retrieves all notifications for a debt list
func (s *NotificationService) GetNotificationsByDebtList(userID uuid.UUID, debtListID uuid.UUID) ([]*models.Notification, error) {
	ctx := context.Background()
	
	// Verify ownership
	belongs, err := s.debtListRepo.BelongsToUser(ctx, debtListID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ownership: %w", err)
	}
	if !belongs {
		return nil, fmt.Errorf("unauthorized: user does not own this debt list")
	}

	return s.notificationRepo.GetByDebtListID(debtListID)
}

// GetNotificationsByInstallment retrieves notifications for a specific installment
func (s *NotificationService) GetNotificationsByInstallment(userID uuid.UUID, debtListID uuid.UUID, installmentNumber int) ([]*models.Notification, error) {
	ctx := context.Background()
	
	// Verify ownership
	belongs, err := s.debtListRepo.BelongsToUser(ctx, debtListID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ownership: %w", err)
	}
	if !belongs {
		return nil, fmt.Errorf("unauthorized: user does not own this debt list")
	}

	return s.notificationRepo.GetByDebtListAndInstallment(debtListID, installmentNumber)
}

// UpdateNotification updates a notification
func (s *NotificationService) UpdateNotification(userID uuid.UUID, notificationID uuid.UUID, enabled bool) error {
	notification, err := s.GetNotification(userID, notificationID)
	if err != nil {
		return err
	}

	notification.Enabled = enabled
	notification.UpdatedAt = time.Now()

	return s.notificationRepo.Update(notification)
}

// DeleteNotification deletes a notification
func (s *NotificationService) DeleteNotification(userID uuid.UUID, notificationID uuid.UUID) error {
	// Verify ownership first
	_, err := s.GetNotification(userID, notificationID)
	if err != nil {
		return err
	}

	return s.notificationRepo.Delete(notificationID)
}

// EnableNotification enables a notification
func (s *NotificationService) EnableNotification(userID uuid.UUID, notificationID uuid.UUID) error {
	return s.UpdateNotification(userID, notificationID, true)
}

// DisableNotification disables a notification
func (s *NotificationService) DisableNotification(userID uuid.UUID, notificationID uuid.UUID) error {
	return s.UpdateNotification(userID, notificationID, false)
}

// SendNotification enqueues a specific notification for delivery.
func (s *NotificationService) SendNotification(notificationID uuid.UUID) error {
	if _, err := s.notificationRepo.GetByID(notificationID); err != nil {
		return fmt.Errorf("notification not found: %w", err)
	}
	s.publishImmediate(notificationID, messaging.JobTypeImmediate)
	return nil
}

// SendManualNotification sends a manual notification
func (s *NotificationService) SendManualNotification(userID uuid.UUID, debtListID uuid.UUID, message string, notificationType string) error {
	ctx := context.Background()
	
	// Verify ownership
	belongs, err := s.debtListRepo.BelongsToUser(ctx, debtListID, userID)
	if err != nil {
		return fmt.Errorf("failed to verify ownership: %w", err)
	}
	if !belongs {
		return fmt.Errorf("unauthorized: user does not own this debt list")
	}

	now := time.Now()
	notification := &models.Notification{
		ID:               uuid.New(),
		DebtListID:       debtListID,
		NotificationType: notificationType,
		RecipientType:    "contact",
		Message:          message,
		Status:           "pending",
		ScheduleType:     "manual",
		ScheduledFor:     &now,
		Enabled:          true,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := s.notificationRepo.Create(notification); err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	s.publishImmediate(notification.ID, messaging.JobTypeManual)
	return nil
}

func (s *NotificationService) publishImmediate(notificationID uuid.UUID, jobType string) {
	if err := s.publisher.PublishNotification(notificationID, jobType); err != nil {
		s.logger.Warn().Err(err).Str("notification_id", notificationID.String()).Msg("Failed to publish immediate notification job")
		return
	}
	if err := s.notificationRepo.MarkQueued(notificationID); err != nil {
		s.logger.Warn().Err(err).Str("notification_id", notificationID.String()).Msg("Failed to mark notification as queued")
	}
}

// ProcessPendingNotifications is handled by the notification-worker service.
func (s *NotificationService) ProcessPendingNotifications() error {
	s.logger.Debug().Msg("ProcessPendingNotifications is handled by notification-worker")
	return nil
}

// SendPaymentConfirmationNotifications enqueues payment confirmation notifications.
func (s *NotificationService) SendPaymentConfirmationNotifications(debtItemID uuid.UUID) error {
	return s.publishEventNotification(debtItemID, "Payment recorded", messaging.JobTypeEvent)
}

// SendPaymentVerificationNotification enqueues payment verification notifications.
func (s *NotificationService) SendPaymentVerificationNotification(debtItemID uuid.UUID, verified bool, reason string) error {
	message := "Payment verified"
	if !verified {
		message = "Payment rejected"
		if reason != "" {
			message = fmt.Sprintf("Payment rejected: %s", reason)
		}
	}
	return s.publishEventNotification(debtItemID, message, messaging.JobTypeEvent)
}

func (s *NotificationService) publishEventNotification(debtItemID uuid.UUID, message string, jobType string) error {
	s.logger.Info().Str("debt_item_id", debtItemID.String()).Str("job_type", jobType).Msg("Publishing event notification")
	// Event notification rows are created by the worker delivery path when extended;
	// for now publish is a no-op placeholder until event rows are created here.
	_ = message
	return nil
}

// DisableNotificationsForPaidInstallment disables notifications when an installment is paid
func (s *NotificationService) DisableNotificationsForPaidInstallment(debtListID uuid.UUID, installmentNumber int) error {
	s.logger.Info().
		Str("debt_list_id", debtListID.String()).
		Int("installment_number", installmentNumber).
		Msg("Disabling notifications for paid installment")

	return s.notificationRepo.DisableNotificationsForInstallment(debtListID, installmentNumber)
}

// getUserSettings gets user settings or returns defaults
func (s *NotificationService) getUserSettings(userID uuid.UUID) (*models.UserSettings, error) {
	ctx := context.Background()
	
	// GetOrCreate returns existing settings or creates default ones
	settings, err := s.userSettingsRepo.GetOrCreate(ctx, userID)
	if err != nil {
		s.logger.Error().Err(err).Str("userID", userID.String()).Msg("Failed to get user settings, using defaults")
		// Return default settings on error
		return &models.UserSettings{
			UserID:                   userID,
			NotificationEmail:        true,
			NotificationSMS:          false,
			NotificationWebhook:      false,
			NotificationReminderDays: []int64{7, 3, 1},
			NotificationTime:         "09:00:00",
			NotificationRecipient:    "both",
		}, nil
	}
	
	return settings, nil
}

// entityToModelDebtList converts entities.DebtList to models.DebtList
func (s *NotificationService) entityToModelDebtList(entity *entities.DebtList) *models.DebtList {
	return &models.DebtList{
		ID:                  entity.ID,
		UserID:              entity.UserID,
		ContactID:           entity.ContactID,
		DebtType:            entity.DebtType,
		TotalAmount:         entity.TotalAmount,
		InstallmentAmount:   entity.InstallmentAmount,
		TotalPaymentsMade:   entity.TotalPaymentsMade,
		TotalRemainingDebt:  entity.TotalRemainingDebt,
		Currency:            entity.Currency,
		Status:              entity.Status,
		DueDate:             entity.DueDate,
		NextPaymentDate:     entity.NextPaymentDate,
		InstallmentPlan:     entity.InstallmentPlan,
		NumberOfPayments:    entity.NumberOfPayments,
		Description:         entity.Description,
		Notes:               entity.Notes,
		CreatedAt:           entity.CreatedAt,
		UpdatedAt:           entity.UpdatedAt,
	}
}

