package messaging

import (
	"context"
	"time"

	"pay-your-dues/internal/config"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// RabbitMQPublisher publishes notification jobs to RabbitMQ.
type RabbitMQPublisher struct {
	conn   *Connection
	cfg    *config.RabbitMQConfig
	logger zerolog.Logger
}

func NewRabbitMQPublisher(cfg *config.RabbitMQConfig, logger zerolog.Logger) (*RabbitMQPublisher, error) {
	conn := NewConnection(cfg.URL)
	if err := conn.WaitForReady(15 * time.Second); err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	if err := DeclareTopology(ch, cfg); err != nil {
		return nil, err
	}

	return &RabbitMQPublisher{
		conn:   conn,
		cfg:    cfg,
		logger: logger.With().Str("component", "rabbitmq_publisher").Logger(),
	}, nil
}

// PublishNotification implements interfaces.NotificationPublisher.
func (p *RabbitMQPublisher) PublishNotification(notificationID uuid.UUID, jobType string) error {
	return p.Publish(NotificationJob{
		NotificationID: notificationID,
		JobType:        jobType,
		Attempt:        1,
	})
}

func (p *RabbitMQPublisher) Publish(job NotificationJob) error {
	body, err := job.Marshal()
	if err != nil {
		return err
	}

	ch, err := p.conn.Channel()
	if err != nil {
		return err
	}

	ctx := context.Background()
	return ch.PublishWithContext(ctx, p.cfg.Exchange, job.RoutingKey(), false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
		Headers: amqp.Table{
			"x-attempt": job.Attempt,
		},
	})
}

func (p *RabbitMQPublisher) PublishToDLQ(job NotificationJob) error {
	body, err := job.Marshal()
	if err != nil {
		return err
	}

	ch, err := p.conn.Channel()
	if err != nil {
		return err
	}

	ctx := context.Background()
	return ch.PublishWithContext(ctx, p.cfg.Exchange, RoutingKeyDLQ, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	})
}

func (p *RabbitMQPublisher) Close() error {
	return p.conn.Close()
}

// NoOpPublisher discards publish calls (useful for tests or when RabbitMQ is disabled).
type NoOpPublisher struct{}

func NewNoOpPublisher() *NoOpPublisher {
	return &NoOpPublisher{}
}

func (p *NoOpPublisher) Publish(_ NotificationJob) error { return nil }

func (p *NoOpPublisher) PublishNotification(_ uuid.UUID, _ string) error { return nil }

func (p *NoOpPublisher) Close() error { return nil }
