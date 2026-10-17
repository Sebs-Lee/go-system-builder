package submission

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/entroforge/go-system-builder/internal/agentauth"
	"github.com/entroforge/go-system-builder/internal/assignment"
	"github.com/entroforge/go-system-builder/internal/repair"
	loopruntime "github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/workpacket"
)

const repairPlanReportAnswerLimit int64 = 20 << 20

// IssueRepairPlanReports delegates publication to the original S9 packet
// Writer. It adds no packet registry or receipt authority of its own.
func (s *Service) IssueRepairPlanReports(ctx context.Context, ownerFilter string) ([]workpacket.RepairPlanReportPacket, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// The original packet Issuer owns pending-marker recovery and its own CAS.
	// Do not pre-read through the read-only service snapshot here: a pending
	// original PACKET-ISSUE must be recoverable by this named mutation path.
	packets, err := repair.IssueRepairPlanReportPackets(s.Root, s.StatePath, s.JournalPath, -1, ownerFilter)
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
		stored, found, err := loopruntime.FindRepairPlanReportPacket(latest.State, packet.ID)
		if err != nil || !found || !sameRepairPlanReportPacket(stored, packet) {
			return nil, fmt.Errorf("S9 issued packet %s is no longer the exact Runtime packet: %v", packet.ID, err)
		}
		current, err := loopruntime.CurrentRepairPlanReportPackets(s.Root, latest.State, packet.Binding.OwnerAgentID)
		if err != nil || !hasRepairPlanReportPacket(current, packet) {
			return nil, fmt.Errorf("S9 issued packet %s no longer matches the current Session/Plan/Assignment tuple: %v", packet.ID, err)
		}
		if err := loopruntime.VerifyRepairPlanReportPacketIssue(s.Root, latest.State, packet); err != nil {
			return nil, err
		}
		if err := loopruntime.VerifyRepairPlanReportCurrent(s.Root, latest.State, packet); err != nil {
			return nil, err
		}
		if err := publishRepairPlanReportView(s.Root, packet); err != nil {
			return nil, err
		}
		draft, err := jsonMarshalIndent(packet.Draft)
		if err != nil {
			return nil, err
		}
		if _, err := workpacket.PublishBounded(s.Root, packet.AnswerPath, append(draft, '\n'), repairPlanReportAnswerLimit); err != nil {
			return nil, err
		}
	}
	return packets, nil
}

// RepairPlanReportSourceForPath resolves a current answer path to its exact
// packet and source root. The pre-execution S9 profile retains its original
// Main-root authoring source.
func (s *Service) RepairPlanReportSourceForPath(path string) (workpacket.RepairPlanReportPacket, string, bool, error) {
	snapshot, err := s.snapshot()
	if err != nil {
		return workpacket.RepairPlanReportPacket{}, "", false, err
	}
	return currentRepairPlanReportSourceForPath(s.Root, snapshot.State, path)
}

// SaveRepairPlanReportFrom accepts a typed author answer through the original
// RepairPlanReport Writer. An optional integration callback remains supported;
// the normal Hook path relies on the registered packet and original source.
func (s *Service) SaveRepairPlanReportFrom(ctx context.Context, owner, path, sourceRoot string, authorize func(root string, state map[string]any) error) (Receipt, error) {
	snapshot, err := s.snapshot()
	if err != nil {
		return Receipt{}, err
	}
	if err := ctx.Err(); err != nil {
		return rejected(snapshot, "", err), err
	}
	packet, actualRoot, found, err := currentRepairPlanReportSourceForPath(s.Root, snapshot.State, path)
	if err != nil {
		return rejected(snapshot, "", err), err
	}
	if !found || owner == "" || packet.Binding.OwnerAgentID != owner {
		err := errors.New("S9 answer path has no current issuance for this registered Builder")
		return rejected(snapshot, "", err), err
	}
	if filepath.Clean(sourceRoot) != filepath.Clean(actualRoot) {
		err := errors.New("S9 pre-execution Worker answer source is unavailable; this packet must be saved from the original Main root")
		return rejected(snapshot, packet.ID, err), err
	}
	validateSource := func(controlRoot string, state map[string]any) error {
		if authorize != nil {
			if err := authorize(controlRoot, state); err != nil {
				return err
			}
		}
		if err := loopruntime.VerifyRepairPlanReportPacketIssue(controlRoot, state, packet); err != nil {
			return err
		}
		prior, accepted, err := repair.IssuedRepairPlanReportReceipt(controlRoot, state, packet.ID, owner)
		if err != nil {
			return err
		}
		if accepted {
			if prior.PacketID != packet.ID || prior.OwnerAgentID != owner || prior.WorkComplete {
				return errors.New("S9 original receipt does not match this accepted packet replay")
			}
			return nil
		}
		if err := loopruntime.VerifyRepairPlanReportCurrent(controlRoot, state, packet); err != nil {
			return err
		}
		return repair.ValidateRepairPlanReportAuthoringStage(controlRoot, state, owner)
	}
	if err := validateSource(s.Root, snapshot.State); err != nil {
		return rejected(snapshot, packet.ID, err), err
	}
	answerBytes, err := workpacket.ReadFile(actualRoot, packet.AnswerPath, repairPlanReportAnswerLimit)
	if err != nil {
		return rejected(snapshot, packet.ID, err), err
	}
	professionalAnswer, issues := workpacket.ValidateRepairPlanReportAnswer(packet, answerBytes)
	if len(issues) != 0 {
		issue := issues[0]
		err := fmt.Errorf("S9 answer %s: %s (%s)", issue.Pointer, issue.Expected, issue.Observed)
		r := rejected(snapshot, packet.ID, err)
		r.FieldErrors = issues
		return r, err
	}
	if err := ctx.Err(); err != nil {
		return rejected(snapshot, packet.ID, err), err
	}
	next, report, submitErr := repair.SubmitIssuedRepairPlanReport(s.Root, s.StatePath, s.JournalPath, repair.IssuedPlanReportRequest{
		ExpectedRevision: snapshot.Revision,
		PacketID:         packet.ID,
		AnswerPath:       packet.AnswerPath,
		OwnerAgentID:     owner,
		SourceRoot:       actualRoot,
		AuthorizeSource:  validateSource,
		OccurredAt:       time.Now().UTC(),
	})
	if submitErr != nil {
		return rejected(snapshot, packet.ID, submitErr), submitErr
	}
	accepted, ok, err := repair.IssuedRepairPlanReportReceipt(s.Root, next.State, packet.ID, owner)
	if err != nil {
		return Receipt{}, err
	}
	if !ok {
		return rejected(next, packet.ID, errors.New("original S9 consumer returned without a verified PlanReport receipt")), nil
	}
	var professional workpacket.RepairPlanReportProfessionalAnswer
	if err := json.Unmarshal(professionalAnswer, &professional); err != nil {
		return Receipt{}, err
	}
	latest, beginErr := ensureS9PlanCheckpoint(s.Root, s.StatePath, s.JournalPath, packet, report.ReportedAt, accepted.AnswerSHA256, professional)
	if latest.Revision == 0 && latest.State == nil {
		latest = next
	}
	receipt, err := acceptedRepairPlanReportReceipt(latest, packet, accepted)
	if err != nil {
		return receipt, err
	}
	if beginErr != nil {
		receipt.ConsumerOwner = "main"
		receipt.ConsumerPending = true
		receipt.WaitingOn = append(receipt.WaitingOn, Waiting{Reason: "original S9 PlanReport is accepted; generic PLAN/activation has not finished", Owner: "Main", Ref: packet.AnswerPath})
		receipt.NextAction = fmt.Sprintf("原 S9 PlanReport 已接收。当前自动连接 generic PLAN/AgentBegin 未完成：%v；保留此精确答案路径 %s，再保存同一问卷以恢复原链。", beginErr, packet.AnswerPath)
		return receipt, nil
	}
	receipt.NextAction = "原 S9 PlanReport 已接收；同一专业答案已自动接入原 generic PLAN/activation 链。继续原 BeginRepairExecution 门，无需另交 generic PLAN。"
	return receipt, nil
}

// ensureS9PlanCheckpoint projects only fields already present in the accepted
// professional PlanReport and exact frozen dispatch metadata, then runs the
// existing AgentBegin chain. The deterministic existing message can be reused
// after any partial original lifecycle commit.
func ensureS9PlanCheckpoint(root, statePath, journalPath string, packet workpacket.RepairPlanReportPacket, occurredAt, answerSHA string, professional workpacket.RepairPlanReportProfessionalAnswer) (loopruntime.Snapshot, error) {
	current, err := loopruntime.NewStore(statePath, journalPath).Snapshot()
	if err != nil {
		return loopruntime.Snapshot{}, err
	}
	accepted, ok, err := repair.IssuedRepairPlanReportReceipt(root, current.State, packet.ID, packet.Binding.OwnerAgentID)
	if err != nil || !ok || accepted.AnswerSHA256 != answerSHA {
		return current, fmt.Errorf("S9 generic PLAN projection requires this exact original accepted answer: %v", err)
	}
	if err := agentauth.ValidateWorkingAuthorization(root, current.State, packet.Binding.OwnerAgentID); err == nil {
		return current, nil
	}
	if err := agentauth.ValidateReportedPlanAuthorization(root, current.State, packet.Binding.OwnerAgentID); err == nil {
		return current, nil
	}
	if err := agentauth.ValidateBlockedPlanAuthorization(root, current.State, packet.Binding.OwnerAgentID); err == nil {
		return current, nil
	}
	var manifest struct {
		AgentDefinitionRef string   `json:"agent_definition_ref"`
		DependsOn          []string `json:"depends_on"`
	}
	if err := json.Unmarshal(packet.Inputs.ManifestAssignment, &manifest); err != nil || manifest.AgentDefinitionRef == "" {
		return current, fmt.Errorf("S9 generic PLAN lacks frozen manifest assignment metadata: %v", err)
	}
	assertions := make([]map[string]string, len(professional.RedChecks))
	for i, check := range professional.RedChecks {
		assertions[i] = map[string]string{"assertion": check.Name, "oracle": check.Command}
	}
	var caseSource struct {
		BugID any `json:"bug_id"`
	}
	if err := json.Unmarshal(packet.Inputs.Case, &caseSource); err != nil {
		return current, err
	}
	bugID := caseSource.BugID
	if bugID == "" {
		bugID = nil
	}
	message := map[string]any{
		"schema_version": "1.0.0", "message_type": "plan_report",
		"message_id": "msg-s9-plan-" + packet.ID, "correlation_id": "corr-s9-plan-" + packet.ID,
		"runtime_id": packet.Binding.RuntimeID, "agent_id": packet.Binding.OwnerAgentID,
		"agent_definition_ref": manifest.AgentDefinitionRef, "task_id": packet.Binding.TaskID,
		"bug_id": bugID, "team_id": packet.Binding.TeamID, "occurred_at": occurredAt,
		"assignment_id": packet.Binding.ManifestAssignmentID, "assignment_revision": 1,
		"objective": professional.Plan, "planned_paths": professional.ProposedPaths,
		"assertion_checks": assertions, "dependencies": manifest.DependsOn,
	}
	data, err := json.Marshal(message)
	if err != nil {
		return current, err
	}
	planPath := filepath.ToSlash(filepath.Join(".claude", "evidence", packet.Binding.RuntimeID,
		fmt.Sprintf("g%d", packet.Binding.BaselineGeneration), "assignments", packet.Binding.OwnerAgentID,
		"s9-plan-"+packet.ID+".json"))
	published, err := workpacket.PublishBounded(root, planPath, data, 1<<20)
	if err != nil {
		return current, err
	}
	if !bytes.Equal(published, data) {
		return current, errors.New("S9 generic PLAN checkpoint already exists with different projected bytes")
	}
	if err := loopruntime.VerifyRepairPlanReportCurrent(root, current.State, packet); err != nil {
		return current, err
	}
	if _, _, err := assignment.AgentBegin(root, statePath, journalPath, assignment.AgentBeginRequest{
		ExpectedRevision: -1, AgentID: packet.Binding.OwnerAgentID, PlanPath: planPath,
	}); err != nil {
		latest, snapErr := loopruntime.NewStore(statePath, journalPath).Snapshot()
		if snapErr == nil {
			current = latest
		}
		return current, err
	}
	return loopruntime.NewStore(statePath, journalPath).Snapshot()
}

// RepairPlanReportReceipt reconstructs an accepted pre-execution receipt by
// packet ID or exact registered answer path. It does not read the answer file.
func (s *Service) RepairPlanReportReceipt(owner, ref string) (Receipt, error) {
	snapshot, err := s.snapshot()
	if err != nil {
		return Receipt{}, err
	}
	packet, found, err := repairPlanReportByReference(s.Root, snapshot.State, ref)
	if err != nil {
		return Receipt{}, err
	}
	if !found || owner == "" || packet.Binding.OwnerAgentID != owner {
		return rejected(snapshot, "", errors.New("S9 packet reference has no matching registered Builder")), nil
	}
	accepted, ok, err := repair.IssuedRepairPlanReportReceipt(s.Root, snapshot.State, packet.ID, owner)
	if err != nil {
		return Receipt{}, err
	}
	if !ok {
		return rejected(snapshot, packet.ID, errors.New("PlanReport has not been accepted")), nil
	}
	return acceptedRepairPlanReportReceipt(snapshot, packet, accepted)
}

func currentRepairPlanReportSourceForPath(root string, state map[string]any, inputPath string) (workpacket.RepairPlanReportPacket, string, bool, error) {
	packets, err := loopruntime.CurrentRepairPlanReportPackets(root, state, "")
	if err != nil {
		return workpacket.RepairPlanReportPacket{}, "", false, err
	}
	return repairPlanReportPathInPackets(root, inputPath, packets)
}

func repairPlanReportByReference(root string, state map[string]any, ref string) (workpacket.RepairPlanReportPacket, bool, error) {
	if packet, found, err := loopruntime.FindRepairPlanReportPacket(state, ref); err != nil || found {
		return packet, found, err
	}
	packets, err := loopruntime.RepairPlanReportPackets(state)
	if err != nil {
		return workpacket.RepairPlanReportPacket{}, false, err
	}
	packet, _, found, err := repairPlanReportPathInPackets(root, ref, packets)
	return packet, found, err
}

func repairPlanReportPathInPackets(root, inputPath string, packets []workpacket.RepairPlanReportPacket) (workpacket.RepairPlanReportPacket, string, bool, error) {
	abs := inputPath
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(root, filepath.FromSlash(inputPath))
	}
	abs, err := filepath.Abs(abs)
	if err != nil {
		return workpacket.RepairPlanReportPacket{}, "", false, err
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return workpacket.RepairPlanReportPacket{}, "", false, err
	}
	var selected *workpacket.RepairPlanReportPacket
	for _, packet := range packets {
		candidate, err := filepath.Abs(filepath.Join(rootAbs, filepath.FromSlash(packet.AnswerPath)))
		if err != nil || filepath.Clean(candidate) != filepath.Clean(abs) {
			continue
		}
		if err := workpacket.CheckPath(rootAbs, packet.AnswerPath); err != nil {
			return workpacket.RepairPlanReportPacket{}, "", false, err
		}
		if selected != nil && selected.ID != packet.ID {
			return workpacket.RepairPlanReportPacket{}, "", false, errors.New("S9 answer path is ambiguous across issued packets")
		}
		copyPacket := packet
		selected = &copyPacket
	}
	if selected == nil {
		return workpacket.RepairPlanReportPacket{}, "", false, nil
	}
	return *selected, rootAbs, true, nil
}

func publishRepairPlanReportView(root string, packet workpacket.RepairPlanReportPacket) error {
	data, err := jsonMarshalIndent(packet)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	stored, err := workpacket.PublishBounded(root, packet.ViewPath, data, repairPlanReportAnswerLimit)
	if err != nil {
		return err
	}
	if !bytes.Equal(stored, data) {
		return fmt.Errorf("S9 issued view differs from the Runtime packet: %s", packet.ViewPath)
	}
	return nil
}

func sameRepairPlanReportPacket(a, b workpacket.RepairPlanReportPacket) bool {
	left, err := workpacket.RepairPlanReportPacketDigest(a)
	if err != nil {
		return false
	}
	right, err := workpacket.RepairPlanReportPacketDigest(b)
	return err == nil && left == right
}

func hasRepairPlanReportPacket(packets []workpacket.RepairPlanReportPacket, expected workpacket.RepairPlanReportPacket) bool {
	for _, packet := range packets {
		if packet.ID == expected.ID && sameRepairPlanReportPacket(packet, expected) {
			return true
		}
	}
	return false
}

func acceptedRepairPlanReportReceipt(snapshot loopruntime.Snapshot, packet workpacket.RepairPlanReportPacket, accepted loopruntime.RepairPlanReportReceipt) (Receipt, error) {
	result := emptyReceipt(snapshot, packet.ID)
	if accepted.PacketID != packet.ID || accepted.OwnerAgentID != packet.Binding.OwnerAgentID ||
		accepted.ReportID != packet.ReportID || accepted.AnswerSHA256 == "" || accepted.WorkComplete {
		return result, errors.New("original S9 PlanReport receipt differs from its issued packet")
	}
	result.SubmissionKey = accepted.AcceptedIdempotencyKey
	result.Accepted = true
	result.ResponsibilityStatus = "accepted"
	result.WorkComplete = false
	result.DomainRefs = []string{accepted.ReportRef}
	result.WaitingOn = []Waiting{}
	result.NextAction = "S9 pre-execution PlanReport 已由原 Writer 接收；继续 Main 的修复流程，收到该回执不会解除 BeginRepairExecution 前的产品写保护。"
	return result, nil
}

func jsonMarshalIndent(v any) ([]byte, error) { return json.MarshalIndent(v, "", "  ") }
