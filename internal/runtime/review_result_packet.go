package runtime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"

	"github.com/entroforge/go-system-builder/internal/reviewmeta"
	"github.com/entroforge/go-system-builder/internal/workpacket"
	"github.com/entroforge/go-system-builder/internal/workstate"
)

const reviewResultPacketTemplateID = "s7-review-result-v1"

// FindReviewResultPacket loads one S7 packet from the original work_packets
// registry. It treats duplicate IDs and malformed or noncanonical S7 rows as
// errors; packet rows are never silently projected into the S5 Packet type.
func FindReviewResultPacket(state map[string]any, id string) (workpacket.ReviewResultPacket, bool, error) {
	if id == "" {
		return workpacket.ReviewResultPacket{}, false, fmt.Errorf("S7 ReviewResult packet ID is required")
	}
	packets, err := ReviewResultPackets(state)
	if err != nil {
		return workpacket.ReviewResultPacket{}, false, err
	}
	var found *workpacket.ReviewResultPacket
	for i := range packets {
		if packets[i].ID != id {
			continue
		}
		if found != nil {
			return workpacket.ReviewResultPacket{}, false, fmt.Errorf("duplicate S7 ReviewResult packet ID %q", id)
		}
		copyPacket := packets[i]
		found = &copyPacket
	}
	if found == nil {
		return workpacket.ReviewResultPacket{}, false, nil
	}
	return *found, true, nil
}

// ReviewResultPackets returns all canonical S7 packet rows from the existing
// work_packets index, sorted by packet ID. It validates global packet ID
// uniqueness across subtypes so callers cannot accidentally select an
// ambiguous identity.
func ReviewResultPackets(state map[string]any) ([]workpacket.ReviewResultPacket, error) {
	entities, ok := state["entities"].(map[string]any)
	if !ok {
		if !PacketEnabled(state) {
			return []workpacket.ReviewResultPacket{}, nil
		}
		return nil, fmt.Errorf("Runtime entities are unavailable")
	}
	rawRows, exists := entities["work_packets"]
	if !exists || rawRows == nil {
		return []workpacket.ReviewResultPacket{}, nil
	}
	rows, ok := rawRows.([]any)
	if !ok {
		return nil, fmt.Errorf("Runtime work_packets index is not an array")
	}
	seenIDs := make(map[string]bool, len(rows))
	packets := make([]workpacket.ReviewResultPacket, 0)
	for index, raw := range rows {
		row, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("Runtime work_packets[%d] is not an object", index)
		}
		id := stringValue(row["packet_id"])
		if id == "" {
			return nil, fmt.Errorf("Runtime work_packets[%d] has no packet_id", index)
		}
		if seenIDs[id] {
			return nil, fmt.Errorf("duplicate Runtime work packet ID %q", id)
		}
		seenIDs[id] = true
		if row["template_id"] != reviewResultPacketTemplateID {
			continue
		}
		packet, err := decodeReviewResultPacket(row)
		if err != nil {
			return nil, fmt.Errorf("decode S7 ReviewResult packet %s: %w", id, err)
		}
		if packet.ID != id {
			return nil, fmt.Errorf("S7 ReviewResult packet key differs from its row ID")
		}
		rebuilt, err := workpacket.RebuildReviewResult(packet)
		if err != nil || !reflect.DeepEqual(rebuilt, packet) {
			return nil, fmt.Errorf("S7 ReviewResult packet %s differs from its canonical rebuild", id)
		}
		packets = append(packets, packet)
	}
	if len(packets) > 0 && !PacketEnabled(state) {
		return nil, fmt.Errorf("S7 ReviewResult packets require the explicit packet contract")
	}
	sort.Slice(packets, func(i, j int) bool { return packets[i].ID < packets[j].ID })
	return packets, nil
}

// CurrentReviewResultPackets returns already-issued packets whose frozen
// tuple still matches the current Runtime ReviewPlan pointer and Assignment
// projection. Consumed rows remain visible for durable receipt lookup and
// exact replay; callers granting a new answer write must separately require
// an unconsumed Assignment. It never derives a Plan from packet inputs.
func CurrentReviewResultPackets(state map[string]any, owner string) ([]workpacket.ReviewResultPacket, error) {
	all, err := ReviewResultPackets(state)
	if err != nil {
		return nil, err
	}
	if len(all) == 0 {
		return []workpacket.ReviewResultPacket{}, nil
	}
	if !PacketEnabled(state) {
		return nil, fmt.Errorf("S7 ReviewResult packets require the explicit packet contract")
	}
	ptr := reviewmeta.FromState(state)
	if ptr == nil {
		return nil, fmt.Errorf("current ReviewPlan pointer is unavailable")
	}
	reviewState, _ := state["review"].(map[string]any)
	round, roundErr := integerField(reviewState, "round")
	baseline, _ := state["baseline"].(map[string]any)
	generation, generationErr := integerField(baseline, "generation")
	runtimeID := stringValue(state["runtime_id"])
	if roundErr != nil || generationErr != nil || round < 1 || generation < 1 || runtimeID == "" || round != ptr.ReviewRound {
		return nil, fmt.Errorf("current ReviewPlan Runtime tuple is incomplete")
	}
	assignments, _ := reviewState["assignments"].(map[string]any)
	if assignments == nil {
		return nil, fmt.Errorf("current ReviewPlan assignment projection is unavailable")
	}
	current := make([]workpacket.ReviewResultPacket, 0, len(all))
	seenAssignments := map[string]bool{}
	for _, packet := range all {
		binding := packet.Binding
		if binding.RuntimeID != runtimeID || binding.BaselineGeneration != generation ||
			binding.ReviewPlanID != ptr.PlanID || binding.ReviewPlanRef != ptr.Path ||
			binding.ReviewPlanSHA256 != ptr.SHA256 || binding.ReviewPlanRevision != ptr.Revision ||
			binding.ReviewRound != ptr.ReviewRound || binding.ReviewRound != round ||
			(binding.VerificationArtifactDigest != nil && *binding.VerificationArtifactDigest != ptr.VerificationArtifactDigest) {
			continue
		}
		row, _ := assignments[binding.AssignmentID].(map[string]any)
		if row == nil || row["agent_id"] != binding.OwnerAgentID {
			continue
		}
		switch workstate.ReviewAssignmentActivity(row) {
		case "dispatched", "blocked":
			if row["result_ref"] != nil {
				continue
			}
		case "consumed":
			if row["result_ref"] == nil {
				continue
			}
		default:
			continue
		}
		claimIDs, ok := checkpointStringSlice(row["claim_ids"])
		if !ok || !sameStrings(claimIDs, binding.ClaimIDs) {
			continue
		}
		if owner != "" && binding.OwnerAgentID != owner {
			continue
		}
		if seenAssignments[binding.AssignmentID] {
			return nil, fmt.Errorf("current S7 ReviewResult Assignment %s has multiple issued packets", binding.AssignmentID)
		}
		seenAssignments[binding.AssignmentID] = true
		current = append(current, packet)
	}
	sort.Slice(current, func(i, j int) bool { return current[i].ID < current[j].ID })
	return current, nil
}

func decodeReviewResultPacket(row map[string]any) (workpacket.ReviewResultPacket, error) {
	data, err := json.Marshal(row)
	if err != nil {
		return workpacket.ReviewResultPacket{}, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var packet workpacket.ReviewResultPacket
	if err := decoder.Decode(&packet); err != nil {
		return workpacket.ReviewResultPacket{}, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return workpacket.ReviewResultPacket{}, fmt.Errorf("unexpected trailing JSON value")
		}
		return workpacket.ReviewResultPacket{}, err
	}
	if packet.TemplateID != reviewResultPacketTemplateID {
		return workpacket.ReviewResultPacket{}, fmt.Errorf("unexpected packet subtype %q", packet.TemplateID)
	}
	return packet, nil
}

func checkpointStringSlice(value any) ([]string, bool) {
	switch values := value.(type) {
	case []string:
		return append([]string(nil), values...), true
	case []any:
		result := make([]string, len(values))
		for i, raw := range values {
			value, ok := raw.(string)
			if !ok {
				return nil, false
			}
			result[i] = value
		}
		return result, true
	default:
		return nil, false
	}
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
