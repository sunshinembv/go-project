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
	GetUserByID(id string) (usersDomain.User, error)
	CreateUser(user usersDomain.User) error
	UpdateUserByID(id string, userReq usersDomain.UserUpdateRequest) (usersDomain.User, error)
	DeleteUserByID(id string) error
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

	user, err := uh.userService.GetUserByID(userID)
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

	if err := uh.userService.CreateUser(user); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"msg": "register success"})
}

func (uh *UsersHandler) UpdateUserByID(ctx *gin.Context) {
	var userUpdateReq usersDomain.UserUpdateRequest
	userID := ctx.Param("id")

	if err := ctx.ShouldBindJSON(&userUpdateReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	updatedUser, err := uh.userService.UpdateUserByID(userID, userUpdateReq)
	if err != nil {
		if errors.Is(err, userErrors.ErrUserNoExists) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"user": updatedUser})
}

func (uh *UsersHandler) DeleteUserByID(ctx *gin.Context) {
	userID := ctx.Param("id")

	if err := uh.userService.DeleteUserByID(userID); err != nil {
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
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "userID not found in context"})
		return
	}

	uid, ok := userID.(string)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "userID is not a string"})
		return
	}
	user, err := uh.userService.GetUserByID(uid)
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
