package persistence

import (
	tasksDomain "go-project/internal/domain/task/models"
	usersDomain "go-project/internal/domain/user/models"

	"github.com/google/uuid"
)

func (s *PersistentStorage) GetUsers() ([]usersDomain.User, error) {
	users, err := s.Mem.GetUsers()
	if err != nil {
		return []usersDomain.User{}, err
	}

	_ = s.persist()
	return users, nil
}

func (s *PersistentStorage) GetUserByUID(uid string) (usersDomain.User, error) {
	user, err := s.Mem.GetUserByUID(uid)
	if err != nil {
		return usersDomain.User{}, err
	}

	_ = s.persist()
	return user, nil
}

func (s *PersistentStorage) CreateUser(user usersDomain.User) (uuid.UUID, error) {
	uid, err := s.Mem.CreateUser(user)
	if err != nil {
		return uuid.Nil, err
	}

	_ = s.persist()
	return uid, nil
}

func (s *PersistentStorage) UpdateUserByUID(uid string, userReq usersDomain.UserUpdateRequest) (string, string, error) {
	name, email, err := s.Mem.UpdateUserByUID(uid, userReq)
	if err != nil {
		return "", "", err
	}

	_ = s.persist()
	return name, email, nil
}

func (s *PersistentStorage) DeleteUserByUID(uid string) error {
	err := s.Mem.DeleteUserByUID(uid)
	if err != nil {
		return err
	}

	_ = s.persist()
	return nil
}

func (s *PersistentStorage) GetUserByEmail(email string) (usersDomain.User, error) {
	user, err := s.Mem.GetUserByEmail(email)
	if err != nil {
		return usersDomain.User{}, err
	}

	_ = s.persist()
	return user, nil
}

func (s *PersistentStorage) GetTasks(uid string) ([]tasksDomain.Task, error) {
	tasks, err := s.Mem.GetTasks(uid)
	if err != nil {
		return []tasksDomain.Task{}, err
	}

	_ = s.persist()
	return tasks, nil
}

func (s *PersistentStorage) GetTaskByTID(uid string, tid string) (tasksDomain.Task, error) {
	task, err := s.Mem.GetTaskByTID(uid, tid)
	if err != nil {
		return tasksDomain.Task{}, err
	}

	_ = s.persist()
	return task, nil
}

func (s *PersistentStorage) CreateTask(uid string, task tasksDomain.Task) (uuid.UUID, error) {
	tid, err := s.Mem.CreateTask(uid, task)
	if err != nil {
		return uuid.Nil, err
	}

	_ = s.persist()
	return tid, nil
}

func (s *PersistentStorage) UpdateTaskByTID(uid string, tid string, req tasksDomain.TaskRequest) (tasksDomain.Task, error) {
	task, err := s.Mem.UpdateTaskByTID(uid, tid, req)
	if err != nil {
		return tasksDomain.Task{}, err
	}

	_ = s.persist()
	return task, nil
}

func (s *PersistentStorage) DeleteTaskByTID(uid string, tid string) error {
	err := s.Mem.DeleteTaskByTID(uid, tid)
	if err != nil {
		return err
	}

	_ = s.persist()
	return nil
}
