package cli

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/entroforge/go-system-builder/internal/policy"
	"github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/submission"
)

func stageRepairResultQuestionnaireContext(ctx context.Context, root string, request policy.Input, service *submission.Service, snapshot runtime.Snapshot, owner string) string {
	if !runtime.PacketEnabled(snapshot.State) {
		return ""
	}
	lifecycle, _ := snapshot.State["lifecycle"].(map[string]any)
	if lifecycle["state"] != "bug_resolution" || lifecycle["phase"] != "fixing" {
		return ""
	}
	if service == nil {
		return "S9 RepairResult status is unavailable; Main must inspect the existing packet before continuing."
	}

	mainSession := owner == "" || owner == "Main"
	state := snapshot.State
	if mainSession && (request.Event == "SessionStart" || request.Event == "PreCompact" || request.Event == "PreToolUse") {
		if _, err := service.IssueRepairResults(ctx, ""); err != nil {
			return "S9 RepairResult issuance is unavailable: " + err.Error()
		}
		latest, err := runtime.NewStore(service.StatePath, service.JournalPath).Snapshot()
		if err != nil {
			return "S9 RepairResult was issued, but its current Runtime projection is unavailable: " + err.Error()
		}
		state = latest.State
	}
	ownerFilter := owner
	if mainSession {
		ownerFilter = ""
	}
	packets, err := runtime.CurrentRepairResultPackets(root, state, ownerFilter)
	if err != nil {
		return "S9 current RepairResult selection is unavailable: " + err.Error()
	}
	if len(packets) == 0 {
		if mainSession {
			return "S9 Main: no current Builder RepairResult questionnaire is available; inspect the original repair dispatch and requirements. Do not guess an answer path."
		}
		return fmt.Sprintf("S9 Builder owner=%s has no current RepairResult questionnaire. Main owns the next repair dispatch; do not guess an answer path.", owner)
	}

	var context strings.Builder
	for _, packet := range packets {
		packetOwner := packet.Binding.Sources.OwnerAgentID
		sourceRoot, _, err := submission.RepairResultRegisteredSource(root, state, packet)
		if err != nil {
			fmt.Fprintf(&context, "S9 Builder %s assignment=%s: registered Worker answer source is unavailable; Main must restore that exact execution before authoring: %v\n", packetOwner, packet.Binding.Sources.RepairAssignmentID, err)
			continue
		}
		view := filepath.Join(sourceRoot, filepath.FromSlash(packet.ViewPath))
		answer := filepath.Join(sourceRoot, filepath.FromSlash(packet.AnswerPath))
		receipt, err := service.RepairResultReceipt(packetOwner, packet.ID)
		if err != nil {
			fmt.Fprintf(&context, "S9 Builder %s assignment=%s packet=%s questionnaire=%s: receipt status unavailable; Main must inspect the original packet: %v\n", packetOwner, packet.Binding.Sources.RepairAssignmentID, packet.ID, answer, err)
			continue
		}
		if receipt.Accepted {
			next := strings.TrimSpace(receipt.NextAction)
			if next == "" {
				next = "Main owns the next integration step."
			}
			fmt.Fprintf(&context, "S9 Builder %s assignment=%s packet=%s: questionnaire %s already accepted; work_complete=%t. Do not fill or submit again. Next: %s\n", packetOwner, packet.Binding.Sources.RepairAssignmentID, packet.ID, answer, receipt.WorkComplete, next)
			continue
		}
		fmt.Fprintf(&context, "S9 Builder %s assignment=%s: read the issued packet at %s; fill only its professional answer at %s, then save. This is the exact registered Worker path; Main owns integration and acceptance.\n", packetOwner, packet.Binding.Sources.RepairAssignmentID, view, answer)
	}
	return strings.TrimSpace(context.String())
}
