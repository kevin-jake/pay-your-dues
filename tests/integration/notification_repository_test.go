package integration

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"pay-your-dues/internal/domain/interfaces"
	"pay-your-dues/internal/models"
	"pay-your-dues/internal/repository"
)

type NotificationRepositoryTestSuite struct {
	suite.Suite
	db   *gorm.DB
	repo interfaces.NotificationRepository
}

func (suite *NotificationRepositoryTestSuite) SetupSuite() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	// SQLite-compatible schema (Notification model uses Postgres-only types for AutoMigrate)
	err = db.Exec(`
		CREATE TABLE notifications (
			id TEXT PRIMARY KEY,
			debt_list_id TEXT NOT NULL,
			debt_item_id TEXT,
			installment_number INTEGER,
			installment_due_date DATETIME,
			notification_type TEXT NOT NULL,
			webhook_type TEXT,
			recipient_type TEXT NOT NULL,
			message TEXT NOT NULL,
			status TEXT DEFAULT 'pending',
			sent_at DATETIME,
			schedule_type TEXT,
			scheduled_for DATETIME,
			cron_job_id TEXT,
			reminder_days_before INTEGER,
			use_custom_schedule INTEGER DEFAULT 0,
			custom_reminder_days TEXT,
			custom_notification_time TEXT,
			custom_message TEXT,
			last_sent_at DATETIME,
			next_run_at DATETIME,
			is_recurring INTEGER DEFAULT 0,
			enabled INTEGER DEFAULT 1,
			recipient_email TEXT,
			recipient_phone TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error
	suite.Require().NoError(err)

	suite.db = db
	suite.repo = repository.NewNotificationRepositoryGORM(db)
}

func (suite *NotificationRepositoryTestSuite) TestBatchSetEnabled() {
	id1 := uuid.New()
	id2 := uuid.New()
	now := time.Now()

	notifications := []*models.Notification{
		{
			ID:               id1,
			DebtListID:       uuid.New(),
			NotificationType: "email",
			RecipientType:    "user",
			Message:          "test",
			Status:           "pending",
			Enabled:          true,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               id2,
			DebtListID:       uuid.New(),
			NotificationType: "sms",
			RecipientType:    "user",
			Message:          "test",
			Status:           "pending",
			Enabled:          true,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
	}

	err := suite.repo.CreateBatch(notifications)
	suite.Require().NoError(err)

	err = suite.repo.BatchSetEnabled([]uuid.UUID{id1, id2}, false)
	suite.Require().NoError(err)

	updated1, err := suite.repo.GetByID(id1)
	suite.Require().NoError(err)
	suite.False(updated1.Enabled)

	updated2, err := suite.repo.GetByID(id2)
	suite.Require().NoError(err)
	suite.False(updated2.Enabled)
}

func (suite *NotificationRepositoryTestSuite) TestRevertToPending() {
	id := uuid.New()
	now := time.Now()

	notification := &models.Notification{
		ID:               id,
		DebtListID:       uuid.New(),
		NotificationType: "email",
		RecipientType:    "user",
		Message:          "test",
		Status:           "queued",
		Enabled:          true,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	suite.Require().NoError(suite.repo.Create(notification))

	err := suite.repo.RevertToPending(id)
	suite.Require().NoError(err)

	updated, err := suite.repo.GetByID(id)
	suite.Require().NoError(err)
	suite.Equal("pending", updated.Status)
}

func TestNotificationRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(NotificationRepositoryTestSuite))
}
