package investigation_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/entroforge/go-system-builder/internal/cli"
	"github.com/entroforge/go-system-builder/internal/investigation"
)

// The synthetic sealed-S7 fixture is established before Case intake. All S8
// registration, issuance, immutable sources and native Stop consumption below
// use the original mechanisms; no source/index alias is registered.
func TestS8TypedFindingSourceHandoffNeedsNoIndexAlias(t *testing.T) {
	content := []byte("SIMULATION_ONLY\n")
	digest := hash(content)
	refs := []string{
		"path:observation.txt#sha256=" + digest,
		fmt.Sprintf("command_output:observation.txt#sha256=%s (bytes=%d, truncated=false)", digest, len(content)),
		"command_output:withheld.txt (withheld: redaction)",
		"artifact:deleted.txt (deleted)",
		"env:SECRET (present; value never captured)",
	}
	f, owners := prepareS8GenerationTwoFixture(t, false, func(f *intakeFixture) {
		writeInvestigationBytes(t, f.root, "observation.txt", content)
		state := map[string]any{}
		if err := json.Unmarshal(mustRead(t, f.statePath), &state); err != nil {
			t.Fatal(err)
		}
		finding := state["entities"].(map[string]any)["findings"].([]any)[0].(map[string]any)
		path := filepath.Join(f.root, finding["path"].(string))
		document := map[string]any{}
		if err := json.Unmarshal(mustRead(t, path), &document); err != nil {
			t.Fatal(err)
		}
		document["evidence_refs"] = refs
		writeInvestigationJSON(t, path, document)
		finding["sha256"] = hash(mustRead(t, path))
		writeInvestigationJSON(t, f.statePath, state)
	})
	owner := owners[0]
	beforeState, beforeJournal := mustRead(t, f.statePath), mustRead(t, f.journalPath)
	writeInvestigationBytes(t, f.root, "observation.txt", []byte("changed\n"))
	if _, err := investigation.IssueHypothesisResultPackets(f.root, f.statePath, f.journalPath, -1, owner.ownerID); err == nil || !strings.Contains(err.Error(), "original Finding") {
		t.Fatalf("issuance adopted changed Finding bytes: %v", err)
	}
	if !bytes.Equal(beforeState, mustRead(t, f.statePath)) || !bytes.Equal(beforeJournal, mustRead(t, f.journalPath)) {
		t.Fatal("failed source preparation changed Runtime/journal")
	}
	writeInvestigationBytes(t, f.root, "observation.txt", content)
	packets, err := investigation.IssueHypothesisResultPackets(f.root, f.statePath, f.journalPath, -1, owner.ownerID)
	if err != nil || len(packets) != 1 {
		t.Fatalf("original typed Finding source issue: %d %v", len(packets), err)
	}
	p := packets[0]
	count := 0
	for _, source := range p.Inputs.EvidenceRefs {
		if source.Origin == "finding" {
			if source.Attestation != "finding_source" || source.SourceFindingID != "finding-1" || source.Status != "" || source.Kind != "" || source.BaselineGeneration != nil || source.ReviewRound != nil {
				t.Fatalf("Finding source gained fabricated index metadata: %+v", source)
			}
			count++
		}
	}
	if count != len(refs) {
		t.Fatalf("lost original Finding sources: %d", count)
	}
	answerPath := writeS8ProfessionalAnswer(t, f.root, p, "SIMULATION: the original frozen observation supports this discriminator.")
	var answer map[string]any
	if err := json.Unmarshal(mustRead(t, filepath.Join(f.root, answerPath)), &answer); err != nil {
		t.Fatal(err)
	}
	answer["answer"].(map[string]any)["evidence_refs"] = []string{refs[0]}
	writeInvestigationJSON(t, filepath.Join(f.root, answerPath), answer)
	beforeState, beforeJournal = mustRead(t, f.statePath), mustRead(t, f.journalPath)
	writeInvestigationBytes(t, f.root, "observation.txt", []byte("changed after issue\n"))
	if _, err := investigation.SubmitIssuedHypothesisResult(f.root, f.statePath, f.journalPath, investigation.IssuedHypothesisResultRequest{ExpectedRevision: -1, PacketID: p.ID, OwnerAgentID: owner.ownerID, AnswerPath: answerPath}); err == nil {
		t.Fatal("accepted a changed frozen Finding source")
	}
	if !bytes.Equal(beforeState, mustRead(t, f.statePath)) || !bytes.Equal(beforeJournal, mustRead(t, f.journalPath)) {
		t.Fatal("failed source consumption changed Runtime/journal")
	}
	writeInvestigationBytes(t, f.root, "observation.txt", content)
	body, _ := json.Marshal(map[string]any{"hook_event_name": "SubagentStop", "agent_id": "opaque-native-finding-source", "cwd": f.root, "last_assistant_message": filepath.Join(f.root, answerPath)})
	var output, diagnostic bytes.Buffer
	code := cli.Run([]string{"hook", "--root", f.root, "--event", "SubagentStop"}, bytes.NewReader(body), &output, &diagnostic)
	if code != 0 || !strings.Contains(output.String()+diagnostic.String(), "original consumer") {
		t.Fatalf("native Stop failed: %d %s %s", code, output.String(), diagnostic.String())
	}
	state := currentS8Snapshot(t, f)
	if _, accepted, err := investigation.IssuedHypothesisResultReceipt(f.root, state.State, p.ID, owner.ownerID); err != nil || !accepted {
		t.Fatalf("original receipt: accepted=%t err=%v", accepted, err)
	}
	for _, raw := range state.State["evidence"].([]any) {
		row := raw.(map[string]any)
		for _, ref := range refs {
			if row["id"] == ref {
				t.Fatal("consumer created a Finding source index alias")
			}
		}
	}
	// Unrelated evidence cannot become a Finding source simply by using the
	// same prefix. A new Hypothesis retains its original index/anchor rules.
	pointer := investigationPointerFromState(t, state.State)
	_, err = investigation.RegisterHypothesis(f.root, f.statePath, f.journalPath, investigation.HypothesisRequest{ExpectedRevision: -1, ExpectedCaseRevision: s8TestInt(pointer["revision"]), ExpectedCaseSHA256: pointer["sha256"].(string), CaseID: pointer["case_id"].(string), HypothesisID: "hypothesis-unbound-path", AssignmentID: "assignment-unbound-path", Statement: "a different question", Invariant: "the value survives", Discriminator: "compare a new observation", ExpectedOutcomes: map[string]any{"support": "lost", "refute": "survives"}, SourceFindingIDs: []string{"finding-1"}, EvidenceRefs: []string{refs[0]}})
	if err == nil {
		t.Fatal("new Hypothesis silently accepted an unindexed path")
	}
}
