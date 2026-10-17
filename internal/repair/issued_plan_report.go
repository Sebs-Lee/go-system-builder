package repair

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/entroforge/go-system-builder/internal/agentauth"
	"github.com/entroforge/go-system-builder/internal/controlproof"
	runtimepkg "github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/schema"
	"github.com/entroforge/go-system-builder/internal/semantic"
	"github.com/entroforge/go-system-builder/internal/workpacket"
	"github.com/entroforge/go-system-builder/internal/workspace"
	"github.com/entroforge/go-system-builder/internal/workstate"
)

const repairPlanReportAnswerLimit int64 = 20 << 20

const repairPlanReportReceiptContract = "s9-plan-report-receipt-v1"

type SubmitIssuedRepairPlanReportRequest struct {
	ExpectedRevision int
	PacketID         string
	AnswerPath       string
	OwnerAgentID     string
	SourceRoot       string
	AuthorizeSource  func(root string, state map[string]any) error
	OccurredAt       time.Time
}

// IssuedPlanReportRequest is a concise alias for Hook/service callers.
type IssuedPlanReportRequest = SubmitIssuedRepairPlanReportRequest

type issuedPlanReportCapability struct {
	packet    workpacket.RepairPlanReportPacket
	receipt   runtimepkg.RepairPlanReportReceipt
	answer    json.RawMessage
	answerSHA string
	source    string
	authorize func(root string, state map[string]any) error
}

// ValidateRepairPlanReportAuthoringStage permits a current issued S9
// PlanReport to be saved before the generic PLAN checkpoint, but only for the
// exact plan_checkpoint Builder still in its original reading state. Callers
// must separately validate this packet's issuance and current Assignment.
func ValidateRepairPlanReportAuthoringStage(root string, state map[string]any, owner string) error {
	if err := agentauth.ValidateWorkingAuthorization(root, state, owner); err == nil {
		return nil
	} else {
		entities, _ := state["entities"].(map[string]any)
		rows, _ := entities["agents"].([]any)
		var found map[string]any
		for _, raw := range rows {
			row, _ := raw.(map[string]any)
			if row == nil || row["id"] != owner {
				continue
			}
			if found != nil {
				return err
			}
			found = row
		}
		if found == nil || workstate.AgentActivity(found) != "reading" || found["dispatch_mode"] != "plan_checkpoint" {
			return err
		}
		return nil
	}
}

func validateIssuedPlanReportCapability(root string, state map[string]any, req SubmitPlanReportRequest, report PlanReport, _ bool) error {
	capability := req.issued
	if capability == nil {
		return errors.New("missing private S9 issued PlanReport capability")
	}
	packet, found, err := runtimepkg.FindRepairPlanReportPacket(state, capability.packet.ID)
	if err != nil || !found || !sameRepairPlanReportPacket(packet, capability.packet) {
		return fmt.Errorf("S9 packet is no longer the exact registered Runtime packet: %v", err)
	}
	if req.Actor != capability.packet.Binding.OwnerAgentID || report.AgentID != capability.packet.Binding.OwnerAgentID || report.ReportID != capability.packet.ReportID ||
		report.SessionID != capability.packet.Binding.SessionID || report.PlanID != capability.packet.Binding.PlanID || report.AssignmentID != capability.packet.Binding.RepairAssignmentID {
		return errors.New("S9 PlanReport actor or original Assignment tuple differs from the private packet capability")
	}
	if report.Receipt == nil || !reflect.DeepEqual(*report.Receipt, capability.receipt) || capability.receipt.PacketID != capability.packet.ID ||
		capability.receipt.AnswerSHA256 != capability.answerSHA || capability.receipt.ReportRef != req.Report.Path ||
		capability.receipt.ReportID != report.ReportID || capability.receipt.WorkComplete {
		return errors.New("S9 PlanReport receipt does not match the private accepted-answer capability")
	}
	if err := runtimepkg.VerifyRepairPlanReportPacketIssue(root, state, capability.packet); err != nil {
		return err
	}
	if err := runtimepkg.VerifyRepairPlanReportCurrent(root, state, capability.packet); err != nil {
		return err
	}
	revision, ok := integerForRepairReport(state["revision"])
	if !ok || revision < 0 {
		return errors.New("S9 current Runtime revision is invalid")
	}
	if err := runtimepkg.VerifyRepairPlanReportSources(root, state, capability.packet, revision+1); err != nil {
		return err
	}
	if capability.authorize != nil {
		if err := capability.authorize(root, state); err != nil {
			return err
		}
	}
	registeredRoot, err := repairPlanReportSourceRoot(root, state, capability.packet)
	if err != nil || !sameAbsolutePath(registeredRoot, capability.source) {
		return errors.New("S9 PlanReport authoring root no longer matches registered execution authority")
	}
	answerBytes, err := workpacket.ReadFile(capability.source, capability.packet.AnswerPath, repairPlanReportAnswerLimit)
	if err != nil {
		return err
	}
	answer, problems := workpacket.ValidateRepairPlanReportAnswer(capability.packet, answerBytes)
	if len(problems) > 0 || workpacket.Digest(answer) != capability.answerSHA || !bytesEqualRepairPlanReportAnswer(answer, capability.answer) {
		return runtimepkg.ErrSubmissionConflict
	}
	if err := ValidateRepairPlanReportAuthoringStage(root, state, capability.packet.Binding.OwnerAgentID); err != nil {
		return err
	}
	issueAt, err := repairPlanReportIssueTime(root, state, capability.packet)
	if err != nil {
		return err
	}
	if report.ReportedAt != issueAt.UTC().Format(time.RFC3339Nano) {
		return errors.New("S9 PlanReport reported_at is not derived from the original PACKET-ISSUE event")
	}
	var professional workpacket.RepairPlanReportProfessionalAnswer
	if err := json.Unmarshal(capability.answer, &professional); err != nil {
		return err
	}
	if report.Plan != professional.Plan || !reflect.DeepEqual(report.ProposedPaths, professional.ProposedPaths) {
		return errors.New("S9 PlanReport professional projection differs from the accepted packet answer")
	}
	wantChecks := make([]RepairCheck, len(professional.RedChecks))
	for i, check := range professional.RedChecks {
		wantChecks[i] = RepairCheck{Name: check.Name, Command: check.Command, Result: check.Result, EvidenceRefs: append([]string(nil), check.EvidenceRefs...)}
	}
	if !reflect.DeepEqual(report.RedChecks, wantChecks) {
		return errors.New("S9 PlanReport red-check projection differs from the accepted packet answer")
	}
	return nil
}

func bytesEqualRepairPlanReportAnswer(left, right []byte) bool {
	return string(left) == string(right)
}

// IssueRepairPlanReportPackets is the S9 adapter to the original runtime
// PACKET-ISSUE writer. It only issues packets for registered, current Builder
// Assignments; the packet is an input contract, not authority to write or to
// begin execution.
func IssueRepairPlanReportPackets(root, statePath, journalPath string, expectedRevision int, ownerFilter string) ([]workpacket.RepairPlanReportPacket, error) {
	current, writer, err := readIssuedRepairPlanReportRuntime(root, statePath, journalPath)
	if err != nil {
		return nil, err
	}
	candidates, err := buildRepairPlanReportPackets(root, current.State, ownerFilter)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		if err := checkRevision(expectedRevision, current.Revision, "runtime S9 PlanReport packet issue"); err != nil {
			return nil, err
		}
		return []workpacket.RepairPlanReportPacket{}, nil
	}
	issued := make([]workpacket.RepairPlanReportPacket, 0, len(candidates))
	for _, packet := range candidates {
		if err := runtimepkg.VerifyRepairPlanReportSources(root, current.State, packet, current.Revision+1); err != nil {
			return nil, fmt.Errorf("S9 packet source closure for Assignment %s: %w", packet.Binding.RepairAssignmentID, err)
		}
		if prior, found, err := runtimepkg.FindRepairPlanReportPacket(current.State, packet.ID); err != nil {
			return nil, err
		} else if found {
			if !sameRepairPlanReportPacket(prior, packet) {
				return nil, fmt.Errorf("S9 packet identity %s is already registered with different frozen sources", packet.ID)
			}
			if err := runtimepkg.VerifyRepairPlanReportPacketIssue(root, current.State, prior); err != nil {
				return nil, fmt.Errorf("S9 packet %s has no exact original issuance: %w", packet.ID, err)
			}
			if err := runtimepkg.VerifyRepairPlanReportCurrent(root, current.State, prior); err != nil {
				return nil, err
			}
			issued = append(issued, prior)
			continue
		}
		// An exact prior PACKET-ISSUE is replayed above before enforcing the
		// caller's original CAS. This lets a mutation-capable Snapshot recover
		// its pending write and then return the exact committed packet without
		// treating that recovery as a conflicting fresh issuance.
		if err := checkRevision(expectedRevision, current.Revision, "runtime S9 PlanReport packet issue"); err != nil {
			return nil, err
		}
		if err := issueOneRepairPlanReportPacket(root, writer, expectedRevision, current, packet); err != nil {
			return nil, err
		}
		latest, err := writer.Snapshot()
		if err != nil {
			return nil, err
		}
		stored, found, err := runtimepkg.FindRepairPlanReportPacket(latest.State, packet.ID)
		if err != nil || !found || !sameRepairPlanReportPacket(stored, packet) {
			return nil, fmt.Errorf("S9 packet %s was not retained exactly by the original Runtime writer: %v", packet.ID, err)
		}
		if err := runtimepkg.VerifyRepairPlanReportPacketIssue(root, latest.State, stored); err != nil {
			return nil, err
		}
		if err := runtimepkg.VerifyRepairPlanReportCurrent(root, latest.State, stored); err != nil {
			return nil, err
		}
		issued = append(issued, stored)
		current = latest
		expectedRevision = latest.Revision
	}
	return issued, nil
}

func issueOneRepairPlanReportPacket(root string, writer *runtimepkg.Store, expected int, current runtimepkg.Snapshot, packet workpacket.RepairPlanReportPacket) error {
	rowBytes, err := json.Marshal(packet)
	if err != nil {
		return err
	}
	var packetRow map[string]any
	if err := json.Unmarshal(rowBytes, &packetRow); err != nil {
		return err
	}
	if err := schema.NewEmbeddedValidator().ValidateBytes("repair-plan-report-work-packet.schema.json", rowBytes); err != nil {
		return err
	}
	cursor := cursor(current.State)
	mutation := runtimepkg.Mutation{
		EventID: packet.IssuanceEventID, TransitionID: "PACKET-ISSUE", Event: "work_packet_issued",
		Actor: "orchestrator", IdempotencyKey: "packet:" + packet.ID, RuntimeID: packet.Binding.RuntimeID,
		From: cursor, To: cursor, BaselineGeneration: packet.Binding.BaselineGeneration,
		OccurredAt: time.Now().UTC(), RetainLastTransition: true,
		Apply: func(state map[string]any) error {
			if err := runtimepkg.VerifyRepairPlanReportSources(root, state, packet, current.Revision+1); err != nil {
				return err
			}
			fresh, err := buildRepairPlanReportPackets(root, state, packet.Binding.OwnerAgentID)
			if err != nil || !containsRepairPlanReportPacket(fresh, packet) {
				return fmt.Errorf("S9 dispatch/source tuple changed before packet publication: %v", err)
			}
			if existing, found, err := runtimepkg.FindRepairPlanReportPacket(state, packet.ID); err != nil {
				return err
			} else if found {
				if sameRepairPlanReportPacket(existing, packet) {
					return errors.New("identical S9 packet appeared during issuance; recover from the original PACKET-ISSUE event")
				}
				return fmt.Errorf("S9 packet ID %s conflicts with an existing Runtime packet", packet.ID)
			}
			all, err := runtimepkg.RepairPlanReportPackets(state)
			if err != nil {
				return err
			}
			for _, other := range all {
				b := other.Binding
				if b.RuntimeID == packet.Binding.RuntimeID && b.BaselineGeneration == packet.Binding.BaselineGeneration &&
					b.ReviewRound == packet.Binding.ReviewRound && b.SessionID == packet.Binding.SessionID &&
					b.RepairAssignmentID == packet.Binding.RepairAssignmentID {
					return fmt.Errorf("RepairAssignment %s already has issued S9 packet %s; inspect its receipt or use named recovery", b.RepairAssignmentID, other.ID)
				}
			}
			entities, ok := state["entities"].(map[string]any)
			if !ok {
				return errors.New("S9 packet issue requires Runtime entities")
			}
			rows, ok := entities["work_packets"].([]any)
			if !ok || rows == nil {
				rows = []any{}
			}
			entities["work_packets"] = append(rows, packetRow)
			return nil
		},
	}
	if expected < 0 {
		_, err = writer.UpdateCurrent(mutation)
	} else {
		_, err = writer.Update(expected, mutation)
	}
	if err == nil {
		return nil
	}
	latest, readErr := writer.Snapshot()
	if readErr == nil {
		stored, found, findErr := runtimepkg.FindRepairPlanReportPacket(latest.State, packet.ID)
		if findErr == nil && found && sameRepairPlanReportPacket(stored, packet) &&
			runtimepkg.VerifyRepairPlanReportPacketIssue(root, latest.State, stored) == nil {
			return nil
		}
	}
	return fmt.Errorf("issue S9 PlanReport packet %s: %w", packet.ID, err)
}

func buildRepairPlanReportPackets(root string, state map[string]any, ownerFilter string) ([]workpacket.RepairPlanReportPacket, error) {
	if !runtimepkg.PacketEnabled(state) {
		return nil, errors.New("S9 typed PlanReport packets require the explicit v1.2 packet contract")
	}
	review, _ := state["review"].(map[string]any)
	pointer, _ := review["repair"].(map[string]any)
	casePointer, _ := review["investigation"].(map[string]any)
	lifecycle, _ := state["lifecycle"].(map[string]any)
	if pointer == nil || (pointer["status"] != "planning" && pointer["status"] != "reproducing") || lifecycle["state"] != "bug_resolution" ||
		(lifecycle["phase"] != "planning" && lifecycle["phase"] != "reproducing") {
		return []workpacket.RepairPlanReportPacket{}, nil
	}
	base, _ := state["baseline"].(map[string]any)
	generation, ok := integerForRepairReport(base["generation"])
	if !ok || generation < 0 {
		return nil, errors.New("S9 PlanReport packet requires a valid current baseline generation")
	}
	round, ok := integerForRepairReport(review["round"])
	if !ok || round < 1 {
		return nil, errors.New("S9 PlanReport packet requires a valid current Review round")
	}
	runtimeID := stringField(state["runtime_id"])
	caseID, caseRef, caseSHA := stringField(casePointer["case_id"]), stringField(casePointer["path"]), stringField(casePointer["sha256"])
	contractRef, contractSHA := stringField(pointer["contract_ref"]), stringField(pointer["contract_sha256"])
	sessionID, sessionRef, sessionSHA := stringField(pointer["session_id"]), stringField(pointer["path"]), stringField(pointer["sha256"])
	planRef, planSHA := stringField(pointer["plan_ref"]), stringField(pointer["plan_sha256"])
	if runtimeID == "" || caseID == "" || caseRef == "" || caseSHA == "" || contractRef == "" || contractSHA == "" || sessionID == "" || sessionRef == "" || sessionSHA == "" || planRef == "" || planSHA == "" {
		return nil, errors.New("S9 PlanReport packet requires exact current Case, Contract, Session, and Plan references")
	}
	caseBytes, err := workpacket.ReadFile(root, caseRef, 20<<20)
	if err != nil || workpacket.Digest(caseBytes) != caseSHA {
		return nil, fmt.Errorf("S9 current Case bytes/SHA differ: %v", err)
	}
	contractBytes, err := readArtifact(root, ArtifactRef{Path: contractRef, SHA256: contractSHA}, "repair-contract.schema.json")
	if err != nil {
		return nil, err
	}
	sessionBytes, err := readArtifact(root, ArtifactRef{Path: sessionRef, SHA256: sessionSHA}, "repair-session.schema.json")
	if err != nil {
		return nil, err
	}
	planBytes, err := readArtifact(root, ArtifactRef{Path: planRef, SHA256: planSHA}, "repair-plan.schema.json")
	if err != nil {
		return nil, err
	}
	if err := workpacket.ValidateUniqueJSON(caseBytes); err != nil {
		return nil, fmt.Errorf("S9 Case source: %w", err)
	}
	var caseDoc map[string]any
	if err := json.Unmarshal(caseBytes, &caseDoc); err != nil || caseDoc["case_id"] != caseID || caseDoc["status"] != "contract_approved" {
		return nil, errors.New("S9 current Case identity/status differs from approved Contract pointer")
	}
	contractRefSHA := ArtifactRef{Path: contractRef, SHA256: contractSHA}
	contract, err := ValidateApprovedContractRef(root, ContractRef{Path: contractRefSHA.Path, SHA256: contractRefSHA.SHA256})
	if err != nil {
		return nil, err
	}
	session, err := ValidateRepairSession(root, ArtifactRef{ID: sessionID, Path: sessionRef, SHA256: sessionSHA})
	if err != nil {
		return nil, err
	}
	plan, err := ValidateRepairPlan(root, ArtifactRef{ID: stringField(pointer["plan_id"]), Path: planRef, SHA256: planSHA})
	if err != nil {
		return nil, err
	}
	if contract.CaseID != caseID || contract.ContractID != stringField(pointer["contract_id"]) || session.SessionID != sessionID ||
		session.RuntimeID != runtimeID || session.BaselineGeneration != generation || session.ContractRef != contractRef || session.ContractSHA256 != contractSHA ||
		plan.SessionID != sessionID || plan.ContractRef != contractRef || plan.ContractSHA256 != contractSHA {
		return nil, errors.New("S9 Case/Contract/Session/Plan source chain differs")
	}
	if err := workpacket.ValidateUniqueJSON(planBytes); err != nil {
		return nil, err
	}
	var planDoc map[string]any
	if err := json.Unmarshal(planBytes, &planDoc); err != nil || planDoc == nil {
		return nil, errors.New("S9 RepairPlan source must be an object")
	}
	entities, _ := state["entities"].(map[string]any)
	owners := stringMapField(pointer["assignment_owners"])
	packets := make([]workpacket.RepairPlanReportPacket, 0)
	for _, assignment := range plan.Assignments {
		owner := owners[assignment.AssignmentID]
		if owner == "" || (ownerFilter != "" && owner != ownerFilter) {
			continue
		}
		if assignment.OwnerAgentID != "" && assignment.OwnerAgentID != owner {
			return nil, fmt.Errorf("S9 RepairAssignment %s registered owner differs from its frozen Plan", assignment.AssignmentID)
		}
		agent, count := uniqueRepairEntity(entities["agents"], owner)
		if count != 1 || agent["role"] != "backend-builder" && agent["role"] != "frontend-builder" && agent["role"] != "test-builder" {
			return nil, fmt.Errorf("S9 RepairAssignment %s requires one registered Builder owner", assignment.AssignmentID)
		}
		teamID := stringField(agent["team_id"])
		team, teamCount := uniqueRepairEntity(entities["teams"], teamID)
		if teamID == "" || teamCount != 1 || team["kind"] != "builder" {
			return nil, fmt.Errorf("S9 RepairAssignment %s requires one registered Builder Team", assignment.AssignmentID)
		}
		manifestRef, manifestSHA := stringField(team["manifest_ref"]), stringField(team["manifest_sha256"])
		manifestBytes, err := workpacket.ReadFile(root, manifestRef, 20<<20)
		if err != nil || workpacket.Digest(manifestBytes) != manifestSHA {
			return nil, fmt.Errorf("S9 registered Builder manifest bytes/SHA differ: %v", err)
		}
		if err := workpacket.ValidateUniqueJSON(manifestBytes); err != nil {
			return nil, err
		}
		var manifest map[string]any
		if err := json.Unmarshal(manifestBytes, &manifest); err != nil || manifest == nil {
			return nil, errors.New("S9 Builder manifest source must be an object")
		}
		prompt := stringField(agent["prompt_ref"])
		manifestAssignmentID := ""
		if prefix := manifestRef + "#"; strings.HasPrefix(prompt, prefix) {
			manifestAssignmentID = strings.TrimPrefix(prompt, prefix)
		}
		if manifestAssignmentID == "" {
			return nil, fmt.Errorf("S9 Builder Agent %s has no exact registered manifest Assignment", owner)
		}
		manifestAssignment, manifestAssignmentCount := uniqueRepairEntityByField(manifest["assignments"], "assignment_id", manifestAssignmentID)
		if manifestAssignmentCount != 1 || manifestAssignment["agent_id"] != owner {
			return nil, fmt.Errorf("S9 Builder manifest Assignment %s does not bind owner %s", manifestAssignmentID, owner)
		}
		manifestAssignmentBytes, err := json.Marshal(manifestAssignment)
		if err != nil {
			return nil, err
		}
		var taskRow map[string]any
		var taskBytes []byte
		taskCount := 0
		taskIDs, ok := stringSliceForRepairReport(agent["task_ids"])
		if !ok {
			return nil, fmt.Errorf("S9 Builder Agent %s task_ids is invalid", owner)
		}
		for _, taskID := range taskIDs {
			task, n := uniqueRepairEntity(entities["tasks"], taskID)
			if n != 1 {
				return nil, fmt.Errorf("S9 Builder Agent %s task %s does not resolve to one registered Runtime TASK", owner, taskID)
			}
			taskOwners, ok := stringSliceForRepairReport(task["owner_agent_ids"])
			if !ok || !containsStringForRepairReport(taskOwners, owner) {
				return nil, fmt.Errorf("S9 Runtime TASK %s is not registered to Builder %s", taskID, owner)
			}
			taskRef, taskSHA := stringField(task["path"]), stringField(task["sha256"])
			registeredTaskBytes, readErr := workpacket.ReadFile(root, taskRef, 20<<20)
			if readErr != nil || taskSHA == "" || workpacket.Digest(registeredTaskBytes) != taskSHA {
				return nil, fmt.Errorf("S9 registered TASK %s bytes/SHA differ: %v", taskID, readErr)
			}
			if err := workpacket.ValidateUniqueJSON(registeredTaskBytes); err != nil {
				return nil, fmt.Errorf("S9 registered TASK %s is invalid JSON: %w", taskID, err)
			}
			var taskDocument map[string]any
			if err := json.Unmarshal(registeredTaskBytes, &taskDocument); err != nil || taskDocument == nil {
				return nil, fmt.Errorf("S9 registered TASK %s must be an object: %v", taskID, err)
			}
			if taskDocument["task_id"] == taskID && taskDocument["session_id"] == sessionID && taskDocument["plan_id"] == plan.PlanID && taskDocument["assignment_id"] == assignment.AssignmentID {
				taskRow = task
				taskBytes = registeredTaskBytes
				taskCount++
			}
		}
		if taskCount != 1 {
			return nil, fmt.Errorf("S9 Builder %s must have exactly one registered TASK for RepairAssignment %s", owner, assignment.AssignmentID)
		}
		taskRef, taskSHA := stringField(taskRow["path"]), stringField(taskRow["sha256"])
		assignmentRaw, err := exactPlanAssignmentRaw(planDoc, assignment.AssignmentID)
		if err != nil {
			return nil, err
		}
		inputs := workpacket.RepairPlanReportInputs{
			Case: json.RawMessage(caseBytes), Contract: json.RawMessage(contractBytes), Session: json.RawMessage(sessionBytes), Plan: json.RawMessage(planBytes),
			Task: json.RawMessage(taskBytes), RepairAssignment: assignmentRaw, ManifestAssignment: json.RawMessage(manifestAssignmentBytes),
			ProspectiveScope: append([]string{}, plan.ProspectiveScope...), ForbiddenScope: append([]string{}, plan.ForbiddenScope...),
		}
		binding := workpacket.RepairPlanReportBinding{
			RuntimeID: runtimeID, BaselineGeneration: generation, ReviewRound: round,
			CaseID: caseID, CaseRef: caseRef, CaseSHA256: caseSHA,
			ContractID: contract.ContractID, ContractRef: contractRef, ContractSHA256: contractSHA,
			SessionID: sessionID, SessionRef: sessionRef, SessionSHA256: sessionSHA,
			PlanID: plan.PlanID, PlanRef: planRef, PlanSHA256: planSHA,
			RepairAssignmentID: assignment.AssignmentID, RepairAssignmentSHA256: workpacket.Digest(assignmentRaw),
			ManifestAssignmentID: manifestAssignmentID, ManifestAssignmentSHA256: workpacket.Digest(manifestAssignmentBytes),
			TeamID: teamID, TaskID: stringField(taskRow["id"]), TaskRef: taskRef, TaskSHA256: taskSHA,
			ManifestRef: manifestRef, ManifestSHA256: manifestSHA, OwnerAgentID: owner,
		}
		packet, err := workpacket.BuildRepairPlanReport(binding, inputs)
		if err != nil {
			return nil, fmt.Errorf("build S9 PlanReport packet for %s: %w", assignment.AssignmentID, err)
		}
		if _, err := repairPlanReportSourceRoot(root, state, packet); err != nil {
			return nil, err
		}
		packets = append(packets, packet)
	}
	return packets, nil
}

func exactPlanAssignmentRaw(plan map[string]any, assignmentID string) (json.RawMessage, error) {
	rows, ok := plan["assignments"].([]any)
	if !ok {
		return nil, errors.New("S9 RepairPlan assignments must be an array")
	}
	var found json.RawMessage
	for _, raw := range rows {
		row, ok := raw.(map[string]any)
		if !ok || row["assignment_id"] != assignmentID {
			continue
		}
		if found != nil {
			return nil, fmt.Errorf("S9 RepairPlan duplicates Assignment %s", assignmentID)
		}
		data, err := json.Marshal(row)
		if err != nil {
			return nil, err
		}
		found = data
	}
	if found == nil {
		return nil, fmt.Errorf("S9 RepairPlan lacks Assignment %s", assignmentID)
	}
	return found, nil
}

func uniqueRepairEntity(values any, id string) (map[string]any, int) {
	return uniqueRepairEntityByField(values, "id", id)
}

func uniqueRepairEntityByField(values any, field, id string) (map[string]any, int) {
	var selected map[string]any
	count := 0
	rows, _ := values.([]any)
	for _, raw := range rows {
		row, _ := raw.(map[string]any)
		if row[field] == id {
			selected = row
			count++
		}
	}
	return selected, count
}

func stringSliceForRepairReport(value any) ([]string, bool) {
	switch values := value.(type) {
	case []string:
		return append([]string(nil), values...), true
	case []any:
		result := make([]string, len(values))
		for i, raw := range values {
			text, ok := raw.(string)
			if !ok {
				return nil, false
			}
			result[i] = text
		}
		return result, true
	default:
		return nil, false
	}
}

func containsStringForRepairReport(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

// SubmitIssuedRepairPlanReport accepts one professional S9 answer through the
// original immutable PlanReport writer. The receipt only records planning
// evidence (work_complete=false); BeginRepairExecution remains a separate
// original checkpoint.
func SubmitIssuedRepairPlanReport(root, statePath, journalPath string, req SubmitIssuedRepairPlanReportRequest) (runtimepkg.Snapshot, PlanReport, error) {
	snapshot, writer, err := readIssuedRepairPlanReportRuntime(root, statePath, journalPath)
	if err != nil {
		return runtimepkg.Snapshot{}, PlanReport{}, err
	}
	packet, found, err := runtimepkg.FindRepairPlanReportPacket(snapshot.State, req.PacketID)
	if err != nil || !found {
		return snapshot, PlanReport{}, fmt.Errorf("S9 issued PlanReport packet %q is unavailable: %v", req.PacketID, err)
	}
	if req.OwnerAgentID == "" || req.OwnerAgentID != packet.Binding.OwnerAgentID {
		return snapshot, PlanReport{}, errors.New("S9 PlanReport owner binding differs from registered Builder")
	}
	if req.AnswerPath != packet.AnswerPath {
		return snapshot, PlanReport{}, errors.New("S9 answer path differs from the exact issued packet")
	}
	actualSource, err := repairPlanReportSourceRoot(root, snapshot.State, packet)
	if err != nil {
		return snapshot, PlanReport{}, err
	}
	if req.SourceRoot == "" || !sameAbsolutePath(req.SourceRoot, actualSource) {
		return snapshot, PlanReport{}, errors.New("S9 answer source root differs from the current registered Main/Worker source")
	}
	if req.AuthorizeSource != nil {
		if err := req.AuthorizeSource(root, snapshot.State); err != nil {
			return snapshot, PlanReport{}, err
		}
	}
	answerBytes, err := workpacket.ReadFile(actualSource, packet.AnswerPath, repairPlanReportAnswerLimit)
	if err != nil {
		return snapshot, PlanReport{}, err
	}
	answer, problems := workpacket.ValidateRepairPlanReportAnswer(packet, answerBytes)
	if len(problems) != 0 {
		return snapshot, PlanReport{}, fmt.Errorf("S9 answer %s: %s (%s)", problems[0].Pointer, problems[0].Expected, problems[0].Observed)
	}
	answerSHA := workpacket.Digest(answer)
	if err := runtimepkg.VerifyRepairPlanReportPacketIssue(root, snapshot.State, packet); err != nil {
		return snapshot, PlanReport{}, err
	}
	if accepted, ok, receiptErr := runtimepkg.VerifiedRepairPlanReportReceipt(root, snapshot.State, packet.ID, req.OwnerAgentID); receiptErr != nil {
		return snapshot, PlanReport{}, receiptErr
	} else if ok {
		if accepted.AnswerSHA256 != answerSHA || accepted.ExecutionRoot == "" || !sameAbsolutePath(accepted.ExecutionRoot, actualSource) {
			return snapshot, PlanReport{}, runtimepkg.ErrSubmissionConflict
		}
		report, err := loadAcceptedRepairPlanReport(root, snapshot.State, accepted)
		return snapshot, report, err
	}
	if err := checkRevision(req.ExpectedRevision, snapshot.Revision, "runtime S9 issued PlanReport answer"); err != nil {
		return snapshot, PlanReport{}, err
	}
	if err := runtimepkg.VerifyRepairPlanReportCurrent(root, snapshot.State, packet); err != nil {
		return snapshot, PlanReport{}, err
	}
	if err := runtimepkg.VerifyRepairPlanReportSources(root, snapshot.State, packet, snapshot.Revision+1); err != nil {
		return snapshot, PlanReport{}, err
	}
 if err := ValidateRepairPlanReportAuthoringStage(root, snapshot.State, req.OwnerAgentID); err != nil {
		return snapshot, PlanReport{}, err
	}
	issueAt, err := repairPlanReportIssueTime(root, snapshot.State, packet)
	if err != nil {
		return snapshot, PlanReport{}, err
	}
	var professional workpacket.RepairPlanReportProfessionalAnswer
	if err := json.Unmarshal(answer, &professional); err != nil {
		return snapshot, PlanReport{}, err
	}
	checks := make([]RepairCheck, len(professional.RedChecks))
	for i, check := range professional.RedChecks {
		checks[i] = RepairCheck{Name: check.Name, Command: check.Command, Result: check.Result, EvidenceRefs: append([]string(nil), check.EvidenceRefs...)}
	}
	receipt, err := newRepairPlanReportReceipt(root, snapshot, packet, answer, answerSHA, actualSource, issueAt)
	if err != nil {
		return snapshot, PlanReport{}, err
	}
	if err := cleanupUnacceptedIssuedPlanReport(root, writer, snapshot, packet, receipt, professional, issueAt); err != nil {
		return snapshot, PlanReport{}, err
	}
	_, ref, err := CreatePlanReport(root, PlanReportRequest{
		Session:      ArtifactRef{ID: packet.Binding.SessionID, Path: packet.Binding.SessionRef, SHA256: packet.Binding.SessionSHA256},
		Plan:         ArtifactRef{ID: packet.Binding.PlanID, Path: packet.Binding.PlanRef, SHA256: packet.Binding.PlanSHA256},
		AssignmentID: packet.Binding.RepairAssignmentID, AgentID: req.OwnerAgentID, ReportID: packet.ReportID,
		PlanText: professional.Plan, RedChecks: checks, ProposedPaths: append([]string(nil), professional.ProposedPaths...),
		OccurredAt: issueAt, issuedReceipt: &receipt,
	})
	if err != nil {
		if recovered, recoveredReport, recoveredErr := recoverAcceptedRepairPlanReport(root, statePath, journalPath, packet, req, answerSHA, actualSource); recoveredErr == nil && recovered {
			return recoveredReport.snapshot, recoveredReport.report, nil
		}
		return snapshot, PlanReport{}, err
	}
	capability := &issuedPlanReportCapability{packet: packet, receipt: receipt, answer: answer, answerSHA: answerSHA, source: actualSource, authorize: req.AuthorizeSource}
	occurredAt := req.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	next, committed, err := SubmitRepairPlanReportToRuntime(root, statePath, journalPath, SubmitPlanReportRequest{
		RuntimeRequest: RuntimeRequest{ExpectedRevision: snapshot.Revision, Actor: req.OwnerAgentID, OccurredAt: occurredAt},
		Report:         ref, issued: capability,
	})
	if err != nil {
		if recovered, recoveredReport, recoveredErr := recoverAcceptedRepairPlanReport(root, statePath, journalPath, packet, req, answerSHA, actualSource); recoveredErr == nil && recovered {
			return recoveredReport.snapshot, recoveredReport.report, nil
		}
		return snapshot, PlanReport{}, err
	}
	if committed.Receipt == nil || !reflect.DeepEqual(*committed.Receipt, receipt) {
		return next, committed, errors.New("original S9 Report consumer returned without the exact issued receipt")
	}
	accepted, ok, err := runtimepkg.VerifiedRepairPlanReportReceipt(root, next.State, packet.ID, req.OwnerAgentID)
	if err != nil {
		return next, committed, err
	}
	if !ok || !reflect.DeepEqual(accepted, receipt) {
		return next, committed, errors.New("original S9 Report consumer did not retain its exact accepted receipt")
	}
	return next, committed, nil
}

func newRepairPlanReportReceipt(root string, snapshot runtimepkg.Snapshot, packet workpacket.RepairPlanReportPacket, answer json.RawMessage, answerSHA, executionRoot string, issueAt time.Time) (runtimepkg.RepairPlanReportReceipt, error) {
	journalBytes, err := controlproof.ReadJournalForState(root, snapshot.State)
	if err != nil {
		return runtimepkg.RepairPlanReportReceipt{}, err
	}
	journal, err := controlproof.InspectJournalData(journalBytes)
	if err != nil {
		return runtimepkg.RepairPlanReportReceipt{}, err
	}
	commitRevision := snapshot.Revision + 1
	if snapshot.Revision < 0 || journal.TailSequence < 0 {
		return runtimepkg.RepairPlanReportReceipt{}, errors.New("S9 receipt cannot predict an invalid original Runtime CAS")
	}
	packetSHA, err := workpacket.RepairPlanReportPacketDigest(packet)
	if err != nil {
		return runtimepkg.RepairPlanReportReceipt{}, err
	}
	b := packet.Binding
	return runtimepkg.RepairPlanReportReceipt{
		Contract: repairPlanReportReceiptContract, PacketID: packet.ID,
		PacketSHA256: packetSHA, AnswerSHA256: answerSHA, SubmittedAnswer: append(json.RawMessage(nil), answer...),
		OwnerAgentID: b.OwnerAgentID, RuntimeID: b.RuntimeID, BaselineGeneration: b.BaselineGeneration, ReviewRound: b.ReviewRound,
		CaseID: b.CaseID, CaseRef: b.CaseRef, CaseSHA256: b.CaseSHA256,
		ContractID: b.ContractID, ContractRef: b.ContractRef, ContractSHA256: b.ContractSHA256,
		SessionID: b.SessionID, SessionRef: b.SessionRef, SessionSHA256: b.SessionSHA256,
		PlanID: b.PlanID, PlanRef: b.PlanRef, PlanSHA256: b.PlanSHA256,
		RepairAssignmentID: b.RepairAssignmentID, RepairAssignmentSHA256: b.RepairAssignmentSHA256,
		ManifestAssignmentID: b.ManifestAssignmentID, ManifestAssignmentSHA256: b.ManifestAssignmentSHA256,
		TeamID: b.TeamID, TaskID: b.TaskID, TaskRef: b.TaskRef, TaskSHA256: b.TaskSHA256,
		ManifestRef: b.ManifestRef, ManifestSHA256: b.ManifestSHA256, ExecutionRoot: executionRoot,
		ReportID: packet.ReportID, ReportRef: filepath.ToSlash(filepath.Join(artifactRoot, "plan-reports", packet.ReportID+".json")),
		WorkComplete: false, IssuanceEventID: packet.IssuanceEventID,
		AcceptedEventID: "evt-s9-issued-plan-report-" + packet.ID, AcceptedIdempotencyKey: "runtime:s9:issued-plan-report:" + packet.ID,
		AcceptedAfterRuntimeRevision: commitRevision, AcceptedSequence: journal.TailSequence + 1,
	}, nil
}

func repairPlanReportIssueTime(root string, state map[string]any, packet workpacket.RepairPlanReportPacket) (time.Time, error) {
	data, err := controlproof.ReadJournalForState(root, state)
	if err != nil {
		return time.Time{}, err
	}
	journal, err := controlproof.InspectJournalData(data)
	if err != nil {
		return time.Time{}, err
	}
	index, ok := journal.EventIndex[packet.IssuanceEventID]
	if !ok {
		return time.Time{}, errors.New("S9 packet issue event has no retained original occurred_at")
	}
	at, err := time.Parse(time.RFC3339Nano, stringField(journal.Events[index]["occurred_at"]))
	if err != nil {
		return time.Time{}, fmt.Errorf("S9 packet issue occurred_at: %w", err)
	}
	return at.UTC(), nil
}

func repairPlanReportSourceRoot(root string, state map[string]any, packet workpacket.RepairPlanReportPacket) (string, error) {
	mainRoot, err := workspace.Canonical(root)
	if err != nil {
		return "", err
	}
	registry, err := workspace.Decode(state)
	if err != nil {
		return "", err
	}
	if registry == nil {
		return mainRoot, nil
	}
	if err := registry.ValidateAuthority(state); err != nil {
		return "", err
	}
	if !repairPlanReportWorkerDeclared(registry, packet) {
		return mainRoot, nil
	}
	return "", errors.New("S9 pre-execution PlanReport has no registered Worker authoring profile; use its Main-root packet path")
}

func repairPlanReportWorkerDeclared(registry *workspace.ExecutionRegistry, packet workpacket.RepairPlanReportPacket) bool {
	if registry == nil {
		return false
	}
	binding := packet.Binding
	identities := map[string]bool{
		binding.RepairAssignmentID:                          true,
		binding.ManifestAssignmentID:                        true,
		manifestAssignmentAlias(binding.RepairAssignmentID): true,
	}
	for key, execution := range registry.Executions {
		if !identities[key] && !identities[execution.AssignmentID] {
			continue
		}
		if execution.RuntimeID == binding.RuntimeID && execution.BaselineGeneration == binding.BaselineGeneration && execution.AgentID == binding.OwnerAgentID {
			return true
		}
	}
	return false
}

func cleanupUnacceptedIssuedPlanReport(root string, writer *runtimepkg.Store, snapshot runtimepkg.Snapshot, packet workpacket.RepairPlanReportPacket, expected runtimepkg.RepairPlanReportReceipt, answer workpacket.RepairPlanReportProfessionalAnswer, issueAt time.Time) error {
	path := filepath.ToSlash(filepath.Join(artifactRoot, "plan-reports", packet.ReportID+".json"))
	data, err := workpacket.ReadFile(root, path, repairPlanReportAnswerLimit)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("preserve unrecognized S9 PlanReport staging path %s: %w", path, err)
	}
	sha := workpacket.Digest(data)
	staged, err := ValidatePlanReport(root, ArtifactRef{Path: path, SHA256: sha})
	if err != nil || staged.Receipt == nil || staged.ReportID != packet.ReportID || staged.AssignmentID != packet.Binding.RepairAssignmentID ||
		staged.AgentID != packet.Binding.OwnerAgentID || staged.SessionID != packet.Binding.SessionID || staged.PlanID != packet.Binding.PlanID ||
		staged.Receipt.PacketID != packet.ID || staged.Receipt.ReportID != packet.ReportID || staged.Receipt.ReportRef != path ||
		staged.Receipt.IssuanceEventID != packet.IssuanceEventID || staged.Receipt.WorkComplete || staged.Receipt.AcceptedEventID != "evt-s9-issued-plan-report-"+packet.ID {
		return fmt.Errorf("preserve existing S9 PlanReport path %s: it is not an exact unaccepted packet staging artifact (%v)", path, err)
	}
	stagedReceipt := *staged.Receipt
	// The only fields allowed to differ from a fresh deterministic retry are
	// the predicted original Runtime revision and journal sequence. A prior
	// failed CAS can make those two coordinates stale while all frozen source,
	// answer and actor evidence remains the same.
	stagedReceipt.AcceptedAfterRuntimeRevision = expected.AcceptedAfterRuntimeRevision
	stagedReceipt.AcceptedSequence = expected.AcceptedSequence
	assignmentRaw, assignmentErr := json.Marshal(packet.Inputs.RepairAssignment)
	var assignment map[string]any
	if assignmentErr == nil {
		assignmentErr = json.Unmarshal(assignmentRaw, &assignment)
	}
	assertionIDs, assertionOK := stringSliceForRepairReport(assignment["assertion_ids"])
	wantChecks := make([]RepairCheck, len(answer.RedChecks))
	for i, check := range answer.RedChecks {
		wantChecks[i] = RepairCheck{Name: check.Name, Command: check.Command, Result: check.Result, EvidenceRefs: append([]string(nil), check.EvidenceRefs...)}
	}
	if assignmentErr != nil || !assertionOK || !reflect.DeepEqual(stagedReceipt, expected) || staged.Status != "reported" ||
		staged.ReportedAt != issueAt.UTC().Format(time.RFC3339Nano) || !reflect.DeepEqual(staged.AssertionIDs, assertionIDs) || staged.Plan != answer.Plan ||
		!reflect.DeepEqual(staged.ProposedPaths, answer.ProposedPaths) || !reflect.DeepEqual(staged.RedChecks, wantChecks) {
		return fmt.Errorf("preserve existing S9 PlanReport path %s: the staged document does not equal this unaccepted answer/source retry", path)
	}
	for _, ref := range existingArtifactRefs(repairPointer(snapshot.State)["plan_report_refs"]) {
		if ref.Path == path {
			return fmt.Errorf("preserve S9 PlanReport staging artifact %s because the current Runtime pointer references it", path)
		}
	}
	removed, err := writer.RemoveUnreferencedArtifact(runtimepkg.ArtifactCleanupRequest{
		ExpectedRevision: snapshot.Revision, ArtifactPath: path, ArtifactSHA256: sha,
		ReferencedPaths: stateArtifactPaths(snapshot.State),
	})
	if err != nil {
		return fmt.Errorf("clean exact unaccepted S9 Report staging artifact %s: %w", path, err)
	}
	if !removed {
		return fmt.Errorf("preserve S9 Report staging artifact %s: pending Runtime state or reference prevents safe cleanup; inspect recovery before retry", path)
	}
	return nil
}

func sameAbsolutePath(left, right string) bool {
	if left == "" || right == "" {
		return false
	}
	a, errA := filepath.Abs(left)
	b, errB := filepath.Abs(right)
	return errA == nil && errB == nil && filepath.Clean(a) == filepath.Clean(b)
}

func sameRepairPlanReportPacket(a, b workpacket.RepairPlanReportPacket) bool {
	left, err := workpacket.RepairPlanReportPacketDigest(a)
	if err != nil {
		return false
	}
	right, err := workpacket.RepairPlanReportPacketDigest(b)
	return err == nil && left == right
}

func containsRepairPlanReportPacket(values []workpacket.RepairPlanReportPacket, expected workpacket.RepairPlanReportPacket) bool {
	for _, value := range values {
		if value.ID == expected.ID && sameRepairPlanReportPacket(value, expected) {
			return true
		}
	}
	return false
}

func integerForRepairReport(value any) (int, bool) {
	switch n := value.(type) {
	case int:
		return n, true
	case int64:
		return int(n), int64(int(n)) == n
	case float64:
		return int(n), float64(int(n)) == n
	case json.Number:
		i, err := n.Int64()
		return int(i), err == nil && int64(int(i)) == i
	default:
		return 0, false
	}
}

func loadAcceptedRepairPlanReport(root string, state map[string]any, receipt runtimepkg.RepairPlanReportReceipt) (PlanReport, error) {
	pointer := repairPointer(state)
	for _, ref := range existingArtifactRefs(pointer["plan_report_refs"]) {
		if ref.Path == receipt.ReportRef {
			report, err := ValidatePlanReport(root, ref)
			if err != nil {
				return PlanReport{}, err
			}
			if report.Receipt == nil || !reflect.DeepEqual(*report.Receipt, receipt) {
				return PlanReport{}, errors.New("S9 accepted report artifact does not contain its exact original receipt")
			}
			return report, nil
		}
	}
	if pointer != nil && stringField(pointer["plan_report_ref"]) == receipt.ReportRef {
		ref := ArtifactRef{Path: receipt.ReportRef, SHA256: stringField(pointer["plan_report_sha256"])}
		report, err := ValidatePlanReport(root, ref)
		if err != nil {
			return PlanReport{}, err
		}
		if report.Receipt == nil || !reflect.DeepEqual(*report.Receipt, receipt) {
			return PlanReport{}, errors.New("S9 accepted report artifact does not contain its exact original receipt")
		}
		return report, nil
	}
	return PlanReport{}, errors.New("S9 accepted receipt is not retained by the original repair pointer")
}

type repairExecutionBeginBinding struct {
	runtimeID          string
	baselineGeneration int
	reviewRound        int
	sessionID          string
	sessionRef         ArtifactRef
	contractID         string
	contractRef        ArtifactRef
	planID             string
	planRef            ArtifactRef
	reportRef          ArtifactRef
	reportRefs         []ArtifactRef
}

func captureRepairExecutionBeginBinding(root string, state map[string]any) (repairExecutionBeginBinding, PlanReport, error) {
	pointer := repairPointer(state)
	if pointer == nil || stringField(pointer["status"]) != "reproducing" {
		return repairExecutionBeginBinding{}, PlanReport{}, errors.New("S9 execution begin requires status=reproducing; submit a PlanReport first")
	}
	lifecycle := mapField(state, "lifecycle")
	if lifecycle["state"] != "bug_resolution" || lifecycle["phase"] != "reproducing" {
		return repairExecutionBeginBinding{}, PlanReport{}, errors.New("S9 execution begin requires the original bug_resolution.reproducing cursor")
	}
	runtimeID := stringField(state["runtime_id"])
	baseGeneration, ok := integerForRepairReport(mapField(state, "baseline")["generation"])
	if runtimeID == "" || !ok || baseGeneration < 0 {
		return repairExecutionBeginBinding{}, PlanReport{}, errors.New("S9 execution begin requires an exact Runtime and baseline generation")
	}
	round, ok := integerForRepairReport(mapField(state, "review")["round"])
	if !ok || round < 1 {
		return repairExecutionBeginBinding{}, PlanReport{}, errors.New("S9 execution begin requires a valid current Review round")
	}
	sessionID := stringField(pointer["session_id"])
	sessionRef, err := pointerArtifact(pointer, "path", "sha256", "current RepairSession")
	if err != nil || sessionID == "" {
		return repairExecutionBeginBinding{}, PlanReport{}, errors.New("S9 execution begin requires the exact current RepairSession")
	}
	contractID := stringField(pointer["contract_id"])
	contractRef, err := pointerArtifact(pointer, "contract_ref", "contract_sha256", "current approved RepairContract")
	if err != nil || contractID == "" {
		return repairExecutionBeginBinding{}, PlanReport{}, errors.New("S9 execution begin requires the exact current approved RepairContract")
	}
	planRef, err := pointerArtifact(pointer, "plan_ref", "plan_sha256", "current RepairPlan")
	if err != nil {
		return repairExecutionBeginBinding{}, PlanReport{}, err
	}
	reportRef, err := pointerArtifact(pointer, "plan_report_ref", "plan_report_sha256", "current RepairPlan report")
	if err != nil {
		return repairExecutionBeginBinding{}, PlanReport{}, err
	}
	reportRefs, err := strictPlanReportRefs(pointer["plan_report_refs"], reportRef)
	if err != nil {
		return repairExecutionBeginBinding{}, PlanReport{}, err
	}
	binding := repairExecutionBeginBinding{
		runtimeID: runtimeID, baselineGeneration: baseGeneration, reviewRound: round,
		sessionID: sessionID, sessionRef: sessionRef, contractID: contractID, contractRef: contractRef,
		planRef: planRef, reportRef: reportRef, reportRefs: reportRefs,
	}
	report, err := validateRepairExecutionBeginState(root, state, binding)
	if err != nil {
		return repairExecutionBeginBinding{}, PlanReport{}, err
	}
	plan, err := ValidateRepairPlan(root, planRef)
	if err != nil {
		return repairExecutionBeginBinding{}, PlanReport{}, err
	}
	binding.planID = plan.PlanID
	return binding, report, nil
}

func validateRepairExecutionBeginState(root string, state map[string]any, expected repairExecutionBeginBinding) (PlanReport, error) {
	pointer := repairPointer(state)
	lifecycle := mapField(state, "lifecycle")
	baseGeneration, baseOK := integerForRepairReport(mapField(state, "baseline")["generation"])
	round, roundOK := integerForRepairReport(mapField(state, "review")["round"])
	if pointer == nil || stringField(state["runtime_id"]) != expected.runtimeID || !baseOK || baseGeneration != expected.baselineGeneration || !roundOK || round != expected.reviewRound ||
		lifecycle["state"] != "bug_resolution" || lifecycle["phase"] != "reproducing" || stringField(pointer["status"]) != "reproducing" ||
		stringField(pointer["session_id"]) != expected.sessionID || stringField(pointer["contract_id"]) != expected.contractID {
		return PlanReport{}, errors.New("S9 Runtime, cursor, generation, Session, or Contract changed before execution begin")
	}
	sessionRef, err := pointerArtifact(pointer, "path", "sha256", "current RepairSession")
	if err != nil || sessionRef != expected.sessionRef {
		return PlanReport{}, errors.New("S9 current RepairSession reference changed before execution begin")
	}
	contractRef, err := pointerArtifact(pointer, "contract_ref", "contract_sha256", "current approved RepairContract")
	if err != nil || contractRef != expected.contractRef {
		return PlanReport{}, errors.New("S9 current approved Contract reference changed before execution begin")
	}
	planRef, err := pointerArtifact(pointer, "plan_ref", "plan_sha256", "current RepairPlan")
	if err != nil || planRef != expected.planRef {
		return PlanReport{}, errors.New("S9 current RepairPlan reference changed before execution begin")
	}
	reportRef, err := pointerArtifact(pointer, "plan_report_ref", "plan_report_sha256", "current RepairPlan report")
	if err != nil || reportRef != expected.reportRef {
		return PlanReport{}, errors.New("S9 current singular PlanReport reference changed before execution begin")
	}
	reportRefs, err := strictPlanReportRefs(pointer["plan_report_refs"], reportRef)
	if err != nil || !sameArtifactRefSequence(reportRefs, expected.reportRefs) {
		return PlanReport{}, errors.New("S9 PlanReport reference set changed before execution begin")
	}
	session, err := ValidateRepairSession(root, expected.sessionRef)
	if err != nil {
		return PlanReport{}, fmt.Errorf("current S9 RepairSession is invalid: %w", err)
	}
	contract, err := ValidateApprovedContractRef(root, ContractRef{Path: expected.contractRef.Path, SHA256: expected.contractRef.SHA256})
	if err != nil {
		return PlanReport{}, fmt.Errorf("current S9 approved Contract is invalid: %w", err)
	}
	plan, err := ValidateRepairPlan(root, expected.planRef)
	if err != nil {
		return PlanReport{}, fmt.Errorf("current S9 RepairPlan is invalid: %w", err)
	}
	if session.SessionID != expected.sessionID || session.RuntimeID != expected.runtimeID || session.BaselineGeneration != expected.baselineGeneration ||
		session.ContractRef != expected.contractRef.Path || session.ContractSHA256 != expected.contractRef.SHA256 || contract.ContractID != expected.contractID ||
		plan.PlanID != expected.planID && expected.planID != "" || plan.SessionID != expected.sessionID || plan.ContractID != expected.contractID ||
		plan.ContractRef != expected.contractRef.Path || plan.ContractSHA256 != expected.contractRef.SHA256 {
		return PlanReport{}, errors.New("S9 immutable Session, Contract, and Plan identity tuple differs from current Runtime authority")
	}
	if len(plan.Assignments) == 0 {
		return PlanReport{}, errors.New("S9 current RepairPlan has no Assignments to cover")
	}
	assignments := make(map[string]RepairAssignment, len(plan.Assignments))
	for _, assignment := range plan.Assignments {
		if assignment.AssignmentID == "" {
			return PlanReport{}, errors.New("S9 current RepairPlan contains an empty Assignment identity")
		}
		if _, duplicate := assignments[assignment.AssignmentID]; duplicate {
			return PlanReport{}, fmt.Errorf("S9 current RepairPlan duplicates Assignment %s", assignment.AssignmentID)
		}
		assignments[assignment.AssignmentID] = assignment
	}
	owners := stringMapField(pointer["assignment_owners"])
	packets, err := runtimepkg.RepairPlanReportPackets(state)
	if err != nil {
		return PlanReport{}, err
	}
	packetsByAssignment := map[string]workpacket.RepairPlanReportPacket{}
	for _, packet := range packets {
		binding := packet.Binding
		if binding.RuntimeID != expected.runtimeID || binding.BaselineGeneration != expected.baselineGeneration || binding.ReviewRound != expected.reviewRound ||
			binding.SessionID != expected.sessionID || binding.PlanID != plan.PlanID {
			continue
		}
		if _, duplicate := packetsByAssignment[binding.RepairAssignmentID]; duplicate {
			return PlanReport{}, fmt.Errorf("S9 RepairAssignment %s has multiple issued PlanReport packets", binding.RepairAssignmentID)
		}
		packetsByAssignment[binding.RepairAssignmentID] = packet
	}
	reportsByAssignment := make(map[string]PlanReport, len(reportRefs))
	var singularReport PlanReport
	for index, ref := range reportRefs {
		report, err := ValidatePlanReport(root, ref)
		if err != nil {
			return PlanReport{}, fmt.Errorf("current S9 PlanReport[%d] is invalid: %w", index, err)
		}
		if report.SessionID != expected.sessionID || report.PlanID != plan.PlanID {
			return PlanReport{}, fmt.Errorf("S9 PlanReport %s is not bound to the current Session and Plan", ref.Path)
		}
		assignment, found := assignments[report.AssignmentID]
		if !found {
			return PlanReport{}, fmt.Errorf("S9 PlanReport %s names Assignment %s outside the current Plan", ref.Path, report.AssignmentID)
		}
		if _, duplicate := reportsByAssignment[report.AssignmentID]; duplicate {
			return PlanReport{}, fmt.Errorf("S9 current pointer has multiple PlanReports for Assignment %s", report.AssignmentID)
		}
		if err := exactIDs(assignment.AssertionIDs, report.AssertionIDs); err != nil {
			return PlanReport{}, fmt.Errorf("S9 PlanReport %s assertion coverage differs from Assignment %s: %w", ref.Path, report.AssignmentID, err)
		}
		if assignment.OwnerAgentID != "" && report.AgentID != assignment.OwnerAgentID {
			return PlanReport{}, fmt.Errorf("S9 PlanReport %s owner differs from immutable Assignment %s", ref.Path, report.AssignmentID)
		}
		if owner := owners[report.AssignmentID]; owner != "" && report.AgentID != owner {
			return PlanReport{}, fmt.Errorf("S9 PlanReport %s owner differs from current Assignment owner %s", ref.Path, owner)
		}
		packet, issued := packetsByAssignment[report.AssignmentID]
		if issued {
			if report.Receipt == nil || packet.ReportID != report.ReportID || report.Receipt.PacketID != packet.ID {
				return PlanReport{}, fmt.Errorf("S9 issued Assignment %s is not covered by its exact accepted PlanReport packet", report.AssignmentID)
			}
			if err := runtimepkg.VerifyRepairPlanReportCurrent(root, state, packet); err != nil {
				return PlanReport{}, err
			}
			verified, accepted, err := runtimepkg.VerifiedRepairPlanReportReceipt(root, state, packet.ID, report.AgentID)
			if err != nil || !accepted || !reflect.DeepEqual(verified, *report.Receipt) {
				return PlanReport{}, fmt.Errorf("S9 Assignment %s receipt is not accepted by the original Runtime journal: %v", report.AssignmentID, err)
			}
		} else if report.Receipt != nil {
			return PlanReport{}, fmt.Errorf("S9 PlanReport %s has a typed receipt without its registered packet", ref.Path)
		}
		if ref.Path == expected.reportRef.Path && ref.SHA256 == expected.reportRef.SHA256 {
			singularReport = report
		}
		reportsByAssignment[report.AssignmentID] = report
	}
	missing := make([]string, 0)
	for assignmentID := range assignments {
		if _, ok := reportsByAssignment[assignmentID]; !ok {
			missing = append(missing, assignmentID)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return PlanReport{}, fmt.Errorf("S9 execution begin is blocked: missing PlanReport for assignments %v; submit one report per assignment", missing)
	}
	if singularReport.ReportID == "" {
		return PlanReport{}, errors.New("S9 singular current PlanReport reference is not present in the full report set")
	}
	return singularReport, nil
}

func strictPlanReportRefs(raw any, singular ArtifactRef) ([]ArtifactRef, error) {
	refs := []ArtifactRef{}
	values, exists := raw.([]any)
	if !exists || len(values) == 0 {
		if raw != nil {
			return nil, errors.New("S9 plan_report_refs is not a non-empty reference array")
		}
		refs = append(refs, singular)
	} else {
		seen := map[string]string{}
		for index, value := range values {
			row, ok := value.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("S9 plan_report_refs[%d] is not an object", index)
			}
			ref := ArtifactRef{Path: stringField(row["path"]), SHA256: stringField(row["sha256"])}
			if ref.Path == "" || ref.SHA256 == "" {
				return nil, fmt.Errorf("S9 plan_report_refs[%d] lacks exact path/SHA", index)
			}
			if old, duplicate := seen[ref.Path]; duplicate {
				if old != ref.SHA256 {
					return nil, fmt.Errorf("S9 plan_report_refs has conflicting SHA values for %s", ref.Path)
				}
				return nil, fmt.Errorf("S9 plan_report_refs duplicates %s", ref.Path)
			}
			seen[ref.Path] = ref.SHA256
			refs = append(refs, ref)
		}
	}
	for _, ref := range refs {
		if ref.Path == singular.Path && ref.SHA256 == singular.SHA256 {
			return refs, nil
		}
	}
	return nil, errors.New("S9 singular current PlanReport is absent from plan_report_refs")
}

func sameArtifactRefSequence(left, right []ArtifactRef) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i].Path != right[i].Path || left[i].SHA256 != right[i].SHA256 {
			return false
		}
	}
	return true
}

type recoveredRepairPlanReport struct {
	snapshot runtimepkg.Snapshot
	report   PlanReport
}

func recoverAcceptedRepairPlanReport(root, statePath, journalPath string, packet workpacket.RepairPlanReportPacket, req SubmitIssuedRepairPlanReportRequest, answerSHA, sourceRoot string) (bool, recoveredRepairPlanReport, error) {
	snapshot, _, err := readIssuedRepairPlanReportRuntime(root, statePath, journalPath)
	if err != nil {
		return false, recoveredRepairPlanReport{}, err
	}
	if req.AuthorizeSource != nil {
		if err := req.AuthorizeSource(root, snapshot.State); err != nil {
			return false, recoveredRepairPlanReport{}, fmt.Errorf("S9 replay source callback no longer matches: %w", err)
		}
	}
	actual, err := repairPlanReportSourceRoot(root, snapshot.State, packet)
	if err != nil || !sameAbsolutePath(actual, sourceRoot) {
		return false, recoveredRepairPlanReport{}, errors.New("S9 replay source root changed")
	}
	answerBytes, err := workpacket.ReadFile(actual, packet.AnswerPath, repairPlanReportAnswerLimit)
	if err != nil {
		return false, recoveredRepairPlanReport{}, err
	}
	answer, problems := workpacket.ValidateRepairPlanReportAnswer(packet, answerBytes)
	if len(problems) != 0 || workpacket.Digest(answer) != answerSHA {
		return false, recoveredRepairPlanReport{}, runtimepkg.ErrSubmissionConflict
	}
	receipt, ok, err := runtimepkg.VerifiedRepairPlanReportReceipt(root, snapshot.State, packet.ID, req.OwnerAgentID)
	if err != nil || !ok {
		return false, recoveredRepairPlanReport{}, err
	}
	if receipt.AnswerSHA256 != answerSHA || !sameAbsolutePath(receipt.ExecutionRoot, sourceRoot) {
		return false, recoveredRepairPlanReport{}, runtimepkg.ErrSubmissionConflict
	}
	report, err := loadAcceptedRepairPlanReport(root, snapshot.State, receipt)
	if err != nil {
		return false, recoveredRepairPlanReport{}, err
	}
	return true, recoveredRepairPlanReport{snapshot: snapshot, report: report}, nil
}

// readIssuedRepairPlanReportRuntime is limited to the typed S9 consumer and
// its explicit recovery check. Unlike observation-only Runtime readers, the
// original mutation-capable Store recovers a durable pending write under the
// Runtime lock before this path reads author bytes or attempts receipt lookup.
// A recovered write is still accepted only after the exact original receipt,
// packet issue, answer bytes, owner, and source/session proof below verify.
func readIssuedRepairPlanReportRuntime(root, statePath, journalPath string) (runtimepkg.Snapshot, *runtimepkg.Store, error) {
	writer := runtimepkg.NewWriter(statePath, journalPath, root, semantic.RuntimeCandidateValidator{})
	snapshot, err := writer.Snapshot()
	if err != nil {
		return runtimepkg.Snapshot{}, nil, err
	}
	return snapshot, writer, nil
}

// IssuedRepairPlanReportReceipt only reconstructs an accepted result from the
// original Report artifact and Runtime journal. It never interprets the draft
// as permission or reads an unaccepted authoring file.
func IssuedRepairPlanReportReceipt(root string, state map[string]any, packetID, owner string) (runtimepkg.RepairPlanReportReceipt, bool, error) {
	return runtimepkg.VerifiedRepairPlanReportReceipt(root, state, packetID, owner)
}
