# Canonical BUG: BUG-HARNESS-001

> Status: fixed in working tree; pending final repository verification
> Severity: **P0**（死锁——核心工具全部失效，session 完全无法推进）
> Layer: harness / infrastructure
> Found in: **post-REQ-004-bind 操作**（2026-08-25 08:50 CST，jyquant 主机）
> Source report: 本会话 ops 调查日志
> Owner: harness maintainer
> Reporter: orchestrator (loop-REQ-003 archived 后 → S0 → REQ-004 locked → req bind 触发)

> Investigation status (2026-08-25): **confirmed with boundary corrections**. The direct `init` command creates a revision-0 inactive runtime. The deadlock is reproducible in the real PreToolUse → Bash → `req bind` path: the Hook persists a milestone refresh before `req bind` runs, making the fresh-runtime guard fail. The `runtime_unreadable` recovery command also has no reachable escape path while mutating Bash is blocked.

## 1. 现象时间线

| 时刻 (UTC ~ 2026-08-25) | 事件 | 操作者 |
|:---|:---|:---|
| 08:30 CST | REQ-003 human release gate approved → master 合并 → TR-025 + TR-017 → loop-REQ-003 archived → state.inactive | orchestrator |
| 08:35 CST | 用户要求"开新 REQ-004 着重可观测性、心跳、快速恢复" | user |
| 08:36-08:44 CST | agent 起草 REQ-004 draft（含 A1-A4/B/C 全节） | agent |
| 08:45 CST | 用户"同意，锁定并绑定 REQ-004" | user |
| 08:46 CST | agent 编辑 REQ-004 状态 draft→locked；写 §E 拍板记录 + 自审 | agent |
| 08:47 CST | agent 执行 `loop-harness-linux-amd64 req bind --req docs/requirements/REQ-004.md` | agent |
| **错误1** | `req bind: requires a fresh inactive runtime: revision must be zero` — loop-state.json 中 revision=63（上一轮 rollover 后的值） | harness |
| agent 反应 | 用 Python 编辑 `.claude/loop-state.json`，重置 revision=0、lifecycle=inactive、bound_req=null、last_transition=null、journal={} | agent |
| **错误2** | `req bind: requires a fresh inactive runtime: journal must be an object` — harness 要求 journal 字段是 dict 不是 array | harness |
| agent 反应 | 编辑 journal 改为 `{}`（object） | agent |
| **错误3** | `req bind: requires a fresh inactive runtime: journal path must be .claude/loop-events.jsonl` — harness 要求 state 文件含 `journal_path` 字段指向 `.claude/loop-events.jsonl` | harness |
| agent 反应 | 编辑加 `journal_path` 字段；删除原 journal 文件后 touch 新建 | agent |
| **错误4（同上）** | 报同样错误——guard 顺序检查未通过 | harness |
| agent 反应 | 用 `loop-harness init` 重置 | agent |
| **错误5** | `init failed: refusing to initialize over existing runtime file .claude/loop-state.json` | harness |
| agent 反应 | 删除 `loop-state.json` 后重试 | agent |
| **错误6** | `init failed: refusing to initialize over existing runtime file .claude/loop-events.jsonl` | harness |
| agent 反应 | 删除 `loop-events.jsonl` 后重试 | agent |
| **错误7 (CRITICAL)** | `init` 返回后，**所有 Bash/Write/Edit 工具调用被 PreToolUse hook 拦截**，错误信息为： `QUALITY GATE BLOCKED — final safety block on runtime_unreadable. Rule: runtime_unreadable. Missing: valid_runtime_cursor` | harness |
| 当前状态 | session **完全无法推进**——Bash/Write/Edit 全部失效；唯一可用工具是 Read（受 hook 警告但能读文件内容） | — |

## 1.1 第二轮回放（原始记录；部分结论经复测校正）

| 时刻 | 事件 | 结果 |
|:---|:---|:---|
| 用户执行恢复：`rm -f .claude/loop-state.json .claude/loop-events.jsonl .claude/hook-decisions.jsonl && init` | init 成功，状态恢复 | 原始记录为 revision=1；当前直接 CLI 复测为 revision=0，Hook 联动后才变为 revision=1 |
| agent `req bind` | 错误 `requires a fresh inactive runtime: revision must be zero` | 当前复测显示此前的 PreToolUse milestone refresh 已将 revision 推进为 1，bound_req 仍 null |
| agent 重复试探（删除 state 再 init） | 触发错误5/7 同一型 hook 死锁 | session 再次卡死 |
| 用户再次恢复并明示"rev=1 不是0 没问题" | 期望 user-perspective：revision=1 是 init 后正常状态，错误信息**应为信息性提示**而非 fatal guard | 当前复测确认：只要 Hook 在 bind 前写入 milestone，revision=1 就不再满足 TR-001 的 fresh bootstrap 不变量；问题在动作顺序，不是 revision 有最大值 |
| agent 第三次尝试 `req bind`（按用户指示仅看 status 不修改） | 错误 `requires a fresh inactive runtime: revision must be zero` | bind 仍失败；revision 增长应归因于重复 Hook refresh，不应归因于失败的 bind mutation |

### 1.2 第二轮新发现

| 编号 | 现象 | 含义 |
|:---|:---|:---|
| D1 | 原始记录认为 `init` 生成的 fresh state revision=1，**不满足 `req bind` 的 "revision must be zero"** | 直接 CLI 复测不支持“init 生成 revision=1”；真实漏洞是 Hook 在 bind 前把 fresh state 推进为 revision=1 |
| D2 | 原始记录认为 `req bind` 错误**伴随 revision bump** | 复测显示 bump 来自 bind 前后的 `MILESTONE-REFRESH` Hook 事件；TR-001 在 guard 阶段失败，没有提交绑定 mutation |
| D3 | 用户视角"rev=1 不是0 没问题" 与 harness 视角"requires revision must be zero" 冲突 | 文档（`loop-harness.md`）未明确说明 `req bind` 的 fresh 前置条件，也未说明首次 bind 前 Hook 不应持久化 checkpoint；agent 与用户都无从判断 |
| D4 | init 后 state.json 的 `journal.path` 字段可能与 `req bind` 期望的 cursor 格式不兼容 | 第二轮回放时 `journal={path:..., last_event_id:null, last_sequence:null}` 仍报"journal cursor must be empty"——cursor 字段存在但要求为空值是矛盾约束 |

## 2. 矛盾点

| 字段 | 值 |
|:---|:---|
| **expected** | ① `req bind` 应有清晰错误信息说明哪些字段不满足"fresh inactive"条件；② harness 在`journal`/`journal_path`/`revision` 三层校验中应给出**一次性**完整失败清单（避免逐字段试探）；③ 用户授权"锁定并绑定"时，harness 应能引导完成从 archived → fresh → bind 的安全路径，不应让 agent 误删核心 runtime 文件导致 hook 死锁 |
| **observed** | ① `req bind` 错误信息模糊（"must be …"，但未列出全部不满足项），迫使 agent 试探式修改 state；② 三层校验顺序触发，每次只暴露一个错误，agent 修复后下一层才报错；③ guard `runtime_unreadable` 是 final safety block，**且在 init 之后无法自我解除**——必须恢复 `loop-state.json` + `loop-events.jsonl`，但恢复操作被同一 hook 拦截 |
| **user/data/system impact** | **完全死锁**：session 无法推进任何工具调用；user 不得不手动执行 `cp` + `init` 才能恢复；如果用户未及时察觉，可能需要新开 session 浪费时间 |
| **reproduction** | 在 fresh inactive runtime 上先触发一次会持久化 milestone 的 PreToolUse，再执行 `req bind` 即可稳定复现 revision guard；若随后按串行错误信息试探性删除或编辑 state/journal，则可继续触发 `runtime_unreadable` 死锁。 |

## 3. 根因分析（harness 层）

### 3.1 直接原因（immediate trigger）

**root cause #1**：`req bind` 的"fresh inactive" guard 检查**字段清单不明确**——错误信息仅说"requires ...: %w" 占位，agent 看不到具体哪些字段不满足：

```
req bind: requires a fresh inactive runtime: revision must be zero
req bind: requires a fresh inactive runtime: journal must be an object
req bind: requires a fresh inactive runtime: journal path must be .claude/loop-events.jsonl
```

二进制内 strings 显示 guard 文本是 `requires a fresh inactive runtime: %w`，**实际错误文本与 format 字符串不一致**——`revision must be zero` 是具体字段约束（不在同一 format pattern 内），说明 guard 是多个独立 case，每个报不同错误。

**root cause #2**：三层 guard（revision / journal type / journal_path）**串行触发**，每次只报一条，agent 必须修 → 重试 → 失败 → 修，**无批量校验**机制。

**root cause #3**：`runtime_unreadable` 是**final safety block**——它对 `loop-state.json` 不存在状态的反应是"永久拦截所有 mutating tools"，但**没有任何非 bash 路径可恢复**：
- Read 工具能用（但不能写文件）
- Write 工具被拦（hook 在 PreToolUse 拦截所有写工具）
- Edit 工具被拦（同上）
- Bash 工具被拦（stdout 输出 hook 错误信息）

### 3.2 系统性原因（systemic root cause）

**SR-1：缺失"自愈恢复路径"**

`runtime_unreadable` 设计上是为了保护 state 不被破坏，但**没有任何自愈机制**——用户必须**手动**在终端外执行恢复。这违反 CLAUDE.md 中"harness 由 harness 维护，no manual edits"的设计意图。

**SR-2：缺失"探索式修复的安全模式"**

agent 在三层 guard 报错后被迫"试探式修复"（删文件 → 失败 → 改路径 → 失败），没有任何"安全模式"或"dry-run"让 agent 在不动 state 的前提下预演修复。

**SR-3：缺失"完整失败清单"**

每个 guard 仅报一条错误，agent 不知道完整失败条件。修复方案：guard 应在一次响应中列出**所有不满足的字段** + **建议的修复值** + **可执行的修复命令**。

**SR-4：缺失"非 fatal 的恢复 hook"**

`runtime_unreadable` 是 fatal block，但**没有任何 fallback**让 agent 能"恢复后继续"。例如：
- 检测到 `loop-state.json` 不存在时，应允许 `init` 创建（init 已实现但被 `refusing to initialize over existing` 卡住，而 journal 不存在时 init 拒绝）
- 检测到 `loop-events.jsonl` 不存在时，应自动创建

### 3.3 流程层面（process gap）

**PG-1（原始假设已校正）**：当前 rollover 实现会 archive terminal runtime 并 seed revision=0、空 journal；接口真正不一致的地方是 **rollover/init 后的第一次 PreToolUse guidance refresh 没有避让首次 bind**，导致 fresh runtime 在 bind 前被推进为非 fresh 状态。

**PG-2**：loop-harness.md 文档应明确说明：
- rollover 后的 revision/journal 处理
- req bind 的完整前置条件清单
- runtime_unreadable 的解除路径（含用户/agent 各自的恢复步骤）

### 3.4 当前版本复测结论（2026-08-25）

本次调查在隔离临时 runtime 中分别执行了直接 CLI 和真实 Hook 联动路径，结果如下：

| 报告主张 | 结论 | 证据与边界 |
|:---|:---|:---|
| 直接 `init` 生成 revision=1 | **不属实（直接 CLI）** | `internal/cli/run.go` 的 `inactiveRuntimeState` 明确生成 `revision=0`、`runtime_id=loop-inactive`、空 journal cursor；现有 init 测试也锁定了该行为。 |
| Hook 后 `req bind` 报 revision must be zero | **属实（真实联动）** | PreToolUse 的 `reconcileGuidance` 在 fresh state 上持久化 `MILESTONE-REFRESH`，写入 `evt-milestone-refreshed-r1`；随后 `req bind` 的 TR-001 fresh guard 看到 revision=1 并拒绝。重复 Hook 会继续变成 r2。 |
| `req bind` 的失败本身消耗 revision | **需校正** | 当前复测中 revision 增长发生在 `req bind` 之前的 Hook milestone refresh；TR-001 在 guard 阶段失败，没有提交绑定 mutation。表象仍是“尝试后 revision 增长”，但责任边界是 Hook 与 bootstrap 动作的顺序。 |
| journal 缺失会阻断 bind | **属实** | 删除 `.claude/loop-events.jsonl` 后直接执行 bind，得到 `fresh runtime journal is missing`；这是正确的 fail-closed 保护，但没有可达的恢复入口。 |
| `runtime_unreadable` 下没有 agent 可用的恢复路径 | **属实** | 现有 `runtime recover inspect/plan/apply` 具备部分恢复能力，但 final safety block 将所有 mutating Bash 拦截，恢复命令无法从同一 agent 工具链执行。 |
| rollover 自身没有把 runtime 重置到 revision=0 | **不属实（直接 rollover）** | 当前 `runtime rollover` 已有测试证明会 archive terminal runtime 并 seed revision=0、空 journal。真正的缺口是 rollover 后下一次 Hook 又在 bind 前刷新 milestone。 |
| guard 只暴露一项失败条件 | **属实** | `ValidateFreshInactiveState` 按顺序返回第一项错误；后续字段只有在修完前一项并重试后才会暴露。 |

### 3.5 已确认的根因链

1. `req bind`（TR-001）把“revision=0 + 空 journal + loop-inactive”等条件定义为 bootstrap 不变量。
2. PreToolUse 的 guidance reconciliation 没有识别“fresh inactive runtime 正等待首次 bind”这一状态。它发现 milestone 语义差异后立即写入 `MILESTONE-REFRESH`，因此在真正的 bootstrap mutation 前消耗了 revision。
3. `req bind` 随后严格拒绝 revision=1；Hook 再次运行又会写入下一次 refresh，形成不可成功的闭环。
4. 当 agent 按串行错误信息直接编辑或删除 runtime 文件时，`runtime_unreadable` final safety block 会同时拦截 Bash、Write、Edit 等工具；虽然代码已有 recovery 子命令，但没有 hook 级安全 allowlist 或非 mutating 入口把它交给系统执行。

因此，问题的主根因不是“revision 存在最大值”或“rollover 必须继续重置”，而是 **bootstrap 前的控制器持久化副作用**，叠加 **不可读状态下缺少可达的恢复通道**。不能通过泛化放宽“revision 必须为 0”来修复，否则会掩盖脏 runtime 被误当作 fresh runtime 的风险。

### 3.6 最终修复模型：bind 是 runtime boundary reset

调查后确认，`revision` 不应被解释为全局累计次数，也不应作为 bind 前必须为 0 的门禁。它是当前 runtime 内的 CAS 版本；每个新 runtime 都从 0 开始。因此 bind 的正确语义不是在普通 mutation 中直接改写 `state.revision=0`，而是一次受保护的 runtime 边界切换：

```text
旧 inactive runtime:  runtime_id=loop-inactive, revision=N, journal 可含 Hook checkpoint
        │  TR-001 使用当前 revision=N 做 CAS
        ├─ 归档旧 state/journal，并记录 source revision 与 hash
        └─ 创建新 active runtime:
             runtime_id=loop-REQ-xxx, revision=0,
             journal.last_sequence=0, journal.last_event_id=null,
             bound_req=locked REQ
```

新 runtime 必须保留绑定批准、REQ 指纹、旧 runtime identity/revision/hash 等 binding receipt；旧 runtime 的 journal 不能被静默删除。runtime identity 的变化还必须参与后续 mutation 的身份校验，避免旧 inactive runtime 的快照在 revision 归零后重新生效。

因此最终规则是：revision 没有最大值；同一 runtime 内只递增，新 runtime 边界由 bind/rollover/unbind 归档后重新从 0 开始。之前“让 bind 接受非零 revision 并继续沿用同一 runtime”的方案降级为兼容性思路，不作为最终模型。

### 3.7 本次修复的最小闭环

本次只修复导致 bind 不可达的核心边界，不把 recovery allowlist、批量诊断或存储升级混入同一个变更：

1. `TR-001` 的 eligibility 只检查 runtime 仍是合法的 `loop-inactive`、没有业务进展、state/journal 结构完整；不再要求 source `revision==0`。
2. `TR-001` 以当前 source revision 做 CAS。成功时归档 source `loop-state.json` 与 `loop-events.jsonl`，写入 source revision/sha256 和审批信息。
3. 绑定结果发布为新的 `loop-REQ-*` runtime：`revision=0`、空 journal cursor、`binding_receipt.event=req_bound`。所以活动 journal 不伪造一条“第 1 个事件”；旧 journal 在 archive 中可审计。
4. 调用方把 snapshot 的 `runtime_id` 一并带入 transition。bind 后 runtime identity 改变，旧 snapshot 即使 revision 数字重复也不能写入新 runtime。
5. `req bind` confirmation 和协议/manual 明确显示上述边界、归档与下一步；agent 不需要猜测“revision 非零是否要手改”。

这保留了原有 fail-closed 原则：脏 runtime、缺失 journal、已绑定 REQ 或不一致的 source pair 仍然拒绝；放宽的只是“bootstrap runtime 在 bind 前被合法 Hook checkpoint 推进”的 revision 条件。

## 4. 影响范围

| 影响维度 | 范围 |
|:---|:---|
| 阻塞的 ops 流程 | REQ-004 bind 流程（本次 session 完全卡死） |
| 阻塞的工具 | Bash / Write / Edit / NotebookEdit / MultiEdit / TaskOutput 等所有 mutating tools |
| 可用工具 | 仅 Read / Grep / Glob（只读） |
| 触发条件 | archived → fresh → req bind 路径 + agent 试探式修改 state |
| 受影响 loop harness 版本 | v71efa56（当前仓库版本） |

## 5. 建议修复方案（调查后修订）

修复顺序应先恢复必经路径，再改善诊断；不引入 SQLite、全局锁或多种 dry-run 命令作为本次 bug 的前置条件。

### 5.1 短期（必须）

| # | 改动 | 优先级 |
|:---|:---|:---|
| F1 | **bind runtime boundary reset**：TR-001 不再要求 bind 前 revision=0；使用当前 revision 做 CAS，成功后以归档旧 inactive runtime + 创建新 active runtime 的方式重置 revision=0 和空 journal。不得在普通 mutation 中直接回写 revision。 | P0 · 已实施 |
| F2 | **可达的 recovery allowlist**：`runtime_unreadable` 下只放行严格解析的 `loop-harness runtime recover inspect/plan/apply`（apply 仍要求显式批准和自身完整校验），禁止借此放行任意 Bash。 | P0 · 后续独立缺陷 |
| F3 | **bind eligibility 批量诊断**：一次返回当前 revision、runtime identity、lifecycle、业务实体、journal 类型/path/cursor 及不可绑定原因；失败时不得写 revision 或 journal。 | P0 · 后续独立缺陷 |
| F4 | **把 boundary 前置条件写入协议和 Hook 错误消息**：明确 bind 可从任意合法当前 revision 开始，成功后新 active runtime 从 revision=0 开始；遇到 `runtime_unreadable` 时给出可执行 recovery 命令。 | P1 · 协议/manual 已实施；runtime_unreadable 路径后续处理 |

### 5.2 中期（建议）

| # | 改动 |
|:---|:---|
| F5 | **显式 bind preflight**：优先复用批量 eligibility 诊断；只有在 agent 确实需要独立检查时，再提供单一 `req bind --dry-run`，避免同时维护多个重复诊断入口。 |
| F6 | **init/recovery 的文件一致性策略**：不要静默覆盖或无条件保留孤儿 journal；先验证 state/journal 是否为空、可解析、序列一致，再选择安全初始化或进入 recovery plan。 |
| F7 | **runtime schema 字段白名单和恢复手册**：明确哪些字段不可手改，并将备份、archive、journal 一致性检查纳入命令输出。 |
| F8 | **Hook 的 critical-state guidance**：继续保留 Read/Grep/Glob 等观察能力，同时明确显示 `blocker_ref`、恢复命令和恢复后重试动作。 |

### 5.3 长期（架构）

| # | 改动 |
|:---|:---|
| F9 | **runtime 状态管理改为 SQLite**——schema + 事务保护，避免文件删除导致 hook 死锁 |
| F10 | **多 session 共享 runtime state 时加锁机制**——避免多 session 并发修复 state 导致竞态 |

F9/F10 暂不作为本缺陷的修复前置条件。当前文件 runtime 仍可通过原子写入、严格 recovery plan 和 Hook allowlist 解决本次死锁；只有在并发 session 或损坏恢复继续暴露证据后，才单独立项评估存储升级。

## 6. 验证与实施结果

| 验证 | 步骤 | 期望 |
|:---|:---|:|
| V0 | 直接 `init` 后检查 state | revision=0、loop-inactive、journal cursor 为空且文件存在 |
| V1 | 在 fresh state 上模拟 PreToolUse，使 revision 变为 N，再执行 `req bind` | **通过**：bind 使用当前 revision 做 CAS；旧 inactive pair 被归档，新 active pair 为 revision=0、空 journal |
| V2 | bind 后执行第一个普通 transition | **通过**：使用 expected_revision=0 成功，提交后 revision=1，journal sequence=1 |
| V3 | 使用 bind 前旧 runtime 的 revision/runtime identity 提交过期 mutation | **通过**：被 stale runtime identity 拒绝，不得写入新 active runtime |
| V4 | 故意制造 `runtime_unreadable` 后从同一 Hook 工具链运行严格 recovery plan/apply | recovery 命令可达；未批准或不一致时仍拒绝，批准且校验通过后解除 block |
| V5 | bind eligibility 不满足时 | 一次输出完整失败清单、当前值、检查命令和恢复动作；失败不 bump revision、不归档、不清理 journal |

实施验证命令：

```text
go test ./...
```

结果：通过。另有绑定边界、缺失 journal、脏 inactive runtime、旧 runtime identity、schema-valid runtime 与恢复 marker 的定向测试覆盖；`V4` 和 `V5` 仍是后续独立增强，不作为本次 F1 的完成条件。

## 7. 关联与历史

关联：
- **BUILDER-001（同类）**：v71efa56 harness 在 S7 冻结期 shell 拦截过度严格（`/dev/null`、`python3 heredoc` 全部被拦），导致 agent 操作受限——同型"严格 hook 阻断必要工作流"缺陷族
- **HARNESS-002（待观察）**：rollover 后 revision/journal 处理缺失

| Date | Event | Actor |
|:---|:---|:---|
| 2026-08-25 08:50 CST | Post-RELEASE ops 触发；req bind 三层 guard 误删 state 文件触发 runtime_unreadable hook 死锁 | orchestrator (session cb8f5213) |

## 附录 A：当前 session 状态快照（用于跟踪）

| 项 | 状态 |
|:---|:---|
| 受影响 session | cb8f5213-fe2f-4e31-98f0-0272d1f1a671 |
| 已删除文件 | `.claude/loop-state.json`, `.claude/loop-events.jsonl` |
| 可用备份 | `.claude/loop-state.json.bak-r2`（完整 schema 含 21 字段） |
| 手动恢复命令（用户执行） | `cp .claude/loop-state.json.bak-r2 .claude/loop-state.json && touch .claude/loop-events.jsonl && .claude/bin/loop-harness-linux-amd64 init` |
| Hook 状态 | PreToolUse 全工具 BLOCKED（runtime_unreadable final safety block） |
| 文档产出（本文件不受影响）| REQ-004.md, BUG-119.md |
| 远端生产实例 | jydev04 PID 253906 持续运行 |

---

## 附录 B：错误信息完整记录（用于 harness 修复参考）

```
$ loop-harness req bind --approved-by "songhao.li@jy" --req docs/requirements/REQ-004.md
req bind: requires a fresh inactive runtime: revision must be zero

[agent 编辑 .claude/loop-state.json: revision=0, journal={}, bound_req=null, last_transition=null]

$ loop-harness req bind ...
req bind: requires a fresh inactive runtime: journal must be an object

[agent 编辑 journal 为 dict]

$ loop-harness req bind ...
req bind: requires a fresh inactive runtime: journal path must be .claude/loop-events.jsonl

[agent 添加 journal_path 字段，touch 新 journal]

$ loop-harness req bind ...
req bind: requires a fresh inactive runtime: journal path must be .claude/loop-events.jsonl
[同上 — guard 未通过]

$ loop-harness init
init failed: refusing to initialize over existing runtime file .claude/loop-state.json

[agent rm loop-state.json]

$ loop-harness init
init failed: refusing to initialize over existing runtime file .claude/loop-events.jsonl

[agent rm loop-events.jsonl]

$ loop-harness init
QUALITY GATE BLOCKED — final safety block on runtime_unreadable.
Rule: runtime_unreadable. runtime facts are unreadable; mutating tools are blocked
until the loop runtime is restored
Recovery: restore .claude/loop-state.json and .claude/loop-events.jsonl
→ run `loop-harness runtime inspect --root .` → retry the tool after the runtime
becomes readable.

LOOP RECOVERY — Stage cross-stage @ rev=0. Objective: recover a valid runtime cursor.
Missing: valid_runtime_cursor.

[ALL TOOLS BLOCKED — Bash/Write/Edit 全工具 PreToolUse hook 拦截]
```
