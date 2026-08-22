// posttooluse.go — the PostToolUse(SendMessage) observer (L3-S7 §8, L4 §7.4).
//
// This event NEVER blocks and never advances lifecycle state on its own:
// the authoritative transitions stay in `runtime agent-event` /
// `runtime review-result submit`. What the observer does is record the
// dispatch envelope the Worker sent (PLAN_REPORT), so the first-write
// barrier (policy rule assignment_write_before_plan) has a fact to read
// and SessionStart recovery can see the plan checkpoint.
//
// Payload identification (platform reality: subagent payloads may not carry
// agent_id): payload agent_id → tool_input.teammate_name matched against
// entities.agents[].id → the sole agent in reading/understanding_submitted
// state as a last-resort fallback. If identification fails, the observation
// is silently skipped (exit 0) — never fabricate an agent binding.
package hook

import (
	"fmt"
	"strings"

	"github.com/entroforge/go-system-builder/internal/policy"
)

// PostToolUseObservation is the outcome the CLI renders.
type PostToolUseObservation struct {
	Recorded  bool   `json:"recorded"`
	AgentID   string `json:"agent_id,omitempty"`
	Message   string `json:"message_type,omitempty"`
	Reason    string `json:"reason,omitempty"`
	SystemMsg string `json:"-"`
}

// HandlePostToolUse observes a PostToolUse(SendMessage) payload. It is
// fail-open by contract: any identification or consistency gap returns an
// observation with Recorded=false and no error.
func HandlePostToolUse(input policy.Input, agents []AgentRow) PostToolUseObservation {
	if input.ToolName != "SendMessage" {
		return PostToolUseObservation{Reason: "not a SendMessage payload"}
	}
	messageType, _ := input.ToolInput["message_type"].(string)
	if messageType == "" {
		messageType, _ = input.ToolInput["type"].(string)
	}
	switch messageType {
	case "plan_report", "blocker_report", "completion_report":
	default:
		return PostToolUseObservation{Reason: "unrelated message type"}
	}
	agentID := identifySender(input, agents)
	if agentID == "" {
		return PostToolUseObservation{Message: messageType, Reason: "sender not identifiable (payload carries no agent_id/teammate_name match)"}
	}
	return PostToolUseObservation{
		Recorded:  true,
		AgentID:   agentID,
		Message:   messageType,
		SystemMsg: fmt.Sprintf("%s observed for %s (authoritative registration stays in runtime agent-event)", messageType, agentID),
	}
}

// AgentRow is the minimal agent fact the observer reads.
type AgentRow struct {
	ID    string
	State string
}

// identifySender applies the three-level identification ladder.
func identifySender(input policy.Input, agents []AgentRow) string {
	if input.AgentID != "" {
		for _, a := range agents {
			if a.ID == input.AgentID {
				return a.ID
			}
		}
	}
	if teammate, _ := input.ToolInput["teammate_name"].(string); teammate != "" {
		for _, a := range agents {
			if a.ID == teammate {
				return a.ID
			}
		}
	}
	// Fallback: exactly one agent waiting on its plan checkpoint.
	var candidates []string
	for _, a := range agents {
		if a.State == "reading" || a.State == "understanding_submitted" {
			candidates = append(candidates, a.ID)
		}
	}
	if len(candidates) == 1 {
		return candidates[0]
	}
	return ""
}

// RenderPostToolUseEnvelope renders the observer output as the hook stdout
// envelope. The decision is always allow-shaped (systemMessage only).
func RenderPostToolUseEnvelope(obs PostToolUseObservation) string {
	msg := obs.SystemMsg
	if msg == "" {
		msg = "PostToolUse observed (no dispatch message captured: " + obs.Reason + ")"
	}
	return fmt.Sprintf(`{"systemMessage": %q}`, strings.ReplaceAll(msg, `"`, `'`))
}
