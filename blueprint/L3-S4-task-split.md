# L3-S4 — 任务拆分（Task Split）

> 层：第三层 ｜ 上游：L2 §S4 ｜ 版本 v4.0.1（v4.0.0 设计定稿 owner 已拍板；+对抗审查处置，见变更记录）

## 1. 要实现什么

把契约条款拆成**单职责、可独立验证的任务**——每个任务一份"完成判据先行"的收尾契约，让"完成"在动手之前就有可判定的定义。

- 进入时：S3 出口（locked 契约集 + oracle 链 + CONTRACTS 索引 = 条款宇宙）。
- 出去时：TR-002 可满足——TASK 批次全 complete、条款覆盖双向闭合、依赖图无环、批次已登记 documents[]。
- 衡量：L2 出口三承诺——每条款至少一任务覆盖（**机器承载**）、写路径重叠有显式串行归属（S5 判断层承载）、判据先行（模板 + S5 承载）。

**本 stage 的载体原则（与 S3 一脉相承）：每个事实只有一个居所。** 指纹/版本/登记态 → runtime documents[]（CONTRACTS-template 第 10 行已为契约立法"锁定状态与依据见 runtime documents[] 与 journal——文件内不再手填"，本 stage 把这条原则推广到任务文档）；条款宇宙 → CONTRACTS 索引；覆盖声明 → TASK §3；场景映射 → S2 模块包；进程状态 → 看板；文档状态 → Status 行。

## 2. 手头有什么（真实机制）

| 机制 | 它能干什么 | 载体 |
|:--|:--|:--|
| TASK 模板 | 头部（Status/Version/Source REQ/Primary contract/Team manifest/Assignment）；§1 一句话可测目标；§2 读序表；§3 条款清单（交付条款，相对 Primary contract）；§3.1 模块影响声明；§4 三类路径+命令类；§5 技能选择；§6 产出与证据；**§7 收尾契约四行 assert**；§8 依赖表；§9 生命周期证据；§10 发现与修复；§11 历史 | `docs/tasks/TASK-template.md`（v4 瘦身形态，见 §3） |
| 模板契约校验（机器） | TASK 模板必含字段检查（现状词表：Team manifest/Assignment ID/Document Manifest/SHA-256/Selected Skills/Lifecycle Evidence/Closing Contract；v4 批次1 改为：去 SHA-256、加 Delivered clauses/Module Impact）+ 禁含旧机制词 | migration/templates.go:68-79 |
| 条款宇宙 | CONTRACTS-{id} 索引的需求覆盖矩阵：REQ→FE/BE/SYNC 条款 cell（`{id} §{n}` 格式）——活矩阵唯一居所（S3 终批） | docs/contracts/CONTRACTS-template.md 需求覆盖矩阵 |
| 批次质量检查（机器） | `tasks check`：批完整性（全批 complete，cancelled 跳过并注记）+ 主契约存在 + 条款覆盖双向（宇宙→声明 / 声明→宇宙）+ 依赖图无环（DFS 报环路径） | internal/semantic/tasks.go（本轮新增）+ internal/cli/tasks_check.go |
| 规划门推进 | TR-002 挂 guard `planning_complete`（契约看 documents[]、任务看磁盘全批）+ guard `tasks_checked`（跑 tasks check）+ action `register_planning_tasks`（TASK 批次入 documents[]） | loop-definition.json TR-002；guards.go；actions.go |
| 证据登记 | `runtime evidence add --kind planning_task`（slot=planning_task_record） | catalog.go:348-350 |
| skill | `specification-planning`（S4 步骤：每 TASK 绑一契约+条款清单+收尾契约+单职责；TR-002 前核对真实条件）；`dag-design`（纯方法论：怎么拆/粒度/关键路径——查环职责已左移机器） | skills/ |
| 看板模板 | 任务矩阵（任务/契约来源/负责人/预估/依赖/状态/阻塞）+ 关键路径文本声明——S4 收口一次性填写+人向总览（执行期进程真相=§9 证据引用；文档 Status 不再承载过程词汇） | `docs/tasks/index-template.md` |

**状态词汇（v4 收敛后）**：TASK 文档 Status 只说文档生命周期 `{draft, complete, cancelled}`；执行进度看 §9 证据引用 + 看板状态列；实体状态机（entities.tasks 八态）无 CLI 入口、有机流程无人推进——如实标注为纸面机制（见 §2.1 #7），是否激活留 S6 裁定。

## 2.1 断点地图（联合调查核实，file:line 为 v3 现状）

| # | 断点 | 证据 | 级别 |
|:--|:--|:--|:--|
| 1 | **TR-002 强分支永不可满足**：`guardPlanningCompleteFn` 先试 documents[] 分支（要求当前代同时有 contract+locked 与 task+complete），失败无条件降级文件名弱回退。契约由 PTR-PLAN-02 登记，TASK 登记动作却挂在 TR-003（晚于 TR-002）——TR-002 时刻 documents[] 必然"有契约无任务"，强分支必失败 | guards.go:347-361；TR-002 actions=[]（loop-definition.json:1051 附近）；TR-003 actions 含 register_execution_batch | **P0** |
| 2 | **弱回退只看一个文件**：checkArtifactStatus 找到一个 locked 契约+一个 complete TASK 即过门——不查全批、不查代际、不查指纹；E2E 靠 fixtures 手播 TASK 条目掩盖（planning_chain.go:45-49） | guards.go:423-455 | P0 |
| 3 | 条款覆盖零机检：L2 承诺"每条款 ≥1 任务"无任何机器承载；§3 四列矩阵中 REQ 列抄索引、output 列抄 §1/§6、VER 列零消费者 | 全库无对应代码；TASK-template.md §3（v3） | P1 |
| 4 | DAG 无环零机检：§8 依赖表全库无解析；team/validator.go:168-229 已有 DFS 环检测+依赖存在性先例（team manifest 的 depends_on）——仅返回 bool 不报环路径，参照重写扩展 | internal/team/validator.go:168-229 | P1 |
| 5 | 手抄指纹三连：§2 读序表 5×64 字符 SHA-256、§5 技能指纹、§9 证据指纹——全部与 runtime documents[] 重复，违反 CONTRACTS-template:10 已立的"文件内不再手填"原则 | TASK-template.md §2/§5/§9（v3） | P1 |
| 6 | 状态词汇三套并存：模板 8 词（v3 文档误记"九态"）/ 实体 8 态 / documents[].status 事实词汇（complete/locked）；机器只读 `complete` | TASK-template.md:3；loop-definition.json:818-900 | P1 |
| 7 | 实体状态机纸面化：AdvanceTask（含 review→done 真检查）仅库级+测试调用，runtime 子命令无 task-event；S1 unbind 在飞检查读 entities.tasks 永远为空——软门静默失效（跨 stage 债，S6 表面） | run.go:906,1091,1135；assignment/task_lifecycle.go:165-172 | P1（如实入档，不在本轮修） |
| 8 | `atomically_lock_execution_batch` 死代码：TR-003 actions 已换 register_execution_batch，registry 孤儿条目制造"有原子锁动作"假象 | actions.go:115,491-493 | P1（删） |
| 9 | §3.1 六列场景表抄 S2 模块包切片（S7 读模块包不读此表）+4 checkbox 中 3 条是 S2 不变量复述 | TASK-template.md §3.1（v3） | P2 |
| 10 | 顺手债：specification-planning skill 双过期（Entry Conditions 列已删 phase 名:13、Inlined Methodology 单 phase 表述:127）；`contracts_checked` 无 guard_specs 词条；L3-S5:35 占位行号过时+§55 两缺口已被 S3 解决；TR-003 登记后 TASK 不受 hook 硬保护（ LockedArtifacts 只认 locked/active，hookctx/loader.go:271-274） | 各处 | P2 |

**重新裁定不算断点**：TR-003 后 TASK 文件不入 LockedArtifacts 是**正确设计**——S6 执行期要写 TASK 文档（§9 生命周期证据），硬保护会挡 S6；版本锁定由 documents[].sha256 对账承载（改动=validate 红）。索引"派生 TASK"列与任务覆盖声明双居所必漂移——v4 删列（owner 已拍板）。§10 Findings/§11 History 两个手抄表（抄 entities.bugs/journal）是单一居所原则的漏网之鱼——主要由 S6-S9 填写，留 S6/S8 设计时裁定，本轮入档不动。

## 3. 选了什么、为什么（含否决）— 设计定稿 v4.0.0（owner 已拍板）

| 子问题 | 选用 | 否决与否决理由（减法） |
|:--|:--|:---|
| 指纹/版本在哪 | **runtime documents[] 唯一居所**：§2/§5/§9 全部删手抄指纹列 | 否决"保留列+报错带全值方便回填"——给坏机制抛光不如删错误类本身；否决"新增 manifest 生成命令"——CLI 原则：可派生参数不设专门命令 |
| 覆盖在哪声明 | **TASK §3 条款清单**（`{contract-id} ｜ §2, §3` 相对 Primary contract）+ 机器聚合对账；**索引删"派生 TASK"列** | 双居所必漂移；索引=条款宇宙（S3 已锁），任务=覆盖声明，各司其职 |
| 条款宇宙权威源 | **CONTRACTS 索引矩阵 cell**（contractClauseCellPattern 可复用，contracts.go:20） | 否决"契约文档内部解析 §n"——契约条款只是表格 cell，无标题结构，无处锚定真伪 |
| 场景覆盖 | **§3.1 缩为模块影响声明**（触及模块 + 真相文件变更即全模块回归承诺） | 六列表抄模块包零消费者；S2 不变量（分支 100%/容量比/fixture 隔离）由模块包承载，不逐任务复述 |
| 状态词汇 | **Status 三词 {draft, complete, cancelled}**；进程状态归看板 | 三套词汇=记忆税+翻译错误面；且词汇收敛反过来保护登记不变式（register 拒绝非 complete，working 等词不复存在=无法误填破坏批次） |
| 算术左移 | **`tasks check` 四查**：批完整性 / 主契约存在 / 条款覆盖双向 / DAG 无环（移植 team DFS） | 否决"§4 写路径交集机检"——目录粒度必误报，实际写入由 S6 激活面+hook 把守，规划期归属是 S5 DV 的判断层职责 |
| TR-002 接线 | guard `planning_complete`（重写：契约看 documents[]+磁盘一致、任务看磁盘全批 complete）+ guard `tasks_checked` + action `register_planning_tasks` | 与 S3 的 PTR-PLAN-02 完全对位（guard 跑真检查、action 登记）；删"documents 优先/文件名回退"双分支——它让强检查静默弱化，改为失败即指路（"PTR-PLAN-02 是否已跑？"） |
| 怎么定义"完成" | **保持** §7 四行 assert 文本块——判据先于实现；tasks check 补**实例级**在场检查（每 TASK 文件含 Closing Contract 块+assert 行，闭合 L2 第三承诺的机检） | criterion_id 结构化是 REQ-040/041 演进方向，不抢跑 |
| 怎么防巨型任务 | **保持** §1 一句话目标+单一 Primary contract+S5 粒度审查 | 否决行数硬上限机检——语义判断，硬阈值误伤 |
| 怎么锁任务 | **保持"不在此锁"**——S4 产 complete；TR-003 `register_execution_batch` 幂等重锁（同代替换语义） | 锁的权威在 S5（双职责 PASS），早锁挡返工 |
| 依赖怎么管 | §8 表（现被 tasks check 解析：TASK- 引用存在+全批 DFS 无环，报环路径）；assignment 引用留给 team validator（已有先例） | dag-design skill 卸下查环职责，回归纯"怎么拆"方法论 |
| 实体状态机 | 如实标注纸面（断点 #7），激活与否留 S6 | 本轮加 task-event CLI = 为零消费者扩命令面 |

## 4. 怎么编排（时间线讲完一件事）

1. **推导**：沿索引条款宇宙逐条推导任务；每任务先写 §1 一句话目标（写不清=拆得不对）与单一 Primary contract。
2. **装填**：§2 读序（无指纹，纯指引）→ §3 交付条款清单 → §3.1 模块影响 → §4 三类路径与命令类 → §5 技能（名称+版本）。
3. **判据**：§7 收尾契约按 4 行 assert 写——"验证命令通过"一行直接引契约 oracle 链的验证方式。
4. **依赖**：§8 声明 TASK- 依赖与所需证据；有环由 `tasks check` 报环路径，重拆。
5. **看板收口**：index 看板填任务矩阵与关键路径——进程状态与给人的总览都在这。
6. **自检**：`tasks check`——批完整性/覆盖双向/无环全绿才请求 TR-002。
7. **登记推进**：TR-002 双 guard（planning_complete+tasks_checked）→ action `register_planning_tasks` 把批次登记进 documents[]（author_agent_id/registered_at/sha256）→ S5。（自动触发首趟必 not ready——GATE 求值需要 documents[] 已有当前代 task 条目，而登记恰是本迁移自己的 action；首趟经手工 CLI apply，与 S3 PTR-PLAN-02 同 posture，非回归。）
8. **锁定在下游**：S5 双职责 PASS 后 TR-003 幂等重锁——本阶段产物仍是可返工的。

## 5. 期望效果

走完 S4：

- **完成有判据**：每任务动手前就有 assert 级判据；S6 兑现、S7 复验都对着这份判据。
- **结构性防住**：多职责巨任务（一句话目标+单契约绑定）、范围蔓延（三类路径+forbidden 含 loop-state）、"差不多完成"（assert+S5 审）、条款漏覆盖（机器双向对账）、批次死锁（DFS 无环）。
- **交给 S5**：TASK 批次在 documents[] 在册——DV-TASK-EXECUTABILITY 的审查对象与 exactSubjects 独立性检查从此有 TASK 侧真数据（author_agent_id）。
- **注意力减负**：模板 126→约 90 行；每任务少算 5 个哈希、少抄两份切片；agent 只在与自己工作局部相关处声明（条款清单/依赖），聚合对账交机器。

## 6. 注意力预算与渐进披露

总评（v3 诊断升级）：语义判断（单职责/判据先行）的载体本来就正确（模板字段即逼问）；真正的病灶是**手抄状态与重复居所**——v4 用"每个事实只有一个居所"一刀切解决。判定尺见 L3-README「注意力分配原则」。

### 6.1 错配与处置（v3 五项 → v4 处置）

| # | v3 错配 | v4 处置 |
|:--|:--|:--|
| 1 | DAG 无环靠 skill 自查 | 左移：tasks check DFS（复用 team/validator.go 先例）；dag-design 回归方法论 |
| 2 | 条款覆盖无机器校验 | 左移：索引宇宙 ↔ 任务声明双向对账 |
| 3 | 双状态词汇靠 guard 弥合 | 根治：Status 三词，进程归看板；v3"九态"系误记（实为 8 词） |
| 4 | 读序 SHA-256 手填 | 删列：指纹唯一居所=runtime documents[]（错误类整体消失） |
| 5 | 占位 action 假象 | 删：atomically_lock_execution_batch（死代码，conformance 单向无回归面） |

新增处置：P0 强分支接线（§3 TR-002 行）；§3 矩阵缩条款清单；§3.1 缩模块声明；实体状态机如实标注纸面；TR-003 后 TASK 不受 hook 硬保护=正确设计（sha 对账承载锁定，硬保护会挡 S6 写 §9）。

### 6.2 阅读预算（谁在何时读什么）

| 角色 | 进入时必读 | 触发式加载 | 永不需要读（机制承载） |
|:--|:--|:--|:--|
| 拆分者 | TASK 模板 + 索引条款宇宙 | 拿不准拆分粒度/依赖设计时载 dag-design（方法论） | 环检测/覆盖对账/指纹——tasks check 承载；状态词汇映射表——词汇已唯一 |
| 人（看板） | index 任务矩阵 + 关键路径 | — | — |

### 6.3 实施清单与真度表（划线=已落地并验证）

**批次 1 模板层**
- [ ] TASK-template v4：Status 三词；§2 列 `Order/Kind/ID/Path/Clauses`（删 Version/SHA-256 列与自我行）；§3 → Delivered Clauses 表（`| Contract | Delivered clauses |` 单行）；§3.1 → Module Impact 两行声明；§5 删 SHA-256 列；§9 列 `Evidence/Reference`（删 Fingerprint 列）
- [ ] CONTRACTS-template 需求覆盖矩阵删"派生 TASK"列
- [ ] migration/templates.go：TASK 必含词去 "SHA-256"、加 "Delivered clauses"/"Module Impact"；测试同步

**批次 2 检查与接线层**
- [ ] internal/semantic/tasks.go：TasksCheck **五查**——①批完整性：TASK-*.md（非模板）全 complete、cancelled 跳过并注记、≥1 存在；②主契约存在：Primary contract 对应 docs/contracts/{id}*.md 在盘；③收尾契约在场：每 TASK 文件含 Closing Contract 块+assert 行；④条款覆盖双向，文法规格——**索引 cell=每条款一个 `{id} §{n}` 记号**（允许多记号并排，跨行集合去重）；**任务 §3=逗号分隔 §n 列表，空=支撑性任务、不计入覆盖聚合**；**cancelled 任务的声明剔除出覆盖聚合**（其条款漏覆盖显红，不许静默击穿）；宇宙下限 ≥1 cell（索引缺失/矩阵空=红，指路）；docs/contracts 每个 FE/BE/SYNC 文件在宇宙有 ≥1 cell（契约↔索引双向，堵假绿洞）；缺向指名 cell、幽灵条款指名任务；⑤DAG：§8 TASK- 引用存在（引用 cancelled 任务=problem）、全批 DFS 无环**报环路径**（参照 team/validator.go 先例重写扩展）
- [ ] internal/cli/tasks_check.go：`tasks check --root [--json]`（输出 tasks/cancelled/clauses_total/clauses_covered/problems）
- [ ] guards.go：`tasks_checked` guard（挂 TR-002，root 取 state）；`planning_complete` 重写（契约=当前代 documents[] contract+locked，磁盘一致**定义为 Status 字段 EqualFold**——沿用 verifyDocumentStatusOnDisk；sha 由登记动作与 reachability 承载，guard 不重复查；无契约条目则指路"PTR-PLAN-02 是否已跑"；任务=磁盘全批 complete；删双分支与文件名回退）
- [ ] actions.go：`register_planning_tasks`（直接复用 registerDocumentsFromDisk，prefixes=["TASK-"]，wantStatus=complete；**不套 actionRegisterExecutionBatch 的 ctx.Evidence 非空前置**——TR-002 required_evidence=[]，套了即永远失败）；registerDocumentsFromDisk 增 cancelled 跳过语义（TR-003 同享）；删 atomically_lock_execution_batch（registry+实现+注释）
- [ ] loop-definition.json TR-002：guards=["planning_complete","tasks_checked"]，actions=["register_planning_tasks"]
- [ ] guard_specs.go：补 `tasks_checked` + `contracts_checked`（S3 遗留）词条；**重写 planning_complete 词条**（现文案"falls back to filename patterns"在删回退后成谎言）
- [ ] 回归面：engine_test.go:175-186,219,400-420（TR-002 夹具改 documents[] 形态）、task039_01_planning_test.go 12 例、req039 planning_chain 夹具、conformance 单向确认

**批次 3 清理与同步层**
- [ ] skills/specification-planning：Entry Conditions 换现行三 phase；Inlined Methodology 三段 phase 表述；S4 步骤 12-14 对齐新模板（条款清单/模块声明/无指纹抄写/TR-002 双 guard）
- [ ] skills/dag-design：卸查环职责表述（Quality Criteria/Operating Procedure 措辞）
- [ ] skills/document-verification：覆盖/DAG 手工审查指令（"every contract clause maps to at least one TASK"等）改为消费 tasks check 输出——其原数据源（派生 TASK 列）已删
- [ ] blueprint 同步：L3-S5:35 占位行+§55 已解决两缺口+§6.1#1+§4 步骤 2（覆盖/DAG 已左移，DV 转消费输出）；L3-S6:34（criterion_id 已推迟，勿当既成事实）；L3-S3 涉占位的历史叙述核对
- [ ] docs/loop-harness.md：tasks check 命令 + TR-002 变更

**批次 4 E2E 与真度清零**
- [ ] TestS4TaskSplitPipelineE2E：绿路径（索引宇宙+条款清单+依赖链）→ 断链红各指名（幽灵条款/漏覆盖 cell/缺依赖 TASK/双任务互依环/批次夹 draft/**cancelled 后其条款漏覆盖显红**）→ TR-002 双 guard 过+**空 evidence 下 action 成功**+登记（author/sha 对磁盘）→ **planning_complete 失败指路文案断言**+**tasks_checked guard 层错误穿透**（非仅 CLI 红）→ **修复回环二趟**（TR-004→改文档→再 TR-002，documents[] 携同代旧条目不误拒——generation 只在 TR-020 递增，回环安全）→ cancelled 跳过 → TR-003 幂等重锁（同代替换不堆叠）
- [ ] 全库 grep **TASK 文档 Status 叙述**同步（activated/working/reported/stale——实际仅 TASK-template.md:3 一处；勿伤实体词汇：reported=BUG 模板与实体态、stale=workgroup 枚举、working=write_scope_enforced 词条）；validate --all 绿；真度表全部划线

## 变更记录

| 日期 | 版本 | 变更 |
|:--|:--|:--|
| 2026-08-14 | v1/v2 | 前两版（被判空洞→叙事不清；v2 的 criterion_id 结构与无环机检系虚构） |
| 2026-08-14 | v3.0.0 | 叙事版；机制事实经 sub-agent 调查核实 | owner 复核 |
| 2026-08-15 | v3.1.0 | 新增 §6 注意力预算与渐进披露 | owner 指示：渐进披露、机制承载规范 |
| 2026-08-16 | v4.0.0 | 联合调查核实（P0 强分支断点/三套词汇/手抄指纹三连等 10 项入档）；设计定稿：单一居所原则推广+tasks check 左移+TR-002 对称接线+模板瘦身+收回两项过度设计（指纹报错抛光/写路径交集机检） | owner 拍板：索引删派生列、Status 三词、其余按建议 |
| 2026-08-16 | v4.0.1 | 设计对抗审查处置（无 P0；P1×5 全采纳）：①条款文法规格化+宇宙下限+契约↔索引双向（堵假绿洞）②cancelled 声明剔除覆盖聚合+死依赖计 problem ③register_planning_tasks 明示无 evidence 前置 ④E2E 补修复回环/空 evidence/指路文案/guard 穿透四断言 ⑤planning_complete 词条重写入批次。P2 随批：DFS 措辞（参照重写非移植）、测试计数 14→12、"磁盘一致"=Status 定义、grep 限域 TASK 文档、自动触发首趟 posture 入档、看板口径改"收口填写+人向总览"、§10/§11 双居所入档留 S6/S8、收尾契约实例级机检（第五查）、L3-S5 §4/L3-S6:34/document-verification skill 同步 | 对抗审查（sub-agent）+ 全采纳 |
