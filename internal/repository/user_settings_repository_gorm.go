package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"

	"pay-your-dues/internal/domain/interfaces"
	"pay-your-dues/internal/models"
)

// userSettingsRepositoryGORM implements the UserSettingsRepository interface using GORM
type userSettingsRepositoryGORM struct {
	db *gorm.DB
}

// NewUserSettingsRepositoryGORM creates a new user settings repository with GORM
func NewUserSettingsRepositoryGORM(db *gorm.DB) interfaces.UserSettingsRepository {
	return &userSettingsRepositoryGORM{
		db: db,
	}
}

// Create creates a new user settings record
func (r *userSettingsRepositoryGORM) Create(ctx context.Context, settings *models.UserSettings) error {
	if err := r.db.WithContext(ctx).Create(settings).Error; err != nil {
		return fmt.Errorf("failed to create user settings: %w", err)
	}
	return nil
}

// GetByUserID retrieves user settings by user ID
func (r *userSettingsRepositoryGORM) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.UserSettings, error) {
	var settings models.UserSettings
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&settings).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Return nil if not found (not an error)
		}
		return nil, fmt.Errorf("failed to get user settings: %w", err)
	}
	return &settings, nil
}

// Update updates an existing user settings record
func (r *userSettingsRepositoryGORM) Update(ctx context.Context, settings *models.UserSettings) error {
	if err := r.db.WithContext(ctx).Save(settings).Error; err != nil {
		return fmt.Errorf("failed to update user settings: %w", err)
	}
	return nil
}

// GetOrCreate retrieves user settings or creates default settings if they don't exist
func (r *userSettingsRepositoryGORM) GetOrCreate(ctx context.Context, userID uuid.UUID) (*models.UserSettings, error) {
	settings, err := r.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if settings != nil {
		return settings, nil
	}

	// Create default settings
	defaultSettings := &models.UserSettings{
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

	if err := r.Create(ctx, defaultSettings); err != nil {
		return nil, fmt.Errorf("failed to create default user settings: %w", err)
	}

	return defaultSettings, nil
}

