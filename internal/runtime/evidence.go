package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/entroforge/go-system-builder/internal/evidence"
)

// EvidenceRequest describes one current, fingerprinted evidence artifact to
// append to Runtime. Evidence is deliberately recorded through Store.Update
// so it participates in the same revision CAS and journal contract as every
// other Runtime mutation.
type EvidenceRequest struct {
	ExpectedRevision int
	ID               string
	Kind             string
	Path             string
	ProducedBy       []string
	ResponsibilityID string
	ReviewRound      *int
	ScopeRefs        []string
	OccurredAt       time.Time
	Validator        CandidateValidator
}

// IsRegisteredEvidenceKind reports whether kind can be persisted by
// RecordEvidence using the shared evidence catalog.
func IsRegisteredEvidenceKind(kind string) bool {
	return evidence.DefaultCatalog().IsRegisteredKind(kind)
}

// RecordEvidence adds one valid evidence item and commits it as a Runtime
// mutation. It rejects absolute or escaping paths, missing artifacts, invalid
// kinds, empty producers, duplicate IDs, and stale revisions before commit.
func RecordEvidence(root, statePath, journalPath string, request EvidenceRequest) (Snapshot, error) {
	if request.ID == "" {
		return Snapshot{}, fmt.Errorf("evidence id is required")
	}
	catalog := evidence.DefaultCatalog()
	if !catalog.IsRegisteredKind(request.Kind) {
		return Snapshot{}, fmt.Errorf("unsupported evidence kind %q; registered kinds: %s", request.Kind, strings.Join(catalog.RegisteredKinds(), ", "))
	}
	// finding_supplement is pipeline-owned: review.SubmitSupplement persists
	// the entities.finding_supplements index row and the evidence entry in one
	// runtime CAS transaction. A manual add would register an evidence row
	// with no supplement index row — a half-registered supplement — so this
	// entry point fails closed with the authoritative path.
	if strings.TrimSpace(request.Kind) == "finding_supplement" {
		return Snapshot{}, fmt.Errorf("evidence kind %q is pipeline-owned: persist supplements with `runtime finding-supplement` (appends the entities.finding_supplements row and the finding_supplement evidence entry in one CAS transaction); manual evidence registration would split the supplement index from the evidence log", request.Kind)
	}
	if len(request.ProducedBy) == 0 {
		return Snapshot{}, fmt.Errorf("evidence produced_by is required")
	}
	for _, producer := range request.ProducedBy {
		if strings.TrimSpace(producer) == "" {
			return Snapshot{}, fmt.Errorf("evidence produced_by contains an empty actor")
		}
	}
	cleanPath, err := safeEvidencePath(root, request.Path)
	if err != nil {
		return Snapshot{}, err
	}
	data, err := os.ReadFile(filepath.Join(root, cleanPath))
	if err != nil {
		return Snapshot{}, fmt.Errorf("read evidence artifact: %w", err)
	}

	store := NewWriter(statePath, journalPath, root, request.Validator)
	snapshot, err := store.Snapshot()
	if err != nil {
		return Snapshot{}, fmt.Errorf("read runtime: %w", err)
	}
	current := snapshot.State
	runtimeID, _ := current["runtime_id"].(string)
	lifecycle, _ := current["lifecycle"].(map[string]any)
	from := map[string]any{"state": lifecycle["state"], "phase": lifecycle["phase"]}

	occurredAt := request.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	producedBy := append([]string(nil), request.ProducedBy...)
	scopeRefs, err := expandRolloverScopeRefs(request.ScopeRefs, runtimeID, request.ExpectedRevision+1)
	if err != nil {
		return Snapshot{}, err
	}
	sha := sha256Hex(data)
	return store.Update(request.ExpectedRevision, Mutation{
		EventID:        fmt.Sprintf("evt-evidence-%s-r%d", request.ID, request.ExpectedRevision+1),
		TransitionID:   "EVIDENCE-RECORD",
		Event:          "evidence_recorded",
		Actor:          "orchestrator",
		IdempotencyKey: fmt.Sprintf("runtime:evidence:%s:%d", request.ID, request.ExpectedRevision),
		RuntimeID:      runtimeID,
		From:           from,
		To:             from,
		EvidenceIDs:    []string{request.ID},
		Message:        "Recorded a current fingerprinted evidence artifact.",
		OccurredAt:     occurredAt,
		Apply: func(state map[string]any) error {
			items, ok := state["evidence"].([]any)
			if !ok {
				return fmt.Errorf("runtime evidence must be an array")
			}
			for _, raw := range items {
				item, _ := raw.(map[string]any)
				if item != nil && item["id"] == request.ID {
					return fmt.Errorf("evidence %s is already registered", request.ID)
				}
			}
			baseline, ok := state["baseline"].(map[string]any)
			if !ok {
				return fmt.Errorf("runtime baseline must be an object")
			}
			generation, err := integerField(baseline, "generation")
			if err != nil {
				return err
			}
			var reviewRound any
			if request.ReviewRound != nil {
				if *request.ReviewRound < 1 {
					return fmt.Errorf("review round must be at least 1")
				}
				reviewRound = *request.ReviewRound
			}
			items = append(items, map[string]any{
				"id":                  request.ID,
				"kind":                request.Kind,
				"path":                cleanPath,
				"sha256":              sha,
				"status":              "valid",
				"baseline_generation": generation,
				"review_round":        reviewRound,
				"produced_by":         producedBy,
				"invalidated_by":      nil,
				"invalidation_rule":   nil,
				"invalidation_reason": nil,
				"responsibility_id":   nullableString(request.ResponsibilityID),
				"scope_refs":          scopeRefs,
			})
			state["evidence"] = items
			state["updated_at"] = occurredAt.UTC().Format(time.RFC3339Nano)
			return nil
		},
	})
}

// expandRolloverScopeRefs resolves the explicit `runtime_rollover:current`
// token to the revision that this evidence commit will produce. This lets a
// human approval evidence item bind to its terminal runtime without manually
// predicting Store.Update's revision increment.
func expandRolloverScopeRefs(scopeRefs []string, runtimeID string, committedRevision int) ([]string, error) {
	if runtimeID == "" {
		return nil, fmt.Errorf("runtime id is required for rollover scope")
	}
	result := append([]string{}, scopeRefs...)
	for index, scope := range result {
		if scope == "runtime_rollover:current" {
			result[index] = fmt.Sprintf("runtime_rollover:%s@%d", runtimeID, committedRevision)
		}
	}
	return result, nil
}

func safeEvidencePath(root, path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("evidence path is required")
	}
	clean := filepath.Clean(path)
	if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("evidence path must stay within repository: %q", path)
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve repository root: %w", err)
	}
	resolvedRoot, err := filepath.EvalSymlinks(rootAbs)
	if err != nil {
		return "", fmt.Errorf("resolve repository root symlinks: %w", err)
	}
	resolvedPath, err := filepath.EvalSymlinks(filepath.Join(rootAbs, clean))
	if err != nil {
		return "", fmt.Errorf("resolve evidence path symlinks: %w", err)
	}
	relative, err := filepath.Rel(resolvedRoot, resolvedPath)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("evidence path must stay within repository: %q", path)
	}
	return filepath.ToSlash(clean), nil
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
