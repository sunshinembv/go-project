package internal

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/rs/zerolog"
)

const (
	defHost  = "0.0.0.0"
	defPort  = "8080"
	defDbDSN = "postgres://user:password@localhost:5432/todolist?sslmode=disable"
	defDebug = false
)

type Config struct {
	Host  string `json:"host"`
	Port  string `json:"port"`
	DbDSN string `json:"dbDSN"`
	Debug bool   `json:"debug"`
}

type setFlags struct {
	Host  bool
	Port  bool
	DbDSN bool
	Debug bool
}

type cliOptions struct {
	ConfigPath string
	Values     Config
	Set        setFlags
}

func ReadConfig() (*Config, error) {
	cfg := Config{
		Host:  defHost,
		Port:  defPort,
		DbDSN: defDbDSN,
		Debug: defDebug,
	}

	cli, err := parseFlags(os.Args[1:])
	if err != nil {
		return nil, err
	}

	configPath := cli.ConfigPath
	if configPath == "" {
		configPath = os.Getenv("CONFIG")
	}

	if configPath != "" {
		if err := readJSONConfig(configPath, &cfg); err != nil {
			return nil, err
		}
	}

	applyExplicitFlags(&cfg, cli)

	if err := applyEnvironment(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (cfg *Config) ConfigureLogger() {
	if cfg.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		return
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

func parseFlags(args []string) (cliOptions, error) {
	var result cliOptions

	fs := flag.NewFlagSet("todo-list", flag.ContinueOnError)

	fs.StringVar(&result.ConfigPath, "c", "", "путь к JSON-конфигурации")
	fs.StringVar(&result.ConfigPath, "config", "", "путь к JSON-конфигурации")

	fs.StringVar(&result.Values.Host, "host", defHost, "указание адреса для запуска сервера")
	fs.StringVar(&result.Values.Port, "port", defPort, "указание порта для запуска сервера")
	fs.StringVar(&result.Values.DbDSN, "dbDSN", defDbDSN, "указание адреса сервера DB")
	fs.BoolVar(&result.Values.Debug, "debug", defDebug, "указание уровня логгирования")

	if err := fs.Parse(args); err != nil {
		return cliOptions{}, err
	}

	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "host":
			result.Set.Host = true
		case "port":
			result.Set.Port = true
		case "dbDSN":
			result.Set.DbDSN = true
		case "debug":
			result.Set.Debug = true
		}
	})

	return result, nil
}

func readJSONConfig(path string, cfg *Config) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open config %q: %w", path, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("Failed to close resource: %v\n", err)
		}
	}()

	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(cfg); err != nil {
		return fmt.Errorf("decode config %q: %w", path, err)
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return fmt.Errorf("config %q contains multiple JSON values", path)
		}
		return fmt.Errorf("decode config %q: %w", path, err)
	}

	return nil
}

func applyExplicitFlags(cfg *Config, cli cliOptions) {
	if cli.Set.Host {
		cfg.Host = cli.Values.Host
	}
	if cli.Set.Port {
		cfg.Port = cli.Values.Port
	}
	if cli.Set.DbDSN {
		cfg.DbDSN = cli.Values.DbDSN
	}
	if cli.Set.Debug {
		cfg.Debug = cli.Values.Debug
	}
}

func applyEnvironment(cfg *Config) error {
	if value := os.Getenv("TODO_LIST_SERVICE_HOST"); value != "" {
		cfg.Host = value
	}
	if value := os.Getenv("TODO_LIST_SERVICE_PORT"); value != "" {
		cfg.Port = value
	}
	if value := os.Getenv("TODO_LIST_SERVICE_DB_DSN"); value != "" {
		cfg.DbDSN = value
	}
	if value := os.Getenv("TODO_LIST_SERVICE_DEBUG"); value != "" {
		debug, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf(
				"invalid TODO_LIST_SERVICE_DEBUG value %q: %w",
				value,
				err,
			)
		}
		cfg.Debug = debug
	}

	return nil
}
