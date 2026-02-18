package internal

import "flag"

const (
	defHost = "127.0.0.1"
	defPort = 8080
)

type Config struct {
	Host  string
	Port  int
	Debug bool
}

func ReadConfig() Config {
	var cfg Config

	flag.StringVar(&cfg.Host, "host", defHost, "указание адреса для запуска сервера")
	flag.IntVar(&cfg.Port, "port", defPort, "указание порта для запуска сервера")
	flag.BoolVar(&cfg.Debug, "debug", false, "указание уровня логгирования")
	flag.Parse()

	return cfg
}
