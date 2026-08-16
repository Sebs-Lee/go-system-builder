# L3-S2 — 设计（Design）

> 层：第三层 ｜ 上游：L2 §S2 + L2 跨阶段全局规则「单一验证分母」 ｜ 版本 v4.0.1（v4.0.1 设计对抗审查处置：并行语义/仲裁/N/A 背书/桥拆两段/交叉格载体/深度自审；v4.0.0 联合审查版：断点地图+双轨汇聚+AC 桥；机制事实经调查核实，含 file:line）

## 1. 要实现什么

产出架构决策（ADR）+（UI 影响时）**模块全量场景真相包**——它是后续契约、任务、验证用例的唯一事实源。

- 进入时：绑定生效的 REQ（含 ui_impact 三值）。
- 出去时：`GATE-PLANNING-DESIGN-COMPLETE` 可满足——一条 `planning_design` 证据（Architect/Orchestrator，pass）+ 架构文档 + 真相包过校验。
- 衡量：**全量与双极性**——场景以模块为全集（不是本需求的子集副本），每条规则的正反分支都有 oracle；`scenario validate` 说绿才算绿。

**v4.0.0 身份扩展（联合审查确立）**：S2 不只是"设计 stage"，它是**全链验证源的出生地**——本 stage 产出的四件套质量（oracle 七字段完备、PATH 绑定真实、browser_required 正确、AC↔CASE 可达）直接决定 S7 完整性门的分母成色与 REQ-040 的挂载基础。因此 S2 的整改以"下游对位"为准绳，不以其自身机制为限。

## 2. 手头有什么（真实机制）

| 机制 | 它能干什么 | 载体 |
|:--|:--|:--|
| ARCHITECTURE 模板 | 12 节结构：目标/上下文/容器/模块职责（含**排除范围**列）/数据流/状态机/数据模型/接口/安全/性能基线/部署回滚/锁定 | `docs/design/architecture/ARCHITECTURE-template.md` |
| ADR 目录 | 决策记录写 `docs/design/decisions/`（skill 流程要求，模板本身无 ADR 段） | specification-planning SKILL.md:28 |
| 场景四件套 | **手写两件**：`scenario-model.json`（facts 分区 + rules→branches，branch 带 `polarity`(positive/negative)、`oracle`、`fixture_id`、`story_refs`/`flow_refs`、`browser_required`）+ `fixture-contract.json`（合成数据装配 setup/cleanup）；**生成两件**：`cases.json` + `scenario-coverage.json`（引擎从手写件派生，字节级防篡改） | `docs/design/prototypes/<module>/`；schema 在 internal/schema/assets/ |
| oracle 结构 | 正向必填 visible[]/terminal_state/persisted_effects/forbidden_side_effects；负向加 rejection（稳定拒绝码/消息）/expected_state/recovery（N/A 须引 source_refs 证明） | scenario-model.schema.json:18-53 |
| `loop-harness scenario` | `generate`（校验+原子生成 cases/coverage）/`validate`（源头校验+生成物字节比对+可选 `--require-specs` 查 Playwright 覆盖 100%） | internal/scenario/engine.go；run.go:169 |
| 覆盖率与正反比 | required branch 100%（构造保证）；负:正 ≥ coverage_profile 下限（critical=3/rule-dense=2/ordinary=1）；doctor/validate 自动跑 scenario 校验（不带 --require-specs） | engine.go:354-363,1220-1238；semantic/validator.go:110 |
| 原型包规则 | index.html+stories.md(S-NNN)+flows.md(F-NNN/PATH-*)+页面 HTML+四件套；HTML 头部字段；sticky 侧栏 6 节；模块任一改动→E2E 全模块回归 | `docs/rules/ui-prototype.md` |
| UI 原型门（hook 侧） | ~~populateUIPrototypeFact 检查真相包并拦截~~——**死残迹**（hook-policy 零匹配该 fact，测试钉死永不触发；调用点 run.go:1855，函数体 run.go:2330+；§6.1 #3 判定，本轮删除） | run_test.go:1137+ |
| skill 群 | `specification-planning`（主：9 步流程，**未显式串联下面四个**——SKILL 与协议对子 skill 零引用，§6.1 #4 的诊断对象）；`scenario-model-design`（四件套方法，**其 Workflow 仍是旧顺序**——step 7 才 reconcile stories，v4.0.0 双轨顺序要求同步修正）；`user-story-design`（S-NNN result-led）；`user-flow-design`（F-NNN/PATH checklist）；`ui-prototyping`（HTML 容器/侧栏契约） | skills/ |
| 规划门推进 | PTR-PLAN-01（design→contracts）挂在 GATE-PLANNING-DESIGN-COMPLETE 上，PreToolUse 自动评估自动迁移 | loop-definition.json:137-157；registry.go:153-155 |
| angles 注册表 | 模块级 append-mostly 自检清单（ANG-{MODULE}-{NNN}，黑名单禁通用词），S5/S7 的 angle_complete guard 消费——S2 设计变更时的模块视角交接输入 | runtime/angles.go；`loop-harness angles` |

### 2.1 联合审查：CASE 验证源的分母断裂地图（v4.0.0，2026-08-15 sub-agent 端到端调查）

流水线在"出生→冻结"（四件套生成+字节防篡改）与"S7 gate 求值"（clean round 纯函数）两段各自机械连续，**但不共享 CASE 粒度分母**——S7 数的是 responsibility_ids。中间五个断点：

| # | 断点 | 形态 | 处置归属 |
|:--|:--|:--|:--|
| 3 | spec 绑定：`--require-specs` 校验扎实（fail-closed TS 解析、case_id+全部 PATH token 须在同一 Playwright 回调体）但自愿触发 | 机械·非强制 | **本轮左移**（进 doctor） |
| 4 | 契约引用：CONTRACTS/BE/FE/SYNC/REQ §F/TASK §3.1 的 BR→CASE→S→F→PATH→Spec 链手抄无校验 | 手抄 | S3 轮（token 对账 cases.json，已记入 L3-S3 整改方向） |
| 6 | 执行分派：e2e-tester 按协议读 cases.json 为"唯一清单"，runtime 无结构化 CASE 执行集 | 纯文本承诺 | REQ-040 FR-002 |
| 7 | 验证计数：三 gate 只查一条 e2e_review pass 字符串；clean round 按职责 | 机械·分母错位 | REQ-040 FR-004（CASE_ID 粒度） |
| 9 | S8 反查分母是另一套宇宙：e2e-coverage 读手维护 inventory（CT/AC/HOOK/TASK-CLOSING），与 CASE 宇宙零映射 | 纯 skill+分母割裂 | 记入 L3-S8 整改方向，作 REQ-040 输入 |

**三个修正既有认知的事实**：①`scenario-coverage.json` 的 required_branch_coverage 是**构造性 100%**（引擎对每个 required case 同时 Required++/Covered++，engine.go:1220-1238）——设计时声明覆盖，非执行覆盖，**不得作执行覆盖证据**（执行覆盖由 S7 证据计数承担）；②angle_declaration 是评审组 pre-review 写的声明证据（REQ-003 FR-001），S2 的 angles 注册表只是其交接输入——§2 表中"消费"表述据此理解；③四件套只对 ui_impact=changed 强制，非 UI REQ 无 CASE 宇宙——设计边界成立但需显式声明（REQ-040 Q-002 预留"无场景包⇒BLOCKED 而非跳过"）。

**对位缺口（本轮核心新增）**：REQ 的验收标准（AC）与场景 CASE 之间**无桥**——链可拼（AC→FR 在 REQ §C3"指向"列；FR→BR→CASE 在 §F 覆盖矩阵与 rule.source_refs）但分段活在三份文档、三个 stage，无一处可见全链、无机器对账。**验证分母的完整形态不是 case.id 单点，而是 AC↔CASE 双向可达**（每条 AC 至少一 CASE 或显式 N/A）。

## 3. 选了什么、为什么（含否决）

| 子问题 | 选用 | 否决与否决理由（减法） |
|:--|:--|:--|
| 怎么逼出全量双极性场景 | scenario-model 的 branch 字段（polarity/oracle 必填）+ 引擎校验（正反比、witness 引用注册 fact、S/F-NNN 必须真实存在） | 否决"自由格式设计文档"——无结构则无全量可言；字段即逼问 |
| 怎么防"需求私有副本" | 模块目录为唯一家（REQ 只是 source_refs）；REQ-template §F 明令"REQ 不能创建私有副本" | 否决"每 REQ 一份场景"——副本必然漂移，模块才是全集的分母 |
| 怎么防生成物篡改 | cases/coverage 由引擎生成 + validate 字节级比对（"stale or tampered"） | 否决"手写 cases"——生成物不可手改，源头单点 |
| 怎么管 UI 影响 | 三值 + `ui_impact_resolved` guard + hook 侧原型门（写契约瞬间查真相包） | 否决"仅文档提醒"——双门（规划 guard+hook 拦截）成本极低 |
| 架构决策怎么落 | ARCHITECTURE 模板 12 节 + ADR 另录 decisions/ | 否决"再造 ADR 模板段"——已有 ADR-template.md 与目录惯例，不重复 |
| spec 覆盖 100% | `--require-specs` 显式开关（doctor 不默认跑） | 如实记录：这是**欠账**——L2 要求拒绝路径非空已由 oracle 保证，但 Playwright spec 级覆盖未进常规校验，属第四层待补 |
| 模块视角积累 | angles 注册表（append-mostly） | 否决"每次全靠 REQ 内描述"——模块教训要跨 REQ 存活 |
| 生产顺序怎么定（v4.0.0 新增，v4.0.1 审查补强） | **双轨汇聚顺序**：系统轨（架构→facts）∥用户轨（§A→stories）→ 汇聚① rules/branches+oracle（facts×FR×stories 三方交叉找全量，**交叉格清单为载体**）→ fixtures（行为定型后造）→ 汇聚② flows/原型（stories×branches，PATH 绑定）→ 收口 generate/validate+AC 桥。**stories 提前**、**fixtures 后移**（理由同 v4.0.0）。**并行语义（v4.0.1）**：同一 agent 的顺序自由（两轨可交错、无强制先后），不是子代理派发；**冲突仲裁（v4.0.1）**：架构约束旅程——stories 挑战架构时升 ADR 人闸裁，不静默让步；**用户轨前置（v4.0.1）**：既有模块演进时先读完整模块包再写新 stories（防与既有 S-NNN 重复/矛盾） | 承载 D4（顺序即课程）+C4（人只拍方向）。否决"逐层拍板漏斗"（S0 形状硬搬）——oracle 是 branch 字段非独立生产步骤；否决"六产物并行铺开"——改一处全链重跑 |
| AC↔CASE 桥怎么承载（v4.0.0 新增，v4.0.1 拆两段） | **两段检查**：AC→FR→BR 段挂汇聚①后即可机检（source_refs 溯源，不需生成物）——源头早查；全链 AC↔CASE 段在收口 validate 复核（需 cases 生成物）。**N/A 须背书（v4.0.1）**：N/A 不是自由文本逃生门——须带类别+指针（非功能 AC→NFR 编号；范围外→REQ §A4 负空间条目），且 N/A 清单进 ADR 人闸拍板包（L2 铁律 1"经独立背书的不适用"的同构） | 承载 D6（覆盖按用例计数、沉默不是不适用——L1:103 直引）。否决"只靠 §F 手抄矩阵"；否决"S7 侧事后核"；否决"N/A 一句理由过门"——无背书的 N/A 是从分母里静默移除验证项 |

## 4. 怎么编排（时间线讲完一件事）

1. **架构决策**：按 ARCHITECTURE 模板 12 节起草；每个实质决策另录 ADR（决策/风险/后果/备选）入 `decisions/`。
2. **UI 影响分岔**：none → 直奔契约；changed → 进入真相包流程；unknown → 停：guard 不放行规划推进（见 §5 已知缺口）。
3. **双轨并进**（changed 时，v4.0.0 顺序修正）：
   - **系统轨**：数据模型/状态机定型 → facts 分区（架构词汇表的落盘）；
   - **用户轨**：从 REQ §A 写 stories（S-NNN 引 REQ-id）——行为缺口的上游种子，与架构并行；
   - **汇聚①行为全量**：rules/branches（正反成对，oracle 随 branch 写）= facts × FR × stories 三方交叉找全（含拒绝路径）；分支的 story_refs 指向已存在的 stories。**交叉格清单（v4.0.1 载体）**：三方交叉的产物是模块包内一份手写清单（fact×FR×story 格 → branch id 或"无分支理由"）——"找全"从叙述变成填空，validate 轻校验其 branch 引用存在；**汇聚①后即跑 AC→FR→BR 源头检查**（早查早回，不等收口）；
   - **fixtures**：branches 定型后造数据（setup/cleanup），一次造对；
   - **汇聚②可走查**：flows（F-NNN/PATH-*）+ 原型页（头部+侧栏契约）= stories 的旅程 × branches 的行为；browser_required 的 branch 在此绑定 PATH。
   **在既有模块包上演进**，不建新副本。
4. **深度自审（收口前，v4.0.1 新增）**：换身份攻击自己的产物——实现者："哪条 oracle 我落不了地或区分不了对错实现？"；e2e-tester："哪条负向 CASE 我按七维度取不了证？"；维护者："哪条规则会和模块演进打架？"——结论一段话入 ADR 包（REQ-040 matcher 落地前的语义质量逼深手段，S0 身份互换自审的同构）。
5. **收口**：`scenario generate` 产 cases/coverage → `scenario validate`（或 doctor）绿：正反比、引用存在性、字节冻结、交叉格清单轻校验、**AC↔CASE 全桥**（每条 AC 至少一 CASE 或背书的 N/A）；不过 → 回汇聚①补负向分支或回 REQ 澄清 AC。
6. **提交规划证据**：登记 `planning_design` 证据（pass）→ 下一次 PreToolUse，controller 评估 GATE-PLANNING-DESIGN-COMPLETE → PTR-PLAN-01 自动迁移进 contracts——无人工拨表。
7. **UI 门的第二道保险**：~~hook 拦截~~（populateUIPrototypeFact 为无消费者残迹，测试钉死永不匹配——v4.0.0 裁定删除）；真实保险是 Quality Gate not_ready（真相包不齐 → missing 指回补包）。

## 5. 期望效果

走完 S2：

- **全量可机检的场景资产**：模块真相包过 `scenario validate`，正反比达标，生成物防篡改——S7 的用例分母与 fixture 装配直接来源于此；
- **结构性防住**：需求私有副本（模块为唯一家）、只写正向（polarity+正反比）、oracle 散文化（schema 必填结构）、UI 影响含糊进契约（guard+hook 双门）、模块教训流失（angles）；
- **交给 S3**：架构决策 + 真相包（锁定指纹）——契约的引用输入。

**缺口处置台账**（v4.0.0 重写为处置视角）：①oracle 无机器 matcher——**归 REQ-040**（判据工作），本轮不抢跑；②`ui_impact_resolved` 未挂 PTR-PLAN-01——**本轮修**（一行 wiring，loop-definition PTR-PLAN-01 guards=[]）；③HTML 头部口径——规则文件自身矛盾（ui-prototype.md §5/§6 写 3-field、:123 门禁写 4-field、代码 hasProtoMetaHeader 执行 4-field）——**本轮统一 4-field**；④`--require-specs` 不进 doctor——**本轮修**；⑤populateUIPrototypeFact 死残迹——**本轮删**（含调用点与钉死测试）。**联合审查新增（§2.1）**：⑥AC↔CASE 无桥——**本轮机制化**（validate 检查项）；⑦构造性覆盖语义、非 UI 二分——**本轮声明**（规则/L2）；⑧S3 token 对账、S8 宇宙映射——**跨 stage 输入记录**（已入 L3-S3/L3-S8 整改方向）。

## 6. 注意力预算与渐进披露

总评：结构层典范（场景四件套的正确性由机器判——正反成对/负正比/防篡改都是 validate 说了算），错配在**两条 wiring 债让最关键防线退回文档层**，以及方法论入口缺渐进分岔。判定尺见 L3-README「注意力分配原则」。

### 6.1 当前错配（什么不对、为什么不对）

| # | 错配 | 为什么不对（L1 根据） |
|:--|:--|:--|
| 1 | `ui_impact_resolved` guard 已实现（guards.go:289-304）但 PTR-PLAN-01 的 guards 为空——unknown 拦截实际不生效 | 防止"UI 影响含糊进设计"的防线停留在 L2 文本（"禁止继续"）——恰是最弱承载层；一行 wiring 债让机制白写（D2：控制点在自然路径上的前提是挂上了） |
| 2 | `--require-specs` 不进 doctor 常规校验 | spec 级 100% 覆盖是文档承诺；机械可判的覆盖留在自觉层（载体三问 1） |
| 3 | `populateUIPrototypeFact` 是无消费者残迹（测试钉死 hook 永不匹配，run_test.go:1137-1148） | 公理三违例：信息量为零还占信道，且误导读者以为有 hook 门 |
| 4 | specification-planning 九步一步铺开 5 个子 skill（合计约 50KB 方法论） | 渐进披露缺失：ui_impact=none 时 UI 三 skill 全部无关，但入口没有显式分岔——agent 倾向一次全载，为无关内容付 token |

### 6.2 阅读预算（v4.0.1 按双轨汇聚顺序重排）

| 时机（轨归属） | 读什么 | 不读什么 |
|:--|:--|:--|
| 进入 S2（共同前置） | ARCHITECTURE 模板 + 绑定的 REQ；**既有模块演进时：完整模块包**（防新 stories 与既有 S-NNN 重复/矛盾） | 协议 S2 全文；5 个 skill 任何一个 |
| ui_impact=none | 到此为止，直奔 S3 | 全部 UI 方法论（三个 skill） |
| 系统轨（架构→facts） | ARCHITECTURE 模板 12 节 + ADR 模板 | 任何 UI/stories 方法论 |
| 用户轨（§A→stories） | user-story-design（用到才载） | 系统轨方法论 |
| 汇聚①（branches+oracle+交叉格清单） | scenario-model-design | flows/原型方法论 |
| fixtures | fixture 契约约定（scenario-model-design 内含） | — |
| 汇聚②（flows/原型） | user-flow-design / ui-prototyping（用到才载） | — |
| 深度自审+收口 | 自审三问（§4 第 4 步）+ validate 的**输出** | "必须正负成对/负正比≥N/生成物不可手改"的规则文本——引擎判，人不判；**AC→FR→BR 源头检查在汇聚①后就跑，不等收口** |

### 6.3 整改方向（v4.0.1 联合审查全景，14 项）

**机制项（8）**：
1. `ui_impact_resolved` 挂 PTR-PLAN-01（P0 wiring）；
2. `--require-specs` 进 doctor（P0 wiring，断点 3 左移）；
3. 删 populateUIPrototypeFact（含调用点+钉死测试）；
4. **AC↔CASE 桥（两段）**：AC→FR→BR 源头检查挂汇聚①后；全链检查在收口 validate——每条 AC 至少一 CASE 或**背书的 N/A**（类别+指针，清单进 ADR 拍板包；承载 D6"沉默不是不适用"）；
5. case_id 格式升 schema pattern（`CASE-XXX-YYY` 从模板约定升为引擎约束；具体 pattern 待终批）；
6. **skill 重构（两件）**：①specification-planning 重排为双轨汇聚顺序——渐进披露触发条件以修订后 §6.2 为契约（轨归属×时机×载哪个子 skill）；边界限定：其 step 6-9 属 S3/S4 的部分不动；"人闸只一个（ADR）"落地时须同步收敛其现有 5 条 Stop Conditions 中的 UI 包评审人触点（存废显式化）。②scenario-model-design 的 Workflow 顺序同步修正（现 step 7 才 reconcile stories——与双轨矛盾）；
7. **交叉格清单载体**（v4.0.1 新增，F7 处置）：汇聚①产物为模块包内手写清单（fact×FR×story 格→branch id 或无分支理由），validate 轻校验引用存在——把"三方交叉找全量"从叙述变成填空（承载 D4）；
8. **深度自审**（v4.0.1 新增，F8 处置）：收口前换实现者/e2e-tester/维护者三身份攻击产物（重点：每条负向 oracle 能否区分对错实现），结论入 ADR 包——REQ-040 matcher 空窗期的语义逼深（承载 D6+公理二）。

**声明项（3）**：
9. 单一分母原则入 L2 全局规则（**已落地 v1.4.1**——AC↔CASE 双向可达，承载 D1+D6；v4.0.1 修复了首次编辑被覆盖的事故）；
10. 构造性覆盖语义澄清（coverage.json 是设计声明，不作执行证据）；
11. 非 UI REQ 无 CASE 宇宙显式声明（REQ-040 Q-002 承接）。

**修正项（1）+跨 stage 输入（2）**：12. HTML 头部统一 4-field（ui-prototype.md §5/§6 的 3-field 改齐 :123 门禁与代码——§5③的对应项，v4.0.0 遗漏编号）；13. S3 token 对账（已入 L3-S3）；14. S8 宇宙映射（已入 L3-S8，作 REQ-040 输入）。

**待 owner 终批**：angles 冻结到 list/commit（C5：未被实战观测的机制不上第三档）；case_id pattern 具体格式。

**保持**：四件套"手写 2+生成 2+字节比对"承载结构；UI 先决用 gate not_ready；ARCHITECTURE 风险驱动结构。

## 变更记录

| 日期 | 版本 | 变更 |
|:--|:--|:--|
| 2026-08-14 | v1/v2 | 前两版（被判空洞→叙事不清；v2 的 expected_observable 系虚构） |
| 2026-08-14 | v3.0.0 | 叙事版；机制事实经 sub-agent 调查核实（oracle/polarity/四件套生成关系/guard 未 wiring 等如实入档） | owner 复核 |
| 2026-08-15 | v3.1.0 | 新增 §6 注意力预算与渐进披露（错配诊断/阅读预算/整改方向），判定尺引 L3-README | owner 指示：渐进披露、机制承载规范、削减平白叙述 |
| 2026-08-15 | v4.0.1 | **设计对抗审查处置（11 findings）**：P1×7——①双轨"并行"语义定为同 agent 顺序自由（非子代理）+冲突仲裁规则（架构约束旅程、挑战升 ADR）+用户轨前置读既有模块包；②AC 的 N/A 升为背书逃生门（类别+指针+进 ADR 拍板包，L2 铁律 1 同构）+"人闸只一个"与 skill 现有 UI 评审触点的存废显式化；③AC 桥拆两段（AC→FR→BR 源头查挂汇聚①后、全链收口复核）；⑤skill 重构拆两件（主 skill 重排+scenario-model-design Workflow 顺序修正）+触发条件以 §6.2 为契约+边界限定；⑦交叉格清单为三方交叉的载体（叙述变填空，D4）；⑧深度自审（三身份攻击负向 oracle，matcher 空窗期的语义逼深，D6+公理二）；⑨L2 分母表行修复（首次编辑被覆盖的事故——内联 replace 未回赋值，第二次写回抹掉）。P2×4——④§6.2 按双轨重排+补既有包前置读；⑥§2 两处现在时改如实+行号修正；⑩§6.3 补 HTML 头部项并升为 14 项全景；⑪D/公理标注补齐（双轨=D4+C4、桥=D6、分母=D1+D6） | owner 指示：注意力引导与思考深度推进是本步命脉，审查设计理念 |
| 2026-08-15 | v4.0.0 | **联合审查版**（S2+S3+S5+S7+S8 联动，sub-agent CASE 流水线端到端调查）：§1 增"验证源出生地"身份；§2.1 新增分母断裂地图（五断点+三事实修正+AC↔CASE 对位缺口）；§3 增双轨汇聚顺序与 AC 桥两条选用（否决"逐层拍板漏斗"——S0 形状硬搬造出不存在的判定层；否决六产物并行）；§4 时间线重写（stories 提前/fixtures 后移/两汇聚判据，hook 第二道保险改为如实——gate not_ready 才是真实保险）；§5 缺口段改处置台账；§6.3 升级为 11 项全景（机制 6+声明 3+跨 stage 输入 2，angles 冻结与 case_id pattern 待终批） | owner 指示：S2 与关联 stage 联合审查；产出对位下游所需；顺序不死板、按依赖重排 |
