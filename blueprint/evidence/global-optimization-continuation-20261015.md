# 全局优化接续记录（2026-10-15）

全局工作状态：`active`。真实模型完整阶段验收为 **S2–S5，4/9**。
统一五态切片的完成记录不代表全局优化完成；包构建、机械测试、真实模型个人交付和阶段
通过分别列证据，不按进程 exit 0 或游标位置推算成功。

## 当前接续快照（收缩决策 / 2026-10-15）

| 工作 | 五态 | 当前证据与下一步 |
| --- | --- | --- |
| 真实 S2–S5 完整链 | done | 4/9；原接收与阶段转换均有证据 |
| 真实 S6 实施/集成 | active | R4 原 Result 已由原入口在 Runtime rev115 接收；原 Main task-integrate 3.039s 返回 preserved/blocked：Bootstrap 系统资产被当成 dirty；未 merge、未运行 race/vet。该缺口待按收缩设计解决，归档协议不继续实施。R4 仍无 S6 stage credit，TASK-002 未完成。R3 热修重试保留真实 dirty inspection checkpoint a3，无 merge/check。R40 三项失败保留；R41 受影响 CLI/integration 回归 exit 0。R2 因专业错误码及 future-Main 断言经原具名修订退役 |
| S6 冻结输入闭包修复 | done | after-r14 新 revision 原 prepare 闭包通过，53/53 冻结输入校验；旧执行未补写 |
| S9 Resume 实现与定向验证 | active | r34 原链通过并已选择性合入 Root；r35/r36 Root 回归通过，r37 current selector 选择用例通过。r36 中 resumed Hook 样本为 8.286s/8.543s，均低于原 10s；这不是普遍性能结论。after-r15 Linux amd64 评估包已构建、fresh install 结构验收通过；尚无真实 S9 阶段通过 |
| Hook 监听/问卷收敛切片 | done | r59/r60/r62 定向验证；after-r18 Claude 实际填卷保存并由原 PTR-PLAN-01 推进 S2→S3；全局完整链未新增阶段 |
| 独立 Impact / Bootstrap 归档扩展 | cancelled | 按用户减负要求停止实施；隔离代码保留参考，不合入。后续领域需求复用原入口，先验证 S6 一个完整闭环 |

以下各节按时间追加并保留失败历史；当前状态以本快照及下方收缩决策为准。

## 最新评估包与真实填写验证（after-r18）

Linux amd64 隔离评估包构建 exit 0（16.387s），归档 SHA-256：
`2391fd64e7100a417bc5dd5b13c486249b1c179ddcd7c9d22587d0770f944e6c`。
只用于评估，不是五平台发布，也未替换 R11 活跃安装。

全新安装/绑定后，原生 Read 自动准备准确绝对问卷路径，重复观察不换卷、不覆盖答案。
真实 Claude 按一般工作请求独立读取问卷和必要业务文件，运行原 scenario bridge /
generate / validate，填写 `architecture_path`、`conclusion` 并保存。原消费者接收，
submission `control_status=advanced`，原 transition 为 `evt-ptr-plan-01-r7`；
Runtime 从 `planning.design` 进入 `planning.contracts`。模型自然 exit 0（79.360s，
宿主报告 18 turns），没有手工 issue/consume，按要求未开始 S3 工作。
该项目的设计和行为包来自测试题材，模型完成核查与问卷交付，不据此宣称从零设计。
这是最新包的单阶段真实填写/原图推进证据，整体完整链仍为 S2–S5，4/9。

原生 doctor、validate --all、docs check 均通过。第一次文档检查把命令拼成
`docscheck` 返回 unknown command（exit 2），原报告保留；正确 `docs check` 复跑
exit 0。原始报告在外部 `reports/r18-questionnaire-probe.json`、
`r18-questionnaire-structural.json` 和 `r18-questionnaire-docs-corrected.json`，
路径/哈希/原接收记录见同名 JSON 接续证据。没有全仓回归声明。

### 下一个已定位问题：原 S6 安装资产识别

只读复查 R11 的当前 R4 Worker：未提交项是 `.claude/agents/backend-builder.md`、
`.claude/bin/loop-harness`、`.claude/settings.json` 及 Bootstrap/frozen-input/pointer
文件；产品目录 `internal/reservation` 和 `go.mod` 没有 dirty status。原 Bootstrap
receipt 已列出 118 个精确文件及哈希，Main execution 的 bootstrap_sha256 已绑定该
receipt；现有 ValidateExecution/ValidateBootstrap 校验这些事实，无需再建归档协议。

原 Inspect 的 worktreeClean 仅识别固定 Runtime 投影，导致安装文件被判 dirty。
后续须在 Inspect、Main 检查和 cleanup 原链审视相同分类；不能只让 Inspect 通过，
也不能忽略整个 `.claude`、把系统资产当产品提交，或用强制清理掩盖 dirty。真实产品
改动和系统文件哈希变化仍要被原检查发现。这仍是待实施项，没有执行 Main race/vet、
merge 或改动 R11。完整链仍 4/9。

## 监听与问卷第一切片（r46–r52）

已落地 PreToolUse graph 监听、取消工具策略/身份/scope 拦截与超时阻断；S2/S3/S4
复用原 issuer 自动准备问卷，进度上下文优先保留准确 answer 路径，保存反馈带回该路径。
S6 workspace 的 PLAN/Result 保存分别调用原 begin/report，不增加接收权威；r52 中
原 canonical S6 用例以保存驱动通过（6.77s），同 Result 重放零新增 Runtime 写入，
后续真实 dirty 检查与具名返工约束保留。S5 正/负路由与 S2–S4 原 followup 测试通过。

r50 完整 Hook/Controller/migration 通过（20.011s）；r52 定向 CLI/Hook/Controller/
migration 全部通过（57.476s）。r46 的旧函数编译引用、r49 的旧重复 denial 提示断言、
r51 的两项旧 MCP 工具拦截断言失败均保留；按用户新契约更新测试后复跑通过。
本轮没有全仓绿或真实阶段完成声明。after-r16 Linux amd64 包已构建；全新原生 Read
自动从零生成一份问卷并给出准确路径，重复仍为一份且保留答案。真实 Claude 两回合
读取正确问卷并列出准确填写路径/问题，exit 0（21.296s），仅为只读入口探针。
它仍引用内部缺项 token；r53 已补为填写/保存动作说明，missingtokens/Hook/Controller
完整包通过（32.407s）。后续 after-r17 探针及 native/S9/Stop 接线见下节；
R4 系统资产集成 blocker 仍未闭合。真实阶段仍为 4/9。

## 问卷路径、保存与提醒的接续验证（r54–r60）

S6 native/workspace 共用原上下文投影的准确 PLAN/Result 路径，保存分别交给原
begin/task-complete 或 workspace begin/report。已接收 Result 不要求重交；Main
集成待办单独提示。S9 fixing 使用原登记 source reader 给出准确答卷位置。Stop/Idle
只读取得当前问卷，反馈原缺项及准确路径，不在停止时发卷、恢复或改 Runtime。
正常角色卡、技能、阶段协议和 workspace 指南同步到“直接给问卷、填写并保存”；
PreToolUse 观察事实与提醒，不承担工具身份、权限或 scope 拦截。

- r54 发现 native Task 在 PLAN 前为 locked，错误过滤导致看不到问卷；修复后 r55
  native PLAN/Result/replay 与 workspace 保存用例通过。
- r56 CLI 与原 S9 candidate/Main integration/final 链定向通过。
- r57 发现 Stop 信任传入 runtime_context 会跳过真实工作区加载；改为读取实际
  Runtime 后，r58 不完整 Stop 带准确路径、零 Runtime 写入及本人完成出口通过。
- r59 六包定向回归通过（151.284s），覆盖 CLI/Hook/Controller/missingtokens/
  submission/investigation；包括原 S8/S9 链，不代表全仓回归。
- 最终提示文字调整后 r60 定向回归通过（80.817s），覆盖规划路径、native/workspace
  S6 保存、重放、具名返工及 Stop。Controller 仅编译，无匹配测试；其测试证据用 r59。
- after-r17 Linux amd64 评估包构建通过；原生探针收到完整绝对填写路径、具体
  architecture_path/conclusion 工序，重复监听保留已填答案。fresh doctor/validate/docscheck
  均 exit 0。第一次真实 Claude 只读探针达到限定回合，记为未完成；全新项目重跑
  自然 exit 0（42.243s，宿主报告 10 turns），正确找到问卷、题目和距离 S3 的缺项。
  这是入口可用性证据，没有修改产品或新增阶段 credit。after-r17 早于本节后续
  native S6/S9/Stop 改动，不能作为后续代码的真实宿主完成证明。

删除保存 adapter 中五处重复 WorkspaceBoundaryDecision 工具策略调用。r61 扩大
回归中原 S9 candidate、并发、sealed-pending、Main integration/final 与 ACK/cleanup
用例通过；整体 exit 1 来自旧 PlanReport “PreToolUse 必须 deny 产品写入”的断言。
该失败保留。进一步移除 S9 PlanReport 提示里的写入授权筛选，已接收后仍返回准确
问卷路径和 Main 下一工序，S7/S8/S9 残留手工 issue 正常提示也已清除。
r62 更新断言后定向回归 exit 0（101.578s），原 S9 PlanReport 原 Writer/重放及
candidate → Main integration → final Result 链再次通过。其余 r61 通过项早于这次
S9 提示修改，不作为新代码的完整回归。

尚未闭合：真实 R4 Main 集成的系统资产 dirty、原 Main race/vet、TASK-002；S9 的
通用 completion 与领域 Result 仍有重复工序；S6–S9 真实完整链尚待验收。阶段仍 4/9。

## 设计收缩与 after-r15（2026-10-15）

用户指出 Hook 与接线持续增重，已停止新增独立 Impact 和 Bootstrap 归档协议。
按用户澄清，Hook 的目的固定为 graph engineering 的监听、遗漏提醒和流转。
直接给出履职问卷，明确工序、每题要求和填写路径；保存后更新原 graph，并提示
距离下一 stage 的具体缺项。取消身份/权限/scope 拦截；问题和缺项共用原阶段要求，
不新造手续与验收。第一切片状态见上节，真实阶段仍为 4/9。详见全局设计 §0。

此前已完成 after-r15 Linux amd64 评估包与 fresh install 验收：install、doctor、
validate、release-graph 和 docs 均 exit 0；202 个源条目和 204 个安装回执条目匹配。
外部报告为 `reports/after-r15-source-inventory.json` 和
`reports/after-r15-install-smoke.json`。此包早于本次设计收缩，不包含新 Hook 方案，
也未替换 R11 活跃安装。真实阶段仍为 4/9，R4 Main 集成仍 blocked，race/vet 未运行。

## 本轮已经定位的问题

- R8 两个真实独立 S5 Reviewer 均完成个人责任；Spec 的五条 Finding 触发原 TR-004
  退回 S2。这是负向路由通过，不是 S5 阶段通过。
- R9 Main 实际修正四份业务文件，原 scenario 校验为 14/14，contracts check 通过；
  两次手工 planning evidence 触发原 PTR，但两份已发行 answer 仍为 null，
  `entities.submissions=[]`。S4 无法从合格 S3 typed 接收重建来源闭包。
- S2 在 design 未被接收时误提示 `contract_set`；已修为当前 S2 的
  `architecture_record`。文件存在或 locked 不能推断阶段责任完成。
- Main skill/protocol 仍教手工 `runtime evidence add`，与 packet 模式要求冲突。
  本轮统一正常操作路径，并在 CLI 写入边界拒绝 packet 模式的规划旁路。
- R9 的 Claude Main Edit 曾被宿主 dontAsk 拒绝。R10 Main 按用户已授权命令使用
  bypassPermissions 并加载项目 Hooks；独立 Worker 仍使用原 launcher/dontAsk。

## S9 non-pass 公开接收切片

保存完整专业 answer 后，`SaveRepairResultFrom` 按 verdict 调用原 candidate 或原
blocker 消费者。blocker 沿原私有 capability 和 `work_blocked` Writer/CAS 接收；
从 prepare 到 CAS 复用原 workspace lock/context。没有第二个 Runtime writer。

回执固定为 accepted、work_complete=false、Main consumer pending；领域 subtype 为
`blocker_recorded`，通用 status 只用五态。当前 original blocker/evidence/PLAN/source
证明成立才投影 blocked 并允许结束本人执行，历史接收不能授 Stop 或写入权限。
blocked 且 Main pending 的 Worker 下一动作是 `await_consumer`。

已通过的机械验证：

- `global-continuation-s9-blocker-r1` 中公开 PostToolUse、原 CAS 与 Stop 测试通过；
  该轮整体 exit 1 来自新增 schema 测试夹具的 field_errors=null，未改产品规则规避。
- 修正夹具为数组后，`global-continuation-s9-blocker-r2` exit 0，覆盖五态、闭合 receipt
  schema、伪造 private proof 拒绝、当前 blocker 责任和 S2/S3 提示回归。
- 扩大回归 `global-continuation-s9-regression-r3` exit 0（276.942 秒），同时覆盖原
  candidate/Main integration/final Result、并发与精确重试、sealed-pending 恢复、
  PlanReport 原 Hook 边界和本轮公开 blocker/Stop；未用 non-pass 接线替代原正向链。
- 公开 Hook 测试覆盖错误 actor/CWD/session 零写入、精确重放零写入、无 candidate/final
  Result 旁路，以及 answer/source 变化时 Stop 拒绝。它使用机械平台 fixture，
  不能记为真实 Claude S9 阶段验收。

## 下一步与尚未完成

1. Main CLI 护栏、操作文档与原 S2/S3/S4 typed 接收链回归已通过：
   `global-continuation-main-entry-r4` exit 0（74.990 秒）。无发行/当前发行时给出精确
   issue/consume 动作；拒绝先读取猜测 artifact；pending 时不恢复事务；旧格式
   packet-off 与非 planning kind 保持原行为。护栏使用真正只读 Runtime store。
2. 冻结新包并校验 manifest，再用 R9 真实修正的 18 份业务输入启动新 R10 Main；
   不复制旧 Runtime/journal/answer/receipt。之后做独立 S5 与原 S6 实施/集成验证。
3. S9 Main resolution authoring/capability、原 CAS 和旧 attempt 停权已通过公开
   Hook 定向机械测试；Resume lineage 与后续 typed Impact/Targeted/Handoff 尚未完成。
4. S7–S10 真实完整链、G08 仓库内 real_host 工厂、跨平台及其余 G01–G08 验收仍待完成。

## R10 真实 Main 与 S5 接点（修复中）

after-r10 Linux amd64 评估包构建与原 release graph/manifest 校验通过；archive SHA：
`6d7dccf977c6d7a3b3bf075fb4a08dd174051762ac4107803d82e2656d27fa4e`。
真实 Main 两个 session 分别为 `d57232c8-0abb-4b03-a959-c0e9d55a714a`
（320.099 秒）和 `0c91b2db-b108-4318-a740-1c82d9d77a74`（210.056 秒），
均自然 exit 0、无超时。三份 S2/S3/S4 typed answer 均得到原 accepted/work_complete
回执，Runtime 到达 S5。

但 S4 的原 system followup 尚 pending 时，另一次普通 Controller cycle 先执行了
TR-002，未在同一事务关联 S4 submission 的 advanced/transition_event_id。
注册 review workgroup 没有改变该接收行；随后一次原 recover 按已变化 origin 将其
标为 replaced，S5 原 issuer 因没有合格 S4 tuple 正确拒绝。没有 S5 packet、审查员
执行或阶段通过；不对 R10 补造原事件，也不把游标 S5 算作 S5 验收。

修复已接线：普通 Main planning 自动推进从同一 snapshot 选择唯一当前 typed
pending followup，交给原 transition CAS 同时消费；显式 followup、原门与来源证明
继续使用原路径。`global-continuation-controller-followup-r12` exit 0；
`global-continuation-main-cli-r13` exit 0（70.624 秒），三条 S2/S3/S4 真实 typed
接收后 pending→普通 Read 自动迁移的机械回归通过，S4 后原 S5FormalInputs 合格。
这一修复不追认 R10 的旧事件。R11 将使用新冻结包和仅 18 份业务文件的新副本验证。

R10 的 38 次 `invalid packet artifact path` 另已定位为 Main milestone.read 的
`docs/architecture/` 等目录被当成单文件读取；不是正式 packet SHA 校验失败。
仅诊断分支现复用原有界 walker 安全展开目录，正式输入仍逐文件校验 SHA。
r13 同时通过 S2/S5 实际读集归因、目录外不识别、缺失/链接目录拒绝、256 entry /
4 MiB 预算与 formal SHA mismatch 负例。原 S2/S3/S4 接收和旧入口兼容回归也通过。

已知独立恢复缺口保留：Main followup 若已被原评估标为终态 `no_transition`，之后门
条件变化时不能把它伪装为 pending/advanced。本轮只接 pending 原接续；该情形仍需
具名 revision/recovery 设计，S5 多 reviewer 的 no_transition 语义未改变。

S9 Main resolution 已独立接线。原历史 reader/current packet 定向验证
`global-continuation-s9-resolution-reader-r8` 已通过；先前 r5/r6 的编译及 reader
事件识别错误保留在原日志。公开 Main Hook/同 CAS/停权整片
`global-continuation-s9-resolution-public-r10` exit 0（93.355 秒），覆盖错误 Main
身份与失败 save、空 rationale 草稿拒绝、原 CAS、contract_conflict→S8、旧 answer
及产品范围拒绝、历史 Stop 失效、精确重放零写与改字节冲突。先前 r9 的老格式
夹具状态断言、原 session 边界更早拒绝的测试断言已修正，未放宽产品守卫。
新片尚未打入 after-r10 包；包/真实模型验收不能继承其仅 blocker 切片的结果。

## 原始记录

本轮包摘要、成功与失败测试的原命令/结果及报告 SHA 见
[机器验证记录](global-optimization-continuation-20261015.json)。after-r11 评估包
archive SHA 为 `0aa2b80b848dae33677834670f60646c817b98c359c205c96f7fc1c1ef72eda7`；
fresh install 与真实 Main 已启动，结果另行更新，不提前计入阶段验收。

实验根目录：
`/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z`。
该目录的 `reports/r9-main-live-chain-review.md`、
`reports/r9-main-s4-issue-diagnostic.md`、`reports/r10-main-refined-seed.json`
保留会话、原诊断、18 项业务白名单及逐文件摘要。旧 R8/R9 包和项目保持冻结。

## R11 当前接续

真实 Main session `a3b27212-e73a-4808-b4f2-718bbbc2f1bf` 在 240.049 秒自然
exit 0。三份当前 typed answer 均 accepted/work_complete/status=done；S4 原接收行
`control_status=advanced`，精确关联原 `evt-tr-002-r16`，原 S5 prepare 成功。
两名独立 reviewer 使用原 dontAsk launcher，19 份正式输入逐文件 SHA 相同；
正在审查，不提前记 S5 完整阶段通过。R10 缺失关联的旧事件保持原样。

`global-continuation-s9-full-regression-r14` exit 0（334.280 秒），覆盖 runtime、
assignment、submission、hook 与 investigation 中原 S9 RepairResult、公开 blocker、
Main resolution、原 lifecycle、candidate/final/recovery 与 PlanReport 边界回归。
该测试对应 after-r11 冻结生产源码；下一 Resume 切片在独立源码副本先固定设计
并实现，不能继承本轮尚未运行的 Resume 验收。

R11 两位原独立 reviewer 均自然 exit 0（172.6 / 186.9 秒），专业结论均 PASS、
findings=[]，各自原接收 accepted/work_complete/status=done。S5 门 missing=[]，
但 TR-003 revision 守卫误将原 S2–S4 packet 按 S5 解析，报
`current revision: revision packet cannot be resolved`。原 recover 未迁移；当前仍
不记 S5 完整通过。已在原守卫中限定 S5 subtype，保留坏 S5/具名 revision 的全部
验证，正在机械回归，后续须新包+冻结状态恢复验证。

编排检查记录：r15 CLI 全量尝试在编译期退出（0.892 秒），原因是一个子代理使用
默认 cwd 将 Resume 部分接线误写主仓，依赖仍在隔离仓；不是 CLI 测试通过记录。
四个文件的新增内容已完整备份并迁回隔离仓，主仓逐字恢复冻结 after-r11 生产
基线。后续 S5 修复为独立增量；不会把中途编译失败记作已运行测试。

r16 混合 registry 正例通过，但新增坏 S5 测试夹具缺少完整 binding，原 decoder
将其读成零值并按不同 Runtime 跳过；未因此扩大产品行为。改用 binding 类型错误
直接验证原 decoder 边界后，r17 exit 0（6.677 秒）：混合 registry、具名 revision
原接收/替换/并发/拒绝例和 TR-003 canonical baseline 正负例通过。after-r12 包
原 graph/manifest 验证与 host build 完成，archive SHA
`9e1eb8ffd2709eb1c3b565dd86aa0bdd7d4e55782a6dc21c77bdfb83bd7b47cd`。
真实恢复将显式调用包内新 CLI，R11 仍保留原 after-r11 安装资产/manifest；不静默
覆盖安装文件，不改 Runtime/作者答案，混合版本验证范围另行标明。

R11 还定位到一个直接增加作者试错的问题：S3 applicability 的原题未声明
`{disposition,rationale}` 对象形状，Main 尝试字符串、verdict 后才成功。新 fresh
template 将直接写明 FE/BE/SYNC 三键及两字段，不预填任何专业判断。R11 目录
采样修复的原 native/stream 中 `invalid packet artifact path` 字面次数为 0；R10
为 38 次。这是该告警消除的实测，不能外推成不同输入轮次的整体性能提升。

## R11 S5 正式恢复通过

after-r12 原 native CLI 在原 R11 宿主执行
`runtime stage-answer recover --root <R11> --agent reviewer-spec`，自然 exit 0，
原 pending 接续提交 TR-003，原接收行仍 accepted/work_complete/status=done，
进入 S6。调用前 state/journal SHA 与冻结诊断相同，未改 Runtime、packet、专业答案、
registry 或安装资产。命令与差异证据保存在
`reports/r11-after-r12-recovery/{command.json,stdout.txt,stderr.txt,state-delta.json}`。

此项明确为 after-r12 外部 CLI + after-r11 已安装 Hooks 的恢复验证，不能称为
after-r12 全新安装整链通过。结合两位真实独立 reviewer 的原 PASS 接收与原 TR-003
事实，S5 完整阶段现记通过，计数更新为 **4/9（S2–S5）**。S6 正在准备真实
实施/集成；不预记其完成。

## Resume 隔离验证与当前真实 S6 阻塞

r18 实际执行 CLI 整包回归（464.990 秒），报告 1 个失败：
`TestProgressProjectionIgnoresRevisionClockAndControlBytes`。目录读取 helper 绕过
原 addSource 对 Runtime/journal observer bytes 的排除；隔离副本已恢复该排除，
r19 原用例及 S3 题面、S6 读单/预算用例通过。没有 CLI 全量绿色结论。

r19 在 `sources/resume-after-r11-worktree` 实际编译与测试（312.121 秒）：
workpacket/runtime/CLI 新用例通过，公开恢复后新 AnswerPath/旧 Stop 测试未报告失败；
但第二次 blocker 和 resumed candidate 接收超出原 Hook 预算，candidate Hook 实测
15.927 秒，最终退出 1。不是完成记录。Resume 生产修改保持隔离，正在消除同一
快照/journal/registry 的重复验证；不扩大 Hook、source、锁或 I/O 预算。

真实 S6 Main 已写出 manifest、注册 `assignment-task-001-build`，尚无 Builder 
交付。其 session 首段自然结束 258.1 秒只做读回，续段被编排 SIGINT 停止，不能
记为 done。Main 实际 manifest 保持原字节，机械注册/prepare 由外部 adapter 沿
原 CLI 接手。prepare 拒绝原 Hook 生成的 `.claude/worktree-reminders/<sessionSHA>.txt`
为 dirty product input。原错误与诊断文件保存在 `reports/r11-s6-dispatch-blocker/`，
未提交、删除、stash 诊断或手改 Runtime。

原 ControlFile 现精确识别 64 位小写 hex 的 `.txt` 及原原子发布 `.delivery-<数字>`，
不忽略整目录。r20 exit 0（3.776 秒），相邻产品文件及原三类 dirty 输入仍拒绝。
after-r13 原 graph/manifest 和 host build 通过，archive SHA
`1e9c64becc78b24589a3f3f756ad9bf3c0397bf174da57b063a410b9d32fe8db`；
下一真实 prepare 使用外部 after-r13 CLI，项目安装/Hook 仍 after-r11。

## r21–r24 与真实 S6 当前失败边界

Resume 所有新生产修改仍在 `sources/resume-after-r11-worktree` 隔离仓。r21 编译失败
来自 private helper 缺失已验证 completionReport 参数；r22 恢复验证器错误要求未发行
packet 已登记；两者均修复并保留失败记录。r23 原“恢复后再次阻塞”用例通过
（70.831 秒整个 Go 命令），并未扩大公开 Hook 或宿主超时预算。

r24 扩大回归 exit 1（558.149 秒）：原 candidate/integration/final/recovery、Main
resolution、Main 读单/S3 题面、控制字节观察器和 S5 revision 回归均通过；恢复后再次
阻塞的公开 Hook 为 9.799 秒，低于原宿主 10 秒上限。唯一失败是恢复后 candidate：
`S9 private candidate capability is not bound to the exact current issued successor packet`。
该失败不是完成记录；正在修复其精确继任绑定，保留原 Writer 与权限边界。

真实 S6 Builder 同 session `864493cd-923d-4181-97d6-edfb1b3b1ce4` 已写出注册范围
内的 go.mod、store.go 和 store_test.go，required checks 按既有 main check-location
配置尚待 Main 实跑。原 Result 拒绝 `.loop-submissions/completion.json`：accepted
PLAN 的第六条断言要求用待接收报告证明“接收已完成”。Main 已完成只读审查，确认此项
是机器接收后置事实而非 TASK Closing Contract 的专业断言；不得放宽 evidence authority、
删除已接收 PLAN 内容或补造回执。原具名修订要求新 owner/team、重新生成专业 PLAN；
旧执行尚未 accepted/merged，不能声称原机制支持原地续接。

同时发现旧 S6 Execution.inputs 漏掉 AssignmentRecord.ReadPaths 声明的 model.md 和
reservation.md。Worker 读到的 checkout 字节与 pinned commit 一致，但不能据此补算
冻结输入闭包通过。修复先在新的源码副本验证，旧执行与输入保持原样。

Builder 最后 Stop 的摘要未保留原 stdin/stop_hook_active，存在既有 loop guard 的
解释空间，尚不能断定 session owner 失配；保留为未定位观察，不虚构根因或验收通过。

### r25 原 Main 解阻预算失败（修复点尚未执行）

r25 原 workspace candidate dirty/stale 负例通过，但 resumed candidate 夹具在前置
Main resolution 的 `git rev-parse --path-format=absolute --git-common-dir` 遇到
`context deadline exceeded`，未执行本轮 candidate 修复点。此路径使用共享公开 Save
6 秒 context；不能仅按错误消息归因于独立 2 秒 source timeout。继续排查同一次
接收中重复 Main/source 校验，不扩展原预算，不把失败当成候选通过。

### R11 S6 原具名修订已接收（rev71）

真实 Main 按原 `register-workgroup --revise-assignment` 登记 Assignment revision 2，
新 owner `builder-task-001-r2`、新 Team `workgroup-REQ-001-build-task-001-r2`。
原 event 为 `evt-s6-assignment-revision-f6ccd4ca0d154f3d1aaf209a9bbea256`。
新 manifest SHA `4adbaba18aec051b26b784c3e24d35ca435b795fb76263e2ece8b8f59ffa9b52`；
旧 manifest/TASK/已接收 PLAN 字节未改；读写范围、required checks 与 baseline 不变。

原 Writer 已将旧 execution 设为 retired、放入现有 execution_history，并删除当前
executions 映射。先前关于旧 key 可能挡住 fresh prepare 的担忧经代码与真实 rev71
双重核对排除，不新增恢复接口。旧代码、checkout、session 均保留历史证据。
下一步等待冻结输入修复包通过，使用原 prepare/launcher 为 revision 2 创建新执行。
报告：`reports/r11-s6-assignment-revision-and-execution-history.md`。

### r26 S6 输入闭包验证边界

source-closure-after-r13-worktree 的 r26 exit 1（109.544 秒）。路径展开、pinned blob、
缺失/越界/链接、checkout 漂移、256 entry/4 MiB 边界，以及原 backend-core 启动/集成/
替换和具名修订退役测试通过。新增完整准备正例失败：目标 fixture 未显式启用 canonical
AssignmentRecord，走了契约限定之外的旧分支。只对该新 fixture 启用契约，并在 prepare
前断言 Record 与两个 ReadPaths 已登记；不放宽生产规则、不改变原 legacy fixtures。
修正后尚待下一串行 Go 命令，当前没有闭包合格或 after-r14 打包结论。

### r27 Resume 原候选至最终回执通过；r28 fixture 仍未闭合

r27 exit 0（149.180 秒），隔离源码包含精确 successor guard、同 verifier 的唯一
当前 slot 校验和删除 Service 层重复 Main 身份校验。保留 Hook 入站、Prepare 与
原 CAS 身份校验；源和 Writer 边界未变。恢复后 candidate 公开 Hook 为 9.274 秒，
原 Main integration 与 final receipt 均通过；旧前驱不恢复 Current、Stop 或 Result。
错误 actor/CWD/session/failed-save、伪造 Resume、重复 slot 与 stale/dirty candidate
负例同时通过。未发行候选兼容新测试尚待随后命令；未声称完整矩阵或真实模型 S9。

r28 exit 1（46.322 秒），新增 canonical fixture 仍缺 TASK 明确 Version，原 Register
在进入准备接线前拒绝。其他边界/旧 backend-core/修订退役测试继续通过。下一轮先
使用完整 canonical 输入跑新增正例，保留两次 fixture 失败，不改生产校验迎合测试。

### after-r14 S6 闭包修复包与 Resume 当前边界

r29 helper 抽取遗漏局部 journal/packets 引用而编译失败，已修正。r30 Runtime 原型、
未发行候选兼容、伪造/重复/过期负例通过，但 resumed blocker/candidate 的公开 Hook
分别为 10.120 / 10.606 秒，超出原 10 秒上限，整轮 exit 1（142.365 秒）。因此
Resume 仍隔离；将同一原边界内连续的 Issue+Current 验证复用同一 request-local
verifier。Prepare 与 CAS 仍分别重开快照复验；历史 standalone Issue 不改。
r33 正在跑扩大原链/新链回归，未提前记为合格。

S6 新 canonical fixture 复用已有完整场景；r31 暴露外部测试 package 误用私有 helper
导致编译失败，改用标准 slices.Contains。新测试按现有受支持选项将相同 required
checks 指定为 Main-owned，避免宿主已知 bwrap probe 限制。r32 exit 0（13.001 秒），
原 register → bind → prepare 的 Record、Execution 输入 SHA、Worker 冻结字节和
pinned Git blob 全部一致。五份代码/测试通过逐文件 baseline 检查后合入，记录在
`reports/after-r14-closure-selective-merge.json`，未覆盖旧执行。

主源码补充简短 PLAN 作者规则、统一 native prepare 入口。after-r14 原 release graph
与 manifest 校验通过，Linux amd64 评估构建 exit 0（12.924 秒），archive SHA：
`bbd9aa82374036061d56eb2f49c0a80b294b8857d44f424e9db237106230eed1`。
它不含隔离中的 Resume。已授权验证代理用包内原 native CLI 准备 revision 2，核完
冻结闭包再走原 dontAsk launcher；R11 Main 安装资产/Hooks 保持 after-r11。

### after-r14 真实 revision 2 prepare 闭包通过

原 prepare exit 0，Runtime rev73，Record revision 2，owner `builder-task-001-r2`，
execution generation 2，状态 ready，check_location=main，base 保持原开发 commit
`1750ae2525534f7bcb83601b7ce0079e2324eb29`。五条声明 ReadPaths 全部进入
53 份 Execution.inputs；53/53 mirror 字节符合冻结 digest。model/state 的 pinned
Git blob、Worker source 与 `.claude/inputs` 三方字节一致，没有补写旧执行。
完整原报告：`reports/r11-s6-after-r14-revision2-prepare.json`。

fresh Builder session 为 `fcde35ad-abed-4eac-957c-bd9d9879f9a9`，原 launcher、dontAsk、
muse、480 秒上限。新执行独立 author；不复制旧代码。Main-owned race/vet 未运行，
不能转写为 Worker 已通过。Root 已将“提交 Result 前可验证”的作者引导落实到
PLAN/Result 题面源码和共用文档，仍待随后 Go/包验证；没有要求 PLAN 描述的未来
实现必须在 PLAN 提交前完成。


## r33–r41 Root 验证、after-r14 安装 smoke 与当前真实 S6（2026-10-15）

### Resume 合入与验证边界

Resume 在 `sources/resume-after-r11-worktree` 的 r33 扩大回归 exit 1；唯一未过项是旧 issuer timeout，其他该轮用例通过。失败记录保留在
`r/reports/global-continuation-s9-expanded-r33.json`。
r34 对该 issuer timeout 修复后，原链/新链定向用例 exit 0（158.537s），报告为
`r/reports/global-continuation-s9-issue-budget-r34.json`。Root 按逐文件核对选择性合入
34 个文件，清单见 `r/reports/resume-selective-merge-after-r34.json`。

合入后的 Root 验证：

- r35：`go test -p=1 ./internal/agentcontext ./internal/cli -count=1 -v` exit 0，
  538.172s；`r/reports/global-continuation-root-cli-full-r35.json`。
- r36：schema、agentcontext、repair、investigation 的选定 S9 / PLAN-Result 用例
  exit 0，505.392s；`r/reports/global-continuation-root-s9-canonical-equality-r36.json`。
  两个 resumed Hook 样本分别为 8.286s、8.543s，低于既有 10s 上限。这里只记录两次
  指定用例观察，不据此声称一般性性能提升。
- r33 的失败没有被删除或改写；r34/r35/r36 的通过范围也不等于完整 S9 或全仓通过。

Resume 已进入 Root 源码，但目前可用评估包仍是 after-r14，未包含该后续合入；
after-r15 尚未构建。本记录不提前引用未来包，也不将选定回归称为真实 S9 阶段通过。

### after-r14 全新安装结构 smoke

对 after-r14 包在新的空目录 `r/runs/after-r14-install-smoke` 执行文档化的
`install --packet`。Install、doctor、`validate --all`、installed release-graph
validate 与 `docs check` 均 exit 0。包 manifest SHA 与安装回执相同
(`264c212c…56a6c7a`)，202 个源条目和 204 个安装回执条目均无摘要不匹配；Linux amd64
native 摘要也与包/回执一致。完整记录：
`r/reports/after-r14-install-smoke.md` 和
`r/reports/after-r14-install-smoke.json`。

这只证明 after-r14 空项目安装与结构验证。初始 Runtime 为 inactive、没有 Agent/TASK/
S7 ReviewAssignment 等实体行，故没有实际验证五态字段；不代表产品阶段、模型、真实
宿主集成、Go 测试或 after-r15 验收。安装生成手册的 Harness 版本为 `dev`，
Runtime/Definition/init schema 是 1.3.0；两者含义不同。

### r37–r41 原 revision recovery / 同快照 execution 坐标回归

r37 `internal/runtime` 与 `internal/workpacket` 选定 repair-result reader/current-selector
用例通过（30.045s）；r38 `internal/hookctx` 与 `internal/workspace` 的 canonical
execution-coordinate resolver、workspace freeze/launch 用例通过（25.308s）；r39
`internal/hookctx` 与 `internal/cli` 的 assignment revision / native context 用例通过
（79.690s）。R39 使用 after-r14 coordinate-hotfix 单 binary，SHA-256 为
`607bacb33096d16808aa46505c8914f396437cd0d4ace449b603dd4ac6c37c94c`。

r40 CLI+integration 矩阵 exit 1（132.140s），三个失败均保留，未从证据中删除：

- `TestCanonicalS6WorkspaceReworkRequiresNamedAssignmentRevision` 的新 CLI fixture 把
  flattened prepare JSON 当作 `workspace.Execution` 解码，导致行为断言未进入目标分支；Root
  已改为直接解码实际 flattened JSON。
- `TestCanonicalMainIntegrationBlocksActualOutOfScopeGitDelta` 和
  `TestCanonicalRevisionReusesOnlyValidatedPredecessorSideParent` 暴露 early-preserve
  条件过宽：所有 `!Ready` inspection 都被提前 preserve，绕过原 out-of-scope
  `ErrScopeViolation`。Root 已收紧到 unmerged 且缺少 `SourceHead`、`TargetHead` 或
  `MergeBase` 坐标时才走该 dirty-inspection preserve 分支；有完整 Git 坐标时继续原 scope
  guard 和 merge-base 检查。

r41 CLI/integration 受影响矩阵 exit 0（128.380s）：CLI 37.362s、integration 35.236s。
`TestCanonicalS6WorkspaceReworkRequiresNamedAssignmentRevision`（7.32s）和
`TestCanonicalDirtyInspectionPreservesCauseAndOffersNamedRevision`（1.71s）通过；原
out-of-scope、predecessor reuse、closing contract、Main check/recovery 用例也通过。报告：
`reports/global-continuation-canonical-revision-recovery-r41.json`。这轮只关闭所列
CLI/integration 回归，不代表全仓、真实模型或 S6 完成；它也没有替代真实 R4 的
Main-owned race/vet 命令。r40 原失败日志继续保留。

### R11 真实 S6 当前边界

真实阶段仍为 S2–S5 **4/9**。R2 的专业错误码与 future-Main 自指断言通过原具名修订
退役；R3 曾有一次真实 429 冷却后同 UUID 成功接收 PLAN/Result，但其 Main integration
未完成。外部 after-r14 coordinate-hotfix binary 沿原 `task-integrate` 重试已 exit 0
（2.944s），实际结果是按 inspection 的 dirty blockers 保存 canonical checkpoint `a3`，
未 merge、未运行 checks；因此不计 S6 完成，也未把原失败伪写成 merge-base 问题。
报告：`reports/r11-s6-r3-original-main-integration-after-coordinate-hotfix.json`。

Main 随后按原 Assignment revision 路径登记 R4：注册 Runtime revision 106，Record SHA
`1145cc4b5a9a993dff54824f92b79681892e50fabc156b97bc3fbf4737da1b83`；原 after-r14 CLI
prepare 成功，53/53 frozen inputs 与 5/5 declared reads 对齐，execution generation 4、
`check_location=main`。Native Builder session
`146738c5-08c5-4e99-91a2-65af6df0cb5e` 的源码 commit
`e82f91dbe87e01274e45cd0bdfa7af2fa2dcd099` 已提交，原 Result 在 Runtime rev115 被接受，
SHA-256 `8e4313bcfcfa7e54c8993bd66b831e0e0d46721046281d5aec1922cbfdccf8c0`。Builder process
在 480.604s 后以 exit 1 结束，stderr 为 `signal: killed`；supervisor 的 `timed_out=false`，
这是原 8 分钟 child limit 后进程结束，不是自然成功退出。Result 已原接收，无需重启或
重报；Root 已开始原 R4 Main `task-integrate` 和串行 checks，最终结果尚未保存。R4 接收
和 Builder 交付不构成 S6 stage credit。

R4 注册、prepare 与 process 证据分别见
`reports/r11-s6-after-r14-revision4-register.json`、
`reports/r11-s6-after-r14-revision4-prepare.json`、
`reports/r11-s6-builder-task-001-r4.process.json`。after-r15 尚未构建；本轮真实阶段数
保持 S2–S5 4/9。

### 后续工作

真实阶段验收仍为 S2–S5 **4/9**。下一步由 Root 对已接收 R4 Result 实跑 Main-owned
race/vet 和原 Main integration，之后再处理 TASK-002；按 after-r15 可用包验证 Resume 与
S9 下游消费者。
新设计的 Impact adapter、dependency/expectation/compatibility 投影以及 observability/
evaluation scripts 仍待设计实施；L3 原有完整 dependency、expectation、兼容性缺口继续
有效。S7–S10 adapters、Main terminal no_transition named-recovery 设计、observability/evaluation scripts、S6–S10 真实阶段链、G08 real-host factory、G01–G08 剩余验收及平台矩阵仍未完成。

## R4 原集成的原生安装资产反例

R4 三个产品文件已提交且原 Result 已接收。原 task-integrate 3.039s 返回 preserved/blocked，
未 merge、未运行 Main checks。Worker 的 60 条 status 包含原 Bootstrap 的三份 tracked
系统 overlay、冻结 inputs 和两个工作区 receipt；Main 也有生成控制资产与登记 worktrees。
原 worktreeClean 只识别少数 runtimeProjection 路径，尚未消费原 Bootstrap 来源证明。
保留 a4 checkpoint、源 commit 与全部输入；不通过提交系统资产、忽略整个 .claude、
手改 Runtime 或强删工作树取得 PASS。下一切片先固定控制资产归属与原 cleanup 契约。
