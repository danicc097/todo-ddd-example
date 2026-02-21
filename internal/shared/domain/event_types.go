package domain

type EventType string

const (
	TodoCreated            EventType = "todo.created"
	TodoCompleted          EventType = "todo.completed"
	TodoTagAdded           EventType = "todo.tag_added"
	TodoTagCreated         EventType = "todo.tag_created"
	WorkspaceCreated       EventType = "workspace.created"
	WorkspaceDeleted       EventType = "workspace.deleted"
	WorkspaceMemberAdded   EventType = "workspace.member_added"
	WorkspaceMemberRemoved EventType = "workspace.member_removed"
	UserCreated            EventType = "user.created"
	UserDeleted            EventType = "user.deleted"
)
