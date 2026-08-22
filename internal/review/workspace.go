package review

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// ---------------------------------------------------------------------------
// E2E cold-start verification workspace (L3-S7 §1.4.1, §3.2): the workspace
// is the only write surface reviewers may use during S7; its digest is
// pinned at registration and re-verified at seal/clean so a spec that changed
// after a result was consumed invalidates the round honestly.
// ---------------------------------------------------------------------------

// WorkspaceDigest computes sha256 over the sorted "relpath:filesha256" lines
// of every file under dir. An empty or missing workspace digests to the
// empty-string hash (the registered cold-start baseline).
func WorkspaceDigest(root, workspaceRel string) (string, error) {
	if workspaceRel == "" {
		return "", nil
	}
	abs := filepath.Join(root, filepath.FromSlash(workspaceRel))
	entries, err := os.ReadDir(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return sha256Of([]byte("")), nil
		}
		return "", fmt.Errorf("read workspace %s: %w", workspaceRel, err)
	}
	_ = entries
	var lines []string
	err = filepath.Walk(abs, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(abs, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		lines = append(lines, filepath.ToSlash(rel)+":"+sha256Of(data))
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(lines)
	return sha256Of([]byte(strings.Join(lines, "\n"))), nil
}

// prepareVerificationWorkspace validates and creates the E2E cold-start
// workspace: it must live under e2e-workspace/ inside the repository so the
// reviewer write-scope rule and the digest recomputation can see it.
func prepareVerificationWorkspace(root string, plan *Plan) (string, error) {
	if plan.VerificationArtifactWorkspace == nil || strings.TrimSpace(*plan.VerificationArtifactWorkspace) == "" {
		return "", nil
	}
	workspace := *plan.VerificationArtifactWorkspace
	if filepath.IsAbs(workspace) || strings.HasPrefix(workspace, "..") {
		return "", fmt.Errorf("verification_artifact_workspace must be repository-relative (got %q)", workspace)
	}
	if !strings.HasPrefix(workspace, "e2e-workspace/") {
		return "", fmt.Errorf("verification_artifact_workspace must live under e2e-workspace/ (got %q); the reviewer write-scope rule only knows that surface", workspace)
	}
	abs := filepath.Join(root, filepath.FromSlash(workspace))
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return "", fmt.Errorf("create verification workspace: %w", err)
	}
	digest, err := WorkspaceDigest(root, workspace)
	if err != nil {
		return "", err
	}
	return digest, nil
}

// verifyResultArtifactDigest binds an E2E result to the workspace digest it
// actually ran against (L3-S7 §3.5): results from a workspace that has since
// drifted are stale, not submittable.
func verifyResultArtifactDigest(root string, plan *Plan, ptr *PlanPointer, result *Result, lens string) error {
	if lens != "e2e" {
		return nil
	}
	if ptr.VerificationArtifactWorkspace == "" {
		if result.VerificationArtifactDigest != nil && *result.VerificationArtifactDigest != "" {
			return fmt.Errorf("result binds a verification artifact digest but the plan declares no workspace")
		}
		return nil
	}
	digest, err := WorkspaceDigest(root, ptr.VerificationArtifactWorkspace)
	if err != nil {
		return err
	}
	if result.VerificationArtifactDigest == nil || *result.VerificationArtifactDigest == "" {
		return fmt.Errorf("E2E results on a cold-start workspace must bind verification_artifact_digest; compute it with `loop-harness s7 workspace-digest` after the spec/fixture write")
	}
	if *result.VerificationArtifactDigest != digest {
		return fmt.Errorf("verification artifact digest mismatch: result binds %s but the workspace now digests to %s; re-run the flows against the current spec before submitting", *result.VerificationArtifactDigest, digest)
	}
	return nil
}

// verifySealedArtifactDigests re-checks every consumed E2E assignment's
// bound digest against the workspace at close time (L3-S7 §10.1.6): a
// workspace that drifted after consumption invalidates the round.
func verifySealedArtifactDigests(root string, ptr *PlanPointer, assignments map[string]any) error {
	if ptr.VerificationArtifactWorkspace == "" {
		return nil
	}
	current, err := WorkspaceDigest(root, ptr.VerificationArtifactWorkspace)
	if err != nil {
		return err
	}
	for id, raw := range assignments {
		row, _ := raw.(map[string]any)
		if row == nil || row["lens"] != "e2e" || row["status"] != "consumed" {
			continue
		}
		bound, _ := row["artifact_digest"].(string)
		if bound != "" && bound != current {
			return fmt.Errorf("assignment %s consumed a result against workspace digest %s, but the workspace now digests to %s; the round is stale (L3-S7 §10.3)", id, bound, current)
		}
	}
	return nil
}

// projectedAssignments returns the assignment projection as it will look
// after this result is consumed (current assignment consumed + digest bound).
func projectedAssignments(state map[string]any, currentAssignment *PlanAssignment, result *Result) map[string]any {
	reviewMap, _ := state["review"].(map[string]any)
	assignments, _ := reviewMap["assignments"].(map[string]any)
	out := map[string]any{}
	for id, raw := range assignments {
		row, _ := raw.(map[string]any)
		if row != nil {
			out[id] = row
		}
	}
	digest := ""
	if result.VerificationArtifactDigest != nil {
		digest = *result.VerificationArtifactDigest
	}
	out[currentAssignment.AssignmentID] = map[string]any{
		"lens": currentAssignment.Lens, "status": "consumed",
		"agent_id": result.ProducerAgentID, "artifact_digest": digest,
	}
	return out
}

// ---------------------------------------------------------------------------
// Capture buffer (L3-S7 §3.6 auto-capture): `loop-harness capture step`
// appends one sanitized timeline step per call; submit merges them into
// findings whose encounter timeline is empty.
// ---------------------------------------------------------------------------

// secretPatterns are the redaction gate: any capture field matching one is
// rejected — the buffer never persists secrets (L3-S7 §6.3).
var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)password\s*[:=]`),
	regexp.MustCompile(`(?i)api[_-]?key\s*[:=]`),
	regexp.MustCompile(`(?i)secret\s*[:=]`),
	regexp.MustCompile(`(?i)bearer\s+[A-Za-z0-9._-]{12,}`),
	regexp.MustCompile(`(?i)token\s*[:=]\s*[A-Za-z0-9._-]{8,}`),
}

// CaptureStep is one sanitized timeline step.
type CaptureStep struct {
	Sequence   int      `json:"sequence"`
	Action     string   `json:"action"`
	Observed   string   `json:"observed"`
	Evidence   []string `json:"evidence_refs,omitempty"`
	CapturedAt string   `json:"captured_at"`
}

// SanitizeCapture rejects any field that smells like a secret.
func SanitizeCapture(step CaptureStep) error {
	fields := map[string]string{
		"action":   step.Action,
		"observed": step.Observed,
	}
	for _, ref := range step.Evidence {
		fields["evidence_ref"] = ref
	}
	for name, value := range fields {
		for _, pattern := range secretPatterns {
			if pattern.MatchString(value) {
				return fmt.Errorf("capture field %q matches a secret pattern (%s); record the redacted ref, not the value (L3-S7 §6.3)", name, pattern.String())
			}
		}
	}
	return nil
}

// CaptureFile returns the buffer path for one assignment.
func CaptureFile(root, runtimeID string, generation int, assignmentID string) string {
	return filepath.Join(root, ".claude", "evidence", runtimeID,
		fmt.Sprintf("g%d", generation), "captures", assignmentID, "steps.jsonl")
}

// LoadCaptureSteps reads the buffer; a missing buffer is empty, not an error.
func LoadCaptureSteps(path string) []CaptureStep {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var steps []CaptureStep
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var step CaptureStep
		if err := json.Unmarshal([]byte(line), &step); err == nil {
			steps = append(steps, step)
		}
	}
	return steps
}

// mergeCapturedTimeline fills an empty encounter timeline from the buffer.
// Findings whose Reviewer wrote a real timeline are never rewritten.
func mergeCapturedTimeline(findings []Finding, steps []CaptureStep) {
	if len(steps) == 0 {
		return
	}
	timeline := make([]TimelineStep, 0, len(steps))
	for _, step := range steps {
		timeline = append(timeline, TimelineStep{
			Sequence:           step.Sequence,
			Action:             step.Action,
			ObservedCheckpoint: step.Observed,
			EvidenceRefs:       step.Evidence,
		})
	}
	for i := range findings {
		if len(findings[i].Encounter.Timeline) == 0 {
			findings[i].Encounter.Timeline = timeline
		}
	}
}
