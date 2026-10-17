package repair

import (
	"reflect"
	"strings"
	"testing"
)

func TestMissingRequiredTargetedReverificationIDsRequirePassAndScope(t *testing.T) {
	root := t.TempDir()
	requests := []TargetedReverificationRequest{
		{
			ReverificationID: "reverify-blocked", RuntimeID: "loop-runtime-1", BugID: "BUG-001", BaselineGeneration: 1,
			OriginalAssignmentID: "repair-assignment-1", PerformingAssignmentID: "assignment-qa",
			ContinuityReason: "test://blocked-log: verifier environment unavailable", ImpactID: "impact-1",
			AssertionResults: []AssertionResult{{AssertionID: "root-1", Result: "blocked", EvidenceRefs: []string{"test://blocked"}}},
			ScopeCompliance:  "pass", Result: "blocked", FailureClass: "blocked",
		},
		{
			ReverificationID: "reverify-failed", RuntimeID: "loop-runtime-1", BugID: "BUG-001", BaselineGeneration: 1,
			OriginalAssignmentID: "repair-assignment-1", PerformingAssignmentID: "assignment-qa",
			ContinuityReason: "test://failure-log: assertion failed", ImpactID: "impact-1",
			AssertionResults: []AssertionResult{{AssertionID: "root-1", Result: "fail", EvidenceRefs: []string{"test://failed"}}},
			ScopeCompliance:  "pass", Result: "fail", FailureClass: "fail_same_cause",
		},
	}
	invalidScope := TargetedReverificationRequest{
		ReverificationID: "reverify-scope", RuntimeID: "loop-runtime-1", BugID: "BUG-001", BaselineGeneration: 1,
		OriginalAssignmentID: "repair-assignment-1", PerformingAssignmentID: "assignment-qa",
		ContinuityReason: "test://scope-log: scope could not be confirmed", ImpactID: "impact-1",
		AssertionResults: []AssertionResult{{AssertionID: "root-1", Result: "pass", EvidenceRefs: []string{"test://root-pass"}}},
		ScopeCompliance:  "fail", Result: "pass",
	}
	if _, _, err := CreateTargetedReverification(root, invalidScope); err == nil {
		t.Fatal("passing TargetedReverification with failed scope must remain invalid under the original schema")
	}
	refs := make([]ArtifactRef, 0, len(requests))
	for _, request := range requests {
		_, ref, err := CreateTargetedReverification(root, request)
		if err != nil {
			t.Fatalf("CreateTargetedReverification(%s): %v", request.ReverificationID, err)
		}
		refs = append(refs, ref)
	}

	required := []string{"reverify-blocked", "reverify-failed"}
	missing, err := missingRequiredTargetedReverificationIDs(root, refs, required)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(missing, required) {
		t.Fatalf("missing required IDs = %#v, want %#v", missing, required)
	}

	blockedOnly := ArtifactRef{Path: refs[0].Path, SHA256: refs[0].SHA256}
	pointer := map[string]any{
		"path": "session.json", "sha256": "session-sha",
		"plan_ref": "plan.json", "plan_sha256": "plan-sha",
		"contract_ref": "contract.json", "contract_sha256": "contract-sha",
		"result_ref": "result.json", "result_sha256": "result-sha",
		"impact_ref": "impact.json", "impact_sha256": "impact-sha",
		"targeted_reverification_artifacts": []any{map[string]any{"path": blockedOnly.Path, "sha256": blockedOnly.SHA256}},
		"required_reverification_ids":       []any{"reverify-blocked"},
	}
	handoff := RepairHandoff{
		SessionRef:                 ArtifactRef{Path: "session.json", SHA256: "session-sha"},
		PlanRef:                    ArtifactRef{Path: "plan.json", SHA256: "plan-sha"},
		ContractRef:                ArtifactRef{Path: "contract.json", SHA256: "contract-sha"},
		ResultRef:                  ArtifactRef{Path: "result.json", SHA256: "result-sha"},
		ChangeImpactRef:            ArtifactRef{Path: "impact.json", SHA256: "impact-sha"},
		TargetedReverificationRefs: []ArtifactRef{blockedOnly},
	}
	if err := validateHandoffAgainstCurrentRepair(root, map[string]any{}, pointer, handoff); err == nil || !strings.Contains(err.Error(), "required_reverification_ids not committed: reverify-blocked") {
		t.Fatalf("failed-only handoff validation error = %v, want missing required reverification", err)
	}
}
