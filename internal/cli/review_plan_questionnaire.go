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

	"github.com/entroforge/go-system-builder/internal/policy"
	"github.com/entroforge/go-system-builder/internal/review"
	"github.com/entroforge/go-system-builder/internal/runtime"
)

func reviewPlanQuestionnairePath(root string, state map[string]any) string {
	return review.PlanQuestionnairePath(root, state)
}

func stageReviewPlanQuestionnaireContext(root string, request policy.Input, snapshot runtime.Snapshot) string {
	message, _ := prepareReviewPlanQuestionnaireContext(root, request, snapshot)
	return message
}

func reviewPlanPreparationFailure(message string) (string, error) {
	return message, errors.New(message)
}

func prepareReviewPlanQuestionnaireContext(root string, request policy.Input, snapshot runtime.Snapshot) (string, error) {
	prepare := request.Event != "Stop" && request.Event != "SubagentStop" && request.Event != "TeammateIdle"
	if !mainQuestionnaireViewer(request) {
		return "", nil
	}
	lifecycle, _ := snapshot.State["lifecycle"].(map[string]any)
	if lifecycle["state"] != "verification" {
		return "", nil
	}
	facts, _ := snapshot.State["review"].(map[string]any)
	round := tolerantInt(facts["round"])
	if ptr := review.PlanPointerFromState(snapshot.State); ptr != nil && ptr.ReviewRound == round {
		return "", nil
	}
	path := reviewPlanQuestionnairePath(root, snapshot.State)
	if path == "" {
		return reviewPlanPreparationFailure("S7 Main ReviewPlan questionnaire is unavailable: Main must restore the original current round and generation. No path was guessed.")
	}
	if info, err := os.Stat(path); os.IsNotExist(err) && prepare {
		plan, notes := review.DraftPlanForRoot(root, snapshot.State, round)
		if plan == nil {
			return reviewPlanPreparationFailure("S7 Main ReviewPlan questionnaire preparation failed at " + path + ": " + strings.Join(notes, "; ") + ". " + reviewPlanPreparationRecovery(root, path))
		}
		data, err := json.MarshalIndent(plan, "", "  ")
		if err == nil {
			err = prepareQuestionnaireText(path, string(data)+"\n")
		}
		if err != nil {
			return reviewPlanPreparationFailure("S7 Main ReviewPlan questionnaire preparation failed at " + path + ": " + err.Error() + ". " + reviewPlanPreparationRecovery(root, path))
		}
		if err := prepareQuestionnaireText(reviewPlanGuidePath(path), reviewPlanGuide(notes)); err != nil {
			return reviewPlanPreparationFailure("S7 Main questionnaire was prepared at " + path + "; Main must restore its planning guidance: " + err.Error() + ". " + reviewPlanPreparationRecovery(root, path))
		}
	} else if err != nil {
		if os.IsNotExist(err) {
			return reviewPlanPreparationFailure("S7 Main ReviewPlan questionnaire is not prepared at " + path + "; Main owns this preparation gap. " + reviewPlanPreparationRecovery(root, path))
		}
		return reviewPlanPreparationFailure("S7 Main ReviewPlan questionnaire is unavailable at " + path + ": " + err.Error() + ". " + reviewPlanPreparationRecovery(root, path))
	} else if !info.Mode().IsRegular() {
		return reviewPlanPreparationFailure("S7 Main ReviewPlan questionnaire is unavailable at " + path + ": expected a regular file. " + reviewPlanPreparationRecovery(root, path))
	}
	guide := reviewPlanGuidePath(path)
	if info, err := os.Stat(guide); os.IsNotExist(err) && prepare {
		plan, notes := review.DraftPlanForRoot(root, snapshot.State, round)
		if plan == nil {
			return reviewPlanPreparationFailure("S7 Main guidance preparation failed at " + guide + ": " + strings.Join(notes, "; ") + ". " + reviewPlanPreparationRecovery(root, path))
		}
		if err := prepareQuestionnaireText(guide, reviewPlanGuide(notes)); err != nil {
			return reviewPlanPreparationFailure("S7 Main questionnaire is retained at " + path + "; guidance preparation failed at " + guide + ": " + err.Error() + ". " + reviewPlanPreparationRecovery(root, path))
		}
	} else if os.IsNotExist(err) {
		return reviewPlanPreparationFailure("S7 Main planning guidance is not prepared at " + guide + ". " + reviewPlanPreparationRecovery(root, path))
	} else if err != nil {
		return reviewPlanPreparationFailure("S7 Main planning guidance is unavailable at " + guide + ": " + err.Error() + ". " + reviewPlanPreparationRecovery(root, path))
	} else if err == nil && !info.Mode().IsRegular() {
		return reviewPlanPreparationFailure("S7 Main planning guidance is unavailable at " + guide + ": expected a regular file. " + reviewPlanPreparationRecovery(root, path))
	}
	return "S7 Main ReviewPlan questionnaire: " + path + ". Planning sources/candidates and unresolved decisions: " + guide + " (diagnostic only; if missing, Main restores preparation, never rewrites the professional answer). For each Claim, name the actual target, a falsifiable assertion, observable pass/fail oracle and concrete review or execution method using its source refs. Complete assignments and applicable coverage; resolve TODO(planner)/PLANNER-REFINE with real professional facts, then save this exact file for the original ReviewPlan consumer. Keep the supplied baseline; missing source or environment belongs to Main preparation, not invented evidence. Reviewer answers remain independent responsibilities.", nil
}

func runSavedReviewPlanQuestionnaire(root string, request policy.Input, stdout, stderr io.Writer) (bool, int) {
	path, saved := savedQuestionnairePath(request)
	if !saved {
		return false, 0
	}
	snapshot, err := runtime.NewStore(filepath.Join(root, ".claude/loop-state.json"), filepath.Join(root, ".claude/loop-events.jsonl")).WithContext(contextForHook(request)).Snapshot()
	if err != nil {
		return false, 0
	}
	expected := reviewPlanQuestionnairePath(root, snapshot.State)
	if expected == "" || filepath.Clean(path) != filepath.Clean(expected) {
		return false, 0
	}
	var output, commandErr bytes.Buffer
	next := "Main prepares and registers eligible Reviewer assignments before native dispatch; each Reviewer saves and returns its own answer for SubagentStop reception."
	roundFacts, _ := snapshot.State["review"].(map[string]any)
	if ptr := review.PlanPointerFromState(snapshot.State); ptr != nil && ptr.ReviewRound == tolerantInt(roundFacts["round"]) {
		_, _, loadErr := review.LoadPlan(root, snapshot.State)
		authored, readErr := os.ReadFile(path)
		registered, registeredErr := os.ReadFile(filepath.Join(root, filepath.FromSlash(ptr.Path)))
		if loadErr != nil || readErr != nil || registeredErr != nil {
			next = "Main restores the original accepted plan/source before checking replay; preserve the answer and original receipt."
			fmt.Fprintf(&output, "ReviewPlan replay source unavailable: %v\n", errors.Join(loadErr, readErr, registeredErr))
		} else if !review.SamePlanAnswer(authored, registered) {
			next = "Use the original named revision/rework path for changed professional content; preserve the accepted plan."
			fmt.Fprintln(&output, "The current ReviewPlan is already registered; changed answers require the original named revision/rework path. Preserve the accepted plan.")
		} else {
			fmt.Fprintln(&output, "The same ReviewPlan was already received; no Runtime write or duplicate registration was performed.")
			next += "\n" + s7PendingReviewGuidance(root, snapshot.State)
		}
	} else {
		received, err := review.RegisterPlan(root, filepath.Join(root, ".claude/loop-state.json"), filepath.Join(root, ".claude/loop-events.jsonl"), review.PlanRequest{Context: contextForHook(request), ExpectedRevision: -1, PlanPath: path})
		if err != nil {
			var author *review.PlanAnswerError
			if errors.As(err, &author) {
				next = "Correct the specific original validation feedback in this same questionnaire and save; no new command or report is required."
			} else {
				next = "Main owns source/environment/Writer recovery; preserve this answer and any original receipt, inspect the concrete failure before resuming. Do not rewrite professional content to repair a system failure."
			}
			fmt.Fprintln(&commandErr, formatFailure("runtime review-plan", err))
		} else {
			ptr := review.PlanPointerFromState(received.State)
			fmt.Fprintf(&output, "review-plan: registered %s for round %d (status %s).\n", ptr.PlanID, ptr.ReviewRound, ptr.Status)
			next += "\n" + s7PendingReviewGuidance(root, received.State)
		}
	}
	context := fmt.Sprintf("S7 Main ReviewPlan questionnaire saved at %s. Next: %s\n%s", path, next, output.String()+commandErr.String())
	_ = json.NewEncoder(stdout).Encode(map[string]any{"hookSpecificOutput": map[string]any{"hookEventName": "PostToolUse", "additionalContext": context}})
	return true, 0
}

func reviewPlanGuidePath(path string) string { return strings.TrimSuffix(path, ".json") + ".md" }
func reviewPlanGuide(notes []string) string {
	var b strings.Builder
	b.WriteString("# S7 planning sources and decisions\n\nThis is derived guidance, not an accepted Plan or an execution permit. Complete the original JSON and save once. A candidate is not a reusable-asset declaration or a passing result.\n\n")
	for _, note := range notes {
		b.WriteString("- " + note + "\n")
	}
	b.WriteString("\nChoose the existing E2E state from actual circumstances; unresolved discovery does not require rebuilding scripts. Select exact cases.json#CASE references and actual script paths; the original receiver fills omitted asset IDs/digests. Supply real selector, route and environment facts rather than generated labels. For cold_start, verification_artifact_workspace must be repository-relative under e2e-workspace/; the original registration prepares that isolated script/fixture surface. Reuse confirmed scripts at their original paths and create only missing assets in that workspace.\n\nREQ prototype links bound automatic discovery hints, not the professional regression boundary. Follow the actual changed providers and consumers through approved architecture/contracts, shared models, API/event/auth/configuration and journey/data paths. When removing a dependency or changing its owner, inspect the previously known consumers as well. Use existing declared sources and identify missing relationship/environment material for Main; no mandatory module graph or whole-project fallback is implied. Explain scope from actual dependencies and keep each scenario's exact source.\n")
	return b.String()
}

func reviewPlanPreparationRecovery(root, path string) string {
	return "Main repairs the stated source/filesystem failure, then resumes original S7 preparation with `loop-harness s7 draft --root " + s7CaptureShellQuote(root) + " --out " + s7CaptureShellQuote(path) + "`; preserve any existing professional answer. Existing integration and transition facts remain recorded; do not resubmit Builder delivery or merge again."
}
