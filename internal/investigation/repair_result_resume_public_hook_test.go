package investigation_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/entroforge/go-system-builder/internal/cli"
	"github.com/entroforge/go-system-builder/internal/hook"
	"github.com/entroforge/go-system-builder/internal/hookctx"
	"github.com/entroforge/go-system-builder/internal/policy"
	"github.com/entroforge/go-system-builder/internal/repair"
	"github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/submission"
)

// Mechanical platform inputs exercise the public boundary, not a real-model S9 pass.
func TestS9ResumePublicHookRestoresOnlySuccessorAnswer(t *testing.T) {
	fixture, owner, predecessor, execution, workerRoot, _, _ := prepareWorkingS9RepairResultPacket(t)
	blocked := encodeS9NonpassRepairResultAnswer(t, predecessor, "blocked")
	if err := os.WriteFile(filepath.Join(workerRoot, filepath.FromSlash(predecessor.AnswerPath)), blocked, 0600); err != nil {
		t.Fatal(err)
	}
	accepted, _ := runS9RepairResultPostToolUseWithRaw(t, fixture.root, owner.agentID, workerRoot, execution.PlatformSessionID, predecessor.AnswerPath)
	if !accepted.Accepted || accepted.WorkComplete {
		t.Fatalf("original blocker not accepted: %+v", accepted)
	}
	path, err := repair.RepairResultResolutionAuthoringPath(predecessor)
	if err != nil {
		t.Fatal(err)
	}
	abs := filepath.Join(fixture.root, filepath.FromSlash(path))
	var draft map[string]any
	if err := json.Unmarshal(mustRead(t, abs), &draft); err != nil {
		t.Fatal(err)
	}
	draft["s9_resolution"].(map[string]any)["rationale"] = "Dependency restored; retry within the original immutable repair scope."
	data, err := json.Marshal(draft)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(abs, data, 0600); err != nil {
		t.Fatal(err)
	}
	resolved, _ := runS9RepairResultHookPayloadWithRaw(t, fixture.root, s9RepairResultHookPayload("", fixture.root, "main-resume-public-session", abs))
	if !resolved.Accepted || resolved.WorkComplete {
		t.Fatalf("Main resolution not accepted: %+v", resolved)
	}
	before := currentS9Snapshot(t, fixture)
	if pending, err := runtime.RepairResultResumePending(fixture.root, before.State, owner.agentID, predecessor.Binding.Sources.RepairAssignmentID); err != nil || !pending {
		t.Fatalf("resolved predecessor did not wait for Main: pending=%t err=%v", pending, err)
	}
	listenerInput, err := json.Marshal(map[string]any{
		"hook_event_name": "SessionStart", "cwd": fixture.root, "session_id": "main-resume-listener-session",
	})
	if err != nil {
		t.Fatal(err)
	}
	var listenerOut, listenerErr bytes.Buffer
	if code := cli.Run([]string{"hook", "--event", "SessionStart", "--root", fixture.root}, bytes.NewReader(listenerInput), &listenerOut, &listenerErr); code != 0 {
		t.Fatalf("original Main SessionStart listener did not prepare the Resume successor: code=%d stdout=%s stderr=%s", code, listenerOut.String(), listenerErr.String())
	}
	var listenerEnvelope struct {
		Output struct {
			Context string `json:"additionalContext"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal(listenerOut.Bytes(), &listenerEnvelope); err != nil {
		t.Fatalf("decode Main Resume listener context: %v; stdout=%s", err, listenerOut.String())
	}
	current := currentS9Snapshot(t, fixture)
	packets, err := runtime.CurrentRepairResultPackets(fixture.root, current.State, owner.agentID)
	if err != nil || len(packets) != 1 {
		t.Fatalf("no unique successor: packets=%v err=%v", packets, err)
	}
	successor := packets[0]
	if successor.ID == predecessor.ID || successor.AnswerPath == predecessor.AnswerPath || successor.Binding.Resume == nil || successor.Binding.Resume.PredecessorPacketID != predecessor.ID {
		t.Fatalf("successor reused the retired identity: %+v", successor)
	}
	sourceRoot, successorExecution, err := submission.RepairResultRegisteredSource(fixture.root, current.State, successor)
	if err != nil || filepath.Clean(sourceRoot) != filepath.Clean(workerRoot) || successorExecution.AssignmentID != execution.AssignmentID || successorExecution.Path != execution.Path || successorExecution.Generation != execution.Generation {
		t.Fatalf("Resume successor lost its original registered Worker source: root=%s execution=%+v err=%v", sourceRoot, successorExecution, err)
	}
	wantView := filepath.Join(sourceRoot, filepath.FromSlash(successor.ViewPath))
	wantAnswer := filepath.Join(sourceRoot, filepath.FromSlash(successor.AnswerPath))
	if !strings.Contains(listenerEnvelope.Output.Context, wantView) || !strings.Contains(listenerEnvelope.Output.Context, wantAnswer) || strings.Contains(listenerEnvelope.Output.Context, filepath.Join(sourceRoot, filepath.FromSlash(predecessor.AnswerPath))) {
		t.Fatalf("Main listener did not give Worker the complete current Resume questionnaire path, or revived the predecessor path: want view=%s answer=%s context=%s", wantView, wantAnswer, listenerEnvelope.Output.Context)
	}
	if pending, err := runtime.RepairResultResumePending(fixture.root, current.State, owner.agentID, successor.Binding.Sources.RepairAssignmentID); err != nil || pending {
		t.Fatalf("verified successor still waits for issuance: pending=%t err=%v", pending, err)
	}
	if err := runtime.VerifyRepairResultPacketCurrent(fixture.root, current.State, predecessor); err == nil {
		t.Fatal("retired predecessor became current again")
	}
	worker, err := hookctx.FromSnapshot(fixture.root, owner.agentID, current)
	if err != nil || worker == nil || worker.PolicyContext.Agent == nil {
		t.Fatalf("Worker projection missing: %v", err)
	}
	if worker.PolicyContext.Agent.RepairResultResumePending || len(worker.PolicyContext.RepairResultAnswerPaths) != 1 || worker.PolicyContext.RepairResultAnswerPaths[0] != successor.AnswerPath {
		t.Fatalf("only new answer should be writable: %+v", worker.PolicyContext)
	}
	for _, tc := range []struct {
		path string
		deny bool
	}{{predecessor.AnswerPath, true}, {successor.AnswerPath, false}} {
		input := policy.Input{Event: "PreToolUse", ToolName: "Write", AgentID: owner.agentID, CWD: workerRoot, SessionID: execution.PlatformSessionID, Runtime: worker.PolicyContext,
			ToolInput: map[string]any{"file_path": filepath.Join(workerRoot, filepath.FromSlash(tc.path))}}
		decision, denied := policy.EvaluateAgentScoped(input)
		if denied != tc.deny {
			t.Fatalf("answer grant %s: denied=%t want=%t decision=%+v", tc.path, denied, tc.deny, decision)
		}
	}
	if mayEnd, err := repair.RepairResultBlockerExecutionMayEnd(fixture.root, current.State, predecessor.ID, owner.agentID); err != nil || mayEnd {
		t.Fatalf("historical blocker revived Stop: end=%t err=%v", mayEnd, err)
	}
	if decision, handled := hook.StopIdleDecision(fixture.root, policy.Input{Event: "SubagentStop", AgentID: owner.agentID, CWD: workerRoot, SessionID: execution.PlatformSessionID}); !handled || decision.Decision != "deny" {
		t.Fatalf("successor ended before new delivery: handled=%t decision=%+v", handled, decision)
	}
	stateBytes, journalBytes := mustRead(t, fixture.statePath), mustRead(t, fixture.journalPath)
	listenerOut.Reset()
	listenerErr.Reset()
	if code := cli.Run([]string{"hook", "--event", "SessionStart", "--root", fixture.root}, bytes.NewReader(listenerInput), &listenerOut, &listenerErr); code != 0 {
		t.Fatalf("exact Main SessionStart successor replay failed: code=%d stderr=%s stdout=%s", code, listenerErr.String(), listenerOut.String())
	}
	if err := json.Unmarshal(listenerOut.Bytes(), &listenerEnvelope); err != nil || !strings.Contains(listenerEnvelope.Output.Context, wantAnswer) {
		t.Fatalf("exact Main listener replay lost the current successor path: err=%v stdout=%s", err, listenerOut.String())
	}
	assertS9RuntimeBytesUnchanged(t, fixture, stateBytes, journalBytes, "successor listener exact replay")
	old, _ := runS9RepairResultPostToolUseWithRaw(t, fixture.root, owner.agentID, workerRoot, execution.PlatformSessionID, predecessor.AnswerPath)
	if old.WorkComplete {
		t.Fatalf("old answer completed successor: %+v", old)
	}
	assertS9RuntimeBytesUnchanged(t, fixture, stateBytes, journalBytes, "retired answer save")
}
