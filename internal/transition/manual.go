// Manual generator — renders `loop-harness manual` markdown and the per-
// transition `loop-harness explain <TR-xxx>` text from loop-definition.json +
// the guard spec registry. Output is the agent-facing checklist of what each
// transition checks.
//
// Format is deliberately minimal: one sentence per guard, no What/Fail/
// Recovery structure. The agent reads the list before calling `forward` and
// self-verifies each item. If a guard fails, the harness error names the
// guard ID; the agent re-reads that one line to understand the gap.
package transition

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/entroforge/go-system-builder/internal/evidence"
)

// ManualOptions controls markdown rendering.
type ManualOptions struct {
	// TargetPath is where the manual will live in the target project.
	TargetPath string
	// HarnessVersion labels the binary that produced this manual.
	HarnessVersion string
	// LoopDefinitionSHA256 pins the manual to the loop-definition.json it was
	// generated from. doctor compares this against the on-disk definition.
	LoopDefinitionSHA256 string
}

// RenderManual produces the full markdown manual.
func RenderManual(def *LoopDefinition, opts ManualOptions) string {
	if def == nil {
		return ""
	}
	var b strings.Builder
	writeHeader(&b, opts)
	writeControlPlaneRecovery(&b)
	writeS11HumanGateway(&b)
	writeTOC(&b, def)
	writeTopLevelTransitions(&b, def)
	writePhaseTransitions(&b, def)
	writeGlobalTransitions(&b, def)
	return b.String()
}

func writeS11HumanGateway(b *strings.Builder) {
	fmt.Fprintf(b, "## S11 Human decision gateway\n\n")
	fmt.Fprintf(b, "`awaiting_human_release` is a non-terminal human gateway. The Controller has no automatic candidate or decision at this cursor. Submit exactly one finite disposition with the explicit Runtime command:\n\n")
	fmt.Fprintf(b, "```bash\n")
	fmt.Fprintln(b, "loop-harness runtime human-decision \\")
	fmt.Fprintln(b, "  --disposition <approve|defer|reject_defect|reject_acceptance|reject_release_audit|abort> \\")
	fmt.Fprintln(b, "  --expected-revision <N> --actor <user|orchestrator> \\")
	fmt.Fprintln(b, "  --decision-evidence <human-decision-reference>")
	fmt.Fprintf(b, "```\n\n")
	fmt.Fprintf(b, "Disposition mapping is fixed: `approve` → TR-025 `release_authorized`; `defer` → TR-026 `paused` (the command binds generated `pause_record=generated:pause_checkpoint`); `reject_defect` → TR-027 S8 investigation (also requires `--finding-evidence`); `reject_acceptance` → TR-028 acceptance; `reject_release_audit` → TR-029 release audit; `abort` → TR-030 `aborted`. Arbitrary target states and transition IDs are not accepted.\n\n")
	fmt.Fprintf(b, "Human approval records release authorization only. Harness has no squash merge, publication, deployment, or formal release permission. Runtime rollover is eligible only from `release_authorized` or `aborted`.\n\n")
}

func writeControlPlaneRecovery(b *strings.Builder) {
	fmt.Fprintf(b, "## Controller recovery protocol\n\n")
	fmt.Fprintf(b, "The Hook is an event trigger for the Loop Controller, not only a guard. "+
		"On `SessionStart`, `PreCompact`, `SubagentStart`, `SubagentStop`, and `TeammateIdle`, "+
		"the Controller reads the Runtime, refreshes the resumable Milestone through CAS, "+
		"and emits a positive `LOOP RECOVERY` packet.\n\n")
	fmt.Fprintf(b, "1. Read the `Next` action and current `Stage` from the Hook packet, then follow its `Read in order` list.\n")
	fmt.Fprintf(b, "2. Read the linked `docs/agent-protocol.md#sN` section before acting.\n")
	fmt.Fprintf(b, "3. If blocked or the Runtime is unclear, read this Manual. Use `runtime reconcile` only when the Hook reports an integrity/CAS recovery condition; do not call `status`/`next` during normal continuation. When the live Quality Gate checklist is unclear, run `loop-harness ready` (diagnostics; never hand-push a Transition from it). `doctor` is schema/manual/policy_ref/metrics only — not stage readiness.\n")
	fmt.Fprintf(b, "4. Execute the one missing deliverable/evidence named by Hook/`ready` `missing[]`; do not invent a parallel lifecycle.\n")
	fmt.Fprintf(b, "5. For `SubagentStop`, complete the report, worktree review, merge-back to the current `develop` integration branch and `completion_ack` checklist before acknowledging the stop. For `TeammateIdle`, re-wake the same teammate. The identical integration chain is available explicitly via `runtime task-integrate --assignment-id <id>` when the automatic SubagentStop payload cannot identify the assignment.\n")
	fmt.Fprintf(b, "6. Builder completion is registered with `runtime task-complete` — one atomic command (message validation + evidence envelope derivation + Agent/TASK advance + evidence registration); the legacy `agent-event completion_reported` + `runtime evidence add` dual write still works but produces a thinner envelope. Before the Builder writes, create its worktree (`git worktree add .worktrees/<assignment-id> -b wt/<assignment-id> develop`) and record `worktree_path`/`branch`/`target_branch` on the manifest row — SubagentStop integration requires them.\n")
	fmt.Fprintf(b, "7. In S7 (verification) the round is driven by runtime verbs, not by hand-pushed transitions: scaffold and register the ReviewPlan (`s7 draft`, `runtime review-plan --file <plan.json>`), dispatch reviewers (`s7 manifest-draft`, `runtime register-workgroup`), consume each Assignment's Canonical ReviewResult (`runtime review-result submit --assignment-id <id> --result <result.json>`; observation steps land via `capture step` / `--captures`), and revise a running plan once per round (`runtime review-plan revise`). `loop-harness s7 status` is the board: it prints the plan line (with the `subject_digest` every result must bind), claim dispositions, and the single next action; a cold-start E2E round also has `loop-harness s7 workspace-digest` for the `verification_artifact_digest` the E2E result must bind. The machine exits are automatic — a sealed ObservationBatch or a machine CleanRound commits TR-008/TR-009 on the next PreToolUse; do not invoke the transition CLI for them. If the PostToolUse auto-activation chain did not fire for a dispatched Worker, recover with `runtime agent-begin --agent-id <id> --plan <plan-report.json>`. Full verb walkthrough: `docs/agent-protocol.md#s7`.\n")
	fmt.Fprintf(b, "8. Stop only at a human Gateway, an external asynchronous wait, or the end of the current turn.\n\n")
	fmt.Fprintf(b, "The persisted `.claude/loop-state.json` `milestone` is a recovery cache, not a second state machine. "+
		"`docs/loop-definition.json` and the Transition Engine remain the authority for legal lifecycle changes.\n\n")
	fmt.Fprintf(b, "During BUG investigation, answer why E2E did not cover or fail the gap "+
		"(`skills/bug-resolution/SKILL.md`; `loop-harness e2e-coverage`). A contracted behavior that broke without a red CT/AC requires a coverage-gap Closing Contract item.\n\n")
	fmt.Fprintf(b, "---\n\n")
}

// RenderTransition renders a single transition for `explain`. Returns empty
// string if no transition matches.
func RenderTransition(def *LoopDefinition, id string) string {
	if def == nil {
		return ""
	}
	spec, ok := findTransitionSpec(def, id)
	if !ok {
		return ""
	}
	var b strings.Builder
	writeTransition(&b, spec)
	return b.String()
}

// internal helpers -------------------------------------------------------

func writeHeader(b *strings.Builder, opts ManualOptions) {
	target := opts.TargetPath
	if target == "" {
		target = ".claude/bin/loop-harness.md"
	}
	fmt.Fprintf(b, "# loop-harness — Transition Checklist\n\n")
	fmt.Fprintf(b, "> What `loop-harness` checks at every transition. Read the\n")
	fmt.Fprintf(b, "> relevant section before calling `forward`; verify each bullet\n")
	fmt.Fprintf(b, "> before requesting the harness to advance.\n\n")
	fmt.Fprintf(b, "- **Path**: `%s`\n", target)
	if opts.HarnessVersion != "" {
		fmt.Fprintf(b, "- **Harness version**: %s\n", opts.HarnessVersion)
	}
	if opts.LoopDefinitionSHA256 != "" {
		fmt.Fprintf(b, "- **Loop definition SHA-256**: `%s`\n", opts.LoopDefinitionSHA256)
	}
	fmt.Fprintf(b, "\n---\n\n")
}

func writeTOC(b *strings.Builder, def *LoopDefinition) {
	fmt.Fprintf(b, "## Contents\n\n")
	for _, t := range def.Transitions {
		fmt.Fprintf(b, "- [`%s`](#%s) %s → %s — %s\n", t.ID, anchor(t.ID), t.From, t.To, oneLineDescription(t.Description))
	}
	phaseOwners := make([]string, 0, len(def.PhaseMachines))
	for owner := range def.PhaseMachines {
		phaseOwners = append(phaseOwners, owner)
	}
	sort.Strings(phaseOwners)
	for _, owner := range phaseOwners {
		machine := def.PhaseMachines[owner]
		if len(machine.Transitions) == 0 {
			continue
		}
		fmt.Fprintf(b, "\n_Phase: %s_\n\n", owner)
		for _, t := range machine.Transitions {
			fmt.Fprintf(b, "- [`%s`](#%s) %s → %s — %s\n", t.ID, anchor(t.ID), t.From, t.To, oneLineDescription(t.Description))
		}
	}
	if len(def.GlobalTransitions) > 0 {
		fmt.Fprintf(b, "\n_Global_\n\n")
		for _, t := range def.GlobalTransitions {
			fmt.Fprintf(b, "- [`%s`](#%s) → %s — %s\n", t.ID, anchor(t.ID), t.To, oneLineDescription(t.Description))
		}
	}
	fmt.Fprintf(b, "\n---\n\n")
}

func writeTopLevelTransitions(b *strings.Builder, def *LoopDefinition) {
	fmt.Fprintf(b, "## Transitions\n\n")
	for _, t := range def.Transitions {
		writeTransition(b, t)
	}
}

func writePhaseTransitions(b *strings.Builder, def *LoopDefinition) {
	phaseOwners := make([]string, 0, len(def.PhaseMachines))
	for owner := range def.PhaseMachines {
		phaseOwners = append(phaseOwners, owner)
	}
	sort.Strings(phaseOwners)
	for _, owner := range phaseOwners {
		machine := def.PhaseMachines[owner]
		if len(machine.Transitions) == 0 {
			continue
		}
		fmt.Fprintf(b, "## Phase transitions: %s\n\n", owner)
		for _, t := range machine.Transitions {
			writeTransition(b, t)
		}
	}
}

func writeGlobalTransitions(b *strings.Builder, def *LoopDefinition) {
	if len(def.GlobalTransitions) == 0 {
		return
	}
	fmt.Fprintf(b, "## Global transitions\n\n")
	for _, t := range def.GlobalTransitions {
		from := strings.Join(t.FromStates, "|")
		spec := TransitionSpec{
			ID:               t.ID,
			From:             from,
			Event:            t.Event,
			To:               t.To,
			Actors:           t.Actors,
			Guards:           t.Guards,
			RequiredEvidence: t.RequiredEvidence,
			Description:      t.Description,
		}
		writeTransition(b, spec)
	}
}

func writeTransition(b *strings.Builder, t TransitionSpec) {
	fmt.Fprintf(b, "### `%s` {#%s}\n\n", t.ID, anchor(t.ID))
	fmt.Fprintf(b, "_%s → %s_\n\n", t.From, t.To)
	fmt.Fprintf(b, "%s\n\n", t.Description)
	if len(t.Guards) == 0 {
		fmt.Fprintf(b, "_No guards._\n\n")
	} else {
		for _, g := range t.Guards {
			writeGuard(b, g)
		}
		fmt.Fprintf(b, "\n")
	}
	if len(t.RequiredEvidence) > 0 {
		fmt.Fprintf(b, "Evidence: %s\n\n", joinBack(t.RequiredEvidence))
		writeEvidenceBindings(b, t)
	}
}

func writeEvidenceBindings(b *strings.Builder, t TransitionSpec) {
	catalog := evidence.DefaultCatalog()
	fmt.Fprintf(b, "Evidence bindings (copy into `runtime transition`):\n\n")
	for _, slot := range t.RequiredEvidence {
		if generator, generated := catalog.Generator(slot); generated {
			fmt.Fprintf(b, "- `%s`: `--evidence %s=%s` (%s)\n", slot, slot, generator.Reference, generator.Description)
		} else {
			fmt.Fprintf(b, "- `%s`: `--evidence %s=<reference>`\n", slot, slot)
		}
		fmt.Fprintf(b, "  Accepted kinds: %s\n", joinBack(catalog.RegisteredAcceptedKinds(slot)))
	}
	fmt.Fprintf(b, "\nIf a binding is missing, retry with the command above; run `loop-harness explain %s` to inspect current candidates.\n\n", t.ID)
}

// ValidateManualEvidenceBindings verifies that a generated Manual contains an
// actionable binding template for every required evidence slot.
func ValidateManualEvidenceBindings(def *LoopDefinition, markdown string) error {
	if def == nil {
		return nil
	}
	for _, spec := range allTransitionSpecs(def) {
		for _, slot := range spec.RequiredEvidence {
			prefix := "--evidence " + slot + "="
			if !strings.Contains(markdown, prefix) {
				return fmt.Errorf("manual missing evidence binding guidance for transition %s slot %s; run `loop-harness manual --root .` to regenerate; expected %s<reference>", spec.ID, slot, prefix)
			}
		}
	}
	return nil
}

func allTransitionSpecs(def *LoopDefinition) []TransitionSpec {
	var specs []TransitionSpec
	specs = append(specs, def.Transitions...)
	owners := make([]string, 0, len(def.PhaseMachines))
	for owner := range def.PhaseMachines {
		owners = append(owners, owner)
	}
	sort.Strings(owners)
	for _, owner := range owners {
		specs = append(specs, def.PhaseMachines[owner].Transitions...)
	}
	for _, global := range def.GlobalTransitions {
		specs = append(specs, TransitionSpec{
			ID:               global.ID,
			RequiredEvidence: global.RequiredEvidence,
		})
	}
	for _, lifecycle := range def.EntityLifecycles {
		specs = append(specs, lifecycle.Transitions...)
	}
	return specs
}

func writeGuard(b *strings.Builder, guardID string) {
	registration, registered := LookupGuardRegistration(guardID)
	enforcement := "unregistered"
	if registered {
		enforcement = string(registration.Enforcement)
	}
	spec, ok := LookupGuardSpec(guardID)
	if !ok {
		fmt.Fprintf(b, "- `%s` [%s] — _no spec_\n", guardID, enforcement)
		return
	}
	fmt.Fprintf(b, "- `%s` [%s] — %s\n", guardID, enforcement, spec.Check)
}

// findTransitionSpec returns the transition spec by ID across all scopes.
func findTransitionSpec(def *LoopDefinition, id string) (TransitionSpec, bool) {
	for _, t := range def.Transitions {
		if t.ID == id {
			return t, true
		}
	}
	for _, machine := range def.PhaseMachines {
		for _, t := range machine.Transitions {
			if t.ID == id {
				return t, true
			}
		}
	}
	for _, g := range def.GlobalTransitions {
		if g.ID == id {
			return TransitionSpec{
				ID:               g.ID,
				From:             strings.Join(g.FromStates, "|"),
				Event:            g.Event,
				To:               g.To,
				Actors:           g.Actors,
				Guards:           g.Guards,
				RequiredEvidence: g.RequiredEvidence,
				Description:      g.Description,
			}, true
		}
	}
	return TransitionSpec{}, false
}

func anchor(id string) string { return strings.ToLower(id) }

func joinBack(values []string) string {
	if len(values) == 0 {
		return "_(none)_"
	}
	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = "`" + v + "`"
	}
	return strings.Join(parts, ", ")
}

func oneLineDescription(s string) string {
	if idx := strings.IndexByte(s, '.'); idx >= 0 {
		return s[:idx+1]
	}
	return s
}

// ManualFilename returns the conventional on-disk filename for the manual.
func ManualFilename() string { return "loop-harness.md" }

// ManualTargetPath returns the conventional target-project path.
func ManualTargetPath() string {
	return filepath.ToSlash(filepath.Join(".claude", "bin", ManualFilename()))
}
