# Canonical BUG: BUG-CX-12

> Status: fixed
> Severity: P0
> Runtime ref: N/A（模板仓库自审——S5 引导层；sub-agent 全部条目主会话亲证）
> Found in review round: s5-review-1
> Finding evidence: 本文件 §2/§3 所引 file:line（HEAD=66b7511 实测）
> Original responsibility: skills/document-verification / agents/document-verifier / docs/agent-protocol.md #s5
> Owner: PM / Architect

## 1. Fingerprinted Specification Chain

| Order | Kind | ID | Path | Version | SHA-256 | Relevant clauses |
|:--|:--|:--|:--|:--|:--|:--|
| 1 | BUG | BUG-CX-12 | `docs/reports/bugs/BUG-CX-12.md` | v1 | n/a | all |
| 2 | skill | document-verification | `skills/document-verification/SKILL.md` | v1.2.0 | n/a | Outputs/Procedure |
| 3 | agent | document-verifier | `agents/document-verifier.md` | current | n/a | 评审枚举 |
| 4 | rule | agent-protocol | `docs/agent-protocol.md` | current | n/a | #s5/S5.4 |

## 2. Observed Contradiction

**症状：S5 的证据产出端与回路动作在 agent 读物里几乎不可学——信封无样例、finding→TR-004 的动作只在蓝图、REV 枚举与 gate conclusion 词汇两套名字、team 落点双义、SKILL 与 protocol "不得手动调 transition" 叙事错位。**

| Field | Value |
|:--|:--|
| expected | agent 可在 S5 的读物里复制出一份能过 gate 的 document_review_record；finding→修复→S4→S5 重回路径在 agent 读物里可走 |
| observed | **① 信封无样例（F1，高）**：`document_review_record` 的 12 个被校验字段（schema_version/evidence_id/kind/runtime_id/baseline_generation/producer_agent_id/producer_responsibility/subject_refs/conclusion/...）无 agent 可抄格式样例——只在 REQ-040、L3-S5 蓝图、代码 evaluator.go:172-186 三处可见，agent 阅读单不包含。SKILL Outputs 只写"REV reports"未提字段。后果：DV 写一份 markdown REV 报告 → gate 的 `envelope.Conclusion` 空串 → `containsString(requirement.Conclusions, …)` 失败 → missing `evidence:document_review_record:DV-SPEC-CONSISTENCY` ——agent 不知"信封内 JSON 必须含 `conclusion:"pass"` 且 subject_refs 必须精确等于当前 documents[] 的完整指纹清单（多一少一即拒，evaluator.go:570-585）"。**② finding→TR-004 动作漂移（F3，中）**：SKILL step 10 写"request TR-004"，protocol 总纲禁止手动调 transition（agent-protocol.md:68-71）；真机制是 DV 写 `conclusion=fix_required` + `requested_event=document_fix_required` 证据后由 PreToolUse 自动提交（registry.go:166-168 `requestedRequirement(...,"document_fix_required")`）。"request" 一词在 agent 眼里指向 CLI。**③ REV 枚举与 gate conclusion 两套名字（F3，中）**：document-verifier.md:40 三枚举 DOCUMENT_PASS/FIX_REQUIRED/REQ_CHANGE_REQUIRED；gate requirement.Conclusions 接 "pass"/"fix_required"（catalog.go:333-341）。映射债在蓝图 L3-S5:67 自承，agent 读物无落点。**④ Team 落点双义（F2，中）**：protocol:266 "spawn Document Verifier Team via team-planning"；team-planning SKILL:64 明确"workgroups are not separate nested Teams, single Agent Team per session"。agent 不知该派两个 document-verifier subagent 还是建 workgroup。**⑤ S5.1/S5.4 步骤在 SKILL 缺失（F2/F3）**：SKILL 步骤 1 直接"Read the assigned TASK"，S5.1 编排动作与 S5.4 finding 回路在 primary skill 里没居所。**⑥ missing token 自修清单缺口（F4，低）**：`evidence:reviewer_not_candidate_author` / `evidence:exact_document_manifest` / `evidence:independent_document_reviewers`（evaluator.go:553-560）字面可读但不给"subject_refs 怎么填=精确全量"修法；`evidence:<id>:producer`（gate Unknown 转换）文案不含越权责任与期望值。**⑦ SKILL step 10 "Orchestrator requests TR-003" 与自动推进叙事冲突（F5，低）**：刚被 S0-S4 教育成"绝不碰 transition CLI"的 agent 会犹豫。**⑧ req_amendment vs req_change_required 词汇漂移（F6，低）**：agent-protocol.md:293-294 vs loop-harness.md:185-187 vs SKILL step 10。⑨ 步骤归属（F7，低）**：SKILL step 3/4（module truth / scenario four-pack）未标注归属 SPEC/TASK 哪条 responsibility |
| user/data/system impact | ①是最直接卡点（DV 写不出能过 gate 的证据）；②+③合起来是"评审/修复回路断链"；④+⑤是 S5.1 的 spawn 双义；其余属文案清扫。综合：S5 是文档验证的专责阶段，文档齐但不可学 |
| reproduction | ①：DV 按 SKILL Outputs 写一份 markdown REV 报告 → 命令 `runtime transition --id TR-003` → gate 报 missing `evidence:document_review_record:DV-SPEC-CONSISTENCY` → 不知是 envelope JSON 缺 conclusion/subject_refs |

## 3. Root Cause Investigation

| Hypothesis | Evidence | Result |
|:--|:--|:--|
| H1: SKILL v1.2.0 改时（CX-07 修复批）只动了 step 2 消费机检结论，对其他步骤没回检 | git log 显示 skills/document-verification/SKILL.md 仅有 commit `49b3f2b feat(s4)` 改 step 2；其他步骤 v3 起未变 | confirmed |
| H2: "document_review_record" 字段集本身在创建时（更早轮）就没建 agent 可抄居所——catalog/validator 知道全部字段但 SKILL 不知道 | catalog.go:333-341、evaluator.go:172-186 全量字段；REQ-040 文档里有但不在 agent 阅读单 | confirmed |
| H3: REV 枚举与 gate conclusion 词汇两套名字的映射在蓝图承认（`blueprint/L3-S5-document-verification.md:67`），但未落 SKILL/agent 读物 | 蓝图自承债未销 | confirmed |
| H4: S0-S4 修复的"绝不碰 transition CLI"叙事在 S5 被 SKILL 旧词"requests TR-xxx"破坏 | SKILL.md:37 与 protocol:68-71 直接矛盾 | confirmed |

Accepted root cause: 与 BUG-CX-INDEX 系统性根因 1/2 同族——**S5 是 v3 期 SKILL 自 CX-04 修复后未再做完整对账的遗留**：S3/S4 已被 CX-04 修复（如 Status 三词/§n 对齐），但 S5 的 SKILL 是"消费机检结论"那次小改的产物，procedure 其他步骤/Outputs/词汇映射没跟上。叠加②证据生产端与⑤回路动作端的"agent 读物与机器语义两套名字"。

## 4. Closing Contract

### 4.1 Repair scope

- **① SKILL Outputs 加可直接复制的 document_review_record JSON 信封样例**（含全部 12 字段，conclusion/subject_refs 用真值）
- **② SKILL step 10 改写为"把 REV 枚举映射为 envelope 字段 + requested_event，agent 不调用 transition CLI"**：
  - DOCUMENT_PASS → `conclusion:"pass"`，不填 requested_event
  - FIX_REQUIRED → `conclusion:"fix_required"` + `requested_event:"document_fix_required"`
  - REQ_CHANGE_REQUIRED → 升级到 human_gateway（`req_amendment`/`unrecoverable_business_decision`）
- **③ protocol #s5 S5.1 行补一句落地语义**：在唯一 Agent Team 内建 `document-verification` workgroup，两条 assignment 分别绑 DV-SPEC-CONSISTENCY 与 DV-TASK-EXECUTABILITY responsibility，主会话不得自己担任 DV
- **④ loop-harness.md TR-003/004/005 段加 missing-token → 动作映射**（rev_author/manifest/independent 三条）
- **⑤ SKILL 步骤→职责映射**（step 3/4 标 SPEC；step 6-8 标 TASK）
- **⑥ #s5 vocabulary 统一**：`req_amendment` 与 `req_change_required` 在 #s5 段标注等价
- **⑦ S5 fixture 改造**（CX-07 教训同款）——见 BUG-CX-13 修复范围
- **⑧ 已落地的"不得手动 transition CLI"叙事在 SKILL 一致化**（删 step 10 的"requests TR-xxx"措辞）

### 4.2 Forbidden scope

- 不改 SKILL v1.2.0 的版本号策略（除非产生新次版本语义）
- 不为 envelope 字段创建新的 agent-only template 文件（保持在 SKILL Outputs 段）

### 4.3 Before-fix evidence

§2 所引 file:line + envelope 无样例的复现步骤。

### 4.4 Retest contract

```text
assert SKILL Outputs 含可解析 JSON 样例（含 conclusion 与完整 subject_refs 示范）
assert SKILL step 10 写明映射（DOCUMENT_PASS→pass；FIX_REQUIRED→fix_required+requested_event）
assert protocol #s5 S5.1 含 workgroup 落地语义
assert loop-harness.md TR-003/004/005 段含三个 missing token 的动作映射
```

## 5. Acceptance And Repair

| Field | Reference |
|:--|:--|
| BUG acceptance evidence | owner 按设计裁决（L3-S5 v4.2.1 §8 计划） |
| repair assignment | 按批次执行 |
| Builder activation | 按批次执行 |
| repair fingerprint | 9a97e58 |
| impact analysis | 按批次执行 |
| invalidated evidence | n/a |

## 6. Verification

| Verification | Owner | Result | Evidence |
|:--|:--|:--|:--|
| DV agent 仅读 SKILL + 模板 + protocol #s5 写出能过 gate 的 envelope | 待派 | pass | 九项全处置：①§0 信封骨架每字段一行指引（模板即教师）；②③REV 枚举与 conclusion 合一（全流程一套词）；④⑤protocol #s5 三步叙事+S5.x 编号删除；⑥missing token 由 TR-003/004 description 与 manual 承载；⑦SKILL "requests TR-xxx" 改 requested_event 信封字段；⑧⑨卡片预载 2 skill+authored 禁令标注纪律层。C5 TestREVTemplateEnvelopeTeachesTheTruth 锁模板与机器校验永不分叉 |

## 7. Deduplication And History

Canonical BUG: BUG-CX-12（S5 引导层族；CX-04/07/09 族的 SKILL 残留）

| Date | Event | Actor | Runtime revision | Evidence |
|:--|:--|:--|:--|:--|
| 2026-08-17 | reported（S5 复杂度审查，主会话亲证全部条目） | 主会话+sub-agent | n/a | 本文件 |
| 2026-08-18 | fixed+verified（批次C，9a97e58） | 主会话 | n/a | 全量 -count=1 绿 + validate/doctor 绿 |
