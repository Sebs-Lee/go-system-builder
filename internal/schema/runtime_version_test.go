package schema_test

import (
	"encoding/json"
	"testing"

	"github.com/entroforge/go-system-builder/internal/schema"
)

func TestRuntimeSchemaSelectsClosedVersion(t *testing.T) {
	data, err := schema.ReadAsset("loop-state.example.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, version                 string
		optimization, unknown, accept bool
	}{
		{"legacy", "1.1.0", false, false, true},
		{"new empty", "1.2.0", false, false, true},
		{"new enabled", "1.2.0", true, false, true},
		{"legacy cannot enable", "1.1.0", true, false, false},
		{"new remains closed", "1.2.0", true, true, false},
		{"unknown", "9.0.0", false, false, false},
		{"missing", "", false, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var state map[string]any
			if err := json.Unmarshal(data, &state); err != nil {
				t.Fatal(err)
			}
			state["schema_version"] = tc.version
			if tc.version == "" {
				delete(state, "schema_version")
			}
			if tc.optimization {
				state["configuration"].(map[string]any)["optimization"] = map[string]any{"packet": true}
				state["entities"].(map[string]any)["work_packets"] = []any{}
				state["entities"].(map[string]any)["submissions"] = []any{}
			}
			if tc.unknown {
				state["entities"].(map[string]any)["unregistered_index"] = []any{}
			}
			instance, err := json.Marshal(state)
			if err != nil {
				t.Fatal(err)
			}
			err = schema.NewEmbeddedValidator().ValidateBytes("loop-state.schema.json", instance)
			if (err == nil) != tc.accept {
				t.Fatalf("accept=%v, validation error: %v", tc.accept, err)
			}
		})
	}
}

func TestLegacySchemaCannotAcceptNewRuntime(t *testing.T) {
	data, err := schema.ReadAsset("loop-state.example.json")
	if err != nil {
		t.Fatal(err)
	}
	var state map[string]any
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatal(err)
	}
	state["schema_version"] = "1.2.0"
	instance, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	if err := schema.NewEmbeddedValidator().ValidateBytes("loop-state-v1.1.schema.json", instance); err == nil {
		t.Fatal("a v1.1 writer must reject a v1.2 instance")
	}
}
