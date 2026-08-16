# L3-S5 — 文档验证（Document Verification）

> 层：第三层 ｜ 上游：L2 §S5 ｜ 版本 v3.2.0（v3.1.0 + S4 联动同步：覆盖/DAG 左移 tasks check、登记路径落地回写）

## 1. 要实现什么

进实现前的最后一闸：**两路独立审查**整条规格链（规格一致性 + 任务可执行性），双 PASS 后原子锁定执行批次——错误在纸上修复成本≈0，进代码后要烧一整轮验证。

- 进入时：S4 出口的 complete 任务批次 + locked 契约 + 真相包。
- 出去时：`GATE-DOCUMENT-PASS` 满足——**两条** document_review 证据（DV-SPEC-CONSISTENCY 与 DV-TASK-EXECUTABILITY，各 conclusion=pass、当前轮）+ 三层独立性检查过 → TR-003 自动提交，批次锁定。
- 衡量：**自己不能给自己签收**——审查者不是产出者、两条证据来自不同 agent、证据精确覆盖被审文档指纹；任一不满足，门不放行。

## 2. 手头有什么（真实机制）

| 机制 | 它能干什么 | 载体 |
|:--|:--|:--|
| document-verification skill | 10 步审查流程：自底向上读→REQ 覆盖双向→模块 current truth→场景一致性（100% 覆盖/oracle 独立/fixture 可行）→契约间一致性→可执行性→链接完整性→Closing Contract 可行性→REV 报告→迁移推荐；四条停止条件（含"审查者是作者"立即上浮人） | skills/document-verification/SKILL.md |
| document-verifier 角色定义 | 只读规格链+只写验证证据；禁 Edit/WebFetch；禁修复被审文档、禁审自己 authored 的维度、禁把缺失覆盖记 N/A；结论三枚举 | agents/document-verifier.md |
| REV 模板 | 头部（status/runtime ref/review round/workgroup manifest/assignment/responsibility/agent/activation）+ §1 指纹化输入表 + §2 职责结论与场景溯源（N/A 须记理由与证据）+ §3 findings（P0-P3/定位/预期/实测/证据）+ §5 证据有效性 + §6 三枚举结果块（DOCUMENT_PASS/FIX_REQUIRED/REQ_CHANGE_REQUIRED）+ 请求的生命周期事件 | docs/reports/review/REV-template.md |
| 独立性三层（机器） | ①事前：team-manifest 必填 `separation_edges`（reason=independence），两端 assignment 共享同一 agent 即拒；②事后：`applyDocumentPassIndependence` 三检查（审查者∉文档作者集合 / subject_refs 与当前 documents **精确全量匹配**（数量+三元组双向）/ 两条证据 producer 互不相同）；③旁路：`unauthorizedProducerConflicts`——任何 valid 证据的 producer_responsibility 越出 gate 允许映射 → gate 直接 Unknown（冲突） | team/validator.go:158-163；evaluator.go:515-561,301-355 |
| 证据信封 | envelope 全字段互证（id/kind/runtime_id/generation/responsibility/produced_by）+ 索引指纹 + 当前轮一致性；document_review 的 subject 必须精确匹配被审文档集 | evaluator.go:162-176,443-513 |
| 双职责门 | GATE-DOCUMENT-PASS = 两条 currentRoundRequirement（各职责一条 pass）；**S5 阶段 review.round=0 即"当前轮"**（轮次递增只在 TR-006 起的 start_review_round） | registry.go:162-165；run.go:663-666；actions.go:223-231 |
| TR-003/TR-004/TR-005 三叉 | 共用 selector：双 PASS→TR-003（锁定进 building）；任一职责 fix_required+requested_event=document_fix_required→TR-004（回 planning 重修）；REQ 级→TR-005（paused，human_boundary） | loop-definition.json:1015-1102 |
| 锁定生效 | TR-003 提交后 documents 中 locked 条目经 hookctx 投影为 LockedArtifacts → PreToolUse 硬阻断写入（恢复指向 versions/g{N+1}/） | hookctx/loader.go:250-299；policy/engine.go:269-274 |
| 两阶段激活 | 审查者自身也要 readback（ready/conflict/missing_input/out_of_scope 四枚举）→激活信封→才能写 REV | two-phase-activation；SKILL.md:60 |

## 3. 选了什么、为什么（含否决）

| 子问题 | 选用 | 否决与否决理由（减法） |
|:--|:--|:---|
| 怎么防自审 | separation edges（事前派发排除）+ producer 三检查（事后机器）+ skill 停止条件（判断层）——三层主备 | 否决"单一审查者通吃两职责"——两职责正交且需互异 producer（门硬性要求两条不同责任证据） |
| 怎么防"纸面对齐" | subject_refs 精确全量匹配——证据必须精确覆盖**当前指纹**的文档集（多一少一都拒） | 否决"抽样引用 subject"——S5 的签收对象是整批，抽样签收=给漂移留门 |
| 返工怎么收敛 | TR-004 回 planning → 修复 → **受影响职责以新指纹重跑**（旧证据 subject 自动失效，不需要显式作废机制） | 否决"每次全量重审"——指纹失配天然圈定重跑范围，省时滞 |
| 并行还是串行 | S5.2/S5.3 并行（协议子阶段不变量），任一有发现即进 S5.4 | 并行买时滞，代价仅编排复杂度一点——值得 |
| 锁的原子性 | TR-003 单事务（CAS 迁移）承载"原子"；无中间可见态 | v3.1 记账的占位 action 与登记欠账**均已解决**（S3 落地 register_locked_contracts/register_execution_batch，S4 落地 register_planning_tasks；占位 action 已删）；原子性由迁移事务保证 |
| N/A 怎么管 | 模板要求 N/A 须记理由与证据（判断层） | 否决"机器枚举合法 N/A"——适用性是语义判断，模板逼问即可 |

## 4. 怎么编排（时间线讲完一件事）

1. **组队（S5.1）**：主会话派两职责任命——team-manifest 声明 separation_edges（independence），validator 拒共享 agent；每个审查者走两阶段激活（readback 四枚举→激活信封）。
2. **并行审查（S5.2/S5.3）**：各审查者按 skill 自底向上读（TASK→契约→REQ→设计→rules），逐层核对：一致性职责盯"验收↔条款↔场景映射+跨文档引用指纹"；可执行性职责消费 `tasks check` 机检结论（覆盖双向/DAG 无环已在 TR-002 把门），再审机器算不了的三件——Closing Contract 可行性/粒度/写路径归属串行。
3. **发现回路（S5.4）**：任一职责出 finding（REV §3，P0-P3 带定位）→ 结论 FIX_REQUIRED + requested_event=document_fix_required → TR-004 回 planning → 修复受影响层 → **仅该职责以新指纹重跑**（旧证据因 subject 失配自动作废）。
4. **双 PASS 收口（S5.5）**：两职责各登记一条 document_review 证据（pass，subject=当前文档集精确指纹）→ 下一次 PreToolUse：gate 求值（两条证据合格 + 三层独立性过）→ TR-003 提交——批次锁定，hook 拦截自此生效。
5. **REQ 级歧义**：结论 REQ_CHANGE_REQUIRED → TR-005 → paused，human_boundary——交人。

## 5. 期望效果

走完 S5：

- **规格链自洽且冻结**：双职责独立签收过同一批精确指纹；此后任何改动都会被锁定拦截拦下（只能走 versions/g{N+1} 新代次）；
- **结构性防住**：自审（三层独立性）、纸面对齐（精确 subject）、半锁状态（单事务）、返工蔓延（指纹失配圈定重跑范围）、礼貌通过（findings 须带定位+证据、N/A 须记理由）；
- **知识共享复利**：审查者读全链建立的心智模型可在 S7 转任对应验证职责；
- **交给 S6**：锁定批次（锁定清单+指纹）——构建读序与授权面的基准。

**如实记录的已知缺口**（供第四层修复清单）：~~①author_agent_id 无生产写入路径~~（S3/S4 登记 action 落地，作者随登记写入）；~~②契约/任务 documents 登记无代码先例~~（PTR-PLAN-02/TR-002/TR-003 三点登记）；③轮次语义：S5 的证据 review_round=0，若 S7 期间产生新 document_review 需注意轮次校验；④REV 模板三枚举与 gate 的 conclusion 词汇（pass/fix_required）不同名，靠登记时映射。

## 6. 注意力预算与渐进披露

总评：独立性靠机制的方向正确（separation_edges 事前排除 + producer 双查事后机器核），第三层（author 集合）是典型的**"文档补机制的洞"**——禁令写了三处，机器检查却无数据。判定尺见 L3-README「注意力分配原则」。

### 6.1 当前错配（什么不对、为什么不对）

| # | 错配 | 为什么不对（L1 根据） |
|:--|:--|:--|
| 1 | ~~`documents[].author_agent_id` 无生产写入路径~~（**已解决**：S3/S4 登记 action 落地；TR-002 起任务批也在册） | 原 D2 论证成立——补的正是那条登记写入；S5 设计时独立性检查从"叙述三处"收敛到机器数据 |
| 2 | document-verification skill 十步未区分必做与按发现触发 | 审查者进入即背十步——前 3 步（读序/指纹/覆盖双向）是必做，后 7 步应由发现触发；渐进披露缺失导致均匀付费 |
| 3 | REV 三枚举（DOCUMENT_PASS/FIX_REQUIRED/REQ_CHANGE_REQUIRED）与 gate conclusion（pass/fix_required）不同名，登记时映射 | 词汇翻译是摩擦+错读面——同一概念两套名字违反"一概念一名字"（naming 规则自己的纪律） |

### 6.2 阅读预算（谁在何时读什么）

| 角色 | 进入时必读 | 触发式加载 | 永不需要读（机制承载） |
|:--|:--|:--|:--|
| 审查者 | 被审文档集（TASK→契约→REQ→设计，自底向上）+ REV 模板 | **发现指向哪层才深读哪层**（步骤 4-9 按发现触发——整改后 skill 应显式标注 [必做]/[触发]） | "不得审自己 authored 的"——separation_edges 事前就派不进来；独立性核验细节——applyDocumentPassIndependence 机器查（evaluator.go:515-561） |
| 主会话 | 双职责结论 + findings 汇总 | — | subject 精确匹配的规则——登记时机器核 |

### 6.3 整改方向

- **补洞**：契约/任务登记时写入 author_agent_id（一条命令级改动，让第三层独立性从空转变真检）；
- **左移（渐进披露结构化）**：skill 步骤标注 [必做]/[触发]，审查者的注意力花在发现上而不是流程背诵上；
- **统一**：REV 三枚举与 gate conclusion 词汇合一，消掉登记时映射；
- **保持**：subject_refs 精确全量匹配（防纸面对齐）、TR-004 指纹失配自动圈定重跑范围（省时滞的机制最优解）。

## 变更记录

| 日期 | 版本 | 变更 |
|:--|:--|:--|
| 2026-08-14 | v1/v2 | 前两版（被判空洞→叙事不清） |
| 2026-08-14 | v3.0.0 | 叙事版；机制事实经 sub-agent 调查核实（三层独立性/精确 subject/轮次=0/author 空转等如实入档） | owner 复核 |
| 2026-08-16 | v3.2.0 | S4 联动同步：覆盖/DAG 审查改为消费 tasks check 机检结论；登记路径/author_agent_id 欠账划线；占位 action 已删 | S4 整改（L3-S4 v4.0.1） |
| 2026-08-15 | v3.1.0 | 新增 §6 注意力预算与渐进披露（错配诊断/阅读预算/整改方向），判定尺引 L3-README | owner 指示：渐进披露、机制承载规范、削减平白叙述 |
