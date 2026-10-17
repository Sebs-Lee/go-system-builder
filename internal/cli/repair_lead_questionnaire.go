package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/entroforge/go-system-builder/internal/policy"
	"github.com/entroforge/go-system-builder/internal/repair"
	"github.com/entroforge/go-system-builder/internal/runtime"
)

// stageRepairLeadQuestionnaireContext gives Main or the dispatched independent
// verifier the exact existing S9 DTO path for the current responsibility.
func stageRepairLeadQuestionnaireContext(root string, request policy.Input, snapshot runtime.Snapshot) string {
	lifecycle := mapFieldCLI(snapshot.State, "lifecycle")
	if stringValue(lifecycle["state"]) != "bug_resolution" {
		return ""
	}
	if !mainQuestionnaireViewer(request) && request.EffectiveAgentID() == "" {
		return ""
	}
	phase, phaseErr := currentRepairLeadPhase(root, snapshot.State)
	if phaseErr != nil {
		return "S9 current TargetedReverification obligations cannot be resolved: " + phaseErr.Error()
	}
	owner := request.EffectiveAgentID()
	main := owner == "" || owner == "Main"
	prepare := repairLeadPreparationEvent(request.Event)
	readOnly := repairLeadReadOnlyEvent(request.Event)

	switch phase {
	case "repair_readback":
		if !main {
			return ""
		}
		state := snapshot.State
		if prepare {
			var err error
			state, err = prepareRepairReadback(root, state)
			if err != nil {
				return "S9 approved Contract readback could not prepare the original RepairSession/RepairPlan: " + err.Error()
			}
		}
		pointer := repairPointerCLI(state)
		contract := repair.ContractRef{Path: stringValue(pointer["contract_ref"]), SHA256: stringValue(pointer["contract_sha256"])}
		session := repair.ArtifactRef{Path: stringValue(pointer["path"]), SHA256: stringValue(pointer["sha256"])}
		plan := repair.ArtifactRef{Path: stringValue(pointer["plan_ref"]), SHA256: stringValue(pointer["plan_sha256"])}
		if contract.Path == "" || session.Path == "" || plan.Path == "" {
			return "S9 repair_readback is missing a current approved Contract, RepairSession, or RepairPlan reference. Restore the original investigation/S9 source; no path was guessed."
		}
		contractPath, _ := absoluteQuestionnairePath(root, contract.Path)
		sessionPath, _ := absoluteQuestionnairePath(root, session.Path)
		planPath, _ := absoluteQuestionnairePath(root, plan.Path)
		if readOnly {
			return fmt.Sprintf("S9 repair_readback is read-only: approved Contract %s; current RepairSession %s; current RepairPlan %s. Stop/Idle did not create or change any artifact.", contractPath, sessionPath, planPath)
		}
		return fmt.Sprintf("S9 repair_readback: review the approved Contract at %s, current RepairSession at %s, and Contract-derived RepairPlan at %s. Session/Plan were produced by their original Runtime APIs; the current RepairAssignments and their exact Builder PlanReport/RepairResult questions are delivered by the original listeners. Main owns dispatch and integration; do not hand-author another plan.", contractPath, sessionPath, planPath)

	case "impact_reconciliation":
		if !main {
			return ""
		}
		path, err := repairLeadQuestionnairePath(root, snapshot.State, "change-impact.json")
		if err != nil {
			return "S9 ChangeImpact questionnaire path is unavailable from current Runtime: " + err.Error()
		}
		if prepare {
			if _, err := prepareRepairImpactQuestionnaire(root, snapshot.State); err != nil {
				return fmt.Sprintf("S9 ChangeImpact questionnaire could not be prepared at %s: %v", path, err)
			}
		}
		if readOnly {
			if questionnaireMissing(path) {
				return "S9 ChangeImpact questionnaire is not prepared at " + path + "; this Stop/Idle event is read-only. The next SessionStart or PreToolUse prepares the original request DTO."
			}
			return "S9 ChangeImpact answer: " + path + ". This Stop/Idle event is read-only; the existing request DTO remains available for Main to complete and save."
		}
		return "S9 ChangeImpact answer: " + path + ". Runtime/REQ/generation, source Case, the exact Session diff, and the Contract-level reverification ID are supplied from the original sources. Complete the professional change types, per-change disposition/rationale, and evidence disposition, then save this exact request. The original CreateChangeImpact and CommitChangeImpact consumers run on save."

	case "targeted_reverification":
		pointer := repairPointerCLI(snapshot.State)
		blockedResume := stringValue(pointer["status"]) == "blocked" && stringValue(pointer["failure_route"]) == "blocked"
		selections, selectionErr := repairTargetedSelections(root, snapshot.State)
		if selectionErr != nil {
			return "S9 independent TargetedReverification source is unavailable: " + selectionErr.Error()
		}
		missing, missingErr := repairMissingTargetedIDs(root, repairPointerCLI(snapshot.State))
		if missingErr != nil {
			return "S9 independent TargetedReverification obligations are unavailable: " + missingErr.Error()
		}
		if main {
			if len(selections) == 0 {
				return "S9 targeted_reverification has no current independently dispatched verifier in Runtime assignment_owners. Main must resolve the original repair dispatch; no verifier identity or answer path was guessed."
			}
			if err := repairTargetedRequestMappingError(selections, missing); err != nil {
				return "S9 targeted_reverification has no current mapping from the ChangeImpact's required IDs to the dispatched verifier assignments: " + err.Error() + ". Main must reconcile the original ChangeImpact/dispatch source; no answer path was guessed."
			}
			var b strings.Builder
			for _, selection := range repairTargetedOutstandingSelections(selections, missing) {
				path, pathErr := repairTargetedQuestionnairePath(root, snapshot.State, selection.performingAssignmentID)
				if pathErr != nil {
					fmt.Fprintf(&b, "S9 independent verifier %s assignment=%s has no resolvable answer path: %v\n", selection.verifierOwner, selection.performingAssignmentID, pathErr)
					continue
				}
				if prepare {
					if _, _, err := prepareRepairTargetedQuestionnaire(root, snapshot.State, selection); err != nil {
						fmt.Fprintf(&b, "S9 independent verifier %s assignment=%s questionnaire preparation failed at %s: %v\n", selection.verifierOwner, selection.performingAssignmentID, path, err)
						continue
					}
				}
				if readOnly && questionnaireMissing(path) {
					fmt.Fprintf(&b, "S9 independent verifier %s assignment=%s: TargetedReverificationRequest is not prepared at %s; Stop/Idle is read-only. The verifier's next SessionStart or PreToolUse prepares it.\n", selection.verifierOwner, selection.performingAssignmentID, path)
					continue
				}
				readOnlyNote := ""
				if readOnly {
					readOnlyNote = " Stop/Idle is read-only."
				}
				if blockedResume {
					fmt.Fprintf(&b, "S9 independent verifier %s assignment=%s: resolve the recorded environmental blocker, then update the same TargetedReverificationRequest at %s with the actual continuity basis and replay result. The original Resume/Create/Commit flow runs on save; Main does not author the independent result.%s\n", selection.verifierOwner, selection.performingAssignmentID, path, readOnlyNote)
				} else {
					fmt.Fprintf(&b, "S9 independent verifier %s assignment=%s: fill and save its TargetedReverificationRequest at %s; Main does not author the independent result.%s\n", selection.verifierOwner, selection.performingAssignmentID, path, readOnlyNote)
				}
			}
			return strings.TrimSpace(b.String())
		}
		for _, selection := range repairTargetedOutstandingSelections(selections, missing) {
			if selection.verifierOwner != owner {
				continue
			}
			if err := repairTargetedRequestMappingError(selections, missing); err != nil {
				return "S9 TargetedReverification assignment cannot be mapped to the current ChangeImpact obligation: " + err.Error() + ". Main must reconcile the original source; no answer was claimed."
			}
			path, err := repairTargetedQuestionnairePath(root, snapshot.State, selection.performingAssignmentID)
			if err != nil {
				return fmt.Sprintf("S9 independent verifier %s assignment=%s has no resolvable answer path: %v", owner, selection.performingAssignmentID, err)
			}
			if prepare {
				if _, _, err := prepareRepairTargetedQuestionnaire(root, snapshot.State, selection); err != nil {
					return fmt.Sprintf("S9 TargetedReverification questionnaire preparation failed at %s: %v", path, err)
				}
			}
			if readOnly && questionnaireMissing(path) {
				return "S9 independent TargetedReverification answer is not prepared at " + path + "; Stop/Idle is read-only. The next verifier SessionStart or PreToolUse prepares this original request DTO."
			}
			readOnlyNote := ""
			if readOnly {
				readOnlyNote = " This Stop/Idle event is read-only."
			}
			if blockedResume {
				return fmt.Sprintf("S9 independent verifier %s assignment=%s: resolve the recorded environmental blocker, then update the same TargetedReverificationRequest at %s with the actual continuity basis and replay result, and save. The original Resume/Create/Commit flow runs on save; Main does not author the independent result.%s", owner, selection.performingAssignmentID, path, readOnlyNote)
			}
			return fmt.Sprintf("S9 independent verifier %s assignment=%s: record the actual assertion results, scope compliance, continuity evidence, verdict, and failure class in %s, then save. The request is bound to the dispatched repair and verifier assignments; the original TargetedReverification consumer checks the independent work.%s", owner, selection.performingAssignmentID, path, readOnlyNote)
		}
		return ""

	case "ready_for_full_review":
		if !main {
			return ""
		}
		if readOnly {
			return "S9 has no remaining professional Main answer. Stop/Idle is read-only; the next Main SessionStart or PreToolUse will pass the current S9 references through the original Changeset/Handoff Runtime consumers and report the committed Handoff and S7 seed paths."
		}
		if !prepare {
			return "S9 has no remaining professional Main answer. The next Main SessionStart or PreToolUse will pass the current S9 references through the original Changeset/Handoff Runtime consumers and report the committed Handoff and S7 seed paths."
		}
		handoffRef, handedOff, err := completeRepairHandoff(root)
		if err != nil {
			return "S9 automatic RepairHandoff could not complete through the original Runtime consumer: " + err.Error()
		}
		handoffPath, _ := absoluteQuestionnairePath(root, handoffRef.Path)
		seedRef := repairPointerCLI(handedOff.State)["review_plan_seed_ref"]
		seedPath, pathErr := absoluteQuestionnairePath(root, stringValue(seedRef))
		if pathErr != nil {
			return fmt.Sprintf("S9 RepairHandoff committed at %s. Runtime did not expose its S7 seed path: %v", handoffPath, pathErr)
		}
		return fmt.Sprintf("S9 RepairHandoff was committed through the original Runtime consumer at %s; the next complete S7 round is seeded from %s. No professional Main answer or human approval was inferred.", handoffPath, seedPath)
	}
	return ""
}

// repairLeadQuestionnairePaths returns only the deterministic paths for the
// current S9 answer phase. It is read-only so Hook fallback observers can
// inspect a saved questionnaire without creating one or parsing prose.
func repairLeadQuestionnairePaths(root string, state map[string]any) []string {
	lifecycle := mapFieldCLI(state, "lifecycle")
	if stringValue(lifecycle["state"]) != "bug_resolution" {
		return nil
	}
	phase, err := currentRepairLeadPhase(root, state)
	if err != nil {
		return nil
	}
	var paths []string
	switch phase {
	case "impact_reconciliation":
		if path, err := repairLeadQuestionnairePath(root, state, "change-impact.json"); err == nil {
			paths = append(paths, path)
		}
	case "targeted_reverification":
		selections, err := repairTargetedSelections(root, state)
		if err != nil {
			return nil
		}
		missing, err := repairMissingTargetedIDs(root, repairPointerCLI(state))
		if err != nil || repairTargetedRequestMappingError(selections, missing) != nil {
			return nil
		}
		for _, selection := range repairTargetedOutstandingSelections(selections, missing) {
			if path, pathErr := repairTargetedQuestionnairePath(root, state, selection.performingAssignmentID); pathErr == nil {
				paths = append(paths, path)
			}
		}
	}
	return paths
}

func currentRepairLeadPhase(root string, state map[string]any) (string, error) {
	pointer := repairPointerCLI(state)
	switch stringValue(pointer["status"]) {
	case "impact_reconciliation":
		return stringValue(pointer["status"]), nil
	case "targeted_reverification":
		missing, err := repairMissingTargetedIDs(root, pointer)
		if err != nil {
			return "", err
		}
		if len(missing) == 0 {
			return "", fmt.Errorf("Runtime status is targeted_reverification but no required reverification remains; preserve the original Runtime source for reconciliation")
		}
		return "targeted_reverification", nil
	case "blocked":
		if stringValue(pointer["failure_route"]) != "blocked" {
			return "", nil
		}
		missing, err := repairMissingTargetedIDs(root, pointer)
		if err != nil {
			return "", err
		}
		if len(missing) == 0 {
			return "", fmt.Errorf("Runtime status is blocked but no required TargetedReverification remains; preserve the original Runtime source for reconciliation")
		}
		return "targeted_reverification", nil
	case "ready_for_full_review":
		missing, err := repairMissingTargetedIDs(root, pointer)
		if err != nil {
			return "", err
		}
		if len(missing) > 0 {
			return "targeted_reverification", nil
		}
		return "ready_for_full_review", nil
	}
	if stringValue(mapFieldCLI(state, "lifecycle")["phase"]) == "repair_readback" {
		return "repair_readback", nil
	}
	return "", nil
}

func repairMissingTargetedIDs(root string, pointer map[string]any) ([]string, error) {
	committed := map[string]bool{}
	for _, ref := range repairArtifactRefs(pointer, "targeted_reverification_artifacts", "", "") {
		value, err := repair.ValidateTargetedReverification(root, ref)
		if err != nil {
			return nil, fmt.Errorf("read current TargetedReverification %s: %w", ref.Path, err)
		}
		if value.Result == "pass" && value.ScopeCompliance == "pass" {
			committed[value.ReverificationID] = true
		}
	}
	missing := []string{}
	for _, id := range stringSliceAny(pointer["required_reverification_ids"]) {
		if !committed[id] {
			missing = append(missing, id)
		}
	}
	return missing, nil
}

func repairLeadPreparationEvent(event string) bool {
	return event == "SessionStart" || event == "PreToolUse" || event == "SubagentStart"
}

func repairLeadReadOnlyEvent(event string) bool {
	return event == "Stop" || event == "SubagentStop" || event == "TeammateIdle" || event == "Idle"
}

func prepareRepairReadback(root string, state map[string]any) (map[string]any, error) {
	pointer := repairPointerCLI(state)
	statePath := filepath.Join(root, ".claude", "loop-state.json")
	journalPath := filepath.Join(root, ".claude", "loop-events.jsonl")
	runtimeID, generation, contract, err := repairReadbackContractSource(root, state)
	if err != nil {
		return nil, err
	}
	suffix, err := repairReadbackIDSuffix(runtimeID, generation, contract.CaseID, contract.ContractID, contract.Ref)
	if err != nil {
		return nil, err
	}
	sameContract := stringValue(pointer["case_id"]) == contract.CaseID && stringValue(pointer["contract_id"]) == contract.ContractID &&
		stringValue(pointer["contract_ref"]) == contract.Ref.Path && stringValue(pointer["contract_sha256"]) == contract.Ref.SHA256
	if stringValue(pointer["contract_id"]) == "" && stringValue(pointer["contract_ref"]) == contract.Ref.Path && stringValue(pointer["contract_sha256"]) == contract.Ref.SHA256 {
		sameContract = stringValue(pointer["case_id"]) == contract.CaseID
	}
	if stringValue(pointer["path"]) == "" || !sameContract {
		_, _, _, err := repair.OpenRepairSession(root, statePath, journalPath, repair.OpenSessionRequest{
			RuntimeRequest: repair.RuntimeRequest{ExpectedRevision: -1, Actor: "Main"},
			SessionID:      "repair-session-" + suffix, CreatedBy: "Main",
		})
		if err != nil {
			return nil, err
		}
	}
	snapshot, err := runtime.NewStore(statePath, journalPath).Snapshot()
	if err != nil {
		return nil, err
	}
	pointer = repairPointerCLI(snapshot.State)
	if stringValue(pointer["plan_ref"]) == "" && stringValue(pointer["status"]) == "contract_ready" {
		runtimeID := strings.TrimSpace(stringValue(snapshot.State["runtime_id"]))
		generation := integerValue(mapFieldCLI(snapshot.State, "baseline")["generation"])
		suffix, err := repairReadbackIDSuffix(runtimeID, generation, stringValue(pointer["case_id"]), stringValue(pointer["contract_id"]), repair.ContractRef{
			Path: stringValue(pointer["contract_ref"]), SHA256: stringValue(pointer["contract_sha256"]),
		})
		if err != nil {
			return nil, err
		}
		_, _, _, err = repair.CompileRepairPlan(root, statePath, journalPath, repair.CompilePlanRequest{
			RuntimeRequest: repair.RuntimeRequest{ExpectedRevision: -1, Actor: "Main"},
			PlanID:         "repair-plan-" + suffix, CreatedBy: "Main",
		})
		if err != nil {
			return nil, err
		}
		snapshot, err = runtime.NewStore(statePath, journalPath).Snapshot()
		if err != nil {
			return nil, err
		}
	}
	return snapshot.State, nil
}

func repairReadbackContractSource(root string, state map[string]any) (string, int, repair.ApprovedContract, error) {
	runtimeID := strings.TrimSpace(stringValue(state["runtime_id"]))
	generation := integerValue(mapFieldCLI(state, "baseline")["generation"])
	if runtimeID == "" || generation < 1 {
		return "", 0, repair.ApprovedContract{}, fmt.Errorf("Runtime id and positive baseline generation are required before S9 can read back")
	}
	investigation := mapFieldCLI(state, "review")["investigation"]
	current, _ := investigation.(map[string]any)
	if current == nil || stringValue(current["status"]) != "contract_approved" {
		return "", 0, repair.ApprovedContract{}, fmt.Errorf("current approved InvestigationCase and RepairContract are required before S9 can read back")
	}
	contract, err := repair.ValidateApprovedContractRef(root, repair.ContractRef{
		Path: stringValue(current["repair_contract_ref"]), SHA256: stringValue(current["repair_contract_sha256"]),
	})
	if err != nil {
		return "", 0, repair.ApprovedContract{}, err
	}
	if contract.CaseID == "" || contract.ContractID == "" || stringValue(current["case_id"]) != contract.CaseID {
		return "", 0, repair.ApprovedContract{}, fmt.Errorf("current InvestigationCase and approved RepairContract identities do not match")
	}
	return runtimeID, generation, contract, nil
}

func repairReadbackIDSuffix(runtimeID string, generation int, caseID, contractID string, contractRef repair.ContractRef) (string, error) {
	if strings.TrimSpace(runtimeID) == "" || generation < 1 || strings.TrimSpace(caseID) == "" || strings.TrimSpace(contractID) == "" || strings.TrimSpace(contractRef.Path) == "" || strings.TrimSpace(contractRef.SHA256) == "" {
		return "", fmt.Errorf("Runtime, baseline generation, Case, and approved Contract reference are required for S9 Session/Plan IDs")
	}
	digest := sha256.Sum256([]byte(fmt.Sprintf("%s\x00%d\x00%s\x00%s\x00%s\x00%s", runtimeID, generation, caseID, contractID, contractRef.Path, contractRef.SHA256)))
	return hex.EncodeToString(digest[:]), nil
}

func repairLeadQuestionnairePath(root string, state map[string]any, name string) (string, error) {
	runtimeID := strings.TrimSpace(stringValue(state["runtime_id"]))
	generation := integerValue(mapFieldCLI(state, "baseline")["generation"])
	round := integerValue(mapFieldCLI(state, "review")["round"])
	segment := dispatchSlug(runtimeID)
	if segment == "" || generation < 1 || round < 1 || name == "" {
		return "", fmt.Errorf("current Runtime id, baseline generation, review round, or questionnaire name is missing")
	}
	pointer := repairPointerCLI(state)
	source := []string{
		stringValue(pointer["case_id"]), stringValue(pointer["contract_id"]), stringValue(pointer["contract_sha256"]),
		stringValue(pointer["session_id"]), stringValue(pointer["plan_ref"]), stringValue(pointer["plan_sha256"]),
	}
	if source[0] == "" || source[3] == "" || source[4] == "" || source[5] == "" {
		return "", fmt.Errorf("current approved Case, RepairSession, and RepairPlan source are required to identify the S9 questionnaire")
	}
	sourceDigest := sha256.Sum256([]byte(strings.Join(source, "\x00")))
	path := filepath.Join(root, ".claude", "submissions", "s9", segment, fmt.Sprintf("g%d", generation), fmt.Sprintf("r%d", round), hex.EncodeToString(sourceDigest[:]), name)
	return filepath.Abs(path)
}

func repairTargetedQuestionnairePath(root string, state map[string]any, assignmentID string) (string, error) {
	segment := dispatchSlug(assignmentID)
	if segment == "" {
		return "", fmt.Errorf("current verifier assignment id is empty")
	}
	return repairLeadQuestionnairePath(root, state, "targeted-reverification-"+segment+".json")
}

func prepareRepairImpactQuestionnaire(root string, state map[string]any) (string, error) {
	path, err := repairLeadQuestionnairePath(root, state, "change-impact.json")
	if err != nil {
		return "", err
	}
	pointer := repairPointerCLI(state)
	sessionRef, err := repairPointerRef(pointer, "path", "sha256", "current RepairSession")
	if err != nil {
		return "", err
	}
	session, err := repair.ValidateRepairSession(root, sessionRef)
	if err != nil {
		return "", fmt.Errorf("read current RepairSession: %w", err)
	}
	changed, err := repair.ComputeSessionChangeset(root, session)
	if err != nil {
		return "", fmt.Errorf("derive exact current Session diff: %w", err)
	}
	boundReq := mapFieldCLI(state, "bound_req")
	caseID := stringValue(pointer["case_id"])
	if caseID == "" {
		return "", fmt.Errorf("current S9 Runtime pointer has no source Case id")
	}
	reverificationIDs, err := repairImpactReverificationIDs(state)
	if err != nil {
		return "", err
	}
	request := repair.ChangeImpactRequest{
		ImpactID:  "impact-" + dispatchSlug(session.SessionID),
		RuntimeID: stringValue(state["runtime_id"]), ReqID: stringValue(boundReq["id"]),
		BaselineGeneration: integerValue(mapFieldCLI(state, "baseline")["generation"]),
		SourceCaseIDs:      []string{caseID}, ChangeTypes: []string{}, ChangedArtifacts: changed,
		Decisions: []repair.ImpactDecision{}, EscalationLevel: "assignment",
		InvalidatedEvidenceIDs: []string{}, SupersededEvidenceIDs: []string{}, RetainedEvidenceIDs: []string{},
		RequiredReverificationIDs: reverificationIDs, AnalyzedBy: "Main",
	}
	return path, writeRepairLeadDraft(path, request)
}

type repairTargetedSelection struct {
	originalAssignmentID   string
	performingAssignmentID string
	repairOwner            string
	verifierOwner          string
	impact                 repair.ChangeImpact
}

func repairImpactReverificationIDs(state map[string]any) ([]string, error) {
	sessionID := stringValue(repairPointerCLI(state)["session_id"])
	if sessionID == "" {
		return nil, fmt.Errorf("current RepairSession is required to identify the Contract reverification obligation")
	}
	impactID := "impact-" + dispatchSlug(sessionID)
	digest := sha256.Sum256([]byte(impactID))
	return []string{"reverify-" + hex.EncodeToString(digest[:])}, nil
}

func repairTargetedSelections(root string, state map[string]any) ([]repairTargetedSelection, error) {
	pointer := repairPointerCLI(state)
	planRef, err := repairPointerRef(pointer, "plan_ref", "plan_sha256", "current RepairPlan")
	if err != nil {
		return nil, err
	}
	plan, err := repair.ValidateRepairPlan(root, planRef)
	if err != nil {
		return nil, fmt.Errorf("read current RepairPlan: %w", err)
	}
	impactRef, err := repairPointerRef(pointer, "impact_ref", "impact_sha256", "current ChangeImpact")
	if err != nil {
		return nil, err
	}
	impact, err := repair.ValidateChangeImpact(root, impactRef)
	if err != nil {
		return nil, fmt.Errorf("read current ChangeImpact: %w", err)
	}
	owners := stringMapCLI(pointer["assignment_owners"])
	selections := []repairTargetedSelection{}
	obligationIDs := impact.RequiredReverificationIDs
	if len(obligationIDs) != 1 {
		return nil, fmt.Errorf("current ChangeImpact has %d required reverification IDs; the original S9 contract flow maps one Contract obligation to one dispatched verifier", len(obligationIDs))
	}
	for _, assignment := range plan.Assignments {
		originalOwner := owners[assignment.AssignmentID]
		performingID := repair.VerificationAssignmentID(assignment.AssignmentID)
		verifierOwner := owners[performingID]
		if verifierOwner == "" {
			continue
		}
		selections = append(selections, repairTargetedSelection{
			originalAssignmentID:   assignment.AssignmentID,
			performingAssignmentID: performingID,
			repairOwner:            originalOwner,
			verifierOwner:          verifierOwner,
			impact:                 impact,
		})
	}
	return selections, nil
}

func repairTargetedRequestMappingError(selections []repairTargetedSelection, missing []string) error {
	if len(selections) == 0 {
		return fmt.Errorf("current Runtime has no independently dispatched verifier assignment for the Contract")
	}
	if len(selections) != 1 {
		return fmt.Errorf("one Contract obligation maps to %d independently dispatched verifier assignments; preserve the original dispatch source and resolve the ambiguity", len(selections))
	}
	if len(missing) == 0 {
		return fmt.Errorf("Runtime has no outstanding required reverification ID")
	}
	if len(missing) != 1 {
		return fmt.Errorf("current ChangeImpact has %d outstanding IDs but the original Contract flow supports one TargetedReverification; preserve the ChangeImpact source for reconciliation", len(missing))
	}
	selection := selections[0]
	if selection.repairOwner == "" || selection.verifierOwner == "" {
		return fmt.Errorf("required ID %s maps to RepairPlan assignment %s, but its original owner or independent verifier is not registered in Runtime assignment_owners", missing[0], selection.originalAssignmentID)
	}
	if selection.verifierOwner == selection.repairOwner {
		return fmt.Errorf("required ID %s maps to assignment %s whose verifier owner is the repair owner; preserve the original independent-dispatch source", missing[0], selection.originalAssignmentID)
	}
	return nil
}

func repairTargetedOutstandingSelections(selections []repairTargetedSelection, missing []string) []repairTargetedSelection {
	if len(selections) != 1 || len(missing) != 1 {
		return nil
	}
	return selections
}

func repairTargetedIDForSelection(selection repairTargetedSelection, selections []repairTargetedSelection, missing []string) (string, error) {
	if len(missing) == 1 && len(selections) == 1 && selections[0].originalAssignmentID == selection.originalAssignmentID {
		return missing[0], nil
	}
	return "", fmt.Errorf("assignment %s has no unique outstanding ChangeImpact obligation", selection.originalAssignmentID)
}

func prepareRepairTargetedQuestionnaire(root string, state map[string]any, selection repairTargetedSelection) (string, repair.TargetedReverificationRequest, error) {
	path, err := repairTargetedQuestionnairePath(root, state, selection.performingAssignmentID)
	if err != nil {
		return "", repair.TargetedReverificationRequest{}, err
	}
	selections, err := repairTargetedSelections(root, state)
	if err != nil {
		return "", repair.TargetedReverificationRequest{}, err
	}
	missing, err := repairMissingTargetedIDs(root, repairPointerCLI(state))
	if err != nil {
		return "", repair.TargetedReverificationRequest{}, err
	}
	if err := repairTargetedRequestMappingError(selections, missing); err != nil {
		return "", repair.TargetedReverificationRequest{}, err
	}
	planRef, err := repairPointerRef(repairPointerCLI(state), "plan_ref", "plan_sha256", "current RepairPlan")
	if err != nil {
		return "", repair.TargetedReverificationRequest{}, err
	}
	if _, err := repair.ValidateRepairPlan(root, planRef); err != nil {
		return "", repair.TargetedReverificationRequest{}, err
	}
	assertions, err := repairContractAssertionResults(root, repairPointerCLI(state))
	if err != nil {
		return "", repair.TargetedReverificationRequest{}, err
	}
	if len(assertions) == 0 {
		return "", repair.TargetedReverificationRequest{}, fmt.Errorf("current approved Contract has no assertion surface")
	}
	reverificationID, err := repairTargetedIDForSelection(selection, selections, missing)
	if err != nil {
		return "", repair.TargetedReverificationRequest{}, err
	}
	request := repair.TargetedReverificationRequest{
		ReverificationID: reverificationID, RuntimeID: stringValue(state["runtime_id"]),
		CaseID:               stringValue(repairPointerCLI(state)["case_id"]),
		BaselineGeneration:   integerValue(mapFieldCLI(state, "baseline")["generation"]),
		OriginalAssignmentID: selection.originalAssignmentID, PerformingAssignmentID: selection.performingAssignmentID,
		ImpactID: selection.impact.ImpactID, AssertionResults: assertions,
	}
	if err := writeRepairLeadDraft(path, request); err != nil {
		return "", repair.TargetedReverificationRequest{}, err
	}
	return path, request, nil
}

func repairContractAssertionResults(root string, pointer map[string]any) ([]repair.AssertionResult, error) {
	contractRef, err := repairPointerRef(pointer, "contract_ref", "contract_sha256", "current approved Contract")
	if err != nil {
		return nil, err
	}
	contract, err := repair.ValidateApprovedContractRef(root, repair.ContractRef{Path: contractRef.Path, SHA256: contractRef.SHA256})
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(contract.Ref.Path)))
	if err != nil {
		return nil, err
	}
	var assertions struct {
		Symptom []json.RawMessage `json:"symptom_assertions"`
		Root    []json.RawMessage `json:"root_invariant_assertions"`
		Gaps    []json.RawMessage `json:"detection_gap_assertions"`
	}
	if err := json.Unmarshal(data, &assertions); err != nil {
		return nil, err
	}
	var results []repair.AssertionResult
	for index := range assertions.Symptom {
		results = append(results, repair.AssertionResult{AssertionID: fmt.Sprintf("symptom-%d", index+1)})
	}
	for index := range assertions.Root {
		results = append(results, repair.AssertionResult{AssertionID: fmt.Sprintf("root-%d", index+1)})
	}
	for index := range assertions.Gaps {
		results = append(results, repair.AssertionResult{AssertionID: fmt.Sprintf("gap-%d", index+1)})
	}
	return results, nil
}

func repairHandoffRequestFromRuntime(root string, state map[string]any) (repair.HandoffRequest, error) {
	pointer := repairPointerCLI(state)
	ref := func(pathKey, shaKey, label string) (repair.ArtifactRef, error) {
		return repairPointerRef(pointer, pathKey, shaKey, label)
	}
	session, err := ref("path", "sha256", "current RepairSession")
	if err != nil {
		return repair.HandoffRequest{}, err
	}
	plan, err := ref("plan_ref", "plan_sha256", "current RepairPlan")
	if err != nil {
		return repair.HandoffRequest{}, err
	}
	contract, err := ref("contract_ref", "contract_sha256", "current RepairContract")
	if err != nil {
		return repair.HandoffRequest{}, err
	}
	result, err := ref("result_ref", "result_sha256", "current RepairResult")
	if err != nil {
		return repair.HandoffRequest{}, err
	}
	impact, err := ref("impact_ref", "impact_sha256", "current ChangeImpact")
	if err != nil {
		return repair.HandoffRequest{}, err
	}
	results := repairArtifactRefs(pointer, "targeted_reverification_artifacts", "", "")
	if len(results) == 0 {
		return repair.HandoffRequest{}, fmt.Errorf("current Runtime has no committed TargetedReverification artifact")
	}
	sessionDocument, err := repair.ValidateRepairSession(root, session)
	if err != nil {
		return repair.HandoffRequest{}, fmt.Errorf("read current RepairSession: %w", err)
	}
	changeset, err := repair.ComputeSessionChangesetRecord(root, sessionDocument)
	if err != nil {
		return repair.HandoffRequest{}, fmt.Errorf("derive current Session Changeset: %w", err)
	}
	changesetRef, err := persistRepairChangeset(root, changeset)
	if err != nil {
		return repair.HandoffRequest{}, fmt.Errorf("persist original Session Changeset: %w", err)
	}
	sessionID := stringValue(pointer["session_id"])
	request := repair.HandoffRequest{
		HandoffID: "repair-handoff-" + dispatchSlug(sessionID),
		Session:   session, Plan: plan, Contract: repair.ContractRef{Path: contract.Path, SHA256: contract.SHA256},
		Result: result, Changeset: changesetRef, ChangeImpact: impact,
		TargetedReverifications: results, HandedOffBy: "Main",
		NextAction: "start the next complete S7 review round from the current repair baseline",
	}
	return request, nil
}

func persistRepairChangeset(root string, changeset repair.Changeset) (repair.ArtifactRef, error) {
	ref, err := repair.PersistChangeset(root, changeset)
	if err == nil {
		return ref, nil
	}
	rel := filepath.ToSlash(filepath.Join(".claude", "review", "repair", "changesets", changeset.ChangesetID+".json"))
	existingRef, readErr := artifactRef(root, rel, changeset.ChangesetID)
	if readErr != nil {
		return repair.ArtifactRef{}, err
	}
	existing, validateErr := repair.ValidateChangeset(root, existingRef)
	if validateErr != nil || existing.Digest != changeset.Digest || existing.SessionID != changeset.SessionID {
		return repair.ArtifactRef{}, err
	}
	return existingRef, nil
}

func completeRepairHandoff(root string) (repair.ArtifactRef, runtime.Snapshot, error) {
	statePath := filepath.Join(root, ".claude", "loop-state.json")
	journalPath := filepath.Join(root, ".claude", "loop-events.jsonl")
	snapshot, err := runtime.NewStore(statePath, journalPath).Snapshot()
	if err != nil {
		return repair.ArtifactRef{}, runtime.Snapshot{}, err
	}
	pointer := repairPointerCLI(snapshot.State)
	if registered := stringValue(pointer["handoff_ref"]); registered != "" {
		ref, err := repairPointerRef(pointer, "handoff_ref", "handoff_sha256", "registered RepairHandoff")
		if err != nil {
			return repair.ArtifactRef{}, snapshot, err
		}
		if _, err := repair.ValidateRepairHandoff(root, ref); err != nil {
			return repair.ArtifactRef{}, snapshot, fmt.Errorf("read registered RepairHandoff: %w", err)
		}
		return ref, snapshot, nil
	}
	if stringValue(pointer["status"]) != "ready_for_full_review" {
		return repair.ArtifactRef{}, snapshot, fmt.Errorf("current repair pointer status is %q, want ready_for_full_review", stringValue(pointer["status"]))
	}
	missing, err := repairMissingTargetedIDs(root, pointer)
	if err != nil {
		return repair.ArtifactRef{}, snapshot, err
	}
	if len(missing) > 0 {
		return repair.ArtifactRef{}, snapshot, fmt.Errorf("Runtime still has outstanding TargetedReverification IDs: %s", strings.Join(missing, ", "))
	}
	request, err := repairHandoffRequestFromRuntime(root, snapshot.State)
	if err != nil {
		return repair.ArtifactRef{}, snapshot, err
	}
	handoff, ref, err := repair.CreateRepairHandoff(root, request)
	if err != nil {
		if !errors.Is(err, os.ErrExist) {
			return repair.ArtifactRef{}, snapshot, err
		}
		rel := filepath.ToSlash(filepath.Join(".claude", "review", "repair", "handoffs", request.HandoffID+".json"))
		ref, err = artifactRef(root, rel, request.HandoffID)
		if err != nil {
			return repair.ArtifactRef{}, snapshot, err
		}
		handoff, err = repair.ValidateRepairHandoff(root, ref)
		if err != nil {
			return repair.ArtifactRef{}, snapshot, fmt.Errorf("read previously created RepairHandoff: %w", err)
		}
	}
	if handoff.HandoffID != request.HandoffID {
		return repair.ArtifactRef{}, snapshot, fmt.Errorf("existing RepairHandoff id %q does not match current S9 session %q", handoff.HandoffID, request.HandoffID)
	}
	committed, err := repair.CommitRepairHandoff(root, statePath, journalPath, repair.CommitHandoffRequest{
		RuntimeRequest: repair.RuntimeRequest{ExpectedRevision: -1, Actor: "Main"}, Handoff: ref,
	})
	if err != nil {
		return repair.ArtifactRef{}, snapshot, fmt.Errorf("commit RepairHandoff %s through the original Runtime consumer: %w", ref.Path, err)
	}
	return ref, committed, nil
}

func repairPointerCLI(state map[string]any) map[string]any {
	return mapFieldCLI(mapFieldCLI(state, "review"), "repair")
}

func repairPointerRef(pointer map[string]any, pathKey, hashKey, label string) (repair.ArtifactRef, error) {
	ref := repair.ArtifactRef{Path: stringValue(pointer[pathKey]), SHA256: stringValue(pointer[hashKey])}
	if ref.Path == "" || ref.SHA256 == "" {
		return repair.ArtifactRef{}, fmt.Errorf("%s is missing from current Runtime", label)
	}
	return ref, nil
}

func questionnaireMissing(path string) bool {
	_, err := os.Stat(path)
	return os.IsNotExist(err)
}

func writeRepairLeadDraft(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return prepareQuestionnaireText(path, string(data)+"\n")
}

func persistRepairLeadQuestionnaire(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

// runSavedRepairLeadQuestionnaire sends an exact saved DTO through its
// original S9 Create/Commit and Runtime Writer functions.
func runSavedRepairLeadQuestionnaire(root string, request policy.Input, stdout, stderr io.Writer) (bool, int) {
	path, saved := savedQuestionnairePath(request)
	if !saved {
		return false, 0
	}
	snapshot, err := runtime.NewStore(filepath.Join(root, ".claude", "loop-state.json"), filepath.Join(root, ".claude", "loop-events.jsonl")).Snapshot()
	if err != nil {
		return false, 0
	}
	pointer := repairPointerCLI(snapshot.State)
	if expected, pathErr := repairLeadQuestionnairePath(root, snapshot.State, "change-impact.json"); pathErr == nil && samePath(path, expected) {
		if stringValue(pointer["impact_ref"]) != "" {
			return true, repairLeadFeedback(stdout, path, "S9 ChangeImpact at "+stringValue(pointer["impact_ref"])+" is already registered. Replay did not write Runtime or register it again; wait for the independent verifier.")
		}
		if stringValue(pointer["status"]) == "impact_reconciliation" {
			return saveRepairImpactQuestionnaire(root, snapshot, path, stdout)
		}
	}
	if selections, selectErr := repairTargetedSelections(root, snapshot.State); selectErr == nil {
		for _, selection := range selections {
			expected, pathErr := repairTargetedQuestionnairePath(root, snapshot.State, selection.performingAssignmentID)
			if pathErr != nil || !samePath(path, expected) {
				continue
			}
			if registered, ok := registeredTargetedForSelection(root, pointer, selection); ok {
				return true, repairLeadFeedback(stdout, path, "S9 TargetedReverification artifact "+registered.Path+" is already registered. Replay did not write Runtime again.")
			}
			phase, phaseErr := currentRepairLeadPhase(root, snapshot.State)
			missing, missingErr := repairMissingTargetedIDs(root, pointer)
			if phaseErr != nil || phase != "targeted_reverification" || missingErr != nil || repairTargetedRequestMappingError(selections, missing) != nil {
				return false, 0
			}
			return saveRepairTargetedQuestionnaire(root, snapshot, path, selection, stdout)
		}
	}
	return false, 0
}

func registeredTargetedForSelection(root string, pointer map[string]any, selection repairTargetedSelection) (repair.ArtifactRef, bool) {
	for _, ref := range repairArtifactRefs(pointer, "targeted_reverification_artifacts", "", "") {
		value, err := repair.ValidateTargetedReverification(root, ref)
		if err == nil && value.Result == "pass" && value.ScopeCompliance == "pass" && value.ImpactID == selection.impact.ImpactID && value.OriginalAssignmentID == selection.originalAssignmentID && value.PerformingAssignmentID == selection.performingAssignmentID {
			return ref, true
		}
	}
	return repair.ArtifactRef{}, false
}

func saveRepairImpactQuestionnaire(root string, snapshot runtime.Snapshot, path string, stdout io.Writer) (bool, int) {
	data, err := os.ReadFile(path)
	if err != nil {
		return true, repairLeadFeedback(stdout, path, fmt.Sprintf("S9 ChangeImpact answer at %s could not be read: %v", path, err))
	}
	var answer repair.ChangeImpactRequest
	if err := json.Unmarshal(data, &answer); err != nil {
		return true, repairLeadFeedback(stdout, path, fmt.Sprintf("S9 ChangeImpact answer at %s is not a valid original request DTO: %v", path, err))
	}
	if len(answer.ChangeTypes) == 0 || len(answer.Decisions) == 0 {
		return true, repairLeadFeedback(stdout, path, "S9 ChangeImpact answer at "+path+" is incomplete: record actual change types and a professional disposition/rationale for each changed surface before saving again.")
	}
	pointer := repairPointerCLI(snapshot.State)
	sessionRef, err := repairPointerRef(pointer, "path", "sha256", "current RepairSession")
	if err != nil {
		return true, repairLeadFeedback(stdout, path, err.Error())
	}
	session, err := repair.ValidateRepairSession(root, sessionRef)
	if err != nil {
		return true, repairLeadFeedback(stdout, path, "S9 ChangeImpact answer could not read the current Session: "+err.Error())
	}
	changed, err := repair.ComputeSessionChangeset(root, session)
	if err != nil {
		return true, repairLeadFeedback(stdout, path, "S9 ChangeImpact answer could not derive the current Session diff: "+err.Error())
	}
	reverificationIDs, err := repairImpactReverificationIDs(snapshot.State)
	if err != nil {
		return true, repairLeadFeedback(stdout, path, "S9 ChangeImpact answer could not derive the Contract reverification obligation from the current RepairSession: "+err.Error())
	}
	boundReq := mapFieldCLI(snapshot.State, "bound_req")
	answer.ImpactID = "impact-" + dispatchSlug(session.SessionID)
	answer.RuntimeID = stringValue(snapshot.State["runtime_id"])
	answer.ReqID = stringValue(boundReq["id"])
	answer.BaselineGeneration = integerValue(mapFieldCLI(snapshot.State, "baseline")["generation"])
	answer.SourceBugIDs = []string{}
	answer.SourceCaseIDs = []string{stringValue(pointer["case_id"])}
	answer.ChangedArtifacts = changed
	if len(answer.RequiredReverificationIDs) == 0 {
		answer.RequiredReverificationIDs = reverificationIDs
	}
	answer.AnalyzedBy = "Main"
	if answer.AnalyzedAt.IsZero() {
		answer.AnalyzedAt = time.Now().UTC()
	}
	if err := persistRepairLeadQuestionnaire(path, answer); err != nil {
		return true, repairLeadFeedback(stdout, path, fmt.Sprintf("S9 ChangeImpact answer at %s could not persist its original DTO: %v", path, err))
	}
	_, ref, err := repair.CreateChangeImpact(root, answer)
	if err != nil {
		return true, repairLeadFeedback(stdout, path, fmt.Sprintf("S9 ChangeImpact request at %s did not create the original artifact: %v", path, err))
	}
	_, err = repair.CommitChangeImpact(root, filepath.Join(root, ".claude", "loop-state.json"), filepath.Join(root, ".claude", "loop-events.jsonl"), repair.CommitImpactRequest{RuntimeRequest: repair.RuntimeRequest{ExpectedRevision: -1, Actor: "Main"}, Impact: ref})
	if err != nil {
		return true, repairLeadFeedback(stdout, path, fmt.Sprintf("S9 ChangeImpact artifact %s was created but the original Runtime consumer did not commit it: %v. Preserve this immutable artifact and follow the named repair source.", ref.Path, err))
	}
	return true, repairLeadFeedback(stdout, path, fmt.Sprintf("S9 ChangeImpact at %s was committed through the original Runtime Writer. The independent verifier owns its separate TargetedReverification answer.", ref.Path))
}

func saveRepairTargetedQuestionnaire(root string, snapshot runtime.Snapshot, path string, selection repairTargetedSelection, stdout io.Writer) (bool, int) {
	data, err := os.ReadFile(path)
	if err != nil {
		return true, repairLeadFeedback(stdout, path, fmt.Sprintf("S9 TargetedReverification answer at %s could not be read: %v", path, err))
	}
	var answer repair.TargetedReverificationRequest
	if err := json.Unmarshal(data, &answer); err != nil {
		return true, repairLeadFeedback(stdout, path, fmt.Sprintf("S9 TargetedReverification answer at %s is not a valid original request DTO: %v", path, err))
	}
	pointer := repairPointerCLI(snapshot.State)
	blockedResume := stringValue(pointer["status"]) == "blocked" && stringValue(pointer["failure_route"]) == "blocked"
	if blockedResume && answer.Result == "blocked" {
		if registered, ok := registeredBlockedTargetedForSelection(root, pointer, selection); ok {
			return true, repairLeadFeedback(stdout, path, "S9 environmental blocker is already recorded by "+registered.Path+". Resolve it, then update the same TargetedReverification questionnaire with the actual continuity basis and replay result; Runtime was not changed.")
		}
	}
	if strings.TrimSpace(answer.ContinuityReason) == "" || strings.TrimSpace(answer.ScopeCompliance) == "" || strings.TrimSpace(answer.Result) == "" {
		return true, repairLeadFeedback(stdout, path, "S9 TargetedReverification answer at "+path+" must record the actual evidence-anchored continuity reason, scope result, and independent verdict before save.")
	}
	for _, assertion := range answer.AssertionResults {
		if strings.TrimSpace(assertion.Result) == "" || len(assertion.EvidenceRefs) == 0 {
			return true, repairLeadFeedback(stdout, path, fmt.Sprintf("S9 TargetedReverification answer at %s is missing the actual result/evidence for assertion %s.", path, assertion.AssertionID))
		}
	}
	answer.RuntimeID = stringValue(snapshot.State["runtime_id"])
	answer.CaseID = stringValue(pointer["case_id"])
	answer.BugID = ""
	answer.BaselineGeneration = integerValue(mapFieldCLI(snapshot.State, "baseline")["generation"])
	answer.OriginalAssignmentID = selection.originalAssignmentID
	answer.PerformingAssignmentID = selection.performingAssignmentID
	answer.ImpactID = selection.impact.ImpactID
	selections, selectionErr := repairTargetedSelections(root, snapshot.State)
	if selectionErr != nil {
		return true, repairLeadFeedback(stdout, path, "S9 TargetedReverification source could not be refreshed: "+selectionErr.Error())
	}
	missing, missingErr := repairMissingTargetedIDs(root, repairPointerCLI(snapshot.State))
	if missingErr != nil {
		return true, repairLeadFeedback(stdout, path, "S9 TargetedReverification obligation could not be refreshed: "+missingErr.Error())
	}
	if mappingErr := repairTargetedRequestMappingError(selections, missing); mappingErr != nil {
		return true, repairLeadFeedback(stdout, path, fmt.Sprintf("S9 TargetedReverification at %s was not submitted: %v. Main must reconcile the original Impact/dispatch source; this verifier must not guess which obligation to claim.", path, mappingErr))
	}
	answer.ReverificationID, err = repairTargetedIDForSelection(selection, selections, missing)
	if err != nil {
		return true, repairLeadFeedback(stdout, path, fmt.Sprintf("S9 TargetedReverification at %s was not submitted: %v. Preserve the original ChangeImpact/Plan source.", path, err))
	}
	if blockedResume {
		if blockedRef, ok := registeredBlockedTargetedForSelection(root, pointer, selection); ok {
			blockedValue, validateErr := repair.ValidateTargetedReverification(root, blockedRef)
			if validateErr == nil {
				blockedAt, parseErr := time.Parse(time.RFC3339Nano, blockedValue.PerformedAt)
				if parseErr == nil && (answer.PerformedAt.IsZero() || answer.PerformedAt.Equal(blockedAt)) {
					answer.PerformedAt = time.Now().UTC()
				}
			}
		}
	}
	if answer.PerformedAt.IsZero() {
		answer.PerformedAt = time.Now().UTC()
	}
	if err := persistRepairLeadQuestionnaire(path, answer); err != nil {
		return true, repairLeadFeedback(stdout, path, fmt.Sprintf("S9 TargetedReverification answer at %s could not persist its original DTO: %v", path, err))
	}
	statePath := filepath.Join(root, ".claude", "loop-state.json")
	journalPath := filepath.Join(root, ".claude", "loop-events.jsonl")
	currentSnapshot := snapshot
	currentPointer := repairPointerCLI(currentSnapshot.State)
	if blockedResume {
		resumed, resumeErr := repair.ResumeTargetedReverification(root, statePath, journalPath, repair.ResumeTargetedRequest{
			RuntimeRequest: repair.RuntimeRequest{ExpectedRevision: currentSnapshot.Revision, Actor: selection.verifierOwner, OccurredAt: answer.PerformedAt},
			Reason:         answer.ContinuityReason,
		})
		if resumeErr != nil {
			return true, repairLeadFeedback(stdout, path, fmt.Sprintf("S9 recorded environmental blocker at %s could not be resumed through the original Runtime consumer: %v", path, resumeErr))
		}
		currentSnapshot = resumed
		currentPointer = repairPointerCLI(currentSnapshot.State)
	}
	var value repair.TargetedReverification
	var ref repair.ArtifactRef
	var createErr error
	if resolvedAtText := stringValue(currentPointer["blocker_resolved_at"]); resolvedAtText != "" {
		resolvedAt, parseErr := time.Parse(time.RFC3339Nano, resolvedAtText)
		if parseErr != nil {
			return true, repairLeadFeedback(stdout, path, "S9 Runtime blocker_resolved_at is invalid; preserve the original Runtime source for reconciliation: "+parseErr.Error())
		}
		value, ref, createErr = repair.CreateTargetedReverificationForResume(root, answer, resolvedAt)
	} else {
		value, ref, createErr = repair.CreateTargetedReverification(root, answer)
	}
	if createErr != nil {
		return true, repairLeadFeedback(stdout, path, fmt.Sprintf("S9 TargetedReverification request at %s did not create the original artifact: %v", path, createErr))
	}
	_, err = repair.CommitTargetedReverification(root, statePath, journalPath, repair.CommitTargetedRequest{RuntimeRequest: repair.RuntimeRequest{ExpectedRevision: currentSnapshot.Revision, Actor: selection.verifierOwner}, Reverification: ref})
	if err != nil {
		return true, repairLeadFeedback(stdout, path, fmt.Sprintf("S9 independent TargetedReverification %s at %s was created but original Runtime consumer did not commit it: %v. Preserve its immutable source and follow the specific consumer feedback.", value.ReverificationID, ref.Path, err))
	}
	return true, repairLeadFeedback(stdout, path, fmt.Sprintf("S9 independent TargetedReverification %s at %s was committed through the original Runtime Writer.", value.ReverificationID, ref.Path))
}

func registeredBlockedTargetedForSelection(root string, pointer map[string]any, selection repairTargetedSelection) (repair.ArtifactRef, bool) {
	var latest repair.ArtifactRef
	for _, ref := range repairArtifactRefs(pointer, "targeted_reverification_artifacts", "", "") {
		value, err := repair.ValidateTargetedReverification(root, ref)
		if err == nil && value.Result == "blocked" && value.FailureClass == "blocked" && value.ImpactID == selection.impact.ImpactID && value.OriginalAssignmentID == selection.originalAssignmentID && value.PerformingAssignmentID == selection.performingAssignmentID {
			latest = ref
		}
	}
	return latest, latest.Path != ""
}

func repairLeadFeedback(stdout io.Writer, path, message string) int {
	message = "Questionnaire: " + path + ". " + message
	_ = json.NewEncoder(stdout).Encode(map[string]any{"hookSpecificOutput": map[string]any{"hookEventName": "PostToolUse", "additionalContext": message}})
	return 0
}
