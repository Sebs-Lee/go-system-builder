package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/entroforge/go-system-builder/internal/assignmentrecord"
	"github.com/entroforge/go-system-builder/internal/policy"
	"github.com/entroforge/go-system-builder/internal/repair"
	"github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/submission"
	"github.com/entroforge/go-system-builder/internal/workpacket"
	"github.com/entroforge/go-system-builder/internal/workspace"
	"github.com/entroforge/go-system-builder/internal/workstate"
)

// This is a read-only projection into Main's existing tool-return context.
// Stop remains the only native answer receiver; no actor mapping or ACK is added.
type nativeReturnSubject struct {
	kind, id, ref string
	record        *assignmentrecord.Record
	targeted      *repairTargetedSelection
}

func nativeMainReturnHook(root string, in policy.Input, out io.Writer) (bool, int) {
	if in.Event != "PostToolUse" || (in.ToolName != "Agent" && in.ToolName != "Task") || in.AgentID != "" && in.AgentID != "Main" || in.TeammateName != "" {
		return false, 0
	}
	if cwd, err := workspace.Canonical(in.CWD); err != nil || cwd != root {
		return false, 0
	}
	if _, err := os.Stat(filepath.Join(root, ".claude/loop-state.json")); err != nil {
		return false, 0
	}
	snap, err := runtime.NewStore(filepath.Join(root, ".claude/loop-state.json"), filepath.Join(root, ".claude/loop-events.jsonl")).Snapshot()
	if err != nil {
		return false, 0
	}
	selectSubjects := func(text string) []nativeReturnSubject {
		var selected []nativeReturnSubject
		for _, row := range objectSliceAny(mapFieldCLI(snap.State, "entities")["work_packets"]) {
			kind := stringValue(row["template_id"])
			switch kind {
			case "s5-document-review-v1", "s7-review-result-v1", "s8-hypothesis-result-v1", "s9-plan-report-v1", "s9-repair-result-v1":
			default:
				continue
			}
			id, view, answer := stringValue(row["packet_id"]), stringValue(row["view_path"]), stringValue(row["answer_path"])
			refs := []string{view, answer}
			if kind == "s9-plan-report-v1" {
				brief, _, _ := nativeS9Paths(workpacket.RepairPlanReportPacket{ID: id, ViewPath: view})
				refs = append(refs, brief)
			}
			for _, ref := range refs {
				if ref != "" && nativeS6ExactReference(text, filepath.Join(root, ref)) {
					selected = append(selected, nativeReturnSubject{kind: kind, id: id, ref: filepath.Join(root, ref)})
					break
				}
			}
		}
		// Do not filter by the new lifecycle: the returned work may just have
		// advanced that lifecycle, while its original reception remains readable.
		if records, err := assignmentrecord.RecordsFromState(snap.State); err == nil {
			for _, r := range records {
				if r.Contract != assignmentrecord.NativeDeliveryContract {
					continue
				}
				brief, _, _ := assignmentrecord.NativePaths(r)
				if nativeS6ExactReference(text, filepath.Join(root, brief)) {
					r := r
					selected = append(selected, nativeReturnSubject{kind: "S6", id: r.AssignmentID, ref: filepath.Join(root, brief), record: &r})
				}
			}
		}
		if stringValue(repairPointerCLI(snap.State)["impact_ref"]) != "" {
			if selections, err := repairTargetedSelections(root, snap.State); err == nil {
				for _, selection := range selections {
					ref, err := repairTargetedQuestionnairePath(root, snap.State, selection.performingAssignmentID)
					if err == nil && nativeS6ExactReference(text, ref) {
						selection := selection
						selected = append(selected, nativeReturnSubject{kind: "Targeted", id: selection.performingAssignmentID, ref: ref, targeted: &selection})
					}
				}
			}
		}
		return selected
	}
	prompt, _ := in.ToolInput["prompt"].(string)
	selected := selectSubjects(prompt)
	if len(selected) == 0 {
		text, _ := in.ToolResponse["text"].(string)
		switch content := in.ToolResponse["content"].(type) {
		case string:
			text += "\n" + content
		case []any:
			for _, raw := range content {
				part, _ := raw.(map[string]any)
				if part["type"] == "text" {
					value, _ := part["text"].(string)
					text += "\n" + value
				}
			}
		}
		selected = selectSubjects(text)
	}
	if len(selected) == 0 {
		return false, 0
	}
	message := "Native return has multiple exact responsibility references; Main must resolve the handoff. No answer was received by this observation."
	if len(selected) == 1 {
		subject := selected[0]
		summary, err := nativeReturnSummary(root, snap, subject)
		if err != nil {
			summary = "Original receipt unavailable; Main preserves this return and inspects the original source/Stop failure before deciding recovery: " + err.Error()
		}
		message = "Native return for " + subject.ref + ". " + summary + " This is a read-only reception summary, not product acceptance. Do not re-dispatch received work merely because Main or the next stage still has work."
	}
	if err := json.NewEncoder(out).Encode(map[string]any{"hookSpecificOutput": map[string]any{"hookEventName": "PostToolUse", "additionalContext": message}}); err != nil {
		return true, 1
	}
	return true, 0
}

func nativeReturnSummary(root string, snap runtime.Snapshot, subject nativeReturnSubject) (string, error) {
	if subject.record != nil {
		r := *subject.record
		agent := contextAgent(snap.State, r.OwnerAgentID)
		ref := stringValue(agent["completion_reported_ref"])
		if ref == "" {
			return "No Result reception is recorded for " + r.Task.ID + ". A background launch can return before completion; Main continues other work and checks the original native execution/Stop diagnostics rather than guessing or re-dispatching.", nil
		}
		current, found, err := assignmentrecord.ForOwner(snap.State, r.OwnerAgentID)
		if err != nil || !found || current.Reference() != r.Reference() || r.RuntimeID != snap.State["runtime_id"] || r.BaselineGeneration != workspace.Generation(snap.State) {
			return "", fmt.Errorf("this exact TASK reference is historical or replaced; Main follows the original current revision rather than applying another Result to it")
		}
		if workstate.AgentActivity(agent) == "cancelled" || workstate.AgentActivity(agent) == "stopped" {
			return "Original TASK is closed; Main preserves its history without re-dispatching it.", nil
		}
		return nativeS6RecordGuidance(root, snap.State, r) + " This projects the recorded reference; original task-receipt/integration checks remain authoritative.", nil
	}
	if subject.targeted != nil {
		var records []string
		for _, ref := range repairArtifactRefs(repairPointerCLI(snap.State), "targeted_reverification_artifacts", "", "") {
			var value repair.TargetedReverification
			if err := repairLeadIndexedSource(root, ref, &value); err != nil {
				return "", err
			}
			if value.ImpactID == subject.targeted.impact.ImpactID && value.OriginalAssignmentID == subject.targeted.originalAssignmentID && value.PerformingAssignmentID == subject.id {
				records = append(records, filepath.Join(root, ref.Path)+" result="+value.Result)
			}
		}
		if len(records) == 0 {
			return "No original Targeted reception recorded; Main checks the actual native execution/Stop diagnostics without re-dispatching by guess.", nil
		}
		return "Original Targeted records: " + strings.Join(records, "; ") + ". Main follows the original repair status " + stringValue(repairPointerCLI(snap.State)["status"]) + " and Handoff to the next complete S7 round; a Targeted result does not replace S7.", nil
	}
	s, err := submission.NewReadOnly(root)
	if err != nil {
		return "", err
	}
	var receipt submission.Receipt
	prefix := ""
	switch subject.kind {
	case "s5-document-review-v1":
		p, _, e := runtime.FindWorkPacket(snap.State, subject.id)
		if e != nil {
			return "", e
		}
		receipt, err = s.ReceiptPacket(p.Binding.Owner, p.ID)
		if err == nil && !receipt.Accepted {
			prefix, receipt.NextAction = nativeS5ReturnDiagnostic(root, snap.State, p)
		}
	case "s7-review-result-v1":
		p, _, e := runtime.FindReviewResultPacket(snap.State, subject.id)
		if e != nil {
			return "", e
		}
		receipt, err = s.ReviewResultReceipt(p.Binding.OwnerAgentID, p.ID)
	case "s8-hypothesis-result-v1":
		p, _, e := runtime.FindHypothesisResultPacket(snap.State, subject.id)
		if e != nil {
			return "", e
		}
		receipt, err = s.HypothesisResultReceipt(p.Binding.OwnerAgentID, p.ID)
	case "s9-repair-result-v1":
		p, _, e := runtime.FindRepairResultPacket(snap.State, subject.id)
		if e != nil {
			return "", e
		}
		receipt, err = s.RepairResultReceipt(p.Binding.Sources.OwnerAgentID, p.ID)
	case "s9-plan-report-v1":
		p, _, e := runtime.FindRepairPlanReportPacket(snap.State, subject.id)
		if e != nil {
			return "", e
		}
		packets, e := runtime.CurrentRepairResultPackets(root, snap.State, p.Binding.OwnerAgentID)
		if e != nil {
			return "", e
		}
		var ids []string
		for _, result := range packets {
			if result.Binding.Sources == p.Binding {
				ids = append(ids, result.ID)
			}
		}
		if len(ids) > 1 {
			return "", fmt.Errorf("multiple current Result packets for this exact repair; Main must resolve the original attempt")
		}
		if len(ids) == 1 {
			receipt, err = s.RepairResultReceipt(p.Binding.OwnerAgentID, ids[0])
		} else {
			receipt, err = s.RepairPlanReportReceipt(p.Binding.OwnerAgentID, p.ID)
			prefix = "Only the PlanReport receipt is available, not the Builder Result. Main inspects the original Stop/source failure; do not re-submit the accepted PlanReport. "
		}
	}
	if err != nil {
		return "", err
	}
	if !receipt.Accepted && subject.kind != "s5-document-review-v1" {
		receipt.NextAction = "尚无原接收事实；后台派发即时返回不代表执行结束。Main 继续可开展的工作，核对实际原生执行与 Stop 诊断，不凭此次返回猜测重派或要求重填。"
		if subject.kind == "s7-review-result-v1" {
			receipt.NextAction += " S7 实际未完成返回仍属于原责任；Main 阅读实际返回和 Stop 的具名作者、原答案路径及具体字段/来源诊断，恢复来源或环境前提，或续跑原 Reviewer 修正专业缺项。已记录 site_lost 时先用原 blocker_resolved 入口解除真实阻塞，再由同一 Reviewer 恢复。不能把已结束但未接收的责任当作仍在后台运行，也不凭 prose 推定验收、重派或编造 verdict/Finding。"
		}
	}
	data, err := json.Marshal(receipt)
	return prefix + string(data), err
}
