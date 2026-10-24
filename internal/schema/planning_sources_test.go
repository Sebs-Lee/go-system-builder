package schema_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/entroforge/go-system-builder/internal/schema"
)

// Runtime embeds the planning packet schemas for offline validation. Issuance
// and persistence must agree on the full source contract, not just a sampled
// packet that happens to fit both sets of constraints.
func TestPlanningSourceSchemasAgreeWithRuntimePersistence(t *testing.T) {
	read := func(name string) map[string]any {
		t.Helper()
		data, err := schema.ReadAsset(name)
		if err != nil {
			t.Fatal(err)
		}
		var value map[string]any
		if err := json.Unmarshal(data, &value); err != nil {
			t.Fatal(err)
		}
		return value
	}
	for template, asset := range map[string]string{
		"s3-contract-v1":      "contract-work-packet.schema.json",
		"s4-task-planning-v1": "task-planning-work-packet.schema.json",
	} {
		want := read(asset)["properties"].(map[string]any)["inputs"]
		for _, runtimeAsset := range []string{"loop-state-v1.2.schema.json", "loop-state-v1.3.schema.json"} {
			count := 0
			var walk func(any)
			walk = func(value any) {
				switch node := value.(type) {
				case map[string]any:
					properties, _ := node["properties"].(map[string]any)
					id, _ := properties["template_id"].(map[string]any)
					if id["const"] == template {
						count++
						if !reflect.DeepEqual(properties["inputs"], want) {
							t.Errorf("%s has a different %s source contract from %s", runtimeAsset, template, asset)
						}
					}
					for _, child := range node {
						walk(child)
					}
				case []any:
					for _, child := range node {
						walk(child)
					}
				}
			}
			walk(read(runtimeAsset))
			if count != 1 {
				t.Fatalf("%s has %d branches for %s, want one", runtimeAsset, count, template)
			}
		}
	}
}
