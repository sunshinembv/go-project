package errors

import "errors"

var (
	ErrUserAlreadyExists = errors.New("user with this email already exists")
	ErrUserNoExists      = errors.New("user with this email no exists")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrUserNotFound      = errors.New("user not found")
)
