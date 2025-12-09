package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"pay-your-dues/internal/handlers"
	"pay-your-dues/internal/mocks"
	"pay-your-dues/internal/models"
	"pay-your-dues/internal/repository"
	"pay-your-dues/internal/services"
)

func TestUserSettingsHandler_GetUserSettings(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupMocks     func(*mocks.MockUserSettingsService)
		setupContext   func(*gin.Context)
		expectedStatus int
		validateBody   func(*testing.T, map[string]interface{})
	}{
		{
			name: "successful get user settings",
			setupMocks: func(mockService *mocks.MockUserSettingsService) {
				userID := uuid.New()
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
				mockService.On("GetUserSettings", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(settings, nil)
			},
			setupContext: func(c *gin.Context) {
				c.Set("user_id", uuid.New())
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "User settings retrieved successfully", body["message"])
				data, ok := body["data"].(map[string]interface{})
				assert.True(t, ok)
				assert.True(t, data["notification_email"].(bool))
				assert.False(t, data["notification_sms"].(bool))
			},
		},
		{
			name: "unauthorized - missing user_id",
			setupMocks: func(mockService *mocks.MockUserSettingsService) {
				// No mock setup needed
			},
			setupContext: func(c *gin.Context) {
				// Don't set user_id
			},
			expectedStatus: http.StatusUnauthorized,
			validateBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "Unauthorized", body["error"])
			},
		},
		{
			name: "service error",
			setupMocks: func(mockService *mocks.MockUserSettingsService) {
				mockService.On("GetUserSettings", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, assert.AnError)
			},
			setupContext: func(c *gin.Context) {
				c.Set("user_id", uuid.New())
			},
			expectedStatus: http.StatusInternalServerError,
			validateBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "Failed to get user settings", body["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockService := &mocks.MockUserSettingsService{}
			tt.setupMocks(mockService)

			logger := zerolog.Nop()
			handler := handlers.NewUserSettingsHandler(mockService, logger)

			// Prepare request
			req := httptest.NewRequest(http.MethodGet, "/api/v1/settings", nil)
			w := httptest.NewRecorder()

			// Setup Gin context
			router := gin.New()
			router.GET("/api/v1/settings", func(c *gin.Context) {
				tt.setupContext(c)
				handler.GetUserSettings(c)
			})

			// Execute
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			var responseBody map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &responseBody)
			assert.NoError(t, err)

			tt.validateBody(t, responseBody)

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestUserSettingsHandler_UpdateUserSettings(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMocks     func(*mocks.MockUserSettingsService)
		setupContext   func(*gin.Context)
		expectedStatus int
		validateBody   func(*testing.T, map[string]interface{})
	}{
		{
			name: "successful update notification email",
			requestBody: map[string]interface{}{
				"notification_email": false,
			},
			setupMocks: func(mockService *mocks.MockUserSettingsService) {
				userID := uuid.New()
				updatedSettings := &models.UserSettings{
					ID:                        uuid.New(),
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
				mockService.On("UpdateUserSettings", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("*models.UpdateUserSettingsRequest")).Return(updatedSettings, nil)
			},
			setupContext: func(c *gin.Context) {
				c.Set("user_id", uuid.New())
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "User settings updated successfully", body["message"])
				data, ok := body["data"].(map[string]interface{})
				assert.True(t, ok)
				assert.False(t, data["notification_email"].(bool))
			},
		},
		{
			name: "successful update webhook configuration",
			requestBody: map[string]interface{}{
				"slack_webhook_url":    "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX",
				"discord_webhook_url":  "https://discord.com/api/webhooks/123456789/abcdefghijklmnopqrstuvwxyz",
				"notification_webhook": true,
			},
			setupMocks: func(mockService *mocks.MockUserSettingsService) {
				userID := uuid.New()
				slackURL := "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"
				discordURL := "https://discord.com/api/webhooks/123456789/abcdefghijklmnopqrstuvwxyz"
				updatedSettings := &models.UserSettings{
					ID:                        uuid.New(),
					UserID:                    userID,
					NotificationEmail:         true,
					NotificationSMS:           false,
					NotificationWebhook:       true,
					NotificationReminderDays:  pq.Int64Array{7, 3, 1},
					NotificationTime:          "09:00:00",
					OverdueReminderFrequency:  "daily",
					EventNotificationsEnabled: true,
					NotifyContactOnPayment:    true,
					DefaultCurrency:           "Php",
					Timezone:                  "UTC",
					SlackWebhookURL:           &slackURL,
					DiscordWebhookURL:         &discordURL,
				}
				mockService.On("UpdateUserSettings", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("*models.UpdateUserSettingsRequest")).Return(updatedSettings, nil)
			},
			setupContext: func(c *gin.Context) {
				c.Set("user_id", uuid.New())
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "User settings updated successfully", body["message"])
				data, ok := body["data"].(map[string]interface{})
				assert.True(t, ok)
				assert.NotNil(t, data["slack_webhook_url"])
				assert.Equal(t, "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX", data["slack_webhook_url"])
			},
		},
		{
			name: "invalid request body",
			requestBody: "invalid json",
			setupMocks: func(mockService *mocks.MockUserSettingsService) {
				// No mock setup needed
			},
			setupContext: func(c *gin.Context) {
				c.Set("user_id", uuid.New())
			},
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "Invalid request body", body["error"])
			},
		},
		{
			name: "unauthorized - missing user_id",
			requestBody: map[string]interface{}{
				"notification_email": false,
			},
			setupMocks: func(mockService *mocks.MockUserSettingsService) {
				// No mock setup needed
			},
			setupContext: func(c *gin.Context) {
				// Don't set user_id
			},
			expectedStatus: http.StatusUnauthorized,
			validateBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "Unauthorized", body["error"])
			},
		},
		{
			name: "service error",
			requestBody: map[string]interface{}{
				"notification_email": false,
			},
			setupMocks: func(mockService *mocks.MockUserSettingsService) {
				mockService.On("UpdateUserSettings", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("*models.UpdateUserSettingsRequest")).Return(nil, assert.AnError)
			},
			setupContext: func(c *gin.Context) {
				c.Set("user_id", uuid.New())
			},
			expectedStatus: http.StatusInternalServerError,
			validateBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "Failed to update user settings", body["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockService := &mocks.MockUserSettingsService{}
			tt.setupMocks(mockService)

			logger := zerolog.Nop()
			handler := handlers.NewUserSettingsHandler(mockService, logger)

			// Prepare request
			var requestBody []byte
			if str, ok := tt.requestBody.(string); ok {
				requestBody = []byte(str)
			} else {
				requestBody, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPut, "/api/v1/settings", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Setup Gin context
			router := gin.New()
			router.PUT("/api/v1/settings", func(c *gin.Context) {
				tt.setupContext(c)
				handler.UpdateUserSettings(c)
			})

			// Execute
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			var responseBody map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &responseBody)
			assert.NoError(t, err)

			tt.validateBody(t, responseBody)

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

// TestUserSettingsHandler_Integration tests the handler with real service and repository
func TestUserSettingsHandler_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Auto-migrate
	err = db.AutoMigrate(
		&models.User{},
		&models.UserSettings{},
	)
	assert.NoError(t, err)

	// Create user
	userID := uuid.New()
	user := &models.User{
		ID:        userID,
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}
	db.Create(user)

	// Initialize repositories and services
	userRepo := repository.NewUserRepositoryGORM(db)
	settingsRepo := repository.NewUserSettingsRepositoryGORM(db)
	logger := zerolog.Nop()
	settingsService := services.NewUserSettingsService(settingsRepo, userRepo, logger)
	handler := handlers.NewUserSettingsHandler(settingsService, logger)

	// Test GET - should create default settings
	req := httptest.NewRequest(http.MethodGet, "/api/v1/settings", nil)
	w := httptest.NewRecorder()

	router := gin.New()
	router.GET("/api/v1/settings", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.GetUserSettings(c)
	})

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var getResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &getResponse)
	assert.NoError(t, err)
	assert.Equal(t, "User settings retrieved successfully", getResponse["message"])

	// Test PUT - update settings
	updateBody := map[string]interface{}{
		"notification_email": false,
		"notification_sms":    true,
		"telegram_bot_token":  "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
		"telegram_chat_id":     "123456789",
	}
	requestBody, _ := json.Marshal(updateBody)

	req = httptest.NewRequest(http.MethodPut, "/api/v1/settings", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.PUT("/api/v1/settings", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.UpdateUserSettings(c)
	})

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var updateResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &updateResponse)
	assert.NoError(t, err)
	assert.Equal(t, "User settings updated successfully", updateResponse["message"])

	data, ok := updateResponse["data"].(map[string]interface{})
	assert.True(t, ok)
	assert.False(t, data["notification_email"].(bool))
	assert.True(t, data["notification_sms"].(bool))
	assert.NotNil(t, data["telegram_bot_token"])
}

