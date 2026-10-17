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

func TestS11QuestionnaireHookDeliversVerifiableS10PackageAndKeepsDecisionHuman(t *testing.T) {
	root, acceptancePath := s10QuestionnaireHookRoot(t)
	releasePath := prepareS10QuestionnaireReleaseAudit(t, root, acceptancePath)
	fillS10QuestionnaireAnswer(t, releasePath, "release_audit", "approved")

	request := map[string]any{
		"hook_event_name": "PostToolUse", "cwd": root, "tool_name": "Write",
		"tool_input":    map[string]any{"file_path": releasePath},
		"tool_response": map[string]any{"file_path": releasePath},
	}
	out := callS10QuestionnaireHook(t, root, request)
	if !strings.Contains(out, "registered as") {
		t.Fatalf("release-audit save did not use the existing S10 consumer: %s", out)
	}
	statePath := filepath.Join(root, ".claude", "loop-state.json")
	state := readS10QuestionnaireState(t, statePath)
	if state["lifecycle"].(map[string]any)["state"] != "release_audit" {
		t.Fatalf("PostToolUse directly advanced the cursor after S10 save: %#v", state["lifecycle"])
	}

	request = map[string]any{"hook_event_name": "PreToolUse", "cwd": root, "tool_name": "Read"}
	out = callS10QuestionnaireHook(t, root, request)
	state = readS10QuestionnaireState(t, statePath)
	if state["lifecycle"].(map[string]any)["state"] != "awaiting_human_release" {
		t.Fatalf("original Controller did not route both current S10 packages to S11: %#v; Hook=%s", state["lifecycle"], out)
	}
	acceptanceEvidencePath := s10QuestionnaireEvidencePath(t, root, state, "acceptance")
	releaseEvidencePath := s10QuestionnaireEvidencePath(t, root, state, "release_audit")
	decisionPath := filepath.Join(root, ".claude", "decisions", "hd-s11-loop-test-g1-r1.json")
	for _, exactPath := range []string{acceptancePath, releasePath, acceptanceEvidencePath, releaseEvidencePath, decisionPath} {
		if !strings.Contains(out, exactPath) {
			t.Errorf("S11 Hook omitted exact current package/application path %s: %s", exactPath, out)
		}
	}
	if _, err := os.Stat(decisionPath); err != nil {
		t.Fatalf("PreToolUse did not prepare the exact human application: %v", err)
	}

	if err := os.Remove(decisionPath); err != nil {
		t.Fatal(err)
	}
	for _, event := range []string{"Stop", "TeammateIdle"} {
		payload := map[string]any{"hook_event_name": event, "cwd": root}
		if event == "TeammateIdle" {
			payload["teammate_name"] = "release-reviewer"
		}
		_, _, code := invokeS10QuestionnaireHook(t, root, payload)
		if code != 0 && code != 2 {
			t.Fatalf("%s Hook failed unexpectedly with code %d", event, code)
		}
		if _, err := os.Stat(decisionPath); !os.IsNotExist(err) {
			t.Fatalf("%s created the S11 human application: %v", event, err)
		}
	}

	request["hook_event_name"] = "PreToolUse"
	out = callS10QuestionnaireHook(t, root, request)
	if !strings.Contains(out, decisionPath) {
		t.Fatalf("next writable Hook did not prepare and name the exact application path: %s", out)
	}
	agentAttempt := []byte(`{"decision":"approve","disposition":"approve","decision_id":"hd-s11-loop-test-g1-r1","reason":"Agent self approval attempt","approved_by":"agent-self-approval","runtime_id":"loop-test","target_cursor":{"state":"awaiting_human_release","phase":null},"scope":"runtime_release:loop-test"}` + "\n")
	if err := os.WriteFile(decisionPath, agentAttempt, 0o644); err != nil {
		t.Fatal(err)
	}
	out = callS10QuestionnaireHook(t, root, request)
	if !strings.Contains(out, decisionPath) {
		t.Fatalf("repeat prepare omitted the existing application path: %s", out)
	}
	preserved, err := os.ReadFile(decisionPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(preserved, agentAttempt) {
		t.Fatalf("repeat prepare overwrote the human-editable application: %q", preserved)
	}

	request = map[string]any{
		"hook_event_name": "PostToolUse", "cwd": root, "agent_id": "agent-self-approval", "tool_name": "Write",
		"tool_input":    map[string]any{"file_path": decisionPath},
		"tool_response": map[string]any{"file_path": decisionPath},
	}
	if out = callS10QuestionnaireHook(t, root, request); out == "" {
		t.Fatal("self-approval save unexpectedly produced no observer output")
	}
	state = readS10QuestionnaireState(t, statePath)
	if state["lifecycle"].(map[string]any)["state"] != "awaiting_human_release" {
		t.Fatalf("Agent-authored approve ended the human gateway: %#v", state["lifecycle"])
	}
	for _, entry := range state["evidence"].([]any) {
		row, _ := entry.(map[string]any)
		if row != nil && row["kind"] == "human_decision" {
			t.Fatalf("Agent-authored application registered human_decision evidence: %#v", row)
		}
	}
}

func TestS10RoutedQuestionnaireOutcomesDoNotDirectlyAdvanceTheirCursor(t *testing.T) {
	t.Run("acceptance review_required", func(t *testing.T) {
		root, acceptancePath := s10QuestionnaireHookRoot(t)
		callS10QuestionnaireHook(t, root, map[string]any{"hook_event_name": "PreToolUse", "cwd": root, "tool_name": "Read"})
		fillS10QuestionnaireAnswer(t, acceptancePath, "acceptance", "review_required")
		request := map[string]any{
			"hook_event_name": "PostToolUse", "cwd": root, "tool_name": "Write",
			"tool_input":    map[string]any{"file_path": acceptancePath},
			"tool_response": map[string]any{"file_path": acceptancePath},
		}
		callS10QuestionnaireHook(t, root, request)
		state := readS10QuestionnaireState(t, filepath.Join(root, ".claude", "loop-state.json"))
		if state["lifecycle"].(map[string]any)["state"] != "acceptance" {
			t.Fatalf("review_required save directly advanced the lifecycle: %#v", state["lifecycle"])
		}
		if !hasS10QuestionnaireEvidence(root, state, "acceptance", "review_required") {
			t.Fatal("review_required professional answer did not reach the original acceptance evidence consumer")
		}
		callS10QuestionnaireHook(t, root, map[string]any{"hook_event_name": "PreToolUse", "cwd": root, "tool_name": "Read"})
		state = readS10QuestionnaireState(t, filepath.Join(root, ".claude", "loop-state.json"))
		lifecycle := state["lifecycle"].(map[string]any)["state"]
		if lifecycle == "release_audit" || lifecycle == "awaiting_human_release" || lifecycle == "release_authorized" {
			t.Fatalf("review_required outcome advanced onto the release path: %v", lifecycle)
		}
	})

	t.Run("release audit blocked", func(t *testing.T) {
		root, acceptancePath := s10QuestionnaireHookRoot(t)
		releasePath := prepareS10QuestionnaireReleaseAudit(t, root, acceptancePath)
		fillS10QuestionnaireAnswer(t, releasePath, "release_audit", "blocked")
		request := map[string]any{
			"hook_event_name": "PostToolUse", "cwd": root, "tool_name": "Write",
			"tool_input":    map[string]any{"file_path": releasePath},
			"tool_response": map[string]any{"file_path": releasePath},
		}
		callS10QuestionnaireHook(t, root, request)
		state := readS10QuestionnaireState(t, filepath.Join(root, ".claude", "loop-state.json"))
		if state["lifecycle"].(map[string]any)["state"] != "release_audit" {
			t.Fatalf("blocked release-audit save directly advanced the lifecycle: %#v", state["lifecycle"])
		}
		if !hasS10QuestionnaireEvidence(root, state, "release_audit", "blocked") {
			t.Fatal("blocked professional answer did not reach the original release-audit evidence consumer")
		}
		callS10QuestionnaireHook(t, root, map[string]any{"hook_event_name": "PreToolUse", "cwd": root, "tool_name": "Read"})
		state = readS10QuestionnaireState(t, filepath.Join(root, ".claude", "loop-state.json"))
		lifecycle := state["lifecycle"].(map[string]any)["state"]
		if lifecycle == "awaiting_human_release" || lifecycle == "release_authorized" {
			t.Fatalf("blocked release-audit outcome advanced to the human release path: %v", lifecycle)
		}
	})
}

func prepareS10QuestionnaireReleaseAudit(t *testing.T, root, acceptancePath string) string {
	t.Helper()
	seedS10QuestionnaireCleanRound(t, root)
	out := callS10QuestionnaireHook(t, root, map[string]any{"hook_event_name": "PreToolUse", "cwd": root, "tool_name": "Read"})
	if !strings.Contains(out, acceptancePath) {
		t.Fatalf("acceptance Hook omitted its exact answer path: %s", out)
	}
	fillS10QuestionnaireAnswer(t, acceptancePath, "acceptance", "pass")
	callS10QuestionnaireHook(t, root, map[string]any{
		"hook_event_name": "PostToolUse", "cwd": root, "tool_name": "Write",
		"tool_input":    map[string]any{"file_path": acceptancePath},
		"tool_response": map[string]any{"file_path": acceptancePath},
	})
	out = callS10QuestionnaireHook(t, root, map[string]any{"hook_event_name": "PreToolUse", "cwd": root, "tool_name": "Read"})
	state := readS10QuestionnaireState(t, filepath.Join(root, ".claude", "loop-state.json"))
	if state["lifecycle"].(map[string]any)["state"] != "release_audit" {
		t.Fatalf("Controller did not advance the genuinely registered acceptance package: %#v; Hook=%s", state["lifecycle"], out)
	}
	releasePath := filepath.Join(root, "s10", "loop-test", "g1", "r1", "release_audit-manifest.json")
	if !strings.Contains(out, releasePath) {
		t.Fatalf("release-audit Hook omitted its exact answer path: %s", out)
	}
	return releasePath
}

func s10QuestionnaireEvidencePath(t *testing.T, root string, state map[string]any, kind string) string {
	t.Helper()
	for _, raw := range state["evidence"].([]any) {
		entry, _ := raw.(map[string]any)
		if entry != nil && entry["kind"] == kind {
			relative, _ := entry["path"].(string)
			if relative == "" {
				t.Fatalf("%s evidence has no registered artifact path", kind)
			}
			return filepath.Join(root, filepath.FromSlash(relative))
		}
	}
	t.Fatalf("no %s package evidence is registered", kind)
	return ""
}

func hasS10QuestionnaireEvidence(root string, state map[string]any, kind, conclusion string) bool {
	for _, raw := range state["evidence"].([]any) {
		entry, _ := raw.(map[string]any)
		if entry == nil || entry["kind"] != kind {
			continue
		}
		path, _ := entry["path"].(string)
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil {
			continue
		}
		var envelope map[string]any
		if json.Unmarshal(data, &envelope) == nil && envelope["conclusion"] == conclusion {
			return true
		}
	}
	return false
}

func invokeS10QuestionnaireHook(t *testing.T, root string, payload map[string]any) (string, string, int) {
	t.Helper()
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	event, _ := payload["hook_event_name"].(string)
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"hook", "--event", event, "--root", root}, bytes.NewReader(data), &stdout, &stderr)
	return stdout.String(), stderr.String(), code
}
