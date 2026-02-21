package domain

import (
	"fmt"
	"strings"

	"github.com/danicc097/todo-ddd-example/internal/apperrors"
	shared "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

var (
	ErrWorkspaceNameEmpty   = shared.NewDomainError(apperrors.InvalidInput, "workspace name cannot be empty")
	ErrWorkspaceNameTooLong = shared.NewDomainError(apperrors.InvalidInput, fmt.Sprintf("workspace name cannot exceed %d characters", workspaceNameMaxLen))
)

const workspaceNameMaxLen = 100

type WorkspaceName struct {
	value string
}

func NewWorkspaceName(val string) (WorkspaceName, error) {
	val = strings.TrimSpace(val)
	if val == "" {
		return WorkspaceName{}, ErrWorkspaceNameEmpty
	}

	if len(val) > workspaceNameMaxLen {
		return WorkspaceName{}, ErrWorkspaceNameTooLong
	}

	return WorkspaceName{value: val}, nil
}

func (n WorkspaceName) String() string {
	return n.value
}
