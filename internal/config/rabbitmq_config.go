package config

import (
	"strconv"
	"time"
)

// RabbitMQConfig holds RabbitMQ and notification worker tuning settings.
type RabbitMQConfig struct {
	URL      string
	Exchange string

	SendQueue       string
	RetryQueue      string
	DLQ             string
	RetryRoutingKey string
	DLQRoutingKey   string

	SchedulerInterval  time.Duration
	SchedulerBatchSize int
	SchedulerLockID    int64

	ConsumerPrefetch    int
	ConsumerConcurrency int
	RetryTTL            time.Duration

	WorkerPort string
}

func LoadRabbitMQConfig() *RabbitMQConfig {
	schedulerBatch, _ := strconv.Atoi(getEnv("SCHEDULER_BATCH_SIZE", "200"))
	consumerPrefetch, _ := strconv.Atoi(getEnv("CONSUMER_PREFETCH", "10"))
	consumerConcurrency, _ := strconv.Atoi(getEnv("CONSUMER_CONCURRENCY", "8"))
	schedulerLockID, _ := strconv.ParseInt(getEnv("SCHEDULER_LOCK_ID", "8675309"), 10, 64)

	schedulerInterval, _ := time.ParseDuration(getEnv("SCHEDULER_INTERVAL", "30s"))
	retryTTL, _ := time.ParseDuration(getEnv("NOTIFICATION_RETRY_TTL", "30s"))

	return &RabbitMQConfig{
		URL:                getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		Exchange:           getEnv("RABBITMQ_EXCHANGE", "notifications.topic"),
		SendQueue:          getEnv("RABBITMQ_SEND_QUEUE", "notifications.send"),
		RetryQueue:         getEnv("RABBITMQ_RETRY_QUEUE", "notifications.send.retry"),
		DLQ:                getEnv("RABBITMQ_DLQ", "notifications.send.dlq"),
		RetryRoutingKey:    getEnv("RABBITMQ_RETRY_ROUTING_KEY", "notification.send.retry"),
		DLQRoutingKey:      getEnv("RABBITMQ_DLQ_ROUTING_KEY", "notification.send.dlq"),
		SchedulerInterval:  schedulerInterval,
		SchedulerBatchSize: schedulerBatch,
		SchedulerLockID:    schedulerLockID,
		ConsumerPrefetch:   consumerPrefetch,
		ConsumerConcurrency: consumerConcurrency,
		RetryTTL:           retryTTL,
		WorkerPort:         getEnv("NOTIFICATION_WORKER_PORT", "8081"),
	}
}
