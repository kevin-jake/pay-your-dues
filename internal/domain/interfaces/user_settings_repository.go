package interfaces

import (
	"context"

	"github.com/google/uuid"

	"pay-your-dues/internal/models"
)

// UserSettingsRepository defines the interface for user settings data access operations
type UserSettingsRepository interface {
	Create(ctx context.Context, settings *models.UserSettings) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*models.UserSettings, error)
	Update(ctx context.Context, settings *models.UserSettings) error
	GetOrCreate(ctx context.Context, userID uuid.UUID) (*models.UserSettings, error)
}

