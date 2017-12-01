package v1

import (
	"net/http"

	"github.com/astronomerio/cs-event-router/integrations"
	"github.com/gin-gonic/gin"
)

type IntegrationsHandler struct {
	client *integrations.Client
}

func NewIntegrationsHandler(client *integrations.Client) *IntegrationsHandler {
	return &IntegrationsHandler{
		client: client,
	}
}

func (i *IntegrationsHandler) Register(router *gin.Engine) {
	router.GET("/integrations", i.getIntegrations)
}

func (i IntegrationsHandler) getIntegrations(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"integrations": i.client.GetAllIntegrations()})
}
