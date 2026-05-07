package db

import (
	"context"
	"errors"
	usersErrors "go-project/internal/domain/user/errors"
	usersDomain "go-project/internal/domain/user/models"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Storage) GetUsers() ([]usersDomain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := s.conn.Query(ctx, "SELECT uid, name, email, password_hash FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []usersDomain.User
	for rows.Next() {
		var user usersDomain.User
		if err := rows.Scan(
			&user.UID, &user.Name, &user.Email, &user.Password,
		); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (s *Storage) CreateUser(user usersDomain.User) (uuid.UUID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var uid uuid.UUID
	if user.UID == "" {
		err := s.conn.QueryRow(
			ctx,
			`INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, $3)
		ON CONFLICT (uid) DO NOTHING;`,
			user.Name, user.Email, user.Password,
		).Scan(&uid)
		if err != nil {
			return uuid.Nil, err
		}
	} else {
		err := s.conn.QueryRow(
			ctx,
			`INSERT INTO users (uid, name, email, password_hash)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (uid) DO NOTHING;`,
			user.UID, user.Name, user.Email, user.Password,
		).Scan(&uid)
		if err != nil {
			return uuid.Nil, err
		}
	}

	return uid, nil
}

func (s *Storage) UpdateUserByUID(uid string, userReq usersDomain.UserUpdateRequest) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var updatedName string
	var updatedEmail string
	err := s.conn.QueryRow(
		ctx,
		"UPDATE users SET name=$1, email=$2 WHERE uid=$3 RETURNING name, email",
		userReq.Name, userReq.Email, uid).Scan(&updatedName, &updatedEmail)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", usersErrors.ErrUserNotFound
		}
		return "", "", err
	}
	return updatedName, updatedEmail, nil
}

func (s *Storage) DeleteUserByUID(uid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmdTag, err := s.conn.Exec(ctx, "DELETE FROM users WHERE uid=$1", uid)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return usersErrors.ErrUserNotFound
	}
	return nil
}

func (s *Storage) GetUserByEmail(email string) (usersDomain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row := s.conn.QueryRow(ctx, "SELECT uid, name, email, password_hash FROM users WHERE email=$1", email)
	var user usersDomain.User
	if err := row.Scan(&user.UID, &user.Name, &user.Email, &user.Password); err != nil {
		return usersDomain.User{}, err
	}
	return user, nil
}

func (s *Storage) GetUserByUID(uid string) (usersDomain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row := s.conn.QueryRow(ctx, "SELECT uid, name, email, password_hash FROM users WHERE uid=$1", uid)
	var user usersDomain.User
	if err := row.Scan(&user.UID, &user.Name, &user.Email, &user.Password); err != nil {
		return usersDomain.User{}, err
	}
	return user, nil
}
