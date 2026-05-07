package taskservice

import (
	"go-project/internal/domain/task/models"

	"github.com/google/uuid"
)

type Repository interface {
	GetTasks(uid string) ([]models.Task, error)
	GetTaskByTID(uid string, tid string) (models.Task, error)
	CreateTask(uid string, task models.Task) (uuid.UUID, error)
	UpdateTaskByTID(uid string, tid string, req models.TaskRequest) (models.Task, error)
	DeleteTaskByTID(uid string, tid string) error
}

type TaskService struct {
	repo Repository
}

func New(repo Repository) *TaskService {
	return &TaskService{
		repo: repo,
	}
}

func (us *TaskService) GetTasks(uid string) ([]models.Task, error) {
	return us.repo.GetTasks(uid)
}

func (us *TaskService) GetTaskByTID(uid string, tid string) (models.Task, error) {
	return us.repo.GetTaskByTID(uid, tid)
}

func (us *TaskService) CreateTask(uid string, req models.TaskRequest) (string, error) {
	var task = models.Task{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
	}
	tid, err := us.repo.CreateTask(uid, task)
	if err != nil {
		return "", err
	}
	return tid.String(), nil
}

func (us *TaskService) UpdateTaskByTID(uid string, tid string, req models.TaskRequest) (models.Task, error) {
	return us.repo.UpdateTaskByTID(uid, tid, req)
}

func (us *TaskService) DeleteTaskByTID(uid string, tid string) error {
	return us.repo.DeleteTaskByTID(uid, tid)
}
