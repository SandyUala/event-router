package routes

import (
	"github.com/gin-gonic/gin"
)

// RouteHandler for Commander
type RouteHandler interface {
	Register(router *gin.Engine)
}
