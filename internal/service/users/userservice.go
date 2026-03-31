package userservice

import (
	userErrors "go-project/internal/domain/user/errors"
	usersDomain "go-project/internal/domain/user/models"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	GetUsers() ([]usersDomain.User, error)
	GetUserByUID(id string) (usersDomain.User, error)
	CreateUser(user usersDomain.User) (uuid.UUID, error)
	UpdateUserByUID(id string, userReq usersDomain.UserUpdateRequest) (string, string, error)
	DeleteUserByUID(id string) error
	GetUserByEmail(email string) (usersDomain.User, error)
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

func (us *UserService) GetUsers() ([]usersDomain.User, error) {
	return us.repo.GetUsers()
}

func (us *UserService) GetUserByUID(id string) (usersDomain.User, error) {
	return us.repo.GetUserByUID(id)
}

func (us *UserService) CreateUser(user usersDomain.User) (string, error) {
	if err := us.valid.Struct(user); err != nil {
		return "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	user.Password = string(hash)

	uid, err := us.repo.CreateUser(user)
	if err != nil {
		return "", err
	}

	return uid.String(), nil
}

func (us *UserService) UpdateUserByUID(id string, userReq usersDomain.UserUpdateRequest) (string, string, error) {
	return us.repo.UpdateUserByUID(id, userReq)
}

func (us *UserService) DeleteUserByUID(id string) error {
	return us.repo.DeleteUserByUID(id)
}

func (us *UserService) LoginUser(userReq usersDomain.UserRequest) (usersDomain.User, error) {
	storageUser, err := us.repo.GetUserByEmail(userReq.Email)
	if err != nil {
		return usersDomain.User{}, err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(storageUser.Password), []byte(userReq.Password)); err != nil {
		return usersDomain.User{}, userErrors.ErrInvalidPassword
	}

	return storageUser, nil
}
