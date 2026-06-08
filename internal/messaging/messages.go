package messaging

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

const (
	JobTypeScheduled = "scheduled"
	JobTypeImmediate = "immediate"
	JobTypeManual    = "manual"
	JobTypeEvent     = "event"

	RoutingKeyScheduled = "notification.send.scheduled"
	RoutingKeyImmediate = "notification.send.immediate"
	RoutingKeyManual    = "notification.send.manual"
	RoutingKeyEvent     = "notification.send.event"
	RoutingKeyRetry     = "notification.send.retry"
	RoutingKeyDLQ       = "notification.send.dlq"
)

// NotificationJob is the thin payload published to RabbitMQ.
type NotificationJob struct {
	NotificationID uuid.UUID `json:"notification_id"`
	JobType        string    `json:"job_type"`
	Attempt        int       `json:"attempt"`
}

func (j NotificationJob) RoutingKey() string {
	switch j.JobType {
	case JobTypeImmediate:
		return RoutingKeyImmediate
	case JobTypeManual:
		return RoutingKeyManual
	case JobTypeEvent:
		return RoutingKeyEvent
	default:
		return RoutingKeyScheduled
	}
}

func (j NotificationJob) Marshal() ([]byte, error) {
	data, err := json.Marshal(j)
	if err != nil {
		return nil, fmt.Errorf("marshal notification job: %w", err)
	}
	return data, nil
}

func UnmarshalNotificationJob(data []byte) (NotificationJob, error) {
	var job NotificationJob
	if err := json.Unmarshal(data, &job); err != nil {
		return NotificationJob{}, fmt.Errorf("unmarshal notification job: %w", err)
	}
	if job.NotificationID == uuid.Nil {
		return NotificationJob{}, fmt.Errorf("notification_id is required")
	}
	if job.Attempt <= 0 {
		job.Attempt = 1
	}
	return job, nil
}
