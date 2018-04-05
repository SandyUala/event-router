package api

import (
	"context"
	"net/http"

	"github.com/astronomerio/event-router/logging"
	"github.com/astronomerio/phoenix/api/routes"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Server is an API client
type Server struct {
	handlers []routes.RouteHandler
	config   *ServerConfig
}

// ServerConfig is a configuraion for the server
type ServerConfig struct {
	APIInterface string
	APIPort      string
}

// NewServer returns a new Server
func NewServer() *Server {
	return &Server{handlers: make([]routes.RouteHandler, 0)}
}

// WithConfig sets the servers config
func (s *Server) WithConfig(config *ServerConfig) *Server {
	s.config = config
	return s
}

// WithRouteHandler appends a new RouteHandler
func (s *Server) WithRouteHandler(rh routes.RouteHandler) *Server {
	s.handlers = append(s.handlers, rh)
	return s
}

// Run starts the webserver
func (s *Server) Run(shutdownChan <-chan struct{}) {
	log := logging.GetLogger(logrus.Fields{"package": "api"})

	router := gin.Default()
	for _, handler := range s.handlers {
		handler.Register(router)
	}

	router.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	srv := &http.Server{
		Addr:    s.config.APIInterface + ":" + s.config.APIPort,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error("Error starting server ", err)
		}
	}()

	<-shutdownChan
	srv.Shutdown(context.Background())
	log.Info("Webserver shut down")
}
