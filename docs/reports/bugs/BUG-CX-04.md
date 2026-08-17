# Canonical BUG: BUG-CX-04

> Status: fixed
> Severity: P1（含一条 P0 级死锁路径）
> Runtime ref: N/A（模板仓库自审——S3/S4 agent 视角复杂度审查）
> Found in review round: complexity-review-1
> Finding evidence: 本文件 §2/§3 所引 file:line（HEAD=9cd52fa 实测）
> Original responsibility: agent-protocol S5.5 / 契约与任务模板 / guards 文案
> Owner: PM / Architect

## 1. Fingerprinted Specification Chain

| Order | Kind | ID | Path | Version | SHA-256 | Relevant clauses |
|:--|:--|:--|:--|:--|:--|:--|
| 1 | BUG | BUG-CX-04 | `docs/reports/bugs/BUG-CX-04.md` | v1 | n/a | all |
| 2 | rule | agent-protocol | `docs/agent-protocol.md` | current | n/a | #s3、S5.5 |
| 3 | template | TASK | `docs/tasks/TASK-template.md` | v4 | n/a | Status/§3/§8 |
| 4 | template | CONTRACTS | `docs/contracts/CONTRACTS-template.md` | current | n/a | 覆盖矩阵 |
| 5 | design | L3-S4 | `blueprint/L3-S4-task-split.md` | v4.0.3 | n/a | Status 三词 |

## 2. Observed Contradiction

**症状一（P0 级）：契约锁定时机文档与机器直接矛盾，可构成死锁。症状二：S3/S4 的机器契约（词表/机检/聚合语义）没有讲进 agent 可读层。**

| Field | Value |
|:--|:--|
| expected | 契约/TASK 的 markdown Status 何时置 locked/complete，文档与机器 guard 要求一致；每项机检在 agent 被拦前可知 |
| observed | ① **死锁链**：protocol S5.5（agent-protocol.md:282）明文「S5.5 is the only legal point at which contracts and TASKs transition from `*-draft` to `locked`」——但 PTR-PLAN-02 的 `register_locked_contracts`（actions.go:352-364）只登记磁盘上 status=locked 的契约；TR-002 的 planning_complete guard（guards.go:427）要求当前代存在已登记 locked 契约。保守遵守 S5.5 的 agent 留 draft → PTR-PLAN-02 提交但登记 0 条 → 进 tasks 阶段 → TR-002 报「run PTR-PLAN-02 first」——而 agent 被禁止手工推迁移（agent-protocol.md:70-72,156-159），且 planning.contracts→tasks 已迁移，PTR-PLAN-02 无法重发。② **`Status: complete` 语义陷阱**：TASK-template.md:3 只给词表不说何时用——工程直觉读成"实现完成"，规划期全填 draft 被 TR-002 拒；真实语义（"文档写完"）只在 blueprint/L3-S4:28（agent 不读）。③ **token 机检不可预知**：contracts check 的五 pattern 文法（contracts.go:14-22）挂 PTR-PLAN-02，但 protocol #s3 与 SKILL step 11 均不提——对比 step 13 明确写了「Run tasks check before requesting TR-002」，S3 无对称自检指令。④ **cancelled 剔除聚合**语义（tasks.go:91-99）在 TASK-template 缺解释——cancel 后突然冒漏覆盖红，agent 不知是设计行为。⑤ **索引 cell §n 与契约正文条款号无人讲需对齐**——CONTRACTS-template.md:48 讲 cell 记号，四个模板均无一字说 §n 必须与目标契约内"本合同条款 §n"一致（L3-S3:60 已承认自证循环并延后）。⑥ 依赖表非 TASK- 引用**静默不建 DAG 边**（tasks.go:21 只认 `^TASK-`；模板 §8 占位符 `{TASK/assignment}` 引导 agent 写 assignment-id）——以为声明了依赖，机器当没看见。⑦ 起草顺序两处矛盾（SKILL:82 FE→BE→SYNC vs protocol:227 BE,FE,SYNC）。⑧ index-template 看板 Status 列用 `pending`（index-template.md:32）与文档 Status 三词无交叉注解 |
| user/data/system impact | ① 是硬死锁（需人工介入 runtime 层才能解）；②-⑥ 造成"按文档做被机器拦、按机器做无文档可依"的往复；⑥ 使依赖声明形同虚设（无环检查缺边） |
| reproduction | 见 ①：守 S5.5 → PTR-PLAN-02 登记 0 条 → TR-002 拒且指路一个被禁止且不可重发的动作 |

## 3. Root Cause Investigation

| Hypothesis | Evidence | Result |
|:--|:--|:--|
| H1: S5.5 的 "locked" 指基线代际锁定（fingerprint 登记/generation），与 markdown Status 字段是两个正交概念，但复用了同一个词且从未区分 | S5.5 上下文（:276 表格）讲的是 atomic_lock=TR-003 的指纹登记；register_locked_contracts 读的是文件 Status；两个 "locked" 在 protocol 里无消歧 | confirmed |
| H2: TR-002/PTR-PLAN-02 的 guard 文案写于"迁移可以从 contracts 阶段重发"的假设下，S4 改为 phase 单向后未复核指路可达性 | guards.go:427「run PTR-PLAN-02 first」对 tasks 阶段的 agent 是不可达动作 | confirmed |
| H3: Status 三词/机检文法/聚合语义是 v4 收口轮（47031b5/49b3f2b）新钉的机器契约，模板与 protocol 只做了减法（删手抄列）没做加法（讲新词表语义） | TASK-template v4 diff：删三处 SHA 列、Status 8→3 词，但三词语义无处落地 | confirmed |
| H4: 依赖静默忽略是实现取巧（prefix 正则过滤天然跳过非匹配行），非有意设计 | tasks.go:21 `taskDepReference` 注释只说格式，未声明"忽略即无边"的语义后果 | confirmed |

Accepted root cause: 与 BUG-CX-03 同族——**机器契约（词表收敛、机检、登记语义）左移后，protocol/模板层未同步**；外加一个独立的词汇碰撞（S5.5 的 locked ≠ 文件 Status 的 locked）。H2 是文案-状态机不匹配：guard 报错没有随 phase 单向化复核"指路可达性"。

## 4. Closing Contract

### 4.1 Repair scope

- `docs/agent-protocol.md`：S5.5 行消歧——「S5.5 的 lock 指指纹/基线代际的原子锁定（TR-003 登记）；契约与 TASK markdown 的 Status 字段在定稿时即置 locked/complete，是登记动作的输入而非其产物」；#s3 actions 补「定稿后将契约 Status 置 locked；运行 `go run ./cmd/loop-harness contracts check` 自检」（与 step 13 对称）；起草顺序两处统一
- `internal/transition/guards.go:427`：报错按当前 phase 分支——planning.tasks 阶段改指「flip the contract markdown Status to locked, then re-run PTR-PLAN-02 path via the next PreToolUse」（或指路 protocol #s3）
- `docs/tasks/TASK-template.md`：Status 行改「draft (writing) / complete (document finished — required at TR-002) / cancelled (out of the batch; its clause declarations drop out of coverage)」；§8 补一行「只有 TASK-* 引用进入 DAG；assignment 引用不被机检追踪」
- `docs/contracts/CONTRACTS-template.md`：cell 记号处补「§n 必须与目标契约正文"本合同条款 §n"一致」；§n 廉价机检（grep 目标契约内 `§{n}` 存在性）加进 semantic.ContractsCheck（附测试）
- `internal/semantic/tasks.go`：§8 首列识别到非 TASK- 引用时发 problem（「dependency reference %q is not machine-tracked; only TASK-* ids join the DAG」）
- `docs/tasks/index-template.md`：看板 Status 列加注（进程状态，非文档 Status）

### 4.2 Forbidden scope

- 不改 PTR-PLAN-02/TR-002 的状态机与登记逻辑（机器侧已正确，是文档矛盾）
- 不恢复 Status 8 词表（v4 收口已裁决）

### 4.3 Before-fix evidence

本文件 §2 所引 file:line（agent-protocol.md:282,70-72 / actions.go:352-364 / guards.go:427 / TASK-template.md:3,95-97 / tasks.go:21,91-99 / CONTRACTS-template.md:48）。

### 4.4 Retest contract

```text
assert protocol S5.5 文本区分两种 locked
assert TR-002 失败文案在 tasks 阶段给出可达动作
assert TASK-template Status 行含 complete 的"文档写完"语义与 cancelled 的剔除语义
assert ContractsCheck 对 §n 指向不存在条款的 cell 报 problem（新增测试 fails_before_and_passes_after）
assert tasks check 对非 TASK- 依赖引用报 problem（新增测试）
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
| 模拟行走：守文档的 agent 能否无死锁通过 S3→S4→TR-002 | 待派 | pass | S5.5 双 locked 消歧；guards 文案 phase 分支；§n 存在性机检（TestContractsCheckFlagsClauseNumberDrift）；非 TASK 依赖显式报错（TestTasksCheckFlagsNonTaskDependency）；TASK/CONTRACTS/index 模板补注 |

## 7. Deduplication And History

Canonical BUG: BUG-CX-04（S3/S4 机器契约未传达 + locked 词汇碰撞族；duplicate of none，与 BUG-CX-03 同根因不同 stage）

| Date | Event | Actor | Runtime revision | Evidence |
|:--|:--|:--|:--|:--|
| 2026-08-17 | reported（复杂度审查 A1-A3/B1-B3/B6/B9） | 主会话+sub-agent | n/a | 本文件 |
| 2026-08-17 | fixed+verified（修复落地，全量测试/validate/doctor 绿） | 主会话 | n/a | S5.5 双 locked 消歧；guards 文案 phase 分支；§n 存在性机检（TestContractsCheckFlagsClauseNumberDrift）；非 TASK 依赖显式报错（TestTasksCheckFlagsNonTaskDependency）；TASK/CONTRACTS/index 模板补注 |
