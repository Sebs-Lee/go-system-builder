# S6 机制审查与返工（2026-10-16）

## 本次结论与范围

S6 本轮有界验收为 done：构建、原返工、真实检查/Git 集成、已知失败提示和原 TR-006 交接已由 R45 验证。R40 提供真实 Builder 专业履职证据；R45 使用机械重放的专业前置与产品，真实重跑 checks、Git、ACK 和清理，并由两次真实 Main probe 验证普通进展与阶段出口，不代表最新包的新 Builder 从零实现。R45 原 `evt-tr-006-r79` 唯一进入 S7 `verification.planned`，准确问卷可读，停止时未填写 S7。本轮生产实现来自 after-r45 冻结包，验收后仅更新本次记录。此前失败保留；本结论不表示全阶段、多平台或发布完成。

## 固定后的最小工作流

工作状态只用 `pending / active / blocked / done / cancelled`。专业结论、原集成 checkpoint、manifest lifecycle 是领域事实，不另建工作状态机。

| 责任 | Agent 实际要做的事 | 系统负责的事 |
| --- | --- | --- |
| Main 派工 | 选择真实执行人员和专业分工，使用原登记、prepare、launch | 生成 manifest 草稿，带入当前 TASK、来源、scope、检查和准确命令 |
| Builder 开工 | 在给出的准确路径填一次专业 PLAN 并保存 | 原 begin/Writer 接收和激活，不另交 PLAN_REPORT |
| Builder 交付 | 实现 TASK、实际自测、提交产品；在准确 Result 路径填实际证据、判断和未完成项 | 复用已接受 PLAN 的断言和索引，带入已知身份/摘要；原 report 接收，不另交信封或完成消息 |
| Main 接收与集成 | 跟踪原 Builder 会话，按当前回执集成、恢复或安排原 revision 返工 | 原 Main 检查、Git 合并、ACK、receipt 精确清理；监听给具体缺项、责任人和原下一动作 |
| S6 出口 | 当前精确 TASK 批次全部完成 | 原 gate/TR-006 校验当前交付、检查和集成，进入 S7 |

Hook 只监听、提醒和流转，不按工具身份、权限、路径或 scope 拦截。Result 接收不等于本人完成，Builder 完成也不等于 Main 已集成。本人自测不能由 Main 的重跑替代。

## 已落地修正

- Main 与 Builder 各收本责任的指引；Main 能观察问卷路径，不会被要求代填、代实现。原 workspace launcher 就是真实派发，同一执行跟踪或恢复原会话。
- Result 自动预填已知断言/判据/条件索引，机器预填不被普通观察误当作作者保存。保留旧题干兼容，作者不选择消息 ID、不重抄元数据。
- 原 commit 请求给准确路径和已知 HEAD，作者只填实际文件及提交说明，原显式 commit 仍执行真实 Git 操作。
- 原接收建议保留 retry、rework 和 pending 诊断；`s6 prepare --revise-assignment` 从当前原 manifest 生成草稿和完整登记命令，保持 TASK、scope、checks、Closing Contract 与来源，仍由原 Writer 发行 revision。
- Main 与 Worker 的 dirty 反馈分别标明实际根路径。原 receipt 精确清理保留；产品 dirty、receipt/hash 不符仍保全现场，不使用 force remove、archive 或系统资产 commit。
- 上游原 TR-002 同步当前已接受 TASK 文档字段；历史已结束审查责任不误占 Builder 容量，历史与未知责任不被随意删除。
- TR-006 解析 Git 的未跟踪目录尾 `/` 标记后，仍执行原字面路径与产品 dirty 检查；保留控制 worktree 不再被误报，真实产品目录仍拒绝。
- 旧 revision 合入后检查失败，不要求它事后变 PASS。仅对准确绑定、原真实 Git merge 可核验、明确尚未验证的历史 checkpoint，不计作 verified contribution；当前任何错误仍拒绝，历史已 verified 贡献保持原投影。原 TR-003 起点至最终 Main 的全部 Git 路径仍进入 S7，残留产品路径沿用原未归属提示，不新增历史证明协议。

## 本轮监听减负与失败保留

本轮合并同一 gate 的重复 canonical 求值；普通 Pre 在同 revision 只读原 milestone，不重复复核或写入；提醒不跑 receipt/Git 链；删除 handoff 特判；提交后复用 Controller Snapshot，减少重复 Reconcile/Load。已有 Main 检查失败只读原 checkpoint 中记录的失败及 receipt/log 路径，并展示原恢复动作，不增加检查或协议。

R43 真实 Main 仍为 building，stream 报告 `QUALITY_GATE UNKNOWN` / `LOOP_GATE_UNKNOWN`、`transition_committed: false`，原 Read 约 288KB 超限，未执行完整链。R44 已原 TR-006 流转且问卷可读，但 SessionStart 触及外层 8 秒 deadline，不能计交接通过；普通进展还遗漏已存在的 govet 失败。相关修复的两次作者窄测失败保留，修正后窄测通过；重叠测试不累加。

## 回归证据

各轮覆盖重叠，不能相加成总数。冻结源清单、命令、原始 JSON 日志与失败记录均保留。

| 轮次 | 结果与边界 |
| --- | --- |
| r131 / r134 | 返工前相关七包定向基线通过；r134 原领域六包 359 项顶层测试通过 |
| r138 / r146 | 原 gate/transition/dispatch 三包整包 250 项通过；r146 qualitygate 整包 105 项通过 |
| r149 | integration 整包 89 项、dispatch 整包 21 项通过；CLI 16 项通过，含原 revision Writer、接收建议与真实 workspace 集成 |
| r150 | TR-006/原构建集成等定向 8 项通过；buildsurface 整包 7 项通过，含真实嵌套 worktree 与产品目录反例 |
| r151 / r152 | 新样本两处 Closing Contract 输入遗漏、一次旧 accepted_unassessed 辅助预期失败均保留。r152 mainreceipt 5 项、buildsurface 7 项通过；integration 90 项通过、1 个新增样本断言失败、1 个宿主子进程 helper 跳过，不能称该轮整包通过 |
| r153 | 修正新增样本后最终定向结果 `{"pass": 8}`；覆盖历史失败→新 revision 修复→原 gate/TR-006→S7 Draft、旧产品残留路径、历史 verified 投影、当前未验证拒绝与绑定/Git 反例。r152→r153 仅删除一条过期测试断言，生产源码完全相同 |

历史冻结证据证明 after-r42 与 r153 的 **1650 项完整源码清单逐项一致**；该结论只适用于当时冻结源，不表示当前源码仍与 after-r42 相同。这是 Linux amd64 宿主评估包，不是五平台发布包。

## 新安装包与真实模型记录

| 现场 | 实际结果 |
| --- | --- |
| R34–R37 | 真实 Main 依次暴露 retired S5 责任占用、TASK 指纹未同步、默认草稿 clean 投影缺失、Main 收到 Builder 本人指令与机器预填误报；均保留失败现场后在新包修正，不计完整通过 |
| R38 | 真实 TASK-001 Builder 自测后因服务冷却中断，由 Main 恢复原会话完成提交、检查、合并及清理。真实 TASK-002 首轮漏自测，被 Main 实际 govet 拦下；后续 Main 手工清理错误，现场保留，不计闭环 |
| R39 | 真实 Main 准确理解未交 Result 的进度、问卷责任与路径；仅为理解和进度定向验证 |
| R40 | 前置与两 TASK 首轮为历史专业产物机械重放。真实 Main 使用新增原返工草稿入口、原登记/prepare/launch；真实 TASK-002 r2 Builder 在提交前实际经历 vet/test 失败、修复、通过及 race 通过，再交原 Result。Main 原检查、合并、ACK、精确清理完成，两个当前 TASK 都 done；出口仍因 Git 目录标记解析失败，未计 S7 通过 |
| R41 | 新包重放专业前置并实际重跑消费者、Git 与检查，两个当前 TASK 均已集成。真实 Main 暴露历史失败 checkpoint 被当作必须 PASS 的贡献，已停止并保留现场，未进入 S7 |
| R42 | 全新安装 after-r42；S2–S6 专业答案和产品均机械重放已保留的 R38/R40 产物，工具重新发行本轮上下文、身份和回执，真实重跑自测、Main 检查、提交、合并、ACK 和清理。首轮模型调用 429 未执行工具，恢复原会话后真实 Main 使用原监听，原 `evt-tr-006-r78` 进入 S7 入口后停止；Hook 输出因预算超时未正常投递，不计反馈体验通过 |
| R45 | after-r45 新安装，机械重放专业前置与产品，真实运行四组件 ready、原 r1 失败→revision r2 检查/Git/ACK/清理；旧失败保留。真实 Main progress probe exit 0、38.708 秒，正确识别 govet failed、原恢复/revision 及准确 checkpoint/receipt/log，不要求重交 accepted Result。真实 Main exit probe exit 0、34.595 秒，startup 已给真实 ReviewPlan 路径，两次 Read 成功（protocol→准确问卷），原 TR-006 进入 S7 后停止未填卷。两次 stream 的 startup、Pre、Stop 均无 deadline/UNKNOWN；进程总时长不是单个 Hook 耗时。 |

R40 原生会话白名单审计证明：真实 Builder 的三组 Go 调用（失败、通过、race）都在提交前；Main 使用原恢复入口，没有重复 native Agent、手写 manifest 或 Bootstrap 清理命令。Result 曾遇到宿主 stale-read 和一次不合法引用的正常纠错，不宣称零错误或一次工具调用保存。

R42 的新模型证据只计 Main 原 Hook/TR-006 出口；不把机械专业前置计作最新包两位新 Builder 从零实现。原失败与新成功分别保留，未在旧现场替换二进制、修 Runtime/context 或改 Git 忽略规则。

## 最终核验与尚存限制

R42 原 task receipts 均 accepted/work_complete/integrated；精确 TASK 分母为 2；Main 检查有真实 receipt/log/hash 和 Git merge/祖先链；当前 Worker 目录已由原 receipt 清理。首轮失败 checkpoint/receipt/log 保持原字节，7 份锁定文档不变；此前场景 123 份 Runtime/journal/入口文件哈希及 41 份实际宿主二进制哈希不变。`tasks check / validate --all / doctor` 均 exit 0，独立只读审计没有修改 Runtime/journal。S7 专业工作未执行。

- 前置重放暴露原 Hook 8 秒预算下重复 Git 来源校验的耗时问题：R41 一次观察有近 300 次 Git 子进程。原公开恢复消费者可完成 pending；R42 前置先保留 Hook 超时，随后以原公开 consumer/recover 完成机械续接；另保留并纠正一次恢复脚本的 Worker cwd 错误。该性能缺口单列，未用本轮 S6 出口成功声称 S5 已消除此问题。R42 的 3 个并发 PreToolUse 也触及预算：原 TR-006 已提交，但结构化成功反馈被丢弃；以 journal 的 request_id 与真实会话/PreToolUse 对齐证明出口，该轮反馈投递未验收；R44 startup 仍超时，后续 R45 的对应进展与出口 probe 已无 deadline/UNKNOWN，旧失败不抹除。
- r141 workspace 整包曾因宿主 Bubblewrap 无法创建 uid map 而失败，该隔离正向检查没有被记为跳过或通过。本次普通 S6 使用原空 Worker runner 合同，Builder 实际运行自测，Main 仍真实重跑绑定检查；未修改隔离实现来消除该宿主限制。
- 本次不代表全仓所有测试、旧 release-graph 扫描缺口、S0–S11 全阶段、完整 S2–S10 模型链或实际发布已完成。

## 证据定位

本机评估根目录：`/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z`。关键只读报告：

- `reports/r40-s6-real-model-behavior.json`：真实 Main/Builder 履职、原生日志哈希与安全摘录。
- `reports/s6-rework-r153-history-final.json`、`reports/r152-r153-only-fixture-change.json`：最终回归与生产源码相同的证明；旧失败日志保留。
- `reports/r42-package-regression-source-match.json`：新包/冻结源逐项对账。
- `reports/r42-s6-failed-task-resume-v1.json`、`reports/r42-s6-completed-batch-resume-v1.json`：机械专业前置归属、真实执行命令及原失败证据。
- `reports/r42-s6-main-exit-resume-1.process.json` 与对应 stream/native 日志：本轮真实 Main 原监听与出口。
- `reports/r42-s6-native-hook-exit.json`：原 PreToolUse 与 journal/session 的准确关联证明原 TR-006 提交；同时明示反馈输出超时，且没有手工转移或 S7 专业填写。
- `reports/r42-s6-final-audit.json`、`reports/r42-s6-revision-facts-v3.json`：实际阶段、精确分母、当前回执、检查、合并、清理、历史保全及旧场景零改动。

- `reports/r45-s6-native-progress.process.json`、`logs/r45-s6-native-progress.stream.jsonl`：真实 Main 已知检查失败提示。
- `reports/r45-s6-native-exit.process.json`、`logs/r45-s6-native-exit.stream.jsonl`：真实 Main 入口/出口与准确问卷读取；原 journal 的唯一 `evt-tr-006-r79` actor 为 `hook_controller`，request_id 为本次会话 `17cd1107-d59a-413d-9202-0dddf6d37818`。
