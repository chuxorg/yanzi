package yanzilibrary

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	ArtifactClassIntent  = "intent"
	ArtifactClassContext = "context"
)

var (
	intentArtifactTypes = map[string]struct{}{
		"prompt":         {},
		"decision":       {},
		"task":           {},
		"change_request": {},
		"checkpoint":     {},
		"note":           {},
	}
	contextArtifactTypes = map[string]struct{}{
		"architecture": {},
		"standard":     {},
		"sop":          {},
		"requirement":  {},
		"policy":       {},
		"constraint":   {},
	}
)

// Artifact represents an artifact stored in the intents ledger table.
type Artifact struct {
	ID        string
	Class     string
	Type      string
	Title     string
	Content   string
	Metadata  string
	CreatedAt string
}

func validateArtifactInput(projectID, class, artifactType, title, content string) error {
	if strings.TrimSpace(projectID) == "" {
		return errors.New("project is required")
	}
	classValue := strings.TrimSpace(class)
	if classValue != ArtifactClassIntent && classValue != ArtifactClassContext {
		return fmt.Errorf("invalid artifact class %q: allowed values are intent, context", class)
	}
	if strings.TrimSpace(title) == "" {
		return errors.New("title is required")
	}
	if strings.TrimSpace(content) == "" {
		return errors.New("content is required")
	}

	typeValue := strings.TrimSpace(artifactType)
	var allowed map[string]struct{}
	switch classValue {
	case ArtifactClassIntent:
		allowed = intentArtifactTypes
	case ArtifactClassContext:
		allowed = contextArtifactTypes
	}
	if _, ok := allowed[typeValue]; !ok {
		if classValue == ArtifactClassIntent {
			return fmt.Errorf("invalid intent type %q: allowed values are prompt, decision, task, change_request, checkpoint, note", artifactType)
		}
		return fmt.Errorf("invalid context type %q: allowed values are architecture, standard, sop, requirement, policy, constraint", artifactType)
	}
	return nil
}

func artifactSystemMeta(projectID string) (string, error) {
	payload, err := json.Marshal(map[string]string{"project": strings.TrimSpace(projectID)})
	if err != nil {
		return "", fmt.Errorf("encode artifact project metadata: %w", err)
	}
	return string(payload), nil
}

func newArtifactID() (string, error) {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("generate artifact id: %w", err)
	}
	return hex.EncodeToString(buf[:]), nil
}

func hashArtifact(id, createdAt, class, artifactType, title, content, metadata string) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		id,
		createdAt,
		class,
		artifactType,
		title,
		content,
		metadata,
	}, "\n")))
	return hex.EncodeToString(sum[:])
}

func nowRFC3339Nano() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}
