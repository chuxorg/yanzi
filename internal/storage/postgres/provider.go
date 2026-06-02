package postgres

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/chuxorg/yanzi/internal/config"
	"github.com/chuxorg/yanzi/internal/storage"
	_ "github.com/lib/pq"
)

// Provider is the Postgres storage provider.
type Provider struct {
	db              *sql.DB
	maxOpenConns    int
	maxIdleConns    int
	connMaxLifetime time.Duration
}

// NewProvider opens a Postgres connection using cfg, configures the pool,
// verifies connectivity, and runs migrations.
func NewProvider(cfg config.PostgresConfig) (*Provider, error) {
	dsn := strings.TrimSpace(cfg.DSN)
	if dsn == "" {
		return nil, errors.New("postgres DSN is required")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	maxOpen := cfg.MaxOpenConns
	if maxOpen <= 0 {
		maxOpen = 25
	}
	maxIdle := cfg.MaxIdleConns
	if maxIdle <= 0 {
		maxIdle = 5
	}
	lifetime := cfg.ConnMaxLifetime
	if lifetime <= 0 {
		lifetime = 300
	}

	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxIdle)
	db.SetConnMaxLifetime(time.Duration(lifetime) * time.Second)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	if err := RunMigrations(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("run postgres migrations: %w", err)
	}

	return &Provider{
		db:              db,
		maxOpenConns:    maxOpen,
		maxIdleConns:    maxIdle,
		connMaxLifetime: time.Duration(lifetime) * time.Second,
	}, nil
}

// Name returns the provider identifier.
func (p *Provider) Name() storage.ProviderName {
	return storage.ProviderPostgres
}

// SQLDB returns the underlying database handle.
func (p *Provider) SQLDB() *sql.DB {
	if p == nil {
		return nil
	}
	return p.db
}

// Close closes the provider handle.
func (p *Provider) Close() error {
	if p == nil || p.db == nil {
		return nil
	}
	return p.db.Close()
}

// Health reports internal readiness for the provider.
func (p *Provider) Health(ctx context.Context) storage.Health {
	health := storage.Health{Provider: storage.ProviderPostgres, Status: storage.HealthReady}
	if p == nil || p.db == nil {
		health.Status = storage.HealthUnavailable
		health.Error = storage.ErrProviderUnavailable.Error()
		return health
	}
	if err := p.db.PingContext(ctx); err != nil {
		health.Status = storage.HealthUnavailable
		health.Error = err.Error()
	}
	return health
}

func (p *Provider) Artifacts() bool    { return true }
func (p *Provider) Projects() bool     { return true }
func (p *Provider) Checkpoints() bool  { return true }
func (p *Provider) Verification() bool { return true }
func (p *Provider) ImportExport() bool { return true }

// --- Artifact Operations ---

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

// CreateArtifact stores an artifact using current Postgres artifact semantics.
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

	var metadataValue interface{}
	if input.Metadata != "" {
		metadataValue = input.Metadata
	}
	if _, err := p.db.ExecContext(
		ctx,
		`INSERT INTO intents (
			id, created_at, author, source_type, title, prompt, response, meta, prev_hash, hash, class, type, content, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
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
		WHERE source_type = 'artifact' AND class = $1
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

// --- Project Operations ---

// CreateProject creates a project using current Postgres project semantics.
func (p *Provider) CreateProject(ctx context.Context, input storage.CreateProjectInput) (storage.Project, error) {
	if p == nil || p.db == nil {
		return storage.Project{}, storage.ErrProviderUnavailable
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return storage.Project{}, errors.New("project name is required")
	}

	exists, err := p.ProjectExists(ctx, name)
	if err != nil {
		return storage.Project{}, err
	}
	if exists {
		return storage.Project{}, fmt.Errorf("project already exists: %s", name)
	}

	description := input.Description
	createdAt := time.Now().UTC()
	createdAtText := createdAt.Format(time.RFC3339Nano)
	hash := hashProjectRecord(name, description, createdAtText)
	if _, err := p.db.ExecContext(
		ctx,
		`INSERT INTO projects (name, description, created_at, prev_hash, hash) VALUES ($1, $2, $3, $4, $5)`,
		name,
		description,
		createdAtText,
		nil,
		hash,
	); err != nil {
		if isUniqueViolation(err) {
			return storage.Project{}, fmt.Errorf("project already exists: %s", name)
		}
		return storage.Project{}, err
	}
	return storage.Project{Name: name, Description: description, CreatedAt: createdAt}, nil
}

// ListProjects returns projects ordered by creation time, oldest first.
func (p *Provider) ListProjects(ctx context.Context) ([]storage.Project, error) {
	if p == nil || p.db == nil {
		return nil, storage.ErrProviderUnavailable
	}
	rows, err := p.db.QueryContext(ctx, `SELECT name, description, created_at FROM projects ORDER BY created_at ASC, name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := make([]storage.Project, 0)
	for rows.Next() {
		var project storage.Project
		var description sql.NullString
		var createdAtText string
		if err := rows.Scan(&project.Name, &description, &createdAtText); err != nil {
			return nil, err
		}
		if description.Valid {
			project.Description = description.String
		}
		project.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAtText)
		if err != nil {
			return nil, fmt.Errorf("parse project created_at for %s: %w", project.Name, err)
		}
		projects = append(projects, project)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return projects, nil
}

// ProjectExists checks whether a project row exists for the provided name.
func (p *Provider) ProjectExists(ctx context.Context, name string) (bool, error) {
	if p == nil || p.db == nil {
		return false, storage.ErrProviderUnavailable
	}
	var count int
	if err := p.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM projects WHERE name = $1`, name).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

// --- Checkpoint Operations ---

// CreateCheckpoint creates a checkpoint using current Postgres checkpoint semantics.
func (p *Provider) CreateCheckpoint(ctx context.Context, input storage.CreateCheckpointInput) (storage.Checkpoint, error) {
	if p == nil || p.db == nil {
		return storage.Checkpoint{}, storage.ErrProviderUnavailable
	}
	project := strings.TrimSpace(input.Project)
	if project == "" {
		return storage.Checkpoint{}, errors.New("project is required")
	}
	summary := strings.TrimSpace(input.Summary)
	if summary == "" {
		return storage.Checkpoint{}, errors.New("summary is required")
	}
	exists, err := p.ProjectExists(ctx, project)
	if err != nil {
		return storage.Checkpoint{}, err
	}
	if !exists {
		return storage.Checkpoint{}, fmt.Errorf("project not found: %s", project)
	}

	createdAt := time.Now().UTC().Format(time.RFC3339Nano)
	previousID, err := p.latestCheckpointID(ctx, project)
	if err != nil {
		return storage.Checkpoint{}, err
	}
	checkpoint := normalizeCheckpoint(storage.Checkpoint{
		Project:              project,
		Summary:              summary,
		CreatedAt:            createdAt,
		ArtifactIDs:          input.ArtifactIDs,
		PreviousCheckpointID: previousID,
	})
	hashValue, err := hashCheckpoint(checkpoint)
	if err != nil {
		return storage.Checkpoint{}, err
	}
	checkpoint.Hash = hashValue

	storedIDs := checkpoint.ArtifactIDs
	if storedIDs == nil {
		storedIDs = []string{}
	}
	artifactJSON, err := json.Marshal(storedIDs)
	if err != nil {
		return storage.Checkpoint{}, err
	}
	var prev interface{}
	if checkpoint.PreviousCheckpointID != "" {
		prev = checkpoint.PreviousCheckpointID
	}

	_, err = p.db.ExecContext(
		ctx,
		`INSERT INTO checkpoints (hash, project, summary, created_at, artifact_ids, previous_checkpoint_id)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		checkpoint.Hash,
		checkpoint.Project,
		checkpoint.Summary,
		checkpoint.CreatedAt,
		string(artifactJSON),
		prev,
	)
	if err != nil {
		return storage.Checkpoint{}, err
	}
	return checkpoint, nil
}

// ListCheckpoints returns project checkpoints ordered newest first.
func (p *Provider) ListCheckpoints(ctx context.Context, project string) ([]storage.Checkpoint, error) {
	if p == nil || p.db == nil {
		return nil, storage.ErrProviderUnavailable
	}
	project = strings.TrimSpace(project)
	if project == "" {
		return nil, errors.New("project is required")
	}
	exists, err := p.ProjectExists(ctx, project)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("project not found: %s", project)
	}
	rows, err := p.db.QueryContext(ctx, `SELECT hash, project, summary, created_at, artifact_ids, previous_checkpoint_id FROM checkpoints WHERE project = $1 ORDER BY created_at DESC`, project)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCheckpoints(rows)
}

// ListAllCheckpoints returns all checkpoints using current CLI all-project ordering.
func (p *Provider) ListAllCheckpoints(ctx context.Context) ([]storage.Checkpoint, error) {
	if p == nil || p.db == nil {
		return nil, storage.ErrProviderUnavailable
	}
	rows, err := p.db.QueryContext(ctx, `SELECT hash, project, summary, created_at, artifact_ids, previous_checkpoint_id FROM checkpoints`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	checkpoints, err := scanCheckpoints(rows)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(checkpoints, func(i, j int) bool {
		if checkpoints[i].Project == checkpoints[j].Project {
			if checkpoints[i].CreatedAt == checkpoints[j].CreatedAt {
				return checkpoints[i].Hash > checkpoints[j].Hash
			}
			return checkpoints[i].CreatedAt > checkpoints[j].CreatedAt
		}
		return checkpoints[i].Project < checkpoints[j].Project
	})
	return checkpoints, nil
}

func (p *Provider) latestCheckpointID(ctx context.Context, project string) (string, error) {
	var hash string
	row := p.db.QueryRowContext(ctx, `SELECT hash FROM checkpoints WHERE project = $1 ORDER BY created_at DESC LIMIT 1`, project)
	if err := row.Scan(&hash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return hash, nil
}

// --- Verification Operations ---

// GetVerificationIntent loads an intent by ID using current verification semantics.
func (p *Provider) GetVerificationIntent(ctx context.Context, id string) (storage.IntentRecord, error) {
	if p == nil || p.db == nil {
		return storage.IntentRecord{}, storage.ErrProviderUnavailable
	}
	record, err := p.getVerificationIntent(ctx, `SELECT id, created_at, author, source_type, title, prompt, response, meta, prev_hash, hash FROM intents WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.IntentRecord{}, fmt.Errorf("%w: intent not found: %s", storage.ErrNotFound, id)
		}
		return storage.IntentRecord{}, err
	}
	return record, nil
}

// GetVerificationIntentByHash loads an intent by hash using current chain traversal semantics.
func (p *Provider) GetVerificationIntentByHash(ctx context.Context, intentHash string) (storage.IntentRecord, error) {
	if p == nil || p.db == nil {
		return storage.IntentRecord{}, storage.ErrProviderUnavailable
	}
	record, err := p.getVerificationIntent(ctx, `SELECT id, created_at, author, source_type, title, prompt, response, meta, prev_hash, hash FROM intents WHERE hash = $1`, intentHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.IntentRecord{}, fmt.Errorf("%w: intent hash not found: %s", storage.ErrNotFound, intentHash)
		}
		return storage.IntentRecord{}, err
	}
	return record, nil
}

func (p *Provider) getVerificationIntent(ctx context.Context, query string, arg string) (storage.IntentRecord, error) {
	var record storage.IntentRecord
	var title sql.NullString
	var meta sql.NullString
	var prevHash sql.NullString
	row := p.db.QueryRowContext(ctx, query, arg)
	if err := row.Scan(
		&record.ID,
		&record.CreatedAt,
		&record.Author,
		&record.SourceType,
		&title,
		&record.Prompt,
		&record.Response,
		&meta,
		&prevHash,
		&record.Hash,
	); err != nil {
		return storage.IntentRecord{}, err
	}
	if title.Valid {
		record.Title = title.String
	}
	if meta.Valid && meta.String != "" {
		record.Meta = []byte(meta.String)
	}
	if prevHash.Valid {
		record.PrevHash = prevHash.String
	}
	return record, nil
}

// --- Import/Export Operations ---

// ListExportItems returns the Postgres-backed export timeline source data.
func (p *Provider) ListExportItems(ctx context.Context, query storage.ExportQuery) ([]storage.ExportItem, int, error) {
	if p == nil || p.db == nil {
		return nil, 0, storage.ErrProviderUnavailable
	}

	captures, captureCount, err := p.listExportCaptures(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	if len(query.MetaFilters) > 0 {
		return mergeExportItems(captures, nil), captureCount, nil
	}

	checkpoints, err := p.listExportCheckpoints(ctx, strings.TrimSpace(query.Project))
	if err != nil {
		return nil, 0, err
	}
	return mergeExportItems(captures, checkpoints), captureCount, nil
}

func (p *Provider) listExportCaptures(ctx context.Context, query storage.ExportQuery) ([]storage.ExportItem, int, error) {
	// Postgres uses ctid for physical row ordering as a proxy for insertion order.
	rows, err := p.db.QueryContext(ctx, `SELECT ctid::text, id, created_at, author, source_type, title, prompt, response, hash, meta, metadata
		FROM intents
		WHERE source_type <> 'artifact'
		ORDER BY created_at ASC, ctid ASC`)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	project := strings.TrimSpace(query.Project)
	items := make([]storage.ExportItem, 0)
	captureCount := 0
	for rows.Next() {
		var (
			ctid  string
			item  storage.ExportItem
			meta  sql.NullString
			extra sql.NullString
			title sql.NullString
		)
		item.Kind = storage.ExportItemCapture
		if err := rows.Scan(
			&ctid,
			&item.Capture.ID,
			&item.Capture.CreatedAt,
			&item.Capture.Author,
			&item.Capture.Source,
			&title,
			&item.Capture.Prompt,
			&item.Capture.Response,
			&item.Capture.Hash,
			&meta,
			&extra,
		); err != nil {
			return nil, 0, err
		}
		item.Timestamp = item.Capture.CreatedAt
		// RowID is not available in Postgres the same way as SQLite rowid.
		// Use 0; ordering relies on created_at + ctid instead.
		item.RowID = 0
		if title.Valid {
			item.Capture.Title = title.String
		}

		metadata, err := mergedStringMetadata(meta.String, extra.String)
		if err != nil {
			continue
		}
		if strings.TrimSpace(metadata["project"]) != project {
			continue
		}
		if !query.IncludeDeleted && metadataDeleted(metadata) {
			continue
		}
		if len(query.MetaFilters) > 0 && !stringMetadataMatchesAll(metadata, query.MetaFilters) {
			continue
		}
		item.Capture.Metadata = metadata

		if exportMetaSource(item.Capture.Source) {
			if len(query.MetaFilters) > 0 {
				continue
			}
			item.Kind = storage.ExportItemMeta
			item.Meta = storage.ExportMeta{
				CreatedAt: item.Capture.CreatedAt,
				Command:   strings.TrimSpace(item.Capture.Prompt),
				Value:     strings.TrimSpace(item.Capture.Response),
			}
			item.Capture = storage.ExportCapture{}
			items = append(items, item)
			continue
		}

		captureCount++
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return items, captureCount, nil
}

func (p *Provider) listExportCheckpoints(ctx context.Context, project string) ([]storage.ExportItem, error) {
	rows, err := p.db.QueryContext(ctx, `SELECT ctid::text, hash, summary, created_at
		FROM checkpoints
		WHERE project = $1
		ORDER BY created_at ASC, ctid ASC`, project)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]storage.ExportItem, 0)
	for rows.Next() {
		var ctid string
		var item storage.ExportItem
		item.Kind = storage.ExportItemCheckpoint
		if err := rows.Scan(&ctid, &item.Checkpoint.Hash, &item.Checkpoint.Summary, &item.Checkpoint.CreatedAt); err != nil {
			return nil, err
		}
		item.Timestamp = item.Checkpoint.CreatedAt
		item.RowID = 0
		item.Checkpoint.Project = project
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// --- Shared helpers (mirrored from sqlite package for parity) ---

func hashArtifact(id, createdAt, class, artifactType, title, content, metadata string) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		id, createdAt, class, artifactType, title, content, metadata,
	}, "\n")))
	return hex.EncodeToString(sum[:])
}

func newArtifactID() (string, error) {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("generate artifact id: %w", err)
	}
	return hex.EncodeToString(buf[:]), nil
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

func hashProjectRecord(name, description, createdAt string) string {
	sum := sha256.Sum256([]byte(name + "\n" + description + "\n" + createdAt))
	return hex.EncodeToString(sum[:])
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique") || strings.Contains(msg, "duplicate key")
}

func scanCheckpoints(rows *sql.Rows) ([]storage.Checkpoint, error) {
	checkpoints := make([]storage.Checkpoint, 0)
	for rows.Next() {
		var checkpoint storage.Checkpoint
		var artifactText string
		var prev sql.NullString
		if err := rows.Scan(&checkpoint.Hash, &checkpoint.Project, &checkpoint.Summary, &checkpoint.CreatedAt, &artifactText, &prev); err != nil {
			return nil, err
		}
		if artifactText != "" {
			if err := json.Unmarshal([]byte(artifactText), &checkpoint.ArtifactIDs); err != nil {
				return nil, err
			}
		}
		if prev.Valid {
			checkpoint.PreviousCheckpointID = prev.String
		}
		checkpoints = append(checkpoints, checkpoint)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return checkpoints, nil
}

func hashCheckpoint(checkpoint storage.Checkpoint) (string, error) {
	preimage, err := canonicalCheckpointPreimage(normalizeCheckpoint(checkpoint))
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(preimage)
	return hex.EncodeToString(sum[:]), nil
}

func normalizeCheckpoint(checkpoint storage.Checkpoint) storage.Checkpoint {
	out := checkpoint
	out.Project = normalizeNewlines(strings.TrimSpace(checkpoint.Project))
	out.Summary = normalizeNewlines(strings.TrimSpace(checkpoint.Summary))
	out.PreviousCheckpointID = normalizeNewlines(checkpoint.PreviousCheckpointID)
	if len(out.ArtifactIDs) > 0 {
		ids := make([]string, len(out.ArtifactIDs))
		for i, id := range out.ArtifactIDs {
			ids[i] = normalizeNewlines(id)
		}
		out.ArtifactIDs = ids
	}
	return out
}

func canonicalCheckpointPreimage(checkpoint storage.Checkpoint) ([]byte, error) {
	if strings.TrimSpace(checkpoint.Project) == "" {
		return nil, errors.New("project is required for hashing")
	}
	if strings.TrimSpace(checkpoint.Summary) == "" {
		return nil, errors.New("summary is required for hashing")
	}
	if checkpoint.CreatedAt == "" {
		return nil, errors.New("created_at is required for hashing")
	}
	createdAt, err := normalizeRFC3339(checkpoint.CreatedAt)
	if err != nil {
		return nil, errors.New("created_at must be RFC3339")
	}
	artifactIDs := checkpoint.ArtifactIDs
	if artifactIDs == nil {
		artifactIDs = []string{}
	}
	artifactJSON, err := json.Marshal(artifactIDs)
	if err != nil {
		return nil, err
	}
	var b strings.Builder
	b.WriteByte('{')
	first := true
	addStringField(&b, &first, "project", checkpoint.Project)
	addStringField(&b, &first, "created_at", createdAt)
	addStringField(&b, &first, "summary", checkpoint.Summary)
	addRawField(&b, &first, "artifact_ids", artifactJSON)
	if checkpoint.PreviousCheckpointID != "" {
		addStringField(&b, &first, "previous_checkpoint_id", checkpoint.PreviousCheckpointID)
	}
	b.WriteByte('}')
	return []byte(b.String()), nil
}

func normalizeRFC3339(value string) (string, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return "", err
	}
	return parsed.UTC().Format(time.RFC3339Nano), nil
}

func normalizeNewlines(value string) string {
	if value == "" {
		return value
	}
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	return value
}

func addStringField(b *strings.Builder, first *bool, name string, value string) {
	if !*first {
		b.WriteByte(',')
	}
	*first = false
	b.WriteByte('"')
	b.WriteString(name)
	b.WriteString(`":`)
	encoded, _ := json.Marshal(value)
	b.Write(encoded)
}

func addRawField(b *strings.Builder, first *bool, name string, raw json.RawMessage) {
	if !*first {
		b.WriteByte(',')
	}
	*first = false
	b.WriteByte('"')
	b.WriteString(name)
	b.WriteString(`":`)
	b.Write(raw)
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

func mergeExportItems(captures, checkpoints []storage.ExportItem) []storage.ExportItem {
	merged := make([]storage.ExportItem, 0, len(captures)+len(checkpoints))
	i := 0
	j := 0
	for i < len(captures) && j < len(checkpoints) {
		if captures[i].Timestamp < checkpoints[j].Timestamp {
			merged = append(merged, captures[i])
			i++
			continue
		}
		if captures[i].Timestamp > checkpoints[j].Timestamp {
			merged = append(merged, checkpoints[j])
			j++
			continue
		}
		if captures[i].RowID <= checkpoints[j].RowID {
			merged = append(merged, captures[i])
			i++
			continue
		}
		merged = append(merged, checkpoints[j])
		j++
	}
	merged = append(merged, captures[i:]...)
	merged = append(merged, checkpoints[j:]...)
	return merged
}

func mergedStringMetadata(primary, secondary string) (map[string]string, error) {
	metadata := map[string]string{}
	for _, raw := range []string{primary, secondary} {
		if strings.TrimSpace(raw) == "" {
			continue
		}
		decoded, err := decodeStringMetadata(raw)
		if err != nil {
			return nil, err
		}
		for key, value := range decoded {
			metadata[key] = value
		}
	}
	return metadata, nil
}

func decodeStringMetadata(raw string) (map[string]string, error) {
	var metadata map[string]string
	if err := json.Unmarshal([]byte(raw), &metadata); err != nil {
		return nil, err
	}
	return metadata, nil
}

func metadataDeleted(metadata map[string]string) bool {
	return strings.EqualFold(strings.TrimSpace(metadata["deleted"]), "true")
}

func stringMetadataMatchesAll(metadata, filters map[string]string) bool {
	for key, value := range filters {
		if metadata[key] != value {
			return false
		}
	}
	return true
}

func exportMetaSource(source string) bool {
	value := strings.ToLower(strings.TrimSpace(source))
	return value == "meta-command" || value == "meta_command" || value == "event"
}
