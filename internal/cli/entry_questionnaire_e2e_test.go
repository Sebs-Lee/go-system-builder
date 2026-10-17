package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/entroforge/go-system-builder/internal/cli"
)

func entryQuestionnaireHook(t *testing.T, root, event, path string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]any{"hook_event_name": event, "cwd": root, "tool_name": "Write", "tool_input": map[string]any{"file_path": path}})
	var out, errout bytes.Buffer
	if code := cli.Run([]string{"hook", "--root", root, "--event", event}, bytes.NewReader(body), &out, &errout); code != 0 {
		t.Fatalf("%s hook failed %d: %s %s", event, code, &out, &errout)
	}
	return out.String()
}

func initEntryQuestionnaireRoot(t *testing.T, reqs map[string]string) string {
	t.Helper()
	root := newUXTestRoot(t, reqs)
	template, err := os.ReadFile(filepath.Join("..", "..", "docs/requirements/REQ-template.md"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "docs/requirements/REQ-template.md"), template, 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errout bytes.Buffer
	if code := cli.Run([]string{"init", "--root", root}, nil, &out, &errout); code != 0 {
		t.Fatalf("init failed %d: %s", code, &errout)
	}
	return root
}

func TestEntryQuestionnaireS0PreparesActualDraftWithoutHumanDecision(t *testing.T) {
	root := initEntryQuestionnaireRoot(t, nil)
	path := filepath.Join(root, "docs/requirements/REQ-001.md")
	out := entryQuestionnaireHook(t, root, "PreToolUse", path)
	var envelope struct {
		Output struct {
			Context    string `json:"additionalContext"`
			Permission string `json:"permissionDecision"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal([]byte(out), &envelope); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(envelope.Output.Context, path) || envelope.Output.Permission == "deny" {
		t.Fatalf("no exact draft or unexpected denial: %s", out)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(data, []byte("> 状态：draft\n")) || !bytes.Contains(data, []byte("> UI impact：unknown\n")) {
		t.Fatalf("draft was not honest: %s", data[:500])
	}
	data = append(data, []byte("\nHuman questions remain unanswered.\n")...)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	entryQuestionnaireHook(t, root, "PreToolUse", path)
	after, _ := os.ReadFile(path)
	if !bytes.Equal(data, after) {
		t.Fatal("repeat listener overwrote draft")
	}
	entryQuestionnaireHook(t, root, "PostToolUse", path)
	assertEntryInactive(t, root)
}

func TestEntryQuestionnaireStopReadsWithoutPreparingDraft(t *testing.T) {
	root := initEntryQuestionnaireRoot(t, nil)
	path := filepath.Join(root, "docs/requirements/REQ-001.md")
	entryQuestionnaireHook(t, root, "Stop", path)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("Stop prepared draft: %v", err)
	}
	assertEntryInactive(t, root)
}

func TestEntryQuestionnaireS1NamesHumanChoicesAndSaveDoesNotBind(t *testing.T) {
	root := initEntryQuestionnaireRoot(t, map[string]string{"REQ-098.md": lockedReqBody})
	path := filepath.Join(root, ".claude/submissions/requirement-binding.md")
	out := entryQuestionnaireHook(t, root, "PreToolUse", path)
	if !strings.Contains(out, path) || !strings.Contains(out, "actual human binding approval") || !strings.Contains(out, "release destination") {
		t.Fatalf("incomplete binding questions: %s", out)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(data, []byte(filepath.Join(root, "docs/requirements/REQ-098.md"))) {
		t.Fatalf("actual requirement path absent: %s", data)
	}
	if err := os.WriteFile(path, append(data, []byte("\nAgent-authored approve text is not a human decision.\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	entryQuestionnaireHook(t, root, "PostToolUse", path)
	assertEntryInactive(t, root)
}

func assertEntryInactive(t *testing.T, root string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, ".claude/loop-state.json"))
	if err != nil {
		t.Fatal(err)
	}
	var state map[string]any
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatal(err)
	}
	if state["lifecycle"].(map[string]any)["state"] != "inactive" {
		t.Fatalf("draft/save created a transition: %s", data)
	}
	if bound, _ := state["bound_req"].(map[string]any); bound != nil && bound["id"] != "" {
		t.Fatalf("draft/save bound a requirement: %+v", bound)
	}
}
