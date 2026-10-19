package cli_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/entroforge/go-system-builder/internal/submission"
	"github.com/entroforge/go-system-builder/internal/workpacket"
	fixtures "github.com/entroforge/go-system-builder/tests/fixtures/req039"
)

func TestTaskPlanningSaveKeepsActualCancelledTaskWithoutAuthorListProof(t *testing.T) {
	root, p := newTaskPlanningAnswerFixture(t)
	cancelled := "docs/dev/tasks/TASK-039-02.md"
	writeContractSource(t, root, cancelled, "# Task\n> Status: cancelled\n> Version: v1.0.0\n> Source REQ refs: REQ-039\n")
	fixtures.CommitCurrentFixture(t, root)
	draft := completedTaskPlanningDraft(p)
	stageJSON(t, root, p.AnswerPath, draft)
	s, err := submission.New(root)
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := s.SaveTaskPlanning(context.Background(), p.AnswerPath)
	if err != nil || !receipt.Accepted {
		t.Fatalf("actual cancelled task was made an author-list qualification: %+v %v", receipt, err)
	}
	snap := stageSnapshot(t, root)
	found := false
	for _, raw := range snap.State["evidence"].([]any) {
		ev := raw.(map[string]any)
		if ev["kind"] != "planning_task" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, ev["path"].(string)))
		if err != nil {
			t.Fatal(err)
		}
		var envelope struct {
			Subjects []workpacket.Subject          `json:"subject_refs"`
			Answer   workpacket.TaskPlanningAnswer `json:"task_planning_answer"`
		}
		if err := json.Unmarshal(data, &envelope); err != nil {
			t.Fatal(err)
		}
		if len(envelope.Answer.TaskPaths) != 1 || envelope.Answer.TaskPaths[0] != draft.Answer.TaskPaths[0] {
			t.Fatalf("professional answer was filled by system: %+v", envelope.Answer)
		}
		for _, subject := range envelope.Subjects {
			if subject.Path == cancelled {
				found = true
			}
		}
	}
	if !found {
		t.Fatal("actual cancelled task fact lost from received record")
	}
}
