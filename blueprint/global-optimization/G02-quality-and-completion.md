# G02 — 质量指令、职责收尾与未决问题的落地设计

> 全局设计；承接 [G02 guideline](../GLOBAL-OPTIMIZATION-GUIDELINES.md#g02--质量指令给出范围优先级与结束条件) 和 [提交回执](../GLOBAL-OPTIMIZATION-DESIGN.md#4-共用基础交付消费与当前回执)。新字段与 API 为拟实现。

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
