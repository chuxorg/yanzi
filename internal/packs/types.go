package packs

import (
	"math"
	"strings"
	"time"
)

// Seed type constants.
const (
	SeedTypeYanzi       = "yanzi"
	SeedTypeProcess     = "process"
	SeedTypeGuardrail   = "guardrail"
	SeedTypeSkill       = "skill"
	SeedTypePersonality = "personality"
)

// ContentSection is a single named section of seed content.
type ContentSection struct {
	Section string `json:"section" yaml:"section"` // e.g. "overview", "constraints"
	Type    string `json:"type" yaml:"type"`       // instruction | guardrail | process | skill | example
	Text    string `json:"text" yaml:"text"`
}

// SeedContent holds the structured sections of a Seed.
type SeedContent struct {
	Sections []ContentSection `json:"sections" yaml:"sections"`
}

// Assemble concatenates all sections in order with a section header before each.
func (sc SeedContent) Assemble() string {
	var b strings.Builder
	for _, s := range sc.Sections {
		b.WriteString("## ")
		b.WriteString(strings.ToUpper(s.Section))
		b.WriteString("\n")
		b.WriteString(s.Text)
		if !strings.HasSuffix(s.Text, "\n") {
			b.WriteString("\n")
		}
	}
	return b.String()
}

// TokenEstimate returns an approximate token count for the assembled content.
// Word count × 1.3, rounded up.
func (sc SeedContent) TokenEstimate() int {
	words := len(strings.Fields(sc.Assemble()))
	return int(math.Ceil(float64(words) * 1.3))
}

// Seed is a discrete reusable context unit.
type Seed struct {
	ID             string      `json:"id" yaml:"id"`
	Name           string      `json:"name" yaml:"name"`
	VersionLabel   string      `json:"version_label,omitempty" yaml:"version_label,omitempty"`
	SeedType       string      `json:"seed_type" yaml:"seed_type"`
	RoleAccessBits RoleBits    `json:"role_access_bits" yaml:"role_access_bits"`
	Description    string      `json:"description,omitempty" yaml:"description,omitempty"`
	Content        SeedContent `json:"content" yaml:"content"`
	TokenEstimate  int         `json:"token_estimate" yaml:"token_estimate"`
	Tags           []string    `json:"tags,omitempty" yaml:"tags,omitempty"`
	AuthorRole     string      `json:"author_role,omitempty" yaml:"author_role,omitempty"`
	CreatedAt      time.Time   `json:"created_at" yaml:"created_at"`
	ArtifactID     string      `json:"artifact_id" yaml:"artifact_id"`
}

// SeedReference is a lightweight pointer to a seed used within a Pack.
type SeedReference struct {
	Name       string `json:"name" yaml:"name"`             // seed name for human readability
	ArtifactID string `json:"artifact_id" yaml:"artifact_id"` // reliable reference
	Override   bool   `json:"override,omitempty" yaml:"override,omitempty"` // true if overrides a parent pack's seed
}

// PackTokenEstimate breaks down token counts for a Pack.
type PackTokenEstimate struct {
	PackContext int  `json:"pack_context" yaml:"pack_context"`
	SeedsTotal  int  `json:"seeds_total" yaml:"seeds_total"`
	Total       int  `json:"total" yaml:"total"`
	Approximate bool `json:"approximate" yaml:"approximate"`
}

// Pack is a named composition of Seeds for a specific agent role.
type Pack struct {
	ID            string            `json:"id" yaml:"id"`
	Name          string            `json:"name" yaml:"name"`
	VersionLabel  string            `json:"version_label,omitempty" yaml:"version_label,omitempty"`
	ExtendsID     string            `json:"extends_id,omitempty" yaml:"extends_id,omitempty"` // artifact ID of parent pack
	Role          RoleBits          `json:"role" yaml:"role"`
	RoleLabel     string            `json:"role_label,omitempty" yaml:"role_label,omitempty"` // user-defined label
	Description   string            `json:"description,omitempty" yaml:"description,omitempty"`
	PackContext   string            `json:"pack_context,omitempty" yaml:"pack_context,omitempty"`
	Seeds         []SeedReference   `json:"seeds" yaml:"seeds"`
	TokenEstimate PackTokenEstimate `json:"token_estimate" yaml:"token_estimate"`
	Tags          []string          `json:"tags,omitempty" yaml:"tags,omitempty"`
	AuthorRole    string            `json:"author_role,omitempty" yaml:"author_role,omitempty"`
	CreatedAt     time.Time         `json:"created_at" yaml:"created_at"`
	ArtifactID    string            `json:"artifact_id" yaml:"artifact_id"`
}

// ComposedSection is one section in a composed prompt result.
type ComposedSection struct {
	Type             string `json:"type"`               // "pack_context" | "seed" | "inherited_seed" | "task"
	SourceArtifactID string `json:"source_artifact_id"`
	SeedName         string `json:"seed_name,omitempty"`
	SeedType         string `json:"seed_type,omitempty"`
	AuthorRole       string `json:"author_role,omitempty"`
	Content          string `json:"content"`
	TokenEstimate    int    `json:"token_estimate"`
}

// ComposeTokenEstimate breaks down token counts for a composed result.
type ComposeTokenEstimate struct {
	PackContext int    `json:"pack_context"`
	SeedsTotal  int    `json:"seeds_total"`
	Task        int    `json:"task"`
	Total       int    `json:"total"`
	Approximate bool   `json:"approximate"`
	ModelHint   string `json:"model_hint,omitempty"`
}

// ComposeWarning is a non-fatal advisory issued during composition.
type ComposeWarning struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	Severity string `json:"severity"` // "advisory" | "warning" | "error"
}

// ComposeResult is the output of Pack composition.
type ComposeResult struct {
	Pack            Pack                 `json:"pack"`
	Sections        []ComposedSection    `json:"sections,omitempty"`
	AssembledPrompt string               `json:"assembled_prompt,omitempty"`
	ClipboardString string               `json:"clipboard_string,omitempty"`
	TokenEstimate   ComposeTokenEstimate `json:"token_estimate"`
	Warnings        []ComposeWarning     `json:"warnings"`
}

// ComposeRequest is the input to Pack composition.
type ComposeRequest struct {
	PackArtifactID string         `json:"pack_artifact_id"`
	TaskContent    string         `json:"task_content,omitempty"`
	TaskArtifactID string         `json:"task_artifact_id,omitempty"`
	ModelHint      string         `json:"model_hint,omitempty"`
	Options        ComposeOptions `json:"options,omitempty"`
}

// ComposeOptions controls which fields are populated in ComposeResult.
type ComposeOptions struct {
	IncludeSections        bool `json:"include_sections"`
	IncludeAssembledPrompt bool `json:"include_assembled_prompt"`
	IncludeClipboardString bool `json:"include_clipboard_string"`
}
