package runtime_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/entroforge/go-system-builder/internal/fileview"
	"github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/workpacket"
)

func packetFixture(t *testing.T) (string, string, string, workpacket.Packet) {
	t.Helper()
	root := t.TempDir()
	statePath := filepath.Join(root, "loop-state.json")
	journalPath := filepath.Join(root, "loop-events.jsonl")
	writeExampleRuntime(t, statePath)
	data, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	var state map[string]any
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatal(err)
	}
	state["schema_version"] = "1.2.0"
	state["lifecycle"] = map[string]any{"state": "document_verification", "phase": nil, "phase_revision": 7}
	state["configuration"].(map[string]any)["optimization"] = map[string]any{"packet": true}
	manifest := []byte(`{"assignments":[{"assignment_id":"assignment-spec","agent_id":"reviewer-spec","responsibility_id":"DV-SPEC-CONSISTENCY","role_family":"document-verifier"}]}`)
	if err := os.WriteFile(filepath.Join(root, "manifest.json"), manifest, 0600); err != nil {
		t.Fatal(err)
	}
	input := []byte("# REQ\n> Status: locked\n")
	if err := os.WriteFile(filepath.Join(root, "req.md"), input, 0600); err != nil {
		t.Fatal(err)
	}
	ref := workpacket.Reference{ID: "REQ-039", Path: "req.md", Version: "v1.0.0", SHA256: workpacket.Digest(input)}
	state["documents"] = []any{map[string]any{"id": ref.ID, "path": ref.Path, "kind": "req", "version": ref.Version, "sha256": ref.SHA256, "status": "registered", "generation": 1}}
	entities := state["entities"].(map[string]any)
	entities["agents"] = []any{map[string]any{"id": "reviewer-spec", "role": "document-verifier", "state": "working", "task_ids": []any{}, "team_id": "team-dv", "definition_ref": "agents/document-verifier.md", "prompt_ref": "manifest.json#assignment-spec", "readback_ref": nil, "activation_ref": nil, "activation_revision": nil, "updated_at": "2026-10-13T00:00:00Z"}}
	entities["teams"] = []any{map[string]any{"id": "team-dv", "platform_team_id": "platform-dv", "kind": "document_verifier", "status": "active", "manifest_ref": "manifest.json", "responsibility_ids": []any{"DV-SPEC-CONSISTENCY"}, "agent_ids": []any{"reviewer-spec"}, "review_round": nil}}
	data, err = json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(statePath, data, 0600); err != nil {
		t.Fatal(err)
	}
	p, err := workpacket.BuildS5(workpacket.Binding{RuntimeID: state["runtime_id"].(string), Generation: 1, StageEntry: 7, Owner: "reviewer-spec", AssignmentID: "assignment-spec", AssignmentSHA: workpacket.Digest(manifest), Responsibility: "DV-SPEC-CONSISTENCY", Attempt: 1}, []workpacket.Reference{ref})
	if err != nil {
		t.Fatal(err)
	}
	return root, statePath, journalPath, p
}

type simultaneousPacketReads struct {
	fileview.Reader
	reads atomic.Int32
	ready chan struct{}
}

func (r *simultaneousPacketReads) ReadFile(path string) ([]byte, error) {
	n := r.reads.Add(1)
	if n == 2 {
		close(r.ready)
	}
	if n <= 2 {
		<-r.ready
	}
	return r.Reader.ReadFile(path)
}

func TestConcurrentDocumentReceptionIsUnique(t *testing.T) {
	for _, different := range []bool{false, true} {
		t.Run(map[bool]string{false: "identical", true: "content conflict"}[different], func(t *testing.T) {
			root, statePath, journalPath, p := packetFixture(t)
			if _, err := runtime.RecordWorkPacket(root, statePath, journalPath, p, fileview.Disk{Root: root}, testCandidateValidator()); err != nil {
				t.Fatal(err)
			}
			envelope := map[string]any{"schema_version": "1.0.0", "evidence_id": "ev-answer", "kind": "document_review", "runtime_id": p.Binding.RuntimeID, "baseline_generation": 1, "producer_agent_id": p.Binding.Owner, "producer_responsibility": p.Binding.Responsibility, "subject_refs": p.Inputs, "conclusion": "pass"}
			data, err := json.Marshal(envelope)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "review.json"), data, 0600); err != nil {
				t.Fatal(err)
			}
			files := &simultaneousPacketReads{Reader: fileview.Disk{Root: root}, ready: make(chan struct{})}
			errs := make([]error, 2)
			var group sync.WaitGroup
			for i := range 2 {
				group.Add(1)
				go func(i int) {
					defer group.Done()
					content := "answer"
					if different && i == 1 {
						content = "different"
					}
					_, errs[i] = runtime.RecordEvidence(root, statePath, journalPath, runtime.EvidenceRequest{ExpectedRevision: -1, ID: "ev-answer", Kind: "document_review", Path: "review.json", ProducedBy: []string{p.Binding.Owner}, ResponsibilityID: p.Binding.Responsibility, Validator: testCandidateValidator(), Submission: &runtime.SubmissionRequest{PacketID: p.ID, Owner: p.Binding.Owner, ContentHash: workpacket.Digest([]byte(content)), Files: files}})
				}(i)
			}
			group.Wait()
			accepted, conflicts := 0, 0
			for _, err := range errs {
				if err == nil {
					accepted++
				} else if errors.Is(err, runtime.ErrSubmissionConflict) {
					conflicts++
				} else {
					t.Fatalf("unexpected concurrent result: %v", err)
				}
			}
			if (!different && accepted != 2) || (different && (accepted != 1 || conflicts != 1)) {
				t.Fatalf("accepted=%d conflicts=%d", accepted, conflicts)
			}
			snapshot, err := runtime.NewStore(statePath, journalPath).Snapshot()
			if err != nil {
				t.Fatal(err)
			}
			if len(snapshot.State["evidence"].([]any)) != 1 || len(snapshot.State["entities"].(map[string]any)["submissions"].([]any)) != 1 {
				t.Fatal("duplicate domain commit")
			}
		})
	}
}

func TestWorkPacketIssuanceIsDurableAndIdempotent(t *testing.T) {
	root, statePath, journalPath, p := packetFixture(t)
	first, err := runtime.RecordWorkPacket(root, statePath, journalPath, p, fileview.Disk{Root: root}, testCandidateValidator())
	if err != nil {
		t.Fatal(err)
	}
	second, err := runtime.RecordWorkPacket(root, statePath, journalPath, p, fileview.Disk{Root: root}, testCandidateValidator())
	if err != nil {
		t.Fatal(err)
	}
	if second.Revision != first.Revision {
		t.Fatal("replay appended a new issuance")
	}
	if len(second.State["entities"].(map[string]any)["work_packets"].([]any)) != 1 {
		t.Fatal("duplicate issuance")
	}
}

func TestWorkPacketRejectsUnregisteredOwnerAndInputDrift(t *testing.T) {
	for _, mutation := range []string{"owner", "stage", "input", "manifest"} {
		t.Run(mutation, func(t *testing.T) {
			root, statePath, journalPath, p := packetFixture(t)
			switch mutation {
			case "owner":
				p.Binding.Owner = "other"
			case "stage":
				p.Binding.StageEntry++
			case "input":
				if err := os.WriteFile(filepath.Join(root, "req.md"), []byte("changed"), 0600); err != nil {
					t.Fatal(err)
				}
			case "manifest":
				if err := os.WriteFile(filepath.Join(root, "manifest.json"), []byte(`{}`), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if mutation == "owner" || mutation == "stage" {
				var err error
				p, err = workpacket.BuildS5(p.Binding, p.Inputs)
				if err != nil {
					t.Fatal(err)
				}
			}
			before, err := os.ReadFile(statePath)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := runtime.RecordWorkPacket(root, statePath, journalPath, p, fileview.Disk{Root: root}, testCandidateValidator()); err == nil {
				t.Fatal("invalid packet accepted")
			}
			after, err := os.ReadFile(statePath)
			if err != nil {
				t.Fatal(err)
			}
			if string(before) != string(after) {
				t.Fatal("rejected issuance changed runtime")
			}
		})
	}
}

func TestWorkPacketCannotIssueAnotherAttemptWithoutRevision(t *testing.T) {
	root, statePath, journalPath, p := packetFixture(t)
	if _, err := runtime.RecordWorkPacket(root, statePath, journalPath, p, fileview.Disk{Root: root}, testCandidateValidator()); err != nil {
		t.Fatal(err)
	}
	p.Binding.Attempt++
	p, err := workpacket.BuildS5(p.Binding, p.Inputs)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runtime.RecordWorkPacket(root, statePath, journalPath, p, fileview.Disk{Root: root}, testCandidateValidator()); err == nil {
		t.Fatal("new attempt issued without a trusted revision event")
	}
}

func TestDocumentAnswerReceptionUsesStableIdentity(t *testing.T) {
	root, statePath, journalPath, p := packetFixture(t)
	if _, err := runtime.RecordWorkPacket(root, statePath, journalPath, p, fileview.Disk{Root: root}, testCandidateValidator()); err != nil {
		t.Fatal(err)
	}
	e := map[string]any{"schema_version": "1.0.0", "evidence_id": "ev-answer", "kind": "document_review", "runtime_id": p.Binding.RuntimeID, "baseline_generation": 1, "producer_agent_id": p.Binding.Owner, "producer_responsibility": p.Binding.Responsibility, "subject_refs": p.Inputs, "conclusion": "pass", "created_at": "2026-10-13T00:00:00Z"}
	data, err := json.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "review.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	request := runtime.EvidenceRequest{ExpectedRevision: -1, ID: "ev-answer", Kind: "document_review", Path: "review.json", ProducedBy: []string{p.Binding.Owner}, ResponsibilityID: p.Binding.Responsibility, Validator: testCandidateValidator(), Submission: &runtime.SubmissionRequest{PacketID: p.ID, Owner: p.Binding.Owner, ContentHash: workpacket.Digest([]byte("answer")), Files: fileview.Disk{Root: root}}}
	first, err := runtime.RecordEvidence(root, statePath, journalPath, request)
	if err != nil {
		t.Fatal(err)
	}
	second, err := runtime.RecordEvidence(root, statePath, journalPath, request)
	if err != nil {
		t.Fatal(err)
	}
	if first.Revision != second.Revision {
		t.Fatal("repeated answer appended evidence")
	}
	rows := second.State["entities"].(map[string]any)["submissions"].([]any)
	if len(rows) != 1 || len(second.State["evidence"].([]any)) != 1 {
		t.Fatal("domain and reception were not one commit")
	}
	row := rows[0].(map[string]any)
	if row["control_status"] != "pending" || row["followup_event_id"] == "" {
		t.Fatal("recovery event was not persisted with domain reception")
	}
	request.Submission.ContentHash = workpacket.Digest([]byte("different answer"))
	if _, err := runtime.RecordEvidence(root, statePath, journalPath, request); !errors.Is(err, runtime.ErrSubmissionConflict) {
		t.Fatalf("same identity accepted different content: %v", err)
	}
}
