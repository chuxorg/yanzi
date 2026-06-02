package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Mode controls whether yanzi runs against local storage or HTTP APIs.
type Mode string

const (
	ModeLocal Mode = "local"
	ModeHTTP  Mode = "http"
)

// Config holds CLI configuration values loaded from disk.
type Config struct {
	Mode    Mode          `yaml:"mode"`
	DBPath  string        `yaml:"db_path"`
	BaseURL string        `yaml:"base_url"`
	Storage StorageConfig `yaml:"storage"`
}

// StorageConfig holds provider selection and per-provider configuration.
type StorageConfig struct {
	Provider string       `yaml:"provider"`
	SQLite   SQLiteConfig `yaml:"sqlite"`
	Postgres PostgresConfig `yaml:"postgres"`
}

// SQLiteConfig holds SQLite-specific configuration.
type SQLiteConfig struct {
	Path string `yaml:"path"`
}

// PostgresConfig holds Postgres-specific configuration.
type PostgresConfig struct {
	DSN             string `yaml:"dsn"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
}

// Environment variable names.
const (
	// LocalDBPathEnvVar is the environment variable that overrides local SQLite resolution.
	LocalDBPathEnvVar = "YANZI_DB_PATH"
	// StorageProviderEnvVar overrides storage.provider.
	StorageProviderEnvVar = "YANZI_STORAGE_PROVIDER"
	// PostgresDSNEnvVar overrides storage.postgres.dsn.
	PostgresDSNEnvVar = "YANZI_POSTGRES_DSN"
	// PostgresMaxConnsEnvVar overrides storage.postgres.max_open_conns.
	PostgresMaxConnsEnvVar = "YANZI_POSTGRES_MAX_CONNS"
)

// Load reads ~/.yanzi/config.yaml and returns the effective CLI configuration.
//
// Environment variables take precedence over config file values. Precedence:
//  1. Environment variables
//  2. config.yaml values
//  3. Defaults
func Load() (Config, error) {
	cfg := Config{
		Mode: ModeLocal,
	}
	path, err := ConfigPath()
	if err != nil {
		return cfg, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			applyDefaults(&cfg)
			applyEnvOverrides(&cfg)
			if err := validateConfig(cfg); err != nil {
				return cfg, err
			}
			return cfg, nil
		}
		return cfg, fmt.Errorf("read config: %w", err)
	}

	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&cfg); err != nil {
		return cfg, fmt.Errorf("invalid config: %w", err)
	}
	if err := ensureEOF(dec); err != nil {
		return cfg, fmt.Errorf("invalid config: %w", err)
	}

	if cfg.Mode == "" {
		cfg.Mode = ModeLocal
	}

	applyDefaults(&cfg)
	applyEnvOverrides(&cfg)

	cfg.BaseURL = strings.TrimSpace(cfg.BaseURL)
	cfg.DBPath = strings.TrimSpace(cfg.DBPath)
	cfg.Storage.Provider = strings.TrimSpace(cfg.Storage.Provider)
	cfg.Storage.Postgres.DSN = strings.TrimSpace(cfg.Storage.Postgres.DSN)

	if err := validateConfig(cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.Mode == ModeLocal && cfg.DBPath == "" {
		if path, err := DefaultDBPath(); err == nil {
			cfg.DBPath = path
		}
	}
	if cfg.Storage.Postgres.MaxOpenConns == 0 {
		cfg.Storage.Postgres.MaxOpenConns = 25
	}
	if cfg.Storage.Postgres.MaxIdleConns == 0 {
		cfg.Storage.Postgres.MaxIdleConns = 5
	}
	if cfg.Storage.Postgres.ConnMaxLifetime == 0 {
		cfg.Storage.Postgres.ConnMaxLifetime = 300
	}
}

func applyEnvOverrides(cfg *Config) {
	if v := strings.TrimSpace(os.Getenv(StorageProviderEnvVar)); v != "" {
		cfg.Storage.Provider = v
	}
	if v := strings.TrimSpace(os.Getenv(PostgresDSNEnvVar)); v != "" {
		cfg.Storage.Postgres.DSN = v
	}
	if v := strings.TrimSpace(os.Getenv(PostgresMaxConnsEnvVar)); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.Storage.Postgres.MaxOpenConns = n
		}
	}
}

func validateConfig(cfg Config) error {
	if cfg.Mode != ModeLocal && cfg.Mode != ModeHTTP {
		return fmt.Errorf("invalid mode: %s", cfg.Mode)
	}
	if cfg.Mode == ModeHTTP && cfg.BaseURL == "" {
		return errors.New("base_url is required when mode=http")
	}
	if cfg.Mode == ModeLocal && cfg.DBPath == "" && cfg.Storage.Provider != "postgres" {
		return errors.New("db_path is required when mode=local")
	}
	provider := strings.TrimSpace(cfg.Storage.Provider)
	if provider == "postgres" && strings.TrimSpace(cfg.Storage.Postgres.DSN) == "" {
		return errors.New("postgres provider requires YANZI_POSTGRES_DSN or storage.postgres.dsn in config")
	}
	return nil
}

func ensureEOF(dec *yaml.Decoder) error {
	var extra any
	if err := dec.Decode(&extra); err == nil {
		return errors.New("multiple YAML documents are not supported")
	} else if !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}

// ConfigPath returns the full path to ~/.yanzi/config.yaml.
func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home: %w", err)
	}
	return filepath.Join(home, ".yanzi", "config.yaml"), nil
}

// StateDir returns the ~/.yanzi directory path.
func StateDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home: %w", err)
	}
	return filepath.Join(home, ".yanzi"), nil
}

// DefaultDBPath returns the default SQLite path under ~/.yanzi.
func DefaultDBPath() (string, error) {
	dir, err := StateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "yanzi.db"), nil
}

// EffectiveLocalDBPath resolves the local SQLite path using deterministic precedence.
//
// Precedence:
//  1. YANZI_DB_PATH
//  2. Config.Storage.SQLite.Path
//  3. Config.DBPath
//  4. DefaultDBPath()
func EffectiveLocalDBPath(cfg Config) (string, error) {
	if override := strings.TrimSpace(os.Getenv(LocalDBPathEnvVar)); override != "" {
		return override, nil
	}
	if path := strings.TrimSpace(cfg.Storage.SQLite.Path); path != "" {
		return path, nil
	}
	if path := strings.TrimSpace(cfg.DBPath); path != "" {
		return path, nil
	}
	return DefaultDBPath()
}

// EffectiveStorageProvider returns the active provider name.
// Defaults to "sqlite" if not configured.
func EffectiveStorageProvider(cfg Config) string {
	p := strings.TrimSpace(cfg.Storage.Provider)
	if p == "" {
		return "sqlite"
	}
	return p
}
