# 仿真缺陷清单（2026-10-24 完整流转仿真）

本文件记录 2026-10-24 三题材完整流转仿真（S0～S7～S9～S7～S10～S11）暴露的、**尚未修复**的框架缺陷。

它不是设计文档，也不是调查流水账：每条只记录「现象 → 证据 → 落点 → 建议修法」，供后续修复与 blueprint 同步更新使用。修复完成后应把对应条目连同修复提交一起删除或标记为已修，不在本文件累积历史。

仿真环境位于仓库外 `/home/sebs/gsb-simulation/`，三题材共用同一份发布包：

| 题材 | 形态 | 终点 | 机械校验 | 机制疑问 |
| --- | --- | --- | --- | --- |
| `source-ledger` | NFR-only、8 份仓库现成 fixture 作对照基线 | **`release_authorized`（终态）**，S0→S7→S9→S7→S10→S11 全通 | 12/12 阶段全绿，`fixture_unchanged=true` | 28 条 |
| `ui-ledger` | UI impact=changed、无候选文档 | S8 完成、S9 planning；S9 发卷被内部 deadline 打断，PlanReport 一度无法接收 | — | 11 条 |
| `callback-api` | 服务行为模块、UI impact=none | S8 完成、S9 planning；S7 接收超时 30 次 | — | 11 条 |

`source-ledger` 的 S8 再次在 `no_competing_hypothesis` 上踩到同一个坑（见 BUG-SIM-01）。`callback-api` 的 S7 接收被 Hook 内部预算反复打断（见 BUG-SIM-17），连续三个窗口停在同一 revision；`ui-ledger` 同一缺陷在 S9 变成死路。

原始反馈原文：`/home/sebs/gsb-simulation/{topic}/project/simulation/feedback-notes.md`（每 10 秒快照到 `feedback-snapshots/`；因框架恢复动作会删除该文件，最终以 `feedback-merged/{topic}.md` 为准），汇总见 `/home/sebs/gsb-simulation/feedback-consolidated.md`。

---

## 一、总体结论

### 1. 减负到位了吗

**到位了一半，而且分界线很清楚。**

到位的部分，有运行时证据：

1. **门禁不再需要猜。** Hook 观测文件里**没有一次工具调用被拦**：source-ledger 1105 次观测中非零退出/含拒绝 3 次（0.3%），callback-api 741 次中 0 次，ui-ledger 854 次中 0 次。策略层只出现 `allow` / `audit`，一次 `deny` 都没有。agent 不再猜「这一步会不会被卡」，机器判定是确定的。
2. **流程骨架完全由 Runtime 托管。** 68 次 `transition_committed` 全部由系统提交；阶段问卷路径一律由 listener 生成，Main 从不自造路径或 envelope；Worker 答卷由原消费者在原生 Stop 接收，Main 不代填。
3. **上下文交付基本够用。** 28 条机制疑问里 18 条**无需试错**，直接由 packet 的 `read` 列表、模板、协议段落回答。`callback-api` 的 S0/S2/S3 三条明确写出「不需要试错」「三处互相印证」「规则表达是明确的、可直接引用的」。
4. **编排者能自我补齐约束。** `source-ledger` 在 S7 答卷被拒后，**自己**把一段顶层形状约束写进了后续 S8 子代理 brief（见 BUG-SIM-04），传播有效。

**但「0 次拦截」有前提。** Hook 的**策略**层一次都没拦过人，Hook 的**时间预算**却三题全部越界：S7 原生答卷接收只有 8 秒内部预算，实测峰值 7680 ms（见 BUG-SIM-17）。也就是说，减负在「判定是否合规」这一层做到位了，在「判定跑不跑得完」这一层没有——超时后 Main 只能自己代交答卷，把「原消费者接收」的责任悄悄接了过去。

没到位的部分，集中在同一个模式上：

> **同一个知识在产品里存在，但没有绑定到需要它的那个人身上。**

`planned_paths` 的语义写在 blueprint 里、Worker 的 brief 里没有；答卷顶层形状的约束 Main 自己写进了 S8 brief、框架生成的 S7 brief 里没有；`no_competing_hypothesis` 的真实语义写在 schema description 里、字段名却表达着相反的意思。

### 2. 流转时还需不需要猜机制

**还要猜，但猜的对象变了。**

| | 仿真前 | 本轮仿真 |
| --- | --- | --- |
| 要猜的 | 门禁怎么过、阶段怎么推进、谁负责什么 | 这张表该长什么样、这个字段该填什么语义 |
| 性质 | 流程机制 | 契约说明 |
| 后果 | 卡死、无出路 | 一到两轮返工后可自行恢复 |

两题（`source-ledger`、`ui-ledger`）都出现过**去 `strings` 翻编译后的 harness 二进制找字段名**的行为（分别 4 次和 6 次）。这是「文档没给出字段名」最直接的证据，也是本清单里多条 P1 的共同来源。

### 3. 流程完整性

`source-ledger` 走通完整链路：`inactive → planning/design → planning/contracts → planning/tasks → document_verification → building → document_verification → building → verification/planned → verification/running → verification/cannot_clean ×2 → verification/observation_sealed → bug_resolution/{investigation,planning,fixing,targeted_reverification} → verification/running → verification/clean → acceptance → release_audit → awaiting_human_release → release_authorized`。

修复前在 S6 卡死的阻塞（Worker 按 brief 写紧凑 `assertion_evidence[].plan_assertion_index`，集成侧按顶层 key 逐字段比对）已定位并修复（提交 `823743c`），修复后两题不再复现。

**三题对照（`loop-events.jsonl` 的 `transition_committed`）.**

| 题材 | 转换总数 | 返工类转换 | 走过的机检 gate |
| --- | --- | --- | --- |
| `source-ledger` | 68 | 8（其中 `REVIEW-PLAN-STALE` ×2 + `REVIEW-PLAN-REVIVE` ×2 全部由**夹具把反馈文件放进产品树**引起） | 19 个全过，含 `GATE-VERIFY-CLEAN-ROUND-PASSED` / `GATE-ACCEPTANCE-COMPLETE` / `GATE-RELEASE-AUDIT-APPROVED` |
| `callback-api` | 59 | 4 | 16 个，停在 `S9-TARGETED-REVERIFICATION` 之后 |
| `ui-ledger` | 46 | 1 | 13 个，停在 `S9-PLAN-COMPILE` 之后 |

**三题的机检 gate 序列逐项相同**，一直一致到各自的中断点。差异全部来自内容形状、契约语义与时间预算，没有一处是「流程本身不确定」。这是「减负在流程机制层已到位」最直接的证据——agent 不需要知道下一步是什么，Runtime 知道。

---

## 二、未修复缺陷

严重度口径：**P0** 阻塞流程或产生错误产物；**P1** 造成返工、误判或可避免的往返；**P2** 措辞歧义、易误用；**P3** 文档缺一句可省的试错。

### BUG-SIM-01 · P0 · `no_competing_hypothesis` 字段名与语义相反（✅ 已修复，提交 `866f4be`）

**现象.** S8 因果收口要求「至少一条 `refuted` 结果，或一条 `no_competing_hypothesis` 声明」。校验只判断该字段**非空**，不判断内容真假。而字段名读作「没有竞争假设」，其真实语义是「没有值得写判别式的可信替代机制」——只有当**确无**竞争假设时才该填。案情明明存在一个可信竞争解释（已登记为判别式的对照面），却因为没有 `refuted` 结果而被迫填写该字段。

**证据.**

**复现 1 — `source-ledger`.** 该题的 Case `.claude/review/investigation/cases/investigation-case-observation-batch-r1-r5.json`（status `contract_approved`）中 `no_competing_hypothesis` 的正文为：

> The one competing mechanism that was credible and registered as the discriminator's contrast — … — **was not refuted by any accepted result, so it is not declared absent**; it is instead retained as the surface on which the accepted supported result located the gap …

即：在一个**已批准**的产物里，框架逼着填写者写下与字段名相反的声明。

**复现 2 — `ui-ledger`（独立复现，结论一致）.** 该题的 Case `.claude/review/investigation/cases/investigation-case-observation-batch-r1-r5.json` 同样是 `contract_approved`，`hypothesis_results` 只有一条、`result = "supported"`、`unexplained_finding_ids = []`（因果闭环**实际已闭合**），但 `no_competing_hypothesis` 仍被填成：

> 本 Case 只有一条注册假设，其结论是 supported，没有任何结果被 refuted，因此按证据说明为什么没有其它可信机制：竞争机制（绑定其实存在于设计与契约层，只是未在 TASK 收尾契约与 Main 检查入口中表述）**已被同一判别式的 refute 分支否证**……

**同一字段、两个题材、同一形态的自相矛盾。**

**补充事实（重要）.** 协议给出的替代出路是「把竞争解释写进 `discriminator` 文本，而不是另起一行 Hypothesis」。`ui-ledger` 正是这么做的——竞争机制 (A)/(B) 的区分写在 `counterfactual_or_discriminator` 里，并给出了明确的否证条件。但**这条路并不满足机器收口校验**：收口只看 `result == "refuted"` 的行数或 `no_competing_hypothesis` 是否非空，不会读 `counterfactual_or_discriminator`。于是 agent 明明已经用判别式处理掉了竞争机制，仍必须再写一份「没有可信竞争机制」的声明。

**落点.**

- `internal/investigation/case_workflow.go:1317-1339`（`competingHypothesisState` / `requireCompetingHypothesisClosure`）
- `internal/schema/assets/review-investigation-case.schema.json:71`（字段 description）
- `internal/investigation/case_workflow.go:119-122`（Go 侧注释）
- `blueprint/l3-stages/L3-S8-finding-investigation.md:155-159`

**建议修法.** 二选一，不要靠措辞糊过去：

- 拆字段：把「确无竞争」与「竞争已由判别式处理」分成两个字段，校验分别识别；
- 或改语义：字段改为记录「竞争假设的处置方式」（含判别式 ID 与结论），校验只要求「存在判别式结果或已声明无竞争」，字段名同步改。

无论哪种，都要让校验能区分「确无竞争」与「有竞争但已被判别式处理」，而不是靠字符串非空。**并且要让「竞争机制写在 `counterfactual_or_discriminator` 里」这条协议出路真正满足收口校验**——现在它不满足，agent 只能两头都写。

**关联反馈.** `source-ledger` F-021；`ui-ledger` S8 Case 记录（未单独记入反馈文件，由本次分析从产物中读出）。

---

### BUG-SIM-02 · P1 · S8 route 问卷指令与消费者引用语义冲突

**现象.** S8 route 问卷的运行时指令原文要求「reuse the fixed reference supplied by the accepted Case result」。但被接受的 Case 结果里的 `path:` 引用是 **S7 捕获**的证据；S8 的消费者要求证据必须由**本阶段消费者**捕获。照抄被接受的引用会被拒。

**证据.** `source-ledger` F-020；被拒的实际引用为 `path:simulation/observations/first-round.txt#sha256=1277…`，来源是 S7 Finding 的固定摘要，不是 S8 捕获物。

**落点.** `internal/cli/investigation_lead_disposition.go:170`（`investigationLeadRouteNextContext` 生成的问卷指令）：

> Reference array shape: {"evidence_refs":["path:<actual-source-path>#sha256=<actual-sha256>"]}; **reuse the fixed reference supplied by the accepted Case result** or an applicable registered ID displayed with the sources.

**建议修法.** 指令改为：`path:` 引用必须指向**本阶段消费者已捕获**的证据；跨阶段继承请使用登记 ID。保留「不要自己计算 path/digest」这半句。

**关联反馈.** `source-ledger` F-020。

---

### BUG-SIM-03 · P1 · S6 brief 未定义 `planned_paths` 语义

**现象.** S6 brief 给出 `Plan` 的字段清单，但只对 `assertion_checks` 给了构造规则，`planned_paths` **只有字段名没有语义**。Worker 第一次按「全部涉及文件」填写，被原消费者拒绝后靠拒绝信息反推语义，原生 resume 重交才通过。

**证据.** `source-ledger` F-013（brief 给字段清单，却没写 `planned_paths` 是「写域」还是「全部涉及文件」）。

**落点.**

- `internal/cli/native_s6_brief.go:13`（`nativeS6LegacyProfessionalFormat`）
- `internal/cli/native_s6_brief.go:22`（`nativeS6CompactProfessionalFormat`，运行时实际使用的紧凑写法，见 `internal/cli/native_s6.go:134`）
- `internal/cli/native_s6_brief.go:24`（`nativeS6ProfessionalFormat`）

**语义其实已存在.** `blueprint/l3-stages/L3-S6-build.md:186`：

> | `planned_paths` | 实际准备改动的产品或测试位置 | 必须落在登记范围，不能扩张 TASK |

Worker 不会读 blueprint，所以这句对 Worker 无效。

**建议修法.** 三处 brief 统一补一句：`planned_paths` = 该 Assignment 的**写域**（实际准备改动的位置），不含读集，且必须落在登记范围内。

**关联反馈.** `source-ledger` F-013。

---

### BUG-SIM-04 · P1 · S7 答卷顶层形状约束缺失（S8 已有等价先例）

**现象.** S7 的 QA Reviewer 把 **view packet 的形状**写回了答卷，顶层 key 变成 `packet_id, template_id, template_sha256, binding, inputs, result_id, issuance_event_id, view_path, answer_path, draft, answer`，被原消费者以 `unknown field "packet_id"` 拒收。

**证据.**

- `source-ledger` F-019；实际拒收消息：`S7 answer /: one closed S7 ReviewResult questionnaire (json: unknown field "packet_id") 保留已发行题干，只填写 answer 中的专业判断与观察。`
- 拒收消息只说「保留已发行题干」，没说顶层 shape 约束。
- **关键对照**：`source-ledger` 在 S8 派发 Investigator 时，**自己**在 brief 里补了一段：

  > CRITICAL FILE-SHAPE REQUIREMENT: your answer file must keep the issued questionnaire's top-level shape — exactly `title`, `instruction`, `read`, `questions`, `answer`. Do NOT replace it with the view packet's shape (`packet_id`, `template_id`, `binding`, `inputs`, `draft`, …)

  这段文本在**整个仓库里不存在**（grep 无匹配），是 Main 自己写的。说明该约束只存在于编排者脑中，没有进入框架生成的 brief。

**落点.** `internal/schema/assets/stage-form-s7-template.json:6`（`instruction` 只说「只修改 answer」）；同类 `stage-form-s8-template.json:6` 同样只有「只修改 answer 并保存本卷的准确路径」。

**建议修法.** S7/S8 及其它 stage-form 的 `instruction` 统一增加一句顶层形状约束（明确列出五个顶层 key，并点名 `packet_id` 等 packet 字段不允许出现在答卷里）。这属于「把 Main 已经自发补上的约束下沉到框架」。

**关联反馈.** `source-ledger` F-019。

---

### BUG-SIM-05 · P1 · 「恢复冻结源」会静默删除 pinned commit 中不存在的文件

**现象.** S7 baseline 漂移的恢复动作是「restore the original pinned source」。当工作树里某个文件在 pinned commit 中不存在时，这个动作**删除**该文件，且不留痕。

**证据.** `ui-ledger` 的 `simulation/feedback-notes.md` 在恢复后被整个删除（`ls` 报 no such file），而不是还原成旧内容。数据靠外部快照才保住。

**次生影响.** 该文件被 S7 的 changed surface 派生选中（见 BUG-SIM-14），删除动作直接改变了冻结对象。

**证据（三次独立复现）.**

1. `ui-ledger`：该文件被整个删除（`ls` 报 no such file），而不是还原成旧内容；数据靠外部快照才保住。
2. `callback-api`：同一文件在 14:42 被人工补回后，**10 分钟内再次被删除**（14:46 复核时已不存在）。该题的 `product_baseline.paths` 把它的状态记为 `deleted`。
3. `source-ledger`：同一现象在 S7 第一轮表现为「计划变 stale 并连带拒收」，其恢复动作把文件字节还原成 pin 时的内容，从而**吃掉 pin 之后写入的 4 条记录**（F-014…F-017）。

**落点.** S7 baseline 恢复路径中执行 pinned source 还原的实现。

**建议修法.** 恢复前先比较 pinned commit 与当前工作树的文件集合：pinned 中不存在但工作树存在的文件应**保留并报告**，而不是删除；或者恢复动作改为「只还原 pinned 中存在且已漂移的路径」。

**关联反馈.** `ui-ledger` / `callback-api` 记录被框架动作删除（数据由 `/home/sebs/gsb-simulation/feedback-snapshots/` 保住）；`source-ledger` F-015 / F-018。

---

### BUG-SIM-06 · P2 · 「worktree is dirty」首行指向错误对象

**现象.** S6 集成被拦时，CLI 首行说 `worktree is dirty`。字面上指被合并的 Builder worktree，实际脏的是 **Main 自己的 integration checkout**（Main 在集成前写的反馈笔记还没提交）。只读 CLI 首行会查错方向。

**证据.** `source-ledger` F-014。CLI 首行与 checkpoint 的 `failure_reason` 指两个不同对象；提示里给出的 `git -C <Main> status --short` 才是判据。

**建议修法.** 首行主语写明，例如 `Main integration checkout is dirty`，并把两条路径都点出来。

**关联反馈.** `source-ledger` F-014。

---

### BUG-SIM-07 · P2 · `revive` 与 `revise` 的分工只存在于一行 CLI 错误里

**现象.**

- `runtime review-plan revive` 报 `baseline still drifted; the round stays stale (use revise for real drift)`——`revive` **不重钉基线**，只在 checkout 与 pinned 完全一致时恢复计划状态；
- `runtime review-plan revise --file …` 报 `prepare ReviewPlan assets: invalid character '#' looking for beginning of value`——`--file` 要的是 **JSON 内联内容**而不是路径，用法不明，也没有示例。

手册里没有这两条分工，恢复者只能靠踩一次错误信息才明白。

**证据.** `source-ledger` F-018。实测有效路径是 `git checkout HEAD -- <冻结路径>` 把字节还原成 pinned，再 `revive`。

**建议修法.** 在 S7 恢复相关文档（`docs/control/agent-protocol.md#s7` 或 blueprint 对应节）补一段：两者的适用条件差异、`revise` 需要的三个参数含义，以及「`--file` 传 JSON 内容」这一用法；`revise --help` 给出可复制的示例。

**关联反馈.** `source-ledger` F-018。

---

### BUG-SIM-08 · P2 · 「pending system_followup」把「接收失败」与「出口待续」写在一起

**现象.** S5 第二份答卷接收后出现 `pending system_followup`，措辞同时指向「答卷被拒」和「阶段出口待续」。实际是：判断**已被接收**，不得重提答卷，followup 指向阶段出口层面的恢复，而不是答卷层面的重交。

**证据.** `source-ledger` F-009（`loop-events.jsonl` sequence 16 `transition_committed TR-003`，`gate_id=GATE-DOCUMENT-PASS`，两个 guard 均通过）。

**建议修法.** 提示文本分开措辞，并明确「do not ask the reviewer to submit again」的适用场景。

**关联反馈.** `source-ledger` F-009。

---

### BUG-SIM-09 · P2 · S9 division 表两表共用 assignment_id，验证者 ID 由系统派生

**现象.** S9 的 Builder 表与独立验证者表都指向同一个 `assignment_id`，而验证者表要求先有 Builder 选择；验证者保存后系统自己派生新的 assignment 与新的 TASK。表单没有一句话说明「两表同 id 正常、验证者 id 由系统派生」，第一次看极易填错。

**证据.** `source-ledger` F-022（系统实际派生出 `repair-assignment-verification-repair-unit-1` 与新的 `TASK-96390047`）。

**建议修法.** 表单增加一句：验证者表沿用 Builder 的 `assignment_id`，不要手写验证者 ID；系统保存时会派生。

**关联反馈.** `source-ledger` F-022。

---

### BUG-SIM-10 · P2 · S9 ChangeImpact 问卷「未准备」没有恢复入口

**现象.** Hook 反复报 `S9 ChangeImpact questionnaire is not prepared at …/change-impact.json; Main/system must prepare the original request when this responsibility starts`，而 `runtime repair status` 说「Main fills and saves the exact ChangeImpact questionnaire supplied by the listener」——但这次没有供问卷。Main 只能从 `docs/examples/s7-s9/README.md` 反推出请求是一个**平 DTO**（无 answer/title/questions 外壳），并自行拼出 `required_reverification_ids`。

**证据.** `source-ledger` F-024。

**次生问题.** `required_reverification_ids` 正常路径由系统生成，问卷未准备时却要 Main 从**合同**取同一组断言 ID，表单没有说明两种来源等价，容易被读成「禁止作者填」。

**建议修法.** 给「问卷未准备」一个明确恢复入口（命令或说明），并在表单里写明 `required_reverification_ids` 的两个合法来源及其等价性。

**关联反馈.** `source-ledger` F-024。

---

### BUG-SIM-11 · P2 · 消费者超时与内容非法未区分

**现象.** `Consumer rejection: context deadline exceeded` 出现在**消费者侧**，且重试可能已经成功。同一份答卷在下一轮原生返回时被正常接收，同一份 plan 在随后保存中也注册成功——即超时不是内容非法。协议只说「不要重写」，却没说重试的具体入口，恢复者容易误以为要重写专业答卷。

**证据.** `source-ledger` F-026。同时同一超时引发三条提示：消费者拒收、`runtime review-plan: context deadline exceeded; Main must inspect original Plan/pending before cleanup`、`S7 Main questionnaire preparation deferred …`，三者其实是同一件事的连锁投影。

**建议修法.** 超时诊断与内容校验分两类措辞，并给出「原生 resume 同一 Worker 原样重交」的具体入口；同时把消费者超时与其派生出的下游提示合并成一条。

**关联反馈.** `source-ledger` F-026。

---

### BUG-SIM-12 · P2 · S10/S11 的字段口径不清

**现象.** 三处：

1. S10 `coverage_inventory[*].expected` 是**系统预填的承诺原文**（REQ / CONTRACTS / BE / TASK 全量正文），字段名没有说明这一点，Main 很容易误以为要自己写「应该怎样」；
2. S11 申请文件里同时存在 `approved_by` / `decision` / `decision_id` / `disposition` / `reason` / `runtime_id` / `scope` / `target_cursor`，提示只让填 `disposition` 和 `reason`，其余空着却不报错——多出来的空字段没有解释，容易以为漏填；
3. S10 的「八个审计主题」只出现在 Hook 文本里，Manifest 本身没有对应字段，「不适用」只能写进 `oracle` 文本，没有结构化落点。

**证据.** `source-ledger` F-028。

**建议修法.** ① 在字段 description 里标注 `expected` 为机器预填基线；② 明确 S11 只填 `disposition`（可选 `reason`），其余字段由系统写入；③ 给审计主题一个结构化落点，或在 Manifest 注释里说明主题与字段的映射。

**关联反馈.** `source-ledger` F-028。

---

### BUG-SIM-13 · P2 · 私有 Plan/Result 与「已接收」之间没有可见反馈

**现象.** S6/S9 的 brief 要求 Worker 返回私有 Result 的绝对路径，容易让人以为「交路径 = 交卷」。实际上私有文件只是 Worker 的专业产物，还要经过原消费者接收，二者之间没有自动的可见反馈。实测 S9 Builder 被会话中断重启后，worktree 里连私有文件都还没写，而 `runtime repair status` 仍显示 `missing_reports`。

**证据.** `source-ledger` F-023。

**建议修法.** 在 brief 或阶段上下文里写明：返回路径**不等于**完成，接收状态以原消费者的回执为准；给出查询接收状态的命令。

**关联反馈.** `source-ledger` F-023。

---

### BUG-SIM-14 · P1 · 未归属路径的「存在与否」会直接否证 S7 ReviewPlan 注册

**现象.** 框架**有**「归属」概念：`buildsurface.Snapshot` 区分「归属某个 TASK 的路径」与 `unattributed_paths`，并且**有意**把未归属路径留在完整审查覆盖内。代码写得很直白：

> `internal/review/draft.go:89-91` — `notes = append(notes, "Main product change has no registered S6 contribution; it remains in full review coverage: "+file)`

所以「未归属路径进入审查覆盖」是**设计意图**，不是缺陷。真正的缺陷在下一层：**这些路径一旦在基线钉住之后出现或消失，就会让 ReviewPlan 注册失败**，而且框架没有提供任何「这条路径不在本轮范围内」的声明方式。

**证据（`ui-ledger`，S7 第一轮；agent 独立定位并做了对照实验）.**

原接收者的实际回执：

> `runtime review-plan: canonical product baseline changed before ReviewPlan commit: frozen deleted subject simulation/feedback-notes.md reappeared on disk; Main must inspect original Plan/pending before cleanup: context deadline exceeded`

agent 给出的机制判定与实测：

> 基线的 `deleted` 状态来自 `start_commit→final_commit` 的 Git diff；而漂移扫描判定的是**磁盘上是否存在**，不是 Git 树里是否存在。`.gitignore` 只能把它移出 Git 树，不能把它移出磁盘。
>
> 实测确认：文件在磁盘上 ⇒ S7 ReviewPlan 注册被拒（"reappeared on disk"）；把文件移出工作树 ⇒ 同一次保存立即通过（`review-plan: registered review-plan-draft-r1 for round 1`）。

**即：`.gitignore` 解决不了这个问题。** 只要该路径在 pinned commit 中被记为 deleted，它就必须在磁盘上也不存在，否则注册必被拒；而任务要求该文件必须存在于那个路径上。两条要求直接互斥。

**证据（`callback-api`，S7 第一轮）.**

1. 产品侧无变更（TASK 交付的是空提交夹具，fixture 字节未改），于是差集里只剩两条非产品路径：`.gitignore`（modified）与 `simulation/feedback-notes.md`（deleted）。两者的 `unattributed_paths` 与 `paths` 完全一致，`tasks` 里的 TASK 路径**不在 `paths` 中**。
2. Main 把自己的反馈笔记写进产品树后，注册报：

   > `runtime review-plan: canonical product baseline changed before ReviewPlan commit: The registration succeeded when 'simulation/feedback-notes.md' is absent from the project tree. Plan 'review-plan-draft-r1' is registered, round 1 run …`

3. Main 为让基线对上，**删掉了自己的反馈文件**；而框架的「恢复冻结源」动作又会**把该文件删掉**（见 BUG-SIM-05）。两者叠加后，「写笔记 → 注册失败 → 删笔记 → 注册成功 → 再写笔记」形成循环，每一轮都会吃掉一份反馈记录。

**证据（`source-ledger`，S7 第二轮）.**

`product_baseline.paths` = `[('simulation/feedback-notes.md','modified')]`，`unattributed_paths` = 同一个文件，`tasks` = `['docs/dev/tasks/TASK-001-01.md']`（不在 `paths` 内）。**整个验证轮次的冻结 subject 只有一个未归属路径**，`change_impact.source_refs` 也只有它。

**落点.**

- `internal/buildsurface/types.go:51-58`（`Snapshot.Paths` / `UnattributedPaths`）
- `internal/buildsurface/read.go:194-198`（未归属路径判定）
- `internal/review/draft.go:86-91`（全部 `Paths` 进入 `ChangedPaths`；未归属路径仅产生一条 note）
- `internal/review/draft.go:284`（`return plan, notes`）与 `internal/repair/s7_seed.go:35-38`（notes 只在 `plan == nil` 时被拼进错误串，**成功路径上被丢弃**）

**建议修法.**

1. 给 changed surface 一个**可声明的排除面**（例如按路径前缀声明本轮审查范围），让「不属于本轮交付」的路径可以被排除，而不是无条件进入完整覆盖；
2. 至少把 `notes` 里的未归属提示**送到 Planner 手上**（现在成功路径上被丢弃），让 Planner 知道本轮审查对象里混入了未归属路径；
3. **漂移扫描的判据需要重新定义**：现在判的是「磁盘上是否存在」，而基线状态来自 Git diff，两者不同源。应当只比对**属于本轮审查范围**的路径，其余路径的增删不应判为 baseline changed——否则任何被 gitignore 的辅助文件都会永久阻塞注册（`.gitignore` 移出 Git 树，但移不出磁盘）；
4. 「基线钉住后文件出现/消失导致注册失败」的错误信息已经点名了文件（这点做得好），但应进一步区分「该路径未归属任何 TASK」与「该路径属于 TASK 却被改动」，前者应给出可执行的处置建议。

**关联反馈.** `ui-ledger` 记录 8（含对照实验）；`source-ledger` F-015 / F-018；`callback-api` S7 第一轮注册失败记录。

---

### BUG-SIM-15 · P3 · TASK 模板两处缺一句话可省一次试错

**现象.**

1. `docs/dev/tasks/TASK-template.md:23`（§2 Document Manifest）路径基准不明。模板示例写 `[本地责任](../contracts/{contract-id}.md)`，看起来像相对仓库根；实际**相对 TASK 文件自身所在目录**解析。`callback-api` 试错一次才确定（`../architecture/...` 必须写 `../../architecture/...`）。
2. 同模板 `docs/dev/tasks/TASK-template.md:146`（§8 Dependencies）**只接受 `TASK-*` 形式**的引用（见第 152 行说明）；无依赖时应保留空表。`callback-api` 写了占位文字 `（无）`，被机器报为「is not machine-tracked」。

**证据.** `callback-api` S4 两条机制疑问。

**建议修法.** §2 加一句「路径相对本 TASK 文件所在目录解析」；§8 加一句「无依赖时保留空表，不要写占位文字」。

**关联反馈.** `callback-api` S4 × 2。

---

### BUG-SIM-16 · P3 · S10/S11 之外的流程提示仍有歧义（归集）

**现象.** 以下均为措辞级歧义，agent 均能自行纠正，但每次都要额外读一轮文档：

| 位置 | 歧义 | 关联反馈 |
| --- | --- | --- |
| S5 | 「两个独立审查责任要不要手工登记 manifest / issue」——协议已说明不必，但入口提示未点明 | `source-ledger` F-008 |
| S6 | brief 要求 `Run the actual TASK self-checks: make check`，而本轮 `make check` 是只读标记校验、应由 Main 跑 | `source-ledger` F-010 |
| S6 | Worker 的空提交算不算「产品提交」 | `source-ledger` F-011 |
| S7 | E2E 判 `not_applicable` 时 Claim 留、Assignment 去 | `source-ledger` F-016 |
| S7 | 同 lens 的多条 Claim 何时可合并成一个 Assignment | `source-ledger` F-017 |
| S8 | 竞争解释应写进 `discriminator` 文本还是另起一行 Hypothesis | `source-ledger` F-021（与 BUG-SIM-01 同源） |
| S9→S7 | 第二轮 seed 把 `assertion`/`method` 也留成 `PLANNER-REFINE`，且默认 `cold_start` 容易被误读成「必须做浏览器 E2E」 | `source-ledger` F-025 |
| S0 | 候选文档到阶段的归属没有权威映射 | `source-ledger` F-001（**夹具侧**：`simulation/candidate-inputs.json` 缺 stage 字段，非框架缺陷） |
| S0 | 「锁定」手势由谁触发、「保存」是否必须用 Write 工具 | `source-ledger` F-002（实测 `cp` 亦可触发 listener） |

---

### BUG-SIM-17 · P0 · S7 原生答卷接收只有 8 秒内部预算，超时后静默降级、逼 Main 代交

**现象.** S7 的原生 review-result 接收跑在 PostToolUse Hook 里，而 Hook 的内部预算默认只有 **8 秒**；S4 早已单独立了 **28 秒**特例，S7 没有。答卷一大，接收就判超时：

> Harness Hook did not complete within its internal budget: context deadline exceeded. Main must inspect the original receipt, current responsibility and remaining handoff before retrying. A missing response does not prove no commit; preserve professional answers and do not repeat accepted work.

**超时后没有任何被指引认可的恢复入口**，而且**退出码是 0**：工具调用在 Claude 侧显示为成功，但 Hook 里的编码与提交已被中止，框架状态只写了一半（finding 已编码、observation batch 已暂存、commit 未完成）。Main 只能自己清理并重放：

> The QA result was encoded (finding + observation batch staged) but the commit timed out. Let me verify they're unindexed, then replay the reception.

**这使「原消费者接收」的责任事实上转移给 Main——而框架是刻意不把这条路交给 Main 的**，它自己就有两条测试锁死这一点：

- `internal/cli/s7_status_test.go:156` — S7 packet status 里**不得**出现 `review-result submit`（"packet status requested manual delivery"）；
- `internal/cli/controller_test.go:73` — S7 的 guidance **不得**包含 `review-result submit`。

也就是说：框架唯一认可的路径是 Hook 接收；Hook 超时后，Main 只能靠翻 CLI（甚至翻二进制）去找一条**被刻意排除在指引外**的代交通道。

**证据（三题全部触发，答卷越大越频繁；次数以 hook 观测文件为准，会话日志里的计数含 agent 复述，偏大）.**

| 题材 | S7 答卷规模 | 内部预算超限次数 | 最慢 PostToolUse | Hook `elapsed_ms` 峰值（成功路径） |
| --- | --- | --- | --- | --- |
| `callback-api` | 107 行、逐 Claim 观察 | **30** | 8242 ms（超限） | 7680 ms |
| `ui-ledger` | 中等 | 3 | 8186 ms（超限） | 10708 ms |
| `source-ledger` | 较大 | 3 | — | 6780 ms |

`callback-api` 的成功路径峰值 **7680 ms 距 8 秒预算只剩 4%**，答卷稍大即越界。

**恢复代价（`callback-api` 逐条统计）.**

- **7 次**删除已暂存的 `review-result-*.json` / finding / observation-batch 产物；
- **6 次**手工调用 `runtime review-result submit`；
- **2 次**重新注册 ReviewPlan。

一个 S7 轮次因此反复重做，题目连续三个 45 分钟窗口仍停在 `verification/running rev=7`。

**S9 变体：同一缺陷在 S9 是死路（`ui-ledger` 独立复现）.**

`ui-ledger` 在 S9 撞上同一堵墙，并且**无法自行恢复**，agent 如实记录了终点：

> Original PlanReport preparation for Builder builder-s9-… remains pending: context deadline exceeded. Its successful dispatch was retained; save this same division answer after resolving the named original source.

> S9 的发卷则多次重试仍未落盘 …… Runtime 里虽有 `plan_report_packets` 记录（status=issued），但 `view_path` 与 `answer_path` 两个文件都不存在，于是导入器找不到 brief，也就 "no current registered Plan responsibility"。

核验结论：

1. **答卷形状不是问题。** 接收器两种形状都接受——`internal/cli/native_s9.go:445-452` 允许 `{"answer":{…}}` 信封式，也允许顶层专业字段的扁平式。S9 brief 也明写「The private outputs below use top-level professional fields, without an answer envelope」。agent 写的扁平三字段形状是对的。
2. **真正的问题是"写在哪里"。** `internal/cli/native_s9.go:425-432` 的接收器从**原生 worktree** 读，并明确拒绝 Main 根目录的副本：
   > actual native worktree PlanReport is missing at %s; **Main's draft cannot substitute for the Worker artifact**
3. **依赖链是闭环的。** work location 由 Main 在 PlanReport 被记录**之后**才准备（`internal/cli/repair_command.go:599`、`:1336`），而接收器又需要 worktree 才能读到那份产物；发卷准备一旦被 deadline 打断，登记与文件同时丢失，就没有可恢复的入口。
4. **终点。** `runtime repair status` = `status: planning`、`missing_reports` 与 `missing_results` 同时非空、`report_ref` 为空；回执 `accepted: false`、`error: "PlanReport has not been accepted"`、**`field_errors: []`**——没有任何字段级诊断。S9.2 的 Result、Main 原集成、ChangeImpact、Targeted 复验、fresh S7 轮次、S10、S11 全部未发生。

> 值得肯定的是：S9 brief 本身写得清楚（相对 worktree 保存、Main 根目录只读、保存不等于接收、只有原消费者回执才算数），且 agent 没有放宽 Gate、没有伪造接收状态、没有改动产品字节——它只是被一个时间预算挡在了门外。

agent 自述的处理路径是：

> The hook ran out of internal budget; the answer is intact and staged. Let me complete the pending reception with the documented recovery entry.
> The hook's internal budget is the blocker. Let me try running the original hook adapter directly with a generous timeout.

**落点.**

- `internal/cli/hook_budget.go:20`（`hookProcessBudget = 8 * time.Second`）
- `internal/cli/hook_budget.go:24`（`s4HookProcessBudget = 28 * time.Second`）
- `internal/cli/hook_budget.go:26-41`（`hookBudget()` 只对 `.claude/submissions/…/S4/` 路径放宽，其余一律 8 秒）
- `internal/cli/hook_budget.go:103`（`context.WithTimeout(ctx, hookProcessBudget)`，第二道 8 秒闸门）与 `:125-139`（超时文案与 `PreToolUse` 的降级返回；PostToolUse 同样以退出码 0 收场）
- `internal/cli/entry_questionnaire.go:263-266`（`mainQuestionnaireViewer` 决定哪些 PostToolUse 会走这条预算路径）
- `internal/cli/s7_status_test.go:156` 与 `internal/cli/controller_test.go:73`（框架刻意不把 `review-result submit` 放进 Main 的指引）
- 平台侧 `settings.json` 的 PostToolUse `timeout=30`（秒）——**抬高平台超时没有用**，内部 8 秒闸门先到期，与 `docs/framework-upgrade.md:23-24` 的说明一致
- `docs/framework-upgrade.md:18-26`（设计意图：Hook 只做快速控制面检查，重活应在 Hook 之外跑；S7 显然没有做到）

**建议修法.**

1. 先对齐设计意图：把 S7 的原生答卷接收移出 Hook（按 `docs/framework-upgrade.md` 的 SubagentStop → `runtime …` 外部调用路径），而不是继续加预算；
2. 若必须留在 Hook 内，则按 S4 的先例给 S7 独立预算，并把判据从「路径含 `/S4/`」改成按阶段判定；
3. **超时必须让调用方看见失败**：现在退出码是 0，Claude 侧以为工具成功，实际状态只写了一半。应返回非零退出码或结构化的「未完成」标记，否则 Main 无从判断该不该重放；
4. 超时文案要给出**具体恢复命令**（当前只有「inspect the original receipt」这类无法直接执行的指示），并明确「不要由 Main 代交答卷」，否则超时会把接收责任悄悄转移给 Main。

**关联反馈.** `callback-api` S7 第一轮（`hook-*.json` 中 30 次 `internal budget`；最慢 8242 ms；7 次删产物 + 6 次手工 submit）；`ui-ledger`、`source-ledger` 各 3 次。

---

### BUG-SIM-18 · P1 · PreToolUse 预算超限时 fail-open，绕过框架自己的 deny 策略

**现象.** 与 BUG-SIM-17 同源（8 秒内部预算），但后果性质不同：**PreToolUse 是框架的硬 deny 执行路径**，而预算超限分支在到达策略判断之前就放行了。

框架自己对「策略不可用」的判定是 **fail-closed**——`internal/policy/unavailable.go:10-32`：

> With no usable policy we cannot trust shell classification or an MCP tool's name to prove that it is read-only. Preserve only explicit inspection tools.

```go
if input.Event != "PreToolUse" { return decision }   // 非 PreToolUse → warn
switch input.ToolName {
case "Read", "Grep", "Glob", "LS", "WebFetch", "WebSearch", "ToolSearch":
    return decision                                   // 只读工具 → warn
default:
    decision.Decision = "deny"                         // 其余 → deny
    decision.MatchedRuleIDs = []string{RulePolicyUnavailable}
    return decision
}
```

但预算超限的分支先短路了：

```go
// internal/cli/hook_budget.go:130-141
fail := func(err error) int {
    ...
    if input.Event == "PreToolUse" {
        return 0            // ← 直接放行，根本没走到 UnavailableDecision
    }
    if decodeErr != nil || policy.UnavailableDecision(input, err).Decision == "deny" {
        return 2
    }
    return 0
}
```

也就是说：**PreToolUse 的门禁评估超时时，框架不但没有按自己的策略 deny，反而比「策略文件缺失」还宽松——任何工具都放行。**

PreToolUse 侧承载的正是这些硬规则：`internal/policy/engine.go:1447`（protected-commands 表驱动硬 deny）、`:1013`（`phaseProductWriteFrozen` 冻结期写入禁止）、`:625` / `:666`（bug_resolution 阶段的写入约束）、`internal/policy/newnative_owned_check.go:27`（原生 Worker 自检归属）。

**证据（本轮实际触发分布）.**

| 题材 | PreToolUse 超限 | PostToolUse 超限 | SubagentStop 超限 |
| --- | --- | --- | --- |
| `callback-api` | **22** | 8 | 0 |
| `ui-ledger` | 2 | 1 | 0 |
| `source-ledger` | 1 | 1 | 1 |

**实际影响：本轮为潜伏，未造成越权。** 全量扫描三题所有会话中的工具调用，**受保护命令族命中 0 次**（`git push` 到受保护分支、`git merge --squash`、`git tag`、`gh release create/edit/delete/upload`、`gh pr merge`、`npm publish`、`goreleaser release`，见 `docs/control/protected-commands.json`）。agent 从未尝试过这些命令，所以放行没有被利用。但这不能证明路径安全——它只是说明本轮没有可利用的机会。

**建议修法.**

1. PreToolUse 的预算超限必须走 `policy.UnavailableDecision`：非只读工具一律 `deny`，与「策略文件缺失」保持同一口径；
2. 预算闸门应按事件类型分别定值：PreToolUse 的评估通常很轻（读策略表 + 运行时事实），8 秒不该被触发；真正需要长预算的是 PostToolUse/SubagentStop 里的重活（见 BUG-SIM-17），不该让它们和 PreToolUse 共用一个常量；
3. fail-open 必须在 stderr 与审计事件里留下**结构化标记**（哪条规则因超时未被评估），否则事后无法区分「评估通过」与「根本没评估」。

**关联反馈.** 无独立反馈条目（agent 只看到 stderr 文案，看不出自己被放行了）；本条由本次分析从 hook 观测文件与代码路径推出。

---

## 三、仿真夹具侧问题（非框架缺陷，记录以免复现时误判）

| 编号 | 问题 | 影响 |
| --- | --- | --- |
| FIX-01 | 反馈文件放在产品树内并被提交 → 进入 changed surface → 成为冻结 subject → 恢复时被还原/删除；其存在与否还会直接否证 S7 ReviewPlan 注册（见 BUG-SIM-14） | `source-ledger` 丢失 F-014…F-017；`ui-ledger` 丢失 1 条；`callback-api` 反复触发 `canonical product baseline changed before ReviewPlan commit` |
| FIX-02 | `simulation/candidate-inputs.json` 只有 sha256，没有 stage 归属字段 | S0 无法确定候选文件归属哪一阶段 |
| FIX-03 | 仿真 runner 的 45 分钟硬上限切断了单次会话的连续性，需要反复 `--resume` | `source-ledger` 用了 3 个窗口才到终点 |
| FIX-04 | 三个题材同时跑时磁盘一度打满，导致编译失败 | 需预留 ≥ 38G |
| FIX-05 | 反馈文件放在产品树内并被提交，会让 S7 的 `product_baseline.paths` 只剩脚手架路径（`unattributed_paths`），真正交付的 TASK 路径不在 `paths` 内 | `callback-api` 与 `ui-ledger` 的 S7 第一轮审查对象实际是测试脚手架，不是产品 |

---

## 四、本清单之外仍处未修状态的存量项

以下问题在本次仿真之前已确认，**本轮未修**（仿真范围之外），一并列出以免遗漏：

1. `docs/rules/ui-prototype.md:83` 与 `skills/ui-prototyping/SKILL.md:192` 仍称页眉四 token 被机器校验，但 `hasProtoMetaHeader` 已删除且无替代实现。
2. `blueprint/l4-mechanisms/L4-hook-platform-wiring.md:33` 仍写旧的 PostToolUse 工具清单，而 `settings.json` 已改为 `*`。
3. `review-plan.schema.json:273` 的 claims 项含 `behavior_surface`，但 `reviewmeta.Claim` 无该字段。
4. `submission.Receipt.attempt_status` 未列入 `submission-receipt.schema.json`（该 schema 未被生产代码引用）。
5. `internal/cli/action_catalog.go` 未随新命令更新。
6. `internal/evidence/local_path.go` 与 S9 捕获没有敏感文件黑名单。
7. 四个存量失败测试：`internal/cli` 的 `TestS6NextAssignsAuthoringToRegisteredBuildersAndIntegrationToMain`、`TestS6TaskDoneWhenReadsNumberedClosingContractFromTaskTemplate`；`internal/integration` 的 `TestS7RegisteredAdditionalMethodReachesNativeViewAndFrozenRecovery`；`internal/investigation` 的 `TestS9RepairPlanReportOriginalHookCoverageReplayAndWriteBarrier`。

---

## 五、核验结论与修复方案（2026-10-24 增补）

> 基调：减负、不让 Agent 猜测机制、用问卷替代传参、不让 Agent 手抄 hash/sha。
> 核验基线：`HEAD 823743c` + 工作区 58 文件未提交改动（约 +596/-103）。下表「已在工作区修复」指代码 + schema + 单测三件套已在工作区出现、待提交与全量回归；「真 Gap」指仍会阻塞、产错误产物或逼 Agent 走被框架刻意排除的路，必须动代码，不能只改措辞。

### 1. 定级总表

| 编号 | 结论 | 依据 |
| --- | --- | --- |
| SIM-01 `no_competing_hypothesis` 语义相反 | ✅ 已修复并提交（`866f4be`） | 新增 `competing_hypothesis_assessment` 三选一收口 + 问卷/flag/SKILL/协议同口径指引；`competing_assessment_test.go` + S8 route 回放测试通过 |
| SIM-02 S8 route 问卷「reuse fixed reference」 | 已在工作区修复 | `internal/cli/investigation_lead_disposition.go` 问卷指令已改：跨阶段用登记 ID，本阶段用消费者已捕获 `path:`，保留「不自算 digest」 |
| SIM-03 `planned_paths` 无语义 | 已在工作区修复 | `internal/cli/native_s6_brief.go:nativeS6CurrentAuthoringGuidance` 明确写域语义，三 brief 共用同一常量 |
| SIM-04 S7/S8 答卷顶层形状 | 已在工作区修复 | `stage-form-s7/s8-template.json:instruction` 已列五个顶层 key 并点名禁止 `packet_id` 等 packet 字段 |
| SIM-05 恢复冻结源静默删文件 | 真 Gap，工作区仅文档缓解 | `internal/cli/action_catalog.go` 只加了「先保现改、禁 blanket reset」文案；恢复路径仍无「pinned 无→现存则保留并报告 / 只还原 pinned 存在且漂移路径」的代码护栏，三次复现未被代码级消除 |
| SIM-06 `worktree is dirty` 主语错 | 已在工作区修复 | `internal/integration/integrate.go:mainWorktreeDirtyFailure/mainWorktreeCheckFailure` 已分 `Main integration checkout is dirty` vs `Builder checkout preserved` |
| SIM-07 `revive/revise` 分工不明 | 已在工作区修复（指引层），缺一小步 | `internal/review/revive.go` 错误串 + `action_catalog.go` + `simulation_recovery_test.go` 已写明分工；缺 `revise --help` 可复制示例与 `agent-protocol.md#s7` 一段话 |
| SIM-08 `pending system_followup` 混义 | 需优化，真文案 Gap | 未见代码改动；`TR-003` 后「已接收、禁重交、followup 指出口恢复」仍挤在一句里 |
| SIM-09 S9 两表同 `assignment_id` | 已在工作区修复大半 | `repair_lead_questionnaire` 系列表单已加「沿用 Builder id、验证者 id 系统派生」；需再跑一次 S9 表单渲染确认 |
| SIM-10 ChangeImpact「未准备」无入口 | 真 Gap（恢复入口缺失） | 无「问卷未准备」恢复命令；`required_reverification_ids` 双来源等价性未进表单 |
| SIM-11 超时 vs 内容非法未区分 | 真 Gap，SIM-17 的子集 | `internal/cli/hook_budget.go:fail()` 仍是同一句话 + 退出码 0；未区分「超时可原样 resume」vs「内容非法须重写」 |
| SIM-12 S10/S11 口径 | 已在工作区修复大半 | `s10-audit-manifest.schema.json` + `audit/human_release questionnaire` 已标 `expected=机器预填`、S11 只填 `disposition/reason`、审计主题结构化落点；剩 Manifest 注释与主题映射表 |
| SIM-13 私有文件≠已接收无反馈 | 已在工作区修复 | `internal/cli/native_s9.go` + S6 brief 已写「回路径≠接收，以原消费者回执为准 + `repair status` 查 `missing_reports/results`」 |
| SIM-14 未归属路径否证注册 | 真 Gap（最大语义债之一） | `internal/review/draft.go:86-91,284` + `internal/repair/s7_seed.go:35-38` 未动：成功路径 `notes` 仍丢弃，未归属路径无声明排除面，`callback-api/source-ledger` 循环仍可复现 |
| SIM-15 TASK 模板两句话 | 已在工作区修复 | `docs/dev/tasks/TASK-template.md:§2/§8` 已加相对基准与空表说明 |
| SIM-16 措辞归集 9 项 | 部分修复，部分需优化 | S6 空提交/`make check` 归属、`s6_prepare_test` 期望已更新；S7 合并规则、S9→S7 `cold_start` 误读、S8 discriminator 出路仍需进 brief/协议正文 |
| SIM-17 S7 Hook 8s 预算 | 真 Gap，P0，工作区未动 | `internal/cli/hook_budget.go:20,26-41,103` 原样：8s 通用、仅 S4 放宽 28s、超时退出码 0、半写状态、无认可恢复入口；三题全触发，`callback-api` 30 次 |
| SIM-18 PreToolUse 预算超限 fail-open | 真 Gap，P1，与 SIM-17 同源但性质不同 | `internal/cli/hook_budget.go:fail()` 在策略判断前对 PreToolUse 直接 `return 0`，绕过 `policy.UnavailableDecision` 的 fail-closed（非只读工具应 deny）；本轮 25 次 PreToolUse 超限，未命中受保护命令，属潜伏未利用 |
| 存量 1–4 | 已在工作区修复 | `docs/rules/ui-prototype.md` + `skills/ui-prototyping/SKILL.md` 去四 token 校验说辞；`L4-hook-platform-wiring.md:33` 改 `*`；`review-plan.schema.json` 删 `behavior_surface`；`submission-receipt.schema.json` 补 `attempt_status=rejected` |
| 存量 5–6 | 已在工作区修复 | `action_catalog.go` 已补新命令；`internal/evidence/local_path.go:+30 ValidateLocalCaptureSource` 黑名单已加 |
| 存量 7（4 个失败测试） | 已在工作区修复，待跑全量 | `s6_prepare_test/integrate/investigation/repair_plan_report` 均有对应更新 |
| FIX-01～05 | 夹具侧，非框架缺陷 | 反馈文件进产品树、`candidate-inputs.json` 缺 stage、45min 切片、磁盘打满；修仿真夹具，不进框架修复分支 |

### 2. 修复总原则（四条红线）

1. 不新增手填字段：所有恢复/重交入口只接受问卷文件路径或登记 ID，digest/sha 由系统回填，CLI 拒收手写 `sha256=`。
2. 问卷替代传参：`revive/revise/ChangeImpact/S9 division/超时重放` 全部走已发行问卷 `answer`，`--file` 只收问卷 JSON 路径，不收内联 JSON。
3. 失败可见：Hook 超时、基线漂移、未归属介入必须非零或结构化 `unfinished` 标记 + 一条可复制命令，禁止退出码 0 + 半写状态。
4. 知识绑到人：brief/`instruction`/表单写清「谁、填什么、不填什么」，文档语义下沉到 Worker 眼前，不指望 Worker 读 blueprint。

### 3. P0：先止血（SIM-17 + SIM-18 + SIM-11 + SIM-05 + SIM-14）

A. S7 接收移出 Hook（SIM-17 根治，SIM-11 顺带）：

- S7 `review-result` 走 `SubagentStop → runtime review-result submit --questionnaire <已发行 answer 路径>` 外部调用，Hook 只做轻量存在性检查，不做编码提交。
- 过渡期（若必须留 Hook）：按阶段判预算，不按路径：`planning/tasks→28s`（沿用 S4）、`verification→28s` 独立预算，删除 `/S4/` 路径嗅探；`hook_budget_test.go` 补 S7 大答卷用例。
- `fail()`：PostToolUse 超时返回退出码 2 + `{"status":"unfinished","replay":"loop-harness runtime review-result submit --questionnaire <原 answer 路径>"}`，文案分两类：超时（内容未判，原样 resume 同一 Worker 重交，禁重写）vs 内容非法（按 `field_errors` 改 `answer`，禁改顶层形状）。
- 合并连锁投影：消费者超时只发一条诊断，`review-plan deferred` 等下游提示收敛为同一 `causal_id`。

D. PreToolUse 预算超限 fail-closed（SIM-18，与 A 同文件、可同分支）：

- `fail()` 里 PreToolUse 分支改走 `policy.UnavailableDecision`：非只读工具一律 deny，与「策略文件缺失」同口径；只读工具（Read/Grep/Glob/LS/WebFetch/WebSearch/ToolSearch）维持 warn 放行。
- 预算闸门按事件类型分别定值：PreToolUse 维持 8s（评估本就轻量，超限即视为异常）；PostToolUse/SubagentStop 的重活走独立长预算（见 A），不共用常量。
- 超限放行/deny 必须在 stderr 与审计事件里留下结构化标记（哪条规则因超时未被评估），事后可区分「评估通过」与「根本没评估」。

B. 安全恢复冻结源（SIM-05 代码护栏）：

- 恢复前做集合比较：`pinned 存在 & 字节漂移 → 还原；pinned 无 & 现存 → 保留 + 报告 + 列入未归属候选；pinned 存在 & 现缺 → 按缺失处理并报告`。禁止无报告删除。
- `revive` 前置检查输出三列表（还原/保留/缺失）+ 一条问卷命令；`action_catalog.go` 文案保留，但以代码护栏为准。

C. 未归属路径排除面（SIM-14）：

1. `buildsurface` 加可声明排除面：按路径前缀声明本轮审查范围（如 `simulation/、*.log、coverage/` 默认候选），Main 在问卷勾选，不手写正则；
2. `internal/review/draft.go` 成功路径 `notes` 不再丢弃，随 seed 送 Planner，并在问卷 `read` 列表置顶未归属提示；
3. 注册失败文案区分「未归属漂移（给排除/纳入二选一命令）」vs「TASK 归属路径被改（给 rework/new-round 命令）」。

### 4. P1：补恢复入口（SIM-10 + SIM-08 + SIM-07 尾巴）

- ChangeImpact 未准备：新增 `loop-harness runtime repair prepare-change-impact --root`（问卷驱动：系统预填 `required_reverification_ids` 来自合同断言，Main 只勾选/确认），表单写明双来源等价；无问卷时 Hook 报错直接给出该命令。
- `pending system_followup`：拆两句：`answer accepted, do not resubmit` + `stage exit pending: <TR/GATE + 恢复命令>`。
- `revise --help`：补可复制示例（问卷路径版），`agent-protocol.md#s7` 补 `revive`（状态恢复，需字节一致）vs `revise`（Claim 受控变更，不动基线）vs `rework/new-round`（真产品变更）三段表。

### 5. P2/P3：措辞下沉（SIM-12 尾巴 + SIM-16 + SIM-09 确认）

- S10 Manifest 注释补八主题→字段映射；S11 表单隐藏系统字段或标「系统写入，留空」。
- S9→S7 seed：`assertion/method=PLANNER-REFINE` 改名「待 Planner 细化（非缺失）」，`cold_start` 改名「无历史基线冷启动（≠必须浏览器 E2E）」。
- S7 合并规则、何时 Claim 合一 Assignment、S8 discriminator 出路（与 SIM-01 联动：判别式处理即满足收口）进 brief 正文，不只放协议。
- S9 两表渲染回归：确认验证者表沿用 `assignment_id` + 系统派生提示可见。

### 6. 验证与收尾

1. `go test ./internal/...` 全量 + `TestS9* / TestS7* / hook_budget` 定向回归；
2. 用 `callback-api`（S7 大答卷）重放超时场景：Hook 峰值超过 28s 不再半写；用 `ui-ledger` 重放恢复删除场景：现存文件保留 + 报告；
3. `blueprint/L3-S7/S8/S9、L4-hook-platform-wiring、L4-state-transition` 同步更新；
4. 按本文件第 5 行约定，修复提交逐条删除对应条目，不累积历史；FIX-01～05 修仿真夹具（反馈文件移出产品树、`candidate-inputs.json` 补 stage）。落地顺序建议：P0（A→B→C）→ P1 → P2/P3；SIM-17 的 Hook 外移与 SIM-05/14 有耦合（都碰基线/产物状态），合在一个分支做，问卷与文案单独小分支做，避免大杂烩提交。
