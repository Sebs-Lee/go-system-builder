# 全阶段减负与重新验收

本轮按用户要求覆盖 **S0–S11 全部阶段**，并覆盖 S9 子相位、返工、阻塞和人闸。
权威原则为 [全局设计 §0](../GLOBAL-OPTIMIZATION-DESIGN.md)：Hook 监听 graph、提醒
遗漏并触发原流转；给清楚工序、问题、准确填写路径和具体剩余工作；填一次并保存。
取消工具身份/权限/scope 拦截，不把同一拦截移到报告入口。专业判断、独立验证、
真实检查和原人类决策保留。已有消费者、状态和写入事务复用，不增加阶段协议。

## 验收纪律

- 每阶段先梳理并固定最小工作流，再改代码、模板、角色/技能和说明，最后测试。
- 作者只回答本责任的专业问题；已知身份、来源、哈希、检查事实由系统带入。
- 进展反馈必须给当前准确 view/answer 路径、具体缺项、下一动作或等待责任人。
  用现有 Markdown 表/文档承载的责任也直接给真实文件路径，不强制新造 JSON 协议。
- 问卷准备/恢复/投递/接收属于系统工作，不要求作者手工 issue/consume、抄信封、
  重交相同 PLAN/结果或读 Runtime 猜路径。缺接线就记录系统缺口并修复。
- 测试验证正常交付、缺项提醒、错误答案修正、重放不重复推进、重启恢复、
  本人完成与阶段出口区分；涉及独立责任、集成或人闸时分别覆盖。
- 人闸测试检验提醒、完整申请和没有授权不推进；不伪造用户决定或实际发布。
- 旧 4/9 是历史完整链证据，不能当作本轮全阶段重新验收通过。after-r18
  的 S2、S3、S4 真实填卷证据可追溯，但不能代替最新合并候选或其他阶段的验收。

## 本轮阶段清单

### S6 本轮有界范围 done（2026-10-16）

S6 构建、原返工、真实检查/Git 集成与监听交接已完成本轮有界验收，见 [S6 机制审查](S6-MECHANISM-REVIEW.md)。
R40 保留真实 Builder 专业履职；R45 after-r45 机械重放专业前置与产品，真实重跑 checks、Git、ACK 与清理，
保留原 r1 失败并由 revision r2 完成交付。两个真实 Main probe 分别正确识别已有 govet 失败及原恢复动作、
验证原 TR-006 唯一进入 S7 verification.planned 和准确问卷可读；startup/Pre/Stop 均无 deadline/UNKNOWN，停在入口未填 S7。

本轮删除重复 gate canonical 求值、普通 Pre 同 revision 复核/写入、提醒中的 receipt/Git 链和 handoff 特判，
提交后复用 Snapshot 减少 Reconcile/Load；原 Main 失败仅只读展示，不加协议。R42 反馈超时、R43 UNKNOWN/Read 超限、
R44 startup 8 秒超时和两次作者窄测失败保留，修复后窄测通过，重叠结果不累加。历史 after-r42 同源清单仅指当时冻结源。
本结论不代表新 Builder 从零实现、全阶段、多平台或发布完成。

### S5 已完成范围（2026-10-16）

S5 范围已完成，见 [S5 机制审查](S5-MECHANISM-REVIEW.md)。先固定 L3-S5 最小契约，
再修原 TR-004 的批次退出、原 TR-003 当前独立双 PASS、历史清单校验和真实 Worker
目录的回执/进度定位；沿用全局五态，作者仍只填原两字段问卷并保存一次。
r122 单 PASS 跳步、r124 历史进度误报、r30 取消后 activation 校验与旧 Worker 入口
失败均保留原证据，已分别修复。最终 r128 的 48 项定向回归及相关四包 102 项完整
回归通过，覆盖有重叠不相加；after-r32 与其 1636 项源码清单完全一致。
全新安装机械回放历史前置（不计真实新专业判断），再由两位真实 muse 独立审查 e8：
均 PASS，原 `evt-tr-003-r54` 唯一进入 S6/e9。各一次脚本保存，无手工系统步骤；
最新 19 项 S4 源闭包和结构检查随原资料投递。新旧三份答卷各两次重放，准确原回执、
路径提示及零写入全部通过；`tasks check / validate --all / doctor` 均 exit 0。
旧清单引发的现场全局校验失败已关闭；未开展 S6 工作，未声称全阶段或发布完成。

S5 后续题干优化已落地：直接解释三结论的边界，通过时保持预填空清单，有问题才填
位置/预期/实际缺口；模板、Role、Skill、协议同步。旧问卷/修订按原模板保留，状态不变。
r130 的 30 项定向回归及 workpacket 整包 37 项通过；真实 muse 两职责×三结论的
六情景理解仿真全对。after-r33 同源评估包已构建，新二进制只读验证 r32 原现场的
validate/doctor 和三份原回执通过、原文件零写入。该补充不是新的完整双评审链验收，
完整链仍引用 after-r32；详见 S5 审查末节。

### 前次执行记录：S4 局部返工（2026-10-16）

用户已要求开始返工。先在 [L3-S4](../l3-stages/L3-S4-task-split.md) 固定最小契约，
再修原接续、原门/TR-002 与准确交付提示，协议/Skill/模板同步。作者仍只完成专业任务
规划、提交 Git 来源产物，填写现有三字段答卷并保存一次；沿用全局五态，不新增协议。
已把 r116 的五个隔离边界探针纳入正式测试；历史审查为基线 23 项通过、边界 4 失败/1 通过，
修后 r121 的 60 个定向用例与来源/接收两包完整回归 14 项通过。after-r29 与回归源码完全一致，
全新安装后真实 muse 15 轮、82.015 秒完成同代返工 S4/e7，补齐 TASK-001 的 `go.mod` 输出并
提交，原 TR-002 `evt-tr-002-r39` 到 S5/e8；一次保存、无手工系统步骤。两次原保存重放零写入，
产物/来源/evidence/Runtime 一致，`tasks check` 通过。前置为原消费者机械重放的历史 finding
与已完成专业答案，不计真实新判断。S4 范围通过；`validate/doctor` 的旧 S5 manifest 失败保留，
全项目验收 false，新 S5 独立评审未执行。

### 前次执行记录：S3 返工与重新验收（2026-10-16）

用户已要求开始 S3 返工。Blueprint 先固定冻结时序、原接续和当前 REQ 批次边界，
实现与 Skill/协议同步完成，见 [S3 机制审查](S3-MECHANISM-REVIEW.md)。r115 的
89 个定向顶层用例通过，fileview/submission 两包完整回归 14 项通过；前述核心包
回归仍按各自冻结源记录，覆盖有重叠，不相加。
after-r27 的真实 Main S2/S3/S4 和两独立 S5 reviewer 均通过，原 TR-003 进入 S6，
未开展 S6 工作。该轮没有 finding，未覆盖返工。另用新安装机械重放历史 finding
作为明确 fixture，真实 muse 暴露 Git 提交视图提示不足：回滚修复后虽过门，业务问题
未修，不能计返工完成。已补准确分支/路径/提交动作和原答卷纠错提示；after-r28
全新安装后真实 muse 保留并提交 BE 修复，原 PTR-PLAN-02 登记新指纹并到 S4，索引未变。
两次原保存重放无写入或重复推进，S4 listener 已给出携带新 BE 指纹的准确问卷路径。
全项目 validate/doctor 仍因旧 S5 entry4 review manifest 引用旧 BE 指纹而 exit 1；旧清单
退役与重建接续尚未完成，未刷新其指纹，也未把全局检查记通过。原失败、双 PASS 和历史测试
输入分别保留，不改写结论，不把机械前置计作本轮真实独立判断。

### 前次执行记录：S5 审查与装包复测（2026-10-16）

按用户逐阶段编排，本次只处理 S5 的问卷、引导、保存接收、缺项提醒、重放恢复和
原出口。工作状态沿用 AGENTS 的 pending/active/blocked/done/cancelled；专业结论
和阶段游标分别记录。本次不启动 S6，也不接手测试题材的规划返工。

- 已核对普通 S5 的两项独立责任；每人只交一份专业 conclusion/findings，原 Writer
  生成身份、subjects、信封和回执。修正文案中的 activation、缺卷转 legacy、重复
  Markdown 报告和手填元数据指引；legacy 仅限明确未启用问卷的旧实例。
- 在既有两道问题中明确原专业检查项和专项判断条件，不增加答卷字段或工作状态。
- r105 定向回归发现阶段已流转后保存重放丢失原回执；已补原已接收路径的 owner
  解析，继续复用原 Save，不另建历史交卷消费者。原接续已经处理时直接返回原回执，
  避免重复 control 及虚假的 system_followup 提示；解析失败给明确错误，不静默丢失。
- 冻结源 r106 的相关定向回归通过 58 个顶层用例（229.642 秒、exit 0），覆盖三种
  结论、原 gate、独立性、冻结来源、重放、并发、修订及中断恢复。r107 仅再明确
  问卷说明中普通 plan_checkpoint 与显式审批的边界；workpacket/workstate 两包
  整包回归通过（4.790 秒、exit 0）。r106→r107 的源码差异仅为该题干说明。
  失败日志保留；该轮仅为本地源码回归，后续装包证据见下条。
- 装包复测接续：after-r24 构建成功，但安装被 Skill description 的 `Use when`
  规范拦下；修正描述后 after-r25 全新安装、原 REQ bind 与真实 S2–S4 前置均成功。
  after-r25 原 S5 答卷由真实 reviewer 交付，原 TR-004 `evt-tr-004-r38` 正确返工；
  该轮 Spec 在来源被替换后停止，没有被记作双 PASS。两次安装后原 Hook 重放均返回
  同一已接收回执，Runtime/journal/答卷字节不变，也没有虚假的 system_followup。
  `validate --all` 与 `doctor` 均 exit 0。模型服务首次 muse 400、恢复 503 的失败记录、
  gpt-6-luna 的同会话恢复及人工停止过期进程分别保留；用户重启中转站后，muse 最小
  调用成功，随后 Muse 已依据原 finding 修测试题材并重新交付 S2/S3；没有进入 S6。
- 完整重审尚未完成：Muse 的 S3 entry6 新问卷已由原 consumer 接收，新 BE-001
  evidence 指纹与磁盘一致，但 Runtime documents 仍保留旧指纹。S3 原门的
  `evaluatePlanningArtifact` 只要找到一份有效锁定合同（本例为 CONTRACTS-001），
  就跳过磁盘补充，漏掉已变更的 BE-001；PTR-PLAN-02 的注册 action 又须先过门才能
  更新旧指纹，形成闭环缺口。已只读定位为门前文档投影不完整，非作者漏填或旧 entry
  抢占；尚未修改 S3 实现。报告 `r25-s5-muse-rework.process.json`，最终原游标为
  planning.contracts / entry6，不能记 S5 双 PASS 或完整返工重审通过。
- 附加工具缺口：创建独立工作目录后，`release-graph validate --installed` 会遍历
  `.worktrees/*/.claude/inputs` 的局部冻结文档，产生 14 条相对链接缺失诊断。
  原安装校验成功；这次附加检查未通过，不能借用 `validate --all` 的通过结果替代。
  此项单独记录，尚未修改该工具的扫描范围。
- after-r23 第二轮真实双评审完成：准确路径均由 Hook 给出，Write 与脚本保存均收到
  原回执，Stop 均正常结束。规格结论 pass；任务结论 fix_required，指出 TASK-002 的
  output paths 漏列已有交付物 `cmd/server/main.go`。原 TR-004 `evt-tr-004-r78`
  退回 planning.design；没有 TR-003 或 TR-005。这证明发现问题后的原路由工作，
  不代表 S5 双 PASS，也不证明尚未安装的后续源码已通过真实模型验证。

本轮真实审计：本机评估目录 `reports/s5-review-r23-attempt-2-audit.json`；源包为
after-r23，原始模型日志和旧答卷保留。本地回归报告为
`reports/s5-review-r106-final.json`、`reports/s5-review-r107-wording-and-states.json`，
各自附有冻结源码清单与命令日志；未把 no tests to run 算作行为覆盖。

| 阶段 | 梳理/设计 | 落地对账 | 本轮回归 | 本轮宿主验证 | 当前重点 |
| --- | --- | --- | --- | --- | --- |
| S0 需求漏斗 | done | active | active | pending | 原 REQ 模板发题、保留草稿和未经人决策不锁定的定向测试通过，等合并候选复跑 |
| S1 绑定 | done | active | active | pending | 实际绑定申请路径和人类选择已接线；保存不自动批准/绑定的测试通过 |
| S2 设计 | done | active | active | done | after-r21 全新安装，真实 Claude 填卷，原消费者推进 S3；12 轮、约 82 秒 |
| S3 合同 | done | done | done | done | r115 定向 89 项通过；after-r28 真实 muse 保留BE修复、原PTR登记、两次重放及S4入口投递通过；旧S5清单全局校验失败另列 |
| S4 任务拆分 | done | done | done | done | r121 定向60项、来源/接收整包14项通过；after-r29真实muse修复TASK、一次保存并原TR-002进S5/e8，两次重放零写入；旧S5 manifest全局失败另列 |
| S5 文档验证 | done | done | done | done | r128定向48项、相关四包102项通过；after-r32真实独立双PASS/原TR-003进S6，六次新旧答卷重放零写入，tasks check/validate/doctor全过；未开展S6 |
| S6 构建 | done | done | done | done | R40 真实 Builder 专业履职；R45 机械前置与真实 checks/Git/ACK/清理，两真实 Main probe 无 deadline/UNKNOWN，正确展示已知失败，唯一 TR-006 进入 S7 且准确问卷可读；未填 S7。R42–R44 和作者窄测失败保留，范围见 S6 审查 |
| S7 验证轮 | done | done | done | done | 仅本轮限定范围：R48 backend Clean；R51 真冷 TR-006/Main Plan 原登记/7问卷准备；R50 synthetic browser bug-v4 原3职责接收→ObservationBatch/TR-008进S8，clean-v2 原3职责接收→CleanRound/TR-009进S10。六段466条直接Hook响应无内部deadline/4秒defer；clean外层720秒截止后同E2E恢复完成、bug编排误约束纠正及旧失败均保留。最新43项不同定向测试PASS（32核心+11原边界）（skip exporter不计）；after-r52构建0，生产Go相对R51无变更，非R52新完整链。未做S8/S10专业工作、非全分支/全平台/全阶段；[详细证据](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r52-s7-closeout-final.json) |
| S8 调查 | done | done | done | done | 仅本轮限定范围：A/B/C/D1 已落地；after-r56 59 项不同顶层定向测试 PASS、0 fail/skip（154.619秒）。真实 Main→一名 native Investigator 本人 supported→原 Case/route 接收→Contract 无授权等待，专业修稿后外部显式仿真人决策经原批准事务到 S9 planning 即停；Runtime rev11/Case r5 contract_approved，无 S9 专业执行。101 条直接 Hook 均 exit0，选定内部 deadline/defer、专业 schema/route 拒收为0。synthetic 上游非新 S7 通过，仿真人决策非真实用户授权；三 schema 单验 PASS，全项目缺 scenario-model 仍失败；非全平台/全分支模型验收。[收尾证据](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r56-s8-closeout-final.json)、[范围说明](S8-MECHANISM-REVIEW.md) |
| S9 修复 | done | done | done | done | 限定本轮：r86/r79单次CLI无人工续接，1624.111264s/exit0/非超时，rev31 repair.closed→freshS7 round2；本人2Plan/2Final、Main真实集成/检查、Impact、独立5项观察、Handoff已接收。r85/r78亦闭环但3次人工续接、累计1706.755652s，非性能对照。r86旧continuity门曾触发、由本人补后接收；r80移除此门仅经原lead全尾机械例及独立install/原case/代码核查，不冒充r80全链nativePASS；synthetic上游非真人S7/S8，freshS7非S7PASS，无新同版全矩阵资格。[范围与证据](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/s9-blueprint-landing-20261018/s9-closure-summary.md) |
| S10 验收 | done | active | active | pending | 原 RecordEvidence、派生 metrics/审计汇总及完整 S10→S11 权威链通过定向回归 |
| S11 发布闸 | done | active | active | pending | 真实 S10 前置链、准确申请/证据路径及 Agent 自批不推进测试通过 |

S9 当前依据仅为 [review 顶部职责/迁移表及最新证据](S9-MECHANISM-REVIEW.md)；下文旧 S9 r号记录与当时安排为历史，不是当前设计或下一轮执行计划。真实必要断言未复验须具体提醒且不推进；删除ID拼写/严格数量/extra项与注册义务证明，不取消独立复验。

`done` 的设计列表示已完成本轮最小工作流梳理，不代表代码或宿主验收通过。
回归列仍为 `active` 的阶段仅有部分定向用例通过，不能算本阶段完整验收。
宿主列的 `done` 仅指各行列明的真实填卷、原消费者/转移及阶段范围证据；S3/S4 的返工
前置明确使用历史 finding fixture，后续本阶段作者为真实 muse。它们不代替全仓回归、已记录
失败的全项目检查或整条 S2–S10 链验收。

## 当前回归证据与缺口

- r63：S0/S1 三个真实 Hook 定向测试通过；来源为隔离快照 `all-stage-entry-r63`。
- r64：S3/S4 原领域接收、恢复及错误答案等定向测试通过；来源为同一快照。
- r65 首次 Main S7 正向测试暴露问卷与生成提醒被当作产品漂移。已修原漂移分类：
  问卷目录属于控制产物，已声明的 frozen subject 仍逐项校验；保留旧轮答案不应使新轮失效。
- r67：S0/S1、Main S7 原登记/重放、S10 清单/保存/重放/缺结论拒收等子测试通过；
  整组仍失败于三项旧提示断言及 S10 测试样本漏掉权威 Claim。正在修正并复跑，未记整组通过。
- r72：S6 原 Git 集成、安装产物投影及三处清理中断恢复的定向回归通过；
  S0/S1、Main S7 保存/登记/重放、默认启用问卷和窄范围漂移分类回归通过。
- r73：S10 实际 Hook 保存清单、原 RecordEvidence 接收及重放测试通过；
  S11 前置样本仍失败。已修正样本分支和证据分母，并补 `s10` 原磁盘来源，等待复跑。
- r74：核心包回归暴露两处编译错误，均已修复但尚待复跑；acceptance、workpacket 包通过。
  workspace 完整测试另受本机 `bwrap` 无法创建 uid map 限制，不能记真实隔离检查通过。
- after-r19 首次合并构建暴露 S9 接收代码三处编译错误，已修复。r77 核心回归中
  submission、hook、hookctx、acceptance、workpacket、schema 六包通过；
  S7 四个测试的问卷形状/流程预期与一处 S9 helper 返回签名仍使整组失败，已修正待复跑。
  workspace 宿主 uid map 限制复现；未将跳过或失败的真实隔离检查计为通过。
- S9 对账收回一次会增加负担的适配器方案：原契约为每 Assignment 一份 RepairResult、
  每 Contract 一份完整 TargetedReverification，不为所有 Assignment 各自强制重复复验。
  当前修复同时处理同 Runtime/generation 再次进入 S9 时的原 Session/Plan 命名和问卷来源碰撞。
- after-r20 构建发现一处 S9 重复声明，已修复；r78 编译检查又发现新增 CLI 测试的
  fixture 循环导入和 S8 Worker 测试两处误用字段，均已修复。r79 的五处新增测试
  编译问题也已修复；该轮完整行为回归因用户暂停而中止，未计通过。
  编译检查中标注 no tests to run 的包不能记行为验收通过。
- r78 的 integration 整包回归通过（158.826 秒）；review 整包仍失败于一个新增 Worker
  用例。该用例暴露原 Runtime checkpoint 被误算产品漂移，以及样本临时 canonical Result
  未清理；已分别补原控制产物分类、清理样本文件，待复跑，未记 review 整包通过。
- S9 环境阻塞恢复：原领域已补仅 PASS/scope PASS 能关闭 required ID，并增加同 ID、
  原恢复时间派生存储版本的创建入口。问卷已接原 Resume 消费者；环境阻塞留在 S9
  targeted_reverification，因果/范围失败才回 S8；完整恢复测试仍待通过。
- r80 全仓回归中，integration 整包通过（115.940 秒），review 整包通过（16.631 秒）。
  整组未通过：CLI 被旧快照测试的格式串检查阻断；S8 样本未提交原冻结输入；
  S9 PlanReport 从错误的信封层级解码出空 generic PLAN，blocker 仍强制 Hook
  身份回调；部分测试仍预期旧身份拒收。上述代码/样本已修，待合并候选复跑。
  migration/semantic 的旧 AGENTS 机制词项契约也已对齐准确问卷路径、统一责任状态
  和原人类决定。repair 包另外暴露恢复样本和领域接线问题，正在处理，不计通过。
- r82：S9 PlanReport 四项原链测试和 RepairResult→Worker 候选→Main 集成完整用例通过；
  migration/semantic 定向回归通过。S8 用例错误地把 packet 相对路径作为实际 Worker
  file_path，已修正样本；两个 repair 用例暴露非法 scope 样本和 Handoff 引用校验缺口。
  原引用集合校验误复用了会过滤控制目录的产品差异比较器，已改为完整 path/hash 集合
  比较，保留未登记 Targeted 必须拒收的负例，等待复跑。
- r83：S5 保存/返工/人类需求决定、TR-003 中断恢复、S4 正式输入→S5 原准备、
  S6 原生 PLAN/Result/修订回执、Main S7 计划登记/重放通过。S9 新测试前置 Case
  指针不完整，已改由原 Ingest/Case/Contract approval 链生成。S10 原 Hook 保存三种
  工具用例通过；发布审计 draft 漏了原 audit_area coverage 行，已补原八个问题与反证，
  保存时从同一专业答案投影旧汇总字段；S10→S11 链仍待复跑，整组未计通过。
- r84：CLI 49 个定向主用例中 45 个通过，含 S0/S1、S2–S4 原跟进、S5 无重复
  PLAN 保存、S6、S7 Main 计划、S10→S11 完整链及无授权不推进。剩余 S5 Stop
  可见缺项已修；S9 的 Case 前置样本改走原假说注册/支持结果/路由/审批流程。
  repair、migration、semantic 整包通过；investigation 350.343 秒，仅一处旧身份
  拒收断言失败，已换成未登记路径与零写入负例，未把整包记通过。
- r86：S9 直接专业 RepairResult 保存→自动原 completion→Worker 候选→Main 集成
  →最终回执/重放通过；Main resolution 原接收、Main SessionStart 自动发继任问卷
  及旧答卷不复活通过（investigation 三项共 82.136 秒）。S5 Stop 的可见路径和
  实际 JSON 缺项提醒通过。S9 Impact 问卷自污染产品 diff 的真实问题已修复。
- r87：机器控制目录不进入 Session capture/diff、未知同级文件仍校验的测试通过，
  原 blocked→Resume→同 ID PASS→Handoff 链通过。r89：S9 两 Assignment、一 Contract
  独立复验的问卷完整链通过，包括阻塞恢复、旧失败证据保留和自动交接新 S7。
  已移除普通 S5 准备/launch 输出中的重复 PLAN/begin/consume；整包旧 ACL 测试
  契约已同步，尚待最终整组执行。以上均非最新包真实 Claude 的全阶段完成证明。
- after-r21：完整冻结候选已构建为 Linux amd64 评估包并通过包内校验；全新安装默认
  启用问卷，r91/r92/r93 的真实 Claude 分别完成 S2/S3/S4，原消费者实际推进至 S5。
  三次均收到准确问卷路径，均未手工 issue/consume；原始报告为
  `all-stage-r91-s2-model.json`、`all-stage-r92-s3-model.json`、`all-stage-r93-s4-model.json`。
  r93b 两独立评审分别用约 149/178 秒完成 S5，原 TR-003 `evt-tr-003-r36` 推进 S6；
  无手工 issue/consume 或通用 PLAN/begin。当前真实链计 4/9（S2–S10），
  其余阶段不能借用历史证据计完成。
- r90：同一冻结合并源的全仓回归已结束（1436 秒），65 包通过，CLI、workspace、
  系统 req039 三包未通过。investigation 436.675 秒、integration 120.827 秒整包通过；
  CLI 的旧提示断言、S5 修订资格和 S7 样本版本已修；系统测试漏 import 已补齐。
  workspace 唯一失败为宿主 `bwrap: setting up uid map: Permission denied`，未计隔离通过。
  S5 修订入口已对齐首次答卷的审批模式，普通问卷不再要求旧通用 PLAN 激活；
  显式审批模式仍使用原授权，并保留当前责任、冻结输入、原答卷及独立执行源校验。
  此生产修复晚于 after-r21，不能声称已由该包实测。
- r95：S5 修订审批模式边界、真实 1.3 激活草稿分类与系统 shared-model Hook 用例
  通过；CLI 因测试引用了不存在的解析 helper 编译失败，已改成就地解析正式
  QUALITY_GATE JSON，等待复跑，不计 CLI 行为测试通过。
- r94 的 S6 实测未通过：原注册命令的帮助隐藏了真实参数，包内又没有明确可读的
  S6 Manifest 示例入口；旧生成手册/动作目录继续指导重复手工提交。Main 随后搜索了
  项目外评估资料，已停止该会话并保留日志，不给 S6 阶段完成证明。普通 S5 注册
  产生的旧激活草稿还使 `validate --all` 失败；这些真实可用性问题在修复，修后需
  新冻结、全新安装复测，不能覆盖现有活跃实例或用提示预填代替产品修复。
- 真实 Claude S3 前两次在工作前返回额度/并发 503，原失败记录保留。用户恢复额度后，
  r75 在 after-r18 包完成 S3（12 轮、约 90 秒），原接收推进至 S4；
  r76 完成 S4（19 轮、约 112 秒），原 TR-002 提交 `evt-tr-002-r23`，实际进入 S5。
  两次均未手工 issue/consume；S4 额外运行一次返回空结果的 recover，作为冗余指导继续排查。

原始日志、逐次命令和退出码在本机隔离评估目录 `claude-closure-20261015T011245Z/reports/`
与 `logs/`。所有失败记录保留；最终合并候选需重新冻结、构建并验证，不能拼接快照冒充整体验收。

## 执行安排

当前按用户编排只处理 S4 返工及 S5 出口；并行修改原接续、原门/事务和文案，
实际 Go 回归与构建由主会话串行运行。
真实 Claude 测试在隔离安装和真实原入口上进行；每次保存原始输出、答案、接收事实
与原转移，不根据模型自述或进程 exit 0 推算通过。发现缺口后修复并复跑当前阶段，
再验证后续阶段；不靠改 Runtime、预填专业答案或替换历史卷跳过工序。
