package sharedmodel

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/entroforge/go-system-builder/internal/fileview"
)

func TestNativeSchemaFileURLsPreservePathsAndLocalReferences(t *testing.T) {
	original := pilot(t)
	root := filepath.Join(t.TempDir(), "project # % 中文")
	if err := os.Rename(original, root); err != nil {
		t.Fatal(err)
	}
	put(t, root, "docs/architecture/data-model/orders.schema.json", `{"$defs":{"request":{"$ref":"model%20part%20%231.schema.json#/$defs/request"},"response":{"type":"object"}}}`)
	dependency := "docs/architecture/data-model/model part #1.schema.json"
	put(t, root, dependency, `{"$defs":{"request":{"type":"object","required":["order_id"],"properties":{"order_id":{"type":"string"}}}}}`)
	files := fileview.Disk{Root: root}
	result := Check(root, files)
	if len(result.Problems) != 0 || result.Rows != 2 {
		t.Fatalf("native schema closure failed: %+v", result)
	}
	if _, ok := result.Files[dependency]; !ok {
		t.Fatal("escaped transitive schema was omitted from the original closure")
	}
	const fragment = "/$defs/request"
	raw := schemaFileURL(filepath.Join(root, filepath.FromSlash(dependency)), fragment)
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "file" || u.Host != "" || u.Fragment != fragment {
		t.Fatalf("native schema URL lost its path or fragment: %q err=%v", raw, err)
	}
	reading := "## Document Manifest\n| 1 | model | MODEL-1 | [read](../../architecture/data-model/orders.schema.json#/$defs/request) | all | request | required |"
	if problems := Reading(root, "docs/dev/tasks/TASK-1.md", reading, files); len(problems) != 0 {
		t.Fatalf("native schema reading failed: %v", problems)
	}
}
