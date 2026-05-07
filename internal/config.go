package internal

import "flag"

const (
	defHost  = "127.0.0.1"
	defPort  = 8080
	defDbDSN = "postgres://user:password@localhost:5432/todolist?sslmode=disable"
)

type Config struct {
	Host  string
	Port  int
	DbDSN string
	Debug bool
}

func ReadConfig() Config {
	var cfg Config

	flag.StringVar(&cfg.Host, "host", defHost, "указание адреса для запуска сервера")
	flag.IntVar(&cfg.Port, "port", defPort, "указание порта для запуска сервера")
	flag.StringVar(&cfg.DbDSN, "dbDSN", defDbDSN, "указание адреса сервера DB")
	flag.BoolVar(&cfg.Debug, "debug", false, "указание уровня логгирования")
	flag.Parse()

	return cfg
}
