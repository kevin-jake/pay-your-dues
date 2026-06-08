package interfaces

import (
	"time"

	"pay-your-dues/internal/models"

	"github.com/google/uuid"
)

// NotificationService defines the interface for notification business logic
type NotificationService interface {
	// Notification creation
	CreateNotification(userID uuid.UUID, req *models.CreateNotificationRequest) (*models.Notification, error)
	CreateNotificationsForDebtList(userID uuid.UUID, debtListID uuid.UUID) error
	CreateNotificationsForInstallment(debtListID uuid.UUID, installmentNumber int, installmentDueDate time.Time) error

	// Notification retrieval
	GetNotification(userID uuid.UUID, notificationID uuid.UUID) (*models.Notification, error)
	GetUserNotifications(userID uuid.UUID, status string, debtListID *uuid.UUID, limit int) ([]*models.Notification, error)
	GetNotificationsByDebtList(userID uuid.UUID, debtListID uuid.UUID) ([]*models.Notification, error)
	GetNotificationsByInstallment(userID uuid.UUID, debtListID uuid.UUID, installmentNumber int) ([]*models.Notification, error)

	// Notification management
	UpdateNotification(userID uuid.UUID, notificationID uuid.UUID, enabled bool) error
	DeleteNotification(userID uuid.UUID, notificationID uuid.UUID) error
	EnableNotification(userID uuid.UUID, notificationID uuid.UUID) error
	DisableNotification(userID uuid.UUID, notificationID uuid.UUID) error

	// Notification sending
	SendNotification(notificationID uuid.UUID) error
	SendManualNotification(userID uuid.UUID, debtListID uuid.UUID, message string, notificationType string) error
	ProcessPendingNotifications() error

	// Event-based notifications
	SendPaymentConfirmationNotifications(debtItemID uuid.UUID) error
	SendPaymentVerificationNotification(debtItemID uuid.UUID, verified bool, reason string) error

	// Installment management
	DisableNotificationsForPaidInstallment(debtListID uuid.UUID, installmentNumber int) error

	// Debt-level schedule management
	GetDebtNotificationSettings(userID uuid.UUID, debtListID uuid.UUID) (*models.DebtNotificationSettingsResponse, error)
	EnableDebtNotifications(userID uuid.UUID, debtListID uuid.UUID, settings *models.CustomNotificationSettings) error
	DisableDebtNotifications(userID uuid.UUID, debtListID uuid.UUID) error
	DeleteNotificationSlot(userID uuid.UUID, debtListID uuid.UUID, installmentNumber *int, scheduledFor time.Time) error
	GetManualSendLimits(userID uuid.UUID, debtListID uuid.UUID) (*models.ManualSendLimits, error)
}

