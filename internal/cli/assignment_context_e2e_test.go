package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/entroforge/go-system-builder/internal/agentcontext"
	"github.com/entroforge/go-system-builder/internal/cli"
	"github.com/entroforge/go-system-builder/internal/schema"
	"github.com/entroforge/go-system-builder/internal/workpacket"
)

func nativeContextHook(t *testing.T, f stageAnswerFixture, owner string) (string, int, string) {
	t.Helper()
	body, _ := json.Marshal(map[string]any{"hook_event_name": "SubagentStart", "agent_id": owner, "cwd": f.root})
	var out, errout bytes.Buffer
	code := cli.Run([]string{"hook", "--root", f.root, "--event", "SubagentStart"}, bytes.NewReader(body), &out, &errout)
	var wire struct {
		Specific struct {
			Context string `json:"additionalContext"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal(out.Bytes(), &wire); err != nil {
		t.Fatalf("Hook wire: %v %s %s", err, out.String(), errout.String())
	}
	return wire.Specific.Context, code, errout.String()
}

func nativeContextRead(t *testing.T, f stageAnswerFixture, owner string) (agentcontext.Reference, map[string]any) {
	t.Helper()
	context, code, errors := nativeContextHook(t, f, owner)
	if code != 0 || !utf8.ValidString(context) || utf8.RuneCountInString(context) > 800 {
		t.Fatalf("wire %d: %s %s", code, context, errors)
	}
	match := regexp.MustCompile(`first Read ([^;]+); sha256=([a-f0-9]{64})`).FindStringSubmatch(context)
	if match == nil {
		t.Fatalf("no complete actual context entrance: %s", context)
	}
	ref := agentcontext.Reference{Path: match[1], SHA256: match[2]}
	data, err := os.ReadFile(filepath.Join(f.root, ref.Path))
	if err != nil || workpacket.Digest(data) != ref.SHA256 {
		t.Fatalf("source %v", err)
	}
	if err := schema.NewEmbeddedValidator().ValidateBytes("native-assignment-context.schema.json", data); err != nil {
		t.Fatal(err)
	}
	var c map[string]any
	if err := json.Unmarshal(data, &c); err != nil {
		t.Fatal(err)
	}
	return ref, c
}

func TestNativeAssignmentContextDeliversActualSourcesAndOriginalPlan(t *testing.T) {
	f := newStageAnswerFixture(t)
	p := f.packets[0]
	before := stageSnapshot(t, f.root)
	ref, c := nativeContextRead(t, f, p.Binding.Owner)
	after := stageSnapshot(t, f.root)
	// The original Controller persists a first lifecycle milestone. Compare
	// actual dispatch/authority instead of treating its original journal as a
	// projection mutation. The read-only team entry is checked separately.
	for _, key := range []string{"entities", "lifecycle", "baseline", "evidence"} {
		left, _ := json.Marshal(before.State[key])
		right, _ := json.Marshal(after.State[key])
		if !bytes.Equal(left, right) {
			t.Fatalf("projection/Controller changed original dispatch or authority: %s", key)
		}
	}
	dispatch := c["dispatch"].(map[string]any)
	if dispatch["owner"] != p.Binding.Owner || dispatch["assignment_id"] != p.Binding.AssignmentID || dispatch["task_id"] != f.taskID || c["work_packet_ref"] != p.ViewPath || c["answer_path"] != p.AnswerPath {
		t.Fatal("invented dispatch or foreign answer")
	}
	if len(c["product_write_paths"].([]any)) != 0 {
		t.Fatal("S5 projected product authority")
	}
	if c["context_selection"].(map[string]any)["mode"] != "shadow" {
		t.Fatal("projection enabled profile")
	}
	inputs := c["frozen_inputs"].(map[string]any)
	for path, hash := range inputs {
		data, err := os.ReadFile(filepath.Join(f.root, path))
		if err != nil || workpacket.Digest(data) != hash {
			t.Fatalf("unavailable actual source %s %v", path, err)
		}
	}
	order := c["input_read_order"].([]any)
	if len(order) < 3 || !strings.HasPrefix(order[0].(string), "docs/dev/tasks/") {
		t.Fatal("lost original bottom-up source")
	}
	planPath := c["plan_draft_path"].(string)
	data, err := os.ReadFile(filepath.Join(f.root, planPath))
	if err != nil {
		t.Fatal(err)
	}
	if err := schema.NewEmbeddedValidator().ValidateBytes("agent-message.schema.json", data); err == nil {
		t.Fatal("empty draft became professional plan")
	}
	var plan map[string]any
	json.Unmarshal(data, &plan)
	plan["objective"] = "Evaluate my original registered independent document responsibility"
	plan["steps"] = []any{map[string]any{"description": "Read the assigned chain and cite concrete gaps", "target": f.taskPath}}
	plan["assertion_checks"] = []any{map[string]any{"assertion": "Required responsibility is fully accounted for", "oracle": "Cited findings compared with the actual frozen inputs"}}
	plan["risks_blockers"] = []any{}
	stageJSON(t, f.root, planPath, plan)
	authored, _ := os.ReadFile(filepath.Join(f.root, planPath))
	again, _ := nativeContextRead(t, f, p.Binding.Owner)
	if again != ref {
		t.Fatal("same original sources created a second projection")
	}
	retried, _ := os.ReadFile(filepath.Join(f.root, planPath))
	if !bytes.Equal(authored, retried) {
		t.Fatal("native retry overwrote professional plan")
	}
	stageCLI(t, "runtime", "agent-begin", "--root", f.root, "--agent-id", p.Binding.Owner, "--plan", planPath)
	if !saveStageHook(t, f.root, f.root, p.Binding.Owner, p.AnswerPath, completedDraft(p, "pass")).WorkComplete {
		t.Fatal("original plan and save did not admit answer")
	}
	context, _, _ := nativeContextHook(t, f, p.Binding.Owner)
	if !strings.Contains(context, "work_complete=true") || strings.Contains(context, "first Read ") {
		t.Fatalf("received responsibility got another planning prompt: %s", context)
	}
}

func TestNativeAssignmentContextDoesNotGuessOwnerOrFallbackFromExecution(t *testing.T) {
	f := newStageAnswerFixture(t)
	context, _, _ := nativeContextHook(t, f, "unknown-owner")
	if strings.Contains(context, "assignment-context/native") || strings.Contains(context, f.packets[0].Binding.Owner) {
		t.Fatal("unknown identity got another owner's view")
	}
	p := f.packets[0]
	f.prepare(t, p)
	context, _, _ = nativeContextHook(t, f, p.Binding.Owner)
	if strings.Contains(context, "assignment-context/native") {
		t.Fatal("independent execution fell back to Main projection")
	}
}

func TestNativeAssignmentContextRejectsMissingSourcesWithoutAuthorOverwrite(t *testing.T) {
	f := newStageAnswerFixture(t)
	p := f.packets[0]
	_, c := nativeContextRead(t, f, p.Binding.Owner)
	path := c["plan_draft_path"].(string)
	before, _ := os.ReadFile(filepath.Join(f.root, path))
	source := filepath.Join(f.root, ".claude/skills/document-verification/SKILL.md")
	if err := os.Remove(source); err != nil {
		t.Fatal(err)
	}
	context, _, _ := nativeContextHook(t, f, p.Binding.Owner)
	if !strings.Contains(context, "context unavailable") || strings.Contains(context, "first Read ") {
		t.Fatalf("missing installed source became ready: %s", context)
	}
	after, _ := os.ReadFile(filepath.Join(f.root, path))
	if !bytes.Equal(before, after) {
		t.Fatal("failed projection overwrote author bytes")
	}
}

func TestRegisteredTeammateContextMatchesNativeAndPreservesOriginalAuthority(t *testing.T) {
	f := newStageAnswerFixture(t)
	// Deliver the original legacy Skill set exactly as an installation does.
	// These are test setup assets, not production fallback from missing files.
	for _, name := range agentcontext.LegacySkills("document-verifier") {
		source, err := os.ReadFile(filepath.Join("../..", "skills", name, "SKILL.md"))
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(f.root, ".claude/skills", name, "SKILL.md")
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, source, 0600); err != nil {
			t.Fatal(err)
		}
	}
	example, err := schema.ReadAsset("agent-message.examples.json")
	if err != nil {
		t.Fatal(err)
	}
	var messages []map[string]any
	json.Unmarshal(example, &messages)
	var template map[string]any
	for _, m := range messages {
		if m["message_type"] == "readback_request" {
			template = m
			break
		}
	}
	template["task_id"] = f.taskID
	docs := []any{}
	state := stageSnapshot(t, f.root).State
	for i, kind := range []string{"task", "contract", "req", "design"} {
		for _, raw := range state["documents"].([]any) {
			doc := raw.(map[string]any)
			if doc["kind"] == kind {
				docs = append(docs, map[string]any{"id": doc["id"], "kind": kind, "path": doc["path"], "version": doc["version"], "sha256": doc["sha256"], "read_order": i + 1})
				break
			}
		}
	}
	template["documents"] = docs
	path := ".claude/registered-readback-template.json"
	stageJSON(t, f.root, path, template)
	// Warm only the original Controller milestone, then compare both entrances.
	nativeRef, _ := nativeContextRead(t, f, f.packets[0].Binding.Owner)
	beforeState, _ := os.ReadFile(filepath.Join(f.root, ".claude/loop-state.json"))
	beforeJournal, _ := os.ReadFile(filepath.Join(f.root, ".claude/loop-events.jsonl"))
	output := stageCLI(t, "team", "launch", "--root", f.root, "--manifest", ".claude/workgroups/REQ-039/review/manifest.json", "--request-template", path)
	var packages []map[string]any
	if err := json.Unmarshal([]byte(output), &packages); err != nil {
		t.Fatal(err)
	}
	if len(packages) != 2 {
		t.Fatal("lost independent owner")
	}
	validator := schema.NewEmbeddedValidator()
	for _, p := range packages {
		data, _ := json.Marshal(p)
		if err := validator.ValidateBytes("agent-message.schema.json", data); err != nil {
			t.Fatal(err)
		}
		ref := p["assignment_context_ref"].(map[string]any)
		actual, err := os.ReadFile(filepath.Join(f.root, ref["path"].(string)))
		if err != nil || workpacket.Digest(actual) != ref["sha256"] {
			t.Fatal("teammate got an unreadable context")
		}
		if p["agent_id"] == f.packets[0].Binding.Owner && (ref["path"] != nativeRef.Path || ref["sha256"] != nativeRef.SHA256) {
			t.Fatal("native and teammate produced different current sources")
		}
	}
	afterState, _ := os.ReadFile(filepath.Join(f.root, ".claude/loop-state.json"))
	afterJournal, _ := os.ReadFile(filepath.Join(f.root, ".claude/loop-events.jsonl"))
	if !bytes.Equal(beforeState, afterState) || !bytes.Equal(beforeJournal, afterJournal) {
		t.Fatal("read-only team projection mutated authority")
	}
	// A schema-shaped request with a guessed subject digest cannot be emitted
	// as the registered owner's launch, even when its file exists.
	docs[0].(map[string]any)["sha256"] = strings.Repeat("a", 64)
	stageJSON(t, f.root, path, template)
	var out, errout bytes.Buffer
	if code := cli.Run([]string{"team", "launch", "--root", f.root, "--manifest", ".claude/workgroups/REQ-039/review/manifest.json", "--request-template", path}, nil, &out, &errout); code == 0 || !strings.Contains(errout.String(), "differs from manifest") {
		t.Fatalf("guessed source became registered launch: %d %s", code, errout.String())
	}
}

func TestNativeAssignmentContextDoesNotIssuePlanForRetiredS5Peer(t *testing.T) {
	f := newStageAnswerFixture(t)
	first := f.packets[0]
	f.begin(t, first, nil)
	receipt := saveStageHook(t, f.root, f.root, first.Binding.Owner, first.AnswerPath, completedDraft(first, "fix_required"))
	if !receipt.WorkComplete || stageSnapshot(t, f.root).State["lifecycle"].(map[string]any)["state"] != "planning" {
		t.Fatal("original FIX did not retire its S5 entry")
	}
	context, _, _ := nativeContextHook(t, f, f.packets[1].Binding.Owner)
	if strings.Contains(context, "first Read ") || !strings.Contains(context, "no longer current") ||
		!strings.Contains(context, "No new plan or source view was issued.") || strings.Contains(context, "Main must repair the original source") {
		t.Fatalf("retired peer got another plan/source view: %s", context)
	}
	if _, err := os.Stat(filepath.Join(f.root, ".claude/assignment-context/native")); !os.IsNotExist(err) {
		t.Fatalf("retired responsibility published new context: %v", err)
	}
}
