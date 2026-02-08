package main

import (
	"context"
	"log"
	"os"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/application"
	todoHttp "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/http"
	todoDb "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/postgres"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	pool, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}
	defer pool.Close()

	todoRepo := todoDb.NewTodoRepo(pool)

	createTodoUC := application.NewCreateTodoUseCase(todoRepo)
	completeTodoUC := application.NewCompleteTodoUseCase(todoRepo)
	getAllTodoUC := application.NewGetAllTodosUseCase(todoRepo)

	todoHandler := todoHttp.NewTodoHandler(createTodoUC, completeTodoUC, getAllTodoUC)

	r := gin.Default()

	v1 := r.Group("/api/v1")
	{
		v1.POST("/todos", todoHandler.Create)
		v1.PATCH("/todos/:id/complete", todoHandler.Complete)
		v1.GET("/todos", todoHandler.GetAll)
	}

	if err := r.Run(":8082"); err != nil {
		log.Fatal(err)
	}
}
