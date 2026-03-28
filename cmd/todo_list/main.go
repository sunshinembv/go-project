package main

import (
	"fmt"
	"go-project/internal"
	inmemory "go-project/internal/repository/in_memory"
	"go-project/internal/server"
	taskservice "go-project/internal/service/tasks"
	userservice "go-project/internal/service/users"
)

func main() {
	cfg := internal.ReadConfig()

	repo := inmemory.NewInMemoryStorage()

	userService := userservice.New(repo)
	taskService := taskservice.New(repo)

	srv := server.New(
		cfg,
		userService,
		taskService,
	)

	if err := srv.Run(); err != nil {
		fmt.Println("Failed to start server")
	}
}
