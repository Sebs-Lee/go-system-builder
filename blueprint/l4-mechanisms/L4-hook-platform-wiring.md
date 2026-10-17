# L4 — Hook 与平台事件接线治理（Hook & Platform Wiring Governance）

> 层：第四层｜机制域：Hook 事件注册、payload 契约、输出/退出码决策契约、graph 监听、履职提醒与阶段流转
>
> 上游：L1 D2（自然路径执法）与公理五；Claude Code Hooks 平台能力基线；全部依赖事件总线的 L3 Stage
>
> 四家族分工：[L4 Agent 调度与治理](L4-agent-dispatch-governance.md)定义「谁在责任上行动」（含 TeammateIdle/SubagentStop 的**决策矩阵本体**）；[L4 权威状态机与迁移事务核心](L4-state-transition-core.md)定义「事实如何合法变更」（含 gate/auto_trigger 引擎机制）；[L4 运行时控制面](L4-runtime-control-plane.md)定义「写什么内容合规」（屏障规则表、脱敏闸等）；**本文档定义这些决策如何坐上同一条平台总线**——payload 怎么进来、结果怎么回去、出错时朝哪个方向失败。另外：平台 Hook 锚点随版本扩展，其目录、触发语义与选点审查流程的唯一权威是 [L4 Claude Code Hook 锚点全图](L4-hook-anchor-catalog.md)——本文 §2 的十类注册项是从该目录选出的**消费快照，不是平台边界**。
>
> 状态：v0.3.0。【当前实现】小节均为代码核对结论（核对日 2026-08-28），附证据位置；未经验证的平台假设在 §10 显式披露。

## 本轮优化的目标边界（2026-10-15，优先于旧接线契约）

引入 Hook 是为了 graph engineering 的监听、遗漏提醒和阶段流转。当前阶段直接
给出明确问卷和准确填写路径，让 Agent 知道要做什么、交什么；相关进展变化时提示
距离下一 stage 还差什么，保存问卷后由原消费者更新事实并按原图流转。

PreToolUse 只提供 observe/guidance 和当前 graph 进度，不对工具调用按身份、权限、
first-write 或 scope 作 deny。问卷问题与缺项提示共用原阶段要求；原检查、审查、人闸
及写入事务继续按图执行，Hook 不另加验收工序。长检查与合并由原责任执行。

本文后续标为历史实现的 enforce、身份授权链和逐工具屏障描述迁移前代码或平台 wire，
不是目标策略；当前代码迁移尚未完成，不能将本次文档修订解释为代码已移除。

目标以[全局设计 §0](../GLOBAL-OPTIMIZATION-DESIGN.md#0-收缩后的实施决策2026-10-15优先于旧方案)
为准；当前代码尚未迁移。

## 0. 准入逻辑：为什么这配一份独立 L4

基石五问自测：

1. 自有对象模型？——有且是全系统唯一的事件侧接口契约：`policy.Input`（官方 payload 无损入口）、`policy.Decision` 六值判定、wire envelope 输出形状（§3–§4）。
2. 五个以上 stage 深度消费？——全部 twelve stages 的每一次工具调用、会话启动、压缩、停机都经它（S6 写面、S7 提交与冻结、S8/S9 双屏障、S5 locked artifact、S11 自动化停机……无一例外）。
3. 独立失败面？——有：识别失败、控制面不可读、超时、payload 字段缺失；fail-open 还是 fail-closed 是本域的核心设计题且曾反复返工。
4. 平台边界接线契约？——它是四家族中与外部平台耦合最深的一个（Claude Code hooks 的 matcher/exit-code/payload 语义），且没有任何别的文档 owns 它。
5. 风险分级与验收？——有：事件分类学（observe/guidance/enforce）、迁移白名单治理、doctor 一致性检查、专项回归测试组。

五问全中。此前它的碎片散在 settings.json、internal/hook/*、internal/migration/templates.go、docs/control/hook-policy.json 和各 L3 的 hook 表格里——典型"人人消费、无人定义"。

## 1. 问题定义与本域目标

Hook 是 D2 的物理载体：所有把控必须挂在结构上必经的路径上。但若每个 handler 自行决定 payload 解析、反馈格式和失败方向，就会出现历史上真实发生过的失效：自造字段假装唤醒、warn 被悄悄升格成 block、audit 决策泄漏进用户可见通道、测试靠专用伪字段通过而真实平台不工作。

本域目标（全域不变量）：

- **单一总线**：所有事件走同一个二进制 `.claude/bin/loop-harness hook --event <E>`，路由收敛于 CLI 的唯一 evaluate 入口；不允许第二个直接注册的命令。
- **官方 payload 无损进入**：平台上有的字段一个不丢地到达策略层；策略层永不发明平台上不存在的输入。
- **输出可预测**：给定决策值与事件类别，envelope 形状、退出码、stderr 用途完全确定（§4 表）。
- **失败方向是设计而非事故**：每条路径的 fail-open / fail-closed 都有名有姓、有注释、有测试（§5）。
- **绝不模拟平台**：Runtime/Hook 都不得伪造"平台已被唤醒/已阻止"的事实（继承调度篇 §10.2 的铁律，本文管其在接线层的表达）。

## 2. 事件注册面

### 2.1 事件消费契约

注册集以本表为设计契约；超时是 Hook 短事务预算，不容纳完整测试或长合并。安装配置、迁移模板、policy 与 schema 同步消费：

| 事件 | matcher | 行为类别 |
|:--|:--|:--|
| `PreToolUse` | `Write\|Edit\|MultiEdit\|Bash\|NotebookEdit\|Task\|TaskUpdate\|Agent\|mcp__.*` | observe/guidance + graph progression cue：给当前责任问卷完整路径、具体缺项和下一动作；graph 事实由原保存消费者更新；不得发 deny、不得做身份/工具权限/scope 判定 |
| `PostToolUse` | `Write\|Edit\|MultiEdit\|SendMessage\|Agent\|Task\|SubagentHandback\|Bash` | observe；问卷保存调用原领域 Writer 更新 graph，再提醒具体缺项/流转；不能撤销已完成动作 |
| `PostToolUseFailure` | `*` | observe（原生失败审计；不阻断）|
| `SessionStart` / `SubagentStart` | （全匹配） | guidance（Agent 上下文）|
| `PreCompact` | （全匹配） | persist（保存恢复检查点；不靠输出注入上下文）|
| `SubagentStop` / `TeammateIdle` | （全匹配） | enforce（stop/idle 决策，矩阵本体在调度篇）|
| `Stop` | （全匹配） | enforce（Main 收工门）|
| `ConfigChange` | `*` | observe（治理配置变更审计；不阻断）|

四条治理规则：

1. **占格先过选点审查**：平台上可用格子远不止当前十类（全集见[锚点全图](L4-hook-anchor-catalog.md)，含 PermissionRequest / TaskCompleted / FileChanged 等 ◐ 候选）。任何新增/退出都先过该目录 §5 六问，再同步三处：本表、迁移白名单、目录状态列——不允许只改 settings。迁移模板中的禁词（PermissionRequest / TaskCompleted）含义是"**未经评审不得接线**"；ConfigChange 已经完成本轮审查，作为只审计观察器接线。
2. **protected_events 对账**：docs/control/hook-policy.json 的 `protected_events` 必须与 settings 注册集完全一致；doctor 校验 `hook_control.policy_ref` 与磁盘 policy 文件的 version+sha256 一致性，漂移时提示 `runtime reconcile-policy-ref`。
3. ** TASK 模板中的 SendMessage 仍是禁词**：TASK 正文不得指挥 Worker 直接发消息——消息纪律属于 agent-message schema 与调度篇，模板只承载任务事实。
4. dry-run（`renderHook=false`）与真实 hook 走同一 evaluate 路径，仅 output 渠道不同——保证被测即所运。

## 3. Payload 契约（输入面）

`policy.Input` 是全部事件的唯一入口类型；官方字段清单及其地位（设计+当前实现一致）：

| 字段 | 来源 | 地位 |
|:--|:--|:--|
| session_id / hook_event_name / tool_name / tool_input | 平台通用 payload | 通用 |
| agent_id | SubagentStop 等携带 | 官方；允许缺失 |
| agent_type | SubagentStart 携带 | 官方；用于唯一匹配 Assignment 简报，不得凭它猜测多个责任 |
| tool_use_id | 工具事件携带 | 官方；用于区分同一回合内的原生失败审计 |
| file_path / error / source | ConfigChange、PostToolUseFailure/SessionStart 携带 | 官方可选；缺失时仍保留事件审计，不发明值 |
| teammate_name / team_name | TeammateIdle 官方 2.1.218 payload | 官方；识别链一级来源 |
| transcript_path / agent_transcript_path / last_assistant_message / stop_hook_active | SubagentStop 官方 payload | 官方；`stop_hook_active` 用于防 hook 自激循环 |

输入同时保留 `cwd`、`tool_response` 与事件声明的 worktree 字段，缺失时不捏造。`CLAUDE_PROJECT_DIR` 用于项目权威根，`cwd` 用于路径/执行环境分类，二者不能混用。

三条铁律：

1. **无损透传**：以上字段逐一进入 Input，可缺失、不可改写、不可丢弃（omitempty 只是序列化整洁，不是许可丢字段）。
2. **禁止发明输入**：任何"为了测试好写而在 payload 里约定私有 agent_id"的做法违宪；测试必须用官方 payload 形状（历史反面教材：自造 agent_id 曾让 TeammateIdle 测试全绿而真实环境不可用）。
3. **有效身份回退一次**：`EffectiveAgentID = AgentID 否则 TeammateName`；回退到此为止，再往下是 §6 识别链的事。

## 4. 输出与决策契约（Output 面）

### 4.1 决策值六集合的 wire 行为

| Decision | stdout | exit | stderr | 说明 |
|:--|:--|:--|:--|:--|
| allow | 无附加信息时空；有 Agent 指引时使用受支持事件的 `hookSpecificOutput.additionalContext` | 0 | − | 不因附带指引改变工具权限 |
| info | 同 allow，附加阶段与下一步 | 0 | − | 状态投影进入 Agent 上下文 |
| warn | 受支持事件的 `hookSpecificOutput.additionalContext` 含事实与恢复动作 | 0 | − | **永远不许升格为 block**；Stop/idle 上持久化后由主会话下一可注入事件投递 |
| audit | 空 | 0 | − | 只落 `.claude/hook-decisions.jsonl`，用户不可见（戒律三）；记录 `elapsed_ms` |
| deny / block @PreToolUse | 迁移前 wire 格式；目标 Hook 不输出此决策 | 不适用 | − | 历史平台能力记录，不作为当前阶段问卷/graph Hook 策略 |
| block @TeammateIdle/SubagentStop | **stdout 静默** | **2** | 单行具体下一步 | 反馈必须经 stderr 回到同一活会话 |

adapter 三戒律（本域宪法，源码注释已固化）：① adapter 不拥有生命周期合法性，只渲染 Controller 结论与正向指导；② 永不把 warn 提升为 block；③ audit 判定零协议输出。

### 4.2 PreToolUse 的 graph 进度投影

Hook 从原 graph/领域门读取只读进度事实，将当前责任的问卷完整路径、具体缺项和下一动作序列化为 `hookSpecificOutput.additionalContext`。保存问卷后的原领域 Writer 负责校验答案、更新事实和执行既有迁移；PreToolUse 不执行权限决定、拒绝工具或提前跑完整业务链。不可解析的当前路径/缺项是系统接线待办，不让 Agent 读取 Runtime 猜测。

内部 rule anchor 可以继续使用代码常量的稳定内部命名；跨进程的
`DecisionEnvelope.rule_id` 和 `matched_rule_ids` 在序列化边界统一为
`HOOK_[A-Z0-9_]+`。这样既不破坏内部 recovery anchor，也避免不同消费者拿到
两套不可验证的规则身份。原生 `ConfigChange`/`PostToolUseFailure` 的审计 envelope
以稳定 payload fingerprint 参与 decision identity：同一 payload 重试幂等，不同
文件路径、错误或来源各自保留审计行；audit-only 写失败只报告而不改变原工具结果。

### 4.3 Agent 可见性与投递对象

设计始终跟随[最新官方 Hooks Reference](https://code.claude.com/docs/en/hooks)，不能因本机版本旧而保留错误 wire。`systemMessage` 在同步 Hook 中用于用户提示，PreToolUse allow/ask 的 permissionDecisionReason 也不能代替 Agent 上下文。SessionStart/SubagentStart 的 additionalContext 放在 hookSpecificOutput 内，不能写在顶层。

Stop/SubagentStop 的 additionalContext 会让会话继续，不能用于 worktree 回收软提醒。SubagentStop 的执行反馈属于子会话；需要主会话处理的回收提醒先写 checkpoint，在父会话 PostToolUse(Agent)、PreToolUse 或 SessionStart 投递。去重按事实版本与接收会话；仅成功输出后的投递记录不能冒充 Agent 已处理。

Agent 工具 `tool_response.status=async_launched` 是启动而非完成。使用 SubagentHandback 的子会话从该工具 `tool_input.message` 收集报告，不能将最终 closing text 当正文。报告抵达仍不等于已提交、已合并或已接收。

父会话 Agent 返回（兼容旧 Task 工具）以及 Bash 中 Harness 的 register-workgroup、task-complete、task-integrate 完成后，直接在现有 PostToolUse 提醒中追加 S6 只读派发投影；此路径不运行 Controller 迁移周期。普通 Bash、异步启动、worker 上下文和非 building 阶段不投递派发摘要。SendMessage 的 plan checkpoint 观察和自动激活保持独立语义。

PreCompact、WorktreeCreate/Remove 等丢弃消息字段的事件不承担 Agent 提醒。上下文限制长度，完整信息以权威记录路径引用。必须在最新 Claude Code 的隔离测试项目验证实际 Agent 接收和接收对象；仅 JSON 合法、stdout 有字或本地 CLI 可执行均不能宣称平台验收完成。

WorktreeCreate 如需接线，先完成 [Worktree 机制](L4-worktree-governance.md)的创建实现；该事件替代平台默认创建，不能当只读观察器注册。

## 5. 失败态度矩阵（本域核心设计）

观察与原领域门分工，适用范围精确到事件×条件：

- **PreToolUse observe/guidance**：graph 或问卷投影暂不可读时不拦工具、不猜身份或 scope；已知当前责任时给出实际问卷完整路径、具体缺项和下一动作。未知/损坏投影形成 Main/系统接线待办，由原领域 Writer/门决定正式答案是否可接收。
- **观察/停止判定类 fail-open**：对象无法识别、Payload 不足以定位责任时不猜绑不伪造——PostToolUse 识别失败=静默 observation、SubagentStop/TeammateIdle 无法定位已知 assignment 或属 one_shot/approval 派单/blocked 有据/completion 已报时一律放行。
- **原领域条件**：问卷保存后的 schema、专业内容、独立性、人闸和阶段迁移仍由对应 Writer/原门求值；Hook 不另加拦截条件。

**历史实现记录（核对日 2026-08-28，迁移前）**：PostToolUse 全程 observer 永不阻断；stop/idle 六条 allow 快捷出口集中在 StopIdleDecision；dynamic Bash 写变异解析不出路径时反手 fail-closed；旧主策略另有首写屏障和 agent-scoped 规则。以上是旧代码/wire 的事实，不是当前目标；代码仍待迁移，本文不宣称这些规则已删除。

## 6. 进度提示的责任路由

平台身份字段只用于把进度提示投递给合适的会话/责任，不构成工具许可、身份认证或 scope 授权。可用的消息来源按以下顺序匹配当前责任：

```text
payload agent_id → teammate_name（顶层）→ tool_input.teammate_name → 唯一活跃候选兜底
```

如果来源缺失或有歧义，不得因而拒绝工具调用，也不得把别人的问卷路径发给当前 Agent。只有成功解析出当前责任及实际问卷完整路径时，才发送该责任的进度提示；否则将投影/派发问题交给 Main 或系统维护责任，不要求 Agent 读取 Runtime 猜身份或路径。回答是否可接收由原领域 Writer/门核验。

## 7. Guidance 与 Enforcement 的分界

| 类别 | 事件 | 物理表现 | 消费者 |
|:--|:--|:--|:--|
| observe | PostToolUse(SendMessage) | 不产 permissionDecision、不动 lifecycle、无 exit 信号 | CLI 落 plan checkpoint（控制面的屏障有了触发事实）|
| observe | PostToolUseFailure / ConfigChange | 原生信号只落 audit，不改变原操作 | S7/S8 证据消费者、治理审计 |
| observe/guidance | SessionStart / SubagentStart / PreToolUse / PostToolUse | `hookSpecificOutput={hookEventName, additionalContext}`；每条责任进度提示含问卷完整路径、具体缺项和下一动作 | Main/Worker 重入座、问卷保存后原消费者更新 graph、履职提醒 |
| checkpoint | PreCompact | 持久化；由后续可注入事件重新投递 | 压缩恢复 |
| enforce | Stop；SubagentStop；TeammateIdle | 原调度/收工决策的既有 wire 行为 | Main、Worker |

判别尺（借用 L3-README 注意力分配三问）：观察信息用于更新原 graph；指导给人与 Main 的注意力；产品条件由既有领域门或人闸处理。PreToolUse 的进度提示必须直接给实际问卷完整路径、具体缺项和下一动作，不能以 deny 恢复包或流程文档替代。

## 8. 与其余三家族的边界（防重复定义）

| 内容 | 权威所在 | 本文角色 |
|:--|:--|:--|
| TeammateIdle/SubagentStop 决策矩阵（何时 block 何时 idle） | 调度篇 §10.2/10.3 | 只管把 block 变成 exit 2 + stderr 到达同一会话 |
| 迁移前逐工具屏障及其旧策略 | 控制面篇 §8 的历史记录 | 不再由 Hook evaluate；原领域 Writer/结果/集成门保留各自产品条件 |
| 原 gate/auto_trigger 与阶段迁移 | 状态机核心篇 §6 | 原 Writer/Controller 求值并流转；Hook 只把当前 graph 进度投影到提示 |
| 采集/脱敏闸的内容规则 | 控制面篇 §9 | 本文不涉及 |

## 9. 治理流程

- 新增/修改任何 handler：先登记 observe/guidance、graph progression 或既有 Stop/idle 类别；PreToolUse 不新增身份/工具权限/scope/first-write deny。问卷保存只接原领域 Writer；消息必须给出实际问卷完整路径、具体缺项和下一动作。Hook wire 与配置迁移仍按原治理校验。
- 移除决策：如 protected_commands 从 hook 主链路退出那样（§10），必须在 hook-policy schema 保留兼容字段的同时把"`不再被咨询`"写成带出处的注释事实，防止后来者误以为清单仍在执法。

## 10. 当前事实边界（诚实条款）

核对日 2026-08-28：

1. **doctor 与平台 smoke 的职责分开**：`doctor` 只做仓库内静态一致性（manual agreement、repository schema 集合、evidence catalog、policy_ref 漂移）；`health` 读取累计运行信号和 Hook 时延。`make platform-smoke` 会在真实进程边界执行 settings 中的命令与官方 payload，Claude CLI 存在且可运行时还报告版本；`make platform-smoke-required` 在没有可运行 Claude CLI 时以环境债退出 77。它不伪造完整 Claude 会话，exit-2 是否让同一 teammate 继续仍需在 disposable project 中按本节验收步骤实测。
2. **protected_commands 已退出 hook 执法**：`extension_protected_commands` 置空且有测试注释锁死"Hook no longer loads docs/control/protected-commands.json"；分类器仅供 Bash 变异路径分析使用。发布保护的现状 = S11 结构性停机 + 人操作纪律，这是有意为之的现状而非未完成事项（相应更正控制面篇 §15 的早期表述）。
3. Stop-idle 判定所依赖的全部决策逻辑收敛在 StopIdleDecision；Main 的 Stop 收工门收敛在 MainStopDecision，并在 Controller cycle 之前 preflight，避免已知非法收工先触发自动迁移或 milestone 写入；两者输出纪律（exit 2 stderr 单行下一步）由 CLI 层执行。
4. SessionStart/SubagentStart 可按当前 Assignment 投影给出履职信息；投影只用于正确送达提醒，不作权限判定。每条进度提示都含当前责任问卷完整路径、具体缺项和下一动作；无法解析时归 Main/系统接线待办，不让 Agent 查看 Runtime 猜测。
5. PostToolUseFailure/ConfigChange 是 audit-only：它们补平台原生信号，不替换已有 wrapper，也不假装拥有拦截能力；写 audit 失败保持 fail-open。
6. 一切 timeouts 固定 10s：若某 handler 未来需要更长预算，超时后的半途状态必须满足状态机篇的可恢复要求，而不是靠加长 timeout 掩盖。

> 本域的缺陷登记册属于目标项目的本地迭代产物，不随模板仓库发布。推进候选锚点时，应在目标项目的 `docs/reports/bugs/` 登记册中建立工单，并在完成后回写本域状态列。

## 11. DoD

- 十类事件注册、白名单、protected_events 三处一致，doctor 可对账；
- 任何事件新增 handler 前都有类别与失败态度登记，任何失败分支都有测试；
- 输出契约表（§4.1）覆盖全部六种决策值且与源码注释一致；
- 无自造 payload 字段存活于任何测试；
- 平台验收入口：先运行 `make platform-smoke`，再在装有 Claude Code 的 disposable project 运行 `make platform-smoke-required` 并记录版本、官方 payload、SubagentStop/TeammateIdle 的 exit-2 同会话继续结果；不要把 `doctor passed` 解读为该实测已完成。

## 变更记录

| 日期 | 版本 | 变更 | 依据 |
|:--|:--|:--|:--|
| 2026-08-28 | v0.1.1 | 边界叙事修正：明确七个注册锚点是消费快照而非平台边界；事件准入规则改写为「占格先过《Hook 锚点全图》六问」，禁词重定性为"未经评审不得接线"；头部挂接锚点全图为平台侧唯一权威 | owner 指示：文档逻辑不应限制只能用现有 hook 挂点；建立平台锚点全集文档供迭代对照审查 |
| 2026-08-28 | v0.2.1 | 对齐质量门诊断字段与 Hook schema；明确内部 rule id 到 `HOOK_*` wire id 的边界转换、原生观察事件按 payload 去重、audit-only 写失败 fail-open，以及 Assignment 歧义不猜绑 | S10-WALK-A 与 HOOK-B07/B09/B06 回归审查 |
| 2026-08-28 | v0.2.0 | 接入 MCP PreToolUse 通配与未知工具兜底、Main Stop 收工门、elapsed_ms 时延审计、SessionStart/SubagentStart additionalContext，以及 PostToolUseFailure/ConfigChange 原生审计观察器；明确 B08/B10/B11 仍为按数据触发的后续项 | HOOK-B01/B03/B04/B05/B06/B07/B09 实装与契约测试 |
| 2026-08-28 | v0.1.0 | 初版：把散落在 settings/migration/internal.hook/hook-policy 各处的事件注册、payload、输出契约、失败态度、识别链、guidance 分界统一为本域唯一权威定义；核实并纠正 protected_commands 已退出 hook 主链路的事实；承接调度篇遗留的真实平台 doctor 缺口的验收归属 | owner 批准的基石抽取批次（其一）；五问自测见 §0 |
| 2026-10-15 | 目标收缩 | PreToolUse 调整为 graph observe/guidance 与流转提醒；逐工具身份/权限/scope/first-write 拦截标为迁移前历史；进度提示要求问卷准确路径、具体缺项和下一动作。代码迁移尚未完成 | GLOBAL 优化 §0 用户决策 |


### 2026-09-18 · v0.3.0

明确 Agent 消息通道、主子会话投递边界和报告/启动区别。

### 2026-09-19 · PLAN_REPORT 的 worker 交接

PostToolUse(SendMessage) 的 control root 始终是 REQ 权威根目录；非根 `cwd` 提交的 `plan_ref` 必须通过当前 Agent 的已登记 Assignment 核对 worktree 归属和同仓库身份，再从该执行目录读取。拒绝未登记 worker、越界路径、符号链接和不匹配的报告身份，不能用根目录同名旧文件替代 worker 报告。

合法 worker 报告按内容哈希导入根目录 `.claude/evidence/plan-handoffs/`，登记与自动推进使用同一个归一化引用；根目录原有报告流程保留。导入文件不能覆盖已有不同内容，报告拒绝或 Runtime 登记失败后的 additionalContext 必须呈现最终失败原因，不能残留 `plan_report observed`。PostToolUse 仍以 exit 0 反馈观察结果，不升级为平台阻断。

## 12. 受控保存消费与有界观察（2026-10-13）

G01 lifecycle additionalContext 使用一个 800 Unicode 字符预算（按 rune、保持合法 UTF-8）。责任进度提示优先包含当前 stage/责任、实际问卷完整路径、具体缺项和下一动作；路径不可截断。若预算不足，省略非必要背景，不能缩写或替换成要求 Agent 读取 Runtime 的指针。身份字段仅用于选择提示接收会话，不作工具授权；身份歧义时交 Main/系统接线待办。

问卷保存由 PostToolUse/宿主保存边界接入原领域 Writer；相关 Write/Edit/MultiEdit 事件更新既有 graph 并触发阶段提示。Hook 不以“已发行/可信 actor/当前 owner”对工具写入作许可判定；原 Writer 按其领域契约决定答案是否接收，拒收反馈具体字段/缺项且保留原业务门。

官方 additionalContext 送达接收/本人完成/等待 owner/字段修正；每条进度提示包含实际问卷完整路径、具体缺项和下一动作。SessionStart/PreCompact 做有界 graph 恢复，Stop 保持原 preflight 顺序。没有 Hook 的宿主也须在保存边界自动调用原消费者；未接线时登记系统接线待办，不要求 Agent 手动 issue/consume 或自行拼恢复命令。

G08通用进展采样需Main明确监听Read/Glob/Grep，置于非dispatch Bash早返回之前；只读当前作业与有限缓存，不跑Controller、不改Runtime、不记全文。首次读取新的必要来源/范围是进展，重复Read/status不是；观察不足与外部等待不能强判stall。生产有限指标沿原spool，真实trace/模型实验留工厂评估工具。官方wire测试与真实Claude验收分开。
