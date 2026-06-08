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
	"pay-your-dues/internal/services"
)

// --------------------------------------------------------------------------
// Send Now limit tests
// --------------------------------------------------------------------------

func TestSendManualNotification_LimitNotReached(t *testing.T) {
	userID := uuid.New()
	debtListID := uuid.New()

	notifRepo := &mockNotificationRepo{}
	debtListRepo := &mocks.MockDebtListRepository{}
	publisher := &mocks.MockNotificationPublisher{}
	logger := zerolog.Nop()

	debtListRepo.On("BelongsToUser", mock.Anything, debtListID, userID).Return(true, nil)
	// 2 sends used — limit (3) not reached
	notifRepo.On("CountManualNotificationsByDebtAndType", debtListID, "email").Return(int64(2), nil)
	publisher.On("PublishNotification", mock.AnythingOfType("uuid.UUID"), messaging.JobTypeManual).Return(nil)

	svc := services.NewNotificationService(
		notifRepo,
		nil,
		debtListRepo,
		nil,
		publisher,
		logger,
	)

	err := svc.SendManualNotification(userID, debtListID, "hello", "email")
	assert.NoError(t, err)
	publisher.AssertExpectations(t)
}

func TestSendManualNotification_LimitReached(t *testing.T) {
	userID := uuid.New()
	debtListID := uuid.New()

	notifRepo := &mockNotificationRepo{}
	debtListRepo := &mocks.MockDebtListRepository{}
	publisher := &mocks.MockNotificationPublisher{}
	logger := zerolog.Nop()

	debtListRepo.On("BelongsToUser", mock.Anything, debtListID, userID).Return(true, nil)
	// 3 sends already used — at limit
	notifRepo.On("CountManualNotificationsByDebtAndType", debtListID, "email").Return(int64(3), nil)

	svc := services.NewNotificationService(
		notifRepo,
		nil,
		debtListRepo,
		nil,
		publisher,
		logger,
	)

	err := svc.SendManualNotification(userID, debtListID, "hello", "email")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "manual send limit reached")
	// publisher must NOT be called
	publisher.AssertNotCalled(t, "PublishNotification", mock.Anything, mock.Anything)
}

func TestSendManualNotification_SMSLimitIndependentOfEmail(t *testing.T) {
	userID := uuid.New()
	debtListID := uuid.New()

	notifRepo := &mockNotificationRepo{}
	debtListRepo := &mocks.MockDebtListRepository{}
	publisher := &mocks.MockNotificationPublisher{}
	logger := zerolog.Nop()

	debtListRepo.On("BelongsToUser", mock.Anything, debtListID, userID).Return(true, nil)
	// SMS has 0 sends; email at limit but we're sending SMS
	notifRepo.On("CountManualNotificationsByDebtAndType", debtListID, "sms").Return(int64(0), nil)
	publisher.On("PublishNotification", mock.AnythingOfType("uuid.UUID"), messaging.JobTypeManual).Return(nil)

	svc := services.NewNotificationService(
		notifRepo,
		nil,
		debtListRepo,
		nil,
		publisher,
		logger,
	)

	err := svc.SendManualNotification(userID, debtListID, "hello", "sms")
	assert.NoError(t, err)
	publisher.AssertExpectations(t)
}

// --------------------------------------------------------------------------
// Disable debt notifications test
// --------------------------------------------------------------------------

func TestDisableDebtNotifications_DeletesAndSetsFlag(t *testing.T) {
	userID := uuid.New()
	debtListID := uuid.New()

	notifRepo := &mockNotificationRepo{}
	debtListRepo := &mocks.MockDebtListRepository{}
	logger := zerolog.Nop()

	debtListRepo.On("BelongsToUser", mock.Anything, debtListID, userID).Return(true, nil)
	debtListRepo.On("UpdateNotificationSettings", mock.Anything, debtListID, mock.MatchedBy(func(m map[string]interface{}) bool {
		v, ok := m["notifications_enabled"]
		return ok && v == false
	})).Return(nil)

	svc := services.NewNotificationService(
		notifRepo,
		nil,
		debtListRepo,
		nil,
		nil,
		logger,
	)

	err := svc.DisableDebtNotifications(userID, debtListID)
	assert.NoError(t, err)
	debtListRepo.AssertExpectations(t)
}

// keep time import used
var _ = time.Now
