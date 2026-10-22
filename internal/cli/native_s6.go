package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/entroforge/go-system-builder/internal/agentcontext"
	"github.com/entroforge/go-system-builder/internal/assignment"
	"github.com/entroforge/go-system-builder/internal/assignmentrecord"
	"github.com/entroforge/go-system-builder/internal/audit"
	"github.com/entroforge/go-system-builder/internal/fileview"
	"github.com/entroforge/go-system-builder/internal/policy"
	"github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/schema"
	"github.com/entroforge/go-system-builder/internal/workpacket"
	"github.com/entroforge/go-system-builder/internal/workspace"
	"github.com/entroforge/go-system-builder/internal/workstate"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
)

func pinNativeS6Source(root, id string, revision int, base string) error {
	body, _ := json.Marshal(map[string]any{"base_commit": base})
	ref := fmt.Sprintf(".claude/assignments/native/%s/r%d/source.json", id, revision)
	published, err := workpacket.PublishBounded(root, ref, body, 4096)
	if err == nil && !bytes.Equal(published, body) {
		return fmt.Errorf("original native preparation source commit differs; preserve it for Main recovery")
	}
	return err
}

// Resume the original prepared input after a registration failure. Never
// overwrite its draft or recapture a different source commit.
func resumeNativeS6Preparation(root, req, task, role, owner, output string, out, diagnostic io.Writer) (bool, int) {
	ref := output
	if ref == "" {
		ref = filepath.ToSlash(filepath.Join(".claude/workgroups", req, task, "manifest.draft.json"))
	}
	data, err := workpacket.ReadFile(root, ref, 1<<20)
	if os.IsNotExist(err) {
		return false, 0
	}
	if err != nil {
		fmt.Fprintln(diagnostic, err)
		return true, 1
	}
	var draft s6ManifestDraft
	if json.Unmarshal(data, &draft) != nil || draft.AssignmentRecordContract != assignmentrecord.NativeDeliveryContract || len(draft.Assignments) != 1 || draft.Assignments[0].AgentID != owner || draft.Assignments[0].RoleFamily != role {
		fmt.Fprintln(diagnostic, "existing preparation has a different frozen contract; preserve it for Main recovery")
		return true, 1
	}
	source, err := workpacket.ReadFile(root, fmt.Sprintf(".claude/assignments/native/%s/r1/source.json", draft.Assignments[0].AssignmentID), 4096)
	var pin struct {
		Base string `json:"base_commit"`
	}
	if err != nil || json.Unmarshal(source, &pin) != nil || pin.Base == "" {
		fmt.Fprintln(diagnostic, "original preparation source commit is unavailable; Main must recover it, never recapture HEAD")
		return true, 1
	}
	for _, doc := range draft.Documents {
		if doc.ID == task {
			return true, registerNativeS6Delivery(root, ref, task, doc.Path, pin.Base, out, diagnostic)
		}
	}
	fmt.Fprintln(diagnostic, "original preparation has no exact TASK source")
	return true, 1
}

func registerNativeS6Delivery(root, manifest, taskID, taskPath, base string, stdout, stderr io.Writer) int {
	data, err := workpacket.ReadFile(root, manifest, 1<<20)
	var draft s6ManifestDraft
	if err != nil || json.Unmarshal(data, &draft) != nil || len(draft.Assignments) != 1 {
		fmt.Fprintln(stderr, "native preparation has no exact manifest Assignment")
		return 1
	}
	if err := pinNativeS6Source(root, draft.Assignments[0].AssignmentID, 1, base); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	statePath, journal := filepath.Join(root, ".claude/loop-state.json"), filepath.Join(root, ".claude/loop-events.jsonl")
	snap, err := assignment.Register(root, statePath, journal, assignment.Request{ExpectedRevision: -1, ManifestPath: filepath.Join(root, manifest), TaskID: taskID, TaskPath: filepath.Join(root, taskPath)})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	records, err := assignmentrecord.RecordsFromState(snap.State)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	for _, record := range records {
		if record.Task.ID == taskID && record.Contract == assignmentrecord.NativeDeliveryContract {
			if err := prepareNativeS6Brief(root, snap.State, record, base, stdout); err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
			return 0
		}
	}
	fmt.Fprintln(stderr, "native TASK registration has no current Record")
	return 1
}

func prepareNativeS6Brief(root string, state map[string]any, r assignmentrecord.Record, base string, out io.Writer) error {
	brief, plan, result := assignmentrecord.NativePaths(r)
	var b strings.Builder
	fmt.Fprintf(&b, "# %s — %s\n\nContract: %s\n\nRead the actual TASK at %s and its linked inputs. Implement only its reviewed scope: %s.\n\n", r.Task.ID, r.RoleFamily, r.Contract, filepath.Join(root, r.Task.Ref), strings.Join(r.EffectiveWritePaths, ", "))
	for _, ref := range r.ReadPaths {
		fmt.Fprintf(&b, "- Read %s\n", filepath.Join(root, ref))
	}
	fmt.Fprintf(&b, "\nDone when:\n")
	for _, condition := range r.DoneWhen {
		fmt.Fprintf(&b, "- %s\n", condition)
	}
	fmt.Fprintf(&b, "\nRun the actual TASK self-checks (Main repeats its checks at integration):\n")
	for _, check := range r.RequiredChecks {
		fmt.Fprintf(&b, "- %s\n", check)
	}
	fmt.Fprintf(&b, "\nUse Claude Code native worktree isolation from Main HEAD %s. Save the following professional JSON files relative to your actual worktree; do not write these drafts into Main. Plan before implementation; save does not activate anything. Implement, self-test, commit only actual TASK products and save Result. Keep these private professional files out of Git; do not force-add them. Return this exact brief reference %s and the absolute path to your Result. Do not run workspace launch, agent-begin or task-complete.\n\nPlan: %s\n```json\n{\"objective\":null,\"planned_paths\":[],\"assertion_checks\":[{\"assertion\":null,\"oracle\":null,\"done_when_indexes\":[]}],\"dependencies\":[],\"risks_blockers\":[]}\n```\nEach assertion has assertion, oracle and done_when_indexes (zero-based indices into Done when). Cover each required condition.\n\nResult: %s\n```json\n{\"summary\":null,\"assertion_evidence\":[{\"assertion\":null,\"oracle\":null,\"evidence_refs\":[]}],\"risks_blockers\":[],\"scope_deviations\":[],\"closing_assessment\":{\"disposition\":null,\"conditions\":[{\"done_when_index\":0,\"judgment\":null,\"plan_assertion_indexes\":[],\"rationale\":null}]}}\n```\nUse exactly the shown field names; no extra nested fields. Each assertion_evidence item has assertion and oracle copied verbatim from its plan assertion, plus nonempty evidence_refs containing exact existing file paths relative to your worktree, within this TASK write scope or its listed Read paths. These are file references: omit line numbers, fragments, descriptions, command text and output. Put commands, outputs and observations in summary or closing rationale; cite their actual evidence file only when it exists within the allowed scope. dependencies, risks_blockers and scope_deviations are arrays of strings. Replace nulls with your own professional answers; do not copy placeholder values. closing_assessment has disposition (complete/incomplete/blocked) and conditions; each condition has done_when_index, judgment (satisfied/incomplete/blocked), plan_assertion_indexes and rationale. Include real self-check commands/results in summary. When blocked, also state the actual blocker and unfinished work directly in your native return. A real blocker is an honest handoff, never a fabricated PASS.\n", base, filepath.Join(root, brief), plan, result)
	if published, err := workpacket.PublishBounded(root, brief, []byte(b.String()), 1<<20); err != nil {
		return err
	} else if !bytes.Equal(published, []byte(b.String())) {
		return fmt.Errorf("original native task brief differs; preserve it for Main recovery")
	}
	bound, _ := state["bound_req"].(map[string]any)
	binding, _ := bound["workspace"].(map[string]any)
	coordinates := map[string]any{"assignment_id": r.AssignmentID, "assignment_revision": r.Revision, "base_commit": base, "target_branch": binding["dev_branch"], "worktree_path": "", "branch": ""}
	data, _ := json.MarshalIndent(coordinates, "", "  ")
	coordinatePath := filepath.Join(root, ".claude/assignments", r.AssignmentID+".json")
	previousBytes, readErr := os.ReadFile(coordinatePath)
	var previous struct {
		Revision int `json:"assignment_revision"`
	}
	_ = json.Unmarshal(previousBytes, &previous)
	if os.IsNotExist(readErr) || (readErr == nil && previous.Revision < r.Revision) {
		if err := os.WriteFile(coordinatePath, data, 0600); err != nil {
			return err
		}
	}
	fmt.Fprintf(out, "Native S6 TASK registered: %s\nMain: use Claude Code's native Agent (%s) with worktree isolation. Carry this exact task brief in its prompt: %s\nBuilder saves professional plan/result in its own worktree. SubagentStop receives them; Main integrates separately. No Team, platform Agent ID or launcher is required.\n", r.Task.ID, r.RoleFamily, filepath.Join(root, brief))
	return nil
}

func nativeS6Records(state map[string]any) []assignmentrecord.Record {
	life, _ := state["lifecycle"].(map[string]any)
	if life["state"] != "building" {
		return nil
	}
	records, err := assignmentrecord.RecordsFromState(state)
	if err != nil {
		return nil
	}
	var out []assignmentrecord.Record
	for _, r := range records {
		if current, found, err := assignmentrecord.ForOwner(state, r.OwnerAgentID); err == nil && found && current.Reference() == r.Reference() && r.Contract == assignmentrecord.NativeDeliveryContract {
			out = append(out, r)
		}
	}
	return out
}

// A known legacy owner keeps its frozen contract. No native actor is rewritten.
func nativeS6Worker(state map[string]any, in policy.Input) bool {
	if in.AgentID == "" {
		return false
	}
	if r, found, _ := assignmentrecord.ForOwner(state, in.AgentID); found {
		return r.Contract == assignmentrecord.NativeDeliveryContract
	}
	return len(nativeS6Records(state)) > 0
}

// The initial task prompt supplies the exact brief. Roles/queue order never bind it.
func nativeS6ExactReference(text, ref string) bool {
	for offset := 0; offset < len(text); {
		index := strings.Index(text[offset:], ref)
		if index < 0 {
			return false
		}
		index += offset
		end := index + len(ref)
		before := index == 0 || strings.ContainsRune(" \t\r\n`\"'(<[{=", rune(text[index-1]))
		after := end == len(text) || strings.ContainsRune(" \t\r\n`\"')>]},;:", rune(text[end]))
		if before && after {
			return true
		}
		offset = end
	}
	return false
}

func nativeS6TaskText(in policy.Input) string {
	text := in.LastAssistantMessage
	if in.AgentTranscriptPath == "" {
		return text
	}
	f, err := os.Open(in.AgentTranscriptPath)
	if err != nil {
		return text
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, 4<<20))
	if err != nil {
		return text
	}
	for _, line := range bytes.Split(data, []byte("\n")) {
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
		var plain string
		if json.Unmarshal(row.Message.Content, &plain) == nil {
			return text + "\n" + plain
		}
		var parts []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}
		if json.Unmarshal(row.Message.Content, &parts) == nil {
			for _, p := range parts {
				if p.Type == "text" {
					text += "\n" + p.Text
				}
			}
			if strings.TrimSpace(text) != strings.TrimSpace(in.LastAssistantMessage) {
				return text
			}
		}
	}
	return text
}

// Native worktrees have no framework bootstrap pointer or copied Runtime.
// Locate only an unambiguous existing Main authority in the same Git repository.
func nativeS6ControlRoot(root string, in policy.Input) string {
	if in.AgentID == "" {
		return root
	}
	if _, err := os.Stat(filepath.Join(root, ".claude/loop-state.json")); err == nil {
		return root
	}
	listing, err := workspace.Git(contextForHook(in), root, "worktree", "list", "--porcelain")
	if err != nil {
		return root
	}
	chosen := ""
	for _, line := range strings.Split(listing, "\n") {
		if !strings.HasPrefix(line, "worktree ") {
			continue
		}
		candidate := strings.TrimPrefix(line, "worktree ")
		if _, err := os.Stat(filepath.Join(candidate, ".claude/loop-state.json")); err != nil {
			continue
		}
		snap, err := runtime.NewStore(filepath.Join(candidate, ".claude/loop-state.json"), filepath.Join(candidate, ".claude/loop-events.jsonl")).WithContext(contextForHook(in)).Snapshot()
		if err != nil || fileview.ValidateAuthority(candidate, snap.State) != nil || len(nativeS6Records(snap.State)) == 0 && !nativeS9Available(candidate, snap.State) && nativeReviewStage(snap.State) == "" {
			continue
		}
		if chosen != "" && chosen != candidate {
			return root
		}
		chosen = candidate
	}
	if chosen != "" {
		return chosen
	}
	return root
}

func nativeS6Hook(root string, in policy.Input, out, diagnostic io.Writer) (bool, int) {
	if _, err := os.Stat(filepath.Join(root, ".claude/loop-state.json")); err != nil {
		return false, 0
	}
	snap, err := runtime.NewStore(filepath.Join(root, ".claude/loop-state.json"), filepath.Join(root, ".claude/loop-events.jsonl")).WithContext(contextForHook(in)).Snapshot()
	if err != nil {
		return false, 0
	}
	if in.Event == "PostToolUse" && (in.ToolName == "Agent" || in.ToolName == "Task" || in.ToolName == "SubagentHandback") && len(nativeS6Records(snap.State)) > 0 {
		return true, 0
	}
	if !nativeS6Worker(snap.State, in) {
		return false, 0
	}
	if in.Event == "SubagentStart" || in.Event == "SubagentStop" {
		if engine, err := policy.Load(filepath.Join(root, "docs/control/hook-policy.json")); err == nil {
			observed := policy.Decision{Decision: "audit", RuleID: "native_s6_lifecycle_observed", Reason: "native S6 lifecycle observed", MatchedRuleIDs: []string{"native_s6_lifecycle_observed"}, Retry: "not_applicable"}
			auditStarted := time.Now()
			defer func() {
				observed.ElapsedMS = time.Since(auditStarted).Milliseconds()
				if err := audit.NewOutbox(filepath.Join(root, ".claude/hook-decisions.jsonl")).Append(engine.Envelope(in, observed, time.Now())); err != nil {
					fmt.Fprintln(diagnostic, "Native lifecycle audit unavailable: "+err.Error())
				}
			}()
		}
	}
	if in.Event == "SubagentStart" {
		body := map[string]any{"hookSpecificOutput": map[string]any{"hookEventName": "SubagentStart", "additionalContext": "Native S6: follow the exact task brief in your native Agent prompt. Read actual inputs, save your professional plan before implementation, continue work and real self-checks without activation, commit and save Result in your own worktree. Return the exact brief and absolute Result reference. Main integrates separately. No framework launch, begin or per-tool receipt is required."}}
		_ = json.NewEncoder(out).Encode(body)
		return true, 0
	}
	if in.Event != "SubagentStop" {
		return true, 0
	} // Ordinary Worker tools do not manage execution.
	records := nativeS6Records(snap.State)
	refs := make([]string, 0, len(records))
	for _, r := range records {
		brief, _, _ := assignmentrecord.NativePaths(r)
		refs = append(refs, filepath.Join(root, brief))
	}
	taskText := nativeReviewTaskText(in, refs)
	var chosen *assignmentrecord.Record
	for _, r := range records {
		brief, _, _ := assignmentrecord.NativePaths(r)
		if nativeS6ExactReference(taskText, filepath.Join(root, brief)) {
			if chosen != nil {
				fmt.Fprintln(diagnostic, "Native S6 return has multiple exact task briefs; Main must resolve the handoff.")
				return true, 0
			}
			r := r
			chosen = &r
		}
	}
	if chosen == nil {
		fmt.Fprintln(diagnostic, "Native S6 return has no exact task brief; preserve the native result for Main, do not guess by role.")
		return true, 0
	}
	// A closed engineering responsibility is not reopened by a late native
	// notification, including one delivered after its worktree was cleaned up.
	switch workstate.AgentActivity(contextAgent(snap.State, chosen.OwnerAgentID)) {
	case "done", "stopped", "cancelled":
		fmt.Fprintln(diagnostic, "Native S6 responsibility already closed; preserve its original receipt/history for Main.")
		return true, 0
	}
	if err := receiveNativeS6(root, in, *chosen, taskText); err != nil {
		var handoff *nativeS6Handoff
		if errors.As(err, &handoff) {
			fmt.Fprintln(diagnostic, "Native S6 blocker returned for Main; no completion receipt created: "+err.Error())
			return true, 0
		}

		return true, nativeStopFailure("Native S6", in.StopHookActive, err, diagnostic)
	}
	return true, 0
}

type nativeS6Handoff struct{ reason string }

func (e *nativeS6Handoff) Error() string { return e.reason }

func receiveNativeS6(root string, in policy.Input, r assignmentrecord.Record, text string) error {
	_, planRel, resultRel := assignmentrecord.NativePaths(r)
	source, err := workspace.Canonical(in.CWD)
	if err != nil {
		return err
	}
	// An explicit absolute output reference can identify the actual native worktree.
	listing, err := workspace.Git(contextForHook(in), root, "worktree", "list", "--porcelain")
	if err != nil {
		return err
	}
	matchedSource := ""
	for _, line := range strings.Split(listing, "\n") {
		if strings.HasPrefix(line, "worktree ") {
			candidate := strings.TrimPrefix(line, "worktree ")
			if nativeS6ExactReference(text, filepath.Join(candidate, resultRel)) {
				if matchedSource != "" && matchedSource != candidate {
					return fmt.Errorf("multiple actual worktree Result references; Main must resolve the handoff")
				}
				matchedSource = candidate
				source = candidate
			}
		}
	}
	if source == root {
		return fmt.Errorf("writing TASK requires its actual native worktree Result, not Main")
	}
	if err := workspace.ValidateCheckout(root, source); err != nil {
		return err
	}
	resultData, err := workpacket.ReadFile(source, resultRel, 1<<20)
	if err != nil {
		return nativeArtifactReadError("read professional Result", err)
	}
	planData, err := workpacket.ReadFile(source, planRel, 1<<20)
	if err != nil {
		var blocked struct {
			Summary string   `json:"summary"`
			Risks   []string `json:"risks_blockers"`
			Closing struct {
				Disposition string `json:"disposition"`
			} `json:"closing_assessment"`
		}
		if json.Unmarshal(resultData, &blocked) == nil && strings.TrimSpace(blocked.Summary) != "" && len(blocked.Risks) > 0 && blocked.Closing.Disposition == "blocked" {
			for _, risk := range blocked.Risks {
				if strings.TrimSpace(risk) != "" {
					return &nativeS6Handoff{reason: blocked.Summary + " / " + risk + "; professional plan unavailable"}
				}
			}
		}
		return nativeArtifactReadError("read professional plan", err)
	}
	projected, err := agentcontext.ApplyRecord(agentcontext.Assignment{}, r)
	if err != nil {
		return err
	}
	plan, _ := agentcontext.PlanTemplate(projected, r.RuntimeID, time.Now().UTC().Format(time.RFC3339Nano), nil)
	result, _ := agentcontext.ResultTemplate(projected, r.RuntimeID, time.Now().UTC().Format(time.RFC3339Nano))
	if err := mergeNativeProfessional(plan, planData, []string{"objective", "planned_paths", "assertion_checks", "dependencies", "risks_blockers"}); err != nil {
		return err
	}
	if err := mergeNativeProfessional(result, resultData, []string{"summary", "assertion_evidence", "risks_blockers", "scope_deviations", "closing_assessment"}); err != nil {
		return err
	}
	plan["assignment_record_contract"] = r.Contract
	result["assignment_record_contract"] = r.Contract
	// Validate the original S6 branches directly so professional mistakes
	// report their exact nested fields instead of an unrelated oneOf branch.
	for _, candidate := range []struct {
		label, definition string
		message           map[string]any
	}{{"plan", "s6PlanReport", plan}, {"Result", "s6CompletionReport", result}} {
		data, err := json.Marshal(candidate.message)
		if err != nil {
			return err
		}
		if err := schema.NewValidator(root).ValidateBytes("agent-message.schema.json#/$defs/"+candidate.definition, data); err != nil {
			var validation *jsonschema.ValidationError
			if errors.As(err, &validation) {
				return &nativeAuthorError{fmt.Errorf("professional %s fields: %w", candidate.label, err)}
			}
			return err
		}
	}
	// The original immutable plan artifact supplies its first receiver timestamp.
	// Repeated/concurrent notifications do not invent a new plan or activation.
	planRef := fmt.Sprintf(".claude/assignments/native/%s/r%d/plan-%s.json", r.AssignmentID, r.Revision, workpacket.Digest(planData))
	if err := publishNativeS6Plan(root, planRef, plan); err != nil {
		return err
	}
	result["occurred_at"] = plan["occurred_at"]
	head, err := workspace.Git(contextForHook(in), source, "rev-parse", "HEAD")
	if err != nil {
		return err
	}
	result["message_id"] = "msg-native-result-" + workpacket.Digest(append(append(planData, resultData...), []byte(head)...))
	resultBytes, _ := json.Marshal(result)
	resultRef := fmt.Sprintf(".claude/assignments/native/%s/r%d/result-%s.json", r.AssignmentID, r.Revision, workpacket.Digest(resultBytes))
	if published, err := workpacket.PublishBounded(root, resultRef, resultBytes, 1<<20); err != nil {
		return err
	} else if !bytes.Equal(published, resultBytes) {
		return fmt.Errorf("immutable native Result source differs")
	}
	coordinatePath := filepath.Join(root, ".claude/assignments", r.AssignmentID+".json")
	coordinateBytes, err := os.ReadFile(coordinatePath)
	if err != nil {
		return err
	}
	var coordinate map[string]any
	if err := json.Unmarshal(coordinateBytes, &coordinate); err != nil {
		return err
	}
	if previous, _ := coordinate["worktree_path"].(string); previous != "" && previous != source {
		return fmt.Errorf("delivery belongs to a different recorded worktree; Main must resolve it")
	}
	if base, _ := coordinate["base_commit"].(string); base == "" {
		return fmt.Errorf("missing original source commit")
	} else if _, err := workspace.Git(contextForHook(in), source, "merge-base", "--is-ancestor", base, "HEAD"); err != nil {
		return fmt.Errorf("native worktree does not contain the declared source commit")
	}
	branch, err := workspace.Git(contextForHook(in), source, "branch", "--show-current")
	if err != nil || branch == "" {
		return fmt.Errorf("native worktree requires a real source branch")
	}
	coordinate["worktree_path"] = source
	coordinate["branch"] = branch
	updated, _ := json.MarshalIndent(coordinate, "", "  ")
	if err := os.WriteFile(coordinatePath, updated, 0600); err != nil {
		return err
	}
	_, _, err = assignment.CompleteTaskWithReceipt(root, filepath.Join(root, ".claude/loop-state.json"), filepath.Join(root, ".claude/loop-events.jsonl"), assignment.CompletionRequest{Context: contextForHook(in), ExpectedRevision: -1, AgentID: r.OwnerAgentID, MessagePath: filepath.Join(root, resultRef), NativePlanPath: planRef, NativeAgentID: in.AgentID, NativeSessionID: in.SessionID})
	return err
}

func publishNativeS6Plan(root, ref string, plan map[string]any) error {
	for attempt := 0; attempt < 2; attempt++ {
		existing, err := workpacket.ReadFile(root, ref, 1<<20)
		if err == nil {
			var received map[string]any
			if err := json.Unmarshal(existing, &received); err != nil {
				return err
			}
			plan["occurred_at"] = received["occurred_at"]
			expected, err := json.Marshal(plan)
			if err != nil {
				return err
			}
			if !bytes.Equal(expected, existing) {
				return fmt.Errorf("immutable native professional plan differs from its original reception")
			}
			return nil
		}
		if !os.IsNotExist(err) {
			return err
		}
		data, err := json.Marshal(plan)
		if err != nil {
			return err
		}
		if published, err := workpacket.PublishBounded(root, ref, data, 1<<20); err == nil {
			if bytes.Equal(published, data) {
				return nil
			}
		} else if attempt == 1 {
			return err
		}
	}
	return fmt.Errorf("native professional plan reception unavailable")
}

func mergeNativeProfessional(target map[string]any, data []byte, allowed []string) error {
	var answer map[string]any
	if err := json.Unmarshal(data, &answer); err != nil {
		return &nativeAuthorError{fmt.Errorf("professional JSON: %w", err)}
	}
	if target == nil {
		return fmt.Errorf("registered TASK has no professional template")
	}
	for key, value := range answer {
		ok := false
		for _, want := range allowed {
			if want == key {
				ok = true
			}
		}
		if !ok {
			return &nativeAuthorError{fmt.Errorf("professional answer cannot set machine field %s", key)}
		}
		target[key] = value
	}
	return nil
}

func nativeS6MainGuidance(root string, state map[string]any) string {
	var lines []string
	for _, r := range nativeS6Records(state) {
		lines = append(lines, nativeS6RecordGuidance(root, state, r))
	}
	return strings.Join(lines, "\n")
}

// This is a Runtime fact/guidance projection, not another completion verifier.
// Main's original task-receipt/integrator still validates checks and provenance.
func nativeS6RecordGuidance(root string, state map[string]any, r assignmentrecord.Record) string {
	brief, _, _ := assignmentrecord.NativePaths(r)
	agent := contextAgent(state, r.OwnerAgentID)
	if workstate.AgentActivity(agent) == "done" {
		return fmt.Sprintf("S6 %s: Main acknowledged the original delivery; continue the reviewed batch.", r.Task.ID)
	}
	if ref := stringValue(agent["completion_reported_ref"]); ref != "" {
		return fmt.Sprintf("S6 %s: professional Result received in the original Runtime record at %s; inspect its original task-receipt for blockers and own checks. Main uses original task-integrate --assignment-id %s; required Main checks and full-batch Gate remain.", r.Task.ID, filepath.Join(root, ref), r.AssignmentID)
	}
	return fmt.Sprintf("S6 %s: Main uses Claude Code native Agent (%s), passing exact brief %s; use native execution/resume, no workspace launcher. Builder owns professional plan, implementation/self-checks and Result; Stop receives delivery.", r.Task.ID, r.RoleFamily, filepath.Join(root, brief))
}

func nativeS6Assignment(state map[string]any, owner string) bool {
	r, found, err := assignmentrecord.ForOwner(state, owner)
	return err == nil && found && r.Contract == assignmentrecord.NativeDeliveryContract
}
