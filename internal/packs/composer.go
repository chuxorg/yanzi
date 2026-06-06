package packs

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

const maxInheritanceDepth = 10

// injection patterns scanned in seed content at compose time (case-insensitive).
var injectionPatterns = []string{
	"ignore all previous instructions",
	"ignore previous instructions",
	"disregard the above",
	"forget everything",
	"you are now",
	"new persona",
}

// Composer assembles a Pack and its Seeds into a composed prompt.
type Composer struct {
	store PackStore
}

// NewComposer returns a Composer backed by store.
func NewComposer(store PackStore) *Composer {
	return &Composer{store: store}
}

// Compose resolves the pack inheritance chain, loads seeds, validates access,
// and assembles the composed prompt result.
func (c *Composer) Compose(ctx context.Context, req ComposeRequest) (ComposeResult, error) {
	if req.PackArtifactID == "" {
		return ComposeResult{}, errors.New("pack_artifact_id is required")
	}

	// Step 1 — Resolve inheritance chain.
	chain, err := c.resolveChain(ctx, req.PackArtifactID, nil)
	if err != nil {
		return ComposeResult{}, err
	}
	root := chain[0]
	leaf := chain[len(chain)-1]

	var warnings []ComposeWarning

	// Step 2 — Resolve seed list with inheritance and overrides.
	resolvedRefs := resolveSeedList(chain)

	// Step 3 — Load Seeds.
	seedMap := make(map[string]Seed, len(resolvedRefs))
	for _, ref := range resolvedRefs {
		seed, err := c.store.GetSeed(ctx, ref.ArtifactID)
		if err != nil {
			warnings = append(warnings, ComposeWarning{
				Code:     "missing_seed",
				Message:  fmt.Sprintf("seed %q (%s) not found: %v", ref.Name, ref.ArtifactID, err),
				Severity: "warning",
			})
			continue
		}
		seedMap[ref.ArtifactID] = seed
	}

	// Step 4 — Validate role access (advisory only).
	for _, ref := range resolvedRefs {
		seed, ok := seedMap[ref.ArtifactID]
		if !ok {
			continue
		}
		if !leaf.Role.Includes(seed.RoleAccessBits) {
			warnings = append(warnings, ComposeWarning{
				Code:     "role_access_violation",
				Message:  fmt.Sprintf("seed %q requires role bits %d; pack has %d", seed.Name, seed.RoleAccessBits, leaf.Role),
				Severity: "advisory",
			})
		}
	}

	// Step 5 — Scan for injection patterns.
	for _, ref := range resolvedRefs {
		seed, ok := seedMap[ref.ArtifactID]
		if !ok {
			continue
		}
		assembled := seed.Content.Assemble()
		lower := strings.ToLower(assembled)
		for _, pattern := range injectionPatterns {
			if strings.Contains(lower, pattern) {
				warnings = append(warnings, ComposeWarning{
					Code:     "injection_pattern",
					Message:  fmt.Sprintf("seed %q contains suspicious pattern: %q", seed.Name, pattern),
					Severity: "warning",
				})
				break
			}
		}
	}

	// Step 6 — Check for yanzi seed.
	hasYanzi := false
	for _, ref := range resolvedRefs {
		if seed, ok := seedMap[ref.ArtifactID]; ok && seed.SeedType == SeedTypeYanzi {
			hasYanzi = true
			break
		}
	}
	if !hasYanzi {
		warnings = append(warnings, ComposeWarning{
			Code:     "missing_yanzi_seed",
			Message:  "no yanzi-type seed found in pack — recommend adding a yanzi seed",
			Severity: "advisory",
		})
	}

	// Step 7 — Build sections.
	var sections []ComposedSection

	// Pack context from innermost (leaf) pack.
	if leaf.PackContext != "" {
		est := tokenEstimateFromText(leaf.PackContext)
		sections = append(sections, ComposedSection{
			Type:             "pack_context",
			SourceArtifactID: leaf.ArtifactID,
			AuthorRole:       leaf.AuthorRole,
			Content:          leaf.PackContext,
			TokenEstimate:    est,
		})
	} else if root.PackContext != "" {
		est := tokenEstimateFromText(root.PackContext)
		sections = append(sections, ComposedSection{
			Type:             "pack_context",
			SourceArtifactID: root.ArtifactID,
			AuthorRole:       root.AuthorRole,
			Content:          root.PackContext,
			TokenEstimate:    est,
		})
	}

	for _, ref := range resolvedRefs {
		seed, ok := seedMap[ref.ArtifactID]
		if !ok {
			continue
		}
		sectionType := "seed"
		if ref.Override {
			sectionType = "inherited_seed"
		}
		content := seed.Content.Assemble()
		sections = append(sections, ComposedSection{
			Type:             sectionType,
			SourceArtifactID: seed.ArtifactID,
			SeedName:         seed.Name,
			SeedType:         seed.SeedType,
			AuthorRole:       seed.AuthorRole,
			Content:          content,
			TokenEstimate:    seed.TokenEstimate,
		})
	}

	var taskEst int
	if req.TaskContent != "" {
		taskEst = tokenEstimateFromText(req.TaskContent)
		sections = append(sections, ComposedSection{
			Type:          "task",
			Content:       req.TaskContent,
			TokenEstimate: taskEst,
		})
	}

	// Step 8 — Build assembled prompt with trust boundary markers.
	var assembled string
	if req.Options.IncludeAssembledPrompt {
		assembled = buildAssembledPrompt(sections, req.TaskContent)
	}

	// Step 9 — Build clipboard string.
	var clipboard string
	if req.Options.IncludeClipboardString {
		te := sumTokenEstimates(sections)
		clipboard = buildClipboardString(leaf, te, assembled)
	}

	// Step 10 — Token estimates.
	var packContextEst, seedsEst int
	for _, s := range sections {
		switch s.Type {
		case "pack_context":
			packContextEst += s.TokenEstimate
		case "seed", "inherited_seed":
			seedsEst += s.TokenEstimate
		}
	}
	total := packContextEst + seedsEst + taskEst

	tokenEstimate := ComposeTokenEstimate{
		PackContext: packContextEst,
		SeedsTotal:  seedsEst,
		Task:        taskEst,
		Total:       total,
		Approximate: true,
		ModelHint:   req.ModelHint,
	}

	result := ComposeResult{
		Pack:          leaf,
		TokenEstimate: tokenEstimate,
		Warnings:      warnings,
	}
	if req.Options.IncludeSections {
		result.Sections = sections
	}
	if req.Options.IncludeAssembledPrompt {
		result.AssembledPrompt = assembled
	}
	if req.Options.IncludeClipboardString {
		result.ClipboardString = clipboard
	}

	return result, nil
}

// resolveChain loads the pack inheritance chain from root to leaf.
func (c *Composer) resolveChain(ctx context.Context, artifactID string, visited []string) ([]Pack, error) {
	if len(visited) > maxInheritanceDepth {
		return nil, fmt.Errorf("pack inheritance depth exceeds %d levels", maxInheritanceDepth)
	}
	for _, v := range visited {
		if v == artifactID {
			return nil, fmt.Errorf("circular pack inheritance detected at %s", artifactID)
		}
	}

	pack, err := c.store.GetPack(ctx, artifactID)
	if err != nil {
		return nil, fmt.Errorf("load pack %s: %w", artifactID, err)
	}

	visited = append(visited, artifactID)

	if pack.ExtendsID == "" {
		return []Pack{pack}, nil
	}

	parentChain, err := c.resolveChain(ctx, pack.ExtendsID, visited)
	if err != nil {
		return nil, err
	}

	return append(parentChain, pack), nil
}

// resolveSeedList merges seed lists across the inheritance chain.
// Root seeds come first; child seeds append or replace by name.
func resolveSeedList(chain []Pack) []SeedReference {
	var resolved []SeedReference
	for _, pack := range chain {
		for _, ref := range pack.Seeds {
			replaced := false
			for i, existing := range resolved {
				if existing.Name == ref.Name {
					r := ref
					r.Override = true
					resolved[i] = r
					replaced = true
					break
				}
			}
			if !replaced {
				resolved = append(resolved, ref)
			}
		}
	}
	return resolved
}

func buildAssembledPrompt(sections []ComposedSection, taskContent string) string {
	var b strings.Builder
	b.WriteString("=== SYSTEM CONTEXT (trusted) ===\n")
	for _, s := range sections {
		if s.Type == "task" {
			continue
		}
		b.WriteString(s.Content)
		if !strings.HasSuffix(s.Content, "\n") {
			b.WriteString("\n")
		}
	}
	if taskContent != "" {
		b.WriteString("\n=== TASK ===\n")
		b.WriteString(taskContent)
		if !strings.HasSuffix(taskContent, "\n") {
			b.WriteString("\n")
		}
		b.WriteString("=== END TASK ===")
	}
	return b.String()
}

func buildClipboardString(pack Pack, tokenTotal int, assembled string) string {
	var b strings.Builder
	b.WriteString("# Yanzi Composed Prompt\n")
	b.WriteString(fmt.Sprintf("# Pack: %s | Role: %s\n", pack.Name, pack.RoleLabel))
	b.WriteString(fmt.Sprintf("# Tokens (approx): %d\n", tokenTotal))
	b.WriteString(fmt.Sprintf("# Generated: %s\n\n", time.Now().UTC().Format(time.RFC3339)))
	b.WriteString(assembled)
	return b.String()
}

func sumTokenEstimates(sections []ComposedSection) int {
	total := 0
	for _, s := range sections {
		total += s.TokenEstimate
	}
	return total
}

func tokenEstimateFromText(text string) int {
	words := len(strings.Fields(text))
	return int(math.Ceil(float64(words) * 1.3))
}
