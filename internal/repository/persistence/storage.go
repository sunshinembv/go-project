package persistence

import (
	"encoding/json"
	tasksDomain "go-project/internal/domain/task/models"
	usersDomain "go-project/internal/domain/user/models"
	inmemory "go-project/internal/repository/in_memory"
	"os"
)

type Dump struct {
	Users []usersDomain.User `json:"users"`
	Tasks []tasksDomain.Task `json:"tasks"`
}

type PersistentStorage struct {
	Mem  *inmemory.Storage
	Path string
}

func LoadFromFile(path string) (Dump, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Dump{}, nil
		}
		return Dump{}, err
	}

	var dump Dump
	err = json.Unmarshal(data, &dump)
	return dump, err
}

func SaveToFile(path string, dump Dump) error {
	data, err := json.MarshalIndent(dump, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (s *PersistentStorage) persist() error {
	users, err := s.Mem.GetUsers()
	if err != nil {
		return err
	}

	tasks, err := s.Mem.GetAllTasks()
	if err != nil {
		return err
	}
	dump := Dump{
		Users: users,
		Tasks: tasks,
	}
	return SaveToFile(s.Path, dump)
}
