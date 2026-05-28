package models

// Artifact represents the current operational API artifact payload.
type Artifact struct {
	ID        string            `json:"id"`
	Class     string            `json:"class"`
	Type      string            `json:"type"`
	Scope     string            `json:"scope,omitempty"`
	Project   string            `json:"project,omitempty"`
	Title     string            `json:"title"`
	Content   string            `json:"content"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt string            `json:"created_at"`
}

// ArtifactCreateRequest captures the current artifact creation shape.
type ArtifactCreateRequest struct {
	Project  string            `json:"project,omitempty"`
	Class    string            `json:"class"`
	Type     string            `json:"type"`
	Scope    string            `json:"scope,omitempty"`
	Title    string            `json:"title"`
	Content  string            `json:"content"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// ArtifactListResponse is the collection response for artifact queries.
type ArtifactListResponse struct {
	Artifacts []Artifact `json:"artifacts"`
}

// Project represents the current operational API project payload.
type Project struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"created_at"`
}

// ProjectCreateRequest captures the current project creation shape.
type ProjectCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// ProjectListResponse is the collection response for project queries.
type ProjectListResponse struct {
	Projects []Project `json:"projects"`
}

// Checkpoint represents the current operational API checkpoint payload.
type Checkpoint struct {
	Hash                 string   `json:"hash"`
	Project              string   `json:"project"`
	Summary              string   `json:"summary"`
	CreatedAt            string   `json:"created_at"`
	ArtifactIDs          []string `json:"artifact_ids,omitempty"`
	PreviousCheckpointID string   `json:"previous_checkpoint_id,omitempty"`
}

// CheckpointCreateRequest captures the current checkpoint creation shape.
type CheckpointCreateRequest struct {
	Project     string   `json:"project"`
	Summary     string   `json:"summary"`
	ArtifactIDs []string `json:"artifact_ids,omitempty"`
}

// CheckpointListResponse is the collection response for checkpoint queries.
type CheckpointListResponse struct {
	Checkpoints []Checkpoint `json:"checkpoints"`
}

// RehydrateCheckpoint represents the current operational API rehydration checkpoint payload.
type RehydrateCheckpoint struct {
	Hash                 string   `json:"hash"`
	Project              string   `json:"project"`
	Summary              string   `json:"summary"`
	CreatedAt            string   `json:"created_at"`
	ArtifactIDs          []string `json:"artifact_ids,omitempty"`
	PreviousCheckpointID string   `json:"previous_checkpoint_id,omitempty"`
}

// RehydrateIntent represents the current operational API rehydration intent payload.
type RehydrateIntent struct {
	ID              string            `json:"id"`
	Timestamp       string            `json:"timestamp"`
	Author          string            `json:"author"`
	SourceType      string            `json:"source_type"`
	Title           string            `json:"title,omitempty"`
	Prompt          string            `json:"prompt"`
	Response        string            `json:"response"`
	PromptSnippet   string            `json:"prompt_snippet"`
	ResponseSnippet string            `json:"response_snippet"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	Hash            string            `json:"hash"`
	PrevHash        string            `json:"prev_hash,omitempty"`
}

// RehydrateResponse is the deterministic operational API rehydration payload.
type RehydrateResponse struct {
	Project        string               `json:"project"`
	HasCheckpoint  bool                 `json:"has_checkpoint"`
	Fallback       bool                 `json:"fallback"`
	FallbackReason string               `json:"fallback_reason,omitempty"`
	FallbackLimit  int                  `json:"fallback_limit,omitempty"`
	Checkpoint     *RehydrateCheckpoint `json:"checkpoint,omitempty"`
	Intents        []RehydrateIntent    `json:"intents"`
}

// ProviderHealth represents the current provider health payload for API status reads.
type ProviderHealth struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// HealthResponse is the minimal operational health/status response.
type HealthResponse struct {
	Version  string         `json:"version"`
	Mode     string         `json:"mode"`
	Provider ProviderHealth `json:"provider"`
}

// StatusResponse is the generic deterministic status payload for non-CRUD route groups.
type StatusResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
