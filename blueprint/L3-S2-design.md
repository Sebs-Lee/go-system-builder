# L3-S2 — 设计（Design）

> 层：第三层 ｜ 上游：L2 §S2 ｜ 版本 v3.1.0（v3.0.0 叙事版 + §6 注意力预算；机制事实经调查核实，含 file:line）

## 1. 要实现什么

产出架构决策（ADR）+（UI 影响时）**模块全量场景真相包**——它是后续契约、任务、验证用例的唯一事实源。

- 进入时：绑定生效的 REQ（含 ui_impact 三值）。
- 出去时：`GATE-PLANNING-DESIGN-COMPLETE` 可满足——一条 `planning_design` 证据（Architect/Orchestrator，pass）+ 架构文档 + 真相包过校验。
- 衡量：**全量与双极性**——场景以模块为全集（不是本需求的子集副本），每条规则的正反分支都有 oracle；`scenario validate` 说绿才算绿。

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

## 4. 怎么编排（时间线讲完一件事）

1. **架构决策**：按 ARCHITECTURE 模板 12 节起草；每个实质决策另录 ADR（决策/风险/后果/备选）入 `decisions/`。
2. **UI 影响分岔**：none → 直奔契约；changed → 进入真相包流程；unknown → 停：guard 不放行规划推进（见 §5 已知缺口）。
3. **场景建模**（changed 时）：`specification-planning` 串起子技能——facts 分区 → rules/branches（正反成对，oracle 逐分支写）→ fixtures（合成数据+cleanup）→ stories（S-NNN 引 REQ-id）→ flows（F-NNN/PATH-*）→ 原型页（头部+侧栏契约）。**在既有模块包上演进**，不建新副本。
4. **生成与校验**：`scenario generate` 产 cases/coverage → `scenario validate`（或 doctor）绿；正反比不足→回 3 补负向分支。
5. **提交规划证据**：登记 `planning_design` 证据（pass）→ 下一次 PreToolUse，controller 评估 GATE-PLANNING-DESIGN-COMPLETE → PTR-PLAN-01 自动迁移进 contracts——无人工拨表。
6. **UI 门的第二道保险**：若真相包不齐而有人试图写契约，hook 在写瞬间拦截（ui_contract_before_prototype）并指路补包。

## 5. 期望效果

走完 S2：

- **全量可机检的场景资产**：模块真相包过 `scenario validate`，正反比达标，生成物防篡改——S7 的用例分母与 fixture 装配直接来源于此；
- **结构性防住**：需求私有副本（模块为唯一家）、只写正向（polarity+正反比）、oracle 散文化（schema 必填结构）、UI 影响含糊进契约（guard+hook 双门）、模块教训流失（angles）；
- **交给 S3**：架构决策 + 真相包（锁定指纹）——契约的引用输入。

**如实记录的已知缺口/漂移**（供第四层修复清单）：①我 v2 版所写 `expected_observable/matcher` 为虚构——真实字段是 **oracle（检查点式，无机器 matcher）**，"预期事实机器核对"要等 REQ-040 的判据工作补齐；②`ui_impact_resolved` guard 已实现但**未挂到 PTR-PLAN-01**（guards 注释自认 wiring 是 follow-on，loop-definition.json:147 guards 为空）——unknown 拦截目前实际不生效；③HTML 头部字段三处说法不一（agent-protocol "4-field"/规则 §5 "3 字段"/代码实际 4 token 含"设计代数"）；④`--require-specs` 的 spec 级 100% 不在 doctor 常规校验内。

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

### 6.3 整改方向

- **P0 wiring**（一行级，性价比最高）：`ui_impact_resolved` 挂 PTR-PLAN-01；`--require-specs` 进 doctor；
- **删减**：删 populateUIPrototypeFact；HTML 头部字段三处口径（协议"4-field"/规则"3 字段"/代码 4 token）以代码为准统一；
- **左移（渐进披露结构化）**：入口 skill 第一步改为显式三路分岔（none→直奔契约 / changed→真相包流程 / unknown→指向澄清动作），把"何时无需读"写在最前面；
- **保持**：四件套"手写 2 + 生成 2 + 字节比对"的承载结构（模块为唯一家、生成物防篡改——机制已最优）。

## 变更记录

| 日期 | 版本 | 变更 |
|:--|:--|:--|
| 2026-08-14 | v1/v2 | 前两版（被判空洞→叙事不清；v2 的 expected_observable 系虚构） |
| 2026-08-14 | v3.0.0 | 叙事版；机制事实经 sub-agent 调查核实（oracle/polarity/四件套生成关系/guard 未 wiring 等如实入档） | owner 复核 |
| 2026-08-15 | v3.1.0 | 新增 §6 注意力预算与渐进披露（错配诊断/阅读预算/整改方向），判定尺引 L3-README | owner 指示：渐进披露、机制承载规范、削减平白叙述 |
