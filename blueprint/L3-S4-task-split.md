# L3-S4 — 任务拆分（Task Split）

> 层：第三层 ｜ 上游：L2 §S4 ｜ 版本 v3.1.0（v3.0.0 叙事版 + §6 注意力预算；机制事实经调查核实，含 file:line）

## 1. 要实现什么

把契约条款拆成**单职责、可独立验证的任务**——每个任务一份"完成判据先行"的收尾契约，让"完成"在动手之前就有可判定的定义。

- 进入时：S3 出口（locked 契约集 + oracle 链）。
- 出去时：`GATE-PLANNING-TASKS-COMPLETE` 可满足——存在 status=complete 的 TASK 文档（磁盘一致）+ `planning_task` 证据（Task Planner/Orchestrator，pass）。
- 衡量：每条契约条款至少被一个任务覆盖；每任务绑定**单一**契约、有收尾契约；写路径重叠有显式串行归属——S5 的 DV-TASK-EXECUTABILITY 审查就是照这份清单逐项过的。

## 2. 手头有什么（真实机制）

| 机制 | 它能干什么 | 载体 |
|:--|:--|:--|
| TASK 模板 | 头部九态 Status/单一契约绑定（Primary contract: {FE/BE/SYNC-id}）；§1 一句话可测目标；§2 Document Manifest（读序表带 SHA-256，修复任务把 BUG 前插 order 1）；§3 契约覆盖矩阵+模块场景覆盖（BR/CASE/Story/PATH/spec/evidence+4 条收尾 checkbox）；§4 Scope（read/**prospective write**/forbidden paths——forbidden 硬编码含 loop-state.json/allowed command classes）；§5 Selected Skills（带版本+指纹）；**§7 Closing Contract：4 行 assert 文本块**（契约条款满足/验证命令通过/changed_paths⊆授权写路径/范围偏差=空）；§8 Dependencies（依赖+所需证据+状态） | `docs/tasks/TASK-template.md` |
| 模板契约校验（机器） | TASK-template 必含字段检查（Team manifest/Assignment ID/Document Manifest/SHA-256/Selected Skills/Lifecycle Evidence/Closing Contract）+ 禁含旧机制词；doctor/validate 自动跑 | migration/templates.go:68-79 |
| 任务实体状态机 | `candidate→reviewed→locked→in_progress→review→done/cancelled`，事件带 guard（如 review→done 需 `required_verification_evidence_present`：收尾契约所需证据须在 runtime.evidence[] 里）——运行时经 runtime 命令推进 | loop-definition.json entity_lifecycles.task；assignment/task_lifecycle.go |
| 规划门推进 | TR-002（tasks→document_verification）挂 GATE-PLANNING-TASKS-COMPLETE，PreToolUse 自动迁移；guard 查"locked 契约+complete TASK 且磁盘状态一致"（无 runtime documents 时回退文件名模式） | loop-definition.json:987-1014；guards.go:305-368 |
| 证据登记 | `runtime evidence add --kind planning_task`（slot=planning_task_record） | catalog.go:348-350 |
| skill | `specification-planning`（S4 步骤：每 TASK 绑一契约+Closing Contract+单职责；scenario 类任务须声明模块回归全扫；TR-002 前核对真实条件）；`dag-design`（判断层：拒绝环、返回环路径——质量指导非门） | skills/ |
| 看板模板 | 任务矩阵（任务/契约来源/负责人/依赖/状态/阻塞）+ 关键路径文本声明 | `docs/tasks/index-template.md` |

注意**状态词汇两套并存**：模板头九态（activated/working/complete…）与 runtime 实体八态（in_progress/done…）不同名——机器只认实体态，磁盘 markdown 状态由 guard 交叉核对（要求 "complete"）。

## 3. 选了什么、为什么（含否决）

| 子问题 | 选用 | 否决与否决理由（减法） |
|:--|:--|:---|
| 怎么定义"完成" | TASK 模板 §7 的 4 行 assert 文本块（最早定义在 bugfix-review.md 六行模板）——判据先于实现（测试即设计） | 如实记录：**无 criterion_id 结构化字段**（那是我 v2 的目标态设计，未落地）；机器只查"所需证据在册"（guard），assert 内容的可执行性由 S5 审——模板逼问+判断层，暂不机检 |
| 怎么防巨型任务 | §1 一句话可测目标字段 + 单一契约绑定（Primary contract 唯一）+ S5 粒度审查 | 否决"行数硬上限机检"——粒度是语义判断，硬阈值误伤大 |
| 怎么管写路径 | §4 三类路径字段（read/write/forbidden）+ command classes 白名单 | 已是声明式；执行期由激活面+hook 接力（见 L3-S6），规划期不重复设门 |
| 怎么管依赖 | §8 Dependencies + dag-design skill（判断层拒绝环）+ 看板关键路径 | 如实记录：**无机器无环校验**（全库无 cycle 检测）——环检测目前靠 skill 方法+S5 审查；这是欠账 |
| 怎么锁任务 | **不在此锁**——S4 产出 complete 态；locked 发生在 S5 的 TR-003 原子锁（document_pass_lock 事件） | 否决"S4 自锁"——锁的权威在 S5（S5.5 是唯一合法锁定点），早锁会挡返工 |
| 覆盖校验 | §3 契约覆盖矩阵（模板）+ S5 DV 审查 | 如实记录：条款→任务覆盖**无机器校验**（semantic 只查任务实体状态合法）——靠模板+人审 |

## 4. 怎么编排（时间线讲完一件事）

1. **推导**：沿契约条款逐条推导任务；每任务先写 §1 一句话目标（写不清=拆得不对）与单一契约绑定。
2. **装填**：§2 读序（带指纹）→ §3 覆盖矩阵（条款+场景 CASE 双向挂）→ §4 三类路径与命令类 → §5 技能选择（带版本）。
3. **判据**：§7 收尾契约按 4 行 assert 写——其中"验证命令通过"一行直接引契约 oracle 链的验证方式；scenario 类任务在 §3.1 勾模块回归全扫。
4. **依赖**：§8 声明依赖与所需证据；按 dag-design 方法自查环（有环返回环路径重拆）。
5. **看板收口**：index 看板填任务矩阵与关键路径——给人的总览。
6. **登记推进**：`runtime evidence add`（planning_task，pass）→ PreToolUse 评估门（complete TASK 文档在+磁盘一致+证据在册）→ TR-002 自动迁移进 S5。
7. **锁定在下游**：S5 双职责 PASS 后 TR-003 的 `document_pass_lock` 才把任务置 locked——本阶段产物仍是可返工的。

## 5. 期望效果

走完 S4：

- **完成有判据**：每任务动手前就有 assert 级判据（含"修复前失败修复后通过"语义承接 bugfix-review 传统）；S6 兑现、S7 复验都对着这份判据；
- **结构性防住**：多职责巨任务（一句话目标+单契约绑定）、范围蔓延（三类路径+forbidden 含 loop-state）、"差不多完成"（assert 文本块+S5 审）、隐性写冲突（§8 依赖显式化）；
- **交给 S5**：TASK 批次——DV-TASK-EXECUTABILITY 的审查对象（覆盖/链接/范围/Closing Contract 可行性/依赖无环）与 TR-003 的锁定标的（task_batch_record 证据= DV 审查记录充当）。

**如实记录的已知缺口**（供第四层修复清单）：①DAG 无环**无机器校验**（dag-design 是判断层；semantic 只查跨 manifest 依赖引用存在，不查环）；②条款覆盖无机器校验；③Closing Contract 无结构化字段（criterion_id/command 是我 v2 的目标态，未进模板）——REQ-041/040 的判据工作可承接；④模板状态词与实体状态词两套并存，靠 guard 磁盘核对弥合；⑤`atomically_lock_execution_batch` 是证据记录占位（actions.go:348-350）。

## 6. 注意力预算与渐进披露

总评：语义判断（单职责/判据先行）的载体正确（模板字段即逼问），错配在**机械判断错留给判断层**——无环、覆盖计数、指纹三件算术活靠 agent 自觉。判定尺见 L3-README「注意力分配原则」。

### 6.1 当前错配（什么不对、为什么不对）

| # | 错配 | 为什么不对（L1 根据） |
|:--|:--|:--|
| 1 | DAG 无环靠 dag-design skill 自查（全库无 cycle 检测） | 公理二分工错位：拓扑排序是算术，却留给判断层；环=批次死锁，机械可判且收益极高（载体三问 1） |
| 2 | 条款→任务覆盖无机器校验 | §3 覆盖矩阵已是结构化表格，可解析可计数；"每条款至少一任务"是 L2 出口的可判定项，现在靠模板自觉+人审 |
| 3 | 双状态词汇（模板九态 vs 实体八态）靠 guard 磁盘核对弥合 | agent 必须同时知道两套词=记忆税+翻译错误面（C1）；词汇本身就该只有一套 |
| 4 | 读序 SHA-256 手填 | 填错要到 S5 才发现；`runtime fingerprint` 已能生成——手填是在制造可错点（D5：便宜确定性被弃用） |
| 5 | `atomically_lock_execution_batch` 是占位 action（actions.go:348-350） | 占位制造"有原子锁动作"的文档假象；原子性实际由 TR-003 迁移事务承载——如实即可，这个名字应删（公理五：机制名与机制实不符） |

### 6.2 阅读预算（谁在何时读什么）

| 角色 | 进入时必读 | 触发式加载 | 永不需要读（机制承载） |
|:--|:--|:--|:--|
| 拆分者 | TASK 模板 + 契约条款清单 | 拿不准拆分粒度/依赖设计时载 dag-design——它是"怎么拆"的方法论，整改后不再是"查环"的检查器 | 环检测/覆盖计数/指纹生成——validate 承载；状态词汇映射表——词汇统一后不存在 |
| 人（看板） | index-template 任务矩阵 + 关键路径声明 | — | — |

### 6.3 整改方向

- **左移**：validate 新增两查——依赖图无环（拓扑排序）+ 条款覆盖计数（解析 §3 矩阵）；manifest 改由 `runtime fingerprint` 生成输出；
- **删减**：任务状态词汇统一为一套；删占位 action（原子性如实由迁移事务承载）；
- **保持**：四行 assert 文本块（判据先于实现，测试即设计）；"不在此锁、锁在 S5"（锁的权威单点）——criterion_id 结构化是 REQ-040/041 的演进方向，不必抢跑。

## 变更记录

| 日期 | 版本 | 变更 |
|:--|:--|:--|
| 2026-08-14 | v1/v2 | 前两版（被判空洞→叙事不清；v2 的 criterion_id 结构与无环机检系虚构） |
| 2026-08-14 | v3.0.0 | 叙事版；机制事实经 sub-agent 调查核实（4 行 assert 模板/双状态词汇/锁在 S5 等如实入档） | owner 复核 |
| 2026-08-15 | v3.1.0 | 新增 §6 注意力预算与渐进披露（错配诊断/阅读预算/整改方向），判定尺引 L3-README | owner 指示：渐进披露、机制承载规范、削减平白叙述 |
