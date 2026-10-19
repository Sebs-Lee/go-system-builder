# 全局优化落地设计

> 定位：独立于 L1–L4 的全局实施设计；承接 [全局优化指南](GLOBAL-OPTIMIZATION-GUIDELINES.md)。
>
> 状态：设计v0.1.0的可行性审查已完成，按用户授权进入实施；逐切片状态见§7。G01–G08、O01–O10、M01–M10 均提供具体方案和验收。本文出现的新包、字段、API、Schema、命令和测试均是拟新增项，除明确列为“现有”者外，不表示已经实现。
>
> 调查基线：2026-10-13 的仓库代码、协议和模板；初次调查交付为设计及可行性审查。后续实施先固定L层，再修改文档/代码；Hook 权限拦截按用户最新 §0 决策取消。

## 0. 收缩后的实施决策（2026-10-18，优先于旧方案）

<a id="当前优先实施契约2026-10-19本轮重验后固定尚未落地"></a>
### 本轮重验后固定契约（2026-10-19，设计固定／实施待验证）

本段优先于以下历史条款，保留所有原失败和证据；不更新全局通过。r96 是新 Runtime 复用真实设计的接续验收，非从零 S0；三路 S2/S3 已原接收，A/B/C 均在 S4 保存遇到预算失败且无 S4 持久接收记录，S4 原生接续仍待 r97，非全局 PASS；独立 Workers、后续职责与真实产品 oracle 尚未完整覆盖。限定修复如下，不新增字段、preflight、状态、proof、队列或更严准入：

1. **条件资料按实际适用关系读取。** 以既有 Markdown 实际链接目标/现有 JSON 实际引用结构确定正向条件依赖；普通裸路径、inline-code 提及或围栏例子不单独声明依赖，不用否定词典。真实链接目标缺失照原错误，不按文件存在性吞缺项。保留原已有实际适用输入及真实缺项；同步 architecture-template 现有 model 段为条件性真实链接指引，不保默认强制 model 路径；旧 fixture 裸正向路径改正常链接，明确不作逐字兼容。无需补造不适用 model.md，沿既有原成功出口及监听接续交准确当前卷。
2. **CASE 集合读取实际身份。** 生成物 CASE 集合取实际 id 字段，不全文扫描 FACT/BR/FX 子串充作 CASE；保留合法完整复合 CASE、真实合同引用及必要覆盖，作者真实简写错误仍提醒。
3. **AC 混合引用逐项保义。** FR/NFR/A4 混合时存在 FR 就逐 FR 保留真实可达关系，补充 NFR/A4 不吞掉 FR；无 FR 才按原 N/A 语义处理。保现有 FR/NFR 缺项检查，不新增 A4 存在性资格。
4. **amend 人意与入口相连。** 已明确批准的具体修订复用原决定，经既有准确参数入口与原 amendment 接收交现卷；manual 权威段源自 internal/transition/manual.go 的原 emitter（打包由 binary manual 生成），与 requirement-funnel 现有 amend 段相连，protocol 必要时仅链接该段。不索逐字命令或重复授权，不添第二清单/审批流程。
5. **完成与准备失败分开呈现。** S2 已完成、S3 准备失败分别记录；已完成答卷不得标作当前待填卷。系统准备失败点明实际系统接续及责任，不一律让作者补材料；真实适用材料缺项仍具体提醒。沿原 pending/active/blocked/done/cancelled，不造状态层，也不据投影差异断言模型误判。

**S2 本片减负边界：** 保五手写源与两派生载体及实际 allow/reject 语义覆盖；正负数量比改风险提示，不作硬门或凑数。统计只引用生成物，不手抄多份；实现者/测试者/维护者三视角合为一段真实语义自审，不另把 direct AC 或数量审计作证明。保留真实语义自审，删除无依据证明另审；不扩展全资料格式重构或新字段/预检。保留 persisted_effects 字段，真实无持久变化可填 []，不靠文字凑非空；真实效果由专业判断承担。非空校验与序列化已同片对齐；[r96 原生报告](/home/sebs/.cache/gsb-gap-closure-20261019T034632Z/R96-REVALIDATION-AND-GAPS.md)记录 B 真实 28 个空集合通过，C 原生 allow/reject 为 12/11、required allow/reject 23/23 并经 S2 原接收（[C 原报告](/home/sebs/.cache/gsb-gap-closure-20261019T034632Z/native-c-todo/CASE-REPORT.md)）。这些局部事实不代表 S4 原生或全局完成；S4 仍待 r97 验证，不预写尚在执行的定向机械重跑全部通过。

**S4 本片固定、实施待验证：** [L3-S4 当前保存契约](l3-stages/L3-S4-task-split.md#s4-current-save-contract-20261019)与[S4 机制审查](global-optimization/S4-MECHANISM-REVIEW.md#s4-current-save-contract-20261019)限定一次专业检查、Writer 一次原子幂等记录、后续只投影原事实并由 TR-002 交独立 S5；禁止跨 Save/prepare/Authorize/followup/receipt 重复来源与身份 proof，非阻塞 advisory 不索 checker 格式证明。原生 A/B 8.113/8.129s 与一次外部 198 Git 诊断分别留证，不归并为 PASS 或纯 Git 主因。

**S3 原误伤小修约束：** shared-model reason 仅要求真实专业说明非空，不增身份或来源资格。真实 JSON 的花括号不作为占位符拒收；保真实缺项，不添新识别规则或证明义务。

限定验证先复现各原错并核原接收/当前卷；随后同版新装原生补未验职责与产品 oracle，才可更新全局验收。原失败、诊断、机械限定结果及原生业务验证分别保留。

### 当前优先实施契约与限定验收（2026-10-19）

本轮小片已实施并完成限定机械回归。r94 原生历史保留：A 到 S4 双卷，C 到 S3 空卷，B 首次 S3 接收仍因 8s Hook deadline 失败；插桩诊断不算 native PASS。r95 fresh 安装仅机械重放已完成的 S1/S2 原 consumer 前缀，随后 B 唯一 native S3 段一次 Edit 正常接收并实际 Read S4 view/answer，仍为空稿；原 Post 附件 8003ms、stderr 空，这是整条 Hook 平台计时，不证明内部超时或性能余量/根治。完整链仍 **0/3**，S4 实际履职、S5–S11、独立 Worker 与产品 oracle 未验。详见[本轮限定报告](/home/sebs/.cache/gsb-global-optimization-20261019T002858Z/R95-ROUND-REPORT.md)。本次三处顶部状态为打包后文档结果更新，r95 二进制及冻结源 manifest 未改，不声称当前全部文档仍逐字等同冻结快照。

本段优先于下方历史落地结论及旧实现保留条款。以下 r91 为原完整链历史基线：本轮同 r91 包三题材 fresh S0–S11 重验 **完整链 0/3**：A 完成 S0–S2、S3 未接收；B/C 原 amendment 后缺当前 S2 卷，B 明示诊断恢复后也未完成 S3；S4–S11 未获本轮原生职责/产品 oracle 验收。旧局部通过不升级为本轮全局通过。证据见[本轮完整报告](/home/sebs/.cache/gsb-global-revalidation-20261018T143345Z/GLOBAL-REVALIDATION-REPORT.md)、[解析审查](/home/sebs/.cache/gsb-global-revalidation-20261018T143345Z/parser-code-review.md)及[预算调用链审查](/home/sebs/.cache/gsb-global-revalidation-20261018T143345Z/case-a-ticket-cli/timeout-code-review.md)。

本轮先按以下边界解决已证阻塞，逐小片审查后实施，不顺带重写专业政策：

合同索引只读取“合同清单”当前实际表的表头、分隔行与连续数据行，表结束后不把适用性或其它专业表当库存。原问卷 `applicability=not_applicable` 的展示行不要求虚构合同文档、locked 状态或版本，也不进入实际交付集合；required 合同的真实文档、状态版本与必要覆盖仍检查。类型展示文字不作精确枚举准入，按原合同 ID family、专业适用性及真实产物读取，不增同义词分类器或字段。

S3 `planning_contract` 原 consumer 在同次接收中一次解析实际交付并生成机器 envelope；CAS 保当前 Runtime/generation/entry/REQ 关联、原 professional semantic 回调及原子提交，删除 CAS 重建已登记发行闭包、重复 `ContractArtifacts` 与 expected 整信封重建。保现存 `Files.Verify`、原 receipt 同答幂等、真实缺项与输出内容检查；机器失败的未接收记录按本次实际生成字节自动形成，不以旧同答 metadata 冲突让作者猜测，原已接收历史不覆盖。后置只读取本次已接收事实，不新增字段、状态、cache、API 或预算；不声称 sharedmodel 全流程只检查一次，transition 原全集 checker 仍保留。r94 限定证据：A 7.590s 到 S4 准确卷，B 8s 仍未接收；solo 诊断 6.655s 已到 S4 cursor，但 issuer 遇 6s deadline 未给卷，诊断不算 fresh 验收；r95 后续限定 S3 接收与 S4 双卷读取见顶部当前记录，不扩称性能根治或完整链通过。

原S2履职结论与文档记录保留为历史事实；S3尚未发行时，按这些已登记design的身份、路径和版本读取当前已提交专业材料，由系统带出本次hash，历史design SHA不作为当前输入资格。真实REQ绑定、当前generation/entry、实际文件与必要覆盖、原Writer/CAS职责保留，不要求作者回滚说明、刷新登记或重新履职。后置门读取同次已验证的当前资料事实，不回拿S2历史hash作为S3资格。

1. **原成功出口交当前责任。** amendment、恢复、返工等原成功入口复用原下一卷 issuer，一次准备真实当前卷并给准确路径；保留旧卷/已接收记录，不让作者手动 issue/consume/recover、猜 generation 目录。人的明确业务选择直接承接必要原事务，姓名仅记录，不要求逐字命令或同一决定的额外执行授权；真实未作的人意仍等待。
2. **监听与实际变化分开。** S3/S4 普通 Pre/Stop 只读当前责任/已记录进度与真实缺项，不重算接收资格。普通只读工具不反复接收未变答案；准确当前答案保存、相关专业资料或真实 Git 输出确有变化时，仍自动交原 consumer 接续，不因工具叫 Bash/MCP 或无 file_path 漏接。沿已有当前答案/资料/输出事实定位，不新建缓存、receipt、状态或变化证明协议；同答已记而原出口未完仍复用原接续，跨 stage 回放不重推。
3. **读取真实专业关联。** AC 可实际关联多个既有 FR，逐项保留必要可达关系；只读当前有效验收表，不把 fenced code、历史引用/样例中的旧 AC 当活义务。cross-matrix 允许同 fact×FR×story 下多个实际 branch，不迫作者为载体唯一键改故事归属。S3 正确读取 REQ 附注、自然 Markdown 合同链接及明确 CASE 范围，不能靠全篇偶然 token 代替当前覆盖。已有登记样例与 Schema 定义分类消费，样例不因未属于 Schema 集合被判为新 Schema；缺真实定义/样例或必要覆盖仍具体提醒，不凭自报 pass 跳过。
4. **机器关联不转作者。** 原系统已知指纹/来源/登记关联由系统使用，清除模板人工组合或复制机器 fingerprint 的义务。专业问卷不新增 metadata、proof、preflight；反馈指明实际材料/条目与下一责任，不要求作者靠猜 parser 格式换接收。
5. **性能只作有界定位。** 已证预算失败不等于已定位内部慢点；本次函数回放已测得写前与 CAS 内重复 prepare，其他未采样函数的耗时仍未知。仅允许原入口小诊断区分 worker、Pre 取材/QG、Save 前校验/CAS/followup，不增加时限、重试层或资格门。S2 负例比例配额、泛化双检/三角色方法及资料大重构只列后续审查，当前不一并修改专业政策。

真实专业工作、完整必要观察/覆盖、实际 Git 与检查、独立 S5/S7 等职责和必要人意均保留，责任状态仍只有 `pending/active/blocked/done/cancelled`。先用本轮原错误作限定回归与实际接收/交接核对；随后补未覆盖职责、以明确前置边界复验，再同版三新装题材闭环及产品 oracle，才能更新全局验收结果。小片实施与限定机械回归不代表全局验收已完成。

### 目标与唯一当前规范

Hook 帮助 Agent 按图完成工序：提前说清本人责任、资料、专业问题、答案形状、准确问卷路径和真实缺项。正常流转为 **展示要求和问卷 → 本人执行并填写 → 保存记录事实 → 提醒真实缺项或推进**。

当前规范以本节、[S0–S2 当前契约与迁移](global-optimization/S0-S2-MECHANISM-REVIEW.md)、[L3-S9 当前规则](l3-stages/L3-S9-repair.md#当前最小工作流2026-10-17重新固定减负设计)及 [L3-S10 当前最小契约](l3-stages/L3-S10-acceptance-audit.md#当前最小契约2026-10-18设计已固定尚未实施)及 [L3-S11 当前最小契约](l3-stages/L3-S11-release-gate.md#当前最小契约2026-10-18)为准；具体逐入口职责见该规则与 [S9 机制审查](global-optimization/S9-MECHANISM-REVIEW.md)。下文含 §7 的旧门槛与历史实现不构成保留授权；文档固定契约不表示生产和宿主链已完成迁移。

### 作者做专业工作，系统带出已知事实

问卷复用本人当前任务与已有 Draft，只问专业判断、实际结果、证据位置和未完成原因，给出必要材料、问题与明确槽位。Plan/Result 作者 view 不附完整 packet/inputs 镜像。JSON 可继续使用，但不要求作者抄身份、hash、revision、发行事件、登记数量或手工信封，不另交相同内容的总结，不手动 issue/consume。

必要断言由系统提供完整问题和槽位，独立复验者真实执行所有必要断言并报告实际观察。断言 ID、数量及注册证明不另作作者的资格门；缺真实观察须明确提醒，不推进完成。

旧兼容元数据 `continuity_reason` 与 `scope_compliance` 可选；未填不拒收、不制造 scope fail，也不作为 Handoff 条件，不得为换取接收而编造 pass 或解释。全部必要断言的真实观察、结果、实际证据及未完成原因仍须如实记录，缺真实结果不推进。

### Hook 读事实，保存只记录一次

Pre、Stop 与普通重复监听只读当前责任、已有答案和工作事实，给准确路径、普通操作命令及具体缺项，不全扫描或重跑业务。Main 在正常责任开始时通过普通工具完成 workspace bind/prepare；原 prepare 提供真实工作位置和 Result 卷。分工或位置选择保存只反馈系统已知参数的准确原命令，不在 Hook 中创建工作区、clone、Bootstrap、运行自检或完成 Main 集成。ready 但缺卷须用同一原 prepare 收尾，不另发准备卷或手动 issue。

保存仅做：**准确定位当前问卷与责任 → 一次必要专业结构校验 → 一次原 Writer 原子幂等记录 → 按真实缺项反馈或推进**。支持普通编辑、脚本和 MCP 保存。并发、同答重放、真实内容冲突、必要存储结构和 CAS 均留在内部原 Writer 边界，不由每层反复 Snapshot、packet schema/Rebuild、journal 发行核验、hash 或 source 证明来重演。

Writer 必要存储结构只服务本次记录的可读性、当前责任关联和并发原子一致性，不能借“校验整个 Runtime”逐卷重建或重验所有历史业务对象与来源链。一次逻辑提交可含必要的多次物理 I/O，不要求存储算法只调用一次 Write。兼容或诊断入口不得成为正常接收依赖；真实损坏走既有存储维护，不在每次保存前预防性运行整条 proof。

身份/权限、scope ACL、发行、来源 hash/receipt/coverage、Bootstrap、lineage 等资格从正常接收整体删除，不能搬到 consumer、helper 或 Handoff。“领域完整性”“原条件”“保持原门”不能作为兜底保留理由。读取已接收事实只投影当前责任和实际进度，不重新证明历史接收合法。

### 真实完成责任与推进边界

S9 消费 S8 已批准 Contract，不重新审批或增加 readback/启动许可。保留真实必要用户决策；删除 generic PLAN/Begin/activation/work_started/completion 授权链。本人 Plan 在本人实施前记录，不等待全员计划；Main 安排真实依赖。工作状态只用 `pending/active/blocked/done/cancelled`。

作者在真实工作区实施或测试并自检，普通 Git 记录实际源码改动；无改动测试不造 diff/commit，无迁移不造引用。前检如实记录 pass/fail/blocked，命令只表示已做工作，不冻结为 Main 未来命令。Main 按实际目标和改动完成真实 Git 集成、必要检查及输出记录；未做的检查不能记 PASS。

Plan/Result 的 fail、blocked、范围偏差可如实记录，不因 scope ACL 拒收。接收成功不等于修复或 stage 完成；缺少真实结果时给出具体剩余责任，不推进。系统从已接收 Final Result 批次按原 last-wins 直接带出实际变更清单，供 Impact 与自动 Handoff 共用；删除 Impact/Handoff 的 exact 集合拒收及物理目录/缓存重扫来源证明，不只把两套来源对齐到通过。Main 填专业影响判断，不手改机器清单。

完整工序：Main 明确批准目标与分工 → 本人 Plan → 实施或测试、自检及本人 Result → 必要 Main 集成与真实检查 → 专业 Impact → 未参与对应实现与合并的独立复验者真实复验全部必要断言 → 正常图完成后自动交接 fresh S7。Handoff 无额外答卷或来源资格层。

不新增模型、作者可见 API、字段/格式、缓存、协议、后台队列、状态层、重试或预算方案。验收看真实工作能否完成、缺项能否准确反馈；Hook exit0、机械 PASS 或历史上游不能替代真实 S9/独立 Targeted/fresh S7 验收。

S10减负三片及作者文案已落地，保原S7当前clean事实与S11真人决定；限定原六函数经r10/r11分段机械验证及r85真实Main到S11待真人（一次人工continue），不是全局或发布完成。[S10-LANDING-REPORT.md](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/s10-reduction-20261018/S10-LANDING-REPORT.md)区分机械/linux-only安装、r84失败与r85实测；不以metrics/fullvalidator代专业判断，不扩S11。

S11本片正常保存自动消费与原退回责任交接已完成限定实施/回归/宿主scope：r88return退审补审再S11、最终r89defect交原S8问题包且监听无false-stale；分轮证据与真实边界见[S11-LANDING-REPORT.md](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/s11-reduction-20261018/S11-LANDING-REPORT.md)，非最终同版全矩阵、Investigator执行/完整S8、真实发布或全局完成。

S0–S2新返工契约已固定：原REQ承载专业判断与真实锁定，不新S0收卷；S1原四问题保存自动TR001，S2原两字段保存交原记录/出口，已知事实系统供给、普通Pre/Stop只读，不索重复身份/来源证明或正常命令手续。实际Git对象与专业设计工作保留；本次实施/限定回归done；r91有限宿主范围通过：MiniMax自然exit0、无外部timeout，639.243807714025s，原四题正常保存绑定develop/main准确目标，S2两字段正常保存经PTR-PLAN-01到rev6/planning.contracts/entry2并备S3卷；cross-matrix无null且未借读包外schema。[r91宿主审计](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/s0-s2-reduction-20261018/native-s012-r91-minimax-audit.md)。本次复用8份真实专业源，非从零全S0创作、全局完成、真实发布或性能提升；S0仅复用历史草稿/未授权等待与真实锁定交接边界及本轮机械证据，不泛化更宽业务。原20函数/7包编译构建与host-only包安装证据见[落地记录](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/s0-s2-reduction-20261018/S0-S2-LANDING-REPORT.md)。r3构建及原19函数通过、r90 linux/amd64包空安装清单0差异。[限定验证与包安装](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/s0-s2-reduction-20261018/verification-20261018T133235.172892Z/package-install-r90.json)；非全仓/全平台，旧证据不当新通过。见[S0–S2机制审查](global-optimization/S0-S2-MECHANISM-REVIEW.md)。

当前r90宿主暴露两项待验修正：S1原四题的Git引用应逐字填写实际选择（如`main`），括注说明放原自由文本，未知目标留空，不加字段/归一化/资格门；cross-matrix模板可选string的branch/no_branch_reason未用项应省略，不写null，保持原schema/validator。到S3不代表此前错误目标记录已正确，r90场景不回改；两项修正已有r91限定机械证据，最新有限宿主复测已通过，范围见当前记录。

## 当前接续入口（2026-10-18）

**本轮 S9 减负设计、实现与限定回归已完成；r79 包的真实 S9 专业链已交接 fresh S7，r80 末片另有局部核查。** r86/r79 单次 CLI、无人工续接，exit0、非 external timeout，**1624.111264s（27.0685min）**；实际 Runtime **rev31、repair.closed、verification/running**。两本人 Plan、两 candidate/Final、Main 真实 Git 集成与两次10/10检查、专业 Impact、未参与实现/合并的独立者五项真实观察、自动 Handoff 及 round2 seed 均已记录。见[进程](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/s9-blueprint-landing-20261018/r86-native-main.process.json)、[实际报告](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/s9-blueprint-landing-20261018/r86-bike-r79-installed/.claude/evidence/s9-validation-report.md)、[独立终态审计](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/s9-blueprint-landing-20261018/r86-terminal-audit.md)与[收口证据](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/s9-blueprint-landing-20261018/s9-closure-summary.md)。**fresh S7 仅入口，Delivery/QA仍pending、clean_round=null，不是S7 PASS；上游S7/S8及审批是明确synthetic fixture，不冒充真人履职或授权。**

r85/r78也完成同一S9专业链，但四段自然结束、三次人工续接，累计 **1706.755652s**；原独立作者纠正旧元数据后才接收，见[终态审计](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/s9-blueprint-landing-20261018/r85-terminal-audit.md)。r86仍触发旧continuity提示，由独立作者本人补后接收；不是零机制困惑，也未验证后续metadata缺省行为。两轮流程/续接不同，耗时不能作控制性能对比。r79反馈片通过[两原例机械验证](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/s9-blueprint-landing-20261018/validation-r10-summary.json)；r80旧Targeted元数据可选片通过[原lead全尾机械例](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/s9-blueprint-landing-20261018/validation-r11-summary.json)及[独立Claude局部安装/原例/代码核查](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/s9-blueprint-landing-20261018/r80-claude-audit.md)（[Root复核](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/s9-blueprint-landing-20261018/r80-audit-root-review.md)）：安装1.40132s、原例PASS2.16s/wall2.21612s。**r80没有另跑完整原生S9，不把r79模型证据继承为r80全链PASS**，也不新增同版全矩阵资格。所有包仅linux/amd64评估；历史失败与未知性能边界保留。

### 历史证据界线

以下保留各轮原记录及其当时待办，不表示当前运行仍将执行或已通过；旧“最新/待运行/下一次”均仅对应当时快照。代码规范仅以 §0 与链接的 L3 当前规则为准。未删除任何原失败、耗时、包或其他阶段证据。

来源统一小片已冻结并通过限定机械验证：r82 原 lead 完整 case PASS（test 4.10s / 包 4.152s / wall 55.536s），原 last-wins 与 stable Changeset identity 两例 PASS（包 0.066s / wall 55.795s），见[原始日志目录](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/r82-impact-typed-source-s7_oq7wa)。case 使用明确 synthetic 专业数据，不是 native PASS。after-r75 host-only 构建 exit 0、55.109531s（[清单](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/after-r75-source-inventory.json)）；本记录晚于锁定快照。r81 已有真实前半两 Final；下一次在 fresh 同 bike 用真实代码、tests、Git 和原 consumer 建立明确标注合成作者的机械前缀至 Impact，仅由 Claude 验证修复尾链，不冒称单轮全新 native 前半闭环。

r81 限定验证已通过：五个既有例全 PASS（包 **13.226s** / wall **37.070s**，[原始日志](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/r81-hook-prepare-split-e9p3j6ki/test.log)）。保存到 workspace 的嵌套准备及 observer 队列遍历已删除；保存/监听给准确命令，Main 普通工具执行原 prepare，原 prepare 负责实际工作区和 Result 卷。覆盖包括 typed normal observer 的显式启动发卷及两个明确 legacy workspace-only 例，不等于完整 S9/fresh S7 通过；后者待全新 r81 真实运行，也未认定 r74 晚 Result 超时根因已解决。after-r74 host-only 构建 exit 0、24.247160s（[清单](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/after-r74-source-inventory.json)）；本工作树验收记录晚于该锁定快照，r80 失败证据保留。


r80 / after-r73 真实运行已在 **331.010s** 停止，首次 Edit 保存触原 8s Hook timeout，**revision 18 / 0 Result**，实际工作区 ready 但 Result 卷未发；见[最终报告](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r80-bike-fixed-r73-main-final.json)与[耗时分析](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r80-bottlenecks-and-timing.md)。该保存/startup边界修正及限定验证见上文，未称原超时已解决。

最新 r79 / after-r72 真实 bike 运行在 **662.697s** 停止，**0 Result，未闭环**：脚本保存已填写的 `check_location` 未被 observer 接到，作者用 Edit 重存后才 prepare，这是实际机制失败；见[耗时与阻塞分析](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r79-bottlenecks-and-timing.md)及[真实最终报告](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r79-bike-fixed-r72-main-final.json)。observer 修复复用既有 Plan、Result、execution 和原队列纯函数；三个原例最新 PASS（包 12.504s / wall 237.680s），仅覆盖首次脚本 prepare、依赖 queued 及未变重放零写，见[原始日志](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/r80-observer-real-plan-awbs2p7c/test.log)。两个首轮 fixture 失败及删除不合法 legacy 后半段的覆盖边界见[记录](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r80-observer-fixture-boundary-adjustment.md)；真实 Final 后释放和完整 S9 在 r80 仍未验证，不宣称原超时已解决。after-r73 已完成 host-only 构建（201.801s，[源码清单](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/after-r73-source-inventory.json)）；本次工作树文档晚于该锁定快照，仅补验收记录。下文保留此前分析和验证历史。

2026-10-18 分析补充：r75 只读 status 基线 **1.624s**（[分析](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r75-readonly-timing-analysis.md)）；r76 在首责任 Final 后，第二个无改动责任经原生 observer 保存 **2.616s**（[分析](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r76-two-responsibility-save-analysis.md)）；r77 同 case CPU 样本 **2.60s**、profile 窗口 **3.03s**，ValidateBytes 累计 **38.08%**、RepairResultPackets **29.23%**、Compile **18.46%**，Compile 已含在 schema 调用内，禁止相加（[分析](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r77-save-cpu-analysis.md)）。这些样本未确定或修复原 r74 超时根因；生产和 r74 现场未改，S9 仍缺 Impact、独立 Targeted 和 fresh S7。

用户已授权实施以下两个既定 §0 小片；冻结源码的三个原定向机械测试通过（runtime 包 0.860s / repair 0.075s / integration 20.948s）。同新源外部单次诊断在首责任 Final 后，以第二个无改动责任、Bash observer 和原生 8s 入口保存通过：2.216615717s、accepted=true，一次 revision 23→24，HEAD 不变、无新 commit；见[原始验证报告](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r78-frozen-mechanical-and-native-observer.md)。这不是严格 before/after 收益证明，不代表 r74 根因已解决或真实模型 S9 PASS；该轮 after-r72 后续真实验证结果见上文 r79。不增加设计、状态、缓存或预算，保留专业必填、当前责任关联、一次原子幂等记录和真实 Main 检查：

- 普通 RepairResult reader 移除逐包 schema 与 Rebuild；Find 只选唯一目标登记行并解码该卷。
- 候选 Accept 删除重复 Issue/bootstrap/权限回调准入及写后 Snapshot/Verified，直接使用原 Writer 已返回的结果反馈；以下旧证据继续作为各自历史边界。

最新 after-r71 的 r74 bike 真实运行已停止：两段累计 **1222.613s**，均 exit 143、非外部 timeout；原轮因 timer 空闲停止，cont1 出现真实 Hook 超时。已接收 **2 Plan（test 为 `[]`）**，backend 的真实 candidate、Final、Main Git 集成和检查通过；test candidate 最终在 **revision 26** 存在，但没有成功接收回执。首个超时早于该 candidate journal 时间，不能据此确定后置投影就是根因。**Impact 0 / Targeted 0 / 未到 fresh S7**，不标 S9 完成；见[最新真实报告](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r74-bike-fixed-r71-main-final.json)。以下保留此前包与工作树验证历史。

after-r70 的真实 bike 运行 r72 已到 **2 Plan / 1 candidate**；Main 已实际 merge 并执行成功的 unittest 检查，但最终 Result 接收因相对/绝对路径混算被拒，先报 checkpoint path escaped，绝对 root 重试后报 Main check record escaped。不能将其写成“未发生合并”，也未完成 S9、独立 Targeted 或 fresh S7。见[真实运行报告](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r72-bike-fixed-r70-main-final.json)。

同包 settings 运行 r73 已将候选写入 Runtime **revision 23**，随后保存回执超时；候选写入不等于最终 Result 或闭环通过，见[原始阻塞报告](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r73-cont2-main-blocker.json)。

工作树已精简 Plan/Result 作者 view 为现有 Draft 专业问卷及准确 Main 材料路径，修正 Main 相对路径读取，并删除候选成功后的重复投影。两原定向机械测试通过：**integration 25.034s / attempt projection 0.086s**；首次 integration 因新增断言将合法材料目录当文件读取而失败，仅修测试断言后原例重跑通过，原失败与 attempt 结果见[首次报告](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r74-final-original-tests-20261017T122709154602Z.json)，最终 integration 见[重跑报告](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r74-integration-directory-assertion-20261017T122836175999Z.json)。这是工作树机械验证；已安装 after-r70 和真实现场未替换，after-r71 正在建包，尚无真实闭环 PASS，不标 S9 完成。

以下为此前入口实测与机械验证历史，不能替代上述最新同源验收边界：

历史 after-r69 已运行并停止：两名作者首交 **2 Plan（含真实 `[]`）/2 ready**；backend 使用普通 Git 提交并做真实 self-check，test 如实报告 blocked。两个专业 Result 已写入，但 PostToolUse 出现 internal budget timeout，停止时 **0 candidate/0 final Result**；运行 486.672s，exit 143。见[原始阻塞报告](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r69-s9-reduction-real-main-blocker.json)。没有完成真实 S9、Targeted 或 fresh S7，尚未证实超时的唯一根因。

r70 独立 Claude 自建离线 bike seed，约 94s，真实测试 **10 项中 9 pass/1 fail**，Git clean。r67/r69 双 clone 使用同 payload、同 prompt，Main 默认预算 900s；实际两组均约 210s 因分工 Post 超时停止：各 **owners 2/backend 问卷 1/test 卷未发/专业 Plan 0/native 0/Result 0**。见[双组原始对比报告](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r70-bike-paired-real-comparison-final.json)。这只定位共同入口阻塞，不能宣布新旧优劣。

本轮已删除 blocked 保存与准确路径定位的重复 source/proof，以及 Result Hook 的 generic 工厂和双重定位；保留当前责任、真实 Git、专业答案和原 Writer 记录。两原合源机械单例曾 **PASS：blocked 49.80s/integration 46.21s/包 96.280s**，见[原始日志](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/logs/r71-final-ge_l7n04.log)；该运行早于最新 Result frontend 小片，不能称作当前最终同源验收。其后已删除 dispatch generic 工厂与 Plan issuer 重复完整 proof，并构建 after-r70；后续真实结果及未测修正以上文为准。

以下保留前轮失败与机械验证历史，陈述仅对应各自当时源码：

after-r68 真实运行 329.364s 后由 Root 停止（exit 143）：诚实 test Plan 的 `proposed_paths: []` 被 receipt `submitted_answer` 残留 `minItems: 1` 拒收，作者被迫改填 owned scope 重交。停止时为 **2 Plan / 0 ready / 0 Result**；这不算正常 Plan 首交通过，也未通过 S9/fresh S7，见[真实阻塞记录](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r68-s9-reduction-real-main-blocker.json)。r68 原公开 integration 单例机械 **PASS，24.982s**，见[集成证据](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r68-public-integration-evidence.json)；该结果不替代真实运行验收。

本片已删除 receipt 的拟改路径非空约束；原公开 CLI Hook 单例由真实派发 test-builder 首交 `[]`，确认原 report 与 receipt 都保持空数组，**PASS，15.270s**，见[原始日志](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/logs/r69-nochange-plan-public-hook-20261017T112024887503Z.log)。同期源码已删除 candidate/final 强制至少 fail/blocked 及 final 对范围偏差一概拒收的门槛，保留真实结果和必要 Main checks；当时这部分 domain/public 复测尚待，after-r69 尚未装包实测；后续当前结果以上文为准。无新设计、协议或字段，不标全阶段 PASS。

历史 after-r66 为 **built_not_run**。after-r67 真实 Claude 到两名本人 PlanReport、两个 ready Worker，RepairResult 接收数 **0**；普通 `commit_command` 仍要求旧提交表单，另有专业 answer 被外层 UI 的 `example.expected`/题干相等门槛拒收。Root 在 529.291s 主动停止（exit 143，非超时），未通过 S9/fresh S7，见[真实阻塞记录](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r67-s9-reduction-real-main-blocker.json)。

本轮源码已删除正常 S9 commit 表单指引，Plan/Result 直接提取 `answer` 并校验原专业字段，忽略外层展示字段；无新协议或字段。三个 validator 直接测试 PASS（0.327s），launch 定向 PASS（3.555s），CLI 编译通过（0.039s）；其后的 r68 结果与当前待验范围见上文，不标全阶段 PASS。

历史 after-r64 已构建且限定机械选择 9/9 PASS，真实模型到两份本人 PlanReport 与两个 ready Worker、Result 接收数 0；coverage/staging/workspace deadline 后停止，未计 S9/fresh S7 通过。详情及保留历史见 [S9 审查 §6](global-optimization/S9-MECHANISM-REVIEW.md#6-当前删减范围与证据边界2026-10-17真实完整链待验)。
S8 已收尾的限定范围与证据继续以下文为准。

当前宿主评估包为 **after-r56，Linux amd64 host-only**（[源码与包清单](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/after-r56-source-inventory.json)，包 SHA256 `0c07c4711afe916d56807e3e936078a36777a6f3f910dbf5e48863b58ba17576`）。S8 本轮 A/B/C/D1 减负切片及限定验证已收尾：59 项不同顶层定向测试 PASS、0 fail、0 skip（154.619 秒）；真实 Main→一名 native Investigator 本人 supported 保存→原 Case/route 接收→Contract 无授权等待，专业修稿后以明确标注的外部仿真人决策调用原批准事务，最终 Runtime rev11/Case r5 `contract_approved` 到 S9 planning 入口即停。三段模型均 exit 0、未 timeout，101 条直接 Hook 响应均 exit 0，选定内部 deadline/defer 与专业 schema/route 拒收为 0；[最终证据](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/reports/r56-s8-closeout-final.json)及[审查与范围说明](global-optimization/S8-MECHANISM-REVIEW.md)统一记录实际边界。上游 synthetic setup 借历史真实 Finding，不是新 S7 通过；human simulation 不是真实用户授权。三原 schema 单验 PASS，借用产品缺 scenario-model 的全项目 validate 失败及旧失败保留；未开展 S9 专业执行、全平台或全分支真实模型验证，不标全局目标完成。

用户要求已扩展为 S0–S11 全部阶段重新梳理、减负与测试；执行顺序和逐阶段状态见
[全阶段重新验收清单](global-optimization/STAGE-REDUCTION-REVALIDATION.md)。旧 4/9 保留为历史证据，
不作为本轮全阶段验收通过。先固定每阶段最小工作流，再做落地与仿真。

**以下为 2026-10-16 历史接续快照，after-r21 / 4/9 不再表示最新宿主或本轮状态。**

PreToolUse 监听与问卷接线已实施：取消逐工具拦截，自动准备当前问卷，完整投递
准确 answer 路径；S6 native/workspace 保存驱动原 begin/report，Stop/Idle 保留
准确路径与缺项，S9 提示复用原问卷/source/receipt。删除保存入口五处重复工具策略。
r59 六包与 r60/r62 定向回归通过；正常技能、角色和协议同步收敛。
该历史快照的 Linux amd64 隔离评估包为 after-r21。全新安装默认启用问卷，真实 Claude
完成 S2、S3、S4 专业核查与保存，原消费者分别经 PTR-PLAN-01、PTR-PLAN-02、
TR-002 推进至 S5；三阶段均未手工 issue/consume。原报告为 r91/r92/r93，
S5 两名独立真实评审也已保存专业问卷，原 TR-003 推进至 S6，没有通用 PLAN/begin；
报告为 r93b。当时该候选的真实阶段链为 **4/9（S2–S10）**；S6 实测暴露的
模板/帮助和旧手工交卷指引正在修复，待新包重新安装验证。

S6 原 Git 集成和清理恢复、S9 专业 PlanReport/RepairResult 自动投影原机器记录、
独立复验阻塞恢复与交接新 S7、S10→S11 及无授权不推进，已有定向回归通过。
同一合并源的 r90 全仓回归已结束：65 包通过，CLI、workspace 和系统 req039 三包
未通过。旧提示断言、S5 修订资格链、S7 样本版本及系统测试 import 已修，待复跑。
宿主无法建立 Bubblewrap user namespace，真实隔离不能记通过；
Main 检查路径仍需真实运行所有所需检查。

逐阶段当前状态与证据统一见[全阶段重新验收清单](global-optimization/STAGE-REDUCTION-REVALIDATION.md)。
历史 4/9、R11/after-r18 及旧失败保留在[历史接续快照](evidence/global-optimization-continuation-20261015.md)
和 §7，不能代替当前候选验收。独立 Impact 与 Bootstrap 归档协议已停止扩展；
后续缺口继续按 §0 复用原入口修复。

## 1. 交付范围与判定

每一点完成设计须具备：现有代码证据、具体差距、改动位置、输入权威、接口或数据结构、实际送达与接收路径、失败恢复、验收用例和兼容迁移。只有 guideline 或“增加提示”不算完成。

专业语义仍由对应角色判断；独立审查、冻结基线、Writer/journal、真实集成、fresh S7 和人闸保持。Hook 操作拦截按 §0 取消。目标是接住真实工作，减少执行者解释机制的负担。

## 2. 逐项设计入口

| 项 | 设计入口 | 核心落点 |
| --- | --- | --- |
| G01 | [上下文与技能](global-optimization/G01-context-and-skills.md) | 统一角色/派发技能选择，知识加载与权限分离，稳定工作包 |
| G02 | [质量与完成边界](global-optimization/G02-quality-and-completion.md) | 收尾断言、未决问题归属、职责回执和指令范围 |
| G03 | [任务粒度](global-optimization/G03-task-granularity.md) | 现有 TASK 分母/DAG 上的上下文与耦合分析 |
| G04 | [项目惯例](global-optimization/G04-project-conventions.md) | 有批准来源的默认选择和可重建参考视图 |
| G05 | [验证复用](global-optimization/G05-verification-reuse.md) | 检查回执完整性、适用性判断、受限复用 |
| G06 | [派发与集成](global-optimization/G06-dispatch-and-integration.md) | 原看板上优先接收与反压，保持真实容量与覆盖 |
| G07 | [执行环境](global-optimization/G07-execution-environment.md) | 检查 profile、平台探测、准确失败类别与恢复 |
| G08 | [评估与观测](global-optimization/G08-evaluation-and-observability.md) | 案例重放、零上下文实战、指标与模型实验 |

## 3. 共用基础：当前工作包

> 以下为原方案与已实施契约的历史说明；新增领域按 §0 收缩，不复制独立发行协议。

### 3.1 调查结论

现有 [team.GenerateReadbackRequests](../internal/team/launch.go) 已聚合文档、技能、scope 和收尾引用；[Hook 生命周期上下文](../internal/hook/lifecycle.go) 已解析 Assignment；[工作区 prepare](../internal/cli/workspace_command.go) 已给精确操作命令。应将它们的公共投影统一，复用已登记事实。

拟新增 `internal/workpacket`，只负责快照到执行视图的确定性投影。它不派发、不修改角色权限、不批准、不建立第二调度状态。原生 subagent、teammate、工作区 Worker 和 Main 的当前责任采用不同入口 adapter，消费同一投影。

### 3.2 包结构与来源

| 字段组 | 权威来源 | 作用 |
| --- | --- | --- |
| binding | 当前 Runtime、REQ generation、真实 owner、Assignment revision、阶段进入事件/轮次、execution generation | 防止旧责任、旧卷和错误身份被消费 |
| job | 当前领域责任、已批准 TASK/Claim/Case/Contract | 目标、本人交付、验收断言、当前动作 |
| inputs | 原 file_sources 的已解析 Git tree 或明确 disk evidence，原阅读清单 | 具体路径、版本、摘要、章节与阅读目的 |
| capabilities | 原角色/Assignment/activation/Runtime/策略的交集及执行环境 | 有效可写范围、真实 cwd、检查入口、操作 argv |
| questions | 与 Harness 同版的注册模板及当前领域展开器 | 本次题目、条件、值域、例子、answer 初值 |
| responsibility | 当前领域接收记录与依赖事实 | 本人完成状态、等待项、下一 owner |

主目录视图路径拟为 `.claude/work-packets/<runtime>/g<generation>/<responsibility>/<packet-id>.json`；Worker 安装为自己的 `.claude/inputs/work-packet.json`。独立Worker目录内填写 `.claude/submissions/stage-answer.json`；共享Main checkout的原生subagent/teammate各填写 `.claude/submissions/<runtime>/g<generation>/<responsibility>/<owner>/<attempt>/answer.json`。工具按发行记录生成安全路径，owner不直接拼接未校验字符串。每个当前attempt只有一个独占入口，发包给完整路径，不要求作者选择目录。Main的只读汇总不供并行手写。

`packet-id` 由发行时的真实 binding、模板版本和输入闭包派生，不能以全局 revision 变化强制重发。阶段进入事件、Assignment revision 或 round 必须区分重入；发包记录使用原 Runtime 的版本化扩展，包文件仅是可重建视图。没有真实可识别 owner 时不给可执行的签名题，也不猜“唯一等待者”。

### 3.3 统一投影与消息送达

拟定接口：`Build(snapshot, fileview, recipient, domainFacts, hostCapabilities) -> Packet`。`domainFacts` 由具名领域 adapter 提供，包含真实 gate 缺口及现有计划/报告/检查点；不得再由文件存在启发式替代领域结论。

只消费当前合法候选/实际选定路由的门缺口，区分可执行责任、已运行责任、其他owner等待与系统动作；不能把所有候选门的missing拼成待办，尤其不能把未触发的人闸/失败分支缺证据要求交给普通执行者。`workpacket`只依赖投影类型与只读接口，不import CLI或Writer；具体领域聚合在CLI/submission adapter完成，避免调用环。

status/next/Hook/launch 将旧字段保留为兼容渲染，新增 `work_packet_ref` 与结构化当前动作。已支持的角色身份决定 recipient；Main 看派发/接收待办，Worker 看自己的作业。安全决策仍在 policy 中，Packet 中的可用操作只是其投影。

有界原生消息先放 packet-id、本人状态、唯一动作、填写/资料路径、等待 owner，再放范围摘要。800 字节限制按 Unicode 边界处理；长 scope 不得挤掉唯一动作。详细资料始终可从包路径读取。

### 3.4 Schema 与版本

现有 `next.schema.json`、`loop-state.schema.json` 的相应对象禁止未知字段；不能直接追加状态字段。新增 `work-packet.schema.json`、authoring Schema、submission receipt Schema，并同步 Go 类型、内嵌资源、catalog 校验和安装测试。

拟采用下一 Runtime Schema 版本 `1.2.0` 承载可选优化配置、发包/接收索引；具体发布前确认该版本未被其他改动占用。旧 Runtime 保持旧形状与旧入口。新字段启用前须升级所有写入二进制，显式迁移 state/journal 配对并保留原恢复标记；旧二进制拒写新数据，不能降级后抹字段。现有 install 只支持新/空目标，不是活跃项目升级器；首版以新安装和旧Runtime兼容shadow为入口，不借install覆盖活跃资产。

当前单个内嵌 `loop-state.schema.json` 的 `schema_version` 是const 1.1.0，不能仅替换为1.2.0。新二进制保留两个完整封闭Schema：拟 `loop-state-v1.1.schema.json` 与 `loop-state-v1.2.schema.json`；在 `schema.Validator.ValidateBytes` 对逻辑名loop-state.schema.json增加固定版本分派，先读实例版本，再以固定白名单选对应Schema，未知/缺失版本按原拒绝规则处理。Writer、semantic/doctor、recovery、实例校验都走同一分派，不根据作者URL或路径选Schema；原v1.1内容保持其封闭约束。

新二进制对v1.1正常写入仍产出v1.1，shadow仅在工厂实验/读投影中计算新包，不写configuration.optimization、发包或接收索引，不开启新consumer。显式迁移后才允许v1.2字段。测试必须涵盖新写入器对旧state/journal的普通更新与pending恢复、对新版本的写入，以及旧二进制拒写新实例；不能只测新安装。

本设计中 `configuration.optimization` 的候选开关为 packet、context selection、receipt、check reuse 和 dispatch advisory；绑定当前 REQ 时固定配置与版本。新实验不暗改活跃责任；先 shadow 对照，启用后由 original domain gates 判定。

## 4. 共用基础：交付消费与当前回执

> 以下保存事件、发行身份及 followup 扩展属于旧方案/兼容实现。目标改为 §0 的
> 履职报告保存与原消费者；迁移前继续保留旧实例的历史和恢复。

### 4.1 现有事实与新增落点

现有 `assignment.CompleteTask` 已原子派生 S6 证据；`review.SubmitResult` 已在一个事务内消费 S7 Result、推进报告/覆盖并封 Batch 或 CleanRound；`runtime.RecordEvidence` 提供通用证据登记。现有函数未统一提供 authoring 模板接收和跨入口幂等键，不能直接承诺外层包一次调用即全部原子完成。

拟新增 `internal/submission` 作为登记模板、接收 answer、调用具名 adapter 和投影 receipt 的应用层。领域 adapter 必须调用原 API，不执行模板中的任意表达式/shell，不持 Writer 锁再递归调用领域 Writer。

### 4.2 接收顺序

1. 宿主从真实成功保存事件取得目标路径；仅处理已发行当前包的独占输入。读取大小有界的完整 JSON，解析失败保留草稿。
2. 通过真实 session/execution/Assignment 绑定 owner；冻结文件字节，并核对发行记录、题目版本及原输入闭包。
3. 校验 answer，null/条件未完成返回字段定位。计算规范化答案摘要；作者不得填写 producer、hash、generation 或内部事件。
4. 从原 API 生成领域 payload，使用稳定 `submission_identity` 与独立content hash。领域事务对同身份的消费强制唯一，校验语义/权限并登记原完成事实与接收索引；同身份不同内容拒绝，同内容重复消费返回原记录。
5. 释放领域事务后，以稳定followup事件及原cursor调用Controller求门；转换事务同时登记该事件结果，跨恢复也至多一次允许转换。长测试/合并派生Main/Integrator待办。
6. 从当前 Runtime 投影 receipt 和下一工作包。已经登记而求值/投影中断时明确系统恢复，不要求执行者重做判断或造信封。

### 4.3 幂等与崩溃恢复

`submission_identity=(runtime,generation,packet,slot,owner,attempt)`不包含答案内容，其摘要作为submission_key；content hash独立存储。领域请求新增可选身份与content hash，Writer锁内检查唯一并写接收关联；两个不同答案对同packet/slot/attempt只能一个成功，另一个返回content_conflict，不能因摘要不同得到第二合法消费键。原 `RecordEvidence` 拒绝重复ID、`CompleteTask`会派生修订ID，不能仅靠外层文件缓存声称幂等。

首次接收后同attempt不可变。修改已接收答案须通过原领域允许的修订/返工事件得到新attempt与新packet；author不能修改机器header自造attempt。原领域无修订入口时必须先实现具名入口，不能泛用supersedes。S5拟新增 `RecordDocumentReviewRevision`：只在同一S5进入事件、同冻结链、原真实owner且旧接收匹配时由Main发行有具体修订原因的新attempt；新结论接收时在同一事务把旧evidence标superseded/invalid、登记新evidence与关联。不改旧envelope。已离开S5或规格改变走原规划/需求返工，人闸保持。并发修订均绑定同previous receipt，其中一个提交后另一个拒绝/重新发题，不无限CAS重试覆盖。

统一索引记录binding、authoring hash、adapter/version、domain artifact refs和待恢复步骤；它描述消费检查点，不决定生命周期。跨域步骤保留明确的domain_committed与后续待办，每个已提交步骤有可核验原记录。响应丢失从原记录重建回执；修订按上述具名领域事件执行。

Controller后处理使用 `followup_event_id=hash(submission_identity,"after-domain")`。领域提交在其原事务中记录post-domain cursor、进入事件/轮次和control_pending。拟扩展ControlRequest/transition.Request携带该事件及expected origin；已有RequestID用于journal关联，但现有按transition/revision生成的幂等键本身不够。转换Apply在同一个Writer提交中检查事件未处理与origin仍匹配，并写followup结果/transition event ref；stage变化和事件处理标记不能分两次写。

恢复先查该事件结果及原journal（包括marker恢复），已转换仅重建当前回执/下一包。当前cursor已由其他事件推进时标followup被当前事实取代，不把旧提交应用到新cursor。求门not_ready且没有转换时记录evaluated_no_transition，未来新的交付事件有自己的followup；unknown/超时可保持pending，但只能对原origin有界重试。重复保存只返回原接收和当前事实，不再开启新控制请求。受影响origin发生返工重入也不得匹配旧事件。

包过期返回具体变更及重新发题责任；不把当前新输入 hash 盖到旧 PASS 上。包登记成功、视图丢失则重建，不重新署名。故障/unknown gate 不判通过，原安全边界保持。

### 4.4 回执字段与职责结束

全局统一状态的目标契约见 [L4 通用工作状态模型](l4-mechanisms/L4-state-transition-core.md#unified-work-state)。
下列接口是现有兼容 wire；领域 receipt 子类型不再扩展成 Agent 必须理解的责任状态。
后续集中资格/完成判断、统一输出，再按版本迁移旧字段，不在现有 Writer 之外增加状态存储。

`Receipt` 的发布接口为：`packet_id`、`submission_key`、`accepted`、`responsibility_status`、`work_complete`、可选 `consumer_owner` / `consumer_pending`、`current_stage`、`domain_refs`、`waiting_on[{reason,owner,ref}]`、字符串 `next_action`、`field_errors[{pointer,expected,observed,repair}]`，以及可选字符串 `error`。

`work_complete=true` 只代表本次责任的交付被原消费者接收，专业失败结论也可能是有效交付。它不代表阶段通过、代码集成或发布授权。仍欠自己检查/结果时为 false；仅等待别人或系统时不向本人重复发题。

receipt 的历史引用与当前状态分别投影；重复旧答案不能复活被 supersede 的结果。受影响责任返工时明确原领域事件与新 attempt，不能在工作已完成回执里静默重新授予产品写入。

单份S5规格审查接收后的回执示意如下。这是接口示例，不是已经实施的Schema或真实证据；真实binding与refs由工具提供：

```json
{
  "packet_id": "packet-example",
  "submission_key": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
  "accepted": true,
  "responsibility_status": "accepted",
  "work_complete": true,
  "current_stage": "S5",
  "domain_refs": ["document-review-spec-example"],
  "waiting_on": [{
    "reason": "另一项独立文档审查尚未交付",
    "owner": "reviewer-task-example",
    "ref": "DV-TASK-EXECUTABILITY"
  }],
  "next_action": "你的审查已接收，可以结束本项责任。",
  "field_errors": []
}
```

所有receipt枚举在发布的Schema固定，界面文案可翻译。waiting_on对应下一个工作的真实owner；next_action对当前recipient说明如何结束或继续，不靠“current_stage仍在S5”推出本人还有工作。

<a id="shared-submission-receipt-contract"></a>

#### 4.4.1 跨领域回执 Schema 闭包（2026-10-14，先固定契约）

共享 `submission-receipt.schema.json` 必须覆盖实际原消费者输出，保持
`additionalProperties=false`。历史 `submission_key` 直接保留原接收键，不重签、不改成
新 SHA：S2–S5 使用原 64 位小写 SHA；S7 使用原 `review-result-<packet digest>`；
S8 使用 `runtime:investigation-case:<case ID>:<positive commit revision>`；S9 使用
`runtime:s9:issued-plan-report:<packet ID>`、
`runtime:s9:issued-repair-candidate:<packet ID>` 或
`runtime:s9:issued-repair-result:<packet ID>`。Schema 对这些格式做封闭联合，
不能放宽为任意非空字符串。`current_stage` 是查询时 Runtime 阶段，不能用它决定
历史接收属于哪个领域。

现有 wire 字段 `responsibility_status` 表达领域接收/处置事实，兼容值固定为
`accepted`、`not_accepted`、`replaced`，以及 S9 的
`candidate_accepted`、`main_result_accepted`、
`main_result_accepted_rework_required`、`main_result_accepted_owner_status_unavailable`。
这些值不作为新增五态的同义词。`accepted=false` 只能是 `not_accepted` 且 `work_complete=false`；其余兼容值须有原
`submission_key` 与非空 `domain_refs`。`replaced` 及两种 Main 历史结果不可完成状态
须 `work_complete=false`；`work_complete=true` 只允许 `accepted`、
`candidate_accepted` 或 `main_result_accepted`，不把它推导为阶段通过。

S9 candidate 状态绑定原 candidate 接收键；其当前完成证明成立时，必须同时投影
`consumer_owner="main"`、`consumer_pending=true`。三个 Main Result 状态绑定原
final Result 接收键，均保留 Main 后续门 pending；仅 `main_result_accepted` 表示当前
Worker 完成证明成立。candidate 当前证明不成立时可保留历史接收、
`work_complete=false` 和具名待办，不能伪造完成。可选 consumer 字段成对出现；
`consumer_pending=true` 须有真实 consumer owner，不接受仅有 pending 的回执。
没有原 candidate 接收时返回 `not_accepted`，不能返回空责任状态的成功查询。
反向亦成立：原 candidate 接收键只能投影 `candidate_accepted`；原 final Result
接收键只能投影上述三个 Main Result 状态，不能换成泛用 `accepted` 隐去当前责任差异。

验收须同时包含 Schema 的正反例和**原 Hook / 公共 receipt 命令实际返回的原始 JSON**，
覆盖 S7、S8、S9 PlanReport、candidate 与 Main Result；不得先反序列化成少数字段再
验证，以免忽略新增/未知字段。此前冻结机制证据与独立安装包保持历史摘要，本节修复
须新增证据；旧包的 validate 通过不证明这个跨领域 wire 已被验证。

### 4.5 首个领域 adapter：S5

现有 S5 要求两个独立 producer，职责为 DV-SPEC-CONSISTENCY 与 DV-TASK-EXECUTABILITY，原 gate 消费 `document_review` evidence。第一版每人只填结论、发现和引用。

adapter 从冻结规格链与真实派发生成 review evidence envelope，经 `runtime.RecordEvidence` 的原校验登记，Controller 继续选择原 PASS/FIX/REQ-change 路由。登记 API 须扩展本节的同事务接收键；不得只给 envelope 文件、不登记证据。不能绕过 producer 独立性，Main 不替 reviewer 签字。

两人都 PASS 才能原门推进；首人完成返回等待另一 owner；FIX 交给原规划返工；REQ-change 消费原人闸。未满足原失败路径前提时给对应责任，不自动跳阶段。真实 Claude 保存事件、PostToolUse 回执、重复保存、基线漂移和重入须实测；原型不代替这些测试。

## 5. 交付策略的实施卡

### O01 — 可开工工单与预检

**改动：** `workpacket.Preflight` 消费原Assignment/领域计划/冻结资料/host capability。workspace prepare/launch与原生派发前置入口共用领域就绪报告；检查内容包括必需input存在与source一致、当前owner/依赖、计划屏障、scope、checks及环境。就绪报告是投影，不给未授权产品动作开权限。

**责任：** Main修派发缺口，规划者补实际收尾/决定，环境owner补运行条件。S6/S9计划阶段可派发合法计划责任，执行题仅在原条件满足后出现；S7先执行可开工的static责任，依赖未满足的behavior责任保留排队。

**失败恢复：** typed `PreflightIssue{code,source_ref,owner,required_fact,repair}`；输入/环境修复后重探测，已准备实例按固定binding恢复。unknown能力不标ready。

**验收：** 缺文档、未提交正式输入、blocked依赖、缺plan、缺环境、scope错误、原生/工作区入口等价性；合法计划仍可交，未授权产品写入仍拒绝。关联G01/G03/G07、M01/M08。

### O02 — 唯一交付责任与视图分区

**改动：** TASK本人交付表、原manifest ResponsibilityIDs与领域分母生成 `owned_outputs`；其他责任通过 `downstream_receivers`只读展示。新workpacket测试禁止把QA/E2E独立判断列为Builder签名题。

**责任：** 规划者分配，Main派发，角色按原职责交付。原S5两人、S7 mandatory/risk责任及separation_edges完整保留。

**失败恢复：** 无owner、重复owner或角色冲突给规划/派发具体来源；不让Worker自领或Main代签。已登记owner修订走原manifest/Assignment路径。

**验收：** 同一builder scope无独立审查签名、required responsibility无遗漏、跨owner提交拒绝、两个独立producer不能共用实际actor。关联G02/G06、M04。

### O03 — 行为样例、断言与oracle

**改动：** workpacket的checks/examples引用已批准合同与当前模块cases.json；S7直接展开原Claim assertion/oracle/method/required_evidence，S9展开approved Contract assertions。S6 Closing Contract无法机械解析时发专业题给规划，不执行任意文本断言。

**责任：** 上游设计者确定预期，Builder实现自己的检查，独立审查者验证与找遗漏。机器只枚举来源与实际运行事实。

**失败恢复：** 预期未决定、用例/引用失效、命令不存在，分别给source和owner；不自动生成语义PASS或补业务规范。

**验收：** 正常/拒绝/无数据例子、缺副作用期望、旧case、不可运行oracle、测试与契约矛盾、新业务决定未被惯例补齐。关联G03/G04/G07、M01/M06。

### O04 — 有效操作与精确修正

**改动：** workpacket capability adapter复用worker `ObservationCommands`、`WorkerInvocation`、prepare check_commands及Main pending argv，补领域固定操作生成器。`next.Action`从结构化action渲染，诊断同步扩展typed owner/source/JSON Pointer/argv；保留旧错误文本供兼容。

**责任：** 工具从真实binding填参数，policy仍求有效权限。专业参数如rework原因以题目收取，答完后生成命令，不把 `<reason>`渲染成可直接执行动作。

**失败恢复：** 缺actual identity/cwd/能力是派发接线待办；unknown token归adapter维护。状态过时重新投影，实际命令仍由原域校验。

**验收：** 空格/中文路径的argv、所有原门实际missing有映射、无占位命令、合法/越权动作与同角色跨工作区、未知门不猜成功。关联G01/G07、M02/M07。

### O05 — JSON注册、发题与保存消费

**改动：** 内嵌template catalog含Stage/责任/authoring Schema、问题/选项/条件/例子和具名adapter；首版S5。authoring与persistent evidence Schema分离；注册验证每题有消费者、每个手续字段有机器生产者。

**接线：** 同时更新Main `settings.json`和Worker Bootstrap生成设置；目前Worker为matcher `*`，仅扩大Main matcher并不能完成消费。`runHook`的PostToolUse早返回分支新增受控提交分支，Write/Edit/MultiEdit按真实payload取路径，其他文件仍是观察路径。受控保存adapter可直接调用同consumer；Bash只有声明的写入adapter可触发，不扫描任意shell字符串猜文件。

**写入边界：** 当前workspace boundary允许自身submissions目录，S9的repair control路径也包含submissions；但 `policy.reviewerWritePathAllowed`（S7）和 `phaseWritePathAllowed`（S8/S10）没有该入口，不能假定新增路径已被允许。拟新增具名authoring-path判定：只允许真实actor当前发行记录的独占答案文件，经原物理路径/symlink检查；作为这些阶段的控制面输入例外，不放开整个 `.claude/`，不授产品写入。确切独占文件校验不能被“仍在仓库内”代替：规范化发行根以后，受控答案路径下级不允许symlink重定向，尤其不能指向另一owner。工作包/Runtime/汇总仍只读，其他owner、旧attempt和未发行路径不给此例外；consumer再独立验证同一binding。原来允许的证据/报告路径继续走原规则，不能用新入口绕过计划屏障或签署别人的判断。workspace中也对受控答案应用此绑定，避免仅靠目录前缀接收。

**提交判断：** 机器由必答题、结论和条件题计算draft/complete，不增加Agent手填“ready”事件。null/缺条件是needs_answer；完整答案立即接收，非法值/引用给拒收；监测草稿与完整提交分别统计。读取损坏/尚未完成的JSON保留草稿和定位，不形成PASS。规划/结果分别发行当前题，先后由原授权保证。

**无Hook：** 只有宿主受控保存或明确tool-return consumer实际接线才声称自动；否则给绑定完整参数的兼容入口，不伪称文件保存自动推进。宿主入口沿可信actor上下文，不收author-supplied身份。

**验收：** 实际官方payload/wire、共享Main目录两人并行且答案不覆盖、Worker独立路径、各阶段合法保存通过PreToolUse后才触发消费；其他owner/旧attempt/未发行文件、路径逃逸、指向仓库内另一owner及仓库外的symlink不能取得新例外。再验最后保存无额外工具、草稿/条件/非法值、权限/输入漂移、所有支持宿主与未支持声明。关联G01/G08、M05/M06。

### O06 — 自动派生事实与合并重复载体

**改动：** submission adapter从真实Assignment、diff、执行回执、证据引用和已提交计划生成原消息/envelope；Markdown报告与兼容BUG仅做可重建投影。S6复用CompleteTask，S7复用SubmitResult，S8/S9/S10按下表分开调用原域API。

| 领域 | 已有入口 | 新adapter留下的专业答案/具体接续 |
| --- | --- | --- |
| S2/S3/S4 | RecordEvidence及原planning转换动作 | 本层决定、资料路径、覆盖/依赖；自动原planning envelope，转换继续登记原documents |
| S5 | RecordEvidence、原document gates | 独立结论与发现；见§4.5 |
| S6 | assignment.CompleteTask、agent-begin与Main task-integrate | 当前计划、实现/检查引用、结果；长集成归Main |
| S7 | review.RegisterPlan、review.SubmitResult | planner的覆盖/oracle和reviewer的Claim判断；hash/Result绑定由工具生成 |
| S8 | investigation intake/case_workflow/contract | 假设、因果、证据、路由、修复意图；每步形成原case checkpoint，人/已有授权审批条件照旧 |
| S9 | repair runtime/result/impact/targeted/handoff | 当前计划、修复结果、独立复验；create/commit外层接续并保留中断checkpoint |
| S10 | acceptance/audit验证、RecordEvidence | 有限inventory逐项判断、反证/风险；自动信封和来源绑定 |
| S0/S1/S11 | 原锁定/绑定/人决定入口 | 发待确认稿与人决定；显式决定到达才留证，不从文件存在推断 |

**重复计划：** S9generic与domain计划保留不同消费者。同一专业计划答案生成共享事实；generic的实际plan bytes绑定先完成，domain红色前置检查仍需真实证据和独立登记。所有Assignment domain reports齐全后原execution begin才开放写入，不能在最终completion追认。

**失败恢复：** 跨域接续复用§4的同事务索引/步骤原记录，不能外层抱Writer锁调用多域写入。不变事实引用来源，派生失败归系统恢复。

**验收：** 原Schema合格、每手续字段有机器来源、专业值不被自动补齐、报告与evidence一致、S9双计划时序及中断、S10无人工hash。关联G02/G05、M06/M07。

### O07 — 最小修正与缺口映射

**改动：** 从Schema ValidationError提取JSON Pointer/允许值，复用现有discriminator pruning保留原详细错误；领域diagnostic补source/owner/expected/observed。gate code/prefix注册到具名gap adapter，mapping test遍历实际候选门与fixture错误矩阵。

**责任：** answer错误归当前作者，资料/决策缺口归原owner，integrity/未知映射归Main/维护。系统知道的事实不能出现在补答题里。

**失败恢复：** 已接收其他slot保持，当前修订保留历史；相关输入变更精确失效，未证明依赖闭包时沿原保守重验，不静默保留旧PASS。FORM_GATE_UNMAPPED有实际缺口与维护责任，不能“继续DRIVE”。

**验收：** 单字段缺失不重填、非法分支只显示该分支错误、不同owner缺口不串、未知gate安全路径、输入漂移与旧引用。关联G05/G07、M02/M07。

### O08 — 责任完成回执与离场

**改动：** Guidance/next/packet引用§4.4 receipt；Stop/Idle从原本人交付与待项求值，Main得到后续派发/集成动作，Worker只见本人工作。

**责任：** 消费者接收，控制面投影，宿主给执行者真实additionalContext/tool return。工作包与receipt引用不能代替真正Result/ack。

**失败恢复：** 旧回执回放报告原记录及当前stage/等待；同事变更REQ或自身结果superseded时不继续旧动作。已接收后的系统pending不能迫使本人重交。

**验收：** 两reviewer先后、本人PASS但另人REQ-change、合法Finding交付、等Main集成、等系统恢复、已完成者Stop、本人缺结果仍受原规则。关联G02/G06、M04。

### O09 — 持久交接与恢复

**改动：** 复用原completion/assignment、review plan dispositions、case/repair pointers和integration checkpoints派生pending；新增submission关联只记录消费恢复，不维护独立可漂移队列。SessionStart/PreCompact作有界补消费/重建；被阻止的Stop保留原preflight顺序，不先写转换。

**责任：** Main/Integrator完成长集成和接收，领域consumer完成短事务；发布仍归原人闸。一个主会话恢复不模拟平台未支持的wake-up能力。

**失败恢复：** 写前中断→无domain完成；domain已提交→重建receipt/求门；迁移已提交→下一包重建；合并已发生→原checkpoint接续；只有log/文件存在不猜域提交成立。键冲突和输入漂移明确拒收/恢复。

**验收：** 每个持久边界故障注入，尤其“转换已提交但响应/外层检查点丢失”；同事件不能再推新cursor。覆盖多进程同身份不同答案、S5重签唯一前驱、旧答案回放、旧轮/Runtime切换、清理失败、合并不重复及journal marker恢复。关联G05/G06/G07、M05/M08。

### O10 — 无进展诊断与恢复责任

**改动：** G08的work_progress观察同作业/缺口/相关输入指纹，忽略纯时间/revision变动。配置只控制框架重复提醒，未知缺口即刻给维护路径；有真实running或外部owner等待不判空转。

**责任：** Main按原缺口补资料、修派发、恢复环境或进入领域返工。专业不确定交具体问题；不增加对人无必要的审批。

**失败恢复：** 达阈值给一次可定位stall而允许合法idle，不将未完成改完成；新证据后恢复原责任。诊断缓存损坏不关安全门，也不无限反复发送同一句提醒。

**验收：** 无变化重试、重复read/status、metadata churn、新证据、正常长运行、等同事/人、缓存损坏、不能自动替换writer。关联G08、M09/M10。

## 6. 覆盖、验收与发布边界

调查、已有测试运行、新设计评审、拟新增测试和真实平台验收分别留证；不把设计覆盖标成实现完成。迁移与启用以最小完整领域链为单位，未启用的入口保持旧路径并明确限制。

### 6.1 M01–M10覆盖矩阵

| 缺口 | 具体方案 | 必须证明的执行结果 |
| --- | --- | --- |
| M01 当前具体交付 | §3、O01/O03/O05、G01/G03 | 不查流程格式即可从包找到题、路径、检查及收尾 |
| M02 真实缺口投影 | §3.3、O04/O07 | 同快照status/next/Hook的动作/owner一致；机器转换不派为手写产物 |
| M03 必需信息送达 | §3.3、G01 §4/6 | 长scope/UTF-8/恢复时仍保留身份、next与包入口 |
| M04 自己完成 | §4.4、O02/O08、G02 | 本人被接收且仅等其他owner时可以离场，无重复催交 |
| M05 最后交付接收 | §4、O05/O09 | 最后保存直接消费/求门或有明确系统pending，无无关工具 |
| M06 可填写草稿 | O05/O06、G01/G03 | 已知值预填、专业题可答、缺上游有owner，没有内部Schema猜测 |
| M07 可执行修正 | O04/O07、G07 | 精确参数与字段定位；未知映射移交维护，原门不放宽 |
| M08 统一开工检查 | O01/O09、G07 | 每个支持入口有领域预检，真实平台不支持被准确报告 |
| M09 普通无进展 | O10、G08 §5 | 同缺口无新信息可诊断，长执行/等待不误判，安全门保持 |
| M10 首次使用验收 | G08、O01–O10 | 完整真实作业完成/失败/恢复，模式/trace缺失如实标记 |

### 6.2 实施切片与依赖

1. **基础投影**：typed domain facts、工作包、字段错误、本人责任视图；shadow核对真实gate。先不切旧手工入口。
2. **S5完整链**：authoring模板→真实独立owner→保存消费→同事务接收键/原evidence→原门/回执→中断恢复，连带Main与Worker设置。通过真实宿主后再删除该入口手工手续。
3. **上下文与质量**：G01选择器和权限等价测试→G02责任/指令→G04参考卡；用G08固定案例对照，不同时改模型。
4. **其他领域adapter**：S2/S3/S4确定性登记，再S6/S7/S8/S9/S10；每个领域先证明语义/时序/恢复，不能从S5通过宣布全部自动。
5. **任务与调度**：G03 advisory→G06接收优先→显式配置反压，基于已有DAG和scope。
6. **检查与环境**：G07 profile及真实平台→G05回执v2与shadow→小范围同职责复用。独立/合并后/fresh检查不变。
7. **观测与stall**：G08有限指标与实战runner贯穿，通用stall先shadow，充分反例后再启用提醒预算。

每个切片有自己的旧入口fallback与版本边界；fallback由工具生成真实参数，不让执行者猜命令。

### 6.3 迁移与回退的具体边界

首版仅在新安装中发行新Schema/内嵌资产，旧项目运行兼容渲染或shadow。当前install拒绝非空目标，不能用于更新活跃项目。

若后续支持活跃迁移，须单独实现维护入口（拟 `runtime optimization migrate`）：核对匹配release manifest/installed receipt；确认无活跃launcher/写入并保全pending；备份state/journal成对及资产；在临时副本验证纯Schema变换；原Writer持锁提交migration事件与配置/索引，沿marker恢复；只对可证明未自改的框架资产发布已验证版本，项目改写资产给明确合并报告，不覆盖。

变换只加版本化可选字段，不改bound REQ、generation、专业结论、权限、旧owner或已提交证据。Worker ready资产不重写，新profile/模板在下一合法发行使用。新Runtime曾提交以后，不用旧二进制回写；功能开关关闭可恢复旧界面/动作，但继续保留新Schema和历史数据。完整恢复需匹配的版本与备份，不能撤销已经发生的Git集成。

该迁移入口本次未实现，未验收前不能在旧项目启用需要新持久字段的consumer。

### 6.4 本次已有测试抽查

执行了team、hook、semantic、dispatch、workspace、integration、metrics、qualitygate的相关既有用例，覆盖readback包、生命周期上下文、冻结输入/bootstrap/launch、DAG/容量、真实board、合并检查/漂移、指标预算和planning门。首轮七个包的选择用例通过，workspace三个用例在根目录夹具验证失败；设置规范化 `TMPDIR=/private/tmp` 后三个workspace用例通过。

此结果验证可复用基础，并暴露路径别名边界；不是完整Go套件、未实施功能或真实Claude的验收。G07记录对应代码和拟补测试，G08区分这些证明层次。

重跑命令与范围：

```sh
go test ./internal/team ./internal/hook ./internal/semantic ./internal/dispatch ./internal/workspace ./internal/integration ./internal/metrics ./internal/qualitygate -run 'Test(GenerateReadbackRequests|BuildLifecycleAdditionalContext|BootstrapResumes|FrozenInputChanges|LaunchUsesOnly|DispatchNoWaveBarrierAndCapacity|DispatchPlanBoundaries|CommittedBoardAndRegistration|PlanCannotDriftOrBecomeLegacy|PostMergeChecksUseCandidate|VerifiedRejectsTargetDrift|TrackedFamilyBudget|PlanningGatesReadDiskDeclaredArtifacts|IdleRecovery)' -count=1
TMPDIR=/private/tmp go test ./internal/workspace -run 'Test(BootstrapResumesAndPreservesChangedAssets|FrozenInputChangesBlockPreparation|LaunchUsesOnlyChildDirectoryAndRegisteredSession)' -count=1
```

### 6.5 可行性审查结果

独立审查核对了八份分项设计、O01–O10与M01–M10覆盖，以及代码中的真实入口/权限/恢复边界。首轮指出的五项问题已修改并复核关闭：提交身份与内容摘要分离、Controller跨事务持久去重、两版完整Schema分派、Worker沿实际Read路径加载技能、普通Read/status的生产观察接线。

后续补充共享checkout的owner/attempt独占答案路径及各阶段写入策略接线，经独立复核未发现新严重遗漏；并将“仓库内symlink指向另一owner”明确列为拒绝用例。通用stall另明确首次读取新必要资料是进展，观察不足不强判空转。

审查结论为设计契约具备可实施性；没有运行新receiver、故障注入或真实Claude，不替代§6.2各切片的实现与行为验收。首个实施单元仍是S5完整发题→保存→接收→原门→回执→恢复链，随后按分项依赖展开。

## 7. 实施状态（2026-10-13）

### 当前核心交付清单（2026-10-15，后续详录为证据）

以下是 §6.2 的具名领域链清单，不以既有 CLI 或夹具通过替代新 adapter。
九个阶段中，S2/S3/S4 Main 与 S5 新机制链已验证 4 项，S6 有局部新能力，
S7/S8/S9 另有后续机械切片，详见 §7.47–7.54；真实模型完整阶段链验收为
**3/9（S2、S3、S4，§7.56）**。S5 真实独立审查仍在验证。这个计数不是全项目百分比。

| 核心 adapter | 当前新增能力 | 尚需完成的新链 |
| --- | --- | --- |
| S2 设计 | Main 原 Writer、保存 Hook、PTR-PLAN-01、回执/恢复及真实 Main 消费/保存已验证，见§7.56 | 多次稳定性、专业子代理入口另列适配 |
| S3 合同 | 原来源闭包、完整合同集、Writer、Main Hook、PTR-PLAN-02、回执/恢复及真实 Main 消费/保存已验证，见§7.56 | 多次稳定性、专业子代理入口另列 |
| S4 TASK | 原合同/条款分母、完整 TASK/波次、Writer、Main Hook、TR-002、回执/恢复及真实 Main 消费/保存已验证，见§7.56 | 多次稳定性；旧实例缺新 S3 来源时沿原恢复 |
| S5 文档审查 | 已发行/保存/原 Writer 接收、独立双责任、原门/回执/具名修订/恢复 | 真实模型送达/保存/恢复验收，后续提升策略仍须证据 |
| S6 构建 | canonical Record/PLAN/Result、本人检查闭包、原 Main 集成与立即前版接续已定向验证；见§7.41/7.43 | TASK 正式来源修订、可信宿主批准、完整 native/workspace 与真实模型链 |
| S7 验证 | 原 Claim/Result/覆盖/封轮 API 存在；共同产品基线→原 ReviewPlan 注册已组合验证，见§7.44 | 批准 Claim 发题、专业结果保存/幂等回执及停止投影，保持独立性与 required/fresh 覆盖 |
| S8 调查 | 原有序调查/修复合同 API 存在 | 当前 Case 答卷→原有序 register/consume/route→批准绑定及恢复 |
| S9 修复 | 已有部分 Worker 固定操作与原交付入口；canonical 产品基线 seed 桥接部分实现，见§7.44 | 计划/修复/独立断言三责任 adapter，原 Session/impact/复验/完整 canonical fresh S7 handoff |
| S10 验收 | 原权威审查全集/审计 API 存在；共享产品路径分母及 all 覆盖检查已定向验证，见§7.44 | 全集反证发题→原验收/审计消费→原人闸前回执及恢复 |

G01–G07 已有的机制详见各切片证据，仍不能称完整模型/平台收益验收；
G08 runner、保存观测和当前 shadow 进展采样只关闭对应机制子项。
下一实施次序为先闭合 S5 真实审查，再推进 S6→S7→S8→S9→S10；必要定向验证随
实现进行，完整回归、五平台构建与安装矩阵在适当里程碑集中执行。
原权限、人闸、独立性、真实集成与 fresh S7 始终是验收条件。

本轮按用户授权进入实施，先固定对应L层契约再改代码。L2新增本人收口规则；L3-S4/S5固定任务收尾与双责任答卷；L4阶段答卷§13固定生产契约，dispatch、hook、runtime、design foundation及workspace profiles分别承载G01–G08的具体消费边界。L1现有D1–D7与独立性原则保持。

| 实施单元 | 当前事实 | 后续必须完成 |
| --- | --- | --- |
| Runtime版本与封闭Schema | 已实现1.1/1.2固定分派；旧Schema保留，新字段独立；schema测试通过 | 新安装/init 1.2 与显式 packet、旧格式/pending 及 native 安装已验证，见§7.6；已发行索引 checkpoint/RR-A 原 pair 恢复见§7.7；真实宿主验收待完成 |
| S5工作包与题目 | 纯投影、内嵌模板、独占路径、冻结题干及字段定位；Main 注册/生命周期发行与独立 Worker execution view 已接线 | 新安装显式启用与正式安装已验证，见§7.6；默认 consumer 提升和真实宿主送达验收待完成 |
| S5领域接收基础 | 原 RecordEvidence 接收索引、规范摘要/精确信封、稳定 followup；Main/Worker 保存、回执、恢复、本人 Stop 与确切源权限已实现，首轮注册/激活/原门/wire 与具名修订已接线，新增修订验证见§7.5 | 默认启用/旧实例恢复、完整安装及真实宿主验收、独立复审 |
| 后续领域及G01–G08 | L层实施契约已固定，生产功能未完成 | 按§6.2切片逐项实施与验收 |

相关验证：schema/workpacket/runtime全包与race测试通过；semantic/assignment/controller与req039系统回归通过。CLI首轮九项受真实工作目录旧Runtime污染，在原HEAD干净目录复现为通过；改为隔离工厂夹具后九项及完整CLI回归通过。本表只记录已执行结果，完整S5链、真实Claude及其余优化尚未完成，不将基础API当保存自动化交付。

备份时基础切片独立审查发现P1：新增Submission接收路径尚未验证真实答卷，也未将调用方摘要与document_review的conclusion/findings/完整subject_refs绑定；非法信封可能占用不可变attempt。2026-10-13按用户要求先提交并推送当前状态作备份，该备份不是可发布的完整功能。接手修复与实际验证见§7.2，完整S5链仍待实现及独立复审。

### 7.1 接手基线与实施顺序（2026-10-13）

接手基线为备份提交`ed61728`，工作区无未提交改动。该提交修改48个文件，包含全局设计、L层消费契约、Runtime版本分派、S5纯投影与接收基础及工厂隔离夹具；未实现submission应用层、保存消费、转换followup去重、具名重签和统一回执。历史测试记录保留为当时结果，接手验证单独记录。

实施沿§6.2依赖展开，每个单元先修改所属L层，再同步生产代码、Schema、安装资产和用户文档：

1. **接收完整性**：按L4阶段答卷§13.3验证冻结authoring、机器计算规范摘要、信封与精确subjects绑定及重放完整性；拒收不占attempt。以畸形/不完整/篡改答卷、信封矛盾、并发异内容及恢复重放关闭P1。
2. **S5完整链**：发行与宿主送达、独占路径权限、受控保存consumer、原门followup事务去重、责任回执/Stop、具名修订和有界恢复；将新安装与旧Runtime兼容、安装资产和官方Hook wire纳入同一切片。
3. **上下文与完成边界**：G01加载选择与权限分离、G02责任收口、G04项目惯例；用G08固定案例取得基线和对照，不同时调整模型默认。
4. **领域扩展和调度**：按§5具名adapter接S2/S3/S4，再S6/S7/S8/S9/S10；G03粒度建议和G06接收优先复用原DAG/看板，不建第二队列。
5. **执行与评价**：G07环境/profile先于G05受限复用，G08指标/无进展观察与首次使用评估贯穿实施。真实宿主不可执行项明确保留待验收，不用fixture成功替代。

各切片的具体实现、实际验证和剩余工作回写本节；仅完成基础API不标完整链通过。先实现到可审阅状态，既有授权范围内持续推进；发布、活跃项目迁移和新的人闸继续消费原明确授权。

### 7.2 接收完整性与控制接续基础

先补L4阶段答卷§13.3/13.4契约，再实现如下基础：

- SubmissionRequest携带冻结的完整答卷。RecordEvidence在原Writer事务内验证注册Schema、唯一JSON键、发行题干、完整answer、机器计算的内容摘要、信封的身份/结论/findings/精确subjects及固定路由；非法请求不留下evidence或submission。同内容重放须保持原domain引用、信封路径和指纹。
- S5具名信封投影由workpacket.DocumentReview提供；FIX派生原document_fix_required，REQ-change不产生自动事件。缩进变化不占新attempt，专业判断变化保持content_conflict。
- Controller和transition.Apply消费稳定followup及原进入身份；转换与advanced/transition_event_id同事务提交；并发、响应丢失和旧进入身份不重复推进。not_ready登记no_transition，unknown保留pending。该基础尚未接到真实保存consumer，不能称为保存自动化。
- 接手时在干净ed61728复现既有Schema分支摘要长度测试失败；按G07契约去除重复分支名，保留逐分支首个错误、unwrap与verbose入口。

上一基础切片已执行：workpacket/runtime/schema/controller/transition普通与race测试通过，相关包go vet及diff检查通过；Controller以真实门与Writer验证转换/接收结果同事务、同步并发只转换一次、重放及unknown。NativeWorkerAdapterChain的Main runner通过；当时semantic/assignment/qualitygate/CLI回归通过。首轮系统套件只有上述Worker隔离子项失败，其他已执行子项通过。这些结果早于§7.3新接点，不能作为新保存完整链或新增授权逻辑的回归结果。

当前环境Linux Bubblewrap运行报setting up uid map: Permission denied；原系统套件的Worker runner子项因此失败，不能记为通过。没有放宽隔离或自动替换runner；真实Claude保存链与其他平台仍待完整consumer实现后验收。新增接点进展与下一步见§7.3。

### 7.3 共享 Main 的生产接点（实施中）

先补 L4-stage-form v0.7.2 与 L3-S5，再编码：

- v1.2 原 workgroup Writer 写入实际 packet_stage_entry/packet_generation；发行与接收检查 manifest 的 Runtime、代数、workgroup、精确 Assignment、当前 subjects 和源视图。旧进入身份的 working reviewer 不自动当作新派发；v1.1 注册形状不追加字段。
- 新 internal/submission 应用层连接原 API。先登记 packet，再独占发布完整 view/draft；已有答卷不覆盖，路径拒绝下级 symlink。规范答案摘要决定具名证据路径，同内容并发复用先发布信封的时间戳；接收时再次调用原计划/激活校验器，锁内仍要求真实 working。领域事务完成后再调用 Controller。
- PostToolUse 在 observer 前消费 Write/Edit/MultiEdit 的真实 file_path，使用真实平台 owner 与重新加载的上下文；Bash 不猜路径，普通消息 observer 保留。回执通过官方 additionalContext 输出；本次拒收不冒充已接收，也不对已落盘的保存发 deny。
- register-workgroup、SessionStart/SubagentStart 接入发行投影，SessionStart/PreCompact 有界恢复 pending；终端 fallback 提供 runtime stage-answer issue/consume/receipt/recover。确切答案路径权限从可信 Runtime 加载，原计划链未成立不授予写入例外。Stop/Idle 从真实领域接收判断本人完成；FIX 路由失效的是门资格，不撤销本人原交付。
- 角色、document-verification Skill 与 agent-protocol 增加“已有真实卷走保存回执，否则保持旧信封流程”的分支。安装/init 默认仍为 v1.1；不要求手改 Runtime 开启。共享 Main 接点已编译通过，尚未进行新接点完整链验收；旧测试结果不计入本切片的新功能验收。

接续顺序：补独立 Worker 的 S5 execution/冻结输入和隔离路径绑定；固定并实现具名修订的 supersedes/失效与并发契约；完成新安装启用及旧实例恢复边界；再验收真实双 owner 最后保存、原 PASS/FIX/人工路由、权限与计划、崩溃和并发恢复。完成这些之后才宣称 S5 整链交付，然后按§6.2继续其余领域和 G01–G08。回执可重建，但磁盘汇总/下阶段卷及通用诊断仍待后续 adapter；完整优化目标保持 active。

### 7.4 独立 S5 execution 与实际接收链验证

先将进入身份、独立根、冻结输入、双入口停用、原计划与本人 Stop 固定到 L4，再落地：

- 原 workspace.prepare/launch 支持已开启 v1.2 的 purpose=document_review；同一 execution 行绑定 packet、stage_entry、owner、execution generation，产品 scope/checks 为空。新 Schema 只进入 v1.2，旧 execution 默认形状保留。
- Bootstrap receipt 覆盖 execution view 与冻结资料，答卷保持可写且不纳入只读摘要。独立源声明后拒绝 Main 旧答卷，接收锁内重新验证当前 execution/view/冻结资料及原激活链；替换保留旧目录并拒绝已收到的原交付被新 execution 覆盖。
- 新 workspace answer/receipt/review-check 采用原 root 路由和确切 argv admission；review-check 复用原 TASK 语义检查且不写 PASS。产品 commit/report/delivery 拒绝 document-review purpose。PostToolUse 可识别真实 Worker 路径，生命周期和角色/Skill 给正确冻结读取与提交入口。
- 顶层 S5 Worker Stop 在任何 Controller mutation 前判断本人已接收或原进入身份已退役；未完成的当前责任阻止，Main 的既有 preflight 不替换。计划前保存仍通过官方 PreToolUse permissionDecision=deny 阻止，不误将该 wire 的 exit 0 当允许。
- 新验收从已有 S5 注册文档基线开始，所有 workgroup、计划/激活、packet、正式保存、领域登记与迁移均用生产 CLI/Writer；没有手植 working Agent 或已接收 review。已证明双 PASS→S6、FIX→planning、REQ-change 留在人工决定前；同内容重放、domain 后接续中断恢复、转换后响应丢失、并发两 Worker 不丢答案且只有一个 TR-003、本人 Stop、错误 owner/view/冻结输入/symlink/旧进入身份拒收。被冻结文档集合同时精确匹配已授权 bound_req 身份/版本/摘要，避免仅 documents[] 自洽却背离需求锁。

已执行普通新接点测试、runtime/assignment/schema/semantic/controller/hook/hookctx/policy 回归，以及相关 go vet；最终完整 CLI 回归通过。runtime/CLI 的选定 S5 接收、源绑定、Stop、恢复和并发用例通过 race 检查；workpacket 同命令报告 no tests to run，不把它算成该命令的新增 race 覆盖。workspace 套件的隔离测试再次复现 bwrap uid map Permission denied，其余子项通过明确 skip 单项的运行验证，不能宣称完整 workspace 套件通过。最终编译与 diff 检查通过。该证据范围不包含真实 Claude child 的能力执行、S0→S4 上游建基线或完整安装升级；默认 init/install 仍未启用 packet。

下一切片：具名修订的 previous receipt/supersedes/资格失效与并发接收，随后默认启用/旧实例恢复与安装验证及真实宿主验收；其余领域 adapter 与 G01–G08 继续保持原完整目标，不将当前 S5 首轮证据当整项完成。

### 7.5 S5 具名修订与提交窗口资格复查

先固定 L4-stage-form v0.7.4 与 workspace profiles，再实现：

- Main `runtime stage-answer revise --previous <真实 submission_key> --reason <具体观察>`；原 Writer 绑定不变的 runtime/generation/stage_entry、owner/Assignment/责任及全部冻结 refs，追加 attempt+1。相同前驱/原因重放唯一新 packet，不同原因冲突；未收到、非当前资格、冻结资料漂移或离开 S5 拒绝。
- 发行同事务暂停旧 evidence（invalid、pending-revision 原因）、记录 revision_packet_id、退役旧 pending followup；新接收同事务登记新 evidence/submission、旧 superseded_by 及最终失效引用。旧 artifact 字节不改，未接收的新草稿不占用 attempt。当前 packet 按唯一 lineage tip 选择，不依赖数组排列。
- 历史 receipt 增加真实 submission_key；被修订的旧交付仍 accepted，但 replaced/work_complete=false，指出新责任。Stop/Idle 不能用旧完成替代新交付。Main 可用 `stage-answer receipt --agent <owner> --packet <packet_id>` 只读查询历史，不能借此恢复旧答案 grant。
- 独立修订持原 Git 调度锁和旧 launch lease；运行中的 child 阻止退役。发行同事务将旧 execution 归 retired history、移除活跃映射，固定新 packet execution_required。原 prepare 在 history 上计算更高执行代数，产生新目录/view/空答卷，旧资产保持，未 ready 时不回退 Main。原计划/激活校验与只读产品范围保持。
- 当前已发行的 S5 责任拒绝手工 evidence add 的并行结论，外层与原 Writer 内均复查。正式安装/旧实例恢复还需验证未索引历史证据和恢复入口的资格，不能以当前测试默认启用所有模式。
- 新提交窗口测试初次失败，定位到 required_evidence 只有一个 document_review 槽，Controller 的两位 reviewer 证明在 Apply 丢失一个；原 joint_document_pass 是 attestation，未覆盖被省略的那份资格。修复为 Controller 携带完整 S5 gate evidence refs，Apply 在原 Writer 内逐项复查 valid/代数/摘要；TR-003 还检查当前修订 tip 已收到有效 PASS，手工迁移不能绕过未收到或非 PASS 的修订。保持原门与人闸，没有用测试放宽判定。

已执行并通过：修订普通用例（Main CLI、历史回执、异原因冲突、空答卷拒收、旧保存拒收、同伴 PASS 等待、新 PASS/FIX/REQ-change 原路由、同内容重放）、独立修订（真实 prepare/begin/save、旧 lease 阻止、Worker root 不授 Main 调度、历史目录和 bootstrap 保留、新代答卷）、提交窗口旧 PASS 失效、runtime 同前驱并发唯一 successor 和失败不变更。runtime/schema/workpacket/assignment/controller/policy/hookctx 回归通过；追加 Apply 复查后 runtime/controller/transition 全包回归通过；semantic/qualitygate/hook 回归通过。workspace 明确 skip 原 bwrap 隔离失败单项后的其余回归通过；不声明完整隔离套件通过。最终完整 CLI 回归通过（240.590s）；修订目标用例 race 通过（runtime 10.031s、CLI 166.293s）。随后追加的“发行已提交但 view/draft 投影丢失”恢复用例在 pass 子项单独 race 验证通过（34.541s），恢复同一 packet 且 Runtime revision 不增加。最终 go build ./...、相关 go vet 和 git diff --check 通过。以上不是实际 Claude child 或完整安装升级验收。

接续顺序：核对新安装/init 的 1.2 版本和显式启用契约，验证旧 1.1 常规写入及 pending 恢复不暗迁移、旧写入器拒绝新版本、installed receipt/匹配资产和证据资格；安装只对新/空目标，不覆盖活跃项目。随后执行真实 Claude 宿主保存/中断/恢复验收并独立复审，再按§6.2推进各领域 adapter、G01–G08。默认尚未开启，真实 Claude 尚未执行，完整优化 goal 保持 active。

### 7.6 新安装/init 1.2、显式 consumer 与原格式恢复

先固定 L4-stage-form §13.1.1 和 runtime control plane，再实现：

- 新/空目录的 install/init 与原自动初始化生成 Runtime 1.2，packet 默认 false；显式 `--packet` 只用于新实例。REQ bind 不重写格式/配置，不生成专业权限。已有 state/journal/init receipt/marker 不覆盖，旧部署不暗迁移。
- 初始化先做完整 Schema、fresh-inactive 和语义校验；对每个已存在组成员做完整字节预检查，已知冲突不发布缺失的 state/journal。原 pending 重试不指定选项时保留原版本/模式，显式不同选项拒绝并保持 marker。新增 runtime-initialization 回执记录原始版本、配置及初始 state 摘要，只给恢复格式提示，不取代当前 Runtime。
- 原普通 Writer 拒绝 schema_version/optimization 的暗改。原人闸 rollover/unbind 在 Writer 锁内采用实际源格式/配置，历史索引随原 pair 归档，新 inactive 索引为空。S0 尚未绑定时只观察、不发卷/授写权，绑定后的消费者仍严格核对权威；避免全新 --packet 在首次 SessionStart 报错。原 reset pending 恢复核对 target 与 sealed source 的格式，混合版本/模式拒绝且保持 state/journal/marker。
- 有效 1.1/1.2 commit pending 原样恢复，新模式保留实际 packet/submission 索引且重放不提交第二次。RR-A legacy 源继续 legacy seed，已知无发行索引的 1.2 源保留格式/配置；已有索引、未知新格式/不可读的新初始化源、混合来源拒绝文档重建。其索引完整重建仍须实施 O09，不能把拒绝当作恢复功能完成。
- 使用当前未提交工作区构建本地 release 包：五个平台二进制真实 cross-build、生成 Manual、205 文件 release manifest 及 release graph 通过；source_revision=ed61728、source_dirty=true，不是发布包。实际 Linux native 二进制对 default/--packet 两种新安装通过 doctor、validate --all、installed release graph、docs check；207 个 installed receipt 文件逐项摘要验证。
- 从 85ea968 的真实 1.1 源构建旧二进制，在两种新安装上的 Schema 校验与实际 evidence add 均拒绝，state/journal/初始回执/安装回执摘要不变。旧 reconcile 对一致 inactive state 可以 no-op 返回成功，该结果不算新格式认证；没有把 no-op 误判成已经拒写或已经兼容。

相关 runtime/schema/controller/transition/assignment/recovery/releasegraph/semantic/hookctx/policy 回归通过；初始化/绑定解绑/旧格式/完整接收索引 pending 恢复的选定 runtime/CLI 用例 race 通过（4.808s/56.597s）。最终完整 CLI 回归通过（221.403s），最终初始化/格式恢复/绑定解绑和 S5 保存、独立源的选定 CLI race 通过（391.826s）；此前 runtime 格式/pending race 4.808s 也通过。未知未来格式及初始化已知冲突分别有额外目标回归/race，最终 go vet 和 diff 检查通过。真实平台探测：Claude Code 2.1.278，loggedIn=false、authMethod=none；没有读取凭据、没有启动真实模型作业，记 NOT EXECUTED。Bubblewrap 宿主限制仍按§7.4保留，没有降级。

本地验收证据位于 `/tmp/gsb-optimization-install-asqxxusa`（临时目录，非持久发布资产）：原 package/final-package、本机安装目录、legacy 二进制、清单和 native acceptance JSON。最终含初始化冲突预检查和 S0 观察修复的 v3 本地包再次完成五平台构建与两模式安装；安装 receipt 全部 207 项摘要通过，实际 packaged SessionStart Hook wire 在两模式下都保持 inactive/无绑定、无 packet/submission，不再报告不存在的 S5 权威错误。此 wire 由测试进程提供真实格式的 payload，不冒充实际 Claude 模型作业。持久摘要见 [native 安装证据](evidence/global-optimization-native-install-20261013.json)。

接续：实现 O09 对已发行/接收索引的完整原事实恢复与迁移审计；同时推进 G01 的完整资料送达/有界上下文、G02 本人收口和其余各领域 adapter。真实 Claude 登录前继续可独立验证的工程工作，consumer 默认不提升；登录可用后补原 professional review、实际 Hook 保存/中断/恢复。完整 goal 保持 active，后续 G01–G08 与 fresh S7/真实集成/人闸要求不缩减。

### 7.7 已发行索引的完整原 pair 恢复

先固定 L4-stage-form v0.7.6 §13.6 与 control plane，再实现：

- 原 Writer 在 1.2 有 packet/submission 索引后，提交及指纹刷新清除 pending 前发布完整 state/journal checkpoint；包含所有旋转段，文件完整 fsync/独占发布后才更新索引，最多保留最新两份。发布冲突/路径别名/超限保留原 pending，恢复只补发布，不重复领域提交。原 RR-A 应用亦在清除 recovery marker 前重建同一 checkpoint。正常状态读取、门禁与权限不改读快照。
- `recover inspect` 给出最新 manifest 位置，`recover plan --checkpoint <manifest>` 生成 `checkpoint_restoration` / `original_pair` 的冻结候选；保持 Runtime/REQ/代数/revision/cursor、原 owner/计划激活/execution registry、修订链、证据失效、接收键和 actual transition stamp，不 import/replay 专业事实。三个 recover leaf 的 help 展示真实选项，不读取/创建 Runtime。
- 原 `apply --approved-by` 沿用人工审核、Runtime 锁、完整隔离、sealed marker、pair 替换与幂等。补齐 journal 旋转段的输入指纹、隔离与退役；避免完整 journal 与旧段叠加，包含缺失成员和中断退役的重试校验。原合法 commit/reset/recovery pending 仍优先于文档计划。
- 完整 pair 做封闭 Schema/语义/摘要/cursor 校验，同时复核每份题干、原发行/接收 journal、专业 envelope/answer 摘要、责任/代数、修订 supersedes 和旧证据失效、原 no_transition/advanced 事件。较新 checkpoint、不同身份/基线/REQ、同 revision 不同事实、已知 journal 更新、原边界归档都拒绝回滚；不从 answer 文件存在推断接收。
- 应用持原 Git 调度锁，在 Runtime 锁内重新核对最新来源并持实际 Worker launch lease；仍运行的 Worker 拒绝应用。沿原 checkout、Inputs、Bootstrap、冻结 review view、REQ committed reader 和计划激活校验，不能把损坏的当前 state 当恢复候选的 authority，也不补造 Worker 答案。来源仍需可验证；此切片不恢复已经丢失的专业 evidence/原计划/Worker 资产。

生产 CLI/Writer 用例从原已注册 S5 文档基线起步，真实 register/begin/保存 Hook/原 gate/revise，然后损坏 active pair：修订 pending 不消耗旧 PASS；新答卷仅经原 TR-003 推进一次；advanced 再恢复不重复迁移；两次原 pair 恢复保持全部索引字节事实。旧快照、重复键、缺失修订链、复活旧 PASS、伪 transition stamp、活动 launch lease 和 symlink index 都有拒绝验证。checkpoint 发布失败保留原 marker，原重试只生成一条 journal；RR-A 在 state/journal 替换与 manifest 前故障后，即使计划/快照目录丢失也从 sealed marker 完成，下一原 Writer 的连续性保持。

相关 runtime/recovery/workpacket/schema/controller/transition/policy/hookctx 全套回归通过（runtime 42.011s）；完整 CLI 回归通过（330.795s），恢复 help 与原 help 兼容回归 0.050s。完整 checkpoint CLI 用例普通 57.736s、race 255.765s 通过；原 Runtime RR-A 回归 4.477s、go vet、diff 检查通过。最后补入的恢复后快照再发布由当前完整 CLI 覆盖；sealed-marker/source-loss 与真实 help 的补充 race 93.794s 通过。

当前未提交工作区再次完成五平台交叉构建和本地 205 文件 release manifest/release graph。实际 Linux native 包在 default/--packet 两种全新安装上通过 doctor、validate --all、installed release graph、docs check 和真实 `recover plan --help` 的 checkpoint 选项；inactive/空索引/原默认开关均保持。包 source_revision=ed61728、source_dirty=true，未发布。持久摘要见 [checkpoint native 安装证据](evidence/global-optimization-checkpoint-native-install-20261013.json)，本地包与验收目录是临时资产。业务恢复测试使用 production CLI 与已注册 S5 基线 fixture，未冒充真实 Claude 专业审查/会话。

下一切片：G01 完整角色/Skill/派发资料送达、统一有界 UTF-8 lifecycle context 和真实 plan scaffold；随后 G02 本人收口、各领域 adapter 及其余 G01–G08。G08 同时记录快照 I/O 与长 journal 的成本/预算，不能把新恢复路径当免费操作。真实 Claude 仍未登录，consumer 默认仍不提升；完整 goal 保持 active，fresh S7、独立判断、真实集成和发布人闸继续适用。

### 7.8 G01 shadow、冻结资料与生命周期预算

先固定 L4 dispatch §18.1 与 Hook wiring §12，再落地：

- 内嵌 agent-context-profiles 集中原8角色 legacy 列表和责任 mandatory 集合；shadow selection 合并 core、实际责任必需与原登记 SkillRefs，每项给原因、安装路径与摘要。没有已批准技术栈来源就不猜模板。team readback 添加同一候选投影，原 skills 完整列表、角色 frontmatter、manifest/activation SkillRefs 与能力兼容回落保持；assignment 用例比较原权限精确等价。缺已安装 Skill 不回落工厂另一份来源。
- 新 1.2 workspace.prepare 从原注册 Agent/Team/manifest 生成实际 owner/assignment/task/mode/责任/读写/检查/收尾投影；原 committed 文档链包含所属 TASK 与登记 depends_on 闭包、共享 REQ/设计/合同，另带实际协议、规则和选中 Skill 来源。context 进入原 Execution.inputs 并复制到 Worker `.claude/inputs/assignment-context.json`，由原 BootstrapReceipt 固定，不设第二派发表。按 TASK→合同→REQ/架构/设计→规则给实际读取序。
- plan_checkpoint 同时交付真实机器身份的 PLAN_REPORT 草稿、专业问题和完整嵌入 Schema；objective/steps/assertion_checks/risks 等保持未答。草稿校验失败、原 begin 不激活，执行者补专业字段后仍经原计划链成为 working。可编辑 plan/answer 不进入只读回执，ready 重试保持作者字节。approval 模式不冒充存在草稿，继续原批准。workspace allowedTools 不新增 Skill，通过已有 Read 指向实际冻结来源。
- 显式 replacement 保留原 context/profile/资料，只重绑新 execution generation，旧目录和作者字节保留。修复 CLI 原 S6/S9-only 检查与已实现 S5 replacement 的不一致：仅原 document_review、当前责任且尚未接收可以替换，不能选择产品 runner；已接收答卷拒绝替换。原来源、原领域、期望代数、checkpoint、Main 和实际 launch lease 限制均保留。原 ready/legacy preparing 不补写新的 context。
- lifecycle 最终 additionalContext 使用一个800 Unicode rune 预算，含 S5 提示和最后提醒；合法 UTF-8、完整动作/身份/责任入口优先，长 scope/checks/done 只作可省略详情，不能剪半命令/路径。实际 Worker 给本地 context/view/answer，legacy Worker 指向原 loop-workspace 指针；未知 child 不发他人责任。保留原 Session recovery/manual/Next 与 S6 派发/集成建议，worktree 提醒只有实际进入 wire 才记录已投递。

新增生产 CLI 用例用真实 prepare 生成的机器字段填计划，仅专业字段由测试作者填写；原 begin 与正式保存 Hook 验证答卷接收。覆盖空草稿拒绝、实际冻结路径/digest、bottom-up 阅读序、ready 作者字节不覆盖、官方格式 SubagentStart 最终800字符 wire、replacement 代数与原 profile 固定、S5 runner 拒绝、接收后替换拒绝，以及 source 漂移不覆盖 Worker。它们不等于真实 Claude 专业执行。

Worker 新用例普通22.853s、补充 replacement 边界8.398s、race110.313s通过；agentcontext/team/hook/assignment/catalog/hookctx race通过（assignment55.931s，部分包使用有效缓存）。workspace 回归12.045s通过，单独复跑原隔离测试0.257s仍因宿主 bwrap uid-map Permission denied 失败，没有放宽或替换隔离。runtime/recovery/schema/workpacket/controller/transition/policy 领域回归通过（runtime55.219s）；相关 go vet 通过。最终完整 CLI 回归341.765s通过；semantic/releasegraph/qualitygate 补充兼容回归通过，最终 diff 检查通过。

当前未提交工作区完成五平台交叉构建、本地 release manifest/release graph 和 default/--packet 两种新安装；native doctor、validate --all、installed release graph、docs check、原 recover help 通过。使用明确的源示例 fixture 作为 readback 输入，打包二进制按自身嵌入 Schema 生成投影，实际安装的 selected Skill 路径/hash逐项吻合；8角色 frontmatter 与原 HEAD 完全相同。source_revision=ed61728、source_dirty=true，未发布。持久摘要见 [context native 安装证据](evidence/global-optimization-context-native-install-20261013.json)；fixture 不是注册业务任务，native 检查不是模型加载验收。

后续仍需三个入口实际宿主的 Skill/Read 消费证据、首次有效操作与 token 对照，以及首个 Builder enabled profile/最小 frontmatter 生成。下一切片先补 native/teammate 的共同冻结资料投影和入口验证，再按已固定契约推进 G02 与其余 adapter/G01–G08。shadow 不算已经减载，consumer 默认不提升；完整 goal 保持 active，真实集成、独立专业判断、fresh S7 与原人闸保持。

### 7.9 G01 native/teammate 共同资料投影

先固定 L4 dispatch §18.2，再实施：

- agentcontext.Assignment、PlanTemplate 和 InputReadingOrder 成为共同事实/机器计划/底部向上读取的生产者，workspace 消费同一逻辑。专业 objective/steps/assertion_checks/risks 保持未答，approval/one_shot 不生成计划 shortcut；原 SkillRefs、权限和领域门不变。
- 没有独立 execution 的当前已注册1.2 owner，在真实格式 SessionStart/SubagentStart 或 registered team launch 可得 assignment_context_ref（path/SHA256）。严格的内嵌 native-assignment-context Schema 固定 Runtime、基线、实际 source_root/agent_state、Agent/Team/manifest/assignment、原职责/检查/收尾、文档/Skill 来源和计划/schema 入口。专业计划草稿置于原证据路径，保留作者字节；已 working 的视图明确复用原计划/激活，不要求重交。
- 原 REQ authority、committed 文档、manifest 每个已声明 subject、实际 installed Skill 和当前 S5 题干均校验；前后重新核对输入及原 Runtime Snapshot，源变动不返回 ready 视图。资料/Schema 使用内容寻址和原 pinned exclusive publisher，拒绝 symlink/字节冲突；文件名用固定摘要，长 owner/assignment 不进入实际文件名。资料 JSON 不成为第二派发表或激活、领域完成来源。当前仍是可重新投影的 shadow，尚未完成 native/teammate enabled 执行版本冻结。
- registered teammate 只为精确原 Team manifest/Agent/TASK 附资料视图，request-template 每个 document 必须匹配原 manifest 的 ID/path/version/SHA；伪造摘要即使文件存在也拒绝。示例/未注册 launch 不生成已注册视图，1.1 不增加该入口。已有独立 execution 保持其原 Inputs/Bootstrap，不回落 Main。
- native lifecycle 保留原800 rune总预算，只向唯一实际 owner 提供完整入口。S5 原题/答卷路径包含在视图内，避免重复长提示；原接收回执优先，不对已完成 owner 再发计划。真实 FIX 退回 planning 后，未答的旧 peer 不再发行资料或计划，交原 Main rework/retirement 路由。没有增加原生 Write/Skill/Bash 权限，也未改变 Stop/人闸。
- 8角色正文、agent-dispatch Skill、agent-protocol 和 workspace 使用说明同步，frontmatter 精确保留原 HEAD 内容。新内嵌 Schema 不要求安装项目读取工厂源码。

生产 CLI 用例从原注册 S5 基线走实际 register/Hook/原 begin/保存/原 FIX：同一 native/teammate owner 获得相同完整视图/hash；实际文件逐项验摘要；机器草稿无效，补专业字段后走原计划与接收；team launch 保持整个 Runtime/journal 字节不变，native 首次 Hook 的原 Controller milestone 写入单独核对不改变 entities/lifecycle/baseline/evidence。覆盖 unknown actor、独立 execution 禁止 Main 回落、缺 installed Skill、不覆盖作者计划、伪造注册文档与真实 S5 旧 peer 退役。实际模型消费未执行。

完整 CLI 回归395.930s通过（启动时尚未补最后的 S5 retirement guard）；native/teammate/Worker 目标回归38.792s、race202.857s通过。最后退役 guard 连同全部 native/teammate 用例普通27.854s、race90.176s通过；不能把较早 full/race 结果说成覆盖了后补用例。agentcontext/team/schema/hook/hookctx race、runtime/controller/transition/policy/assignment/semantic/releasegraph 回归（runtime55.419s）、workspace 除原宿主受限用例24.603s、最终 go vet 和 diff 检查通过。原 bwrap 限制沿用§7.8已实际复跑记录，没有放宽或回落。

含最终 guard 的未提交工作区重新构建五平台，本地205文件 release manifest/release graph通过；default/--packet 两种全新 native 安装 doctor、validate --all、installed release graph、docs check、原 recover help通过。8角色 native 指引实际安装，frontmatter与原 HEAD一致；示例 readback逐项验证安装 Skill hash且不包含伪造的 registered context。source_revision=ed61728、source_dirty=true，未发布。持久摘要见 [native context 安装证据](evidence/global-optimization-native-context-install-20261013.json)。此 native 安装检查的输入是明确例子 fixture，不是注册业务派发；业务派发由上述 production CLI 用例验证，二者都不冒充真实 Claude 消费。

接续：先固定并实现 native/teammate 的执行 profile/version binding 和首个 Builder enabled 生成路径，再补真实宿主消费/首次有效操作/token 对照；G02本人收口与其余 G01–G08、各领域 adapter 按原依赖继续。安装与视图投影成功不算 token 改善或整项完成。consumer 默认不提升；完整 goal保持active，原独立专业判断、真实集成、required coverage、fresh S7和人闸持续适用。

### 7.10 G01 固定执行资料绑定与首个 Builder 角色生成器

先固定 L4 dispatch §18.3，再实施：

- Main `runtime assignment-context prepare --agent-id <owner>` 将已验证共同资料的 context_ref/generation/profile version/hash/mode 写入原1.2 Agent.context_binding。沿原 Writer/CAS/journal，保留原 last_transition、专业计划/激活与状态；1.1不暗迁移。原资料发布不覆盖作者计划，登记失败不产生激活。普通准备重试和同内容并发只提交一次。
- 已绑定 native lifecycle/registered readback 校验并复用固定引用，不运行当前 selector替换旧profile。原 manifest/dispatch/资料/Schema不符则提示恢复；原捕获 agent_state 与当前 Runtime 状态分开，working不重新生成计划。独立execution声明后拒绝Main回落，保持原Inputs/Bootstrap。
- 显式 `--replace-generation <old>`仅消费原S5具名修订：实际新卷previous_submission_key引用已接收旧卷，原Runtime/基线/进入/owner/assignment保留；profile、Skill、角色源和Schema固定。无前驱、普通刷新、错误代数拒绝；成功后响应丢失重试复用已提交新代数。旧引用、旧目录和作者字节不删除。
- 固定资料绑定亦触发complete-pair checkpoint，即使没有packet/submission。RR-A文档重建拒绝丢失此固定事实。恢复索引校验原CONTEXT-PREPARE journal的owner/generation/ref-key、资料字节、原角色/Skill/Schema。合法文档修订导致旧资料投递过时仍保留此事实，不能重新选择profile或伪称ready；原专业domain/source验证继续适用。
- `RenderCoreRole`及 `runtime assignment-context render-role`从实际安装backend-builder生成四项core预加载候选；仅替换skills列表，其他frontmatter字段和整段权限/责任正文保持字节相同，LF/CRLF均保留。重复/不支持的skills格式、不同角色和未支持的rollout角色拒绝。返回实际source/hash、生成字节/hash、core列表，明确enabled=false。没有改变任何安装/运行角色或授权SkillRefs；生成器不算实际enabled执行。

新增production CLI用例实际register/prepare/原begin/保存/revise：固定引用贯穿reading→working，幂等和两并发只有一个Writer提交；来源漂移不回落并保持原绑定；恢复匹配资产后复用原引用；原S5接收前驱换代后原答卷仍可接收，换代响应丢失不重提交；独立execution后拒绝native fallback。checkpoint真实原pair验收固定索引，损坏资料拒绝；保持完整journal cursor仅改CTX事件/key时，特定拒绝“缺原准备journal”，不以cursor错误冒充此覆盖。packet=false基线仅以原生产register/prepare生成knowledge绑定，未捏造packet/reception，实际checkpoint读/restore校验及race通过。

完整CLI407.057s、native/teammate/Worker相关race314.637s通过（启动早于后补换代重试与render-role）；最终绑定/角色全部目标用例34.014s、race113.595s通过。后补无packet checkpoint普通1.321s、race7.085s通过；agentcontext/schema race、runtime/recovery/schema/controller/transition/policy/assignment/hookctx/team领域回归（runtime45.658s）、最终相关go vet与diff检查通过。较早完整套件不计入后补用例覆盖，后补部分已有独立生产入口和race验收。

最终未提交工作区五平台包完成205文件manifest/release graph，两种全新native安装通过doctor、validate --all、installed graph、docs check、recover help；prepare help显示实际replace-generation/agent-id，实际打包render-role返回可验摘要的四core候选和enabled=false，正文与安装源精确相同，原8角色frontmatter不变。持久证据见 [context binding 安装验收](evidence/global-optimization-context-binding-install-20261013.json)，source_revision=ed61728/source_dirty=true，未发布。例子readback仍不冒充注册任务；业务冻结/修订/恢复由上述production CLI用例验收，真实Claude未执行，bwrap原宿主限制不调整。

下一切片把生成结果沿新Builder execution.prepare/Bootstrap/launch实际装配为首个enabled profile，并保持原授权和来源固定；native/teammate角色装配及真实宿主消费/token/首次有效操作对照继续。随后推进G02本人收口、各领域adapter和其余G01–G08。默认仍shadow/packet consumer不提升；完整goal保持active，required coverage、fresh S7、真实集成、独立专业判断和人闸不缩减。

### 7.11 G01 首个 workspace Backend enabled profile

先固定 L4 dispatch §18.4，再实现：

- 新1.2 S6 backend-builder 的显式 `workspace prepare --context-profile backend-core` 实际启用四core预加载，默认仍shadow。错误角色/阶段/旧1.1拒绝；选项仅prepare可用，错误用法在副作用前拒绝。既有ready/preparing省略选择复用原profile，显式不同值拒绝，不自动改安装或活跃实例。
- 原实际角色生成字节/source/target/hash、原角色source/hash、core集合与profile信息进入冻结context及Execution.inputs。Bootstrap仅在新Worker的backend-builder角色目标使用固定生成字节，原receipt固定它；其他角色/Skill完整安装，额外selected来源以Read可达。ReadDispatchContext验证生成源/原源及非skills字段/正文的精确等价，不按当前selector重新生成旧profile。
- 原 manifest/activation SkillRefs、name/model/tools/disallowedTools/permissionMode、正文权限和领域责任保持；Core selection不传入权限派生，Main角色/安装回执不变，allowedTools仍为Read/Glob/Grep/Write/Edit/Bash。launch使用原登记角色名和固定Worker目标，提示enabled仅控制加载；审批/计划、required_checks、commit/report/集成/发布边界不改。
- 原 execution replacement保留同一profile/Skill/生成字节，仅换代数；旧目录/作者字节保留，源漂移不覆盖Worker。返回context_mode来自实际冻结数据，不能将enabled ready重试报告成shadow。角色正文、agent-dispatch、agent-protocol和使用说明区分shadow与enabled，未改原frontmatter授权。

生产CLI新用例从原注册S5文档基线出发，实际两owner原begin/保存双PASS→原TR-003/S6，再登记真实Builder，prepare仅专业计划字段由作者填写。原计划、受控git commit、真实候选 `go test ./...`、原completion report与Main task-integrate都实际运行；非squash merge与原required check完成后确实将产品改变落入Main。没有手植working/已完成/检查回执或用core元数据代替专业/集成交付。原Main角色与阶段前activation字节精确不变，命令类别比较原envelope；api-contracts原本允许build，未错误把原权限变化当此次扩权。

替身claude通过真实launch/CWD/原session/allowedTools接线，读取实际四core角色文件及全部selected Skill路径并验hash，且提示为enabled。它只证明子进程装配/读路径，不是实际Claude模型或token验收。另验证默认shadow保留Main原角色字节、旧shadow不可retrofit、replace同源、源漂移拒绝、S5 reviewer不可得Builder profile，责任分母/原门没有缩减。

新增完整目标用例79.892s、完整CLI596.439s、core/原Worker/native binding目标race522.031s通过。agentcontext/workspace race通过（workspace29.978s，除原已实际记录的bwrap宿主受限用例），runtime/schema/controller/transition/policy/assignment/integration/hookctx/team回归通过（runtime79.019s），最终相关go vet与diff通过。并行套件成本如实记录，不把耗时增加或core4项解释成已测token改善。

最终dirty工作区完成五平台交叉构建及205文件manifest/release graph，两种新native安装doctor、validate --all、installed graph、docs check、recover help和实际workspace.prepare context-profile选项通过；candidate render仍明确enabled=false，默认8角色frontmatter与原HEAD一致。持久摘要见 [Backend core安装证据](evidence/global-optimization-backend-core-install-20261013.json)，source_revision=ed61728/source_dirty=true，未发布。原业务链由production CLI验证，native安装示例不伪称注册业务；真实Claude消费/token/首次有效操作仍未执行，原隔离限制不弱化。

接续：native/teammate的enabled角色装配及真实宿主消费/成本对照继续；G01仍未整项完成。同时按依赖进入G02本人交付/等待归属与各领域adapter，再推进G03–G08。consumer默认不提升，完整goal保持active；原required coverage、独立专业判断、fresh S7、真实集成及人闸保持。

### 7.12 G02 本人执行收尾与 Main 消费边界

先固定 G02 §7 和 L4 dispatch §10.4.1，再落地只读 ResponsibilityFacts；其来源
仅原 Agent/Assignment 与原 S5 当前已接收事实。delivery_recorded 与
execution_may_end 分开，本人 PLAN_REPORT/Result 欠项、消费者 Main、待消费及
缺具体 closing contract 可区分。报告存在、旧终态/blocked 兼容允许结束都不
等于检查通过，本切片不新增 work_complete 授权或另一个进度账本。

Stop/Idle 复用原本人兼容矩阵。SubagentStop/TeammateIdle 指引标明 Worker 等待和
Main 消费，缺报告时不覆盖为 task-integrate 动作。独立已登记 Worker 的普通
Stop 经原 session/root、精确 CWD、当前 execution/Bootstrap 和本人事实核验后
直接返回，不执行 Main 的 Stop 队列/控制循环。原 S5 发题接收 Stop 单独保留，
Main 原 Stop、required checks、完整批次、独立审查、人闸仍保留。

TASK §6 分本人交付/后续消费两表，保留独立 DV/QA/E2E 与真实集成责任；AGENTS
和 loop-orchestration、agent-dispatch、agent-protocol 按当前责任收尾。未决风险
须带具体证据/影响走原 BLOCKER，不能将范围外风险自动免责或扩大 Worker 权限。

新增原 S5→S6 Builder 实际交付链断言：缺 Result 普通 Stop exit 2 且不写游标；
Main CWD 不冒充 Worker；原 loop guard 保留；报告登记后只读本人事实指向 Main
待集成；无自造 agent_id 的顶层 Stop 使用原 session/root，TeammateIdle 使用
官方 teammate_name，均允许结束；普通 Stop 不写 Main 游标。随后原非 squash
merge 和 required Go test 仍完成。投影矩阵覆盖计划/报告/接收/ack/旧终态/
queued/批准等待/BLOCKER/其他 owner，不把这些用例解释成真实模型实验。

投影矩阵与 hook/hookctx/policy race通过；原职责链目标36.813s，普通 Worker Stop
目标37.645s，职责链 race163.716s通过。首次完整CLI494.556s发现三处缺失原
SubagentStop事件标签的输出回归；保留新语义并补回标签，复测23.435s、完整CLI
474.074s通过。其后最后的只读指引收紧（兼容终态/BLOCKER不声称报告已记录，
具体集成命令须有worktree和completion）另测21.451s通过；证据明确完整套件
与最后指引用例的快照边界。Controller/Transition/Assignment/Integration/Review
回归、最终相关vet与diff通过。

最终五平台构建、205文件manifest及两种新native安装的doctor、validate --all、
release graph、docs、固定profile来源和新职责模板检查通过。见
[本人收尾安装证据](evidence/global-optimization-owned-completion-install-20261013.json)；
dirty source=ed61728，未发布、未覆盖活跃安装，真实Claude/token实验未执行。
最后包sha256=4e4a4b84376a1d50f4613a1ff280be9b6a3063a538f3f93838857ba1f6d0acf4。

后续仍需领域 work_complete
与 checks/输入失效/返工责任回执、缺收尾契约的新派发约束、各 adapter 和真实
模型 A/B；G02 和完整 goal 保持未完成。

### 7.13 G02 S6 当前责任回执、owner 登记与失败接收恢复

先固定 G02 §8 和 L3-S6 §7.3，再新增只读 `runtime task-receipt` 与闭合
builder-receipt Schema。复用原 qualifiedEvidence、原报告引用/索引/哈希、当前
TASK 指纹与 TASK/Agent/assignment owner 绑定；区分 delivery_recorded、accepted、
work_complete、integrated，空/失败/缺漏 required checks、scope deviation、缺
closing contract 明确列本人欠项。checks_source=original_report，查询没有执行
新检查，也不把报告状态或任意 assert 文本伪装为行为真值。

本人合格、Main 未 verified 集成时本人完成而 Main 待消费；原 report-bound
checkpoint 校验 TASK/assignment/generation、报告 hash/时间和 bound branch
真实可达性后才投影 integrated。原登记报告损坏/失效、TASK 输入改变指向 Main
恢复/原返工；不让作者重复承担系统恢复。输出前复核 committed view、当前读源
与 Runtime snapshot；不缓存授权，不写状态或推动门。S9 修复 Actor 明确拒绝
用 S6 报告冒充 RepairResult 完成。

真实原 S5→S6 Builder 链发现并修复两处原接线缺口：原 register-workgroup 的
同一 Writer/CAS 对既有 TASK 加入明确派发的 Builder owner，保留旧 owner 及其余
TASK 事实；原 task-integrate 在 Inspect 拒绝时通过 Integrator !Ready 保全路径
落 durable preserved checkpoint，再写 Milestone。保全前不 checkout；不存在
worktree 的失败也能记录恢复事实，既有 verified/merge 后恢复保护仍保留。

prepare 对 S6 Builder 返回精确 builder_receipt_command，Worker 根路由核验
本人 owner/assignment；Hook 只接纳这条固定只读适配命令，组合 shell 仍拒绝。
父命令不截走此查询的 leaf help。agent-protocol、agent-dispatch、工作区说明
和生成 manual 同步实际路径；S5 原 answer receipt、S9 原完成及权限不改。

实际失败检查→原报告→失败集成保全→Main 原 rework→旧 ref 清空/历史保留→
修正后真实检查/新报告→Main 实际 merge/required check→当前 integrated 回执
完整链已通过；另验证报告损坏转 Main 恢复、Worker 拒查其他 owner、只读不写
游标、真实 Hook 固定命令/组合命令、完整接收/检查点矩阵和闭合 Schema。

原职责/返工链43.333s、真实 Hook/query链29.657s通过。首次完整CLI537.074s
发现保全前checkout缺口；修复后边界/业务链41.379s、完整CLI503.955s通过。
随后最终额外committed-view复核另测49.446s通过；证据记录完整套件与最后
read-only source recheck的快照边界。职责返工/显式接收race110.519s，最终
Integrator/qualitygate race9.512s/3.974s，qualitygate/policy/assignment race
3.970s/6.580s/67.051s通过。真实Git检查点矩阵0.299s、Schema闭合矩阵与
领域回归、最终相关vet/diff通过；未将原报告checks说成查询执行的新检查。

最终五平台构建、205文件manifest及默认/packet两种新native安装doctor、
validate --all、release graph、docs、leaf help、冻结profile/来源和新资料检查
通过。首次help拦截已修复。见
[S6本人回执安装证据](evidence/global-optimization-builder-receipt-install-20261013.json)，
package sha256=95224ce41b18a088736834ab2b243ab37be7d9009b9e18c42de4b9828929fc1e，
dirty source=ed61728；未发布或覆盖活跃安装，真实Claude/token实验未执行。

本切片不代表
S6 typed WorkPacket adapter、其他领域责任回执、G03–G08 或真实模型消费/token
实验完成；完整 goal 保持 active，原批次覆盖、独立审查、fresh S7、人闸保留。

### 7.14 G03 原 TASK 的版本化读集与粒度分析

先固定 G03 §7，再实现 task_context.go，并接入原 TasksCheckResult：保留原
reference_loads 字符串，增加 context_analysis_version=1.0.0、task_contexts、
granularity_advice。原 Problems/覆盖/DAG/取消分母及出口判定不由 advisory 更改。

每条 Manifest 行保留 Mode、条件未知、Purpose、Clauses、TASK 来源与 path/anchor。
旧行明确 legacy_unspecified；optional/conditional 不冒充 required。可定位的
Markdown 标题/中文 slug/inline 或标题前显式 anchor按章节字节统计；代码块内
标题忽略，重复/缺失/非 Markdown anchor保守计全文并说明原因。按同一文件的
字节区间并集去重，分别列 mode 和 total；目录不递归展开、缺文件/解析超限/
越界明确不可解析，Bytes=null不伪装成功0字节。分析缓存文件内容，1 MiB仅为
解析上限，不能伪称既有 Reader I/O 已限流或计过实际token。

原 committed/disk file-source 契约保留；裸路径沿原repo-relative规则，链接沿
原sharedmodel.Resolve。磁盘symlink出Root拒绝，已声明git_tree不被Working-tree
symlink替换为另一事实源。Closing Contract仅提取明确已声明检查入口，未执行
命令或assert，未知/占位/自然语言断言保留为专业欠项。

同文件写入、同required读集且scope重叠、多写入路径组和缺具体检查入口给带
TASK/路径/来源的问题。没有任意bytes/token阈值，没有自动拆并、改scope或DAG。
S4 tasks check --json与S5原独立workspace review-check实际返回同分析；文本只
给短摘要。规划/审查Skill改为按独立验收、实际读集和变更耦合判断，去除30KB/
一次compact/可拆即缺陷的自动结论；S4蓝图同步现状。

semantic/sharedmodel/transition/qualitygate回归4.102s/1.661s/14.780s/0.603s，
semantic/sharedmodel race14.185s/3.617s，原S4/S5入口目标7.467s和race27.158s
通过；最终heading/cache细化后的parser与入口0.078s/14.193s通过。完整CLI
538.015s通过（在最终heading/cache细化前启动，后者有独立最终用例），最终
相关vet/diff通过。证据明确每轮范围，不用静态字符串验证声称模型成本改善。

最终五平台构建、205文件manifest，两种新native安装doctor、validate --all、
release graph、docs、实际consumer Skills/固定来源和原S4 task JSON fixture
通过。初次复制平台binary因磁盘满失败，只清理明确本任务过期临时安装副本后
重新安装通过；原live Go regression没有重启。见
[TASK context安装证据](evidence/global-optimization-task-context-install-20261013.json)，
package sha256=3836b92644fa9268a1c9bef92c60efa368f126402542e3fa7f205a8f9a578660，
dirty source=ed61728，未发布/覆盖活跃安装。

G03行为实验、正式WorkPacket额外冻结分析及解析器
版本、已批准合同/当前CASE样例投影仍待接续；当前Scope只证明分析及原S4/S5
入口接线，不证明compact/交接/token改善。完整goal继续active，其他G01–G08与
typed领域adapter、原独立专业判断、真实集成、fresh S7、人闸保持。

### 7.15 G03 新 S5 Packet 冻结分析与原接收/修订/恢复链

先固定 G03 §8，新增可选 task_analysis ref及闭合artifact Schema。新原issuer
仅对原登记TASK集合生成版本1.0.0读集/advice，内容寻址发布，引用进入packet ID
和Draft.read；旧已发行packet/原v1身份不补字段。正式inputs及answer/domain
envelope责任范围不扩大，没有从advisory自动推出专业结论。

原ValidateWorkPacket在原发行/保存Writer锁内检查分析hash/schema/parser/完整
formal refs及可读来源hash；不按当前parser重算旧分析。具名revision、checkpoint
index恢复重建使用同一可选冻结引用；历史checkpoint读取原不可变artifact字节，
丢失/损坏拒绝并要求恢复。原场景没有分析引用时仍严格沿原投影。

原PlanDocumentReview把分析与可读来源纳入Execution.inputs；原Bootstrap安装
到.claude/inputs并由原receipt固定。原执行视图额外校验这些字节；S5 review-check
从原分析读取context/advice，机械Problems/覆盖/DAG继续原求值。工作区说明与
document-verification Skill同步冻结/漂移/旧包口径，不更新活跃安装。

新增实际S5发行→prepare/冻结安装→原begin/answer接收→Main具名revision保留
同分析→损坏分析发行重试拒绝且不写游标→恢复原字节/幂等重试完整用例11.991s
通过；旧v1 identity/typed identity/正式subjects未扩大/具名binding重建/任意
analysis path拒绝单测通过。原S5 receiver/本人Stop目标23.709s，接收/修订race
133.792s，runtime/workpacket/workspace回归26.787s/0.034s/7.461s、runtime/
workpacket race140.845s/1.408s，完整CLI556.009s、相关vet/diff通过。

最终五平台构建、205文件manifest及两种全新native安装doctor、validate --all、
release graph、docs、冻结分析consumer资料与实际S4 fixture通过。见
[冻结TASK分析安装证据](evidence/global-optimization-frozen-task-analysis-install-20261013.json)，
sha256=dc3a9abbf5eac1039f88bde81ad9c89c2ddc08f1330fe66f0c013a258c3b04e0，
dirty source=ed61728，未发布。原bwrap宿主受限未伪称通过；真实模型消费/token/
compact/首次有效操作没有执行。批准合同/当前CASE样例、S6等typed adapters、
其余G01–G08和真实验收继续，完整goal保持active；原独立专业判断、完整覆盖、
真实集成、fresh S7及发布人闸保留。

### 7.16 G03 当前 CASE 行为样例与原 S5 冻结来源闭包

先固定 G03 §9；S4 只从 TASK Manifest 明确 cases.json 路径与 Clauses 的 CASE
引用选择样例，不扫描全仓库、不把同模块所有 CASE 推为本人责任。原 Scenario
source/schema/generator 验证当前 cases/coverage 等于源生成结果，保留原 CASE、
rule/branch/polarity、required/browser_required、oracle/fixture；输入仅解析
原 witness 的 fact partition.value，没有执行测试/浏览器或推测预期。

读集 Source/Mode/Purpose/condition_unknown 保留，同 CASE 不同 Manifest 行
各自保留出处，conditional 不变成必需。无 CASE 引用、未存在 CASE、失效生成
或未验证源包给 unresolved 问题，不新增 TasksCheck Problems 或自动开工门。
样例来源闭包记录七个模块文件及原校验实际读取的绑定 REQ/FR，hash 是实际
读取字节的 SHA256；中途同路径漂移拒绝。原 FR join 仍严格验证当前绑定 REQ，
不退化为 shape-only，也不把 registry/例子存在当新需求批准。

新冻结 artifact 可选 behavior_projection_version/behavior_examples，内容
hash 固定该投影；旧分析不补字段。原 Packet formal inputs/结论范围不扩大，
原 Execution/Bootstrap 安装闭包。原 review-check 返回冻结样例，原 validator/
answer consumer 核验来源，具名修订保留同一分析。规划/审查 Skill 和工作区
说明同步参考样例、未决责任、真实执行与冻结来源的边界。

端到端用例验证原发行、独立 prepare、完整来源安装、原 begin、review-check/
answer 因来源漂移拒绝且 Runtime 未变、恢复原字节后接收及具名修订；并覆盖
未知/缺 CASE、手改生成物、投影中途来源变化、条件和不同出处保留。最终包
五平台交叉构建、205 文件 manifest、两种全新 linux/amd64 安装与真实 Scenario
生成/S4 投影通过。完整回归状态及各轮源码范围见
[CASE 样例安装证据](evidence/global-optimization-task-behavior-install-20261013.json)。
完整 CLI 597.054s 通过（启动后最后一行读集出处去重细化另以最终目标35.824s
及 race84.129s验证），原接收/独立来源 race242.118s、runtime race158.406s，
最终 scenario/semantic/workpacket race14.767s/16.631s/1.254s、vet/diff通过。
最终 package sha256=6c316065e6a72ae4978ac0ff2b7ca5598e7892a050a1e070843018ca5d2d6003，
dirty source=ed61728，未发布或覆盖活跃安装。
真实 Claude、浏览器产品结果、token/compact 对照未执行，不由静态输出替代。

下一步 G04 §7 已固定首个候选卡、原来源/批准验证与 registry 切片顺序；resolver
尚未实现。G03 实战对照、其余 G01–G08、各领域 typed adapter 与完整验收继续，
完整 goal 保持 active；原独立专业判断、完整覆盖、真实集成、fresh S7 和人闸保留。

### 7.17 G04 原来源惯例索引、S4 投影与新 S5 冻结链

先固定 G04 §7–§8 的来源/批准/输出契约，再落地闭合 registry/projection Schema、
空模板和原来源 resolver。按明确 TASK scope 的目录边界选择条目；无 registry
只为本人明确引用的 API/错误处理 Rule 提供候选，不扫描代码推断默认或栈。

当前 declared_default 必须符合原 locked Rules 索引/Rule ID/hash，或命中原
登记 locked 架构/决定的 id/path/version/hash，且默认文字实际位于唯一来源
段落。自写 approved 不证明批准；缺来源、过期示例、重复 ID、重叠 scope 的
不同默认及 migration/superseded 均不能成为当前选择。仅核验来源一致性，
语义适用性、例子是否吻合及偏离仍交原专业责任，不补业务值域或权限。

每批共享一次 bounded source/schema 缓存，逐 TASK 保留实际闭包。disk 检查
真实路径越界，git_tree 保持原提交字节；正式示例限定原 bound git_tree，不
因原声明 disk 就偷偷替换来源。调查修正 ScopedPlanningFiles 的 authoring
来源误标，新增原 document-review prepare 的冻结摘要一致性比较，原 dirty
输入拒绝、激活、权限、coverage/DAG 和门继续。

S4 JSON 返回 task_conventions，新 S5 原 issuer 用 formal refs 求证并冻结
同一 projection/解析器/真实来源；原 Bootstrap 安装和来源校验、answer 接收、
具名修订/恢复复用原分析链。旧 artifact 可选字段不补齐，原 formal inputs/
结论范围不扩大。Rules/project-map/规划/审查 Skill、三个 Builder 的条件化
阅读说明和工作区文档同步；实际 Builder/native 投影入口仍待 G04 §9 接线。

单测覆盖 absent/候选、来源段落、原登记版本、冲突/重复、模块边界、迁移、
过期示例、非法字段、共享读缓存、隐藏路径拒读、磁盘越界与提交树抗 symlink
漂移。原 S5 真链验证冻结 Rule/registry/index/提交树示例、dirty prepare 原门
拒绝且 Runtime 未变、独立消费、来源漂移原 answer 拒收、恢复接收和同分析
revision。最终五平台构建、206 文件 manifest，两种全新 native 安装及真实
S4 候选/默认/冲突投影通过，原 TASK verdict 不变。完整终态和源码范围见
[项目惯例安装证据](evidence/global-optimization-project-conventions-install-20261013.json)。
完整 CLI 723.201s 通过（在最终 formal-example guard 前启动，后者有最终
目标49.939s与race140.453s），最终 resolver/semantic/workpacket race2.657s/
20.132s/1.294s，原 runtime race178.185s，相关 vet/diff通过。workspace 全部
尝试仅原 bwrap UID-map 用例失败；仅排除该已记录宿主限制的回归20.190s通过。
package sha256=7f54a9511e804b004b81f11bbdbc8910fc0856f192176a1d8f1b012e878c04f8，
dirty source=ed61728，未发布或覆盖活跃安装。

原 workspace bwrap UID-map 宿主限制仍如实记录，隔离不弱化。真实 Claude、
浏览器执行、模型成本/风格返工对照未执行；当前测试不代替这些验收。G04
§9 固定下步 Builder/native 的原来源冻结契约，整个 G01–G08/typed adapters/
独立专业判断/required coverage/真实集成/fresh S7/人闸 goal 继续 active。

### 7.18 G04 实际 Builder/workspace 与固定 native 惯例上下文

先固定 G04 §10，再在原 context 中增加可选 conventions bundle。S5 复用原已
发行分析，S6 从原唯一登记 TASK 路径按 Task/Assignment 共同范围求证；原
generation 的完整 formal refs 提供架构/决定来源，不从 registry 状态批准。
常用规范前缀与末尾 /** 可证明交集，其余未证明模式列具体问题，不扩大权限。

真实 registry/index/规范/示例摘要进入原 frozen_inputs、Bootstrap 和 native
input_read_order；bundle 校验版本/TASK 路径/来源关系，不重新选择默认。
发布及原登记 Writer 锁内复核 catalog.file_sources、开发分支和真实字节；
native S5 的原发题校验也使用同一 source contract。固定 native 重放验证原
formal proof、旧 ready/context/analysis 缺字段不补，不换 profile/授权 Skill。
checkpoint 核验不可变 bundle 与原输入关系，不伪造历史专业结论。

实际 S5→S6 fixture 经原 prepare/launch/begin/受控 commit/真实 Go 检查/report/
非 squash Main verified integration 通过，包含本人修改已引用示例源码；
native S5 与已发行分析逐项一致，native Builder 固定来源/真实读序/重放/漂移
拒绝且 Runtime 未变通过。启动替身只证明进程装配和实际接收到的上下文，
不证明 Claude/模型成本。原计划、激活、权限、原检查/集成及 fresh S7 不变。

最终五平台交叉构建、206 项资产、两种全新 native 安装及旧 Skill/frontmatter/
来源与新协议说明验证通过。各轮源码范围和完整回归终态见
[实际惯例上下文安装证据](evidence/global-optimization-convention-context-install-20261013.json)。
最终入口目标40.299s、race159.866s，原 checkpoint/替换目标73.259s、相关
runtime race161.085s及vet/diff通过。首轮完整CLI在686.090s以原checkpoint
发布 ENOSPACE 终止，未记通过；只清理已确认本任务过期安装副本后，同一
完整范围在该切片最终源码重跑586.579s通过，终态记入上述证据。归档/源码/证据/活跃测试
目录保留；原pending保留和隔离边界不改。
真实模型/风格对照、其余 G01–G08 和所有剩余 named adapters/完整平台验收
继续，完整 goal active，required coverage、独立判断及人闸保持。

### 7.19 G07 Main 检查入口预检、只读能力与真实故障分类

先固定 G07 §7，再修正原 Main PreflightChecks 直接返回成功的缺口。原
prepare/launch 发布 ready/预留 session 前检查继承环境的实际 sh 和简单
字面检查工具入口，明确缺失时拒绝；不执行检查/安装/版本脚本探测。复杂
shell、变量/控制流明确 unverified，fully_verified=false 不伪称完整依赖/
配置/fixture/服务已验证。locked: 非可执行职责保留，旧 profile 不启用复用。

原 ObservationCommands/Hook 精确入口新增 check-capabilities，只读且重复
原 owner/assignment/路径/frozen inputs/Bootstrap 校验；Worker 继续真实
Bubblewrap/seccomp 预检，不变成 Main 无隔离退路。公共 typed Failure 保留
code/category/owner/观察/修复/重试条件/receipt/log 及原 error unwrap 链；
真实 Main/Worker 回执兼容旧字段，process_exit 不能自动解释为产品断言
失败，取消/超时/诊断 I/O/input_drift 保留原 checkpoint 与候选。

用例证明字面/引用空格中文路径、缺工具/缺 shell、复杂语法未验证、预检
不执行工具、原 prepare 缺工具时 Runtime 未变，以及真实非零退出/超时/
诊断存储故障和回执来源。相关领域/race 与原 Builder actual verified
integration 通过。原生安装发现 observe --help 错误退出2，已修正并以最终
入口及原生验收证明成功。最终五平台构建、206项资产，两种全新 native
安装和能力入口/协议资产通过；源码范围与完整回归终态见
[检查能力安装证据](evidence/global-optimization-check-capability-install-20261013.json)。
完整CLI671.261s通过（启动后仅 help 退出码修正，以最终目标20.522s和native
help另验），原入口/真实集成目标44.945s、race200.439s，checkexec/integration/
workspace race1.119s/8.398s/12.868s，相关vet/diff通过。workspace仅排除先前
已明确的原bwrap UID-map宿主限制用例，隔离不弱化。最终package sha256=
94b614eaffdb6bc6e693115b745e5227b233d46bc0117bbb50bf4077ca10dfa5，dirty
source=ed61728，未发布或覆盖活跃安装。

完整 CheckProfile、依赖/服务/资源/环境指纹、G05 正式适用性与受限复用、
所有其他 G01–G08/named adapters 和真实模型/平台/实战验收继续，完整 goal
active。required checks、真实 Main 集成、独立判断/fresh S7 和人闸保持。

### 7.20 G07 显式 CheckProfile、实际指纹与 v2 观测回执

先固定 G07 §8，在原 manifest Assignment 增加可选闭合 CheckProfile。原
required_checks 索引/命令必须一致，新1.2执行固定 location、1–300秒timeout、
环境政策与来源，不增加第二命令。旧 Runtime 不登记新 profile、旧 ready
execution不补字段，原 replace/adopt/rebind 保留同一原资料。

依赖/fixture 的真实路径和摘要进入原 frozen inputs；环境仅批准白名单，
unset与空字符串区分，declared_only不继承全部环境。原 Main/Worker runner
执行同一检查并采集实际受测树/runner/白名单环境/依赖/fixture指纹及原
Runtime/generation/Assignment/owner/责任，时限不超原外层截止。原 workspace/
native context 暴露冻结profile，来源与原责任校验保持。

profile-backed v2 回执保留旧字段、真实失败、日志摘要和 corruption integrity；
输入/环境/runner变化不能PASS，非零退出、取消/超时与诊断故障沿原现场恢复。
完整性值不是可信接收；当前 reuse_eligible=false、所有检查仍实际执行，
Main required/post integration、独立判断与fresh S7不因profile省略。

实际S5→S6 prepare/begin/受控commit/真实Go检查/report/Main verified integration
及profile责任回执、替换保留、native固定暴露/重放通过，相关领域/race通过。
最终五平台构建、206项资产，两种全新native安装与协议资产通过；各轮源码
范围和完整回归终态见
[CheckProfile安装证据](evidence/global-optimization-check-profiles-install-20261013.json)。
完整CLI768.699s通过，原profile实际集成/替换/native目标59.978s及race199.523s
通过；checkexec/integration/assignment/runtime/workspace race1.316s/10.908s/
59.089s/146.740s/15.240s，最终observation Schema追加后目标0.050s/8.723s
通过。仅排除原bwrap UID-map宿主限制用例，隔离不弱化；vet/diff通过。最终
package sha256=4ff066c301d5f9f22fd855971782e9ca5044c3f2b6ac691db441b2627e2f7b0a，
dirty source=ed61728，未发布或覆盖活跃安装；终态验证UTC日期为2026-10-14。

G05 原工具回执可信登记、完整适用性/同责任复用键、并发/中断恢复仍需落地。
继承环境只记录声明事实，不能声称完整复用闭包；真实Worker隔离宿主限制、
模型/平台/实战对照仍如实保留。完整G01–G08/named adapters/coverage/fresh S7/
人闸goal保持active，不因当前采集和安装测试缩小完成定义。

### 7.21 G05 原工具回执来源与 shadow 适用性（2026-10-14）

先固定 G05 §2.1/§3 与 L4 的原接收边界，再接入实际 runner 捕获：摘要来自
实际生成的内存字节，回执原子发布后才给进程内接收器，不提供用户回执导入。
Worker 经原 Writer/CAS 复核 Runtime/generation/owner/execution/profile/激活
职责后登记，Main 在原 integration checkpoint 保留工具引用。登记只表达工具
事实，不造专业结论、gate PASS、责任完成或 fresh 判断。Worker 索引至多64项，
原 checkpoint 恢复核对原 reception journal；幂等重放复核当前身份。

共享核验要求登记摘要、规范化 integrity、原 binding/profile 和日志字节均
一致。改写后重新封印、截断、日志漂移、未登记来源和另一责任不能作为新
复用来源。原 rebind 只移动 evidence 坐标，保留原字节/摘要/checkout，不授权
跨 checkout 复用；旧领域回执不强制迁移，原 S9 接收条件保持。

`check-reuse` 为原固定 ObservationCommands/Hook 的只读投影。首切片一律
action=run；完整同责任输入匹配仅报告 shadow would_reuse，未知权限/隔离
条件不命中。Main 必需检查、独立责任、新假设和 fresh 均重跑，所有 runner
仍实际执行。此切片不是 G05 启用复用或并发去重完成。

原 Writer/CAS/重放/checkpoint 来源、真实 Main 检查及 S5→S6 Go 集成、重封印
拒绝、输入/责任边界和显式搬迁测试通过，相关领域与 race 通过。五平台构建、
206项资产及两种全新 native 安装通过；完整 CLI 回归在用户要求停手时中止，
未取得终态 PASS；最终范围
与结果见 [G05 首切片证据](evidence/global-optimization-check-receipts-install-20261014.json)。

下一切片仍需闭合 Worker 隔离/权限与实际命令环境键，在批准纯检查上完成
shadow 对照后接入同职责启用、同键并发/中断恢复。真实隔离 Worker、模型与
平台/实战对照未被本次构建或单元 fixture 代替；整体 G01–G08/named adapters/
独立覆盖/fresh S7/人闸整体工作未完成；用户已要求停手，goal 为 paused。
未发布或覆盖活跃安装。

### 7.22 G05 原权限指纹与闭合静态 shadow（2026-10-14，目标恢复）

用户已恢复目标，goal 为 active。先固定 G05 §2.2 与 L4 契约，再移除调用者
PermissionKnown 布尔值：新 Worker 观测绑定实际 Bubblewrap、shell/ELF loader/
shared libraries、原 seccomp/argv policy、platform/kernel、uid/gid/groups、
Main/common/Worker/cwd 与原 scope/checks/inputs/Bootstrap authority。采集与
原执行共用 argv/filter 生产者，Bootstrap 改写在执行前拒绝，权限/动态库/
策略变化及旧回执缺字段不能命中。继承环境继续标未知，不提升 runner。

首版闭合命令只接受 true/false 和受测树内 test -f/-s；arbitrary shell、外部
程序、隐藏控制/环境路径、expansion/glob/逃逸路径不能仅凭 whole_tree 声明
取得资格。只读 shadow 与 RunCheck 使用同一实际临时副本及归一化模式。
依旧 action=run/reuse_eligible=false，Main 必需检查、独立责任与 fresh 保持。

领域/相关 Schema 与 race 通过；实际 ELF/filter/copy/Worker 尝试和 Bootstrap
篡改测试通过。真实 Worker 尝试保留宿主 bwrap UID-map Permission denied，
没有未隔离 fallback，也未声称支持宿主的成功 Worker 或真实模型验收。
五平台构建和两种全新 native 安装通过；完整 CLI 首次因 ENOSPACE 失败，
原失败用例10.251s复跑通过，tmpfs全套复跑759.424s通过；结果见
[权限切片证据](evidence/global-optimization-check-permissions-install-20261014.json)。

首轮构建因磁盘不足失败，改用本任务可执行 tmpfs 临时目录后成功；本目标
临时归档逐个校验 SHA256 后完整转存，原路径保留链接，未删除共享缓存。
全部旧归档/JSON/失败日志保留，临时存储可能过期。

后续仍需同职责显式启用、同键并发/中断恢复、支持宿主准确性与实际对照，
再继续 G06/G08 与各阶段 adapter；整体完成定义保持，未发布或覆盖活跃安装。

### 7.23 G05 显式静态复用、同键锁与中断诊断（2026-10-14）

先固定 G05 §3.1 与 L4，再新增原 profile 的可选 reuse_mode；缺省 shadow，
enabled 仅原 Worker 闭合静态语法与 declared_only/none/whole_tree。原 Schema/
manifest/execution/native context 冻结同一模式和摘要；旧 shadow 回执不升级。
实际 RunCheck 用当前副本/权限/原登记记录决定是否返回旧引用，不新造 PASS；
Main required integration、独立判断、S9/fresh 与新假设均重跑。

原进程锁串行同责任/profile/源树请求，原 Main integration 锁及 Writer 接收
仍有效。竞争返回有界 run 引用；取得锁后重读原 Runtime/执行/owner/S9
authority。诊断为 waiting/running/finished/interrupted，缺失或损坏不能成为
缓存来源；真实子进程死亡释放锁并保存 interrupted，无完成或 PASS 推断。
原固定 check-running/Hook 观察最多64项，不作为存活证明或第二调度状态。

闭合/启用政策、shadow 不升级、fresh、进程竞争和真实 kill/recovery、相关
领域与 race、Main实际检查/原 Writer目标回归通过；五平台构建及两种新安装
通过。真实隔离失败保留，支持宿主 exact reuse 正向用例因 UID-map 限制 SKIP，
未把该条件跳过记为真实复用验收。完整 CLI 855.394s通过；具体源码范围与终态见
[显式复用切片证据](evidence/global-optimization-check-reuse-install-20261014.json)。

仍需支持宿主上实际 shadow/enable 准确性与并发只运行一次完整链、真实模型/
平台对照，以及 G06/G08 和具名阶段 adapter。整体 goal active，未发布或覆盖
活跃安装；原范围、独立覆盖、fresh S7 和人闸不缩减。

### 7.24 G06 原接收优先建议首切片（2026-10-14）

先固定 G06 §2.1 契约，再落地原 `s6 status --advice`。JSON 保留原 Board，
追加 reception_advice；原 pending 仅追加 TASK ID/checkpoint 引用。优先接收
释放更多直接消费者的原结果，同分稳定排序；具体 argv 沿原 pending，专业
纠正理由作为 question。CLI 前后校验 Runtime 身份/generation/revision，
读失败或过时提示刷新。原 Main checkpoint guidance 消费同一只读投影。

实际容量须显式声明，未知预算为 null，零为无槽位。
`--max-unreceived-results` 只限建议列表，未接收、失败、verified 待 ack、
cleanup 分列；cleanup 不占该上限。原 next、TASK/DAG、注册、权限、Main
实际集成检查与 required coverage 继续由原入口负责。默认 status 保持原形状。
该切片尚未实现 Runtime 持久反压/滞回或进程存活观测。

原 CLI 容量/建议上限只读契约及真实 S5→S6 Builder 报告→Main Go 检查→
集成完成链已通过，验证原 argv 与预算释放。领域 race 已通过；CLI race
首次遇到 S5 前置转换未完成，同一链独立重试通过，原因尚未确定并保留
初次失败。系统回归发现原启动预检问题被短上下文挤出，以及接收建议未进入
Main SessionStart 字段，已修复并补原生有界上下文验收。系统包与 Hook race
已通过；成功 Worker 隔离子用例因宿主 uid-map 禁用明确排除，未放宽隔离。
最终代码五平台构建、206 资产及默认/packet 两种全新安装通过；包 SHA256
为 3ad85e31b726acae26f599cac0c06d3751cc76295a251bfc896b8482a0db43c7。
完整 CLI 最终回归已取得 PASS 724.533s；其他 53 包通过，系统包修正后
PASS 136.136s（另有两个成功隔离用例因当前宿主限制明确排除）。初次失败/
重试及最终来源摘要均保留，真实模型/吞吐未执行。结果登记在
[首切片证据](evidence/global-optimization-reception-install-20261014.json)。
全局目标仍 active，后续继续 G06 配置/观测/吞吐、G08 与具名阶段 adapter。

### 7.25 G06 初始化冻结的持久反压与 ACK 报告绑定（2026-10-14）

先固定 G06 §3.1，再落地可选 dispatch_backpressure 配置：仅新 1.2
init/install 明确选择触发/解除阈值，初始化重试须匹配；普通 Writer/reset
保留冻结配置。原 Runtime 的 dispatch_pressure 只持久化身份、代次和迟滞，
原 Builder 报告与 Agent ACK 事务维护；读操作不写状态，不新增队列。

原接收投影共用 integration.Pending 和 HookContext.FromSnapshot。计数按
原 Assignment，含失败/待 ack；当前已确认报告与 checkpoint 摘要匹配才
解除，cleanup 单列。原 ACK 事务记录已登记的本人报告摘要，重交清除旧
ACK/摘要并重新 reported。原 TASK 全局最新报告不得覆盖另一个 owner 的
canonical 指针；无法解析 Builder 来源时不把未知计数当零。

仅启用配置下新的 S6 Builder 注册在原 Writer 内暂缓；默认/旧 Runtime、
其他阶段、已登记工作、原依赖/冲突/owner、实际 Main 检查、独立性和 required
覆盖保持。来源解析共用原只读作业授权实现，assignment API/原验证链保留。
未实现自动取消在途工作、平台 spawn、运行存活观测或吞吐对照。

原 CLI 配置/中断重试和真实 S5→S6 报告→注册暂缓→Main Go 检查/集成→
解除链，以及领域/CLI race 已通过；真实 ACK 保留后的 checkpoint cleanup/
摘要失配模拟补充通过，不把模拟称为真实 cleanup 故障。五平台、206 资产及
packet/持久反压两个独立开关的四种全新安装通过；包 SHA256 为
46200f13f439777ce978baeb70e14875a17efde933b845bf749a17e034478d33。
完整 CLI PASS 1082.278s；其余包完整回归取得终态 PASS，系统包
PASS 199.729s（两个当前宿主不支持的成功隔离用例明确排除）。后追加的
cleanup/摘要失配用例单独 PASS 32.326s，未混称真实 cleanup 故障或吞吐验收。
详见 [持久反压切片证据](evidence/global-optimization-pressure-install-20261014.json)。
整体 goal active，切片不代表真实模型/宿主复用或全局优化完成。

### 7.26 G06 原 OS 工作区锁的只读运行观测（2026-10-14）

先固定 G06 §4.1，再接原 filelock/LockWorkspace：只读打开已有普通锁文件，
非阻塞探测实际 OS 锁并释放，前后检查同一 inode；不创建/写锁、Runtime、
journal、checkpoint 或第二份运行记录。缺失、可用、占用与未知分开；闭合
integration-observation Schema 明确瞬时、authorizes=false，不推断 PID/具体
Assignment 身份，也不填 PASS/完成。

原 Main `runtime workspace integration-status` 与 s6 advice/Main checkpoint
接线，busy 时给 observe_integration 的完整只读 argv，保留 pending/兼容候选。
实际执行仍由原锁/Writer/领域门负责。未知给来源诊断，普通工具/Stop/Idle
不新增扫描或 gate。修正 controller 外层只对有派发计划接线的旧条件，默认
无计划/无配置/未占用仍不添加接收建议。

真实子进程持锁/死亡释放、只读/损坏/链接行为已验证。真实 Main Go 必需
检查在受控等待点期间，原第二次集成被锁拒绝，没有重复检查；释放后原 Go
执行与接收完成。测试 wrapper 最终执行真实 Go，不伪造结果。初次原生提示
遗漏的失败已保留，修正后真实链/提示重试 PASS 36.481s。

领域 race 通过，CLI race 首次在 S5 前置转换失败，独立重试 PASS 122.096s，
未确认重复负载失败原因。五平台/206 资产/四种新安装通过；初次帮助信息
stdout 断言问题仅修正 smoke 脚本，失败记录保留。包 SHA256 为
3c329bf18820522a33ea2b54495ce0b9cb68de0eae4ee0e6687b30cd35e10627。
完整 CLI PASS 1211.371s，其他包回归 PASS，系统包 PASS 181.918s；两个
宿主不支持的成功隔离用例明确排除。后续 S5 测试夹具诊断/接续改动独立
验证，不混称该完整回归已覆盖它们。证据见
[原锁运行观测切片](evidence/global-optimization-observation-install-20261014.json)。
这是资源忙碌观测，不是全阶段真实模型/吞吐验收；整体目标仍 active。

### 7.27 S5 末答卷有界接续与集成夹具误判边界（2026-10-14）

原 Save 在领域 Writer 接收后调用有界 Controller；异常可保留原 pending
followup，WorkComplete 不承诺同 Hook 立即进入 S6。实际 Builder 夹具原先
仅断言马上转换；补充末回执/等待/原 milestone 诊断，并仅当明确声明
system_followup 时调用原 Main stage-answer recover。恢复后仍须原 TR-003
提交与 building；缺输入、原门拒绝、未声明延迟和恢复仍失败继续报错。
生产行为、预算与权限未变，不注入 Runtime 或重复专业答卷。

受控末 reviewer 取消验证：两独立责任已被原 Writer 接收，末次控制中断后
Main 通过原门恢复 TR-003，保持两份 submission 与证据不增加，PASS 10.908s。
诊断版本实际 Main 链 race 三轮 PASS 371.437s；含原恢复夹具与受控末提交的
三轮 race PASS 517.798s。观察切片完整 CLI 已终态 PASS 1211.371s，编译时
尚未包含这些后加测试，证据严格分列。

最初并发负载失败尚未重新复现；受控机制不能证明历史失败根因，保留
调查项。详见 [S5 接续证据](evidence/global-optimization-s5-deferred-followup-20261014.json)。
全局目标 active，继续 G06 吞吐与 G08/具名 adapter，不缩减验收范围。

### 7.28 G08 工厂案例 runner 与机制/宿主证据分离（2026-10-14）

先固定 G08 §2.1，新增工厂 tools/optimization-eval（不进 release include）。
十二类案例记录公开输入/来源、原 fixture/oracle、角色、版本、变量、资源与
合法结束条件。实际 dirty 源码按白名单与逐文件模式/SHA 冻结到独立真实
Git checkout；源项目不改，执行后校验冻结输入及新来源变化。

原 Go -json/-count=1 固定 argv 执行，保存原事件/stderr/摘要；必须指定
Test/subtest 真正 run/pass，不以包成功/skip/无测试计成功。执行/日志有界，
超时终止本任务进程树。每次 run 唯一目录，失败现场保留；公开视图无
oracle/fixture 字段，缺 usage/Read/Skill trace 明确 unavailable。

本次冻结1120个工厂文件，十二机制目标全部实际通过，累计进程时长
264.859s（不是产品吞吐）；runner unit race PASS 1.240s，vet/diff 通过。
real_host 的十二案例实际记录 not_run 并返回非成功，没有执行机制替身；
真实发行/oracle 隔离、十二实际业务工程 fixture、模型轨迹/A-B/holdout 仍待
实施。两次源快照不同，不比较它们为性能收益。完整证据见
[评估 runner 切片](evidence/global-optimization-eval-runner-20261014.json)。

过程中根盘 ENOSPACE 导致设计写中断；先保全本任务23个归档到内存文件
系统并验证 SHA、保持原路径，再恢复 G08 完整原文及新契约。未删除共享
缓存或其他用户文件；临时源树/日志保留且可能到期。全局目标仍 active。

### 7.29 G08 首个生产保存观测与验收夹具清理（2026-10-14）

先固定 G08 §4.1，再把已识别 S5 答卷的原 Save 调用接入原 observer spool。
两个新增 family 分别记录保存调用计数与耗时；闭合 outcome 区分中断、
读取失败、草稿/无效答卷、完整答卷拒收和三个有效接收结论。原 Writer
确认持久接收后才记 received，有效 Finding/业务未决与阶段 PASS 分开。
一次发布同时承载计数/耗时；原 CAS 内部重试不另计，统计失败不改回执。
health/doctor 都有消费者，family budget 为 18，并覆盖省略字段的静态检查。

计数包含重放，不代表唯一交付或首交成功率；耗时为本次 Save 调用观测，
不推算跨进程等待、真实 usage 或性能改善。原 compaction 的并发和删除
中断验证两个 family 不丢失/不重复。真实 Hook 草稿、授权拒收、接收、
重放、冲突与指标目录故障验证，原业务事实和权限仍由原领域负责。

指标 race PASS 1.468s；原 Hook/三结论普通回归 PASS 28.967s。
领域 race 和非 CLI 全包通过，系统包 PASS 150.133s；宿主不支持的两个
成功隔离用例仍精确排除。CLI race 首次并发运行失败，保留原记录；
九目标独立复跑 PASS 510.845s，包含实际并行保存、原中断恢复和原 Main
检查锁。历史失败原因未确认，不能从独立通过推断已解决所有负载问题。

重复保存夹具先由原 Main 恢复明确 pending 接续，再保持处理后重放不增
revision 的原断言；核对原接收 key/证据/内容与证据数量不变。补充真实
独立 Worker 的受控延迟：首次取消过早被原授权拒收，改为原完整语义
验证通过后的取消边界，普通/延迟两分支 race PASS 73.779s。
这些后加测试与之前编译的九目标回归分列，不称完整 CLI 回归。

另确认并清理了本任务旧原锁测试残留进程。夹具改为临时目录移除前释放
并等原集成完成，目录消失时 wrapper 报错退出；正常忙锁/提前返回两个
真实 Main 链 race 分别 PASS 104.48s/99.17s，随后未观察到对应残留进程。
生产进程预算、原 required 检查、权限和接续逻辑未改变。

五平台交叉构建、206 个资产摘要与四种本机新安装通过；原生指标消费者
验收输入明确为合成 telemetry，真实生产者/Hook 证据来自上述原 CLI 链，
不冒充模型作业。包 SHA256 为
ba614afd5996b4d94da83be647572dc64f38e0b332f9d097b5caed19a540631d。
详见 [S5 保存观测证据](evidence/global-optimization-submission-metrics-20261014.json)。

全局目标 active。G08 唯一/首交统计、跨进程时间线、通用 stall、真实宿主
与模型对照，以及其他具名阶段 adapter/原门完整验收仍待继续实施。

### 7.30 G08 shadow 进展采样及实施节奏纠偏（2026-10-14）

固定 G08 §5.1 后，新增有界 work_progress 缓存与 typed notice Schema。
真实 owner/session 解析后、普通 Bash 早返回前采样；Main 加 Read/Glob/Grep，
Worker 原 matcher 保持。默认 shadow，只记录诊断；明确 hook
--progress-mode notify 可验提示 wire，工具 payload 不能开启提醒。
采样只读原 Runtime/packet/Assignment/milestone 与有界专业资料，不跑
Controller、不写 Runtime/journal、不更改完成或权限。

新必要资料的实际返回字节/范围、专业答卷与责任代码变化重置；纯
revision/时间/message_id 变化不重置。原运行锁、人闸、外部责任等待、
已接收交付、未知读取范围、缓存损坏/锁忙/饱和均不强判空转。
缓存最多64KiB/64范围，读取资料4MiB/16来源，原兼容 Runtime 观察锁
预算25ms；pending 原 pair 不修复，普通 Writer/Snapshot 行为保持。

必要定向验证已执行：机制/原 Main 与 S5 Worker wire race、影子默认和
payload 不能提升提醒、实际 notice Schema、原 pending/有界 Runtime
观察均通过。记录各次编译/范围，不声称完整 CLI、真实模型或所有阶段
专业 gap adapter 已验收。证据见
[G08 进展采样](evidence/global-optimization-work-progress-20261014.json)。

按用户侧聊纠偏，核心交付清单已集中到 §7 开头；不再把持续加固/重复
打包当核心 adapter 进展。当前不重跑完整平台矩阵；下一项恢复 S2 设计
资料登记链，再逐项推进 S3/S4/S6/S7/S8/S9/S10。完整目标仍 active。

### 7.31 S2 设计 adapter 的题目、绑定与领域投影（2026-10-14）

在 L3-S2 固定 Main planning.design 接收契约后，新增 DesignPacket 与独立
PlanningBinding：来自实际 runtime/generation/stage_entry/locked REQ，
不伪造 Assignment 或 Architect agent。内嵌 S2 模板只让规划者填写
原架构产物路径与明确自审 pass；不根据文件存在生成专业结论。
答卷采用独立闭合 Schema，拒绝重复字段、题干变化、不完整和逃逸路径。

原 planning_design 信封由冻结 REQ 与实际架构引用投影，producer 为
Main/Orchestrator；它不是 gate PASS 或转换证明。原 fileview 的 authority
root、冻结 REQ 视图/checkout 摘要与实际 planning.design 游标检查已实现。
投影 race PASS 1.132s，权威/资料漂移/错误阶段检查 race PASS 1.099s。
首次 Runtime 层误引用 workspace 形成调用环，已改复用原 fileview 权威
接口并保留编译失败记录；没有放宽根绑定或虚构主体。

此阶段只完成模板/投影/权威校验，尚未发行到原 Runtime、未调用领域
Writer 接收或 PTR-PLAN-01、未接线 Hook/回执/恢复。核心完成计数仍
S5 机制链 1/9，S2/S6 为局部能力；完整模型链仍 0/9。
下一改动明确扩展原 work_packets/submissions 的具名 subtype 分派，
并同步原完整 pair checkpoint，而非建立第二生命周期或省略恢复。
详见 [S2 设计投影基础证据](evidence/global-optimization-s2-foundation-20261014.json)。

### 7.32 S2 Main 保存、原门与原完整 pair 恢复（2026-10-14）

在 §7.31 基础上，原 v1.2 work_packets/submissions 增加显式 S2 子类型；
1.1 和 S5 的闭合子类型保持。发行调用原 Writer，保存调用原 RecordEvidence，
同事务记录 planning_design、接收摘要与 pending followup。信封精确匹配
冻结 REQ 和实际 locked ARCHITECTURE 的路径、版本与字节摘要；机器不代答 pass。
unknown UI impact 在原 Writer 接收校验中拒绝，需求澄清仍走原人闸。

Main 的实际 Write/Edit/MultiEdit Hook 与终端 consume 共用消费者，原 Controller
执行 PTR-PLAN-01，原动作登记 design。回执依赖原接收与原证据摘要；同内容
重放不增加 revision，异内容拒绝；接收后的中断只恢复原 followup。Main Stop
仍依据项目原状态，设计责任回执不授予项目结束资格。

原完整 pair checkpoint 按子类型验证发行、接收与 journal 转换关系，再将
临时 S5 投影交原 S5 校验器；持久化完整 state/journal 未裁剪。实际 CLI fixture
验证损坏后走原 recover plan/apply，恢复保留原全部事实，重放不重复转换。
伪造 owner/content digest/发行记录/转换记录均拒绝。

定向 race 覆盖 S2 issue→Main 保存→原接收→原门/登记→回执/重放、拒收、
中断 followup、原 pair 恢复，以及原 S5 三路兼容；证据与各自执行范围见
[S2 原接收链证据](evidence/global-optimization-s2-reception-20261014.json)。
完整平台矩阵与真实模型未执行，核心机制计数为 2/9，模型全链仍 0/9。
下一项为 S3：先固定设计输入闭包、合同集与原门的接收契约，再实现同路径 adapter。

### 7.33 S3 契约 adapter 的设计与投影基础（2026-10-14）

L3-S3 固定 Main planning.contracts 的接收契约：实际 registered design 和
REQ、条件来源闭包与目录清单进入冻结输入；专业作者显式交付 CONTRACTS
索引、完整 FE/BE/SYNC 合同集、每端适用性及理由。原 contracts check、
scenario bridge、PTR-PLAN-02/register_locked_contracts 与后续独立审查保留。
原 checker 的 draft/REQ 范围/条款 matcher 缝隙按 L3-S3 §5.2 披露。

新增 ContractPacket/ContractInputs 与独立闭合答卷 Schema；排序和复制集合
形成稳定身份，调用方修改原数组不会改包；草稿不生成 pass，缺判断、缺
理由、适用性与合同集矛盾、重复/越界路径、改题干、额外字段均拒绝。
Runtime 从原 planning.contracts 和原 documents[].generation/locked design
投影真实 Main binding，并检查冻结 REQ/design/条件来源的原视图与实际字节。

目前只关闭题目/绑定/已提供输入的校验基础。源闭包完整解析、原 Writer
发行/领域信封、同事务接收、Hook/回执/原门及原 pair 恢复仍未接线；不能
把提供者给出的来源集合当完整性证明。完整核心机制计数仍 2/9，模型全链
仍 0/9。定向范围与首次 generation 字段错误保留在
[S3 投影基础证据](evidence/global-optimization-s3-foundation-20261014.json)。

### 7.34 S3 原接收链与条件来源（2026-10-14）

在 §7.33 后接通原 Writer 的具名 S3 发行、planning_contract 接收和 pending
followup。条件来源按原 scenario bridge 模块集合、相关 accepted ADR、本地
模型引用及 sharedmodel.Check 实际 schema/样例闭包解析；合同输出自身不冻结。
源和目录成员复核失败、不明确的 shared model policy、交付集遗漏/draft、
索引矛盾均拒绝。原 contracts check 在应用层及受信 Writer 回调重复执行，
避免 Runtime→semantic 依赖环，PTR-PLAN-02 原 guards/actions 保留。

信封保留 canonical contract_answer，以原证据摘要重建历史判断，不信任可改
作者文件。原门通过严格 S3 已发行/接收/信封/完整来源校验，读取尚未登记的
条件来源事实；原 subjectsMatch 算法与原合同登记动作保持。Main issue/save/
receipt/recover、原转换与混合 S2/S3 pair 恢复已定向验证。证据见
[S3/S4 原接收链](evidence/global-optimization-s3-s4-reception-20261014.json)。

### 7.35 S4 Main 原 TASK 收口与完整恢复（2026-10-14）

先按 L3-S4 §11 固定设计，再实现独立 typed packet/answer。冻结当前正式合同
全集、索引条款分母、原 S3 已接收来源以及 TASK 模板/派发规则。真实 bound REQ
路径保留描述后缀，不按 ID 伪造路径；旧实例缺新 S3 原接收时不补造来源。
作者只答 conclusion/index_path/完整 task_paths，依赖和条款来自正式 TASK。
原 tasks check/waves-v1 在接收前和 Writer 受信回调执行，原 TR-002 保留。

取消 TASK 仍在完整交付集，门适配只投影其真实 cancelled 状态，不能算 complete
或删掉分母。planning_task 信封带 task_planning_answer，Main 保存/回执/重放/
pending followup 与原完整 S2/S3/S4 state/journal pair 恢复已实际 CLI/race 验证。
未答完、缺真实 prospective write scope、冻结规则漂移拒绝且不登记新接收。

首轮合并回归的 Schema resource ID/局部引用错误已修复；夹具缺正式 REQ 登记、
缺 shared-model policy/实际 Scope 声明和错误读取索引 runtime_id 的问题保留失败
日志。race 中 6 秒保存 followup 可合法 pending，夹具由原 Main recovery 收口，
生产预算保持。定向修复后结果与边界见同一证据；未重跑完整平台/安装矩阵。
当前核心机制 4/9，真实模型全链仍 0/9，完整全局目标继续 active。
下一项按已固定 L3-S6 契约接唯一 Assignment Record，再接计划/结果原消费者。

### 7.36 S6 唯一 Assignment Record 首切片（2026-10-14）

按 L3-S6 已固定的显式 `assignment_record_contract: "s6-v1.2"` 契约，
当前已接入原 `register-workgroup` Writer：同事务登记 canonical
`entities.assignments[]` 与 TASK owner、Team、Agent；以真实锁定 TASK、
当前 reviewed dispatch row、实际安装角色/Skills、CheckProfile sources
构造并冻结记录，保留原 CAS、pending 与 staged activation 清理。

实际 Register 定向夹具已证明记录与实体同事务落盘、激活范围取
`EffectiveWritePaths`、无 marker 不启用，以及来源漂移不改 state/journal
且不留 activation。薄 PLAN 仅要求 objective/planned_paths/assertion_checks/
dependencies/risks_blockers，由本人填写；原 checkpoint 消费者校验当前
Record identity/revision/digest、冻结来源、计划路径与依赖。

本切片继续接 native/workspace context、Execution 与 activation 的记录引用；
其合并回归与证据以本节后续记录为准。新增 Record 的修订入口、专业 Result
与实际 Git/runner 事实绑定、Main Inspect/Integrate/checkpoint/receipt、
TR-006 精确 TASK 集合及 fresh S7 尚未完成。机械 scope 当前固定
manifest scope × write/output × reviewed TASK；仓库尚无可计算的角色最大路径
契约，沿用原角色及 policy 约束，不宣称已新增该能力。

原登记/薄 PLAN/activation/context/Execution 首切片已通过 8 个相关包的
合并定向 race 与 vet，原 native/workspace context CLI 兼容组 race 通过。
新增 canonical native CLI 场景已通过登记→context→作者补五字段→begin；
其中发现的原 `req workspace` 缺审计字段已修复；包含该真实绑定入口的
完整定向重跑通过，错误 REQ 分支不改原 state/journal。证据见 [S6 Assignment Record 首切片](evidence/global-optimization-s6-assignment-record-20261014.json)。

Result 的封闭专业答卷/Record 草稿/AnswerDigest 纯层已落地并通过定向 race；
原 CompleteTask 与实际 evidence/Git/runner/checkpoint 消费尚未连接，
仍不产生专业 PASS 或 integration verified。

S6 仍为 partial，核心机制仍 **4/9**；真实模型全链仍 **0/9**。
完整全局目标保持 active，剩余 S6/S7/S8/S9/S10 与宿主验收继续推进。

### 7.37 S6 薄 Result 原接收与内容幂等（2026-10-14）

显式 canonical Assignment 的四字段 Result 已接入原 `task-complete` Writer。
消费者重读当前 Record 和已接收 PLAN，在同一 mutation 保存实际证据的耐久
副本、Result、原 evidence index 与 Agent/TASK 当前引用；记录独立的正整数
Result revision 和 AnswerDigest。同 key 异答案拒绝，同答案重放返回原回执，
历史重放保留当前指针；新答案清除旧 ACK，要求 Main 重新验证绑定。

真实 CLI 定向夹具已通过原 REQ workspace 绑定、Register、native context、
作者 PLAN、AgentBegin、Result 接收与只读 task-receipt 链，覆盖错误 tuple、
Worker 假 checks 零写入、重放、新 revision、历史回执及损坏证据拒绝。
共享 Result 来源与 Main facts DTO、封闭 completion Schema 定向 race 通过。
原 Store pending 三个持久边界的重建夹具也已通过 race：只读拒绝、清理保留
Result bytes、原恢复收口 state/journal 且不重复追加，恢复后可重建原接收回执。
该夹具重建真实已接收 state/event，不声称执行过真实进程崩溃实验。
证据见 [S6 Result 接收切片](evidence/global-optimization-s6-result-acceptance-20261014.json)。

本证据中的 Main facts 为线格式夹具；真实 Main 集成、完整 required checks、
新 Result 重检及 canonical Gate/回执合并验收另行进行。专业 Closing Contract
尚未闭合，当前 `accepted_unassessed` 与 `work_complete=false` 明确表示该欠项；
接收成功或机器 verified 不产生专业 PASS。Result 草稿交付、Record 修订失效、
fresh S7、后续 adapter 与宿主验收继续推进。

### 7.38 S6 原 Main 集成事实、回执与 TR-006（2026-10-14）

canonical Main 已沿原 Inspect/Integrator/锁/checkpoint CAS 接线。真实 Git 夹具
经原 Register→AgentBegin→CompleteTask 后，提交实际 worker commit、执行
non-squash Main merge 与原全量 Required Checks，持久完整 changed paths/digest、
实际 runner receipts/log SHA、tested head、Record/Result revision/digest。
调用者自报 PASS、另传 check runner 或检查列表不能替换这份原契约。

同一真实夹具调用公开 CurrentBuilderReceipt 和原 TR-006 Gate：集成前
false/not_ready，Main verified 后 true/satisfied；原 Writer 接收 Result r2
立即使旧绑定失效，Main 重跑检查后恢复。原两父 merge 不重复创建，精确重放
不重复检查；损坏证据强制重检，实际失败保全，恢复成功后再 verified。
真实越界 Git diff、旧 Record/Result 投影拒绝。专业状态始终单独保留
accepted_unassessed/work_complete=false，不以这份机器证据产生专业 PASS。

上述跨包三场景正常及 race 通过；canonical/legacy Result 与 Builder receipt/
batch 合并定向 race、原 Main recovery/idempotency 定向 race 与相关包 vet
通过。证据见 [S6 Result Main 集成切片](evidence/global-optimization-s6-result-main-20261014.json)。
该证据为真实工具在临时 Git 项目的机制验收，未执行真实模型全链或新鲜 S7。
剩余本人专业收尾、Record 修订、Result 草稿完整交付和后续 adapter 继续推进。

S6 保持 partial；核心机制仍 **4/9**，真实模型全链仍 **0/9**，目标 active。

### 7.39 S6 显式本人收尾与 Main 等待（2026-10-14）

先固定 L3-S6/G02 契约，再由原 Register 冻结 `closing_contract` 和显式
`worker_checks: []`、post_merge mode、原 Main CheckProfiles。新 PLAN 逐条映射
冻结 DoneWhen，Result 由本人填写 closing_assessment；同一原 CompleteTask
校验绑定、登记专业答案和耐久来源。缺契约不推为空，旧 Record/digest 不迁移。

原 CLI/race 已验证本人完整→work_complete=true 等待 Main、合法受阻答卷接收
但本人未完整、未映射 PLAN/缺专业判断零写入拒绝。native/workspace 草稿保留
作者内容且拒绝绑定/嵌套字段漂移；新 contract workspace 将 Main 原检查固定
到 Main，显式 worker location 拒绝，未新增默认重复 Worker 检查。

真实 Git 原 Main 合并/全量 profiled check/Gate 定向及合并 race 通过：本人完整
且 Main verified 时两种责任分别成立；新 incomplete/risk/scope 答卷阻断原
Inspect/Integrate/ACK，即使调用者伪造 ready/cleanup 也保全旧 merge/facts、
当前 Gate 未满足。多原因重放不增加 checkpoint revision。非空 WorkerChecks
明确 unavailable，不能借用 Main receipts，实际 Main runner/merge 不运行。

Record/Result/Gate、PLAN/Schema/context/workspace、CLI 新/旧 canonical 路由、
Main 新/旧 canonical 路由定向合并 race，以及九个相关包 vet 通过；文档已同步。
证据见 [S6 本人收尾切片](evidence/global-optimization-s6-closing-20261014.json)。

下一项先固定非空本人检查的原 runner/Writer 接收与机器闭合摘要幂等，再实现；
随后落实 Assignment revision 原 Writer、审批新版本绑定、native/workspace 完整
派发与恢复场景，继续 fresh S7 和后续 S8/S9/S10 adapter。安装/平台里程碑及
真实模型/宿主验收尚待执行。本轮临时 Git 机制验证不能替代上述验收。
S6 仍 partial；核心机制 **4/9**，真实模型全链 **0/9**，完整目标保持 active。


### 7.40 S6 非空本人检查、原 Result 闭合与恢复（2026-10-14）

先固定非空集合、独立收件索引、机器闭合摘要及 alias Writer 契约，再复用原
workspace/native 隔离 runner、process capture 和 Runtime CAS。Main RequiredChecks
保持独立；工具字段不能由专业答卷提供。当前 Record/owner/profile、源快照、
bootstrap 或 native 固定 context/实际 activation 在原接收时核对。

原 CompleteTask 保存每项 WorkerCheckFact、receipt/log 原字节耐久副本与
ToolClosureDigest。已登记 key 读取原结果重放、零写；新 key 按
(AnswerDigest, ToolClosureDigest) 识别版本。同内容的新键由原 Writer 固定 alias，
不增加 Result 或改变当前指针/ACK；新工具事实可使同专业答案形成新 Result。
原恢复 checkpoint 复验每类原 journal 接收事件，不能凭文件存在恢复授权。

当前源摘要另进入机器内容身份，missing/stale 仍不产生 PASS，连续改源不会因
同 pending 状态 alias 到旧耐久来源。workspace sealed Record/完整 execution
snapshot/bootstrap 均经耐久重放校验；相对本地 runner 指纹保留脚本坐标与
字节摘要，避免随机 scratch 路径使相同输入失效。native 精确受信命令在混合
workspace registry 场景的窄权限路径已接线，未知身份/改写调用继续拒绝。

真实 native CLI 原检查→Result 已验证两次宿主实际 fail、原 receipt/log/SHA
耐久副本、精确重放、alias、改源 fresh 后的新 Result 与不可改历史。当前宿主
Bubblewrap/user namespace 能力不足；workspace prepare 在原能力门拒绝且
state/journal 原字节不变，其完整 runner→Result 未执行。纯 PASS DTO/Gate
夹具验证不等于适用宿主 PASS。未新增 fallback 或真实模型/平台成功声明。

定向 race、原 Main closing 阻断/Gate、受影响包 vet 与各入口测试记录见
[非空本人检查切片](evidence/global-optimization-s6-owned-checks-20261014.json)。
Assignment revision 原 Writer/人闸版本绑定设计继续固定后实施；native sidecar
来源绑定、workspace 全场景、适用宿主 PASS、fresh S7、后续 S8/S9/S10 adapter
及最终安装/平台/真实模型验收仍未完成。S6 partial；核心机制 **4/9**、
真实模型全链 **0/9**，完整目标 active。

### 7.41 S6 原 Assignment 修订发行与独立 Main 槽位（2026-10-14）

先固定修订契约，再接入原 Register 单一 Writer。原 workspace 锁覆盖发行与
Main merge/check/cleanup，稳定意图和完整来源摘要绑定前版 ref/SHA；新版本由
Writer 连续分配。初次 r1 与新版本保存有界的真实来源字节快照，历史 Record
保持原 digest，快照仅供审计。部分发布失败的清理意图在发布前登记。

首切片支持同 exact locked TASK 与 plan_checkpoint，fresh Team/Owner；旧
Result/ACK 当前投影失效，旧 execution 转历史。CLI 返回原发行回执，历史 r2
在 r3 后重放为非当前版本且零写，不回滚当前版本。同前版并发只有一个赢家；
缺失或错误的原 journal 发行事件不能恢复授权。人闸与 TASK 来源变更仍拒绝。

当前权威 Record 选择 Main checkpoint：r1 保持原路径，r2+ 使用独立 aN 槽位。
真实 Git 夹具完成 r1 原 Register/PLAN/Result/Main，再完成 r2 原发行、新 PLAN/
Result、新 scoped commit、双父 merge、fresh 全量 Main checks、责任回执与
TR-006；Main 前 not_ready、后 satisfied，r1 原 checkpoint 字节保全。旧 Owner
不能提交 Result/ACK；历史 canonical Result 不进入 legacy 完成集合，精确 TASK
分母保持原定义。

最新合并代码的 Main 完整链正常测试通过（18.397s）；较早的独立 Main 链 race
通过（91.003s），最终 Writer 锁与并发另经合并 assignment/CLI race 覆盖，
不将较早 race 声称为最新 Main 组合的 race。原 pending 重建恢复、来源快照/
Result 接收、Gate/context/path 定向 race 与十个受影响包 vet 记录见
[Assignment 修订切片](evidence/global-optimization-s6-assignment-revision-20261014.json)。

下一步固定高风险原生 Plan approval 的可信来源与新版本绑定，并审计跨 Record
前版 Main 合并依据和 TASK 正式来源修订。旧 verified/checks/ACK 不继承；native
sidecar、适用宿主 PASS、完整 workspace、fresh S7、S8/S9/S10 与最终安装/
平台/真实模型验收仍未完成。S6 partial；核心机制 **4/9**、真实模型全链
**0/9**，完整目标 active。

### 7.42 S6 → S7 薄 Result 空分母保护（2026-10-14）

定位到原 S7 completion 读取器仅取 changed_paths，而 canonical 薄 Result 的
真实变化面属于独立 Main facts。现先固定 L3-S7 接续契约并在原读取器拒绝将
assignment_record 标记的薄 Result 解释成空分母；null/畸形标记或作者提供
changed_paths 也不能回退 legacy。原 ReviewPlan 基线校验与 S10 外部分母
读取获得明确诊断，不生成完成事实；S10 的独立 acceptance 读取器也已接同一
保护，实际原 Gate 返回 external_baseline_unverifiable。合法 legacy 空变化面保持原行为。

定向 review race（1.182s）覆盖上述保护、目录/symlink 路径约束与原 Reviewer
独立性；acceptance/qualitygate 定向 race（1.124s/1.228s）另覆盖 S10 读取与
真实 Gate 拒绝，见 [S7/S10 分母保护记录](evidence/global-optimization-s7-canonical-baseline-guard-20261014.json)。
原 Main facts 读取能力和累计产品变化投影仍待接线；r2 最新 diff 不能遗漏 r1
已合入且仍存在的变化。该保护不构成 S7 完整链完成。核心机制 **4/9**、真实
模型全链 **0/9**，完整目标 active。

后续只读审查发现 S10 显式 all 曾在读取原证据前返回，现 rooted acceptance
入口先校验来源，再仅免除集合比较；薄 Result 或缺失原文件仍拒绝。最新
review/acceptance 定向 race（1.192s/1.128s）通过。共享 integration 施工期间
一次 qualitygate 构建未完成，该中间日志不算通过；构建修正后，实际原 Gate
拒绝链的定向 race 已另行通过（1.259s）。

### 7.43 S6 立即前版 Main 合并依据接续（2026-10-14）

先固定前版依据契约，再实现独立历史 Result 审计 API 与原 Integrator resolver。
普通 LoadAccepted 继续限定 current Owner/有效证据；历史 API 复验连续前后
Record、原 supersession 事件关联、PLAN/答卷/本人工具事实与耐久证据闭包，
拒绝重复 evidence ID。历史审计不产生当前权限，r3 后仍可审计 r1→r2。

原 Main 锁内先验证完整原 revision receipt/journal，再找当前最高 Record 的
立即前版。仅已有完整源快照、唯一被该发行事件替代的 Result、独立原 Main
checkpoint/checks/log/profile 与真实双父 merge 关系齐全时，允许从旧 side
parent 继续提交的分支进入原集成。实际 merge-base 必须等于前版 SourceHead，
前版 merge 仍在当前 Main，候选必须唯一；当前 r2 delta 按 r2 scope 校验。

新版本继续创建自己的 merge、执行全部 fresh Main checks、保存自己的 Main
facts 与责任回执；旧 checks/verified/ACK 不继承。真实 Git 正常测试（19.676s）
验证缺原发行 journal、当前范围越界、前版 Main log 篡改均不移动 Main 或运行
新检查，恢复后原 RetryPreserved 完成并通过 TR-006，r1 checkpoint 字节保全。
历史 API 定向 race（1.759s）及受影响包 vet 已通过，组合 Main race（87.200s）
覆盖旧源续接与从当前 Main 建新分支的完整链。独立审查后另补历史 execution
generation 由原 retired registry 绑定取得的检查，最终定向正常/race 再通过
（20.109s/83.880s），最终 integration vet 通过。持久记录见
[前版 Main 接续切片](evidence/global-optimization-s6-predecessor-main-20261014.json)。

legacy 前版缺 source bundle、多个旧 merge 候选、无新提交仍保留 pending。
旧 checkpoint 尚无初次 verified 时的独立原始摘要锚点；本切片继续信任原
受保护 Writer 目录，不声称抵抗整套 facts/receipt/log 连同摘要被重写。发行
journal 暂未显式列旧 Result ID/path，关联使用原 Writer 的 invalidated_by。
完整 TASK 来源修订已固定必须经 TR-007→S4/TR-002→S5 双审→TR-003 锁定的
设计，尚未落地。高风险平台 gate 与真实人类决定的可信来源已区分，版本绑定
仍待宿主接线；fresh S7 的累计产品变化和删除依据读取、后续 adapter/安装/
真实平台验收未完成。只读定位确认原 TR-003 pinned View.Commit 尚未持久化；
REQ bound_commit 和每个 execution BaseCommit 均不能代替共同起点。下一切片
已在 L3-S7 固定：原 TR-003 首次锁定事务保存同源起点，同 generation 返工
保持该起点，S7 固定最终 Main、计算真实全量差异并保存删除前像来源；历史
无来源的 generation 不补猜。核心机制 **4/9**、真实模型全链 **0/9**，完整目标 active。

### 7.44 canonical S6 共同基线 → S7 注册与 S10 完整变化面（2026-10-14）

按 L3-S7 已固定契约，原 1.2 TR-003 `register_execution_batch` 同事务保存
实际 pinned View 的 `s6-main-start-v1`，将摘要写进原 action receipt；同
generation 保留首次起点，缺原记录且已提交过 TR-003 时拒绝补猜。读取完整
rotation journal、原提交身份和 cursor，旧 Runtime 不被暗迁移。

原 Main 回执强验证提取到只读 `mainreceipt`；原 Runtime journal/发行依据提取
到 `controlproof`，Runtime 保留原 API 与 Writer。`reviewmeta/runtimemeta` 仅
共享原 DTO，解决消费者依赖环，不新增控制面。`buildsurface` 从上述原权威
重建 exact TASK、当前与连续历史 Record/Result/PLAN、完整 Main 工具闭包和
Git 合并依据，再固定首次 start..最终 Main 的全产品变化面。保留空格/换行
字面路径及真实删除提交/前像；未归属 checkpoint 的变化仍进入分母并诊断。

ReviewPlan 保存 typed `product_baseline` 与其摘要，冻结全部 TASK/变化路径；
原 Register/Revise/Submit/Revive、capture 和 CleanRound 消费复验来源及最终
HEAD。原 TR-009/015/017 Writer 同步拒绝旧 clean 标签对应的漂移来源。
TR-006 原 builder_report_record 槽允许真实 canonical completion_report，
但须经完整 buildsurface 校验且匹配原已接收 Result，不制造兼容报告。
S10 读取原已注册 ReviewPlan 的同源产品快照，显式 all 保留完整路径分母。
S9 原 TR-012 seed 使用同一快照和完整 required QA/TASK，仍需 Planner 专业
修订与 fresh S7；seed 不成为 PASS。

独立复核补强两处：原 pinned Plan 已声明 product_baseline 时，缺当前
Assignment registry 不允许降级 legacy，坏指针也拒绝；当前 TASK 与首次
TR-003 TaskSet 逐项匹配，拒绝增删或改身份/版本/摘要。本切片仅支持 TASK
来源未变的 Assignment revision；将来正式来源修订须读取原完整返工/重锁
证明后更新当前分母，仍保留首次产品起点，不能永久冻结合法后续设计变更。

真实临时 Git 组合用例已通过原 S5 Gate/TR-003 → Register/PLAN/Result →
Main Integrate → TR-006 Gate/Writer → authored Draft/原 S7 RegisterPlan。
读取已注册快照成功；缺起点、Main 日志篡改、最终 HEAD 漂移拒绝，新增
未归属产品路径仍进入新快照。该用例使用机制夹具，不声称模型专业验证。
共享 Git/控制依据、ReviewPlan/S10 消费与原 Main 回执定向 race 已通过；
S9 只验证 canonical seed 转换与原 legacy handoff 用例，完整 canonical
S6→S9→fresh S7 尚未验收。最终复核与来源/日志摘要记录于
[本切片证据](evidence/global-optimization-s7-product-baseline-20261014.json)。

旧 r1 仅有 receipt source bundle、物理 Main 根目录迁移、TASK 正式来源修订、
高风险宿主批准绑定、完整 S7 typed 保存 adapter 与 S8/S9/S10 全链、安装及
真实平台/模型验收仍有欠项。受保护 Writer 目录的整套重写抵抗边界沿用
§7.43。核心机制 **4/9**、真实模型全链 **0/9**，完整目标保持 active。

### 7.45 S7 Reviewer 停止依据与原消费者回执切片（2026-10-14）

先在 L3-S7 与 L4 stage-form §13.7 固定契约：S7 的普通 completion_report
仅通知；typed 答卷必须绑定受信派发 owner、原 Plan/Assignment/精确 Claim、
subject digest 与发行路径，仍由原 `review.SubmitResult` 校验/消费/封轮。
下一 Writer 切片用现有一次性 result_id 作稳定键，在同 CAS 存完整作者答卷
摘要、effective capture 闭包及可重建 receipt；同键同摘要零写重放，异摘要
冲突，响应丢失先恢复原 pending，不重复 Finding/CleanRound。旧 1.1 保持原形。
typed 发行/保存 Hook 与完整专业接收链尚未验收；原 Writer 的首个幂等回执
基础已接线并定向验证，见§7.46。

独立核对发现原 `ProjectResponsibility` 把任意 completion 通知或 reported
状态视作 S7 已交付。原 Stop/Idle 现从同一 Runtime Snapshot 投影派发与
接收事实，共享 `ReviewResultConsumed` 核验 current Plan/round/generation、
owner 与 dispatched TASK set、exact Assignment/Claim dispositions、consumed
result_ref、唯一 valid/live ledger、原 envelope 路径/摘要及坐标。读取复用
原 `workpacket.ReadFile` 的有界 regular/no-symlink 规则；不新增 Writer。
未消费的通用通知不能放行，已消费 Reviewer 不代等全局 CleanRound。
原 queued、真实 blocker/site-lost 与非 S7 一次性派发边界保持。独立复核继续
补强：S7 one_shot 仍须消费 Result；已知 ReviewAssignment owner 的坏上下文
不得降级 legacy；S7 manifest blocked 不能替代原 Runtime blocker 事实；
重复 evidence ID 即使同路径不同 kind 也拒绝。最终定向 race 已通过；
site-lost 同时保留原 running/cannot_clean/discovery_draining 路径，拒绝已封轮状态。
Controller 完成提示也消费同一 proof，避免与实际停止控制矛盾。

Hook 责任投影/原 StopIdle 与 CLI 提示定向 race 通过；CLI 原通知消费回归
及新增实际 loader 的 S7 reported/generic completion 拒绝用例通过。正向
Hook 回执夹具为合规 Runtime/evidence 投影，尚未由真正 `SubmitResult` 生成；
完整 professional answer→原 Writer→Hook 停止组合仍待下一切片，不能以此
提升核心完整链计数或真实模型验收。日志与源摘要见
[本切片证据](evidence/global-optimization-s7-consumed-stop-idle-20261014.json)。
核心机制 **4/9**、真实模型完整链 **0/9**，完整目标保持 active。

### 7.46 S7 原 SubmitResult 完整答卷与稳定回执（2026-10-14）

按§7.45/L4 §13.7 契约，只有原 1.2 packet 模式在现有消费 CAS 中写
`s7-review-result-receipt-v1`；使用一次性 result_id 作稳定键，绑定原 Plan
ref/SHA/revision、generation/round、Assignment 完整摘要、owner/subject、
作者答卷与 capture merge 后有效答卷摘要、原 Result ref/SHA 和实际提交事件。
旧 1.1/非 packet 答卷保持原 state/envelope 形状，不暗迁移。

原 envelope 保留既有领域投影，增加完整 submitted/effective answer。
规范摘要拒绝重复键并规范嵌套对象键顺序；机器 tuple 单独校验。原 Writer
先恢复 pending，再在读取 live CaptureDir、过期 ExpectedRevision 或已封轮
状态检查前核对原已接收回执。完全相同返回原持久事实，不重复 Runtime、
journal、指标、Finding、CleanRound/ObservationBatch；异摘要/机器 tuple、
新 ID 覆盖已消费 Assignment、无完整旧回执或损坏来源均拒绝。新提交仍走
原完整 Claim/capture/evidence/Finding/site-lost/独立性校验，CAS 内再核 Plan/owner。

受控 Plan 修订比较完整 Assignment 内容与逐 Claim owner；metadata 变化、
删除旧 Assignment 后移动 Claim 到新 ID 都使旧答案失效、要求 fresh answer。
真正未变 Assignment 保留原回执，经原 immutable Plan 与当前对应关系可重放。
Result 输入为有界 20 MiB regular file，仍支持 Main 外的合法 Worker 输入路径。
产物写入执行 file/目录 sync；两格式的清理统一走原 Runtime 锁内
RemoveUnreferencedArtifact，绑定 revision/ref/SHA 并保留 pending 所引用的产物。

独立复核修复了 raw JSON 嵌套键顺序、incoming generation、effective tuple、
证据失效字段、未限长读取、未知 CAS 后误删 pending 产物、Claim owner 转移
及 receipt Go struct 的恢复序列化顺序问题。回执先 round-trip 为纯 JSON map，
使原 state 哈希与 marker 重读一致，不改变 Runtime 的原哈希算法。

真实原 SubmitResult 用例验证普通 PASS→原 CleanRound、Finding→原 batch、
同键并发一条提交与精确 retry、响应丢失/缺 CaptureDir 重放、完整源篡改拒绝、
修订失效与未变项保留。由真实已接受 state/event 构造 sealed commit marker，
恢复提交前 pair 后通过原 Writer 完成两条恢复链，不重复 Finding/CleanRound。
最终 receipt/revision/legacy 定向 race **156.852s** 通过，新增 owner/input/cleanup
用例另列最终 race；不将较早测试当作后补覆盖。证据见
[S7 Writer 回执切片](evidence/global-optimization-s7-review-result-receipt-20261014.json)。

TR-006 Writer 另复验 canonical 全批 Main 来源，不因调用方选兼容 aggregate
报告 kind 而跳过；当前已登记 completion 的 canonical marker 或当前 pinned
Plan 声明也不能因 registry 缺失降级，历史 generation 的 Plan 不选择新代模式。
这仍是原消费者与来源机制，typed scaffold 发行、可信 owner 保存 Hook、完整
professional Reviewer→回执→停止/门链、后续 S8/S9/S10 adapter 和安装/真实宿主
验收仍待完成。核心机制 **4/9**、真实模型全链 **0/9**，完整目标 active。

### 7.47 S7 typed 发卷、可信保存与原 Gate 机制链（2026-10-14）

先固定 L4 stage-form §13.7.1，再在原 Runtime `work_packets` 增加封闭
`s7-review-result-v1` subtype。Main 从原 pinned Plan、实际派发 Assignment、
精确 Claim/owner/subject 发行；queued 不发卷，重复发卷保留作者字节。
完整 journal 与 checkpoint 复验发行及 immutable Plan 来源，不新增接收表或 Writer。

专业卷面只让 Reviewer 填判断与证据，verdict 草稿为 null。原成功
Write/Edit/MultiEdit 的精确目标经 Hook actor、原 PLAN/working、登记派发与
实际 source root 校验，再调用原 `SubmitResult`。原 Claim/Finding/capture/独立性
校验及 CleanRound/ObservationBatch 消费保持；packet、答卷与 execution_root
在同一原接受事务绑定回执，Schema 要求新 tuple 三字段全部出现或全部缺省。

本轮审计修复三处权限/重放问题：终端 `stage-answer consume --agent` 不能
授予代交身份；声明独立 execution 后，E2E artifact 摘要须读取实际 Worker
workspace，封轮、原 CleanRound Gate 与直接 TR-009 同源复验；已消费保存
不能跳过当前作者文件和 source 授权，修改后的答案仍返回 conflict，缺文件
使用独立只读 receipt 查询。原 pending-safe 清理继续复用原锁内引用/摘要检查。

发行提前检查累计状态预算；原 Runtime Writer 在 pending 前按实际最终
序列化检查 checkpoint state 16 MiB / journal 64 MiB 上限，拒绝不能恢复的
发布。原 journal schema compiler 在一次完整读取内复用，每条事件仍验证
Schema、identity、sequence、runtime/cursor；不缓存跨次读取的事实。
角色/Skill、AGENTS 安装模板与使用文档已同步此次流程和身份边界。

真实临时 Git 定向链已通过原 S5→TR-003/S6 Main Integrate→TR-006、
原 S7 Register/Reviewer Register/PLAN、发卷、公开 CLI PostToolUse 保存、
原责任回执、修改答案冲突且 state/journal 不变、恢复原字节零写重放、
CleanRound 及原 TR-009 Gate/Apply；最终 normal **141.027s**。答案/前序
evidence 为机制夹具，不等同模型专业验证。第一次最终 race **292.176s 失败**，
首次保存返回 `context deadline exceeded`，失败日志保留。按 L4 先固定契约，
只跨 Validator 复用编译后的 immutable embedded Schema，所有输入仍逐次
校验，非内嵌路径保持原行为，不缓存事实或源文件。定向并发 normal/race
通过；原始日志微基准从约 2.420 ms/次降为 0.141 ms/次。保留原 6 秒 Hook
deadline 的最终同链 race **276.138s 通过**；不推广为真实宿主性能保证。
另以每次 accepted save、异答 conflict 与 exact replay 都启动独立 test-binary
子进程的冷 CLI 变体通过 race **304.354s**，仍保留原 6 秒 deadline。该验证
覆盖进程内编译缓存无法跨 Hook 复用的实际边界，仍非真实 Claude 平台验收。
最新来源与日志见
[发行与保存链证据](evidence/global-optimization-s7-issued-result-20261014.json)。

独立 Worker artifact/source/session 的定向验证已通过，但完整实际 Worker
capture 组合、真实平台 agent_id/session 与真实模型链未执行。TASK 正式
来源修订、物理 Main 根迁移、完整 canonical S9→fresh S7、后续 adapter、
安装及真实宿主覆盖仍待完成；受保护控制文件整体重写边界不变。
核心完整链 **4/9**、真实模型全链 **0/9**，完整目标 active。

下一切片 S8 已按 L4 stage-form §13.8 固定 HypothesisResult 设计并完成原
InvestigationCase Writer 机制链验证；具体范围与限制见 §7.48。Case receipt 保存
后继 ref/revision 与原事件，不嵌入自引用 SHA；读时由原 pointer/revision_history
重建。S9 PlanReport adapter 已开始，尚不由 S8 用例背书。

### 7.48 S8 HypothesisResult 原 Case Writer 与恢复链（2026-10-14）

按 L4 stage-form §13.8 固定的 typed contract，在原 Runtime `work_packets` /
`PACKET-ISSUE` 中发行 `s8-hypothesis-result-v1`。原 Register 保存实际
generation 和 Investigator Team 的 manifest SHA；卷从注册的 manifest/TASK、
Assignment、PLAN 与当前 Case/route 闭包投影。专业 HypothesisResult 仍由原
Case revision CAS 写入 `hypothesis_results[]`，Receipt 在该行，不新增 Result
表、queue 或 Writer。

真实 generation-2 机制夹具覆盖原 Register → AgentBegin/PLAN → Issue →
公开 CLI PostToolUse → 原 Case/Runtime receipt → Stop/Idle。它检查 Register
后的 manifest 修改被拒、未经 human approval gate 的 `plan_approval_required`
答卷被拒、已接收 owner 可结束本人责任而未答 owner 仍被阻止；先接收一个
并行 Hypothesis 后仍可恢复/接收兄弟答案，兄弟 Case revision 不丢原 packet。
保存对当前文件和 source authorization 做复验；原 route consumer 仅更新 pointer
而 Case 仍保留旧 route 的情形被拒。再以新 Case revision 清除 pointer 的消费标记、
恢复旧可见 tuple，旧 unanswered packet 仍因原 Case 连续修订闭包及完整 journal
中的原 `S8-ROUTE-CONSUME` 共同复验而拒绝，防止通过改后再改回复活。原已接受
receipt 仍可只读查询。

在单进程并发 direct submit 与 sealed pending recovery 定向用例中，两个同卷答案
只提交一个原 Case revision/event；精确重试零写。恢复用例从一次真实接受的
before/after state 与 journal event 构造原 sealed marker，恢复提交前 pair 后由原
Writer 恢复同一提交；cleanup 在 marker 未决时保留 marker 和 Case 文件。删除作者
答案文件后，原 receipt 仍可只读查询，保存重放则拒绝缺失文件。Combined gen2
链加 recovery 的 normal **10.948s**、race **90.704s** 通过；Runtime source/
continuity race **15.922s**、Adapter normal/race 和 packet/source targeted 结果
见 [S8 证据](evidence/global-optimization-s8-issued-hypothesis-result-20261014.json)。

这些是原机制的真实注册、文件字节、Git/Runtime、Writer 与恢复夹具，不是实际
Claude Investigator 启动或模型专业判断。Sealed recovery 用真实接受产物构造 marker
并恢复 pair，不是对进程在 fsync 边界执行硬杀。旧 Case 整体替换且原 pointer/
revision history 不再保留时，历史 receipt 查询 fail-closed；不扫描目录猜测。
新增 intake Case 文件会同步父目录；Unix 使用目录 sync，Windows 仅验证目录且
不承诺断电后的目录项持久性。

完整安装与独立 Investigator prepare/launch、supported-host 验收、真实模型或
产品侧执行链仍未完成。Global 核心完整链保持 **4/9**，真实模型完整链保持
**0/9**。S9 PlanReport adapter 已开始后，其共用 Runtime/workpacket/schema 文件
与上述 S8 定向证据分别评估；该 S8 测试选择不作为 S9 验收。7.47 的 S7 验收内容
与证据文件未改；此处只更新原先预告的下一切片状态并追加 S8 记录。

### 7.49 S9 执行前 PlanReport 原接收与恢复（2026-10-14）

先按 L4 stage-form §13.9 固定契约，再沿原 approved Case/Contract、
OpenRepairSession、RepairPlan 与实际 repair dispatch，在既有 work_packets /
PACKET-ISSUE 发行 `s9-plan-report-v1`。原 Register 在 packet mode 的 S9
Builder 原 Team 行保存实际 manifest SHA。中性题干绑定完整源与本人原
RepairAssignment、manifest alias、Team/TASK/owner；专业答案仅 plan、
red_checks、proposed_paths，不提供审批、运行 SHA、身份或控制状态。

成功保存沿受信平台 actor、原 PLAN/working 与原来源校验，调用原
CreatePlanReport / SubmitRepairPlanReportToRuntime / PTR-BUG-10。原 immutable
PlanReport 内保存可选完整规范答案 receipt，无所在文件自引用 SHA，无第二
接收表或 Writer。该回执 `work_complete=false`；它不释放产品写入，也不让
Builder 提前完成。Main 仍经原 BeginRepairExecution 收齐全部报告，Apply 内
重新验证同一 Session/Contract/Plan、generation/round、报告集合及逐 Assignment
coverage/receipt 后才放行。兄弟报告按原 CAS 当前 refs 合并，不能丢行。

实际组合验收发现并修复：Runtime TASK 行只有 id/path/SHA，发行须通过原
Agent.task_ids 找记录后读实际 TASK JSON 的 Session/Plan/Assignment tuple；
RepairPlan 的行不保存动态 owner，Hook 必须消费原 assignment_owners，再以
原 packet 的 manifest alias、Team/TASK/owner 与 journal 来源共同证明身份。
现有 registered execution 可以 canonical 或 manifest alias 为 key，两者均须
按本人 Runtime/generation/owner 识别；不支持的 pre-execution Worker 不回退
Main。已接受后撤回 fresh answer write grant，保存重放仍读当前作者字节和
源授权；异答冲突，精确重放先于 caller stale revision/当前 phase，缺作者文件
只能用只读 receipt 查询。

恢复沿原 Writer sealed pending。具名 typed issuer/Submit 可先恢复再验原事实；
公开 stage-answer issue 也由原具备验证器的 Writer 恢复正式 snapshot 后路由，
不信 raw state hint；status/receipt 与 Hook lifecycle 查询仍只读。原案例从实际
第二次 PACKET-ISSUE 的 before/after 构 marker（retain_last_transition=true），
公开发行重试恢复 exact pair，保留两个原 packet 和作者字节、无重复事件。
同卷并发只落一个 PTR-BUG-10，报告 pending 恢复后原回执不变；构造 marker
是机制恢复边界测试，不是 fsync 边界硬杀或断电实验。

原 S9 immutable 文件函数统一使用安全 pinned parent、exclusive 创建、文件
sync/close 与父目录同步，读取限定 20 MiB regular/no-symlink。报告同时保存
专业投影与 submitted_answer，故在最终序列化后、任何发布前另验 20 MiB；
11 MiB 单体答案双投影超预算的 Schema 合法用例被提前拒绝，不能接收后
才发现回执不可读。Windows 继续受既有目录同步能力限制。回执读回时规范
内嵌 JSON，避免原 indented serializer 与 compact 答案的字节表示差异被误判。

共享 registry 审计另修原 S5 RecordWorkPacket 的槽比较：只解释 S5 subtype
的 stage_entry，历史 S2/S3/S4/S7/S8/S9 绑定不作为 S5 结构读取；发行明确使用
packet 的 generation。槽级共存及原实际 S5 发行回归另记日志，不冒充完整
S9→S5 重入模型链。

原实际 normal 组合 **7.286s**；四 target 合并 race 中公开 Hook 链 **69.46s**、
同卷并发 **10.74s**、报告 sealed 恢复 **11.11s**、公开发行 sealed 恢复
**13.29s**，investigation 包共 **105.763s**；alias 身份单测所在 repair 包
**1.129s**。全程未放宽原 6 秒 Hook deadline。Runtime 来源、完整答卷与回执、
Schema、原批次/回环、预算、安全文件与 adapter 的定向结果、失败及中断记录见
[本切片证据](evidence/global-optimization-s9-issued-plan-report-20261014.json)。
stale-staging 的兄弟提交后安全重建单项 normal **2.972s**、race **28.605s**
通过；最终 Runtime source/receipt/S5-slot normal **1.829s**、race **16.720s**
通过。五目标临时发行包构建、manifest verify、全新目录 --packet 安装及安装后的
release-graph/doctor/validate --all 均通过，见
[冻结包安装证据](evidence/global-optimization-s9-disposable-install-20261014.json)。
初次跨平台 helper 编译失败通过 Linux/非 Linux 文件拆分修复；再次打包发现的
发行文档链接指向未打包 blueprint，通过文档引用修复关闭，失败原日志保留。
该 -03 包与主证据 source snapshot 在下一 RepairResult 生产编辑前冻结；后续
worktree 变化不继承本包验收。跨编译不构成 macOS/Windows 运行证明。

PlanReport 仅完成 S9 执行前一段。完整 RepairResult typed candidate→Main
实际集成/check receipts→原 Result、独立 targeted、Impact/Handoff 与 fresh S7
组合仍待实施；Main checkout 的 S9 native registered checks 路径尚无接线，
check_location=main 只是既有 Worker worktree 在合并后由 Main 执行检查，不能
把 S6 native-owned worker checks 当作 S9 能力。实际模型/平台执行、完整宿主矩阵
与其余全局卡继续保持待完成。核心完整链 **4/9**、真实模型完整链 **0/9**，
完整目标保持 active。

<a id="s9-repair-result-progress"></a>

### 7.50 S9 RepairResult 原链与定向恢复通过（2026-10-14，新包待验收）

在 [L4 stage-form §13.10](l4-mechanisms/L4-stage-form.md#s9-repair-result-adapter)
固定原 approved source/已接收 PlanReport/Begin/实际 registered execution 后发行
专业结果卷；答案形成原 immutable WorkspaceCandidate，原 DeliveryRef/SHA 与
workspace_updated 事件证明接收。最终 HEAD/diff/blob SHA 在保存接收时机器
采集，不要求先写旧自由 proposal，也不提前冻结尚未完成的最终 HEAD。
原 ReworkRef/file SHA/event 是 attempt lineage，不能仅靠未递增的 execution
代数识别返工。题干、候选/最终 receipt 与原事件各自不含所在文件 self-SHA。

本人候选及本人 required checks 齐全时遵守 G02，work_complete=true、Main
consumer_pending；Main-owned checks 不冒充 Worker 欠项，额外明确冻结的
WorkerChecks 仍须各自原 OwnedCheckReceipts。Main-only check 拓扑可在 typed
candidate 保留明确检查 pending；最后原 Result pass 仍要真实 Main checkpoint、
current TestedHead、exact frozen profiles/check receipts/log SHA 及原专业答案。
Main 锁再进入原 Runtime CAS，在原 Submit Apply 重复验证，不新增 Writer。
checkpoint 的 ACK/cleanup 合法变化与其不可变 verified facts 分开，Result-before-
ACK 中断从原索引恢复。另一 task-integrate/Controller 入口也必须遵守同一原
candidate 与 Result 接收，不能只修保存/Stop 而保留集成绕过路径。

第一接线仍为 separate registered Worker worktree→Main，未把 Main-owned
checks 说成 S9 native Main-checkout runner。原 candidate 当前只支持 pass/no
scope deviation；typed 非 pass 原接收及后续 Targeted/Impact/Handoff/fresh S7
仍是必做剩余项，不用本次设计固定或 PlanReport 验收冒充已完成。当前已审查
原 API 并固定本契约。中性九字段专业卷面/Schema 的四项 normal 通过（0.320s），
Runtime canonical roundtrip/不可变投影四项 normal 通过（0.033s），报告来源的
保守 profile 格式和共享 S5 slot normal 通过（0.025s）；分别见
[DTO](evidence/logs/s9-repair-result-dto-targeted-normal-r2-20261014.log)、
[Runtime helper](evidence/logs/s9-repair-result-runtime-helper-normal-r2-20261014.log)、
[profile/slot](evidence/logs/s9-repair-result-profiles-and-shared-slot-normal-20261014.log)。
之后修复卷面将真实 S9 omitempty 字段误当 S6 必填的接线缺口；真实无 Record/
ClosingContract 的 S9 Execution 不需补造该元数据。合并 DTO/source normal 通过
（workpacket 0.337s、Runtime 1.038s），见
[合并来源契约](evidence/logs/s9-repair-result-dto-and-runtime-source-normal-r2-20261014.log)。
原候选事件/refSHA/当前 delivery 指针的一致性、actor/代数/key 负向的机械回执
夹具 normal 1.595s 通过，见
[候选事实读取契约](evidence/logs/s9-repair-result-runtime-candidate-proof-normal-r2-20261014.log)。
这些是组装原 source/journal 的读取器夹具，不是原审批/dispatch/候选 Writer 的
实际调用，也不证明真实 Main merge/check 或最终 Result。
Runtime 原来源/回执与 checkpoint、领域 private capability/CAS、prepare 原报告
profile 冻结及公开 Hook/CLI/Stop 分别接线中。领域包已编译；公开 receipt 查询的
pending 回归通过，原 state/journal/marker 字节不变，见
[只读查询与 adapter 编译](evidence/logs/s9-repair-result-adapter-readonly-20261014.log)。
候选接收已补原 reported PLAN 与 completion envelope 的索引/精确 SHA/原事件；
provenance 冻结原报告 ref/SHA/event，历史读取不依赖当前 Agent 标签。机械来源
夹具含当前报告的错状态/索引/SHA/代数/额外事件引用/重复 TASK 负例，以及历史
candidate 在 Agent 后续推进后的冻结报告读取，normal 2.195s 通过，见
[原完成报告与候选事实](evidence/logs/s9-repair-result-reported-candidate-proof-normal-r2-20261014.log)。
原 legacy completion mutation 显式保留实际 baseline generation，不能在 gen2
事件中默认为 gen1。首轮新增 helper 错把生命周期 Mutation.Event 当作 raw journal
event，已按原 Writer 的 transition_committed / transition_id 修正，失败日志保留。
实际 gen2 separate worktree 原链已通过：原 approved source、dispatch/PLAN、
双 Assignment PlanReport 与全量 Begin、实际 prepare/launch、Worker 原 commit/
completion_report、Main 发卷、公开 PostToolUse 接收 candidate、本人 Stop 放行，
再由原 Main merge/ProfileCheckRunner 实际检查、typed Result、ACK/cleanup，最后
公开只读 receipt 仍显示本人完成。专业答案是机械夹具，launch 使用测试替身；
没有运行真实 Claude 模型。第一轮通过 normal r17 为 54.05s，扩展无 capability
原 candidate/Result 提交拒绝与精确 PlanReport bytes/SHA 篡改拒绝的 r18 为
46.89s，分别见
[原链 r17](evidence/logs/s9-repair-result-original-hook-chain-normal-r17-20261014.log)、
[扩展原链 r18](evidence/logs/s9-repair-result-original-hook-chain-normal-r18-20261014.log)。
这两轮候选 Hook 实际响应为 **8.584s / 5.880s**；原 6 秒 context 配置未改，
响应包含边界外的处理，不能以配置值或单轮通过承诺响应始终低于 6 秒。

原链暴露并已修正安装态默认 Agent 的 catalog authority、未 prepare 合法兄弟的
批量发卷选择、原 raw journal transition/event 坐标及 Main 集成 canonical/
manifest alias 绑定。别名仅能来自同一冻结 registered execution 的精确 ID 与
定义，不能做任意后缀匹配；candidate 还须对应同一 ResultID。原 Main 锁内本人
责任验证传递相同 context 保持重入；最终 Main proof 先调用原
integration.VerifyCapturedChecks，再做逐 profile 的完整覆盖复验。

定向 normal 回归通过，涵盖 DTO、Runtime 来源/完成报告/候选与最终回执、
领域 attempt/legacy fence、CLI 原 dispatch 与 alias 绑定、pending 查询无写入，
见 [定向 normal r2](evidence/logs/s9-repair-result-scoped-regression-normal-r2-20261014.log)。
其中 hook/policy/submission 三包仅编译，没有命中行为测试；不能称为三包行为
回归通过。五包同目标定向 race 亦通过（workpacket 4.130s、Runtime 27.071s、
repair 3.608s、CLI 2.149s、hookctx 1.168s），见
[定向 race](evidence/logs/s9-repair-result-scoped-regression-race-20261014.log)。
这不是完整原链的 race 运行。扩展原链 r19 normal 48.215s 通过，新增验证
Worker checkout 已清理之后，以旧发行 revision 精确重放仍返回同一不可变
Result/ref/SHA/bytes，Runtime state/journal 零写入；非空伪造 Main verified facts
具名拒绝，execution=complete 时原 Hook 本人责任投影仍有效，见
[最终回执重放 r19](evidence/logs/s9-repair-result-original-hook-chain-normal-r19-20261014.log)。
提取共享真实 candidate fixture 后首次编译缺失原产品 expected bytes 已修复，
失败日志保留。normal-r2 重新验证原主链 54.17s 通过，但 final sealed-pending
恢复在最后公开 CLI ACK 失败：首次 typed Inspect 的 ValidateDelivery 有意跳过
generic completion binding，重入却补入该 binding，触发重跑并改变原 verified
facts；严格回执校验正确拒绝，不能标作恢复通过。见
[恢复缺口原失败](evidence/logs/s9-repair-result-original-chain-and-final-recovery-normal-r2-20261014.log)。
§13.10 已先固定 typed 已接收 Result 的重入必须保持原 checkpoint binding、
Integrate 前后均复验原 receipt；实现仅在 registered typed Result 分支保留原
Inspection，legacy/unregistered 仍沿原 refresh。verified 期间 Main target 前进
也须在 Integrate 前拒绝，保全原 checkpoint/Result/Runtime；该情形自动新建
授权后继验证尝试仍未实现，不误推荐只支持未 verified 状态的 workspace rework。
normal-r4 的原主链 56.46s、final recovery 44.41s 分别通过；后者已真实验证
pending 下公开只读 receipt 零写入、原 Writer exact pair 恢复、公开 integrate
自行恢复 sealed Result 后 ACK/cleanup、immutable facts 保真及真实 Main empty
commit 漂移的具名零写拒绝。测试中的 reset 仅恢复隔离 fixture，不是产品恢复
步骤。见 [恢复修复后的运行](evidence/logs/s9-repair-result-original-chain-and-recovery-normal-r4-20261014.log)。
该整条命令仍为 **FAIL**：candidate 并发测试错误要求原有界集成锁的两个调用
均立即成功，遇到合法 ErrIntegrationBusy 被断言拒绝；现修为仅接受该具名竞争
诊断、至少一个成功，竞争结束后同旧请求精确重试并证明零额外写入。未知错误
仍失败，锁预算未改；normal-r5 三个目标均通过、命令 exit 0，总计 154.724s：候选并发/原接收器恢复
44.26s、公开 Hook→Main 原链 63.92s、final sealed-pending/公开恢复/ACK/漂移保全
46.45s，见 [最终三目标 normal](evidence/logs/s9-repair-result-original-chain-and-recovery-normal-r5-20261014.log)。
三个候选 Hook 的完整响应为 8.126 / 7.998 / 8.180s，原 context 配置不变，
仍不承诺低于 6 秒。候选原 WORKSPACE CAS 保留 retain_last_transition=true，
final Result CAS 为 false；sealed marker 均从实际原 before/after pair 构造。
只读查询拒绝 pending，原 mutation-capable reader 恢复精确 pair；标记引用产物
清理被阻止，candidate/Result 原事件各恰一次。此处不是 hardkill 实验，也不是
整条链的 race 验收。后续 Worker-owned 检查矩阵、final 并发/stale staging、
原 rework 新 lineage、非 pass typed recovery 与 Targeted/Impact/Handoff/fresh S7
及真实平台仍待补齐。已修复重入代码的 CLI/hookctx 定向 race 再通过（2.118s / 1.140s），见
[重入修复后的定向 race](evidence/logs/s9-repair-result-cli-resume-scoped-race-20261014.log)；
范围仍是命中的 helper/入口/责任投影及编译，不是原链整体 race。
本轮日志含原失败/中断，来源与验收范围统一冻结到
[S9 RepairResult 机制证据](evidence/global-optimization-s9-issued-repair-result-20261014.json)。
独立新包安装验收待实际执行，不沿用 PlanReport 冻结包。

r12/r13 曾在原 Hook context 内超时；r13 CPU profile 显示反复展开原 packet/
审批/Schema/来源闭包的开销，已按 §13.10 缩减同次同 revision 重复 preflight 与
trusted-source callback 的重复领域复验，CAS Apply 仍完整重新验证。没有增大
原预算、跨 CAS 缓存外部权限/文件或省略实际 Main 检查。
[超时 r12](evidence/logs/s9-repair-result-original-hook-chain-normal-r12-20261014.log)、
[profile 原失败 r13](evidence/logs/s9-repair-result-original-hook-chain-profile-r13-20261014.log)
与其后诊断日志均保留；r11/r15 为人工中断，不计作测试通过。
早期 DTO regexp/optional shape、reader 编译、专业夹具缺引用及并行接线的
compile-only 失败日志均已保留，
后续通过不覆盖它们。后续代码/Schema/Hook/文档和证据继续补齐。全局核心完整链
**4/9**、真实模型完整链 **0/9**，目标保持 active。


<a id="s9-repair-result-package"></a>

### 7.51 S9 RepairResult 本地包与临时安装通过（2026-10-14）

在 §7.50 的生产代码、测试和出货文档冻结后，独立构建
`s9-issued-repair-result-review`，新目录为
`/home/sebs/.cache/gsb-adapter-work-mp771m5c/s9-repair-result-disposable-20261014-01`。
五目标（darwin arm64/amd64、linux arm64/amd64、windows amd64）构建完成，
解包、206 项 release manifest 精确验证、全新临时项目 `install --packet`、
installed release-graph、doctor、`validate --all` 七阶段全部 exit 0。
构建 452.442s，执行窗口 23:03:18–23:11:01 UTC。原命令、退出码、日志与文件
SHA、完整 status/diff 和前后 inventory 见
[独立包与安装证据](evidence/global-optimization-s9-repair-result-disposable-install-20261014.json)。

包 SHA256 为 `d4235a8c80381a5698418adec8471c730504868bf8298a71bed6f4e3a2f241d1`；
manifest SHA256 为 `321ef706c27285944f7ccd9e332d28d46a824bc51cb76c0de0763f7e47e13469`。
源 HEAD 是 `ed61728e7bfd417dd86dd5d91b4bcae94ca8b742`，dirty=true；实际输入是
整个当前工作树，不把其他既有优化改动归因于本次 S9 切片。构建前后 2406 项
tracked/untracked 非忽略文件清单完全相同，inventory 文件 SHA 均为
`64151745f40a8ee7f75d128f4a9bce2df76c1a47d09b60ba4eadee2f596cb9ce`；
完整 porcelain、`git diff --binary HEAD` 前后 SHA 也一致。

包 manifest 的 platform_acceptance 仍为 `not_attested_by_packaging`：这是本地
交叉构建和临时安装，不是真实 macOS/Windows/Claude/model 验收。没有发布、
tag、提交或安装当前项目；旧 PlanReport -03 包及其证据保持原冻结值。
新包证据 JSON 在 inventory 比较结束后生成，未包含在该包中。本条及 L3/L4
状态补注亦发生在冻结之后；不刷新 §7.50 机制证据或包记录的旧 source hashes，
截至该包验收状态补注完成时，生产代码、测试和出货文档没有追加编辑；
此后新切片另记来源与验证，不继承该包的覆盖。

以下为包冻结时的下一切片待办，后续进展按新增记录更新。先固定 working 阶段 non-pass 的原 `AGENT-LIFECYCLE/work_blocked`
接收与 `blocker_resolved` 后 successor packet lineage。已做只读 API/依赖评审，
当前只有仓库外草案，尚未正式固定 §13.11 或实施；需保留四 verdict、完整答卷，
不要求诚实阻塞作者先报 completed 或全部检查 pass，不另建 Writer/Registry。
特别要闭合原解阻事件/refSHA→新 packet/answer→当前责任/完成报告时间序列，
保持旧答卷不可覆盖，并明确 Apply 复验、不可变 staging 和 workspace→Runtime
锁序。其余 §7.50 检查矩阵及 Targeted/Impact/Handoff/fresh S7、G01–G08 和真实
宿主验收仍是必做项。核心完整链 **4/9**、真实模型完整链 **0/9**，完整 goal active。

<a id="shared-receipt-and-final-retry-progress"></a>

### 7.52 共享回执 Schema、S9 final 重试与宿主能力拒绝（2026-10-14）

在 §7.51 包冻结后继续实施。先固定 §4.4.1 与 L4 §13.5，再补共享
`submission-receipt.schema.json`：封闭七个现有责任状态、原领域接收键格式联合、
S9 candidate/final 的双向 key/status 绑定、本人完成与 Main pending、成对 consumer
字段约束。`current_stage` 仍是查询时阶段，不能推断历史来源。修复 S9 已发题但尚无
candidate 时的公开查询：现在返回具名 `not_accepted`，不再返回空责任状态。

原 Hook 测试直接验证完整原始 receipt JSON，再解码测试投影，覆盖 S5、S7、S8、
S9 PlanReport/candidate；公开 S9 Main Result 查询也校验完整 wire。新增 query 证明
未接收时不写 Runtime/journal。这里没有以少数字段重新序列化后冒充完整 wire 验证。

正常验证四条串行命令均 exit 0（命令墙钟包含编译）：

| 范围 | 墙钟 | 关键结果 |
| --- | ---: | --- |
| Schema + 原 S8 receipt 投影 | 53.532s | 封闭格式/状态/consumer 正反例，历史阶段不作域判别 |
| S5 公开 Main/独立 Worker 回执 | 32.237s | 原接收/路由与 frozen source receipt 两条通过 |
| S7 原 dispatch→Hook→receipt→CleanRound | 54.990s | 完整原链与两层 JSON parser 通过 |
| S8/S9 原链与恢复七目标 | 180.299s | 双 Assignment PlanReport、candidate/Result pending 恢复、final 并发、旧 staging、Main 原链全部通过 |

最后一条 Go package 为 165.595s；其中 final 并发精确重试 24.91s、兄弟原
`workspace prepare` CAS 后 stale staging 重建 30.21s。两者只有一个原
`S9-RESULT-SUBMIT`；同旧请求重试不多写，合法 sibling 保留，真正 stale revision
拒绝不改 staging，刷新 expected revision 后只重建接受坐标、不改专业结果或
已冻结 Main facts。首次新增测试因未用变量编译失败的日志保留，移除变量后的先行
r2 两目标亦通过（59.546s）；不把首次命令改记为 PASS。

随后收紧 Schema 负例，使 Main work-complete/consumer 等约束在正确 key 下独立
受测；最终源码的 schema/submission 定向 `-race` 再通过（包 1.215s / 1.223s，
命令墙钟 81.159s）。该 race 范围不是整条 Hook→Main 回归。

另有一条原 S9 Worker prepare 宿主能力拒绝测试通过（6.68s，命令墙钟 10.432s）：
从原审批/派发、双 PlanReport、Begin、bind 出发，真实
`ExecutionRegistry.PreflightChecks` 拒绝隔离能力，公开 worker-location prepare
给同一具名能力诊断；Runtime pair/revision、Execution 登记、Worker 路径、Main HEAD、
产品 Git status 与 red target 保持。允许 preflight 前产生 ignored Main context
投影；不声称全文件系统零写。该测试本机实际运行、未 skip，**不证明 Worker check
成功或产品命令实际 fail**。本机仍缺支持 user namespaces 的隔离能力；Worker
pass/missing receipt/tamper/真实产品 fail 矩阵待受支持宿主，不能借 native 或 fake
runner 绕过原隔离。S9 PlanReport-derived `Execution.Checks` 与 S6 Record-frozen
`WorkerChecks` 是不同来源，S9 尚无后者的合法发行契约。

新日志、完整命令/退出码、选定源码/文档摘要与保留失败见
[本轮机制证据](evidence/global-optimization-shared-submission-receipt-20261014.json)。
旧 §7.50 机制、§7.51 tar/manifest/安装证据保持冻结；其 tar、安装证据 JSON 与
原运行 metadata 摘要本轮核对未变。旧包不覆盖本轮 Schema/代码/测试/出货文档，
本轮尚未另行打包安装。

下一切片已固定到 [L4 §13.11](l4-mechanisms/L4-stage-form.md#s9-repair-result-blocker)：
working non-pass 走原 Agent lifecycle/CAS，Main 受信 Hook resolution 后有唯一新
Resume packet/answer；不可变 message→artifact、原回执、Stop 与工作完成分别证明。
这仍是**设计固定、实现未开始**；`blocker_recorded` 新 wire/schema 不在本轮七状态内。
机器 stale 的 operational-owner 恢复、completed 但不可集成答案的领域接点亦未闭合。
Targeted/Impact/Handoff/fresh S7、剩余 G01–G08 和真实平台/模型仍是必做项。
核心完整链 **4/9**、真实模型完整链 **0/9**，goal active；无 commit/tag/push/发布。

<a id="s9-blocker-foundation-progress"></a>

### 7.53 S9 blocker 数据模型与原 lifecycle generation 修复（2026-10-14）

在 §7.52 冻结之后，按 L4 §13.11 开始下一实施片，当前只完成数据模型基础和原
生命周期生产者的一处修复。新增 closed `repair-result-blocker.schema.json`、
`RepairResultBlockerArtifact` 与只读 Receipt DTO：保留完整原 binding、execution /
platform session / HEAD、九字段专业答案及 canonical digest、不可变 message ref/SHA；
artifact 不携带自身摘要或预测接受坐标。坐标只属于只读 receipt 投影，不是权限来源。
Schema 自身拒绝 null 专业值和 `completed` blocker；未知/重复字段、错 answer digest、
缺 artifact 的 receipt 均拒绝。原 canonical RepairAssignment / manifest Assignment
两种 execution ID 拼写均受测，权威 registry 匹配仍待原 reader 完成。

实施前置验证实际复现了 `AdvanceAgent` 的 generation 错记：原 Store 在 Mutation
未设 `BaselineGeneration` 时默认 1，generation 2/3 的原
`AGENT-LIFECYCLE/transition_committed` 都写成了 1；generation 1 正常。
[原复现命令](evidence/logs/s9-blocker-lifecycle-generation-normal-r1-20261014.log)
exit 1 保留。现 producer 显式记录原 snapshot generation，在同一 Apply 核对
runtime/generation；未重写任何旧 event、回执或 evidence。

后续验证全部通过：

| 范围 | Go package 时间 | 命令墙钟 |
| --- | --- | ---: |
| blocker 模型 + 原 lifecycle 正常定向 | runtime 0.097s、assignment 0.284s | 54.289s |
| 原 S8 gen-2、S9 PlanReport、S9 Hook→Main→Result→ACK 三链 | investigation 53.467s | 70.488s |
| 模型与原 lifecycle 定向 race | runtime 1.561s、assignment 2.746s | 88.871s |
| 补齐 canonical/manifest ID 拼写后的最终模型 race | runtime 1.569s | 32.909s |

generation 1/2/3 原 readback event 的 key、before/after revision、event type 与实际代数
均验证；旧 activation/blocker-resolved API 正常定向也通过。原 S9 正向链仍实际执行
Main check、Result、ACK/cleanup，不以模型构造数据替代该链。
新模型单测是格式/局部摘要自洽验证，**不是 typed blocker 原接收或宿主身份验收**。
本轮 5 日志含 1 原故障 FAIL、4 PASS，见
[blocker 基础证据](evidence/global-optimization-s9-blocker-foundation-20261014.json)。
§7.52 的机制证据及更早包/证据保持冻结，本片代码不继承其安装验收。

下一步接原 evidence kind/catalog/同 CAS 行、权威历史 reader、私有 blocker/resolution
capability、原 lifecycle CAS guard、Hook/Stop 与 Resume 发题/消费；当前没有新增 Writer、
registry 或公开 non-pass 接收能力，shared receipt 的 `blocker_recorded` 尚未启用。
机器 stale operational-owner、completed 不可集成结果的原领域接点、受支持宿主
Worker check 矩阵、后续 typed consumers 与 G01–G08 仍未完成。
核心完整链 **4/9**、真实模型完整链 **0/9**，goal active；未提交/发布/安装当前项目。

<a id="s9-blocker-reader-progress"></a>

### 7.54 S9 blocker 历史 reader、消息与 evidence 约束（2026-10-15）

按 L4 §13.11 继续实现基础边界，尚未接通 non-pass 生产者。本片新增权威历史
reader：从原注册 packet、完整轮转 journal、content-addressed artifact/message
与唯一 evidence 行重建原 `AGENT-LIFECYCLE` 接受坐标，不依赖当前 Agent 标签、
Worker checkout 或当前源文件；当前 evidence 被 invalidated 后仍保留原接收事实。
该事实不授当前 Stop、完成或写入权限。

审查时先固定了原幂等 key 的唯一 committed 接受契约，再实现核对：重复
PACKET-ISSUE key 或生命周期 key 即使引用其他 evidence 也拒绝，noncommitted
warning 不算第二次接收。合法 candidate-only packet 返回“没有 blocker”；
同 packet scope 的 resolution 行不误算为第二份 blocker。错误 kind/path、孤立行/事件、
摘要漂移和 owner/assignment/source tuple 不一致均拒绝。

原 agent-message 新增闭合、可省略的顶层 `s9_blocker` / `s9_resolution`，保留原
字符串 body 和 legacy 消息。两种 pipeline-owned evidence kind 进入 v1.2 Schema
和 catalog，但不能单独登记、恢复 import 或满足 gate slot。generic RecordEvidence
在读 artifact/Runtime 前具名拒绝。resolution 按 L4 正文共用 `blocker-messages`
目录。新增 blocked PLAN authorizer 直接复用原校验器，在真实 blocked state 验证
原 PLAN/activation，未克隆 state 伪装 working，也不授 resolve/Stop 权限。

最终消息/evidence schema 与 catalog、blocked 授权、Runtime reader 及受影响的旧
PlanReport/RepairResult 用例均通过普通与 race 定向验证。历史 reader 的正例是
**机械构造的完整证据 fixture，不是原 capability Writer 或真实 Hook 接收验收**。
首轮编译失败和旧 fixture 的 symbolic TASK ID 不满足原 message Schema 的失败均保留；
fixture 改用原规则内的 `TASK-901` 后重跑，未放宽 Schema。初次 race 启动后补强了
失效 evidence 的断言，最终较大范围 normal/race 已复验该固定源码。
完整命令、失败及范围见
[本片冻结证据](evidence/global-optimization-s9-blocker-reader-20261015.json)。

下一实施片正在准备 repair-owned opaque capability、原 lifecycle 同 CAS evidence/
Agent pointer/journal、全历史 target 的 generic guard；本片证据不覆盖这些并行的新文件。
实际 Hook/回执 `blocker_recorded`、当前未解阻 Stop proof、Main resolution 与 Resume
仍待接线，机器 stale operational-owner 与 completed 不可集成答卷的额外接收仍待设计。
未新增打包或安装验收；此前所有日志、证据和包保持冻结，不刷新旧 source hashes。
核心完整链 **4/9**、真实模型 **0/9**，goal active，未提交或发布。

### 7.55 真实 Claude 冷启动对比与入口闭环切片（2026-10-15，实施中）

按用户提供的环境启动真实 Claude Code 2.1.299；配置模型别名为 `muse`，底层型号
未独立核验。旧 HEAD 与冻结 dirty 版本各三个独立执行副本，另有搭建/复核会话。
Linux amd64 实验包完成原安装；原始日志、进程结果和独立审查保持冻结，摘要见
[真实宿主探索证据](evidence/global-optimization-real-host-exploratory-20261015.json)。

受控 S1 两边均通过原绑定，REQ 摘要一致，耗时 24.164s/25.225s，不证明提速。
空项目首次 SessionStart 的 additionalContext 为 1182/289 UTF-8 字节，只证明该
字段变短，不代表全会话 token。规格链旧版停 S2，dirty 版本在 S3 达到时限；
均无真实 S5 reviewer 分发或领域回执。真实完整链仍 **0/9**，不把 SDK success
或 Main 自检包计作独立审查通过。探索组漏投验收文件及未固定 Git 参数的工装问题
单独记录，不归咎产品，也不对该组做耗时胜负比较。

用户要求继续后，本切片优先关闭实测入口问题，先固定契约：

1. [L3-S1](l3-stages/L3-S1-bind.md) 规定绑定提示带全部必填参数，真实审批人不由
   Git 身份推导，已提交开发分支事实与本地环境配置明确区分。
2. [L4 控制面 §12.2](l4-mechanisms/L4-runtime-control-plane.md) 规定原 Apply
   错误经已有 conflicts 送达 Hook；ready 只读且明确不评估完整 transition guards。
3. [L3-S2](l3-stages/L3-S2-design.md) 固定无 UI 的服务仍保留 FR→CASE，沿用
   browser_required=false 的五个源文件与两个生成文件，不制造 HTML 或新 adapter。
4. [L4 阶段答卷](l4-mechanisms/L4-stage-form.md) 补 Main S2–S4 已有 issuer 的
   正常领取/精确路径提示，避免真实作者退回手造 evidence 的旧手续。

实现及定向验证进行中；本节不继承旧包验收，不宣称当前新代码已有真实 S5 结果。
下一真实作业采用准备好的业务规格/CASE 输入，经原 CLI/Writer 建立阶段，再运行
原 reviewer 派发、PLAN、保存接收和 gate。不能复制活跃 Runtime 冒充新目录绑定，
不能将 FR 改成 NFR、删除 bridge 或以 fakeClaude 替代真实宿主。

### 7.56 真实 Main S2–S4 接收闭环与 S5 宿主断点（2026-10-15）

§7.55 后按实测先修 Main 精确答卷授权及首次 Worker bootstrap 的 Git 继承识别。
新包 `real-host-closure-r3-20261015` 的干净副本只由测试控制器安装、提交候选业务
文档并通过原 `req bind` 进入 S2；S2–S4 的专业结论与答卷均由真实 Claude Code
会话填写，没有预置规划接收结果。381.509s 后，三个原接收回执均 accepted、
work_complete，原 PTR-PLAN-01、PTR-PLAN-02、TR-002 分别提交，停在 S5 入口。
父会话另以实际安装二进制运行 doctor、validate --all、scenario validate，均通过。
14/14 CASE 是场景设计覆盖，尚不代表业务 API 执行通过。

真实完整阶段链由 **0/9 更新为 3/9（S2、S3、S4）**；旧探索证据的 0/9 保持历史
原值。代码、实验包及 20 个原日志/Runtime/答卷/evidence 文件摘要已冻结，见
[真实 Main 验收证据](evidence/global-optimization-real-main-s2-s4-20261015.json)。
单次成功不证明稳定性或提速；首份 S2 答卷丢失题干被拒，模型修正后才接收。

S5 已真实启动两名独立 reviewer，但当前尚无独立审查接收闭环。实测发现：

- 原 S4 将 TASK/dispatch plan 标为 complete，S5 的 registered/locked 过滤遗漏这些
  产出；S5 也未继承已接收 S4 的设计、ADR 与服务 CASE 冻结源闭包。
- S5 接收/receipt 扫描共享 registry 时误读前序 S2 subtype，须按原 typed reader 路由。
- 宿主 dontAsk 拒绝 `.claude/submissions` 写入。独立真 CLI 对照中，具名 Hook
  allow 下该路径仍被拒，而普通工作目录答卷写入成功；需固定 Worker 本地可编辑
  路径与冻结执行上下文的映射。Main canonical packet/evidence 和原 Worker 权限
  模式不变。依据与实测日志见实验目录 reports/host-authoring-path-probe-summary.json
  及 [Claude Code protected paths](https://code.claude.com/docs/en/permission-modes#protected-paths)。

上述 S5 修复正在按蓝图实施，不能继承 R3 的安装验收。首轮 reviewer 达到原 8 分钟
期限，恢复会话仍未接收，SDK exit 0 不算审查成功。原 activation envelope 被 doctor
误作 full Agent message 的诊断另行记录，尚未修复。下一验收使用新包、新副本和原
PLAN/activation/answer/gate，保留所有失败日志。S6–S10 真实链及 S9 non-pass 的
Hook/Stop、Main resolution、Resume lineage 等全局待项仍未完成。

### 7.57 全局工作状态统一设计与正式输入共享读取（2026-10-15）

用户指出本轮优化目标是减少 Agent 对机制的猜测。真实 S5 派发已连续暴露跨层输入
状态理解不一致：R5 原 S4 接收并进入 S5 后，workspace prepare 仍因 complete TASK
不在 registered/locked 白名单中而失败；此轮没有启动 reviewer，不能计入真实通过。

现将全局工作状态统一契约固定在 [L4 §0.1](l4-mechanisms/L4-state-transition-core.md#unified-work-state)：
以责任 attempt 为单位复用 pending/active/blocked/done/cancelled；阶段、结论、冻结
和权限不混作状态；集中正式输入资格及本人完成读模型，所有 CLI/Hook/Gate/Stop
共用，逐步移除局部字符串分支和重复字段。该模型替代各阶段自定义完成语义，
不新增工作注册表、Writer、持久化投影或 Agent 填表步骤。

第一片已落地 `runtime.CurrentFormalInputDocuments`，复用原 S5FormalInputs 的
已接收 S4 闭包证明；非 S5 仍采用当前代 registered/locked 条件。S5 原文档读取、
workspace context 与 convention context/复核共用该入口，调用者不再自行列状态
白名单。返回只有原 Reference 与 Kind，不把旧 Status 暴露给调用者重新解释。

定向通过 5 个顶层测试：共享 selector 1 个，原 S4→S5 workspace prepare 1 个，
既有 S5/native Builder convention 与 S4 closure 回归 3 个。新原链用例也验证过代
和来源漂移拒绝。初轮新夹具遗漏 workspace bind，次轮遗漏正式 protocol/rules，
补齐原安装前提后通过；没有放宽生产 guard。日志、源码摘要及证明边界见
[统一状态与正式输入切片证据](evidence/global-optimization-unified-state-model-20261015.json)。

接续次序为共享完成/交接投影→文案及必要版本迁移→真实模型闭环验证。五态尚未
全面迁移，S5 新包/真实 reviewer 复测尚未执行；本片不提升已验收计数，真实完整
阶段链仍为 **3/9**，S5–S10 继续。


### 7.58 统一五态落地与职责覆盖（2026-10-15）

本批按 §7.57 与 L4 §0.1 的目标继续落地。新 Runtime 1.3 的 Agent、TASK、S7
ReviewAssignment、Change WorkItem 使用五态；新 RepairPlan 1.1 的 Assignment
从 pending 开始。状态与原事件在既有 Writer/CAS 内共同记录，原活动只在具名边界
还原供原权限检查使用。旧 Runtime/Plan 保留原格式及 journal/receipt 字节，不做
原地迁移。Agent 的简短定义仍直接放在入口 AGENTS。

本人回执、Hook/Stop/Idle、CLI 与派发复用共享模型。Worker 本人 done 与 Main
集成是不同责任；依赖派发须消费本人完成和合格集成。S5 reviewer 的 TASK 登记
不证明 S6 Builder 开始。文档 complete、专业 verdict、stage/phase、资源准备与
集成 checkpoint 保留原含义，不能替代工作状态或权限。

[覆盖清单](evidence/global-optimization-unified-state-coverage-20261015.md)列出所有
实际工作字段、出口与兼容边界。本批工作状态统一、定向回归与真实 Claude 仿真已闭环。
R7 Main 在新格式下经原 evidence/transition 走完 S2–S4；R8 双 Reviewer 的原回执均为
`done`，Stop 正常放行。Spec 的真实 Finding 经 TR-004 退回 S2，由 Main 接手；这不代表
S5 阶段通过。PLAN 保存后给出精确原 begin 命令，进度提醒复用同一责任证明；19 项真实
输入闭包不再触发 16 文件误报，原来源哈希与总读取预算保留。

只有原接收回执与实际迁移可计入阶段验收。全局完整阶段验收仍为 S2–S4 **3/9**；
本批包校验、回归范围、五个旧 CLI 夹具的定向复验、宿主隔离测试限制及仿真原始证据
见[验证记录](evidence/global-optimization-unified-state-application-20261015.json)。

### 7.59 Main 规划入口闭包与 S9 non-pass 公开接收（2026-10-15，实施中）

本轮从真实 R9 Main 暴露的接口断点继续：手工 planning evidence 可推进原 PTR，
但不能替代已发行 packet 的 typed 接收，S4 因缺少 S3 接收闭包无法发卷。固定
packet 模式只经原 answer consumer 登记规划证据，CLI 提前拒绝旁路并给当前
issue/consume 动作；packet-off 旧流程保留。S2 同时移除提前提示 S3 contract_set
的推断。业务修正与可信接收分别验收，不把 Main 自审提升为独立 review。

S9 完整 non-pass answer 的公开 Save/PostToolUse、原 work_blocked CAS、共享回执与
同源 Stop proof 已接线并通过定向机械测试。blocker 是 Main 待接手的未完成责任，
工作状态仍用统一五态；历史 accepted 不授当前 Stop 或写入权限。Main resolution
的公开 Hook、同原 lifecycle CAS 接收与旧 attempt 停权已通过定向机械验证。
新 Resume lineage 和后续 typed consumers 仍待完成。完整阶段计数保持 **3/9**。

本轮进度、失败原因、验证层次与下一步见
[接续记录](evidence/global-optimization-continuation-20261015.md)。


<a id="s9-resume-r36-current"></a>

### 7.60 Resume 合入、R36 选择回归与 S6 当前边界（2026-10-15）

r33 在隔离 Resume 副本的扩大回归唯一失败为旧 issuer timeout，失败记录保留；
r34 修复后指定原链/新链用例通过，并选择性合入 Root 的 34 个文件。Root r35
`internal/agentcontext` 与 `internal/cli` 全量包测试通过（538.172s）；r36
schema/agentcontext/repair/investigation 的选定用例通过（505.392s）。r36 两个
resumed Hook 样本为 8.286s 和 8.543s，均低于原 10s 预算；仅为所测样本，
不外推为普遍性能提升。报告及选择性合并清单见
`r/reports/global-continuation-s9-expanded-r33.json`、
`r/reports/global-continuation-s9-issue-budget-r34.json`、
`r/reports/resume-selective-merge-after-r34.json`、
`r/reports/global-continuation-root-cli-full-r35.json` 和
`r/reports/global-continuation-root-s9-canonical-equality-r36.json`。

after-r14 在独立空目录的安装、doctor、validate、installed release-graph 与 docs
结构 smoke 均通过，见 `r/reports/after-r14-install-smoke.md/json`。该包没有
Resume 合入后的源代码；after-r15 尚未构建。空 Runtime 没有实体行，故这不是五态
实体值或产品阶段验收。

真实 S6 R3 的正确 PLAN/Result 已沿原入口接收，当前只到 `completion_reported`。
Main-owned checks/commit integration 与 TASK-002 仍未完成。R2 的专业错误码和
future-Main 断言通过原具名修订退役；R3 曾有一次真实 429 冷却后同 UUID 成功。最新原 task-integrate 在 0.417s 处因
assignment execution binding mismatch 失败，尚未运行 checks/checkpoint；Root 正修
canonical loader.projectRecord 的同快照 execution 坐标投影及负例。此前 commit-before-report
顺序尚未验证，launcher/frozen-context/docs/AGENTS 提示更新不代表已经恢复。完整阶段验收保持 **4/9（S2–S5）**。

新设计的 Impact adapter 和其 dependency、expectation、兼容性投影尚未实现；
L3 原缺口仍有效。S7–S10 adapters（含 S9 typed Impact/Targeted/Handoff）、Main terminal no_transition
named-recovery 设计、observability/evaluation、G08 real-host factory、G01–G08 剩余验收、
S6–S10 真实链和平台矩阵不因本轮定向
通过而关闭。当前动作与边界见[接续记录](evidence/global-optimization-continuation-20261015.md)。

## 变更记录

| 日期 | 版本 | 内容 |
| --- | --- | --- |
| 2026-10-13 | v0.1.0 | 调查G01–G08与原实现，形成独立分项设计、共用工作包/消费接口、O01–O10实施卡、M01–M10覆盖及迁移/验收方案；修复并复核审查问题，保留待实施/真实宿主验收边界 |

<a id="s6-current-r41"></a>

### 7.61 R37–R41 回归与 R11 S6 R4 当前状态（2026-10-15）

r37 `internal/runtime` / `internal/workpacket` 的 repair-result reader/current-selector
选择用例通过（30.045s）；r38 `internal/hookctx` / `internal/workspace` 的 canonical
execution-coordinate、freeze 与 launch 选择用例通过（25.308s）；r39 `internal/hookctx` /
`internal/cli` 的 assignment revision / native context 选择用例通过（79.690s）。r39
使用的 after-r14 coordinate-hotfix 单 binary SHA-256 为
`607bacb33096d16808aa46505c8914f396437cd0d4ace449b603dd4ac6c37c94c`。

r40 CLI/integration 矩阵 exit 1（132.140s），三项失败保留：
`TestCanonicalS6WorkspaceReworkRequiresNamedAssignmentRevision` 是新 fixture 对
flattened prepare JSON 使用错误 DTO 解码；`TestCanonicalMainIntegrationBlocksActualOutOfScopeGitDelta`
与 `TestCanonicalRevisionReusesOnlyValidatedPredecessorSideParent` 则发现 early-preserve
对所有 `!Ready` 检查过宽，绕过原 `ErrScopeViolation`。Root 已更正 fixture，并将
preserve 条件限于 unmerged 且 SourceHead、TargetHead 或 MergeBase 缺失的情况；有完整
Git 坐标时仍走原 scope/merge-base guard。

r41 CLI/integration 受影响矩阵 exit 0（128.380s；CLI 37.362s、integration 35.236s）。
新增 `TestCanonicalS6WorkspaceReworkRequiresNamedAssignmentRevision`（7.32s）、dirty
pending preserve（1.71s）、原 out-of-scope、predecessor reuse、closing contract、Main
check/recovery 用例通过。该结果只覆盖记录的选择集，不代表全仓通过或代替 R4 实际
race/vet 与 task-integrate。r40 三项失败仍作为历史记录保留，未被覆盖。

R3 的坐标热修后原 `task-integrate` exit 0（2.944s），但将真实 dirty inspection blockers
保存为 canonical checkpoint `a3`，无 merge、无 checks，因此仍不计 S6 完成。当前 R4
通过原 Assignment revision 注册为 revision 4（registration Runtime rev106，Record SHA
`1145cc4b5a9a993dff54824f92b79681892e50fabc156b97bc3fbf4737da1b83`）。原 CLI prepare
生成 execution generation 4，53/53 inputs 和 5/5 declared reads 对齐，
`check_location=main`。Native Builder session `146738c5-08c5-4e99-91a2-65af6df0cb5e`
提交 source commit `e82f91dbe87e01274e45cd0bdfa7af2fa2dcd099` 后，其 Result 在 Runtime
rev115 被接收（SHA-256 `8e4313bcfcfa7e54c8993bd66b831e0e0d46721046281d5aec1922cbfdccf8c0`）。
Builder process 480.604s 后以 exit 1 / `signal: killed` 结束，supervisor 并未标记
timeout；Result 已接收，无需重启。Root 已开始原 Main task-integrate 与所需 serial checks；
最终报告尚未保存，之后继续 TASK-002。Result 接收不是 S6 stage credit，阶段计数仍为
**4/9（S2–S5）**。

after-r15 尚未构建；Resume 的机械回归不计真实 S9 阶段通过。S7–S10 adapters、typed
Impact/Targeted/Handoff、observability/evaluation、Main terminal no_transition recovery、
G08 real-host factory、G01–G08 剩余验收和平台矩阵仍开放。

### 2026-10-15 真实链追加进展

R11 Main 三份 typed 交付及两名独立 S5 reviewer 的 PASS 已由原消费者接收。
共享 packet registry 的 S5 revision 分派修复通过机械回归后，after-r12 原 CLI
在保留 after-r11 安装资产的 R11 宿主恢复原 pending，提交 TR-003 进入 S6。
完整阶段计数保持 **S2–S5，4/9**。其后 Resume 已从隔离副本选择性合入并通过
r35/r36 指定 Root 回归，r37–r41 记录的 reader/coordinate/revision-recovery 选择矩阵
通过；当前发行包仍 after-r14，after-r15 未构建。R11 S6 当前为 R4：原 Result 已接收，
Root 已启动 Main checks/integration 串行验证，TASK-002 未完成；R3 热修重试仅保留实际
dirty checkpoint。S6–S10 不提前计通过，全局工作仍 active；
当前状态和原始证据见[接续记录](evidence/global-optimization-continuation-20261015.md)。
