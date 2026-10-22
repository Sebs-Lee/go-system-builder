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
	prepare := request.Event != "Stop" && request.Event != "SubagentStop" && request.Event != "TeammateIdle"
	if !mainQuestionnaireViewer(request) {
		return ""
	}
	lifecycle, _ := snapshot.State["lifecycle"].(map[string]any)
	if lifecycle["state"] != "verification" {
		return ""
	}
	facts, _ := snapshot.State["review"].(map[string]any)
	round := tolerantInt(facts["round"])
	if ptr := review.PlanPointerFromState(snapshot.State); ptr != nil && ptr.ReviewRound == round {
		return ""
	}
	path := reviewPlanQuestionnairePath(root, snapshot.State)
	if path == "" {
		return "S7 Main ReviewPlan questionnaire is unavailable: Main must restore the original current round and generation. No path was guessed."
	}
	if _, err := os.Stat(path); os.IsNotExist(err) && prepare {
		plan, notes := review.DraftPlanForRoot(root, snapshot.State, round)
		if plan == nil {
			return "S7 Main ReviewPlan questionnaire preparation failed at " + path + ": " + strings.Join(notes, "; ")
		}
		data, err := json.MarshalIndent(plan, "", "  ")
		if err == nil {
			err = prepareQuestionnaireText(path, string(data)+"\n")
		}
		if err != nil {
			return "S7 Main ReviewPlan questionnaire preparation failed at " + path + ": " + err.Error()
		}
		if err := prepareQuestionnaireText(reviewPlanGuidePath(path), reviewPlanGuide(notes)); err != nil {
			return "S7 Main questionnaire was prepared at " + path + "; Main must restore its planning guidance: " + err.Error()
		}
	} else if err != nil {
		if os.IsNotExist(err) {
			return "S7 Main ReviewPlan questionnaire is not prepared at " + path + "; the next tool listener must prepare it. Main owns this system gap."
		}
		return "S7 Main ReviewPlan questionnaire is unavailable at " + path + ": " + err.Error()
	}
	guide := reviewPlanGuidePath(path)
	if _, err := os.Stat(guide); os.IsNotExist(err) && request.Event == "SessionStart" {
		_, notes := review.DraftPlanForRoot(root, snapshot.State, round)
		_ = prepareQuestionnaireText(guide, reviewPlanGuide(notes))
	}
	return "S7 Main ReviewPlan questionnaire: " + path + ". Planning sources/candidates and unresolved decisions: " + guide + " (diagnostic only; if missing, Main restores preparation, never rewrites the professional answer). For each Claim, name the actual target, a falsifiable assertion, observable pass/fail oracle and concrete review or execution method using its source refs. Complete assignments and applicable coverage; resolve TODO(planner)/PLANNER-REFINE with real professional facts, then save this exact file for the original ReviewPlan consumer. Keep the supplied baseline; missing source or environment belongs to Main preparation, not invented evidence. Reviewer answers remain independent responsibilities."
}

func runSavedReviewPlanQuestionnaire(root string, request policy.Input, stdout, stderr io.Writer) (bool, int) {
	path, saved := savedQuestionnairePath(request)
	if !saved {
		return false, 0
	}
	snapshot, err := runtime.NewStore(filepath.Join(root, ".claude/loop-state.json"), filepath.Join(root, ".claude/loop-events.jsonl")).Snapshot()
	if err != nil {
		return false, 0
	}
	expected := reviewPlanQuestionnairePath(root, snapshot.State)
	if expected == "" || filepath.Clean(path) != filepath.Clean(expected) {
		return false, 0
	}
	var output, commandErr bytes.Buffer
	next := "Main dispatches native Agents with the registered exact view/answer paths; each Reviewer saves and returns its own answer for SubagentStop reception."
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
		}
	} else {
		received, err := review.RegisterPlan(root, filepath.Join(root, ".claude/loop-state.json"), filepath.Join(root, ".claude/loop-events.jsonl"), review.PlanRequest{ExpectedRevision: -1, PlanPath: path})
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
	b.WriteString("\nChoose the existing E2E state from actual circumstances; unresolved discovery does not require rebuilding scripts. Select exact cases.json#CASE references and actual script paths; the original receiver fills omitted asset IDs/digests. Supply real selector, route and environment facts rather than generated labels. Keep cross-module consumers in the professional scope.\n")
	return b.String()
}
