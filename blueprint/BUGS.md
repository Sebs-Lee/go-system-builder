# 仿真缺陷清单（2026-10-24 完整流转仿真）

本文件记录 2026-10-24 三题材完整流转仿真（S0～S7～S9～S7～S10～S11）暴露的、**尚未修复**的框架缺陷。

开头的仿真现象和计时保留其原执行环境含义；当前缺陷状态与修复边界以各条及定级表为准。尤其普通 PreToolUse 已按进度监听合同实现，不能从旧门禁措辞或历史 exit 0 统计推出需要恢复工具 deny。

它不是设计文档，也不是调查流水账：每条只记录「现象 → 证据 → 落点 → 建议修法」，供后续修复与 blueprint 同步更新使用。修复完成后应把对应条目连同修复提交一起删除或标记为已修，不在本文件累积历史。

仿真环境位于仓库外 `/home/sebs/gsb-simulation/`，三题材共用同一份发布包：

| 题材 | 形态 | 终点 | 机械校验 | 机制疑问 |
| --- | --- | --- | --- | --- |
| `source-ledger` | NFR-only、8 份仓库现成 fixture 作对照基线 | **`release_authorized`（终态）**，S0→S7→S9→S7→S10→S11 全通 | 12/12 阶段全绿，`mechanical_checks_passed=true`，`fixture_unchanged=true` | 28 条 |
| `callback-api` | 服务行为模块、UI impact=none | **`release_authorized`（终态）**，rev=23、clean_round=2 | 21 个机检 gate 全过，`mechanical_checks_passed=true`，`fixture_unchanged=true`，`registered_evidence_sources_intact=true` | 28 条 |
| `ui-ledger` | UI impact=changed、无候选文档 | **未到终点**：`verification/running` rev=20，S7 第二轮 7 份责任里 5 份 `done`、2 份 `active` | — | 36 条 |

`source-ledger` 的 S8 再次在 `no_competing_hypothesis` 上踩到同一个坑（见 BUG-SIM-01）。`callback-api` 的 S7 接收被 Hook 内部预算反复打断（见 BUG-SIM-17），连续多个窗口停在同一 revision，最后靠"删 stale 产物 + 清空工作树 + 换新原生 agent"突破并跑到终态；`ui-ledger` 同一缺陷在 S9 变成死路、在 S7 第二轮变成**闭环**（见 BUG-SIM-17 的接收与恢复问题）。

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

### BUG-SIM-11 · P2 · 接收中断、专业拒错和后继未完成混淆

**现象.** 原仿真同一答案或 Plan 在后续接收成功，首次却收到 `Consumer rejection: context deadline exceeded`，并同时出现拒收、Plan 注册中断、材料准备延期等提示。超时不能证明内容非法，缺少回执也不能证明未提交。

**设计.** 按 [Hook §10.2～§10.3](l4-mechanisms/L4-hook-platform-wiring.md#102-各层工作与复用边界)贯通一次准确输入与分段恢复：未提交时复用原输入，pending 形成后恢复原事务，接收成立后只续原出口／材料。反馈分别表达专业字段错误、来源故障、并发冲突、传输中断和已接收后继未完成；不要求作者改内容、重存、删产物或换 ID。

**实现状态.** 一次答案输入已贯通服务与原消费者，pending／回执／本次暂存按进度续接，基线 context 和取消错误链已接通。Result 尾段锚定原工作截止后 1.4 秒，Plan 为 400 毫秒，延迟返回不重新发放预算。冻结验收曾发现 OS 锁释放后，遗留兼容标记的 30 秒过期窗口超过新进程约 5 秒锁预算，阻断立即恢复。本轮原锁入口已改为：证明原 PID 退出后立即回收标记，已知活进程不因标记年龄被驱逐；未知旧格式仍沿原兼容规则。[真实进程回归](../internal/review/result_process_recovery_test.go)继续检验新进程原入口恢复、同答幂等与篡改／缺失 intent 拒绝，不能用执行 defer 的 panic 测试代替。仍无跨进程自动执行者，强制终止后的无人触发恢复不在已实现能力内。

**验收.** 覆盖两层读取之间改文件、有限恢复固定原输入、各提交窗口中断、响应丢失和后继失败；同时保留真实专业拒错及同身份不同内容冲突。与 SIM-17 共用正常路径优化，故障语义仍须独立验证。

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

### BUG-SIM-17 · P0 · 专业接收链重复处理历史与来源，耗尽 Hook 预算并放大恢复负担

**现象.** 原仿真 S7／S9 接收多次耗尽内部预算，已填写答案不能稳定变成持久接收及后继交接。Main 曾通过删暂存、手工代交或换 Agent 排障，这些不是正常作者合同。历史 `ui-ledger` SubagentStop 中位约 6,445 ms、p90 约 7,265 ms；这些数值描述原现场，不是当前版本性能。

**问题关系.** SIM-17 是整条正常接收与中断交接的主问题。SIM-11 是其故障分类／恢复侧，SIM-20 是转录历史成本放大器；SIM-18 必须按现行进度监听合同处理，不能恢复普通工具门禁。SIM-19 改善每次处理的状态体积，与调用链重复工作正交，二者可独立实施。

**原排查位置与修复分工.**

| 位置 | 具体问题 | 修复方向 |
| --- | --- | --- |
| `submission.SaveReviewResultFrom` → `review.SubmitIssuedResult` | 相邻两层各读、校验答案 | 原边界固定一次输入，向下传实际字节及绑定，恢复沿同次输入 |
| `detectUndeclaredProductDriftWithState` | 枚举全部历史 packet 后逐份验证；无请求 context 的入口又重建完整 registry／解析 journal | 按实际待判断路径选择证明，贯通 context，共用一次准确索引与解析 |
| `buildsurface.Read／Verify`、`gitBytes` | 固定来源重复重建，部分调用另起 Background／局部时限 | 固定计算同次复用，当前连续性在原边界复核，Git／锁继承剩余预算 |
| Writer、checkpoint | 重复完整候选处理，正文体积放大成本 | 相同候选下复用准确结构计算；独立恢复和落盘完整性仍校验，存储按 SIM-19 简化 |
| 原生返回与后继 | 已接收和出口／材料失败须区别；重跑整链会放大故障 | 先认原事实再续未完成步骤，不重收作者答案 |
| 审计追加 | 每次扫描历史去重，锁另设 30 秒上限 | 保留事件和去重，采用可验证复用／索引，等待受父请求约束 |

**设计依据.** [Hook §10.2](l4-mechanisms/L4-hook-platform-wiring.md#102-各层工作与复用边界)定义一次接收顺序、准确复用、当前连续性与外围读取；[§10.3](l4-mechanisms/L4-hook-platform-wiring.md#103-暂存中断与原事务续跑)定义恢复；[§10.4](l4-mechanisms/L4-hook-platform-wiring.md#104-实现落点依赖与尚未闭合的路径)标明现状和依赖；[§10.6](l4-mechanisms/L4-hook-platform-wiring.md#106-整条接收链的验收与观测)定义整链验收。各阶段保持原责任、原消费者和原 Writer，不新建接收状态机。

**范围与现状.** 一次输入、按实际路径批量证明、基线／固定 Git 的请求预算与复用、按进度恢复、审计索引及外层失败观测已在工作区接入生产调用链；锁内仍核实际磁盘证据，取消／局部超时不写成内容非法或 Plan stale。当前原生 S7 Stop 不无条件发行下一 Reviewer 问卷，最后结果仍可续 S10；不能继续用“每次接收必做两件重活”概括当前机制。主修是减少实际工作及按进度续接，不是增加预算、无条件拆成两个平台事件、等待下一普通工具或交给 Main 手工 submit。

**验收缺口.** 需用当前版本复测完整链路、历史增长、并发和各中断窗口。组件缓存命中、旧仿真数字、单份答卷变小或换新 Agent 后偶次成功，均不能替代验收。

**冻结验收复现的机制缺口.** 除 SIM-11 的死亡进程兼容锁外，[Main 返回取消检查](../internal/cli/hook_acceptance_cancellation_linux_test.go)证明 S9 摘要在请求取消后仍打开原工作包正文，相关 lookup／diagnostic 仍另起 Background；[诊断与基线检查](../internal/integration/hook_diagnostic_acceptance_test.go)证明原机制发布的 `.claude/diagnostics/planning/` 准确诊断文件会进入未声明产品漂移，进而阻断原 Gate。两者都不能靠让作者补字段、删诊断或重开验证轮解决。[S9 证据路径检查](../internal/cli/hook_acceptance_s9_evidence_test.go)还证明，按指引推荐的 `path:logs/check.txt` 保存检查证据后，原候选消费者以未跟踪日志导致 Worker 不干净而拒收；不能用额外证据提交／清理掩盖交接缺口。

**验收后的定向收敛.** 按 [Hook §10.7](l4-mechanisms/L4-hook-platform-wiring.md#107-已复现交接断点的收敛设计)补齐上述边界：Main 返回的正文、只读回执与 S9 诊断继承父 context；显示诊断按准确文件和有限内容分类，共用于 S7 产品漂移、S9 基线和集成检查；S9 原消费者只投影本次已捕获且字节匹配的未跟踪日志输出，贯通候选、Main 集成及已验证后的清理，Main 不可变证据保留。未知邻居、源码、已跟踪或已在原 Session 基线中的文件、变更字节仍保留原拒绝条件；未增加作者字段、日志提交／忽略手续或工具门禁。[S9 交接回归](../internal/cli/s9_preferred_evidence_integration_test.go)覆盖推荐路径到 Main Final／清理及证据保留，补足仅候选通过的缺口。原 smoke 工具按 PostToolUse 30 秒、其他事件 10 秒校验，不改产品时限。四项原失败用例与相邻负例为本轮验收范围；真实 Claude S9、完整强杀窗口、整链增长及其他未归因红灯仍未闭合，SIM-17／11 保持开放。

**当前复验结论.** 四项原失败断言与相邻边界共 13 项已通过，原 smoke、安装检查及适用 race 复验通过；常规回归仍有红灯。真实 S9 最终形成 Main 集成检查、Final 和 complete，但经历首次 Stop 超时、未引用临时脚本拒收、额外本地忽略规则及替代原 actor 继续，原平台用例仍失败。模型工具请求携带的 `Agent.resume` 在实际 Hook 输入中消失且启动新 actor，须核对客户端继续接口，不能把模型自行替代当预定流程通过。另一个具名缺口是 [Main 集成入口](../internal/cli/workspace_delivery.go)：execution 尚无 DeliveryRef 时直接以空路径读取，只报 `invalid packet artifact path`，未交付“候选未形成”的前置事实和原恢复责任。展示 `Next: .` 与 Hook 文档 §12.2 的旧 smoke 描述也仍有残留。以上不由扩大日志豁免或增加作者手续解决；完整增长、中断与无额外指引的安装包流程仍待验。

**执行清单.** 使用 [Hook 接收、指引与恢复验收清单](l4-mechanisms/L4-hook-acceptance-checklist.md)分别收集 SIM-17／11／18／20 与正交的 SIM-19 证据；特别检查实际模型渠道、接手责任、恢复动作及是否依赖额外作者步骤。未验条目保持待验，不因清单或实现存在就关闭缺陷。

**关联反馈.** `callback-api` S7 接收；`ui-ledger` 记录 20／21／22／24／25／27；`source-ledger` F-026。

---

### BUG-SIM-18 · P1 · PreToolUse 超时的旧门禁归因与现行监听合同冲突

**原记录.** 历史仿真统计到 312 次 PreToolUse 超限且 exit 0。这个记录可以证明观察经常未完成，不能仅凭退出码证明当前框架绕过了一项本应执行的工具 deny。

**现行合同.** [Hook 普通入口](l4-mechanisms/L4-hook-platform-wiring.md#61-本人保存调用本人消费者)提供进度与准确材料；[PreToolUse 渲染器](../internal/hook/pretooluse.go)不输出 `permissionDecision`，[对应测试](../internal/hook/pretooluse_test.go)要求旧 block 反馈不转换为工具 deny。`UnavailableDecision` 和兼容 adapter 的旧能力不能推导当前入口必须限制工具。

**仍需解决的问题.** 超时可能丢失当前进度、错误诊断或必要交接反馈；外层真实失败已接入现有 observation spool 和 health，低于平台告警线也保留。必须区分“工具继续”“观察未完成”“领域已接收”与“阶段已推进”，不把其中一个当作另一个的证明。

**修复方向.** 与 SIM-17 共用轻量分流、准确状态复用和有界观测；按 [Hook §10.6](l4-mechanisms/L4-hook-platform-wiring.md#106-整条接收链的验收与观测)保留外层传输失败事实及已知接收信息。废弃“所有非只读工具 deny”和仅缩短 Pre 时限的旧建议。领域来源、批准、责任与转换条件继续由原消费者／Writer 核对，不能用放宽接收换取非阻断。

**验收.** 普通 Pre 超时／未知不新增工具 deny，不发布部分成功，不制造接收或阶段事实；真实领域缺项仍在原入口拒收。保持作者权限与交付动作不变。

---

### BUG-SIM-19 · P1 · Runtime 内嵌历史工作包全文，放大后续读取与提交成本

**现象.** 已签发工作包的完整 `binding`、`inputs` 和 `draft` 随责任和轮次追加到 `entities.work_packets`。普通状态读取及上下文转换承担整份状态的解析成本，写事务还要复制、校验和持久化完整状态，并发布 checkpoint。每个 Hook 的路径不同，不能概括为“每次 Hook 都执行全部校验”；历史正文持续进入当前处理路径的问题仍然成立。

**证据与归因边界.** 原仿真记录的 `ui-ledger` 状态为 742,336 B，Reviewer 后续接收反复接近或超过时限。体积增长是成本放大因素，但该记录不能单独证明全部超时由状态体积导致；SIM-17 的串行工作量、来源扫描和运行环境须分别测量。不能由一份较小答卷失败推导答卷体积永远不影响成本。

工作包膨胀不仅来自 S7 题目；S9 PlanReport 与 RepairResult 还重复携带共同冻结来源。只移出 S7 draft，不能视为整个 Runtime 存储问题已经解决。

**落点.** [S7 签发](../internal/review/result_packet.go)保存完整 packet；[Runtime Writer](../internal/runtime/store.go)和 [checkpoint](../internal/runtime/checkpoint.go)处理完整状态；[S7 registry 缓存](../internal/runtime/review_result_read_cache.go)已减少同请求重复重建，但命中仍需计算完整 registry 的内容身份，且未改变持久化表示。

**修复设计.** 采用[运行控制面 §2.4～§2.6](l4-mechanisms/L4-runtime-control-plane.md#24-保留事实与保存全文是不同决策)的主状态索引与独立完整正文分工，按当前责任读取。保留专业产物、原签发／接收事实和历史问卷，不要求 Agent 改变输入、答案、保存、结束或恢复操作。优先移出完整 packet 并减少重复处理；共同来源去重按收益另行评估，不先引入全局对象库或后台队列。

原问卷恢复依赖原冻结内容和受支持的模板身份。`DraftPlanForRoot` 生成当前 ReviewPlan 草稿，不是历史问卷恢复器；不得用当前 Plan 或最新模板补造旧题目。正文先发布、引用后提交，旧格式、缺失展示文件、pending、checkpoint 和归档的兼容要求由[状态机核心 §8.4](l4-mechanisms/L4-state-transition-core.md#84-工作包正文外置的事务与兼容边界)统一定义。

已有 16 MiB checkpoint 状态硬上限保留；不新增未经测量校准的固定 500 KB 降级阈值。验收按 [Hook §10.5](l4-mechanisms/L4-hook-platform-wiring.md#105-存储减负的行为等价与性能验收)同时检查作者行为等价、历史恢复、最终结果交接与增长场景下的实际成本，不能以缓存命中或文件缩小代替。

**实现状态.** 工作区已实现 S7 ReviewResult、S9 PlanReport／RepairResult 的轻索引、完整不可变正文、按责任读取、原 Writer 新写及显式旧状态维护，并接入旧 pending 和独立 checkpoint／恢复计划正文副本。其他阶段仍内嵌，无 GC 或正文内部共同来源去重。主状态精简不等于删除完整恢复内容；按行为、状态体积与真实读取成本分别验收。

**关联反馈.** `ui-ledger` 记录 22。

---
### BUG-SIM-20 · P1 · 转录任务定位仍有整段读取与重复 fallback 成本

**现象与权重.** 原实现整段读取并逐行解析转录。历史测量解析约 30～96 ms，不能据此解释整个接收链约 6.4 秒，也不能以换 Agent 后偶次成功证明唯一根因。它是 SIM-17 的次要历史成本放大器；有界读取达到上限后，也不能继续宣称成本与会话总长度无限线性增长。

**当前落点.** [分块定位器](../internal/cli/native_task_transcript.go)已接入 Review 与 S6 入口，按 64 KiB 倒序读取，近尾准确命中立即停止；原尾／头各 4 MiB 覆盖范围保留，fallback 共用累计 8 MiB 上限及区间缓存，重叠字节不重复读取。普通无关事件继续跳过转录。诊断 helper 区分匹配、边界、读取失败与取消，兼容字符串入口保持原交付形式。

**设计.** 按 [Hook §10.2.3](l4-mechanisms/L4-hook-platform-wiring.md#1023-转录与审计也属于接收成本)分块倒序读取，命中原准确任务即停，fallback 共用累计读取及时间预算。保留真实 user 任务、原引用边界、历史别名、多责任冲突和工具结果排除规则。任务定位不替代原 packet／owner／来源证明。

不直接改成任意最后 200 行或 256 KB，不用最终助手消息无条件替代真实任务，也不要求作者缩短会话或换 Agent。无匹配、越界和 I/O 失败分别可诊断，不能静默选错责任。新路径须证明任务选择与恢复行为等价，并测实际读取字节、解析行数与整链耗时。

**关联反馈.** `callback-api` 与 `ui-ledger` 的转录排障记录；完整接收故障仍归 SIM-17。

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
| SIM-11 接收中断与内容拒错 | 死亡兼容锁已定向修复；完整中断矩阵待验 | 原入口按已退出 PID 立即回收，活进程负例及真实强杀回归保留；无人触发自动执行尚未实现 |
| SIM-12 S10/S11 口径 | ✅ 已修复并提交（`6795e11`） | schema description + 双问卷口径 + 两个文案断言测试；S11 申请无 schema 校验记为后续 gap |
| SIM-13 私有文件≠已接收无反馈 | ✅ 已修复并提交（`4d3138d`） | S9 brief 已有；S6 brief 补同款 `s6 status` 查询句；断线续跑机制本来全 |
| SIM-14 未归属路径否证注册 | ✅ 已修复并提交（`15c4efd`） | deleted surface 不进 frozen_subjects，注册只看归属；漂移文案改挂牌口径；测试跟新语义 |
| SIM-15 TASK 模板两句话 | ✅ 已修复并提交（`82b0fc3`） | §2 相对本 TASK 目录 + §8 空表禁占位；语义测试通过 |
| SIM-16 措辞归集 9 项 | ✅ 已修复并提交（`09be9e8`） | S7 问卷补 E2E N/A/合并/seed 口径 + 中断恢复；S0 锁定触发=落盘；余项此前已修 |
| SIM-17 接收链工作量与交接 | 四项及相邻复验通过；整体仍有红灯 | 推荐日志到 Main Final 已验；真实平台继续兼容、缺候选的集成诊断、全量回归及整链增长／中断矩阵未闭合 |
| SIM-18 PreToolUse 超时归因 | 旧全局 deny 修法废弃；监听与观测仍需减负 | 现行入口不决定工具许可；超时不得误报接收／推进，领域条件仍由原消费者检查 |
| SIM-19 Runtime 状态随轮次膨胀 | S7／两类 S9 外置和恢复已落盘，验证中 | 新写轻索引，旧状态显式维护，保留完整原题及独立恢复副本；其他阶段仍内嵌，无 GC 或内部共同来源去重 |
| SIM-20 转录历史放大器 | 分块早停与共享 fallback 预算已落盘，验证中 | 保留准确任务与原尾／头覆盖范围，命中早停且重叠区间不重读；不增加作者动作 |
| 存量 1–4 | 已在工作区修复 | `docs/rules/ui-prototype.md` + `skills/ui-prototyping/SKILL.md` 去四 token 校验说辞；`L4-hook-platform-wiring.md:33` 改 `*`；`review-plan.schema.json` 删 `behavior_surface`；`submission-receipt.schema.json` 补 `attempt_status=rejected` |
| 存量 5–6 | 已在工作区修复 | `action_catalog.go` 已补新命令；`internal/evidence/local_path.go:+30 ValidateLocalCaptureSource` 黑名单已加 |
| 存量 7（4 个失败测试） | 已在工作区修复，待跑全量 | `s6_prepare_test/integrate/investigation/repair_plan_report` 均有对应更新 |
| FIX-01～05 | 夹具侧，非框架缺陷 | 反馈文件进产品树、`candidate-inputs.json` 缺 stage、45min 切片、磁盘打满；修仿真夹具，不进框架修复分支 |

### 2. 修复总原则（四条红线）

1. 不新增手填字段：所有恢复/重交入口只接受问卷文件路径或登记 ID，digest/sha 由系统回填，CLI 拒收手写 `sha256=`。
2. 问卷替代传参：`revive/revise/ChangeImpact/S9 division/超时重放` 全部走已发行问卷 `answer`，`--file` 只收问卷 JSON 路径，不收内联 JSON。
3. 失败可见：传输中断、来源漂移、专业拒错和已接收后继未完成分别反馈，保留准确原引用和具名缺口。退出码遵循事件合同，exit 0 不证明接收；不以统一非零、工具 deny 或要求作者代交作为恢复方案。
4. 知识绑到人：brief/`instruction`/表单写清「谁、填什么、不填什么」，文档语义下沉到 Worker 眼前，不指望 Worker 读 blueprint。

### 3. 接收主线与相关修复（SIM-17／11，兼顾 SIM-18／20）

A. S7 同步接收的机制减负（SIM-17／SIM-11，替代此前“移到手工 submit／加长预算”建议）：

- 原保存／SubagentStop 继续调用原消费者，正常本次完成适用出口与下一材料交接；不让 Main 代交、作者重存或等待下一普通工具触发。
- 框架复用同次初始路由快照、准确不可变来源与已验证registry／journal解析；固定Git批读最多四路，当前来源与Writer仍核真实条件。
- 暂存归属intent、原Writer互斥发布与消费、有限收尾和原输入续跑承担中断恢复。外层保留具体stderr，部分stdout不证明成功；超时和取消指标独立于内容拒错。
- 不新增生命周期状态机、不改变工具权限、不增作者字段或手填摘要；无跨进程执行者时如实保留强制终止后的执行限制。SIM-18 遵循当前进度监听合同，不恢复普通工具门禁。

D. PreToolUse 监听与传输观测（SIM-18）：

- 保持普通工具非阻断，移除将旧 `UnavailableDecision` 当作当前入口修法的建议。
- 与 SIM-17 共用状态读取和预算减负；内部超时单独可见，不以平台告警线或 exit 0 推导观察完成。
- 原具名领域消费者保持来源、责任、批准与转换检查，未知不得生成接收或推进事实。

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
2. 用当前版本重放 `callback-api`／`ui-ledger` 原接收及中断窗口，在既定事件预算内检查输入、暂存、pending、接收与后继；增长场景分别记录历史 packet、日志和转录的成本，不把 28 秒 S4 特例移用到 S7。恢复删除场景确认现存文件保留并准确报告；
3. `blueprint/L3-S7/S8/S9、L4-hook-platform-wiring、L4-state-transition` 同步更新；
4. 按本文件第 5 行约定，修复提交逐条删除对应条目，不累积历史；FIX-01～05 修仿真夹具（反馈文件移出产品树、`candidate-inputs.json` 补 stage）。落地顺序建议：P0（A→B→C）→ P1 → P2/P3；SIM-17 按原接收链的输入、来源、事务与反馈边界分层闭合；SIM-19 的格式改造独立验证兼容，不要求先外移 Hook 或新增作者步骤。
