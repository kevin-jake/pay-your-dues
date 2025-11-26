package workers

import (
	"context"
	"fmt"
	"time"

	"pay-your-dues/internal/config"
	"pay-your-dues/internal/domain/interfaces"
	"pay-your-dues/internal/models"
	"pay-your-dues/internal/services/notification"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// NotificationWorker processes and sends pending notifications
type NotificationWorker struct {
	notificationRepo interfaces.NotificationRepository
	debtListRepo     interfaces.DebtListRepository
	contactRepo      interfaces.ContactRepository
	userRepo         interfaces.UserRepository

	emailSender    *notification.EmailSender
	smsSender      *notification.SMSSender
	webhookService *notification.WebhookService
	templateEngine *notification.TemplateEngine
	contactFetcher *notification.ContactFetcher

	config *config.NotificationConfig
	logger zerolog.Logger

	stopChan chan struct{}
	running  bool
}

// NewNotificationWorker creates a new notification worker
func NewNotificationWorker(
	notificationRepo interfaces.NotificationRepository,
	debtListRepo interfaces.DebtListRepository,
	contactRepo interfaces.ContactRepository,
	userRepo interfaces.UserRepository,
	db *gorm.DB,
	cfg *config.NotificationConfig,
	logger zerolog.Logger,
) *NotificationWorker {
	emailSender := notification.NewEmailSender(cfg)
	smsSender := notification.NewSMSSender(cfg)
	webhookService := notification.NewWebhookService()
	templateEngine := notification.NewTemplateEngine()
	contactFetcher := notification.NewContactFetcher(db, debtListRepo, contactRepo, userRepo)

	return &NotificationWorker{
		notificationRepo: notificationRepo,
		debtListRepo:     debtListRepo,
		contactRepo:      contactRepo,
		userRepo:         userRepo,
		emailSender:      emailSender,
		smsSender:        smsSender,
		webhookService:   webhookService,
		templateEngine:   templateEngine,
		contactFetcher:   contactFetcher,
		config:           cfg,
		logger:           logger,
		stopChan:         make(chan struct{}),
		running:          false,
	}
}

// Start begins processing notifications
func (w *NotificationWorker) Start() {
	if w.running {
		w.logger.Warn().Msg("Notification worker is already running")
		return
	}

	w.running = true
	w.logger.Info().Msg("Starting notification worker")

	go w.processLoop()
}

// Stop stops the notification worker
func (w *NotificationWorker) Stop() {
	if !w.running {
		return
	}

	w.logger.Info().Msg("Stopping notification worker")
	close(w.stopChan)
	w.running = false
}

// IsRunning returns whether the worker is currently running
func (w *NotificationWorker) IsRunning() bool {
	return w.running
}

// processLoop is the main processing loop
func (w *NotificationWorker) processLoop() {
	ticker := time.NewTicker(w.config.WorkerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopChan:
			w.logger.Info().Msg("Notification worker stopped")
			return
		case <-ticker.C:
			w.processPendingNotifications()
		}
	}
}

// processPendingNotifications processes all pending notifications
func (w *NotificationWorker) processPendingNotifications() {
	ctx := context.Background()

	// Get pending notifications that are scheduled to be sent
	now := time.Now()
	notifications, err := w.notificationRepo.GetScheduledNotifications(now, w.config.BatchSize)
	if err != nil {
		w.logger.Error().Err(err).Msg("Failed to get pending notifications")
		return
	}

	if len(notifications) == 0 {
		return
	}

	w.logger.Info().
		Int("count", len(notifications)).
		Msg("Processing pending notifications")

	for i := range notifications {
		if err := w.processNotification(ctx, notifications[i]); err != nil {
			w.logger.Error().
				Err(err).
				Str("notification_id", notifications[i].ID.String()).
				Msg("Failed to process notification")
		}
	}
}

// processNotification processes a single notification
func (w *NotificationWorker) processNotification(ctx context.Context, notif *models.Notification) error {
	// Get debt list information
	debtList, err := w.debtListRepo.GetByID(ctx, notif.DebtListID)
	if err != nil {
		return fmt.Errorf("failed to get debt list: %w", err)
	}

	// Get contact information
	contactInfo, err := w.contactFetcher.GetContactInfoForDebtList(notif.DebtListID)
	if err != nil {
		return fmt.Errorf("failed to get contact info: %w", err)
	}

	// Get user settings for notification preferences
	userSettings, err := w.getUserSettings(debtList.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user settings: %w", err)
	}

	// Calculate days until due
	daysUntilDue := 0
	if notif.InstallmentDueDate != nil {
		daysUntilDue = int(time.Until(*notif.InstallmentDueDate).Hours() / 24)
	} else {
		daysUntilDue = int(time.Until(debtList.DueDate).Hours() / 24)
	}

	// Prepare template data
	installmentNum := 0
	if notif.InstallmentNumber != nil {
		installmentNum = *notif.InstallmentNumber
	}
	
	installmentTotal := 0
	if debtList.NumberOfPayments != nil {
		installmentTotal = *debtList.NumberOfPayments
	}

	remainingInstallments := 0
	if notif.InstallmentNumber != nil && debtList.NumberOfPayments != nil {
		remainingInstallments = *debtList.NumberOfPayments - *notif.InstallmentNumber
	}

	installmentDueDate := time.Now()
	if notif.InstallmentDueDate != nil {
		installmentDueDate = *notif.InstallmentDueDate
	}

	notifData := notification.TemplateData{
		UserFirstName:          contactInfo.UserFirstName,
		UserLastName:           contactInfo.UserLastName,
		UserEmail:              contactInfo.UserEmail,
		ContactName:            contactInfo.ContactName,
		ContactEmail:           *contactInfo.Email,
		ContactPhone:           *contactInfo.Phone,
		Amount:                 debtList.TotalAmount,
		Currency:               debtList.Currency,
		DueDate:                debtList.DueDate,
		TotalDebt:              debtList.TotalAmount,
		RemainingDebt:          debtList.TotalRemainingDebt,
		PaidAmount:             debtList.TotalPaymentsMade,
		DebtType:               debtList.DebtType,
		InstallmentPlan:        debtList.InstallmentPlan,
		InstallmentNumber:      installmentNum,
		InstallmentTotal:       installmentTotal,
		InstallmentDueDate:     installmentDueDate,
		InstallmentAmount:      debtList.InstallmentAmount,
		RemainingInstallments:  remainingInstallments,
		DaysUntilDue:           daysUntilDue,
	}

	// Send notification based on type
	var sendErr error
	switch notif.NotificationType {
	case "email":
		sendErr = w.sendEmailNotification(notif, notifData, contactInfo, userSettings)
	case "sms":
		sendErr = w.sendSMSNotification(notif, notifData, contactInfo, userSettings)
	case "webhook":
		sendErr = w.sendWebhookNotification(notif, notifData, userSettings)
	default:
		sendErr = fmt.Errorf("unsupported notification type: %s", notif.NotificationType)
	}

	// Update notification status
	now := time.Now()
	if sendErr != nil {
		// Mark as failed and retry later if not exceeded max retries
		// For now, just mark as failed
		if err := w.notificationRepo.UpdateStatus(notif.ID, "failed", &now); err != nil {
			w.logger.Error().Err(err).Str("notification_id", notif.ID.String()).Msg("Failed to update notification status")
		}
		return sendErr
	}

	// Mark as sent
	if err := w.notificationRepo.UpdateStatus(notif.ID, "sent", &now); err != nil {
		return fmt.Errorf("failed to update notification status: %w", err)
	}

	w.logger.Info().
		Str("notification_id", notif.ID.String()).
		Str("type", notif.NotificationType).
		Msg("Notification sent successfully")

	return nil
}

// sendEmailNotification sends an email notification
func (w *NotificationWorker) sendEmailNotification(
	notif *models.Notification,
	data notification.TemplateData,
	contactInfo *notification.ContactInfo,
	userSettings *models.UserSettings,
) error {
	// Determine recipient email
	recipientEmail := notif.RecipientEmail
	if recipientEmail == nil || *recipientEmail == "" {
		if notif.RecipientType == "user" {
			// Send to user's email (from user_contacts)
			recipientEmail = contactInfo.Email
		} else {
			// Send to contact's email
			recipientEmail = contactInfo.Email
		}
	}

	if recipientEmail == nil || *recipientEmail == "" {
		return fmt.Errorf("no email address available for notification")
	}

	// Render email template
	var body string
	if userSettings.CustomEmailMessage != nil && *userSettings.CustomEmailMessage != "" {
		body = w.templateEngine.RenderHTML(*userSettings.CustomEmailMessage, data)
	} else {
		// Use default template (embedded in template engine)
		defaultTemplate := notification.GetDefaultEmailTemplate()
		body = w.templateEngine.RenderHTML(defaultTemplate, data)
	}

	// Send email
	subject := "Payment Reminder"
	if notif.InstallmentNumber != nil {
		subject = fmt.Sprintf("Payment Reminder - Installment #%d", *notif.InstallmentNumber)
	}

	return w.emailSender.SendEmail(*recipientEmail, subject, body)
}

// sendSMSNotification sends an SMS notification
func (w *NotificationWorker) sendSMSNotification(
	notif *models.Notification,
	data notification.TemplateData,
	contactInfo *notification.ContactInfo,
	userSettings *models.UserSettings,
) error {
	// Determine recipient phone
	recipientPhone := notif.RecipientPhone
	if recipientPhone == nil || *recipientPhone == "" {
		if notif.RecipientType == "user" {
			// Send to user's phone (from user_contacts)
			recipientPhone = contactInfo.Phone
		} else {
			// Send to contact's phone
			recipientPhone = contactInfo.Phone
		}
	}

	if recipientPhone == nil || *recipientPhone == "" {
		return fmt.Errorf("no phone number available for notification")
	}

	// Render SMS template
	var body string
	if userSettings.CustomSMSMessage != nil && *userSettings.CustomSMSMessage != "" {
		body = w.templateEngine.Render(*userSettings.CustomSMSMessage, data)
	} else {
		// Use default template (embedded in template engine)
		defaultTemplate := notification.GetDefaultSMSTemplate()
		body = w.templateEngine.Render(defaultTemplate, data)
	}

	// Send SMS
	return w.smsSender.SendSMS(*recipientPhone, body)
}

// sendWebhookNotification sends a webhook notification
func (w *NotificationWorker) sendWebhookNotification(
	notif *models.Notification,
	data notification.TemplateData,
	userSettings *models.UserSettings,
) error {
	if notif.WebhookType == nil || *notif.WebhookType == "" {
		return fmt.Errorf("webhook type not specified")
	}

	return w.webhookService.SendNotification(*notif.WebhookType, userSettings, data)
}

// getUserSettings gets user settings or returns defaults
func (w *NotificationWorker) getUserSettings(userID uuid.UUID) (*models.UserSettings, error) {
	// TODO: Get from database
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

