package interfaces

import "github.com/google/uuid"

// NotificationPublisher publishes notification delivery jobs to a message queue.
type NotificationPublisher interface {
	PublishNotification(notificationID uuid.UUID, jobType string) error
	Close() error
}
