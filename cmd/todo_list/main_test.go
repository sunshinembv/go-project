package main

import (
	"context"
	"os"
	"testing"

	taskModels "go-project/internal/domain/task/models"
	userModels "go-project/internal/domain/user/models"
	"go-project/internal/repository/persistence"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func inTempDirectory(t *testing.T) {
	t.Helper()
	oldDirectory, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(t.TempDir()))
	t.Cleanup(func() { require.NoError(t, os.Chdir(oldDirectory)) })
}

func TestNewPersistent(t *testing.T) {
	inTempDirectory(t)
	dump := persistence.Dump{
		Users: []userModels.User{
			{
				Name:     "Name",
				Email:    "test@example.com",
				Password: "hash",
			},
		},
		Tasks: []taskModels.Task{
			{
				UID:    "legacy-user",
				Title:  "title",
				Status: taskModels.NewStatus,
			},
		},
	}
	require.NoError(t, persistence.SaveToFile("dump.json", dump))

	repo := NewPersistent()
	t.Cleanup(func() { require.NoError(t, repo.Close()) })

	users, err := repo.GetUsers()
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, dump.Users[0].Email, users[0].Email)

	tasks, err := repo.GetTasks("legacy-user")
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, dump.Tasks[0].Title, tasks[0].Title)
}

func TestNewRepoFallback(t *testing.T) {
	inTempDirectory(t)

	repo := NewRepo(context.Background(), "://invalid")
	require.NotNil(t, repo)
	require.NoError(t, repo.Close())

	_, err := os.Stat("dump.json")
	require.NoError(t, err)
}
