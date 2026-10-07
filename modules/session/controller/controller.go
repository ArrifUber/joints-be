package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"joints-be/modules/session/dto"
	"joints-be/modules/session/service"
	"joints-be/pkg/response"
)

type SessionController struct {
	service  service.SessionService
	validate *validator.Validate
}

func NewSessionController(service service.SessionService, validate *validator.Validate) *SessionController {
	return &SessionController{
		service:  service,
		validate: validate,
	}
}

func (c *SessionController) CreateSession(ctx *gin.Context) {
	var req dto.CreateSessionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error("Invalid request payload", err.Error()))
		return
	}

	if err := c.validate.Struct(&req); err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, response.Error("Validation failed", err.Error()))
		return
	}

	res, err := c.service.CreateSession(ctx.Request.Context(), req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.Error("Failed to create session", err.Error()))
		return
	}

	ctx.JSON(http.StatusCreated, response.Success("Session created successfully", res))
}

func (c *SessionController) GetSession(ctx *gin.Context) {
	id := ctx.Param("id")

	res, err := c.service.GetSessionByID(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, response.Error("Session not found", err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success("Session retrieved successfully", res))
}

