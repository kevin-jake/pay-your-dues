package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"pay-your-dues/internal/domain/interfaces"
	"pay-your-dues/internal/models"
)

// UserSettingsHandler handles user settings-related HTTP requests
type UserSettingsHandler struct {
	settingsService interfaces.UserSettingsService
	logger          zerolog.Logger
}

// NewUserSettingsHandler creates a new user settings handler
func NewUserSettingsHandler(settingsService interfaces.UserSettingsService, logger zerolog.Logger) *UserSettingsHandler {
	return &UserSettingsHandler{
		settingsService: settingsService,
		logger:          logger.With().Str("handler", "user_settings").Logger(),
	}
}

// GetUserSettings handles GET /api/v1/settings
func (h *UserSettingsHandler) GetUserSettings(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	requestID := getRequestID(c)
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error().Str("request_id", requestID).Msg("User ID not found in context")
		c.JSON(http.StatusUnauthorized, NewErrorResponse("Unauthorized", "user_id_missing", requestID))
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		h.logger.Error().Str("request_id", requestID).Msg("Invalid user ID type in context")
		c.JSON(http.StatusUnauthorized, NewErrorResponse("Unauthorized", "invalid_user_id", requestID))
		return
	}

	logger := h.logger.With().Str("request_id", requestID).Str("user_id", userUUID.String()).Str("method", "GetUserSettings").Logger()

	settings, err := h.settingsService.GetUserSettings(ctx, userUUID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get user settings")
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to get user settings", err.Error(), requestID))
		return
	}

	logger.Info().Msg("User settings retrieved successfully")
	c.JSON(http.StatusOK, NewSuccessResponse("User settings retrieved successfully", settings, requestID))
}

// UpdateUserSettings handles PUT /api/v1/settings
func (h *UserSettingsHandler) UpdateUserSettings(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	requestID := getRequestID(c)
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error().Str("request_id", requestID).Msg("User ID not found in context")
		c.JSON(http.StatusUnauthorized, NewErrorResponse("Unauthorized", "user_id_missing", requestID))
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		h.logger.Error().Str("request_id", requestID).Msg("Invalid user ID type in context")
		c.JSON(http.StatusUnauthorized, NewErrorResponse("Unauthorized", "invalid_user_id", requestID))
		return
	}

	logger := h.logger.With().Str("request_id", requestID).Str("user_id", userUUID.String()).Str("method", "UpdateUserSettings").Logger()

	var req models.UpdateUserSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn().Err(err).Msg("Invalid request body")
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request body", err.Error(), requestID))
		return
	}

	settings, err := h.settingsService.UpdateUserSettings(ctx, userUUID, &req)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to update user settings")
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to update user settings", err.Error(), requestID))
		return
	}

	logger.Info().Msg("User settings updated successfully")
	c.JSON(http.StatusOK, NewSuccessResponse("User settings updated successfully", settings, requestID))
}

