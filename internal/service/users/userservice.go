package userservice

import (
	userErrors "go-project/internal/domain/user/errors"
	"go-project/internal/domain/user/models"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	GetUsers() ([]models.User, error)
	GetUserByID(id string) (models.User, error)
	CreateUser(user models.User) error
	UpdateUserByID(id string, userReq models.UserUpdateRequest) (models.User, error)
	DeleteUserByID(id string) error
	GetUserByEmail(email string) (models.User, error)
}

type UserService struct {
	repo  Repository
	valid *validator.Validate
}

func New(repo Repository) *UserService {
	return &UserService{
		repo:  repo,
		valid: validator.New(),
	}
}

func (us *UserService) GetUsers() ([]models.User, error) {
	return us.repo.GetUsers()
}

func (us *UserService) GetUserByID(id string) (models.User, error) {
	return us.repo.GetUserByID(id)
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

	return us.repo.CreateUser(user)
}

func (us *UserService) UpdateUserByID(id string, userReq models.UserUpdateRequest) (models.User, error) {
	return us.repo.UpdateUserByID(id, userReq)
}

func (us *UserService) DeleteUserByID(id string) error {
	return us.repo.DeleteUserByID(id)
}

func (us *UserService) LoginUser(userReq models.UserRequest) (models.User, error) {
	userInMemory, err := us.repo.GetUserByEmail(userReq.Email)
	if err != nil {
		return models.User{}, err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(userInMemory.Password), []byte(userReq.Password)); err != nil {
		return models.User{}, userErrors.ErrInvalidPassword
	}

	return userInMemory, nil
}
