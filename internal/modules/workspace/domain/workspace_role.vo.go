package domain

import (
	"fmt"
)

type WorkspaceRole string

const (
	RoleOwner  WorkspaceRole = "OWNER"
	RoleMember WorkspaceRole = "MEMBER"
	RoleGuest  WorkspaceRole = "GUEST"
)

func NewWorkspaceRole(role string) (WorkspaceRole, error) {
	r := WorkspaceRole(role)
	switch r {
	case RoleOwner, RoleMember, RoleGuest:
		return r, nil
	default:
		return "", fmt.Errorf("invalid role: %s", role)
	}
}
