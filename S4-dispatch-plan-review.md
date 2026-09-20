# S4 dispatch plan 落地审查

审查日期：2026-09-19。范围：当前未提交工作区的 waves-v1 实现、正式机制、模板、实际 CLI/Hook/Writer 接线。初次审查只新增测试和报告；用户随后授权修复，修复记录见文末。未提交或回退已有修改。

## 优化与结论

本次优化把 S4 的 TASK 集合提升为可审签的整体派发计划：TASK 持有真实依赖和写域，REQ index 持有成员、波次和资源顺序；TR-002 独立登记 dispatch_plan，S5 精确审签，TR-003 冻结。S6 基于当前 Runtime、Result、owner、assignment 和集成 checkpoint 持续计算候选与容量，不引入整波屏障。workgroup 注册在原 Writer 事务中再次检查实际输入。

初次审查时主链路已落地，但存在以下四项缺口。下文发现及失败结果保留为历史证据；后续修复状态和复验记录见文末。

## 审查发现

### F1 / P1：output_paths 可以扩大已审 TASK 的写权限

位置：`internal/assignment/register.go:146-150`、`internal/assignment/activation_envelope.go:94-101`。

注册时只把 manifest 的 `write_paths` 交给 `ValidateWorkPackage` 检查，但 activation 的 `allowed_write_paths` 取 `write_paths ∪ output_paths`。实际 CLI 反例：TASK 只允许 `web/pages`，manifest 设置 `write_paths=[web/pages]`、`output_paths=[server/api]`，注册退出 0、Runtime revision 从 1 到 2，并生成允许两个目录的 activation。对照案例把 `server/api` 放入 write_paths 时会正常拒绝。

影响：派发入口承诺的写域边界与实际下发权限不一致；也可能使依据 TASK 写域计算的冲突集合不完整。此反例证明授权扩大，不声称已完成越界产品写入或绕过最终集成检查。

建议：统一“有效可写范围”的定义，注册校验和 activation 使用同一集合；如需额外证据输出权限，应显式限定证据路径，不能任意扩大产品目录。

### F2 / P2：资源约束中的非法数据行被静默丢弃

位置：`internal/semantic/dispatch_plan.go:277-280,293-295`。

两个独立反例均得到空 Problems：

- Resource order 的完整四列数据行 `| malformed-before | TASK-042-01 | resource:test-db | missing task id |` 被当作应跳过内容，未报告非法前置 TASK。
- 两个 TASK 均声明 `| shared-test-db | mutable database |`；由于缺少规范要求的 `resource:` 前缀，两者资源声明直接消失，计划仍通过。

影响：编写者已填写的互斥/顺序意图可因格式错误无声失效。应明确识别表头和分隔行，对其余非空数据行检查列数、ID、资源名及必填字段。模板的三级 Resources 标题和样例的二级标题均能正确解析，不属于缺陷。

### F3 / P2：示例公开的相对 root 命令误报合同缺失

位置：`internal/semantic/tasks.go:59-89`、`internal/fileview/contracts.go:18-28`。

`docs/examples/dispatch-plan/README.md` 给出的相对路径入口实际退出 1：6 个 TASK 均报告 BE-042 文件不存在，`clauses_total=0`。同一目录改用绝对路径则退出 0，6 个 TASK、1 个条款且覆盖完整。原因是 TasksCheckWithFiles 保留相对 root 构造路径，Disk reader 又拼接一次 Root；内部计划/TASK 读取的规范化未覆盖合同读取。

这属于 S4 复用入口暴露的路径问题，不断言其完全由 S4 新引入。应统一根目录规范化和 reader 路径约定，保留相对/绝对输入等价回归。

### F4 / P2：Result 内容更新但 created_at 不变时，仍可复用旧集成验证

位置：`internal/qualitygate/dispatch_progress.go:114-119`。`internal/controller/cycle.go:140-155` 使用 Writer 刷新允许变动的证据指纹，completion_report 位于正式 mutable evidence 清单中。

当前判定检查最新 Result 的 SHA，却仅以 `verified_at >= created_at` 关联集成 checkpoint，没有把 checkpoint 绑定到被验证 Result 的内容版本。反例改变 Result 的 changed_paths 并保留 created_at，通过真实 `Store.RefreshEvidenceFingerprints` 持久化新 SHA 后，`PlannedBuilderProgress` 仍返回 integrated；直接 Board 反例同样失败。因此实施清单所说“更新后的 Result 不能借旧验证放行”目前只覆盖时间戳前移场景。

建议：集成验证记录 Result 的 evidence ID/摘要或等价不可变版本，投影要求当前版本与记录一致；内容变更后重新验证。不能把可保留的创建时间当作内容版本。

证据等级：这是生产 Writer 刷新与生产进度求值的组件集成反例，使用构造的 Runtime、checkpoint 和简化 validator；另由静态代码确认 Controller 接线。未跑完整生产 Controller→实际 Git 集成→Result 改写链，不把它称作平台 E2E。

## 验证分层

由三个已有 `gpt-5.6-luna / max` 子代理分别审查静态计划、Hook 主链和运行态投影，主审独立执行实际注册反例及复跑关键测试。

| 层级 | 实际覆盖 | 结果 |
| --- | --- | --- |
| 真实 Hook/CLI + 临时 Git 项目 | S3→S4→S5→S6；TR-002 登记独立计划；精确 subjects；TASK 分母仍为 6；dirty 副本不替换已提交输入；S6 冻结写保护 | 通过 |
| 真实 CLI/Writer 注册 | 正常派发；前置未满足、write_paths 越界、脏 TASK、重复 owner 拒绝；拒绝后 Runtime 字节不变 | 通过 |
| 真实 CLI/Writer 授权反例 | output_paths 扩大 activation 权限 | 验收失败，F1 |
| S6 CLI | JSON、容量声明、只读 Runtime、legacy 恢复可读 | 通过 |
| 运行态投影组件 | 容量 2、提前释放、未集成阻塞、失效证据、旧 assignment、Result 时间更新、checkpoint 恢复、写域冲突 | 通过 |
| 静态计划组件 | 正式样例、REQ 成员闭包、跨 REQ、联合 DAG/资源环、依赖、路径、元信息、标题层级 | 通过 |
| 非法资源与路径入口反例 | 两类非法资源行；相对/绝对 root 等价性 | 验收失败，F2/F3 |
| Result 内容版本反例 | 相同 created_at、新 SHA、旧 checkpoint；Board 与 Writer 刷新后求值 | 验收失败，F4（组件集成） |

Hook 流程使用生成的审查证据夹具；注册测试预置已冻结 S6 状态；投影测试构造 Runtime/Result/checkpoint 事实。这些验证覆盖本地 Harness 接线，不代表真实 Claude Code 多 Agent 平台并行执行、目标产品联调或效率收益已实测。

## 可复现命令

正常验证：

本轮全量 `go test -p 2 ./...` 退出 0（含缓存命中）；新增主链/注册/静态计划/投影/CLI 审查测试另以 `-count=1` 执行通过，全量之后补充的 Writer 刷新测试亦单独复跑通过。doctor 退出 0，结构、schema、示例、语义链接和手册检查通过；Runtime 健康指标另报。`git diff --check` 通过。

```bash
go test ./tests/system/req039 -run '^(TestAuditDispatchPlanHookSpineCommittedSubjectsAndFreeze|TestS4AuditRegistrationChecksActualInputs)$' -count=1 -v
go test ./internal/semantic -run '^TestAuditDispatchPlan' -count=1
go test ./internal/dispatch ./internal/cli -run '^TestAuditDispatch' -count=1 -v
go test -p 2 ./...
```

初次审查的失败验收使用以下 opt-in 命令。修复后已取消这些开关，原测试现在全部默认运行；保留命令仍可复跑：

```bash
S4_AUDIT_REPRO=1 go test ./tests/system/req039 -run '^TestS4AuditOutputPathsCannotExpandReviewedScope$' -count=1 -v
L4_S4_AUDIT_REPRO=1 go test ./internal/semantic -run 'TestAuditDispatchPlan(MalformedResourceOrderProbe|MalformedTaskResourceProbe|RelativeRootProbe)$' -count=1 -v
S4_DISPATCH_AUDIT_REPRO=1 go test ./internal/dispatch -run '^TestAuditDispatch(RuntimeWriter)?MutableResultInvalidatesOldCheckpoint$' -count=1 -v
```

独立二进制入口：

```bash
go build -o /tmp/s4-audit-loop-harness ./cmd/loop-harness
/tmp/s4-audit-loop-harness tasks check --root docs/examples/dispatch-plan/project --req REQ-042 --json
/tmp/s4-audit-loop-harness tasks check --root "$PWD/docs/examples/dispatch-plan/project" --req REQ-042 --json
/tmp/s4-audit-loop-harness doctor --root .
```

本地证据：`/tmp/s4-audit-spine-registration.log`、`/tmp/s4-audit-output-scope.log`、`/tmp/s4-audit-semantic-probes.log`、`/tmp/s4-audit-runtime-cli.log`、`/tmp/s4-audit-example-relative.json`、`/tmp/s4-audit-example-absolute.json`、`/tmp/s4-audit-full.log`、`/tmp/s4-audit-doctor.log`。临时日志不是项目长期证据；新增测试保留完整复现方法。

F4 补充日志：`/tmp/s4-audit-mutable-result.log`。


## 修复记录（用户授权后，2026-09-19）

- **F1**：activation 和计划注册共享 `effectiveActivationWritePaths`，注册在 Writer 事务中校验 write_paths/output_paths 的完整并集。两个路径可分别位于 TASK 批准目录的不同子目录；未批准的产品或证据路径不能借 output_paths 授权。无计划的 legacy 注册保持原行为。
- **F2**：只跳过空行、正式表头及分隔行；其余资源数据行检查列数、非空资源名和 TASK 端点，错误会阻止计划通过。保留合法资源及同波次冲突正例。
- **F3**：TasksCheckWithFiles 统一规范化绝对 root，合同、TASK、计划使用一致路径。实际二进制的相对和绝对 root 命令均退出 0，6 TASK、1 条款且覆盖完整。
- **F4**：Inspect 记录 Result 的仓库相对路径和 SHA256；Integrate 在检查前后重新核对内容，在验证成功时写 checkpoint；planned quality gate 要求当前 Result 与 checkpoint 的路径/摘要一致。报告改变或旧记录缺少绑定时重新运行检查，复用已记录 merge；复验前检查目标分支仍包含该 merge。单纯 ack/cleanup 不刷新摘要或 verified_at。legacy 无绑定的 checkpoint 不自动成为新计划完成事实。

原失败反例均已转为默认回归。新增 `TestS4ResultBindingThroughTaskIntegrate` 使用真实临时 Git worktree 和 CLI：首次集成→检查→complete→清理；同 created_at 修改报告→检查命令计数增加→摘要和验证时间更新→merge HEAD 不变；删除报告→拒绝且原 checkpoint 不变。这补齐了初审 F4 仅有 Writer 组件反例的验证层次；完整新计划阶段门仍由主链测试与进度求值测试分别覆盖，不宣称真实平台多 Agent 执行。

已独立通过：四个核心包（assignment、semantic、dispatch、qualitygate）`-count=1`；实际注册/Hook 主链/Result 集成 E2E（`/tmp/s4-fix-e2e.log`）；额外的 sibling output 与未批准证据路径回归（`/tmp/s4-fix-registration-final.log`）；5 个 integration 绑定/恢复/验证期间改写/缺失报告/目标历史失效测试（`/tmp/s4-fix-integration-final.log`）；示例相对/绝对路径；doctor（`/tmp/s4-fix-final-doctor.log`）；`git diff --check`。完整回归首轮 `go test -p 2 ./...` 退出 0（`/tmp/s4-fix-full.log`），最终稳定版本的全量复跑同样退出 0（`/tmp/s4-fix-final-full.log`）。四项问题均已修复并通过复验，变更未提交。
