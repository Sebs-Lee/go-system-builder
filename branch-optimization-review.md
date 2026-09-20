# 分支优化综合复审记录

日期：2026-09-20。范围：L4 worktree/Hook、S3 shared-model contracts、S4 dispatch plan 的全部工作区变更及前轮修复。

后续完整回审、Luna / max 子代理 E2E 及统一修复的当前状态见 [落地回审记录](branch-optimization-recheck.md)。下文保留前轮调查与验收历史，其中待验证边界以后续记录为准。

原始提交保存当时实现与复审测试，不表示当时以下问题已经修复。2026-09-20 主清单五项已修复，隔离全量及相关包 race 回归通过；后续修复记录见文末；原问题描述保留为调查依据。已完成前轮专项修复的详情见三个专项 review 文档。综合复审的核心结论是：单一 Runtime 权威、Git 中的正式输入、共享模型契约与依赖调度的方向一致，但维护入口和跨模块交接仍存在断点。

## 原始待修复问题（后续状态见文末）

1. **P1：维护命令可重新认可未经复审的冻结内容。** 在已通过评审的任务修改并提交后，registration 原本拒绝漂移；执行 `runtime fingerprint` 或 `runtime reconcile-policy-ref` 后，同一未经 S5 复审的范围可注册成功。两条路径调用广义 `RefreshFingerprints`，更新了冻结任务的摘要。应限制维护刷新范围，并让正式产品变化走评审失效流程。复现：`BRANCH_AUDIT_REPRO=1 go test ./tests/system/req039 -run '^TestBranchAuditMaintenancePreservesReviewedSubjects$' -count=1 -v`。

2. **P1：正常完成上报的 Result 来源分叉。** `task-complete` 为 task 写入规范化 Result 引用，但 assignment 的 CompletionRef 仍指向原始消息；Integrator 读取 assignment，planned quality gate 读取 task。两者的路径/摘要绑定不能自然闭合。已复现实际完成上报后的引用差异，探针继续覆盖集成 checkpoint；完整后继任务释放仍需在修复后验证。复现入口：`L4_AUDIT_REPRO=1 go test ./tests/system/req039 -run '^TestL4AuditAssignmentAndTaskCompletionRefsShareCanonicalSource$' -count=1 -v`。

3. **P2：有效写入范围在集成端不一致（静态确认，待完整复现）。** 注册与 activation 使用 `write_paths ∪ output_paths`，hookctx assignment 投影及 Integrator 越界检查只保留 write_paths。合法的同一 TASK 范围内输出若位于更窄 write_paths 之外，可能被集成拒绝。应共用有效范围计算，避免新增权限规则。

4. **P2：S5 阶段的手动状态命令提前给出可调度提示。** `s6 status` 在 document_verification 状态仍输出 Ready/Next batch，尽管实际 registration 会拒绝，Hook 自动提示也有 building 状态保护。应明确预览状态或在批准前隐藏调度建议。复现：`BRANCH_GUIDANCE_AUDIT_REPRO=1 go test ./internal/cli -run '^TestAuditS6StatusDoesNotAdvertiseBeforeDocumentPass$' -count=1 -v`。

5. **P2：8 个模板仍引用旧任务总表路径。** 合约、需求、project-map、验收与 release audit 模板仍使用 `docs/tasks/index.md`，而当前机制要求 REQ 专属路径。应统一为 `docs/tasks/index-REQ-<id>.md`，减少 agent 在跨阶段交接时自行猜测。复现：`BRANCH_GUIDANCE_AUDIT_REPRO=1 go test ./internal/cli -run '^TestAuditGuidanceUsesREQScopedPlanPath$' -count=1 -v`。

## 原始复审的验证边界与简化建议

前轮完整 Go 测试及 doctor 通过；本轮复审的已知问题通过环境变量开启探针，默认测试跳过这些探针。因此默认测试通过不能作为上述缺陷已解决的证明。

S9 修复输入从 authority 到 worktree 的实际交付、原生平台 Worktree hook wiring、capture 执行内容与证据摘要的绑定，以及 dispatch 与 controller 的文件来源配置一致性，仍需补齐验证。新增 integration 审计文件中的 checkpoint 投影和 S9 输入探针用于后续调查，不能直接等同于已证明阶段门失效。

后续修复应优先统一 Result 来源、有效写入范围和文件来源解析；维护命令不应成为额外的评审批准入口。保留现有生命周期和单一权威，让 agent 只需遵循一套阶段提示与输入路径，避免为这些断点叠加新的状态机或例外规则。


## 2026-09-20 · 五项问题调查与修复

### 调查证据

在 `fff2b1b` 的独立临时副本中运行原始探针，确认问题 1、2、4、5。测试 Git 命令使用临时身份，避免依赖开发者全局配置；没有更改真实仓库身份或分支。

问题 3 增补真实 Git/CLI 反例：注册 TASK 的写域为 `packages/validation/source`，输出为同一已批准 TASK 内的 `packages/validation/generated/result.json`。原提交能注册并完成上报，却在集成时报告 `worktree diff is outside the assignment write scope`。这把原来的静态怀疑补成了可执行反例。

完整链还发现 `task-complete --root <project>` 将默认 state/journal 当作当前 cwd 的路径，导致项目根与命令 cwd 不同时读错状态。原提交的范围反例显式传入 state/journal 以继续调查；修复后的回归保留默认路径，验证实际 CLI 使用方式。

### 修复结果

1. **维护刷新**：公共 `RefreshFingerprints` 仅更新 Harness definition/policy 元数据。REQ、登记文档（含设计、合同、计划与 TASK）、TASK entity 和 evidence 保留已登记摘要；`fingerprint` 报告 `drifted` 并给出既有变更/复审或证据登记入口。两条维护命令都不能重新认可未经审查的 TASK 范围。集成使用的显式 mutable-evidence 刷新入口保持独立。
2. **Result 来源**：`task-complete` 在同一事务中为 Agent 与 TASK 写入相同的规范 Result；从 reported/done 重报时也更新引用。Hook assignment 投影优先使用所属 TASK 的规范引用，兼容旧 Runtime 的原始消息指针，并隔离其他 owner。Integrator 和计划进度门因此消费同一路径与摘要。
3. **有效写域**：抽取 `pathscope.EffectiveWrites`，注册/activation 与 Hook assignment 投影共用 write/output 并集；Integrator 消费该投影，保留 legacy scope 回退。并集仍受注册端已评审 TASK 范围约束，不扩大批准权限。
4. **阶段提示**：非 building 状态下，`s6 status` 文本/JSON 只说明当前生命周期和派发不可用，不提供 Ready/Next batch。building 的现有派发投影保持可用。
5. **模板路径**：八个模板统一使用 `docs/tasks/index-REQ-{id}.md` 或对应相对链接，与现有 index 模板占位符一致。
6. **交付链附带修复**：`task-complete` 的 state/journal 默认路径与显式相对路径统一以 `--root` 解析。

原主清单探针已转为默认测试；其余调查探针仍保持原有边界。新增 `TestBranchDeliveryReleasesSuccessorWithSiblingOutput` 覆盖真实注册、worktree 创建、输出提交、规范完成上报、集成/清理与后继注册：仅有报告时拒绝后继；两个前置验证集成后释放后继，不等待无关的第三项工作。清理后重报会再次阻止后继，重新验证后恢复资格且不重复 merge。Agent working 状态由夹具设置，不宣称验证了平台 spawn/激活。

### 验证记录

- 原提交反例：`/tmp/branch-audit-baseline.log`、`/tmp/branch-delivery-baseline.log`、`/tmp/branch-delivery-baseline-scope.log`。最初缺少 Git 身份和依赖下载失败属于环境/夹具问题，已排除后复现，不计为产品缺陷。
- 新交付链定向回归通过：`/tmp/branch-delivery-fixed.log`、`/tmp/branch-delivery-resubmit.log`。
- 构建、`go vet ./...`、`make fmt-check` 和 `git diff --check` 通过。
- 当前源目录缺少 `.claude/loop-state.json`，直接 doctor/validate 失败；在复制当前模板的隔离副本中执行 init 后，doctor 与 validate --all 均通过。没有初始化或替换真实仓库 Runtime。
- 初始化当前文件副本后，`go test -p 2 ./...` 全部通过（exit 0），包含最终交付/重报回归；REQ-039 系统包 165.622s。日志 `/tmp/branch-fix-full-isolated.log`。
- `go test -race -p 2 ./internal/runtime ./internal/assignment ./internal/hookctx ./internal/integration ./internal/dispatch ./internal/cli ./tests/system/req039` 全部通过（exit 0）；CLI 384.546s、REQ-039 392.304s，无数据竞争报告。日志 `/tmp/branch-fix-race-verified.log`。
- 源目录直接全量测试还受既有夹具影响：semantic 的五项测试要求本地 Runtime；schema 的摘要长度断言依赖绝对路径长度，并在原提交上独立复现失败（1727 / 3453 bytes），当前源路径为 1727 / 3455 bytes，整数除法边界导致失败。本轮没有放宽该断言；初始化的隔离副本用于完整验证，不能据此宣称源目录直接运行全绿。
- 与本轮相关的测试修整：维护测试使用格式合法的旧 SHA；新 loader 夹具补齐 revision；CLI 的根路径回归移到临时项目，不再要求或覆盖开发者 Runtime。中间失败已保留在日志，旧运行在夹具更新后被新运行替代，不计为最终通过。

### 保留的验收边界

本轮关闭范围是主清单五项及直接阻塞交付链的 CLI 路径问题。另行复核代码后，S9 非 Git 输入仍由 authority 生成并以相对 read_paths 声明，原生 WorktreeCreate/Remove 仍未在 settings 注册，capture 尚缺“dirty 成功不能归到干净提交”的完整消费端验收，dispatch 的来源规则仍硬编码而 controller 消费 Definition.FileSources。这些既有边界不因本轮五项回归通过而变为已验收。checkpoint 的 milestone 字符串/对象投影探针也不等同于依赖释放门失效；本轮实际依赖释放链读取并验证持久 checkpoint。

验证进程使用临时 GIT_AUTHOR/GIT_COMMITTER 身份；隔离副本的 18 个变更 Go 文件已逐一与当前工作区比对一致。改动保留在工作区，未提交或发布。

临时日志仅用于本轮核查，长期复现入口是仓库中的默认回归测试。
