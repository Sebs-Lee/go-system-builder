package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/workpacket"
	"github.com/entroforge/go-system-builder/internal/workspace"
)

// Display only: neither this file nor its absence grants reception or recovery.
// Original responsibility/source coordinates and exact file observations keep a rejection from being
// projected as the current gap after the answer, source or responsibility changes.
type nativeS9Diagnostic struct {
	Scope   string            `json:"scope"`
	Runtime string            `json:"runtime"`
	Message string            `json:"message"`
	Files   map[string]string `json:"files"`
}

func nativeS9DiagnosticPath(key string) string {
	return ".claude/diagnostics/native-s9/" + workpacket.Digest([]byte(key)) + ".json"
}

func nativeS9FileObservation(path string) string {
	data, err := workpacket.ReadFile(filepath.Dir(path), filepath.Base(path), 20<<20)
	if os.IsNotExist(err) {
		return "missing"
	}
	if err != nil {
		return "unreadable: " + err.Error()
	}
	return workpacket.Digest(data)
}

func recordNativeS9Diagnostic(root string, snap runtime.Snapshot, key, owner string, paths []string, failure error) {
	parent, name, err := workpacket.OpenParent(root, nativeS9DiagnosticPath(key), failure != nil)
	if err != nil {
		return
	}
	defer parent.Close()
	if failure == nil {
		_ = parent.Remove(name)
		return
	}
	scope := nativeS9DiagnosticScope(root, snap.State, key)
	if scope == "" {
		return
	}
	value := nativeS9Diagnostic{Scope: scope, Runtime: stringValue(snap.State["runtime_id"]), Files: map[string]string{}}
	value.Message = fmt.Sprintf("Last S9 Stop attempted reception for %s; exact brief/answer references: %v. Consumer diagnostic: %s. This is the last attempt, not a new validation or completion. Main restores sources/prerequisites or resumes the original professional responsibility; preserve any accepted partial result.", owner, paths, nativeFailureDetail(failure))
	var completion *repairExitFailure
	if errors.As(failure, &completion) {
		value.Message = fmt.Sprintf("Original Targeted observations for %s are accepted; Main owns the unfinished completion outlet. Exact answer references: %v. %s Preserve the accepted independent observations and resume only the original Main completion; do not resume the verifier or repeat its observations. This describes the last completion attempt, not a new validation.", owner, paths, nativeFailureDetail(failure))
		if repairCurrentS7Entry(snap.State) {
			if path := reviewPlanQuestionnairePath(root, snap.State); path != "" {
				paths = append(paths, path, reviewPlanGuidePath(path))
			}
		}
	}
	for _, path := range paths {
		value.Files[path] = nativeS9FileObservation(path)
	}
	data, err := json.Marshal(value)
	if err != nil || len(data) > 64<<10 {
		return
	}
	f, err := parent.OpenFile(name, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	_, _ = f.Write(data)
	_ = f.Close()
}

func nativeS9ReturnDiagnostic(root string, snap runtime.Snapshot, key string) string {
	data, err := workpacket.ReadFile(root, nativeS9DiagnosticPath(key), 64<<10)
	if err != nil {
		return ""
	}
	var value nativeS9Diagnostic
	if json.Unmarshal(data, &value) != nil || value.Scope == "" || value.Scope != nativeS9DiagnosticScope(root, snap.State, key) || value.Runtime != stringValue(snap.State["runtime_id"]) {
		return ""
	}
	for path, observed := range value.Files {
		if nativeS9FileObservation(path) != observed {
			return ""
		}
	}
	return value.Message
}

// Compare recorded identities, not the global write counter: receiving another
// Builder's work must not erase this responsibility's unanswered Stop. These
// are display coordinates only; source validation stays with the consumer.
func nativeS9DiagnosticScope(root string, state map[string]any, key string) string {
	pointer := repairPointerCLI(state)
	source := map[string]any{}
	for _, field := range []string{"case_id", "case_ref", "case_sha256", "contract_id", "contract_ref", "contract_sha256", "session_id", "session_ref", "session_sha256", "plan_id", "plan_ref", "plan_sha256"} {
		source[field] = pointer[field]
	}
	var responsibility any
	packet, found, err := runtime.FindRepairPlanReportPacket(state, key)
	if err != nil {
		return ""
	}
	if found {
		b := packet.Binding
		if b.RuntimeID != state["runtime_id"] || b.BaselineGeneration != workspace.Generation(state) || b.SessionID != pointer["session_id"] || b.PlanRef != pointer["plan_ref"] || stringMapCLI(pointer["assignment_owners"])[b.RepairAssignmentID] != b.OwnerAgentID {
			return ""
		}
		responsibility = packet
	} else {
		// A Targeted answer path is determined by the original source coordinates
		// and one recorded assignment. Resolve it without re-reading Plan/Impact or
		// rerunning the independent-verification source closure.
		var chosen []string
		for assignment, owner := range stringMapCLI(pointer["assignment_owners"]) {
			if path, err := repairTargetedQuestionnairePath(root, state, assignment); err == nil && path == key {
				if chosen != nil {
					return ""
				}
				chosen = []string{assignment, owner}
			}
		}
		if chosen == nil {
			return ""
		}
		responsibility = chosen
		source["impact_ref"], source["impact_sha256"] = pointer["impact_ref"], pointer["impact_sha256"]
	}
	encoded, err := json.Marshal([]any{state["runtime_id"], workspace.Generation(state), lifecycleState(state), nestedStateValue(state, "review", "round"), source, responsibility})
	if err != nil {
		return ""
	}
	return workpacket.Digest(encoded)
}
