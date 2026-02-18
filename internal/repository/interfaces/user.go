package interfaces

import "go-project/internal/domain/user/models"

type IUserStorage interface {
	GetUsers() ([]models.User, error)
	GetUserByID(id string) (models.User, error)
	CreateUser(user models.User) error
	UpdateUserByID(id string, userReq models.UserUpdateRequest) (models.User, error)
	DeleteUserByID(id string) error
	GetUserByEmail(email string) (models.User, error)
}
