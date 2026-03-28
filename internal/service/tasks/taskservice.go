package taskservice

import (
	"go-project/internal/domain/task/models"
)

type Repository interface {
	GetTasks() ([]models.Task, error)
	GetTaskByID(id string) (models.Task, error)
	CreateTask(task models.Task) (models.Task, error)
	UpdateTaskByID(id string, task models.Task) (models.Task, error)
	DeleteTaskByID(id string) error
}

type TaskService struct {
	repo Repository
}

func New(repo Repository) *TaskService {
	return &TaskService{
		repo: repo,
	}
}

func (us *TaskService) GetTasks() ([]models.Task, error) {
	return us.repo.GetTasks()
}

func (us *TaskService) GetTaskByID(id string) (models.Task, error) {
	return us.repo.GetTaskByID(id)
}

func (us *TaskService) CreateTask(task models.Task) (models.Task, error) {
	return us.repo.CreateTask(task)
}

func (us *TaskService) UpdateTaskByID(id string, task models.Task) (models.Task, error) {
	return us.repo.UpdateTaskByID(id, task)
}

func (us *TaskService) DeleteTaskByID(id string) error {
	return us.repo.DeleteTaskByID(id)
}
