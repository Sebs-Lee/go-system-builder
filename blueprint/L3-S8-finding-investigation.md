# L3-S8 — 发现调查（Finding Investigation）

> 层：第三层 ｜ 上游：L2 §S8 ｜ 版本 v3.0.0（叙事版；机制事实经调查核实——现状与 REQ-041 设计态分开标注）

## 1. 要实现什么

把验证发现转化为**证据支撑的规范缺陷报告**：先复现、深查根因、正确归类处置（六路由），产出 Builder-ready 的 canonical BUG——"看一眼就派修"在这里被结构性地挡住。

- 进入时：TR-008 落下的 draft BUG 批次（指纹去重过、默认 P0）+ 原始发现证据。
- 出去时：每个发现落到五归宿之一（accepted / final-rejected / duplicate / spec-rework / req-change）；每个 accepted BUG 有根因证据 + Closing Contract。
- 衡量：**报告不合格退回的是报告不是发现**（PTR-BUG-03 语义）——没有根因证据的 BUG 拿不到修复派发。

## 2. 手头有什么（真实机制）

| 机制 | 它能干什么 | 载体 |
|:--|:--|:--|
| S8 双相位 | `investigation`→`bug_report_review`（PTR-BUG-01：guard root_cause_evidence_complete，evidence=finding+root_cause_record）；裁决后 PTR-BUG-02（accepted→S9）/ PTR-BUG-03（rejected→回 investigation） | loop-definition.json:432-524 |
| BUG 实体生命周期（10 态） | draft→investigating→pending_approval→accepted→…→closed/rejected/duplicate；事件带 param 级 guard（bug_report_submitted 须 root_cause_evidence_complete+bug_closing_contract_complete；进入 investigating 时 attempt_count+1 并过重试上限） | entity_lifecycles.bug；assignment/bug_lifecycle.go |
| 六处置路由 | 实现缺陷→S9；假阳性/瞬态→final rejection（整批无 accepted 才可 TR-022 回 verification 开新轮）；duplicate→指向存活 canonical（closed 的不可再链）；规格→TR-023 回 planning.design（带 invalidate_affected_evidence）；REQ→TR-024 paused；报告不合格→PTR-BUG-03 退回重查 | agent-protocol.md:353-361；loop-definition.json:1622-1717 |
| 深查指令层（三层文本机制） | 协议六步（复现→根因→分类→三条件合并→canonical BUG→裁决）；bug-resolution skill（步骤 1-5 + **E2E 覆盖缺口硬步骤**："为什么 E2E 没拦住"，须查场景清单跑 e2e-coverage，契约坏而测试没红则 Closing Contract 必含 coverage gap 项）；R-P06 规则（**禁从 diff 开始**、六行 Closing Contract 模板、历史样本回归、数据修复 dry-run/apply 证据门槛） | agent-protocol.md:344-350；SKILL.md:31；docs/rules/bugfix-review.md |
| 合并三条件 | 仅当**同一用户可见矛盾+同一根因+兼容 Closing Contract** 才可合并（"never merge solely because they touch the same file"）——判断层规则 | agent-protocol.md:348；SKILL.md:32 |
| 注册指纹去重 | `sha256(finding_source+sorted(evidence_refs))` 在存活 BUG 间必须唯一，冲突即 ErrDuplicateBug；指纹折进 path 字段 | assignment/register.go:266-290 |
| BUG 模板（28 字段） | §3 假设表（Hypothesis/Evidence/confirmed-rejected）+ 四子字段 Closing Contract（repair scope/forbidden scope/before-fix evidence/retest contract）+ §7 去重声明位 | docs/reports/bugs/BUG-template.md；review-evidence.schema.json:45-84 |
| 独立性（机器） | `BUG_CLOSE_BY_FINDER_FORBIDDEN`：close 时强制 actor∉original_finder_agent_ids（身份级检查）；retest 的 original_finder_assigned guard | bug_lifecycle.go:163-180,262-322 |
| gate | GATE-BUG-DRAFTS-READY（本轮 blocking finding + root_cause_record/Investigator/complete）；GATE-CANONICAL-BUGS-ACCEPTED/REJECTED（bug_batch_record + accepted/rejected 结论） | registry.go:244-253 |
| 停止条件 | 根因无证据、REQ change、no-repair 处置缺证据、重试到顶、原 finder 不可用——停给人 | SKILL.md:52-57 |

## 3. 选了什么、为什么（含否决）

| 子问题 | 选用 | 否决与否决理由（减法） |
|:--|:--|:--|
| 深查怎么承载 | 三层文本机制（协议六步/skill 步骤+硬钩子/R-P06 规则）+ param 存在性 guard（root_cause_evidence、closing_contract 不填迁移不了）+ BUG 模板假设表 | 如实记录：**mechanism 词表/blast_radius/结构化根因是 REQ-041 设计态**（schema 无此字段）；现状深度=指令层逼问+param 门槛 |
| 去重靠什么 | 注册指纹去重（机器）+ 合并三条件（判断层）+ duplicate 迁移指向存活 canonical | 否决"机器自动合并"——合并是语义判断（三条件），机器只做"同发现不重复注册" |
| 怎么防"报告空转" | PTR-BUG-03 的"退报告不退发现"语义 + skill 停止条件 | 把劣质报告和无害发现区分开——劣质报告重写，不该拖着发现不处理 |
| 根因 guard | 现状是存根（evidenceBackedGuard 只查非空，guards.go:100-108 自认）——真语义靠 param 存在性+gate 证据要求 | 如实记录；REQ-041 的词表化是升级路径，但需先清 guard-theater 债 |
| 失败计数 | `same_contract_failure_count` 现为死计数器（恒 0、仅 checkRetryLimits 读） | 如实记录：SKILL 写了"fail→increment"但 runtime 无实现——REQ-041 FR-006 的前置 |
| 无修复出口 | TR-022 整批无 accepted 才可用，回 verification.delivery 开新轮 | 否决"发现全被否决就地结案"——没有新完整轮，旧轮证据状态就没交代 |

## 4. 怎么编排（时间线讲完一件事）

1. **接收**：draft BUG 批次落地（指纹去重已做，重复发现不重复建 BUG）。
2. **调查（investigation 相位）**：每个 BUG 按六步走——复现（确定性或确定性证据）→根因（从证据不从 diff；假设表逐条 confirmed/rejected）→ 影响与同类（合并三条件判断；E2E 类发现必过覆盖缺口钩子）→ Closing Contract 四子字段（R-P06 六行模板落 retest contract）。
3. **报告（bug_report_submitted）**：param 门槛（root_cause_evidence+bug_closing_contract）不填过不了迁移；PTR-BUG-01 进评审。
4. **裁决（bug_report_review 相位）**：主会话逐 BUG 过六处置路由——accepted（PTR-BUG-02 → S9 修复派发）/ rejected（附 rejection_reason；报告劣质则 PTR-BUG-03 退回重查）/ duplicate（duplicate_of 指向存活 canonical）/ spec_rework（TR-023 回 planning 并失效受影响证据）/ req_change（TR-024 paused 交人）。
5. **收口**：整批无 accepted 且全部 final-rejected/duplicate 闭环 → TR-022 回 verification.delivery 开新完整轮。

## 5. 期望效果

走完 S8：

- **每个发现有归宿**：五路出清，无"发现的 bug 就消失了"；修复派发拿到的 BUG 都带根因证据+边界+重验契约；
- **结构性防住**：浅查直进修复（param 门槛+假设表）、一因拆多（指纹去重+三条件）、自己关自己的 BUG（BUG_CLOSE_BY_FINDER_FORBIDDEN）、E2E 漏网无人问（覆盖缺口硬步骤）、劣质报告拖死发现（退报告不退发现）；
- **交给 S9**：accepted BUG（含 repair/forbidden scope + Closing Contract）——修复授权面与判据锚。

**如实记录的现状边界**（REQ-041 设计态）：①根因无受控词表（去重靠指纹+三条件判断，无 mechanism 检索）；②blast_radius 非结构化（impact 是字符串数组）；③root_cause/closing_contract 的 PTR-BUG-02 guard 是存根；④same_contract_failure_count 不自增（深查触发无机器支撑）；⑤canonicalBug 28 字段的书写负担——REQ-041 的减法对象。

## 变更记录

| 日期 | 版本 | 变更 |
|:--|:--|:--|
| 2026-08-14 | v1/v2 | 前两版（被判空洞→叙事不清；v2 的七步/词表是设计态冒充现状） |
| 2026-08-14 | v3.0.0 | 叙事版；现状（六处置路由/param 门槛/指纹去重/关闭独立性）与 REQ-041 设计态分栏 | owner 复核 |
