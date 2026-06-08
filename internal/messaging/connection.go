package messaging

import (
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Connection manages a reconnecting RabbitMQ connection and channel.
type Connection struct {
	url string

	mu          sync.Mutex
	conn        *amqp.Connection
	channel     *amqp.Channel
	notifyClose chan *amqp.Error
}

func NewConnection(url string) *Connection {
	return &Connection{url: url}
}

func (c *Connection) Channel() (*amqp.Channel, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.channel != nil && c.conn != nil && !c.conn.IsClosed() {
		return c.channel, nil
	}

	if err := c.connectLocked(); err != nil {
		return nil, err
	}
	return c.channel, nil
}

func (c *Connection) connectLocked() error {
	if c.conn != nil && !c.conn.IsClosed() {
		_ = c.conn.Close()
	}

	conn, err := amqp.Dial(c.url)
	if err != nil {
		return fmt.Errorf("dial rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("open rabbitmq channel: %w", err)
	}

	c.conn = conn
	c.channel = ch
	c.notifyClose = make(chan *amqp.Error, 1)
	c.conn.NotifyClose(c.notifyClose)

	go c.watchClose()

	return nil
}

func (c *Connection) watchClose() {
	c.mu.Lock()
	notify := c.notifyClose
	c.mu.Unlock()

	if notify == nil {
		return
	}

	err, ok := <-notify
	if !ok || err == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.channel = nil
	c.conn = nil
}

func (c *Connection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.channel != nil {
		_ = c.channel.Close()
		c.channel = nil
	}
	if c.conn != nil && !c.conn.IsClosed() {
		return c.conn.Close()
	}
	return nil
}

func (c *Connection) WaitForReady(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := c.Channel(); err == nil {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("rabbitmq not ready after %s", timeout)
}
