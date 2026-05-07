package users

import (
	"errors"
	"go-project/internal/domain"
	userErrors "go-project/internal/domain/user/errors"
	usersDomain "go-project/internal/domain/user/models"
	"go-project/internal/service/auth"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserService interface {
	GetUsers() ([]usersDomain.User, error)
	GetUserByUID(id string) (usersDomain.User, error)
	CreateUser(user usersDomain.User) (string, error)
	UpdateUserByUID(id string, userReq usersDomain.UserUpdateRequest) (string, string, error)
	DeleteUserByUID(id string) error
	LoginUser(userReq usersDomain.UserRequest) (usersDomain.User, error)
}

type UsersHandler struct {
	userService UserService
	jwtSigner   auth.HS256Signer
}

func New(userService UserService, jwtSigner auth.HS256Signer) *UsersHandler {
	return &UsersHandler{
		userService: userService,
		jwtSigner:   jwtSigner,
	}
}

func (uh *UsersHandler) GetUsers(ctx *gin.Context) {

	users, err := uh.userService.GetUsers()
	if err != nil {
		if errors.Is(err, userErrors.ErrUserNoExists) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"users": users})
}

func (uh *UsersHandler) GetUserByID(ctx *gin.Context) {
	userID := ctx.Param("id")

	user, err := uh.userService.GetUserByUID(userID)
	if err != nil {
		if errors.Is(err, userErrors.ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"user": user})
}

func (uh *UsersHandler) Register(ctx *gin.Context) {
	var user usersDomain.User

	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	uid, err := uh.userService.CreateUser(user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"uid": uid})
}

func (uh *UsersHandler) UpdateUserByID(ctx *gin.Context) {
	var userUpdateReq usersDomain.UserUpdateRequest
	userID := ctx.Param("id")

	if err := ctx.ShouldBindJSON(&userUpdateReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	updatedName, updatedEmail, err := uh.userService.UpdateUserByUID(userID, userUpdateReq)
	if err != nil {
		if errors.Is(err, userErrors.ErrUserNoExists) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"updatedName":  updatedName,
		"updatedEmail": updatedEmail,
	})
}

func (uh *UsersHandler) DeleteUserByID(ctx *gin.Context) {
	userID := ctx.Param("id")

	if err := uh.userService.DeleteUserByUID(userID); err != nil {
		if errors.Is(err, userErrors.ErrUserNoExists) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"msg": "user deleted"})
}

func (uh *UsersHandler) Login(ctx *gin.Context) {
	var userReq usersDomain.UserRequest
	if err := ctx.ShouldBindJSON(&userReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := uh.userService.LoginUser(userReq)
	if err != nil {
		if errors.Is(err, userErrors.ErrInvalidPassword) || errors.Is(err, userErrors.ErrUserNoExists) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	access, err := uh.jwtSigner.NewAccessToken(user.UID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	refresh, err := uh.jwtSigner.NewRefreshToken(user.UID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.SetCookie("refresh_token", refresh, domain.CoockeiMaxAge, "/", "localhost", false, true)
	ctx.JSON(http.StatusOK, gin.H{
		"user":   user,
		"access": access,
	})
}

func (uh *UsersHandler) Profile(ctx *gin.Context) {
	uid := ctx.GetString("userID")
	if uid == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := uh.userService.GetUserByUID(uid)
	if err != nil {
		if errors.Is(err, userErrors.ErrUserNoExists) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (uh *UsersHandler) Refresh(ctx *gin.Context) {
	refreshToken, err := ctx.Cookie("refresh_token")
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	claims, err := uh.jwtSigner.ParseRefreshToken(refreshToken, auth.ParseOptions{
		ExpectedIssuer:   uh.jwtSigner.Issuer,
		ExpectedAudience: uh.jwtSigner.Audience,
		AllowedMethods:   []string{"HS256"},
		Leeway:           domain.LeewayTimeout,
	})
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	access, err := uh.jwtSigner.NewAccessToken(claims.Subject)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	newRefresh, err := uh.jwtSigner.NewRefreshToken(claims.Subject)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.SetCookie("refresh_token", newRefresh, domain.CoockeiMaxAge, "/", "localhost", false, true)
	ctx.JSON(http.StatusOK, gin.H{"access": access})
}
