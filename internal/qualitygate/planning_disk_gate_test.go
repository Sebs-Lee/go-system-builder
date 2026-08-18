package qualitygate_test

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/entroforge/go-system-builder/internal/qualitygate"
	"github.com/entroforge/go-system-builder/internal/runtime"
)

// listingFiles extends memoryFiles with directory listing so the evaluator
// can discover disk-declared artifacts (BUG-CX-07).
type listingFiles map[string][]byte

func (m listingFiles) ReadFile(path string) ([]byte, error) {
	data, ok := m[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return append([]byte(nil), data...), nil
}

func (m listingFiles) ReadDir(dir string) ([]os.DirEntry, error) {
	prefix := ""
	if dir != "." {
		prefix = strings.TrimSuffix(dir, "/") + "/"
	}
	var entries []os.DirEntry
	seen := map[string]bool{}
	for path := range m {
		if !strings.HasPrefix(path, prefix) {
			continue
		}
		name := strings.SplitN(strings.TrimPrefix(path, prefix), "/", 2)[0]
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		entries = append(entries, memoryDirEntry{name: name})
	}
	return entries, nil
}

type memoryDirEntry struct{ name string }

func (e memoryDirEntry) Name() string { return e.name }
func (e memoryDirEntry) IsDir() bool  { return !strings.HasSuffix(e.name, ".md") }
func (e memoryDirEntry) Type() os.FileMode { return 0 }
func (e memoryDirEntry) Info() (os.FileInfo, error) { return nil, os.ErrNotExist }

// TestPlanningGatesReadDiskDeclaredArtifacts pins BUG-CX-07: the planning
// gates' document precondition must be satisfiable by disk-declared facts
// (contract Status: locked / task Status: complete) — the registration that
// documents[] carries is produced by the gated transitions themselves, so
// requiring it up front deadlocks the hook auto-advance path.
func TestPlanningGatesReadDiskDeclaredArtifacts(t *testing.T) {
	evaluator := newTestEvaluator(t)

	contractData := []byte("# BE-001\n\n> 状态：locked\n> 版本：v1.0.0\n")
	envelope := map[string]any{
		"schema_version":          "1.0.0",
		"evidence_id":             "ev-contract",
		"kind":                    "planning_contract",
		"runtime_id":              "loop-test",
		"baseline_generation":     1,
		"producer_agent_id":       "planner-1",
		"producer_responsibility": "Contract Planner",
		"subject_refs": []any{map[string]any{
			"path": "docs/contracts/BE-001.md", "version": "v1.0.0",
			"sha256": sha256Hex(contractData),
		}},
		"conclusion": "pass",
		"created_at": "2026-08-17T00:00:00Z",
	}
	envelopeData, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	input := qualitygate.Input{
		Snapshot: runtime.Snapshot{
			Revision: 3,
			State: map[string]any{
				"runtime_id": "loop-test",
				"lifecycle":  map[string]any{"state": "planning", "phase": "contracts", "phase_revision": float64(1)},
				"baseline":   map[string]any{"generation": float64(1)},
				"review":     map[string]any{"round": float64(0)},
				"documents":  []any{},
				"evidence": []any{map[string]any{
					"id": "ev-contract", "kind": "planning_contract", "path": "evidence/contract.json",
					"sha256": sha256Hex(envelopeData), "status": "valid", "baseline_generation": float64(1),
					"review_round": nil, "produced_by": []any{"planner-1"}, "invalidated_by": nil,
					"responsibility_id": "Contract Planner", "scope_refs": []any{},
				}},
			},
		},
		TransitionID: "PTR-PLAN-02",
		GateID:       "GATE-PLANNING-CONTRACTS-COMPLETE",
		Files: listingFiles{
			"docs/contracts/BE-001.md": contractData,
			"evidence/contract.json":   envelopeData,
		},
	}
	result, err := evaluator.Evaluate(context.Background(), input)
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if result.Status != qualitygate.StatusSatisfied {
		t.Fatalf("BUG-CX-07: disk-declared locked contract + qualified planning evidence must satisfy the gate without a pre-existing documents[] registration; got status=%q missing=%#v", result.Status, result.Missing)
	}
}

// TestPlanningGatesStillRefuseWhenDiskAlsoLacks: the disk fallback must not
// weaken the gate — no disk contract and no registration stays NOT_READY.
func TestPlanningGatesStillRefuseWhenDiskAlsoLacks(t *testing.T) {
	evaluator := newTestEvaluator(t)
	input := planningInputForGate(t, "GATE-PLANNING-CONTRACTS-COMPLETE", "PTR-PLAN-02", "contracts", "planning_contract_record", "Contract Planner")
	result, err := evaluator.Evaluate(context.Background(), input)
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if result.Status != qualitygate.StatusNotReady {
		t.Fatalf("gate must stay NOT_READY when neither disk nor documents[] declares the artifact, got %q", result.Status)
	}
}

// TestPlanningDesignGateReadsDiskDeclaredArchitecture pins BUG-CX-13 A3:
// the S2 exit gate must accept a disk-declared locked architecture document
// without a pre-existing documents[] registration (the registration happens
// at PTR-PLAN-01's commit; nothing produced it before, so the organic S2
// exit was deadlocked).
func TestPlanningDesignGateReadsDiskDeclaredArchitecture(t *testing.T) {
	evaluator := newTestEvaluator(t)
	archData := []byte("# ARCHITECTURE-001\n\n> 状态：locked\n> 版本：v1.0.0\n")
	reqData := []byte("# REQ-001\n\n> 状态：locked\n> 版本：v1.0.0\n")
	envelope := map[string]any{
		"schema_version":          "1.0.0",
		"evidence_id":             "ev-design",
		"kind":                    "planning_design",
		"runtime_id":              "loop-test",
		"baseline_generation":     1,
		"producer_agent_id":       "architect-1",
		"producer_responsibility": "Architect",
		"subject_refs": []any{
			map[string]any{"path": "docs/design/architecture/ARCHITECTURE-001.md", "version": "v1.0.0", "sha256": sha256Hex(archData)},
			map[string]any{"path": "docs/requirements/REQ-001.md", "version": "v1.0.0", "sha256": sha256Hex(reqData)},
		},
		"conclusion": "pass",
		"created_at": "2026-08-18T00:00:00Z",
	}
	envelopeData, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	input := qualitygate.Input{
		Snapshot: runtime.Snapshot{
			Revision: 2,
			State: map[string]any{
				"runtime_id": "loop-test",
				"lifecycle":  map[string]any{"state": "planning", "phase": "design", "phase_revision": float64(1)},
				"baseline":   map[string]any{"generation": float64(1)},
				"review":     map[string]any{"round": float64(0)},
				"documents": []any{ // req registered by bind; design NOT registered (organic pre-commit state)
					map[string]any{"id": "REQ-001", "kind": "req", "path": "docs/requirements/REQ-001.md", "version": "v1.0.0", "sha256": sha256Hex(reqData), "status": "locked", "generation": float64(1)},
				},
				"evidence": []any{map[string]any{
					"id": "ev-design", "kind": "planning_design", "path": "evidence/design.json",
					"sha256": sha256Hex(envelopeData), "status": "valid", "baseline_generation": float64(1),
					"review_round": nil, "produced_by": []any{"architect-1"}, "invalidated_by": nil,
					"responsibility_id": "Architect", "scope_refs": []any{},
				}},
			},
		},
		TransitionID: "PTR-PLAN-01",
		GateID:       "GATE-PLANNING-DESIGN-COMPLETE",
		Files: listingFiles{
			"docs/design/architecture/ARCHITECTURE-001.md": archData,
			"docs/requirements/REQ-001.md":         reqData,
			"evidence/design.json":                 envelopeData,
		},
	}
	result, err := evaluator.Evaluate(context.Background(), input)
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if result.Status != qualitygate.StatusSatisfied {
		t.Fatalf("BUG-CX-13: disk-declared locked architecture + registered req must satisfy the S2 exit gate; got status=%q missing=%#v conflicts=%v", result.Status, result.Missing, result.Conflicts)
	}
}

// TestDocumentPassGateFlagsRegisteredDocumentDrift pins BUG-CX-11 B3:
// a registered document whose on-disk sha no longer matches must block
// GATE-DOCUMENT-PASS with the path named — otherwise a document the
// reviewers never saw can be re-registered from disk and locked into
// building by TR-003's commit.
func TestDocumentPassGateFlagsRegisteredDocumentDrift(t *testing.T) {
	evaluator := newTestEvaluator(t)
	contractData := []byte("# BE-001\n\n> 状态：locked\n> 版本：v1.0.0\n")
	driftedData := []byte("# BE-001 (edited after review)\n\n> 状态：locked\n> 版本：v1.0.0\n")
	doc := func(data []byte) map[string]any {
		return map[string]any{"id": "BE-001", "kind": "contract", "path": "docs/contracts/BE-001.md", "version": "v1.0.0", "sha256": sha256Hex(data), "status": "locked", "generation": float64(1)}
	}
	input := qualitygate.Input{
		Snapshot: runtime.Snapshot{
			Revision: 5,
			State: map[string]any{
				"runtime_id": "loop-test",
				"lifecycle":  map[string]any{"state": "document_verification", "phase": nil, "phase_revision": float64(1)},
				"baseline":   map[string]any{"generation": float64(1)},
				"review":     map[string]any{"round": float64(0)},
				// registered with the ORIGINAL sha; disk carries the edited bytes
				"documents": []any{doc(contractData)},
				"evidence":  []any{},
			},
		},
		TransitionID: "TR-003",
		GateID:       "GATE-DOCUMENT-PASS",
		Files: listingFiles{
			"docs/contracts/BE-001.md": driftedData,
		},
	}
	result, err := evaluator.Evaluate(context.Background(), input)
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	blocked := false
	for _, conflict := range result.Conflicts {
		if strings.Contains(conflict, "document_drift:docs/contracts/BE-001.md") {
			blocked = true
		}
	}
	if !blocked {
		t.Fatalf("BUG-CX-11: a registered document drifting on disk must produce a document_drift conflict naming the path; got status=%q conflicts=%v missing=%v", result.Status, result.Conflicts, result.Missing)
	}
}
