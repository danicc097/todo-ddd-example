package application

import (
	"github.com/danicc097/todo-ddd-example/internal/shared/application"
)

type AuthUseCases struct {
	Login        application.RequestHandler[LoginCommand, LoginResponse]
	Register     application.RequestHandler[RegisterCommand, RegisterUserResponse]
	InitiateTOTP application.RequestHandler[application.Void, string]
	VerifyTOTP   application.RequestHandler[VerifyTOTPCommand, VerifyTOTPResponse]
}
