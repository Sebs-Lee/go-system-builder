# S0–S2 当前机制迁移

## 本轮补充边界（2026-10-19，设计固定／实施待验证）

[GLOBAL §0 本轮固定契约](../GLOBAL-OPTIMIZATION-DESIGN.md#本轮重验后固定契约2026-10-19设计固定实施待验证) 优先。明确人意不重复索取，原 amend 准确入口接原事务及当前卷；S2 接收完成与 S3 准备失败分别呈现，旧完成卷不是当前待填卷，系统发行失败不笼统变作者补料。条件资料以既有 Markdown 实际链接目标/现有 JSON 实际引用结构确定正向关联；裸路径、inline-code 和围栏例子不单独造义务，不用否定词典，真实链接缺失照原错误。architecture-template 的 model 段同步为条件性真实链接指引，旧 fixture 裸正向路径改链接，不作逐字兼容。

S2 保五手写源/两派生载体、实际 allow/reject 语义和必要覆盖；正负数量比仅作风险提示，不作硬门、不凑数。统计引用生成物，不多处手抄；实现者/测试者/维护者三视角只写一段真实语义自审，不新增 direct AC/数量证明。真实无持久变化允许 persisted_effects=[]，字段仍保留、真实效果仍须判断；现有非空 validator/schema 与 omitempty 尚需实施对齐。本片不重构全资料格式、不添字段/预检/状态/proof；真实语义自审保留，无依据证明另审。

AC 混合 FR/NFR/A4 逐项保义，有 FR 就逐 FR 保可达、补充 NFR/A4 不吞 FR，无 FR 才 N/A，不增 A4 存在性资格。全部历史失败/证据保留；本段只是固定设计，未执行验证，未更新全局通过。

## 当前契约与验证边界（2026-10-18）

当前规范为 GLOBAL §0 与 L3-S0/S1/S2 顶部当前最小契约；旧主体和历史通过不构成保留旧门或本轮验收。实施/限定回归done；r91有限宿主范围通过：MiniMax自然exit0、无外部timeout，639.243807714025s，原四题正常保存绑定develop/main准确目标，S2两字段正常保存经PTR-PLAN-01到rev6/planning.contracts/entry2并备S3卷；cross-matrix无null且未借读包外schema。[r91宿主审计](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/s0-s2-reduction-20261018/native-s012-r91-minimax-audit.md)。本次复用8份真实专业源，非从零全S0创作、全局完成、真实发布或性能提升；S0仅复用历史草稿/未授权等待与真实锁定交接边界及本轮机械证据，不泛化更宽业务。原20函数/7包编译构建与host-only包安装证据见[落地记录](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/s0-s2-reduction-20261018/S0-S2-LANDING-REPORT.md)。r3构建与原19函数实际RUN/PASS，源前后不变，r90 linux/amd64包空安装成功且清单0差异。[限定验证与包安装](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/s0-s2-reduction-20261018/verification-20261018T133235.172892Z/package-install-r90.json)。非全仓/全平台验证。


当前r90宿主暴露两项待验修正：S1原四题的Git引用应逐字填写实际选择（如`main`），括注说明放原自由文本，未知目标留空，不加字段/归一化/资格门；cross-matrix模板可选string的branch/no_branch_reason未用项应省略，不写null，保持原schema/validator。到S3不代表此前错误目标记录已正确，r90场景不回改；两项修正已有r91限定机械证据，最新有限宿主复测已通过，范围见当前记录。

- S0没有独立Runtime阶段。原REQ承载专业判断和实际人明确锁定；不新建S0收卷、证据、状态或确认门。同时明确锁定并启动可直接承接，只有锁定未授权启动才等待真实绑定决定。
- S1沿原四问题和TR001：选择REQ、实际授权记录、开发分支、发布目标。系统带出现成事实和准确路径；真实决定正常保存自动消费。姓名只记录不认证，无决定留空。原req bind仅兼容恢复。
- S2沿原architecture_path/conclusion与实际设计工作，正常保存原记录/出口。专业材料缺失具体提醒；不以身份、整包相等、来源hash或重建证明作为接收资格。

全程沿用pending/active/blocked/done/cancelled；文档状态、阶段游标和专业结论不另造工作状态机。原Writer/CAS、当前责任关联、原子幂等和真实并发冲突保留；一次逻辑记录不限定一次物理I/O。

## 实际落点与有限次序

| 小片 | 现落点 | 删除/接通 | 保留 |
| --- | --- | --- | --- |
| S0/S1入口与绑定 | cli/entry_questionnaire.go；questionnaire_observer.go；run.go的runReqBind | 普通Pre准备、保存后要求命令和机械二次确认；真实责任开始/REQ lock出口准备，四问题保存接原事务 | 真实锁定/启动决定、Git目标与已提交REQ对象；不自动commit/push/发布 |
| S2正常接收与投递 | workpacket/planning.go；submission/design.go；runtime/planning_authority.go、planning_submission.go；既有CLI当前卷投影 | Main身份、全packet/Draft相等、开发source hash、重复envelope证明；空结论草稿；原TR001完成口准备 | 原两字段、实际相关UI/服务/ADR设计、当前卷/责任、同答幂等 |
| S2深层事实与出口 | runtime/planning_checkpoint.go、submission_control.go；transition/engine.go的PTR-PLAN-01；qualitygate/evaluator.go；controller/cycle.go、planning_followup.go的S2分支 | 全envelope/evidence/fileview及exact主体重证明；已记未完成同save继续出口，完成旧卷只读 | 已记录专业结论、实际材料可见缺项、原Writer/CAS和原S3交接 |

以上是既有路径收口，不新增作者API、字段、协议、缓存或调度器。实施时相关作者文案同边界同步，不用旧正文恢复资格。

## 入口、异常与幂等

- SessionStart、实际责任开始或原完成出口一次准备；普通Pre/Stop只读当前事实和准确缺项，不做业务准备或来源复证。
- S1已消费申请不能只凭非inactive阻止重放：unbind/rollover可回inactive。初次沿原requirement-binding.md；之后仅读既有.claude/runtime-archive目录名原时间后缀，取最新basename定位当前申请路径，不读历史state/hash。原TR001本身已生成默认归档边界，不因自定义unbind归档目录新增映射；作者不维护metadata，不增加attempt/确认flag。旧申请可读不重绑，新责任授权槽无真实决定则空。
- S2已接收但原出口未提交，同份正常保存继续原出口且不重复登记；已换stage旧卷回放不产生新责任。真实补充与并发冲突沿现记录边界，不新修订/恢复协议。
- 不代人锁定/授权，不猜分支或发布目标；不以自报pass代替实际设计工作，也不以全包历史proof代替专业判断。

## 后续有限验收

验收复用现有直接公共用例与新装宿主实际责任链，本轮原19函数清单及逐项日志见外部落地报告，不新增全仓/平台矩阵。需分别观察准确题干与路径、空决定/草稿零业务写、真实决定正常保存原出口、同答零重复、已记未完成出口接续、旧申请不重绑和真实相关设计材料；没有执行的边界明确列出。

历史entry_questionnaire_e2e的S0草稿/Stop只读及S1保存仍inactive，只证明旧局部行为，不证明新绑定自动链；历史after-r21 S2宿主结果同样不自动继承为本轮通过。无S0独立receiver本身不是缺陷，也不是新增门的理由。

当前执行明细与失败保留见[落地报告](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/s0-s2-reduction-20261018/S0-S2-LANDING-REPORT.md)。本轮这些blueprint记录晚于验证源，不在评估包发布白名单，不改变r90代码/作者资产身份。

历史停点及当前复测见[落地报告](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/s0-s2-reduction-20261018/S0-S2-LANDING-REPORT.md)：旧现场/session保留；r91有限S1/S2正常保存到S3已实测通过，不能将r90到S3或服务恢复本身当通过。

未改的后续文案简化点：AGENTS通用Driver第7步反证与specification-planning三角色自审使作者在validate PASS后继续探针/反证和ADR调整；这是实际专业工作指引负担，不是Hook接收资格，本片不再改源或构包，也不称全部减负已终结。
