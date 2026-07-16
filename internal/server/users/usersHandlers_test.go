package users

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-project/internal/domain"
	userErrors "go-project/internal/domain/user/errors"
	usersDomain "go-project/internal/domain/user/models"
	userMocks "go-project/internal/mocks/users"
	"go-project/internal/service/auth"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	ErrInvalidToken    = errors.New("invalid token")
	ErrInvalidLogin    = errors.New("invalid login")
	ErrInvalidRegister = errors.New("invalid register")
	ErrUserService     = errors.New("service error")
)

func TestLogin(t *testing.T) {
	type want struct {
		statusCode        int
		validAccessToken  bool
		validRefreshToken bool
		user              usersDomain.User
		cookie            string
		err               error
	}

	type test struct {
		name        string
		req         string
		method      string
		callService bool
		want        want
	}

	validUser := usersDomain.User{
		UID:      "user-123",
		Name:     "Name1",
		Email:    "test@example.com",
		Password: "password123",
	}

	tests := []test{
		{
			name:        "valid login",
			req:         `{"email":"test@example.com","password":"password123"}`,
			method:      http.MethodPost,
			callService: true,
			want: want{
				statusCode:        http.StatusOK,
				validAccessToken:  true,
				validRefreshToken: true,
				user:              validUser,
				cookie:            "refresh_token",
			},
		},
		{
			name:   "invalid json",
			req:    `{invalid`,
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:        "invalid password",
			req:         `{"email":"test@example.com","password":"password123"}`,
			method:      http.MethodPost,
			callService: true,
			want: want{
				statusCode: http.StatusUnauthorized,
				err:        userErrors.ErrInvalidPassword,
			},
		},
		{
			name:        "user does not exist",
			req:         `{"email":"test@example.com","password":"password123"}`,
			method:      http.MethodPost,
			callService: true,
			want: want{
				statusCode: http.StatusUnauthorized,
				err:        userErrors.ErrUserNoExists,
			},
		},
		{
			name:        "service error",
			req:         `{"email":"test@example.com","password":"password123"}`,
			method:      http.MethodPost,
			callService: true,
			want: want{
				statusCode: http.StatusInternalServerError,
				err:        ErrUserService,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			userServiceMock := userMocks.NewMockUserService(t)

			if tc.callService {
				var reqObj usersDomain.UserRequest
				err := json.Unmarshal([]byte(tc.req), &reqObj)
				require.NoError(t, err)

				userServiceMock.EXPECT().
					LoginUser(reqObj).
					Return(tc.want.user, tc.want.err)
			}

			signer := auth.HS256Signer{
				Secret:     []byte("Secret123321"),
				Issuer:     "todo_list-service",
				Audience:   "todo_list-client",
				AccessTTL:  domain.AccessTTL,
				RefreshTTL: domain.RefreshTTL,
			}

			opt := auth.ParseOptions{
				ExpectedIssuer:   signer.Issuer,
				ExpectedAudience: signer.Audience,
				AllowedMethods:   []string{"HS256"},
				Leeway:           domain.LeewayTimeout,
			}

			gin.SetMode(gin.ReleaseMode)
			router := gin.New()
			us := New(userServiceMock, signer)
			router.POST("/login", us.Login)

			req := httptest.NewRequest(tc.method, "/login", strings.NewReader(tc.req))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			require.Equal(t, tc.want.statusCode, resp.Code)

			if tc.want.statusCode != http.StatusOK {
				type errorResponse struct {
					Error string `json:"error"`
				}

				var respData errorResponse
				require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &respData))
				if tc.want.err != nil {
					assert.Equal(t, tc.want.err.Error(), respData.Error)
				} else {
					assert.NotEmpty(t, respData.Error)
				}
				return
			}

			type loginResponse struct {
				AccessToken string `json:"access"`
			}

			var respData loginResponse
			require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &respData))
			require.NotEmpty(t, respData.AccessToken)

			claims, err := signer.ParseAccessToken(respData.AccessToken, opt)
			if tc.want.validAccessToken {
				require.NoError(t, err)
				require.NotNil(t, claims)
				assert.Equal(t, tc.want.user.UID, claims.UserID)
			} else {
				require.ErrorIs(t, err, auth.ErrInvalidToken)
			}

			cookies := resp.Result().Cookies()
			require.NotEmpty(t, cookies)
			assert.Equal(t, tc.want.cookie, cookies[0].Name)

			refreshClaims, err := signer.ParseRefreshToken(cookies[0].Value, opt)
			if tc.want.validRefreshToken {
				require.NoError(t, err)
				require.NotNil(t, refreshClaims)
				assert.Equal(t, tc.want.user.UID, refreshClaims.Subject)
			} else {
				require.ErrorIs(t, err, auth.ErrInvalidToken)
			}
		})
	}
}

func TestGetUsers(t *testing.T) {
	type want struct {
		statusCode int
		users      []usersDomain.User
		err        error
	}

	type test struct {
		name   string
		method string
		want   want
	}

	tests := []test{
		{
			name:   "valid get users",
			method: http.MethodGet,
			want: want{
				statusCode: http.StatusOK,
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
			name:   "users not found",
			method: http.MethodGet,
			want: want{
				statusCode: http.StatusNotFound,
				err:        userErrors.ErrUserNoExists,
			},
		},
		{
			name:   "service error",
			method: http.MethodGet,
			want: want{
				statusCode: http.StatusInternalServerError,
				err:        ErrUserService,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			userServiceMock := userMocks.NewMockUserService(t)
			userServiceMock.EXPECT().
				GetUsers().
				Return(tc.want.users, tc.want.err)

			gin.SetMode(gin.ReleaseMode)
			router := gin.New()
			us := New(userServiceMock, auth.HS256Signer{})
			router.GET("/users", us.GetUsers)

			req := httptest.NewRequest(tc.method, "/users", nil)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			require.Equal(t, tc.want.statusCode, resp.Code)

			if tc.want.err != nil {
				type errorResponse struct {
					Error string `json:"error"`
				}

				var respData errorResponse
				require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &respData))
				assert.Equal(t, tc.want.err.Error(), respData.Error)
				return
			}

			type usersResponse struct {
				Users []usersDomain.User `json:"users"`
			}

			var respData usersResponse
			require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &respData))
			assert.Equal(t, tc.want.users, respData.Users)
		})
	}
}

func TestGetUserByID(t *testing.T) {
	type want struct {
		statusCode int
		user       usersDomain.User
		err        error
	}

	type test struct {
		name   string
		userID string
		method string
		want   want
	}

	tests := []test{
		{
			name:   "valid get user",
			userID: "user-123",
			method: http.MethodGet,
			want: want{
				statusCode: http.StatusOK,
				user: usersDomain.User{
					UID: "user-123", Name: "Name", Email: "test@example.com",
				},
			},
		},
		{
			name:   "user not found",
			userID: "user-123",
			method: http.MethodGet,
			want: want{
				statusCode: http.StatusNotFound,
				err:        userErrors.ErrUserNotFound,
			},
		},
		{
			name:   "service error",
			userID: "user-123",
			method: http.MethodGet,
			want: want{
				statusCode: http.StatusInternalServerError,
				err:        ErrUserService,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			userServiceMock := userMocks.NewMockUserService(t)
			userServiceMock.EXPECT().
				GetUserByUID(tc.userID).
				Return(tc.want.user, tc.want.err)

			gin.SetMode(gin.ReleaseMode)
			router := gin.New()
			us := New(userServiceMock, auth.HS256Signer{})
			router.GET("/users/:id", us.GetUserByID)

			req := httptest.NewRequest(tc.method, "/users/"+tc.userID, nil)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			require.Equal(t, tc.want.statusCode, resp.Code)

			if tc.want.err != nil {
				type errorResponse struct {
					Error string `json:"error"`
				}

				var respData errorResponse
				require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &respData))
				assert.Equal(t, tc.want.err.Error(), respData.Error)
				return
			}

			type userResponse struct {
				User usersDomain.User `json:"user"`
			}

			var respData userResponse
			require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &respData))
			assert.Equal(t, tc.want.user, respData.User)
		})
	}
}

func TestRegister(t *testing.T) {
	type want struct {
		statusCode int
		uid        string
		err        error
	}

	type test struct {
		name        string
		req         string
		method      string
		callService bool
		want        want
	}

	validRequest := `{"name":"Name","email":"test@example.com","password":"password123"}`

	tests := []test{
		{
			name:        "valid register",
			req:         validRequest,
			method:      http.MethodPost,
			callService: true,
			want: want{
				statusCode: http.StatusOK,
				uid:        "user-123",
			},
		},
		{
			name:   "invalid json",
			req:    `{invalid`,
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:        "service error",
			req:         validRequest,
			method:      http.MethodPost,
			callService: true,
			want: want{
				statusCode: http.StatusInternalServerError,
				err:        ErrUserService,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			userServiceMock := userMocks.NewMockUserService(t)

			if tc.callService {
				var reqObj usersDomain.User
				err := json.Unmarshal([]byte(tc.req), &reqObj)
				require.NoError(t, err)

				userServiceMock.EXPECT().
					CreateUser(reqObj).
					Return(tc.want.uid, tc.want.err)
			}

			gin.SetMode(gin.ReleaseMode)
			router := gin.New()
			us := New(userServiceMock, auth.HS256Signer{})
			router.POST("/users", us.Register)

			req := httptest.NewRequest(tc.method, "/users", strings.NewReader(tc.req))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			require.Equal(t, tc.want.statusCode, resp.Code)

			if tc.want.err != nil {
				type errorResponse struct {
					Error string `json:"error"`
				}

				var respData errorResponse
				require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &respData))
				assert.Equal(t, tc.want.err.Error(), respData.Error)
				return
			}

			if tc.want.uid != "" {
				type registerResponse struct {
					UID string `json:"uid"`
				}

				var respData registerResponse
				require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &respData))
				assert.Equal(t, tc.want.uid, respData.UID)
			}
		})
	}
}

func TestUpdateUserByID(t *testing.T) {
	type want struct {
		statusCode  int
		updatedName string
		updatedMail string
		err         error
	}

	type test struct {
		name        string
		userID      string
		req         string
		method      string
		callService bool
		want        want
	}

	validRequest := `{"name":"Updated","email":"updated@example.com"}`

	tests := []test{
		{
			name:        "valid update",
			userID:      "user-123",
			req:         validRequest,
			method:      http.MethodPut,
			callService: true,
			want: want{
				statusCode:  http.StatusOK,
				updatedName: "Updated",
				updatedMail: "updated@example.com",
			},
		},
		{
			name:   "invalid json",
			userID: "user-123",
			req:    `{invalid`,
			method: http.MethodPut,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:        "user not found",
			userID:      "user-123",
			req:         validRequest,
			method:      http.MethodPut,
			callService: true,
			want: want{
				statusCode: http.StatusNotFound,
				err:        userErrors.ErrUserNoExists,
			},
		},
		{
			name:        "service error",
			userID:      "user-123",
			req:         validRequest,
			method:      http.MethodPut,
			callService: true,
			want: want{
				statusCode: http.StatusInternalServerError,
				err:        ErrUserService,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			userServiceMock := userMocks.NewMockUserService(t)

			if tc.callService {
				var reqObj usersDomain.UserUpdateRequest
				err := json.Unmarshal([]byte(tc.req), &reqObj)
				require.NoError(t, err)

				userServiceMock.EXPECT().
					UpdateUserByUID(tc.userID, reqObj).
					Return(tc.want.updatedName, tc.want.updatedMail, tc.want.err)
			}

			gin.SetMode(gin.ReleaseMode)
			router := gin.New()
			us := New(userServiceMock, auth.HS256Signer{})
			router.PUT("/users/:id", us.UpdateUserByID)

			req := httptest.NewRequest(tc.method, "/users/"+tc.userID, strings.NewReader(tc.req))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			require.Equal(t, tc.want.statusCode, resp.Code)

			if tc.want.err != nil {
				type errorResponse struct {
					Error string `json:"error"`
				}

				var respData errorResponse
				require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &respData))
				assert.Equal(t, tc.want.err.Error(), respData.Error)
				return
			}

			if tc.callService {
				type updateResponse struct {
					UpdatedName  string `json:"updatedName"`
					UpdatedEmail string `json:"updatedEmail"`
				}

				var respData updateResponse
				require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &respData))
				assert.Equal(t, tc.want.updatedName, respData.UpdatedName)
				assert.Equal(t, tc.want.updatedMail, respData.UpdatedEmail)
			}
		})
	}
}

func TestDeleteUserByID(t *testing.T) {
	type want struct {
		statusCode int
		message    string
		err        error
	}

	type test struct {
		name   string
		userID string
		method string
		want   want
	}

	tests := []test{
		{
			name:   "valid delete",
			userID: "user-123",
			method: http.MethodDelete,
			want: want{
				statusCode: http.StatusOK,
				message:    "user deleted",
			},
		},
		{
			name:   "user not found",
			userID: "user-123",
			method: http.MethodDelete,
			want: want{
				statusCode: http.StatusNotFound,
				err:        userErrors.ErrUserNoExists,
			},
		},
		{
			name:   "service error",
			userID: "user-123",
			method: http.MethodDelete,
			want: want{
				statusCode: http.StatusInternalServerError,
				err:        ErrUserService,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			userServiceMock := userMocks.NewMockUserService(t)
			userServiceMock.EXPECT().
				DeleteUserByUID(tc.userID).
				Return(tc.want.err)

			gin.SetMode(gin.ReleaseMode)
			router := gin.New()
			us := New(userServiceMock, auth.HS256Signer{})
			router.DELETE("/users/:id", us.DeleteUserByID)

			req := httptest.NewRequest(tc.method, "/users/"+tc.userID, nil)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			require.Equal(t, tc.want.statusCode, resp.Code)

			if tc.want.err != nil {
				type errorResponse struct {
					Error string `json:"error"`
				}

				var respData errorResponse
				require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &respData))
				assert.Equal(t, tc.want.err.Error(), respData.Error)
				return
			}

			type deleteResponse struct {
				Message string `json:"msg"`
			}

			var respData deleteResponse
			require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &respData))
			assert.Equal(t, tc.want.message, respData.Message)
		})
	}
}

func TestProfile(t *testing.T) {
	type want struct {
		statusCode int
		user       usersDomain.User
		err        error
	}

	type test struct {
		name        string
		userID      string
		method      string
		callService bool
		want        want
	}

	tests := []test{
		{
			name:        "valid profile",
			userID:      "user-123",
			method:      http.MethodGet,
			callService: true,
			want: want{
				statusCode: http.StatusOK,
				user: usersDomain.User{
					UID: "user-123", Name: "Name", Email: "test@example.com",
				},
			},
		},
		{
			name:   "unauthorized",
			method: http.MethodGet,
			want: want{
				statusCode: http.StatusUnauthorized,
			},
		},
		{
			name:        "user not found",
			userID:      "user-123",
			method:      http.MethodGet,
			callService: true,
			want: want{
				statusCode: http.StatusNotFound,
				err:        userErrors.ErrUserNoExists,
			},
		},
		{
			name:        "service error",
			userID:      "user-123",
			method:      http.MethodGet,
			callService: true,
			want: want{
				statusCode: http.StatusInternalServerError,
				err:        ErrUserService,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			userServiceMock := userMocks.NewMockUserService(t)

			if tc.callService {
				userServiceMock.EXPECT().
					GetUserByUID(tc.userID).
					Return(tc.want.user, tc.want.err)
			}

			gin.SetMode(gin.ReleaseMode)
			router := gin.New()
			router.Use(func(ctx *gin.Context) {
				if tc.userID != "" {
					ctx.Set("userID", tc.userID)
				}
				ctx.Next()
			})
			us := New(userServiceMock, auth.HS256Signer{})
			router.GET("/profile", us.Profile)

			req := httptest.NewRequest(tc.method, "/profile", nil)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			require.Equal(t, tc.want.statusCode, resp.Code)

			if tc.want.err != nil {
				type errorResponse struct {
					Error string `json:"error"`
				}

				var respData errorResponse
				require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &respData))
				assert.Equal(t, tc.want.err.Error(), respData.Error)
				return
			}

			if tc.callService {
				var respData usersDomain.User
				require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &respData))
				assert.Equal(t, tc.want.user, respData)
			}
		})
	}
}

func TestRefresh(t *testing.T) {
	type want struct {
		statusCode int
		validToken bool
		userID     string
		cookie     string
	}

	type test struct {
		name   string
		method string
		cookie *http.Cookie
		want   want
	}

	signer := auth.HS256Signer{
		Secret:     []byte("Secret123321"),
		Issuer:     "todo_list-service",
		Audience:   "todo_list-client",
		AccessTTL:  domain.AccessTTL,
		RefreshTTL: domain.RefreshTTL,
	}

	refreshToken, err := signer.NewRefreshToken("user-123")
	require.NoError(t, err)

	tests := []test{
		{
			name:   "valid refresh",
			method: http.MethodPost,
			cookie: &http.Cookie{Name: "refresh_token", Value: refreshToken},
			want: want{
				statusCode: http.StatusOK,
				validToken: true,
				userID:     "user-123",
				cookie:     "refresh_token",
			},
		},
		{
			name:   "missing cookie",
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusUnauthorized,
			},
		},
		{
			name:   "invalid refresh token",
			method: http.MethodPost,
			cookie: &http.Cookie{Name: "refresh_token", Value: "invalid"},
			want: want{
				statusCode: http.StatusUnauthorized,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			userServiceMock := userMocks.NewMockUserService(t)

			gin.SetMode(gin.ReleaseMode)
			router := gin.New()
			us := New(userServiceMock, signer)
			router.POST("/refresh", us.Refresh)

			req := httptest.NewRequest(tc.method, "/refresh", nil)
			if tc.cookie != nil {
				req.AddCookie(tc.cookie)
			}
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			require.Equal(t, tc.want.statusCode, resp.Code)

			if !tc.want.validToken {
				type errorResponse struct {
					Error string `json:"error"`
				}

				var respData errorResponse
				require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &respData))
				assert.NotEmpty(t, respData.Error)
				return
			}

			type refreshResponse struct {
				AccessToken string `json:"access"`
			}

			var respData refreshResponse
			require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &respData))
			require.NotEmpty(t, respData.AccessToken)

			claims, err := signer.ParseAccessToken(respData.AccessToken, auth.ParseOptions{
				ExpectedIssuer:   signer.Issuer,
				ExpectedAudience: signer.Audience,
				AllowedMethods:   []string{"HS256"},
				Leeway:           domain.LeewayTimeout,
			})
			require.NoError(t, err)
			assert.Equal(t, tc.want.userID, claims.UserID)

			cookies := resp.Result().Cookies()
			require.NotEmpty(t, cookies)
			assert.Equal(t, tc.want.cookie, cookies[0].Name)
			assert.NotEqual(t, refreshToken, cookies[0].Value)
		})
	}
}
