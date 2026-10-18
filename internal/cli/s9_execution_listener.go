package cli

import (
	"github.com/entroforge/go-system-builder/internal/policy"
)

// These are the existing S9 lifecycle phases. Unknown and legacy BUG phases
// retain their original Controller path.
func s9ProfessionalPhase(phase string) bool {
	switch phase {
	case "repair_readback", "planning", "reproducing", "fixing", "targeted_reverification", "ready_for_full_review":
		return true
	}
	return false
}

// Professional questions and actual source documents precede optional
// recovery material. This projection neither prepares work nor guesses a
// native Worker owner from its role.
func s9QuestionnaireReadOrder(root string, state map[string]any, request policy.Input) []string {
	owner := request.EffectiveAgentID()
	main := mainQuestionnaireViewer(request) && (owner != "" || request.AgentType == "" && request.AgentTranscriptPath == "")
	if owner == "" && (!main || request.Event == "SubagentStop" || request.Event == "TeammateIdle") {
		return nil
	}
	var paths []string
	appendPath := func(path string) {
		if path != "" {
			paths = appendUnique(paths, path)
		}
	}
	if main {
		owner = ""
		for _, path := range repairDispatchPaths(root, state) {
			appendPath(path)
		}
		for _, path := range repairVerifierDispatchPaths(root, state) {
			appendPath(path)
		}
	}
	phase := stringValue(mapFieldCLI(state, "lifecycle")["phase"])
	status := stringValue(repairPointerCLI(state)["status"])
	result := phase == "fixing"
	packetPhase := phase == "planning" || phase == "reproducing" || phase == "fixing" && (status == "repairing" || status == "blocked")
	if packets, err := s9CurrentQuestionnaires(root, state, owner, result); err == nil && packetPhase {
		for _, packet := range packets {
			source, _, _ := s9ListenerResultSource(root, state, packet)
			if source != "" {
				appendPath(resolveRootPath(source, packet.answer))
				appendPath(resolveRootPath(source, packet.view))
			}
		}
	}
	pointer := repairPointerCLI(state)
	for _, key := range []string{"contract_ref", "path", "plan_ref"} {
		if ref := stringValue(pointer[key]); ref != "" {
			appendPath(resolveRootPath(root, ref))
		}
	}
	if ref := stringValue(investigationPointer(state)["path"]); ref != "" {
		appendPath(resolveRootPath(root, ref))
	}
	return paths
}
