package main

import (
	"fmt"
	"go-project/internal"
	inmemory "go-project/internal/repository/in_memory"
	"go-project/internal/server"
)

func main() {
	cfg := internal.ReadConfig()

	storage := inmemory.NewInMemoryStorage()

	srv := server.NewServer(cfg, storage)

	if err := srv.Run(); err != nil {
		fmt.Println("Failed to start server")
	}
}
