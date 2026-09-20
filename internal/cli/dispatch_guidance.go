package cli

import (
	"github.com/entroforge/go-system-builder/internal/dispatch"
	"github.com/entroforge/go-system-builder/internal/policy"
	"strings"
)

// Refresh only at lifecycle checkpoints, not on each tool invocation. Errors
// are observations, never new Stop/Idle or ordinary-tool gates.
func applyDispatchGuidance(root string, state map[string]any, event string, input policy.Input, guidance *policy.Guidance) {
	if input.EffectiveAgentID() != "" {
		return
	}
	switch event {
	case "SessionStart":
	case "PostToolUse":
		switch input.ToolName {
		case "Agent", "Task":
			if input.ToolResponse["status"] == "async_launched" {
				return
			}
		case "Bash":
			command, _ := input.ToolInput["command"].(string)
			if !strings.Contains(command, "loop-harness") || !(strings.Contains(command, "task-integrate") || strings.Contains(command, "task-complete") || strings.Contains(command, "register-workgroup")) {
				return
			}
		default:
			return
		}
	default:
		return
	}

	board, err := dispatch.Load(root, state, 0)
	if err != nil {
		guidance.Automation = append(guidance.Automation, "S6 dispatch projection unavailable; inspect `loop-harness s6 status` before choosing another TASK")
		return
	}
	guidance.Automation = append(guidance.Automation, dispatch.Advisory(board)...)
}
