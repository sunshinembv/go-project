package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-project/internal"
	"go-project/internal/domain"
	usersDomain "go-project/internal/domain/user/models"
	taskMocks "go-project/internal/mocks/tasks"
	userMocks "go-project/internal/mocks/users"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	userService := userMocks.NewMockUserService(t)
	taskService := taskMocks.NewMockTaskService(t)
	user := usersDomain.User{
		Name:     "Name",
		Email:    "test@example.com",
		Password: "password123",
	}
	userService.EXPECT().CreateUser(user).Return("user-123", nil)

	srv := New(internal.Config{
		Host: "127.0.0.1",
		Port: "9090",
	}, userService, taskService)

	require.NotNil(t, srv)
	require.NotNil(t, srv.srv)
	assert.Equal(t, "127.0.0.1:9090", srv.srv.Addr)
	assert.Equal(t, domain.ReadHeaderTimeout, srv.srv.ReadHeaderTimeout)
	require.NotNil(t, srv.srv.Handler)

	req := httptest.NewRequest(http.MethodPost, "/users/", strings.NewReader(`{"name":"Name","email":"test@example.com","password":"password123"}`))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	srv.srv.Handler.ServeHTTP(resp, req)
	require.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "user-123")

	req = httptest.NewRequest(http.MethodGet, "/tasks/", nil)
	resp = httptest.NewRecorder()
	srv.srv.Handler.ServeHTTP(resp, req)
	require.Equal(t, http.StatusUnauthorized, resp.Code)

	req = httptest.NewRequest(http.MethodPost, "/refresh", nil)
	resp = httptest.NewRecorder()
	srv.srv.Handler.ServeHTTP(resp, req)
	require.Equal(t, http.StatusUnauthorized, resp.Code)

	require.NoError(t, srv.Shutdown(context.Background()))
}
