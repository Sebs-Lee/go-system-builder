# S11 机制审查与三片迁移

## 0. 当前边界（2026-10-18）

依据[L3-S11当前最小契约](../l3-stages/L3-S11-release-gate.md#当前最小契约2026-10-18)及[GLOBAL§0](../GLOBAL-OPTIMIZATION-DESIGN.md#0-收缩后的实施决策2026-10-18优先于旧方案)。本片正常保存自动消费与原退回责任交接已落地，限定实施/回归/宿主scope完成；机械与包安装分轮有证，不声称最终同版全矩阵。

r88真实return通过普通保存TR029→本人补审→TR017→新空S11；最终r89真实defect通过普通保存TR027→真实Case/问题/dispatch/packet，发行后多次Pre保持current、无false-stale、Objective来源中性。见[native-s11-r88-audit.md](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/s11-reduction-20261018/native-s11-r88-audit.md)、[native-s11-r89-audit.md](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/s11-reduction-20261018/native-s11-r89-audit.md)及[S11-LANDING-REPORT.md](/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z/diagnostics/s11-reduction-20261018/S11-LANDING-REPORT.md)。r88旧stale、r87 setup与r86命令路径证据原样保留；旧错不算首次成功。上游synthetic S7/复用实际prior S10、模拟操作者选择不是本项目真实发布授权；未执行Investigator/完整S8或全局完工。模型wall不等于Hook性能，重复提示噪声仅为观察到的文案负担。

系统复用S10已接收可读报告、实际结论/风险及准确路径，人明确六处置及必要理由。Main如实填已预备问卷原disposition/可选reason并保存，其余metadata由系统带出；PostToolUse自动调用既有原Writer入口，在同次逻辑记录消费并执行原出口，不要求正常CLI或第二凭证。空选择草稿不写Runtime；无真实选择留空等待，不默认批准。真实人意是职责要求，Hook不认证身份、不加human_confirmed或额外证明；actor只是系统已有调用记录。原human-decision/decision-evidence只兼容恢复，不作正常职责或验证步骤。

TR027的原S8 intake仅对明确人类问题来源作局部兼容，可放宽既有Batch专属必需约束；不新增字段/schema/状态，不把decision ID冒充Batch或虚构Finding，不重写全S8。人类Case本就不存在的可选Batch/digest字段与现binding空值是同一事实，不构成Case变更或stale资格；不新增来源证明。

## 1. 明确删除 / 保留

| 删除额外资格 | 保留实际职责与内部一致性 |
|---|---|
| 人/Main重审S10全包/hash/开发源/validity登记证明 | 当前已接收结论/真实风险与准确候选对应；缺实际材料准确提醒 |
| 人手填Runtime/context/decision ID、逐字命令、revision算术 | 人明确真实处置与必要理由，Main忠实搬运，不推断批准 |
| 缺先注册finding/placeholder拒接缺陷驳回 | 实际问题/现finding交S8原责任；不把路由当BUG已调查 |
| Pre/Stop准备业务、下一Pre补交接、复用consumed旧申请 | TR017准备一次；读当前报告定位新决定；普通监听只读 |
| 原S10旧pass在退回后重新自动推进 | 保观察、交原reason/准确路径，清被退回步骤旧最终结论，明确待办 |

保留原Writer/CAS、当前责任/明确选择关联、逻辑原子记录、消费与幂等事实。不得搬旧proof到helper/consumer，或新建状态、字段、schema、API、角色、缓存、队列、提交协议。五态是工作状态，六处置与既有生命周期游标另属原语义。

## 2. 现有落点与三小片

下列三片保留本轮迁移依据与原定位；当前已落地范围及验证以§0为准，行号不构成新运行协议。

### 片一：入口展示

S10→S11交接：audit已收卷但原TR017未提交时，下次同份正常保存必须继续既有出口，不重复登记；已完成跨stage回放不重推进。不能用registered/current evidence的early return截断正常自动交接，不新增重试状态/证明/协议或调整预算。

- `cli/human_release_questionnaire.go:15 stageHumanReleaseQuestionnaireContext`、`:77 humanReleaseQuestionnairePath`、`:96 prepareHumanDecisionApplication`、`:123 humanDecisionApplicationContext`：将一次准备接原TR017完成口，普通Pre/Stop只读；系统展示S10现可读报告/风险/准确路径与六处置去向，不要求对账整包。
- 返回S11按当前已知S10报告定位新决定，消费过的旧申请不复用；沿现对象，不造attempt字段。
- 同片同步L3/协议/AGENTS的人闸说明：等的是人决定，不拦普通工具，不把无风险变成必填“唯一未决事实”。

### 片二：原决定消费者

正常提示直接给原问卷路径、六处置/原两字段含义及自动保存出口；兼容恢复命令help如实说明原参数，不把作者引去猜actions。

- 原S11问卷保存接收与PostToolUse监听复用当前路径/原disposition与reason，直接调用已有decision/Writer消费入口，自动走固定出口；草稿不写Runtime、同答不重跑、换阶段旧卷不生成新责任。`cli/run.go`的原human-decision仅兼容恢复，不在正常保存后追加命令。
- `transition/engine.go:792 validateHumanDecisionScope`、`:842 validateHumanDecisionArtifact`及原Writer调用：删重复来源/hash/开发源/身份标签资格；保当前责任、明确选择、消费事实、CAS与原子一致性，不另建许可层。
- `runtime/s11_migration.go:23 HumanReleaseTransitionID`复用原六枚举映射；不新增target或状态机，不代做发布。
- 同片同步命令帮助/示例/协议：人不拼机器信封，不承诺本地actor字符串完成身份认证；无真实明确决定不接受批准。

### 片三：返工与作者文案

- 原TR028/029将人reason及准确路径交原S10责任；保原观察，清被退回步骤的旧最终结论，不能旧pass冒新答。
- 原TR027接实际问题/现finding交S8，不要求先登记finding才让人选择；不造占位finding或假调查结果。
- defer/abort/approve仍沿原出口，批准不执行发布；外部发布及rollover不是本片实现。
- 同片同步AGENTS生命周期逐字命令句、协议S11、Skill、release/handoff指南与现模板：只复用已接收材料，不再添Agent/人重复表格或全量证据清单。

## 3. 有限验证与诚实界线

后续只选既有六处置公共路径及一次退回S10再到S11往返，保原日志/幂等/真实并发边界；具体原用例由实现后核定，不新造全矩阵或附加通过门。

验证必须看到：真实明确处置才写决定；同答不重复消费；缺陷问题不需假finding；退回旧最终结论不能自动推进；普通Pre/Stop零业务写；重新到S11不复用已消费申请。无明确选择不默认approve。

Claude仿真由测试操作者显式给模拟决定，必须标evaluation，不是本项目真人发布授权；记录实际材料/原入口/分支与耗时，不用历史S10计数代替本轮。Linux评估/机械/模型观察各记范围；不冒全平台或真实发布完成。

旧L3主体、历史scope/hash/完整命令和“尚无draft”描述仅作追溯，不是本轮规范或当前接线结论。当前限定自动两分支已有上述事实，最终defect发卷后只读监听无false-stale；Investigator实际执行/结果不在已完成范围，旧命令证据不继承为新链或完整S8验收。
