# S8 机制审查与减负方案

> 2026-10-17 审查历史与本轮限定范围收尾。
> A/B/C/D1 已实施；after-r56 定向回归及 Linux amd64 宿主真实调查证据见 §5。
> 优先契约：[GLOBAL-OPTIMIZATION-DESIGN.md §0](../GLOBAL-OPTIMIZATION-DESIGN.md)。
> 复用原领域对象与消费者，不新增协议、状态机、Cache、独立发行或接收回执。

## 1. 审查结论与范围

普通 Investigator 已有单份专业 HypothesisResult 路径：只填 `answer`，保存后由原消费者接收；
普通 `plan_checkpoint` 不再另交 generic PLAN_REPORT，不手动 issue/consume。
证据：`agents/investigator.md:27`、`:72`；`internal/workpacket/hypothesis_result.go:208`。
最初审查发现 Main 专业入口、首次/恢复提示、复杂题形和监听重复工作的缺口，历史证据保留在 §3。
本轮已完成这些减负切片及限定验证；不表示全分支、全平台或全阶段验收完成。

Hook 只监听、提醒和流转。当前正常保存传入 `authorize=nil`，并非仍有 Hook 身份鉴权拒收：
`internal/cli/stage_answer.go:658`；`internal/submission/hypothesis_result.go:62`。
owner/source/Case/dispatch 的匹配属于原领域消费完整性，不能连同工具 ACL 一起删除。
当前 Pre 已有直接允许工具调用的路径；本方案不重写权限，也不机械新增 `s8GraphControlForHook` 特例。

## 2. 已固定的最小专业工作流

| 原工序 | 专业责任与填写内容 | 系统带出与流转 |
| --- | --- | --- |
| S7 交接与 intake | Main 阅读 immutable Findings、边界与 capture gaps，判断分组理由 | 给出 sealed Batch/Finding 实际路径、当前 Case 和既有缺项 |
| 竞争假设与派发 | Main 提出可证伪问题、判别条件、来源与停止条件，分派专业调查 | 复用原 register/dispatch；已有 ID、来源和绑定不让作者重抄 |
| 假设调查 | Investigator 从原现场出发执行 discriminator，填写真实观察、证据、专业判断及未解释项 | 给准确 view/answer 路径；保存调用原 HypothesisResult 消费者 |
| 因果汇总与 route | Main 汇总竞争解释、因果链、blast radius、detection gap，选择原 Case route | 按当前工序渐进展开原字段，复用原 Case/route 消费者 |
| Contract 审阅与批准 | Main 完成修复不变量、完整范围和断言；原授权主体审阅最终稿 | 审阅前投影当前机器字段；原 approve 绑定最终真实字节与授权 |
| 出口 | Main 填写当前路由所需专业判断、处理真实人决策 | 系统在满足原条件后调用原 consume/交接，按原 graph 进入 S9、S2、fresh S7 或保持调查/等待 |

Main 本来就负责分派专业调查；平台真实 Agent spawn 仍由 Main 通过原平台能力完成。
Hook 不代替专业分派，不伪造进程已启动，也不成为跨领域总控制器。
Main 专业问题沿已有工序展示，复用原对象与消费者，不另建一份主协议或第二份汇总报告。
正常 Main 路径也不手动 issue/consume、不拆专业内容另拼登记命令；保存不构成真实审批，获有效原授权后才调用原批准事务。
个人交付完成与整个 S8 出口分开判断；统一工作五态仍为 `pending/active/blocked/done/cancelled`。
Case 的 `investigating/contract_approved`、Hypothesis 的 `supported/refuted/inconclusive` 是领域事实。
这些原领域事实不应被当作新增工作状态，也不因五态统一而被删除。

## 3. 审查时历史：五类确定缺口与最小修正

本节保留实施前的静态审查快照。源码行号、缺口及“最小修正”均指审查时版本，
不作为当前仍有这些缺陷的断言；实际落地和验证范围见 §5。

### 3.1 Main 入口与尾段仍要求解释机制和搬运元数据

- `internal/cli/stage_answer.go:1435`：没有注册问卷就要求恢复 Case/manifest/TASK/dispatch，正常首次入口也落此文案。
- `internal/cli/investigation_command.go:1410`：register/dispatch 的下一步仍使用 Case/Hypothesis/Assignment/Agent 占位符。
- 同文件 `:546`、`:592`：route 读取三个独立 JSON 文件；`:664` 又要求独立 Contract 文件、hash 和批准命令。
- 同文件 `:961`、`:992`：intake 生成的 Contract placeholder 固化早期 Case revision。
- `internal/investigation/contract.go:164`：批准要求 draft revision 等于当前 Case revision；期间假设、结果、route 都会推进 Case。

最小修正：正常入口展示当前专业问题及实际资料；无卷先区分未登记、待派发与真实准备故障。
Main 的因果/范围/断言只填一次，按原工序投影进 Case、route、Contract 所需内容。
系统在实际审阅前带出最终 Case revision、exact Finding set 等机器字段，给出实际稿件路径。
审批绑定最终真实文件字节及其 hash；审批后不得再自动修改同一获批稿以“补齐元数据”。
原 Contract 领域校验和真实授权继续保留；不让 Main 为满足消费者手抄当前 revision。

### 3.2 题形、角色说明和恢复动作仍不一致

- `internal/schema/assets/stage-form-s8-template.json:16`、`:36`：边界和新假设示例仅解释要求，没有合法 JSON 值/对象。
- `internal/schema/assets/stage-answer-s8.schema.json:90`：新假设实际要求六组字段，含嵌套 `expected_outcomes{support,refute}`。
- `internal/workpacket/hypothesis_result.go:642`：草稿默认 `new_hypotheses=[]`，作者看不到非空项形状。
- `agents/investigator.md:57` 与 `:102`：前段禁止 Worker route/register，后段却给出 Main 的重入、登记和派发动作。
- 同文件 `:4`、`:86`：禁用 Edit 与建议 Write/Edit/MultiEdit 保存相冲突；`:23` 仍让作者核对来源 hash。
- L3-S8 `:298`、`:695` 仍要求新假设带 Assignment，与问卷机器生成身份的正常路径口径不一致。

最小修正：在既有题干/example 内给合法数组、完整新假设对象和冻结 Finding JSON Pointer 示例。
只让作者选择真实使用的证据与边界；系统核对已知来源，不能自动预填专业结论。
Investigator 只收到本人调查、专业阻塞和保存动作；Main 重入/派发说明明确归 Main。
保存说明统一为编辑并保存准确路径，兼容已有脚本/扩展工具监听，不限定某一种编辑工具。
阻塞填写位置仍待验证：用现有专业文本字段说明已做工作、缺口和 Main 下一动作，不新增 blocker 协议。

### 3.3 准确路径在启动、已接收和异常反馈中可能丢失

- `internal/cli/run.go:3116`：SessionStart/SubagentStart 中非空通用 Assignment context 覆盖 stageAnswerContext。
- `internal/cli/assignment_context.go:403`、`:450`：覆盖内容只有 Assignment sources/hash/状态，不含 S8 exact answer。
- `internal/cli/stage_answer.go:1481`、`:1487`：过期、receipt 异常和已接收反馈未附已有问卷准确路径。
- `internal/hook/stopidle.go:911`、`:946`：恢复动作只指向“前文 questionnaire context”的路径。

最小修正：已有 S8 问卷时，进入/恢复/收尾上下文保留其准确 answer 路径，通用资料摘要不能覆盖它。
异常也附可解析的实际路径；确实未生成或来源不可解析时，明确准备缺口和 Main/系统处理方。
本人已接收时说明本人可收尾及 Case 后续责任，不再让作者搜索路径或重复交付。
这是静态分支可确定的投递缺陷；尚未在本轮真实宿主复现，不声称已复现超时。

### 3.4 提醒触发通用恢复、全量发行与已接收答案重送

- `internal/cli/stage_answer.go:766`、`:780`：SessionStart/PreCompact 先走通用 New/Recover，再确定 S8 专业上下文。
- 同文件 `:1417`、`:1431`：Main 的 SessionStart/PreCompact/每次 PreToolUse 都调用全量 `IssueHypothesisResults`。
- `internal/cli/questionnaire_observer.go:239`：S7 已接收项跳过；S8 没有同等跳过，`:86` 只比较最初 draft。
- 因而已接收且偏离 draft 的 S8 文件会在后续普通 Bash/MCP 观察时再次送往原消费者。
- 原消费者在 `internal/investigation/hypothesis_result_receipt.go:100` 对同答幂等、异答 conflict，当前没有旁路。

最小修正：提醒读当前领域事实和问卷；仅在原责任确有准备缺口或变化时调用必要的原准备动作。
未变的已接收答案不重送，判断复用原 receipt/当前答案事实，不增加独立缓存或接收索引。
发生异答保存仍交原消费者并返回 conflict；恢复不能静默覆盖第一次已接收答案。
脚本与扩展工具保存继续走已有监听和共用消费者；Stop/Idle 同次完整判定只做一次。
目前只能证明重复调用结构，不能据此宣称本轮测得延迟、超时或性能收益。

### 3.5 原消费者前后存在重复读取与校验，删除边界待验证

- `internal/submission/hypothesis_result.go:93`、`:97` 读取并验证 authoring 文件。
- `internal/investigation/hypothesis_result_receipt.go:87`、`:91` 随即再次读取并验证同一输入。
- `internal/investigation/case_workflow.go:733` 与 `:857`、`:875` 分别在预检和 CAS 内调用 validateCurrent。

最小修正：后续先区分纯题形计算、可复用的已读字节和必须在 Writer/CAS 内重新证明的当前事实。
只删除经窄回归证明冗余的预检；保留 CAS 必要复验、原幂等、并发冲突和 pending 恢复。
本文不直接判定两次检查完全等价，也不承诺直接移除某次读盘即可安全。

## 4. 专业质量与授权保留项

必须保留 immutable Findings、原现场边界、可证伪竞争假设、真实 discriminator/evidence 和未解释项。
因果闭环、blast radius、detection gap、修复不变量、完整范围及原断言不能因表单减负而省略。
显式 `plan_approval_required`、真实 Contract 审批/有效原授权和 REQ 人闸继续执行。
原 Writer/CAS、来源完整性、同答幂等/异答 conflict 和原 route/consume 恢复保持。

## 5. 本轮落地与限定验证

| 切片 | 已实施范围 | 保留边界 |
| --- | --- | --- |
| A | L3 顶部固定最小工作流；角色、技能与原九题模板同步准确路径、合法 JSON 示例及现有字段阻塞说明 | Investigator 只交本人 answer；不加字段或第二份报告，Main 仍负责真实平台派发 |
| B | 启动上下文保留专业路径；listener/observer 复用当前领域事实，跳过未填写及未变已接收内容，补齐恢复与收尾路径 | 不增缓存或工具 ACL；异答仍交原消费者冲突判定，普通工具不跑完整业务链 |
| C1/C2a | Main 按 intake→假设/派发→因果/route 渐进填答；保存复用原 Ingest、register/dispatch、route/consume | 已知元数据系统带出，不手工 issue/consume、拆三 JSON 或重复专业内容；不新增全结果 PASS 前置 |
| C2b | Contract 专业填写投影原 schema 稿件；保存时校验，审阅前固定字节；匹配原已注册 human_decision 或有效原授权后调用原批准事务 | 保存不批准；普通提醒不重跑完整 Contract 校验，已审稿不覆盖，原授权/CAS 保留 |
| D1 | 删除 submission 外层紧接原领域消费者的重复 ReadFile/题形预检，复用原错误字段路径 | 原消费者全检查、CAS 必要复验、幂等及 receipt 保留；未改领域 validateCurrent 并发逻辑 |

冻结候选为 **after-r56，Linux amd64 host-only**；[源码与包清单](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/after-r56-source-inventory.json)
记录包 SHA256 `0c07c4711afe916d56807e3e936078a36777a6f3f910dbf5e48863b58ba17576`。
[定向回归](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r56-s8-combined-final.json)
为 **59 项不同顶层测试 PASS、0 fail、0 skip，154.619 秒**，不与早期切片测试重复累计。
当前 Go/schema 与冻结 R56 一致。

[R56 最终收尾报告](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r56-s8-closeout-final.json)
（SHA256 `ed3c56562ff66a3f72294e84cc7923aa6c77b5b8f7a77ddcf7834bddc4999c78`）确认：

- 真实 Main 提出单个可信假设并启动一名 native Investigator；Worker 本人保存 supported 答案，由原 Case 接收。保存的 parent_tool_use_id 对应该 native Agent，Main 未代写。
- Main 一次 route 保存经原校验接收，再填写 Contract；无授权时真实停在 S8，未进入 S9。模型工具未手工调用领域登记/消费命令。
- 专业审阅反馈仅修正 scope paths 为 `server.py` 与 `web/e2e/settings/settings.spec.ts`；修稿时旧稿字节和 Runtime 不变，并保留新旧稿。这是专业审阅修正，不是程序新增 gate。
- 最终待批稿 SHA256 为 `76bc2e8540a0e4dbacafe6f0558cfb95df18dd28a50305310623406fac255f21`。显式 `simulation=true` 的外部 human_decision 输入经原 evidence add 注册后，原批准事务接收；最终 Runtime rev11 为 `bug_resolution.planning`，Case r5 为 `contract_approved`。只到 S9 planning 机器入口，未开展 S9 专业执行。
- 三段模型分别为 224.409、102.849、27.834 秒，均 exit 0、未 timeout；101 条直接 Hook 响应（79+16+6）均 exit 0。选定内部 deadline/defer 模式和专业 schema/route 拒收均为 0；产品、Finding、trace hash 不变。没有受控性能对照，不能据此声称性能收益。

[R54 辅助真实报告](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r54-s8-model-summary.json)
另保留两名 native Investigator、原 Case 双 supported、route、无授权等待及仿真授权出口证据；
它不代替后续 R56 题干与被动去重版本的验证。

### 验证范围与未覆盖项

最初规划的十个关键场景保留如下；这是覆盖清单，不是十项均已真实模型通过：

1. S7→S8 冷入口：正常无卷不报恢复，Main 得到实际资料与当前专业问题。
2. Main register/dispatch 后真实平台启动 Investigator，首次上下文包含 exact answer 路径。
3. 普通调查 supported/refuted/inconclusive：只保存一份专业结果，不追加 generic PLAN。
4. 非空新假设与 boundary 引用：合法示例可直接使用，身份由系统生成并等待 Main 派发。
5. 证据/环境阻塞：明确现有字段填写方式和 Main 下一动作，不伪造已执行观察。
6. Write、脚本或扩展工具保存：共用原消费者，无手工 issue/consume 和额外登记信封。
7. 未变已接收项遇普通工具调用不重送；异答保存仍 conflict，重复保存不重复推进。
8. Session 恢复、Stop/Idle、来源过期：每次保留路径或明确准备缺口，个人与阶段完成分开。
9. 同 Case 兄弟结果、并发与 pending 恢复：原 CAS/幂等边界及必要事务内复验保持。
10. Main 原 route/Contract/consume 出口：最终字节先审后批，无授权不进 S9，REQ 仍走人闸。

真实模型覆盖 supported 主路径、native 派发/本人保存、route、Contract 无授权等待、
专业修稿及仿真授权至 S9 入口；refuted/inconclusive、环境阻塞恢复、非空新假设、
MCP 特定保存、冲突/崩溃注入等仍仅有机制测试或尚未验证，不能称全分支通过。

上游是 synthetic sealed S7 setup，借用历史真实 Finding/观测，未复制旧 Runtime，
不是新 S7 通过或完整 S7–S9 产品链。原 Runtime 1.3、Batch、Finding 三 schema 单验 PASS；
借用产品缺 `docs/design/prototypes/settings/scenario-model.json`，全项目 `validate --all` 仍失败，
未为通过检查补造不相关专业材料。首次 producer constant 错误的拒收现场及各次失败原样保留。
外部 human_decision 是明确标注的仿真输入，不代表真实用户审阅或授权。
20261014 记录仍只证明历史机制夹具，不是真实模型专业验收。
本轮未做全平台、S9 修复/独立复验或全分支真实模型验收；不标记全局九阶段或 S0–S11 目标完成。
