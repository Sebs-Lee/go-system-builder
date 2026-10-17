package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/entroforge/go-system-builder/internal/acceptance"
	"github.com/entroforge/go-system-builder/internal/policy"
	"github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/semantic"
)

var auditPathSegment = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

func currentAuditKind(state map[string]any) string {
	lifecycle, _ := state["lifecycle"].(map[string]any)
	switch stringValue(lifecycle["state"]) {
	case acceptance.ManifestAcceptance:
		return acceptance.ManifestAcceptance
	case acceptance.ManifestReleaseAudit:
		return acceptance.ManifestReleaseAudit
	default:
		return ""
	}
}

func stageAuditQuestionnaireContext(root string, request policy.Input, snapshot runtime.Snapshot) string {
	if !mainQuestionnaireViewer(request) {
		return ""
	}
	kind := currentAuditKind(snapshot.State)
	if kind == "" {
		return ""
	}
	status := inspectS10Artifact(root, snapshot.State, kind)
	if status.State == "ready" || status.State == "review_required" || status.State == "blocked" {
		path, err := absoluteQuestionnairePath(root, status.ManifestPath)
		if err != nil {
			return fmt.Sprintf("S10 %s evidence is registered, but its manifest path cannot be resolved: %v", kind, err)
		}
		return fmt.Sprintf("S10 %s questionnaire is already saved at %s; the original evidence and Controller path will evaluate it. Wait for Main's next stage instruction.", kind, path)
	}
	path, err := auditQuestionnairePath(root, snapshot.State, kind)
	if err != nil {
		return fmt.Sprintf("S10 %s is active, but the system could not prepare its exact manifest path from the current Runtime: %v. This is a wiring/source gap; do not guess a path or reconstruct the inventory from other Runtime files.", kind, err)
	}
	absPath, err := absoluteQuestionnairePath(root, path)
	if err != nil {
		return fmt.Sprintf("S10 %s is active, but its exact manifest path cannot be resolved: %v", kind, err)
	}
	if questionnairePreparationEvent(request.Event) {
		if _, _, err := prepareAuditQuestionnaire(root, snapshot.State, kind); err != nil {
			return fmt.Sprintf("S10 %s is active, but the system could not prepare its exact manifest path %s from the current Runtime: %v. Do not guess a path or reconstruct the inventory from other Runtime files.", kind, absPath, err)
		}
		return fmt.Sprintf("S10 %s manifest answer: %s. The system filled current Runtime/generation/review round and authoritative inventory IDs. Fill the professional expected/oracle/owner, disposition, counterevidence, evidence, risks, and outcome fields in this file, then save it. For release audit, answer the eight audit_area coverage rows once; the listener derives the original audit_areas summary and metrics. The existing evidence/Controller consumer runs after a valid save; do not author an envelope or a second report.", kind, absPath)
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(path))); os.IsNotExist(err) {
		return fmt.Sprintf("S10 %s answer draft is not prepared yet; the exact path is %s. This Hook event is read-only. SessionStart or PreToolUse prepares it; Stop/Idle only reports the missing draft.", kind, absPath)
	} else if err != nil {
		return fmt.Sprintf("S10 %s answer draft at %s could not be inspected: %v", kind, absPath, err)
	}
	return fmt.Sprintf("S10 %s manifest answer is at %s. This Hook event is read-only; review or continue filling this existing draft, then save it. The original evidence/Controller consumer runs after a valid save.", kind, absPath)
}

func questionnairePreparationEvent(event string) bool {
	return event == "SessionStart" || event == "PreToolUse"
}

func absoluteQuestionnairePath(root, relative string) (string, error) {
	if strings.TrimSpace(relative) == "" {
		return "", fmt.Errorf("the registered answer path is empty")
	}
	return filepath.Abs(filepath.Join(root, filepath.FromSlash(relative)))
}

func prepareAuditQuestionnaire(root string, state map[string]any, kind string) (string, acceptance.Manifest, error) {
	status := inspectS10Artifact(root, state, kind)
	if status.State == "ready" || status.State == "review_required" || status.State == "blocked" {
		return status.ManifestPath, acceptance.Manifest{}, nil
	}
	path, err := auditQuestionnairePath(root, state, kind)
	if err != nil {
		return "", acceptance.Manifest{}, err
	}
	draftPath, err := absoluteQuestionnairePath(root, path)
	if err != nil {
		return "", acceptance.Manifest{}, err
	}
	if _, err := os.Stat(draftPath); err == nil {
		return path, acceptance.Manifest{}, nil
	} else if !os.IsNotExist(err) {
		return "", acceptance.Manifest{}, fmt.Errorf("inspect existing manifest %s: %w", draftPath, err)
	}
	baseline, err := acceptance.BuildS10ExternalBaseline(root, state, nil)
	if err != nil {
		return "", acceptance.Manifest{}, fmt.Errorf("build current changed-path inventory: %w", err)
	}
	authority, err := acceptance.BuildS10InventoryAuthority(root, state, baseline)
	if err != nil {
		return "", acceptance.Manifest{}, fmt.Errorf("build current authoritative inventory: %w", err)
	}
	manifest := auditManifestDraft(state, kind, authority)
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return "", acceptance.Manifest{}, fmt.Errorf("encode current manifest draft: %w", err)
	}
	if err := prepareQuestionnaireText(draftPath, string(data)+"\n"); err != nil {
		return "", acceptance.Manifest{}, fmt.Errorf("prepare current manifest draft %s: %w", draftPath, err)
	}
	return path, manifest, nil
}

func auditManifestDraft(state map[string]any, kind string, authority acceptance.InventoryAuthority) acceptance.Manifest {
	manifest := acceptance.Manifest{
		SchemaVersion:      acceptance.SchemaVersionAcceptance,
		ManifestType:       kind,
		RuntimeID:          stringValue(state["runtime_id"]),
		BaselineGeneration: integerValue(nestedStateValue(state, "baseline", "generation")),
		ReviewRound:        integerValue(nestedStateValue(state, "review", "round")),
		CoverageInventory:  make([]acceptance.CoverageItem, 0),
		Counterevidence:    make([]acceptance.CounterevidenceItem, 0),
		Risks:              []acceptance.RiskItem{},
		TechnicalDebt:      []acceptance.DebtItem{},
		BlockingFindings:   []acceptance.BlockingFinding{},
	}
	add := func(category string, ids []string) {
		for _, id := range ids {
			source := id
			if category == "changed_path" && !strings.HasPrefix(id, "path:") {
				id = "path:" + id
			}
			manifest.CoverageInventory = append(manifest.CoverageInventory, acceptance.CoverageItem{
				ID: id, Category: category, SourceRefs: []string{source},
				EvidenceRefs: []string{},
			})
		}
	}
	add("requirement", authority.RequirementIDs)
	add("contract", authority.ContractIDs)
	add("task", authority.TaskIDs)
	add("claim", authority.ClaimIDs)
	add("changed_path", authority.ChangedPaths)
	if kind == acceptance.ManifestReleaseAudit {
		add("audit_area", acceptance.AuditAreaIDs)
	}
	for _, item := range manifest.CoverageInventory {
		manifest.Counterevidence = append(manifest.Counterevidence, acceptance.CounterevidenceItem{
			ID: "counter:" + item.ID, InventoryID: item.ID, EvidenceRefs: []string{},
		})
	}
	manifest.Metrics = acceptance.DeriveMetrics(manifest)
	return manifest
}

// The original manifest carries an audit-area summary. New questionnaires
// collect that judgment in the coverage rows once and project the summary on
// save. Previously authored summaries remain subject to the original validator.
func projectAuditAreaSummary(manifest *acceptance.Manifest) {
	if manifest.ManifestType != acceptance.ManifestReleaseAudit || len(manifest.AuditAreas) != 0 {
		return
	}
	for _, item := range manifest.CoverageInventory {
		if item.Category == "audit_area" {
			manifest.AuditAreas = append(manifest.AuditAreas, acceptance.AuditArea{
				ID: item.ID, Conclusion: item.Disposition, Owner: item.Owner,
				EvidenceRefs: append([]string(nil), item.EvidenceRefs...),
			})
		}
	}
}

func auditQuestionnairePath(root string, state map[string]any, kind string) (string, error) {
	runtimeID := strings.TrimSpace(stringValue(state["runtime_id"]))
	if runtimeID == "" {
		return "", fmt.Errorf("current Runtime has no runtime_id")
	}
	generation := integerValue(nestedStateValue(state, "baseline", "generation"))
	round := integerValue(nestedStateValue(state, "review", "round"))
	if generation < 1 || round < 1 {
		return "", fmt.Errorf("current Runtime has no positive baseline generation and review round")
	}
	segment := auditPathSegment.ReplaceAllString(runtimeID, "-")
	if segment == "" || segment == "." || segment == ".." {
		return "", fmt.Errorf("current runtime_id cannot form an in-repository answer path")
	}
	dir := filepath.ToSlash(filepath.Join("s10", segment, fmt.Sprintf("g%d", generation), fmt.Sprintf("r%d", round)))
	return filepath.ToSlash(filepath.Join(dir, kind+"-manifest.json")), nil
}

func runSavedAuditQuestionnaire(root string, request policy.Input, stdout, stderr io.Writer) (bool, int) {
	if failed, _ := request.ToolResponse["is_error"].(bool); failed {
		return false, 0
	}
	actualPath, saved := savedQuestionnairePath(request)
	if !saved {
		return false, 0
	}
	snapshot, err := runtime.NewStore(filepath.Join(root, ".claude/loop-state.json"), filepath.Join(root, ".claude/loop-events.jsonl")).Snapshot()
	if err != nil {
		return false, 0
	}
	kind := currentAuditKind(snapshot.State)
	if kind == "" {
		return false, 0
	}
	status := inspectS10Artifact(root, snapshot.State, kind)
	if (status.State == "ready" || status.State == "review_required" || status.State == "blocked") && status.ManifestPath != "" {
		expected, pathErr := absoluteQuestionnairePath(root, status.ManifestPath)
		if pathErr == nil && filepath.Clean(actualPath) == filepath.Clean(expected) {
			return true, auditQuestionnaireFeedback(stdout, fmt.Sprintf("S10 %s manifest %s is already registered with its exact current fingerprint. Replay did not register evidence or advance lifecycle; wait for Main/Controller.", kind, expected))
		}
		return false, 0
	}
	path, pathErr := auditQuestionnairePath(root, snapshot.State, kind)
	if pathErr != nil {
		return false, 0
	}
	expected, err := absoluteQuestionnairePath(root, path)
	if err != nil || filepath.Clean(actualPath) != filepath.Clean(expected) {
		return false, 0
	}
	artifactPath, err := safeS10Path(root, path)
	if err != nil {
		return true, auditQuestionnaireFeedback(stdout, fmt.Sprintf("S10 %s questionnaire at exact path %s was not registered: %v", kind, expected, err))
	}

	reject := func(err error) (bool, int) {
		return true, auditQuestionnaireFeedback(stdout, fmt.Sprintf("S10 %s questionnaire at exact path %s was saved but not registered: %v. Correct this same draft and save again.", kind, expected, err))
	}
	data, err := os.ReadFile(artifactPath)
	if err != nil {
		return reject(fmt.Errorf("read saved manifest: %w", err))
	}
	var manifest acceptance.Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return reject(fmt.Errorf("decode saved manifest: %w", err))
	}
	if manifest.ManifestType != kind || manifest.RuntimeID != stringValue(snapshot.State["runtime_id"]) ||
		manifest.BaselineGeneration != integerValue(nestedStateValue(snapshot.State, "baseline", "generation")) ||
		manifest.ReviewRound != integerValue(nestedStateValue(snapshot.State, "review", "round")) {
		return reject(fmt.Errorf("manifest_type, runtime_id, baseline_generation, and review_round must match the current Runtime"))
	}
	if strings.TrimSpace(manifest.Outcome) == "" || !validAuditOutcome(kind, manifest.Outcome) {
		return reject(fmt.Errorf("outcome is required; record the professional conclusion in this manifest, do not infer PASS"))
	}
	actor := "Main" // The original S10 responsibility, not a Hook identity claim.
	baseline, err := acceptance.BuildS10ExternalBaseline(root, snapshot.State, nil)
	if err != nil {
		return reject(fmt.Errorf("current changed-path baseline is unverifiable: %w", err))
	}
	authority, err := acceptance.BuildS10InventoryAuthority(root, snapshot.State, baseline)
	if err != nil {
		return reject(fmt.Errorf("current authoritative inventory is unverifiable: %w", err))
	}
	projectAuditAreaSummary(&manifest)
	manifest.Metrics = acceptance.DeriveMetrics(manifest)
	data, err = json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return reject(fmt.Errorf("encode manifest with derived metrics: %w", err))
	}
	data = append(data, '\n')
	if _, err := acceptance.ValidateForOutcomeWithBaselineAndAuthority(data, kind, manifest.Outcome, baseline, authority); err != nil {
		return reject(err)
	}
	if err := os.WriteFile(artifactPath, data, 0o644); err != nil {
		return reject(fmt.Errorf("write machine-derived metrics to the same manifest: %w", err))
	}

	manifestSHA := sha256HexForArtifact(data)
	shortSHA := manifestSHA[:12]
	evidenceID := fmt.Sprintf("s10-%s-g%d-r%d-%s-%s", auditPathSegment.ReplaceAllString(manifest.RuntimeID, "-"), manifest.BaselineGeneration, manifest.ReviewRound, kind, shortSHA)
	responsibility := "Acceptance"
	if kind == acceptance.ManifestReleaseAudit {
		responsibility = "Release Auditor"
	}
	envelopePath := strings.TrimSuffix(path, "-manifest.json") + "-evidence-" + shortSHA + ".json"
	envelope := map[string]any{
		"schema_version": "1.0.0", "evidence_id": evidenceID, "kind": kind,
		"runtime_id": manifest.RuntimeID, "baseline_generation": manifest.BaselineGeneration,
		"review_round": manifest.ReviewRound, "producer_agent_id": actor,
		"producer_responsibility": responsibility, "subject_refs": []string{},
		"conclusion": manifest.Outcome, "audit_manifest_path": path,
		"audit_manifest_sha256": manifestSHA,
	}
	envelopeData, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return reject(fmt.Errorf("encode machine evidence envelope: %w", err))
	}
	envelopeData = append(envelopeData, '\n')
	envelopeSHA := sha256HexForArtifact(envelopeData)
	if existing, ok := registeredAuditEvidence(snapshot.State, evidenceID, envelopePath, envelopeSHA); ok {
		return true, auditQuestionnaireFeedback(stdout, fmt.Sprintf("S10 %s manifest %s already has current evidence %s; replay did not re-register or advance lifecycle.", kind, expected, existing))
	}
	if evidenceHasID(snapshot.State, evidenceID) {
		return reject(fmt.Errorf("evidence id %s already exists with a different path or fingerprint; no duplicate was registered", evidenceID))
	}
	if err := writeAuditArtifactExclusive(root, envelopePath, envelopeData); err != nil {
		return reject(err)
	}
	round := manifest.ReviewRound
	_, err = runtime.RecordEvidence(root,
		filepath.Join(root, ".claude/loop-state.json"), filepath.Join(root, ".claude/loop-events.jsonl"),
		runtime.EvidenceRequest{
			ExpectedRevision: -1, ID: evidenceID, Kind: kind, Path: envelopePath,
			ProducedBy: []string{actor}, ResponsibilityID: responsibility, ReviewRound: &round,
			Validator: semantic.RuntimeCandidateValidator{},
		})
	if err != nil {
		latest, snapshotErr := runtime.NewStore(filepath.Join(root, ".claude/loop-state.json"), filepath.Join(root, ".claude/loop-events.jsonl")).Snapshot()
		if snapshotErr == nil {
			if existing, ok := registeredAuditEvidence(latest.State, evidenceID, envelopePath, sha256HexForArtifact(envelopeData)); ok {
				return true, auditQuestionnaireFeedback(stdout, fmt.Sprintf("S10 %s manifest %s is registered as %s; concurrent replay did not write a duplicate.", kind, expected, existing))
			}
		}
		return reject(fmt.Errorf("register original S10 evidence: %w", err))
	}
	return true, auditQuestionnaireFeedback(stdout, fmt.Sprintf("S10 %s manifest saved at %s and registered as %s. Runtime/Controller will evaluate the original gate; do not manually transition or publish.", kind, expected, evidenceID))
}

func validAuditOutcome(kind, outcome string) bool {
	if kind == acceptance.ManifestAcceptance {
		return outcome == "pass" || outcome == "review_required"
	}
	return kind == acceptance.ManifestReleaseAudit && (outcome == "approved" || outcome == "approved_with_risk" || outcome == "blocked")
}

func registeredAuditEvidence(state map[string]any, id, path, sha string) (string, bool) {
	for _, raw := range stateEvidence(state) {
		entry, _ := raw.(map[string]any)
		if entry == nil || stringValue(entry["id"]) != id {
			continue
		}
		if stringValue(entry["status"]) == "valid" && stringValue(entry["path"]) == path && stringValue(entry["sha256"]) == sha {
			return id, true
		}
		return id, false
	}
	return "", false
}

func evidenceHasID(state map[string]any, id string) bool {
	for _, raw := range stateEvidence(state) {
		entry, _ := raw.(map[string]any)
		if entry != nil && stringValue(entry["id"]) == id {
			return true
		}
	}
	return false
}

func writeAuditArtifactExclusive(root, rel string, data []byte) error {
	path, err := safeS10Path(root, rel)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create S10 evidence directory for %s: %w", path, err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if os.IsExist(err) {
		existing, readErr := os.ReadFile(path)
		if readErr == nil && bytes.Equal(existing, data) {
			return nil
		}
		return fmt.Errorf("existing evidence path %s has different bytes and was not overwritten", path)
	}
	if err != nil {
		return fmt.Errorf("create S10 evidence envelope %s: %w", path, err)
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return fmt.Errorf("write S10 evidence envelope %s: %w", path, err)
	}
	return file.Close()
}

func auditQuestionnaireFeedback(stdout io.Writer, message string) int {
	_ = json.NewEncoder(stdout).Encode(map[string]any{"hookSpecificOutput": map[string]any{"hookEventName": "PostToolUse", "additionalContext": message}})
	return 0
}
