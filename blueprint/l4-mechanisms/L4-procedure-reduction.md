# L4 — 手续收敛

> 层：第四层｜机制域：Agent 面对的命令链、名字、重复载体和占位符如何收成既有门禁能消费的一次写入
>
> 上游：L1 D1/D2/D3/D4 与公理五；[L4 阶段答卷](L4-stage-form.md)；[L4 revision](L4-revision-usage.md)；[L4 Agent 调度](L4-agent-dispatch-governance.md)；[L4 项目级设计基础](L4-project-design-foundation.md)
>
> 状态：v0.3.0 目标设计。本文提出的统一手续收敛尚未在生产实现；既有机制的实现情况见第 12 节。[S5 隔离原型](../../tools/stage-form-prototype/README.md)仅验证交互契约。本文不新增状态机，不放宽既有门禁。答卷篇规定生成、保存与消费；本文规定执行者应保留哪些工作、移走哪些手续。

## 1. 核对

2026-10-12 对照 `docs/control/agent-protocol.md`、`internal/cli/action_catalog.go`、`internal/cli/repair_command.go`、`internal/cli/investigation_command.go`、`internal/assignment/lifecycle.go`、`internal/review/draft.go`、`internal/review/manifest_draft.go`、`internal/review/plan.go`、`internal/schema/assets/agent-message.schema.json`。六条里有四条按原话成立，两条要收窄。

| 发现 | 核对 | 收窄 |
|:--|:--|:--|
| 一条判断要敲一串命令 | 成立。S7 目录是 `s7 draft` → `runtime review-plan` → `s7 manifest-draft` → `runtime register-workgroup` → `runtime review-result submit`，协议还要求从 `s7 status` 抄 `subject_digest`，冷启动再抄 `s7 workspace-digest`。S8 是 `ingest`、`hypothesis register`、`dispatch`、`hypothesis result`、`consume`、`route`（三份 JSON）、`contract approve`（`--approval-hash` 与 evidence id），另有可选 `project`。S9 的 `impact`、`targeted`、`handoff` 都是 `create` 与 `commit` 两步，前后还有 `session open`、`plan compile`、`dispatch`、`plan-report submit`、`execution begin`、`result submit`、`changeset compute` | revision 篇第 3.3 节已经要求复合动作在一个写入里完成。缺口在命令面还没照做 |
| 同一位置有四套名字 | 成立。协议末尾对照表同时使用 S0–S11、`lifecycle`、`phase`、`TR`/`PTR`。S8 的机器相位是 `bug_resolution.investigation`。S7、S9 两行写明当前相位名仍是兼容投影，与目标名不一致 | 内部游标、journal 和诊断保留机器名。收的是给 Agent 的恢复包和阶段正文 |
| 同一事实写了两到三遍 | 部分成立。完成报告 schema 将 `activation_id`、`requested_event` 列为告警必填，协议要求 Agent 对 readback 文件跑 `shasum` 填 `approved_readback_sha256`。S10 同时要 Markdown、审计 manifest 和带 `audit_manifest_sha256` 的 `evidence add`。S8 的 `project` 在批准后仍出现在 `next` 段落里 | REQ → 设计 → 契约 → TASK 这条链是追溯结构，保留。收的是全文复述，以及要 Agent 亲手写的派生字段和兼容投影 |
| 生成稿里有会拒收的占位符 | 部分成立。`DraftPlan` 与 `DraftManifest` 写入 `TODO(planner)`，`ValidatePlan` 和身份校验拒绝该字面量。manifest 注释写明控制面能证明的字段预填，其余留占位 | 两类分开。控制面已经知道的值由工具写入或在派发时盖上。真正缺的上游（例如没有可读的 `cases.json`）由门指出缺的输入。留给 Agent 的判断用答卷空位，空位是 `null`，不用会像正文的占位词 |
| 派发是另一台状态机 | 成立。`lifecycle.go` 的正式事件是 12 个：`readback_started`、`readback_submitted`、`understanding_approved`、`understanding_rejected`、`document_conflict_reported`、`activation_sent`、`work_started`、`completion_reported`、`completion_acknowledged`、`work_blocked`、`blocker_resolved`、`shutdown_approved`。协议仍要求 `PLAN_REPORT` 走 SendMessage，并同时说明 `SubagentStop` 与 `TeammateIdle` | `plan_checkpoint` 已有自动链，这 12 个事件留在 Writer 内。Agent 路径是派发和把结果写入答卷 |
| 进环前要先走分类树 | 成立。`agent-protocol.md` 的 S0 第 1 步把 UI 影响、投入级别、thin/full、Foundation 状态写在同一个动作里 | 分类表的权威仍是设计基础篇。S0 只问人两件事，其余由 `design-foundation` 计算并给出待确认稿 |

## 2. 留下的判断

下列能力继续由原来的门负责：

- 交付、质量、端到端各自给出结论，审查者不审自己写的实现。
- 表象先形成证据充分的因果模型和批准后的 RepairContract，再改产品。
- 需求锁定和发布决定由人做出。
- 答卷写完只表示判断已提交。空位仍空、路径读不到、发现没有可定位的观察、实现超出任务范围，门仍不过。

## 3. 一次写入

当前责任对 Agent 的工作入口是实际派发或填写生成的当前题目。已接收者另有明确的完成/等待回执，不会继续被要求派发或交卷。每阶段 Main 只有一份逻辑汇总，各执行者填写独占分卷；具体题目、阅读目的和示例在卷里，不藏在零散提示中。

S7 的覆盖与 Claim 划分消费已经批准的 ReviewPlan，发卷时绑定原冻结资料，关键说明随卷保存。规划者仍须判断覆盖和风险，机器不从文件清单发明完整语义覆盖。审查者填结论、观察、发现和适用性理由；消费时核对发卷版本并计算实际验证产物 digest，不能把当前新 subject 的哈希绑定旧判断。

S8 的假设、因果、影响面、检测缺口和路由是同一份调查卷里的空位。批准 RepairContract 时，Hook 计算草稿哈希并登记一次性 evidence。三份旁路 JSON 和 `--approval-hash` 不再是 Agent 步骤。

S9 里 Agent 要写的是修复计划、修复结果和定向重验结论。`changeset` 由工具按会话基线计算。`impact`、`targeted`、`handoff` 的 create 与 commit 收成同一次写入：卷里的判断就是提交内容，Writer 在同一事务里绑定当前 Session、Contract 和路径哈希。`targeted resume` 仍是受阻之后的恢复动作，由 `repair status` 的 `next` 点名。

诊断命令保留：`s7 status`、`runtime investigation status`、`runtime repair status`。它们的第一行是当前 S 阶段和唯一下一动作。

## 4. 一套对外名字

恢复包、Milestone 和阶段 Skill 只出现两样东西：当前 `S00`–`S11`，以及第 3 节的那一个入口。

`planning`、`verification`、`bug_resolution`、相位名、`TR-*`、`PTR-BUG-*` 留在 `loop-definition.json`、journal 和 `doctor`。S7、S8、S9 的对外阶段以 Main Spine 为准，不再把「目标名」和「当前兼容相位」同时交给 Agent。

`status` 展开项可以附带机器游标，放在下一动作之后，供诊断阅读。

## 5. 一个手写载体

答卷是 Agent 的手写入口。Markdown 报告、证据信封和兼容 BUG 在写入通过校验后由工具生成，供人读和旧消费者使用。

从 Agent 必填字段中去掉这些派生量：`message_id`、`correlation_id`、`activation_id`、`requested_event`、`approved_readback_sha256`、`subject_digest`、`verification_artifact_digest`、`audit_manifest_sha256`、`approval_hash`。身份和哈希按答卷篇第 4、6、7 节，从真实派发绑定与被冻结的输入写入；answer 自填身份不赋予权限。

证据引用只收仓库路径或捕获 id。`path:#sha256=` 这种由 Agent 拼接的形式不再作为交卷格式。

S2、S3、S4 的卷记录本层新增的决定和指向上一层的路径。契约层记录分端责任与协议差异，任务层记录范围、收尾检查和依赖。已在上一份卷里写过的目标、约束和验收不再整段抄入。S5 继续核对链接和增量是否与锁定基线一致。

`runtime investigation project` 留在兼容命令里，不出现在 S8 的 `next` 中。

## 6. 生成稿只留空位

发卷时分三种格子：

| 格子 | 写入 | Agent |
|:--|:--|:--|
| 控制面已有的值 | 终值。文档路径、版本、代数、轮次、digest 在生成时绑定。`agent_id` 在真实派发时盖上，草稿里不出现该字段 | 不改 |
| 缺上游 | 不生成看起来已完成的计划。回执写明缺哪份输入，例如没有可读的场景清单 | 先补上游 |
| 要判断 | `null` 或空对象。Claim 的断定、不适用理由、发现的 expected 与 observed | 写入答卷 |

`TODO(planner)` 和 `PLANNER-REFINE` 不再写入生成稿。注册校验继续拒绝这两个字面量，用来守住旧草稿。

## 7. 派发收成两步

12 个 Agent 事件继续由 Writer 推进。`plan_checkpoint` 的自动链保持为内部实现。

Agent 看到的派发是：主会话按 `next` 派出责任；执行者读当前分卷，只填写自己的题目。需要写前计划时系统先发计划题，接收后才继续实现并暴露结果题；不能在最终 completion 时追认计划。Hook 从对应保存和返回消费当前作业，并推进原内部事件。

高风险任务仍可要求人看计划。该要求出现在 `next` 里，作为一次明确的等待，而不是第二套激活说明。

Main 的调度动作是派出 `next` 点名的责任，或收下已经写入的结果。普通计划不需要 Main 再回一句批准。

SendMessage 仍可作为平台上的观察点。它不是完成一份工作的必经手续；必经手续是写入答卷。

## 8. S0 只问两件事

设计基础篇继续定义投入级别、thin/full 和 Foundation 状态。S0 的 Agent 动作收成两问：

1. 这次是否改变用户可见界面。
2. 人是否确认这版需求。

`design-foundation` 根据 `project-map` 和现有 `DESIGN.md` 提取事实、计算可判定缺口，并生成投入/路径建议。投资取舍与审美方向仍由设计者提出、按原职责确认，不把它们伪装成确定性计算。缺口未补上时，回执指向具体稿和需要作出的判断，Agent 不从协议分类树里猜手续。

两问是可见入口的收敛，不删除需求漏斗中的实质价值判断或人类锁定。S01 启动与 S11 发布也必须消费明确人决定，不能由文件存在或 JSON 填 approve 推断。

## 9. 落地顺序

1. 先打通 S05 注册模板、发卷、真实独立 owner、保存消费、原门、回执与恢复的一条完整链，不能只改 status 第一行。
2. S03/S04 确定性登记接入同一输入路径，去掉手续空位；恢复包只给当前责任和填写路径。
3. S6/S7/S8/S9/S10 各自 adapter 接原领域 API，去掉手抄身份、digest、派生 envelope 和事件；保留实质判断与计划先决条件。
4. S9 impact/targeted/handoff 的 create/commit 由 consumer 接续；长集成保留系统待办，失败现场沿原恢复协议保全。
5. 生成稿去掉字面占位符；缺上游给具体字段/owner。S2–S4 增量答案，S0 分类给已计算事实与可审阅建议。
6. 每阶段真实保存→接收→原门验收通过后，才移走 Skill/角色/协议中的旧手工入口；未接线环境明确给生成好的兼容动作。

每一步完成前，对应阶段的原门测试仍然有效。手续减少以 Agent 必填字段、必跑动词、协议查找和已完成责任的重复指令次数下降来计，不以删掉检查来计。

## 10. 按零上下文执行者复审最新提交

以下均是设计/接线问题，不能靠让执行者再认真阅读解决。原 2026-10-12 核对保留为当时事实，新增契约以答卷 v0.6.0 为准。

| 问题 | 会怎样继续打转 | 收敛后的契约 |
| --- | --- | --- |
| 只有字段空位，没有旁边的问题/值域 | 猜 conclusion 用 fail 还是 finding，空 finding 要删还是填 | Schema 与问题一起注册；题目、枚举、条件、示例、读序随卷 |
| 只说写入即交卷，没有成功保存后的 consumer | 当前 PostToolUse 不监听 Write/Edit 且提前进入观察分支；工作已结束仍没有消费事件 | 保存后接入原领域 adapter 并立即求门，真实 wire 测试与恢复兜底 |
| 所有人手写同一份 JSON | 过期整卷覆盖已回答责任，迫使他人补交 | 独占输入、Writer 合并、Main 单一汇总；不新增第二事实源 |
| 每次用当前文件新哈希覆盖登记 | 旧结论可能被重新绑定新产品，假装仍通过 | 发卷冻结输入；消费核对当前适用性；合法可变 evidence 才按白名单刷新 |
| S1/S11 被称为只记状态 | 文件齐全被误当授权，或隐藏人闸使 Agent 无限尝试 | 显式人决定到达后自动记账；回执明确等待谁裁定 |
| 一阶段强行一次最终回答 | S6/S9 的写前计划只能在结果后补，造成权限死锁或追认 | 同一责任路径按原时序只发当前题目，先计划后结果 |
| 只给阶段未通过，没有责任接收回执 | 已完成者把等待同事/集成/系统恢复误当自身未完成 | work_complete 与 Stage 分开，待办有 owner；本人的已接收回答不重复要求 |
| 生成 JSON 后仍让 Agent 注册旧信封 | 问卷只是增加一个文件，未减少手续 | adapter 派生原 evidence/消息/路由；每个原门的手续字段有机器生产者 |

不能让所有剩余工作都归“填写 JSON”。实际代码、原型、commit、真实测试、派发、合并和人判断仍是工作；卷消除的是模糊交接、重复记录与猜命令。每个未通过项必须归为具体待答题、真实交付先决条件、人决定或有恢复路径的系统待办；无归属项是模板/adapter 缺陷。

## 11. 首次使用验收卡

只给新的执行者当前卷与必要工程资料，不给 L4 设计历史，让它完成一轮实际责任，观察：

1. 首次保存前是否搜索内部 Schema、哈希命令或事件名；每次搜索都指向一个需要移入题目或机器的缺口。
2. 报错能否只改所指的 answer JSON Pointer，已有回答是否保留。
3. 最后保存之后是否立即收到接收和推进/失败路由；无须新 prompt 或无关工具。
4. 同事未交卷、Main 待集成、系统恢复和人闸时，是否明确自身已完成且知道等待谁。
5. 旧卷、重复事件、并行输入和阶段重入是否保留历史并防止错绑。

隔离原型是契约试验；真实 Claude Hook 与原领域 gate 接线通过以前，以上自动化不能标为实现完成。覆盖 S5 的不同结论后，再按阶段推广，避免一次创造十二套未经实战验证的模板。

## 12. 协议要求到执行机制的覆盖审查

2026-10-13 对照 [Agent Protocol](../../docs/control/agent-protocol.md)、[Harness Manual](../../loop-harness.md) 与以下实现。这里审查的是要求是否在实际执行入口变成可消费的工作，不能把“协议中已经描述”或“Hook 输出过一句提醒”计为自动化完成。本节是代码与契约审查，不是一次真实 Agent 全链路运行验收。

### 12.1 已落地的基础

- Controller/Writer 已负责原门求值、单事件至多一次允许的转换、Runtime 与 journal 的持久化；普通阶段未满足不因此阻止合法工作。参见 [控制周期](../../internal/controller/cycle.go) 与 [PreToolUse 投影](../../internal/hook/pretooluse.go)。
- SessionStart/SubagentStart 已有原生 `additionalContext` 注入，并按真实 Assignment 匹配，歧义时不冒充另一个 owner。参见 [生命周期上下文](../../internal/hook/lifecycle.go)。
- 普通 `plan_checkpoint` 已有 PLAN_REPORT 消费和激活自动链；这是现成能力，不能重新列为全无实现。参见 [自动链](../../internal/assignment/auto_chain.go)。
- 工作区 prepare/launch 已检查冻结输入、执行身份、角色和检查环境，prepare 输出精确操作命令；pending 从持久化报告/checkpoint 派生恢复命令。参见 [工作区命令](../../internal/cli/workspace_command.go)、[启动](../../internal/workspace/launch.go) 与 [工作区说明](../../docs/workspace-integration.md)。
- S7 的部分校验已返回结构化 missing/repair/next/verify；TeammateIdle 的重复恢复提醒已有同事实预算。参见 [领域诊断](../../internal/review/gate_diagnostic.go) 与 [Idle 预算](../../internal/hook/idle_budget.go)。

以上证明存在控制、投影和恢复能力。它们尚未统一保证执行者只凭当前作业包就能完成、交付并退出自己的责任。

### 12.2 尚未闭合的能力

| 编号 | 现有落点 | 执行者仍须猜什么 | 所需机制与验收 |
| --- | --- | --- | --- |
| M01 当前动作的具体交付 | [Guidance](../../internal/cli/controller.go) 的正文和 [普通工具摘要](../../internal/hook/progress.go) 仍指向协议/说明书；ReadOrder 包含 AGENTS、Runtime、阶段协议后再读工程输入 | 当前具体交哪一份东西、放哪、哪些题要答 | 发包直接给已解析的作业、交付路径和当前题；协议链接供追溯。隐藏流程说明后仍能交付；必要业务/工程资料照读 |
| M02 当前工作由真实缺口派生 | [projectionContracts/contractFor](../../internal/cli/projection.go) 含静态与启发式 missing；[projectNext](../../internal/cli/run.go) 含阶段概述，真实 gate 另给 missing | 阶段概述、投影 missing、实际门缺口应如何合成下一动作 | 同一领域投影将当前门缺口映射为带 owner 的作业、等待或系统恢复；status/next/Hook 消费同一结果。每个实际缺口有归属，机器转换不列为 Agent 交付 |
| M03 强制信息随职责送达 | [SubagentStart 上下文](../../internal/hook/lifecycle.go) 已带 scope/checks/done_when，但 scope 排在 next 前，整串按 800 字节截断；无 done_when 的兼容情况要求再读 Assignment | 被截掉的下一动作，以及旧 manifest 缺失的收尾条件 | 有界上下文先保留身份、当前动作、收尾与作业包路径，长范围放包内；职责原生必需项不得静默丢失。长 scope 与恢复场景验收 |
| M04 自己完成与全局完成分开 | [Guidance](../../internal/policy/engine.go) 以 stage/missing/blocked 为主；Stop/Idle 有报告判定，集成也有 checkpoint | 自己还欠交付，还是同事/Main/系统欠工作 | 每个接收结果投影 responsibility_status、work_complete、waiting_on、next_owner；Worker 只见自己待办，Main 见派发/接收待办。本人已接收且仅等其他 owner 时不重复催交 |
| M05 最后一次交付触发接收与求门 | [settings](../../settings.json) 的 PostToolUse 不匹配 Write/Edit；[PostToolUse 分支](../../internal/cli/run.go) 是观察路径；普通阶段推进由后续 PreToolUse 检查 | 文件已经保存为何还未接收，是否需要再随便调用一个工具 | 答卷保存事件接消费器，原领域写入后求门，或记录明确的系统待办并恢复；每事件最多一次允许转换，不在 Hook 跑长合并。最后交付后无额外 prompt/无关工具也有明确回执 |
| M06 不合格草稿转成可回答题目 | [ReviewPlan](../../internal/review/draft.go) 与 [manifest 草稿](../../internal/review/manifest_draft.go) 有 TODO；工作区 launch 已给 submission 路径，却仍要求使用协议 schemas | 哪些是缺输入，哪些要专业判断，结构和枚举怎么写 | 已知字段由机器生成，缺上游返回对应 owner，要判断的字段给 null、题目/值域/示例。不保留需要执行者搜索 Schema 才能开始的空壳 |
| M07 修正动作覆盖与可执行性 | [missing-token legend](../../internal/missingtokens/missingtokens.go) 覆盖 Builder 与若干 S7 门，许多动作仍有占位符；未识别 token 要求查看门实现/explain。部分 S7 错误已有结构化诊断 | 参数取哪里、谁执行、要改哪项；是否还须读 Harness 源码 | 从实际对象生成参数完整的动作，回答错误定位到字段/证据；未知映射归 adapter 维护/系统恢复，不派给业务 Worker 猜。每类实际 gate 缺口验收 |
| M08 开工就绪检查的统一覆盖 | 工作区 prepare/launch 已有输入、环境、身份预检；原生调度还有另一路前置检查和提醒 | 本次入口能否获得同等完整的可开工条件 | 各实际支持入口消费领域对应的就绪判据；缺资料、缺授权、缺环境分别归责任人。不能以“工作区入口做了预检”宣称所有阶段/平台均已覆盖 |
| M09 无进展的识别与责任移交 | [Idle 预算](../../internal/hook/idle_budget.go) 只限制指定 TeammateIdle 恢复提醒；[恢复 replay](../../internal/controller/replay.go) 防循环用于恢复过程 | 普通连续操作遇到相同缺口时还要尝试多久 | 按作业和具体缺口记录是否有新输入、修改或证据；重复无进展形成明确诊断和恢复责任。预算不能把未完成改成完成，也不把正常长运行或等待当成死循环 |
| M10 首次使用的行为验收 | 已有 Hook wire、文档动词存在、领域门和恢复测试；本次看到的这些检查分别证明实现与引用一致 | 不读流程文档的执行者是否仍能完成 | 增加跨入口的首次使用实战验收：只提供作业包/工程输入，记录查协议、找 Schema、拒收、重复操作、交后滞留；通过才宣称手续已收敛 |

M02 不是已复现的所有阶段投影错误；代码表明投影与门缺口有不同生产路径，统一一致性需要逐阶段验证。M04 也不否定既有报告/集成终态；缺口在统一对执行者暴露责任完成状态。审查结论必须保持这些边界。

### 12.3 提醒何时成为机制

以下三类要求分别落实，不能统统改成硬阻断，也不能只增加提示字数：

| 要求 | 机制落点 | 例子 |
| --- | --- | --- |
| 系统已知的事实与确定性手续 | 派发/接收 adapter 自动生成，保留原领域门 | owner、输入绑定、路径、事件、摘要、具体命令 |
| 必须先满足的工程条件 | 开工前预检或原有权限/依赖门，失败归对应 owner | 冻结输入可读、环境可用、写前计划、人决定 |
| 专业判断 | 当前题与必要工程资料同时送达，接收时核对证据和独立责任 | 根因、覆盖、风险、设计取舍 |

“去阅读另一份流程文件”可以是故障诊断的补充；正常任务的操作格式、当前责任和收尾条件必须在工作入口闭合。业务规格与源代码仍是必要阅读，不能把不读流程手册扩大为不读工程输入。

## 13. 将既有协议要求落到机制的次序

1. **先闭合责任与交付事件（M02/M04/M05）**：真实缺口归 owner，接收后给责任完成回执，最后交付触发消费与原门求值。先做 S5 完整链，再按领域扩展；Main 的合法长集成是明确待办，不丢给已完成 Worker。
2. **让当前作业可直接执行（M01/M03/M06/M07）**：复用既有 Assignment、模板 registry、领域诊断和工作区精确命令；必需题、路径、收尾和已解析动作一起送达。有界消息保留入口，详细包保持同一 Runtime/Assignment 绑定。
3. **统一开工与恢复（M08/M09）**：扩展各入口预检覆盖、记录无进展原因与恢复 owner。系统故障不能变成无限的“继续 DRIVE”，等待人或其他 owner 也不算当前执行者欠交付。
4. **逐项验收再收旧提醒（M10）**：新增/修改协议的每项强制要求，登记其生产者、交付事件、消费者、失败归属和实际入口测试。只有这些落点存在，才计为机制已实现；仅有文档或输出字符串记为指导/部分实现。

本顺序复用原 Runtime、Assignment、领域门和 checkpoint，不新增一套生命周期。第 12 节是实施缺口清单；本次审查未修改生产 Harness。

## 变更记录

| 日期 | 版本 | 变更 |
|:--|:--|:--|
| 2026-10-12 | v0.1.0 | 按命令面、协议和 schema 核对六条减负发现，写入手续收敛。未实现 |
| 2026-10-13 | v0.2.0 | 从无设计上下文执行者复审，补保存消费、字段旁引导、独占分卷、工作完成回执、冻结/人闸/计划边界与逐阶段真实验收；生产接线仍未实现 |
| 2026-10-13 | v0.3.0 | 对照协议、Harness Manual 与实现，区分已有基础和 M01–M10 执行缺口，补提醒转机制的落点与责任/接收优先的实施次序；未修改生产 Harness |
