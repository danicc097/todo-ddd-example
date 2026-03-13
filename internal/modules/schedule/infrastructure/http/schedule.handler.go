package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	api "github.com/danicc097/todo-ddd-example/internal/generated/api"
	"github.com/danicc097/todo-ddd-example/internal/modules/schedule/application"
	infraHttp "github.com/danicc097/todo-ddd-example/internal/shared/infrastructure/http"
)

type ScheduleHandler struct {
	uc application.ScheduleUseCases
}

func NewScheduleHandler(uc application.ScheduleUseCases) *ScheduleHandler {
	return &ScheduleHandler{uc: uc}
}

func (h *ScheduleHandler) CommitTask(c *gin.Context) {
	req, ok := infraHttp.BindJSON[api.CommitTaskRequest](c)
	if !ok {
		return
	}

	if _, ok := infraHttp.Execute(c, h.uc.CommitTask, application.CommitTaskCommand{
		TodoID: req.TodoId,
		Cost:   req.Cost,
		Date:   req.Date.String(),
	}); ok {
		c.Status(http.StatusNoContent)
	}
}
