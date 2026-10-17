package submission

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/entroforge/go-system-builder/internal/assignment"
	"github.com/entroforge/go-system-builder/internal/integration"
	"github.com/entroforge/go-system-builder/internal/repair"
	loopruntime "github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/workpacket"
)

// RepairResultResolutionForPath selects an exact Main authoring path. Selection
// grants no identity, source authority, or permission to resume the Worker.
func (s *Service) RepairResultResolutionForPath(path string) (workpacket.RepairResultPacket, bool, error) {
	snapshot, err := s.snapshot()
	if err != nil {
		return workpacket.RepairResultPacket{}, false, err
	}
	return repairResultResolutionForPath(s.Root, snapshot.State, path)
}

func repairResultResolutionForPath(root string, state map[string]any, inputPath string) (workpacket.RepairResultPacket, bool, error) {
	if !filepath.IsAbs(inputPath) {
		inputPath = filepath.Join(root, filepath.FromSlash(inputPath))
	}
	abs, err := filepath.Abs(inputPath)
	if err != nil {
		return workpacket.RepairResultPacket{}, false, err
	}
	packets, err := loopruntime.RepairResultPackets(state)
	if err != nil {
		return workpacket.RepairResultPacket{}, false, err
	}
	var selected *workpacket.RepairResultPacket
	for _, packet := range packets {
		rel, err := repair.RepairResultResolutionAuthoringPath(packet)
		if err != nil {
			return workpacket.RepairResultPacket{}, false, err
		}
		if filepath.Clean(abs) != filepath.Clean(filepath.Join(root, filepath.FromSlash(rel))) {
			continue
		}
		if err := workpacket.CheckPath(root, rel); err != nil {
			return workpacket.RepairResultPacket{}, false, err
		}
		if selected != nil {
			return workpacket.RepairResultPacket{}, false, errors.New("S9 Main resolution path has ambiguous original issuance")
		}
		copy := packet
		selected = &copy
	}
	if selected == nil {
		return workpacket.RepairResultPacket{}, false, nil
	}
	return *selected, true, nil
}

// SaveRepairResultResolutionFrom consumes the exact registered Main authoring
// path. The packet, accepted blocker, typed resolution and Agent lifecycle CAS
// own domain acceptance; authorizeMain remains optional for compatibility.
func (s *Service) SaveRepairResultResolutionFrom(ctx context.Context, path string, authorizeMain func(string, map[string]any) error) (Receipt, error) {
	ctx, unlock, err := integration.LockWorkspace(ctx, s.Root)
	if err != nil {
		return Receipt{}, err
	}
	defer unlock()
	snapshot, err := s.snapshot()
	if err != nil {
		return Receipt{}, err
	}
	packet, found, err := repairResultResolutionForPath(s.Root, snapshot.State, path)
	if err != nil || !found {
		if err == nil {
			err = errors.New("S9 Main resolution path has no exact original issuance")
		}
		return rejected(snapshot, packet.ID, err), err
	}
	// Preparation reopens the locked snapshot and verifies the original packet,
	// blocker and authored decision before publishing immutable message bytes.
	rel, err := repair.RepairResultResolutionAuthoringPath(packet)
	if err != nil {
		return rejected(snapshot, packet.ID, err), err
	}
	owner := packet.Binding.Sources.OwnerAgentID
	capability, prior, err := repair.PrepareIssuedRepairResultResolution(ctx, s.Root, s.StatePath, s.JournalPath, repair.PrepareIssuedRepairResultResolutionRequest{
		ExpectedRevision: snapshot.Revision, PacketID: packet.ID, OwnerAgentID: owner,
		AuthoringPath: rel, AuthorizeMain: authorizeMain,
	})
	if err != nil {
		return rejected(snapshot, packet.ID, err), err
	}
	if prior == nil {
		if capability == nil {
			err := errors.New("original S9 Main resolution preparation returned no capability")
			return rejected(snapshot, packet.ID, err), err
		}
		next, advanceErr := assignment.AdvanceAgentWithContext(ctx, s.Root, s.StatePath, s.JournalPath, assignment.AgentEventRequest{
			ExpectedRevision: snapshot.Revision, AgentID: owner, Event: "blocker_resolved",
			MessagePath:                  filepath.Join(s.Root, filepath.FromSlash(capability.MessageRef())),
			IssuedRepairResultResolution: capability,
		})
		if advanceErr != nil {
			return rejected(snapshot, packet.ID, advanceErr), advanceErr
		}
		snapshot = next
	} else {
		snapshot, err = s.snapshot()
		if err != nil {
			return Receipt{}, err
		}
	}
	return repairResultResolutionReceipt(s.Root, snapshot, packet, owner)
}

func repairResultResolutionReceipt(root string, snapshot loopruntime.Snapshot, packet workpacket.RepairResultPacket, owner string) (Receipt, error) {
	result := emptyReceipt(snapshot, packet.ID)
	original, accepted, err := loopruntime.VerifiedRepairResultResolutionReceipt(root, snapshot.State, packet.ID, owner)
	if err != nil {
		return result, err
	}
	if !accepted {
		return result, errors.New("original S9 Main resolution CAS has no verifiable receipt")
	}
	result.Accepted = true
	result.ResponsibilityStatus = "blocker_resolution_recorded"
	result.SubmissionKey = original.AcceptedIdempotencyKey
	result.WorkComplete = false
	result.ConsumerOwner, result.ConsumerPending = "main", true
	result.DomainRefs = []string{original.MessageRef, original.BlockerRef}
	result.WaitingOn = []Waiting{{Reason: "a new verified Resume successor is required", Owner: "Main", Ref: packet.ID}}
	result.NextAction = "Main 恢复决定已由原生命周期接收；原责任尝试已结束，旧答卷路径不再接收。Main 下一次进入 SessionStart 或 PreToolUse 时，原监听器会准备继任 RepairResult 的准确问卷路径；Worker 按新路径继续。"
	return result, nil
}

// RepairResultResolutionReceipt reads the original Main decision without
// recovering a pending transaction or granting the retired Worker an action.
func (s *Service) RepairResultResolutionReceipt(path string) (Receipt, error) {
	snapshot, err := s.snapshot()
	if err != nil {
		return Receipt{}, err
	}
	packet, found, err := repairResultResolutionForPath(s.Root, snapshot.State, path)
	if err != nil || !found {
		if err == nil {
			err = errors.New("S9 Main resolution path has no original packet")
		}
		return rejected(snapshot, packet.ID, err), err
	}
	return repairResultResolutionReceipt(s.Root, snapshot, packet, packet.Binding.Sources.OwnerAgentID)
}

// Publication is an authoring projection after the original blocker commit.
// Existing Main edits survive exact Worker replays; publication owns no CAS.
func publishRepairResultResolutionDraft(root string, snapshot loopruntime.Snapshot, packet workpacket.RepairResultPacket, owner string) (string, error) {
	path, data, err := repair.RepairResultResolutionDraft(root, snapshot.State, packet.ID, owner)
	if err != nil {
		return "", err
	}
	if path == "" || len(data) == 0 {
		return "", fmt.Errorf("original S9 blocker has no Main resolution draft")
	}
	_, err = workpacket.PublishBounded(root, path, data, repairResultAnswerLimit)
	return path, err
}
