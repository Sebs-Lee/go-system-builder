---
name: agent-dispatch
description: Use when spawning, reassigning, or reactivating any specialized Builder or Reviewer Agent; covers the three L4 dispatch modes (plan_checkpoint continuous execution by default, plan_approval_required for high-risk work, one_shot for idempotent single actions)
category: methodology
version: 1.0.0
---
# Agent Dispatch

## Authority
Dispatch modes and the agent lifecycle are defined in `docs/loop-definition.json` (`entity_lifecycles.agent`) and `blueprint/L4-agent-dispatch-governance.md` §3.3/§7; the runtime events are executed by `loop-harness runtime agent-event` / `runtime task-complete`.

## Entry Conditions
- An Agent Definition exists under `agents/<role>.md`.
- A validated team manifest assignment exists (S7 reviewer assignments carry `claim_ids`; S6 builders carry the TASK binding).
- The document chain (TASK/BUG → contracts → REQ → design) is readable and fingerprinted.

## Required Inputs
| Input | Path / field | Why |
|:--|:--|:--|
| Assignment | manifest `assignments[]` (agent_id, dispatch_mode, scope) | dispatch coordinates and mode |
| Document chain | locked REQ/design/contracts/TASK | what the Worker must read first |
| Runtime state | `.claude/loop-state.json` | CAS revision for every event |

## Dispatch modes
| mode | flow | use when |
|:--|:--|:--|
| `plan_checkpoint` (default) | Worker reads assignment → sends ONE PLAN_REPORT (message_type `plan_report`) → **continues immediately**; Main stays silent when aligned, sends CORRECTION only on semantic drift | ordinary work — the vast majority |
| `plan_approval_required` | readback (`readback_response`) → Main `understanding_approved` → activation envelope | high-risk or irreversible work (destructive tests, production-adjacent data, spec changes) |
| `one_shot` | single idempotent action, no plan checkpoint | trivial fire-and-forget tasks |

## Procedure
1. Launch the Worker with the assignment (register-workgroup already stamped `dispatch_mode` on the agent row; default `plan_checkpoint`).
2. `plan_checkpoint`: the Worker sends one PLAN_REPORT — assignment_id/revision, objective, planned paths, steps, assertion checks, dependencies, risks — and starts working without waiting. Main reviews asynchronously; silence means aligned, CORRECTION means drift.
3. `plan_approval_required`: the Worker submits a readback; Main approves (`understanding_approved`) before the activation envelope is accepted.
4. Register the plan/readback with `loop-harness runtime agent-event --event readback_submitted --message <file>`; then `activation_sent` (the hash chain binds the envelope to the submitted plan/readback file bytes — compute with `shasum -a 256 <file>` / `sha256sum <file>` after writing it).
5. `work_started` → work → `runtime task-complete` (Builders) or `runtime review-result submit` (S7 Reviewers).

## Outputs
- Agent lifecycle events under CAS (journal-visible).
- PLAN_REPORT / readback evidence files bound by the activation hash chain.

## Exit Conditions
- The Worker reaches `reported` (result submitted) and the controller consumes it; or a BLOCKER path resolves; or shutdown is approved.

## Stop Conditions
- Activation hash mismatch (plan file drifted) — recompute and resubmit.
- Requested scope exceeds the Agent Definition — revise the assignment, never widen silently.
- `plan_approval_required` work without an approval — do not self-activate.

## Non-Goals
- Do not treat "no reply from Main" as permission for the main session to do the delegated work.
- Do not add approval rounds to ordinary `plan_checkpoint` work "to be safe" — that is the complexity this mode exists to remove.
- Do not use `one_shot` for anything with side effects.

## Inlined Methodology
Loop Engineering dispatches Agents in three modes (L4 §3.3). The default is continuous execution: one structured PLAN_REPORT is the checkpoint — it makes the plan inspectable and correctable while the work proceeds, without a synchronous approval wait. The two-round readback → approval → activation flow remains for genuinely high-risk or irreversible work, and the activation hash chain (`approved_readback_sha256`) applies in both modes: the envelope proves which plan the Worker actually saw. The first-write barrier (PreToolUse) blocks a dispatched Worker's first product mutation until its PLAN_REPORT is recorded; PostToolUse(SendMessage) observes the report automatically when the platform payload identifies the sender.
