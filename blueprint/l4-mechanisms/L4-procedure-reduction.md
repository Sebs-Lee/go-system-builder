# L4 — 手续收敛

> 层：第四层｜机制域：Agent 面对的命令链、名字、重复载体和占位符如何收成既有门禁能消费的一次写入
>
> 上游：L1 D1/D2/D3/D4 与公理五；[L4 阶段答卷](L4-stage-form.md)；[L4 revision](L4-revision-usage.md)；[L4 Agent 调度](L4-agent-dispatch-governance.md)；[L4 项目级设计基础](L4-project-design-foundation.md)
>
> 状态：v0.1.0 设计。未实现。本文不新增状态机，不放宽既有门禁。答卷篇规定「写到哪」；本文规定写之前还要经过的手续。

## 1. 核对

2026-10-12 对照 `docs/control/agent-protocol.md`、`internal/cli/action_catalog.go`、`internal/cli/repair_command.go`、`internal/cli/investigation_command.go`、`internal/assignment/lifecycle.go`、`internal/review/draft.go`、`internal/review/manifest_draft.go`、`internal/review/plan.go`、`internal/schema/assets/agent-message.schema.json`。六条里有四条按原话成立，两条要收窄。

| 发现 | 核对 | 收窄 |
|:--|:--|:--|
| 一条判断要敲一串命令 | 成立。S7 目录是 `s7 draft` → `runtime review-plan` → `s7 manifest-draft` → `runtime register-workgroup` → `runtime review-result submit`，协议还要求从 `s7 status` 抄 `subject_digest`，冷启动再抄 `s7 workspace-digest`。S8 是 `ingest`、`hypothesis register`、`dispatch`、`hypothesis result`、`consume`、`route`（三份 JSON）、`contract approve`（`--approval-hash` 与 evidence id），另有可选 `project`。S9 的 `impact`、`targeted`、`handoff` 都是 `create` 与 `commit` 两步，前后还有 `session open`、`plan compile`、`dispatch`、`plan-report submit`、`execution begin`、`result submit`、`changeset compute` | revision 篇第 3.3 节已经要求复合动作在一个写入里完成。缺口在命令面还没照做 |
| 同一位置有四套名字 | 成立。协议末尾对照表同时使用 S0–S11、`lifecycle`、`phase`、`TR`/`PTR`。S8 的机器相位是 `bug_resolution.investigation`。S7、S9 两行写明当前相位名仍是兼容投影，与目标名不一致 | 内部游标、journal 和诊断保留机器名。收的是给 Agent 的恢复包和阶段正文 |
| 同一事实写了两到三遍 | 部分成立。完成报告 schema 将 `activation_id`、`requested_event` 列为告警必填，协议要求 Agent 对 readback 文件跑 `shasum` 填 `approved_readback_sha256`。S10 同时要 Markdown、审计 manifest 和带 `audit_manifest_sha256` 的 `evidence add`。S8 的 `project` 在批准后仍出现在 `next` 段落里 | REQ → 设计 → 契约 → TASK 这条链是追溯结构，保留。收的是全文复述，以及要 Agent 亲手写的派生字段和兼容投影 |
| 生成稿里有会拒收的占位符 | 部分成立。`DraftPlan` 与 `DraftManifest` 写入 `TODO(planner)`，`ValidatePlan` 和身份校验拒绝该字面量。manifest 注释写明控制面能证明的字段预填，其余留占位 | 两类分开。控制面已经知道的值由工具写入或在派发时盖上。真正缺的上游（例如没有可读的 `cases.json`）由门指出缺的输入。留给 Agent 的判断用答卷空位，空位是 `null`，不用会像正文的占位词 |
| 派发是另一台状态机 | 成立。`lifecycle.go` 的正式事件是 12 个：`readback_started`、`readback_submitted`、`understanding_approved`、`understanding_rejected`、`document_conflict_reported`、`activation_sent`、`work_started`、`completion_reported`、`completion_acknowledged`、`work_blocked`、`blocker_resolved`、`shutdown_approved`。协议仍要求 `PLAN_REPORT` 走 SendMessage，并同时说明 `SubagentStop` 与 `TeammateIdle` | `plan_checkpoint` 已有自动链，这 12 个事件留在 Writer 内。Agent 路径是派发和把结果写入答卷 |
| 进环前要先走分类树 | 成立。`agent-protocol.md` 的 S0 第 1 步把 UI 影响、投入级别、thin/full、Foundation 状态写在同一个动作里 | 分类表的权威仍是设计基础篇。S0 只问人两件事，其余由 `design-foundation` 计算并给出待确认稿 |

## 2. 留下的判断

下列能力继续由原来的门负责：

- 交付、质量、端到端各自给出结论，审查者不审自己写的实现。
- 表象先形成证据充分的因果模型和批准后的 RepairContract，再改产品。
- 需求锁定和发布决定由人做出。
- 答卷写完只表示判断已提交。空位仍空、路径读不到、发现没有可定位的观察、实现超出任务范围，门仍不过。

## 3. 一次写入

每个阶段对 Agent 只有两个入口：派发，以及写入该阶段那一份答卷。`next` 里的句子就是这两个入口之一。

S7 的覆盖、Claim 划分和 digest 在发卷时从当前 Runtime 与已登记文档算好，写在卷外的只读说明里。审查者填的是结论、发现和适用性理由。写入时 Hook 绑定 `subject_digest` 与 workspace digest。

S8 的假设、因果、影响面、检测缺口和路由是同一份调查卷里的空位。批准 RepairContract 时，Hook 计算草稿哈希并登记一次性 evidence。三份旁路 JSON 和 `--approval-hash` 不再是 Agent 步骤。

S9 里 Agent 要写的是修复计划、修复结果和定向重验结论。`changeset` 由工具按会话基线计算。`impact`、`targeted`、`handoff` 的 create 与 commit 收成同一次写入：卷里的判断就是提交内容，Writer 在同一事务里绑定当前 Session、Contract 和路径哈希。`targeted resume` 仍是受阻之后的恢复动作，由 `repair status` 的 `next` 点名。

诊断命令保留：`s7 status`、`runtime investigation status`、`runtime repair status`。它们的第一行是当前 S 阶段和唯一下一动作。

## 4. 一套对外名字

恢复包、Milestone 和阶段 Skill 只出现两样东西：当前 `S00`–`S11`，以及第 3 节的那一个入口。

`planning`、`verification`、`bug_resolution`、相位名、`TR-*`、`PTR-BUG-*` 留在 `loop-definition.json`、journal 和 `doctor`。S7、S8、S9 的对外阶段以 Main Spine 为准，不再把「目标名」和「当前兼容相位」同时交给 Agent。

`status` 展开项可以附带机器游标，放在下一动作之后，供诊断阅读。

## 5. 一个手写载体

答卷是 Agent 的手写入口。Markdown 报告、证据信封和兼容 BUG 在写入通过校验后由工具生成，供人读和旧消费者使用。

从 Agent 必填字段中去掉这些派生量：`message_id`、`correlation_id`、`activation_id`、`requested_event`、`approved_readback_sha256`、`subject_digest`、`verification_artifact_digest`、`audit_manifest_sha256`、`approval_hash`。身份和哈希按答卷篇第 4 节，由 Hook 从路径和当前 Runtime 写入。

证据引用只收仓库路径或捕获 id。`path:#sha256=` 这种由 Agent 拼接的形式不再作为交卷格式。

S2、S3、S4 的卷记录本层新增的决定和指向上一层的路径。契约层记录分端责任与协议差异，任务层记录范围、收尾检查和依赖。已在上一份卷里写过的目标、约束和验收不再整段抄入。S5 继续核对链接和增量是否与锁定基线一致。

`runtime investigation project` 留在兼容命令里，不出现在 S8 的 `next` 中。

## 6. 生成稿只留空位

发卷时分三种格子：

| 格子 | 写入 | Agent |
|:--|:--|:--|
| 控制面已有的值 | 终值。文档路径、版本、代数、轮次、digest 在生成时绑定。`agent_id` 在真实派发时盖上，草稿里不出现该字段 | 不改 |
| 缺上游 | 不生成看起来已完成的计划。回执写明缺哪份输入，例如没有可读的场景清单 | 先补上游 |
| 要判断 | `null` 或空对象。Claim 的断定、不适用理由、发现的 expected 与 observed | 写入答卷 |

`TODO(planner)` 和 `PLANNER-REFINE` 不再写入生成稿。注册校验继续拒绝这两个字面量，用来守住旧草稿。

## 7. 派发收成两步

12 个 Agent 事件继续由 Writer 推进。`plan_checkpoint` 的自动链保持为内部实现。

Agent 看到的派发是：主会话按 `next` 派出责任；执行者读派发说明，做完后把结果写入本阶段答卷里属于自己的那一块。Hook 从这次写入和子代理返回取得计划与完成，并推进内部事件。

高风险任务仍可要求人看计划。该要求出现在 `next` 里，作为一次明确的等待，而不是第二套激活说明。

Main 的调度动作是派出 `next` 点名的责任，或收下已经写入的结果。普通计划不需要 Main 再回一句批准。

SendMessage 仍可作为平台上的观察点。它不是完成一份工作的必经手续；必经手续是写入答卷。

## 8. S0 只问两件事

设计基础篇继续定义投入级别、thin/full 和 Foundation 状态。S0 的 Agent 动作收成两问：

1. 这次是否改变用户可见界面。
2. 人是否确认这版需求。

`design-foundation` 根据 `project-map` 和现有 `DESIGN.md` 算出投入级别、路径和覆盖缺口，并生成待确认稿。缺口未补上时，锁定动作停住，回执指向那一份稿。Agent 不在协议正文里选择 local、core、extended 与 Foundation 状态的组合。

## 9. 落地顺序

1. 恢复包和三个 status 的第一行改为「当前 S 阶段 + 一个入口」。机器名移到展开项。
2. 答卷写入时绑定 digest、身份和批准哈希。S7、S8、S10 去掉手抄字段。
3. S9 的 impact、targeted、handoff 改为一次写入。
4. 生成稿去掉 `TODO(planner)`，缺上游改为门回执。
5. 派发说明改为「派出、写入答卷」。12 事件留在内部。
6. S2–S4 改为增量空位。S0 改为两问，分类计算留在设计基础命令。

每一步完成前，对应阶段的原门测试仍然有效。手续减少以 Agent 必填字段和必跑动词的数量下降来计，不以删掉检查来计。

## 变更记录

| 日期 | 版本 | 变更 |
|:--|:--|:--|
| 2026-10-12 | v0.1.0 | 按命令面、协议和 schema 核对六条减负发现，写入手续收敛。未实现 |
