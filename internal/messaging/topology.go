package messaging

import (
	"fmt"

	"pay-your-dues/internal/config"

	amqp "github.com/rabbitmq/amqp091-go"
)

// DeclareTopology creates the exchange, queues, bindings, and dead-letter setup.
func DeclareTopology(ch *amqp.Channel, cfg *config.RabbitMQConfig) error {
	if err := ch.ExchangeDeclare(
		cfg.Exchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("declare exchange: %w", err)
	}

	dlqArgs := amqp.Table{}
	if _, err := ch.QueueDeclare(
		cfg.DLQ,
		true,
		false,
		false,
		false,
		dlqArgs,
	); err != nil {
		return fmt.Errorf("declare dlq: %w", err)
	}

	if err := ch.QueueBind(cfg.DLQ, RoutingKeyDLQ, cfg.Exchange, false, nil); err != nil {
		return fmt.Errorf("bind dlq: %w", err)
	}

	retryArgs := amqp.Table{
		"x-message-ttl":             int32(cfg.RetryTTL.Milliseconds()),
		"x-dead-letter-exchange":    cfg.Exchange,
		"x-dead-letter-routing-key": RoutingKeyScheduled,
	}
	if _, err := ch.QueueDeclare(
		cfg.RetryQueue,
		true,
		false,
		false,
		false,
		retryArgs,
	); err != nil {
		return fmt.Errorf("declare retry queue: %w", err)
	}

	if err := ch.QueueBind(cfg.RetryQueue, RoutingKeyRetry, cfg.Exchange, false, nil); err != nil {
		return fmt.Errorf("bind retry queue: %w", err)
	}

	sendArgs := amqp.Table{
		"x-dead-letter-exchange":    cfg.Exchange,
		"x-dead-letter-routing-key": RoutingKeyRetry,
	}
	if _, err := ch.QueueDeclare(
		cfg.SendQueue,
		true,
		false,
		false,
		false,
		sendArgs,
	); err != nil {
		return fmt.Errorf("declare send queue: %w", err)
	}

	for _, key := range []string{RoutingKeyScheduled, RoutingKeyImmediate, RoutingKeyManual, RoutingKeyEvent} {
		if err := ch.QueueBind(cfg.SendQueue, key, cfg.Exchange, false, nil); err != nil {
			return fmt.Errorf("bind send queue (%s): %w", key, err)
		}
	}

	return nil
}
