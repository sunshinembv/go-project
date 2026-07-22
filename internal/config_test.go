package internal

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	envConfig = "CONFIG"
	envHost   = "TODO_LIST_SERVICE_HOST"
	envPort   = "TODO_LIST_SERVICE_PORT"
	envDBDSN  = "TODO_LIST_SERVICE_DB_DSN"
	envDebug  = "TODO_LIST_SERVICE_DEBUG"
)

func prepareConfigTest(t *testing.T, flags []string) {
	t.Helper()

	oldArgs := os.Args
	oldCommandLine := flag.CommandLine
	t.Cleanup(func() {
		os.Args = oldArgs
		flag.CommandLine = oldCommandLine
	})

	for _, env := range []string{envConfig, envHost, envPort, envDBDSN, envDebug} {
		t.Setenv(env, "")
	}

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = append([]string{"test"}, flags...)
}

func writeConfigFile(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.json")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	return path
}

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

			testCfg, err := ReadConfig()
			require.NoError(t, err)

			assert.Equal(t, tc.want.cfg, *testCfg)
		})
	}
}

func TestReadConfigFromJSON(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		content string
		want    Config
	}{
		{
			name:   "short flag",
			source: "c",
			content: `{
				"host": "json-host",
				"port": "7070",
				"dbDSN": "jsonDBDSN",
				"debug": true
			}`,
			want: Config{
				Host:  "json-host",
				Port:  "7070",
				DbDSN: "jsonDBDSN",
				Debug: true,
			},
		},
		{
			name:    "long flag with partial config",
			source:  "config",
			content: `{"port":"6060"}`,
			want: Config{
				Host:  defHost,
				Port:  "6060",
				DbDSN: defDbDSN,
				Debug: defDebug,
			},
		},
		{
			name:    "CONFIG environment variable",
			source:  "env",
			content: `{"host":"config-env-host","debug":true}`,
			want: Config{
				Host:  "config-env-host",
				Port:  defPort,
				DbDSN: defDbDSN,
				Debug: true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			prepareConfigTest(t, nil)
			configPath := writeConfigFile(t, tc.content)

			switch tc.source {
			case "c", "config":
				os.Args = append(os.Args, "-"+tc.source, configPath)
			case "env":
				t.Setenv(envConfig, configPath)
			default:
				t.Fatalf("unknown config source %q", tc.source)
			}

			cfg, err := ReadConfig()

			require.NoError(t, err)
			assert.Equal(t, tc.want, *cfg)
		})
	}
}

func TestReadConfigFileHasLowerPriority(t *testing.T) {
	prepareConfigTest(t, []string{
		"-host", "flag-host",
		"-debug=false",
	})

	configPath := writeConfigFile(t, `{
		"host": "json-host",
		"port": "5050",
		"dbDSN": "jsonDBDSN",
		"debug": true
	}`)
	os.Args = append(os.Args, "-c", configPath)
	t.Setenv(envPort, "4040")
	t.Setenv(envDBDSN, "envDBDSN")

	cfg, err := ReadConfig()

	require.NoError(t, err)
	assert.Equal(t, Config{
		Host:  "flag-host",
		Port:  "4040",
		DbDSN: "envDBDSN",
		Debug: false,
	}, *cfg)
}

func TestReadConfigErrors(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		writeFile bool
		wantError string
	}{
		{
			name:      "config file does not exist",
			wantError: "open config",
		},
		{
			name:      "malformed JSON",
			content:   `{"host":`,
			writeFile: true,
			wantError: "decode config",
		},
		{
			name:      "unknown JSON field",
			content:   `{"unknown":"value"}`,
			writeFile: true,
			wantError: "unknown field",
		},
		{
			name:      "multiple JSON values",
			content:   `{"host":"first"} {}`,
			writeFile: true,
			wantError: "multiple JSON values",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			configPath := filepath.Join(t.TempDir(), "config.json")
			if tc.writeFile {
				require.NoError(t, os.WriteFile(configPath, []byte(tc.content), 0o600))
			}
			prepareConfigTest(t, []string{"-c", configPath})

			cfg, err := ReadConfig()

			require.Error(t, err)
			assert.Nil(t, cfg)
			assert.ErrorContains(t, err, tc.wantError)
		})
	}
}

func TestReadConfigInvalidDebugEnvironment(t *testing.T) {
	prepareConfigTest(t, nil)
	t.Setenv(envDebug, "not-a-bool")

	cfg, err := ReadConfig()

	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.ErrorContains(t, err, envDebug)
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
