package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/storos/sdlc-agent/configuration-api/services"
)

type WebhookHandler struct {
	service *services.WebhookService
	logger  *logrus.Logger
}

func NewWebhookHandler(service *services.WebhookService, logger *logrus.Logger) *WebhookHandler {
	return &WebhookHandler{
		service: service,
		logger:  logger,
	}
}

func (h *WebhookHandler) GetWebhookEvents(c *gin.Context) {
	jiraProjectKey := c.Query("jira_project_key")

	var webhookEvents interface{}
	var err error

	if jiraProjectKey != "" {
		webhookEvents, err = h.service.GetByJiraProjectKey(c.Request.Context(), jiraProjectKey)
	} else {
		webhookEvents, err = h.service.GetAll(c.Request.Context())
	}

	if err != nil {
		h.logger.WithError(err).Error("Failed to get webhook events")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get webhook events"})
		return
	}

	c.JSON(http.StatusOK, webhookEvents)
}

func (h *WebhookHandler) GetWebhookEvent(c *gin.Context) {
	id := c.Param("id")

	webhookEvent, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		h.logger.WithError(err).WithField("id", id).Error("Failed to get webhook event")
		c.JSON(http.StatusNotFound, gin.H{"error": "Webhook event not found"})
		return
	}

	c.JSON(http.StatusOK, webhookEvent)
}
