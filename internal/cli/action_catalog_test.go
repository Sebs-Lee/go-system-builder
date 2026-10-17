package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/entroforge/go-system-builder/internal/cli"
)

func TestActionsCatalogShowsCanonicalReviewRepairLoop(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := cli.Run([]string{"actions"}, strings.NewReader(""), &stdout, &stderr); code != 0 {
		t.Fatalf("actions failed: code=%d stderr=%s", code, stderr.String())
	}
	for _, want := range []string{
		"Canonical Agent action catalog",
		"S6 Builder implementation",
		"exact PLAN and Result view/answer paths",
		"original begin flow",
		"original report/completion Writer",
		"runs its TASK checks",
		"s6 prepare --revise-assignment <id> --revision-intent-key <new-key>",
		"runtime workspace pending",
		"runtime task-integrate --assignment-id <id>",
		"ReviewPlan questionnaire",
		"s7 manifest-draft --assignment <id>",
		"original ReviewResult consumer",
		"runtime investigation ingest --root <root> --grouping-rationale",
		"runtime investigation hypothesis register",
		"original Case consumer",
		"Main listener uses the original readback to prepare the RepairSession and RepairPlan",
		"Main dispatches each real RepairAssignment with `runtime repair dispatch`",
		"original PlanReport consumer",
		"exact Hook-provided RepairResult",
		"Hook-provided ChangeImpact answer",
		"TargetedReverification answer",
		"existing handoff consumer",
		"S10: use `s10 status` for read-only progress",
		"original consumer validates it, derives machine metadata",
		"only the human's explicit decision uses `runtime human-decision`",
		"fresh complete S7 round",
		"not a second delivery after saving a Hook-provided answer",
		"runtime repair session open`, `runtime repair plan compile`",
		"runtime transition` is legacy/low-level",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("action catalog missing %q:\n%s", want, stdout.String())
		}
	}
	for _, obsoleteNormalFlow := range []string{
		"s7 draft --out <plan.json>",
		"runtime review-plan --file <plan.json>",
		"runtime review-result submit --assignment-id <id> --result <result.json>",
		"runtime investigation hypothesis result` →",
		"runtime repair plan-report submit --file <plan-report.json>",
		"runtime repair result submit --file <result.json>",
		"runtime repair impact create|commit",
		"s10 manifest validate --file <manifest.json>` → register evidence",
	} {
		if strings.Contains(stdout.String(), obsoleteNormalFlow) {
			t.Fatalf("action catalog still presents %q as normal flow:\n%s", obsoleteNormalFlow, stdout.String())
		}
	}
}

func TestPrimaryCommandHelpIsActionable(t *testing.T) {
	for _, args := range [][]string{
		{"s7", "--help"},
		{"s7", "status", "--help"},
		{"s10", "--help"},
		{"runtime", "--help"},
		{"runtime", "repair", "--help"},
	} {
		var stdout, stderr bytes.Buffer
		if code := cli.Run(args, strings.NewReader(""), &stdout, &stderr); code != 0 {
			t.Fatalf("%v help failed: code=%d stdout=%s stderr=%s", args, code, stdout.String(), stderr.String())
		}
		if !strings.Contains(stdout.String(), "Usage:") || strings.Contains(stderr.String(), "requires") {
			t.Fatalf("%v help is not a clean actionable response: stdout=%s stderr=%s", args, stdout.String(), stderr.String())
		}
	}
}
