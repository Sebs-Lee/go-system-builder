# L4 — 阶段答卷

> 层：第四层｜机制域：每个阶段一份答卷，判断或状态变更如何进入既有门禁
>
> 上游：L1 D3/D4/D5 与公理二、公理五；L2 各阶段出口；[L4 运行时控制面](L4-runtime-control-plane.md)；[L4 revision](L4-revision-usage.md)；命令链、对外名字、重复载体和占位符的收敛见 [L4 手续收敛](L4-procedure-reduction.md)
>
> 状态：v0.7.6 独立 S5 execution、具名修订与完整 pair 恢复已接线；发行暂停、接收替代、当前 attempt 回执和原 Apply 提交窗口复查的普通验收通过。首个监听/问卷切片已有 r59 六包定向回归通过，native/workspace S6 保存与 Stop 准确路径已接线；真实完整阶段仍为 4/9，S6 Main 集成未闭环。最终回归记录见全局设计§7.5；新安装/init 1.2 和显式 packet、旧格式保持与 native 安装验收见全局设计§7.6；已发行索引的 checkpoint/RR-A 恢复与故障验收见全局设计§7.7；完整安装与真实 Claude 验收仍待完成。另有[隔离的 S5 交互原型](../../tools/stage-form-prototype/README.md)，仅验证填写与回执契约。本文不新增状态机。

## 1. 机制

### 收缩后的目标（2026-10-15，优先于本文旧方案）

阶段答卷是一份**提前说明预期的履职问卷**。Hook 承接 graph engineering 的监听与
流转，提醒 Agent 完成工序、避免跳步和遗漏。问卷把当前责任、工序、所需资料、
每题要求/答案形状、实际交付物和准确填写路径一次给清楚。

正常路径：**看到要求和问卷 → 执行并填写 → 保存 → 更新 graph → 提示剩余工序或进入下一 stage**。
详见[全局设计 §0](../GLOBAL-OPTIMIZATION-DESIGN.md#0-收缩后的实施决策2026-10-15优先于旧方案)。

- Agent 填专业答案、产物路径和具体阻塞；系统已知信息自动带入。JSON 可保留，
  不再手抄哈希/身份/事件，不重复交总结或信封，也不手动发行、接收问卷。
- 每条进度反馈携带本责任问卷的准确路径；摘要预算不得裁掉路径，不让作者读
  Runtime、猜目录或运行领取命令找位置。缺少问卷时报告系统生成缺口。
- 问题与下一 stage 的缺项共用原 graph/阶段要求。反馈明确指出哪道题、哪个工序
  或哪项交付还缺什么；原检查、审查和人闸按图执行，Hook 不另加验收工序。
- 取消 Hook 身份、工具权限与 scope 拦截。普通工具调用不跑完整业务链；相关结果
  变化时更新必要事实并提醒，问卷保存时走原消费者及阶段求值。
- 复用原 Writer/CAS、幂等与恢复；重复保存不重复推进。本人责任与整个阶段出口
  分别判断，长检查和合并由原责任执行。

首个监听/问卷切片已按此实施并通过 r59 六包定向回归；native/workspace S6 保存与
Stop 准确路径已接线。后续 r62 验证删除保存入口重复工具策略及 S9 原接收链；
after-r18 的真实 Claude 核查设计、填卷保存后经原消费者/PTR-PLAN-01 从 S2 推进 S3，
自然结束，无手工 issue/consume。真实完整链仍为 4/9，S6 Main 集成未闭环，
全部领域和宿主迁移尚未完成。
以下原发行、权限守卫和接收机制保留用于定位旧实现；与本节
冲突的拦截在迁移时删除，不继续复制。§13.12 的独立 Impact 协议与 §13.14 的
Bootstrap 归档协议停止实施，隔离代码不合入。

### 当前兼容实现：自动准备、准确路径与原消费者保存（2026-10-15）

显式 packet 模式下，Main 的生命周期提醒会通过原幂等 issuer 自动准备当前问卷；
包括 Main 的 PreToolUse 提醒，不要求 Agent 手工运行 `runtime stage-answer issue`。
提醒从当前 Runtime 选择器和原注册 Worker execution/source reader 取得当前卷，
直接附上发行记录中的实际 view 与 answer 路径，并将其置于长背景说明之前；摘要预算不裁掉
该交付说明。S9 RepairResult 的路径以原 registered Worker execution 为根，不能只拼 Main
root；Main-owned planning packet 则用其当前发行的 answer path。Worker 按已公开的题目与要求填写专业答案、
产物路径或具体阻塞，并保存到提醒给出的 answer 路径；缺卷或 source 不可用时明确
报告准备/恢复缺口，不让作者查 Runtime、猜路径或找内部 Schema。旧格式与未开启
packet 的实例保留原路径。

PreToolUse 只监听工具调用、补充当前提醒并提示 graph/阶段流转，不核验调用者身份、
不按 answer path 授权工具，也不因问卷路径拒绝普通工具调用。保存仍进入原 typed/domain
consumer：由原 packet 与绑定来源复验答卷和证据，通过原 Writer/CAS 记录接收；提示本身
不登记 evidence、不批准专业结论，也不代表阶段已推进。

在显式 packet 模式中，S2/S3/S4 的 `planning_design`、`planning_contract`、
`planning_task` evidence 必须由当前已发行 answer 的 typed consumer 接收。公开命令
`runtime evidence add` 只有 evidence artifact 的路径和登记元数据，没有原 packet answer
与原 consumer 的接收事实；它不得旁路登记上述 kind，否则会留下没有 typed submission 的
半闭包，后续阶段无法按原接收事实恢复。Hook/原 service 自动准备问卷并给出准确 answer
路径；若准备失败，应报告具体恢复缺口，不要求作者运行 issue/consume 命令或从磁盘文件、
自由 evidence payload、历史 PTR 推断 typed answer/补造 submission。直接调用原领域
service/Writer 的 typed Save 保持原契约；packet 未启用的 legacy 实例继续沿用
`runtime evidence add` 原行为。

packet 模式下，普通 Controller cycle 在 S2/S3/S4 的原 gate 满足并准备自动推进时，
也必须从本次 pinned Runtime snapshot 中选出当前阶段、锁定 REQ、generation、stage-entry、
Main owner 与 exact typed packet 唯一匹配的 accepted `pending` submission followup，
并将该原 followup ID 交给既有 transition Apply。缺少或出现多个匹配项时不得以空 ID
推进；阶段 gate 满足本身不替代 typed consumer 的接收事实。Apply 在原 Writer 锁内复验
pending/source/original-entry，并把 cursor 转移和该 submission 的 `advanced` /
`TransitionEventID` 同 CAS 提交。显式 followup 仍按原入口处理，非 S2/S3/S4 与 packet-off
保持原行为。已经发生的历史空 ID 转移不得依据当前 cursor、旧 receipt 或后续 recover
追认成 `advanced`；旧 followup 仍按原 replaced/invalidated 规则处理。

## 2. 文件

Main 汇总目录是 `.claude/forms/`。文件名只用于定位，不标注待填或已完成；汇总由 Writer 生成，是原领域记录的可重建视图。过门和重做不改汇总文件名，已接收历史记录不被覆盖。

```text
.claude/forms/REQ-<req-id>-S<nn>-<slug>.json
```

`S<nn>` 用两位，使 `S05` 排在 `S10` 前面。`ls .claude/forms` 时，同一需求按阶段号排在一起。

| 阶段 | 文件 |
|:--|:--|
| S00 需求 | `REQ-006-S00-requirement.json` |
| S01 绑定 | `REQ-006-S01-bind.json` |
| S02 设计 | `REQ-006-S02-design.json` |
| S03 合同 | `REQ-006-S03-contracts.json` |
| S04 任务 | `REQ-006-S04-tasks.json` |
| S05 文档审查 | `REQ-006-S05-document-verification.json` |
| S06 构建 | `REQ-006-S06-build.json` |
| S07 验证 | `REQ-006-S07-verification.json` |
| S08 调查 | `REQ-006-S08-investigation.json` |
| S09 修复 | `REQ-006-S09-repair.json` |
| S10 验收 | `REQ-006-S10-acceptance.json` |
| S11 发布 | `REQ-006-S11-release.json` |

同一阶段若有多名审查者或多个任务，它们是汇总里的不同 slot。每名执行者填写独占分卷，不能同时手写这份汇总文件；分卷是输入，不是第二 Runtime 或第二权威。

## 3. 发卷时写什么

进入阶段时，生成器先从当前责任和已验证上游枚举题目，将机器已登记事实自动带入。它不能推断尚未作出的专业 PASS，也不能由文件存在推断人工决定。

无新增判断：自动记录已有状态和来源，按原门求值。S00/S01/S11 只有在显式人类决定已经到达后才可自动留证和走固定路由，不属于“文件齐全即可自动通过”的阶段。S02–S04 的确定性登记可自动接住；S09 的计划、修复结果和独立复验仍有先后与责任要求。

不能只凭状态通过：写入空位，并提示：

```text
你的填写卷在当前 execution-root 的 .claude/submissions/stage-answer.json
工作做完后，把答案填进这个文件。
Main 的阶段汇总由系统生成，你不用编辑它。
```

S05 汇总有两块，系统各发一份独占填写文件。每份卷同时携带阅读目的、问题、选项和条件示例，执行者无需到另一份长协议里找字段含义：

```json
{
  "title": "S05 文档审查 · 规格一致性",
  "instruction": "只修改 answer。null 表示未回答；完整保存即交卷。",
  "read": [
    { "path": "docs/requirements/REQ-006.md", "purpose": "核对需求承诺。" },
    { "path": "docs/dev/contracts/CONTRACTS-006.md", "purpose": "核对协议与分端条款。" },
    { "path": "docs/dev/tasks/index-REQ-006.md", "purpose": "核对任务覆盖。" }
  ],
  "questions": {
    "conclusion": {
      "question": "需求、合同、任务是否一致？",
      "choices": ["pass", "fix_required", "req_change_required"]
    },
    "findings": {
      "question": "未通过时记录具体发现；通过填 []。",
      "example": {
        "path": "docs/dev/contracts/SYNC-006.md#permission",
        "expected": "无权限时拒绝操作。",
        "observed": "条款没有定义拒绝结果。"
      }
    }
  },
  "answer": { "conclusion": null, "findings": [] }
}
```

S06 按批准 TASK，S07 按已批准 Claim，S08 按 Case，S09 按 Contract assertion，S10 按审查全集展开题目。覆盖 ID、内部事件和聚合 PASS 不是执行者的题目。关键阅读与填写说明随卷保存；next 只指当前责任和唯一填写路径。

`pass` 用空数组表示没有发现，不生成一条全空 finding，也不与未处置发现共存。`fix_required` 或 `req_change_required` 若没有可定位的 path、expected 与 observed，就没有回答完整。null 保留为草稿；字段错误给具体 JSON Pointer、允许值和例子，而不是让 Agent 重新找规范。

## 4. 过门

主要触发是 authoring 文件成功保存后的 PostToolUse；状态事实或人类决定到达也可触发对应 adapter。不能等待下一次无关工具调用才消费答案。事件接线见 §8。

Hook 接着做三步：

1. 识别当前卷和真实 owner，校验 authoring answer，复核发卷时冻结的责任与输入。路径按原 file_sources/Git tree 契约读取，读不到定位到答案字段；Agent 自填哈希、身份和事件不构成权威。
2. 具名 adapter 调用既有领域事务，由 Writer 生成 evidence、身份、代数、哈希和路由。相同保存幂等重放；修订生成新记录并明确 supersedes/失效关系，不抹除历史。基线漂移不能把新哈希盖到旧 PASS 上。
3. 立即求原门，满足则执行原合法自动迁移并生成下一卷；有发现走原失败路由。登记已提交而后续求值/生成失败时，持久化系统待办并自动恢复，不能要求 Agent 再造一份登记。

填完空位本身不构成通过。双人结论都在、路径与当前登记一致、不通过时有可定位的 finding、需求变更停在人闸、实现差异落在任务范围内，这些检查仍由原来的门负责。

## 5. 模板注册与字段闭合

模板随匹配版本的 Harness 发布，在既有 Definition/catalog 中登记 Stage、责任、authoring Schema、题目/选项/条件/示例和具名领域 adapter。复用原责任分母与输入来源，不另建工作流引擎或长期 registry。

JSON Schema 只约束形状，不能自动提供好问题或证明语义质量。注册校验须证明：题目选项和条件与 Schema 相符；每个必填答案有 adapter 消费者；原门的每个手续字段有机器生产者。authoring Schema 允许草稿 null，持久 evidence Schema 由 Writer 生成，二者不混用。

生成只用具名受限展开器，例如 TASK Closing Contract、Assignment Claim、Case 当前假设；不执行模板中的任意脚本、shell 或通用路由表达式。上游缺失时给具体缺口，不能把无资料的占位计划当作可填写终稿。

若卷已答齐而剩余 gate 无法映射到题目、真实交付条件、人闸或系统待办，报 FORM_GATE_UNMAPPED 给维护者。这是 adapter 漏接，不是执行者未完成。

## 6. 并行、身份与旧卷

Main 汇总只由 Writer 写。每个执行者收到独占 authoring 路径；worktree 中复用 `.claude/submissions/stage-answer.json`。该文件包含本责任的题目和 answer，写入只更新本人的 slot，Writer 锁内串行合并。两个过期整卷不能互相覆盖。

一个事实源、多份独占输入的区别必须保留：分卷不带可 merge 的 Runtime/journal；完成事实仍住在原控制面。执行者只见一个填写入口，Main 只见一个阶段汇总。

owner 来自可信派发/execution/session/Assignment 绑定，不从 answer.agent_id、文件名或“唯一等待 Agent”猜。身份无法识别时明确报告派发接线问题，不伪造 producer；汇总者也不能替独立 reviewer 签字。

卷的机器绑定包括 Runtime identity、REQ generation、阶段重入/轮次、Assignment 尝试、模板版本和冻结输入闭包。路径稳定不表示授权永久相同；旧执行、旧轮和旧 Runtime 的保存不能自动绑定当前上下文。

## 7. 接收、求门与恢复

哈希等字段由机器从权威输入生成，但只能绑定实际被审版本。Formal subject 从声明的已解析 Git tree 读取；产品/锁定规格漂移拒收旧结论，合法 mutable evidence 仍按原白名单同步。全仓 HEAD 移动不必然使卷失效，比较该卷冻结的实际内容闭包。

首版复用既有领域 API 的原子事务，不能在持锁 Apply 中递归调用另一个 Writer。跨领域步骤如尚不能一次提交，应保存 durable 消费检查点，由原恢复路径继续求值；不得提前承诺所有步骤同属一个原子事务。

相同卷实例、slot 和答案内容重复投递不追加证据或迁移。回执引用原接收记录，同时报告当前阶段；不能用旧回执冒充当前通过。修改已接收答案保留历史，阶段已经推进时消费原返工流程，不就地重写冻结依据。

登记成功、下一卷生成失败时从已提交 Runtime 重建投影。系统正在恢复不是执行者缺题。每次事件最多执行 Definition 允许的一次自动阶段迁移，Hook 不跑长测试、合并或无限循环。

## 8. 保存事件和无 Hook 环境

接手前 settings 的 PostToolUse matcher 是 SendMessage/Agent/Task/SubagentHandback/Bash，不含 Write/Edit，`internal/cli/run.go` 的 PostToolUse 提前进入观察分支。当前已增加 Write/Edit/MultiEdit，并在 observer 前接入已开启 v1.2 的具名答卷 consumer 和官方 additionalContext 回执；native/workspace S6 保存与 Stop 准确路径已接线，r59 六包定向回归通过。完整阶段仍为 4/9，S6 Main 集成、完整宿主安装与真实 Claude 验收仍待完成。按 [Hook 接线](L4-hook-platform-wiring.md)治理后续宿主验收。

主触发覆盖实际支持的 Write/Edit/MultiEdit 与能识别目标文件的写入事件，只消费已登记当前卷。PreToolUse 只监听并补充提醒，不核验身份、不授权特定文件、不拒绝普通工具调用。PostToolUse 观察成功写入并调用原答卷 consumer；consumer 拒收只表示未形成领域完成事实，保留草稿并给修正回执，Hook 不撤销过去的写入。

SessionStart/PreCompact/PreToolUse 生命周期提醒为 Main 自动准备当前问卷并附精确填写路径；它们不替代保存后的领域接收。子代理返回由 Main 接收保全结果。Stop 按持久接收事实判断本人是否完成，并提供实际问卷路径；已完成者不因系统 pending 或别人未交卷反复唤醒。当前 native/workspace S6 的保存与 Stop 准确路径已接线，S6 Main 集成仍未闭环。

短事务超时或长集成由已有系统待办/Integrator 接续，回执点名 owner。无 Hook 的 Codex/终端采用同一 consumer API 的受控保存入口或宿主工具返回边界显式消费；只有文件落盘的环境标为未自动接线。兼容命令由宿主给完整实际参数，不让 Agent 猜事件。

## 9. 不能在最终交卷时追认的先决条件

S00/S01/S11 的需求锁定、绑定启动和发布处置仍消费显式人工决定。文件存在、状态字样、保存、超时或 Agent 写 approve 都不能生成授权；人工决定到达后自动留证并走固定去向。

S6/S9 写前计划、S8 RepairContract 批准、S9 独立复验必须守原先后与分工。系统在同一责任路径只暴露当前题目：先填计划，接收后继续工作，再填结果。不能拿最终 completion 追认写前权限；高风险保留原计划人闸。

产品提交和 Main 集成检查是真实交付条件，不能靠 JSON 的 completed 替代。Worker 已报告而 Main 待集成时，回执说明 Worker 责任结束，下一动作归 Main/Integrator。

## 10. 各阶段留下的工作

| Stage | 执行者回答/产出 | 系统派生与消费 |
| --- | --- | --- |
| S00/S01 | 需求整理、人确认及显式启动/分支目标 | 待确认稿、锁定记录、绑定与回执；不猜授权 |
| S02 | 本层设计决定、产物路径 | 来源校验、planning evidence 和具体缺口 |
| S03 | 共享模型/SYNC/分端责任与适用性 | 输入闭包、contract evidence、原覆盖检查 |
| S04 | TASK 范围/断言/依赖和计划 | 列表、条款/DAG 检查与规划登记 |
| S05 | 独立结论与具体发现 | producer、精确 subject、信封、固定路由 |
| S06 | 当前计划、实现/检查和结果 | Assignment 生命周期、真实 diff、Main 集成 |
| S07 | 已列 Claim 结论/观察、Finding/受阻理由 | digest、Result、覆盖、CleanRound/Batch |
| S08 | 当前 Case 假设/证据/因果和修复意图 | 有序 register/consume/route、版本与批准绑定 |
| S09 | 计划、修复结果、独立断言结果 | Session、changeset、impact、复验及新 S7 handoff |
| S10 | 逐项验收/反证/审计判断、风险 | 从原权威枚举全集、metrics、报告和 evidence |
| S11 | 人选择六种处置之一 | 当前 package、一次性决定和固定路由；不发布 |

覆盖规划和风险判定仍由专业角色承担；生成器枚举现有权威输入，不替规划者判断未知业务或自动补 PASS。首版具体实现每个 adapter，不把这张表冒充所有领域已接线。

## 11. 回执明确谁已完成

回执至少给 work_complete、当前 Stage、接收状态、待完成责任/owner 和一个 next；错误给 answer JSON Pointer、观察值、允许值/示例和修正动作。

```text
S05 · 你的规格审查已接收，工作已完成。
本阶段等待：任务可执行性审查（reviewer-task）。
你无需继续登记或推进；Main 等另一份答卷到达。
```

```text
S05 · /answer/findings/0/observed 仍为空。
你选择了 fix_required，需要写出在所引路径上实际看到的矛盾。
只补这一格并保存；其他已接收责任保持有效。
```

题已齐、系统待恢复时，明确“答卷已接收；系统恢复推进”，不再给执行者补填指令。未知 gate 缺口触发 §5 的维护缺陷，不能笼统反复说“继续 DRIVE”。

## 12. 落地与判据

先打通 S05 完整链：模板注册 → 真实独立派发 → 生成分卷 → 保存事件 → 原 document_review evidence → 原 PASS/FIX/REQ-change gate → 下一卷/人闸 → 回执。再接 S03/S04 的确定性登记，随后覆盖其余领域；S01/S11 不作为文件存在型捷径。

上线须证明：

- 无模板上下文的执行者，只读发卷说明和工程输入即可首次正确回答，不搜索内部命令或 Schema。
- 最后一次完整保存立即接收/求门，不依赖新 prompt、随机 Bash、重复 ready；实测保存到接收/推进延迟。
- 未答、条件缺失、非法值和缺引用均定位到当前卷字段；无消费者的题不进入必填项。
- 并行输入不丢答案、不冒名；重复事件、旧轮、阶段重入、Runtime 切换和基线漂移不污染当前通过。
- 接收后中断、响应丢失和下一卷生成失败可恢复；原 Writer marker/journal 恢复测试继续通过。
- 专业失败走原路由；计划屏障、独立性、真实集成、全量复验与人工授权仍有效。
- 已完成者不会因系统待办或别人未交卷反复干活；unmapped gate 给维护者明确恢复路径。
- 真实 Claude 保存→Hook→领域门→additionalContext 的验收，与纯函数原型、CI/交叉编译分别留证。

实战观察首次交卷拒收原因、查找协议/命令次数、已完成者收到重复指令次数、保存到推进延迟。优化题目、adapter 或接线，不能靠放宽质量门改善数字。

## 13. 生产实施契约（2026-10-13）

本节固定首个S5实施单元，并细化§4/6/7的事务与发行边界；全局优化仍是独立文档，具体机制契约归本层。实现状态以代码/测试及全局设计实施记录为准。

### 13.1 版本与权威对象

Runtime 1.2新增可选configuration.optimization.packet开关、entities.work_packets与entities.submissions。1.1完整封闭Schema保留，逻辑Schema按实例版本固定分派；旧形状不能写新字段。新安装提供1.2，旧Runtime不得暗迁移。显式启用前只读shadow。

workpacket是纯投影包；绑定runtime/generation、lifecycle state/phase/phase_revision（作为单调阶段进入身份）、真实owner、Assignment/manifest内容指纹、责任slot、attempt、template/version及冻结subject refs。全局revision与updated_at不是发包依据。发行与消费索引由原Writer登记，磁盘工作包/回执/汇总可重建。

共享Main checkout答案为.claude/submissions/<runtime>/g<generation>/<responsibility>/<owner>/<attempt>/answer.json；Worker独立目录为.claude/submissions/stage-answer.json。路径从发行记录生成且每attempt独占，不让作者选择。机器题干不可改，answer只含conclusion/findings；authoring允许null草稿，完整与非法答案由原Schema及条件判定，不增加ready题。

### 13.1.1 新安装与重置的版本契约

新/空目录的 install、init 和原 Hook 自动初始化采用 Runtime 1.2.0；packet 默认 false，只有新安装/init 的显式 `--packet` 才开启 S5 consumer。实际 Claude 验收前不将该 consumer 推为默认。开关仅选择机读接收形式，不生成 REQ 授权、计划、专业结论或人闸决定；后续 REQ bind 保持其版本/配置。未绑定的可证明无业务进展 S0 只观察、不解析尚不存在的 S5 权威、不发卷或授写；不可用的 bound 权威仍拒绝。

init 不覆盖已有 state/journal/marker。初始化中断时以原 pending fresh_state 的完整字节完成；重试未指定开关时保留原模式，显式开关与 marker 冲突拒绝，不把旧 1.1 marker 升级。发布前在匹配新二进制上执行完整 Schema 和语义校验，再写第一个文件。初始化同组发布只读 runtime-initialization.json 回执，记录初始 Schema/optimization 及初始 state 摘要；它只防止恢复猜错格式，不取代当前 Runtime 权威。

rollover/unbind 在原人闸与 Writer 锁内根据实际源 Runtime 保持 schema_version 及完整 optimization 配置；历史 packet/submission/执行目录留在原成对归档，新 inactive 的索引为空。不把旧 1.1 升为 1.2，也不把新实例降级或悄悄关闭 consumer。pending reset 恢复核对归档源与 target 的格式/配置一致，已批准的旧 1.1 marker 保持原样。

安装仍只对新/空目标，release manifest 与 installed receipt 覆盖真实二进制、Hook、角色/Skill、模板和初始化状态。旧写入器须在实际写入 1.2 前拒绝，Schema 校验须失败，不能抹掉新字段；旧程序的观察/no-op 成功不能当作新格式兼容证明。活跃迁移入口尚未实现，存量 1.1 普通写入/原 pending 恢复保留旧形状。

RR-A 的文档导入不能重建已发行 packet、规范接收索引、实际 owner/激活或 followup 结果。本切片保留 legacy 源的 1.1 seed；可证明尚无发行/接收索引的 1.2 源保留其格式/配置。已有索引、混合源版本或新格式源不可读时拒绝文档重建，保留原配对、pending 和原恢复证据，直到实现可证明的索引恢复。该限制不阻止原 Writer 对有效 1.2 pending 事务做原样恢复，不以手植新索引宣称支持。

### 13.1.2 S5 formal inputs 与共享 typed packet 路由

packet-enabled S5 的冻结输入是当前代 registered/locked 文档，加上唯一被原路由接受的当前 S4 输出闭包。仅当 `s4-task-planning-v1` 绑定当前 Runtime、代数、锁定 REQ 和 Main 的 S4 进入身份，且同一 packet 有唯一 `advanced` submission、原接收 evidence 与 TR-002 journal 记录时，才可投影 S4 来源。沿原 S4 checkpoint 与完整 journal 验证该 tuple，再用已冻结 S4 packet 和原 S3 解析器重建 source closure；不从文件系统扫描或猜上游材料。输入包含原 `Inputs.Sources`（稳定 ID 从 path 派生；缺失版本显式记为 `unversioned`）及答卷点名并经原 S4 校验的 complete TASK 和 dispatch plan。输出还必须与当前代 Runtime document 的 kind/id/path/version/SHA/status 精确一致；取消、未接收、旧代/旧 REQ、摘要漂移或额外 TASK/plan 均拒绝。没有 typed S4 packet 的旧实例保持原 registered/locked docs 投影。S4 来源中的 `TASK-template.md` 可作为冻结读集输入，但任务分析必须按文件名排除模板，不把它计为 TASK 实例。

发行、消费授权和 Worker 冻结都重算相同闭包，并按冻结 SHA 校验原文件视图与 checkout。S5 gate 的 `subject_refs` 只可通过同一 accepted-S4 证明扩展只读 document facts；不写入 `Runtime.documents`，也不放宽通用 subjects matcher。

共享 `entities.work_packets` registry 的消费者按 `template_id` 路由到各自 typed reader。S5 遍历只将 `s5-document-review-v1` 交给 `FindWorkPacket`；明确识别并略过合法的 `s2-design-v1`、`s3-contract-v1`、`s4-task-planning-v1`、`s7-review-result-v1`、`s8-hypothesis-result-v1`、`s9-plan-report-v1` 与 `s9-repair-result-v1`。缺失、未知或畸形 subtype 必须拒绝，不能当成无关行略过；S5 的 `FindWorkPacket` 仍拒绝错误子型。

### 13.2 实际接线和权限

register-workgroup成功及真实SubagentStart/SessionStart按当前合法责任发行/恢复；没有真实身份不给签名题。发卷不取代原readback/activation/计划屏障。Main与Worker的Write/Edit/MultiEdit PostToolUse在最后成功保存消费，不依赖下次调用。Main普通工具仍走纯observer。

发行器从当前 Runtime 的 team/agent、team.manifest_ref、prompt_ref 和 manifest 的精确 Assignment 行推导 owner/slot，冻结上述当前 formal input closure；不能用作者指定的 owner、输入子集或磁盘草稿替代这些事实。先经原 Writer 登记 packet，再发布只读 packet view 和独占 draft；重放只补缺失投影，存在的答卷保持原字节。view 损坏须报错，不能覆盖已填写答卷。实际 Main 路径含 a<attempt>-<packet_id>，因此同代重入也不会复用旧文件。独立 Worker 必须先具备受登记的 S5 execution 与相同冻结输入绑定；该入口未落地前不得借共享 Main 答卷宣称独立目录已支持。

冻结集必须精确包含当前已锁定 bound_req 的 id/path/version/sha256；仅 documents[] 自洽而与已授权 REQ 基线不同也不能发卷或新接收。该检查不修改基线、授权或原 TR-004 的 req_baseline_unchanged 人闸边界。

开启 packet 的 v1.2 S5 workgroup 登记在原 team 行写 packet_stage_entry 与 packet_generation，两字段由原 Writer 的实际 candidate 派生；发卷与接收均要求同进入身份/代数，防止重入时把旧 working reviewer 当作新派发。manifest 的 runtime_id/baseline_generation/workgroup_id 必须匹配 Runtime、当前基线和该 team。无此登记 stamp 的旧派发不能自动补授权，只能重新完成原注册；v1.1 team 形状保持原样。

独立 S5 Worker 复用原 workspace.prepare/launch 与 execution registry。仅已开启 v1.2、当前真实 S5 Assignment 可准备 purpose=document_review 的 execution；原 execution 行固定 work_packet_id、stage_entry、execution_generation、owner、独立根和冻结 inputs，产品 write_paths 与 executable required_checks 均为空。审查可读冻结资料和原观察 adapter，交付是本人的专业答卷；不发送 Builder 的 commit/report/deliver，也不因没有产品测试而伪造测试 PASS。原计划/激活使用已有 workspace.begin，plan_approval_required 的原人工激活仍由 Main 执行。

packet 本身保持不可变；Worker bootstrap 安装 `.claude/inputs/work-packet.json` 的 execution view，内含原 packet、固定执行根/代数/owner、独占 authoring_path 和 frozen_input_root。该 view 与所有冻结资料属于 bootstrap receipt 的只读资产；答卷 `.claude/submissions/stage-answer.json` 是可写草稿，不进入 bootstrap hash。read[].path 是权威资料的相对路径，在 Worker 中从 frozen_input_root 下读取，题干不改。bootstrap 重放仅补缺失草稿，不覆盖已填写答案或 ready 资产。

S5 execution 的 preparing 记录提交后，原共享 Main 答卷不再是可消费或可写入口；不能因独立目录未 ready 而退回 Main。消费必须从当前 registry 的确切 Worker 根、owner、execution generation 和 packet 解析，并验证 view/bootstrap/冻结资料；原 Writer 锁内再次检查该源绑定与原计划链。新接收还须当前 S5 进入身份；已接收重放可返回原交付与当前阶段，但仍不能复用被替换的 execution。替换保留旧目录且生成新 execution generation 和空草稿；原接收不可覆盖，已接收后走领域恢复或新派发。S5 Worker 的 Stop 只判断该原责任：本人已接收、或原进入身份已被原路由取代可结束；未提交且当前责任仍有效则在任何 Controller 变更前阻止。Main Stop 的既有 preflight 顺序保持。

领取允许 spawned/reading 等尚未开工的合法派发状态；首次接收则要求原 working 状态，并在原领域 Writer 锁内调用原计划/激活校验器，核对实际 Runtime/Assignment/readback/activation 链。此校验是内部 adapter 注入的可信能力，不能来自 Hook payload、authoring JSON 或 CLI 参数；缺失能力拒收。领卷或保存答案都不生成计划或激活事实。

PreToolUse仅对当前真实actor已发行的确切答案文件提供控制面输入例外；通过规范化根、物理包含与无下级symlink检查，不放开整个submissions或.claude。工作包/Runtime/汇总不能借此变可写。S7 reviewer与S8/S10 phase规则同步接例外，其他owner/旧attempt/未发行文件和指向仓库内另一owner的symlink拒绝；消费端再核对binding，PostToolUse拒收不撤销保存、不deny。

### 13.3 一次接收与具名修订

submission_identity=(runtime,generation,packet,slot,owner,attempt)，其key不含答案摘要；content hash单独记录。原RecordEvidence的Writer事务内原子追加document_review与submission索引；同身份同内容重放原记录，同身份不同内容content_conflict。冻资料漂移、跨owner、旧stage_entry及非法结论不登记完成。原producer独立性及门检查不变。

接收请求必须携带本次成功保存后冻结的完整authoring字节。原Writer事务内，先按注册Schema及发行题干验证答卷，再将完整answer规范编码为JSON并自行计算content hash；调用方提供的hash仅可作为待核对值，不能成为接收事实。null、条件题不完整、未知字段、重复JSON键及题干修改均不占用attempt。已接收答案以规范answer内容判断幂等，空白、缩进与对象字段顺序不构成新内容；finding数组的顺序和具体字符串保持原判断。

document_review必须是该已验证answer和发行packet的确定性投影：conclusion/findings完整一致，subject_refs精确覆盖冻结的path/version/sha256集合，无重复、遗漏或额外对象；runtime/generation/producer/responsibility/evidence_id与真实接收请求一致。仅fix_required派生document_fix_required，pass和req_change_required不得携带自动事件；S5不填写review_round。接收入口拒绝不匹配的信封，不以稍后gate拒绝替代接收校验。相同内容重放还须匹配原登记信封的路径、内容指纹与domain引用，不能在完成后换信封或借旧回执改写结论。

application adapter 在消费前按已发行路径读取一次有大小上限的完整答卷，拒绝 root 下任何 symlink；保存工具的真实 file_path 只用于定位发行记录，不承载身份。具名领域信封路径包含 packet_id 与规范 answer 摘要，采用完整文件的独占发布；同内容并发保留先发布的 created_at，不写半文件、不覆盖其他判断。已登记后从原 evidence path/sha 重建信封引用；不重新生成时间戳。孤立但未登记信封不构成完成事实，重放仍走原 Writer 校验。接收事务完成后才调用 Controller，不能嵌套 Writer。

首次接收 attempt 不可变。S5 的 RecordDocumentReviewRevision 只由 Main 调用，以真实 previous submission_key 和具体原因发行 attempt+1；固定原 runtime/generation/stage_entry、完整冻结资料、owner、assignment 和 responsibility。原因非空且不超过 1024 字节。未接收、已被修订、已离开 S5 或冻结链变化不能产生新分支。同 previous、同原因重放唯一 successor；不同原因冲突。重放历史发行不产生新工作授权。

发行事务在原 Writer 内同时追加新 packet、记录旧 submission 的 revision_packet_id、将旧 evidence 标 invalid（修订等待），并将旧 pending followup 标 replaced。暂停旧证据必须发生在发行时：修订未收到期间，旧 PASS 不能与同伴 PASS 合并推进原门。新接收事务登记新 evidence/submission，并记录旧 submission 的 superseded_by 和最终失效引用；历史 artifact 不改。既有原门仍只消费 valid evidence，不添加宽松通道。若原迁移先获锁并离开 S5，发行失败；若发行先获锁，旧控制接续锁内复查后不能迁移。Controller 将真实 S5 gate 求值引用的完整 evidence 集带给 Apply；原 required_evidence 的单槽映射不能丢失第二位 reviewer。Apply 的原 Writer 内复查所有这些引用的 valid/代数/摘要；TR-003 另确认当前修订 tip 已收到有效 PASS，未收到或非 PASS 不可经手工迁移绕过。

已发行当前 S5 packet 的责任只接受答卷 consumer；手工 evidence add 不能登记并行替代结论。该检查在原 Writer 内重复，保留未发行答卷的旧模式。正式启用与旧实例恢复必须核对未索引的历史有效审查，不能暗迁移其资格。

新 binding 的 previous_submission_key/revision_reason 是 Writer 派生的固定题外事实，不能从答案输入。旧回执保持 accepted 的历史事实，但 responsibility_status=replaced、work_complete=false，并指出当前答卷；Stop/Idle 只认当前 attempt。旧答卷保存不重新接收或控制，历史回执允许只读查询。

独立 Worker 修订先持原 Git 调度锁与旧 launch lease；仍运行的 Worker 不能退役。发行同事务将旧 execution 留入 retired history，移除活跃映射，新 packet 固定 execution_required。发行后不回退 Main authoring；由原 prepare/bootstrap 创建更高 execution generation 的新目录、新冻结 view 和空白答卷，旧目录与答卷保留。相同责任和冻结权限不重新铸造批准或专业 PASS。资料/责任变化或取消修订沿原返工/人闸，不自动复活旧证据。

### 13.4 至多一次控制接续

domain事务记录stable followup_event_id、origin state/phase/phase_revision和control_pending。ControlRequest/transition.Request携带这组事实，Apply的原Writer事务内核对未处理与origin相同，并把转换及followup结果一起提交。已有request_id或按transition/revision的key不能代替这项去重。

转换已提交但回执/外层响应丢失：恢复只重建receipt与当前包；cursor已被其他事件推进：标replaced，不能把旧交付应用新阶段。not_ready记evaluated_no_transition，未来新交付有自己的事件；unknown只对原origin有界重试。旧Runtime/重入/同答案重复保存不重复迁移。长测试/合并归Main/Integrator，Hook仅短事务。

生产索引的control_status固定为pending/no_transition/advanced/replaced，其中no_transition表示evaluated_no_transition。advanced在转换事务内同时写transition_event_id；Controller进入求门前检查真实submission索引和原进入身份，已处理只返回当前事实，原进入身份改变则由短Writer事务登记replaced。转换Apply在锁内重复检查pending与原进入身份，转换后的stamp与cursor原子提交；重复或并发请求不生成第二转换/计数。没有合法转换的已完成求值沿短Writer事务登记no_transition；unknown保留pending供有界恢复，不将错误记为已求值完成。

### 13.5 完成回执与验收

回执包括accepted、responsibility_status、work_complete、current_stage、domain_refs、waiting_on的reason/owner/ref、当前recipient next_action和JSON Pointer字段错误。work_complete只表达本人交付已接收，可包含合法Finding；阶段PASS、Main集成和人授权另判。系统pending与同事未交不重复催本人。

跨领域 wire 的封闭枚举、原 `submission_key` 格式联合、S9 candidate/final 状态与
consumer 字段约束，按 [全局 §4.4.1](../GLOBAL-OPTIMIZATION-DESIGN.md#shared-submission-receipt-contract)
固定。共享 Schema 必须直接验证实际 Hook / 公共命令的完整原始 receipt JSON；
不能将 S5 专用枚举误当成所有已接线领域的发布接口。当前补齐工作另记新证据，
不改已冻结安装包与机制快照。

receipt 是 Runtime 索引与原 evidence 的可重建只读投影，不是第二状态源；acceptance/错误回执须报告本次保存的接收结果，不能把同身份异内容的拒收伪装为本次成功。领域接收成功而求门超时/unknown 时保留 accepted/work_complete，waiting_on 点名 Main 的系统接续。未提交责任点名已注册的真实 owner，未派发责任归 Main；req_change_required 点名人工需求决定。回执丢失可从 durable reception 重建，不能要求执行者改答案或再造 completion 消息。

原返工迁移消费 FIX 并将该 evidence 标 invalid，不撤销原交付已接收的事实；work_complete 与 evidence 的当前 gate 资格分开。同代历史签字可说明本人已交付，不能授权下一阶段或新责任；代数变化、当前 S5 重入或新 attempt 必须匹配新派发，旧完成不能替代新交付。其他责任的派发投影暂不可读时由 Main 恢复，不把已收到的本人交付退回未完成。

必须用真实Writer/领域门证明：并发两个owner不丢答案；同attempt异内容冲突；domain提交后/转换提交后故障恢复；两PASS/FIX/REQ-change原路由；原权限、计划、独立职责全集、人闸及冻结边界。安装和官方wire测试后再独立审查；真实Claude实测另记版本与结果，不用fixture替代。

### 13.6 已发行索引的完整恢复（v0.7.6 实施契约）

RR-A 文档重建不恢复 packet/submission/Agent 激活。已发行工作的恢复来源必须是原 Writer 在同一把 Runtime 锁内验证并保存的完整 state/journal pair；journal 包含旋转段，不能仅取当前尾文件。1.2 有发行/接收索引后，每次原提交和指纹刷新在清除 pending marker 前发布完整 checkpoint。发布失败保留原 pending，重试补发布，不重新接收或迁移。checkpoint 是恢复证据，不供正常门禁读取，也不是第二 Runtime。

checkpoint 含精确字节摘要、Runtime/代数/revision/cursor；完整保存派发、激活、execution、修订 lineage、失效状态、唯一接收键和已处理接续事件。最新索引最多保留两份完整 pair；已审核 plan 自带冻结候选，不依赖旧 checkpoint 永久存在。单份 state 上限16 MiB、完整 journal 上限64 MiB，超限保留 pending 并明确报告，不静默截断。先完整落盘、再发布索引、再回收旧副本；未完成目录不能成为恢复来源。

Main 用 `runtime recover plan --checkpoint <manifest>` 明确选择来源，沿原 `apply --approved-by` 人闸和隔离/marker/原子替换流程。plan 不 replay 原已处理的 followup。只接受最新 checkpoint；当前可读 Runtime 或 journal 中任何更新、不同 Runtime/代数/REQ 或同版本不同事实都拒绝，已通过 bind/unbind/rollover 关闭的 Runtime 不复活。原合法 pending 总是优先。

完整 pair、封闭 Schema、语义、packet/接收/证据/修订/事件关联必须同时校验；不能通过重算 manifest 摘要补造专业事实。实际 workspace/REQ/原计划/冻结 Worker 源仍需可验证。应用前持原 Git 调度锁及相关 Worker launch lease；活跃子进程拒绝应用。恢复不能复制一个运行中的执行者，也不能补写它的答卷。已 advanced/replaced 的记录保持已处理；新 attempt 未交保持未交，旧 PASS 保持失效。domain 已接收但控制 pending 的记录由已有有界补消费接续。

验收覆盖真实发行/原激活/保存/修订、两 PASS 原迁移、响应丢失、pair 半替换、journal 旋转、索引发布失败、源漂移和 launch lease。安装和真实宿主结论分开记录。

### 13.7 S7 typed ReviewResult adapter（2026-10-14 首个 Writer receipt 切片历史契约）

> 本节保留首片设计与当时状态；当前发行、Hook 与验收进展以§13.7.1及全局设计§7.47为准。

S7 的 authoring scaffold 是输入视图，`review.Result` 是原领域 Writer 消费和持久化的
Canonical ReviewResult。S7 adapter 不复用 S6 专用 `agentcontext.ResultTemplate`，也不把
S5/S6 的通用 `stage-answer.json` 解释成 ReviewResult；不增加第二份结果表、queue 或
Writer。直接 `runtime review-result submit` 与未来保存 Hook 最终都必须调用
`review.SubmitResult`，沿原 Assignment 消费、Finding、Claim projection、CleanRound /
ObservationBatch 与 pause 路由。

**当前实现状态。** 原 `review.SubmitResult` 已在 Runtime v1.2 packet mode 接入 Writer
receipt：现有显式 CLI 仍提交原 `review.Result`；Writer 固定 `result_id`/idempotency tuple，持久化
完整 `submitted_answer` 与 `effective_answer`，并对专业答案字段计算 digest，在原消费 CAS 中写入 receipt。
重放校验原 pinned Plan、Assignment tuple、Result/Finding 文件和索引、journal commit；同键同答卷
为零写重放，同键异答卷冲突。未变 Assignment 在具名 Plan 修订后保留原 receipt；被修订影响的
Assignment 清掉当前 receipt。原 Runtime pending recovery 与引用/摘要保护的暂存清理继续适用。
这不是 typed 答卷发行或保存 Hook：受信 issuance、exact-path save Hook 和 Reviewer 专业答卷到原
Writer 的平台链仍未接通。Writer 的 Claim-owner 转移失效、20 MiB regular-file 输入预算与 pending-safe 清理已通过最终定向验证；详见全局设计§7.46。

**发行的机器绑定。** 每个 Reviewer 的唯一卷由受信的当前 Assignment 派发生成，至少固定：
Runtime 与 baseline generation、当前 ReviewPlan ref 和经原 Plan reader 验证的 digest、
review round、Assignment ID/revision、真实注册并派发的 owner、该 Assignment 的精确 Claim
集合，以及当前 frozen subjects 的 `subject_digest`。这些值只显示为机器题干，不属于
可编辑答案。模板同时签发一个精确 authoring path/issuance identity；保存 Hook 以成功
Write/Edit/MultiEdit 的真实目标路径查找发行记录，再用可信 session/owner、当前 Runtime
和 Assignment 复验绑定。路径、文件名、`producer_agent_id`、CLI `--agent-id` 或“唯一等待
者”都不能授予身份。旧 Plan、owner、generation、round、Assignment revision、Claim 集或
subject digest 变化后，旧卷必须拒收。没有精确发行记录时保持现有显式提交入口和错误路径，
不猜测 owner 或路径。

**专业答案和投影。** scaffold 让 Reviewer填写 `claim_results`、必要的
`blocked_claims`、真实观察/方法/evidence refs、Findings/encounter、deviations、
site-lost 声明及本人判断的 Result-level `verdict`。未回答字段可为草稿 null；不会预填
`pass`，也不能由全绿 checks、捕获数据、Claim 覆盖率或 Finding 是否为空自动合成 verdict。
适配器只从发行记录补机器 tuple，再把已完整填写的专业字段映射为现有 `review.Result`，
随后仍由 `review.SubmitResult` 对 schema、PLAN、精确 Claim coverage、producer 独立性、
evidence/capture provenance 和 Finding readiness 做原校验。作者填写的 check/ref/SHA 只
是答卷内容；机器 capture 的原始字节、command/window/Plan 绑定仍由现有 capture consumer
验证，不能互相替代，也不凭此新增 required-check 分母。

**同一 Writer 的摘要、回执和重放。** 原 `SubmitResult` 是唯一接受事务。它在登记 Result
evidence/index、Assignment consumed、Claim disposition、Findings 和现有 round consumer 的
同一 Runtime CAS 中记录发行身份、稳定 submission key、完整规范化 professional-answer
digest，以及可从原 Result evidence ref/SHA 与提交事件重建的 receipt；immutable artifact
bytes 继续沿原 staged-write/cleanup 边界处理。digest 覆盖全部已接受
答案字段（包括顺序敏感的 Claim/blocked dispositions、evidence refs、Findings/encounter、
deviations、site-lost 和 verdict）；machine tuple 由发行绑定单独校验。JSON 键顺序、空白
和缩进不改变 digest；重复键、未知字段、题干/绑定编辑或不完整答案不占用 key。Result ID
由首次发行或接收固定，不能在重放时重选。

同一 issuance key 与相同完整 answer digest 重放时，原 Writer 先验证并返回原 Result/receipt，
不追加 Runtime revision/event/evidence/Finding，不重复消费 Claim，也不再生成
CleanRound/ObservationBatch。相同 key、不同 digest 是 conflict；一个 Assignment 不允许借新
Result ID 或重发卷覆盖已消费判断。两个并发相同提交由原 Writer 锁/CAS 只登记一次。CAS
前可确定失败时，按 `SubmitResult` 原有 staged-artifact 清理；CAS 返回不明或外层响应丢失时，
先读原 Runtime state/journal、Result index 与 durable artifact，若提交已发生则重建同一 receipt；
若未提交才按原恢复路径清理孤立 staging 并重试。不得用第二 Writer 或另一个成功标记修补。

接收回执立即说明该 Reviewer 本人是否已交付、原 Result ref、当前 owner next action；另一
Assignment 未交或全局 CleanRound 尚未形成，不把已接收 owner 改报为 pending，也不要求其
等待全局 gate。阶段仍由原 ReviewPlan/Gate/Controller 根据所有 required Claims 和原路由
决定。此处固定目标；同 Writer receipt、digest 与精确重放基础已实现，typed scaffold、trusted
issuance 和保存 Hook 仍未实现。

**已实现 Writer receipt 切片的兼容边界。** 现有显式 CLI 使用一次性 `result_id` 作提交键，
与 Runtime/generation/Plan ID 与 revision/round/Assignment/producer/subject digest 一起
绑定，不添加 Reviewer 手填的第二键。完整 `submitted_answer` 与规范摘要保存到原
Result envelope；capture merge 后的有效答案另行闭合，重放不再消费变化的 CaptureDir。
Assignment 的 `result_receipt` 在原消费 CAS 写入并绑定原 event/index/artifact；具名
Plan revision 清当前 receipt、保留旧历史。历史未记录完整摘要的结果不补猜为可重放。
孤立 O_EXCL 文件仍不是接收事实；原 pending 先恢复，缺可靠无引用证明时保留文件
并给恢复诊断。第一切片不声称已发行 trusted-owner scaffold 或自动保存接收。

**剩余实施顺序与验收。** 原 Writer 的稳定 key、AnswerDigest、receipt 和恢复基础已接入；接下来
发行当前 Assignment 的 typed scaffold 与 exact-path identity，再接保存 Hook、将专业答案投影至
原 `review.Result` 并调用原 Writer。完整 adapter 验收应覆盖错 owner/旧 tuple/
Claim 漏项或多项、伪造 capture/check、专业 verdict 不被合成、同 key 同摘要零写重放、同 key
异摘要冲突、并发提交、CAS 已提交但响应丢失、无重复 Finding/CleanRound，以及原 Gate 和
本人回执不等待全局门。首个 Writer 切片不代表 typed S7 adapter 已完成；以上发行与 Hook
验收仍待实施。

### 13.7.1 S7 发行索引、卷面与原 Writer 接口（2026-10-14 固定，机制已接线）

使用已有 Runtime v1.2 `optimization.packet` 显式开关，不另加阶段开关。新的
`s7-review-result-v1` subtype 追加到原 `entities.work_packets`；不在
`entities.submissions` 登记第二份 S7 接收事实。尚未发行的 Assignment 保留原显式
`review-result submit` 行为；一旦当前 Assignment 已有有效发行，该入口必须要求原发行
绑定，不能换一个 Result ID 绕过卷面。历史未发行 receipt 不补造 packet。

中性 `workpacket.ReviewResultPacket` 持有封闭 Binding：`runtime_id`、
`baseline_generation`、`review_plan_id/ref/sha256/revision`、`review_round`、
`assignment_id/sha256`、`owner_agent_id`、精确有序 `claim_ids`、`subject_digest`、
nullable `verification_artifact_digest`。Inputs 保存该 Assignment 的原 Claim 题干与原
frozen subjects JSON；由 review 包从经校验的 pinned Plan 提取，不能由答卷提供。
packet ID 由 Binding、Inputs 与嵌入模板 SHA 确定；`result_id` 与
`issuance_event_id` 从 packet ID 派生，不参与自身哈希。view/answer 路径亦由 packet ID
派生，位于 `.claude/work-packets/.../S7/...` 与
`.claude/submissions/.../S7/.../<owner>/answer.json`。原发行事件固定
`PACKET-ISSUE`、orchestrator、`packet:<packet_id>`；事件不添加专业结论。

Draft 包含只读标题、说明、资料与问题，以及专业 `answer` JSON。answer 仅允许
`claim_results`、`blocked_claims`、`checks`、`findings`、`deviations`、`site_lost`、
`verdict`，以及 E2E 专业执行证据声明 `verification_artifact_digest`；后者不授予身份，
由作者对实际执行的 workspace 字节声明，原 `verifyResultArtifactDigest` 仍核对磁盘与原
workspace/seal 规则。cold-start workspace 是 Reviewer 的允许写面，不能在发卷时冻结空
workspace 摘要阻止随后写 spec/fixture；此时 Binding 的该摘要为 null，消费保留专业
执行声明并在原 CAS 再验证。若发行已固定非 null 摘要则答卷必须完全一致。草稿 Claim conclusion 与 verdict 为 null。严格 schema、重复键拒绝和
原 Draft 对比共同禁止未知字段与题干编辑。workpacket 不依赖 review 类型；完成的
专业 JSON 由 review 包映射至唯一原 `review.Result`。整卷预算 20 MiB，拒绝 symlink、
非 regular file 和根外目标。累计发行仍受原 Runtime 16 MiB checkpoint 预算约束，
在接受 CAS 前拒绝超限并预留 Writer 元数据空间，不能先提交再留下无法恢复的 marker。

发行由 Main 调用 `review.IssueResultPackets`，原 Runtime Writer 在锁内复验 pinned
Plan、generation/round、完整 Assignment digest、精确 Claim、真实派发 owner 与注册
Reviewer 身份。只发行 dispatched Assignment；queued 等原消费者释放后再发行。
相同 tuple 重试保持原卷与作者字节；具名修订产生新 tuple，旧卷不取得新写权限。
runtime 只持有 subtype lookup 和历史 checkpoint proof，不依赖 review；恢复先校验
冻结 packet 与原完整 journal 的发行事件，再从临时投影移除 S7 subtype 交给 S5
校验器，持久状态不被过滤或改写。

`review.SubmitIssuedResult` 接受 adapter 已解析的 owner 与精确发行路径，读取磁盘
professional answer，补机器 tuple，调用同一个原 SubmitResult 核心。内部发行 capability
不能由 Result JSON 或 CLI `--agent` 赋予。当前 S7 终端 consume 没有受信宿主 caller
capability，因此拒绝用 `--agent` 代交已发行卷；终端仍可读取 receipt，未发行卷保留
原 explicit submit。不能把 working/PLAN 合格的 owner 字符串当作 caller 身份。
首次提交在 preflight 和原接受 CAS 内再次
验证原 packet、Plan/Assignment/owner/subject 与 `agentauth.ValidateWorkingAuthorization`；
原独立性、capture、Finding、Claim coverage 和 round consumer 全部保留。receipt 与原
Result envelope 增加可选 `packet_id`/`packet_sha256`，在原接受 CAS 闭合发行身份；无发行
的旧 receipt 形状保持原样。重放还验证同一原发行与真实 caller owner，但不要求已 reported
的 Reviewer 再变 working；旧卷不因其专业摘要相同而恢复当前资格。

Worker 使用同一个 packet 的相对答案路径，读取根由原 ready execution/session-owner
校验确定。adapter 在 preflight 与原消费 CAS 复验 source execution（Runtime/gen、owner、
Assignment、物理根与原 execution proof），冻结一次实际落盘 bytes 后交同一个消费者；
不复制为 Main 作者文件。已有独立 execution 时 Main 答案入口退休；发行仅向该受信
execution 发布只读卷面和未答草稿，已填写字节不覆盖。

原 S7 product-drift/capture 扫描把经原 work_packets 与发行 journal 证明的确切 S7
view/answer 文件视为控制面输入。例外按文件，不放开 work-packets/submissions 整目录；
没有发行、损坏来源或 symlink 的同目录文件仍按原 drift 规则拒绝。declared frozen subject
继续先核 SHA，不能借控制面路径掩盖已冻结产品变更。

typed E2E 的执行摘要在首次接收与原 CAS 使用已认证 execution root 下的原
`verification_artifact_workspace`，不能用 Main 的同名目录替代 Worker 实际执行面。
packet-bound receipt 与原 Result envelope 记录 `execution_root`，原提交事件 identity
也绑定它；旧无 packet receipt 不添加此字段。封轮/CleanRound/原 Gate 对所有已接收
typed E2E rows 使用各自原 packet/receipt 与原 execution registry 复验源及 workspace
摘要，保持全体同一允许 workspace 版本的既有规则。零写重放只复验原接收/source
identity，不重新执行捕获或把变化的 live workspace 重新当成旧答案。只读
`s7 workspace-digest --root <Worker>` 通过原 control pointer 找到 Main Plan，并计算
实际 Worker 根下的目录；该查询没有提交权限。

保存 Hook 使用原平台 payload 的 agent_id 与同一 Runtime snapshot 的精确派发匹配。
Worker cwd 必须额外通过原 execution/session-owner 绑定，payload ID 与 session owner
不一致时拒绝；Main 缺 Reviewer 身份时不能代交。native 身份沿既有宿主输入边界，不能
防同机用户手工伪造 Hook stdin，真实 Claude 映射仍待宿主验收。PreToolUse 只开放当前
owner 已发行的精确答案路径，PostToolUse 对成功保存给反馈与原责任回执，不撤销写入。

<a id="s8-hypothesis-result-adapter"></a>

### 13.8 S8 HypothesisResult adapter（2026-10-14，固定契约，机制链已验证，平台验收待完成）

权威保持原 `InvestigationCase`。`RegisterHypothesis` 的 Assignment ID 是待派发身份，
不是已派发证明；必须等原 investigator manifest/TASK 经 `assignment.Register` 登记，
原 `ENTITY-REGISTER` 事件须记录登记时真实 baseline generation（不能沿用
Writer 的缺省 1）；v1.2 packet mode 的 Investigator Team 在同一原 Register
事务保存 `manifest_sha256`，TASK 继续保存原 sha256。发行/接收/恢复复验这份
原登记 SHA，不能把 Register 后改过的同路径文件当成原派发。历史缺登记
SHA 或错代事件不从磁盘补猜，须 Main 经原具名新派发恢复。再由 Main 在已有 work_packets 中发行 `s8-hypothesis-result-v1`，沿原 PACKET-ISSUE
事件，不增加 CaseResult 表、submissions、queue 或第二 Writer。

封闭 Binding 固定 Runtime/generation/review round、发行时 Case ref/SHA/revision、
ObservationBatch ref/SHA/baseline digest、Case 精确 Finding 集、Hypothesis ID 与问题
投影摘要/精确源 Finding 子集、原 Assignment 身份和 revision（或原 registered
manifest 身份）、Team、TASK ref/SHA、manifest ref/SHA 与真实 owner；另固定 Case route tuple 摘要。Inputs 保存
statement/invariant/discriminator/expected outcomes、原 Finding encounter/boundary
资料、原 status/route/route_reason/route_consumed_at/route_consumer（缺省字段规范为 null）
与经原 evidence index/SHA 验证的资料引用。发行只允许未决定 route 或未消费的
`investigate_more`；不能仅凭 status=investigating 放行已决定的其他出口。
packet/结果 key/view/answer 路径
由原发行确定；题干不从专业答案重建，路径或预填 ID 不授予身份。

专业 answer 只允许原 `HypothesisResultRequest` 的 method、evidence_refs、
source_boundary_refs、observed、counterfactual_or_discriminator、
result(supported/refuted/inconclusive)、explains_finding_ids、does_not_explain，
以及 L3 允许的 new_hypotheses 内容。草稿专业结论为 null；没有 route、causal model、
root cause、RepairContract 或 approval 字段。新 Hypothesis 的 ID/Assignment ID
由原消费者按 packet identity 与数组位置稳定分配，作者只给问题与来源；随后仍须
Main 用原 dispatch 登记，不能在结果 CAS 中另启动 Worker。

保存沿 S7 的受信宿主 actor、同 snapshot 的原派发与原 PLAN/working 校验；独立
Worker 另验原 execution/session/source root。仅读已发行的精确 bounded regular
answer，未知字段、重复键、题干编辑、身份/源变化拒绝。已有发行后旧 direct flags
不能绕过；终端 owner 字符串不能代替宿主身份。尚无发行的旧流程保留既有兼容入口。

专业答案最终进入原 `SubmitHypothesisResult` / `updateCaseRevision`。在 preflight
及原 Case Writer CAS 内复验 packet、Hypothesis/Assignment/owner/工作授权、当前
Case active/route、generation/round、原 route tuple 必须同时匹配不可变 Case 文件和 Runtime investigation pointer；
原 `ConsumeCaseRoute` 可只在 pointer 记录 consumption，不能只读文件而漏过
已消费出口，也不能用结果追加清掉原 consumption。原 Finding exact set、EvidenceRefs 的 SHA/有效性/
当前代数与轮次，以及 source_boundary_refs 确实来自被冻结 Finding encounter。
`source_boundary_refs` 接受原 intake 的 `finding:<source-id>`，或该 Finding 上
实际存在的 `#/encounter/...` JSON pointer；另外仅允许 sealed ObservationBatch
原 blocked Claim 已登记的精确 boundary evidence ref。不可凭任意 `scheme://`
字符串生成 boundary 证明，也不从可编辑 Case 补造该 allowlist。
解释/未解释 Finding 与新 Hypothesis 的来源不能超出原边界。原 HypothesisResult
行记录可选 typed receipt（packet ref/SHA、完整规范答案摘要、owner、派发来源 tuple、
原事件及后继 Case ref/revision）；接收事实仍在原 immutable Case revision 中。
Case 内嵌 receipt 不保存其所在文件的 SHA，避免自引用摘要；读取器从原提交事件、
当前 pointer 或完整原 revision_history 重建后继 ref/SHA，并校验实际不可变字节。
索引 evidence 按原有效性/代数/轮次/SHA 校验；原 `scheme://` execution anchor 保留
既有执行锚语义，不伪造内容 SHA 或声称仅凭字符串已验证实际执行。

多个 Hypothesis 可并行。发行时 Case ref/SHA/revision 保留作审计；兄弟结果追加的
无关 Case revision 不使其他卷失效。接收须证明当前仍为同 Case/ObservationBatch/
baseline/Finding 集、同一未答 Hypothesis 与原问题/Assignment/owner/资料，且冻结的 route/消费 tuple 无任何变化
且未终态，然后对最新 Case pointer 做原 CAS。自身问题、来源、派发或路由变化后
旧卷拒绝并具名重发；Main 重复发行对仍兼容的槽保留原 packet 和作者字节。
不能在相关 tuple 改后又改回时复活旧卷：原完整 Case revision_history 自发行
版本到当前逐版证明上述本人闭包未变；源版本缺失或不连续拒绝。原 journal 中
发行后此 Case 的 `S8-ROUTE-CONSUME` 也使未答旧卷永久失效，即使后续 Case
修订重置当前 consumption。已接受答案仍只读报告其原事实，不能变成新授权。

相同 packet/完整答案摘要重放先验证原 Case 结果、发行与 journal，零写返回原回执；
异答冲突。实际保存重放仍须读取当前答案并比较，丢失作者文件时用独立只读 receipt
查询恢复事实。CAS 不明先恢复原 Writer，再查原 Case/Runtime/事件，不猜成功或删
可能被 pending 引用的产物。本人接收与全 Case 的解释/路由进展分开，原 route consumer
继续判断整体状态。已发行 Investigator 的 Stop/Idle/Controller 完成提示也只
消费同一原 Case receipt proof；通用 completion_report、reported 或 one_shot
不能替代 HypothesisResult。接收 owner 可结束本人责任，不等待全 Case route/批准；
fresh 写授权撤回不抹去原已交付身份。未发行的旧 Investigator 保留原行为，
原 blocker 或正式已退休责任沿原路由处理。

`UpdateCaseRoute`、`ConsumeCaseRoute` 与 `ApproveContract` 保持原职责。S9 批准仍
需要绑定当前 Case/Contract、精确 draft hash 的有效、未消费 scoped human_decision
（或原已授权 delegation），在原批准 CAS 消费；答案和 approved-by 名字不构成人闸。
高风险实验仍走原计划/人闸，结果答卷不补造授权。此契约已根据原 API 只读审计固定，
中性卷面、Schema、原来源/回执读取器及 Writer/Hook 已接线。真实 generation-2
Register/PLAN → public PostToolUse → Case/receipt → Stop/Idle 组合、同一 Case
兄弟结果接续、原 approval gate 回归、route pointer/history 失效保护、并发唯一提交与
sealed pending 恢复均有定向机制覆盖。Combined normal/race 分别 10.948s/90.704s
通过；Runtime source/continuity race 15.922s，Adapter normal/race 亦通过，日志与
源 SHA 见 [S8 机制证据](../evidence/global-optimization-s8-issued-hypothesis-result-20261014.json)。

答案与 source 注册闭包由测试夹具提供，不能当作真实模型专业判断；Register/PLAN
和公开 Hook 调用也不是独立 Claude Investigator 的实际 prepare/launch。Pending
测试用真实已接受 state/journal 构造 sealed marker 并恢复提交前 pair，不是进程硬杀。
若 Case 整体被替换且不再保留原 pointer/revision history，历史 receipt 读取 fail-closed。
Intake 父目录 Unix sync 已接入；Windows 当前只验证目录，不保证断电后的目录项持久性。
完整安装、支持平台/真实宿主链与真实模型仍未验证；全局核心链 4/9、真实模型链
0/9 不变。S9 PlanReport 正在启动，其共享 Runtime/workpacket/schema 变化不包含在
本 S8 验收范围内。

<a id="s9-plan-report-adapter"></a>
### 13.9 S9 PlanReport adapter（2026-10-14，已接线，原组合与恢复验收通过）

本切片只把执行前专业计划接入原 `CreatePlanReport` /
`SubmitRepairPlanReportToRuntime` 和 `PTR-BUG-10`。RepairResult、独立 targeted
reverification、ChangeImpact 与 Handoff 分别继续走原消费者，下一切片逐段固定。
PlanReport 接收是执行前里程碑；receipt 的 `work_complete=false`，不能凭它
结束 Builder、满足 Stop/Idle、授予产品写入或代替完整 RepairResult/Main 集成。
原 Main `BeginRepairExecution` 仍须全体 Assignment 报告完整后才能放行。

批准来源沿原 S8 scoped human decision/delegation 的实际消费与原
OpenRepairSession，不从 approved-by 字符串或作者答案补造授权。Main 只为
原 `runtime repair dispatch` 已实际登记、当前 planning/reproducing 的 Builder
发行 `s9-plan-report-v1`，复用原 work_packets/PACKET-ISSUE。排队、独立 targeted
verifier 和只有磁盘 manifest 的未登记身份不能发卷。原 Register 在 packet mode
且请求绑定 S9 RepairAssignment 时，在同一原 Team 行保存实际 manifest SHA；
旧无 SHA 记录须具名新登记，不能发行时从当前文件补猜。

封闭 Binding 固定 Runtime/generation/round、当前 approved Case ref/SHA/ID 与
Contract ref/SHA/ID、Session ref/SHA/ID、RepairPlan ref/SHA/ID、原 RepairAssignment
ID/规范行摘要、其 unit/assertion/scope/dependency/lock 闭包、真实 owner、原
manifest assignment alias/规范行摘要、Team/TASK ref/SHA、manifest ref/SHA。
Inputs 提供上述原不可变文档、本人 Assignment 与原 TASK、prospective/forbidden
scope。实际文件 bounded regular/no-symlink、唯一 JSON 键、Schema/SHA，原
ENTITY-REGISTER 与 PACKET-ISSUE 的 actor/runtime/generation/cursor/ref 对均复验；
不能用重新组装的专业答案当题干。其他 Builder 追加 PlanReport 不使本人卷失效；
Session/Plan/Contract/Case/owner/本人派发来源变化使旧未答卷拒绝，具名恢复。

专业 answer 仅 `plan`、`red_checks`、`proposed_paths` 三项，草稿均为 null；
red check 结构沿原 name/command/result/evidence_refs，至少一项 fail/blocked，
原 execution anchor 形状约束保留。anchor 字符串不声称实际命令已执行。
路径须非空且同时在原 Assignment 与 prospective scope 内、在 forbidden 外；
不添加作者给定 ID、assertion list、审批、status、reported_at 或产物 SHA。
系统从冻结源推导原 report/session/plan/assignment/assertion/owner 字段。

实际 successful save 由受信 Hook actor 和原 source/session 提供身份；preflight
及原 Submit CAS 都检查同 snapshot 的原派发、当前计划来源、原 generic
PLAN/working。独立 Worker 另验原 execution root/session；没有支持的 prepare/
launch profile 必须具名拒绝，不能回退 Main。终端 --agent 与 body owner 均不
授予代交；旧 direct submit 对已有发行的本人目标拒绝，未发行旧路径保持兼容。
原 Submit 的 Apply 内补齐上述来源/状态复验，不允许 preflight 后旧源提交。

稳定 report ID、事件 ID、idempotency key 和 reported_at 由原发行身份/时间
确定。完整规范 submitted_answer 与摘要、packet ID/SHA、owner/源 tuple、原
接收事件/key/sequence/revision 存在原 immutable PlanReport 的可选 typed receipt；
不嵌入所在文件 SHA，读取器从原 repair pointer 的 report ref/SHA 与完整 journal
验证原事实；不增加第二 registry/Writer。原当前 Session 的 plan_report_refs
保留多 Assignment 报告，接收 CAS 从当前原指针合并，不能覆盖兄弟报告。

相同 packet/完整答案零写返回同一原 receipt，异答 conflict。实际 save replay
仍读取当前作者字节与源授权；缺作者文件只允许独立只读 receipt 查询。幂等查询
先于 caller ExpectedRevision 或当前 execution phase 检查，接受事实不等于新授权。
并发与中断沿原 Writer sealed pending 恢复后重查原报告/事件；不猜成功、不删
pending 引用的 staging 文件。发布前遵守现有 state/journal exact budget。
历史 Session 已整体被替换且原索引未保留时，具名报告无法从当前 authority
恢复，不扫目录猜旧接受事实。此限制不构成新 registry 的理由。

完整 immutable 报告同时保存专业投影与 submitted_answer。20 MiB 答案输入上限
不保证其双份投影仍落在产物预算内；原文件函数在任何创建/发布前按最终
序列化字节另验 20 MiB 上限，超限具名拒绝，不能先提交再得到不可读回执。
原 S9 immutable artifact 文件函数保持一个入口，读时限制 20 MiB regular
no-symlink 并复验 Schema/SHA；写时固定原相对路径与逐级目录，exclusive
创建后 sync 文件、关闭并同步父目录，再交原 Runtime CAS。目录同步结果
不明保留文件，后续只经原 pending-safe 引用/字节校验清理；Windows 沿既有
目录同步能力，不宣称断电持久性。未增加第二种领域 Writer。

验收须包含真实原批准/Session/Plan/dispatch/PLAN→发行→公开 Hook 保存→原报告
与事件→重复零写/异答拒绝、兄弟并发、源/身份/phase 负向和 sealed pending
恢复；原 execution begin 前产品写屏障、全报告完整性与 Builder 未完成边界均保留。
本契约已落地至原领域 Writer、Runtime/Schema、Hook/权限及角色资产。原批准→
Session/Plan→实际 generation-2 派发/PLAN→公开发卷/保存→原报告/回执→
全报告 Begin 的实际机制链，以及同卷并发、sealed 报告恢复、公开发卷 pending
恢复与 alias source 拒绝的合并 race 已通过，Hook 仍为原 6 秒 deadline。
专业结论为机制夹具；此处不是完整 S9 RepairResult/targeted/Handoff adapter、
真实 Worker/platform/model 或全安装矩阵完成声明。当前证据与剩余工作见
[全局 §7.49](../GLOBAL-OPTIMIZATION-DESIGN.md) 和
[S9 本切片证据](../evidence/global-optimization-s9-issued-plan-report-20261014.json)。

<a id="s9-repair-result-adapter"></a>
### 13.10 S9 RepairResult adapter（2026-10-14，原正向链、定向恢复与本地包通过，平台待验收）

本切片从 §13.9 的原 PlanReport 全覆盖及 Main `BeginRepairExecution` 继续，
把专业结果接入原 `CreateWorkspaceCandidate`、`workspace_updated` delivery 与
Main `workspace integrate` → `SubmitRepairResultToRuntime` / `S9-RESULT-SUBMIT`。
候选是 Worker 的原交付，最终 RepairResult 是 Main 对实际集成结果的原接收；
两个事实分开验证，不增加 submissions 表、candidate registry、第二 Writer、
调度器或由作者自报的 PASS 证明。Targeted、ChangeImpact、Handoff 与 fresh S7
仍须各原消费者独立完成；本节契约及原正向链已落地；定向验证进展见 [全局 §7.50](../GLOBAL-OPTIMIZATION-DESIGN.md#s9-repair-result-progress)，不声明这些后续段已完成。

#### 发卷与不变来源

在原状态 `repairing`、原 approved Case/Contract/Session/Plan 与本人已接收
PlanReport 有效，且原 registered execution 实际 prepare/ready 后、candidate
形成之前，Main 经原 work_packets / PACKET-ISSUE 发行 `s9-repair-result-v1`。
不得先要求作者手填旧 `repair-result.json` 再重复填写同一专业答案。来源沿
§13.9 的原审批消费、实际 dispatch、TASK/manifest/Team/owner 与 PLAN 闭包，
另固定本人精确 PlanReport ref/SHA、其 receipt/原接收事实、精确 unit 集和原
assertion/scope/dependency/lock。旧缺这些原登记事实的身份须具名恢复，不能从
当前磁盘补造原授权。
批量发卷只选择已实际 prepare 的 Assignment；兄弟尚无 execution 时保持未发卷，
不能阻止已就绪目标发卷。已有同目标 execution 但 owner/代数/key/状态或来源
不一致时具名拒绝，不能把损坏登记解释成尚未 prepare；显式指定未就绪 owner
时给出原 prepare 恢复动作。

Binding 另固定真实 Execution 的 assignment canonical/manifest alias、Runtime/
baseline generation/execution generation、Worker path、branch/target branch、
BaseCommit、check_location、write scope、inputs、原 frozen check profiles 与
bootstrap/source/session 授权闭包。题干提供原源文档与本人执行定义；不冻结
尚未完成的最终 SourceHead，不把 coding 中正常 HEAD 改变误判为题干变化。
执行定义摘要排除 ToolCheckReceipts 等进展行和 DeliveryRef 等接收指针，但这些
进展仍在接收时按原 authority 实时复验，不从排除摘要推导可以忽略。

原 `workspace rework` 不递增 execution generation，故 Binding 必须额外固定
当前 ReworkRef 及其文件 SHA，初轮明确为空；非空时复验原不可变 rework receipt、
原 `integration_rework_requested` 事件、原待改范围和当前 execution。新 rework
清 delivery、封存失败 checkpoint，须具名发行绑定新 lineage 的卷。旧 packet/
candidate/receipt 保留审计事实，不能因可见 source tuple 改后又改回获得新授权。
同槽重复发行保留作者字节；具名恢复不重置未确认的 pending marker。

#### 专业答案与机器事实

封闭 answer 保留 L3-S9 §3.4 的专业责任，不仅搬运旧薄 API 的六个字段：
`verdict`（completed/blocked/contract_conflict/stale）、`before_fix_results`、
`unit_results`、`detection_gap_artifacts`、`evidence_refs`、`scope_deviations`、
`residual_risks`、`migration_ref`、`rollback_ref`。unit_id 是题干给定的精确槽，
作者填写 status/evidence_refs、implementation_summary、restored_invariant_refs；
原 wire 名 unit_results 对应 L3 的 repair_unit_results，不另存第二份数组。
before_fix_results 对精确原 assertion_id 槽逐项填写非空 observed_red 文本和
资料引用，不能由 PlanReport 的命令 fail 字符串补造逐 assertion 的专业观测。
detection_gap_artifacts 是专业资料引用集合，不自报 artifact SHA 或新 authority。
草稿专业结论为 null，不从全绿命令推断 completed；Schema/路径形状不证明业务
判断。兼容迁移沿原 Contract 的 migration_ref/rollback_ref 要求。

原领域 Result.result 是不同 wire enum。本切片仅在原 candidate 路径接受
completed，机器投影为 pass；原 Result 和完整答案仍保留专业 verdict，不能
反向从 pass 补造专业 completed。blocked/contract_conflict/stale 不压平到同一
blocked，不将合同冲突误报为实际越界；在后续各自原 recovery adapter 固定并
接线前，本切片具名拒收这些候选，保留原 blocker/保全路径，不发成功 receipt。
机器 authority freshness 失效也拒收并要求原 stale/session 恢复，不能用作者
verdict 获得有效来源。旧未发行 API 的 fail 路径保持原语义。原 RepairResult/
RepairUnitResult 扩展可选上述专业字段，typed 接收强制完整，legacy 不补造。
资料 ref 字符串仅是专业引用声明；资格仍须原 evidence consumer/真实字节证明，
不能把形状合法字符串等同于机器已验证 ArtifactRef。
全局 ChangedArtifacts 仍由实际 diff 生成；逐 unit 的改动归因不能仅凭同在
Assignment 就猜测。原 Unit scope 无充分唯一归因时，保留完整实际 diff 与专业
unit 结果，并将更细粒度归因列为尚未完成的 L3 语义，不据此宣称完整 canonical
S9。身份、审批、时间、SHA、ChangedArtifacts、BeforeFixChecks、Checks、
checkpoint、ACK、状态或权限均不由答卷提供。

BeforeFixChecks 无条件取本人原已接收 PlanReport.RedChecks，PlanReportRef
也取同一精确原索引；不能保留作者自带非空 before_fix_checks。保存时由机器
读取真实已提交、干净 Worker HEAD、BaseCommit..HEAD 的实际 diff/blob SHA
（含删除的原字节），在原 Assignment/prospective scope 内、forbidden 外，
构造原 candidate Proposal。专业投影须与完整规范答案逐项相等，最终结果再按
原 Session 实际 Main diff 核验；不能把候选 SHA 当 Main 文件或最终检查证明。

第一实施拓扑仍是原 registered separate Worker worktree → Main merge/check。
check_location=main 表示该 worktree 的声明检查由 Main 在集成后执行，不表示
在 Main checkout 原生编程。S9 native Main-checkout checks 与 S6 canonical
ClosingContract.WorkerChecks 尚未形成相同原路径，不能复用其名字宣称已支持。
不支持的 topology、source/session 或 isolation 具名拒绝，无不受限 runner 回退。

原 S9 dispatch 在专业 PlanReport 之前登记，不能凭空预填后来才由 Builder 提交
的 red-check 命令；它也没有 S6 ClosingContract。S9 的原 prepare 必须先消费
已接收 PlanReport 的实际命令集合，再由 Main 明确选择原 worker/main 位置，
在同一原 Execution prepare CAS 用报告 ref/SHA、命令原顺序和所选位置冻结
确定性的保守 CheckProfiles：whole_tree、inherited environment、unknown side
effects、reuse=never、原超时上限；环境/依赖列表显式空。此格式描述执行来源，
不声明测试通过，不推断业务 oracle，不伪造 S6 closing_contract。已有执行不得
在原位补写 profile 或改变位置；缺 profile 的旧 execution 须沿原具名替换恢复。
Runtime 来源及原接收 Apply 再核对命令仍来自同一原报告、profile 仍为该冻结
定义；Main 仍实际执行每项。若已有原专业派发 profile 则须来源严格一致并原样
冻结，不能用保守缺省覆盖它的声明；本次初始 S9 路径没有这些旧声明。
原 Main runner 按命令定位 profile，不能区分同名命令的多个 index；当前保守
准备分支对重复 command 具名拒绝并要求原计划恢复，不静默去重、代填 PASS 或
接收一个之后无法证明 exact coverage 的候选。该限制须在后续原 runner/index
扩展中关闭，不能计为全部 S9 检查矩阵已支持。

#### 原 candidate 接收与本人责任

保留原真实 completion_report/Agent=reported 前置；completion_report 不自动
调用 deliver，也不能替代 typed candidate。先报告、再保存专业答卷形成候选，
Stop 在两者之间看到欠交 candidate 时继续原阻断并给出具名下一动作。受信
Hook 仅接受平台 owner、原精确 execution/source 和已发行 answer 路径；终端
owner 字符串、body agent_id 或 Main root 路径不能授予 Worker 代交身份。
本代原 S9 使用非 S6 Record 的 completion envelope：校验原 Agent/TASK 指针、
evidence 索引及其精确文件 SHA、owner/Runtime/代数/TASK，并核对唯一原
`BUILDER-RESULT` 原接收事件（生命周期事件为 completion_reported，原 Writer 的
raw journal event 为 transition_committed）。报告必须晚于 Begin/本人当前
Rework；可在本次发卷前或后，但必须早于 candidate 接收。candidate provenance 冻结
`completion_report_ref`、`completion_report_sha256`、`completion_report_event_id`；
只读历史回执也验证此原报告先于 candidate 接收，不以当前 reported 标签代替。

原 immutable WorkspaceCandidate 增加可选 typed provenance：packet ID/SHA、
完整规范 submitted_answer/摘要、owner/source/attempt tuple。它不含自身文件 SHA
或预测的接收 sequence，从而保留 canonical 全文 SHA 作为原文件名/DeliverySHA。
原 `workspace_updated` CAS 登记 Execution.DeliveryRef/DeliverySHA256，同时
EvidenceIDs 指向精确 candidate；Apply 复验发行、当前 approved 源、owner/PLAN/
reported、execution/attempt、答案当前字节、干净 SourceHead/diff 与候选全文。
兄弟 execution 的当前记录按原 CAS 合并，不能用旧整份 registry 覆盖当前进展。
候选 receipt 从原 pointer、唯一原提交事件与 candidate SHA/provenance 重建；
仅存在文件不表示接收，typed provenance 也不能自授予原事件或私有 capability。
已有 typed 发行时旧自由 proposal/deliver flags 必须进入同一原消费者或拒绝。

原 Hook 6 秒预算保持不变。同一 revision 的 preflight 可复用本次读取已核验的
execution/packet/journal 解析结果，减少重复全闭包与 Git 调用；CAS Apply 仍从
当前 state 重新验证来源、PLAN/reported/completion、答案、HEAD/实际 diff 和 checks。
Hook 的 trusted-source callback 负责平台 owner/session/实际 registered checkout
身份；领域 permission/source closure 归原接收器逐边界验证，不在 callback 内
嵌套重复整套原 receipt/审批/发卷校验。外部文件或授权结论不跨 revision/CAS
复用，也不以增大超时或跳过当前 proof 使慢路径通过。当前两轮原链候选 Hook
实际响应为 8.584s / 5.880s；6 秒是原 context 配置，不能据此承诺完整响应
均低于 6 秒，仍需后续性能与原宿主验证。

检查归属从原 frozen execution/profile 推导。check_location=worker 的本人
required checks 必须逐项消费原已登记 ToolCheckReceipts，复验实际 bytes/log SHA、
profile/Runtime/generation/owner/execution/subject/current code 与原 runner 事实；
作者的 checks 数组或自由 PASS 不算证明。check_location=main 的 e.Checks 归 Main；另有明确冻结的 e.WorkerChecks 时，
仍逐项消费原 OwnedCheckReceipts，不能把该独立集合也推为空；S9 未接线的
canonical ClosingContract 来源不得伪造。两种本人集合都确认为空时才没有本人
检查欠项，typed candidate 可显式保留机器 Checks 空及 Main pending；仅对可信
已发行 typed candidate 的原验证分支允许此待消费态。旧未发行 candidate 保持
原非空 passing checks 规则，最终 Result pass 仍严格需要 Main 实际全量检查。

原 candidate 路径目前只接收 pass、全 unit pass、无 scope deviation 的交付；
专业非 pass 不能被改写为 pass 或静默集成。此切片对非 pass 保留原 blocker/
保全/领域 recovery 并给出具名动作，尚未接入的 typed 非 pass 原接收不得输出
成功 receipt。补齐其原 recovery adapter 是后续必做项，不计入完整 S9 已完成。
后续非 pass 接收不得继承 completed candidate 的 completion_report/reported
前置：真实受阻作者应能在 working、来源仍有效时报告，不要求先谎报完成。
原 `work_blocked` 只支持 working→blocked；现有 generic blocker 的自由 body、
Agent 标签和 TASK blocker 字符串都缺 packet/完整答案/当前 attempt 的接收证明，
不能据此输出 typed receipt 或当作现成恢复 adapter。优先沿原 blocker 消费器
扩展精确绑定与同 CAS 复验；具体入口和结构在下一切片固定后才实施，禁止
泛增 reported→blocked 状态边或另起队列。需区分答复已接收、当前 owner 可结束、
修复工作完成三个事实；非 pass 不得释放 Main 集成、全局成功或 fresh S7 门。
机器发现来源 stale 时仍拒绝旧卷；作者填 stale 不能覆盖实时授权失效。

遵守 G02：当本人 candidate 已在原接收、全部本人 required checks/报告齐全且
当前输入有效，可投影 `work_complete=true`、`consumer_owner=main`、
`consumer_pending=true`，让本人结束，Main 继续其集成责任。PlanReport 自身仍为
false。Main-owned checks 或全局阶段未过不能反推 Worker 欠项；本人检查缺失/
失效则 false，原有效 blocker 可结束当前执行但不等于任务完成。Stop/Idle/
Controller 统一消费此原 proof，generic reported/completion/one_shot 不能替代。
最终 receipt 中冻结的 work_complete 是原接收事实；当前 Stop/status/receipt
责任投影另匹配当前 execution 定义摘要、owner/代数/attempt 与精确 delivery。
合法 ACK/cleanup 后的 ready→complete 不恢复已删除 Worker 目录，也不使责任
重新变欠交；明确 rework 或执行定义改变后，旧 receipt 仅保留历史 Accepted，
不能重新授予当前作业完成。
Main 的另一 `task-integrate`/HandleSubagentStopForController 路径对已发行 S9
必须核验同一 candidate 并路由原 workspace integrate，或具名拒绝并提供该命令；
不能从 generic completion report 直接绕过 typed delivery/Main Result 接收。

#### Main 实际集成与最终原 Result

保持 workspace→Runtime 锁顺序，沿原 `integration.LockWorkspace` context
可重入能力覆盖 Inspect/Integrate/Main checks/最终 Result 接收；原领域的窄
trusted integration 入口也持有/取得该锁，不能只在 CLI preflight 检查一次。
在 Main 原 verified checkpoint 后，读取原 candidate 与当前 authority，证明
SourceHead/BaseCommit、worktree/target、Runtime/gen/execution/owner 和当前 Main
HEAD=TestedHead；按原 Session/Assignment 重算实际产物，再由原机器检查结果
生成最终 Checks。原 receipt 同时包含完整答案、packet/candidate/refSHA、精确
PlanReport、execution/attempt 与原接收事件/key/sequence/revision。

Main proof 绑定原 checkpoint ref 与 verified facts 投影摘要：原 source/base/
merge/tested head、required check/profile 集、ToolCheckReceipts/log SHA/runner
闭包。复用 `integration.VerifyCapturedChecks` 及原 ProfileCheckRunner 证据，
另逐项证明 exact declared checks（locked metadata 除外）都有匹配实际 receipt，
绑定 Runtime/baseline/owner/execution/responsibility 与 tested head、cwd、exit/status。
仅循环已存在 receipts 的空集或可编辑 JSON 的 pass 字段不能满足覆盖。原最后
领域 Writer 的 private issued/integration capability 才可携带上述机器事实；
preflight 与原 Submit Apply 同时复验 packet/完整答案、当前源/PLAN/owner、
candidate/refSHA、execution、checkpoint/head/check receipts 和 exact result batch。
直接旧 result submit 对已发行目标拒绝；未发行兼容入口不据此声称有 Main proof。
结果 refs 按 Apply 的最新原集合合并，保留其他 Assignment 已接受行与原批次门。

完整 typed receipt 只存在原 immutable RepairResult 可选字段，由原
S9-RESULT-SUBMIT CAS 索引；无所在文件自引用 SHA，无旁路结果表。原 checkpoint
在 Result 后还要 ACK/cleanup，故 receipt 固定 verified facts，而非要求整个可变
checkpoint raw SHA 永远不变；可记接受时 raw SHA 作审计。后续原
verified→acknowledged→cleanup_pending→complete 与有原来源的合法 rebind 可继续，
冻结的验证事实不得变化。最终接受后从原 Result/索引/journal 重建历史事实，
不能用重跑检查、当前 stage 或作者文件存在代替原接收。ACK/cleanup 仍归原 Main。
已登记 typed Result 的原 CLI 重入，在任何 Integrate 前先从原 Result/索引/
journal 与当前 checkpoint 复验精确 receipt，之后仍再次复验。初次 S9 Inspect
通过 ValidateDelivery 校验 typed candidate 时有意不额外绑定 generic completion
report；恢复必须保持该 checkpoint 原 Inspection，不可用当前 Agent 的普通
completion 指针补入新 binding，误触发“报告改变”重跑检查并改写原 verified
facts。typed provenance 原完成报告 ref/SHA/event 仍逐次复验；不能借此忽略
原报告篡改、缺失或当前 Main HEAD/检查覆盖失效。legacy 恢复保持其原 report
refresh 规则。已接收 typed Result 且 checkpoint 仍为 verified 时，在同一锁内
先确认当前 target HEAD 等于原 TestedHead；不等则具名拒绝并保全 checkpoint/
Result/Runtime，不能先调用 Integrate 自动重跑、改坏原事实后再拒绝。已 ACK 的
cleanup/历史回执仍允许 Main 后续合法前进。真正新验证结果须进入具名恢复/
新接收，不能改写已收 Result。

#### 重试、恢复与验收

同 packet/完整答案/原 candidate 和 accepted receipt 精确重放零写；异答或
不同 SourceHead 的旧 attempt 冲突，须原 rework 后新卷。实际 save replay
仍验当前作者字节和受信 source，缺作者文件走只读 receipt 查询；历史 receipt
不授予新的编辑或新任务完成。已接受 Result 后尚未 ACK 的重试从原索引与
checkpoint 继续 ACK，不重复写 Result；checkpoint 丢失具名拒绝，不猜已验证。
原 sealed pending 不明先由原 Writer 恢复 exact pair，再查 candidate/Result/
journal，pending 引用产物不清理。未接受 staging 只能经原 pending-safe
cleanup 和精确字节/当前引用验证重试，遵守原 20 MiB artifact、state/journal
最终预算与安全文件/目录同步边界。历史 execution/Session 无原保留 authority
时具名拒绝，不扫目录重建接受事实。

必须覆盖原 approved gen2 chain、Begin 前写屏障、实际 prepare/PLAN/report/
发卷→公开保存→原 candidate→本人 Stop/Main pending→真实 Git merge/Main checks→
原 Result→ACK/replay；同卷并发、兄弟 refs、source/HEAD/profile/answer/身份负向、
篡改 check JSON/log/缺项/旧 subject、原 rework 同 generation 新 lineage、
Result-before-ACK 与 sealed pending 恢复。Main-only checks 与 Worker-owned
checks 分别验收；实际 host 不支持隔离时记录未执行，不伪造 runner receipt。
Main-only checks 原正向链、候选并发及 candidate/Result sealed-pending 原恢复
normal 已通过；定向 helper/CLI race 与整条链 race 分别记录。final 并发/stale
staging、原 rework lineage、Worker-owned 检查矩阵及真实支持平台验收继续补齐。
独立五目标包/临时安装通过，详见 [全局 §7.51](../GLOBAL-OPTIMIZATION-DESIGN.md#s9-repair-result-package)；
此状态补注在机制/包 source snapshot 冻结后追加，不刷新旧摘要，也不提高全局
4/9 核心链和 0/9 真实模型链计数。

<a id="s9-repair-result-blocker"></a>

### 13.11 S9 non-pass RepairResult 与 blocker/resume（2026-10-14，设计固定，数据模型基础已实施，接收待接线）

本节固定四种专业 verdict 的接收边界，数据模型基础不代表接收链已经实现。`completed` 保持
§13.10 原 candidate→Main integration→Result 路径；`blocked`、`contract_conflict`
与机器 tuple 仍 fresh 时作者报告的 `stale`，由原 Agent lifecycle `work_blocked`
接收完整答卷和 typed receipt，Agent 进入 `blocked`，但不生成 candidate/Result、
不接受 merge、不宣告 PASS/WorkComplete。机器 tuple 已漂移时，拒绝旧 packet，
保全答卷和 staging；不能用答卷里的 `stale` 覆盖机器 freshness。已 reported 的
Agent 仍不能退回 blocked。读取／接收 Professional Answer 与 `CleanInputs` 分离：
non-pass 不要求 clean commit、无 dirty/untracked 产品变更或全绿 checks；但必须
继续执行原 `ExecutionRegistry.ValidateInputs`、`ValidateExecution`、packet source
验证、PLAN/authorization 验证，并验证任何被引用的真实 receipt。跳过 `CleanInputs`
绝不等于跳过冻结授权输入的新鲜性检查。

新接收的分支固定如下；只读历史回执或已接收的精确重放仍按原接收事实返回，
不因此恢复产品写权限。

| 专业答案与当前条件 | 原消费者行为 |
| --- | --- |
| `completed`、所有 unit `pass`、无 scope deviation，且原 reported/completion/check/source 条件成立 | 沿 §13.10 接收 candidate；Main 集成/检查及最终 Result 仍分开执行 |
| `blocked`、`contract_conflict` 或作者 `stale`，机器 tuple fresh 且 Agent 真正 `working` | 沿本节原 `work_blocked` 接收完整答案，保持专业 verdict 原值 |
| `completed` 但 unit fail/blocked、scope deviation 或其他 candidate 条件不成立 | 保全原答案/staging，具名拒绝；不自动转为 blocker，不发完成回执 |
| non-pass 且 Agent 已 `reported`，或任一新接收的机器 tuple 已 stale | 保全并拒绝；不创造 `reported → blocked` 或用旧 packet 接新责任 |

Schema 可表达 `completed` 但不可集成的完整答案，不等于当前 adapter 已有原领域
接收路径。未来若接收这种诊断，必须另行固定与原 `result=fail`、本人责任及生命周期
接点的关系；本节不宣称所有 schema-valid 九字段答案都可接收。

#### 原 Main Hook 身份来源与 lifecycle capability

Main 对普通 blocker resolve 的授权只沿现有 Claude Hook 信任边界建立，不接受
文本 actor/flag：原 `hook → cli.run.evaluate → runSavedStageAnswerHook` 路径在真实
`PostToolUse` 成功回调中识别 exact resolution authoring file。必须同时要求：

- `Event=PostToolUse` 且实际 `ToolName` 为 Write/Edit/MultiEdit；读取平台传入的
  `tool_input.file_path`，规范化为该 blocker packet 唯一派生的精确 Main-tree
  authoring path。Bash 文本、其他文件、目录或相似前缀均不能触发。
- `CWD` canonical 后等于 `ResolveControl` 得到的原 Main root，且 `PointerRoot`
  证明它不是 Worker checkout；平台输入的 `AgentID`、`TeammateName` 必须为空，
  `SessionID` 必须非空。再从同一 Runtime/workspace registry 核对该 session 未被
  登记给 Worker；不把 `workspace.RequireMain` 或 Main-root CWD 单独当作身份。
- Hook 事件必须来自已配置的 Claude `PostToolUse` 与本机同 OS control bundle。
  当前协议把该宿主回调当信任边界；Hook JSON/`policy.Input.Source` 本身没有签名，
  也没有抵抗本机进程伪造 stdin 的认证承诺。`ToolUseID`/规范事件指纹可绑定幂等，
  不能单独证明主体。若部署不满足本项目既有 Hook 信任假设，必须停止 typed resolve。

当前 Worker 身份可由 `workspace.SessionOwner` 的 Worker pointer + registered
Execution + `PlatformSessionID` + `CWD` 验证；Main 则没有独立持久化的平台 session
注册，主 CWD 上 `SessionOwner` 返回空 owner。S7/S8/S9 Main Hook 目前也以顶层
Main session、无 Agent ID 与 canonical Main CWD 路由。该上下文足以沿本项目既有
本机 Hook 信任边界派生一次性 capability，但不是密码学 Main principal 证明。
`controller.ControlRequest` 只承载同一 Hook 解码字段；`runtime agent-event` 的
`--agent-id`、`--message`、原生命周期 `Actor="orchestrator"`、自写 message、
`--actor` 或单独 `LockWorkspace` 均不授予 resolve authority。对已发行 S9 target，
直接 CLI `runtime agent-event --event blocker_resolved` 必须因缺 capability 拒绝。

Hook 只在 exact authoring path、current blocker/packet/owner/attempt、原 `repairing`
pointer 与全部来源验证成功后，构造进程内不可序列化的 opaque capability。它通过
`AgentEventRequest` 的可选 `IssuedRepairResultBlocker` /
`IssuedRepairResultResolution` capability 字段携带 blocker 或 resolution proof；
其 repair-owned 类型的内容为私有，JSON/CLI 不能构造有效值。能力绑定 Host
session/CWD/工具事件 tuple、答卷或不可变 message
bytes/ref/SHA、前 blocker artifact ref/SHA 与当前 Runtime/generation/Assignment/Execution
tuple。随后调用统一签名
`assignment.AdvanceAgentWithContext(ctx, root, statePath, journalPath, request)`；保留原
`AdvanceAgent` wrapper。该入口须委托给同一 `AGENT-LIFECYCLE` Writer/CAS，不创建第二
Writer。Main workspace lock 从 Hook prepare 持有到原 lifecycle CAS，workspace→Runtime
锁序固定，传同一个 context 做可重入锁；Apply 使用当前 CAS state 和只读原 journal
复验，不能在 Apply 里再 Snapshot/开 Writer。Mutation 显式写当前
`BaselineGeneration`。私有 capability 仅用来满足原 lifecycle 的 typed guard，不能
代替 CAS、agent transition 或原 `applyAgentEventParams`。
原 lifecycle 仅在携带 typed capability 的分支获取/重入该 workspace lock；legacy
ACK/AutoChain 等既有 wrapper 调用不统一新增锁，避免已持锁且没有传 context 的旧调用
二次取锁。generic blocker/resolve 的全历史 guard 在 preflight 和同一 Apply 都复核，
缺 capability 的已发行 target 拒绝，不能因保持旧锁行为而绕过。

Blocker prepare 的具体接口固定为 repair-owned
`PrepareIssuedRepairResultBlocker(ctx, root, statePath, journalPath, request)`，
返回 opaque `IssuedRepairResultBlockerCapability` 或已经复核的历史 Receipt；
request 携带 exact packet/answer/owner/registered source root、原 platform source
verifier 与 caller CAS revision。capability 字段私有且不得序列化，不能存储可重入
lock context 供锁释放后复用。原 lifecycle Apply 将当次 context 显式传入 capability
校验，并重新核对 CAS state、原 source/PLAN、answer/message/artifact bytes 与 observed
HEAD 后只追加该 evidence 行；Agent pointer 与 transition 仍由原 lifecycle 处理。
message timestamp 从原 PACKET-ISSUE event 派生，ID/正文稳定生成，不含 caller revision
或重试时钟，避免同一答案重试得到不同 content address。自由 `evidence_refs` 只保留为
作者声明，不按字符串启发式将其认作机器 receipt；实际机器事实走原冻结来源与校验器。

验收须区分完整 CLI/Hook 的 race 与 receiver/CAS 的 race。若插桩的前置发题超出原
6 秒预算，保留该失败，不加长时限；可用同一生产源码快照正常构建的 CLI 子进程执行
原公开发题入口，再在 race 测试进程中验证接收/CAS/并发/恢复。子进程 binary SHA 与
生产源码清单必须绑定，测试不可使用不同源码的旧二进制；该拆分只提供 receiver race
证据，不声称前置 CLI 子进程或整条 Hook 链已经 race 通过。

#### Resolution artifact、接收和恢复

专用 Main authoring 文件路径由当前 packet 的 `AnswerPath` 精确派生：将终段
`<owner>/answer.json` 替换为 `<owner>/main-resolution.json`。它不是 Worker answer，
也不扩大 policy 对 Worker 的写权限。原 blocker 接收后可在 Main root 发布该精确
草稿；机器 tuple 从原 immutable blocker/message 派生，仅 rationale、evidence refs
和说明由 Main 填写。空 rationale 不可消费，已有草稿绝不覆盖。草稿发布失败不抹掉
已提交 blocker；重放可恢复相同草稿，不能再写 Runtime。contract_conflict 不发布普通
retry 草稿，仍走 S8 人闸。
该文件沿用原 `agent-message` 的
`blocker_resolution` envelope，并为 S9 subtype 增加可选、闭合的 `s9_resolution` 对象；
对已发行 S9 target，此对象成为必需，对 legacy blocker message 不增加要求。闭合
对象只保存 `decision`、rationale、evidence refs、当前 blocker packet/artifact ref+SHA
与必要的当前 Attempt tuple；不允许 author 自报身份、Actor、接受 revision/sequence、
lifecycle event ID 或自 SHA。这个精确文件是 Main authoring input。Hook 校验 bounded 原字节
和闭合 envelope，按原字节计算 SHA，再复制到不可变路径
`.claude/evidence/<runtime>/g<generation>/S9/blocker-messages/<packet_id>/<message_sha256>.json`；
该 copy 是唯一 lifecycle message source，原 `AdvanceAgent.MessagePath` 与
`Agent.blocker_resolved_ref` 指向 copy。可编辑 authoring path 只供诊断，历史接收不能
依赖它仍存在或仍相同；不为同一份 copy 再建独立 message writer/evidence artifact。
Resolution message 本身引用当前 blocker packet/artifact 与 decision，不引用自身 SHA。
原 CAS evidence row 与 journal `EvidenceIDs` 绑定前 blocker receipt、该不可变 message
ref/SHA；message ref 单向指向实际 bytes。Answer digest 与 message digest 不通过预测
Runtime revision 或循环引用计算。

消息 subtype 均位于原 agent-message **顶层**，保留原字符串 `body`，不把它改成
对象。`s9_blocker` 固定为 packet_id、packet_sha256、answer_path、answer_sha256；
其中 answer SHA 是 canonical 九字段专业答案摘要，authoring path 只是原答案位置，
历史 reader 不靠仍可修改的作者文件证明接收。`s9_resolution` 固定为 decision
（仅 `retry`）、rationale、evidence_refs、packet_id、packet_sha256、blocker_ref、
blocker_sha256、execution_assignment_id、execution_generation、execution_sha256。
原 packet SHA / execution digest 冻结其余 Runtime/Session/Plan/Contract/attempt，
Apply 必须复验这些事实；消息中不允许接受坐标、self SHA 或自报 Main 身份。
Contract conflict 的原 S8 路由不通过 `retry` 自动完成。两个 subtype 对 legacy
message 可省略；issued target 在原 Writer 中强制其确切 subtype 和私有 capability。

Typed `work_blocked` 同样只扩展原 `AdvanceAgent`：同一完整 S9 target slot（Runtime、
generation、round、Session、Plan、Contract、Repair 与 Manifest Assignment、owner、TASK）
已有任何历史发行 packet 时，泛用 blocker report 没有 capability 必须拒绝；即使 source
漂移令 fresh selector 不再选中，也不能退回 generic 路径。不同完整 slot 且未发行的 legacy
Agent 保持原行为。接收按不可循环的两步顺序形成事实。先由原 typed answer 与当前 packet/source
生成严格的原 `blocker_report` lifecycle message，只引用 packet/answer ref+SHA，不引用
尚未生成的 artifact；严格校验后按规范 message bytes 计算 SHA 并保存到
`.claude/evidence/<runtime>/g<generation>/S9/blocker-messages/<packet_id>/<message_sha256>.json`。
再生成 canonical blocker artifact，含完整 professional answer、不可变 message ref/SHA、
packet 与当前 source/attempt proof，按 artifact 全字节 digest 保存到
`.claude/evidence/<runtime>/g<generation>/S9/blockers/<packet_id>/<artifact_sha256>.json`。
因此 message→answer 是单向引用，artifact→message 是单向引用，没有回指环。原 CAS
的 `Agent.work_blocked_ref` 指向不可变 message；同一 evidence row 和原
`AGENT-LIFECYCLE` journal `EvidenceIDs` 绑定该 message 与 artifact。exact
artifact/message/owner/source replay 在当前状态与 caller revision 检查前核验完整原
receipt 并零写返回，异答案或异 message conflict；submission key 固定为原 Writer 的
`runtime:agent:<owner>:work_blocked:<beforeRevision>`，不是新 registry 或 SHA key。
历史 reader 对 packet issue key 与 blocker lifecycle key 分别检查完整 journal：
每个 key 只能有一个 outcome=committed 的事件，且须是预期 event ID/transition；
第二个 committed 事件即使引用其他 evidence 也构成冲突。失败或拒绝审计不计为
接受，不因其与成功重试共用 key 而抹掉原接收事实。
staging 走原 pending-safe cleanup/recovery：未决 marker 不删产物；只有原 Writer
recovery 判断提交结果并且精确 SHA 未被 state/journal 引用时才可清理。

Blocker artifact 的闭合 DTO 最少固定：schema/contract version、packet ID/SHA/issue
event、完整原 `RepairResultBinding`（含 Sources、PlanReport、Execution digest、
适用 Rework/Resume）、规范 registered execution root/generation/PlatformSessionID、
BaseCommit 与 observed HEAD、原 AnswerPath、完整 canonical 九字段 answer 及其
digest、不可变 message ref/SHA。以上 machine tuple 必须和原 packet 冻结的 execution
projection、原授权资料 ref/SHA 对照；不把自由 evidence refs 升格为机器证明。
artifact 不含自己的 ref/SHA、预测接受 revision/sequence、重试时钟或生命周期 event
ID/key；这些接受坐标由原 evidence/journal reader 重建。新接收在真实 working
PLAN 下验证；后续 blocked 状态的 proof 通过显式 blocked-state authorizer 复用原
PLAN/activation 校验，不复制或改写 state 来伪装 `working`。

原 evidence 索引只扩展两个 pipeline-owned kind：`repair_result_blocker` 与
`repair_result_resolution`，仅 v1.2 state schema 可容纳。两者均不能用
`runtime evidence add` 或恢复 import 单独登记，不增加任何 gate 的合格证据 slot。
Blocker 行的 id/path 均为 content-addressed artifact ref，sha 为实际原字节摘要，
produced_by 只含原 owner，baseline_generation/review_round 来自 packet，
responsibility_id 为原 `BUILD-WORK-PACKAGE`，scope_refs 是 packet ID 与
RepairAssignment ID 的无重复精确集合。Resolution 行 id/path 均为不可变 resolution
message ref，sha 为实际原字节摘要，produced_by 为 `main`，responsibility_id 为 null，
scope_refs 是 packet ID、RepairAssignment ID、前 blocker artifact ref 的精确集合。
初始 status=valid、invalidation 字段为 null。行、原 Agent pointer 与原 lifecycle
event 在同一 CAS 提交；历史 receipt 不因当前 gate status 变化丢失原接收事实，
当前 Stop 仍单独验证未 resolve 的唯一 blocker。原 `EvidenceIDs` 精确绑定所需的
message/artifact refs，无重复；不以顺序另造接受身份。

前 blocker 仍是历史。`blocker_resolved` 必须从原 `blocked` 状态发生，evidence
绑定 exact 前 blocker receipt 与新 resolution/message；Agent 的 `blocker_resolved_ref`
只作索引提示。后续 Resume lineage 必须从当前 Runtime/evidence 与完整 journal
验证唯一的 blocker、精确 generation/before-after revision/sequence/event ID/key、
`EvidenceIDs` 和 answer/message refs/SHA，并验证它发生在该 blocker 之后；不能只看
Agent 字段、当前文件或时间戳。新 S9 packet identity/provenance 包含前 blocker
packet ID/packet SHA、artifact ref/SHA 和精确 `blocker_resolved` event ID/key/sequence、
immutable message ref/SHA，使用新 AnswerPath。沿原完整 Runtime journal 的有界
64 MiB reader 从末端反向解析 lifecycle chronology；必须唯一匹配该前 blocker 之后的
resolve，拒绝丢失/重复/逆序证据和任何重复引用形成的 lineage cycle。Agent 上的旧
`blocker_resolved_ref` 只作候选索引，不能提供事实。新 packet 的 Resume tuple 必须
绑定到 current packet selector，并进入 candidate、final Result 与 completion provenance；
这些消费者重新验证同一前驱与 Resume digest，不能从旧 receipt 推断新执行。
resolution 原 CAS 通过后，旧 packet 必须立即退出 current selector，旧 answer 和原
产品 scope 同时停权；不能只依赖 Agent 被原 `blocker_resolved` 推回 working。
Resume successor 尚未发行时，责任投影明确等待 Main，不催 Worker 重交旧答卷，也不
由历史 blocker/resolution receipt 授 Stop、WorkComplete 或新产品写权限。
旧 packet/answer/receipt 不能复活为当前 Stop/WorkComplete/新接收；新 completed
completion report 必须晚于原 Begin、适用 Rework 与 Resume event。

#### Resume successor 实施约束（2026-10-15）

不新增责任状态、生命周期事件、Writer 或注册表。原 Main `stage-answer issue` 在同一
workspace lock 内由原 blocker/resolution receipt 派生后继。`RepairResultBinding` 仅新增
可选 `resume`（旧包省略该键，旧 identity 不变），闭合字段为
`predecessor_packet_id`、`predecessor_packet_sha256`、`blocker_ref`、`blocker_sha256`、
`resolution_message_ref`、`resolution_message_sha256`、`resolution_event_id`、
`resolution_idempotency_key`、`resolution_before_revision`、`resolution_after_revision`、
`resolution_sequence`。这些值全部由原已验证 receipt/packet 派生，不由作者填写。
原 Sources、PlanReport、Execution、Rework 精确相等才可接续；changed-generation/rework
沿原独立入口处理，contract_conflict 仍回 S8。resolution 公开 receipt 不新增必填字段。

原 packet builder 将完整 resume 纳入 identity，故产生新 AnswerPath；candidate/final
provenance 携带同一可选 resume 和其 canonical SHA256。唯一当前后继由原 packet registry
与 journal 验证，拒绝同一 resolution 的分叉、循环、缺项、反序及假 message；原锁/CAS
保证重放得到同一后继。尚未签发时继续等待 Main，只有完整验证的当前后继才解除
ResumePending。验证读取复用现有有界 journal，不扩大 Hook 时间或读取预算。

新 completion report 必须晚于原 Begin、适用 Rework、当前 resolution 以及后继 issue；
旧 completion、旧 candidate/answer、旧 Stop receipt 均不能授权后继。原候选入口仍拦截
未带 typed capability 的调用，但不能因历史前驱与当前后继共用 execution 就拒绝
唯一合法后继。后继再次 non-pass 时，原 typed work_blocked CAS 仅在完整后继 proof
证明其绑定上次 resolution 后清除 Agent 的旧 `blocker_resolved_ref` 索引；旧不可变
message/journal 全部保留。必须覆盖 blocked → resolved → successor → 再次 blocked，
不能让第一轮保留索引破坏第二轮。

#### Owner responsibility、回执 wire 与原 S8 人闸

typed blocker receipt 和 completed candidate 是独立事实。新 blocker 接收回执须明确
投影 `accepted=true`、`work_complete=false`、责任状态 `blocker_recorded`、
`consumer_owner="main"` 与 `consumer_pending=true`；它不是 delivery 或 professional PASS。
其中 `blocker_recorded` 只保留为原领域的 `responsibility_status` 接收事实，
通用 `status` 仍只有 AGENTS 的五态：仅当前原 blocker/source/blocked PLAN 证明成立
时投影 `blocked`。历史接收本身不证明当前阻塞，也不授结束或写入权限。
保存入口沿原 `SaveRepairResultFrom` 校验完整答案，再按专业 verdict 选择原
candidate 或 `work_blocked` 消费器。blocker 分支从 prepare 到原 lifecycle CAS
持有同一 workspace lock/context；只读 receipt 从原 journal/evidence 重建，
精确重放不得再写一次 Runtime。CLI Hook 保留相同真实 actor/CWD/session 绑定。
原 lifecycle submission key 必须保持
`runtime:agent:<owner>:work_blocked:<beforeRevision>`，由原
`AGENT-LIFECYCLE` Writer/CAS 产生，不创建独立 registry 或以 artifact SHA 代替幂等键。
共享 receipt schema 的 `responsibility_status` 与 submission-key allowlist 已在同一
接线 slice 中显式增列 `blocker_recorded` 与上述精确原 lifecycle key 格式；两者必须
双向匹配，并固定 accepted、work_complete、consumer 与 domain refs 的组合。
其他 status/key 保持闭合约束，不复用已有值伪装 blocker。
当前 exact blocker receipt 且 Agent 确为 blocked，才可独立投影
`ExecutionMayEnd=true`；这不要求共享回执 schema 新增外部字段，内部 responsibility
fact 可以承载。已 resolve 的历史 receipt 仍可 `accepted=true`，但绝不意味着当前
`ExecutionMayEnd`；Stop/Idle
必须从同一只读 Runtime snapshot 验证唯一、未失效的当前 blocker receipt/journal/source
projection。它同时投影 `DeliveryRecorded/WorkComplete=false`、`ConsumerOwner=Main`、
`ConsumerPending=true`，不等待全局 ACK，也不能显示任务完成。generic
`completion_report`、`one_shot`、manifest blocked 文本或单独 Agent.status 均不能放行。
typed blocker 闭合不能让 Main 集成或下游 S7 开始。

Main resolution 的原接收回执单独使用领域事实
`responsibility_status=blocker_resolution_recorded` 与原
`runtime:agent:<owner>:blocker_resolved:<beforeRevision>` key，两者在共享 schema
双向绑定。固定 accepted=true、work_complete=false、status=active、Main consumer
pending，并引用新 immutable message 与前 blocker。它证明恢复决定已接收，
不证明 Worker 交付完成；Resume successor 未发行时下一动作仍是等待 Main。

`contract_conflict` 一律不能在同一 Contract/Session/Plan 上调用普通
`blocker_resolved` 续跑。它保全到原 S8 人闸：必须有新 Contract 的原审批 evidence
或显式授权的 bounded review、精确审批源/SHA、消费记录、`S8-REPAIR-CONTRACT-APPROVAL`
journal proof，再重新建立 Session、Plan、Assignment/dispatch 与新 packet attempt。
现有 `ApproveContract`/`VerifyRepairPlanReportSources` 可复验这个原链；
`approved_by`/`produced_by` 字符串不是独立身份凭证，不能拿旧批准替代新 Contract。

#### 尚未闭合的恢复边界与实现状态

若机器 tuple 已 stale，旧 packet 不可接收也不授 Stop。当前没有可证明的 typed
Main operational-owner receipt/Stop 路径来接收因 source 已损坏、缺失或漂移而无法
满足正常 packet 校验的未决答卷；这项仍待单独设计。必须保全原 answer/message/marker
并输出具名恢复动作，不能通过 generic `work_blocked`、自由 blocker message、重新读
当前文件后冒称旧 source、或虚构 Main decision 接收。遗留无 typed packet 的 legacy
路径继续原样。

当前已新增严格的 `RepairResultBlockerArtifact` / 只读 Receipt DTO 与独立 artifact
Schema，并验证完整 non-pass answer、canonical answer digest、不可混入接受坐标、
两种原 execution ID 拼写及缺失/重复/未知字段拒绝。原 Agent lifecycle mutation
也已显式记录实际 baseline generation 并在 Apply 核对；generation 2/3 默认写成 1
的原故障已复现后修复。进展与范围见
[全局 §7.53](../GLOBAL-OPTIMIZATION-DESIGN.md#s9-blocker-foundation-progress)。

历史 receipt reader、原 message/evidence Schema 与 catalog、blocked PLAN authorizer
已完成定向 normal/race，见[全局 §7.54](../GLOBAL-OPTIMIZATION-DESIGN.md#s9-blocker-reader-progress)。
其机械 fixture 只验证历史 reader，不代表真实 typed producer 接受。
原 `AgentEventRequest`/`AdvanceAgentWithContext` 的 blocker 私有 capability 与同 CAS
接点已存在并通过直接 API 机械链验证。公开 Save/PostToolUse、闭合 receipt
schema 与同源 Stop/Controller proof 已接线，公开 Hook 与原 CAS 定向测试通过；
这仍是机械验证，不是 S9 真实模型整阶段验收。Main resolution authoring Hook、
私有 capability、同 CAS 的 message/evidence/Agent pointer、历史 receipt 与旧 scope
退役也已通过公开 Hook 定向机械验证；精确重放零写入，错误身份/更改重放拒绝，
contract_conflict 不进入普通 retry。Resume successor/lineage 已合入主源码，公开恢复、
第二次 blocker、原 candidate→Main integration→final Result 机械回归通过；旧答案
不恢复写权限，未发行后继时 Worker 等待 Main。各接点仍沿原 schema 与原
evidence/journal/lifecycle CAS，保持单 Writer、原 transition、原 S8 人闸和未发行
legacy 行为。该记录不代表真实模型 S9 整阶段或后续 typed consumers 已验收。

### 13.12 S9 ChangeImpact typed adapter over the existing PTR-BUG-05 (旧方案，停止实施)

本切片只将 Main 的专业影响判断接入现有 `CreateChangeImpact` / `CommitChangeImpact`
与 `PTR-BUG-05`。它是原 ChangeImpact 的阶段答卷入口，不新增领域消费者、Runtime
Writer、状态机或第二接收账本；也不扩展原 ChangeImpact 的领域接受规则。当前无
`Contract.impact_expectations`，本契约不新增它，不增加 unexpected-change 拒绝门、Plan
revision 或 S8 路由。L3-S9 §3.5/§6 的 expectation 对账、完整 scope/dependency graph、
迁移/兼容/部署分析仍未实现，不得以本 adapter 宣称 L3 全闭合。此节只固定后续实现边界。

#### Main 入口和冻结来源

仅在显式 packet 模式、当前 S9 `impact_reconciliation` 责任、当前 approved Contract/
Session/Plan、完整当前 Assignment Result 批次都可验证时发行 `s9-change-impact-v1`。
typed 路径要求批次中每个当前 Result 都有其原始有效 typed receipt；混有未带 typed receipt
的旧 Result 时不发行此卷，packet-off 仍使用原 `repair impact create|commit` 路径。
该责任由原 `review.repair.status=impact_reconciliation` 表示；原 lifecycle 此时仍为
`bug_resolution.fixing`，不新增同名 phase。Issue 路由优先检查原 repair pointer；
PTR-BUG-05 的 from cursor 按该真实原值冻结和验证。
Main 身份、平台 session 与 CWD 均来自原 Hook/可信运行上下文。Worker、命令参数
`--actor`、答卷字段或磁盘上同名文件不能取得 Main 保存权限。复用原 stage-answer
issue/save/receipt 与 Runtime Store 的 sealed-pending 恢复；只允许当前 packet 的精确
answer path。没有新 Store、Writer、Status 或接受旁路。

packet 必须从同一 Runtime snapshot 和原不可变源固定：

- packet body 固定 Runtime ID、锁定 REQ 与 baseline generation。packet ID/SHA 与原 issue
  event 的 before/after Runtime revision、idempotency key 和 journal sequence 作为外层发行
  回执从原 registry/journal 核对；不将 packet 自身 SHA 或包含自身 ID 的 issue event
  写入其哈希输入，不从当前 revision 或文件时间推算发行事实；
- 当前 Investigation/Case pointer 实际存储的 Case ref/SHA 与 Case ID，并验证它与
  approved Contract 的 `case_id` 相同。Case 路径/哈希只能取原 pointer，不按 ID 拼路径；
- 当前 approved Contract ref/SHA/ID，RepairSession ref/SHA/ID，RepairPlan ref/SHA/ID；
- 当前 Plan assignment 顺序中的完整 Result 批次。每项固定 Assignment ID、Result ref/SHA，
  typed receipt 中的 packet ID/SHA、answer SHA、owner、Runtime/generation、Candidate
  ref/SHA、accepted event ID、accepted idempotency key、accepted-after-Runtime-revision
  和 accepted sequence；读原 `RepairResultReceipt` / Result ref，不从 current Agent 标签
  或文件存在补造接收；
- 发行时由原 `ComputeSessionChangesetRecord` 生成并由原 `PersistChangeset` 保存的
  Changeset ref/SHA、Changeset ID、Session ID、digest 与完整 `artifacts[{path,sha256,status}]`。
  保存前后都重算 Session 实际 diff 并精确比较 path/SHA/status。现有 digest 仅对排序后的
  `path:sha256` 求哈希，不包含 status；status 必须另作 exact set 比较。Runtime 没有
  Changeset pointer，不能伪称该 ref 是当前 pointer；它是由原函数生成、packet 精确绑定的
  immutable artifact。相同 source tuple 重发先读取并验证既有 Changeset，保留原
  `computed_at`；若尚无文件才生成并独占发布。不能因本次时钟不同改写同 ID 文件，
  或把时间差当作新 packet/新来源；
- 当前 Runtime evidence ledger 的规范排序摘要。行只投影实际存在的字段：`id`、`kind`、
  `path`、`sha256`、`status`、`baseline_generation`、`review_round`、`produced_by`、
  `invalidated_by`、`invalidation_rule`、`invalidation_reason`、`responsibility_id`、
  `scope_refs`。旧行缺失的可选字段按原数据表示保留，不补造 `consumed_by`、`consumed_at`
  或 `dimension`。此摘要绑定发行时 ledger，不替代原 Commit 对当前 ledger 的验证。

若以上任何一项缺失、重复、source 不一致或无法读取，不能发行可填写卷；走各对象原有
恢复入口。文件名、外部自由文本、当前 cursor 或历史 receipt 不能补成当前 authority。

Inputs 用 UTF-8 字符串保存各原 JSON 源的完整字节，包括原缩进和末尾换行；原 SHA
按这些字节校验。不能把 `json.RawMessage` 嵌入 packet 后经 Marshal 的重排字节当作
原 artifact；packet 编解码后仍须保持原 pin。ledger 行使用同一规范字段投影与排序。
Changeset 既有路径只有截断 digest，复用还须核对完整 digest、当前 Session ID、
`source=session_diff` 与含 status 的精确 artifacts；跨 Session 或截断碰撞直接冲突，
不覆盖原文件，也不借同 ID 假定来源相同。

#### 专业答案、引用 allowset 与既有 ledger 映射

封闭 answer 仅暴露现有 ChangeImpact 专业判断：`change_types[]`、原结构
`decisions[]`、`evidence_dispositions[]`，以及可选 `escalation_level`。若省略后者，沿
原 `CreateChangeImpact` 默认 `assignment`。`decisions[]` 保持现有对象字段
`source_id,target_id,relation,rule_id,decision,responsibility_id,scope,rationale,
recovery_evidence`；Schema 允许的 decision 为 `invalidate|supersede|retain|reverify`，
`rule_id` 必须匹配 `IM-...`，`relation` 仍为非空字符串。`change_types` 只能取当前
`review-evidence.schema.json` 的 closed enum：`requirement|design|contract|task|
implementation|test|configuration|database|dependency|documentation_only`；旧示例中的
`regression_test` 不是该 Schema 允许值。这里不创建新的 Rule 或 Relation 字典。

引用按字段语义分开冻结 allowset，而非让所有字段共享一袋字符串：

- `source_bug_ids` / `source_case_ids` 由机器从当前有效源派生；至少使用 Contract 绑定的
  当前 Case ID，Bug ID 仅在当前源有权威登记时才附加，不能让作者填任意 ID；
- `decisions[].source_id` 只列 packet 冻结的当前来源对象 ID；`target_id` 只列当前
  Result/evidence、Contract unit/assertion 或 Changeset artifact 对象 ID；`responsibility_id`
  只列当前 Assignment/Runtime 登记的责任 ID，并保留 Schema 允许的 `null`；三者各自
  单独生成 allowset，不相互借用。`scope[]` 只列冻结 Changeset 的真实 changed path，精确
  文件路径本身满足原 Create 对每个实际 changed path 的覆盖检查。所有 ID/path 都须来自
  packet 当前闭包，不凭似是而非的 ID、路径相似或目录存在补建对象。`recovery_evidence[]`
  单独沿原 consumer 接受的形状校验：当前 Runtime evidence ID、当前 pinned S9 artifact、
  带 SHA-256 的内容寻址本地 `path:` 引用，或非空 `scheme://subject`。最后一种是原消费者
  认可的外部引用字符串，不代表本地存在或已解析的外部对象。Schema/领域代码没有 `relation`
  枚举或 `rule_id` 注册表，不能把它们描述为已存在的机器 allowset，也不能新增平行 registry；
- `evidence_dispositions[]` 每项只有 `{evidence_id, disposition}`，disposition 是
  `invalidate|supersede|retain`。每个 evidence ID 在答案中至多出现一次，可以不选择。候选 ID
  必须是当前 Runtime ledger 中已登记且 `baseline_generation` 等于冻结 generation 的行；
  `supersede` 与 `retain` 还必须在保存前为 `status=valid`（原消费者把缺失 status 按
  `valid` 读取）。原消费者对 `invalidate` 只要求行已登记且属于当前 generation，没有保存前
  status=valid 检查，adapter 不另加此门。机器将每项唯一映射到原 artifact 的
  `invalidated_evidence_ids[]`、`superseded_evidence_ids[]`、`retained_evidence_ids[]` 三组之一；
  答案不再重复填写这三组数组。

`decisions[]` 是逐个 source/target/relation 的专业分析；`evidence_dispositions[]` 是
ledger 动作，二者不互相复制、不从一个数组猜另一个。只有原 consumer 已有的交叉关系继续
生效：每个 retained ID 必须有 `decision=retain` 且 `target_id` 精确等于该 evidence ID 的
decision，并有非空 rationale；retain/supersede decision 保留原 recovery_evidence 要求。
不新增 supersede 与 decision 行之间的领域关联门。未选中的 evidence 不会被答卷隐式标为
retain；原 Commit 仍按 `impact.ComputeImpact` 对未列出的受影响 evidence 执行原默认失效。
该计算目前依赖 `scope_refs` 的现有投影，不是完整依赖图，也不证明 L3 的影响传播已实现。

`required_reverification_ids[]`、`source_*_ids[]`、`changed_artifacts[]`、`impact_id`、
Runtime/REQ/generation、Main `analyzed_by` 和 `analyzed_at` 全由机器产生，answer 不提供
这些身份或 receipt 字段。对每条 `decision=reverify`，按 `decisions[]` 中 1-based 序号，
对 `{packet_id, decision_ordinal, decision}`（其中 decision 是按结构字段 canonical JSON
编码的原 `ImpactDecision`）计算 SHA-256，形成 `reverify-<64位小写十六进制摘要>`；对结果
ID 字典序排序后写入原数组，空集合仍可为空。此 ID 沿用现有 `reverify-` 形状并由原
Handoff exact set 消费。原 ChangeImpact wire 没有 `occurred_at`；虽然旧 Request struct 有
同名参数，原 `CreateChangeImpact` 不持久化它，typed answer/provenance 不得添加该字段。

#### 原 PTR-BUG-05 Apply、回执、幂等与 legacy fence

adapter 将专业答案映射回现有 `ChangeImpactRequest` 和原 immutable ChangeImpact artifact，
再走唯一的 `CommitChangeImpact` / `PTR-BUG-05` Writer。ChangeImpact 增加可选、closed 的
`typed_provenance,omitempty`，仅 typed 路径出现，字段固定为 packet ID/SHA、原 issue event ID、
answer SHA、issue Runtime revision/sequence、当前 Runtime/REQ/generation、Case/Contract/
Session/Plan 的 `{id,ref,sha256}`、按 Plan 顺序排列的完整 Result receipt tuples、
`changeset:{ref,sha256,changeset_id,session_id,digest,artifacts[{path,sha256,status}]}` 和
ledger digest。Result tuple 的字段就是本节冻结清单中的 Assignment/Result/packet/answer/
owner/Runtime/generation/Candidate/accepted event/key/revision/sequence。关闭对象不接受额外键，
并逐项等于同 packet 与当前原始来源；`analyzed_by` / `analyzed_at` 保持 ChangeImpact 顶层原字段，
由可信 Main 与机器时钟写入。缺省时旧 artifact 的字段和序列化字节保持不变；不更换现有
schema_version。artifact 不嵌入自己的 SHA、PTR-BUG-05 接收坐标或预测 receipt。

typed private capability 只证明本次保存来自可信 Main Hook、精确当前 packet 与完整 answer；
它不授予新领域通过条件。在原 `PTR-BUG-05` Apply、原 Runtime Writer 锁内，typed 分支必须
重新验证 capability 绑定及完整冻结 tuple：当前 Case/Contract/Session/Plan、Result 批次和
各 typed receipt、Changeset ref/SHA 与重算 exact diff、evidence ledger 摘要、answer bytes、
Main authoring context 和原 authority-freshness。任一不一致则原 CAS 不写。不能以 `CommitChangeImpact`
现有锁前 preflight 取代 Apply 内验证；本要求只补 typed 来源一致性，旧未发行 legacy
`CommitChangeImpact` 的原规则和行为不顺手改变。Commit 继续执行现有 exact Result/actual
Session diff 比较、ledger 验证与 PTR-BUG-05 transition，不另造影响质量门。

一次提交后，原 Runtime pointer 的 `impact_ref`/`impact_sha256` 与 journal 中唯一已提交的
`PTR-BUG-05` event 是接收事实。receipt 精确重建原坐标：令 `beforeRevision = runtimeCommitRevision(expectedRevision, snapshot)`，
`afterRevision = beforeRevision + 1`；event ID 为 `evt-s9-impact-r<afterRevision>`，幂等 key 为
`runtime:s9:impact:<impact artifact path>:<beforeRevision>`，event 为
`change_impact_reconciled`，`EvidenceIDs` 恰为 `[impact artifact path]`；同时核对原
`PTR-BUG-05`、baseline generation、精确 from cursor 与 to=`bug_resolution/targeted_reverification`、
提交后 Runtime revision 和 journal sequence。
不建第二 receipt registry。相同 packet 和 answer 的精确重放只在当前 pointer 仍指向同 artifact
且上述原 journal event/receipt 唯一匹配时返回原 receipt，不再 CAS。相同 packet 改答卷是
conflict，不能覆写 immutable artifact。artifact 已创建但 current pointer/journal 尚无接收事实是
pending，交原 Store sealed-pending 恢复；不因文件存在宣告成功，也不以新事件重试覆盖旧状态。

仅创建 artifact、尚未进入 Writer 时可能根本没有 Store pending marker。这时保留并核对
同 ID artifact 的完整 canonical 字节（包括原机器时间），在原 phase 与冻结来源仍有效时
续作同一个 `CommitChangeImpact`；不能重新取时间生成第二份 artifact。已有 marker 才走
原 Store pair 恢复。两种情况均须原 CAS/journal 接收后才返回 accepted。

一旦该 exact target 已发行 typed packet，原 CLI/legacy create|commit 不能绕过 packet 对同一
current Runtime/Contract/Session/Plan/Result/Changeset target 接收；只能由原 typed save/capability
入口消费。未发行 typed packet 的 target 和 packet-off Runtime 均保留原 legacy API/CLI。旧
ChangeImpact artifact 缺 provenance 永远按 legacy artifact 读取，不能因 `analyzed_by=Main`
推定 typed 接收。

#### 明确未完成

此 adapter 不校验 Contract impact expectations（当前 Contract/schema 没该字段）、不要求
unexpected changed 触发 S8/S2 或 Plan revision、不实现 prospective/actual 三分分类、完整
authority/contract/module/data-flow dependency graph、scope reconciliation、migration/
compatibility/deployment decision 或完整 required-reverification assertion/owner map。
现有 `CommitChangeImpact` 消费的 `required_reverification_ids` 是 artifact 中 ID 的原样登记，
后续 Handoff 做 exact-set 检查；typed adapter 从 reverify decision 稳定派生 ID，但这不等于
实现 L3 的完整检验矩阵或证明下游所有 assertion/owner 均已安排。以上是独立 domain/schema/
consumer gap，不能在本优化切片中伪装成 packet 完整性或悄然新增拒绝规则。

### 13.13 S6 canonical Hook assignment 与工作顺序（2026-10-15，窄补充）

canonical Hook 的 assignment 坐标必须从同一次 Runtime snapshot 投影当前 exact Record 与
Execution 的绑定，并核对 owner、Record ref/SHA、Runtime/baseline generation 和
`ready|complete` 状态。
Manifest 只表达计划坐标，不能覆盖之后登记的 e3 Execution，也不能使已退役 Execution
恢复为当前。Runtime Record 或 owner/ref/SHA/generation 不匹配时 fail closed；绑定工作区尚无当前
Execution，或 execution 为 preparing/退役时清空坐标。二者均不从 sidecar、manifest
或目录内容 fallback 猜 owner/execution。未绑定的旧实例保留原 manifest 投影。

正常 Builder 顺序是 PLAN → 原 begin → 实现与本人 checks → 原产品 commit → Result/report →
Main integration。报告只记录已报告事实，不授予后续产品写入或 commit 权限。遗漏原 commit 时，
保留原失败 checkpoint。canonical S6 由 Main 以原 `register-workgroup --revise-assignment`
及具名 intent、fresh manifest/Team/Agent 发行新 Record；不在同一 Record 内重开，
不将 superseded Result 作为当前有效历史重新接收。原 `workspace rework` 仅保留其
legacy/S9 契约；环境恢复仍可复验原 integration。没有新 attempt 状态或历史授权，
不回退 lifecycle，也不放宽 commit 屏障。

### 13.14 Native Bootstrap 与集成 clean/cleanup（旧归档方案，停止实施）

本节仅保留设计历史。按全局设计 §0，优先将执行资产与产品工作树分离；不继续
归档 manifest、额外来源分类与归档恢复协议。原权限、Worker CWD 与 Main 检查保持。

产品 clean 与系统安装资产分别按原来源证明判断，不增加状态。未绑定的旧执行保留
原检查。已绑定 Main 复用原 `ControlFile` / `CleanInputs` 与 `AllExecutions` 的精确
工作树坐标分类；不能忽略整个 `.claude` 或任意嵌套仓库。Worker 从原 pointer、
同一 Main Runtime、当前 exact Execution 和 `BootstrapSHA256` 复验原 BootstrapReceipt，
并沿原 `ValidateExecution` / `ValidateBootstrap` 核对路径、字节与 Git 来源。

只有未暂存的 ` M` 或 `??` 且路径与当前 receipt.Files 精确相等、SHA 仍相符时，
才将它识别为系统资产；receipt 自身必须与 Main 固定 SHA 相符，pointer 必须等于
同一已登记执行的机器元组。产品 dirty、未知控制文件、staged/rename/delete、链接、
asset/receipt/pointer 漂移均拒绝。Inspector 的序列化 allowlist 不是权限；Integrator
持原 workspace lock 时重新验证当前来源。R3 产品未提交仍拒绝，R4 系统 overlay
不再冒充产品修改；实际 scope、diff、检查与接收门不变。

Git 的非强制 remove 仍不能删除 tracked overlay。仅在原 verified、Main ACK 和
`cleanup_pending` 已证明、原 source HEAD 不变且无未接收产品变更时，允许原 cleanup
把该执行目录原子归档到当前 checkpoint 下由 execution generation/source SHA 派生的
固定目录，保留源码、bootstrap 与本地提交资料。归档前独占发布 manifest，绑定原
checkpoint、Record、Execution、source HEAD 和 bootstrap SHA；目的路径不得覆盖。
同时短尝试并持有原 `workspace-launch/<runtime>/g<baseline>-<assignment>-e<generation>.lock`
到归档与登记清理结束；launch lease busy 时保留 `cleanup_pending`，不凭 session ID
或 Result 接收推断进程退出。不新增 PID/会话状态；锁序沿原 workspace lock → launch lease。
归档只是证据保存，不能作为新执行或 Git 权威。之后继续原 targeted、非 force 的
`git worktree remove` 清理旧登记，不 prune 其他目录、不强删或提交系统资产。

归档前失败保留原目录；归档后失败保持原 `cleanup_pending`，重试使用既有“原路径
缺失、原登记尚在”的恢复分支，不重复归档或合并。原 verified/check/ACK 事实不因
归档重写，Runtime Execution 仅在原 durable complete 后投影 complete。未知差异始终
保留待诊断。验收须真实 bootstrap 覆盖 tracked launcher/settings/role，并覆盖产品与
系统漂移拒绝、原检查/非 squash merge、归档中断恢复和非强制登记清理。

## 变更记录

| 日期 | 版本 | 变更 |
| --- | --- | --- |
| 2026-10-15 | 收缩决策 | 按用户纠正，取消 Hook 身份/权限/scope 拦截；答卷收缩为履职清单兼报告，保存触发原消费者核验与阶段推进；停止独立 Impact 与 Bootstrap 归档扩展。 |
| 2026-10-12 | v0.5.0 | 初始阶段答卷设计；未实现 |
| 2026-10-13 | v0.6.0 | 按无设计上下文执行者补齐自描述 JSON、独占分卷、保存消费、明确回执、冻结/幂等/恢复、人闸与原领域适配；附隔离 S5 原型，生产接线仍未实现 |
| 2026-10-13 | v0.7.0 | 固定生产发包/索引版本、独占输入、权限接线、同身份唯一接收、具名修订和跨事务followup去重；进入实施 |
| 2026-10-13 | v0.7.1 | 补冻结authoring字节、规范answer摘要、精确信封绑定、重放完整性与control结果/转换事件同事务契约；接收和控制基础已实现，保存入口仍待接线 |
| 2026-10-13 | v0.7.2 | 固定实际派发的 S5 进入/代数标记、锁内原计划授权、完整独占发布、回执与门资格的区别；共享 Main 接点已编码，Worker、修订与完整验收待完成 |
| 2026-10-13 | v0.7.3 | 复用原 execution registry 接独立 S5 Worker，冻结 view/资料及锁内源校验，停用 Main 双入口，按本人责任守 Stop；真实注册/激活/原门/保存 wire 验证通过，修订与正式启用待完成 |

| 2026-10-13 | v0.7.4 | 实现具名修订的发行暂停、接收替代、唯一 lineage、历史回执和独立 execution 新代；修复原 Apply 丢失第二 reviewer 资格的提交窗口缺口，普通验收通过 |
| 2026-10-13 | v0.7.5 | 新安装/init 1.2 与显式 --packet；固定原 Writer 格式和重置/pending 恢复，初始化前校验及字节冲突预检查，native 安装与真实旧写入器拒写验证；真实 Claude 未登录 |
| 2026-10-13 | v0.7.6 | 固定已发行索引的完整 pair checkpoint、原 RR-A 审核恢复、防回滚、源与 launch lease 校验；进入实施，结果见全局设计后续记录 |
| 2026-10-14 | v0.7.8 | S9 typed candidate→Main 实际检查→原 Result→ACK/cleanup 原链及定向并发/sealed-pending 恢复通过；固定已收 Result 重入与 Main target 漂移的事实保全，非 pass/后续消费者和真实平台仍待完成 |
| 2026-10-14 | v0.7.7 | 固定并验证 S7/S8 与 S9 PlanReport 原消费者/回执/恢复；PlanReport 冻结包已临时安装验收；新增 §13.10 的 S9 RepairResult 原候选→Main 实际检查→原 Result 契约，分层接线中，非完整 S9 完成声明 |
| 2026-10-15 | v0.7.9 | 固定 packet-enabled S5 的 accepted S4 formal input/subject 闭包、legacy 兼容和共享 typed packet subtype 路由规则。 |
| 2026-10-15 | v0.7.10 | 固定 S9 typed ChangeImpact 对原 PTR-BUG-05 的来源/专业字段/幂等边界（未接线）；S6 当前执行坐标和 commit-before-report，canonical 修正复用原具名 Assignment revision。 |

### S5 revision 守卫的共享 registry 边界（2026-10-15）

S2–S9 各 typed packet 共用原 `work_packets` registry。TR-003 的具名 S5 revision
守卫仅解码 `s5-document-review-v1`；其他 subtype 交回各自原消费者，不能误用
S5 lookup 阻断已满足的门。坏 S5 row、重复 identity、未接收/非 PASS revision
继续 fail closed。R11 双 reviewer 真实 PASS 暴露该分派错误；修复必须先通过
混合 registry 与原 revision 负例，再由冻结原状态上的正式 recover 验证。

### S3 作者字段形状（R11 真实反馈，2026-10-15）

题目直接声明 `answer.applicability` 为含 FE、BE、SYNC 三键的对象；每个值为仅含
`disposition` 和 `rationale` 的对象。disposition 只能 required/not_applicable，
rationale 由作者给出具体理由；机器不预填专业判断。R11 原题缺少字段形状，导致
两次不必要的试错与项目中不存在的 schema 查找。修复进入下一 fresh template；
正在恢复的 R11 链保留已发行 template identity；安装与外部 native CLI 的版本边界见接续证据。

### S6 read 类别展开（R11 真实反馈，2026-10-15）

`locked contracts` / `locked TASKs` 是投影类别，不是文件名。Next/read_order 与
诊断采样复用同一个只读展开：从当前 generation 的 documents[] 选择精确 kind、
locked 状态的实际 path，排序去重。未注册类别保持未解决，不能从目录存在推断
正式输入。已有 Runtime 的旧文字读单也使用同一展开；不扩大权限、正式 SHA 或
4 MiB/256 entry 预算，缺失真实文件继续报告。

### Native 执行准备入口（R11 S6 实测，2026-10-15）

已注册的 native Worker 统一从原 `runtime workspace prepare --assignment <id>
--agent <owner> --root <Main>` 准备，再使用返回的原 launcher。原 registry
Plan/Materialize 拥有实际 checkout/branch；manifest 的计划坐标不是已存在执行的
证明。不要把 legacy `runtime worktree-create` 串到 native prepare 前；它的
已有 checkout 校验在计划路径尚不存在时只能拒绝，不能删除校验后接受缺失目录。
AGENTS、S6 协议与 dispatch skill 使用同一正常入口，保留原 legacy CLI 兼容语义。

### 原工作区 advisory 与 CleanInputs（R11 S6，2026-10-15）

原 Hook 的 `.claude/worktree-reminders/<64位小写hex>.txt` 及同目录原原子发布
临时文件 `.delivery-<十进制随机数>` 是诊断 control bytes，不能因自身生成阻断
下一原工作区 prepare。沿原 ControlFile 分类识别这两种形状；不忽略整目录、
不提交/删除诊断来绕过，不改变 `.claude` 其他真实输入及产品 dirty 检查。

### PLAN 专业断言与接收后置事实（R11 S6，2026-10-15）

Main 注册前将 `done_when` 表达为交付物/业务行为的可检验条件；提交、接收、集成仍由
既有流程及原回执证明。Worker 的 PLAN assertion/oracle 必须能在本次报告接收前用
实际产物或已发生检查验证。待接收的同一报告不能证明“本报告已被接收”。
若已接收 PLAN 把原 receiver 的接收事实当作自身前置断言，走既有具名修订，
不删除已收断言、不白名单化待收报告、不伪造或提前引用未来回执；closing-contract
已有 DoneWhen 覆盖约束不因此跳过。该区分写入共用 AGENTS 与 PLAN 作者引导，
不新增状态、字段、Writer 或语义猜测校验器。
canonical S6 Builder 的 manifest 作者引导同时说明此 DoneWhen 边界；QA 等其他
领域的收尾条件保持各自契约。已规范包可作结构相等比较，但该比较不替代
读取器的 Schema/重建、发行来源、当前 slot 和真实调用者校验。

### S6 声明读集的原 Git 来源冻结（2026-10-15，实施约束）

新 native S6 execution 的原 canonical AssignmentRecord ReadPaths 必须在 Plan 前
展开为精确文件并进入 Execution.Inputs 和原 Worker inputs。每项来自绑定开发分支
的同一个 pinned commit 的 regular Git blob；checkout 字节必须与其相等，拒绝
缺失、链接、越界或不一致。目录/glob 按既有 path matcher 作有界具体化，总计
最多 256 entry / 4 MiB；不允许误用诊断 walker 作为正式来源。已有 scope、Agent、
session、激活、checksum 与原 Writer/Plan/Materialize 不变。原已准备 execution
缺少声明来源时仍按原 Plan 拒绝，不补写其冻结 SHA 或 bootstrap。旧 manifest/
legacy source 模式沿原契约，不以源码更新静默升级现有 execution。R11 缺失
model/state 的旧执行保留原反例及 pinned worktree 阅读轨迹；新准备须独立验证。
