package cmd

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/chuxorg/yanzi/internal/config"
	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
)

// RunRehydrate renders the latest checkpoint and artifacts since.
func RunRehydrate(args []string) error {
	if len(args) != 0 {
		return errors.New("usage: yanzi rehydrate")
	}

	project, err := loadActiveProject()
	if err != nil {
		return err
	}
	if project == "" {
		return errors.New("no active project set")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	switch cfg.Mode {
	case config.ModeLocal:
		if _, ok := os.LookupEnv("YANZI_DB_PATH"); !ok {
			if err := os.Setenv("YANZI_DB_PATH", cfg.DBPath); err != nil {
				return fmt.Errorf("set YANZI_DB_PATH: %w", err)
			}
		}
	case config.ModeHTTP:
		return errors.New("rehydrate is not available in http mode")
	default:
		return fmt.Errorf("invalid mode: %s", cfg.Mode)
	}

	payload, err := yanzilibrary.RehydrateProject(project)
	if err != nil {
		if errors.Is(err, yanzilibrary.ErrCheckpointNotFound) {
			return errors.New("no checkpoint found for active project")
		}
		return err
	}

	intents := payload.IntentsSince
	sort.SliceStable(intents, func(i, j int) bool {
		if intents[i].CreatedAt.Equal(intents[j].CreatedAt) {
			return intents[i].ID < intents[j].ID
		}
		return intents[i].CreatedAt.Before(intents[j].CreatedAt)
	})

	fmt.Printf("Project: %s\n", payload.Project)
	fmt.Println("Latest Checkpoint:")
	fmt.Printf("* CreatedAt: %s\n", payload.LatestCheckpoint.CreatedAt)
	fmt.Printf("* Summary: %s\n", payload.LatestCheckpoint.Summary)
	fmt.Println("Artifacts Since Checkpoint:")
	if len(intents) == 0 {
		fmt.Println("  (none)")
		return nil
	}
	for i, intent := range intents {
		fmt.Printf("%d. %s %s %s\n", i+1, intent.ID, intent.CreatedAt.Format(time.RFC3339Nano), "intent")
	}
	return nil
}
