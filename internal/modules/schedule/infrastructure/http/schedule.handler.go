package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/danicc097/todo-ddd-example/internal/apperrors"
	api "github.com/danicc097/todo-ddd-example/internal/generated/api"
	"github.com/danicc097/todo-ddd-example/internal/modules/schedule/application"
	sharedApp "github.com/danicc097/todo-ddd-example/internal/shared/application"
)

type ScheduleHandler struct {
	commitTaskHandler sharedApp.RequestHandler[application.CommitTaskCommand, application.CommitTaskResponse]
}

func NewScheduleHandler(ct sharedApp.RequestHandler[application.CommitTaskCommand, application.CommitTaskResponse]) *ScheduleHandler {
	return &ScheduleHandler{
		commitTaskHandler: ct,
	}
}

func (h *ScheduleHandler) CommitTask(c *gin.Context) {
	var req api.CommitTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.New(apperrors.InvalidInput, err.Error()))
		return
	}

	_, err := h.commitTaskHandler.Handle(c.Request.Context(), application.CommitTaskCommand{
		TodoID: req.TodoId,
		Cost:   req.Cost,
		Date:   req.Date.String(),
	})
	if err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}
