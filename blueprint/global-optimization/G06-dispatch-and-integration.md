# G06 — 派发、接收与积压反压的落地设计

> 全局设计；承接 [G06 guideline](../GLOBAL-OPTIMIZATION-GUIDELINES.md#g06--并行速度匹配接收能力)。新调度建议与参数为拟实现，原DAG/权限继续有效。

## 1. 调查证据与结论

[dispatch.LoadWithFiles](../../internal/dispatch/board.go) 已核对冻结plan/TASK，读取真实进度，以 capacity减running给slots；[NextDispatch](../../internal/semantic/dispatch_plan.go) 优先释放直接依赖较多的TASK，处理资源/写入冲突及排队，reported不能释放消费者。

[dispatch.Advisory](../../internal/dispatch/guidance.go) 已展示eligible/reported/waiting；[dispatch_guidance](../../internal/cli/dispatch_guidance.go) 只在指定Main检查点发提示。[workspace pending](../../internal/cli/workspace_pending.go) 从原报告/checkpoint派生命令，集成保持Git锁和持久恢复。

结论：无需新队列或调度器。增加接收优先的只读当前动作投影与容量建议，并在实际注册时继续执行原约束。

## 2. 当前动作的确定性选择

在dispatch包拟新增 `reception_advisory.go`，输入Board、pendingIntegrations、已声明实际capacity和优化配置，输出 `ScheduleAdvice{action_kind, subject, owner, argv, reason, dispatch_budget}`。

优先顺序：有明确可恢复的未接收Result/checkpoint→Main消费；阻断当前关键依赖的待接收结果优先；其次可派发TASK；已启动长集成/检查则给观测引用而不重复启动。阻断需代码返工时给现有rework路径，原因由专业题填写，命令不带假装可直接执行的占位符。

首版不估算复杂关键路径时长；沿用当前直接依赖排序，并在同分时稳定按TASK ID。它是建议优先级，不新增禁止合法独立工作的授权门。

## 3. 反压与容量

实际capacity继续由支持的平台/操作者提供，未知时只给ready，不声称可派几人。已报告但未接收量从原canonical completion和checkpoint计数，失败/preserved单列，不能当成已完成。

拟配置 `max_unreceived_results` 与解除阈值，作为可观测派发预算上限；配置缺省首版仅提示integrate_first，不改变slots。启用阈值时超出者保持queued，required set不变；已有冲突注册门照常。

观测running、reported、verified待ack和cleanup分别计数。cleanup积压给清理待办，不以任意总量持续阻塞全部开发；只有真实资源冲突仍阻止相应任务。

## 4. 并发与接收

Main操作从原pending的完整argv派生。异步Agent启动只表示平台已接受启动，后续completion事实必须来自原报告；Worker完成不等Main接收才允许结束本人。

注册在Writer事务内重新核对owner、依赖和冲突；两个同时派发不能靠同一旧board突破范围。Integration沿原Git锁、checkpoint和Main目标检查串行化；反压不跳过合并后检查。

S7仍按coverage_complete保留所有required责任，真实slots限制执行速度；M04回执让已接收reviewer离场，Main调度余项，不重唤醒该reviewer。

## 5. 故障与恢复

Board/投影读取失败给Main具体诊断，不新增普通工具或Stop gate。过时advice在实际注册/集成时由原校验拒绝，重新投影而非改plan。

Main重启从原pending重建接收顺序；合并中断沿checkpoint继续，不重复合并。reported但报告不可读是接收恢复项，不推断worker死亡/另派writer。真实冲突保全树与报告，由原rework/规划路径修复。

## 6. 验收与切片

在board/guidance/pending测试加：多个ready+一个reported优先接收；解除阈值后补位；容量未知；独立慢任务不挡consumer；reported不释放；blocked/rework与cleanup分开；stale advice并发注册；重启排序稳定。

原scope/owner/资源环/输入漂移/集成恢复测试继续通过。实战测集成吞吐、交后等待、积压与冲突，防止为了较少WIP延迟必须覆盖。

第一版只改Main当前动作投影，随后在明确配置时启用反压。配置变化不改冻结TASK/plan；旧入口仍给原advisory，不声称平台已经自动spawn或接收。
