package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"pay-your-dues/internal/domain/interfaces"
	"pay-your-dues/internal/models"
	"pay-your-dues/internal/repository"
)

type UserSettingsRepositoryTestSuite struct {
	suite.Suite
	db              *gorm.DB
	settingsRepo    interfaces.UserSettingsRepository
}

func (suite *UserSettingsRepositoryTestSuite) SetupSuite() {
	// Setup in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	// Auto-migrate all models
	err = db.AutoMigrate(
		&models.User{},
		&models.UserSettings{},
	)
	suite.Require().NoError(err)

	suite.db = db
	suite.settingsRepo = repository.NewUserSettingsRepositoryGORM(db)
}

func (suite *UserSettingsRepositoryTestSuite) SetupTest() {
	// Ensure clean database state before each test
	suite.db.Exec("DELETE FROM user_settings")
	suite.db.Exec("DELETE FROM users")
}

func (suite *UserSettingsRepositoryTestSuite) TearDownTest() {
	// Clean up database after each test
	suite.db.Exec("DELETE FROM user_settings")
	suite.db.Exec("DELETE FROM users")
}

func (suite *UserSettingsRepositoryTestSuite) TestCreate() {
	ctx := context.Background()
	userID := uuid.New()

	// Create a user first (required for foreign key)
	user := &models.User{
		ID:        userID,
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}
	suite.db.Create(user)

	settings := &models.UserSettings{
		ID:                        uuid.New(),
		UserID:                    userID,
		NotificationEmail:         true,
		NotificationSMS:           false,
		NotificationWebhook:       false,
		NotificationReminderDays:  pq.Int64Array{7, 3, 1},
		NotificationTime:          "09:00:00",
		OverdueReminderFrequency:  "daily",
		EventNotificationsEnabled: true,
		NotifyContactOnPayment:    true,
		DefaultCurrency:           "Php",
		Timezone:                  "UTC",
	}

	err := suite.settingsRepo.Create(ctx, settings)
	suite.NoError(err)

	// Verify settings were created
	var retrieved models.UserSettings
	err = suite.db.Where("user_id = ?", userID).First(&retrieved).Error
	suite.NoError(err)
	suite.Equal(userID, retrieved.UserID)
	suite.True(retrieved.NotificationEmail)
	suite.Equal(pq.Int64Array{7, 3, 1}, retrieved.NotificationReminderDays)
}

func (suite *UserSettingsRepositoryTestSuite) TestGetByUserID() {
	ctx := context.Background()
	userID := uuid.New()

	// Create a user first
	user := &models.User{
		ID:        userID,
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}
	suite.db.Create(user)

	// Create settings
	settings := &models.UserSettings{
		ID:                        uuid.New(),
		UserID:                    userID,
		NotificationEmail:         true,
		NotificationSMS:           true,
		NotificationWebhook:       false,
		NotificationReminderDays:  pq.Int64Array{14, 7, 1},
		NotificationTime:          "10:00:00",
		OverdueReminderFrequency:  "weekly",
		EventNotificationsEnabled: true,
		NotifyContactOnPayment:    true,
		DefaultCurrency:           "USD",
		Timezone:                  "America/New_York",
	}
	suite.db.Create(settings)

	// Retrieve settings
	retrieved, err := suite.settingsRepo.GetByUserID(ctx, userID)
	suite.NoError(err)
	suite.NotNil(retrieved)
	suite.Equal(userID, retrieved.UserID)
	suite.True(retrieved.NotificationEmail)
	suite.True(retrieved.NotificationSMS)
	suite.Equal(pq.Int64Array{14, 7, 1}, retrieved.NotificationReminderDays)
	suite.Equal("10:00:00", retrieved.NotificationTime)
	suite.Equal("USD", retrieved.DefaultCurrency)
	suite.Equal("America/New_York", retrieved.Timezone)

	// Test non-existent user
	nonExistentID := uuid.New()
	retrieved, err = suite.settingsRepo.GetByUserID(ctx, nonExistentID)
	suite.NoError(err)
	suite.Nil(retrieved)
}

func (suite *UserSettingsRepositoryTestSuite) TestUpdate() {
	ctx := context.Background()
	userID := uuid.New()

	// Create a user first
	user := &models.User{
		ID:        userID,
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}
	suite.db.Create(user)

	// Create initial settings
	settings := &models.UserSettings{
		ID:                        uuid.New(),
		UserID:                    userID,
		NotificationEmail:         true,
		NotificationSMS:           false,
		NotificationWebhook:       false,
		NotificationReminderDays:  pq.Int64Array{7, 3, 1},
		NotificationTime:          "09:00:00",
		OverdueReminderFrequency:  "daily",
		EventNotificationsEnabled: true,
		NotifyContactOnPayment:    true,
		DefaultCurrency:           "Php",
		Timezone:                  "UTC",
	}
	suite.db.Create(settings)

	// Update settings
	settings.NotificationEmail = false
	settings.NotificationSMS = true
	settings.NotificationReminderDays = pq.Int64Array{14, 7, 1}
	settings.NotificationTime = "10:00:00"

	err := suite.settingsRepo.Update(ctx, settings)
	suite.NoError(err)

	// Verify update
	var retrieved models.UserSettings
	err = suite.db.Where("user_id = ?", userID).First(&retrieved).Error
	suite.NoError(err)
	suite.False(retrieved.NotificationEmail)
	suite.True(retrieved.NotificationSMS)
	suite.Equal(pq.Int64Array{14, 7, 1}, retrieved.NotificationReminderDays)
	suite.Equal("10:00:00", retrieved.NotificationTime)
}

func (suite *UserSettingsRepositoryTestSuite) TestGetOrCreate_Existing() {
	ctx := context.Background()
	userID := uuid.New()

	// Create a user first
	user := &models.User{
		ID:        userID,
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}
	suite.db.Create(user)

	// Create existing settings
	existingSettings := &models.UserSettings{
		ID:                        uuid.New(),
		UserID:                    userID,
		NotificationEmail:         true,
		NotificationSMS:           true,
		NotificationWebhook:       true,
		NotificationReminderDays:  pq.Int64Array{14, 7, 1},
		NotificationTime:          "10:00:00",
		OverdueReminderFrequency:  "weekly",
		EventNotificationsEnabled: true,
		NotifyContactOnPayment:    true,
		DefaultCurrency:           "USD",
		Timezone:                  "America/New_York",
	}
	suite.db.Create(existingSettings)

	// Get or create should return existing
	settings, err := suite.settingsRepo.GetOrCreate(ctx, userID)
	suite.NoError(err)
	suite.NotNil(settings)
	suite.Equal(existingSettings.ID, settings.ID)
	suite.True(settings.NotificationEmail)
	suite.Equal("USD", settings.DefaultCurrency)
}

func (suite *UserSettingsRepositoryTestSuite) TestGetOrCreate_New() {
	ctx := context.Background()
	userID := uuid.New()

	// Create a user first
	user := &models.User{
		ID:        userID,
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}
	suite.db.Create(user)

	// Get or create should create default settings
	settings, err := suite.settingsRepo.GetOrCreate(ctx, userID)
	suite.NoError(err)
	suite.NotNil(settings)
	suite.Equal(userID, settings.UserID)
	suite.True(settings.NotificationEmail)
	suite.False(settings.NotificationSMS)
	suite.False(settings.NotificationWebhook)
	suite.Equal(pq.Int64Array{7, 3, 1}, settings.NotificationReminderDays)
	suite.Equal("09:00:00", settings.NotificationTime)
	suite.Equal("daily", settings.OverdueReminderFrequency)
	suite.True(settings.EventNotificationsEnabled)
	suite.True(settings.NotifyContactOnPayment)
	suite.Equal("Php", settings.DefaultCurrency)
	suite.Equal("UTC", settings.Timezone)
}

func (suite *UserSettingsRepositoryTestSuite) TestUpdateWebhookConfigurations() {
	ctx := context.Background()
	userID := uuid.New()

	// Create a user first
	user := &models.User{
		ID:        userID,
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}
	suite.db.Create(user)

	// Create settings
	settings := &models.UserSettings{
		ID:                        uuid.New(),
		UserID:                    userID,
		NotificationEmail:         true,
		NotificationSMS:           false,
		NotificationWebhook:       false,
		NotificationReminderDays:  pq.Int64Array{7, 3, 1},
		NotificationTime:          "09:00:00",
		OverdueReminderFrequency:  "daily",
		EventNotificationsEnabled: true,
		NotifyContactOnPayment:    true,
		DefaultCurrency:           "Php",
		Timezone:                  "UTC",
	}
	suite.db.Create(settings)

	// Update webhook configurations
	slackURL := "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"
	telegramToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	telegramChatID := "123456789"
	discordURL := "https://discord.com/api/webhooks/123456789/abcdefghijklmnopqrstuvwxyz"

	settings.SlackWebhookURL = &slackURL
	settings.TelegramBotToken = &telegramToken
	settings.TelegramChatID = &telegramChatID
	settings.DiscordWebhookURL = &discordURL
	settings.NotificationWebhook = true

	err := suite.settingsRepo.Update(ctx, settings)
	suite.NoError(err)

	// Verify webhook configurations
	var retrieved models.UserSettings
	err = suite.db.Where("user_id = ?", userID).First(&retrieved).Error
	suite.NoError(err)
	suite.NotNil(retrieved.SlackWebhookURL)
	suite.Equal(slackURL, *retrieved.SlackWebhookURL)
	suite.NotNil(retrieved.TelegramBotToken)
	suite.Equal(telegramToken, *retrieved.TelegramBotToken)
	suite.NotNil(retrieved.TelegramChatID)
	suite.Equal(telegramChatID, *retrieved.TelegramChatID)
	suite.NotNil(retrieved.DiscordWebhookURL)
	suite.Equal(discordURL, *retrieved.DiscordWebhookURL)
	suite.True(retrieved.NotificationWebhook)
}

func TestUserSettingsRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserSettingsRepositoryTestSuite))
}

