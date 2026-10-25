# 仿真缺陷清单（2026-10-24 完整流转仿真）

本文件记录 2026-10-24 三题材完整流转仿真（S0～S7～S9～S7～S10～S11）暴露的、**尚未修复**的框架缺陷。

它不是设计文档，也不是调查流水账：每条只记录「现象 → 证据 → 落点 → 建议修法」，供后续修复与 blueprint 同步更新使用。修复完成后应把对应条目连同修复提交一起删除或标记为已修，不在本文件累积历史。

仿真环境位于仓库外 `/home/sebs/gsb-simulation/`，三题材共用同一份发布包：

| 题材 | 形态 | 终点 | 机械校验 | 机制疑问 |
| --- | --- | --- | --- | --- |
| `source-ledger` | NFR-only、8 份仓库现成 fixture 作对照基线 | **`release_authorized`（终态）**，S0→S7→S9→S7→S10→S11 全通 | 12/12 阶段全绿，`mechanical_checks_passed=true`，`fixture_unchanged=true` | 28 条 |
| `callback-api` | 服务行为模块、UI impact=none | **`release_authorized`（终态）**，rev=23、clean_round=2 | 21 个机检 gate 全过，`mechanical_checks_passed=true`，`fixture_unchanged=true`，`registered_evidence_sources_intact=true` | 28 条 |
| `ui-ledger` | UI impact=changed、无候选文档 | **未到终点**：`verification/running` rev=20，S7 第二轮 7 份责任里 5 份 `done`、2 份 `active` | — | 36 条 |

`source-ledger` 的 S8 再次在 `no_competing_hypothesis` 上踩到同一个坑（见 BUG-SIM-01）。`callback-api` 的 S7 接收被 Hook 内部预算反复打断（见 BUG-SIM-17），连续多个窗口停在同一 revision，最后靠"删 stale 产物 + 清空工作树 + 换新原生 agent"突破并跑到终态；`ui-ledger` 同一缺陷在 S9 变成死路、在 S7 第二轮变成**闭环**（见 BUG-SIM-17 的死锁链）。

**`ui-ledger` 的中断点是本轮唯一的"没跑完"**，如实记录：S7 第二轮 `assignment-qa-pattern-idiom-fit` 答卷已写好却接收不出，`assignment-qa-testability-oracle` 被派发但工作包从未准备；两者把这一轮钉在 `active`。主会话转录已累积到 **31 MB**，每轮 `--resume` 都在继续加长它。

**截至 2026-10-25 07:15，该题已连续 4 个窗口停在 `rev 90`**，试过的手段全部无效：删陈旧产物并重存答卷、清空遗留 worktree（`.claude/worktrees` 现为 0 文件、工作树干净、`git status` 30 ms）、换新原生 agent 重派（2 次，转录 455 KB / 570 KB）、给宿主进程降优先级、留静默窗口、逐个函数读框架实现找"不依赖时序的出路"。**没有一条跳出 `S7 reception incomplete` / `canonical frozen baseline drift scan: context deadline exceeded`。** 看护器仍在自动续跑（上限 16 个窗口），若后续突破再补记。

原始反馈原文：`/home/sebs/gsb-simulation/{topic}/project/simulation/feedback-notes.md`（每 10 秒快照到 `feedback-snapshots/`；因框架恢复动作会删除该文件，最终以 `feedback-merged/{topic}.md` 为准），汇总见 `/home/sebs/gsb-simulation/feedback-consolidated.md`。

**反馈条数以 `feedback-merged/` 为准**（`merge.py` 对 40 份快照去重合并）：`source-ledger` **28** 条、`callback-api` **28** 条、`ui-ledger` **36** 条，合计 **92** 条。每条按「原始提示原文 → 原本的疑问 → 实际怎么判断 → 最终是否正确」四段记录。

---

## 一、总体结论

### 1. 减负到位了吗

**到位了一半，而且分界线很清楚。**

到位的部分，有运行时证据：

1. **门禁不再需要猜。** Hook 观测文件里**没有一次工具调用被拦**：三题合计 **7 494 次** hook 观测（`source-ledger` 1 105 / `callback-api` 2 785 / `ui-ledger` 3 604），**退出码全部为 0**；框架的决策日志（`final-hook-decisions.jsonl`）只有 `allow` 与 `audit`（289+20 / 989+36 / 1 172+39），**`deny` 0 次**。agent 不再猜「这一步会不会被卡」，机器判定是确定的。

   > **统计口径更正（2026-10-25 06:55）.** 仿真辅助脚本 `status.py` 输出的 `hook errors: denied=1, rejected=29`（`callback-api` `denied=2, rejected=45`）**不是框架的拒绝次数**——那是对整个 hook 观测文件体做子串匹配的结果，命中的是 **agent 自己的问题文本**（例如 S7 brief 里写着 "Explicit incorrect digests are rejected"）。**以退出码与决策日志为准：策略层拦截 0 次。** 本条结论不变，但辅助脚本的统计口径需修正。
2. **流程骨架完全由 Runtime 托管。** 68 次 `transition_committed` 全部由系统提交；阶段问卷路径一律由 listener 生成，Main 从不自造路径或 envelope；Worker 答卷由原消费者在原生 Stop 接收，Main 不代填。
3. **上下文交付基本够用。** `source-ledger` 的 28 条机制疑问里 18 条**无需试错**，直接由 packet 的 `read` 列表、模板、协议段落回答。`callback-api` 的 S0/S2/S3 三条明确写出「不需要试错」「三处互相印证」「规则表达是明确的、可直接引用的」。

   > **全量重算（92 条合并反馈，关键词扫描口径）.** 以「试错/试了/失败/被拒/拒收/报错/重试/超时/猜/不确定/返工/卡住/绕/逆向/翻源码/读源码」为试错信号逐条扫描：`source-ledger` 29 条中 **13 条**含信号（44%）、`callback-api` 28 条中 **22 条**（78%）、`ui-ledger` 36 条中 **28 条**（75%），合计 63/94 ≈ **67%**。**`source-ledger` 的干净条目数（16）与人工判读的 18 条基本吻合，说明这个口径可用。** 题材之间差异明显：`source-ledger` 有 8 份仓库现成 fixture 作对照基线，问答差距最小；另两题一半以上的疑问都要靠试错或逆向才能收敛。
4. **编排者能自我补齐约束。** `source-ledger` 在 S7 答卷被拒后，**自己**把一段顶层形状约束写进了后续 S8 子代理 brief（见 BUG-SIM-04），传播有效。

**但「0 次拦截」有前提。** Hook 的**策略**层一次都没拦过人，Hook 的**时间预算**却三题全部越界：**S7 接收事件（`SubagentStop`）耗时中位数 3 265 ms → 6 377 ms → 6 445 ms，已贴着约 6 658 ms 的内部预算运行**（`ui-ledger` p90 7 265 ms、49 次直接超限；框架自记 S7 答卷提交被拒 52 次 / 接受 5 次）。也就是说，减负在「判定是否合规」这一层做到位了，在「判定跑不跑得完」这一层没有——超时后 Main 只能自己代交答卷，把「原消费者接收」的责任悄悄接了过去。

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

**三题都出现过去 `strings` 翻编译后的 harness 二进制找字段名的行为**（`source-ledger` 13 次、`callback-api` 8 次、`ui-ledger` 38 次）。这是「文档没给出字段名」最直接的证据，也是本清单里多条 P1 的共同来源。

**把三题「绕过文档找答案」的行为全量统计后，结论更硬：每一题都绕过了，而且越绕越深。**

统计口径：对 `~/.claude/projects/-home-sebs-gsb-simulation-<topic>-project/` 下的**主会话与全部子代理转录**逐条扫描 `tool_use`，按工具与目标路径分类（`Bash` 取完整命令行、`Read` 取 `file_path`、`Grep` 取 `path`+`pattern`）。判定规则：`strings` 命令且目标为 harness 二进制 → 翻二进制；命令/路径命中 `go-system-builder/internal/`、`go-system-builder/cmd/` 或「仓库路径 + `.go`」 → 读实现源码；命中 `BUGS.md` → 翻缺陷清单。

| 题材 | `strings` 翻编译后的二进制 | 读框架实现源码 | 翻缺陷清单 | 合计 |
| --- | --- | --- | --- | --- |
| `callback-api` | 8 | **234** | 15 | **257** |
| `ui-ledger` | **38** | **80** | 0 | **118** |
| `source-ledger` | 13 | 0 | 0 | **13** |

合计 **388 次**绕过文档去反推机制。

> **修正记录.** 本文件早先记的是「93 次」，那是**只数主会话早期片段的人工统计，严重低估**；上表是可复现的全量脚本统计口径（抽检 18 条命中样本，全部为真实的源码阅读或二进制翻查，仅 1 条因命令里出现 `harness` 字样而误命中）。修正后结论方向不变，但**量级差 4 倍**，说明"绕过文档"不是偶发行为，而是贯穿全程的主要求解手段。

**所有绕过行为 100% 发生在 Main 会话**（子代理 0 次）——Main 是那个既要知道"下一步是什么"、又要去猜"这个字段到底什么意思"的角色。

`callback-api` 最极端：它不满足于翻二进制，而是直接去读**框架的 Go 实现**来搞清楚"修订后 reviewer 该怎么重派"——`internal/review/revise.go`、`internal/review/result_packet.go`、`internal/cli/run.go`、`internal/assignment/register.go`、`internal/workpacket/review_result.go`、`internal/buildsurface/*.go`，并 grep `"workgroup %s is already registered"`、`"reviewer_dispatched"`、`"canonical frozen baseline drift scan"`、`"agent-begin"`。agent 自述：

> The harness source is available. Let me read the actual re-dispatch logic.

`ui-ledger` 的路径更长：先 `strings` 翻二进制找字段名（38 次），发现不够用之后**转去读框架源码**（80 次），最后一路读到 `detectUndeclaredProductDriftWithState`、`VerifyReviewResultPacketIssue`、`acquireLegacyLockContext` 这些**实现细节**，并开始逐个函数找"一条不依赖时序的出路"。它的原话：

> Let me try a different angle on the stale-artifact conflict — the consumer offers "change the artifact id" as an alternative to deleting the file. Let me check whether the plan revision is validated at consumption time.

> Let me check for a stale legacy lock sentinel — `acquireLegacyLockContext` waits 30 s before evicting one, which would consume the whole budget.

也就是说，**当文档答不上来时，agent 唯一能走通的路是读实现**。这说明"猜机制"的残余成本不是一次试错，而是把框架当黑盒逆向一遍——**真实部署里没有源码可读时，这条路直接断掉**。而这正是本轮最大的风险面：仿真里 agent 之所以最终能收敛（`callback-api` 到终态），恰恰是因为它能读到源码和这份缺陷清单。

### 3. 流程完整性

`source-ledger` 走通完整链路：`inactive → planning/design → planning/contracts → planning/tasks → document_verification → building → document_verification → building → verification/planned → verification/running → verification/cannot_clean ×2 → verification/observation_sealed → bug_resolution/{investigation,planning,fixing,targeted_reverification} → verification/running → verification/clean → acceptance → release_audit → awaiting_human_release → release_authorized`。

修复前在 S6 卡死的阻塞（Worker 按 brief 写紧凑 `assertion_evidence[].plan_assertion_index`，集成侧按顶层 key 逐字段比对）已定位并修复（提交 `823743c`），修复后两题不再复现。

**三题对照（`loop-events.jsonl` 的 `transition_committed`）.**

| 题材 | 转换总数 | 返工类转换 | 走过的机检 gate |
| --- | --- | --- | --- |
| `source-ledger` | 68 | `ReviewPlan marked stale` ×2 + `revived` ×2（全部由**夹具把反馈文件放进产品树**引起） | **21 个全过**，含 `GATE-VERIFY-CLEAN-ROUND-PASSED` / `GATE-ACCEPTANCE-COMPLETE` / `GATE-RELEASE-AUDIT-APPROVED` → `release_authorized` |
| `callback-api` | 76 | `marked stale` ×1 | **21 个全过** → `release_authorized` |
| `ui-ledger` | 81 | `marked stale` ×3 + `revived` ×3（S9 第二轮计划的漂移回环） | 18 个，停在 `S9-REPAIR-HANDOFF` 之后、`GATE-VERIFY-CLEAN-ROUND-PASSED` 之前 |

**三题的前 18 个机检 gate 逐项相同**，一字不差：

> `GATE-PLANNING-DESIGN-COMPLETE` → `GATE-PLANNING-CONTRACTS-COMPLETE` → `GATE-PLANNING-TASKS-COMPLETE` → `GATE-DOCUMENT-PASS` → `GATE-BUILDER-BATCH-READY` → `GATE-VERIFY-BLOCKING-FINDING` → `S8-INVESTIGATION-INTAKE` → `S8-INVESTIGATION-CASE-REVISION` ×3 → `S8-REPAIR-CONTRACT-APPROVAL` → `S9-SESSION-OPEN` → `S9-PLAN-COMPILE` → `S9-PLAN-REPORT` → `S9-RESULT-SUBMIT` → `S9-IMPACT-RECONCILE` → `S9-TARGETED-REVERIFICATION` → `S9-REPAIR-HANDOFF`

差异全部来自内容形状、契约语义与时间预算，没有一处是「流程本身不确定」。这是「减负在流程机制层已到位」最直接的证据——agent 不需要知道下一步是什么，Runtime 知道。

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

### BUG-SIM-02 · P1 · S8 route 问卷指令与消费者引用语义冲突（✅ 已修复，提交 `cb70c61`）

**现象.** S8 route 问卷的运行时指令原文要求「reuse the fixed reference supplied by the accepted Case result」。但被接受的 Case 结果里的 `path:` 引用是 **S7 捕获**的证据；S8 的消费者要求证据必须由**本阶段消费者**捕获。照抄被接受的引用会被拒。

**证据.** `source-ledger` F-020；被拒的实际引用为 `path:simulation/observations/first-round.txt#sha256=1277…`，来源是 S7 Finding 的固定摘要，不是 S8 捕获物。

**落点.** `internal/cli/investigation_lead_disposition.go:170`（`investigationLeadRouteNextContext` 生成的问卷指令）：

> Reference array shape: {"evidence_refs":["path:<actual-source-path>#sha256=<actual-sha256>"]}; **reuse the fixed reference supplied by the accepted Case result** or an applicable registered ID displayed with the sources.

**建议修法.** 指令改为：`path:` 引用必须指向**本阶段消费者已捕获**的证据；跨阶段继承请使用登记 ID。保留「不要自己计算 path/digest」这半句。

**关联反馈.** `source-ledger` F-020。

---

### BUG-SIM-03 · P1 · S6 brief 未定义 `planned_paths` 语义（✅ 已修复，提交 `18bbf11`）

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

### BUG-SIM-04 · P1 · S7 答卷顶层形状约束缺失（S8 已有等价先例）（✅ 已修复，提交 `438f8e7`）

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

### BUG-SIM-05 · P1 · 「恢复冻结源」会静默删除 pinned commit 中不存在的文件（✅ 已修复，提交 `3a26782`）

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

### BUG-SIM-06 · P2 · 「worktree is dirty」首行指向错误对象（✅ 已修复，提交 `619f538`）

**现象.** S6 集成被拦时，CLI 首行说 `worktree is dirty`。字面上指被合并的 Builder worktree，实际脏的是 **Main 自己的 integration checkout**（Main 在集成前写的反馈笔记还没提交）。只读 CLI 首行会查错方向。

**证据.** `source-ledger` F-014。CLI 首行与 checkpoint 的 `failure_reason` 指两个不同对象；提示里给出的 `git -C <Main> status --short` 才是判据。

**建议修法.** 首行主语写明，例如 `Main integration checkout is dirty`，并把两条路径都点出来。

**关联反馈.** `source-ledger` F-014。

---

### BUG-SIM-07 · P2 · `revive` 与 `revise` 的分工只存在于一行 CLI 错误里（✅ 已修复，提交 `e9c026f`，指引层见 `3a26782`）

**现象.**

- `runtime review-plan revive` 报 `baseline still drifted; the round stays stale (use revise for real drift)`——`revive` **不重钉基线**，只在 checkout 与 pinned 完全一致时恢复计划状态；
- `runtime review-plan revise --file …` 报 `prepare ReviewPlan assets: invalid character '#' looking for beginning of value`——`--file` 要的是 **JSON 内联内容**而不是路径，用法不明，也没有示例。

手册里没有这两条分工，恢复者只能靠踩一次错误信息才明白。

**证据.** `source-ledger` F-018。实测有效路径是 `git checkout HEAD -- <冻结路径>` 把字节还原成 pinned，再 `revive`。

**建议修法.** 在 S7 恢复相关文档（`docs/control/agent-protocol.md#s7` 或 blueprint 对应节）补一段：两者的适用条件差异、`revise` 需要的三个参数含义，以及「`--file` 传 JSON 内容」这一用法；`revise --help` 给出可复制的示例。

**关联反馈.** `source-ledger` F-018。

---

### BUG-SIM-08 · P2 · 「pending system_followup」把「接收失败」与「出口待续」写在一起（✅ 已修复，提交 `bd95960`，整理见 `blueprint/SIM-08-followup-note.md`）

**现象.** S5 第二份答卷接收后出现 `pending system_followup`，措辞同时指向「答卷被拒」和「阶段出口待续」。实际是：判断**已被接收**，不得重提答卷，followup 指向阶段出口层面的恢复，而不是答卷层面的重交。

**证据.** `source-ledger` F-009（`loop-events.jsonl` sequence 16 `transition_committed TR-003`，`gate_id=GATE-DOCUMENT-PASS`，两个 guard 均通过）。

**建议修法.** 提示文本分开措辞，并明确「do not ask the reviewer to submit again」的适用场景。

**关联反馈.** `source-ledger` F-009。

---

### BUG-SIM-09 · P2 · S9 division 表两表共用 assignment_id，验证者 ID 由系统派生（✅ 已修复，提交 `ff92af9`）

**现象.** S9 的 Builder 表与独立验证者表都指向同一个 `assignment_id`，而验证者表要求先有 Builder 选择；验证者保存后系统自己派生新的 assignment 与新的 TASK。表单没有一句话说明「两表同 id 正常、验证者 id 由系统派生」，第一次看极易填错。

**证据.** `source-ledger` F-022（系统实际派生出 `repair-assignment-verification-repair-unit-1` 与新的 `TASK-96390047`）。

**建议修法.** 表单增加一句：验证者表沿用 Builder 的 `assignment_id`，不要手写验证者 ID；系统保存时会派生。

**关联反馈.** `source-ledger` F-022。

---

### BUG-SIM-10 · P2 · S9 ChangeImpact 问卷「未准备」没有恢复入口（✅ 已修复，提交 `7f76365`）

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

**调查结论（2026-10-26，未修，留待 SIM-17 一并根治）.**
超时不是内容错，也不是消费者判死刑，是 S7 大答卷的重活（验签名、跑消费者、写状态）塞进了 Claude Code 给 Hook 的 8 秒窗口，计时器到即 `context deadline exceeded`，答卷状态半写。同一份答卷下一轮原样重交即被接收（F-026）。
近治是把 `hook_budget.go:fail()` 拆两类文案（超时：内容未判，原样 resume 同一 Worker 重交，禁重写；内容非法：按 field_errors 改 answer），不动预算；根治归 SIM-17（重活移出 Hook，走 `review-result submit` 外部调用，Hook 只做轻量检查）。退出码与连锁投影合并也在 SIM-17。

**关联反馈.** `source-ledger` F-026。

---

### BUG-SIM-12 · P2 · S10/S11 的字段口径不清（✅ 已修复，提交 `6795e11`）

**现象.** 三处：

1. S10 `coverage_inventory[*].expected` 是**系统预填的承诺原文**（REQ / CONTRACTS / BE / TASK 全量正文），字段名没有说明这一点，Main 很容易误以为要自己写「应该怎样」；
2. S11 申请文件里同时存在 `approved_by` / `decision` / `decision_id` / `disposition` / `reason` / `runtime_id` / `scope` / `target_cursor`，提示只让填 `disposition` 和 `reason`，其余空着却不报错——多出来的空字段没有解释，容易以为漏填；
3. S10 的「八个审计主题」只出现在 Hook 文本里，Manifest 本身没有对应字段，「不适用」只能写进 `oracle` 文本，没有结构化落点。

**证据.** `source-ledger` F-028。

**建议修法.** ① 在字段 description 里标注 `expected` 为机器预填基线；② 明确 S11 只填 `disposition`（可选 `reason`），其余字段由系统写入；③ 给审计主题一个结构化落点，或在 Manifest 注释里说明主题与字段的映射。

**关联反馈.** `source-ledger` F-028。

---

### BUG-SIM-13 · P2 · 私有 Plan/Result 与「已接收」之间没有可见反馈（✅ 已修复，提交 `4d3138d`）

**现象.** S6/S9 的 brief 要求 Worker 返回私有 Result 的绝对路径，容易让人以为「交路径 = 交卷」。实际上私有文件只是 Worker 的专业产物，还要经过原消费者接收，二者之间没有自动的可见反馈。实测 S9 Builder 被会话中断重启后，worktree 里连私有文件都还没写，而 `runtime repair status` 仍显示 `missing_reports`。

**证据.** `source-ledger` F-023。

**建议修法.** 在 brief 或阶段上下文里写明：返回路径**不等于**完成，接收状态以原消费者的回执为准；给出查询接收状态的命令。

**关联反馈.** `source-ledger` F-023。

---

### BUG-SIM-14 · P1 · 未归属路径的「存在与否」会直接否证 S7 ReviewPlan 注册（✅ 已修复，提交 `15c4efd`）

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

### BUG-SIM-15 · P3 · TASK 模板两处缺一句话可省一次试错（✅ 已修复，提交 `82b0fc3`）

**现象.**

1. `docs/dev/tasks/TASK-template.md:23`（§2 Document Manifest）路径基准不明。模板示例写 `[本地责任](../contracts/{contract-id}.md)`，看起来像相对仓库根；实际**相对 TASK 文件自身所在目录**解析。`callback-api` 试错一次才确定（`../architecture/...` 必须写 `../../architecture/...`）。
2. 同模板 `docs/dev/tasks/TASK-template.md:146`（§8 Dependencies）**只接受 `TASK-*` 形式**的引用（见第 152 行说明）；无依赖时应保留空表。`callback-api` 写了占位文字 `（无）`，被机器报为「is not machine-tracked」。

**证据.** `callback-api` S4 两条机制疑问。

**建议修法.** §2 加一句「路径相对本 TASK 文件所在目录解析」；§8 加一句「无依赖时保留空表，不要写占位文字」。

**关联反馈.** `callback-api` S4 × 2。

---

### BUG-SIM-16 · P3 · S10/S11 之外的流程提示仍有歧义（归集）（✅ 已修复，提交 `09be9e8`）

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

### BUG-SIM-17 · P0 · S7 答卷接收的 hook 事件承担两件重活、共用一份时限，超时后静默降级逼 Main 代交

> **本条与 BUG-SIM-20 的关系（2026-10-25 06:45 实测修正）.** BUG-SIM-20（读转录并逐行解析）曾被写成本条的"根因"，但直接计时显示它**只占 30–96 ms**，不解释 6.4 秒。本条记录的是**时限机制本身**的缺陷（预算共用、超时退出码 0、无认可恢复入口、错误文案误导），**它才是主因**；SIM-20 只是让接收成本随重试次数上升的次要放大器。两条叠加才产生本轮观察到的现象。

**现象.** S7 的原生 review-result 接收跑在 PostToolUse Hook 里，而 Hook 的内部预算只有约 **6.6 秒**（`hookProcessBudget = 8s` 减去启动与前置开销；S4 另立了 28 秒特例）。超时时的回执是：

> Harness Hook did not complete within its internal budget: context deadline exceeded. Main must inspect the original receipt, current responsibility and remaining handoff before retrying. A missing response does not prove no commit; preserve professional answers and do not repeat accepted work.

**超时后没有任何被指引认可的恢复入口，而且退出码是 0**：工具调用在 Claude 侧显示成功，但 Hook 里的编码与提交已被中止，框架状态只写了一半（finding 已编码、observation batch 已暂存、commit 未完成）。Main 只能自己清理并重放：

> The QA result was encoded (finding + observation batch staged) but the commit timed out. Let me verify they're unindexed, then replay the reception.

**这使「原消费者接收」的责任事实上转移给 Main——而框架是刻意不把这条路交给 Main 的**，它自己就有两条测试锁死这一点：

- `internal/cli/s7_status_test.go:156` — S7 packet status 里**不得**出现 `review-result submit`（"packet status requested manual delivery"）；
- `internal/cli/controller_test.go:73` — S7 的 guidance **不得**包含 `review-result submit`。

也就是说：框架唯一认可的路径是 Hook 接收；Hook 超时后，Main 只能靠翻 CLI（甚至翻二进制）去找一条**被刻意排除在指引外**的代交通道。

**真正的成本结构（`ui-ledger` 逐项排除后给出，本条据此定稿）.**

同一份 ReviewResult 连试 11 次收不进来，而同轮另外 5 份 1–3 次就过。`ui-ledger` 做了逐项排除，结论是**负载只是放大器，根因是 hook 事件的工作量结构**：

| 候选原因 | 排除依据 |
| --- | --- |
| 答卷体积 | 已压到本轮最小（10 600 B，staged 结果 5 463 B）；同项目姊妹工程被接收的 round-2 QA 结果是 **51 502 B** ⇒ 提交开销与载荷无关 |
| 答卷内容/字段 | 失败件与已成功件的 staged 结果**逐字段 diff**，除 id 与正文外完全同形；消费者本已完整解析（补上了 4 条证据的 `#sha256=`、算出 `subject_digest`、写出完整 `effective_answer`），**只是没来得及提交** |
| packet / manifest / plan | 与已成功件全量 diff：只有 id、名称、时间戳等身份字符串不同 |
| 基线真的漂移 | **不是**：跟踪树指纹、内容哈希、HEAD、frozen `changed_surface` 的 sha、frozen `deleted` 主体从 git 历史复现的 blob sha 全部一致；`git fsck` 干净；`runtime recover inspect` 报 active runtime `readable` 且与 checkpoint 的 sha 一致 |
| 机器负载 | load average 5.17→8.64（并发 `go build`）期间失败；**把负载降到 1.44/1.79 后重试仍然失败** ⇒ 负载是放大器，不是根因 |

把剩下的三条串起来，就是根因：

1. **一个 hook 事件干两件重活。** 同一次 hook 上下文里同时出现「把这一份责任带进原生 Agent」（= 接收上一份 ReviewResult）与「准备下一份问卷」，**两件各自约 6 秒的活共用一份约 6.6 秒的时限**。谁的活先被尝试，谁就吃掉预算。
2. **Runtime 状态随轮次变大。** `loop-state.json` 达 **742 336 B**，其中 `entities` **459 463 B**，逐份存着 work-packet 问卷 draft。状态越大，每次 hook 要做的序列化+校验越慢——于是**越靠后的责任越难**（已完成的 5 份每份都要先经历一次「释放队列 → 备出下一份问卷」）。
3. **基线漂移扫描遍历整个工作树**，包括**未跟踪**的遗留 S9 worktree（26 MB / 742 文件、3.7 MB）。每次扫描都要走它们。

**「两件重活」的损伤在两侧都留了痕（`ui-ledger` 第二轮权威账本）.**

`loop-state.json` 的 `review.assignments` 是权威账本，S7 第二轮 7 份责任的真实状态（2026-10-25 06:21 实读）：

| assignment | status | 答卷 | 工作包 |
| --- | --- | --- | --- |
| `assignment-dv-traceability` | `done` | 有 | 有 |
| `assignment-qa-debt-operability` | `done` | 有 | 有 |
| `assignment-qa-design-boundary` | `done` | 有 | 有 |
| `assignment-qa-logic-state-error` | `done` | 有 | 有 |
| `assignment-qa-maintainability` | `done` | 有 | 有 |
| `assignment-qa-pattern-idiom-fit` | **`active`** | 有（10 565 B，sha256 前后一致） | 有 |
| `assignment-qa-testability-oracle` | **`active`** | **无** | **无** |

两条 `active` 恰好各对应 hook 事件的两件重活：

1. `qa-pattern-idiom-fit` = **接收侧**卡死：答卷早已写好并被反复重存（sha256 前后一致、字段未改），框架消费不出来；
2. `qa-testability-oracle` = **准备侧**卡死：它的 activation 已落盘（`.claude/evidence/workgroup-qa-testability-oracle/TASK-001-01/activation-s7-g1-r2-qa-testability-oracle.json` 与 `.claude/workgroups/manifest-g1-r2-assignment-qa-testability-oracle.json`），说明 **Reviewer 已经被派发**，但**工作包从未准备出来**——它连问卷都没拿到，自然交不出答卷。

**"准备下一份问卷"与"接收上一份 ReviewResult"共用一份时限，代价是可以对得上的**：同一轮里 5 份顺利进账、1 份卡在接收、1 份根本没发卷。这条账本把"一个 hook 事件干两件重活"从推测变成了可核对的事实。

**完整死锁链（hook 自己的 stderr 逐环扣上，2026-10-25 06:18 实读）.**

> S7 reception incomplete; Main owns recovery for s7-g1-r2-qa-pattern-idiom-fit at …. Preserve the original answer, observations and any accepted receipt; this return does not establish review completion. … Do not invent a verdict, product Finding or blocked Claim to exit. **canonical product baseline changed before ReviewResult commit: canonical frozen baseline drift scan: context deadline exceeded; staged ReviewResult artifacts were retained because lock-protected cleanup could not prove they were unreferenced: .claude/evidence/loop-REQ-001/g1/reviews/s7-g1-r2-qa-pattern-idiom-fit/review-result-a62f24be….json**

把这条拆开，就是一个闭环：

1. **同一次 hook 事件的前序操作已经吃光了预算**（编码 finding + 暂存 observation batch + 准备下一份问卷 + 编码/落盘），漂移扫描只拿到剩余时间 → 超时 → 错误被包装成「canonical product baseline changed」；
2. 因为扫描没跑完，兜底清理**无法证明暂存产物未被引用**（`internal/review/submit.go:627-629`），于是把**已经生成好的** ReviewResult 留在盘上；
3. 下一次接收看到盘上已有同 id 产物 → 判为陈旧产物 → 拒绝；
4. Main 删掉产物重试 → 回到第 1 步。

**即：清理逻辑把"扫描超时"翻译成了"有东西还被引用"，于是每轮都留下一颗毒丸；而毒丸的存在让下一轮的工作量更大，于是预算更不够——这是一个自我维持的死循环。** agent 的实测动作与此完全一致——它反复做的正是"删 stale 产物 + 保持工作树干净 + 让同一 Reviewer 重存答卷"：

> Stale staged artifact removed and the answer re-saved byte-identical in the same command; before/after sha256 equal, file valid JSON, no field changed.

**这条链路的成本还会随重试次数自我放大**（BUG-SIM-20）。本轮主会话转录已膨胀到 **31 MB**，reviewer 子代理转录 **2.4 MB**；接收路径上的固定开销（读转录）随重试次数增长，漂移扫描的开销随工作树增长，两者挤在同一个 8 秒闸门里。

**agent 已经开始逆向锁机制本身.** 它在转录里自述：「the consumer offers "change the artifact id" as an alternative to deleting the file」「check for a stale legacy lock sentinel — `acquireLegacyLockContext` waits 30 s before evicting one, which would consume the whole budget」「the state is described as a rebuildable projection, and a smaller rebuilt state would fit the budget」。**这是"猜机制"退化成"逆向一切"的又一步：它开始逐个函数读框架实现去找一条不依赖时序的出路。**

**agent 已经把逃生手段升级到"改宿主"（2026-10-25 06:58 现场记录）.**

`ui-ledger` 的 Main 在第 10 个窗口里做了三件事，顺序很清楚：

1. **量化宿主抢占**：用 `top`/`vmstat` 测出"host is stealing 53 % CPU with 39 runnable tasks"，并给出自己的结论——*"the S7 reception needs ~5.5 s against the platform's fixed 6 s hook budget"*；
2. **去改框架之外的东西**：找到一个 `minimax-code` 进程（pid 2179906，父进程是一个已运行 2.2 天的 bash 会话；**不是本条仿真会话的 2678589**），确认它没有活跃会话子进程后，**用 `renice` 把它降到 nice 19**，好让 harness hook 拿到 CPU；
3. **换新原生 agent 重派**：起一个全新 `s7-r2-qa-pattern-idiom-fit-retry` 子代理（转录几乎为零），让它按多轮保存循环重交同一份答卷。

> **这条行为本身就是一个结论.** 当框架自己的时限撑不住时，agent 的出路是**走出框架去调宿主**——改进程优先级、换 agent、重存文件、删产物。这不是"文档没写清楚"，是**框架没有给出任何在预算内完成接收的办法**。同时它也再次印证了"负载是放大器不是根因"：**基线接收成本（约 5.5 s）本来就贴着 6 s 的预算**，宿主只要拿走一点 CPU 就会翻车；把负载压到 1.44 仍失败，正是因为余量本来就只有 15%。

**余量有多紧.** `reception/s7` 平均 **5 728 ms**，内部时限约 **6 658 ms**——**只有 15% 余量**。任何一点负载增长、状态增长或工作树膨胀都会越界。

> `callback-api` 独立测得同一组数字并给出了自己的结论：*"The phase metrics show `reception/s7` at ~5.4 s against a 6 s internal budget — marginal, and it scales with registry size."* 它还特意去**已安装的二进制**（`823743c`）而不是上游 HEAD 里验证漂移扫描的证明循环——因为上游 HEAD 已被并发修复改动过。这条与本清单的测量完全吻合。

**实测有效的处置（`ui-ledger` 验证过）.**

> （a）先删 stale staged 产物；（b）把仓库根下的遗留隔离 worktree 移进被忽略的 `.claude/worktrees/`，让工作树完全干净；（c）让同一 Reviewer 重存自己那份答卷。

搬迁后 `git status --porcelain` 由 `?? .worktrees/` 变为完全干净，跟踪树指纹与内容哈希不变、HEAD 不变、产品字节未动；**同一 Reviewer 重存答卷后立刻接收成功（rev 84）**。

> **但这套处置在 `ui-ledger` 上只走通了前半程（2026-10-25 07:15 修正）.** 遗留 worktree 确实已被清空（`.claude/worktrees` 现为 0 文件）、工作树保持干净、`git status` 只要 30 ms，**但接收仍然失败**——第 9–11 个窗口里 `rev` 一直停在 **90**，`assignment-qa-pattern-idiom-fit` 与 `assignment-qa-testability-oracle` 始终 `active`。换新原生 agent（转录 455 KB / 570 KB）、降宿主进程优先级、静默窗口等待，**都没有把这题推过 S7 第二轮**。
>
> **结论要修正：** 上面这套处置对 `callback-api` 有效、对 `ui-ledger` 无效。差别不在"有没有做"，而在**卡点的位置**：`callback-api` 卡在 hook 前置工作（转录扫描），省出几十毫秒就能过；`ui-ledger` 卡在**预算被前序操作吃光后的下游**（漂移扫描 → 产物被保留 → 陈旧冲突），而这套处置**完全不触及**那个环节。**这也说明"处置有效"不能推广到"问题解决"——`ui-ledger` 是本轮唯一没跑到终点的题材，而它恰好就是处置看起来最标准的那一题。**

**错误文案把「超时」包装成「基线变了」，极具误导性——而且这是绝大多数情况（2026-10-25 07:10 全量分类）.**

| 题材 | 漂移扫描**超时** | 漂移扫描**真漂移**（磁盘上重新出现） | 「未声明产品文件」真漂移 | 「stale 状态持久化」也超时 | hook 总预算超时 |
| --- | --- | --- | --- | --- | --- |
| `source-ledger` | 1 | 0 | 0 | 0 | 3 |
| `callback-api` | 34 | 1 | 0 | 1 | 72 |
| `ui-ledger` | **99** | 10 | 0 | 4 | **301** |
| **合计** | **134** | **11** | **0** | **5** | **376** |

> **134 次扫描超时 vs 11 次真漂移——92% 的「基线变了」其实根本没跑到判定。** 「undeclared product file(s) … outside frozen_subjects」这条**真正的漂移判决一次都没出现过**。

实测基线指纹、HEAD、frozen `changed_surface` 的 sha、frozen `deleted` 主体复现的 blob sha **全部一致**，基线根本没变。agent 自己的判语是：

> 错误文案把"扫描没跑完"包装成了"基线变了"，容易误判成真缺陷。

**而且这个超时是「预算被前序操作吃光」造成的，不是扫描本身慢。** 三条实测证据：

1. `internal/review/workspace.go:306-313` —— 扫描函数**继承 hook 的同一个 deadline context**，并在循环里逐段检查 `parent.Err()`；它拿到的是**剩余预算**，不是 8 秒。
2. 扫描真正要做的 IO 都很快：`git status --porcelain` **30 ms**、`git diff --name-status HEAD` **45 ms**、`loop-state.json` 解析 **24 ms**。没有任何一项接近 6 秒。
3. **最关键的一条**：最新一次回执的后半句是

   > ReviewPlan frozen subject baseline: canonical frozen baseline drift scan: context deadline exceeded; **failed to persist ReviewPlan stale status: context deadline exceeded**

   ——**扫描之后紧接着的"把 stale 状态写回去"也超时了**。如果是扫描自己慢，它把预算吃完也就算了；但"写一个状态"这种微秒级操作也超时，只能说明**预算在扫描开始之前就已经没了**。

同一条检查有两种落点：漂移能快速判定时给出可执行的原因（`frozen deleted subject … reappeared on disk`），跑不完时只吐一个 `context deadline exceeded`，却顶着「基线变了」的外壳。**这是本轮最容易把人带偏的一条报错。**

**证据（三题全部触发，次数以 hook 观测文件为准；会话日志里的计数含 agent 复述，偏大）.**

| 题材 | PreToolUse 超限 | PostToolUse 超限 | SubagentStop 超限 | S7 答卷提交 被拒 / 被接受 |
| --- | --- | --- | --- | --- |
| `callback-api` | 51 | 20 | 20 | — |
| `ui-ledger` | **258** | 6 | **49** | **52 / 5** |
| `source-ledger` | 3 | 1 | 3 | — |

> **`ui-ledger` 的 `loop-metrics.json` 里记着 `loop_s7_result_submits_total = {accepted: 5, rejected: 52}`——被拒率 91%。** 这是框架自己记的账：同一轮 52 次提交被拒、5 次被接受。（`ui-ledger` 第一轮的 7 份责任里只有 2 份进了账，所以被拒数含第一轮。）

**hook 各事件的耗时分布（`ui-ledger`，全量 3 500+ 次观测）.**

| 事件 | 次数 | 中位 | p90 | max | ≥4 s | 预算超时 |
| --- | --- | --- | --- | --- | --- | --- |
| `PreToolUse` | 1 645 | 1 876 ms | 8 169 ms | 8 604 ms | **542** | **247** |
| `PostToolUse` | 1 624 | 796 ms | 2 711 ms | 10 708 ms | 31 | 6 |
| **`SubagentStop`** | **75** | **6 445 ms** | **7 265 ms** | 8 296 ms | **60** | **49** |
| `SubagentStart` | 74 | 635 ms | 1 074 ms | 1 590 ms | 0 | 0 |
| `Stop` | 41 | 755 ms | 1 023 ms | 1 474 ms | 0 | 0 |

三条读法：

1. **`SubagentStop` 就是 S7 接收事件**，75 次里 **49 次超限**、中位 6 445 ms —— 内部预算 6 658 ms 已被吃满，p90 直接越过预算线。
2. **`PreToolUse` 的中位只有 1.9 秒，但 p90 是 8.2 秒**——长尾极重，而且 247 次直接越过内部预算（这 247 次全部以退出码 0 放行，见 BUG-SIM-18）。Main 每做一次工具调用，都可能被自己的门禁评估拖住 8 秒。
3. `PostToolUse` 中位 796 ms 很正常，**问题集中在 SubagentStop（接收）与 PreToolUse（评估）两个事件上**，与「一个 hook 事件干两件重活」的判断一致。

**恢复代价（`callback-api` 逐条统计）.**

- **7 次**删除已暂存的 `review-result-*.json` / finding / observation-batch 产物；
- **6 次**手工调用 `runtime review-result submit`；
- **2 次**重新注册 ReviewPlan。

一个 S7 轮次因此反复重做，题目连续多个 45 分钟窗口停在同一 revision。

**S9 变体：同一缺陷在 S9 是死路（`ui-ledger` 独立复现）.**

`ui-ledger` 在 S9 撞上同一堵墙，并且一度**无法自行恢复**：

> Original PlanReport preparation for Builder builder-s9-… remains pending: context deadline exceeded. Its successful dispatch was retained; save this same division answer after resolving the named original source.

> S9 的发卷则多次重试仍未落盘 …… Runtime 里虽有 `plan_report_packets` 记录（status=issued），但 `view_path` 与 `answer_path` 两个文件都不存在，于是导入器找不到 brief，也就 "no current registered Plan responsibility"。

核验结论：

1. **答卷形状不是问题。** 接收器两种形状都接受——`internal/cli/native_s9.go:445-452` 允许 `{"answer":{…}}` 信封式，也允许顶层专业字段的扁平式。S9 brief 也明写「The private outputs below use top-level professional fields, without an answer envelope」。agent 写的扁平三字段形状是对的。
2. **真正的问题是「写在哪里」。** `internal/cli/native_s9.go:425-432` 的接收器从**原生 worktree** 读，并明确拒绝 Main 根目录的副本：
   > actual native worktree PlanReport is missing at %s; **Main's draft cannot substitute for the Worker artifact**
3. **依赖链是闭环的。** work location 由 Main 在 PlanReport 被记录**之后**才准备（`internal/cli/repair_command.go:599`、`:1336`），而接收器又需要 worktree 才能读到那份产物；发卷准备一旦被 deadline 打断，登记与文件同时丢失，就没有可恢复的入口。
4. **终点.** `runtime repair status` = `status: planning`、`missing_reports` 与 `missing_results` 同时非空、`report_ref` 为空；回执 `accepted: false`、`error: "PlanReport has not been accepted"`、**`field_errors: []`**——没有任何字段级诊断。

> 值得肯定的是：S9 brief 本身写得清楚（相对 worktree 保存、Main 根目录只读、保存不等于接收、只有原消费者回执才算数），且 agent 没有放宽 Gate、没有伪造接收状态、没有改动产品字节——它只是被一个时间预算挡在了门外。

**落点.**

- `internal/cli/hook_budget.go:20`（`hookProcessBudget = 8 * time.Second`）
- `internal/cli/hook_budget.go:24`（`s4HookProcessBudget = 28 * time.Second`）
- `internal/cli/hook_budget.go:26-41`（`hookBudget()` 只对 `.claude/submissions/…/S4/` 路径放宽，其余一律 8 秒）
- `internal/cli/hook_budget.go:103`（`context.WithTimeout(ctx, hookProcessBudget)`，第二道 8 秒闸门）与 `:125-139`（超时文案与 `PreToolUse` 的降级返回；PostToolUse 同样以退出码 0 收场）
- `internal/cli/entry_questionnaire.go:263-266`（`mainQuestionnaireViewer` 决定哪些 PostToolUse 会走这条预算路径）
- `internal/cli/s7_status_test.go:156` 与 `internal/cli/controller_test.go:73`（框架刻意不把 `review-result submit` 放进 Main 的指引）
- 平台侧 `settings.json` 的 PostToolUse `timeout=30`（秒）——**抬高平台超时没有用**，内部 8 秒闸门先到期，与 `docs/framework-upgrade.md:23-24` 的说明一致
- `docs/framework-upgrade.md:18-26`（设计意图：Hook 只做快速控制面检查，重活应在 Hook 之外跑；S7 显然没有做到）
- **具体那个"重活"在哪：** `internal/review/workspace.go:175` 的 `detectUndeclaredProductDriftWithState(root, plan, state, root, contexts...)`——扫描 git diff 基线与未跟踪文件，找出未声明进 `frozen_subjects` 的产品文件。它是**整个仓库规模**的遍历，而它跑在 S7 ReviewPlan 注册路径里，也就是 Hook 的时限之内。

**框架自己的阈值与这个预算直接矛盾.**

`internal/metrics/metrics.go` 早就为"Hook 跑太慢"备好了判据与告警：

- `internal/metrics/metrics.go:39`（`defaultHookTimeoutMS = 10000`）与 `:524-529`（`hookPlatformTimeoutMS`：PostToolUse 30s，其余 10s）；
- `:400`（`healthDegraded`：任一 Hook 采样 ≥ 平台超时的 **40%** 即判定系统降级）与 `:512-520`（同样口径的 WARNING，且对 PreToolUse 额外打印 **"a platform timeout can bypass PreToolUse enforcement"**）。

把三个数放在一起：

| 事件 | 平台超时 | 框架降级阈值（40%） | 框架内部预算 | 关系 |
| --- | --- | --- | --- | --- |
| PreToolUse | 10 s | **4 s** | 8 s | **内部预算是降级阈值的 2 倍** |
| PostToolUse | 30 s | **12 s** | 8 s | 内部预算**低于**阈值 |
| PostToolUse（S4 特例） | 30 s | 12 s | 28 s | 已越过阈值 |

两个后果：

1. **PreToolUse**：任何触到内部预算的评估，按框架自己的标准就是"降级"，而框架的 metrics 模块恰恰为这个事件打印了"平台超时可以绕过 PreToolUse 强制"。框架知道这个失效模式，却让内部预算（8s）高于它自己设的降级线（4s）——这正是 BUG-SIM-18 fail-open 的成因。
2. **PostToolUse**：真正的杀手是 8 秒内部闸门，而框架的观测阈值是 12 秒。**被内部预算掐掉的 Hook（实测 8 242 ms）永远低于告警线，`healthDegraded()` 返回 false，也不会打出 WARNING。** 框架对自己正在遭受的这种失败完全失明——`callback-api` 一题就发生了 8 次。

**建议修法.**

1. **把「备下一份问卷」与「接收上一份 ReviewResult」拆成两个 hook 事件**。这是 `ui-ledger` 给出的最直接推论：先在一次"只做准备"的事件里把待准备的问卷备出来，随后的 Stop 事件就不需要准备，整份预算留给接收。
2. **超时必须让调用方看见失败**：现在退出码是 0，Claude 侧以为工具成功，实际状态只写了一半。应返回非零退出码或结构化的「未完成」标记，否则 Main 无从判断该不该重放。
3. **超时与漂移要分开措辞**：漂移检查跑不完时应说"检查超时，未能判定基线"，而不是"基线变了"。
4. 超时文案要给出**具体恢复命令**（当前只有「inspect the original receipt」这类无法直接执行的指示），并明确「不要由 Main 代交答卷」，否则超时会把接收责任悄悄转移给 Main。
5. **让观测覆盖真实失效点**：降级阈值应按**内部预算**算（PostToolUse 取 8s 的 40% ≈ 2.6s），而不是平台超时的 40%；否则框架对自己最频繁的失败失明。
6. 长期方向仍是 `docs/framework-upgrade.md` 写的那条：重活移出 Hook，按 SubagentStop → `runtime …` 外部调用路径走；给 S7 加预算只是治标。

**关联反馈.** `callback-api` S7 第一轮（`hook-*.json` 中 30 次 `internal budget`；最慢 8 242 ms；7 次删产物 + 6 次手工 submit）；`ui-ledger` 记录 20/21/22/24/25/27（该题后半程的逐项排除表与"漂移≠超时"判据）；`source-ledger`、`ui-ledger` 各 3 次预算超限。

**一份意外的旁证.** `callback-api` 的 agent 在读框架源码反推 S7 重派机制时，执行了

> `cd /home/sebs/go-system-builder grep -n -i "revis" blueprint/BUGS.md | head -20`

——它 grep 到了本文件（仿真环境与框架仓库同机，未提交的缺陷清单对它可见），并读走了 BUG-SIM-14/15/17 里关于 S7 基线漂移与重派恢复的描述。**agent 是靠这份缺陷清单解开自己的卡点的。**

这件事有两层含义：

1. **正面**：本清单记录的现象、恢复入口和代码落点，对真实使用者确实有用——它没有停在"这里有个坑"，而是写清了怎么绕过去、绕过去的代价是什么。
2. **反面**：当文档答不上来时，agent 会穷尽一切可用信息源——协议文档 → 编译后的二进制 → 框架实现源码 → 缺陷记录。**"猜机制"实际上已经升级为"逆向一切"。** 真实部署里没有源码、也没有缺陷清单可读时，这条路直接断掉；本清单被 agent 读到这件事本身，也提醒不要把仓库里的未提交分析当作 agent 可用的信息来源。

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

**证据（本轮实际触发分布，2026-10-25 06:40 全量重算 `hook-*.json`）.**

统计口径：逐个读 `{topic}/hook-*.json`，取 `started_ns`/`finished_ns` 算耗时，取 `stderr` 是否含 `internal budget` / `context deadline` 判定是否超限，取 `exit_code` 判定是否放行。

| 题材 | PreToolUse 超限（**全部 exit 0 = fail-open**） | PostToolUse 超限 | SubagentStop 超限 | SubagentStop 耗时中位数 | 内部预算 |
| --- | --- | --- | --- | --- | --- |
| `source-ledger` | 3 | 1 | 3 | **3 265 ms** | ≈6 658 ms |
| `callback-api` | 51 | 20 | 20 | **6 377 ms** | ≈6 658 ms |
| `ui-ledger` | **258** | 6 | 49 | **6 445 ms** | ≈6 658 ms |
| **合计** | **312** | 27 | 72 | | |

**这张表里最关键的一列是 SubagentStop 耗时中位数：3 265 ms → 6 377 ms → 6 445 ms。** 它随题材复杂度单调上升，而内部预算是死的 6 658 ms——**`ui-ledger` 已经贴在天花板上运行**（p90 = 7 265 ms、max = 8 296 ms、75 次里 60 次 ≥4 s、49 次直接超限）。这不是"偶尔慢一下"，是**接收路径的常态耗时已经等于预算**。

同时 `ui-ledger` 的 PreToolUse 超限达到 **258 次**（占其 1 645 次 PreToolUse 的 15.7%），每次都意味着**一次完整的门禁评估被丢弃**。

**实际影响：本轮为潜伏，未造成越权（这一点本轮复核后仍然成立）。** 用 `docs/control/protected-commands.json` 的**完整 19 条受保护命令族**逐条比对三题全部 fail-open 事件（`git push` 到受保护分支、`git merge --squash`、`git tag`、`gh release create/edit/delete/upload`、`gh pr merge`、`npm publish`、`goreleaser release`、`terraform apply/destroy`、`kubectl apply/delete`、`aws s3 cp/sync`、`sh -c`、`bash -c`、`npx --`）——**命中 0 次**。那 258 次 fail-open 放行的是 `python3 - <<'PY'` 读取状态、`find ... -delete`、`git stash/checkout`、`timeout 60 ./.claude/bin/loop-harness runtime review-result submit`、读框架源码等操作。

> **诚实口径.** 「命中 0 次」只能说明**本轮没有可利用的机会**，不能说明路径安全。实际暴露面比早先记录的「25 次」大了一个数量级（**312 次**），其中包含 `find -delete` 这类删除动作与手工 `runtime review-result submit` 这类代交通道——它们每次都在**未经任何策略评估**的情况下直接执行了。

**落点.**

- `internal/cli/hook_budget.go:130-141`（预算超限分支：`PreToolUse` 直接 `return 0`，根本没走到 `UnavailableDecision`）
- `internal/policy/unavailable.go:10-32`（框架自己的 fail-closed 口径：非只读工具一律 deny）
- `internal/policy/engine.go:1447`（protected-commands 表驱动硬 deny）、`:1013`（`phaseProductWriteFrozen` 冻结期写入禁止）、`:625` / `:666`（bug_resolution 阶段的写入约束）、`internal/policy/newnative_owned_check.go:27`（原生 Worker 自检归属）
- `internal/metrics/metrics.go:39`（`defaultHookTimeoutMS = 10000`）、`:524-529`（`hookPlatformTimeoutMS`：PostToolUse 30s，其余 10s）、`:400`（`healthDegraded` 阈值 = 平台超时的 40% = **PreToolUse 4s**）、`:512-520`（WARNING 文案里已经明写 **"a platform timeout can bypass PreToolUse enforcement"**）

**框架自己知道这个失效模式，却把内部预算设在降级线之上。** PreToolUse 的平台超时是 10s，框架自己的降级阈值是 4s，而内部预算是 **8s——正好是阈值的两倍**。也就是说：框架为「PreToolUse 被绕过强制」专门写了告警，却让内部预算允许这种情况发生两倍于自己的容忍线。

> 这一段是 `callback-api` 的 agent 自己在读框架源码时挖出来的（它去翻 `internal/metrics/metrics.go` 是为了搞清楚"基线扫描的 8 秒耗在哪"）。它找到的是框架的**观测口径**，但没有意识到**内部预算与该口径的冲突本身**就是缺陷。

**建议修法.**

1. PreToolUse 的预算超限必须走 `policy.UnavailableDecision`：非只读工具一律 `deny`，与「策略文件缺失」保持同一口径；
2. **把 PreToolUse 的内部预算降到降级阈值以内**（≤4s），否则框架自己的健康判据和它自己的预算互相打架；
3. 预算闸门应按事件类型分别定值：PreToolUse 的评估通常很轻（读策略表 + 运行时事实），8 秒不该被触发；真正需要长预算的是 PostToolUse/SubagentStop 里的重活（见 BUG-SIM-17），不该让它们和 PreToolUse 共用一个常量；
4. fail-open 必须在 stderr 与审计事件里留下**结构化标记**（哪条规则因超时未被评估），否则事后无法区分「评估通过」与「根本没评估」。

**关联反馈.** 无独立反馈条目（agent 只看到 stderr 文案，看不出自己被放行了）；本条由本次分析从 hook 观测文件与代码路径推出。

---

### BUG-SIM-19 · P1 · Runtime 状态随轮次膨胀，把每个 hook 的固定成本越推越高

**现象.** S7 一个轮次里每多一个 assignment，`loop-state.json` 就变大一截，而**每个 hook 事件都要把整份状态序列化并校验一遍**。`ui-ledger` 在第二轮实测：

| 题材 | `loop-state.json` 体积 |
| --- | --- |
| `source-ledger` | 611 186 B |
| `callback-api` | 684 520 B |
| `ui-ledger` | **742 336 B**（其中 `entities` 459 463 B） |

`entities` 里逐份存着 work-packet 的问卷 draft（每份含完整的题目、read 列表与答案模板）。于是：

1. **越靠后的责任越难被接收**——这解释了为什么同一轮里前 5 份 1–3 次就过，最后一份连试 11 次都过不来；
2. **单份答卷的体积反而不重要**（被接收的 51 502 B 结果 vs 失败的 5 463 B staged 结果），真正吃时间的是**框架自己那份越来越大的状态**；
3. 每多一个 reviewer assignment，就要多存一份完整问卷 draft，而问卷内容对状态机是**可再生的**（`DraftPlanForRoot` 随时能重建）。

**证据.** `ui-ledger` 记录 22 的原话：

> Runtime 状态（`loop-state.json` 742336 B，其中 `entities` 459463 B、逐份存着 work-packet 问卷 draft）随之变大，两件重活的耗时一起抬升，落在时限边缘上。

**建议修法.**

1. **不要把问卷 draft 存进 Runtime 状态。** draft 是 `DraftPlanForRoot` 可以确定性重建的派生数据，应当按需重建、或落到 `.claude/` 下按 packet id 寻址的文件，而不是塞进主状态并每次全量序列化+校验；
2. 给 Runtime 状态加**体积上限与增长告警**，并在 hook 预算里为状态体积预留时间，而不是假设它恒定；
3. 状态体积本身应当进入 `healthDegraded` 的判据——状态越大，Hook 越慢，这是可测量的。

**方案 A 详细设计（2026-10-25 增补，暂不动手）.**

> 结论先行：`entities.work_packets` 里真正需要持久化的是**身份+绑定**（packet id、assignment、轮次、基线 pin、digest），问卷 draft（题目全文、read 列表、答案模板）是 `DraftPlanForRoot(root, state, round)` 可确定性重建的派生数据。瘦身 = 存引用不存全文。

- **写入侧**（`internal/review/result_packet.go:150`）：`entities["work_packets"]` 改存 `{packet_id, assignment_id, review_round, baseline_generation, digest, draft_ref}`，其中 `draft_ref` 指向 `.claude/review/packets/<packet-id>.json`；draft 全文落该文件（Main/系统写，不手填）。
- **读取侧**：`FindReviewResultPacketContext` / `ReviewResultPackets`（`internal/runtime`）按 id 先读状态行，draft 按需从文件重建；`DeepEqual` 身份比对只比身份+绑定+digest，不比 draft 全文（draft 重建确定性由 `DraftPlanForRoot` 单测锁）。
- **校验侧**：16 MiB 预算上限保留；新增 entities 体积告警（超 500 KB 在 hook 里报 degraded，不阻断）。
- **测试**：重建一致性测试（同一 root+state+round 两次 `DraftPlanForRoot` 输出一致）；瘦身前后 packet 身份比对等价测试；体积告警触发测试。
- **与 SIM-17 的关系**：SIM-17 拆事件减少单次 Hook 的活，SIM-19 瘦身减少每次 Hook 的固定成本，两者正交，建议同分支先后做，先 SIM-19（改动小、收益确定）再 SIM-17（结构大）。

**关联反馈.** `ui-ledger` 记录 22。

---
### BUG-SIM-20 · P0 · 原生 agent 转录被整份读入并逐行 JSON 解析，与重试次数成正比地吃掉 Hook 时限

**这是本轮最重要的一条，也是 SIM-17 的首要根因。**

> **2026-10-25 06:45 实测修正（重要）.** 本条最初被写成"首要根因"，但直接计时后**必须下调其权重**：转录解析本身**只占 30–96 ms**（见下方「实测各环节耗时」），**不是 6.4 秒的来源**。本条仍然成立且仍然是真实缺陷——它让接收成本**随重试次数单调上升**——但**6.4 秒的真正来源是"一次 hook 事件里串行做了 5–10 次框架级操作"**，即 BUG-SIM-17 描述的结构性问题。**修 SIM-17 的"拆两件重活"与"重活移出 Hook"是主修，本条是次要的成本放大器。**

**现象.** S7/S6 的原生接收在 hook 里调用 `nativeReviewTaskText` / `nativeS6TaskText`，为了"从转录里找出最后一条提到当前精确引用的真实任务"，会：

1. 打开原生 agent 自己的**会话转录文件**（平台 payload 里的 `agent_transcript_path`）；
2. 读入**最多 4 MiB**；
3. 对**每一行**做一次 `json.Unmarshal`。

代码里有明确的"限制 IO"的意图，但限制值定得太宽：

```go
// internal/cli/native_review_hooks.go:331-352
func nativeReviewTaskText(in policy.Input, refs []string) string {
    f, err := os.Open(in.AgentTranscriptPath)
    ...
    const limit = 4 << 20                       // 4 MiB
    ...
    data, err := io.ReadAll(io.LimitReader(f, limit))
    lines := bytes.Split(data, []byte("\n"))
    for _, line := range lines {
        var row struct { ... }                  // ← 每行一次 json.Unmarshal
        json.Unmarshal(line, &row)
```

```go
// internal/cli/native_s6.go:272（S6 同一模式）
data, err := io.ReadAll(io.LimitReader(f, 4<<20))
```

**为什么会形成死循环.** 转录文件随 agent 的每一次尝试**追加**。第一次接收失败 → 同一个 reviewer 被 resume/重派 → 它的转录变长 → 下一次接收要读+解析更多 → 更容易超时 → 再失败 → 再重派……**代价随重试次数自我放大**，而预算只有约 6 秒。

**证据（转录大小与结果完全对应）.**

| 题材 | reviewer 子代理转录 | 结果 |
| --- | --- | --- |
| `source-ledger` | `agent-a8fe8023….jsonl` **700 KB** | ✅ 走通 S0→S11 全程 |
| `callback-api` | `agent-ab4d291d….jsonl` **831 KB** | 卡在 S7 第二轮 |
| `ui-ledger` | `agent-afe8bd11….jsonl` **2.4 MB**（重派 11+ 次，持续增长） | 卡在 S7 第二轮 |

主会话转录同向：`source-ledger` 7.3 MB、`callback-api` 22 MB、`ui-ledger` 28 MB。

**两个 agent 各自独立定位到了这一点.**

`callback-api`：

> The error changed to `signal: killed` — the drift scan's git-status subprocess now gets *zero* remaining budget, because the hook's pre-work (**notably scanning my 706 KB resumed-agent transcript**) consumed the 6 s request context. A fresh native Agent has a tiny transcript, so let me dispatch a new one to deliver the same answer.

`ui-ledger`：

> The error points at the *second* baseline verification (inside the CAS). Let me check whether the per-event bookkeeping files grow without bound — that would turn "fits in budget" into "doesn't" over a long run.

**agent 给出的处置本身就是这个诊断的验证**：换一个**全新的**原生 agent 去交同一份答卷——新 agent 的转录几乎为零，接收立刻成功。这与"接收成本 ∝ 转录大小"完全吻合。

> **⚠️ 2026-10-25 07:15 反证：这个处置在 `ui-ledger` 上并不成立.** `callback-api` 的 agent 报告"换新原生 agent 后接收立刻成功"，但 `ui-ledger` 在第 10 个窗口里**同样换新原生 agent 重派**（`agent-a00918d91322e21d5` 转录 **455 KB**、`agent-a34ad5122d3de60ba` 转录 **570 KB**，都明显小于原 agent 的 920 KB），**两个新 agent 的 SubagentStop 全部仍然失败**：
>
> | agent | 转录 | SubagentStop 耗时 | 结果 |
> | --- | --- | --- | --- |
> | `a886569ec7f89cebd`（原） | 920 KB | 8 117 ms | `S7 reception incomplete` |
> | `a00918d91322e21d5`（新） | 455 KB | 8 117 ms | `S7 reception incomplete` |
> | `a34ad5122d3de60ba`（新） | 570 KB | 6 520 ms | `S7 reception incomplete` |
>
> **所以"转录大小"不是接收成功的充分条件。** 两题的差异在于：`callback-api` 的卡点确实在 hook 前置工作（转录扫描）上，换新 agent 能省出预算；`ui-ledger` 的卡点更靠后——**预算在前序操作就被吃光**，漂移扫描与 stale 状态持久化连续超时，于是**每一轮都会留下一颗被保留的陈旧产物**，形成自我维持的死循环（本节上文「完整死锁链」）。**换 agent 能省 4 MiB 解析里的几十毫秒，省不出整个 hook 事件缺的 1–2 秒。**
>
> 这条反证进一步坐实了前面的修正：**主修是 BUG-SIM-17（拆事件、重活移出 Hook），本条只是次要放大器。**

**实测各环节耗时（`ui-ledger` 项目现场，2026-10-25 06:45）.**

| 环节 | 实测 | 说明 |
| --- | --- | --- |
| `loop-harness version`（**二进制冷启动**） | **555 ms** | 单是启动一个进程就要 0.55 s |
| `loop-harness s7 status` | 389 ms | 一次只读状态查询 |
| `loop-harness runtime recover inspect` | 570 ms | 一次只读运行时检查 |
| `loop-state.json` 解析（499 KB） | 24 ms | **不是瓶颈** |
| `git status --porcelain` | 30 ms | 工作树当前干净 |
| `git diff --name-status HEAD` | 45 ms | |
| 主会话转录尾部 4 MiB 逐行 `json.Unmarshal`（30.4 MiB 转录） | **44 ms**（1 460 行） | **不是瓶颈** |
| 子代理转录逐行解析（最大 2.32 MiB / 715 行） | **96 ms** | **不是瓶颈** |
| **`SubagentStop` 事件实测总耗时** | **中位 6 445 ms / p90 7 265 ms / max 8 296 ms** | 内部预算 ≈6 658 ms |

**这张表给出的结论很明确：没有任何一个单独的慢操作。** 二进制启动 0.55 s × 若干次串行、每次状态查询 0.4–0.6 s、加上 Git 与文件 I/O，再叠加 S7 接收路径上「加载状态 → 扫工作树 → 编码 finding → 暂存 observation batch → 提交 → **再准备下一份问卷** → 写状态」，**在一次 8 秒的闸门里串行跑完，就是 6.4 秒的来源**。

所以问题不是"某一处写得慢"，而是**"把一串各自 0.4–0.6 s 的框架操作塞进一次 hook 事件、并且给它们共用一份 6.6 s 的预算"**。这条修正直接强化了 BUG-SIM-17 的第 1、6 条建议（拆事件、重活移出 Hook），也说明**只优化转录读取（SIM-20）救不了这个场景**。

**落点.**

- `internal/cli/native_review_hooks.go:331-364`（S7：`nativeReviewTaskText`，4 MiB + 逐行 `json.Unmarshal`）
- `internal/cli/native_s6.go:262-285`（S6：`nativeS6TaskText`，同样的 4 MiB + 逐行解析）
- `internal/cli/hook_budget.go:20`（8 秒内部预算）、`:128`（`context.WithDeadline`）
- `internal/policy/engine.go:322`（`AgentTranscriptPath` 来自平台 payload）

**建议修法.**

1. **不要整读整解析。** 目标只是"找到最后一条提到当前引用的真实任务"，应当**从文件尾部反向扫描、只解析有限行数**（例如最后 200 行），命中即停。现在的做法是读 4 MiB 再逐行解析全部。
2. **把上限从 4 MiB 降到与预算匹配的量级**（例如 256 KB），并在截断时**明确报告"未能在预算内定位任务引用"**，而不是静默回退到 `nativeS6TaskText`。
3. **优先用平台已经给好的 `last_assistant_message`**：payload 里本来就有它，`nativeS6TaskText` 已经优先用它；S7 侧应同样优先使用，避免为找一句话去翻整份历史。
4. **更根本的**：接收不应依赖"从历史转录里猜"agent 到底做了什么。正确来源是**本次派发时写入的 packet/assignment 记录**——它本来就在 Runtime 里，转录只是旁证。
5. 建议在 hook 时限之外做**独立的耗时分解**（转录读取 / 漂移扫描 / 状态序列化各占多少），否则每次超时都只能靠猜。

**关联反馈.** `callback-api` 与 `ui-ledger` 后半程的逐项排除记录（由本次分析从会话与转录体积对照中坐实）；两者在 S7 第二轮的"最后一份答卷永远收不进来"现象即由此导致。

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
| SIM-02 S8 route 问卷「reuse fixed reference」 | ✅ 已修复并提交（`cb70c61`） | 问卷指令改为封存 S7 引用可复用/新路径 S8 捕获；`finding_source_handoff_test.go` 锁封存引用路由 |
| SIM-03 `planned_paths` 无语义 | ✅ 已修复并提交（`18bbf11`） | 新 brief append 写集合定义 + 协议 S6 同句；历史 compact/legacy 字节冻结不动 |
| SIM-04 S7/S8 答卷顶层形状 | ✅ 已修复并提交（`438f8e7`） | 模板 + L3-S7/S8 协议四处同句；S9 结构不同已确认不改，S2–S5 无返工不扩大 |
| SIM-05 恢复冻结源静默删文件 | ✅ 已修复并提交（`3a26782`） | 新增 `runtime review-plan frozen-diff` 只读三列表 + revive 指路；Main 只传 `--root`，pinned SHA 来自 plan |
| SIM-06 `worktree is dirty` 主语错 | ✅ 已修复并提交（`619f538`） | 集成失败路径已带归属；深挖补齐清理三处裸返回，全路径 `Builder checkout %q is dirty` |
| SIM-07 `revive/revise` 分工不明 | ✅ 已修复并提交（`e9c026f`，指引层见 `3a26782`） | help 示例 + 错误串 + 协议三段表现已齐；补 `--file` 非 JSON 前置拒收 |
| SIM-08 `pending system_followup` 混义 | ✅ 已修复并提交（`bd95960`） | S5 板拆两句 + S2/S4/S5 回执加定语；整理见 `blueprint/SIM-08-followup-note.md` |
| SIM-09 S9 两表同 `assignment_id` | ✅ 已修复并提交（`ff92af9`） | 验证者指引点名共用 Builder anchor + 禁手写验证者 ID；问卷测试锁两句 |
| SIM-10 ChangeImpact「未准备」无入口 | ✅ 已修复并提交（`7f76365`） | 新增 `impact prepare` + 三处指路 + ID 双来源等价口径；顺手正 70 行措辞 |
| SIM-11 超时 vs 内容非法未区分 | 真 Gap，SIM-17 的子集 | `internal/cli/hook_budget.go:fail()` 仍是同一句话 + 退出码 0；未区分「超时可原样 resume」vs「内容非法须重写」 |
| SIM-12 S10/S11 口径 | ✅ 已修复并提交（`6795e11`） | schema description + 双问卷口径 + 两个文案断言测试；S11 申请无 schema 校验记为后续 gap |
| SIM-13 私有文件≠已接收无反馈 | ✅ 已修复并提交（`4d3138d`） | S9 brief 已有；S6 brief 补同款 `s6 status` 查询句；断线续跑机制本来全 |
| SIM-14 未归属路径否证注册 | ✅ 已修复并提交（`15c4efd`） | deleted surface 不进 frozen_subjects，注册只看归属；漂移文案改挂牌口径；测试跟新语义 |
| SIM-15 TASK 模板两句话 | ✅ 已修复并提交（`82b0fc3`） | §2 相对本 TASK 目录 + §8 空表禁占位；语义测试通过 |
| SIM-16 措辞归集 9 项 | ✅ 已修复并提交（`09be9e8`） | S7 问卷补 E2E N/A/合并/seed 口径 + 中断恢复；S0 锁定触发=落盘；余项此前已修 |
| SIM-17 S7 Hook 8s 预算 | 真 Gap，P0，工作区未动 | `internal/cli/hook_budget.go:20,26-41,103` 原样：8s 通用、仅 S4 放宽 28s、超时退出码 0、半写状态、无认可恢复入口；三题全触发。**2026-10-25 06:40 全量重算：`SubagentStop`（=S7 接收事件）超限 `source-ledger` 3 / `callback-api` 20 / `ui-ledger` **49** 次，耗时中位数 **3 265 → 6 377 → 6 445 ms**（内部预算 ≈6 658 ms，`ui-ledger` 已贴天花板，p90 7 265 ms）；`ui-ledger` 框架自记 `loop_s7_result_submits_total = {accepted:5, rejected:52}`（被拒率 91%）。根因：一个 hook 事件同时做"备下一份问卷"+"接收上一份 ReviewResult"，两件约 6s 的活共用一份约 6.6s 时限；配合 Runtime 状态膨胀（SIM-19）与漂移扫描遍历工作树。机器负载是放大器不是根因（负载降到 1.44 仍失败）。**2026-10-25 06:45 实测修正：6.4s 没有单一慢操作——二进制冷启动 555ms、`s7 status` 389ms、`runtime recover inspect` 570ms、状态解析 24ms、git 30–45ms、转录解析 30–96ms，串行累加即达预算** |
| SIM-18 PreToolUse 预算超限 fail-open | 真 Gap，P1，与 SIM-17 同源但性质不同 | `internal/cli/hook_budget.go:fail()` 在策略判断前对 PreToolUse 直接 `return 0`，绕过 `policy.UnavailableDecision` 的 fail-closed（非只读工具应 deny）。**2026-10-25 06:40 全量重算：PreToolUse 超限 `source-ledger` 3 / `callback-api` 51 / `ui-ledger` **258** 次，合计 **312 次，退出码全部为 0（即全部 fail-open 放行）**；`ui-ledger` 一次 PreToolUse 事件数 1 645（占 15.7%）。按 `docs/control/protected-commands.json` 全部 19 条受保护命令族逐条比对，**命中 0 次 → 仍属潜伏未利用**，但放行内容包含 `find -delete`、`git stash/checkout`、手工 `runtime review-result submit` 等未经任何策略评估的操作。根因补充：PreToolUse 降级阈值 4s（10s×40%），内部预算 8s 正好是阈值两倍；框架 metrics 已打印过 "a platform timeout can bypass PreToolUse enforcement" |
| SIM-19 Runtime 状态随轮次膨胀 | 真 Gap，P1（本次分析新增） | `loop-state.json` 611 KB→742 KB（`entities` 459 KB，逐份存 work-packet 问卷 draft）；每个 hook 都要全量序列化+校验，导致**越靠后的 assignment 越难被接收**；问卷 draft 是 `DraftPlanForRoot` 可确定性重建的派生数据，不应进主状态。**实测：状态解析本身只要 24 ms，不是 6.4s 的主因；它的危害是「越靠后越难被接收」与「全量序列化+校验随轮次线性变慢」** |
| SIM-20 原生 agent 转录被整读整解析 | 真 Gap，P0，**本次分析新增**；**2026-10-25 实测后由「SIM-17 首要根因」下调为「次要成本放大器」** | `internal/cli/native_review_hooks.go:331-352`（S7）与 `internal/cli/native_s6.go:272`（S6）每次接收都 `os.Open(AgentTranscriptPath)` → 读最多 **4 MiB** → **逐行 `json.Unmarshal`**；转录随重试次数追加，接收成本自我放大。走通全程的 `source-ledger` reviewer 转录 700 KB；卡住的 `callback-api` 831 KB、`ui-ledger` 2.4 MB（重派 11+ 次）、主会话 30.4 MB。**但直接计时：主会话尾部 4 MiB 解析 44 ms、子代理最大转录解析 96 ms——只占 6.4s 的 1.5%，不是首要根因。主修仍是 SIM-17 的「拆事件 / 重活移出 Hook」** |
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
