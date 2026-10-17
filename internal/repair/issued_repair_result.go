package repair

import (
	"context"
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
	"github.com/entroforge/go-system-builder/internal/assignmentrecord"
	"github.com/entroforge/go-system-builder/internal/checkexec"
	"github.com/entroforge/go-system-builder/internal/controlproof"
	"github.com/entroforge/go-system-builder/internal/integration"
	runtimepkg "github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/schema"
	"github.com/entroforge/go-system-builder/internal/semantic"
	"github.com/entroforge/go-system-builder/internal/workpacket"
	"github.com/entroforge/go-system-builder/internal/workspace"
)

const repairResultAnswerLimit int64 = 20 << 20

const repairResultCandidateContract = "s9-repair-result-candidate-v1"
const repairResultFinalContract = "s9-repair-result-receipt-v1"

var errRepairResultExecutionNotRegistered = errors.New("S9 RepairResult assignment has no registered Worker execution")

type AcceptIssuedRepairResultCandidateRequest struct {
	ExpectedRevision int
	PacketID         string
	AnswerPath       string
	OwnerAgentID     string
	SourceRoot       string
	// AuthorizeSource is an optional integration callback. The original
	// registered packet/execution readers remain the source of domain facts.
	AuthorizeSource  func(root string, state map[string]any) error
	EnsureCompletion func(context.Context, runtimepkg.Snapshot, workpacket.RepairResultPacket, workspace.Execution, RepairResultRequest) (runtimepkg.RepairResultCompletionReport, error)
	OccurredAt       time.Time
}

type SubmitIssuedRepairResultRequest struct {
	ExpectedRevision int
	PacketID         string
	OwnerAgentID     string
	MainProof        runtimepkg.RepairResultMainProof
	OccurredAt       time.Time
}

type issuedRepairResultCapability struct {
	packet         workpacket.RepairResultPacket
	provenance     runtimepkg.RepairResultCandidateProvenance
	beforeRevision int
}

type issuedRepairResultFinalCapability struct {
	packet     workpacket.RepairResultPacket
	provenance runtimepkg.RepairResultCandidateProvenance
	receipt    runtimepkg.RepairResultReceipt
	ctx        context.Context
	mainRoot   string
	cleanup    func(RepairResult) error
}

// IssueRepairResultPackets freezes the current, registered S9 PlanReport and
// execution definition into the neutral packet registry. The packet is only
// an input contract: it does not accept a result, deliver a candidate, or
// complete the original Main integration checkpoint.
func IssueRepairResultPackets(root, statePath, journalPath string, expectedRevision int, ownerFilter string) ([]workpacket.RepairResultPacket, error) {
	ctx, unlock, err := integration.LockWorkspace(context.Background(), root)
	if err != nil {
		return nil, err
	}
	defer unlock()
	current, writer, err := readIssuedRepairResultRuntime(root, statePath, journalPath)
	if err != nil {
		return nil, err
	}
	if mainRoot := workspaceMainRoot(root, current.State); !sameAbsolutePath(mainRoot, root) {
		return nil, errors.New("S9 packet issuance root must be the original registered Main checkout")
	}
	candidates, err := buildRepairResultPackets(ctx, root, current.State, ownerFilter)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		if err := checkRevision(expectedRevision, current.Revision, "runtime S9 RepairResult packet issue"); err != nil {
			return nil, err
		}
		return []workpacket.RepairResultPacket{}, nil
	}
	issued := make([]workpacket.RepairResultPacket, 0, len(candidates))
	callerRevisionChecked := false
	for _, packet := range candidates {
		if prior, found, err := runtimepkg.FindRepairResultPacket(current.State, packet.ID); err != nil {
			return nil, err
		} else if found {
			if !sameRepairResultPacket(prior, packet) {
				return nil, fmt.Errorf("S9 packet identity %s is already registered with different frozen sources", packet.ID)
			}
			if err := runtimepkg.VerifyRepairResultPacketIssueAndCurrent(root, current.State, prior); err != nil {
				return nil, fmt.Errorf("S9 packet %s failed exact issuance/current verification: %w", packet.ID, err)
			}
			issued = append(issued, prior)
			continue
		}
		// The caller's explicit revision pins the initial issuance transaction.
		// Publishing a sibling packet advances Runtime; subsequent packets in
		// this same serialized batch use the refreshed internal snapshot/CAS.
		// Exact prior issuances above remain replayable before this check.
		if !callerRevisionChecked {
			if err := checkRevision(expectedRevision, current.Revision, "runtime S9 RepairResult packet issue"); err != nil {
				return nil, err
			}
			callerRevisionChecked = true
		}
		if err := runtimepkg.VerifyRepairResultSources(root, current.State, packet, current.Revision+1); err != nil {
			return nil, fmt.Errorf("S9 packet source closure for Assignment %s: %w", packet.Binding.Sources.RepairAssignmentID, err)
		}
		if err := issueOneRepairResultPacket(ctx, root, writer, current, packet); err != nil {
			return nil, err
		}
		latest, err := writer.Snapshot()
		if err != nil {
			return nil, err
		}
		stored, found, err := runtimepkg.FindRepairResultPacket(latest.State, packet.ID)
		if err != nil || !found || !sameRepairResultPacket(stored, packet) {
			return nil, fmt.Errorf("S9 packet %s was not retained exactly by the original Runtime writer: %v", packet.ID, err)
		}
		if err := runtimepkg.VerifyRepairResultPacketIssueAndCurrent(root, latest.State, stored); err != nil {
			return nil, err
		}
		issued = append(issued, stored)
		current = latest
	}
	return issued, nil
}

func issueOneRepairResultPacket(ctx context.Context, root string, writer *runtimepkg.Store, current runtimepkg.Snapshot, packet workpacket.RepairResultPacket) error {
	rowBytes, err := json.Marshal(packet)
	if err != nil {
		return err
	}
	if err := schema.NewEmbeddedValidator().ValidateBytes("repair-result-work-packet.schema.json", rowBytes); err != nil {
		return err
	}
	var row map[string]any
	if err := json.Unmarshal(rowBytes, &row); err != nil {
		return err
	}
	from := cursor(current.State)
	mutation := runtimepkg.Mutation{
		EventID: packet.IssuanceEventID, TransitionID: "PACKET-ISSUE", Event: "work_packet_issued",
		Actor: "orchestrator", IdempotencyKey: "packet:" + packet.ID, RuntimeID: packet.Binding.Sources.RuntimeID,
		From: from, To: from, BaselineGeneration: packet.Binding.Sources.BaselineGeneration,
		OccurredAt: time.Now().UTC(), RetainLastTransition: true,
		Apply: func(state map[string]any) error {
			if err := runtimepkg.VerifyRepairResultSources(root, state, packet, current.Revision+1); err != nil {
				return err
			}
			fresh, err := buildRepairResultPackets(ctx, root, state, packet.Binding.Sources.OwnerAgentID)
			if err != nil || !containsRepairResultPacket(fresh, packet) {
				return fmt.Errorf("S9 dispatch/source tuple changed before RepairResult packet publication: %v", err)
			}
			if prior, found, err := runtimepkg.FindRepairResultPacket(state, packet.ID); err != nil {
				return err
			} else if found {
				if sameRepairResultPacket(prior, packet) {
					return errors.New("identical S9 packet appeared during issuance; recover from the original PACKET-ISSUE event")
				}
				return fmt.Errorf("S9 packet ID %s conflicts with an existing Runtime packet", packet.ID)
			}
			entities, _ := state["entities"].(map[string]any)
			if entities == nil {
				entities = map[string]any{}
				state["entities"] = entities
			}
			rows, _ := entities["work_packets"].([]any)
			entities["work_packets"] = append(rows, row)
			return nil
		},
	}
	_, err = writer.Update(current.Revision, mutation)
	if err == nil {
		return nil
	}
	latest, readErr := writer.Snapshot()
	if readErr == nil {
		stored, found, findErr := runtimepkg.FindRepairResultPacket(latest.State, packet.ID)
		if findErr == nil && found && sameRepairResultPacket(stored, packet) && runtimepkg.VerifyRepairResultPacketIssue(root, latest.State, stored) == nil {
			return nil
		}
	}
	return fmt.Errorf("issue S9 RepairResult packet %s: %w", packet.ID, err)
}

func buildRepairResultPackets(ctx context.Context, root string, state map[string]any, ownerFilter string) ([]workpacket.RepairResultPacket, error) {
	if !runtimepkg.PacketEnabled(state) {
		return nil, errors.New("S9 typed RepairResult packets require the explicit v1.2 packet contract")
	}
	life, _ := state["lifecycle"].(map[string]any)
	pointer := repairPointer(state)
	if pointer == nil || pointer["status"] != "repairing" || life["state"] != "bug_resolution" || life["phase"] != "fixing" {
		return []workpacket.RepairResultPacket{}, nil
	}
	planRef, err := pointerArtifact(pointer, "plan_ref", "plan_sha256", "current RepairPlan")
	if err != nil {
		return nil, err
	}
	plan, err := ValidateRepairPlan(root, planRef)
	if err != nil {
		return nil, err
	}
	owners := stringMapField(pointer["assignment_owners"])
	packets := []workpacket.RepairResultPacket{}
	for _, assignment := range plan.Assignments {
		owner := owners[assignment.AssignmentID]
		if owner == "" || (ownerFilter != "" && owner != ownerFilter) {
			continue
		}
		if err := validateRepairAssignmentReady(root, pointer, plan, assignment); err != nil {
			// Dependencies and locks are original S9 scheduler gates. Do not issue
			// an authoring packet until this Assignment is the original ready slot.
			continue
		}
		reportRef, ok := planReportForAssignment(root, pointer["plan_report_refs"], assignment.AssignmentID)
		if !ok {
			continue
		}
		report, err := ValidatePlanReport(root, reportRef)
		if err != nil {
			return nil, err
		}
		if report.Receipt == nil || report.AgentID != owner || report.AssignmentID != assignment.AssignmentID || report.PlanID != plan.PlanID {
			return nil, fmt.Errorf("S9 typed RepairResult requires the exact accepted PlanReport receipt for %s", assignment.AssignmentID)
		}
		reportPacket, found, err := runtimepkg.FindRepairPlanReportPacket(state, report.Receipt.PacketID)
		if err != nil || !found || reportPacket.ReportID != report.ReportID || reportPacket.Binding.OwnerAgentID != owner {
			return nil, fmt.Errorf("S9 original PlanReport packet for %s is unavailable: %v", assignment.AssignmentID, err)
		}
		acceptedReport, accepted, err := runtimepkg.VerifiedRepairPlanReportReceipt(root, state, reportPacket.ID, owner)
		if err != nil || !accepted || acceptedReport.ReportRef != reportRef.Path || acceptedReport.ReportID != report.ReportID || acceptedReport.WorkComplete {
			return nil, fmt.Errorf("S9 PlanReport for %s is not the exact accepted pre-execution report: %v", assignment.AssignmentID, err)
		}
		if err := runtimepkg.VerifyRepairPlanReportPacketIssue(root, state, reportPacket); err != nil {
			return nil, err
		}
		execution, err := repairResultExecutionFromState(ctx, state, reportPacket.Binding, root)
		if err != nil {
			if errors.Is(err, errRepairResultExecutionNotRegistered) && ownerFilter == "" {
				// A global issue call covers only Assignments that have reached an
				// original prepared Worker execution. An unprepared sibling is not a
				// malformed packet source and must not hide another ready target.
				continue
			}
			if errors.Is(err, errRepairResultExecutionNotRegistered) {
				return nil, fmt.Errorf("S9 Assignment %s has no prepared Worker execution for requested owner %s; prepare that Assignment before issuing its result packet", assignment.AssignmentID, ownerFilter)
			}
			return nil, err
		}
		executionRaw, err := json.Marshal(execution)
		if err != nil {
			return nil, err
		}
		var executionMap map[string]any
		if err := json.Unmarshal(executionRaw, &executionMap); err != nil {
			return nil, err
		}
		projection, err := runtimepkg.ProjectRepairResultExecution(executionMap)
		if err != nil {
			return nil, err
		}
		reportBytes, err := workpacket.ReadFile(root, reportRef.Path, repairResultAnswerLimit)
		if err != nil || workpacket.Digest(reportBytes) != reportRef.SHA256 {
			return nil, fmt.Errorf("S9 accepted PlanReport bytes changed: %v", err)
		}
		rework := json.RawMessage("null")
		reworkSHA := ""
		if execution.ReworkRef != "" {
			reworkBytes, err := workpacket.ReadFile(root, execution.ReworkRef, repairResultAnswerLimit)
			if err != nil {
				return nil, fmt.Errorf("read exact S9 rework source: %w", err)
			}
			rework = json.RawMessage(reworkBytes)
			reworkSHA = workpacket.Digest(reworkBytes)
		}
		binding := workpacket.RepairResultBinding{
			Sources: reportPacket.Binding, PlanReportRef: reportRef.Path, PlanReportSHA256: reportRef.SHA256,
			ExecutionSHA256: workpacket.Digest(projection), ReworkRef: execution.ReworkRef, ReworkSHA256: reworkSHA,
		}
		resume, err := runtimepkg.RepairResultResumeBinding(root, state, binding)
		if err != nil {
			return nil, fmt.Errorf("derive S9 Resume predecessor for Assignment %s: %w", assignment.AssignmentID, err)
		}
		binding.Resume = resume
		inputs := workpacket.RepairResultInputs{
			Sources: reportPacket.Inputs, PlanReport: json.RawMessage(reportBytes), Execution: projection, ReworkReceipt: rework,
		}
		packet, err := workpacket.BuildRepairResult(binding, inputs)
		if err != nil {
			return nil, fmt.Errorf("build S9 RepairResult packet for %s: %w", assignment.AssignmentID, err)
		}
		if err := runtimepkg.VerifyRepairResultPacketCurrent(root, state, packet); err != nil {
			return nil, err
		}
		packets = append(packets, packet)
	}
	return packets, nil
}

func repairResultExecutionFromState(ctx context.Context, state map[string]any, binding workpacket.RepairPlanReportBinding, root string) (workspace.Execution, error) {
	registry, err := workspace.Decode(state)
	if err != nil {
		return workspace.Execution{}, fmt.Errorf("decode original S9 Worker execution registry: %w", err)
	}
	if registry == nil {
		return workspace.Execution{}, errRepairResultExecutionNotRegistered
	}
	if err := registry.ValidateAuthority(state); err != nil {
		return workspace.Execution{}, err
	}
	identities := map[string]bool{binding.RepairAssignmentID: true, binding.ManifestAssignmentID: true, manifestAssignmentAlias(binding.RepairAssignmentID): true}
	var selected *workspace.Execution
	registeredIdentity := false
	for key, execution := range registry.Executions {
		if !identities[key] && !identities[execution.AssignmentID] {
			continue
		}
		registeredIdentity = true
		if key != execution.AssignmentID {
			return workspace.Execution{}, fmt.Errorf("S9 registered Worker execution key %q differs from its Assignment identity %q", key, execution.AssignmentID)
		}
		if execution.RuntimeID != binding.RuntimeID || execution.BaselineGeneration != binding.BaselineGeneration || execution.AgentID != binding.OwnerAgentID {
			return workspace.Execution{}, fmt.Errorf("S9 registered Worker execution %s differs from the issued Runtime/generation/owner tuple", execution.AssignmentID)
		}
		if selected != nil {
			return workspace.Execution{}, errors.New("S9 typed RepairResult has ambiguous canonical/manifest execution")
		}
		copy := execution
		selected = &copy
	}
	if selected == nil {
		if !registeredIdentity {
			return workspace.Execution{}, errRepairResultExecutionNotRegistered
		}
		return workspace.Execution{}, errors.New("S9 typed RepairResult has no unique registered execution for this owner/Assignment")
	}
	if selected.Status != "ready" || selected.BootstrapSHA256 == "" || !filepath.IsAbs(selected.Path) || filepath.Clean(selected.Path) != selected.Path {
		return workspace.Execution{}, errors.New("S9 typed RepairResult requires a ready, bootstrapped canonical Worker execution")
	}
	if err := registry.ValidateExecution(ctx, *selected, false); err != nil {
		return workspace.Execution{}, err
	}
	if err := registry.ValidateInputs(*selected); err != nil {
		return workspace.Execution{}, err
	}
	return *selected, nil
}

func containsRepairResultPacket(values []workpacket.RepairResultPacket, wanted workpacket.RepairResultPacket) bool {
	for _, value := range values {
		if sameRepairResultPacket(value, wanted) {
			return true
		}
	}
	return false
}

func sameRepairResultPacket(a, b workpacket.RepairResultPacket) bool {
	// Both packets must already come from the canonical builder or the Runtime
	// reader, which validates Schema and canonical rebuild before returning.
	// Equality is not issuance, source, current-slot or caller authorization.
	return reflect.DeepEqual(a, b)
}

func readIssuedRepairResultRuntime(root, statePath, journalPath string) (runtimepkg.Snapshot, *runtimepkg.Store, error) {
	writer := runtimepkg.NewWriter(statePath, journalPath, root, semantic.RuntimeCandidateValidator{})
	snapshot, err := writer.Snapshot()
	if err != nil {
		return runtimepkg.Snapshot{}, nil, err
	}
	return snapshot, writer, nil
}

// AcceptIssuedRepairResultCandidate saves the original S9 answer through the
// existing WorkspaceCandidate writer and existing WORKSPACE Runtime CAS.
// Professional input never supplies source/check hashes, candidate identity,
// or a completion bit; the source tree and runner receipts are captured here.
func AcceptIssuedRepairResultCandidate(ctx context.Context, root, statePath, journalPath string, req AcceptIssuedRepairResultCandidateRequest) (runtimepkg.Snapshot, WorkspaceCandidate, ArtifactRef, error) {
	lockedCtx, unlock, err := integration.LockWorkspace(ctx, root)
	if err != nil {
		return runtimepkg.Snapshot{}, WorkspaceCandidate{}, ArtifactRef{}, err
	}
	defer unlock()
	ctx = lockedCtx
	mainSnapshot, writer, err := readIssuedRepairResultRuntime(root, statePath, journalPath)
	if err != nil {
		return runtimepkg.Snapshot{}, WorkspaceCandidate{}, ArtifactRef{}, err
	}
	if mainRoot := workspaceMainRoot(root, mainSnapshot.State); !sameAbsolutePath(mainRoot, root) {
		return mainSnapshot, WorkspaceCandidate{}, ArtifactRef{}, errors.New("S9 candidate receiver root must be the original registered Main checkout")
	}
	packet, found, err := runtimepkg.FindRepairResultPacket(mainSnapshot.State, req.PacketID)
	if err != nil || !found {
		return mainSnapshot, WorkspaceCandidate{}, ArtifactRef{}, fmt.Errorf("S9 issued RepairResult packet %q is unavailable: %v", req.PacketID, err)
	}
	if req.OwnerAgentID == "" || req.OwnerAgentID != packet.Binding.Sources.OwnerAgentID {
		return mainSnapshot, WorkspaceCandidate{}, ArtifactRef{}, errors.New("S9 Worker owner binding differs from the issued RepairAssignment")
	}
	if req.AnswerPath != packet.AnswerPath {
		return mainSnapshot, WorkspaceCandidate{}, ArtifactRef{}, errors.New("S9 answer path differs from the exact issued packet")
	}
	if req.AuthorizeSource != nil {
		if err := req.AuthorizeSource(root, mainSnapshot.State); err != nil {
			return mainSnapshot, WorkspaceCandidate{}, ArtifactRef{}, err
		}
	}
	accepted, acceptedOK, err := runtimepkg.VerifiedRepairResultCandidateReceipt(root, mainSnapshot.State, packet.ID, req.OwnerAgentID)
	if err != nil {
		return mainSnapshot, WorkspaceCandidate{}, ArtifactRef{}, err
	}
	var completionReport runtimepkg.RepairResultCompletionReport
	var registry *workspace.ExecutionRegistry
	var execution workspace.Execution
	if acceptedOK {
		// An accepted answer is historical evidence. Do not require the current
		// Agent/TASK pointers to remain in `reported`; the receipt reader already
		// revalidates this frozen completion tuple against the original event and
		// source bytes. An optional integration callback, if supplied, ran above.
		completionReport = runtimepkg.RepairResultCompletionReport{
			Ref: accepted.Provenance.CompletionReportRef, SHA256: accepted.Provenance.CompletionReportSHA256,
			EventID: accepted.Provenance.CompletionReportEventID,
		}
	}
	actualSource := ""
	if acceptedOK {
		actualSource = accepted.Provenance.ExecutionRoot
	} else {
		registry, execution, err = currentRepairExecution(ctx, root, mainSnapshot.State, packet, false)
		if err != nil {
			return mainSnapshot, WorkspaceCandidate{}, ArtifactRef{}, err
		}
		actualSource, err = workspace.Canonical(execution.Path)
		if err != nil || actualSource != execution.Path {
			return mainSnapshot, WorkspaceCandidate{}, ArtifactRef{}, errors.New("S9 RepairResult authoring Worker path is not its canonical registered root")
		}
	}
	if req.SourceRoot == "" || !sameAbsolutePath(req.SourceRoot, actualSource) {
		return mainSnapshot, WorkspaceCandidate{}, ArtifactRef{}, errors.New("S9 answer source root differs from the registered Worker execution")
	}
	answerBytes, err := workpacket.ReadFile(actualSource, packet.AnswerPath, repairResultAnswerLimit)
	if err != nil {
		return mainSnapshot, WorkspaceCandidate{}, ArtifactRef{}, err
	}
	answer, problems := workpacket.ValidateRepairResultAnswer(packet, answerBytes)
	if len(problems) != 0 {
		return mainSnapshot, WorkspaceCandidate{}, ArtifactRef{}, fmt.Errorf("S9 answer %s: %s (%s)", problems[0].Pointer, problems[0].Expected, problems[0].Observed)
	}
	answerSHA := workpacket.Digest(answer)
	if acceptedOK {
		if err := runtimepkg.VerifyRepairResultPacketIssue(root, mainSnapshot.State, packet); err != nil {
			return mainSnapshot, WorkspaceCandidate{}, ArtifactRef{}, err
		}
		if accepted.Provenance.AnswerSHA256 != answerSHA || accepted.Provenance.SubmittedAnswer == nil || string(accepted.Provenance.SubmittedAnswer) != string(answer) || !sameAbsolutePath(accepted.Provenance.ExecutionRoot, actualSource) ||
			!sameRepairResultCompletionReport(completionReport, runtimepkg.RepairResultCompletionReport{Ref: accepted.Provenance.CompletionReportRef, SHA256: accepted.Provenance.CompletionReportSHA256, EventID: accepted.Provenance.CompletionReportEventID}) {
			return mainSnapshot, WorkspaceCandidate{}, ArtifactRef{}, runtimepkg.ErrSubmissionConflict
		}
		head, headErr := workspace.Git(ctx, actualSource, "rev-parse", "HEAD")
		if headErr != nil {
			return mainSnapshot, WorkspaceCandidate{}, ArtifactRef{}, headErr
		}
		if head != accepted.Provenance.SourceHead {
			return mainSnapshot, WorkspaceCandidate{}, ArtifactRef{}, fmt.Errorf("S9 exact replay source HEAD differs from its accepted candidate; inspect the receipt or use explicit rework")
		}
		ref := ArtifactRef{Path: accepted.CandidateRef, SHA256: accepted.CandidateSHA256}
		candidate, err := ReadWorkspaceCandidate(root, ref)
		if err != nil {
			return mainSnapshot, WorkspaceCandidate{}, ArtifactRef{}, err
		}
		return mainSnapshot, candidate, ref, nil
	}
	if err := checkRevision(req.ExpectedRevision, mainSnapshot.Revision, "runtime S9 issued RepairResult answer"); err != nil {
		return mainSnapshot, WorkspaceCandidate{}, ArtifactRef{}, err
	}
	current, err := writer.Snapshot()
	if err != nil {
		return mainSnapshot, WorkspaceCandidate{}, ArtifactRef{}, err
	}
	if current.Revision != mainSnapshot.Revision {
		return current, WorkspaceCandidate{}, ArtifactRef{}, runtimepkg.ErrStaleRevision
	}
	if !reflect.DeepEqual(current.State, mainSnapshot.State) {
		return current, WorkspaceCandidate{}, ArtifactRef{}, runtimepkg.ErrStaleRevision
	}
	currentPacket, ok, err := runtimepkg.FindRepairResultPacket(current.State, packet.ID)
	if err != nil || !ok || !sameRepairResultPacket(currentPacket, packet) {
		return current, WorkspaceCandidate{}, ArtifactRef{}, errors.New("S9 packet changed before Worker candidate capture")
	}
	if !sameAbsolutePath(execution.Path, actualSource) {
		return current, WorkspaceCandidate{}, ArtifactRef{}, errors.New("S9 Worker source root changed before candidate capture")
	}
	if req.AuthorizeSource != nil {
		if err := req.AuthorizeSource(root, current.State); err != nil {
			return current, WorkspaceCandidate{}, ArtifactRef{}, err
		}
	}
	// Verify the strict issuance source closure and live current slot/attempt
	// against one request-local packet registry and Resume verifier.
	if err := runtimepkg.VerifyRepairResultPacketIssueAndCurrent(root, current.State, currentPacket); err != nil {
		return current, WorkspaceCandidate{}, ArtifactRef{}, err
	}
	answerBytes, err = workpacket.ReadFile(actualSource, packet.AnswerPath, repairResultAnswerLimit)
	if err != nil {
		return current, WorkspaceCandidate{}, ArtifactRef{}, err
	}
	answerNow, problems := workpacket.ValidateRepairResultAnswer(packet, answerBytes)
	if len(problems) != 0 || workpacket.Digest(answerNow) != answerSHA || string(answerNow) != string(answer) {
		return current, WorkspaceCandidate{}, ArtifactRef{}, runtimepkg.ErrSubmissionConflict
	}
	if err := ensureRepairResultCompleted(answer); err != nil {
		return current, WorkspaceCandidate{}, ArtifactRef{}, err
	}
	if !sameAbsolutePath(execution.Path, actualSource) {
		return current, WorkspaceCandidate{}, ArtifactRef{}, errors.New("S9 Worker answer source does not match its current registered execution")
	}
	if err := verifyRepairResultExecutionChecks(ctx, registry, execution); err != nil {
		return current, WorkspaceCandidate{}, ArtifactRef{}, err
	}
	// Run the original candidate domain checks before writing the generic
	// completion checkpoint. The preview is not published; CreateWorkspaceCandidate
	// repeats the same checks after the original completion consumer succeeds.
	previewProvenance, err := newRepairResultCandidateProvenance(ctx, root, current, packet, execution, answer, answerSHA, completionReport)
	if err != nil {
		return current, WorkspaceCandidate{}, ArtifactRef{}, err
	}
	previewProposal, err := repairResultProposal(packet, answer, execution, registry.MainRoot, previewProvenance)
	if err != nil {
		return current, WorkspaceCandidate{}, ArtifactRef{}, err
	}
	previewProposal.issued = &issuedRepairResultCapability{packet: packet, provenance: previewProvenance, beforeRevision: current.Revision}
	preview, err := prepareWorkspaceCandidate(ctx, root, current.State, execution, previewProposal)
	if err != nil {
		return current, WorkspaceCandidate{}, ArtifactRef{}, err
	}
	completionReport, err = runtimepkg.CurrentRepairResultCompletionReport(root, current.State, currentPacket)
	if err != nil {
		if req.EnsureCompletion == nil || repairResultCompletionPointerPresent(current.State, req.OwnerAgentID) {
			return current, WorkspaceCandidate{}, ArtifactRef{}, fmt.Errorf("S9 original completion report/PLAN prerequisite: %w", err)
		}
		if workingErr := agentauth.ValidateWorkingAuthorization(root, current.State, req.OwnerAgentID); workingErr != nil {
			return current, WorkspaceCandidate{}, ArtifactRef{}, fmt.Errorf("S9 original completion report/PLAN prerequisite: %w", err)
		}
		completionReport, err = req.EnsureCompletion(ctx, current, packet, execution, preview.Proposal)
		if err != nil {
			return current, WorkspaceCandidate{}, ArtifactRef{}, fmt.Errorf("S9 automatic original completion: %w", err)
		}
		current, err = writer.Snapshot()
		if err != nil {
			return runtimepkg.Snapshot{}, WorkspaceCandidate{}, ArtifactRef{}, err
		}
		mainSnapshot = current
		currentPacket, ok, err = runtimepkg.FindRepairResultPacket(current.State, packet.ID)
		if err != nil || !ok || !sameRepairResultPacket(currentPacket, packet) {
			return current, WorkspaceCandidate{}, ArtifactRef{}, errors.New("S9 packet changed after automatic completion")
		}
		if req.AuthorizeSource != nil {
			if err := req.AuthorizeSource(root, current.State); err != nil {
				return current, WorkspaceCandidate{}, ArtifactRef{}, err
			}
		}
		if err := runtimepkg.VerifyRepairResultPacketIssueAndCurrent(root, current.State, currentPacket); err != nil {
			return current, WorkspaceCandidate{}, ArtifactRef{}, err
		}
		freshRegistry, freshExecution, err := currentRepairExecution(ctx, root, current.State, currentPacket, false)
		if err != nil || !reflect.DeepEqual(freshExecution, execution) {
			return current, WorkspaceCandidate{}, ArtifactRef{}, fmt.Errorf("S9 Worker execution changed after automatic completion: %v", err)
		}
		registry, execution = freshRegistry, freshExecution
		if err := verifyRepairResultExecutionChecks(ctx, registry, execution); err != nil {
			return current, WorkspaceCandidate{}, ArtifactRef{}, err
		}
		answerBytes, err = workpacket.ReadFile(actualSource, packet.AnswerPath, repairResultAnswerLimit)
		if err != nil {
			return current, WorkspaceCandidate{}, ArtifactRef{}, err
		}
		answerNow, problems := workpacket.ValidateRepairResultAnswer(packet, answerBytes)
		if len(problems) != 0 || workpacket.Digest(answerNow) != answerSHA || string(answerNow) != string(answer) {
			return current, WorkspaceCandidate{}, ArtifactRef{}, runtimepkg.ErrSubmissionConflict
		}
		reportedCompletion, reportErr := runtimepkg.CurrentRepairResultCompletionReport(root, current.State, currentPacket)
		if reportErr != nil || !sameRepairResultCompletionReport(reportedCompletion, completionReport) {
			return current, WorkspaceCandidate{}, ArtifactRef{}, fmt.Errorf("S9 automatic original completion did not retain the exact current report: %v", reportErr)
		}
		completionReport = reportedCompletion
	}
	if err := agentauth.ValidateReportedPlanAuthorization(root, current.State, req.OwnerAgentID); err != nil {
		return current, WorkspaceCandidate{}, ArtifactRef{}, err
	}
	currentCompletion, err := runtimepkg.CurrentRepairResultCompletionReport(root, current.State, currentPacket)
	if err != nil || !sameRepairResultCompletionReport(currentCompletion, completionReport) {
		return current, WorkspaceCandidate{}, ArtifactRef{}, fmt.Errorf("S9 original completion report changed before candidate capture: %v", err)
	}
	provenance, err := newRepairResultCandidateProvenance(ctx, root, current, packet, execution, answer, answerSHA, completionReport)
	if err != nil {
		return current, WorkspaceCandidate{}, ArtifactRef{}, err
	}
	proposal, err := repairResultProposal(packet, answer, execution, registry.MainRoot, provenance)
	if err != nil {
		return current, WorkspaceCandidate{}, ArtifactRef{}, err
	}
	proposal.issued = &issuedRepairResultCapability{packet: packet, provenance: provenance, beforeRevision: current.Revision}
	candidate, ref, err := CreateWorkspaceCandidate(ctx, root, current.State, execution, proposal)
	if err != nil {
		return current, WorkspaceCandidate{}, ArtifactRef{}, err
	}
	if candidate.Provenance == nil || !reflect.DeepEqual(*candidate.Provenance, provenance) {
		return current, WorkspaceCandidate{}, ArtifactRef{}, errors.New("original WorkspaceCandidate writer did not preserve typed answer provenance")
	}
	key := "runtime:s9:issued-repair-candidate:" + packet.ID
	eventID := "evt-s9-issued-repair-candidate-" + packet.ID
	evidenceIDs := []string{ref.Path, packet.ID, packet.Binding.Sources.RepairAssignmentID}
	at := occurred(req.OccurredAt)
	mutation := runtimepkg.Mutation{
		EventID: eventID, TransitionID: "WORKSPACE", Event: "workspace_updated", Actor: req.OwnerAgentID,
		IdempotencyKey: key, RuntimeID: packet.Binding.Sources.RuntimeID, EvidenceIDs: evidenceIDs,
		From: cursor(current.State), To: cursor(current.State), BaselineGeneration: packet.Binding.Sources.BaselineGeneration,
		RetainLastTransition: true, OccurredAt: at, Message: "accepted issued S9 professional result as an immutable workspace candidate; Main integration remains pending",
		Apply: func(state map[string]any) error {
			if err := runtimepkg.VerifyRepairResultPacketIssueAndCurrent(root, state, packet); err != nil {
				return err
			}
			if req.AuthorizeSource != nil {
				if err := req.AuthorizeSource(root, state); err != nil {
					return err
				}
			}
			if err := agentauth.ValidateReportedPlanAuthorization(root, state, req.OwnerAgentID); err != nil {
				return err
			}
			appliedCompletion, err := runtimepkg.CurrentRepairResultCompletionReport(root, state, packet)
			if err != nil || !sameRepairResultCompletionReport(appliedCompletion, completionReport) {
				return fmt.Errorf("S9 original completion report changed before candidate Apply: %v", err)
			}
			if err := validateRepairResultMainThreadlessCandidate(root, state, packet, execution, provenance, ref); err != nil {
				return err
			}
			freshRegistry, freshExecution, err := currentRepairExecution(ctx, root, state, packet, false)
			if err != nil {
				return err
			}
			if !reflect.DeepEqual(freshExecution, execution) {
				return errors.New("S9 registered Worker execution changed before candidate CAS")
			}
			if err := verifyRepairResultExecutionChecks(ctx, freshRegistry, freshExecution); err != nil {
				return err
			}
			actualAnswer, err := workpacket.ReadFile(actualSource, packet.AnswerPath, repairResultAnswerLimit)
			if err != nil {
				return err
			}
			acceptedAnswer, problems := workpacket.ValidateRepairResultAnswer(packet, actualAnswer)
			if len(problems) != 0 || workpacket.Digest(acceptedAnswer) != answerSHA || string(acceptedAnswer) != string(answer) {
				return runtimepkg.ErrSubmissionConflict
			}
			workspaceRegistry, err := workspace.Decode(state)
			if err != nil || workspaceRegistry == nil {
				return fmt.Errorf("S9 Workspace registry disappeared before candidate CAS: %v", err)
			}
			entry := workspaceRegistry.Executions[freshExecution.AssignmentID]
			if entry.AssignmentID != freshExecution.AssignmentID || entry.DeliveryRef != "" && (entry.DeliveryRef != ref.Path || entry.DeliverySHA256 != ref.SHA256) {
				return errors.New("S9 Worker already has a conflicting delivery candidate")
			}
			if entry.DeliveryRef == ref.Path && entry.DeliverySHA256 == ref.SHA256 {
				return errors.New("identical S9 candidate appeared during receipt; recover its original WORKSPACE event")
			}
			entry.DeliveryRef, entry.DeliverySHA256 = ref.Path, ref.SHA256
			workspaceRegistry.Executions[freshExecution.AssignmentID] = entry
			state["workspace"] = workspace.Encode(workspaceRegistry)
			return nil
		},
	}
	_, err = writer.Update(current.Revision, mutation)
	if err != nil {
		if recovered, recoveredSnapshot, recoveredCandidate, recoveredRef, recoverErr := recoverAcceptedRepairResultCandidate(root, writer, packet, req, answerSHA, actualSource); recoverErr == nil && recovered {
			return recoveredSnapshot, recoveredCandidate, recoveredRef, nil
		}
		return current, WorkspaceCandidate{}, ArtifactRef{}, err
	}
	next, err := writer.Snapshot()
	if err != nil {
		return runtimepkg.Snapshot{}, WorkspaceCandidate{}, ArtifactRef{}, err
	}
	acceptedReceipt, acceptedAtCommit, err := runtimepkg.VerifiedRepairResultCandidateReceipt(root, next.State, packet.ID, req.OwnerAgentID)
	if err != nil || !acceptedAtCommit || acceptedReceipt.CandidateRef != ref.Path || acceptedReceipt.CandidateSHA256 != ref.SHA256 || !reflect.DeepEqual(acceptedReceipt.Provenance, provenance) {
		return next, candidate, ref, fmt.Errorf("original S9 Workspace CAS did not retain the exact candidate receipt: %v", err)
	}
	return next, candidate, ref, nil
}

func ensureRepairResultCompleted(raw json.RawMessage) error {
	var answer workpacket.RepairResultProfessionalAnswer
	if err := json.Unmarshal(raw, &answer); err != nil {
		return err
	}
	if answer.Verdict != "completed" {
		return fmt.Errorf("S9 typed result verdict %q is preserved for named recovery; this slice only accepts completed candidate delivery", answer.Verdict)
	}
	if len(answer.ScopeDeviations) != 0 {
		return errors.New("S9 completed result declares scope deviations; preserve it and route the scope change through approved recovery")
	}
	for _, unit := range answer.UnitResults {
		if unit.Status != "pass" {
			return errors.New("S9 completed result has a non-passing unit; preserve it for named recovery")
		}
	}
	return nil
}

func currentRepairExecution(ctx context.Context, root string, state map[string]any, packet workpacket.RepairResultPacket, allowDelivered bool) (*workspace.ExecutionRegistry, workspace.Execution, error) {
	b := packet.Binding.Sources
	registry, err := workspace.Decode(state)
	if err != nil || registry == nil {
		return nil, workspace.Execution{}, fmt.Errorf("S9 typed result requires original Workspace execution registry: %v", err)
	}
	var selected *workspace.Execution
	for key, execution := range registry.Executions {
		if key != execution.AssignmentID || (execution.AssignmentID != b.RepairAssignmentID && execution.AssignmentID != b.ManifestAssignmentID) || execution.RuntimeID != b.RuntimeID || execution.BaselineGeneration != b.BaselineGeneration || execution.AgentID != b.OwnerAgentID {
			continue
		}
		if selected != nil {
			return nil, workspace.Execution{}, errors.New("S9 packet resolves to multiple current canonical/manifest Worker executions")
		}
		copy := execution
		selected = &copy
	}
	if selected == nil {
		return nil, workspace.Execution{}, errors.New("S9 packet has no exact current registered Worker execution")
	}
	if selected.Status != "ready" || (!allowDelivered && selected.DeliveryRef != "") || selected.BootstrapSHA256 == "" || selected.ReworkRef != packet.Binding.ReworkRef {
		return nil, workspace.Execution{}, errors.New("S9 candidate requires the current ready execution and attempt")
	}
	// workspace.Decode already runs ValidateAuthority against this exact state.
	// Keep the explicit execution/input checks here without re-reading every
	// registered Assignment source a second time.
	if err := registry.ValidateExecution(ctx, *selected, false); err != nil {
		return nil, workspace.Execution{}, err
	}
	if err := registry.ValidateInputs(*selected); err != nil {
		return nil, workspace.Execution{}, err
	}
	return registry, *selected, nil
}

func verifyRepairResultExecutionChecks(ctx context.Context, registry *workspace.ExecutionRegistry, execution workspace.Execution) error {
	if registry == nil {
		return errors.New("S9 check receipt verification requires the original Workspace registry")
	}
	if execution.CheckLocation != "worker" && execution.CheckLocation != "main" {
		return errors.New("S9 execution has no original frozen check location")
	}
	if execution.CheckLocation == "main" && len(execution.ToolCheckReceipts) != 0 {
		return errors.New("S9 Main-owned required checks cannot be represented as Worker receipts")
	}
	if execution.CheckLocation == "worker" {
		for index, command := range execution.Checks {
			if strings.HasPrefix(strings.TrimSpace(command), "locked:") {
				continue
			}
			profile, ok := workspace.ProfileAt(execution, index)
			if !ok || profile.Command != command {
				return fmt.Errorf("S9 Worker required check %d lacks its original frozen profile", index)
			}
			ref, ok := latestRepairResultToolReceipt(execution.ToolCheckReceipts, profile)
			if !ok {
				return fmt.Errorf("S9 Worker required check %d has no current original receipt", index)
			}
			if err := verifyRepairResultToolReceipt(ctx, registry.MainRoot, execution, profile, ref, false); err != nil {
				return err
			}
		}
	}
	for index, check := range execution.WorkerChecks {
		ref, ok := latestRepairResultOwnedReceipt(execution.OwnedCheckReceipts, index)
		if !ok {
			return fmt.Errorf("S9 Worker-owned required check %d has no original receipt", index)
		}
		if _, err := workspace.VerifyOwnedCheckReceipt(filepath.Join(registry.MainRoot, ".claude/evidence/workspace-checks"), execution, index, ref); err != nil {
			return err
		}
		verified, err := checkexec.VerifyToolReceipt(filepath.Join(registry.MainRoot, ".claude/evidence/workspace-checks"), ref)
		if err != nil || !verified.Passed {
			return fmt.Errorf("S9 Worker-owned check %d receipt is not passing: %v", index, err)
		}
		_, observed, err := workspace.ObserveOwnedCheckSource(ctx, execution.Path, check.Profile)
		if err != nil || !sameRepairResultObservation(verified.Observation, observed) {
			return fmt.Errorf("S9 Worker-owned check %d source no longer matches its passed receipt: %v", index, err)
		}
	}
	return nil
}

func latestRepairResultToolReceipt(refs []checkexec.ToolReceipt, profile checkexec.Profile) (checkexec.ToolReceipt, bool) {
	for index := len(refs) - 1; index >= 0; index-- {
		if refs[index].Command == profile.Command && refs[index].ProfileDigest == profile.Digest() {
			return refs[index], true
		}
	}
	return checkexec.ToolReceipt{}, false
}

func latestRepairResultOwnedReceipt(refs []workspace.OwnedCheckReceipt, index int) (checkexec.ToolReceipt, bool) {
	for row := len(refs) - 1; row >= 0; row-- {
		if refs[row].WorkerCheckIndex == index {
			return refs[row].Receipt, true
		}
	}
	return checkexec.ToolReceipt{}, false
}

func verifyRepairResultToolReceipt(ctx context.Context, evidenceRoot string, execution workspace.Execution, profile checkexec.Profile, ref checkexec.ToolReceipt, main bool) error {
	responsibility := "owned_check"
	if main {
		responsibility = "main_required_integration"
	}
	expected := checkexec.Binding{RuntimeID: execution.RuntimeID, Generation: execution.BaselineGeneration, AssignmentID: execution.AssignmentID, Owner: execution.AgentID, ExecutionGeneration: execution.Generation, Responsibility: responsibility}
	if ref.Binding != expected || ref.Checkout != execution.Path || ref.Command != profile.Command || ref.ProfileDigest != profile.Digest() {
		return errors.New("S9 check receipt differs from its exact current execution/profile binding")
	}
	verified, err := checkexec.VerifyToolReceipt(evidenceRoot, ref)
	if err != nil || !verified.Passed {
		return fmt.Errorf("S9 check receipt did not pass original verification: %v", err)
	}
	_, observed, err := workspace.ObserveOwnedCheckSource(ctx, execution.Path, profile)
	if err != nil || !sameRepairResultObservation(verified.Observation, observed) {
		return fmt.Errorf("S9 Worker check source no longer matches its received receipt: %v", err)
	}
	return nil
}

func sameRepairResultObservation(a, b checkexec.Observation) bool {
	return a.ProfileID == b.ProfileID && a.ProfileVersion == b.ProfileVersion && a.ProfileDigest == b.ProfileDigest && a.SubjectDigest == b.SubjectDigest && a.RunnerDigest == b.RunnerDigest && a.EnvironmentDigest == b.EnvironmentDigest && a.DependencyDigest == b.DependencyDigest && a.FixtureDigest == b.FixtureDigest && a.VerificationScope == b.VerificationScope && reflect.DeepEqual(a.Environment, b.Environment)
}

func newRepairResultCandidateProvenance(ctx context.Context, root string, snapshot runtimepkg.Snapshot, packet workpacket.RepairResultPacket, execution workspace.Execution, answer json.RawMessage, answerSHA string, completionReport runtimepkg.RepairResultCompletionReport) (runtimepkg.RepairResultCandidateProvenance, error) {
	packetSHA, err := workpacket.RepairResultPacketDigest(packet)
	if err != nil {
		return runtimepkg.RepairResultCandidateProvenance{}, err
	}
	head, err := workspace.Git(ctx, execution.Path, "rev-parse", "HEAD")
	if err != nil || head == "" {
		return runtimepkg.RepairResultCandidateProvenance{}, fmt.Errorf("read current S9 Worker HEAD: %v", err)
	}
	jBytes, err := controlproof.ReadJournalForState(root, snapshot.State)
	if err != nil {
		return runtimepkg.RepairResultCandidateProvenance{}, err
	}
	j, err := controlproof.InspectJournalData(jBytes)
	if err != nil || controlproof.ValidateStateJournalPair(snapshot.State, j) != nil {
		return runtimepkg.RepairResultCandidateProvenance{}, fmt.Errorf("inspect S9 Runtime journal before candidate receipt: %v", err)
	}
	if snapshot.Revision < 0 || j.TailSequence < 0 {
		return runtimepkg.RepairResultCandidateProvenance{}, errors.New("S9 candidate receipt cannot predict invalid original Runtime coordinates")
	}
	workerReceipts, ownedReceipts, err := repairResultOwnReceipts(execution)
	if err != nil {
		return runtimepkg.RepairResultCandidateProvenance{}, err
	}
	b := packet.Binding
	resumeSHA, err := workpacket.RepairResultResumeSHA256(b.Resume)
	if err != nil {
		return runtimepkg.RepairResultCandidateProvenance{}, fmt.Errorf("digest S9 Resume predecessor tuple: %w", err)
	}
	return runtimepkg.RepairResultCandidateProvenance{
		Contract: repairResultCandidateContract, PacketID: packet.ID, PacketSHA256: packetSHA,
		IssuanceEventID: packet.IssuanceEventID, AnswerSHA256: answerSHA, SubmittedAnswer: append(json.RawMessage(nil), answer...),
		OwnerAgentID: b.Sources.OwnerAgentID, RuntimeID: b.Sources.RuntimeID, BaselineGeneration: b.Sources.BaselineGeneration,
		ReviewRound: b.Sources.ReviewRound, ExecutionAssignmentID: execution.AssignmentID, ExecutionGeneration: execution.Generation,
		ExecutionSHA256: b.ExecutionSHA256, ExecutionRoot: execution.Path, SourceHead: head, BaseCommit: execution.BaseCommit,
		PlanReportRef: b.PlanReportRef, PlanReportSHA256: b.PlanReportSHA256, ReworkRef: b.ReworkRef, ReworkSHA256: b.ReworkSHA256,
		Resume: b.Resume, ResumeSHA256: resumeSHA,
		CompletionReportRef: completionReport.Ref, CompletionReportSHA256: completionReport.SHA256, CompletionReportEventID: completionReport.EventID,
		WorkerCheckReceipts: workerReceipts, OwnedCheckReceipts: ownedReceipts,
	}, nil
}

func sameRepairResultCompletionReport(left, right runtimepkg.RepairResultCompletionReport) bool {
	return left.Ref != "" && left.SHA256 != "" && left.EventID != "" && left == right
}

func repairResultCompletionPointerPresent(state map[string]any, owner string) bool {
	entities, _ := state["entities"].(map[string]any)
	rows, _ := entities["agents"].([]any)
	for _, raw := range rows {
		agent, _ := raw.(map[string]any)
		if stringField(agent["id"]) == owner && stringField(agent["completion_reported_ref"]) != "" {
			return true
		}
	}
	return false
}

// repairResultOwnReceipts projects receipts already checked by
// verifyRepairResultExecutionChecks in the same preflight. The lifecycle CAS
// reopens and verifies them again against its current source tree.
func repairResultOwnReceipts(execution workspace.Execution) ([]checkexec.ToolReceipt, []runtimepkg.RepairResultOwnedCheckReceipt, error) {
	worker := []checkexec.ToolReceipt{}
	owned := []runtimepkg.RepairResultOwnedCheckReceipt{}
	if execution.CheckLocation == "worker" {
		for index, command := range execution.Checks {
			if strings.HasPrefix(strings.TrimSpace(command), "locked:") {
				continue
			}
			profile, ok := workspace.ProfileAt(execution, index)
			if !ok || profile.Command != command {
				return nil, nil, fmt.Errorf("S9 Worker check %d lacks frozen profile", index)
			}
			ref, ok := latestRepairResultToolReceipt(execution.ToolCheckReceipts, profile)
			if !ok {
				return nil, nil, fmt.Errorf("S9 Worker check %d lacks receipt", index)
			}
			worker = append(worker, ref)
		}
	}
	for index := range execution.WorkerChecks {
		ref, ok := latestRepairResultOwnedReceipt(execution.OwnedCheckReceipts, index)
		if !ok {
			return nil, nil, fmt.Errorf("S9 Worker-owned check %d lacks receipt", index)
		}
		owned = append(owned, runtimepkg.RepairResultOwnedCheckReceipt{WorkerCheckIndex: index, Receipt: ref})
	}
	sort.Slice(worker, func(i, j int) bool { return worker[i].ProfileDigest < worker[j].ProfileDigest })
	sort.Slice(owned, func(i, j int) bool { return owned[i].WorkerCheckIndex < owned[j].WorkerCheckIndex })
	return worker, owned, nil
}

func repairResultProposal(packet workpacket.RepairResultPacket, raw json.RawMessage, execution workspace.Execution, mainRoot string, provenance runtimepkg.RepairResultCandidateProvenance) (RepairResultRequest, error) {
	var answer workpacket.RepairResultProfessionalAnswer
	if err := json.Unmarshal(raw, &answer); err != nil {
		return RepairResultRequest{}, err
	}
	units := make([]RepairUnitResult, len(answer.UnitResults))
	for i, unit := range answer.UnitResults {
		units[i] = RepairUnitResult{UnitID: unit.UnitID, Status: unit.Status, EvidenceRefs: append([]string{}, unit.EvidenceRefs...), ImplementationSummary: unit.ImplementationSummary, RestoredInvariantRefs: append([]string{}, unit.RestoredInvariantRefs...)}
	}
	before := make([]RepairBeforeFixResult, len(answer.BeforeFixResults))
	for i, result := range answer.BeforeFixResults {
		before[i] = RepairBeforeFixResult{AssertionID: result.AssertionID, ObservedRed: result.ObservedRed, EvidenceRefs: append([]string{}, result.EvidenceRefs...)}
	}
	checks := []RepairCheck{}
	if execution.CheckLocation == "worker" {
		for _, ref := range provenance.WorkerCheckReceipts {
			relative, err := filepath.Rel(mainRoot, ref.Path)
			if err != nil || !filepath.IsLocal(relative) {
				return RepairResultRequest{}, fmt.Errorf("S9 Worker check receipt escaped the original Main evidence root: %v", err)
			}
			evidence := "path:" + filepath.ToSlash(relative) + "#sha256=" + ref.SHA256
			checks = append(checks, RepairCheck{Name: ref.Command, Command: ref.Command, Result: "pass", EvidenceRefs: []string{evidence}})
		}
	}
	return RepairResultRequest{
		Contract: ContractRef{Path: packet.Binding.Sources.ContractRef, SHA256: packet.Binding.Sources.ContractSHA256},
		Session:  ArtifactRef{ID: packet.Binding.Sources.SessionID, Path: packet.Binding.Sources.SessionRef, SHA256: packet.Binding.Sources.SessionSHA256},
		Plan:     ArtifactRef{ID: packet.Binding.Sources.PlanID, Path: packet.Binding.Sources.PlanRef, SHA256: packet.Binding.Sources.PlanSHA256},
		ResultID: packet.ResultID, ProducerAgentID: packet.Binding.Sources.OwnerAgentID, UnitResults: units,
		Verdict: answer.Verdict, BeforeFixResults: before, DetectionGapArtifacts: append([]string{}, answer.DetectionGapArtifacts...), EvidenceRefs: append([]string{}, answer.EvidenceRefs...),
		ChangedArtifacts: []ChangedArtifact{}, Result: "pass", AssignmentID: packet.Binding.Sources.RepairAssignmentID,
		PlanReport:      ArtifactRef{Path: packet.Binding.PlanReportRef, SHA256: packet.Binding.PlanReportSHA256},
		BeforeFixChecks: []RepairCheck{}, Checks: checks, ScopeDeviations: append([]string{}, answer.ScopeDeviations...),
		MigrationRef: answer.MigrationRef, RollbackRef: answer.RollbackRef, ResidualRisks: append([]string{}, answer.ResidualRisks...),
	}, nil
}

func validateRepairResultMainThreadlessCandidate(root string, state map[string]any, packet workpacket.RepairResultPacket, execution workspace.Execution, provenance runtimepkg.RepairResultCandidateProvenance, candidate ArtifactRef) error {
	if candidate.Path == "" || candidate.SHA256 == "" || candidate.SHA256 != sha256PathBase(candidate.Path) {
		return errors.New("S9 candidate must use its exact content-addressed original path")
	}
	if provenance.PacketID != packet.ID || provenance.OwnerAgentID != packet.Binding.Sources.OwnerAgentID || provenance.ExecutionRoot != execution.Path || provenance.ExecutionAssignmentID != execution.AssignmentID || provenance.SourceHead == "" {
		return errors.New("S9 candidate typed provenance differs from current original execution")
	}
	return nil
}

func sha256PathBase(path string) string {
	base := strings.TrimSuffix(filepath.Base(path), ".json")
	return base
}

func workspaceMainRoot(fallback string, state map[string]any) string {
	registry, err := workspace.Decode(state)
	if err == nil && registry != nil && registry.MainRoot != "" {
		return registry.MainRoot
	}
	return fallback
}

// guardIssuedRepairResultCandidate prevents the original public candidate
// constructor from bypassing a current typed packet. The capability is
// package-private and is installed only after the issuer has checked the
// answer, Worker, plan, source and own-check facts.
func guardIssuedRepairResultCandidate(_ string, state map[string]any, execution workspace.Execution, capability *issuedRepairResultCapability) error {
	packets, err := runtimepkg.RepairResultPackets(state)
	if err != nil {
		return err
	}
	baseline, _ := state["baseline"].(map[string]any)
	review, _ := state["review"].(map[string]any)
	pointer, _ := review["repair"].(map[string]any)
	life, _ := state["lifecycle"].(map[string]any)
	gen := integerRepairResult(baseline["generation"])
	round := integerRepairResult(review["round"])
	matched := []workpacket.RepairResultPacket{}
	for _, packet := range packets {
		b := packet.Binding.Sources
		var definition struct {
			AssignmentID string `json:"assignment_id"`
			RuntimeID    string `json:"runtime_id"`
			Generation   int    `json:"execution_generation"`
			AgentID      string `json:"agent_id"`
			Path         string `json:"worktree_path"`
		}
		if err := json.Unmarshal(packet.Inputs.Execution, &definition); err != nil {
			return err
		}
		if b.OwnerAgentID != execution.AgentID || definition.AssignmentID != execution.AssignmentID || definition.RuntimeID != execution.RuntimeID || definition.AgentID != execution.AgentID || definition.Generation != execution.Generation || !sameAbsolutePath(definition.Path, execution.Path) {
			continue
		}
		if b.RuntimeID != state["runtime_id"] || b.BaselineGeneration != gen || b.ReviewRound != round || pointer["session_id"] != b.SessionID || pointer["plan_ref"] != b.PlanRef || pointer["plan_sha256"] != b.PlanSHA256 || pointer["status"] != "repairing" || life["state"] != "bug_resolution" || life["phase"] != "fixing" {
			continue
		}
		matched = append(matched, packet)
	}
	if capability == nil {
		if len(matched) > 0 {
			return fmt.Errorf("S9 Assignment %s has a retained issued RepairResult packet; use its exact answer and original candidate receiver for the unique current packet", matched[0].Binding.Sources.RepairAssignmentID)
		}
		return nil
	}
	revision, hasRevision := state["revision"]
	if capability.beforeRevision < 0 || !hasRevision || integerRepairResult(revision) != capability.beforeRevision {
		return runtimepkg.ErrStaleRevision
	}
	var selected *workpacket.RepairResultPacket
	for i := range matched {
		// RepairResultPackets already validates and canonically rebuilds every
		// registered row; the private capability holds the canonical packet
		// produced by the same issuer. Compare that exact typed value without
		// rebuilding both digests a second time in this guard.
		if !reflect.DeepEqual(capability.packet, matched[i]) {
			continue
		}
		if selected != nil {
			return errors.New("S9 Runtime packet registry duplicates the private candidate packet")
		}
		copy := matched[i]
		selected = &copy
	}
	if selected == nil {
		return errors.New("S9 private candidate capability is not bound to the exact current issued successor packet")
	}
	resumeSHA, err := workpacket.RepairResultResumeSHA256(selected.Binding.Resume)
	if err != nil {
		return err
	}
	if capability.provenance.PacketID != selected.ID || capability.provenance.OwnerAgentID != execution.AgentID || capability.provenance.ExecutionAssignmentID != execution.AssignmentID || capability.provenance.ExecutionRoot != execution.Path || capability.provenance.ExecutionGeneration != execution.Generation || capability.provenance.RuntimeID != execution.RuntimeID || capability.provenance.BaselineGeneration != execution.BaselineGeneration || capability.provenance.ReworkRef != execution.ReworkRef || !reflect.DeepEqual(capability.provenance.Resume, selected.Binding.Resume) || capability.provenance.ResumeSHA256 != resumeSHA {
		return errors.New("S9 private candidate capability differs from its unique current packet, Resume predecessor, or Worker execution")
	}
	return nil
}

// rejectLegacyRepairResultForIssuedTarget prevents callers of the original
// untyped Result API from bypassing an issued S9 packet. Once any packet exists
// for the current Session/Plan/Assignment tuple, later attempts remain in the
// typed path even when a fresh packet selector would no longer return them.
func rejectLegacyRepairResultForIssuedTarget(state, pointer map[string]any, plan RepairPlan, planRef ArtifactRef, assignmentID string) error {
	packets, err := runtimepkg.RepairResultPackets(state)
	if err != nil {
		return err
	}
	review, _ := state["review"].(map[string]any)
	for _, packet := range packets {
		if repairResultPacketMatchesLegacyTarget(packet, stringField(state["runtime_id"]), baselineGeneration(state), integerRepairResult(review["round"]), stringField(pointer["session_id"]), plan, planRef, assignmentID) {
			return fmt.Errorf("S9 Assignment %s already has an issued typed RepairResult packet for this Session/Plan; use its exact answer and original candidate receiver", assignmentID)
		}
	}
	return nil
}

// repairResultPacketMatchesLegacyTarget resolves the canonical and registered
// manifest spellings of one Assignment before the original legacy Result
// writer decides whether its untyped path remains available.
func repairResultPacketMatchesLegacyTarget(packet workpacket.RepairResultPacket, runtimeID string, generation, round int, sessionID string, plan RepairPlan, planRef ArtifactRef, assignmentID string) bool {
	binding := packet.Binding.Sources
	assignmentMatches := binding.RepairAssignmentID == assignmentID || binding.ManifestAssignmentID == assignmentID || manifestAssignmentAlias(binding.RepairAssignmentID) == assignmentID
	return assignmentMatches && binding.RuntimeID == runtimeID && binding.BaselineGeneration == generation &&
		binding.ReviewRound == round && binding.SessionID == sessionID && binding.PlanID == plan.PlanID &&
		binding.PlanRef == planRef.Path && binding.PlanSHA256 == planRef.SHA256
}

func integerRepairResult(value any) int {
	switch n := value.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	case json.Number:
		v, _ := n.Int64()
		return int(v)
	default:
		return 0
	}
}

func recoverAcceptedRepairResultCandidate(root string, writer *runtimepkg.Store, packet workpacket.RepairResultPacket, req AcceptIssuedRepairResultCandidateRequest, answerSHA, sourceRoot string) (bool, runtimepkg.Snapshot, WorkspaceCandidate, ArtifactRef, error) {
	snapshot, err := writer.Snapshot()
	if err != nil {
		return false, runtimepkg.Snapshot{}, WorkspaceCandidate{}, ArtifactRef{}, err
	}
	receipt, ok, err := runtimepkg.VerifiedRepairResultCandidateReceipt(root, snapshot.State, packet.ID, req.OwnerAgentID)
	if err != nil || !ok {
		return false, snapshot, WorkspaceCandidate{}, ArtifactRef{}, err
	}
	if receipt.Provenance.AnswerSHA256 != answerSHA || !sameAbsolutePath(receipt.Provenance.ExecutionRoot, sourceRoot) {
		return false, snapshot, WorkspaceCandidate{}, ArtifactRef{}, runtimepkg.ErrSubmissionConflict
	}
	candidateRef := ArtifactRef{Path: receipt.CandidateRef, SHA256: receipt.CandidateSHA256}
	candidate, err := ReadWorkspaceCandidate(root, candidateRef)
	if err != nil {
		return false, snapshot, WorkspaceCandidate{}, ArtifactRef{}, err
	}
	return true, snapshot, candidate, candidateRef, nil
}

// RepairResultWorkComplete projects the registered Worker owner's current S9
// responsibility. An accepted Main Result remains sufficient historical work
// proof only while its exact delivery attempt is still current; before Main
// integration, the current candidate, clean source HEAD and frozen owner-run
// checks must agree. Main-owned checks remain consumer-pending.
func RepairResultWorkComplete(root string, state map[string]any, packetID, owner string) (bool, error) {
	return RepairResultWorkCompleteWithContext(context.Background(), root, state, packetID, owner)
}

// RepairResultWorkCompleteWithContext projects the registered Worker owner's
// current S9 responsibility while preserving a caller's reentrant workspace
// lock context. CLI delivery paths that already hold that lock must use this
// entry point instead of starting a second lock acquisition.
func RepairResultWorkCompleteWithContext(ctx context.Context, root string, state map[string]any, packetID, owner string) (bool, error) {
	packet, found, err := runtimepkg.FindRepairResultPacket(state, packetID)
	if err != nil || !found {
		return false, err
	}
	if owner == "" || owner != packet.Binding.Sources.OwnerAgentID {
		return false, errors.New("S9 work-complete owner differs from issued RepairAssignment")
	}
	if acceptedReceipt, accepted, err := runtimepkg.VerifiedRepairResultReceipt(root, state, packet.ID, owner); err != nil {
		return false, err
	} else if accepted {
		// A retained Main Result proves historical completion but must not
		// reactivate Worker responsibility after an explicit rework opened a
		// different execution attempt. Project work_complete only while the
		// accepted candidate still names the current registered attempt.
		return repairResultAttemptStillCurrent(state, acceptedReceipt.Provenance, acceptedReceipt.CandidateRef, acceptedReceipt.CandidateSHA256), nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, unlock, err := integration.LockWorkspace(ctx, workspaceMainRoot(root, state))
	if err != nil {
		return false, err
	}
	defer unlock()
	packets, err := runtimepkg.CurrentRepairResultPackets(root, state, owner)
	if err != nil {
		return false, err
	}
	current := false
	for _, candidate := range packets {
		if candidate.ID == packet.ID && sameRepairResultPacket(candidate, packet) {
			current = true
			break
		}
	}
	if !current {
		return false, nil
	}
	if err := runtimepkg.VerifyRepairResultPacketIssueAndCurrent(root, state, packet); err != nil {
		return false, err
	}
	receipt, accepted, err := runtimepkg.VerifiedRepairResultCandidateReceipt(root, state, packet.ID, owner)
	if err != nil || !accepted {
		return false, err
	}
	registry, execution, err := currentRepairExecution(ctx, root, state, packet, true)
	if err != nil {
		return false, err
	}
	if !sameAbsolutePath(execution.Path, receipt.Provenance.ExecutionRoot) || execution.AssignmentID != receipt.Provenance.ExecutionAssignmentID || execution.Generation != receipt.Provenance.ExecutionGeneration || execution.ReworkRef != receipt.Provenance.ReworkRef {
		return false, errors.New("S9 accepted candidate no longer binds the current Worker execution/attempt")
	}
	if head, err := workspace.Git(ctx, execution.Path, "rev-parse", "HEAD"); err != nil || head != receipt.Provenance.SourceHead {
		return false, errors.New("S9 accepted candidate source HEAD changed; perform named rework before reporting owner completion")
	}
	statusArgs := append([]string{"status", "--porcelain", "--untracked-files=all", "--", "."}, workspace.WorkerSubmissionStatusExcludes()...)
	status, err := workspace.Git(ctx, execution.Path, statusArgs...)
	if err != nil || status != "" {
		return false, errors.New("S9 accepted candidate Worker source is no longer clean")
	}
	if err := verifyRepairResultExecutionChecks(ctx, registry, execution); err != nil {
		return false, err
	}
	candidateRef := ArtifactRef{Path: receipt.CandidateRef, SHA256: receipt.CandidateSHA256}
	candidate, err := ValidateWorkspaceCandidate(root, state, execution, candidateRef)
	if err != nil || candidate.Provenance == nil || !reflect.DeepEqual(*candidate.Provenance, receipt.Provenance) {
		return false, fmt.Errorf("S9 original candidate delivery is not exact: %v", err)
	}
	return true, nil
}

func repairResultAttemptStillCurrent(state map[string]any, provenance runtimepkg.RepairResultCandidateProvenance, candidateRef, candidateSHA string) bool {
	registry, err := workspace.Decode(state)
	if err != nil || registry == nil {
		return false
	}
	var selected *workspace.Execution
	for _, execution := range registry.Executions {
		if execution.RuntimeID != provenance.RuntimeID || execution.BaselineGeneration != provenance.BaselineGeneration ||
			execution.AgentID != provenance.OwnerAgentID || execution.AssignmentID != provenance.ExecutionAssignmentID ||
			execution.Generation != provenance.ExecutionGeneration {
			continue
		}
		if selected != nil {
			return false
		}
		copy := execution
		selected = &copy
	}
	if selected == nil || (selected.Status != "ready" && selected.Status != "complete") || selected.ReworkRef != provenance.ReworkRef ||
		selected.DeliveryRef != candidateRef || selected.DeliverySHA256 != candidateSHA || !sameAbsolutePath(selected.Path, provenance.ExecutionRoot) {
		return false
	}
	encoded, err := json.Marshal(selected)
	if err != nil {
		return false
	}
	var currentDefinition map[string]any
	if json.Unmarshal(encoded, &currentDefinition) != nil {
		return false
	}
	projection, err := runtimepkg.ProjectRepairResultExecution(currentDefinition)
	return err == nil && workpacket.Digest(projection) == provenance.ExecutionSHA256
}

// SubmitIssuedRepairResult consumes the exact accepted Worker candidate only
// after the original S9 Main integration checkpoint has verified every
// frozen required profile. It serializes on the Main workspace lock before
// reading/mutating Runtime, and the original Result writer repeats the proof
// inside its Runtime CAS.
func SubmitIssuedRepairResult(ctx context.Context, root, statePath, journalPath string, req SubmitIssuedRepairResultRequest) (runtimepkg.Snapshot, RepairResult, ArtifactRef, error) {
	lockedCtx, unlock, err := integration.LockWorkspace(ctx, root)
	if err != nil {
		return runtimepkg.Snapshot{}, RepairResult{}, ArtifactRef{}, err
	}
	defer unlock()
	ctx = lockedCtx

	current, writer, err := readIssuedRepairResultRuntime(root, statePath, journalPath)
	if err != nil {
		return runtimepkg.Snapshot{}, RepairResult{}, ArtifactRef{}, err
	}
	if mainRoot := workspaceMainRoot(root, current.State); !sameAbsolutePath(mainRoot, root) {
		return current, RepairResult{}, ArtifactRef{}, errors.New("S9 final Result must be submitted by the registered Main checkout")
	}
	packet, found, err := runtimepkg.FindRepairResultPacket(current.State, req.PacketID)
	if err != nil || !found {
		return current, RepairResult{}, ArtifactRef{}, fmt.Errorf("S9 issued RepairResult packet %q is unavailable: %v", req.PacketID, err)
	}
	if req.OwnerAgentID == "" || req.OwnerAgentID != packet.Binding.Sources.OwnerAgentID {
		return current, RepairResult{}, ArtifactRef{}, errors.New("S9 final Result owner differs from the issued RepairAssignment")
	}
	if err := runtimepkg.VerifyRepairResultPacketIssue(root, current.State, packet); err != nil {
		return current, RepairResult{}, ArtifactRef{}, err
	}

	// A committed Result is queried before lifecycle, PLAN, expected revision,
	// or live Worker checks. It remains an immutable historical receipt after
	// normal acknowledgement/cleanup; a changed Main facts projection is not a
	// replay and must go through the existing rework path.
	accepted, acceptedOK, err := runtimepkg.VerifiedRepairResultReceipt(root, current.State, packet.ID, req.OwnerAgentID)
	if err != nil {
		return current, RepairResult{}, ArtifactRef{}, err
	}
	if acceptedOK {
		result, ref, ok, err := existingIssuedRepairResult(root, current.State, accepted.ResultID)
		if err != nil || !ok {
			return current, RepairResult{}, ArtifactRef{}, fmt.Errorf("S9 accepted Result artifact is unavailable: %v", err)
		}
		checksDoc, err := repairResultChecksDocument(result.Checks)
		if err != nil {
			return current, RepairResult{}, ArtifactRef{}, err
		}
		_, currentProof, err := readRepairResultMainProof(root, packet, accepted.Provenance, result.Checks)
		if err != nil {
			return current, RepairResult{}, ArtifactRef{}, fmt.Errorf("revalidate retained S9 Main facts for exact replay: %w", err)
		}
		if currentProof.CheckpointRef != accepted.MainProof.CheckpointRef || currentProof.VerifiedFactsSHA256 != accepted.MainProof.VerifiedFactsSHA256 || string(currentProof.VerifiedFacts) != string(accepted.MainProof.VerifiedFacts) {
			return current, RepairResult{}, ArtifactRef{}, errors.New("S9 exact replay Main verified facts changed; inspect the retained Result and use named rework")
		}
		if len(req.MainProof.VerifiedFacts) != 0 {
			requestFacts, factsErr := canonicalRuntimeFacts(req.MainProof.VerifiedFacts)
			if factsErr != nil || workpacket.Digest(requestFacts) != accepted.MainProof.VerifiedFactsSHA256 || string(requestFacts) != string(accepted.MainProof.VerifiedFacts) {
				return current, RepairResult{}, ArtifactRef{}, errors.New("S9 exact replay caller Main facts differ from the accepted receipt")
			}
		}
		if err := runtimepkg.VerifyRepairResultMainProof(root, packet, accepted.Provenance, accepted.MainProof, checksDoc); err != nil {
			return current, RepairResult{}, ArtifactRef{}, fmt.Errorf("revalidate original S9 Main proof: %w", err)
		}
		return current, result, ref, nil
	}

	if req.ExpectedRevision < 0 {
		return current, RepairResult{}, ArtifactRef{}, errors.New("fresh typed S9 Result commit requires an explicit Runtime expected revision")
	}
	if err := checkRevision(req.ExpectedRevision, current.Revision, "runtime issued S9 RepairResult"); err != nil {
		return current, RepairResult{}, ArtifactRef{}, err
	}
	if err := runtimepkg.VerifyRepairResultPacketCurrent(root, current.State, packet); err != nil {
		return current, RepairResult{}, ArtifactRef{}, err
	}
	issueRevision, issueAt, err := repairResultIssueCoordinates(root, current.State, packet)
	if err != nil {
		return current, RepairResult{}, ArtifactRef{}, err
	}
	if err := runtimepkg.VerifyRepairResultSources(root, current.State, packet, issueRevision); err != nil {
		return current, RepairResult{}, ArtifactRef{}, err
	}
	candidateReceipt, candidateOK, err := runtimepkg.VerifiedRepairResultCandidateReceipt(root, current.State, packet.ID, req.OwnerAgentID)
	if err != nil || !candidateOK {
		return current, RepairResult{}, ArtifactRef{}, fmt.Errorf("S9 Main integration requires the exact original accepted Worker candidate: %v", err)
	}
	registry, execution, err := currentRepairExecution(ctx, root, current.State, packet, true)
	if err != nil {
		return current, RepairResult{}, ArtifactRef{}, err
	}
	if execution.Status != "ready" || execution.DeliveryRef != candidateReceipt.CandidateRef || execution.DeliverySHA256 != candidateReceipt.CandidateSHA256 || !sameAbsolutePath(execution.Path, candidateReceipt.Provenance.ExecutionRoot) || execution.AssignmentID != candidateReceipt.Provenance.ExecutionAssignmentID || execution.Generation != candidateReceipt.Provenance.ExecutionGeneration || execution.ReworkRef != candidateReceipt.Provenance.ReworkRef {
		return current, RepairResult{}, ArtifactRef{}, errors.New("S9 accepted candidate no longer matches its current registered Worker execution/attempt")
	}
	head, err := workspace.Git(ctx, execution.Path, "rev-parse", "HEAD")
	if err != nil || head != candidateReceipt.Provenance.SourceHead {
		return current, RepairResult{}, ArtifactRef{}, errors.New("S9 Worker source HEAD changed after candidate acceptance; use explicit rework")
	}
	statusArgs := append([]string{"status", "--porcelain", "--untracked-files=all", "--", "."}, workspace.WorkerSubmissionStatusExcludes()...)
	status, err := workspace.Git(ctx, execution.Path, statusArgs...)
	if err != nil || status != "" {
		return current, RepairResult{}, ArtifactRef{}, fmt.Errorf("S9 Worker source changed after candidate acceptance: %s %v", status, err)
	}
	if err := verifyRepairResultExecutionChecks(ctx, registry, execution); err != nil {
		return current, RepairResult{}, ArtifactRef{}, err
	}
	candidateRef := ArtifactRef{Path: candidateReceipt.CandidateRef, SHA256: candidateReceipt.CandidateSHA256}
	candidate, err := ValidateWorkspaceCandidate(root, current.State, execution, candidateRef)
	if err != nil || candidate.Provenance == nil || !reflect.DeepEqual(*candidate.Provenance, candidateReceipt.Provenance) {
		return current, RepairResult{}, ArtifactRef{}, fmt.Errorf("S9 accepted candidate artifact no longer validates: %v", err)
	}
	if err := ensureRepairResultCompleted(candidateReceipt.Provenance.SubmittedAnswer); err != nil {
		return current, RepairResult{}, ArtifactRef{}, err
	}
	checks, err := repairResultMainChecks(root, packet, candidateReceipt.Provenance)
	if err != nil {
		return current, RepairResult{}, ArtifactRef{}, err
	}
	checkpoint, proof, err := readRepairResultMainProof(root, packet, candidateReceipt.Provenance, checks)
	if err != nil {
		return current, RepairResult{}, ArtifactRef{}, err
	}
	if req.MainProof.CheckpointRef != proof.CheckpointRef || req.MainProof.CheckpointSHA256 != proof.CheckpointSHA256 || req.MainProof.VerifiedFactsSHA256 != proof.VerifiedFactsSHA256 || string(req.MainProof.VerifiedFacts) != string(proof.VerifiedFacts) {
		return current, RepairResult{}, ArtifactRef{}, errors.New("S9 caller Main proof differs from the exact retained verified checkpoint")
	}
	if head, err := workspace.Git(ctx, root, "rev-parse", "HEAD"); err != nil || head != checkpoint.TestedHead {
		return current, RepairResult{}, ArtifactRef{}, errors.New("Main HEAD moved after the original required checks; rerun the registered integration checks")
	}
	checksDoc, err := repairResultChecksDocument(checks)
	if err != nil {
		return current, RepairResult{}, ArtifactRef{}, err
	}
	if err := runtimepkg.VerifyRepairResultMainProof(root, packet, candidateReceipt.Provenance, proof, checksDoc); err != nil {
		return current, RepairResult{}, ArtifactRef{}, err
	}
	pointer := repairPointer(current.State)
	if pointer == nil || stringField(pointer["status"]) != "repairing" || stringField(pointer["session_id"]) != packet.Binding.Sources.SessionID || stringField(pointer["plan_ref"]) != packet.Binding.Sources.PlanRef || stringField(pointer["plan_sha256"]) != packet.Binding.Sources.PlanSHA256 {
		return current, RepairResult{}, ArtifactRef{}, errors.New("S9 current Session/Plan/repair phase differs from the issued result")
	}
	planRef, err := pointerArtifact(pointer, "plan_ref", "plan_sha256", "current RepairPlan")
	if err != nil {
		return current, RepairResult{}, ArtifactRef{}, err
	}
	plan, err := ValidateRepairPlan(root, planRef)
	if err != nil {
		return current, RepairResult{}, ArtifactRef{}, err
	}
	assignment, ok := assignmentByID(plan.Assignments, packet.Binding.Sources.RepairAssignmentID)
	if !ok {
		return current, RepairResult{}, ArtifactRef{}, errors.New("S9 issued Result Assignment disappeared from current Plan")
	}
	if err := validateRepairAssignmentReady(root, pointer, plan, assignment); err != nil {
		return current, RepairResult{}, ArtifactRef{}, err
	}
	if err := checkS9AuthorityFreshness(root, pointer, candidate.Proposal.ChangedArtifacts); err != nil {
		return current, RepairResult{}, ArtifactRef{}, err
	}
	resultPath := artifactRoot + "/results/" + packet.ResultID + ".json"
	sequence := 1
	journalBytes, err := controlproof.ReadJournalForState(root, current.State)
	if err != nil {
		return current, RepairResult{}, ArtifactRef{}, err
	}
	journal, err := controlproof.InspectJournalData(journalBytes)
	if err != nil || controlproof.ValidateStateJournalPair(current.State, journal) != nil {
		return current, RepairResult{}, ArtifactRef{}, fmt.Errorf("inspect original S9 Runtime journal before final Result: %v", err)
	}
	sequence = journal.TailSequence + 1
	receipt := runtimepkg.RepairResultReceipt{
		Contract: repairResultFinalContract, Provenance: candidateReceipt.Provenance,
		CandidateRef: candidateReceipt.CandidateRef, CandidateSHA256: candidateReceipt.CandidateSHA256,
		ResultID: packet.ResultID, ResultRef: resultPath, MainProof: proof, WorkComplete: true,
		AcceptedEventID:              "evt-s9-issued-repair-result-" + packet.ID,
		AcceptedIdempotencyKey:       "runtime:s9:issued-repair-result:" + packet.ID,
		AcceptedAfterRuntimeRevision: current.Revision + 1, AcceptedSequence: sequence,
	}
	capability := &issuedRepairResultFinalCapability{packet: packet, provenance: candidateReceipt.Provenance, receipt: receipt, ctx: ctx, mainRoot: root}
	proposal := candidate.Proposal
	proposal.Checks = checks
	proposal.ResultID = packet.ResultID
	proposal.ProducerAgentID = req.OwnerAgentID
	proposal.AssignmentID = packet.Binding.Sources.RepairAssignmentID
	proposal.PlanReport = ArtifactRef{Path: packet.Binding.PlanReportRef, SHA256: packet.Binding.PlanReportSHA256}
	proposal.Verdict = "completed"
	proposal.Result = "pass"
	proposal.Receipt = &receipt
	proposal.issuedFinal = capability
	proposal.OccurredAt = issueAt
	if err := capability.validate(root, current.State, proposal); err != nil {
		return current, RepairResult{}, ArtifactRef{}, fmt.Errorf("S9 Main proof preflight: %w", err)
	}
	capability.cleanup = func(expected RepairResult) error {
		return cleanUnacceptedRepairResultStaging(root, writer, current, expected)
	}
	if req.OccurredAt.IsZero() {
		req.OccurredAt = time.Now().UTC()
	}
	return SubmitRepairResultToRuntime(root, statePath, journalPath, SubmitResultRuntimeRequest{
		RuntimeRequest: RuntimeRequest{ExpectedRevision: current.Revision, Actor: req.OwnerAgentID, OccurredAt: req.OccurredAt},
		Result:         proposal, issued: capability,
	})
}

func (capability *issuedRepairResultFinalCapability) validate(root string, state map[string]any, request RepairResultRequest) error {
	if capability == nil || capability.ctx == nil || capability.packet.ID == "" || request.issuedFinal != capability || request.Receipt == nil || !reflect.DeepEqual(request.Receipt, &capability.receipt) {
		return errors.New("S9 final Result lacks its private original Main integration capability")
	}
	packet, found, err := runtimepkg.FindRepairResultPacket(state, capability.packet.ID)
	if err != nil || !found || !sameRepairResultPacket(packet, capability.packet) {
		return errors.New("S9 issued packet changed before original Result CAS")
	}
	if err := runtimepkg.VerifyRepairResultPacketIssueAndCurrent(root, state, packet); err != nil {
		return err
	}
	issueRevision, _, err := repairResultIssueCoordinates(root, state, packet)
	if err != nil {
		return err
	}
	if err := runtimepkg.VerifyRepairResultSources(root, state, packet, issueRevision); err != nil {
		return err
	}
	candidateReceipt, accepted, err := runtimepkg.VerifiedRepairResultCandidateReceipt(root, state, packet.ID, request.ProducerAgentID)
	if err != nil || !accepted || candidateReceipt.CandidateRef != capability.receipt.CandidateRef || candidateReceipt.CandidateSHA256 != capability.receipt.CandidateSHA256 || !reflect.DeepEqual(candidateReceipt.Provenance, capability.provenance) {
		return fmt.Errorf("S9 final Result candidate receipt changed: %v", err)
	}
	registry, execution, err := currentRepairExecution(capability.ctx, root, state, packet, true)
	if err != nil {
		return err
	}
	if execution.Status != "ready" || execution.DeliveryRef != candidateReceipt.CandidateRef || execution.DeliverySHA256 != candidateReceipt.CandidateSHA256 || execution.AgentID != request.ProducerAgentID || execution.AssignmentID != capability.provenance.ExecutionAssignmentID || execution.Generation != capability.provenance.ExecutionGeneration || !sameAbsolutePath(execution.Path, capability.provenance.ExecutionRoot) || execution.ReworkRef != capability.provenance.ReworkRef {
		return errors.New("S9 current Worker execution no longer matches the accepted candidate")
	}
	head, err := workspace.Git(capability.ctx, execution.Path, "rev-parse", "HEAD")
	if err != nil || head != capability.provenance.SourceHead {
		return errors.New("S9 Worker source HEAD changed before final Result CAS")
	}
	statusArgs := append([]string{"status", "--porcelain", "--untracked-files=all", "--", "."}, workspace.WorkerSubmissionStatusExcludes()...)
	status, err := workspace.Git(capability.ctx, execution.Path, statusArgs...)
	if err != nil || status != "" {
		return fmt.Errorf("S9 Worker source changed before final Result CAS: %s %v", status, err)
	}
	if err := verifyRepairResultExecutionChecks(capability.ctx, registry, execution); err != nil {
		return err
	}
	candidateRef := ArtifactRef{Path: candidateReceipt.CandidateRef, SHA256: candidateReceipt.CandidateSHA256}
	candidate, err := ValidateWorkspaceCandidate(root, state, execution, candidateRef)
	if err != nil || candidate.Provenance == nil || !reflect.DeepEqual(*candidate.Provenance, candidateReceipt.Provenance) {
		return fmt.Errorf("S9 exact candidate artifact changed before final Result CAS: %v", err)
	}
	checkpoint, proof, err := readRepairResultMainProof(root, packet, capability.provenance, request.Checks)
	if err != nil {
		return err
	}
	if checkpoint.State != integration.StateVerified && checkpoint.State != integration.StateAcknowledged && checkpoint.State != integration.StateCleanupPending && checkpoint.State != integration.StateComplete {
		return errors.New("S9 original Main checkpoint is no longer in a verified lifecycle")
	}
	if proof.CheckpointRef != capability.receipt.MainProof.CheckpointRef || proof.CheckpointSHA256 != capability.receipt.MainProof.CheckpointSHA256 || proof.VerifiedFactsSHA256 != capability.receipt.MainProof.VerifiedFactsSHA256 || string(proof.VerifiedFacts) != string(capability.receipt.MainProof.VerifiedFacts) {
		return errors.New("S9 Main checkpoint changed between preflight and original Result CAS")
	}
	if head, err := workspace.Git(capability.ctx, capability.mainRoot, "rev-parse", "HEAD"); err != nil || head != checkpoint.TestedHead {
		return errors.New("Main HEAD changed between check verification and original Result CAS")
	}
	checksDoc, err := repairResultChecksDocument(request.Checks)
	if err != nil {
		return err
	}
	if err := runtimepkg.VerifyRepairResultMainProof(root, packet, capability.provenance, proof, checksDoc); err != nil {
		return err
	}
	if request.Verdict != "completed" || request.Result != "pass" || request.ResultID != packet.ResultID || request.AssignmentID != packet.Binding.Sources.RepairAssignmentID || request.ProducerAgentID != packet.Binding.Sources.OwnerAgentID || !reflect.DeepEqual(request.ChangedArtifacts, candidate.Proposal.ChangedArtifacts) {
		return errors.New("S9 final Result projection differs from its complete accepted candidate")
	}
	if err := checkS9AuthorityFreshness(root, repairPointer(state), request.ChangedArtifacts); err != nil {
		return err
	}
	pointer := repairPointer(state)
	planRef, err := pointerArtifact(pointer, "plan_ref", "plan_sha256", "current RepairPlan")
	if err != nil {
		return err
	}
	plan, err := ValidateRepairPlan(root, planRef)
	if err != nil {
		return err
	}
	assignment, ok := assignmentByID(plan.Assignments, packet.Binding.Sources.RepairAssignmentID)
	if !ok {
		return errors.New("S9 Result Assignment disappeared before original Result CAS")
	}
	if err := validateRepairAssignmentReady(root, pointer, plan, assignment); err != nil {
		return err
	}
	return nil
}

func repairResultIssueCoordinates(root string, state map[string]any, packet workpacket.RepairResultPacket) (int, time.Time, error) {
	data, err := controlproof.ReadJournalForState(root, state)
	if err != nil {
		return 0, time.Time{}, err
	}
	journal, err := controlproof.InspectJournalData(data)
	if err != nil {
		return 0, time.Time{}, err
	}
	if err := controlproof.ValidateStateJournalPair(state, journal); err != nil {
		return 0, time.Time{}, err
	}
	index, found := journal.EventIndex[packet.IssuanceEventID]
	if !found {
		return 0, time.Time{}, errors.New("S9 original RepairResult PACKET-ISSUE event is missing")
	}
	event := journal.Events[index]
	after, err := integerRepairResultEvent(event, "after_revision")
	before, beforeErr := integerRepairResultEvent(event, "before_revision")
	baselineGeneration, generationErr := integerRepairResultEvent(event, "baseline_generation")
	currentRevision := integerRepairResult(state["revision"])
	actor, _ := event["actor"].(map[string]any)
	from, _ := event["from"].(map[string]any)
	to, _ := event["to"].(map[string]any)
	if err != nil || beforeErr != nil || generationErr != nil ||
		event["event"] != "transition_committed" || event["outcome"] != "committed" ||
		event["transition_id"] != "PACKET-ISSUE" || event["event_id"] != packet.IssuanceEventID ||
		event["idempotency_key"] != "packet:"+packet.ID || event["runtime_id"] != packet.Binding.Sources.RuntimeID ||
		actor["type"] != "orchestrator" || actor["id"] != "orchestrator" ||
		baselineGeneration != packet.Binding.Sources.BaselineGeneration || after != before+1 || after > currentRevision ||
		from["state"] != "bug_resolution" || from["phase"] != "fixing" || !reflect.DeepEqual(from, to) {
		return 0, time.Time{}, errors.New("S9 original RepairResult issue event coordinates are invalid")
	}
	text, _ := event["occurred_at"].(string)
	at, err := time.Parse(time.RFC3339Nano, text)
	if err != nil || text == "" {
		return 0, time.Time{}, errors.New("S9 original RepairResult issue timestamp is invalid")
	}
	return after, at.UTC(), nil
}

func integerRepairResultEvent(event map[string]any, key string) (int, error) {
	switch n := event[key].(type) {
	case int:
		return n, nil
	case int64:
		return int(n), nil
	case float64:
		return int(n), nil
	case json.Number:
		v, err := n.Int64()
		return int(v), err
	default:
		return 0, fmt.Errorf("S9 event %s is not an integer", key)
	}
}

func repairResultCheckpointPath(root string, packet workpacket.RepairResultPacket) (string, string, error) {
	var execution struct {
		AssignmentID string                            `json:"assignment_id"`
		Record       *assignmentrecord.RecordReference `json:"assignment_record"`
	}
	if err := json.Unmarshal(packet.Inputs.Execution, &execution); err != nil || execution.AssignmentID == "" {
		return "", "", errors.New("S9 frozen execution lacks an original checkpoint slot identity")
	}
	path := ""
	if execution.Record != nil {
		path = assignmentrecord.IntegrationCheckpointPath(root, *execution.Record)
	} else {
		path = filepath.Join(root, ".claude", "evidence", packet.Binding.Sources.RuntimeID, fmt.Sprintf("g%d", packet.Binding.Sources.BaselineGeneration), "worktree", execution.AssignmentID, "checkpoint.json")
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", "", err
	}
	relative, err := filepath.Rel(root, absolute)
	if err != nil || !filepath.IsLocal(relative) {
		return "", "", errors.New("S9 original checkpoint path escaped the Main root")
	}
	return absolute, filepath.ToSlash(relative), nil
}

func repairResultMainChecks(root string, packet workpacket.RepairResultPacket, provenance runtimepkg.RepairResultCandidateProvenance) ([]RepairCheck, error) {
	_, checkpointPath, err := repairResultCheckpointPath(root, packet)
	if err != nil {
		return nil, err
	}
	data, err := workpacket.ReadFile(root, checkpointPath, 20<<20)
	if err != nil {
		return nil, fmt.Errorf("read current S9 Main checkpoint: %w", err)
	}
	var checkpoint integration.Checkpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return nil, fmt.Errorf("decode S9 Main checkpoint: %w", err)
	}
	var execution struct {
		Checks   []string            `json:"required_checks"`
		Profiles []checkexec.Profile `json:"check_profiles"`
	}
	if err := json.Unmarshal(packet.Inputs.Execution, &execution); err != nil {
		return nil, err
	}
	profiles := map[int]checkexec.Profile{}
	for _, profile := range execution.Profiles {
		if _, duplicate := profiles[profile.CheckIndex]; duplicate {
			return nil, errors.New("S9 frozen Main check profiles have a duplicate index")
		}
		profiles[profile.CheckIndex] = profile
	}
	out := []RepairCheck{}
	seen := map[string]bool{}
	for index, command := range execution.Checks {
		if strings.HasPrefix(strings.TrimSpace(command), "locked:") {
			continue
		}
		profile, ok := profiles[index]
		if !ok || profile.Command != command {
			return nil, fmt.Errorf("S9 Main required check %d has no exact registered profile", index)
		}
		var selected *checkexec.ToolReceipt
		for i := range checkpoint.ToolCheckReceipts {
			ref := &checkpoint.ToolCheckReceipts[i]
			if ref.Command != command || ref.ProfileDigest != profile.Digest() {
				continue
			}
			if selected != nil {
				return nil, fmt.Errorf("S9 Main required check %d has duplicate checkpoint receipts", index)
			}
			selected = ref
		}
		if selected == nil || selected.Path == "" || selected.SHA256 == "" || seen[selected.Path] {
			return nil, fmt.Errorf("S9 Main required check %d has no unique exact captured receipt", index)
		}
		seen[selected.Path] = true
		relative, err := filepath.Rel(root, selected.Path)
		if err != nil || !filepath.IsLocal(relative) {
			return nil, errors.New("S9 Main check receipt escaped the registered Main evidence root")
		}
		out = append(out, RepairCheck{Name: command, Command: command, Result: "pass", EvidenceRefs: []string{"path:" + filepath.ToSlash(relative) + "#sha256=" + selected.SHA256}})
	}
	if len(out) == 0 {
		return nil, errors.New("S9 final Main proof has no non-locked required checks")
	}
	_ = provenance // The Runtime verifier binds every receipt to this exact provenance.
	return out, nil
}

func readRepairResultMainProof(root string, packet workpacket.RepairResultPacket, provenance runtimepkg.RepairResultCandidateProvenance, checks []RepairCheck) (integration.Checkpoint, runtimepkg.RepairResultMainProof, error) {
	var zero integration.Checkpoint
	checkpointPath, relative, err := repairResultCheckpointPath(root, packet)
	if err != nil {
		return zero, runtimepkg.RepairResultMainProof{}, err
	}
	data, err := workpacket.ReadFile(root, relative, 20<<20)
	if err != nil {
		return zero, runtimepkg.RepairResultMainProof{}, fmt.Errorf("read registered S9 Main checkpoint: %w", err)
	}
	var checkpoint integration.Checkpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return zero, runtimepkg.RepairResultMainProof{}, err
	}
	if err := integration.VerifyCapturedChecks(checkpointPath, checkpoint); err != nil {
		return zero, runtimepkg.RepairResultMainProof{}, fmt.Errorf("verify original S9 Main captured checks: %w", err)
	}
	switch checkpoint.State {
	case integration.StateVerified, integration.StateAcknowledged, integration.StateCleanupPending, integration.StateComplete:
	default:
		return zero, runtimepkg.RepairResultMainProof{}, fmt.Errorf("S9 Main checkpoint state %q is not verified", checkpoint.State)
	}
	facts, err := runtimepkg.ProjectRepairResultMainFacts(data)
	if err != nil {
		return zero, runtimepkg.RepairResultMainProof{}, err
	}
	proof := runtimepkg.RepairResultMainProof{
		CheckpointRef: relative, CheckpointSHA256: workpacket.Digest(data),
		VerifiedFacts: facts, VerifiedFactsSHA256: workpacket.Digest(facts),
	}
	checksDoc, err := repairResultChecksDocument(checks)
	if err != nil {
		return zero, runtimepkg.RepairResultMainProof{}, err
	}
	if checkpoint.AssignmentID == "" || checkpoint.BaselineGeneration != packet.Binding.Sources.BaselineGeneration || checkpoint.SourceHead != provenance.SourceHead || checkpoint.TestedHead == "" || checkpoint.WorktreePath == "" {
		return zero, runtimepkg.RepairResultMainProof{}, errors.New("S9 checkpoint does not bind the original candidate source and generation")
	}
	if err := runtimepkg.VerifyRepairResultMainProof(root, packet, provenance, proof, checksDoc); err != nil {
		return zero, runtimepkg.RepairResultMainProof{}, err
	}
	return checkpoint, proof, nil
}

func repairResultChecksDocument(checks []RepairCheck) (map[string]any, error) {
	encoded, err := json.Marshal(map[string]any{"checks": checks})
	if err != nil {
		return nil, err
	}
	var document map[string]any
	if err := json.Unmarshal(encoded, &document); err != nil {
		return nil, err
	}
	return document, nil
}

func canonicalRuntimeFacts(facts json.RawMessage) (json.RawMessage, error) {
	if err := workpacket.ValidateUniqueJSON(facts); err != nil {
		return nil, err
	}
	var value any
	decoder := json.NewDecoder(strings.NewReader(string(facts)))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	return json.Marshal(value)
}

func existingIssuedRepairResult(root string, state map[string]any, resultID string) (RepairResult, ArtifactRef, bool, error) {
	pointer := repairPointer(state)
	if pointer == nil {
		return RepairResult{}, ArtifactRef{}, false, nil
	}
	refs := existingArtifactRefs(pointer["result_refs"])
	if len(refs) == 0 && stringField(pointer["result_ref"]) != "" {
		refs = []ArtifactRef{{Path: stringField(pointer["result_ref"]), SHA256: stringField(pointer["result_sha256"])}}
	}
	for _, ref := range refs {
		result, err := ValidateRepairResult(root, ref)
		if err != nil {
			return RepairResult{}, ArtifactRef{}, false, err
		}
		if result.ResultID == resultID {
			return result, ref, true, nil
		}
	}
	return RepairResult{}, ArtifactRef{}, false, nil
}

func repairResultPointerRows(pointer map[string]any, next ArtifactRef, root string) ([]any, error) {
	var refs []ArtifactRef
	if raw, exists := pointer["result_refs"]; exists && raw != nil {
		parsed, err := artifactRefsFromAny(raw, "result_refs")
		if err != nil {
			return nil, err
		}
		refs = parsed
	} else if stringField(pointer["result_ref"]) != "" {
		ref, err := pointerArtifact(pointer, "result_ref", "result_sha256", "current RepairResult")
		if err != nil {
			return nil, err
		}
		refs = []ArtifactRef{ref}
	}
	rows := make([]any, 0, len(refs)+1)
	seenPaths := map[string]bool{}
	seenAssignments := map[string]bool{}
	for _, ref := range refs {
		if ref.Path == next.Path || seenPaths[ref.Path] {
			return nil, fmt.Errorf("S9 current result index repeats or already contains %s", ref.Path)
		}
		seenPaths[ref.Path] = true
		result, err := ValidateRepairResult(root, ref)
		if err != nil {
			return nil, fmt.Errorf("validate retained S9 Result before appending typed sibling: %w", err)
		}
		if seenAssignments[result.AssignmentID] {
			return nil, fmt.Errorf("S9 current Result index repeats Assignment %s", result.AssignmentID)
		}
		seenAssignments[result.AssignmentID] = true
		rows = append(rows, map[string]any{"path": ref.Path, "sha256": ref.SHA256})
	}
	rows = append(rows, map[string]any{"path": next.Path, "sha256": next.SHA256})
	return rows, nil
}

// cleanUnacceptedRepairResultStaging removes only the exact immutable Result
// bytes produced for this unaccepted typed attempt. A pending Runtime write,
// current pointer reference, revision change, changed bytes, or any content
// difference outside the two predicted acceptance coordinates keeps the file
// in place for recovery rather than risking deletion of accepted evidence.
func cleanUnacceptedRepairResultStaging(root string, writer *runtimepkg.Store, snapshot runtimepkg.Snapshot, expected RepairResult) error {
	if writer == nil || expected.Receipt == nil || expected.ResultID == "" || expected.Receipt.ResultRef == "" {
		return errors.New("S9 unaccepted Result cleanup requires the exact staged typed Result")
	}
	path := expected.Receipt.ResultRef
	data, err := workpacket.ReadFile(root, path, repairResultAnswerLimit)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("preserve unrecognized S9 Result staging path %s: %w", path, err)
	}
	sha := workpacket.Digest(data)
	staged, err := ValidateRepairResult(root, ArtifactRef{Path: path, SHA256: sha})
	if err != nil || staged.Receipt == nil || staged.ResultID != expected.ResultID || staged.AssignmentID != expected.AssignmentID ||
		staged.SessionID != expected.SessionID || staged.PlanID != expected.PlanID || staged.ProducerAgentID != expected.ProducerAgentID ||
		staged.Receipt.CandidateRef != expected.Receipt.CandidateRef || staged.Receipt.CandidateSHA256 != expected.Receipt.CandidateSHA256 ||
		staged.Receipt.ResultRef != path || staged.Receipt.ResultID != expected.ResultID ||
		staged.Receipt.AcceptedEventID != expected.Receipt.AcceptedEventID || staged.Receipt.AcceptedIdempotencyKey != expected.Receipt.AcceptedIdempotencyKey ||
		staged.Receipt.WorkComplete != expected.Receipt.WorkComplete {
		return fmt.Errorf("preserve existing S9 Result staging path %s: it is not the exact unaccepted packet/candidate artifact (%v)", path, err)
	}
	// A retry can advance only after another sibling Result was accepted. The
	// stable packet issue time keeps SubmittedAt fixed; only predicted Runtime
	// revision and journal sequence may be refreshed before this Result's CAS.
	staged.Receipt.AcceptedAfterRuntimeRevision = expected.Receipt.AcceptedAfterRuntimeRevision
	staged.Receipt.AcceptedSequence = expected.Receipt.AcceptedSequence
	if !reflect.DeepEqual(staged, expected) {
		return fmt.Errorf("preserve existing S9 Result staging path %s: staged content differs from this exact unaccepted retry", path)
	}
	for _, ref := range existingArtifactRefs(repairPointer(snapshot.State)["result_refs"]) {
		if ref.Path == path {
			return fmt.Errorf("preserve S9 Result staging artifact %s because the current Runtime pointer already references it", path)
		}
	}
	removed, err := writer.RemoveUnreferencedArtifact(runtimepkg.ArtifactCleanupRequest{
		ExpectedRevision: snapshot.Revision, ArtifactPath: path, ArtifactSHA256: sha,
		ReferencedPaths: stateArtifactPaths(snapshot.State),
	})
	if err != nil {
		return fmt.Errorf("clean exact unaccepted S9 Result staging artifact %s: %w", path, err)
	}
	if !removed {
		return fmt.Errorf("preserve S9 Result staging artifact %s: pending Runtime state or a current reference prevents safe cleanup; inspect recovery before retry", path)
	}
	return nil
}
