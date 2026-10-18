package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/entroforge/go-system-builder/internal/investigation"
	"github.com/entroforge/go-system-builder/internal/policy"
	"github.com/entroforge/go-system-builder/internal/repair"
	"github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/schema"
	"github.com/entroforge/go-system-builder/internal/workpacket"
)

// Only professional fields from the original RepairContract schema. The Case
// supplies its existing cause and binding; this is not another causal report.
type investigationLeadContractAnswer struct {
	ViolatedInvariant        string           `json:"violated_invariant"`
	ArchitectureIntent       string           `json:"architecture_intent"`
	RepairUnits              []map[string]any `json:"repair_units"`
	ProspectiveScope         []string         `json:"prospective_scope"`
	ForbiddenScope           []string         `json:"forbidden_scope"`
	CompatibilityMigration   string           `json:"compatibility_migration"`
	SymptomAssertions        []string         `json:"symptom_assertions"`
	RootInvariantAssertions  []string         `json:"root_invariant_assertions"`
	DetectionGapAssertions   []string         `json:"detection_gap_assertions"`
	RegressionSurfaces       []string         `json:"regression_surfaces"`
	StopEscalationConditions []string         `json:"stop_escalation_conditions"`
}

func investigationLeadContractPaths(root string, state map[string]any) []string {
	life, pointer := mapFieldCLI(state, "lifecycle"), investigationPointer(state)
	if life["state"] != "bug_resolution" || life["phase"] != "investigation" || pointer == nil || pointer["status"] == "contract_approved" {
		return nil
	}
	_, document, _, err := investigationLeadSource(root, state)
	if err != nil || document["route"] != "s9_repair" {
		return nil
	}
	route, err := investigationLeadRoutePath(root, state, document)
	if err != nil {
		return nil
	}
	return []string{filepath.Join(filepath.Dir(route), "contract-"+stringValue(pointer["sha256"])+".json")}
}

func stageInvestigationLeadContractContext(root string, request policy.Input, snapshot runtime.Snapshot) string {
	if !mainQuestionnaireViewer(request) {
		return ""
	}
	paths := investigationLeadContractPaths(root, snapshot.State)
	if len(paths) == 0 {
		return ""
	}
	path := paths[0]
	_, document, _, err := investigationLeadSource(root, snapshot.State)
	if err != nil {
		return "S8 Main Contract questionnaire: " + path + ". Original Case recovery is required: " + err.Error()
	}
	if repairLeadPreparationEvent(request.Event) {
		draft := investigationLeadContractAnswer{RepairUnits: []map[string]any{{"id": "repair-unit-1", "description": "", "scope": []string{}, "assertion_ids": []string{}}}, ProspectiveScope: []string{}, ForbiddenScope: []string{}, SymptomAssertions: []string{}, RootInvariantAssertions: []string{}, DetectionGapAssertions: []string{}, RegressionSurfaces: []string{}, StopEscalationConditions: []string{}}
		if err := writeRepairLeadDraft(path, draft); err != nil {
			return fmt.Sprintf("S8 Main Contract questionnaire: %s. Main/system preparation failed: %v", path, err)
		}
	}
	if questionnaireMissing(path) {
		return "S8 Main Contract questionnaire: " + path + ". Not prepared yet; Main's next SessionStart or PreToolUse prepares it. This reminder does not approve or change the Case."
	}
	finalPath, finalBytes, err := investigationLeadContractProjection(root, path, document)
	if err != nil {
		return "S8 Main Contract questionnaire: " + path + ". Fill violated_invariant, architecture_intent, complete repair_units (id/description/scope/assertion_ids), prospective_scope, forbidden_scope, compatibility_migration, symptom_assertions, root_invariant_assertions, detection_gap_assertions, regression_surfaces and stop_escalation_conditions, then save. The existing Case cause is carried automatically; do not repeat its causal report or supply approval metadata. Current original-schema gap: " + err.Error()
	}
	stored, readErr := os.ReadFile(finalPath)
	if readErr != nil || !bytes.Equal(stored, finalBytes) {
		return "S8 Main Contract questionnaire: " + path + ". Fill violated_invariant/architecture_intent as strings; repair_units as objects with id/description and scope arrays such as [\"internal/module\"], plus assertion_ids=[\"all\"] or original symptom-1/root-1/gap-1 assertion IDs. Fill prospective_scope/forbidden_scope, symptom_assertions/root_invariant_assertions/detection_gap_assertions, regression_surfaces and stop_escalation_conditions as string arrays, and compatibility_migration as a string. The Case cause/binding is supplied automatically. Save once to fix final review draft bytes at " + finalPath + ". No approval is implied."
	}
	if request.Event == "PreToolUse" {
		if approval, ok := investigationLeadContractApproval(root, snapshot.State, document, finalPath, finalBytes); ok {
			next, err := investigation.ApproveContract(root, filepath.Join(root, ".claude/loop-state.json"), filepath.Join(root, ".claude/loop-events.jsonl"), approval)
			if err != nil {
				return "S8 Main Contract questionnaire: " + path + ". Final review draft: " + finalPath + ". Original approval validation remains outstanding: " + err.Error()
			}
			prepared, prepareErr := prepareRepairEntry(root, filepath.Join(root, ".claude/loop-state.json"), filepath.Join(root, ".claude/loop-events.jsonl"), next.State)
			if prepareErr != nil {
				return "S8 Original approval recorded. S9 entry preparation: " + prepareErr.Error()
			}
			return fmt.Sprintf("S8 Main Contract questionnaire: %s. Original approval recorded Contract %s; S9 Session/Plan and Main division drafts are prepared. %s", path, stringValue(investigationPointer(next.State)["repair_contract_ref"]), repairLeadExistingQuestionnaires(root, prepared))
		}
	}
	return "S8 Main Contract questionnaire: " + path + ". Final review draft: " + finalPath + "; SHA256=" + workpacket.Digest(finalBytes) + ". Wait for the actual human decision or an existing bounded delegation plus technical review matching these exact bytes. Main must not self-author a human decision. The next Main tool listener passes matching original evidence to the original approval transaction; saving alone does not authorize S9."
}

// Content-addressed original-schema drafts keep previously reviewed bytes
// immutable. A professional edit yields another draft, never rewrites a review.
func investigationLeadContractProjection(root, path string, document map[string]any) (string, []byte, error) {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return "", nil, err
	}
	data, err := workpacket.ReadFile(root, filepath.ToSlash(relative), 20<<20)
	if err != nil {
		return "", nil, err
	}
	return investigationLeadContractProjectionBytes(root, data, document)
}

// Project already-read answer bytes without schema/scope validation or I/O.
// Explicit saves and the original approval transaction keep their validators.
func investigationLeadContractProjectionBytes(root string, data []byte, document map[string]any) (string, []byte, error) {
	if err := workpacket.ValidateUniqueJSON(data); err != nil {
		return "", nil, err
	}
	var answer investigationLeadContractAnswer
	if err := investigationLeadDecode(data, &answer); err != nil {
		return "", nil, err
	}
	professional, _ := json.Marshal(answer)
	var draft map[string]any
	_ = json.Unmarshal(professional, &draft)
	caseID := stringValue(document["case_id"])
	draft["schema_version"], draft["status"] = "1.0.0", "draft"
	draft["repair_contract_id"], draft["case_id"] = "repair-contract-"+caseID, caseID
	draft["revision"], draft["source_finding_ids"] = document["revision"], document["source_finding_ids"]
	draft["root_cause_statement"] = document["primary_root_cause"]
	draft["causal_model_ref"] = "case://" + caseID + "/causal-model"
	finalBytes, err := json.MarshalIndent(draft, "", "  ")
	if err != nil {
		return "", nil, err
	}
	finalBytes = append(finalBytes, '\n')
	finalPath := filepath.Join(root, ".claude", "review", "investigation", "contracts", fmt.Sprintf("draft-%s-r%d-%s.json", caseID, integerValue(document["revision"]), workpacket.Digest(finalBytes)))
	return finalPath, finalBytes, nil
}

func runSavedInvestigationLeadContract(root string, request policy.Input, stdout io.Writer) (bool, int) {
	path, saved := savedQuestionnairePath(request)
	if !saved || !strings.HasPrefix(filepath.Base(path), "contract-") {
		return false, 0
	}
	current, err := runtime.NewStore(filepath.Join(root, ".claude/loop-state.json"), filepath.Join(root, ".claude/loop-events.jsonl")).Snapshot()
	if err != nil || investigationPointer(current.State) == nil {
		return false, 0
	}
	feedback := func(message string) (bool, int) { return true, repairLeadFeedback(stdout, path, message) }
	paths := investigationLeadContractPaths(root, current.State)
	if len(paths) == 0 {
		return investigationLeadContractApprovedReplay(root, path, current.State, feedback)
	}
	if !samePath(path, paths[0]) {
		if filepath.Dir(path) != filepath.Dir(paths[0]) {
			return false, 0
		}
		return feedback("S8 Contract source is stale after the Case changed. Main must recover the current original Case responsibility at " + paths[0] + "; the earlier answer and reviewed draft were preserved, not silently rebound.")
	}
	_, document, _, err := investigationLeadSource(root, current.State)
	if err != nil {
		return feedback("Original Case recovery is required: " + err.Error())
	}
	finalPath, data, err := investigationLeadContractProjection(root, path, document)
	if err != nil {
		return feedback("S8 Contract professional answer is incomplete or invalid under the original schema: " + err.Error())
	}
	if err := schema.NewEmbeddedValidator().ValidateBytes("repair-contract.schema.json", data); err != nil {
		return feedback("S8 Contract original schema validation: " + err.Error())
	}
	if err := repair.ValidateContractUnitScopes(data); err != nil {
		return feedback("S8 Contract original unit scope validation: " + err.Error())
	}
	if err := prepareQuestionnaireText(finalPath, string(data)); err != nil {
		return feedback("S8 original-schema review draft could not be fixed: " + err.Error())
	}
	actual, err := os.ReadFile(finalPath)
	if err != nil || !bytes.Equal(actual, data) {
		return feedback("S8 fixed review draft differs from its exact projected bytes; Main must recover that original draft. No approved or reviewed bytes were overwritten.")
	}
	return feedback("S8 professional Contract answer is projected once to final review draft " + finalPath + "; SHA256=" + workpacket.Digest(data) + ". Wait for a genuine matching decision or original bounded authorization/technical review. Save is not approval; repeated save does not write Runtime.")
}

// The selector reads only already registered original artifacts. It neither
// manufactures decisions nor replaces the original approval/delegation checks.
func investigationLeadContractApproval(root string, state, document map[string]any, path string, draft []byte) (investigation.ContractRequest, bool) {
	var contract map[string]any
	_ = json.Unmarshal(draft, &contract)
	for _, row := range objectSliceAny(state["evidence"]) {
		kind := stringValue(row["kind"])
		if row["status"] != "valid" || (kind != "human_decision" && kind != "repair_contract_review") || runtime.HumanDecisionEvidenceConsumed(row) {
			continue
		}
		artifact, err := investigationLeadReadObject(root, stringValue(row["path"]), stringValue(row["sha256"]))
		if err != nil || artifact["runtime_id"] != state["runtime_id"] || artifact["case_id"] != document["case_id"] || artifact["contract_id"] != contract["repair_contract_id"] || artifact["approval_hash"] != workpacket.Digest(draft) {
			continue
		}
		actor, delegation := stringValue(artifact["approved_by"]), ""
		if kind == "human_decision" {
			if artifact["decision"] != "approve_contract" || artifact["decision_id"] != row["id"] || !s8ContractContains(row["scope_refs"], "s8_contract_approval:"+stringValue(state["runtime_id"])) {
				continue
			}
			if actor == "" {
				producers := stringSliceAny(row["produced_by"])
				if len(producers) > 0 {
					actor = producers[0]
				}
			}
		} else {
			actor, delegation = stringValue(artifact["reviewed_by"]), stringValue(artifact["delegation_evidence_id"])
			if artifact["decision"] != "approve_bounded_repair" || delegation == "" || !s8ContractContains(row["scope_refs"], "s8_contract_review:"+stringValue(document["case_id"])) {
				continue
			}
		}
		if actor == "" || !s8ContractContains(row["produced_by"], actor) {
			continue
		}
		return investigation.ContractRequest{ExpectedRevision: tolerantInt(state["revision"]), CaseID: stringValue(document["case_id"]), ContractPath: path, ApprovedBy: actor, ApprovalHash: workpacket.Digest(draft), ApprovalEvidenceID: stringValue(row["id"]), DelegationEvidenceID: delegation}, true
	}
	return investigation.ContractRequest{}, false
}

func s8ContractContains(value any, want string) bool {
	for _, got := range stringSliceAny(value) {
		if got == want {
			return true
		}
	}
	return false
}

func investigationLeadContractHasProfessionalContent(_ string, data []byte) bool {
	if workpacket.ValidateUniqueJSON(data) != nil {
		return true
	}
	var answer investigationLeadContractAnswer
	if investigationLeadDecode(data, &answer) != nil {
		return true
	}
	if strings.TrimSpace(answer.ViolatedInvariant+answer.ArchitectureIntent+answer.CompatibilityMigration) != "" || len(answer.ProspectiveScope)+len(answer.ForbiddenScope)+len(answer.SymptomAssertions)+len(answer.RootInvariantAssertions)+len(answer.DetectionGapAssertions)+len(answer.RegressionSurfaces)+len(answer.StopEscalationConditions) > 0 {
		return true
	}
	for _, unit := range answer.RepairUnits {
		if strings.TrimSpace(stringValue(unit["description"])) != "" || len(stringSliceAny(unit["scope"])) > 0 || len(stringSliceAny(unit["assertion_ids"])) > 0 {
			return true
		}
	}
	return false
}

func investigationLeadContractApprovedReplay(root, path string, state map[string]any, feedback func(string) (bool, int)) (bool, int) {
	pointer := investigationPointer(state)
	if pointer["status"] != "contract_approved" {
		return false, 0
	}
	approved, err := investigationLeadReadObject(root, stringValue(pointer["repair_contract_ref"]), stringValue(pointer["repair_contract_sha256"]))
	if err != nil {
		return feedback("S8 approved Contract recovery is required: " + err.Error())
	}
	// The original immutable pre-approval Case is the source of this answer.
	revision := integerValue(approved["revision"]) - 1
	caseID := stringValue(approved["case_id"])
	caseRef := filepath.ToSlash(filepath.Join(".claude/review/investigation/cases", fmt.Sprintf("%s-r%d.json", caseID, revision)))
	caseBytes, err := workpacket.ReadFile(root, caseRef, 20<<20)
	if err != nil {
		return feedback("S8 original pre-approval Case recovery is required: " + err.Error())
	}
	_, batchSHA := investigationLeadBatchReference(state)
	expected := filepath.Join(root, ".claude/submissions/s8", dispatchSlug(stringValue(state["runtime_id"])), fmt.Sprintf("g%d", integerValue(mapFieldCLI(state, "baseline")["generation"])), fmt.Sprintf("r%d", integerValue(mapFieldCLI(state, "review")["round"])), batchSHA, "contract-"+workpacket.Digest(caseBytes)+".json")
	if !samePath(path, expected) {
		return false, 0
	}
	var document map[string]any
	if json.Unmarshal(caseBytes, &document) != nil {
		return feedback("S8 original pre-approval Case is unreadable; Main must recover it.")
	}
	_, data, err := investigationLeadContractProjection(root, path, document)
	if err != nil || workpacket.Digest(data) != stringValue(approved["approval_hash"]) {
		return feedback("S8 Contract answer conflicts with the already approved final draft; no approved content or authorization was changed.")
	}
	return feedback("S8 original Contract approval is already recorded at " + resolveRootPath(root, stringValue(pointer["repair_contract_ref"])) + ". Repeated save is read-only; continue with the original S9 handoff.")
}
