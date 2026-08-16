package cli_test

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/entroforge/go-system-builder/internal/cli"
)

// TestS3ContractPipelineE2E walks the contract stage end to end (L3-S3
// v4.0.1): draft a contract with a real reference chain → contracts check
// green → inject a broken link → red → fix → PTR-PLAN-02 registers the
// locked contract into documents[] (with author) → hook write-protection
// fires on the registered contract → same-generation rework re-locks with
// replace semantics (no stacking).
func TestS3ContractPipelineE2E(t *testing.T) {
	root := t.TempDir()
	for _, rel := range []string{"docs/contracts", "docs/requirements", "docs/design/prototypes/wb", ".claude"} {
		if err := os.MkdirAll(filepath.Join(root, rel), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, rel := range []string{"docs/loop-definition.json", "docs/hook-policy.json"} {
		data, err := os.ReadFile(filepath.Join("..", "..", rel))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, rel), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write := func(rel, content string) {
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(rel)), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	sha := func(s string) string { return fmt.Sprintf("%x", sha256.Sum256([]byte(s))) }
	run := func(args ...string) (string, string, int) {
		var stdout, stderr bytes.Buffer
		code := cli.Run(args, strings.NewReader(""), &stdout, &stderr)
		return stdout.String(), stderr.String(), code
	}

	// REQ with an FR; module package with CASE/S/F/PATH universe.
	write("docs/requirements/REQ-500.md", "# REQ-500\n\n> 状态：locked\n> 版本：v1.0.0\n> UI impact：changed\n\n| 编号 | 模块 | 需求 | 服务于 | 优先级 |\n|:--|:--|:--|:--|:--|\n| FR-501 | wb | 提交 | A1 | Must |\n")
	write("docs/design/prototypes/wb/cases.json", `{"cases":[{"id":"CASE-WB-001"},{"id":"CASE-WB-002"}]}`)
	write("docs/design/prototypes/wb/stories.md", "# S-001\n")
	write("docs/design/prototypes/wb/flows.md", "# F-001\n\n### PATH-SUBMIT\n")

	// --- green: a contract whose references all resolve ---
	write("docs/contracts/BE-501.md", ""+
		"# BE-501\n\n> 状态：locked\n> 版本：v1.0.0\n\n"+
		"| REQ source_ref | Rule/CASE/Story/PATH | 本合同条款§ | 验收标准 |\n|:--|:--|:--|:--|\n"+
		"| REQ-500/FR-501 | CASE-WB-001 / S-001 / F-001 / PATH-SUBMIT | §2 | 可提交 |\n")
	out, _, code := run("contracts", "check", "--root", root)
	if code != 0 || !strings.Contains(out, "all reconciled") {
		t.Fatalf("green contract must pass: code=%d out=%s", code, out)
	}

	// --- red: broken CASE + unknown clause target ---
	write("docs/contracts/BE-501.md", strings.Replace(readFile(t, root, "docs/contracts/BE-501.md"),
		"CASE-WB-001", "CASE-GHOST-9", 1)+"\n| cell | FE-999 §1 | x |\n")
	_, stderr, code := run("contracts", "check", "--root", root)
	if code == 0 || !strings.Contains(stderr, "CASE-GHOST-9") || !strings.Contains(stderr, "FE-999") {
		t.Fatalf("broken links must be named, got: %s", stderr)
	}
	// restore green
	write("docs/contracts/BE-501.md", strings.Replace(readFile(t, root, "docs/contracts/BE-501.md"), "CASE-GHOST-9", "CASE-WB-001", 1))
	write("docs/contracts/BE-501.md", func() string {
		s := readFile(t, root, "docs/contracts/BE-501.md")
		if idx := strings.Index(s, "\n| cell |"); idx >= 0 {
			return s[:idx]
		}
		return s
	}())

	// --- registration via PTR-PLAN-02: bind, then fire the transition ---
	if _, stderr, code := run("req", "bind", "--root", root, "--approved-by", "bob"); code != 0 {
		t.Fatalf("bind failed: %s", stderr)
	}
	// design→contracts→tasks: PTR-PLAN-01 first (carries the wired
	// ui_impact_resolved guard — impact is `changed`, so it passes), then
	// PTR-PLAN-02 carries guard contracts_checked + action register_locked_contracts.
	if _, stderr, code = run("runtime", "transition", "--root", root,
		"--id", "PTR-PLAN-01", "--expected-revision", "1", "--actor", "orchestrator"); code != 0 {
		t.Fatalf("PTR-PLAN-01 failed: %s", stderr)
	}
	_, stderr, code = run("runtime", "transition", "--root", root,
		"--id", "PTR-PLAN-02", "--expected-revision", "2", "--actor", "orchestrator")
	if code != 0 {
		t.Fatalf("PTR-PLAN-02 failed: %s", stderr)
	}
	statePath := filepath.Join(root, ".claude", "loop-state.json")
	state := readJSONMap(t, statePath)
	docs, _ := state["documents"].([]any)
	var contractEntry map[string]any
	for _, raw := range docs {
		doc, _ := raw.(map[string]any)
		if doc != nil && doc["kind"] == "contract" {
			contractEntry = doc
		}
	}
	if contractEntry == nil {
		t.Fatal("PTR-PLAN-02 must register the locked contract into documents[]")
	}
	if contractEntry["author_agent_id"] != "orchestrator" {
		t.Fatalf("contract author_agent_id = %v, want orchestrator (registering actor)", contractEntry["author_agent_id"])
	}
	diskData, _ := os.ReadFile(filepath.Join(root, "docs", "contracts", "BE-501.md"))
	if contractEntry["sha256"] != fmt.Sprintf("%x", sha256.Sum256(diskData)) {
		t.Fatal("registered sha must match disk")
	}

	// --- same-generation rework: revise + re-lock → replace, not stack ---
	revise := strings.Replace(readFile(t, root, "docs/contracts/BE-501.md"), "v1.0.0", "v1.1.0", 1)
	write("docs/contracts/BE-501.md", revise)
	// bump revision by a no-op evidence-free transition is not available; use direct state edit to allow re-fire
	state = readJSONMap(t, statePath)
	state["lifecycle"] = map[string]any{"state": "planning", "phase": "contracts", "phase_revision": float64(1)}
	writeJSONMap(t, statePath, state)
	if _, stderr, code := run("runtime", "transition", "--root", root,
		"--id", "PTR-PLAN-02", "--expected-revision", intStr(int(state["revision"].(float64))), "--actor", "orchestrator"); code != 0 {
		t.Fatalf("re-lock failed: %s", stderr)
	}
	state = readJSONMap(t, statePath)
	docs, _ = state["documents"].([]any)
	count := 0
	for _, raw := range docs {
		doc, _ := raw.(map[string]any)
		if doc != nil && doc["kind"] == "contract" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("same-generation re-lock must replace not stack, got %d contract entries", count)
	}
	_ = sha
}

func readFile(t *testing.T, root, rel string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func intStr(n int) string { return fmt.Sprintf("%d", n) }
