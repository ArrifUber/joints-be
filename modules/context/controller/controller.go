package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"joints-be/modules/context/service"
	"joints-be/pkg/response"
)

type ContextController struct {
	service service.ContextService
}

func NewContextController(svc service.ContextService) *ContextController {
	return &ContextController{service: svc}
}

// ListContexts handles GET /sessions/:sessionId/contexts
func (c *ContextController) ListContexts(ctx *gin.Context) {
	sessionID := ctx.Param("sessionId")

	contexts, err := c.service.ListBySession(ctx.Request.Context(), sessionID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.Error("Failed to retrieve contexts", nil))
		return
	}

	ctx.JSON(http.StatusOK, response.Success("Contexts retrieved", contexts))
}

