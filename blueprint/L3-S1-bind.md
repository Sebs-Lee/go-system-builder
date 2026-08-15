# L3-S1 — 绑定（Bind）

> 层：第三层 ｜ 上游：L2 §S1 ｜ 版本 v3.1.0（v3.0.0 叙事版 + §6 注意力预算；机制事实经调查核实，含 file:line）

## 1. 要实现什么

一条显式命令，把人锁定的需求登记为**唯一授权对象**，并落权威状态起点——从这一刻起，"哪些工作被授权、做到哪了"有唯一可审计的答案。

- 进入时：inactive 的空 runtime（`init` 造的空壳）+ 一份 locked REQ。
- 出去时：runtime 处于 `planning/design`（直落 S2），`bound_req` 带指纹入册，baseline generation=1，授权记录与审计首条落盘。
- 衡量：**一项目一活需求**——任何第二条 REQ 在结构上无处安放；此后每次 hook 事件都能从状态文件读到"当前在干什么"。

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

## 3. 选了什么、为什么（含否决）

| 子问题 | 选用 | 否决与否决理由（减法） |
|:--|:--|:--|
| 怎么保证"显式授权" | 一条人执行的命令 + `--approved-by` 记名 | 否决"身份强校验"——当前只是非空字符串，人工边界靠协议（human_boundary=true）非密码学；如实记录：够用，因为威胁模型是"防 agent 越权"而非"防人冒充人" |
| 怎么防双需求 | 四层唯一性（from=inactive/空日志/新鲜校验/幂等键） | 否决"支持多 REQ 并行"——单需求单周期是 L2 铁律，复杂度不值得 |
| 怎么保证写入安全 | 原子写+pending 自愈+reconcile 对账 | 否决"绑定失败人工修状态文件"——状态文件永不手编是 D1 的底线 |
| 绑定前查什么 | doctor+validate（环境健康） | 否决"重复 REQ 内容审查"——内容质量属 S0（模板自检+人）；bind 只查可绑定性 |
| 绑定后怎么让全系统知道 | hook 每事件重读状态文件 | 否决"绑定事件广播/通知机制"——重读比订阅简单且无漏报 |
| 终态后重启 | rollover（人审批证据+归档） | 与 TR-020（改需求：代际+1、下游全失效）是两条不同路径，不合并——一个是换周期，一个是周期内换基线 |

## 4. 怎么编排（时间线讲完一件事）

1. **健康自检**：主会话跑 `doctor` + `validate --all`——修的是环境（定义/策略/schema/指纹漂移），不是 REQ 内容。
2. **人执行绑定**：`req bind --req <path> --approved-by <身份>` → CLI 校验 → `transition.Apply(TR-001)`：五 guard 顺序求值 → `bind_loop_req`（指纹入册、runtime_id=loop-{REQ}、generation=1）与 `record_loop_authorization`（授权记录）→ 原子提交 → 日志首条。
3. **落点校验**：主会话核对 done_when——bound_req 指纹=磁盘实算、cursor=planning/design、日志含绑定事件。
4. **控制面即刻生效**：下一次任何 hook 事件（SessionStart/PreToolUse）从状态文件读到 bound_req——S2 的第一个 PreToolUse 就带着授权上下文；未绑定项目则投影为 "S0: bind one human-locked REQ"。
5. **崩溃恢复**：绑定中途挂 → pending marker 自愈或 `runtime reconcile` 对账；终态后的重启走 rollover（人证据+归档），不是重跑 bind。

## 5. 期望效果

走完 S1：

- **授权唯一且可审计**：一活需求、指纹入册、授权记名、日志留痕；
- **结构性防住**：双绑定（四层唯一性）、半绑定状态（原子提交）、绑定后漂移（指纹 mismatch 即可见）、"绕过绑定开工"（一切门以已绑定为先决，未绑定投影只指向"去绑定"）；
- **交给 S2**：runtime 处于 planning/design + milestone 初始值（单一下一步）——S2 从权威投影起步，不靠记忆。

**如实记录的已知缺口**（供第四层修复清单）：①TR-001 五 guard 中仅 `no_other_active_loop` 有真实语义体，其余四个只查"证据非空"（guards.go:242-270 自注 "guard-theater"）；②两个证据槽是字符串 ID 而非真实证据引用（run.go:224）；③文档要求日志含 `req_bound`，代码实际写入的事件名是 `loop_requested`（engine.go:125 vs agent-protocol.md:197）；④loader 读不到 req id 时 fail-loud 回退硬编码 "REQ-039"（loader.go:763-781）。

## 6. 注意力预算与渐进披露

总评：**全系统注意力分配的样板**——零方法论阅读，一条命令 + 机器自证；本 stage 只有实现债，没有分配债。判定尺见 L3-README「注意力分配原则」。

### 6.1 当前错配（什么不对、为什么不对）

| # | 错配 | 为什么不对（L1 根据） |
|:--|:--|:--|
| 1 | TR-001 五 guard 中 4 个是证据非空桩（guards.go:112-115），仅 no_other_active_loop 有真语义（:249-275） | 公理四违例：桩制造"有五道门"的审计假象（实际一道）——用机制的名字支付了叙述的成本，却没买到确定性 |
| 2 | 文档要求 journal 含 `req_bound`，代码实际写 `loop_requested`（engine.go:125 vs agent-protocol.md:197） | 公理五违例：文档与机制同名异指，按文档重建理由的人拿到错误事实 |
| 3 | loader 读不到 req id 时回退哨兵 "REQ-039"（loader.go:763-783），注释自称 fail-loud | 形式上 fail-silent：哨兵字符串会流进投影文本；`NO_BOUND_REQ` 投影语义已存在却未复用 |

### 6.2 阅读预算（谁在何时读什么）

| 角色 | 进入时必读 | 触发式加载 | 永不需要读（机制承载） |
|:--|:--|:--|:--|
| 主会话 | doctor / validate 的**输出**（修环境，不读文档） | rollover 仅终态后发生（run.go:1202-1245） | 唯一性/原子写/崩溃恢复的全部机制细节——harness 承载，出错时报错自解释（D3：拒绝信息自我解释） |
| 人 | bind 一条命令 + `--approved-by` 记名 | — | — |

### 6.3 整改方向

- **删减**：四个桩 guard 折叠为一个通用 evidence-present 检查，真语义 guard 单列（guard-theater 清偿的局部执行）；
- **对齐**：事件名统一（`loop_requested` 与 `req_bound` 择一，以代码或协议为权威改另一方）；
- **删减**：REQ-039 哨兵改为复用 `NO_BOUND_REQ` 投影；
- **保持**：其余全部不动——S1 是其他 stage 整改的对齐样板（最小机制、最强控制、零文档依赖）。

## 变更记录

| 日期 | 版本 | 变更 |
|:--|:--|:--|
| 2026-08-14 | v1/v2 | 前两版（被判空洞→叙事不清） |
| 2026-08-14 | v3.0.0 | 叙事版；机制事实经 sub-agent 调查核实（含 guard-theater 等四项诚实缺口入档） | owner 复核 |
| 2026-08-15 | v3.1.0 | 新增 §6 注意力预算与渐进披露（错配诊断/阅读预算/整改方向），判定尺引 L3-README | owner 指示：渐进披露、机制承载规范、削减平白叙述 |
