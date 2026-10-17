package repair

import (
	"context"
	"fmt"
	"time"

	"github.com/entroforge/go-system-builder/internal/agentauth"
	runtimepkg "github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/workstate"
)

// RepairResultBlockerExecutionMayEnd verifies the current S9 non-pass blocker
// proof for Stop/Idle projections. The Runtime snapshot is supplied by the
// caller; this helper performs no Runtime read or write. A historical receipt
// is insufficient: the exact owner must still be blocked on its original
// message, the original PLAN authorization and packet sources must remain
// current, the registered Worker execution/HEAD must match, and the evidence
// must not be invalidated or resolved.
func RepairResultBlockerExecutionMayEnd(root string, state map[string]any, packetID, owner string) (bool, error) {
	if root == "" || state == nil || packetID == "" || owner == "" || !runtimepkg.PacketEnabled(state) {
		return false, nil
	}
	if !sameAbsolutePath(root, workspaceMainRoot(root, state)) {
		return false, nil
	}
	packet, found, err := runtimepkg.FindRepairResultPacket(state, packetID)
	if err != nil || !found {
		return false, err
	}
	binding := packet.Binding.Sources
	if binding.OwnerAgentID != owner {
		return false, fmt.Errorf("S9 blocker owner differs from the issued RepairAssignment")
	}

	agent, found, err := repairResultBlockerAgent(state, owner)
	if err != nil || !found {
		return false, err
	}
	if workstate.AgentActivity(agent) != "blocked" ||
		stringField(agent["work_blocked_ref"]) == "" ||
		stringField(agent["blocker_resolved_ref"]) != "" {
		return false, nil
	}
	if event := stringField(agent["last_lifecycle_event"]); event != "" && event != "work_blocked" {
		return false, nil
	}

	receipt, accepted, err := runtimepkg.VerifiedRepairResultBlockerReceipt(root, state, packet.ID, owner)
	if err != nil || !accepted {
		return false, err
	}
	if stringField(agent["work_blocked_ref"]) != receipt.Artifact.MessageRef {
		return false, nil
	}
	if !currentRepairResultBlockerEvidence(state, receipt.ArtifactRef) || repairResultBlockerResolved(state, packet.ID, receipt.ArtifactRef) {
		return false, nil
	}

	// This reader validates the actual original PLAN/activation chain for a
	// blocked Agent. It deliberately does not fabricate a working state.
	if err := agentauth.ValidateBlockedPlanAuthorization(root, state, owner); err != nil {
		return false, fmt.Errorf("S9 blocker has no current original blocked PLAN authorization: %w", err)
	}

	currentPackets, err := runtimepkg.CurrentRepairResultPackets(root, state, owner)
	if err != nil {
		return false, err
	}
	current := 0
	for _, candidate := range currentPackets {
		if candidate.ID == packet.ID {
			current++
		}
	}
	if current != 1 {
		return false, nil
	}
	if err := runtimepkg.VerifyRepairResultPacketIssue(root, state, packet); err != nil {
		return false, fmt.Errorf("S9 blocker packet issue is no longer verifiable: %w", err)
	}
	if err := runtimepkg.VerifyRepairResultPacketCurrent(root, state, packet); err != nil {
		return false, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := verifyAcceptedBlockerBytes(
		ctx, root, state, packet, receipt, receipt.Artifact.ExecutionRoot,
		receipt.Artifact.SourceHead, receipt.Artifact.SubmittedAnswer,
	); err != nil {
		return false, nil
	}
	return true, nil
}

func repairResultBlockerAgent(state map[string]any, owner string) (map[string]any, bool, error) {
	entities, _ := state["entities"].(map[string]any)
	agents, ok := entities["agents"].([]any)
	if !ok {
		return nil, false, fmt.Errorf("Runtime Agent registry is unavailable for S9 blocker projection")
	}
	var selected map[string]any
	for _, raw := range agents {
		row, _ := raw.(map[string]any)
		if row == nil || stringField(row["id"]) != owner {
			continue
		}
		if selected != nil {
			return nil, false, fmt.Errorf("Runtime has ambiguous Agent rows for S9 blocker owner %q", owner)
		}
		selected = row
	}
	return selected, selected != nil, nil
}

func currentRepairResultBlockerEvidence(state map[string]any, artifactRef string) bool {
	rows, ok := state["evidence"].([]any)
	if !ok || artifactRef == "" {
		return false
	}
	count := 0
	for _, raw := range rows {
		row, _ := raw.(map[string]any)
		if row == nil || stringField(row["id"]) != artifactRef {
			continue
		}
		count++
		if row["status"] != "valid" || row["invalidated_by"] != nil ||
			row["invalidation_rule"] != nil || row["invalidation_reason"] != nil {
			return false
		}
	}
	return count == 1
}

func repairResultBlockerResolved(state map[string]any, packetID, artifactRef string) bool {
	rows, _ := state["evidence"].([]any)
	for _, raw := range rows {
		row, _ := raw.(map[string]any)
		if row == nil || row["kind"] != "repair_result_resolution" {
			continue
		}
		refs, ok := row["scope_refs"].([]any)
		if !ok {
			continue
		}
		packetMatch, blockerMatch := false, false
		for _, ref := range refs {
			value, _ := ref.(string)
			packetMatch = packetMatch || value == packetID
			blockerMatch = blockerMatch || value == artifactRef
		}
		if packetMatch || blockerMatch {
			return true
		}
	}
	return false
}

func repairResultStateRevision(state map[string]any) int {
	return integerRepairResult(state["revision"])
}
