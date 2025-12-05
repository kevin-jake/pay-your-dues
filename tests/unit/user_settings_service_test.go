package unit

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/rs/zerolog"

	"pay-your-dues/internal/domain/entities"
	"pay-your-dues/internal/mocks"
	"pay-your-dues/internal/models"
	"pay-your-dues/internal/services"
)

func TestUserSettingsService_GetUserSettings(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name          string
		userID        uuid.UUID
		setupMocks    func(*mocks.MockUserSettingsRepository, *mocks.MockUserRepository)
		expectedError string
		validateResult func(*testing.T, *models.UserSettings)
	}{
		{
			name:   "get existing user settings",
			userID: userID,
			setupMocks: func(settingsRepo *mocks.MockUserSettingsRepository, userRepo *mocks.MockUserRepository) {
				user := &entities.User{
					ID:        userID,
					Email:     "test@example.com",
					FirstName: "Test",
					LastName:  "User",
				}
				userRepo.On("GetByID", mock.Anything, userID).Return(user, nil)

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
				settingsRepo.On("GetOrCreate", mock.Anything, userID).Return(settings, nil)
			},
			expectedError: "",
			validateResult: func(t *testing.T, settings *models.UserSettings) {
				assert.NotNil(t, settings)
				assert.Equal(t, userID, settings.UserID)
				assert.True(t, settings.NotificationEmail)
				assert.False(t, settings.NotificationSMS)
				assert.Equal(t, "09:00:00", settings.NotificationTime)
			},
		},
		{
			name:   "create default settings when not found",
			userID: userID,
			setupMocks: func(settingsRepo *mocks.MockUserSettingsRepository, userRepo *mocks.MockUserRepository) {
				user := &entities.User{
					ID:        userID,
					Email:     "test@example.com",
					FirstName: "Test",
					LastName:  "User",
				}
				userRepo.On("GetByID", mock.Anything, userID).Return(user, nil)

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
				settingsRepo.On("GetOrCreate", mock.Anything, userID).Return(defaultSettings, nil)
			},
			expectedError: "",
			validateResult: func(t *testing.T, settings *models.UserSettings) {
				assert.NotNil(t, settings)
				assert.Equal(t, userID, settings.UserID)
				assert.True(t, settings.NotificationEmail)
				assert.Equal(t, pq.Int64Array{7, 3, 1}, settings.NotificationReminderDays)
			},
		},
		{
			name:   "user not found",
			userID: userID,
			setupMocks: func(settingsRepo *mocks.MockUserSettingsRepository, userRepo *mocks.MockUserRepository) {
				userRepo.On("GetByID", mock.Anything, userID).Return(nil, assert.AnError)
			},
			expectedError: "user not found",
			validateResult: func(t *testing.T, settings *models.UserSettings) {
				assert.Nil(t, settings)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockSettingsRepo := &mocks.MockUserSettingsRepository{}
			mockUserRepo := &mocks.MockUserRepository{}
			tt.setupMocks(mockSettingsRepo, mockUserRepo)

			// Create service
			logger := zerolog.Nop()
			service := services.NewUserSettingsService(mockSettingsRepo, mockUserRepo, logger)

			// Execute
			ctx := context.Background()
			settings, err := service.GetUserSettings(ctx, tt.userID)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				tt.validateResult(t, nil)
			} else {
				assert.NoError(t, err)
				tt.validateResult(t, settings)
			}

			// Verify mock expectations
			mockSettingsRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestUserSettingsService_UpdateUserSettings(t *testing.T) {
	userID := uuid.New()
	settingsID := uuid.New()

	tests := []struct {
		name          string
		userID        uuid.UUID
		request       *models.UpdateUserSettingsRequest
		setupMocks    func(*mocks.MockUserSettingsRepository, *mocks.MockUserRepository)
		expectedError string
		validateResult func(*testing.T, *models.UserSettings)
	}{
		{
			name:   "update notification email setting",
			userID: userID,
			request: &models.UpdateUserSettingsRequest{
				NotificationEmail: boolPtr(false),
			},
			setupMocks: func(settingsRepo *mocks.MockUserSettingsRepository, userRepo *mocks.MockUserRepository) {
				user := &entities.User{
					ID:        userID,
					Email:     "test@example.com",
					FirstName: "Test",
					LastName:  "User",
				}
				userRepo.On("GetByID", mock.Anything, userID).Return(user, nil)

				existingSettings := &models.UserSettings{
					ID:                        settingsID,
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
				settingsRepo.On("GetOrCreate", mock.Anything, userID).Return(existingSettings, nil)

				updatedSettings := &models.UserSettings{
					ID:                        settingsID,
					UserID:                    userID,
					NotificationEmail:         false,
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
				settingsRepo.On("Update", mock.Anything, mock.MatchedBy(func(s *models.UserSettings) bool {
					return s.UserID == userID && s.NotificationEmail == false
				})).Return(nil).Run(func(args mock.Arguments) {
					s := args.Get(1).(*models.UserSettings)
					*s = *updatedSettings
				})
			},
			expectedError: "",
			validateResult: func(t *testing.T, settings *models.UserSettings) {
				assert.NotNil(t, settings)
				assert.Equal(t, userID, settings.UserID)
				assert.False(t, settings.NotificationEmail)
			},
		},
		{
			name:   "update webhook configuration",
			userID: userID,
			request: &models.UpdateUserSettingsRequest{
				TelegramBotToken: stringPtr("123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"),
				TelegramChatID:   stringPtr("123456789"),
			},
			setupMocks: func(settingsRepo *mocks.MockUserSettingsRepository, userRepo *mocks.MockUserRepository) {
				user := &entities.User{
					ID:        userID,
					Email:     "test@example.com",
					FirstName: "Test",
					LastName:  "User",
				}
				userRepo.On("GetByID", mock.Anything, userID).Return(user, nil)

				existingSettings := &models.UserSettings{
					ID:                        settingsID,
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
				settingsRepo.On("GetOrCreate", mock.Anything, userID).Return(existingSettings, nil)

				botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
				chatID := "123456789"
				updatedSettings := &models.UserSettings{
					ID:                        settingsID,
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
					TelegramBotToken:          &botToken,
					TelegramChatID:            &chatID,
				}
				settingsRepo.On("Update", mock.Anything, mock.MatchedBy(func(s *models.UserSettings) bool {
					return s.UserID == userID && s.TelegramBotToken != nil && *s.TelegramBotToken == botToken
				})).Return(nil).Run(func(args mock.Arguments) {
					s := args.Get(1).(*models.UserSettings)
					*s = *updatedSettings
				})
			},
			expectedError: "",
			validateResult: func(t *testing.T, settings *models.UserSettings) {
				assert.NotNil(t, settings)
				assert.Equal(t, userID, settings.UserID)
				assert.NotNil(t, settings.TelegramBotToken)
				assert.Equal(t, "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11", *settings.TelegramBotToken)
				assert.NotNil(t, settings.TelegramChatID)
				assert.Equal(t, "123456789", *settings.TelegramChatID)
			},
		},
		{
			name:   "update multiple fields",
			userID: userID,
			request: &models.UpdateUserSettingsRequest{
				NotificationEmail:        boolPtr(true),
				NotificationSMS:          boolPtr(true),
				NotificationWebhook:      boolPtr(true),
				NotificationReminderDays: &pq.Int64Array{14, 7, 1},
				NotificationTime:          stringPtr("10:00:00"),
			},
			setupMocks: func(settingsRepo *mocks.MockUserSettingsRepository, userRepo *mocks.MockUserRepository) {
				user := &entities.User{
					ID:        userID,
					Email:     "test@example.com",
					FirstName: "Test",
					LastName:  "User",
				}
				userRepo.On("GetByID", mock.Anything, userID).Return(user, nil)

				existingSettings := &models.UserSettings{
					ID:                        settingsID,
					UserID:                    userID,
					NotificationEmail:         false,
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
				settingsRepo.On("GetOrCreate", mock.Anything, userID).Return(existingSettings, nil)

				updatedSettings := &models.UserSettings{
					ID:                        settingsID,
					UserID:                    userID,
					NotificationEmail:         true,
					NotificationSMS:           true,
					NotificationWebhook:       true,
					NotificationReminderDays:  pq.Int64Array{14, 7, 1},
					NotificationTime:          "10:00:00",
					OverdueReminderFrequency:  "daily",
					EventNotificationsEnabled: true,
					NotifyContactOnPayment:    true,
					DefaultCurrency:           "Php",
					Timezone:                  "UTC",
				}
				settingsRepo.On("Update", mock.Anything, mock.MatchedBy(func(s *models.UserSettings) bool {
					return s.UserID == userID && s.NotificationEmail == true && s.NotificationSMS == true
				})).Return(nil).Run(func(args mock.Arguments) {
					s := args.Get(1).(*models.UserSettings)
					*s = *updatedSettings
				})
			},
			expectedError: "",
			validateResult: func(t *testing.T, settings *models.UserSettings) {
				assert.NotNil(t, settings)
				assert.Equal(t, userID, settings.UserID)
				assert.True(t, settings.NotificationEmail)
				assert.True(t, settings.NotificationSMS)
				assert.True(t, settings.NotificationWebhook)
				assert.Equal(t, pq.Int64Array{14, 7, 1}, settings.NotificationReminderDays)
				assert.Equal(t, "10:00:00", settings.NotificationTime)
			},
		},
		{
			name:   "user not found",
			userID: userID,
			request: &models.UpdateUserSettingsRequest{
				NotificationEmail: boolPtr(false),
			},
			setupMocks: func(settingsRepo *mocks.MockUserSettingsRepository, userRepo *mocks.MockUserRepository) {
				userRepo.On("GetByID", mock.Anything, userID).Return(nil, assert.AnError)
			},
			expectedError: "user not found",
			validateResult: func(t *testing.T, settings *models.UserSettings) {
				assert.Nil(t, settings)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockSettingsRepo := &mocks.MockUserSettingsRepository{}
			mockUserRepo := &mocks.MockUserRepository{}
			tt.setupMocks(mockSettingsRepo, mockUserRepo)

			// Create service
			logger := zerolog.Nop()
			service := services.NewUserSettingsService(mockSettingsRepo, mockUserRepo, logger)

			// Execute
			ctx := context.Background()
			settings, err := service.UpdateUserSettings(ctx, tt.userID, tt.request)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				tt.validateResult(t, nil)
			} else {
				assert.NoError(t, err)
				tt.validateResult(t, settings)
			}

			// Verify mock expectations
			mockSettingsRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}

