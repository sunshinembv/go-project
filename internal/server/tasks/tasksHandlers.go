package tasks

import (
	"errors"
	taskErrors "go-project/internal/domain/task/errors"
	tasksDomain "go-project/internal/domain/task/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TaskService interface {
	GetTasks(uid string) ([]tasksDomain.Task, error)
	GetTaskByTID(uid string, tid string) (tasksDomain.Task, error)
	CreateTask(uid string, req tasksDomain.TaskRequest) (string, error)
	UpdateTaskByTID(uid string, tid string, req tasksDomain.TaskRequest) (tasksDomain.Task, error)
	DeleteTaskByTID(uid string, tid string) error
}

type TasksHandler struct {
	taskService TaskService
}

func New(taskService TaskService) *TasksHandler {
	return &TasksHandler{
		taskService: taskService,
	}
}

func (th *TasksHandler) GetTasks(ctx *gin.Context) {
	uid := ctx.GetString("userID")
	if uid == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	tasks, err := th.taskService.GetTasks(uid)
	if err != nil {
		if errors.Is(err, taskErrors.ErrTaskNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

func (th *TasksHandler) GetTaskByTID(ctx *gin.Context) {
	uid := ctx.GetString("userID")
	if uid == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	tid := ctx.Param("id")

	task, err := th.taskService.GetTaskByTID(uid, tid)
	if err != nil {
		if errors.Is(err, taskErrors.ErrTaskNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"task": task})
}

func (th *TasksHandler) CreateTask(ctx *gin.Context) {
	uid := ctx.GetString("userID")
	if uid == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req tasksDomain.TaskRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	tid, err := th.taskService.CreateTask(uid, req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"tid": tid})
}

func (th *TasksHandler) UpdateTaskByTID(ctx *gin.Context) {
	uid := ctx.GetString("userID")
	if uid == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req tasksDomain.TaskRequest
	tid := ctx.Param("id")

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	updatedTask, err := th.taskService.UpdateTaskByTID(uid, tid, req)
	if err != nil {
		if errors.Is(err, taskErrors.ErrTaskNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"task": updatedTask})
}

func (th *TasksHandler) DeleteTaskByTID(ctx *gin.Context) {
	uid := ctx.GetString("userID")
	if uid == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	tid := ctx.Param("id")

	if err := th.taskService.DeleteTaskByTID(uid, tid); err != nil {
		if errors.Is(err, taskErrors.ErrTaskNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"msg": "task deleted"})
}
