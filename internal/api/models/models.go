package models

// ArtifactSummary represents the current operational API artifact list payload.
type ArtifactSummary struct {
	ID        string            `json:"id"`
	CreatedAt string            `json:"created_at"`
	Project   string            `json:"project,omitempty"`
	Author    string            `json:"author"`
	Source    string            `json:"source"`
	Title     string            `json:"title"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// Artifact represents the current operational API artifact detail payload.
type Artifact struct {
	ID        string            `json:"id"`
	CreatedAt string            `json:"created_at"`
	Project   string            `json:"project,omitempty"`
	Author    string            `json:"author"`
	Source    string            `json:"source"`
	Title     string            `json:"title"`
	Prompt    string            `json:"prompt"`
	Response  string            `json:"response"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	PrevHash  string            `json:"prev_hash,omitempty"`
	Hash      string            `json:"hash"`
}

// ArtifactResponse is the detail response for artifact reads.
type ArtifactResponse struct {
	Artifact Artifact `json:"artifact"`
}

// ArtifactListResponse is the collection response for artifact queries.
type ArtifactListResponse struct {
	Artifacts []ArtifactSummary `json:"artifacts"`
}

// ArtifactCaptureRequest captures the POST /v0/artifacts capture payload.
type ArtifactCaptureRequest struct {
	Author     string            `json:"author"`
	SourceType string            `json:"source_type,omitempty"`
	Title      string            `json:"title,omitempty"`
	Prompt     string            `json:"prompt"`
	Response   string            `json:"response"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	Project    string            `json:"project,omitempty"`
	PrevHash   string            `json:"prev_hash,omitempty"`
}

// ArtifactCaptureResponse is the deterministic capture artifact response.
type ArtifactCaptureResponse struct {
	ID         string            `json:"id"`
	CreatedAt  string            `json:"created_at"`
	Author     string            `json:"author"`
	SourceType string            `json:"source_type"`
	Title      string            `json:"title,omitempty"`
	Prompt     string            `json:"prompt"`
	Response   string            `json:"response"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	PrevHash   string            `json:"prev_hash,omitempty"`
	Hash       string            `json:"hash"`
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
