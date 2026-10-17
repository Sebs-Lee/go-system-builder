# G06 — 派发、接收与积压反压的落地设计

> 全局设计；承接 [G06 guideline](../GLOBAL-OPTIMIZATION-GUIDELINES.md#g06--并行速度匹配接收能力)。接收建议与显式持久反压工程切片已验证；存活观测与吞吐对照待实施，原DAG/权限继续有效。

## 1. 调查证据与结论

[dispatch.LoadWithFiles](../../internal/dispatch/board.go) 已核对冻结plan/TASK，读取真实进度，以 capacity减running给slots；[NextDispatch](../../internal/semantic/dispatch_plan.go) 优先释放直接依赖较多的TASK，处理资源/写入冲突及排队，reported不能释放消费者。

[dispatch.Advisory](../../internal/dispatch/guidance.go) 已展示eligible/reported/waiting；[dispatch_guidance](../../internal/cli/dispatch_guidance.go) 只在指定Main检查点发提示。[workspace pending](../../internal/cli/workspace_pending.go) 从原报告/checkpoint派生命令，集成保持Git锁和持久恢复。

结论：无需新队列或调度器。增加接收优先的只读当前动作投影与容量建议，并在实际注册时继续执行原约束。

## 2. 当前动作的确定性选择

### 2.1 首切片接口与原消费入口（2026-10-14）

原 `s6 status` 增加显式 `--advice`，JSON 保留原 Board，附 reception_advice；
原 pendingIntegration 仅追加 TASK ID 与 checkpoint 引用，来源仍是原已登记
Assignment/当前 checkpoint。纯投影按直接 consumer 数、TASK ID、Assignment ID
稳定排序，恢复/接收先于新派发；cleanup 单列，缺少专业返工理由只给 question
和原只读 pending 入口，不输出带 `<placeholder>` 的可执行 argv。

原 pending 的 complete 但 Runtime 未完成投影/Result 已变仍按原恢复处理；
reported 不释放 consumer。明确 capacity 才给整数 dispatch_budget，未知为 null。
首切片 `--max-unreceived-results` 仅限制 advice 的推荐预算，不改原 Board.next、
TASK/DAG、注册授权或 Runtime 配置；真正配置反压的冻结/注册接线另行实施。
Main checkpoint guidance 消费同一投影，读失败只给原诊断，不新增普通工具/
Stop/Idle gate。不得以 checkpoint 状态推断当前有存活集成进程。

已新增 `dispatch/reception_advisory.go`，输入原 Board、pending 投影、显式容量及可选建议上限，输出 `ReceptionAdvice`：当前动作、稳定 pending_actions、分别计数和建议预算。`current_action` 包含 action_kind、subject、owner、原 argv、reason、checkpoint_ref 和必要 question。CLI 前后核对 Runtime 身份、generation、revision；不一致要求刷新。checkpoint 与命令执行之间仍可能变化，执行时由原消费事务再验证。

优先顺序：有明确可恢复的未接收Result/checkpoint→Main消费；阻断当前关键依赖的待接收结果优先；其次可派发TASK；已启动长集成/检查则给观测引用而不重复启动。阻断需代码返工时给现有rework路径，原因由专业题填写，命令不带假装可直接执行的占位符。

首版不估算复杂关键路径时长；沿用当前直接依赖排序，并在同分时稳定按TASK ID。它是建议优先级，不新增禁止合法独立工作的授权门。

## 3. 反压与容量

实际capacity继续由支持的平台/操作者提供，未知时只给ready，不声称可派几人。已报告但未接收量从原canonical completion和checkpoint计数，失败/preserved单列，不能当成已完成。

拟配置 `max_unreceived_results` 与解除阈值，作为可观测派发预算上限；配置缺省首版仅提示integrate_first，不改变slots。启用阈值时超出者保持queued，required set不变；已有冲突注册门照常。

观测running、reported、verified待ack和cleanup分别计数。cleanup积压给清理待办，不以任意总量持续阻塞全部开发；只有真实资源冲突仍阻止相应任务。

### 3.1 持久反压实施契约（2026-10-14）

新 1.2 Runtime 的 `configuration.optimization.dispatch_backpressure` 可选对象
固定 `max_unreceived_results >= 1` 与 `0 <= resume_unreceived_results < max`。
由 `init/install --dispatch-max-unreceived-results N --dispatch-resume-unreceived-results M`
显式选择；缺省不启用，也不迁移 1.1 或已有 Runtime。初始化中断重试必须
匹配原选择；普通 Writer、reset/recover 保留冻结配置，不能现场开启或调阈值。
单次 `s6 status --max-unreceived-results` 继续是建议参数，不能开启持久门。

接收列表沿原 Assignment/报告/checkpoint 生成，CLI 与注册事务共用原投影。
提供从原 Writer snapshot 加载 HookContext 的只读入口，注册不得再次读取另一个
Runtime 版本。U 按当前 Assignment 的原未接收结果计数，含失败及 verified
待 ack，不以 TASK 数或 Agent 数替代；cleanup 不计。当前报告与 checkpoint
绑定失配仍算未接收，不能靠旧 ACK/complete 释放新报告。原 Main ACK 已登记、
checkpoint 绑定仍一致时可解除反压，不要求 cleanup 成功。

显式配置下原 Agent ACK 事务另保存 `completion_acknowledged_reports`（原
canonical report path→已登记 SHA），只引用本人的原 TASK 报告；与 checkpoint
当前绑定一致才算已接收。重交报告沿原完成 consumer 接收时清除旧 ACK/摘要，
回到 reported 等待原 Main 对新报告的确认；默认和旧 Runtime 行为不变。

原 Runtime 可选 `dispatch_pressure` 只保存 Runtime ID、generation、paused
这三个迟滞投影字段，不保存队列或新授权。原 Builder 完成报告与 Agent ACK
Writer 在同一事务中维护它：U 达到 max 则 paused；paused 保持到 U <= resume。
读操作只投影，不写 Runtime；新代次忽略旧迟滞。注册事务复核实际 U/迟滞，
启用且 paused 时只暂缓新的 S6 Builder 注册，保留当前结果、已登记职责与
全部 required 集合。接收、恢复、其他阶段和非 Builder 职责继续走原门。

配置启用时原看板把可派 Builder 显示为 queued、next 为空；原 CLI/guidance
展示解除阈值和当前接收量。未启用保持原形状及原注册行为。运行中的 Builder
可以继续交付，使 U 暂时超过触发阈值；此阈值不是取消在途作业或预留未来
结果配额。并发新注册仍由原 Writer/CAS 与依赖/冲突门串行复核。

验收包括初始化/中断选择、冻结/reset 保留、多个 Assignment 计数、触发与
中间迟滞、原 ACK 后解除且 cleanup 保留、旧报告绑定失配、代次变化、未知
容量、默认兼容，以及原注册/完成/接收命令真实链。新字段均须闭合 Schema。

## 4. 并发与接收

### 4.1 原工作区锁的运行观测契约（2026-10-14）

新增只读 `runtime workspace integration-status --root <Main>`，使用原
LockWorkspace 的同一规范化 checkout 与 `.claude/integration.lock`。
以只读方式打开已有普通文件，非阻塞探测实际 OS 锁后立即释放探测锁；
不创建目录/锁文件，不改 checkpoint、Runtime、journal 或任何诊断记录。
输出 `busy/available/not_created/unknown`，绑定实际观察的 checkout/lock_ref，
明确是瞬时资源状态、不是权限/PASS/责任完成或具体进程身份。

原 `s6 status --advice`/指定 Main checkpoint 消费同一观察。实际 busy 时
current_action 为 observe_integration，给完整原只读入口，pending/recommended
保留供查看；不再次建议执行接收命令。available/not_created 时仍按原 pending
选接收；unknown 给来源诊断，执行原命令时原锁与 Writer 继续复核。不得把
旧 running.json、checkpoint 或 PID 当锁持有者的可信身份。

实际注册/集成/check 命令继续按原锁串行化，不靠这个读取授权或绕过门。
锁可能在观察后立即变化；不以 wall-clock 判任务死亡，不扫描全部工具事件。
验收须覆盖真实子进程持锁与死亡释放、缺失/损坏/链接锁的只读行为、实际
长 Main check期间观察且第二请求不启动重复检查，以及原完成后重新接收。

Windows 使用原 LockFileEx 非阻塞独占锁；只读 handle 符合官方要求的
GENERIC_READ/GENERIC_WRITE 任一种访问权。API 文档核对不代表原生 Windows
验收，失败仍返回 unknown，见 [Microsoft LockFileEx 文档](https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-lockfileex)。

Main操作从原pending的完整argv派生。异步Agent启动只表示平台已接受启动，后续completion事实必须来自原报告；Worker完成不等Main接收才允许结束本人。

注册在Writer事务内重新核对owner、依赖和冲突；两个同时派发不能靠同一旧board突破范围。Integration沿原Git锁、checkpoint和Main目标检查串行化；反压不跳过合并后检查。

S7仍按coverage_complete保留所有required责任，真实slots限制执行速度；M04回执让已接收reviewer离场，Main调度余项，不重唤醒该reviewer。

## 5. 故障与恢复

Board/投影读取失败给Main具体诊断，不新增普通工具或Stop gate。过时advice在实际注册/集成时由原校验拒绝，重新投影而非改plan。

Main重启从原pending重建接收顺序；合并中断沿checkpoint继续，不重复合并。reported但报告不可读是接收恢复项，不推断worker死亡/另派writer。真实冲突保全树与报告，由原rework/规划路径修复。

## 6. 验收与切片

在board/guidance/pending测试加：多个ready+一个reported优先接收；解除阈值后补位；容量未知；独立慢任务不挡consumer；reported不释放；blocked/rework与cleanup分开；stale advice并发注册；重启排序稳定。

原scope/owner/资源环/输入漂移/集成恢复测试继续通过。实战测集成吞吐、交后等待、积压与冲突，防止为了较少WIP延迟必须覆盖。

第一版已接原 Main 当前动作投影，随后在明确配置时启用反压。配置变化不改冻结TASK/plan；旧入口仍给原advisory，不声称平台已经自动spawn或接收。首切片的测试覆盖纯投影排序/预算、原 CLI 未知/零/实际容量、原 next 不变及状态/journal 只读；真实 S5→S6 Builder 报告到 Main 实际 Go 检查与集成完成链验证了原 pending argv 和预算释放。此证据不代表真实模型、平台吞吐或持久反压已验收。

后续顺序：先固定原 Runtime 配置中阈值/解除阈值的冻结和修改契约，再确定原注册事务消费预算的显式启用边界；随后接原集成/检查的存活观测，最后做并发积压与吞吐对照。不能让只读首切片参数隐式升级成注册门。

首切片最终验收：完整 CLI PASS 724.533s，系统包 PASS 136.136s，Hook 全包
race PASS 1.884s，领域及 CLI 定向 race 通过；五平台构建和两种全新安装通过。
初次 S5 竞态前置转换失败与独立成功重试均保留，原因尚未确定；两个成功
Worker 隔离用例受宿主 bwrap uid-map 禁用而排除，未计成功。完整记录见
[接收建议证据](../evidence/global-optimization-reception-install-20261014.json)。

持久反压切片最终结果：完整 CLI PASS 1082.278s，其余包完整回归 PASS；
领域 race、CLI 定向 race、默认/packet/反压四种新安装及五平台构建通过。
cleanup/摘要失配通过真实 ACK 保留后的 checkpoint 模拟验证；多 Assignment
计数/迟滞通过纯机制验证，未称真实并发吞吐。真实并发接收、运行观测和
cleanup 故障实验继续保留，见
[持久反压证据](../evidence/global-optimization-pressure-install-20261014.json)。
