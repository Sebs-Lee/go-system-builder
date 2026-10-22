package catalog_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/entroforge/go-system-builder/internal/catalog"
)

func nativeRoleRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "agents"), 0755); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir("../../agents")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		data, err := os.ReadFile(filepath.Join("../../agents", entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "agents", entry.Name()), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestNativeRoleCatalogChecksProfessionalOutputsWithoutLegacyActivation(t *testing.T) {
	root := nativeRoleRoot(t)
	if err := catalog.ValidateAgents(root); err != nil {
		t.Fatalf("actual native roles rejected by old execution vocabulary: %v", err)
	}
	for _, tc := range []struct{ role, remove string }{
		{"frontend-builder", "## Output Contract"},
		{"frontend-builder", "SubagentStop"},
		{"frontend-builder", "PlanReport"},
		{"frontend-builder", "RepairResult"},
		{"frontend-builder", "task-brief"},
		{"document-verifier", "## Output Contract"},
		{"document-verifier", "SubagentStop"},
		{"document-verifier", "questionnaire"},
		{"qa", "independent"},
		{"qa", "frozen"},
	} {
		t.Run(tc.role+" missing "+tc.remove, func(t *testing.T) {
			path := filepath.Join(root, "agents", tc.role+".md")
			original, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			// Both capitalizations may occur in headings/prose. This fixture
			// removes the responsibility rather than adding obsolete words.
			body := strings.ReplaceAll(string(original), tc.remove, "removed")
			upper := strings.ToUpper(tc.remove[:1]) + tc.remove[1:]
			body = strings.ReplaceAll(body, upper, "removed")
			if body == string(original) {
				t.Fatal("negative fixture did not remove an actual contract")
			}
			if err := os.WriteFile(path, []byte(body), 0644); err != nil {
				t.Fatal(err)
			}
			defer os.WriteFile(path, original, 0644)
			if err := catalog.ValidateAgents(root); err == nil || !strings.Contains(err.Error(), tc.role) || !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tc.remove)) {
				t.Fatalf("missing native responsibility was accepted or diagnosed as unrelated: %v", err)
			}
		})
	}
}
