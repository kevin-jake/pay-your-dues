package services

import (
	"context"
	"fmt"
	"time"

	"pay-your-dues/internal/config"
	"pay-your-dues/internal/domain/entities"
	"pay-your-dues/internal/domain/interfaces"
	"pay-your-dues/internal/models"
	"pay-your-dues/internal/services/notification"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// NotificationService implements the NotificationService interface
type NotificationService struct {
	notificationRepo interfaces.NotificationRepository
	templateRepo     interfaces.NotificationTemplateRepository
	debtListRepo     interfaces.DebtListRepository
	debtItemRepo     interfaces.DebtItemRepository
	contactRepo      interfaces.ContactRepository
	userRepo         interfaces.UserRepository

	emailSender     *notification.EmailSender
	smsSender       *notification.SMSSender
	webhookService  *notification.WebhookService
	templateEngine  *notification.TemplateEngine
	contactFetcher  *notification.ContactFetcher
	paymentChecker  *notification.PaymentChecker

	config *config.NotificationConfig
	logger zerolog.Logger
}

// NewNotificationService creates a new notification service instance
func NewNotificationService(
	notificationRepo interfaces.NotificationRepository,
	templateRepo interfaces.NotificationTemplateRepository,
	debtListRepo interfaces.DebtListRepository,
	debtItemRepo interfaces.DebtItemRepository,
	contactRepo interfaces.ContactRepository,
	userRepo interfaces.UserRepository,
	db *gorm.DB,
	cfg *config.NotificationConfig,
	logger zerolog.Logger,
) interfaces.NotificationService {
	emailSender := notification.NewEmailSender(cfg)
	smsSender := notification.NewSMSSender(cfg)
	webhookService := notification.NewWebhookService()
	templateEngine := notification.NewTemplateEngine()
	
	contactFetcher := notification.NewContactFetcher(
		db,
		debtListRepo,
		contactRepo,
		userRepo,
	)
	
	paymentChecker := notification.NewPaymentChecker(
		db,
		debtListRepo,
		debtItemRepo,
	)

	return &NotificationService{
		notificationRepo: notificationRepo,
		templateRepo:     templateRepo,
		debtListRepo:     debtListRepo,
		debtItemRepo:     debtItemRepo,
		contactRepo:      contactRepo,
		userRepo:         userRepo,
		emailSender:      emailSender,
		smsSender:        smsSender,
		webhookService:   webhookService,
		templateEngine:   templateEngine,
		contactFetcher:   contactFetcher,
		paymentChecker:   paymentChecker,
		config:           cfg,
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
	if debtList.DebtType == "onetime" {
		return s.createOneTimeNotifications(s.entityToModelDebtList(debtList), userSettings)
	}

	// For installment debts, create notifications for each installment
	return s.createInstallmentNotifications(s.entityToModelDebtList(debtList), userSettings)
}

// createOneTimeNotifications creates notifications for one-time debts
func (s *NotificationService) createOneTimeNotifications(debtList *models.DebtList, settings *models.UserSettings) error {
	reminderDays := []int64{7, 3, 1} // Default
	if settings.NotificationReminderDays != nil && len(settings.NotificationReminderDays) > 0 {
		reminderDays = settings.NotificationReminderDays
	}

	notificationTime := "09:00:00"
	if settings.NotificationTime != "" {
		notificationTime = settings.NotificationTime
	}

	var notifications []*models.Notification

	for _, daysBefore := range reminderDays {
		scheduledFor := debtList.DueDate.AddDate(0, 0, -int(daysBefore))
		
		// Parse notification time
		t, _ := time.Parse("15:04:05", notificationTime)
		scheduledFor = time.Date(
			scheduledFor.Year(), scheduledFor.Month(), scheduledFor.Day(),
			t.Hour(), t.Minute(), t.Second(), 0, scheduledFor.Location(),
		)

		// Only create if scheduled time is in the future
		if scheduledFor.After(time.Now()) {
			daysBefore32 := int(daysBefore)
			notification := &models.Notification{
				ID:                 uuid.New(),
				DebtListID:         debtList.ID,
				NotificationType:   "email",
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

	if len(notifications) > 0 {
		return s.notificationRepo.CreateBatch(notifications)
	}

	return nil
}

// createInstallmentNotifications creates notifications for installment debts
func (s *NotificationService) createInstallmentNotifications(debtList *models.DebtList, settings *models.UserSettings) error {
	if debtList.NumberOfPayments == nil || *debtList.NumberOfPayments == 0 {
		return fmt.Errorf("invalid number of payments for installment debt")
	}

	reminderDays := []int64{7, 3, 1} // Default
	if settings.NotificationReminderDays != nil && len(settings.NotificationReminderDays) > 0 {
		reminderDays = settings.NotificationReminderDays
	}

	notificationTime := "09:00:00"
	if settings.NotificationTime != "" {
		notificationTime = settings.NotificationTime
	}

	var notifications []*models.Notification
	currentDueDate := debtList.DueDate

	for i := 1; i <= *debtList.NumberOfPayments; i++ {
		installmentNumber := i

		for _, daysBefore := range reminderDays {
			scheduledFor := currentDueDate.AddDate(0, 0, -int(daysBefore))
			
			// Parse notification time
			t, _ := time.Parse("15:04:05", notificationTime)
			scheduledFor = time.Date(
				scheduledFor.Year(), scheduledFor.Month(), scheduledFor.Day(),
				t.Hour(), t.Minute(), t.Second(), 0, scheduledFor.Location(),
			)

			// Only create if scheduled time is in the future
			if scheduledFor.After(time.Now()) {
				daysBefore32 := int(daysBefore)
				notification := &models.Notification{
					ID:                 uuid.New(),
					DebtListID:         debtList.ID,
					InstallmentNumber:  &installmentNumber,
					InstallmentDueDate: &currentDueDate,
					NotificationType:   "email",
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

		// Calculate next due date based on installment plan
		currentDueDate = s.calculateNextDueDate(currentDueDate, debtList.InstallmentPlan)
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
	if userSettings.NotificationReminderDays != nil && len(userSettings.NotificationReminderDays) > 0 {
		reminderDays = userSettings.NotificationReminderDays
	}

	var notifications []*models.Notification

	for _, daysBefore := range reminderDays {
		scheduledFor := installmentDueDate.AddDate(0, 0, -int(daysBefore))
		
		if scheduledFor.After(time.Now()) {
			daysBefore32 := int(daysBefore)
			notification := &models.Notification{
				ID:                 uuid.New(),
				DebtListID:         debtListID,
				InstallmentNumber:  &installmentNumber,
				InstallmentDueDate: &installmentDueDate,
				NotificationType:   "email",
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

// SendNotification sends a specific notification
func (s *NotificationService) SendNotification(notificationID uuid.UUID) error {
	notification, err := s.notificationRepo.GetByID(notificationID)
	if err != nil {
		return fmt.Errorf("notification not found: %w", err)
	}

	// Check if notification should be sent
	// TODO: Implement payment status checking here

	// Send based on notification type
	switch notification.NotificationType {
	case "email":
		// TODO: Implement email sending
		s.logger.Info().Str("notification_id", notificationID.String()).Msg("Sending email notification")
	case "sms":
		// TODO: Implement SMS sending
		s.logger.Info().Str("notification_id", notificationID.String()).Msg("Sending SMS notification")
	case "webhook":
		// TODO: Implement webhook sending
		s.logger.Info().Str("notification_id", notificationID.String()).Msg("Sending webhook notification")
	}

	// Update status
	now := time.Now()
	return s.notificationRepo.UpdateStatus(notificationID, "sent", &now)
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

	// Create and send notification immediately
	notification := &models.Notification{
		ID:               uuid.New(),
		DebtListID:       debtListID,
		NotificationType: notificationType,
		RecipientType:    "contact",
		Message:          message,
		Status:           "pending",
		ScheduleType:     "manual",
		Enabled:          true,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.notificationRepo.Create(notification); err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	return s.SendNotification(notification.ID)
}

// ProcessPendingNotifications processes all pending notifications
func (s *NotificationService) ProcessPendingNotifications() error {
	now := time.Now()
	notifications, err := s.notificationRepo.GetScheduledNotifications(now, 100)
	if err != nil {
		return fmt.Errorf("failed to get pending notifications: %w", err)
	}

	s.logger.Info().Int("count", len(notifications)).Msg("Processing pending notifications")

	for _, notification := range notifications {
		if err := s.SendNotification(notification.ID); err != nil {
			s.logger.Error().
				Err(err).
				Str("notification_id", notification.ID.String()).
				Msg("Failed to send notification")
		}
	}

	return nil
}

// SendPaymentConfirmationNotifications sends notifications when a payment is made
func (s *NotificationService) SendPaymentConfirmationNotifications(debtItemID uuid.UUID) error {
	// TODO: Implement payment confirmation notifications
	s.logger.Info().Str("debt_item_id", debtItemID.String()).Msg("Sending payment confirmation notifications")
	return nil
}

// SendPaymentVerificationNotification sends a notification when payment is verified/rejected
func (s *NotificationService) SendPaymentVerificationNotification(debtItemID uuid.UUID, verified bool, reason string) error {
	// TODO: Implement payment verification notifications
	s.logger.Info().
		Str("debt_item_id", debtItemID.String()).
		Bool("verified", verified).
		Msg("Sending payment verification notification")
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
	// Try to get from database
	// For now, return default settings
	return &models.UserSettings{
		UserID:                   userID,
		NotificationEmail:        true,
		NotificationSMS:          false,
		NotificationWebhook:      false,
		NotificationReminderDays: []int64{7, 3, 1},
		NotificationTime:         "09:00:00",
	}, nil
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

