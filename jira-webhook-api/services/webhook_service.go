package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/storos/sdlc-agent/jira-webhook-api/models"
	"github.com/storos/sdlc-agent/jira-webhook-api/repositories"
	"github.com/streadway/amqp"
)

var (
	ErrInvalidPayload       = errors.New("invalid webhook payload")
	ErrNotInDevelopment     = errors.New("status change is not 'In Development'")
	ErrRabbitMQNotConnected = errors.New("RabbitMQ connection not available")
)

// WebhookService handles webhook processing
type WebhookService struct {
	repo           *repositories.WebhookRepository
	rabbitConn     *amqp.Connection
	rabbitChannel  *amqp.Channel
	exchangeName   string
	logger         *logrus.Logger
}

// NewWebhookService creates a new webhook service
func NewWebhookService(
	repo *repositories.WebhookRepository,
	rabbitConn *amqp.Connection,
	logger *logrus.Logger,
) (*WebhookService, error) {
	channel, err := rabbitConn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open RabbitMQ channel: %w", err)
	}

	service := &WebhookService{
		repo:          repo,
		rabbitConn:    rabbitConn,
		rabbitChannel: channel,
		exchangeName:  "webhook.development.request",
		logger:        logger,
	}

	// Declare exchange
	err = channel.ExchangeDeclare(
		service.exchangeName, // name
		"topic",              // type
		true,                 // durable
		false,                // auto-deleted
		false,                // internal
		false,                // no-wait
		nil,                  // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	return service, nil
}

// ProcessWebhook processes incoming JIRA webhook
func (s *WebhookService) ProcessWebhook(ctx context.Context, payload *models.JiraWebhookPayload) error {
	s.logger.WithFields(logrus.Fields{
		"issue_key": payload.Issue.Key,
		"event":     payload.WebhookEvent,
	}).Info("Processing webhook")

	// Validate payload
	if err := s.validatePayload(payload); err != nil {
		s.logger.WithError(err).Error("Invalid webhook payload")
		return err
	}

	// Check if status changed to "In Development"
	isInDevelopment, previousStatus := s.isStatusChangedToInDevelopment(payload)
	if !isInDevelopment {
		s.logger.WithField("status", payload.Issue.Fields.Status.Name).Debug("Status not 'In Development', ignoring")
		return ErrNotInDevelopment
	}

	s.logger.WithFields(logrus.Fields{
		"from_status": previousStatus,
		"to_status":   payload.Issue.Fields.Status.Name,
	}).Info("Detected 'In Development' status change")

	// Store webhook event
	event := &models.WebhookEvent{
		JiraIssueID:    payload.Issue.ID,
		JiraIssueKey:   payload.Issue.Key,
		JiraProjectKey: payload.Issue.Fields.Project.Key,
		Summary:        payload.Issue.Fields.Summary,
		Description:    payload.Issue.Fields.Description,
		Status:         payload.Issue.Fields.Status.Name,
		PreviousStatus: previousStatus,
		EventType:      payload.WebhookEvent,
		RawPayload:     payload,
	}

	if err := s.repo.Create(ctx, event); err != nil {
		s.logger.WithError(err).Error("Failed to store webhook event")
		return fmt.Errorf("failed to store webhook event: %w", err)
	}

	s.logger.WithField("event_id", event.ID.Hex()).Info("Webhook event stored")

	// Publish to RabbitMQ
	if err := s.publishToRabbitMQ(ctx, event); err != nil {
		s.logger.WithError(err).Error("Failed to publish to RabbitMQ")
		return fmt.Errorf("failed to publish to RabbitMQ: %w", err)
	}

	// Mark as processed
	if err := s.repo.MarkProcessed(ctx, event.ID); err != nil {
		s.logger.WithError(err).Warn("Failed to mark event as processed")
		// Don't return error - message was already published
	}

	s.logger.WithField("issue_key", payload.Issue.Key).Info("Webhook processed successfully")
	return nil
}

// validatePayload checks if the webhook payload is valid
func (s *WebhookService) validatePayload(payload *models.JiraWebhookPayload) error {
	if payload.Issue.Key == "" {
		return fmt.Errorf("%w: missing issue key", ErrInvalidPayload)
	}
	if payload.Issue.Fields.Project.Key == "" {
		return fmt.Errorf("%w: missing project key", ErrInvalidPayload)
	}
	if payload.Issue.Fields.Summary == "" {
		return fmt.Errorf("%w: missing issue summary", ErrInvalidPayload)
	}
	return nil
}

// isStatusChangedToInDevelopment checks if status changed to "In Development"
func (s *WebhookService) isStatusChangedToInDevelopment(payload *models.JiraWebhookPayload) (bool, string) {
	currentStatus := payload.Issue.Fields.Status.Name

	// Check current status
	if currentStatus != "In Development" {
		return false, ""
	}

	// Check if there's a changelog
	if payload.Changelog == nil || len(payload.Changelog.Items) == 0 {
		// No changelog, assume it's a new issue in "In Development"
		return true, ""
	}

	// Find status change in changelog
	for _, item := range payload.Changelog.Items {
		if item.Field == "status" {
			if item.ToString == "In Development" {
				return true, item.FromString
			}
		}
	}

	return false, ""
}

// publishToRabbitMQ publishes development request to RabbitMQ
func (s *WebhookService) publishToRabbitMQ(ctx context.Context, event *models.WebhookEvent) error {
	if s.rabbitChannel == nil {
		return ErrRabbitMQNotConnected
	}

	// Create development request message
	request := &models.DevelopmentRequest{
		JiraIssueID:    event.JiraIssueID,
		JiraIssueKey:   event.JiraIssueKey,
		JiraProjectKey: event.JiraProjectKey,
		Summary:        event.Summary,
		Description:    event.Description,
	}

	// Marshal to JSON
	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Publish message
	err = s.rabbitChannel.Publish(
		s.exchangeName,                  // exchange
		"webhook.development."+event.JiraProjectKey, // routing key
		false,                           // mandatory
		false,                           // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // persistent message
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"issue_key": event.JiraIssueKey,
		"exchange":  s.exchangeName,
		"routing_key": "webhook.development." + event.JiraProjectKey,
	}).Info("Published message to RabbitMQ")

	return nil
}

// Close closes RabbitMQ channel
func (s *WebhookService) Close() error {
	if s.rabbitChannel != nil {
		return s.rabbitChannel.Close()
	}
	return nil
}
