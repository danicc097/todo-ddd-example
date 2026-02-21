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

func (r WorkspaceRole) String() string {
	return string(r)
}

func (r WorkspaceRole) MarshalText() ([]byte, error) {
	return []byte(r), nil
}

func (r *WorkspaceRole) UnmarshalText(text []byte) error {
	vo, err := NewWorkspaceRole(string(text))
	if err != nil {
		return err
	}

	*r = vo

	return nil
}
