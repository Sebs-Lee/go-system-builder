package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/submission"
	"github.com/entroforge/go-system-builder/internal/workpacket"
	"github.com/entroforge/go-system-builder/internal/workspace"
)

func TestNativeS9OriginalFlatConsumeReceivesAndProjectsReplay(t *testing.T) {
	root, worker, p, _, _ := nativeS9WorkerFixture(t)
	_, _, relative := nativeS9Paths(p)
	path := filepath.Join(worker, relative)
	var out, diagnostic bytes.Buffer
	if code := runStageAnswer([]string{"consume", "--root", root, "--path", path}, &out, &diagnostic); code != 0 {
		t.Fatalf("public original-input recovery: %d %s %s", code, &out, &diagnostic)
	}
	snap := repairDispatchSnapshot(t, root)
	registry, err := workspace.Decode(snap.State)
	if err != nil || registry == nil || registry.Executions[p.Binding.ManifestAssignmentID].DeliveryRef == "" {
		t.Fatalf("did not receive original candidate: %v %s", err, &out)
	}
	before := repairDispatchRuntimeBytes(t, root)
	// System successor recovery never needs the now-consumed private files.
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(worker, p.AnswerPath)); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	diagnostic.Reset()
	if code := runStageAnswer([]string{"consume", "--root", root, "--path", path}, &out, &diagnostic); code != 0 {
		t.Fatalf("accepted replay: %d %s %s", code, &out, &diagnostic)
	}
	repairDispatchAssertUnchanged(t, root, before)
}

func TestNativeS9OriginalConsumeDoesNotReimportAcceptedPlan(t *testing.T) {
	root, worker, p, _, _ := nativeS9WorkerFixture(t)
	service, err := submission.New(root)
	if err != nil {
		t.Fatal(err)
	}
	var diagnostic bytes.Buffer
	if err := importNativeS9PlanReport(context.Background(), root, worker, p, service, &diagnostic); err != nil {
		t.Fatal(err)
	}
	if receipt, err := service.SaveRepairPlanReportFrom(context.Background(), p.Binding.OwnerAgentID, filepath.Join(root, p.AnswerPath), root, nil); err != nil || !receipt.Accepted {
		t.Fatalf("plan receipt: %#v %v", receipt, err)
	}
	if err := os.WriteFile(filepath.Join(worker, p.AnswerPath), []byte("new unfinished private Plan"), 0600); err != nil {
		t.Fatal(err)
	}
	_, _, relative := nativeS9Paths(p)
	if _, err := consumeRegisteredStageAnswer(context.Background(), service, filepath.Join(worker, relative)); err != nil {
		t.Fatal(err)
	}
}

func TestNativeS9OriginalConsumePreservesSavedTypedInput(t *testing.T) {
	root, worker, p, _, professional := nativeS9WorkerFixture(t)
	ctx := context.Background()
	service, err := submission.New(root)
	if err != nil {
		t.Fatal(err)
	}
	var diagnostic bytes.Buffer
	if err := importNativeS9PlanReport(ctx, root, worker, p, service, &diagnostic); err != nil {
		t.Fatal(err)
	}
	if r, err := service.SaveRepairPlanReportFrom(ctx, p.Binding.OwnerAgentID, filepath.Join(root, p.AnswerPath), root, nil); err != nil || !r.Accepted {
		t.Fatalf("plan: %#v %v", r, err)
	}
	base := repairLeadGit(t, root, "rev-parse", "HEAD")
	if err := recordNativeS9Source(ctx, root, worker, base, p); err != nil {
		t.Fatal(err)
	}
	packets, err := service.IssueRepairResults(ctx, p.Binding.OwnerAgentID)
	if err != nil || len(packets) != 1 {
		t.Fatalf("packets: %d %v", len(packets), err)
	}
	packet := packets[0]
	raw, _ := json.Marshal(packet.Draft)
	if nativeS9TypedInputSaved(packet, raw) {
		t.Fatal("original draft claimed as saved typed answer")
	}
	draft := packet.Draft
	draft.Answer, _ = json.Marshal(professional)
	if err := persistRepairLeadQuestionnaire(filepath.Join(worker, packet.AnswerPath), draft); err != nil {
		t.Fatal(err)
	}
	saved, err := os.ReadFile(filepath.Join(worker, packet.AnswerPath))
	if err != nil {
		t.Fatal(err)
	}
	if !nativeS9TypedInputSaved(packet, saved) {
		t.Fatal("saved typed answer treated as draft")
	}
	_, _, relative := nativeS9Paths(p)
	if err := os.WriteFile(filepath.Join(worker, relative), []byte("changed unfinished flat artifact"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := consumeRegisteredStageAnswer(ctx, service, filepath.Join(worker, relative)); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(filepath.Join(worker, packet.AnswerPath))
	if err != nil || !bytes.Equal(saved, after) {
		t.Fatalf("saved typed bytes replaced: %v", err)
	}
}

func TestNativeS9OriginalConsumeRejectsMainAndWrongPacket(t *testing.T) {
	root, _, p, _, _ := nativeS9WorkerFixture(t)
	service, err := submission.New(root)
	if err != nil {
		t.Fatal(err)
	}
	_, _, relative := nativeS9Paths(p)
	before := repairDispatchRuntimeBytes(t, root)
	for _, path := range []string{filepath.Join(root, relative), filepath.Join(root, ".claude/submissions/native-s9/wrong-plan/result.json")} {
		if _, err := consumeRegisteredStageAnswer(context.Background(), service, path); err == nil {
			t.Fatalf("invalid source accepted: %s", path)
		}
		repairDispatchAssertUnchanged(t, root, before)
	}
}

func TestNativeS9ExplicitStopChangeStillConflictsAfterRecovery(t *testing.T) {
	root, worker, p, input, professional := nativeS9WorkerFixture(t)
	service, err := submission.New(root)
	if err != nil {
		t.Fatal(err)
	}
	_, _, relative := nativeS9Paths(p)
	if _, err := consumeRegisteredStageAnswer(context.Background(), service, filepath.Join(worker, relative)); err != nil {
		t.Fatal(err)
	}
	before := repairDispatchRuntimeBytes(t, root)
	professional.ResidualRisks = []string{"explicit new professional content"}
	if err := persistRepairLeadQuestionnaire(filepath.Join(worker, relative), professional); err != nil {
		t.Fatal(err)
	}
	var diagnostic bytes.Buffer
	input.Event = "SubagentStop"
	err = receiveNativeS9Builder(root, repairDispatchSnapshot(t, root).State, input, p, input.LastAssistantMessage, &diagnostic)
	if !errors.Is(err, runtime.ErrSubmissionConflict) {
		t.Fatalf("new explicit professional content lost original conflict: %v %s", err, &diagnostic)
	}
	repairDispatchAssertUnchanged(t, root, before)
}

// This fixture captures a real accepted candidate and its actual Writer event,
// then exposes the same staged/durable boundaries as the Store integrity tests.
// It does not invent a business event or a replacement candidate.
func nativeS9OriginalCandidateBoundary(t *testing.T) (root, worker string, plan workpacket.RepairPlanReportPacket, packet workpacket.RepairResultPacket, previousState, previousJournal, targetState, targetJournal []byte) {
	t.Helper()
	root, worker, plan, _, professional := nativeS9WorkerFixture(t)
	ctx := context.Background()
	service, err := submission.New(root)
	if err != nil {
		t.Fatal(err)
	}
	var diagnostic bytes.Buffer
	if err := importNativeS9PlanReport(ctx, root, worker, plan, service, &diagnostic); err != nil {
		t.Fatal(err)
	}
	if r, err := service.SaveRepairPlanReportFrom(ctx, plan.Binding.OwnerAgentID, filepath.Join(root, plan.AnswerPath), root, nil); err != nil || !r.Accepted {
		t.Fatalf("plan receipt: %#v %v", r, err)
	}
	if err := recordNativeS9Source(ctx, root, worker, repairLeadGit(t, root, "rev-parse", "HEAD"), plan); err != nil {
		t.Fatal(err)
	}
	packets, err := service.IssueRepairResults(ctx, plan.Binding.OwnerAgentID)
	if err != nil || len(packets) != 1 {
		t.Fatalf("packet issue: %d %v", len(packets), err)
	}
	packet = packets[0]
	draft := packet.Draft
	draft.Answer, _ = json.Marshal(professional)
	if err := persistRepairLeadQuestionnaire(filepath.Join(worker, packet.AnswerPath), draft); err != nil {
		t.Fatal(err)
	}
	previousState, err = os.ReadFile(service.StatePath)
	if err != nil {
		t.Fatal(err)
	}
	previousJournal, err = os.ReadFile(service.JournalPath)
	if err != nil {
		t.Fatal(err)
	}
	if r, err := service.SaveRepairResultFrom(ctx, plan.Binding.OwnerAgentID, filepath.Join(worker, packet.AnswerPath), worker, nil); err != nil || !r.Accepted {
		t.Fatalf("candidate receipt: %#v %v", r, err)
	}
	targetState, err = os.ReadFile(service.StatePath)
	if err != nil {
		t.Fatal(err)
	}
	targetJournal, err = os.ReadFile(service.JournalPath)
	if err != nil {
		t.Fatal(err)
	}
	return
}

func TestNativeS9OriginalConsumePendingPrecedesMissingTypedAndChangedFlat(t *testing.T) {
	for _, stateWritten := range []bool{false, true} {
		name := "before_state_publication"
		if stateWritten {
			name = "after_state_before_journal"
		}
		t.Run(name, func(t *testing.T) {
			root, worker, p, packet, previousState, previousJournal, targetState, targetJournal := nativeS9OriginalCandidateBoundary(t)
			statePath := filepath.Join(root, ".claude/loop-state.json")
			journalPath := filepath.Join(root, ".claude/loop-events.jsonl")
			markerBytes := nativeS9OriginalPendingMarker(t, previousState, targetState, targetJournal)
			var marker map[string]any
			if err := json.Unmarshal(markerBytes, &marker); err != nil {
				t.Fatal(err)
			}
			marker["retain_last_transition"] = true
			markerBytes, err := json.MarshalIndent(marker, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			currentState := previousState
			if stateWritten {
				currentState = targetState
			}
			if err := os.WriteFile(statePath, currentState, 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(journalPath, previousJournal, 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(statePath+".commit-pending.json", markerBytes, 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.Remove(filepath.Join(worker, packet.AnswerPath)); err != nil {
				t.Fatal(err)
			}
			_, _, relative := nativeS9Paths(p)
			flatPath := filepath.Join(worker, relative)
			replacement := []byte("later unrelated/unfinished flat bytes")
			if err := os.WriteFile(flatPath, replacement, 0600); err != nil {
				t.Fatal(err)
			}
			// A read-only query reports the boundary and cannot finish the pair.
			if _, err := runtime.NewStore(statePath, journalPath).Snapshot(); !errors.Is(err, runtime.ErrPendingRuntimeOperation) {
				t.Fatalf("missing real pending boundary: %v", err)
			}
			var out, diagnostic bytes.Buffer
			if code := runStageAnswer([]string{"consume", "--root", root, "--path", flatPath}, &out, &diagnostic); code != 0 {
				t.Fatalf("public pending recovery: %d %s %s", code, &out, &diagnostic)
			}
			afterState, err := os.ReadFile(statePath)
			if err != nil {
				t.Fatal(err)
			}
			afterJournal, err := os.ReadFile(journalPath)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(afterState, targetState) || !bytes.Equal(afterJournal, targetJournal) {
				t.Fatal("pending recovery replaced original candidate or added a second business event")
			}
			if _, err := os.Stat(statePath + ".commit-pending.json"); !os.IsNotExist(err) {
				t.Fatalf("original pending marker not retired: %v", err)
			}
			if _, err := os.Stat(filepath.Join(worker, packet.AnswerPath)); !os.IsNotExist(err) {
				t.Fatalf("recovery recreated typed input from changed flat: %v", err)
			}
			if after, err := os.ReadFile(flatPath); err != nil || !bytes.Equal(after, replacement) {
				t.Fatalf("recovery changed later private flat bytes: %v", err)
			}
			beforeReplay := repairDispatchRuntimeBytes(t, root)
			out.Reset()
			diagnostic.Reset()
			if code := runStageAnswer([]string{"consume", "--root", root, "--path", flatPath}, &out, &diagnostic); code != 0 {
				t.Fatalf("pending successor replay: %d %s %s", code, &out, &diagnostic)
			}
			repairDispatchAssertUnchanged(t, root, beforeReplay)
		})
	}
}

func TestNativeS9OriginalConsumeMissingTypedDoesNotOverwriteStagedCandidate(t *testing.T) {
	root, worker, p, packet, previousState, previousJournal, _, _ := nativeS9OriginalCandidateBoundary(t)
	accepted := repairDispatchSnapshot(t, root)
	registry, err := workspace.Decode(accepted.State)
	if err != nil {
		t.Fatal(err)
	}
	candidatePath := filepath.Join(root, registry.Executions[p.Binding.ManifestAssignmentID].DeliveryRef)
	staged, err := os.ReadFile(candidatePath)
	if err != nil {
		t.Fatal(err)
	}
	// Expose the real staged artifact before its original CAS/pending publication.
	statePath := filepath.Join(root, ".claude/loop-state.json")
	journalPath := filepath.Join(root, ".claude/loop-events.jsonl")
	if err := os.WriteFile(statePath, previousState, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(journalPath, previousJournal, 0600); err != nil {
		t.Fatal(err)
	}
	typedPath := filepath.Join(worker, packet.AnswerPath)
	typedBytes, err := os.ReadFile(typedPath)
	if err != nil {
		t.Fatal(err)
	}
	original := repairDispatchSnapshot(t, root)
	capture := newNativeS9AttemptCapture(context.Background(), root, original, p.ID, p.Binding.OwnerAgentID, "Builder")
	capture.SetTypedInput(typedPath)
	capture.SetStage("typed_result_saved")
	capture.ObserveInput(typedPath, typedBytes)
	if err := capture.Publish(context.Background(), errors.New("controlled candidate reception interruption before pending publication")); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(typedPath); err != nil {
		t.Fatal(err)
	}
	_, _, relative := nativeS9Paths(p)
	flatPath := filepath.Join(worker, relative)
	replacement := []byte("later different flat input")
	if err := os.WriteFile(flatPath, replacement, 0600); err != nil {
		t.Fatal(err)
	}
	before := repairDispatchRuntimeBytes(t, root)
	var out, diagnostic bytes.Buffer
	if code := runStageAnswer([]string{"consume", "--root", root, "--path", flatPath}, &out, &diagnostic); code == 0 {
		t.Fatalf("missing original typed input silently replaced from flat: %s %s", &out, &diagnostic)
	}
	repairDispatchAssertUnchanged(t, root, before)
	out.Reset()
	diagnostic.Reset()
	if code := runStageAnswer([]string{"consume", "--root", root, "--path", flatPath}, &out, &diagnostic); code == 0 {
		t.Fatalf("repeated missing typed recovery lost its saved-input window: %s %s", &out, &diagnostic)
	}
	repairDispatchAssertUnchanged(t, root, before)
	if after, err := os.ReadFile(candidatePath); err != nil || !bytes.Equal(after, staged) {
		t.Fatalf("staged immutable candidate replaced: %v", err)
	}
	if _, err := os.Stat(filepath.Join(worker, packet.AnswerPath)); !os.IsNotExist(err) {
		t.Fatalf("missing typed bytes fabricated from flat: %v", err)
	}
}

func TestNativeS9OriginalStopUsesNewProfessionalInputAfterSourceGap(t *testing.T) {
	root, worker, p, input, professional := nativeS9WorkerFixture(t)
	_, _, relative := nativeS9Paths(p)
	professional.EvidenceRefs = []string{"path:logs/original-missing-check.txt"}
	if err := persistRepairLeadQuestionnaire(filepath.Join(worker, relative), professional); err != nil {
		t.Fatal(err)
	}
	var diagnostic bytes.Buffer
	if err := receiveNativeS9Builder(root, repairDispatchSnapshot(t, root).State, input, p, input.LastAssistantMessage, &diagnostic); err == nil {
		t.Fatal("missing actual evidence unexpectedly accepted")
	}
	partial := repairDispatchSnapshot(t, root)
	registry, err := workspace.Decode(partial.State)
	if err != nil || registry == nil || registry.Executions[p.Binding.ManifestAssignmentID].DeliveryRef != "" {
		t.Fatalf("source gap has an accepted candidate: %v", err)
	}
	// This new actual Stop is a new explicit professional save. The old adapted
	// but unaccepted typed answer must not mask its corrected observations.
	professional.EvidenceRefs = []string{"internal/api/one.go"}
	professional.ResidualRisks = []string{"new professional observation after restoring the source gap"}
	if err := persistRepairLeadQuestionnaire(filepath.Join(worker, relative), professional); err != nil {
		t.Fatal(err)
	}
	diagnostic.Reset()
	if err := receiveNativeS9Builder(root, partial.State, input, p, input.LastAssistantMessage, &diagnostic); err != nil {
		t.Fatalf("new Stop swallowed by stale unaccepted typed bytes: %v %s", err, &diagnostic)
	}
	accepted := repairDispatchSnapshot(t, root)
	registry, err = workspace.Decode(accepted.State)
	if err != nil {
		t.Fatal(err)
	}
	candidatePath := filepath.Join(root, registry.Executions[p.Binding.ManifestAssignmentID].DeliveryRef)
	raw, err := os.ReadFile(candidatePath)
	if err != nil {
		t.Fatal(err)
	}
	var candidate struct {
		Provenance struct {
			SubmittedAnswer json.RawMessage `json:"submitted_answer"`
		} `json:"provenance"`
	}
	if err := json.Unmarshal(raw, &candidate); err != nil {
		t.Fatal(err)
	}
	var actual workpacket.RepairResultProfessionalAnswer
	if err := json.Unmarshal(candidate.Provenance.SubmittedAnswer, &actual); err != nil {
		t.Fatal(err)
	}
	if len(actual.ResidualRisks) != 1 || actual.ResidualRisks[0] != professional.ResidualRisks[0] {
		t.Fatalf("new Stop did not record its actual new professional input: %s", candidate.Provenance.SubmittedAnswer)
	}
}

func TestNativeS9OriginalConsumeMissingUnadaptedDraftCanAdaptExplicitFlat(t *testing.T) {
	root, worker, p, packet, previousState, previousJournal, _, _ := nativeS9OriginalCandidateBoundary(t)
	accepted := repairDispatchSnapshot(t, root)
	registry, err := workspace.Decode(accepted.State)
	if err != nil {
		t.Fatal(err)
	}
	// Retain the original pre-adaptation issuance but remove all later candidate
	// facts/artifacts. There is no successful adaptation observation or pending.
	if err := os.Remove(filepath.Join(root, registry.Executions[p.Binding.ManifestAssignmentID].DeliveryRef)); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".claude/loop-state.json"), previousState, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".claude/loop-events.jsonl"), previousJournal, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(worker, packet.AnswerPath)); err != nil {
		t.Fatal(err)
	}
	_, _, relative := nativeS9Paths(p)
	var out, diagnostic bytes.Buffer
	if code := runStageAnswer([]string{"consume", "--root", root, "--path", filepath.Join(worker, relative)}, &out, &diagnostic); code != 0 {
		t.Fatalf("missing unadapted draft prevented original explicit flat adaptation: %d %s %s", code, &out, &diagnostic)
	}
	accepted = repairDispatchSnapshot(t, root)
	registry, err = workspace.Decode(accepted.State)
	if err != nil || registry.Executions[p.Binding.ManifestAssignmentID].DeliveryRef == "" {
		t.Fatalf("explicit original flat did not reach original consumer: %v", err)
	}
}

// Match the original Store commit-pending wire protocol using a real candidate
// state and its original final journal event, as in Store integrity fixtures.
func nativeS9OriginalPendingMarker(t *testing.T, previousStateBytes, targetStateBytes, targetJournalBytes []byte) []byte {
	t.Helper()
	var previousState, targetState map[string]any
	if err := json.Unmarshal(previousStateBytes, &previousState); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(targetStateBytes, &targetState); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(targetJournalBytes)), "\n")
	if len(lines) == 0 || lines[len(lines)-1] == "" {
		t.Fatal("target journal has no original committed event")
	}
	var event map[string]any
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &event); err != nil {
		t.Fatal(err)
	}
	eventBytes, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	stateHash := func(state map[string]any) string {
		encoded, err := json.MarshalIndent(state, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		return workpacket.Digest(append(encoded, '\n'))
	}
	marker := map[string]any{
		"schema_version": "1.0.0", "previous_state_sha256": stateHash(previousState),
		"previous_revision": previousState["revision"], "state_sha256": stateHash(targetState),
		"journal_event_sha256": workpacket.Digest(append(eventBytes, '\n')), "request_id": event["request_id"],
		"idempotency_key": event["idempotency_key"], "retain_last_transition": true,
		"state": targetState, "journal_event": event,
	}
	encoded, err := json.MarshalIndent(marker, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return append(encoded, '\n')
}

func TestNativeS9OriginalConsumeAcceptedPlanSurvivesMissingDisplayBrief(t *testing.T) {
	root, worker, p, input, _ := nativeS9WorkerFixture(t)
	brief, pin, relative := nativeS9Paths(p)
	pinPath := filepath.Join(root, pin)
	pinBytes, err := os.ReadFile(pinPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(pinPath); err != nil {
		t.Fatal(err)
	}
	input.Event = "SubagentStop"
	var out, diagnostic bytes.Buffer
	if handled, code := nativeS9Hook(root, input, &out, &diagnostic); !handled || code != 0 {
		t.Fatalf("original failed Stop: %t %d %s", handled, code, &diagnostic)
	}
	partial := repairDispatchSnapshot(t, root)
	registry, err := workspace.Decode(partial.State)
	if err != nil || registry != nil && len(registry.Executions) != 0 {
		t.Fatalf("pin failure incorrectly registered a source: %v", err)
	}
	service, err := submission.New(root)
	if err != nil {
		t.Fatal(err)
	}
	if receipt, err := service.RepairPlanReportReceipt(p.Binding.OwnerAgentID, p.ID); err != nil || !receipt.Accepted {
		t.Fatalf("original partial Plan missing: %#v %v", receipt, err)
	}
	// Lose the display brief while the exact captured original native binding
	// still identifies the same failed source. Missing pin stays a source error.
	if err := os.Remove(filepath.Join(root, brief)); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	diagnostic.Reset()
	if code := runStageAnswer([]string{"consume", "--root", root, "--path", filepath.Join(worker, relative)}, &out, &diagnostic); code == 0 || !strings.Contains(diagnostic.String(), pin) {
		t.Fatalf("captured binding lost original missing pin error: %d %s %s", code, &out, &diagnostic)
	}
	if err := os.WriteFile(pinPath, pinBytes, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(worker, p.AnswerPath)); err != nil {
		t.Fatal(err)
	}
	beforePackets, err := runtime.CurrentRepairPlanReportPackets(root, partial.State, "")
	if err != nil {
		t.Fatal(err)
	}
	out.Reset()
	diagnostic.Reset()
	if code := runStageAnswer([]string{"consume", "--root", root, "--path", filepath.Join(worker, relative)}, &out, &diagnostic); code != 0 {
		t.Fatalf("display loss swallowed original responsibility: %d %s %s", code, &out, &diagnostic)
	}
	accepted := repairDispatchSnapshot(t, root)
	afterPackets, err := runtime.CurrentRepairPlanReportPackets(root, accepted.State, "")
	if err != nil || len(afterPackets) != len(beforePackets) || len(afterPackets) != 1 || afterPackets[0].ID != p.ID {
		t.Fatalf("recovery replaced original Plan responsibility: %#v %v", afterPackets, err)
	}
	registry, err = workspace.Decode(accepted.State)
	if err != nil || registry == nil || registry.Executions[p.Binding.ManifestAssignmentID].DeliveryRef == "" {
		t.Fatalf("same original packet was not received: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, brief)); !os.IsNotExist(err) {
		t.Fatalf("consume rebuilt display preparation instead of receiving original input: %v", err)
	}
}

func TestNativeS9OriginalRecoveryPreparationPreservesFrozenLegacySource(t *testing.T) {
	root, worker, p, _, _ := nativeS9WorkerFixture(t)
	service, err := submission.New(root)
	if err != nil {
		t.Fatal(err)
	}
	_, _, relative := nativeS9Paths(p)
	if _, err := consumeRegisteredStageAnswer(context.Background(), service, filepath.Join(worker, relative)); err != nil {
		t.Fatal(err)
	}
	snap := repairDispatchSnapshot(t, root)
	registry, err := workspace.Decode(snap.State)
	if err != nil {
		t.Fatal(err)
	}
	frozen := registry.Executions[p.Binding.ManifestAssignmentID]
	frozen.Purpose = ""
	frozen.BootstrapSHA256 = strings.Repeat("a", 64)
	frozen.PlatformSessionID = "original-frozen-independent-session"
	registry.Executions[p.Binding.ManifestAssignmentID] = frozen
	snap.State["workspace"] = workspace.Encode(registry)
	// Even with an original-looking source pin present, native preparation
	// evidence must never override a registered independent frozen consumer.
	if prepared, err := nativeS9RecoveryPrepared(context.Background(), root, snap.State, p); err != nil || prepared {
		t.Fatalf("frozen legacy source was not cleanly retained by its original consumer: native=%t err=%v", prepared, err)
	}
}

func TestNativeS9OriginalConsumeRejectsUnpreparedPacketWithoutGuessingNative(t *testing.T) {
	root, worker, p, _, _ := nativeS9WorkerFixture(t)
	brief, pin, relative := nativeS9Paths(p)
	if err := os.Remove(filepath.Join(root, pin)); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, brief)); err != nil {
		t.Fatal(err)
	}
	before := repairDispatchRuntimeBytes(t, root)
	var out, diagnostic bytes.Buffer
	if code := runStageAnswer([]string{"consume", "--root", root, "--path", filepath.Join(worker, relative)}, &out, &diagnostic); code == 0 {
		t.Fatalf("unknown preparation claimed as native: %s %s", &out, &diagnostic)
	}
	repairDispatchAssertUnchanged(t, root, before)
}
