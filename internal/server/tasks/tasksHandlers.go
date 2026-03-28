package tasks

import (
	"errors"
	taskErrors "go-project/internal/domain/task/errors"
	tasksDomain "go-project/internal/domain/task/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TaskService interface {
	GetTasks() ([]tasksDomain.Task, error)
	GetTaskByID(id string) (tasksDomain.Task, error)
	CreateTask(task tasksDomain.Task) (tasksDomain.Task, error)
	UpdateTaskByID(id string, task tasksDomain.Task) (tasksDomain.Task, error)
	DeleteTaskByID(id string) error
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

	tasks, err := th.taskService.GetTasks()
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

func (th *TasksHandler) GetTaskByID(ctx *gin.Context) {
	taskID := ctx.Param("id")

	task, err := th.taskService.GetTaskByID(taskID)
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
	var task tasksDomain.Task

	if err := ctx.ShouldBindJSON(&task); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	taskInStorage, err := th.taskService.CreateTask(task)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"task": taskInStorage})
}

func (th *TasksHandler) UpdateTaskByID(ctx *gin.Context) {
	var task tasksDomain.Task
	taskID := ctx.Param("id")

	if err := ctx.ShouldBindJSON(&task); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	updatedTask, err := th.taskService.UpdateTaskByID(taskID, task)
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

func (th *TasksHandler) DeleteTaskByID(ctx *gin.Context) {
	taskID := ctx.Param("id")

	if err := th.taskService.DeleteTaskByID(taskID); err != nil {
		if errors.Is(err, taskErrors.ErrTaskNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"msg": "task deleted"})
}
