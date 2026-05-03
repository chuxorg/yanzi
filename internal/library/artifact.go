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
	ContextScopeGlobal   = "global"
	ContextScopeProject  = "project"
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
		"requirement":     {},
		"process_rule":    {},
		"coding_standard": {},
		"reference":       {},
		"note":            {},
	}
)

// Artifact represents an artifact stored in the intents ledger table.
type Artifact struct {
	ID        string
	Class     string
	Type      string
	Scope     string
	Project   string
	Title     string
	Content   string
	Metadata  string
	CreatedAt string
}

func validateArtifactInput(projectID, class, artifactType, title, content, scope string) error {
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
		if strings.TrimSpace(projectID) == "" {
			return errors.New("project is required")
		}
		allowed = intentArtifactTypes
	case ArtifactClassContext:
		if err := validateContextScope(scope); err != nil {
			return err
		}
		if strings.TrimSpace(scope) == ContextScopeProject && strings.TrimSpace(projectID) == "" {
			return errors.New("project is required for project-scoped context")
		}
		allowed = contextArtifactTypes
	}
	if _, ok := allowed[typeValue]; !ok {
		if classValue == ArtifactClassIntent {
			return fmt.Errorf("invalid intent type %q: allowed values are prompt, decision, task, change_request, checkpoint, note", artifactType)
		}
		return fmt.Errorf("invalid context type %q: allowed values are requirement, process_rule, coding_standard, reference, note", artifactType)
	}
	return nil
}

func validateContextScope(scope string) error {
	switch strings.TrimSpace(scope) {
	case ContextScopeGlobal, ContextScopeProject:
		return nil
	default:
		return fmt.Errorf("invalid context scope %q: allowed values are global, project", scope)
	}
}

func artifactSystemMeta(projectID, scope string) (string, error) {
	payload := map[string]string{}
	if project := strings.TrimSpace(projectID); project != "" {
		payload["project"] = project
	}
	if scopeValue := strings.TrimSpace(scope); scopeValue != "" {
		payload["scope"] = scopeValue
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("encode artifact project metadata: %w", err)
	}
	return string(payloadJSON), nil
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

type artifactSystemFields struct {
	Project   string
	Scope     string
	Deleted   bool
	DeletedAt string
}

func artifactSystemFieldsFromMeta(metaText string) (artifactSystemFields, error) {
	if metaText == "" {
		return artifactSystemFields{}, nil
	}

	var payload map[string]string
	if err := json.Unmarshal([]byte(metaText), &payload); err != nil {
		return artifactSystemFields{}, err
	}

	fields := artifactSystemFields{
		Project:   strings.TrimSpace(payload["project"]),
		Scope:     strings.TrimSpace(payload["scope"]),
		Deleted:   strings.EqualFold(strings.TrimSpace(payload["deleted"]), "true"),
		DeletedAt: strings.TrimSpace(payload["deleted_at"]),
	}
	if fields.Scope == "" {
		if fields.Project != "" {
			fields.Scope = ContextScopeProject
		} else {
			fields.Scope = ContextScopeGlobal
		}
	}
	return fields, nil
}
