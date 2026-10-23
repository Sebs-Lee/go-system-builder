# Task: TASK-{id}

> Status: draft
> Reading policy: linked-v1
> Dispatch policy: waves-v1
> （document status: draft=still writing；complete=the planning document is finished and eligible for its original gate——required across the whole batch at TR-002, says nothing about implementation；cancelled=out of the batch, its §3 clause declarations drop out of coverage so any gap resurfaces. Responsibility progress uses the separate five-state model in `docs/rules/state-machine.md`; do not infer implementation `done` from this line.）
> Version: v1.0.0
> Source REQ refs: REQ-{id} / none
> Module current truth: `docs/design/prototypes/{module}/` / N/A
> Primary contract: {FE/BE/SYNC-id}
> Closing Contract: TASK-{id}#closing-contract
> Execution: runtime owns assignment, owner, responsibility status and integration; do not fill them into this frozen TASK.

## 1. Objective

{One testable objective and user-visible value.}

<!-- 用一句话解释本任务的交付目标，再按独立完成判据、责任和真实变更耦合判断是否拆分。
结合 §2 读集、§4 写入范围及执行中的 compact/返工观察调查粒度；句式、文件数或一次 compact 不能单独决定拆分，也不能据此省略必要输入。拆分纪律全文见 specification-planning step 11。 -->

## 2. Document Manifest

Read order for the builder. Fingerprints, versions, and lock state live in
runtime documents[] (.claude/loop-state.json) — this table never hand-copies
them.

| Order | Kind | ID | Path | Clauses | Purpose | Mode |
|:---|:---|:---|:---|:---|:---|:---|
| 1 | contract | {contract-id} | [Local responsibility](../contracts/{contract-id}.md#{stable-anchor}) | §{n} | Understand this delivery | required |
| 2 | sync | SYNC-{id} | [Operation and recovery](../contracts/SYNC-{id}.md#{stable-anchor}) | §{n} | Understand shared behavior | required |
| 3 | model | {native-definition} | [Shared model entry](../../architecture/data-model/MODEL-{id}.md) | {definition} | Locate the exact data source | required |

Remove inapplicable rows; keep the smallest complete ordered reading set. Use
`conditional` only with an explicit condition in Purpose, and `optional` for
background. Follow [shared-model reading rules](../../rules/shared-model-contracts.md).
Do not add a foundation dependency unless another task actually produces an
input this implementation needs. Returning to this TASK closes the reading path.

Repair assignments prepend the canonical BUG as order 1 and shift the remaining
documents. The request remains the authority for exact order.

## 3. Delivered Clauses

Which clauses of the primary contract this TASK delivers. The CONTRACTS index
is the clause universe; `tasks check` aggregates these declarations against it.

| Contract | Delivered clauses |
|:---|:---|
| {FE/BE/SYNC-id} | §{n}, §{n} |

An empty clause list means a support TASK — legitimate, but excluded from
coverage aggregation.

### 3.1 Module Impact

Modules touched: `{module}` / N/A. Scenario truth lives in the module package
(`docs/design/prototypes/{module}/`) — never create a per-REQ or per-round
scenario/spec copy. Any change to a current-module truth file triggers a full
module regression sweep.

## 4. Scope

| Type | Paths / Commands |
|:---|:---|
| read paths | `{path}` |
| prospective write paths | `{path}` |
| forbidden paths | `.claude/loop-state.json`, `{path}` |
| allowed command classes | test / lint / build / read-only |
| output paths | `{implementation/test/report paths}` |

This scope records the planned dispatch and integration-check boundaries for
the TASK. Hook events observe progress and provide reminders; they do not
authorize or deny individual tool calls by identity, path, or scope.

## 5. Selected Skills

Record additional skills needed for this TASK's actual risk or responsibility.
The original role and assignment supply default skills; do not repeat their
metadata here. Remove the example row when no additional selection is needed.

| Skill | Category | Source | Version | Applicability |
|:---|:---|:---|:---|:---|
| {skill} | best-practice | `.claude/skills/{skill}/SKILL.md` | {version} | {risk/responsibility} |

## 6. Outputs And Evidence

### 6.1 This TASK's assigned deliverables

Keep only outputs owned by this TASK's dispatched responsibility. A Builder
produces implementation, its scoped tests and its completion report; each
independent reviewer receives a separate original assignment.

| Output | Path | Acceptance |
|:---|:---|:---|
| implementation | `{path}` | {contract assertion} |
| tests | `{path}` | {behavior/failure coverage} |
| completion report | `{agent-message-path}` | schema-valid |

### 6.2 Subsequent consumption and independent verification

These responsibilities retain their original gates and required coverage.
They are not additional Builder outputs or permission to author another role's
evidence. Runtime owns the actual assignment and receipt references.

| Responsibility / consumer | Evidence or action | Acceptance |
|:---|:---|:---|
| Main / integration | bound development branch and integration receipt | original required checks pass after integration |
| Delivery Verifier | `docs/reports/review/REV-{id}.md` | assigned dimension result |
| QA | `docs/reports/qa/QA-{id}.md` | assigned dimension result |
| E2E | `docs/reports/e2e/E2E-{id}.md` | real-browser flow result |

## 7. Closing Contract

Cover observable behavior, this responsibility's own checks, its reviewed
scope, and a usable handoff. Use as many assertions as the actual conditions
need; the four examples below are not a line-count requirement. These are
professional completion criteria, not commands executed by the checker.

```text
assert {contract clause} == satisfied
assert {verification command} == pass
assert actual changes stay within the reviewed TASK and assigned write scope
assert completion_result explains each condition and any actual scope deviation
```

An actual deviation must be reported and resolved through the original scope
and change process; reporting it alone does not authorize a wider scope.

## 8. Dependencies

| Dependency | Required artifact / verification |
|:---|:---|
| TASK-{id} | {committed artifact integrated and verified on the bound development branch} |

依赖列只认 `TASK-*` 引用——只有 TASK 引用进入 DAG（`tasks check` 的环检测与拓扑）。assignment 级依赖不被机检追踪，需要跨任务顺序时写成 TASK 依赖或在收尾契约中声明。

## 9. Execution entry

Read the REQ's [overall plan](index-REQ-{id}.md) and
[dispatch rules](../../rules/dispatch-plan.md). Runtime owns lifecycle evidence.

### Resources

Declare only mutable external resources requiring exclusive use (not read-only
contracts/models). Use the same stable name across tasks; leave empty if none.

| Resource | Purpose |
|:---|:---|

## 10. Findings And Repairs

Findings, canonical BUGs, repair assignments and re-verification are runtime
records. Follow the BUG procedure and the execution entry; do not append
execution history to this frozen task document.

## 11. Document History

| Date | Version | Design change |
|:---|:---|:---|
| | | |
