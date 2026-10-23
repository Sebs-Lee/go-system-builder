package cli_test

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/entroforge/go-system-builder/internal/cli"
	"github.com/entroforge/go-system-builder/internal/runtime"
	fixtures "github.com/entroforge/go-system-builder/tests/fixtures/req039"
)

func TestS4OriginalSavePreparesS5WithoutMainMetadataOrConfirmation(t *testing.T) {
	root, packet := newTaskPlanningAnswerFixture(t, func(root string) {
		for _, rel := range []string{"docs/control/agent-protocol.md", "docs/rules"} {
			copyS6RecordContextTree(t, filepath.Join(fixtures.RepoRoot(t), rel), filepath.Join(root, rel))
		}
		for _, rel := range []string{"agents/document-verifier.md", "skills/document-verification", "skills/specification-planning", "skills/api-contracts", "skills/testing-strategy", "skills/agent-dispatch"} {
			copyS6RecordContextTree(t, filepath.Join(fixtures.RepoRoot(t), rel), filepath.Join(root, ".claude", rel))
		}
	})
	stageJSON(t, root, packet.AnswerPath, completedTaskPlanningDraft(packet))
	path := filepath.Join(root, packet.AnswerPath)
	body, _ := json.Marshal(map[string]any{
		"hook_event_name": "PostToolUse", "session_id": "normal-s4-to-s5", "cwd": root,
		"tool_name": "Write", "tool_input": map[string]any{"file_path": path}, "tool_response": map[string]any{},
	})
	var out, diag bytes.Buffer
	if code := cli.Run([]string{"hook", "--event", "PostToolUse", "--root", root}, bytes.NewReader(body), &out, &diag); code != 0 {
		t.Fatalf("original save failed %d: %s %s", code, &out, &diag)
	}
	snapshot := stageSnapshot(t, root)
	if snapshot.State["lifecycle"].(map[string]any)["state"] != "document_verification" {
		t.Fatalf("normal S4 outlet did not enter S5: %s %s", &out, &diag)
	}
	packets, err := runtime.S5Packets(root, snapshot.State, "")
	if err != nil || len(packets) != 2 {
		t.Fatalf("S4 completion left machine preparation to Main: packets=%d err=%v out=%s diag=%s", len(packets), err, &out, &diag)
	}
	for _, p := range packets {
		if _, found, err := runtime.FindWorkPacket(snapshot.State, p.ID); err != nil || !found {
			t.Fatalf("S5 responsibility not actually issued: %s %v", p.ID, err)
		}
		for _, ref := range []string{p.Binding.Owner, p.Binding.Responsibility, filepath.Join(root, p.ViewPath), filepath.Join(root, p.AnswerPath)} {
			if !strings.Contains(out.String(), ref) {
				t.Fatalf("handoff did not give actual S5 dispatch reference %q: %s", ref, &out)
			}
		}
	}
}
