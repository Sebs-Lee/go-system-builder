package cli_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/entroforge/go-system-builder/internal/submission"
	"github.com/entroforge/go-system-builder/internal/transition"
	"github.com/entroforge/go-system-builder/internal/workpacket"
	"github.com/entroforge/go-system-builder/internal/workspace"
	"github.com/entroforge/go-system-builder/internal/workstate"
	fixtures "github.com/entroforge/go-system-builder/tests/fixtures/req039"
)

func currentS5ReworkFixture(t *testing.T) (string, []workpacket.Packet) {
	t.Helper()
	root, p := newTaskPlanningAnswerFixture(t)
	for _, rel := range []string{"docs/control/agent-protocol.md", "docs/rules"} {
		copyS6RecordContextTree(t, filepath.Join(fixtures.RepoRoot(t), rel), filepath.Join(root, rel))
	}
	fixtures.CommitCurrentFixture(t, root)
	stageJSON(t, root, p.AnswerPath, completedTaskPlanningDraft(p))
	s, err := submission.New(root)
	if err != nil {
		t.Fatal(err)
	}
	if r, err := s.SaveTaskPlanning(context.Background(), p.AnswerPath); err != nil || !r.Accepted {
		t.Fatalf("S4 original reception: %+v %v", r, err)
	}
	settleMainPlanningFollowups(t, root)
	return root, registerAcceptedS5Reviewers(t, root, "TASK-039-01", "docs/dev/tasks/TASK-039-01.md")
}

func TestS5ReworkRetiresOriginalWorkersAndPreservesReceivedReplay(t *testing.T) {
	root, packets := currentS5ReworkFixture(t)
	stageCLI(t, "runtime", "workspace", "bind", "--root", root)
	for _, p := range packets {
		stageCLI(t, "runtime", "workspace", "prepare", "--root", root, "--assignment", p.Binding.AssignmentID, "--agent", p.Binding.Owner)
	}
	before := stageSnapshot(t, root)
	registry, err := workspace.Decode(before.State)
	if err != nil || registry == nil || len(registry.Executions) != 2 {
		t.Fatalf("original review executions: %+v %v", registry, err)
	}
	manifestPaths := map[string]string{}
	for _, raw := range before.State["entities"].(map[string]any)["teams"].([]any) {
		row := raw.(map[string]any)
		path := row["manifest_ref"].(string)
		data, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			t.Fatal(err)
		}
		manifestPaths[path] = string(data)
	}
	answers := map[string]workpacket.Draft{}
	for i, p := range packets {
		e := registry.Executions[p.Binding.AssignmentID]
		paths, err := registry.WorkerSubmissionPaths(e)
		if err != nil {
			t.Fatal(err)
		}
		verdict := "pass"
		if i == 1 {
			verdict = "fix_required"
		}
		draft := completedDraft(p, verdict)
		if verdict == "fix_required" {
			draft.Answer.Findings = []workpacket.Finding{{Path: "docs/dev/tasks/TASK-039-01.md#scope", Expected: "all deliverables named", Observed: "a concrete missing output"}}
		}
		answers[p.ID] = draft
		receipt := saveStageHook(t, root, e.Path, p.Binding.Owner, paths.Answer, draft)
		if !receipt.Accepted || !receipt.WorkComplete {
			t.Fatalf("independent judgment was not received: %+v", receipt)
		}
		if i == 0 {
			current, err := workspace.Decode(stageSnapshot(t, root).State)
			if err != nil || len(current.Executions) != 2 || len(current.History) != 0 {
				t.Fatalf("partial review prematurely retired the current pair: %+v %v", current, err)
			}
		}
	}
	after := stageSnapshot(t, root)
	if after.State["lifecycle"].(map[string]any)["state"] != "planning" {
		t.Fatalf("original finding did not route TR-004: %v", after.State["lifecycle"])
	}
	retired, err := workspace.Decode(after.State)
	if err != nil || len(retired.Executions) != 0 || len(retired.History) != 2 {
		t.Fatalf("original executions were not moved to history: %+v %v", retired, err)
	}
	for _, e := range retired.History {
		if e.Status != "retired" {
			t.Fatalf("historical execution remained dispatchable: %+v", e)
		}
		if _, err := os.Stat(e.Path); err != nil {
			t.Fatalf("rework removed the original worktree: %v", err)
		}
	}
	for _, raw := range after.State["entities"].(map[string]any)["agents"].([]any) {
		if workstate.AgentActivity(raw.(map[string]any)) != "stopped" {
			t.Fatalf("superseded review responsibility remained active: %#v", raw)
		}
	}
	for path, data := range manifestPaths {
		actual, err := os.ReadFile(filepath.Join(root, path))
		if err != nil || string(actual) != data {
			t.Fatalf("rework overwrote the historical manifest %s: %v", path, err)
		}
	}
	for _, p := range packets {
		e := registry.Executions[p.Binding.AssignmentID]
		paths, err := registry.WorkerSubmissionPaths(e)
		if err != nil {
			t.Fatal(err)
		}
		stateBytes, _ := os.ReadFile(filepath.Join(root, ".claude/loop-state.json"))
		journalBytes, _ := os.ReadFile(filepath.Join(root, ".claude/loop-events.jsonl"))
		r := saveStageHook(t, e.Path, e.Path, p.Binding.Owner, paths.Answer, answers[p.ID])
		stateAfter, _ := os.ReadFile(filepath.Join(root, ".claude/loop-state.json"))
		journalAfter, _ := os.ReadFile(filepath.Join(root, ".claude/loop-events.jsonl"))
		if !r.Accepted || !r.WorkComplete || string(stateBytes) != string(stateAfter) || string(journalBytes) != string(journalAfter) {
			t.Fatalf("retired Worker replay lost the original receipt or wrote control: %+v", r)
		}
	}
}

func TestS5DirectTR003CannotSkipCurrentIndependentPair(t *testing.T) {
	root, packets := currentS5ReworkFixture(t)
	p := packets[0]
	stageJSON(t, root, p.AnswerPath, completedDraft(p, "pass"))
	s, err := submission.New(root)
	if err != nil {
		t.Fatal(err)
	}
	r, err := s.Save(context.Background(), p.Binding.Owner, p.AnswerPath)
	if err != nil || !r.Accepted || len(r.DomainRefs) != 1 {
		t.Fatalf("first independent reception: %+v %v", r, err)
	}
	before := stageSnapshot(t, root)
	stateBytes, _ := os.ReadFile(s.StatePath)
	journalBytes, _ := os.ReadFile(s.JournalPath)
	_, err = transition.Apply(root, s.StatePath, s.JournalPath, transition.Request{
		TransitionID: "TR-003", Actor: "orchestrator", ExpectedRevision: before.Revision,
		Evidence: map[string]string{"document_review_record": r.DomainRefs[0]}, Files: s.Files,
	})
	if err == nil {
		t.Fatal("direct TR-003 skipped the second current independent review")
	}
	afterState, _ := os.ReadFile(s.StatePath)
	afterJournal, _ := os.ReadFile(s.JournalPath)
	if stageSnapshot(t, root).Revision != before.Revision || string(stateBytes) != string(afterState) || string(journalBytes) != string(afterJournal) {
		t.Fatal("rejected incomplete review pair changed the Runtime/journal")
	}
}

func TestS5ReworkCancelsUnreceivedPeerWithoutAcceptingANewAnswer(t *testing.T) {
	root, packets := currentS5ReworkFixture(t)
	stageCLI(t, "runtime", "workspace", "bind", "--root", root)
	for _, p := range packets {
		stageCLI(t, "runtime", "workspace", "prepare", "--root", root, "--assignment", p.Binding.AssignmentID, "--agent", p.Binding.Owner)
	}
	registry, err := workspace.Decode(stageSnapshot(t, root).State)
	if err != nil || registry == nil {
		t.Fatal(err)
	}
	first, peer := packets[0], packets[1]
	e := registry.Executions[first.Binding.AssignmentID]
	paths, err := registry.WorkerSubmissionPaths(e)
	if err != nil {
		t.Fatal(err)
	}
	draft := completedDraft(first, "fix_required")
	draft.Answer.Findings = []workpacket.Finding{{Path: "docs/dev/tasks/TASK-039-01.md#scope", Expected: "all deliverables named", Observed: "a concrete missing output"}}
	if r := saveStageHook(t, root, e.Path, first.Binding.Owner, paths.Answer, draft); !r.Accepted {
		t.Fatalf("original finding reception failed: %+v", r)
	}
	retired, err := workspace.Decode(stageSnapshot(t, root).State)
	if err != nil || len(retired.Executions) != 0 || len(retired.History) != 2 {
		t.Fatalf("unreceived peer was not retired: %+v %v", retired, err)
	}
	peerExecution := registry.Executions[peer.Binding.AssignmentID]
	peerPaths, err := registry.WorkerSubmissionPaths(peerExecution)
	if err != nil {
		t.Fatal(err)
	}
	stageJSON(t, peerExecution.Path, peerPaths.Answer, completedDraft(peer, "pass"))
	s, err := submission.New(root)
	if err != nil {
		t.Fatal(err)
	}
	beforeState, _ := os.ReadFile(s.StatePath)
	beforeJournal, _ := os.ReadFile(s.JournalPath)
	if r, err := s.Save(context.Background(), peer.Binding.Owner, filepath.Join(peerExecution.Path, peerPaths.Answer)); err == nil || r.Accepted {
		t.Fatalf("retired unreceived peer submitted a new judgment: %+v %v", r, err)
	}
	if code, msg := stageHookCode(t, peerExecution.Path, peerExecution.Path, peer.Binding.Owner, "Stop", ""); code != 0 || strings.Contains(msg, "workspace root unavailable") {
		t.Fatalf("cancelled peer was asked to complete an obsolete questionnaire: %d %s", code, msg)
	}
	afterState, _ := os.ReadFile(s.StatePath)
	afterJournal, _ := os.ReadFile(s.JournalPath)
	if string(beforeState) != string(afterState) || string(beforeJournal) != string(afterJournal) {
		t.Fatal("retired unreceived peer changed the Runtime/journal")
	}
}
