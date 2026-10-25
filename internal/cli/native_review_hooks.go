package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/entroforge/go-system-builder/internal/metrics"
	"github.com/entroforge/go-system-builder/internal/policy"
	"github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/submission"
	"github.com/entroforge/go-system-builder/internal/workpacket"
	"github.com/entroforge/go-system-builder/internal/workspace"
	"github.com/entroforge/go-system-builder/internal/workstate"
)

func nativeReviewStage(state map[string]any) string {
	if !runtime.PacketEnabled(state) {
		return ""
	}
	life := mapFieldCLI(state, "lifecycle")
	switch life["state"] {
	case "document_verification":
		return "S5"
	case "verification":
		return "S7"
	case "bug_resolution":
		if life["phase"] == "investigation" {
			return "S8"
		}
	}
	return ""
}

// This is only the platform Hook boundary for read-only professional work.
// Original packets and domain consumers own all responsibility and result facts;
// there is no native actor registry, launcher, queue, activation or new Result.
func nativeReviewHook(root string, in policy.Input, out, diagnostic io.Writer) (bool, int) {
	snap, err := nativeRoutingSnapshot(root, in)
	if err != nil {
		return false, 0
	}
	stage := nativeReviewStage(snap.State)
	if stage == "" {
		return false, 0
	}
	if in.AgentID == "" || in.AgentID == "Main" {
		if in.Event == "PostToolUse" && (in.ToolName == "Agent" || in.ToolName == "Task") {
			return true, 0
		}
		return false, 0
	}
	if prior := contextAgent(snap.State, in.AgentID); stringValue(prior["activation_ref"]) != "" {
		return false, 0 // Explicitly activated legacy responsibility, not a new prerequisite.
	}
	// A frozen independently installed session keeps its recorded recovery.
	for _, raw := range mapFieldCLI(mapFieldCLI(snap.State, "workspace"), "executions") {
		e, _ := raw.(map[string]any)
		if e["agent_id"] == in.AgentID && stringValue(e["bootstrap_sha256"]) != "" {
			return false, 0
		}
	}
	if in.Event != "SubagentStart" && in.Event != "SubagentStop" {
		return true, 0
	}
	if engine, err := policy.Load(filepath.Join(root, "docs/control/hook-policy.json")); err == nil {
		observed := policy.Decision{Decision: "audit", RuleID: "native_review_lifecycle_observed", Reason: stage + " native professional lifecycle observed", Retry: "not_applicable"}
		auditStarted := time.Now()
		defer func() {
			observed.ElapsedMS = time.Since(auditStarted).Milliseconds()
			if err := appendHookAudit(contextForHook(in), root, engine.Envelope(in, observed, time.Now())); err != nil {
				fmt.Fprintln(diagnostic, err)
			}
		}()
	}
	if in.Event == "SubagentStart" && stage != "S7" {
		_ = json.NewEncoder(out).Encode(map[string]any{"hookSpecificOutput": map[string]any{"hookEventName": in.Event, "additionalContext": stage + " native professional work: follow the exact original questionnaire view and answer references in Main's task prompt. Read its actual frozen subjects, perform your independent review/investigation and save the professional answer at the supplied path. Return that exact reference; SubagentStop calls the original consumer. Ordinary tools do not activate or receive work. Claude Code owns execution/resume; no Team, platform ID lookup, workspace launcher or generic report is required. Real domain approval and independence remain. Unrelated native work follows its own task; do not infer a formal responsibility from your role."}})
		return true, 0
	}
	type candidate struct {
		owner, answer, source string
		packet                string
		refs                  []string
		err                   error
		diagnosticOnly        bool
	}
	var candidates []candidate
	var sourceErr error
	selectAnswer := func(id, assignment, view, path string, generation int, runtimeID string, aliases ...string) {
		actual, err := questionnaireSourceRoot(root, snap.State, assignment, id, runtimeID, generation)
		refs := append([]string{filepath.Join(root, view), filepath.Join(root, path)}, aliases...)
		if err == nil {
			refs = append(refs, filepath.Join(actual, view), filepath.Join(actual, path))
		}
		candidates = append(candidates, candidate{owner: id, answer: filepath.Join(actual, path), source: actual, refs: refs, err: err})
	}
	switch stage {
	case "S5":
		for _, row := range objectSliceAny(mapFieldCLI(snap.State, "entities")["work_packets"]) {
			if row["template_id"] != "s5-document-review-v1" {
				continue
			}
			p, found, err := runtime.FindWorkPacket(snap.State, stringValue(row["packet_id"]))
			if err != nil {
				sourceErr = err
				continue
			}
			if !found {
				continue
			}
			if p.Binding.RuntimeID != snap.State["runtime_id"] || p.Binding.Generation != workspace.Generation(snap.State) || p.Binding.StageEntry != tolerantInt(mapFieldCLI(snap.State, "lifecycle")["phase_revision"]) {
				continue
			}
			if record, accepted, err := runtime.FindSubmission(snap.State, p.ID); err != nil {
				sourceErr = err
				continue
			} else if accepted && (record.RevisionPacketID != "" || record.SupersededBy != "") {
				continue
			}
			var aliases []string
			frozenAnswer, frozenSource := "", ""
			if registry, err := workspace.Decode(snap.State); err == nil && registry != nil {
				if execution, declared, err := workspace.ReviewExecution(snap.State, p); err == nil && declared && execution.Path != "" {
					if paths, err := registry.WorkerSubmissionPaths(execution); err == nil {
						frozenSource, frozenAnswer = execution.Path, filepath.Join(execution.Path, paths.Answer)
						aliases = []string{frozenAnswer, filepath.Join(execution.Path, workpacket.WorkerViewPath)}
					}
				}
			}
			selectAnswer(p.Binding.Owner, p.Binding.AssignmentID, p.ViewPath, p.AnswerPath, p.Binding.Generation, p.Binding.RuntimeID, aliases...)
			candidates[len(candidates)-1].packet = p.ID
			if frozenAnswer != "" {
				candidates[len(candidates)-1].answer, candidates[len(candidates)-1].source = frozenAnswer, frozenSource
			}
		}
	case "S7":
		packets, err := runtime.CurrentReviewResultPacketsContext(runtime.WithPacketBodyRoot(contextForHook(in), root), snap.State, "")
		if err != nil {
			sourceErr = err
		}
		for _, p := range packets {
			row := mapFieldCLI(mapFieldCLI(snap.State, "review"), "assignments")[p.Binding.AssignmentID]
			assignment, _ := row.(map[string]any)
			if in.Event == "SubagentStart" && workstate.ReviewAssignmentActivity(assignment) == "consumed" {
				continue // Reception history is not a new professional task.
			}
			selectAnswer(p.Binding.OwnerAgentID, p.Binding.AssignmentID, p.ViewPath, p.AnswerPath, p.Binding.BaselineGeneration, p.Binding.RuntimeID)
			candidates[len(candidates)-1].packet = p.ID
			if candidates[len(candidates)-1].err != nil {
				// Report the exact issued reference even when its execution source
				// cannot be resolved. The source error still prevents reception.
				candidates[len(candidates)-1].answer = filepath.Join(root, p.AnswerPath)
			}
		}
	case "S8":
		current, currentErr := runtime.CurrentHypothesisResultPackets(root, snap.State, "")
		currentIDs := map[string]bool{}
		for _, p := range current {
			currentIDs[p.ID] = true
		}
		// A damaged Case or changed question must not erase the original
		// responsibility from the return. Registry references identify only
		// the diagnostic owner here; only a current packet can be received.
		packets, err := runtime.HypothesisResultPackets(snap.State)
		if err != nil {
			sourceErr = err
		}
		for _, p := range packets {
			if p.Binding.RuntimeID != snap.State["runtime_id"] || p.Binding.BaselineGeneration != workspace.Generation(snap.State) || p.Binding.ReviewRound != tolerantInt(mapFieldCLI(snap.State, "review")["round"]) || p.Binding.CaseID != investigationPointer(snap.State)["case_id"] {
				continue
			}
			selectAnswer(p.Binding.OwnerAgentID, p.Binding.AssignmentID, p.ViewPath, p.AnswerPath, p.Binding.BaselineGeneration, p.Binding.RuntimeID)
			candidates[len(candidates)-1].packet = p.ID
			candidates[len(candidates)-1].diagnosticOnly = currentErr != nil || !currentIDs[p.ID]
			if candidates[len(candidates)-1].err == nil {
				if currentErr != nil {
					candidates[len(candidates)-1].err = currentErr
				} else if !currentIDs[p.ID] {
					candidates[len(candidates)-1].err = fmt.Errorf("S8 questionnaire %s is no longer a current responsibility; Main must inspect the original Case question, source and route before resuming; no answer was received", p.ID)
				}
			}
			if candidates[len(candidates)-1].source == "" {
				// Keep the issued reference actionable when its original source
				// is unavailable. This does not permit reception from Main.
				candidates[len(candidates)-1].answer = filepath.Join(root, p.AnswerPath)
			}
		}
	}
	var refs []string
	for _, c := range candidates {
		refs = append(refs, c.refs...)
	}
	text := nativeReviewTaskText(in, refs)
	owner, answer, source, packet := "", "", "", ""
	matched := 0
	matchedCurrent := false
	for _, c := range candidates {
		for _, ref := range c.refs {
			if nativeS6ExactReference(text, ref) {
				// A resumed task may mention a superseded question alongside
				// its current one. Historical references supply diagnostics only
				// when no current responsibility was identified.
				if c.diagnosticOnly && matchedCurrent {
					break
				}
				if !c.diagnosticOnly && !matchedCurrent {
					matched = 0
					matchedCurrent = true
				}
				matched++
				owner, answer, source, sourceErr = c.owner, c.answer, c.source, c.err
				packet = c.packet
				break
			}
		}
	}
	// An exact task reference cannot override an existing owner binding to a
	// different current packet. As in S6, surface the handoff mismatch without
	// selecting either responsibility or restricting ordinary tools.
	if stage == "S7" && in.Event == "SubagentStart" && matched == 1 {
		for _, c := range candidates {
			if c.owner == in.AgentID && c.packet != packet {
				message := fmt.Sprintf("S7 native lifecycle handoff responsibility mismatch: actor %s is the registered owner of packet %s (answer %s), but the exact task reference selects packet %s (owner %s, answer %s). Main resolves the original actor/reference conflict before resuming either responsibility. No answer was inspected or received.", in.AgentID, c.packet, c.answer, packet, owner, answer)
				_ = json.NewEncoder(out).Encode(map[string]any{"hookSpecificOutput": map[string]any{"hookEventName": in.Event, "additionalContext": message}})
				fmt.Fprintln(diagnostic, message)
				return true, 0
			}
		}
	}
	// A platform ID may be the exact registered owner. Role names and the
	// number of pending packets never establish this binding.
	if stage == "S7" && in.Event == "SubagentStart" && matched == 0 {
		for _, c := range candidates {
			if c.owner == in.AgentID {
				matched++
				owner, answer, source, sourceErr, packet = c.owner, c.answer, c.source, c.err, c.packet
			}
		}
	}
	if matched != 1 {
		if matched > 1 {
			fmt.Fprintln(diagnostic, stage+" lifecycle handoff has multiple exact professional responsibility references; Main must resolve the handoff")
		}
		return true, 0
	}
	if in.Event == "SubagentStart" {
		view := ""
		if p, found, err := runtime.FindReviewResultPacketContext(runtime.WithPacketBodyRoot(contextForHook(in), root), snap.State, packet); err == nil && found {
			view = filepath.Join(source, p.ViewPath)
			if source == "" {
				view = filepath.Join(root, p.ViewPath)
			}
		}
		message := "S7 current professional responsibility for " + owner + ": packet " + packet + "; view " + view + "; answer " + answer + ". Read that exact original questionnaire, review only its assigned Claims against the frozen subjects, save the professional answer and return that reference. SubagentStop calls the original consumer. Ordinary tools are freely available; no activation or platform ID lookup is required."
		if sourceErr != nil {
			message += " Main must restore the original source: " + nativeFailureDetail(sourceErr)
		}
		_ = json.NewEncoder(out).Encode(map[string]any{"hookSpecificOutput": map[string]any{"hookEventName": in.Event, "additionalContext": message}})
		return true, 0
	}
	if sourceErr == nil {
		service, err := submission.NewContext(contextForHook(in), root, true)
		if err != nil {
			sourceErr = err
		} else {
			var receipt submission.Receipt
			switch stage {
			case "S5":
				receipt, sourceErr = service.Save(contextForHook(in), owner, answer)
			case "S7":
				receipt, sourceErr = service.SaveReviewResultFrom(contextForHook(in), owner, answer, source, nil)
			case "S8":
				receipt, sourceErr = service.SaveHypothesisResultFrom(contextForHook(in), owner, answer, source, nil)
			}
			if len(receipt.FieldErrors) != 0 {
				sourceErr = &workpacket.AnswerError{Label: stage + " answer", Fields: receipt.FieldErrors}
			}
			if sourceErr == nil && !receipt.Accepted {
				sourceErr = fmt.Errorf("original professional consumer did not accept this answer: %s", receipt.Error)
			}
			if sourceErr == nil {
				fmt.Fprintln(diagnostic, stage+" native professional answer received by the original consumer. "+receipt.NextAction)
			}
		}
	}
	if sourceErr != nil {
		if stage == "S5" {
			submission.RecordPlanningDiagnostic(root, snap.State, nativeS5DiagnosticKind(packet), fmt.Sprintf("owner %s; answer %s: %s", owner, answer, nativeFailureDetail(sourceErr)))
			return true, nativeS5HandoffFailure(owner, answer, sourceErr, diagnostic)
		}
		if stage == "S7" {
			submission.RecordPlanningDiagnostic(root, snap.State, nativeS7DiagnosticKind(packet), fmt.Sprintf("owner %s; packet %s; answer %s: %s", owner, packet, answer, nativeFailureDetail(sourceErr)))
			return true, nativeS7HandoffFailure(owner, answer, sourceErr, diagnostic)
		}
		if stage == "S8" {
			submission.RecordPlanningDiagnostic(root, snap.State, "S8 native Stop "+packet, fmt.Sprintf("owner %s; answer %s: %s", owner, answer, nativeFailureDetail(sourceErr)))
			return true, nativeS8HandoffFailure(owner, answer, sourceErr, diagnostic)
		}
		return true, nativeStopFailure(stage+" native professional handoff", in.StopHookActive, sourceErr, diagnostic)
	}
	if stage == "S5" {
		submission.RecordPlanningDiagnostic(root, snap.State, nativeS5DiagnosticKind(packet), "")
	}
	if stage == "S7" {
		submission.RecordPlanningDiagnostic(root, snap.State, nativeS7DiagnosticKind(packet), "")
		// Reception is already durable. Continue only the original clean-round
		// exit; a continuation failure never turns it into an unreceived answer.
		statePath := filepath.Join(root, ".claude/loop-state.json")
		journalPath := filepath.Join(root, ".claude/loop-events.jsonl")
		completionContext, cancelCompletion := hookCompletionContext(contextForHook(in))
		defer cancelCompletion()
		current, err := runtime.NewStore(statePath, journalPath).WithContext(completionContext).Snapshot()
		if err != nil {
			fmt.Fprintln(diagnostic, "S7 ReviewResult is received; Main restores current Runtime access before continuing the original clean-round exit: "+err.Error()+". Recovery: "+s7CaptureShellQuote(filepath.Join(root, ".claude/bin/loop-harness"))+" runtime review-continue --root "+s7CaptureShellQuote(root))
		} else if _, handoff := continueS7Acceptance(completionContext, root, statePath, journalPath, current); handoff != "" {
			fmt.Fprintln(diagnostic, handoff)
		}
	}
	if stage == "S8" {
		submission.RecordPlanningDiagnostic(root, snap.State, "S8 native Stop "+packet, "")
	}
	return true, 0
}

// A resumed native professional can receive new original task references. Use the latest
// actual task that names a current exact reference, never a tool_result or role.
// Bound transcript IO, and keep the final message for explicit handoff evidence.
func nativeReviewTaskText(in policy.Input, refs []string) string {
	defer metrics.Span(contextForHook(in), "hook/native_routes/task_text")()
	text, _ := nativeTaskTextDiagnostic(in, refs, true)
	return text
}
