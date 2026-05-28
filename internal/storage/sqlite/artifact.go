package sqlite

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/chuxorg/yanzi/internal/storage"
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

// CreateArtifact stores an artifact using current SQLite artifact semantics.
func (p *Provider) CreateArtifact(ctx context.Context, input storage.CreateArtifactInput) (storage.Artifact, error) {
	if p == nil || p.db == nil {
		return storage.Artifact{}, storage.ErrProviderUnavailable
	}
	class := strings.TrimSpace(input.Class)
	scope := strings.TrimSpace(input.Scope)
	if class == storage.ArtifactClassContext && scope == "" {
		scope = storage.ContextScopeProject
	}
	if err := validateArtifactInput(input.Project, class, input.Type, input.Title, input.Content, scope); err != nil {
		return storage.Artifact{}, err
	}
	project := strings.TrimSpace(input.Project)
	if project != "" {
		exists, err := p.ProjectExists(ctx, project)
		if err != nil {
			return storage.Artifact{}, err
		}
		if !exists {
			return storage.Artifact{}, fmt.Errorf("project not found: %s", project)
		}
	}

	id, err := newArtifactID()
	if err != nil {
		return storage.Artifact{}, err
	}
	createdAt := time.Now().UTC().Format(time.RFC3339Nano)
	systemMeta, err := artifactSystemMeta(project, scope)
	if err != nil {
		return storage.Artifact{}, err
	}
	hashValue := hashArtifact(id, createdAt, class, strings.TrimSpace(input.Type), input.Title, input.Content, input.Metadata)

	var metadataValue any
	if input.Metadata != "" {
		metadataValue = input.Metadata
	}
	if _, err := p.db.ExecContext(
		ctx,
		`INSERT INTO intents (
			id, created_at, author, source_type, title, prompt, response, meta, prev_hash, hash, class, type, content, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id,
		createdAt,
		"yanzi",
		"artifact",
		input.Title,
		input.Content,
		input.Content,
		systemMeta,
		nil,
		hashValue,
		class,
		strings.TrimSpace(input.Type),
		input.Content,
		metadataValue,
	); err != nil {
		return storage.Artifact{}, err
	}
	return storage.Artifact{
		ID:        id,
		Class:     class,
		Type:      strings.TrimSpace(input.Type),
		Scope:     scope,
		Project:   project,
		Title:     input.Title,
		Content:   input.Content,
		Metadata:  input.Metadata,
		CreatedAt: createdAt,
	}, nil
}

// ListArtifacts lists artifacts using current project/class/type filtering semantics.
func (p *Provider) ListArtifacts(ctx context.Context, query storage.ArtifactQuery) ([]storage.Artifact, error) {
	if p == nil || p.db == nil {
		return nil, storage.ErrProviderUnavailable
	}
	class := strings.TrimSpace(query.Class)
	artifactType := strings.TrimSpace(query.Type)
	if err := validateArtifactClassAndType(class, artifactType); err != nil {
		return nil, err
	}
	artifacts, err := p.listArtifactsAllProjects(ctx, class, artifactType, query.IncludeDeleted)
	if err != nil {
		return nil, err
	}
	project := strings.TrimSpace(query.Project)
	if project == "" {
		return artifacts, nil
	}
	filtered := make([]storage.Artifact, 0, len(artifacts))
	for _, artifact := range artifacts {
		if artifact.Project == project {
			filtered = append(filtered, artifact)
		}
	}
	return filtered, nil
}

// ListVisibleContextArtifacts lists context artifacts with current visibility rules.
func (p *Provider) ListVisibleContextArtifacts(ctx context.Context, query storage.ContextArtifactQuery) ([]storage.Artifact, error) {
	if p == nil || p.db == nil {
		return nil, storage.ErrProviderUnavailable
	}
	artifactType := strings.TrimSpace(query.Type)
	if artifactType != "" {
		if _, ok := contextArtifactTypes[artifactType]; !ok {
			return nil, fmt.Errorf("invalid context type %q: allowed values are requirement, process_rule, coding_standard, reference, note", query.Type)
		}
	}
	scopeFilter := strings.TrimSpace(query.Scope)
	if scopeFilter != "" {
		if err := validateContextScope(scopeFilter); err != nil {
			return nil, err
		}
	}
	projectFilter := strings.TrimSpace(query.Project)
	if !query.AllProjects && scopeFilter == storage.ContextScopeGlobal && projectFilter != "" {
		return nil, errors.New("--project cannot be used with --scope global")
	}

	artifacts, err := p.listArtifactsAllProjects(ctx, storage.ArtifactClassContext, artifactType, query.IncludeDeleted)
	if err != nil {
		return nil, err
	}
	if query.AllProjects {
		return filterAllProjectContextArtifacts(artifacts, scopeFilter), nil
	}

	targetProject := projectFilter
	if targetProject == "" {
		targetProject = strings.TrimSpace(query.ActiveProject)
	}
	filtered := make([]storage.Artifact, 0, len(artifacts))
	for _, artifact := range artifacts {
		if !contextArtifactVisible(artifact, scopeFilter, targetProject, projectFilter != "") {
			continue
		}
		filtered = append(filtered, artifact)
	}
	return filtered, nil
}

// GetVisibleContextArtifact resolves a visible context artifact by full id or unique prefix.
func (p *Provider) GetVisibleContextArtifact(ctx context.Context, idPrefix, activeProject string) (storage.Artifact, error) {
	idPrefix = strings.TrimSpace(idPrefix)
	if idPrefix == "" {
		return storage.Artifact{}, errors.New("context id is required")
	}
	artifacts, err := p.ListVisibleContextArtifacts(ctx, storage.ContextArtifactQuery{ActiveProject: activeProject, IncludeDeleted: true})
	if err != nil {
		return storage.Artifact{}, err
	}

	var matches []storage.Artifact
	for _, artifact := range artifacts {
		if artifact.ID == idPrefix || strings.HasPrefix(artifact.ID, idPrefix) {
			matches = append(matches, artifact)
		}
	}
	switch len(matches) {
	case 0:
		return storage.Artifact{}, fmt.Errorf("context artifact not found: %s", idPrefix)
	case 1:
		return matches[0], nil
	default:
		return storage.Artifact{}, fmt.Errorf("context artifact id is ambiguous: %s", idPrefix)
	}
}

func (p *Provider) listArtifactsAllProjects(ctx context.Context, class, artifactType string, includeDeleted bool) ([]storage.Artifact, error) {
	rows, err := p.db.QueryContext(
		ctx,
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

	artifacts := make([]storage.Artifact, 0)
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
		artifact.Project = fields.Project
		if artifact.Class == storage.ArtifactClassContext {
			artifact.Scope = fields.Scope
		}
		if artifactType != "" && artifact.Type != artifactType {
			continue
		}
		artifacts = append(artifacts, artifact)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return artifacts, nil
}

func validateArtifactInput(projectID, class, artifactType, title, content, scope string) error {
	classValue := strings.TrimSpace(class)
	if classValue != storage.ArtifactClassIntent && classValue != storage.ArtifactClassContext {
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
	case storage.ArtifactClassIntent:
		if strings.TrimSpace(projectID) == "" {
			return errors.New("project is required")
		}
		allowed = intentArtifactTypes
	case storage.ArtifactClassContext:
		if err := validateContextScope(scope); err != nil {
			return err
		}
		if strings.TrimSpace(scope) == storage.ContextScopeProject && strings.TrimSpace(projectID) == "" {
			return errors.New("project is required for project-scoped context")
		}
		allowed = contextArtifactTypes
	}
	if _, ok := allowed[typeValue]; !ok {
		if classValue == storage.ArtifactClassIntent {
			return fmt.Errorf("invalid intent type %q: allowed values are prompt, decision, task, change_request, checkpoint, note", artifactType)
		}
		return fmt.Errorf("invalid context type %q: allowed values are requirement, process_rule, coding_standard, reference, note", artifactType)
	}
	return nil
}

func validateArtifactClassAndType(class, artifactType string) error {
	class = strings.TrimSpace(class)
	switch class {
	case storage.ArtifactClassIntent:
		if artifactType == "" {
			return nil
		}
		if _, ok := intentArtifactTypes[strings.TrimSpace(artifactType)]; ok {
			return nil
		}
		return fmt.Errorf("invalid intent type %q: allowed values are prompt, decision, task, change_request, checkpoint, note", artifactType)
	case storage.ArtifactClassContext:
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

func validateContextScope(scope string) error {
	switch strings.TrimSpace(scope) {
	case storage.ContextScopeGlobal, storage.ContextScopeProject:
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
			fields.Scope = storage.ContextScopeProject
		} else {
			fields.Scope = storage.ContextScopeGlobal
		}
	}
	return fields, nil
}

func scanArtifactRow(scanner interface {
	Scan(dest ...any) error
}) (storage.Artifact, string, error) {
	var artifact storage.Artifact
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
		return storage.Artifact{}, "", err
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

func contextArtifactVisible(artifact storage.Artifact, scopeFilter, targetProject string, projectFilterApplied bool) bool {
	switch scopeFilter {
	case storage.ContextScopeGlobal:
		return artifact.Scope == storage.ContextScopeGlobal
	case storage.ContextScopeProject:
		if targetProject == "" {
			return false
		}
		return artifact.Scope == storage.ContextScopeProject && artifact.Project == targetProject
	}

	if projectFilterApplied {
		return artifact.Scope == storage.ContextScopeProject && artifact.Project == targetProject
	}
	if artifact.Scope == storage.ContextScopeGlobal {
		return true
	}
	return targetProject != "" && artifact.Scope == storage.ContextScopeProject && artifact.Project == targetProject
}

func filterAllProjectContextArtifacts(artifacts []storage.Artifact, scopeFilter string) []storage.Artifact {
	filtered := make([]storage.Artifact, 0, len(artifacts))
	for _, artifact := range artifacts {
		switch scopeFilter {
		case "":
		case storage.ContextScopeGlobal:
			if artifact.Scope != storage.ContextScopeGlobal {
				continue
			}
		case storage.ContextScopeProject:
			if artifact.Scope != storage.ContextScopeProject {
				continue
			}
		}
		filtered = append(filtered, artifact)
	}
	return filtered
}
