package main

import (
	"context"
	"errors"
	"fmt"
	"go-project/internal"
	"go-project/internal/repository/db"
	inmemory "go-project/internal/repository/in_memory"
	"go-project/internal/repository/persistence"
	"go-project/internal/server"
	taskservice "go-project/internal/service/tasks"
	userservice "go-project/internal/service/users"
	storageinterfaces "go-project/internal/storage_interfaces"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cfg := internal.ReadConfig()
	cfg.ConfigureLogger()

	dbDSN := cfg.DbDSN
	repo := NewRepo(ctx, dbDSN)
	defer func() {
		if err := repo.Close(); err != nil {
			fmt.Printf("Failed to close repository: %v\n", err)
		}
	}()

	userService := userservice.New(repo)
	taskService := taskservice.New(repo)

	srv := server.New(
		*cfg,
		userService,
		taskService,
	)

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- srv.Run()
	}()

	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("Server error: %v\n", err)
		}
		return
	case <-shutdownCtx.Done():
		fmt.Println("Shutdown signal received")
	}

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		fmt.Printf("Failed to shutdown server gracefully: %v\n", err)
		return
	}

	if err := <-serverErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("Server error: %v\n", err)
		return
	}

	fmt.Println("Server stopped gracefully")
}

func NewRepo(ctx context.Context, dbDSN string) storageinterfaces.Repositories {

	repo, err := db.New(ctx, dbDSN)
	if err != nil {
		fmt.Printf("DB connection error: %v", err)
		return NewPersistent()
	}
	if err := db.RunMigrations(dbDSN); err != nil {
		fmt.Printf("DB migrations error: %v", err)
		if closeErr := repo.Close(); closeErr != nil {
			fmt.Printf("DB close error: %v", closeErr)
		}
		return NewPersistent()
	}

	dump, _ := persistence.LoadFromFile("dump.json")

	for _, user := range dump.Users {
		_, _ = repo.CreateUser(user)
	}

	for _, task := range dump.Tasks {
		if task.Deleted {
			continue
		}
		_, _ = repo.CreateTask(task.UID, task)
	}

	_ = os.Remove("dump.json")

	return repo
}

func NewPersistent() storageinterfaces.Repositories {
	repo := inmemory.New()
	dump, _ := persistence.LoadFromFile("dump.json")

	for _, user := range dump.Users {
		if _, err := repo.CreateUser(user); err != nil {
			fmt.Printf("Failed to restore user: %v", err)
		}
	}
	for _, task := range dump.Tasks {
		if _, err := repo.CreateTask(task.UID, task); err != nil {
			fmt.Printf("Failed to restore task: %v", err)
		}
	}

	persistent := &persistence.PersistentStorage{
		Mem:  repo,
		Path: "dump.json",
	}

	return persistent
}
