package hook

import (
	"github.com/entroforge/go-system-builder/internal/policy"
	"github.com/entroforge/go-system-builder/internal/repair"
	loopruntime "github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/workspace"
)

// RepairResultAssignmentBinding identifies the S9 Builder responsibility
// from the current repairing pointer's original Assignment→owner projection.
// A missing or ambiguous owner projection remains bound with an empty ID so a
// generic completion report cannot make the known responsibility disappear.
func RepairResultAssignmentBinding(root string, state map[string]any, owner string) (assignmentID string, bound bool) {
	if state == nil || owner == "" || !loopruntime.PacketEnabled(state) {
		return "", false
	}
	lifecycle, _ := state["lifecycle"].(map[string]any)
	if lifecycle["state"] != "bug_resolution" || lifecycle["phase"] != "fixing" {
		return "", false
	}
	review, _ := state["review"].(map[string]any)
	pointer, _ := review["repair"].(map[string]any)
	if pointer == nil || pointer["status"] != "repairing" {
		return "", false
	}
	owners, _ := pointer["assignment_owners"].(map[string]any)
	for id, rawOwner := range owners {
		if id == "" || rawOwner != owner {
			continue
		}
		if bound {
			return "", true
		}
		assignmentID, bound = id, true
	}
	if !bound {
		return "", false
	}
	packets, err := loopruntime.RepairResultPackets(state)
	if err != nil {
		// A malformed packet registry must not turn an issued typed task back
		// into a generic completion path.
		return assignmentID, true
	}
	reviewMap, _ := state["review"].(map[string]any)
	runtimeID := workspace.RuntimeID(state)
	generation := workspace.Generation(state)
	round := repairResultInteger(reviewMap["round"])
	sessionID, _ := pointer["session_id"].(string)
	planRef, _ := pointer["plan_ref"].(string)
	planSHA, _ := pointer["plan_sha256"].(string)
	for _, packet := range packets {
		b := packet.Binding.Sources
		if b.RepairAssignmentID == assignmentID && b.RuntimeID == runtimeID && b.BaselineGeneration == generation &&
			b.ReviewRound == round && b.SessionID == sessionID && b.PlanRef == planRef && b.PlanSHA256 == planSHA {
			return assignmentID, true
		}
	}
	return "", false
}

func repairResultInteger(value any) int {
	switch n := value.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	default:
		return 0
	}
}

// RepairResultWorkComplete is the only S9 stop/idle completion proof. A
// candidate receipt by itself is not enough: the original repair consumer
// must revalidate the current candidate/source/attempt and every owner-owned
// check for this exact RepairAssignment.
func RepairResultWorkComplete(root string, state map[string]any, agent *policy.AgentContext) bool {
	if agent == nil || !agent.RepairResultAssignment || agent.ID == "" || agent.RepairAssignmentID == "" || state == nil {
		return false
	}
	assignmentID, bound := RepairResultAssignmentBinding(root, state, agent.ID)
	if !bound || assignmentID != agent.RepairAssignmentID {
		return false
	}
	packets, err := loopruntime.CurrentRepairResultPackets(root, state, agent.ID)
	if err != nil {
		return false
	}
	var selectedID string
	for _, packet := range packets {
		binding := packet.Binding.Sources
		if binding.OwnerAgentID != agent.ID || binding.RepairAssignmentID != assignmentID {
			continue
		}
		if selectedID != "" {
			return false
		}
		if loopruntime.VerifyRepairResultPacketIssue(root, state, packet) != nil {
			return false
		}
		if loopruntime.VerifyRepairResultPacketCurrent(root, state, packet) != nil {
			// It is no longer a fresh write responsibility. The historical
			// final-receipt path below may still prove this exact attempt complete.
			continue
		}
		selectedID = packet.ID
	}
	if selectedID != "" {
		complete, err := repair.RepairResultWorkComplete(root, state, selectedID, agent.ID)
		return err == nil && complete
	}

	// A completed execution can leave the fresh selector after its exact
	// accepted final Result is durable. Preserve only that already-issued
	// responsibility: locate packets for the same current Session/Plan and
	// Assignment, then require both the original verified final receipt and the
	// domain's current-attempt completion proof. This branch grants no new save
	// or delivery permission.
	review, _ := state["review"].(map[string]any)
	pointer, _ := review["repair"].(map[string]any)
	runtimeID := workspace.RuntimeID(state)
	generation := workspace.Generation(state)
	round := repairResultInteger(review["round"])
	sessionID, _ := pointer["session_id"].(string)
	planRef, _ := pointer["plan_ref"].(string)
	planSHA, _ := pointer["plan_sha256"].(string)
	all, err := loopruntime.RepairResultPackets(state)
	if err != nil {
		return false
	}
	var finalPacketID string
	for _, packet := range all {
		binding := packet.Binding.Sources
		if binding.OwnerAgentID != agent.ID || binding.RepairAssignmentID != assignmentID ||
			binding.RuntimeID != runtimeID || binding.BaselineGeneration != generation || binding.ReviewRound != round ||
			binding.SessionID != sessionID || binding.PlanRef != planRef || binding.PlanSHA256 != planSHA {
			continue
		}
		if loopruntime.VerifyRepairResultPacketIssue(root, state, packet) != nil {
			return false
		}
		if _, accepted, receiptErr := loopruntime.VerifiedRepairResultReceipt(root, state, packet.ID, agent.ID); receiptErr != nil {
			return false
		} else if !accepted {
			continue
		}
		if finalPacketID != "" {
			return false
		}
		finalPacketID = packet.ID
	}
	if finalPacketID == "" {
		return false
	}
	complete, err := repair.RepairResultWorkComplete(root, state, finalPacketID, agent.ID)
	return err == nil && complete
}
