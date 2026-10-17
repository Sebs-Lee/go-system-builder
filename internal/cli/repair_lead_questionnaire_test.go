package cli

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/entroforge/go-system-builder/internal/investigation"
	"github.com/entroforge/go-system-builder/internal/policy"
	"github.com/entroforge/go-system-builder/internal/repair"
	"github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/semantic"
)

func TestS9RepairLeadStopAndIdleDoNotGuessPathWithoutAuthority(t *testing.T) {
	root := repairLeadFreshRoot(t)
	state := repairLeadBaseState(t, root, "bug_resolution", "fixing", 0)
	mapFieldCLI(state, "review")["repair"] = map[string]any{"status": "impact_reconciliation"}
	snapshot := runtime.Snapshot{State: state}
	for _, event := range []string{"Stop", "Idle"} {
		context := stageRepairLeadQuestionnaireContext(root, policy.Input{Event: event}, snapshot)
		if !strings.Contains(context, "current approved Case, RepairSession, and RepairPlan source are required") {
			t.Fatalf("%s guessed a ChangeImpact path without current source authority: %s", event, context)
		}
		if paths := repairLeadQuestionnairePaths(root, state); len(paths) != 0 {
			t.Fatalf("%s path selector invented a questionnaire path without current authority: %#v", event, paths)
		}
	}

	// S9 Handoff has no Main-authored professional fields. A ready Stop/Idle
	// reports its read-only behavior without creating a handoff or new form.
	ready := repairLeadBaseState(t, root, "bug_resolution", "ready_for_full_review", 0)
	mapFieldCLI(ready, "review")["repair"] = map[string]any{"status": "ready_for_full_review", "required_reverification_ids": []any{}}
	for _, event := range []string{"Stop", "Idle"} {
		context := stageRepairLeadQuestionnaireContext(root, policy.Input{Event: event}, runtime.Snapshot{State: ready})
		if !strings.Contains(context, "no remaining professional Main answer") || !strings.Contains(context, "read-only") {
			t.Fatalf("%s did not report read-only automatic handoff: %s", event, context)
		}
	}
	if got := repairLeadQuestionnairePaths(root, ready); len(got) != 0 {
		t.Fatalf("ready handoff has no professional answer path, got %#v", got)
	}
}

func TestS9RepairLeadPathSelectorUsesCurrentPhaseWithoutPreparing(t *testing.T) {
	root := repairLeadFreshRoot(t)
	state := repairLeadBaseState(t, root, "bug_resolution", "fixing", 0)
	mapFieldCLI(state, "review")["repair"] = map[string]any{"status": "impact_reconciliation"}
	if paths := repairLeadQuestionnairePaths(root, state); len(paths) != 0 {
		t.Fatalf("path selector guessed without current Case/Session/Plan source: %#v", paths)
	}
	ready := repairLeadBaseState(t, root, "bug_resolution", "ready_for_full_review", 0)
	mapFieldCLI(ready, "review")["repair"] = map[string]any{"status": "ready_for_full_review", "required_reverification_ids": []any{}}
	if got := repairLeadQuestionnairePaths(root, ready); len(got) != 0 {
		t.Fatalf("ready handoff should not manufacture an answer path: %#v", got)
	}

	state = repairLeadBaseState(t, root, "bug_resolution", "targeted_reverification", 0)
	if got := repairLeadQuestionnairePaths(root, state); len(got) != 0 {
		t.Fatalf("targeted path selector guessed without a current plan/impact/dispatch: %#v", got)
	}
}

func TestS9RepairLeadTargetedPathUsesRegisteredIndependentAssignment(t *testing.T) {
	root := repairLeadFreshRoot(t)
	contractRef := writeRepairLeadTestContract(t, root, "investigation-case-lead-path")
	session, sessionRef, err := repair.CreateRepairSession(root, repair.SessionRequest{
		Contract: contractRef, SessionID: "repair-session-lead-path", RuntimeID: "loop-req039-ct", ReqID: "REQ-039", BaselineGeneration: 1, CreatedBy: "Main",
	})
	if err != nil {
		t.Fatal(err)
	}
	plan, planRef, err := repair.CreateRepairPlan(root, repair.PlanRequest{Contract: contractRef, Session: sessionRef, PlanID: "repair-plan-lead-path", CreatedBy: "Main"})
	if err != nil {
		t.Fatal(err)
	}
	changedHash := strings.Repeat("a", 64)
	impact, impactRef, err := repair.CreateChangeImpact(root, repair.ChangeImpactRequest{
		ImpactID: "impact-lead-path", RuntimeID: "loop-req039-ct", ReqID: "REQ-039", BaselineGeneration: 1,
		SourceCaseIDs: []string{"investigation-case-lead-path"}, ChangeTypes: []string{"implementation"},
		ChangedArtifacts:          []repair.ArtifactRef{{ID: "changed-api", Path: "internal/api/payload.go", SHA256: changedHash}},
		Decisions:                 []repair.ImpactDecision{{SourceID: "finding-1", TargetID: "claim-1", Relation: "invalidates", RuleID: "IM-API", Decision: "reverify", Scope: []string{"internal/api/payload.go"}, Rationale: "the changed authority must be checked", RecoveryEvidence: []string{"test://lead-path"}}},
		RequiredReverificationIDs: []string{"reverify-lead-path"}, AnalyzedBy: "Main",
	})
	if err != nil {
		t.Fatal(err)
	}
	if impact.ImpactID == "" {
		t.Fatal("CreateChangeImpact returned no current Impact")
	}
	assignmentID := plan.Assignments[0].AssignmentID
	verifierID := repair.VerificationAssignmentID(assignmentID)
	state := repairLeadBaseState(t, root, "bug_resolution", "targeted_reverification", 0)
	state["runtime_id"] = "loop-req039-ct"
	review := mapFieldCLI(state, "review")
	review["repair"] = map[string]any{
		"status":  "targeted_reverification",
		"case_id": "investigation-case-lead-path", "contract_id": "repair-contract-investigation-case-lead-path",
		"contract_ref": contractRef.Path, "contract_sha256": contractRef.SHA256, "session_id": session.SessionID,
		"path": sessionRef.Path, "sha256": sessionRef.SHA256,
		"plan_ref": planRef.Path, "plan_sha256": planRef.SHA256,
		"impact_ref": impactRef.Path, "impact_sha256": impactRef.SHA256,
		"required_reverification_ids": []any{"reverify-lead-path"},
		"assignment_owners":           map[string]any{assignmentID: "builder-lead-path", verifierID: "verifier-lead-path"},
	}
	want, err := repairTargetedQuestionnairePath(root, state, verifierID)
	if err != nil {
		t.Fatal(err)
	}
	got := repairLeadQuestionnairePaths(root, state)
	if len(got) != 1 || got[0] != want {
		t.Fatalf("targeted paths = %#v, want registered verifier path %s", got, want)
	}
	if _, err := os.Stat(want); !os.IsNotExist(err) {
		t.Fatalf("read-only path selector created verifier questionnaire: %v", err)
	}
	selections, err := repairTargetedSelections(root, state)
	if err != nil || len(selections) != 1 {
		t.Fatalf("legacy single obligation source did not resolve one selection: selections=%#v err=%v", selections, err)
	}
	_, legacyDraft, err := prepareRepairTargetedQuestionnaire(root, state, selections[0])
	if err != nil || legacyDraft.ReverificationID != "reverify-lead-path" {
		t.Fatalf("legacy single ID was not preserved through the original Targeted DTO: id=%q err=%v", legacyDraft.ReverificationID, err)
	}
}

func TestS9RepairLeadImpactTargetedAndHandoffUseOriginalConsumers(t *testing.T) {
	root := repairLeadFreshRoot(t)
	// Install dispatch inputs before the Session captures their authority.
	writeRepairLeadFile(t, root, "docs/control/agent-protocol.md", "# Agent protocol\n")
	writeRepairLeadFile(t, root, "agents/test-builder.md", "# Test Builder\n")
	caseID := "investigation-case-lead-chain"
	repairLeadApprovedCase(t, root, caseID, []map[string]any{
		{"id": "unit-lead-questionnaire-one", "description": "restore the first source authority", "scope": []string{"internal/api/payload.go"}, "assertion_ids": []string{"all"}},
		{"id": "unit-lead-questionnaire-two", "description": "restore the second source authority", "scope": []string{"internal/api/other.go"}, "assertion_ids": []string{"all"}},
	})
	statePath := filepath.Join(root, ".claude", "loop-state.json")
	journalPath := filepath.Join(root, ".claude", "loop-events.jsonl")
	current, err := runtime.NewStore(statePath, journalPath).Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	opened, _, sessionRef, err := repair.OpenRepairSession(root, statePath, journalPath, repair.OpenSessionRequest{
		RuntimeRequest: repair.RuntimeRequest{ExpectedRevision: current.Revision, Actor: "Main"},
		SessionID:      "repair-session-lead-chain", CreatedBy: "Main",
	})
	if err != nil {
		t.Fatal(err)
	}
	planSnapshot, plan, planRef, err := repair.CompileRepairPlan(root, statePath, journalPath, repair.CompilePlanRequest{
		RuntimeRequest: repair.RuntimeRequest{ExpectedRevision: opened.Revision, Actor: "Main"},
		PlanID:         "repair-plan-lead-chain", CreatedBy: "Main",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Assignments) != 2 {
		t.Fatalf("multi-assignment fixture compiled %d assignments, want 2", len(plan.Assignments))
	}
	assignments := plan.Assignments
	changedPaths := []string{"internal/api/payload.go", "internal/api/other.go"}
	builders := []string{"builder-lead-chain-1", "builder-lead-chain-2"}
	expectedRevision := planSnapshot.Revision
	for index, assignment := range assignments {
		_, reportRef, err := repair.CreatePlanReport(root, repair.PlanReportRequest{
			Session: sessionRef, Plan: planRef, AssignmentID: assignment.AssignmentID,
			AgentID: builders[index], ReportID: fmt.Sprintf("repair-plan-report-lead-chain-%d", index+1),
			PlanText:      "restore the source authority",
			RedChecks:     []repair.RepairCheck{{Name: "reported failure", Command: "go test ./internal/api", Result: "fail", EvidenceRefs: []string{fmt.Sprintf("test://red-%d", index+1)}}},
			ProposedPaths: []string{changedPaths[index]},
		})
		if err != nil {
			t.Fatal(err)
		}
		if _, _, err := repair.SubmitRepairPlanReportToRuntime(root, statePath, journalPath, repair.SubmitPlanReportRequest{
			RuntimeRequest: repair.RuntimeRequest{ExpectedRevision: expectedRevision, Actor: builders[index]}, Report: reportRef,
		}); err != nil {
			t.Fatal(err)
		}
		expectedRevision++
	}
	if _, err := repair.BeginRepairExecution(root, statePath, journalPath, repair.BeginRepairExecutionRequest{RuntimeRequest: repair.RuntimeRequest{ExpectedRevision: expectedRevision, Actor: "Main"}}); err != nil {
		t.Fatal(err)
	}
	expectedRevision++
	changedBytes := [][]byte{
		[]byte("package api\n\nfunc FixedPayload() {}\n"),
		[]byte("package api\n\nfunc FixedOther() {}\n"),
	}
	for index, path := range changedPaths {
		absolute := filepath.Join(root, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(absolute), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(absolute, changedBytes[index], 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var resultSnapshot runtime.Snapshot
	var resultRef repair.ArtifactRef
	for index, assignment := range assignments {
		resultSnapshot, _, resultRef, err = repair.SubmitRepairResultToRuntime(root, statePath, journalPath, repair.SubmitResultRuntimeRequest{
			RuntimeRequest: repair.RuntimeRequest{ExpectedRevision: expectedRevision, Actor: builders[index]},
			Result: repair.RepairResultRequest{
				ResultID: fmt.Sprintf("repair-result-lead-chain-%d", index+1), AssignmentID: assignment.AssignmentID, ProducerAgentID: builders[index],
				UnitResults:      []repair.RepairUnitResult{{UnitID: assignment.UnitIDs[0], Status: "pass", EvidenceRefs: []string{fmt.Sprintf("test://repair-unit-%d", index+1)}}},
				ChangedArtifacts: []repair.ChangedArtifact{{Path: changedPaths[index], SHA256: repairLeadDigest(changedBytes[index]), Status: "added"}},
				Checks:           []repair.RepairCheck{{Name: "post-fix", Command: "go test ./...", Result: "pass", EvidenceRefs: []string{fmt.Sprintf("test://green-%d", index+1)}}}, Result: "pass",
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		expectedRevision++
	}
	repairPointer := repairPointerCLI(resultSnapshot.State)
	if got := stringValue(repairPointer["status"]); got != "impact_reconciliation" {
		t.Fatalf("RepairResult pointer status = %q, want impact_reconciliation", got)
	}
	if got := stringValue(mapFieldCLI(resultSnapshot.State, "lifecycle")["phase"]); got != "fixing" {
		t.Fatalf("RepairResult changed lifecycle phase to %q; routing must use the repair pointer status", got)
	}

	impactPath, err := repairLeadQuestionnairePath(root, resultSnapshot.State, "change-impact.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, event := range []string{"Stop", "Idle"} {
		context := stageRepairLeadQuestionnaireContext(root, policy.Input{Event: event}, resultSnapshot)
		if !strings.Contains(context, impactPath) || !strings.Contains(context, "read-only") {
			t.Fatalf("%s did not report the current sourced path as read-only: %s", event, context)
		}
		if _, err := os.Stat(impactPath); !os.IsNotExist(err) {
			t.Fatalf("%s created the ChangeImpact questionnaire: %v", event, err)
		}
	}
	impactContext := stageRepairLeadQuestionnaireContext(root, policy.Input{Event: "PreToolUse"}, resultSnapshot)
	if !strings.Contains(impactContext, impactPath) {
		t.Fatalf("Main context omitted the actual ChangeImpact path %s: %s", impactPath, impactContext)
	}
	var impactAnswer repair.ChangeImpactRequest
	readRepairLeadJSON(t, impactPath, &impactAnswer)
	wantReverificationIDs, err := repairImpactReverificationIDs(resultSnapshot.State)
	if err != nil || len(impactAnswer.RequiredReverificationIDs) != 1 || impactAnswer.RequiredReverificationIDs[0] != wantReverificationIDs[0] {
		t.Fatalf("Impact draft required IDs = %#v, want one current Contract obligation %#v (err=%v)", impactAnswer.RequiredReverificationIDs, wantReverificationIDs, err)
	}
	var feedback bytes.Buffer
	impactSave := repairLeadSaveInput(root, impactPath, "")
	if handled, _ := runSavedRepairLeadQuestionnaire(root, impactSave, &feedback, &bytes.Buffer{}); !handled || !strings.Contains(feedback.String(), "incomplete") {
		t.Fatalf("incomplete professional Impact answer was not rejected: handled=%v feedback=%s", handled, feedback.String())
	}
	unchanged, err := runtime.NewStore(statePath, journalPath).Snapshot()
	if err != nil || unchanged.Revision != resultSnapshot.Revision {
		t.Fatalf("incomplete Impact answer wrote Runtime: revision=%d want=%d err=%v", unchanged.Revision, resultSnapshot.Revision, err)
	}
	impactAnswer.ChangeTypes = []string{"implementation"}
	impactAnswer.Decisions = []repair.ImpactDecision{{
		SourceID: caseID, TargetID: "claim-lead-chain", Relation: "invalidates", RuleID: "IM-API", Decision: "reverify",
		Scope:            append([]string(nil), changedPaths...),
		Rationale:        "the changed API surfaces require independent assertion runs",
		RecoveryEvidence: []string{resultRef.Path},
	}}
	// The author leaves machine-required IDs alone; save preserves the supplied
	// Contract obligation while creating the original Impact.
	impactAnswer.RequiredReverificationIDs = nil
	writeRepairLeadJSON(t, impactPath, impactAnswer)
	feedback.Reset()
	if handled, _ := runSavedRepairLeadQuestionnaire(root, impactSave, &feedback, &bytes.Buffer{}); !handled || !strings.Contains(feedback.String(), "committed through the original Runtime Writer") {
		t.Fatalf("Impact did not reach its original Runtime consumer: handled=%v feedback=%s", handled, feedback.String())
	}
	impactSnapshot, err := runtime.NewStore(statePath, journalPath).Snapshot()
	if err != nil || stringValue(repairPointerCLI(impactSnapshot.State)["status"]) != "targeted_reverification" {
		t.Fatalf("original Impact consumer did not register Targeted obligation: status=%q err=%v", stringValue(repairPointerCLI(impactSnapshot.State)["status"]), err)
	}
	registeredRequired := stringSliceAny(repairPointerCLI(impactSnapshot.State)["required_reverification_ids"])
	if len(registeredRequired) != 1 || registeredRequired[0] != wantReverificationIDs[0] {
		t.Fatalf("Runtime registered required IDs = %#v, want one Contract obligation %#v", registeredRequired, wantReverificationIDs)
	}
	impactRevision := impactSnapshot.Revision
	feedback.Reset()
	if handled, _ := runSavedRepairLeadQuestionnaire(root, impactSave, &feedback, &bytes.Buffer{}); !handled || !strings.Contains(feedback.String(), "Replay did not write Runtime") {
		t.Fatalf("Impact replay was not recognized: handled=%v feedback=%s", handled, feedback.String())
	}
	impactReplay, err := runtime.NewStore(statePath, journalPath).Snapshot()
	if err != nil || impactReplay.Revision != impactRevision {
		t.Fatalf("Impact replay changed Runtime revision: %d -> %d err=%v", impactRevision, impactReplay.Revision, err)
	}

	verifierOwners := []string{"verifier-lead-chain-1"}
	for index, assignment := range assignments[:1] {
		var dispatchOut, dispatchErr bytes.Buffer
		if code := Run([]string{"runtime", "repair", "dispatch", "--root", root, "--assignment-id", assignment.AssignmentID, "--agent-id", verifierOwners[index], "--independent-verification"}, strings.NewReader(""), &dispatchOut, &dispatchErr); code != 0 {
			t.Fatalf("dispatch actual independent verifier %d: code=%d stdout=%s stderr=%s", index, code, dispatchOut.String(), dispatchErr.String())
		}
	}
	targetPaths := make([]string, len(verifierOwners))
	targetSaves := make([]policy.Input, len(verifierOwners))
	dispatchedSnapshot, err := runtime.NewStore(statePath, journalPath).Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	for index, assignment := range assignments[:1] {
		path, err := repairTargetedQuestionnairePath(root, dispatchedSnapshot.State, repair.VerificationAssignmentID(assignment.AssignmentID))
		if err != nil {
			t.Fatal(err)
		}
		targetPaths[index] = path
	}
	if paths := repairLeadQuestionnairePaths(root, dispatchedSnapshot.State); len(paths) != 1 {
		t.Fatalf("fresh Contract-level path selector returned %#v, want one path for the actual verifier", paths)
	} else {
		for index, path := range paths {
			if path != targetPaths[index] {
				t.Fatalf("multi-verifier path[%d] = %s, want assignment path %s", index, path, targetPaths[index])
			}
		}
	}
	selections, err := repairTargetedSelections(root, dispatchedSnapshot.State)
	if err != nil {
		t.Fatal(err)
	}
	if mappingErr := repairTargetedRequestMappingError(selections, []string{"reverify-legacy-one", "reverify-legacy-two"}); mappingErr == nil || !strings.Contains(mappingErr.Error(), "supports one TargetedReverification") {
		t.Fatalf("ambiguous old multi-ID impact was not reported as a source gap: %v", mappingErr)
	}
	if mappingErr := repairTargetedRequestMappingError(append(selections, selections[0]), wantReverificationIDs); mappingErr == nil || !strings.Contains(mappingErr.Error(), "maps to 2 independently dispatched verifier assignments") {
		t.Fatalf("ambiguous multi-verifier source was not reported: %v", mappingErr)
	}
	for index := range assignments[:1] {
		targetSnapshot, err := runtime.NewStore(statePath, journalPath).Snapshot()
		if err != nil {
			t.Fatal(err)
		}
		targetContext := stageRepairLeadQuestionnaireContext(root, policy.Input{Event: "PreToolUse", AgentID: verifierOwners[index]}, targetSnapshot)
		targetPath := targetPaths[index]
		if !strings.Contains(targetContext, targetPath) {
			t.Fatalf("dispatched verifier %s context omitted its TargetedReverification path %s: %s", verifierOwners[index], targetPath, targetContext)
		}
		for otherIndex, otherPath := range targetPaths {
			if otherIndex != index && otherPath != "" && strings.Contains(targetContext, otherPath) {
				t.Fatalf("verifier %s received another assignment's TargetedReverification path %s: %s", verifierOwners[index], otherPath, targetContext)
			}
		}
		var targetAnswer repair.TargetedReverificationRequest
		readRepairLeadJSON(t, targetPath, &targetAnswer)
		if got, want := targetAnswer.ReverificationID, wantReverificationIDs[0]; got != want {
			t.Fatalf("Targeted DTO ID = %q, want the current Contract obligation %q", got, want)
		}
		if len(targetAnswer.AssertionResults) != 3 {
			t.Fatalf("Targeted DTO has %d assertions, want all 3 current Contract assertion slots", len(targetAnswer.AssertionResults))
		}
		targetSaves[index] = repairLeadSaveInput(root, targetPath, verifierOwners[index])
		if index == 0 {
			feedback.Reset()
			if handled, _ := runSavedRepairLeadQuestionnaire(root, targetSaves[index], &feedback, &bytes.Buffer{}); !handled || !strings.Contains(feedback.String(), "must record") {
				t.Fatalf("incomplete independent result was not rejected: handled=%v feedback=%s", handled, feedback.String())
			}
			unchanged, err = runtime.NewStore(statePath, journalPath).Snapshot()
			if err != nil || unchanged.Revision != targetSnapshot.Revision {
				t.Fatalf("incomplete Targeted answer wrote Runtime: revision=%d want=%d err=%v", unchanged.Revision, targetSnapshot.Revision, err)
			}
		}
		targetAnswer.ContinuityReason = "test://blocked-run: the verifier environment was unavailable before the assertion replay"
		targetAnswer.ScopeCompliance = "pass"
		targetAnswer.Result = "blocked"
		targetAnswer.FailureClass = "blocked"
		for i := range targetAnswer.AssertionResults {
			targetAnswer.AssertionResults[i].Result = "blocked"
			targetAnswer.AssertionResults[i].EvidenceRefs = []string{"test://blocked-assertion"}
		}
		writeRepairLeadJSON(t, targetPath, targetAnswer)
		feedback.Reset()
		if handled, _ := runSavedRepairLeadQuestionnaire(root, targetSaves[index], &feedback, &bytes.Buffer{}); !handled || !strings.Contains(feedback.String(), "committed through the original Runtime Writer") {
			t.Fatalf("blocked Targeted answer for %s did not reach original Runtime consumer: handled=%v feedback=%s", verifierOwners[index], handled, feedback.String())
		}
		blockedSnapshot, err := runtime.NewStore(statePath, journalPath).Snapshot()
		blockedPointer := repairPointerCLI(blockedSnapshot.State)
		if err != nil || stringValue(blockedPointer["status"]) != "blocked" || stringValue(blockedPointer["failure_route"]) != "blocked" {
			t.Fatalf("blocked Targeted consumer status=%q route=%q err=%v", stringValue(blockedPointer["status"]), stringValue(blockedPointer["failure_route"]), err)
		}
		if paths := repairLeadQuestionnairePaths(root, blockedSnapshot.State); len(paths) != 1 || paths[0] != targetPath {
			t.Fatalf("blocked Contract responsibility did not retain its original answer path %s: %#v", targetPath, paths)
		}
		blockedContext := stageRepairLeadQuestionnaireContext(root, policy.Input{Event: "PreToolUse", AgentID: verifierOwners[index]}, blockedSnapshot)
		if !strings.Contains(blockedContext, targetPath) || !strings.Contains(blockedContext, "environmental blocker") || !strings.Contains(blockedContext, "actual continuity basis") {
			t.Fatalf("blocked verifier context omitted the same questionnaire or resume instruction: %s", blockedContext)
		}
		var persistedBlocked repair.TargetedReverificationRequest
		readRepairLeadJSON(t, targetPath, &persistedBlocked)
		blockedPerformedAt := persistedBlocked.PerformedAt
		if blockedPerformedAt.IsZero() {
			t.Fatal("blocked Targeted save did not persist its machine PerformedAt")
		}
		blockedRefs := repairArtifactRefs(blockedPointer, "targeted_reverification_artifacts", "", "")
		if len(blockedRefs) != 1 {
			t.Fatalf("blocked Runtime references = %#v, want the original failed attempt", blockedRefs)
		}
		blockedArtifactPath := filepath.Join(root, filepath.FromSlash(blockedRefs[0].Path))
		blockedBytes, err := os.ReadFile(blockedArtifactPath)
		if err != nil {
			t.Fatal(err)
		}
		for _, event := range []string{"Stop", "Idle"} {
			readOnly := stageRepairLeadQuestionnaireContext(root, policy.Input{Event: event, AgentID: verifierOwners[index]}, blockedSnapshot)
			if !strings.Contains(readOnly, targetPath) || !strings.Contains(readOnly, "read-only") {
				t.Fatalf("%s omitted the existing blocked questionnaire or read-only note: %s", event, readOnly)
			}
		}
		readOnlyAfter, err := runtime.NewStore(statePath, journalPath).Snapshot()
		if err != nil || readOnlyAfter.Revision != blockedSnapshot.Revision {
			t.Fatalf("blocked Stop/Idle context changed Runtime revision: %d -> %d err=%v", blockedSnapshot.Revision, readOnlyAfter.Revision, err)
		}
		feedback.Reset()
		if handled, _ := runSavedRepairLeadQuestionnaire(root, targetSaves[index], &feedback, &bytes.Buffer{}); !handled || !strings.Contains(feedback.String(), "already recorded") {
			t.Fatalf("replayed blocked questionnaire was not reported without resubmission: handled=%v feedback=%s", handled, feedback.String())
		}
		replayedBlocked, err := runtime.NewStore(statePath, journalPath).Snapshot()
		if err != nil || replayedBlocked.Revision != blockedSnapshot.Revision {
			t.Fatalf("blocked questionnaire replay changed Runtime revision: %d -> %d err=%v", blockedSnapshot.Revision, replayedBlocked.Revision, err)
		}
		if after, readErr := os.ReadFile(blockedArtifactPath); readErr != nil || !bytes.Equal(after, blockedBytes) {
			t.Fatalf("blocked replay changed the old failed artifact: %v", readErr)
		}

		persistedBlocked.ContinuityReason = "test://replay-log: restored the verifier environment and re-ran the approved assertions"
		persistedBlocked.Result = "pass"
		persistedBlocked.FailureClass = ""
		persistedBlocked.PerformedAt = time.Time{}
		for i := range persistedBlocked.AssertionResults {
			persistedBlocked.AssertionResults[i].Result = "pass"
			persistedBlocked.AssertionResults[i].EvidenceRefs = []string{"test://assertion-replay-pass"}
		}
		writeRepairLeadJSON(t, targetPath, persistedBlocked)
		feedback.Reset()
		if handled, _ := runSavedRepairLeadQuestionnaire(root, targetSaves[index], &feedback, &bytes.Buffer{}); !handled || !strings.Contains(feedback.String(), "committed through the original Runtime Writer") {
			t.Fatalf("resolved blocker did not resume/create/commit the pass result: handled=%v feedback=%s", handled, feedback.String())
		}
		var persistedPass repair.TargetedReverificationRequest
		readRepairLeadJSON(t, targetPath, &persistedPass)
		if persistedPass.PerformedAt.IsZero() || persistedPass.PerformedAt.Equal(blockedPerformedAt) {
			t.Fatalf("resumed Targeted save did not persist a fresh PerformedAt: blocked=%s resumed=%s", blockedPerformedAt, persistedPass.PerformedAt)
		}
		afterTarget, err := runtime.NewStore(statePath, journalPath).Snapshot()
		if err != nil || stringValue(repairPointerCLI(afterTarget.State)["status"]) != "ready_for_full_review" {
			t.Fatalf("resumed Targeted consumer did not reach handoff status: status=%q err=%v", stringValue(repairPointerCLI(afterTarget.State)["status"]), err)
		}
		resumedRefs := repairArtifactRefs(repairPointerCLI(afterTarget.State), "targeted_reverification_artifacts", "", "")
		if len(resumedRefs) != 2 {
			t.Fatalf("resume replaced or omitted the original blocked attempt: refs=%#v", resumedRefs)
		}
		foundResume := false
		for _, resumedRef := range resumedRefs {
			value, validateErr := repair.ValidateTargetedReverification(root, resumedRef)
			if validateErr != nil {
				t.Fatal(validateErr)
			}
			if resumedRef.Path == blockedRefs[0].Path {
				if value.Result != "blocked" || value.FailureClass != "blocked" {
					t.Fatalf("original failed artifact was rewritten: %#v", value)
				}
				continue
			}
			if value.Result == "pass" && value.ReverificationID == persistedPass.ReverificationID && strings.Contains(resumedRef.Path, "-resume-") {
				foundResume = true
			}
		}
		if !foundResume {
			t.Fatalf("Runtime did not retain a new same-ID resumed PASS artifact: %#v", resumedRefs)
		}
		if after, readErr := os.ReadFile(blockedArtifactPath); readErr != nil || !bytes.Equal(after, blockedBytes) {
			t.Fatalf("resumed attempt changed the old failed artifact: %v", readErr)
		}
		remaining := repairLeadQuestionnairePaths(root, afterTarget.State)
		if len(remaining) != 0 {
			t.Fatalf("completed Contract responsibility remained outstanding: paths=%#v", remaining)
		}
		completedContext := stageRepairLeadQuestionnaireContext(root, policy.Input{Event: "PreToolUse", AgentID: verifierOwners[0]}, afterTarget)
		if strings.Contains(completedContext, targetPaths[0]) {
			t.Fatalf("completed verifier still received its answer path: %s", completedContext)
		}
		feedback.Reset()
		if handled, _ := runSavedRepairLeadQuestionnaire(root, targetSaves[index], &feedback, &bytes.Buffer{}); !handled || !strings.Contains(feedback.String(), "Replay did not write Runtime") {
			t.Fatalf("Targeted replay for %s was not recognized: handled=%v feedback=%s", verifierOwners[index], handled, feedback.String())
		}
		replayedTarget, err := runtime.NewStore(statePath, journalPath).Snapshot()
		if err != nil || replayedTarget.Revision != afterTarget.Revision {
			t.Fatalf("Targeted replay for %s changed Runtime revision: %d -> %d err=%v", verifierOwners[index], afterTarget.Revision, replayedTarget.Revision, err)
		}
	}
	readySnapshot, err := runtime.NewStore(statePath, journalPath).Snapshot()
	if err != nil || stringValue(repairPointerCLI(readySnapshot.State)["status"]) != "ready_for_full_review" {
		t.Fatalf("all independent verifiers did not reach handoff status: status=%q err=%v", stringValue(repairPointerCLI(readySnapshot.State)["status"]), err)
	}
	targetRevision := readySnapshot.Revision
	for _, event := range []string{"Stop", "Idle"} {
		context := stageRepairLeadQuestionnaireContext(root, policy.Input{Event: event}, readySnapshot)
		if !strings.Contains(context, "no remaining professional Main answer") || !strings.Contains(context, "read-only") {
			t.Fatalf("%s did not report read-only automatic Handoff behavior: %s", event, context)
		}
	}
	if _, err := os.Stat(filepath.Join(root, ".claude", "review", "repair", "handoffs")); !os.IsNotExist(err) {
		t.Fatalf("Stop/Idle created a Handoff artifact directory: %v", err)
	}
	readOnlySnapshot, err := runtime.NewStore(statePath, journalPath).Snapshot()
	if err != nil || readOnlySnapshot.Revision != targetRevision || stringValue(repairPointerCLI(readOnlySnapshot.State)["handoff_ref"]) != "" {
		t.Fatalf("Stop/Idle changed Runtime before active Main event: revision=%d handoff=%q err=%v", readOnlySnapshot.Revision, stringValue(repairPointerCLI(readOnlySnapshot.State)["handoff_ref"]), err)
	}

	handoffContext := stageRepairLeadQuestionnaireContext(root, policy.Input{Event: "PreToolUse"}, readySnapshot)
	handoffSnapshot, err := runtime.NewStore(statePath, journalPath).Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	handoffPointer := repairPointerCLI(handoffSnapshot.State)
	handoffRef, err := repairPointerRef(handoffPointer, "handoff_ref", "handoff_sha256", "committed RepairHandoff")
	if err != nil {
		t.Fatalf("normal Main event did not automatically commit Handoff: %v; context=%s", err, handoffContext)
	}
	handoffPath, err := absoluteQuestionnairePath(root, handoffRef.Path)
	if err != nil || !strings.Contains(handoffContext, handoffPath) {
		t.Fatalf("handoff context omitted actual artifact path %s: %s err=%v", handoffPath, handoffContext, err)
	}
	seedPath, err := absoluteQuestionnairePath(root, stringValue(handoffPointer["review_plan_seed_ref"]))
	if err != nil || !strings.Contains(handoffContext, seedPath) {
		t.Fatalf("handoff context omitted actual S7 seed path %s: %s err=%v", seedPath, handoffContext, err)
	}
	if got := stringValue(mapFieldCLI(handoffSnapshot.State, "lifecycle")["state"]); got != "verification" {
		t.Fatalf("original Handoff consumer did not start S7, lifecycle state=%q", got)
	}
	if _, err := os.Stat(seedPath); err != nil {
		t.Fatalf("registered S7 seed is missing at its reported path: %v", err)
	}
	committedAgain, replayed, err := completeRepairHandoff(root)
	if err != nil || committedAgain.Path != handoffRef.Path || replayed.Revision != handoffSnapshot.Revision {
		t.Fatalf("repeated automatic Handoff changed the result: ref=%#v revision=%d want=%d err=%v", committedAgain, replayed.Revision, handoffSnapshot.Revision, err)
	}
}

func repairLeadSaveInput(root, path, agentID string) policy.Input {
	return policy.Input{Event: "PostToolUse", AgentID: agentID, CWD: root, ToolName: "Write", ToolInput: map[string]any{"file_path": path}}
}

func readRepairLeadJSON(t *testing.T, path string, destination any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, destination); err != nil {
		t.Fatal(err)
	}
}

func writeRepairLeadJSON(t *testing.T, path string, value any) {
	t.Helper()
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeRepairLeadFile(t *testing.T, root, relative, value string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(value), 0o644); err != nil {
		t.Fatal(err)
	}
}

func repairLeadDigest(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

func repairLeadFreshRoot(t *testing.T) string {
	t.Helper()
	root := initFormatRoot(t)
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"init", "--root", root}, strings.NewReader(""), &stdout, &stderr); code != 0 {
		t.Fatalf("initialize local S9 fixture: code=%d stderr=%s stdout=%s", code, stderr.String(), stdout.String())
	}
	return root
}

func repairLeadBaseState(t *testing.T, root, lifecycleState, phase string, revision int) map[string]any {
	t.Helper()
	reqPath := "docs/requirements/REQ-039-loop-control-plane.md"
	reqBytes := []byte("# REQ-039\n> Status: locked\n> Version: v2.0.0\n")
	writeRepairLeadFile(t, root, reqPath, string(reqBytes))
	state, err := inactiveRuntimeState(root, time.Now().UTC())
	if err != nil {
		t.Fatalf("build local S9 Runtime from current initialization facts: %v", err)
	}
	state["runtime_id"] = "loop-req039-ct"
	state["revision"] = float64(revision)
	state["lifecycle"] = map[string]any{
		"state": lifecycleState, "phase": repairLeadPhaseValue(phase), "phase_revision": 0,
	}
	milestone := mapFieldCLI(state, "milestone")
	milestone["stage"] = "S8"
	milestone["lifecycle_state"] = lifecycleState
	milestone["lifecycle_phase"] = repairLeadPhaseValue(phase)
	milestone["source_revision"] = float64(revision)
	state["authorization"] = map[string]any{
		"mode": "binding", "command": "loop-harness req bind", "actor": "user",
		"occurred_at": "2026-10-01T00:00:00Z",
	}
	state["bound_req"] = map[string]any{
		"id": "REQ-039", "path": reqPath, "version": "v2.0.0", "sha256": repairLeadDigest(reqBytes),
		"status": "locked", "approved_by": "user", "approved_at": "2026-10-01T00:00:00Z",
		"metadata": map[string]any{"ui_impact": "none"},
	}
	state["baseline"] = map[string]any{"generation": 1, "captured_at": "2026-10-01T00:00:00Z"}
	state["review"] = map[string]any{"round": 1, "clean_round": nil}
	return state
}

func repairLeadPhaseValue(phase string) any {
	if phase == "" {
		return nil
	}
	return phase
}

func repairLeadWriteState(t *testing.T, root string, state map[string]any) {
	t.Helper()
	bound, _ := state["bound_req"].(map[string]any)
	if bound == nil {
		t.Fatal("local S9 fixture is missing its locked REQ binding")
	}
	head := repairLeadCommitFixture(t, root)
	bound["workspace"] = map[string]any{
		"project_root": root, "dev_branch": "test-development",
		"release_upstream": "origin/release", "bound_commit": head,
	}
	putInitMap(t, filepath.Join(root, ".claude", "loop-state.json"), state)
}

func repairLeadCommitFixture(t *testing.T, root string) string {
	t.Helper()
	if _, err := os.Stat(filepath.Join(root, ".git")); os.IsNotExist(err) {
		repairLeadGit(t, root, "init", "-q", "-b", "test-development")
	} else if err != nil {
		t.Fatalf("inspect local S9 fixture Git repository: %v", err)
	}
	repairLeadGit(t, root, "config", "--local", "user.name", "S9 Fixture")
	repairLeadGit(t, root, "config", "--local", "user.email", "s9-fixture@example.invalid")
	repairLeadGit(t, root, "config", "--local", "commit.gpgsign", "false")
	info := filepath.Join(root, ".git", "info")
	if err := os.MkdirAll(info, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(info, "exclude"), []byte(".claude/\n.worktrees/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	repairLeadGit(t, root, "add", "-A")
	repairLeadGit(t, root, "commit", "--allow-empty", "-qm", "S9 test fixture")
	repairLeadGit(t, root, "branch", "-f", "origin/release", "HEAD")
	return repairLeadGit(t, root, "rev-parse", "--verify", "refs/heads/test-development^{commit}")
}

func repairLeadGit(t *testing.T, root string, args ...string) string {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", root}, args...)...)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, output)
	}
	return strings.TrimSpace(string(output))
}

func TestS9RepairReadbackUsesOriginalSessionAndPlanWriters(t *testing.T) {
	root := repairLeadFreshRoot(t)
	state := repairLeadApprovedCase(t, root, "investigation-case-lead-readback", []map[string]any{{
		"id": "unit-lead-questionnaire", "description": "restore the source authority",
		"scope": []string{"internal/api"}, "assertion_ids": []string{"all"},
	}})

	context := stageRepairLeadQuestionnaireContext(root, policy.Input{Event: "PreToolUse"}, runtime.Snapshot{State: state})
	prepared, err := runtime.NewStore(filepath.Join(root, ".claude/loop-state.json"), filepath.Join(root, ".claude/loop-events.jsonl")).Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	pointer := repairPointerCLI(prepared.State)
	for _, key := range []string{"contract_ref", "path", "plan_ref"} {
		refPath := stringValue(pointer[key])
		if refPath == "" {
			t.Fatalf("original writer did not register %s: %#v", key, pointer)
		}
		absPath, err := absoluteQuestionnairePath(root, refPath)
		if err != nil || !strings.Contains(context, absPath) {
			t.Fatalf("readback context omitted exact %s path %s: %s (err=%v)", key, absPath, context, err)
		}
	}
	if stringValue(pointer["status"]) != "planning" || !strings.Contains(context, "Contract-derived RepairPlan") {
		t.Fatalf("readback did not use the original Session/Plan APIs: pointer=%#v context=%s", pointer, context)
	}
	_, err = prepareRepairReadback(root, prepared.State)
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := runtime.NewStore(filepath.Join(root, ".claude/loop-state.json"), filepath.Join(root, ".claude/loop-events.jsonl")).Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if replayed.Revision != prepared.Revision {
		t.Fatalf("repeated readback preparation wrote Runtime again: revision %d -> %d", prepared.Revision, replayed.Revision)
	}
}

func TestS9RepairReadbackIDsIncludeCurrentCaseAndContract(t *testing.T) {
	root := repairLeadFreshRoot(t)
	caseA, caseB := "investigation-case-lead-readback-a", "investigation-case-lead-readback-b"
	stateA := repairLeadApprovedCase(t, root, caseA, []map[string]any{{
		"id": "unit-lead-questionnaire-a", "description": "restore Case A source authority",
		"scope": []string{"internal/api"}, "assertion_ids": []string{"all"},
	}})
	first, err := prepareRepairReadback(root, stateA)
	if err != nil {
		t.Fatal(err)
	}
	firstPointer := repairPointerCLI(first)
	firstPlan, err := repair.ValidateRepairPlan(root, repair.ArtifactRef{Path: stringValue(firstPointer["plan_ref"]), SHA256: stringValue(firstPointer["plan_sha256"])})
	if err != nil {
		t.Fatal(err)
	}
	firstQuestionnairePath, err := repairLeadQuestionnairePath(root, first, "change-impact.json")
	if err != nil {
		t.Fatal(err)
	}

	// Seed the same Runtime and generation from a different approved Case while
	// leaving Case A's immutable Session/Plan artifacts in the shared root.
	stateB := repairLeadApprovedCase(t, root, caseB, []map[string]any{{
		"id": "unit-lead-questionnaire-b", "description": "restore Case B source authority",
		"scope": []string{"internal/api"}, "assertion_ids": []string{"all"},
	}})
	second, err := prepareRepairReadback(root, stateB)
	if err != nil {
		t.Fatal(err)
	}
	secondPointer := repairPointerCLI(second)
	secondPlan, err := repair.ValidateRepairPlan(root, repair.ArtifactRef{Path: stringValue(secondPointer["plan_ref"]), SHA256: stringValue(secondPointer["plan_sha256"])})
	if err != nil {
		t.Fatal(err)
	}
	secondQuestionnairePath, err := repairLeadQuestionnairePath(root, second, "change-impact.json")
	if err != nil {
		t.Fatal(err)
	}
	if firstPointer["session_id"] == secondPointer["session_id"] || firstPlan.PlanID == secondPlan.PlanID {
		t.Fatalf("different Cases reused immutable S9 artifacts at identical Runtime/generation: session %q/%q plan %q/%q", firstPointer["session_id"], secondPointer["session_id"], firstPlan.PlanID, secondPlan.PlanID)
	}
	if firstPointer["contract_ref"] == secondPointer["contract_ref"] {
		t.Fatalf("different Cases reused the same approved Contract: %q", firstPointer["contract_ref"])
	}
	if firstQuestionnairePath == secondQuestionnairePath {
		t.Fatalf("different Case/Plan sources reused the same same-round questionnaire path %s", firstQuestionnairePath)
	}
	if firstPointer["case_id"] != caseA || secondPointer["case_id"] != caseB {
		t.Fatalf("readback did not follow current approved Case: first=%q second=%q", firstPointer["case_id"], secondPointer["case_id"])
	}
}

// repairLeadApprovedCase creates the S9 source through the same S7->S8 and
// Case/Contract writers used by the product. The ObservationBatch and Finding
// are minimal schema-valid fixture facts; the InvestigationCase is never
// hand-pinned into Runtime.
func repairLeadApprovedCase(t *testing.T, root, caseID string, units []map[string]any) map[string]any {
	t.Helper()
	state := repairLeadBaseState(t, root, "bug_resolution", "investigation", 0)
	token := dispatchSlug(caseID)
	findingID := "finding-" + token
	batchID := "observation-batch-" + token
	occurredAt := "2026-10-01T00:00:00Z"
	finding := map[string]any{
		"schema_version": "1.0.0", "finding_id": findingID, "claim_id": "claim-" + token,
		"lens": "qa", "severity": "P1", "expected": "the expected behavior holds",
		"authority_refs": []string{"REQ-039"}, "observed": "the observed behavior deviates",
		"observation_mode": "code_inspection",
		"encounter": map[string]any{
			"journey_summary":  "inspect the bound implementation and observe the deviation",
			"inspection_entry": "internal/api", "symbol_trail": "entry -> boundary",
			"wall_action": "inspect the boundary", "first_bad_checkpoint": "the invariant is false",
			"terminal_state": "review stopped",
		},
		"reproducibility": "always", "evidence_refs": []string{"test://repair-lead-finding"},
	}
	findingBytes, err := json.MarshalIndent(finding, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	findingBytes = append(findingBytes, '\n')
	findingRel := filepath.ToSlash(filepath.Join(".claude", "evidence", "findings", findingID+".json"))
	writeRepairLeadFile(t, root, findingRel, string(findingBytes))
	mapFieldCLI(state, "entities")["findings"] = []any{map[string]any{
		"finding_id": findingID, "path": findingRel, "sha256": repairLeadDigest(findingBytes),
		"claim_id": finding["claim_id"], "assignment_id": "assignment-s8-fixture", "lens": "qa",
		"severity": "P1", "observation_mode": "code_inspection", "original_finder": "reviewer-fixture",
		"review_round": 1, "created_at": occurredAt,
	}}
	batch := map[string]any{
		"schema_version": "1.0.0", "observation_batch_id": batchID, "conclusion": "sealed",
		"evidence_id": batchID, "kind": "observation_batch", "runtime_id": "loop-req039-ct",
		"producer_agent_id": "round-consumer", "producer_responsibility": "Orchestrator",
		"review_plan_id": "review-plan-repair-lead-fixture", "review_round": 1,
		"baseline_generation": 1, "subject_digest": strings.Repeat("a", 64),
		"finding_ids": []string{findingID}, "drain_policy": "complete_required_claims",
		"claim_coverage_summary": map[string]any{
			"total_required": 1, "pass": 0, "finding": 1, "not_applicable": 0,
			"blocked": 0, "blocked_claims": []any{}, "plan_revision": 1,
		},
		"unobserved_claim_ids": []any{}, "original_finder_routes": []any{},
		"investigation_readiness": []any{map[string]any{"finding_id": findingID, "status": "ready"}},
		"sealed_at":               occurredAt, "sealed_by": "round-consumer", "revision": 1,
	}
	batchBytes, err := json.MarshalIndent(batch, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	batchBytes = append(batchBytes, '\n')
	batchRel := filepath.ToSlash(filepath.Join(".claude", "evidence", batchID+".json"))
	writeRepairLeadFile(t, root, batchRel, string(batchBytes))
	mapFieldCLI(state, "review")["observation_batch"] = map[string]any{
		"batch_id": batchID, "path": batchRel, "sha256": repairLeadDigest(batchBytes),
		"finding_ids": []string{findingID}, "drain_policy": "complete_required_claims", "sealed_at": occurredAt,
	}
	repairLeadWriteState(t, root, state)
	statePath := filepath.Join(root, ".claude", "loop-state.json")
	journalPath := filepath.Join(root, ".claude", "loop-events.jsonl")
	if err := os.WriteFile(journalPath, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	ingested, err := investigation.Ingest(root, statePath, journalPath, investigation.IngestRequest{
		ExpectedRevision: 0, CaseID: caseID, GroupingRationale: "the sealed fixture Finding is a bounded repair Case",
		OccurredAt: time.Date(2026, 10, 1, 0, 0, 1, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("original InvestigationCase Ingest writer: %v", err)
	}
	caseRevision := func(state map[string]any) int {
		pointer := mapFieldCLI(mapFieldCLI(state, "review"), "investigation")
		switch value := pointer["revision"].(type) {
		case float64:
			return int(value)
		case int:
			return value
		default:
			t.Fatalf("active InvestigationCase has no numeric revision: %#v", pointer["revision"])
			return 0
		}
	}
	pointer := mapFieldCLI(mapFieldCLI(ingested.State, "review"), "investigation")
	assignmentID := "assignment-hypothesis-" + token
	hypothesisID := "hypothesis-" + token
	registered, err := investigation.RegisterHypothesis(root, statePath, journalPath, investigation.HypothesisRequest{
		ExpectedRevision: ingested.Revision, ExpectedCaseRevision: caseRevision(ingested.State), ExpectedCaseSHA256: stringValue(pointer["sha256"]),
		CaseID: caseID, HypothesisID: hypothesisID, AssignmentID: assignmentID,
		Statement:        "the source authority is duplicated at the reported boundary",
		Invariant:        "the boundary has one authoritative source",
		Discriminator:    "trace the observed value across the source boundary",
		ExpectedOutcomes: map[string]any{"support": "the source differs at the boundary", "refute": "the source is consistent at the boundary"},
		SourceFindingIDs: []string{findingID}, EvidenceRefs: []string{"test://repair-lead-finding"},
		OccurredAt: time.Date(2026, 10, 1, 0, 0, 2, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("original hypothesis registration writer: %v", err)
	}
	pointer = mapFieldCLI(mapFieldCLI(registered.State, "review"), "investigation")
	result, err := investigation.SubmitHypothesisResult(root, statePath, journalPath, investigation.HypothesisResultRequest{
		ExpectedRevision: registered.Revision, ExpectedCaseRevision: caseRevision(registered.State), ExpectedCaseSHA256: stringValue(pointer["sha256"]),
		CaseID: caseID, HypothesisID: hypothesisID, AssignmentID: assignmentID,
		Method: "read-only inspection of the cited source boundary", EvidenceRefs: []string{"test://repair-lead-finding"},
		SourceBoundaryRefs: []string{"finding-source-boundary"},
		Observed:           "the observed value crosses two conflicting source definitions",
		Counterfactual:     "one authoritative source would preserve the same value across the boundary",
		Result:             "supported",
		ExplainsFindingIDs: []string{findingID},
		DoesNotExplain:     []string{},
		OccurredAt:         time.Date(2026, 10, 1, 0, 0, 3, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("original HypothesisResult writer: %v", err)
	}
	pointer = mapFieldCLI(mapFieldCLI(result.State, "review"), "investigation")
	routed, err := investigation.UpdateCaseRoute(root, statePath, journalPath, investigation.RouteRequest{
		ExpectedRevision: result.Revision, ExpectedCaseRevision: caseRevision(result.State), ExpectedCaseSHA256: stringValue(pointer["sha256"]),
		CaseID: caseID, Route: "s9_repair", RouteReason: "the fixture requires a bounded implementation repair",
		PrimaryRootCause: "the source authority is duplicated",
		CausalModel: map[string]any{
			"trigger": "payload crosses the boundary", "violated_invariant": "one owner",
			"faulty_mechanism": "duplicate authority", "propagation": "the boundary diverges",
			"symptoms": []string{"the Finding is reproduced"},
		},
		NoCompetingHypothesis: "the sealed observation supports no credible competing mechanism",
		BlastRadius:           map[string]any{"paths": []string{"internal/api"}},
		DetectionGap:          map[string]any{"gap_type": "contract", "evidence_refs": []string{"test://repair-lead-detection-gap"}},
		OccurredAt:            time.Date(2026, 10, 1, 0, 0, 4, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("original InvestigationCase route writer: %v", err)
	}
	approvedCaseRevision := caseRevision(routed.State)
	contractID := "repair-contract-" + token
	draftRel := filepath.ToSlash(filepath.Join(".claude", "review", "investigation", "contracts", "drafts", contractID+".json"))
	draft := map[string]any{
		"schema_version": "1.0.0", "repair_contract_id": contractID, "case_id": caseID,
		"revision": approvedCaseRevision, "status": "draft", "source_finding_ids": []string{findingID},
		"root_cause_statement": "the source authority is duplicated", "violated_invariant": "one owner",
		"causal_model_ref": "case://" + caseID + "/causal-model", "architecture_intent": "restore one source authority",
		"repair_units": units, "prospective_scope": []string{"internal/api"},
		"forbidden_scope":            []string{"docs/requirements"},
		"symptom_assertions":         []string{"the observed behavior is corrected"},
		"root_invariant_assertions":  []string{"one authority remains"},
		"detection_gap_assertions":   []string{"the boundary test detects drift"},
		"stop_escalation_conditions": []string{"the approved scope expands"},
	}
	draftBytes, err := json.MarshalIndent(draft, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	draftBytes = append(draftBytes, '\n')
	writeRepairLeadFile(t, root, draftRel, string(draftBytes))
	approvalHash := repairLeadDigest(draftBytes)
	decisionID := "ev-contract-approval-" + token
	decisionRel := filepath.ToSlash(filepath.Join(".claude", "decisions", decisionID+".json"))
	decisionBytes, err := json.MarshalIndent(map[string]any{
		"decision": "approve_contract", "decision_id": decisionID, "runtime_id": "loop-req039-ct",
		"case_id": caseID, "contract_id": contractID, "approved_by": "main-session", "approval_hash": approvalHash,
	}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	decisionBytes = append(decisionBytes, '\n')
	writeRepairLeadFile(t, root, decisionRel, string(decisionBytes))
	evidenceSnapshot, err := runtime.RecordEvidence(root, statePath, journalPath, runtime.EvidenceRequest{
		ExpectedRevision: routed.Revision, ID: decisionID, Kind: "human_decision", Path: decisionRel,
		ProducedBy: []string{"main-session"}, ScopeRefs: []string{"s8_contract_approval:loop-req039-ct"},
		Validator: semantic.RuntimeCandidateValidator{},
	})
	if err != nil {
		t.Fatalf("original human approval evidence writer: %v", err)
	}
	approved, err := investigation.ApproveContract(root, statePath, journalPath, investigation.ContractRequest{
		ExpectedRevision: evidenceSnapshot.Revision, CaseID: caseID, ContractPath: draftRel,
		ApprovedBy: "main-session", ApprovalHash: approvalHash, ApprovalEvidenceID: decisionID,
		OccurredAt: time.Date(2026, 10, 1, 0, 0, 5, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("original RepairContract approval writer: %v", err)
	}
	return approved.State
}

func writeRepairLeadTestContract(t *testing.T, root, caseID string) repair.ContractRef {
	return writeRepairLeadContractWithUnits(t, root, caseID, []map[string]any{{
		"id": "unit-lead-questionnaire", "description": "restore the source authority",
		"scope": []string{"internal/api"}, "assertion_ids": []string{"all"},
	}})
}

func writeRepairLeadContractWithUnits(t *testing.T, root, caseID string, units []map[string]any) repair.ContractRef {
	t.Helper()
	rel := filepath.ToSlash(filepath.Join(".claude", "review", "investigation", "contracts", "repair-contract-"+dispatchSlug(caseID)+".json"))
	value := map[string]any{
		"schema_version": "1.0.0", "repair_contract_id": "repair-contract-" + dispatchSlug(caseID),
		"case_id": caseID, "revision": 1, "status": "approved",
		"source_finding_ids": []string{"finding-lead-readback"}, "root_cause_statement": "two authorities drift",
		"violated_invariant": "one authority", "causal_model_ref": "case://lead-readback/model",
		"architecture_intent": "restore one authority",
		"repair_units":        units,
		"prospective_scope":   []string{"internal/api"}, "forbidden_scope": []string{"docs/requirements"},
		"symptom_assertions":         []string{"the reported behavior is corrected"},
		"root_invariant_assertions":  []string{"one authority remains"},
		"detection_gap_assertions":   []string{"the existing test catches drift"},
		"stop_escalation_conditions": []string{"the approved scope expands"},
		"approved_by":                "human", "approved_at": "2026-10-01T00:00:00Z", "approval_hash": strings.Repeat("a", 64),
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	data = append(data, '\n')
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(data)
	return repair.ContractRef{Path: rel, SHA256: hex.EncodeToString(digest[:])}
}
