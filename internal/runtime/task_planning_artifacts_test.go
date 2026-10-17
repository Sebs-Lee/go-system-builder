package runtime_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/entroforge/go-system-builder/internal/fileview"
	"github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/workpacket"
)

func TestTaskPlanningArtifactsReturnsExactREQBatchIncludingCancelled(t *testing.T) {
	root, state, files := taskPlanningRuntimeFixture(t)
	inputs, err := runtime.ResolveTaskPlanningInputs(root, state, files)
	if err != nil {
		t.Fatal(err)
	}
	binding, err := runtime.CurrentTaskPlanningBinding(state)
	if err != nil {
		t.Fatal(err)
	}
	p, err := workpacket.BuildTaskPlanning(binding, inputs)
	if err != nil {
		t.Fatal(err)
	}
	writeTaskPlanningFile(t, root, "docs/dev/tasks/TASK-REQ-001-01.md", "> Status: complete\n> Version: v1.0.0\n> Source REQ refs: REQ-001\n")
	writeTaskPlanningFile(t, root, "docs/dev/tasks/TASK-REQ-001-02.md", "> Status: cancelled\n> Version: v1.0.0\n> Source REQ refs: REQ-001\n")
	writeTaskPlanningFile(t, root, "docs/dev/tasks/index-REQ-001.md", "# REQ-001 dispatch\n> REQ: REQ-001\n> Dispatch policy: waves-v1\n> Revision: 1\n> Status: complete\n\n## W1 First wave\n\n- [ ] [TASK-REQ-001-01](TASK-REQ-001-01.md)\n")
	answer := workpacket.TaskPlanningAnswer{
		Conclusion: stringPtr("pass"), IndexPath: stringPtr("docs/dev/tasks/index-REQ-001.md"),
		TaskPaths: []string{"docs/dev/tasks/TASK-REQ-001-02.md", "docs/dev/tasks/TASK-REQ-001-01.md"},
	}
	refs, err := runtime.TaskPlanningArtifacts(root, state, p, answer, files)
	if err != nil {
		t.Fatal(err)
	}
	if len(refs) != 3 || refs[0].Path != "docs/dev/tasks/index-REQ-001.md" || refs[1].Path != "docs/dev/tasks/TASK-REQ-001-01.md" || refs[2].Path != "docs/dev/tasks/TASK-REQ-001-02.md" {
		t.Fatalf("artifacts must be index first then all sorted tasks: %#v", refs)
	}
}

func TestTaskPlanningArtifactsRejectsOmittedCancelledOrUnplannedTask(t *testing.T) {
	root, state, files := taskPlanningRuntimeFixture(t)
	inputs, err := runtime.ResolveTaskPlanningInputs(root, state, files)
	if err != nil {
		t.Fatal(err)
	}
	binding, _ := runtime.CurrentTaskPlanningBinding(state)
	p, err := workpacket.BuildTaskPlanning(binding, inputs)
	if err != nil {
		t.Fatal(err)
	}
	writeTaskPlanningFile(t, root, "docs/dev/tasks/TASK-REQ-001-01.md", "> Status: complete\n> Source REQ refs: REQ-001\n")
	writeTaskPlanningFile(t, root, "docs/dev/tasks/TASK-REQ-001-02.md", "> Status: cancelled\n> Source REQ refs: REQ-001\n")
	writeTaskPlanningFile(t, root, "docs/dev/tasks/index-REQ-001.md", "# REQ-001 dispatch\n> REQ: REQ-001\n> Dispatch policy: waves-v1\n> Revision: 1\n> Status: complete\n\n## W1\n- [ ] [TASK-REQ-001-01](TASK-REQ-001-01.md)\n")
	answer := workpacket.TaskPlanningAnswer{Conclusion: stringPtr("pass"), IndexPath: stringPtr("docs/dev/tasks/index-REQ-001.md"), TaskPaths: []string{"docs/dev/tasks/TASK-REQ-001-01.md"}}
	if _, err := runtime.TaskPlanningArtifacts(root, state, p, answer, files); err == nil {
		t.Fatal("an answer that omitted the cancelled TASK was accepted")
	}
	answer.TaskPaths = append(answer.TaskPaths, "docs/dev/tasks/TASK-REQ-001-02.md")
	writeTaskPlanningFile(t, root, "docs/dev/tasks/index-REQ-001.md", "# REQ-001 dispatch\n> REQ: REQ-001\n> Dispatch policy: waves-v1\n> Revision: 1\n> Status: complete\n\n## W1\n")
	if _, err := runtime.TaskPlanningArtifacts(root, state, p, answer, files); err == nil {
		t.Fatal("an index that omits a complete TASK was accepted")
	}
}

func TestTaskPlanningPacketRejectsUnregisteredContractAdditionAndSourceDrift(t *testing.T) {
	t.Run("extra contract", func(t *testing.T) {
		root, state, files := taskPlanningRuntimeFixture(t)
		inputs, err := runtime.ResolveTaskPlanningInputs(root, state, files)
		if err != nil {
			t.Fatal(err)
		}
		binding, _ := runtime.CurrentTaskPlanningBinding(state)
		p, err := workpacket.BuildTaskPlanning(binding, inputs)
		if err != nil {
			t.Fatal(err)
		}
		writeTaskPlanningFile(t, root, "docs/dev/contracts/FE-002.md", "> Status: locked\n> Version: v1.0.0\n")
		if err := runtime.ValidateTaskPlanningPacket(root, state, p, files); err == nil {
			t.Fatal("an unregistered extra contract was omitted from the frozen contract closure")
		}
	})
	t.Run("design drift", func(t *testing.T) {
		root, state, files := taskPlanningRuntimeFixture(t)
		inputs, err := runtime.ResolveTaskPlanningInputs(root, state, files)
		if err != nil {
			t.Fatal(err)
		}
		binding, _ := runtime.CurrentTaskPlanningBinding(state)
		p, err := workpacket.BuildTaskPlanning(binding, inputs)
		if err != nil {
			t.Fatal(err)
		}
		writeTaskPlanningFile(t, root, "docs/architecture/ARCHITECTURE-001.md", "> Status: locked\n> Version: v1.0.0\nchanged\n")
		if err := runtime.ValidateTaskPlanningPacket(root, state, p, files); err == nil {
			t.Fatal("S3 registered design source drift was accepted")
		}
	})
}

func TestTaskPlanningRejectsS3EnvelopeRuntimeMismatch(t *testing.T) {
	root, state, files := taskPlanningRuntimeFixture(t)
	evidencePath := ".claude/evidence/E-S3-001.json"
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(evidencePath)))
	if err != nil {
		t.Fatal(err)
	}
	var envelope map[string]any
	if err := json.Unmarshal(data, &envelope); err != nil {
		t.Fatal(err)
	}
	envelope["runtime_id"] = "runtime-other"
	mutated, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	writeBytes(t, root, evidencePath, mutated)
	evidenceRow := state["evidence"].([]any)[0].(map[string]any)
	evidenceRow["sha256"] = workpacket.Digest(mutated)
	if _, err := runtime.ResolveTaskPlanningInputs(root, state, files); err == nil {
		t.Fatal("S3 evidence envelope with a different runtime identity was accepted")
	}
}

func TestTaskPlanningInputsProjectAcceptedS3SourcesIncludingSharedModels(t *testing.T) {
	root, state, files := taskPlanningRuntimeFixture(t)
	modelPath := "docs/models/REQ-001.json"
	modelBytes := []byte(`{"type":"object","title":"REQ-001 model"}`)
	writeBytes(t, root, modelPath, modelBytes)
	addTaskPlanningS3Subject(t, root, state, workpacket.Subject{Path: modelPath, SHA256: workpacket.Digest(modelBytes)})
	state["documents"] = append(state["documents"].([]any), map[string]any{
		"id": "shared-model:" + modelPath, "kind": "design", "path": modelPath,
		"version": "", "sha256": workpacket.Digest(modelBytes), "status": "locked", "generation": 1,
	})

	inputs, err := runtime.ResolveTaskPlanningInputs(root, state, files)
	if err != nil {
		t.Fatal(err)
	}
	sources := map[string]workpacket.Subject{}
	for _, source := range inputs.Sources {
		sources[source.Path] = source
	}
	for _, path := range []string{
		"docs/architecture/ARCHITECTURE-001.md",
		"docs/design/decisions/ADR-001.md",
		modelPath,
	} {
		if _, ok := sources[path]; !ok {
			t.Fatalf("accepted S3 source %s was not projected into S4: %#v", path, inputs.Sources)
		}
	}
	if sources[modelPath].Version != "" || sources[modelPath].SHA256 != workpacket.Digest(modelBytes) {
		t.Fatalf("S3 shared-model subject was not preserved exactly: %+v", sources[modelPath])
	}

	writeBytes(t, root, modelPath, []byte(`{"type":"object","title":"changed"}`))
	if _, err := runtime.ResolveTaskPlanningInputs(root, state, files); err == nil {
		t.Fatal("changed S3 shared-model subject was not rejected by the original frozen-input validator")
	}
}

func TestTaskPlanningInputsAllowDeclaredForeignREQContractBatch(t *testing.T) {
	root, state, files := taskPlanningRuntimeFixture(t)
	writeTaskPlanningFile(t, root, "docs/dev/contracts/CONTRACTS-002.md", "# CONTRACTS-002\n> Requirement: REQ-002\n> Status: locked\n> Version: v1.0.0\n\n## 合同清单\n| Contract ID | Type | Name | Status | Version | Notes |\n|:---|:---|:---|:---|:---|:---|\n| BE-002 | backend | api | locked | v1.0.0 | foreign |\n")
	writeTaskPlanningFile(t, root, "docs/dev/contracts/BE-002.md", "> Status: locked\n> Version: v1.0.0\n")
	if _, err := runtime.ResolveTaskPlanningInputs(root, state, files); err != nil {
		t.Fatalf("declared foreign REQ contract batch blocked the accepted S3/S4 closure: %v", err)
	}

	writeTaskPlanningFile(t, root, "docs/dev/contracts/SYNC-999.md", "> Status: draft\n")
	if _, err := runtime.ResolveTaskPlanningInputs(root, state, files); err == nil {
		t.Fatal("unowned extra contract document was accepted alongside a foreign REQ batch")
	}
}

func taskPlanningRuntimeFixture(t *testing.T) (string, map[string]any, fileview.Reader) {
	t.Helper()
	root := t.TempDir()
	for _, dir := range []string{".claude/evidence", "docs/requirements", "docs/architecture", "docs/design/decisions", "docs/dev/contracts", "docs/dev/tasks", "docs/rules"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0700); err != nil {
			t.Fatal(err)
		}
	}
	reqPath := "docs/requirements/REQ-001.md"
	designPath := "docs/architecture/ARCHITECTURE-001.md"
	conditionalPath := "docs/design/decisions/ADR-001.md"
	indexPath := "docs/dev/contracts/CONTRACTS-001.md"
	contractPath := "docs/dev/contracts/FE-001.md"
	req := []byte("# REQ-001\n> Status: locked\n> Version: v1.0.0\n")
	design := []byte("# Architecture\n> Status: locked\n> Version: v1.0.0\n")
	conditional := []byte("# ADR-001\n> Status: accepted\n> Version: v1.0.0\n")
	index := []byte("# CONTRACTS-001\n> Requirement: REQ-001\n> Status: locked\n> Version: v1.0.0\n\n## 合同清单\n| Contract ID | Type | Name | Status | Version | Notes |\n|:---|:---|:---|:---|:---|:---|\n| FE-001 | frontend | ui | locked | v1.0.0 | entry |\n\n## 需求覆盖矩阵\n| Contract | Clause | Purpose |\n|:---|:---|:---|\n| FE-001 §1 | user view | visible |\n")
	contract := []byte("# FE-001\n> Status: locked\n> Version: v1.0.0\n")
	taskTemplate := []byte("# Task template\n")
	dispatchRules := []byte("# Dispatch plan rules\n")
	writeBytes(t, root, reqPath, req)
	writeBytes(t, root, designPath, design)
	writeBytes(t, root, conditionalPath, conditional)
	writeBytes(t, root, indexPath, index)
	writeBytes(t, root, contractPath, contract)
	writeBytes(t, root, "docs/dev/tasks/TASK-template.md", taskTemplate)
	writeBytes(t, root, "docs/rules/dispatch-plan.md", dispatchRules)

	reqRef := workpacket.Reference{ID: "REQ-001", Path: reqPath, Version: "v1.0.0", SHA256: workpacket.Digest(req)}
	designRef := workpacket.Reference{ID: "ARCHITECTURE-001", Path: designPath, Version: "v1.0.0", SHA256: workpacket.Digest(design)}
	s3Binding := workpacket.PlanningBinding{RuntimeID: "runtime-001", Generation: 1, StageEntry: 2, Stage: "S3", Owner: "Main", Requirement: reqRef}
	s3Packet, err := workpacket.BuildContracts(s3Binding, workpacket.ContractInputs{
		Designs:     []workpacket.Reference{designRef},
		Sources:     []workpacket.Subject{{Path: conditionalPath, Version: "v1.0.0", SHA256: workpacket.Digest(conditional)}},
		Directories: map[string][]string{"docs/design/decisions": {"ADR-001.md"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	answer := workpacket.ContractAnswer{
		Conclusion: stringPtr("pass"), IndexPath: stringPtr(indexPath), ContractPaths: []string{contractPath},
		Applicability: map[string]workpacket.ContractApplicability{
			"FE":   {Disposition: "required", Rationale: "The REQ includes a user-facing view."},
			"BE":   {Disposition: "not_applicable", Rationale: "This REQ has no backend behavior."},
			"SYNC": {Disposition: "not_applicable", Rationale: "This REQ has no shared integration behavior."},
		},
	}
	artifacts := []workpacket.Reference{
		{ID: "CONTRACTS-001", Path: indexPath, Version: "v1.0.0", SHA256: workpacket.Digest(index)},
		{ID: "FE-001", Path: contractPath, Version: "v1.0.0", SHA256: workpacket.Digest(contract)},
	}
	s3Evidence, err := workpacket.ContractEvidence(s3Packet, "E-S3-001", artifacts, answer, time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	evidenceData, err := json.Marshal(s3Evidence)
	if err != nil {
		t.Fatal(err)
	}
	evidencePath := ".claude/evidence/E-S3-001.json"
	writeBytes(t, root, evidencePath, evidenceData)
	packetData, _ := json.Marshal(s3Packet)
	var packetRow map[string]any
	_ = json.Unmarshal(packetData, &packetRow)
	state := map[string]any{
		"schema_version": "1.2.0", "runtime_id": "runtime-001",
		"configuration": map[string]any{"optimization": map[string]any{"packet": true}},
		"lifecycle":     map[string]any{"state": "planning", "phase": "tasks", "phase_revision": 3},
		"baseline":      map[string]any{"generation": 1},
		"bound_req":     map[string]any{"id": "REQ-001", "path": reqPath, "version": "v1.0.0", "sha256": workpacket.Digest(req), "status": "locked", "workspace": map[string]any{"project_root": root}},
		"documents": []any{
			map[string]any{"id": designRef.ID, "kind": "design", "path": designRef.Path, "version": designRef.Version, "sha256": designRef.SHA256, "status": "locked", "generation": 1},
			map[string]any{"id": "CONTRACTS-001", "kind": "contract", "path": indexPath, "version": "v1.0.0", "sha256": workpacket.Digest(index), "status": "locked", "generation": 1},
			map[string]any{"id": "FE-001", "kind": "contract", "path": contractPath, "version": "v1.0.0", "sha256": workpacket.Digest(contract), "status": "locked", "generation": 1},
		},
		"entities": map[string]any{
			"work_packets": []any{packetRow},
			"submissions": []any{map[string]any{
				"submission_key": workpacket.Digest([]byte("submission:" + s3Packet.ID)), "packet_id": s3Packet.ID, "owner": "Main",
				"content_sha256": mustContractAnswerDigest(t, answer), "evidence_id": "E-S3-001", "followup_event_id": workpacket.Digest([]byte(workpacket.Digest([]byte("submission:"+s3Packet.ID)) + ":after-domain")),
				"control_status": "advanced", "origin_state": "planning", "origin_phase": "contracts", "origin_entry": 2, "transition_event_id": "evt-PTR-PLAN-02",
			}},
		},
		"evidence": []any{map[string]any{
			"id": "E-S3-001", "kind": "planning_contract", "baseline_generation": 1,
			"responsibility_id": "Orchestrator", "produced_by": []any{"Main"}, "status": "valid", "path": evidencePath, "sha256": workpacket.Digest(evidenceData),
			"review_round": nil, "invalidated_by": nil, "invalidation_rule": nil, "invalidation_reason": nil, "scope_refs": []any{},
		}},
	}
	return root, state, fileview.Disk{Root: root}
}

func writeTaskPlanningFile(t *testing.T, root, path, content string) {
	t.Helper()
	writeBytes(t, root, path, []byte(content))
}

func addTaskPlanningS3Subject(t *testing.T, root string, state map[string]any, subject workpacket.Subject) {
	t.Helper()
	evidencePath := ".claude/evidence/E-S3-001.json"
	evidenceData, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(evidencePath)))
	if err != nil {
		t.Fatal(err)
	}
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(evidenceData, &envelope); err != nil {
		t.Fatal(err)
	}
	var subjects []workpacket.Subject
	if err := json.Unmarshal(envelope["subject_refs"], &subjects); err != nil {
		t.Fatal(err)
	}
	subjects = append(subjects, subject)
	envelope["subject_refs"], err = json.Marshal(subjects)
	if err != nil {
		t.Fatal(err)
	}
	updated, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	writeBytes(t, root, evidencePath, updated)
	state["evidence"].([]any)[0].(map[string]any)["sha256"] = workpacket.Digest(updated)
}

func writeBytes(t *testing.T, root, path string, data []byte) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, data, 0600); err != nil {
		t.Fatal(err)
	}
}

func stringPtr(value string) *string { return &value }

func mustContractAnswerDigest(t *testing.T, answer workpacket.ContractAnswer) string {
	t.Helper()
	digest, err := workpacket.ContractAnswerDigest(answer)
	if err != nil {
		t.Fatal(err)
	}
	return digest
}
