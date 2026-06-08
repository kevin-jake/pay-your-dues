package handlers

import (
	"net/http"
	"strconv"
	"time"

	"pay-your-dues/internal/domain/interfaces"
	"pay-your-dues/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// NotificationHandler handles notification-related HTTP requests
type NotificationHandler struct {
	notificationService interfaces.NotificationService
	logger              zerolog.Logger
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(
	notificationService interfaces.NotificationService,
	logger zerolog.Logger,
) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
		logger:              logger,
	}
}

// CreateNotification creates a new manual notification
// POST /api/v1/notifications
func (h *NotificationHandler) CreateNotification(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse("Unauthorized", "user_id_missing", getRequestID(c)))
		return
	}

	var req models.CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to bind create notification request")
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request body", "invalid_request", getRequestID(c)))
		return
	}

	notification, err := h.notificationService.CreateNotification(userID.(uuid.UUID), &req)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create notification")
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err.Error(), "create_failed", getRequestID(c)))
		return
	}

	h.logger.Info().
		Str("notification_id", notification.ID.String()).
		Str("user_id", userID.(uuid.UUID).String()).
		Msg("Notification created successfully")

	c.JSON(http.StatusCreated, NewSuccessResponse("Notification created successfully", notification, getRequestID(c)))
}

// ListNotifications lists notifications for the authenticated user
// GET /api/v1/notifications
func (h *NotificationHandler) ListNotifications(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse("Unauthorized", "user_id_missing", getRequestID(c)))
		return
	}

	status := c.Query("status")
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}

	var debtListID *uuid.UUID
	if debtListIDStr := c.Query("debt_list_id"); debtListIDStr != "" {
		parsed, err := uuid.Parse(debtListIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid debt list ID", "invalid_id", getRequestID(c)))
			return
		}
		debtListID = &parsed
	}

	notifications, err := h.notificationService.GetUserNotifications(userID.(uuid.UUID), status, debtListID, limit)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list notifications")
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err.Error(), "list_failed", getRequestID(c)))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Notifications retrieved successfully", notifications, getRequestID(c)))
}

// GetNotificationsByDebtList gets all notifications for a debt list
// GET /api/v1/notifications/debt-lists/:debt_list_id
func (h *NotificationHandler) GetNotificationsByDebtList(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse("Unauthorized", "user_id_missing", getRequestID(c)))
		return
	}

	debtListIDStr := c.Param("debt_list_id")
	debtListID, err := uuid.Parse(debtListIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid debt list ID", "invalid_id", getRequestID(c)))
		return
	}

	notifications, err := h.notificationService.GetNotificationsByDebtList(userID.(uuid.UUID), debtListID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get notifications")
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err.Error(), "get_failed", getRequestID(c)))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Notifications retrieved successfully", notifications, getRequestID(c)))
}

// GetNotificationsByInstallment gets notifications for a specific installment
// GET /api/v1/debts/:id/installments/:num/notifications
func (h *NotificationHandler) GetNotificationsByInstallment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse("Unauthorized", "user_id_missing", getRequestID(c)))
		return
	}

	debtListIDStr := c.Param("id")
	debtListID, err := uuid.Parse(debtListIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid debt list ID", "invalid_id", getRequestID(c)))
		return
	}

	installmentNumStr := c.Param("num")
	installmentNum, err := strconv.Atoi(installmentNumStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid installment number", "invalid_installment", getRequestID(c)))
		return
	}

	notifications, err := h.notificationService.GetNotificationsByInstallment(userID.(uuid.UUID), debtListID, installmentNum)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get installment notifications")
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err.Error(), "get_failed", getRequestID(c)))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Installment notifications retrieved successfully", notifications, getRequestID(c)))
}

// GetNotification gets a single notification by ID
// GET /api/v1/notifications/:id
func (h *NotificationHandler) GetNotification(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse("Unauthorized", "user_id_missing", getRequestID(c)))
		return
	}

	notificationIDStr := c.Param("id")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid notification ID", "invalid_id", getRequestID(c)))
		return
	}

	notification, err := h.notificationService.GetNotification(userID.(uuid.UUID), notificationID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get notification")
		c.JSON(http.StatusNotFound, NewErrorResponse(err.Error(), "not_found", getRequestID(c)))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("Notification retrieved successfully", notification, getRequestID(c)))
}

// EnableNotification enables a notification
// PATCH /api/v1/notifications/:id/enable
func (h *NotificationHandler) EnableNotification(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse("Unauthorized", "user_id_missing", getRequestID(c)))
		return
	}

	notificationIDStr := c.Param("id")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid notification ID", "invalid_id", getRequestID(c)))
		return
	}

	if err := h.notificationService.EnableNotification(userID.(uuid.UUID), notificationID); err != nil {
		h.logger.Error().Err(err).Msg("Failed to enable notification")
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err.Error(), "enable_failed", getRequestID(c)))
		return
	}

	h.logger.Info().
		Str("notification_id", notificationID.String()).
		Msg("Notification enabled successfully")

	c.JSON(http.StatusOK, NewSuccessResponse("Notification enabled successfully", nil, getRequestID(c)))
}

// DisableNotification disables a notification
// PATCH /api/v1/notifications/:id/disable
func (h *NotificationHandler) DisableNotification(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse("Unauthorized", "user_id_missing", getRequestID(c)))
		return
	}

	notificationIDStr := c.Param("id")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid notification ID", "invalid_id", getRequestID(c)))
		return
	}

	if err := h.notificationService.DisableNotification(userID.(uuid.UUID), notificationID); err != nil {
		h.logger.Error().Err(err).Msg("Failed to disable notification")
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err.Error(), "disable_failed", getRequestID(c)))
		return
	}

	h.logger.Info().
		Str("notification_id", notificationID.String()).
		Msg("Notification disabled successfully")

	c.JSON(http.StatusOK, NewSuccessResponse("Notification disabled successfully", nil, getRequestID(c)))
}

// DeleteNotification deletes a notification
// DELETE /api/v1/notifications/:id
func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse("Unauthorized", "user_id_missing", getRequestID(c)))
		return
	}

	notificationIDStr := c.Param("id")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid notification ID", "invalid_id", getRequestID(c)))
		return
	}

	if err := h.notificationService.DeleteNotification(userID.(uuid.UUID), notificationID); err != nil {
		h.logger.Error().Err(err).Msg("Failed to delete notification")
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err.Error(), "delete_failed", getRequestID(c)))
		return
	}

	h.logger.Info().
		Str("notification_id", notificationID.String()).
		Msg("Notification deleted successfully")

	c.JSON(http.StatusOK, NewSuccessResponse("Notification deleted successfully", nil, getRequestID(c)))
}

// ScheduleNotifications schedules notifications for a debt list
// POST /api/v1/notifications/schedule
func (h *NotificationHandler) ScheduleNotifications(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse("Unauthorized", "user_id_missing", getRequestID(c)))
		return
	}

	var req struct {
		DebtListID uuid.UUID `json:"debt_list_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to bind schedule notifications request")
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request body", "invalid_request", getRequestID(c)))
		return
	}

	if err := h.notificationService.CreateNotificationsForDebtList(userID.(uuid.UUID), req.DebtListID); err != nil {
		h.logger.Error().Err(err).Msg("Failed to schedule notifications")
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err.Error(), "schedule_failed", getRequestID(c)))
		return
	}

	h.logger.Info().
		Str("debt_list_id", req.DebtListID.String()).
		Msg("Notifications scheduled successfully")

	c.JSON(http.StatusOK, NewSuccessResponse("Notifications scheduled successfully", nil, getRequestID(c)))
}

// SendManualNotification sends a manual notification
// POST /api/v1/notifications/send
func (h *NotificationHandler) SendManualNotification(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse("Unauthorized", "user_id_missing", getRequestID(c)))
		return
	}

	var req struct {
		DebtListID       uuid.UUID `json:"debt_list_id" binding:"required"`
		Message          string    `json:"message" binding:"required"`
		NotificationType string    `json:"notification_type" binding:"required,oneof=email sms"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to bind send manual notification request")
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request body", "invalid_request", getRequestID(c)))
		return
	}

	if err := h.notificationService.SendManualNotification(userID.(uuid.UUID), req.DebtListID, req.Message, req.NotificationType); err != nil {
		h.logger.Error().Err(err).Msg("Failed to send manual notification")
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err.Error(), "send_failed", getRequestID(c)))
		return
	}

	h.logger.Info().
		Str("debt_list_id", req.DebtListID.String()).
		Msg("Manual notification sent successfully")

	c.JSON(http.StatusOK, NewSuccessResponse("Notification sent successfully", nil, getRequestID(c)))
}

// GetDebtNotificationSettings returns merged notification settings for a debt list
// GET /api/v1/notifications/debt-lists/:debt_list_id/settings
func (h *NotificationHandler) GetDebtNotificationSettings(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse("Unauthorized", "user_id_missing", getRequestID(c)))
		return
	}
	debtListID, err := uuid.Parse(c.Param("debt_list_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid debt list ID", "invalid_id", getRequestID(c)))
		return
	}
	result, err := h.notificationService.GetDebtNotificationSettings(userID.(uuid.UUID), debtListID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get debt notification settings")
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err.Error(), "get_failed", getRequestID(c)))
		return
	}
	c.JSON(http.StatusOK, NewSuccessResponse("Debt notification settings retrieved", result, getRequestID(c)))
}

// EnableDebtNotifications enables scheduled notifications for a debt list
// POST /api/v1/notifications/debt-lists/:debt_list_id/schedule
func (h *NotificationHandler) EnableDebtNotifications(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse("Unauthorized", "user_id_missing", getRequestID(c)))
		return
	}
	debtListID, err := uuid.Parse(c.Param("debt_list_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid debt list ID", "invalid_id", getRequestID(c)))
		return
	}
	var req struct {
		Settings *models.CustomNotificationSettings `json:"settings"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		// Body is optional — ignore binding error
		req.Settings = nil
	}
	if err := h.notificationService.EnableDebtNotifications(userID.(uuid.UUID), debtListID, req.Settings); err != nil {
		h.logger.Error().Err(err).Msg("Failed to enable debt notifications")
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err.Error(), "enable_failed", getRequestID(c)))
		return
	}
	c.JSON(http.StatusOK, NewSuccessResponse("Notifications enabled", nil, getRequestID(c)))
}

// DisableDebtNotifications disables and bulk-deletes all notifications for a debt list
// DELETE /api/v1/notifications/debt-lists/:debt_list_id/schedule
func (h *NotificationHandler) DisableDebtNotifications(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse("Unauthorized", "user_id_missing", getRequestID(c)))
		return
	}
	debtListID, err := uuid.Parse(c.Param("debt_list_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid debt list ID", "invalid_id", getRequestID(c)))
		return
	}
	if err := h.notificationService.DisableDebtNotifications(userID.(uuid.UUID), debtListID); err != nil {
		h.logger.Error().Err(err).Msg("Failed to disable debt notifications")
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err.Error(), "disable_failed", getRequestID(c)))
		return
	}
	c.JSON(http.StatusOK, NewSuccessResponse("Notifications disabled", nil, getRequestID(c)))
}

// DeleteDebtNotificationSlot bulk-deletes all channel rows for one (installment, scheduled_for) slot
// DELETE /api/v1/notifications/debt-lists/:debt_list_id/slots
func (h *NotificationHandler) DeleteDebtNotificationSlot(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse("Unauthorized", "user_id_missing", getRequestID(c)))
		return
	}
	debtListID, err := uuid.Parse(c.Param("debt_list_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid debt list ID", "invalid_id", getRequestID(c)))
		return
	}
	var req models.DeleteNotificationSlotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request body: scheduled_for is required", "invalid_request", getRequestID(c)))
		return
	}
	if req.ScheduledFor.IsZero() {
		c.JSON(http.StatusBadRequest, NewErrorResponse("scheduled_for is required", "invalid_request", getRequestID(c)))
		return
	}
	if err := h.notificationService.DeleteNotificationSlot(userID.(uuid.UUID), debtListID, req.InstallmentNumber, req.ScheduledFor); err != nil {
		h.logger.Error().Err(err).Msg("Failed to delete notification slot")
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err.Error(), "delete_failed", getRequestID(c)))
		return
	}
	c.JSON(http.StatusOK, NewSuccessResponse("Notification slot deleted", nil, getRequestID(c)))
}

// GetManualSendLimits returns per-channel manual send usage for a debt list
// GET /api/v1/notifications/debt-lists/:debt_list_id/send-limits
func (h *NotificationHandler) GetManualSendLimits(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse("Unauthorized", "user_id_missing", getRequestID(c)))
		return
	}
	debtListID, err := uuid.Parse(c.Param("debt_list_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid debt list ID", "invalid_id", getRequestID(c)))
		return
	}
	limits, err := h.notificationService.GetManualSendLimits(userID.(uuid.UUID), debtListID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get send limits")
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err.Error(), "get_failed", getRequestID(c)))
		return
	}
	c.JSON(http.StatusOK, NewSuccessResponse("Send limits retrieved", limits, getRequestID(c)))
}

// SendManualNotificationWithLimitCheck wraps SendManualNotification and returns 422 on limit exceeded
// POST /api/v1/notifications/send
func (h *NotificationHandler) SendManualNotificationHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse("Unauthorized", "user_id_missing", getRequestID(c)))
		return
	}
	var req struct {
		DebtListID       uuid.UUID `json:"debt_list_id" binding:"required"`
		Message          string    `json:"message" binding:"required"`
		NotificationType string    `json:"notification_type" binding:"required,oneof=email sms"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to bind send manual notification request")
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid request body", "invalid_request", getRequestID(c)))
		return
	}
	if err := h.notificationService.SendManualNotification(userID.(uuid.UUID), req.DebtListID, req.Message, req.NotificationType); err != nil {
		if isLimitError(err) {
			c.JSON(http.StatusUnprocessableEntity, NewErrorResponse(err.Error(), "limit_exceeded", getRequestID(c)))
			return
		}
		h.logger.Error().Err(err).Msg("Failed to send manual notification")
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err.Error(), "send_failed", getRequestID(c)))
		return
	}
	h.logger.Info().Str("debt_list_id", req.DebtListID.String()).Msg("Manual notification sent successfully")
	c.JSON(http.StatusOK, NewSuccessResponse("Notification sent successfully", nil, getRequestID(c)))
}

func isLimitError(err error) bool {
	return err != nil && len(err.Error()) > 0 && containsString(err.Error(), "manual send limit reached")
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && stringContains(s, substr))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// keep time import used
var _ = time.Now

