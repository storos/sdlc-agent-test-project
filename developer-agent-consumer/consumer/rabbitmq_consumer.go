package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"github.com/storos/sdlc-agent/developer-agent-consumer/models"
)

const (
	maxRetries     = 5
	retryDelay     = 2 * time.Second
	prefetchCount  = 1
	queueName      = "develop"
	errorQueueName = "develop_error"
	exchangeName   = "webhook.development.request"
)

type MessageHandler func(context.Context, *models.DevelopmentRequest) error

type RabbitMQConsumer struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	logger       *logrus.Logger
	handler      MessageHandler
	done         chan struct{}
}

func NewRabbitMQConsumer(rabbitMQURL string, handler MessageHandler, logger *logrus.Logger) (*RabbitMQConsumer, error) {
	var conn *amqp.Connection
	var err error

	// Retry connection with exponential backoff
	for i := 0; i < maxRetries; i++ {
		conn, err = amqp.Dial(rabbitMQURL)
		if err == nil {
			break
		}
		logger.Warnf("Failed to connect to RabbitMQ (attempt %d/%d): %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			time.Sleep(retryDelay * time.Duration(i+1))
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ after %d attempts: %w", maxRetries, err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Set QoS (prefetch count)
	if err := channel.Qos(prefetchCount, 0, false); err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	consumer := &RabbitMQConsumer{
		conn:    conn,
		channel: channel,
		logger:  logger,
		handler: handler,
		done:    make(chan struct{}),
	}

	// Declare queues and bindings
	if err := consumer.setupQueues(); err != nil {
		consumer.Close()
		return nil, err
	}

	return consumer, nil
}

func (c *RabbitMQConsumer) setupQueues() error {
	// Declare main queue
	_, err := c.channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue %s: %w", queueName, err)
	}

	// Declare error queue
	_, err = c.channel.QueueDeclare(
		errorQueueName, // name
		true,           // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare error queue %s: %w", errorQueueName, err)
	}

	// Bind main queue to exchange with wildcard routing key
	err = c.channel.QueueBind(
		queueName,                  // queue name
		"webhook.development.*",    // routing key pattern (matches all projects)
		exchangeName,               // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue to exchange: %w", err)
	}

	c.logger.Infof("Queues declared and bound: %s, %s", queueName, errorQueueName)
	return nil
}

func (c *RabbitMQConsumer) Start(ctx context.Context) error {
	msgs, err := c.channel.Consume(
		queueName, // queue
		"",        // consumer tag
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	c.logger.Info("RabbitMQ consumer started, waiting for messages...")

	go func() {
		for {
			select {
			case <-ctx.Done():
				c.logger.Info("Context cancelled, stopping consumer")
				close(c.done)
				return
			case msg, ok := <-msgs:
				if !ok {
					c.logger.Warn("Message channel closed")
					close(c.done)
					return
				}
				c.processMessage(ctx, msg)
			}
		}
	}()

	return nil
}

func (c *RabbitMQConsumer) processMessage(ctx context.Context, msg amqp.Delivery) {
	c.logger.WithFields(logrus.Fields{
		"message_id": msg.MessageId,
		"body_size":  len(msg.Body),
	}).Info("Received message")

	var request models.DevelopmentRequest
	if err := json.Unmarshal(msg.Body, &request); err != nil {
		c.logger.Errorf("Failed to unmarshal message: %v", err)
		c.sendToErrorQueue(msg.Body, fmt.Sprintf("Invalid JSON: %v", err))
		msg.Nack(false, false)
		return
	}

	// Process the message with handler
	if err := c.handler(ctx, &request); err != nil {
		c.logger.WithFields(logrus.Fields{
			"jira_issue_key": request.JiraIssueKey,
			"error":          err.Error(),
		}).Error("Failed to process message")
		c.sendToErrorQueue(msg.Body, err.Error())
		msg.Nack(false, false)
		return
	}

	// Acknowledge successful processing
	if err := msg.Ack(false); err != nil {
		c.logger.Errorf("Failed to acknowledge message: %v", err)
	}

	c.logger.WithFields(logrus.Fields{
		"jira_issue_key": request.JiraIssueKey,
	}).Info("Message processed successfully")
}

func (c *RabbitMQConsumer) sendToErrorQueue(body []byte, errorMsg string) {
	errorMessage := map[string]interface{}{
		"original_message": string(body),
		"error":            errorMsg,
		"timestamp":        time.Now().Format(time.RFC3339),
	}

	errorBody, err := json.Marshal(errorMessage)
	if err != nil {
		c.logger.Errorf("Failed to marshal error message: %v", err)
		return
	}

	err = c.channel.Publish(
		"",             // exchange
		errorQueueName, // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        errorBody,
		},
	)
	if err != nil {
		c.logger.Errorf("Failed to publish to error queue: %v", err)
	}
}

func (c *RabbitMQConsumer) Wait() {
	<-c.done
}

func (c *RabbitMQConsumer) Close() error {
	c.logger.Info("Closing RabbitMQ consumer")
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
	return nil
}
