package taskservice

import (
	"go-project/internal/domain/task/models"
	"go-project/internal/repository/interfaces"
)

type TaskService struct {
	db interfaces.ITaskStorage
}

func NewTaskService(db interfaces.ITaskStorage) *TaskService {
	return &TaskService{
		db: db,
	}
}

func (us *TaskService) GetTasks() ([]models.Task, error) {
	return us.db.GetTasks()
}

func (us *TaskService) GetTaskByID(id string) (models.Task, error) {
	return us.db.GetTaskByID(id)
}

func (us *TaskService) CreateTask(task models.Task) (models.Task, error) {
	return us.db.CreateTask(task)
}

func (us *TaskService) UpdateTaskByID(id string, task models.Task) (models.Task, error) {
	return us.db.UpdateTaskByID(id, task)
}

func (us *TaskService) DeleteTaskByID(id string) error {
	return us.db.DeleteTaskByID(id)
}
