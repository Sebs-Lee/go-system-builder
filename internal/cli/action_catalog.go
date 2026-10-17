package cli

import (
	"fmt"
	"io"
	"strings"
)

// wantsHelp is checked before subcommand-specific flag parsing. The standard
// flag package writes ErrHelp to stderr and returns exit 2 for nested commands;
// agents need help to be a successful, discoverable action instead.
func wantsHelp(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" || arg == "help" {
			return true
		}
	}
	return false
}

func printCommandHelp(w io.Writer, usage, detail string) {
	fmt.Fprintf(w, "Usage: %s\n\n", usage)
	if detail != "" {
		fmt.Fprintln(w, detail)
	}
	fmt.Fprintln(w, "Run `loop-harness actions` for the canonical stage action sequence and compatibility notes.")
}

// runActions is intentionally a small, stable index rather than another
// scheduler. It is the one agent-facing map from stage responsibility to the
// runtime verb that owns the mutation.
func runActions(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 && !wantsHelp(args) {
		fmt.Fprintf(stderr, "actions: unknown option %q\n", args[0])
		return 2
	}
	printActionCatalog(stdout)
	return 0
}

func printActionCatalog(w io.Writer) {
	fmt.Fprintln(w, "Canonical Agent action catalog")
	fmt.Fprintln(w, "For normal work, follow the current Hook instructions and exact questionnaire paths. Save each professional answer once; PostToolUse sends it to the original domain consumer. Main still owns real dispatch, integration, required checks, and human decisions. Do not use `runtime transition` directly.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "S6 Builder implementation")
	fmt.Fprintln(w, "  1. Main checks the TASK batch and uses `s6 prepare` to project the current TASK-bound manifest draft; review it and register the exact returned path with `runtime register-workgroup`. Use `runtime workspace prepare` for each registered Assignment, then `runtime workspace launch`; prepare supplies its frozen inputs and exact questionnaire paths.")
	fmt.Fprintln(w, "  2. The Hook supplies the Builder's exact PLAN and Result view/answer paths. The Builder saves the professional PLAN to trigger the original begin flow, implements the TASK, runs its TASK checks, commits the scoped delivery, then saves the Result with actual check outcomes for the original report/completion Writer.")
	fmt.Fprintln(w, "  3. Main runs `runtime task-integrate --assignment-id <id>` and the registered required checks. The complete verified batch advances to S7.")
	fmt.Fprintln(w, "  4. For failed canonical integration needing code correction, Main follows the Hook's rework argv or `runtime workspace pending`. Use `s6 prepare --revise-assignment <id> --revision-intent-key <new-key>` with the new Builder identities for the exact revision draft and original registration command; then prepare and launch that new attempt.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "S7 complete verification")
	fmt.Fprintln(w, "  1. Main completes the Hook-provided ReviewPlan questionnaire and saves it for the original RegisterPlan consumer.")
	fmt.Fprintln(w, "  2. Main creates manifests for the actual Assignments with `s7 manifest-draft --assignment <id>` and registers/dispatches them with `runtime register-workgroup`. The Hook supplies each Reviewer the exact answer path; save the professional ReviewResult once for the original ReviewResult consumer.")
	fmt.Fprintln(w, "  3. Main follows `s7 status` for Claim coverage, blockers, and the next action. The original consumer seals Findings or records a clean round.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "S8 root-cause investigation")
	fmt.Fprintln(w, "  1. After S7 seals an ObservationBatch and the normal Hook advances to S8, Main consumes that seal with `runtime investigation ingest --root <root> --grouping-rationale <evidence-based grouping>`.")
	fmt.Fprintln(w, "  2. Main registers each falsifiable hypothesis with `runtime investigation hypothesis register` and dispatches a real Investigator Assignment with `runtime investigation dispatch`. The Investigator saves the exact Hook-provided answer; the original Case consumer validates and records it.")
	fmt.Fprintln(w, "  3. Main follows `runtime investigation status`, routes the Case using the supported results and evidence, then uses the original RepairContract approval route. A human approval or bounded-repair grant must be real; a technical review is not a human decision.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "S9 complete repair and return")
	fmt.Fprintln(w, "  1. For an approved Contract, the Main listener uses the original readback to prepare the RepairSession and RepairPlan. Main dispatches each real RepairAssignment with `runtime repair dispatch`; the Hook supplies each Builder's exact PlanReport path, which saves to the original PlanReport consumer.")
	fmt.Fprintln(w, "  2. After the original graph advances to fixing, Builders save the exact Hook-provided RepairResult; the original Result and candidate consumers validate the answer and real changes. Main performs original integration and required checks.")
	fmt.Fprintln(w, "  3. Main saves the Hook-provided ChangeImpact answer for its original consumer. Main dispatches an independent verifier, which saves the exact TargetedReverification answer for its original consumer.")
	fmt.Fprintln(w, "  4. After targeted PASS, the next Main Hook runs the existing handoff consumer and prepares the S7 seed. Main dispatches the fresh complete S7 round; targeted PASS alone does not exit S9.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "S10/S11")
	fmt.Fprintln(w, "  S10: use `s10 status` for read-only progress. Fill the Hook-provided acceptance/audit manifest and save it once; the original consumer validates it, derives machine metadata, and records the evidence.")
	fmt.Fprintln(w, "  S11: fill and save the request at the exact pending release path supplied by the Hook. An agent may prepare the request but cannot decide or approve release; only the human's explicit decision uses `runtime human-decision`.")
	fmt.Fprintln(w, "  S9 never goes directly to S10; every product change returns to a fresh S7 review.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Diagnostics and compatibility")
	fmt.Fprintln(w, "  Read-only status: `s7 status`, `runtime investigation status`, `runtime repair status`, `s10 status`, `health`, `doctor`, `validate --all`.")
	fmt.Fprintln(w, "  Direct `runtime review-plan`, `runtime review-result submit`, `runtime investigation hypothesis result`, `runtime repair session open`, `runtime repair plan compile`, `s10 manifest validate/render`, `runtime evidence add`, and direct S9 create/submit/commit verbs are diagnostic, compatibility, or recovery interfaces—not a second delivery after saving a Hook-provided answer.")
	fmt.Fprintln(w, "  `runtime investigation consume` is used only when the current Case route explicitly requires that original route consumer. `runtime transition` is legacy/low-level; normal continuation follows the current stage instructions.")
	fmt.Fprintln(w, "  Help is non-mutating: `loop-harness s7 --help`, `s10 --help`, or `runtime repair dispatch --help` exits 0 and points back here.")
}

func compactHelpName(args []string) string {
	parts := make([]string, 0, len(args))
	for _, arg := range args {
		if arg == "--help" || arg == "-h" || arg == "help" {
			continue
		}
		parts = append(parts, arg)
	}
	return strings.Join(parts, " ")
}
