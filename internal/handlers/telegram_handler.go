package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"pay-your-dues/internal/services"
)

// TelegramHandler handles Telegram webhook and subscription endpoints
type TelegramHandler struct {
	telegramService *services.TelegramSubscriptionService
	webhookSecret   string
	logger          zerolog.Logger
}

// NewTelegramHandler creates a new Telegram handler
func NewTelegramHandler(
	telegramService *services.TelegramSubscriptionService,
	webhookSecret string,
	logger zerolog.Logger,
) *TelegramHandler {
	return &TelegramHandler{
		telegramService: telegramService,
		webhookSecret:   webhookSecret,
		logger:          logger.With().Str("handler", "telegram").Logger(),
	}
}

// HandleWebhook processes incoming Telegram webhook updates
// POST /api/telegram/webhook
func (h *TelegramHandler) HandleWebhook(c *gin.Context) {
	// Verify webhook secret from header if configured
	// Telegram sends the secret in X-Telegram-Bot-Api-Secret-Token header
	if h.webhookSecret != "" {
		token := c.GetHeader("X-Telegram-Bot-Api-Secret-Token")
		if token != h.webhookSecret {
			h.logger.Warn().Msg("Invalid Telegram webhook secret")
			c.Status(http.StatusUnauthorized)
			return
		}
	}

	var update services.TelegramUpdate
	if err := c.ShouldBindJSON(&update); err != nil {
		h.logger.Error().Err(err).Msg("Failed to parse Telegram update")
		c.Status(http.StatusBadRequest)
		return
	}

	if err := h.telegramService.HandleWebhookUpdate(c.Request.Context(), &update); err != nil {
		h.logger.Error().Err(err).Msg("Failed to handle Telegram update")
		// Still return 200 to prevent Telegram from retrying
	}

	// Always return 200 OK to Telegram
	c.Status(http.StatusOK)
}

// GenerateLinkCodeResponse represents the response for link code generation
type GenerateLinkCodeResponse struct {
	Code         string `json:"code"`
	BotLink      string `json:"bot_link"`
	BotUsername  string `json:"bot_username,omitempty"`
	Instructions string `json:"instructions"`
	ExpiresIn    int    `json:"expires_in"` // seconds
}

// GenerateLinkCode generates a code for user to link their Telegram account
// POST /api/v1/settings/telegram/link
func (h *TelegramHandler) GenerateLinkCode(c *gin.Context) {
	requestID := getRequestID(c)

	// Check if Telegram is configured
	if !h.telegramService.IsConfigured() {
		c.JSON(http.StatusServiceUnavailable, NewErrorResponse(
			"Telegram notifications are not configured",
			"telegram_not_configured",
			requestID,
		))
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse("Unauthorized", "user_id_missing", requestID))
		return
	}

	code, err := h.telegramService.GenerateLinkCode(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to generate Telegram link code")
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err.Error(), "generate_code_failed", requestID))
		return
	}

	// Get bot info for the link
	botUsername := ""
	botLink := ""
	if botInfo, err := h.telegramService.GetBotInfo(); err == nil {
		botUsername = botInfo.UserName
		botLink = "https://t.me/" + botInfo.UserName
	}

	response := GenerateLinkCodeResponse{
		Code:         code,
		BotLink:      botLink,
		BotUsername:  botUsername,
		Instructions: "Open the Telegram bot and send this code to link your account",
		ExpiresIn:    600, // 10 minutes
	}

	h.logger.Info().
		Str("user_id", userID.(uuid.UUID).String()).
		Str("code", code).
		Msg("Generated Telegram link code")

	c.JSON(http.StatusOK, NewSuccessResponse("Link code generated successfully", response, requestID))
}

// UnlinkTelegram removes Telegram subscription for the user
// DELETE /api/v1/settings/telegram/link
func (h *TelegramHandler) UnlinkTelegram(c *gin.Context) {
	requestID := getRequestID(c)

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse("Unauthorized", "user_id_missing", requestID))
		return
	}

	if err := h.telegramService.UnlinkAccount(c.Request.Context(), userID.(uuid.UUID)); err != nil {
		h.logger.Error().Err(err).Msg("Failed to unlink Telegram account")
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err.Error(), "unlink_failed", requestID))
		return
	}

	h.logger.Info().
		Str("user_id", userID.(uuid.UUID).String()).
		Msg("Telegram account unlinked successfully")

	c.JSON(http.StatusOK, NewSuccessResponse("Telegram account unlinked successfully", nil, requestID))
}

// GetTelegramStatus returns the Telegram subscription status for the user
// GET /api/v1/settings/telegram/status
func (h *TelegramHandler) GetTelegramStatus(c *gin.Context) {
	requestID := getRequestID(c)

	// Check if Telegram is configured at app level
	isConfigured := h.telegramService.IsConfigured()

	// Get bot info if configured
	var botUsername string
	var botLink string
	if isConfigured {
		if botInfo, err := h.telegramService.GetBotInfo(); err == nil {
			botUsername = botInfo.UserName
			botLink = "https://t.me/" + botInfo.UserName
		}
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Telegram status retrieved", gin.H{
		"configured":   isConfigured,
		"bot_username": botUsername,
		"bot_link":     botLink,
	}, requestID))
}

