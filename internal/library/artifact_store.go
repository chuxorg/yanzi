package yanzilibrary

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// CreateArtifact stores a new artifact for a project.
func CreateArtifact(projectID, class, artifactType, title, content, metadata string) (Artifact, error) {
	scope := ""
	if strings.TrimSpace(class) == ArtifactClassContext {
		scope = ContextScopeProject
	}
	if err := validateArtifactInput(projectID, class, artifactType, title, content, scope); err != nil {
		return Artifact{}, err
	}

	db, err := InitDB()
	if err != nil {
		return Artifact{}, err
	}
	defer func() {
		_ = db.Close()
	}()

	return createArtifact(db, projectID, class, artifactType, title, content, metadata, scope)
}

// CreateContextArtifact stores a new context artifact using the Phase 6 scope rules.
func CreateContextArtifact(projectID, artifactType, scope, title, content, metadata string) (Artifact, error) {
	if err := validateArtifactInput(projectID, ArtifactClassContext, artifactType, title, content, scope); err != nil {
		return Artifact{}, err
	}

	db, err := InitDB()
	if err != nil {
		return Artifact{}, err
	}
	defer func() {
		_ = db.Close()
	}()

	return createArtifact(db, projectID, ArtifactClassContext, artifactType, title, content, metadata, scope)
}

func createArtifact(db *sql.DB, projectID, class, artifactType, title, content, metadata, scope string) (Artifact, error) {
	if db == nil {
		return Artifact{}, fmt.Errorf("artifact store is not initialized")
	}
	if strings.TrimSpace(projectID) != "" {
		exists, err := projectExists(context.Background(), db, projectID)
		if err != nil {
			return Artifact{}, err
		}
		if !exists {
			return Artifact{}, ProjectNotFoundError{Name: projectID}
		}
	}

	id, err := newArtifactID()
	if err != nil {
		return Artifact{}, err
	}
	createdAt := nowRFC3339Nano()
	systemMeta, err := artifactSystemMeta(projectID, scope)
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
		Scope:     scope,
		Project:   projectID,
		Title:     title,
		Content:   content,
		Metadata:  metadata,
		CreatedAt: createdAt,
	}, nil
}

// ListArtifacts lists artifacts for a project and class, optionally filtered by type.
func ListArtifacts(projectID, class, artifactType string, includeDeleted bool) ([]Artifact, error) {
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

	return listArtifacts(db, projectID, class, artifactType, includeDeleted)
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
		return fmt.Errorf("invalid context type %q: allowed values are requirement, process_rule, coding_standard, reference, note", artifactType)
	default:
		return fmt.Errorf("invalid artifact class %q: allowed values are intent, context", class)
	}
}

func listArtifacts(db *sql.DB, projectID, class, artifactType string, includeDeleted bool) ([]Artifact, error) {
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
		fields, err := artifactSystemFieldsFromMeta(meta.String)
		if err != nil || fields.Project != projectID {
			continue
		}
		if !includeDeleted && fields.Deleted {
			continue
		}
		artifact.Project = fields.Project
		if artifact.Class == ArtifactClassContext {
			artifact.Scope = fields.Scope
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
	fields, err := artifactSystemFieldsFromMeta(metaText)
	if err != nil {
		return "", err
	}
	return fields.Project, nil
}

func contextScopeFromMeta(metaText string) (string, error) {
	fields, err := artifactSystemFieldsFromMeta(metaText)
	if err != nil {
		return "", err
	}
	return fields.Scope, nil
}

// ListVisibleContextArtifacts returns global context plus project context visible to the caller.
func ListVisibleContextArtifacts(activeProject, artifactType, scopeFilter, projectFilter string, includeDeleted bool) ([]Artifact, error) {
	if artifactType != "" {
		if _, ok := contextArtifactTypes[strings.TrimSpace(artifactType)]; !ok {
			return nil, fmt.Errorf("invalid context type %q: allowed values are requirement, process_rule, coding_standard, reference, note", artifactType)
		}
	}
	if strings.TrimSpace(scopeFilter) != "" {
		if err := validateContextScope(scopeFilter); err != nil {
			return nil, err
		}
	}
	if strings.TrimSpace(scopeFilter) == ContextScopeGlobal && strings.TrimSpace(projectFilter) != "" {
		return nil, errors.New("--project cannot be used with --scope global")
	}

	targetProject := strings.TrimSpace(projectFilter)
	if targetProject == "" {
		targetProject = strings.TrimSpace(activeProject)
	}

	db, err := InitDB()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = db.Close()
	}()

	rows, err := db.Query(
		`SELECT id, class, type, title, content, metadata, meta, created_at
		FROM intents
		WHERE source_type = 'artifact' AND class = ?
		ORDER BY created_at DESC, id DESC`,
		ArtifactClassContext,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	artifacts := make([]Artifact, 0)
	for rows.Next() {
		artifact, metaText, err := scanArtifactRow(rows)
		if err != nil {
			return nil, err
		}
		fields, err := artifactSystemFieldsFromMeta(metaText)
		if err != nil {
			continue
		}
		if !includeDeleted && fields.Deleted {
			continue
		}
		artifact.Scope = fields.Scope
		artifact.Project = fields.Project

		if artifactType != "" && artifact.Type != strings.TrimSpace(artifactType) {
			continue
		}
		if !contextArtifactVisible(artifact, strings.TrimSpace(scopeFilter), targetProject, strings.TrimSpace(projectFilter) != "") {
			continue
		}
		artifacts = append(artifacts, artifact)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return artifacts, nil
}

// GetVisibleContextArtifact resolves a visible context artifact by full id or unique prefix.
func GetVisibleContextArtifact(idPrefix, activeProject string) (Artifact, error) {
	idPrefix = strings.TrimSpace(idPrefix)
	if idPrefix == "" {
		return Artifact{}, errors.New("context id is required")
	}

	artifacts, err := ListVisibleContextArtifacts(activeProject, "", "", "", true)
	if err != nil {
		return Artifact{}, err
	}

	var matches []Artifact
	for _, artifact := range artifacts {
		if artifact.ID == idPrefix || strings.HasPrefix(artifact.ID, idPrefix) {
			matches = append(matches, artifact)
		}
	}
	switch len(matches) {
	case 0:
		return Artifact{}, fmt.Errorf("context artifact not found: %s", idPrefix)
	case 1:
		return matches[0], nil
	default:
		return Artifact{}, fmt.Errorf("context artifact id is ambiguous: %s", idPrefix)
	}
}

func scanArtifactRow(scanner interface {
	Scan(dest ...any) error
}) (Artifact, string, error) {
	var artifact Artifact
	var metadata sql.NullString
	var meta sql.NullString
	var title sql.NullString
	var content sql.NullString
	if err := scanner.Scan(
		&artifact.ID,
		&artifact.Class,
		&artifact.Type,
		&title,
		&content,
		&metadata,
		&meta,
		&artifact.CreatedAt,
	); err != nil {
		return Artifact{}, "", err
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
	return artifact, meta.String, nil
}

func contextArtifactVisible(artifact Artifact, scopeFilter, targetProject string, projectFilterApplied bool) bool {
	switch scopeFilter {
	case ContextScopeGlobal:
		return artifact.Scope == ContextScopeGlobal
	case ContextScopeProject:
		if targetProject == "" {
			return false
		}
		return artifact.Scope == ContextScopeProject && artifact.Project == targetProject
	}

	if projectFilterApplied {
		return artifact.Scope == ContextScopeProject && artifact.Project == targetProject
	}
	if artifact.Scope == ContextScopeGlobal {
		return true
	}
	return targetProject != "" && artifact.Scope == ContextScopeProject && artifact.Project == targetProject
}
