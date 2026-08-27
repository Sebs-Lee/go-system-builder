# BUG 登记册 — Hook 锚点选择全面审查（2026-08-28）

> 来源：对初始现役七个 Hook 锚点与平台 31 锚点全集的五维审查（时序覆盖 / 能力匹配 / 经济性 / 冗余度 / 可验证性）。本轮已将其中部分候选接入，当前状态以总表和“本轮实现对账”为准。
> 事实基准：`.claude/settings.json`（即仓库根 settings.json 的安装模板）、`internal/hook/*`、`internal/policy/*` 代码核对，以及官方 Hooks Reference 与《[L4 Hook 锚点全图](../../blueprint/L4-hook-anchor-catalog.md)》。
> 关联设计文档：《[Hook 与平台事件接线](../../blueprint/L4-hook-platform-wiring.md)》《[Agent 调度与治理](../../blueprint/L4-agent-dispatch-governance.md)》《[运行时控制面](../../blueprint/L4-runtime-control-plane.md)》。
> 状态词汇：`Open` 未处理｜`In-Review` 待 owner 裁决方向｜`Fixed` 已修（附提交号）｜`Partially fixed` 已完成可独立验收的子集｜`Verification debt` 代码已具备但等待外部平台实测｜`Deferred` 有意延后并记录触发条件｜`Wontfix`（附理由与复核触发条件）。本轮变更尚未提交，因此表中 Fixed 先标注“工作树，待提交”。

## 评级定义

| 级别 | 判定标准 |
|:--|:--|
| P0 | 存在**现实的静默绕过面**：屏障/门在可预见的正常使用路径上失效且无任何告警 |
| P1 | 核心执法能力**悬置或缺位**：历史上实际发生过的失效因此无法被机械拦截 |
| P2 | 已占格子能力欠用 / 缺失观测——不影响正确性但浪费已付费的平台能力或不可诊断 |
| P3 | 结构预留 / 可移植性 / 远期架构改善 |

---

## 总表

| ID | 级别 | 类别 | 标题 | 状态 |
|:--|:-:|:--|:--|:-:|
| HOOK-B01 | P0 | matcher 盲区 | PreToolUse 对 MCP 工具完全失明 + 策略引擎未知工具默认放行 | Fixed（工作树，待提交）|
| HOOK-B02 | P1 | 平台验证债 | SubagentStop/TeammateIdle 强制通道未经真实平台 doctor 实测 | Verification debt |
| HOOK-B03 | P1 | 主侧缺口 | 主会话无收工门：Main 宣布结束不受任何机械拦截 | Fixed（工作树，待提交）|
| HOOK-B04 | P2 | 观测缺失 | 决策链路零耗时记录；PreToolUse 超时=静默执法豁免窗口 | Fixed（工作树，待提交）|
| HOOK-B05 | P2 | 能力欠用 | SessionStart 只发文本横幅：additionalContext/source 分叉/watchPaths 三项未用 | Partially fixed（watchPaths deferred）|
| HOOK-B06 | P2 | 能力欠用 | SubagentStart 只报出生：Assignment 简报未注入子代理上下文 | Fixed（工作树，待提交）|
| HOOK-B07 | P2 | 信号替代 | 工具失败观测靠自建探针，未消费平台原生 PostToolUseFailure | Fixed（native audit observer；wrapper retained）|
| HOOK-B08 | P3 | 架构改善 | 冻结面漂移感知是拉模式：FileChanged+watchPaths 哨兵体系未建立 | Deferred（数据/噪声门槛未满足）|
| HOOK-B09 | P3 | 治理盲区 | loop-definition/hook-policy 运行期外部改动无第二只眼（ConfigChange） | Fixed（audit-only）|
| HOOK-B10 | P3 | 结构预留 | 单条 PreToolUse 注册项混载快慢工具，timeout 无法分级 | Deferred（等 B04 指标触发）|
| HOOK-B11 | P3 | 可移植性 | matcher 不含 PowerShell，Windows 模板可移植性存疑 | Deferred（跨平台批次）|

---

## HOOK-B01 · PreToolUse 对 MCP 工具完全失明（P0）

**现象（证据）**
settings.json 注册的 matcher 为 `Write|Edit|MultiEdit|Bash|NotebookEdit|Task|TaskUpdate|Agent`，不含任何 `mcp__<server>__<tool>` 模式。平台侧 MCP 工具命名规则为 `mcp__<server>__<tool>`（锚点全图 §2）。policy 引擎 `Evaluate` 在无任何规则命中时返回默认 allow（`internal/policy/types.go` 主路径），dynamic-Bash 的 fail-closed 特例只覆盖 Bash 解析失败场景。

**根因**
六条写屏障的触发前提全部是"工具被 matcher 命中且能分类出变异路径"。MCP 工具既不在 matcher 里、策略层也没有针对未知工具身份下变异嫌疑的兜底分支——两个条件同时落空。

**影响**
接入任何一个带文件系统写能力或命令执行能力的 MCP server（fs-server、shell-server、浏览器 agent 等都属此类），即可完整绕过全部六条屏障（locked_artifact_write、reviewer_product_write、repair_* 双门等）与 gate 投影路径。S7 验证冻结、S9 执行门前锁定在这些通道前形同虚设。这是静默失效：没有任何告警产生。

**修复方案**
双管齐下：
1. matcher 增补 MCP 通配（对照锚点全图 §2 命名语法：服务器通配必须带 `.*` 后缀；插件内置 server 有 scoped 前缀需一并覆盖）；
2. 策略引擎增加身份化兜底：携带 Agent 身份的调用若命中"未被匹配到已知分类的疑似变异工具"（无明确只读签名），按 dynamic-Bash 同款语义处置（verification 相 fail-closed，其余相 warn 可调）。
主会话不受影响；只读签名白名单要独立于本修复维护。

**验收标准**
- 用一个模拟 `mcp__evil__write_file` 的官方 payload 走 hook 路径：agent 身份下 verification 相得到 block，envelope 带 rule id；
- 主会话同 payload 得到 allow 或 warn（不误伤）；
- 全量既有回归（含豁免面三测试）不回退；
- migration 白名单与锚点目录状态列同步更新。

**依赖 / 归属**：matcher 改动走《锚点全图》§5 六问快审（本文即依据）；策略改动归控制面篇 §8 屏障家族的增补条目。本轮已按该方案立项并实现，待提交后将状态从工作树更新为 Fixed。

---

## HOOK-B02 · SubagentStop/TeammateIdle 强制通道未经真实平台实测（P1）

**现象**
两格的全部强制力构建在"exit 2 + stderr 会把单行反馈送回同一活会话并使其继续工作"这一平台上。该假设至今只在 2.1.218 文档语义层面成立，从未在真实 Claude Code 进程中执行过 doctor 验证（环境限制，调度篇 §15 与接线篇 §10.1 双登记）。fail-open 合同（stopidle.go 头注释）意味着：假设不成立时这两格不会报错，只会静默退化为提醒。

**影响**
"计划后误 idle""拿计划当交卷"两类历史实测失效（调度篇 §12、S7 代入测试 R 系列问题）会原样回归，且退化是无声的——没有失败信号提示我们强制力已经消失。

**修复方案**
1. **前置**：目标环境跑通 exit-2 往返 doctor（最小脚本：teammate 空转→hook exit 2→观察其是否带 stderr 继续同一 assignment）。此为调度篇遗留最高优先缺口，本文承接其落地责任；
2. **冗余预案（先设计后接线）**：把 `TaskCompleted(continue:false)` 预登记为 B 通道——它与 stop-idle 走不同平台通路（任务完成事件 vs 会话停止事件），恰好构成独立失效域。仅当 doctor 证明 A 通道不可靠时启用；
3. **可观测补偿**：无论 doctor 结论如何，hook-decisions.jsonl 应记录 stop/idle 每次判定的 outcome 分布，使"exit 2 发出后 teammate 是否真的继续"成为可以事后统计的问题（对接 HOOK-B04 的观测底座）。

**验收标准**
- 目标环境下留有可复跑的 doctor 脚本与结论记录（通过/不通过 + 版本号）；
- 通道 B 预案在接线篇 §10 登记，含启用条件与启用步骤，禁止未经评审直接上线。

**依赖 / 归属**：目标环境可用性（外部条件）；通道 B 归属调度篇。**非代码缺陷，属验证债务登记**。

---

## HOOK-B03 · 主会话无收工门（P1）

**现象**
Worker 侧有两道停机门（SubagentStop/TeammateIdle），但 Main 会话的回合结束（平台 `Stop` 锚点）完全没有注册。"active workers 未收口不得宣布总任务结束"目前只是调度篇 §11 第 8 步的文字承诺与 L4 目录里的 ◐ 候选备忘。

**根因**
当初按 Worker 问责设计的停机面没有镜像到编排者自身。历史代入测试中 Main 抢停/宣布结束正是反复出现的失效（README changelog 记录的两轮 E2E 代入测试均有涉及）。

**影响**
DRIVE 循环可以在 ready/queued 责任尚未派发、Result 尚未消费的情况下自然收敛结束，交付循环无声中断——这与 B02 相反，属于"闸门本来就该有而根本没装"。

**修复方案**
新注册 `Stop` 锚点，判定语义刻意求轻（毫秒级、每次回合结束都会触发）：
- 读 runtime 一行状态：存在 ready/queued 未派发 Assignment **或** result_submitted 未 consumed → decision:block + reason 给出唯一下一步（派发最优先项 / 进入消费）；其余情形一律秒级放行；
- 尊重 `stop_hook_active` 与平台连续 8 次 block 上限（锚点全图 §4 Stop 行）；
- blocked 理由区分两种文案："仍有待派发责任"（引导继续 DRIVE）vs "有结果待消费"（引导收口），避免把"合法等待后台完成"误拦成违规收工；
- 失败态度：控制面不可读时 fail-open（收工门宁少勿滥，避免堵死用户的正常退出）。

**验收标准**
- 五个典型场景测试：有 queued 未派发→block；有未消费 Result→block；活跃 worker 后台运行中→allow（等待合法）；runtime 缺失/不可读→allow；连续 block≥8 次后平台接管→人工确认语义正确；
- handler 自身耗时 p99 < 50ms（因每回合必触发）；
- migration 白名单与锚点目录状态列同步。

**依赖 / 归属**：新增锚点准入走《锚点全图》§5 六问；行为边界已按“只拦未派发责任/未消费 Result、活跃后台不误拦”实现并锁定测试。

---

## HOOK-B04 · 决策链路零耗时观测 + 超时旁路不可见（P2）

**现象**
决策信封与 `.claude/hook-decisions.jsonl` 均不含 handler 执行耗时字段；doctor 无任何 hook 时延检查。全部七格统一 timeout=10s（settings 内层 hook 对象），但预算消耗情况永远不可回答。

**影响**
平台的超时语义决定了这不仅是性能问题：PreToolUse 的 command 家族超时后**不拦截**而是落入常规权限流（锚点全图 §2）。换言之每一次 handler 卡顿都是一次静默的执法豁免——磁盘慢、锁等待（audit outbox 有跨进程锁）、未来新增分析逻辑都可能触碰这条线，而我们没有任何手段发现它已被触碰。

**修复方案**
1. envelope 增加 `elapsed_ms`（适配层记时，含 controller cycle + policy 求值全程）；
2. doctor 增加检查项：读最近 N 条决策记录，报告 p95/max 并在超过阈值（如 timeout 的 40%）时告警"时延逼近超时旁路窗口"；
3. 完成本项后再评估是否拆分 Bash 与文件写两条注册项以分级 timeout（见 HOOK-B10）——用数据决定，不预先拍脑袋。

**验收标准**
- 新字段出现在 jsonl 与 DecisionEnvelope schema；
- doctor 子命令可输出分布摘要；
- 构造一次人为慢查询（如给 loader 注入延迟）能在 doctor 输出中显形。

---

## HOOK-B05 · SessionStart 平台能力只用了一半（P2）

**现象**
现实现仅输出 systemMessage 文本横幅（adapter 生命周期包络）。三个未用的平台能力：
1. `additionalContext` 直接注入上下文（无需模型主动读取横幅）；
2. source 三分叉（startup/resume/fork 各自不同的恢复包——resume 场景用户带着旧上下文回来，需要的投影和冷启动不同）；
3. `watchPaths` 动态下发监视清单（HOOK-B08 的天然挂点，先行打通接口层）。

**修复方案**
分三步走，各自独立合入：① 已把恢复包核心内容（当前 stage/revision/唯一下一步命令/read anchor）经 additionalContext 注入，systemMessage 保留作终端可见摘要；② 当前先保留官方 source 值并在上下文中披露，source-specific 变更摘要待有真实 resume 差异需求后再加；③ watchPaths 未在本轮接线，留给 HOOK-B08 的动态冻结面方案，避免现在先引入一套 watch 生命周期。

**验收标准**
- 官方 payload 回放：additionalContext 出现在会话上下文（以 debug 日志或集成测试断言）；
- 三种 source 各自产物差异快照；
- watchPaths 生效性由 HOOK-B08 接手后回归。

---

## HOOK-B06 · SubagentStart 只报出生，未注入简报（P2）

**现象**
现实现仅渲染出生横幅（RenderWithRoot 生命周期分支）。平台语义明示该格可向子代理上下文注入 additionalContext（锚点全图 §4 行）。

**影响**
Assignment ref、effective_scope、done_when 目前只能依赖 spawn prompt 的自觉抄写进入子代理上下文——这是 D1 权威外置原则在子代理入口的一个漏点：权威事实的到达方式是"请它记得读"，而不是"必然在场"。

**修复方案**
从 runtime 当前 assignment 库存中定位本次 spawn 对应的责任（识别输入含 agent_type/名称），将其 identity/authority/plan 区块的关键字段生成紧凑 additionalContext（上限控制在一屏内，细节仍留在 Assignment Record 本体）；无法唯一定位时维持现状横幅（fail-open）。

**验收标准**
- dispatch 后首屏出现 scope/done_when 机械文本的端到端断言；
- 多 assignment 并存且无法唯一归属时不注入、不猜测。

---

## HOOK-B07 · 工具失败观测靠自建探针（P2）

**现象**
S7 evidence window 冻结等失败观测由内部 wrapper/探针承担；平台原生信号 `PostToolUseFailure` 未注册。两者重叠：凡是 wrapper 能看到的工具失败，平台也会发出（且权限拒绝类/参数校验类不发的边界更干净）。

**修复方案**
本轮先注册并消费 PostToolUseFailure 作为**原生审计信号**，保留 wrapper 作为运行在自己进程内的补充信号源；evidence window 的语义和冻结触发暂不迁移，仍归控制面篇 §9。待两路信号做等价性回放后再决定是否删除 wrapper 触发职责，避免观测切换本身造成证据缺口。

**验收标准**
- 同一注入失败序列在切前切后产出等价 evidence window 快照；
- 权限拒绝不产生 window（符合平台不发语义，需在控制面篇补一句口径说明）。

---

## HOOK-B08 · 冻结面漂移感知是拉模式（P3）

**现象**
locked artifacts / frozen subjects 被动过后，发现途径只有下一次 guard 求值或主动 `runtime fingerprint` 刷新——被动依赖"有人再来查"。 drifted 期间 stale 基线持续被引用而无提醒。

**修复方案（体系而非单点）**
FileChanged 锚点 + SessionStart/CwdChanged 的 watchPaths 动态清单：watch 清单=当前 generation 的冻结面（bound REQ、locked documents、frozen_subjects 代表文件），命中变更即时落一条审计并向看板投影 drift 警示（不直接 invalidate——判定权仍在 guard）。工程重点在 watch 清单的生命周期管理（generation bump/amend 时换防）。

**验收标准**
- 仓库外进程篡改 locked document 在下一秒内产生 drift 审计行；
- amend 后旧 watch 清单失效新清单生效；
- 频繁编辑项目热文件不产生噪声（清单只含冻结面）。

**依赖**：HOOK-B05 ③ 先行打通 watchPaths 下发接口。

---

## HOOK-B09 · 治理资产缺第二只眼（P3）

**现象**
loop-definition / hook-policy / settings 的运行期外部改动（手工编辑、脚本）当前完全不可见；PreToolUse 只能看到经过工具的改动。

**修复方案**
注册 `ConfigChange` 作审计型消费者：记录变更来源与文件，对治理资产变更形成可检索审计。接受两个平台限制并在设计里补偿：policy_settings 层拦不住；该事件不承担拦截，所有输入都写入 audit 行。

**验收标准**
外部修改 docs/hook-policy.json 期间产生可检索的 ConfigChange 审计行；拦截动作与 audit 行一一对应。

---

## HOOK-B10 · 单条注册混载快慢工具，timeout 无法分级（P3）

**现象**
PreToolUse 目前是一条 matcher 串起八种工具、单一 10s 预算。Bash 需要 classifier 解析链（重），文件写类路径快。一旦未来引入分析型逻辑，慢工具会把整格预算抬高或逼出超时旁路（HOOK-B04 的放大器）。

**修复方案**
结构预留：拆为 `Write|Edit|MultiEdit|NotebookEdit`（短预算）与 `Bash|Task|TaskUpdate|Agent`（长预算）两条注册项。**触发条件数据化**：仅当 HOOK-B04 的耗时统计显示某类 p99 越过阈值的一半才执行拆分，不预防性拆。

---

## HOOK-B11 · matcher 不含 PowerShell（P3)

**现象**
平台上 PowerShell 是与 Bash 并列的执行类工具；现 matcher 仅含 Bash。当前 darwin/macOS 使用无影响，但模板向 Windows 团队分发时 exec 面屏障整体失明。

**修复方案**
一行 matcher 增补 + 分类器声明不支持时的安全姿态（建议对 PowerShell 命令体在未支持解析器时沿用 dynamic-bash fail-closed 兜底）。随下一个兼容性批次顺手处理。

---

## 本轮实现对账（2026-08-28）

本轮按“先补静默绕过，再补必经上下文，最后补只读观测”的顺序落地。实现不把每个 Hook 变成新状态机，所有变更仍收敛在同一个 `hook` CLI 入口和现有 Runtime/Assignment 事实源。

| 条目 | 已落地 | Agent 可见的恢复/下一步 |
|:--|:--|:--|
| B01 | `settings.json` 的 PreToolUse matcher 覆盖 `mcp__.*`；未知且无路径的 MCP 工具在 Worker/verification 场景不再静默 allow；已知只读 MCP 保留豁免；带路径 MCP 复用既有写屏障 | block/warn 消息明确要求先完成工具分类；Worker 按提示改用 activated Assignment 允许工具 |
| B03 | `Stop` 读取已有 review assignments；未派发责任或未消费 Result 才 exit 2，活跃后台 Worker 不误拦，Runtime 不可读 fail-open | stderr 明确区分“先派发”与“先消费 Result” |
| B04 | DecisionEnvelope 与 outbox 记录 `elapsed_ms`；`doctor` 输出 count/p95/max，并在达到 10s 的 40% 时告警 | 告警明确说明 PreToolUse 超时可能形成静默旁路；B10 仍由数据触发 |
| B05 | SessionStart 使用原生 `additionalContext` 注入 source/stage/revision/next/read；SubagentStart 使用原生 `additionalContext` 注入唯一匹配 Assignment 的 scope/done_when/required_checks | 上下文控制在 800 字节；多 Assignment 无法唯一匹配时不猜、不注入 |
| B06 | Assignment brief 从 Runtime workgroup manifest 读取，不从 spawn prompt 猜测 | assignment_id/task_id/scope/done_when/required_checks/next 均在子代理首屏可见 |
| B07 | 注册并消费 `PostToolUseFailure`，落去重审计 envelope；保留现有 SendMessage/wrapper 观察路径 | 原生失败信号是证据输入，不改变工具结果、不替代 evidence window 语义 |
| B09 | 注册并消费 `ConfigChange`，落去重审计 envelope；不声称能阻止 policy_settings 层变更 | 审计记录包含 source/file_path（若平台提供） |

以下项目本轮明确不实现，避免为了“覆盖所有平台格子”制造复杂度：B08 需要先证明 FileChanged/watchPaths 的动态列表不会引入热文件噪声；B10 只有在 B04 指标达到阈值后才拆分 timeout；B11 进入跨平台兼容批次；B02 仍需目标环境中的真实 Claude Code doctor，仓库内测试不能冒充平台实证。

### A 类历史结论复核

本次复测没有重新打开已修复的 A1/A3/A5/A6/A7：agent identity 已在 observer 识别链校验，draft/schema 的空对象与 not_applicable 约束已有测试和恢复文案；A2/A4/A8 属生成器契约摩擦，不是本轮 Hook 锚点的静默旁路。它们若再次出现，应回到相应 S7 plan/schema 工单，不在 Hook 层重复造容错状态机。

---

## 附录 A · 本轮裁定"现阶段不采用"的平台锚点（防重复评估）

| 锚点 | 一句话理由 | 复核触发条件 |
|:--|:--|:--|
| PermissionRequest | 替用户行使批准权，触及人闸模型（人闸契约修订前禁入；迁移模板维持禁词） | owner 专项裁决 |
| UserPromptSubmit | 每 turn 高频 × 收益稀薄，违反成本公理；SessionStart 包 + 看板足以承载 | 出现"会话中段方向丢失"的实际失效案例 |
| TaskCreated / TaskCompleted | 与 stopidle 职能重叠；TaskCompleted(continue:false) 登记为 B02 的 B 通道预案 | B02 doctor 证明 exit-2 通道不可靠时启用 |
| WorktreeCreate | 配置即整体替换平台建树行为；须先统一 S6/S9 隔离树创建来源 | 创建来源统一后重新评估 |
| MessageDisplay | 显示层脱敏不改 transcript，产品化阶段再议 | 模板进入对外发布阶段 |
| Elicitation 族 / Notification / Setup / StopFailure / CwdChanged / DirectoryAdded / InstructionsLoaded / PostCompact / UserPromptExpansion / PermissionDenied | 暂无对应失控或纯运维通报用途 | 各自对应失控首次出现时 |

## 附录 B · 待 owner 裁决

| # | 事项 | 关联 |
|:--|:--|:--|
| 裁决一 | HOOK-B01（MCP 通配 + 未知工具兜底）是否授权立项修复 | 已执行；提交后关闭 |
| 裁决二 | HOOK-B03（Stop 收工门）是否进入下一批次及其行为边界确认 | 已执行；提交后关闭 |
| 裁决三 | PermissionRequest 长期去向：维持禁词待议 vs 安排专项 | 附录 A 第一行 |

## 变更记录

| 日期 | 变更 |
|:--|:--|
| 2026-08-28 | 初版：五维审查产出的 11 条登记项（1×P0 / 2×P1 / 4×P2 / 4×P3）+ 不采用清单 + 三项裁决请求 |
| 2026-08-28 | 本轮：完成 B01/B03/B04/B06/B07/B09；B05 完成 additionalContext 子集并明确 watchPaths 延后；B02 保留为真实平台验证债，B08/B10/B11 明确按触发条件延期；同步 settings、policy、migration、L4 接线与锚点目录 |

---

## 附：S10 全弧实走修复批（2026-08-28 晚，owner 授权执行）

以 r13 血缘沙箱实走 S10 时确认并处置的缺陷与新发现：

### 已修复（沙箱复验通过）

| ID | 根因 | 处置 | 回归 |
|:--|:--|:--|:--|
| S10-WALK-A | `qualityGateEnvelopeFields` 只透传 missing[]，评估器给出的完整冲突串（如 `evidence_ref_missing`）从未到达 agent；systemMessage 同样只报裸 LOOP_GATE_UNKNOWN。根因实例：manifest 引用 `reverify-r13-1` 而注册 id 是 `reverify-r13-1.json`——逐字匹配失败但诊断不可见 | wire 字段补 `conflicts`/`error_code`；unknown 恢复包追加 `Conflicts:` 段 | `internal/cli/s10_wire_test.go`、`internal/hook/pretooluse_conflicts_test.go` |
| S10-WALK-B | `runtime evidence add` 对 S10 双 kind 入库 `review_round=null`，看板立即判 stale；CLI 虽有隐藏 `--review-round` 旗标但教练链零提及 | RecordEvidence 对 acceptance/release_audit 家族自动从信封继承 review_round；kind 拒绝文案补双词汇映射说明；注册回显收敛为一行摘要（不再倾倒全量 State） | `TestRecordEvidenceS10InheritsReviewRoundFromEnvelope` |
| S10-WALK-C | `s10 status` 引用审计口径松于 gate（不校验文件哈希），且严格审计会遮蔽 blocked/review_required 的路由呈现；hook-policy schema 枚举落后于并行开发的锚点扩张 | inspect 判据与 gate 对齐；路由型结论先于严格引用审计返回；枚举同步 `Stop`/`PostToolUseFailure`/`ConfigChange` | 既有 `TestS10StatusReportsBlockedManifestRoute` 复绿 |

修复后实测正门弧线：acceptance leg（manifest→envelope 注册→TR-015 committed）→ release_audit leg（八区 manifest→TR-017 committed）→ `awaiting_human_release`；blocked 分支经 explain 给出的手工绑定（`pause_record=generated:pause_checkpoint`）推至 `paused` 且暂停检查点含 from_state/document_fingerprints。

### 新发现 · 本轮已修复（工作树，待提交）

| ID | 发现 | 建议方向 |
|:--|:--|:--|
| S10-WALK-D | **TR-018 三重矛盾**：Hook 自动门因 `pause_record` 生成令牌无法合成而恒 not_ready；手工推动时 `release-auditor` 被 actors 白名单拒绝（仅 orchestrator 可推）；explain 尾注又写"Let the Controller take TR-018; do not call runtime transition manually" | Fixed：由 transition catalog 标记生成型 evidence 为 postcondition，gate 不再把 `pause_record` 当 precondition；Controller 自动执行 `capture_pause_checkpoint`，文档删除手工绑定路径 |
| S10-WALK-E | counterevidence 行 `outcome=unknown` 仍然强制非空 `evidence_refs`，与自身报错文案"…or mark the result UNKNOWN"矛盾 | Fixed：仅 `unknown` 允许空引用；其它已判定结果仍要求 evidence refs，UNKNOWN 继续阻断 gate |
| S10-WALK-F | acceptance 清单必填 `metrics.audit_area_coverage` 且恒须 =1（该域对其无意义） | Fixed：acceptance 只要求 requirement/contract/changed_path；release_audit 才要求 audit_area coverage |
| S10-WALK-G | S10 相位下纯探测性 PreToolUse 每次都提交 milestone_refreshed 使 revision +1（本次走查人为推高 10+ 个修订号） | Fixed：milestone semantic identity 和 journal idempotency key 排除 observed/source revision 等 volatile fields；revision 仍无上限，只有语义未变时不再写入 |

### 本轮继续发现并修复的契约缺口（工作树，待提交）

| 范围 | 根因 | 处置与回归 |
|:--|:--|:--|
| Hook envelope schema | Controller wire 已新增 `quality_gate.error_code/conflicts`，但 `hook-decision.schema.json` 的 `additionalProperties=false` 未同步字段 | schema 增加两个字段；`TestQualityGateWireWithConflictFieldsSatisfiesHookDecisionSchema` 锁定真实序列化结果 |
| Native observer 去重 | `DecisionEnvelope` identity 只含 session/event/agent 等公共字段，`ConfigChange`/`PostToolUseFailure` 的不同 path/error 会被 outbox 当成同一事件 | identity 纳入稳定 native payload fingerprint；同 payload 仍幂等，不同 payload 各留审计；`TestDistinctConfigChangeEventsAreNotDeduplicated` 回归 |
| Native observer 错误姿态 | audit-only observer 的 policy/outbox 失败返回非零，可能把“只观察”误变成工具失败/平台阻断 | observer 保持 fail-open，仅 stderr 报告；`TestNativeObserverAuditFailureRemainsFailOpen` 回归 |
| Assignment 归属 | SubagentStart/Task fallback 在多 assignment 或无 agent_id 时取第一行，可能把别人的 scope 注入当前 agent | 只有显式 agent 匹配或唯一未绑定 assignment 才建立上下文；歧义返回空并要求重新定位；`TestLoadFullDoesNotGuessFirstAssignmentWhenAgentBindingIsAmbiguous` 回归 |
| Main Stop 顺序 | Stop 收工门在 Controller cycle 之后才判定，可能先提交自动迁移/刷新 milestone，再发现有未消费 Result | Main Stop pending/未消费判定在 Controller 前 preflight；阻断直接走审计和 exit 2；`TestMainStopHookBlocksUnconsumedReviewResult` 以可自动迁移 fixture 验证 revision 不变 |
| wire rule id | 内部 anchor 使用 lowercase id，但 hook-decision schema 要求稳定 `HOOK_*` 形式，实际 envelope 在规则命中时不可验证 | 只在序列化边界 canonicalize `rule_id/matched_rule_ids`，内部比较与 recovery anchor 不变；`TestEnvelopeRuleIDsSatisfyHookDecisionSchema` 回归 |
