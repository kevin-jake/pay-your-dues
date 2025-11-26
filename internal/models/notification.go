package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Notification struct {
	ID                   uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	
	// Core relationships
	DebtListID           uuid.UUID  `json:"debt_list_id" gorm:"type:uuid;not null;index"`
	DebtItemID           *uuid.UUID `json:"debt_item_id" gorm:"type:uuid;index"` // For event-based notifications
	
	// Installment tracking
	InstallmentNumber    *int       `json:"installment_number" gorm:"index:idx_debt_installment"`
	InstallmentDueDate   *time.Time `json:"installment_due_date" gorm:"index:idx_installment_due"`
	
	// Notification settings
	NotificationType     string     `json:"notification_type" gorm:"not null"` // 'email', 'sms', 'webhook'
	WebhookType          *string    `json:"webhook_type"` // 'slack', 'telegram', 'discord'
	RecipientType        string     `json:"recipient_type" gorm:"not null"` // 'user' or 'contact'
	
	Message              string     `json:"message" gorm:"not null"`
	Status               string     `json:"status" gorm:"default:'pending';index"` // 'pending', 'sent', 'failed', 'cancelled'
	SentAt               *time.Time `json:"sent_at"`
	
	// Scheduling fields
	ScheduleType         string     `json:"schedule_type"` // 'reminder', 'overdue', 'manual', 'event'
	ScheduledFor         *time.Time `json:"scheduled_for" gorm:"index:idx_scheduled_for"`
	CronJobID            *string    `json:"cron_job_id"`
	ReminderDaysBefore   *int       `json:"reminder_days_before"`
	
	// Custom schedule
	UseCustomSchedule       bool          `json:"use_custom_schedule" gorm:"default:false"`
	CustomReminderDays      pq.Int64Array `json:"custom_reminder_days" gorm:"type:integer[]"`
	CustomNotificationTime  *string       `json:"custom_notification_time"`
	CustomMessage           *string       `json:"custom_message"`
	
	// Recurring
	LastSentAt           *time.Time `json:"last_sent_at"`
	NextRunAt            *time.Time `json:"next_run_at" gorm:"index:idx_next_run"`
	IsRecurring          bool       `json:"is_recurring" gorm:"default:false"`
	Enabled              bool       `json:"enabled" gorm:"default:true"`
	
	// Optional recipient override (normally fetched from user_contacts)
	RecipientEmail       *string    `json:"recipient_email,omitempty"`
	RecipientPhone       *string    `json:"recipient_phone,omitempty"`
	
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type CreateNotificationRequest struct {
	DebtListID           uuid.UUID  `json:"debt_list_id" binding:"required"`
	InstallmentNumber    *int       `json:"installment_number"`
	InstallmentDueDate   *time.Time `json:"installment_due_date"`
	NotificationType     string     `json:"notification_type" binding:"required,oneof=email sms webhook"`
	WebhookType          *string    `json:"webhook_type" binding:"omitempty,oneof=slack telegram discord"`
	RecipientType        string     `json:"recipient_type" binding:"required,oneof=user contact"`
	Message              string     `json:"message" binding:"required"`
	RecipientEmail       *string    `json:"recipient_email,omitempty"`
	RecipientPhone       *string    `json:"recipient_phone,omitempty"`
	ScheduledFor         *time.Time `json:"scheduled_for,omitempty"` // For manual testing
	ScheduleType         string     `json:"schedule_type,omitempty"` // 'reminder', 'overdue', 'manual', 'event'
} 