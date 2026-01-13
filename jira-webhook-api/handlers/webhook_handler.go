package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/storos/sdlc-agent/jira-webhook-api/models"
	"github.com/storos/sdlc-agent/jira-webhook-api/services"
)

// WebhookHandler handles webhook HTTP requests
type WebhookHandler struct {
	service *services.WebhookService
	logger  *logrus.Logger
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(service *services.WebhookService, logger *logrus.Logger) *WebhookHandler {
	return &WebhookHandler{
		service: service,
		logger:  logger,
	}
}

// HandleWebhook processes incoming JIRA webhook
// POST /webhook
func (h *WebhookHandler) HandleWebhook(c *gin.Context) {
	var payload models.JiraWebhookPayload

	// Bind JSON payload
	if err := c.ShouldBindJSON(&payload); err != nil {
		h.logger.WithError(err).Error("Failed to parse webhook payload")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid webhook payload",
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"issue_key": payload.Issue.Key,
		"event":     payload.WebhookEvent,
		"status":    payload.Issue.Fields.Status.Name,
	}).Info("Received webhook")

	// Process webhook
	err := h.service.ProcessWebhook(c.Request.Context(), &payload)
	if err != nil {
		// Check if it's just not "In Development" status
		if err == services.ErrNotInDevelopment {
			h.logger.Debug("Webhook ignored - not 'In Development' status")
			c.JSON(http.StatusOK, gin.H{
				"message": "Webhook received but ignored (not 'In Development' status)",
			})
			return
		}

		// Log and return error for other cases
		h.logger.WithError(err).Error("Failed to process webhook")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process webhook",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Webhook processed successfully",
		"issue_key": payload.Issue.Key,
	})
}

// HealthCheck returns health status
// GET /health
func (h *WebhookHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"service": "jira-webhook-api",
	})
}
