package cache

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/db/types"
)

type keys struct{}

// Keys defines cache keys construction.
var Keys keys

const (
	prefixTodoAgg   = "todo:agg"
	prefixTodoRead  = "todo:read"
	prefixTag       = "tag"
	prefixWorkspace = "ws"
	prefixRateLimit = "ratelimit"
	prefixUser      = "user"
	prefixCacheTags = "cache_tags"
)

func (keys) TodoAggregate(id types.TodoID) string {
	return fmt.Sprintf("%s:%s", prefixTodoAgg, id)
}

func (keys) TodoReadModel(id types.TodoID) string {
	return fmt.Sprintf("%s:%s", prefixTodoRead, id)
}

func (keys) TodoWorkspaceCollection(wsID types.WorkspaceID, revision string) string {
	return fmt.Sprintf("%s:collection:GetAllByWorkspace:%s:rev:%s", prefixTodoRead, wsID, revision)
}

func (keys) TodoWorkspaceCollectionPaginated(wsID types.WorkspaceID, limit, offset int32, revision string) string {
	return fmt.Sprintf("%s:limit:%d:offset:%d", keys{}.TodoWorkspaceCollection(wsID, revision), limit, offset)
}

func (keys) IdempotencyKey(id uuid.UUID) string {
	return "idempotency:" + id.String()
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
