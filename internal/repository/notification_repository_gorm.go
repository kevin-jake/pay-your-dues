package repository

import (
	"time"

	"pay-your-dues/internal/domain/interfaces"
	"pay-your-dues/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationRepositoryGORM implements the NotificationRepository interface using GORM
type NotificationRepositoryGORM struct {
	db *gorm.DB
}

// NewNotificationRepositoryGORM creates a new notification repository instance
func NewNotificationRepositoryGORM(db *gorm.DB) interfaces.NotificationRepository {
	return &NotificationRepositoryGORM{db: db}
}

// Create creates a new notification
func (r *NotificationRepositoryGORM) Create(notification *models.Notification) error {
	if notification.ID == uuid.Nil {
		notification.ID = uuid.New()
	}
	return r.db.Create(notification).Error
}

// GetByID retrieves a notification by ID
func (r *NotificationRepositoryGORM) GetByID(id uuid.UUID) (*models.Notification, error) {
	var notification models.Notification
	err := r.db.Where("id = ?", id).First(&notification).Error
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

// Update updates a notification
func (r *NotificationRepositoryGORM) Update(notification *models.Notification) error {
	return r.db.Save(notification).Error
}

// Delete deletes a notification
func (r *NotificationRepositoryGORM) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&models.Notification{}).Error
}

// GetByDebtListID retrieves all notifications for a debt list
func (r *NotificationRepositoryGORM) GetByDebtListID(debtListID uuid.UUID) ([]*models.Notification, error) {
	var notifications []*models.Notification
	err := r.db.Where("debt_list_id = ?", debtListID).
		Order("created_at DESC").
		Find(&notifications).Error
	if err != nil {
		return nil, err
	}
	return notifications, nil
}

// GetByDebtListAndInstallment retrieves notifications for a specific installment
func (r *NotificationRepositoryGORM) GetByDebtListAndInstallment(debtListID uuid.UUID, installmentNumber int) ([]*models.Notification, error) {
	var notifications []*models.Notification
	err := r.db.Where("debt_list_id = ? AND installment_number = ?", debtListID, installmentNumber).
		Order("scheduled_for ASC").
		Find(&notifications).Error
	if err != nil {
		return nil, err
	}
	return notifications, nil
}

// GetPendingNotifications retrieves pending notifications up to the specified limit
func (r *NotificationRepositoryGORM) GetPendingNotifications(limit int) ([]*models.Notification, error) {
	var notifications []*models.Notification
	err := r.db.Where("status = ? AND enabled = ?", "pending", true).
		Order("scheduled_for ASC").
		Limit(limit).
		Find(&notifications).Error
	if err != nil {
		return nil, err
	}
	return notifications, nil
}

// GetScheduledNotifications retrieves notifications scheduled before the specified time
func (r *NotificationRepositoryGORM) GetScheduledNotifications(beforeTime time.Time, limit int) ([]*models.Notification, error) {
	var notifications []*models.Notification
	err := r.db.Where("status = ? AND enabled = ? AND scheduled_for <= ?", "pending", true, beforeTime).
		Order("scheduled_for ASC").
		Limit(limit).
		Find(&notifications).Error
	if err != nil {
		return nil, err
	}
	return notifications, nil
}

// GetByStatus retrieves notifications by status
func (r *NotificationRepositoryGORM) GetByStatus(status string, limit int) ([]*models.Notification, error) {
	var notifications []*models.Notification
	query := r.db.Where("status = ?", status).Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	err := query.Find(&notifications).Error
	if err != nil {
		return nil, err
	}
	return notifications, nil
}

// CreateBatch creates multiple notifications in a batch
func (r *NotificationRepositoryGORM) CreateBatch(notifications []*models.Notification) error {
	if len(notifications) == 0 {
		return nil
	}

	// Set IDs for notifications that don't have one
	for _, notification := range notifications {
		if notification.ID == uuid.Nil {
			notification.ID = uuid.New()
		}
	}

	return r.db.Create(&notifications).Error
}

// UpdateStatus updates the status of a notification
func (r *NotificationRepositoryGORM) UpdateStatus(id uuid.UUID, status string, sentAt *time.Time) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if sentAt != nil {
		updates["sent_at"] = sentAt
		updates["last_sent_at"] = sentAt
	}
	return r.db.Model(&models.Notification{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateNextRun updates the next run time of a notification
func (r *NotificationRepositoryGORM) UpdateNextRun(id uuid.UUID, nextRunAt time.Time) error {
	return r.db.Model(&models.Notification{}).
		Where("id = ?", id).
		Update("next_run_at", nextRunAt).Error
}

// DisableNotification disables a notification
func (r *NotificationRepositoryGORM) DisableNotification(id uuid.UUID) error {
	return r.db.Model(&models.Notification{}).
		Where("id = ?", id).
		Update("enabled", false).Error
}

// EnableNotification enables a notification
func (r *NotificationRepositoryGORM) EnableNotification(id uuid.UUID) error {
	return r.db.Model(&models.Notification{}).
		Where("id = ?", id).
		Update("enabled", true).Error
}

// DisableNotificationsForInstallment disables all notifications for a specific installment
func (r *NotificationRepositoryGORM) DisableNotificationsForInstallment(debtListID uuid.UUID, installmentNumber int) error {
	return r.db.Model(&models.Notification{}).
		Where("debt_list_id = ? AND installment_number = ? AND status = ? AND enabled = ?",
			debtListID, installmentNumber, "pending", true).
		Updates(map[string]interface{}{
			"enabled": false,
			"status":  "cancelled",
		}).Error
}

// GetActiveNotificationsForInstallment retrieves active notifications for a specific installment
func (r *NotificationRepositoryGORM) GetActiveNotificationsForInstallment(debtListID uuid.UUID, installmentNumber int) ([]*models.Notification, error) {
	var notifications []*models.Notification
	err := r.db.Where("debt_list_id = ? AND installment_number = ? AND enabled = ? AND status = ?",
		debtListID, installmentNumber, true, "pending").
		Find(&notifications).Error
	if err != nil {
		return nil, err
	}
	return notifications, nil
}

// HasPendingNotifications checks if there are pending notifications for a debt list
func (r *NotificationRepositoryGORM) HasPendingNotifications(debtListID uuid.UUID, installmentNumber *int) (bool, error) {
	var count int64
	query := r.db.Model(&models.Notification{}).
		Where("debt_list_id = ? AND status = ? AND enabled = ?", debtListID, "pending", true)

	if installmentNumber != nil {
		query = query.Where("installment_number = ?", *installmentNumber)
	}

	err := query.Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// CountNotificationsByDebtList counts all notifications for a debt list
func (r *NotificationRepositoryGORM) CountNotificationsByDebtList(debtListID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Notification{}).
		Where("debt_list_id = ?", debtListID).
		Count(&count).Error
	return count, err
}

// Helper method to check if notification belongs to user
func (r *NotificationRepositoryGORM) BelongsToUser(notificationID uuid.UUID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.Notification{}).
		Joins("JOIN debt_lists ON notifications.debt_list_id = debt_lists.id").
		Where("notifications.id = ? AND debt_lists.user_id = ?", notificationID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetUserNotifications retrieves notifications for a user with optional filters
func (r *NotificationRepositoryGORM) GetUserNotifications(userID uuid.UUID, status string, debtListID *uuid.UUID, limit int) ([]*models.Notification, error) {
	var notifications []*models.Notification
	query := r.db.Joins("JOIN debt_lists ON notifications.debt_list_id = debt_lists.id").
		Where("debt_lists.user_id = ?", userID).
		Order("notifications.created_at DESC")

	if status != "" {
		query = query.Where("notifications.status = ?", status)
	}
	if debtListID != nil {
		query = query.Where("notifications.debt_list_id = ?", *debtListID)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&notifications).Error
	if err != nil {
		return nil, err
	}
	return notifications, nil
}

// GetNotificationWithDebtList retrieves a notification with its associated debt list
func (r *NotificationRepositoryGORM) GetNotificationWithDebtList(id uuid.UUID) (*models.Notification, error) {
	var notification models.Notification
	err := r.db.Preload("DebtList").Where("id = ?", id).First(&notification).Error
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

// DeleteByDebtListID deletes all notifications for a debt list (cascade delete)
func (r *NotificationRepositoryGORM) DeleteByDebtListID(debtListID uuid.UUID) error {
	return r.db.Where("debt_list_id = ?", debtListID).Delete(&models.Notification{}).Error
}

// GetOverdueNotifications retrieves notifications that are overdue
func (r *NotificationRepositoryGORM) GetOverdueNotifications(limit int) ([]*models.Notification, error) {
	var notifications []*models.Notification
	now := time.Now()
	
	query := r.db.Where("status = ? AND enabled = ? AND scheduled_for < ?",
		"pending", true, now).
		Order("scheduled_for ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&notifications).Error
	if err != nil {
		return nil, err
	}
	return notifications, nil
}

// UpdateCronJobID updates the cron job ID for a notification
func (r *NotificationRepositoryGORM) UpdateCronJobID(id uuid.UUID, cronJobID string) error {
	return r.db.Model(&models.Notification{}).
		Where("id = ?", id).
		Update("cron_job_id", cronJobID).Error
}

// GetByCronJobID retrieves a notification by its cron job ID
func (r *NotificationRepositoryGORM) GetByCronJobID(cronJobID string) (*models.Notification, error) {
	var notification models.Notification
	err := r.db.Where("cron_job_id = ?", cronJobID).First(&notification).Error
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

// CancelStaleNotifications cancels notifications that are far past their scheduled time
func (r *NotificationRepositoryGORM) CancelStaleNotifications(daysOld int) error {
	cutoffDate := time.Now().AddDate(0, 0, -daysOld)
	return r.db.Model(&models.Notification{}).
		Where("status = ? AND scheduled_for < ?", "pending", cutoffDate).
		Updates(map[string]interface{}{
			"status":  "cancelled",
			"enabled": false,
		}).Error
}

// GetPendingNotificationsByUserID retrieves all pending notifications for a user
func (r *NotificationRepositoryGORM) GetPendingNotificationsByUserID(userID uuid.UUID) ([]*models.Notification, error) {
	var notifications []*models.Notification
	err := r.db.Joins("JOIN debt_lists ON notifications.debt_list_id = debt_lists.id").
		Where("debt_lists.user_id = ? AND notifications.status = ? AND notifications.enabled = ?",
			userID, "pending", true).
		Find(&notifications).Error
	if err != nil {
		return nil, err
	}
	return notifications, nil
}

// ClaimDueNotifications atomically claims due notifications for the scheduler.
func (r *NotificationRepositoryGORM) ClaimDueNotifications(beforeTime time.Time, limit int) ([]*models.Notification, error) {
	var notifications []*models.Notification

	err := r.db.Raw(`
		WITH claimed AS (
			SELECT id
			FROM notifications
			WHERE status = 'pending'
				AND enabled = true
				AND scheduled_for <= ?
			ORDER BY scheduled_for ASC
			LIMIT ?
			FOR UPDATE SKIP LOCKED
		)
		UPDATE notifications n
		SET status = 'queued', updated_at = NOW()
		FROM claimed c
		WHERE n.id = c.id
		RETURNING n.*
	`, beforeTime, limit).Scan(&notifications).Error
	if err != nil {
		return nil, err
	}

	return notifications, nil
}

// RevertToPending returns a claimed notification to pending after a publish failure.
func (r *NotificationRepositoryGORM) RevertToPending(id uuid.UUID) error {
	return r.db.Model(&models.Notification{}).
		Where("id = ? AND status = 'queued'", id).
		Updates(map[string]interface{}{
			"status":     "pending",
			"updated_at": time.Now(),
		}).Error
}

// MarkQueued marks a notification as queued after a successful publish.
func (r *NotificationRepositoryGORM) MarkQueued(id uuid.UUID) error {
	return r.db.Model(&models.Notification{}).
		Where("id = ? AND status = 'pending'", id).
		Updates(map[string]interface{}{
			"status":     "queued",
			"updated_at": time.Now(),
		}).Error
}

// BatchSetEnabled updates enabled flag for multiple notifications in one query.
func (r *NotificationRepositoryGORM) BatchSetEnabled(ids []uuid.UUID, enabled bool) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.Model(&models.Notification{}).
		Where("id IN ?", ids).
		Update("enabled", enabled).Error
}

// UpdatePendingNotificationsByUserID updates all pending notifications for a user
func (r *NotificationRepositoryGORM) UpdatePendingNotificationsByUserID(userID uuid.UUID, updates map[string]interface{}) error {
	// Get all debt list IDs for this user
	var debtListIDs []uuid.UUID
	err := r.db.Model(&models.DebtList{}).
		Where("user_id = ?", userID).
		Pluck("id", &debtListIDs).Error
	if err != nil {
		return err
	}

	if len(debtListIDs) == 0 {
		return nil // No debt lists, nothing to update
	}

	// Update all pending notifications for these debt lists
	return r.db.Model(&models.Notification{}).
		Where("debt_list_id IN ? AND status = ? AND enabled = ?",
			debtListIDs, "pending", true).
		Updates(updates).Error
}

// DeleteByDebtListIDAndSlot deletes all notifications for a specific (installment, scheduled_for) slot
func (r *NotificationRepositoryGORM) DeleteByDebtListIDAndSlot(debtListID uuid.UUID, installmentNumber *int, scheduledFor time.Time) error {
	query := r.db.Where("debt_list_id = ? AND DATE(scheduled_for) = DATE(?)", debtListID, scheduledFor)
	if installmentNumber != nil {
		query = query.Where("installment_number = ?", *installmentNumber)
	} else {
		query = query.Where("installment_number IS NULL")
	}
	return query.Delete(&models.Notification{}).Error
}

// DeleteReminderNotificationsByDebtList deletes reminder and overdue notifications for a debt list
func (r *NotificationRepositoryGORM) DeleteReminderNotificationsByDebtList(debtListID uuid.UUID) error {
	return r.db.Where("debt_list_id = ? AND schedule_type IN ?", debtListID, []string{"reminder", "overdue"}).
		Delete(&models.Notification{}).Error
}

// CountManualNotificationsByDebtAndType counts manual sends for a debt + channel
func (r *NotificationRepositoryGORM) CountManualNotificationsByDebtAndType(debtListID uuid.UUID, notificationType string) (int64, error) {
	var count int64
	err := r.db.Model(&models.Notification{}).
		Where("debt_list_id = ? AND notification_type = ? AND schedule_type = ?",
			debtListID, notificationType, "manual").
		Count(&count).Error
	return count, err
}

