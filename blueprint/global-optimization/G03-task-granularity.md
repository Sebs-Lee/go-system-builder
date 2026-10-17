# G03 — 任务粒度与上下文耦合的落地设计

> 全局设计；承接 [G03 guideline](../GLOBAL-OPTIMIZATION-GUIDELINES.md#g03--按上下文和耦合程度拆任务)。§7 首个分析器及原 S4/S5 输出接线已实现；工作包冻结分析和行为实验仍待实施。

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

## 7. 首个分析器契约

在原 TasksCheckResult 增加 context_analysis_version、task_contexts 与
granularity_advice，保留原 reference_loads 字符串入口。新分析仅附加资料；不得
增加 Problems、修改覆盖/DAG、自动拆并或更改任何 TASK/权限/派发状态。

按原 Document Manifest 逐行保留 path/anchor、Mode、Purpose、Clauses 和 TASK
出处。没有 Mode 的旧行标 legacy_unspecified，conditional 始终保留条件与
condition_unknown，不自行判满足。路径沿原 sharedmodel.Resolve 处理链接，裸
路径沿原 repo-relative 规则；越界/远程/多个链接/模板占位记录无法解析。

每个文件内容在分析中缓存一次、解析上限1 MiB（既有 Reader 的底层读取不伪称
受该解析上限限制）；目录不递归展开，显式列 directory_unexpanded，
读取失败列 unreadable，不伪装0字节成功。Markdown 标题/中文 slug/明确 heading
anchor 唯一且可定位时，按该标题到下一同级/上级标题统计字节；代码块内标题
忽略，重复/未定位/JSON anchor 保守计完整文件并说明原因。以文件字节区间并集
去重，分别列 required、conditional、optional、legacy_unspecified 读集与字节；
total_unique_bytes 仅为可读已解析资料，不是 token/难度估计或强制阈值。

Closing Contract 中只提取明确的 `assert <command> == pass` 检查入口及来源；
占位或自然语言断言保留为 unresolved_assertion，不执行 DSL、不认行为 PASS。
同模块写入及重复 required 读集、明确同文件写入、单任务跨模块/跨合同收尾给
带 TASK/路径/条款的 advisory；仅观测耦合，不推导任意大小阈值或自动改批准输入。

S4 的 tasks check --json 与 S5 原 workspace review-check 返回同一分析，文本
入口只显示短建议与 JSON 入口。原 S5 WorkPacket 的输入约束继续固定来源，解析器
版本在分析输出明确；正式工作包额外冻结分析及真实模型成本实验仍需后续切片。

具体实施和验证范围见 [全局设计 §7.14](../GLOBAL-OPTIMIZATION-DESIGN.md#714-g03-原-task-的版本化读集与粒度分析)
与 [安装/测试证据](../evidence/global-optimization-task-context-install-20261013.json)。

## 8. 新 S5 工作包冻结分析

Packet 增加可选 task_analysis {path, sha256, parser_version}，仅新发行包由原
submission issuer生成；旧已发行包保持原身份/题干，不补字段。分析是内容寻址
只读 artifact，包含解析器版本、原 Packet.inputs 精确集合、批准 TASK 的分析
与 advisory。TASK discovery 限原登记 task refs；补充读取仍沿原 file-source
契约，记录实际来源哈希或 unavailable，缺预期不能生成行为真值。

分析引用进入新 Packet identity 和 Draft.read，原正式 inputs/结论/领域 envelope
的 scope 不扩大。原 ValidateWorkPacket 在发行/保存的 Writer 锁内核验 artifact
path/hash/schema/parser版本、完整 formal refs；不是重新运行当前 parser来覆盖
旧分析。具名修订重建保留同一分析引用，checkpoint恢复及历史读取验证其不可变
关系，不反推专业 PASS；丢 artifact 明确需要恢复原字节，不生成替代答案。

新独立 Worker prepare将分析作为原 Execution.inputs/Bootstrap只读输入安装；
原 review-check 的机械 Problems/覆盖/DAG 仍实时从同一原 formal view求值，其
context/advice取已冻结 artifact。额外来源漂移须明确失效/恢复，不能回复旧缓存
的继续授权；分析从来不授权下一步。行为样例只可从批准合同/当前 CASE来源生成，
此切片首先完成分析链，样例来源及真实模型对照保持后续明确待项。

冻结分析链已落地，完整实现与验证见
[全局设计 §7.15](../GLOBAL-OPTIMIZATION-DESIGN.md#715-g03-新-s5-packet-冻结分析与原接收修订恢复链)
和 [安装/测试证据](../evidence/global-optimization-frozen-task-analysis-install-20261013.json)。

## 9. 批准 TASK 引用的当前 CASE 行为样例

只从 TASK Manifest 明确 cases.json 路径及 Clauses 中 CASE-id 选择样例，不扫描
全仓库、不用模块相同推断全部 CASE 都属于本人。Scenario 的原 source-package/
schema/generator校验用于证明当前 cases/coverage 等于源生成结果；不得把手改
cases.json 当标准。原 formal TASK/contract refs 仍是批准边界；样例只是参考，
不能添加需求、改变 coverage 分母、替代 Builder owned checks 或独立 S7。

投影保留 CASE/rule/branch/polarity/required/browser_required、原 oracle/fixture，
并保留原读集 Mode/Purpose/condition_unknown，conditional 不变成必需条件。
输入仅由原 witness→fact partition.value解析。保留来源闭包 path/hash，含模型、
fixture、cross-matrix、stories/flows和当前 cases/coverage。
原 Scenario 校验实际读取的绑定 REQ/FR 也属于闭包，必须与原 formal REQ 引用及
摘要一致；不能省略 FR join 或退化为 shape-only。投影中途读取同路径发生变化
必须拒绝，冻结的 hash 对应实际用于生成样例的字节。
没有明确 CASE 引用、CASE不存在、源包未验证、缺预期等给具体 unresolved 问题和原 TASK 来源；不填
猜测输入/预期，不将缺样例作为自动开工或放行门。

同一投影供 S4分析、S5 review-check使用，并作为可选 behavior_projection_version
与behavior_examples加入新冻结分析artifact。旧artifact不补字段；新数据进入
内容hash，Worker冻结来源闭包，接收/修订沿原来源hash核验。生成不执行浏览器、
测试或assert，不将标准样例存在解释为当前产品通过。合同自由文本期待不能
自动变成事实；仍要求原CASE/规格责任补齐可观察oracle。

此切片的实现与原接收/恢复/修订测试见
[全局设计 §7.16](../GLOBAL-OPTIMIZATION-DESIGN.md#716-g03-当前-case-行为样例与原-s5-冻结来源闭包)
及 [CASE 样例安装证据](../evidence/global-optimization-task-behavior-install-20261013.json)。
真实产品/浏览器行为和模型成本实验仍为后续验收。
