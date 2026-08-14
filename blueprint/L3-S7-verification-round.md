# L3-S7 — 完整验证轮（Verification Round）

> 层：第三层 ｜ 上游：L2 §S7 ｜ 版本 v3.0.0（叙事版；机制事实经调查核实——**现状与 REQ-040 设计态明确分开标注**）

## 1. 要实现什么

同一轮内的完整发现：三组正交验证（交付符合性 / 工程质量 / 真实浏览器行为）按序跑完，四守卫判定干净轮——干净轮是通向验收的唯一门票。

- 进入时：S6 完成报告批次（TR-006 已把 review.round+1、相位置 delivery）。
- 出去时：要么干净轮记录成立（PTR-VERIFY-04→TR-009 进 S10），要么产出阻断发现（TR-008 进 S8）。
- 衡量：**同轮、全维度、无失效证据、无未闭阻断缺陷**——四守卫是纯函数，谁算都一个结果。

## 2. 手头有什么（真实机制）

| 机制 | 它能干什么 | 载体 |
|:--|:--|:--|
| verification 相位机 | delivery→qa→e2e_browser→clean_round_evaluation→clean_round_passed，前组是后组的守卫；PTR-VERIFY-01/02/03 各挂一组 gate（PreToolUse 自动迁移） | loop-definition.json phase_machines.verification |
| 三组 gate | 各要求**一条当前轮 review_record**（delivery_review/Delivery Verifier、qa_review/QA、e2e_review/E2E Browser，各 conclusion=pass）+ 迁移级 required_evidence 含 team_manifest_record | registry.go:190-198；loop-definition.json |
| **干净轮四守卫（纯函数）** | `EvaluateCleanRound`：①同轮性（七类验证证据 review_round==当前轮，历史 planning/building 证据不否决）②维度完备（**三个 team kind 必须都在**+每职责当前轮 valid PASS 证据，无职责集合也 fail）③无失效（当轮 invalid 即 fail）④无未闭阻断（P0 阻断态枚举；已闭 P0 还须有当轮 targeted_reverification）；引擎守卫/CLI `verification clean-round`/gate 三处复用同一函数 | internal/verification/clean_round.go:41-135；guards.go:258-262 |
| angle 完成度 guard | PTR 的 delivery/qa/e2e_angle_complete 消费 angle_declaration 证据+team manifest 的 inherited_angles；维度三值不可跨用；ui_impact=none 时 E2E 允许 N=1 的放宽 | transition/angle_complete_guard.go |
| 真交互约束（协议层） | E2E 须"从声明入口进入、按序点击声明的控件、避免 URL 直跳（除非 flow 声明该入口）、每步记录可见断言"；每个 PATH-* 要么有 step-level evidence 要么显式 N/A 带理由 | agent-protocol.md:323,330-331 |
| e2e-tester 角色定义 | 输出八件套：stack-state 表/测试架构图/API CRUD walk（真实 HTTP code）/CDP findings/证据清单（JSONL+PNG）/状态码分布/BUG drafts/幂等重执行说明；负向 CASE 七维度（visible/terminal_state/persisted_effects/forbidden_side_effects/rejection/expected_state/recovery，recovery N/A 须 source_refs+理由）；spec 只写 web/e2e/<module>/，禁发明 flow/CASE（缺失→DV finding 停止） | agents/e2e-tester.md:36-58 |
| delivery-verifier / qa 角色 | 前者："REQ is a source reference, not a smaller test scope"，一职责一结论；后者：PASS/N/A/finding 三态+可复现证据+BUG drafts，禁"用工具替代判断""接受缺失的负向覆盖" | agents/delivery-verifier.md；agents/qa.md |
| 发现流转 | TR-008（guard blocking_findings_present，on_guard_failure=reject）：`recordFindingBatch` 按指纹（reporter+body+path 的 sha256）去重建 draft BUG（默认 P0），`set_bug_phase_investigation` 进 S8 | loop-definition.json；actions.go:407+ |
| 轮次管理 | review.round 由 TR-006/TR-012 的 start_review_round 递增并清 clean_round；干净轮成立则 review.clean_round=当前轮 | actions.go:219-239,591-593；loop-state.schema.json |
| clean-round-evaluation skill | 第一步就跑 `verification clean-round` CLI；结果严格三值 pass/incomplete/blocked，禁加权分；targeted re-verification 不满足任何守卫；旧记录不可改写只能失效重跑 | skills/clean-round-evaluation/SKILL.md |

## 3. 选了什么、为什么（含否决）

| 子问题 | 选用 | 否决与否决理由（减法） |
|:--|:--|:--|
| 三组顺序还是并行 | **顺序**（delivery→qa→e2e）：前组的干净是后组的守卫，同轮证据不重叠 | 否决"三组并行省时间"——并行会让发现互相追尾（delivery 的发现 invalidate 掉 qa 正在引的证据）；顺序是正确性选择 |
| 干净轮怎么判 | 纯函数四守卫（一处实现三处复用），无加权、无部分分 | 否决"评分制"——连续比精确更重要；三值结果让"差一点"无处藏 |
| 覆盖怎么数 | **按职责**：从 team manifest 收集 responsibility_ids，每职责要当轮 valid PASS | 如实记录：**按用例计数（CASE 粒度）是 REQ-040 设计态，未实现**——当前粒度到职责为止；E2E 侧靠 PATH-* step-level evidence 的协议要求兜 |
| E2E 真交互怎么保 | 协议契约（入口/按序/禁直跳/逐步断言）+ 角色定义（Forbidden Actions+输出八件套含 CDP） | 如实记录：**机械核对（trace 交互事件/HAR 匹配）是 REQ-040 设计态**——当前真交互是判断层承诺+证据抽查 |
| 发现怎么进缺陷流 | 指纹去重的 finding batch → draft BUG（默认 P0）→ S8 | 否决"发现直接修复"——那是 Fixes that Fail 的入口 |
| 维度跨用 | angle guard 禁止（delivery 的 angle 不能满足 e2e_angle_complete） | 否决"复用省证据"——跨用即口径造假 |

## 4. 怎么编排（时间线讲完一件事）

1. **组队**：三个 workgroup（delivery_verifier/qa/e2e_browser 三 team kind 必须都在——四守卫 Check 2 的硬前提）；每任命两阶段激活；E2E 组每任命预载 e2e-browser-testing+playwright-e2e 并枚举将执行的 USER-FLOW 文件与 PATH-* ID。
2. **delivery 相位**：交付验证者对照规格逐职责核——REQ 覆盖/契约实现/任务完成/集成/回归；结论+证据登记（delivery_review，当轮）。angle guard 查本维度 angle 完成度 → gate 过 → PTR-VERIFY-01 进 qa。
3. **qa 相位**：QA 组按四基线+风险触发维度出独立结论（qa_review）。→ PTR-VERIFY-02 进 e2e_browser。
4. **e2e_browser 相位**：一任命一职责，从声明入口按序真实交互；负向 CASE 七维度逐步取证（JSONL/PNG/CDP）；spec 缺 flow/CASE → 停止并作为 DV finding 上报。结论（e2e_review）→ PTR-VERIFY-03 进评估。
5. **发现即分流**：任一相位产出阻断发现（finding_record，blocking，requested_event=blocking_findings_reported）→ TR-008 直接进 S8——不必等三轮跑完。
6. **干净轮评估**：无阻断发现 → clean-round-evaluation skill → CLI 求值四守卫 → pass 则 PTR-VERIFY-04 落 clean_round 记录（review.clean_round=当前轮）→ TR-009 进 S10；incomplete → PTR-VERIFY-05 回 delivery 开新轮。

## 5. 期望效果

走完 S7（干净轮成立时）：

- **同轮全维度可复算**：四守卫纯函数任何人重算同结果；三 team kind+全职责+无失效+无未闭阻断，一条不满足都不叫干净；
- **结构性防住**：跨轮凑数（同轮性）、维度缺失（三 team kind 硬前提）、维度跨用（angle guard）、"局部重验当完成"（targeted re-verification 不满足任何守卫）、发现被吞（指纹去重直进 BUG 流）；
- **交给 S10**：干净轮记录+冻结指纹——验收的前提；**交给 S8**（若有）：去重后的 draft BUG 批次。

**如实记录的现状边界**（REQ-040 设计态，未实现——本文不描述为现状）：①机械预检（P 系）/原始证据清单/佐证记录/`backed_by_raw_evidence_and_corroboration` 守卫——全仓零实现；②CASE 粒度覆盖计数（当前=职责粒度）；③真交互的机器核对（当前=协议+角色承诺+证据抽查）；④guard-theater 现象（guards.go:100-106 自注：多数 guard 是证据非空桩，真语义在 validateCurrentEvidence 与 EvaluateCleanRound）。这些是 REQ-040 的改造对象，实现后本文 §2/§4 相应升级。

## 变更记录

| 日期 | 版本 | 变更 |
|:--|:--|:--|
| 2026-08-14 | v1/v2 | 前两版（被判空洞→叙事不清；v2 的预检/佐证剧本是设计态冒充现状） |
| 2026-08-14 | v3.0.0 | 叙事版；现状（四守卫/相位机/angle/发现流转）与 REQ-040 设计态明确分栏 | owner 复核 |
