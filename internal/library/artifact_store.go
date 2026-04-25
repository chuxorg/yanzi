package yanzilibrary

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

// CreateArtifact stores a new artifact for a project.
func CreateArtifact(projectID, class, artifactType, title, content, metadata string) (Artifact, error) {
	if err := validateArtifactInput(projectID, class, artifactType, title, content); err != nil {
		return Artifact{}, err
	}

	db, err := InitDB()
	if err != nil {
		return Artifact{}, err
	}
	defer func() {
		_ = db.Close()
	}()

	return createArtifact(db, projectID, class, artifactType, title, content, metadata)
}

func createArtifact(db *sql.DB, projectID, class, artifactType, title, content, metadata string) (Artifact, error) {
	if db == nil {
		return Artifact{}, fmt.Errorf("artifact store is not initialized")
	}
	exists, err := projectExists(context.Background(), db, projectID)
	if err != nil {
		return Artifact{}, err
	}
	if !exists {
		return Artifact{}, ProjectNotFoundError{Name: projectID}
	}

	id, err := newArtifactID()
	if err != nil {
		return Artifact{}, err
	}
	createdAt := nowRFC3339Nano()
	systemMeta, err := artifactSystemMeta(projectID)
	if err != nil {
		return Artifact{}, err
	}
	hashValue := hashArtifact(id, createdAt, class, artifactType, title, content, metadata)

	var metadataValue any
	if metadata != "" {
		metadataValue = metadata
	}

	if _, err := db.Exec(
		`INSERT INTO intents (
			id, created_at, author, source_type, title, prompt, response, meta, prev_hash, hash, class, type, content, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id,
		createdAt,
		"yanzi",
		"artifact",
		title,
		content,
		content,
		systemMeta,
		nil,
		hashValue,
		class,
		artifactType,
		content,
		metadataValue,
	); err != nil {
		return Artifact{}, err
	}

	return Artifact{
		ID:        id,
		Class:     class,
		Type:      artifactType,
		Title:     title,
		Content:   content,
		Metadata:  metadata,
		CreatedAt: createdAt,
	}, nil
}

// ListArtifacts lists artifacts for a project and class, optionally filtered by type.
func ListArtifacts(projectID, class, artifactType string) ([]Artifact, error) {
	if err := validateArtifactClassAndType(class, artifactType); err != nil {
		return nil, err
	}
	if projectID == "" {
		return nil, fmt.Errorf("project is required")
	}

	db, err := InitDB()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = db.Close()
	}()

	return listArtifacts(db, projectID, class, artifactType)
}

func validateArtifactClassAndType(class, artifactType string) error {
	class = strings.TrimSpace(class)
	switch class {
	case ArtifactClassIntent:
		if artifactType == "" {
			return nil
		}
		if _, ok := intentArtifactTypes[strings.TrimSpace(artifactType)]; ok {
			return nil
		}
		return fmt.Errorf("invalid intent type %q: allowed values are prompt, decision, task, change_request, checkpoint, note", artifactType)
	case ArtifactClassContext:
		if artifactType == "" {
			return nil
		}
		if _, ok := contextArtifactTypes[strings.TrimSpace(artifactType)]; ok {
			return nil
		}
		return fmt.Errorf("invalid context type %q: allowed values are architecture, standard, sop, requirement, policy, constraint", artifactType)
	default:
		return fmt.Errorf("invalid artifact class %q: allowed values are intent, context", class)
	}
}

func listArtifacts(db *sql.DB, projectID, class, artifactType string) ([]Artifact, error) {
	rows, err := db.Query(
		`SELECT id, class, type, title, content, metadata, meta, created_at
		FROM intents
		WHERE source_type = 'artifact' AND class = ?
		ORDER BY created_at DESC, id DESC`,
		class,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	artifacts := make([]Artifact, 0)
	for rows.Next() {
		var artifact Artifact
		var metadata sql.NullString
		var meta sql.NullString
		var title sql.NullString
		var content sql.NullString
		if err := rows.Scan(
			&artifact.ID,
			&artifact.Class,
			&artifact.Type,
			&title,
			&content,
			&metadata,
			&meta,
			&artifact.CreatedAt,
		); err != nil {
			return nil, err
		}
		project, err := projectFromMeta(meta.String)
		if err != nil || project != projectID {
			continue
		}
		if artifactType != "" && artifact.Type != artifactType {
			continue
		}
		if title.Valid {
			artifact.Title = title.String
		}
		if content.Valid {
			artifact.Content = content.String
		}
		if metadata.Valid {
			artifact.Metadata = metadata.String
		}
		artifacts = append(artifacts, artifact)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return artifacts, nil
}

func projectFromMeta(metaText string) (string, error) {
	if metaText == "" {
		return "", nil
	}
	var payload map[string]string
	if err := json.Unmarshal([]byte(metaText), &payload); err != nil {
		return "", err
	}
	return strings.TrimSpace(payload["project"]), nil
}
