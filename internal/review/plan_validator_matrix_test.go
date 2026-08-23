package review

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// §14.1 plan-validator matrix: the coverage rules that gate registration
// beyond the exact-set partition checks already covered by review_test.go.
// ---------------------------------------------------------------------------

// loadFixturePlan returns a fresh decoded copy of the shared fixture plan.
func loadFixturePlan(t *testing.T) *Plan {
	t.Helper()
	root := t.TempDir()
	data, err := os.ReadFile(writePlanFile(t, root))
	if err != nil {
		t.Fatal(err)
	}
	var plan Plan
	if err := json.Unmarshal(data, &plan); err != nil {
		t.Fatal(err)
	}
	return &plan
}

func dropClaims(plan *Plan, ids ...string) {
	drop := map[string]bool{}
	for _, id := range ids {
		drop[id] = true
	}
	kept := plan.Claims[:0]
	for _, claim := range plan.Claims {
		if !drop[claim.ClaimID] {
			kept = append(kept, claim)
		}
	}
	plan.Claims = kept
	assignments := plan.Assignments[:0]
	for _, assignment := range plan.Assignments {
		claimIDs := assignment.ClaimIDs[:0]
		for _, claimID := range assignment.ClaimIDs {
			if !drop[claimID] {
				claimIDs = append(claimIDs, claimID)
			}
		}
		assignment.ClaimIDs = claimIDs
		assignments = append(assignments, assignment)
	}
	plan.Assignments = assignments
}

// §14.1: Main 临时把 capacity policy 改成 bounded_flow —— revision/schema
// gate 拒绝；S7 policy 固定为 coverage_complete。
func TestValidatePlanRejectsNonCoverageCompletePolicy(t *testing.T) {
	plan := loadFixturePlan(t)
	plan.DispatchCapacityPolicy = "bounded_flow"
	if err := ValidatePlan(plan); err == nil || !strings.Contains(err.Error(), "coverage_complete") {
		t.Fatalf("bounded_flow policy must be rejected, got %v", err)
	}
}

// §14.1: 有产品实现变化却 DV 或 QA 为 N=0 —— 拒绝；纯文档例外需
// coverage_justification 证明。
func TestValidatePlanRejectsZeroClaimLensWithoutJustification(t *testing.T) {
	// delivery N=0.
	plan := loadFixturePlan(t)
	dropClaims(plan, "claim-dv-1")
	if err := ValidatePlan(plan); err == nil || !strings.Contains(err.Error(), "zero required delivery Claims") {
		t.Fatalf("zero DV claims must be rejected, got %v", err)
	}
	// Docs-only exception: a non-empty coverage_justification legalizes it.
	plan = loadFixturePlan(t)
	dropClaims(plan, "claim-dv-1")
	justification := "pure docs change: TASK-101 touches only docs/reports; no product behavior moved (source: change_impact)"
	plan.CoverageJustification = &justification
	if err := ValidatePlan(plan); err != nil {
		t.Fatalf("justified docs-only DV N=0 must pass, got %v", err)
	}
	// qa N=0.
	plan = loadFixturePlan(t)
	dropClaims(plan, "claim-qa-1", "claim-qa-2")
	if err := ValidatePlan(plan); err == nil || !strings.Contains(err.Error(), "zero required qa Claims") {
		t.Fatalf("zero QA claims must be rejected, got %v", err)
	}
	// An empty/whitespace justification is not a justification.
	plan = loadFixturePlan(t)
	dropClaims(plan, "claim-qa-1", "claim-qa-2")
	blank := "   "
	plan.CoverageJustification = &blank
	if err := ValidatePlan(plan); err == nil {
		t.Fatal("blank coverage_justification must not legalize a zero-claim lens")
	}
}

// §14.1: E2E cold_start —— 必须声明隔离写面且至少有一个 required e2e
// Claim；blank matrix 不得压成一个 generic Agent。
func TestValidatePlanColdStartRequirements(t *testing.T) {
	// cold_start without a workspace is rejected.
	plan := loadFixturePlan(t)
	plan.E2ECoverageState = "cold_start"
	if err := ValidatePlan(plan); err == nil || !strings.Contains(err.Error(), "verification_artifact_workspace") {
		t.Fatalf("cold_start without workspace must be rejected, got %v", err)
	}
	// Workspace declared but every e2e claim is N/A: cold start still
	// requires at least one required e2e Claim.
	plan = loadFixturePlan(t)
	plan.E2ECoverageState = "cold_start"
	workspace := "e2e-workspace/plan-t-1"
	plan.VerificationArtifactWorkspace = &workspace
	if err := ValidatePlan(plan); err == nil || !strings.Contains(err.Error(), "at least one required e2e Claim") {
		t.Fatalf("cold_start without a required e2e claim must be rejected, got %v", err)
	}
	// Promote the e2e claim to required and dispatch it: the matrix closes.
	plan = loadFixturePlan(t)
	plan.E2ECoverageState = "cold_start"
	plan.VerificationArtifactWorkspace = &workspace
	for i := range plan.Claims {
		if plan.Claims[i].ClaimID == "claim-e2e-na" {
			plan.Claims[i].Applicability = "required"
			plan.Claims[i].NARationale = ""
			plan.Claims[i].Target = "user login flow"
		}
	}
	plan.Assignments = append(plan.Assignments, PlanAssignment{
		AssignmentID: "assignment-e2e-1", Lens: "e2e", ClaimIDs: []string{"claim-e2e-na"},
		NonOverlapBoundary: "owns the declared flows", ExecutionWave: "behavior",
	})
	if err := ValidatePlan(plan); err != nil {
		t.Fatalf("closed cold-start matrix must pass, got %v", err)
	}
}

// §14.1: E2E 不适用 —— 必须有 impact/source/rationale；没有任何 N/A e2e
// Claim 的 not_applicable 结论被拒绝。
func TestValidatePlanNotApplicableRequiresNAClaim(t *testing.T) {
	plan := loadFixturePlan(t)
	dropClaims(plan, "claim-e2e-na")
	if err := ValidatePlan(plan); err == nil || !strings.Contains(err.Error(), "applicability=not_applicable") {
		t.Fatalf("not_applicable without an N/A e2e claim must be rejected, got %v", err)
	}
}

// §14.1: ReviewPlan 漏掉 Closing Contract obligation —— 每个
// current-generation TASK 必须出现在至少一个 Claim 的 source_refs，
// 缺口在注册时拒绝并指出缺失 source。
func TestRegisterPlanRejectsTaskCoverageGap(t *testing.T) {
	root := t.TempDir()
	state := baseVerificationState()
	state["documents"] = []any{
		map[string]any{
			"id": "TASK-101", "kind": "task", "path": "docs/tasks/TASK-101.md",
			"version": "v1.0.0", "sha256": strings.Repeat("2", 64),
			"status": "locked", "generation": 1,
		},
	}
	statePath, journalPath := writeState(t, root, state)

	_, err := RegisterPlan(root, statePath, journalPath, PlanRequest{
		ExpectedRevision: 1, PlanPath: writePlanFile(t, root),
	})
	if err == nil || !strings.Contains(err.Error(), "TASK-101") {
		t.Fatalf("plan dropping TASK-101 must be rejected with the missing source named, got %v", err)
	}

	// Add TASK-101 to a claim's source_refs: the coverage diff closes and
	// registration succeeds.
	planPath := writePlanFile(t, root)
	data, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatal(err)
	}
	var body map[string]any
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatal(err)
	}
	for _, raw := range body["claims"].([]any) {
		claim := raw.(map[string]any)
		if claim["claim_id"] == "claim-dv-1" {
			claim["source_refs"] = append(claim["source_refs"].([]any), "TASK-101")
		}
	}
	out, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(planPath, append(out, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	snap, err := RegisterPlan(root, statePath, journalPath, PlanRequest{
		ExpectedRevision: 1, PlanPath: planPath,
	})
	if err != nil {
		t.Fatalf("plan covering TASK-101 must register, got %v", err)
	}
	if ptr := PlanPointerFromState(snap.State); ptr == nil || ptr.Status != "running" {
		t.Fatalf("plan pointer wrong after coverage fix: %+v", ptr)
	}
}

// TASKs from earlier generations do not constrain the current round's plan.
func TestRegisterPlanIgnoresPriorGenerationTasks(t *testing.T) {
	root := t.TempDir()
	state := baseVerificationState()
	state["documents"] = []any{
		map[string]any{
			"id": "TASK-099", "kind": "task", "path": "docs/tasks/TASK-099.md",
			"version": "v1.0.0", "sha256": strings.Repeat("3", 64),
			"status": "locked", "generation": 0,
		},
	}
	statePath, journalPath := writeState(t, root, state)
	if _, err := RegisterPlan(root, statePath, journalPath, PlanRequest{
		ExpectedRevision: 1, PlanPath: writePlanFile(t, root),
	}); err != nil {
		t.Fatalf("prior-generation TASK must not gate registration, got %v", err)
	}
}
