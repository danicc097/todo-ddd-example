package application

import "github.com/danicc097/todo-ddd-example/internal/generated/db"

func BadReturn() db.Todo { // want "Arch violation: Function BadReturn returns type from internal/generated/db. Use a Domain entity or API DTO instead."
	return db.Todo{}
}

func BadParam(t db.Todo) { // want "Arch violation: Function BadParam uses type from internal/generated/db in parameters. Use a Domain entity or API DTO instead."
}

type BadStruct struct {
	T db.Todo // want "Arch violation: Struct BadStruct has field using type from internal/generated/db. Use a Domain entity or API DTO instead."
}
