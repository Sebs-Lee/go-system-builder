package runtime

import (
	"github.com/entroforge/go-system-builder/internal/controlproof"
)

// RepairResultResumePending reports a recorded resolution awaiting its current
// successor questionnaire. It reads current responsibility facts, not proofs.
func RepairResultResumePending(root string, state map[string]any, owner, assignment string) (bool, error) {
	if !PacketEnabled(state) || owner == "" || assignment == "" {
		return false, nil
	}
	packets, err := RepairResultPackets(state)
	if err != nil {
		return false, err
	}
	current := currentRepairResultCandidates(state, owner, packets)
	for _, packet := range current {
		if packet.Binding.Sources.RepairAssignmentID == assignment {
			return false, nil
		}
	}
	for _, packet := range packets {
		b := packet.Binding.Sources
		if b.OwnerAgentID != owner || b.RepairAssignmentID != assignment {
			continue
		}
		resume, err := RepairResultResumeBinding(root, state, packet.Binding)
		if err != nil {
			return false, err
		}
		if resume != nil {
			return true, nil
		}
	}
	return false, nil
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
