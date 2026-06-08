package services

import (
	"context"
	"fmt"
	"time"

	"pay-your-dues/internal/config"
	"pay-your-dues/internal/domain/interfaces"
	"pay-your-dues/internal/messaging"
	"pay-your-dues/internal/models"
	notifpkg "pay-your-dues/internal/services/notification"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// NotificationDeliveryService delivers notifications consumed from RabbitMQ.
type NotificationDeliveryService struct {
	notificationRepo interfaces.NotificationRepository
	debtListRepo     interfaces.DebtListRepository
	contactRepo      interfaces.ContactRepository
	userRepo         interfaces.UserRepository

	emailSender    *notifpkg.EmailSender
	smsSender      *notifpkg.SMSSender
	webhookService *notifpkg.WebhookService
	templateEngine *notifpkg.TemplateEngine
	contactFetcher *notifpkg.ContactFetcher
	paymentChecker *notifpkg.PaymentChecker

	db     *gorm.DB
	config *config.NotificationConfig
	logger zerolog.Logger

	settingsCache map[uuid.UUID]*models.UserSettings
	contactCache  map[uuid.UUID]*notifpkg.ContactInfo
}

func NewNotificationDeliveryService(
	notificationRepo interfaces.NotificationRepository,
	debtListRepo interfaces.DebtListRepository,
	debtItemRepo interfaces.DebtItemRepository,
	contactRepo interfaces.ContactRepository,
	userRepo interfaces.UserRepository,
	db *gorm.DB,
	cfg *config.NotificationConfig,
	logger zerolog.Logger,
) *NotificationDeliveryService {
	return &NotificationDeliveryService{
		notificationRepo: notificationRepo,
		debtListRepo:     debtListRepo,
		contactRepo:      contactRepo,
		userRepo:         userRepo,
		emailSender:      notifpkg.NewEmailSender(cfg),
		smsSender:        notifpkg.NewSMSSender(cfg),
		webhookService:   notifpkg.NewWebhookService(cfg.TelegramBotToken),
		templateEngine:   notifpkg.NewTemplateEngine(),
		contactFetcher:   notifpkg.NewContactFetcher(db, debtListRepo, contactRepo, userRepo),
		paymentChecker:   notifpkg.NewPaymentChecker(db, debtListRepo, debtItemRepo),
		db:               db,
		config:           cfg,
		logger:           logger.With().Str("service", "notification_delivery").Logger(),
		settingsCache:    make(map[uuid.UUID]*models.UserSettings),
		contactCache:     make(map[uuid.UUID]*notifpkg.ContactInfo),
	}
}

func (s *NotificationDeliveryService) ClearCaches() {
	s.settingsCache = make(map[uuid.UUID]*models.UserSettings)
	s.contactCache = make(map[uuid.UUID]*notifpkg.ContactInfo)
}

// ProcessJob handles a notification job from RabbitMQ.
func (s *NotificationDeliveryService) ProcessJob(ctx context.Context, job messaging.NotificationJob) error {
	notif, err := s.notificationRepo.GetByID(job.NotificationID)
	if err != nil {
		return fmt.Errorf("notification not found: %w", err)
	}

	switch notif.Status {
	case "sent", "skipped", "cancelled":
		s.logger.Debug().
			Str("notification_id", notif.ID.String()).
			Str("status", notif.Status).
			Msg("Notification already terminal, skipping")
		return nil
	}

	if !notif.Enabled {
		now := time.Now()
		return s.notificationRepo.UpdateStatus(notif.ID, "skipped", &now)
	}

	shouldSend, err := s.paymentChecker.ShouldSendNotificationForInstallment(notif.DebtListID, notif.InstallmentNumber)
	if err != nil {
		return fmt.Errorf("payment check failed: %w", err)
	}
	if !shouldSend {
		now := time.Now()
		if err := s.notificationRepo.UpdateStatus(notif.ID, "skipped", &now); err != nil {
			return err
		}
		s.logger.Info().
			Str("notification_id", notif.ID.String()).
			Msg("Skipped notification because installment is already paid")
		return nil
	}

	if err := s.deliverNotification(ctx, notif); err != nil {
		if job.Attempt >= s.config.MaxRetry {
			now := time.Now()
			if updateErr := s.notificationRepo.UpdateStatus(notif.ID, "failed", &now); updateErr != nil {
				return fmt.Errorf("mark failed: %w (original: %v)", updateErr, err)
			}
			return nil
		}
		return err
	}

	now := time.Now()
	if err := s.notificationRepo.UpdateStatus(notif.ID, "sent", &now); err != nil {
		return fmt.Errorf("failed to update notification status: %w", err)
	}

	s.logger.Info().
		Str("notification_id", notif.ID.String()).
		Str("type", notif.NotificationType).
		Msg("Notification sent successfully")

	return nil
}

func (s *NotificationDeliveryService) deliverNotification(ctx context.Context, notif *models.Notification) error {
	debtList, err := s.debtListRepo.GetByID(ctx, notif.DebtListID)
	if err != nil {
		return fmt.Errorf("failed to get debt list: %w", err)
	}

	contactInfo, err := s.getContactInfo(notif.DebtListID)
	if err != nil {
		return fmt.Errorf("failed to get contact info: %w", err)
	}

	userSettings, err := s.getUserSettings(debtList.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user settings: %w", err)
	}

	if !shouldNotifyRecipient(userSettings, notif.RecipientType) {
		now := time.Now()
		if err := s.notificationRepo.UpdateStatus(notif.ID, "skipped", &now); err != nil {
			return err
		}
		return nil
	}

	daysUntilDue := 0
	if notif.InstallmentDueDate != nil {
		daysUntilDue = int(time.Until(*notif.InstallmentDueDate).Hours() / 24)
	} else {
		daysUntilDue = int(time.Until(debtList.DueDate).Hours() / 24)
	}

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

	contactEmail := ""
	if contactInfo.Email != nil {
		contactEmail = *contactInfo.Email
	}
	contactPhone := ""
	if contactInfo.Phone != nil {
		contactPhone = *contactInfo.Phone
	}

	notifData := notifpkg.TemplateData{
		RecipientType:         notif.RecipientType,
		UserFirstName:         contactInfo.UserFirstName,
		UserLastName:          contactInfo.UserLastName,
		UserEmail:             contactInfo.UserEmail,
		ContactName:           contactInfo.ContactName,
		ContactEmail:          contactEmail,
		ContactPhone:          contactPhone,
		Amount:                debtList.TotalAmount,
		Currency:              debtList.Currency,
		DueDate:               debtList.DueDate,
		TotalDebt:             debtList.TotalAmount,
		RemainingDebt:         debtList.TotalRemainingDebt,
		PaidAmount:            debtList.TotalPaymentsMade,
		DebtType:              debtList.DebtType,
		InstallmentPlan:       debtList.InstallmentPlan,
		InstallmentNumber:     installmentNum,
		InstallmentTotal:      installmentTotal,
		InstallmentDueDate:    installmentDueDate,
		InstallmentAmount:     debtList.InstallmentAmount,
		RemainingInstallments: remainingInstallments,
		DaysUntilDue:          daysUntilDue,
	}

	switch notif.NotificationType {
	case "email":
		return s.sendEmailNotification(notif, notifData, contactInfo, userSettings)
	case "sms":
		return s.sendSMSNotification(notif, notifData, contactInfo, userSettings)
	case "webhook":
		return s.sendWebhookNotification(notif, notifData, userSettings)
	case "slack":
		return s.sendSlackNotification(notifData, userSettings)
	case "telegram":
		return s.sendTelegramNotification(notifData, userSettings)
	case "discord":
		return s.sendDiscordNotification(notifData, userSettings)
	default:
		return fmt.Errorf("unsupported notification type: %s", notif.NotificationType)
	}
}

func (s *NotificationDeliveryService) getContactInfo(debtListID uuid.UUID) (*notifpkg.ContactInfo, error) {
	if cached, ok := s.contactCache[debtListID]; ok {
		return cached, nil
	}
	info, err := s.contactFetcher.GetContactInfoForDebtList(debtListID)
	if err != nil {
		return nil, err
	}
	s.contactCache[debtListID] = info
	return info, nil
}

func (s *NotificationDeliveryService) getUserSettings(userID uuid.UUID) (*models.UserSettings, error) {
	if cached, ok := s.settingsCache[userID]; ok {
		return cached, nil
	}

	var settings models.UserSettings
	err := s.db.Where("user_id = ?", userID).First(&settings).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			defaults := &models.UserSettings{
				UserID:                    userID,
				NotificationEmail:         true,
				NotificationSMS:           false,
				NotificationWebhook:       false,
				NotificationReminderDays:  []int64{7, 3, 1},
				NotificationTime:          "09:00:00",
				OverdueReminderFrequency:  "daily",
				EventNotificationsEnabled: true,
				NotifyContactOnPayment:    true,
				NotificationRecipient:     "both",
				DefaultCurrency:           "Php",
				Timezone:                  "UTC",
			}
			s.settingsCache[userID] = defaults
			return defaults, nil
		}
		return nil, fmt.Errorf("failed to fetch user settings: %w", err)
	}

	s.settingsCache[userID] = &settings
	return &settings, nil
}

func shouldNotifyRecipient(settings *models.UserSettings, recipientType string) bool {
	recipientSetting := settings.NotificationRecipient
	if recipientSetting == "" {
		recipientSetting = "both"
	}

	switch recipientSetting {
	case "both":
		return true
	case "user":
		return recipientType == "user"
	case "contact":
		return recipientType == "contact"
	default:
		return true
	}
}

func (s *NotificationDeliveryService) sendEmailNotification(
	notif *models.Notification,
	data notifpkg.TemplateData,
	contactInfo *notifpkg.ContactInfo,
	userSettings *models.UserSettings,
) error {
	recipientEmail := notif.RecipientEmail
	if recipientEmail == nil || *recipientEmail == "" {
		recipientEmail = contactInfo.Email
	}
	if recipientEmail == nil || *recipientEmail == "" {
		return fmt.Errorf("no email address available for notification")
	}

	var body string
	if userSettings.CustomEmailMessage != nil && *userSettings.CustomEmailMessage != "" {
		body = s.templateEngine.RenderHTML(*userSettings.CustomEmailMessage, data)
	} else {
		body = s.templateEngine.RenderHTML(notifpkg.GetDefaultEmailTemplate(), data)
	}

	subject := "Payment Reminder"
	if notif.InstallmentNumber != nil {
		subject = fmt.Sprintf("Payment Reminder - Installment #%d", *notif.InstallmentNumber)
	}

	return s.emailSender.SendEmail(*recipientEmail, subject, body)
}

func (s *NotificationDeliveryService) sendSMSNotification(
	notif *models.Notification,
	data notifpkg.TemplateData,
	contactInfo *notifpkg.ContactInfo,
	userSettings *models.UserSettings,
) error {
	recipientPhone := notif.RecipientPhone
	if recipientPhone == nil || *recipientPhone == "" {
		recipientPhone = contactInfo.Phone
	}
	if recipientPhone == nil || *recipientPhone == "" {
		return fmt.Errorf("no phone number available for notification")
	}

	var body string
	if userSettings.CustomSMSMessage != nil && *userSettings.CustomSMSMessage != "" {
		body = s.templateEngine.Render(*userSettings.CustomSMSMessage, data)
	} else {
		body = s.templateEngine.Render(notifpkg.GetDefaultSMSTemplate(), data)
	}

	return s.smsSender.SendSMS(*recipientPhone, body)
}

func (s *NotificationDeliveryService) sendWebhookNotification(
	notif *models.Notification,
	data notifpkg.TemplateData,
	userSettings *models.UserSettings,
) error {
	if notif.WebhookType == nil || *notif.WebhookType == "" {
		return fmt.Errorf("webhook type not specified")
	}
	if !s.webhookService.IsWebhookConfigured(*notif.WebhookType, userSettings) {
		return fmt.Errorf("%s webhook not configured", *notif.WebhookType)
	}
	return s.webhookService.SendNotification(*notif.WebhookType, userSettings, data)
}

func (s *NotificationDeliveryService) sendSlackNotification(data notifpkg.TemplateData, userSettings *models.UserSettings) error {
	if !s.webhookService.IsWebhookConfigured("slack", userSettings) {
		return fmt.Errorf("slack webhook not configured")
	}
	return s.webhookService.SendNotification("slack", userSettings, data)
}

func (s *NotificationDeliveryService) sendTelegramNotification(data notifpkg.TemplateData, userSettings *models.UserSettings) error {
	if !s.webhookService.IsWebhookConfigured("telegram", userSettings) {
		return fmt.Errorf("telegram not linked")
	}
	return s.webhookService.SendNotification("telegram", userSettings, data)
}

func (s *NotificationDeliveryService) sendDiscordNotification(data notifpkg.TemplateData, userSettings *models.UserSettings) error {
	if !s.webhookService.IsWebhookConfigured("discord", userSettings) {
		return fmt.Errorf("discord webhook not configured")
	}
	return s.webhookService.SendNotification("discord", userSettings, data)
}
