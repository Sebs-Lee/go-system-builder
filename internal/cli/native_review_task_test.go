package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/entroforge/go-system-builder/internal/policy"
)

func TestNativeReviewResumeUsesLatestExactTaskAndIgnoresToolResults(t *testing.T) {
	oldRef, newRef := "/main/questions/old.md", "/main/questions/new.md"
	path := filepath.Join(t.TempDir(), "native.jsonl")
	var transcript strings.Builder
	appendRow := func(content any) {
		row, _ := json.Marshal(map[string]any{"type": "user", "message": map[string]any{"role": "user", "content": content}})
		transcript.Write(row)
		transcript.WriteByte('\n')
	}
	appendRow("First review " + oldRef)
	appendRow([]map[string]string{{"type": "text", "text": "Now review " + newRef}})
	appendRow([]map[string]string{{"type": "tool_result", "text": "Quoted old output " + oldRef}})
	appendRow("Continue your work")
	if err := os.WriteFile(path, []byte(transcript.String()), 0600); err != nil {
		t.Fatal(err)
	}
	text := nativeReviewTaskText(policy.Input{AgentTranscriptPath: path, LastAssistantMessage: "Finished " + newRef}, []string{oldRef, newRef})
	if strings.Contains(text, oldRef) || !nativeS6ExactReference(text, newRef) {
		t.Fatalf("resumed responsibility was confused with initial task/tool output: %s", text)
	}
	// A final handoff naming two real responsibilities is still ambiguous.
	text = nativeReviewTaskText(policy.Input{AgentTranscriptPath: path, LastAssistantMessage: oldRef + "\n" + newRef}, []string{oldRef, newRef})
	if !nativeS6ExactReference(text, oldRef) || !nativeS6ExactReference(text, newRef) {
		t.Fatal(text)
	}
}

func TestNativeS9DoesNotCaptureS8WithRetainedRepairPointer(t *testing.T) {
	state := map[string]any{"lifecycle": map[string]any{"state": "bug_resolution", "phase": "investigation"}, "review": map[string]any{"repair": map[string]any{"status": "targeted_reverification"}}}
	root := t.TempDir()
	if nativeS9Available(root, state) || len(nativeS9PlanPackets(root, state)) != 0 {
		t.Fatal("retained S9 state captured current S8 responsibility")
	}
	state["lifecycle"].(map[string]any)["phase"] = "fixing"
	if !nativeS9Available(root, state) {
		t.Fatal("actual S9 targeted work lost its lifecycle handler")
	}
}

func TestNativeReviewResumeReadsBoundedTranscriptTail(t *testing.T) {
	oldRef, newRef := "/main/old/brief.md", "/main/current/brief.md"
	var transcript strings.Builder
	appendRow := func(content any) {
		row, _ := json.Marshal(map[string]any{"type": "user", "message": map[string]any{"role": "user", "content": content}})
		transcript.Write(row)
		transcript.WriteByte('\n')
	}
	appendRow("Earlier " + oldRef)
	appendRow([]map[string]string{{"type": "tool_result", "text": strings.Repeat("x", 4<<20) + oldRef}})
	appendRow("Resumed actual task " + newRef)
	path := filepath.Join(t.TempDir(), "large-native.jsonl")
	if err := os.WriteFile(path, []byte(transcript.String()), 0600); err != nil {
		t.Fatal(err)
	}
	text := nativeReviewTaskText(policy.Input{AgentTranscriptPath: path, LastAssistantMessage: "Finished."}, []string{oldRef, newRef})
	if !nativeS6ExactReference(text, newRef) || nativeS6ExactReference(text, oldRef) {
		t.Fatalf("bounded tail lost current responsibility: %s", text)
	}
}
