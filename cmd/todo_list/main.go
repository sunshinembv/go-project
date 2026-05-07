package main

import (
	"fmt"
	"go-project/internal"
	"go-project/internal/repository/db"
	inmemory "go-project/internal/repository/in_memory"
	"go-project/internal/repository/persistence"
	"go-project/internal/server"
	taskservice "go-project/internal/service/tasks"
	userservice "go-project/internal/service/users"
	storageinterfaces "go-project/internal/storage_interfaces"
	"os"
)

func main() {
	cfg := internal.ReadConfig()
	dbDSN := cfg.DbDSN

	repo := NewRepo(dbDSN)

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

func NewRepo(dbDSN string) storageinterfaces.Repositories {

	repo, err := db.New(dbDSN)
	if err != nil {
		fmt.Printf("DB connection error: %v", err)
		return NewPersistent()
	}
	if err := db.RunMigrations(dbDSN); err != nil {
		fmt.Printf("DB migrations error: %v", err)
		return NewPersistent()
	}

	dump, _ := persistence.LoadFromFile("dump.json")

	for _, user := range dump.Users {
		_, _ = repo.CreateUser(user)
	}

	for _, task := range dump.Tasks {
		_, _ = repo.CreateTask(task.UID, task)
	}

	_ = os.Remove("dump.json")

	return repo
}

func NewPersistent() storageinterfaces.Repositories {
	repo := inmemory.New()
	dump, _ := persistence.LoadFromFile("dump.json")

	for _, user := range dump.Users {
		repo.CreateUser(user)
	}
	for _, task := range dump.Tasks {
		repo.CreateTask(task.UID, task)
	}

	persistent := &persistence.PersistentStorage{
		Mem:  repo,
		Path: "dump.json",
	}

	return persistent
}
