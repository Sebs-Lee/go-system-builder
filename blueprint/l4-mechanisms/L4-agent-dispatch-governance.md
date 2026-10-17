# L4 — Agent 调度与治理机制（Sub-agent / Agent Team Dispatch Governance）

> 层：第四层｜机制域：跨 Stage Agent 调度、执行连续性与治理｜平台基线：Claude Code 2.1.218
>
> 上游：L1 D1～D7 与五公理、L2 生命周期目标、所有需要委派职责的 L3 Stage
>
> 下游：实现资产：Skill、Agent Definition、Schema、Harness、Hook、Assignment Record、测试和迁移任务
>
> 状态：v0.5.0 目标设计。本文定义机制应达到的统一行为；§15 及各 L3 的当前事实节列出现有实现与目标的差距，不把设计意图写成既成事实。

## 0. L4 的位置：为什么这不是另一份 L3

L3 按 Stage 回答：

- 这个阶段为什么存在；
- 需要完成什么；
- 哪些角色、产物和门共同形成阶段闭环；
- 失败时回到哪个生命周期位置。

L4 按跨阶段机制回答：

- 同一类工具机制在多个 Stage 中如何保持一致；
- Claude Code 平台对象、项目 Runtime 与持久化记录如何对齐；
- 派发、消息、等待、idle、stop、恢复和收口的精确语义是什么；
- Hook、Harness、Schema 和 Skill 各自承担哪一段责任；
- 哪些 L3 只引用机制，不能各写一套变体。

因此，Agent 调度不属于 S6 私有能力。S5 的独立文档审查、S6 的 Builder、S7 的 Delivery/QA/E2E、S8 的并行假设调查、S9 的修复与定向复验，都消费同一套调度治理机制。

本 L4 是第一份跨 Stage 工具机制设计。L3 只声明“此处需要哪种派发模式和阶段特有完成条件”；派发细节以本文为唯一目标态设计。

## 1. 问题定义与机制目标

### 1.1 当前要解决的根问题

Agent 委派不是一次 API 调用，而是一条跨多个控制面的执行链：

1. Main Agent 判断是否值得委派；
2. 选择普通 Sub-agent 还是 Agent Team teammate；
3. 固定 TASK、责任、scope、依赖和完成条件；
4. Worker 阅读真实项目文档并形成计划；
5. Worker 持续执行并向 Main Agent 回报；
6. Hook 识别过早 idle、过早 stop、缺计划、缺报告和越界；
7. Runtime 记录持久事实；
8. 结果被集成、验证或阶段 Gate 消费；
9. 中断后能够恢复同一责任，而不是重新猜测上下文。

任何一段没有闭合，都会产生典型失效：

- 第一轮回读后 Worker 进入 idle，Main Agent 来不及第二轮授意；
- 项目 Runtime 把 Agent 改回 active，但 Claude Code 平台会话没有被唤醒；
- Main Agent 以为自己在等结果，实际 Worker 已停止或隐藏；
- Worker 报告了计划，却没有继续执行；
- Main Agent 没收到计划就自行接管已委派责任；
- teammate idle 时被自动分配下一项任务，但上一项尚未集成或消费；
- 同一事实分别存在于 Team task、Runtime agent、TASK state、message 和 evidence 中，彼此漂移；
- Hook 只输出提醒，没有使用 Claude Code 的真实 decision control。

### 1.2 目标

本机制要实现：

- **统一选择**：Main Agent 能确定何时自执行、何时用普通 Sub-agent、何时用 Agent Team；
- **连续执行**：默认 Worker 回报计划后不停顿，直接完成 assignment；
- **风险分层**：普通任务使用计划检查点，高风险任务使用原生 Plan approval；
- **单一权威**：Assignment 是责任、scope、计划状态和结果状态的权威来源；
- **真实控制**：Hook 使用平台支持的 continue/block/exit 语义，不用 Runtime 文案模拟唤醒；
- **最少打断**：Main Agent 只在计划缺失、计划偏移、执行偏移、阻塞或完成时介入；
- **可恢复**：同一 Agent 可继续时继续同一 Agent；不可继续时从持久 Assignment 和 checkpoint 重派；
- **可观测收敛**：每个 assignment 能回答“当前在哪、还差什么、下一步由谁做”。
- **容量不裁责任**：L3 可声明 `bounded_flow` 或 `coverage_complete`；物理并发容量只决定启动顺序，不得让 coverage-critical Stage 删除 required Assignment。

### 1.3 非目标

- 不用 Agent 数量制造“团队感”；但当独立 coverage/context 确实存在时，不以少 Agent 为优化目标；
- 不要求所有任务都经过人工或 Main Agent 的计划批准；
- 不把聊天记录当作需求、scope 或完成状态的权威来源；
- 不让 Hook 判断业务设计优劣；
- 不允许 teammate 自己扩大任务、创建需求或越过人闸；
- 不用长协议弥补工具没有接线的问题；
- 不为 Claude Code 2.1.218 以前的行为保留第二套协议；环境不满足时升级或修复环境。

### 1.4 复杂度—收益减法审计

| 现有/候选机制 | 工程判断 | 处置 |
|:--|:--|:--|
| 所有 Worker 固定 readback → approval → activation | 协调、idle 和上下文切换成本高；普通可回滚任务没有等量风险收益 | 删除为默认，只在高风险任务保留原生 Plan approval |
| 把 PLAN_REPORT 当第一轮 final response | 直接终止 Sub-agent 本轮，制造恢复/重派和“第二次授意” | 禁止；改为运行中 SendMessage，发送后连续执行 |
| Main 对每份正常计划回复“批准开工” | 不改变 scope、权限或风险，只增加消息和等待 | 删除；计划无偏移时静默 |
| 大而全的 readback 字段复制 TASK/契约 | 重复真相、增加 token、字段无人消费 | 收缩为 objective/paths/steps/checks/dependencies/risks |
| Runtime 把状态改成 active 来模拟平台唤醒 | 两个控制面语义不相等，形成假恢复 | 删除；只用 SendMessage、exit 2 或平台 resume |
| TeammateIdle 自动领取/分配下一 TASK | 上一结果可能未消费，绕过依赖、冲突和 WIP 重算 | 删除；idle 只处理当前 Assignment |
| Team task、Runtime agent、消息、evidence 各存一份完成事实 | 多写者导致漂移，Gate 无法知道该信谁 | Assignment/Result 为工程权威，其他只投影 |
| 用自造 hook 字段和 systemMessage 假装控制 | 测试通过但真实 Claude Code 不会继续 Worker | 删除；使用官方 payload 与 exit/decision control |
| 每个 Stage 各写一套派发协议 | 漂移成本随 L3 数量线性增长，修一处不能全局生效 | 上移 L4；L3 只声明消费模式和阶段完成条件 |
| 所有任务都建 Agent Team | Team token、冲突与收口成本对小任务收益不足 | 按 one-shot / Sub-agent / teammate 三路选择 |
| 所有 Stage 固定 WIP=2/token budget | 容量易预测，但会让 S7 等 coverage-critical Stage 用漏检换短期成本 | 增加 stage capacity policy；默认 `bounded_flow`，S7 使用 `coverage_complete`，排队不裁 required coverage |
| 强制周期性 PROGRESS 消息 | 常态噪声高，Main 注意力收益低 | 默认可选；只有长任务或超时诊断启用 |
| 继续扩写 agent-protocol 长文 | 依赖模型读完、记住并自觉执行，不在自然路径上 | 协议只保留入口/不变量；细则埋入 Skill、Assignment、Hook、Result 和报错反馈 |

### 1.5 机制准入规则

候选机制只有满足以下至少一项，才进入持久化控制面：

1. 能阻止一个确定性非法转换或越权动作；
2. 能产生下游必须消费的事实；
3. 能减少 Main/Reviewer 的重复判断或重复取证。

只用于解释、统计或展示的内容必须是计算视图，不新增 ID、生命周期和审批门。S7 的 Claims、Assignment 和 Result 是事实；静态/E2E/Discovery coverage、Ready/Queued/Dispatchable、Case ledger 和 dashboard 都是视图。`coverage_complete` 只保证 required work 不因平台容量被裁剪，不授权 Agent 无来源地递归扩大任务范围。

## 2. Claude Code 对象模型与项目权威

### 2.1 平台对象

| 对象 | 平台语义 | 适合承担 |
|:--|:--|:--|
| Main session / lead | 当前会话的编排者和结果消费者 | 调度、例外判断、计划偏移判断、结果收口 |
| 普通 Sub-agent | 一次独立执行上下文，最终结果返回调用者；实际可前台或后台运行 | 自包含的一次性研究、检查、隔离写入任务 |
| 可寻址自定义 Sub-agent | 有稳定名称/ID，可在运行中通过 SendMessage 回报/纠偏，结束后可恢复 | 需要计划回执、worktree 隔离或中途指导的一次性责任 |
| Agent Team teammate | 独立 Claude Code 会话，共享 Team task 视图，可直接 SendMessage | 持续执行、并行责任、计划回执、跨任务协作 |
| Agent Definition | 角色静态边界、工具、模型和最大能力 | 定义“这个角色最多能做什么” |
| Plan mode / Plan approval | teammate 只读规划并向 lead 请求批准 | 高风险任务的执行前审批 |
| Task list | Claude Code 的协作和依赖视图 | 可视化 Ready/Pending/In Progress/Completed |
| Worktree | 隔离修改和保留失败现场 | Builder/repair 等写入责任 |
| SendMessage | Agent 间消息与运行中纠偏通道 | PLAN_REPORT、BLOCKER、CORRECTION、状态通知 |

### 2.2 项目权威分工

| 事实 | 唯一权威 | 平台对象的地位 |
|:--|:--|:--|
| 当前生命周期 Stage | .claude/loop-state.json | Team task 不决定 Stage |
| assignment 身份、Owner、scope、revision | Assignment Record | spawn prompt 和 Team task 是投影 |
| TASK/规格真相 | 已登记文档及指纹 | 消息只引用，不复制为新真相 |
| plan 是否已回报 | Assignment Record 的 plan checkpoint | mailbox/transcript 是传输和审计来源 |
| Worker 是否正在平台运行 | Claude Code teammate/subagent 状态 | Runtime 不能凭空唤醒平台 |
| assignment 是否完成 | canonical Result + 消费检查 | Agent final text 不是完成状态 |
| 写入是否可接受 | 实际 diff + integration/consumer gate | planned_paths 不是实际 diff |
| Stage 是否可推进 | 对应 Quality Gate | teammate/task completed 不自动推进 Stage |

核心原则：

> Runtime 记录工程事实，Claude Code 管理活会话；两者互相投影，但任何一方都不能伪装成另一方。

跨 worktree 还有一条硬边界：

> 产品写入面可以按 worktree 隔离；Assignment、plan、Result、checkpoint 等调度控制面必须解析到同一个项目级共享存储，并使用 CAS/revision 写入。它们不能跟随 Worker 当前 cwd 各写一份，也不能靠 git merge 汇总运行态。

L4 只规定这个一致性契约；共享存储的具体路径、锁和事务格式由实现代码与 Schema 定义。PostToolUse/PreToolUse/Harness 若无法定位共享控制面必须失败并给出恢复动作，不得退化为写 Worker worktree 中的“本地真相”。

### 2.3 平台基线

本设计以 Claude Code 2.1.218 为最低能力基线，要求：

- Agent Teams 已启用；
- Agent/Task/SendMessage 能力可用；
- teammate 支持计划批准；
- SubagentStop 和 TeammateIdle 的真实 continue/block 能力必须由 doctor/system test 确认；若平台不支持，使用共享 checkpoint + 原 Assignment 重派，不模拟唤醒；
- PostToolUse 能观察 SendMessage；
- Hook 输入保留官方字段；
- named Agent 可被 Main Agent 定向消息和恢复。

平台拓扑还必须通过 doctor/系统测试确认以下行为，而不能靠名字猜测：

- 交互会话开启 Agent Teams 后，带 `name` 的 Agent 调用通常生成 teammate；使用调用级 `isolation` 或 fork 时仍是 Sub-agent；
- `isolation: worktree` 写在 Agent Definition 中，不保证该 Definition 作为 teammate 运行时仍有隔离；
- 自定义 Sub-agent 可被 SendMessage 寻址和恢复；内建 Explore/Plan 是 one-shot，不能作为需要恢复的 Worker；
- 交互环境的 fork mode 可能让 Sub-agent 实际在后台运行，Skill 必须读取调用返回的真实拓扑，不能假设“我要求 foreground 就一定阻塞”。

doctor 应检查这些能力。缺失时的处理是升级 Claude Code、启用配置或修复安装，不是切换到一套弱化的旧版流程。

官方能力来源：

- [Agent Teams](https://code.claude.com/docs/en/agent-teams)
- [Sub-agents](https://code.claude.com/docs/en/sub-agents)
- [Hooks reference](https://code.claude.com/docs/en/hooks)
- [Tools reference](https://code.claude.com/docs/en/tools-reference)

## 3. 派发前决策：先选执行拓扑，再写 Prompt

### 3.1 是否委派

Main Agent 按以下顺序判断：

1. 责任是否足够独立，能由一个明确的产物或结论收口；
2. 是否能给出真实文档入口、scope、依赖和 done_when；
3. 委派是否带来上下文隔离、专业化、并行或独立性收益；
4. 对普通 `bounded_flow` 责任，协调和回收成本是否低于 Main Agent 自执行成本；对 `coverage_complete` required responsibility，成本只影响拓扑和排队，不否决委派；
5. 是否存在必须由不同 Agent 背书的独立性要求。

以下情况默认不委派：

- 只需一次很小的局部修改；
- 责任与 Main Agent 当前工作高度交织，需要频繁来回；
- TASK 尚未自包含；
- 多人必须同时编辑同一表面；
- 仅为了满足“使用了 Sub-agent”的形式要求。

### 3.2 执行拓扑选择

| 场景 | 选择 |
|:--|:--|
| 单次、自包含、只需最终结果 | 普通 Sub-agent |
| 多个互相独立的一次性调查 | 多个后台 Sub-agent |
| 需要计划中间回执、持续执行、直接纠偏，且必须 worktree 隔离 | 可寻址自定义 Sub-agent + 调用级 `isolation: worktree` |
| 需要计划中间回执、持续执行、直接纠偏，且可按只读/路径分区 | Agent Team teammate |
| 多责任共享 Task DAG、需要团队消息 | Agent Team |
| 高风险且需要平台原生 Plan gate | Agent Team teammate + Plan mode |
| 需要独立审查/背书 | 不同 teammate 或不同 Sub-agent |
| 高频来回、上下文不可拆 | Main Agent 自执行 |

普通 Sub-agent 的 final response 是本轮结束。若业务需要“回计划后继续做”，PLAN_REPORT 必须在 Worker 仍运行时通过 SendMessage 发给 Main Agent，不能把计划写成 final response。若已错误结束，则用返回的 agent ID/name 恢复同一自定义 Sub-agent；恢复是异常路径，不是默认调度。

Agent Teams 打开时，`name`、fork 和调用级 `isolation` 会影响 Agent 调用最终落成 Sub-agent 还是 teammate。派发器必须记录平台返回的真实 topology。需要原生 worktree 隔离时优先 Sub-agent；需要共享 Task DAG/原生 Plan approval 时优先 teammate。两者都需要时，不能假装平台自动同时提供：应拆分写表面、让 Main session 自身运行在隔离 worktree，或选择更重要的约束并在 Assignment 中明示。

### 3.3 三种标准派发模式

| dispatch_mode | 语义 | 默认用途 |
|:--|:--|:--|
| one_shot | Worker 完成全部责任后一次返回结果 | 自包含研究、轻量检查、无需中间治理 |
| plan_checkpoint | Worker 先发送计划回执，不等批准，继续执行并最终交卷 | 默认 Builder、Verifier、QA、调查和修复任务 |
| plan_approval_required | Worker 进入原生 Plan mode，平台原生 Plan gate 通过后才能实施；是否需要 lead/人操作由 pin 的平台版本和模式决定 | 高风险、难回滚、scope 尚需确认的任务 |

不得再用“所有 specialized Agent 都走硬两阶段”作为统一规则。

派发身份是另一条低复杂度硬边界：manifest-draft 可以输出
`TODO(planner):agent-id-...` 作为 authoring 草稿，但该值不是 Agent identity。
readback 请求生成、`register-workgroup`、activation envelope、`runtime agent-event`、
PostToolUse sender binding、auto-chain 和 `runtime agent-begin` 只接受非空、无控制字符且非占位符的真实平台 ID；不强制
`agent-` 前缀。观察到非法可选身份时，PostToolUse 继续尝试已验证的
`teammate_name`，找不到则 fail-open 且不伪造绑定。

## 4. 风险分类：什么时候必须批准计划

### 4.1 默认采用 plan_checkpoint 的条件

以下条件同时成立时，计划是观测点而不是审批门：

- assignment scope 已由上游 TASK/责任固定；
- 修改发生在隔离 worktree；
- 改动可回滚；
- 不修改 locked truth；
- 不执行发布、部署或外部不可恢复动作；
- 实际 diff 会在集成时再次校验；
- Required Checks 已知；
- Main Agent 能在收到计划后通过 SendMessage 纠偏。

这种情况下，让 Worker 等待批准的收益低于 idle、kill、上下文切换和调度停顿成本。

### 4.2 触发 plan_approval_required

任一条件成立就使用原生 Plan approval：

- 数据库 migration、数据回填、不可逆数据变化；
- 公共 API、共享 schema、事件协议或跨模块核心接口变化；
- 权限、安全、密钥、支付、审计或合规边界；
- 会影响多个 active assignment 的共享写表面；
- 大范围删除、移动、生成或机械改写；
- 需要扩大已锁定 TASK 的 prospective scope；
- 任务定义仍有多个高成本解释；
- worktree 不能提供足够隔离；
- 失败后恢复成本显著高于一次计划审批等待。

### 4.3 风险决定由谁做

- 机械触发项由 dispatch planner 根据 TASK risk、路径和命令类型实算；
- Main Agent 可以升级为 approval_required；
- Worker 发现新风险时发送 BLOCKER/PLAN_ESCALATION，不得自降风险；
- 降级 approval_required 必须产生新的 Assignment revision；
- 人只在 REQ 变化、外部授权和不可派生业务决策上介入。

## 5. Assignment：调度治理的唯一工程记录

### 5.1 Assignment Record

每次派发必须有一个权威 Assignment Record：

| 字段组 | 必须表达 |
|:--|:--|
| identity | assignment_id、revision、task/responsibility ref、owner_agent_id、role |
| topology | dispatch_mode、实际 agent/team identity、foreground/background/team、resume address |
| authority | effective_scope、worktree/source/target；其他权限引用 Role/TASK，不复制全文 |
| plan | status、reported_at、digest、planned_paths、approval ref（若适用） |
| result | assignment status、result_ref、blocker_ref、integration/consumption ref |
| recovery | last checkpoint、replacement provenance |

Stage、generation、依赖、优先级、文档指纹、Closing Contract、Required Checks 和 expected outputs 已属于 TASK/责任的事实，Assignment 只保存引用或派发时指纹，不复制第二份可独立修改的内容。`dispatch_capacity_policy` 属于 Stage/ReviewPlan 权威；冲突、Ready/Queued/Dispatchable 和 active capacity 是调度器实算视图，不作为 Assignment 的第二套长期配置维护。

### 5.2 一个事实只维护一次

- Team task 从 TASK + Assignment 投影 subject、dependency、owner 和平台状态；
- spawn prompt 从 Assignment 生成；
- Agent Role 只提供静态最大能力；
- PLAN_REPORT 更新 Assignment 的 plan checkpoint；
- Completion Result 更新 result 状态；
- Runtime task/agent 列表是聚合视图；
- Gate 读取 Assignment/Result，不要求 Agent 手写重复 envelope。

所有更新以 `assignment_id + revision` 做 CAS；Worker worktree 只承载产品差异和必要的工作产物，不承载可被 merge 的活调度状态。

### 5.3 revision 规则

以下变化必须生成新 revision：

- Owner 变化；
- dispatch_mode 变化；
- effective_scope 扩大；
- TASK/规格指纹变化；
- Required Checks 实质变化；
- worktree/target 变化；
- 计划被判定为与 assignment 语义冲突。

纯文字澄清、顺序调整或不改变授权边界的纠偏可以使用 SendMessage，不必制造 revision。

### 5.4 统一工作状态与计划凭据

Assignment 的责任状态只使用 [状态机核心篇 §0.1](L4-state-transition-core.md#unified-work-state)
的 pending/active/blocked/done/cancelled。旧设计中的
`draft → ready → dispatched → running → result_submitted → consumed` 不再是一台
需要独立维护的主状态机：这些名称仅描述已存在的草稿、派发、运行、提交和消费事实。
ready 不另写状态，result_submitted 不能直接证明 done，consumed 须指明原消费者
及消费对象。替换由原新 attempt/撤销事实表达，不新增 replaced 工作状态。

计划的提交、漂移和批准由原 PLAN/approval 凭据验证，不另存计划状态或重复投影，
也不将其扩为 reading/approved/activated/working 工作状态。下列只说明兼容凭据结果，
不构成另一条需要 Agent 操作的生命周期：

```text
not_required
missing -> reported -> drifted
missing -> approval_pending -> approved
                            -> rejected -> approval_pending
```

- `plan_checkpoint`：reported 即满足首写屏障；drifted 只阻止首次写入或触发纠偏；
- `plan_approval_required`：approved 才满足写入屏障；
- 平台的 running/idle/stopped 是活会话事实，不回写成 Assignment 的同名伪状态；
- 平台 Team task 的 completed 是宿主字段，由 Main/scheduler 在原依赖消费条件
  满足后更新；不直接代表五态中的本人 done，也不把宿主字段名传播成另一套内部枚举。
- Worker 交付、Main 集成和阶段收口按不同责任分别投影；一个 Assignment 含多个
  checkpoint 时，当前动作必须明确是哪项责任，不把 PLAN 已接收当成修复已完成。

### 5.5 Canonical Result

每个执行型 Assignment 只交一份最小 Result：

| 字段 | 目的 |
|:--|:--|
| assignment_id / revision | 防止结果投错责任或消费过期结果 |
| outcome | completed / blocked / failed |
| output_refs | 指向代码、报告、finding、证据或其他真实产物 |
| actual_changed_paths | 与 effective_scope 做实际差异校验 |
| checks | 命令/用例、退出状态、evidence refs；不复制日志全文 |
| deviations / residual_risks | 暴露计划与实际差异及剩余不确定性 |
| handoff | 下一消费者需要执行的一个明确动作 |

Worker 仅写出 Result 文件不证明完成；原消费者验证交付及本人 required checks 后
才可将本人责任投影 done。Main 的集成/阶段吸收另有原消费事实与责任，依赖解锁仍须
满足其原条件。失败和阻塞也必须产出 Result/Blocker 引用，不能靠最后一条聊天猜测现场。

## 6. 调度模型：复用 TASK DAG，分离覆盖承诺与物理容量

### 6.1 一张权威图、一个即时冲突视图

统一调度器复用已有 TASK 依赖 DAG，不另建第二份依赖图。每次调度时只计算：

1. **依赖是否满足**：上游 Result 已 consumed，而不只是 Worker 声称完成；
2. **即时冲突集**：当前候选与 active assignment 是否在写路径、共享 schema、migration、生成文件、独立性或稀缺工具上冲突。

不持久化完整冲突图；只有真实冲突才在对应 Assignment/Blocker 上记录原因。这样保留可判定调度，又不引入需要持续同步的第二套项目图。

### 6.2 Ready 条件

assignment 只有同时满足以下条件才能进入 Ready Set：

- 上游 L3 已产生该责任所需的权威输入；
- 依赖已 consumed，而不只是 Worker 声称完成；
- 没有与 active assignment 的未解决冲突；
- Owner 和角色合法；
- dispatch_mode 已确定；
- worktree/只读边界可建立；
- 没有必须先由人裁决的问题。

满足这些语义条件即进入 Ready Set。平台并发槽是否空闲不改变 Ready，也不能使 Assignment 消失；调度器另计算 Dispatchable Set，槽位不足的 Ready Assignment 显示为 queued。

### 6.3 优先级

Ready assignment 默认按以下顺序：

1. 关键路径；
2. 解锁最多后续责任；
3. 高不确定性、早反馈价值高；
4. 可复用同一 Agent 当前上下文；
5. 稳定 ID 顺序，保证调度可复现。

### 6.4 Stage capacity policy

Stage/ReviewPlan 必须从两个策略中选择一个：

| policy | 适用 | 调度语义 |
|:--|:--|:--|
| `bounded_flow` | 普通 Builder、修复或协调成本可能高于并行收益的工作 | active slot 上限由真实平台槽位、明确资源约束和可观察的集成积压决定，不默认固定为 2；超出者保持 Ready/queued |
| `coverage_complete` | S7 等“漏掉一个独立验证面会制造后续返工”的质量发现工作 | required Assignment 的逻辑数量、Ready/queued 集合和 token 支出不设上限；所有独立 scope 必须保留，平台槽位释放后持续补位 |

两种策略都遵守：

- 同一即时写冲突集默认只允许一个 Assignment active；冲突改变顺序，不取消责任；
- 真实 platform capacity、账号、端口、数据集和 worktree 决定 Dispatchable Set；
- `coverage_complete` 下禁止通过 token budget、`max_reviewers` 或固定 WIP 把 required Assignment 改成 N/A/cancelled，或合并成认知过载 scope；
- `coverage_complete` 下仍须由初始 ReviewPlan 和一次受控 revision 闭合 required set；它不意味着每个 Result 都能继续产生新的 Assignment；
- `bounded_flow` 的 WIP 也是 active concurrency 上限，不是 Backlog/Ready coverage 上限；
- idle 不等于可以 self-claim 下一项，只有 scheduler 重新计算 Ready/Dispatchable 后才能派发。

资源锁释放时，Assignment 状态、Claim disposition 与 queued Agent 的唤醒必须在同一消费事务内完成：Assignment 从 `planned/queued` 变为 `dispatched` 后，Agent 从 `queued` 进入 `reading`，再由正常 PLAN_REPORT/activation 必经路径继续。queued Agent 在尚未释放前不应被 stop/idle 门要求提交计划；释放后也不能直接跳过计划检查。

## 7. 默认协议：双回执、单次连续执行

### 7.1 现实原型

该机制的现实原型不是“两次开工许可”，而是教师收两次材料：

1. 第一次收计划，确认学生理解题目和答题路线；
2. 第二次收最终试卷，检查实际结果。

学生交完计划后继续答题。只有计划跑题、漏题或遇到阻塞时，教师才打断。

### 7.2 正常时序

```mermaid
sequenceDiagram
    autonumber
    participant Main as Main Agent / Lead
    participant Tool as Dispatch + Runtime
    participant Worker as Addressable Sub-agent / Teammate
    participant Hook as Claude Hooks
    participant Consumer as Integrator / Reviewer / Gate

    Main->>Tool: dispatch ready assignment
    Tool-->>Worker: spawn prompt + Assignment ref
    Worker->>Worker: read TASK/responsibility -> contracts -> REQ -> design/rules
    Worker->>Main: SendMessage PLAN_REPORT (not final response)
    Hook->>Tool: PostToolUse records plan checkpoint
    Worker->>Worker: continue immediately; no approval wait
    Worker->>Worker: implement/review/test
    Worker->>Main: SendMessage BLOCKER/CORRECTION only when needed
    Main-->>Worker: SendMessage correction only on drift
    Worker->>Tool: submit canonical Result
    Worker->>Main: SendMessage COMPLETION
    Tool->>Consumer: integrate or consume Result
    Consumer-->>Tool: accepted/blocked
```

### 7.3 Worker 的硬指令

plan_checkpoint 模式的 spawn instruction 必须明确：

```text
Read the assigned authoritative documents before the first mutation.
Send one PLAN_REPORT to main/lead with the assignment ID/revision,
objective, planned paths, implementation/review steps, required checks,
dependencies, risks, and blockers.

PLAN_REPORT is a checkpoint sent through SendMessage, not your final response
and not the end of your turn.
Do not wait for approval. Continue the assignment immediately after sending it.
Stop only for a real blocker, an authority/scope conflict, or completion.
```

不能只写“先汇报计划”，否则模型很容易把汇报解释成 final response。

### 7.4 PLAN_REPORT 最小结构

| 字段 | 作用 | 消费者 |
|:--|:--|:--|
| assignment_id / revision | 防止计划投错责任或过期 | PostToolUse、Main Agent |
| objective | 暴露任务理解偏差 | Main Agent |
| planned_paths | 与 effective_scope 做机械子集检查 | Harness/Hook |
| steps | 让 Main Agent看见实现顺序和关键决策 | Main Agent |
| assertion_checks | 防止只写代码不证明 Closing Contract | Main Agent、Result validator |
| dependencies | 暴露错误 Ready 判断 | scheduler |
| risks / blockers | 触发纠偏、审批升级或 blocked | Main Agent |

计划不复制整份 assignment，不重复机器已经知道的 allowed/forbidden 列表。

### 7.5 计划检查

计划收到后分两层检查：

- **机器立即检查**：assignment/revision 当前、planned_paths 在 scope 内、Required Checks 未缺失、依赖声明一致；
- **Main Agent 异步判断**：目标理解、步骤是否偏题、是否漏关键集成边、是否触发高风险审批。

机器检查失败时不允许第一次写入。Main Agent 语义判断正常时不回复；发现偏移时通过 SendMessage 发送 CORRECTION。

### 7.6 一次性首写屏障

计划检查点应埋在第一次写入的必经之路：

1. Worker 通过 SendMessage 发送稳定类型的 PLAN_REPORT；
2. PostToolUse(SendMessage) 验证目标是 main/lead、assignment/revision 正确；
3. Harness 将 digest 和 reported_at 原子写入 Assignment；
4. Worker 首次调用 Write/Edit/NotebookEdit/写入型 Bash；
5. PreToolUse 读取 Assignment；
6. plan_checkpoint 模式且 plan.status != reported 时，拒绝本次写入并告诉 Worker先发送计划；
7. plan 已记录后不再重复打扰。

这是“先形成计划”的自助屏障，不是“等 Main Agent 批准”的停工门。

## 8. 高风险协议：原生 Plan approval

### 8.1 时序

```mermaid
sequenceDiagram
    participant Main as Lead
    participant Worker as Teammate in Plan mode
    participant Platform as Claude Code native Plan gate
    participant Host as Host adapter
    participant Runtime as Assignment Runtime

    Main->>Worker: dispatch approval_required assignment
    Worker->>Worker: read-only investigation and plan
    Worker->>Platform: native plan approval request
    Platform->>Host: native decision observation, when exposed by this platform version
    Host->>Runtime: consume observation; validate current Record and exact plan
    alt native gate approved
        Runtime->>Runtime: append bound platform receipt; plan.status=approved
        Worker->>Worker: exit Plan mode and execute
    else native gate rejected
        Runtime->>Runtime: append bound rejection receipt; plan.status=rejected
        Worker->>Worker: remain in Plan mode and revise
    end
```

### 8.2 与默认模式的区别

| 维度 | plan_checkpoint | plan_approval_required |
|:--|:--|:--|
| 计划作用 | 早期观测 | 执行授权 |
| 发送后 | 立即继续 | 等待平台原生 gate 结果 |
| 写入屏障 | 只要求计划已发送 | 要求计划已 approved |
| Main Agent 正常动作 | 静默 | 消费实际原生 gate 结果；需要时纠偏，不自行伪造批准 |
| 适用成本 | 低 | 高，但由高风险定价 |

`plan_approval_required` 这一名称只说明平台原生 Plan gate 是执行前置条件；平台可能
自动推进该 gate，也可能要求 lead 操作，具体依赖 pin 的平台版本和模式。无论哪种情况，
Runtime 都只记录实际观察到的平台结果；平台自动通过不是人类审批。

### 8.3 禁止伪造批准

- teammate 或其他 Agent 的消息不能充当用户权限批准；
- Main Agent 的自然语言回复只能纠偏，不能自行成为平台批准 receipt 或人类决定；
- REQ 变化和外部不可派生授权仍交人；
- approval 绑定 assignment revision，revision 变化后自动失效。

### 8.4 原生 Plan gate 的来源与 Record 绑定

`PLAN_APPROVAL_REQUEST` 是 Worker 提交的计划请求，不是批准。`plan_approval_required`
的权威来源必须是 Claude Code/Agent Teams 原生协议产生、由 Host adapter 直接观察的
批准或拒绝事件；普通 SendMessage 文本、Worker 自写文件、CLI `--approved-by`、
Runtime `Actor`、session ID 或手工填写的 evidence ref 都不能伪造该事件。平台事件若
没有稳定的原生 request/event identity、来源身份和对应计划内容，Harness 不得生成
“已批准”记录。

平台 Plan gate 与 human approval 是不同事实。平台行为随版本和会话模式核实；若原生
teammate plan request 在该版本/模式下由平台自动放行、没有 lead review 或人类提示，
receipt 只能表示 `platform_plan_gate=passed`，不能表示 `human_approved`。需要真正
人类签字的路径必须捕获平台提供的、可验证的显式人类决定事件；当前 Runtime 没有把
普通 CLI 字段或对话文本变成该信号的能力。适配器不能观察到可信来源时，canonical
高风险 Assignment 保持 `approval_pending` 且所有写入被拒绝。

对 canonical Assignment，Writer 消费的原生 Plan gate receipt 至少绑定：

- namespace 与身份：`runtime_id`、`baseline_generation`、`assignment_id`、当前
  `assignment_revision`、当前 `Assignment RecordReference + SHA256`；
- 新版本的精确计划：当前 PLAN_REPORT ref/SHA、平台原生 approval request/event ID
  与该平台事件所批准的计划摘要；
- 平台观察上下文：由注册/Host adapter 核实的 Team、teammate/worker 会话身份，
  原生事件中的 decision、source kind 和平台版本。`human_decision` 另需平台实际
  认证的人类决策来源，不能由上述平台 gate 字段推导。

Host adapter 只把原生观察交给既有 Runtime Writer；它不单独改 assignment、activation
或 lifecycle。Writer 在同一 CAS mutation 中重算当前 Record 和 PLAN ref/SHA、检查
dispatch mode 与身份完全相符，再 append immutable receipt 并更新正交的 plan status。
重复同一平台 event ID 只返回原 receipt；异 event/异绑定冲突拒绝。拒绝结果保持在
Plan mode 并等待 Worker 改计划；计划字节变化、Record revision 变化、owner 或
namespace 变化都令旧 receipt 不再满足当前写屏障。首次 PreToolUse 及后续写入都要
检查当前 Record/PLAN 与该 receipt 的完整绑定；原生平台的只读 Plan 限制仍须保留，
Runtime receipt 不可代替平台本身的权限控制。

现有 legacy `understanding_approved` 不能给 canonical revision 供证：
`internal/assignment/lifecycle.go` 只要求 Agent 从 `understanding_submitted` 转移，
输入仍是 Worker 的 `readback_response`，Runtime mutation 的 Actor 固定为
`orchestrator`；`internal/agentauth/activation.go` 的 hash chain 只验证 Worker readback
字节 SHA 和 `message_id`。activation schema 中的 `approval_evidence_id` 不是 Writer
消费的批准源。不能把该旧状态、CLI `approved-by`/Actor/session 或 Worker-authored
evidence 回填成新 Record 的批准。

最小可落地顺序：先针对仓库声明的 Claude Code 版本验证 Host 能否取得并关联原生 Plan
decision、具体 request/计划内容和实际 source identity；其后接一个只接受该 Host
observation 的适配器与原 Runtime Writer receipt 消费路径，再把首写 Hook 接到精确
Record/PLAN receipt 校验。若该版本只提供平台自动 gate，则仅接入平台 gate receipt，
需要人签字的 assignment 继续 fail closed，等待独立、可验证的人类决定来源。接线前
只开放 `plan_checkpoint` canonical revision；高风险 Assignment 可停在只读规划阶段，
不得用 legacy activation 或手工证据开启写入。

## 9. 等待、消息和执行连续性的精确语义

### 9.1 普通 Sub-agent

| 运行方式 | Main Agent 行为 | 结果语义 |
|:--|:--|:--|
| 实际 foreground | Agent 调用阻塞到 Sub-agent 本轮完成 | 只适合必须先拿最终结果再继续；交互 fork mode 下不能假定一定 foreground |
| 实际 background | Main 继续其他工作，最终结果通过 completion notification 到达 | 适合独立并行任务，不轮询 transcript/task row |
| running addressable agent | Main/Worker 可通过 SendMessage 交换 PLAN_REPORT/CORRECTION | plan_checkpoint 的正常中间回执链 |
| stopped addressable agent | Main 用 SendMessage(agent ID/name) 恢复同一上下文，平台在后台继续 | 仅用于误停、追加责任或纠偏，不作为正常计划链 |

Sub-agent 的 canonical final result 仍通过 Agent completion 返回，但计划可以在运行中通过 SendMessage 到达 Main Agent。需要中途计划回执时必须使用可寻址的自定义 Sub-agent；内建 Explore/Plan 不返回可恢复 ID，不适用。若 Sub-agent 只把计划作为 final response，SubagentStop 应先阻止结束；已经结束时恢复同一 agent，不能新建一个“第二轮 Agent”。

### 9.2 Agent Team teammate

- teammate 通过 SendMessage 主动回报计划、blocker 和完成；
- 消息自动到达 lead，lead 不轮询 mailbox；
- teammate 发送 PLAN_REPORT 后继续当前 turn；
- lead 可在 Worker 运行中发送 CORRECTION；
- teammate idle 仍可被定向消息唤醒，但正常路径不依赖先 idle 再唤醒；
- Team task 是调度视图，不是完成权威。

### 9.3 Main Agent 如何等待

Main Agent 不进入“等待计划后第二次授意”的 activation barrier。

它遵循：

1. 派发后继续调度其他 Ready responsibility；
2. 不自执行已经委派的同一责任；
3. 收到 PLAN_REPORT 时快速做偏移判断；
4. 正常计划不回复；
5. 没有其他 Ready 工作时，明确等待 teammate 完成，而不是宣布结束；
6. 收到 completion 后进入消费/集成；
7. 只有 blocker、drift、approval request 或 completion 需要产生响应。

Plan mode 属于 approval_required Worker，不属于 lead 的普通等待状态。

把“需要收到什么”直接映射到平台等待方式：

| Main 需要收到 | 派发/等待方式 | 禁止做法 |
|:--|:--|:--|
| 同步最终结果 | 实际 foreground Sub-agent；Agent 调用本身阻塞 | 再造一层轮询 |
| 并行最终结果 | background Sub-agent；等待 completion notification | 读 transcript 猜完成、因 UI row 消失而重派 |
| 计划 + 最终结果，计划不需批准 | 可寻址 Sub-agent 或 teammate；PLAN_REPORT 用 SendMessage，Main 继续调度，最终结果走 completion/COMPLETION | 把计划当 final、收到计划后再发“可以开始” |
| 必须经过原生 Plan gate 后执行 | teammate Plan mode；等待原生 PLAN_APPROVAL_REQUEST 和平台 gate 结果 | 用普通文本假装 approval protocol |
| 已误停 Sub-agent 的第二次授意 | SendMessage 到原 agent ID/name，自动恢复后台执行 | 再次 Agent 调用创建新上下文 |

因此不存在一个通用的“主会话等待模式”。foreground 的调用本身就是等待；background completion 和 teammate message 都由平台投递。调度 Skill 要做的是保持 assignment 未收口、继续处理其他 Ready 工作，并在没有本地工作时等待平台通知，而不是轮询、结束总任务或让 Worker 先 idle。

### 9.4 消息类型

统一使用：

| 类型 | 方向 | 是否终止当前工作 |
|:--|:--|:--|
| PLAN_REPORT | Worker → Main | 否 |
| PLAN_APPROVAL_REQUEST | Worker → Main | 是，停在 Plan mode |
| PLAN_APPROVAL_RESPONSE | 原生平台/Lead session → Worker | 否，只有平台原生 gate 结果通过才继续 |
| CORRECTION | Main → Worker | 否 |
| BLOCKER | Worker → Main | 取决于是否仍有可做工作 |
| PROGRESS | Worker → Main | 否，默认不强制 |
| COMPLETION | Worker → Main | 是，进入 Result 消费 |
| SHUTDOWN_REQUEST | Main → Worker | 是，仅团队收口 |

不得为每个生命周期微状态创建一条强制消息。

## 10. Hook 与 Harness 的真实控制契约

### 10.1 Hook 事件地图

| 事件 | 真实责任 |
|:--|:--|
| PreToolUse(Agent/Task) | 派发前检查 Assignment、dispatch_mode、worktree、stage capacity policy、Dispatchable slot、冲突和角色；不得因槽满删除 queued responsibility |
| SubagentStart | 注入唯一匹配 Assignment 的真实 scope、`done_when`、Required Checks、连续执行语义和最小上下文 |
| PostToolUse(SendMessage) | 捕获 PLAN_REPORT/BLOCKER/COMPLETION，更新 Assignment checkpoint |
| PreToolUse(Write/Edit/Bash...) | 首写计划屏障、locked artifact、不可逆动作、scope 前置诊断 |
| PreToolUse(TaskUpdate) | owner/status 变化必须对应有效 Assignment 和 scheduler 决策；阻止 teammate 擅自 self-claim 下一责任 |
| TeammateIdle | 在平台支持真实 continue/block 控制且 doctor 已确认时，阻止未完成 teammate 进入 idle；只处理当前 assignment |
| SubagentStop | 在平台支持真实 stop decision 且 doctor 已确认时，阻止普通 Sub-agent 以计划代替完成，或缺 Result 时停止 |
| TaskCompleted | 防止 Team task 在 canonical Result 被消费前标 completed |
| SessionStart | 从 Assignment/Result 重建调度板，识别失联 Worker |
| PreCompact | 确保 Assignment、plan、result 和 resume address 已持久化 |

### 10.2 TeammateIdle 决策矩阵

Hook 必须使用官方 teammate_name、team_name、transcript_path 定位对象，不依赖平台不会提供的自造 agent_id。

| 当前事实 | Hook 行为 |
|:--|:--|
| 无匹配 Assignment | 允许 idle，记录诊断；不得猜第一个 idle Agent |
| plan_checkpoint 且计划未报告 | exit 2：要求发送 PLAN_REPORT 后继续 |
| plan_approval_required 且等待批准 | 允许保持平台 Plan waiting；不得改成普通 active |
| 计划已报告、assignment 未完成、无 blocker | exit 2：继续当前 assignment |
| blocked 但没有 BLOCKER | exit 2：先提交 blocker |
| blocked 且 BLOCKER 有效 | 允许 idle |
| Result 缺失但 Worker 声称完成 | exit 2：提交 canonical Result |
| Result 有效、等待消费 | 允许 idle；不自动分配下一项 |
| assignment consumed | 允许 idle；scheduler 可另行派发 |

exit 2 的 stderr 必须是给 teammate 的一条具体下一步，不能只输出 LOOP RECOVERY 长文。

`exit 2` 和 `decision=block` 是目标平台控制，不得通过自造字段模拟。落地前必须用最新 Claude Code 的官方 payload 做 doctor/system test，确认反馈会回到同一活会话。若平台能力缺失或 payload 无法定位原 Agent，Hook 不得假装完成唤醒：持久化 Assignment checkpoint，报告一个可恢复 blocker，并由 scheduler 依据原 Assignment 重派。禁止因为一个未验证的平台假设再增加一套“虚拟 active”状态。

### 10.3 SubagentStop 决策矩阵

| 当前事实 | Hook 行为 |
|:--|:--|
| final response 只有计划，assignment 要求执行 | decision=block：计划不是完成，请继续 |
| assignment 未完成且无 blocker | decision=block：继续当前责任 |
| blocked 但无 BLOCKER | decision=block：先报告 blocker |
| canonical Result 缺失 | decision=block：提交 Result |
| Result 有效 | 允许 stop，父会话进入消费 |

必须使用 stop_hook_active 避免 Hook 自身造成无限停止循环。

### 10.4 Hook 输出纪律

- 要让 teammate 继续，必须使用 Claude Code 支持的 exit 2 或 decision=block；
- systemMessage 是用户提示；Agent 指引使用接线篇规定的 `hookSpecificOutput.additionalContext`；Stop/SubagentStop 上该字段会续跑，不用于回收软提醒；
- exit 2/decision=block 只有在官方 payload doctor 通过后才作为强制控制；未通过时走 checkpoint + replacement，不制造假唤醒；
- Runtime 状态变化不等于平台会话已唤醒；
- Hook 只判机械事实和控制下一步，不判断计划的业务质量；
- 所有官方输入字段必须无损进入 adapter；
- 测试必须使用官方 payload，禁止靠额外 agent_id 让测试通过。

### 10.4.1 临时 worktree 的调度责任

派发与交付消费 [L4 Worktree](L4-worktree-governance.md) 的唯一生命周期。主会话从 REQ 开发绑定解析创建基线，确保上游正式产出已提交；运行 evidence 按声明交接，不复制整个控制面。Assignment 记录 source commit 与接收目标。

子会话提交成果并报告，主会话在项目根目录合并、联合验证、接收和清理。SubagentStop 不运行完整测试/长合并，不要求已停止的子会话再次 stop 才完成清理；隔离中的子会话不承担根目录 merge。未回收与积压只产生给主会话的软提醒，不增加调度 stop/idle 矩阵的 block 条目。

独立启动的已登记 Worker 顶层会话同样按本人矩阵处理普通 Stop：原
session/root 登记解析 owner，再校验当前 execution、精确 CWD 和 Bootstrap；
本人 PLAN_REPORT/Result 缺失继续原阻断，结果已记录或原合法结束事实成立则返回。
此路径不运行 Main 的 Stop 队列或全局控制循环，不因待集成或同事未交重新催本人。
报告存在仅代表 delivery_recorded；required checks、领域接收、原返工与人闸仍由
原消费者决定。旧 blocked/终态兼容允许结束不伪装成有效证据。

### 10.5 指引应埋在哪里

`agent-protocol.md` 只保留调度入口、三种 dispatch_mode 和“哪里查权威”的短索引。持续执行所需指引分布在 Worker 的必经路径：

| 载体/时机 | 埋入的唯一内容 | 为什么在这里 |
|:--|:--|:--|
| TASK / Closing Contract | responsibility、输入、done_when、Required Checks | 没有它就无法完成任务 |
| Assignment generator | owner、真实 topology、scope、dispatch_mode、revision | 派发时必须生成，避免 Prompt 自由发挥 |
| Agent Definition | 角色最大工具/写路径/专业方法边界 | 定义长期稳定能力，不复制单任务事实 |
| dispatch Skill | 拓扑选择、风险分类、等待语义、异常恢复 | 只在 Main 准备委派时加载 |
| spawn prompt | Assignment ref、最小读序、“PLAN_REPORT 不是 final，发送后继续” | Worker 第一轮必然接收 |
| PLAN_REPORT schema | objective/paths/steps/checks/dependencies/risks | 填写动作本身暴露理解偏差 |
| PostToolUse(SendMessage) | 记录计划/阻塞/完成 checkpoint | 消息发送必经路径 |
| 首次 mutation 的 PreToolUse | 缺计划、revision 过期、approval 缺失的一个下一步 | 开工必经路径，只触发一次 |
| PreToolUse(TaskUpdate) | 禁止未授权 self-claim/越权完成 | Team task 状态变化必经路径 |
| TeammateIdle / SubagentStop | 计划不能当交卷、缺 Result 不能结束 | Worker 准备停下的最后必经路径 |
| Result schema + consumer | actual diff、checks、deviation、消费确认 | “做完”变成下游可消费事实 |
| SessionStart / PreCompact | 恢复地址、未消费结果、单一下一步 | 跨会话连续性的必经路径 |

同一句规则只能有一个主承载点。其他位置最多放链接和一句触发提示，不能复制成长篇协议；错误反馈只给当前缺失事实和下一动作，不把整份 L4 注入 Worker。

## 11. Main Agent 调度循环

Main Agent 在任何 spawn、消息、idle、stop、result 或 SessionStart 后运行：

```text
DISPATCH_LOOP()
1. Read current Stage and authoritative Assignment inventory.
2. Read dispatch_capacity_policy; recompute DAG, conflicts, Ready Set, queued set, platform slots and Dispatchable Set.
3. Consume inbound PLAN/BLOCKER/COMPLETION messages.
4. For PLAN_REPORT:
   - run mechanical checks;
   - judge semantic drift;
   - stay silent if aligned;
   - send CORRECTION or revise/reassign if drifted.
5. For BLOCKER:
   - continue all unaffected work;
   - resolve locally, revise assignment, route spec change, or ask human only at a real gateway.
6. For COMPLETION:
   - validate Result;
   - integrate or pass to the stage-specific consumer;
   - mark consumed only after that consumer succeeds.
7. Dispatch the highest-priority Dispatchable assignment; on coverage_complete keep every other required Ready assignment queued and refill all free slots.
8. If active workers remain and no Dispatchable work exists, wait; do not cancel Ready/queued work.
9. Close teammates only after all their assignments are consumed or intentionally cancelled.
```

Main Agent 不能：

- 在 Worker 回计划后机械回复“批准，可以开始”；
- 因 teammate 暂时没有新消息就重派；
- 抢做 active assignment；
- 在 TeammateIdle 中直接分配未重新求值的下一 TASK；
- 把 Team task completed 当作 Stage gate；
- 因协调成本高而静默缩小任务范围。

## 12. 中断、偏移与恢复

### 12.1 计划缺失

- 首写屏障拒绝第一次 mutation；
- TeammateIdle/SubagentStop 阻止以无计划或只有计划的状态结束；
- 反馈只要求发送 PLAN_REPORT 并继续；
- 不需要 Main Agent 发第二轮 activation。

### 12.2 计划偏移

偏移分三级：

| 级别 | 示例 | 处理 |
|:--|:--|:--|
| 可纠偏 | 顺序不佳、漏一个 owned check | SendMessage CORRECTION，继续同一 revision |
| 授权变化 | 新增写路径、改 Required Checks、触发高风险 | 新 Assignment revision，必要时进入 Plan approval |
| 规格冲突 | TASK/contract/REQ 无法共同成立 | 停止对应责任，走 L3 的 spec-change/human route |

Main Agent 纠偏后不需要重新讲完整任务，只发送差异和原因。

### 12.3 执行偏移

- planned_paths 只是早期预测；
- 实际 changed paths 由 git diff 或消费者实算；
- scope 越界在 integration/Result consumer 阻断；
- worktree 保留所有偏移现场；
- 可回滚的普通偏移优先修回，不立即 kill/reassign；
- 只有 Owner 不再可信、上下文损坏或平台 Agent 不可恢复时才替换。

### 12.4 idle、hidden、stopped 与 killed

必须区分：

| 平台现象 | 含义 | 恢复 |
|:--|:--|:--|
| idle | 当前 turn 结束，teammate 仍可寻址 | 官方 continue/block 已验证时用 SendMessage/exit 2；否则保留 checkpoint 并由 scheduler 恢复/重派 |
| hidden | UI 隐藏 idle row，不等于会话销毁 | 按名称/ID 消息唤醒 |
| completed / TaskStop stopped | Sub-agent 本轮结束但仍可寻址 | SendMessage resume 原 agent ID/name，或消费结果 |
| user-cancelled / killed | 用户在 UI 取消，或平台强制终止；SendMessage 可能拒绝自动恢复 | 保留现场；能由用户重开则恢复，否则按 checkpoint 重派 |
| shutdown | teammate 被正常关闭 | 仅在 assignment consumed/cancelled 后允许 |
| session lost | 平台上下文不可恢复 | 从持久 Assignment/plan/result/checkpoint 重派 |

### 12.5 重派

重派必须记录：

- replacement_of_agent_id；
- 原 assignment/revision；
- 新 Owner；
- 已完成计划和工作摘要；
- worktree/checkpoint 接管方式；
- 为什么不能继续原 Agent；
- 独立性约束是否仍成立。

禁止因一次 idle 就创建替代 Agent。

## 13. 跨 L3 消费地图

| L3 Stage | 典型责任 | 默认拓扑/模式 | Stage 特有消费 |
|:--|:--|:--|:--|
| S0 需求设计 | 辅助研究、方案对抗 | one_shot；Main 保持需求责任 | 结果仅供人和 Main 形成 REQ，不得替人锁定 |
| S1 绑定 | 通常不委派 | Main | 生命周期授权不能由 Sub-agent 执行 |
| S2 设计 | 模块设计、原型研究、技术调查 | plan_checkpoint Sub-agent/teammate；独立研究可 one_shot | 设计包/ADR 被 S2 收口 |
| S3 契约 | FE/BE/SYNC 分面起草与对账 | Agent Team + plan_checkpoint | 契约锁定前由 Main/Stage gate 聚合 |
| S4 TASK 拆分 | 模块拆分、DAG/Closing Contract 检查 | one_shot 或 plan_checkpoint | TASK batch、DAG 与整体派发计划交付 S5 |
| S5 文档验证 | 两个独立 reviewer responsibility | Agent Team；read-only plan_checkpoint | 两份独立结论和包含派发计划的 exact subjects |
| S6 构建 | frontend/backend/test Builder | 隔离写入优先 Sub-agent + worktree；路径可分时可用 teammate；默认 plan_checkpoint，高风险 approval | 批准计划的实时投影 + Builder Result + integration |
| S7 完整验证 | DV/QA/E2E Claims 与独立验证面 | ReviewPlan DAG + 1..N Assignments + plan_checkpoint | 当前 round 的 ReviewResult、Finding 和 CleanRound/ObservationBatch |
| S8 调查 | 多假设调查、Original Finder 责任 | 多 teammate plan_checkpoint；Investigator lifecycle bridge 已接入 Runtime，通用 PLAN_REPORT 按 fingerprinted manifest 绑定 | InvestigationCase/HypothesisResult/RepairContract；canonical BUG 仅为批准后的兼容投影；实际平台进程启动仍由 Claude/Agent Team 负责 |
| S9 修复 | repair Builder、定向复验 | 隔离写入优先 Sub-agent；默认 plan_checkpoint，高风险 teammate approval | fix Result、影响失效、targeted reverify |
| S10 验收审计 | 证据汇编、独立系统审计 | read-only teammate/one_shot | ACC 与 audit 结论 |
| S11 发布闸 | 发布前信息整理 | read-only one_shot 可辅助 | 最终 release decision 仍由人 |

所有 L3 可以增加 Stage 特有 done_when，但不得重定义：

- PLAN_REPORT 是否终止 turn；
- TeammateIdle 和 SubagentStop 的平台语义；
- Assignment 生命周期；
- dispatch_mode；
- plan approval 风险规则；
- Main Agent 的等待和纠偏方式。

## 14. L1 准则与现实原型映射

| L1 准则 | L4 落点 |
|:--|:--|
| D1 权威外置 | Assignment/Result/checkpoint 持久化；平台会话只承载活执行 |
| D2 自然路径观测 | SendMessage、首写、idle、stop、task complete、session start 都有控制点 |
| D3 门是顾问 | 默认计划不等批准；错误反馈给一个可执行下一步；高风险才停 |
| D4 引导性产物 | PLAN_REPORT 的少量字段迫使 Worker 暴露理解、步骤和验证计划 |
| D5 三级强制 | Prompt 引导 → 结构化消息/Assignment → 首写/stop/idle 机械控制 |
| D6 三方收敛 | Worker 判断并执行，机器公证 scope/状态，Main/独立角色消费结论 |
| D7 收敛可观测 | Ready/running/reported/consumed + plan/result checkpoint 显示单调进度 |
| 公理一 原型 | 教师先收答题计划、后收试卷；工程上对应 kickoff plan + implementation handoff |
| 公理二 分工 | Main 调度和判偏移，Worker 执行，Hook 判机械事实，人握需求/发布 |
| 公理三 消费 | 每个 plan/result 字段都有 scheduler、Main、consumer 或 Gate 消费者 |
| 公理四 成本 | 普通任务取消审批等待，高风险任务保留定价后的保险 |
| 公理五 传达 | 纠偏原因随 CORRECTION，Hook 失败随下一步到达 Worker |

## 15. 跨阶段消费与实现边界

当前实现以代码、回归测试及 [L4 运行时控制面消费地图](L4-runtime-control-plane.md#14-跨-l3-消费地图) 为准。平台进程的启动、唤醒和恢复必须通过真实平台验证，不能仅凭 Runtime 状态或合成 payload 宣称成功。

调度层的统一闭环只覆盖“谁负责、何时可写、何时可停、如何恢复”；它不替 S7/S8/S9 做业务判断。跨阶段当前结论如下：

| 边界 | 调度层已经保证 | 仍不能声称已经保证 |
|:--|:--|:--|
| S7 → S8 | ReviewResult/ObservationBatch 的 Assignment、消息、CAS、exact-set 消费和 S8 入口指引；Investigator 已由专用 workgroup/Task/Agent/activation bridge 登记，通用 PLAN_REPORT 可按 manifest 绑定 | Claude/Agent Team 的实际进程启动仍是平台动作；Runtime 只登记控制面和恢复事实，不伪造外部 spawn 成功 |
| S8 → S9 | approved RepairContract 才能打开 S9；Contract hash/revision 进入 Session/Plan；通用 PLAN_REPORT 可按 fingerprinted manifest 校验 | BUG projection、通用 PLAN_REPORT 和 S9 domain PlanReport 不自动互换；它们仍是不同契约，领域报告必须单独提交 |
| S9 → S7 | Handoff、baseline、seed、round+1 和清理旧投影在 Runtime CAS 中形成可恢复事实 | seed 已等同完整 ReviewPlan；S9 的 `depends_on/resource_locks` 已被调度器执行；这些仍需最终 S7 注册/消费或后续专门实现 |

调度机制的复杂度边界保持不变：保留首写屏障、PLAN_REPORT、Stop/Idle 控制、Assignment/CAS 和 exact-set 消费，因为每一项都有实际消费者和恢复动作；不新增第二套 Investigator lifecycle、CaseSet 或 S9 scheduler，除非先补齐明确的权威对象、next command、失败恢复和回归测试。涉及两个计划回报时，平台通用 `agent-message.plan_report` 只证明 Agent 已理解并可继续，S9 `repair-plan-report` 才证明领域执行计划已提交；两者不能用文档简称合并。

## 16. 验收测试、指标与 Definition of Done

### 16.1 必须有的系统测试

| 场景 | 期望 |
|:--|:--|
| Sub-agent/teammate 发送 PLAN_REPORT 后继续调用工具 | 不进入 idle/stop，不需要 Main approval |
| plan_checkpoint 首次写入前没有 PLAN_REPORT | PreToolUse 拒绝并指导发送计划后重试 |
| PLAN_REPORT assignment/revision 错误 | 不记录 checkpoint，拒绝首写 |
| planned_paths 超 scope | 计划标 drifted，Main 收到纠偏项 |
| teammate 在计划前 idle | TeammateIdle exit 2，反馈送回同一 teammate |
| teammate 计划后、完成前 idle | exit 2，继续当前 assignment |
| teammate blocked 且 blocker 有效 | 允许 idle，Main 收到 blocker |
| teammate Result 有效 | 允许 idle，等待 consumer，不自动分配下一项 |
| 普通 Sub-agent 只返回计划 | SubagentStop block，要求继续工作 |
| 普通 Sub-agent 完成并有 Result | 允许 stop，父会话收到结果 |
| approval_required 未批准就写 | 拒绝 |
| approval_required 批准后写 | 允许，approval 与 revision 对应 |
| Main 运行中发送 CORRECTION | 同一 Agent 接收并继续，不重派 |
| Sub-agent 把计划作为 final response | SubagentStop block；若已结束则 SendMessage 原 ID 恢复，不新建 Agent |
| 开启 Agent Teams 后使用调用级 isolation | 真实 topology 记录为 Sub-agent，并得到 worktree 隔离 |
| 隔离 Sub-agent 发送 PLAN_REPORT | Main 与 Worker 读取同一共享 Assignment checkpoint；worktree 中不产生第二份运行态 |
| teammate 尝试 self-claim 未授权 Task | PreToolUse(TaskUpdate) 拒绝，等待 scheduler 派发 Assignment |
| session 恢复时平台 Agent 不存在 | 从 Assignment/checkpoint 生成明确 replacement，不伪称 resume |
| 两个 assignment 写冲突 | 第二个不进入 Dispatchable，保留责任并在锁释放后重新求值 |
| `bounded_flow` active slots 已满 | 新 Assignment 保持 Ready/queued，不被取消；槽释放后派发 |
| S7 `coverage_complete` 有 20 个 required Ready、平台 6 个槽 | 6 个 Dispatchable/running、14 个 queued；持续补位且 exact set 不减少 |
| `coverage_complete` 配置 token budget/max_reviewers 试图裁 required Assignment | schema/dispatch gate 拒绝；成本只可观测，不可改变 required set |
| 使用官方 TeammateIdle payload、无 agent_id | 精确定位 teammate，不猜测 |

### 16.2 运营指标

- dispatch 到 PLAN_REPORT 的时间；
- PLAN_REPORT 到首次 mutation 的时间；
- plan_checkpoint 被 Main 判偏移的比例；
- 因缺计划触发首写屏障的比例；
- 计划后误 idle/stop 的比例；
- 同一 Agent 恢复成功率；
- 因错误重派损失的 token 和时间；
- 每 assignment Main Agent 消息次数；
- approval_required 占比及实际捕获风险数；
- stage capacity policy、Ready/queued/Dispatchable/active 数、cycle time、lead time；
- coverage_complete required Assignment retention，目标 100%；
- 因平台容量排队的时间与因 scope 冲突排队的时间；
- plan 字段中没有消费者的比例；
- Runtime 与 Claude Code 平台状态不一致次数。

### 16.3 机制 Definition of Done

只有以下条件全部成立，本 L4 才算落地：

- 普通任务不再需要 readback 后第二次授意；
- Worker 发送计划后在同一 turn 连续执行；
- Main Agent 正常计划静默，只有例外才打断；
- 高风险任务能稳定进入原生 Plan approval；
- TeammateIdle/SubagentStop 使用真实平台控制；
- Assignment 是 Owner、scope、plan、result 和 recovery 的唯一权威；
- 隔离 worktree 与 Main 共享同一调度控制面，运行态不经 git merge 汇总；
- Team task 与 Runtime 是一致投影；
- 没有 Agent 因一次 idle 被无因重派；
- 首写、完成、集成/消费和 Stage Gate 形成可判定闭环；
- 所有受影响 L3 引用同一 L4，不再复制各自的派发协议；
- `bounded_flow` 与 `coverage_complete` 可由 Stage 明确选择；物理容量只影响 Dispatchable/queue，不会裁掉 coverage-critical required Assignment；
- 系统测试使用 Claude Code 2.1.218 官方 payload；
- 指标能证明等待减少而偏移、越界和返工没有恶化。

## 变更记录

| 日期 | 版本 | 变更 | 依据 |
|:--|:--|:--|:--|
| 2026-08-20 | v0.3.0 | 增加机制准入规则；明确 Claims/Assignment/Result 为事实，coverage/ledger/board 为视图；限制 `coverage_complete` 不得递归扩张；补充 exit 2/decision=block 必须通过官方 payload doctor，失败时走 checkpoint + replacement | 终审发现调度设计仍可能引入重复控制面，并且不能把未验证的平台唤醒能力写成既成事实 |
| 2026-08-20 | v0.2.0 | 新增 `bounded_flow / coverage_complete` Stage capacity policy；将 Ready 与 Dispatchable/physical slots 分离；S7 coverage-critical required Assignments 不受固定 WIP、Reviewer 数量或 token budget 裁剪，容量不足只进入 queued | S7 的 DV/QA/E2E 完整发现属于必要质量投入；通用 L4 若仍强制 WIP=2，会让 L3 设计无法落到真实 Agent/Task 派发必经路径 |
| 2026-08-20 | v0.1.0 | 建立第一份 L4：统一 Sub-agent/Agent Team 派发模式、双回执连续执行、风险 Plan approval、Hook 控制、恢复与跨 L3 消费地图 | owner 指示：L4 为横跨多个 L3 的抽象工具层；两轮对话应改为计划回执后连续执行 |


### 2026-09-18 · v0.4.0

调度消费临时 worktree 机制，由主会话完成接收与清理，回收仅软提醒。


## 共享模型与合同依赖（2026-09-19）

TASK 是角色阅读入口，先交代范围和完成标准，再用精确链接展开合同、SYNC、模型。共同设计内容一致不要求整个 checkout SHA 相同；依赖真实共享产物的任务待回收后派发。 机制权威见 [L4 共享模型与合同治理](L4-shared-model-contract-governance.md)。


<a id="dispatch-plan"></a>
## 17. 整体派发计划与持续并行执行

> 规范修订：v0.5.0，2026-09-19。适用于新规划批次；本节定义正式机制；waves-v1 首版已接入计划解析、注册/冻结、S6 候选投影和 workgroup 注册校验。实际平台派发仍由 Main 执行，不把候选建议当作 Agent 已启动。

### 17.1 对象与唯一权威

S4 必须交付每个 REQ 的整体派发计划，规范入口为 `docs/dev/tasks/index-REQ-<id>.md`；已有 index 可保留原路径并登记。计划沿用现有 index，不同时维护另一份 DISPATCH 和人工进度看板。

| 对象 | 拥有的事实 | 消费方式 |
|:--|:--|:--|
| TASK | 目标、读写范围、真实产物依赖、交付物、Closing Contract | 计划链接引用；依赖和 scope 从 TASK 解析，不双写 |
| 整体派发计划 | 当前 REQ 任务集合、推荐波次、并行依据、必要资源顺序、编排理由 | S4 交付，S5 审签，S6 消费 |
| TR-003 execution batch | 当前 generation 批准计划指纹与精确 TASK 集合 | 执行分母；计划文档不能计作 TASK |
| Assignment / Result / checkpoint | 实际 owner、尝试、结果及有效集成事实 | 复用控制面，不新增 Wave 生命周期 |
| S6 实时清单 | 就绪、排队、执行、待集成、阻塞和完成的派生视图 | 可重建，不反向写权威状态 |

开发主分支和发布上游只从 REQ 绑定读取。设计输入按上游约定读取已提交 Git tree；运行证据和控制事实遵循各自明确的来源契约。

### 17.2 最小交付格式

计划声明所属 REQ、格式版本 `waves-v1`、计划修订号、文档状态；链接需求、合同与执行入口。正文按稳定局部编号 W1、W2…列出 TASK Markdown 文件链接，每个非取消 TASK 恰好出现一次，以 `- [ ]` 构成简洁工作清单。

每波说明并行边界和放行依据；特殊资源顺序只在必要时记录具体对象、先后与原因。复用 TASK/manifest 的 required checks，不再复制验证命令。状态 complete 表示计划文档完整，不表示代码完成。

S5 冻结后，计划中的空框保持不变；实时勾选只在 S6 投影中由当前有效 verified 集成事实生成。Agent、完成时间、证据、依赖满足状态不得回填冻结计划或 TASK。导出的实时清单标明生成时刻和 runtime revision，只是快照。

### 17.3 依赖、资源与波次

1. 真实产物依赖只在 TASK 声明：消费者等待生产者的产物回到根目录开发分支并验证。只读相同模型、合同或需求不产生实现依赖。
2. 资源冲突优先通过收窄写域、唯一 owner、拆公共修改或隔离资源解决；无法消除才明确串行次序。不能为了资源排队伪造 TASK 产物依赖。
3. 波次从真实依赖的拓扑层推导，再考虑必要资源顺序。同波必须不存在内部依赖和未解决互斥；所有显式资源顺序与 TASK DAG 联合查环。
4. 波次是推荐布局，不是整波屏障。某 TASK 的实际前置全部满足即可进入候选，无须等上一波无关慢任务。因资源次序延后的任务仍须遵守该明确次序。
5. 同父目录不自动判冲突，不同文件不自动判安全；检查公共导出、lockfile、迁移编号、生成输出和可变测试资源。worktree 隔离不能证明合并或语义兼容。
6. 宽路径重叠给具体诊断；首版不自动认定同文件不同区块可并行。需通过可执行的隔离/汇总办法或调整归属解决，不能用一句“安全”绕过已有 scope 控制。

容量只决定当次派发数量，不改变逻辑波次、依赖或必需集合。主会话/配置声明真实可用容量；工具展示 active、queued、待集成数和限制原因。无充分理由不固定 WIP=2，也不超过平台能力；不实现预测排程器或全局最优搜索。

### 17.4 候选求值与实际派发

S4 对账、S6 投影和派发入口共用计划/依赖语义。S6 从批准计划与精确执行集合开始，对当前有效依赖消费、owner、Result、checkpoint、阻塞、即时冲突和容量重新求值；冲突应包含 reported 但尚未消费的写入责任。

在互相兼容的候选中，优先解锁下游多、位于关键依赖链或有早反馈价值的任务，最后按稳定 TASK ID 排序。没有工期数据时不宣称获得准确工期关键路径。

启动、恢复、结果到达、集成完成、阻塞变化和槽位释放均触发重算。派发入口在实际操作前复核版本、owner 与资源条件；建议清单不等于授权，更不证明平台 Agent 已启动。复用既有事务、消息与平台动作，不新建后台调度服务。

主会话按“计划 → 实时清单 → 下一批 TASK → 精确阅读清单”阅读；Builder 从自身 TASK 开始，按需查波次及前置。第一版优先扩展现有 `s6 status` 显示下一批，是否增加其他命令取决于真实动作，不在规范中假定尚不存在的命令可用。

### 17.5 完成、恢复与提示边界

Worker 提交子分支，Main 按 [worktree 治理](L4-worktree-governance.md) 合并到根目录绑定开发分支、运行检查并登记有效 checkpoint，随后及时清理。结果上报、Agent 停止或仅存在 merge commit 都不能单独勾选完成或释放消费者；清理仍为软提醒。

失败只阻塞相关后代，无关兼容任务继续。恢复从批准计划和现有控制面重建；历史已失效或被当前尝试替代的成功不能继续满足依赖。视图与阶段门共用当前有效事实判定，防止看板与门结论不一致。

缺派、闲置容量、待集成结果和待清理 worktree 在已有 Agent 可见通道给短提示，不新增 Stop/Tool 门禁。结构错误与冻结漂移进入已有阶段检查；真正缺前置或互斥使候选等待并显示原因。平台接线以实施时最新官方机制核验，不以终端显示代替 Agent 收到消息。

### 17.6 版本与验收契约

S5 后修改成员、依赖、写域、波次或资源顺序走已有规划返工和重新审签；容量变化、实际进度和按真实前置提前启动不修改计划，不重做 S5。证据 SHA 自动刷新不适用于正式计划。

历史执行 generation 缺少计划时明确标记 legacy，仍可恢复已有工作，不补造审签；新规划批次必须满足新契约。计划遗漏不能自证完整，需与当前 REQ TASK 声明和合同覆盖共同对账，包括不覆盖条款的 support TASK。

机制验收必须覆盖：独立三任务同时候选；A→B 且慢 C 独立时 B 提前释放；只读模型两端并行；公共写入/外部资源冲突；资源顺序与 TASK 依赖联合环；跨 REQ/遗漏/重复/取消成员；未提交或审签后漂移；未集成不得完成；容量不足持续补位；恢复无重复 owner；失效历史 checkpoint 不复用；计划不污染 TASK 分母或 S2 设计主体。只完成模板或看板展示不足以证明本节已落地。

修订依据与取舍记录保留在根目录报告；本文拥有机制规则，L3 只拥有各阶段交付与审查责任。

### 17.7 首版适配边界

waves-v1 使用具体仓库相对文件/目录写域，不接受 glob 作为新计划授权范围；共享可变资源在 TASK Resources 中以 `resource:<name>` 声明。重叠是 S4/S5 诊断，实际派发保守保留互斥，不靠模糊文本例外绕过。没有范围冲突的就绪项可在实际槽位内并行。

当前注册入口一 TASK 对应一个 Builder writer；正式输入和最终可写范围（write_paths 与 output_paths 的并集）必须与审核 TASK 一致。S6 新计划的完成判定共用当前 report、owner 和对应 checkpoint；检查点同时绑定 `completion_report_path` 与 `completion_report_sha256`，`verified_at` 只在集成检查成功时记录。报告内容改变或旧检查点缺少内容绑定时，必须重新运行集成检查；单纯 ack/cleanup 不得更新摘要或时间以追认报告。历史无这些字段的 legacy 检查点继续走原恢复路径，不当作新计划已验证事实。

容量由调用方显式给定；不声明容量时只展示状态而不选择实际批次。计划求值是确定性的启发式排序，非最优排程或自动唤醒器。

### 17.8 主会话实时派发摘要

派发摘要在父会话 SessionStart 及 PostToolUse 的 Agent 返回、Harness register-workgroup/task-complete/task-integrate 检查点生成，进入已有 additionalContext 指引。摘要列出就绪候选、待回收结果和等待原因；每类最多四项，完整清单通过 s6 status 阅读。未声明容量时只说明逻辑候选，不推断实际槽位或声称已派发。

SubagentStop/TeammateIdle 不用普通 systemMessage 冒充父会话投递；异步 Agent 启动不是完成。Worker 不接收全局派发摘要，普通工具调用不重复扫描计划。投影失败只提醒检查，不改变既有 allow/block 判定。

## 18. 当前工作包、加载与接收优先契约（2026-10-13）

实施全局G01/G02/G06时，Assignment仍是派发权威。workpacket是当前recipient的只读投影，按本人owned_outputs、done_when、checks、冻结读集、当前唯一操作、等待owner展开；Main看接收/调度，Worker只看本人作业。长scope不能挤掉动作/路径，短消息按Unicode边界截断。答案接收与离场见[L4阶段答卷§13](L4-stage-form.md#13-生产实施契约2026-10-13)。

加载选择由一个版本化profile生产：核心∪原责任必需∪已确认技术栈∪Assignment登记技能。SelectedContextSkills与原授权SkillRefs分离，不能以减载改变activation命令能力或触发未知技能默认授权。teammate/native/workspace共用选择；workspace没有Skill工具时用已有Read消费安装固定路径与digest，ready资产不覆写。

### 18.1 G01 首个实施切片：shadow 与实际资料送达

首版 profile 集中原角色 legacy 列表、建议 core 与 responsibility 必需集合。`shadow` 记录候选加载集合、来源/理由与实际安装路径/digest，不改变 legacy 加载、角色 frontmatter、manifest SkillRefs 或 activation。项目技术栈未形成已批准机器来源时，只消费已登记 Assignment SkillRefs，不从 project.yaml 空模板或自然语言猜技术栈。legacy 列表及 mandatory skills 只有一个生产者；缺失安装文件不能回落工厂同名文件。

新 1.2 workspace prepare 从原 Runtime 注册 Agent/Team 和冻结 manifest 生成 Assignment context，绑定 owner/assignment/runtime/代数/execution generation、原技能、职责、读写/输出/检查/收尾和安装 Skill digest。作为原 Execution.inputs 的一项冻结来源复制进 `.claude/inputs/assignment-context.json`，由 BootstrapReceipt 验证；不新增第二派发表，也不修改已有 ready/legacy preparing 资产。重试复用同一冻结 context；来源不同拒绝。Bootstrap 交付原内嵌 agent-message Schema，Worker 无需访问工厂 Go 源码。

冻结资料包含已绑定 committed tree 的共享 REQ/设计/合同，以及所属 TASK 与原登记 depends_on 的依赖闭包；原 agent-protocol 与 docs/rules 一同交付。读取次序先 TASK、合同，再 REQ/架构/设计、规则；不从本地未提交文档生成正式来源。显式 execution replacement 保留原 profile、派发与资料摘要，只重绑新 execution generation，旧资料及作者字节保留；S5 replacement 仅限原已注册、尚未接收的 document_review，禁止更换产品 check runner。来源/lease/接收限制和原人工边界继续成立，不能借 replacement 升级 profile。

plan_checkpoint 作业同时给原 PLAN_REPORT 的填写支架：机器填真实身份、manifest revision=1、时间与派发路径，专业 objective/steps/assertion_checks 等保留未答。草稿不符合正式 plan Schema，begin 仍走原校验/激活；支架不铸造专业计划、批准或 working 状态。可编辑 plan.json 不进入只读 BootstrapReceipt，也不覆盖既有作者字节；完整 schema、问题与填写说明在冻结 context。S5 的 planned_paths 使用原消费者授予的独占答卷路径，产品写域保持空。

shadow 只能证明选择和送达候选，不证明宿主实际加载或 token 改善。teammate 输出同一 shadow selection；native context 引导读取真实 manifest/技能，workspace 用已有 Read 读取固定 context/Skill 路径，allowedTools 不新增 Skill。后续通过权限等价/资料读取/宿主案例后，再启用首个 Builder profile 和生成最小 frontmatter；本切片不提前宣称 G01 全部完成。

### 18.2 G01 native/teammate 的共同资料投影

新1.2 的已注册 owner 可取得共同 Assignment 资料视图：实际 Runtime/基线、Agent/Team/manifest、职责、原 SkillRefs、scope/checks/done、已绑定 committed 文档及原安装协议/规则、同一 shadow selection 和真实机器身份的未答计划支架。视图同时注明实际 source_root/agent_state；已 working 沿用原接受计划，不再要求重交。计划字段/问题/读取排序由共同生产者提供，不能三入口各写一份。native/teammate 阅读原权威根的完整文件路径，workspace 阅读原 Bootstrap 固定的副本；路径坐标不同不能混成同一物理来源。

native 仅向平台唯一已识别 owner 投递；teammate launch 只有输入 manifest 精确匹配当前注册 Runtime/Agent/Team/Assignment 才附共同视图。示例/未注册 launch 保持明确的候选投影，不生成“已注册”资料或授写权。旧1.1 不追加注册资料入口。源资料/身份不完整、漂移或 symlink 给 Main 具体来源，不宣称 ready，也不从其他 owner、工厂 Skill 或不同 manifest 回落。

共同资料 JSON 是内容寻址的只读派发投影，由实际来源再次校验后独占发布，给 path/hash；不是第二 Runtime、激活或领域完成来源。shadow 投影可随当前原来源重新生成；这不等于 enabled profile 的执行版本迁移。已有 workspace execution 始终走其原 Inputs/Bootstrap 路径，禁止 native 资料作为独立执行的 Main 回落。正式 enabled rollout 前仍需落实 native/teammate 的固定执行版本和实际宿主消费验收。

plan_checkpoint 在原可写专业证据路径给机器填写的未答草稿，专业判断由 owner 填写；草稿不纳入只读来源，不覆盖已存在作者字节。approval/one_shot 不生成可激活的计划草稿。正式 PLAN_REPORT 仍走原 SendMessage(plan_ref)/agent-begin；S5 原题、答卷、责任回执及当前进入身份继续由具名消费者判定。schema 以实际嵌入版本、独占 path/hash 提供，不要求目标项目访问工厂源码。

native 最终 lifecycle 只给完整资料入口与 hash/原模式，并消费原800 rune预算；详细内容在视图内。若 S5 资料视图已含本人 packet/answer，避免在 wire 重复长路径挤掉 next。完成回执优先于新计划提示，未知身份不发行视图。以上投影不新增 Write/Skill/Bash 权限，不代答、代签或改变原门。

接收优先建议从原Board与integration checkpoint派生，不建立第二队列。reported不释放依赖，capacity未知不猜slot；默认仅建议integrate_first，显式反压只限制新派发预算、不删除required set。注册继续在原Writer内重验owner/DAG/scope冲突，真实集成及检查保持。

### 18.3 G01 固定 native/teammate 执行资料与首个 Builder 生成路径

原 Main 通过 `runtime assignment-context prepare --agent-id <actual-owner>` 将已验证的共同资料引用登记到原1.2 Agent.context_binding，含固定 path/SHA、context generation 与原 profile version/hash/mode。原 Writer/CAS/journal 保持唯一提交者；资料发布与专业草稿仍独占、不覆盖，失败不登记 ready/激活。旧1.1拒绝此新写入；未准备的 shadow 仍可投影，已准备者 lifecycle/readback 必须消费固定来源，不能重算到新版 profile。Agent工作状态、计划/激活/完成仍以当前原 Runtime 为准，资料中的 agent_state 仅表示准备时捕获值。

绑定后每次送达校验原 Runtime/基线/owner/Assignment/Team/mode、原 manifest、来源摘要、固定 Schema与当前S5题干；源丢失/漂移/绑定损坏直接给原恢复路径，不回落新投影。已有独立 execution 继续原 Inputs/Bootstrap，不创建 native绑定。当前接收和退役优先于资料提示。原 Main prepare 重试同代数是幂等；禁止通过通用刷新替换既有 profile。

S5具名修订允许显式 `--replace-generation <old>`：仅当前实际新packet的 previous_submission_key 精确指向已接收旧packet，且owner/assignment/runtime/基线/进入身份保持。新 context代数保留原 profile/Skill来源/Schema和角色源；只允许原修订链决定的文档/卷/答案变化。已过阶段、旧代数、没有原接收前驱、来源版本变化拒绝。旧引用与作者字节保留，不用临时文件存在证明修订。

首个 Builder profile 的生成器只修改原 backend-builder frontmatter 的 skills 为共同profile核心集合；name/model/tools/disallowedTools/permissionMode及正文权限/责任保持，原 manifest/activation SkillRefs 不由减载集合替代。生成结果、来源/hash和profile/hash一起固定后才能进入新的execution；已有ready/preparing不覆盖。新增enabled执行入口仅用于已合法准备的新Builder，默认继续shadow，真实宿主未执行不得宣称消费/token收益。实现生成器和生成验收不是启用完成；原生/teammate实际role装配与首个workspace启用继续按同一契约接线。

固定 context_binding 也触发原 complete-pair checkpoint；RR-A文档重建不得丢失该索引。checkpoint复核固定资料字节、Schema、role/Skill源及原CONTEXT-PREPARE journal的owner/generation/ref-key。合法S5文档修订可能让旧资料投递过时，恢复保留该事实并提示原换代，不能因此重算profile或伪称ready。

首个角色生成器由 `runtime assignment-context render-role` 返回实际来源与生成字节，明确enabled=false；候选生成不等于真正execution的enabled profile。已有角色/安装回执不覆写，下一实施切片沿原新execution准备/Bootstrap接入生成结果。

### 18.4 G01 首个 workspace Builder enabled profile

新1.2 S6 backend-builder execution.prepare 可显式 `--context-profile backend-core`；默认shadow。只有首次合法新execution接受选择，其他角色/阶段/旧1.1拒绝enabled选择。已准备实例省略选择复用原profile，显式不同选择拒绝；原replacement只重绑execution代数并保留固定profile/生成源，不用replacement升级角色。

prepare从实际原角色生成core前置列表，仅skills块变化；生成字节的source/target/hash、原角色source/hash、核心集合及profile信息进入原冻结context和Execution.inputs。Bootstrap在新Worker目标角色路径安装这份固定字节并纳入原Receipt，其余角色/Skill资产仍完整安装，Main资产/安装回执不变。缺源、symlink、生成字节冲突或旧Worker资产不同拒绝，不覆写已有ready/preparing。

原manifest/activation SkillRefs仍是权限来源，name/model/tools/disallowedTools/permissionMode与正文不变；selected额外技能保持固定安装路径/摘要，由原Read入口消费。workspace.allowedTools不新增Skill。launch仅使用原登记角色名和固定Worker目标，提示说明enabled只控制加载，原计划/人闸/实际检查/commit/report/integrate不改。

验收从原S5基线经真实双PASS到S6，再注册实际Builder并走prepare/原begin/原commit/原report/原集成；比较Main角色/激活字节与有效权限，Worker实际core前置4项、额外来源可读，重试/replace固定，漂移拒绝，实际子进程argv/CWD/角色文件符合冻结数据。fixture/fake宿主只能证明装配和读路径；真实Claude消费及token对照未执行不得计为通过。enabled是显式新执行装配，不是自动全项目迁移或G01整体完成。
## G06 原看板的接收建议（2026-10-14）

`dispatch.AdviseReception` 纯投影原 Board 与 pendingIntegration，按直接
consumer 数、TASK ID、Assignment ID 排序。CLI `s6 status --advice` 与原
Main checkpoint guidance 使用同一投影；普通工具、Worker Stop/Idle 不增加
扫描或门。Board 与 pending 的 Runtime 身份、generation、revision 必须
一致；建议不成为可提交授权，原注册/集成事务仍复核。

原 pending 追加 TASK ID 和 checkpoint 引用；不新增队列。失败恢复保留
原重试 argv，专业纠正理由作为问题，不能把占位符当作可执行命令。
容量须显式声明，`--max-unreceived-results` 仅限制建议，原 Board.next
保持。未接收与 verified 待 ack 占建议上限，cleanup 单列。暂无 Runtime
持久阈值、自动 spawn、滞回或进程存活证明，后续按 G06 设计逐项落实。

## G06 原 Writer 内的持久反压（2026-10-14）

配置只在新 Runtime 的 init/install 明确选择并冻结。原 Runtime 可选
 dispatch_pressure 仅记录 Runtime ID/generation/paused，作为迟滞状态；不是
第二队列。原接收投影搬入 integration.Pending，HookContext.FromSnapshot
消费 Writer 持有的状态，禁止在同一事务拼接另一个 Runtime 版本。

原 Builder 报告与 Agent ACK 事务维护迟滞，原 workgroup 注册仅对 building
的 Builder 暂缓；原依赖/owner/冲突校验继续复核。ACK 的报告摘要沿原
Runtime evidence 登记来源生成，旧 ACK 不释放新报告；未确认或失配结果仍
计接收积压，匹配 ACK 后 cleanup 单列。只读建议不落迟滞，不改阈值。
原只读作业授权校验抽至 agentauth 复用，保留 assignment 接口及相同链验证。

## G06 只读 OS 锁观测（2026-10-14）

filelock.Probe 只打开已有普通锁文件，非阻塞探测并释放，不创建或修改文件；
前后复核同一 inode，替换/链接/权限问题给 unknown。integration.ObserveWorkspace
复用原 LockWorkspace 的规范化 checkout；对不同证据根仍应观察实际 Git 根。
闭合 observation Schema 不含可信 actor/PID，authorizes 恒 false。

Main 原 checkpoint 和显式 s6 advice 接线；busy 原资源时观察动作优先，
原 pending 与可派建议保持。普通工具/Stop/Idle 不增加门。观察与执行之间
允许变化，实际锁、原 Writer、Main required 检查与 fresh S7 继续负责执行。
真实子进程死亡后的释放以及真实 Main Go 检查的并发拒绝必须验收，不能
用单个 checkpoint 状态或预设运行记录替代。
