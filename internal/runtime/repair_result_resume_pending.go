package runtime

import (
	"fmt"

	"github.com/entroforge/go-system-builder/internal/controlproof"
	"github.com/entroforge/go-system-builder/internal/workpacket"
)

// RepairResultResumePending projects retirement from the original resolution
// receipt. It releases that projection only when one exact current successor
// is registered and its complete Resume lineage verifies.
func RepairResultResumePending(root string, state map[string]any, owner, assignment string) (bool, error) {
	if !PacketEnabled(state) || owner == "" || assignment == "" {
		return false, nil
	}
	packets, err := RepairResultPackets(state)
	if err != nil {
		return false, err
	}
	candidates := currentRepairResultCandidates(state, owner, packets)
	if len(candidates) == 0 {
		return false, nil
	}
	journal, err := repairResultResumeJournal(root, state)
	if err != nil {
		return false, err
	}
	verifier, err := newRepairResultResumeVerifier(root, state, journal, packets)
	if err != nil {
		return false, err
	}
	var latest workpacket.RepairResultPacket
	latestSequence := -1
	for _, packet := range candidates {
		if packet.Binding.Sources.RepairAssignmentID != assignment {
			continue
		}
		receipt, resolved, err := verifier.resolution(packet, owner)
		if err != nil {
			return false, fmt.Errorf("verify S9 Resume predecessor resolution: %w", err)
		}
		if !resolved {
			continue
		}
		// A later independent rework attempt has its own authorized source
		// binding and does not inherit this predecessor's Resume requirement.
		execution, err := RepairResultExecution(state, packet)
		if err == nil && stringValue(execution["rework_ref"]) != packet.Binding.ReworkRef {
			continue
		}
		if receipt.AcceptedSequence > latestSequence {
			latest = packet
			latestSequence = receipt.AcceptedSequence
		}
	}
	if latestSequence < 0 {
		return false, nil
	}
	current, err := currentRepairResultPacketsWithVerifier(verifier, owner, candidates)
	if err != nil {
		return false, err
	}
	for _, packet := range current {
		if packet.Binding.Sources.RepairAssignmentID != assignment || !sameRepairResultResumeBase(packet.Binding, latest.Binding) {
			continue
		}
		if packet.Binding.Resume == nil || packet.Binding.Resume.PredecessorPacketID != latest.ID {
			return false, fmt.Errorf("S9 current packet does not continue the latest accepted resolution")
		}
		// The shared request verifier already checked the full unique registered
		// chain and confirmed that this packet is unresolved/current.
		return false, nil
	}
	return true, nil
}

func repairResultResumeJournal(root string, state map[string]any) (controlproof.JournalInspection, error) {
	data, err := controlproof.ReadJournalForState(root, state)
	if err != nil {
		return controlproof.JournalInspection{}, err
	}
	journal, err := controlproof.InspectJournalData(data)
	if err != nil {
		return controlproof.JournalInspection{}, err
	}
	if err := controlproof.ValidateStateJournalPair(state, journal); err != nil {
		return controlproof.JournalInspection{}, err
	}
	return journal, nil
}
