package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/entroforge/go-system-builder/internal/cli"
	"github.com/entroforge/go-system-builder/internal/runtime"
)

func observedQuestionnaireHook(t *testing.T, root, tool string, input, response map[string]any) string {
	t.Helper()
	body, _ := json.Marshal(map[string]any{
		"hook_event_name": "PostToolUse", "tool_name": tool, "cwd": root,
		"tool_input": input, "tool_response": response,
	})
	var out, errout bytes.Buffer
	if code := cli.Run([]string{"hook", "--root", root, "--event", "PostToolUse"}, bytes.NewReader(body), &out, &errout); code != 0 {
		t.Fatalf("observed save failed: %d %s %s", code, &out, &errout)
	}
	return out.String()
}

func TestQuestionnaireObserverReceivesScriptAndCustomToolSaves(t *testing.T) {
	for _, tool := range []string{"Bash", "mcp__sandbox__execute", "mcp__sandbox__write_file"} {
		t.Run(tool, func(t *testing.T) {
			root, packet := newDesignFixture(t, "none")
			path := filepath.Join(root, packet.AnswerPath)
			stageJSON(t, root, packet.AnswerPath, designDraft(packet))
			input := map[string]any{"command": "script writes the existing questionnaire; its text is not parsed"}
			if tool == "mcp__sandbox__write_file" {
				input = map[string]any{"file_path": path}
			}
			out := observedQuestionnaireHook(t, root, tool, input, map[string]any{})
			if !strings.Contains(out, path) || !strings.Contains(out, "S2 SUBMISSION RECEIPT") {
				t.Fatalf("no exact observed-save receipt: %s", out)
			}
			snapshot := stageSnapshot(t, root)
			record, found, err := runtime.FindSubmission(snapshot.State, packet.ID)
			if err != nil || !found || record.ControlStatus != "advanced" {
				t.Fatalf("original save/transition missing: %+v %v", record, err)
			}
			if snapshot.State["lifecycle"].(map[string]any)["phase"] != "contracts" {
				t.Fatal("original S2 stage did not advance")
			}
			assertDesignCheckpoint(t, root)
		})
	}
}

func TestQuestionnaireObserverSkipsUnchangedDraftAndFailedTool(t *testing.T) {
	root, packet := newDesignFixture(t, "none")
	before := stageSnapshot(t, root).Revision
	out := observedQuestionnaireHook(t, root, "Bash", map[string]any{"command": "true"}, map[string]any{})
	if strings.Contains(out, "SUBMISSION RECEIPT") || stageSnapshot(t, root).Revision != before {
		t.Fatalf("untouched draft invoked its consumer: %s", out)
	}
	stageJSON(t, root, packet.AnswerPath, designDraft(packet))
	out = observedQuestionnaireHook(t, root, "Bash", map[string]any{"command": "false"}, map[string]any{"is_error": true})
	if strings.Contains(out, "SUBMISSION RECEIPT") || stageSnapshot(t, root).Revision != before {
		t.Fatalf("failed tool consumed an older answer: %s", out)
	}
	if _, found, err := runtime.FindSubmission(stageSnapshot(t, root).State, packet.ID); err != nil || found {
		t.Fatalf("failed tool registered a submission: %v %v", found, err)
	}
}

func TestS6ObserverSkipsMachineOnlyResultAndRoutesMainCorrectionToBuilder(t *testing.T) {
	fixture := readyS6CanonicalResultCLI(t)
	beforeState, beforeJournal := s6CanonicalRuntimePair(t, fixture.root)
	observed := observedQuestionnaireHook(t, fixture.root, "Bash", map[string]any{"command": "true"}, map[string]any{})
	if strings.Contains(observed, "S6 Completion questionnaire saved") {
		t.Fatalf("Main's ordinary Bash observation submitted the system-prefilled Result: %s", observed)
	}
	s6AssertCanonicalRuntimePair(t, fixture.root, beforeState, beforeJournal)

	resultBytes, err := os.ReadFile(filepath.Join(fixture.root, filepath.FromSlash(fixture.resultPath)))
	if err != nil {
		t.Fatal(err)
	}
	var draft map[string]any
	if err := json.Unmarshal(resultBytes, &draft); err != nil {
		t.Fatal(err)
	}
	draft["summary"] = "A partial answer that still needs the registered Builder's evidence."
	stageJSON(t, fixture.root, fixture.resultPath, draft)
	observed = observedQuestionnaireHook(t, fixture.root, "Bash", map[string]any{"command": "true"}, map[string]any{})
	if !strings.Contains(observed, fixture.record.OwnerAgentID) || !strings.Contains(observed, fixture.resultPath) ||
		!strings.Contains(observed, "owns correcting this answer") || strings.Contains(observed, "Main should fill") {
		t.Fatalf("Main's observer did not route the incomplete Result correction to its registered Builder: %s", observed)
	}
	s6AssertCanonicalRuntimePair(t, fixture.root, beforeState, beforeJournal)

	evidence := []byte("The registered Builder's actual Result evidence.\n")
	s6WriteRecordContextFile(t, fixture.root, fixture.sourceRef, evidence)
	draft["summary"] = "The registered Builder implemented the reviewed TASK and cited its delivered evidence."
	draft["assertion_evidence"] = []any{map[string]any{
		"assertion": fixture.assertion, "oracle": fixture.oracle, "evidence_refs": []string{fixture.sourceRef},
	}}
	draft["risks_blockers"], draft["scope_deviations"] = []any{}, []any{}
	if raw, ok := draft["closing_assessment"].(map[string]any); ok {
		raw["disposition"] = "complete"
		conditions, _ := raw["conditions"].([]any)
		for _, value := range conditions {
			condition, _ := value.(map[string]any)
			condition["judgment"] = "satisfied"
			condition["rationale"] = "The registered TASK condition is satisfied by the cited delivery evidence."
		}
	}
	stageJSON(t, fixture.root, fixture.resultPath, draft)
	observed = observedQuestionnaireHook(t, fixture.root, "Bash", map[string]any{"command": "true"}, map[string]any{})
	if !strings.Contains(observed, "S6 Completion questionnaire saved") || !strings.Contains(observed, fixture.resultPath) {
		t.Fatalf("professional script save did not reach the original Result consumer: %s", observed)
	}
	snapshot := stageSnapshot(t, fixture.root)
	entities := snapshot.State["entities"].(map[string]any)
	var resultRef string
	for _, raw := range entities["agents"].([]any) {
		row := raw.(map[string]any)
		if row["id"] == fixture.record.OwnerAgentID {
			resultRef, _ = row["completion_reported_ref"].(string)
		}
	}
	if resultRef == "" {
		t.Fatalf("original TASK completion consumer did not record the Builder Result: %s", observed)
	}
	s6AssertCanonicalBuilderTaskReceipt(t, fixture, resultRef)
}

func TestS7ObserverSkipsMachineDraftAndReceivesProfessionalScriptSave(t *testing.T) {
	root := worktreeProjectRoot(t)
	path := filepath.Join(root, ".claude/submissions/loop-REQ-WORKTREE/g1/S7/r1/review-plan.json")
	entryQuestionnaireHook(t, root, "PreToolUse", path)
	before := stageSnapshot(t, root).Revision
	out := observedQuestionnaireHook(t, root, "Bash", map[string]any{"command": "true"}, map[string]any{})
	if strings.Contains(out, "S7 Main ReviewPlan questionnaire saved") || stageSnapshot(t, root).Revision != before {
		t.Fatalf("untouched Main plan invoked its consumer: %s", out)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var plan map[string]any
	if err := json.Unmarshal(data, &plan); err != nil {
		t.Fatal(err)
	}
	claims, _ := plan["claims"].([]any)
	if len(claims) == 0 {
		t.Fatal("fixture has no professional claims")
	}
	claims[0].(map[string]any)["oracle"] = "The locked acceptance obligations map to observed implementation evidence."
	stageJSON(t, root, ".claude/submissions/loop-REQ-WORKTREE/g1/S7/r1/review-plan.json", plan)
	out = observedQuestionnaireHook(t, root, "mcp__sandbox__execute", map[string]any{}, map[string]any{})
	if !strings.Contains(out, "S7 Main ReviewPlan questionnaire saved") || !strings.Contains(out, path) || !strings.Contains(out, "original validation feedback") {
		t.Fatalf("partial professional plan did not reach original consumer: %s", out)
	}
	if stageSnapshot(t, root).Revision != before {
		t.Fatal("incomplete professional plan changed Runtime")
	}
	authored := minimalReviewPlan(t, t.TempDir(), "review-plan-observed-save")
	data, err = os.ReadFile(authored)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	out = observedQuestionnaireHook(t, root, "Bash", map[string]any{"command": "script saved the plan"}, map[string]any{})
	if !strings.Contains(out, "review-plan-observed-save") || !strings.Contains(out, path) {
		t.Fatalf("complete script save was not registered: %s", out)
	}
	accepted := stageSnapshot(t, root).Revision
	if replay := entryQuestionnaireHook(t, root, "PostToolUse", path); !strings.Contains(replay, "already received") || stageSnapshot(t, root).Revision != accepted {
		t.Fatalf("registered plan replay was not idempotent: %s", replay)
	}
}
