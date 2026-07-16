package internal

import (
	"flag"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestReadConfig(t *testing.T) {
	type want struct {
		cfg Config
	}

	type test struct {
		name     string
		flags    []string
		envSetup func(t *testing.T)
		want     want
	}

	tests := []test{
		{
			name: "config with flags",
			flags: []string{
				"-host", "localhost",
				"-port", "9090",
				"-dbDSN", "mockDBDSN",
				"-debug=true",
			},
			want: want{
				cfg: Config{
					Host:  "localhost",
					Port:  "9090",
					DbDSN: "mockDBDSN",
					Debug: true,
				},
			},
		},
		{
			name: "config with envs",
			envSetup: func(t *testing.T) {
				t.Setenv("TODO_LIST_SERVICE_HOST", "123.123.123.123")
				t.Setenv("TODO_LIST_SERVICE_PORT", "8888")
				t.Setenv("TODO_LIST_SERVICE_DB_DSN", "mockDBDSN")
			},
			want: want{
				cfg: Config{
					Host:  "123.123.123.123",
					Port:  "8888",
					DbDSN: "mockDBDSN",
					Debug: false,
				},
			},
		},
		{
			name: "default values",
			want: want{
				cfg: Config{
					Host:  "0.0.0.0",
					Port:  "8080",
					DbDSN: "postgres://user:password@localhost:5432/todolist?sslmode=disable",
					Debug: false,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			oldArgs := os.Args
			oldCommandLine := flag.CommandLine

			t.Cleanup(func() {
				os.Args = oldArgs
				flag.CommandLine = oldCommandLine
			})

			t.Setenv("TODO_LIST_SERVICE_HOST", "")
			t.Setenv("TODO_LIST_SERVICE_PORT", "")
			t.Setenv("TODO_LIST_SERVICE_DB_DSN", "")

			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			os.Args = []string{"test"}

			if len(tc.flags) > 0 {
				os.Args = append(os.Args, tc.flags...)
			}

			if tc.envSetup != nil {
				tc.envSetup(t)
			}

			testCfg := ReadConfig()

			assert.Equal(t, tc.want.cfg, *testCfg)
		})
	}
}

func TestConfigureLogger(t *testing.T) {
	oldLevel := zerolog.GlobalLevel()
	t.Cleanup(func() {
		zerolog.SetGlobalLevel(oldLevel)
	})

	type want struct {
		level zerolog.Level
	}

	type test struct {
		name  string
		debug bool
		want  want
	}

	tests := []test{
		{
			name:  "debug",
			debug: true,
			want: want{
				level: zerolog.DebugLevel,
			},
		},
		{
			name:  "info",
			debug: false,
			want: want{
				level: zerolog.InfoLevel,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := Config{
				Debug: tc.debug,
			}
			cfg.ConfigureLogger()

			assert.Equal(t, tc.want.level, zerolog.GlobalLevel())
		})
	}
}
