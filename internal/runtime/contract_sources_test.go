package runtime_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/entroforge/go-system-builder/internal/fileview"
	"github.com/entroforge/go-system-builder/internal/runtime"
	"github.com/entroforge/go-system-builder/internal/sharedmodel"
	"github.com/entroforge/go-system-builder/internal/workpacket"
)

func TestResolveContractInputsClosesScenarioAndLocalDesignDependencies(t *testing.T) {
	root, state := contractSourceFixture(t, true)
	inputs, err := runtime.ResolveContractInputs(root, state, fileview.Disk{Root: root})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"docs/design/decisions/ADR-001.md",
		"docs/design/decisions/ADR-002.md",
		"docs/architecture/data-model/model.md",
		"docs/architecture/data-model/order.schema.json",
		"docs/architecture/data-model/money.schema.json",
		"docs/architecture/data-model/order-example.json",
		"docs/design/prototypes/orders/scenario-model.json",
		"docs/design/prototypes/orders/stories.md",
		"docs/design/prototypes/orders/assets/banner.svg",
	}
	paths := make(map[string]bool, len(inputs.Sources))
	for _, source := range inputs.Sources {
		paths[source.Path] = true
	}
	for _, path := range want {
		if !paths[path] {
			t.Errorf("frozen conditional sources omit %s", path)
		}
	}
	if !paths["docs/design/decisions/ADR-003.md"] {
		t.Fatal("accepted ADR associated with the bound REQ was not included")
	}
	if paths["docs/design/decisions/ADR-004.md"] {
		t.Fatal("an ADR associated with a different REQ was included")
	}
	if paths["docs/dev/contracts/SYNC-001.md"] {
		t.Fatal("the S3 output contract was frozen as an input")
	}
	if _, ok := inputs.Directories["docs/design/prototypes"]; !ok {
		t.Fatal("the module discovery directory is not frozen")
	}
	if _, ok := inputs.Directories["docs/design/prototypes/orders/assets"]; !ok {
		t.Fatal("nested package membership is not frozen")
	}

	binding, _, err := runtime.CurrentContractBinding(state)
	if err != nil {
		t.Fatal(err)
	}
	packet, err := workpacket.BuildContracts(binding, inputs)
	if err != nil {
		t.Fatal(err)
	}
	writeContractSourceFile(t, root, "docs/dev/contracts/SYNC-001.md", "# Updated S3 output\n")
	writeContractSourceFile(t, root, "docs/dev/contracts/FE-002.md", "# Newly authored S3 output\n")
	if err := runtime.ValidateContractPacket(root, state, packet, fileview.Disk{Root: root}); err != nil {
		t.Fatalf("changes to the S3 output contract set affected frozen inputs: %v", err)
	}
}

func TestResolveContractInputsFreezesMissingScenarioRoot(t *testing.T) {
	root, state := contractSourceFixture(t, false)
	inputs, err := runtime.ResolveContractInputs(root, state, fileview.Disk{Root: root})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := inputs.Directories["docs/design"]; !ok {
		t.Fatal("nearest bounded parent manifest does not prove prototypes root absence")
	}
	binding, _, err := runtime.CurrentContractBinding(state)
	if err != nil {
		t.Fatal(err)
	}
	packet, err := workpacket.BuildContracts(binding, inputs)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "docs/design/prototypes"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := runtime.ValidateContractPacket(root, state, packet, fileview.Disk{Root: root}); err == nil || !strings.Contains(err.Error(), "source directory changed") {
		t.Fatalf("new prototypes root did not invalidate the frozen absence: %v", err)
	}
}

func TestResolveContractInputsDoesNotInferEmptyScenarioApplicability(t *testing.T) {
	root, state := contractSourceFixture(t, false)
	req := []byte("# REQ-001\n> Status: locked\n> Version: v1.0.0\n\n| ID | AC | Target |\n|:--|:--|:--|\n| AC-001 | Complete order | FR-001 |\n")
	writeContractSourceFile(t, root, "docs/requirements/REQ-001.md", string(req))
	state["bound_req"].(map[string]any)["sha256"] = workpacket.Digest(req)
	if _, err := runtime.ResolveContractInputs(root, state, fileview.Disk{Root: root}); err == nil || !strings.Contains(err.Error(), "no module package exists") {
		t.Fatalf("empty scenario set was accepted for an FR-targeted criterion: %v", err)
	}
}

func TestResolveContractInputsDoesNotFreezeCurrentSharedModelOutputClosure(t *testing.T) {
	root, state := contractSourceFixture(t, false)
	writeSharedModelIndex(t, root, false)

	inputs, err := runtime.ResolveContractInputs(root, state, fileview.Disk{Root: root})
	if err != nil {
		t.Fatal(err)
	}
	paths := make(map[string]bool, len(inputs.Sources))
	for _, source := range inputs.Sources {
		paths[source.Path] = true
	}
	for _, path := range []string{
		"docs/architecture/data-model/indexed-order.schema.json",
		"docs/architecture/data-model/indexed-detail.schema.json",
		"docs/architecture/data-model/indexed-valid.json",
		"docs/architecture/data-model/indexed-invalid.json",
	} {
		if paths[path] {
			t.Errorf("S3 output closure was frozen before the professional answer: %s", path)
		}
	}
	for _, output := range []string{
		"docs/dev/contracts/CONTRACTS-001.md",
		"docs/dev/contracts/FE-001.md",
	} {
		if paths[output] {
			t.Errorf("S3 contract output was frozen as an input: %s", output)
		}
	}
}

func TestResolveContractInputsFreezesUpstreamWhileS3SharedModelClosureRemainsOutput(t *testing.T) {
	root, state := contractSourceFixture(t, false)
	writeSharedModelIndex(t, root, false)
	inputs, err := runtime.ResolveContractInputs(root, state, fileview.Disk{Root: root})
	if err != nil {
		t.Fatal(err)
	}
	binding, _, err := runtime.CurrentContractBinding(state)
	if err != nil {
		t.Fatal(err)
	}
	packet, err := workpacket.BuildContracts(binding, inputs)
	if err != nil {
		t.Fatal(err)
	}

	// The S3 index and consumer contracts are writable outputs and do not
	// contribute to the packet's upstream frozen source set.
	writeContractSourceFile(t, root, "docs/dev/contracts/FE-001.md", "# Updated FE contract\n\n## Shared model inputs\n\n[Indexed order](../../architecture/data-model/indexed-order.schema.json)\n\nNew consumer explanation.\n")
	writeContractSourceFile(t, root, "docs/dev/contracts/SYNC-001.md", "# Updated S3 output contract\n")
	if err := runtime.ValidateIssuedContractPacket(root, state, packet, fileview.Disk{Root: root}); err != nil {
		t.Fatalf("editing contract outputs changed frozen inputs: %v", err)
	}

	// Adding a baseline declaration changes the closure discovered by the
	// original submission checker, not the upstream issuance.
	writeContractSourceFile(t, root, "docs/architecture/data-model/indexed-refund.schema.json", `{"type":"object","properties":{"reason":{"type":"string"}},"required":["reason"]}`)
	writeContractSourceFile(t, root, "docs/architecture/data-model/indexed-refund-valid.json", `{"reason":"duplicate"}`)
	writeContractSourceFile(t, root, "docs/architecture/data-model/indexed-refund-invalid.json", `{}`)
	writeContractSourceFile(t, root, "docs/dev/contracts/FE-001.md", "# Updated FE contract\n\n## Shared model inputs\n\n[Indexed order](../../architecture/data-model/indexed-order.schema.json)\n[Indexed refund](../../architecture/data-model/indexed-refund.schema.json)\n")
	writeSharedModelIndex(t, root, true)
	if err := runtime.ValidateIssuedContractPacket(root, state, packet, fileview.Disk{Root: root}); err != nil {
		t.Fatalf("S3 model output closure changed the upstream issuance: %v", err)
	}
}

func TestResolveContractInputsDoesNotRunCurrentSharedModelCheckerAtIssue(t *testing.T) {
	root, state := contractSourceFixture(t, false)
	writeContractSourceFile(t, root, "docs/dev/contracts/CONTRACTS-001.md", "# CONTRACTS-001\n> 需求: REQ-001\n\nNo explicit shared-model applicability decision.\n")
	if _, err := runtime.ResolveContractInputs(root, state, fileview.Disk{Root: root}); err != nil {
		t.Fatalf("current S3 outputs blocked upstream issuance: %v", err)
	}
}

func TestValidateIssuedContractPacketAcceptsLegacyFrozenSharedModelInputs(t *testing.T) {
	root, state := contractSourceFixture(t, false)
	writeSharedModelIndex(t, root, false)
	files := fileview.Disk{Root: root}
	inputs, err := runtime.ResolveContractInputs(root, state, files)
	if err != nil {
		t.Fatal(err)
	}
	checked := sharedmodel.Check(root, files, "REQ-001")
	if len(checked.Problems) != 0 || len(checked.Warnings) != 0 {
		t.Fatalf("legacy source fixture invalid: problems=%v warnings=%v", checked.Problems, checked.Warnings)
	}
	for path, data := range checked.Files {
		if strings.HasPrefix(path, "docs/dev/contracts/") {
			continue
		}
		inputs.Sources = append(inputs.Sources, workpacket.Subject{Path: path, SHA256: workpacket.Digest(data)})
	}
	binding, _, err := runtime.CurrentContractBinding(state)
	if err != nil {
		t.Fatal(err)
	}
	packet, err := workpacket.BuildContracts(binding, inputs)
	if err != nil {
		t.Fatal(err)
	}
	if err := runtime.ValidateIssuedContractPacket(root, state, packet, files); err != nil {
		t.Fatalf("valid historical S3 output closure was rejected: %v", err)
	}
	writeContractSourceFile(t, root, "docs/architecture/data-model/indexed-order.schema.json", `{"type":"object","properties":{"changed":{"type":"string"}}}`)
	if err := runtime.ValidateIssuedContractPacket(root, state, packet, files); err == nil || !strings.Contains(err.Error(), "frozen input drifted") {
		t.Fatalf("historically frozen model source drift was accepted: %v", err)
	}
}

func TestValidateIssuedContractPacketRejectsUpstreamAndTemplateDrift(t *testing.T) {
	issued := func(t *testing.T) (string, map[string]any, workpacket.ContractInputs, workpacket.PlanningBinding, workpacket.ContractPacket, fileview.Reader) {
		t.Helper()
		root, state := contractSourceFixture(t, true)
		files := fileview.Disk{Root: root}
		inputs, err := runtime.ResolveContractInputs(root, state, files)
		if err != nil {
			t.Fatal(err)
		}
		binding, _, err := runtime.CurrentContractBinding(state)
		if err != nil {
			t.Fatal(err)
		}
		packet, err := workpacket.BuildContracts(binding, inputs)
		if err != nil {
			t.Fatal(err)
		}
		return root, state, inputs, binding, packet, files
	}
	t.Run("missing frozen input", func(t *testing.T) {
		root, state, inputs, binding, _, files := issued(t)
		missing := inputs
		missing.Sources = append([]workpacket.Subject(nil), inputs.Sources[:len(inputs.Sources)-1]...)
		changed, err := workpacket.BuildContracts(binding, missing)
		if err != nil {
			t.Fatal(err)
		}
		if err := runtime.ValidateIssuedContractPacket(root, state, changed, files); err == nil || !strings.Contains(err.Error(), "source closure differs") {
			t.Fatalf("issuance missing a frozen source was accepted: %v", err)
		}
	})
	t.Run("new upstream source", func(t *testing.T) {
		root, state, inputs, _, packet, files := issued(t)
		designPath := inputs.Designs[0].Path
		changedDesign := "# Architecture\n> Status: locked\n> Version: v1.0.0\n\n[New source](data-model/new-source.md).\n"
		writeContractSourceFile(t, root, designPath, changedDesign)
		writeContractSourceFile(t, root, "docs/architecture/data-model/new-source.md", "# New source\n")
		state["documents"].([]any)[0].(map[string]any)["sha256"] = workpacket.Digest([]byte(changedDesign))
		if err := runtime.ValidateIssuedContractPacket(root, state, packet, files); err == nil {
			t.Fatal("newly discovered upstream source reused an old issuance")
		}
	})
	t.Run("template changed", func(t *testing.T) {
		root, state, _, _, packet, files := issued(t)
		changed := packet
		changed.TemplateSHA = "changed-template"
		if err := runtime.ValidateIssuedContractPacket(root, state, changed, files); err == nil {
			t.Fatal("changed S3 template reused an old issuance")
		}
	})
}

func TestCurrentContractBindingExcludesRegisteredS3SharedModelDocuments(t *testing.T) {
	root, state := contractSourceFixture(t, false)
	modelPath := "docs/architecture/data-model/registered.schema.json"
	modelData := []byte(`{"type":"object"}`)
	writeContractSourceFile(t, root, modelPath, string(modelData))
	docs := state["documents"].([]any)
	state["documents"] = append(docs, map[string]any{
		"id": "shared-model:" + modelPath, "kind": "design", "path": modelPath,
		"version": "unversioned", "sha256": workpacket.Digest(modelData), "status": "locked", "generation": 1,
	})
	_, designs, err := runtime.CurrentContractBinding(state)
	if err != nil {
		t.Fatal(err)
	}
	if len(designs) != 1 || designs[0].ID != "ARCHITECTURE-001" {
		t.Fatalf("S3 shared-model registration was mistaken for an S2 design: %#v", designs)
	}
}

func TestResolveContractInputsRejectsUnacceptedADRForBoundREQ(t *testing.T) {
	root, state := contractSourceFixture(t, false)
	writeContractSourceFile(t, root, "docs/design/decisions/ADR-005.md", "> 状态: proposed\n> 关联需求: `docs/requirements/REQ-001.md`\n")
	if _, err := runtime.ResolveContractInputs(root, state, fileview.Disk{Root: root}); err == nil || !strings.Contains(err.Error(), "still proposed") {
		t.Fatalf("unaccepted current-REQ ADR did not block source closure: %v", err)
	}
}

func TestResolveContractInputsFreezesModuleAdditionAndDeletion(t *testing.T) {
	t.Run("addition", func(t *testing.T) {
		root, state := contractSourceFixture(t, true)
		inputs, err := runtime.ResolveContractInputs(root, state, fileview.Disk{Root: root})
		if err != nil {
			t.Fatal(err)
		}
		binding, _, _ := runtime.CurrentContractBinding(state)
		packet, err := workpacket.BuildContracts(binding, inputs)
		if err != nil {
			t.Fatal(err)
		}
		writeContractSourceFile(t, root, "docs/design/prototypes/refunds/scenario-model.json", `{"module":"refunds"}`)
		if err := runtime.ValidateContractPacket(root, state, packet, fileview.Disk{Root: root}); err == nil || !strings.Contains(err.Error(), "source directory changed") {
			t.Fatalf("added module directory did not invalidate prior source set: %v", err)
		}
	})
	t.Run("deletion", func(t *testing.T) {
		root, state := contractSourceFixture(t, true)
		inputs, err := runtime.ResolveContractInputs(root, state, fileview.Disk{Root: root})
		if err != nil {
			t.Fatal(err)
		}
		binding, _, _ := runtime.CurrentContractBinding(state)
		packet, err := workpacket.BuildContracts(binding, inputs)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.RemoveAll(filepath.Join(root, "docs/design/prototypes/orders")); err != nil {
			t.Fatal(err)
		}
		if err := runtime.ValidateContractPacket(root, state, packet, fileview.Disk{Root: root}); err == nil {
			t.Fatal("deleted module package did not invalidate prior source set")
		}
	})
}

func TestResolveContractInputsRejectsFileViewCheckoutMismatch(t *testing.T) {
	root, state := contractSourceFixture(t, false)
	files := pinnedContractFiles{root: root, overrides: map[string][]byte{
		"docs/requirements/REQ-001.md": []byte("upstream REQ bytes"),
	}}
	if _, err := runtime.ResolveContractInputs(root, state, files); err == nil || !strings.Contains(err.Error(), "original checkout SHA") {
		t.Fatalf("a different upstream REQ was accepted: %v", err)
	}
}

func contractSourceFixture(t *testing.T, withModule bool) (string, map[string]any) {
	t.Helper()
	root := t.TempDir()
	reqPath := "docs/requirements/REQ-001.md"
	designPath := "docs/architecture/ARCHITECTURE-001.md"
	req := []byte("# REQ-001\n> Status: locked\n> Version: v1.0.0\n")
	design := []byte("# Architecture\n> Status: locked\n> Version: v1.0.0\n\nSee `docs/architecture/data-model/model.md`. [Decision](../design/decisions/ADR-001.md). [contract output](../dev/contracts/SYNC-001.md).\n")
	writeContractSourceFile(t, root, reqPath, string(req))
	writeContractSourceFile(t, root, designPath, string(design))
	writeContractSourceFile(t, root, "docs/dev/contracts/SYNC-001.md", "# SYNC output\n")
	writeContractSourceFile(t, root, "docs/design/decisions/ADR-001.md", "# ADR-001\n[follow-up](ADR-002.md)\n")
	writeContractSourceFile(t, root, "docs/design/decisions/ADR-002.md", "# ADR-002\n")
	writeContractSourceFile(t, root, "docs/design/decisions/ADR-003.md", "> 状态: accepted\n> 关联需求: `docs/requirements/REQ-001.md`\n")
	writeContractSourceFile(t, root, "docs/design/decisions/ADR-004.md", "> 状态: accepted\n> 关联需求: `docs/requirements/REQ-002.md`\n")
	writeContractSourceFile(t, root, "docs/architecture/data-model/model.md", "# Model\n[order schema](order.schema.json) [example](order-example.json)\n")
	writeContractSourceFile(t, root, "docs/architecture/data-model/order.schema.json", `{"type":"object","properties":{"money":{"$ref":"money.schema.json"}}}`)
	writeContractSourceFile(t, root, "docs/architecture/data-model/money.schema.json", `{"type":"string"}`)
	writeContractSourceFile(t, root, "docs/architecture/data-model/order-example.json", `{"money":"USD 10"}`)
	if withModule {
		writeContractSourceFile(t, root, "docs/design/prototypes/orders/scenario-model.json", `{"module":"orders","rules":[{"source_refs":["REQ-001/FR-001","docs/design/decisions/ADR-001.md"]}]}`)
		writeContractSourceFile(t, root, "docs/design/prototypes/orders/fixture-contract.json", `{"module":"orders"}`)
		writeContractSourceFile(t, root, "docs/design/prototypes/orders/cross-matrix.json", `{"module":"orders"}`)
		writeContractSourceFile(t, root, "docs/design/prototypes/orders/cases.json", `{"module":"orders"}`)
		writeContractSourceFile(t, root, "docs/design/prototypes/orders/scenario-coverage.json", `{"module":"orders"}`)
		writeContractSourceFile(t, root, "docs/design/prototypes/orders/index.html", "<html></html>\n")
		writeContractSourceFile(t, root, "docs/design/prototypes/orders/page.html", "<html></html>\n")
		writeContractSourceFile(t, root, "docs/design/prototypes/orders/flows.md", "# Flow\n")
		writeContractSourceFile(t, root, "docs/design/prototypes/orders/stories.md", "# Stories\n[ADR](../../decisions/ADR-002.md)\n")
		writeContractSourceFile(t, root, "docs/design/prototypes/orders/assets/banner.svg", "<svg></svg>\n")
	}
	state := map[string]any{
		"schema_version": "1.2.0",
		"runtime_id":     "loop-contract-inputs",
		"configuration":  map[string]any{"optimization": map[string]any{"packet": true}},
		"baseline":       map[string]any{"generation": 1},
		"lifecycle":      map[string]any{"state": "planning", "phase": "contracts", "phase_revision": 3},
		"bound_req": map[string]any{
			"id": "REQ-001", "path": reqPath, "version": "v1.0.0", "sha256": workpacket.Digest(req), "status": "locked",
			"workspace": map[string]any{"project_root": root},
		},
		"documents": []any{map[string]any{
			"id": "ARCHITECTURE-001", "kind": "design", "path": designPath, "version": "v1.0.0", "sha256": workpacket.Digest(design), "status": "locked", "generation": 1,
		}},
	}
	return root, state
}

func writeContractSourceFile(t *testing.T, root, path, content string) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}

func writeSharedModelIndex(t *testing.T, root string, addRefund bool) {
	t.Helper()
	index := "# CONTRACTS-001\n> 需求: REQ-001\n> Shared model policy: json-schema-v1\n\n## Shared model baseline\n\n| Operation | Slot | Schema | Consumers | Valid example | Structural negative |\n|:--|:--|:--|:--|:--|:--|\n| create | order | [Indexed order](../../architecture/data-model/indexed-order.schema.json) | [FE contract](FE-001.md) | [valid](../../architecture/data-model/indexed-valid.json) | [invalid](../../architecture/data-model/indexed-invalid.json) |\n"
	if addRefund {
		index += "| create | refund | [Indexed refund](../../architecture/data-model/indexed-refund.schema.json) | [FE contract](FE-001.md) | [valid](../../architecture/data-model/indexed-refund-valid.json) | [invalid](../../architecture/data-model/indexed-refund-invalid.json) |\n"
	}
	writeContractSourceFile(t, root, "docs/dev/contracts/CONTRACTS-001.md", index)
	writeContractSourceFile(t, root, "docs/architecture/data-model/indexed-order.schema.json", `{"type":"object","properties":{"detail":{"$ref":"indexed-detail.schema.json"}},"required":["detail"]}`)
	writeContractSourceFile(t, root, "docs/architecture/data-model/indexed-detail.schema.json", `{"type":"object","properties":{"name":{"type":"string"}},"required":["name"]}`)
	writeContractSourceFile(t, root, "docs/architecture/data-model/indexed-valid.json", `{"detail":{"name":"Order"}}`)
	writeContractSourceFile(t, root, "docs/architecture/data-model/indexed-invalid.json", `{"detail":{}}`)
	consumer := "# FE-001\n\n## Shared model inputs\n\n[Indexed order](../../architecture/data-model/indexed-order.schema.json)\n"
	if addRefund {
		consumer += "[Indexed refund](../../architecture/data-model/indexed-refund.schema.json)\n"
	}
	writeContractSourceFile(t, root, "docs/dev/contracts/FE-001.md", consumer)
}

type pinnedContractFiles struct {
	root      string
	overrides map[string][]byte
}

func (p pinnedContractFiles) ReadFile(path string) ([]byte, error) {
	if data, ok := p.overrides[filepath.ToSlash(filepath.Clean(path))]; ok {
		return append([]byte(nil), data...), nil
	}
	return os.ReadFile(filepath.Join(p.root, filepath.FromSlash(path)))
}

func (p pinnedContractFiles) ReadDir(path string) ([]os.DirEntry, error) {
	return os.ReadDir(filepath.Join(p.root, filepath.FromSlash(path)))
}
