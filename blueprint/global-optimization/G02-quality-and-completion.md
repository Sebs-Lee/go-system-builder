# G02 — 质量指令、职责收尾与未决问题的落地设计

> 全局设计；承接 [G02 guideline](../GLOBAL-OPTIMIZATION-GUIDELINES.md#g02--质量指令给出范围优先级与结束条件) 和 [提交回执](../GLOBAL-OPTIMIZATION-DESIGN.md#4-共用基础交付消费与当前回执)。§7 执行收尾投影已落地；S6 v1.2 canonical 完成回执与 Main 消费为部分实现，不能据此声称完整 fresh-S7、S9 或真实模型验收完成。

### 当前接线状态（2026-10-14）

canonical S6→S7 的机器来源已接到原链：TR-003 在原 Writer 持久化 pinned Main
start receipt，`controlproof` 用 journal 校验发行；`mainreceipt` 复核原 accepted
Result、版本化 checkpoint、runner/tool receipts 和实际 Git contribution；
`buildsurface` 汇总精确 TASK、current/historical Record Main contributions 与完整
start..final 变化面，并验证原 typed ReviewPlan 的完整 `frozen_subjects`。S10 的
显式 `affected_paths=all` 表示全量审查范围；canonical product baseline 仍须逐项
覆盖完整 `changed_path` 分母，不能省略覆盖行或来源验证。

canonical S9 seed 现在在发布和 TR-012 状态变更前复验此 baseline，再沿原
`DraftPlanForRoot` 保留全部 TASK/产品路径、DV 与 QA 机器骨架；专业 assertion、
oracle、method 留给 Planner refine。TR-006 的 `builder_report_record` 只作为兼容
slot；原 Engine 只接受 `buildsurface.Read` 已证明属于当前 Main batch 的原
`completion_report` 引用，catalog 声明或作者路径本身不够。

S7 的原 `review.SubmitResult` 已有 v1.2 packet-mode Writer receipt 首片：保留完整
`submitted_answer`/`effective_answer`，并对专业答案字段计算 digest，在原消费 CAS 固定 Plan/Assignment/owner
tuple、Result ref/SHA 和 journal 坐标；原 pinned Plan、Result/Finding 字节、Evidence 索引和
journal 支持同键同答卷的只读重放。同一 Plan 修订中未变 Assignment 保留当前 receipt，受影响
Assignment 清除当前 receipt。此接线仍使用显式 `review.Result` 提交，没有受信 typed 答卷发行
或精确路径保存 Hook，专业 Reviewer 平台链仍待接通。Claim-owner 转移失效、20 MiB 输入预算与 pending-safe 清理已通过最终定向验证，
待 root 收口确认。

这是机器来源、Writer 回执基础与 seed handoff 的部分接线，不是 S7/S9 全链完成。canonical S7
typed 答卷发行、保存 Hook 与专业 Claim/Assignment/ReviewResult 平台链仍待接；canonical S9 fresh round、
native/workspace 全链、安装/平台覆盖和真实模型/宿主验收也未完成。当前记录的 core
mechanism 为 4/9、real-model complete chains 为 0/9，数字不因本轮定向测试改变。
对应的目标验证见 [TR-003 receipt race](../evidence/logs/s7-main-start-transition-race-20261014.log)、
[Main receipt race](../evidence/logs/s7-mainreceipt-race-20261014.log)、
[canonical surface normal run](../evidence/logs/s7-canonical-surface-targeted-20261014.log)、
[canonical surface race, 66.376s](../evidence/logs/s7-canonical-surface-race-20261014.log)、
[S7/S10 consumer race](../evidence/logs/s7-s10-canonical-product-consumers-race-20261014.log)
和 [S9 bridge targeted test](../evidence/logs/s9-fresh-s7-bridge-workflow-targeted-20261014.log)。

## 1. 调查证据与结论

- [AGENTS 模板](../../AGENTS-template.md) 的全局质量措辞要求广泛反证与未知处理，DRIVE 以全局 stage 为循环单位。
- [TASK 模板](../../docs/dev/tasks/TASK-template.md) 同列实现、测试、DV、QA、E2E；[Backend Builder](../../agents/backend-builder.md) 则禁止独立审查和其他角色产物，执行入口有呈现冲突。
- [hookctx.AssignmentContext](../../internal/hookctx/types.go) 已有 DoneWhen、RequiredChecks、ResponsibilityIDs、CompletionRef；[lifecycle.go](../../internal/hook/lifecycle.go) 缺 done_when 时要求再读 manifest。
- [CompleteTask](../../internal/assignment/completion.go) 和 [SubmitResult](../../internal/review/submit.go) 已有领域接收事实；[Guidance](../../internal/policy/engine.go) 尚未统一表达自己的完成与等待归属。

结论：必须先把交付与未决事项定位到原责任，再改措辞；单纯要求“不要过度检查”没有可验证的结束边界。

## 2. 具体改动

TASK §6 分成本 TASK 所属交付与后续消费责任两张表，后者只列责任/接收条件，不让 Builder 填 QA/E2E。已有职责命名保持，不新增审批角色。

新发行的 Assignment 使用原 done_when/required_checks/expected outputs，通过工作包引用已批准的验收集合；缺具体收尾时由 Main 补规划/派发输入，不能生成含混工单。旧 manifest 保留兼容，但包清楚标 missing completion contract。

拟定领域 `ResponsibilityFacts`：当前 owner、必交集合、原已接收 refs、本人待项、其他 owner 待项、系统恢复待项。`workpacket` 从这些事实生成本人状态，不能从 stage未过推断所有人未完成。

AGENTS/loop-orchestration 只保留有范围的质量动作：核对当前覆盖全集，按当前断言取证，有阻断风险走原路由，已接收后结束该责任。Main 继续保护整个工程结果，其阶段任务与 Worker 的局部收尾明确分开。

## 3. 未决事项与额外检查

未决问题使用当前 answer 的条件题，包含具体断言/决定、来源观察、影响与请求处理的责任。owner 由原派发、规格责任或人闸映射，作者提出问题不能自行赋予权限。

四种去向：本人缺证据→继续取证；业务未决定→原规格/设计责任；其他责任未交→调度等待；环境/接线错误→Main/系统恢复。范围外新风险保全证据并移交；若影响当前验收或安全，当前责任明确受阻，不简单以“范围外”免责。

新增检查记录到原检查回执/审查证据，原因来自 changed input、interrupted、invalid evidence、new hypothesis、independent review、post integration 或原 required check。这个原因由事件/输入比较派生为主，不增加 Worker 每次跑命令前的填表手续。

## 4. 回执与结束路径

全局工作状态及其唯一读模型以 [L4 §0.1](../l4-mechanisms/L4-state-transition-core.md#unified-work-state)
为准。以下 work_complete、execution_may_end、consumer_pending 是当前兼容接口，
不是三套可以独立变更的状态机；按该契约集中派生并逐步消除重复判断。本人交付、
Main 集成和独立审查是不同责任，统一使用相同状态词，不共用一个对象 completion。

domain adapter 接收成功后给 `work_complete` 与下一 owner。例如 S6 Worker 的结果登记成立，但集成未完成：本人完成，Main 待集成，全局 S6 未完成。若自己的检查仍缺/失败，则 work_complete=false，指向具体本人欠项。

Stop/TeammateIdle 从原接收记录和本人待项判定，不因全局 missing 或系统 pending反复唤醒。已有有效 BLOCKER允许按原规则结束当前执行；BLOCKER不等于任务完成或权限放开。

已接收修订仍走原返工/新尝试，不通过重新解释 AGENTS 直接修改冻结输入。人闸显示真实等待决定，不由质量指令自动授权。

## 5. 失败与恢复

角色/TASK 责任冲突归派发或模板缺陷，给 Main具体冲突来源；没有消费者的强制产物不能派给 Worker。未决问题找不到 owner 时登记维护缺口，不默认为执行者全包。

接收成功后响应丢失，从 receipt/domain ref重建本人完成；同事 REQ-change 或原输入漂移后，重复旧交付投影当前等待/失效事实，不能回复缓存的“可以继续”。

## 6. 验收与切片

1. TASK 模板/launch 包测试：Builder必交表不含独立 QA/E2E，后续责任仍保留且 gate分母未减少。
2. hookctx/packet 测试：无 done_when 的新派发给具体缺口；旧入口标兼容，不伪造收尾通过。
3. Stop/Idle 集成测试：本人接收后等另一 reviewer、Main集成、系统恢复或人决定，均不重复催本人；本人缺结果时继续原阻断/恢复。
4. 回执矩阵：有效 PASS、合法 Finding、BLOCKER、报告缺失、自己的检查失败、范围外阻断风险、超时和返工分别判定。
5. 固定案例 A/B 比较原/新质量措辞的额外检查与范围扩张，同时比较遗漏风险。措辞影响是行为实验，不能由字符串单测证明。

先补职责投影和完成回执，再同步改 TASK、角色、AGENTS 与 Skill；新旧入口共存时根据 capability/profile显示实际路径。原 independent/fresh-S7/human gate测试继续通过，完成条件变化要走原设计/批准路径。

## 7. 首个实施切片：执行收尾投影

先复用原 Hook 已读取的 Agent、Assignment 和 S5 已接收事实，生成只读
`ResponsibilityFacts`：owner、assignment、delivery_recorded、execution_may_end、
own_pending、consumer_owner、consumer_pending、missing_closing_contract。投影不写
Runtime，不从全局阶段 missing 推断本人欠项，也不把报告存在推断为检查通过。

`execution_may_end` 保留原 Stop/Idle 兼容矩阵：queued 可结束；one_shot、等待原批准、
原终态、原 blocked/BLOCKER、原完成报告等 legacy 规则适用于非 S7。S7 Reviewer（包括
one_shot 派发）须有当前精确 ReviewResult 已被原 round consumer 消费，或有原 Runtime
`work_blocked` 事件/当前 `site_lost` 事务及其有效证据；通用 `completion_report`、Agent
终态或 manifest 的 blocked 文本不够。已登记 S7 owner 若 context/Assignment 绑定不可用会被
阻断，不能降级成未知身份。只有 payload 无法定位已知 Agent 或 Runtime 不可读时保留原 fail-open。
其余返回本人 PLAN_REPORT 或 Result 欠项。`delivery_recorded` 仅表示原引用/报告或 S5 接收存在；
`work_complete` 仍由领域接收、required checks 和当前输入效力决定，本切片不新增此授权字段。

SubagentStop/TeammateIdle 的指引分别标明 Worker 交付与 Main 消费；集成检查仍由
原 task-integrate 执行，其他 reviewer、系统恢复及全局覆盖仍归原消费者。缺具体
done_when 只投影派发缺口，旧 manifest 不自动通过新契约，也不新增兼容门禁。
TASK §6 拆为本人交付和后续独立消费两表；AGENTS 的 Worker 入口按本人责任收尾，
Main 的 DRIVE 保持全局阶段覆盖。验收先证明投影/真实 Hook 输出和原业务链，再做
领域回执、返工效力与真实模型行为实验；不能以本切片声称 G02 全部完成。

独立 Builder 顶层会话的普通 Stop 先由原 session/root 注册定位本人，再核验当前
execution/CWD/Bootstrap，复用同一 Stop/Idle 本人矩阵，直接返回平台结果；不进入
Main 的 Stop preflight 或控制循环。当前 owner 未能定位时不猜测身份；S5 继续使用
原已接收发题视图的专用 Stop 路径。本人缺结果的 Stop 不写游标；收到原报告后
Main 待集成仍保留在原 pending/checkpoint。Stop loop guard 不放宽实际工具权限。

实施与回归进展见 [全局设计 §7.12](../GLOBAL-OPTIMIZATION-DESIGN.md#712-g02-本人执行收尾与-main-消费边界)
和 [安装/测试证据](../evidence/global-optimization-owned-completion-install-20261013.json)。
本切片不代表领域完成回执或真实模型行为实验已完成。

## 8. S6 原结果的当前责任回执

新增只读 `runtime task-receipt --agent-id <owner> --assignment-id <id>`，在原
控制 Runtime、原 Assignment、REQ-bound development branch 的 file-source
契约上投影；Worker 只能查询自己的登记身份。不新建提交状态或缓存授权。

使用原 completion_report ref 和 evidence index、原 qualifiedEvidence 校验及
TASK 的当前 path/version/hash；要求原 TASK/Agent owner 和 manifest assignment
一致。报告存在与其原证据仍合格分开；空 checks、fail/blocked/not_run、原 scope
deviations 和 declared required_checks 未覆盖分别列本人欠项。checks_source 明确
为原报告，不把作者声明伪装成此查询执行过命令。缺 done_when 另列契约缺口；不
执行自由文本 assert 或编造行为 oracle。work_complete 表示当前原领域交付满足
这些接收条件，不表示整体批次或独立复验通过。

本人合格但集成未 verified 时，work_complete=true、waiting_on 指向 Main 原
task-integrate；原 report-bound checkpoint 经身份、报告 hash、时间和 bound
branch 的真实 merge ancestry 校验后才投影已集成。本人证据失效/输入指纹变化
则 work_complete=false；Main 的失败集成/系统恢复与本人欠项分开，不能从全局
stage missing 反推本人未交。原 rework 清空旧 report ref 后回执必须显示新一次
尚未提交；重新提交从当前 refs 重建，不返回旧 receipt。

回执不写状态、不执行检查/集成、不推进门，不提升权限。输出有闭合 Schema；
读取结束复核 Runtime revision 和 committed source view，漂移拒绝。验收覆盖
当前报告、他人/旧 generation/错误 assignment、丢失或篡改报告、TASK 输入变化、
失败/缺漏检查与 scope deviation、等 Main、真实集成后、rework 和重复只读查询。

原 register-workgroup 对已存在 TASK 的 S6 Builder 登记必须在同一原 Writer/CAS
内把明确登记的 Builder owner 加入 TASK.owner_agent_ids，去重且保留既有 owner
及 TASK 状态、路径、指纹和结果引用。只能取本次已校验 manifest 的 Builder
assignment；不是 Worker 自领、补写运行态或从报告反推 owner。否则 manifest/Agent
已有合法 owner、TASK 却保留早先审查 owner，当前原领域资格始终不成立。回执不得
用“忽略 TASK.owner”掩盖该接线缺口。

原 task-integrate 的预集成 Inspect 拒绝分支，须先经原 Integrator 的 !Ready
保全路径写 durable preserved checkpoint，再投影 Milestone。仅提示“已保全”但
没有原 checkpoint 会让 workspace rework 无法定位失败责任。此分支不合并、
不验证通过、不 ACK、不清理；已有 verified/merge 后恢复分支保持原保护。

此切片已实现，实际失败检查/原保全/原返工/重新报告/真实集成链、当前回执矩阵、
Hook、回归和安装证据见 [全局设计 §7.13](../GLOBAL-OPTIMIZATION-DESIGN.md#713-g02-s6-当前责任回执owner-登记与失败接收恢复)
及 [S6回执安装证据](../evidence/global-optimization-builder-receipt-install-20261013.json)。
S6 typed WorkPacket、其他领域 adapter/回执与真实模型实验仍待接续。

## 9. canonical S6 本人收尾下一切片

设计已固定在 [L3-S6 本人专业收尾契约](../l3-stages/L3-S6-build.md#本人专业收尾的实施契约显式空本人检查已实现)：
原 Main Register 冻结显式 closing_contract 与本人 runner 集合；Main 的
required_checks 独立保留。PLAN 完整映射冻结 DoneWhen，Result 逐条回答
closing_assessment，原消费者校验覆盖与实际来源。专业判断是本人交付
收尾，不自动证明自由文本 oracle、集成或独立 S7。

显式空本人检查集合先实现本人完成并等待 Main；集合缺失不推为空，已有
Record 不回填。非空集合的原 workspace/native runner、收件和当前源快照
校验已接线，Result 持久 receipt/log 与工具闭合身份；新 key 按答案/闭合
身份区分版本，alias 经原 Writer 固定，已登记 key 重放零写。非 pass/current
proof 明确 pending，不由 Main 检查补位。实际宿主能力门继续生效。显式空集合切片已实现并通过原 CLI 登记/PLAN/Result/回执以及真实
Main 集成/Gate 的定向验证；非空完整链、修订与适用宿主 PASS 验收仍在推进。
旧缺契约 Record 保留 unassessed，不暗迁移。
证据见 [S6 本人收尾切片](../evidence/global-optimization-s6-closing-20261014.json)。

机器产品面到 S7 的来源接线和 canonical S9 seed bridge 已有独立定向验证；它们不
补齐本节仍待完成的 full-chain、typed S7 professional adapter、canonical S9 fresh
round、安装/平台和真实模型验收，也不改变上面的 4/9、0/9 进度计数。


## 10. Assignment revision 首个原 Writer 切片

exact locked TASK/plan_checkpoint 下，原 Register 可用稳定意图发行下一 Record；
fresh Team/Agent 重新 PLAN/Result，原 Main 使用 aN checkpoint、实际新 merge 与
fresh 全量 checks，旧 Result/ACK/verified 不继承。真实 r1→r2 Main/回执/TR-006
已正常验证，较早独立 Main 链另经 race 验证，历史源字节与 checkpoint 保全。发行共享 Main workspace
锁，pending/replay/并发单赢家经原 Writer 验证，历史回执不会回滚 current。

前版合并来源现仅支持完整审计的立即前版 side-parent 续接：历史 Record/Result/
原来源快照、发行 journal、前版 Main 检查及真实 Git 父节点均需有效；新版本
仍生成自己的 merge 与 fresh 全量 Main checks。legacy 缺快照、歧义来源和
无新提交继续保全。人闸版本绑定、TASK 来源变化与适用宿主 PASS 仍未完成。
当前证据范围与剩余边界见 [L3-S6 修订契约](../l3-stages/L3-S6-build.md)
及 [原修订验收记录](../evidence/global-optimization-s6-assignment-revision-20261014.json)。
