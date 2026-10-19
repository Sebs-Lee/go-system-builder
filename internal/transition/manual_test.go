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
		"explicitly named legacy/recovery command rejects an artifact",
		"runtime s7-budget-decision",
	} {
		if !strings.Contains(manual, want) {
			t.Fatalf("S7 manual guidance must contain %q, got:\n%s", want, manual)
		}
	}
}

func TestManualQuestionnairePathIsTheNormalS6S7S9Flow(t *testing.T) {
	manual := transition.RenderManual(&transition.LoopDefinition{}, transition.ManualOptions{})

	for _, want := range []string{
		"docs/examples/s6-builder/README.md",
		"runtime workspace prepare --root <authority-root> --assignment <id> --agent <owner>",
		"exact PLAN and Result paths",
		"runtime workspace launch",
		"Do not run `runtime task-complete`",
		"exact Hook-supplied ReviewPlan path",
		"Do not run `runtime review-result submit`",
		"one independent TargetedReverification per approved Contract",
		"no author-filled Handoff is needed",
	} {
		if !strings.Contains(manual, want) {
			t.Fatalf("normal questionnaire path must contain %q, got:\n%s", want, manual)
		}
	}

	for _, stale := range []string{
		"Use the returned begin/commit/report/check commands",
		"team launch --manifest <path> --request-template <template>",
		"runtime review-plan --file <plan.json>",
		"send the generic PLAN_REPORT and submit the S9 domain PlanReport",
	} {
		if strings.Contains(manual, stale) {
			t.Fatalf("manual still advertises a duplicate normal-path step %q", stale)
		}
	}
}

func TestManualS10GuidanceNamesManifestAndClosedLoop(t *testing.T) {
	manual := transition.RenderManual(&transition.LoopDefinition{}, transition.ManualOptions{})

	for _, want := range []string{
		"exact acceptance or release-audit manifest answer path",
		"coverage/counterevidence row once",
		"eight audit-area rows",
		"original consumer validates the saved manifest and derives its envelope and metrics",
		"Do not hand-run manifest validate/render",
		"zero UNKNOWN",
		"S8 → S9 → a fresh complete S7",
		"runtime human-decision",
	} {
		if !strings.Contains(manual, want) {
			t.Fatalf("S10 manual guidance must contain %q, got:\n%s", want, manual)
		}
	}
	for _, stale := range []string{"audit_manifest_path", "audit_manifest_sha256", "loop-harness s10 manifest validate"} {
		if strings.Contains(manual, stale) {
			t.Fatalf("normal S10 guidance must not ask the author to copy or submit %q", stale)
		}
	}
}

func TestManualApprovedAmendmentUsesOriginalReceiverAndCurrentQuestionnaire(t *testing.T) {
	def := &transition.LoopDefinition{Transitions: []transition.TransitionSpec{{ID: "TR-020", From: "paused", To: "planning", RequiredEvidence: []string{"human_decision_record", "req_lock_record"}}}}
	manual := transition.RenderManual(def, transition.ManualOptions{})
	for _, want := range []string{
		"## Approved requirement amendment {#approved-requirement-amendment}",
		"skip this step when already paused",
		"runtime pause --root <project-root> --approved-by <actual-human> --reason <approved-correction>",
		"same REQ ID", "raise its semantic version", "bound development branch",
		"req amend --root <project-root> --req docs/requirements/REQ-<id>.md --approved-by <actual-human>",
		"exact current S2 questionnaire view/answer paths", "do not ask for the same approval again",
	} {
		if !strings.Contains(manual, want) {
			t.Fatalf("normal amendment guidance missing %q", want)
		}
	}
	legacy := transition.RenderTransition(def, "TR-020")
	if !strings.Contains(legacy, "Low-level diagnostic evidence slots only") || !strings.Contains(legacy, "#approved-requirement-amendment") || strings.Contains(legacy, "copy into `runtime transition`") || strings.Contains(legacy, "retry with the command above") {
		t.Fatalf("TR020 still presented as normal evidence composition: %s", legacy)
	}
	if err := transition.ValidateManualEvidenceBindings(def, manual); err != nil {
		t.Fatal(err)
	}
}
