# Canonical BUG: BUG-CX-13

> Status: reported
> Severity: P0
> Runtime ref: N/A（模板仓库自审——S5 fixture 与可达性）
> Found in review round: s5-review-1
> Finding evidence: 本文件 §2/§3 所引 file:line（HEAD=66b7511 实测）
> Original responsibility: tests/fixtures/req039 / tests/system/req039 / planner gate (evaluator evaluatePlanningDesign)
> Owner: PM / Architect

## 1. Fingerprinted Specification Chain

| Order | Kind | ID | Path | Version | SHA-256 | Relevant clauses |
|:--|:--|:--|:--|:--|:--|:--|
| 1 | BUG | BUG-CX-13 | `docs/reports/bugs/BUG-CX-13.md` | v1 | n/a | all |
| 2 | code | evaluator | `internal/qualitygate/evaluator.go` | current | n/a | evaluatePlanningDesign |
| 3 | rule | loop-definition | `docs/loop-definition.json` | current | n/a | PTR-PLAN-01 |
| 4 | fixture | req039 | `tests/fixtures/req039/fixtures.go` | current | n/a | SeedPlanningDesignComplete |
| 5 | design | L3-S5 | `blueprint/L3-S5-document-verification.md` | current | n/a | S5 fixture |

## 2. Observed Contradiction

**症状：S2 design 出链堵死 + S5 fixture 仍手工播种（CX-07 教训同款）。**

| Field | Value |
|:--|:--|
| expected | S2 design 出链（planning.design→contracts）纯有机可达（disk-declared + commit-time register）；S5 fixture 在去除手工 documents/evidence 播种后真实可达 |
| observed | **① design gate 无磁盘回退**：`evaluatePlanningDesign`（evaluator.go:188-211）从 documents[] 查 kind=req/locked 与 kind=design/locked，**未走 diskDeclaredArtifacts**（磁盘回退只加在 evaluatePlanningArtifact，evaluator.go:235）。**② PTR-PLAN-01 actions=[]（loop-definition）**——没有任何动作把 ARCH-* 文档登记进 documents[]。**③ SeedPlanningDesignComplete 手工播种**（fixtures.go:836-861）ARCH-039 + 写 evidence index——与 S3/S4 fixture 已按 CX-07 教训改为"磁盘+commit 时登记"不同步（planning_chain.go:29-30,52 明确注释"不再手工种 documents"），S5 fixture 没改。**④ spine S2→S3 步一直依赖 fixture 手工播种**：`RequireLifecycleTransition`（hook_helpers.go:135）声称 PRE-toolUse 自动 commit，但纯有机链上 design gate 永不通过——`ARCH-039` 无 register 路径，而 seeded ARCH-039 来自 fixture 手工注入。后果：spine 测试在 S2 出链这一段**实际验证的是 fixture 播种 + 手工 transition 的组合**，而非真正的 hook auto-advance |
| user/data/system impact | 这是 CX-07 修复的**深化**——S2 design 出链与 S5 fixture 同属"fixture 掩盖可达性"模式，spine 测试过去从未在纯有机链上跑通 S2 出口（与 9cd52fa 修复后 f2c23a4 记录的"spine 首次 PASS"是同一性质——它把 S3/S4 段改干净了但 S2 出链未改）。一旦遇到真实项目无 fixture 播种，S2 永远卡住 |
| reproduction | 在全新目录跑 hook-driven spine（无 fixture 播种）：走到 planning.design→PreToolUse → GATE-PLANNING-DESIGN-COMPLETE 报 missing `document:design:locked`，无任何"在磁盘上写 ARCH.md"的提示 |

## 3. Root Cause Investigation

| Hypothesis | Evidence | Result |
|:--|:--|:--|
| H1: design 文档登记路径从未被设计——PTR-PLAN-01 actions=[]；registerDocumentsFromDisk 只支持 contract/task 两种 kind（actions.go:373），无 design | loop-definition.json:191 actions:[] + actions.go:373 actionKind 参数 | confirmed |
| H2: design gate 修 CX-07 时只补了 contract/task（evaluatePlanningArtifact 内部已分 kind），evaluatePlanningDesign 未同步 | evaluator.go:188-211 走的是 currentDocuments，未走 diskDeclaredArtifacts | confirmed |
| H3: S5 fixture 与 planning_chain 各自修改，S5 fixture 没改 | fixtures.go:836-861 直接 AppendEvidence+documents | confirmed |
| H4: spine 测试的"首次 PASS"主要靠 fixture 掩护 | spine 跑 S2 出链步（s2_to_s11_hook_driven_test.go:34-46）显式调 SeedPlanningDesignComplete 注入 ARCH-039；不注入则不通过 | confirmed |

Accepted root cause: **CX-07 修复是部分修复**——只补了 contract/task 两类的磁盘回退与 TR-002 幂等补登，但 **PTR-PLAN-01 无登记动作 + design gate 无磁盘回退**这两条链上游的可达性未动；S5 fixture 同步缺位使得 spine 测试无法真实暴露。深根因仍是**机制左移时"自然路径可达性"未被验证**——同一错误模式在 S2 design 出链复发。

## 4. Closing Contract

### 4.1 Repair scope

- **① 给 PTR-PLAN-01 加 register_design_document 动作**（或扩展 registerDocumentsFromDisk 接受 kind=design），与 contract/task 同步登记 ARCH-* 文档入 documents[]（author_agent_id=hook_controller，version 与 status=locked 同步）
- **② evaluatePlanningDesign 加磁盘回退**（复用 diskDeclaredArtifacts，目录 = docs/design/architecture，识别 ARCH-* 文件 + status/locked；与 contract/task 同结构）
- **③ S5 fixture 改造**——删除 SeedPlanningDesignComplete / WriteDocumentVerificationPassEvidence 的手工 documents/evidence 播种，对齐 planning_chain.go 的 CX-07 改法：ARCH-* 文件写入磁盘 + PTR-PLAN-01 自动登记，document_review_record 由 `runtime evidence add` 路径生成
- **④ spine 测试覆盖纯有机链 S2→S5**（现有测试已对 S3/S4 改干净，对 S2/S5 部分补一层）
- **⑤ 文档**：protocol #s2 增"PTR-PLAN-01 落盘与登记"的产出物明示；agents/document-verifier.md 减"手工播种"的隐式依赖

### 4.2 Forbidden scope

- 不改 kind=design 的指纹算法（复用 ARCH-*.md 的 SHA-256）
- 不让 design 登记走 planning_chain 的 contract 路径（design 与 contract 是不同维度）

### 4.3 Before-fix evidence

§2 所引 file:line + 纯有机 spine 复现步骤。

### 4.4 Retest contract

```text
assert evaluatePlanningDesign 走磁盘回退（新增 pin 测试，参照 contract/task 的 TestPlanningGatesReadDiskDeclaredArtifacts）
assert PTR-PLAN-01 提交后 documents[] 含 kind=design
assert spine S2→S5 全程纯有机可达（去 fixture 播种）
```

## 5. Acceptance And Repair

| Field | Reference |
|:--|:--|
| BUG acceptance evidence | pending |
| repair assignment | pending |
| Builder activation | pending |
| repair fingerprint | pending |
| impact analysis | pending |
| invalidated evidence | n/a |

## 6. Verification

| Verification | Owner | Result | Evidence |
|:--|:--|:--|:--|
| 纯有机链可达性 E2 | | pending | — |

## 7. Deduplication And History

Canonical BUG: BUG-CX-13（S2 design 出链与 S5 fixture 同步族；CX-07 修复的深化——S2 出链堵死 + fixture 掩盖可达性）

| Date | Event | Actor | Runtime revision | | |
|:--|:--|:--|:--|:--|
| 2026-08-17 | reported（S5 机器链审查，主会话亲证） | 主会话+sub-agent | n/a | 本文件 |