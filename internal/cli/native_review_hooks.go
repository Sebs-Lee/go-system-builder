package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/entroforge/go-system-builder/internal/audit"
	"github.com/entroforge/go-system-builder/internal/policy"
	"github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/submission"
	"github.com/entroforge/go-system-builder/internal/workpacket"
	"github.com/entroforge/go-system-builder/internal/workspace"
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
	snap, err := runtime.NewStore(filepath.Join(root, ".claude/loop-state.json"), filepath.Join(root, ".claude/loop-events.jsonl")).WithContext(contextForHook(in)).Snapshot()
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
			if err := audit.NewOutbox(filepath.Join(root, ".claude/hook-decisions.jsonl")).Append(engine.Envelope(in, observed, time.Now())); err != nil {
				fmt.Fprintln(diagnostic, err)
			}
		}()
	}
	if in.Event == "SubagentStart" {
		_ = json.NewEncoder(out).Encode(map[string]any{"hookSpecificOutput": map[string]any{"hookEventName": in.Event, "additionalContext": stage + " native professional work: follow the exact original questionnaire view and answer references in Main's task prompt. Read its actual frozen subjects, perform your independent review/investigation and save the professional answer at the supplied path. Return that exact reference; SubagentStop calls the original consumer. Ordinary tools do not activate or receive work. Claude Code owns execution/resume; no Team, platform ID lookup, workspace launcher or generic report is required. Real domain approval and independence remain. Unrelated native work follows its own task; do not infer a formal responsibility from your role."}})
		return true, 0
	}
	type candidate struct {
		owner, answer, source string
		packet                string
		refs                  []string
		err                   error
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
		packets, err := runtime.CurrentReviewResultPackets(snap.State, "")
		if err != nil {
			sourceErr = err
		}
		for _, p := range packets {
			selectAnswer(p.Binding.OwnerAgentID, p.Binding.AssignmentID, p.ViewPath, p.AnswerPath, p.Binding.BaselineGeneration, p.Binding.RuntimeID)
			if candidates[len(candidates)-1].err != nil {
				// Report the exact issued reference even when its execution source
				// cannot be resolved. The source error still prevents reception.
				candidates[len(candidates)-1].answer = filepath.Join(root, p.AnswerPath)
			}
		}
	case "S8":
		packets, err := runtime.CurrentHypothesisResultPackets(root, snap.State, "")
		if err != nil {
			sourceErr = err
		}
		for _, p := range packets {
			if p.Binding.RuntimeID != snap.State["runtime_id"] || p.Binding.BaselineGeneration != workspace.Generation(snap.State) || p.Binding.CaseID != investigationPointer(snap.State)["case_id"] {
				continue
			}
			selectAnswer(p.Binding.OwnerAgentID, p.Binding.AssignmentID, p.ViewPath, p.AnswerPath, p.Binding.BaselineGeneration, p.Binding.RuntimeID)
		}
	}
	var refs []string
	for _, c := range candidates {
		refs = append(refs, c.refs...)
	}
	text := nativeReviewTaskText(in, refs)
	owner, answer, source, packet := "", "", "", ""
	matched := 0
	for _, c := range candidates {
		for _, ref := range c.refs {
			if nativeS6ExactReference(text, ref) {
				matched++
				owner, answer, source, sourceErr = c.owner, c.answer, c.source, c.err
				packet = c.packet
				break
			}
		}
	}
	if matched != 1 {
		if matched > 1 {
			fmt.Fprintln(diagnostic, stage+" return has multiple exact professional responsibility references; Main must resolve the handoff")
		}
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
			return true, nativeS7HandoffFailure(owner, answer, sourceErr, diagnostic)
		}
		return true, nativeStopFailure(stage+" native professional handoff", in.StopHookActive, sourceErr, diagnostic)
	}
	if stage == "S5" {
		submission.RecordPlanningDiagnostic(root, snap.State, nativeS5DiagnosticKind(packet), "")
	}
	return true, 0
}

// A resumed native professional can receive new original task references. Use the latest
// actual task that names a current exact reference, never a tool_result or role.
// Bound transcript IO, and keep the final message for explicit handoff evidence.
func nativeReviewTaskText(in policy.Input, refs []string) string {
	f, err := os.Open(in.AgentTranscriptPath)
	if err != nil {
		return nativeS6TaskText(in)
	}
	defer f.Close()
	const limit = 4 << 20
	info, err := f.Stat()
	if err != nil {
		return nativeS6TaskText(in)
	}
	truncated := info.Size() > limit
	if truncated {
		if _, err := f.Seek(-limit, io.SeekEnd); err != nil {
			return nativeS6TaskText(in)
		}
	}
	data, err := io.ReadAll(io.LimitReader(f, limit))
	if err != nil {
		return nativeS6TaskText(in)
	}
	lines := bytes.Split(data, []byte("\n"))
	if truncated && len(lines) > 0 {
		lines = lines[1:]
	}
	latest := ""
	for _, line := range lines {
		var row struct {
			Type    string `json:"type"`
			Message struct {
				Role    string          `json:"role"`
				Content json.RawMessage `json:"content"`
			} `json:"message"`
		}
		if json.Unmarshal(line, &row) != nil || row.Type != "user" || row.Message.Role != "user" {
			continue
		}
		var task string
		if json.Unmarshal(row.Message.Content, &task) != nil {
			var parts []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}
			if json.Unmarshal(row.Message.Content, &parts) != nil {
				continue
			}
			for _, p := range parts {
				if p.Type == "text" {
					task += "\n" + p.Text
				}
			}
		}
		for _, ref := range refs {
			if nativeS6ExactReference(task, ref) {
				latest = task
				break
			}
		}
	}
	if strings.TrimSpace(latest) == "" {
		return nativeS6TaskText(in)
	}
	return in.LastAssistantMessage + "\n" + latest
}
