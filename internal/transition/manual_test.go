package transition_test

import (
	"strings"
	"testing"

	"github.com/entroforge/go-system-builder/internal/transition"
)

func TestManualHeaderContainsNoVolatileTimestamp(t *testing.T) {
	manual := transition.RenderManual(&transition.LoopDefinition{}, transition.ManualOptions{
		TargetPath:           "loop-harness.md",
		HarnessVersion:       "dev",
		LoopDefinitionSHA256: strings.Repeat("a", 64),
	})
	if strings.Contains(manual, "**Generated**") {
		t.Fatalf("manual output must be deterministic; got volatile timestamp header: %s", manual)
	}
}

func TestManualEvidenceBindingGuidanceIncludesSlotTemplateAndAcceptedKinds(t *testing.T) {
	manual := transition.RenderManual(&transition.LoopDefinition{
		Transitions: []transition.TransitionSpec{
			{
				ID:               "TR-006",
				From:             "building",
				To:               "verification",
				RequiredEvidence: []string{"builder_report_record", "team_manifest_record"},
				Description:      "Start a review round.",
			},
		},
	}, transition.ManualOptions{})

	for _, want := range []string{
		"--evidence builder_report_record=<reference>",
		"--evidence team_manifest_record=<reference>",
		"Accepted kinds: `builder_report`, `agent_completion`",
		"Accepted kinds: `builder_report`, `team_manifest`",
	} {
		if !strings.Contains(manual, want) {
			t.Fatalf("manual must contain %q, got:\n%s", want, manual)
		}
	}
}

func TestManualIncludesExplicitS11HumanDecisionGuidance(t *testing.T) {
	manual := transition.RenderManual(&transition.LoopDefinition{}, transition.ManualOptions{})

	for _, want := range []string{
		"runtime human-decision",
		"approve",
		"defer",
		"reject_defect",
		"reject_acceptance",
		"reject_release_audit",
		"abort",
		"release_authorized",
		"Harness has no squash merge, publication, deployment, or formal release permission",
	} {
		if !strings.Contains(manual, want) {
			t.Fatalf("S11 manual guidance must contain %q, got:\n%s", want, manual)
		}
	}
}

func TestManualS7RecoveryGuidanceNamesRepairFactsAndTypedEvidence(t *testing.T) {
	manual := transition.RenderManual(&transition.LoopDefinition{}, transition.ManualOptions{})

	for _, want := range []string{
		"coverage_inventory",
		"e2e_assets",
		"capture step --finding <id> --claim <id>",
		"rejected command includes the missing facts, repair action, next command",
		"runtime s7-budget-decision",
	} {
		if !strings.Contains(manual, want) {
			t.Fatalf("S7 manual guidance must contain %q, got:\n%s", want, manual)
		}
	}
}
