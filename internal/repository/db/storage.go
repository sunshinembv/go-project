package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type Storage struct {
	conn *pgxpool.Pool
}

func New(ctx context.Context, dbDSN string) (*Storage, error) {
	conn, err := pgxpool.New(ctx, dbDSN)
	if err != nil {
		return nil, err
	}

	const maxAttempts = 10

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err = conn.Ping(ctx)
		if err == nil {
			return &Storage{
				conn: conn,
			}, nil
		}

		sleep := time.Duration(1<<uint(attempt-1)) * time.Second
		if sleep > 30*time.Second {
			sleep = 30 * time.Second
		}

		log.Debug().Str(
			"todo_list_db",
			fmt.Sprintf(
				"failed to connect to db, attempt %d/%d: %v",
				attempt,
				maxAttempts,
				err,
			),
		).Send()

		select {
		case <-time.After(sleep):
		case <-ctx.Done():
			conn.Close()
			return nil, ctx.Err()
		}
	}

	conn.Close()

	return nil, fmt.Errorf("failed to connect to db after %d attempts: %w", maxAttempts, err)
}

func (s *Storage) Close() {
	s.conn.Close()
}

func RunMigrations(dbDSN string) error {
	m, err := migrate.New(
		"file://migrations",
		dbDSN,
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no new migrations")
			return nil
		}

		return err
	}

	fmt.Println("migrations applied successfully")
	return nil
}
