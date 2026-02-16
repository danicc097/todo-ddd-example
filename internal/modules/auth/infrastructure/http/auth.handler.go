package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/danicc097/todo-ddd-example/internal/apperrors"
	api "github.com/danicc097/todo-ddd-example/internal/generated/api"
	"github.com/danicc097/todo-ddd-example/internal/modules/auth/application"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	sharedApp "github.com/danicc097/todo-ddd-example/internal/shared/application"
	"github.com/negrel/secrecy"
)

type AuthHandler struct {
	loginHandler    sharedApp.RequestHandler[application.LoginCommand, application.LoginResponse]
	registerHandler sharedApp.RequestHandler[application.RegisterCommand, userDomain.UserID]
	initiateHandler sharedApp.RequestHandler[sharedApp.Void, string]
	verifyHandler   sharedApp.RequestHandler[application.VerifyTOTPCommand, application.VerifyTOTPResponse]
}

func NewAuthHandler(
	login sharedApp.RequestHandler[application.LoginCommand, application.LoginResponse],
	register sharedApp.RequestHandler[application.RegisterCommand, userDomain.UserID],
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
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.New(apperrors.InvalidInput, err.Error()))
		return
	}

	resp, err := h.loginHandler.Handle(c.Request.Context(), application.LoginCommand{
		Email:    req.Email,
		Password: *secrecy.NewSecret(req.Password),
	})
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"accessToken": resp.AccessToken})
}

func (h *AuthHandler) Register(c *gin.Context, params api.RegisterParams) {
	var req struct {
		Email    string `json:"email"`
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.New(apperrors.InvalidInput, err.Error()))
		return
	}

	id, err := h.registerHandler.Handle(c.Request.Context(), application.RegisterCommand{
		Email:    req.Email,
		Name:     req.Name,
		Password: *secrecy.NewSecret(req.Password),
	})
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id.UUID})
}

func (h *AuthHandler) InitiateTOTP(c *gin.Context) {
	uri, err := h.initiateHandler.Handle(c.Request.Context(), sharedApp.Void{})
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"provisioningUri": uri})
}

func (h *AuthHandler) VerifyTOTP(c *gin.Context) {
	var req struct {
		Code string `json:"code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.New(apperrors.InvalidInput, err.Error()))
		return
	}

	resp, err := h.verifyHandler.Handle(c.Request.Context(), application.VerifyTOTPCommand{Code: req.Code})
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"accessToken": resp.AccessToken})
}
