# loop-harness — Transition Checklist

> What `loop-harness` checks at every transition. Read the
> relevant section before calling `forward`; verify each bullet
> before requesting the harness to advance.

- **Path**: `loop-harness.md`
- **Harness version**: dev
- **Loop definition SHA-256**: `3fa7c1ee8b2fb648ce8254c4d3d2fae62d1049c12b202dda19340a2af07cdb61`

---

## Controller recovery protocol

The Hook is an event trigger for the Loop Controller, not only a guard. On `SessionStart`, `PreCompact`, `SubagentStart`, `SubagentStop`, and `TeammateIdle`, the Controller reads the Runtime, refreshes the resumable Milestone through CAS, and emits a positive `LOOP RECOVERY` packet.

1. Read the `Next` action and current `Stage` from the Hook packet, then follow its `Read in order` list.
2. Read the linked `docs/agent-protocol.md#sN` section before acting.
3. If blocked or the Runtime is unclear, read this Manual. Use `runtime reconcile` only when the Hook reports an integrity/CAS recovery condition; do not call `status`/`next` during normal continuation. When the live Quality Gate checklist is unclear, run `loop-harness ready` (diagnostics; never hand-push a Transition from it). `doctor` is schema/manual/policy_ref/metrics only — not stage readiness.
4. Execute the one missing deliverable/evidence named by Hook/`ready` `missing[]`; do not invent a parallel lifecycle.
5. For `SubagentStop`, complete the report, worktree review, merge-back to the current `develop` integration branch and `completion_ack` checklist before acknowledging the stop. For `TeammateIdle`, re-wake the same teammate.
6. Stop only at a human Gateway, an external asynchronous wait, or the end of the current turn.

The persisted `.claude/loop-state.json` `milestone` is a recovery cache, not a second state machine. `docs/loop-definition.json` and the Transition Engine remain the authority for legal lifecycle changes.

During BUG investigation, answer why E2E did not cover or fail the gap (`skills/bug-resolution/SKILL.md`; `loop-harness e2e-coverage`). A contracted behavior that broke without a red CT/AC requires a coverage-gap Closing Contract item.

---

## S11 Human decision gateway

`awaiting_human_release` is a non-terminal human gateway. The Controller has no automatic candidate or decision at this cursor. Submit exactly one finite disposition with the explicit Runtime command:

```bash
loop-harness runtime human-decision \
  --disposition <approve|defer|reject_defect|reject_acceptance|reject_release_audit|abort> \
  --expected-revision <N> --actor <user|orchestrator> \
  --decision-evidence <human-decision-reference>
```

Disposition mapping is fixed: `approve` → TR-025 `release_authorized`; `defer` → TR-026 `paused` (the command binds generated `pause_record=generated:pause_checkpoint`); `reject_defect` → TR-027 S8 investigation (also requires `--finding-evidence`); `reject_acceptance` → TR-028 acceptance; `reject_release_audit` → TR-029 release audit; `abort` → TR-030 `aborted`. Arbitrary target states and transition IDs are not accepted.

Human approval records release authorization only. Harness has no squash merge, publication, deployment, or formal release permission. Runtime rollover is eligible only from `release_authorized` or `aborted`.

## Contents

- [`TR-001`](#tr-001) inactive → planning — Start exactly one Loop for the named locked REQ.
- [`TR-002`](#tr-002) planning → document_verification — Planning completes when at least one CONTRACTS-*.
- [`TR-003`](#tr-003) document_verification → building — Lock only the exact contract and task versions jointly verified.
- [`TR-004`](#tr-004) document_verification → planning — Non-REQ document findings return to planning.
- [`TR-005`](#tr-005) document_verification → paused — REQ changes always return control to the human.
- [`TR-006`](#tr-006) building → verification — Start a complete review round after all activated Builders report.
- [`TR-007`](#tr-007) building → planning — Non-REQ execution conflicts return through planning and document verification.
- [`TR-008`](#tr-008) verification → bug_resolution — Any blocking Delivery, QA, or E2E Browser finding means the REQ batch is incomplete and must enter S8 investigation before repair.
- [`TR-009`](#tr-009) verification → acceptance — Acceptance requires one complete, current and blocker-free review round.
- [`TR-010`](#tr-010) verification → paused — A required REQ change pauses automation.
- [`TR-011`](#tr-011) verification → paused — Security, compliance or equivalent blockers require human intervention.
- [`TR-012`](#tr-012) bug_resolution → verification — Only the ready_for_full_review handoff checkpoint may enter a complete Delivery + QA + E2E Browser round; targeted re-verification never substitutes for that round.
- [`TR-013`](#tr-013) bug_resolution → planning — Repair-driven specification changes return through planning and document verification.
- [`TR-014`](#tr-014) bug_resolution → paused — Repair work cannot modify the locked REQ.
- [`TR-015`](#tr-015) acceptance → release_audit — Release audit starts only from current ACC and clean-round evidence.
- [`TR-016`](#tr-016) acceptance → verification — Acceptance discrepancies restart the complete review.
- [`TR-017`](#tr-017) release_audit → awaiting_human_release — Approved or approved-with-risk audit reaches the human release boundary.
- [`TR-018`](#tr-018) release_audit → paused — A blocked release audit pauses the Loop.
- [`TR-019`](#tr-019) paused → RESUME_FROM_PAUSE — Resume the exact validated state, phase and entity checkpoint.
- [`TR-020`](#tr-020) paused → planning — A changed locked REQ starts a new planning generation.
- [`TR-021`](#tr-021) paused → aborted — Only a human may permanently abort the Loop.
- [`TR-022`](#tr-022) bug_resolution → verification — When the completed S8 disposition batch contains no accepted BUG, every finding is finally rejected without product/specification change or duplicate-linked to a canonical BUG with no remaining repair, the Loop returns to verification for a fresh complete round.
- [`TR-023`](#tr-023) bug_resolution → planning — Finding-level specification rework (S8) routes back to planning.
- [`TR-024`](#tr-024) bug_resolution → paused — A finding that requires modifying the locked REQ cannot proceed autonomously; the Loop pauses for human amendment.
- [`TR-025`](#tr-025) awaiting_human_release → release_authorized — Record human release authorization without performing merge, publication, deployment, or formal release.
- [`TR-026`](#tr-026) awaiting_human_release → paused — Defer the release decision and capture the S11 checkpoint before entering paused.
- [`TR-027`](#tr-027) awaiting_human_release → bug_resolution — Route a human defect rejection to bug-resolution investigation with its finding evidence.
- [`TR-028`](#tr-028) awaiting_human_release → acceptance — Re-enter acceptance and invalidate only prior acceptance and release-audit evidence.
- [`TR-029`](#tr-029) awaiting_human_release → release_audit — Re-enter release audit and invalidate only prior release-audit evidence.
- [`TR-030`](#tr-030) awaiting_human_release → aborted — Record a human release abort without performing any release side effect.

_Phase: bug_resolution_

- [`PTR-BUG-01`](#ptr-bug-01) investigation → bug_report_review — A finding cannot become repair work before root-cause evidence exists.
- [`PTR-BUG-02`](#ptr-bug-02) bug_report_review → repair_readback — The orchestrator approves canonical BUGs and duplicate mappings.
- [`PTR-BUG-03`](#ptr-bug-03) bug_report_review → investigation — Insufficient BUG reports return to investigation.
- [`PTR-BUG-04`](#ptr-bug-04) repair_readback → fixing — Repair execution uses the same two-phase activation gate.
- [`PTR-BUG-05`](#ptr-bug-05) fixing → targeted_reverification — Every repair invalidates affected historical PASS evidence before recheck.
- [`PTR-BUG-06`](#ptr-bug-06) targeted_reverification → ready_for_full_review — A targeted pass never substitutes for the full review round; it records the S9-to-S7 handoff checkpoint.
- [`PTR-BUG-07`](#ptr-bug-07) targeted_reverification → investigation — Failed repair verification restarts root-cause investigation.

_Phase: planning_

- [`PTR-PLAN-01`](#ptr-plan-01) design → contracts — Advance formal planning from design to contracts after the design quality gate passes.
- [`PTR-PLAN-02`](#ptr-plan-02) contracts → tasks — Advance formal planning from contracts to tasks after the contract quality gate passes.

_Phase: verification_

- [`PTR-VERIFY-01`](#ptr-verify-01) delivery → qa — QA cannot start until all required delivery dimensions pass.
- [`PTR-VERIFY-02`](#ptr-verify-02) qa → e2e_browser — Real-browser E2E follows complete QA evidence.
- [`PTR-VERIFY-03`](#ptr-verify-03) e2e_browser → clean_round_evaluation — Clean-round evaluation follows complete real-browser E2E evidence.
- [`PTR-VERIFY-04`](#ptr-verify-04) clean_round_evaluation → clean_round_passed — Only one complete, current and blocker-free round passes.
- [`PTR-VERIFY-05`](#ptr-verify-05) clean_round_evaluation → delivery — Mixed, incomplete or stale evidence restarts the full review.

_Global_

- [`GTR-001`](#gtr-001) → paused — A user may pause from any active state.
- [`GTR-002`](#gtr-002) → paused — Any required locked REQ change pauses automation.
- [`GTR-003`](#gtr-003) → paused — Production-data, security, compliance or irreversible actions require human approval.
- [`GTR-004`](#gtr-004) → paused — Configured repair limits never silently close a BUG.
- [`GTR-005`](#gtr-005) → paused — Runtime/document inconsistency fails closed.

---

## Transitions

### `TR-001` {#tr-001}

_inactive → planning_

Start exactly one Loop for the named locked REQ.

- `req_exists` [evidence_attestation] — A REQ file exists at the path supplied to `loop-harness req bind`.
- `req_locked` [evidence_attestation] — The REQ file declares status=locked and a non-empty version field in its frontmatter.
- `req_questions_non_blocking` [evidence_attestation] — Any open questions recorded in the REQ's 待澄清问题 section are tagged non-blocking, or none exist.
- `pm_context_matches_req` [evidence_attestation] — docs/project-map.md current stage and bound REQ references are consistent with the REQ being bound.
- `no_other_active_loop` [semantic_check] — No other runtime is currently bound to a different REQ; the Loop model allows one active REQ per project.

Evidence: `req_lock_record`, `loop_authorization_record`

Evidence bindings (copy into `runtime transition`):

- `req_lock_record`: `--evidence req_lock_record=<reference>`
  Accepted kinds: `human_decision`
- `loop_authorization_record`: `--evidence loop_authorization_record=<reference>`
  Accepted kinds: `human_decision`

If a binding is missing, retry with the command above; run `loop-harness explain TR-001` to inspect current candidates.

### `TR-002` {#tr-002}

_planning → document_verification_

Planning completes when at least one CONTRACTS-*.md is locked AND at least one TASK-*.md is complete. Direct file+status check via the planning_complete guard.

- `planning_complete` [semantic_check] — At least one current-baseline contract document has status=locked and a matching on-disk markdown status, AND at least one current-baseline task document has status=complete with a matching on-disk status (aligned with GATE-PLANNING-TASKS-COMPLETE). Falls back to CONTRACTS-*.md / TASK-*.md filename patterns when runtime documents are absent.

### `TR-003` {#tr-003}

_document_verification → building_

Lock only the exact contract and task versions jointly verified.

- `joint_document_pass` [evidence_attestation] — _no spec_
- `verified_versions_current` [evidence_attestation] — Every document the joint review verified (REQ, contracts, tasks, design) still matches the fingerprints captured when the pass verdict was recorded, so the lock cannot advance on drifted inputs.

Evidence: `document_review_record`, `contract_set_record`, `task_batch_record`

Evidence bindings (copy into `runtime transition`):

- `document_review_record`: `--evidence document_review_record=<reference>`
  Accepted kinds: `document_review`
- `contract_set_record`: `--evidence contract_set_record=<reference>`
  Accepted kinds: `document_review`
- `task_batch_record`: `--evidence task_batch_record=<reference>`
  Accepted kinds: `document_review`

If a binding is missing, retry with the command above; run `loop-harness explain TR-003` to inspect current candidates.

### `TR-004` {#tr-004}

_document_verification → planning_

Non-REQ document findings return to planning.design for reworking the failing artifacts.

- `req_baseline_unchanged` [evidence_attestation] — The locked REQ's sha256 recorded in runtime.bound_req still matches the file at runtime.bound_req.path, so the rework loop cannot silently advance on a changed REQ.

Evidence: `document_review_record`

Evidence bindings (copy into `runtime transition`):

- `document_review_record`: `--evidence document_review_record=<reference>`
  Accepted kinds: `document_review`

If a binding is missing, retry with the command above; run `loop-harness explain TR-004` to inspect current candidates.

### `TR-005` {#tr-005}

_document_verification → paused_

REQ changes always return control to the human.

_No guards._

Evidence: `document_review_record`, `pause_record`

Evidence bindings (copy into `runtime transition`):

- `document_review_record`: `--evidence document_review_record=<reference>`
  Accepted kinds: `document_review`
- `pause_record`: `--evidence pause_record=generated:pause_checkpoint` (generated pause checkpoint)
  Accepted kinds: `human_decision`

If a binding is missing, retry with the command above; run `loop-harness explain TR-005` to inspect current candidates.

### `TR-006` {#tr-006}

_building → verification_

Start a complete review round after all activated Builders report.

- `all_builder_tasks_in_review` [evidence_attestation] — Every TASK assignment in the current Builder workgroup has reached at least the `reported` state in runtime.entities.tasks[].state.
- `builder_reports_complete` [evidence_attestation] — Each Builder assignment's runtime record references a valid completion_report evidence item with a matching fingerprint.
- `verification_team_manifest_complete` [evidence_attestation] — A Delivery Verifier team manifest is registered in runtime.entities.teams[] with all mandatory responsibilities (VER-REQ-GAP, VER-SPEC-GAP, VER-MODULE-COMPLETE) plus any risk-triggered responsibilities.

Evidence: `builder_report_record`, `team_manifest_record`

Evidence bindings (copy into `runtime transition`):

- `builder_report_record`: `--evidence builder_report_record=<reference>`
  Accepted kinds: `builder_report`, `agent_completion`
- `team_manifest_record`: `--evidence team_manifest_record=<reference>`
  Accepted kinds: `builder_report`, `team_manifest`

If a binding is missing, retry with the command above; run `loop-harness explain TR-006` to inspect current candidates.

### `TR-007` {#tr-007}

_building → planning_

Non-REQ execution conflicts return through planning and document verification.

- `req_baseline_unchanged` [evidence_attestation] — The locked REQ's sha256 recorded in runtime.bound_req still matches the file at runtime.bound_req.path, so the rework loop cannot silently advance on a changed REQ.

Evidence: `change_impact_record`

Evidence bindings (copy into `runtime transition`):

- `change_impact_record`: `--evidence change_impact_record=<reference>`
  Accepted kinds: `change_impact`

If a binding is missing, retry with the command above; run `loop-harness explain TR-007` to inspect current candidates.

### `TR-008` {#tr-008}

_verification → bug_resolution_

Any blocking Delivery, QA, or E2E Browser finding means the REQ batch is incomplete and must enter S8 investigation before repair.

- `blocking_findings_present` [evidence_attestation] — At least one finding evidence item with severity=blocking exists for the current review round, referencing a canonical BUG.

Evidence: `finding_record`

Evidence bindings (copy into `runtime transition`):

- `finding_record`: `--evidence finding_record=<reference>`
  Accepted kinds: `bug`

If a binding is missing, retry with the command above; run `loop-harness explain TR-008` to inspect current candidates.

### `TR-009` {#tr-009}

_verification → acceptance_

Acceptance requires one complete, current and blocker-free review round.

- `verification_phase_clean_round_passed` [semantic_check] — verification.EvaluateCleanRound reports that runtime.review.clean_round equals runtime.review.round, i.e. a complete clean round has just been recorded for the active round.
- `same_review_round` [semantic_check] — verification.EvaluateCleanRound reports that every `pass` evidence item under consideration was recorded against runtime.review.round, so the round cannot close on stale evidence from a prior round.
- `no_invalidated_pass_evidence` [semantic_check] — verification.EvaluateCleanRound reports no `pass` evidence in runtime.evidence[] carries an `invalidated_at` stamp for the active review round.
- `no_open_blocking_bugs` [semantic_check] — No BUG with severity=P0 in runtime.entities.bugs[] remains in an open state (`investigating`, `accepted`, `assigned`, `fixing`, `retesting`) for the current review round.

Evidence: `clean_round_record`

Evidence bindings (copy into `runtime transition`):

- `clean_round_record`: `--evidence clean_round_record=<reference>`
  Accepted kinds: `clean_round`

If a binding is missing, retry with the command above; run `loop-harness explain TR-009` to inspect current candidates.

### `TR-010` {#tr-010}

_verification → paused_

A required REQ change pauses automation.

_No guards._

Evidence: `review_result_record`, `pause_record`

Evidence bindings (copy into `runtime transition`):

- `review_result_record`: `--evidence review_result_record=<reference>`
  Accepted kinds: `delivery_review`, `qa_review`, `e2e_review`
- `pause_record`: `--evidence pause_record=generated:pause_checkpoint` (generated pause checkpoint)
  Accepted kinds: `human_decision`

If a binding is missing, retry with the command above; run `loop-harness explain TR-010` to inspect current candidates.

### `TR-011` {#tr-011}

_verification → paused_

Security, compliance or equivalent blockers require human intervention.

_No guards._

Evidence: `review_result_record`, `pause_record`

Evidence bindings (copy into `runtime transition`):

- `review_result_record`: `--evidence review_result_record=<reference>`
  Accepted kinds: `delivery_review`, `qa_review`, `e2e_review`
- `pause_record`: `--evidence pause_record=generated:pause_checkpoint` (generated pause checkpoint)
  Accepted kinds: `human_decision`

If a binding is missing, retry with the command above; run `loop-harness explain TR-011` to inspect current candidates.

### `TR-012` {#tr-012}

_bug_resolution → verification_

Only the ready_for_full_review handoff checkpoint may enter a complete Delivery + QA + E2E Browser round; targeted re-verification never substitutes for that round.

- `bug_phase_ready_for_full_review` [evidence_attestation] — Every BUG from the round has reached `bug_phase_ready_for_full_review`, i.e. targeted re-verification is complete and the bug sub-machine is ready to fold back into the main verification review.
- `all_targeted_reverification_passed` [semantic_check] — Every P0 BUG in runtime.entities.bugs[] has advanced past `retesting`/`fixing`/`investigating` so no blocking bug remains awaiting targeted re-verification.

Evidence: `targeted_reverification_record`, `change_impact_record`

Evidence bindings (copy into `runtime transition`):

- `targeted_reverification_record`: `--evidence targeted_reverification_record=<reference>`
  Accepted kinds: `targeted_reverification`
- `change_impact_record`: `--evidence change_impact_record=<reference>`
  Accepted kinds: `change_impact`

If a binding is missing, retry with the command above; run `loop-harness explain TR-012` to inspect current candidates.

### `TR-013` {#tr-013}

_bug_resolution → planning_

Repair-driven specification changes return through planning and document verification.

- `req_baseline_unchanged` [evidence_attestation] — The locked REQ's sha256 recorded in runtime.bound_req still matches the file at runtime.bound_req.path, so the rework loop cannot silently advance on a changed REQ.

Evidence: `change_impact_record`, `repair_record`

Evidence bindings (copy into `runtime transition`):

- `change_impact_record`: `--evidence change_impact_record=<reference>`
  Accepted kinds: `change_impact`
- `repair_record`: `--evidence repair_record=<reference>`
  Accepted kinds: `bug`

If a binding is missing, retry with the command above; run `loop-harness explain TR-013` to inspect current candidates.

### `TR-014` {#tr-014}

_bug_resolution → paused_

Repair work cannot modify the locked REQ.

_No guards._

Evidence: `repair_record`, `pause_record`

Evidence bindings (copy into `runtime transition`):

- `repair_record`: `--evidence repair_record=<reference>`
  Accepted kinds: `bug`
- `pause_record`: `--evidence pause_record=generated:pause_checkpoint` (generated pause checkpoint)
  Accepted kinds: `human_decision`

If a binding is missing, retry with the command above; run `loop-harness explain TR-014` to inspect current candidates.

### `TR-015` {#tr-015}

_acceptance → release_audit_

Release audit starts only from current ACC and clean-round evidence.

- `acc_complete` [evidence_attestation] — An acceptance evidence item recorded for the current review round is referenced from runtime.evidence[] and its fingerprint matches the on-disk acceptance record.
- `clean_round_still_valid` [semantic_check] — verification.EvaluateCleanRound still reports a passing clean round at the current baseline generation, so neither baselines nor evidence drifted between round capture and release.

Evidence: `acceptance_record`, `clean_round_record`

Evidence bindings (copy into `runtime transition`):

- `acceptance_record`: `--evidence acceptance_record=<reference>`
  Accepted kinds: `acceptance`
- `clean_round_record`: `--evidence clean_round_record=<reference>`
  Accepted kinds: `clean_round`

If a binding is missing, retry with the command above; run `loop-harness explain TR-015` to inspect current candidates.

### `TR-016` {#tr-016}

_acceptance → verification_

Acceptance discrepancies restart the complete review.

_No guards._

Evidence: `acceptance_record`, `change_impact_record`

Evidence bindings (copy into `runtime transition`):

- `acceptance_record`: `--evidence acceptance_record=<reference>`
  Accepted kinds: `acceptance`
- `change_impact_record`: `--evidence change_impact_record=<reference>`
  Accepted kinds: `change_impact`

If a binding is missing, retry with the command above; run `loop-harness explain TR-016` to inspect current candidates.

### `TR-017` {#tr-017}

_release_audit → awaiting_human_release_

Approved or approved-with-risk audit reaches the human release boundary.

- `release_audit_approved` [evidence_attestation] — A release-audit approval evidence item referencing docs/release_audits/ and signed by an authorized actor is recorded in runtime.evidence[].
- `acc_complete` [evidence_attestation] — An acceptance evidence item recorded for the current review round is referenced from runtime.evidence[] and its fingerprint matches the on-disk acceptance record.
- `clean_round_still_valid` [semantic_check] — verification.EvaluateCleanRound still reports a passing clean round at the current baseline generation, so neither baselines nor evidence drifted between round capture and release.

Evidence: `release_audit_record`, `acceptance_record`

Evidence bindings (copy into `runtime transition`):

- `release_audit_record`: `--evidence release_audit_record=<reference>`
  Accepted kinds: `release_audit`
- `acceptance_record`: `--evidence acceptance_record=<reference>`
  Accepted kinds: `acceptance`

If a binding is missing, retry with the command above; run `loop-harness explain TR-017` to inspect current candidates.

### `TR-018` {#tr-018}

_release_audit → paused_

A blocked release audit pauses the Loop.

_No guards._

Evidence: `release_audit_record`, `pause_record`

Evidence bindings (copy into `runtime transition`):

- `release_audit_record`: `--evidence release_audit_record=<reference>`
  Accepted kinds: `release_audit`
- `pause_record`: `--evidence pause_record=generated:pause_checkpoint` (generated pause checkpoint)
  Accepted kinds: `human_decision`

If a binding is missing, retry with the command above; run `loop-harness explain TR-018` to inspect current candidates.

### `TR-019` {#tr-019}

_paused → RESUME_FROM_PAUSE_

Resume the exact validated state, phase and entity checkpoint.

- `resume_checkpoint_valid` [semantic_check] — A pause checkpoint exists at runtime.pause with a non-empty `reason` and `required_human_action`, so the resume transition has a captured state to restore.
- `baselines_unchanged` [evidence_attestation] — The baseline document fingerprints captured at pause time still match the on-disk files in docs/contracts/ and docs/tasks/ so the resume cannot quietly advance on drifted inputs.

Evidence: `human_decision_record`, `pause_record`

Evidence bindings (copy into `runtime transition`):

- `human_decision_record`: `--evidence human_decision_record=<reference>`
  Accepted kinds: `human_decision`
- `pause_record`: `--evidence pause_record=generated:pause_checkpoint` (generated pause checkpoint)
  Accepted kinds: `human_decision`

If a binding is missing, retry with the command above; run `loop-harness explain TR-019` to inspect current candidates.

### `TR-020` {#tr-020}

_paused → planning_

A changed locked REQ starts a new planning generation.

- `updated_req_locked` [evidence_attestation] — The updated REQ file referenced by the re-bind request declares status=locked with a strictly higher version than the currently bound REQ, and its sha256 matches the on-disk file.

Evidence: `human_decision_record`, `req_lock_record`

Evidence bindings (copy into `runtime transition`):

- `human_decision_record`: `--evidence human_decision_record=<reference>`
  Accepted kinds: `human_decision`
- `req_lock_record`: `--evidence req_lock_record=<reference>`
  Accepted kinds: `human_decision`

If a binding is missing, retry with the command above; run `loop-harness explain TR-020` to inspect current candidates.

### `TR-021` {#tr-021}

_paused → aborted_

Only a human may permanently abort the Loop.

- `human_abort_approved` [evidence_attestation] — A human-approval evidence item signed by an authorized actor is referenced from runtime.evidence[] permitting the runtime to move to the `aborted` state.

Evidence: `human_decision_record`

Evidence bindings (copy into `runtime transition`):

- `human_decision_record`: `--evidence human_decision_record=<reference>`
  Accepted kinds: `human_decision`

If a binding is missing, retry with the command above; run `loop-harness explain TR-021` to inspect current candidates.

### `TR-022` {#tr-022}

_bug_resolution → verification_

When the completed S8 disposition batch contains no accepted BUG, every finding is finally rejected without product/specification change or duplicate-linked to a canonical BUG with no remaining repair, the Loop returns to verification for a fresh complete round. An insufficient BUG report remains in investigation; spec and REQ changes use TR-023 and TR-024.

- `no_accepted_bugs` [evidence_attestation] — No BUG in runtime.entities.bugs[] for the current review round is in state `accepted`, `assigned`, `fixing`, or `retesting`. Used by TR-022 to confirm a finding-level Loop exit to verification is safe (no accepted BUG requires the S9 repair flow).
- `bug_report_review_complete` [evidence_attestation] — Every blocking S7 finding has a recorded disposition in runtime.evidence[] (accepted canonical BUG, rejected BUG, duplicate link, spec rework handoff, or REQ change pause). Used by TR-022 to confirm the orchestrator has classified every blocking finding before exiting the bug_resolution phase.

Evidence: `bug_batch_record`

Evidence bindings (copy into `runtime transition`):

- `bug_batch_record`: `--evidence bug_batch_record=<reference>`
  Accepted kinds: `bug`

If a binding is missing, retry with the command above; run `loop-harness explain TR-022` to inspect current candidates.

### `TR-023` {#tr-023}

_bug_resolution → planning_

Finding-level specification rework (S8) routes back to planning.design for a new TR-002 cycle; complements TR-013 which handles repair-level spec change.

- `req_baseline_unchanged` [evidence_attestation] — The locked REQ's sha256 recorded in runtime.bound_req still matches the file at runtime.bound_req.path, so the rework loop cannot silently advance on a changed REQ.

Evidence: `bug_batch_record`, `change_impact_record`

Evidence bindings (copy into `runtime transition`):

- `bug_batch_record`: `--evidence bug_batch_record=<reference>`
  Accepted kinds: `bug`
- `change_impact_record`: `--evidence change_impact_record=<reference>`
  Accepted kinds: `change_impact`

If a binding is missing, retry with the command above; run `loop-harness explain TR-023` to inspect current candidates.

### `TR-024` {#tr-024}

_bug_resolution → paused_

A finding that requires modifying the locked REQ cannot proceed autonomously; the Loop pauses for human amendment. Complements TR-014 which handles repair-level REQ change.

_No guards._

Evidence: `bug_batch_record`, `pause_record`

Evidence bindings (copy into `runtime transition`):

- `bug_batch_record`: `--evidence bug_batch_record=<reference>`
  Accepted kinds: `bug`
- `pause_record`: `--evidence pause_record=generated:pause_checkpoint` (generated pause checkpoint)
  Accepted kinds: `human_decision`

If a binding is missing, retry with the command above; run `loop-harness explain TR-024` to inspect current candidates.

### `TR-025` {#tr-025}

_awaiting_human_release → release_authorized_

Record human release authorization without performing merge, publication, deployment, or formal release.

_No guards._

Evidence: `human_decision_record`

Evidence bindings (copy into `runtime transition`):

- `human_decision_record`: `--evidence human_decision_record=<reference>`
  Accepted kinds: `human_decision`

If a binding is missing, retry with the command above; run `loop-harness explain TR-025` to inspect current candidates.

### `TR-026` {#tr-026}

_awaiting_human_release → paused_

Defer the release decision and capture the S11 checkpoint before entering paused.

_No guards._

Evidence: `human_decision_record`, `pause_record`

Evidence bindings (copy into `runtime transition`):

- `human_decision_record`: `--evidence human_decision_record=<reference>`
  Accepted kinds: `human_decision`
- `pause_record`: `--evidence pause_record=generated:pause_checkpoint` (generated pause checkpoint)
  Accepted kinds: `human_decision`

If a binding is missing, retry with the command above; run `loop-harness explain TR-026` to inspect current candidates.

### `TR-027` {#tr-027}

_awaiting_human_release → bug_resolution_

Route a human defect rejection to bug-resolution investigation with its finding evidence.

_No guards._

Evidence: `human_decision_record`, `finding_record`

Evidence bindings (copy into `runtime transition`):

- `human_decision_record`: `--evidence human_decision_record=<reference>`
  Accepted kinds: `human_decision`
- `finding_record`: `--evidence finding_record=<reference>`
  Accepted kinds: `bug`

If a binding is missing, retry with the command above; run `loop-harness explain TR-027` to inspect current candidates.

### `TR-028` {#tr-028}

_awaiting_human_release → acceptance_

Re-enter acceptance and invalidate only prior acceptance and release-audit evidence.

_No guards._

Evidence: `human_decision_record`

Evidence bindings (copy into `runtime transition`):

- `human_decision_record`: `--evidence human_decision_record=<reference>`
  Accepted kinds: `human_decision`

If a binding is missing, retry with the command above; run `loop-harness explain TR-028` to inspect current candidates.

### `TR-029` {#tr-029}

_awaiting_human_release → release_audit_

Re-enter release audit and invalidate only prior release-audit evidence.

_No guards._

Evidence: `human_decision_record`

Evidence bindings (copy into `runtime transition`):

- `human_decision_record`: `--evidence human_decision_record=<reference>`
  Accepted kinds: `human_decision`

If a binding is missing, retry with the command above; run `loop-harness explain TR-029` to inspect current candidates.

### `TR-030` {#tr-030}

_awaiting_human_release → aborted_

Record a human release abort without performing any release side effect.

- `human_abort_approved` [evidence_attestation] — A human-approval evidence item signed by an authorized actor is referenced from runtime.evidence[] permitting the runtime to move to the `aborted` state.

Evidence: `human_decision_record`

Evidence bindings (copy into `runtime transition`):

- `human_decision_record`: `--evidence human_decision_record=<reference>`
  Accepted kinds: `human_decision`

If a binding is missing, retry with the command above; run `loop-harness explain TR-030` to inspect current candidates.

## Phase transitions: bug_resolution

### `PTR-BUG-01` {#ptr-bug-01}

_investigation → bug_report_review_

A finding cannot become repair work before root-cause evidence exists.

- `root_cause_evidence_complete` [evidence_attestation] — A root-cause evidence item (failure mode, triggering input, and minimal-repro path) is referenced from runtime.evidence[] for the BUG promoted out of `investigating`.

Evidence: `finding_record`, `root_cause_record`

Evidence bindings (copy into `runtime transition`):

- `finding_record`: `--evidence finding_record=<reference>`
  Accepted kinds: `bug`
- `root_cause_record`: `--evidence root_cause_record=<reference>`
  Accepted kinds: `bug`

If a binding is missing, retry with the command above; run `loop-harness explain PTR-BUG-01` to inspect current candidates.

### `PTR-BUG-02` {#ptr-bug-02}

_bug_report_review → repair_readback_

The orchestrator approves canonical BUGs and duplicate mappings.

- `canonical_bug_mapping_complete` [evidence_attestation] — Each finding evidence item in the round maps to a canonical BUG id in runtime.entities.bugs[] so the verification phase cannot advance on unresolved duplicates.
- `bug_closing_contracts_complete` [evidence_attestation] — Every BUG raised in the current review round has its closing-contract evidence referenced from runtime.evidence[] before the verification phase can advance.

Evidence: `bug_batch_record`

Evidence bindings (copy into `runtime transition`):

- `bug_batch_record`: `--evidence bug_batch_record=<reference>`
  Accepted kinds: `bug`

If a binding is missing, retry with the command above; run `loop-harness explain PTR-BUG-02` to inspect current candidates.

### `PTR-BUG-03` {#ptr-bug-03}

_bug_report_review → investigation_

Insufficient BUG reports return to investigation.

_No guards._

Evidence: `bug_batch_record`

Evidence bindings (copy into `runtime transition`):

- `bug_batch_record`: `--evidence bug_batch_record=<reference>`
  Accepted kinds: `bug`

If a binding is missing, retry with the command above; run `loop-harness explain PTR-BUG-03` to inspect current candidates.

### `PTR-BUG-04` {#ptr-bug-04}

_repair_readback → fixing_

Repair execution uses the same two-phase activation gate.

- `repair_understanding_approved` [evidence_attestation] — An understanding-approval evidence item is referenced from runtime.evidence[] confirming the assigned Builder understood the root-cause writeup before activation.
- `repair_activation_recorded` [evidence_attestation] — A repair-activation evidence item is referenced from runtime.evidence[] confirming the assigned Builder has started the repair task recorded in runtime.entities.tasks[].

Evidence: `activation_record`

Evidence bindings (copy into `runtime transition`):

- `activation_record`: `--evidence activation_record=<reference>`
  Accepted kinds: `agent_activation`

If a binding is missing, retry with the command above; run `loop-harness explain PTR-BUG-04` to inspect current candidates.

### `PTR-BUG-05` {#ptr-bug-05}

_fixing → targeted_reverification_

Every repair invalidates affected historical PASS evidence before recheck.

- `repair_reports_complete` [evidence_attestation] — Every repair task spawned for BUGs in the round has a completion_report evidence item referenced from runtime.evidence[].

Evidence: `repair_record`, `change_impact_record`

Evidence bindings (copy into `runtime transition`):

- `repair_record`: `--evidence repair_record=<reference>`
  Accepted kinds: `bug`
- `change_impact_record`: `--evidence change_impact_record=<reference>`
  Accepted kinds: `change_impact`

If a binding is missing, retry with the command above; run `loop-harness explain PTR-BUG-05` to inspect current candidates.

### `PTR-BUG-06` {#ptr-bug-06}

_targeted_reverification → ready_for_full_review_

A targeted pass never substitutes for the full review round; it records the S9-to-S7 handoff checkpoint.

- `original_finder_reverification_complete` [evidence_attestation] — An original-finder re-verification evidence item is referenced from runtime.evidence[] confirming the agent that raised the BUG has re-tested the fix.

Evidence: `targeted_reverification_record`

Evidence bindings (copy into `runtime transition`):

- `targeted_reverification_record`: `--evidence targeted_reverification_record=<reference>`
  Accepted kinds: `targeted_reverification`

If a binding is missing, retry with the command above; run `loop-harness explain PTR-BUG-06` to inspect current candidates.

### `PTR-BUG-07` {#ptr-bug-07}

_targeted_reverification → investigation_

Failed repair verification restarts root-cause investigation.

_No guards._

Evidence: `targeted_reverification_record`

Evidence bindings (copy into `runtime transition`):

- `targeted_reverification_record`: `--evidence targeted_reverification_record=<reference>`
  Accepted kinds: `targeted_reverification`

If a binding is missing, retry with the command above; run `loop-harness explain PTR-BUG-07` to inspect current candidates.

## Phase transitions: planning

### `PTR-PLAN-01` {#ptr-plan-01}

_design → contracts_

Advance formal planning from design to contracts after the design quality gate passes.

_No guards._

### `PTR-PLAN-02` {#ptr-plan-02}

_contracts → tasks_

Advance formal planning from contracts to tasks after the contract quality gate passes.

_No guards._

## Phase transitions: verification

### `PTR-VERIFY-01` {#ptr-verify-01}

_delivery → qa_

QA cannot start until all required delivery dimensions pass.

- `delivery_angle_complete` [semantic_check] — _no spec_

Evidence: `team_manifest_record`, `delivery_review_record`

Evidence bindings (copy into `runtime transition`):

- `team_manifest_record`: `--evidence team_manifest_record=<reference>`
  Accepted kinds: `builder_report`, `team_manifest`
- `delivery_review_record`: `--evidence delivery_review_record=<reference>`
  Accepted kinds: `delivery_review`

If a binding is missing, retry with the command above; run `loop-harness explain PTR-VERIFY-01` to inspect current candidates.

### `PTR-VERIFY-02` {#ptr-verify-02}

_qa → e2e_browser_

Real-browser E2E follows complete QA evidence.

- `qa_angle_complete` [semantic_check] — _no spec_

Evidence: `team_manifest_record`, `qa_review_record`

Evidence bindings (copy into `runtime transition`):

- `team_manifest_record`: `--evidence team_manifest_record=<reference>`
  Accepted kinds: `builder_report`, `team_manifest`
- `qa_review_record`: `--evidence qa_review_record=<reference>`
  Accepted kinds: `qa_review`

If a binding is missing, retry with the command above; run `loop-harness explain PTR-VERIFY-02` to inspect current candidates.

### `PTR-VERIFY-03` {#ptr-verify-03}

_e2e_browser → clean_round_evaluation_

Clean-round evaluation follows complete real-browser E2E evidence.

- `e2e_angle_complete` [semantic_check] — _no spec_

Evidence: `team_manifest_record`, `e2e_review_record`

Evidence bindings (copy into `runtime transition`):

- `team_manifest_record`: `--evidence team_manifest_record=<reference>`
  Accepted kinds: `builder_report`, `team_manifest`
- `e2e_review_record`: `--evidence e2e_review_record=<reference>`
  Accepted kinds: `e2e_review`

If a binding is missing, retry with the command above; run `loop-harness explain PTR-VERIFY-03` to inspect current candidates.

### `PTR-VERIFY-04` {#ptr-verify-04}

_clean_round_evaluation → clean_round_passed_

Only one complete, current and blocker-free round passes.

- `same_review_round` [semantic_check] — verification.EvaluateCleanRound reports that every `pass` evidence item under consideration was recorded against runtime.review.round, so the round cannot close on stale evidence from a prior round.
- `all_required_dimensions_passed` [semantic_check] — verification.EvaluateCleanRound reports that every required verification dimension (delivery, qa, and any risk-triggered dimensions) has a current `pass` verdict for the active review round.
- `no_invalidated_pass_evidence` [semantic_check] — verification.EvaluateCleanRound reports no `pass` evidence in runtime.evidence[] carries an `invalidated_at` stamp for the active review round.
- `no_open_blocking_bugs` [semantic_check] — No BUG with severity=P0 in runtime.entities.bugs[] remains in an open state (`investigating`, `accepted`, `assigned`, `fixing`, `retesting`) for the current review round.

Evidence: `clean_round_record`

Evidence bindings (copy into `runtime transition`):

- `clean_round_record`: `--evidence clean_round_record=<reference>`
  Accepted kinds: `clean_round`

If a binding is missing, retry with the command above; run `loop-harness explain PTR-VERIFY-04` to inspect current candidates.

### `PTR-VERIFY-05` {#ptr-verify-05}

_clean_round_evaluation → delivery_

Mixed, incomplete or stale evidence restarts the full review.

_No guards._

Evidence: `clean_round_record`

Evidence bindings (copy into `runtime transition`):

- `clean_round_record`: `--evidence clean_round_record=<reference>`
  Accepted kinds: `clean_round`

If a binding is missing, retry with the command above; run `loop-harness explain PTR-VERIFY-05` to inspect current candidates.

## Global transitions

### `GTR-001` {#gtr-001}

_planning|document_verification|building|verification|bug_resolution|acceptance|release_audit → paused_

A user may pause from any active state.

_No guards._

Evidence: `human_decision_record`, `pause_record`

Evidence bindings (copy into `runtime transition`):

- `human_decision_record`: `--evidence human_decision_record=<reference>`
  Accepted kinds: `human_decision`
- `pause_record`: `--evidence pause_record=generated:pause_checkpoint` (generated pause checkpoint)
  Accepted kinds: `human_decision`

If a binding is missing, retry with the command above; run `loop-harness explain GTR-001` to inspect current candidates.

### `GTR-002` {#gtr-002}

_planning|document_verification|building|verification|bug_resolution|acceptance|release_audit → paused_

Any required locked REQ change pauses automation.

_No guards._

Evidence: `pause_record`

Evidence bindings (copy into `runtime transition`):

- `pause_record`: `--evidence pause_record=generated:pause_checkpoint` (generated pause checkpoint)
  Accepted kinds: `human_decision`

If a binding is missing, retry with the command above; run `loop-harness explain GTR-002` to inspect current candidates.

### `GTR-003` {#gtr-003}

_planning|document_verification|building|verification|bug_resolution|acceptance|release_audit → paused_

Production-data, security, compliance or irreversible actions require human approval.

_No guards._

Evidence: `pause_record`

Evidence bindings (copy into `runtime transition`):

- `pause_record`: `--evidence pause_record=generated:pause_checkpoint` (generated pause checkpoint)
  Accepted kinds: `human_decision`

If a binding is missing, retry with the command above; run `loop-harness explain GTR-003` to inspect current candidates.

### `GTR-004` {#gtr-004}

_verification|bug_resolution → paused_

Configured repair limits never silently close a BUG.

_No guards._

Evidence: `pause_record`, `bug_batch_record`

Evidence bindings (copy into `runtime transition`):

- `pause_record`: `--evidence pause_record=generated:pause_checkpoint` (generated pause checkpoint)
  Accepted kinds: `human_decision`
- `bug_batch_record`: `--evidence bug_batch_record=<reference>`
  Accepted kinds: `bug`

If a binding is missing, retry with the command above; run `loop-harness explain GTR-004` to inspect current candidates.

### `GTR-005` {#gtr-005}

_planning|document_verification|building|verification|bug_resolution|acceptance|release_audit → paused_

Runtime/document inconsistency fails closed.

_No guards._

Evidence: `pause_record`

Evidence bindings (copy into `runtime transition`):

- `pause_record`: `--evidence pause_record=generated:pause_checkpoint` (generated pause checkpoint)
  Accepted kinds: `human_decision`

If a binding is missing, retry with the command above; run `loop-harness explain GTR-005` to inspect current candidates.

