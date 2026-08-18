# L3-S4 — 任务拆分（Task Split）

> 层：第三层 ｜ 上游：L2 §S4 ｜ 机制事实经调查核实，含 file:line

## 1. 要实现什么

把契约条款拆成**单职责、可独立验证的任务**——每个任务一份"完成判据先行"的收尾契约，让"完成"在动手之前就有可判定的定义。

- 进入时：S3 出口（locked 契约集 + oracle 链 + CONTRACTS 索引 = 条款宇宙）。
- 出去时：TR-002 可满足——TASK 批次全 complete、条款覆盖双向闭合、依赖图无环、批次已登记 documents[]。
- 衡量：L2 出口三承诺——每条款至少一任务覆盖（机器承载）、写路径重叠有显式串行归属（S5 判断层）、判据先行（模板+S5 承载）。

**载体原则（与 S3 一脉相承）：每个事实只有一个居所。** 指纹/版本/登记态 → runtime documents[]；条款宇宙 → CONTRACTS 索引；覆盖声明 → TASK §3；场景映射 → S2 模块包；进程状态 → 看板；文档状态 → Status 行。

**拆分纪律（builder 视角，写在 specification-planning step 11）**：一句话说不成交付物、或出现"以及/然后"→ 拆；FE+BE+SYNC 不混进同一任务；类型/schema/迁移是地基，下游任务必须声明对它的依赖。**compact 警示：尽量避免 subagent 中途 compact——builder 丢任务信息是灾难性表现**。宁可多拆一个任务，也不要让 builder 读着读着上下文被压缩；每个任务的 §2 清单只引用它真正需要的条款切片。规模直觉锚：必读合计 ~30KB / 触碰 ~8 文件 / 改动 ~400 行（参考非门槛，`tasks check` 输出 reference load 供对照）。

## 2. 手头有什么（真实机制）

| 机制 | 它能干什么 | 载体 |
|:--|:--|:--|
| TASK 模板 | 头部（Status 三词/Version/Primary contract）；§1 一句话可测目标（含一句话测试与 compact 警示注）；§2 读序表（无指纹列）；§3 交付条款清单；§3.1 模块影响声明；§4 三类路径+命令类；§5 技能选择；**§7 收尾契约四行 assert**；§8 依赖表（只认 TASK-* 引用入 DAG）；§9 生命周期证据 | `docs/tasks/TASK-template.md` |
| 模板契约校验（机器） | TASK 模板必含字段检查（Delivered Clauses/Module Impact；不含旧机制词） | migration/templates.go |
| 条款宇宙 | CONTRACTS-{id} 索引的需求覆盖矩阵——活矩阵唯一居所 | docs/contracts/CONTRACTS-template.md |
| `tasks check`（机器五查） | ①批完整性（全批 complete，cancelled 剔除聚合、漏覆盖显红）②主契约存在 ③收尾契约在场（块+assert 行）④条款覆盖双向（宇宙↔声明，缺向指名 cell、幽灵条款指名任务；契约↔索引双向堵假绿洞；宇宙下限）⑤DAG 无环（DFS 报环路径）。**另输出每任务 reference load**（必读 KB+写路径数，info-only 参考） | internal/semantic/tasks.go；internal/cli/tasks_check.go |
| 规划门推进 | TR-002 挂 guard `planning_complete`（契约看 documents[]+磁盘一致、任务看磁盘全批）+ guard `tasks_checked`（跑 tasks check）+ actions `register_locked_contracts`（幂等补登）+ `register_planning_tasks`（TASK 批次入 documents[]） | loop-definition.json TR-002；guards.go；actions.go |
| 证据登记 | `runtime evidence add --kind planning_task`（slot=planning_task_record）；信封教学见 specification-planning 的 Planning Evidence Envelopes 节 | catalog.go:348-350 |
| skill | `specification-planning` step 11（拆分纪律+compact 警示）/ step 12（Status 翻转+信封登记+tasks check+reference load 对照）；`dag-design`（纯方法论：怎么拆/粒度/关键路径——查环已左移机器） | skills/ |
| 看板模板 | 任务矩阵+关键路径——S4 收口一次性填写+人向总览（看板状态列是进程状态，与文档 Status 三词正交） | `docs/tasks/index-template.md` |

**状态词汇**：TASK 文档 Status 只说文档生命周期 `{draft, complete, cancelled}`（complete=文档写完，与实现无关）；执行进度看 §9 证据引用+看板。实体状态机（entities.tasks 八态）无 CLI 入口、有机流程无人推进——纸面机制，是否激活留 S6 裁定。

## 3. 选了什么、为什么（含否决）

| 子问题 | 选用 | 否决与否决理由（减法） |
|:--|:--|:--|
| 指纹/版本在哪 | **runtime documents[] 唯一居所**：§2/§5/§9 无手抄指纹列 | 否决"保留列+报错带全值方便回填"——给坏机制抛光不如删错误类本身；否决"manifest 生成命令"——可派生参数不设专门命令 |
| 覆盖在哪声明 | **TASK §3 条款清单**+ 机器聚合对账；索引无"派生 TASK"列 | 双居所必漂移；索引=条款宇宙、任务=覆盖声明，各司其职 |
| 条款宇宙权威源 | **CONTRACTS 索引矩阵 cell**（`{id} §{n}`） | 否决"契约文档内部解析 §n"——条款只是表格 cell，无标题结构无处锚定 |
| 场景覆盖 | **§3.1 缩为模块影响声明**（触及模块+全模块回归承诺） | 六列表抄模块包零消费者；S2 不变量由模块包承载，不逐任务复述 |
| 状态词汇 | **Status 三词**；进程归看板 | 三套词汇=记忆税+翻译错误面；词汇收敛保护登记不变式（无法误填破坏批次） |
| 算术左移 | **tasks check 五查** | 否决"§4 写路径交集机检"——目录粒度必误报；实际写入由 S6 激活面+hook 把守，规划期归属 S5 判断层 |
| TR-002 接线 | 双 guard（真检查）+ 双 action（登记）；无文件名回退——失败即指路 | 与 S3 的 PTR-PLAN-02 完全对位 |
| 怎么定义"完成" | **保持** §7 四行 assert 文本块+ tasks check 实例级在场检查 | criterion_id 结构化是 REQ-040/041 演进方向，不抢跑 |
| 怎么防巨型任务 | §1 一句话测试+单一 Primary contract+S5 粒度审查+compact 警示（提示词形态）+reference load 参考数字 | 否决行数硬上限机检——语义判断，硬阈值误伤且限制模型创造力（owner 裁定：提示词+审查要点形态，非硬门禁） |
| 怎么锁任务 | **不在此锁**——S4 产 complete；TR-003 `register_execution_batch` 幂等重锁（同代替换） | 锁的权威在 S5（双职责 PASS），早锁挡返工 |
| 依赖怎么管 | §8 表（tasks check 解析：引用存在+DFS 无环报环路径）；非 TASK 引用显式报错（不静默丢边） | dag-design skill 卸下查环职责，回归纯方法论 |

## 4. 怎么编排（时间线讲完一件事）

1. **推导**：沿索引条款宇宙逐条推导任务；每任务先写 §1 一句话目标（写不清=拆得不对）与单一 Primary contract。
2. **装填**：§2 读序（只引用真正需要的条款切片）→ §3 交付条款清单 → §3.1 模块影响 → §4 三类路径与命令类 → §5 技能。
3. **判据**：§7 收尾契约按 4 行 assert 写——"验证命令通过"一行直接引契约 oracle 链的验证方式。
4. **依赖**：§8 声明 TASK- 依赖与所需证据；有环由 `tasks check` 报环路径，重拆。
5. **看板收口**：index 看板填任务矩阵与关键路径。
6. **自检**：`tasks check`——五查全绿+reference load 对照规模锚（超了先想想能不能拆）。
7. **登记推进**：翻各 TASK Status 行为 complete → 登记 planning_task 信封 → TR-002 双 guard 过 → actions 登记批次进 documents[]（author_agent_id/sha256）→ S5。
8. **锁定在下游**：S5 双职责 PASS 后 TR-003 幂等重锁——本阶段产物仍是可返工的。

## 5. 期望效果

走完 S4：

- **完成有判据**：每任务动手前就有 assert 级判据；S6 兑现、S7 复验都对着这份判据。
- **结构性防住**：多职责巨任务（一句话测试+单契约绑定+拆分纪律）、范围蔓延（三类路径+forbidden 含 loop-state）、"差不多完成"（assert+S5 审）、条款漏覆盖（机器双向对账）、批次死锁（DFS 无环）、上下文烧穿（compact 警示+reference load 参考）。
- **交给 S5**：TASK 批次在 documents[] 在册——DV-TASK-EXECUTABILITY 的审查对象与独立性检查从此有 TASK 侧真数据。
- **注意力减负**：每任务少算 5 个哈希、少抄两份切片；agent 只在与自己工作局部相关处声明，聚合对账交机器。

## 6. 注意力预算与渐进披露

总评：语义判断（单职责/判据先行/compact 避免）的载体是模板字段与提示词（字段即逼问）；机器承载全部算术（覆盖/环/指纹/load）。判定尺见 L3-README「注意力分配原则」。

### 6.1 阅读预算（谁在何时读什么）

| 角色 | 进入时必读 | 触发式加载 | 永不需要读（机制承载） |
|:--|:--|:--|:--|
| 拆分者 | TASK 模板 + 索引条款宇宙 + step 11 拆分纪律 | 拿不准拆分粒度/依赖设计时载 dag-design（方法论） | 环检测/覆盖对账/指纹——tasks check 承载；状态词汇映射表——词汇已唯一 |
| 人（看板） | index 任务矩阵 + 关键路径 | — | — |
