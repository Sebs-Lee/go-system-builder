package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/submission"
	"github.com/entroforge/go-system-builder/internal/workpacket"
	"github.com/entroforge/go-system-builder/internal/workspace"
)

func nativeS9Capture(captures []*nativeS9AttemptCapture) *nativeS9AttemptCapture {
	if len(captures) == 0 {
		return nil
	}
	return captures[0]
}

// Only the exact private native Result coordinate or original Targeted request
// enters this adapter. Typed paths retain their existing submission selector.
func consumeNativeS9Original(ctx context.Context, service *submission.Service, path string) (any, bool, error) {
	flat := strings.Contains(filepath.ToSlash(path), "/.claude/submissions/native-s9/") && filepath.Base(path) == "result.json"
	targeted := strings.Contains(filepath.ToSlash(path), "/.claude/submissions/s9/") && strings.HasPrefix(filepath.Base(path), "targeted-reverification-") && strings.HasSuffix(path, ".json")
	if !flat && !targeted {
		return nil, false, nil
	}
	snap, err := runtime.NewWriter(service.StatePath, service.JournalPath, service.Root, service.Validator).WithContext(ctx).Snapshot()
	if err != nil {
		return nil, true, err
	}
	if !runtime.PacketEnabled(snap.State) {
		return nil, true, fmt.Errorf("original S9 input has no current packet-enabled responsibility")
	}
	var feedback bytes.Buffer
	if targeted {
		selections, err := repairTargetedSelections(service.Root, snap.State)
		if err != nil {
			return nil, true, err
		}
		for _, selection := range selections {
			ref, err := repairTargetedQuestionnairePath(service.Root, snap.State, selection.performingAssignmentID)
			if err != nil {
				return nil, true, err
			}
			if filepath.Clean(ref) != filepath.Clean(path) {
				continue
			}
			if accepted, ok := registeredTargetedForSelection(service.Root, repairPointerCLI(snap.State), selection); ok {
				return map[string]any{"input_ref": path, "accepted_ref": accepted.Path, "next_action": "Targeted is already received. " + repairHandoffProjectionContext(service.Root, snap.State)}, true, nil
			}
			status := stringValue(repairPointerCLI(snap.State)["status"])
			if status != "targeted_reverification" && status != "blocked" {
				return nil, true, fmt.Errorf("Targeted answer has no current unreceived independent responsibility or accepted receipt")
			}
			capture := newNativeS9AttemptCapture(ctx, service.Root, snap, ref, selection.verifierOwner, "Targeted")
			capture.SetResponsibilityRef(ref)
			err = receiveNativeS9Targeted(ctx, service.Root, snap, ref, &feedback, capture)
			if publishErr := capture.Publish(ctx, err); publishErr != nil {
				fmt.Fprintln(&feedback, "S9 reception diagnostic could not be delivered:", publishErr)
			}
			return map[string]any{"input_ref": path, "next_action": strings.TrimSpace(feedback.String())}, true, err
		}
		return nil, true, fmt.Errorf("Targeted answer path has no exact current independent responsibility")
	}
	life := mapFieldCLI(snap.State, "lifecycle")
	if life["state"] != "bug_resolution" || !s9ProfessionalPhase(stringValue(life["phase"])) {
		return nil, true, fmt.Errorf("native Result has no current professional S9 responsibility")
	}
	var selected *workpacket.RepairPlanReportPacket
	var source string
	packets, err := runtime.CurrentRepairPlanReportPacketsContext(ctx, service.Root, snap.State, "")
	if err != nil {
		return nil, true, err
	}
	for _, packet := range packets {
		_, _, result := nativeS9Paths(packet)
		suffix := string(filepath.Separator) + filepath.FromSlash(result)
		if !strings.HasSuffix(path, suffix) {
			continue
		}
		prepared, err := nativeS9RecoveryPrepared(ctx, service.Root, snap.State, packet)
		if err != nil {
			return nil, true, err
		}
		if !prepared {
			return nil, true, fmt.Errorf("original Plan packet has no verified native preparation; preserve its frozen/legacy consumer")
		}
		if selected != nil {
			return nil, true, fmt.Errorf("native Result path selects multiple current responsibilities")
		}
		packet := packet
		selected, source = &packet, strings.TrimSuffix(path, suffix)
	}
	if selected == nil {
		return nil, true, fmt.Errorf("native Result path has no exact current original Plan packet")
	}
	// CanonicalProspective permits a cleaned accepted source; the frozen
	// execution path is still checked by the shared receiver before projection.
	source, err = workspace.CanonicalProspective(source)
	if err != nil {
		return nil, true, err
	}
	capture := newNativeS9AttemptCapture(ctx, service.Root, snap, selected.ID, selected.Binding.OwnerAgentID, "Builder")
	brief, _, _ := nativeS9Paths(*selected)
	capture.SetResponsibilityRef(filepath.Join(service.Root, brief))
	capture.SetInput(path)
	err = receiveNativeS9BuilderFrom(ctx, service.Root, snap.State, *selected, source, true, &feedback, capture)
	if publishErr := capture.Publish(ctx, err); publishErr != nil {
		fmt.Fprintln(&feedback, "S9 reception diagnostic could not be delivered:", publishErr)
	}
	return map[string]any{"input_ref": path, "next_action": strings.TrimSpace(feedback.String())}, true, err
}

// A published draft is not an adapted answer. Once adaptation has saved bytes,
// retry those bytes through the original Writer instead of replacing them from
// a mutable flat file. The Writer owns any staged/durable pending pair.
func recoverNativeS9TypedResult(ctx context.Context, service *submission.Service, state map[string]any, plan workpacket.RepairPlanReportPacket, source string, projectAccepted bool, diagnostic io.Writer, capture *nativeS9AttemptCapture) (bool, error) {
	packets, err := runtime.CurrentRepairResultPacketsContext(ctx, service.Root, state, plan.Binding.OwnerAgentID)
	if err != nil {
		return false, err
	}
	for _, packet := range packets {
		if packet.Binding.Sources.RepairAssignmentID != plan.Binding.RepairAssignmentID {
			continue
		}
		registry, err := workspace.Decode(state)
		if err != nil {
			return true, err
		}
		if registry == nil {
			return true, fmt.Errorf("original typed Result has no registered native source")
		}
		execution, ok := registry.Executions[plan.Binding.ManifestAssignmentID]
		if !ok || execution.Path != source || execution.AgentID != plan.Binding.OwnerAgentID || execution.RuntimeID != plan.Binding.RuntimeID || execution.BaselineGeneration != plan.Binding.BaselineGeneration || execution.Purpose != workspace.NativeDeliveryPurpose {
			return true, fmt.Errorf("original native Result source differs from its frozen execution")
		}
		receipt, err := service.WithContext(ctx).RepairResultReceipt(plan.Binding.OwnerAgentID, packet.ID)
		if err != nil {
			return true, err
		}
		if receipt.Accepted && (receipt.ResponsibilityStatus == "main_result_accepted" || receipt.ResponsibilityStatus == "candidate_accepted" || receipt.WorkComplete) {
			if !projectAccepted {
				return false, nil
			}
			fmt.Fprintln(diagnostic, receipt.NextAction)
			return true, nil
		}
		path := filepath.Join(source, packet.AnswerPath)
		if err := ctx.Err(); err != nil {
			return true, err
		}
		raw, err := workpacket.ReadFile(source, packet.AnswerPath, 20<<20)
		if err != nil && capture != nil {
			capture.ObserveInputError(path, err)
		}
		if os.IsNotExist(err) {
			// An issued slot may already have an immutable staged candidate even
			// before the Writer publishes its pending marker. Without its exact
			// typed bytes, a mutable flat file cannot replace that original input.
			if nativeS9SavedTypedAttempt(ctx, service.Root, state, plan, path) {
				if capture != nil {
					capture.SetTypedInput(path)
					capture.SetStage("typed_result_missing")
				}
				return true, fmt.Errorf("original saved typed Result is missing at %s; preserve staged artifacts and restore its original saved input or paired Runtime pending write: %w", path, err)
			}
			return false, nil
		}
		if err != nil {
			return true, err
		}
		if capture != nil {
			capture.SetTypedInput(path)
			capture.ObserveInput(path, raw)
		}
		if !nativeS9TypedInputSaved(packet, raw) {
			return false, nil
		}
		if capture != nil {
			capture.SetStage("typed_result_saved")
		}
		if capture != nil {
			capture.EnterConsumer("typed_result_consumer")
		}
		receipt, err = service.SaveRepairResultFrom(ctx, plan.Binding.OwnerAgentID, path, source, nil)
		if err != nil {
			return true, err
		}
		if !receipt.Accepted {
			return true, fmt.Errorf("original RepairResult was not accepted: %s", receipt.Error)
		}
		if capture != nil {
			capture.ConsumerSucceeded()
		}
		fmt.Fprintln(diagnostic, receipt.NextAction)
		return true, nil
	}
	return false, nil
}

// nativeS9TypedInputSaved compares the original issued draft with bytes already
// read by the caller. It performs no I/O and makes no receipt/pending claim.
// Malformed non-draft bytes remain the original typed consumer's responsibility.
func nativeS9TypedInputSaved(packet workpacket.RepairResultPacket, raw []byte) bool {
	draft, err := json.Marshal(packet.Draft)
	if err != nil {
		return true
	}
	var actual, empty any
	return json.Unmarshal(raw, &actual) != nil || json.Unmarshal(draft, &empty) != nil || !reflect.DeepEqual(actual, empty)
}

// This observes a prior successful typed write, including when the current
// file has disappeared. It supplies no answer bytes or source authorization.
// Recovery uses the stable original binding to name the missing
// saved input and refuses to substitute a mutable private answer.
func nativeS9SavedTypedAttempt(ctx context.Context, root string, state map[string]any, plan workpacket.RepairPlanReportPacket, path string) bool {
	if ctx.Err() != nil {
		return false
	}
	raw, err := workpacket.ReadFile(root, nativeS9DiagnosticPath(plan.ID), 64<<10)
	if err != nil {
		return false
	}
	var value nativeS9Diagnostic
	if json.Unmarshal(raw, &value) != nil || value.Runtime != stringValue(state["runtime_id"]) || value.Scope != nativeS9DiagnosticScopeContext(ctx, root, state, plan.ID) || value.Attempt == nil {
		return false
	}
	for _, observation := range []nativeS9ConsumerFailure{
		{Message: value.Message, Files: value.Files, Attempt: value.Attempt},
		func() nativeS9ConsumerFailure {
			if value.ConsumerFailure != nil {
				return *value.ConsumerFailure
			}
			return nativeS9ConsumerFailure{}
		}(),
	} {
		attempt := observation.Attempt
		if attempt == nil {
			continue
		}
		observed := observation.Files[path]
		saved := (attempt.Stage == "typed_result_saved" || attempt.Stage == "typed_result_consumer" || attempt.Stage == "result_received") && len(observed) == 64
		missingSaved := attempt.Stage == "typed_result_missing" && observed == "missing"
		if attempt.Kind == "Builder" && attempt.Owner == plan.Binding.OwnerAgentID && attempt.TypedInputRef == path && (saved || missingSaved) {
			return true
		}
	}
	return false
}

// A display brief may disappear after the Plan has been received. Select from
// original packet coordinates and native preparation evidence; an existing
// frozen independent execution always retains its original consumer.
func nativeS9RecoveryPrepared(ctx context.Context, root string, state map[string]any, plan workpacket.RepairPlanReportPacket) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	registry, err := workspace.Decode(state)
	if err != nil {
		return false, err
	}
	if registry != nil {
		if execution, ok := registry.Executions[plan.Binding.ManifestAssignmentID]; ok {
			if execution.Purpose != workspace.NativeDeliveryPurpose || execution.BootstrapSHA256 != "" || execution.PlatformSessionID != "" {
				return false, nil
			}
			b := plan.Binding
			if execution.AgentID != b.OwnerAgentID || execution.RuntimeID != b.RuntimeID || execution.BaselineGeneration != b.BaselineGeneration {
				return false, fmt.Errorf("original native preparation differs from the current Plan responsibility")
			}
			return true, nil
		}
	}
	brief, pin, _ := nativeS9Paths(plan)
	info, err := os.Lstat(filepath.Join(root, pin))
	if err == nil {
		return info.Mode().IsRegular(), nil
	}
	if !os.IsNotExist(err) {
		return false, err
	}
	raw, err := workpacket.ReadFile(root, nativeS9DiagnosticPath(plan.ID), 64<<10)
	if err != nil {
		return false, nil
	}
	var value nativeS9Diagnostic
	if json.Unmarshal(raw, &value) != nil || value.Attempt == nil || value.Runtime != stringValue(state["runtime_id"]) || value.Scope != nativeS9DiagnosticScopeContext(ctx, root, state, plan.ID) {
		return false, nil
	}
	attempt := value.Attempt
	return attempt.Kind == "Builder" && attempt.Owner == plan.Binding.OwnerAgentID && attempt.ResponsibilityRef == filepath.Join(root, brief), nil
}
