package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/entroforge/go-system-builder/internal/cli"
	"github.com/entroforge/go-system-builder/internal/fileview"
	"github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/transition"
	"github.com/entroforge/go-system-builder/internal/workpacket"
	fixtures "github.com/entroforge/go-system-builder/tests/fixtures/req039"
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

func TestPlanningObserverUsesAnswerAndBoundCommitWithoutRepeatingUnchangedSave(t *testing.T) {
	for _, stage := range []string{"S3", "S4"} {
		t.Run(stage, func(t *testing.T) {
			var root, path, packetID string
			var complete any
			if stage == "S3" {
				var p workpacket.ContractPacket
				root, p = newContractAnswerFixture(t)
				path, packetID, complete = p.AnswerPath, p.ID, completedContractDraft(p)
			} else {
				var p workpacket.TaskPlanningPacket
				root, p = newTaskPlanningAnswerFixture(t)
				path, packetID, complete = p.AnswerPath, p.ID, completedTaskPlanningDraft(p)
			}
			// These fixtures copy the repository's actual default source declarations.
			catalog, err := transition.LoadCatalog(root)
			if err != nil {
				t.Fatal(err)
			}
			view, err := fileview.New(root, "refs/heads/"+nestedDevelopmentBranch(t, root), catalog.Definition.FileSources)
			if err != nil {
				t.Fatal(err)
			}
			if source, err := view.Source("docs/dev"); err != nil || source != "git_tree" {
				t.Fatalf("actual production source: %s %v", source, err)
			}
			hook := func(tool, saved, session string, failed bool) string {
				t.Helper()
				input := map[string]any{"command": "arbitrary script text; never classified as save authority"}
				if saved != "" {
					input = map[string]any{"file_path": filepath.Join(root, saved)}
				}
				body, _ := json.Marshal(map[string]any{"hook_event_name": "PostToolUse", "tool_name": tool, "cwd": root, "session_id": session, "tool_input": input, "tool_response": map[string]any{"is_error": failed}})
				var out, stderr bytes.Buffer
				if code := cli.Run([]string{"hook", "--root", root, "--event", "PostToolUse"}, bytes.NewReader(body), &out, &stderr); code != 0 {
					t.Fatalf("hook %d: %s %s", code, &out, &stderr)
				}
				return out.String()
			}
			receipt := stage + " SUBMISSION RECEIPT"
			save := func(conclusion string) {
				stageJSON(t, root, path, map[string]any{"answer": map[string]any{"conclusion": conclusion}})
			}
			requireReceipt := func(out string) {
				t.Helper()
				if !strings.Contains(out, receipt) {
					t.Fatalf("original consumer was not invoked: %s", out)
				}
			}
			save("pass") // Deliberately incomplete; the original consumer reports missing work.
			requireReceipt(hook("Edit", path, "planning-observer", false))
			for i := 0; i < 9; i++ {
				if out := hook("Bash", "", "planning-observer", false); strings.Contains(out, receipt) {
					t.Fatalf("unchanged Bash %d repeated reception: %s", i, out)
				}
			}
			requireReceipt(hook("Edit", path, "planning-observer", false))
			save("fail")
			// A failed tool or passive read must not swallow the next real save.
			hook("Bash", "", "planning-observer", true)
			hook("Read", path, "planning-observer", false)
			requireReceipt(hook("Bash", "", "planning-observer", false))
			save("blocked")
			requireReceipt(hook("mcp__sandbox__execute", "", "planning-observer", false))
			writeContractSource(t, root, "docs/dev/observer-note.md", "# Real professional edit\n")
			hook("Write", "docs/dev/observer-note.md", "planning-observer", false)
			fixtures.CommitCurrentFixture(t, root)
			requireReceipt(hook("Bash", "", "planning-observer", false))
			if out := hook("Bash", "", "planning-observer", false); strings.Contains(out, receipt) {
				t.Fatalf("unchanged committed output repeated: %s", out)
			}
			requireReceipt(hook("Bash", "", "", false)) // Missing telemetry keeps the original observer.
			// A damaged optional diagnostic slot must not suppress real work.
			cacheDir := filepath.Join(root, ".claude/hook-notices/work-progress")
			entries, err := os.ReadDir(cacheDir)
			if err != nil {
				t.Fatal(err)
			}
			for _, entry := range entries {
				if !strings.HasSuffix(entry.Name(), ".json") {
					continue
				}
				cachePath := filepath.Join(cacheDir, entry.Name())
				previous, err := os.ReadFile(cachePath)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(cachePath, []byte("damaged"), 0600); err != nil {
					t.Fatal(err)
				}
				requireReceipt(hook("Bash", "", "planning-observer", false))
				if err := os.WriteFile(cachePath, previous, 0600); err != nil {
					t.Fatal(err)
				}
			}
			// An actual disk override for professional docs conservatively uses the
			// original observer, without adding a new source qualification rule.
			definitionPath := filepath.Join(root, "docs/control/loop-definition.json")
			original, err := os.ReadFile(definitionPath)
			if err != nil {
				t.Fatal(err)
			}
			var definition map[string]any
			if err := json.Unmarshal(original, &definition); err != nil {
				t.Fatal(err)
			}
			definition["file_sources"] = append(definition["file_sources"].([]any), map[string]any{"path": "docs/dev", "source": "disk"})
			edited, _ := json.Marshal(definition)
			if err := os.WriteFile(definitionPath, edited, 0600); err != nil {
				t.Fatal(err)
			}
			requireReceipt(hook("Bash", "", "planning-observer", false))
			if err := os.WriteFile(definitionPath, original, 0600); err != nil {
				t.Fatal(err)
			}
			stageJSON(t, root, path, complete)
			requireReceipt(hook("mcp__sandbox__execute", "", "planning-observer", false))
			record, accepted, err := runtime.FindSubmission(stageSnapshot(t, root).State, packetID)
			if err != nil || !accepted || record.ControlStatus != "advanced" {
				t.Fatalf("real completed answer did not advance: %+v %v", record, err)
			}

		})
	}
}

func nestedDevelopmentBranch(t *testing.T, root string) string {
	t.Helper()
	ref, err := fileview.DevelopmentRef(stageSnapshot(t, root).State)
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimPrefix(ref, "refs/heads/")
}

func TestPlanningObserverMissingPacketWaitsForCommitAndPreservesReadChange(t *testing.T) {
	root, p := newContractAnswerFixture(t)
	state := stageSnapshot(t, root).State
	// Synthetic fresh entry retains the prior packet and journal as history;
	// it does not delete a recorded issuance and attempt to recreate its event.
	state["lifecycle"].(map[string]any)["phase_revision"] = p.Binding.StageEntry + 1
	stageJSON(t, root, ".claude/loop-state.json", state)
	source := p.Inputs.Designs[0].Path
	original, err := os.ReadFile(filepath.Join(root, source))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, source)); err != nil {
		t.Fatal(err)
	}
	fixtures.CommitCurrentFixture(t, root)
	hook := func(tool string) string {
		t.Helper()
		input := map[string]any{"command": "ordinary script"}
		if tool == "Read" || tool == "Write" {
			input = map[string]any{"file_path": filepath.Join(root, source)}
		}
		body, _ := json.Marshal(map[string]any{"hook_event_name": "PostToolUse", "tool_name": tool, "cwd": root, "session_id": "missing-packet-observer", "tool_input": input})
		var out, stderr bytes.Buffer
		if code := cli.Run([]string{"hook", "--root", root, "--event", "PostToolUse"}, bytes.NewReader(body), &out, &stderr); code != 0 {
			t.Fatalf("hook %d: %s %s", code, &out, &stderr)
		}
		return stderr.String()
	}
	if diagnostic := hook("Bash"); diagnostic == "" {
		t.Fatal("first missing packet observation did not try its original preparation")
	}
	for i := 0; i < 9; i++ {
		if diagnostic := hook("Bash"); diagnostic != "" {
			t.Fatalf("unchanged missing packet Bash %d repeated preparation: %s", i, diagnostic)
		}
	}
	// A normal explicit save still retries a temporarily incomplete original outlet.
	if diagnostic := hook("Write"); diagnostic == "" {
		t.Fatal("explicit save lost its original preparation outlet")
	}
	writeContractSource(t, root, source, string(original))
	fixtures.CommitCurrentFixture(t, root)
	hook("Read") // Must not consume the changed commit observation.
	preparation := hook("Bash")
	if preparation == "" {
		t.Fatal("real commit did not retry the original issuer")
	}
	snapshot := stageSnapshot(t, root)
	var currentPath string
	for _, raw := range snapshot.State["entities"].(map[string]any)["work_packets"].([]any) {
		row := raw.(map[string]any)
		if row["template_id"] == "s3-contract-v1" && row["packet_id"] != p.ID {
			currentPath, _ = row["answer_path"].(string)
		}
	}
	if currentPath == "" {
		t.Fatal("real commit did not prepare a current registered packet:", preparation)
	}
	if _, err := os.Stat(filepath.Join(root, currentPath)); err != nil {
		t.Fatal(err)
	}
	// A deleted registered answer is a stable missing-file fact, then changes
	// naturally when the original issuer restores it.
	if err := os.Remove(filepath.Join(root, currentPath)); err != nil {
		t.Fatal(err)
	}
	hook("Bash")
	if _, err := os.Stat(filepath.Join(root, currentPath)); err != nil {
		t.Fatal("missing answer was not restored:", err)
	}
	hook("Bash") // New draft replaces missing marker; preparation has no work.
	if diagnostic := hook("Bash"); diagnostic != "" {
		t.Fatalf("restored answer repeated preparation: %s", diagnostic)
	}
}
