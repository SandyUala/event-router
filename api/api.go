package api

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/DeanThompson/ginpprof"
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

	ginpprof.Wrap(router)

	if string(port[0]) != ":" {
		port = ":" + port
	}
	srv := &http.Server{
		Addr:    port,
		Handler: router,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logger.Error(err)
		}
	}()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	for sig := range sigchan {
		logger.Infof("Webserver caught signal %v: terminating", sig)
		srv.Close()
	}
	logger.Info("Shutdown Webserver")
	return nil
}
