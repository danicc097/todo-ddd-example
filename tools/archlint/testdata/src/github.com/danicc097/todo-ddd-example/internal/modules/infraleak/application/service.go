package application // want "Arch violation: Application package github.com/danicc097/todo-ddd-example/internal/modules/infraleak/application imports infrastructure package github.com/danicc097/todo-ddd-example/internal/infrastructure/crypto. Depend on domain ports instead."

import "github.com/danicc097/todo-ddd-example/internal/infrastructure/crypto"


var _ = crypto.MasterKeyLength
