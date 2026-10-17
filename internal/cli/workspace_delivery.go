package cli

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/entroforge/go-system-builder/internal/checkexec"
	"github.com/entroforge/go-system-builder/internal/hookctx"
	"github.com/entroforge/go-system-builder/internal/integration"
	"github.com/entroforge/go-system-builder/internal/repair"
	"github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/semantic"
	"github.com/entroforge/go-system-builder/internal/workpacket"
	"github.com/entroforge/go-system-builder/internal/workspace"
)

func runWorkspaceDelivery(args []string, stdout, stderr io.Writer) int {
	verb := args[0]
	fs := flag.NewFlagSet("runtime workspace "+verb, flag.ContinueOnError)
	fs.SetOutput(stderr)
	root := fs.String("root", ".", "Main or registered Worker root")
	id := fs.String("assignment", "", "registered assignment")
	agent := fs.String("agent", "", "registered owner")
	retry := fs.Bool("retry-preserved", false, "retry the same immutable candidate after resolving its recorded failure")
	if err := parseWorkspaceFlags(fs, args[1:]); err != nil {
		return 2
	}
	fail := func(err error) int { fmt.Fprintln(stderr, err); return 1 }
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	ctx, release, err := integration.LockWorkspace(ctx, *root)
	if err != nil {
		return fail(err)
	}
	defer release()
	statePath, journalPath := filepath.Join(*root, ".claude/loop-state.json"), filepath.Join(*root, ".claude/loop-events.jsonl")
	writer := runtime.NewWriter(statePath, journalPath, *root, semantic.RuntimeCandidateValidator{})
	snap, err := writer.Snapshot()
	if err != nil {
		return fail(err)
	}
	b, err := workspace.Decode(snap.State)
	if err != nil {
		return fail(err)
	}
	if b == nil {
		return fail(fmt.Errorf("bind Main first"))
	}
	if err = b.Validate(ctx, *root); err != nil {
		return fail(err)
	}
	e, ok := b.Execution(*id, workspace.RuntimeID(snap.State), workspace.Generation(snap.State))
	if !ok || e.AgentID != *agent || (e.Status != "ready" && !(verb == "integrate" && e.Status == "complete")) {
		return fail(fmt.Errorf("delivery requires the registered execution and owner"))
	}
	if e.Purpose == workspace.DocumentReviewPurpose {
		return fail(fmt.Errorf("S5 document review has no product delivery/integration; use the answer receipt"))
	}
	loaded, err := hookctx.LoadFull(*root, e.AgentID)
	if err != nil {
		return fail(err)
	}
	var a *hookctx.AssignmentContext
	for i := range loaded.Assignments {
		if loaded.Assignments[i].AssignmentID == e.AssignmentID && loaded.Assignments[i].OwnerAgentID == e.AgentID {
			a = &loaded.Assignments[i]
			break
		}
	}
	if a == nil {
		return fail(fmt.Errorf("delivery assignment is no longer registered"))
	}
	if verb == "deliver" {
		if loaded.PolicyContext.Agent == nil || loaded.PolicyContext.Agent.LifecycleActivity() != "reported" {
			return fail(fmt.Errorf("record the genuine Worker completion_report before proposing S9 delivery; the final domain RepairResult is submitted only after merge"))
		}
		if err = b.ValidateExecution(ctx, e, false); err != nil {
			return fail(err)
		}
		if err = b.ValidateInputs(e); err != nil {
			return fail(err)
		}
		submissionPaths, err := b.WorkerSubmissionPaths(e)
		if err != nil {
			return fail(err)
		}
		path := filepath.Join(e.Path, filepath.FromSlash(submissionPaths.RepairResult))
		physical, err := workspace.Canonical(path)
		if err != nil || physical != path {
			return fail(fmt.Errorf("candidate proposal must be a local Worker submission"))
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return fail(err)
		}
		var proposal repair.RepairResultRequest
		if err = json.Unmarshal(data, &proposal); err != nil {
			return fail(err)
		}
		candidate, ref, err := repair.CreateWorkspaceCandidate(ctx, *root, snap.State, e, proposal)
		if err != nil {
			return fail(err)
		}
		if e.DeliveryRef != "" && (e.DeliveryRef != ref.Path || e.DeliverySHA256 != ref.SHA256) {
			return fail(fmt.Errorf("execution already has an immutable delivery; preserve it for explicit recovery"))
		}
		if e.DeliveryRef == "" {
			e.DeliveryRef, e.DeliverySHA256 = ref.Path, ref.SHA256
			b.Executions[e.AssignmentID] = e
			key := fmt.Sprintf("workspace-delivery-r%d", snap.Revision+1)
			snap, err = writer.Update(snap.Revision, runtime.Mutation{EventID: key, TransitionID: "WORKSPACE", Event: "workspace_updated", Actor: e.AgentID, RuntimeID: e.RuntimeID, IdempotencyKey: key, RetainLastTransition: true, OccurredAt: time.Now().UTC(), Message: "immutable S9 candidate proposed; no final result or PASS registered", Apply: func(state map[string]any) error { state["workspace"] = workspace.Encode(b); return nil }})
			if err != nil {
				return fail(err)
			}
		}
		return encodeJSON(stdout, map[string]any{"candidate": candidate, "artifact_ref": ref, "revision": snap.Revision, "next_command_argv": []string{"loop-harness", "runtime", "workspace", "integrate", "--root", *root, "--assignment", e.AssignmentID, "--agent", e.AgentID}})
	}
	if err = workspace.RequireMain(*root); err != nil {
		return fail(err)
	}
	ref := repair.ArtifactRef{Path: e.DeliveryRef, SHA256: e.DeliverySHA256}
	candidate, err := repair.ReadWorkspaceCandidate(*root, ref)
	if err != nil {
		return fail(err)
	}
	if candidate.RuntimeID != e.RuntimeID || candidate.BaselineGeneration != e.BaselineGeneration || candidate.ExecutionGeneration != e.Generation || candidate.AssignmentID != e.AssignmentID || candidate.BaseCommit != e.BaseCommit {
		return fail(fmt.Errorf("candidate execution identity mismatch"))
	}
	typedPacket, typedResult, err := currentRepairResultPacketForDelivery(*root, snap.State, e, candidate)
	if err != nil {
		return fail(err)
	}
	resultRef, registered, err := repair.RegisteredWorkspaceResult(*root, snap.State, candidate)
	if err != nil {
		return fail(err)
	}
	if !registered {
		if _, err = repair.ValidateWorkspaceCandidate(*root, snap.State, e, ref); err != nil {
			return fail(err)
		}
	}
	if typedResult && !registered {
		if err := runtime.VerifyRepairResultPacketCurrent(*root, snap.State, typedPacket); err != nil {
			return fail(err)
		}
		accepted, acceptedOK, acceptedErr := runtime.VerifiedRepairResultCandidateReceipt(*root, snap.State, typedPacket.ID, e.AgentID)
		if acceptedErr != nil || !acceptedOK || accepted.CandidateRef != e.DeliveryRef || accepted.CandidateSHA256 != e.DeliverySHA256 {
			return fail(fmt.Errorf("typed S9 candidate lacks its exact original delivery receipt: %v", acceptedErr))
		}
		complete, completeErr := repair.RepairResultWorkCompleteWithContext(ctx, *root, snap.State, typedPacket.ID, e.AgentID)
		if completeErr != nil {
			return fail(fmt.Errorf("typed S9 candidate owner/check proof is unavailable; preserve the candidate: %w", completeErr))
		}
		if !complete {
			return fail(fmt.Errorf("typed S9 candidate is retained, but its professional answer or current owner-owned checks are incomplete; preserve the Worker and do not merge"))
		}
	}
	cpPath := workspaceExecutionCheckpointPath(*root, e)
	cp, found, err := integration.DefaultCheckpointStore().Load(cpPath)
	if err != nil {
		return fail(err)
	}
	if found && ((cp.SourceHead != "" && cp.SourceHead != candidate.SourceHead) || cp.WorktreePath != e.Path || cp.TargetBranch != e.TargetBranch || cp.BaselineGeneration != e.BaselineGeneration) {
		return fail(fmt.Errorf("candidate/checkpoint coordinates differ"))
	}
	if registered && !found {
		return fail(fmt.Errorf("registered candidate result has no integration checkpoint"))
	}
	if registered && typedResult {
		// A typed S9 Result freezes the Main facts from the checkpoint that was
		// originally verified. Validate that receipt before any recovery call
		// can change the checkpoint's verification facts.
		verified, receiptFound, verifyErr := runtime.VerifiedRepairResultReceipt(*root, snap.State, typedPacket.ID, e.AgentID)
		if verifyErr != nil || !receiptFound || verified.ResultRef != resultRef.Path || verified.ResultID != typedPacket.ResultID {
			return fail(fmt.Errorf("registered S9 Result does not have its exact original candidate/Main checkpoint receipt: %v", verifyErr))
		}
		if cp.State == integration.StateVerified {
			targetHead, headErr := workspace.Git(ctx, *root, "rev-parse", cp.TargetBranch)
			if headErr != nil {
				return fail(fmt.Errorf("cannot verify Main target HEAD for an accepted S9 Result: %w", headErr))
			}
			if targetHead != cp.TestedHead {
				return fail(fmt.Errorf("Main target HEAD moved after the accepted S9 Result; preserve its verified checkpoint, inspect runtime repair status and runtime workspace integration-status, and arrange a new authorized verification attempt"))
			}
		}
	}
	config := integration.IntegrateConfig{Root: *root, GitRoot: *root, RuntimeID: e.RuntimeID, RequiredChecks: e.Checks, CheckRunner: integration.ProfileCheckRunner(e.CheckProfiles, checkexec.Binding{RuntimeID: e.RuntimeID, Generation: e.BaselineGeneration, AssignmentID: e.AssignmentID, Owner: e.AgentID, ExecutionGeneration: e.Generation, Responsibility: "main_required_integration"})}
	var inspected integration.Inspection
	if found && (cp.State == integration.StateReady || cp.State == integration.StateMerged || cp.State == integration.StateVerified || cp.State == integration.StateAcknowledged || cp.State == integration.StateCleanupPending || cp.State == integration.StateComplete) {
		inspected = integration.InspectionFromCheckpoint(cp)
		if !(registered && typedResult) {
			inspected, err = integration.RefreshCompletionBinding(*root, e.RuntimeID, a.CompletionRef, inspected)
			if err != nil {
				return fail(err)
			}
		}
	} else {
		if err = b.CleanInputs(ctx); err != nil {
			return fail(err)
		}
		if err = b.ValidateExecution(ctx, e, false); err != nil {
			return fail(err)
		}
		a.WorktreePath, a.Branch, a.TargetBranch = e.Path, e.Branch, e.TargetBranch
		a.WritePaths = e.WritePaths
		locked := []string{}
		for _, v := range loaded.PolicyContext.LockedArtifacts {
			locked = append(locked, v.Path)
		}
		inspectCfg := integration.InspectConfig{LockedArtifacts: locked, ValidateDelivery: func(ctx context.Context, in integration.Inspection) error {
			if in.SourceHead != candidate.SourceHead {
				return fmt.Errorf("candidate source commit changed")
			}
			_, err := repair.ValidateWorkspaceCandidate(*root, snap.State, e, ref)
			return err
		}}
		if found && cp.MergeCommit != "" {
			inspectCfg.RecoveryBase, err = integration.RecoveryBaseForCheckpoint(ctx, *root, cp)
			if err != nil {
				return fail(err)
			}
		}
		inspected, err = integration.Inspect(ctx, integration.InspectRequest{Root: *root, Assignment: *a, TargetBranch: e.TargetBranch, BaselineGeneration: e.BaselineGeneration, RuntimeID: e.RuntimeID}, inspectCfg)
		if err != nil {
			return fail(err)
		}
	}
	merged, err := integration.Integrate(ctx, integration.IntegrateRequest{Inspection: inspected, RetryPreserved: *retry}, config)
	if err != nil {
		return fail(err)
	}
	cp = merged.Checkpoint
	if cp.State != integration.StateVerified && cp.State != integration.StateAcknowledged && cp.State != integration.StateCleanupPending && cp.State != integration.StateComplete {
		return fail(fmt.Errorf("candidate is not verified: %s %s", cp.State, cp.FailureReason))
	}
	if !registered {
		head, err := workspace.Git(ctx, *root, "rev-parse", "HEAD")
		if err != nil || head != cp.TestedHead {
			return fail(fmt.Errorf("Main moved after candidate verification"))
		}
		snap, err = writer.Snapshot()
		if err != nil {
			return fail(err)
		}
		candidate.Proposal.Checks, err = verifiedDomainChecks(*root, e, cp)
		if err != nil {
			return fail(err)
		}
		var result repair.RepairResult
		if typedResult {
			proof, proofErr := repairResultMainProof(*root, e, cp)
			if proofErr != nil {
				return fail(proofErr)
			}
			snap, result, resultRef, err = repair.SubmitIssuedRepairResult(ctx, *root, statePath, journalPath, repair.SubmitIssuedRepairResultRequest{
				ExpectedRevision: snap.Revision,
				PacketID:         typedPacket.ID,
				OwnerAgentID:     e.AgentID,
				MainProof:        proof,
				OccurredAt:       time.Now().UTC(),
			})
		} else {
			snap, result, resultRef, err = repair.SubmitRepairResultToRuntime(*root, statePath, journalPath, repair.SubmitResultRuntimeRequest{RuntimeRequest: repair.RuntimeRequest{ExpectedRevision: snap.Revision, Actor: e.AgentID}, Result: candidate.Proposal})
		}
		if err != nil {
			return fail(err)
		}
		if result.Result != "pass" {
			return fail(fmt.Errorf("domain result %s is %s; preserve Worker and follow the recorded repair recovery route", result.ResultID, result.Result))
		}
	} else if typedResult {
		verified, found, verifyErr := runtime.VerifiedRepairResultReceipt(*root, snap.State, typedPacket.ID, e.AgentID)
		if verifyErr != nil || !found || verified.ResultRef != resultRef.Path || verified.ResultID != typedPacket.ResultID {
			return fail(fmt.Errorf("registered S9 Result does not have its exact original candidate/Main checkpoint receipt: %v", verifyErr))
		}
	}
	if a.CompletionAckRef == "" {
		for _, other := range b.Executions {
			if other.AgentID != e.AgentID || other.AssignmentID == e.AssignmentID || other.DeliveryRef == "" || other.RuntimeID != e.RuntimeID || other.BaselineGeneration != e.BaselineGeneration {
				continue
			}
			c, err := repair.ReadWorkspaceCandidate(*root, repair.ArtifactRef{Path: other.DeliveryRef, SHA256: other.DeliverySHA256})
			if err != nil {
				return fail(err)
			}
			if _, done, err := repair.RegisteredWorkspaceResult(*root, snap.State, c); err != nil {
				return fail(err)
			} else if !done {
				return fail(fmt.Errorf("register verified domain result for %s before acknowledging shared owner", other.AssignmentID))
			}
		}
		snap, err = acknowledgeVerifiedIntegration(*root, snap, a, cp)
		if err != nil {
			return fail(err)
		}
	}
	finished, err := integration.Integrate(ctx, integration.IntegrateRequest{Inspection: integration.InspectionFromCheckpoint(cp), Acknowledge: true, Cleanup: true}, config)
	if err != nil {
		return fail(err)
	}
	snap, err = persistExecutionCompletion(*root, snap, e.AssignmentID)
	if err != nil {
		return fail(err)
	}
	return encodeJSON(stdout, map[string]any{"checkpoint": finished.Checkpoint, "result_ref": resultRef, "revision": snap.Revision, "next": "continue independent targeted verification and a fresh S7; human release remains required"})
}

// currentRepairResultPacketForDelivery identifies an issued typed S9 packet
// from immutable candidate provenance. If no provenance exists, any issued
// packet for this Assignment still prevents fallback to the legacy Writer.
func currentRepairResultPacketForDelivery(root string, state map[string]any, execution workspace.Execution, candidate repair.WorkspaceCandidate) (workpacket.RepairResultPacket, bool, error) {
	packets, err := runtime.RepairResultPackets(state)
	if err != nil {
		return workpacket.RepairResultPacket{}, false, err
	}
	if candidate.Provenance != nil {
		packet, found, err := runtime.FindRepairResultPacket(state, candidate.Provenance.PacketID)
		if err != nil || !found {
			return workpacket.RepairResultPacket{}, false, fmt.Errorf("candidate references an unregistered S9 packet: %v", err)
		}
		packetSHA, digestErr := workpacket.RepairResultPacketDigest(packet)
		if digestErr != nil {
			return workpacket.RepairResultPacket{}, false, digestErr
		}
		matches, matchErr := repairResultPacketMatchesExecution(packet, execution)
		if matchErr != nil {
			return workpacket.RepairResultPacket{}, false, matchErr
		}
		if !matches || !repairResultCandidateMatchesPacketExecution(candidate, packet, execution, packetSHA) {
			return workpacket.RepairResultPacket{}, false, fmt.Errorf("candidate typed S9 packet differs from the exact registered execution")
		}
		if err := runtime.VerifyRepairResultPacketIssue(root, state, packet); err != nil {
			return workpacket.RepairResultPacket{}, false, err
		}
		return packet, true, nil
	}
	var selected *workpacket.RepairResultPacket
	for _, packet := range packets {
		binding := packet.Binding.Sources
		if binding.SessionID != candidate.SessionID || binding.PlanID != candidate.PlanID || binding.ContractSHA256 != candidate.ContractSHA256 {
			continue
		}
		matches, matchErr := repairResultPacketMatchesExecution(packet, execution)
		if matchErr != nil {
			return workpacket.RepairResultPacket{}, false, matchErr
		}
		if !matches {
			continue
		}
		if selected != nil {
			return workpacket.RepairResultPacket{}, false, fmt.Errorf("multiple typed S9 packets bind this execution")
		}
		copyPacket := packet
		selected = &copyPacket
	}
	if selected == nil {
		return workpacket.RepairResultPacket{}, false, nil
	}
	return workpacket.RepairResultPacket{}, false, fmt.Errorf("issued S9 packet requires its exact typed candidate provenance; legacy Result submission is refused")
}

// repairResultPacketMatchesExecution first locates packets by the canonical
// or manifest Assignment identity, then proves that the frozen Execution is
// the exact registered workspace.Execution definition. The Assignment alias
// is not authority by itself; the frozen identity and definition digest must
// both agree with the registered execution.
func repairResultPacketMatchesExecution(packet workpacket.RepairResultPacket, execution workspace.Execution) (bool, error) {
	binding := packet.Binding.Sources
	if binding.OwnerAgentID != execution.AgentID || binding.RuntimeID != execution.RuntimeID ||
		binding.BaselineGeneration != execution.BaselineGeneration ||
		(execution.AssignmentID != binding.ManifestAssignmentID && execution.AssignmentID != binding.RepairAssignmentID) {
		return false, nil
	}
	var frozen struct {
		AssignmentID        string `json:"assignment_id"`
		RuntimeID           string `json:"runtime_id"`
		BaselineGeneration  int    `json:"baseline_generation"`
		ExecutionGeneration int    `json:"execution_generation"`
		AgentID             string `json:"agent_id"`
	}
	if len(packet.Inputs.Execution) == 0 || json.Unmarshal(packet.Inputs.Execution, &frozen) != nil {
		return true, fmt.Errorf("typed S9 packet has no valid frozen registered execution")
	}
	if frozen.AssignmentID != execution.AssignmentID || frozen.RuntimeID != execution.RuntimeID ||
		frozen.BaselineGeneration != execution.BaselineGeneration || frozen.ExecutionGeneration != execution.Generation ||
		frozen.AgentID != execution.AgentID {
		return true, fmt.Errorf("typed S9 packet frozen execution identity differs from the exact registered execution")
	}
	encoded, err := json.Marshal(execution)
	if err != nil {
		return true, err
	}
	var current map[string]any
	if err := json.Unmarshal(encoded, &current); err != nil {
		return true, err
	}
	projection, err := runtime.ProjectRepairResultExecution(current)
	if err != nil {
		return true, err
	}
	if workpacket.Digest(projection) != packet.Binding.ExecutionSHA256 || !bytes.Equal(projection, bytes.TrimSpace(packet.Inputs.Execution)) {
		return true, fmt.Errorf("typed S9 packet frozen execution definition differs from the exact registered execution")
	}
	return true, nil
}

func repairResultCandidateMatchesPacketExecution(candidate repair.WorkspaceCandidate, packet workpacket.RepairResultPacket, execution workspace.Execution, packetSHA string) bool {
	provenance := candidate.Provenance
	if provenance == nil {
		return false
	}
	binding := packet.Binding
	sources := binding.Sources
	return candidate.RuntimeID == execution.RuntimeID && candidate.BaselineGeneration == execution.BaselineGeneration &&
		candidate.ExecutionGeneration == execution.Generation && candidate.AssignmentID == execution.AssignmentID &&
		candidate.BaseCommit == execution.BaseCommit && candidate.SourceHead != "" &&
		candidate.SessionID == sources.SessionID && candidate.PlanID == sources.PlanID && candidate.ContractSHA256 == sources.ContractSHA256 &&
		candidate.Proposal.ResultID == packet.ResultID && candidate.Proposal.AssignmentID == sources.RepairAssignmentID && candidate.Proposal.ProducerAgentID == execution.AgentID &&
		provenance.Contract == "s9-repair-result-candidate-v1" && provenance.PacketID == packet.ID &&
		provenance.PacketSHA256 == packetSHA && provenance.IssuanceEventID == packet.IssuanceEventID &&
		provenance.OwnerAgentID == execution.AgentID && provenance.RuntimeID == execution.RuntimeID &&
		provenance.BaselineGeneration == execution.BaselineGeneration && provenance.ReviewRound == sources.ReviewRound &&
		provenance.ExecutionAssignmentID == execution.AssignmentID && provenance.ExecutionGeneration == execution.Generation &&
		provenance.ExecutionSHA256 == binding.ExecutionSHA256 && provenance.ExecutionRoot == execution.Path &&
		provenance.SourceHead == candidate.SourceHead && provenance.BaseCommit == execution.BaseCommit &&
		provenance.PlanReportRef == binding.PlanReportRef && provenance.PlanReportSHA256 == binding.PlanReportSHA256 &&
		provenance.ReworkRef == binding.ReworkRef && provenance.ReworkSHA256 == binding.ReworkSHA256
}

func repairResultMainProof(root string, execution workspace.Execution, checkpoint integration.Checkpoint) (runtime.RepairResultMainProof, error) {
	path := workspaceExecutionCheckpointPath(root, execution)
	relative, err := filepath.Rel(root, path)
	if err != nil || !filepath.IsLocal(relative) {
		return runtime.RepairResultMainProof{}, fmt.Errorf("S9 Main checkpoint escaped its registered slot")
	}
	data, err := workpacket.ReadFile(root, filepath.ToSlash(relative), 20<<20)
	if err != nil {
		return runtime.RepairResultMainProof{}, fmt.Errorf("read verified S9 Main checkpoint: %w", err)
	}
	checkpointHash := sha256.Sum256(data)
	facts, err := runtime.ProjectRepairResultMainFacts(data)
	if err != nil {
		return runtime.RepairResultMainProof{}, err
	}
	factsHash := sha256.Sum256(facts)
	if checkpoint.State != integration.StateVerified && checkpoint.State != integration.StateAcknowledged && checkpoint.State != integration.StateCleanupPending && checkpoint.State != integration.StateComplete {
		return runtime.RepairResultMainProof{}, fmt.Errorf("S9 Main checkpoint is not verified")
	}
	return runtime.RepairResultMainProof{
		CheckpointRef:       filepath.ToSlash(relative),
		CheckpointSHA256:    hex.EncodeToString(checkpointHash[:]),
		VerifiedFacts:       facts,
		VerifiedFactsSHA256: hex.EncodeToString(factsHash[:]),
	}, nil
}
