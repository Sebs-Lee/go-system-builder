package qualitygate

import (
	"encoding/json"
	"github.com/entroforge/go-system-builder/internal/fileview"
	"github.com/entroforge/go-system-builder/internal/runtime"
	"os"
	"path/filepath"
	"testing"
)

func TestPlannedProgressRequiresCurrentReportAndAssignment(t *testing.T) {
	root := t.TempDir()
	put := func(p string, v any) []byte {
		t.Helper()
		b, _ := json.Marshal(v)
		os.MkdirAll(filepath.Dir(filepath.Join(root, p)), 0755)
		if e := os.WriteFile(filepath.Join(root, p), b, 0644); e != nil {
			t.Fatal(e)
		}
		return b
	}
	report := ".claude/result.json"
	checkpoint := ".claude/evidence/run/g1/worktree/a1/checkpoint.json"
	env := map[string]any{"evidence_id": "E1", "kind": "completion_report", "task_id": "TASK-042-01", "runtime_id": "run", "baseline_generation": 1, "producer_agent_id": "builder", "conclusion": "completed", "checks": []any{map[string]any{"name": "test", "result": "pass"}}, "created_at": "2026-09-19T10:00:00Z"}
	b := put(report, env)
	cp := map[string]any{"task_id": "TASK-042-01", "assignment_id": "a1", "baseline_generation": 1, "state": "verified", "verified_at": "2026-09-19T11:00:00Z"}
	cp["completion_report_path"] = report
	cp["completion_report_sha256"] = sha256Hex(b)
	put(checkpoint, cp)
	task := map[string]any{"id": "TASK-042-01", "state": "review", "owner_agent_ids": []any{"builder"}, "completion_report_ref": report}
	index := map[string]any{"id": "E1", "path": report, "kind": "completion_report", "status": "valid", "baseline_generation": 1, "sha256": sha256Hex(b)}
	agent := map[string]any{"id": "builder", "state": "reported", "prompt_ref": "manifest#a1"}
	state := map[string]any{"runtime_id": "run", "baseline": map[string]any{"generation": 1}, "entities": map[string]any{"tasks": []any{task}, "agents": []any{agent}}, "evidence": []any{index}}
	input := Input{Root: root, Snapshot: runtime.Snapshot{State: state}, Files: fileview.Disk{Root: root}}
	check := func(want string) {
		t.Helper()
		if got := PlannedBuilderProgress(input)["TASK-042-01"].State; got != want {
			t.Fatalf("want %s got %s", want, got)
		}
	}
	check("integrated")
	delete(cp, "completion_report_sha256")
	put(checkpoint, cp)
	check("reported") // Old checkpoints need fresh verification for planned dispatch.
	cp["completion_report_sha256"] = sha256Hex(b)
	cp["completion_report_path"] = ".claude/other-result.json"
	put(checkpoint, cp)
	check("reported")
	cp["completion_report_path"] = report
	put(checkpoint, cp)
	check("integrated")
	agent["prompt_ref"] = "manifest#a2"
	check("reported")
	agent["prompt_ref"] = "manifest#a1"
	index["invalidated_by"] = "superseded"
	check("reported")
	delete(index, "invalidated_by")
	env["created_at"] = "2026-09-19T12:00:00Z"
	b = put(report, env)
	index["sha256"] = sha256Hex(b)
	check("reported")
	env["created_at"] = "2026-09-19T10:00:00Z"
	env["checks"] = []any{map[string]any{"result": "fail"}}
	b = put(report, env)
	index["sha256"] = sha256Hex(b)
	check("reported")
	task["state"] = "blocked"
	check("blocked")
}
