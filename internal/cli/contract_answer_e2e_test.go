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
	"github.com/entroforge/go-system-builder/internal/qualitygate"
	"github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/submission"
	"github.com/entroforge/go-system-builder/internal/transition"
	"github.com/entroforge/go-system-builder/internal/workpacket"
	fixtures "github.com/entroforge/go-system-builder/tests/fixtures/req039"
)

func newContractAnswerFixture(t *testing.T, configure ...func(string)) (string, workpacket.ContractPacket) {
	t.Helper()
	root, design := newDesignFixture(t, "none")
	stageJSON(t, root, design.AnswerPath, designDraft(design))
	stageCLI(t, "runtime", "stage-answer", "consume", "--root", root, "--agent", "Main", "--path", design.AnswerPath)
	settleMainPlanningFollowups(t, root)
	writeContractSource(t, root, "docs/design/README.md", "# Design sources\n")
	writeContractSource(t, root, "docs/dev/contracts/CONTRACTS-039.md", "# Contract batch\n> 需求：REQ-039\n> 状态：locked\n> 版本：v1.0.0\n> Shared model policy: none\n> Shared model reason: This fixture delivers no shared wire model.\n\n## 合同清单\n\n| 合同ID | 类型 | 负责人 | 状态 | 版本 | 联调点 |\n|:---|:---|:---|:---|:---|:---|\n| BE-039 | 后端 | Main | locked | v1.0.0 | N/A |\n\n## 需求覆盖矩阵\n\n| REQ source_ref | FE 合同条款 | BE 合同条款 | SYNC 条款 |\n|:---|:---|:---|:---|\n")
	writeContractSource(t, root, "docs/dev/contracts/BE-039.md", "# Backend contract\n> 状态：locked\n> 版本：v1.0.0\n\n## Contract clauses\n\n| 本合同条款 | 条款 |\n|:---|:---|\n| §1 | Fixture service behavior |\n")
	for _, setup := range configure {
		setup(root)
	}
	// Only authored artifacts are committed in this temporary fixture.
	// The original source declarations determine which bytes consumers read.
	fixtures.CommitCurrentFixture(t, root)
	var p workpacket.ContractPacket
	if err := json.Unmarshal([]byte(stageCLI(t, "runtime", "stage-answer", "issue", "--root", root)), &p); err != nil {
		t.Fatal(err)
	}
	return root, p
}

func settleMainPlanningFollowups(t *testing.T, root string) {
	t.Helper()
	snap := stageSnapshot(t, root)
	rows, _ := snap.State["entities"].(map[string]any)["submissions"].([]any)
	for _, raw := range rows {
		row, _ := raw.(map[string]any)
		if row["owner"] == "Main" && row["control_status"] == "pending" {
			s, err := submission.New(root)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := s.Recover(context.Background(), "Main", 2); err != nil {
				t.Fatal(err)
			}
			return
		}
	}
}

func writeContractSource(t *testing.T, root, path, content string) {
	t.Helper()
	full := filepath.Join(root, path)
	if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}

func completedContractDraft(p workpacket.ContractPacket) workpacket.ContractDraft {
	pass, index := "pass", "docs/dev/contracts/CONTRACTS-039.md"
	a := p.Draft
	a.Answer = workpacket.ContractAnswer{Conclusion: &pass, IndexPath: &index, ContractPaths: []string{"docs/dev/contracts/BE-039.md"}, Applicability: map[string]workpacket.ContractApplicability{"FE": {Disposition: "not_applicable", Rationale: "The fixture requirement has no client."}, "BE": {Disposition: "required", Rationale: "This batch specifies the service."}, "SYNC": {Disposition: "not_applicable", Rationale: "There is no separate client/provider boundary in this fixture."}}}
	return a
}

func TestContractAnswerMainHookOriginalRouteAndHistoricalReceipt(t *testing.T) {
	root, p := newContractAnswerFixture(t)
	assertDesignCheckpoint(t, root)
	stageJSON(t, root, p.AnswerPath, completedContractDraft(p))
	body, _ := json.Marshal(map[string]any{"hook_event_name": "PostToolUse", "tool_name": "Write", "tool_input": map[string]any{"file_path": filepath.Join(root, p.AnswerPath)}, "cwd": root})
	var out, errout bytes.Buffer
	if code := cli.Run([]string{"hook", "--root", root, "--event", "PostToolUse"}, bytes.NewReader(body), &out, &errout); code != 0 {
		t.Fatalf("S3 Hook %d %s %s", code, out.String(), errout.String())
	}
	if !strings.Contains(out.String(), "S3 SUBMISSION RECEIPT") {
		t.Fatal(out.String())
	}
	snap := stageSnapshot(t, root)
	record, found, err := runtime.FindSubmission(snap.State, p.ID)
	if err != nil || !found || record.ControlStatus != "advanced" {
		service, _ := submission.New(root)
		if service != nil {
			catalog, _ := transition.LoadCatalog(root)
			registry, _ := qualitygate.NewRegistry(catalog)
			gate, gateErr := qualitygate.NewEvaluator(registry).Evaluate(context.Background(), qualitygate.Input{Root: root, Snapshot: snap, TransitionID: "PTR-PLAN-02", GateID: "GATE-PLANNING-CONTRACTS-COMPLETE", Files: service.Files})
			t.Logf("original S3 gate: %+v %v", gate, gateErr)
		}
		t.Fatalf("original S3 reception: %+v %v %s", record, err, out.String())
	}
	life := snap.State["lifecycle"].(map[string]any)
	if life["state"] != "planning" || life["phase"] != "tasks" {
		t.Fatalf("original S3 route: %+v", life)
	}
	registered := map[string]bool{}
	for _, raw := range snap.State["documents"].([]any) {
		row := raw.(map[string]any)
		if row["kind"] == "contract" {
			registered[row["path"].(string)] = true
		}
	}
	if !registered["docs/dev/contracts/CONTRACTS-039.md"] || !registered["docs/dev/contracts/BE-039.md"] {
		t.Fatal("original PTR did not register complete batch", registered)
	}
	assertDesignCheckpoint(t, root)
	stageCLI(t, "runtime", "stage-answer", "consume", "--root", root, "--agent", "Main", "--path", p.AnswerPath)
	if stageSnapshot(t, root).State["revision"] != snap.State["revision"] {
		t.Fatal("replay changed original Runtime")
	}
	// Historical receipt must reconstruct the original professional facts;
	// current authoring bytes can no longer dictate its applicability/hash.
	writeContractSource(t, root, p.AnswerPath, "corrupt mutable answer\n")
	var receipt submission.Receipt
	if err := json.Unmarshal([]byte(stageCLI(t, "runtime", "stage-answer", "receipt", "--root", root, "--agent", "Main", "--packet", p.ID)), &receipt); err != nil {
		t.Fatal(err)
	}
	if !receipt.Accepted || !receipt.WorkComplete {
		t.Fatalf("original historical receipt: %+v", receipt)
	}
}

func TestContractAnswerPendingAndOriginalPairRecovery(t *testing.T) {
	root, p := newContractAnswerFixture(t)
	stageJSON(t, root, p.AnswerPath, completedContractDraft(p))
	s, err := submission.New(root)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.Validator = cancelValidatedReception{CandidateValidator: s.Validator, packetID: p.ID, cancel: cancel}
	r, err := s.SaveContracts(ctx, p.AnswerPath)
	if err != nil || !r.Accepted || !r.WorkComplete || len(r.WaitingOn) == 0 {
		t.Fatalf("pending S3: %+v %v", r, err)
	}
	before := stageSnapshot(t, root)
	old, found, err := runtime.FindSubmission(before.State, p.ID)
	if err != nil || !found || old.ControlStatus != "pending" {
		t.Fatalf("pending original reception: %+v %v", old, err)
	}
	path, pair := latestStageCheckpoint(t, root)
	for _, mutation := range []string{"replaced_without_decision", "wrong_owner", "wrong_professional_digest"} {
		t.Run(mutation, func(t *testing.T) {
			state := stageSnapshot(t, root).State
			rows := state["entities"].(map[string]any)["submissions"].([]any)
			for _, raw := range rows {
				row := raw.(map[string]any)
				if row["packet_id"] != p.ID {
					continue
				}
				switch mutation {
				case "replaced_without_decision":
					row["control_status"] = "replaced"
				case "wrong_owner":
					row["owner"] = "other"
				case "wrong_professional_digest":
					row["content_sha256"] = strings.Repeat("0", 64)
				}
			}
			if err := runtime.ValidateCheckpointIndices(root, state, pair.JournalBytes); err == nil {
				t.Fatal("forged S3 checkpoint accepted")
			}
		})
	}
	damageStagePair(t, root)
	plan := checkpointPlan(t, root, path, pair.State)
	stageCLI(t, "runtime", "recover", "apply", "--root", root, "--plan", plan, "--approved-by", "operator")
	if !reflect.DeepEqual(stageSnapshot(t, root).State, before.State) {
		t.Fatal("S2/S3 original pair facts changed during restore")
	}
	stageCLI(t, "runtime", "stage-answer", "recover", "--root", root, "--agent", "Main")
	after := stageSnapshot(t, root)
	next, _, err := runtime.FindSubmission(after.State, p.ID)
	if err != nil || next.ControlStatus != "advanced" || old.Key != next.Key || old.EvidenceID != next.EvidenceID || old.ContentHash != next.ContentHash || len(after.State["evidence"].([]any)) != len(before.State["evidence"].([]any)) {
		t.Fatalf("recovery changed professional reception: %+v %v", next, err)
	}
	assertDesignCheckpoint(t, root)
	stageCLI(t, "runtime", "stage-answer", "recover", "--root", root, "--agent", "Main")
	if stageSnapshot(t, root).State["revision"] != after.State["revision"] {
		t.Fatal("recovered followup replay changed Runtime")
	}
}

func TestContractAnswerConditionalADRPassesOriginalGate(t *testing.T) {
	root, p := newContractAnswerFixture(t, func(root string) {
		writeContractSource(t, root, "docs/design/decisions/ADR-039.md", "# Decision\n> 状态：accepted\n> 版本：v1.0.0\n> 关联需求：REQ-039\n\nAn explicitly accepted decision for this fixture.\n")
	})
	if len(p.Inputs.Sources) != 1 || p.Inputs.Sources[0].Path != "docs/design/decisions/ADR-039.md" {
		t.Fatalf("accepted ADR omitted: %+v", p.Inputs)
	}
	stageJSON(t, root, p.AnswerPath, completedContractDraft(p))
	stageCLI(t, "runtime", "stage-answer", "consume", "--root", root, "--agent", "Main", "--path", p.AnswerPath)
	snap := stageSnapshot(t, root)
	record, found, err := runtime.FindSubmission(snap.State, p.ID)
	if err != nil || !found || record.ControlStatus != "advanced" {
		t.Fatalf("S3 conditional original gate: %+v %v", record, err)
	}
	assertDesignCheckpoint(t, root)
}

func TestContractAnswerRejectsDraftAndSourceDriftBeforeReception(t *testing.T) {
	for _, mode := range []string{"draft", "unknown_contract", "source_drift"} {
		t.Run(mode, func(t *testing.T) {
			root, p := newContractAnswerFixture(t)
			stageJSON(t, root, p.AnswerPath, completedContractDraft(p))
			switch mode {
			case "draft":
				writeContractSource(t, root, "docs/dev/contracts/BE-039.md", "# Draft\n> Status: draft\n> Version: v1.0.0\n")
			case "unknown_contract":
				writeContractSource(t, root, "docs/dev/contracts/FE-039.md", "# Extra draft\n> Status: draft\n> Version: v1.0.0\n")
			case "source_drift":
				writeContractSource(t, root, "docs/design/prototypes/new-module/scenario-model.json", "{}\n")
			}
			before := stageSnapshot(t, root)
			s, err := submission.New(root)
			if err != nil {
				t.Fatal(err)
			}
			if r, err := s.SaveContracts(context.Background(), p.AnswerPath); err == nil || r.Accepted {
				t.Fatalf("invalid S3 accepted: %+v %v", r, err)
			}
			after := stageSnapshot(t, root)
			if before.Revision != after.Revision || len(before.State["evidence"].([]any)) != len(after.State["evidence"].([]any)) {
				t.Fatal("rejected S3 changed original reception")
			}
		})
	}
}
