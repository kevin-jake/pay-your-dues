package database

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"pay-your-dues/internal/models"
)

type Database struct {
	DB *gorm.DB
}

func NewDatabase(dsn string) (*Database, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Auto migrate the schema
	if err := db.AutoMigrate(
		&models.User{},
		&models.UserSettings{},
		&models.Contact{},
		&models.UserContact{},
		&models.DebtList{},
		&models.DebtItem{},
		&models.Notification{},
		&models.NotificationTemplate{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %v", err)
	}

	// Add unique constraint on user_contacts (user_id, email)
	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_user_contacts_user_email 
		ON user_contacts(user_id, email) 
		WHERE email IS NOT NULL
	`).Error; err != nil {
		return nil, fmt.Errorf("failed to create unique index on user_contacts: %v", err)
	}

	// Add custom indexes for notification system
	if err := addNotificationIndexes(db); err != nil {
		return nil, fmt.Errorf("failed to create notification indexes: %v", err)
	}

	log.Println("Database connected and migrated successfully")

	return &Database{DB: db}, nil
}

// addNotificationIndexes creates custom indexes for the notification system
func addNotificationIndexes(db *gorm.DB) error {
	// Composite index for notifications by debt_list and installment
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_notifications_debt_installment
		ON notifications(debt_list_id, installment_number)
	`).Error; err != nil {
		return err
	}

	// Partial index for scheduled pending notifications
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_notifications_scheduled_pending
		ON notifications(scheduled_for)
		WHERE status = 'pending'
	`).Error; err != nil {
		return err
	}

	// Partial index for queued notifications awaiting consumer delivery
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_notifications_queued
		ON notifications(scheduled_for)
		WHERE status = 'queued'
	`).Error; err != nil {
		return err
	}

	// Partial index for enabled next run
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_notifications_next_run_enabled
		ON notifications(next_run_at)
		WHERE enabled = true
	`).Error; err != nil {
		return err
	}

	// Index on schedule_type for filtering
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_notifications_schedule_type
		ON notifications(schedule_type)
	`).Error; err != nil {
		return err
	}

	// Index on recipient_type for filtering
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_notifications_recipient_type
		ON notifications(recipient_type)
	`).Error; err != nil {
		return err
	}

	log.Println("✅ Notification indexes created successfully")
	return nil
}

func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
} 