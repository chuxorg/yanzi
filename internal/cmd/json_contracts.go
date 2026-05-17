package cmd

const (
	machineContractSchemaVersion = 1

	jsonKindHistoryExport = "history_export"
	jsonKindContextExport = "context_export"
	jsonKindRehydrate     = "rehydrate"
	jsonKindStatus        = "status"
	jsonKindArtifactTypes = "artifact_types"
)

type contractJSONCheckpoint struct {
	Hash                 string   `json:"hash"`
	Project              string   `json:"project"`
	Summary              string   `json:"summary"`
	CreatedAt            string   `json:"created_at"`
	ArtifactIDs          []string `json:"artifact_ids"`
	PreviousCheckpointID string   `json:"previous_checkpoint_id,omitempty"`
}
