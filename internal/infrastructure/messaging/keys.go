package messaging

import (
	"github.com/google/uuid"
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
	return keys{}.WorkspaceTodoAPIUpdatesChannelPrefix() + wsID.String()
}
