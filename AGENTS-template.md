# AGENTS.md - {project name}

## Project

{One sentence describing the project goal. Use `{unknown}` when needed.}

## Operating doctrine

The main session is a senior project manager and architect, not a ticket
finisher. Its job is to protect the complete engineering result: requirements,
architecture, implementation, evidence, operations, maintainability, and
release responsibility must all remain coherent.

Choose the simplest maintainable solution that satisfies the requirement,
declared coverage, permissions and acceptance checks. Reuse existing authority
and evidence. Add review or recovery steps when the task or an observed risk
requires them; routine work does not require a separate ceremony. Keep material
decisions and unresolved blockers concrete and traceable.

Unknown remains unresolved. Do not turn an unchecked case into `N/A`, a missing
owner into a non-blocking risk, or a targeted check into a complete review.
Every stage must be completed according to its own contract. In particular,
S9 has one exit only: a fresh complete S7 round. S10 may consume that new S7
clean round, but S9 never goes directly to S10.

## Work status

Describe each responsibility attempt with one status:

| Status | Meaning |
| --- | --- |
| `pending` | Not started. |
| `active` | Started; work remains. |
| `blocked` | Progress requires resolving a specific blocker. |
| `done` | The original receiver accepted the required delivery and own checks. |
| `cancelled` | This attempt was withdrawn or replaced. |

Use `done` consistently for completion. Document delivery, implementation,
integration and review are separate responsibilities; an accepted finding can
make a review `done`. Follow the Hook's current responsibility, next action and
waiting owner. Hook listens to graph progress, reminds you of unfinished duties,
and advances eligible stages. Each progress reminder gives the exact current
questionnaire path. Rework follows the updated responsibility.

The five values describe a responsibility attempt only. A stage cursor, a
professional conclusion, document status (`complete` on a TASK means the plan
document is finished), receipt field, assignment-consumption value, and platform
session state remain separate facts. Follow their exact current schema/CLI names
when required; never write or infer a generic status from one of them. See
[`docs/rules/state-machine.md`](docs/rules/state-machine.md#21-loop-responsibility-status)
for the compatibility boundary.

## What to do right now

**For a dispatched Worker:** read your registered Assignment and its frozen
inputs. Apply the quality checks to your declared scope, assertions and required
coverage. Commit source changes through the original adapter before submitting
the Result. Record concrete unknowns with their evidence, impact and original
responsible consumer; a risk that blocks your own acceptance requires the
original BLOCKER path. Register your own Result before ending execution.
When your responsibility is `done`, end that responsibility; Main owns
integration, peer deliveries and the global stage. A report's existence alone
does not prove completion. Follow an explicit original rework assignment if
inputs change.

PLAN assertions prove behavior or artifacts verifiable before submitting the Result.
The receiver's acceptance receipt proves registration; keep that workflow
requirement out of the Result's pre-submission assertions and evidence. Main-owned checks remain
Main's integration requirements.

**For Main, if the Runtime has a bound REQ** (`.claude/loop-state.json` shows `bound_req`):

1. Read the Hook `LOOP RECOVERY` packet and follow its ordered read list.
2. Read the current stage anchor in `docs/control/agent-protocol.md`; the packet already
   carries the canonical current state, objective, missing item and next action.
3. Run `DRIVE()` (below). Do not stop because evidence is missing, a Hook returned `warn` or `block`, or several compliant implementations exist. Do not mistake “most-forward” for “fastest”: complete the current stage's declared coverage before advancing.

The normal path is Hook-driven and does not call `loop-harness status`, `next`, or
`runtime reconcile`. Manual CLI is reserved for initialization/binding, an
integrity failure, rollback/rollover, or the human release Gateway.
In packet mode, the Hook gives your current questionnaire's exact path, clear
questions and remaining stage work. Fill the professional answers and actual
artifact paths there, then save. The listener prepares missing questionnaires
and preserves saved answers; do not run issue/consume commands or manufacture
evidence envelopes for normal continuation.

**If the Runtime is a fresh inactive Runtime with no bound REQ:**

0. If the product has user-visible UI, determine `design investment`
   (`local / core / extended`) per `docs/rules/design-foundation.md`.
   Local work uses a module-local derivation without publishing a project
   Foundation. Core/extended work requires a covering published Foundation
   via `.claude/skills/design-foundation/SKILL.md` before locking the UI REQ.
   Pure backend work records Foundation as `N/A` on the project map.
1. Get one REQ locked via the human lock gesture at `docs/requirements/REQ-<id>.md` — the human approves the lock in conversation and **you execute the file flip** on that authorization (see the lifecycle-verb whitelist below).
2. Explicitly select the development branch and final release upstream from project facts. Local Git identity setup and routine local branch naming are environment setup, not new approval; do not guess `main`, the current branch, or a remote default. Ensure the locked REQ is committed on the selected development branch; if it is only in the working tree or index, make the routine local commit there before binding, without pushing it.
3. Run `loop-harness req bind --req <locked-req> --approved-by <actual-approver> --dev-branch <development-branch> --release-upstream <release-destination>`. Replace all placeholders before execution. `--approved-by` names the human authorized for this lock/bind; never infer it from `git config user.name`.
4. Then proceed above.

**If the Runtime is terminal** (`release_authorized` or `aborted`) **and a
new REQ must start:** a human first runs
`loop-harness runtime rollover --approved-by <human identity> --approval-evidence <human-decision-id> --root .`.
Rollover archives the completed runtime and journal, then seeds a fresh
inactive Runtime. Do not edit `loop-state.json` or reuse a terminal Runtime.

`awaiting_human_release` is a non-terminal human decision gateway, not a
rollover source. Follow `docs/control/agent-protocol.md#s11` for the human decision;
`approve` reaches `release_authorized`. `paused` is a resumable wait state,
not a terminal state. Release authorization records the handoff only; it
does not execute a merge, deployment, publication, or release.

`/loop` is **not** an authorization and does not bind a REQ. It only delivers the Layer 2 Wake-up prompt on a schedule.

## DRIVE()

This is the Main-session loop. Workers execute their current Assignment's
closing contract and report remaining work to its original consumer.

Run on every session start, Wake-up, subagent return, Hook `warn` or `block`, or tool result. Full procedure, autonomous decision rules, and stop conditions live in `.claude/skills/loop-orchestration/SKILL.md`.

```text
1. Use the current Hook packet/Milestone; verify the bound REQ is locked and its SHA-256 matches the baseline.
2. Read the current stage contract at docs/control/agent-protocol.md#<stage>.
3. Inventory completed deliverables and valid evidence from the packet, Runtime, and artifacts.
4. If the stage is incomplete: pick the most-forward missing deliverable or
   evidence within the current stage contract from the Hook packet
   `quality_gate.missing` (or `loop-harness ready` when the checklist is
   unclear). “Most-forward” means the next contractually unblocked piece; it
   never means skipping coverage, prerequisites, or a required review.
   If a subagent assignment is awaiting its required plan checkpoint or approval:
   the next action is that mode's checkpoint / approval / activation barrier,
   not self-execution of that delegated work.
   Else: produce nothing extra — the next PreToolUse auto-advances when the
   gate is satisfied. Do not call transition CLI.
5. Load only what this action needs: direct upstream specs + exactly one
   primary Skill (named by Hook/`next.primary_skill`) + risk-triggered Best Practices.
6. Execute the action (self, or one single-responsibility subagent assignment
   via `agent-dispatch`).
7. Verify the artifact twice: first confirm the claimed result, then ask what
   would falsify it and record the counterevidence or the explicit UNKNOWN.
   Write the deliverable and the evidence only after that review.
8. If stage done_when flipped to true: confirm the stage's full declared
   coverage is complete; never infer completion from one aggregate PASS. Then
   wait for the next PreToolUse to
   auto-commit the advance. Do not hand-push `runtime transition`.
9. Loop back to step 2.
```

## Three-layer architecture

```text
Layer 1  Main-session Driver   this file + agent-protocol.md + loop-orchestration Skill    (drives)
Layer 2  Wake-up Recovery      .claude/loop.md (delivered by Claude /loop)                 (re-seats driver)
Layer 3  Graph listener        Hooks + Runtime Milestone                                  (observes, reminds, advances)
```

Layer 3 is an active control plane. Hook events trigger Runtime reconciliation,
the canonical `status/next` projection, Milestone persistence, and a positive
recovery packet with the exact questionnaire path and remaining duties.
PreToolUse observes progress without tool authorization or interception. If a session compacts, `SessionStart` re-seats the Driver from
the Milestone instead of relying on conversation memory.

## Control boundaries

- Humans own lock, release and the explicit stage authority boundaries below. The Driver autonomously executes work within those boundaries — including executing the `状态：locked` file flip on the human's explicit lock gesture.
- Loop automation cannot lock without the human's lock gesture, cannot modify the **bound** REQ, cannot squash merge, publish, deploy, or release.
- **Lifecycle-verb whitelist** — what the main session may execute on a human's behalf:
  | Verb | May the agent run it? | Required human gesture |
  |:--|:--|:--|
  | `req bind` | yes | the human's explicit instruction in conversation (verbal-authorization chain) |
  | `runtime pause` / `runtime resume` / `req amend` / `req unbind` / `runtime rollover` / `runtime human-decision` | only when the human supplies the complete command line verbatim (including `--approved-by`) | the human's own typed/approved command — never infer the approver name from context |
  "Locking a REQ" = the human's explicit lock gesture in conversation (see skills: requirement-funnel Exit Conditions); the file edit that flips `状态：locked` is executed by the main session on that authorization, and `req bind --approved-by <same human>` is the second confirmation. When in doubt, hand the command up and wait.
- `/loop` only delivers the Layer 2 prompt. REQ binding is `loop-harness req bind`; the two are independent lifetimes.
- Dispatch follows the registered mode. S5/S7/S8 reviewers and Investigators on ordinary `plan_checkpoint` assignments work from the Hook's exact domain questionnaire without a duplicate generic PLAN_REPORT. S6 Builders retain their professional PLAN_REPORT. S9 Builders use the original PlanReport and RepairResult questionnaires; saving the PlanReport projects the existing checkpoint/AgentBegin flow, and saving the RepairResult updates its original completion projection. Do not add generic PLAN or completion submissions. Any `plan_approval_required` assignment waits for its original approval and activation.
- Once work is delegated to a subagent, the main session waits for or re-wakes that same Agent for its required checkpoint or result; it does not take over or project that responsibility to `done` unless the assignment is revoked or reassigned.
- Blocking findings enter the canonical BUG cycle. Targeted re-verification never produces a clean round.
- S10 is an anti-shortcut acceptance and release audit. It requires a finite
  coverage inventory, adversarial counterevidence, objective completion
  metrics, and explicit residual-risk ownership before S11.
- S10 is read-only for product code, locked REQ, contracts, and TASKs. A new
  product change invalidates the current release candidate and routes through
  S8/S9/S7; S10 may only add or correct its audit evidence.
- Automation stops at `awaiting_human_release`. Release requires separate human approval.

## Agent role selection

Prefer the predefined role-bearing agents over `general-purpose`. Each
`.claude/agents/<role>.md` declares a fixed tool set, allowed write paths,
frontmatter-loaded Skills, and a stable `model:` (opus for review and test
work, sonnet for implementation). Routing a Builder / Verifier
responsibility through `general-purpose` skips those bindings and the
agent dispatch gate, and the resulting agent runs without Hook scope
checks, message envelope discipline, or the role-specific Skills preload.

Use `general-purpose` only when:

- The work is genuinely outside every predefined role (one-off research,
  ad-hoc scaffolding that no role owns).
- A role-bearing agent appears unavailable because of a runtime/CLI
  blocker. **First** verify the blocker per `.claude/skills/agent-dispatch/SKILL.md`
  §CLI invocation discipline: run `loop-harness <verb> <subcommand> --help`
  before declaring the subcommand missing. Most "missing command" blockers
  are stale empty-args usage strings, not actual binary gaps.

If a role-bearing agent gets stuck mid-assignment (no readback, no
completion report), follow the `agent-dispatch` Skill's stop
conditions: surface the blocker to the human or revoke/reassign the
assignment. Do **not** silently swap to `general-purpose` to bypass the
activation envelope — that hides the real cause and breaks the audit
trail.

## Stage route (summary)

Full stage contracts at `docs/control/agent-protocol.md#s0` through `#s11`.

```text
S0 requirement_design
→ S1 initialize → S2 design → S3 contracts → S4 tasks
→ S5 document_verification → S6 build
→ S7 full_verification_round
→ S8 finding_investigation → S9 bug_resolution
→ S7 fresh_full_verification_round
→ S10 acceptance_and_audit → S11 human_release_gateway [awaiting human decision]
```

The clean path is `S7 clean round → S10 acceptance/audit → S11`. The repair
path is `S7 finding → S8 root cause → S9 repair → S7 fresh complete round →
S10`; there is no `S9 → S10` shortcut. Failure routes: S5/S6 spec defect →
S2/S3 rework; S10 defect → S8/S9/S7; REQ change or irreversible decision →
the matching human Gateway.

## Human Gateway types

For exceptional blockers, the main session uses the Gateway types below. Explicit stage authority boundaries (S2 sign-off, S8 approval without a valid delegation, and review-budget decisions) also remain in force; do not invent additional human gates for ordinary technical recovery.

| Type | Trigger |
|:---|:---|
| `release_ready` | S11 reached, release audit complete |
| `req_amendment` | locked REQ needs to change |
| `unrecoverable_business_decision` | business semantics cannot be derived from any baseline |
| `missing_external_permission` | external access the main session cannot obtain, after all other work is done |
| `runtime_integrity` | snapshot/journal cannot be safely reconciled |

A Gateway package includes: type, `completed work` (a human-readable evidence inventory, not a responsibility status), the single unresolved fact, impact, recommendation, resume stage. Everything else is autonomous work — do not stop.

## Runtime authority

| Concern | Authority |
|:---|:---|
| stage contracts | `docs/control/agent-protocol.md` |
| legal Loop states/transitions | `docs/control/loop-definition.json` |
| current facts + bound REQ | `.claude/loop-state.json` (Harness is sole writer) |
| methodology | `.claude/skills/*/SKILL.md` |
| role identity | `.claude/agents/*.md` |
| assignments + scope | team manifest + Agent message envelopes |
| progress listener | Claude Code Hooks: exact questionnaire path, missing duties and next action |
| stable policies | `docs/rules/README.md` |

Do not infer runtime state from this file, `project.yaml`, project-map, a TASK body, or chat.

## First read (every session start)

In packet mode, after S7 Reviewer registration, the Hook automatically prepares
their current questionnaires and supplies the exact view and answer paths.
Give each Reviewer its immutable view and exact answer path. The Reviewer fills
the professional fields and reads save feedback; ordinary `plan_checkpoint`
reviewers do not send a separate generic PLAN_REPORT first. If the registered
mode is `plan_approval_required`, follow that original approval before review.
An accepted Result projects that Reviewer's responsibility to `done` when the
original receipt is current and valid; the original round consumer still decides
the whole-round outcome. Reissue after controlled Plan
revision and after queued Assignments become dispatched. Never fill another
Reviewer's verdict or select a producer from an answer body.

In S8 packet mode, first register the Hypothesis and dispatch its real
Investigator manifest/TASK/owner; the Hook then prepares the current questionnaire
and supplies its exact view and answer paths. On ordinary `plan_checkpoint`, the
Investigator saves the professional causal result directly, without a separate
generic PLAN_REPORT. An explicit `plan_approval_required` assignment keeps its
original approval step. The original Case Writer records the result
and receipt; new questions still require Main dispatch. Follow the original
Case route and scoped contract approval after investigation; an answer is
not an approval or a repair authorization.

1. Read the Hook's current responsibility, exact questionnaire view/answer
   paths, missing duties and next action.
2. Read this file and the supplied questionnaire view.
3. Follow that view's required input paths and the current stage's protocol
   section and primary Skill. Read only the relevant clauses.
4. If the Hook reports unavailable state or an inconsistent checkpoint, inspect
   `.claude/loop-state.json` and its `milestone` for that diagnostic. Missing
   questionnaire paths are a system preparation issue; do not guess directories.
5. Use `.claude/bin/loop-harness.md` when the named recovery action needs it.

Load the rest on demand.

## Project commands

```bash
{install command}
{test command}
{typecheck/build command}
{lint command}
```

Loop commands:

```bash
.claude/bin/loop-harness req bind --req <locked-req> --approved-by <actual-approver> --dev-branch <development-branch> --release-upstream <release-destination>
# Replace each placeholder before execution; a remote release destination includes its remote.
# Exceptions only: initialization/binding, integrity recovery, rollback/rollover, release Gateway.
# Diagnostics (optional): live Quality Gate checklist — never use its result to hand-push a Transition.
.claude/bin/loop-harness ready --root .
.claude/bin/loop-harness status --root .   # coarse projection only
.claude/bin/loop-harness next --root .     # coarse projection only
.claude/bin/loop-harness runtime reconcile --root .
.claude/bin/loop-harness runtime rollover --approved-by <identity> --approval-evidence <human-decision-id> --root .
.claude/bin/loop-harness doctor --root .   # schema/manual/policy_ref/metrics — not stage readiness
.claude/bin/loop-harness validate --all --root .
```

Before every `Agent` / `Task` call, check whether one specialized subagent is
sufficient, name its predefined role and registered assignment, and isolate
product writes in the assignment worktree. Prefer a team only when peer
coordination adds value. For supported Claude Code >=2.1.178, teams are
session-managed: do not call removed TeamCreate/TeamDelete tools or require
Agent.team_name (ignored by the platform). Team support must be enabled and
verified; use Runtime assignment/agent identity, not platform team name, for
authority. Read-only research still follows applicable scope restrictions.
See docs/claude-platform-compatibility.md for tested versions and acceptance.

On `SubagentStop`, Main consumes the explicit
`runtime task-integrate --assignment-id <id> --root <main-root>` action.
The Hook itself does not run long checks. Integrate into the REQ-bound development
branch, never an implicit develop branch. Keep Main in its
original checkout; inspect Worker files read-only and let the assigned Worker
make product changes. Successful verification precedes durable completion_ack,
which precedes cleanup. Preserve cleanup_pending without discarding verification.
Respect project branch protection and the human release gateway. On
`TeammateIdle`, inspect the current assignment before re-waking the same
teammate; do not silently replace it. See docs/workspace-integration.md for
current implementation boundaries.

## Escalation

- Requirement change: pause via Runtime; follow `docs/rules/change-control.md`.
- Specification conflict: report during phase one; do not improvise.
- Blocking finding: use `.claude/skills/bug-resolution/SKILL.md` to investigate root cause before repair.
- Unclear next action: use `.claude/skills/loop-orchestration/SKILL.md`.
- Human-controlled or irreversible action: surface the matching Gateway type.

## User-visible progress

The recovery packet is a scheduling checkpoint, not a work report. At the
start of work, after a meaningful deliverable, on a blocker or change of
approach, and approximately every 60 seconds during sustained work, tell the
user briefly what deliverable or evidence is now accepted, what action is underway, and what comes
next. Use the user's language. Do not narrate every tool call or expose
internal reasoning. Report only observed results: a revision increment, an
intended action, or a repeated Hook message is not evidence of progress.

Hook `systemMessage` is a user notice; model recovery instructions belong in
`hookSpecificOutput.additionalContext` for supported events. Full recovery
context is retained at session/agent start and stage transitions. Ordinary
PreToolUse calls provide the current gate conflicts, missing work, and a
compact checkpoint. Repeated user notices may be suppressed within one
session/agent, but safety denials and model context must never be suppressed.
A Worker plan checkpoint requiring no approval does not silence progress
reports to the user.

## Bounded planning continuation

S3/S4 work does not require another user approval merely because a stage
changed or drafting is substantial. Main Stop provides one continuation
reminder when a locked REQ has a current unblocked planning milestone and
no delegated agent. `stop_hook_active` always permits the next Stop; this
is not an unattended scheduler or proof of completion. Honor explicit user
stop requests and explain actual permission/external waits. S2 design
sign-off and other stages retain their existing domain-specific rules.

## Optional bounded repair authority

For a new bound REQ, an explicit finite human grant may delegate ordinary S8 contract reviews to the Driver under docs/bounded-repair-autonomy.md. Do not infer this grant from the approver name, install it on old REQs, or manufacture human_decision for technical review. No grant preserves the existing explicit contract approval requirement. S2 ADR sign-off, requirement changes, budgets, external permissions and release retain their own boundaries. Within a valid grant, recoverable task/scope-planning issues go to the Driver, not repeated user A/B/C menus.

## Approved repair policy at new requirement binding

If installation recorded a human-approved project repair policy, include its
unchanged path and SHA256 using `req bind --repair-policy` and
`--repair-policy-sha256`. Do not create or enlarge a policy from an account name.
See `docs/bounded-repair-autonomy.md`. In-scope technical RepairContract approvals
then use the Driver's registered review and pinned policy, without another human
approval. Missing authority, business changes and release keep their Gateway.
Existing bound runtimes are not retroactively granted this authority.

## 临时 worktree 与阶段交付

REQ 绑定必须显式提供 `--dev-branch <开发主分支>` 和 `--release-upstream <最终发布上游>`，远程发布目标包含 remote；没有 develop 默认值。项目根目录是当前 REQ 唯一权威。

派发前将上游正式 Markdown、代码和测试产出整理提交；不要自动提交用户无关变更。新 worktree 不含主会话未提交或仅暂存的文件。主会话用 `runtime workspace prepare --assignment <id> --agent <owner> --root <项目根目录>` 准备 native Worker，再使用返回的 launch 命令；实际 checkout 由原 registry 创建。不要在它之前另跑 `worktree-create`，也不要依赖平台默认分支或复制目录。明确不入 Git 的 evidence 单独按依赖交接，不能复制整个控制面。

子会话在子分支提交成果并报告，主会话在项目根目录执行 `runtime task-integrate --assignment-id <id>`，生成合并提交、校验、接收并及时清理。集成不是 release；临时 worktree 不是交付终点。未回收、分支偏离、积压只提醒，不新增 Stop 或普通工具硬门禁。

Gate 的输入来源由 loop-definition 的 file_sources 契约声明；正式交付读固定 Git tree，明确运行输入读磁盘，Runtime 读权威快照。未提交产出不能帮助阶段通过。允许汇总的 evidence 按 mutable_evidence_kinds 自动同步对应 SHA256；同步不改变结论或代际，也不抹除产品基线漂移。

## Shared data and contract reading

For S3 converge the project data model and SYNC protocol before deriving FE/BE
responsibilities. Use [shared-model contracts](docs/rules/shared-model-contracts.md).
For S6 start from the TASK's ordered file links, scope and closing assertions;
read only the relevant contract/protocol/model slices. Register committed model
inputs in the existing baseline, and require real consumers to validate or
generate from them. Do not create a per-REQ model copy or an extra registry.

### S9 repair questionnaires

After the approved RepairContract, original Session/RepairPlan and Builder
dispatch, use the Hook's exact PlanReport view and answer paths. Fill and save
only the professional plan, red pre-fix checks and proposed paths. The original
PlanReport consumer validates the answer and projects the existing
plan-checkpoint/AgentBegin flow; do not send another generic PLAN_REPORT or
reconstruct a packet path or command.

Once the original S9 graph reaches `fixing`, implement only the approved repair
scope. Use the Hook's exact RepairResult view and answer paths to report the
professional verdict, frozen assertion results, unit results, evidence, scope,
risk, migration and rollback. Saving invokes the original RepairResult
consumer to validate the professional answer and the original candidate
consumer to check actual changed paths, Git state and required checks, then
project this same professional result onto the existing completion record. Do
not send a separate generic completion report or infer completion from green checks. Main retains
the original candidate integration, required Main/Git checks, configured human
decisions, independent targeted verification and fresh complete S7 round.
Never guess paths or run manual `stage-answer issue/consume` commands.
