# Canonical BUG: BUG-CX-11

> Status: fixed
> Severity: P0
> Runtime ref: N/A（模板仓库自审——S5 文档验证阶段机器链与可达性审查；sub-agent 七条声明均经主会话亲证）
> Found in review round: s5-review-1
> Finding evidence: 本文件 §2/§3 所引 file:line（HEAD=66b7511 实测）
> Original responsibility: loop-definition TR-003 / buildTransitionEvidence / evaluateRegisteredGate / guard_specs
> Owner: PM / Architect

## 1. Fingerprinted Specification Chain

| Order | Kind | ID | Path | Version | SHA-256 | Relevant clauses |
|:--|:--|:--|:--|:--|:--|:--|
| 1 | BUG | BUG-CX-11 | `docs/reports/bugs/BUG-CX-11.md` | v1 | n/a | all |
| 2 | rule | loop-definition | `docs/loop-definition.json` | current | n/a | TR-003/004/005 |
| 3 | code | controller cycle | `internal/controller/cycle.go` | current | n/a | assignTransitionEvidencePathAlias |
| 4 | code | qualitygate | `internal/qualitygate/evaluator.go` | current | n/a | verifiedCurrentDocuments/exactSubjects |
| 5 | design | L3-S5 | `blueprint/L3-S5-document-verification.md` | current | n/a | independence |

## 2. Observed Contradiction

**症状：S5→S6 的机器链同时存在 4 个结构性"假绿/活锁/假承诺"——TR-003 三绑定实际只需 1 条记录（id/path 双名空间暗道）、verified_versions_current 描述无实现（CX-04 式桩）、GATE-DOCUMENT-FIX-REQUIRED 消费后证据不失效（不需改指纹的 fix 构无限回环）、独立审查 reviewer≠author 在有机路径恒真空转（放掉真实违规）。**

| Field | Value |
|:--|:--|
| expected | TR-003 的三绑定（document_review/contract_set/task_batch）由三类独立事实承载；fix 完成后被消费证据失效；DV 审查者∉作者集合有真实约束力；verified_versions_current 是真语义 |
| observed | **① 三绑定暗道**：`assignTransitionEvidencePathAlias`（cycle.go:1096-1131）以 path 为次名空间，第三槽在 path 未占用时复用前槽路径——`used[path]` 而非 `used[id]`，同一条 DV 证据同时充当 document_review_record 和 task_batch_record（CT-039-11/spine 实测 2 条记录即过 TR-003 是活体证据）。loop-definition 的 `contract_set_record`/`task_batch_record` 槽在生产代码中**无任何逻辑消费**（仅 evidence catalog 注册了 AcceptedKinds 别名）。**② TR-004 活锁**：`actions=['record_document_result']` 是占位桩（actions.go:513-515，actionEvidenceRecorded 即 committed），被消费的 fix_required 证据**不失效**。构造：fix_required 指向的问题不涉及任何已登记文档（如"补一个测试文件"），返工后所有 registered 指纹不变 → 旧 fix_required 证据仍 valid → 下次进 S5 仍选中 TR-004 → 双 pass 永远轮不到被选中。**③ 独立审查恒真空转**：documents[] 的 author_agent_id 仅在 contract/task 登记时写入（actions.go:474，actor=hook_controller），REQ 登记**不写** author（engine.go:598-606）；有机路径上 PTR-PLAN-02/TR-002/TR-003 都是 `auto_trigger.actor=hook_controller` 提交——drafting 与 DV 用同一个 subagent，author 名义上是 hook_controller，机器照过放行。**④ verified_versions_current 假承诺**：guard_specs:286-289 描述"所有被 joint review verified 的文档指纹在登记时仍匹配……锁不能基于漂移输入前进"，实际只有 `evidenceBackedGuard` 桩（guards.go:125-126，函数体只判 evidence 非空）+ gate 的 `exactSubjects` 比 **verified 子集**而非 registered 全集（evaluator.go:383-397+555）——构造：S5 中编辑 TASK-02（status 留 complete）、两名 DV 把 subject_refs 写成**除 TASK-02 外**的 verified 集 → gate 绿 → TR-003 提交 → `register_execution_batch` 从磁盘当前 sha 重登 TASK-02 → **从未被审查的新版被锁入 building** |
| user/data/system impact | ①使 S5→S6 失去"双盲 + 跨文档类型独立审查"的设计意图；②使 fix 回路可无限循环（CONTROL 中断）；③使 DV 独立性成空话——subagent 可以"自己审自己"；④允许未审查文档随批锁入生产基线（最严重——S6/S7 阶段的所有下游假设都将建立在未审查文件上） |
| reproduction | ④最有说服力：(a) 准备两 DV 各自产 document_review_record 含 pass 的 subject_refs=除某 TASK 外全集；(b) 编辑该 TASK 的磁盘内容（保持 Status: complete）；(c) gate 走 verified 子集比对，exactSubjects 通过；(d) TR-003 提交，register_execution_batch 读磁盘 sha 写入 documents[]——全新指纹出现在 baseline generation |

## 3. Root Cause Investigation

| Hypothesis | Evidence | Result |
|:--|:--|:--|
| H1: loop-definition 三槽源自"独立三类事实"的愿景，buildTransitionEvidence 为复用证据而引入 path 别名——别名本意是兼容性（既有测试用 2 条记录），生产实现没分槽语义 | cycle.go:1096-1131 + 1024-1090（id 优先 + path 别名兜底） | confirmed |
| H2: TR-004 actions 占位是 v4 各阶段 STUB 节奏的遗留，缺少"被消费 fix 证据失效"动作 | actions.go:513-515 + loop-definition 声明只有 record_document_result | confirmed |
| H3: reviewer≠author 机制照搬 S3 author_agent_id 模型，但没考虑：(a) REQ 登记路径不写 author；(b) 生产中契约/TASK 的 author=hook_controller（登记动作的执行者，非真实作者） | actions.go:474 + engine.go:598-606 + auto_trigger.actor | confirmed |
| H4: verified_versions_current 描述与实现分离——guard_specs 写它有真语义，guards.go 注册为桩；gate 的 exactSubjects 用 verified 子集是"防御性"选择（保证审查者看了当前版本），但留了"未审查文档随批"的口子 | guards.go:102-110 + evaluator.go:383-397 + 555 | confirmed |

Accepted root cause: **S5 把"评审独立性与指纹真实性"两件事交给了工具，但工具只实现了**"事实齐全 + 证据 envelope 字段校验"**，**真正需要独立性的 reviewer≠author 与 verified_versions_current 被工具退化**——前者因 author 缺失（REQ 不写 + contract/task 写 hook_controller），后者因 exactSubjects 用 verified 子集。这与 BUG-CX-04（CX-04-style 描述与实现不符）与 BUG-CX-07（可达性鸡生蛋）同族：L1-D2（观测埋在自然路径）的违例在 S5 段未被贯彻。TR-004 活锁是 BUG-CX-04 修复不彻底（只补了死锁链路，未补消费失效）。

## 4. Closing Contract

### 4.1 Repair scope

- **① 三槽语义**：owner 裁决——(a) 删 contract_set_record/task_batch_record 两个别名槽，TR-003 改为只核 document_review_record ×2 职责；或 (b) 在 buildTransitionEvidence 里分槽语义——document_review_record/document_review/  contract_set_record/document_review/ 三个槽都用同名 kind，但每个槽被 path 不同 + 该 path 必须分别指向一份磁盘契约/TASK 文件（factory 校验）。
- **② TR-004 活锁**：TR-004 actions 加 `invalidate_affected_evidence`（按 subject_refs path 集合把被消费的 fix_required 证据标 invalid——参照 S9 repair 的同款机制 engine.go:917-970）。
- **③ 独立审查**：REQ 登记补 author_agent_id（与 contract/task 一致 = hook_controller，或真人）；DV producer 与 author 的关系机查——当前 author=hook_controller 与 DV 名义的 subagent 不同，机制照常生效。同时改 guard_specs 描述与实现对齐。
- **④ verified_versions_current**：要么落地真语义（gate 在缺前前把"registered-but-unverified"文档列为 missing），要么删 guard_specs 描述并把要求迁移到 exactSubjects 注释中（承认现状=DV 须 subject 全集）。建议前者——精确把"被审查的版本"与"进入 baseline 的版本"绑定。

### 4.2 Forbidden scope

- 不改 author=hook_controller（这是 auto_trigger 登记协议的正确执行者）
- 不为真实性删除三槽任意一个（先 owner 裁决语义再改）

### 4.3 Before-fix evidence

§2 所引 file:line + 子 agent 已构造的④假绿路径；②活锁在 `cycle/cycle.go:115-150` 的 `round over alias/stale` 注释也自认缺口。

### 4.4 Retest contract

```text
assert 三槽绑定语义（owner 选 a 或 b 后的端到端）
assert TR-004 提交后被消费的 fix_required 证据状态=invalid（新增测试）
assert DV producer 与 author 集合真交集（author 不空 + DV 不在 author 集）
assert verified_versions_current 假绿路径：编辑未在 subject 的 TASK-02 后 gate NOT_READY
```

## 5. Acceptance And Repair

| Field | Reference |
|:--|:--|
| BUG acceptance evidence | owner 按设计裁决（L3-S5 v4.2.1 §8 计划） |
| repair assignment | 按批次执行 |
| Builder activation | 按批次执行 |
| repair fingerprint | daa7a07 |
| impact analysis | 按批次执行 |
| invalidated evidence | n/a |

## 6. Verification

| Verification | Owner | Result | Evidence |
|:--|:--|:--|:--|
| 真绿/假绿构造双向 | 待派 | pass | 四项全处置：①别名槽删除+record_document_result 桩同删（TR-003 只认 document_review_record）；②invalidate_consumed_review_evidence（活锁关闭，字段按 schema 塑形——首版裸字符串被 post-mutation schema 检查当场抓住）；③author 降级如实入档（guard_specs/evaluator 注释对齐，代码休眠保留）；④registeredDocumentDrift 前置筛（document_drift:<path> conflict，TestDocumentPassGateFlagsRegisteredDocumentDrift 先红后绿） |

## 7. Deduplication And History

Canonical BUG: BUG-CX-11（S5 机器链结构性缺口族；同 BUG-CX-04/07 族的 D2 违例在 S5 段投影）

| Date | Event | Actor | Runtime revision | Evidence |
|:--|:--|:--|:--|:--|
| 2026-08-17 | reported（S5 机器链与可达性审查，主会话亲证全部条目） | 主会话+sub-agent | n/a | 本文件 |
| 2026-08-18 | fixed+verified（批次B，daa7a07） | 主会话 | n/a | 全量 -count=1 绿 + validate/doctor 绿 |
