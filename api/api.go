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
	server   *http.Server
	config   *ServerConfig
}

// ServerConfig is a configuraion for the server
type ServerConfig struct {
	APIInterface string
	APIPort      string
}

// NewServer returns a new Server
func NewServer(config *ServerConfig) *Server {
	// Create our new Server
	s := &Server{handlers: make([]routes.RouteHandler, 0)}

	// Set the config
	s.config = config

	// Create actual http server
	s.server = &http.Server{
		Addr: config.APIInterface + ":" + config.APIPort,
	}

	// Return our Server
	return s
}

// WithRouteHandler appends a new RouteHandler
func (s *Server) WithRouteHandler(rh routes.RouteHandler) *Server {
	s.handlers = append(s.handlers, rh)
	return s
}

// Serve starts the webserver
func (s *Server) Serve(shutdownChan <-chan struct{}) {
	log := logging.GetLogger(logrus.Fields{"package": "api"})

	router := gin.Default()
	router.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// Loop through all configured handlers and add them to our route handler
	for _, handler := range s.handlers {
		handler.Register(router)
	}

	// Set the route handler on the internal http server
	s.server.Handler = router

	go func() {
		if err := s.server.ListenAndServe(); err != nil {
			log.Error(err)
		}
	}()

	<-shutdownChan
	log.Info("Webserver received shutdown signal")
}

// Close cleans up and shutsdown the webserver
func (s *Server) Close() {
	log := logging.GetLogger(logrus.Fields{"package": "api"})

	s.server.Shutdown(context.Background())
	log.Info("Webserver has been shut down")
}
