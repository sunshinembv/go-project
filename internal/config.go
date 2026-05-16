package internal

import (
	"cmp"
	"flag"
	"os"

	"github.com/rs/zerolog"
)

const (
	defHost  = "0.0.0.0"
	defPort  = "8080"
	defDbDSN = "postgres://user:password@localhost:5432/todolist?sslmode=disable"
	defDebug = false
)

type Config struct {
	Host  string
	Port  string
	DbDSN string
	Debug bool
}

func ReadConfig() *Config {
	var cfg Config

	flag.StringVar(&cfg.Host, "host", defHost, "указание адреса для запуска сервера")
	flag.StringVar(&cfg.Port, "port", defPort, "указание порта для запуска сервера")
	flag.StringVar(&cfg.DbDSN, "dbDSN", defDbDSN, "указание адреса сервера DB")
	flag.BoolVar(&cfg.Debug, "debug", defDebug, "указание уровня логгирования")
	flag.Parse()

	cfg.Host = cmp.Or(os.Getenv("TODO_LIST_SERVICE_HOST"), cfg.Host)
	cfg.Port = cmp.Or(os.Getenv("TODO_LIST_SERVICE_PORT"), cfg.Port)
	cfg.DbDSN = cmp.Or(os.Getenv("TODO_LIST_SERVICE_DB_DSN"), cfg.DbDSN)

	return &cfg
}

func (cfg *Config) ConfigureLogger() {
	if cfg.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		return
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}
