package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Client for API
type Client struct {
	logger  *logrus.Entry
	debug   bool
	handler *Handler
}

// NewClient returns a new API Client
func NewClient(log *logrus.Logger, debug bool, handler *Handler) *Client {
	client := &Client{
		logger:  log.WithFields(logrus.Fields{"package": "api"}),
		debug:   debug,
		handler: handler,
	}
	return client
}

// Serve starts the HTTP server
func (client *Client) Serve(port string) error {
	logger := client.logger.WithFields(logrus.Fields{"function": "Serve"})
	logger.Debug("Entered Serve")
	var router *gin.Engine
	if client.debug {
		gin.SetMode(gin.DebugMode)
		router = gin.Default()
	} else {
		gin.SetMode(gin.ReleaseMode)
		router = gin.New()
		router.Use(gin.Recovery())
	}

	router.GET("/org/:id", client.handler.getOrg)
	router.POST("/org", client.handler.postOrg)
	router.POST("/org/:id", client.handler.postOrg)
	router.POST("/org/:id/reprovision", client.handler.postOrgReprovision)
	router.DELETE("/org/:id", client.handler.deleteOrg)
	router.GET("/apps", client.handler.getApps)
	router.POST("/dag", client.handler.deployDag)
	router.GET("/provisions", client.handler.getProvisions)
	router.GET("/resources", client.handler.getAvailableResources)
	router.GET("/metrics", client.handler.getMetrics)
	router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	return router.Run(port)
}
