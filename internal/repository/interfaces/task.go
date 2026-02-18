package interfaces

import "go-project/internal/domain/task/models"

type ITaskStorage interface {
	GetTasks() ([]models.Task, error)
	GetTaskByID(id string) (models.Task, error)
	CreateTask(task models.Task) (models.Task, error)
	UpdateTaskByID(id string, task models.Task) (models.Task, error)
	DeleteTaskByID(id string) error
}
