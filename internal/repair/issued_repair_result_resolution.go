package repair

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/entroforge/go-system-builder/internal/agentauth"
	"github.com/entroforge/go-system-builder/internal/integration"
	runtimepkg "github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/schema"
	"github.com/entroforge/go-system-builder/internal/workpacket"
	"github.com/entroforge/go-system-builder/internal/workstate"
)

const repairResultResolutionMessageLimit = 1 << 20

// PrepareIssuedRepairResultResolutionRequest names the exact Main resolution
// file. AuthorizeMain is an optional compatibility callback; the original
// packet/blocker and resolution validators own domain acceptance.
type PrepareIssuedRepairResultResolutionRequest struct {
	ExpectedRevision int
	PacketID         string
	OwnerAgentID     string
	AuthoringPath    string
	AuthorizeMain    func(root string, state map[string]any) error
}

// IssuedRepairResultResolutionCapability is an opaque, single-use proof for
// the original Agent lifecycle CAS. Its unexported fields prevent JSON or CLI
// inputs from manufacturing Main authority, evidence IDs, or message bytes.
type IssuedRepairResultResolutionCapability struct {
	packet         workpacket.RepairResultPacket
	blocker        runtimepkg.RepairResultBlockerReceipt
	messageRef     string
	messageSHA256  string
	messageBytes   []byte
	authoringPath  string
	owner          string
	taskID         string
	mainRoot       string
	beforeRevision int
	generation     int
	authorizeMain  func(root string, state map[string]any) error
}

// RepairResultResolutionAuthoringPath derives Main's one authoring path from
// the registered packet's exact <owner>/answer.json path.
func RepairResultResolutionAuthoringPath(packet workpacket.RepairResultPacket) (string, error) {
	return workpacket.RepairResultResolutionAuthoringPath(packet)
}

// RepairResultResolutionDraft returns an incomplete Main authoring envelope
// for the exact currently unresolved blocker. It does not write the file or
// authorize lifecycle transitions. The empty rationale marks the draft as
// incomplete until Main supplies its decision rationale.
func RepairResultResolutionDraft(root string, state map[string]any, packetID, owner string) (string, []byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	packet, found, err := runtimepkg.FindRepairResultPacket(state, packetID)
	if err != nil || !found {
		return "", nil, fmt.Errorf("S9 resolution packet %q is unavailable: %v", packetID, err)
	}
	path, err := RepairResultResolutionAuthoringPath(packet)
	if err != nil {
		return "", nil, err
	}
	if owner == "" || owner != packet.Binding.Sources.OwnerAgentID {
		return "", nil, errors.New("S9 resolution owner differs from the issued RepairAssignment")
	}
	if !sameAbsolutePath(root, workspaceMainRoot(root, state)) {
		return "", nil, errors.New("S9 resolution draft requires the original Main checkout")
	}
	blocker, err := validateCurrentRepairResultResolutionBlocker(ctx, root, state, packet, owner, nil)
	if err != nil {
		return "", nil, err
	}
	if err := rejectContractConflictResolution(blocker); err != nil {
		return "", nil, err
	}
	messageBytes, err := workpacket.ReadFile(root, blocker.Artifact.MessageRef, repairResultResolutionMessageLimit)
	if err != nil {
		return "", nil, fmt.Errorf("read original immutable S9 blocker message: %w", err)
	}
	if workpacket.Digest(messageBytes) != blocker.Artifact.MessageSHA256 {
		return "", nil, errors.New("original immutable S9 blocker message differs from its accepted SHA")
	}
	if err := schema.NewEmbeddedValidator().ValidateBytes("agent-message.schema.json", messageBytes); err != nil {
		return "", nil, fmt.Errorf("validate original immutable S9 blocker message: %w", err)
	}
	var original map[string]any
	if err := json.Unmarshal(messageBytes, &original); err != nil || original == nil {
		return "", nil, fmt.Errorf("decode original immutable S9 blocker message: %v", err)
	}
	draft, err := buildRepairResultResolutionEnvelope(original, packet, blocker, "", []any{}, "")
	if err != nil {
		return "", nil, err
	}
	data, err := canonicalJSON(draft)
	if err != nil {
		return "", nil, err
	}
	return path, data, nil
}

// PrepareIssuedRepairResultResolution validates the current Main authoring
// bytes against the accepted blocker and stages an exact immutable message.
// The returned capability is consumed only by Agent's original lifecycle CAS.
// A nil capability with a receipt is an exact historical replay and does not
// request another Runtime mutation.
func PrepareIssuedRepairResultResolution(
	ctx context.Context,
	root, statePath, journalPath string,
	req PrepareIssuedRepairResultResolutionRequest,
) (*IssuedRepairResultResolutionCapability, *runtimepkg.RepairResultResolutionReceipt, error) {
	if ctx == nil {
		return nil, nil, errors.New("S9 resolution preparation requires the caller context")
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
	if !sameAbsolutePath(root, workspaceMainRoot(root, snapshot.State)) {
		return nil, nil, errors.New("S9 resolution receiver root must be the original registered Main checkout")
	}
	packet, found, err := runtimepkg.FindRepairResultPacket(snapshot.State, req.PacketID)
	if err != nil || !found {
		return nil, nil, fmt.Errorf("S9 issued RepairResult packet %q is unavailable: %v", req.PacketID, err)
	}
	if req.OwnerAgentID == "" || req.OwnerAgentID != packet.Binding.Sources.OwnerAgentID {
		return nil, nil, errors.New("S9 resolution owner differs from the issued RepairAssignment")
	}
	expectedPath, err := RepairResultResolutionAuthoringPath(packet)
	if err != nil {
		return nil, nil, err
	}
	if req.AuthoringPath != expectedPath {
		return nil, nil, errors.New("S9 resolution path differs from the exact current Main authoring path")
	}
	if req.AuthorizeMain != nil {
		if err := req.AuthorizeMain(root, snapshot.State); err != nil {
			return nil, nil, err
		}
	}

	// History is checked before current state and expected revision so an exact
	// accepted resolution can be replayed without another CAS.
	existing, accepted, err := runtimepkg.VerifiedRepairResultResolutionReceipt(root, snapshot.State, packet.ID, req.OwnerAgentID)
	if err != nil {
		return nil, nil, err
	}
	if accepted {
		if err := verifyExactRepairResultResolutionReplay(root, req.AuthoringPath, packet, existing); err != nil {
			return nil, nil, err
		}
		return nil, &existing, nil
	}
	if req.ExpectedRevision < 0 || req.ExpectedRevision != snapshot.Revision {
		return nil, nil, runtimepkg.ErrStaleRevision
	}
	blocker, err := validateCurrentRepairResultResolutionBlocker(ctx, root, snapshot.State, packet, req.OwnerAgentID, nil)
	if err != nil {
		return nil, nil, err
	}
	if err := rejectContractConflictResolution(blocker); err != nil {
		return nil, nil, err
	}

	authoringBytes, err := workpacket.ReadFile(root, req.AuthoringPath, repairResultResolutionMessageLimit)
	if err != nil {
		return nil, nil, fmt.Errorf("read exact S9 Main resolution authoring file: %w", err)
	}
	originalMessage, err := readAcceptedBlockerMessage(root, blocker)
	if err != nil {
		return nil, nil, err
	}
	if err := validateRepairResultResolutionAuthoring(authoringBytes, originalMessage, packet, blocker); err != nil {
		return nil, nil, err
	}
	messageRef, storedBytes, err := publishRepairResultResolutionMessage(root, packet, authoringBytes)
	if err != nil {
		return nil, nil, err
	}
	messageSHA := workpacket.Digest(storedBytes)
	if messageSHA == "" {
		return nil, nil, errors.New("S9 resolution immutable message has no digest")
	}
	binding := packet.Binding.Sources
	capability := &IssuedRepairResultResolutionCapability{
		packet: packet, blocker: blocker, messageRef: messageRef, messageSHA256: messageSHA,
		messageBytes: append([]byte(nil), storedBytes...), authoringPath: req.AuthoringPath,
		owner: req.OwnerAgentID, taskID: binding.TaskID, mainRoot: root,
		beforeRevision: snapshot.Revision, generation: binding.BaselineGeneration,
		authorizeMain: req.AuthorizeMain,
	}
	return capability, nil, nil
}

// MessageRef returns the immutable Main resolution message consumed by the
// original blocker_resolved lifecycle transition.
func (capability *IssuedRepairResultResolutionCapability) MessageRef() string {
	if capability == nil {
		return ""
	}
	return capability.messageRef
}

// EvidenceIDs binds the new message and its exact predecessor blocker in the
// original AGENT-LIFECYCLE event.
func (capability *IssuedRepairResultResolutionCapability) EvidenceIDs() []string {
	if capability == nil || capability.messageRef == "" || capability.blocker.ArtifactRef == "" {
		return nil
	}
	return []string{capability.messageRef, capability.blocker.ArtifactRef}
}

// ValidateLifecycleRequest binds the capability to one precise original CAS.
func (capability *IssuedRepairResultResolutionCapability) ValidateLifecycleRequest(root, owner, event, messagePath string, beforeRevision, generation int) error {
	if err := capability.validate(); err != nil {
		return err
	}
	if !sameAbsolutePath(root, capability.mainRoot) || owner != capability.owner || event != "blocker_resolved" ||
		!blockerRequestPathMatches(root, capability.messageRef, messagePath) || beforeRevision != capability.beforeRevision || generation != capability.generation {
		return runtimepkg.ErrSubmissionConflict
	}
	return nil
}

// AcceptedReplay returns true only for the same accepted message bytes and
// receipt. It intentionally does not require the now-resolved packet to stay
// current or the Worker source to remain unchanged.
func (capability *IssuedRepairResultResolutionCapability) AcceptedReplay(ctx context.Context, root string, state map[string]any, owner, event, messagePath string) (bool, error) {
	if err := capability.validate(); err != nil {
		return false, err
	}
	if ctx == nil {
		return false, errors.New("S9 resolution replay requires the original writer context")
	}
	if !sameAbsolutePath(root, capability.mainRoot) || owner != capability.owner || event != "blocker_resolved" ||
		!blockerRequestPathMatches(root, capability.messageRef, messagePath) {
		return false, runtimepkg.ErrSubmissionConflict
	}
	if capability.authorizeMain != nil {
		if err := capability.authorizeMain(root, state); err != nil {
			return false, err
		}
	}
	receipt, found, err := runtimepkg.VerifiedRepairResultResolutionReceipt(root, state, capability.packet.ID, owner)
	if err != nil || !found {
		return false, err
	}
	if receipt.BlockerRef != capability.blocker.ArtifactRef || receipt.BlockerSHA256 != capability.blocker.ArtifactSHA256 ||
		receipt.MessageRef != capability.messageRef || receipt.MessageSHA256 != capability.messageSHA256 {
		return false, runtimepkg.ErrSubmissionConflict
	}
	if err := verifyCapabilityResolutionBytes(root, capability); err != nil {
		return false, err
	}
	return true, nil
}

// ApplyLifecycleCAS is invoked from the original Agent Writer's Apply callback.
// It rechecks the current blocked PLAN/source/attempt, predecessor receipt and
// exact saved bytes before appending the evidence row.
func (capability *IssuedRepairResultResolutionCapability) ApplyLifecycleCAS(
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
		return errors.New("S9 resolution lifecycle Apply requires the original writer context")
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
		return errors.New("S9 resolution Runtime identity changed before the original lifecycle CAS")
	}
	baseline, _ := state["baseline"].(map[string]any)
	if integerRepairResult(baseline["generation"]) != generation {
		return errors.New("S9 resolution baseline generation changed before the original lifecycle CAS")
	}
	if capability.authorizeMain != nil {
		if err := capability.authorizeMain(root, state); err != nil {
			return err
		}
	}
	if !reflect.DeepEqual(messageBytes, capability.messageBytes) {
		return runtimepkg.ErrSubmissionConflict
	}
	if err := verifyCapabilityResolutionBytes(root, capability); err != nil {
		return err
	}
	packet, found, err := runtimepkg.FindRepairResultPacket(state, capability.packet.ID)
	if err != nil || !found || !sameRepairResultPacket(packet, capability.packet) {
		return fmt.Errorf("S9 resolution packet changed before original Agent lifecycle CAS: %v", err)
	}
	blocker, err := validateCurrentRepairResultResolutionBlocker(ctx, root, state, packet, owner, &capability.blocker)
	if err != nil {
		return err
	}
	if err := rejectContractConflictResolution(blocker); err != nil {
		return err
	}
	if accepted, err := resolutionReceiptAlreadyExists(root, state, packet, owner); err != nil {
		return err
	} else if accepted {
		return errors.New("S9 resolution was concurrently accepted; recover the exact original lifecycle event")
	}
	return appendRepairResultResolutionEvidence(state, packet, blocker, ArtifactRef{Path: capability.messageRef, SHA256: capability.messageSHA256})
}

func (capability *IssuedRepairResultResolutionCapability) validate() error {
	if capability == nil || capability.packet.ID == "" || capability.owner == "" || capability.taskID == "" ||
		capability.mainRoot == "" || capability.messageRef == "" || capability.messageSHA256 == "" ||
		len(capability.messageBytes) == 0 || capability.authoringPath == "" || capability.blocker.ArtifactRef == "" ||
		capability.blocker.ArtifactSHA256 == "" || capability.beforeRevision < 0 || capability.generation < 1 {
		return errors.New("invalid or zero S9 RepairResult resolution capability")
	}
	return nil
}

func validateCurrentRepairResultResolutionBlocker(
	ctx context.Context,
	root string,
	state map[string]any,
	packet workpacket.RepairResultPacket,
	owner string,
	expected *runtimepkg.RepairResultBlockerReceipt,
) (runtimepkg.RepairResultBlockerReceipt, error) {
	if ctx == nil {
		return runtimepkg.RepairResultBlockerReceipt{}, errors.New("S9 resolution requires the original caller context")
	}
	if !runtimepkg.PacketEnabled(state) || !sameAbsolutePath(root, workspaceMainRoot(root, state)) {
		return runtimepkg.RepairResultBlockerReceipt{}, errors.New("S9 resolution requires the original packet-mode Main Runtime")
	}
	if owner == "" || owner != packet.Binding.Sources.OwnerAgentID {
		return runtimepkg.RepairResultBlockerReceipt{}, errors.New("S9 resolution owner differs from the issued RepairAssignment")
	}
	stored, found, err := runtimepkg.FindRepairResultPacket(state, packet.ID)
	if err != nil || !found || !sameRepairResultPacket(stored, packet) {
		return runtimepkg.RepairResultBlockerReceipt{}, fmt.Errorf("S9 resolution packet differs from its exact Runtime issuance: %v", err)
	}
	blocker, accepted, err := runtimepkg.VerifiedRepairResultBlockerReceipt(root, state, packet.ID, owner)
	if err != nil {
		return runtimepkg.RepairResultBlockerReceipt{}, err
	}
	if !accepted {
		return runtimepkg.RepairResultBlockerReceipt{}, errors.New("S9 resolution requires the original accepted typed blocker receipt")
	}
	if expected != nil && !reflect.DeepEqual(blocker, *expected) {
		return runtimepkg.RepairResultBlockerReceipt{}, runtimepkg.ErrSubmissionConflict
	}
	if _, resolved, err := runtimepkg.VerifiedRepairResultResolutionReceipt(root, state, packet.ID, owner); err != nil {
		return runtimepkg.RepairResultBlockerReceipt{}, err
	} else if resolved {
		return runtimepkg.RepairResultBlockerReceipt{}, runtimepkg.ErrSubmissionConflict
	}
	agent, found, err := repairResultBlockerAgent(state, owner)
	if err != nil || !found {
		return runtimepkg.RepairResultBlockerReceipt{}, fmt.Errorf("S9 resolution owner Agent is unavailable: %v", err)
	}
	projection, err := workstate.AgentProjection(agent, fmt.Sprint(state["schema_version"]))
	if err != nil || projection.Status != workstate.Blocked || projection.Activity != "blocked" {
		return runtimepkg.RepairResultBlockerReceipt{}, fmt.Errorf("S9 resolution requires the exact blocked owner state: %v", err)
	}
	lifecycleEvent := stringField(agent["last_lifecycle_event"])
	if (workstate.Unified(fmt.Sprint(state["schema_version"])) && lifecycleEvent != "work_blocked") || (lifecycleEvent != "" && lifecycleEvent != "work_blocked") ||
		stringField(agent["work_blocked_ref"]) != blocker.Artifact.MessageRef || stringField(agent["blocker_resolved_ref"]) != "" {
		return runtimepkg.RepairResultBlockerReceipt{}, errors.New("S9 resolution owner is not blocked on this exact unresolved blocker message")
	}
	if !currentRepairResultBlockerEvidence(state, blocker.ArtifactRef) {
		return runtimepkg.RepairResultBlockerReceipt{}, errors.New("S9 resolution predecessor blocker evidence is invalidated or not current")
	}
	if err := agentauth.ValidateBlockedPlanAuthorization(root, state, owner); err != nil {
		return runtimepkg.RepairResultBlockerReceipt{}, fmt.Errorf("S9 resolution has no current original blocked PLAN authorization: %w", err)
	}
	current, err := runtimepkg.CurrentRepairResultPackets(root, state, owner)
	if err != nil {
		return runtimepkg.RepairResultBlockerReceipt{}, err
	}
	if len(current) != 1 || current[0].ID != packet.ID {
		return runtimepkg.RepairResultBlockerReceipt{}, errors.New("S9 resolution packet is not the unique current blocker attempt")
	}
	if err := runtimepkg.VerifyRepairResultPacketIssueAndCurrent(root, state, packet); err != nil {
		return runtimepkg.RepairResultBlockerReceipt{}, err
	}
	execution, head, err := verifyIssuedRepairResultBlockerSource(ctx, root, state, packet, blocker.Artifact.ExecutionRoot, blocker.Artifact.SourceHead)
	if err != nil {
		return runtimepkg.RepairResultBlockerReceipt{}, err
	}
	if head != blocker.Artifact.SourceHead || execution.AssignmentID != blocker.Artifact.ExecutionAssignmentID ||
		execution.Generation != blocker.Artifact.ExecutionGeneration || execution.PlatformSessionID != blocker.Artifact.PlatformSessionID ||
		execution.Path != blocker.Artifact.ExecutionRoot || execution.BaseCommit != blocker.Artifact.BaseCommit {
		return runtimepkg.RepairResultBlockerReceipt{}, errors.New("S9 resolution Worker execution/attempt differs from the accepted blocker tuple")
	}
	answerBytes, err := workpacket.ReadFile(execution.Path, packet.AnswerPath, repairResultAnswerLimit)
	if err != nil {
		return runtimepkg.RepairResultBlockerReceipt{}, err
	}
	answer, problems := workpacket.ValidateRepairResultAnswer(packet, answerBytes)
	if len(problems) != 0 || workpacket.Digest(answer) != blocker.Artifact.AnswerSHA256 || !bytes.Equal(answer, blocker.Artifact.SubmittedAnswer) {
		return runtimepkg.RepairResultBlockerReceipt{}, runtimepkg.ErrSubmissionConflict
	}
	return blocker, nil
}

func rejectContractConflictResolution(blocker runtimepkg.RepairResultBlockerReceipt) error {
	var answer workpacket.RepairResultProfessionalAnswer
	if err := json.Unmarshal(blocker.Artifact.SubmittedAnswer, &answer); err != nil {
		return fmt.Errorf("decode accepted S9 blocker verdict: %w", err)
	}
	if answer.Verdict == "contract_conflict" {
		return errors.New("S9 contract_conflict cannot use retry on the same Contract/Session/Plan; resolve through the original S8 human Contract-approval gate")
	}
	return nil
}

func buildRepairResultResolutionEnvelope(original map[string]any, packet workpacket.RepairResultPacket, blocker runtimepkg.RepairResultBlockerReceipt, body string, evidenceRefs []any, rationale string) (map[string]any, error) {
	if original == nil {
		return nil, errors.New("S9 resolution requires the verified original blocker message envelope")
	}
	packetSHA, err := workpacket.RepairResultPacketDigest(packet)
	if err != nil {
		return nil, err
	}
	message := make(map[string]any, len(original)+1)
	for key, value := range original {
		message[key] = value
	}
	delete(message, "s9_blocker")
	message["message_type"] = "blocker_resolution"
	message["message_id"] = resolutionMessageID(packet.ID, blocker.ArtifactSHA256)
	message["body"] = body
	message["s9_resolution"] = map[string]any{
		"decision": "retry", "rationale": rationale, "evidence_refs": evidenceRefs,
		"packet_id": packet.ID, "packet_sha256": packetSHA,
		"blocker_ref": blocker.ArtifactRef, "blocker_sha256": blocker.ArtifactSHA256,
		"execution_assignment_id": blocker.Artifact.ExecutionAssignmentID,
		"execution_generation":    blocker.Artifact.ExecutionGeneration,
		"execution_sha256":        packet.Binding.ExecutionSHA256,
	}
	return message, nil
}

func validateRepairResultResolutionAuthoring(data, originalBytes []byte, packet workpacket.RepairResultPacket, blocker runtimepkg.RepairResultBlockerReceipt) error {
	if len(data) == 0 || len(data) > repairResultResolutionMessageLimit {
		return errors.New("S9 Main resolution authoring file is empty or exceeds 1 MiB")
	}
	if err := workpacket.ValidateUniqueJSON(data); err != nil {
		return fmt.Errorf("S9 Main resolution JSON: %w", err)
	}
	if err := schema.NewEmbeddedValidator().ValidateBytes("agent-message.schema.json", data); err != nil {
		return fmt.Errorf("S9 Main resolution message schema: %w", err)
	}
	var original map[string]any
	if err := json.Unmarshal(originalBytes, &original); err != nil || original == nil {
		return fmt.Errorf("decode accepted S9 blocker message: %v", err)
	}
	var actual map[string]any
	if err := json.Unmarshal(data, &actual); err != nil || actual == nil {
		return fmt.Errorf("decode S9 Main resolution message: %v", err)
	}
	body, ok := actual["body"].(string)
	if !ok {
		return errors.New("S9 Main resolution body must retain its original string field")
	}
	actualResolution, ok := actual["s9_resolution"].(map[string]any)
	if !ok {
		return errors.New("S9 Main resolution lacks its closed s9_resolution object")
	}
	rationale, ok := actualResolution["rationale"].(string)
	if !ok {
		return errors.New("S9 Main resolution rationale must retain its schema-defined string field")
	}
	evidenceRefs, ok := actualResolution["evidence_refs"].([]any)
	if !ok {
		return errors.New("S9 Main resolution evidence_refs must be an array")
	}
	expected, err := buildRepairResultResolutionEnvelope(original, packet, blocker, body, evidenceRefs, rationale)
	if err != nil {
		return err
	}
	actualBytes, err := json.Marshal(actual)
	if err != nil {
		return err
	}
	expectedBytes, err := json.Marshal(expected)
	if err != nil {
		return err
	}
	if !bytes.Equal(actualBytes, expectedBytes) {
		return errors.New("S9 Main resolution changed a machine-bound blocker, packet, owner, or execution field")
	}
	return nil
}

func readAcceptedBlockerMessage(root string, blocker runtimepkg.RepairResultBlockerReceipt) ([]byte, error) {
	message, err := workpacket.ReadFile(root, blocker.Artifact.MessageRef, repairResultResolutionMessageLimit)
	if err != nil {
		return nil, fmt.Errorf("read accepted S9 blocker lifecycle message: %w", err)
	}
	if workpacket.Digest(message) != blocker.Artifact.MessageSHA256 {
		return nil, errors.New("accepted S9 blocker lifecycle message differs from its immutable SHA")
	}
	if err := schema.NewEmbeddedValidator().ValidateBytes("agent-message.schema.json", message); err != nil {
		return nil, fmt.Errorf("validate accepted S9 blocker lifecycle message: %w", err)
	}
	return message, nil
}

func publishRepairResultResolutionMessage(root string, packet workpacket.RepairResultPacket, data []byte) (string, []byte, error) {
	if len(data) == 0 || len(data) > repairResultResolutionMessageLimit {
		return "", nil, errors.New("S9 immutable resolution message is empty or exceeds 1 MiB")
	}
	if err := schema.NewEmbeddedValidator().ValidateBytes("agent-message.schema.json", data); err != nil {
		return "", nil, fmt.Errorf("validate immutable S9 resolution message: %w", err)
	}
	b := packet.Binding.Sources
	digest := workpacket.Digest(data)
	messageRef := path.Join(".claude/evidence", b.RuntimeID, fmt.Sprintf("g%d", b.BaselineGeneration), "S9", "blocker-messages", packet.ID, digest+".json")
	stored, err := workpacket.PublishBounded(root, messageRef, data, repairResultResolutionMessageLimit)
	if err != nil {
		return "", nil, err
	}
	if !bytes.Equal(stored, data) || workpacket.Digest(stored) != digest {
		return "", nil, runtimepkg.ErrSubmissionConflict
	}
	return messageRef, stored, nil
}

func verifyExactRepairResultResolutionReplay(root, authoringPath string, packet workpacket.RepairResultPacket, receipt runtimepkg.RepairResultResolutionReceipt) error {
	if receipt.PacketID != packet.ID || receipt.OwnerAgentID != packet.Binding.Sources.OwnerAgentID {
		return runtimepkg.ErrSubmissionConflict
	}
	authoring, err := workpacket.ReadFile(root, authoringPath, repairResultResolutionMessageLimit)
	if err != nil {
		return fmt.Errorf("read exact replayed S9 Main resolution authoring file: %w", err)
	}
	immutable, err := workpacket.ReadFile(root, receipt.MessageRef, repairResultResolutionMessageLimit)
	if err != nil {
		return fmt.Errorf("read accepted immutable S9 Main resolution message: %w", err)
	}
	if workpacket.Digest(immutable) != receipt.MessageSHA256 || !bytes.Equal(authoring, immutable) {
		return runtimepkg.ErrSubmissionConflict
	}
	return nil
}

func verifyCapabilityResolutionBytes(root string, capability *IssuedRepairResultResolutionCapability) error {
	immutable, err := workpacket.ReadFile(root, capability.messageRef, repairResultResolutionMessageLimit)
	if err != nil || !bytes.Equal(immutable, capability.messageBytes) || workpacket.Digest(immutable) != capability.messageSHA256 {
		return fmt.Errorf("S9 immutable resolution message changed before the original lifecycle CAS: %v", err)
	}
	if err := schema.NewEmbeddedValidator().ValidateBytes("agent-message.schema.json", immutable); err != nil {
		return err
	}
	authoring, err := workpacket.ReadFile(root, capability.authoringPath, repairResultResolutionMessageLimit)
	if err != nil || !bytes.Equal(authoring, capability.messageBytes) {
		return fmt.Errorf("S9 Main resolution authoring bytes changed after Prepare: %v", err)
	}
	return nil
}

func resolutionReceiptAlreadyExists(root string, state map[string]any, packet workpacket.RepairResultPacket, owner string) (bool, error) {
	_, found, err := runtimepkg.VerifiedRepairResultResolutionReceipt(root, state, packet.ID, owner)
	return found, err
}

func appendRepairResultResolutionEvidence(state map[string]any, packet workpacket.RepairResultPacket, blocker runtimepkg.RepairResultBlockerReceipt, message ArtifactRef) error {
	rows, ok := state["evidence"].([]any)
	if !ok {
		return errors.New("Runtime evidence index is unavailable for the original blocker_resolved lifecycle CAS")
	}
	for _, raw := range rows {
		row, _ := raw.(map[string]any)
		if row == nil {
			continue
		}
		if row["id"] == message.Path || row["path"] == message.Path {
			return fmt.Errorf("S9 resolution evidence path already exists; recover or inspect the original lifecycle event")
		}
		if row["kind"] == "repair_result_resolution" {
			refs, _ := row["scope_refs"].([]any)
			if anyString(refs, packet.ID) {
				return fmt.Errorf("S9 packet already has a resolution evidence row; inspect its original lifecycle event")
			}
		}
	}
	b := packet.Binding.Sources
	row := map[string]any{
		"id": message.Path, "kind": "repair_result_resolution", "path": message.Path, "sha256": message.SHA256,
		"status": "valid", "baseline_generation": b.BaselineGeneration, "review_round": b.ReviewRound,
		"produced_by": []any{"main"}, "invalidated_by": nil, "invalidation_rule": nil,
		"invalidation_reason": nil, "responsibility_id": nil,
		"scope_refs": []any{packet.ID, b.RepairAssignmentID, blocker.ArtifactRef},
	}
	state["evidence"] = append(rows, row)
	return nil
}

func anyString(values []any, wanted string) bool {
	for _, raw := range values {
		if value, ok := raw.(string); ok && value == wanted {
			return true
		}
	}
	return false
}

func resolutionMessageID(packetID, blockerSHA string) string {
	return "msg-s9-resolution-" + strings.TrimPrefix(packetID, "packet-") + "-" + blockerSHA[:16]
}
