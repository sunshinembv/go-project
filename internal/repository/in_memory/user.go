package inmemory

import (
	"go-project/internal/domain/user/errors"
	"go-project/internal/domain/user/models"

	"github.com/google/uuid"
)

func (us *Storage) GetUsers() ([]models.User, error) {
	users := make([]models.User, 0, len(us.users))
	for _, user := range us.users {
		users = append(users, user)
	}
	return users, nil
}

func (us *Storage) GetUserByUID(uid string) (models.User, error) {
	user, exists := us.users[uid]
	if !exists {
		return models.User{}, errors.ErrUserNotFound
	}
	return user, nil
}

func (us *Storage) CreateUser(user models.User) (uuid.UUID, error) {
	for _, userInStorage := range us.users {
		if user.Email == userInStorage.Email {
			return uuid.Nil, errors.ErrUserAlreadyExists
		}
	}

	uid := uuid.New()
	user.UID = uid.String()

	us.users[user.UID] = user
	return uid, nil
}

func (us *Storage) UpdateUserByUID(uid string, userReq models.UserUpdateRequest) (string, string, error) {
	userInStorage, err := us.GetUserByUID(uid)
	if err != nil {
		return "", "", err
	}

	userInStorage.Name = userReq.Name
	userInStorage.Email = userReq.Email

	us.users[userInStorage.UID] = userInStorage

	return userInStorage.Name, userInStorage.Email, nil
}

func (us *Storage) DeleteUserByUID(uid string) error {
	if _, err := us.GetUserByUID(uid); err != nil {
		return err
	}

	delete(us.users, uid)
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
