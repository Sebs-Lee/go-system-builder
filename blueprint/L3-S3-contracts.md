# L3-S3 — 契约（Contracts）

> 层：第三层 ｜ 上游：L2 §S3 + L2 全局规则「单一验证分母」 ｜ 版本 v4.0.1（v4.0.1 设计对抗审查处置：P0×1+P1×7；v4.0.0 联合审查版：契约流水线断点地图+三查+documents 登记落地；v3.1.0 增 §6；机制事实经调查核实，含 file:line）

## 1. 要实现什么

把需求+设计翻译成**分端执行契约**（索引 + FE/BE/SYNC），让前后端对同一接口的理解在锁定的文本上重合——契约是"防止各自想象"的共同事实声明，也是 Builder 的唯一边界。

- 进入时：S2 出口（架构决策 + 模块真相包）。
- 出去时：`GATE-PLANNING-CONTRACTS-COMPLETE` 可满足——存在 locked 契约文档 + `planning_contract` 证据（Contract Planner/Orchestrator，pass）。
- 衡量：**每条验收标准都能沿 REQ→Rule→CASE→Story→PATH→条款 的链走通**，且 SYNC 契约里前端行为与后端行为逐项对得上——两端没有各自想象的空间。

**v4.0.0 身份与最大发现（联合审查确立）**：契约流水线是"两端机械、中段手抄、锁点断源"——S2 侧校验与 gate/迁移链真实连续，但全部语义承载（覆盖/对照/同构/指纹）手抄，且 **TR-003 的锁定从未真正登记契约**（documents[] 的 contract/task 写入在生产代码零先例）。后果：**三套已建成机制在契约上全部空转**——hook 保护（管道真、水源无）、S5 独立性检查（author_agent_id 恒空）、代际保护（仅 req 豁免）。本 stage 整改的第一杠杆：接上水源+补上中段对账。

## 2. 手头有什么（真实机制）

| 机制 | 它能干什么 | 载体 |
|:--|:--|:--|
| CONTRACTS 索引模板 | 头部稳定性元数据（状态/版本/PM/Contractor/锁定依据=runtime@revision+transition+DV 证据）；**UI 设计包输入表**（真相包全集文件+Fingerprint 列）；**需求覆盖矩阵**（REQ→Rule→CASE→Story→PATH→Spec→条款→TASK 一表打穿） | `docs/contracts/CONTRACTS-template.md` |
| BE/FE 契约模板 | §2 范围+**排除**；§3 输出契约表+**需求条款映射**（source_ref/Rule/CASE/条款§n/验收标准）+ Rule→CASE→…→Spec→Evidence 表（BE oracle 同场景包结构：visible/terminal_state/…负例 rejection/expected_state/recovery）；FE 另有场景与测试映射表（CASE id/polarity/visible oracle/PATH/spec）与联调点表（METHOD PATH） | `BE/FE-contract-template.md` |
| SYNC 契约模板 | **两端形状对照的真正载体**：UI 设计包映射四列表（真相文件/字段-错误-状态-权限-副作用/前端行为/后端行为）；wire shape/error/idempotency/state 断言表；原始 HTTP+JSON 样例；错误码表（错误码/HTTP 状态/场景/前端行为）；契约测试用例表（CT-xxx） | `SYNC-contract-template.md` |
| 契约锁定校验（机器） | 门求值只查 documents 中存在 kind=contract∧status=locked（不读契约内容）；磁盘状态字段核对由 **`planning_complete` guard**（挂 TR-002）承担（"planning not complete: contract {path}"）——注意：`verified_versions_current` 本身是证据非空桩（guards.go:122），真核对在 planning_complete（guards.go:326-373） | qualitygate/evaluator.go:203-233；guards.go:326-373 |
| 证据登记 | `runtime evidence add --kind planning_contract`（slot 与 kind 双名设计：slot=planning_contract_record，kind=planning_contract） | catalog.go:345-347；run.go:1259-1300 |
| 规划门推进 | PTR-PLAN-02（contracts→tasks）挂 GATE-PLANNING-CONTRACTS-COMPLETE，PreToolUse 自动迁移；on_guard_failure=warn_and_retry | loop-definition.json:162-181 |
| UI 先决（真实形态） | **Quality Gate not_ready**：写契约而真相包不齐 → 门不满足（missing=document:contract:locked），不前进；~~populateUIPrototypeFact~~ **已删除**（S2 v4.0.2，含调用点） | agent-protocol.md:235 |
| 规则 | api-design.md："APIs are contracts, not implementation details"——API 字段必须记入 SYNC、X-Request-ID、ISO 8601 UTC、金额禁浮点、错误码必定义调用方行为、兼容性矩阵（必填新增=breaking→ADR）；naming.md：一概念一名字、契约文件命名 CONTRACTS/FE/BE/SYNC-{NNN}.md | `docs/rules/api-design.md`、`naming.md` |
| skill | `specification-planning`（S3 主流程：**起草顺序固定 FE→BE→SYNC**，每份链 REQ source_ref+真相包）；`api-contracts`（locked contract owns interface semantics，禁静默扩展）；`http-api-design`/`openapi-swagger`（质量指引） | skills/ |

### 2.1 联合审查：契约流水线断点地图（v4.0.0，2026-08-15 sub-agent 端到端调查）

十二断点按环节（形态 | 关键事实）：

| 环节 | 形态 | 关键事实 |
|:--|:--|:--|
| 引用提取（契约表格→token） | 手抄无校验 | 四模板表头结构稳定可写解析器，但无人写过（semantic 生产代码对 contract 零命中） |
| 条款覆盖（FR→§n→TASK） | 手抄无校验 | gate 只查"存在一条 locked 契约"（evaluator.go:203-233），不读内容 |
| 两端对照（SYNC 四列） | 纯文本承诺（**有意保留**） | 语义对齐是人审（本文件 §3 已否决机器 diff） |
| oracle 同构（契约链↔场景包） | 手抄无校验（词表逐字同名） | model.go 七字段 vs BE:57/FE:63——同名无对账 |
| 指纹锁定（Fingerprint 列/锁定依据行） | 纯文本承诺 | 列手填；"锁定依据"行全库零读者（纯装饰） |
| **documents[] 登记** | **无路径（最大断点）** | 生产代码仅 req 两处写入+recovery；`atomically_lock_execution_batch` 是证据非空占位（actions.go:359-361） |
| hook 保护 | 机械连续但前提断裂 | policy 管道真（engine.go:264-297），契约从不进 LockedArtifacts——水源未接水管干烧 |
| 代际处置 | 纯文本承诺 | amend 后旧代契约退出 hook 保护（loader 代次过滤）、指纹被照磁盘重算（仅 req 豁免） |
| S4 绑定（manifest 手抄） | 手抄无校验 | TASK §2 的 contract 行 id/路径/版本/指纹全手抄 |
| S7 REV 追溯表 | 手抄无校验 | REV:30-34 表 token 无人核 |
| 证据链→gate→迁移 | 机械连续（薄） | planning_contract→gate→PTR-PLAN-02 通 |
| 八件套同步（S2→S3 表面） | **新断点** | CONTRACTS:44/REQ:106/FE:41-44/prototypes README 仍 7 文件清单——S2 v4.0.2 已强制第 8 文件 |

**三套空转机制清单**（登记落地后一次性激活）：hook 保护 / S5 独立性检查（author_agent_id）/ 代际保护。

**S2 成果咬合**：FE 场景映射表与 cases.json 字段名一致（id/polarity/branch_id/oracle.visible/flow_refs）；BE oracle 链与场景七字段词表逐字同名——token 对账的解析基础已备；§F 覆盖矩阵首列 `REQ-{id}/FR-{id}` 与 bridge 的 source_refs 解析格式一致但整表无机器读口（桥只解析 §C 的 AC 表）。

## 3. 选了什么、为什么（含否决）

| 子问题 | 选用 | 否决与否决理由（减法） |
|:--|:--|:--|
| 怎么表达两端一致性 | SYNC 模板四列对照表（同一事实行的前端行为/后端行为并排）+ 原始 HTTP 样例 + 契约测试用例表 | 否决"机器 diff 两端 schema"——形状对照是人的语义对齐，表格+CT 用例已够；机器 diff 只能查结构查不了意图 |
| 怎么承载追溯 | 模板三张映射表（需求条款映射/覆盖矩阵/oracle 链）——**结构靠模板、语义靠 S5 评审** | 如实记录：semantic 校验**不含**契约链接检查（grep 零命中），我 v2 声称的"无孤儿机检"不存在——这是欠账不是现状 |
| 怎么管 UI 先决 | Quality Gate not_ready（真相包不齐→门不满足→不推进） | 已否决 hook deny/warn——项目有明确测试钉死"Hook 永不匹配 ui_contract_before_prototype"；用门比用拦截温和且够用（D3：门是顾问） |
| 起草顺序 | FE→BE→SYNC（skill 钦定） | 顺序即架构立场：先界面事实（用户可见行为），再后端实现，再对齐两端 |
| 三查怎么承载（v4.0.1 修正：查②延后） | 新 `loop-harness contracts check` 子命令，**挂 PTR-PLAN-02 的 guard 链**（收口即门判，非自觉命令——审查发现 D2 接线缺失）+doctor/validate --all：①断链引用（CASE/S/F/PATH/FR token 须存在于模块包与 REQ；**分工边界：S2 桥管 REQ 侧 AC↔CASE，本查只对账契约侧 token 存在性，互不越界**）②~~孤儿条款~~ **延后**——契约模板不存在编号条款清单（§n 仅在映射表内出现=自证循环，审查发现检查对象不存在），待契约结构化条款载体（REQ-040 判断）后再立 ③UI 设计包输入表 Fingerprint 列与磁盘实算比对 ④CONTRACTS 覆盖矩阵条款格存在性（P0 处置的落点）| 否决"并入 scenario validate"——载体不同；否决"S5 人审兜底"；否决"孤儿条款硬做"——对象不存在硬做=恒过的仪式；**解析边界声明（non-goal）**：只提 token 形态不析语义（oracle 单元格/验收标准列是自由文本）；承载 D5+D2+公理三 |
| documents[] 登记怎么落地（v4.0.1 修正 P1：喂食点前移） | **两处登记**：①PTR-PLAN-02（contracts→tasks）新 action `register_locked_contracts`——S3→S4 的出口门 GATE-PLANNING-CONTRACTS-COMPLETE 消费 documents[]，喂食必须发生在门求值前（审查发现：TR-003 在 S5，喂不上 S3 自己的出口门）；扫描 docs/contracts/*.md 状态=locked 者+CONTRACTS 索引清单，写入 documents[]（同代同 id **替换不堆叠**——修复 appendDocument 纯追加导致的同代重锁死锁）；②TR-003 的 action 登记任务（complete 态）——S5 锁的执行批次含任务。状态不对的行为：**fail 不 skip**（skip=静默部分登记→exactSubjects 下游费解缺失，违公理五） | 否决"仅 TR-003 一处登记"——GATE-PLANNING-CONTRACTS-COMPLETE 只认 documents[]，无磁盘回退（evaluator.go:209-233），喂食点必须前移；否决"CLI 副作用写"——迁移 action 是唯一带原子性的自然路径；承载 D2+D6 |
| 锁定依据行（v4.0.1 修正 P1：回填不可行，改为指路） | **删除模板的手填"锁定依据"行**，替换为一行指路："锁定状态与依据见 runtime documents[] 与 journal（.claude/loop-events.jsonl）"——回填是被自己击败的：迁移事务只覆盖 state，事务外文件写要么回滚留脏文件、要么提交后 documents sha 与磁盘失配（审查发现 P1-3）；锁定事实的权威居所本就是 documents[]+journal（D1），文件里再写一份是第二权威 | 否决"TR-003 回填"——时序上必破坏刚钉死的指纹；否决"保留手填装饰"——零读者+会漂移；承载 D1（单一权威） |
| 代际豁免怎么推广（v4.0.0 新增） | RefreshFingerprints 与 reachability 的 superseded 代豁免从"仅 req"推广到全 kind（旧代一律不可变历史，不得被刷新改写）；**配套（v4.0.1 P1 处置）**：loop-state schema 的 documentReference 补 author_agent_id 字段（现 additionalProperties:false 无此字段，写入即被 schema 拒——审查发现地基缺失）；值来源=登记 action 的 request actor；**实效如实降级**：主会话起草场景下"审查者∉作者"恒真（作者=主会话），独立性收益仅在起草被委派给 sub-agent 时为真——激活是真激活（空转→有数据），收益边界如实记录 | 否决"维持 req 特例"——登记落地后契约也有多代，特例口径即漂移源；与 L3-S1 v4.5.1"旧代非 req 不得移动"规则合并为统一口径：**旧代条目=冻结历史（不刷新、不移动、hook 按 loader 代次过滤自然退出保护）** |
| 覆盖矩阵机读口（v4.0.1 修正 P0） | **活矩阵唯一居所=CONTRACTS 索引的需求覆盖矩阵**（契约类文档，agent 可写）；REQ §F 保持骨架并注明"活矩阵在 CONTRACTS-{id}，本表仅锁定时快照"——§F 被 hook 人-only+基线不可变三重锁死，作机检输入不可行（审查发现）。`contracts check` 校验 CONTRACTS 矩阵行的 FE/BE/SYNC 条款格（`{id} §{n}`）须存在于对应契约文件 | 否决"§F 直填"——REQ 是锁定基线，agent 填写被 hook 拦/人-only 禁/指纹漂移崩下游（P0 处置）；否决"留 S5 人审"——汇总视图错抄毒性最大；承载 D6 |
| 锁定语义 | 模板 status 字段 + S5 原子锁把它升格为机器事实（documents 登记）→ 此后写入受锁定产物拦截 | 否决"起草期就锁"——S3 产物允许迭代；锁的权威在 S5，S3 只声明 |
| 契约变更 | 兼容性矩阵：可选字段新增=兼容（更新 SYNC+CT）；必填新增/删除=breaking（ADR+版本升级）+ change-control | 否决"一律重锁"——兼容变更走轻路径，breaking 才重——减法 |

## 4. 怎么编排（时间线讲完一件事）

1. **UI 先决自检**：ui_impact=changed 时先确认真相包就绪——否则写契约也白写（门不会满足，missing 会指回来）。
2. **FE 起草**：按 FE 模板——§3 逐项挂原型包文件（current/locked 状态）、场景与测试映射表把 CASE/polarity/oracle/PATH 抄进契约、联调点列 METHOD+PATH。
3. **BE 起草**：范围+排除、需求条款映射逐行 source_ref、oracle 链与场景包同构（BE 的验证断言直接复用场景语言）。
4. **SYNC 对齐**：四列对照表逐行写"同一事实的两端行为"；错误码表定义到前端行为级；原始 HTTP 样例+CT 用例落纸。
5. **索引收口**：CONTRACTS 索引汇总清单、UI 输入表（含指纹列）、覆盖矩阵——验收标准→条款的最后一遍人肉对账。
6. **登记推进**：`runtime evidence add`（planning_contract，pass）→ 下一次 PreToolUse 评估门（locked 契约文档存在+磁盘状态一致+证据在册）→ PTR-PLAN-02 自动迁移进 tasks。
6.5. **机检收口（v4.0.1，门判非自觉）**：PTR-PLAN-02 的 guard 链跑 `contracts check`——断链引用/Fingerprint 实算/CONTRACTS 矩阵条款格全过才放行（doctor/validate --all 同源复跑）；不绿 → 修复对应表格行后重跑。
7. **返工回路**：S5 审查发现契约矛盾 → 主会话修复 → 仅受影响职责用新指纹重跑（agent-protocol.md:266）。
8. **登记时间线（v4.0.1 前移）**：PTR-PLAN-02 action 登记 locked 契约入 documents[]（同代同 id 替换不堆叠）——出口门自此有粮；TR-003 action 登记 complete 任务——执行批次闭环；hook 保护自 PTR-PLAN-02 起对契约生效。

## 5. 期望效果

走完 S3：

- **两端无各自想象**：SYNC 对照表让每个字段/错误/状态/副作用的前后端行为并排可见；oracle 链让验证语言从场景包一路贯到契约；
- **结构性防住**：接口语义漂移（locked contract owns semantics+禁静默扩展）、breaking 变更溜过（兼容性矩阵+ADR）、命名漂移（一概念一名字）、UI 契约先于原型（门 not_ready）；
- **交给 S4**：契约条款（§n 编号）+ oracle 链——TASK 收尾契约判据的素材；**交给 S5**：契约集——文档验证的对象与 `contract_set_record` 证据的标的。

**缺口处置台账**（v4.0.0 重写为处置视角）：①契约链接/孤儿/指纹无机器校验 → **本轮落地** `contracts check` 三查+§F 轻校验；②v2 虚构机检门 → 已在 v3.0.0 修正为如实；③`atomically_lock_execution_batch` 占位 → **本轮落地**真登记（契约+任务入 documents[]）；④契约无结构化数据形态 → **保持**（token 对账在 markdown 表格上做即可，结构化 schema 归 REQ-040 判断）；⑤八件套模板不同步（S2 v4.0.2 新增断点）→ **本轮修**（CONTRACTS/REQ/FE/README 四处 7→8 文件）；⑥锁定依据行零读者 → **本轮机写**（TR-003 回填）；⑦代际豁免 req 特例 → **本轮推广全 kind**。**如实记录（仍然成立）**：SYNC 四列两端对照保持人审（有意设计）；"契约集证据"kind 与普通文档审查同 kind 的标的绑定弱化——登记落地后标的物变为 documents[] 本身，证据 kind 问题降级为记账细节。

## 6. 注意力预算与渐进披露

总评：**全系统文档承载最重的 stage**——机械可判的追溯完整性全靠自觉填表 + S5 人审，左移空间最大。判定尺见 L3-README「注意力分配原则」。

### 6.1 当前错配（什么不对、为什么不对）

| # | 错配 | 为什么不对（L1 根据） |
|:--|:--|:--|
| 1 | 追溯三张表全靠模板自觉填 + S5 人审，semantic 对契约零覆盖（grep 零命中） | 引用存在性/指纹新鲜度是机械可判项（载体三问 1）；孤儿条款/断链 CASE/陈旧指纹现在要等到 S5 人力发现——"纸上修复成本≈0"的红利被错误的发现点放弃（发现越晚越贵） |
| 2 | 索引的 Fingerprint 列手填 | 机械可算（`runtime fingerprint` 命令已存在）却让人填=制造可错点且无人核验——便宜确定性反被弃用（D5 反向） |
| 3 | SYNC 四列对照 + CT 用例 + HTTP 样例 + 错误码表对纯后端 REQ 大半 N/A | "必填但无关"的 N/A 扫填训练 agent 忽略字段——D4 逼问失效为格式填充（Eroding Goals 的入口） |
| 4 | api-design 规则 6KB 全文靠读 | 字段级规范（金额禁浮点/X-Request-ID/ISO 8601）的正确居所是模板行内注释——写到该字段时才见；进阶段就通读=为尚未写到的字段预付注意力 |

### 6.2 阅读预算（谁在何时读什么）

| 角色 | 进入时必读 | 触发式加载 | 永不需要读（机制承载） |
|:--|:--|:--|:--|
| 起草者 | FE/BE/SYNC 模板（起草顺序 FE→BE→SYNC 由 skill 一句话承载，不需读规则全文） | 写到具体字段时读该字段的行内规范/api-design 对应节 | "追溯必须完整"的叙述——链接校验应点名缺口（整改后）；UI 先决——gate not_ready 会指回来 |
| S5 审查者 | **语义对齐两处**：SYNC 四列（两端行为是否同一事实）+ **oracle 翻译抽查**（FE/BE 映射表的 oracle 翻译是否忠实于场景包原字段——手抄即翻译，这是 S3 最深的思考动作，不能无消费者） | — | 机械核对（引用存在/指纹）——交机器；**分工边界**：S2 桥管 REQ 侧 AC↔CASE，contracts check 只对账契约侧 token |

### 6.3 整改方向（v4.0.1 修订全景，机制 6+保持 3）

**机制项（6）**：
1. **`contracts check`**（挂 PTR-PLAN-02 guard 链+doctor）：断链引用/Fingerprint 实算/CONTRACTS 矩阵条款格——孤儿条款延后（对象不存在）；只提 token 不析语义（边界声明）；
2. **PTR-003→PTR-PLAN-02 双点登记**：PTR-PLAN-02 action `register_locked_contracts`（locked 契约，同代同 id 替换）；TR-003 action 登记 complete 任务；fail 不 skip；
3. ~~锁定依据行机写~~ → **删手填行换指路行**（权威居所=documents[]+journal）；
4. **代际豁免推广全 kind** + schema documentReference 补 author_agent_id（值=登记 actor；主会话场景实效边界如实记录）；
5. **八件套同步**：CONTRACTS/REQ/FE 模板+prototypes README 的 7→8 文件；
6. **appendDocument 替换语义**：同代同 id 条目替换不堆叠（修复同代重锁死锁）。

**保持项（3）**：SYNC 四列人审（有意设计）；FE→BE→SYNC 起草顺序；oracle 词表同构（token 对账覆盖）。

**E2E 计划（v4.0.1）**：契约全链——起草四契约（含手抄链）→ `contracts check` 绿 → 断链注入红 → 修复 → PTR-PLAN-02 登记契约（documents[] 含 author_agent_id）+门过 → hook 拦截契约写入 → 同代返工重锁（替换不堆叠）→ TR-003 登记任务 → amend 后代际处置（旧代冻结）。

## 变更记录

| 日期 | 版本 | 变更 |
|:--|:--|:--|
| 2026-08-14 | v1/v2 | 前两版（被判空洞→叙事不清；v2 的契约机检门系虚构） |
| 2026-08-14 | v3.0.0 | 叙事版；机制事实经 sub-agent 调查核实（模板三张映射表/SYNC 四列对照/UI 门=not_ready/追溯无机检等如实入档） | owner 复核 |
| 2026-08-15 | v4.0.1 | **设计对抗审查处置（P0×1+P1×7+P2×7）**：P0——§F 轻校验输入被 hook 人-only/基线不可变三重锁死，**活矩阵唯一居所改判为 CONTRACTS 索引**（REQ §F 留骨架注指路）；P1——documents 喂食点前移至 PTR-003-02（出口门无粮问题）+appendDocument 同代同 id 替换语义（重锁死锁）；锁定依据行回填改删除+指路（事务外文件写破坏指纹）；孤儿条款延后（检查对象不存在——模板无编号条款清单，§n 仅在映射表内=自证）；oracle 翻译抽查补进 S5 审查预算（S3 最深思考动作的消费者）；author_agent_id 补 schema 字段+主会话场景实效边界如实降级；三查从自愿命令改挂 PTR-PLAN-02 guard 链（D2 接线）；P2——§2 两处失实修正（verified_versions_current 张冠李戴/populateUIPrototypeFact 过去时）、步骤重排、oracle"token 对账覆盖"自相矛盾修正、bridge/check 分工边界声明、计数对齐 | 设计对抗审查（sub-agent）：方案闭合 half 返工 |
| 2026-08-15 | v4.0.0 | **联合审查版**（S3+S4+S5+S7 联动，sub-agent 契约流水线端到端调查）：§1 增"两端机械、中段手抄、锁点断源"定位与三套空转机制清单；§2.1 新增十二断点地图（documents[] 登记无路径为最大断点——生产代码仅 req 写入，TR-003 占位；hook 保护/S5 独立性/代际保护三套建成机制空转；S2 八件套未同步为新断点）；§3 新增五条选用（contracts check 三查/TR-003 真登记/锁定依据机写/代际豁免推广/§F 轻校验，各含否决）；§4 补机检与登记时间线；§5 改处置台账（七项全处置）；§6.3 升 8 项全景（机制 5+保持 3） | owner 批准六决策点按建议执行 |
| 2026-08-15 | v3.1.0 | 新增 §6 注意力预算与渐进披露（错配诊断/阅读预算/整改方向），判定尺引 L3-README | owner 指示：渐进披露、机制承载规范、削减平白叙述 |
