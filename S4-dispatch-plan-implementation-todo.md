# S4 整体派发计划落地清单

> 2026-09-19 · waves-v1 首版完成。依据 [优化方案](S4-dispatch-plan-optimization.md) 与 [正式机制](blueprint/L4-agent-dispatch-governance.md#dispatch-plan)。本文件记录实施进度，不是目标项目的派发计划或新的规范权威。

- [x] 1. 升级 index/TASK、规划与派发指引，明确静态计划和动态进度。
- [x] 2. 实现计划解析、当前 REQ 成员闭包、依赖/资源顺序联合检查。
- [x] 3. 接入 tasks check、TR-002 注册、S5 subjects、TR-003 冻结；计划不得污染任务分母。
- [x] 4. 实现共享候选求值，S6 展示波次 todo、阻塞原因及下一批，容量不足持续补位。
- [x] 5. 派发入口复核计划与当前事实，接入已有 Agent 提示；保持 legacy 恢复且不伪造平台运行。
- [x] 6. 增加示例与边界回归：提前释放、只读模型并行、资源环、成员遗漏/跨 REQ、漂移、未集成与恢复。
- [x] 7. 运行相关及全量检查，更新蓝图实现边界和本清单，报告实际完成与限制。

顺序：1 → 2 → 3 → 4 → 5 → 6/7。每项以实际实现和验证为准打勾，不以文件存在代替闭环。工作区已有其他任务改动，本次不回退或提交这些改动。

## 首版实施记录

- 计划格式：`waves-v1`，升级现有 index；TASK 只保留静态任务书，执行进度从 runtime 投影。新 REQ/TASK 模板显式采用该格式。
- `tasks check --req <REQ>`：按需求筛选发现范围，对账成员、波次、真实依赖及资源顺序；JSON 输出有问题时返回非零。宽写域重叠给诊断，实际调度保守互斥。
- TR-002 登记独立 `dispatch_plan` 文档；S5 精确 subjects 自动包含计划；TR-003 比对 S4 登记内容，不能刷新哈希掩盖变更；成功后将计划和 TASK 投影为 locked，复用既有文档保护和集成检查。计划不会被算作 TASK 或架构。
- `s6 status --capacity <实际并发总槽位> [--json]`：展示波次 todo、状态/原因、下一批；未声明容量时不擅自假设槽位。真实前置满足即可释放，不增加全波屏障。
- workgroup 注册在既有 Writer 事务中核对真实依赖、owner、范围与传给 Builder 的 TASK 字节。当前适配一 TASK 一个 writer；不是自动平台 spawn。
- 新计划的阶段门和派发投影共用当前 Result/owner/assignment/checkpoint 判定。集成 checkpoint 绑定 Result 路径/SHA256 并记录 `verified_at`；内容改变必须重跑检查，单纯 ack/cleanup 不更新成功检查时间或摘要。
- Hook 的已有 Agent 指引给出计划与实时清单入口；无新增 Stop/普通工具硬门。清理 worktree 仍为提醒。

## 验证与边界

新增测试覆盖六任务三波、容量、提前释放、已报告未集成、重复/缺失/跨 REQ、资源联合环、写域冲突诊断、Git 输入和 dirty 副本分离、冻结漂移、重复 owner、越界派发、失效证据/旧 assignment、计划 subjects 与 TASK 分母隔离，以及实际 CLI JSON 只读投影。

最终 `go test -p 2 ./...` 通过（`/tmp/s4-delivery-tests.log`）；doctor 的结构/schema/示例/链接与手册同步检查通过（`/tmp/s4-doctor-final.log`），运行健康指标另行报告。示例 `tasks check --req REQ-042 --json` 通过；36 个具体本地文件链接和 `git diff --check` 通过。初轮针对测试发现模板迁移检查仍要求旧执行态字段、JSON 失败退出码测试旧约定和标准 Markdown 表头分隔行误判，已同步修正。新 S5 测试曾误用“子集匹配”辅助函数，已改为检验实际精确 subjects 语义。

首版使用具体相对文件/目录写域，不支持 glob 授权或同文件区块级并行豁免；共享资源需显式声明。调度采用确定性启发式，不做预测/全局最优排程。历史执行无计划时明确标识 legacy，不补造历史审签。真实 Claude Code 多 Agent 并行、目标产品联调和效率提升数据尚未实测。

改动未提交；工作区其他任务已有改动未回退。

## Agent 实时指引补齐（2026-09-19）

- [x] 在父会话 SessionStart / PostToolUse(Agent 或 Harness 派发、完成、集成动作) 生成就绪任务、待回收结果与阻塞摘要，直接进入已有 Agent 指引。
- [x] 限制摘要长度；不猜测平台容量、不自动 spawn、不因投影失败增加 Stop/Tool 门。
- [x] 验证摘要、事件接线和既有 Hook 回归。

本轮验证：`go test -p 2 ./internal/dispatch ./internal/cli ./internal/hook` 通过；实际候选到 `hookSpecificOutput.additionalContext` 的 CLI/wire 回归通过；手册已重新生成，doctor 与 `git diff --check` 通过。测试证明本地接线与输出内容，未宣称真实 Claude Code 会话接收已实测。改动未提交。


## 审查修复（2026-09-19）

已落实 [审查报告](S4-dispatch-plan-review.md) 的四项修复：注册校验实际 write/output 权限并集；非法资源行显式报错；相对 root 统一规范化；Result 内容版本与集成 checkpoint 绑定并支持不重复合并的重新验证。原失败验收已纳入默认测试，补充实际 CLI 的 Result 改写/重验/清理后恢复测试。最终验证结果见审查报告修复记录。
