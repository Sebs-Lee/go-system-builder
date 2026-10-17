package submission

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/entroforge/go-system-builder/internal/assignment"
	"github.com/entroforge/go-system-builder/internal/integration"
	"github.com/entroforge/go-system-builder/internal/repair"
	loopruntime "github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/workpacket"
)

// A blocker uses the existing lifecycle Writer and its opaque proof. Holding
// the original workspace lock through prepare and CAS keeps the source tuple
// stable; this adapter owns no additional Runtime transaction.
func (s *Service) saveRepairResultBlocker(ctx context.Context, packet workpacket.RepairResultPacket, owner, sourceRoot, verdict string, authorize func(string, map[string]any) error) (Receipt, error) {
	ctx, unlock, err := integration.LockWorkspace(ctx, s.Root)
	if err != nil {
		return Receipt{}, err
	}
	defer unlock()
	snapshot, err := s.snapshot()
	if err != nil {
		return Receipt{}, err
	}
	capability, prior, err := repair.PrepareIssuedRepairResultBlocker(ctx, s.Root, s.StatePath, s.JournalPath, repair.AcceptIssuedRepairResultBlockerRequest{
		ExpectedRevision: snapshot.Revision, PacketID: packet.ID, AnswerPath: packet.AnswerPath,
		OwnerAgentID: owner, SourceRoot: sourceRoot, AuthorizeSource: authorize,
	})
	if err != nil {
		return rejected(snapshot, packet.ID, err), err
	}
	if prior == nil {
		if capability == nil {
			err := errors.New("original S9 blocker preparation returned no proof")
			return rejected(snapshot, packet.ID, err), err
		}
		next, advanceErr := assignment.AdvanceAgentWithContext(ctx, s.Root, s.StatePath, s.JournalPath, assignment.AgentEventRequest{
			ExpectedRevision: snapshot.Revision, AgentID: owner, Event: "work_blocked",
			MessagePath:               filepath.Join(s.Root, filepath.FromSlash(capability.MessageRef())),
			IssuedRepairResultBlocker: capability,
		})
		if advanceErr != nil {
			return rejected(snapshot, packet.ID, advanceErr), advanceErr
		}
		snapshot = next
	} else {
		// Prepare may recover an already committed marker. Rebuild from that
		// original pair instead of returning the pre-prepare snapshot.
		snapshot, err = s.snapshot()
		if err != nil {
			return Receipt{}, err
		}
	}
	result, found, err := repairResultBlockerReceipt(s.Root, snapshot, packet, owner)
	if err == nil && !found {
		err = errors.New("original S9 blocker CAS has no verifiable receipt")
	}
	if err == nil && found && result.currentRepairResultBlocker {
		if verdict == "contract_conflict" {
			result.NextAction = "原 Contract 冲突答卷已接收，本次执行可结束，交付未完成；Main 必须回原 S8 新 Contract 审批人闸，不能在同一 Contract/Session/Plan 上普通 retry。"
		} else if path, publishErr := publishRepairResultResolutionDraft(s.Root, snapshot, packet, owner); publishErr != nil {
			result.Error = "original blocker accepted; Main resolution draft publication failed: " + publishErr.Error()
			result.NextAction = "原阻塞接收已保留；Main 草稿发布失败。保全来源并原样重放本答卷以恢复草稿，不重建 Runtime 接收。"
		} else {
			result.NextAction = "原阻塞答卷已接收，本次执行可结束，交付未完成；等待 Main 填写并保存 " + path + " 的恢复理由与依据。"
		}
	}
	return result, err
}

func repairResultBlockerReceipt(root string, snapshot loopruntime.Snapshot, packet workpacket.RepairResultPacket, owner string) (Receipt, bool, error) {
	original, found, err := loopruntime.VerifiedRepairResultBlockerReceipt(root, snapshot.State, packet.ID, owner)
	if err != nil || !found {
		return Receipt{}, found, err
	}
	result := emptyReceipt(snapshot, packet.ID)
	result.Accepted, result.WorkComplete = true, false
	result.ResponsibilityStatus = "blocker_recorded"
	result.SubmissionKey = original.AcceptedIdempotencyKey
	result.DomainRefs = []string{original.ArtifactRef, original.Artifact.MessageRef}
	result.ConsumerOwner, result.ConsumerPending = "main", true
	mayEnd, currentErr := repair.RepairResultBlockerExecutionMayEnd(root, snapshot.State, packet.ID, owner)
	result.currentRepairResultBlocker = currentErr == nil && mayEnd
	result.WaitingOn = append(result.WaitingOn, Waiting{Reason: "original S9 blocker requires Main recovery", Owner: "Main", Ref: original.ArtifactRef})
	result.NextAction = "原阻塞答卷已接收；当前阻塞证明成立时可结束本次执行，Main 按原阻塞恢复路径接手。本人交付尚未完成。"
	if !result.currentRepairResultBlocker {
		result.NextAction = "原阻塞答卷的历史接收已保留；当前阻塞或来源证明已变化，用 runtime repair status 检查原恢复责任，历史回执不授结束或写入权限。"
	}
	if currentErr != nil {
		result.Error = "original blocker accepted; current execution-end proof unavailable: " + currentErr.Error()
	}
	return result, true, nil
}
