package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type UserSettings struct {
	ID                  uuid.UUID `json:"id" gorm:"type:uuid;primary_key"`
	UserID              uuid.UUID `json:"user_id" gorm:"type:uuid;not null;uniqueIndex"`
	NotificationEmail   bool      `json:"notification_email" gorm:"default:true"`
	NotificationSMS     bool      `json:"notification_sms" gorm:"default:false"`
	NotificationWebhook bool      `json:"notification_webhook" gorm:"default:false"`
	
	// Notification schedule settings
	NotificationReminderDays   pq.Int64Array `json:"notification_reminder_days" gorm:"type:integer[];default:'{7,3,1}'"`
	NotificationTime           string        `json:"notification_time" gorm:"default:'09:00:00'"`
	OverdueReminderFrequency   string        `json:"overdue_reminder_frequency" gorm:"default:'daily'"`
	CustomEmailMessage         *string       `json:"custom_email_message"`
	CustomSMSMessage           *string       `json:"custom_sms_message"`
	
	// Webhook configurations
	SlackWebhookURL     *string   `json:"slack_webhook_url"`
	TelegramBotToken    *string   `json:"telegram_bot_token"`
	TelegramChatID      *string   `json:"telegram_chat_id"`
	DiscordWebhookURL   *string   `json:"discord_webhook_url"`
	
	// Event notification settings
	EventNotificationsEnabled bool   `json:"event_notifications_enabled" gorm:"default:true"`
	NotifyContactOnPayment    bool   `json:"notify_contact_on_payment" gorm:"default:true"`
	NotificationRecipient     string `json:"notification_recipient" gorm:"default:'both'"` // 'user', 'contact', or 'both'
	
	DefaultCurrency     string    `json:"default_currency" gorm:"default:'Php'"`
	Timezone           string    `json:"timezone" gorm:"default:'UTC'"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	
	// Relationships
	// User relationship is handled in User model to avoid circular reference
}

type UpdateUserSettingsRequest struct {
	NotificationEmail            *bool         `json:"notification_email"`
	NotificationSMS              *bool         `json:"notification_sms"`
	NotificationWebhook          *bool         `json:"notification_webhook"`
	NotificationReminderDays     *pq.Int64Array `json:"notification_reminder_days"`
	NotificationTime             *string       `json:"notification_time"`
	OverdueReminderFrequency     *string       `json:"overdue_reminder_frequency"`
	CustomEmailMessage           *string       `json:"custom_email_message"`
	CustomSMSMessage             *string       `json:"custom_sms_message"`
	SlackWebhookURL              *string       `json:"slack_webhook_url"`
	TelegramBotToken             *string       `json:"telegram_bot_token"`
	TelegramChatID               *string       `json:"telegram_chat_id"`
	DiscordWebhookURL            *string       `json:"discord_webhook_url"`
	EventNotificationsEnabled    *bool         `json:"event_notifications_enabled"`
	NotifyContactOnPayment       *bool         `json:"notify_contact_on_payment"`
	NotificationRecipient        *string       `json:"notification_recipient"` // 'user', 'contact', or 'both'
	DefaultCurrency              *string       `json:"default_currency"`
	Timezone                     *string       `json:"timezone"`
} 