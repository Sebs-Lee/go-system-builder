package submission

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/entroforge/go-system-builder/internal/assignment"
	"github.com/entroforge/go-system-builder/internal/repair"
	loopruntime "github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/workpacket"
	"github.com/entroforge/go-system-builder/internal/workspace"
)

const repairResultAnswerLimit int64 = 20 << 20

// IssueRepairResults delegates packet creation and PACKET-ISSUE to the
// original Repair Writer. This adapter only publishes the canonical view and
// answer draft into each exact registered Worker checkout.
func (s *Service) IssueRepairResults(ctx context.Context, ownerFilter string) ([]workpacket.RepairResultPacket, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// The named original issuer owns recovery of an exact pending PACKET-ISSUE
	// pair. Do not put a read-only service snapshot in front of that recovery.
	packets, err := repair.IssueRepairResultPackets(s.Root, s.StatePath, s.JournalPath, -1, ownerFilter)
	if err != nil {
		return nil, err
	}
	for _, packet := range packets {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		latest, err := s.snapshot()
		if err != nil {
			return nil, err
		}
		// This one snapshot-local verification includes exact registry equality,
		// issuance/source closure and the unique current Assignment slot. Physical
		// checkout validation below still runs against this latest snapshot.
		if err := loopruntime.VerifyRepairResultPacketIssueAndCurrent(s.Root, latest.State, packet); err != nil {
			return nil, err
		}
		sourceRoot, execution, err := RepairResultRegisteredSource(s.Root, latest.State, packet)
		if err != nil {
			return nil, err
		}
		registry, err := workspace.Decode(latest.State)
		if err != nil || registry == nil {
			return nil, fmt.Errorf("S9 result packet requires its registered Worker execution: %v", err)
		}
		if err := registry.ValidateExecution(ctx, execution, false); err != nil {
			return nil, err
		}
		if err := publishRepairResultAt(sourceRoot, packet); err != nil {
			return nil, err
		}
	}
	return packets, nil
}

// RepairResultSourceForPath resolves only a current packet in its registered
// execution. It is a selector, not proof of the Hook caller's identity.
func (s *Service) RepairResultSourceForPath(path string) (workpacket.RepairResultPacket, string, bool, error) {
	snapshot, err := s.snapshot()
	if err != nil {
		return workpacket.RepairResultPacket{}, "", false, err
	}
	return currentRepairResultSourceForPath(s.Root, snapshot.State, path)
}

// SaveRepairResultFrom accepts the exact issued answer through the original
// candidate/workspace_updated or blocker/work_blocked Writer. The packet's
// registered Worker source is selected again inside the original CAS.
func (s *Service) SaveRepairResultFrom(ctx context.Context, owner, path, sourceRoot string, authorize func(root string, state map[string]any) error) (Receipt, error) {
	snapshot, err := s.snapshot()
	if err != nil {
		return Receipt{}, err
	}
	if err := ctx.Err(); err != nil {
		return rejected(snapshot, "", err), err
	}
	packet, actualRoot, found, err := currentRepairResultSourceForPath(s.Root, snapshot.State, path)
	if err != nil {
		return rejected(snapshot, "", err), err
	}
	if !found || owner == "" || packet.Binding.Sources.OwnerAgentID != owner {
		err := errors.New("S9 result answer path has no current issuance for this registered Builder")
		return rejected(snapshot, "", err), err
	}
	want, err := filepath.Abs(sourceRoot)
	if err != nil {
		return rejected(snapshot, packet.ID, err), err
	}
	got, err := filepath.Abs(actualRoot)
	if err != nil || filepath.Clean(want) != filepath.Clean(got) {
		err := errors.New("S9 result answer source root differs from the registered Worker execution")
		return rejected(snapshot, packet.ID, err), err
	}
	if authorize != nil {
		if err := authorize(s.Root, snapshot.State); err != nil {
			return rejected(snapshot, packet.ID, err), err
		}
	}
	answerBytes, err := workpacket.ReadFile(actualRoot, packet.AnswerPath, repairResultAnswerLimit)
	if err != nil {
		return rejected(snapshot, packet.ID, err), err
	}
	answer, issues := workpacket.ValidateRepairResultAnswer(packet, answerBytes)
	if len(issues) != 0 {
		err := fmt.Errorf("S9 RepairResult answer %s: %s (%s)", issues[0].Pointer, issues[0].Expected, issues[0].Observed)
		r := rejected(snapshot, packet.ID, err)
		r.FieldErrors = issues
		return r, err
	}
	if err := ctx.Err(); err != nil {
		return rejected(snapshot, packet.ID, err), err
	}
	var professional workpacket.RepairResultProfessionalAnswer
	if err := json.Unmarshal(answer, &professional); err != nil {
		return rejected(snapshot, packet.ID, err), err
	}
	if professional.Verdict != "completed" {
		return s.saveRepairResultBlocker(ctx, packet, owner, actualRoot, professional.Verdict, authorize)
	}
	next, _, _, acceptErr := repair.AcceptIssuedRepairResultCandidate(ctx, s.Root, s.StatePath, s.JournalPath, repair.AcceptIssuedRepairResultCandidateRequest{
		ExpectedRevision: snapshot.Revision,
		PacketID:         packet.ID,
		AnswerPath:       packet.AnswerPath,
		OwnerAgentID:     owner,
		SourceRoot:       actualRoot,
		AuthorizeSource:  authorize,
		EnsureCompletion: func(ctx context.Context, current loopruntime.Snapshot, issued workpacket.RepairResultPacket, execution workspace.Execution, proposal repair.RepairResultRequest) (loopruntime.RepairResultCompletionReport, error) {
			return ensureS9OriginalCompletion(ctx, s.Root, s.StatePath, s.JournalPath, current, issued, execution, proposal)
		},
		OccurredAt: time.Now().UTC(),
	})
	if acceptErr != nil {
		return rejected(next, packet.ID, acceptErr), acceptErr
	}
	return acceptedRepairResultCandidateReceipt(s.Root, next, packet, owner)
}

// ensureS9OriginalCompletion projects the already validated professional
// RepairResult plus original Git/check facts into the existing completion
// message consumer. It introduces no completion schema or parallel receipt.
func ensureS9OriginalCompletion(ctx context.Context, root, statePath, journalPath string, current loopruntime.Snapshot, packet workpacket.RepairResultPacket, execution workspace.Execution, proposal repair.RepairResultRequest) (loopruntime.RepairResultCompletionReport, error) {
	answerBytes, err := workpacket.ReadFile(execution.Path, packet.AnswerPath, repairResultAnswerLimit)
	if err != nil {
		return loopruntime.RepairResultCompletionReport{}, err
	}
	answer, issues := workpacket.ValidateRepairResultAnswer(packet, answerBytes)
	if len(issues) != 0 {
		return loopruntime.RepairResultCompletionReport{}, fmt.Errorf("S9 answer changed before original completion: %s", issues[0].Pointer)
	}
	var professional workpacket.RepairResultProfessionalAnswer
	if err := json.Unmarshal(answer, &professional); err != nil {
		return loopruntime.RepairResultCompletionReport{}, err
	}
	var manifestAssignment struct {
		AgentDefinitionRef string `json:"agent_definition_ref"`
	}
	if err := json.Unmarshal(packet.Inputs.Sources.ManifestAssignment, &manifestAssignment); err != nil || manifestAssignment.AgentDefinitionRef == "" {
		return loopruntime.RepairResultCompletionReport{}, fmt.Errorf("S9 completion requires the frozen manifest agent_definition_ref: %v", err)
	}
	changed := make([]string, 0, len(proposal.ChangedArtifacts))
	for _, artifact := range proposal.ChangedArtifacts {
		changed = append(changed, artifact.Path)
	}
	checks := make([]map[string]string, 0, len(proposal.Checks)+len(execution.WorkerChecks))
	for _, check := range proposal.Checks {
		checks = append(checks, map[string]string{"name": check.Name, "command": check.Command, "result": check.Result})
	}
	for _, check := range execution.WorkerChecks {
		checks = append(checks, map[string]string{"name": check.Command, "command": check.Command, "result": "pass"})
	}
	summaries := make([]string, 0, len(professional.UnitResults))
	for _, unit := range professional.UnitResults {
		if text := strings.TrimSpace(unit.ImplementationSummary); text != "" {
			summaries = append(summaries, unit.UnitID+": "+text)
		}
	}
	summary := fmt.Sprintf("S9 RepairResult %s answer sha256=%s", packet.ResultID, workpacket.Digest(answer))
	if len(summaries) > 0 {
		summary += ": " + strings.Join(summaries, "; ")
	}
	bugID := s9CompletionBugID(packet.Inputs.Sources.Case)
	message := map[string]any{
		"schema_version": "1.0.0", "message_type": "completion_report",
		"message_id": "msg-s9-result-" + packet.ID, "correlation_id": "corr-s9-result-" + packet.ID,
		"runtime_id": packet.Binding.Sources.RuntimeID, "agent_id": packet.Binding.Sources.OwnerAgentID,
		"agent_definition_ref": manifestAssignment.AgentDefinitionRef, "task_id": packet.Binding.Sources.TaskID,
		"bug_id": bugID, "team_id": packet.Binding.Sources.TeamID,
		"occurred_at": time.Now().UTC().Format(time.RFC3339Nano), "status": "completed",
		"summary": summary, "changed_paths": changed, "reviewed_paths": []string{}, "checks": checks,
		"evidence_refs": professional.EvidenceRefs, "remaining_risks": professional.ResidualRisks,
		"scope_deviations": professional.ScopeDeviations, "requested_event": "completion_reported",
	}
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return loopruntime.RepairResultCompletionReport{}, err
	}
	file, err := os.CreateTemp("", "gsb-s9-completion-*.json")
	if err != nil {
		return loopruntime.RepairResultCompletionReport{}, err
	}
	messagePath := file.Name()
	defer os.Remove(messagePath)
	if _, err := file.Write(messageBytes); err != nil {
		_ = file.Close()
		return loopruntime.RepairResultCompletionReport{}, err
	}
	if err := file.Close(); err != nil {
		return loopruntime.RepairResultCompletionReport{}, err
	}
	_, _, completeErr := assignment.CompleteTaskWithReceipt(root, statePath, journalPath, assignment.CompletionRequest{
		ExpectedRevision: current.Revision, AgentID: packet.Binding.Sources.OwnerAgentID,
		MessagePath: messagePath, OccurredAt: time.Now().UTC(),
	})
	latest, snapshotErr := loopruntime.NewStore(statePath, journalPath).Snapshot()
	if snapshotErr != nil {
		return loopruntime.RepairResultCompletionReport{}, snapshotErr
	}
	report, reportErr := loopruntime.CurrentRepairResultCompletionReport(root, latest.State, packet)
	if reportErr == nil && verifyS9ProjectedCompletion(root, report, summary, changed, checks, professional.ScopeDeviations) == nil {
		return report, nil
	}
	if completeErr != nil {
		return loopruntime.RepairResultCompletionReport{}, fmt.Errorf("original task completion consumer: %w", completeErr)
	}
	if reportErr != nil {
		return loopruntime.RepairResultCompletionReport{}, reportErr
	}
	return loopruntime.RepairResultCompletionReport{}, errors.New("original S9 completion differs from the validated RepairResult/Git/check projection")
}

func s9CompletionBugID(raw json.RawMessage) any {
	var source struct {
		BugID        string   `json:"bug_id"`
		SourceBugIDs []string `json:"source_bug_ids"`
		BugIDs       []string `json:"bug_ids"`
	}
	if json.Unmarshal(raw, &source) == nil {
		if source.BugID != "" {
			return source.BugID
		}
		if len(source.SourceBugIDs) == 1 {
			return source.SourceBugIDs[0]
		}
		if len(source.BugIDs) == 1 {
			return source.BugIDs[0]
		}
	}
	return nil
}

func verifyS9ProjectedCompletion(root string, report loopruntime.RepairResultCompletionReport, summary string, changed []string, checks []map[string]string, deviations []string) error {
	data, err := workpacket.ReadFile(root, report.Ref, repairResultAnswerLimit)
	if err != nil || workpacket.Digest(data) != report.SHA256 {
		return fmt.Errorf("read original completion projection: %v", err)
	}
	var document map[string]any
	if err := json.Unmarshal(data, &document); err != nil {
		return err
	}
	for key, expected := range map[string]any{"summary": summary, "changed_paths": changed, "checks": checks, "scope_deviations": deviations} {
		got, _ := json.Marshal(document[key])
		want, _ := json.Marshal(expected)
		if string(got) != string(want) {
			return fmt.Errorf("original completion %s differs from the S9 domain projection", key)
		}
	}
	return nil
}

// RepairResultReceipt is a read-only view of the original blocker, candidate or final receipt.
// Main verification/final Result remains a separate original consumer.
func (s *Service) RepairResultReceipt(owner, ref string) (Receipt, error) {
	snapshot, err := s.snapshot()
	if err != nil {
		return Receipt{}, err
	}
	packet, found, err := repairResultByReference(s.Root, snapshot.State, ref)
	if err != nil {
		return Receipt{}, err
	}
	if !found || owner == "" || packet.Binding.Sources.OwnerAgentID != owner {
		return rejected(snapshot, "", errors.New("S9 packet reference has no matching registered Builder")), nil
	}
	if result, found, err := repairResultBlockerReceipt(s.Root, snapshot, packet, owner); err != nil || found {
		return result, err
	}
	final, finalFound, finalErr := loopruntime.VerifiedRepairResultReceipt(s.Root, snapshot.State, packet.ID, owner)
	if finalErr != nil {
		return Receipt{}, finalErr
	}
	if finalFound {
		result := emptyReceipt(snapshot, packet.ID)
		if final.Provenance.PacketID != packet.ID || final.Provenance.OwnerAgentID != owner || final.ResultID != packet.ResultID || final.ResultRef == "" || final.CandidateRef == "" {
			return Receipt{}, errors.New("S9 original final Result receipt differs from its packet")
		}
		result.Accepted = true
		// The original receipt is immutable historical acceptance. Project
		// current owner completion from the original domain consumer so an
		// accepted Result cannot revive responsibility after the execution was
		// replaced or explicitly reworked.
		workComplete, completionErr := repair.RepairResultWorkComplete(s.Root, snapshot.State, packet.ID, owner)
		result.WorkComplete = completionErr == nil && workComplete
		result.SubmissionKey = final.AcceptedIdempotencyKey
		result.ConsumerOwner = "main"
		result.ConsumerPending = true
		result.DomainRefs = []string{final.CandidateRef, final.ResultRef}
		if completionErr != nil {
			result.ResponsibilityStatus = "main_result_accepted_owner_status_unavailable"
			result.WaitingOn = append(result.WaitingOn, Waiting{Reason: "accepted Result is historical; current Worker completion proof could not be resolved", Owner: owner, Ref: packet.ID})
			result.NextAction = "原 Main Result 已接收且保留其历史回执；当前 Builder 完成状态无法核验，先用 runtime repair status 检查执行/返工记录，不要把历史 receipt 当作当前完成。"
		} else if !workComplete {
			result.ResponsibilityStatus = "main_result_accepted_rework_required"
			result.WaitingOn = append(result.WaitingOn, Waiting{Reason: "accepted Result is historical; current Worker attempt no longer matches its candidate", Owner: owner, Ref: packet.ID})
			result.NextAction = "原 Main Result 已接收且保留其历史回执；当前执行代次或返工已变化，按 runtime repair status 给出的原命名返工路径继续，不要把旧 receipt 当作当前完成。"
		} else {
			result.ResponsibilityStatus = "main_result_accepted"
			result.NextAction = "原 Main Result 已接收；继续原 checkpoint ACK/cleanup 与 S9/S7 后续门，不重复提交答卷。"
		}
		return result, nil
	}
	return acceptedRepairResultCandidateReceipt(s.Root, snapshot, packet, owner)
}

func currentRepairResultSourceForPath(root string, state map[string]any, inputPath string) (workpacket.RepairResultPacket, string, bool, error) {
	packets, err := loopruntime.CurrentRepairResultPackets(root, state, "")
	if err != nil {
		return workpacket.RepairResultPacket{}, "", false, err
	}
	packet, source, found, err := repairResultPathInRoots(root, state, inputPath, packets)
	if err != nil || found {
		return packet, source, found, err
	}
	// An accepted packet may no longer be in the fresh selector after the
	// lifecycle advances. A delayed Hook replay can still locate its exact old
	// answer from the original durable candidate receipt; this path grants no
	// fresh write because the domain Writer performs the receipt/digest check.
	all, err := loopruntime.RepairResultPackets(state)
	if err != nil {
		return workpacket.RepairResultPacket{}, "", false, err
	}
	return acceptedRepairResultPathWithSource(root, state, inputPath, all)
}

func repairResultByReference(root string, state map[string]any, ref string) (workpacket.RepairResultPacket, bool, error) {
	if packet, found, err := loopruntime.FindRepairResultPacket(state, ref); err != nil || found {
		return packet, found, err
	}
	packets, err := loopruntime.RepairResultPackets(state)
	if err != nil {
		return workpacket.RepairResultPacket{}, false, err
	}
	packet, _, found, err := repairResultPathInRoots(root, state, ref, packets)
	if err != nil || found {
		return packet, found, err
	}
	packet, _, found, err = acceptedRepairResultPathWithSource(root, state, ref, packets)
	return packet, found, err
}

func acceptedRepairResultPathInRoots(root string, state map[string]any, inputPath string, packets []workpacket.RepairResultPacket) (workpacket.RepairResultPacket, bool, error) {
	packet, _, found, err := acceptedRepairResultPathWithSource(root, state, inputPath, packets)
	return packet, found, err
}

func acceptedRepairResultPathWithSource(root string, state map[string]any, inputPath string, packets []workpacket.RepairResultPacket) (workpacket.RepairResultPacket, string, bool, error) {
	abs, err := filepath.Abs(inputPath)
	if err != nil {
		return workpacket.RepairResultPacket{}, "", false, err
	}
	if !filepath.IsAbs(inputPath) {
		abs, err = filepath.Abs(filepath.Join(root, filepath.FromSlash(inputPath)))
		if err != nil {
			return workpacket.RepairResultPacket{}, "", false, err
		}
	}
	var selected *workpacket.RepairResultPacket
	selectedRoot := ""
	for _, candidate := range packets {
		accepted, ok, receiptErr := loopruntime.VerifiedRepairResultCandidateReceipt(root, state, candidate.ID, candidate.Binding.Sources.OwnerAgentID)
		if receiptErr != nil {
			return workpacket.RepairResultPacket{}, "", false, receiptErr
		}
		acceptedRoot := ""
		if ok && accepted.Provenance.PacketID == candidate.ID && accepted.Provenance.OwnerAgentID == candidate.Binding.Sources.OwnerAgentID {
			acceptedRoot = accepted.Provenance.ExecutionRoot
		} else {
			blocker, blocked, err := loopruntime.VerifiedRepairResultBlockerReceipt(root, state, candidate.ID, candidate.Binding.Sources.OwnerAgentID)
			if err != nil {
				return workpacket.RepairResultPacket{}, "", false, err
			}
			if !blocked || blocker.Artifact.PacketID != candidate.ID || blocker.Artifact.Binding.Sources.OwnerAgentID != candidate.Binding.Sources.OwnerAgentID {
				continue
			}
			acceptedRoot = blocker.Artifact.ExecutionRoot
		}
		sourceRoot, err := workspace.Canonical(acceptedRoot)
		if err != nil || filepath.Clean(sourceRoot) == filepath.Clean(root) {
			continue
		}
		var frozenExecution struct {
			Path string `json:"worktree_path"`
		}
		if json.Unmarshal(candidate.Inputs.Execution, &frozenExecution) != nil || frozenExecution.Path == "" {
			continue
		}
		frozenRoot, err := workspace.Canonical(frozenExecution.Path)
		if err != nil || filepath.Clean(frozenRoot) != filepath.Clean(sourceRoot) {
			continue
		}
		candidatePath, err := filepath.Abs(filepath.Join(sourceRoot, filepath.FromSlash(candidate.AnswerPath)))
		if err != nil || filepath.Clean(candidatePath) != filepath.Clean(abs) {
			continue
		}
		if err := workpacket.CheckPath(sourceRoot, candidate.AnswerPath); err != nil {
			return workpacket.RepairResultPacket{}, "", false, err
		}
		copyPacket := candidate
		if selected != nil && (selected.ID != copyPacket.ID || filepath.Clean(selectedRoot) != filepath.Clean(sourceRoot)) {
			return workpacket.RepairResultPacket{}, "", false, errors.New("S9 accepted answer path is ambiguous across issued packets")
		}
		selected, selectedRoot = &copyPacket, sourceRoot
	}
	if selected == nil {
		return workpacket.RepairResultPacket{}, "", false, nil
	}
	return *selected, selectedRoot, true, nil
}

func repairResultPathInRoots(root string, state map[string]any, inputPath string, packets []workpacket.RepairResultPacket) (workpacket.RepairResultPacket, string, bool, error) {
	abs, err := filepath.Abs(inputPath)
	if err != nil {
		return workpacket.RepairResultPacket{}, "", false, err
	}
	if !filepath.IsAbs(inputPath) {
		abs, err = filepath.Abs(filepath.Join(root, filepath.FromSlash(inputPath)))
		if err != nil {
			return workpacket.RepairResultPacket{}, "", false, err
		}
	}
	var selected *workpacket.RepairResultPacket
	selectedRoot := ""
	for _, packet := range packets {
		sourceRoot, _, err := RepairResultRegisteredSource(root, state, packet)
		if err != nil {
			// A stale/historical packet is not a source-root grant. It remains
			// readable by packet ID through its original receipt.
			continue
		}
		candidate, err := filepath.Abs(filepath.Join(sourceRoot, filepath.FromSlash(packet.AnswerPath)))
		if err != nil || filepath.Clean(candidate) != filepath.Clean(abs) {
			continue
		}
		if err := workpacket.CheckPath(sourceRoot, packet.AnswerPath); err != nil {
			return workpacket.RepairResultPacket{}, "", false, err
		}
		if selected != nil && (selected.ID != packet.ID || filepath.Clean(selectedRoot) != filepath.Clean(sourceRoot)) {
			return workpacket.RepairResultPacket{}, "", false, errors.New("S9 answer path is ambiguous across issued packets")
		}
		copyPacket := packet
		selected, selectedRoot = &copyPacket, sourceRoot
	}
	if selected == nil {
		return workpacket.RepairResultPacket{}, "", false, nil
	}
	return *selected, selectedRoot, true, nil
}

// RepairResultRegisteredSource reuses the original execution/source selector for
// both the consumer and its questionnaire view; it does not authorize tool calls.
func RepairResultRegisteredSource(root string, state map[string]any, packet workpacket.RepairResultPacket) (string, workspace.Execution, error) {
	registry, err := workspace.Decode(state)
	if err != nil || registry == nil {
		return "", workspace.Execution{}, fmt.Errorf("S9 result packet requires a registered Worker execution: %v", err)
	}
	if err := registry.ValidateAuthority(state); err != nil {
		return "", workspace.Execution{}, err
	}
	b := packet.Binding.Sources
	var selected *workspace.Execution
	for _, id := range []string{b.RepairAssignmentID, b.ManifestAssignmentID} {
		if id == "" {
			continue
		}
		e, ok := registry.Execution(id, b.RuntimeID, b.BaselineGeneration)
		if !ok {
			continue
		}
		if e.AssignmentID != id || e.AgentID != b.OwnerAgentID || e.Status != "ready" {
			return "", workspace.Execution{}, errors.New("S9 packet execution owner, identity, or status differs")
		}
		if selected != nil && (selected.AssignmentID != e.AssignmentID || filepath.Clean(selected.Path) != filepath.Clean(e.Path)) {
			return "", workspace.Execution{}, errors.New("S9 packet has ambiguous canonical/manifest Worker executions")
		}
		copyExecution := e
		selected = &copyExecution
	}
	if selected == nil {
		return "", workspace.Execution{}, errors.New("S9 packet has no exact registered Worker execution")
	}
	rootPath, err := workspace.Canonical(selected.Path)
	if err != nil || rootPath != selected.Path {
		return "", workspace.Execution{}, errors.New("S9 registered Worker source root is unavailable or noncanonical")
	}
	if filepath.Clean(rootPath) == filepath.Clean(root) {
		return "", workspace.Execution{}, errors.New("S9 typed RepairResult does not support Main-native execution")
	}
	return rootPath, *selected, nil
}

func publishRepairResultAt(root string, packet workpacket.RepairResultPacket) error {
	view, err := json.MarshalIndent(packet, "", "  ")
	if err != nil {
		return err
	}
	view = append(view, '\n')
	stored, err := workpacket.PublishBounded(root, packet.ViewPath, view, repairResultAnswerLimit)
	if err != nil {
		return err
	}
	if !bytes.Equal(stored, view) {
		return fmt.Errorf("S9 result issued view differs from Runtime packet")
	}
	draft, err := json.MarshalIndent(packet.Draft, "", "  ")
	if err != nil {
		return err
	}
	_, err = workpacket.PublishBounded(root, packet.AnswerPath, append(draft, '\n'), repairResultAnswerLimit)
	return err
}

func acceptedRepairResultCandidateReceipt(root string, snapshot loopruntime.Snapshot, packet workpacket.RepairResultPacket, owner string) (Receipt, error) {
	result := emptyReceipt(snapshot, packet.ID)
	accepted, ok, err := loopruntime.VerifiedRepairResultCandidateReceipt(root, snapshot.State, packet.ID, owner)
	if err != nil {
		return result, err
	}
	if !ok {
		result = rejected(snapshot, packet.ID, errors.New("S9 candidate has not been accepted by the original workspace_updated Writer"))
		result.NextAction = "候选尚未由原 workspace_updated Writer 接收；保持工作区并重新提交已发行答卷。"
		return result, nil
	}
	if accepted.Provenance.PacketID != packet.ID || accepted.Provenance.OwnerAgentID != owner || accepted.CandidateRef == "" || accepted.CandidateSHA256 == "" || accepted.AcceptedEventID == "" {
		return result, errors.New("original S9 candidate receipt differs from its issued packet")
	}
	result.Accepted = true
	result.SubmissionKey = accepted.AcceptedIdempotencyKey
	result.ResponsibilityStatus = "candidate_accepted"
	result.DomainRefs = []string{accepted.CandidateRef}
	result.WaitingOn = []Waiting{}
	result.WorkComplete, err = repair.RepairResultWorkComplete(root, snapshot.State, packet.ID, owner)
	if err != nil {
		result.WorkComplete = false
		result.WaitingOn = append(result.WaitingOn, Waiting{Reason: "S9 owner execution/check proof needs recovery", Owner: owner, Ref: packet.ID})
		result.NextAction = "保全原 candidate；修复已登记的 Worker source/check receipt 后，再查此 packet 的原回执。"
		return result, nil
	}
	if result.WorkComplete {
		result.ConsumerOwner = "main"
		result.ConsumerPending = true
		result.NextAction = "本人 S9 candidate 与 owner-owned checks 已由原消费者接收；Main 仍须执行原集成、required checks 与 Result 接收。"
	} else {
		result.WaitingOn = append(result.WaitingOn, Waiting{Reason: "candidate/required Worker checks are not complete or current", Owner: owner, Ref: packet.ID})
		result.NextAction = "继续原 Worker execution，完成本人 frozen checks 与 candidate 接收条件后再结束。"
	}
	return result, nil
}

func sameRepairResultPacket(a, b workpacket.RepairResultPacket) bool {
	left, err := workpacket.RepairResultPacketDigest(a)
	if err != nil {
		return false
	}
	right, err := workpacket.RepairResultPacketDigest(b)
	return err == nil && left == right
}

func hasRepairResultPacket(values []workpacket.RepairResultPacket, expected workpacket.RepairResultPacket) bool {
	for _, packet := range values {
		if packet.ID == expected.ID && sameRepairResultPacket(packet, expected) {
			return true
		}
	}
	return false
}
