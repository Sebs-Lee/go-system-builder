# S3 共享模型优化落地审查

> F1/F2/F3 已修复，最终全量回归及 doctor 通过，详情见文末。以下审查正文保留修复前的发现与复现证据；原失败探针现已转为默认回归，不再需要环境变量。

日期：2026-09-19。对象：当前工作区，包含前序未提交实现。本轮只增加审查记录与测试，不修改生产实现，也不提交 Git。

结论：优化主干已经落地，但未形成完整闭环。确认三项实现缺陷：冻结模型可经 worker 集成改写、跨模型稳定 `$id` 冲突漏检、同一 REQ 跨索引 operation/slot 冲突漏检。常规测试通过不能替代这些失败验收用例。

## 优化的实际含义

本次把前后端各自推测数据、最后用 SYNC 对账，改为共享模型与 SYNC 按业务切片先共同收敛，再形成分端责任。共享模型拥有结构，SYNC 拥有交互协议，FE/BE 拥有实现职责；TASK 使用有序、带目的的链接进入这套共同依据。S3 登记 Schema、样例及实际加载的传递依赖，S5 将其纳入审查 subject 和冻结基线。

审查依据是 `S3-shared-model-contracts-optimization.md` 第 17 节、`blueprint/L4-shared-model-contract-governance.md` 和安装规则。第 1–16 节中的未来能力不全部视为本版承诺。JSON Schema 为首版适配器；没有要求补做 OpenAPI/Proto、多语言生成平台或目标产品浏览器。

继续使用三个 `gpt-5.6-luna` / `max` 子代理分别审查 Git/冻结、Schema/引用和 TASK/HTTP 消费。主代理负责跨模块接线、反例独立复跑与整体验证。

## F1 · P1：S6 集成入口未消费冻结模型清单，worker 可以改写已锁定模型

位置：`internal/cli/controller.go:1431–1438`；`internal/integration/inspect.go:322–347`。

Runtime 中的 shared-model design document 能被 Hook loader 正确投影为 LockedArtifact，直接在权威根目录 Write 同一路径也会被拒绝。但 `task-integrate` 给 Inspect 的只有 Assignment 的 RequiredChecks，没有把 Runtime 的锁定路径传入；检查器只识别 `locked:` 提示，不自行读取 Runtime。

真实 Git + CLI 反例：S6 已登记 `shared-model:internal/orders.schema.json`，状态 locked、generation=1；worker 将 `order_id` 从 string 改成 number 并提交。`runtime task-integrate` 返回 exit 0、`state=integrated`、checkpoint `complete`，开发分支前进，根模型实际变为 number。这是对冻结后修改的确定性检查缺失，不要求工具判断任意语义兼容性。该缺口位于复用的 Integrator 接线，S3 新机制依赖这条路径，因此仅证明 S5 冻结成功还不能证明交付闭环。

测试在 `tests/system/req039/audit_sharedmodel_integration_test.go`。反例从有效的锁定 documents 投影启动，不宣称该例单独执行了整个规划流程。配对正向控制通过同一 fixture 的 PreToolUse 确认根目录写入被拒绝；fixture 的 milestone 已与 S6 lifecycle 对齐，排除了旧阶段投影造成的误报。

```sh
S3_AUDIT_REPRO=1 go test ./tests/system/req039 -run '^TestS3AuditFrozenModel' -count=1 -v
```

实跑结果：直接根目录写保护通过；worker 合并验收断言失败。建议在合并前使用权威 Runtime 的冻结集合检查源差异，并在集成事务边界核对，不能只依赖任务自己声明的检查命令。

## F2 · P2：不同 Schema 可以声明相同稳定 `$id`，重复权威未被发现

位置：`internal/sharedmodel/check.go:185–190`。

每个 baseline 行单独创建 jsonschema Compiler；跨行没有核对规范化后的稳定标识与定义来源。反例在首行实际选中的 request subschema 和另一文件的 response Schema 上放置相同 `$id`，两者字段约束不同，消费者均正确引用，正反样例也满足各自约束。检查仍返回空 Problems。

这不是缺少消费者引用，而是 A03 的同一稳定标识对应多个作者源没有闭合。无需建设业务 registry，但应在所检查的模型资源闭包内识别冲突标识，诊断两处来源，并允许同一源被多个消费者正常复用。

```sh
L4_S3_AUDIT_REPRO=1 go test ./internal/sharedmodel -run '^TestAuditSharedModelDuplicateStableIDProbe$' -count=1 -v
```

子代理与主代理均独立运行，验收断言失败：`expected duplicate stable-$id rejection, got []string(nil)`。

## F3 · P2：同一 REQ 跨索引的 operation/slot 冲突漏检

位置：`internal/sharedmodel/check.go:151`。

operation/slot 的去重表在每个 CONTRACTS 文件内部重新初始化。同一次 REQ 范围检查中，两份索引均显式声明 `REQ-001`，均定义 `cancelOrder/request`，一份要求 `order_id:string`，另一份要求 `order_id:integer`。两者拥有各自正确的 FE/BE/SYNC 引用和正反样例，单独检查均通过；合并后仍返回空 Problems 和 Warnings。

该反例不是同源模型的合法复用，而是同一批次、同一交互槽位指向互相冲突的结构定义。建议将权威冲突检查提升到 REQ 批次作用域，并输出双方索引及 Schema 来源。

```sh
L4_SHARED_MODEL_AUDIT_REPRO=1 go test ./internal/cli -run '^TestAuditSharedModelReadingDuplicateAuthorityAcrossIndexes$' -count=1 -v
```

子代理与主代理均独立运行，验收断言失败：预期 duplicate operation/slot 诊断，实际 `problems=[] warnings=[]`。

## 文档残留与验收边界

- `skills/specification-planning/SKILL.md:161` 仍称类型/schema/迁移是地基、下游必须依赖；同文件末尾又允许独立生成并行，仅真实共享实现需依赖。`blueprint/L3-S3-contracts.md:88–90` 的 T2/T3/T4 表仍呈现 FE→BE→SYNC 的旧分解，而正文已改共同模型/SYNC先行。建议直接改执行入口和分解表，避免靠末尾附加说明消解旧路径。这是文案一致性风险，本轮没有声称通过真实 Planner 会话复现了错误拆分。
- Consumer 检查验证 Shared model inputs 中出现相同 Schema 引用，没有独立的 operation/slot→Schema 映射语法。正式蓝图也未定义这种消费者映射，因此没有把“交换两个 slot 的文字标签”另报成实现缺陷。
- 真实 HTTP 是工厂的合成消费者/提供者；200、400、409 和无副作用校验不等于目标产品 FE/BE、UI 恢复、权限及并发完成的全链验收。
- 本轮验证了链接可解析，没有证明新 Agent 只拿 TASK 就能独立完成真实任务；导航回链测试也不替代原有执行 DAG 环检查。
- 可变 evidence 自动刷新与模型冻结职责仍分开。本次没有把显式 `runtime fingerprint` 命令等同于自动 evidence 刷新，也不借此作未经复现的扩大结论。

## 已验证的正常路径

| 场景 | 层级 | 结果 |
| --- | --- | --- |
| 示例合同通过，非法正样例返回非零 JSON 诊断 | 独立编译的 Harness 二进制进程 | 通过：合法 exit 0，order_id:number exit 1 |
| 本地 anchor、递归 $ref、REQ 范围隔离、远程依赖拒绝 | 解析/检查器 | 通过 |
| Git symlink 与内部 worktree 的磁盘别名拒绝 | 真实 Git + 文件视图/检查器 | 通过 |
| dirty 磁盘不污染固定 Git view，branch 移动后 Verify 拒绝 | 真实 Git + 检查器 | 通过 |
| 已提交闭包登记；根工作区未跟踪 schema 不能补齐 Git 输入 | 真实 Git + transition registry action | 通过 |
| 传递依赖漂移、依赖增加和删除在 S5 被拒绝；S3 重登记后可冻结 | 真实 Git + transition registry action | 通过 |
| 不同 checkout commit、共同闭包相同 | 真实 linked worktree + transition action | 通过 |
| 必读坏锚点、重复 path#fragment、非法 Mode | Reading / TasksCheck / CLI tasks check | 通过，诊断传入自然 S4 检查 |
| 同文件不同 fragment、导航回链、optional 历史坏链 | Reading | 不误拒 |
| 正常取消、结构拒绝、业务拒绝无副作用、Mock 枚举偏移 | 本地 httptest HTTP + Schema 边界校验 | 通过；属于 synthetic integration |

正式 action 集成测试直接调用 transition registry，不能将其写成已经穿过完整的 Runtime Writer、Gate、Hook 全阶段 E2E。另行补充的自然 Hook 链结果和最终全量结果记录在文末。

## 可重复执行的审查资产

本轮新增五个测试文件，生产实现保持原状：

- `internal/sharedmodel/audit_sharedmodel_sources_test.go`：源视图、引用、REQ 范围及 F2。
- `internal/transition/audit_shared_models_e2e_test.go`：真实 Git 闭包登记、冻结漂移与同源不同 commit。
- `internal/cli/audit_sharedmodel_reading_e2e_test.go`：阅读路径、自然 S4 检查、HTTP 边界消费及 F3。
- `tests/system/req039/audit_sharedmodel_integration_test.go`：根目录锁保护正向控制及 F1 的 worker 合并反例。
- `tests/system/req039/audit_sharedmodel_spine_test.go`：真实 Hook/CLI 的 S3→S4→S5→S6 推进。

三个已确认的失败验收探针通过上文环境变量显式启用，默认跳过，以免审查提交本身破坏既有默认回归；默认绿色不代表三项缺陷修复。后续修复时应将相应用例改成默认必跑，并确认旧实现失败、新实现通过。

额外执行了独立二进制的 `contracts check --json` 正反例，测试目录为 `/tmp/s3-audit-cli-uyzuajv1`；非法正样例准确返回 exit 1。验证日志保留于 `/tmp/s3-audit-*.log`，这些临时文件不是仓库交付依赖。

## 自然 Hook 链验证

`go test ./tests/system/req039 -run '^TestAuditSharedModelHookSpineCommittedSource$' -count=1 -v` 最终通过（主代理独立复跑 4.166s）。测试使用隔离 Git 仓库与真实 Hook/CLI 路径，执行以下阶段推进，没有手动 transition：

1. 根磁盘存在但未提交的 Schema 不能让 S3 通过，Runtime 不登记模型。
2. 提交模型后，PreToolUse 提交 PTR-PLAN-02，进入 S4；Schema 与三个样例登记为带内容哈希的 locked design documents。
3. 补齐 TASK 条款与规划证据后，Hook 提交 TR-002，进入 S5。
4. S5 审查 subject 包含模型，Hook 提交 TR-003，进入 S6；模型仍为 locked 且哈希存在。

这里的任务与审查证据由测试夹具准备，验证的是 Harness 对这些正式输入的消费，不是实际调用 Planner/Reviewer 模型生成证据。用例中的 `worker-only.json` 只是文件名，该否定分支实际测试根目录未跟踪文件；真实 linked worker 的集成路径另由 F1 用例覆盖。

首次组装此用例曾因 Go 声明语法及缺少已复制合同的 TASK 条款覆盖而失败，修正的仅是测试夹具；最终独立复跑通过，没有将这些中间失败计为产品缺陷。

## 最终回归结果

- `go test -p 2 ./...`：最终 exit 0，覆盖最终 Hook 用例；`tests/system/req039` 45.934s，日志 `/tmp/s3-audit-final-full.log`。
- `go test ./internal/sharedmodel ./internal/transition ./internal/cli -run 'Test(AuditShared|ConsumerProviderPilot)' -count=1`：全部通过；日志 `/tmp/s3-audit-final-targeted.log`。
- `doctor --root .`：通过结构、示例、语义链接及手册同步检查；历史运行时健康指标单独呈现，不宣称被清零。
- `git diff --check`：通过。
- F1/F2/F3：分别显式启用失败探针，均由主代理确认复现；不是已修复项。

建议修复顺序为 F1 冻结集合接入集成边界，再补 F2/F3 的批次级冲突检查，最后消除规划文案的旧路径。本轮没有改生产代码，没有自动 commit，也没有目标业务项目的真实前后端浏览器试点。

## 后续修复（用户授权后，2026-09-19）

上述三项缺陷均已修改，审查正文保留为修复前证据。本节描述修复后的工作区；没有提交 Git。

- **F1**：Controller 将 Runtime 冻结文档路径通过独立 `InspectRequest.LockedPaths` 传给 Integrator，不再依赖任务自行声明 `locked:` 检查。Inspection 保留锁集合，合并锁内再检查实际源差异。Git 差异关闭重命名折叠并使用 NUL 分隔，保证删除/重命名不会隐藏被冻结的旧路径。
- **F2**：在同一 REQ 批次中累积编译器实际加载的 Schema 资源标识，解析嵌套资源与相对 ID，拒绝不同源声明同一标识；源文件和标识排序保证诊断稳定。同源复用合法，样例文件及 Schema 的 default/const/examples 数据不会变成资源。资源位置、draft4 的 id、旧 draft 的 `$ref` 旁置字段和锚点行为与现有编译器对齐；使用相同 JSON 解码器避免大数导致身份检查被跳过。
- **F3**：operation/slot 权威检查从单文件提升到有效 REQ 作用域；无元数据时按 CONTRACTS 文件名推导 REQ。不同 REQ 隔离，同 REQ 同源复用允许，冲突源给出明确诊断。
- **文档**：规划 skill 的依赖规则改为真实共享实现依赖；S3 分解表、流程图和任务说明统一为模型/SYNC 共同收敛后 FE/BE 并行。安装规则补充批次级权威与冻结集成约束。

三个失败探针已移除环境变量开关，成为默认回归。F2 测试更名为 `TestAuditSharedModelDuplicateStableID`；前文带 Probe 的命令仅保留历史记录。新增回归验证：冻结文件修改/删除/重命名被明确拒绝、目标与 worker HEAD 保持、worker 保留；无关实现正常 non-squash 合并及清理；合并边界再次检查；同源复用、不同 REQ、传递依赖嵌套冲突、数据 `$id`、相对资源及 draft 边界。

修复后的定向验证命令（无需环境变量）：

```sh
go test ./internal/integration -count=1
go test ./internal/sharedmodel ./internal/cli ./internal/transition -run 'Test(AuditShared|ConsumerProviderPilot|ResourceIdentity)' -count=1
go test ./tests/system/req039 -run 'Test(S3Audit|AuditSharedModelHookSpine)' -count=1 -v
```

上述定向回归通过。此前第一次修复后全量遇到工作区并行 S4 dispatch-plan 编辑及测试重命名的编译中间态，不能计为通过；没有为规避该失败覆盖并行改动。最终全量与 doctor 结果另列下文。

修复收尾验证：

- worker 修改、删除、重命名拒绝及无关实现合并：主代理独立运行全部通过（1.941s）；子代理另外完成 `internal/integration` 和完整 `tests/system/req039`（48.705s）。
- 最终 Schema/Reading/transition 定向测试：全部通过，分别为 0.938s、0.111s、3.902s；日志 `/tmp/s3-fix-targeted.log`。Schema 资源语义的六个边界用例也通过。
- `git diff --check`：通过。
- 中间一次全量和 doctor **未通过**：工作区期间出现的 S4 变更将 TASK 模板中的 `Team manifest:` 移除，但当时 `internal/migration/templates.go` 仍要求它，导致 `TestHarnessTemplateMigration`、`TestS2DualTrackConvergenceE2E` 和 doctor 失败；`TestS4TasksCheckEmptyRoot` 当时仍期望旧的 JSON 模式 exit 0，实际返回 1。这些文件未由本次 S3 修复修改。历史日志 `/tmp/s3-fix-final-full.log`、`/tmp/s3-fix-final-doctor.log`。

随后并行 S4 变更完成同步，上述三个失败用例复跑通过（`/tmp/s3-fix-s4-recheck.log`），doctor 再次通过（`/tmp/s3-fix-latest-doctor.log`）。本次没有覆盖并行实现来规避验证，也没有把中间态结果冒充全量通过。

**最终结果**：`go test -p 2 ./...` exit 0，全部通过，包含默认缺陷回归、S3→S6 Hook 链和 worker 集成；`tests/system/req039` 74.597s。日志 `/tmp/s3-fix-verified-full.log`。最终 doctor 和 `git diff --check` 均通过。独立 Luna/max 子代理只读复核未发现新增阻断性问题。修改保留在工作区，未自动 commit。
