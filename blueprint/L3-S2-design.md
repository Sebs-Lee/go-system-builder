# L3-S2 — 设计（Design）

> 层：第三层 ｜ 上游：L2 §S2 + L2「REQ 授权生命周期」分母条款 ｜ 版本 v4.0.0（v4.0.0 联合审查版：CASE 验证源断点地图+双轨汇聚顺序+AC↔CASE 桥；v3.1.0 增 §6；机制事实经调查核实，含 file:line）

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
| UI 原型门（hook 侧） | 写 `docs/contracts/` 且 ui_impact=changed 时，`populateUIPrototypeFact` 检查真相包完整性（7 文件+S/F-NNN 引 REQ-id+scenario 校验过），不完整→hook 拦截 | run.go:2040-2075；ui_scenario.go:66-120 |
| skill 群 | `specification-planning`（主：9 步流程，串起下面四个）；`scenario-model-design`（四件套方法）；`user-story-design`（S-NNN result-led）；`user-flow-design`（F-NNN/PATH checklist）；`ui-prototyping`（HTML 容器/侧栏契约） | skills/ |
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
| 生产顺序怎么定（v4.0.0 新增） | **双轨汇聚顺序**：系统轨（架构→facts）∥用户轨（§A→stories）→ 汇聚① rules/branches+oracle（facts×FR×stories 三方交叉找全量）→ fixtures（行为定型后造）→ 汇聚② flows/原型（stories×branches，PATH 绑定）→ 收口 generate/validate+AC 桥。**stories 提前**（现状排在 fixtures 后——story 是行为缺口的上游种子，建模后才写则模型用不上交叉视角；且 branch.story_refs 引 stories，先写免返工）、**fixtures 后移**（数据需求等 branch 定型） | 否决"逐层拍板漏斗"（S0 形状硬搬）——oracle 是 branch 字段非独立生产步骤，"判定层"不存在；否决"六产物并行铺开"（现状）——改一处全链重跑。人闸只一个（ADR 方向），其余机闸 |
| AC↔CASE 桥怎么承载（v4.0.0 新增） | `scenario validate` 新检查项：解析绑定 REQ 的 AC 表，每条 AC 须经 FR→BR→CASE 可达或有显式 N/A 理由——单一分母的完整形态 | 否决"只靠 §F 手抄矩阵"——分段无对账即断点；否决"S7 侧事后核"——错误的发现点越晚越贵 |

## 4. 怎么编排（时间线讲完一件事）

1. **架构决策**：按 ARCHITECTURE 模板 12 节起草；每个实质决策另录 ADR（决策/风险/后果/备选）入 `decisions/`。
2. **UI 影响分岔**：none → 直奔契约；changed → 进入真相包流程；unknown → 停：guard 不放行规划推进（见 §5 已知缺口）。
3. **双轨并进**（changed 时，v4.0.0 顺序修正）：
   - **系统轨**：数据模型/状态机定型 → facts 分区（架构词汇表的落盘）；
   - **用户轨**：从 REQ §A 写 stories（S-NNN 引 REQ-id）——行为缺口的上游种子，与架构并行；
   - **汇聚①行为全量**：rules/branches（正反成对，oracle 随 branch 写）= facts × FR × stories 三方交叉找全（含拒绝路径）；分支的 story_refs 指向已存在的 stories；
   - **fixtures**：branches 定型后造数据（setup/cleanup），一次造对；
   - **汇聚②可走查**：flows（F-NNN/PATH-*）+ 原型页（头部+侧栏契约）= stories 的旅程 × branches 的行为；browser_required 的 branch 在此绑定 PATH。
   **在既有模块包上演进**，不建新副本。
4. **收口**：`scenario generate` 产 cases/coverage → `scenario validate`（或 doctor）绿：正反比、引用存在性、字节冻结、**AC↔CASE 桥**（每条 AC 至少一 CASE 或显式 N/A）；不过 → 回汇聚①补负向分支或回 REQ 澄清 AC。
5. **提交规划证据**：登记 `planning_design` 证据（pass）→ 下一次 PreToolUse，controller 评估 GATE-PLANNING-DESIGN-COMPLETE → PTR-PLAN-01 自动迁移进 contracts——无人工拨表。
6. **UI 门的第二道保险**：~~hook 拦截~~（populateUIPrototypeFact 为无消费者残迹，测试钉死永不匹配——v4.0.0 裁定删除）；真实保险是 Quality Gate not_ready（真相包不齐 → missing 指回补包）。

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

### 6.2 阅读预算（渐进披露的正面样板）

| 时机 | 读什么 | 不读什么 |
|:--|:--|:--|
| 进入 S2 | ARCHITECTURE 模板 + 绑定的 REQ（模板字段即逼问） | 协议 S2 全文；5 个 skill 任何一个 |
| ui_impact=none | 到此为止，直奔 S3 | 全部 UI 方法论（三个 skill） |
| ui_impact=changed：建场景模型时 | scenario-model-design | stories/flows/原型方法论 |
| 写 stories / flows / 原型页 时 | user-story-design / user-flow-design / ui-prototyping（用到才载） | 其余 |
| 正确性核验（任何时刻） | validate 的**输出** | "必须正负成对/负正比≥N/生成物不可手改"的规则文本——引擎判，人不判 |

### 6.3 整改方向（v4.0.0 联合审查全景，11 项）

**机制项（6）**：
1. `ui_impact_resolved` 挂 PTR-PLAN-01（P0 wiring）；
2. `--require-specs` 进 doctor（P0 wiring，断点 3 左移）；
3. 删 populateUIPrototypeFact（含调用点+钉死测试）；
4. **AC↔CASE 桥**：`scenario validate` 新检查——每条 AC 至少经 FR→BR→CASE 可达或有显式 N/A（断点地图的核心修复，REQ-040 分母的直接加固）；
5. case_id 格式升 schema pattern（`CASE-XXX-YYY` 从模板约定升为引擎约束；具体 pattern 待终批）；
6. specification-planning 重构为**双轨汇聚顺序**（§4 新时间线：stories 提前/fixtures 后移/单一人闸/两个汇聚判据；子 skill 方法论保留，按轨渐进加载；替代 v3.1.0 的浅改方案"入口三路分岔"——分岔保留为双轨的入口形态）。

**声明项（3）**：
7. 单一分母原则入 L2 全局规则（AC↔CASE 双向可达，case.id 为 S2→S7 的唯一验证分母）；
8. 构造性覆盖语义澄清（coverage.json 是设计声明，不作执行证据）；
9. 非 UI REQ 无 CASE 宇宙显式声明（REQ-040 Q-002 承接）。

**跨 stage 输入记录（2）**：10. S3 token 对账（已入 L3-S3）；11. S8 宇宙映射（已入 L3-S8，作 REQ-040 输入）。

**待 owner 终批**：angles 冻结到 list/commit（C5：未被实战观测的机制不上第三档）；case_id pattern 具体格式。

**保持**：四件套"手写 2+生成 2+字节比对"承载结构；UI 先决用 gate not_ready；ARCHITECTURE 风险驱动结构。

## 变更记录

| 日期 | 版本 | 变更 |
|:--|:--|:--|
| 2026-08-14 | v1/v2 | 前两版（被判空洞→叙事不清；v2 的 expected_observable 系虚构） |
| 2026-08-14 | v3.0.0 | 叙事版；机制事实经 sub-agent 调查核实（oracle/polarity/四件套生成关系/guard 未 wiring 等如实入档） | owner 复核 |
| 2026-08-15 | v3.1.0 | 新增 §6 注意力预算与渐进披露（错配诊断/阅读预算/整改方向），判定尺引 L3-README | owner 指示：渐进披露、机制承载规范、削减平白叙述 |
| 2026-08-15 | v4.0.0 | **联合审查版**（S2+S3+S5+S7+S8 联动，sub-agent CASE 流水线端到端调查）：§1 增"验证源出生地"身份；§2.1 新增分母断裂地图（五断点+三事实修正+AC↔CASE 对位缺口）；§3 增双轨汇聚顺序与 AC 桥两条选用（否决"逐层拍板漏斗"——S0 形状硬搬造出不存在的判定层；否决六产物并行）；§4 时间线重写（stories 提前/fixtures 后移/两汇聚判据，hook 第二道保险改为如实——gate not_ready 才是真实保险）；§5 缺口段改处置台账；§6.3 升级为 11 项全景（机制 6+声明 3+跨 stage 输入 2，angles 冻结与 case_id pattern 待终批） | owner 指示：S2 与关联 stage 联合审查；产出对位下游所需；顺序不死板、按依赖重排 |
