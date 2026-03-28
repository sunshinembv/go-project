package userservice

import (
	userErrors "go-project/internal/domain/user/errors"
	"go-project/internal/domain/user/models"
	"go-project/internal/repository/interfaces"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	db    interfaces.IUserStorage
	valid *validator.Validate
}

func NewUserService(db interfaces.IUserStorage) *UserService {
	return &UserService{
		db:    db,
		valid: validator.New(),
	}
}

func (us *UserService) GetUsers() ([]models.User, error) {
	return us.db.GetUsers()
}

func (us *UserService) GetUserByID(id string) (models.User, error) {
	return us.db.GetUserByID(id)
}

func (us *UserService) CreateUser(user models.User) error {
	if err := us.valid.Struct(user); err != nil {
		return err
	}

	uid := uuid.New().String()
	user.UID = uid

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hash)

	return us.db.CreateUser(user)
}

func (us *UserService) UpdateUserByID(id string, userReq models.UserUpdateRequest) (models.User, error) {
	return us.db.UpdateUserByID(id, userReq)
}

func (us *UserService) DeleteUserByID(id string) error {
	return us.db.DeleteUserByID(id)
}

func (us *UserService) LoginUser(userReq models.UserRequest) (models.User, error) {
	userInMemory, err := us.db.GetUserByEmail(userReq.Email)
	if err != nil {
		return models.User{}, err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(userInMemory.Password), []byte(userReq.Password)); err != nil {
		return models.User{}, userErrors.ErrInvalidPassword
	}

	return userInMemory, nil
}
