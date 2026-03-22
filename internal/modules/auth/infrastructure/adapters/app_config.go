package adapters

import (
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/messaging"
	"github.com/danicc097/todo-ddd-example/internal/modules/auth/domain"
)

type MessagingAppConfig struct{}

var _ domain.AppConfig = (*MessagingAppConfig)(nil)

func NewMessagingAppConfig() *MessagingAppConfig {
	return &MessagingAppConfig{}
}

func (c *MessagingAppConfig) DisplayName() string {
	return messaging.Keys.AppDisplayName()
}
