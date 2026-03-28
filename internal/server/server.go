package server

import (
	"fmt"
	"go-project/internal"
	"go-project/internal/domain"
	"go-project/internal/repository/interfaces"
	"go-project/internal/server/auth"
	"go-project/internal/server/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TodoListApi struct {
	srv       *http.Server
	db        interfaces.IStorage
	jwtSigner auth.HS256Signer
}

func NewServer(cgf internal.Config, db interfaces.IStorage) *TodoListApi {
	signer := auth.HS256Signer{
		Secret:     []byte("Secret123321"),
		Issuer:     "todo_list-service",
		Audience:   "todo_list-client",
		AccessTTL:  domain.AccessTTL,
		RefreshTTL: domain.RefreshTTL,
	}
	httpSrv := http.Server{
		Addr:              fmt.Sprintf("%s:%d", cgf.Host, cgf.Port),
		ReadHeaderTimeout: domain.ReadHeaderTimeout,
	}

	api := TodoListApi{
		srv:       &httpSrv,
		db:        db,
		jwtSigner: signer,
	}

	api.configRouter()

	fmt.Printf("Server addr %s", fmt.Sprintf("%s:%d", cgf.Host, cgf.Port))

	return &api
}

func (api *TodoListApi) Run() error {
	return api.srv.ListenAndServe()
}

func (api *TodoListApi) Shutdown() error {
	return nil
}

func (api *TodoListApi) configRouter() {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	users := router.Group("/users")

	users.POST("/", api.register)
	users.POST("/login", api.login)
	users.GET("/profile", middleware.AuthMiddleware(api.jwtSigner), api.profile)

	users.GET("/", middleware.AuthMiddleware(api.jwtSigner), api.getUsers)
	users.GET("/:id", middleware.AuthMiddleware(api.jwtSigner), api.getUserByID)
	users.PUT("/:id", middleware.AuthMiddleware(api.jwtSigner), api.updateUserByID)
	users.DELETE("/:id", middleware.AuthMiddleware(api.jwtSigner), api.deleteUserByID)

	tasks := router.Group("/tasks")

	tasks.GET("/", middleware.AuthMiddleware(api.jwtSigner), api.getTasks)
	tasks.GET("/:id", middleware.AuthMiddleware(api.jwtSigner), api.getTaskByID)
	tasks.POST("/", middleware.AuthMiddleware(api.jwtSigner), api.createTask)
	tasks.PUT("/:id", middleware.AuthMiddleware(api.jwtSigner), api.updateTaskByID)
	tasks.DELETE("/:id", middleware.AuthMiddleware(api.jwtSigner), api.deleteTaskByID)

	router.POST("/refresh", api.refresh)

	api.srv.Handler = router
}

func (srv *TodoListApi) refresh(ctx *gin.Context) {
	refreshToken, err := ctx.Cookie("refresh_token")
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	claims, err := srv.jwtSigner.ParseRefreshToken(refreshToken, auth.ParseOptions{
		ExpectedIssuer:   srv.jwtSigner.Issuer,
		ExpectedAudience: srv.jwtSigner.Audience,
		AllowedMethods:   []string{"HS256"},
		Leeway:           domain.LeewayTimeout,
	})
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	access, err := srv.jwtSigner.NewAccessToken(claims.Subject)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	newRefresh, err := srv.jwtSigner.NewRefreshToken(claims.Subject)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.SetCookie("refresh_token", newRefresh, domain.CoockeiMaxAge, "/", "localhost", false, true)
	ctx.JSON(http.StatusOK, gin.H{"access": access})
}
