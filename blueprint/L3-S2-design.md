# L3-S2 — 设计（Design）

> 层：第三层 ｜ 上游：L2 §S2 + L2 跨阶段全局规则「单一验证分母」 ｜ 机制事实经调查核实，含 file:line

## 1. 要实现什么

产出架构决策（ADR）+（UI 影响时）**模块全量场景真相包**——它是后续契约、任务、验证用例的唯一事实源。

- 进入时：绑定生效的 REQ（含 ui_impact 三值）。
- 出去时：`GATE-PLANNING-DESIGN-COMPLETE` 可满足——一条 `planning_design` 证据（Architect/Orchestrator，pass）+ 架构文档（Status: locked）+ 真相包过校验。
- 衡量：**全量与双极性**——场景以模块为全集（不是本需求的子集副本），每条规则的正反分支都有 oracle；`scenario validate` 说绿才算绿。

**身份**：S2 不只是"设计 stage"，它是**全链验证源的出生地**——本 stage 产出的四件套质量（oracle 七字段完备、PATH 绑定真实、browser_required 正确、AC↔CASE 可达）直接决定 S7 完整性门的分母成色。因此 S2 的设计以"下游对位"为准绳，不以其自身机制为限。

## 2. 手头有什么（真实机制）

| 机制 | 它能干什么 | 载体 |
|:--|:--|:--|
| ARCHITECTURE 模板 | 12 节结构：目标/上下文/容器/模块职责（含排除范围列）/数据流/状态机/数据模型/接口/安全/性能基线/部署回滚/锁定；定稿翻 `状态：locked`（PTR-PLAN-01 只登记 locked 的架构文档） | `docs/design/architecture/ARCHITECTURE-template.md` |
| ADR 目录与签核包 | 决策记录写 `docs/design/decisions/ADR-<id>.md`，含 Depth Self-Review 与 Endorsed N/A 两个固定段——S2 唯一人闸的拍板包 | ADR-template.md |
| 场景四件套 | **手写两件**：`scenario-model.json`（facts 分区 + rules→branches，branch 带 `polarity`(positive/negative)、`oracle`、`fixture_id`、`story_refs`/`flow_refs`、`browser_required`）+ `fixture-contract.json`（合成数据装配 setup/cleanup）；**生成两件**：`cases.json` + `scenario-coverage.json`（引擎从手写件派生，字节级防篡改） | `docs/design/prototypes/<module>/`；schema 在 internal/schema/assets/ |
| oracle 结构 | 正向必填 visible[]/terminal_state/persisted_effects/forbidden_side_effects；负向加 rejection（稳定拒绝码/消息）/expected_state/recovery（N/A 须引 recovery_source_refs+recovery_reason） | scenario-model.schema.json:18-53 |
| cross-matrix | 模块包内手写的汇聚①载体：fact×FR×story 每格指向 branch 或记无分支理由——"找全"从叙述变填空；机器地板：每 fact/每 story 至少一格、branch 的 rule 须真实引用该格 FR、理由 ≥8 字符含字母、req_ref 只认 bound REQ | cross-matrix.go |
| `loop-harness scenario` | `generate`（校验+原子生成 cases/coverage）/`validate`（源头校验+生成物字节比对+AutoSpecs——spec 树存在即查 Playwright 覆盖）/`bridge`（AC→FR→BR 源头检查） | internal/scenario/engine.go；run.go:169 |
| 覆盖率与正反比 | required branch 100%（构造性——设计声明非执行证据）；负:正 ≥ coverage_profile 下限（critical=3/rule-dense=2/ordinary=1）；doctor/validate 自动跑 | engine.go:354-363,1220-1238 |
| 原型包规则 | index.html+stories.md(S-NNN 恰三位)+flows.md(F-NNN/PATH-*)+页面 HTML（4-field 头部：设计代数/更新/路由/index 链接）+四件套+cross-matrix；模块任一改动→E2E 全模块回归 | `docs/rules/ui-prototype.md` |
| skill 群 | `specification-planning`（主：Step 0 分流+双轨汇聚+S3/S4 步骤+Planning Evidence Envelopes 节）；`scenario-model-design`（四件套方法+oracle 字段语义表）；`user-story-design`/`user-flow-design`/`ui-prototyping`（按轨按需加载） | skills/ |
| 规划门推进 | PTR-PLAN-01（design→contracts）挂 GATE-PLANNING-DESIGN-COMPLETE：locked 架构文档（磁盘声明或已注册）+ planning_design 证据——磁盘事实可直接满足，登记在 commit 时完成 | evaluator.go:188-244 |
| angles 注册表 | 模块级 append-mostly 自检清单（ANG-{MODULE}-{NNN}，黑名单禁通用词），S5/S7 的 angle_complete guard 消费 | runtime/angles.go；`loop-harness angles` |

### 2.1 分母链全景（联合审查确立）

出生→冻结（四件套生成+字节防篡改）与 S7 gate 求值两段各自机械连续；**单一验证分母**（L2 全局规则）要求 AC↔CASE 双向可达（每条 AC 至少一 CASE 或经背书的 N/A）：

| 链段 | 承载 |
|:--|:--|
| AC→FR→BR | `scenario bridge`（源头检查，汇聚①后即可跑）+ validate 复核 |
| AC↔CASE 全链 | `scenario validate`（需 cases 生成物） |
| cross-matrix join | validateCrossMatrix 接入 bound REQ（FR 表对账/branch↔source_refs 真实引用/每 fact 每 story 至少一格） |
| N/A 背书 | 类别+指针（非功能 AC→NFR 编号；范围外→REQ §A4 负空间条目）；自由文本拒绝——无背书的 N/A 是从分母里静默移除验证项 |
| 分母右端 | S7 按 CASE_ID 粒度计数（REQ-040 承接）；S8 反查宇宙映射（REQ-040 输入） |

**三个设计边界**：①`scenario-coverage.json` 的 required_branch_coverage 是构造性 100%（设计时声明），不得作执行覆盖证据（执行覆盖由 S7 证据计数承担）；②angle_declaration 是评审组声明证据（REQ-003），S2 的 angles 注册表只是其交接输入；③四件套只对 ui_impact=changed 强制——非 UI REQ 无 CASE 宇宙（显式设计边界，REQ-040 Q-002 预留"无场景包⇒BLOCKED 而非跳过"）。

## 3. 选了什么、为什么（含否决）

| 子问题 | 选用 | 否决与否决理由（减法） |
|:--|:--|:--|
| 怎么逼出全量双极性场景 | scenario-model 的 branch 字段（polarity/oracle 必填）+ 引擎校验（正反比、witness 引用注册 fact、S/F-NNN 必须真实存在） | 否决"自由格式设计文档"——无结构则无全量可言；字段即逼问 |
| 怎么防"需求私有副本" | 模块目录为唯一家（REQ 只是 source_refs） | 否决"每 REQ 一份场景"——副本必然漂移，模块才是全集的分母 |
| 怎么防生成物篡改 | cases/coverage 由引擎生成 + validate 字节级比对 | 否决"手写 cases"——生成物不可手改，源头单点 |
| 怎么管 UI 影响 | 三值 + `ui_impact_resolved` guard 挂 PTR-PLAN-01 + gate not_ready（真相包不齐→missing 指回补包） | 否决"仅文档提醒"；也否决"hook 侧 fact 拦截"——无消费者的死机制已删，gate 才是真实保险 |
| 架构决策怎么落 | ARCHITECTURE 模板 12 节 + ADR 另录 decisions/（含自审与 N/A 段） | 否决"再造 ADR 模板段"——已有模板与目录惯例，不重复 |
| spec 覆盖 100% | AutoSpecs：spec 树存在即 doctor/validate 强制（缺失不罚——S6+ 工件时序） | 否决"显式开关"——机械可判的覆盖不留在自觉层 |
| 模块视角积累 | angles 注册表（append-mostly） | 否决"每次全靠 REQ 内描述"——模块教训要跨 REQ 存活 |
| 生产顺序怎么定 | **双轨汇聚**：系统轨（架构→facts）∥用户轨（§A→stories）→ 汇聚① rules/branches+oracle（facts×FR×stories 三方交叉，cross-matrix 为载体）→ fixtures（行为定型后造）→ 汇聚② flows/原型（PATH 绑定）→ 收口。**stories 前置**（branch.story_refs 的依赖）、**fixtures 后移**（数据需要分支定型）。**并行语义**：同一 agent 的顺序自由，不是子代理派发；**冲突仲裁**：架构约束旅程，stories 挑战架构时升 ADR 人闸裁；**用户轨前置**：既有模块演进先读完整模块包再写新 stories | 承载 D4+C4。否决"逐层拍板漏斗"（S0 形状硬搬——oracle 是 branch 字段非独立生产步骤）；否决"六产物并行铺开"（改一处全链重跑） |
| AC↔CASE 桥怎么承载 | **两段检查**：AC→FR→BR 源头段挂汇聚①后（不需生成物，早查早回）；全链段在收口 validate 复核。N/A 须背书（类别+指针），N/A 清单进 ADR 拍板包 | 承载 D6。否决"只靠 §F 手抄矩阵"；否决"S7 侧事后核"；否决"N/A 一句理由过门" |

## 4. 怎么编排（时间线讲完一件事）

1. **架构决策**：按 ARCHITECTURE 模板 12 节起草；每个实质决策另录 ADR 入 `decisions/`。
2. **UI 影响分岔（Step 0）**：none → 直奔契约；changed → 进入真相包流程；unknown → 停：`ui_impact_resolved` guard 不放行规划推进。
3. **双轨并进**（changed 时）：
   - **系统轨**：数据模型/状态机定型 → facts 分区；
   - **用户轨**：从 REQ §A 写 stories（S-NNN 引 REQ-id）；既有模块演进时先读完整模块包；
   - **汇聚①行为全量**：rules/branches（正反成对，oracle 随 branch 写）= facts × FR × stories 三方交叉找全（含拒绝路径），交叉格清单（cross-matrix）为载体；汇聚①后即跑 AC→FR→BR 源头检查；
   - **fixtures**：branches 定型后造数据；
   - **汇聚②可走查**：flows + 原型页（4-field 头部）= stories 的旅程 × branches 的行为；browser_required 的 branch 绑定 PATH。
   在既有模块包上演进，不建新副本。
4. **深度自审（收口前）**：换身份攻击自己的产物——实现者："哪条 oracle 我落不了地或区分不了对错实现？"；e2e-tester："哪条负向 CASE 我按七维度取不了证？"（visible/terminal_state/persisted_effects/forbidden_side_effects，负向加 rejection/expected_state/recovery）；维护者："哪条规则会和模块演进打架？"——结论一段话入 ADR 包的 Depth Self-Review 段。
5. **收口**：翻 ARCHITECTURE `状态：locked` → `scenario generate` 产 cases/coverage → `scenario validate` 绿（正反比/引用存在/字节冻结/cross-matrix/AC↔CASE 全桥）→ 登记 planning_design JSON 信封（见 specification-planning 的 Planning Evidence Envelopes 节）。
6. **自动推进**：下一次 PreToolUse，controller 评估 GATE-PLANNING-DESIGN-COMPLETE → PTR-PLAN-01 自动迁移进 contracts（register_design_documents 登记架构文档）——无人工拨表。

## 5. 期望效果

走完 S2：

- **全量可机检的场景资产**：模块真相包过 `scenario validate`，正反比达标，生成物防篡改——S7 的用例分母与 fixture 装配直接来源于此；
- **结构性防住**：需求私有副本（模块为唯一家）、只写正向（polarity+正反比）、oracle 散文化（schema 必填结构）、UI 影响含糊进契约（guard+gate 双门）、交叉遗漏（cross-matrix 机器地板）、N/A 逃逸（背书制）、模块教训流失（angles）；
- **交给 S3**：架构决策（locked）+ 真相包（锁定指纹）+ planning_design 证据——契约的引用输入。

已知缺口（供第四层）：oracle 语义质量无机器 matcher——归 REQ-040（判据工作），深度自审是空窗期的判断层逼深手段。

## 6. 注意力预算与渐进披露

总评：结构层典范（场景四件套的正确性由机器判——正反成对/负正比/防篡改都是 validate 说了算），方法论入口按轨分岔按需加载。判定尺见 L3-README「注意力分配原则」。

### 6.1 阅读预算（按双轨汇聚顺序）

| 时机（轨归属） | 读什么 | 不读什么 |
|:--|:--|:--|
| 进入 S2（共同前置） | ARCHITECTURE 模板 + 绑定的 REQ；既有模块演进时：完整模块包 | 协议 S2 全文；5 个 skill 任何一个 |
| ui_impact=none | 到此为止，直奔 S3 | 全部 UI 方法论（三个 skill） |
| 系统轨（架构→facts） | ARCHITECTURE 模板 12 节 + ADR 模板 | 任何 UI/stories 方法论 |
| 用户轨（§A→stories） | user-story-design（用到才载） | 系统轨方法论 |
| 汇聚①（branches+oracle+cross-matrix） | scenario-model-design | flows/原型方法论 |
| fixtures | fixture 契约约定（scenario-model-design 内含） | — |
| 汇聚②（flows/原型） | user-flow-design / ui-prototyping（用到才载） | — |
| 深度自审+收口 | 自审三问（§4 第 4 步）+ validate 的**输出** + Planning Evidence Envelopes 节 | "必须正负成对/负正比≥N/生成物不可手改"的规则文本——引擎判，人不判；AC→FR→BR 源头检查在汇聚①后就跑，不等收口 |
