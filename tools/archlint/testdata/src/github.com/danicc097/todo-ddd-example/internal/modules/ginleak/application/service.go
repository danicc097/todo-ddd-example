package application

import (
	"context"

	"github.com/gin-gonic/gin"
)

func CreateTodo(c *gin.Context) { // want "Arch violation: Application layer function CreateTodo uses \\*gin.Context. Use context.Context instead."
}

func UpdateTodo(ctx context.Context, c *gin.Context) { // want "Arch violation: Application layer function UpdateTodo uses \\*gin.Context. Use context.Context instead."
}
