# G03 — 任务粒度与上下文耦合的落地设计

> 全局设计；承接 [G03 guideline](../GLOBAL-OPTIMIZATION-GUIDELINES.md#g03--按上下文和耦合程度拆任务)。新分析器、字段与测试为拟实现。

## 1. 调查证据与结论

[semantic.TasksCheckWithFiles](../../internal/semantic/tasks.go) 已检查 CONTRACTS 条款全集、TASK 覆盖、完整状态、收尾块存在与 DAG，并给 ReferenceLoads。当前统计按 manifestPaths 累加文件字节及写入路径数量，未证明实际 token、独立验收或任务难度；收尾检查主要识别 assert 文本存在。

[LoadDispatchPlan/NextDispatch](../../internal/semantic/dispatch_plan.go) 已处理已提交规划、TASK 依赖、Resource order、写入冲突与容量；无整波等待。已批准输入不可运行中悄悄拆并。

结论：复用现有任务分析，增加有解释性的粒度建议与可执行收尾投影；不增加一个用任意 token 阈值拒绝任务的门。

## 2. 具体改动与接口

在 `internal/semantic` 拟新增 `task_context.go`，将 ReferenceLoads 改为兼容字符串加结构化 `TaskContextSummary`：TASK ref、required/conditional/optional 读集、去重字节数、实际读取章节、写入模块、依赖和冲突、可解析检查入口、收尾引用、无法解析的项目。

解析原 TASK 表格与确认条件；条件未知时保留 unknown并归规划责任。不能将所有目录引用展开为整个仓库。Markdown anchor存在且范围可定位时统计章节，否则记录整个文件与保守原因。无法读取的必要资料是具体缺口，不算 0 字节。

拟新增 `AnalyzeTaskBatch(plan, summaries) -> []Advisory`，只给观察和建议：读集高度重复且写入相同模块、独立模块但收尾混杂、跨任务频繁依赖、明确同文件冲突。第一版不自动改 TASK、分母、scope 或 DAG。

S4 规划者从工作包回答当前建议；S5 的 DV-TASK-EXECUTABILITY 消费同分析，明确判断独立验收、依赖必要性与粒度。每项建议有源 TASK/条款/路径，不能只说“太大”。

## 3. 收尾与行为样例

TASK 的检查命令从已声明 required_checks/Closing Contract提取到工作包；未知自由文本保留原引用并交规划者补具体检查。不能建立任意 assert DSL执行器，也不能把 `assert X == satisfied` 自动当行为真值。

样例来自批准的合同、当前模块 cases.json 和已有测试，附输入/预期/副作用及来源；缺预期时给专业题。Builder只负责自己的检查，独立 QA/E2E仍在原责任中。

新 TASK模板强调一句交付目标、本人必交、真实依赖、引用式行为断言；不重复上游全文。拆分建议必须保留每条已批准合同覆盖和可执行依赖。

## 4. 执行数据反馈与规划变更

G08记录每 TASK 的 compact、重复读、交接和返工观测。仅有一次 compact不自动判任务错误；分析其资料负担、执行时间与变更耦合。规划者据证据选择裁无关读集、拆分或合并。

已冻结批次需要变更时，走原 planning/review/invalidation路径，更新 TASK及dispatch plan，再验证条款全集、取消依赖和资源/DAG；不能把新取消项从分母删掉而留下覆盖缺口。

## 5. 失败与恢复

- 读集 anchor/条件解析失败：给具体原表格位置，必要项归规划，背景项作有界提示。
- 行为预期未决：归合同/设计责任，不生成猜测测试。
- 粒度建议与工程约束冲突：保留理由，由 S5验收可执行性；字节数不是强制预算。
- 规划建议文件丢失：从 frozen TASK/plan重建；不建立第二任务状态。

## 6. 验收与实施切片

1. 扩展 tasks/shared-model reading测试：重复引用去重、required/conditional/optional、不可读输入、中文 anchor、目录引用上界。
2. TASK只包含 assert文字而无可执行 oracle时，投影明确专业缺口；已有模式不伪造 PASS。
3. 保持现有 `TestDispatchPlanBoundaries`、`TestDispatchNoWaveBarrierAndCapacity`、取消/跨 REQ/资源环/范围测试。
4. 三类实战案例：大且可独立、过碎且共享写入、合理多文件任务。比较 compact、交接、冲突和首交合格率。

先只发结构化 advisory给 S4/S5，得到实际数据后再调整模板；新解析器版本固定在工作包，旧TASK兼容保留原检查。新任务的缺收尾问题仍走现有可执行性职责，不增加“上下文超限”硬门。
