# L3-S1 — 绑定与授权生命周期（Bind & Authorization Lifecycle）

> 层：第三层 ｜ 上游：L2 §S1 + L2「REQ 授权生命周期」 ｜ 机制现状经全面调查核实，含 file:line

## 1. 要实现什么

一条显式命令，把人锁定的需求登记为**唯一授权对象**，并落权威状态起点——从这一刻起，"哪些工作被授权、做到哪了"有唯一可审计的答案。

- 进入时：inactive 的空 runtime（`init` 造的空壳）+ 一份 locked REQ。
- 出去时：runtime 处于 `planning/design`（直落 S2），`bound_req` 带指纹入册，baseline generation=1，授权记录与审计首条落盘。
- 衡量：**一项目一活需求**——任何第二条 REQ 在结构上无处安放；此后每次 hook 事件都能从状态文件读到"当前在干什么"。

**扩展职责**：S1 同时是 REQ 授权生命周期的**控制面之家**（L2「REQ 授权生命周期」节的落地）。七个动词中：**进入、退出（unbind/abort）、重新绑定、修订**的命令面长在本 stage；**暂停/恢复**的触发器分散在各 stage 的 failure_route（TR-005/TR-010/defer 等——它们本就是各 stage 的失败路由），检查点与漂移校验机制的控制面描述归此。REQ 文件状态只回答"这份文件是什么"（locked=冻结基线，**不是"进行中"**）；"走到哪了"的唯一权威是 runtime 与 runtime-archive（D1）。

## 2. 手头有什么（真实机制）

| 机制 | 它能干什么 | 载体 |
|:--|:--|:--|
| `req bind --req --approved-by` | 唯一的绑定入口：CLI 侧校验（locked/版本/REQ- 前缀）→ 引擎侧复核（SHA-256、元数据、UI impact 三值） | run.go:177-232；engine.go:462-541 |
| TR-001 迁移 | 绑定=一次状态机迁移（inactive→planning/design）：guard（no_other_active_loop）+ 2 个 action + 2 个证据槽；`human_boundary=true`、自动化不合格 | loop-definition.json；guards.go:154-281 |
| 绑定唯一性四层 | from=inactive 硬约束 + 空日志强制 + 新鲜 inactive 校验 + 幂等键/事件查重 | engine.go:141；store.go:1215-1222,235-239 |
| 原子写与恢复 | 临时文件+rename 原子提交；pending marker 崩溃自愈；`runtime reconcile` 快照/日志对账 | store.go:2316-2336,970-982 |
| 健康自检 | `doctor`（Manual 一致性/仓库语义/证据目录/策略指纹漂移/指标）与 `validate --all` | run.go:1398-1485 |
| hook 状态加载 | 每次事件重读 loop-state 的 bound_req/lifecycle——绑定即刻对控制平面可见，无通知机制的必要 | hookctx/loader.go:12-25 |
| rollover（终态后） | 归档旧 runtime 到 runtime-archive，落新 inactive 壳——终态后重新可绑定的唯一路径；REQ 状态行落章（locked→archived）+ 双指纹回执 | run.go:1202-1245；store.go:537-555 |
| 人工闸 scope 校验 | 十个人工转换（GTR-001/TR-019/020/021/025..030）声明 `human_decision_scope`：evidence 必须携带 `<scope>:<runtime_id>@<revision>`——一次批准只授权一个动词一个 revision（跨动词/跨 revision 复用即拒） | loop-definition.json；engine.go validateRequest |

### 2.1 授权生命周期控制面（七动词）

| 动词 | 机制 | 说明 |
|:--|:--|:--|
| 进入 | TR-001 + `req bind`（自动发现/自动 init/人话输出+`--json`/git 身份提示）+ `req list` 三色清单 | 唯一性四层兜底 |
| 暂停 | GTR-001~005 + `capture_pause_checkpoint`（单次捕获不变式、富快照含指纹/generation/round）；`runtime pause` 封装（决策工件+human_decision 证据登记+GTR-001 一条命令） | GTR-002/003/005 由各 stage 失败路由接续；GTR-004 桥无生产调用方（如实记录） |
| 恢复 | TR-019 + 指纹漂移拒绝；`runtime resume` 封装，漂移→指路 amend（哨兵错误 ErrBaselineDrift 分流） | `baselines_unchanged` 真校验在 action 层（re-hash），guard 体仅查 evidence 非空——guard/action 职责如实记录 |
| 修订 | TR-020：increment → update_bound_req（locked/指纹/版本严格递增，换入 bound_req+新代 documents 条目）→ 下游证据全作废 | 引擎不变式：离开 paused 即清 checkpoint；旧 REQ 保持锁定（loader 跨代保护）；`req amend` 一条命令 |
| 退出 unbind | `req unbind`：Store.Unbind 镜像 Rollover（disposition=unbound、审批 scope 独立、在飞实体软门 --force，forced+in_flight 落 manifest） | 任意非终态含 paused；归档留痕+回池 |
| 退出 abort | TR-021（paused→aborted）/ TR-030（S11 人闸） | guard/action 声明桩如实记录（真语义在 evidence 校验与 action） |
| 正常结束 | TR-025 + rollover（审批四要素强校验 + 崩溃安全归档）+ REQ 落章（状态行+双指纹回执） | 审批scope 独立防串用 |
| 重新绑定 | unbind 回池再绑 / rollover 后新周期 | unbound 归档不排除（换目标后绑回正当）；同 REQ 终态重绑无护栏（显式 --req 可绑，待观测） |

**如实记录的信任边界**：七个人闸的"human-only"实质是"evidence-only"——威胁模型是防 agent 越权（scope 校验挡复用），不是防人冒充人；agent 代跑与人执行在仓库层面不可区分（`--approved-by` 自我声明），唯一人闸是 Claude Code 权限提示，靠协议纪律（主会话不主动敲人闸命令）。

### 2.2 CLI 策略与命令矩阵

**五原则**（每条指认 L1）：①人闸命令人话化、代理命令双形态（`--json`）——C4/D3；②机器代办一切可派生参数（自动发现/身份提示/evidence 自动构造），人只给选择与记名；③拒绝即指路——公理五；④一个动词一条主命令，`--req/--force/--json` 是披露深层非必经；⑤人话输出与 journal 可互证——D1/D6。

| 命令 | 演员 | 时机 | 输出契约 |
|:--|:--|:--|:--|
| 口头授权 + 代跑 | 人→主会话 | S0 锁定后 | 人说"绑定 REQ-xxx"→ 主会话按投影给出的命令行代跑 `req bind` → Claude Code 工具权限提示原生确认 → 人话输出；三层各司其职：手势在对话、确认在权限提示、记录在 journal |
| `req bind` | 人 | — | 人话 4 行（bound/sha 前缀/cursor+generation/event）+next；`--json` |
| `req list` | agent/人 | 任何时候 | 三色清单（draft/locked 可绑/locked 已终态+归档位置）；`--json` |
| `req unbind` | 人 | 任何非终态 | 在飞实体软门→`--force`；成功附池提示 |
| `runtime pause --reason --approved-by` | 人 | 工作态 | 内部自动登记 human_decision + pause_record |
| `runtime resume` | 人 | paused | 漂移校验人话；漂移→指路 amend |
| `runtime human-decision` / `rollover` | 人 | S11/终态后 | rollover 含 REQ 落章确认行 |
| `status/next/doctor/validate/explain/dry-run` | agent | 按需 | 只读双形态 |

**调用纪律**：主会话不主动敲任何人闸命令（只在人显式指令后代笔）；状态获取走 hook 投影；生命周期迁移由 controller 自动做，手工 `runtime transition` 仅限 reconcile 指引下的恢复。

**绑定入口形态**：口头授权 + 主会话代跑（投影带完整命令行 → 人一句话确认 → Claude Code 权限提示原生确认 → journal 记名）。不设内置 /goal——它是持续工作驱动器（逐轮判定+Auto 权限），语义与人在场单点授权相反，且在手势/确认/审计三层均不新增价值（公理一查证后否决）。

## 3. 选了什么、为什么（含否决）

| 子问题 | 选用 | 否决与否决理由（减法） |
|:--|:--|:--|
| 怎么保证"显式授权" | 一条人执行的命令 + `--approved-by` 记名 + scope 校验（防证据复用） | 否决"身份强校验"——威胁模型是防 agent 越权而非防人冒充人（如实记录：够用） |
| 怎么防双需求 | 四层唯一性（from=inactive/空日志/新鲜校验/幂等键） | 否决"支持多 REQ 并行"——单需求单周期是 L2 铁律 |
| 怎么保证写入安全 | 原子写+pending 自愈+reconcile 对账 | 否决"绑定失败人工修状态文件"——状态文件永不手编是 D1 的底线 |
| 绑定前查什么 | doctor+validate（环境健康）+ 控制面指纹 preflight（definition/policy 与 runtime 记录不一致即拒） | 否决"重复 REQ 内容审查"——内容质量属 S0；bind 只查可绑定性 |
| 绑定后怎么让全系统知道 | hook 每事件重读状态文件 | 否决"绑定事件广播/通知机制"——重读比订阅简单且无漏报 |
| 终态后重启 | rollover（人审批证据+归档+REQ 落章） | 与 TR-020（改需求：代际+1、下游全失效）是两条不同路径——一个是换周期，一个是周期内换基线 |
| 退出怎么做 | `req unbind`：镜像 rollover——归档 disposition=unbound、REQ 回可绑定池；非终态任意时刻可用（含 paused）；在飞实体软门 + `--force`（forced+清单落 manifest） | 否决"pause→TR-021 两步退出"作为唯一路径——为退一扇进错的门先造一次暂停记录，成本与语义双输；否决"agent 可解绑"——撤销授权是授权域动作，人-only 与 bind 对称 |
| archived 在哪落章 | rollover 时刻由 harness 写 REQ 状态行（locked→archived）+ journal 记 `req_archived` 双指纹；基线内容区永不动 | 否决"approve 时刻落章"——approve 只授权发布、周期未关；rollover 统一覆盖两终态且本就是周期关闭点 |
| bindable 怎么算 | `req list`：locked REQ 文件 − 当前已绑 − 终态归档引用（unbound 归档不排除——换目标后绑回正当） | 否决"REQ 文件状态独判"——locked=冻结基线不是进行中，终态信息在 archive |
| 一次批准授权多少 | scope 校验：`<scope>:<runtime_id>@<revision>`——一个动词一个 revision | 否决"批准可跨动词/跨 revision 复用"——一次拍板只值一件事（公理二：人闸的每次介入都应最小化） |

## 4. 怎么编排（时间线讲完一件事）

1. **健康自检**：主会话跑 `doctor` + `validate --all`——修的是环境（定义/策略/schema/指纹漂移），不是 REQ 内容。
2. **人执行绑定**：`req bind --approved-by <身份>` → CLI 校验 → `transition.Apply(TR-001)`：guard 求值 → `bind_loop_req`（指纹入册、runtime_id=loop-{REQ}、generation=1）与 `record_loop_authorization` → 原子提交 → 日志首条。
3. **落点校验**：主会话核对 done_when——bound_req 指纹=磁盘实算、cursor=planning/design、日志含绑定事件。
4. **控制面即刻生效**：下一次任何 hook 事件从状态文件读到 bound_req——S2 的第一个 PreToolUse 就带着授权上下文；未绑定项目则投影为 S0 起步引导。
5. **崩溃恢复**：绑定中途挂 → pending marker 自愈或 `runtime reconcile` 对账；终态后的重启走 rollover，不是重跑 bind。
6. **生命周期动词**：暂停——各 stage 失败路由触发或人 `runtime pause`，checkpoint 富快照落盘；恢复——人 `runtime resume`，逐文件核对暂停时刻指纹，漂移即拒并指路修订；修订——人批新 generation 后 `req amend` 换入新 REQ 并保持旧 REQ 锁定；退出——非终态（含 paused）`req unbind`（留痕归档回池），paused/S11 处 abort（终态）；结束——approve→rollover，REQ 状态行落章 archived+双指纹入册；重绑——unbound 回池再绑 / rollover 后新周期。

## 5. 期望效果

走完 S1：
- **授权唯一且可审计**：一活需求、指纹入册、授权记名、日志留痕；
- **结构性防住**：双绑定（四层唯一性）、半绑定状态（原子提交）、绑定后漂移（指纹 mismatch 即可见）、"绕过绑定开工"（一切门以已绑定为先决）；
- **生命周期控制面**：七动词各有一条人话命令可达；撤销与结束全程留痕（弃周期在 archive 可审计、REQ 落章带双指纹）；bindable 判定机器可算（归档扫描，不以文件状态为准）；人的注意力只花在记名与拍板；
- **交给 S2**：runtime 处于 planning/design + milestone 初始值——S2 从权威投影起步，不靠记忆。

## 6. 注意力预算与渐进披露

总评：机制层是全系统的分配样板（零方法论阅读、机器自证）。S1 的注意力对象是**人**与**审计者**：人的注意力只该花在记名与拍板，审计者的注意力不该被假门消耗。判定尺见 L3-README「注意力分配原则」。

### 6.1 阅读预算（谁在何时读什么）

| 角色 | 进入时必读 | 触发式加载 | 永不需要读（机制承载） |
|:--|:--|:--|:--|
| 主会话 | doctor / validate 的**输出**（修环境，不读文档） | rollover 仅终态后发生（run.go:1202-1245） | 唯一性/原子写/崩溃恢复的全部机制细节——harness 承载，出错时报错自解释（D3） |
| 人 | bind 一条命令 + `--approved-by` 记名 | — | — |
