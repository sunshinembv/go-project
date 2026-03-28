package inmemory

import (
	taskModels "go-project/internal/domain/task/models"
	userModels "go-project/internal/domain/user/models"
)

type Storage struct {
	users map[string]userModels.User
	tasks map[string]taskModels.Task
}

func NewInMemoryStorage() *Storage {
	return &Storage{
		users: make(map[string]userModels.User),
		tasks: make(map[string]taskModels.Task),
	}
}
