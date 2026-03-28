package server

import (
	"errors"
	"go-project/internal/domain"
	userErrors "go-project/internal/domain/user/errors"
	"go-project/internal/domain/user/models"
	userService "go-project/internal/service/user_service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (srv *TodoListApi) getUsers(ctx *gin.Context) {
	usecase := userService.NewUserService(srv.db)

	users, err := usecase.GetUsers()
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

func (srv *TodoListApi) getUserByID(ctx *gin.Context) {
	userID := ctx.Param("id")

	usecase := userService.NewUserService(srv.db)
	user, err := usecase.GetUserByID(userID)
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

func (srv *TodoListApi) register(ctx *gin.Context) {
	var user models.User

	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	usecase := userService.NewUserService(srv.db)

	if err := usecase.CreateUser(user); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"msg": "register success"})
}

func (srv *TodoListApi) updateUserByID(ctx *gin.Context) {
	var userUpdateReq models.UserUpdateRequest
	userID := ctx.Param("id")

	if err := ctx.ShouldBindJSON(&userUpdateReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	usecase := userService.NewUserService(srv.db)

	updatedUser, err := usecase.UpdateUserByID(userID, userUpdateReq)
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

func (srv *TodoListApi) deleteUserByID(ctx *gin.Context) {
	userID := ctx.Param("id")

	usecase := userService.NewUserService(srv.db)
	if err := usecase.DeleteUserByID(userID); err != nil {
		if errors.Is(err, userErrors.ErrUserNoExists) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"msg": "user deleted"})
}

func (srv *TodoListApi) login(ctx *gin.Context) {
	var userReq models.UserRequest
	if err := ctx.ShouldBindJSON(&userReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	usecase := userService.NewUserService(srv.db)
	user, err := usecase.LoginUser(userReq)
	if err != nil {
		if errors.Is(err, userErrors.ErrInvalidPassword) || errors.Is(err, userErrors.ErrUserNoExists) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	access, err := srv.jwtSigner.NewAccessToken(user.UID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	refresh, err := srv.jwtSigner.NewRefreshToken(user.UID)
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

func (srv *TodoListApi) profile(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "userID not found in context"})
		return
	}

	usecase := userService.NewUserService(srv.db)
	uid, ok := userID.(string)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "userID is not a string"})
		return
	}
	user, err := usecase.GetUserByID(uid)
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
