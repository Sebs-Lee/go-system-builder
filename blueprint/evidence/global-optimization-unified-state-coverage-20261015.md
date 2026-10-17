# 统一工作状态覆盖清单（2026-10-15）

工作进度仅使用 `pending / active / blocked / done / cancelled`。Agent 入口定义在
`AGENTS-template.md` 的 Work status；设计依据为 L4-state-transition-core §0.1。
下面列实现范围；测试、冻结包和真实仿真结果另见同批 application JSON，不能由本清单推定验收通过。

| 需要统一的工作对象/出口 | 实现入口 | 当前规则 |
| --- | --- | --- |
| Agent 责任进度 | `internal/workstate/activity.go`、assignment lifecycle/register/completion/revision | Runtime 1.3 存五态与原 last_lifecycle_event；原权限检查由事件还原，active 不授写权限 |
| TASK 执行进度 | 同一 workstate + task_lifecycle | 绑定真实责任；文档交付与实现责任分开；新尝试由原 revision 流程发行 |
| S7 ReviewAssignment | `workstate/review_assignment.go`、review、runtime/store、semantic/validator | 新格式五态与原事件匹配；合格 Finding 可完成审查责任，不等于阶段 PASS |
| Change WorkItem | `change/record.go`、runtime/change、CLI projection | 新写 pending；旧 open 仅在兼容边界读取，旧 Runtime 写回原格式 |
| S9 RepairPlan Assignment | `repair/session.go`、repair-plan schema | 新 Plan 1.1 使用五态，初值 pending；旧 Plan 1.0 保留原字节；实时进度由原 Runtime/消费者决定 |
| 本人回执 | `submission/work_status.go`、qualitygate/builder_work_status | JSON status 由原事实即时派生；accepted 与本人检查同时成立才 done；当前阻塞/撤销优先 |
| Hook/Stop/Idle、CLI 工作指引与进度提醒 | hook/responsibility、hookctx/loader、policy.AgentContext、CLI controller/work_progress | 共用当前责任证明与五态；已接收交付在退阶段后仍可结束，不重复催交；Stop/权限保留原检查，不由状态文字放行 |
| S6 派发与依赖看板 | qualitygate/dispatch_progress、semantic/dispatch_plan、dispatch | status 五态；已交付、已集成、可派发与容量是独立事实；下游仍要求 done 且合格集成 |
| 正式输入资格 | runtime.CurrentFormalInputDocuments、workspace/context、conventions | S5 复用原 S4 接收与完整来源闭包；调用者不再枚举文档旧 status |

## 保留各自含义的字段

- 全局 stage/phase、S9 repair cursor：流程位置，决定原迁移门。
- 专业 verdict、CASE/check 结果、Finding disposition：专业结论与覆盖依据。
- 文档 complete/locked、evidence valid/invalid：工件交付、冻结、来源效力。
- Workspace preparing/ready/complete、Team 收撤：执行资源及协作容器事实。
- Integration checkpoint 的 merged/verified/acknowledged/cleanup_pending：原事务恢复步骤，是集成责任证明的输入。
- Receipt 的 accepted/work_complete/consumer 字段：原接收协议事实，status 由它们及当前绑定重建。

这些字段不能替代本人工作状态，也不能凭重命名获得权限、完成证明或阶段放行。

## 兼容及失败边界

Runtime 1.1/1.2、历史 journal/packet/receipt 不做原地改写。新 Runtime 1.3 的工作字段
严格五态；缺少或错配原事件时拒绝。原 Writer/CAS、身份、6 秒 Hook 总预算、100ms
锁预算、独立审查、coverage、fresh S7 和人闸继续由原入口核验。

共用词表不要求每个领域发明五条迁移。例如 S7 原 Plan revision 重建当前 Assignment；
旧尝试由原 packet/journal/replacement 记录保留。没有原取消事件的裸 cancelled
Assignment 仍应拒绝，不能为凑状态新增一套取消协议。

## 本批验证结果

实现、schema、原接收链及兼容路径已完成定向回归；完整 CLI 扫描发现的五个旧夹具问题
已修正并逐项复验，没有再次重跑整个 CLI 包。宿主的 bwrap uid-map 权限限制仍使一项
隔离测试无法通过，原生产隔离检查保留。

R7 真实 Main 在新 Runtime 1.3 下经原 evidence/transition 走完 S2–S4；R8 两位独立
审查员均经原消费者得到 `status=done`，Stop 正常放行。Spec 的五条 Finding 触发原
TR-004 退回 S2，个人责任仍已完成。真实闭包包含 19 项正式输入。

仿真发现的 PLAN 保存误导和 16 文件提醒上限已修复并重测：保存冻结 PLAN 后直接给出
原 `begin` 命令，进度投影复用同一接收证明，来源 SHA 与 4 MiB 总预算继续校验。
本批工作状态统一已闭环；全局完整阶段验收仍为 **3/9**。证据和局限见
[验证记录](global-optimization-unified-state-application-20261015.json)。
