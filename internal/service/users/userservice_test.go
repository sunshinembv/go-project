package userservice

import (
	"errors"
	"strings"
	"testing"

	userErrors "go-project/internal/domain/user/errors"
	usersDomain "go-project/internal/domain/user/models"
	userMocks "go-project/internal/mocks/userservice"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

var errRepository = errors.New("repository error")

func TestGetUsers(t *testing.T) {
	type want struct {
		users []usersDomain.User
		err   error
	}

	type test struct {
		name string
		want want
	}

	tests := []test{
		{
			name: "success",
			want: want{
				users: []usersDomain.User{
					{
						UID:   "user-123",
						Name:  "Name",
						Email: "test@example.com",
					},
				},
			},
		},
		{
			name: "repository error",
			want: want{
				err: errRepository,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := userMocks.NewMockRepository(t)
			repo.EXPECT().GetUsers().Return(tc.want.users, tc.want.err)

			users, err := New(repo).GetUsers()

			require.ErrorIs(t, err, tc.want.err)
			assert.Equal(t, tc.want.users, users)
		})
	}
}

func TestGetUserByUID(t *testing.T) {
	type want struct {
		user usersDomain.User
		err  error
	}

	type test struct {
		name   string
		userID string
		want   want
	}

	tests := []test{
		{
			name:   "success",
			userID: "user-123",
			want: want{
				user: usersDomain.User{
					UID:   "user-123",
					Name:  "Name",
					Email: "test@example.com",
				},
			},
		},
		{
			name:   "repository error",
			userID: "user-456",
			want: want{
				err: errRepository,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := userMocks.NewMockRepository(t)
			repo.EXPECT().GetUserByUID(tc.userID).Return(tc.want.user, tc.want.err)

			user, err := New(repo).GetUserByUID(tc.userID)

			require.ErrorIs(t, err, tc.want.err)
			assert.Equal(t, tc.want.user, user)
		})
	}
}

func TestUpdateUserByUID(t *testing.T) {
	type want struct {
		name  string
		email string
		err   error
	}

	type test struct {
		name   string
		userID string
		req    usersDomain.UserUpdateRequest
		want   want
	}

	tests := []test{
		{
			name:   "success",
			userID: "user-123",
			req: usersDomain.UserUpdateRequest{
				Name:  "Updated",
				Email: "updated@example.com",
			},
			want: want{
				name:  "Updated",
				email: "updated@example.com",
			},
		},
		{
			name:   "repository error",
			userID: "user-456",
			req: usersDomain.UserUpdateRequest{
				Name:  "Updated",
				Email: "updated@example.com",
			},
			want: want{
				err: errRepository,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := userMocks.NewMockRepository(t)
			repo.EXPECT().UpdateUserByUID(tc.userID, tc.req).Return(
				tc.want.name,
				tc.want.email,
				tc.want.err,
			)

			name, email, err := New(repo).UpdateUserByUID(tc.userID, tc.req)

			require.ErrorIs(t, err, tc.want.err)
			assert.Equal(t, tc.want.name, name)
			assert.Equal(t, tc.want.email, email)
		})
	}
}

func TestDeleteUserByUID(t *testing.T) {
	type want struct {
		err error
	}

	type test struct {
		name   string
		userID string
		want   want
	}

	tests := []test{
		{
			name:   "success",
			userID: "user-123",
			want: want{
				err: nil,
			},
		},
		{
			name:   "repository error",
			userID: "user-456",
			want: want{
				err: errRepository,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := userMocks.NewMockRepository(t)
			repo.EXPECT().DeleteUserByUID(tc.userID).Return(tc.want.err)

			err := New(repo).DeleteUserByUID(tc.userID)

			require.ErrorIs(t, err, tc.want.err)
		})
	}
}

func TestCreateUser(t *testing.T) {
	uid := uuid.MustParse("7f676b16-2db7-4ed5-9afc-9b4e54f95563")

	type want struct {
		repoCalled bool
		repoErr    error
		err        bool
	}

	type test struct {
		name string
		user usersDomain.User
		want want
	}

	tests := []test{
		{
			name: "success",
			user: usersDomain.User{
				Name:     "Name",
				Email:    "test@example.com",
				Password: "password123",
			},
			want: want{
				repoCalled: true,
			},
		},
		{
			name: "invalid email",
			user: usersDomain.User{
				Name:     "Name",
				Email:    "invalid",
				Password: "password123",
			},
			want: want{
				err: true,
			},
		},
		{
			name: "short password",
			user: usersDomain.User{
				Name:     "Name",
				Email:    "test@example.com",
				Password: "short",
			},
			want: want{
				err: true,
			},
		},
		{
			name: "bcrypt error",
			user: usersDomain.User{
				Name:     "Name",
				Email:    "test@example.com",
				Password: strings.Repeat("a", 73),
			},
			want: want{
				err: true,
			},
		},
		{
			name: "repository error",
			user: usersDomain.User{
				Name:     "Name",
				Email:    "test@example.com",
				Password: "password123",
			},
			want: want{
				repoCalled: true,
				repoErr:    errRepository,
				err:        true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := userMocks.NewMockRepository(t)
			var stored usersDomain.User
			if tc.want.repoCalled {
				repo.EXPECT().CreateUser(mock.MatchedBy(func(user usersDomain.User) bool {
					stored = user
					return user.Name == tc.user.Name && user.Email == tc.user.Email
				})).Return(uid, tc.want.repoErr)
			}

			got, err := New(repo).CreateUser(tc.user)

			if tc.want.err {
				require.Error(t, err)
				assert.Empty(t, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, uid.String(), got)
			require.NoError(t, bcrypt.CompareHashAndPassword([]byte(stored.Password), []byte(tc.user.Password)))
			assert.NotEqual(t, tc.user.Password, stored.Password)
		})
	}
}

func TestLoginUser(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)
	storageUser := usersDomain.User{
		UID:      "user-123",
		Email:    "test@example.com",
		Password: string(hash),
	}

	type want struct {
		user usersDomain.User
		err  error
	}

	type test struct {
		name     string
		password string
		repoErr  error
		want     want
	}

	tests := []test{
		{
			name:     "success",
			password: "password123",
			want: want{
				user: storageUser,
			},
		},
		{
			name:     "invalid password",
			password: "wrong-password",
			want: want{
				user: storageUser,
				err:  userErrors.ErrInvalidPassword,
			},
		},
		{
			name:     "repository error",
			password: "password123",
			repoErr:  errRepository,
			want: want{
				err: errRepository,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := usersDomain.UserRequest{
				Email:    "test@example.com",
				Password: tc.password,
			}
			repo := userMocks.NewMockRepository(t)
			repo.EXPECT().GetUserByEmail(req.Email).Return(tc.want.user, tc.repoErr)

			got, err := New(repo).LoginUser(req)

			if tc.want.err != nil {
				require.ErrorIs(t, err, tc.want.err)
				assert.Empty(t, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, storageUser, got)
		})
	}
}
