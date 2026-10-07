package routes

import (
	"github.com/gin-gonic/gin"
	"joints-be/modules/transcript/controller"
)

// RegisterTranscriptRoutes registers transcript routes under /sessions/:sessionId/transcripts.
func RegisterTranscriptRoutes(router *gin.RouterGroup, ctrl *controller.TranscriptController) {
	transcriptRoutes := router.Group("/sessions/:sessionId/transcripts")
	{
		transcriptRoutes.POST("", ctrl.CreateTranscript)
		transcriptRoutes.GET("", ctrl.ListTranscripts)
	}
}

