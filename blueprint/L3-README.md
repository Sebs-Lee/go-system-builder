# L3 README — 第三层：各 Stage 详细设计（索引与共用词表）

> 层结构（以 L1 §六 为准）：第一层哲学蓝图 → 第二层生命周期实战目标 → **第三层 = 每个 stage 的详细落地设计**（含该环节的内容→工具映射与门禁逻辑）→ 第四层实现规格（REQ/架构/契约/任务）→ 第五层实现 → 第六层运营回灌。

## 文件索引

| 文件 | Stage | 一句话 |
|:--|:--|:--|
| L3-S0-requirement-design.md | S0 需求设计 | 把人的意图固化为可锁定的需求基线 |
| L3-S1-bind.md | S1 绑定 | 显式授权与权威状态起点 |
| L3-S2-design.md | S2 设计 | 架构决策 + 模块全量场景真相包 |
| L3-S3-contracts.md | S3 契约 | 分端执行契约与双向追溯 |
| L3-S4-task-split.md | S4 任务拆分 | 单职责任务 + 可判定收尾契约 |
| L3-S5-document-verification.md | S5 文档验证 | 独立双路审查 + 原子锁定 |
| L3-S6-build.md | S6 构建 | 两阶段授权下的实现与如实报告 |
| L3-S7-verification-round.md | S7 完整验证轮 | 三组正交 + 用例粒度 + 双重背书 |
| L3-S8-finding-investigation.md | S8 发现调查 | 深查七步与规范缺陷处置 |
| L3-S9-repair.md | S9 修复 | 边界约束修复 + 证据失效 + 定向重验 |
| L3-S10-acceptance-audit.md | S10 验收与审计 | 验收汇编 + 系统不变量审计 + 债务登记 |
| L3-S11-release-gate.md | S11 人工发布闸 | 不可默认的人闸与回滚路 |

## 共用机制清单（真实载体——L3 只允许引用本清单的真实机制名，禁用抽象代号）

抽象代号（如 T1/T4）会迫使读者查对照表=制造猜测成本（L1 公理五）。第三层直接写真实机制：

### A. Hook（`.claude/settings.json` 注册的生命周期事件——状态切换与派发执法的触发器）

| 事件 | 承载的职责 |
|:--|:--|
| `SessionStart` | 重投影当前状态：发恢复包（当前 stage/单一下一步/读序），从权威状态重入座（compact 恢复正门） |
| `PreToolUse` | **状态切换主力**：控制循环→质量门评估→满足则自动迁移（CAS）→安全决策（越界写拦截/锁定产物阻断） |
| `PreToolUse`（子代理派发匹配） | **派发前预检提醒**：单人 vs 团队？角色模板选对没？worktree 隔离？team_name 带了吗？ |
| `SubagentStart` | 派发瞬间提醒（预检答案落地、任务简报要求） |
| `SubagentStop` | 要求完成报告 + worktree→develop 集成检查清单 + completion_ack |
| `TeammateIdle` | 重唤醒**同一**队友（禁换人） |
| `PreCompact` | 持久化可恢复检查点（给下一个 SessionStart） |

### B. Harness（`loop-harness` 二进制——确定性引擎）

| 能力 | 关键命令/机制 |
|:--|:--|
| 状态机迁移（唯一写者，CAS） | `req bind` / `runtime transition --id TR-xxx` |
| 两阶段授权三事件 | `runtime agent-event`（readback_submitted → understanding_approved → activated） |
| 证据登记+指纹 | `runtime evidence add`（登记 id/kind/path/sha256/produced_by） |
| 读回信封生成 | `team launch --manifest --request-template`（指纹化读回请求） |
| 门禁/健康 | `ready`（门清单）/ `doctor` / `validate --all` |
| 缺陷事件 | `runtime bug-event`；干净轮求值 `verification clean-round` |

### C. 模板（`docs/**/*-template*`——字段即逼问，D4 主体）

REQ / TASK（读序、允许路径、收尾契约、职责）/ BE·FE·SYNC 契约 / BUG（七步）/ REV·QA·E2E·ACC 报告 / 场景四件套（scenario-model·cases·coverage·fixture-contract）/ 原型包头部。

### D. 角色定义（`agents/*.md`）

frontend/backend/test-builder（构建者）；document/delivery-verifier、qa、e2e-tester（验证者）——各带工具集/写路径/预载技能。

### E. 技能（`skills/*/SKILL.md`——方法论按需加载）

`two-phase-activation`（两阶段流程）/ `team-planning`（组队）/ `loop-orchestration`（驱动）/ `bug-resolution`（深查）/ `clean-round-evaluation` 等。

**编排三原则**：①模板负责"声明结构"（字段逼问），harness 负责"事实求值"（指纹/门/迁移），hook 负责"在自然事件上执法与提醒"；②同一职责多机制必须声明主备；③优先写机制的真实命令/事件名。

## 文档骨架（每份 L3-Sn 统一——五问讲清一件事）

1. **要实现什么**：本 stage 的转化目标（进入时是什么 → 出去时是什么，衡量标准一句话）。
2. **手头有什么**：可用的现有机制盘点（从下方机制清单挑相关的，各一句"它能干什么"）。
3. **选了什么、为什么**：选用表 + **否决表**（同样重要：哪个备选因复杂度>收益被删——我们是做减法，不是从零设计）。
4. **怎么编排**：按时间线把一件事讲完（从进入到出口，谁在什么时机做什么、信息怎么流动）。
5. **期望效果**：走完后系统处于什么状态、防住了什么、给下游交了什么。

**减法纪律**：每个机制的选择都要过"复杂度 vs 收益"；优先复用现有机制；说不出收益的机制不进编排。字段级规格不写在 L3（活在模板/实现里，属第四层），L3 只讲清"为什么这样安排"。

## 纪律

- 每份文档的每条设计标注承载的 D/公理（映射纪律）。
- 引用 L2 条目时只引用不复述；发现与 L2 冲突 → 停下修订 L2，不在 L3 私改。
- 修订走 L1 演化协议。

## 变更记录

| 日期 | 变更 |
|:--|:--|
| 2026-08-14 | 建立第三层：README（T1-T7 词表 + 骨架 + 索引）与 L3-S0…L3-S11 共 12 份 stage 详细设计 | owner 指示：第三层=每 stage 详细落地设计 |
| 2026-08-14 | 文件名补全 stage 英文名（L3-Sn-xxx），索引同步 | owner 指示 |
| 2026-08-14 | 骨架 §6 升级为「机制编排与职责覆盖」（含覆盖/重叠对账要求） | owner 指示 |
| 2026-08-14 | 删除 T1-T7 抽象代号，改为真实机制清单（hook 事件表/harness 能力/模板/角色/技能）——公理五：不制造猜测成本 | owner 复核 |
| 2026-08-14 | 骨架改为五问叙事（目标/手头工具/选择含否决/编排/期望效果）；确立减法纪律；字段级规格移至第四层 | owner 复核：讲清楚一件事 > 机制堆砌 |
