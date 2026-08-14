# L3-S3 — 契约（Contracts）

> 层：第三层 ｜ 上游：L2 §S3 ｜ 版本 v3.0.0（叙事版；机制事实经调查核实，含 file:line）

## 1. 要实现什么

把需求+设计翻译成**分端执行契约**（索引 + FE/BE/SYNC），让前后端对同一接口的理解在锁定的文本上重合——契约是"防止各自想象"的共同事实声明，也是 Builder 的唯一边界。

- 进入时：S2 出口（架构决策 + 模块真相包）。
- 出去时：`GATE-PLANNING-CONTRACTS-COMPLETE` 可满足——存在 locked 契约文档 + `planning_contract` 证据（Contract Planner/Orchestrator，pass）。
- 衡量：**每条验收标准都能沿 REQ→Rule→CASE→Story→PATH→条款 的链走通**，且 SYNC 契约里前端行为与后端行为逐项对得上——两端没有各自想象的空间。

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

## 3. 选了什么、为什么（含否决）

| 子问题 | 选用 | 否决与否决理由（减法） |
|:--|:--|:--|
| 怎么表达两端一致性 | SYNC 模板四列对照表（同一事实行的前端行为/后端行为并排）+ 原始 HTTP 样例 + 契约测试用例表 | 否决"机器 diff 两端 schema"——形状对照是人的语义对齐，表格+CT 用例已够；机器 diff 只能查结构查不了意图 |
| 怎么承载追溯 | 模板三张映射表（需求条款映射/覆盖矩阵/oracle 链）——**结构靠模板、语义靠 S5 评审** | 如实记录：semantic 校验**不含**契约链接检查（grep 零命中），我 v2 声称的"无孤儿机检"不存在——这是欠账不是现状 |
| 怎么管 UI 先决 | Quality Gate not_ready（真相包不齐→门不满足→不推进） | 已否决 hook deny/warn——项目有明确测试钉死"Hook 永不匹配 ui_contract_before_prototype"；用门比用拦截温和且够用（D3：门是顾问） |
| 起草顺序 | FE→BE→SYNC（skill 钦定） | 顺序即架构立场：先界面事实（用户可见行为），再后端实现，再对齐两端 |
| 锁定语义 | 模板 status 字段 + S5 原子锁把它升格为机器事实（documents 登记）→ 此后写入受锁定产物拦截 | 否决"起草期就锁"——S3 产物允许迭代；锁的权威在 S5，S3 只声明 |
| 契约变更 | 兼容性矩阵：可选字段新增=兼容（更新 SYNC+CT）；必填新增/删除=breaking（ADR+版本升级）+ change-control | 否决"一律重锁"——兼容变更走轻路径，breaking 才重——减法 |

## 4. 怎么编排（时间线讲完一件事）

1. **UI 先决自检**：ui_impact=changed 时先确认真相包就绪——否则写契约也白写（门不会满足，missing 会指回来）。
2. **FE 起草**：按 FE 模板——§3 逐项挂原型包文件（current/locked 状态）、场景与测试映射表把 CASE/polarity/oracle/PATH 抄进契约、联调点列 METHOD+PATH。
3. **BE 起草**：范围+排除、需求条款映射逐行 source_ref、oracle 链与场景包同构（BE 的验证断言直接复用场景语言）。
4. **SYNC 对齐**：四列对照表逐行写"同一事实的两端行为"；错误码表定义到前端行为级；原始 HTTP 样例+CT 用例落纸。
5. **索引收口**：CONTRACTS 索引汇总清单、UI 输入表（含指纹列）、覆盖矩阵——验收标准→条款的最后一遍人肉对账。
6. **登记推进**：`runtime evidence add`（planning_contract，pass）→ 下一次 PreToolUse 评估门（locked 契约文档存在+磁盘状态一致+证据在册）→ PTR-PLAN-02 自动迁移进 tasks。
7. **返工回路**：S5 审查发现契约矛盾 → 主会话修复 → 仅受影响职责用新指纹重跑（agent-protocol.md:266）。

## 5. 期望效果

走完 S3：

- **两端无各自想象**：SYNC 对照表让每个字段/错误/状态/副作用的前后端行为并排可见；oracle 链让验证语言从场景包一路贯到契约；
- **结构性防住**：接口语义漂移（locked contract owns semantics+禁静默扩展）、breaking 变更溜过（兼容性矩阵+ADR）、命名漂移（一概念一名字）、UI 契约先于原型（门 not_ready）；
- **交给 S4**：契约条款（§n 编号）+ oracle 链——TASK 收尾契约判据的素材；**交给 S5**：契约集——文档验证的对象与 `contract_set_record` 证据的标的。

**如实记录的已知缺口**（供第四层修复清单）：①契约链接/孤儿/指纹新鲜度**无机器校验**（semantic 零覆盖；索引模板的 Fingerprint 列是手填的）——追溯完备目前全靠模板+人审；②我 v2 版的 G1/G4"机检门"是虚构；③`atomically_lock_execution_batch` 动作目前是证据记录占位（actions.go:348-350），原子锁的"原子"依赖迁移事务而非独立动作；④契约条款是 Markdown 表格（§n），无结构化数据形态——REQ-040 的内容 schema 工作若覆盖契约需另立。

## 变更记录

| 日期 | 版本 | 变更 |
|:--|:--|:--|
| 2026-08-14 | v1/v2 | 前两版（被判空洞→叙事不清；v2 的契约机检门系虚构） |
| 2026-08-14 | v3.0.0 | 叙事版；机制事实经 sub-agent 调查核实（模板三张映射表/SYNC 四列对照/UI 门=not_ready/追溯无机检等如实入档） | owner 复核 |
