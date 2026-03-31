package inmemory

import (
	"go-project/internal/domain/task/errors"
	"go-project/internal/domain/task/models"

	"github.com/google/uuid"
)

func (us *Storage) GetAllTasks() ([]models.Task, error) {
	tasks := make([]models.Task, 0, len(us.tasks))
	for _, task := range us.tasks {
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (us *Storage) GetTasks(uid string) ([]models.Task, error) {
	tasks := make([]models.Task, 0, len(us.tasks))
	for _, task := range us.tasks {
		if task.UID == uid {
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}

func (us *Storage) GetTaskByTID(uid string, tid string) (models.Task, error) {
	task, exists := us.tasks[tid]
	if !exists || task.UID != uid {
		return models.Task{}, errors.ErrTaskNotFound
	}
	return task, nil
}

func (us *Storage) CreateTask(uid string, task models.Task) (uuid.UUID, error) {
	tid := uuid.New()
	_, exists := us.tasks[tid.String()]

	if exists {
		tid = uuid.New()
	}

	task.TID = tid.String()
	task.UID = uid
	us.tasks[task.TID] = task
	return tid, nil
}

func (us *Storage) UpdateTaskByTID(uid string, tid string, req models.TaskRequest) (models.Task, error) {
	taskInStorage, err := us.GetTaskByTID(uid, tid)
	if err != nil {
		return models.Task{}, err
	}

	taskInStorage = models.Task{
		TID:         tid,
		UID:         uid,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
	}

	us.tasks[taskInStorage.TID] = taskInStorage
	return taskInStorage, nil
}

func (us *Storage) DeleteTaskByTID(uid string, tid string) error {
	if _, err := us.GetTaskByTID(uid, tid); err != nil {
		return err
	}

	delete(us.tasks, tid)
	return nil
}
