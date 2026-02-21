package domain

import (
	"fmt"
	"strings"

	"github.com/danicc097/todo-ddd-example/internal/apperrors"
	shared "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

var ErrWorkspaceDescriptionTooLong = shared.NewDomainError(apperrors.InvalidInput, fmt.Sprintf("workspace description cannot exceed %d characters", workspaceDescriptionMaxLen))

const workspaceDescriptionMaxLen = 255

type WorkspaceDescription struct {
	value string
}

func NewWorkspaceDescription(val string) (WorkspaceDescription, error) {
	val = strings.TrimSpace(val)
	if len(val) > workspaceDescriptionMaxLen {
		return WorkspaceDescription{}, ErrWorkspaceDescriptionTooLong
	}

	return WorkspaceDescription{value: val}, nil
}

func (d WorkspaceDescription) String() string {
	return d.value
}
