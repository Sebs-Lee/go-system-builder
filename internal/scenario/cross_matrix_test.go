package scenario_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/entroforge/go-system-builder/internal/scenario"
)

func writeCrossMatrix(t *testing.T, root, module string, matrix map[string]any) {
	t.Helper()
	writeJSON(t, filepath.Join(root, "docs/design/prototypes", module, "cross-matrix.json"), matrix)
}

// TestCrossMatrixRequiredAndValidated pins the carrier: a module package
// without cross-matrix.json is incomplete, and references must resolve.
func TestCrossMatrixRequiredAndValidated(t *testing.T) {
	root := newScenarioRoot(t, "investor-workbench", "ordinary")
	if err := os.Remove(filepath.Join(root, "docs/design/prototypes", "investor-workbench", "cross-matrix.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := scenario.GenerateModule(root, "investor-workbench"); err == nil || !strings.Contains(err.Error(), "cross-matrix.json") {
		t.Fatalf("missing cross-matrix must fail generate, got %v", err)
	}

	// Unknown branch reference.
	root = newScenarioRoot(t, "investor-workbench", "ordinary")
	writeCrossMatrix(t, root, "investor-workbench", map[string]any{
		"module": "investor-workbench",
		"entries": []any{map[string]any{
			"fact": "fact-investor", "req_ref": "REQ-INV-001", "story": "S-001", "branch": "branch-nonexistent",
		}},
	})
	if _, err := scenario.GenerateModule(root, "investor-workbench"); err == nil || !strings.Contains(err.Error(), "unknown branch") {
		t.Fatalf("unknown branch reference must fail, got %v", err)
	}

	// Silence is not N/A: neither branch nor reason.
	root = newScenarioRoot(t, "investor-workbench", "ordinary")
	writeCrossMatrix(t, root, "investor-workbench", map[string]any{
		"module": "investor-workbench",
		"entries": []any{map[string]any{
			"fact": "fact-investor", "req_ref": "REQ-INV-001", "story": "S-001",
		}},
	})
	if _, err := scenario.GenerateModule(root, "investor-workbench"); err == nil || !(strings.Contains(err.Error(), "silence is not N/A") || strings.Contains(err.Error(), "anyOf")) {
		t.Fatalf("cell without branch or reason must fail, got %v", err)
	}

	// A no-branch reason is a valid cell.
	root = newScenarioRoot(t, "investor-workbench", "ordinary")
	writeCrossMatrix(t, root, "investor-workbench", map[string]any{
		"module": "investor-workbench",
		"entries": []any{map[string]any{
			"fact": "fact-investor", "req_ref": "REQ-INV-001", "story": "S-001", "no_branch_reason": "covered by the same rule's negative branch",
		}},
	})
	if _, err := scenario.GenerateModule(root, "investor-workbench"); err != nil {
		t.Fatalf("reasoned no-branch cell must pass, got %v", err)
	}
}

// TestCaseIDPatternPinned pins the denominator format.
func TestCaseIDPatternPinned(t *testing.T) {
	root := newScenarioRoot(t, "investor-workbench", "ordinary")
	model := loadModel(t, root)
	branch := model["rules"].([]any)[0].(map[string]any)["branches"].([]any)[0].(map[string]any)
	branch["case_id"] = "case_lower"
	writeModel(t, root, model)
	if _, err := scenario.GenerateModule(root, "investor-workbench"); err == nil || !(strings.Contains(err.Error(), "verification denominator") || strings.Contains(err.Error(), "does not match pattern")) {
		t.Fatalf("lowercase case id must be rejected, got %v", err)
	}
}
