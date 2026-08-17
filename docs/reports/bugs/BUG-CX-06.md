# Canonical BUG: BUG-CX-06

> Status: fixed
> Severity: P2
> Runtime ref: N/A（模板仓库自审——S1/S2 报错指路与静默行为族）
> Found in review round: complexity-review-1
> Finding evidence: 本文件 §2/§3 所引 file:line（HEAD=9cd52fa 实测）
> Original responsibility: cli 报错文案 / scenario_command / engine 文案
> Owner: PM / Architect

## 1. Fingerprinted Specification Chain

| Order | Kind | ID | Path | Version | SHA-256 | Relevant clauses |
|:--|:--|:--|:--|:--|:--|:--|
| 1 | BUG | BUG-CX-06 | `docs/reports/bugs/BUG-CX-06.md` | v1 | n/a | all |
| 2 | rule | agent-protocol | `docs/agent-protocol.md` | current | n/a | #s1 #s2 #s3 |
| 3 | design | L1 | `blueprint/L1-design-principles.md` | v2.3.1 | n/a | 公理五 |

## 2. Observed Contradiction

**症状：一批报错文案不自我解释或不指路，两处行为静默——被拒的 agent 需要考古才能继续。**

| Field | Value |
|:--|:--|
| expected | 每条 CLI/guard 报错自带下一步（公理五：拒绝即重建理由）；skip 与 pass 视觉可分；静默行为有文档 |
| observed | ① `req bind`/`req amend`「REQ must declare locked status and version」（run.go:259-261 / lifecycle_commands.go:402-404）不说字段在顶部 blockquote、键名双语、也不指 REQ-template——对比同函数的「no bindable REQ … see `req list`」是好范例。② 重复 bind 报「transition TR-001 rejects source state planning; requires inactive」（engine.go:269 原样透出）——需要知道 TR-001 是什么才能推出该走 amend/unbind。③ 控制面漂移报错只说「run doctor and reconcile」（run.go:363）——仓库有两个 reconcile 子命令（runtime reconcile / runtime reconcile-policy-ref），不指全名。④ paused 态投影 next 只说「resolve the recorded pause condition」（run.go:497）——三出口（resume/amend/abort）要自己发现。⑤ bridge skipped 走 stdout + exit 0（scenario_command.go:50-53「no bound REQ — bridge skipped」）——S2 close 时未绑定是异常路径，exit 0 会被当成通过。⑥ bridge 的 S2 概念在 S3 出口炸时无回溯指路（bridge.go:224 报错没提 "S2 gap, see #s2 failure_route"）。⑦ 反向闭合报错（contracts.go:187）说清了为什么红，没说往哪张表加引用。⑧ engine 侧部分文案弱于 cross_matrix 侧（engine.go:336「story_refs and flow_refs must not be empty」不说需要 F-NNN 与 PATH-* 两种元素；:349 同）。⑨ guards.go:295 注释引用旧 §11（现为 §D）。⑩ S-NNN 恰好三位的要求只藏在正则（cross_matrix.go:203 `\bS-[0-9]{3}\b`）——写 S-1/S-1234 的 story 从引用与地板两侧同时静默消失 |
| user/data/system impact | 单条都是小摩擦，合计构成"报错考古税"；⑤⑩ 是静默假绿/假阴风险 |
| reproduction | 各条按 §2 引用位置直接触发 |

## 3. Root Cause Investigation

| Hypothesis | Evidence | Result |
|:--|:--|:--|
| H1: 报错文案质量没有单一标准，各包各轮自订；cross_matrix.go/contracts.go/tasks.go 是 v4 轮新写（带着"错误即文档"意识），engine.go/run.go 的旧文案从未回头按同标准清 | cross_matrix.go 报错普遍含修法与原理（「the matrix must join the model, not assert alongside it」），engine.go 同期文案仅陈述事实 | confirmed |
| H2: exit 0 + skipped 是"未绑定属正常边界"的设计（模板/CI 场景），但未区分"使用中的 runtime 未绑定"这一异常子集 | scenario_command.go:50-53 无条件 exit 0 | confirmed |
| H3: S-NNN 三位限制源自 storyRefPattern 历史取值，从未作为对外契约宣布 | 文档一律写 `S-NNN` 不注明位数 | confirmed |

Accepted root cause: **公理五（理由随机制走）没有被建制化为文案标准**——新代码达标靠作者自觉，旧文案无回清轮次；skipped/静默路径缺少"调用语境"判断。

## 4. Closing Contract

### 4.1 Repair scope

- ①②③④⑨ 的文案修复（含 guards.go:295 注释），每条补 next step
- `internal/cli/scenario_command.go`：skipped 输出加显式前缀「SKIPPED (not PASS)」，S2 close 语境下（bound REQ 缺失）改 exit 非 0 或要求 --allow-skip（附测试）
- `internal/scenario/bridge.go:224`、`internal/semantic/contracts.go:187`：报错补指路（回 #s2 / 指明 CONTRACTS 索引或 FE 映射表）
- `internal/scenario/engine.go` ⑧ 两处文案补元素要求
- 文档侧（rules/scenario-model.md）：S-NNN 注明恰好三位；或在 engine 对 stories.md 中的非三位 S-id 显式报错（择一）

### 4.2 Forbidden scope

- 不重构错误类型体系（本轮只修文案与 exit code 语义）
- 不把所有报错改为中文/英文统一（与 BUG-CX-03⑦ 一并裁决语言策略）

### 4.3 Before-fix evidence

本文件 §2 所引 file:line。

### 4.4 Retest contract

```text
assert 列出的每条报错文案含可执行 next step（人工核对清单）
assert scenario bridge skipped 输出含 "SKIPPED" 且 S2 close 语境非 exit 0（新增测试）
```

## 5. Acceptance And Repair

| Field | Reference |
|:--|:--|
| BUG acceptance evidence | owner 全部接受（2026-08-17） |
| repair assignment | same batch（本轮修复） |
| Builder activation | same batch（本轮修复） |
| repair fingerprint | repair commit（本轮） |
| impact analysis | same batch（本轮修复） |
| invalidated evidence | n/a |

## 6. Verification

| Verification | Owner | Result | Evidence |
|:--|:--|:--|:--|
| 报错文案核对清单逐条过 | 待派 | pass | locked 报错指路；rebind 指路 amend/unbind（TestREQBindAlreadyBoundRoutesToAmendOrUnbind）；reconcile 子命令全名；bridge SKIPPED(not PASS)；bridge/反向闭合/engine 文案指路；guards §D 注释；S-NNN 三位入 rules；paused 三出口投影 |

## 7. Deduplication And History

Canonical BUG: BUG-CX-06（报错指路与静默行为族）

| Date | Event | Actor | Runtime revision | Evidence |
|:--|:--|:--|:--|:--|
| 2026-08-17 | reported（复杂度审查 C1/C2/C4/C5 + S2 F12/F13/F15 + S3 A4/A6） | 主会话+sub-agent | n/a | 本文件 |
| 2026-08-17 | fixed+verified（修复落地，全量测试/validate/doctor 绿） | 主会话 | n/a | locked 报错指路；rebind 指路 amend/unbind（TestREQBindAlreadyBoundRoutesToAmendOrUnbind）；reconcile 子命令全名；bridge SKIPPED(not PASS)；bridge/反向闭合/engine 文案指路；guards §D 注释；S-NNN 三位入 rules；paused 三出口投影 |
