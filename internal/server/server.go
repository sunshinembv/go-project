package server

import (
	"context"
	"fmt"
	"go-project/internal"
	"go-project/internal/domain"
	"go-project/internal/server/middleware"
	"go-project/internal/server/tasks"
	"go-project/internal/server/users"
	"go-project/internal/service/auth"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Server struct {
	srv *http.Server
}

func New(
	cgf internal.Config,
	userService users.UserService,
	taskService tasks.TaskService,
) *Server {
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

	uh := users.New(userService, signer)
	th := tasks.New(taskService)
	r := configureRouter(signer, uh, th)

	httpSrv.Handler = r

	return &Server{
		srv: &httpSrv,
	}
}

func (s *Server) Run() error {
	return s.srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}

func configureRouter(
	jwtSigner auth.HS256Signer,
	uh *users.UsersHandler,
	th *tasks.TasksHandler,
) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	users := router.Group("/users")

	users.POST("/", uh.Register)
	users.POST("/login", uh.Login)
	users.GET("/profile", middleware.AuthMiddleware(jwtSigner), uh.Profile)

	users.GET("/", middleware.AuthMiddleware(jwtSigner), uh.GetUsers)
	users.GET("/:id", middleware.AuthMiddleware(jwtSigner), uh.GetUserByID)
	users.PUT("/:id", middleware.AuthMiddleware(jwtSigner), uh.UpdateUserByID)
	users.DELETE("/:id", middleware.AuthMiddleware(jwtSigner), uh.DeleteUserByID)

	tasks := router.Group("/tasks")

	tasks.GET("/", middleware.AuthMiddleware(jwtSigner), th.GetTasks)
	tasks.GET("/:id", middleware.AuthMiddleware(jwtSigner), th.GetTaskByTID)
	tasks.POST("/", middleware.AuthMiddleware(jwtSigner), th.CreateTask)
	tasks.PUT("/:id", middleware.AuthMiddleware(jwtSigner), th.UpdateTaskByTID)
	tasks.DELETE("/:id", middleware.AuthMiddleware(jwtSigner), th.DeleteTaskByTID)

	router.POST("/refresh", uh.Refresh)

	return router
}
