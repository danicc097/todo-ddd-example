package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	api "github.com/danicc097/todo-ddd-example/internal/generated/api"
	"github.com/danicc097/todo-ddd-example/internal/modules/auth/application"
	sharedApp "github.com/danicc097/todo-ddd-example/internal/shared/application"
	infraHttp "github.com/danicc097/todo-ddd-example/internal/shared/infrastructure/http"
)

type AuthUseCases struct {
	Login        sharedApp.RequestHandler[application.LoginCommand, application.LoginResponse]
	Register     sharedApp.RequestHandler[application.RegisterCommand, application.RegisterUserResponse]
	InitiateTOTP sharedApp.RequestHandler[sharedApp.Void, string]
	VerifyTOTP   sharedApp.RequestHandler[application.VerifyTOTPCommand, application.VerifyTOTPResponse]
}

type AuthHandler struct {
	uc AuthUseCases
}

func NewAuthHandler(uc AuthUseCases) *AuthHandler {
	return &AuthHandler{uc: uc}
}

func (h *AuthHandler) Login(c *gin.Context) {
	req, ok := infraHttp.BindJSON[api.LoginRequestBody](c)
	if !ok {
		return
	}

	resp, ok := infraHttp.Execute(c, h.uc.Login, application.LoginCommand{
		Email:    string(req.Email),
		Password: req.Password,
	})
	if ok {
		c.JSON(http.StatusOK, api.LoginResponseBody{AccessToken: resp.AccessToken})
	}
}

func (h *AuthHandler) Register(c *gin.Context, params api.RegisterParams) {
	req, ok := infraHttp.BindJSON[api.RegisterUserRequestBody](c)
	if !ok {
		return
	}

	resp, ok := infraHttp.Execute(c, h.uc.Register, application.RegisterCommand{
		Email:    string(req.Email),
		Name:     req.Name,
		Password: req.Password,
	})
	if ok {
		c.JSON(http.StatusCreated, api.IdResponse{Id: resp.ID.UUID()})
	}
}

func (h *AuthHandler) InitiateTOTP(c *gin.Context) {
	resp, ok := infraHttp.Execute(c, h.uc.InitiateTOTP, sharedApp.Void{})
	if ok {
		c.JSON(http.StatusOK, api.InitiateTOTPResponseBody{ProvisioningUri: resp})
	}
}

func (h *AuthHandler) VerifyTOTP(c *gin.Context) {
	req, ok := infraHttp.BindJSON[api.VerifyTOTPRequestBody](c)
	if !ok {
		return
	}

	resp, ok := infraHttp.Execute(c, h.uc.VerifyTOTP, application.VerifyTOTPCommand{Code: req.Code})
	if ok {
		c.JSON(http.StatusOK, api.LoginResponseBody{AccessToken: resp.AccessToken})
	}
}
