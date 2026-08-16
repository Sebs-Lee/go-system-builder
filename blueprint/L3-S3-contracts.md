# L3-S3 — 契约（Contracts）

> 层：第三层 ｜ 上游：L2 §S3 + L2 全局规则「单一验证分母」 ｜ 版本 v4.0.0（v4.0.0 联合审查版：契约流水线断点地图+三查+documents 登记落地；v3.1.0 增 §6；机制事实经调查核实，含 file:line）

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
| 契约锁定校验（机器） | 门求值要求 documents 中存在 kind=contract∧status=locked；guard `verified_versions_current` 逐份核对磁盘 markdown 状态字段一致（"planning not complete: contract {path}"） | qualitygate/evaluator.go:203-233；guards.go:334-381 |
| 证据登记 | `runtime evidence add --kind planning_contract`（slot 与 kind 双名设计：slot=planning_contract_record，kind=planning_contract） | catalog.go:345-347；run.go:1259-1300 |
| 规划门推进 | PTR-PLAN-02（contracts→tasks）挂 GATE-PLANNING-CONTRACTS-COMPLETE，PreToolUse 自动迁移；on_guard_failure=warn_and_retry | loop-definition.json:162-181 |
| UI 先决（真实形态） | **Quality Gate not_ready**：写契约而真相包不齐 → 门不满足（missing=document:contract:locked），不前进；**不是 hook 拦截**——populateUIPrototypeFact 是无消费者的残迹，测试明确"Hook 永不匹配该 fact" | agent-protocol.md:235；run_test.go:1137-1148 |
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
| 三查怎么承载（v4.0.0 新增） | 新 `loop-harness contracts check` 子命令，挂 doctor/validate --all：解析四模板的结构稳定表格，查①断链引用（CASE/S/F/PATH/FR token 须存在于模块包与 REQ——对账 cases.json/stories/flows/REQ 表）②孤儿条款（条款 §n 须被需求条款映射表或覆盖矩阵引用）③UI 设计包输入表 Fingerprint 列与磁盘实算比对 | 否决"并入 scenario validate"——契约是文档不是场景包，载体不同；否决"S5 人审兜底"——机械可判的引用存在性留在人审是最贵的核对；承载 D5（便宜确定性大量做）+公理三（三查各有消费者：S4 绑定/S5 审查/S7 追溯） |
| documents[] 登记怎么落地（v4.0.0 新增） | **TR-003 的 `atomically_lock_execution_batch` 从占位变真**：锁定时刻扫描 locked 契约+complete 任务的磁盘文件，写入 documents[]（id/kind/path/version/sha256/generation/**author_agent_id**）——锁定即登记，原子性由迁移事务承载 | 否决"S5 证据登记时 CLI 副作用写"——锁的原子性是 S5 出口定义的一半；否决"只登记契约"——S5 锁的是执行批次（契约+任务）；承载 D2（锁定是自然路径上的必经点）+D6（三套空转机制一次性激活） |
| 锁定依据行谁写（v4.0.0 新增） | TR-003 登记时回填契约头部的锁定依据行（runtime@revision/transition/DV 证据）——从手填装饰变机写事实 | 否决"删掉该行"——它的消费者是审计者（回溯锁定来源），机写比删除多留住一条审计线索；承载 D1（输出与 journal 可互证） |
| 代际豁免怎么推广（v4.0.0 新增） | RefreshFingerprints 与 reachability 的 superseded 代豁免从"仅 req"推广到全 kind（旧代一律不可变历史，不得被刷新改写） | 否决"维持 req 特例"——登记落地后契约也有多代，特例口径即漂移源；与 L3-S1 v4.5.1"旧代非 req 不得移动"规则合并为统一口径：**旧代条目=冻结历史（不刷新、不移动、hook 按 loader 代次过滤自然退出保护）** |
| §F 覆盖矩阵机读口（v4.0.0 新增） | `contracts check` 加轻校验：§F 行的 FE/BE/SYNC 条款格（`{id} §{n}`）须存在于对应契约文件 | 否决"留 S5 人审"——§F 是全链汇总视图，错抄毒性最大（一处错格污染 S4/S7 的追溯起点）；承载 D6（分母链的汇总视图也要可机核） |
| 锁定语义 | 模板 status 字段 + S5 原子锁把它升格为机器事实（documents 登记）→ 此后写入受锁定产物拦截 | 否决"起草期就锁"——S3 产物允许迭代；锁的权威在 S5，S3 只声明 |
| 契约变更 | 兼容性矩阵：可选字段新增=兼容（更新 SYNC+CT）；必填新增/删除=breaking（ADR+版本升级）+ change-control | 否决"一律重锁"——兼容变更走轻路径，breaking 才重——减法 |

## 4. 怎么编排（时间线讲完一件事）

1. **UI 先决自检**：ui_impact=changed 时先确认真相包就绪——否则写契约也白写（门不会满足，missing 会指回来）。
2. **FE 起草**：按 FE 模板——§3 逐项挂原型包文件（current/locked 状态）、场景与测试映射表把 CASE/polarity/oracle/PATH 抄进契约、联调点列 METHOD+PATH。
3. **BE 起草**：范围+排除、需求条款映射逐行 source_ref、oracle 链与场景包同构（BE 的验证断言直接复用场景语言）。
4. **SYNC 对齐**：四列对照表逐行写"同一事实的两端行为"；错误码表定义到前端行为级；原始 HTTP 样例+CT 用例落纸。
5. **索引收口**：CONTRACTS 索引汇总清单、UI 输入表（含指纹列）、覆盖矩阵——验收标准→条款的最后一遍人肉对账。
6. **登记推进**：`runtime evidence add`（planning_contract，pass）→ 下一次 PreToolUse 评估门（locked 契约文档存在+磁盘状态一致+证据在册）→ PTR-PLAN-02 自动迁移进 tasks。
7. **机检三查（收口前，v4.0.0）**：`contracts check` 绿——断链引用/孤儿条款/指纹实算/§F 条款格全过；不绿 → 修复对应表格行后重跑（人审只保留语义对齐：SYNC 四列的两端行为是否同一事实）。
8. **返工回路**：S5 审查发现契约矛盾 → 主会话修复 → 仅受影响职责用新指纹重跑（agent-protocol.md:266）。
9. **锁定即登记（S5 侧，v4.0.0）**：TR-003 提交时真登记——契约+任务入 documents[]（含 author_agent_id），契约头部锁定依据行机写回填；hook 保护自此对契约生效。

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
| S5 审查者 | 只审**语义对齐**（SYNC 四列两端行为是否同一事实） | — | 机械核对（引用存在/指纹）——交机器 |

### 6.3 整改方向（v4.0.0 联合审查全景，8 项）

**机制项（5）**：
1. **`contracts check` 三查+§F 轻校验**：断链引用（token 对账模块包/REQ）/孤儿条款/Fingerprint 实算/§F 条款格存在性——挂 doctor/validate --all；
2. **TR-003 真登记**：`atomically_lock_execution_batch` 扫描 locked 契约+complete 任务入 documents[]（含 author_agent_id、generation）；
3. **锁定依据行机写**：TR-003 回填契约头部（runtime@revision/transition/DV 证据）；
4. **代际豁免推广**：RefreshFingerprints/reachability 的 superseded 豁免全 kind；
5. **八件套同步**：CONTRACTS/REQ/FE 模板+prototypes README 的 7→8 文件。

**保持项（3）**：SYNC 四列人审（有意设计）；FE→BE→SYNC 起草顺序；oracle 词表同构（token 对账覆盖）。

**E2E 计划**：契约全链——起草四契约（含手抄链）→ `contracts check` 绿 → 断链注入红 → 修复 → TR-003 登记（documents[] 含 contract+task、锁定依据回填）→ hook 拦截契约写入 → amend 后代际处置（旧代冻结）。

## 变更记录

| 日期 | 版本 | 变更 |
|:--|:--|:--|
| 2026-08-14 | v1/v2 | 前两版（被判空洞→叙事不清；v2 的契约机检门系虚构） |
| 2026-08-14 | v3.0.0 | 叙事版；机制事实经 sub-agent 调查核实（模板三张映射表/SYNC 四列对照/UI 门=not_ready/追溯无机检等如实入档） | owner 复核 |
| 2026-08-15 | v4.0.0 | **联合审查版**（S3+S4+S5+S7 联动，sub-agent 契约流水线端到端调查）：§1 增"两端机械、中段手抄、锁点断源"定位与三套空转机制清单；§2.1 新增十二断点地图（documents[] 登记无路径为最大断点——生产代码仅 req 写入，TR-003 占位；hook 保护/S5 独立性/代际保护三套建成机制空转；S2 八件套未同步为新断点）；§3 新增五条选用（contracts check 三查/TR-003 真登记/锁定依据机写/代际豁免推广/§F 轻校验，各含否决）；§4 补机检与登记时间线；§5 改处置台账（七项全处置）；§6.3 升 8 项全景（机制 5+保持 3） | owner 批准六决策点按建议执行 |
| 2026-08-15 | v3.1.0 | 新增 §6 注意力预算与渐进披露（错配诊断/阅读预算/整改方向），判定尺引 L3-README | owner 指示：渐进披露、机制承载规范、削减平白叙述 |
