package storageinterfaces

import (
	taskservice "go-project/internal/service/tasks"
	userservice "go-project/internal/service/users"
)

type Repositories interface {
	userservice.Repository
	taskservice.Repository
}
