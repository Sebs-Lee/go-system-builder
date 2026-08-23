// s7_status.go provides a read-only board view of the current S7 review
// round: the registered ReviewPlan, every Claim's disposition, assignment
// consumption, current-round Findings, and the sealed batch / clean round
// state. It is the agent's answer to "where do I stand" without parsing
// raw loop-state JSON (L3-S7 §12.2 D7).
package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/entroforge/go-system-builder/internal/metrics"
	"github.com/entroforge/go-system-builder/internal/review"
	"github.com/entroforge/go-system-builder/internal/runtime"
)

// runS7Command is the `loop-harness s7` dispatcher: `status` is the
// read-only board, `draft` scaffolds a ReviewPlan, `manifest-draft`
// scaffolds the reviewer team-manifest for one plan Assignment, and
// `workspace-digest` prints the current verification-artifact digest an
// E2E cold-start ReviewResult must bind (L3-S7 §3.5).
func runS7Command(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || (args[0] != "status" && args[0] != "draft" && args[0] != "manifest-draft" && args[0] != "workspace-digest") {
		fmt.Fprintln(stderr, "s7 requires <status|draft|manifest-draft|workspace-digest>")
		return 2
	}
	if args[0] == "workspace-digest" {
		flags := flag.NewFlagSet("s7 workspace-digest", flag.ContinueOnError)
		flags.SetOutput(stderr)
		root := flags.String("root", ".", "repository root")
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		return runS7WorkspaceDigest(*root, stdout)
	}
	if args[0] == "manifest-draft" {
		flags := flag.NewFlagSet("s7 manifest-draft", flag.ContinueOnError)
		flags.SetOutput(stderr)
		root := flags.String("root", ".", "repository root")
		assignmentID := flags.String("assignment", "", "ReviewPlan Assignment id to dispatch (required)")
		out := flags.String("out", "", "write the draft manifest JSON here (default stdout)")
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		if *assignmentID == "" {
			fmt.Fprintln(stderr, "s7 manifest-draft requires --assignment <assignment-id>")
			return 2
		}
		return runS7ManifestDraft(*root, *assignmentID, *out, stdout)
	}
	if args[0] == "draft" {
		flags := flag.NewFlagSet("s7 draft", flag.ContinueOnError)
		flags.SetOutput(stderr)
		root := flags.String("root", ".", "repository root")
		out := flags.String("out", "", "write the draft plan JSON here (default stdout)")
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		return runS7Draft(*root, *out, stdout)
	}
	flags := flag.NewFlagSet("s7 status", flag.ContinueOnError)
	flags.SetOutput(stderr)
	root := flags.String("root", ".", "repository root")
	if err := flags.Parse(args[1:]); err != nil {
		return 2
	}
	return runS7Status(*root, stdout)
}

// runS7Status reads the current runtime state and prints the S7 review
// board to stdout. It performs no writes.
func runS7Status(root string, stdout io.Writer) int {
	statePath := filepath.Join(root, ".claude", "loop-state.json")
	journalPath := filepath.Join(root, ".claude", "loop-events.jsonl")
	store := runtime.NewStore(statePath, journalPath)
	snapshot, err := store.Snapshot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "read runtime: %v\n", err)
		return 1
	}
	state := snapshot.State

	reviewMap, _ := state["review"].(map[string]any)
	round := 0
	if value, ok := reviewMap["round"].(float64); ok {
		round = int(value)
	}
	ptr := review.PlanPointerFromState(state)
	fmt.Fprintf(stdout, "S7 review board (round %d)\n", round)
	if ptr == nil {
		fmt.Fprintln(stdout, "(no ReviewPlan registered — create one and run `runtime review-plan --file <plan.json>`)")
		return 0
	}
	fmt.Fprintf(stdout, "plan: %s status=%s revision=%d e2e_coverage=%s\n",
		ptr.PlanID, ptr.Status, ptr.Revision, ptr.E2ECoverageState)

	plan, _, planErr := review.LoadPlan(root, state)
	dispositions := review.Dispositions(state)

	// Claims grouped by lens, in plan order when loadable.
	fmt.Fprintln(stdout, "\nclaims:")
	claimOrder := make([]string, 0, len(dispositions))
	if planErr == nil {
		for _, claim := range plan.Claims {
			claimOrder = append(claimOrder, claim.ClaimID)
		}
	} else {
		for claimID := range dispositions {
			claimOrder = append(claimOrder, claimID)
		}
		sort.Strings(claimOrder)
	}
	for _, claimID := range claimOrder {
		disp, ok := dispositions[claimID]
		if !ok {
			continue
		}
		line := fmt.Sprintf("  %s [%s] %s", claimID, disp.Lens, disp.Disposition)
		if disp.Applicability == "not_applicable" {
			line += " (plan-level N/A)"
		}
		if disp.AssignmentID != "" {
			line += fmt.Sprintf(" <- %s", disp.AssignmentID)
		}
		if len(disp.FindingIDs) > 0 {
			line += fmt.Sprintf(" findings=%v", disp.FindingIDs)
		}
		fmt.Fprintln(stdout, line)
	}

	// Assignments with agent binding and consumption state.
	reviewAssignments, _ := reviewMap["assignments"].(map[string]any)
	fmt.Fprintln(stdout, "\nassignments:")
	assignmentIDs := make([]string, 0, len(reviewAssignments))
	for id := range reviewAssignments {
		assignmentIDs = append(assignmentIDs, id)
	}
	sort.Strings(assignmentIDs)
	for _, id := range assignmentIDs {
		row, _ := reviewAssignments[id].(map[string]any)
		if row == nil {
			continue
		}
		agent, _ := row["agent_id"].(string)
		if agent == "" {
			agent = "(not dispatched — run `runtime register-workgroup`)"
		}
		fmt.Fprintf(stdout, "  %s [%s] status=%s agent=%s\n", id, row["lens"], row["status"], agent)
	}

	// Current-round findings.
	findings := review.RoundFindings(state)
	if len(findings) > 0 {
		fmt.Fprintln(stdout, "\nfindings:")
		for _, row := range findings {
			fmt.Fprintf(stdout, "  %s [%s/%s] claim=%s finder=%s\n",
				row["finding_id"], row["lens"], row["severity"], row["claim_id"], row["original_finder"])
		}
	}

	// S7 operational metrics summary (L3-S7 §14.2 machine-collectible
	// subset). Read-only: the runtime verbs record these series.
	if summary, err := metrics.FormatS7(root, round); err == nil {
		fmt.Fprintln(stdout, "")
		fmt.Fprintln(stdout, summary)
	}

	// Exit state and the single next action.
	fmt.Fprintln(stdout, "")
	pending := review.UndispositionedRequired(state)
	if batch, _ := reviewMap["observation_batch"].(map[string]any); batch != nil {
		fmt.Fprintf(stdout, "observation_batch: sealed as %s (%d findings) — the next PreToolUse auto-commits TR-008 (do not invoke the transition CLI)\n",
			batch["batch_id"], len(batchFindingIDs(batch)))
		return 0
	}
	if ptr.Status == "clean" {
		fmt.Fprintln(stdout, "clean round: machine CleanRound registered — the next PreToolUse auto-commits TR-009 (do not invoke the transition CLI)")
		return 0
	}
	if len(pending) > 0 {
		fmt.Fprintf(stdout, "pending required claims: %d — next: dispatch/consume results via `runtime review-result submit`\n", len(pending))
	} else {
		fmt.Fprintln(stdout, "all required claims dispositioned; round consumer closes on the next submit")
	}
	return 0
}

func batchFindingIDs(batch map[string]any) []any {
	ids, _ := batch["finding_ids"].([]any)
	return ids
}

// readLifecycleCursor reports the active lifecycle state/phase for
// disclosure messages; both default to "unknown" so an absent lifecycle
// block never silently produces a stage-less error.
func readLifecycleCursor(state map[string]any) (string, string) {
	stage, phase := "unknown", "unknown"
	lc, _ := state["lifecycle"].(map[string]any)
	if value, _ := lc["state"].(string); value != "" {
		stage = value
	}
	if value, _ := lc["phase"].(string); value != "" {
		phase = value
	}
	return stage, phase
}

// runS7Draft scaffolds a ReviewPlan from the current runtime facts
// (L3-S7 §4.2 planner assist). Read-only: it never mutates state; the
// planner reviews the TODO markers before registering.
//
// Gating rationale (L3-S7 §11.1, blue/L3-S7-verification-round.md): the
// review round is opened by the S6→S7 transition, not by handcrafting
// `review.round`. A draft emitted outside the verification stage would
// be register-rejected (state != verification) and a Planner encouraged
// to fix the stage instead of the plan. We surface this disclosure with
// the current stage, the legal entry path (TR-006/TR-012/TR-016), and one
// next action so the agent does not invent a parallel lifecycle.
func runS7Draft(root, out string, stdout io.Writer) int {
	statePath := filepath.Join(root, ".claude", "loop-state.json")
	journalPath := filepath.Join(root, ".claude", "loop-events.jsonl")
	snapshot, err := runtime.NewStore(statePath, journalPath).Snapshot()
	if err != nil {
		fmt.Fprintf(stdout, "read runtime: %v\n", err)
		return 1
	}
	round := 0
	if reviewMap, ok := snapshot.State["review"].(map[string]any); ok {
		if value, ok := reviewMap["round"].(float64); ok {
			round = int(value)
		}
	}
	if round < 1 {
		stage, phase := readLifecycleCursor(snapshot.State)
		fmt.Fprintf(stdout,
			"current stage is %s (phase=%s); S7 enters automatically when S6 commits a TASK batch via TR-006 (bug_resolution re-entry: TR-012; acceptance re-entry: TR-016) — complete the current stage's missing work (see `loop-harness next` / `loop-harness ready`) instead of handcrafting a ReviewPlan\n",
			stage, phase)
		return 1
	}
	if stage, _ := readLifecycleCursor(snapshot.State); stage != "verification" {
		fmt.Fprintf(stdout,
			"current stage is %s; S7 (verification) ReviewPlan drafting is only legal while the lifecycle stage is `verification` — round=%d is open but the ReviewPlan register-verb (`runtime review-plan --file <plan.json>`) will reject with `a ReviewPlan can only be registered in the verification stage`. Finish the current stage's missing work first.\n",
			stage, round)
		return 1
	}
	plan, notes := review.DraftPlan(snapshot.State, round)
	data, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		fmt.Fprintf(stdout, "encode draft: %v\n", err)
		return 1
	}
	if out != "" {
		if err := os.WriteFile(out, append(data, '\n'), 0o644); err != nil {
			fmt.Fprintf(stdout, "write draft: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "draft ReviewPlan written to %s\n", out)
	} else {
		fmt.Fprintln(stdout, string(data))
	}
	for _, note := range notes {
		fmt.Fprintf(stdout, "note: %s\n", note)
	}
	return 0
}

// runS7WorkspaceDigest prints the current verification-artifact digest of
// the registered plan's cold-start workspace. An E2E ReviewResult must bind
// exactly this value as verification_artifact_digest (L3-S7 §3.5); computing
// it by hand is error-prone, so the submit-time error message points here.
func runS7WorkspaceDigest(root string, stdout io.Writer) int {
	statePath := filepath.Join(root, ".claude", "loop-state.json")
	journalPath := filepath.Join(root, ".claude", "loop-events.jsonl")
	snapshot, err := runtime.NewStore(statePath, journalPath).Snapshot()
	if err != nil {
		fmt.Fprintf(stdout, "read runtime: %v\n", err)
		return 1
	}
	ptr := review.PlanPointerFromState(snapshot.State)
	if ptr == nil || ptr.VerificationArtifactWorkspace == "" {
		fmt.Fprintln(stdout, "no verification artifact workspace is pinned (the plan is not registered or e2e_coverage_state is not cold_start); E2E results do not bind a workspace digest")
		return 0
	}
	digest, err := review.WorkspaceDigest(root, ptr.VerificationArtifactWorkspace)
	if err != nil {
		fmt.Fprintf(stdout, "compute workspace digest: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "%s  %s\n", digest, ptr.VerificationArtifactWorkspace)
	return 0
}
