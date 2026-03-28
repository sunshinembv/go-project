package inmemory

import (
	"go-project/internal/domain/user/errors"
	"go-project/internal/domain/user/models"
)

func (us *Storage) GetUsers() ([]models.User, error) {
	users := make([]models.User, 0, len(us.users))
	for _, user := range us.users {
		users = append(users, user)
	}
	return users, nil
}

func (us *Storage) GetUserByID(id string) (models.User, error) {
	user, exists := us.users[id]
	if !exists {
		return models.User{}, errors.ErrUserNotFound
	}
	return user, nil
}

func (us *Storage) CreateUser(user models.User) error {
	for _, userInStorage := range us.users {
		if user.Email == userInStorage.Email {
			return errors.ErrUserAlreadyExists
		}
	}
	us.users[user.UID] = user
	return nil
}

func (us *Storage) UpdateUserByID(id string, userReq models.UserUpdateRequest) (models.User, error) {
	userInStorage, err := us.GetUserByID(id)
	if err != nil {
		return models.User{}, err
	}

	userInStorage.Name = userReq.Name
	userInStorage.Email = userReq.Email

	us.users[userInStorage.UID] = userInStorage

	return userInStorage, nil
}

func (us *Storage) DeleteUserByID(id string) error {
	if _, err := us.GetUserByID(id); err != nil {
		return err
	}

	delete(us.users, id)
	return nil
}

func (us *Storage) GetUserByEmail(email string) (models.User, error) {
	for _, userInStorage := range us.users {
		if email == userInStorage.Email {
			return userInStorage, nil
		}
	}
	return models.User{}, errors.ErrUserNoExists
}
