# S4 机制审查（2026-10-16）

## 结论与范围

**原问卷、消费者、任务检查与 TR-002 主链可保留，S4 需要局部返工。**
审查阶段完成文档/实现对账、三个独立只读审查和隔离复现；当时未修改生产实现。
用户随后授权局部返工，现已完成 S4 实现、定向回归及新包真实履职验收；答案 schema 和全局工作状态不变。工作责任仍使用 `pending / active / blocked / done / cancelled`；TASK 的
`complete/cancelled` 是原文档字段，不改造成另一套工作状态。

现有 S4 正常链及 pending 恢复等 23 个顶层基线用例通过。新增五个边界探针中，
四个失败、一个通过：已推进后的原答卷重放正常，不把其他阶段的重放问题套到 S4。
NotReady 用例是受控 gate 仿真，不声称真实模型已产生同一现场。

## 审查确认的缺口

| 项目 | 证据与影响 | 最小返工方向 |
| --- | --- | --- |
| 可以直接跳过当前答卷 | `docs/control/loop-definition.json` 的 TR-002 无 required evidence；当前接收核对在 `internal/controller/planning_followup.go` 的自动路径中。隔离调用原 `transition.Apply(TR-002)`，没有 S4 接收记录仍进入 S5。 | 在原 TR-002 事务验证当前 S4 接收及完整批次；明确未启用问卷的旧流程保留原路径。这是阶段完成条件，不加工具身份/权限门禁。 |
| 未过门耗尽原接续 | `internal/runtime/submission_control.go` 在 S4 NotReady 后写 `no_transition`；S4 `SubmissionFollowup`、`currentPlanningSubmission` 仅接受 pending，Recover 的旧记录兼容只覆盖 S3。探针中随后普通 Read 和 Recover 都不能接续。 | 未过门保留当前原接续；兼容同入口已有记录，已接收专业答案保持原样，作者不再交卷或手改状态。复用原 Writer、回执和恢复路径。 |
| 保存拒绝仍让作者猜提交方式 | `readTaskPlanningFrozenFile` 只报 source-view/checkout 差异；未提交 TASK 的回执既不说明准确分支与提交动作，也不带准确 answer 路径，反而使用通用具名修订提示。 | 入场直接给准确 index、已知 TASK/候选路径、TASK/index 模板与绑定分支；按原 Git 来源提交专业产物后保存一次答卷。尚未接收时修正并保存原路径；冻结的上游输入继续原校验。 |
| 真实漂移被误报为缺交卷 | `internal/qualitygate/evaluator.go:202` 忽略 `verifiedTaskPlanningPacketSources` 错误。探针改动已接收 TASK 后，门只报 `evidence:planning_task_record` 缺项，丢失具体 TASK 漂移原因。 | 继续用现有完整批次投影，保留准确错误和责任；未接收才提示填写，已接收漂移应给原变更/返工动作。S4 已投影完整输出，不需要重做 S3 的合同集合修复。 |
| 旧指引增加作者步骤 | L3 §3/§4.5 仍要求手工检查、登记信封，§11 要作者处理 pending/recover 并停止 no_transition；`specification-planning` Step 12 重复要求 tasks check。Hook 的 S4 文案只给 view/answer，未提供强制的 index 路径及 index 模板。TASK 模板 §4 仍称 Hook policy 参与动态工具权限求交。 | Blueprint、协议、Skill、模板统一为专业规划→提交专业产物→填答保存；机检和证据由原链负责。粒度/语义建议仍供作者判断，检查命令用于具体诊断。scope 作为派发/集成检查依据，Hook 监听提醒。 |

保存重放探针验证：返回同一 submission receipt，Runtime、journal、answer 三个文件
字节均不变，且没有虚假的 system followup；此项不列为已复现缺陷。

## 应固定的最小工作流

1. 当前 `planning.tasks` listener 给准确 view/answer、冻结合同/条款分母、输出与模板路径。
2. Main 规划本 REQ 的 TASK、条款、专业 scope、Closing Contract、真实依赖与 waves-v1
   派发计划。TASK ID、粒度及语义由作者判断；路径规则和已知文件由系统说明。
3. 完成当前专业产物；原声明为 Git 来源的文件提交到 Hook 给出的绑定开发分支。
4. Main 填现有 `conclusion / index_path / task_paths` 并保存。原消费者做批次机检、
   evidence 和接收；不另交信封、重复 PLAN、身份或 hash。
5. 原门以完整已接收事实判定，TR-002 原事务登记并推进 S5。未过门给具体缺项、
   原 answer 路径及下一责任；系统继续原接续，不要求作者猜 recover/issue/consume。
6. S5 独立审查及原人闸保留；已接收不等于 S5 PASS 或代码完成。

## 旧 S5 清单的接续归属

r28 当前为 S4/e7，原 listener 已发行准确问卷，并绑定新 BE 指纹 `e8a5fd…`。
旧 review team/manifest 属于 S5/e4，仍引用旧 BE `a04fa8…`。TR-004 只使已消费的
review evidence 失效，未退役清单；全局 validator 仍扫描 Runtime 引用的旧 manifest，
因此 `validate --all` 和 `doctor` 失败。

旧 team 不会被当前 entry 的 S5 selector 当作新派发；当前 S4 输入准备未被阻断。
S4 正式交付后，S5 仍须从最新已接收 S4 tuple 和当前 S5 entry 创建其评审批次。
旧清单退役/历史校验边界属于返工后的评审批次接续缺口，应明确处理，不能让 Main
刷新旧 manifest/Runtime 指纹、复用旧 PASS 或增加一份作者答卷来消除错误。
本次只确定归属，未扩改 S5 生命周期实现。

## 证据与限度

本机评估根：`/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z`。

| 记录 | 结果与限度 |
| --- | --- |
| `reports/s4-audit-r116-source.json` | 当前完整源码冻结清单，包含未跟踪生产文件。 |
| `reports/s4-audit-r116-baseline.json`、同名日志 | 62.262 秒，23 个顶层用例通过，exit 0；覆盖正常保存、原 TR-002、pending 恢复、来源拒收及 S5 来源消费等，不是全仓回归。 |
| `reports/s4-audit-r116-probe-inventory.json`、`s4-audit-r116-probes.patch` | 审查时五个新增探针只存在于隔离 regression workspace，可由补丁复现；返工阶段已纳入 `internal/cli/s4_continuation_test.go`，保留原断言。 |
| `reports/s4-audit-r116-probes.json`、同名日志 | 41.115 秒，4 失败、1 通过，exit 1。失败项如上，未改写为通过。 |
| `reports/r27-s3-prefix-s4.json` | after-r27 真实 muse 首次 S4，58.561 秒，准确路径、保存接收、原 TR-002 进 S5；没有手工 issue/consume。属于此前候选，不能代替新 BE 返工后的 S4 验收。 |
| `reports/r28-s3-exit-listener.json`、对应 S4/e7 packet | 新 BE 输入与准确问卷已送达，状态仍 tasks/e7，当前 S4 提交为 0；未执行本轮 S4 专业工作，不代表已通过 TR-002。 |

## 返工实施与验收

本次按以下顺序实施：先把上述最小流程固定到 L3/协议/Skill/模板，再补原事务接收校验、原接续与准确诊断。
保留当前正常保存、完整批次/条款/DAG、冻结源、pending 恢复和已通过的重放行为。
补齐五项边界回归，并覆盖本 REQ、取消项、同代返工重入和 S4→新 S5 输入选择。
最后全新装包，用真实 muse 交付新 BE 对应的 S4 到 S5 出口；旧清单全局校验的失败与
处理结果单列，不借历史正常通过代替本轮验收，不启动 S6 实现。

### 当前实施记录

- 已先固定 L3-S4 最小契约，协议/Skill/TASK 模板同步删除普通作者的系统步骤。
- 已补原问卷准确输出/模板/候选路径与 Git 分支提示，已接收时不再要求保存同卷。
- 原门区分未接收与已接收来源冲突；原 TR-002 事务复核接收及完整批次，只登记已接收来源。
- 当前原接续保留 NotReady，兼容同入口旧 no_transition；历史 entry 不复活。
- 冻结源 r121：60 个定向顶层用例通过（217.866 秒、exit 0）；fileview/submission 两包完整回归 14 项通过（3.744 秒、exit 0）。此前 r117–r120 的失败日志保留。
- after-r29 Linux amd64 评估包构建成功，1631 项源码清单与 r121 完全一致。新安装按原接收/TR-004 机械重放前置，S4/e7 问卷未预填；机械前置不计真实新专业判断。
- 真实 muse 15 轮、82.015 秒，依新题干自审发现 TASK-001 §4 output paths 漏列 §6 的 `go.mod`，修正后提交 `7cd59e5`。只用一次 Bash 完成 Git 提交、一次 Write 保存原答卷，无手工 issue/consume/recover、重复机检或手改 Runtime。新 BE 保持 `e8a5fd95…`。
- 当前 e7 原接收唯一，原 TR-002 `evt-tr-002-r39` 推进到 S5/e8；完整 evidence subjects、实际产物、Runtime 登记一致。Main 最后发言引用当时 pending 回执，最终游标以 journal/Runtime 为准：随后原监听已完成接续。
- 两次原保存重放均返回相同回执，Runtime/journal/answer 字节不变，无虚假 system followup。安装包 `tasks check` exit 0。
- `validate --all`、`doctor` 均 exit 1，仍是旧 S5/e4 manifest 的 BE 指纹 `a04fa8…` 与当前 `e8a5fd…` 不同；旧 manifest 和 r25/r28 现场均未修改。S4 范围验收通过，全项目验收为 false；未启动 S5 新独立评审或 S6。

### 本轮最终证据

| 记录 | 事实 |
| --- | --- |
| `reports/s4-rework-r121-source.json`、`s4-rework-r121-targeted.json` | 最终统一源码和 60 项定向回归；包括五个原探针、同入口兼容/历史拒绝、取消项/完整集合、事务原子拒绝及 S5 实际输入消费。 |
| `reports/s4-rework-r121-source-receipt-full.json` | fileview/submission 两包完整回归 14 项通过；不是全仓回归。 |
| `reports/after-r29-source-inventory.json` | Linux amd64 评估构建，源码与 r121 相同；不声称五平台发布验收。 |
| `reports/r29-s4-rework-setup.json` | 全新安装、原消费者机械前置/原 TR-004、当前 S4/e7 与未填问卷；冻结来源和原现场哈希。 |
| `reports/r29-s4-rework-model.json`、`logs/r29-s4-rework-model-model.stdout.jsonl` | 真实 muse 的准确路径投递、专业修复、提交和一次保存；模型 exit 0。 |
| `reports/r29-s4-rework-acceptance.json` | 原接收/TR-002、产物/来源/登记对账、两次零写入重放：`S4_scoped_passed=true`；保留 `whole_project_passed=false`。 |

下一处已确认缺口是 S5 返工后旧评审批次的退役与当前 entry 重建。应由原生命周期处理，
不能照旧诊断建议让 Main 刷旧 manifest 指纹或复用旧评审 PASS；由用户安排后续阶段。


### 后续 S5 接续结果

S4 当时的旧 S5 清单全局校验失败，现已由 S5 返工在 after-r32 全新场景关闭，
见 [S5 机制审查的最终验收](S5-MECHANISM-REVIEW.md#最终候选与验收)。原 TR-004
退出旧评审批次，旧清单保留原 SHA；新 entry 使用最新已接收 S4 来源。真实独立双
PASS 与原 TR-003、六次原回执重放，以及现场 `validate --all / doctor` 均通过。
本页此前 after-r29 的失败结论及原现场保持原样，不能将后续结果回填为旧包通过。
