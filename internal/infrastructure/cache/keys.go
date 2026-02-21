package cache

import (
	"fmt"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/db/types"
)

type keys struct{}

// Keys defines cache keys construction.
var Keys keys

func (keys) Todo(id types.TodoID) string {
	return fmt.Sprintf("todo_query:%s", id)
}

func (keys) TodoWorkspaceCollection(wsID types.WorkspaceID) string {
	return fmt.Sprintf("todo_query:collection:GetAllByWorkspace:%s", wsID)
}

func (keys) Tag(id types.TagID) string {
	return fmt.Sprintf("tag:%s", id)
}

func (keys) Workspace(id types.WorkspaceID) string {
	return fmt.Sprintf("ws:%s", id)
}

func (keys) WorkspaceRevision(id types.WorkspaceID) string {
	return fmt.Sprintf("ws:%s:revision", id)
}

func (keys) AllWorkspaces() string {
	return "ws:collection:FindAll:all"
}

func (keys) RateLimit(opID, ip string) string {
	return fmt.Sprintf("ratelimit:%s:%s", opID, ip)
}

func (keys) TOTPUsedCode(userID types.UserID, code string) string {
	return fmt.Sprintf("user:%s:used_totp:%s", userID.String(), code)
}

func (keys) WorkspaceTag(wsID types.WorkspaceID) string {
	return "ws:" + wsID.String()
}

func (keys) TagSet(tag string) string {
	return "cache_tags:" + tag
}

func (keys) TodoAPIUpdatesChannel() string {
	return "todo_api_updates"
}
