package unit

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pay-your-dues/internal/messaging"
)

func TestNotificationJob_RoutingKey(t *testing.T) {
	tests := []struct {
		jobType     string
		expectedKey string
	}{
		{messaging.JobTypeScheduled, messaging.RoutingKeyScheduled},
		{messaging.JobTypeImmediate, messaging.RoutingKeyImmediate},
		{messaging.JobTypeManual, messaging.RoutingKeyManual},
		{messaging.JobTypeEvent, messaging.RoutingKeyEvent},
		{"unknown", messaging.RoutingKeyScheduled},
	}

	for _, tt := range tests {
		job := messaging.NotificationJob{JobType: tt.jobType}
		assert.Equal(t, tt.expectedKey, job.RoutingKey())
	}
}

func TestNotificationJob_MarshalUnmarshal(t *testing.T) {
	id := uuid.New()
	job := messaging.NotificationJob{
		NotificationID: id,
		JobType:        messaging.JobTypeManual,
		Attempt:        2,
	}

	data, err := job.Marshal()
	require.NoError(t, err)

	decoded, err := messaging.UnmarshalNotificationJob(data)
	require.NoError(t, err)
	assert.Equal(t, id, decoded.NotificationID)
	assert.Equal(t, messaging.JobTypeManual, decoded.JobType)
	assert.Equal(t, 2, decoded.Attempt)
}

func TestUnmarshalNotificationJob_DefaultAttempt(t *testing.T) {
	id := uuid.New()
	data := []byte(`{"notification_id":"` + id.String() + `","job_type":"immediate"}`)

	decoded, err := messaging.UnmarshalNotificationJob(data)
	require.NoError(t, err)
	assert.Equal(t, 1, decoded.Attempt)
}

func TestNoOpPublisher(t *testing.T) {
	publisher := messaging.NewNoOpPublisher()
	err := publisher.PublishNotification(uuid.New(), messaging.JobTypeImmediate)
	assert.NoError(t, err)
	assert.NoError(t, publisher.Close())
}
