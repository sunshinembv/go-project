package db

import (
	"context"
	"errors"
	tasksErrors "go-project/internal/domain/task/errors"
	tasksDomain "go-project/internal/domain/task/models"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Storage) GetTasks(uid string) ([]tasksDomain.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := s.conn.Query(ctx, "SELECT tid, user_id, title, description, status FROM tasks WHERE user_id=$1", uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []tasksDomain.Task
	for rows.Next() {
		var task tasksDomain.Task
		if err := rows.Scan(
			&task.TID, &task.UID, &task.Title, &task.Description, &task.Status,
		); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s *Storage) CreateTask(uid string, task tasksDomain.Task) (uuid.UUID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var tid uuid.UUID
	err := s.conn.QueryRow(
		ctx,
		`INSERT INTO tasks (user_id, title, description, status)
		VALUES ($1, $2, $3, $4) RETURNING TID`,
		uid, task.Title, task.Description, task.Status,
	).Scan(&tid)
	if err != nil {
		return uuid.Nil, err
	}

	return tid, nil
}

func (s *Storage) UpdateTaskByTID(uid string, tid string, req tasksDomain.TaskRequest) (tasksDomain.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var updatedTitle string
	var updatedDescription string
	var updatedStatus tasksDomain.Status
	err := s.conn.QueryRow(
		ctx,
		`UPDATE tasks SET title=$1, description=$2, status=$3 WHERE user_id=$4 AND tid=$5 
		RETURNING title, description, status`,
		req.Title, req.Description, req.Status, uid, tid,
	).Scan(&updatedTitle, &updatedDescription, &updatedStatus)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return tasksDomain.Task{}, tasksErrors.ErrTaskNotFound
		}
		return tasksDomain.Task{}, err
	}

	return tasksDomain.Task{
		TID:         tid,
		UID:         uid,
		Title:       updatedTitle,
		Description: updatedDescription,
		Status:      updatedStatus,
	}, nil
}

func (s *Storage) DeleteTaskByTID(uid string, tid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmdTag, err := s.conn.Exec(ctx, "DELETE FROM tasks WHERE user_id=$1 AND tid=$2", uid, tid)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return tasksErrors.ErrTaskNotFound
	}
	return nil
}

func (s *Storage) GetTaskByTID(uid string, tid string) (tasksDomain.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row := s.conn.QueryRow(ctx, "SELECT tid, user_id, title, description, status FROM tasks WHERE user_id=$1 AND tid=$2", uid, tid)
	var task tasksDomain.Task
	if err := row.Scan(&task.TID, &task.UID, &task.Title, &task.Description, &task.Status); err != nil {
		return tasksDomain.Task{}, err
	}

	return task, nil
}
