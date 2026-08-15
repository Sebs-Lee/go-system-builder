# L3-S10 — 验收与发布审计（Acceptance & Audit）

> 层：第三层 ｜ 上游：L2 §S10 ｜ 版本 v3.1.0（v3.0.0 叙事版 + §6 注意力预算；机制事实经调查核实）

## 1. 要实现什么

验收回答"**承诺的都兑现了吗**"（逐条证据），审计回答"**系统还站得住吗**"（跨变更不变量）——干净轮证明的是"这次改动对"，审计证明的是整体没被破坏。

- 进入时：干净轮记录（review.clean_round=当前轮）+ 全部有效证据。
- 出去时：TR-017 到 `awaiting_human_release`——ACC(pass)+审计(approved/approved_with_risk)+干净轮仍有效三件齐。
- 衡量：**每条验收标准都能映射到同轮有效证据**；审计八区全过、无阻断发现。

## 2. 手头有什么（真实机制）

| 机制 | 它能干什么 | 载体 |
|:--|:--|:--|
| 四条迁移四道门 | TR-015（acc_complete+clean_round_still_valid→release_audit）/ TR-016（验收差异→回 verification 开新轮，带失效 action）/ TR-017（审计过→人闸）/ TR-018（审计阻断→paused）；对应 gate：acceptance_record(pass)+**当前轮**clean_round_record；release_audit_record(approved|approved_with_risk)+acceptance_record(pass) | loop-definition.json:1420-1545；registry.go:217-232 |
| clean_round_still_valid（真 guard） | 委托 `EvaluateCleanRound` 纯函数复算——验收/审计推进时干净轮必须**仍然**成立（S7 之后若有任何漂移，这里现形） | guards.go:257-261 |
| ACC 模板 | 六章：指纹化基线（REQ/真相包/契约/任务四行 SHA-256）→ 干净轮表（三组 manifest+轮次+PASS+有效性）→ **需求验收逐条映射**（source_ref/Rule/CASE/Story/PATH→期望-证据-结果）→ 交付与运维（部署顺序/迁移数据处理/回滚/交接）→ 残余非阻断风险（Tracking artifact 可挂 TD）→ 结论（passed/blocked）；尾注"Acceptance does not authorize release" | docs/reports/acceptance/ACC-template.md |
| 发布审计模板+规则 | 16 章模板：状态机/事务 UoW/并发幂等/数据模型与迁移/调用点/可观测性/验证证据/文档与发布边界（含"技术债已明确标注"检查项）/阻断发现（ARA-编号）/非阻断风险/10 问签核/三枚举终判/人的边界；R-P05 规则：**8 个审计区**、**恰好 8 项阻断发现**（状态无出口/迁移缺失/raw SQL 未验/唯一键无历史检查/并发创建只靠 SELECT/关键 DB 行为仅 mock/范围外代码混入/无效 session 记录）、BLOCKED=不合并不发布 | docs/release_audits/TEMPLATE.md；docs/rules/release-architecture-audit.md |
| acceptance-and-handoff skill | 8 步：复验干净轮（stale→要求新完整轮）→ 组装 ACC → 证据同代际同轮核对 → TR-015 → 执行审计 → 发现分类（可纠正→回完整轮；阻断→记录暂停；完整性→交人）→ 组装人闸交接包（clean-round ID/hash、ACC、审计证据、代际、"automation stops"显式声明）→ TR-017；五条停止条件 | skills/acceptance-and-handoff/SKILL.md |
| 失败路由 | 审计缺陷→S8（缺陷流）；报告不完整（非缺陷）→S6；REQ 缺口→req_amendment | agent-protocol.md:444 |

## 3. 选了什么、为什么（含否决）

| 子问题 | 选用 | 否决与否决理由（减法） |
|:--|:--|:--|
| 验收怎么防"总体感觉" | ACC 模板逐条映射表（每条验收→证据 ID）+ gate 要求 acceptance_record(pass) | 否决"机检全部映射"——证据语义关联是判断；模板逼逐条、gate 逼签收 |
| 干净轮怎么防"过期" | `clean_round_still_valid` 真复算（不是比对缓存字段） | 否决"记一个 clean_round 标志就信"——四守卫现算，漂移当场暴露 |
| 审计怎么防走过场 | 固定 8 区+8 阻断项清单+三枚举+独立结论记录 | 否决"审计即又一次 code review"——R-P05 明令"not another code review"，只查系统不变量 |
| approved_with_risk | 审计 gate 接受两档通过——带非阻断风险也到人闸，风险显式留档 | 否决"有风险一律 BLOCKED"——风险分寸是人的判断，交 S11 裁 |
| 债务登记 | ACC 第 5 章风险表 Tracking=TD 挂账 | 如实记录：**无独立债务章节**（我 L2 写的"债务登记表"是设计态）——现状债务=非阻断风险的一种追踪形态 |

## 4. 怎么编排（时间线讲完一件事）

1. **复验前提**：skill 第一步重算干净轮（stale/被取代→停，要求新完整轮）。
2. **组装 ACC**：指纹基线四行→干净轮表→逐条验收映射（每条挂证据 ID）→交付运维（迁移/回滚）→残余风险挂 TD。
3. **推进**：登记 acceptance_record(pass) → PreToolUse 评估 gate（+当前轮 clean_round 复算）→ TR-015 进审计。
4. **执行审计**：按 16 章模板走 8 区；发现分类——可纠正→TR-016 回 verification 开新轮（带证据失效）；阻断→记录→TR-018 paused；无阻断→登记 release_audit_record(approved|approved_with_risk)。
5. **到闸**：TR-017——三 guard（audit_approved+acc_complete+clean_round_still_valid）过 → `awaiting_human_release`，交接包五要素+automation stops 声明，交 S11。

## 5. 期望效果

走完 S10：

- **验收逐条可溯、干净轮到闸仍成立**（推进点复算是最后一道防漂移）；
- **结构性防住**：总体感觉式验收（逐条映射）、审计变 code review（固定清单+系统视角）、风险溜过（approved_with_risk 显式留档）、"验收即授权发布"（模板尾注+TR-025 只在 S11）；
- **交给 S11**：交接包（clean-round ID/hash、ACC、审计证据、代际、人须做的事项）。

**如实记录**：①`acc_complete`/`release_audit_approved` guard 本体是证据非空桩，真语义在 gate 的 verdict 要求层；②债务无独立登记结构（挂 TD 追踪）；③`releasegraph` 校验的是**模板发布树**（tarball 引用完整性/无实例工件泄漏），与业务系统的发布审计同名不同物——勿混。

## 6. 注意力预算与渐进披露

总评：分配健康——逐条映射（模板逼）、干净轮真复算（机器判）、8 区固定清单（可核对而非散文）。主要问题是审计模板的 **N/A 扫填面**与一处命名撞车。判定尺见 L3-README「注意力分配原则」。

### 6.1 当前错配（什么不对、为什么不对）

| # | 错配 | 为什么不对（L1 根据） |
|:--|:--|:--|
| 1 | 审计 16 章模板对多数 REQ 存在大量 N/A 章节 | N/A 扫填侵蚀逼问——8 区是主干、其余是条件章节；"必填但无关"训练 agent 跳读（同 S3 错配 3） |
| 2 | releasegraph 与业务发布审计同名异义（校验的是模板发布树的引用完整性） | 命名撞车消耗每一次解释成本（本文即需专门辨析一条——这就是证据）；公理五：读者从名字重建的理由是错的 |
| 3 | （如实记录的非错配）债务登记无独立结构、挂 TD 追踪 | 这是**正确**的克制——补结构=新仪式（公理三），维持现状即可；L2"债务登记表"的表述应向现状对齐而非反向加机制 |

### 6.2 阅读预算（谁在何时读什么）

| 角色 | 进入时必读 | 触发式加载 | 永不需要读（机制承载） |
|:--|:--|:--|:--|
| 审计者 | 8 区清单 + 恰好 8 项阻断发现（固定、可逐项勾对的清单——不是散文规则） | 触及迁移/并发/数据模型等区才读该区章节细节 | 干净轮有效性——clean_round_still_valid 机器复算（guards.go:257-261）；"审计不是 code review"的告诫——固定清单本身就是这个语义的机制化 |
| 人（S11 消费者） | ACC 结论 + 残余风险表 | — | 逐条证据——汇总已由 ACC 模板逼出 |

### 6.3 整改方向

- **重组**：审计模板按"8 区主干 + 条件章节"分层（不适用=不出现，而非填 N/A）；
- **改名**：releasegraph 消歧（如 template-release-check）；
- **保持**：逐条映射、真复算、approved_with_risk 两档、BLOCKED=不合并不发布（机制已最优）。

## 变更记录

| 日期 | 版本 | 变更 |
|:--|:--|:--|
| 2026-08-14 | v1/v2 | 前两版（被判空洞→叙事不清） |
| 2026-08-14 | v3.0.0 | 叙事版；机制事实经调查核实（四门四迁移/clean_round 真复算/8区8阻断/releasegraph 同名辨析） | owner 复核 |
| 2026-08-15 | v3.1.0 | 新增 §6 注意力预算与渐进披露（错配诊断/阅读预算/整改方向），判定尺引 L3-README | owner 指示：渐进披露、机制承载规范、削减平白叙述 |
