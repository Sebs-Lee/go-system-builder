# L4 Worktree / Hook 整改落地审查

审查日期：2026-09-18。审查对象：当前工作区（包含原有未提交实现与新增文件），不是仅审查 HEAD。

最新状态（2026-09-19）：F1–F4 已修复，原始反例已加入默认回归；最终 `go test ./...`、doctor 和 diff 检查均通过。修复记录位于文末，平台及完整交付链的剩余验收边界保持明确。

初次审查结论（修复前）：基础机制和多条正常链路已经落地，但尚不能判定整改完成。本轮确认 2 个 P1、2 个 P2：根目录唯一写者、worker 报告交接、回执恢复和清理恢复仍有缺口；原有全量测试通过不足以覆盖这些验收边界。

本文件是模板工厂自身的审查记录，不是目标项目的 REQ/BUG 或发布验收证据。审查期间未修改生产实现、原有设计文档或用户的 Git 分支/提交。

## 架构与审查口径

项目是 Claude Code 工程流程模板与 Go Harness。主会话依照 AGENTS/agent-protocol 驱动 S0–S11；loop 提示负责恢复；settings 注册的 Hook 经 CLI、Controller、Quality Gate、Transition Engine 进入 Runtime Writer。Runtime/journal 是唯一状态权威，worktree 是执行面，Integrator 负责根目录接收。

本轮阅读 README、docs/README、L2 生命周期、L4 worktree/状态机/运行时/Hook 机制与根目录整改报告，将 §3–6、§8–9 的要求按以下路径核查：

1. REQ 显式分支绑定 → 固定提交 → worktree 创建/复用。
2. 子目录执行与提交 → 根目录合并 → 联合检查 → 接收 → 清理及重试。
3. 上游来源声明 → Git/disk 读取 → Gate → 同快照 Transition。
4. 可变 evidence → 引用闭包刷新 → 冻结主体保护。
5. 父子 Hook 输入 → 报告登记 → Agent 可见反馈。

按用户指定，三个子代理均使用 `gpt-5.6-luna`、`max`，分别执行 worktree、阶段来源、Hook/evidence 审查与测试。主代理负责跨模块审查与独立反例。

## 已确认问题（以下保留修复前证据，后续修复见文末）

### F1 · P1：Runtime Writer 没有统一校验绑定的权威目录

位置：`internal/cli/run.go:2124`；`internal/runtime/store.go:1142`。

`runtime evidence add` 直接使用调用者的 root/state/journal 创建 Writer。Writer 的锁仅绑定传入的 state 路径，没有检查它是否属于 `bound_req.workspace.project_root`。新增的 `ValidateAuthority` 目前只在 worktree 创建、部分 Controller/Transition 入口使用，没有成为所有写操作的不变量。

真实 Git + CLI 反例：创建合法已绑定 Runtime → 创建 linked worktree → 将根目录 Runtime 原样复制过去（绑定仍明确指向原根目录）→ 在 worker 中登记 evidence。命令 exit 0，输出 `recorded=true, revision=1`；子 Runtime 改变，原根 Runtime 字节完全不变。因此这不是仅“可能错写”，而是已经复现了独立写者。

复现：

```bash
L4_AUDIT_REPRO=1 go test ./tests/system/req039 \
  -run '^TestAuditCopiedRuntimeCannotRecordEvidence$' -count=1 -v
```

验收断言当前失败。测试位于 `tests/system/req039/audit_authority_e2e_test.go`，默认跳过未修复反例；显式开启后要求副本被拒绝且无写入。建议在公共写边界验证权威 root 与实际 state/journal 坐标，并覆盖所有恢复、证据、任务、审查及修复写入口；不要仅再补单个 CLI 分支。

### F2 · P2：合并回执恢复只识别当前 HEAD，分支前进后无法闭环

位置：`internal/cli/controller.go:1365–1379`。

目前恢复逻辑仅检查开发分支当前 HEAD 是否恰好是 target/source 的双父合并提交。实际合并已成功但 receipt 尚未落盘时，如果开发分支随后出现另一个普通提交，合并仍在祖先历史中，却无法恢复。后续 Inspect 又因源分支已无未合并提交而拒绝，持久 checkpoint 留在 `ready`。

反例保留真实 worker checkout，模拟合并回执丢失，并在根分支追加提交后通过 `runtime task-integrate` 重试。`TestL4AuditLostMergeReceiptAfterTargetAdvance` 当前失败。应沿绑定开发分支历史查找可证明身份的 merge，并核对记录的两个父提交及祖先关系，不能只检查 HEAD 或重新合并。

### F3 · P2：清理失败重试会重复运行已通过的联合检查

位置：`internal/integration/integrate.go:150–160`、`internal/integration/integrate.go:358` 的 `preserveAfterCAS`。

所有失败统一落 `preserved`；只要已有 merge，恢复就退回 `merged`，丢失之前已经 verified/acknowledged 的阶段。测试先完成校验，再使 worker 出现未跟踪文件触发清理保留，恢复干净后重试：RequiredChecks 运行计数从 1 变成 2。根合并没有重复，但这仍违反“清理失败只重试清理”；昂贵或带副作用的校验会被重复执行，甚至让已经接收的任务重新卡在校验阶段。

`TestL4AuditCleanupRetryDoesNotReRunChecks` 当前失败。建议保留最后成功阶段或将清理失败保持在 `cleanup_pending`，只重复清理所需的当前文件/HEAD/活动锁检查。

两个恢复反例的复现命令：

```bash
L4_AUDIT_REPRO=1 go test ./tests/system/req039 \
  -run 'TestL4Audit(CleanupRetryDoesNotReRunChecks|LostMergeReceiptAfterTargetAdvance)$' \
  -count=1 -v
```

### F4 · P1：worker 本地报告没有导入根控制面，且拒绝后仍给 Agent 成功提示

位置：`internal/cli/run.go:2758–2762`、`internal/cli/run.go:2822–2845`；`internal/hook/posttooluse.go:140–147`。

Builder 定义要求 worker 写 PLAN_REPORT 再传 `plan_ref`。Hook 的 `--root` 仍是权威根，payload `cwd` 才是 worker；但 `validatePlanReportCheckpoint` 将相对引用直接拼接根目录，未按已登记的 worker 坐标读取并导入，也没有输出引用归一化流程。只在 worker `.claude/plans/plan.json` 存在的有效报告被报为根目录文件缺失。

独立复现使用真实 linked worktree 和 CLI Hook，并设正向对照：同一合法报告放在根目录时能登记；删除根目录副本而保留 worker 文件后，Hook exit 0、stderr 为 `plan_report rejected: read plan_ref ... no such file or directory`。此时 stdout 的 additionalContext 却仍为 `plan_report observed for agent-s9 ...`：失败分支改了 Recorded/Reason，没有清除之前生成的 SystemMsg，渲染器优先输出旧消息。Agent 看不到正确的失败与恢复动作。

复现：

```bash
L4_AUDIT_REPRO=1 go test ./internal/cli \
  -run '^TestAuditHooksE2EPlanRefUsesAuthorityRoot$' -count=1 -v
```

测试位于 `internal/cli/audit_hooks_e2e_test.go`。子代理最初测试误断言了 wire 中不存在的 `recorded` 字段；主代理已修正为业务验收断言、增加真实 Git worktree，并独立复跑确认上述失败。因此不把旧测试本身的错误算作生产缺陷。

建议通过受控入口根据 assignment/cwd 核对 worker 身份、导入合法输出、归一化并登记根目录引用；读取相对路径不能简单全局改为 cwd。失败输出必须以最终观察结果重建，避免拒绝后仍提示已收到。

## 正常路径与新增测试

| 案例 | 层级 | 本轮结果 |
| --- | --- | --- |
| 子分支新增源文件、二进制、删除文件，双父合并后 ack/清理 | 真实 Git + Integrator；此案例未配置业务 checks | 通过 |
| 子目录遗漏 untracked 成果，保留现场；提交后重试 | 真实 Git + Integrator | 通过 |
| 根目录存在 dirty 文件，拒绝合并；整理后重试 | 真实 Git + Integrator | 通过 |
| 合并后检查失败，重试沿用 merge receipt | 真实 Git + Integrator，注入可计数失败的 CheckRunner | 通过，无重复合并 |
| 根 checkout 偏离绑定分支，按绑定 commit 创建并复用 | 真实 Git + CLI `worktree-create` | 通过 |
| 根目录接收并清理；回执丢失但 merge 仍为 HEAD | 真实 Git + CLI `task-integrate` | 通过 |
| formal 文档只在子分支提交，或根目录 untracked/staged | 真实 Git + CLI PreToolUse/Controller | 均 `not_ready`，不迁移 |
| 根目录绑定分支提交 formal 文档，同时消费 disk evidence | 同上，Hook cwd 指向 worker | `transition_committed=true`，推进到 planning/contracts |
| 复制 Runtime 后在 worker 登记 evidence | 真实 Git + CLI | 失败，F1 |
| 回执丢失后 dev 前进 | 真实 Git + CLI | 失败，F2 |
| 清理失败恢复仅重试清理 | 真实 Git + Integrator | 失败，F3 |
| worker 本地合法 PLAN_REPORT 交给根目录 Hook | 真实 Git + CLI PostToolUse；根目录同报告为正向对照 | 失败，F4；拒绝后 additionalContext 仍提示 observed |

worktree 正常案例可重跑：

```bash
go test ./tests/system/req039 \
  -run 'Test(L4AuditRealGitClosedLoop|L4AuditUntrackedResultIsPreservedAndRetryable|L4AuditDirtyAuthorityRetry|L4AuditCheckFailureRetry|WorktreeCreateUsesBoundCommitAndReusesAssignment|TaskIntegrateMergesWorktreeToVerified|IntegrationRecoversLostMergeReceiptWithoutMergingAgain)$' \
  -count=1 -v
```

其中正常 Integrator 组件链不是“真实 Agent 自主完成全流程”的证明。失败探针默认 opt-in 是为保存审查反例；默认测试绿色不能当作这些验收项已修复。

表中的 CLI 系统用例通过实际 `cli.Run` 入口在测试进程内执行，Git 操作调用真实子进程。独立 Harness 二进制的进程边界由下方 smoke 另外覆盖；这些层级不混称为 Claude 原生平台会话。

来源链用例：`go test ./tests/system/req039 -run '^TestAuditCommittedStageSourcesE2E$' -count=1 -v`。主代理独立复跑通过；来源子代理另跑 5 次均通过，并完成 `go test ./tests/system/req039 -count=1 -v`。该包仍有既有 blocker skip 和本轮 opt-in skip，不能把 package PASS 等同于每个验收项都执行。

来源测试只跑到 planning/contracts，不是完整 S7→S8→S9 的红绿测试交付链。末尾 dirty Go 测试仅证明 dirty 文件不属于权威 Git tree，未证明所有结果消费者都会拒绝把该成功结果归到 HEAD。整改 §9.3.7 的被测内容绑定仍需补充专门验收。

## 基线验证

- `go test ./...`：初始基线及新增审查测试落盘后均通过（包含缓存命中；四个未修复验收反例默认跳过）。最终 system/req039 包运行 42.478s。
- `go run ./cmd/loop-harness doctor --root .`：通过结构 schema、examples、semantic links、manual 一致性检查。健康输出另观察到 45 次 milestone stale_revision 计数，不能据此归因于本轮缺陷。
- `git diff --check`：通过。
- `go test ./internal/hook ./internal/runtime -run 'Test.*(Evidence|Fingerprint|AdditionalContext|Guidance|Lifecycle|Stop|Async|PostTool)' -count=1`：通过，补充非缓存的 Hook/evidence 针对性回归。
- 当前源码编译为 `/tmp/l4-audit-loop-harness`，复制配置到独立临时项目并 `init` 后，执行 `tools/claude-hook-smoke.sh --root <临时项目> --harness /tmp/l4-audit-loop-harness --require-platform`：exit 0。settings 边界和 SessionStart/SubagentStart/TeammateIdle/SubagentStop/PreToolUse 五个进程入口均通过；检测到本机 Claude Code `2.1.268`。该 smoke 未启动模型会话。

上述通过不抵消新增反例，也不证明真实 Claude 父子会话已经验收。

本轮共保留 4 个新增测试文件、9 个测试函数：5 个常规测试、4 个 `L4_AUDIT_REPRO=1` 验收反例。四个反例均单独实跑并失败，失败原因分别对应 F1–F4；当前不是“全部验收通过”。未自动修改生产逻辑或提交 Git。

## 尚未闭合的机制与验收项

- **原生创建接线**：`runtime worktree-create` 已证明按绑定分支创建；但 `settings.json` 未注册 WorktreeCreate/Remove，当前也没有证据证明 `Agent(isolation=worktree)` 的平台创建会调用此服务并登记 assignment/base_commit。CLI 正常路径不能替代整改 §3 的原生创建纳管要求。应明确选择经过评审的自定义创建接线，或完成原生创建后的绑定核对与台账接收，并验证不会重复创建两套 worktree。
- **S9 输入交接**：`internal/cli/repair_command.go:265–280` 仍以相对路径将 RepairContract/Session/Plan 放入 manifest/read_paths，`internal/cli/worktree.go:126–149` 创建 Git checkout 并登记坐标。当前测试未证明不入 Git 的依赖如何成为 worker 可读取、带版本的输入，也未完成“worker 红绿测试 → 报告导入 → 根目录复验 → evidence 刷新”的全链。这部分不能仅由已提交 Markdown 的阶段来源用例替代。
- **被测提交绑定**：当前新增 dirty-test 案例未到验证结果消费端。还需用“提交版本失败、dirty workspace 成功”的反例，证明成功 evidence 不能宣称提交版本通过。

## 平台证据边界

真实临时 Git 与 Harness CLI/Hook payload 的集成测试可以验证程序链路；由 Codex 拉起 Luna 子代理执行这些测试，不能等同于 Claude Code 原生 Agent 的端到端验收。

本轮核对官方文档：[worktrees](https://code.claude.com/docs/en/worktrees)、[Hooks reference](https://code.claude.com/docs/en/hooks)、[subagents](https://code.claude.com/docs/en/sub-agents)。官方明确 `CLAUDE_PROJECT_DIR` 保留启动根目录、Hook payload 的 `cwd` 跟随 worktree。真实平台仍需独立证明唯一标记被正确父/子 Agent 消费、后台交接、压缩恢复及活动 worktree 清理行为。

## 2026-09-19 · 已确认缺陷修复

本节记录用户授权后的实现修复；上面的 F1–F4 和失败输出保留为修复前证据。修复范围是四个已复现缺陷，不将平台接线、S9 全流程输入交接或被测提交绑定的待验收项自动标为完成。继续使用 `gpt-5.6-luna` / `max` 子代理，分别负责集成恢复、报告交接和离线恢复兼容；主代理负责 Writer 权威约束、交叉检查与整体回归。

- **F1**：公共 Writer 核对已绑定的 root/state/journal，当前状态和 pending 候选均检查；副本、非标准存储坐标和独立文件别名无法写入。恢复计划通过显式离线能力构建候选，不给普通 CLI 参数增加绕过入口。新增 copied Runtime、fingerprint/reconcile/pending、替代 state/journal、文件别名拒绝及项目目录别名允许的回归。
- **F2**：丢失回执时搜索绑定开发分支第一父历史，按顺序匹配记录的 target/source 父提交，并核对可达性；开发分支前进仍可恢复，其他 merge 不能冒充该回执。
- **F3**：失败 checkpoint 保存 `resume_state`。清理失败继续清理，不重跑已成功的检查；组件层重新 Inspect 返回已合并也能保留恢复阶段。worker 新提交和活动锁仍阻止清理。旧 checkpoint 没有恢复阶段时保持保守重验。
- **F4**：从登记的同仓库 worker 导入报告，使用内容哈希生成根目录不可覆盖的引用，登记与自动推进消费同一引用；保留原有根目录报告路径。拒绝越界、符号链接、身份不匹配及不同报告替换已登记引用。拒绝或登记失败的反馈不再显示 observed。

四个原始验收反例已取消 `L4_AUDIT_REPRO` 开关，直接参加默认测试。重跑命令：

```bash
go test ./tests/system/req039 -run 'Test(AuditCopiedRuntime|RuntimeAuthority|L4Audit|AuditCommittedStageSourcesE2E)' -count=1
go test ./internal/cli -run '^TestAuditHooksE2EPlanRefUsesAuthorityRoot$' -count=1
go test ./internal/runtime ./internal/recovery ./tests/system/recovery -count=1
```

Git/CLI 专项回归已由主代理独立运行通过；完整验证结果见下方最终记录。

最终验证（子代理完成全部修改后再次运行）：

- `go test ./...`：exit 0；CLI 105.535s，Runtime 25.378s，system/req039 84.287s，system/recovery 45.654s，其余包全部通过（含缓存）。
- 原始 Hook 交接反例独立非缓存运行通过（2.015s）；Git/CLI 专项独立非缓存运行通过（17.951s）。
- 新增 pending authority 与 offline pair 边界测试独立运行通过（0.252s），覆盖活动文件、交换活动文件、同文件、文件别名及根外路径拒绝。
- `go run ./cmd/loop-harness doctor --root .`：通过结构 schema、examples、semantic links 和 manual 一致性检查。
- `git diff --check`：通过。

过程中发现一处旧诊断测试使用非标准 Runtime 文件路径，将其夹具调整为 `.claude/loop-state.json` / `.claude/loop-events.jsonl` 后，继续验证原有 missing evidence 诊断。另一次测试启动恰逢并行 S3 修改调整接口，发生瞬时 vet 失败；最终全量是在接口一致后重新运行并通过。没有为通过测试放宽生产 Writer 规则，也没有修改或提交真实项目分支。
