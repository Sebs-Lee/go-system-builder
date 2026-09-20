# Worktree、Evidence 与 Hook 消息链路整改调查

调查日期：2026-09-18。状态：缺陷调查已完成；正式设计已迁入 blueprint 对应机制篇，运行时修复与验收进度在本文跟踪。

本文记录当前分支的四项缺陷及统一整改方案。它是工厂仓库的缺陷报告，不是设计蓝图，也不是目标项目的 REQ、BUG 或验收证据。设计规范以 `blueprint/` 对应机制篇为权威；本报告保留缺陷现象、根因、修复范围和验收证据。文中的目标行为来自本轮用户要求；与历史设计冲突的地方应同步修订历史设计，而不是继续维护旧行为。

用户澄清后的设计基准：始终按照最新 Claude Code 版本及官方机制完善设计，本机旧版本不构成设计约束。worktree 成果应先在子分支提交，再以合并提交回到项目根目录的分支；这种集成是正确的，不是 release。项目根目录始终是唯一权威，整改重点是及时合并、校验、接收和清理，不能让临时 worktree 成为独立交付终点。

后续澄清：在 REQ 绑定时显式声明开发主分支和最终发布的上游分支，不默认使用 develop，也不静默以当前分支或远程默认分支补全。worktree 以开发主分支为创建来源和回收目标；发布上游只用于最终发布。未合并、分支偏离和回收积压通过 agent 可见的警示提醒处理，不为这些状态新增严格门禁。

阶段交付要求：每一 stage 的主要 Markdown 文档、代码、测试等产出应及时 commit。stage 检查以根目录 REQ 开发主分支的已提交内容为依据；未提交内容不能帮助通过阶段检查。只有明确不入 Git 的 evidence 及 Runtime 控制状态按其专用规则从根目录读取。这项阶段资格要求与“worktree 回收仅提醒”的要求并行，不能混为工具调用硬拦截。

文件视图职责澄清：由上游 stage/gate 的输入契约逐项声明读取来源，文件视图据此读磁盘或 Git tree。上述已提交交付要求由相应上游契约表达，不在通用 FileView 中硬编码为“所有文件都读 Git”，也不由扩展名或 evidence 路径猜测来源。

## 1. 项目与调查边界

本项目不是业务应用，而是 Claude Code 的可复用工程流程模板和 Go Harness：

- `AGENTS-template.md`、`docs/agent-protocol.md`：主会话职责与 S0–S11 流程。
- `loop-template.md`：恢复与唤醒提示。
- `settings.json` → `internal/cli/run.go:evaluate` → Controller / policy → `internal/hook/`：事件、状态推进、防护、输出。
- `internal/runtime/`：Runtime、锁、持久化与 journal；`internal/qualitygate/`、`internal/transition/`：证据检查与迁移。
- `internal/integration/`：现有 worktree 检查、合并与清理。
- `internal/migration/`、schema、blueprint、skills：模板分发及契约，修改功能时须一并更新。

四项问题共有的架构缺口：没有明确区分项目的持久化交付面、子会话的临时执行面、可更新的证据内容指纹、以及人和 agent 各自的消息通道。

## 2. 根因一：没有统一的 worktree 文件边界

### 已确认事实

1. `internal/repair/artifact.go:captureRepositoryBaseline` 从根目录递归读取文件。它排除 `.git` 和部分 `.claude` 控制面，但不排除 `.worktrees/` 或 `.claude/worktrees/`。
2. `ComputeSessionChangeset` 复用该扫描器。因此 worktree 创建、修改和删除会污染项目修复差异，既可能出现额外文件，也可能出现虚假的删除。
3. `internal/review/e2e_inventory.go` 的排除表没有 `.worktrees`；子树测试文件可能被算作主项目 E2E 资产。
4. `internal/review/workspace.go:detectUndeclaredProductDrift` 的允许漂移目录没有 worktree 语义；未忽略的 worktree 路径可能被报告为未声明产品变化。
5. `internal/cli/capture_exec.go:snapshotArtifacts` 又维护一份不同的目录排除表。
6. `policy.Input` 没有保存官方 `cwd`；`reviewerRelativePath` 使用 Hook 进程的 `filepath.Abs` 解析相对路径。平台执行目录、项目根目录与子 worktree 没有统一模型。写面判定同样没有明确的 worktree 边界。

### 复现

临时测试在空项目下分别创建 `.worktrees/probe/internal/probe.go` 和 `.claude/worktrees/probe/internal/probe.go`，调用真实 `captureRepositoryBaseline`。结果两个文件都被列为项目 artifact。临时测试已移除。

### 整改设计

建立共享的 WorkspaceContext / PathScope 解析层，区分：

- `project_root`：本次主会话确认的持久化项目目录。即使该目录自身是 Git linked worktree，也不能擅自改成 Git 的主 checkout。
- `execution_root` / `cwd`：本次工具或子会话的实际执行目录。
- `git_common_dir`：判断 worktree 是否属于同一仓库的身份依据。
- `managed_worktrees`：本次项目管理的临时目录及 assignment 归属。

从 Git worktree 注册信息、项目 worktree 台账及约定临时目录共同构造排除边界，不能只匹配一个目录名。路径解析须处理绝对路径、相对路径、符号链接和目录分隔边界；不能将 `worktrees-backup` 误判为 `worktrees`。

项目级扫描、差异、资产发现和 Hook 文件检查都消费同一边界：临时 worktree 文件不计入项目交付面；归并到根目录后，才按根目录路径接受既有检查。直接对临时文件的 Hook 检查也应先分类，不得拿根目录的文件状态误判它。

这不意味着“只要 cwd 在 worktree 就跳过整个 Hook”：一次 Bash 同时改 worktree 和根目录时，只排除前者；发布动作、共享控制面变更与 worktree 生命周期仍有自己的职责。

## 3. 根因二：合并交付方式正确，但根目录权威与回收闭环不完整

### 已确认事实

| 位置 | 当前行为 | 后果 |
| --- | --- | --- |
| `AGENTS-template.md` 的 SubagentStop 指令；`internal/cli/run.go` 的 agent 激活提示 | 指定 `develop` 为来源/目标 | 未绑定主会话所在项目分支 |
| `internal/cli/controller.go:HandleSubagentStopForController` | 缺省 `targetBranch = "develop"` | 可能把工作拿到错误的分支 |
| `internal/integration/inspect.go:Inspect` | 要求 source tracked files 干净，并存在超出 merge-base 的提交 | 符合先提交再集成的要求；须补足未跟踪交付文件的核对 |
| `internal/integration/integrate.go` | 要求根目录干净，必要时 checkout 目标分支 | 保护未提交内容合理；但不应为迁就 assignment 目标而擅自切换根目录分支 |
| `internal/integration/git_worktree.go:performMerge` | `git merge --no-ff --no-verify -m ...` | 生成合并提交符合要求；`--no-verify` 与其注释冲突是独立问题 |
| `HandleSubagentStopForController` | 首次执行不 acknowledge、不 cleanup；依赖后续调用 | 一次完成事件不保证回收闭环 |
| `internal/hook/mainstop.go` | 只检查 review assignment 是否派发/消费 | 主会话可以在有待回收 worktree 时收工 |
| `settings.json` | 无 WorktreeCreate/WorktreeRemove；PostToolUse 仅 SendMessage | 缺少原生创建/移除接线和 Agent 返回后的主会话交接 |

此外，`worktreeClean` 忽略 untracked 文件。Git remove 没有实际使用 force，不能据此声称它一定会删掉新文件；但这些新文件没有进入现有以 commit diff 为核心的成果检查，本身已经是交付缺口。

现有 checkpoint 可在 `preserved` 状态原样返回，修复冲突后不能假定重复调用就会恢复；清理响应丢失也需要专门验证，而不是只依赖函数注释。

### 目标生命周期

```text
REQ 绑定的根目录主分支及其已提交基线
  → 为 assignment 创建临时 worktree
  → 子会话开发及局部校验
  → 子分支提交完整成果（含新增、修改、删除）
  → 主会话将子分支合并回项目根目录的 REQ 主分支，生成合并提交
  → 根目录联合校验 + evidence 指纹同步
  → 记录已接收事实
  → 删除临时 worktree，回收名额
  → 项目根目录分支保留权威成果与集成历史
```

归并不推进 release，不要求人工发布授权，不切换根目录分支。子分支提交及根目录合并提交是正式的任务集成方式；只有合并回根目录并校验接收后，子任务才完成交付。worktree 目录是临时资源，已集成的提交历史保留在根目录分支中。

### 落地要求

1. 创建时记录 project_root、assignment、owner session、REQ 开发主分支、基线 commit 和子分支。以显式绑定的开发主分支最新提交创建，不能硬编码 develop，也不能在根目录 checkout 变化后悄悄改用其他分支。任务依赖根目录未提交内容时，主会话应先整理所需基线，不得默默遗漏依赖或自动提交无关用户变更。
2. 子会话交付前提交完整任务成果，核对未跟踪的新文件、二进制、删除、重命名和模式变化；不能用整个 worktree 覆盖根目录，也不能把“tracked files 干净”等同于“成果全部提交”。
3. 对根目录合并串行化，保留生成合并提交的流程。合并前核对项目根目录身份和当前分支；分支发生变化时由主会话明确处理，不通过自动 checkout 消解不一致。
4. 保留根目录已有未提交内容及用户暂存状态。存在影响合并的变更时交给主会话整理，不能自动 stash、reset 或覆盖。根目录干净检查本身不是本轮要移除的缺陷。
5. 合并、校验、ack、清理的事实持久化且可重试。通过固定的 source commit、根目录 merge commit 和 Git 祖先关系核对“已合并但状态未写成功”，避免重复合并；合并后校验失败应从校验阶段恢复。
6. 校验接收后及时清理；清理失败只重试清理，不能重新合并。冲突、校验失败、子 worktree 又出现未提交内容或新提交时保留现场，不能靠强删实现数量上限。
7. Runtime、journal 和接收记录以项目根目录为唯一权威。子 worktree 的控制面副本不独立推进权威状态；生命周期记录复用现有 assignment/checkpoint，Git 注册表用于核对实际资源。

### REQ 开发主分支、发布上游绑定与温和提醒

在 `req bind` 增加两个显式声明（现已实现）：

- `--dev-branch`：REQ 开发主分支，所有子 worktree 从该分支创建，完成后合并回该分支。
- `--release-upstream`：最终发布的上游分支，开发成果完成验收后进入该目标；不作为日常 worktree 回收目标。

两者均由绑定操作明确声明，不默认 develop/main/master，不默认为根目录当前分支，也不从 Git tracking upstream 推断发布目标。缺失时要求补全绑定信息；这属于初始化参数完整性，不是新增运行期未合并门禁。

绑定至少记录项目根目录身份、开发分支完整引用、发布上游明确引用（涉及远程时包含 remote）和绑定时 commit；这些是 Runtime 绑定事实，不要求修改已锁定 REQ 正文。绑定时 commit 仅用于溯源，后续创建 worktree 应使用开发主分支的最新提交，不能永远从初始 commit 派生。assignment 的 target_branch 从该绑定派生，不再独立缺省为 develop。

```text
REQ 开发主分支 → 子 worktree 分支 → 提交并合回 REQ 开发主分支
                                           ↓
                                  最终发布到声明的上游分支
```

声明发布上游不等于授权发布，既有人工发布边界保持独立；日常 worktree 集成不进入发布流程。

此处“主分支”是该 REQ 的集成分支，可以是 feature/bugfix 分支，不特指 Git 的 main/master。后续根目录 checkout 到其他分支，不自动改变 REQ 的绑定。合法更换集成目标应显式更新绑定；不通过自动切分支或默默改绑定来掩盖偏离。

复用 assignment/checkpoint 记录的 source commit，检查它是否已包含在绑定主分支历史中；分支名相同、子分支已经 commit 或 worktree 已删除都不等于已接收。另行标记 worktree 尚有未提交成果、源分支出现新提交、已合并但未校验、已接收但未清理。只追踪当前 REQ 所属任务；无法读取 Git 状态时提示“无法确认”，不猜测已交付或未交付。

提醒内容包含 REQ、绑定主分支、相关 assignment/worktree、当前状态及下一步。例如：“REQ-xxx 的主分支是 bugfix/example；task-a 的提交尚未合入该分支，请在根目录集成并校验后清理 worktree。”

主会话在子任务结果返回、后续工具事件或恢复会话时收到 additionalContext。相同状态去重，任务完成、分支变化或积压变化时再次提醒。Stop 本身若采用 additionalContext 会触发继续执行，不能用它伪装成纯提醒；普通收工不因本项警示被阻断或强制续跑，待办保留到下一次可消费事件。

### Hook 防护与主会话职责

- 原生创建必须接入同一生命周期服务；手工 `git worktree add` 也需纳管，不能仅覆盖 `Agent(isolation=worktree)`。
- 同 assignment 优先复用现有 worktree；创建登记需处理并发，避免同一任务重复分配而不自知。
- 项目并发预算作为积压提醒阈值；已完成待归并或待清理的 worktree 也计入占用。提醒主会话优先归并/清理，不因达到阈值新增 Hook deny 或创建硬门禁。
- Hook 将子会话完成持久化为待归并事项，并在主会话可消费的事件上提醒。未合并、分支偏离或未清理不会触发本项新增的 Stop block、exit 2 或强制续跑。
- 未登记的实际 worktree 必须可见，但不能直接推定其属于本任务并删除。根目录扫描仍须识别它们的临时/独立 checkout 边界。
- 10 秒 Hook 只负责短事务、判定和反馈。完整测试及大规模归并由主会话调用集成服务；不在 SubagentStop 内执行无界测试。当前 `context.Background()` 加同步 RequiredChecks 的做法有超时后部分完成的风险。

主会话模板应明确写出：worktree 是暂存且可丢弃的执行环境；子分支先提交，主会话及时合并回本次项目根目录分支，校验接收后清理。项目根目录是唯一权威。子会话说完成、生成报告、或仅在临时分支 commit，都不等于主会话交付完成；该集成不属于 release。

## 4. 根因三：把 evidence 的内容指纹当作不可变授权

### 已确认事实

- `internal/transition/engine.go:validateCurrentEvidence` 直接以 sha256 不一致拒绝 evidence 引用。
- `internal/qualitygate/review_round_gates.go` 将 Finding 文件哈希变化加入 `hash_mismatch`。
- `internal/qualitygate/evaluator.go` 的多个消费者遇到哈希不一致就跳过 evidence；S10 还会因 envelope 的 `audit_manifest_sha256` 不一致产生 conflict。表现不只是明确报错，也可能是证据突然“缺失”。
- `internal/semantic/validator.go:validateReviewManifestReferences` 独立比较 team manifest 的 `documents[].sha256`，存在多处绑定。
- `Store.RefreshFingerprints` 已具备加锁刷新能力，但没有作为普通 Hook 中的 evidence 自动同步步骤。它还会刷新 documents、bound_req、definition、policy 和 tasks，不能把它原样接到每次 Hook 上代替 evidence 专用同步。

### 整改设计

将“内容改变”和“证据语义失效”拆开。worktree 成果归并后，evidence 字节变化是正常事实：Hook 自动更新该文件的当前 sha256 绑定，不因此新增审批、拒绝、重注册或手工恢复门禁。

在 Controller 读取用于 gate 的一致快照之前执行 evidence 专用 reconcile：

1. 以根目录的实际证据文件为准，找到当前 evidence 及对应的当前绑定；不从子 worktree 副本刷新主项目状态。
2. 内容 hash 变化时自动刷新。同路径若还绑定在当前 manifest/index/envelope，沿明确的证据引用关系同步；嵌套 envelope 先更新内层引用，再重新计算外层文件 hash。
3. 使用既有 Runtime 锁与持久化恢复机制实现幂等更新。跨文件更新要有可恢复写入顺序；不能出现 runtime 更新了、另一个 manifest 仍旧错配而再次卡住。
4. 同步后重新读取快照，继续既有内容检查。新 hash 不代表测试通过、审批通过或发布已授权；不能顺带改结论、status、责任人或恢复已 invalidated 的证据。
5. 已归档历史不重写。锁定 REQ、契约、产品冻结基线、policy/definition 等非 evidence 的授权指纹，不因该修复被全量“洗平”。对于混合 manifest，按引用对象的语义区分，不能按字段都叫 sha256 来批量替换。
6. 文件确实缺失、JSON 已损坏或出现真实内容冲突时，保留既有可读性/语义错误。单纯 hash 不一致不再产生 blocker。

不要增加“证明本次修改确实来自 worktree 才允许刷新”的新门禁；本需求的目标是正常更新 evidence 时自动收敛，不是额外证明修改来源。

## 5. 根因四：Hook 发出了字符串，却未证明 agent 收到了它

### 源码与测试的问题

- `internal/hook/pretooluse.go` 把恢复包写在 `systemMessage` 和 allow 的 `permissionDecisionReason`，没有 `hookSpecificOutput.additionalContext`。
- `internal/hook/adapter.go:renderSystemMessage` 将 lifecycle 的 `additionalContext` 放在顶层；应按最新官方契约修正为事件专用嵌套输出形状，不以旧版是否兼容该形状作为保留理由。
- `internal/hook/lifecycle_context_test.go` 反而断言 SessionStart 不得出现 `hookSpecificOutput`。测试将现有输出当成了平台标准。
- `settings.json` 的 PostToolUse 仅监听 SendMessage，没有 Agent 返回结果到主会话的专用处理；`policy.Input` 也未建模 `tool_response`。
- `tools/claude-hook-smoke.sh` 验证命令、退出码、JSON 可解析和 CLI 版本，不启动真实 agent 会话。因此当前 smoke 通过不能证明消息进入上下文。

### 官方机制核对

截至调查日，官方区分了用户提示 `systemMessage` 和模型上下文 `hookSpecificOutput.additionalContext`；PreToolUse 的 allow/ask reason 给用户，deny reason 给模型。PreCompact 丢弃 systemMessage。SubagentStop 的反馈作用于子会话；给主会话的前台 Agent 返回提示应走 PostToolUse。WorktreeCreate 会替代默认创建动作，不能只登记一个观察器；WorktreeRemove 不能靠 systemMessage 通知 agent。

来源：[Claude Code Hooks reference](https://code.claude.com/docs/en/hooks#add-context-for-claude)、[PreToolUse decision control](https://code.claude.com/docs/en/hooks#pretooluse-decision-control)、[SubagentStop](https://code.claude.com/docs/en/hooks#subagentstop)、[WorktreeCreate](https://code.claude.com/docs/en/hooks#worktreecreate)。

本机调查时 `claude --version` 为 **2.1.268**，仅作为测试环境记录。设计始终跟随最新 Claude Code 版本及官方契约，包括最新文档描述的 SubagentHandback 报告路径；不为本机旧版本降级设计或保留旧协议。平台验收应使用当时最新版本，并记录实际版本和结果；旧环境测试不能替代最新版本验收。

### 修复边界

- 统一由事件适配器决定“给谁、哪个字段、是否继续执行”；不得把一份 systemMessage 当作所有事件的通用通知。
- PreToolUse 的日常指导用 additionalContext；“流程允许”不必主动覆盖 Claude 本来的权限判断。内部 quality_gate 结构继续保留在 outbox，agent 需要的内容通过支持的文本字段传递。
- SessionStart/SubagentStart 采用文档对应的事件输出；PreCompact 保存恢复事实，压缩后由 SessionStart(compact) 注入。
- SubagentStop 记录完成和待归并事项；前台 Agent 的 PostToolUse 给主会话提供交接。后台 Agent 的启动返回不等于完成，须依靠已持久化待办及主会话后续 Hook 消费，不能误判 async_launched。
- 按最新机制接入 SubagentHandback 的报告输入；不能继续假定 last_assistant_message 或 Agent 返回的 content 必然包含完整交付报告。
- Stop/SubagentStop 的阻断与非阻断续跑语义分别处理，避免为了“让 agent 看见”把所有消息升级成 block；worktree 新建/移除按事件特有协议适配。
- 按最新版本维护事件能力表及平台契约测试；更新 settings、迁移白名单、protected_events、schema、blueprint 和测试，保持分发结果一致。

## 6. 实施顺序与验收

建议分四个可独立检查的改动，但最终按完整链路验收：

1. 修复 agent 消息通道与平台契约测试；补齐 cwd/tool_response 等实际需要的官方输入。
2. 统一 worktree 边界，接入扫描器、Hook 路径分类与共享控制面定位。
3. 在 REQ 绑定时显式记录开发主分支与最终发布上游，完善 worktree 的创建、合并、ack、清理与主会话积压提醒；移除 develop 硬编码及隐式分支默认值，保留子分支提交和根目录合并提交，落实根目录唯一权威，不为未回收状态新增硬门禁。
4. 引入 evidence 专用自动指纹同步，连接归并完成和 Controller 快照读取；同步相关绑定消费者。

验收至少覆盖：

| 场景 | 预期 |
| --- | --- |
| 两类默认路径、自定义目录、外部 linked worktree | 项目扫描不计入暂存文件；根目录文件仍检查 |
| REQ 绑定分别声明开发主分支与发布上游 | 两者独立持久化；缺失不静默补为 develop、当前分支或 Git tracking upstream |
| worktree 创建与回收，发布上游与开发主分支不同 | 只使用开发主分支；不合入发布上游，不触发发布 |
| 同一命令同时改 worktree 和根目录 | 只排除暂存写面 |
| 从当前功能分支创建，根目录已有待提交变更 | 来源分支正确；明确已提交基线及任务依赖，不切 develop，不覆盖用户变更 |
| 两个 worktree 先后提交交付，有新文件、删除和二进制 | 子分支提交完整，根目录生成合并提交并保留全部成果 |
| 子 worktree 遗漏未跟踪文件或在接收后新增变更 | 不把旧 source commit 的接收事实当作新内容已接收，不误清理 |
| 子 worktree 存在 Runtime/journal 副本 | 所有权威读写定位根目录，不产生分裂的控制面 |
| 子任务一次完成，无第二次 SubagentStop | 主会话可见并完成归并、校验与清理 |
| 重复事件、并发创建、归并后崩溃、清理响应丢失 | 正确统计积压并提醒，不重复合并、不丢文件；可恢复 |
| 绑定 REQ 后根目录切到其他分支，或成果合到了其他分支 | 绑定目标不变，准确提醒偏离/未回收；不自动 checkout、不新增 deny |
| 未合并或积压提醒反复触发、主会话 Stop | 提醒去重，普通收工不被本项警示阻断或强制续跑 |
| 冲突、源文件归并后再次变化、未知归属 worktree | 不误删；指出具体待处理事项 |
| 多个任务合并同一 evidence 文件 | 内容可正常合并时自动同步全部当前 hash 绑定，Hook 继续 |
| evidence 同步前后重复 Hook | 无多余语义迁移；不会凭刷新 hash 生成通过结论 |
| 非 evidence 的 REQ/契约/产品基线被改 | 不因 evidence 同步而自动消除原有检查 |
| 真实 Claude 会话中注入唯一测试标记 | agent 能在下一轮使用标记；终端显示不算成功 |
| 前台、后台 Agent；启动、压缩恢复、停止 | 提醒正确接收方，不误把后台启动当交付 |
| 最新 Claude Code 的 SubagentHandback 报告路径 | 读取真实报告输入，不把结束语误作报告；记录最新版本平台验收结果 |

## 7. 本轮验证结果与未验证边界

- 相关六个包的现有测试通过：`go test ./internal/hook ./internal/integration ./internal/qualitygate ./internal/runtime ./internal/repair ./internal/review`。
- 临时扫描探针复现了两类 worktree 目录进入产品 artifact 的问题，探针已清除。
- 已查询本机 Claude 版本及最新官方 Hook 文档。
- 未运行会消耗实际模型请求的 Claude 会话；agent 接收、后台交接、原生 worktree 清理事件的端到端行为仍属实施后的平台验收项。
- 本轮只新增此调查设计文档；没有修改运行时代码、提交 Git 或操作用户 worktree。现有测试通过是调查基线，不是修复已完成。

## 8. S7–S8 未提交产出进入 S9 的边界调查

### 8.1 已实测的 Git 行为

在独立临时 Git 仓库中，从明确的本地开发分支创建真实 worktree，结果如下。实验结束后临时仓库已删除，未操作当前项目分支或 worktree。

| 根目录状态 | 新 worktree 中的结果 |
| --- | --- |
| report.md 已跟踪，内容从 v1 改成 v2，未提交 | 仍是已提交的 v1 |
| staged.md 已跟踪，v2 已 git add，未提交 | 仍是已提交的 v1 |
| new-contract.md 新建、未跟踪 | 不存在 |
| .claude/loop-state.json 被忽略 | 不存在 |
| .env 被忽略 | 不存在 |
| worktree 创建后，根目录再提交上述报告和契约 | 旧 worktree 仍停在原提交，不自动更新 |

实验还确认：子 worktree 的 `.git` 是指针文件，两个工作目录的 git common dir 相同。Git 历史共享，不代表磁盘文件、HEAD 或暂存区共享。

因此，绑定 REQ 开发分支解决的是“从哪里分支、合到哪里”，并没有解决“派发时需要的输入是否在那个提交里”。不能只检查分支名就宣称 Builder 获取了最新 S7–S8 产出。

### 8.2 Claude Code 原生机制与框架责任

最新官方文档明确：原生隔离默认使用 Git worktree；Claude 增加会话绑定、文件/命令隔离、环境文件携带和资源清理。配置 WorktreeCreate 时才由自定义 hook 替代原生创建，并需要自行处理文件携带。并不存在默认将项目工作区完整复制给子会话的承诺。

需按最新文档核对的边界包括：默认 fresh 与本地 head 的区别、head 在嵌套 worktree 中的含义、`.worktreeinclude` 只选择被忽略文件、原生隔离对根目录写入及 Git 重定向的限制、活动 worktree 锁及不同来源 worktree 的清理规则。默认环境不是完整进程/服务隔离，不能仅靠目录隔离推定端口或数据库隔离。

来源：[原生 worktree](https://code.claude.com/docs/en/worktrees)、[子会话隔离](https://code.claude.com/docs/en/sub-agents#write-subagent-files)、[Git worktree](https://git-scm.com/docs/git-worktree)。这里对 Claude 行为的判断来自官方契约，不是对最新 Claude 二进制内部调用的跟踪实测。

### 8.3 项目现有 S9 链路的缺口

`internal/cli/repair_command.go` 的 dispatch 将 RepairContract、RepairSession、RepairPlan 等以相对路径和指纹写入 manifest 的 documents/read_paths；它并未在这里物化一份完整的 worktree 输入包。Builder 从隔离目录解释这些相对路径时，可能读到旧版本或找不到文件。

`internal/cli/worktree_shared_control_plane_test.go` 验证了显式 `--root <项目根目录>` 可以把 CLI 写入集中到根目录。这是进程内 CLI/Git 测试，不是 Claude 原生隔离环境中的端到端测试。它不能证明新平台允许所有“子会话去根目录执行 CLI”的调用，也不能证明测试执行目录正确。

`internal/repair/runtime.go`、`internal/repair/artifact.go` 的多个调用以同一个 root 读取权威输入、产出报告或计算修复变化。实施时必须审查每个 root 参数的职责：权威状态所在目录和实际被测代码目录应分开，不能一律替换为根目录，也不能一律改为 worktree。

可能出现的错误闭环：根目录 Runtime 是新的，Builder 代码却是旧的；或者 Builder 修好了子目录代码，校验命令实际在根目录测试未修复代码。前者导致错误修复基线，后者导致错误测试结果。这两项是由现有目录模型推导出的风险，尚未作为真实 Claude 会话故障复现。

### 8.4 输入交接方案（按已提交阶段产出要求修订）

为每次 S9 assignment 明确下列输入及权威归属；主要阶段产出先提交再作为下阶段依据，输入副本不再作为未提交正式文档通过 stage 的替代路径：

| 对象 | 传递方式 | 权威归属 |
| --- | --- | --- |
| 产品代码及需要版本控制的测试/文档基线 | 从 REQ 开发主分支明确的 commit 检出；需要的未提交代码由主会话先整理成范围明确的 checkpoint commit | 根目录开发主分支 |
| 主要 S7–S8 Markdown 报告、修复文档及其他需版本控制的阶段产出 | 提交到开发主分支后随代码基线检出；不能通过复制未提交文件绕过 stage 检查 | 根目录开发主分支的已提交版本 |
| 明确不入 Git 的 evidence，例如任务所需的运行证据及控制面旁文件 | 派发时按显式依赖清单读取根目录原件或导出只读快照；记录原始路径、版本及哈希 | 根目录原件；副本仅供读取 |
| Runtime、journal、生命周期推进和合并接收 | 保持单一根目录控制面；由主会话或受控 Hook/服务接收子会话报告后写入 | 根目录，禁止通过 Git 合并运行态副本 |

主要文档和代码必须及时 commit，但不要求把 Runtime 和所有临时文件入库，也不能 blanket `git add -A`。阶段 commit 和发布是两件事。evidence 例外必须按明确的 artifact 类型/目录职责登记，不能因为文件位于 docs/reports、被 gitignore 忽略或叫 evidence 就自动豁免。只读运行证据副本不参与 Builder 的交付 commit；其输出在根目录登记前先由主会话导入并归一化引用路径。输入包只补充允许不入 Git 的输入，不替代已提交文档基线。

先确定任务输入快照，再派发 Builder；创建 worktree 后根目录新增的输入需显式更新交接，不能假设实时同步。快照带上权威版本，输入变化时通知主会话和 Builder，不因可合并 evidence 的 hash 变化新增门禁。根目录 evidence 自动刷新仍按第 4 节执行，但刷新 hash 不会让正在工作的 Builder 自动读到新内容。

保持只读输入快照与动态 Runtime 分离，可以避免用整个 `.claude` 的复制或软链接来“修复不可见性”；前者会制造第二份运行态，后者会让子会话直接共享可变文件并模糊写入责任。

### 8.5 边界清单与验证要求

| 边界 | 具体危险 | 处理方向 |
| --- | --- | --- |
| 未提交主要报告或契约 | 文件缺失，或更隐蔽地读到旧版 | 不能作为 stage 通过依据，提交到开发主分支后再消费 |
| 未提交产品代码/复现测试 | S7 检查的基线不同于 S9 修复基线 | 明确代码 commit，主会话整理任务所需变更 |
| 普通 untracked 文件 | 误以为 .worktreeinclude 能携带任何文件 | 正式阶段产出先提交；仅明确豁免的证据由框架交接 |
| 根目录在派发后继续变化 | 同分支名下已是另一个代码/证据版本 | 记录实际 commit 与输入快照版本，显式通知更新 |
| 子任务 B 依赖 A | A 已完成或已提交，但 B 创建时尚未合入 | A 合回开发主分支后再为 B 确定基线；旧 B 不自动更新 |
| 多 worker 读写同一控制面 | 丢写、旧状态覆盖或各自推进 lifecycle | 单一根目录 writer/锁；子会话报告由权威入口消费 |
| root/cwd 混用 | 根目录测试替代子目录测试，或报告登记到子目录 | 分离 control root 与 execution root，证据记录被测 commit/目录 |
| 子会话尝试返回根目录合并 | 与平台原生隔离冲突 | 由根目录主会话接收合并；不让 Builder 绕过隔离 |
| 手工创建、自定义 hook、原生创建混用 | 来源、环境初始化和清理责任不同 | 同一台账核对来源，逐条验证，不假设平台代为清理 |
| 只检查已提交 diff | 遗漏未跟踪的新实现/新报告 | 子任务交付清单与文件状态对账，接收后才清理 |
| 提前清理或旧会话恢复 | 活动任务失去目录，恢复时实际 cwd 变化 | 尊重活动锁；恢复时重新定位；未接收成果不删除 |
| Git 历史共享 | 强改共享 refs 或删除分支影响其他 worktree | 任务分支唯一归属，不把 worktree 当完整仓库沙箱 |
| 环境文件、依赖目录或服务共享 | 本地测试互相污染，测试环境不一致 | 显式初始化任务环境，隔离写入的缓存/端口/数据库 |

验收增加一条完整路径：S7–S8 生成主要文档与运行证据 → 主要文档未提交时不能支撑阶段推进 → 文档 commit、运行证据按例外登记 → S9 compile/dispatch → Builder 从正确 commit 和允许的 evidence 输入读取，运行红绿测试 → 主会话接收报告、合并提交、根目录复验 → evidence 绑定自动刷新 → 清理。需要同时证明权威 Runtime 始终只有根目录一份。

## 9. 上游输入契约决定文件视图，交付检查消费已提交产出

### 9.1 源码定位

`internal/controller/cycle.go:diskFiles` 是生产 Quality Gate 的文件视图，ReadFile/ReadDir 直接调用根目录的 os.ReadFile/os.ReadDir；它并不区分 commit、index 和工作区。因此未提交的新文档、修改或暂存内容有机会成为 gate 的有效输入。这是“主会话通过 stage，但新 worktree 缺少该 stage 产出”的根因之一。

`internal/qualitygate/evaluator.go:FileView` 已提供统一读取接口，是整改入口；但不是唯一入口。Transition Engine、语义校验、目录发现和其他领域包还直接读取磁盘，需要逐个核对。只改 ReadFile 而保持 ReadDir 扫工作区，或 gate 读 Git 而 transition 又读磁盘，都会继续产生两套事实。

### 9.2 上游声明与读取职责

上游 stage/gate 定义或其明确引用的 artifact 契约，负责声明每项输入的来源要求；Controller 结合 REQ 的分支绑定解析具体来源；文件视图只执行该要求，Evaluator 负责检查内容。不能让文件视图自行决定“什么算已经交付”。

拟议的每项输入要求包含：artifact/path 或发现范围、`source`（`git_tree` / `disk`），以及对应的来源坐标。`git_tree` 在本次求值时解析成明确的 repository + commit SHA；`disk` 明确权威 root 或执行 root，并记录实际读取内容的版本/哈希。具体字段名在实施时与现有 catalog/schema 对齐，不另起一套并行规则。

| 上游检查要求 | 文件视图行为 | 结论边界 |
| --- | --- | --- |
| 主要阶段产出必须已提交到 REQ 开发主分支 | 从绑定分支解析出的固定 Git tree 发现并读取 | 可用于正式阶段交付资格 |
| 明确不入 Git 的运行 evidence | 读取根目录指定磁盘证据及绑定 | 可用于对应 evidence 检查，按既有规则刷新哈希 |
| 开发中的草稿预检、当前执行现场检查 | 上游明确要求时读取指定工作区磁盘 | 只描述当前工作区，不替代正式已提交交付结论 |

同一 gate 可以同时消费 Git 中的正式文档和磁盘 evidence；同一路径也可在不同检查中被要求读取不同来源。路由至少包含输入身份/来源要求，不能仅以 path 作全局缓存键。文件发现、ReadDir、ReadFile、哈希和语义校验必须沿用同一来源，避免目录来自磁盘、内容来自 Git 的混合视图。

来源要求必须由上游契约传递，不由 Builder 的工具输入、自称 evidence 或文件是否 gitignored 临时决定。来源不明确时修正或报告上游契约缺失，不静默默认磁盘；指定来源中缺少文件时也不自动换另一来源兜底。

### 9.3 已提交交付契约的读取规则

以下规则适用于上游声明“必须已提交”的正式阶段输入，不是所有 FileView 调用的全局模式：

1. 开始一次需要已提交输入的 stage 检查时，解析根目录 REQ 开发主分支的明确 commit SHA。根目录当前 checkout 与绑定分支不同则提醒偏离，不能悄悄把其他分支的 HEAD 当成该 REQ 的开发事实。
2. 上游要求已提交的产出，其发现、读取、内容哈希和语义校验统一使用该 commit 的 Git tree/blob。不能只证明某个路径曾经 commit，然后仍读取磁盘最新内容；已跟踪文件的未提交修改同样不能通过这种方式被采信。
3. 新文件未 commit、仅 git add、commit 只在未合回的子分支上，均不能作为根目录开发主分支已经完成该阶段的依据。普通文件不在已提交树中时返回 not_ready，提示所缺产出及提交动作，不回退磁盘读取。
4. 只有显式允许不入 Git 的 evidence 走根目录证据视图，并按第 4 节自动刷新其对应 sha256。Runtime/journal 继续由单一控制面读取，不要求提交它们。禁止用 evidence 自动刷新覆盖普通文档的已提交版本绑定。
5. stage 所需版本、当前 generation、文档注册信息及证据引用需与该 commit 中的实际内容匹配。工作区的 v2 不能弥补已提交 v1 的不足；如果 v1 本来就满足当前阶段要求，检查结论只声明 v1 满足，不能宣称未提交 v2 已交付。
6. 一次判定与紧随的状态迁移必须使用一致的 commit 快照；正式推进前重新核对分支引用，发生变化则重新评估。Runtime CAS 本身不能证明 Git 分支未移动。
7. 测试/验证证据必须描述实际被测代码版本。若测试在带未提交产品修改的工作区运行，其通过结果不能冒充已提交代码的验证结果；应使用对应提交的干净测试快照或等价的内容对账。仅给结果填写 HEAD 字符串不够。

### 9.4 工作节奏与反馈

主会话在各 stage 的正式产出完成后，提交本阶段范围内的主要文档和代码，再由 Hook 重新评估。Hook 不代替 agent 执行 blanket add/commit，也不阻断它为修复 not_ready 而进行的编辑、测试和提交。

未提交产出只导致 stage 尚不具备推进条件，通过 additionalContext 提醒“哪些正式产出需要提交”；不因此拒绝普通工具调用。worktree 未清理等资源积压仍仅提醒；但某 stage 确实依赖的交付代码没有合回开发主分支时，该阶段自然缺少合格输入，不能拿子分支提交代替它。

处理“阶段迁移自己生成产出”的循环依赖：先物化需版本控制的正式文档，通知主会话 commit，再检查并登记迁移；Runtime 投影和明确的运行证据可继续由迁移事务生成。不能要求一个尚未执行的迁移先生成并提交其产物，也不能为打破循环而允许未提交正式文档过关。

提交后应有明确的重评估时机，例如经评审接线的 PostToolUse(Bash) 或下一次正常 Hook；PreToolUse 阶段不能在 commit 尚未执行时假装它已成功。提交到开发主分支无需 push，也不等于 release。

### 9.5 必须覆盖的回归场景

- 同一路径 Git 为 v1、磁盘为 v2：上游要求 git_tree 时读 v1，要求 disk 时读 v2，不串用缓存。
- 同一 gate 的正式文档来自 Git、运行 evidence 来自磁盘，分别遵循各自上游契约。
- 指定来源中缺文件而另一来源存在时，不回退；缺少来源声明时不隐式使用磁盘。
- 工作区草稿预检通过，不被提升为已提交阶段交付通过。

- 未跟踪的新 Markdown 看起来完整，stage 仍不能消费它。
- 已跟踪文件在工作区或暂存区被补齐，已提交版本仍不完整，stage 不通过。
- 同样的产出提交到开发主分支后成为候选输入；其他内容条件仍照常校验。
- 文件只在子 worktree 分支提交，未合回，不计入开发主分支阶段交付。
- 允许不入 Git 的 evidence 可正常被消费和更新哈希，不要求 Runtime/journal commit。
- 普通报告不能通过放入忽略目录或自称 evidence 绕过提交要求。
- Git tree 目录发现不包含工作区新文件；transition 不重新读取另一份磁盘版本。
- Git 分支在 gate 判定和迁移之间移动，不提交基于过时快照的推进。
- 产品代码未提交，但工作区测试成功，不能用该结果证明提交版本通过。
- 无关工作区脏文件不触发全仓 clean 门禁；阶段资格仅依据其正式输入与证据。


## 10. 正式设计落点与实施记录

本报告留在仓库根目录，记录缺陷与验收；正式机制已经归入以下设计文件：

| 范围 | 设计权威 | 实现状态 |
| --- | --- | --- |
| 根目录权威、REQ 开发/发布分支、worktree 生命周期与回收 | [L4 Worktree](blueprint/L4-worktree-governance.md)、[S1](blueprint/L3-S1-bind.md)、[S6](blueprint/L3-S6-build.md)、[S9](blueprint/L3-S9-repair.md) | 已实现；见下方验证范围 |
| 阶段交付、上游逐输入来源、Gate → transition 一致快照 | [L2](blueprint/L2-lifecycle-plan.md)、[L3 共用契约](blueprint/L3-README.md)、[L4 状态机核心](blueprint/L4-state-transition-core.md) | 已实现；见下方验证范围 |
| Evidence 汇总后的定向 sha256 自动同步 | [L4 运行时控制面](blueprint/L4-runtime-control-plane.md) | 已实现；见下方验证范围 |
| Agent 可见消息、主子会话与事件边界 | [L4 Hook 接线](blueprint/L4-hook-platform-wiring.md)、[锚点目录](blueprint/L4-hook-anchor-catalog.md)、[调度](blueprint/L4-agent-dispatch-governance.md) | 已实现；见下方验证范围 |

实施顺序：缺陷报告 → 正式蓝图 → 安装规范/契约/schema → 实现与针对性回归 → 最新 Claude Code 平台验收。蓝图修订不代表运行时代码已满足设计；后续代码提交应逐项回填实际测试结果，不以文档链接替代验收证据。

### 10.1 已实现的对应落点

- `internal/workspace`、`internal/cli/worktree.go`：双分支显式绑定、历史绑定补全、根目录身份核对、按绑定提交创建与复用、Git 实际 checkout 归属检查、主会话事实去重提醒。
- `internal/pathscope`：约定临时目录、Git 注册的自定义/外部 worktree、符号链接和目录边界；根目录扫描与 Hook 目标分类消费同一实现。
- `internal/fileview`、Controller、transition、契约/任务/场景语义检查：上游 `file_sources` 驱动目录发现与文件读取，正式产出使用固定提交，迁移前后核对引用；旧 Runtime 缺少绑定不能回退磁盘推进。独立草稿检查明确选择磁盘。
- Integrator：主会话显式回收、根目录串行执行、普通合并提交、检查后再次确认工作区未被修改、接收确认与非强制清理；已合并失败可重试，合并响应丢失按双父提交核对恢复，清理响应丢失可幂等完成。
- Runtime evidence reconcile：仅当前有效类别，内层引用先同步再计算外层哈希；复用 fingerprint pending marker 恢复文件与 Runtime 更新。受控范围之外的产品/REQ/冻结主体不更新。并发证据变化保留新内容，撤销过时刷新意图后重算，不新增来源证明门禁。
- Hook：采用官方事件专用 `hookSpecificOutput.additionalContext`；普通建议不覆盖平台权限决定；PreToolUse deny 用官方 JSON/exit 0；PreCompact 只持久化，Stop 家族不以软提醒强制续跑；处理 Agent 异步启动边界和结构化 SubagentHandback 报告。
- 恢复计划也显式声明 `--dev-branch` / `--release-upstream`，将绑定纳入计划指纹，重放消费提交视图。

### 10.2 验证范围

回归覆盖 Git/磁盘混合来源、暂存/未跟踪文件不合格、分支移动、worktree 路径隔离与根目录保护、绑定创建/重试复用、普通双父合并与清理、校验产生新修改后的保留/重试、合并响应丢失恢复、证据引用内外层同步及冻结主体保护、Hook 官方输出结构与提醒去重。2026-09-18 验证结果：`go test ./...` 全量通过；`go run ./cmd/loop-harness doctor --root .` 通过结构 schema、示例、语义链接与生成手册一致性检查；`git diff --check` 通过。补充的子目录扫描排除已纳入最终全量回归。

最新 Claude Code 的真实父/子 Agent 接收验收尚未执行。本轮依据官方最新契约改代码并验证输出/流程；这不能替代最新平台会话中的实际接收实测，也不将本机旧版本视为设计基准。

## 2026-09-19 · 审查反例整改

针对落地审查中复现的四个缺陷，补齐公共 Runtime Writer 的权威目录/存储坐标校验（含 pending 恢复与显式离线恢复能力）、历史双父合并回执恢复、按成功阶段恢复清理，以及登记 worker 的 PLAN_REPORT 导入与最终失败反馈。四个反例已转为默认回归，并增加自定义存储坐标、错误合并、worker HEAD 变化、未知 worker 和符号链接等反向用例。

修复前证据、实现说明与验证记录见 [`L4-worktree-hook-remediation-review.md`](L4-worktree-hook-remediation-review.md)。这次修复不把原生 worktree 创建接线、S9 完整输入交接及被测提交绑定的待验收项自动视为完成；真实 Claude 父子会话验收仍须单独完成。
