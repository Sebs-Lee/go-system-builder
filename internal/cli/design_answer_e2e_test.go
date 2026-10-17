package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/entroforge/go-system-builder/internal/cli"
	"github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/submission"
	"github.com/entroforge/go-system-builder/internal/workpacket"
	fixtures "github.com/entroforge/go-system-builder/tests/fixtures/req039"
)

// Only the authorized planning baseline and authored documents are seeded.
// Issuance, reception, transition and registration use original production code.
func newDesignFixture(t *testing.T, impact string) (string, workpacket.DesignPacket) {
	t.Helper()
	root := fixtures.FreshRoot(t)
	state := fixtures.BaseState(t, root, "planning", "design", 0)
	state["review"].(map[string]any)["round"] = 0
	state["schema_version"] = "1.2.0"
	state["configuration"].(map[string]any)["optimization"] = map[string]any{"packet": true}
	state["bound_req"].(map[string]any)["metadata"].(map[string]any)["ui_impact"] = impact
	entities := state["entities"].(map[string]any)
	entities["work_packets"], entities["submissions"] = []any{}, []any{}
	bound := state["bound_req"].(map[string]any)
	state["documents"] = []any{map[string]any{"id": bound["id"], "kind": "req", "path": bound["path"], "version": bound["version"], "sha256": bound["sha256"], "status": "locked", "generation": 1}}
	arch := filepath.Join(root, "docs/architecture/ARCHITECTURE-039.md")
	if err := os.MkdirAll(filepath.Dir(arch), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(arch, []byte("# ARCHITECTURE-039\n> Status: locked\n> Version: v1.0.0\n\nAuthored design fixture.\n"), 0600); err != nil {
		t.Fatal(err)
	}
	fixtures.WriteState(t, root, state)
	var p workpacket.DesignPacket
	if err := json.Unmarshal([]byte(stageCLI(t, "runtime", "stage-answer", "issue", "--root", root)), &p); err != nil {
		t.Fatal(err)
	}
	return root, p
}

func designDraft(p workpacket.DesignPacket) workpacket.DesignDraft {
	a := p.Draft
	conclusion, path := "pass", "docs/architecture/ARCHITECTURE-039.md"
	a.Answer = workpacket.DesignAnswer{Conclusion: &conclusion, ArchitecturePath: &path}
	return a
}

func assertDesignCheckpoint(t *testing.T, root string) {
	t.Helper()
	s := stageSnapshot(t, root)
	journal, err := os.ReadFile(filepath.Join(root, ".claude/loop-events.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if err := runtime.ValidateCheckpointIndices(root, s.State, journal); err != nil {
		t.Fatal(err)
	}
}

func TestDesignAnswerMainSaveOriginalTransitionAndReplay(t *testing.T) {
	root, p := newDesignFixture(t, "none")
	assertDesignCheckpoint(t, root)
	stageJSON(t, root, p.AnswerPath, designDraft(p))
	body, _ := json.Marshal(map[string]any{"hook_event_name": "PostToolUse", "tool_name": "Write", "tool_input": map[string]any{"file_path": filepath.Join(root, p.AnswerPath)}, "cwd": root})
	var out, errout bytes.Buffer
	if code := cli.Run([]string{"hook", "--root", root, "--event", "PostToolUse"}, bytes.NewReader(body), &out, &errout); code != 0 {
		t.Fatalf("Hook %d %s %s", code, out.String(), errout.String())
	}
	if !strings.Contains(out.String(), "S2 SUBMISSION RECEIPT") {
		t.Fatalf("missing S2 receipt: %s", out.String())
	}
	s := stageSnapshot(t, root)
	record, found, err := runtime.FindSubmission(s.State, p.ID)
	if err != nil || !found || record.ControlStatus != "advanced" {
		t.Fatalf("original reception: %+v %v %s", record, err, out.String())
	}
	life := s.State["lifecycle"].(map[string]any)
	if life["state"] != "planning" || life["phase"] != "contracts" {
		t.Fatalf("original route: %+v", life)
	}
	registered := false
	for _, raw := range s.State["documents"].([]any) {
		d := raw.(map[string]any)
		if d["path"] == "docs/architecture/ARCHITECTURE-039.md" {
			registered = true
		}
	}
	if !registered {
		t.Fatal("original action did not register design")
	}
	assertDesignCheckpoint(t, root)
	var receipt submission.Receipt
	if err := json.Unmarshal([]byte(stageCLI(t, "runtime", "stage-answer", "consume", "--root", root, "--agent", "Main", "--path", p.AnswerPath)), &receipt); err != nil {
		t.Fatal(err)
	}
	if !receipt.Accepted || !receipt.WorkComplete {
		t.Fatalf("replay receipt: %+v", receipt)
	}
	if stageSnapshot(t, root).State["revision"] != s.State["revision"] {
		t.Fatal("same answer replay changed Runtime")
	}
	assertDesignCheckpoint(t, root)
	journal, err := os.ReadFile(filepath.Join(root, ".claude/loop-events.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	for _, mutation := range []string{"wrong_owner", "wrong_digest", "missing_issuance", "wrong_transition"} {
		t.Run(mutation, func(t *testing.T) {
			state := stageSnapshot(t, root).State
			entities := state["entities"].(map[string]any)
			row := entities["submissions"].([]any)[0].(map[string]any)
			modifiedJournal := journal
			switch mutation {
			case "wrong_owner":
				row["owner"] = "other"
			case "wrong_digest":
				row["content_sha256"] = strings.Repeat("0", 64)
			case "missing_issuance":
				modifiedJournal = bytes.ReplaceAll(journal, []byte("PACKET-ISSUE"), []byte("OTHER-ISSUE"))
			case "wrong_transition":
				modifiedJournal = bytes.ReplaceAll(journal, []byte("PTR-PLAN-01"), []byte("PTR-PLAN-02"))
			}
			if err := runtime.ValidateCheckpointIndices(root, state, modifiedJournal); err == nil {
				t.Fatal("forged checkpoint accepted")
			}
		})
	}
}

func TestDesignAnswerRejectsIncompleteDriftAndUnknownUI(t *testing.T) {
	for _, mode := range []string{"incomplete", "requirement_drift", "unknown_ui"} {
		t.Run(mode, func(t *testing.T) {
			impact := "none"
			if mode == "unknown_ui" {
				impact = "unknown"
			}
			root, p := newDesignFixture(t, impact)
			if mode != "incomplete" {
				stageJSON(t, root, p.AnswerPath, designDraft(p))
			}
			if mode == "requirement_drift" {
				if err := os.WriteFile(filepath.Join(root, p.Binding.Requirement.Path), []byte("drift"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			s, err := submission.New(root)
			if err != nil {
				t.Fatal(err)
			}
			before := stageSnapshot(t, root)
			if r, err := s.SaveDesign(context.Background(), p.AnswerPath); err == nil || r.Accepted {
				t.Fatalf("rejected answer accepted: %+v %v", r, err)
			}
			after := stageSnapshot(t, root)
			if before.State["revision"] != after.State["revision"] || len(after.State["evidence"].([]any)) != 0 {
				t.Fatal("rejected design occupied domain reception")
			}
		})
	}
}

func TestDesignAnswerPendingRecoveryPreservesReception(t *testing.T) {
	root, p := newDesignFixture(t, "none")
	stageJSON(t, root, p.AnswerPath, designDraft(p))
	s, err := submission.New(root)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.Validator = cancelValidatedReception{CandidateValidator: s.Validator, packetID: p.ID, cancel: cancel}
	r, err := s.SaveDesign(ctx, p.AnswerPath)
	if err != nil || !r.Accepted || !r.WorkComplete || len(r.WaitingOn) == 0 {
		t.Fatalf("pending receipt: %+v %v", r, err)
	}
	before := stageSnapshot(t, root)
	prior, found, err := runtime.FindSubmission(before.State, p.ID)
	if err != nil || !found || prior.ControlStatus != "pending" {
		t.Fatalf("pending record: %+v %v", prior, err)
	}
	assertDesignCheckpoint(t, root)
	stageCLI(t, "runtime", "stage-answer", "recover", "--root", root, "--agent", "Main")
	after := stageSnapshot(t, root)
	next, _, err := runtime.FindSubmission(after.State, p.ID)
	if err != nil || next.ControlStatus != "advanced" || next.Key != prior.Key || next.ContentHash != prior.ContentHash || next.EvidenceID != prior.EvidenceID || len(after.State["evidence"].([]any)) != len(before.State["evidence"].([]any)) {
		t.Fatalf("recovery changed domain reception: %+v %v", next, err)
	}
	assertDesignCheckpoint(t, root)
	revision := after.State["revision"]
	stageCLI(t, "runtime", "stage-answer", "recover", "--root", root, "--agent", "Main")
	if stageSnapshot(t, root).State["revision"] != revision {
		t.Fatal("recovery replay changed Runtime")
	}
	changed := designDraft(p)
	other := "docs/architecture/ARCHITECTURE-other.md"
	changed.Answer.ArchitecturePath = &other
	stageJSON(t, root, p.AnswerPath, changed)
	fresh, err := submission.New(root)
	if err != nil {
		t.Fatal(err)
	}
	if r, err := fresh.SaveDesign(context.Background(), p.AnswerPath); err == nil || r.Accepted {
		t.Fatalf("different professional answer replay accepted: %+v %v", r, err)
	}
	if stageSnapshot(t, root).State["revision"] != revision {
		t.Fatal("conflicting replay changed Runtime")
	}
}

func TestDesignAnswerOriginalPairRecovery(t *testing.T) {
	root, p := newDesignFixture(t, "none")
	stageJSON(t, root, p.AnswerPath, designDraft(p))
	stageCLI(t, "runtime", "stage-answer", "consume", "--root", root, "--agent", "Main", "--path", p.AnswerPath)
	before := stageSnapshot(t, root)
	path, pair := latestStageCheckpoint(t, root)
	if !reflect.DeepEqual(pair.State, before.State) {
		t.Fatal("checkpoint dropped S2 facts")
	}
	damageStagePair(t, root)
	plan := checkpointPlan(t, root, path, pair.State)
	stageCLI(t, "runtime", "recover", "apply", "--root", root, "--plan", plan, "--approved-by", "operator")
	if !reflect.DeepEqual(stageSnapshot(t, root).State, before.State) {
		t.Fatal("original pair restoration changed S2 reception or cursor")
	}
	stageCLI(t, "runtime", "stage-answer", "consume", "--root", root, "--agent", "Main", "--path", p.AnswerPath)
	if stageSnapshot(t, root).State["revision"] != before.State["revision"] {
		t.Fatal("restored replay repeated original transition")
	}
	assertDesignCheckpoint(t, root)
}

func TestContractAuthorityUsesOriginalS2RegisteredDesign(t *testing.T) {
	root, design := newDesignFixture(t, "none")
	stageJSON(t, root, design.AnswerPath, designDraft(design))
	stageCLI(t, "runtime", "stage-answer", "consume", "--root", root, "--agent", "Main", "--path", design.AnswerPath)
	snap := stageSnapshot(t, root)
	binding, refs, err := runtime.CurrentContractBinding(snap.State)
	if err != nil || binding.Stage != "S3" || len(refs) != 1 {
		t.Fatalf("original S2 registration: %+v %+v %v", binding, refs, err)
	}
	p, err := workpacket.BuildContracts(binding, workpacket.ContractInputs{Designs: refs})
	if err != nil {
		t.Fatal(err)
	}
	service, err := submission.New(root)
	if err != nil {
		t.Fatal(err)
	}
	files := service.Files
	if err := runtime.ValidateContractPacket(root, snap.State, p, files); err != nil {
		t.Fatal(err)
	}
	if err := runtime.ValidateContractPacket(t.TempDir(), snap.State, p, files); err == nil {
		t.Fatal("other root granted S3 authority")
	}
	if err := os.WriteFile(filepath.Join(root, refs[0].Path), []byte("changed design"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := runtime.ValidateContractPacket(root, snap.State, p, files); err == nil {
		t.Fatal("changed design reused S3 frozen input")
	}
	snap.State["lifecycle"].(map[string]any)["phase"] = "tasks"
	if _, _, err := runtime.CurrentContractBinding(snap.State); err == nil {
		t.Fatal("S3 Main fabricated outside contracts")
	}
}
