package repair

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/entroforge/go-system-builder/internal/agentauth"
	"github.com/entroforge/go-system-builder/internal/integration"
	runtimepkg "github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/schema"
	"github.com/entroforge/go-system-builder/internal/workpacket"
	"github.com/entroforge/go-system-builder/internal/workspace"
)

// AcceptIssuedRepairResultBlockerRequest names the issued S9 answer and its
// registered Worker source. AuthorizeSource remains an optional compatibility
// callback; the original packet, Workspace execution, source, owner, and CAS
// checks establish the domain facts when it is nil.
type AcceptIssuedRepairResultBlockerRequest struct {
	ExpectedRevision int
	PacketID         string
	AnswerPath       string
	OwnerAgentID     string
	SourceRoot       string
	AuthorizeSource  func(root string, state map[string]any) error
}

// IssuedRepairResultBlockerCapability is an opaque, one-transition capability
// for the original Agent lifecycle CAS. Its fields are private so callers
// cannot manufacture evidence IDs, answer bytes, source identity, or authority.
// It deliberately stores no context; every operation receives its caller's
// context and can reuse an outer workspace lock.
type IssuedRepairResultBlockerCapability struct {
	packet          workpacket.RepairResultPacket
	artifact        runtimepkg.RepairResultBlockerArtifact
	artifactRef     ArtifactRef
	artifactBytes   []byte
	messageRef      string
	messageBytes    []byte
	answer          json.RawMessage
	sourceRoot      string
	sourceHead      string
	owner           string
	taskID          string
	mainRoot        string
	beforeRevision  int
	generation      int
	authorizeSource func(root string, state map[string]any) error
}

// PrepareIssuedRepairResultBlocker verifies an issued packet and its complete
// answer, then stages immutable message and answer artifacts for the original
// AGENT-LIFECYCLE writer. A nil capability with a receipt means an exact
// previously accepted answer was found; no new Runtime mutation is implied.
func PrepareIssuedRepairResultBlocker(
	ctx context.Context,
	root, statePath, journalPath string,
	req AcceptIssuedRepairResultBlockerRequest,
) (*IssuedRepairResultBlockerCapability, *runtimepkg.RepairResultBlockerReceipt, error) {
	if ctx == nil {
		return nil, nil, errors.New("S9 blocker preparation requires the caller context")
	}
	lockedCtx, unlock, err := integration.LockWorkspace(ctx, root)
	if err != nil {
		return nil, nil, err
	}
	defer unlock()
	ctx = lockedCtx

	snapshot, _, err := readIssuedRepairResultRuntime(root, statePath, journalPath)
	if err != nil {
		return nil, nil, err
	}
	if mainRoot := workspaceMainRoot(root, snapshot.State); !sameAbsolutePath(mainRoot, root) {
		return nil, nil, errors.New("S9 blocker receiver root must be the original registered Main checkout")
	}
	packet, found, err := runtimepkg.FindRepairResultPacket(snapshot.State, req.PacketID)
	if err != nil || !found {
		return nil, nil, fmt.Errorf("S9 issued RepairResult packet %q is unavailable: %v", req.PacketID, err)
	}
	if req.OwnerAgentID == "" || req.OwnerAgentID != packet.Binding.Sources.OwnerAgentID {
		return nil, nil, errors.New("S9 blocker owner differs from the issued RepairAssignment")
	}
	if req.AnswerPath != packet.AnswerPath {
		return nil, nil, errors.New("S9 blocker answer path differs from the exact issued packet")
	}
	// Original accepted history is checked before revision, working-state, PLAN,
	// and current-source gates. A replay is still bound to the same author bytes
	// and source HEAD below; it never creates another lifecycle transition.
	existing, accepted, err := runtimepkg.VerifiedRepairResultBlockerReceipt(root, snapshot.State, packet.ID, req.OwnerAgentID)
	if err != nil {
		return nil, nil, err
	}
	if accepted {
		if req.AuthorizeSource != nil {
			if err := req.AuthorizeSource(root, snapshot.State); err != nil {
				return nil, nil, err
			}
		}
		if req.SourceRoot != existing.Artifact.ExecutionRoot {
			return nil, nil, runtimepkg.ErrSubmissionConflict
		}
		if err := verifyAcceptedBlockerBytes(ctx, root, snapshot.State, packet, existing, req.SourceRoot, existing.Artifact.SourceHead, existing.Artifact.SubmittedAnswer); err != nil {
			return nil, nil, err
		}
		return nil, &existing, nil
	}

	if req.ExpectedRevision < 0 || req.ExpectedRevision != snapshot.Revision {
		return nil, nil, runtimepkg.ErrStaleRevision
	}
	if err := rejectRepairResultBlockerOutcome(root, snapshot.State, packet, req.OwnerAgentID); err != nil {
		return nil, nil, err
	}
	if err := runtimepkg.VerifyRepairResultPacketIssueAndCurrent(root, snapshot.State, packet); err != nil {
		return nil, nil, err
	}
	if err := agentauth.ValidateWorkingAuthorization(root, snapshot.State, req.OwnerAgentID); err != nil {
		return nil, nil, err
	}
	if req.AuthorizeSource != nil {
		if err := req.AuthorizeSource(root, snapshot.State); err != nil {
			return nil, nil, err
		}
	}
	if req.SourceRoot == "" {
		return nil, nil, errors.New("S9 blocker requires the exact registered Worker execution root")
	}

	execution, head, err := verifyIssuedRepairResultBlockerSource(ctx, root, snapshot.State, packet, req.SourceRoot, "")
	if err != nil {
		return nil, nil, err
	}
	answer, answerSHA, err := readIssuedRepairResultBlockerAnswer(req.SourceRoot, packet)
	if err != nil {
		return nil, nil, err
	}
	issueAfter, issueAt, err := repairResultIssueCoordinates(root, snapshot.State, packet)
	if err != nil || issueAfter >= snapshot.Revision+1 {
		return nil, nil, fmt.Errorf("S9 blocker source issue coordinates are unavailable or inconsistent: %v", err)
	}
	packetSHA, err := workpacket.RepairResultPacketDigest(packet)
	if err != nil {
		return nil, nil, err
	}
	manifestAssignment, err := blockerManifestAssignment(packet)
	if err != nil {
		return nil, nil, err
	}
	b := packet.Binding.Sources
	base := path.Join(".claude/evidence", b.RuntimeID, fmt.Sprintf("g%d", b.BaselineGeneration), "S9")
	message := map[string]any{
		"schema_version":       "1.0.0",
		"message_type":         "blocker_report",
		"message_id":           blockerMessageID(packet.ID, answerSHA),
		"correlation_id":       blockerCorrelationID(packet.ID),
		"runtime_id":           b.RuntimeID,
		"agent_id":             b.OwnerAgentID,
		"agent_definition_ref": manifestAssignment["agent_definition_ref"],
		"task_id":              b.TaskID,
		"bug_id":               nil,
		"team_id":              b.TeamID,
		"occurred_at":          issueAt.UTC().Format(time.RFC3339Nano),
		"body":                 "S9 issued RepairResult answer reports a blocker; the complete professional answer is preserved with this message.",
		"s9_blocker":           map[string]any{"packet_id": packet.ID, "packet_sha256": packetSHA, "answer_path": packet.AnswerPath, "answer_sha256": answerSHA},
	}
	messageRef, messageBytes, err := writeBlockerImmutable(root, path.Join(base, "blocker-messages", packet.ID), "agent-message.schema.json", message)
	if err != nil {
		return nil, nil, err
	}
	artifact := runtimepkg.RepairResultBlockerArtifact{
		Contract: "s9-repair-result-blocker-v1", SchemaVersion: "1.0.0",
		PacketID: packet.ID, PacketSHA256: packetSHA, IssuanceEventID: packet.IssuanceEventID,
		Binding: packet.Binding, ExecutionAssignmentID: execution.AssignmentID,
		ExecutionGeneration: execution.Generation, ExecutionRoot: execution.Path,
		PlatformSessionID: execution.PlatformSessionID, BaseCommit: execution.BaseCommit,
		SourceHead: head, AnswerPath: packet.AnswerPath,
		SubmittedAnswer: append(json.RawMessage(nil), answer...), AnswerSHA256: answerSHA,
		MessageRef: messageRef.Path, MessageSHA256: messageRef.SHA256,
	}
	artifactRef, artifactBytes, err := writeBlockerImmutable(root, path.Join(base, "blockers", packet.ID), "repair-result-blocker.schema.json", artifact)
	if err != nil {
		return nil, nil, err
	}
	capability := &IssuedRepairResultBlockerCapability{
		packet: packet, artifact: artifact, artifactRef: artifactRef,
		artifactBytes: append([]byte(nil), artifactBytes...), messageRef: messageRef.Path,
		messageBytes: append([]byte(nil), messageBytes...), answer: append(json.RawMessage(nil), answer...),
		sourceRoot: execution.Path, sourceHead: head, owner: req.OwnerAgentID,
		taskID: b.TaskID, mainRoot: root, beforeRevision: snapshot.Revision,
		generation: b.BaselineGeneration, authorizeSource: req.AuthorizeSource,
	}
	return capability, nil, nil
}

// MessageRef returns the exact immutable blocker message path that the
// original Agent lifecycle request must consume.
func (capability *IssuedRepairResultBlockerCapability) MessageRef() string {
	if capability == nil {
		return ""
	}
	return capability.messageRef
}

// EvidenceIDs returns the only two original lifecycle evidence references:
// the immutable message and its complete answer artifact.
func (capability *IssuedRepairResultBlockerCapability) EvidenceIDs() []string {
	if capability == nil || capability.messageRef == "" || capability.artifactRef.Path == "" {
		return nil
	}
	return []string{capability.messageRef, capability.artifactRef.Path}
}

// ClearResolvedBlockerIndex permits the original Agent lifecycle CAS to clear
// an old resolution pointer only after ApplyLifecycleCAS has reverified the
// exact Resume predecessor. The referenced immutable history is retained.
func (capability *IssuedRepairResultBlockerCapability) ClearResolvedBlockerIndex() bool {
	return capability != nil && capability.packet.Binding.Resume != nil
}

// ValidateLifecycleRequest binds the pre-existing lifecycle entry to this
// capability. The lifecycle writer remains the sole owner of Agent state.
func (capability *IssuedRepairResultBlockerCapability) ValidateLifecycleRequest(root, owner, event, messagePath string, beforeRevision, generation int) error {
	if err := capability.validate(); err != nil {
		return err
	}
	if !sameAbsolutePath(root, capability.mainRoot) || owner != capability.owner || event != "work_blocked" ||
		!blockerRequestPathMatches(root, capability.messageRef, messagePath) || beforeRevision != capability.beforeRevision || generation != capability.generation {
		return runtimepkg.ErrSubmissionConflict
	}
	return nil
}

// AcceptedReplay proves that this exact already-accepted blocker can be
// returned by the original lifecycle API without requiring the Agent to be
// working again. It never accepts a changed author answer or Worker HEAD.
func (capability *IssuedRepairResultBlockerCapability) AcceptedReplay(ctx context.Context, root string, state map[string]any, owner, event, messagePath string) (bool, error) {
	if err := capability.validate(); err != nil {
		return false, err
	}
	if ctx == nil {
		return false, errors.New("S9 blocker replay requires the original operation context")
	}
	receipt, found, err := runtimepkg.VerifiedRepairResultBlockerReceipt(root, state, capability.packet.ID, owner)
	if err != nil || !found {
		return false, err
	}
	if !sameAbsolutePath(root, capability.mainRoot) || owner != capability.owner || event != "work_blocked" ||
		!blockerRequestPathMatches(root, capability.messageRef, messagePath) || receipt.ArtifactRef != capability.artifactRef.Path ||
		receipt.ArtifactSHA256 != capability.artifactRef.SHA256 || !reflect.DeepEqual(receipt.Artifact, capability.artifact) {
		return false, runtimepkg.ErrSubmissionConflict
	}
	if capability.authorizeSource != nil {
		if err := capability.authorizeSource(root, state); err != nil {
			return false, err
		}
	}
	if err := verifyAcceptedBlockerBytes(ctx, root, state, capability.packet, receipt, capability.sourceRoot, capability.sourceHead, capability.answer); err != nil {
		return false, err
	}
	return true, nil
}

// ApplyLifecycleCAS is called only from the original Agent Runtime mutation's
// Apply callback. It revalidates the issued packet, working PLAN, answer,
// source/session, immutable files, and exact pre-CAS revision before adding
// one typed evidence row. It does not write Runtime or mutate Agent fields.
func (capability *IssuedRepairResultBlockerCapability) ApplyLifecycleCAS(
	ctx context.Context,
	root string,
	state map[string]any,
	messageBytes []byte,
	owner, taskID string,
	beforeRevision, generation int,
) error {
	if err := capability.validate(); err != nil {
		return err
	}
	if ctx == nil {
		return errors.New("S9 blocker lifecycle Apply requires the original writer context")
	}
	lockedCtx, unlock, err := integration.LockWorkspace(ctx, root)
	if err != nil {
		return err
	}
	defer unlock()
	ctx = lockedCtx
	if !sameAbsolutePath(root, capability.mainRoot) || owner != capability.owner || taskID != capability.taskID ||
		beforeRevision != capability.beforeRevision || generation != capability.generation {
		return runtimepkg.ErrStaleRevision
	}
	if got, ok := state["revision"]; !ok || integerRepairResult(got) != beforeRevision {
		return runtimepkg.ErrStaleRevision
	}
	if state["runtime_id"] != capability.packet.Binding.Sources.RuntimeID {
		return errors.New("S9 blocker Runtime identity changed before the original lifecycle CAS")
	}
	baseline, _ := state["baseline"].(map[string]any)
	if integerRepairResult(baseline["generation"]) != generation {
		return errors.New("S9 blocker baseline generation changed before the original lifecycle CAS")
	}
	if err := agentauth.ValidateWorkingAuthorization(root, state, owner); err != nil {
		return err
	}
	if capability.authorizeSource != nil {
		if err := capability.authorizeSource(root, state); err != nil {
			return err
		}
	}
	packet, found, err := runtimepkg.FindRepairResultPacket(state, capability.packet.ID)
	if err != nil || !found || !sameRepairResultPacket(packet, capability.packet) {
		return fmt.Errorf("S9 blocker packet changed before original Agent lifecycle CAS: %v", err)
	}
	if err := runtimepkg.VerifyRepairResultPacketIssueAndCurrent(root, state, packet); err != nil {
		return err
	}
	agent, found, err := repairResultBlockerAgent(state, owner)
	if err != nil || !found {
		return fmt.Errorf("S9 blocker owner Agent is unavailable before original lifecycle CAS: %v", err)
	}
	resolvedRef := stringField(agent["blocker_resolved_ref"])
	if packet.Binding.Resume == nil {
		if resolvedRef != "" {
			return errors.New("S9 blocker owner retains a prior resolution index but this packet has no verified Resume lineage")
		}
	} else {
		if resolvedRef != packet.Binding.Resume.ResolutionMessageRef {
			return errors.New("S9 resumed blocker owner resolution index differs from its exact Resume predecessor")
		}
	}
	if err := rejectRepairResultBlockerOutcome(root, state, packet, owner); err != nil {
		return err
	}
	execution, head, err := verifyIssuedRepairResultBlockerSource(ctx, root, state, packet, capability.sourceRoot, capability.sourceHead)
	if err != nil {
		return err
	}
	if head != capability.sourceHead || execution.AssignmentID != capability.artifact.ExecutionAssignmentID || execution.Generation != capability.artifact.ExecutionGeneration || execution.PlatformSessionID != capability.artifact.PlatformSessionID || execution.BaseCommit != capability.artifact.BaseCommit {
		return errors.New("S9 blocker execution tuple changed before original Agent lifecycle CAS")
	}
	answer, answerSHA, err := readIssuedRepairResultBlockerAnswer(capability.sourceRoot, packet)
	if err != nil {
		return err
	}
	if answerSHA != capability.artifact.AnswerSHA256 || string(answer) != string(capability.answer) {
		return runtimepkg.ErrSubmissionConflict
	}
	if !reflect.DeepEqual(messageBytes, capability.messageBytes) {
		return runtimepkg.ErrSubmissionConflict
	}
	if err := verifyBlockerImmutableFiles(root, capability); err != nil {
		return err
	}
	if accepted, err := blockerReceiptAlreadyExists(root, state, packet, owner); err != nil {
		return err
	} else if accepted {
		return fmt.Errorf("S9 blocker was concurrently accepted; recover the exact original lifecycle event")
	}
	return appendRepairResultBlockerEvidence(state, packet, capability.artifactRef)
}

func (capability *IssuedRepairResultBlockerCapability) validate() error {
	if capability == nil || capability.packet.ID == "" || capability.owner == "" || capability.taskID == "" ||
		capability.mainRoot == "" || capability.sourceRoot == "" || capability.sourceHead == "" ||
		capability.messageRef == "" || len(capability.messageBytes) == 0 || len(capability.answer) == 0 ||
		capability.artifactRef.Path == "" || capability.artifactRef.SHA256 == "" ||
		len(capability.artifactBytes) == 0 || capability.beforeRevision < 0 || capability.generation < 1 {
		return errors.New("invalid or zero S9 RepairResult blocker capability")
	}
	return nil
}

func verifyAcceptedBlockerBytes(ctx context.Context, root string, state map[string]any, packet workpacket.RepairResultPacket, receipt runtimepkg.RepairResultBlockerReceipt, sourceRoot, expectedHead string, expectedAnswer json.RawMessage) error {
	stored, found, err := runtimepkg.FindRepairResultPacket(state, packet.ID)
	if err != nil || !found || !sameRepairResultPacket(stored, packet) {
		return fmt.Errorf("S9 blocker replay packet differs from original Runtime issuance: %v", err)
	}
	if receipt.Artifact.PacketID != packet.ID || receipt.Artifact.AnswerPath != packet.AnswerPath || !sameAbsolutePath(sourceRoot, receipt.Artifact.ExecutionRoot) {
		return runtimepkg.ErrSubmissionConflict
	}
	execution, head, err := verifyIssuedRepairResultBlockerSource(ctx, root, state, packet, sourceRoot, expectedHead)
	if err != nil || execution.AssignmentID != receipt.Artifact.ExecutionAssignmentID || execution.Generation != receipt.Artifact.ExecutionGeneration || execution.PlatformSessionID != receipt.Artifact.PlatformSessionID || execution.BaseCommit != receipt.Artifact.BaseCommit || head != receipt.Artifact.SourceHead {
		return fmt.Errorf("S9 exact blocker replay Worker execution/HEAD differs from its accepted source; preserve history and use explicit resolution/rework: %v", err)
	}
	answerBytes, err := workpacket.ReadFile(execution.Path, packet.AnswerPath, repairResultAnswerLimit)
	if err != nil {
		return err
	}
	answer, problems := workpacket.ValidateRepairResultAnswer(packet, answerBytes)
	if len(problems) != 0 || workpacket.Digest(answer) != receipt.Artifact.AnswerSHA256 ||
		string(answer) != string(receipt.Artifact.SubmittedAnswer) || string(answer) != string(expectedAnswer) {
		return runtimepkg.ErrSubmissionConflict
	}
	return nil
}

func verifyIssuedRepairResultBlockerSource(ctx context.Context, root string, state map[string]any, packet workpacket.RepairResultPacket, requestedRoot, expectedHead string) (workspace.Execution, string, error) {
	registry, execution, err := currentRepairExecution(ctx, root, state, packet, false)
	if err != nil {
		return workspace.Execution{}, "", err
	}
	canonicalRoot, err := workspace.Canonical(requestedRoot)
	if err != nil || canonicalRoot != execution.Path || execution.Path != requestedRoot {
		return workspace.Execution{}, "", errors.New("S9 blocker source root differs from its exact canonical registered Worker checkout")
	}
	var frozen map[string]any
	if err := json.Unmarshal(packet.Inputs.Execution, &frozen); err != nil || frozen == nil {
		return workspace.Execution{}, "", errors.New("S9 blocker packet has an invalid frozen execution definition")
	}
	currentBytes, err := json.Marshal(execution)
	if err != nil {
		return workspace.Execution{}, "", err
	}
	var current map[string]any
	if err := json.Unmarshal(currentBytes, &current); err != nil || current == nil {
		return workspace.Execution{}, "", errors.New("S9 current Workspace execution cannot be projected")
	}
	projection, err := runtimepkg.ProjectRepairResultExecution(current)
	if err != nil || workpacket.Digest(projection) != packet.Binding.ExecutionSHA256 {
		return workspace.Execution{}, "", fmt.Errorf("S9 registered Worker execution definition differs from its packet: %v", err)
	}
	if frozen["assignment_id"] != execution.AssignmentID || execution.RuntimeID != packet.Binding.Sources.RuntimeID ||
		execution.BaselineGeneration != packet.Binding.Sources.BaselineGeneration || execution.AgentID != packet.Binding.Sources.OwnerAgentID ||
		execution.PlatformSessionID == "" || frozen["platform_session_id"] != execution.PlatformSessionID ||
		frozen["base_commit"] != execution.BaseCommit || frozen["worktree_path"] != execution.Path ||
		integerRepairResult(frozen["execution_generation"]) != execution.Generation {
		return workspace.Execution{}, "", errors.New("S9 blocker current execution owner/session/base tuple differs from the frozen assignment")
	}
	if registry == nil {
		return workspace.Execution{}, "", errors.New("S9 blocker has no original registered Workspace registry")
	}
	head, err := workspace.Git(ctx, execution.Path, "rev-parse", "HEAD")
	if err != nil || !validBlockerCommit(head) {
		return workspace.Execution{}, "", fmt.Errorf("read S9 blocker Worker source HEAD: %v", err)
	}
	if expectedHead != "" && head != expectedHead {
		return workspace.Execution{}, "", errors.New("S9 blocker Worker source HEAD changed before the original lifecycle CAS")
	}
	return execution, head, nil
}

func readIssuedRepairResultBlockerAnswer(sourceRoot string, packet workpacket.RepairResultPacket) (json.RawMessage, string, error) {
	data, err := workpacket.ReadFile(sourceRoot, packet.AnswerPath, repairResultAnswerLimit)
	if err != nil {
		return nil, "", err
	}
	answer, problems := workpacket.ValidateRepairResultAnswer(packet, data)
	if len(problems) != 0 {
		return nil, "", fmt.Errorf("S9 blocker answer %s: %s (%s)", problems[0].Pointer, problems[0].Expected, problems[0].Observed)
	}
	var professional workpacket.RepairResultProfessionalAnswer
	if err := json.Unmarshal(answer, &professional); err != nil {
		return nil, "", err
	}
	switch professional.Verdict {
	case "blocked", "contract_conflict", "stale":
	default:
		return nil, "", fmt.Errorf("S9 blocker receipt only preserves blocked, contract_conflict, or stale professional verdicts")
	}
	return answer, workpacket.Digest(answer), nil
}

func rejectRepairResultBlockerOutcome(root string, state map[string]any, packet workpacket.RepairResultPacket, owner string) error {
	if _, found, err := runtimepkg.VerifiedRepairResultCandidateReceipt(root, state, packet.ID, owner); err != nil {
		return err
	} else if found {
		return errors.New("S9 packet already has an accepted WorkspaceCandidate; blocker receipt would conflict")
	}
	if _, found, err := runtimepkg.VerifiedRepairResultReceipt(root, state, packet.ID, owner); err != nil {
		return err
	} else if found {
		return errors.New("S9 packet already has an accepted Main Result; blocker receipt would conflict")
	}
	return nil
}

func blockerReceiptAlreadyExists(root string, state map[string]any, packet workpacket.RepairResultPacket, owner string) (bool, error) {
	_, found, err := runtimepkg.VerifiedRepairResultBlockerReceipt(root, state, packet.ID, owner)
	return found, err
}

func appendRepairResultBlockerEvidence(state map[string]any, packet workpacket.RepairResultPacket, ref ArtifactRef) error {
	rows, ok := state["evidence"].([]any)
	if !ok {
		return errors.New("Runtime evidence index is unavailable for the original blocker lifecycle CAS")
	}
	for _, raw := range rows {
		row, _ := raw.(map[string]any)
		if row != nil && (row["id"] == ref.Path || row["path"] == ref.Path) {
			return fmt.Errorf("S9 blocker evidence path already exists; recover or inspect the original lifecycle event")
		}
	}
	b := packet.Binding.Sources
	row := map[string]any{
		"id": ref.Path, "kind": "repair_result_blocker", "path": ref.Path, "sha256": ref.SHA256,
		"status": "valid", "baseline_generation": b.BaselineGeneration, "review_round": b.ReviewRound,
		"produced_by": []any{b.OwnerAgentID}, "invalidated_by": nil, "invalidation_rule": nil,
		"invalidation_reason": nil, "responsibility_id": "BUILD-WORK-PACKAGE",
		"scope_refs": []any{packet.ID, b.RepairAssignmentID},
	}
	state["evidence"] = append(rows, row)
	return nil
}

func verifyBlockerImmutableFiles(root string, capability *IssuedRepairResultBlockerCapability) error {
	message, err := workpacket.ReadFile(root, capability.messageRef, 1<<20)
	if err != nil || !reflect.DeepEqual(message, capability.messageBytes) || workpacket.Digest(message) != capability.artifact.MessageSHA256 {
		return fmt.Errorf("S9 immutable blocker message changed before original lifecycle CAS: %v", err)
	}
	if err := schema.NewEmbeddedValidator().ValidateBytes("agent-message.schema.json", message); err != nil {
		return err
	}
	artifact, err := workpacket.ReadFile(root, capability.artifactRef.Path, 20<<20)
	if err != nil || !reflect.DeepEqual(artifact, capability.artifactBytes) || workpacket.Digest(artifact) != capability.artifactRef.SHA256 {
		return fmt.Errorf("S9 immutable blocker artifact changed before original lifecycle CAS: %v", err)
	}
	if err := schema.NewEmbeddedValidator().ValidateBytes("repair-result-blocker.schema.json", artifact); err != nil {
		return err
	}
	var decoded runtimepkg.RepairResultBlockerArtifact
	if err := json.Unmarshal(artifact, &decoded); err != nil || !reflect.DeepEqual(decoded, capability.artifact) {
		return fmt.Errorf("S9 immutable blocker artifact no longer matches its capability: %v", err)
	}
	return nil
}

func blockerManifestAssignment(packet workpacket.RepairResultPacket) (map[string]any, error) {
	var assignment map[string]any
	if err := json.Unmarshal(packet.Inputs.Sources.ManifestAssignment, &assignment); err != nil || assignment == nil {
		return nil, errors.New("S9 blocker packet has no valid frozen manifest Assignment")
	}
	if stringField(assignment["agent_definition_ref"]) == "" || stringField(assignment["agent_id"]) != packet.Binding.Sources.OwnerAgentID {
		return nil, errors.New("S9 blocker frozen manifest Assignment does not name the issued owner/Agent definition")
	}
	return assignment, nil
}

func blockerMessageID(packetID, answerSHA string) string {
	return "msg-s9-blocker-" + strings.TrimPrefix(packetID, "packet-") + "-" + answerSHA
}

func blockerCorrelationID(packetID string) string {
	return "corr-s9-blocker-" + strings.TrimPrefix(packetID, "packet-")
}

func validBlockerCommit(value string) bool {
	if len(value) != 40 && len(value) != 64 {
		return false
	}
	for _, c := range value {
		if !('0' <= c && c <= '9') && !('a' <= c && c <= 'f') {
			return false
		}
	}
	return true
}

func blockerRequestPathMatches(root, ref, supplied string) bool {
	if supplied == ref {
		return true
	}
	want, err := repositoryPath(root, ref)
	if err != nil || !filepath.IsAbs(supplied) {
		return false
	}
	got, err := filepath.Abs(supplied)
	if err != nil {
		return false
	}
	return filepath.Clean(got) == filepath.Clean(want)
}

func writeBlockerImmutable(root, directory, schemaName string, document any) (ArtifactRef, []byte, error) {
	data, err := canonicalJSON(document)
	if err != nil {
		return ArtifactRef{}, nil, err
	}
	limit := int64(20 << 20)
	if schemaName == "agent-message.schema.json" {
		limit = 1 << 20
	}
	if int64(len(data)) > limit {
		return ArtifactRef{}, nil, fmt.Errorf("S9 immutable blocker %s exceeds %d bytes", schemaName, limit)
	}
	if err := schema.NewEmbeddedValidator().ValidateBytes(schemaName, data); err != nil {
		return ArtifactRef{}, nil, fmt.Errorf("validate immutable S9 blocker artifact: %w", err)
	}
	sha := sha256Bytes(data)
	ref := ArtifactRef{Path: path.Join(directory, sha+".json"), SHA256: sha}
	if existing, readErr := workpacket.ReadFile(root, ref.Path, limit); readErr == nil {
		if !reflect.DeepEqual(existing, data) || sha256Bytes(existing) != ref.SHA256 {
			return ArtifactRef{}, nil, runtimepkg.ErrSubmissionConflict
		}
		if err := syncBlockerArtifactParent(root, ref.Path); err != nil {
			return ArtifactRef{}, nil, err
		}
		return ref, existing, nil
	} else if !errors.Is(readErr, os.ErrNotExist) {
		return ArtifactRef{}, nil, readErr
	}
	written, err := writeImmutable(root, ref.Path, schemaName, document)
	if err != nil {
		if !errors.Is(err, os.ErrExist) {
			return ArtifactRef{}, nil, err
		}
		// Another exact retry may have created the same content-addressed file
		// after our read. Reuse it only after a bounded no-symlink byte check.
		existing, readErr := workpacket.ReadFile(root, ref.Path, limit)
		if readErr == nil && reflect.DeepEqual(existing, data) && sha256Bytes(existing) == ref.SHA256 {
			if syncErr := syncBlockerArtifactParent(root, ref.Path); syncErr != nil {
				return ArtifactRef{}, nil, syncErr
			}
			return ref, existing, nil
		}
		return ArtifactRef{}, nil, err
	}
	if written != ref {
		return ArtifactRef{}, nil, errors.New("immutable blocker writer returned a different content-addressed reference")
	}
	return written, data, nil
}

func syncBlockerArtifactParent(root, relative string) error {
	filePath, err := repositoryPath(root, relative)
	if err != nil {
		return err
	}
	return runtimepkg.SyncDirectory(filepath.Dir(filePath))
}
