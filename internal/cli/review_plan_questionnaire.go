package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
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
	} else if err != nil {
		if os.IsNotExist(err) {
			return "S7 Main ReviewPlan questionnaire is not prepared at " + path + "; the next tool listener must prepare it. Main owns this system gap."
		}
		return "S7 Main ReviewPlan questionnaire is unavailable at " + path + ": " + err.Error()
	}
	return "S7 Main ReviewPlan questionnaire: " + path + ". Complete the original Claim assertions/oracles/methods, assignments and required coverage; resolve every TODO(planner), then save this exact file. The original ReviewPlan consumer validates and registers it. Reviewer answers remain independent responsibilities."
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
	next := "Main dispatches the registered review assignments; each Reviewer fills and saves its own exact questionnaire."
	code := 0
	roundFacts, _ := snapshot.State["review"].(map[string]any)
	if ptr := review.PlanPointerFromState(snapshot.State); ptr != nil && ptr.ReviewRound == tolerantInt(roundFacts["round"]) {
		_, _, loadErr := review.LoadPlan(root, snapshot.State)
		authored, readErr := os.ReadFile(path)
		registered, registeredErr := os.ReadFile(filepath.Join(root, filepath.FromSlash(ptr.Path)))
		var a, b any
		if loadErr != nil || readErr != nil || registeredErr != nil || json.Unmarshal(authored, &a) != nil || json.Unmarshal(registered, &b) != nil || !reflect.DeepEqual(a, b) {
			code = 1
			fmt.Fprintln(&output, "The current ReviewPlan is already registered; changed answers require the original named revision/rework path. Preserve the accepted plan.")
		} else {
			fmt.Fprintln(&output, "The same ReviewPlan was already received; no Runtime write or duplicate registration was performed.")
		}
	} else {
		code = runRuntime([]string{"review-plan", "--root", root, "--file", path}, &output, &commandErr)
	}
	if code != 0 {
		next = "Correct the specific original validation feedback and save this same questionnaire; do not hand-edit Runtime or issue another report."
	}
	context := fmt.Sprintf("S7 Main ReviewPlan questionnaire saved at %s. Next: %s\n%s", path, next, output.String()+commandErr.String())
	_ = json.NewEncoder(stdout).Encode(map[string]any{"hookSpecificOutput": map[string]any{"hookEventName": "PostToolUse", "additionalContext": context}})
	return true, 0
}
