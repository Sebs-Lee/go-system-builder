package hook_test

import (
	"testing"

	"github.com/entroforge/go-system-builder/internal/hook"
	"github.com/entroforge/go-system-builder/internal/policy"
)

// The observer never blocks and never errors; identification gaps produce a
// silent observation (L3-S7 §8 / L4 §7.4 platform-reality ladder).
func TestPostToolUseObservationLadder(t *testing.T) {
	agents := []hook.AgentRow{
		{ID: "agent-build-1", State: "working"},
		{ID: "agent-qa-1", State: "reading"},
	}

	cases := []struct {
		name       string
		input      policy.Input
		wantRecord bool
		wantAgent  string
	}{
		{
			name: "payload agent_id wins",
			input: policy.Input{
				ToolName: "SendMessage", AgentID: "agent-build-1",
				ToolInput: map[string]any{"message_type": "plan_report", "plan_ref": ".claude/plan-report.json"},
			},
			wantRecord: true, wantAgent: "agent-build-1",
		},
		{
			name: "teammate_name match",
			input: policy.Input{
				ToolName:  "SendMessage",
				ToolInput: map[string]any{"message_type": "plan_report", "plan_ref": ".claude/plan-report.json", "teammate_name": "agent-qa-1"},
			},
			wantRecord: true, wantAgent: "agent-qa-1",
		},
		{
			name: "sole reading agent fallback",
			input: policy.Input{
				ToolName:  "SendMessage",
				ToolInput: map[string]any{"message_type": "plan_report", "plan_ref": ".claude/plan-report.json"},
			},
			wantRecord: true, wantAgent: "agent-qa-1",
		},
		{
			name: "unrelated tool passes silently",
			input: policy.Input{
				ToolName:  "Bash",
				ToolInput: map[string]any{"command": "go test ./..."},
			},
			wantRecord: false,
		},
		{
			name: "unrelated message type passes silently",
			input: policy.Input{
				ToolName: "SendMessage", AgentID: "agent-build-1",
				ToolInput: map[string]any{"message_type": "chitchat"},
			},
			wantRecord: false,
		},
		{
			name: "plan report without authoritative file ref stays silent",
			input: policy.Input{
				ToolName: "SendMessage", AgentID: "agent-build-1",
				ToolInput: map[string]any{"message_type": "plan_report"},
			},
			wantRecord: false,
		},
		{
			name: "unidentifiable sender passes silently",
			input: policy.Input{
				ToolName:  "SendMessage",
				ToolInput: map[string]any{"message_type": "blocker_report"},
			},
			wantRecord: true, wantAgent: "agent-qa-1", // sole waiting agent fallback
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			obs := hook.HandlePostToolUse(tc.input, agents)
			if obs.Recorded != tc.wantRecord {
				t.Fatalf("Recorded = %v, want %v (reason=%q)", obs.Recorded, tc.wantRecord, obs.Reason)
			}
			if tc.wantRecord && obs.AgentID != tc.wantAgent {
				t.Fatalf("AgentID = %q, want %q", obs.AgentID, tc.wantAgent)
			}
		})
	}
}

// Ambiguity must fail silent, never guess.
func TestPostToolUseAmbiguousFallbackSilent(t *testing.T) {
	agents := []hook.AgentRow{
		{ID: "agent-a", State: "reading"},
		{ID: "agent-b", State: "understanding_submitted"},
	}
	obs := hook.HandlePostToolUse(policy.Input{
		ToolName:  "SendMessage",
		ToolInput: map[string]any{"message_type": "plan_report", "plan_ref": ".claude/plan-report.json"},
	}, agents)
	if obs.Recorded {
		t.Fatalf("ambiguous sender must not record, got %q", obs.AgentID)
	}
}
