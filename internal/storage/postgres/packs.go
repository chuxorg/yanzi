package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/chuxorg/yanzi/internal/packs"
)

var _ packs.PackStore = (*PostgresPackStore)(nil)

// PostgresPackStore implements packs.PackStore backed by the Postgres provider.
type PostgresPackStore struct {
	db *sql.DB
}

// NewPackStore returns a PackStore backed by db.
func NewPackStore(db *sql.DB) *PostgresPackStore {
	return &PostgresPackStore{db: db}
}

// CreateSeed inserts a new seed record.
func (s *PostgresPackStore) CreateSeed(ctx context.Context, seed packs.Seed) (packs.Seed, error) {
	id, err := newArtifactID()
	if err != nil {
		return packs.Seed{}, err
	}
	seed.ID = id

	contentJSON, err := json.Marshal(seed.Content)
	if err != nil {
		return packs.Seed{}, fmt.Errorf("marshal seed content: %w", err)
	}
	tagsJSON, err := json.Marshal(seed.Tags)
	if err != nil {
		return packs.Seed{}, fmt.Errorf("marshal seed tags: %w", err)
	}

	now := time.Now().UTC()
	if seed.CreatedAt.IsZero() {
		seed.CreatedAt = now
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO seeds
		 (id, artifact_id, name, version_label, seed_type, role_access_bits,
		  description, content, token_estimate, tags, author_role, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		seed.ID, seed.ArtifactID, seed.Name, seed.VersionLabel, seed.SeedType,
		int(seed.RoleAccessBits), seed.Description, string(contentJSON),
		seed.TokenEstimate, string(tagsJSON), seed.AuthorRole, seed.CreatedAt.UTC(),
	)
	if err != nil {
		return packs.Seed{}, fmt.Errorf("insert seed: %w", err)
	}
	return seed, nil
}

// GetSeed returns a seed by artifact ID.
func (s *PostgresPackStore) GetSeed(ctx context.Context, artifactID string) (packs.Seed, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, artifact_id, name, version_label, seed_type, role_access_bits,
		        description, content, token_estimate, tags, author_role, created_at
		 FROM seeds WHERE artifact_id = $1`,
		artifactID,
	)
	seed, err := scanPGSeed(row)
	if errors.Is(err, sql.ErrNoRows) {
		return packs.Seed{}, fmt.Errorf("%w: seed not found: %s", packs.ErrNotFound, artifactID)
	}
	return seed, err
}

// GetSeedByName returns the most recently created seed with the given name.
func (s *PostgresPackStore) GetSeedByName(ctx context.Context, name string) (packs.Seed, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, artifact_id, name, version_label, seed_type, role_access_bits,
		        description, content, token_estimate, tags, author_role, created_at
		 FROM seeds WHERE name = $1 ORDER BY created_at DESC LIMIT 1`,
		name,
	)
	seed, err := scanPGSeed(row)
	if errors.Is(err, sql.ErrNoRows) {
		return packs.Seed{}, fmt.Errorf("%w: seed not found: %s", packs.ErrNotFound, name)
	}
	return seed, err
}

// ListSeeds returns seeds matching filter.
func (s *PostgresPackStore) ListSeeds(ctx context.Context, filter packs.SeedFilter) ([]packs.Seed, error) {
	query := `SELECT id, artifact_id, name, version_label, seed_type, role_access_bits,
		             description, content, token_estimate, tags, author_role, created_at
		      FROM seeds WHERE TRUE`
	args := []any{}
	n := 1
	if filter.SeedType != "" {
		query += fmt.Sprintf(" AND seed_type = $%d", n)
		args = append(args, filter.SeedType)
		n++
	}
	if filter.MinRoleBits > 0 {
		query += fmt.Sprintf(" AND role_access_bits >= $%d", n)
		args = append(args, int(filter.MinRoleBits))
		n++
	}
	if filter.Name != "" {
		query += fmt.Sprintf(" AND name = $%d", n)
		args = append(args, filter.Name)
		n++
	}
	_ = n
	query += " ORDER BY created_at DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list seeds: %w", err)
	}
	defer rows.Close()

	var result []packs.Seed
	for rows.Next() {
		var (
			id, artifactID, name, seedType, contentStr, tagsStr string
			roleAccessBits, tokenEstimate                       int
			createdAt                                           time.Time
			nullVersionLabel, nullDescription, nullAuthorRole   sql.NullString
		)
		if err := rows.Scan(
			&id, &artifactID, &name, &nullVersionLabel, &seedType, &roleAccessBits,
			&nullDescription, &contentStr, &tokenEstimate, &tagsStr, &nullAuthorRole, &createdAt,
		); err != nil {
			return nil, fmt.Errorf("scan seed row: %w", err)
		}
		seed, err := buildPGSeed(id, artifactID, name, nullVersionLabel.String, seedType,
			roleAccessBits, nullDescription.String, contentStr, tokenEstimate, tagsStr,
			nullAuthorRole.String, createdAt)
		if err != nil {
			return nil, err
		}
		result = append(result, seed)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate seeds: %w", err)
	}
	return result, nil
}

// DeleteSeed removes a seed record.
func (s *PostgresPackStore) DeleteSeed(ctx context.Context, artifactID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM seeds WHERE artifact_id = $1`, artifactID)
	if err != nil {
		return fmt.Errorf("delete seed: %w", err)
	}
	return nil
}

// CreatePack inserts a new pack record.
func (s *PostgresPackStore) CreatePack(ctx context.Context, pack packs.Pack) (packs.Pack, error) {
	id, err := newArtifactID()
	if err != nil {
		return packs.Pack{}, err
	}
	pack.ID = id

	seedsJSON, err := json.Marshal(pack.Seeds)
	if err != nil {
		return packs.Pack{}, fmt.Errorf("marshal pack seeds: %w", err)
	}
	tokenEstJSON, err := json.Marshal(pack.TokenEstimate)
	if err != nil {
		return packs.Pack{}, fmt.Errorf("marshal token estimate: %w", err)
	}
	tagsJSON, err := json.Marshal(pack.Tags)
	if err != nil {
		return packs.Pack{}, fmt.Errorf("marshal pack tags: %w", err)
	}

	now := time.Now().UTC()
	if pack.CreatedAt.IsZero() {
		pack.CreatedAt = now
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO packs
		 (id, artifact_id, name, version_label, extends_id, role_bits, role_label,
		  description, pack_context, seeds, token_estimate, tags, author_role, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		pack.ID, pack.ArtifactID, pack.Name, pack.VersionLabel, pack.ExtendsID,
		int(pack.Role), pack.RoleLabel, pack.Description, pack.PackContext,
		string(seedsJSON), string(tokenEstJSON), string(tagsJSON), pack.AuthorRole,
		pack.CreatedAt.UTC(),
	)
	if err != nil {
		return packs.Pack{}, fmt.Errorf("insert pack: %w", err)
	}
	return pack, nil
}

// GetPack returns a pack by artifact ID.
func (s *PostgresPackStore) GetPack(ctx context.Context, artifactID string) (packs.Pack, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, artifact_id, name, version_label, extends_id, role_bits, role_label,
		        description, pack_context, seeds, token_estimate, tags, author_role, created_at
		 FROM packs WHERE artifact_id = $1`,
		artifactID,
	)
	p, err := scanPGPack(row)
	if errors.Is(err, sql.ErrNoRows) {
		return packs.Pack{}, fmt.Errorf("%w: pack not found: %s", packs.ErrNotFound, artifactID)
	}
	return p, err
}

// GetPackByName returns the most recently created pack with the given name.
func (s *PostgresPackStore) GetPackByName(ctx context.Context, name string) (packs.Pack, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, artifact_id, name, version_label, extends_id, role_bits, role_label,
		        description, pack_context, seeds, token_estimate, tags, author_role, created_at
		 FROM packs WHERE name = $1 ORDER BY created_at DESC LIMIT 1`,
		name,
	)
	p, err := scanPGPack(row)
	if errors.Is(err, sql.ErrNoRows) {
		return packs.Pack{}, fmt.Errorf("%w: pack not found: %s", packs.ErrNotFound, name)
	}
	return p, err
}

// ListPacks returns packs matching filter.
func (s *PostgresPackStore) ListPacks(ctx context.Context, filter packs.PackFilter) ([]packs.Pack, error) {
	query := `SELECT id, artifact_id, name, version_label, extends_id, role_bits, role_label,
		             description, pack_context, seeds, token_estimate, tags, author_role, created_at
		      FROM packs WHERE TRUE`
	args := []any{}
	n := 1
	if filter.Role > 0 {
		query += fmt.Sprintf(" AND role_bits = $%d", n)
		args = append(args, int(filter.Role))
		n++
	}
	if filter.Name != "" {
		query += fmt.Sprintf(" AND name = $%d", n)
		args = append(args, filter.Name)
		n++
	}
	_ = n
	query += " ORDER BY created_at DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list packs: %w", err)
	}
	defer rows.Close()

	var result []packs.Pack
	for rows.Next() {
		var (
			id, artifactID, name, seedsStr, tokenEstStr, tagsStr string
			roleBits                                             int
			createdAt                                           time.Time
			nullVersionLabel, nullExtendsID, nullRoleLabel       sql.NullString
			nullDescription, nullPackContext, nullAuthorRole     sql.NullString
		)
		if err := rows.Scan(
			&id, &artifactID, &name, &nullVersionLabel, &nullExtendsID, &roleBits, &nullRoleLabel,
			&nullDescription, &nullPackContext, &seedsStr, &tokenEstStr, &tagsStr, &nullAuthorRole, &createdAt,
		); err != nil {
			return nil, fmt.Errorf("scan pack row: %w", err)
		}
		p, err := buildPGPack(id, artifactID, name, nullVersionLabel.String, nullExtendsID.String,
			roleBits, nullRoleLabel.String, nullDescription.String, nullPackContext.String,
			seedsStr, tokenEstStr, tagsStr, nullAuthorRole.String, createdAt)
		if err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate packs: %w", err)
	}
	return result, nil
}

// DeletePack removes a pack record.
func (s *PostgresPackStore) DeletePack(ctx context.Context, artifactID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM packs WHERE artifact_id = $1`, artifactID)
	if err != nil {
		return fmt.Errorf("delete pack: %w", err)
	}
	return nil
}

// RecordTokenUsage inserts a token usage record.
func (s *PostgresPackStore) RecordTokenUsage(ctx context.Context, usage packs.TokenUsage) error {
	id, err := newArtifactID()
	if err != nil {
		return err
	}
	if usage.ID == "" {
		usage.ID = id
	}
	if usage.RecordedAt.IsZero() {
		usage.RecordedAt = time.Now().UTC()
	}
	approx := 0
	if usage.Approximate {
		approx = 1
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO token_usage
		 (id, project, phase, task, artifact_id, pack_id, token_count, approximate, model_hint, recorded_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		usage.ID, usage.Project, usage.Phase, usage.Task, usage.ArtifactID, usage.PackID,
		usage.TokenCount, approx, usage.ModelHint, usage.RecordedAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("insert token usage: %w", err)
	}
	return nil
}

// GetTokenUsage returns aggregated token usage matching filter.
func (s *PostgresPackStore) GetTokenUsage(ctx context.Context, filter packs.TokenFilter) (packs.TokenUsageSummary, error) {
	query := `SELECT phase, task, token_count, approximate FROM token_usage WHERE project = $1`
	args := []any{filter.Project}
	n := 2
	if filter.Phase != "" {
		query += fmt.Sprintf(" AND phase = $%d", n)
		args = append(args, filter.Phase)
		n++
	}
	if filter.Task != "" {
		query += fmt.Sprintf(" AND task = $%d", n)
		args = append(args, filter.Task)
		n++
	}
	if !filter.Since.IsZero() {
		query += fmt.Sprintf(" AND recorded_at >= $%d", n)
		args = append(args, filter.Since.UTC())
		n++
	}
	_ = n

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return packs.TokenUsageSummary{}, fmt.Errorf("query token usage: %w", err)
	}
	defer rows.Close()

	summary := packs.TokenUsageSummary{
		Project: filter.Project,
		ByPhase: map[string]int{},
		ByTask:  map[string]int{},
	}
	for rows.Next() {
		var phase, task sql.NullString
		var count, approxInt int
		if err := rows.Scan(&phase, &task, &count, &approxInt); err != nil {
			return packs.TokenUsageSummary{}, fmt.Errorf("scan token usage: %w", err)
		}
		summary.TotalTokens += count
		if phase.Valid && phase.String != "" {
			summary.ByPhase[phase.String] += count
		}
		if task.Valid && task.String != "" {
			summary.ByTask[task.String] += count
		}
		if approxInt == 1 {
			summary.Approximate = true
		}
	}
	if err := rows.Err(); err != nil {
		return packs.TokenUsageSummary{}, fmt.Errorf("iterate token usage: %w", err)
	}
	return summary, nil
}

func scanPGSeed(row *sql.Row) (packs.Seed, error) {
	var (
		id, artifactID, name, seedType, contentStr, tagsStr string
		roleAccessBits, tokenEstimate                       int
		createdAt                                           time.Time
		nullVersionLabel, nullDescription, nullAuthorRole   sql.NullString
	)
	if err := row.Scan(
		&id, &artifactID, &name, &nullVersionLabel, &seedType, &roleAccessBits,
		&nullDescription, &contentStr, &tokenEstimate, &tagsStr, &nullAuthorRole, &createdAt,
	); err != nil {
		return packs.Seed{}, err
	}
	return buildPGSeed(id, artifactID, name, nullVersionLabel.String, seedType,
		roleAccessBits, nullDescription.String, contentStr, tokenEstimate, tagsStr,
		nullAuthorRole.String, createdAt)
}

func buildPGSeed(id, artifactID, name, versionLabel, seedType string,
	roleAccessBits int, description, contentStr string, tokenEstimate int,
	tagsStr, authorRole string, createdAt time.Time) (packs.Seed, error) {

	var content packs.SeedContent
	if err := json.Unmarshal([]byte(contentStr), &content); err != nil {
		return packs.Seed{}, fmt.Errorf("unmarshal seed content: %w", err)
	}
	var tags []string
	if tagsStr != "" && tagsStr != "null" {
		if err := json.Unmarshal([]byte(tagsStr), &tags); err != nil {
			return packs.Seed{}, fmt.Errorf("unmarshal seed tags: %w", err)
		}
	}
	return packs.Seed{
		ID:             id,
		ArtifactID:     artifactID,
		Name:           name,
		VersionLabel:   versionLabel,
		SeedType:       seedType,
		RoleAccessBits: packs.RoleBits(roleAccessBits),
		Description:    description,
		Content:        content,
		TokenEstimate:  tokenEstimate,
		Tags:           tags,
		AuthorRole:     authorRole,
		CreatedAt:      createdAt,
	}, nil
}

func scanPGPack(row *sql.Row) (packs.Pack, error) {
	var (
		id, artifactID, name, seedsStr, tokenEstStr, tagsStr string
		roleBits                                             int
		createdAt                                           time.Time
		nullVersionLabel, nullExtendsID, nullRoleLabel       sql.NullString
		nullDescription, nullPackContext, nullAuthorRole     sql.NullString
	)
	if err := row.Scan(
		&id, &artifactID, &name, &nullVersionLabel, &nullExtendsID, &roleBits, &nullRoleLabel,
		&nullDescription, &nullPackContext, &seedsStr, &tokenEstStr, &tagsStr, &nullAuthorRole, &createdAt,
	); err != nil {
		return packs.Pack{}, err
	}
	return buildPGPack(id, artifactID, name, nullVersionLabel.String, nullExtendsID.String,
		roleBits, nullRoleLabel.String, nullDescription.String, nullPackContext.String,
		seedsStr, tokenEstStr, tagsStr, nullAuthorRole.String, createdAt)
}

func buildPGPack(id, artifactID, name, versionLabel, extendsID string,
	roleBits int, roleLabel, description, packContext string,
	seedsStr, tokenEstStr, tagsStr, authorRole string, createdAt time.Time) (packs.Pack, error) {

	var seedRefs []packs.SeedReference
	if seedsStr != "" && seedsStr != "null" {
		if err := json.Unmarshal([]byte(seedsStr), &seedRefs); err != nil {
			return packs.Pack{}, fmt.Errorf("unmarshal pack seeds: %w", err)
		}
	}
	var tokenEst packs.PackTokenEstimate
	if tokenEstStr != "" && tokenEstStr != "null" {
		if err := json.Unmarshal([]byte(tokenEstStr), &tokenEst); err != nil {
			return packs.Pack{}, fmt.Errorf("unmarshal token estimate: %w", err)
		}
	}
	var tags []string
	if tagsStr != "" && tagsStr != "null" {
		if err := json.Unmarshal([]byte(tagsStr), &tags); err != nil {
			return packs.Pack{}, fmt.Errorf("unmarshal pack tags: %w", err)
		}
	}
	return packs.Pack{
		ID:            id,
		ArtifactID:    artifactID,
		Name:          name,
		VersionLabel:  versionLabel,
		ExtendsID:     extendsID,
		Role:          packs.RoleBits(roleBits),
		RoleLabel:     roleLabel,
		Description:   description,
		PackContext:   packContext,
		Seeds:         seedRefs,
		TokenEstimate: tokenEst,
		Tags:          tags,
		AuthorRole:    authorRole,
		CreatedAt:     createdAt,
	}, nil
}
