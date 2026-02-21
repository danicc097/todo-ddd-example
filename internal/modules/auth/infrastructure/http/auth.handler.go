package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/danicc097/todo-ddd-example/internal/apperrors"
	api "github.com/danicc097/todo-ddd-example/internal/generated/api"
	"github.com/danicc097/todo-ddd-example/internal/modules/auth/application"
	sharedApp "github.com/danicc097/todo-ddd-example/internal/shared/application"
)

type AuthHandler struct {
	loginHandler    sharedApp.RequestHandler[application.LoginCommand, application.LoginResponse]
	registerHandler sharedApp.RequestHandler[application.RegisterCommand, application.RegisterUserResponse]
	initiateHandler sharedApp.RequestHandler[sharedApp.Void, string]
	verifyHandler   sharedApp.RequestHandler[application.VerifyTOTPCommand, application.VerifyTOTPResponse]
}

func NewAuthHandler(
	login sharedApp.RequestHandler[application.LoginCommand, application.LoginResponse],
	register sharedApp.RequestHandler[application.RegisterCommand, application.RegisterUserResponse],
	initiate sharedApp.RequestHandler[sharedApp.Void, string],
	verify sharedApp.RequestHandler[application.VerifyTOTPCommand, application.VerifyTOTPResponse],
) *AuthHandler {
	return &AuthHandler{
		loginHandler:    login,
		registerHandler: register,
		initiateHandler: initiate,
		verifyHandler:   verify,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req api.LoginRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.New(apperrors.InvalidInput, err.Error()))
		return
	}

	resp, err := h.loginHandler.Handle(c.Request.Context(), application.LoginCommand{
		Email:    string(req.Email),
		Password: req.Password,
	})
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.LoginResponseBody{AccessToken: resp.AccessToken})
}

func (h *AuthHandler) Register(c *gin.Context, params api.RegisterParams) {
	var req api.RegisterUserRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.New(apperrors.InvalidInput, err.Error()))
		return
	}

	resp, err := h.registerHandler.Handle(c.Request.Context(), application.RegisterCommand{
		Email:    string(req.Email),
		Name:     req.Name,
		Password: req.Password,
	})
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, api.IdResponse{Id: resp.ID.UUID()})
}

func (h *AuthHandler) InitiateTOTP(c *gin.Context) {
	uri, err := h.initiateHandler.Handle(c.Request.Context(), sharedApp.Void{})
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.InitiateTOTPResponseBody{ProvisioningUri: uri})
}

func (h *AuthHandler) VerifyTOTP(c *gin.Context) {
	var req api.VerifyTOTPRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.New(apperrors.InvalidInput, err.Error()))
		return
	}

	resp, err := h.verifyHandler.Handle(c.Request.Context(), application.VerifyTOTPCommand{Code: req.Code})
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.LoginResponseBody{AccessToken: resp.AccessToken})
}
