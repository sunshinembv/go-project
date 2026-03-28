package inmemory

import (
	"go-project/internal/domain/task/errors"
	"go-project/internal/domain/task/models"

	"github.com/google/uuid"
)

func (us *Storage) GetTasks() ([]models.Task, error) {
	tasks := make([]models.Task, 0, len(us.tasks))
	for _, task := range us.tasks {
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (us *Storage) GetTaskByID(id string) (models.Task, error) {
	task, exists := us.tasks[id]
	if !exists {
		return models.Task{}, errors.ErrTaskNotFound
	}
	return task, nil
}

func (us *Storage) CreateTask(task models.Task) (models.Task, error) {
	uid := uuid.New().String()
	_, exists := us.tasks[uid]

	if exists {
		uid = uuid.New().String()
	}

	task.UID = uid
	us.tasks[task.UID] = task
	return task, nil
}

func (us *Storage) UpdateTaskByID(id string, task models.Task) (models.Task, error) {
	taskInStorage, err := us.GetTaskByID(id)
	if err != nil {
		return models.Task{}, err
	}

	task.UID = taskInStorage.UID

	us.tasks[task.UID] = task
	return task, nil
}

func (us *Storage) DeleteTaskByID(id string) error {
	if _, err := us.GetTaskByID(id); err != nil {
		return err
	}

	delete(us.tasks, id)
	return nil
}
