# 分支优化综合复审记录

日期：2026-09-20。范围：L4 worktree/Hook、S3 shared-model contracts、S4 dispatch plan 的全部工作区变更及前轮修复。

本次提交保存当前实现与复审测试，不表示以下问题已经修复。已完成前轮专项修复的详情见三个专项 review 文档。综合复审的核心结论是：单一 Runtime 权威、Git 中的正式输入、共享模型契约与依赖调度的方向一致，但维护入口和跨模块交接仍存在断点。

## 待修复问题

1. **P1：维护命令可重新认可未经复审的冻结内容。** 在已通过评审的任务修改并提交后，registration 原本拒绝漂移；执行 `runtime fingerprint` 或 `runtime reconcile-policy-ref` 后，同一未经 S5 复审的范围可注册成功。两条路径调用广义 `RefreshFingerprints`，更新了冻结任务的摘要。应限制维护刷新范围，并让正式产品变化走评审失效流程。复现：`BRANCH_AUDIT_REPRO=1 go test ./tests/system/req039 -run '^TestBranchAuditMaintenancePreservesReviewedSubjects$' -count=1 -v`。

2. **P1：正常完成上报的 Result 来源分叉。** `task-complete` 为 task 写入规范化 Result 引用，但 assignment 的 CompletionRef 仍指向原始消息；Integrator 读取 assignment，planned quality gate 读取 task。两者的路径/摘要绑定不能自然闭合。已复现实际完成上报后的引用差异，探针继续覆盖集成 checkpoint；完整后继任务释放仍需在修复后验证。复现入口：`L4_AUDIT_REPRO=1 go test ./tests/system/req039 -run '^TestL4AuditAssignmentAndTaskCompletionRefsShareCanonicalSource$' -count=1 -v`。

3. **P2：有效写入范围在集成端不一致（静态确认，待完整复现）。** 注册与 activation 使用 `write_paths ∪ output_paths`，hookctx assignment 投影及 Integrator 越界检查只保留 write_paths。合法的同一 TASK 范围内输出若位于更窄 write_paths 之外，可能被集成拒绝。应共用有效范围计算，避免新增权限规则。

4. **P2：S5 阶段的手动状态命令提前给出可调度提示。** `s6 status` 在 document_verification 状态仍输出 Ready/Next batch，尽管实际 registration 会拒绝，Hook 自动提示也有 building 状态保护。应明确预览状态或在批准前隐藏调度建议。复现：`BRANCH_GUIDANCE_AUDIT_REPRO=1 go test ./internal/cli -run '^TestAuditS6StatusDoesNotAdvertiseBeforeDocumentPass$' -count=1 -v`。

5. **P2：8 个模板仍引用旧任务总表路径。** 合约、需求、project-map、验收与 release audit 模板仍使用 `docs/tasks/index.md`，而当前机制要求 REQ 专属路径。应统一为 `docs/tasks/index-REQ-<id>.md`，减少 agent 在跨阶段交接时自行猜测。复现：`BRANCH_GUIDANCE_AUDIT_REPRO=1 go test ./internal/cli -run '^TestAuditGuidanceUsesREQScopedPlanPath$' -count=1 -v`。

## 验证边界与简化建议

前轮完整 Go 测试及 doctor 通过；本轮复审的已知问题通过环境变量开启探针，默认测试跳过这些探针。因此默认测试通过不能作为上述缺陷已解决的证明。

S9 修复输入从 authority 到 worktree 的实际交付、原生平台 Worktree hook wiring、capture 执行内容与证据摘要的绑定，以及 dispatch 与 controller 的文件来源配置一致性，仍需补齐验证。新增 integration 审计文件中的 checkpoint 投影和 S9 输入探针用于后续调查，不能直接等同于已证明阶段门失效。

后续修复应优先统一 Result 来源、有效写入范围和文件来源解析；维护命令不应成为额外的评审批准入口。保留现有生命周期和单一权威，让 agent 只需遵循一套阶段提示与输入路径，避免为这些断点叠加新的状态机或例外规则。
