package storageinterfaces

import (
	taskservice "go-project/internal/service/tasks"
	userservice "go-project/internal/service/users"
	"io"
)

type Repositories interface {
	userservice.Repository
	taskservice.Repository
	io.Closer
}
