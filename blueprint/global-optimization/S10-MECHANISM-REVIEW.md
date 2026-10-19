# S10 机制审查与三片减负规划

## 0. 当前契约与证据边界（2026-10-18）

依据 [GLOBAL§0](../GLOBAL-OPTIMIZATION-DESIGN.md#0-收缩后的实施决策2026-10-18优先于旧方案)与 [L3-S10当前最小契约](../l3-stages/L3-S10-acceptance-audit.md#当前最小契约2026-10-18设计已固定尚未实施)。
三片生产与同片作者文案已落地；下表保留60379d4后实施前定位，行号/“改为”是迁移依据，不代表当前仍有旧门。旧L3主体不构成恢复资格授权。

限定机械证据为r10主程序构建/两包test-c及五原函数PASS，原S11失败因测试helper取旧unknown记录；r11仅修helper并重编CLI/重跑原S11 PASS1.994370s，生产与文档对r10等价，未称r11全六重跑。见[validation-r11-summary.json](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/s10-reduction-20261018/validation-r11-summary.json)。r85为linux/amd64 host-only评估包，构包35.359s、原installer全新空根安装1.370564s均exit0，inventory零差异；非全仓/全平台验证。

真实r85从明确synthetic S7经原TR009的入口开始，Main复用已有真实输出并阅读产品，保存acceptance/audit，原TR015/017到rev7 `awaiting_human_release`；无人工代填专业答案。首段84.640263s自然结束于audit，同会话一次人工continue后136.265592s完成，合计220.905855s（不含续接间隔），不是单轮全自动。见[native-s10-r85-audit.md](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/s10-reduction-20261018/native-s10-r85-audit.md)及[native-s10-r85-installed/.claude/evidence/s10-validation-report.md](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/s10-reduction-20261018/native-s10-r85-installed/.claude/evidence/s10-validation-report.md)。上游synthetic S7非新模型S7 PASS，S11申请草稿已备但decision/disposition/reason/approved_by均未填，无human_decision evidence/真人批准/发布；重复Hook噪声仅观测，样本耗时不证明性能收益。r84两段642.164147s的实际失败与猜结构/旧指引问题保留于[native-s10-r84-terminal-audit.md](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/s10-reduction-20261018/native-s10-r84-terminal-audit.md)，不用作者主观耗时估计。

完整收口与源码/包身份见[S10-LANDING-REPORT.md](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/s10-reduction-20261018/S10-LANDING-REPORT.md)；历史或S9通过不自动计入本轮范围。

## 1. 正常职责与最小Hook边界

保留acceptance→release_audit两个已有专业步骤；TR-009完成准备acceptance，TR-015完成准备audit准确卷。
系统带现成承诺原文、预期、准确来源/责任，不发明oracle/适用性；audit引用既验收结论，不复制全部REQ/Contract/Claim表。
八主题为固定短提示，Main/专业作者判断实际涉及，Hook不建相关性/完整性引擎、不认证N/A理由。
Main汇入真实通知的必要审查结果保存现Manifest并保留原作者归属，不每行另派Agent/另交表，Hook不验身份资格。
S7适用观察直接复用，仅具体缺口/矛盾/本步特有风险补检查，不默认重测、新日志或逐行反证。
行ID/category/owner及Runtime/generation/round由现任务带出，owner不让作者猜；风险责任按实际报告。
原字段分工：expected/source_refs系统已有；disposition本人结论；evidence_refs有现成输出/文件才引用。
oracle仅必要时补判断依据/实际观察/未完成原因，原字符串不强每行非空、不发明反证或文字解析协议；na_reason真N/A一次说明。
已有来源与oracle实际依据可表达履职，evidence_refs空或由此派生的UnsupportedPassCount不作否决，不为数组新建日志/文件。
Hook只需问卷可读、已使用字段能按现有类型解码、准确当前责任关联；读空/unknown/fail与已报阻断，原记录幂等和本步出口。
不得以原fullvalidator“只跑一次”保留语义审查；Hook不判覆盖/证据充分或真实/独立/风险零，metrics只显示。
草稿空outcome可保存原文件，只反馈具体未答项、不算最终提交/自动返工；明确review_required/blocked才沿原负向路由。
同专业同答只读；当前卷真实补充沿现内容ID/RecordEvidence和最新匹配记录，不一概异答拒收、不新修订协议。
仅真实并发/旧轮/当前责任冲突沿原边界处理；换stage旧卷重放不新建责任。
当前clean消费原S7当轮结果与既有失效事实，缺实际工作具体说明，疑似影响交Main，不全盘hash/来源证明。
真实承诺漏项、必要独立审查与相关风险仍由专业人负责，不能以总PASS替代；无对象不造N/A/任务/证据。
风险只说明实际影响/责任/处置，不逐项强要ticket/recovery引用；S11真人决定保留，不扩其实现或工具ACL。

## 2. 入口与保存的实有调用

| 环节 | 当前函数/位置 | 本轮保留与删除 |
|---|---|---|
| 责任入口 | `cli/entry_questionnaire.go:173`→`stageAuditQuestionnaireContext`，`audit_questionnaire.go:21/33` | 原kind/路径与当前责任定位保留；此入口未见caller身份保存准入，不凭空造删除项。 |
| 草稿 | `audit_questionnaire.go:57–64/82/119–158/178` | SessionStart/Pre准备，系统只放ID、expected/oracle/owner空，八区全铺槽；改原TR-009/015责任入口准备，给已有原文/材料，专业人判断相关性。 |
| 普通检查 | `s10_command.go:445–578/604–663 inspectS10Artifact` | 当前读卷仍hash/重建baseline/authority/validator/扫描注册证据；改已记录事实投影，不搬完整证明。 |
| 显式保存 | `run.go:2965`→`questionnaire_observer.go:22–40`→`audit_questionnaire.go:196` | 准确已知路径与原接收保留，不新增typed协议或手工issue/consume。 |
| 脚本/MCP | `run.go:2968`→observer`:61/147–153/219–225` | 复用同保存入口，无需shell解析；当前重探测较重且`:227`后还扫无关packet/workspace，正常S10应收窄。 |
| 专业记录 | `audit_questionnaire.go:244–269/275–318` | 空outcome草稿不当最终提交；最小可读/当前责任与原Writer保留，删机器字段精确对齐与baseline/authority/fullvalidator。 |
| 当前后续 | `run.go:2965–2966/3042–3047/3550` | 保存直接return，下一事件才Controller；完成出口接原迁移，不能把下一Pre当正常后半工序。 |
| 人类报告 | `s10_command.go:58/304/344–357 runS10ManifestRender` | 保存未调renderer，旧L3“自动render”不实；正常完成出口复用现renderer一次，必要交接直接复用RenderMarkdown，不经CLI再验证proof，草稿/同答不render。 |

## 3. 四层consumer重复资格及本次删点

| 层/代码 | 已确认重复或形式门 | 删除/保留界线 |
|---|---|---|
| CLI保存 | `audit_questionnaire.go:253–268` | baseline/authority/完整validator在正常保存重建；收至问卷可读/已使用字段能按现有类型解码/当前责任关联，不换名保留语义覆盖或证据审查。 |
| Runtime记录 | `runtime/evidence.go:43/87/103/137–157` | 又ValidateEvidenceArtifact、ResolveManifest、baseline/authority/fullvalidate；只S10分支去重复，保原记录/CAS，不改其他stage。 |
| QualityGate | `qualitygate/evaluator.go:505–590` | envelope/hash、baseline/authority/fullvalidator、missingS10EvidenceRefs登记资格；改已接收当前结论、空/unknown/fail/已报阻断；不认证证据充分或风险归零。 |
| Transition | `transition/guards.go:313/329/345`→`manifest.go:1227–1302` | 当前clean round真实条件保留；ACC/audit又ledger/hash/全来源重证应删，不整体宣称原guard不变。 |
| 来源构造 | `manifest.go:139/455–610/680/1180–1212` | 整份上游注册hash和envelope→manifest指纹再验证不成为每次正常接收/查询前提；现有机器记录载体不另造协议。 |
| 清单exact | `manifest.go:623–676/1060–1113` | ID/source/path集合missing+extra拒收；系统供给目标，缺真实承诺提醒，不因ID数量/额外诚实材料拒收。 |
| 空类别 | `manifest.go:848–860` | requirement/contract/changed_path硬凑类别取消；没有对象不造N/A。 |
| 单独反证 | `manifest.go:860–904` | exact一行一counterevidence及强制另交证据取消；相关反例责任由作者履行，同观察可复用；空引用不自动当伪PASS。 |
| 八区与重复摘要 | `manifest.go:909–941`、`projectAuditAreaSummary:164` | 八区exact数量、N/A强证据、非空旧摘要第二次自洽资格取消；专业行及可选audit_areas原样保留；展示汇总不改写已报unknown/fail，保存不制造第二份镜像摘要；metrics仅显示。 |
| 身份/票据拼写 | `manifest.go:980–992/1032` | owner=Builder标签判独立、URL/PROJECT-123限定删除；实际独立人、真实风险owner与可定位处理保留。 |

上述证明不得搬到helper/gate/transition；真实工作责任不能借删资格省略，但Hook不替人鉴定观察充分性。
真正存储损坏走原维护边界；诊断兼容函数可留但不能成为正常保存/推进前提。
工作状态沿用五态，原阶段/结论单独投影；不删除全局状态机或真实当前clean round。

## 4. 最多三片落地（每片同改其直接作者文案）

### 片1：原责任入口与问题供给

- 原TR-009/015完成出口准备现卷：已有承诺原文、准确材料与原字段说明，不生成业务oracle或适用性。
- 普通Pre/Stop/重复observer只读，不prepare/fullproof/无关workspace扫描；Main实际安排审查，不加新角色/卷。
- 同片同步协议S10、Skill/模板/examples：复用S7观察、audit不复制验收全表、八主题固定提示而非八任务。

### 片2：原记录与专业责任边界

- CLI→RecordEvidence S10专属链删除fullvalidator/schema重证、exact/hash/注册/身份/票据门；不留“一次fullvalidator”。
- 仅可读/已使用字段能按现有类型解码/准确责任关联；Hook读空/unknown/fail/已报阻断，不判覆盖/证据充分/独立/风险零；metrics仅展示。
- 空outcome草稿正常留原文件，同答只读，真实补充沿现内容ID/原RecordEvidence与最新匹配语义（inspectS10Artifact:461）。
- 原Writer/CAS处理真实并发/旧轮责任冲突，不新增修订协议；已换stage旧卷不造新责任。
- 同片改专业字段说明及无ref自动伪PASS/强制每行oracle、反证、ticket等旧文案，不增观察字段/格式解析协议。

### 片3：读事实与本步完成出口

- status/gate/原TR-015/017消费已记录当前结论和原S7当轮结果/失效事实，不递归复证或由Hook做专业充分性判断。
- 仅实际完成报告执行本步一个原出口并给下一准确卷/既有人闸材料；必要交接投影直接原RenderMarkdown一次。
- 草稿/同答不render/重跑，不每次编辑连跑stage；明确review_required/blocked沿原负向路由，不等下一Pre。
- 反馈展开具体问题、实际路径、已有结论/未完成原因、下一责任；未知原因明确未知，不给泛计数/恢复建议。
- 同片改S10冲突退出指引，保留S11真人决定、不发布、不扩S11实现。

## 5. 文案同片迁移与有限验证

协议S10、acceptance-and-handoff Skill、AGENTS-template.md的S10/通用Driver相关句、docs/rules/release-architecture-audit.md、ACC/audit模板/examples随原三片同步，不另设文案工程；本项为实际入口文案遗漏的待修范围，未表示已修复。
删shortest-path-worst、逐行强反证/证据、空类别造行/全八区数量、copy hash/信封、unknown停工具、strong-block发布等旧句。
当前L3含Hook能做/不能做、原字段与正常任务/反馈示例；旧主体只保历史，GLOBAL同时链接S9/S10。
验证只复用Root限定原例，不铺旧proof矩阵：
- `cli.TestS10QuestionnaireHookUsesPreparedManifestAndMainProducer`：准确卷、草稿/真实补充/重放和本步出口。
- 原manifest直接例按新Hook最小边界修旧资格断言，不以旧fullvalidator洗绿当减负。
- 限定正常S10→S11入口及一个原负向路由：真结论/缺项可读、原cleanround事实、Agent不代真人批准。
机械后独立fresh安装真实专业作者完成acceptance/相关audit；synthetic上游明标，不借S9种子跳S7。
记录实际到达/原错/未知耗时，不用历史count或exit0推通过，不新造同版全矩阵资格。
本轮仅二轮深度规划，未生产/测试/构包/模型，不扩S11。
