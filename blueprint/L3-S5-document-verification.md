# L3-S5 — 文档验证（Document Verification）

> 层：第三层 ｜ 上游：L2 §S5 ｜ 版本 v4.3.1（落地后首轮重走处置：登记步入教学链/requested_event 死值修正/锁定阶段感知/B2 测试兑现；v4.3.0 实施完成：四批次 A-D 落地并全量验证；v4.2.0 机制复杂度/收益二次审计：REV 降为 findings-only / 卡片预载 9 skill 砍到 2 / S5.x 子阶段编号删除 / review_round 出模板 / subject_refs 手动复制的"故意不做命令"入档）

## 1. S5 是什么，为什么要有它

**S5 做一件事：在写第一行代码之前，让两个没有参与写作的人，各自把整条规格链核对一遍，双人签字后整批冻结。**

为什么值得一个独立阶段：

- S2/S3/S4 的产物是同一条会话链写出来的——写的人容易"自己看着都对"。错误进代码之前改是纸面成本，进代码之后改要烧一整轮构建加验证；
- S6 的 Builder 是照 TASK 干活的——如果 TASK 引用的契约条款根本对不上、或收尾契约要求的证据现有工具做不出来，Builder 会在半路卡死；
- 冻结（原子锁）之后，任何人改这批文档都会被 hook 拦下，必须走新代次——S6/S7 的所有验证都建立在"基线不会悄悄变"之上。

所以 S5 的产出只有三样：

1. **两份审查结论**（一份规格一致性、一份任务可执行性，各来自一个独立审查者）；
2. **一批被这两份结论精确覆盖的文档指纹**（不是"看过了"，是"看的就是这一版"）;
3. **一次原子锁定**（TR-003，之后这批文档只读不改）。

进入条件：S4 出口——契约 locked、任务批 complete、`tasks check` 已过。出去条件：两条 `document_review` 证据 conclusion=pass + 独立性检查过 → TR-003 自动提交。

## 2. 谁参与，各自干什么

S5 只有三种角色：

| 角色 | 是谁 | 干什么 |
|:--|:--|:--|
| 主会话（Orchestrator） | 一直存在的编排者 | 只做一件事：派活。把两个审查职责任命给两个不同的 subagent，之后等结论。**不参与审查本身** |
| 审查者 A | 一个 document-verifier subagent | 拿到 **DV-SPEC-CONSISTENCY** 职责：核对"规格链自洽"——REQ 的每条验收标准能走到契约条款，契约之间的数据形状/错误码/状态机在 FE/BE/SYNC 边界上一致，场景包与契约映射不矛盾 |
| 审查者 B | 另一个 document-verifier subagent | 拿到 **DV-TASK-EXECUTABILITY** 职责：核对"任务能执行"——每个 TASK 有可判定的收尾契约、粒度单一、写路径冲突有串行归属、引用的每份文档真实存在且指纹匹配 |

两个职责为什么分开、为什么并行：它们看同一批文档但问的问题正交（"纸面对不对" vs "照着能不能干"）；并行省时滞；互为对方的第二双眼睛。

**独立性怎么保证**（这是 S5 机制的立足点，分两层机器检查 + 一层纪律）：

- 事前（派发时）：team-manifest 必须声明 separation_edges（reason=independence）——两个职责派给同一个 agent 直接被 validator 拒绝；
- 事后（收口时）：gate 机器核对两条证据的 producer 必须是不同 agent，且每条证据的 subject_refs 必须**精确等于**当前 documents[] 的完整指纹集（多一份少一份都拒——防止"看的是旧版"或"只看了一半就签字"）；
- 纪律层：审查者若发现自己参与写过被审文档，stop condition 立即上浮（文档级禁令；author 数据在机器层不区分真实作者，这条不做机器承诺——如实记录，不虚称三层）。

审查者卡片只预载 two-phase-activation 与 document-verification 两个 skill——其余按激活信封指名加载（渐进披露；预载九个 skill 是上下文噪音）。诚实限制：两个审查者是同一模型同一套卡片，独立性是**程序性的**（不同上下文窗口、不同职责透镜），不是认识论意义上的双盲——比自审强，但文档不暗示更多。

## 3. 审查者产出什么（一个必产物 + 一个条件产物）

**必产物：document_review_record 证据信封**（`docs/reports/review/REV-<runid>-<resp>.json`，机器消费）
→ 给 gate 看的签字单，激活后第一件事就落骨架。**模板即教师**：REV-template §0 的骨架每个字段带一行"填什么"，agent 复制模板的过程就是读字段含义的过程——引导嵌在必填结构里（D4 彻底形态），SKILL 不再承担字段教学。核心三个字段：

- `producer_responsibility`：自己是哪条职责（DV-SPEC-CONSISTENCY 或 DV-TASK-EXECUTABILITY）；
- `conclusion`：三选一——`pass` / `fix_required` / `req_change_required`（与 gate 同词，全流程无第二套枚举）；
- `subject_refs`：自己核对过的文档指纹清单。

`subject_refs` **必须手动**从 `.claude/loop-state.json` 的 documents[] 逐条复制 `{path, version, sha256}`——这是故意不做命令的：逐条抄写迫使审查者与"我签的到底是哪一版"对峙，这个笨拙动作本身就是审查的锚。自动 scaffold 会把这次深度思考优化掉（写明此处，防未来被"改进"）。review_round 字段模板不列（S5=轮 0，缺省即正确；误填反而静默失配）。

**条件产物：REV 报告**（`docs/reports/review/REV-<runid>-<resp>.md`，markdown）
→ 只在有 finding 时才写：给修复者看的过程与定位（P0-P3/哪份文档哪一条/预期 vs 实测/证据路径），N/A 须记理由。**双 PASS 不产 REV 文件**——没有人读一份"都挺好"的审查报告（公理三：无消费者的产物是仪式）。

## 4. 机制清单（每台机器管什么）

| 机制 | 管什么 | 在哪 |
|:--|:--|:--|
| team-planning + separation_edges | 派发时挡"两职责同 agent" | team/validator.go |
| two-phase-activation | 审查者先 readback 证明读懂了任务再动手 | skills/two-phase-activation |
| GATE-DOCUMENT-PASS | 收口判定：两条证据（各职责一条 pass、当前轮）+ producer 互异 + subject 精确匹配 | internal/qualitygate/evaluator.go |
| TR-003 | 双 PASS 后由 PreToolUse 自动提交——原子锁定批次进 documents[] | docs/loop-definition.json |
| TR-004 / TR-005 | 失败路由（见 §5 时间线第 4/5 步） | 同上 |
| tasks check 机检 | S4 已把覆盖/DAG 把过关——审查者 B 消费其结论，不重算（机器算过的不再人审） | loop-harness tasks check |
| hook 锁定拦截 | TR-003 后改这批文档直接被 PreToolUse 拒绝 | hookctx/loader.go |

## 5. 完整时间线（agent 视角，一步一步）

**第 1 步 · 派活（主会话）**
按 team-planning 建两职责任命：两个 document-verifier subagent，分别绑 DV-SPEC-CONSISTENCY 与 DV-TASK-EXECUTABILITY，manifest 里声明 separation_edges。各审查者走 two-phase-activation（readback → 激活信封）。protocol 不再用 S5.1-S5.5 子阶段编号——S5 就是三步：派活 → 审查 → 收口三岔路。

**第 2 步 · 落信封骨架（每个审查者，激活后第一件事）**
把 document_review_record 的 JSON 骨架写到自己的 REV 文件（12 字段全列，值可空）。字段即问题——写骨架时就读懂了自己要交什么。

**第 3 步 · 并行审查（两个审查者同时）**
- 职责 A（规格一致性）：自底向上读 TASK→契约→REQ→设计，核对验收↔条款映射、跨文档引用指纹、契约间边界一致、场景映射；
- 职责 B（任务可执行性）：跑 `loop-harness tasks check` 读机检结论（覆盖/DAG 机器已判），再审机器判不了的三件——收尾契约可行性、粒度、写路径串行归属；
- 有 finding 才写 REV 报告（带定位），没有则只在信封收口。

**第 4 步 · 收口（三条路，按结论走）**
- 双 pass：各自登记 document_review 证据（conclusion=pass，subject_refs=当前 documents[] 精确指纹）→ 下一次 PreToolUse：gate 求值（两条证据 + 独立性两查）→ **TR-003 自动提交，批次锁定**，S5 结束进 S6。agent 不调用任何 transition 命令。
- 任一 fix_required：该审查者信封填 `conclusion=fix_required` + `requested_event=document_fix_required` → TR-004 自动提交，回 planning → 主会话修复被标记的文档 → **只有受影响的职责以新指纹重跑**（旧证据因 subject 指纹失配自动作废——不需要显式失效动作，这是机制的省时滞最优解）→ 回到第 3 步。
- REQ 级歧义（规格链写不出一致解读）：`conclusion=req_change_required` → TR-005 → runtime paused（human_boundary）——交人裁决。

**第 5 步 · 冻结生效**
TR-003 提交后，documents[] 中这批条目被 hook 投影为 LockedArtifacts——S6 起任何人想改这批文档，PreToolUse 直接拒绝，指路 versions/g{N+1}/ 新代次。

## 6. 设计取舍记录（为什么这样、否决过什么）

| 问题 | 选择 | 否决的与理由 |
|:--|:--|:--|
| 审查者几人 | 两人两职责，并行 | 单人通吃——两问题正交且 gate 要求互异 producer；三人加仲裁——现复杂度下无收益 |
| 防自审 | 事前 separation_edges + 事后 producer 双查 + 纪律层 | "author 机器检查"第三层——author 数据不具备（REQ 不写 author，契约/TASK 的登记 author=hook_controller 即执行者非真实作者），机器层恒空转；与其虚称三层，不如如实两层 + 纪律 |
| 防纸面签字 | subject_refs 精确全量匹配 | 抽样引用——签收对象是整批，抽样=给漂移留门 |
| 返工范围 | 指纹失配自动作废旧 pass 证据 + TR-004 的 invalidate_consumed_review_evidence 消费触发的 fix 记录 | 仅指纹失配——不改已登记文档的 fix 会无限重入（活锁，批次 B 修正）；全量重审——时滞浪费 |
| TR-003 证据槽 | 只要两条 document_review（各职责一条） | ~~contract_set_record / task_batch_record 第三槽~~——曾声明"三类独立事实"，实际 path 别名允许同一条记录顶三槽（假独立），v4.0.0 删除 |
| TR-004 动作 | 无 action（纯指纹机制收敛回路） | ~~record_document_result 占位 action~~——无逻辑消费，v4.0.0 删除 |
| 结论词汇 | pass / fix_required / req_change_required（REV 与 gate 同词） | ~~REV 自造 DOCUMENT_PASS 等大写枚举再登记时映射~~——同一概念两套名字是翻译摩擦与错读源，v4.0.0 合一 |
| SKILL 步骤 | 3 必做（落骨架/审/收口登记）+ finding 触发式展开 | 10 步均匀手册——后 7 步是发现后才需要的，预读是均匀付费 |
| 字段教学放哪 | 信封模板每个字段带一行"填什么"（模板即教师） | SKILL 手册教字段——预读与填写分离，抄模板时不带走理解 |
| PASS 要不要 REV 报告 | 不要（findings-only） | 每次都写——"都挺好"的报告无消费者，纯仪式 |
| subject_refs 生成 | 手动逐条复制 | 自动 scaffold 命令——会把"签的是哪一版"的对峙优化掉 |
| 审查者卡片预载 | 仅 2 skill（激活信封按需指名其余） | 预载 9 skill——与渐进披露机制自相矛盾的上下文噪音 |
| 阶段叙事 | 三步：派活→审查→收口三岔路 | S5.1-S5.5 子阶段编号——SKILL 从不使用，双义源头（CX-12 F2） |

## 7. 期望效果与如实缺口

走完 S5：规格链两路独立签收并冻结；自审、纸面对齐、半锁、返工蔓延、礼貌通过各有结构性防线；S6 拿到锁定批次作为读序与授权基准；审查者建立的全链心智可在 S7 转任验证。

如实记录的缺口与限制（供修复清单）：
1. author 纪律层无机器数据支撑（见 §6 防自审行）——若未来要真挡，需 REQ 登记写真实作者；
2. 独立性是程序性的（同模型同卡片、不同上下文与职责透镜），不是认识论双盲——如实声明，不虚称；
3. S5 证据 review_round=0；S7 期间若产生新 document_review，注意轮次校验（模板已删该字段防误填）；
4. S5 fixture（SeedDocumentPassS5/SeedDocumentFixRequired）仍手工播种且整体替换 documents[]/evidence——CX-13 A4 只改了 design 播种；spine 的 S5 段依赖它。待按 A4 模式改造（诚实记录，非已完成项）。
5. 词汇映射（三个名字、两个命名空间，已在 REV-template §0 注 4 固化）：信封结论词 `req_change_required`（gate 词汇）＝ protocol 人闸名 `req_amendment`（= `req amend` 命令路径）＝机器事件名 `document_req_change_required`（loop-definition TR-005，机器名不进 agent 读物）。envelope 的 requested_event 在该分支留空（人闸不走自动路由）。

## 8. 落地规划（L4 批次——审查与复验对着本表逐行核对）

设计修正一处（随批次 B 落地）：§6 取舍表"TR-004 无 action（纯指纹机制收敛）"只对"修复改了已登记文档"的情况成立；finding 不改任何已登记文档时旧 fix_required 证据永真、TR-004 无限重入（CX-11②活锁）。修正为：TR-004 挂 **`invalidate_consumed_review_evidence`**（把本次 commit 引用的 fix_required 证据置 invalid）——精确的消费失效，不是恢复占位桩。

### 批次 A：S2 design 出链打通（BUG-CX-13 全部；S5 有机路径的前置）——✅ 完成（6749b1d）

| # | 改动 | 文件 | 内容 | 复验 |
|:--|:--|:--|:--|:--|
| A1 | 登记 | docs/loop-definition.json | PTR-PLAN-01 `actions: []` → `[register_design_documents]` | loop-definition 含新 action 名 |
| A2 | 动作 | internal/transition/actions.go | `register_design_documents`：扫描 `docs/design/architecture/ARCH-*.md`，status=locked 登记入 documents[]（复用 registerDocumentsFromDisk，kind=design） | 新测试：PTR-PLAN-01 提交后 documents[] 含 kind=design |
| A3 | 门回退 | internal/qualitygate/evaluator.go | `evaluatePlanningDesign` 增磁盘回退（复用 diskDeclaredArtifacts：目录 docs/design/architecture，前缀 ARCH-） | 新 pin：磁盘 locked ARCH 满足 gate（无 documents[] 登记） |
| A4 | fixture 去播种 | tests/fixtures/req039/fixtures.go | `SeedPlanningDesignComplete` 删 documents[]/evidence 手工注入 → 磁盘写 ARCH + `runtime evidence add` 生产路径 | grep fixture 无 documents[] 直写 |
| A5 | 有机复验 | spine 测试 | S2→S3 段纯有机（无播种、无手动 transition） | spine 全程 PASS 非 SKIP |

边界：不动 design 指纹算法；不把 design 塞进 contract 登记路径。

### 批次 B：S5 机器链修正（BUG-CX-11 ①②③④）——✅ 完成（daa7a07）

| # | 改动 | 文件 | 内容 | 复验 |
|:--|:--|:--|:--|:--|
| B1 | 删别名槽 | docs/loop-definition.json | TR-003 required_evidence 删 `contract_set_record` + `task_batch_record`，只剩 `document_review_record`（CX-11①方案 a） | 一条 DV 记录不再能顶多槽 |
| B2 | 活锁 | loop-definition + actions.go | TR-004 actions = `[invalidate_consumed_review_evidence]`（按 request 引用的证据 ref 置 invalid） | 新测试：TR-004 commit 后 fix 证据 invalid、不再重入 |
| B3 | 漂移真语义 | internal/qualitygate/evaluator.go | GATE-DOCUMENT-PASS 求值前：当前代任一 documents[] 条目磁盘 sha ≠ 登记 sha → conflict `document_drift:<path>`（堵"编辑未审文档随批锁入"） | 新测试：编辑已登记文档后 gate NOT_READY/Unknown 且点名 path |
| B4 | author 降级 | guard_specs.go + evaluator 注释 | 描述改"数据驱动：author_agent_id 存在时才检查"（代码保留休眠，不虚称三层） | guard_specs 文案与实现对齐 |
| B5 | 手册 | make manual | loop-definition 变更再生成 | loop-harness.md 与 definition 一致 |
| B6 | 测试 | 新增 | B1/B2/B3 各一 pin（见各行） | 全绿 |

边界：不改 TR-005 事件名；不删 evaluator 的 author 检查代码；不动 S6+ 转换。

### 批次 C：引导层落地（BUG-CX-12 全部 + v4.2.0 设计）——✅ 完成（9a97e58）

| # | 改动 | 文件 | 内容 | 复验 |
|:--|:--|:--|:--|:--|
| C1 | protocol 重写 | docs/agent-protocol.md #s5 | 三步叙事（派活→审查→收口三岔路）；删 S5.1-S5.5 子阶段表；actions 首条=落信封骨架；conclusion 词汇 pass/fix_required/req_change_required；词汇映射入档（§7.4：结论词↔人闸名↔机器事件名） | #s5 无 S5.x 编号、无大写枚举 |
| C2 | SKILL v2.0 | skills/document-verification/SKILL.md | 砍到：角色契约 + 3 必做 + finding 触发（REV findings-only）+ 停止条件；字段教学移走 | 步骤数 ≤5、无字段教学内容 |
| C3 | REV 模板 | docs/reports/review/REV-template.md | §0=12 字段骨架每字段一行"填什么"（subject_refs 注明手动复制故意无命令）；删 review_round；结论段小写三值；markdown 部分改 findings-only | 骨架字段集与 evaluator 校验字段一致 |
| C4 | 卡片瘦身 | agents/document-verifier.md | skills 9→2；Output Contract 小写结论词；"不审 authored 维度"标注纪律层（非机器承诺） | frontmatter skills==2 |
| C5 | 模板教真话测试 | 新增 | 按 REV-template §0 骨架填出的信封过 evaluator 全字段校验（模板与机器校验永不分叉——模板改坏即红） | 测试存在且绿 |

边界：不新建 envelope 模板文件（骨架单一居所在 REV-template §0）；不做 `evidence scaffold` 命令。

### 批次 D：收口——✅ 完成（本 commit）

| # | 改动 | 内容 | 复验 |
|:--|:--|:--|:--|
| D1 | BUG 报告处置 | CX-11/12/13 全部写回 fixed + 验证证据 | 三份报告状态=fixed |
| D2 | 索引 | BUG-CX-INDEX 更新 | 索引含本轮记录 |
| D3 | L3-S5 划线 | §6 TR-004 行修正 + 版本 v4.3.0（实施完成） | 本文档版本一致 |
| D4 | 全量验证 | `go test -count=1` + `validate --all` + `doctor` + spine 有机复跑 | 全绿 |

### 执行顺序与全局禁区

依赖：A → B → C → D（C5 模板测试与 spine 有机复跑依赖 A 打通的登记链；B6① 依赖 B1）。

全局禁区（任何批次不得触碰）：
- 不动 S6 及以后任何转换/gate/skill（S5 收口即止）
- 不改 evidence catalog 的 kind 语义（只删 loop-definition 的槽引用）
- 不为"简化"重命名任何机器可见名（transition id / event / kind 稳定；词汇合一只发生在 agent 读物层）
- subject_refs 保持手动复制（设计 §3 已写明故意不做命令）

## 变更记录

| 日期 | 版本 | 变更 |
|:--|:--|:--|
| 2026-08-14 | v1/v2 | 前两版（被判空洞→叙事不清） |
| 2026-08-14 | v3.0.0 | 叙事版；机制事实经 sub-agent 调查核实 |
| 2026-08-15 | v3.1.0 | 新增注意力预算与渐进披露 |
| 2026-08-16 | v3.2.0 | S4 联动：审查消费 tasks check 机检结论；登记欠账划线 |
| 2026-08-17 | v4.0.0 | 减法重做：删 TR-003 别名双槽与 TR-004 占位 action；REV 枚举与 gate 词汇合一；author 机器检查明示降级为纪律层；SKILL 砍到必做+触发 |
| 2026-08-17 | v4.1.0 | v4.0.0 被判"改动说明而非设计"——按讲明白一件事重排叙事，设计内容与 v4.0.0 一致 |
| 2026-08-18 | v4.3.1 | **落地后重走处置（两路 sub-agent + 主会话亲证全部条目）**：①H1 教学链补登记步（runtime evidence add 命令行入模板注 2/SKILL 步骤 3/protocol——此前照教学走必然卡死）；②requested_event 死值修正（req_change 分支留空走人闸；三名词汇映射入档 §7.5，机器事件名 document_req_change_required 不进 agent 读物）；③B2 复验测试兑现（TestTR004InvalidatesConsumedFixRecord——批次 D 台账曾虚记，本轮补齐）；④锁定时序实装阶段感知（policy 比较当前阶段与 LockedFromStage：登记即投影但 S6 前可写、旧代次恒锁——TR-004 修复回路不再被 hook 误挡）；⑤字段口径 12→11/10 如实（四处）；⑥M2 结论错配延迟冲突（有合格者不误报，全不合格才点名——直接版误伤 bug 双 requirement，测试现场抓出）+M3 exactSubjects 差集点名+空 path/sha 逃逸封堵；⑦M1 重跑口径如实（任一文档变指纹两份信封同失效，另一职责至少重签）；⑧spine "纯有机"如实限定 + S5 fixture 播种入缺口 4；⑨C5 补 fix_required 变体 |
| 2026-08-18 | v4.3.0 | **实施完成**：四批次按 §8 逐行落地（A=6749b1d/B=daa7a07/C=9a97e58/D=本 commit），每行复验判据通过；§6 TR-004 行按批次 B 修正；执行中测试现场抓出并修正三处首版错误（裸字符串失效字段、ARCH- 前缀未对齐 protocol 命名、模板占位类型）；spine PASS——如实限定：S2→S4 段纯有机（去播种）；S5→S6 段转换由 hook 驱动但 DV 证据仍 fixture 播种（SeedDocumentPassS5 未改造，见 §7 缺口 4） |
| 2026-08-18 | v4.2.1 | 新增 §8 落地规划（L4 四批次 A-D，含每行复验判据与全局禁区——审查与复验对着本表逐行核对）；设计修正预告：TR-004 需 `invalidate_consumed_review_evidence`（§6 该行随批次 B 修正） |
| 2026-08-17 | v4.2.0 | **机制复杂度/收益二次审计**（owner 指示"考虑复杂度和收益的比率……引导埋在必经之路"）：①REV 报告降为 findings-only（双 PASS 不产 REV——无消费者的产物是仪式）；②信封模板升级为字段级脚手架（模板即教师，SKILL 卸下字段教学）；③subject_refs 手动复制的"故意不做命令"入档（抄写即签收对峙，防未来被自动化掉）；④审查者卡片预载 9 skill 砍到 2（与激活信封的渐进披露对齐）；⑤S5.1-S5.5 子阶段编号删除（三步叙事：派活→审查→收口三岔路）；⑥review_round 出模板（S5=0 缺省即正确）；⑦诚实限制入档：独立性是程序性的，非认识论双盲 |