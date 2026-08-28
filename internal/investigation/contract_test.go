package investigation_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/entroforge/go-system-builder/internal/investigation"
	"github.com/entroforge/go-system-builder/internal/schema"
)

func TestApproveContractRequiresExactCaseFindingCoverage(t *testing.T) {
	fixture := newIntakeFixture(t, []string{"finding-1", "finding-2"})
	setContractLifecycle(t, fixture)
	if _, err := investigation.Ingest(fixture.root, fixture.statePath, fixture.journalPath, investigation.IngestRequest{
		ExpectedRevision:  0,
		GroupingRationale: "the sealed batch is the provisional grouping boundary",
	}); err != nil {
		t.Fatalf("Ingest() error = %v", err)
	}
	prepareCaseForContractApproval(t, fixture)
	contractPath := writeContractDraft(t, fixture.root, []string{"finding-1"})
	_, err := investigation.ApproveContract(fixture.root, fixture.statePath, fixture.journalPath, investigation.ContractRequest{
		ExpectedRevision: 1,
		CaseID:           "investigation-case-observation-batch-r1",
		ContractPath:     contractPath,
		ApprovedBy:       "main-session",
	})
	if err == nil || !strings.Contains(err.Error(), "exact") || !strings.Contains(err.Error(), "finding-2") {
		t.Fatalf("ApproveContract() error = %v, want exact Finding coverage guidance", err)
	}
}

func TestApproveContractRequiresCausalClosureBeforeS9(t *testing.T) {
	fixture := newIntakeFixture(t, []string{"finding-1"})
	setContractLifecycle(t, fixture)
	if _, err := investigation.Ingest(fixture.root, fixture.statePath, fixture.journalPath, investigation.IngestRequest{
		ExpectedRevision:  0,
		GroupingRationale: "the sealed batch is the provisional grouping boundary",
	}); err != nil {
		t.Fatalf("Ingest() error = %v", err)
	}
	_, err := investigation.ApproveContract(fixture.root, fixture.statePath, fixture.journalPath, investigation.ContractRequest{
		ExpectedRevision: 1,
		CaseID:           "investigation-case-observation-batch-r1",
		ContractPath:     writeContractDraft(t, fixture.root, []string{"finding-1"}),
		ApprovedBy:       "main-session",
	})
	if err == nil || !strings.Contains(err.Error(), "unexplained Finding IDs") {
		t.Fatalf("ApproveContract() error = %v, want causal-closure guidance", err)
	}
}

func TestApproveContractCommitsApprovedContractAndCaseRevision(t *testing.T) {
	fixture := newIntakeFixture(t, []string{"finding-2", "finding-1"})
	setContractLifecycle(t, fixture)
	if _, err := investigation.Ingest(fixture.root, fixture.statePath, fixture.journalPath, investigation.IngestRequest{
		ExpectedRevision:  0,
		GroupingRationale: "the sealed batch is the provisional grouping boundary",
	}); err != nil {
		t.Fatalf("Ingest() error = %v", err)
	}
	prepareCaseForContractApproval(t, fixture)
	contractPath := writeContractDraft(t, fixture.root, []string{"finding-1", "finding-2"})
	snapshot, err := investigation.ApproveContract(fixture.root, fixture.statePath, fixture.journalPath, investigation.ContractRequest{
		ExpectedRevision: 1,
		CaseID:           "investigation-case-observation-batch-r1",
		ContractPath:     contractPath,
		ApprovedBy:       "main-session",
	})
	if err != nil {
		t.Fatalf("ApproveContract() error = %v", err)
	}
	if snapshot.Revision != 2 {
		t.Fatalf("revision = %d, want 2", snapshot.Revision)
	}
	lifecycle := snapshot.State["lifecycle"].(map[string]any)
	if lifecycle["phase"] != "repair_readback" {
		t.Fatalf("lifecycle phase = %v, want repair_readback", lifecycle["phase"])
	}
	review := snapshot.State["review"].(map[string]any)
	pointer := review["investigation"].(map[string]any)
	if pointer["status"] != "contract_approved" {
		t.Fatalf("investigation status = %v, want contract_approved", pointer["status"])
	}
	if pointer["repair_contract_ref"] == nil || pointer["repair_contract_sha256"] == nil {
		t.Fatalf("approved pointer missing Contract ref/hash: %#v", pointer)
	}
	approvedContractPath := filepath.Join(fixture.root, filepath.FromSlash(pointer["repair_contract_ref"].(string)))
	contractBytes, err := os.ReadFile(approvedContractPath)
	if err != nil {
		t.Fatalf("read approved Contract: %v", err)
	}
	if err := schema.NewEmbeddedValidator().ValidateBytes("repair-contract.schema.json", contractBytes); err != nil {
		t.Fatalf("approved Contract schema: %v", err)
	}
	var contractDocument map[string]any
	if err := json.Unmarshal(contractBytes, &contractDocument); err != nil {
		t.Fatal(err)
	}
	if contractDocument["revision"] != float64(2) || contractDocument["status"] != "approved" {
		t.Fatalf("approved Contract = %#v", contractDocument)
	}
	casePath := pointer["path"].(string)
	caseBytes, err := os.ReadFile(filepath.Join(fixture.root, filepath.FromSlash(casePath)))
	if err != nil {
		t.Fatalf("read approved Case: %v", err)
	}
	if err := schema.NewEmbeddedValidator().ValidateBytes("review-investigation-case.schema.json", caseBytes); err != nil {
		t.Fatalf("approved Case schema: %v", err)
	}
	var caseDocument map[string]any
	if err := json.Unmarshal(caseBytes, &caseDocument); err != nil {
		t.Fatal(err)
	}
	if caseDocument["revision"] != float64(2) || caseDocument["status"] != "contract_approved" {
		t.Fatalf("approved Case = %#v", caseDocument)
	}
	journalBytes, err := os.ReadFile(fixture.journalPath)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(journalBytes)), "\n")
	var event map[string]any
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &event); err != nil {
		t.Fatal(err)
	}
	if event["transition_id"] != "S8-REPAIR-CONTRACT-APPROVAL" {
		t.Fatalf("Contract approval must use its own transition id, got %#v", event["transition_id"])
	}
}

func writeContractDraft(t *testing.T, root string, findingIDs []string) string {
	t.Helper()
	draft := map[string]any{
		"schema_version": "1.0.0", "repair_contract_id": "repair-contract-r1", "case_id": "investigation-case-observation-batch-r1", "revision": 1,
		"status": "draft", "source_finding_ids": findingIDs,
		"root_cause_statement": "the authoritative payload contract is duplicated and drifts across the boundary",
		"violated_invariant":   "one source of truth owns the payload shape",
		"causal_model_ref":     "case://investigation-case-observation-batch-r1/causal-model",
		"architecture_intent":  "restore one authoritative contract boundary",
		"repair_units":         []any{map[string]any{"id": "repair-unit-1", "description": "centralize the payload contract"}},
		"prospective_scope":    []any{"internal/api", "frontend/client"}, "forbidden_scope": []any{"docs/requirements"},
		"symptom_assertions":         []any{"finding-1 restored", "finding-2 restored"},
		"root_invariant_assertions":  []any{"payload contract has one owner"},
		"detection_gap_assertions":   []any{"contract test fails on field drift"},
		"stop_escalation_conditions": []any{"REQ or locked contract must change"},
	}
	data, err := json.MarshalIndent(draft, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "repair-contract-draft.json")
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func setContractLifecycle(t *testing.T, fixture *intakeFixture) {
	t.Helper()
	data, err := os.ReadFile(fixture.statePath)
	if err != nil {
		t.Fatal(err)
	}
	var state map[string]any
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatal(err)
	}
	state["lifecycle"] = map[string]any{"state": "bug_resolution", "phase": "investigation", "phase_revision": 0}
	updated, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fixture.statePath, append(updated, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}

func prepareCaseForContractApproval(t *testing.T, fixture *intakeFixture) {
	t.Helper()
	stateBytes, err := os.ReadFile(fixture.statePath)
	if err != nil {
		t.Fatal(err)
	}
	var state map[string]any
	if err := json.Unmarshal(stateBytes, &state); err != nil {
		t.Fatal(err)
	}
	pointer := state["review"].(map[string]any)["investigation"].(map[string]any)
	casePath := filepath.Join(fixture.root, filepath.FromSlash(pointer["path"].(string)))
	caseBytes, err := os.ReadFile(casePath)
	if err != nil {
		t.Fatal(err)
	}
	var caseDocument map[string]any
	if err := json.Unmarshal(caseBytes, &caseDocument); err != nil {
		t.Fatal(err)
	}
	caseDocument["unexplained_finding_ids"] = []any{}
	caseDocument["causal_model"] = map[string]any{"trigger": "payload crosses boundary", "violated_invariant": "one owner", "faulty_mechanism": "duplicate schema", "propagation": "decoder rejects fields", "symptoms": []any{"finding-1", "finding-2"}}
	caseDocument["primary_root_cause"] = "the payload contract has two incompatible owners"
	caseDocument["blast_radius"] = map[string]any{"paths": []any{"internal/api"}}
	caseDocument["detection_gap"] = map[string]any{"gap_type": "contract", "evidence_refs": []any{"evidence://drift"}}
	caseDocument["no_competing_hypothesis"] = "the sealed occurrence is only consistent with the payload-contract drift mechanism"
	caseDocument["route"] = "s9_repair"
	caseDocument["route_reason"] = "implementation boundary must be repaired"
	updatedCase, err := json.MarshalIndent(caseDocument, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	updatedCase = append(updatedCase, '\n')
	if err := os.WriteFile(casePath, updatedCase, 0o644); err != nil {
		t.Fatal(err)
	}
	pointer["sha256"] = hash(updatedCase)
	updatedState, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fixture.statePath, append(updatedState, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}
