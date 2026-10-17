package cli

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/entroforge/go-system-builder/internal/policy"
	"github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/workpacket"
)

// mainPlanningNext carries the Main-only normal continuation guidance for an
// explicit packet-mode S2/S3/S4 cursor. It is a projection of Runtime facts;
// it never issues a packet or reads authoring files.
type mainPlanningNext struct {
	Action      string
	Automation  string
	IssueAction bool
}

func isMainPlanningViewer(input policy.Input) bool {
	return input.EffectiveAgentID() == "" && input.Runtime.Agent == nil
}

func mainPlanningNextForState(state map[string]any, stage string) *mainPlanningNext {
	if !runtime.PacketEnabled(state) {
		return nil
	}

	var packetID, viewPath, answerPath, phase string
	var binding workpacket.PlanningBinding
	var found, conflicting bool
	switch stage {
	case "S2":
		phase = "design"
		expected, err := runtime.CurrentDesignPacket(state)
		if err == nil {
			binding = expected.Binding
			stored, issued, conflict := currentDesignPlanningPacket(state, expected)
			conflicting = conflict
			if issued {
				packetID, viewPath, answerPath, found = stored.ID, stored.ViewPath, stored.AnswerPath, true
			}
		}
	case "S3":
		phase = "contracts"
		currentBinding, designs, err := runtime.CurrentContractBinding(state)
		if err == nil {
			binding = currentBinding
			stored, issued, conflict := currentContractPlanningPacket(state, currentBinding, designs)
			conflicting = conflict
			if issued {
				packetID, viewPath, answerPath, found = stored.ID, stored.ViewPath, stored.AnswerPath, true
			}
		}
	case "S4":
		phase = "tasks"
		currentBinding, err := runtime.CurrentTaskPlanningBinding(state)
		if err == nil {
			binding = currentBinding
			stored, issued, conflict := currentTaskPlanningPacket(state, currentBinding)
			conflicting = conflict
			if issued {
				packetID, viewPath, answerPath, found = stored.ID, stored.ViewPath, stored.AnswerPath, true
			}
		}
	default:
		return nil
	}

	stageName := stage + " Main"
	if conflicting {
		return unverifiedPlanningRecord(stageName)
	}
	if !found {
		return &mainPlanningNext{
			Action:      fmt.Sprintf("%s questionnaire is being prepared by the Hook listener; use the exact answer path in its delivery. If preparation fails, Main resolves the named missing input", stageName),
			Automation:  fmt.Sprintf("The Hook prepares the current %s questionnaire automatically and preserves saved answers; the author does not run issue or consume", stage),
			IssueAction: false,
		}
	}

	if submission, received, err := runtime.FindSubmission(state, packetID); err != nil {
		return unverifiedPlanningRecord(stageName)
	} else if received {
		if submission.Owner != "Main" || submission.PacketID != packetID || submission.OriginState != "planning" || submission.OriginPhase != phase || submission.OriginEntry != binding.StageEntry {
			return unverifiedPlanningRecord(stageName)
		}
		return &mainPlanningNext{
			Action:     fmt.Sprintf("%s answer for the current Main packet was already received; do not edit or resubmit it, and continue from the existing gate result", stageName),
			Automation: fmt.Sprintf("The current %s Main submission is already registered in Runtime; leave its answer unchanged and let the original gate/CAS determine the next cursor", stage),
		}
	}

	return &mainPlanningNext{
		Action:      fmt.Sprintf("%s packet %s is current: read the exact frozen view %s, fill only the answer and actual artifact paths in %s, and save it through the normal file hook for the original receiver", stageName, packetID, viewPath, answerPath),
		Automation:  fmt.Sprintf("Fill the current %s questionnaire at the exact answer path above and save; the Hook restores missing views automatically while preserving saved answers", stage),
		IssueAction: false,
	}
}

func unverifiedPlanningRecord(stage string) *mainPlanningNext {
	return &mainPlanningNext{
		Action:     fmt.Sprintf("Runtime has a %s packet or submission whose current binding cannot be confirmed; do not edit an answer file or issue another packet, and use the original Runtime recovery path", stage),
		Automation: "Runtime contains a planning record whose current REQ, owner, or stage-entry binding cannot be confirmed; no authoring path or issue action is projected",
	}
}

func currentDesignPlanningPacket(state map[string]any, expected workpacket.DesignPacket) (workpacket.DesignPacket, bool, bool) {
	entities, _ := state["entities"].(map[string]any)
	rows, _ := entities["work_packets"].([]any)
	count := 0
	slotSeen := false
	for _, raw := range rows {
		row, _ := raw.(map[string]any)
		if row["template_id"] != expected.TemplateID {
			continue
		}
		data, err := json.Marshal(row)
		if err != nil {
			continue
		}
		var envelope struct {
			Binding workpacket.PlanningBinding `json:"binding"`
		}
		if json.Unmarshal(data, &envelope) != nil {
			return workpacket.DesignPacket{}, false, true
		}
		if !samePlanningSlot(envelope.Binding, expected.Binding) {
			continue
		}
		slotSeen = true
		if envelope.Binding != expected.Binding {
			return workpacket.DesignPacket{}, false, true
		}
		count++
	}
	if count != 1 {
		return workpacket.DesignPacket{}, false, slotSeen
	}
	stored, found, err := runtime.FindDesignPacket(state, expected.ID)
	if err != nil || !found || !reflect.DeepEqual(stored, expected) {
		return workpacket.DesignPacket{}, false, true
	}
	return stored, true, false
}

func samePlanningSlot(left, right workpacket.PlanningBinding) bool {
	return left.RuntimeID == right.RuntimeID && left.Generation == right.Generation && left.StageEntry == right.StageEntry
}

func currentContractPlanningPacket(state map[string]any, binding workpacket.PlanningBinding, designs []workpacket.Reference) (workpacket.ContractPacket, bool, bool) {
	entities, _ := state["entities"].(map[string]any)
	rows, _ := entities["work_packets"].([]any)
	var candidateIDs []string
	slotSeen := false
	for _, raw := range rows {
		row, _ := raw.(map[string]any)
		if row["template_id"] != "s3-contract-v1" {
			continue
		}
		data, err := json.Marshal(row)
		if err != nil {
			continue
		}
		var envelope struct {
			Binding workpacket.PlanningBinding `json:"binding"`
		}
		if json.Unmarshal(data, &envelope) != nil {
			return workpacket.ContractPacket{}, false, true
		}
		if !samePlanningSlot(envelope.Binding, binding) {
			continue
		}
		slotSeen = true
		var packet workpacket.ContractPacket
		if json.Unmarshal(data, &packet) != nil || !samePlanningSlot(packet.Binding, binding) {
			return workpacket.ContractPacket{}, false, true
		}
		if packet.Binding != binding || !reflect.DeepEqual(packet.Inputs.Designs, designs) || packet.ID == "" {
			return workpacket.ContractPacket{}, false, true
		}
		candidateIDs = append(candidateIDs, packet.ID)
	}
	if len(candidateIDs) != 1 {
		return workpacket.ContractPacket{}, false, slotSeen
	}
	packet, found, err := runtime.FindContractPacket(state, candidateIDs[0])
	if err != nil || !found {
		return workpacket.ContractPacket{}, false, true
	}
	expected, err := workpacket.BuildContracts(binding, packet.Inputs)
	if err != nil || !reflect.DeepEqual(expected, packet) {
		return workpacket.ContractPacket{}, false, true
	}
	return packet, true, false
}

func currentTaskPlanningPacket(state map[string]any, binding workpacket.PlanningBinding) (workpacket.TaskPlanningPacket, bool, bool) {
	entities, _ := state["entities"].(map[string]any)
	rows, _ := entities["work_packets"].([]any)
	var candidateIDs []string
	slotSeen := false
	for _, raw := range rows {
		row, _ := raw.(map[string]any)
		if row["template_id"] != "s4-task-planning-v1" {
			continue
		}
		data, err := json.Marshal(row)
		if err != nil {
			continue
		}
		var envelope struct {
			Binding workpacket.PlanningBinding `json:"binding"`
		}
		if json.Unmarshal(data, &envelope) != nil {
			return workpacket.TaskPlanningPacket{}, false, true
		}
		if !samePlanningSlot(envelope.Binding, binding) {
			continue
		}
		slotSeen = true
		var packet workpacket.TaskPlanningPacket
		if json.Unmarshal(data, &packet) != nil || !samePlanningSlot(packet.Binding, binding) {
			return workpacket.TaskPlanningPacket{}, false, true
		}
		if packet.Binding != binding || packet.ID == "" {
			return workpacket.TaskPlanningPacket{}, false, true
		}
		candidateIDs = append(candidateIDs, packet.ID)
	}
	if len(candidateIDs) != 1 {
		return workpacket.TaskPlanningPacket{}, false, slotSeen
	}
	packet, found, err := runtime.FindTaskPlanningPacket(state, candidateIDs[0])
	if err != nil || !found {
		return workpacket.TaskPlanningPacket{}, false, true
	}
	expected, err := workpacket.BuildTaskPlanning(binding, packet.Inputs)
	if err != nil || !reflect.DeepEqual(expected, packet) {
		return workpacket.TaskPlanningPacket{}, false, true
	}
	return packet, true, false
}
