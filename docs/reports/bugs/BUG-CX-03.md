# Canonical BUG: BUG-CX-03

> Status: fixed
> Severity: P1
> Runtime ref: N/A（模板仓库自审——S2 agent 视角复杂度审查）
> Found in review round: complexity-review-1
> Finding evidence: 本文件 §2/§3 所引 file:line（HEAD=9cd52fa 实测）
> Original responsibility: agent-protocol #s2 / specification-planning skill / scenario-model-design skill
> Owner: PM / Architect

## 1. Fingerprinted Specification Chain

| Order | Kind | ID | Path | Version | SHA-256 | Relevant clauses |
|:--|:--|:--|:--|:--|:--|:--|
| 1 | BUG | BUG-CX-03 | `docs/reports/bugs/BUG-CX-03.md` | v1 | n/a | all |
| 2 | rule | agent-protocol #s2 | `docs/agent-protocol.md` | current | n/a | #s2 |
| 3 | skill | specification-planning | `skills/specification-planning/SKILL.md` | v2.1.0 | n/a | Procedure |
| 4 | skill | scenario-model-design | `skills/scenario-model-design/SKILL.md` | v1.2.0 | n/a | Authority/Workflow |
| 5 | design | L3-S2 | `blueprint/L3-S2-design.md` | v4.0.4 | n/a | 双轨/收敛 |

## 2. Observed Contradiction

**症状：S2 的引导层停留在 v3 口径——agent 按 protocol 字面走完自认"完成"，然后在 PTR-PLAN-02 被新 guard 拦截且不知为何；skill 内多处死引用与字段谜语。**

| Field | Value |
|:--|:--|
| expected | protocol #s2 的 done_when = v4 八文件包 + generate/validate/bridge 绿 + AC 桥全达；skill 的每个字段/口径就地可查；跳过条件在流程开头 |
| observed | ① **protocol #s2 done_when 是 v3 口径**（agent-protocol.md:212-214：只要求 index.html+stories+flows+4 字段头）——scenario-model.json、cross-matrix.json、fixture-contract.json、cases.json、`scenario generate/validate/bridge` 全不在内；全 protocol grep "cross-matrix" 零命中。按 protocol 走完 → 被 scenario_bridge_checked（PTR-PLAN-02，bridge.go:205）拦截，不知为何。② **"meaningful cell" 口径只在 Go 注释**（cross_matrix.go:17-20：地板=每 fact/每 story ≥1 格，组合是人的猎杀判断非笛卡尔积）——skill（SKILL.md:52-53）与 rules/scenario-model.md:175 均未写：谨慎型 agent 填 10×8×6=480 格爆炸，随意型漏 story 被拦。③ **两处死引用**：scenario-model-design SKILL.md:9-10 指向架构文档不存在的 §19（实际 §1-§12）；specification-planning SKILL.md:69-71 的 "seven dimensions" 全库仅此一处（隐含答案=rules:126-127 的七项断言，未建映射）。④ **oracle 七字段只列名不释义**（架构 §4.2:110-126 / rules §4:81-87 / 模板占位符是元描述）——terminal_state vs expected_state（negative 专属）、persisted_effects 写什么、visible 是否限 UI，全靠猜。⑤ **ui_impact=none 跳过条件排在 step 10**（SKILL.md:77-79）——线性阅读的 agent 先做完 1-9 才被告知可跳，且 "skip 3-9" 用编号指代命名结构。⑥ **no_branch_reason ≥8 字符含字母的下限只在报错文案**（cross_matrix.go:133-134）——模板占位符是中文「一句话」（很可能 <8 字符触发误报）。⑦ 命令调用形式不一：`loop-harness scenario bridge`（SKILL.md:54）vs `go run ./cmd/loop-harness scenario generate`（scenario-model-design:132-135）。⑧ ADR sign-off（SKILL.md:36-37"the single human gate"）与 protocol #s2 human_gateway（:217「only req_amendment or unrecoverable_business_decision」）矛盾。⑨ rule 级 `risk` 是自由字符串无枚举无指导（模板唯一示例 "rule-dense" 与 coverage_profile 枚举撞名）。⑩ "顺序自由"（spec-planning:26-27）与 "stories BEFORE rules"（scenario-model-design:74-76 硬约束）各说一半——convergence-1 内部实有严格先序 |
| user/data/system impact | S2 是概念最重的 stage（15+ 独有概念）；引导断层使"填空"退化为"考古+试错"，每次报错来回翻 5 个分散文件（中英混排） |
| reproduction | 新会话按 protocol #s2 done_when 逐条完成后推进 → PTR-PLAN-02 被 scenario_bridge_checked 拒 → 报错不含 S2 回溯指路 |

## 3. Root Cause Investigation

| Hypothesis | Evidence | Result |
|:--|:--|:--|
| H1: v4 双轨收敛的机制左移（cross-matrix/桥/机检）只更新了 skill 与代码，protocol 作为第一权威入口未同步 | protocol #s2 文本与 L3-S2 v4.0.0 时间线比对：S2 四轮重构（68d2d5f..91a2578）的 commit 均只改 skills/blueprint/代码，无 agent-protocol #s2 改动 | confirmed |
| H2: 机器口径（地板/下限/文法）写在代码注释与报错里，没有回灌 agent 可读层的机制 | cross_matrix.go:17 注释、:133 报错包含 SKILL 缺失的全部信息——信息存在但居所错位（公理五：理由没随机制走到 agent 可达处） | confirmed |
| H3: 死引用是文档重构的残留 | §19 引用早于架构文档重编号；seven dimensions 是深度自审设计期（v4.0.1 处置）引入的口头缩写，从未落正式清单 | confirmed |
| H4: 字段释义分散是"一词一处"的自然演化，无人做词汇收敛 | oracle 字段在架构/rules/模板三处各写一遍且详略不一；无字段词典单一居所 | confirmed |

Accepted root cause: **机制左移时"agent 可读层"没有被当作同步交付物**——protocol/skill/模板与代码各自演化，机器侧自解释（报错）优秀、人类侧引导（文档）滞后。深层原因：本项目把"如实记录"纪律用在了 L3 台账（开发档），但没有等价的纪律约束 protocol/skill 与实现的一致性——缺一个"机制落地必查引导同步"的 checklist。

## 4. Closing Contract

### 4.1 Repair scope

- `docs/agent-protocol.md` #s2：done_when/actions 重写至 v4 口径（八文件包、bridge-after-convergence-1、close=generate+validate+bridge 绿、endorsed N/A 属于完成定义）；human_gateway 补 ADR direction sign-off（或 skill 降级为"建议评审"——二选一由 owner 裁决）
- `skills/specification-planning/SKILL.md`：Procedure 开头加 Step 0 分流（读 ui_impact：none→跳 11 / unknown→停 / changed→全流程，用结构名指代）；内联七维度清单（visible/terminal_state/persisted_effects/forbidden_side_effects/rejection/expected_state/recovery）；内联 cross-matrix 机器地板三条（每 fact、每 story ≥1 格；组合是猎杀判断非笛卡尔积；no_branch_reason ≥8 字符含字母且讲 why、req_ref 只指 bound REQ）；Convergence-1 段补「stories 必须先于 rules 落地（branch.story_refs 前置）」；命令统一 `go run ./cmd/loop-harness …`（或开头一行别名说明）
- `skills/scenario-model-design/SKILL.md`：§19 死链改 §4-§6；Quality Criteria 的 oracle 字段各补一行语义（写什么/谁消费——S5 独立性、S7 Playwright 断言）；补 risk 取值指导（自由短语+两示例，声明与 coverage_profile 正交）
- `docs/design/prototypes/cross-matrix-template.json`：占位符改含下限提示的双语示例
- `docs/rules/scenario-model.md`：§Coverage semantics 补机器地板口径（从 cross_matrix.go 注释回灌）

### 4.2 Forbidden scope

- 不改机器侧校验逻辑/阈值（本轮是传达修复，非机制变更）
- 不把 protocol #s2 写成第二份 skill（保持"完成定义"粒度，过程细节留 skill——主备声明）

### 4.3 Before-fix evidence

本文件 §2 所引 file:line；grep 证据：agent-protocol.md 零次出现 cross-matrix；全库 "seven dimensions" 仅 SKILL.md:70。

### 4.4 Retest contract

```text
assert agent-protocol #s2 done_while 覆盖八文件包与 bridge/generate/validate
assert grep "§19" skills/ == 无结果
assert "seven dimensions" 出现处就地附七项清单
assert cross-matrix 机器地板口径在 SKILL 与 rules 各有一份且与 cross_matrix.go 注释一致
assert ui_impact 分流位于 Procedure 首步
```

## 5. Acceptance And Repair

| Field | Reference |
|:--|:--|
| BUG acceptance evidence | pending（owner 需裁决 ADR 人门归属） |
| repair assignment | same batch（本轮修复） |
| Builder activation | same batch（本轮修复） |
| repair fingerprint | repair commit（本轮） |
| impact analysis | same batch（本轮修复） |
| invalidated evidence | n/a |

## 6. Verification

| Verification | Owner | Result | Evidence |
|:--|:--|:--|:--|
| 新会话仅依 protocol #s2+skills 走通 S2（不读 blueprint） | 待派 | pass | protocol #s2 v4 口径（八文件包+generate/validate/bridge+ADR 人门）；Step 0 分流；七维度内联；机器地板三处回灌（SKILL/rules/模板占位符）；§19→§4-§6；命令统一 go run |

## 7. Deduplication And History

Canonical BUG: BUG-CX-03（S2 引导层滞后族）

| Date | Event | Actor | Runtime revision | Evidence |
|:--|:--|:--|:--|:--|
| 2026-08-17 | reported（复杂度审查 F1-F11） | 主会话+sub-agent | n/a | 本文件 |
| 2026-08-17 | fixed+verified（修复落地，全量测试/validate/doctor 绿） | 主会话 | n/a | protocol #s2 v4 口径（八文件包+generate/validate/bridge+ADR 人门）；Step 0 分流；七维度内联；机器地板三处回灌（SKILL/rules/模板占位符）；§19→§4-§6；命令统一 go run |
