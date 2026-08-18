# L3-S3 — 契约（Contracts）

> 层：第三层 ｜ 上游：L2 §S3 + L2 全局规则「单一验证分母」 ｜ 机制事实经调查核实，含 file:line

## 1. 要实现什么

把需求+设计翻译成**分端执行契约**（索引 + FE/BE/SYNC），让前后端对同一接口的理解在锁定的文本上重合——契约是"防止各自想象"的共同事实声明，也是 Builder 的唯一边界。

- 进入时：S2 出口（架构决策 + 模块真相包）。
- 出去时：`GATE-PLANNING-CONTRACTS-COMPLETE` 可满足——locked 契约文档（磁盘声明或已注册）+ `planning_contract` 证据（Contract Planner/Orchestrator，pass）。
- 衡量：**每条验收标准都能沿 REQ→Rule→CASE→Story→PATH→条款 的链走通**，且 SYNC 契约里前端行为与后端行为逐项对得上——两端没有各自想象的空间。

## 2. 手头有什么（真实机制）

| 机制 | 它能干什么 | 载体 |
|:--|:--|:--|
| CONTRACTS 索引模板 | 头部稳定性元数据；**需求覆盖矩阵**（REQ→Rule→CASE→Story→PATH→条款→TASK 一表打穿，条款宇宙唯一居所）；锁定依据行为指路行（权威=runtime documents[]+journal） | `docs/contracts/CONTRACTS-template.md` |
| BE/FE 契约模板 | §2 范围+排除；§3 输出契约表+需求条款映射（source_ref/Rule/CASE/条款§n/验收标准）+ Rule→CASE→…→Spec→Evidence 表（BE oracle 同场景包结构）；FE 另有场景与测试映射表与联调点表 | `BE/FE-contract-template.md` |
| SYNC 契约模板 | **两端形状对照的真正载体**：上游需求表（含本合同条款列——SYNC 自己的条款号唯一声明居所）；UI 设计包映射四列表；wire shape/error/idempotency/state 断言表；原始 HTTP+JSON 样例；错误码表；契约测试用例表 | `SYNC-contract-template.md` |
| `contracts check`（机器） | 挂 PTR-PLAN-02 guard 链 + doctor/validate：①断链引用（CASE/S/F/PATH/FR token 须存在于模块包与 REQ；**分母权威=scenario-model.json 的 branch case_id**——model↔cases 不一致显红，生成物篡改无法缩小分母）②条款格数字精确匹配（索引 cell 的 §n 须在目标契约正文声明）③反向闭合（每个 CASE 必被某契约引用）④指纹列与磁盘实算比对⑤空仓地板（零契约不放行） | semantic/contracts.go；guards.go |
| 证据登记 | `runtime evidence add --kind planning_contract`（slot=planning_contract_record，kind=planning_contract）；信封教学见 specification-planning 的 Planning Evidence Envelopes 节 | catalog.go:345-347 |
| documents[] 登记 | PTR-PLAN-02 action `register_locked_contracts`：扫描 docs/contracts 状态=locked 者写入 documents[]（同代同 id 替换不堆叠；无状态=跳过、错状态=fail）——hook 保护自此对契约生效；TR-002 挂幂等补登 | actions.go |
| 规划门推进 | PTR-PLAN-02（contracts→tasks）挂 GATE-PLANNING-CONTRACTS-COMPLETE（contracts_checked+scenario_bridge_checked 双 guard），PreToolUse 自动迁移 | loop-definition.json:162-181 |
| UI 先决（真实形态） | **Quality Gate not_ready**：写契约而真相包不齐 → 门不满足，不前进 | agent-protocol.md:235 |
| 规则 | api-design："APIs are contracts"；X-Request-ID、ISO 8601 UTC、金额禁浮点、错误码必定义调用方行为、兼容性矩阵（必填新增=breaking→ADR）；naming：一概念一名字 | `docs/rules/api-design.md`、`naming.md` |
| skill | `specification-planning` step 10（起草顺序 FE→BE→SYNC、条款对齐、翻状态、信封登记、contracts check）；`api-contracts`（locked contract owns interface semantics，禁静默扩展） | skills/ |

### 2.1 语义分工（哪些归机器、哪些归人）

| 环节 | 归属 |
|:--|:--|
| 引用存在性（token 对账） | 机器（contracts check ①） |
| 条款号一致性（索引↔契约正文） | 机器（contracts check ②） |
| 分母完整性（CASE↔契约双向） | 机器（contracts check ③+bridge 分工：S2 桥管 REQ 侧 AC↔CASE，本查只对账契约侧） |
| 指纹新鲜度 | 机器（contracts check ④+documents 登记） |
| 两端对照（SYNC 四列语义对齐） | 人审（S5）——形状对照是语义判断，机器 diff 查结构查不了意图 |
| oracle 翻译忠实度 | 人审（S5 翻译抽查——手抄即翻译，S3 最深的思考动作） |
| 条款语义/粒度 | 人审（S5） |

## 3. 选了什么、为什么（含否决）

| 子问题 | 选用 | 否决与否决理由（减法） |
|:--|:--|:--|
| 怎么表达两端一致性 | SYNC 模板四列对照表（同一事实行的前端行为/后端行为并排）+ 原始 HTTP 样例 + 契约测试用例表 | 否决"机器 diff 两端 schema"——形状对照是人的语义对齐，表格+CT 用例已够 |
| 怎么承载追溯 | 模板三张映射表（结构）+ contracts check（token/条款/分母/指纹——机械四查）+ S5 评审（语义） | 否决"全靠 S5 人审兜底"——机械可判项左移到收口即门判 |
| 怎么管 UI 先决 | Quality Gate not_ready | 否决 hook deny/warn——用门比用拦截温和且够用（D3） |
| 起草顺序 | FE→BE→SYNC | 顺序即架构立场：先界面事实，再后端实现，再对齐两端 |
| 孤儿条款检查 | 延后——契约模板不存在编号条款清单（§n 仅在映射表内=自证循环），待结构化条款载体（REQ-040 判断）后再立 | 否决"硬做"——对象不存在硬做=恒过的仪式 |
| documents[] 登记点 | PTR-PLAN-02 action（出口门消费 documents[]，喂食必须在门求值前）+ TR-002 幂等补登（tasks 阶段恢复路径） | 否决"仅 TR-003 一处登记"——喂不上 S3 自己的出口门；否决"CLI 副作用写"——迁移 action 是唯一带原子性的自然路径 |
| 锁定依据行 | 模板不手填，指路到 runtime documents[]+journal | 否决"迁移回填"——事务外文件写要么回滚留脏文件、要么提交后 sha 失配；锁定事实的权威居所本就是 documents[]+journal（D1） |
| 代际豁免 | RefreshFingerprints 与 reachability 的 superseded 代豁免覆盖全 kind——旧代条目=冻结历史（不刷新、不移动） | 否决"req 特例"——登记落地后契约也有多代，特例口径即漂移源 |
| 活矩阵居所 | CONTRACTS 索引的需求覆盖矩阵（契约类文档，agent 可写）；REQ §F 留骨架注指路 | 否决"§F 直填"——REQ 被 hook 人-only+基线不可变三重锁死，作机检输入不可行 |
| 锁定语义 | 模板 status 字段 + PTR-PLAN-02 登记把它升格为机器事实 → 此后写入受锁定产物拦截 | 否决"起草期就锁"——S3 产物允许迭代；锁的权威在登记，S3 只声明 |
| 契约变更 | 兼容性矩阵：可选字段新增=兼容（更新 SYNC+CT）；必填新增/删除=breaking（ADR+版本升级）+ change-control | 否决"一律重锁"——兼容变更走轻路径，breaking 才重 |

## 4. 怎么编排（时间线讲完一件事）

1. **UI 先决自检**：ui_impact=changed 时先确认真相包就绪——否则写契约也白写（门不会满足，missing 会指回来）。
2. **FE 起草**：按 FE 模板——§3 逐项挂原型包文件、场景与测试映射表把 CASE/polarity/oracle/PATH 抄进契约、联调点列 METHOD+PATH。
3. **BE 起草**：范围+排除、需求条款映射逐行 source_ref、oracle 链与场景包同构。
4. **SYNC 对齐**：四列对照表逐行写"同一事实的两端行为"；错误码表定义到前端行为级；原始 HTTP 样例+CT 用例落纸。
5. **索引收口**：CONTRACTS 索引汇总清单、覆盖矩阵（每格 `{id} §{n}` 须与目标契约正文的「本合同条款」列一致）——验收标准→条款的最后一遍人肉对账。
6. **机检+登记+推进**：翻各契约状态行（模板「状态」行）为 locked → 登记 planning_contract 信封 → `contracts check` 绿 → 下一次 PreToolUse 评估门（双 guard）→ PTR-PLAN-02 自动迁移进 tasks（register_locked_contracts 同步登记，hook 保护生效）。
7. **返工回路**：S5 审查发现契约矛盾 → fix_required → TR-004 回 planning → 主会话修复 → 受影响职责以新指纹重跑。

## 5. 期望效果

走完 S3：

- **两端无各自想象**：SYNC 对照表让每个字段/错误/状态/副作用的前后端行为并排可见；oracle 链让验证语言从场景包一路贯到契约；
- **结构性防住**：接口语义漂移（locked contract owns semantics+禁静默扩展）、breaking 变更溜过（兼容性矩阵+ADR）、命名漂移（一概念一名字）、UI 契约先于原型（门 not_ready）、断链/假分母/条款错位/指纹陈旧（contracts check 四查+空仓地板）；
- **交给 S4**：契约条款（§n）+ oracle 链——TASK 收尾契约判据的素材；**交给 S5**：契约集（已登记 documents[]）——文档验证的对象。

## 6. 注意力预算与渐进披露

总评：机械四查左移后，S3 的注意力重心只剩**语义层**——SYNC 对照与 oracle 翻译。判定尺见 L3-README「注意力分配原则」。

### 6.1 阅读预算（谁在何时读什么）

| 角色 | 进入时必读 | 触发式加载 | 永不需要读（机制承载） |
|:--|:--|:--|:--|
| 起草者 | FE/BE/SYNC 模板（起草顺序由 skill 承载） | 写到具体字段时读该字段的行内规范/api-design 对应节 | "追溯必须完整"的叙述——contracts check 点名缺口；UI 先决——gate not_ready 指回来 |
| S5 审查者 | **语义对齐两处**：SYNC 四列（两端行为是否同一事实）+ oracle 翻译抽查（映射表的 oracle 翻译是否忠实于场景包原字段——手抄即翻译，不能无消费者） | — | 机械核对（引用/条款/分母/指纹）——contracts check 承载 |
