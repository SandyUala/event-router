package routes

import "github.com/gin-gonic/gin"

type RouteHandler interface {
	Register(router *gin.Engine)
}
