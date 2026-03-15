package http

import "github.com/danicc097/todo-ddd-example/internal/generated/db"

func BadHandlerReturn() db.Todo { // want "Arch violation: Function BadHandlerReturn returns type from internal/generated/db. Use a Domain entity or API DTO instead."
	return db.Todo{}
}

func BadHandlerParam(t db.Todo) { // want "Arch violation: Function BadHandlerParam uses type from internal/generated/db in parameters. Use a Domain entity or API DTO instead."
}

type BadHandlerStruct struct {
	T db.Todo // want "Arch violation: Struct BadHandlerStruct has field using type from internal/generated/db. Use a Domain entity or API DTO instead."
}
