package policy_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/entroforge/go-system-builder/internal/policy"
)

// ---------------------------------------------------------------------------
// §14.1: Reviewer 尝试写产品代码 —— PreToolUse hard deny，recovery 给出
// 授权 evidence 路径；授权写面（.claude/、docs/reports/、ReviewPlan
// verification_artifact_workspace）和非 verification 阶段不受影响
// (L3-S7 §1.4.1, §8)。
// ---------------------------------------------------------------------------

func evaluateWrite(t *testing.T, engine *policy.Engine, tool, path, stage, workspace string) policy.Decision {
	t.Helper()
	decision, err := engine.Evaluate(policy.Input{
		Event:     "PreToolUse",
		ToolName:  tool,
		ToolInput: map[string]any{"file_path": path},
		Runtime: policy.RuntimeContext{
			CurrentState:          stage,
			VerificationWorkspace: workspace,
		},
	})
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	return decision
}

func TestReviewerProductWriteHardDeny(t *testing.T) {
	engine, err := policy.Load(filepath.Join("..", "..", "docs", "hook-policy.json"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	decision := evaluateWrite(t, engine, "Write", "internal/example/service.go", "verification", "e2e-workspace/plan-1")
	if decision.Decision != "block" || decision.RuleID != policy.RuleReviewerProductWrite {
		t.Fatalf("product write in verification must hard deny, got %q (%s)", decision.Decision, decision.RuleID)
	}
	// The deny names the authorized evidence surfaces so the Reviewer can
	// proceed without a human round-trip.
	joined := strings.Join(decision.Recovery, "\n")
	if !strings.Contains(joined, ".claude/") || !strings.Contains(joined, "verification_artifact_workspace") {
		t.Fatalf("recovery must give the authorized evidence paths, got %v", decision.Recovery)
	}
	if decision.AffectedPath != "internal/example/service.go" {
		t.Fatalf("affected path = %q", decision.AffectedPath)
	}

	// Locked spec writes are product-surface writes too.
	decision = evaluateWrite(t, engine, "Edit", "docs/contracts/CONTRACTS-001.md", "verification", "")
	if decision.Decision != "block" || decision.RuleID != policy.RuleReviewerProductWrite {
		t.Fatalf("locked spec edit in verification must hard deny, got %q", decision.Decision)
	}
}

func TestReviewerWriteAuthorizedSurfacesStayOpen(t *testing.T) {
	engine, err := policy.Load(filepath.Join("..", "..", "docs", "hook-policy.json"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	workspace := "e2e-workspace/plan-1"
	allowed := []struct{ tool, path string }{
		{"Write", ".claude/evidence/loop-REQ-1/g1/reviews/agent-qa-1/result.json"}, // control plane
		{"Write", "docs/reports/review/REV-001.md"},                                // report projection
		{"Write", "e2e-workspace/plan-1/flow.spec.ts"},                             // cold-start workspace
		{"Edit", "e2e-workspace/plan-1/fixtures/seed.json"},                        // cold-start fixture
	}
	for _, tc := range allowed {
		decision := evaluateWrite(t, engine, tc.tool, tc.path, "verification", workspace)
		if decision.Decision == "block" {
			t.Fatalf("authorized surface %s must stay open in verification, got block (%s)", tc.path, decision.RuleID)
		}
	}
	// A write into a *different* workspace than the plan declared is still a
	// product-surface write.
	decision := evaluateWrite(t, engine, "Write", "e2e-workspace/other/spec.ts", "verification", workspace)
	if decision.Decision != "block" {
		t.Fatalf("writes outside the declared workspace must deny, got %q", decision.Decision)
	}
}

func TestReviewerProductWriteRuleScopedToVerificationStage(t *testing.T) {
	engine, err := policy.Load(filepath.Join("..", "..", "docs", "hook-policy.json"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	// Outside the verification stage the frozen-baseline rule does not fire
	// (other stages' guards own their write surfaces).
	decision := evaluateWrite(t, engine, "Write", "internal/example/service.go", "building", "")
	if decision.RuleID == policy.RuleReviewerProductWrite {
		t.Fatal("reviewer_product_write must not fire outside the verification stage")
	}
	// Non-write tools are out of scope even in verification.
	decision = evaluateWrite(t, engine, "Bash", "", "verification", "")
	if decision.RuleID == policy.RuleReviewerProductWrite {
		t.Fatal("reviewer_product_write must not fire for non-write tools")
	}
}
