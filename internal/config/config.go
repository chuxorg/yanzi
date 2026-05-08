package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	Mode    Mode   `yaml:"mode"`
	DBPath  string `yaml:"db_path"`
	BaseURL string `yaml:"base_url"`
}

// Load reads ~/.yanzi/config.yaml and returns the effective CLI configuration.
//
// Problem:
// The CLI needs one deterministic source of truth for local and optional HTTP
// runtime settings.
//
// Solution:
// Load reads the config file, applies defaults, trims values, and validates
// mode-specific requirements before returning.
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
	cfg.BaseURL = strings.TrimSpace(cfg.BaseURL)
	cfg.DBPath = strings.TrimSpace(cfg.DBPath)

	if cfg.Mode != ModeLocal && cfg.Mode != ModeHTTP {
		return cfg, fmt.Errorf("invalid mode: %s", cfg.Mode)
	}
	if cfg.Mode == ModeHTTP && cfg.BaseURL == "" {
		return cfg, errors.New("base_url is required when mode=http")
	}
	if cfg.Mode == ModeLocal && cfg.DBPath == "" {
		return cfg, errors.New("db_path is required when mode=local")
	}

	return cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.Mode == ModeLocal && cfg.DBPath == "" {
		if path, err := DefaultDBPath(); err == nil {
			cfg.DBPath = path
		}
	}
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
