package api

import (
	"net/http"

	"github.com/astronomerio/event-router/api/routes"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var (
	log   = logrus.WithField("package", "api")
	Debug = false
)

type Client struct {
	handlers []routes.RouteHandler
}

func NewClient() *Client {
	return &Client{handlers: make([]routes.RouteHandler, 0)}
}

func (c *Client) AppendRouteHandler(rh routes.RouteHandler) {
	c.handlers = append(c.handlers, rh)
}

func (c *Client) Serve(port string) error {
	logger := log.WithFields(logrus.Fields{"function": "Serve"})
	logger.Debug("Entered Serve")
	var router *gin.Engine
	if Debug {
		gin.SetMode(gin.DebugMode)
		router = gin.Default()
	} else {
		gin.SetMode(gin.ReleaseMode)
		router = gin.New()
		router.Use(gin.Recovery())
	}

	for _, handler := range c.handlers {
		handler.Register(router)
	}

	router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})
	return router.Run(port)
}
