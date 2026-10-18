# S9 机制审查与减负方案

## 0. 当前职责与可执行迁移清单（2026-10-18，本轮设计已固定）

本节是当前迁移规范，优先方向见 [GLOBAL §0](../GLOBAL-OPTIMIZATION-DESIGN.md)。本轮设计、实现与限定机械回归已完成；r79真实S9专业链及r80局部核查的具体范围见本节当前验证边界，不表示S7或全版本/全平台通过。**下文旧 §1–5 全部为历史，不是继续适用的规范；§6 也仅保留历次实施/失败记录及当时计划，不构成当前待办。** 历史中的“原消费者领域完整性”“原条件保留”、授权/发行/来源证明等兜底表述，不得恢复或限制本节的删除边界。

S9 完成真实修复、Main 集成检查、影响判断及独立复验，再自动交接 fresh S7。Hook 只接本人已保存专业答案、读当前已记录事实并提示准确下一步；普通 Pre/重复观察不发行、不重建 packet、不证明历史、不 prepare。接收限于必要专业结构、当前责任与本次 attempt 的定位，以及一次原 Writer 原子幂等记录。机器身份、路径、当前代码/变更事实由系统带出；不得以 exact 集合反向拒收作者。store 的 CAS/原子性是内部记录边界，不是在每层重放 proof 的许可。

| 工序与具体入口 | 谁做真实工作 | Hook 接收/只读职责 | 本轮明确删除边界 |
| --- | --- | --- | --- |
| 分工：`runSavedRepairDispatch`、`s9ObservedDispatchRegistered`、`stageRepairDispatchContext` | Main 阅读 Contract、依赖与实际责任，选择 Builder/独立 verifier，并真实平台启动/通知 | 接分工选择，给当前准确卷和参数齐全的原命令；按现有选择/责任读进度 | 删除普通工具身份/工具/范围 ACL；不靠 caller ID/alias 证明有权保存，不以同卷 script/MCP 无 file_path 为由漏接；不在保存/观察内 bind/prepare、调队列或物化 |
| Plan：`Service.SaveRepairPlanReportFrom` → `SubmitIssuedRepairPlanReport`、`BeginRepairExecution` | 每个 Builder 本人做必要前检并提出计划；没有执行条件如实 blocked，无改动责任可 `proposed_paths=[]` | 接 plan/red_checks/proposed_paths；本人计划记录后正常工作可开始，给 Main 原 prepare 命令 | 删除全员 Plan 屏障、forced red/非空路径、generic PLAN/Begin/activation 双交；删除 `AuthorizeSource`、来源 hash/journal/receipt/发行资格、packet Rebuild/issue-current-bootstrap 重验证。Begin 是工作事实，不是写权限；别人已 fixing 不阻本人晚到 Plan |
| workspace：`runRuntimeWorkspace` 的 bind/prepare、`publishS9Result` | Main 在普通工具执行实际 bind/prepare；准备实际 Worker、必要资料及原 Result 卷，真实启动作者 | 保存只提示原命令；Pre 只读；只有工作区和准确 Result 卷实际就绪才提示开工 | 删除 save Hook→bind/prepare/clone/Bootstrap/Result发行嵌套及被动队列准备；删除宿主工具/身份 ACL 与历史来源资格复放。保留 Git/工作区操作的真实成功或具体失败，不用状态 ready 冒充发卷成功 |
| 自检与 Result：`Service.SaveRepairResultFrom` → `AcceptIssuedRepairResultCandidate` / `CreateWorkspaceCandidate` | Builder 完成修复、必要自检与真实日志；有改动用普通 Git；无改动测试不造 diff/commit | 接当前责任/attempt 的专业结果并一次记录 candidate；实际专业缺项准确提示 | 删除第二 generic completion/交付报告、forced fail/blocked、新commit/diff形式门；删除 source/bootstrap/hash/journal/issue/receipt/packet rebuild 资格链、范围偏差 blanket 拒收及重复回执证明。偏差与风险如实交 Main 判断 |
| Main 集成：`runRuntimeWorkspace` integrate → `integration.Integrate` → `SubmitIssuedRepairResult` | Main 查看真实 diff/冲突，按目标和实际改动执行 `--check-command`，真实集成并记录 Final；无改动正常检查交付 | 读 candidate/Final 与检查事实，提示尚未完成的 Main 工作 | 删除 WritePaths scope ACL、generic 工厂/双定位、已接受 Plan/Result 的资格重放；不把无配置/未运行检查登记成 PASS。实际 Git冲突、命令失败或必要检查未做保持待办 |
| Impact：`prepareRepairImpactQuestionnaire`、`saveRepairImpactQuestionnaire` → `CreateChangeImpact` / `CommitChangeImpact` | Main 判断实际变更影响和证据处置；引用真实资料位置/输出说明理由 | 系统从已记录 Final 批次带出变更；接专业影响并记录，给独立复验责任 | 删除 `exactChangedArtifactSet` 准入、`validateCurrentRepairResults` 递归来源资格、`validateChangeImpactEvidenceLedger` / `validateImpactRecoveryEvidence` 的注册ID/generation/status/hash/URI/锚点证明。不能由机器依赖图否决专业 retain 判断，也不能让作者删系统带出的字段适应消费者 |
| Targeted：`prepareRepairTargetedQuestionnaire`、`saveRepairTargetedQuestionnaire` → `CreateTargetedReverification` / `CommitTargetedReverification` | **未参与实现和合并的独立人**真正复验全部必要 Contract 断言，亲自填写实际结果/证据；Main 真实分工/通知，不代答 | 系统给每项槽位及断言原文；接真实独立报告。**缺任何必要项的真实结果不推进**，具体提醒哪项未做；fail/blocked 如实记录 | 删除 `checkS9AuthorityFreshness` 全面基线证明、`evidence.ValidateRefs` 注册/hash/round门、continuity URI要求、`bindTargetedReverificationIdentities` 的alias/owner池/显式actor准入、注册义务证明；删除 `exactContractAssertionCoverage` 中ID拼写/严格数量/extra项拒收，保留实际逐项漏验提醒，不能因此忽略漏验或取消独立性 |
| Handoff/fresh S7：`completeRepairHandoff` → `CreateRepairHandoff` / `CommitRepairHandoff` | 已记录修复、Main必要检查、专业影响和独立必要复验真实完成后，系统自动汇总交接并生成准确 fresh S7 入口；不开始新S7 | 读尚缺的实际工序；无新增 Main Handoff 答卷，不要求手工 create/commit 或重复报告 | 删除 `checkS9AuthorityFreshness`、`checkHandoffRefs` / `validateHandoffAgainstCurrentRepair` 的递归hash/transitive/journal/receipt重验、refs exact-set、重复ledger/exact变更集、required-ID登记资格及为注册seed倒逼S9覆盖证明。缺实际独立复验仍待办；seed存在不等于新S7 PASS |

旧兼容元数据 `continuity_reason` 与 `scope_compliance` 可选；未填不拒收、不制造 scope fail，也不作为 Handoff 条件，不得为换取接收而编造 pass 或解释。全部必要断言的真实观察、结果、实际证据及未完成原因仍须如实记录，缺真实结果不推进。

**删除范围须在真正接收路径落实，不能搬到 consumer/helper/Handoff 再执行。** 专业证据可以是真实文件位置和输出，不要求 URI、Runtime 注册ID或hash链资格。保留对本次必要答案的可读性/结构与真实缺项提示；历史证明失败不得强迫回 S8、重发卷或重交同一专业内容。不得新增机制、状态、缓存、ID映射、receipt、补偿协议或复验矩阵。

<a id="s9-current-validation"></a>

### 当前验证边界（不增加计数）

**本轮 S9 减负设计、实现与限定回归已完成；r79 包的真实 S9 专业链已交接 fresh S7，r80 末片另有局部核查。** r86/r79 单次 CLI、无人工续接，exit0、非 external timeout，**1624.111264s（27.0685min）**；实际 Runtime **rev31、repair.closed、verification/running**。两本人 Plan、两 candidate/Final、Main 真实 Git 集成与两次10/10检查、专业 Impact、未参与实现/合并的独立者五项真实观察、自动 Handoff 及 round2 seed 均已记录。见[进程](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/s9-blueprint-landing-20261018/r86-native-main.process.json)、[实际报告](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/s9-blueprint-landing-20261018/r86-bike-r79-installed/.claude/evidence/s9-validation-report.md)、[独立终态审计](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/s9-blueprint-landing-20261018/r86-terminal-audit.md)与[收口证据](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/s9-blueprint-landing-20261018/s9-closure-summary.md)。**fresh S7 仅入口，Delivery/QA仍pending、clean_round=null，不是S7 PASS；上游S7/S8及审批是明确synthetic fixture，不冒充真人履职或授权。**

r85/r78也完成同一S9专业链，但四段自然结束、三次人工续接，累计 **1706.755652s**；原独立作者纠正旧元数据后才接收，见[终态审计](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/s9-blueprint-landing-20261018/r85-terminal-audit.md)。r86仍触发旧continuity提示，由独立作者本人补后接收；不是零机制困惑，也未验证后续metadata缺省行为。两轮流程/续接不同，耗时不能作控制性能对比。r79反馈片通过[两原例机械验证](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/s9-blueprint-landing-20261018/validation-r10-summary.json)；r80旧Targeted元数据可选片通过[原lead全尾机械例](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/s9-blueprint-landing-20261018/validation-r11-summary.json)及[独立Claude局部安装/原例/代码核查](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/s9-blueprint-landing-20261018/r80-claude-audit.md)（[Root复核](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/s9-blueprint-landing-20261018/r80-audit-root-review.md)）：安装1.40132s、原例PASS2.16s/wall2.21612s。**r80没有另跑完整原生S9，不把r79模型证据继承为r80全链PASS**，也不新增同版全矩阵资格。所有包仅linux/amd64评估；历史失败与未知性能边界保留。

## 历史记录入口

以下 §1–5 为2026-10-17初始静态审查，§6为历次实施日志。其行号、“当前”“待实施”及当时后续计划均只描述历史；与本节冲突的旧门槛不再适用，不得按旧计划恢复资格证明。

## 1. 历史：范围与最小正常工序

S9 消费已批准 RepairContract，不重新猜根因；目标是完整修复、重建可信度，再回 fresh S7。
Hook 监听、提醒和流转，不对普通工具按身份、权限或范围拦截；原消费者负责领域完整性。
专业作者只填本人专业答案，系统带出已知身份、revision、hash、来源、实际 diff 和检查事实。

| 原工序 | 专业责任 | 正常交付与流转 |
| --- | --- | --- |
| Contract readback 与分工 | Main 阅读修复意图、单元依赖、完整范围和断言，选择 Builder 及独立复验分工 | 复用原 Session/Plan/dispatch；系统给资料与当前题目，Main 完成真实平台启动 |
| 实施前计划与前检 | Builder 提出有限计划、拟改路径，执行真实 red/blocked 前检 | 在准确 PlanReport answer 路径保存；原消费者投影 lifecycle checkpoint，按原条件进入执行 |
| 完整修复与结果 | Builder 修复根因、补 detection gap，处理兼容/迁移/回滚并运行必要检查 | 保存准确 RepairResult；系统消费原候选和 completion，Main 真实集成并运行 Main checks |
| 实际影响对账 | Main/Repair Lead 判断依赖影响及旧证据处置 | 保存专业 ChangeImpact；系统从 Session baseline/真实 diff 带出机器字段，原消费者提交 |
| 独立定向复验 | 未参与对应实现和合并的 verifier 按 Contract 执行全部必要断言 | 保存准确 TargetedReverification；原消费者校验独立性、覆盖与证据，失败按原路恢复 |
| fresh S7 交接 | 本步无新增专业答卷；Main 按交接结果继续新 S7 责任 | 原条件满足后系统调用 Handoff，生成原 seed；不手工 create/commit 或重复填 Handoff |

PlanReport 与 RepairResult 是执行前、执行后两次不同的必要专业工序，不是重复报告。
每次保存不再另交同内容 generic PLAN/completion，也不手动 issue/consume 或抄机器信封。
显式 `plan_approval_required`、真实授权、人闸、独立性、原 Writer/CAS 和幂等恢复均保留。
工作五态仍为 `pending/active/blocked/done/cancelled`；原 phase、unit 状态及领域结果不强压成五态。

## 2. 历史：已存在的原消费者接线与证据边界

- `internal/submission/repair_plan_report.go:146` 调用原 `SubmitIssuedRepairPlanReport`；`:184` 已说明同一专业答案接入 generic PLAN/activation，无须第二份计划。
- `internal/submission/repair_result.go:129` 将 non-pass 交原 blocker；`:132` 调用原候选接收，并于 `:139–153` 用 `EnsureCompletion` 投影原 completion，实际没有第二份专业完成报告门。
- `internal/cli/repair_lead_questionnaire.go:886–890` 保存 Impact 后调用原 Create/Commit；`:983–990` 保存 Targeted 后调用原 Create/Commit；`:747` 已接原 Handoff Commit。
- Main resolution、Targeted blocked/resume 及历史证据都有原路径；本轮应缩减重复准备与反馈缺口，不另建恢复协议。
- `internal/cli/repair_command.go:1031` 的“先 genuine completion_report 再准备卷”是旧提示，不能据此误称现有 Result 保存仍需作者双交。
- 预执行 PlanReport 使用 Main authority root 是原设计，不把它当作执行根错误；实际实现、Git 和候选来源仍按原责任消费。

R56 真实 S8 收尾只到 S9 planning 入口：原 Plan 有两个 pending Assignment，
`assignment_owners` 为空，尚无 Builder 问卷；没有执行 S9 模型专业工作。
证据见 [R56 最终报告](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r56-s8-closeout-final.json)。
其中 synthetic 上游、仿真人决策及宿主范围限制仍适用，不能当作 S9 已验证链。

## 3. 历史：六类确定缺口与最小收缩方向

以下为当前源码静态分支证据；没有实测性能、超时或真实 S9 专业完成结论。

### 3.1 Main 正常 planning/reproducing 没有专业分派入口

`internal/cli/repair_lead_questionnaire.go:226–265` 选择后段与 readback，不处理 planning/reproducing 分工。
`internal/cli/stage_answer.go:1393` 将正常未派发也提示为“restore RepairPlan/dispatch and PLAN checkpoint”。
Main 因而仍需猜当前该派谁、如何关联修复与复验责任；R56 空 owner 只是入口事实，不是失败的 S9 调查。
最小修正：沿原 Plan/Assignment 展示当前专业分工题与已知资料，保存调用原 dispatch；
真实平台 spawn 仍由 Main 完成，Hook 不代做，不另建主协议或发行回执。

### 3.2 普通监听仍走完整 Controller/恢复，Stop 提醒重复履职证明

`internal/cli/controller.go:70–72` 只让 investigation 走轻观察，S9 被排除；
`internal/cli/run.go:3047` 因而进入 `runControlCycleForHook`。
`internal/cli/stage_answer.go:773–782` 在 SessionStart/PreCompact 使用通用 New/Recover。
`internal/cli/run.go:3010` 先生成 Stop guidance，`:3140` 再执行原 StopIdle，当前反馈路径会重复读取履职证明。
最小修正：提醒读取当前原事实和准确路径；必要原转移仍由原消费者触发。
同次 Stop/Idle 完整证明只求值一次，保留原完成条件，不把校验移成逐工具权限门。

### 3.3 每次 Main Pre 全量准备，并重复校验后段来源

`internal/cli/stage_answer.go:1377–1378` 每次 Main Pre 调用全量 IssuePlan；
`internal/cli/repair_result_questionnaire.go:28–29` 同样全量 IssueResult，`:63` 逐卷读取领域 receipt。
`internal/cli/repair_lead_questionnaire.go:421–425` 提醒准备 Impact 时验证 Session 并重算 Git；
Targeted 准备又经 `:482`、`:600–605` 验证/读取原来源，非只展示现有题目。
最小修正：只有缺卷或真实来源变化才做必要原准备；普通提醒读已登记事实，不重跑完整链。
保存、原 CAS 及收尾所需的真实复验继续保留；不新增准备缓存或旁路事实库。

### 3.4 已接收答卷在普通工具观察中重送，resolution 发现过重

`internal/cli/questionnaire_observer.go:95–99` 仅为 S8 跳过未变 accepted 内容；
S9 已填写且偏离初始 draft 的答案可被普通 Bash/MCP 当作保存重送原消费者。
同文件 `:267` 为发现 Main resolution 调用完整 Draft；
`internal/repair/issued_repair_result_resolution.go:80` 在 Draft 内复验当前 blocker。
最小修正：用原接收事实跳过未变内容，真实异答仍交原冲突判定；
已知 resolution 路径用既有纯 `RepairResultResolutionAuthoringPath`（同文件 `:55`），不增缓存。

### 3.5 后段题形与恢复准确路径不足，角色口径也未同步

Impact 草稿 `internal/cli/repair_lead_questionnaire.go:438–443` 给空 decisions，
但 `internal/repair/model.go:325–337` 要多字段关系/处置对象，没有相邻合法实例可直接照填。
Targeted `repair_lead_questionnaire.go:615–623` 只生成 assertion ID，不呈现断言原文、合法枚举和结果示例。
`internal/submission/repair_result_blocker.go:64–68` 首次接收带 Main resolution 路径；
后续 receipt `:85–90` 只说恢复/查 status，丢失当前准确填写路径。
最小修正：在既有题形内展示原字段、合法枚举、实际 ID 与断言原文，每次恢复反馈保留准确路径；机器绑定自动带出。

同步角色/Skill 短段，不扩展机制：Builder 角色混 Main dispatch CLI，FE/Test 还写 backend 默认值
（`agents/frontend-builder.md:59–62`、`agents/test-builder.md:61–64`）；
三 Builder 的“必须 failing red”（如 `agents/backend-builder.md:71`）与 L3-S9 `:426` 的 red/blocked 冲突；
`skills/bug-resolution/SKILL.md:211–212` 未发卷即退旧 delivery，应限显式 legacy 模式，正常缺卷归系统准备；
原 finder 角色只规定 S7 答卷（`agents/qa.md:32,48`、`agents/e2e-tester.md:30,62`、`agents/delivery-verifier.md:25,42`），需补 S9 独立 Targeted 分支；
`docs/control/agent-protocol.md:690` 的 tools/command-class activation 限制应收为专业范围/停止条件，不能延续工具 ACL。

### 3.6 已登记后直接回“重放成功”，不比较新保存的专业内容

`internal/cli/repair_lead_questionnaire.go:802–803` 有 impact_ref 就直接返回已登记；
`:815–816` 有 Targeted PASS 也直接返回，未读取这次答案，异答可能被当同答。
最小修正：复用原不可变对象与幂等规则比较专业内容；同答只读、异答明确原冲突/恢复。
不为比较新增 receipt、store、状态或覆盖第一次接收的答案。

## 4. 历史：兼容范围与必须保留的专业边界

`repair_lead_questionnaire.go:489,514,520` 限单义务/单 verifier，原 `internal/repair/runtime.go:1308–1319`
已允许在 ready_for_full_review 补齐剩余原义务；这不是本轮扩展多 Contract batch 的理由。
单 Contract 多 Builder 可以只派一名独立 verifier 复验全 Contract，不能直接说这种拓扑失败。
原事实有明确关系时保留原消费；真正无映射时明确 Main 专业分工缺口，不新建映射协议。
真实 Git、Main 集成/task-integrate、required checks、完整断言、独立性和原授权/CAS 均须保留。
历史 CLI/packet 测试和迁移快照是兼容资料，不当作普通作者步骤或当前真实模型通过证据。

## 5. 历史：最小返工切片与验证范围

| 顺序 | 最小切片 | 验证重点 |
| --- | --- | --- |
| A | 先固定 L3 顶部最小契约、角色与完整题形 | 作者知道本人做什么、交什么、填哪里；Plan/Result 各一次，机器信息不重抄 |
| B | 补 Main 入口专业分工，接原 dispatch 与真实平台启动 | 正常空 owner 不误报恢复，真实 Builder 本人获准确卷，不由 Main 代写 |
| C | 轻监听、被动去重与后段原消费者反馈 | 普通工具不全链；准确恢复路径；未变已接收不重送，异答仍冲突 |
| D | 窄回归后全新装包 Claude 执行实际 S9 | 完整修复、必要 blocked 恢复、独立 Targeted、原 Handoff 至 fresh S7，不借 S8 PASS |

关键场景限于：冷入口分工/派发；真实 Plan 前检→Result；Main 集成与 checks；
Impact 合法题形；独立 Targeted 全断言；必要 blocked/resume；普通工具/重复保存/异答；
准确恢复与原 CAS/Handoff 至 fresh S7。先窄测试再真实宿主，不扩成全平台或全分支矩阵。
以上均为待实施/待验证范围；本轮没有运行这些测试或模型，也未宣布 S9 返工完成。


<a id="6-当前删减范围与证据边界2026-10-17真实完整链待验"></a>

## 6. 历史实施与验证记录（非当前规范，含当时计划）

来源统一小片已冻结并通过限定机械验证：r82 原 lead 完整 case PASS（test 4.10s / 包 4.152s / wall 55.536s），原 last-wins 与 stable Changeset identity 两例 PASS（包 0.066s / wall 55.795s），见[原始日志目录](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/r82-impact-typed-source-s7_oq7wa)。case 使用明确 synthetic 专业数据，不是 native PASS。after-r75 host-only 构建 exit 0、55.109531s（[清单](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/after-r75-source-inventory.json)）；本记录晚于锁定快照。r81 已有真实前半两 Final；下一次在 fresh 同 bike 用真实代码、tests、Git 和原 consumer 建立明确标注合成作者的机械前缀至 Impact，仅由 Claude 验证修复尾链，不冒称单轮全新 native 前半闭环。

r81 最新真实运行已停止：两份 Final 均通过，暂计两段 **901.014s + 554.553s = 1455.567s**，Runtime **revision 28**；Main Impact 的系统字段带出两个未 tracked 的 pyc，但 consumer 又按 Final 批次拒收，形成机制矛盾。Impact artifact 已创建但未 commit，**0 Targeted / 未到 fresh S7**，见[最终报告](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r81-bike-fixed-r74-main-final.json)与[耗时分析](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r81-bottlenecks-and-timing.md)，不标闭环通过。该来源边界已实施，限定机械验证及真实尾链待验范围见上文：

S9 的 changed_artifacts 与自动 Handoff Changeset 统一取已接收 Final Result 批次记录的实际 changed artifacts，沿原 last-wins 规则归并；不重新 WalkDir 物理 Session 目录或缓存求第二套来源证明。系统带出已知变更集，Main 只填写专业影响判断，不要求作者删除机器字段；consumer 不以物理 diff 重做接收准入，不增加 pyc 黑名单、状态、格式或 API，真实测试、Main 集成及独立 Targeted 保留。

r81 限定验证已通过：五个既有例全 PASS（包 **13.226s** / wall **37.070s**，[原始日志](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/r81-hook-prepare-split-e9p3j6ki/test.log)）。保存到 workspace 的嵌套准备及 observer 队列遍历已删除；保存/监听给准确命令，Main 普通工具执行原 prepare，原 prepare 负责实际工作区和 Result 卷。覆盖包括 typed normal observer 的显式启动发卷及两个明确 legacy workspace-only 例，不等于完整 S9/fresh S7 通过；后者待全新 r81 真实运行，也未认定 r74 晚 Result 超时根因已解决。after-r74 host-only 构建 exit 0、24.247160s（[清单](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/after-r74-source-inventory.json)）；本工作树验收记录晚于该锁定快照，r80 失败证据保留。

r80 / after-r73 真实运行已在 **331.010s** 停止：首次 Edit 选择出现原 8s Hook timeout，ready 已记录到 **revision 18 / 0 Result**，Result 卷未发；见[最终报告](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r80-bike-fixed-r73-main-final.json)与[耗时分析](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r80-bottlenecks-and-timing.md)。只读调用链显示保存 Hook 嵌套 workspace bind/prepare、Bootstrap 物化与 Result 发行，违背 §0；这不算实际 startup 完成。该失败后的最小修正边界如下，当前实现与限定验证见上文：

S9 分工/工作位置保存只记录选择并反馈现有原 CLI 的准确 workspace bind/prepare 命令，由 Main 在普通工具中执行正常 startup；保存 Hook 不嵌套创建工作区、物化 Bootstrap 或发行 Result。原 prepare 自带 Result 发卷和准确路径；已有 ready execution 但缺 Result 卷时，仍由 Main 运行同一原 prepare 完成 startup，不手工 issue。只有实际工作区与 Result 卷均就绪才可按该位置开工，保存成功不等于 ready。Pre 和重复监听只读；不增加后台、缓存、状态、预算或第三份问卷。

最新 r79 / after-r72 真实 bike 运行在 **662.697s** 停止，**0 Result，未闭环**：脚本保存已填写的 `check_location` 未被 observer 接到，作者用 Edit 重存后才 prepare，这是实际机制失败；见[耗时与阻塞分析](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r79-bottlenecks-and-timing.md)及[真实最终报告](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r79-bike-fixed-r72-main-final.json)。observer 修复复用既有 Plan、Result、execution 和原队列纯函数；三个原例最新 PASS（包 12.504s / wall 237.680s），仅覆盖首次脚本 prepare、依赖 queued 及未变重放零写，见[原始日志](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/r80-observer-real-plan-awbs2p7c/test.log)。两个首轮 fixture 失败及删除不合法 legacy 后半段的覆盖边界见[记录](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r80-observer-fixture-boundary-adjustment.md)；真实 Final 后释放和完整 S9 在 r80 仍未验证，不宣称原超时已解决。after-r73 已完成 host-only 构建（201.801s，[源码清单](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/after-r73-source-inventory.json)）；本次工作树文档晚于该锁定快照，仅补验收记录。下文保留此前分析和验证历史。

2026-10-18 分析补充：r75 只读 status 基线 **1.624s**（[分析](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r75-readonly-timing-analysis.md)）；r76 在首责任 Final 后，第二个无改动责任经原生 observer 保存 **2.616s**（[分析](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r76-two-responsibility-save-analysis.md)）；r77 同 case CPU 样本 **2.60s**、profile 窗口 **3.03s**，ValidateBytes 累计 **38.08%**、RepairResultPackets **29.23%**、Compile **18.46%**，Compile 已含在 schema 调用内，禁止相加（[分析](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r77-save-cpu-analysis.md)）。这些样本未确定或修复原 r74 超时根因；生产和 r74 现场未改，S9 仍缺 Impact、独立 Targeted 和 fresh S7。

用户已授权实施以下两个既定 §0 小片；冻结源码的三个原定向机械测试通过（runtime 包 0.860s / repair 0.075s / integration 20.948s）。同新源外部单次诊断在首责任 Final 后，以第二个无改动责任、Bash observer 和原生 8s 入口保存通过：2.216615717s、accepted=true，一次 revision 23→24，HEAD 不变、无新 commit；见[原始验证报告](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r78-frozen-mechanical-and-native-observer.md)。这不是严格 before/after 收益证明，不代表 r74 根因已解决或真实模型 S9 PASS；该轮 after-r72 后续真实验证结果见上文 r79。不增加设计、状态、缓存或预算，保留专业必填、当前责任关联、一次原子幂等记录和真实 Main 检查：

- 普通 RepairResult reader 移除逐包 schema 与 Rebuild；Find 只选唯一目标登记行并解码该卷。
- 候选 Accept 删除重复 Issue/bootstrap/权限回调准入及写后 Snapshot/Verified，直接使用原 Writer 已返回的结果反馈；以下旧证据继续作为各自历史边界。

最新 after-r71 的 r74 bike 真实运行已停止：两段累计 **1222.613s**，均 exit 143、非外部 timeout；原轮因 timer 空闲停止，cont1 出现真实 Hook 超时。已接收 **2 Plan（test 为 `[]`）**，backend 的真实 candidate、Final、Main Git 集成和检查通过；test candidate 最终在 **revision 26** 存在，但没有成功接收回执。首个超时早于该 candidate journal 时间，不能据此确定后置投影就是根因。**Impact 0 / Targeted 0 / 未到 fresh S7**，不标 S9 完成；见[最新真实报告](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r74-bike-fixed-r71-main-final.json)。以下保留此前包与工作树验证历史。

after-r70 的真实 bike 运行 r72 已到 **2 Plan / 1 candidate**；Main 已实际 merge 并执行成功的 unittest 检查，但最终 Result 接收因相对/绝对路径混算被拒，先报 checkpoint path escaped，绝对 root 重试后报 Main check record escaped。不能将其写成“未发生合并”，也未完成 S9、独立 Targeted 或 fresh S7。见[真实运行报告](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r72-bike-fixed-r70-main-final.json)。

同包 settings 运行 r73 已将候选写入 Runtime **revision 23**，随后保存回执超时；候选写入不等于最终 Result 或闭环通过，见[原始阻塞报告](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r73-cont2-main-blocker.json)。

工作树已精简 Plan/Result 作者 view 为现有 Draft 专业问卷及准确 Main 材料路径，修正 Main 相对路径读取，并删除候选成功后的重复投影。两原定向机械测试通过：**integration 25.034s / attempt projection 0.086s**；首次 integration 因新增断言将合法材料目录当文件读取而失败，仅修测试断言后原例重跑通过，原失败与 attempt 结果见[首次报告](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r74-final-original-tests-20261017T122709154602Z.json)，最终 integration 见[重跑报告](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r74-integration-directory-assertion-20261017T122836175999Z.json)。这是工作树机械验证；已安装 after-r70 和真实现场未替换，after-r71 正在建包，尚无真实闭环 PASS，不标 S9 完成。

以下为此前入口实测与机械验证历史，不能替代上述最新同源验收边界：

历史 after-r69 已运行并停止：两名作者首交 **2 Plan（含真实 `[]`）/2 ready**；backend 使用普通 Git 提交并做真实 self-check，test 如实报告 blocked。两个专业 Result 已写入，但 PostToolUse 出现 internal budget timeout，停止时 **0 candidate/0 final Result**；运行 486.672s，exit 143。见[原始阻塞报告](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r69-s9-reduction-real-main-blocker.json)。没有完成真实 S9、Targeted 或 fresh S7，尚未证实超时的唯一根因。

r70 独立 Claude 自建离线 bike seed，约 94s，真实测试 **10 项中 9 pass/1 fail**，Git clean。r67/r69 双 clone 使用同 payload、同 prompt，Main 默认预算 900s；实际两组均约 210s 因分工 Post 超时停止：各 **owners 2/backend 问卷 1/test 卷未发/专业 Plan 0/native 0/Result 0**。见[双组原始对比报告](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r70-bike-paired-real-comparison-final.json)。这只定位共同入口阻塞，不能宣布新旧优劣。

本轮已删除 blocked 保存与准确路径定位的重复 source/proof，以及 Result Hook 的 generic 工厂和双重定位；保留当前责任、真实 Git、专业答案和原 Writer 记录。两原合源机械单例曾 **PASS：blocked 49.80s/integration 46.21s/包 96.280s**，见[原始日志](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/logs/r71-final-ge_l7n04.log)；该运行早于最新 Result frontend 小片，不能称作当前最终同源验收。其后已删除 dispatch generic 工厂与 Plan issuer 重复完整 proof，并构建 after-r70；后续真实结果及未测修正以上文为准。

以下保留前轮失败与机械验证历史，陈述仅对应各自当时源码：

after-r68 真实运行 329.364s 后由 Root 停止（exit 143）：诚实 test Plan 的 `proposed_paths: []` 被 receipt `submitted_answer` 残留 `minItems: 1` 拒收，作者被迫改填 owned scope 重交。停止时为 **2 Plan / 0 ready / 0 Result**；这不算正常 Plan 首交通过，也未通过 S9/fresh S7，见[真实阻塞记录](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r68-s9-reduction-real-main-blocker.json)。r68 原公开 integration 单例机械 **PASS，24.982s**，见[集成证据](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r68-public-integration-evidence.json)；该结果不替代真实运行验收。

本片已删除 receipt 的拟改路径非空约束；原公开 CLI Hook 单例由真实派发 test-builder 首交 `[]`，确认原 report 与 receipt 都保持空数组，**PASS，15.270s**，见[原始日志](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/logs/r69-nochange-plan-public-hook-20261017T112024887503Z.log)。同期源码已删除 candidate/final 强制至少 fail/blocked 及 final 对范围偏差一概拒收的门槛，保留真实结果和必要 Main checks；当时这部分 domain/public 复测尚待，after-r69 尚未装包实测；后续当前结果以上文为准。无新设计、协议或字段，不标全阶段 PASS。

历史 after-r66 为 **built_not_run**。after-r67 真实 Claude 到两名本人 PlanReport、两个 ready Worker，RepairResult 接收数 **0**；普通 `commit_command` 仍要求旧提交表单，另有专业 answer 被外层 UI 的 `example.expected`/题干相等门槛拒收。Root 在 529.291s 主动停止（exit 143，非超时），未通过 S9/fresh S7，见[真实阻塞记录](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r67-s9-reduction-real-main-blocker.json)。

本轮源码已删除正常 S9 commit 表单指引，Plan/Result 直接提取 `answer` 并校验原专业字段，忽略外层展示字段；无新协议或字段。三个 validator 直接测试 PASS（0.327s），launch 定向 PASS（3.555s），CLI 编译通过（0.039s）；其后的 r68 结果与当前待验范围见上文，不标全阶段 PASS。

历史 after-r64 已构建，冻结 source 关键机械选择 **9/9 PASS**，见
[机械证据](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/after-r64-s9-reduction-mechanical-20261017.json)。
真实模型已到两名 native 作者本人 PlanReport 接收、Begin、两个 ready worker，**RepairResult 接收数为 0**。
过程中出现 legacy coverage 噪音、并发 staging 清理误拒收（同轮随后本人 Edit 已 accepted），
以及 Pre 准备 workspace 的锁 deadline。用户因此要求停手并重新梳理，模型已停止。
[模型原始日志](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/logs/r64-s9-reduction-main-normal.stream.jsonl)
保留实际顺序；未完成真实 S9/fresh S7 验收，不能标 done。

本轮已在源码落限定删减；r65 真实 Main 只读运行 24.89s 后，Root 因已知同版 bug 主动停止，不算 Claude 复现或 PASS。after-r66 仅 built_not_run；after-r67 的实际失败与当前待验范围以上文为准，尚无本轮真实 S9/fresh S7 通过。当前契约见
[L3-S9 当前最小工作流](../l3-stages/L3-S9-repair.md#当前最小工作流2026-10-17重新固定减负设计)
与 [GLOBAL §0](../GLOBAL-OPTIMIZATION-DESIGN.md)：

本轮已落的保存/流转边界：专业 PlanReport 接收全 pass 前检及拟改范围偏差，删除正常保存前后的重复来源认证；S9 lead/verifier 普通 context 只读，不准备 Session/Plan、Impact、Targeted 或 Handoff。真实 Contract 批准后准备原 Session/Plan 和 Main 分工草稿；历史显式 plan compile 成功后只读现 Plan 准备相同草稿。真实 Main integration 收尾刷新原 snapshot，仅在 impact_reconciliation 准备 Impact；其余状态反馈实际剩余责任。Impact 保存成功给已派 verifier 准备 Targeted；targeted 阶段的 verifier 分工保存只准备该本人卷；最后 Targeted 成功后调用原 Handoff 并报告实际 fresh S7 seed。lead 路径沿同 Session/source 保持稳定，删除多余 review round 目录。

限定机械证据：Plan 专业字段/原 domain 前检两例通过（workpacket 0.250s、repair 0.864s，submission 编译通过）；真实批准入口与只读提醒两例通过（CLI 3.627s）；原 Impact→独立 Targeted→Handoff 单例通过（CLI 9.492s，含原同答零写及既有 blocked/resume 段，未扩大矩阵）。显式 session open→plan compile→两原分工草稿单例通过（CLI 2.108s）。早期同组公开 integration 单例因三处旧 WorkspaceAuthority.Checks 测试引用编译失败（整组 exit 1）；之后保留的中间失败依次暴露重复 Begin、prepare 嵌套锁、空 profile 必填、旧 prepare 输出解码/来源拒收断言、空 receipt 列表序列化、旧 fixture 重复 issue 超时，以及 launch 动态 session 使原卷失效。该单例各次原始 external 日志已持久归档，见[工作树证据清单](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r66-s9-public-integration-worktree-evidence.json)，没有覆盖中间失败记录。

前一工作树覆盖（不代表 r67 后新源）：原 `TestS9RepairResultUsesOriginalWorkerCandidateAndMainIntegrationChain` 单例 **PASS，46.840s，exit 0**，见[最终日志](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/logs/r66-s9-public-integration-worktree-20261017T104904042563Z-stable-definition.log)。它直接消费真实 prepare 已给的同一卷，跨 launch 保存个人 Result，Main 通过 `--check-command` 实际执行 `test -f internal/api/payload.go`，走原 Result/final/Main ACK/cleanup 及既有回执重放；不再手工重发卷。此结果不是 r65 冻结源码同版验证，也不代表真实 S9 整阶段、fresh S7 模型或全平台通过。

- 原 issuer 沿用已有 operation ctx，避免同一准备动作重取锁；无配置时不强制 profile，原投影把缺失 receipt 列表记为空数组，不生成 receipt/profile/PASS。
- 动态 platform session 不再进入稳定执行定义而使原卷失效；无改动计划的 `proposed_paths: []` 按本人真实答案保留。

以下仍是本轮设计边界；未由以上限定证据覆盖的分支不据此宣称完成：

- 目的在于避免跳步遗漏，让作者知道做什么、填什么、准确路径和还差什么；Plan/Result 仅记录本人实施前后工作，不作泛授权或完成证明，不新增全员先交 Plan 的屏障。
- Main/系统在责任开始前准备真实位置与当前卷，阶段变更在原位置提供下一卷。Hook 只读已记录事实提醒；真实保存记录答案一次并推进。普通工具不重跑业务、重复准备、全扫描或索取来源 hash/receipt/coverage 证明；Hook 不做调度、测试 runner 或工作区恢复。
- 复用现有保存监听并支持编辑和脚本；并发和重保存内部处理，不让作者理解 pending/staging 或重交已接收答案，不新增协议、缓存或 payload 解析器。
- 问卷只问专业判断、真实结果、证据位置和未完成原因；已知任务、ID、diff、命令事实系统带出。JSON 可继续且题形明确，本轮不新定字段/schema，不因旧 schema 字段多要求全抄。
- 前检 command 是已做工作记录，不冻结为 Main 后续命令；Main 按目标与实际改动确定并真实执行必要检查。无改动测试有效，无需造 diff/commit；无需迁移无需双 refs；前检如实 pass/fail/blocked，不强制造 red。
- 删除身份/权限/scope ACL、泛 PLAN→AgentBegin→activation→work_started→completion、source hash/receipt/coverage 额外准入，不移到 consumer 保留。接收只保最小可读结构、明确必填专业项/真实遗漏反馈、必要一致性和一次正确记录；原子性/幂等内部处理。工作五态仅 pending/active/blocked/done/cancelled。
- 真实工序保留 Main 阅读 S8 已批准 Contract 并明确目标/断言/分工与真实位置、本人 PlanReport、修复或测试及自检/RepairResult、有改动 Main 集成与必要检查、Main ChangeImpact、未参与对应实现/合并者独立 Targeted 全部必要断言，完成后自动 fresh S7，无额外 Handoff 答卷；真实测试和明确用户审批照常履行。

S9 入口已有 S8 批准 Contract，Main 阅读并明确目标、断言、分工，不重新审批或另设 readback/启动许可，不逐步新增人审。
工作区准备是一次正常动作，系统给清晰现入口和结果，无第三份准备问卷。
普通工具不应触 Begin、缺卷发行或 workspace prepare；本轮 lead/verifier context 与保存后流转已按责任开始/工序切换拆开，其他分支的完整验收仍待真实链。
解除 `RedChecks.Command → delivery Checks` 绑定，删除 `scopeAllows`、packet/source proof 多层额外资格重验，不能藏进保存 consumer。
fail/blocked/有偏差的真实报告可记录，记录成功不等于修复完成；完成依实际未完成工序判断。
必要一致性只确保答案准确记录到当前责任，不兜底保留旧门槛；故障优先删不必要依赖，不新增并发重试、ID 映射或补偿层。

下一轮验收先走一个真实正常 S9 到 fresh S7，再做必要的真实 blocked 提醒与普通工具/同答不重复执行小验证。
不为旧门槛扩大矩阵；看是否还要猜机制、补协议、重复交卷、ordinary Hook 重准备。
Hook exit 0 或机械测试数量不能代替减负验收。

以下 R59–R63 是旧实现的失败历史，保留其当时顺序与机械证据，不继续作为当前门槛。

| 历史快照 | 已确认事实与未通过边界 | 证据 |
| --- | --- | --- |
| R59 | 机械 51 项是历史基线；真实 Main 指示代存且 Builder 提前改 Main-root 产品，本次实际失败 | [正常流程失败报告](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r59-s9-normal-flow-failure.json) |
| R60 | 两名 Builder 本人 Plan 与原 Begin 已过；fixture 未跟踪已安装 assets，原 prepare 因 Git 来源拒绝。新基线已修 fixture，未放松原 Git 校验，本次非完整通过 | [fixture 基线阻塞](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r60-s9-fixture-baseline-blocker.json) |
| R61 | 两份 Plan 与 Worker 已过；原 AgentBegin 投影仍 pending 时先 Begin，使原恢复来源失效。R62 修正原投影先于 Begin，本次非完整通过 | [Plan 激活顺序失败](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r61-s9-plan-activation-order-failure.json) |
| R62 | 冻结源码定向 15 项通过（14 CLI 加原 Plan public Hook；初次共享缓存编译失败及独立缓存重跑均保留）。真实两人 Plan 首交、原 activation/work_started 先于唯一 Begin、两 execution ready、unit1 completion/candidate 成功；unit2 已登记但物理卷缺失，全量 issuer 反复触 4 秒 deadline，进程自然结束，非完整通过 | [机械定向](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r62-s9-mechanical-final.json)、[缺卷全量准备失败](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r62-s9-missing-result-bulk-issue-failure.json) |
| R63 | 仅将 L1 缺卷准备收缩为从原事实选择真正缺卷 owner、调用该本人原 issuer；8 项匹配源码的定向测试及构建通过。真实两份 Plan、原激活、唯一 Begin、两 execution ready 与两份 Result view/answer 实体均已到位，前段 Hook 超时为0；但两作者将实际前检命令写成文字概述，不适合原 prepare 冻结后由 Main 原样执行，提前判本次输入不合格并终止保留，未完整通过 | [增量定向](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r63-s9-listener-delta-mechanical.json)、[冻结源码与构建](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/after-r63-source-inventory.json)、[不可执行冻结命令失败](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r63-s9-nonrunnable-frozen-check-authoring-failure.json) |

历史 R63 包 SHA-256：`d105a60c6d30d4a2056138ed88dac23e7c1a7c0324ac794e7a8685fac4acfa6c`。
8 项是 R63 缺卷增量覆盖，15 项是 R62 冻结定向，51 项是 R59 历史基线，不能合计为 R63 的完整回归。

当前尚无完整真实 S9 闭环或 fresh S7 交接通过证据，也未验证全平台、全分支真实模型。
上游 synthetic fixture 借用历史真实 Finding，不是新 S7 通过；S8 的仿真人决策及入口证据不替代 S9 验收。
后续仍须实际 Main 集成/checks、Impact、独立 Targeted、必要 blocked 恢复与原 Handoff 到 fresh S7 的完整证据；
不增加全局阶段完成计数，不把在途模型或已准备现场写成 PASS。
