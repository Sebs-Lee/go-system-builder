package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/entroforge/go-system-builder/internal/acceptance"
	"github.com/entroforge/go-system-builder/internal/policy"
	"github.com/entroforge/go-system-builder/internal/runtime"
)

func stageHumanReleaseQuestionnaireContext(root string, request policy.Input, snapshot runtime.Snapshot) string {
	if !mainQuestionnaireViewer(request) {
		return ""
	}
	if stringValue(nestedStateValue(snapshot.State, "lifecycle", "state")) != "awaiting_human_release" {
		return ""
	}
	var packages []string
	for _, kind := range []string{acceptance.ManifestAcceptance, acceptance.ManifestReleaseAudit} {
		status := inspectS10Artifact(root, snapshot.State, kind)
		if status.State != "ready" || status.ManifestPath == "" || status.EvidenceID == "" {
			return fmt.Sprintf("S11 is waiting for the human, but the current %s evidence package path cannot be verified from this Runtime. Do not infer a file from templates or directory names; Main must repair the S10 package references.", kind)
		}
		entry := evidenceEntryByID(snapshot.State, status.EvidenceID)
		if entry == nil {
			return fmt.Sprintf("S11 %s package evidence %s has no current Runtime path; do not guess.", kind, status.EvidenceID)
		}
		evidencePath, err := absoluteQuestionnairePath(root, stringValue(entry["path"]))
		if err != nil {
			return fmt.Sprintf("S11 %s package evidence %s has an unresolved Runtime path: %v", kind, status.EvidenceID, err)
		}
		manifestPath, err := absoluteQuestionnairePath(root, status.ManifestPath)
		if err != nil {
			return fmt.Sprintf("S11 %s package manifest path is unresolved: %v", kind, err)
		}
		packages = append(packages, fmt.Sprintf("%s evidence %s: %s (manifest %s)", kind, status.EvidenceID, evidencePath, manifestPath))
	}
	packageText := strings.Join(packages, "\n")
	runtimeScope := "runtime_release:" + stringValue(snapshot.State["runtime_id"])
	legacyScope := fmt.Sprintf("%s@%d", runtimeScope, snapshot.Revision)
	for _, raw := range stateEvidence(snapshot.State) {
		entry, _ := raw.(map[string]any)
		if entry == nil || stringValue(entry["kind"]) != "human_decision" || stringValue(entry["status"]) != "valid" || runtime.HumanDecisionEvidenceConsumed(entry) {
			continue
		}
		if !evidenceIncludesScope(entry, runtimeScope) && !evidenceIncludesScope(entry, legacyScope) {
			continue
		}
		path, err := absoluteQuestionnairePath(root, stringValue(entry["path"]))
		if err == nil {
			return fmt.Sprintf("S11 human review package paths:\n%s\nExisting current human decision evidence %s: %s. Only the human chooses a disposition; the existing gateway consumes this registered evidence. This Hook does not publish or deploy.", packageText, stringValue(entry["id"]), path)
		}
	}

	path, decisionID, err := humanReleaseQuestionnairePath(root, snapshot.State)
	if err != nil {
		return fmt.Sprintf("S11 human review package paths:\n%s\nThe exact human decision application path cannot be derived from the current Runtime: %v. Do not guess a path or disposition.", packageText, err)
	}
	if questionnairePreparationEvent(request.Event) {
		if err := prepareHumanDecisionApplication(path, snapshot.State, decisionID); err != nil {
			return fmt.Sprintf("S11 human review package paths:\n%s\nThe system could not prepare the human decision application at %s: %v", packageText, path, err)
		}
		return humanDecisionApplicationContext(packageText, path, decisionID)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Sprintf("S11 human review package paths:\n%s\nThe human decision application is not prepared yet; its exact path is %s. This Hook event is read-only. SessionStart or PreToolUse prepares it; Stop/Idle only reports the missing application.", packageText, path)
	} else if err != nil {
		return fmt.Sprintf("S11 human decision application at %s could not be inspected: %v", path, err)
	}
	return humanDecisionApplicationContext(packageText, path, decisionID)
}

func humanReleaseQuestionnairePath(root string, state map[string]any) (string, string, error) {
	runtimeID := strings.TrimSpace(stringValue(state["runtime_id"]))
	if runtimeID == "" {
		return "", "", fmt.Errorf("current Runtime has no runtime_id")
	}
	generation := integerValue(nestedStateValue(state, "baseline", "generation"))
	round := integerValue(nestedStateValue(state, "review", "round"))
	if generation < 1 || round < 1 {
		return "", "", fmt.Errorf("current Runtime has no positive baseline generation and review round")
	}
	segment := auditPathSegment.ReplaceAllString(runtimeID, "-")
	if segment == "" || segment == "." || segment == ".." {
		return "", "", fmt.Errorf("current runtime_id cannot form the human decision application path")
	}
	decisionID := fmt.Sprintf("hd-s11-%s-g%d-r%d", segment, generation, round)
	path, err := absoluteQuestionnairePath(root, filepath.ToSlash(filepath.Join(".claude", "decisions", decisionID+".json")))
	return path, decisionID, err
}

func prepareHumanDecisionApplication(path string, state map[string]any, decisionID string) error {
	data, err := json.MarshalIndent(map[string]any{
		"decision": "", "disposition": "", "decision_id": decisionID,
		"reason": "", "approved_by": "", "runtime_id": stringValue(state["runtime_id"]),
		"target_cursor": map[string]any{"state": "awaiting_human_release", "phase": nil},
		"scope":         "runtime_release:" + stringValue(state["runtime_id"]),
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("encode human decision application: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create human decision application directory: %w", err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if os.IsExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("create human decision application: %w", err)
	}
	if _, err := file.Write(append(data, '\n')); err != nil {
		_ = file.Close()
		return fmt.Errorf("write human decision application: %w", err)
	}
	return file.Close()
}

func humanDecisionApplicationContext(packageText, path, decisionID string) string {
	return fmt.Sprintf("S11 human review package paths:\n%s\nHuman decision application: %s (decision_id %s). The human must personally choose and fill decision/disposition, reason, and approved_by; valid dispositions are approve, defer, reject_defect, reject_acceptance, reject_release_audit, or abort. This is only an application draft: saving it does not register human_decision evidence, approve the release, or publish/deploy. The existing human gateway still requires registered decision evidence and the actual human actor.", packageText, path, decisionID)
}

func evidenceIncludesScope(entry map[string]any, want string) bool {
	refs, _ := entry["scope_refs"].([]any)
	for _, ref := range refs {
		if stringValue(ref) == want {
			return true
		}
	}
	return false
}

func evidenceEntryByID(state map[string]any, id string) map[string]any {
	for _, raw := range stateEvidence(state) {
		entry, _ := raw.(map[string]any)
		if entry != nil && stringValue(entry["id"]) == id {
			return entry
		}
	}
	return nil
}
