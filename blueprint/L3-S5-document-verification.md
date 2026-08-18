# L3-S5 — 文档验证（Document Verification）

> 层：第三层 ｜ 上游：L2 §S5 ｜ 机制事实经调查核实，含 file:line

## 1. S5 是什么，为什么要有它

**S5 做一件事：在写第一行代码之前，让两个没有参与写作的人，各自把整条规格链核对一遍，双人签字后整批冻结。**

为什么值得一个独立阶段：

- S2/S3/S4 的产物是同一条会话链写出来的——写的人容易"自己看着都对"。错误进代码之前改是纸面成本，进代码之后改要烧一整轮构建加验证；
- S6 的 Builder 是照 TASK 干活的——如果 TASK 引用的契约条款根本对不上、或收尾契约要求的证据现有工具做不出来，Builder 会在半路卡死；
- 冻结（原子锁）之后，任何人改这批文档都会被 hook 拦下，必须走新代次——S6/S7 的所有验证都建立在"基线不会悄悄变"之上。

所以 S5 的产出只有三样：

1. **两份审查结论**（一份规格一致性、一份任务可执行性，各来自一个独立审查者）；
2. **一批被这两份结论精确覆盖的文档指纹**（不是"看过了"，是"看的就是这一版"）；
3. **一次原子锁定**（TR-003，之后这批文档只读不改）。

进入条件：S4 出口——契约 locked、任务批 complete、`tasks check` 已过。出去条件：两条 `document_review` 证据 conclusion=pass + 独立性检查过 → TR-003 自动提交。

## 2. 谁参与，各自干什么

S5 只有三种角色：

| 角色 | 是谁 | 干什么 |
|:--|:--|:--|
| 主会话（Orchestrator） | 一直存在的编排者 | 只做一件事：派活。把两个审查职责任命给两个不同的 subagent（并按 REQ/契约特征核对触发条件、在激活信封指名 Triggered Deep-Dives），之后等结论。**不参与审查本身** |
| 审查者 A | 一个 document-verifier subagent | 拿到 **DV-SPEC-CONSISTENCY** 职责：核对"规格链自洽"——REQ 的每条验收标准能走到契约条款，契约之间的数据形状/错误码/状态机在 FE/BE/SYNC 边界上一致，场景包与契约映射不矛盾；外加三项深挖（§3.2 #1-#3） |
| 审查者 B | 另一个 document-verifier subagent | 拿到 **DV-TASK-EXECUTABILITY** 职责：核对"任务能执行"——不是纸面存在性检查，而是**代入 builder 视角回答"我拿到这个任务单能顺利干完吗"**（五问+半问见 §3.1）；外加持表三个触发式专项（§3.2 #5-#8） |

两个职责为什么分开、为什么并行：它们看同一批文档但问的问题正交（"纸面对不对" vs "照着能不能干"）；并行省时滞；互为对方的第二双眼睛。

**独立性怎么保证**（这是 S5 机制的立足点，分两层机器检查 + 一层纪律）：

- 事前（派发时）：team-manifest 必须声明 separation_edges（reason=independence）——两个职责派给同一个 agent 直接被 validator 拒绝；
- 事后（收口时）：gate 机器核对两条证据的 producer 必须是不同 agent，且每条证据的 subject_refs 必须**精确等于**当前 documents[] 的完整指纹集（多一份少一份都拒——防止"看的是旧版"或"只看了一半就签字"）；
- 纪律层：审查者若发现自己参与写过被审文档，stop condition 立即上浮（文档级禁令；author 数据在机器层不区分真实作者，这条不做机器承诺——如实记录，不虚称三层）。

审查者卡片只预载 two-phase-activation 与 document-verification 两个 skill——其余按激活信封指名加载（渐进披露）。诚实限制：两个审查者是同一模型同一套卡片，独立性是**程序性的**（不同上下文窗口、不同职责透镜），不是认识论意义上的双盲——比自审强，但文档不暗示更多。

## 3. 审查者产出什么（一个必产物 + 一个条件产物）

**必产物：document_review_record 证据信封**（`docs/reports/review/REV-<runid>-<resp>.json`，机器消费）
→ 给 gate 看的签字单，激活后第一件事就落骨架。**模板即教师**：REV-template §0 的骨架每个字段带一行"填什么"，agent 复制模板的过程就是读字段含义的过程——引导嵌在必填结构里（D4 彻底形态），SKILL 不再承担字段教学。核心三个字段：

- `producer_responsibility`：自己是哪条职责（DV-SPEC-CONSISTENCY 或 DV-TASK-EXECUTABILITY）；
- `conclusion`：三选一——`pass` / `fix_required` / `req_change_required`（与 gate 同词，全流程无第二套枚举）；
- `subject_refs`：自己核对过的文档指纹清单。

`subject_refs` **必须手动**从 `.claude/loop-state.json` 的 documents[] 逐条复制 `{path, version, sha256}`——这是故意不做命令的：逐条抄写迫使审查者与"我签的到底是哪一版"对峙，这个笨拙动作本身就是审查的锚。自动 scaffold 会把这次深度思考优化掉（写明此处，防未来被"改进"）。review_round 字段模板不列（S5=轮 0，缺省即正确；误填反而静默失配）。

**登记**：信封写盘后必须 `runtime evidence add --kind document_review`（含 `--expected-revision`；重签用 `-r2` 递增后缀新 ID——同 ID 会被拒）——未登记的信封 gate 看不见。详见 REV-template §0 注 2。

**条件产物：REV 报告**（`docs/reports/review/REV-<runid>-<resp>.md`，markdown）
→ 只在有 finding 时才写：给修复者看的过程与定位（P0-P3/哪份文档哪一条/预期 vs 实测/证据路径），N/A 须记理由。**双 PASS 不产 REV 文件**——没有人读一份"都挺好"的审查报告（公理三：无消费者的产物是仪式）。

### 3.1 可执行性五问+半问（TASK-EXECUTABILITY 的深挖标准）

审查者 B 对每个 TASK 回答五个问题+半问——每一问都有判定测试，不是感觉：

**问 1 · 单一职责——一句话测试**：用一句话说出该 TASK 的交付物。说不成一句、或句子里出现"以及/然后"→ 拆。三个具体信号：§3 条款跨多张契约；FE+BE+SYNC 混进同一任务（跨层切换即上下文损耗）；收尾契约的 assert 超过四行（一个交付物不该需要四条以上断言）。

**问 2 · 单窗口可行——compact 是灾难性表现**：如果我是 builder，读完 §2 清单 + 写完 §4 路径，会不会在中途撞 compact？**subagent 中途 compact 丢任务信息 = 杂音与错误开发的头号来源，是灾难性表现**——这不是硬性尺寸门禁（不限制模型的创造力，绕的活该绕），而是拆分者与审查者共持的强警示：能拆小就拆小，能裁清单就裁清单（只引用需要的条款切片）。判定依据：规模直觉锚（必读合计 ~30KB / 触碰 ~8 文件 / 改动 ~400 行）+ `tasks check` 输出的 reference load 参考值——数字是参考不是门槛，审查者判"这个任务会不会真的撞上"。

**问 3 · 语义连贯——DAG 的机器外一半**：`tasks check` 查了无环与引用存在，查不了语义序。三查：**地基先于依赖者**（类型/schema/迁移任务的下游有没有声明依赖它）；**缺失边**（B 任务用到 A 任务的产物却没声明依赖——builder 会读到半成品）；**假边**（复制粘贴来的依赖，实际无读写关系——虚增串行浪费时滞）。

**问 4 · 自包含——防探索浪费**：任务单给的是精确锚点（路径+条款号+行区间）还是"自己去理解模块"？builder 需要 grep 找活干 = 任务书写得不合格，探索烧掉的上下文直接挤占问 2 的预算。

**问 5 · 可测性前向**：收尾契约要求的证据在 S7 三角度（DV/QA/E2E）下真的能产出吗？验收标准本身可判定吗（"响应快"无指标=到 S7 只能靠猜）？

**附加半问 · 批次节奏**：DAG 机器已判无环——再看关键路径有没有单链瓶颈、有没有假依赖把可并行的任务串起来。

五问的分工：全部纯判断（审查者），唯问 2 有机器参考数字（reference load，不产 problem；该数字同时供问 5 判断证据产出成本）。

### 3.2 审查角度全景（S5 是最后一道设计闸）

组织原则：**核心双职责必审 + 按需触发的专项角度**（触发条件来自 REQ/契约的特征——D5 精神：按观测到的风险升级，不做均匀付费）。八个角度的处置：

| # | 审查角度 | 审什么 | 触发条件 | 承载 |
|:--|:--|:--|:--|:--|
| 1 | 需求覆盖端到端 | 每条 AC 沿 REQ→条款→TASK→收尾契约 assert 走通最后一公里（bridge 只到 CASE）——抽 2-3 条走全程 | 恒常 | 职责 A 深挖 |
| 2 | NFR 落地追踪 | REQ 非功能表的每行有没有落进契约条款（NFR 最易静默掉队——Eroding Goals 的经典入口）——没落地的每一行都是 finding | REQ 有 NFR 行 | 职责 A 深挖 |
| 3 | 负向与错误路径完备 | 契约错误码表 ↔ 场景包负向分支 ↔ 权限拒绝用例三方对账（REQ 流程表的权限列有没有负向场景） | API 类契约 / REQ 有权限列 | 职责 A 深挖 |
| 4 | 可测性前向 | 收尾契约证据 S7 三角度真能产出？验收标准可判定？ | 恒常 | 职责 B 第五问 |
| 5 | 迁移与破坏性变更处置 | **默认干净断裂，不默认向后兼容**（owner 裁定：非必须不兼容——兼容是必须辩护的技术债）。审查者查**决策与登记**而非兼容实现：选了兼容的有没有登记（决策理由/影响面/移除路径+责任人）？破坏性变更的数据迁移任务在不在批里？无登记的兼容 = 无声负债，直接 finding | BE 契约含数据模型变更 | 触发式附加段（归 B） |
| 6 | 外部集成韧性 | 每个外部调用点的超时/重试/降级声明；SYNC 把外部错误翻译成本地错误码 | SYNC 契约或 REQ 声明外部依赖 | 触发式附加段（归 B） |
| 7 | 批次节奏 | 关键路径单链瓶颈、假依赖串行 | — | 职责 B 半问 |
| 8 | 风险触发验证就位 | critical 的 REQ——S7 的 risk-triggered 验证维度"座位"在不在（required 分支/PATH 预检；负向齐备性归深挖 #3 不重复） | coverage_profile=critical | 触发式附加段（归 B） |

触发式专项不设第三任命（防职责表膨胀）——主会话派活时按 REQ/契约特征核对触发表并在激活信封指名（expected_outputs 或附言均可）；审查者命中触发条件可**自查自救直接开审**——漏审比越权严重。

**暂缓项**（有意识不做，理由入档）：现有代码库的债务适配——需要读实现代码，超出文档验证职责边界，归 S6 builder 的 Best Practices 触发与 S7 的实现级验证；契约内状态机与持久层一致性——待 BE 契约模板有独立状态机节后再审。

## 4. 机制清单（每台机器管什么）

| 机制 | 管什么 | 在哪 |
|:--|:--|:--|
| team-planning + separation_edges | 派发时挡"两职责同 agent" | team/validator.go |
| two-phase-activation | 审查者先 readback 证明读懂了任务再动手；信封承载触发段指名 | skills/two-phase-activation |
| GATE-DOCUMENT-PASS | 收口判定：两条证据（各职责一条 pass、当前轮）+ producer 互异 + subject 精确匹配 + **registered-document drift 前置筛**（任一当前代条目磁盘 sha ≠ 登记 sha → `document_drift:<path>` conflict——堵"编辑未审文档随批锁入"） | internal/qualitygate/evaluator.go |
| TR-003 | 双 PASS 后由 PreToolUse 自动提交——原子锁定批次进 documents[]；登记 execution batch（空批显式失败） | docs/loop-definition.json |
| TR-004 / TR-005 | 失败路由（见 §5 时间线第 4/5 步）；TR-004 挂 `invalidate_consumed_review_evidence`（消费触发的 fix 记录置 invalid） | 同上 |
| tasks check 机检 | S4 已把覆盖/DAG 把过关——审查者 B 消费其结论，不重算 | loop-harness tasks check |
| reference load 统计 | 每任务必读 KB + 写路径数——**仅供参考不产 problem**（提示词与审查要点形态，不锁死创造力）；目录行按 0 计（自估真实体积） | 同上 |
| 写路径重叠检测 | 两任务 prospective write paths 交集非空且无依赖串行声明 → problem（"串行归属"的机器地板） | 同上 |
| hook 锁定拦截（阶段感知） | 登记即投影进 LockedArtifacts，但写拦截自锁阶段起才激活（contract/task/design=S6）；S5 修复回路（TR-004）可写；旧代次恒锁；wire 路径经共享投影接入 | policy/engine.go；hookctx.LockedArtifactsFromSnapshot |

## 5. 完整时间线（agent 视角，一步一步）

**第 1 步 · 派活（主会话）**
按 team-planning 建两职责任命：两个 document-verifier subagent，分别绑 DV-SPEC-CONSISTENCY 与 DV-TASK-EXECUTABILITY，manifest 里声明 separation_edges；按 REQ/契约特征核对触发表，命中项在激活信封指名。各审查者走 two-phase-activation（readback → 激活信封）。S5 就是三步：派活 → 审查 → 收口三岔路。

**第 2 步 · 落信封骨架（每个审查者，激活后第一件事）**
把 document_review_record 的 JSON 骨架写到自己的 REV 文件（11 字段全列，值可空）。字段即问题——写骨架时就读懂了自己要交什么。

**第 3 步 · 并行审查（两个审查者同时）**
- 职责 A：自底向上读 TASK→契约→REQ→设计，核对验收↔条款映射、跨文档引用指纹、契约间边界一致、场景映射 + 三项深挖（AC→assert 抽样 / NFR 落地 / 负向三方对账）；
- 职责 B：跑 `loop-harness tasks check` 消费机检结论（覆盖/DAG 机器已判，不重算），再审五问+半问 + 触发式专项（信封指名或自查命中）；
- 有 finding 才写 REV 报告（带定位；缺失型 finding 的 Location 填"应出现处"，Observed 记 absent），没有则只在信封收口。

**第 4 步 · 收口（三条路，按结论走）**
- 双 pass：各自登记 document_review 证据（conclusion=pass，subject_refs=当前 documents[] 精确指纹）→ 下一次 PreToolUse：gate 求值（两条证据 + 独立性两查 + drift 筛）→ **TR-003 自动提交，批次锁定**，S5 结束进 S6。agent 不调用任何 transition 命令。
- 任一 fix_required：该审查者信封填 `conclusion=fix_required` + `requested_event=document_fix_required` → TR-004 自动提交（消费的 fix 记录置 invalid），回 planning → 主会话修复被标记的文档 → **受影响职责重新审查，另一职责至少以新指纹重签信封**（任一文档变指纹，两份旧 pass 信封同时失配——subject 全量匹配不区分谁受影响；重签用 -r2 后缀新 ID）→ 回到第 3 步。
- REQ 级歧义（规格链写不出一致解读）：`conclusion=req_change_required`（requested_event 留空——人闸不走自动路由）→ TR-005 → runtime paused（human_boundary）——交人裁决 amendment 或放弃。

**第 5 步 · 冻结生效**
TR-003 提交后，documents[] 中这批条目的写拦截自 S6 起激活——任何人想改这批文档，PreToolUse 直接拒绝，指路 versions/g{N+1}/ 新代次。

## 6. 设计取舍记录（为什么这样、否决过什么）

| 问题 | 选择 | 否决的与理由 |
|:--|:--|:--|
| 审查者几人 | 两人两职责，并行 | 单人通吃——两问题正交且 gate 要求互异 producer；三人加仲裁——现复杂度下无收益 |
| 防自审 | 事前 separation_edges + 事后 producer 双查 + 纪律层 | "author 机器检查"第三层——author 数据不具备（REQ 不写 author，契约/TASK 的登记 author=hook_controller 即执行者非真实作者），机器层恒空转；与其虚称三层，不如如实两层 + 纪律 |
| 防纸面签字 | subject_refs 精确全量匹配 + drift 前置筛 | 抽样引用——签收对象是整批，抽样=给漂移留门 |
| 返工范围 | 指纹失配自动作废旧 pass 证据 + TR-004 的失效动作消费触发的 fix 记录 | 仅指纹失配——不改已登记文档的 fix 会无限重入（活锁）；全量重审——时滞浪费 |
| TR-003 证据槽 | 只要两条 document_review（各职责一条） | contract_set_record / task_batch_record 第三槽——曾声明"三类独立事实"，实际 path 别名允许同一条记录顶三槽（假独立），已删除 |
| TR-004 动作 | invalidate_consumed_review_evidence（精确消费失效） | 无 action（纯指纹机制）——不覆盖"不改登记文档的 fix"；占位桩——无逻辑消费 |
| 结论词汇 | pass / fix_required / req_change_required（REV 与 gate 同词） | REV 自造大写枚举再登记时映射——同一概念两套名字是翻译摩擦与错读源 |
| SKILL 步骤 | 3 必做（落骨架/审/收口登记）+ finding 触发式展开 | 10 步均匀手册——后 7 步是发现后才需要的，预读是均匀付费 |
| 可执行性审到多深 | 五问+半问；"避免 compact"以提示词（S4 拆分时）+ 审查要点（S5）的形态进入流程，机器只出参考数字 | 纸面存在性检查——审不出"builder 拿到干不完"；也否决两类过刚：硬阈值门禁（限制创造力——owner 裁定不取）与 token 精确计数（伪精度） |
| 拆分引导埋哪 | 主战场在 S4 拆分时（TASK 模板字段即问题），S5 是第二道核对 | 全压在 S5 审查——拆错的成本在 S5 才发现，晚了一整轮返工 |
| 字段教学放哪 | 信封模板每个字段带一行"填什么"（模板即教师） | SKILL 手册教字段——预读与填写分离，抄模板时不带走理解 |
| PASS 要不要 REV 报告 | 不要（findings-only） | 每次都写——"都挺好"的报告无消费者，纯仪式 |
| subject_refs 生成 | 手动逐条复制 | 自动 scaffold 命令——会把"签的是哪一版"的对峙优化掉 |
| 审查者卡片预载 | 仅 2 skill（激活信封按需指名其余） | 预载 9 skill——与渐进披露机制自相矛盾的上下文噪音 |
| 阶段叙事 | 三步：派活→审查→收口三岔路 | S5.1-S5.5 子阶段编号——SKILL 从不使用，双义源头 |
| 锁定时序 | 登记即投影、S6 前可写（修复回路）、旧代恒锁 | 登记即全锁——挡死 TR-004 修复；S6 才投影——wire 路径要重投影 |
| 专项角度形态 | 触发式附加段（归 B，信封指名+自查自救） | 第三任命——职责表膨胀；均匀全审——为无风险 REQ 付费 |

## 7. 期望效果与如实缺口

走完 S5：规格链两路独立签收并冻结；自审、纸面对齐、半锁、返工蔓延、礼貌通过、NFR 掉队、断链最后一公里、迁移债各有结构性防线；S6 拿到锁定批次作为读序与授权基准；审查者建立的全链心智可在 S7 转任验证。

如实记录的缺口与限制（供修复清单）：
1. author 纪律层无机器数据支撑（见 §6 防自审行）——若未来要真挡，需 REQ 登记写真实作者；
2. 独立性是程序性的（同模型同卡片、不同上下文与职责透镜），不是认识论双盲——如实声明，不虚称；
3. S5 证据 review_round=0；S7 期间若产生新 document_review，注意轮次校验（模板已删该字段防误填）；
4. S5 fixture 派生化：有机链的 documents[] 注册优先（信封 subject 直接取当前注册指纹）；仅压缩前置场景才写文件+建注册；轮次不强制为 1；证据索引追加而非替换。诚实边界：证据索引条目仍由 fixture 构造（内存态模式，gate 时全量校验）；
5. 词汇映射（三个名字、两个命名空间，REV-template §0 注 4 固化）：信封结论词 `req_change_required`（gate 词汇）＝ protocol 人闸名 `req_amendment`（= `req amend` 命令路径）＝机器事件名 `document_req_change_required`（loop-definition TR-005，机器名不进 agent 读物）；
6. 两条诊断不对称观察（如实不修）：subject 超集静默不合格、延迟冲突在真实堵点为 subject 漂移时可能指错原因——均属诊断噪音，且静默跳过正是"指纹失配自动作废"的预期通道，修反成害。
