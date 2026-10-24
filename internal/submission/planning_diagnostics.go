package submission

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/entroforge/go-system-builder/internal/workpacket"
)

// Planning diagnostics describe the last attempted outlet, never its authority.
// Scope excludes Runtime revision: unrelated writes must not erase the cause.
// A new entry cannot inherit an old failure. Reads never rerun a source closure.
func planningDiagnosticPath(state map[string]any, kind string) string {
	life, _ := state["lifecycle"].(map[string]any)
	baseline, _ := state["baseline"].(map[string]any)
	scope, _ := json.Marshal([]any{state["runtime_id"], baseline["generation"], life["state"], life["phase"], life["phase_revision"], kind})
	return ".claude/diagnostics/planning/" + workpacket.Digest(scope) + ".json"
}

// RecordPlanningDiagnostic is best-effort display storage. Failure cannot turn
// an accepted answer or committed transition into a failed business result.
func RecordPlanningDiagnostic(root string, state map[string]any, kind, message string) {
	parent, name, err := workpacket.OpenParent(root, planningDiagnosticPath(state, kind), message != "")
	if err != nil {
		return
	}
	defer parent.Close()
	if message == "" {
		_ = parent.Remove(name)
		return
	}
	data, err := json.Marshal(struct {
		Message string `json:"message"`
		At      string `json:"at"`
	}{message, time.Now().UTC().Format(time.RFC3339Nano)})
	if err != nil || len(data) > 64<<10 {
		return
	}
	var random [16]byte
	if _, err := rand.Read(random[:]); err != nil {
		return
	}
	tmp := fmt.Sprintf(".diagnostic-%x.tmp", random)
	f, err := parent.OpenFile(tmp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	defer parent.Remove(tmp)
	_, writeErr := f.Write(data)
	closeErr := f.Close()
	if writeErr == nil && closeErr == nil {
		_ = parent.Rename(tmp, name)
	}
}

// PlanningDiagnostic returns only a bounded last-attempt diagnostic. Missing or
// damaged display storage is ignored; receipts and gates remain authoritative.
func PlanningDiagnostic(root string, state map[string]any, kind string) string {
	data, err := workpacket.ReadFile(root, planningDiagnosticPath(state, kind), 64<<10)
	if err != nil {
		return ""
	}
	var diagnostic struct {
		Message string `json:"message"`
	}
	if json.Unmarshal(data, &diagnostic) != nil || diagnostic.Message == "" {
		return ""
	}
	return " Last " + kind + " attempt: " + diagnostic.Message
}

// PlanningRecoveryCommand names the existing Main technical recovery entry.
// It never invents a new packet or asks the author to copy identity fields.
func PlanningRecoveryCommand(root, action string) string {
	quote := func(value string) string { return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'" }
	return quote(filepath.Join(root, ".claude/bin/loop-harness")) + " runtime stage-answer " + action + " --root " + quote(root) + " --agent Main"
}
