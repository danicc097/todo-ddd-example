package domain

type EventType string

const (
	TodoCreated            EventType = "todo.created"
	TodoCompleted          EventType = "todo.completed"
	TodoTagAdded           EventType = "todo.tag_added"
	WorkspaceCreated       EventType = "workspace.created"
	WorkspaceDeleted       EventType = "workspace.deleted"
	WorkspaceMemberAdded   EventType = "workspace.member_added"
	WorkspaceMemberRemoved EventType = "workspace.member_removed"
)
