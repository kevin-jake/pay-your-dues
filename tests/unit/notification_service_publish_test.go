package unit

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"pay-your-dues/internal/messaging"
	"pay-your-dues/internal/mocks"
	"pay-your-dues/internal/models"
	"pay-your-dues/internal/services"
)

func TestNotificationService_SendManualNotification_PublishesJob(t *testing.T) {
	userID := uuid.New()
	debtListID := uuid.New()

	notificationRepo := &mockNotificationRepo{}
	debtListRepo := &mocks.MockDebtListRepository{}
	publisher := &mocks.MockNotificationPublisher{}
	logger := zerolog.Nop()

	debtListRepo.On("BelongsToUser", mock.Anything, debtListID, userID).Return(true, nil)
	// no sends yet
	notificationRepo.On("CountManualNotificationsByDebtAndType", debtListID, "email").Return(int64(0), nil)
	publisher.On("PublishNotification", mock.AnythingOfType("uuid.UUID"), messaging.JobTypeManual).Return(nil)

	service := services.NewNotificationService(
		notificationRepo,
		nil,
		debtListRepo,
		nil,
		publisher,
		logger,
	)

	err := service.SendManualNotification(userID, debtListID, "hello", "email")
	assert.NoError(t, err)
	publisher.AssertExpectations(t)
}

type mockNotificationRepo struct {
	mock.Mock
	created *models.Notification
}

func (m *mockNotificationRepo) Create(notification *models.Notification) error {
	m.created = notification
	return nil
}

func (m *mockNotificationRepo) GetByID(id uuid.UUID) (*models.Notification, error) {
	if m.created != nil && m.created.ID == id {
		return m.created, nil
	}
	return nil, assert.AnError
}

func (m *mockNotificationRepo) Update(notification *models.Notification) error { return nil }
func (m *mockNotificationRepo) Delete(id uuid.UUID) error                      { return nil }
func (m *mockNotificationRepo) GetByDebtListID(debtListID uuid.UUID) ([]*models.Notification, error) {
	return nil, nil
}
func (m *mockNotificationRepo) GetByDebtListAndInstallment(debtListID uuid.UUID, installmentNumber int) ([]*models.Notification, error) {
	return nil, nil
}
func (m *mockNotificationRepo) GetPendingNotifications(limit int) ([]*models.Notification, error) {
	return nil, nil
}
func (m *mockNotificationRepo) GetScheduledNotifications(beforeTime time.Time, limit int) ([]*models.Notification, error) {
	return nil, nil
}
func (m *mockNotificationRepo) GetByStatus(status string, limit int) ([]*models.Notification, error) {
	return nil, nil
}
func (m *mockNotificationRepo) CreateBatch(notifications []*models.Notification) error { return nil }
func (m *mockNotificationRepo) UpdateStatus(id uuid.UUID, status string, sentAt *time.Time) error {
	return nil
}
func (m *mockNotificationRepo) UpdateNextRun(id uuid.UUID, nextRunAt time.Time) error { return nil }
func (m *mockNotificationRepo) DisableNotification(id uuid.UUID) error                { return nil }
func (m *mockNotificationRepo) EnableNotification(id uuid.UUID) error                   { return nil }
func (m *mockNotificationRepo) DisableNotificationsForInstallment(debtListID uuid.UUID, installmentNumber int) error {
	return nil
}
func (m *mockNotificationRepo) GetActiveNotificationsForInstallment(debtListID uuid.UUID, installmentNumber int) ([]*models.Notification, error) {
	return nil, nil
}
func (m *mockNotificationRepo) HasPendingNotifications(debtListID uuid.UUID, installmentNumber *int) (bool, error) {
	return false, nil
}
func (m *mockNotificationRepo) CountNotificationsByDebtList(debtListID uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *mockNotificationRepo) GetPendingNotificationsByUserID(userID uuid.UUID) ([]*models.Notification, error) {
	return nil, nil
}
func (m *mockNotificationRepo) GetUserNotifications(userID uuid.UUID, status string, debtListID *uuid.UUID, limit int) ([]*models.Notification, error) {
	return nil, nil
}
func (m *mockNotificationRepo) UpdatePendingNotificationsByUserID(userID uuid.UUID, updates map[string]interface{}) error {
	return nil
}
func (m *mockNotificationRepo) ClaimDueNotifications(beforeTime time.Time, limit int) ([]*models.Notification, error) {
	return nil, nil
}
func (m *mockNotificationRepo) RevertToPending(id uuid.UUID) error { return nil }
func (m *mockNotificationRepo) MarkQueued(id uuid.UUID) error      { return nil }
func (m *mockNotificationRepo) BatchSetEnabled(ids []uuid.UUID, enabled bool) error {
	return nil
}
func (m *mockNotificationRepo) DeleteByDebtListID(debtListID uuid.UUID) error { return nil }
func (m *mockNotificationRepo) DeleteByDebtListIDAndSlot(debtListID uuid.UUID, installmentNumber *int, scheduledFor time.Time) error {
	return nil
}
func (m *mockNotificationRepo) DeleteReminderNotificationsByDebtList(debtListID uuid.UUID) error {
	return nil
}
func (m *mockNotificationRepo) CountManualNotificationsByDebtAndType(debtListID uuid.UUID, notificationType string) (int64, error) {
	args := m.Called(debtListID, notificationType)
	return args.Get(0).(int64), args.Error(1)
}
