package interfaces

import (
	"time"

	"pay-your-dues/internal/models"

	"github.com/google/uuid"
)

// NotificationRepository defines the interface for notification data access
type NotificationRepository interface {
	// Basic CRUD operations
	Create(notification *models.Notification) error
	GetByID(id uuid.UUID) (*models.Notification, error)
	Update(notification *models.Notification) error
	Delete(id uuid.UUID) error

	// Query operations
	GetByDebtListID(debtListID uuid.UUID) ([]*models.Notification, error)
	GetByDebtListAndInstallment(debtListID uuid.UUID, installmentNumber int) ([]*models.Notification, error)
	GetPendingNotifications(limit int) ([]*models.Notification, error)
	GetScheduledNotifications(beforeTime time.Time, limit int) ([]*models.Notification, error)
	GetByStatus(status string, limit int) ([]*models.Notification, error)

	// Batch operations
	CreateBatch(notifications []*models.Notification) error
	UpdateStatus(id uuid.UUID, status string, sentAt *time.Time) error
	UpdateNextRun(id uuid.UUID, nextRunAt time.Time) error
	DisableNotification(id uuid.UUID) error
	EnableNotification(id uuid.UUID) error

	// Installment-specific operations
	DisableNotificationsForInstallment(debtListID uuid.UUID, installmentNumber int) error
	GetActiveNotificationsForInstallment(debtListID uuid.UUID, installmentNumber int) ([]*models.Notification, error)

	// Status checking
	HasPendingNotifications(debtListID uuid.UUID, installmentNumber *int) (bool, error)
	CountNotificationsByDebtList(debtListID uuid.UUID) (int64, error)

	// User-specific operations
	GetPendingNotificationsByUserID(userID uuid.UUID) ([]*models.Notification, error)
	UpdatePendingNotificationsByUserID(userID uuid.UUID, updates map[string]interface{}) error

	// Queue scheduler operations
	ClaimDueNotifications(beforeTime time.Time, limit int) ([]*models.Notification, error)
	RevertToPending(id uuid.UUID) error
	MarkQueued(id uuid.UUID) error
	BatchSetEnabled(ids []uuid.UUID, enabled bool) error
}

// NotificationTemplateRepository defines the interface for notification template data access
type NotificationTemplateRepository interface {
	// Basic CRUD operations
	Create(template *models.NotificationTemplate) error
	GetByID(id uuid.UUID) (*models.NotificationTemplate, error)
	Update(template *models.NotificationTemplate) error
	Delete(id uuid.UUID) error

	// Query operations
	GetByUserID(userID uuid.UUID) ([]*models.NotificationTemplate, error)
	GetByType(templateType string) ([]*models.NotificationTemplate, error)
	GetDefaultTemplates() ([]*models.NotificationTemplate, error)
	GetUserTemplate(userID uuid.UUID, templateType string) (*models.NotificationTemplate, error)

	// Template management
	SetAsDefault(id uuid.UUID, userID uuid.UUID) error
	UnsetDefault(userID uuid.UUID, templateType string) error
}

