package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/entroforge/go-system-builder/internal/audit"
	"github.com/entroforge/go-system-builder/internal/integration"
	"github.com/entroforge/go-system-builder/internal/policy"
	"github.com/entroforge/go-system-builder/internal/repair"
	"github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/semantic"
	"github.com/entroforge/go-system-builder/internal/submission"
	"github.com/entroforge/go-system-builder/internal/workpacket"
	"github.com/entroforge/go-system-builder/internal/workspace"
)

// These are professional input/output references, not platform execution slots.
func nativeS9Paths(p workpacket.RepairPlanReportPacket) (brief, source, result string) {
	dir := filepath.ToSlash(filepath.Join(filepath.Dir(p.ViewPath), p.ID, "native"))
	return dir + "/brief.md", dir + "/source.json", ".claude/submissions/native-s9/" + p.ID + "/result.json"
}

func nativeS9PlanPackets(root string, state map[string]any) []workpacket.RepairPlanReportPacket {
	if stringValue(mapFieldCLI(state, "lifecycle")["state"]) != "bug_resolution" || !s9ProfessionalPhase(stringValue(mapFieldCLI(state, "lifecycle")["phase"])) {
		return nil
	}
	packets, err := runtime.RepairPlanReportPackets(state)
	if err != nil {
		return nil
	}
	pointer := repairPointerCLI(state)
	var current []workpacket.RepairPlanReportPacket
	for _, p := range packets {
		b := p.Binding
		if b.RuntimeID != state["runtime_id"] || b.BaselineGeneration != workspace.Generation(state) || b.SessionID != pointer["session_id"] || b.PlanRef != pointer["plan_ref"] || stringMapCLI(pointer["assignment_owners"])[b.RepairAssignmentID] != b.OwnerAgentID {
			continue
		}
		brief, _, _ := nativeS9Paths(p)
		if _, err := os.Stat(filepath.Join(root, brief)); err == nil {
			current = append(current, p)
		}
	}
	return current
}

func prepareNativeS9Brief(root string, p workpacket.RepairPlanReportPacket) (string, error) {
	brief, pin, result := nativeS9Paths(p)
	base := ""
	if data, err := workpacket.ReadFile(root, pin, 4096); err == nil {
		var saved struct {
			Base string `json:"base_commit"`
		}
		if err := json.Unmarshal(data, &saved); err != nil || saved.Base == "" {
			return "", fmt.Errorf("original native S9 source commit is unavailable")
		}
		base = saved.Base
	} else if !os.IsNotExist(err) {
		return "", err
	} else {
		base, err = workspace.Git(context.Background(), root, "rev-parse", "HEAD")
		if err != nil {
			return "", err
		}
		body, _ := json.Marshal(map[string]string{"base_commit": base})
		published, err := workpacket.PublishBounded(root, pin, body, 4096)
		if err != nil {
			return "", err
		}
		if !bytes.Equal(body, published) {
			return "", fmt.Errorf("original native S9 baseline differs; preserve it")
		}
	}
	text := fmt.Sprintf(`# S9 repair — %s

Read the original professional question %s and its actual source materials.
This brief supersedes any legacy workspace/activation instructions in that question;
retain its professional fields, assigned units, assertions and actual evidence duties.
Use Claude Code native worktree isolation from committed Main HEAD containing prepared
baseline %s. Respect the actual dependency order. Before implementation,
perform actual prechecks and personally save the professional PlanReport at %s relative
to your actual native worktree. Write only plan, red_checks and proposed_paths in this
private JSON; the original Main-root question is read-only input, not a Worker write
location. Claude Code forbids isolated Workers from editing the shared Main checkout.
Continue your repair and real self-checks without waiting for a save receipt or activation.
Commit only product deliverables. Keep professional files out of Git.
Save your professional Result at %s relative to your actual native worktree:

`+"```json"+`
{"verdict":"","before_fix_results":[],"unit_results":[],"detection_gap_artifacts":[],"evidence_refs":[],"scope_deviations":[],"residual_risks":[],"migration_ref":"","rollback_ref":""}
`+"```"+`

Use verdict completed, blocked, stale or contract_conflict truthfully. before_fix_results
items have assertion_id, observed_red, evidence_refs. unit_results items have unit_id,
status, evidence_refs, implementation_summary, restored_invariant_refs. Use the exact
assigned units/assertions from the question and actual evidence, never invented PASS.
unit_results.status is pass, fail or blocked; implementation_summary and observed_red
are strings. Every evidence_refs and restored_invariant_refs is an array of strings.
detection_gap_artifacts, top-level evidence_refs, scope_deviations and residual_risks
are arrays of strings, never artifact objects. migration_ref and rollback_ref are strings;
use "" if not applicable. Keep every shown field; use [] for an empty array. Do not add
machine metadata or extra nested fields. Record commands/observations in the professional
text fields and reference actual files as strings.
Return this exact brief %s and the absolute Result path. SubagentStop receives the
professional artifacts; Main owns integration and necessary checks. No workspace
prepare/launch, session reservation, Team, generic readback, begin or activation.
`, p.Binding.RepairAssignmentID, filepath.Join(root, p.ViewPath), base, p.AnswerPath, result, filepath.Join(root, brief))
	published, err := workpacket.PublishBounded(root, brief, []byte(text), 1<<20)
	if err != nil {
		return "", err
	}
	if !bytes.Equal(published, []byte(text)) {
		return "", fmt.Errorf("original native S9 brief differs; preserve it")
	}
	return fmt.Sprintf("Main: use Claude Code native Agent with worktree isolation and carry this exact S9 brief: %s. Read-only professional question: %s; private PlanReport output relative to the actual native worktree: %s. Builder owns its professional plan, repair and Result; SubagentStop receives them. Never ask the isolated Worker to write Main. No workspace prepare/launch or platform ID is required. ", filepath.Join(root, brief), filepath.Join(root, p.ViewPath), p.AnswerPath), nil
}

func nativeS9Available(root string, state map[string]any) bool {
	if len(nativeS9PlanPackets(root, state)) > 0 {
		return true
	}
	pointer := repairPointerCLI(state)
	status := stringValue(pointer["status"])
	return stringValue(mapFieldCLI(state, "lifecycle")["state"]) == "bug_resolution" && s9ProfessionalPhase(stringValue(mapFieldCLI(state, "lifecycle")["phase"])) && (status == "targeted_reverification" || status == "blocked" && pointer["failure_route"] == "blocked" && len(repairArtifactRefs(pointer, "targeted_reverification_artifacts", "", "")) > 0)
}

func nativeS9Hook(root string, in policy.Input, out, diagnostic io.Writer) (bool, int) {
	snap, err := runtime.NewStore(filepath.Join(root, ".claude/loop-state.json"), filepath.Join(root, ".claude/loop-events.jsonl")).WithContext(contextForHook(in)).Snapshot()
	if err != nil || !runtime.PacketEnabled(snap.State) || !nativeS9Available(root, snap.State) {
		return false, 0
	}
	if in.AgentID == "" || in.AgentID == "Main" {
		if in.Event == "PostToolUse" && (in.ToolName == "Agent" || in.ToolName == "Task") {
			return true, 0
		}
		return false, 0
	}
	// Previously installed independent sessions keep their frozen recovery path.
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
		observed := policy.Decision{Decision: "audit", RuleID: "native_s9_lifecycle_observed", Reason: "native S9 lifecycle observed", Retry: "not_applicable"}
		auditStarted := time.Now()
		defer func() {
			observed.ElapsedMS = time.Since(auditStarted).Milliseconds()
			if err := audit.NewOutbox(filepath.Join(root, ".claude/hook-decisions.jsonl")).Append(engine.Envelope(in, observed, time.Now())); err != nil {
				fmt.Fprintln(diagnostic, err)
			}
		}()
	}
	if in.Event == "SubagentStart" {
		context := "Follow the exact S9 professional brief or independent Targeted answer reference in your native prompt. Save professional artifacts and continue without activation or per-tool receipts. Return those exact references. Main owns integration/checks; actual native execution and resume belong to Claude Code. Unrelated native work keeps its own task."
		context += nativeS9TargetedStartContext(root, snap.State, in)
		_ = json.NewEncoder(out).Encode(map[string]any{"hookSpecificOutput": map[string]any{"hookEventName": in.Event, "additionalContext": context}})
		return true, 0
	}
	packets := nativeS9PlanPackets(root, snap.State)
	var refs, targetedRefs []string
	for _, p := range packets {
		brief, _, _ := nativeS9Paths(p)
		refs = append(refs, filepath.Join(root, brief))
	}
	status := stringValue(repairPointerCLI(snap.State)["status"])
	if status == "targeted_reverification" || status == "blocked" {
		if selections, err := repairTargetedSelections(root, snap.State); err == nil {
			for _, selection := range selections {
				if ref, err := repairTargetedQuestionnairePath(root, snap.State, selection.performingAssignmentID); err == nil {
					targetedRefs = append(targetedRefs, ref)
				}
			}
		}
	}
	refs = append(refs, targetedRefs...)
	text := nativeReviewTaskText(in, refs)
	var selected *workpacket.RepairPlanReportPacket
	for _, p := range packets {
		brief, _, _ := nativeS9Paths(p)
		if nativeS6ExactReference(text, filepath.Join(root, brief)) {
			if selected != nil {
				fmt.Fprintln(diagnostic, "S9 return has multiple exact briefs; Main must resolve it")
				return true, 0
			}
			p := p
			selected = &p
		}
	}
	// A verifier may quote an earlier Builder's brief as background material.
	// Multiple exact responsibilities are ambiguous even across S9 branches;
	// never consume the Builder merely because it was checked first.
	if selected != nil {
		for _, ref := range targetedRefs {
			if nativeS6ExactReference(text, ref) {
				fmt.Fprintln(diagnostic, "S9 return has both Builder and Targeted responsibility references; Main must resolve the handoff")
				return true, 0
			}
		}
	}

	if selected != nil {
		err = receiveNativeS9Builder(root, snap.State, in, *selected, text, diagnostic)
	} else {
		err = receiveNativeS9Targeted(contextForHook(in), root, snap, text, diagnostic)
	}
	if err != nil {
		var blocker *nativeS6Handoff
		if errors.As(err, &blocker) {
			fmt.Fprintln(diagnostic, "Native S9 blocker returned for Main; no completion receipt created: "+err.Error())
			return true, 0
		}
		return true, nativeStopFailure("Native S9", in.StopHookActive, err, diagnostic)
	}
	return true, 0
}

func nativeS9TargetedStartContext(root string, state map[string]any, in policy.Input) string {
	status := stringValue(repairPointerCLI(state)["status"])
	if status != "targeted_reverification" && status != "blocked" {
		return ""
	}
	selections, err := repairTargetedSelections(root, state)
	if err != nil {
		return ""
	}
	var paths []string
	for _, selection := range selections {
		if path, err := repairTargetedQuestionnairePath(root, state, selection.performingAssignmentID); err == nil {
			paths = append(paths, path)
		}
	}
	text := nativeReviewTaskText(in, paths)
	for _, path := range paths {
		if nativeS6ExactReference(text, path) {
			return repairStopConditionsContext(root, state)
		}
	}
	return ""
}

func receiveNativeS9Builder(root string, state map[string]any, in policy.Input, p workpacket.RepairPlanReportPacket, text string, diagnostic io.Writer) error {
	if registry, err := workspace.Decode(state); err == nil && registry != nil {
		if source, exists := registry.Executions[p.Binding.ManifestAssignmentID]; exists && source.Purpose == workspace.NativeDeliveryPurpose && source.Status == "complete" {
			fmt.Fprintln(diagnostic, "Native S9 responsibility already integrated and cleaned; preserve the original receipt.")
			return nil
		}
	}
	_, pin, result := nativeS9Paths(p)
	source := in.CWD
	listing, err := workspace.Git(contextForHook(in), root, "worktree", "list", "--porcelain")
	if err != nil {
		return err
	}
	matched := ""
	for _, line := range strings.Split(listing, "\n") {
		if !strings.HasPrefix(line, "worktree ") {
			continue
		}
		candidate := strings.TrimPrefix(line, "worktree ")
		if nativeS6ExactReference(text, filepath.Join(candidate, result)) {
			if matched != "" && matched != candidate {
				return fmt.Errorf("multiple actual Result roots; Main must resolve the handoff")
			}
			matched, source = candidate, candidate
		}
	}
	source, err = workspace.Canonical(source)
	if err != nil {
		return err
	}
	if source == root {
		return fmt.Errorf("return the actual native worktree Result, not Main")
	}
	if err := workspace.ValidateCheckout(root, source); err != nil {
		return err
	}
	professional, err := workpacket.ReadFile(source, result, 20<<20)
	if err != nil {
		return nativeArtifactReadError("read professional Result or return the actual blocker", err)
	}
	var verdict workpacket.RepairResultProfessionalAnswer
	if err := workpacket.ValidateUniqueJSON(professional); err != nil {
		return &nativeAuthorError{err}
	}
	if err := json.Unmarshal(professional, &verdict); err != nil {
		return &nativeAuthorError{err}
	}
	service := &submission.Service{Root: root, StatePath: filepath.Join(root, ".claude/loop-state.json"), JournalPath: filepath.Join(root, ".claude/loop-events.jsonl"), Validator: semantic.RuntimeCandidateValidator{}}
	if err := importNativeS9PlanReport(contextForHook(in), root, source, p, service, diagnostic); err != nil {
		if (verdict.Verdict == "blocked" || verdict.Verdict == "contract_conflict") && strings.TrimSpace(strings.Join(verdict.ResidualRisks, "; ")) != "" {
			return &nativeS6Handoff{reason: "PlanReport unavailable; actual blocker returned to Main without completion: " + strings.Join(verdict.ResidualRisks, "; ")}
		}
		return err
	}
	if receipt, err := service.SaveRepairPlanReportFrom(contextForHook(in), p.Binding.OwnerAgentID, filepath.Join(root, p.AnswerPath), root, nil); err != nil {
		if (verdict.Verdict == "blocked" || verdict.Verdict == "contract_conflict") && len(verdict.ResidualRisks) > 0 && strings.TrimSpace(strings.Join(verdict.ResidualRisks, " ")) != "" {
			return &nativeS6Handoff{reason: "PlanReport unavailable; actual blocker returned to Main without completion: " + strings.Join(verdict.ResidualRisks, "; ")}
		}
		return err
	} else if !receipt.Accepted {
		return fmt.Errorf("original PlanReport was not accepted: %s", receipt.Error)
	}
	var saved struct {
		Base string `json:"base_commit"`
	}
	data, err := workpacket.ReadFile(root, pin, 4096)
	if err != nil {
		return fmt.Errorf("read original native source pin %s: %w", filepath.Join(root, pin), err)
	}
	if err := json.Unmarshal(data, &saved); err != nil || saved.Base == "" {
		return fmt.Errorf("original native source pin missing")
	}
	base := saved.Base
	// A dependency may advance Main between professional preparation and the
	// native worktree's actual creation. Observe its common Git base at first
	// return, then preserve that base for every retry/integration of this source.
	if registry, err := workspace.Decode(state); err != nil {
		return err
	} else if registry != nil && registry.Executions[p.Binding.ManifestAssignmentID].Purpose == workspace.NativeDeliveryPurpose {
		base = registry.Executions[p.Binding.ManifestAssignmentID].BaseCommit
	} else {
		mainHead, err := workspace.Git(contextForHook(in), root, "rev-parse", "HEAD")
		if err != nil {
			return err
		}
		base, err = workspace.Git(contextForHook(in), source, "merge-base", "HEAD", mainHead)
		if err != nil {
			return err
		}
	}
	if _, err := workspace.Git(contextForHook(in), source, "merge-base", "--is-ancestor", saved.Base, base); err != nil {
		return fmt.Errorf("native Git source does not contain the prepared baseline")
	}
	if err := recordNativeS9Source(contextForHook(in), root, source, base, p); err != nil {
		return err
	}
	packets, err := service.IssueRepairResults(contextForHook(in), p.Binding.OwnerAgentID)
	if err != nil {
		return err
	}
	for _, packet := range packets {
		if packet.Binding.Sources.RepairAssignmentID != p.Binding.RepairAssignmentID {
			continue
		}
		// Only author-owned professional fields are transferred into the original
		// questionnaire. No platform identity or result metadata is model-authored.
		var answer workpacket.RepairResultProfessionalAnswer
		decoder := json.NewDecoder(bytes.NewReader(professional))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&answer); err != nil {
			return &nativeAuthorError{err}
		}
		draft := packet.Draft
		// Preserve explicit null/missing values for the original professional
		// validator; decoding into Go strings must not turn unknown into "".
		draft.Answer = append(json.RawMessage(nil), professional...)
		if err := persistRepairLeadQuestionnaire(filepath.Join(source, packet.AnswerPath), draft); err != nil {
			return err
		}
		receipt, err := service.SaveRepairResultFrom(contextForHook(in), p.Binding.OwnerAgentID, filepath.Join(source, packet.AnswerPath), source, nil)
		if err != nil {
			return err
		}
		if !receipt.Accepted {
			return fmt.Errorf("original RepairResult was not accepted: %s", receipt.Error)
		}
		fmt.Fprintln(diagnostic, "Native S9 professional Result received; Main owns original integration and checks. "+receipt.NextAction)
		return nil
	}
	return fmt.Errorf("original Result consumer has no current responsibility for this Builder")
}

// Import only this returned worktree's original professional payload. The
// shared questionnaire is a receiver input, never a Worker write location.
func importNativeS9PlanReport(ctx context.Context, root, source string, p workpacket.RepairPlanReportPacket, service *submission.Service, diagnostic io.Writer) error {
	raw, err := workpacket.ReadFile(source, p.AnswerPath, 20<<20)
	if os.IsNotExist(err) {
		receipt, receiptErr := service.RepairPlanReportReceipt(p.Binding.OwnerAgentID, p.ID)
		if receiptErr == nil && receipt.Accepted {
			return nil
		}
		return &nativeAuthorError{fmt.Errorf("actual native worktree PlanReport is missing at %s; Main's draft cannot substitute for the Worker artifact", filepath.Join(source, p.AnswerPath))}
	}
	if err != nil {
		return err
	}
	if err := workpacket.ValidateUniqueJSON(raw); err != nil {
		return &nativeAuthorError{err}
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil || object == nil {
		return &nativeAuthorError{fmt.Errorf("actual native PlanReport must be a JSON object: %v", err)}
	}
	professional := json.RawMessage(raw)
	if answer, exists := object["answer"]; exists {
		for _, key := range []string{"plan", "red_checks", "proposed_paths"} {
			if _, mixed := object[key]; mixed {
				return &nativeAuthorError{fmt.Errorf("actual native PlanReport mixes answer and professional fields")}
			}
		}
		professional = answer
	}
	draft := p.Draft
	draft.Answer = professional
	data, err := json.Marshal(draft)
	if err != nil {
		return err
	}
	answer, fields := workpacket.ValidateRepairPlanReportAnswer(p, data)
	if len(fields) != 0 {
		return &workpacket.AnswerError{Label: "actual native PlanReport", Fields: fields}
	}
	draft.Answer = answer
	if err := persistRepairLeadQuestionnaire(filepath.Join(root, p.AnswerPath), draft); err != nil {
		return err
	}
	fmt.Fprintf(diagnostic, "Imported original professional PlanReport from actual native worktree: %s\n", filepath.Join(source, p.AnswerPath))
	return nil
}

// This records a returned Git source for the existing Result/integration consumers.
// It never creates a worktree, bootstrap, Claude session, launch lease or process.
func recordNativeS9Source(parent context.Context, root, source, base string, p workpacket.RepairPlanReportPacket) error {
	ctx, unlock, err := integration.LockWorkspace(parent, root)
	if err != nil {
		return err
	}
	defer unlock()
	store := runtime.NewWriter(filepath.Join(root, ".claude/loop-state.json"), filepath.Join(root, ".claude/loop-events.jsonl"), root, semantic.RuntimeCandidateValidator{}).WithContext(ctx)
	snap, err := store.Snapshot()
	if err != nil {
		return err
	}
	registry, err := workspace.Decode(snap.State)
	if err != nil {
		return err
	}
	if registry == nil {
		registry, err = workspace.New(ctx, root)
		if err != nil {
			return err
		}
	}
	if err := registry.Validate(ctx, root); err != nil {
		return err
	}
	b := p.Binding
	if stringMapCLI(repairPointerCLI(snap.State)["assignment_owners"])[b.RepairAssignmentID] != b.OwnerAgentID || snap.State["runtime_id"] != b.RuntimeID || workspace.Generation(snap.State) != b.BaselineGeneration {
		return fmt.Errorf("native S9 responsibility changed before reception")
	}
	if old, exists := registry.Executions[b.ManifestAssignmentID]; exists {
		if old.AgentID == b.OwnerAgentID && old.RuntimeID == b.RuntimeID && old.BaselineGeneration == b.BaselineGeneration && old.Path == source && old.BaseCommit == base && old.BootstrapSHA256 == "" && old.PlatformSessionID == "" {
			return registry.ValidateExecution(ctx, old, false)
		}
		return fmt.Errorf("an existing frozen source differs; Main must preserve and resolve it")
	}
	branch, err := workspace.Git(ctx, source, "branch", "--show-current")
	if err != nil || branch == "" {
		return fmt.Errorf("native source needs its actual Git branch")
	}
	plan, err := repairDispatchPlan(root, snap.State)
	if err != nil {
		return err
	}
	var scope []string
	for _, a := range plan.Assignments {
		if a.AssignmentID == b.RepairAssignmentID {
			scope = a.Scope
		}
	}
	if len(scope) == 0 {
		return fmt.Errorf("current repair scope missing")
	}
	e := workspace.Execution{Purpose: workspace.NativeDeliveryPurpose, AssignmentID: b.ManifestAssignmentID, RuntimeID: b.RuntimeID, BaselineGeneration: b.BaselineGeneration, Generation: 1, AgentID: b.OwnerAgentID, Path: source, Branch: branch, TargetBranch: registry.Branch, BaseCommit: base, Status: "ready", WritePaths: scope, Checks: []string{}, Inputs: map[string]string{}, CheckLocation: "main"}
	if err := registry.ValidateExecution(ctx, e, false); err != nil {
		return err
	}
	registry.Executions[e.AssignmentID] = e
	key := "native-s9-source-" + p.ID
	_, err = store.Update(snap.Revision, runtime.Mutation{EventID: key, TransitionID: "WORKSPACE", Event: "workspace_updated", Actor: "orchestrator", RuntimeID: b.RuntimeID, BaselineGeneration: b.BaselineGeneration, IdempotencyKey: key, RetainLastTransition: true, OccurredAt: time.Now().UTC(), Message: "record returned native S9 Git source", Apply: func(state map[string]any) error { state["workspace"] = workspace.Encode(registry); return nil }})
	return err
}

func receiveNativeS9Targeted(ctx context.Context, root string, snap runtime.Snapshot, text string, diagnostic io.Writer) error {
	status := stringValue(repairPointerCLI(snap.State)["status"])
	if status != "targeted_reverification" && status != "blocked" {
		return nil
	}
	selections, err := repairTargetedSelections(root, snap.State)
	if err != nil {
		return err
	}
	var chosen *repairTargetedSelection
	var path string
	for _, selection := range selections {
		ref, err := repairTargetedQuestionnairePath(root, snap.State, selection.performingAssignmentID)
		if err != nil {
			return err
		}
		if nativeS6ExactReference(text, ref) {
			if chosen != nil {
				return fmt.Errorf("multiple Targeted responsibility references; Main must resolve it")
			}
			selection := selection
			chosen, path = &selection, ref
		}
	}
	if chosen == nil {
		return nil
	}
	// Keep the original independent-verification rules; native IDs are observed
	// in the audit envelope, never substituted for the engineering verifier owner.
	answer, err := readRepairTargetedAnswer(path)
	if err != nil {
		return err
	}
	var feedback bytes.Buffer
	_, code := saveRepairTargetedAnswerContext(ctx, root, snap, path, *chosen, answer, &feedback)
	var message struct {
		Output struct {
			Context string `json:"additionalContext"`
		} `json:"hookSpecificOutput"`
	}
	_ = json.Unmarshal(feedback.Bytes(), &message)
	fmt.Fprintln(diagnostic, message.Output.Context)
	if code != 0 {
		return fmt.Errorf("original Targeted consumer: %s", feedback.String())
	}
	latest, err := runtime.NewStore(filepath.Join(root, ".claude/loop-state.json"), filepath.Join(root, ".claude/loop-events.jsonl")).WithContext(ctx).Snapshot()
	if err != nil {
		return err
	}
	pointer := repairPointerCLI(latest.State)
	for _, ref := range repairArtifactRefs(pointer, "targeted_reverification_artifacts", "", "") {
		var value repair.TargetedReverification
		if err := repairLeadIndexedSource(root, ref, &value); err != nil {
			continue
		}
		if value.ImpactID == chosen.impact.ImpactID && value.OriginalAssignmentID == chosen.originalAssignmentID && value.PerformingAssignmentID == chosen.performingAssignmentID {
			if err := repairLeadTargetedReplay(root, path, ref); err == nil {
				return nil
			}
		}
	}
	return fmt.Errorf("Targeted answer was not recorded: %s", message.Output.Context)
}
