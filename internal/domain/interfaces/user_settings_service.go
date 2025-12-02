package interfaces

import (
	"context"

	"github.com/google/uuid"

	"pay-your-dues/internal/models"
)

// UserSettingsService defines the interface for user settings operations
type UserSettingsService interface {
	GetUserSettings(ctx context.Context, userID uuid.UUID) (*models.UserSettings, error)
	UpdateUserSettings(ctx context.Context, userID uuid.UUID, req *models.UpdateUserSettingsRequest) (*models.UserSettings, error)
	GetOrCreateUserSettings(ctx context.Context, userID uuid.UUID) (*models.UserSettings, error)
}

