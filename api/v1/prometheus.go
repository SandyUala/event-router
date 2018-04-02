package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PrometheusHandler is a Prometheus route handler
type PrometheusHandler struct{}

// NewPrometheusHandler creates a new PrometheusHandler
func NewPrometheusHandler() *PrometheusHandler {
	return &PrometheusHandler{}
}

// Register registers this route handler with the given Engine
func (p *PrometheusHandler) Register(router *gin.Engine) {
	router.GET("/metrics", p.getMetrics)
}

func (p *PrometheusHandler) getMetrics(c *gin.Context) {
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}
