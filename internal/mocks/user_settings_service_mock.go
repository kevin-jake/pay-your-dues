package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"pay-your-dues/internal/models"
)

// MockUserSettingsService is a mock implementation of UserSettingsService
type MockUserSettingsService struct {
	mock.Mock
}

func (m *MockUserSettingsService) GetUserSettings(ctx context.Context, userID uuid.UUID) (*models.UserSettings, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserSettings), args.Error(1)
}

func (m *MockUserSettingsService) UpdateUserSettings(ctx context.Context, userID uuid.UUID, req *models.UpdateUserSettingsRequest) (*models.UserSettings, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserSettings), args.Error(1)
}

func (m *MockUserSettingsService) GetOrCreateUserSettings(ctx context.Context, userID uuid.UUID) (*models.UserSettings, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserSettings), args.Error(1)
}

