package http

import "github.com/google/uuid"

type createTodoRequest struct {
	Title string `json:"title" binding:"required"`
}

type todoResponse struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	CreatedAt string    `json:"createdAt"`
}
