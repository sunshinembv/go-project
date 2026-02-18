package server

import (
	"errors"
	taskErrors "go-project/internal/domain/task/errors"
	"go-project/internal/domain/task/models"
	taskService "go-project/internal/service/task_service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (srv *TodoListApi) getTasks(ctx *gin.Context) {
	usecase := taskService.NewTaskService(srv.db)

	tasks, err := usecase.GetTasks()
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

func (srv *TodoListApi) getTaskByID(ctx *gin.Context) {
	taskID := ctx.Param("id")

	usecase := taskService.NewTaskService(srv.db)
	task, err := usecase.GetTaskByID(taskID)
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

func (srv *TodoListApi) createTask(ctx *gin.Context) {
	var task models.Task

	if err := ctx.ShouldBindJSON(&task); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	usecase := taskService.NewTaskService(srv.db)

	taskInStorage, err := usecase.CreateTask(task)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"task": taskInStorage})
}

func (srv *TodoListApi) updateTaskByID(ctx *gin.Context) {
	var task models.Task
	taskID := ctx.Param("id")

	if err := ctx.ShouldBindJSON(&task); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	usecase := taskService.NewTaskService(srv.db)

	updatedTask, err := usecase.UpdateTaskByID(taskID, task)
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

func (srv *TodoListApi) deleteTaskByID(ctx *gin.Context) {
	taskID := ctx.Param("id")

	usecase := taskService.NewTaskService(srv.db)
	if err := usecase.DeleteTaskByID(taskID); err != nil {
		if errors.Is(err, taskErrors.ErrTaskNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"msg": "task deleted"})
}
