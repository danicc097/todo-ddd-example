package testutils

import (
	"fmt"

	"github.com/google/uuid"

	sharedDomain "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

// RandomEventType returns a unique event type for testing.
func RandomEventType() sharedDomain.EventType {
	return sharedDomain.EventType(fmt.Sprintf("test.event.%s", uuid.New()))
}
