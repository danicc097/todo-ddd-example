package domain

import (
	"encoding/json"
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

func (n WorkspaceName) MarshalJSON() ([]byte, error) {
	return json.Marshal(n.value)
}

func (n WorkspaceName) MarshalText() ([]byte, error) {
	return []byte(n.value), nil
}

func (n *WorkspaceName) UnmarshalText(text []byte) error {
	vo, err := NewWorkspaceName(string(text))
	if err != nil {
		return err
	}

	*n = vo

	return nil
}
