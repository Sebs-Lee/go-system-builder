package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/entroforge/go-system-builder/internal/assignmentrecord"
	"github.com/entroforge/go-system-builder/internal/cli"
	"github.com/entroforge/go-system-builder/internal/submission"
)

func mainNativeReturn(t *testing.T, root, prompt string, response ...string) string {
	t.Helper()
	text := "Finished."
	if len(response) > 0 {
		text = response[0]
	}
	return mainNativeReturnResponse(t, root, prompt, map[string]any{"content": []any{map[string]any{"type": "text", "text": text}}})
}

func mainNativeReturnResponse(t *testing.T, root, prompt string, response map[string]any) string {
	t.Helper()
	body, _ := json.Marshal(map[string]any{"hook_event_name": "PostToolUse", "cwd": root, "tool_name": "Agent", "tool_input": map[string]any{"prompt": prompt}, "tool_response": response})
	before := dryRunTree(t, root)
	var out, diagnostic bytes.Buffer
	if code := cli.Run([]string{"hook", "--root", root, "--event", "PostToolUse"}, bytes.NewReader(body), &out, &diagnostic); code != 0 {
		t.Fatalf("Main return: %d %s %s", code, &out, &diagnostic)
	}
	if !reflect.DeepEqual(before, dryRunTree(t, root)) {
		t.Fatal("Main return observation changed project files")
	}
	if out.Len() == 0 {
		return ""
	}
	var result struct {
		Output struct {
			Event string `json:"hookEventName"`
			Text  string `json:"additionalContext"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal(out.Bytes(), &result); err != nil || result.Output.Event != "PostToolUse" {
		t.Fatalf("Main did not receive Hook context: %v %s", err, &out)
	}
	return result.Output.Text
}

func TestNativeMainReturnReadsS5ReceiptAcrossStageWithoutReceivingAnswers(t *testing.T) {
	f := newStageAnswerFixture(t)
	p := f.packets[0]
	stageJSON(t, f.root, p.AnswerPath, completedDraft(p, "pass"))
	ref := filepath.Join(f.root, p.ViewPath)
	if text := mainNativeReturn(t, f.root, ref); !strings.Contains(text, `"accepted":false`) {
		t.Fatalf("unreceived answer was not distinguished from receipt: %s", text)
	}
	if text := mainNativeReturn(t, f.root, "Unrelated exploratory task"); strings.Contains(text, "Native return") {
		t.Fatalf("unrelated task was claimed: %s", text)
	}
	for _, p := range f.packets {
		stageJSON(t, f.root, p.AnswerPath, completedDraft(p, "pass"))
		if code, _, diagnostic := nativeReviewEvent(t, f.root, "SubagentStop", "opaque-reviewer", filepath.Join(f.root, p.ViewPath)); code != 0 {
			t.Fatal(diagnostic)
		}
		if err := os.Remove(filepath.Join(f.root, p.AnswerPath)); err != nil {
			t.Fatal(err)
		}
		text := mainNativeReturn(t, f.root, filepath.Join(f.root, p.ViewPath))
		if !strings.Contains(text, `"accepted":true`) || !strings.Contains(text, `"work_complete":true`) || !strings.Contains(text, p.ID) {
			t.Fatalf("Main missed the original receipt: %s", text)
		}
	}
	if stageSnapshot(t, f.root).State["lifecycle"].(map[string]any)["state"] != "building" {
		t.Fatal("last S5 reception did not enter S6")
	}
	if text := mainNativeReturn(t, f.root, "Continue the native task", ref); !strings.Contains(text, `"packet_id":"`+p.ID+`"`) || !strings.Contains(text, `"accepted":true`) {
		t.Fatalf("return reference did not identify original receipt: %s", text)
	}
	if text := mainNativeReturn(t, f.root, ref, filepath.Join(f.root, f.packets[1].ViewPath)); !strings.Contains(text, `"packet_id":"`+p.ID+`"`) || strings.Contains(text, `"packet_id":"`+f.packets[1].ID+`"`) {
		t.Fatalf("background quote replaced actual dispatch reference: %s", text)
	}
	if text := mainNativeReturn(t, f.root, ref+"\n"+filepath.Join(f.root, f.packets[1].ViewPath)); !strings.Contains(text, "multiple exact") {
		t.Fatalf("ambiguous return selected one responsibility: %s", text)
	}
}

func TestNativeMainReturnReadsS6ResultWithoutPrivateArtifacts(t *testing.T) {
	root, source, r := nativeDeliveryFixture(t)
	nativeCommit(t, source)
	nativeAnswer(t, source, r)
	brief, plan, result := assignmentrecord.NativePaths(r)
	if code, output := nativeEvent(t, root, source, "SubagentStop", filepath.Join(root, brief)+"\n"+filepath.Join(source, result), false); code != 0 {
		t.Fatal(output)
	}
	for _, path := range []string{plan, result} {
		if err := os.Remove(filepath.Join(source, path)); err != nil {
			t.Fatal(err)
		}
	}
	text := mainNativeReturn(t, root, filepath.Join(root, brief))
	if !strings.Contains(text, "Result received") || !strings.Contains(text, "task-integrate") || !strings.Contains(text, r.Task.ID) {
		t.Fatalf("Main missed the original Result and its own next responsibility: %s", text)
	}
}

func TestNativeMainS7ReturnDoesNotTreatAnotherResponsibilityBodyAsHistorical(t *testing.T) {
	root, p := s7ListenerFixture(t)
	state := stageSnapshot(t, root).State
	entities := state["entities"].(map[string]any)
	rows := entities["work_packets"].([]any)
	var other map[string]any
	packetData, _ := json.Marshal(p)
	if err := json.Unmarshal(packetData, &other); err != nil {
		t.Fatal(err)
	}
	// A second responsibility for the same owner has an invalid body. Only
	// the registered identity/coordinates are needed for current membership.
	other["packet_id"] = "unrelated-same-owner-body"
	other["view_path"] = "unrelated-same-owner/view.md"
	other["answer_path"] = "unrelated-same-owner/answer.json"
	other["binding"].(map[string]any)["assignment_id"] = "unrelated-same-owner-assignment"
	other["inputs"] = "invalid original body"
	entities["work_packets"] = append(rows, other)
	assignments := state["review"].(map[string]any)["assignments"].(map[string]any)
	data, _ := json.Marshal(assignments[p.Binding.AssignmentID])
	var otherAssignment map[string]any
	if err := json.Unmarshal(data, &otherAssignment); err != nil {
		t.Fatal(err)
	}
	assignments["unrelated-same-owner-assignment"] = otherAssignment
	stageJSON(t, root, ".claude/loop-state.json", state)
	text := mainNativeReturn(t, root, filepath.Join(root, p.ViewPath))
	if !strings.Contains(text, `"packet_id":"`+p.ID+`"`) || !strings.Contains(text, `"accepted":false`) || strings.Contains(text, "historical or replaced") || strings.Contains(text, "summary could not be refreshed") {
		t.Fatalf("unrelated body erased current responsibility: %s", text)
	}
}

func TestNativeMainS7ReturnKeepsSubmissionDiagnosticWhenOwnBodyFails(t *testing.T) {
	root, p := s7ListenerFixture(t)
	state := stageSnapshot(t, root).State
	prior := "original submission succeeded at Runtime revision 42; receipt verification failed for " + p.Binding.OwnerAgentID
	submission.RecordPlanningDiagnostic(root, state, "S7 native Stop "+p.ID, prior)
	for _, raw := range state["entities"].(map[string]any)["work_packets"].([]any) {
		row := raw.(map[string]any)
		if row["packet_id"] == p.ID {
			row["inputs"] = "invalid selected original body"
		}
	}
	stageJSON(t, root, ".claude/loop-state.json", state)
	text := mainNativeReturn(t, root, filepath.Join(root, p.ViewPath))
	for _, want := range []string{prior, "current receipt is unknown", p.ID, "runtime review-continue", "runtime audit-prepare"} {
		if !strings.Contains(text, want) {
			t.Fatalf("lost original diagnostic or read failure %q: %s", want, text)
		}
	}
	if strings.Contains(text, `"accepted":false`) || strings.Contains(text, `"accepted":true`) || strings.Contains(text, "historical or replaced") {
		t.Fatalf("selected body failure manufactured reception state: %s", text)
	}
}
