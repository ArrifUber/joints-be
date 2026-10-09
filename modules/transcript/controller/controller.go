package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"joints-be/modules/transcript/dto"
	"joints-be/modules/transcript/service"
	"joints-be/pkg/response"
)

type TranscriptController struct {
	service  service.TranscriptService
	validate *validator.Validate
}

func NewTranscriptController(svc service.TranscriptService, validate *validator.Validate) *TranscriptController {
	return &TranscriptController{
		service:  svc,
		validate: validate,
	}
}

// CreateTranscript handles POST /sessions/:id/transcripts
// Returns 202 Accepted immediately after saving — never waits for AI processing.
func (c *TranscriptController) CreateTranscript(ctx *gin.Context) {
	sessionID := ctx.Param("id")
	if sessionID == "" {
		sessionID = ctx.Param("sessionId")
	}

	var req dto.CreateTranscriptRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error("Invalid request payload", err.Error()))
		return
	}

	if err := c.validate.Struct(&req); err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, response.Error("Validation failed", err.Error()))
		return
	}

	res, err := c.service.AcceptTranscript(ctx.Request.Context(), sessionID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSessionNotFound):
			ctx.JSON(http.StatusNotFound, response.Error("Session not found", nil))
		case errors.Is(err, service.ErrDuplicateSequence):
			ctx.JSON(http.StatusConflict, response.Error("Transcript sequence already exists for this session", nil))
		default:
			ctx.JSON(http.StatusInternalServerError, response.Error("Failed to accept transcript", nil))
		}
		return
	}

	// 202 Accepted: backend received and will process asynchronously
	ctx.JSON(http.StatusAccepted, response.Success("Transcript accepted", res))
}

// ListTranscripts handles GET /sessions/:id/transcripts
func (c *TranscriptController) ListTranscripts(ctx *gin.Context) {
	sessionID := ctx.Param("id")
	if sessionID == "" {
		sessionID = ctx.Param("sessionId")
	}

	chunks, err := c.service.ListBySession(ctx.Request.Context(), sessionID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.Error("Failed to retrieve transcripts", nil))
		return
	}

	ctx.JSON(http.StatusOK, response.Success("Transcripts retrieved", chunks))
}

