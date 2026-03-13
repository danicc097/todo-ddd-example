package application

import (
	"github.com/danicc097/todo-ddd-example/internal/shared/application"
)

type WorkspaceUseCases struct {
	Onboard      application.RequestHandler[OnboardWorkspaceCommand, OnboardWorkspaceResponse]
	AddMember    application.RequestHandler[AddWorkspaceMemberCommand, AddWorkspaceMemberResponse]
	RemoveMember application.RequestHandler[RemoveWorkspaceMemberCommand, RemoveWorkspaceMemberResponse]
	Delete       application.RequestHandler[DeleteWorkspaceCommand, DeleteWorkspaceResponse]
}
