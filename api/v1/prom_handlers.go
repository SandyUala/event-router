package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PromHandler struct {
}

func NewPromHandler() *PromHandler {
	return &PromHandler{}
}

func (p *PromHandler) Register(router *gin.Engine) {
	router.GET("/metrics", p.getMetrics)
}

func (p *PromHandler) getMetrics(c *gin.Context) {
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}
