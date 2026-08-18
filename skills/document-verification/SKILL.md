---
name: document-verification
description: Use when S5 document verification assigns you one review responsibility — write the envelope first (REV-template §0), findings only when found
category: methodology
version: 2.0.0
---
# Document Verification

## Authority
You produce evidence; the gate evaluates it. Runtime authority lives in `docs/loop-definition.json`; the S5 stage contract lives in `docs/agent-protocol.md #s5`（三步：派活→审查→收口三岔路）; the envelope skeleton and findings format live in `docs/reports/review/REV-template.md`.

## Role Contract

You are one of two independent reviewers of the frozen spec chain. You verify
exactly one responsibility — `DV-SPEC-CONSISTENCY`（规格链自洽吗？）or
`DV-TASK-EXECUTABILITY`（Builder 照着能干吗？）— assigned in your activation
envelope. You never repair what you review.

## Entry Conditions
- The Loop is in `document_verification` and you hold an activation envelope naming exactly one responsibility.
- The spec chain carries fingerprints (path + version + SHA-256 + baseline generation).

## Required Inputs

Activation envelope (responsibility, report path, fingerprints, skills to
load), the locked spec chain (REQ → design → contracts → tasks), and the
module current-truth package when the REQ touches UI.

## Procedure — 3 mandatory, 1 triggered

1. **落信封骨架（激活后第一件事）**：复制 `docs/reports/review/REV-template.md` §0 到你的报告路径 `docs/reports/review/REV-{runid}-{resp}.json`——每个字段带一行填写指引，写骨架即读懂要交什么。`subject_refs` 从 `.claude/loop-state.json` 的 documents[] 逐条复制（故意没有自动命令——逐条抄写就是"我签的是哪一版"的对峙，这一步的笨拙是审查的锚）。
2. **审查（按职责）**：
   - SPEC-CONSISTENCY：自底向上读 TASK → 主契约 → 关联契约 → locked REQ → 设计/rules，核对验收↔条款映射、跨文档引用指纹、FE/BE/SYNC 边界一致（数据形状/错误码/状态机/API 面）、场景包与契约映射不矛盾。
   - TASK-EXECUTABILITY：跑 `go run ./cmd/loop-harness tasks check --root .` 消费机检结论（覆盖双向/DAG 机器已判，不重算），再审机器判不了的三件——每个 TASK 的收尾契约可判定且现有工具能产出其证据、粒度单一、写路径冲突有显式串行归属。
3. **收口**：信封回填 `conclusion`（pass / fix_required / req_change_required——与 gate 同词，全流程没有第二套枚举）与 `requested_event`（仅 fix_required 填 document_fix_required）。此后不调用任何 transition 命令——PreToolUse 按你的 conclusion 自动路由。
4. **触发——有 finding 才写 REV 报告**：按 `REV-template.md` §1-§5 写（findings 表带 P0-P3/定位/预期/实测/证据；N/A 须记理由与证据）。双 pass 不产报告——没有人读"都挺好"。

## Outputs

- 必产物：`REV-{runid}-{resp}.json` 证据信封（12 字段，机器校验）。
- 条件产物：`REV-{runid}-{resp}.md` findings 报告（仅有 finding 时）。

## Exit Conditions

- 信封 conclusion 已回填且 subject_refs 精确等于当前 documents[] 指纹集（多一少一 gate 都拒）。

## Stop Conditions

Stop immediately and surface to the human if any of:

- A fingerprint changed mid-review (the artifact was edited after you started).
- A required document layer is missing entirely.
- Two authorities contradict and cannot be reconciled without a REQ decision → `req_change_required`.
- You recognize you authored (or materially drafted) an artifact under review — independence is lost（纪律层：机器无法替你判这一条，你不说没人知道——这是 S5 对你的唯一诚信要求）。

## Non-Goals

- Do not repair the reviewed documents — your fix_required routes them back to planning.
- Do not lock the batch or activate Builders — TR-003 does that after both reviewers pass.
- Do not treat missing coverage as N/A, and do not accept an unverifiable closing contract as executable.
