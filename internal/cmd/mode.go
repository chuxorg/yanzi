package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/chuxorg/yanzi/internal/config"
)

// RunMode shows or sets the runtime mode.
func RunMode(args []string) error {
	if len(args) == 0 {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		fmt.Printf("Current mode: %s\n", formatMode(cfg))
		return nil
	}

	if len(args) > 1 {
		return modeUsageError()
	}

	switch args[0] {
	case "local":
		return setModeLocal()
	case "http":
		return setModeHTTP()
	default:
		return modeUsageError()
	}
}

func setModeLocal() error {
	dir, err := config.StateDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	dbPath, err := config.DefaultDBPath()
	if err != nil {
		return err
	}
	content := fmt.Sprintf("mode: local\ndb_path: %s\n", dbPath)
	if err := writeConfig(content); err != nil {
		return err
	}
	fmt.Println("Mode set to local.")
	return nil
}

func setModeHTTP() error {
	dir, err := config.StateDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	baseURL := "http://localhost:8080"
	content := fmt.Sprintf("mode: http\nbase_url: %s\n", baseURL)
	if err := writeConfig(content); err != nil {
		return err
	}
	fmt.Printf("Mode set to http (%s).\n", baseURL)
	fmt.Println("Run 'libraryd' to start the server.")
	return nil
}

func writeConfig(content string) error {
	path, err := config.ConfigPath()
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

func modeUsageError() error {
	return errors.New("usage: yanzi mode [local|http]")
}

func formatMode(cfg config.Config) string {
	switch cfg.Mode {
	case config.ModeHTTP:
		return fmt.Sprintf("http (%s)", cfg.BaseURL)
	default:
		return "local"
	}
}
