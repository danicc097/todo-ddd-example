package messaging

import (
	"fmt"

	"github.com/google/uuid"

	sharedDomain "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

type keys struct{}

// Keys defines messaging channel/topic keys construction.
var Keys keys

func (keys) TodoAPIUpdatesChannel() string {
	return "todo_api_updates"
}

func (keys) WorkspaceTodoAPIUpdatesChannelPrefix() string {
	return "todo_api_updates:ws:"
}

func (keys) WorkspaceTodoAPIUpdatesChannel(wsID uuid.UUID) string {
	return fmt.Sprintf("%s%s", keys{}.WorkspaceTodoAPIUpdatesChannelPrefix(), wsID)
}

func (keys) EventRoutingKey(eventType sharedDomain.EventType, aggID uuid.UUID) string {
	return fmt.Sprintf("%s.%s", eventType, aggID)
}
