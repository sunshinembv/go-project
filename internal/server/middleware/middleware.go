package middleware

import (
	"go-project/internal/domain"
	"go-project/internal/server/auth"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(signer auth.HS256Signer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization header"})
			ctx.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := signer.ParseAccessToken(tokenStr, auth.ParseOptions{
			ExpectedIssuer:   signer.Issuer,
			ExpectedAudience: signer.Audience,
			AllowedMethods:   []string{"HS256"},
			Leeway:           domain.LeewayTimeout,
		})

		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			ctx.Abort()
			return
		}

		ctx.Set("userID", claims.UserID)
		ctx.Next()
	}
}
