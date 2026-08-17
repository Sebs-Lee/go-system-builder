# L3-S1 — 绑定与授权生命周期（Bind & Authorization Lifecycle）

> 层：第三层 ｜ 上游：L2 §S1 + L2「REQ 授权生命周期」 ｜ 版本 v4.6.0（v4.6.0 人工闸 scope 校验推广至全部人工转换；v4.5.x 深度终审/两轮对抗审查处置；v4.0.0 扩展为七动词控制面之家；机制现状经 sub-agent 全面调查核实，含 file:line；设计态与现状分开标注）

## 1. 要实现什么

一条显式命令，把人锁定的需求登记为**唯一授权对象**，并落权威状态起点——从这一刻起，"哪些工作被授权、做到哪了"有唯一可审计的答案。

- 进入时：inactive 的空 runtime（`init` 造的空壳）+ 一份 locked REQ。
- 出去时：runtime 处于 `planning/design`（直落 S2），`bound_req` 带指纹入册，baseline generation=1，授权记录与审计首条落盘。
- 衡量：**一项目一活需求**——任何第二条 REQ 在结构上无处安放；此后每次 hook 事件都能从状态文件读到"当前在干什么"。

**v4.0.0 扩展职责**：S1 同时是 REQ 授权生命周期的**控制面之家**（L2「REQ 授权生命周期」节的落地）。七个动词中：**进入、退出（unbind/abort）、重新绑定、修订**的命令面长在本 stage；**暂停/恢复**的触发器分散在各 stage 的 failure_route（TR-005/TR-010/defer 等——它们本就是各 stage 的失败路由），检查点与漂移校验机制的控制面描述归此。REQ 文件状态只回答"这份文件是什么"（locked=冻结基线，**不是"进行中"**）；"走到哪了"的唯一权威是 runtime 与 runtime-archive（D1）。

## 2. 手头有什么（真实机制）

| 机制 | 它能干什么 | 载体 |
|:--|:--|:--|
| `req bind --req --approved-by` | 唯一的绑定入口：CLI 侧校验（locked/版本/REQ- 前缀）→ 引擎侧复核（SHA-256、元数据、UI impact 三值） | run.go:177-232；engine.go:462-541 |
| TR-001 迁移 | 绑定=一次状态机迁移（inactive→planning/design）：5 个 guard + 2 个 action + 2 个证据槽；`human_boundary=true`、自动化不合格 | loop-definition.json；guards.go:154-281 |
| 绑定唯一性四层 | from=inactive 硬约束 + 空日志强制 + 新鲜 inactive 校验 + 幂等键/事件查重 | engine.go:141；store.go:1215-1222,235-239 |
| 原子写与恢复 | 临时文件+rename 原子提交；pending marker 崩溃自愈；`runtime reconcile` 快照/日志对账 | store.go:2316-2336,970-982 |
| 健康自检 | `doctor`（Manual 一致性/仓库语义/证据目录/策略指纹漂移/指标）与 `validate --all` | run.go:1398-1485 |
| hook 状态加载 | 每次事件重读 loop-state 的 bound_req/lifecycle——绑定即刻对控制平面可见，无通知机制的必要 | hookctx/loader.go:12-25 |
| rollover（终态后） | 归档旧 runtime 到 runtime-archive，落新 inactive 壳——终态后重新可绑定的唯一路径 | run.go:1202-1245；store.go:537-555 |

### 2.1 授权生命周期控制面（七动词现状真度——2026-08-15 sub-agent 调查核实）

| 动词 | 机制 | 真度 | 缺口 |
|:--|:--|:--|:--|
| 进入 | TR-001 + `req bind`（如 §2） | ✅ 全通 | ~~bind UX 六项待做~~ → **P0 已落地（2026-08-15）**：自动发现（接归档排除+终态排除）/ 自动 init / 人话输出（--json 保留）/ 身份提示（git identity 检测，attest 保持显式）/ 投影带完整命令行；`req list` 三色清单上线 |
| 暂停 | GTR-001~005 + `capture_pause_checkpoint`（**真**，actions.go:248-292：单次捕获不变式、富快照含指纹/generation/round/idempotency keys） | ◐→✅ 用户路径 | **`runtime pause` 封装已落地（P1）**：决策工件+human_decision 证据登记+GTR-001 一条命令；GTR-002/003/005 仍零触发器（各 stage 失败路由接续，后续批次）；GTR-004 桥仍无生产调用方 |
| 恢复 | TR-019 哨兵 `RESUME_FROM_PAUSE` 解析（EG:93-103）+ 指纹漂移拒绝 | ✅ 结构错位 | **`runtime resume` 封装已落地（P1）**，漂移→指路 amend；`baselines_unchanged` guard 仍是桩（真校验在 action 层）——guard/action 职责归位仍欠 |
| 修订 | TR-020：increment → **update_bound_req**（新，真校验：locked/指纹/版本严格递增，换入 bound_req+新代 documents 条目）→ 证据全作废 | ✅（P3） | pause 残留已修（引擎不变式：离开 paused 即清 checkpoint）；旧 REQ 保持锁定（loader 对 req 文档跨代保护）；`updated_req_locked` guard 仍为声明桩（真语义在 action，Manual 如实标注）；versions 目录仍为程序性约定；CLI `req amend` 一条命令 |
| 退出 unbind | `req unbind`：Store.Unbind 镜像 Rollover（disposition=unbound、审批 scope 独立、在飞实体软门 --force，forced+in_flight 落 manifest） | ✅（P2+审查修复） | **任意非终态含 paused**（v4.5.0 修正：曾违终裁拒绝 paused，造成"漂移时 resume 被拒→无路可退"死区）；归档留痕+回池；paused 解绑时 checkpoint 随档留存；E2E+审查测试钉死 |
| 退出 abort | TR-021（paused→aborted）/ TR-030（S11 人闸） | ◐→✅ | pause 残留已修；TR-021 走 `runtime transition`（证据经 evidence add 登记）+E2E 行为测试；guard/action 声明桩如实记录 |
| 正常结束 | TR-025 + rollover（审批四要素强校验 + 崩溃安全归档） | ✅ 样板 | **REQ 落章已落地（P2）**：状态行 locked→archived + req-archive.json 双指纹回执（封存哈希不动）；E2E 钉死 |
| 重新绑定 | unbind 回池再绑 / rollover 后新周期 | ✅（P2） | unbound 归档不排除（真实形态已存在）；同 REQ 终态重绑仍无护栏（显式 --req 可绑，如实记录待观测）；E2E 覆盖两条重绑路径 |

### 2.2 CLI 策略与命令矩阵（v4.1.0 设计态）

**五原则**（每条指认 L1）：①人闸命令人话化、代理命令双形态（`--json`）——C4/D3；②机器代办一切可派生参数（自动发现/身份提示/evidence 自动构造），人只给选择与记名——提案拍板纪律的 CLI 版；③拒绝即指路——公理五（"runtime 缺失"类错误应消失而非指路）；④一个动词一条主命令，`--req/--force/--json` 是披露深层非必经；⑤人话输出与 journal 可互证（指纹前缀/revision/事件名在 journal 有对应行）——D1/D6。

| 命令 | 演员 | 时机 | 输出契约 |
|:--|:--|:--|:--|
| 口头授权 + 代跑 | 人→主会话 | S0 锁定后 | 人说"绑定 REQ-xxx"（对话手势）→ 主会话按投影给出的命令行代跑 `req bind` → Claude Code 工具权限提示原生确认 → 人话输出；三层各司其职：手势在对话、确认在权限提示、记录在 journal |
| `req bind` | 人 | — | 人话 4 行（bound/sha 前缀/cursor+generation/event）+next；`--json` |
| `req list` | agent/人 | 任何时候 | 三色清单（draft/locked 可绑/locked 已终态+归档位置）；`--json` |
| `req unbind` | 人 | 任何非终态 | 在飞实体软门→`--force`；成功附池提示 |
| `runtime pause --reason --approved-by` | 人 | 工作态 | 内部自动登记 human_decision（复用 scope 展开模式）+生成 pause_record |
| `runtime resume` | 人 | paused | 漂移校验人话；漂移→指路 amend |
| `runtime human-decision` / `rollover` | 人 | S11/终态后 | 已有；rollover 增 REQ 落章确认行 |
| `status/next/doctor/validate/explain/dry-run` | agent | 按需 | 只读双形态 |

**调用纪律**：主会话不主动敲任何人闸命令（只在人显式指令后代笔）；状态获取走 hook 投影（manual CLI 只保留初始化/绑定/对账/回滚/闸门）；生命周期迁移由 controller 自动做，手工 `runtime transition` 仅限 reconcile 指引下的恢复。

**实施批次**（已全部完成，2026-08-15）：P0 纯 UX → P1 修复与统一 → P2 新动词 → P3 amend 完整化 → E2E 全链 → 两轮对抗审查处置。每批闭环均执行：实现+测试→doctor/validate→§2.1 真度表更新→自审+第三人审查。

**已 owner 终批（2026-08-15，按工程建议执行）**：① archived 落章时刻=rollover；② bind UX 四参数（多候选列清单要求显式 --req / 自动 init 内嵌 / 人话输出默认+--json / 身份提示保持显式 attest）；③ unbind=任意非终态+在飞软门；④ `req archive` 与引导换绑维持 backlog（C5）。

## 3. 选了什么、为什么（含否决）

| 子问题 | 选用 | 否决与否决理由（减法） |
|:--|:--|:--|
| 怎么保证"显式授权" | 一条人执行的命令 + `--approved-by` 记名 | 否决"身份强校验"——当前只是非空字符串，人工边界靠协议（human_boundary=true）非密码学；如实记录：够用，因为威胁模型是"防 agent 越权"而非"防人冒充人" |
| 怎么防双需求 | 四层唯一性（from=inactive/空日志/新鲜校验/幂等键） | 否决"支持多 REQ 并行"——单需求单周期是 L2 铁律，复杂度不值得 |
| 怎么保证写入安全 | 原子写+pending 自愈+reconcile 对账 | 否决"绑定失败人工修状态文件"——状态文件永不手编是 D1 的底线 |
| 绑定前查什么 | doctor+validate（环境健康） | 否决"重复 REQ 内容审查"——内容质量属 S0（模板自检+人）；bind 只查可绑定性 |
| 绑定后怎么让全系统知道 | hook 每事件重读状态文件 | 否决"绑定事件广播/通知机制"——重读比订阅简单且无漏报 |
| 终态后重启 | rollover（人审批证据+归档） | 与 TR-020（改需求：代际+1、下游全失效）是两条不同路径，不合并——一个是换周期，一个是周期内换基线 |
| 退出怎么做（v4.0.0 新增，已落地） | `req unbind`：镜像 rollover 机械——runtime 归档 disposition=unbound、新 inactive、REQ 回可绑定池；非终态任意时刻可用（含 paused）；在飞实体（in_progress/blocked 任务、planned/active teams）软门 + `--force`（forced+清单落 manifest） | 否决"pause→TR-021 两步退出"作为唯一路径——为退一扇进错的门先造一次暂停记录，成本与语义双输；否决"agent 可解绑"——撤销授权是授权域动作，人-only 与 bind 对称 |
| archived 在哪落章（v4.0.0 新增） | rollover 时刻由 harness 写 REQ 状态行（locked→archived）+ journal 记 `req_archived` 双指纹（from-sha/to-sha）；基线内容区永不动 | 否决"approve 时刻落章"——approve 只授权发布、周期未关（人批后未发布的窗口期语义含糊）；rollover 统一覆盖 approve/abort 两终态且本就是周期关闭点 |
| 绑定的入口形态（v4.2.0 修正） | 口头授权 + 主会话代跑：投影带完整命令行（P0）→ 人一句话确认 → 代跑经 Claude Code 工具权限提示原生确认 → journal 记名。**撤回 v4.0.0 的 /goal 提案**——机制查证（2026-08-15）：Claude Code 内置 /goal 是持续工作驱动器（evaluator 逐轮判定、驱动 agent 自主干活、推荐配 Auto 权限模式），语义与"人在场单点授权"相反；且它在手势（对话事件）/确认（权限提示已有）/审计（journal 才是权威）三层均不新增价值。docs/README 原否决"no separate /goal required"**维持成立**，理由升级 | 否决"/goal 触发绑定"——公理一违例的自我修正：望文生义引入未查证机制；公理四：无消费者的新机制（新增 commands/ 分发类别+维护成本） |
| bindable 怎么算（v4.0.0 新增，已落地） | `req list`：locked REQ 文件 − 当前 runtime 已绑 − 归档件 disposition∈{release_authorized,aborted} 引用的（**unbound 归档不排除**——换目标后绑回是正当的） | 否决"REQ 文件状态独判"——locked=冻结基线不是进行中，终态信息在 archive；否决"归档一律排除"——unbind 回池语义 |
| pause 残留怎么修（v4.0.0 新增） | TR-020/TR-021 的 action 链补清 `state["pause"]`（与 TR-019 restore 同一语义位） | 不是新机制是 bug 修复：现状残留会让下一次暂停必然失败（"would overwrite"，actions.go:249-251） |

## 4. 怎么编排（时间线讲完一件事）

1. **健康自检**：主会话跑 `doctor` + `validate --all`——修的是环境（定义/策略/schema/指纹漂移），不是 REQ 内容。
2. **人执行绑定**：`req bind --req <path> --approved-by <身份>` → CLI 校验 → `transition.Apply(TR-001)`：五 guard 顺序求值 → `bind_loop_req`（指纹入册、runtime_id=loop-{REQ}、generation=1）与 `record_loop_authorization`（授权记录）→ 原子提交 → 日志首条。
3. **落点校验**：主会话核对 done_when——bound_req 指纹=磁盘实算、cursor=planning/design、日志含绑定事件。
4. **控制面即刻生效**：下一次任何 hook 事件（SessionStart/PreToolUse）从状态文件读到 bound_req——S2 的第一个 PreToolUse 就带着授权上下文；未绑定项目则投影为 "S0: bind one human-locked REQ"。
5. **崩溃恢复**：绑定中途挂 → pending marker 自愈或 `runtime reconcile` 对账；终态后的重启走 rollover（人证据+归档），不是重跑 bind。
6. **生命周期动词（v4.5.2 起全部已落地）**：暂停——各 stage 失败路由触发或人 `runtime pause`，checkpoint 富快照落盘；恢复——人 `runtime resume`，逐文件核对暂停时刻指纹，漂移即拒并指路修订；修订——人批新 generation 后 `req amend` 真正换入新 REQ 并保持旧 REQ 锁定；退出——非终态（含 paused）`req unbind`（留痕归档回池），paused/S11 处 abort（终态）；结束——approve→rollover，REQ 状态行落章 archived+双指纹入册；重绑——unbound 回池再绑 / rollover 后新周期，`req list`+投影带命令+口头授权一条链。

## 5. 期望效果

走完 S1：
- **授权唯一且可审计**：一活需求、指纹入册、授权记名、日志留痕；
- **结构性防住**：双绑定（四层唯一性）、半绑定状态（原子提交）、绑定后漂移（指纹 mismatch 即可见）、"绕过绑定开工"（一切门以已绑定为先决，未绑定投影只指向"去绑定"）；
- **交给 S2**：runtime 处于 planning/design + milestone 初始值（单一下一步）——S2 从权威投影起步，不靠记忆。

**v4 生命周期控制面的期望效果**（v4.5.2 起已兑现，真度表 §2.1 为准）：七动词各有一条人话命令可达；撤销与结束全程留痕（弃周期在 archive 可审计、REQ 落章带双指纹）；bindable 判定机器可算（归档扫描，不以文件状态为准）；人的注意力只花在记名与拍板；审计面与事实一致（真度表 §2.1 每行从 ◐/✖ 收敛到 ✅）。

**缺口处置台账**（v4.5.2 重写为处置视角——以下七项在调查时均为真缺口，现全部处置完毕）：①TR-001 四个证据非空桩 → **已删**（guards 5→1，真语义单列）；②证据槽合成字符串 → **已做实** path@sha256 + approved-by 记名（含 v4.5.2 补上的 recover_command 遗漏点）；③事件名分叉 → **已统一** req_bound（definition+schema example+文档，journal 行以 transition_id=TR-001 溯源）；④REQ-039 哨兵 → **已改** UNBOUND 标签；⑤pause 残留 → **已修**（引擎不变式：离开 paused 即清）；⑥TR-020 后 REQ 失锁 → **已修**（loader 对 req 文档跨代保护）；⑦GTR-005 死 fact → **已清除**（不可达 runtime 的 fail-open 事实如实入档注释）。**共性如实记录（仍然成立）**：七个人闸的"human-only"实质是"evidence-only"——威胁模型是防 agent 越权，不是防人冒充人；agent 代跑与人执行在仓库层面不可区分（详见 v4.5.1 台账）。

## 6. 注意力预算与渐进披露

总评（v4.0.0 修订）：机制层仍是全系统的分配样板（零方法论阅读、机器自证），但 2026-08-15 实测暴露**UX 债与生命周期动词的可达性债**：绑定一次要 init→doctor→validate→bind（敲全路径+身份）→肉眼核对 JSON 输出；GTR-001 暂停要人手拼两段 evidence 引用。判定尺见 L3-README「注意力分配原则」——S1 的注意力对象是**人**与**审计者**：人的注意力只该花在记名与拍板，审计者的注意力不该被假门消耗。

### 6.1 当前错配（什么不对、为什么不对）

| # | 错配 | 为什么不对（L1 根据） |
|:--|:--|:--|
| 1 | ~~TR-001 五 guard 中 4 个是证据非空桩~~ → **已删**（guards 5→1，v4.4.0；§5 台账①） | 处置前为公理四违例 |
| 2 | ~~事件名分叉~~ → **已统一 req_bound**（P1） | 处置前为公理五违例 |
| 3 | ~~REQ-039 哨兵~~ → **已改 UNBOUND**（P1） | 处置前为形式上 fail-silent |
| 4 | ~~bind UX 六处摩擦~~ → **已改造**（P0：一条命令/自动发现/自动 init/人话输出/身份提示，v4.3.0） | 处置前为 C4/D3 违例 |
| 5 | ~~生命周期动词无 CLI 封装~~ → **已封装**（runtime pause/resume、req unbind/amend/list，v4.4.0；v4.6.0 起人闸由 scope 校验机器强制） | 处置前为公理二分工错位 |

### 6.2 阅读预算（谁在何时读什么）

| 角色 | 进入时必读 | 触发式加载 | 永不需要读（机制承载） |
|:--|:--|:--|:--|
| 主会话 | doctor / validate 的**输出**（修环境，不读文档） | rollover 仅终态后发生（run.go:1202-1245） | 唯一性/原子写/崩溃恢复的全部机制细节——harness 承载，出错时报错自解释（D3：拒绝信息自我解释） |
| 人 | bind 一条命令 + `--approved-by` 记名 | — | — |

### 6.3 整改方向（v4.0.0 五块全景）

1. **bind 简化六项**：自动发现（接 archive+终态+unbound 精化）/ 自动 init / 内置 preflight / 人话成功输出（--json 保留）/ 身份提示（检测 git identity，保持显式 attest）/ 投影带完整命令行；
2. **真实性五项**：TR-001 桩 guard 5→1（只留 no_other_active_loop，pm_context_matches_req 概念已随 PM Todo 删除死亡）/ 事件名统一 req_bound / 证据槽做实 path@sha256 / REQ-039 哨兵改 NO_BOUND_REQ / kind 字符串统一 "req"；
3. **REQ 状态机**：locked→archived 在 rollover 落章（改状态行+双指纹 journal `req_archived` 事件）；
4. **口头授权链** + `req list`（bindable 计算：locked − 当前绑定 − 终态归档引用，unbound 不排除）+ 投影带完整命令行（P0，agent 被告知跑什么、人不需记语法）；
5. **req unbind**：Store.Unbind 镜像 Rollover（disposition=unbound、在飞实体软门、--force）；**同步修三 bug**——①pause 残留（TR-020/021 清 state["pause"]）、②TR-020 后旧 REQ 失锁（generation 过滤）、③GTR-005 死 fact 接线或删除。"TR-020 新 REQ 入 runtime"是 amend 完整化功能（P3），不算 bug——与 §5 缺口 ⑤⑥⑦ 及 §2.2 批次口径对齐。
另：`runtime pause`/`runtime resume` 封装命令（把 evidence 构造收回机器）；TR-019 的漂移校验从 action 归位 guard 层。

## 变更记录

| 日期 | 版本 | 变更 |
|:--|:--|:--|
| 2026-08-14 | v1/v2 | 前两版（被判空洞→叙事不清） |
| 2026-08-14 | v3.0.0 | 叙事版；机制事实经 sub-agent 调查核实（含 guard-theater 等四项诚实缺口入档） | owner 复核 |
| 2026-08-15 | v3.1.0 | 新增 §6 注意力预算与渐进披露（错配诊断/阅读预算/整改方向），判定尺引 L3-README | owner 指示：渐进披露、机制承载规范、削减平白叙述 |
| 2026-08-15 | v4.0.0 | **扩展为"绑定与授权生命周期"**：承接 L2 v1.4.0 七动词全图；§2.1 新增控制面现状真度表（sub-agent 全面调查：五动词覆盖度+三项新 bug——pause 残留/TR-020 失锁/GTR-005 死 fact）；§3 新增六条选用裁决（unbind/archived 落章 rollover//goal 推翻原否决/bindable 计算/pause 残留修复）；§4 补生命周期设计态时间线；§6 更新 UX 债与动词可达性债；6.3 整改方向升级为五块全景 | owner 指示：REQ 可退出、正常结束、暂停、恢复、重新绑定——完整生命周期；先派 sub-agent 摸清现状再从 L1/L2 定安家 |
| 2026-08-15 | v4.1.0 | 新增 §2.2 CLI 策略与命令矩阵：五原则（人话化双形态/机器代办可派生参数/拒绝即指路/一动词一命令/输出与 journal 可互证）+ 命令矩阵（/goal、req bind/list/unbind、runtime pause/resume 封装）+ 调用纪律（主会话不主动敲人闸命令）+ 四实施批次 P0-P3 | owner 指示：明确 CLI 工具的运用策略与做法 |
| 2026-08-15 | v4.2.0 | **撤回 /goal 提案**（v4.0.0 引入、经 Claude Code 机制查证后自我修正）：内置 /goal 是持续工作驱动器（evaluator 逐轮判定+推荐 Auto 权限模式），语义与人在场单点授权相反；口头授权三层结构替代（手势在对话/确认在 Claude Code 工具权限提示/记录在 journal），零新增机制；docs/README 原否决维持成立、理由升级。§2.2/§3/§4/§6.3 及 P2 批次同步 | owner 指示：调查 /goal 复杂度收益比，查证 Claude Code 机制，对比口头告知方案 |
| 2026-08-15 | v4.2.1 | **多轮后自审修正七项**：F1 未终批参数不再以"裁决"口吻入档（L2 裁决段补终批状态，§2.2 增"实施前待 owner 终批"四点）；F2 §2.1 补"退出 unbind"行（零覆盖漏行）；F3 §6.3 三 bug 口径与 §5/§2.2 对齐（"新 REQ 入 runtime"是 P3 功能非 bug）；F4 L2 mermaid 全角冒号修正；F5 §5 补 v4 生命周期期望效果；F6 P0 批次补 protocol #s1 联动；F7 L2 显式声明文件 archived 是可读性镜像、bindable 判定权威是归档扫描 | owner 指示：多轮对话可能偏离，重梳逻辑自审 |
| 2026-08-15 | v4.3.0 | **P0 落地**：`req bind` 六项 UX 改造（自动发现/自动 init 复用 writeInactiveRuntime/人话确认输出+`--json`/git 身份提示）+ `req list` 命令（bindable 计算：locked − 当前绑定 − 终态归档引用，unbound 不排除）+ S0 投影带完整命令行（projectNext inactive 分支）+ protocol #s1 三步收敛为一步（actions 重写，输出即核对）。新增 7 个 UX 测试（req_bind_ux_test.go）；§2.1 进入/重新绑定两行真度更新 | owner 终批四点按工程建议执行 |
| 2026-08-15 | v4.4.0 | **P1/P2/P3 + E2E 全部落地**：P1（pause 不变式修复/TR-020 失锁修复/GTR-005 死 fact 清除/桩 5→1/req_bound/path@sha256/哨兵 UNBOUND/kind 统一/runtime pause+resume 封装）；P2（req unbind：Store.Unbind 镜像归档+在飞软门；rollover REQ 落章：状态行+双指纹回执）；P3（amend 完整化：update_bound_req 真校验+换入 runtime+req amend 命令）；E2E `TestLifecycleVerbChainE2E` 钉死七动词全链（进入→暂停→恢复→再暂停→修订→解绑→重绑→放弃→归档落章→新周期）。§2.1 真度表同步收敛 | owner 指示：先完整做完修改，再用 E2E 覆盖 |
| 2026-08-15 | v4.5.0 | **第三方对抗审查（sub-agent）处置**：P1×2 已修——①unbind 放开 paused（曾违"任意非终态"终裁并制造漂移死区，checkpoint 随档留存）；②原地 amend 修复——reachability 与 RefreshFingerprints 对 superseded req 代次豁免（旧代历史不可机检漂移/不可被刷新改写，hook 层仍按 path 保护）。P2×7 已修——amend 强制同 ID（异 ID 指路 unbind）与数字版本格式、--force 的 forced+in_flight 落 manifest、落章原子写+路径遏制+CRLF 行尾保留、bind 输出补 revision（原则⑤齐）、validateRolloverApproval 死代码删除、hook 不可达注释改如实（fail-open）、docs/README mermaid 事件名。**审查另揪出一个真 bug**：inFlightEntities 读 task["status"] 而 schema 字段是 state——软门从未生效，已修（含 blocked 态）并钉测试。补测 7 项：scope 隔离单测/paused→unbind/--force 落档/unbind 回执不授权 rollover/漂移→resume 拒绝指路 amend/req list --json/amend 异 ID 与非数字版本拒绝。P3 backlog 如实入档：动词 Apply 失败遗留孤儿 decision 证据、三个旧测试的合成证据格式漂移（因豁免仍绿） | owner 指示：第三方审查启动 |
| 2026-08-15 | v4.5.1 | **第二轮对抗审查处置**：处置验证 7/7 属实；P2×1 已修——team 在飞判定按 status 过滤（planned/active 才阻断；teams 只增不删，原实现会让软门永久误伤并催生 --force 习惯化）；P3×7 已修/入档——落章原子写保留原文件权限+父目录 fsync（原 CreateTemp 0600 会剥掉 git 跟踪 REQ 的组/他读）、reachability 豁免统一要求 generation>0（与 RefreshFingerprints 口径一致）、CLI scope 隔离测试改为仅前缀错误（revision 对齐，隔离证明纯净）、CRLF 翻转走真实 rollover 路径钉测、amend 输出如实化（hook 保护是按路径的，文件移走即止）、pause/resume/amend 输出补 revision（原则⑤全覆盖）。**如实入档两项**：①残余信任边界——生命周期动词由 agent 代跑时，仓库层面与人执行不可区分（hook 无 actor 区分，--approved-by 自我声明+detectGitIdentity 还会主动建议身份），唯一人闸是 Claude Code 权限提示，属设计接受的威胁模型（防 agent 越权靠协议纪律）；②非 req 旧代 documents 无豁免（契约应重做到新代是 L2 意图），但**旧代非 req 文档不得移动**，否则 validate 报 not reachable——S2 落地 documents 写入路径后此口径需重审 | owner 指示：再做一遍对抗审查 |
| 2026-08-15 | v4.5.2 | **深度终审处置**：①recover_command 证据格式遗漏（生产代码仍 `#lock`）→ 统一 path@sha256，P1 该项至此真正做全；②bind 补控制面指纹 preflight（definition/policy 与 runtime 记录不一致即拒并指路 doctor——原"self-preflights"仅覆盖解析层，合法漂移静默通过）；③journal 字面歧义 → protocol done_when 与 docs/README 改为真实表述（journal 行是 TR-001 commit，语义事件名在 state）；④README bind 示例更新（--req 可选+自动发现）；⑤本文档时态大扫除：§3 两行"设计态"、§4"待封装"、§5 缺口段七项改处置台账、§2.2 批次标注完成、§6.1 两行标已处置——已实现仍写待做是最直接的歧义源；⑥hookctx 三处 REQ-039 注释通用化；⑦四个旧测试的合成证据格式统一（消"哪种格式才对"歧义） | owner 指示：深度审查确认整改到位无遗漏无歧义 |
| 2026-08-17 | v4.6.0 | **人工闸 scope 校验推广（P0 修复）**：validateLifecycleApproval 此前仅 unbind/rollover 在 Store 层强制，pause/resume/amend/abort/六个人工发布转换只有 CLI 提示——agent 可自注册 human_decision 证据复用通过。现 loop-definition 为 GTR-001/TR-019/020/021/025..030 声明 `human_decision_scope`，engine validateRequest 强制 evidence 携带 `<scope>:<runtime_id>@<revision>`——一次批准只授权一个动词一个 revision（跨动词/跨 revision 复用即拒，测试钉死）；baselines_unchanged 改哨兵错误 ErrBaselineDrift（CLI errors.Is 分流）；checkpoint 两个无消费者字段（entity_snapshot_revision/committed_idempotency_keys）删除；amend 输出改为"文件原地不动"（hook 路径锁如实）；版本比较加溢出防护 | L1 价值观复审（sub-agent）：公理二/D2 违例处置 |
| 2026-08-17 | v4.6.1 | **对抗审查处置（第三方 sub-agent ×3）**：①scope cited 匹配收窄——优先只认 human_decision_record 槽位引用（额外 key 不再放宽匹配）；②决策工件补 authorization_revision（artifact 记 R，scope 绑 @R+1，审计语义对齐）；③guards.go "only four real checks" 陈旧注释改真；④changelog v4.3.0 归位+尾部换行。**如实入档**：evidence add 仍可自由铸造 scoped human_decision 证据（self-mint 未闭，闭它需 actor 身份机制，超出仓库层威胁模型——与 v4.5.1 残余信任边界同源，唯一人闸仍是 Claude Code 权限提示）；被删 checkpoint 字段会使存量 paused 状态在 validate 下非法（dogfood 阶段可接受，转换写入路径不受影响） | owner 指示：拉起第三方对抗审查 |
