# L4 — 阶段答卷

> 层：第四层｜机制域：每个阶段一份答卷，判断或状态变更如何进入既有门禁
>
> 上游：L1 D3/D4/D5 与公理二、公理五；L2 各阶段出口；[L4 运行时控制面](L4-runtime-control-plane.md)；[L4 revision](L4-revision-usage.md)；命令链、对外名字、重复载体和占位符的收敛见 [L4 手续收敛](L4-procedure-reduction.md)
>
> 状态：v0.7.0 实施契约已固定，生产接线进行中。未完成的入口不宣称已自动化。另有[隔离的 S5 交互原型](../../tools/stage-form-prototype/README.md)，仅验证填写与回执契约，不证明真实 Hook、Git、身份、领域门或崩溃恢复已经接线。本文不新增状态机。

## 1. 机制

一个需求的每个阶段只有一份逻辑答卷。进入阶段时，系统按注册的 JSON 模板生成当前责任的题目，把填写路径放进 `next`；Agent 不查找外部样例或自己补占位符。并行分卷与汇总见 §6。

设计要消除的失败是：工作已完成，但执行者不知道还要造信封、抄哈希、选事件、登记证据或调用哪个命令，于是一直在原阶段猜手续。验收对象是一个没有模板设计背景的执行者，它只拿到发卷说明、必要工程资料和 JSON，就能正确交卷。

正常路径为 **领卷 → 做工作并填写 answer → 保存 → 系统接收、求门、给回执**。完整且原门满足时自动通过；有发现时进入原失败路由。机器已知道的字段和登记手续不得成为人工题目。

卷区分两种内容；人工决定另按 §9 处理：

- **要判断。** 空位已经摆好。工作做完后，Agent 把答案写入这个文件，写入就是交卷。
- **只记状态。** 已有机器记录足以证明的事实由系统填写，不再让 Agent 抄写。文件状态不能代替人锁定、绑定授权或发布决定。

然后走同一条路：Hook 按文件路径计算指纹、记下内容，再跑这个阶段原来的门。门通过才进入下一阶段，并生成下一份卷。门不过就停住，回执说明缺什么，要改的仍是这一份。

Agent 不填写哈希、代数、身份和事件名，也不另跑提交命令。

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

当前 settings 的 PostToolUse matcher 是 SendMessage/Agent/Task/SubagentHandback/Bash，不含 Write/Edit；`internal/cli/run.go` 的 PostToolUse 也提前进入观察分支。只生成 JSON 或扩大 matcher 都不足以实现保存消费，须同步注册、payload 路径提取、consumer、官方 additionalContext 回执和端到端测试，按 [Hook 接线](L4-hook-platform-wiring.md)治理。

主触发覆盖实际支持的 Write/Edit/MultiEdit 与能提取真实目标的受控写入 adapter，只消费已登记当前卷。PreToolUse 守权限，不消费尚未落盘的新答案。PostToolUse 无法撤销过去的写入，拒收只表示未形成领域完成事实，保留草稿并给修正回执。

SessionStart/PreCompact 有界补消费当前输入与下一卷投影；子代理返回由 Main 接收保全结果。Stop 判断持久接收事实，已完成者不因系统 pending 或别人未交卷反复唤醒。修改 Stop 接线仍须守现有 preflight 契约，不能让被阻止的 Stop 先提交迁移。

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

### 13.2 实际接线和权限

register-workgroup成功及真实SubagentStart/SessionStart按当前合法责任发行/恢复；没有真实身份不给签名题。发卷不取代原readback/activation/计划屏障。Main与Worker的Write/Edit/MultiEdit PostToolUse在最后成功保存消费，不依赖下次调用。Main普通工具仍走纯observer。

PreToolUse仅对当前真实actor已发行的确切答案文件提供控制面输入例外；通过规范化根、物理包含与无下级symlink检查，不放开整个submissions或.claude。工作包/Runtime/汇总不能借此变可写。S7 reviewer与S8/S10 phase规则同步接例外，其他owner/旧attempt/未发行文件和指向仓库内另一owner的symlink拒绝；消费端再核对binding，PostToolUse拒收不撤销保存、不deny。

### 13.3 一次接收与具名修订

submission_identity=(runtime,generation,packet,slot,owner,attempt)，其key不含答案摘要；content hash单独记录。原RecordEvidence的Writer事务内原子追加document_review与submission索引；同身份同内容重放原记录，同身份不同内容content_conflict。冻资料漂移、跨owner、旧stage_entry及非法结论不登记完成。原producer独立性及门检查不变。

首次接收attempt不可变。S5的RecordDocumentReviewRevision必须在同进入身份/冻结链/原owner上，由Main记录具体原因并发行新attempt绑定唯一previous receipt；接收时同事务失效旧evidence/登记新evidence，旧envelope不改。并发同previous只有一个接收。离开S5或规格改变沿原返工/需求路由，不就地追认旧结论。

### 13.4 至多一次控制接续

domain事务记录stable followup_event_id、origin state/phase/phase_revision和control_pending。ControlRequest/transition.Request携带这组事实，Apply的原Writer事务内核对未处理与origin相同，并把转换及followup结果一起提交。已有request_id或按transition/revision的key不能代替这项去重。

转换已提交但回执/外层响应丢失：恢复只重建receipt与当前包；cursor已被其他事件推进：标replaced，不能把旧交付应用新阶段。not_ready记evaluated_no_transition，未来新交付有自己的事件；unknown只对原origin有界重试。旧Runtime/重入/同答案重复保存不重复迁移。长测试/合并归Main/Integrator，Hook仅短事务。

### 13.5 完成回执与验收

回执包括accepted、responsibility_status、work_complete、current_stage、domain_refs、waiting_on的reason/owner/ref、当前recipient next_action和JSON Pointer字段错误。work_complete只表达本人交付已接收，可包含合法Finding；阶段PASS、Main集成和人授权另判。系统pending与同事未交不重复催本人。

必须用真实Writer/领域门证明：并发两个owner不丢答案；同attempt异内容冲突；domain提交后/转换提交后故障恢复；两PASS/FIX/REQ-change原路由；原权限、计划、独立职责全集、人闸及冻结边界。安装和官方wire测试后再独立审查；真实Claude实测另记版本与结果，不用fixture替代。

## 变更记录

| 日期 | 版本 | 变更 |
| --- | --- | --- |
| 2026-10-12 | v0.5.0 | 初始阶段答卷设计；未实现 |
| 2026-10-13 | v0.6.0 | 按无设计上下文执行者补齐自描述 JSON、独占分卷、保存消费、明确回执、冻结/幂等/恢复、人闸与原领域适配；附隔离 S5 原型，生产接线仍未实现 |
| 2026-10-13 | v0.7.0 | 固定生产发包/索引版本、独占输入、权限接线、同身份唯一接收、具名修订和跨事务followup去重；进入实施 |
