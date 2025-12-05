package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"pay-your-dues/internal/models"
)

// MockUserSettingsRepository is a mock implementation of UserSettingsRepository
type MockUserSettingsRepository struct {
	mock.Mock
}

func (m *MockUserSettingsRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.UserSettings, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserSettings), args.Error(1)
}

func (m *MockUserSettingsRepository) Create(ctx context.Context, settings *models.UserSettings) error {
	args := m.Called(ctx, settings)
	return args.Error(0)
}

func (m *MockUserSettingsRepository) Update(ctx context.Context, settings *models.UserSettings) error {
	args := m.Called(ctx, settings)
	return args.Error(0)
}

func (m *MockUserSettingsRepository) GetOrCreate(ctx context.Context, userID uuid.UUID) (*models.UserSettings, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserSettings), args.Error(1)
}

