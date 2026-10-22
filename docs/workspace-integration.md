# 工作区绑定、执行与恢复

> 新 S5/S7/S8 只读责任由 Main 将原问卷准确 view/answer 交 Claude Code 原生 Agent。独立上下文完成专业审查/调查，保存私人答卷并返回引用，SubagentStop 调用原消费者；普通 Worker 工具不接收、不激活。不要求 workspace prepare、bootstrap、平台 Team 或 actor/owner 相等。已登记独立来源继续使用原路径，丢失时不得回落 Main。真实批准、独立性、波次与原 Gate 保留。

> 新 S9 Builder/Targeted 同样交给 Claude Code 原生执行。Main 携带准确 brief；Start/Stop 衔接专业资料与结果。新责任不执行 workspace prepare/launch。Stop 返回后保存在原来源记录中的 `purpose=native_delivery` 仅供原 Result、Main integrate/checks/cleanup 消费，不包含 bootstrap、平台 session、启动租约，也不能被 launcher 启动。下文预建 Worker 与 session 操作仅适用于已冻结的旧 S9 执行；独立 Targeted 的专业独立性与原消费者保留。

> 新 S6 使用 Claude Code 原生 Agent/worktree 和 `s6-native-v1`；参见 [S6 协议](control/agent-protocol.md#s6)。本文 workspace prepare/launch/session/Bootstrap 路径服务已登记旧执行，不再是新 S6 的普通派发入口。不要为同一责任启动两名 Builder。

责任进度按 [`docs/rules/state-machine.md`](rules/state-machine.md#21-loop-responsibility-status)
使用 `pending/active/blocked/done/cancelled`。本文出现的 Runtime lifecycle、
`ready` execution、CLI/receipt 字段和 platform session 值均是当前兼容协议，
须原样使用但不作为第二套责任状态；Worker 不手写状态，按原 receipt 与
Hook projection 判断本人责任，阶段仍由原 Gate 推进。

Main 始终是用户当前项目 checkout，目标分支取绑定时的当前分支。当前是 `test2` 就合回 `test2`；没有通用 `develop` 默认值，不自动 checkout。Main 本身可以是已有 linked worktree，Git common directory 只用于验证仓库身份。

新 install/init 默认启用专业问卷（`packet=true`），普通新流程不需要再传 `--packet`。
`--packet=false` 仅是新 Runtime 的显式 opt-out；已有 Runtime、legacy 格式和 pending
初始化继续使用各自已经冻结的设置，不在原地暗迁移。问卷路径由 Hook 投影，作者不手工
issue/consume，也不从 Runtime 猜路径。

## 主目录准备与执行

```sh
loop-harness runtime workspace bind --root /project
loop-harness runtime workspace prepare --root /project --assignment assignment-example --agent agent-example
loop-harness runtime workspace launch --root /project --assignment assignment-example --agent agent-example
```

对普通 S6 bound-workspace 派发，Main 先用原 manifest 注册 Assignment，再对每个已登记
Assignment 运行 `workspace prepare --root <authority-root> --assignment <id> --agent <owner>`。
读取 prepare 返回的 `execution`、`assignment_context_ref` 和 `submission_paths` 中 PLAN/Result 的实际路径，
然后以同一 root、Assignment、owner 运行 `workspace launch`。Hook 为当前责任再给精确
view/answer 路径；作者照该路径填写专业答案并保存。prepare/launch 创建并登记原 Worker
checkout；这个普通路径不再运行 `team launch`、导出 readback template 或手工建 worktree。
完整可复制 manifest 见 [`docs/examples/s6-builder/README.md`](examples/s6-builder/README.md)。

prepare 消费已经登记的 assignment/owner，接受 building 或 repairing 阶段的 S9 Assignment。S9 使用 Hook 给出的准确 PlanReport/RepairResult 路径保存专业答案，由原 consumer 投影生命周期和候选事实；Main 集成、必需检查、独立 TargetedReverification 与人类闸仍保留。不要把后文列出的低层兼容命令加成默认步骤。审批、派发、S7 领域审核登记等仍走 Main 控制端，原显式审批仍等待其人闸。

派发前检查 staged、unstaged、untracked 输入，存在未提交源码/文档就列出文件并拒绝，不自动 add-all、commit 或 stash。派发记录精确 base_commit、输入摘要、scope、checks、owner 和执行代次；恢复不重新捕获当前 HEAD。先持久化 preparing，再创建 worktree、安装资产，最后 ready。

Worker 安装当前原生二进制、最小 Hook 设置、已安装的角色/技能及 `.claude/inputs` 下的冻结输入文档；不复制 Runtime、journal、凭据或整个 Main `.claude`。Bootstrap 回执记录文件摘要，改变资产后不能直接复用旧执行。

launch 只设置子进程 cwd，不切换主会话目录。启动前登记独立 Claude UUID，持有进程锁防止重复启动，通过安装的 `--agent` 角色执行，默认 30 分钟进程时限。Worker 的可用工具由安装角色配置决定；没有 skip-permissions。PreToolUse 只监听当前问卷并提醒准确的 view/answer 绝对路径，不检查工具权限或 scope，也不拦截写入。需要其他原生工具的任务应按批准流程调整角色配置或执行方案。

重试用同一命令加 `--resume`，只恢复已登记 UUID。若进程在 Claude 创建会话前失败，resume 会如实失败；使用显式 replace 保留旧执行后重新派发。退出码 0 只代表 Claude 进程正常退出，不代表任务、S7 或发布完成。启动尝试日志/回执位于 Main `.claude/workspace-launch`。

## Worker 的受控操作

### S5 独立文档审查

新发行的 S5 packet 可包含内容寻址 `task_analysis`：固定解析器版本、正式
输入集合、TASK读集与粒度 advisory。prepare 将分析及其可读来源安装到
`.claude/inputs/` 并纳入原 Bootstrap 哈希。review-check 的 context/advice
读取原冻结字节，原机械 Problems/覆盖/DAG 仍独立求值。分析不签署 PASS、
不授权写入或下一阶段；来源漂移/原分析丢失须恢复或走原返工。旧已发行包
不会在重试时补分析字段；具名修订保留同一分析而不按新 parser 重算。

新分析可包含 `behavior_projection_version` / `behavior_examples`：仅投影 TASK
Manifest 明确引用的当前 CASE，经原 Scenario 源包与生成结果校验，保留真实
witness 输入、oracle、fixture、七个模块文件及实际校验的绑定 REQ 摘要。Mode、未知条件继续保留；
未决预期交回原规格责任。来源闭包也安装并冻结，漂移拒绝继续消费。这里没有
执行产品测试或浏览器，样例不是 PASS；旧分析不会追加此字段。

新分析还可冻结 `conventions_resolver_version` / `task_conventions`。只消费本人
scope 内有原 Rules 或登记文档证明的选择，保留例外及来源模式/摘要；registry
自称 approved 不构成批准。`candidate/conflict/stale/migration/superseded` 不
作为当前默认。prepare 核验实际安装输入等于冻结分析，接收继续检查原来源；
修订不刷新索引，旧分析不补字段。缺索引不会阻止普通授权工程选择，有影响
验收的决定缺口交原 Architect。参考投影不授予权限或替代专业结论。

Runtime v1.2/v1.3 Builder prepare 的 assignment-context 可包含 `conventions`：按原登记
TASK 与 Assignment 的共同范围求证，真实 Rule/registry/规范/示例摘要进入
原 frozen_inputs，Worker 从 `.claude/inputs/<原路径>` 读取。native 固定
context 同样列出实际来源和 input_read_order，登记锁内复核原提交视图。
S5 context 复用原已发行分析，旧已发行分析/ready execution/context binding
不补字段。该资料只说明执行开始时的参考；本人按原权限改变示例源码仍须
原 checks/report/Main verified integration。冲突/旧例子/未决范围交原责任，
不扩大 write_paths、授权 Skill 或专业结论。

当前问卷已由原消费者发行且 Runtime 启用问卷时，同一 `workspace prepare` 可准备只读 S5 document-review execution。新 install/init 默认 Runtime 1.3 且问卷开启，无需额外传 `--packet`；`--packet=false` 仅用于新 Runtime opt-out。已有 Runtime 或 pending 初始化沿用原冻结配置，不要手改或尝试原地切换。该执行绑定当前阶段、packet、owner 和独立目录代数；准备记录提交后，共享 Main 的同一答卷入口停用。

prepare/launch 给出完整命令及以下路径：

| 字段 | 用途 |
| --- | --- |
| work_packet_ref | `.claude/inputs/work-packet.json`，含原 packet 与不可变 execution view |
| frozen_input_root | packet read[].path 从此目录下读取，题干中的权威相对路径保持原样 |
| answer_path | 本执行独占 `.claude/submissions/stage-answer.json`，只填写 answer |
| review_check_command | 原 TASK 语义检查的只读结果，不执行产品命令、不登记测试 PASS |
| receipt_command | 领域事实的只读回执 |

普通 `plan_checkpoint` S5 只有专业 answer 问卷：读取 `work_packet_ref`，在准确 `answer_path` 填 `answer` 并保存；原 PostToolUse consumer 接收并投影责任状态，不要求 generic PLAN、begin 或手工 answer command。只有当前 prepare 输出确实含 `plan_draft_path`/`begin_command` 的显式 `plan_approval_required` 或冻结 legacy execution 才按其原计划/审批流程继续；`answer_command` 是兼容入口，不是普通保存步骤。

成功 Write/Edit/MultiEdit 保存会返回正式领域接收回执。合法 Finding 经原接收者接受且本人检查满足后，将本人责任投影为 `done`；同事未交、系统 pending 或人工需求决定交给回执指明的 owner。Worker Stop 在 Controller 变更前检查本人责任，已接收或旧进入身份已被原路由取代时可结束。S5 execution 不提供产品 write scope、commit、Builder report、delivery 或 integration。view 和冻结资料计入 bootstrap 回执，答卷不计入其只读摘要；重放不得覆盖答案或 ready 资产。

### S2 Main 设计答卷

启用问卷的 v1.2/v1.3 Runtime（新 install/init 默认开启）在 `planning.design` 由 Hook 根据当前 Main/runtime/generation/stage-entry 自动确保原设计问卷。PreToolUse 提供当前 `view_path` 与 `answer_path` 的绝对路径；Main 读取 view，在原架构文档中完成设计与 locked 声明，只在 answer 的 `answer` 填写 `conclusion=pass` 和真实 `architecture_path`，然后保存。成功的 Write/Edit/MultiEdit 由原 PostToolUse consumer 接收；不需要手工运行 `stage-answer issue` 或 `consume`。这些 CLI 命令保留为显式诊断/兼容入口，不属于正常作者流程。

接收登记原 `planning_design` 和 pending followup，原门决定
`PTR-PLAN-01`，原动作登记架构文档。未答完、REQ 漂移、架构未 locked、
UI impact=unknown 都不能有效接收；需求澄清沿原人闸处理。同内容重放
返回原回执，已接收后中断只恢复 followup。Main 的阶段回执不等于项目
可结束，也不替代后续独立验证。默认旧实例保持原流程。

### S3 Main 契约答卷

在实际 cursor 为 `planning.contracts` 时，Hook 自动确保 S3 答卷；在
`planning.design` 时仍提供 S2 问卷。每次 PreToolUse 按当前 `status`/原 Hook
cursor 返回准确 view/answer 绝对路径。Main 读取冻结资料，只在 answer 的 `answer` 填入：

| S3 字段 | 填写内容 |
| --- | --- |
| conclusion | 明确的专业自审结论 `pass` |
| index_path | 实际的 `docs/dev/contracts/CONTRACTS-<id>.md` 索引路径 |
| contract_paths | 本次完整且适用的 FE/BE/SYNC 合同文件路径数组 |
| applicability | `FE`、`BE`、`SYNC` 各自为 `required` 或 `not_applicable`，并给出与 REQ/locked design 对应的具体 rationale |

`not_applicable` 不是机器推断：必须说明该端为何不适用；`required` 必须与
交付的同类合同相符，整体仍须有至少一份 locked 合同。不要填写 S2 的
`architecture_path`，也不要把 S2 的 UI impact 或设计结论复制成 S3 字段。作者
不填写 producer、source refs、证据 ID、摘要或接收键。

Hook 发行绑定实际 `planning.contracts` 的 Main/runtime/generation/stage-entry；
答卷读取原 locked REQ 和 PTR-PLAN-01 已登记的 S2 design，并冻结按原引用闭合的
模块真相、ADR/共享模型和条件来源。S3 的 CONTRACTS/FE/BE/SYNC 是本阶段的
交付候选；它们要由本次原检查和 PTR-PLAN-02 消费、登记，不能假定发行时已经
被 PTR-PLAN-02 登记。源文件摘要及目录成员清单由发行端从上游原登记与引用解析
重建，不由答卷中的 `contract_paths` 自证完整。必需来源新增、删除、漂移或
stage/baseline 变化时，尚未接收的旧答卷不能覆盖新事实，应按原规划/修复路径
重发。

Main 保存只完成接收请求；原 `contracts check` 和 `scenario bridge` 仍是 S3
机械检查，原 `GATE-PLANNING-CONTRACTS-COMPLETE` 与 PTR-PLAN-02 仍决定推进，
并由 `register_locked_contracts` 登记原合同与 shared inputs。答卷的 `pass` 不
等于 gate PASS。缺条件输入、合同集合不完整、检查失败或需改变 REQ 时，修复
相应源并重跑原检查；涉及 REQ/breaking 决策继续走原 human gateway，不能用
S3 自审批准。PTR-PLAN-02 后的 S5 双审仍独立。

同一个 packet/attempt 的相同答案重放返回原领域回执；不同内容冲突。成功接收
后中断由原恢复流程续作 pending followup，不重复登记 `planning_contract`
或转换。需要查询回执或恢复时，Main 可使用原 `receipt`/`recover` 诊断入口；
手工 `issue`/`consume` 是兼容操作，不属于常规作者步骤。回执说明本次 Main 责任、原 evidence refs、等待 owner/系统动作和下一步；
它不代表 S3 gate 已通过，更不授予结束项目资格。命令、原接收、保存 Hook、
回执与恢复的机制已通过定向验证；实际模型/宿主消费仍需单独验收。

### S4 Main TASK 规划答卷

实际阶段为 `planning.tasks` 时，Hook 自动确保 S4 问卷并提供当前 view 与
answer 的准确绝对路径。读取 view，只在 answer 填写 `conclusion=pass`、精确当前
REQ 的 `index_path` 和完整 `task_paths`；取消项也必须包含。依赖、条款覆盖、
写范围和波次在正式 TASK/索引中声明，原 `tasks check` 校验这些事实。保存登记
原 `planning_task` 与 pending followup；原门和 `TR-002` 决定进入 S5。需要查询或
恢复时由 Main 使用 `receipt`/`recover` 诊断入口，手工 `issue`/`consume` 仅保留
兼容路径，不属于正常作者流程。pending 恢复不修改已经接收的专业判断；原完整
状态恢复保留三个子类型。

S4 首版要求可验证的原 S3 接收和登记来源；旧实例没有这些来源时按原流程
恢复，不从磁盘补造历史或自动迁移。机制已通过定向验证，真实模型验收待做。

### S5 具名修订

已接收判断不能就地修改。Main 在 S5 原进入身份、owner、Assignment 和冻结资料均不变时，以原回执的 `submission_key` 和具体原因调用：

```sh
loop-harness runtime stage-answer revise --root <Main> --previous <submission_key> --reason '<具体要重新审查的观察>'
```

同原回执、同原因重放返回同一新 packet；不同原因冲突。发行立即暂停旧证据资格，修订等待期间旧 PASS 不能推进阶段。新答案接收后原 Writer 原子登记替代关系，原 PASS/FIX/需求人闸路由保持。旧回执仍说明历史交付已接收；兼容字段 `responsibility_status=replaced`、`work_complete=false` 指向新 packet，当前 attempt 只有新答卷经原接收后才可能投影为 `done`。资料变化或已离开 S5 走原返工/需求流程。

独立执行须先结束旧 Claude 会话，Main 再发行修订。旧 execution 与目录留存，新 packet 要求独立源；调用原 `workspace prepare` 为相同 Assignment/owner 建立更高执行代数的新目录，再使用它提供的 view 和 answer 路径。不改旧 view、旧答卷或产品权限，不回退 Main 答卷，也不伪造新批准。相同原责任与冻结链的有效计划/激活授权仍由原校验器检查。

Main 可只读恢复历史回执：

```sh
loop-harness runtime stage-answer receipt --root <Main> --agent <原owner> --packet <原packet_id>
```

旧 Worker 根已退役，不再接受提交或执行命令；历史读取由 Main 查询原领域记录。当前已发行责任不通过手工 `runtime evidence add` 登记另一份审查。新 install/init 默认打开问卷 consumer；`--packet=false` 是新 Runtime 的明确 opt-out，旧实例和 pending 初始化不暗迁移。版本、初始化回执与恢复边界见[安装指南](guides/install.md)。

### 构建与修复执行

prepare 输出原 execution、准确 submission paths 和兼容命令。普通新路径按 Hook 给出的准确 view/answer 路径保存专业答卷；保存时由原 consumer 投影生命周期和领域事实。prepare/Hook 中出现的低层 begin/report/deliver 命令不是额外正常工序：仅在当前 Hook 明确给出恢复动作或 Runtime 明确走未发行的 legacy 路径时才调用。

| 字段 | 输入与行为 |
| --- | --- |
| `submission_paths` 中的 PLAN / Result 路径 | 当前 S6 assignment 的专业答卷路径；优先用 Hook 给出的 exact view/answer 路径保存，原 Writer 自动登记 begin/Result |
| check_commands | 执行已声明检查的零基索引；由当前 Result/检查问题要求时运行 |
| commit_command | 使用返回的 `submission_paths.commit`；只暂存明确文件并正常 Git commit，保留真实 source delivery |
| builder_receipt_command | S6 Builder 的只读当前交付/本人欠项/Main 待消费查询；S9 不使用此完成口径 |
| `begin_command` / `report_command` | 低层兼容入口：仅用于未发行 legacy Runtime 或 Hook 明确指示的恢复，不与已保存 PLAN/Result 重复调用 |
| `s9_delivery_command` | 未发行 S9 RepairResult packet 的 legacy 候选入口；packet-enabled 正常路径保存精确 RepairResult 答卷 |

在 packet-enabled S9 正常路径中，Hook 先给出 PlanReport 的准确问卷路径。Builder 填专业计划并保存；原 PlanReport Writer 投影既有 checkpoint/AgentBegin，图监听再给出下一步。Builder 随后在准确 RepairResult 路径填写专业结论和实际结果并保存；原 consumer 复核登记的实际 Git/source 与必要检查，接收候选并投影既有 completion。无需另交 generic `completion_report`、手工执行 `repair execution begin` 或运行 `workspace deliver`。Main 仍运行原候选集成和全部必需 Main checks；原 ChangeImpact、每个批准 Contract 一份独立 TargetedReverification、Handoff 和 fresh S7 保持。保存 PlanReport 本身不代表修复完成，保存 RepairResult 本身也不等于 Main 集成或独立复验通过。

未发行 typed RepairResult packet 的 legacy Runtime 可继续使用 `repair-result.json` 和 `s9_delivery_command` 原契约；只有该 legacy 路径明确要求时才提交原 generic completion。不要把 legacy 步骤带入 packet-enabled 新流程。

生命周期 JSON 使用既有 agent-message schema。普通 canonical S6 Builder 已在准确的
`submission_paths.commit` 提供原提交请求草稿，并带入起始 Worker HEAD；只填写
`message` 和实际源码文件 `paths`，然后运行返回的 `commit_command`。保存请求本身不会提交，
重新准备也不覆盖已有请求。旧实例的原请求格式如下；HEAD 漂移仍由原入口明确报错：

```json
{"expected_head":"<当前 Worker 的完整 HEAD SHA>","message":"implement assignment","paths":["src/example.go","src/example_test.go"]}
```

commit 拒绝目录、路径越界、未声明暂存文件和 HEAD 漂移。仓库 Git hooks 保持启用，失败保留 index；相同请求可识别已完成提交，避免重复提交。仓库 hooks 属于受信任项目代码，提交入口不是 OS 沙箱；发现 hook 携带未声明文件时保留提交并报错。

Builder 普通 S6 顺序是：保存 Hook 提供的 PLAN answer → 按当前 TASK 实施并运行本人检查 →
按 `submission_paths.commit` 提交明确的源码改动 → 保存 Hook 提供的 Result answer →
Main 集成。原消费者根据保存的答卷登记对应生命周期；无需再运行 begin/report CLI。
源码变更须先按 `submission_paths.commit` 写明确请求，再执行原 `commit_command`；
Result 接收后，其交付源码快照已固定；后续修正走原返工或具名修订。漏提交时由 Main 保留原集成失败 checkpoint；
canonical S6 用原 `register-workgroup --revise-assignment` 具名修订形成新 attempt，
旧格式符合条件时沿原 `workspace rework`。不手改状态或直接 Git 提交。

check 当前支持 Linux amd64/arm64，要求 Bubblewrap、用户命名空间和 seccomp。检查在不超过 512 MiB 的副本中运行，排除 `.claude`/嵌套 `.worktrees`；Main 和 Git 元数据只读，禁止网络及 Unix socket、io_uring。声明命令需要缓存时使用沙箱内 `/tmp`。检查最长 5 分钟，日志/回执保存在 Main `.claude/evidence/workspace-checks`。不复制回构建产物，不把该回执代替实际合并检查。其他系统明确返回不可用，不降级成任意 shell。

Workspace registry 将 Worker 路径关联到原 Main Runtime，Assignment 的 write scope 是原派发契约。PreToolUse 不检查身份、权限或 scope，也不拒绝文件写入；原 PostToolUse/domain consumer 仍按各自语义校验是否接收已保存内容，并把拒绝原因作为反馈。Hook 不是 OS 级任意程序隔离。

### S6 显式 Assignment Record 与薄 PLAN

新的 building 1.2 Builder manifest 可显式声明
`"assignment_record_contract": "s6-v1.2"`，并由原 `register-workgroup` 登记。
登记必须绑定当前锁定 TASK 与完整 reviewed dispatch plan；无标记沿用原流程。
实际角色、Skill、manifest 来自原安装路径，Schema/profile 来自 binary
嵌入字节，目标项目不需要 harness 的 `internal/` 源码目录。

在 `building` 阶段，PreToolUse 从当前 ready、非 DocumentReview 的 workspace
登记或原 native Assignment context 投影 PLAN/Result 问卷，并每次给出准确的
绝对路径。此 Hook 只监听并提醒，不检查工具权限或拦截写入。保存匹配的 PLAN
会调用原 `agent-begin`；保存匹配的 Result 会调用原 `task-complete`。原命令继续
校验专业答卷并推进自己的状态；Result 接收后，Main 的原 Inspect/Integrate 与
checks 仍需单独完成。常规 Worker 不手工运行 `stage-answer issue`/`consume`。

native/workspace context 回指同一 Runtime Record 的 ID、revision 和 SHA，
写入范围与检查要求由记录投影。读取 context 提供的 PLAN 草稿与 Schema，
本人填写 `objective`、`planned_paths`、`assertion_checks`、`dependencies`、
`risks_blockers` 五个专业字段。该显式分支不填 `steps`；依赖与风险的草稿
为 null，确认没有后才写 `[]`。禁止手改机器绑定或扩展有效范围。原
SendMessage(plan_ref)/`agent-begin` 激活链继续校验冻结来源；高风险模式
继续使用原审批闸。

canonical Result 通过原 `task-complete` 接收，填写 `summary`、
`assertion_evidence`、`risks_blockers`、`scope_deviations`。每个 assertion/oracle
须与已接收 PLAN 一致，并给出实际 evidence refs；不能填写机器 checks 或
changed_paths。原 Writer 保存接收时的耐久证据快照并返回 Result ref/SHA、
Result revision、AnswerDigest。相同答案重试保留 message_id，修正答案使用新
message_id；历史重放不改变当前 Result。副本损坏走原恢复，不覆盖旧事实。

此显式 S6 Builder 的 `done_when` 描述本人交付，须能在提交 Result 前验证。
接收回执在提交后证明 Result 注册；Main 的 `required_checks` 仍由集成者实跑。

#### 显式 S6 closing contract

新派发只能由 Main 在原 manifest 明确选择 `closing_contract`：
`contract: "s6-closing-v1"`，并提供 `worker_checks` 数组；同一 assignment
要求同一 assignment 明确 `integration_check_mode: "post_merge"`；原来的 Main
`check_profiles` 必须是 `location: "main"`。原 Register 将这些条件冻结进当前
Assignment Record。`worker_checks: []` 表示原派发明确没有本人 runner checks；
缺少 closing contract 的旧 Record 不补字段，也不回溯成新契约。

closing contract 下，PLAN 的每项 `assertion_checks` 都要填非空、去重的
`done_when_indexes`，引用 Record 中 DoneWhen 的零基索引并覆盖每条 DoneWhen。
写 Result 时使用原 native/workspace context 提供的 `result_path` 和完整
`result_draft`，并查看 `result_questions`/`result_schema_ref`。Record、revision、
Runtime、owner 等机器字段已经由原 context 填好；不要手工重建或修改它们。只
填写草稿中的五个 null 专业字段：`summary`、`assertion_evidence`、
`risks_blockers`、`scope_deviations` 和 `closing_assessment`。

`closing_assessment` 要逐条判断每个 DoneWhen，格式与字段说明见 context：
`disposition` 为 `complete`、`incomplete` 或 `blocked`；这是专业评估结果值，不是责任状态；每个条件带零基
`done_when_index`、`satisfied`/`incomplete`/`blocked` 的 `judgment`、引用已接收
PLAN assertions 的 `plan_assertion_indexes` 和非空 `rationale`。complete 要求每条
判断都为 satisfied；声明风险/阻断或 scope deviation 时本人不能闭合。它是本人
专业判断，不是 runner PASS 或独立验收。

显式空 `worker_checks` 下，完整且来源有效的 complete 判断、无风险/偏差时，原
Builder receipt 可显示兼容字段 `work_complete=true`，将本人责任投影为 `done` 并将 Main integration 单独列为等待。
Main 仍须完成原合并与全部 Main `required_checks`。合法 incomplete/blocked、未决
风险/偏差或本人检查没有当前原 PASS 证明时，Main Inspect/Integrate 与 ACK 保持
阻断，TR-006 not ready。Main machine checkpoint 与本人专业闭合分别投影；
fresh S7 仍由原独立流程决定。

非空本人检查使用 context/prepare 返回的 `owned_check_commands`：workspace
命令选 `--owned-check <index>`，不能与 Main 的 `--check` 混用；native 使用
`runtime assignment-context check` 的冻结索引。命令与 profile 来自 Record，
检查必须经原隔离 runner 捕获，失败也保留原失败回执。缺隔离能力不会切到
未隔离执行。native 当前要求已固定 context、已激活 PLAN 及 manifest 冻结的
实际 worktree；后来 sidecar 中的坐标或审批模式尚不能借此获得授权。

Result 接收保存每项 `worker_check_facts` 和 `tool_closure_digest` 的机器事实及
receipt/log 耐久副本，专业草稿不能填写这些字段。同 message_id/答案重试使用
原耐久回执、零写；新 message_id 才收集当前工具事实。同答案但工具事实变化
会产生新 Result revision，重新等待 Main；内容相同的新键由原 Writer 记录
alias 后返回原 Result，之后 alias 重放也零写。搬迁或 native HEAD 变化后的
新接收需要 fresh 检查，历史 Result 重放无需恢复已清理的 worker。

非空 native runner 实际 fail→Result/alias/改源新版本已通过定向 race；workspace
当前宿主在 prepare 能力门拒绝，未执行完整链。Assignment revision 首个
exact-TASK/plan_checkpoint 切片已接原 Register Writer，要求 fresh Team/Agent；
r2 的 Main 检查使用独立 aN 槽位，旧 checks/ACK 不继承。审批版本绑定、TASK
来源版本变更、完整 S6 及适用宿主 PASS 验收仍未完成。旧合并来源仅支持下述
立即前版的完整审计续接。详细设计和验收证据属于源码仓库维护资料，不随模板发行；
本指南只描述当前可安装的操作和限制。


#### 原 Assignment 修订入口

代码修正时，Hook 给出原重试和具名修订命令；也可用只读
`runtime workspace pending --root .` 查看当前完整指令。Main 通过原准备入口生成
fresh manifest，只选择新的执行人员和修订意图，不重抄原合同或猜清单路径：

```bash
loop-harness s6 prepare --root . --task TASK-001-01 \
  --role backend-builder --agent-id <new-agent> --platform-team-id <new-team> \
  --revise-assignment assignment-task-001-01 \
  --revision-intent-key task-001-01-charter-r2
```

准备返回 `.claude/workgroups/<REQ>/<TASK>/manifest-s6-r<N>.json` 的准确路径，
保留原范围、检查及履职条件，并带出原 Record ref/SHA 的登记命令。它不登记或
改 Runtime，也不覆盖已有草稿。Main 审核后执行返回的原 `register-workgroup`
命令，再按原 `workspace prepare → launch` 派发新 Builder。
直接登记的兼容格式如下：

```bash
loop-harness runtime register-workgroup --root . \
  --manifest .claude/workgroups/manifest-s6-r2.json \
  --task-id TASK-001-01 --task docs/dev/tasks/TASK-001-01.md \
  --revise-assignment assignment-task-001-01 \
  --revision-intent-key task-001-01-charter-r2
```

同意图重试保留 revision-intent-key；改动输入需要新键。当前首个切片要求
same Runtime/baseline、同 exact locked TASK ID/path/SHA 以及 plan_checkpoint。
原 Writer 分配下一 Assignment revision，调用者不能指定。若显式断言旧版本，
`--expected-record-ref`（RecordReference JSON）必须与 `--expected-record-sha256`
成对提供；缺省时由原 preflight 捕获当前前版，锁内再验证。原 Main workspace
锁与 Runtime Writer 锁串行化发行及集成，busy/stale 不产生新版本。

revision 输出额外提供 `assignment_revision_receipt`：原 Record ref/SHA、
revision、原 event/commit 与 current/replayed 标记。历史意图重放返回历史原
发行回执，不把当前 Runtime 提交序号当发行序号，也不回滚现有最高版本。
旧 Owner 的当前 PLAN/Result/ACK 失效，历史字节与源快照保留；新 Owner 重新
取得 context、填写 PLAN/Result，并等待自己版本的原 Main 检查。人闸版本
绑定与 TASK 来源版本变化仍未开放。

新源分支可从立即前版已合入的 source commit 继续提交，但原 Integrator 必须
先验证前版 Record 的完整源快照、精确 superseded Result、原发行 journal、
独立 checkpoint/原检查日志及真实双父 merge 关系。通过后只允许接续集成，
新版本仍生成自己的 merge 并运行全部 Main checks。缺来源、多个旧 merge
候选或没有新提交时保留 pending；可从当前 Main 建立新分支走原正常流程。
前版 checkpoint 和 ACK 不变成新版本的完成凭据。

#### 当前 canonical S6 → S7 接续边界

canonical 薄 Result 不提供机器 changed paths。canonical S7 来源现由原 TR-003
pinned Main-start receipt、原 journal、Assignment Record/Result、版本化 Main
checkpoint、完整 runner/tool receipts 和真实 Git diff 共同重建：`controlproof`
验证原 journal/start receipt，`mainreceipt` 验证单项 Main contribution，
`buildsurface.Read` 汇总当前与历史 Record 对应的合法 contribution 及 start..final
全路径，`ReadRegistered` 再检查 pinned typed ReviewPlan 的完整
`frozen_subjects`。包含删除路径；不能退回作者的 `changed_paths`、`scope_refs` 或
被遗漏的路径。S10 显式 `affected_paths=all` 表示全量审查范围；canonical product
baseline 仍须逐项覆盖机器 `changed_path` 分母，不能省略覆盖行或来源验证。缺失或
漂移仍是不可验证，不视为“无变化”。

TR-006 的 `builder_report_record` 是对原 `completion_report` 的兼容候选 kind，
不是另一个 Builder Result。原 Engine 只在原 Writer 已登记的当前 Result ref 属于
`buildsurface.Read` 验证的当前 Main contribution 时接受；本人闭合、精确 TASK、
共同起点和 Main 事实任一缺失仍拒绝。legacy completion 保持原路径。

canonical S9 的 TR-012 seed bridge 已复用同一 baseline：seed 发布前和原 Writer
变更状态前复验它，再保留完整任务/路径覆盖、DV 与六项 QA 骨架；assertion、oracle、
method 仍由 Planner refine，seed 不判定 E2E 适用性。这是机器来源和 seed 的接线，
不代表 full canonical S9 fresh round 或真实模型审查已完成。typed S7 原发卷/保存
机制链的后续验收见本页 S7 专业答卷节；native/workspace 完整链、安装/平台覆盖
以及真实模型/宿主验收分别记录。当前进度与 canonical surface、S9 bridge 的定向
测试记录保留在源码仓库的 evidence 日志中，不属于安装资产；机制测试通过不能
替代真实模型完整链验收。

### S6 本人交付回执

Builder 报告登记后，Worker 使用 prepare 返回的精确 `builder_receipt_command`；
Main 可查询：

```sh
loop-harness runtime task-receipt --root /project --agent-id agent-example --assignment-id assignment-example
```

回执只读原 TASK/Agent、报告索引/哈希、当前正式 TASK 指纹和原 integration
checkpoint。`delivery_recorded` 与 `accepted`、`work_complete` 分开；本人检查
失败/未运行/缺漏、scope deviation、缺 closing contract 明确列入 `own_pending`。
`checks_source=original_report` 表示读取原报告，不表示查询重新执行过命令。
本人合格但尚未集成时，兼容字段 `work_complete=true` 且 `waiting_on` 指向 Main；
当前原 receipt 与本人检查有效时，本人责任投影为 `done`；
`integrated=true` 另需原真实 merge/verified、当前报告绑定与目标分支可达性。

上述本人合格条件为原通用 completion 的已实现语义。显式 `s6-v1.2` 使用
`checks_source=main_checkpoint`，同时显示接收 Result/PLAN/Record 和耐久证据
引用，以及原 Main 实际 diff、完整 required check receipts。带新 closing contract
时，只有显式空本人检查集合、有效 complete 专业判断且无风险/偏差才可令
`work_complete=true`。合法 incomplete/blocked 或 Main 闭合阻断保持本人欠项；
Main verified 不能代替该判断。缺 closing contract 的旧 Record 保持
`professional_status=accepted_unassessed` 与 `professional_closing_contract_unproven`，
不追补或暗迁移。

登记证据丢失/被改或 TASK 输入改变时，回执失效并指向 Main 的恢复/原返工；不
让作者重复承担系统恢复。Main 经原 `workspace rework` 重新开放失败任务后，旧
结果引用清空，回执显示本次尚未提交，历史文件保留。查询不改变 Runtime、权限、
TR-006 完整批次条件、S7 独立性或发布人闸。S5 仍用 answer receipt，S9 仍用原
RepairResult；这个 S6 查询不能把通用修复报告当作修复完成。

## 合并、S9 与继续流程

```sh
loop-harness runtime task-integrate --root /project --assignment-id assignment-example
loop-harness runtime workspace pending --root /project
```

未指定 `integration_check_mode` 时保留 Worker 预检与 Main 合并后检查；普通任务不默认选择 `post_merge`。新 S6 closing contract 的首个空本人集合路径要求原 manifest 显式选择 `post_merge`，并把 Main profile 固定为 `location: main`；不把 Main 命令复制成 Worker checks。两种模式都检查工作区身份、范围和冻结文档，并且都必须完成 Main 合并后的必需检查。closing 未闭合时，原 Main Inspect/Integrate 保全检查点并阻止 ACK；canonical Result 的路径与 SHA256 一起绑定到回执，Result 改变须重新验证，不能复用旧 PASS。首次接收确认前目标 HEAD 改变也须重验；已经确认后的纯清理重试保留原验证回执。

Main 从原生返回回执及 SessionStart 的当前职责提示消费显式待办，不在 Hook 中同步跑长合并。PreCompact 保持静默，不代收答卷或另外创建阶段 checkpoint。pending 从原报告/checkpoint 生成命令，不建立第二套调度状态。

RepairResult 被原 consumer 接收后，Main 用原集成和必需检查完成候选接收；TargetedReverification 每个获批 Contract 一份，之后原 Handoff 返回 S7。canonical S6 generation 的 TR-012 会按原链路提供 S7 seed，Planner 仍须完成专业 Claim review；Session 累计差异、独立复验和 fresh S7 门禁继续生效。未发行 typed Result packet 的 legacy Runtime 保留原 `repair-result.json` / `s9_delivery_command` 契约。自动投影不代表 Main 集成、独立复验或人类发布授权通过；完整 canonical S9 fresh round 尚未验收。

typed S9 的 `completed` 答卷仍须全部 unit pass、无 scope deviation，并通过原候选与
Main 集成链。`blocked`、`contract_conflict`、`stale` 走原 blocker 接收，回执保持
`work_complete=false`，Worker 等待 Main 指定消费者。Main 的原 resolution 接收后，
可恢复项发行绑定前驱与 resolution 原事件的新 packet/answer；旧答案不恢复产品权限。
`contract_conflict` 仍回原 S8 人闸。相同解阻事件重放不重复接收，不能改填 pass。

已接收 typed Result 后、ACK 前的中断，重试同一 `workspace integrate` 会复验原回执并继续 ACK/cleanup，保留原检查事实。若 Main target 已离开原 TestedHead，命令在改写 checkpoint 前拒绝；先用 `runtime repair status` 和 `runtime workspace integration-status` 查看状态，由 Main 安排新的授权验证尝试。此情形的自动后继尝试尚未接线，不能用 `--retry-preserved` 或仅限未 verified checkpoint 的 `workspace rework` 绕过原回执。ACK 之后的清理重试仍允许 Main 后续合法前进。

失败保留源树/checkpoint。环境问题解决后使用 `--retry-preserved` 重验原候选。若需要修改代码，由 Main 执行 `workspace rework --root /project --assignment <id> --agent <owner> --reason <具体修正原因>`：仅接受未 verified 的失败 checkpoint，归档旧 checkpoint/候选，按实体生命周期把 reported owner/TASK review 退回 working/in_progress，并将原 completion evidence 标为 superseded；完成后 Worker 修改、commit、report，S9 再 deliver 新候选。不会覆盖旧候选或把失败改成 PASS。中断的返工也由 pending 投影恢复命令。已验证交付通过真实 completion_acknowledged 生命周期确认，同一 owner 的其他交付也必须验证。清理失败停在 cleanup_pending，重试继续清理，不强删、不全局 prune。清理完成后，由同一 durable checkpoint 将 ExecutionRegistry 投影为 complete；若 Runtime CAS 前中断，pending 会继续提示 Main 重试。complete 实例不能重新启动 Worker；未发行 typed Result 的兼容实例仍允许 Main 重复集成或报告重验，已接收 typed Result 保持原不可变验证事实。已有 failed 领域结果要走原领域恢复路径，不能通过替换 worktree 抹掉失败。

上述同 owner 的 `workspace rework` 用于旧格式及原 S9 候选。canonical S6 有显式
Assignment Record 时，修正代码／重新报告沿[原 Assignment 修订入口](#原-assignment-修订入口)：
Main 提供 fresh manifest、Team 和 Agent，以新 Record 形成下一 attempt。
原 checkpoint、Result、源树保留；不把 superseded Result 当作当前有效交付，
也不在同一 Record 内回退 reported owner。

## 显式迁移与恢复

所有恢复命令只能由 Main 执行，并要求 `--reason`。

- `workspace adopt --root /project --assignment <id> --agent <owner> --request historical-execution.json --reason ...`：请求是完整 Execution JSON，必须提供真实历史 base_commit、输入摘要、原 branch/path/target、runtime/baseline/execution generation、scope/checks。校验当前登记 assignment、仓库归属和历史差异；拒绝第二份 Runtime/journal。不会把当前 HEAD 猜成历史基线，缺失来源时需重新建立已批准基线。旧 PASS 不升级成新 verified。
- `workspace replace --root /project --assignment <id> --agent <owner> --execution-generation <旧代次> --reason ...`：要求没有活跃 launcher 锁、已登记 completion/candidate/integration checkpoint。保留旧树及 execution_history、原 owner 和冻结基线，新建代次/路径/分支。已有交付必须先走原集成/领域恢复；不能借替换绕过它。
- `workspace rebind --root /new/project --request relocation.json --reason ...`：先由 Git 完成 native worktree move/repair，再重绑 Harness。要求旧 Main 已不存在、明确 runtime/旧新 Main/旧新 common directory/branch/expected_head，以及每个活动和历史 Worker 的路径映射；拒绝仍活跃 launcher 及未完成交付/checkpoint。先保存 before/after 迁移意图，再 CAS 改绑定，随后迁移已完成 checkpoint 的坐标并更新活动 Worker 指针；相同请求可继续中断的更新。该入口不自动搬文件或修 Git 元数据，也不改变目标分支。

relocation.json 示例（路径映射必须完整）：

```json
{"runtime_id":"loop-example","old_main_root":"/old/project","old_common_dir":"/old/project/.git","new_main_root":"/new/project","new_common_dir":"/new/project/.git","branch":"test2","expected_head":"<当前完整 SHA>","paths":{"/old/project/.worktrees/assignment-example-g1-e1":"/new/project/.worktrees/assignment-example-g1-e1"}}
```

Runtime workspace 仍是 schema 1.1.0 的可选扩展，绑定前统一更新二进制；历史数据原样保留。执行新模式前做 Runtime/journal 配对备份，旧二进制不应写入不认识的扩展。回滚程序不能撤销已发生的 Git 合并。

## Hook 与指标维护

原生入口用外层子进程约束整个 Hook 为 8 秒，低于平台配置的 10 秒。完整结果返回前不输出部分 allow；超时/异常的潜在写操作返回阻断码 2，只读和生命周期通知保留诊断通道。中断的 Runtime 事务继续由原恢复协议处理。

计数型指标统一追加独立观察文件，不等待汇总锁。health/doctor、SessionStart 和 PreCompact 执行可恢复压缩：原子记录累计值和已消费文件，再删原文件，避免重复计数。普通 PreToolUse 不做压缩。持续长会话没有这些边界事件时仍应定期运行 health，指标不是业务恢复证据；不能对无限长、完全无维护的会话承诺磁盘恒定。

S5 已识别答卷的保存调用在 health/doctor 显示
`loop_submission_save_total` 与 `loop_submission_save_duration_ms`。
草稿/无效答卷记为 incomplete，完整答卷的领域拒收另记 complete_rejected；
原 Writer 持久接收的 pass、fix_required、req_change_required 分别显示。
计数包含重复保存，不能用它计算唯一交付数或首交成功率；耗时是本次
Save 调用耗时，包含同步接续，不能当作跨进程等待时间或模型用量。
统计写入失败不改变原回执，指标本身不决定权限、完成或阶段转换。

## 有界进展观察

Main 的 PostToolUse 监听 Read/Glob/Grep，Worker 沿用原 matcher。
原 Runtime/派发/答卷/收尾事实与必要资料的实际读取范围相同时，连续
5 次支持的 Read/status 观察记录一次 typed stall。默认 shadow，只保存到
诊断缓存；维护者显式 hook --progress-mode notify 才追加 WORK PROGRESS
提示。安装/Worker Bootstrap 不默认启用提醒。
新资料范围、专业字段或责任范围代码变化重置计数；原运行锁、人闸、
外部责任等待和已接收交付抑制提示。未知范围、未适配工具、缓存损坏、
锁忙或资源超限只标观察不足，不作为空转事实。

缓存位于 .claude/hook-notices/work-progress，每个 owner/session 一个槽，
每槽最多64KiB、64个资料范围。缓存故障时 Main 可核对并移除指定槽后
重试；它不是恢复或业务接收证据。提示只给当前缺口和恢复 owner，
不改变阶段、授权、required 检查或本人完成状态。

当前覆盖 S5 发行答卷、原 Assignment 收尾事实及 Main 原 milestone；
其他阶段的专业缺口 adapter 和真实模型轨迹仍须分别验收。

## 验收范围

临时真实 Git、进程、领域链路、故障恢复测试验证源码行为。真实 Claude S0–S11 和目标 Windows/macOS 运行验收需分别执行；未登录环境与交叉编译不能替代这些验收。检查隔离当前为 Linux 特性，不把其他平台的明确不可用说成支持。发布验收应保存所用版本、平台、操作结果和失败恢复证据。

## Completion report identity

An assignment's explicit `CompletionRef` is authoritative: missing or unreadable
reports block integration and recovery. Neither operation substitutes another
report. Without an explicit reference or a previously bound inspection, discovery
uses only `.claude/evidence/<current-runtime>/g<current-generation>/assignments/<assignment>/completion.json`.
Unknown Runtime/generation and historical or alternate layouts require explicit
recovery; the framework does not scan other runtimes or select the first candidate.
The normal evidence identity/hash and quality gates still apply.

## Worker 观察与检查能力

prepare 输出 `observation_commands`，包含 cwd、status、diff、staged、log 五个固定视图。
例如使用 prepare 返回的完整命令调用 `runtime workspace observe --view log`。
命令已经包含 Worker root、assignment 和 owner，不必自己构造身份。原始 `pwd` 可以使用；
原始 Git shell 被限制时，Hook 返回对应 observe 命令。观察不写完成证据，不改变 Runtime revision。
Git 调用禁用外部 diff/textconv、pager、fsmonitor 和惰性抓取，清除 Git 环境覆盖，最多输出 1 MiB、执行 10 秒。
不接受自定义 flags、重定向或复合 shell；源码文件可正常保存，commit adapter 仍按原规则校验显式暂存文件和提交范围。PreToolUse 不拦截源码写入；原 Assignment scope、consumer 校验与 Main 集成检查继续按各自流程生效。

`runtime workspace prepare --check-location worker` 是兼容默认，prepare/launch 会先探测实际
隔离能力和 512 MiB 检查树上限。网络依赖需预先准备离线缓存。
不适用时可在新执行实例上显式选择 `--check-location main`，由 Main 集成检查通道运行注册检查；
这不放开 Worker shell，不跳过 required_checks、独立验证或发布人闸。
已存在执行实例的 runner 不可切换；没有交付/checkpoint 的实例可按原 replace 约束替换。
replace 默认保留原 runner；可在满足原有无交付、无 checkpoint、无活动 launcher 的条件下显式指定 `--check-location main` 或 `worker`，新代次记录选择，旧代次保持历史。
macOS/Windows 用户应明确选择 Main runner，不能把 Claude 自带 sandbox 等同于此框架的 Linux 检查实现。

rebind 的迁移 intent 保存不可变源快照和请求，只允许路径坐标改变；Runtime/journal 沿用原提交恢复机制。
已完成交付可以迁移，保留 immutable candidate、报告哈希、source/merge/tested SHA 和检查凭据原文；绝对路径按显式映射更新，历史检查 CWD 保存在 checkpoint 中。未完成交付仍拒绝迁移。旧版 ready 条目若对应完整 complete checkpoint，可在迁移后通过独立 Runtime CAS 恢复终态，无需重建已清理 Worker。
同一请求可恢复 intent 已写入、Runtime 已写但 journal 未完成、checkpoint 尚未迁移、终态尚未投影及部分 Worker pointer 更新的窗口。迁移后可重复集成；重新提交 canonical Result 仍须重验，不能凭历史 PASS 放行。
普通 Writer 仍拒绝旧 authority 外的访问；遇到非迁移 pending 事务需恢复原路径处理该事务，不能 force 跳过。

观察 adapter 会在 Git 身份检查前拒绝继承的 `GIT_DIR`、`GIT_WORK_TREE`、`GIT_TRACE` 等 Git 环境覆盖变量；按提示从调用环境清除后再执行。pager 和 external diff 环境设置会被忽略，不能由观察命令启用。

### 冻结派发资料与计划支架（G01 shadow）

新 1.2 workspace prepare 返回 `context_mode=shadow` 和
`assignment_context_ref=.claude/inputs/assignment-context.json`。先读此文件：
它给出实际 owner/assignment/Team/TASK、职责、冻结资料、原技能、检查和收尾条件。
技能候选保留来源与摘要；当前 legacy 加载、角色 frontmatter 与激活权限均保持，
未宣称已经减载或实际加载完成。技术栈没有已批准机器来源时不从空模板猜测。

只有原 Assignment/prepare context 明确提供 `plan_draft_path` 时，才填写该 Plan
草稿的专业字段。S6 Builder 的 objective、steps、assertion_checks、risks_blockers
由本人判断；保存准确路径后，原 consumer 自动执行 PLAN/AgentBegin 校验与投影，
不再手工运行 prepare 返回的 begin。普通 S5/S7/S8 专业问卷不要求另造 generic
PLAN 文件；跟随 Hook 给出的 domain answer 路径。计划草稿在保存并被接收前不是
有效 PLAN_REPORT 或批准；显式 `plan_approval_required` 仍走原人闸。plan/answer
作者文件不进入只读 BootstrapReceipt，重试不覆盖作者字节。

技能、原协议和规则文件均需安装来源可读；缺失由 Main 修复，不回落工厂。
新冻结来源进入原 Execution.inputs/BootstrapReceipt，漂移拒绝继续使用；已有 ready
或 legacy preparing 资产不自动重写。显式 replace 保留原冻结来源/profile，重绑
execution generation；原目录与 context 保留。生命周期提示使用统一800 Unicode
字符预算，完整动作与读入口优先，长 scope 和提示不会挤掉它们。

### 首个 Backend execution 的 core profile

新1.2 S6 backend-builder 可由 Main 显式准备：

```sh
loop-harness runtime workspace prepare --root . \
  --assignment <actual-assignment> --agent <actual-owner> \
  --context-profile backend-core --check-location main
```

默认仍 shadow。返回 context_mode=enabled，固定 context 包含原 profile/hash 和
generated_role 的来源、目标及摘要。新 Worker 的 backend-builder.md 只预加载
四core；额外 selected Skills 由原 Read 工具读安装路径，allowedTools 不新增 Skill。
原角色 tools/model/permissionMode/disallowedTools、正文、manifest/activation
SkillRefs 和 required_checks 保持；Main 角色与安装回执不改。Main runner 的
选择仍走原检查通道，不跳过检查。

既有 ready/preparing 不能通过新选择换 profile；省略参数沿用已固定选择，
显式不同值拒绝。显式原 replace 保留同一生成字节和 profile，仅换 execution
代数。其他角色/阶段/旧1.1不能启用此首版。role/source/hash/symlink 冲突须
恢复原来源，不覆写现有资产。原 begin/commit/report/integrate、专业判断、
独立验证与发布人闸保持。宿主实际消费/token改善需要另行验收。

### Native/teammate 的注册资料视图

没有独立 execution 的已注册1.2 owner 可在 native lifecycle 或 registered
`team launch` 得到 `assignment_context_ref`（path/SHA256）。在原 Main 根读取，
视图列出实际 source_root、agent_state、原 manifest、资料顺序与安装 Skill 摘要。
计划支架和 Schema 与 workspace 使用同一生成逻辑；原 Agent 已 working 时复用
原接受的计划/激活，不因为支架出现就重新提交。已有作者文件不覆盖。

示例/未注册 launch 没有注册资料视图；注册 launch 的文档必须精确匹配原
manifest 引用。缺失、漂移、错误 owner 或 symlink 由 Main 修复。此视图是
shadow 资料，不是 ready、权限或完成证明。已经声明的独立 execution 始终走
原冻结副本，不回落 Main；旧1.1 不增加此资料入口。

### 固定 native/teammate profile 与生成候选

Main 运行 `runtime assignment-context prepare --root . --agent-id <actual-owner>`
将现有1.2注册责任的资料引用/代数/profile登记在原 Agent.context_binding。之后
lifecycle/readback复用原来源，不重新选择当前二进制的profile。原资料漂移或
损坏需要恢复匹配来源；实际Agent状态、计划/激活和完成仍来自当前Runtime，
不能把视图捕获的agent_state当当前授权。已声明独立execution继续原冻结副本。

S5具名修订仅允许 `--replace-generation <old>`，新卷必须指向已接收前驱，
原role/profile/Skill/Schema保持。成功后重试保持原代数，不覆盖作者文件。
绑定亦触发完整pair checkpoint；文档重建不能丢失固定profile后再选新版本。

`runtime assignment-context render-role --root . --role backend-builder` 给出
实际安装来源/SHA、核心Skill列表和生成的角色字节，enabled=false。它只减小
Skill preload块，不修改安装文件或Runtime；实际执行启用和宿主消费仍待接线。

### 已发行答卷的 Runtime 恢复

1.2 Runtime 发行工作包后，原 Writer 会在清除提交/指纹 pending 前保存完整
state/journal checkpoint（包含旋转段），索引只保留最新两份。它是恢复证据，
正常门禁仍只读原 Runtime。快照失败会保留原 pending，修复报告的问题后由原
Writer 重试；不要删除 marker 或把答卷重新登记。

先运行 `runtime recover inspect --root . --req <locked-REQ-path>`，从输出取得
`latest_checkpoint`，然后生成可审核的原 pair 计划：

```bash
.claude/bin/loop-harness runtime recover plan --root . \
  --req <locked-REQ-path> --dev-branch <bound-development-branch> \
  --release-upstream <bound-release-destination> --checkpoint <latest_checkpoint>
```

核对 `checkpoint_restoration` / `original_pair`、候选摘要、责任与激活、修订及
失效记录、原 cursor 与迁移事件后，按原人闸执行
`runtime recover apply --root . --plan <plan-path> --approved-by <human-operator>`。
原合法 pending 优先于计划；半替换后重复同一 apply，不能重新生成接收键。
已处理接续保持已处理，未交的新 attempt 保持未交，旧 PASS 不复活。
计划应用拒绝仍持 launch lease 的 Worker、较新事实、资料漂移和已关闭 Runtime。
现存 Worker 冻结资料与原计划仍需完整可验证；缺少可信完整来源时，不从文件存在猜测
专业签字。单份快照 state 上限16 MiB、完整 journal 上限64 MiB，超限明确保留 pending。

### 发布前真实会话验收

在隔离的新项目中，使用已登录的 Claude Code，记录 CLI 版本、操作系统和 Harness 提交：

1. 在自选开发分支（例如 `test2`）提交输入，绑定 Main；准备并启动两个不同 scope 的 Worker，确认 Main 的目录/分支不变。
2. 在 Worker 执行 `pwd` 以及 prepare 返回的全部观察命令；原始 Git 命令被拒绝时，确认 Claude 能读取 recovery 并改用观察 adapter。
3. 验证本项目所需离线依赖可用、声明检查成功；缺少 bwrap/用户命名空间或检查树超限时，确认派发前失败且没有新增执行/会话记录。通用预检不会推断任意项目的依赖是否齐全；应在采用 Worker runner 前验证项目声明的依赖条件。
4. 完成编辑、begin/commit/report/deliver 与 Main 集成；错误 owner、跨 Worker 和复合 shell 应拒绝。
5. 注入合并后检查失败，确认保留候选和 checkpoint、未放行后继；环境修复后 retry，代码失败采用 rework 新候选，不自动 reset。
6. 停止 launcher 后迁移项目，先修复原生 Git 链接，再按显式映射 rebind；中断并重试，检查 Runtime/journal 与各 Worker pointer 一致。
7. Linux Worker runner、macOS/Windows Main runner 分别记录原生执行结果；未执行项标记待验收，不能以交叉编译代替。

## 检查能力与故障回执

Main 检查模式在原 prepare/launch 前核验实际 shell 与简单字面命令的执行
入口；明确缺失工具不发布 ready，不替 Worker 自动改 runner。它不会执行
检查、安装或版本脚本来探测环境。复杂 shell/变量/控制流保留 legacy 语义，
明确标为 unverified；依赖、配置、fixture 与服务尚未完整 profile 化，因此
入口可用不等于环境全部 verified。

使用 prepare 给出的 `observation_commands["check-capabilities"]` 只读查看
当前环境；同一原 owner/assignment、路径和 Bootstrap 校验仍适用。Worker
观察会走原隔离预检，不能变成未隔离执行入口；支持平台/host 限制如实列出。

实际检查失败回执保留旧字段并附 failure：code/category/owner/check、观察、
恢复/重试条件及 receipt/log。非零退出只说明 process_exit，不能凭输出猜
assertion_failed。取消/超时/诊断 I/O 与 input_drift 归明确责任；原候选与
checkpoint 保留。当前旧检查没有完整环境/profile 键，复用仍为 never，
Main 必需合并检查和独立 fresh S7 必须照原合同执行。

## 显式 CheckProfile

Runtime v1.2/v1.3 Assignment 的原 manifest 可选 `check_profiles`，由嵌入
`check-profile.schema.json` 校验。每个 profile 的 check_index/command 必须
逐字命中原 required_checks，不新增/替换命令；location 必须与新执行选择
一致，timeout_seconds 为1–300秒且更短的外层截止仍优先。旧 ready execution
不补字段，原 replace/adopt/rebind 保留原资料。依赖/fixture路径与摘要进入
原 frozen inputs，prepare/launch 核验真实来源和白名单环境，不运行安装脚本。

`inherited` 保持继承环境；`declared_only` 仅传 profile 的批准白名单事实，
null 表示必须 unset，空字符串不等于 unset。当前 worker/main 的实际回执
profile_observation 采集受测树、runner、环境、依赖/fixture指纹，check_binding
注明原Runtime/generation/Assignment/owner/责任。profile检查产生 v2 回执，
附日志摘要和 receipt_integrity；这只证明完整性，不证明领域批准或可信复用。

当前仍全部实际执行，reuse_eligible=false。不得用自写回执、文件存在或
自重算完整性值替代原领域接收。Main必需合并检查、作者与独立责任、fresh S7
保持；完整 G05 启用复用/并发/中断恢复接线尚需继续。

## 工具回执登记与 shadow 观察

新 profile 回执由实际 runner 捕获内存字节摘要；Worker 通过原 Writer/CAS
登记到当前 Execution，Main 通过原 integration checkpoint 保存可信引用。
登记含原责任绑定、profile、checkout、回执与日志摘要，不增加专业 PASS 或
完成事实。核验拒绝改写后重新封印的回执、日志漂移、未登记来源和责任变化。
旧回执继续供原领域消费，不能自动升级为新复用凭据。

prepare 返回的 `observation_commands["check-reuse"]` 是固定只读入口。
当前所有决策均为 `action=run`；Main 必需检查、独立职责和 fresh 检查照常
执行。Worker 权限/隔离指纹未知或实际预检不可用时返回 unknown_permission_conditions；
即使纯函数 shadow 的完整同责任键匹配，也仍实际执行。原显式 rebind 仅移动
证据坐标，保留原 checkout、回执字节和摘要，移动后的 checkout 不自动命中。

新 Worker 观测记录实际 Bubblewrap、shell/动态库、seccomp/argv policy、
platform/kernel、uid/gid/groups 和原 scope/冻结输入/Bootstrap 摘要。采集与
原执行共用同一隔离参数和过滤器；真实预检失败不切换 runner、不启用复用。
旧回执没有 permissions 字段时仍按原领域消费，但不能作为新 shadow 命中来源。

首版闭合静态语法仅 `true`、`false`、`test -f <相对文件>` 和 `test -s <相对文件>`。
路径必须位于实际受测树内；任意 shell/外部程序不能仅凭 whole_tree 声明获得
命中资格。只读投影使用真实检查的临时副本算法，包含实际归一化目录模式。
缺省仍为 shadow，所有原检查继续实际执行。

## 显式静态复用与进行中诊断

新 manifest profile 可选 `reuse_mode: enabled`，仅允许原 Worker 的闭合静态
检查、same_input_same_responsibility、declared_only 和无副作用。缺省/旧 profile
仍 shadow；此字段属于冻结摘要，旧 shadow 回执不能升级。第一次仍实际执行，
之后只有当前实际输入、权限和原已登记成功回执完全一致时返回原引用。
不产生新 PASS；Main 必需检查、独立判断、S9/fresh 均实际重跑。

同责任/profile/源树请求通过进程锁串行。竞争请求返回原 run.json 诊断引用，
原 Main integration 锁和 Writer 接收仍保持。`observation_commands["check-running"]`
只读查看至多64项诊断；它不是存活证明、完成或接收凭据。进程死亡后重新
取得锁会保留 interrupted 记录，不从未完成文件推断 PASS。取得锁后重新读取
原 Runtime/执行/owner，再核对实际受测副本；损坏或缺失回执只能重跑。
## S6 接收优先建议（G06 首切片）

Main 可运行 `loop-harness s6 status --root . --advice --capacity 3 --json`；
`3` 应替换成平台实际总槽位。输出保留原看板，并追加 `reception_advice`。
它从原 Assignment、完成报告与 integration checkpoint 重建待接收动作，
优先接收能释放更多直接消费者的结果，同分按 TASK/Assignment ID 排序。
`current_action.argv` 是原 pending 的具体入口；依然由原集成检查和 Writer
核实当前权限与结果。收到返工问题时先填写实际纠正理由。

省略 capacity 时预算为 null；显式零表示当前无槽位。
`--max-unreceived-results 1` 只收紧 `recommended_tasks` 的建议预算，
不改 `next`、注册门或 required 集合。未接收、失败、verified 待确认、cleanup
分别计数；cleanup 不计入这个上限。该参数尚未成为 Runtime 持久配置或
注册反压门。Runtime 在读取期间变化会要求刷新；建议仍可能在执行前过时，
执行原命令时继续复核。checkpoint 状态不证明进程存活。

## 新 Runtime 的显式持久反压

初始化或全新安装时可选择：

```sh
loop-harness init --root . --dispatch-max-unreceived-results 3 --dispatch-resume-unreceived-results 1
```

install 接受相同两个参数。max 至少 1；resume 必须小于 max，缺省为 0。
配置随原初始化回执冻结，已有 Runtime、普通更新与 reset 不可修改它。
默认不启用；单次 status 的上限仍只用于建议。

原 Assignment 的未接收结果达到 max 时，新的 S6 Builder 注册暂缓，看板
候选显示 queued；已登记作业照常交付。迟滞持续到原 Main 接收量降至 resume。
在途交付可能超过触发值，不能据此取消作业或缩小 required 覆盖。
原 ACK 登记确认的报告路径与 SHA；匹配当前 checkpoint 才释放，新报告清除
旧 ACK 绑定。cleanup 失败单列恢复，不延长已确认结果的反压。
源 manifest 无法解析时先恢复原登记来源，不能把未知接收量当零。

## 原 Main 集成锁的只读观测

`loop-harness runtime workspace integration-status --root .` 读取原工作区锁的
瞬时状态：busy、available、not_created 或 unknown。它不创建锁或诊断文件，
不改 Runtime、journal 或 checkpoint。锁文件存在不表示正在运行；busy 来自
实际 OS 锁探测，而不是 PID 或旧记录。输出遵循 integration-observation Schema。

在实际 busy 期间，`s6 status --advice` 的当前动作为 observe_integration，给
完整只读入口，继续保留待接收结果与兼容候选。这个信号没有权限或 PASS，
不证明具体 Assignment/进程身份；观察后状态可能立即变化。unknown 给来源
诊断，真正执行仍由原锁和 Writer 校验。普通工具和 Worker Stop/Idle 不新增
扫描或门，Main 只在原指定检查点刷新。

## S7 已发行专业答卷

Main 完成原 ReviewPlan 与专业登记，原准备入口确保当前 dispatched Assignment 的 `s7-review-result-v1` 问卷。波次和依赖保留；未就绪责任不提前取得专业结论。Main 将准确 view/answer 路径带入原生 Agent，Reviewer 阅读真实冻结材料，独立执行其 Claims/lens，填写原 `answer` 并返回准确引用。SubagentStop 调用原 `review.SubmitIssuedResult`，校验当前 Plan/round、owner、来源、Claim exact set、证据与原批准模式。普通工具事件不提供逐保存回执，不要求 generic PLAN/activation。

新只读责任可直接使用原 Main 根下私人问卷，不创建 execution。已登记独立 Worker 使用其确切来源，缺失时不能回落 Main；cold-start E2E 必须如实填写实际执行 workspace 的 `verification_artifact_digest`，原消费者仍验证来源及摘要。显式 `plan_approval_required` 等待原审批。

相同已接收答卷重放不追加 Result/Finding，不同答案沿原冲突/具名修订处理。`runtime stage-answer receipt --agent <owner> --path <issued-path> --root <Main>` 只读恢复事实；普通工作不手动 issue/consume。原 explicit/冻结 teammate 保存入口继续服务其原兼容责任。本文不声明新 S7 原生模型专业验收已完成。

## S8 已发行假设答卷

Main 沿原 Case/Hypothesis 与 `runtime investigation dispatch` 登记工程责任，原准备入口确保准确 view/answer。Main 携带该引用原生启动或恢复 Investigator；逻辑 owner 不要求等于实际平台 actor，登记不代表已经启动。Investigator 阅读真实 encounter/材料，填写本人方法、观察、证据、判别与结论，保存原答卷并返回引用。SubagentStop 调用原 HypothesisResult/Case Writer，普通工具不激活、不接收。

新只读责任不要求独立 execution/session。已登记独立目录沿原来源校验，缺失时不得换用 Main。普通 `plan_checkpoint` 无 generic PLAN；显式批准模式保留原批准。原 Writer 同 CAS 保存结果与回执，无关兄弟结果追加不使有效问卷失效，自身来源/派发/路由变化仍拒收。source boundary 使用实际 `finding:<source-id>` 或 encounter pointer；不编造 ref。

新 Hypothesis/Assignment 由原领域生成，Main 仍须真正派发；答案不批准 route、Contract 或 S9 修复。重复 Stop 沿原幂等接收，改写已接收答案仍冲突。只读 receipt、冻结旧保存入口和原人闸保留。S8 当前实际 lifecycle 为 `bug_resolution/investigation`，残留 S9 pointer 不得抢占它。真实原生模型专业验收另记。

## S9 issued PlanReport boundary

The S9 pre-execution questionnaire binds the original approved Case/Contract,
Session, RepairPlan, actual repair dispatch and its registered TASK/manifest SHA.
When S9 is current, Main's Hook automatically ensures the questionnaire; each
PreToolUse returns its exact view/answer absolute paths. The trusted platform
saves the Builder's three professional fields through the original
PlanReport Writer. The saved answer projects the original checkpoint/AgentBegin;
the graph listener then reports the next concrete step. No generic PLAN_REPORT or
manual execution-begin is a second prerequisite. A receipt is queryable through
the existing `stage-answer receipt` entry; acceptance of PlanReport is not proof
that repair, Main integration, or verification is complete. The Hook later gives
the exact RepairResult path; saving its professional answer runs the original
candidate consumer and completion projection. Main integration, independent
targeted verification, Handoff and fresh S7 remain separate original conditions.

Manual `stage-answer issue`/`consume` and direct Plan/Result lifecycle commands
remain diagnostic/compatibility paths; they are not part of the normal S9 author
flow. The original Report and candidate Writers remain authoritative.

This questionnaire adapter currently supports its registered Main-root source;
it does not make a separate independent Worker execution appear registered.
If the original dispatch selects an independent execution, use its exact path only
when the Hook supplies it; a missing route/path is a Main/system integration gap,
not a reason for the Builder to guess a source or run `workspace deliver` manually.
Source-level mechanism tests and native CLI fixtures do not establish a live
platform Builder run or complete the full S9 adapter.

## Shared submission receipt

`stage-answer receipt` and the platform save Hook return the same closed
`submission-receipt.schema.json` interface. `submission_key` is the original
domain acceptance identity: a generic submission digest, an S7 Result ID, an S8
Case commit key, or an S9 PlanReport/candidate/Result key. It is not always a
SHA256 value. `current_stage` describes the Runtime at query time and does not
identify the historical receipt's originating domain.

`accepted` records original reception; `work_complete` is a compatibility
projection that requires the current owner-responsibility proof. A S9 PlanReport
can be accepted while the responsibility remains `active`. The S9
`candidate_accepted` receipt marks Main integration pending and can project the
Worker responsibility to `done` when its current checks are valid. `main_result_accepted` preserves
Main's final reception with its subsequent gates still pending. The historical
states `main_result_accepted_rework_required` and
`main_result_accepted_owner_status_unavailable` keep `work_complete=false`;
follow their named recovery action. An unreceived candidate returns
`not_accepted`, not a blank responsibility status.

When present, `consumer_owner` and `consumer_pending` appear together. Main's
pending work does not itself reopen the Worker's `done` responsibility or
authorize another stage. Treat `submission_key`, `domain_refs`, and status as
read-only projections; obtain a current receipt through the original command.

## Main S2/S3/S4 planning answer

With questionnaires enabled (the default for a fresh install/init), PreToolUse reads the current lifecycle cursor
and original typed packet binding for `planning.design`, `planning.contracts`, or
`planning.tasks`. It idempotently ensures that stage's S2/S3/S4 questionnaire,
preserves any saved draft, and returns the exact view/answer absolute paths. Main
reads the frozen view, fills only the professional answer and actual artifact paths,
then saves it. The original PostToolUse consumer records the typed submission and
leaves progression to its original gate/CAS. PreToolUse is a read-and-remind listener;
it does not inspect tool permissions or block writes.

Manual `stage-answer issue`/`consume` remains only for explicit diagnostic or
compatibility use, not the normal author flow. `receipt` and `recover` remain available
for queries and pending recovery. The projection consults Runtime's
`entities.work_packets` and existing `Current*` Main binding facts; it does not infer
packets from files on disk, register evidence, or approve a professional answer.
Historical stage entries, other REQs, other owners, and unverifiable bindings do not
expose authoring paths. Packet-off instances retain legacy guidance, and Worker
guidance keeps its original Assignment responsibility.
