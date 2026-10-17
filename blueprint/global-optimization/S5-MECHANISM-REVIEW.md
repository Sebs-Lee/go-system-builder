# S5 机制审查与返工（2026-10-16）

当前结论：**S5 范围返工与最终装包验收通过**。after-r32 的真实双评审、原 TR-003、
新旧答卷六次零写入重放，以及现场 `tasks check / validate --all / doctor` 均通过。
原游标停在 S6 入口，本轮没有开展 S6 工作。

## 范围与最小流程

本轮接续 S4 返工后的 S5，先固定 [L3-S5](../l3-stages/L3-S5-document-verification.md)
的最小契约，再修改原生命周期、原门及说明，最后装包验证。工作状态沿用
`pending / active / blocked / done / cancelled`；专业结论、阶段游标和已有历史执行
字段保持各自用途。验收截止 S6 入口，不开展 S6 工作。

1. Main 从最新已接收 S4 批次组织两个独立责任，以当前 S5 entry 的新身份和新 manifest
   路径准备 reviewer；原 team-planning/register/prepare 发出准确 view/answer 路径。
2. Reviewer 读取已投递的专业来源和原任务分析/结构检查，回答现有
   `conclusion / findings`，保存一次。元数据、检查、证据和接续由原系统完成。
3. 第一位完成后等待另一责任；当前两位独立 PASS 才可由原 TR-003 推进。
   finding 仍走原 TR-004，REQ 方向问题仍保留原人闸。
4. 原 TR-004 在同一事务退出旧批次；产物、原答卷和回执保留。新 S5 entry 从最新 S4
   重新派发，不刷新旧 manifest 指纹、不复用旧 PASS、不要求作者复制来源哈希。
5. 已接收旧答卷重放返回原回执；当前缺项给准确路径、下一动作和责任人。

## 已确认的缺口及返工

| 缺口 | 证据与影响 | 最小修改 |
| --- | --- | --- |
| 旧批次未退出 | TR-004 仅使引用的 fix evidence 失效，旧 team/agent/执行仍在当前集合；旧 manifest 与新 BE 指纹冲突，导致 S4 出口后的 validate/doctor 失败。 | 原 TR-004 一并退出该 gen/entry 的 reviewer、review evidence 与 document_review 执行；执行放入已有 history，保留实际文件。 |
| 单份 PASS 可直达 S6 | r122 隔离探针经原接收仅交一份 PASS，直接原 TR-003 仍成功，跳过第二独立责任。 | 原 TR-003 事务校验当前完整独立双 PASS，包括原接收、修订 tip 和冻结专业来源；原 gate 同步过滤旧 entry 证据。 |
| 全局检查混淆历史与当前 | 已发行清单固定旧来源；旧返工清单的历史指纹不应被拿来匹配当前文件，也不能随意改写。 | 历史身份须由原 packet/owner/team/assignment/entry 及原 manifest SHA 证明；历史仍检查 schema、Runtime、路径和原清单字节。当前及无证明的旧流程继续严格检查，最新 S5 清单在 S6 仍严格。 |
| 执行退出后原回执无法恢复 | 原 Worker 来源仅查询当前执行 map，转入 history 会丢失原来源。 | 复用原执行查询读取精确 packet 对应的已有 history；仅已接收答卷允许从 retired 来源恢复原回执，未接收不得借此提交新判断。 |
| 重复机检与准备提示不足 | reviewer 被要求自己运行结构检查；S5 入场缺少当前清单候选路径。 | 在原冻结 task-analysis 内带入原检查结果；现有旧产物不重写。入场给原 Skill 和 `review-e<entry>/manifest.json`，由原准备步骤发行实际问卷。 |

## 验证记录

评估根：`/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z`。

- `reports/s5-audit-r122-source.json`：取自原 r121 的隔离基线，仅加入明确的单 PASS
  边界探针；不是把在改实现冒充基线。
- `reports/s5-audit-r122-direct-pair.json`：exit 1，21.606 秒；原 TR-003 跳过第二人，
  失败记录保留。
- `reports/s5-rework-r123-source.json`：合并返工的完整源码清单，含未跟踪生产文件。
  定向集成回归 exit 0，213.692 秒；当时尚未完成装包与当前 entry 的真实独立双评审。

后续验收须分别列出：正常双 PASS、finding 返工、原人闸、当前与历史来源校验、
已接收 Worker 重放、原 pending 恢复；机械前置不计真实新专业判断。
此前 after-r27 的双 PASS 及 after-r29 的 S4 验收保留为历史证据，不替代本轮新候选。

- `reports/s5-rework-r124-original-chain.json`：普通三结论、并发、具名修订及中断恢复通过，
  但退出后的 PostToolUse 进度只查当前 execution，出现目录不可用误报，整轮 exit 1。
  原断言与失败日志保留；已接收回执正常并不能代替进度入口正常。
- `reports/s5-rework-r125-retirement-progress.json`：修复原进度读取已有历史执行后，
  退出后本人回执/进度、尚未答卷 peer 的取消与拒绝新判断、旧 entry gate 选择均通过，
  exit 0，66.976 秒。未增加作者动作或工作状态。


### 装包发现的补充缺口

- after-r30 全新安装已由原消费者重放前置进入 S5/e8，两个真实 muse reviewer 各
  Read 后 Write 一次，双 PASS 后原 `evt-tr-003-r52` 进入 S6。规格 23 Read / 1 Write，
  任务 15 Read / 1 Write，均无 Bash；但全局 validate/doctor 尚未通过，不能记最终验收。
- 取消一个从未 activation 的 reviewer，不代表它曾生成真实 activation message。
  `activation_revision=null` 的 shutdown/stopped/cancelled 被旧 fallback 误当成已激活，
  导致把原预置能力文件按真实消息 schema 校验。已去掉该错误推断；真实 activation
  revision 存在时仍严格校验源消息，两版 Runtime 的回归覆盖此边界。r127 的
  fileview/workpacket/submission/semantic 四包整包回归通过（40.597 秒、exit 0）。
- r127 完整 S5 定向链通过（207.590 秒、exit 0），但真实目录重放还暴露原
  `ResolveControl` 只识别 ready 执行；指定 Main root 的旧测试漏掉安装 Hook 的 Worker
  root 入口。r30 两次旧答卷重放均在入口报 stale pointer，原文件未写，失败保留。
  现补原目录→原 controller 的精确历史 S5 定位，并将端到端重放改为实际 Worker root；
  原来源/原接收继续拒绝退休目录提交新判断，其他 retired Builder 规则保持原验收。
- CLI receipt 和 Hook receipt 的区别是后者在 `next_action` 增加准确问卷路径。
  原接收键、证据、完成事实必须一致；验收分别核对原回执事实与准确路径提示，
  不把提示增补误判为新接收。r30 原失败报告不覆盖。

最终候选已另行全新装包，结果如下。r30/r31 现场二进制和失败证据保持原样。


## 最终候选与验收

| 证据 | 结果 |
| --- | --- |
| `reports/s5-rework-r128-source.json`、`after-r32-source-inventory.json`、`r32-s5-source-match.json` | 1636 项完整源码 SHA 清单一致，包含未跟踪生产文件；Linux amd64 评估包构建成功，未做五平台发布。 |
| `reports/s5-rework-r128-worker-root.json` | 9 个定向顶层用例通过，110.433 秒、exit 0；真实 Worker root 的历史回执、取消 peer、原目录来源校验及普通 retired Builder 边界。 |
| `reports/s5-rework-r128-final-chain.json` | 最终同源 48 项定向顶层用例通过，215.604 秒、exit 0；包括相邻接收边界、三种结论、当前双 PASS、返工、并发、修订、中断恢复和原人闸。 |
| `reports/s5-rework-r128-four-packages.json` | fileview/workpacket/submission/semantic 四包完整回归 102 项通过，8.059 秒、exit 0；与定向用例有重叠，不相加，不称全仓回归。 |
| `reports/r32-s5-rework-setup.json` | 全新安装；历史 finding 与已完成 S2/S3/S4 专业答案仅机械回放 answer 对象，由原消费者/原 TR-004 进入新 S5/e8。当前双评审未预填；r25/r28/r29 原场景字节保持不变。 |
| `reports/r32-s5-preflight-history-replay.json` | 从实际旧 Worker root 调用原安装 Hook，准确原回执恢复成功；Runtime、journal、answer 字节不变。 |
| `reports/r32-s5-e8-workers-live.json`、两份原生模型日志 | 两个独立 muse session，规格 163.700 秒、任务 132.884 秒，均 exit 0；分别真实 PASS，无 findings。原 `evt-tr-003-r54` 唯一推进到 building/e9（S6 入口）。 |
| `reports/r32-s5-native-tool-audit.json` | 规格 25 Read / 2 Bash，任务 24 Read / 2 Bash；各有一次脚本保存原问卷，无额外 PLAN、信封、报告、issue/consume/recover。另一次 Bash 是读卷/来源核对，不是必交步骤。此前 r30 的 Write 保存路径也有真实通过证据。 |
| `reports/r32-s5-rework-acceptance.json` | 当前两份原接收、不同 producer、19 项最新 S4/BE/TASK 源闭包和冻结 task-analysis 一致。结构结果为 2 TASK、9/9 clauses、0 problems。旧 e4 team retired、两项责任 cancelled、两执行保留 history；旧清单原 SHA 和文件保留、旧 finding evidence invalid。 |
| 同一验收报告的六次重放及 installed checks | 当前两份与旧 e4 已接收答卷各重放两次，原接收事实/准确路径提示一致，无虚假 system_followup，Runtime/journal/答卷全部零写入。`tasks check`、`validate --all`、`doctor` 均 exit 0。 |

验收报告的 `scoped_passed=true` 表示本轮 S5 范围通过；其 `whole_project_passed=true`
仅表示该安装现场上述三条命令均通过，不代表 S0–S11 全阶段优化或发布完成。
REQ 变更原人闸由定向回归验证，真实双 PASS 场景没有替人作需求决定。
旧清单导致的 S4 出口全局校验失败在本轮新场景已关闭；旧失败日志未改写。
最终结果写入后仅更新本审查、全阶段台账和 S4 后续说明，生产源码与候选保持一致。

下一阶段由用户编排 S6 审查；本轮没有派发 Builder 或产生 S6 专业交付。


## 后续优化：直接解释问卷选择（2026-10-16）

按用户对“是否仍需猜测字段”的追问，先固定 L3-S5，再同步实际问卷、Role、Skill
和协议。作者流程统一为：**审查明确列出的事项，选择结论；有问题才补充清单，保存一次。**

- `pass`：本责任未发现问题，保持预填 `findings: []`，不用另写“无问题”报告。
- `fix_required`：REQ 含义明确，修正设计、合同或 TASK 即可满足。
- `req_change_required`：REQ 本身存在歧义、冲突或缺失，须由人类明确或变更；
  选择该项不表示已获变更授权。
- 有问题才填 `path`（所给被审文件，可附锚点）、`expected`（应有内容）、
  `observed`（实际缺口）；不用抄身份、指纹或机检汇总。

格式、枚举、原门与全局五态均不变。新发行用新题干；本轮优化前的冻结模板原字节
随包保留，原 packet/具名修订按其绑定的模板 SHA 重建，不改写已发行问卷、答案或回执。
仅允许已知的这两版模板；未知模板不静默套用新题干。

| 证据 | 结果及范围 |
| --- | --- |
| `reports/s5-guidance-r130-source.json`、`after-r33-source-inventory.json` | 1637 项完整源码清单相同；after-r33 Linux amd64 评估包构建成功。 |
| `reports/s5-guidance-r130-chain.json` | 30 项定向顶层用例通过，186.474 秒、exit 0；含新题干、旧模板重建、三结论原路由、当前独立审查、返工、并发与具名修订。 |
| `reports/s5-guidance-r130-packet-full.json` | workpacket 整包 37 项通过，1.764 秒、exit 0；与定向覆盖有重叠，不相加。 |
| `reports/r33-s5-guidance-interpretation-acceptance.json`、原 stream/process 日志 | 一个真实 muse session 读取实际新题干，六个独立情景（两个职责 × 通过/修文档/需求待人决策）全部选择正确；通过保留空清单，非通过仅交三项定位字段；1 轮、21.016 秒、0 工具调用。这是题干理解仿真，不是真实项目双审签字或完整阶段链。 |
| `reports/r33-s5-prior-runtime-readonly.json` | 新二进制只读检查原 r32 真实现场：validate/doctor 及三份已接收答卷原回执均 exit 0；四份已发行 S5 packet 的模板被保留。Runtime、journal、原安装二进制、view、Main 草稿和 Worker 答卷哈希全部不变。没有覆盖安装或替换原场景二进制。 |

本补充没有重跑 after-r33 的完整项目双评审链；最近的完整双审/六次重放验收仍是
上节 after-r32，不能将其冒充新包的新一轮完整链。新题干的三种路由由本地原消费者
回归覆盖，表达是否清楚由上述六情景真实模型仿真覆盖。仍停在 S6 入口，未开展 S6。
