package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/entroforge/go-system-builder/internal/cli"
	"github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/submission"
	"github.com/entroforge/go-system-builder/internal/workpacket"
	fixtures "github.com/entroforge/go-system-builder/tests/fixtures/req039"
)

func TestS4OriginalSavePreparesS5WithoutMainMetadataOrConfirmation(t *testing.T) {
	root, packet := s4HandoffFixture(t)
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

func s4HandoffFixture(t *testing.T) (string, workpacket.TaskPlanningPacket) {
	return newTaskPlanningAnswerFixture(t, func(root string) {
		for _, rel := range []string{"docs/control/agent-protocol.md", "docs/rules"} {
			copyS6RecordContextTree(t, filepath.Join(fixtures.RepoRoot(t), rel), filepath.Join(root, rel))
		}
		for _, rel := range []string{"agents/document-verifier.md", "skills/document-verification", "skills/specification-planning", "skills/api-contracts", "skills/testing-strategy", "skills/agent-dispatch"} {
			copyS6RecordContextTree(t, filepath.Join(fixtures.RepoRoot(t), rel), filepath.Join(root, ".claude", rel))
		}
	})
}

func TestS4NamedOutletsDeliverPreparedS5InSameResponse(t *testing.T) {
	for _, action := range []string{"consume", "recover", "post"} {
		t.Run(action, func(t *testing.T) {
			root, p := s4HandoffFixture(t)
			stageJSON(t, root, p.AnswerPath, completedTaskPlanningDraft(p))
			if action == "recover" || action == "post" {
				service, err := submission.New(root)
				if err != nil {
					t.Fatal(err)
				}
				ctx, cancel := pendingMainFollowupContext(service, p.ID)
				r, err := service.SaveTaskPlanning(ctx, p.AnswerPath)
				cancel()
				if err != nil || !r.Accepted {
					t.Fatalf("pending reception: %+v %v", r, err)
				}
				record, found, lookupErr := runtime.FindSubmission(stageSnapshot(t, root).State, p.ID)
				if lookupErr != nil || !found {
					t.Fatalf("original receipt %v", lookupErr)
				}
				writeLegacyS3Evaluation(t, root, record) // reconstruct the same historical no_transition stamp for S4
			}
			args := []string{"runtime", "stage-answer", action, "--root", root, "--agent", "Main"}
			if action == "consume" {
				args = append(args, "--path", p.AnswerPath)
			}
			var output string
			if action == "post" {
				body, _ := json.Marshal(map[string]any{"hook_event_name": "PostToolUse", "session_id": "s4-original-post-followup", "cwd": root, "tool_name": "Read", "tool_input": map[string]any{"file_path": filepath.Join(root, "docs/control/agent-protocol.md")}, "tool_response": map[string]any{}})
				var out, diag bytes.Buffer
				if code := cli.Run([]string{"hook", "--event", "PostToolUse", "--root", root}, bytes.NewReader(body), &out, &diag); code != 0 {
					t.Fatalf("post %d: %s %s", code, &out, &diag)
				}
				output = out.String()
			} else {
				output = stageCLI(t, args...)
			}
			snap := stageSnapshot(t, root)
			packets, err := runtime.S5Packets(root, snap.State, "")
			if err != nil || len(packets) != 2 {
				t.Fatalf("packets %d %v %s", len(packets), err, output)
			}
			for _, packet := range packets {
				for _, ref := range []string{packet.ViewPath, packet.AnswerPath, "source root", root} {
					if !strings.Contains(output, ref) {
						t.Fatalf("missing %s: %s", ref, output)
					}
				}
			}
			before := snap.Revision
			again := stageCLI(t, "runtime", "stage-answer", "consume", "--root", root, "--agent", "Main", "--path", p.AnswerPath)
			if !strings.Contains(again, "source root") || stageSnapshot(t, root).Revision != before {
				t.Fatal("replay changed original reception", again)
			}
		})
	}
}

func TestS4PreparationFailureSurvivesSessionAndIssueRepair(t *testing.T) {
	root, p := s4HandoffFixture(t)
	missing := filepath.Join(root, ".claude/skills/document-verification/SKILL.md")
	data, err := os.ReadFile(missing)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(missing); err != nil {
		t.Fatal(err)
	}
	stageJSON(t, root, p.AnswerPath, completedTaskPlanningDraft(p))
	out := stageCLI(t, "runtime", "stage-answer", "consume", "--root", root, "--agent", "Main", "--path", p.AnswerPath)
	if !strings.Contains(out, "stage-answer issue") || !strings.Contains(out, "document-verification") {
		t.Fatalf("failure lost repair %s", out)
	}
	snap := stageSnapshot(t, root)
	if snap.State["lifecycle"].(map[string]any)["state"] != "document_verification" {
		t.Fatal("transition revoked")
	}
	session := s3DeliveredContext(t, designGuidanceEvent(t, root, "s5-prepare-failure-new-session", "SessionStart", "Read"))
	if !strings.Contains(session, "document-verification") || !strings.Contains(session, "stage-answer issue") {
		t.Fatal("session lost cause", session)
	}
	if stageSnapshot(t, root).Revision != snap.Revision {
		t.Fatal("session mutated preparation")
	}
	if err := os.WriteFile(missing, data, 0600); err != nil {
		t.Fatal(err)
	}
	stageCLI(t, "runtime", "stage-answer", "issue", "--root", root, "--agent", "Main")
	repaired := stageSnapshot(t, root)
	if diagnostic := submission.PlanningDiagnostic(root, repaired.State, "preparation"); diagnostic != "" {
		t.Fatal("stale cause", diagnostic)
	}
	packets, err := runtime.S5Packets(root, repaired.State, "")
	if err != nil || len(packets) != 2 {
		t.Fatalf("repair %d %v", len(packets), err)
	}
	// A later partial publication failure must retain the peer's authored
	// bytes and both original packet identities when issue resumes it.
	authored := []byte("{\"author_work\":\"keep this answer\"}\n")
	answerPath := filepath.Join(root, packets[0].AnswerPath)
	if err := os.WriteFile(answerPath, authored, 0600); err != nil {
		t.Fatal(err)
	}
	blockedView := filepath.Join(root, packets[1].ViewPath)
	if err := os.Remove(blockedView); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(blockedView, 0700); err != nil {
		t.Fatal(err)
	}
	var failedOut, failedErr bytes.Buffer
	if code := cli.Run([]string{"runtime", "stage-answer", "issue", "--root", root, "--agent", "Main"}, nil, &failedOut, &failedErr); code == 0 {
		t.Fatal("blocked publication succeeded")
	}
	if err := os.Remove(blockedView); err != nil {
		t.Fatal(err)
	}
	stageCLI(t, "runtime", "stage-answer", "issue", "--root", root, "--agent", "Main")
	if actual, err := os.ReadFile(answerPath); err != nil || !bytes.Equal(actual, authored) {
		t.Fatalf("peer answer overwritten: %s %v", actual, err)
	}
	resumed, err := runtime.S5Packets(root, stageSnapshot(t, root).State, "")
	if err != nil || len(resumed) != 2 || resumed[0].ID != packets[0].ID || resumed[1].ID != packets[1].ID {
		t.Fatalf("partial recovery replaced original packets %v %v", resumed, err)
	}
}

func TestS4ReceptionFieldFailurePersistsAcrossSessionAndClears(t *testing.T) {
	root, p := s4HandoffFixture(t)
	draft := completedTaskPlanningDraft(p)
	invalid := "/outside/index.md"
	draft.Answer.IndexPath = &invalid
	stageJSON(t, root, p.AnswerPath, draft)
	service, err := submission.New(root)
	if err != nil {
		t.Fatal(err)
	}
	r, err := service.SaveTaskPlanning(context.Background(), p.AnswerPath)
	if err == nil || r.Accepted {
		t.Fatal("invalid path received")
	}
	for _, event := range []string{"PreToolUse", "SessionStart"} {
		text := s3DeliveredContext(t, designGuidanceEvent(t, root, "s4-field-"+event, event, "Read"))
		if !strings.Contains(text, "index_path") || !strings.Contains(text, "outside") {
			t.Fatalf("%s lost cause %s", event, text)
		}
	}
	// The original source failure remains actionable after a fresh session.
	stageJSON(t, root, p.AnswerPath, completedTaskPlanningDraft(p))
	sourcePath := p.Inputs.Contracts[0].Path
	originalSource, readErr := os.ReadFile(filepath.Join(root, sourcePath))
	if readErr != nil {
		t.Fatal(readErr)
	}
	writeContractSource(t, root, sourcePath, string(originalSource)+"\nchanged accepted contract description\n")
	fixtures.CommitCurrentFixture(t, root)
	service, err = submission.New(root)
	if err != nil {
		t.Fatal(err)
	}
	r, err = service.SaveTaskPlanning(context.Background(), p.AnswerPath)
	if err == nil || r.Accepted {
		t.Fatal("changed original source accepted")
	}
	sourceSession := s3DeliveredContext(t, designGuidanceEvent(t, root, "s4-source-new-session", "SessionStart", "Read"))
	if !strings.Contains(sourceSession, sourcePath) {
		t.Fatal("session lost precise source", sourceSession)
	}
	writeContractSource(t, root, sourcePath, string(originalSource))
	fixtures.CommitCurrentFixture(t, root)
	service, err = submission.New(root)
	if err != nil {
		t.Fatal(err)
	}
	origin := stageSnapshot(t, root)
	stageJSON(t, root, p.AnswerPath, completedTaskPlanningDraft(p))
	r, err = service.SaveTaskPlanning(context.Background(), p.AnswerPath)
	if err != nil || !r.Accepted {
		t.Fatalf("repair %+v %v", r, err)
	}
	if diagnostic := submission.PlanningDiagnostic(root, origin.State, "reception"); diagnostic != "" {
		t.Fatal("stale field cause", diagnostic)
	}
}
