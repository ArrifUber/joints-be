package routes

import (
	"github.com/gin-gonic/gin"
	"joints-be/modules/context/controller"
)

func RegisterContextRoutes(router *gin.RouterGroup, ctrl *controller.ContextController) {
	contextRoutes := router.Group("/sessions/:sessionId/contexts")
	{
		contextRoutes.GET("", ctrl.ListContexts)
	}
}

