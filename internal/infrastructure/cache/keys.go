package cache

import (
	"fmt"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/db/types"
)

type keys struct{}

// Keys defines cache keys construction.
var Keys keys

const (
	prefixTodo      = "todo_query"
	prefixTag       = "tag"
	prefixWorkspace = "ws"
	prefixRateLimit = "ratelimit"
	prefixUser      = "user"
	prefixCacheTags = "cache_tags"
)

func (keys) Todo(id types.TodoID) string {
	return fmt.Sprintf("%s:%s", prefixTodo, id)
}

func (keys) TodoWorkspaceCollection(wsID types.WorkspaceID) string {
	return fmt.Sprintf("%s:collection:GetAllByWorkspace:%s", prefixTodo, wsID)
}

func (keys) TodoWorkspaceCollectionPaginated(wsID types.WorkspaceID, limit, offset int32) string {
	return fmt.Sprintf("%s:limit:%d:offset:%d", keys{}.TodoWorkspaceCollection(wsID), limit, offset)
}

func (keys) Tag(id types.TagID) string {
	return fmt.Sprintf("%s:%s", prefixTag, id)
}

func (keys) Workspace(id types.WorkspaceID) string {
	return fmt.Sprintf("%s:%s", prefixWorkspace, id)
}

func (keys) WorkspaceRevision(id types.WorkspaceID) string {
	return fmt.Sprintf("%s:%s:revision", prefixWorkspace, id)
}

func (keys) RateLimit(opID, ip string) string {
	return fmt.Sprintf("%s:%s:%s", prefixRateLimit, opID, ip)
}

func (keys) TOTPUsedCode(userID types.UserID, code string) string {
	return fmt.Sprintf("%s:%s:used_totp:%s", prefixUser, userID, code)
}

func (keys) WorkspaceTag(wsID types.WorkspaceID) string {
	return fmt.Sprintf("%s:%s", prefixWorkspace, wsID)
}

func (keys) TagSet(tag string) string {
	return fmt.Sprintf("%s:%s", prefixCacheTags, tag)
}
