package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockNotificationPublisher is a mock implementation of NotificationPublisher.
type MockNotificationPublisher struct {
	mock.Mock
}

func (m *MockNotificationPublisher) PublishNotification(notificationID uuid.UUID, jobType string) error {
	args := m.Called(notificationID, jobType)
	return args.Error(0)
}

func (m *MockNotificationPublisher) Close() error {
	args := m.Called()
	return args.Error(0)
}
