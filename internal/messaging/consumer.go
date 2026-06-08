package messaging

import (
	"context"
	"fmt"
	"sync"
	"time"

	"pay-your-dues/internal/config"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
)

// DeliveryHandler processes a notification job pulled from the queue.
type DeliveryHandler func(ctx context.Context, job NotificationJob) error

// Consumer consumes notification jobs from RabbitMQ.
type Consumer struct {
	conn    *Connection
	cfg     *config.RabbitMQConfig
	handler DeliveryHandler
	logger  zerolog.Logger

	wg       sync.WaitGroup
	stopChan chan struct{}
}

func NewConsumer(cfg *config.RabbitMQConfig, handler DeliveryHandler, logger zerolog.Logger) (*Consumer, error) {
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

	return &Consumer{
		conn:     conn,
		cfg:      cfg,
		handler:  handler,
		logger:   logger.With().Str("component", "rabbitmq_consumer").Logger(),
		stopChan: make(chan struct{}),
	}, nil
}

func (c *Consumer) Start() error {
	for i := 0; i < c.cfg.ConsumerConcurrency; i++ {
		c.wg.Add(1)
		go c.runWorker(i)
	}
	return nil
}

func (c *Consumer) runWorker(workerID int) {
	defer c.wg.Done()

	logger := c.logger.With().Int("worker_id", workerID).Logger()

	for {
		select {
		case <-c.stopChan:
			return
		default:
		}

		if err := c.consumeOnce(logger); err != nil {
			logger.Error().Err(err).Msg("Consumer iteration failed, retrying")
		}
	}
}

func (c *Consumer) consumeOnce(logger zerolog.Logger) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}

	if err := ch.Qos(c.cfg.ConsumerPrefetch, 0, false); err != nil {
		return fmt.Errorf("set qos: %w", err)
	}

	deliveries, err := ch.Consume(
		c.cfg.SendQueue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("consume queue: %w", err)
	}

	for {
		select {
		case <-c.stopChan:
			return nil
		case delivery, ok := <-deliveries:
			if !ok {
				return fmt.Errorf("delivery channel closed")
			}
			c.handleDelivery(logger, delivery)
		}
	}
}

func (c *Consumer) handleDelivery(logger zerolog.Logger, delivery amqp.Delivery) {
	job, err := UnmarshalNotificationJob(delivery.Body)
	if err != nil {
		logger.Error().Err(err).Msg("Invalid job payload, sending to DLQ")
		_ = delivery.Nack(false, false)
		return
	}

	if attempt, ok := delivery.Headers["x-attempt"].(int32); ok && job.Attempt <= 1 {
		job.Attempt = int(attempt)
	}

	ctx := context.Background()
	if err := c.handler(ctx, job); err != nil {
		logger.Error().
			Err(err).
			Str("notification_id", job.NotificationID.String()).
			Int("attempt", job.Attempt).
			Msg("Failed to process notification job")
		_ = delivery.Nack(false, false)
		return
	}

	if err := delivery.Ack(false); err != nil {
		logger.Error().Err(err).Msg("Failed to ACK delivery")
	}
}

func (c *Consumer) Stop() {
	close(c.stopChan)
	c.wg.Wait()
	_ = c.conn.Close()
}
