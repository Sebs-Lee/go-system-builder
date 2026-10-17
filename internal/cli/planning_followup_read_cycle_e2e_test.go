package cli_test

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/entroforge/go-system-builder/internal/controller"
	"github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/submission"
)

func TestMainPlanningReadCycleConsumesAcceptedS2Followup(t *testing.T) {
	root, packet := newDesignFixture(t, "none")
	stageJSON(t, root, packet.AnswerPath, designDraft(packet))
	s, err := submission.New(root)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := pendingMainFollowupContext(s, packet.ID)
	defer cancel()
	receipt, err := s.SaveDesign(ctx, packet.AnswerPath)
	if err != nil || !receipt.Accepted || !receipt.WorkComplete {
		t.Fatalf("accept S2 before ordinary cycle: %+v %v", receipt, err)
	}
	assertReadCycleConsumesPlanningFollowup(t, s, packet.ID, "PTR-PLAN-01", "design", "contracts")
}

func TestMainPlanningReadCycleConsumesAcceptedS3Followup(t *testing.T) {
	root, packet := newContractAnswerFixture(t)
	stageJSON(t, root, packet.AnswerPath, completedContractDraft(packet))
	s, err := submission.New(root)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := pendingMainFollowupContext(s, packet.ID)
	defer cancel()
	receipt, err := s.SaveContracts(ctx, packet.AnswerPath)
	if err != nil || !receipt.Accepted || !receipt.WorkComplete {
		t.Fatalf("accept S3 before ordinary cycle: %+v %v", receipt, err)
	}
	assertReadCycleConsumesPlanningFollowup(t, s, packet.ID, "PTR-PLAN-02", "contracts", "tasks")
}

func TestMainPlanningReadCycleConsumesAcceptedS4FollowupAndClosesS5Inputs(t *testing.T) {
	root, packet := newTaskPlanningAnswerFixture(t)
	stageJSON(t, root, packet.AnswerPath, completedTaskPlanningDraft(packet))
	s, err := submission.New(root)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := pendingMainFollowupContext(s, packet.ID)
	defer cancel()
	receipt, err := s.SaveTaskPlanning(ctx, packet.AnswerPath)
	if err != nil || !receipt.Accepted || !receipt.WorkComplete {
		t.Fatalf("accept S4 before ordinary cycle: %+v %v", receipt, err)
	}
	snapshot := assertReadCycleConsumesPlanningFollowup(t, s, packet.ID, "TR-002", "tasks", "")
	formal, err := runtime.S5FormalInputs(root, snapshot.State, s.Files)
	if err != nil || len(formal.References) == 0 {
		t.Fatalf("accepted S4 closure is not consumable by S5: references=%d err=%v", len(formal.References), err)
	}
}

func pendingMainFollowupContext(s *submission.Service, packetID string) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	s.Validator = cancelValidatedReception{CandidateValidator: s.Validator, packetID: packetID, cancel: cancel}
	return ctx, cancel
}

func assertReadCycleConsumesPlanningFollowup(t *testing.T, s *submission.Service, packetID, transitionID, fromPhase, toPhase string) runtime.Snapshot {
	t.Helper()
	before := stageSnapshot(t, s.Root)
	prior, found, err := runtime.FindSubmission(before.State, packetID)
	if err != nil || !found || prior.ControlStatus != "pending" || prior.TransitionEventID != "" {
		t.Fatalf("typed Main answer did not leave one pending followup: %+v found=%t err=%v", prior, found, err)
	}

	result, err := controller.RunControlCycle(context.Background(), controller.ControlRequest{
		Root: s.Root, CWD: s.Root, Files: s.Files, Event: "PreToolUse", ToolName: "Read",
	})
	if err != nil || result.Error != "" {
		t.Fatalf("ordinary PreToolUse/Read cycle: result=%+v err=%v", result, err)
	}
	if !result.QualityGate.TransitionCommitted || result.QualityGate.CandidateTransition != transitionID {
		t.Fatalf("ordinary Read did not commit original %s transition: %+v", transitionID, result.QualityGate)
	}
	if result.Snapshot.State == nil {
		t.Fatal("ordinary Read cycle returned no committed snapshot")
	}
	record, found, err := runtime.FindSubmission(result.Snapshot.State, packetID)
	if err != nil || !found || record.ControlStatus != "advanced" || record.TransitionEventID == "" {
		t.Fatalf("original Main followup was not stamped advanced with the transition: %+v found=%t err=%v", record, found, err)
	}
	lifecycle := result.Snapshot.State["lifecycle"].(map[string]any)
	if lifecycle["state"] != "planning" && transitionID != "TR-002" {
		t.Fatalf("planning route state = %v", lifecycle)
	}
	if transitionID == "TR-002" {
		if lifecycle["state"] != "document_verification" || lifecycle["phase"] != nil {
			t.Fatalf("TR-002 cursor = %+v", lifecycle)
		}
	} else if lifecycle["phase"] != toPhase {
		t.Fatalf("%s destination phase = %v, want %s", transitionID, lifecycle["phase"], toPhase)
	}
	assertPlanningTransitionJournalEvent(t, s.Root, record.TransitionEventID, transitionID, fromPhase)
	return result.Snapshot
}

func assertPlanningTransitionJournalEvent(t *testing.T, root, eventID, transitionID, fromPhase string) {
	t.Helper()
	file, err := os.Open(filepath.Join(root, ".claude/loop-events.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 4096), 2<<20)
	for scanner.Scan() {
		var event map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			t.Fatal(err)
		}
		if event["event_id"] != eventID {
			continue
		}
		from, _ := event["from"].(map[string]any)
		beforeRevision, ok := event["before_revision"].(float64)
		if event["transition_id"] != transitionID || from["state"] != "planning" || from["phase"] != fromPhase || !ok {
			t.Fatalf("followup points to a different original transition event: %+v", event)
		}
		if event["idempotency_key"] != fmt.Sprintf("runtime:%s:%d", transitionID, int(beforeRevision)) {
			t.Fatalf("transition event has noncanonical original idempotency key: %+v", event)
		}
		if event["event_id"] != fmt.Sprintf("evt-%s-r%d", strings.ToLower(transitionID), int(beforeRevision)+1) {
			t.Fatalf("transition event identity does not match original revision: %+v", event)
		}
		return
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	t.Fatalf("journal does not contain followup's exact transition event %q", eventID)
}
