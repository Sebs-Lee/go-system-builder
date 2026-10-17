package qualitygate_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/entroforge/go-system-builder/internal/qualitygate"
)

func TestS10GateRefusesCanonicalResultWithoutMainProjection(t *testing.T) {
	evaluator := newTestEvaluator(t)
	root := t.TempDir()
	manifest := validS10Manifest(t, "acceptance")
	input := s10GateInput(t, "GATE-ACCEPTANCE-COMPLETE", "TR-015", "acceptance", map[string]any{
		"audit_manifest_path": "s10/acceptance-manifest.json", "audit_manifest_sha256": sha256Hex(manifest),
	})
	input.Root = root
	files := input.Files.(memoryFiles)
	files["s10/acceptance-manifest.json"] = manifest
	envelope := []byte(`{"kind":"completion_report","assignment_record":{"assignment_revision":2}}`)
	if err := os.MkdirAll(filepath.Join(root, "evidence"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "evidence/canonical-result.json"), envelope, 0600); err != nil {
		t.Fatal(err)
	}
	files["evidence/canonical-result.json"] = envelope
	input.Snapshot.State["evidence"] = append(input.Snapshot.State["evidence"].([]any), map[string]any{
		"id": "ev-canonical-result", "kind": "completion_report", "path": "evidence/canonical-result.json",
		"sha256": sha256Hex(envelope), "status": "valid", "baseline_generation": 1,
		"produced_by": []any{"builder-1"}, "invalidated_by": nil,
		"responsibility_id": "BUILD-WORK-PACKAGE", "scope_refs": []any{},
	})
	result, err := evaluator.Evaluate(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != qualitygate.StatusUnknown || result.ErrorCode != qualitygate.ErrorGateUnknown ||
		!containsPrefix(result.Conflicts, "s10:acceptance_manifest:ev-acc:external_baseline_unverifiable") {
		t.Fatalf("canonical Result without Main projection must fail closed, got %+v", result)
	}
}
