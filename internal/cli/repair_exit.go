package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/entroforge/go-system-builder/internal/policy"
	"github.com/entroforge/go-system-builder/internal/review"
	"github.com/entroforge/go-system-builder/internal/runtime"
)

func repairExitRecovery(root string) string {
	return fmt.Sprintf("runtime repair handoff complete --root %q", root)
}

// A received Targeted conclusion can outlive a failed successor. Keep that
// failure distinct from author-owned answer omissions when returning to Main.
type repairExitFailure struct {
	cause   error
	message string
}

func (e *repairExitFailure) Error() string { return e.message }
func (e *repairExitFailure) Unwrap() error { return e.cause }

func repairCurrentS7Entry(state map[string]any) bool {
	p := repairPointerCLI(state)
	entry := mapFieldCLI(mapFieldCLI(state, "review"), "round_entry")
	return lifecycleState(state) == "verification" && stringValue(entry["transition_id"]) == "TR-012" &&
		integerValue(entry["round"]) == integerValue(nestedStateValue(state, "review", "round")) &&
		integerValue(entry["baseline_generation"]) == integerValue(nestedStateValue(state, "baseline", "generation")) &&
		stringValue(p["review_plan_seed_ref"]) != "" && entry["review_plan_seed_ref"] == p["review_plan_seed_ref"] &&
		stringValue(p["handoff_ref"]) != "" && entry["repair_handoff_ref"] == p["handoff_ref"]
}

// This projection never retries a Writer or prepares files. The explicit
// completion outlet owns recovery after partial Targeted/Handoff success.
func repairHandoffProjectionContext(root string, state map[string]any) string {
	p := repairPointerCLI(state)
	if stringValue(p["handoff_ref"]) == "" && stringValue(p["status"]) != "ready_for_full_review" {
		return "" // An accepted non-pass or partial Targeted is not a Handoff duty.
	}
	message := ""
	if stringValue(p["handoff_ref"]) == "" {
		message = "S9 original completion recovery: `" + repairExitRecovery(root) + "`. Preserve accepted Targeted observations; this outlet only completes the original Handoff and S7 preparation."
	}
	if ref := stringValue(p["handoff_ref"]); ref != "" {
		message += " Recorded Handoff: " + resolveRootPath(root, ref) + "."
	} else if session := stringValue(p["session_id"]); session != "" {
		message += " Original Handoff source: " + filepath.Join(root, ".claude/review/repair/handoffs", "repair-handoff-"+dispatchSlug(session)+".json") + "."
	}
	if s7BudgetGateRequired(state, root) {
		path, _, err := s7BudgetDecisionApplicationForState(root, state)
		if err == nil {
			message += " Original human budget application: " + path + "."
		}
	}
	if seed := stringValue(p["review_plan_seed_ref"]); repairCurrentS7Entry(state) {
		message += " Fresh S7 seed: " + resolveRootPath(root, seed) + "."
		if plan := review.PlanPointerFromState(state); plan != nil && plan.ReviewRound == integerValue(nestedStateValue(state, "review", "round")) {
			return message + " The current S7 ReviewPlan is already registered; follow its current independent review responsibilities without submitting it again."
		}
		if path := reviewPlanQuestionnairePath(root, state); path != "" {
			message += " S7 Main ReviewPlan questionnaire: " + path + "; planning guide: " + reviewPlanGuidePath(path) + ". Complete and save the original questionnaire for registration."
			for _, source := range []string{path, reviewPlanGuidePath(path)} {
				if info, err := os.Stat(source); err != nil || !info.Mode().IsRegular() {
					message += " Preparation is incomplete at " + source + "; resume `" + repairExitRecovery(root) + "`."
				}
			}
		}
	}
	return message
}

// Only actual completion and explicit recovery call this function; ordinary
// Pre/status projections cannot advance a round or manufacture a decision.
func completeRepairExit(root string) (string, error) {
	ref, snapshot, err := completeRepairHandoff(root)
	if err != nil {
		message := "Original Handoff remains incomplete: " + err.Error() + ". "
		if ref.Path != "" {
			message += "Existing Handoff source: " + resolveRootPath(root, ref.Path) + ". "
		}
		if s7BudgetGateRequired(snapshot.State, root) {
			path, prepareErr := prepareS7BudgetDecisionQuestionnaire(root, snapshot.State)
			if prepareErr != nil {
				message += "Original budget application preparation failed: " + prepareErr.Error() + ". "
			} else {
				message += "Actual review budget is exhausted; original human decision application: " + path + ". Only its real human decision can authorize the successor. "
			}
		}
		return message + repairHandoffProjectionContext(root, snapshot.State), err
	}
	message := "Handoff recorded at " + resolveRootPath(root, ref.Path) + "."
	if !repairCurrentS7Entry(snapshot.State) {
		return message, nil
	}
	// System-owned preparation at the actual completion outlet, even when the
	// triggering event was a Worker stop. This does not receive a plan answer.
	guidance, prepareErr := prepareReviewPlanQuestionnaireContext(root, policy.Input{Event: "PostToolUse", CWD: root}, snapshot)
	if guidance != "" {
		message += " " + guidance
	}
	return message + " " + repairHandoffProjectionContext(root, snapshot.State), prepareErr
}

func runRuntimeRepairHandoffComplete(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("runtime repair handoff complete", flag.ContinueOnError)
	fs.SetOutput(stderr)
	bindUsage(fs, "runtime repair handoff complete")
	root := fs.String("root", ".", "repository root")
	if err := parseWorkspaceFlags(fs, args); err != nil {
		return 2
	}
	message, err := completeRepairExit(*root)
	if err != nil {
		fmt.Fprintln(stderr, message)
		return 1
	}
	snapshot, readErr := runtime.NewStore(filepath.Join(*root, ".claude/loop-state.json"), filepath.Join(*root, ".claude/loop-events.jsonl")).Snapshot()
	if readErr != nil {
		fmt.Fprintln(stderr, readErr)
		return 1
	}
	return encodeJSON(stdout, map[string]any{"revision": snapshot.Revision, "next_action": message})
}
