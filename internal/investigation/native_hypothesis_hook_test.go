package investigation_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/entroforge/go-system-builder/internal/cli"
	"github.com/entroforge/go-system-builder/internal/investigation"
	"github.com/entroforge/go-system-builder/internal/submission"
)

func TestNativeS8StopReceivesOriginalCaseAnswerWithoutGenericPlanOrActorAlias(t *testing.T) {
	fixture, owners := prepareS8GenerationTwoFixture(t, false)
	owner := owners[0]
	before := currentS8Snapshot(t, fixture)
	packets, err := investigation.IssueHypothesisResultPackets(fixture.root, fixture.statePath, fixture.journalPath, before.Revision, owner.ownerID)
	if err != nil || len(packets) != 1 {
		t.Fatalf("original issue: %d %v", len(packets), err)
	}
	p := packets[0]
	path := writeS8ProfessionalAnswer(t, fixture.root, p, "Native Investigator observed the source-boundary discrepancy without modifying product files.")
	query := func(flag, ref, agent string) (submission.Receipt, int) {
		t.Helper()
		beforeState, beforeJournal := mustRead(t, fixture.statePath), mustRead(t, fixture.journalPath)
		var out, diagnostic bytes.Buffer
		code := cli.Run([]string{"runtime", "stage-answer", "receipt", "--root", fixture.root, "--agent", agent, flag, ref}, nil, &out, &diagnostic)
		var receipt submission.Receipt
		if code == 0 {
			if err := json.Unmarshal(out.Bytes(), &receipt); err != nil {
				t.Fatalf("public S8 receipt wire: %v %s %s", err, &out, &diagnostic)
			}
		}
		if !bytes.Equal(beforeState, mustRead(t, fixture.statePath)) || !bytes.Equal(beforeJournal, mustRead(t, fixture.journalPath)) {
			t.Fatal("S8 receipt query mutated Runtime/journal")
		}
		return receipt, code
	}
	if receipt, code := query("--packet", p.ID, owner.ownerID); code != 0 || receipt.Accepted || receipt.WorkComplete {
		t.Fatalf("an authored but unreceived answer became completion: %+v code=%d", receipt, code)
	}
	active := false
	call := func(event, actor, ref string) (int, string, string) {
		body, _ := json.Marshal(map[string]any{"hook_event_name": event, "stop_hook_active": active, "agent_id": actor, "agent_type": "investigator", "session_id": "actual-native-parent", "cwd": fixture.root, "last_assistant_message": ref, "tool_name": "Write", "tool_input": map[string]any{"file_path": filepath.Join(fixture.root, path)}})
		var out, diagnostic bytes.Buffer
		code := cli.Run([]string{"hook", "--root", fixture.root, "--event", event}, bytes.NewReader(body), &out, &diagnostic)
		return code, out.String(), diagnostic.String()
	}
	before = currentS8Snapshot(t, fixture)
	for _, event := range []string{"SubagentStart", "PreToolUse", "PostToolUse", "TeammateIdle"} {
		if code, out, diagnostic := call(event, "opaque-native-investigator", ""); code != 0 || event != "SubagentStart" && out != "" {
			t.Fatalf("native %s: %d %s %s", event, code, out, diagnostic)
		}
		if currentS8Snapshot(t, fixture).Revision != before.Revision {
			t.Fatalf("native %s changed Runtime", event)
		}
	}
	if code, _, diagnostic := call("PostToolUse", "", ""); code != 0 || currentS8Snapshot(t, fixture).Revision != before.Revision {
		t.Fatalf("Main received S8 answer before native Stop: %d %s", code, diagnostic)
	}
	answerPath := filepath.Join(fixture.root, path)
	valid, err := os.ReadFile(answerPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(answerPath, []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	if code, _, diagnostic := call("SubagentStop", "opaque-native-investigator", filepath.Join(fixture.root, p.ViewPath)); code != 2 || !strings.Contains(diagnostic, "correct") {
		t.Fatalf("professional error lost its owner: %d %s", code, diagnostic)
	}
	active = true
	if code, _, diagnostic := call("SubagentStop", "opaque-native-investigator", filepath.Join(fixture.root, p.ViewPath)); code != 0 || !strings.Contains(diagnostic, "Main owns recovery") {
		t.Fatalf("repeated correction was not bounded: %d %s", code, diagnostic)
	}
	if currentS8Snapshot(t, fixture).Revision != before.Revision {
		t.Fatal("invalid answer fabricated a Case result")
	}
	if err := os.WriteFile(answerPath, valid, 0600); err != nil {
		t.Fatal(err)
	}
	active = false

	if code, out, diagnostic := call("SubagentStop", "opaque-native-investigator", filepath.Join(fixture.root, p.ViewPath)); code != 0 || out != "" || !strings.Contains(diagnostic, "original consumer") {
		t.Fatalf("native Stop: %d %s %s", code, out, diagnostic)
	}
	mainBody, _ := json.Marshal(map[string]any{"hook_event_name": "PostToolUse", "cwd": fixture.root, "tool_name": "Agent", "tool_input": map[string]any{"prompt": filepath.Join(fixture.root, p.ViewPath)}})
	var mainOut, mainDiagnostic bytes.Buffer
	if code := cli.Run([]string{"hook", "--root", fixture.root, "--event", "PostToolUse"}, bytes.NewReader(mainBody), &mainOut, &mainDiagnostic); code != 0 {
		t.Fatalf("Main return: %d %s", code, &mainDiagnostic)
	}
	var message struct {
		Output struct {
			Text string `json:"additionalContext"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal(mainOut.Bytes(), &message); err != nil || !strings.Contains(message.Output.Text, `"accepted":true`) || !strings.Contains(message.Output.Text, p.ID) {
		t.Fatalf("Main missed original Case reception: %v %s", err, &mainOut)
	}
	accepted := currentS8Snapshot(t, fixture)
	if _, found, err := investigation.IssuedHypothesisResultReceipt(fixture.root, accepted.State, p.ID, owner.ownerID); err != nil || !found {
		t.Fatalf("original Case receipt missing: %t %v", found, err)
	}
	if code, _, diagnostic := call("SubagentStop", "fresh-native-actor-after-resume", filepath.Join(fixture.root, p.AnswerPath)); code != 0 {
		t.Fatalf("native replay: %d %s", code, diagnostic)
	}
	if currentS8Snapshot(t, fixture).Revision != accepted.Revision {
		t.Fatal("same answer after native resume changed Runtime")
	}
	canonical, found, err := investigation.IssuedHypothesisResultReceipt(fixture.root, accepted.State, p.ID, owner.ownerID)
	if err != nil || !found {
		t.Fatalf("original accepted receipt: %v", err)
	}
	casePath := filepath.Join(fixture.root, canonical.SuccessorCaseRef)
	caseBytes := mustRead(t, casePath)
	acceptedState, acceptedJournal := mustRead(t, fixture.statePath), mustRead(t, fixture.journalPath)
	writeS8ProfessionalAnswer(t, fixture.root, p, "Changed submitted content is a conflicting answer, not a receipt query.")
	if _, err := submitS8Original(fixture, accepted.Revision, p, path, owner.ownerID); err == nil || !strings.Contains(err.Error(), "conflicts") {
		t.Fatalf("actual changed submission bypassed the original digest conflict: %v", err)
	}
	if err := os.WriteFile(answerPath, []byte("broken private answer"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"corrupt-answer", "removed-author-files"} {
		if mode == "removed-author-files" {
			for _, file := range []string{answerPath, filepath.Join(fixture.root, p.ViewPath)} {
				if err := os.Remove(file); err != nil {
					t.Fatal(err)
				}
			}
		}
		for flag, ref := range map[string]string{"--packet": p.ID, "--path": answerPath} {
			receipt, code := query(flag, ref, owner.ownerID)
			if code != 0 || !receipt.Accepted || !receipt.WorkComplete || receipt.PacketID != p.ID || len(receipt.DomainRefs) != 1 || receipt.DomainRefs[0] != canonical.SuccessorCaseRef {
				t.Fatalf("%s lost original accepted result with %s: %+v code=%d", mode, flag, receipt, code)
			}
		}
		mainOut.Reset()
		mainDiagnostic.Reset()
		if code := cli.Run([]string{"hook", "--root", fixture.root, "--event", "PostToolUse"}, bytes.NewReader(mainBody), &mainOut, &mainDiagnostic); code != 0 {
			t.Fatalf("%s Main receipt return: code=%d %s", mode, code, &mainDiagnostic)
		}
		if err := json.Unmarshal(mainOut.Bytes(), &message); err != nil || !strings.Contains(message.Output.Text, `"accepted":true`) || !strings.Contains(message.Output.Text, canonical.SuccessorCaseRef) {
			t.Fatalf("%s Main asked for private answer instead of original receipt: %v %s", mode, err, &mainOut)
		}
	}
	if receipt, code := query("--packet", p.ID, "different-investigator"); code == 0 && (receipt.Accepted || receipt.WorkComplete) {
		t.Fatalf("receipt lent completion to a different owner: %+v", receipt)
	}
	if !bytes.Equal(caseBytes, mustRead(t, casePath)) || !bytes.Equal(acceptedState, mustRead(t, fixture.statePath)) || !bytes.Equal(acceptedJournal, mustRead(t, fixture.journalPath)) {
		t.Fatal("receipt recovery or changed-answer conflict rewrote original Case/Runtime/journal")
	}
	if err := os.WriteFile(casePath, append(append([]byte(nil), caseBytes...), '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	if receipt, code := query("--packet", p.ID, owner.ownerID); code == 0 && receipt.Accepted {
		t.Fatalf("author-file independence bypassed damaged canonical Case proof: %+v", receipt)
	}
}
