# Canonical BUG: BUG-CX-07

> Status: reported
> Severity: P0
> Runtime ref: N/A（模板仓库自审——第二轮 agent 视角复杂度审查；所有结论经主会话亲自核验）
> Found in review round: complexity-review-2
> Finding evidence: 本文件 §2/§3 所引 file:line（HEAD=f2c23a4 实测）
> Original responsibility: qualitygate evaluator / transition actions / guards 文案
> Owner: PM / Architect

## 1. Fingerprinted Specification Chain

| Order | Kind | ID | Path | Version | SHA-256 | Relevant clauses |
|:--|:--|:--|:--|:--|:--|:--|
| 1 | BUG | BUG-CX-07 | `docs/reports/bugs/BUG-CX-07.md` | v1 | n/a | all |
| 2 | rule | loop-definition | `docs/loop-definition.json` | current | n/a | PTR-PLAN-02/TR-002/gates |
| 3 | code | evaluator | `internal/qualitygate/evaluator.go` | current | n/a | evaluatePlanningArtifact |
| 4 | design | L3-S4 | `blueprint/L3-S4-task-split.md` | v4.0.3 | n/a | 推进链 |

## 2. Observed Contradiction

**症状 A（自动路径不可达）：planning 阶段两个门要求的事实只能由被门放行的转换自己产生——hook 自动推进在 planning 阶段疑似整体死锁。**
**症状 B（死锁残留入口+虚假承诺）：契约 Status 字段缺失时登记静默 skip；TR-002 的 tasks 阶段报错承诺"下一个 PreToolUse 会重跑登记"，机械上不存在该路径。**

| Field | Value |
|:--|:--|
| expected | 门的前置事实（documents[] 注册）在门评估前就可达产；或门直接读 agent 能产出的磁盘事实；被拒报错指向可达动作 |
| observed | **症状 A**：`evaluatePlanningArtifact`（internal/qualitygate/evaluator.go:203-226）对 GATE-PLANNING-CONTRACTS-COMPLETE / GATE-PLANNING-TASKS-COMPLETE 只扫描 `documents[]` 中 kind=contract/locked、kind=task/complete 的条目，缺失即 NOT_READY（missing=`document:contract:locked` / `document:task:complete`）；而注册动作 `register_locked_contracts` 挂在 PTR-PLAN-02（loop-definition phase transition，from=contracts）、`register_planning_tasks` 挂在 TR-002——都只在**被门放行后的 commit 里**执行（actions.go:93 起，全仓无其他调用点）。门要求转换的产物作前置 → 纯自动路径鸡生蛋。佐证：所有 planning E2E（contracts_e2e/tasks_e2e/dual_track）都靠手动 `runtime transition` 推进——恰是 protocol 明令 agent 禁用的命令（agent-protocol.md:70,158）。missing 词汇（`document:task:complete`/`planning_task_record`）在 docs/、skills/、loop-harness.md **零出现**（grep 亲证 0 命中）——agent 收到 NOT READY 时无从知道该造什么。**症状 B**：`registerDocumentsFromDisk` 对无 Status 字段的文件静默 skip（actions.go:447-451"declares no batch membership … skip"）→ 契约缺 Status 时 PTR-PLAN-02 提交且登记 0 条 → tasks 阶段 TR-002 拒 → f2c23a4 的 phase 分支文案（guards.go:433）说"flip the contract markdown Status to locked … and the next PreToolUse re-runs the registration"——但 PTR-PLAN-02 from=contracts，tasks 阶段永不再评估，`register_locked_contracts` 无任何重跑路径（亲证：actions.go:93 唯一注册点）→ 同一报错无限循环 |
| user/data/system impact | 这是"真实使用一直卡在 planning"的直接结构性根因之一：agent 做完全部正确的事，hook 永远 NOT READY，且报错词汇无文档、修复文案指向不存在的动作 |
| reproduction | 症状 A：绑定 REQ → 走到 planning.contracts → 写全四契约（Status: locked）→ 触发任意 PreToolUse → gate 报 missing `document:contract:locked`，documents[] 恒空。症状 B：契约文件删掉 `状态：` 行 → PTR-PLAN-02 提交（登记 0）→ tasks 阶段按报错 flip Status → 下一个 PreToolUse 无事发生 |

## 3. Root Cause Investigation

| Hypothesis | Evidence | Result |
|:--|:--|:--|
| H1: 门的前置事实与注册动作时序倒置——设计时假设"注册先行"，但注册被挂在被门转换的 actions 里 | evaluator.go:203-226 纯查 documents[]；注册仅在 PTR-PLAN-02/TR-002 commit 路径（loop-definition + actions.go:93 唯一注册表）；无任何播种/housekeeping 路径 | confirmed |
| H2: 自动路径从未被端到端测试，全部 E2E 用手动 transition（禁用命令）绕过了它，故缺陷不可见 | contracts_e2e/tasks_e2e/dual_track 的推进语句均为 `runtime transition --id PTR-PLAN-02/TR-002`；无任何"仅靠 PreToolUse 走完 planning"的测试 | confirmed |
| H3: requested-event 旁路可解 | `qualifiedRequestedEvents` 在单候选 cursor 早退（evaluator.go:379-382 `len(candidates)<2 → nil`）；planning 各 phase 恰为单候选 | confirmed（旁路不可用） |
| H4: 症状 B 是 f2c23a4 修复文案的错误承诺，非新机制缺口 | 文案承诺的重跑路径不存在（PTR-PLAN-02 from=contracts 硬约束）；静默 skip 入口在 v4.0.2 就有 | confirmed |

Accepted root cause: **门（评估侧）与登记（转换侧）的前置时序从未做过可达性验证**——"gate requires what the gated transition produces"。深层原因与 BUG-CX-04 同源但更深：v4 把 planning_complete guard 改成了读磁盘事实（disk-declared complete/locked），却漏了同一条链上游的 qualitygate 仍是 documents[] 前置；而测试全部走手动迁移，使自动路径的可达性从未进入回归面。症状 B 是同一时序倒置在报错文案上的投影（承诺了一个架构上不存在的重跑）。

## 4. Closing Contract

### 4.1 Repair scope

方向（owner 已裁或待裁，按工程建议）——**门的前置改读磁盘事实，登记保持为转换职责**（与 planning_complete guard 的 v4 哲学一致：agent 产文件，机器管登记）：
- `internal/qualitygate/evaluator.go` evaluatePlanningArtifact：documents[] 查不到时，经 `input.Files.ReadFile` 读 `docs/contracts/*.md` / `docs/tasks/TASK-*.md` 的顶部 Status（复用 ParseMarkdownField 语义），磁盘声明 locked/complete 即视为事实满足（指纹核验仍在 commit 后由登记+reachability 承担）；missing 文案同步改为磁盘动作（"flip the contract markdown Status to locked"）
- `internal/transition/guards.go:433` 文案随门语义修正（不再承诺 PreToolUse 重跑）
- 可选加固：TR-002 actions 增幂等 `register_locked_contracts`（同代替换语义已支持，engine appendDocument 注释明说合法）——封症状 B 的静默-skip 入口
- 新增测试：**仅靠 hook/PreToolUse 语义走完 planning.contracts→tasks→document_verification** 的可达性 E2E（不用 runtime transition）——把自动路径拉进回归面
- `loop-harness.md` TR-002/PTR-PLAN-02 小节补一句磁盘 Status 是门的前置（missing 词汇不再无文档）

### 4.2 Forbidden scope

- 不改 selector/gate 结构（SEL-PLANNING-OUTCOME）
- 不放开 agent 手动 transition 的禁令（正解是自动路径可达，不是绕行合法化）
- 不把注册从转换挪到 gate 评估路径（评估器保持纯函数；登记仍是有 journal 的 commit 动作）

### 4.3 Before-fix evidence

§2 所引 file:line + 复现两条（主会话亲证：evaluator 纯查 documents[]；actions.go:93 唯一注册点；missing 词汇文档 0 命中；单候选早退）。

### 4.4 Retest contract

```text
assert 端到端 E2E：写齐磁盘契约（Status: locked）→ PreToolUse 推进 PTR-PLAN-02，无手动 transition
assert 同链走完 tasks → TR-002 自动 commit 且登记 ≥1 contract/task
assert 无 Status 契约不再静默 skip（登记 0 时 PTR-PLAN-02 报错指路，或 TR-002 幂等补登）
assert TR-002 报错文案的动作可达（模拟执行后有状态变化）
assert missing 词汇在 loop-harness.md 有解释
```

## 5. Acceptance And Repair

| Field | Reference |
|:--|:--|
| BUG acceptance evidence | pending（owner 拍板修复方向） |
| repair assignment | pending |
| Builder activation | pending |
| repair fingerprint | pending |
| impact analysis | pending |
| invalidated evidence | n/a |

## 6. Verification

| Verification | Owner | Result | Evidence |
|:--|:--|:--|:--|
| 自动路径可达性 E2E（无手动 transition） | 待派 | pending | — |

## 7. Deduplication And History

Canonical BUG: BUG-CX-07（planning 推进链可达性；BUG-CX-04 修复的深化——同一时序倒置的架构层）

| Date | Event | Actor | Runtime revision | Evidence |
|:--|:--|:--|:--|:--|
| 2026-08-17 | reported（二轮复杂度审查 N1/N2，主会话亲证） | 主会话+sub-agent（复核） | n/a | 本文件 |
