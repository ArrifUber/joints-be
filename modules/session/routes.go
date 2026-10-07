package routes

import (
	"github.com/gin-gonic/gin"
	"joints-be/modules/session/controller"
)

func RegisterSessionRoutes(router *gin.RouterGroup, ctrl *controller.SessionController) {
	sessionRoutes := router.Group("/sessions")
	{
		sessionRoutes.POST("", ctrl.CreateSession)
		sessionRoutes.GET("/:id", ctrl.GetSession)
	}
}
