package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/storos/sdlc-agent/configuration-api/services"
)

type DevelopmentHandler struct {
	service *services.DevelopmentService
	logger  *logrus.Logger
}

func NewDevelopmentHandler(service *services.DevelopmentService, logger *logrus.Logger) *DevelopmentHandler {
	return &DevelopmentHandler{
		service: service,
		logger:  logger,
	}
}

func (h *DevelopmentHandler) GetDevelopments(c *gin.Context) {
	jiraProjectKey := c.Query("jira_project_key")

	var developments interface{}
	var err error

	if jiraProjectKey != "" {
		developments, err = h.service.GetByJiraProjectKey(c.Request.Context(), jiraProjectKey)
	} else {
		developments, err = h.service.GetAll(c.Request.Context())
	}

	if err != nil {
		h.logger.WithError(err).Error("Failed to get developments")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get developments"})
		return
	}

	c.JSON(http.StatusOK, developments)
}

func (h *DevelopmentHandler) GetDevelopment(c *gin.Context) {
	id := c.Param("id")

	development, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		h.logger.WithError(err).WithField("id", id).Error("Failed to get development")
		c.JSON(http.StatusNotFound, gin.H{"error": "Development not found"})
		return
	}

	c.JSON(http.StatusOK, development)
}
