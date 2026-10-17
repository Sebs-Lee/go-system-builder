# S3 机制审查（2026-10-16）

## 结论与范围

本次按用户要求回到 S3，检查入口、问卷、专业交付、保存接收、阶段门、登记、
重放、恢复与返工重入。原审查发现的局部问题已返工，**S3 接收、原门/PTR、重放及下一
入口投递已通过本轮验收；全项目检查仍有旧 S5 review manifest 接续缺口，整体优化未完成。**
原偏移发生在失败后的系统接续及把机制工作交给作者的指引。旧 r25 现场保留，不热换
其二进制或手改 Runtime；下述失败过程保留原记录。

## 当前返工进展

Blueprint 已先固定并同步 S3 的协议与 Skill；原实现已修改，专业答卷字段与全局五态
保持原契约。门前投影改为当前已接收交付集；未过门保留原接续，兼容同入口的旧
`no_transition`；发行复用原卷并只冻结上游，新模型闭包在原提交时捕获。
PTR-PLAN-02 按当前批次登记，原出口携带完整 subjects 供下游读取。
准确产物候选/实际路径和模板由 Hook 给出；已接收答案不再被提示重复填写。

- 冻结源码 r115 的 89 个定向顶层用例通过（189.263 秒）；范围含 S3、必要上下游
  接续、旧记录恢复、重放、冷启动模型与其他 REQ 索引共存；不是全仓回归。
- r112 的 workpacket/submission/controller/qualitygate/transition 五包整包回归通过：
  319 个顶层用例、17.915 秒；runtime/semantic/schema/hook/workstate 五包整包回归
  459 个顶层用例通过（88.105 秒）。最后补原事务履职校验后，r114 又完整复跑
  controller/transition/semantic 三包通过（34.965 秒）。定向与整包覆盖有重叠，不能相加。
- PTR-PLAN-02 在问卷模式下必须使用当前原接收记录；直接转移不能跳过履职答案。
  这是原阶段完成条件，不是工具身份/权限拦截。未启用问卷的 legacy 路径继续保留。
- after-r27 Linux amd64 评估包构建成功，未作五平台正式发布；全新安装与原绑定成功。
  真实 muse 完成 S2/S3/S4；S3 仅保存一次，由原 PTR-PLAN-02 推进。两位真实 S5
  reviewer 均交付 `pass`、空 findings，原 TR-003 `evt-tr-003-r38` 进入 building，
  未开展 S6 工作。该轮没有 TR-004；只期待返工的两个 campaign wrapper 因而未通过，
  这表示预期分支未发生，不能改写专业结论，也不能据此证明返工通过。
  after-r26 首次 S3 已通过，但因最终代码补充而停止旧评估，不能计为完整返工验收。
- 返工分支另用新安装实例，机械重放 r27 已交付的 S2/S3/S4 专业答案，原机检继续运行；
  把 r25 真实已接收的 Task finding 专业内容作为明确标注的历史测试输入，由当前 packet
  和消费者生成新来源证据，原 TR-004 `evt-tr-004-r23` 回到 design/e5。后续 Main 的
  S2/S3 由真实 muse 执行。该轮暴露额外提示缺口：未提交 BE 改动时，原消费者报告
  checkout/source-view 差异，通用拒收提示又提到具名修订；Main 误解为本次合同不可改，
  回滚 BE 后以 pass 推进。阶段流转、两次原回执重放及 validate/doctor 虽通过，
  原专业问题未修，故 `reports/r27-s3-historical-finding-result.json` 明确为未通过。
  已补准确的 Git 分支/产物提交指引及未接收答案的原路径纠错提示，复测结果见 after-r28。
  此处不声称本轮 reviewer 重新独立发现问题，
  不复制旧 envelope/subjects，也不把机械前置计为真实模型履职。r25 原现场指纹保持不变。
- r109 的两个失败分别为新增测试同时破坏索引、S3 登记分支误用于 S4 共用动作；
  已修正。r110 编译错误、r111 冷启动测试缺少出口条款/模板也已修正并由 r112 复跑。
  r113 的新增直接跳转负测已通过，旧磁盘登记用例随后明确使用未启用问卷的模式；
  r114 整组通过。所有失败日志保留，不改写为通过。
- r115 另完整复跑 fileview/submission 两包：14 个顶层用例通过（2.277 秒）。
  新用例明确验证未提交专业产物时给出原路径/分支/提交动作，提交后用同一 packet 和答案
  接收并保留专业修复；disk 来源不增加提交要求，原声明的字节校验保持。
- after-r28 已构建并全新安装，使用同一历史 finding fixture。真实 muse 在当前入口修改
  BE-001，提交 `bcc5e457…` 到 `ab-development` 并保留专业修复；CONTRACTS 索引未变、
  BE 已变，复现原故障形状。原 S3 接收后 PTR-PLAN-02 `evt-ptr-plan-02-r34` 进入
  tasks/e7，BE 登记、evidence subjects 和磁盘均为 `e8a5fd955c…`。没有回滚或手刷指纹。
  两次原 PostToolUse 重放返回同一回执，Runtime/journal/答卷字节不变，无虚假接续。
  下一次原 PreToolUse 正常给出 S4 问卷，其 inputs 包含新 BE 指纹；未填写 S4 答卷。

### 仍未通过的全项目检查

after-r28 的 `validate --all` 与 `doctor` 均 exit 1：旧 S5 entry4 的 review manifest
仍引用旧 BE 指纹 `a04fa8…`，当前为 tasks/e7、新指纹 `e8a5fd…`。原 TR-004 只失效被
消费的 review evidence，未退役该清单引用；全局 validator 仍按所有引用清单核对当前磁盘。
这是返工后评审批次/清单接续的剩余缺口；S3 原事务已登记正确事实，原 S4 listener 也能
读取新合同。保留两项失败及 raw campaign 的 `passed=false`，没有刷新旧清单或删历史。
下一阶段需按原评审批次生命周期处理，不能指导作者手工刷新 Runtime/hash 来抹平诊断。
TR-004 事件的静态责任字段与实际 Task finding owner 不一致也单独记录；finding 来源以
原专业答卷及引用的 evidence 为准，本轮未扩改该 S5 事件字段。

本地报告：`reports/s3-rework-r115-final.json`、`reports/s3-rework-r115-receipt-packages.json`、
`reports/s3-rework-r114-final.json`、`reports/s3-rework-r114-shared.json`、
`reports/s3-rework-r112-core-packages.json`、`reports/s3-rework-r112-runtime-and-guidance.json`；
当前包清单为 `reports/after-r28-source-inventory.json`，此前装包模型记录为
`reports/r27-s3-rework-campaign.json`、`reports/r27-s3-rework-spec-continuation.json`；
独立审计为 `reports/r27-s3-rework-independent-audit.json`。历史 finding 仿真前置为
`reports/r27-s3-historical-finding-setup.json`；当前复测为
`reports/r28-s3-historical-finding-campaign.json`。S3 范围结论为
`reports/r28-s3-scoped-acceptance.json`，原始结果为
`reports/r28-s3-historical-finding-result.json`，出口检查为 `reports/r28-s3-exit-listener.json`，
独立审计为 `reports/r28-s3-historical-finding-independent-audit.json`。范围结论明确
`s3_acceptance_passed=true`、`whole_project_validation_passed=false`，不覆盖失败记录。

## 应保留的最小工作流

1. 进入 S3，listener 给出当前问卷的准确 view/answer 路径、冻结输入、已知产物路径与模板。
2. Main 设计共同模型及适用的 FE/BE/SYNC 合同，完成索引与专业自审；Git 来源的产物按 Hook 给出的开发分支提交。
3. Main 只填现有专业答案并保存一次；原消费者运行检查、生成 evidence 并返回原回执。
4. 系统根据本批完整合同事实评估原门；满足时由 PTR-PLAN-02 原事务登记并进入 S4。
5. 未满足时给出具体文件、条目、下一动作或系统接续责任；修复后由原链继续。

工作状态沿用 AGENTS 的 `pending / active / blocked / done / cancelled`。
专业结论、阶段游标与既有接续记录不另造一套作者状态。问卷接收不等于阶段通过。
合同 ID、适用性、条款、兼容性及 oracle 翻译仍需专业判断；原检查、S5 独立审查
与 REQ/breaking 人类决定继续承担各自职责。

## 原审查确认的返工项

| 项目 | 证据及影响 | 最小处理范围 |
| --- | --- | --- |
| 门前合同集合不完整 | `internal/qualitygate/evaluator.go` 的 `evaluatePlanningArtifact` 只要找到任一有效 locked contract，就不补磁盘集合。返工时 CONTRACTS 索引未变、BE 已变，旧 BE 被指纹过滤，新 BE 又未补入，导致新 evidence 无法匹配。原登记 action 必须过门后才运行。 | 在 S3 门前只读取得本批完整当前合同事实，继续核对原答卷、来源及指纹；PTR-PLAN-02 仍负责正式登记。不能要求作者刷新 Runtime/hash。 |
| 未过门却耗尽原接续 | `internal/controller/cycle.go` 在未转换且 gate 非 unknown 时调用 `RecordSubmissionEvaluation`；`internal/runtime/submission_control.go` 将原接续记为 `no_transition`。但原 `internal/controller/planning_followup.go::currentPlanningSubmission` 只接受 pending，Recover 也只扫描 pending。真实 e6 已处于此组合。 | 保留 S3 原接续重试并兼容同入口的旧记录，不要求重交或手改状态。r114 已覆盖旧记录的 listener 重试与 recover；旧 r25 原现场继续只读保留。 |
| 提示没有解释真实缺口 | `contractReceipt` 对已接收答案给 `work_complete=true` 和泛化的“按原门继续”；质量门将来源投影错误丢弃后只留下 `planning_contract_record`。现场已有新 evidence，作者却被提示缺 evidence。 | 原回执与进度提示联合区分本人接收、系统接续和阶段出口；显示准确答卷路径、具体缺失文件/指纹/检查条目及责任人。不要把系统投影失败解释成作者再交一份信封。 |
| 产物路径指引不足 | `internal/workpacket/contracts.go` 的 Read 仅含上游输入；S3 模板要求填 `index_path`、`contract_paths`，没有列明当前可确定的产物路径。 | 在现有 view/提示中列出已有及命名规则可确定的准确路径、实际模板路径。新增合同 ID 由 T1 设计明确，FE/BE/SYNC 适用性由作者判断，不能让系统猜专业范围。 |
| 文档与 Skill 仍增加步骤 | L3 流程图仍写“补新证据信封”；`skills/api-contracts/SKILL.md` 无条件要求 Assignment read-back/approval/activation 与 TASK 起读；specification-planning 和 loop-orchestration 要作者手动执行保存消费者及原门已运行的机检。 | S3 Main 指引统一为冻结输入→专业工作→准确路径保存。Assignment 方法限定实际有 Assignment 的工作；手工检查命令只作诊断选项；证据身份由系统生成。 |

前两项直接影响流转；后三项影响 Agent 是否必须猜机制。已有正常保存重放检查通过，
本次没有把 S5 的重放缺陷未经验证地套到 S3。

## 已固定并落地的返工设计边界

- **冻结时序与 S3 产物**：发行只冻结已确定的上游 REQ、S2 注册设计及其明确引用的
  条件资料，不从本阶段正在创作的 CONTRACTS 索引发现发行闭包。S3 共同模型、结构化
  例子及其依赖属于本次交付；原 shared-model checker 在保存时解析完整依赖并写入现有
  evidence `subject_refs`，之后按原 hash 逐字节验证。作为 S2 上游输入的既有模型不能
  静默修改。不得要求 Main 删除发行、回 S2 重交普通 S3 产物或手动推进游标。
- **REQ 批次边界**：一份答卷对应当前 bound REQ 的一个 CONTRACTS 索引及其完整合同集。
  历史或其他 REQ 的索引不进入当前答卷，也不能作为当前批次 coverage 的证据；原全局
  checks 保留，但 gate 对当前 REQ 另核完整合同事实，其他 REQ 不得帮助当前范围变绿。
- **门未通过时的原接续**：不造状态或协议。保留原接续可重试，并兼容已有 `no_transition`
  记录；系统按原答卷、当前完整交付 facts 和当前门结果重新评估。答卷已接收时不得要求
  作者重交；未接收时才依据具体拒收字段修改现有答卷。
- **准确路径与提示**：现有 S3 view/提示给出能确定的准确上游与产物路径、对应模板、
  answer 路径、具体缺项和下一动作。新合同 ID 与端适用性是 T1/Main 专业设计，不让
  机器猜；不增加专业答案字段。
- **普通作者步骤**：Main 只填一次既有专业答案并保存，消费者和原 gate 执行检查并生成
  evidence，PTR-PLAN-02 正式登记。没有 Assignment、手工信封、重复手工检查或 issue/consume；
  手工检查命令只作失败诊断。S5 独立审查及 REQ/breaking 人闸保留。

这些边界已固定到 [L3-S3](../l3-stages/L3-S3-contracts.md)，实现与本地回归已完成，
真实模型返工与 S3 出口已按上述范围验收；全项目检查缺口另列，下述隔离审计证据保留原失败事实。

## 本次证据

本机评估根目录：
`/home/sebs/.cache/gsb-adapter-work-mp771m5c/claude-closure-20261015T011245Z`。

| 证据 | 结果与限度 |
| --- | --- |
| `reports/r25-s5-prefix-s3.json` | after-r25 首轮真实 Claude 保存原 S3 问卷，PTR-PLAN-02 `evt-ptr-plan-02-r12` 进入 S4。 |
| `reports/r25-s5-muse-rework.process.json` 与同名 native 日志 | muse 恢复后的真实返工：新 e6 答卷接收，未再次通过 S3。进程 exit 0 不作为阶段通过证明。 |
| `runs/r25-s5-retest/.claude/loop-state.json` | revision 51，planning.contracts / entry6；当前 submission 为 `no_transition`。BE 磁盘及新 evidence 一致、Runtime 旧 BE 指纹未更新，CONTRACTS 索引指纹仍匹配。 |
| `reports/s3-audit-r108-source.json` | 冻结当前完整源码的逐文件 SHA 清单。 |
| `reports/s3-audit-r108-probes.patch` | 两个新增探针仅放隔离执行目录，可应用到冻结源码复现；未混入生产实现。 |
| `reports/s3-audit-r108-behavior.json`、同名日志 | 60.047 秒，9 个顶层用例通过、1 个失败，整组 exit 1。原首轮/恢复/拒收及普通重放通过；新增“索引未变、BE 已变、旧登记仍在”用例失败，报告缺 planning_contract_record。 |
| `reports/s3-audit-r108-contracts.json`、同名日志 | 30.519 秒，workpacket/runtime/qualitygate/controller 的 25 个合同与规划定向顶层用例通过，exit 0；不包含上述失败探针，不代表 S3 完整验收。 |

## 局部返工的闭环顺序

1. 先统一 S3 Blueprint 的正常流程、未过门接续、冻结时序、输入变化重入和本批范围；删旧手工信封指引。
2. 修原门的完整事实投影与原接续，给具体诊断；同步问卷、Skill 和协议，普通作者不增加步骤。
3. 回归首次交付、错误答案修正、来源变化、保存重放、中断恢复、同代返工重入和多索引边界。
4. 新冻结包全新安装；当本轮独立审查实际给出双 PASS 时保留该结论。返工分支使用明确标注的
   历史 finding fixture，通过当前原消费者/TR-004，后续真实 muse 走 S2→S3→PTR-PLAN-02，
   到 S4 出口停止。不热换 r25 实例，不把测试前置冒充本轮真实独立 finding。

本轮 S3 的设计、实现、本地回归与上述宿主范围验收记为 done；旧 S5 manifest 的全项目
校验失败继续保留为接续工作，不能把本阶段通过写成全局优化完成。
