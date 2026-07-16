package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-project/internal/service/auth"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func middlewareSigner() auth.HS256Signer {
	return auth.HS256Signer{
		Secret:     []byte("Secret123321"),
		Issuer:     "todo_list-service",
		Audience:   "todo_list-client",
		AccessTTL:  time.Minute,
		RefreshTTL: time.Hour,
	}
}

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	signer := middlewareSigner()
	validToken, err := signer.NewAccessToken("user-123")
	require.NoError(t, err)

	type want struct {
		statusCode int
		userID     string
	}

	type test struct {
		name   string
		header string
		want   want
	}

	tests := []test{
		{
			name:   "valid token",
			header: "Bearer " + validToken,
			want: want{
				statusCode: http.StatusOK,
				userID:     "user-123",
			},
		},
		{
			name: "missing header",
			want: want{
				statusCode: http.StatusUnauthorized,
			},
		},
		{
			name:   "invalid scheme",
			header: "Basic value",
			want: want{
				statusCode: http.StatusUnauthorized,
			},
		},
		{
			name:   "invalid token",
			header: "Bearer invalid",
			want: want{
				statusCode: http.StatusUnauthorized,
			},
		},
		{
			name:   "empty bearer token",
			header: "Bearer ",
			want: want{
				statusCode: http.StatusUnauthorized,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/protected", AuthMiddleware(signer), func(ctx *gin.Context) {
				ctx.JSON(http.StatusOK, gin.H{"userID": ctx.GetString("userID")})
			})

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			require.Equal(t, tc.want.statusCode, resp.Code)
			if tc.want.userID != "" {
				assert.JSONEq(t, `{"userID":"`+tc.want.userID+`"}`, resp.Body.String())
			} else {
				assert.Contains(t, resp.Body.String(), "error")
			}
		})
	}
}
