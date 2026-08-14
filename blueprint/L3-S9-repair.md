# L3-S9 — 修复（Repair）

> 层：第三层 ｜ 上游：L2 §S9 ｜ 版本 v3.0.0（叙事版；机制事实经调查核实——现状与设计态分开标注）

## 1. 要实现什么

按缺陷的边界修复：BUG 范围授权 → 先红后绿落地 → **失效受污染的历史通过证据** → 原发现职责定向重验 → 交接"仅差完整轮"。铁律：定向重验只证明这个 BUG 关了，唯一出路是全新完整轮。

- 进入时：S8 accepted 的 BUG 批次（带 repair/forbidden scope + Closing Contract）。
- 出去时：TR-012 从 `ready_for_full_review` 检查点进 verification——round+1、clean_round 清空、相位回 delivery，新完整轮开始。
- 衡量：**修复扰动过什么，就失效什么证据**——新轮引用不到任何被修复污染的旧 PASS。

## 2. 手头有什么（真实机制）

| 机制 | 它能干什么 | 载体 |
|:--|:--|:--|
| 修复任务与授权 | `repair_assigned` 事件（guard repair_task_and_builder_present，param 写入 repair_task_id/builder id）；TASK 模板规定修复任务把 canonical BUG **前插 order 1**；激活走与 S6 同一套两阶段（PTR-BUG-04：repair_understanding_approved+repair_activation_recorded→activation_record）——**"BUG 边界收窄"是协议/模板层约定**（写路径与命令类按修复所需声明），引擎无 BUG-scope 专属校验 | loop-definition.json:914-921,525-555；TASK-template.md:31 |
| 失效动作（核心） | `invalidate_affected_evidence` 挂在 PTR-BUG-05 的 action 链首位：`ComputeImpact(state, AffectedPaths)` 按 scope_refs 路径重叠筛 valid 证据 → 标 invalid（invalidated_by/规则/原因）；REQ 变更走另一条 `increment_baseline_and_invalidate`（整代际失效） | actions.go:363-386；impact/analysis.go:44-122 |
| 修复报告门 | PTR-BUG-05 evidence=repair_record(reported)+change_impact_record(recorded)；changeImpact schema：changed_artifacts 为 reference 数组、十类 change_types、四档 escalation、失效/替代/保留/须重验四清单 | loop-definition.json:574-577；registry.go:257-261；review-evidence.schema.json |
| 定向重验 | targetedReverification schema 15 字段：result 四枚举（pass/fail/blocked/**scope_changed**）+ if-then 约束（pass⇒scope_compliance=pass 且断言全 pass）；BUG 侧 retest_started→closing_contract_passed/failed 两事件（guard=证据 param 非空）；PTR-BUG-06/07 gate 要求 targeted_reverification_record（责任=**Original Finder**，当前轮） | review-evidence.schema.json:153-201；bug_lifecycle.go:243-268 |
| 独立性（机器，反向禁令） | `BUG_CLOSE_BY_FINDER_FORBIDDEN`：任何进 closed 的迁移强制 actor∉original_finder_agent_ids（身份级检查，catalog 没列也查）——注意 guard_spec 文档写的是正向"原发现者任 re-tester"，**实现语义相反**（禁令） | bug_lifecycle.go:157-180,243-256 |
| 交接检查点 | TR-012 `from_phase=ready_for_full_review` 引擎三处强制相位匹配；guards 里 `all_targeted_reverification_passed` 是真语义（所有 P0 不在未闭环态）；actions=start_review_round（round+1 且 **clean_round=null**）+set_verification_phase_delivery | loop-definition.json:1311+；engine.go:429-455；actions.go:223-231 |
| 重试上限 | `checkRetryLimits`（重入 investigating 时查 max_attempts/same_contract；超限报错"pause the Loop"但**无自动桥**）；typed `RepairLimitError`→`DispatchRepairLimitExceeded`→GTR-004 paused 仅在 adapter 接线路径生效 | bug_lifecycle.go:343-376；repair_limit.go；adapter/dispatch.go:38-57 |
| 模板级追踪 | TASK 模板的 BUG 追踪表（Finding→Canonical BUG→Impact→Repair→Reverify→Status） | TASK-template.md:114-116 |

## 3. 选了什么、为什么（含否决）

| 子问题 | 选用 | 否决与否决理由（减法） |
|:--|:--|:--|
| 失效-重验顺序怎么保 | 绑死在 PTR-BUG-05 的 action 链内（进 targeted_reverification 相位的**同一个迁移**先失效）——顺序=迁移原子性 | 如实记录：我 v2 宣称的"时间戳硬校验"不存在；且 **AffectedPaths 为空时失效是静默 no-op**——顺序保证退化为"动作被调过"，真正的防线是重验证据的当前轮约束+新轮四守卫 |
| 重验谁来跑 | gate 要求责任=Original Finder（当前轮记录）+ 关 BUG 的反向禁令双保险 | 否决"任意验证者复验"——原发现者最知道当初怎么发现的；但如实记录：正向"指派原发现者"无机器强制，靠 gate 责任字段 |
| 授权面 | 复用 S6 两阶段全套（同 activation_record slot），scope 由修复任务声明收窄 | 否决"为修复再造一套授权机制"——同一套已够，收窄在声明层（减法）；代价=引擎不校验 BUG-scope（诚实记录） |
| 定向≠全轮 | TR-012 的相位检查点+round+1+clean_round=null 三件套——新轮四守卫从零再算 | 否决"重验过就补章"——结构性无路 |
| 上限超了怎么办 | GTR-004 paused（带 pause checkpoint） | 如实记录：自动桥只在 adapter 路径；AdvanceBug 报的是字符串错误，需上层识别 |
| scope_changed | 重验结果第四枚举——重验中发现修复超范围，显式回炉而非硬判 | 给"边界声明错了"一条诚实的出路 |

## 4. 怎么编排（时间线讲完一件事）

1. **派修（S9.1）**：主会话建修复任务（BUG 前插 order 1，读序从缺陷报告开始）、指定 builder → repair_assigned → PTR-BUG-04 两阶段（读回批准→激活，声明面=BUG 边界）。
2. **修复（S9.2）**：先红（Closing Contract 判据红证据留存）→修→绿 → fix_reported（fix_ref 入实体）→ 登记repair_record(reported)。
3. **失效（S9.3，同一迁移内）**：PTR-BUG-05 提交——action 链先跑 `invalidate_affected_evidence`（按改动路径命中 scope 的历史证据标 invalid），再记修复完成；change_impact_record 落四清单。
4. **定向重验（S9.4）**：原发现职责对**该 BUG 的维度**复验（断言逐条+scope_compliance）→ pass：closing_contract_passed→closed（actor 不得是发现者，机器强制）；fail：closing_contract_failed→回 investigating（attempt+1，撞上限→paused）；scope_changed：回 S8 重定边界。
5. **交接（S9.5）**：PTR-BUG-06→ready_for_full_review（持久检查点，崩溃可恢复）→ TR-012：round+1、clean_round 清空、相位回 delivery——**新完整轮**，S9 期间的重验证据因轮次已+1 天然不能充当新轮证据。

## 5. 期望效果

走完 S9：

- **证据诚实**：被修复污染的 PASS 已失效（invalidated_by 可追溯到迁移）；change_impact 四清单给出失效/替代/保留/须重验的完整账；
- **结构性防住**：修一个坏两个（scope 声明+越界拦截沿用 S6）、"重验过=完成"（相位检查点+三件套重置）、自己关自己的 BUG（反向禁令）、无限重试（上限→paused）；
- **交给 S7（新轮）**：检查点+replacements 清单——新轮必重验集；同轮性从新轮号重新起算。

**如实记录的现状边界**：①AffectedPaths 空时失效静默跳过（调用方必须传全改动路径，否则失效不发生）；②"原发现者正向指派"无机器强制（gate 责任字段+反向禁令是两道可用的墙）；③guard_spec 与实现语义相反（original_finder_assigned 文档正向/实现禁令）；④same_contract_failure_count 死计数器确认（深查触发无机器支撑，REQ-041 前置）；⑤repair_task_id 仅查非空字符串不验实体存在。

## 变更记录

| 日期 | 版本 | 变更 |
|:--|:--|:--|
| 2026-08-14 | v1/v2 | 前两版（被判空洞→叙事不清；v2 的时序门 G3 系虚构） |
| 2026-08-14 | v3.0.0 | 叙事版；现状（action 链绑定失效/反向禁令/三件套重置）与设计态分栏 | owner 复核 |
