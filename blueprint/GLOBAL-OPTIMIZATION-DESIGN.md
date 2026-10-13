# 全局优化落地设计

> 定位：独立于 L1–L4 的全局实施设计；承接 [全局优化指南](GLOBAL-OPTIMIZATION-GUIDELINES.md)。
>
> 状态：设计v0.1.0的可行性审查已完成，按用户授权进入实施；逐切片状态见§7。G01–G08、O01–O10、M01–M10 均提供具体方案和验收。本文出现的新包、字段、API、Schema、命令和测试均是拟新增项，除明确列为“现有”者外，不表示已经实现。
>
> 调查基线：2026-10-13 的仓库代码、协议和模板；初次调查交付为设计及可行性审查。后续实施先固定L层，再修改文档/代码，保留原权限、角色独立性和模型默认。

## 1. 交付范围与判定

每一点完成设计须具备：现有代码证据、具体差距、改动位置、输入权威、接口或数据结构、实际送达与接收路径、失败恢复、验收用例和兼容迁移。只有 guideline 或“增加提示”不算完成。

专业语义仍由对应角色判断；既有权限、独立性、冻结基线、Writer/journal、真实集成、fresh S7 和人闸保持。目标是接住真实工作，减少执行者解释机制的负担。

## 2. 逐项设计入口

| 项 | 设计入口 | 核心落点 |
| --- | --- | --- |
| G01 | [上下文与技能](global-optimization/G01-context-and-skills.md) | 统一角色/派发技能选择，知识加载与权限分离，稳定工作包 |
| G02 | [质量与完成边界](global-optimization/G02-quality-and-completion.md) | 收尾断言、未决问题归属、职责回执和指令范围 |
| G03 | [任务粒度](global-optimization/G03-task-granularity.md) | 现有 TASK 分母/DAG 上的上下文与耦合分析 |
| G04 | [项目惯例](global-optimization/G04-project-conventions.md) | 有批准来源的默认选择和可重建参考视图 |
| G05 | [验证复用](global-optimization/G05-verification-reuse.md) | 检查回执完整性、适用性判断、受限复用 |
| G06 | [派发与集成](global-optimization/G06-dispatch-and-integration.md) | 原看板上优先接收与反压，保持真实容量与覆盖 |
| G07 | [执行环境](global-optimization/G07-execution-environment.md) | 检查 profile、平台探测、准确失败类别与恢复 |
| G08 | [评估与观测](global-optimization/G08-evaluation-and-observability.md) | 案例重放、零上下文实战、指标与模型实验 |

## 3. 共用基础：当前工作包

### 3.1 调查结论

现有 [team.GenerateReadbackRequests](../internal/team/launch.go) 已聚合文档、技能、scope 和收尾引用；[Hook 生命周期上下文](../internal/hook/lifecycle.go) 已解析 Assignment；[工作区 prepare](../internal/cli/workspace_command.go) 已给精确操作命令。应将它们的公共投影统一，复用已登记事实。

拟新增 `internal/workpacket`，只负责快照到执行视图的确定性投影。它不派发、不修改角色权限、不批准、不建立第二调度状态。原生 subagent、teammate、工作区 Worker 和 Main 的当前责任采用不同入口 adapter，消费同一投影。

### 3.2 包结构与来源

| 字段组 | 权威来源 | 作用 |
| --- | --- | --- |
| binding | 当前 Runtime、REQ generation、真实 owner、Assignment revision、阶段进入事件/轮次、execution generation | 防止旧责任、旧卷和错误身份被消费 |
| job | 当前领域责任、已批准 TASK/Claim/Case/Contract | 目标、本人交付、验收断言、当前动作 |
| inputs | 原 file_sources 的已解析 Git tree 或明确 disk evidence，原阅读清单 | 具体路径、版本、摘要、章节与阅读目的 |
| capabilities | 原角色/Assignment/activation/Runtime/策略的交集及执行环境 | 有效可写范围、真实 cwd、检查入口、操作 argv |
| questions | 与 Harness 同版的注册模板及当前领域展开器 | 本次题目、条件、值域、例子、answer 初值 |
| responsibility | 当前领域接收记录与依赖事实 | 本人完成状态、等待项、下一 owner |

主目录视图路径拟为 `.claude/work-packets/<runtime>/g<generation>/<responsibility>/<packet-id>.json`；Worker 安装为自己的 `.claude/inputs/work-packet.json`。独立Worker目录内填写 `.claude/submissions/stage-answer.json`；共享Main checkout的原生subagent/teammate各填写 `.claude/submissions/<runtime>/g<generation>/<responsibility>/<owner>/<attempt>/answer.json`。工具按发行记录生成安全路径，owner不直接拼接未校验字符串。每个当前attempt只有一个独占入口，发包给完整路径，不要求作者选择目录。Main的只读汇总不供并行手写。

`packet-id` 由发行时的真实 binding、模板版本和输入闭包派生，不能以全局 revision 变化强制重发。阶段进入事件、Assignment revision 或 round 必须区分重入；发包记录使用原 Runtime 的版本化扩展，包文件仅是可重建视图。没有真实可识别 owner 时不给可执行的签名题，也不猜“唯一等待者”。

### 3.3 统一投影与消息送达

拟定接口：`Build(snapshot, fileview, recipient, domainFacts, hostCapabilities) -> Packet`。`domainFacts` 由具名领域 adapter 提供，包含真实 gate 缺口及现有计划/报告/检查点；不得再由文件存在启发式替代领域结论。

只消费当前合法候选/实际选定路由的门缺口，区分可执行责任、已运行责任、其他owner等待与系统动作；不能把所有候选门的missing拼成待办，尤其不能把未触发的人闸/失败分支缺证据要求交给普通执行者。`workpacket`只依赖投影类型与只读接口，不import CLI或Writer；具体领域聚合在CLI/submission adapter完成，避免调用环。

status/next/Hook/launch 将旧字段保留为兼容渲染，新增 `work_packet_ref` 与结构化当前动作。已支持的角色身份决定 recipient；Main 看派发/接收待办，Worker 看自己的作业。安全决策仍在 policy 中，Packet 中的可用操作只是其投影。

有界原生消息先放 packet-id、本人状态、唯一动作、填写/资料路径、等待 owner，再放范围摘要。800 字节限制按 Unicode 边界处理；长 scope 不得挤掉唯一动作。详细资料始终可从包路径读取。

### 3.4 Schema 与版本

现有 `next.schema.json`、`loop-state.schema.json` 的相应对象禁止未知字段；不能直接追加状态字段。新增 `work-packet.schema.json`、authoring Schema、submission receipt Schema，并同步 Go 类型、内嵌资源、catalog 校验和安装测试。

拟采用下一 Runtime Schema 版本 `1.2.0` 承载可选优化配置、发包/接收索引；具体发布前确认该版本未被其他改动占用。旧 Runtime 保持旧形状与旧入口。新字段启用前须升级所有写入二进制，显式迁移 state/journal 配对并保留原恢复标记；旧二进制拒写新数据，不能降级后抹字段。现有 install 只支持新/空目标，不是活跃项目升级器；首版以新安装和旧Runtime兼容shadow为入口，不借install覆盖活跃资产。

当前单个内嵌 `loop-state.schema.json` 的 `schema_version` 是const 1.1.0，不能仅替换为1.2.0。新二进制保留两个完整封闭Schema：拟 `loop-state-v1.1.schema.json` 与 `loop-state-v1.2.schema.json`；在 `schema.Validator.ValidateBytes` 对逻辑名loop-state.schema.json增加固定版本分派，先读实例版本，再以固定白名单选对应Schema，未知/缺失版本按原拒绝规则处理。Writer、semantic/doctor、recovery、实例校验都走同一分派，不根据作者URL或路径选Schema；原v1.1内容保持其封闭约束。

新二进制对v1.1正常写入仍产出v1.1，shadow仅在工厂实验/读投影中计算新包，不写configuration.optimization、发包或接收索引，不开启新consumer。显式迁移后才允许v1.2字段。测试必须涵盖新写入器对旧state/journal的普通更新与pending恢复、对新版本的写入，以及旧二进制拒写新实例；不能只测新安装。

本设计中 `configuration.optimization` 的候选开关为 packet、context selection、receipt、check reuse 和 dispatch advisory；绑定当前 REQ 时固定配置与版本。新实验不暗改活跃责任；先 shadow 对照，启用后由 original domain gates 判定。

## 4. 共用基础：交付消费与当前回执

### 4.1 现有事实与新增落点

现有 `assignment.CompleteTask` 已原子派生 S6 证据；`review.SubmitResult` 已在一个事务内消费 S7 Result、推进报告/覆盖并封 Batch 或 CleanRound；`runtime.RecordEvidence` 提供通用证据登记。现有函数未统一提供 authoring 模板接收和跨入口幂等键，不能直接承诺外层包一次调用即全部原子完成。

拟新增 `internal/submission` 作为登记模板、接收 answer、调用具名 adapter 和投影 receipt 的应用层。领域 adapter 必须调用原 API，不执行模板中的任意表达式/shell，不持 Writer 锁再递归调用领域 Writer。

### 4.2 接收顺序

1. 宿主从真实成功保存事件取得目标路径；仅处理已发行当前包的独占输入。读取大小有界的完整 JSON，解析失败保留草稿。
2. 通过真实 session/execution/Assignment 绑定 owner；冻结文件字节，并核对发行记录、题目版本及原输入闭包。
3. 校验 answer，null/条件未完成返回字段定位。计算规范化答案摘要；作者不得填写 producer、hash、generation 或内部事件。
4. 从原 API 生成领域 payload，使用稳定 `submission_identity` 与独立content hash。领域事务对同身份的消费强制唯一，校验语义/权限并登记原完成事实与接收索引；同身份不同内容拒绝，同内容重复消费返回原记录。
5. 释放领域事务后，以稳定followup事件及原cursor调用Controller求门；转换事务同时登记该事件结果，跨恢复也至多一次允许转换。长测试/合并派生Main/Integrator待办。
6. 从当前 Runtime 投影 receipt 和下一工作包。已经登记而求值/投影中断时明确系统恢复，不要求执行者重做判断或造信封。

### 4.3 幂等与崩溃恢复

`submission_identity=(runtime,generation,packet,slot,owner,attempt)`不包含答案内容，其摘要作为submission_key；content hash独立存储。领域请求新增可选身份与content hash，Writer锁内检查唯一并写接收关联；两个不同答案对同packet/slot/attempt只能一个成功，另一个返回content_conflict，不能因摘要不同得到第二合法消费键。原 `RecordEvidence` 拒绝重复ID、`CompleteTask`会派生修订ID，不能仅靠外层文件缓存声称幂等。

首次接收后同attempt不可变。修改已接收答案须通过原领域允许的修订/返工事件得到新attempt与新packet；author不能修改机器header自造attempt。原领域无修订入口时必须先实现具名入口，不能泛用supersedes。S5拟新增 `RecordDocumentReviewRevision`：只在同一S5进入事件、同冻结链、原真实owner且旧接收匹配时由Main发行有具体修订原因的新attempt；新结论接收时在同一事务把旧evidence标superseded/invalid、登记新evidence与关联。不改旧envelope。已离开S5或规格改变走原规划/需求返工，人闸保持。并发修订均绑定同previous receipt，其中一个提交后另一个拒绝/重新发题，不无限CAS重试覆盖。

统一索引记录binding、authoring hash、adapter/version、domain artifact refs和待恢复步骤；它描述消费检查点，不决定生命周期。跨域步骤保留明确的domain_committed与后续待办，每个已提交步骤有可核验原记录。响应丢失从原记录重建回执；修订按上述具名领域事件执行。

Controller后处理使用 `followup_event_id=hash(submission_identity,"after-domain")`。领域提交在其原事务中记录post-domain cursor、进入事件/轮次和control_pending。拟扩展ControlRequest/transition.Request携带该事件及expected origin；已有RequestID用于journal关联，但现有按transition/revision生成的幂等键本身不够。转换Apply在同一个Writer提交中检查事件未处理与origin仍匹配，并写followup结果/transition event ref；stage变化和事件处理标记不能分两次写。

恢复先查该事件结果及原journal（包括marker恢复），已转换仅重建当前回执/下一包。当前cursor已由其他事件推进时标followup被当前事实取代，不把旧提交应用到新cursor。求门not_ready且没有转换时记录evaluated_no_transition，未来新的交付事件有自己的followup；unknown/超时可保持pending，但只能对原origin有界重试。重复保存只返回原接收和当前事实，不再开启新控制请求。受影响origin发生返工重入也不得匹配旧事件。

包过期返回具体变更及重新发题责任；不把当前新输入 hash 盖到旧 PASS 上。包登记成功、视图丢失则重建，不重新署名。故障/unknown gate 不判通过，原安全边界保持。

### 4.4 回执字段与职责结束

拟定 `Receipt`：`submission_key`、`accepted`、`responsibility_status`、`work_complete`、`current_stage`、`domain_refs`、`waiting_on[{reason,owner,ref}]`、`next_action`、`field_errors[{pointer,expected,observed,repair}]`。

`work_complete=true` 只代表本次责任的交付被原消费者接收，专业失败结论也可能是有效交付。它不代表阶段通过、代码集成或发布授权。仍欠自己检查/结果时为 false；仅等待别人或系统时不向本人重复发题。

receipt 的历史引用与当前状态分别投影；重复旧答案不能复活被 supersede 的结果。受影响责任返工时明确原领域事件与新 attempt，不能在工作已完成回执里静默重新授予产品写入。

单份S5规格审查接收后的回执示意如下。这是接口示例，不是已经实施的Schema或真实证据；真实binding与refs由工具提供：

```json
{
  "accepted": true,
  "responsibility_status": "accepted",
  "work_complete": true,
  "current_stage": "S5",
  "domain_refs": ["document-review-spec-example"],
  "waiting_on": [{
    "reason": "另一项独立文档审查尚未交付",
    "owner": "reviewer-task-example",
    "ref": "DV-TASK-EXECUTABILITY"
  }],
  "next_action": {
    "owner": "reviewer-spec-example",
    "kind": "finish_responsibility",
    "instruction": "你的审查已接收，可以结束本项责任。"
  },
  "field_errors": []
}
```

所有receipt枚举在发布的Schema固定，界面文案可翻译。waiting_on对应下一个工作的真实owner；next_action对当前recipient说明如何结束或继续，不靠“current_stage仍在S5”推出本人还有工作。

### 4.5 首个领域 adapter：S5

现有 S5 要求两个独立 producer，职责为 DV-SPEC-CONSISTENCY 与 DV-TASK-EXECUTABILITY，原 gate 消费 `document_review` evidence。第一版每人只填结论、发现和引用。

adapter 从冻结规格链与真实派发生成 review evidence envelope，经 `runtime.RecordEvidence` 的原校验登记，Controller 继续选择原 PASS/FIX/REQ-change 路由。登记 API 须扩展本节的同事务接收键；不得只给 envelope 文件、不登记证据。不能绕过 producer 独立性，Main 不替 reviewer 签字。

两人都 PASS 才能原门推进；首人完成返回等待另一 owner；FIX 交给原规划返工；REQ-change 消费原人闸。未满足原失败路径前提时给对应责任，不自动跳阶段。真实 Claude 保存事件、PostToolUse 回执、重复保存、基线漂移和重入须实测；原型不代替这些测试。

## 5. 交付策略的实施卡

### O01 — 可开工工单与预检

**改动：** `workpacket.Preflight` 消费原Assignment/领域计划/冻结资料/host capability。workspace prepare/launch与原生派发前置入口共用领域就绪报告；检查内容包括必需input存在与source一致、当前owner/依赖、计划屏障、scope、checks及环境。就绪报告是投影，不给未授权产品动作开权限。

**责任：** Main修派发缺口，规划者补实际收尾/决定，环境owner补运行条件。S6/S9计划阶段可派发合法计划责任，执行题仅在原条件满足后出现；S7先执行可开工的static责任，依赖未满足的behavior责任保留排队。

**失败恢复：** typed `PreflightIssue{code,source_ref,owner,required_fact,repair}`；输入/环境修复后重探测，已准备实例按固定binding恢复。unknown能力不标ready。

**验收：** 缺文档、未提交正式输入、blocked依赖、缺plan、缺环境、scope错误、原生/工作区入口等价性；合法计划仍可交，未授权产品写入仍拒绝。关联G01/G03/G07、M01/M08。

### O02 — 唯一交付责任与视图分区

**改动：** TASK本人交付表、原manifest ResponsibilityIDs与领域分母生成 `owned_outputs`；其他责任通过 `downstream_receivers`只读展示。新workpacket测试禁止把QA/E2E独立判断列为Builder签名题。

**责任：** 规划者分配，Main派发，角色按原职责交付。原S5两人、S7 mandatory/risk责任及separation_edges完整保留。

**失败恢复：** 无owner、重复owner或角色冲突给规划/派发具体来源；不让Worker自领或Main代签。已登记owner修订走原manifest/Assignment路径。

**验收：** 同一builder scope无独立审查签名、required responsibility无遗漏、跨owner提交拒绝、两个独立producer不能共用实际actor。关联G02/G06、M04。

### O03 — 行为样例、断言与oracle

**改动：** workpacket的checks/examples引用已批准合同与当前模块cases.json；S7直接展开原Claim assertion/oracle/method/required_evidence，S9展开approved Contract assertions。S6 Closing Contract无法机械解析时发专业题给规划，不执行任意文本断言。

**责任：** 上游设计者确定预期，Builder实现自己的检查，独立审查者验证与找遗漏。机器只枚举来源与实际运行事实。

**失败恢复：** 预期未决定、用例/引用失效、命令不存在，分别给source和owner；不自动生成语义PASS或补业务规范。

**验收：** 正常/拒绝/无数据例子、缺副作用期望、旧case、不可运行oracle、测试与契约矛盾、新业务决定未被惯例补齐。关联G03/G04/G07、M01/M06。

### O04 — 有效操作与精确修正

**改动：** workpacket capability adapter复用worker `ObservationCommands`、`WorkerInvocation`、prepare check_commands及Main pending argv，补领域固定操作生成器。`next.Action`从结构化action渲染，诊断同步扩展typed owner/source/JSON Pointer/argv；保留旧错误文本供兼容。

**责任：** 工具从真实binding填参数，policy仍求有效权限。专业参数如rework原因以题目收取，答完后生成命令，不把 `<reason>`渲染成可直接执行动作。

**失败恢复：** 缺actual identity/cwd/能力是派发接线待办；unknown token归adapter维护。状态过时重新投影，实际命令仍由原域校验。

**验收：** 空格/中文路径的argv、所有原门实际missing有映射、无占位命令、合法/越权动作与同角色跨工作区、未知门不猜成功。关联G01/G07、M02/M07。

### O05 — JSON注册、发题与保存消费

**改动：** 内嵌template catalog含Stage/责任/authoring Schema、问题/选项/条件/例子和具名adapter；首版S5。authoring与persistent evidence Schema分离；注册验证每题有消费者、每个手续字段有机器生产者。

**接线：** 同时更新Main `settings.json`和Worker Bootstrap生成设置；目前Worker为matcher `*`，仅扩大Main matcher并不能完成消费。`runHook`的PostToolUse早返回分支新增受控提交分支，Write/Edit/MultiEdit按真实payload取路径，其他文件仍是观察路径。受控保存adapter可直接调用同consumer；Bash只有声明的写入adapter可触发，不扫描任意shell字符串猜文件。

**写入边界：** 当前workspace boundary允许自身submissions目录，S9的repair control路径也包含submissions；但 `policy.reviewerWritePathAllowed`（S7）和 `phaseWritePathAllowed`（S8/S10）没有该入口，不能假定新增路径已被允许。拟新增具名authoring-path判定：只允许真实actor当前发行记录的独占答案文件，经原物理路径/symlink检查；作为这些阶段的控制面输入例外，不放开整个 `.claude/`，不授产品写入。确切独占文件校验不能被“仍在仓库内”代替：规范化发行根以后，受控答案路径下级不允许symlink重定向，尤其不能指向另一owner。工作包/Runtime/汇总仍只读，其他owner、旧attempt和未发行路径不给此例外；consumer再独立验证同一binding。原来允许的证据/报告路径继续走原规则，不能用新入口绕过计划屏障或签署别人的判断。workspace中也对受控答案应用此绑定，避免仅靠目录前缀接收。

**提交判断：** 机器由必答题、结论和条件题计算draft/complete，不增加Agent手填“ready”事件。null/缺条件是needs_answer；完整答案立即接收，非法值/引用给拒收；监测草稿与完整提交分别统计。读取损坏/尚未完成的JSON保留草稿和定位，不形成PASS。规划/结果分别发行当前题，先后由原授权保证。

**无Hook：** 只有宿主受控保存或明确tool-return consumer实际接线才声称自动；否则给绑定完整参数的兼容入口，不伪称文件保存自动推进。宿主入口沿可信actor上下文，不收author-supplied身份。

**验收：** 实际官方payload/wire、共享Main目录两人并行且答案不覆盖、Worker独立路径、各阶段合法保存通过PreToolUse后才触发消费；其他owner/旧attempt/未发行文件、路径逃逸、指向仓库内另一owner及仓库外的symlink不能取得新例外。再验最后保存无额外工具、草稿/条件/非法值、权限/输入漂移、所有支持宿主与未支持声明。关联G01/G08、M05/M06。

### O06 — 自动派生事实与合并重复载体

**改动：** submission adapter从真实Assignment、diff、执行回执、证据引用和已提交计划生成原消息/envelope；Markdown报告与兼容BUG仅做可重建投影。S6复用CompleteTask，S7复用SubmitResult，S8/S9/S10按下表分开调用原域API。

| 领域 | 已有入口 | 新adapter留下的专业答案/具体接续 |
| --- | --- | --- |
| S2/S3/S4 | RecordEvidence及原planning转换动作 | 本层决定、资料路径、覆盖/依赖；自动原planning envelope，转换继续登记原documents |
| S5 | RecordEvidence、原document gates | 独立结论与发现；见§4.5 |
| S6 | assignment.CompleteTask、agent-begin与Main task-integrate | 当前计划、实现/检查引用、结果；长集成归Main |
| S7 | review.RegisterPlan、review.SubmitResult | planner的覆盖/oracle和reviewer的Claim判断；hash/Result绑定由工具生成 |
| S8 | investigation intake/case_workflow/contract | 假设、因果、证据、路由、修复意图；每步形成原case checkpoint，人/已有授权审批条件照旧 |
| S9 | repair runtime/result/impact/targeted/handoff | 当前计划、修复结果、独立复验；create/commit外层接续并保留中断checkpoint |
| S10 | acceptance/audit验证、RecordEvidence | 有限inventory逐项判断、反证/风险；自动信封和来源绑定 |
| S0/S1/S11 | 原锁定/绑定/人决定入口 | 发待确认稿与人决定；显式决定到达才留证，不从文件存在推断 |

**重复计划：** S9generic与domain计划保留不同消费者。同一专业计划答案生成共享事实；generic的实际plan bytes绑定先完成，domain红色前置检查仍需真实证据和独立登记。所有Assignment domain reports齐全后原execution begin才开放写入，不能在最终completion追认。

**失败恢复：** 跨域接续复用§4的同事务索引/步骤原记录，不能外层抱Writer锁调用多域写入。不变事实引用来源，派生失败归系统恢复。

**验收：** 原Schema合格、每手续字段有机器来源、专业值不被自动补齐、报告与evidence一致、S9双计划时序及中断、S10无人工hash。关联G02/G05、M06/M07。

### O07 — 最小修正与缺口映射

**改动：** 从Schema ValidationError提取JSON Pointer/允许值，复用现有discriminator pruning保留原详细错误；领域diagnostic补source/owner/expected/observed。gate code/prefix注册到具名gap adapter，mapping test遍历实际候选门与fixture错误矩阵。

**责任：** answer错误归当前作者，资料/决策缺口归原owner，integrity/未知映射归Main/维护。系统知道的事实不能出现在补答题里。

**失败恢复：** 已接收其他slot保持，当前修订保留历史；相关输入变更精确失效，未证明依赖闭包时沿原保守重验，不静默保留旧PASS。FORM_GATE_UNMAPPED有实际缺口与维护责任，不能“继续DRIVE”。

**验收：** 单字段缺失不重填、非法分支只显示该分支错误、不同owner缺口不串、未知gate安全路径、输入漂移与旧引用。关联G05/G07、M02/M07。

### O08 — 责任完成回执与离场

**改动：** Guidance/next/packet引用§4.4 receipt；Stop/Idle从原本人交付与待项求值，Main得到后续派发/集成动作，Worker只见本人工作。

**责任：** 消费者接收，控制面投影，宿主给执行者真实additionalContext/tool return。工作包与receipt引用不能代替真正Result/ack。

**失败恢复：** 旧回执回放报告原记录及当前stage/等待；同事变更REQ或自身结果superseded时不继续旧动作。已接收后的系统pending不能迫使本人重交。

**验收：** 两reviewer先后、本人PASS但另人REQ-change、合法Finding交付、等Main集成、等系统恢复、已完成者Stop、本人缺结果仍受原规则。关联G02/G06、M04。

### O09 — 持久交接与恢复

**改动：** 复用原completion/assignment、review plan dispositions、case/repair pointers和integration checkpoints派生pending；新增submission关联只记录消费恢复，不维护独立可漂移队列。SessionStart/PreCompact作有界补消费/重建；被阻止的Stop保留原preflight顺序，不先写转换。

**责任：** Main/Integrator完成长集成和接收，领域consumer完成短事务；发布仍归原人闸。一个主会话恢复不模拟平台未支持的wake-up能力。

**失败恢复：** 写前中断→无domain完成；domain已提交→重建receipt/求门；迁移已提交→下一包重建；合并已发生→原checkpoint接续；只有log/文件存在不猜域提交成立。键冲突和输入漂移明确拒收/恢复。

**验收：** 每个持久边界故障注入，尤其“转换已提交但响应/外层检查点丢失”；同事件不能再推新cursor。覆盖多进程同身份不同答案、S5重签唯一前驱、旧答案回放、旧轮/Runtime切换、清理失败、合并不重复及journal marker恢复。关联G05/G06/G07、M05/M08。

### O10 — 无进展诊断与恢复责任

**改动：** G08的work_progress观察同作业/缺口/相关输入指纹，忽略纯时间/revision变动。配置只控制框架重复提醒，未知缺口即刻给维护路径；有真实running或外部owner等待不判空转。

**责任：** Main按原缺口补资料、修派发、恢复环境或进入领域返工。专业不确定交具体问题；不增加对人无必要的审批。

**失败恢复：** 达阈值给一次可定位stall而允许合法idle，不将未完成改完成；新证据后恢复原责任。诊断缓存损坏不关安全门，也不无限反复发送同一句提醒。

**验收：** 无变化重试、重复read/status、metadata churn、新证据、正常长运行、等同事/人、缓存损坏、不能自动替换writer。关联G08、M09/M10。

## 6. 覆盖、验收与发布边界

调查、已有测试运行、新设计评审、拟新增测试和真实平台验收分别留证；不把设计覆盖标成实现完成。迁移与启用以最小完整领域链为单位，未启用的入口保持旧路径并明确限制。

### 6.1 M01–M10覆盖矩阵

| 缺口 | 具体方案 | 必须证明的执行结果 |
| --- | --- | --- |
| M01 当前具体交付 | §3、O01/O03/O05、G01/G03 | 不查流程格式即可从包找到题、路径、检查及收尾 |
| M02 真实缺口投影 | §3.3、O04/O07 | 同快照status/next/Hook的动作/owner一致；机器转换不派为手写产物 |
| M03 必需信息送达 | §3.3、G01 §4/6 | 长scope/UTF-8/恢复时仍保留身份、next与包入口 |
| M04 自己完成 | §4.4、O02/O08、G02 | 本人被接收且仅等其他owner时可以离场，无重复催交 |
| M05 最后交付接收 | §4、O05/O09 | 最后保存直接消费/求门或有明确系统pending，无无关工具 |
| M06 可填写草稿 | O05/O06、G01/G03 | 已知值预填、专业题可答、缺上游有owner，没有内部Schema猜测 |
| M07 可执行修正 | O04/O07、G07 | 精确参数与字段定位；未知映射移交维护，原门不放宽 |
| M08 统一开工检查 | O01/O09、G07 | 每个支持入口有领域预检，真实平台不支持被准确报告 |
| M09 普通无进展 | O10、G08 §5 | 同缺口无新信息可诊断，长执行/等待不误判，安全门保持 |
| M10 首次使用验收 | G08、O01–O10 | 完整真实作业完成/失败/恢复，模式/trace缺失如实标记 |

### 6.2 实施切片与依赖

1. **基础投影**：typed domain facts、工作包、字段错误、本人责任视图；shadow核对真实gate。先不切旧手工入口。
2. **S5完整链**：authoring模板→真实独立owner→保存消费→同事务接收键/原evidence→原门/回执→中断恢复，连带Main与Worker设置。通过真实宿主后再删除该入口手工手续。
3. **上下文与质量**：G01选择器和权限等价测试→G02责任/指令→G04参考卡；用G08固定案例对照，不同时改模型。
4. **其他领域adapter**：S2/S3/S4确定性登记，再S6/S7/S8/S9/S10；每个领域先证明语义/时序/恢复，不能从S5通过宣布全部自动。
5. **任务与调度**：G03 advisory→G06接收优先→显式配置反压，基于已有DAG和scope。
6. **检查与环境**：G07 profile及真实平台→G05回执v2与shadow→小范围同职责复用。独立/合并后/fresh检查不变。
7. **观测与stall**：G08有限指标与实战runner贯穿，通用stall先shadow，充分反例后再启用提醒预算。

每个切片有自己的旧入口fallback与版本边界；fallback由工具生成真实参数，不让执行者猜命令。

### 6.3 迁移与回退的具体边界

首版仅在新安装中发行新Schema/内嵌资产，旧项目运行兼容渲染或shadow。当前install拒绝非空目标，不能用于更新活跃项目。

若后续支持活跃迁移，须单独实现维护入口（拟 `runtime optimization migrate`）：核对匹配release manifest/installed receipt；确认无活跃launcher/写入并保全pending；备份state/journal成对及资产；在临时副本验证纯Schema变换；原Writer持锁提交migration事件与配置/索引，沿marker恢复；只对可证明未自改的框架资产发布已验证版本，项目改写资产给明确合并报告，不覆盖。

变换只加版本化可选字段，不改bound REQ、generation、专业结论、权限、旧owner或已提交证据。Worker ready资产不重写，新profile/模板在下一合法发行使用。新Runtime曾提交以后，不用旧二进制回写；功能开关关闭可恢复旧界面/动作，但继续保留新Schema和历史数据。完整恢复需匹配的版本与备份，不能撤销已经发生的Git集成。

该迁移入口本次未实现，未验收前不能在旧项目启用需要新持久字段的consumer。

### 6.4 本次已有测试抽查

执行了team、hook、semantic、dispatch、workspace、integration、metrics、qualitygate的相关既有用例，覆盖readback包、生命周期上下文、冻结输入/bootstrap/launch、DAG/容量、真实board、合并检查/漂移、指标预算和planning门。首轮七个包的选择用例通过，workspace三个用例在根目录夹具验证失败；设置规范化 `TMPDIR=/private/tmp` 后三个workspace用例通过。

此结果验证可复用基础，并暴露路径别名边界；不是完整Go套件、未实施功能或真实Claude的验收。G07记录对应代码和拟补测试，G08区分这些证明层次。

重跑命令与范围：

```sh
go test ./internal/team ./internal/hook ./internal/semantic ./internal/dispatch ./internal/workspace ./internal/integration ./internal/metrics ./internal/qualitygate -run 'Test(GenerateReadbackRequests|BuildLifecycleAdditionalContext|BootstrapResumes|FrozenInputChanges|LaunchUsesOnly|DispatchNoWaveBarrierAndCapacity|DispatchPlanBoundaries|CommittedBoardAndRegistration|PlanCannotDriftOrBecomeLegacy|PostMergeChecksUseCandidate|VerifiedRejectsTargetDrift|TrackedFamilyBudget|PlanningGatesReadDiskDeclaredArtifacts|IdleRecovery)' -count=1
TMPDIR=/private/tmp go test ./internal/workspace -run 'Test(BootstrapResumesAndPreservesChangedAssets|FrozenInputChangesBlockPreparation|LaunchUsesOnlyChildDirectoryAndRegisteredSession)' -count=1
```

### 6.5 可行性审查结果

独立审查核对了八份分项设计、O01–O10与M01–M10覆盖，以及代码中的真实入口/权限/恢复边界。首轮指出的五项问题已修改并复核关闭：提交身份与内容摘要分离、Controller跨事务持久去重、两版完整Schema分派、Worker沿实际Read路径加载技能、普通Read/status的生产观察接线。

后续补充共享checkout的owner/attempt独占答案路径及各阶段写入策略接线，经独立复核未发现新严重遗漏；并将“仓库内symlink指向另一owner”明确列为拒绝用例。通用stall另明确首次读取新必要资料是进展，观察不足不强判空转。

审查结论为设计契约具备可实施性；没有运行新receiver、故障注入或真实Claude，不替代§6.2各切片的实现与行为验收。首个实施单元仍是S5完整发题→保存→接收→原门→回执→恢复链，随后按分项依赖展开。

## 7. 实施状态（2026-10-13）

本轮按用户授权进入实施，先固定对应L层契约再改代码。L2新增本人收口规则；L3-S4/S5固定任务收尾与双责任答卷；L4阶段答卷§13固定生产契约，dispatch、hook、runtime、design foundation及workspace profiles分别承载G01–G08的具体消费边界。L1现有D1–D7与独立性原则保持。

| 实施单元 | 当前事实 | 后续必须完成 |
| --- | --- | --- |
| Runtime版本与封闭Schema | 已实现1.1/1.2固定分派；旧Schema保留，新字段独立；schema测试通过 | 新安装默认、旧实例实际写入/恢复与启用边界的端到端验证 |
| S5工作包与题目 | 已实现纯投影、内嵌JSON模板、独占路径、冻结题干及字段定位；单元测试通过 | 注册/启动时实际发题、发布安装资产、Worker路径与宿主送达 |
| S5领域接收基础 | 原RecordEvidence事务同时写接收索引/followup pending；重复/异内容/并发/漂移测试通过 | authoring consumer、具名修订、Controller转换去重、回执/恢复、Stop与权限接线 |
| 后续领域及G01–G08 | L层实施契约已固定，生产功能未完成 | 按§6.2切片逐项实施与验收 |

相关验证：schema/workpacket/runtime全包与race测试通过；semantic/assignment/controller与req039系统回归通过。CLI首轮九项受真实工作目录旧Runtime污染，在原HEAD干净目录复现为通过；改为隔离工厂夹具后九项及完整CLI回归通过。本表只记录已执行结果，完整S5链、真实Claude及其余优化尚未完成，不将基础API当保存自动化交付。

基础切片独立审查发现一项待修复P1：新增Submission接收路径尚未验证真实答卷，也未将调用方摘要与document_review的conclusion/findings/完整subject_refs绑定；非法信封可能占用不可变attempt。下一步必须先补拒绝测试，在原Writer事务内验证规范答卷、自行计算内容摘要并核对领域信封，关闭审查问题后再推进保存消费与控制接续。2026-10-13按用户要求先提交并推送当前状态作备份，该备份不是可发布的完整功能。

## 变更记录

| 日期 | 版本 | 内容 |
| --- | --- | --- |
| 2026-10-13 | v0.1.0 | 调查G01–G08与原实现，形成独立分项设计、共用工作包/消费接口、O01–O10实施卡、M01–M10覆盖及迁移/验收方案；修复并复核审查问题，保留待实施/真实宿主验收边界 |
