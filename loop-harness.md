# loop-harness — Transition Checklist

> What `loop-harness` checks at every transition. Read the
> relevant section before calling `forward`; verify each bullet
> before requesting the harness to advance.

- **Path**: `.claude/bin/loop-harness.md`
- **Harness version**: dev
- **Loop definition SHA-256**: `7a8d25d4165a2f4da1a92ae88f8e673447c2f18ea636cacd0a99e7647bbb3dd9`

---

## Controller recovery protocol

The Hook is an event trigger for the Loop Controller, not only a guard. On `SessionStart`, `PreCompact`, `SubagentStart`, `SubagentStop`, and `TeammateIdle`, the Controller reads the Runtime, refreshes the resumable Milestone through CAS, and emits a positive `LOOP RECOVERY` packet.

1. Read the `Next` action and current `Stage` from the Hook packet, then follow its `Read in order` list.
2. Read the linked `docs/agent-protocol.md#sN` section before acting.
3. If blocked or the Runtime is unclear, read this Manual. Use `runtime reconcile` only when the Hook reports an integrity/CAS recovery condition; do not call `status`/`next` during normal continuation. When the live Quality Gate checklist is unclear, run `loop-harness ready` (diagnostics; never hand-push a Transition from it). `doctor` is schema/manual/policy_ref/metrics only — not stage readiness.
4. Execute the one missing deliverable/evidence named by Hook/`ready` `missing[]`; do not invent a parallel lifecycle.
5. For `SubagentStop`, complete the report, worktree review, merge-back to the current `develop` integration branch and `completion_ack` checklist before acknowledging the stop. For `TeammateIdle`, re-wake the same teammate. The identical integration chain is available explicitly via `runtime task-integrate --assignment-id <id>` when the automatic SubagentStop payload cannot identify the assignment.
6. Builder completion is registered with `runtime task-complete` — one atomic command (message validation + evidence envelope derivation + Agent/TASK advance + evidence registration); the legacy `agent-event completion_reported` + `runtime evidence add` dual write still works but produces a thinner envelope. Before the Builder writes, create its worktree (`git worktree add .worktrees/<assignment-id> -b wt/<assignment-id> develop`) and record `worktree_path`/`branch`/`target_branch` on the manifest row — SubagentStop integration requires them.
7. Stop only at a human Gateway, an external asynchronous wait, or the end of the current turn.

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
- [`TR-002`](#tr-002) planning → document_verification — Planning completes when contracts are locked and the TASK batch is fully complete on disk (planning_complete), with clause coverage, DAG acyclicity, and closing contracts verified (tasks_checked).
- [`TR-003`](#tr-003) document_verification → building — Lock only the exact contract and task versions jointly verified.
- [`TR-004`](#tr-004) document_verification → planning — Non-REQ document findings return to planning.
- [`TR-005`](#tr-005) document_verification → paused — REQ changes always return control to the human.
- [`TR-006`](#tr-006) building → verification — Start a complete review round after every TASK in the TR-003 exact execution batch has a Builder Result with passing checks, no unapproved scope deviations, and a verified integration checkpoint.
- [`TR-007`](#tr-007) building → planning — Non-REQ execution conflicts return through planning and document verification.
- [`TR-008`](#tr-008) verification → bug_resolution — A sealed ObservationBatch carries the exact immutable Finding set (with encounters and evidence refs) into S8 diagnosis; S8 never re-reproduces symptoms by default.
- [`TR-009`](#tr-009) verification → acceptance — Acceptance requires the machine CleanRound: every required Claim of the current ReviewPlan has a consumed pass Result, no current-round Finding exists, and the clean-round snapshot is recomputed by the clean_round_valid guard at promotion time.
- [`TR-010`](#tr-010) verification → paused — A ReviewResult verdict of req_change_required pauses the loop for a human REQ decision.
- [`TR-011`](#tr-011) verification → paused — A ReviewResult verdict of release_blocked pauses the loop for a human release decision.
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

- [`PTR-BUG-01`](#ptr-bug-01) investigation → bug_report_review — A sealed ObservationBatch (the exact S7 Finding set) cannot become repair work before root-cause evidence exists.
- [`PTR-BUG-02`](#ptr-bug-02) bug_report_review → repair_readback — The orchestrator approves canonical BUGs and duplicate mappings.
- [`PTR-BUG-03`](#ptr-bug-03) bug_report_review → investigation — Insufficient BUG reports return to investigation.
- [`PTR-BUG-04`](#ptr-bug-04) repair_readback → fixing — Repair execution uses the same agent dispatch gate (plan_checkpoint by default).
- [`PTR-BUG-05`](#ptr-bug-05) fixing → targeted_reverification — Every repair invalidates affected historical PASS evidence before recheck.
- [`PTR-BUG-06`](#ptr-bug-06) targeted_reverification → ready_for_full_review — A targeted pass never substitutes for the full review round; it records the S9-to-S7 handoff checkpoint.
- [`PTR-BUG-07`](#ptr-bug-07) targeted_reverification → investigation — Failed repair verification restarts root-cause investigation.

_Phase: planning_

- [`PTR-PLAN-01`](#ptr-plan-01) design → contracts — Advance formal planning from design to contracts after the design quality gate passes.
- [`PTR-PLAN-02`](#ptr-plan-02) contracts → tasks — Advance formal planning from contracts to tasks after the contract quality gate passes.

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

Planning completes when contracts are locked and the TASK batch is fully complete on disk (planning_complete), with clause coverage, DAG acyclicity, and closing contracts verified (tasks_checked). Gate readiness facts: a TASK declaring `Status: complete` on disk (or a registered complete task) plus valid planning_task evidence — the missing token `document:task:complete` means neither was found. On commit, locked contracts and complete tasks are registered into documents[].

- `planning_complete` [semantic_check] — At least one current-baseline contract document has status=locked with a matching on-disk markdown Status field (contracts are registered by PTR-PLAN-02), AND every docs/tasks/TASK-*.md declares status complete or cancelled with at least one complete (the batch is registered by TR-002's own register_planning_tasks action). Fingerprints are owned by registration and reachability, not re-checked here.
- `tasks_checked` [semantic_check] — S4's mechanical close (semantic.TasksCheck) runs at TR-002: the TASK batch is fully complete (cancelled tasks excluded), every task has an existing primary contract and a Closing Contract block, clause coverage between the CONTRACTS index universe and TASK §3 declarations closes in both directions, and the §8 dependency graph is acyclic (cycle path reported).

### `TR-003` {#tr-003}

_document_verification → building_

Lock only the exact contract and task versions jointly verified.

- `joint_document_pass` [evidence_attestation] — _no spec_
- `verified_versions_current` [evidence_attestation] — Every current-generation registered document still matches its on-disk sha, so the lock cannot advance on drifted inputs. The real check runs in GATE-DOCUMENT-PASS's registered-document drift screen (a `document_drift:<path>` conflict blocks the gate); the guard body itself only rejects an empty evidence map.

Evidence: `document_review_record`

Evidence bindings (copy into `runtime transition`):

- `document_review_record`: `--evidence document_review_record=<reference>`
  Accepted kinds: `document_review`

If a binding is missing, retry with the command above; run `loop-harness explain TR-003` to inspect current candidates.

### `TR-004` {#tr-004}

_document_verification → planning_

Non-REQ document findings return to planning.design for reworking the failing artifacts.

- `req_baseline_unchanged` [semantic_check] — The locked REQ's sha256 recorded in runtime.bound_req still matches the file at runtime.bound_req.path, so the rework loop cannot silently advance on a changed REQ.

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

Start a complete review round after every TASK in the TR-003 exact execution batch has a Builder Result with passing checks, no unapproved scope deviations, and a verified integration checkpoint. S7 verification planning starts from the real integrated diff at its own entry; the building stage no longer demands S7 team-manifest evidence it cannot legitimately produce.

_No guards._

Evidence: `builder_report_record`

Evidence bindings (copy into `runtime transition`):

- `builder_report_record`: `--evidence builder_report_record=<reference>`
  Accepted kinds: `builder_report`, `agent_completion`

If a binding is missing, retry with the command above; run `loop-harness explain TR-006` to inspect current candidates.

### `TR-007` {#tr-007}

_building → planning_

Non-REQ execution conflicts return through planning and document verification.

- `req_baseline_unchanged` [semantic_check] — The locked REQ's sha256 recorded in runtime.bound_req still matches the file at runtime.bound_req.path, so the rework loop cannot silently advance on a changed REQ.

Evidence: `change_impact_record`

Evidence bindings (copy into `runtime transition`):

- `change_impact_record`: `--evidence change_impact_record=<reference>`
  Accepted kinds: `change_impact`

If a binding is missing, retry with the command above; run `loop-harness explain TR-007` to inspect current candidates.

### `TR-008` {#tr-008}

_verification → bug_resolution_

A sealed ObservationBatch carries the exact immutable Finding set (with encounters and evidence refs) into S8 diagnosis; S8 never re-reproduces symptoms by default. One BUG draft is created per Finding, deduplicated by finding content hash.

- `observation_batch_sealed` [semantic_check] — _no spec_

Evidence: `observation_batch_record`

Evidence bindings (copy into `runtime transition`):

- `observation_batch_record`: `--evidence observation_batch_record=<reference>`
  Accepted kinds: `observation_batch`

If a binding is missing, retry with the command above; run `loop-harness explain TR-008` to inspect current candidates.

### `TR-009` {#tr-009}

_verification → acceptance_

Acceptance requires the machine CleanRound: every required Claim of the current ReviewPlan has a consumed pass Result, no current-round Finding exists, and the clean-round snapshot is recomputed by the clean_round_valid guard at promotion time.

- `clean_round_valid` [semantic_check] — _no spec_

Evidence: `clean_round_record`

Evidence bindings (copy into `runtime transition`):

- `clean_round_record`: `--evidence clean_round_record=<reference>`
  Accepted kinds: `clean_round`

If a binding is missing, retry with the command above; run `loop-harness explain TR-009` to inspect current candidates.

### `TR-010` {#tr-010}

_verification → paused_

A ReviewResult verdict of req_change_required pauses the loop for a human REQ decision. The verdict transaction (runtime review-result submit) already created the single authoritative pause checkpoint; this transition only moves the cursor.

- `pause_checkpoint_recorded` [semantic_check] — _no spec_

Evidence: `review_result_record`

Evidence bindings (copy into `runtime transition`):

- `review_result_record`: `--evidence review_result_record=<reference>`
  Accepted kinds: `review_result`, `delivery_review`, `qa_review`, `e2e_review`

If a binding is missing, retry with the command above; run `loop-harness explain TR-010` to inspect current candidates.

### `TR-011` {#tr-011}

_verification → paused_

A ReviewResult verdict of release_blocked pauses the loop for a human release decision. The verdict transaction (runtime review-result submit) already created the single authoritative pause checkpoint; this transition only moves the cursor.

- `pause_checkpoint_recorded` [semantic_check] — _no spec_

Evidence: `review_result_record`

Evidence bindings (copy into `runtime transition`):

- `review_result_record`: `--evidence review_result_record=<reference>`
  Accepted kinds: `review_result`, `delivery_review`, `qa_review`, `e2e_review`

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

- `req_baseline_unchanged` [semantic_check] — The locked REQ's sha256 recorded in runtime.bound_req still matches the file at runtime.bound_req.path, so the rework loop cannot silently advance on a changed REQ.

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
- `baselines_unchanged` [evidence_attestation] — Every document fingerprint captured at pause time matches the on-disk file, so the resume cannot quietly advance on drifted inputs. The re-hash runs in TR-019's restore_from_pause action (fail-closed, sentinel ErrBaselineDrift routes the CLI to amendment); the guard body itself only rejects an empty evidence map.

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

- `human_abort_approved` [evidence_attestation] — The transition's evidence validation enforces that the cited human_decision evidence is current and scoped to `runtime_abort:<runtime_id>@<revision>` (human_decision_scope on TR-021/TR-030) — one approval authorizes exactly one abort at one revision; the guard body itself only rejects an empty evidence map.

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

- `req_baseline_unchanged` [semantic_check] — The locked REQ's sha256 recorded in runtime.bound_req still matches the file at runtime.bound_req.path, so the rework loop cannot silently advance on a changed REQ.

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
  Accepted kinds: `finding`, `bug`

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

- `human_abort_approved` [evidence_attestation] — The transition's evidence validation enforces that the cited human_decision evidence is current and scoped to `runtime_abort:<runtime_id>@<revision>` (human_decision_scope on TR-021/TR-030) — one approval authorizes exactly one abort at one revision; the guard body itself only rejects an empty evidence map.

Evidence: `human_decision_record`

Evidence bindings (copy into `runtime transition`):

- `human_decision_record`: `--evidence human_decision_record=<reference>`
  Accepted kinds: `human_decision`

If a binding is missing, retry with the command above; run `loop-harness explain TR-030` to inspect current candidates.

## Phase transitions: bug_resolution

### `PTR-BUG-01` {#ptr-bug-01}

_investigation → bug_report_review_

A sealed ObservationBatch (the exact S7 Finding set) cannot become repair work before root-cause evidence exists.

- `root_cause_evidence_complete` [evidence_attestation] — A root-cause evidence item (failure mode, triggering input, and minimal-repro path) is referenced from runtime.evidence[] for the BUG promoted out of `investigating`.

Evidence: `observation_batch_record`, `root_cause_record`

Evidence bindings (copy into `runtime transition`):

- `observation_batch_record`: `--evidence observation_batch_record=<reference>`
  Accepted kinds: `observation_batch`
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

Repair execution uses the same agent dispatch gate (plan_checkpoint by default).

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

Advance formal planning from design to contracts after the design quality gate passes. Gate readiness facts: an ARCHITECTURE-*.md whose top status line (状态/Status) declares `locked` plus a valid planning_design JSON envelope (responsibility Architect, conclusion pass — a markdown file as --path yields `evidence:<id>:schema`). Missing tokens `document:design:locked` / `evidence:planning_design_record` mean one of these is absent.

- `ui_impact_resolved` [semantic_check] — runtime.bound_req.metadata.ui_impact is not `unknown`, so the planning phase is not paused on the SM-003 gate waiting for §11 of the REQ to clarify UI impact.

### `PTR-PLAN-02` {#ptr-plan-02}

_contracts → tasks_

Advance formal planning from contracts to tasks after the contract quality gate passes. Gate readiness facts: a contract declaring `Status: locked` on disk (or an already-registered locked contract in runtime documents[]) plus valid planning_contract evidence — the missing token `document:contract:locked` means neither was found; flip the contract's top status line（状态/Status）to locked.

- `contracts_checked` [semantic_check] — S3's mechanical close (semantic.ContractsCheck) runs at PTR-PLAN-02: contract token references resolve against REQ FR tables and module packages, clause cells point at known contracts, and fingerprint columns match disk.
- `scenario_bridge_checked` [semantic_check] — S2's AC↔CASE bridge (scenario.GuardBridgeChecked) runs at PTR-PLAN-02: every AC of the bound REQ reaches a rule via FR source_refs (with branches), or carries an endorsed N/A (NFR id / §A4). With no module packages at all, only fully N/A-endorsed REQs pass — an AC pointing at FR- with nothing citing it is a broken denominator.

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

---

## `capture` — automatic encounter timeline collection

The `capture` verb is the harness-side half of the §3.6 / §6.3 / §8
capture-buffer 必经路径: any wrapper that the E2E Reviewer (or a test
runner, browser harness, CLI helper) calls writes sanitized timeline
steps into the Assignment's capture buffer, and `verification result
submit --captures` merges the buffer into Findings whose encounter
timeline is empty. There are two forms:

- `loop-harness capture step` — manual single-step append for the
  Reviewer. Use it when the agent already ran the action by hand and
  needs to record the sanitized action / observed / evidence refs.
- `loop-harness capture exec` — automatic collection around a wrapped
  command. Use it for the E2E command_flow observation mode (§3.6):
  the wrapper runs the command once, collects every timeline field
  the spec requires, and freezes the evidence window on a non-zero
  exit so the Reviewer can label the failure boundary in the Finding.

The two forms share the same buffer file (one JSONL line per step) and
the same redaction gate; the manual form is for human-written steps,
the exec form is for wrapper-driven steps.

### `loop-harness capture exec` — auto-collect a command_flow step

```text
loop-harness capture exec --assignment <id> [flags] -- <command...>
```

Runs the wrapped command once and appends one command_flow step to the
Assignment's capture buffer. The wrapper never swallows the child's
exit code: a non-zero exit freezes the evidence window and is passed
straight back to the caller (the harness fails the same way the
command did).

Flags:

| Flag | Default | Meaning |
|:--|:--|:--|
| `--assignment` | (required) | Plan Assignment the step belongs to. The capture buffer is addressed by `(runtime_id, baseline.generation, assignment_id)`. |
| `--cwd` | `$PWD` | Working directory for the wrapped command. The wrapper records the resolved absolute path on the step. |
| `--summary-bytes` | 4096 | Per-stream bytes kept inline in the buffered step (head summary). The full stream is persisted as a typed evidence file. |
| `--max-evidence-bytes` | 1 MiB | Hard cap on each persisted stream evidence file. Overflow is truncated and the truncation is recorded honestly in the ref (`truncated=true, bytes=N`). |
| `--max-artifacts` | 20 | Maximum produced / modified / deleted artifact refs recorded per step. Overflow is reported as `+K more artifact(s) omitted by --max-artifacts` on the step. |
| `--artifact-depth` | 3 | Directory depth scanned under `cwd` for the before/after artifact digest diff. `.git`, `.claude`, `node_modules` are always skipped. |
| `--root` | `.` | Repository root (used to address the capture buffer). |

The wrapper always records the following fields on the step:

- `sequence` — auto-numbered starting at 1, derived from the current
  buffer length so two wrappers do not collide.
- `captured_at` — UTC RFC3339Nano, taken at command start.
- `action` — `exec: <rendered command> (cwd: <abs>) (tool: <name> — <version>)`.
  The command is rendered with `strconv.Quote` for any arg that
  contains whitespace or quotes; the rendered string is what passes
  the redaction gate, never raw environment.
- `observed` — `exit=<code> duration=<d>; stdout: <head>; stderr: <head>`
  for a clean run, or `FAILED (wall-action candidate; evidence window
  frozen — Reviewer: annotate last_good_checkpoint / wall_action /
  first_bad_checkpoint) exit=<code> duration=<d>; ...` for a non-zero
  exit.
- `evidence_refs[]` — one ref per recorded artifact:
  - `command_output:<rel>#sha256=<hex> (bytes=<n>, truncated=<t>)`
    for the full stdout stream, persisted at
    `.claude/evidence/<runtime>/g<n>/captures/<assignment>/exec/step-NNN-stdout.log`.
  - the same shape for the stderr stream.
  - `artifact:<rel>#sha256=<hex> (bytes=<n>, new|modified)` for each
    file the wrapped command produced or modified under `cwd`.
  - `artifact:<rel> (deleted)` for each file the command removed.
  - `env:<NAME> (present; value never captured)` for every environment
    variable whose *name* matches the sensitive-name pattern
    (`token|password|passwd|secret|api[_-]?key|private[_-]?key|access[_-]?key|credential`).
    The presence is recorded; the value never is.

Redaction (§6.3宁拒勿放): two gates run on the wrapper's input/output.
The pre-exec gate refuses to spawn the command if the rendered
command line matches a secret pattern
(`password=`, `secret=`, `api_key=`, `bearer <token>`, `token=<value>`,
etc.); the command is not executed, no buffer file is written, and
the wrapper exits with code 2. The post-exec gate runs over the
persisted stream: if the stream's bytes match a secret pattern, the
evidence file is replaced by a placeholder
(`[capture withheld: stream matched a secret redaction pattern; full
output not persisted (L3-S7 §6.3)]`) and the ref records a capture
gap. The buffer step's `observed` field never includes the redacted
value in either case.

Evidence window freeze: a non-zero exit writes
`failure.json` next to the buffer with the assignment, sequence,
captured_at, command, exit code, `wall_action_candidate=true`, and
`evidence_window_frozen=true`. The buffered step is also annotated
with the `FAILED` prefix. The Reviewer's only remaining work is to
open the Finding, label `last_good_checkpoint` (the last clean step
before this one), `wall_action` (this step), and
`first_bad_checkpoint` (the next clean step after, or "no recovery
yet" if the run ended here). The wrapper does not attempt to
recover, retry, or fix — that is the S8 Investigator's job.

Memory and size invariants: the wrapper does not hold the full
stream in memory. The stream recorder caps the persisted evidence
file at `--max-evidence-bytes`, hashes only the persisted bytes,
and remembers the total stream length so the truncation is recorded
honestly. The artifact walk caps the number of files digested
(`execWalkEntriesMax`), the directory depth (`--artifact-depth`),
and the number of refs persisted (`--max-artifacts`), so a runaway
tree diff cannot flood the buffer.

### Product-side wrappers (browser, Playwright, test runner, trace collector)

The harness provides the buffer file format, the redaction gate, and
the `verification result submit --captures` merge binding. The
injection half — the wrapper that observes a Playwright action, a
test-runner invocation, or a network/console event and decides when
to call `loop-harness capture` — belongs to the product side. The
boundary is intentional: the harness stays runtime-agnostic and the
product keeps ownership of its own observation surfaces.

The contract a product-side wrapper must satisfy to be drop-in
compatible with the harness:

1. The wrapper is invoked once per material action (browser step,
   test-runner call, network/console event). It assembles one
   capture step and calls the appropriate `loop-harness capture`
   form. For CLI-like actions, `loop-harness capture exec -- <cmd>`
   is the reference: sequence/time, cwd, tool version, sanitized
   command, exit code, bounded stdout/stderr summaries bound to
   typed evidence files, and the produced/modified artifact digest
   diff. For browser actions, the equivalent step is
   `loop-harness capture step --assignment <id> --action
   "navigate: <route>" --observed "<url, title, status>" --evidence
   shot.png,trace.json,net-3.har`. For test runner actions, the
   equivalent step is `loop-harness capture step --action "test:
   <name>" --observed "<pass|fail, duration>" --evidence
   junit.xml,coverage.json`. The shape — sanitized action, observed
   checkpoint, evidence refs — is identical across all forms.
2. The wrapper runs the redaction gate *before* persisting: any
   field matching a secret pattern
   (`password=`, `secret=`, `api_key=`, `bearer <token>`,
   `token=<value>`) is withheld, replaced by a redacted ref
   (`env:<NAME> (present; value never captured)`,
   `command_output:<rel> (withheld: secret pattern matched; full
   output not persisted)`). When in doubt, refuse to record the
   value (宁拒勿放).
3. Sensitive env var *names* are recorded as presence-only refs;
   values are never captured. The naming list is the same regex
   the harness uses:
   `(?i)(token|password|passwd|secret|api[_-]?key|private[_-]?key|access[_-]?key|credential)`.
4. On a non-zero outcome (test failure, network 5xx, console error
   block, exception), the wrapper writes the equivalent of
   `failure.json` next to the buffer with the step's sequence and a
   `wall_action_candidate=true` flag, and annotates the buffered
   step with the `FAILED` prefix. The wrapper does not retry, fix,
   or skip — those are S8/S9 work.
5. The wrapper does not need to read the runtime; the harness
   resolves the buffer path from `(runtime_id, baseline.generation,
   assignment_id)`. The product side just passes `--assignment` and
   lets the harness do the addressing.
6. Long outputs are bounded: the wrapper caps any inline summary
   (~4 KiB) and persists the full stream as a typed evidence file
   with its own size cap. The reference implementation for a CLI
   command is `loop-harness capture exec` itself; a Playwright
   wrapper should follow the same pattern with `--summary-bytes`
   mapped to the JSONL step field and the full network/console
   trace persisted as `<assignment>/exec/step-NNN-<kind>.json`.

The harness provides the buffer file format, the redaction gate, and
the merge into the Finding's encounter timeline. The product
provides the observation trigger and the typed evidence files
behind each ref. The two halves meet at the buffer file; no other
integration is required.

---

## Sandbox: entering S7 directly (test/demo only)

S7 is normally entered automatically: `building → verification` is
TR-006 (`runtime transition --id TR-006`), triggered by the Hook
when every TASK in the TR-003 exact execution batch has a Builder
Result with passing checks. `bug_resolution → verification` is
TR-012, and `acceptance → verification` is TR-016. The state
machine is the only authority for these transitions; the harness
deliberately has no seed/back-door that flips the lifecycle to
`verification` outside a registered transition.

But sandboxed test/demo work (an E2E tester, a recovery rehearsal,
a doc walk-through) sometimes needs the S7 ReviewPlan scaffold
without driving a real builder batch. The recommended path is:

1. Use a Go test helper that builds a real runtime via the
   published verbs. The fixture helpers in
   `internal/review/review_test.go` are the canonical reference:
   `baseVerificationState()` returns a complete, schema-valid
   `verification` runtime with `round=1`, and
   `registerFixturePlan(t, root, statePath, journalPath)` calls
   `RegisterPlan` exactly the way production does, so the harness
   emits the right CAS/transition events and the round starts in
   the same shape an E2E user would see. **Mirror that pattern;
   don't handcraft `loop-state.json` if you can avoid it.**
2. When a real harness verb is not appropriate (a one-off screencast
   on a clone that never runs `go test`), the minimum viable
   `loop-state.json` shape that lets `loop-harness s7 draft` and
   `runtime review-plan --file` proceed is documented below. Every
   required field is enforced by `internal/schema/assets/loop-state.schema.json`;
   the runtime's semantic validator
   (`internal/runtime/store.go:validateLifecycle`) rejects unknown
   state/phase combinations.

Minimum viable `loop-state.json` for a sandbox S7 run (test/demo
ONLY; never use in a real product release path):

| Top-level path | Required shape | Why it matters |
|:--|:--|:--|
| `lifecycle.state` | `"verification"` | `internal/review/register.go` rejects any plan registration whose lifecycle state is not `verification`; `s7 draft` itself now exits with a stage-specific hint when this is wrong (see the `runS7Draft` disclosure). |
| `lifecycle.phase` | `"planned"` or `"running"` | Phase must be one of the names registered in `docs/loop-definition.json` under the `verification` owner_state. |
| `lifecycle.phase_revision` | integer `>= 0` | Required by the loop-state schema; bumped under CAS by every transition. |
| `review.round` | integer `>= 1` | A round value of 0 is the "no S7 round yet" sentinel; `s7 draft` refuses to scaffold below round 1. |
| `review.clean_round` | `null` | Must be null while the round is open; `clean_round_valid` recomputes it at TR-009 promotion time. |
| `bound_req.id` / `path` / `sha256` / `version` | non-empty | The bound REQ is the round's identity anchor: every TASK claim's `source_refs`, the manifest's `documents[]`, and the `verification_artifact_digest` all derive from it. |
| `documents[]` (TASK rows) | one entry per current-generation TASK, each with `kind=task`, `id`, `path`, `sha256`, `generation == baseline.generation`, `version` | `ValidatePlanTaskCoverage` (L3-S7 §4.4) refuses to register a plan whose `Claims[].source_refs` doesn't cover every current-generation TASK; without these rows the draft has no DV claim to begin with. |
| `entities.tasks[]` | a row per TASK referenced by `documents[]` | The Hook, the runtime transition guard, and the manifest validator all read this projection. |
| `baseline.generation` | integer `> 0` | `register.go` rejects a plan whose `baseline_generation` does not match the runtime's; the draft reads it for the DV/QA pivot. |
| `baseline.{unit_test_status, integration_test_status, build_status, integration_checkpoint_verified}` | the exact keys the S6 hook writes | These are the fields TR-006 consults; if you skip S6 to enter S7 directly, populate the keys whose values are `passed` so the on-disk evidence envelope is honest about being faked. |

Two warnings before you ship this anywhere besides a sandboxed
clone:

- CleanRound / acceptance consume the **full** evidence chain. A
  handcrafted `verification` runtime will pass plan registration,
  ReviewResult submission, and ObservationBatch sealing, but
  `clean_round_valid` and the release-audit guard read the
  builder-result history directly — they will reject any release
  whose history was manufactured rather than produced. Treat the
  table above as a way to exercise the S7 surface, not to bypass
  S6.
- The five E2E tester harnesses (`tests/system/req039/...`) and
  the recovery regression in `internal/cli/controller_s7_recovery_test.go`
  all drive S7 through `registerFixturePlan` rather than handcrafted
  JSON. New test files should do the same: the helpers
  (`baseVerificationState` + `registerFixturePlan` in
  `internal/review/review_test.go`, `writeManifestDraftFixture` in
  `internal/cli/s7_manifest_draft_test.go`) are the supported way
  to land in S7 deterministically.

