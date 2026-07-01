package db

import (
	"context"
	"errors"
	tasksErrors "go-project/internal/domain/task/errors"
	tasksDomain "go-project/internal/domain/task/models"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

const deletedTasksBatchSize uint64 = 10

func (s *Storage) GetTasks(uid string) ([]tasksDomain.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := s.conn.Query(
		ctx,
		"SELECT tid, user_id, title, description, status, deleted FROM tasks WHERE user_id=$1 AND deleted=false",
		uid,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []tasksDomain.Task
	for rows.Next() {
		var task tasksDomain.Task
		if err := rows.Scan(
			&task.TID, &task.UID, &task.Title, &task.Description, &task.Status, &task.Deleted,
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
	var taskDeleted bool
	err := s.conn.QueryRow(
		ctx,
		`UPDATE tasks SET title=$1, description=$2, status=$3
		WHERE user_id=$4 AND tid=$5 AND deleted=false
		RETURNING title, description, status, deleted`,
		req.Title, req.Description, req.Status, uid, tid,
	).Scan(&updatedTitle, &updatedDescription, &updatedStatus, &taskDeleted)

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
		Deleted:     taskDeleted,
	}, nil
}

func (s *Storage) DeleteTaskByTID(uid string, tid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmdTag, err := s.conn.Exec(
		ctx,
		"UPDATE tasks SET deleted=true WHERE user_id=$1 AND tid=$2 AND deleted=false",
		uid,
		tid,
	)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return tasksErrors.ErrTaskNotFound
	}

	if s.deletedTasksCount.Add(1) >= deletedTasksBatchSize {
		s.startDeletedTasksWorker()
	}

	return nil
}

func (s *Storage) GetTaskByTID(uid string, tid string) (tasksDomain.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row := s.conn.QueryRow(
		ctx,
		"SELECT tid, user_id, title, description, status, deleted FROM tasks WHERE user_id=$1 AND tid=$2 AND deleted=false",
		uid,
		tid,
	)
	var task tasksDomain.Task
	if err := row.Scan(
		&task.TID,
		&task.UID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Deleted,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return tasksDomain.Task{}, tasksErrors.ErrTaskNotFound
		}
		return tasksDomain.Task{}, err
	}

	return task, nil
}

func (s *Storage) startDeletedTasksWorker() {
	if !s.deletedTasksCleanupRunning.CompareAndSwap(false, true) {
		return
	}

	s.deletedTasksWorker.Add(1)
	go s.runDeletedTasksWorker()
}

func (s *Storage) runDeletedTasksWorker() {
	defer s.deletedTasksWorker.Done()

	for {
		countBeforeBatch, ok := s.takeDeletedTasksBatch()
		if !ok {
			if s.releaseDeletedTasksWorker() {
				return
			}
			continue
		}

		if err := s.deleteMarkedTasks(); err != nil {
			log.Error().Err(err).Msg("failed to delete marked tasks")
			s.deletedTasksCount.Add(deletedTasksBatchSize)
			if s.releaseDeletedTasksWorkerAfterFailure(countBeforeBatch) {
				return
			}
		}
	}
}

func (s *Storage) takeDeletedTasksBatch() (uint64, bool) {
	for {
		count := s.deletedTasksCount.Load()
		if count < deletedTasksBatchSize {
			return 0, false
		}
		if s.deletedTasksCount.CompareAndSwap(count, count-deletedTasksBatchSize) {
			return count, true
		}
	}
}

func (s *Storage) releaseDeletedTasksWorker() bool {
	s.deletedTasksCleanupRunning.Store(false)
	if s.deletedTasksCount.Load() < deletedTasksBatchSize {
		return true
	}

	return !s.deletedTasksCleanupRunning.CompareAndSwap(false, true)
}

func (s *Storage) releaseDeletedTasksWorkerAfterFailure(countBeforeBatch uint64) bool {
	s.deletedTasksCleanupRunning.Store(false)
	if s.deletedTasksCount.Load() <= countBeforeBatch {
		return true
	}

	return !s.deletedTasksCleanupRunning.CompareAndSwap(false, true)
}

func (s *Storage) deleteMarkedTasks() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if _, err := tx.Exec(ctx, "DELETE FROM tasks WHERE deleted=true"); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
