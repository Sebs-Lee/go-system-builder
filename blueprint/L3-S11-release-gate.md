# L3-S11 — 人工发布闸（Release Gate）

> 层：第三层 ｜ 上游：L2 §S11 ｜ 版本 v3.1.0（v3.0.0 叙事版 + §6 注意力预算；机制事实经调查核实）

## 1. 要实现什么

自动化在此**结构性停机**：提交交接包，等待人从六枚举中显式选一个处置——默认、超时、沉默都不构成批准；发布动作本身永远在人手里。

- 进入时：`awaiting_human_release` + 交接包（clean-round/ACC/审计证据+人须做的事项+"automation stops"声明）。
- 出去时：六处置之一落章（TR-025..030，全部 human_boundary）——approve 只是**记录**授权，merge/publish/deploy 仍由人在 harness 外执行。
- 衡量：**"没人拦"≠"有人批"**——唯一合法的 approve 来源是显式人闸命令。

## 2. 手头有什么（真实机制）

| 机制 | 它能干什么 | 载体 |
|:--|:--|:---|
| 无候选自动迁移（结构性） | controller 对 `awaiting_human_release/release_authorized/aborted` 直接返回空候选——即使 catalog 意外带 auto-trigger 也不会被自动执行；catalog 另列 `automated_s11_decision` 为 strong_block | controller/cycle.go:748-756；loop-definition.json forbidden_events |
| 人闸命令 | `runtime human-decision --disposition <六值> --expected-revision --actor --decision-evidence`——六处置→TR-025..030 **固定 switch、不收目标状态参数**；缺任一必填/未知 disposition → exit 2 零副作用；reject_defect 强制 --finding-evidence；defer 自动补 pause_record | run.go:1065-1133；runtime/s11_migration.go:23-44 |
| 六处置路由 | approve→TR-025 终态（记录授权，"without performing merge…"）；defer→TR-026 paused+S11 检查点；reject_defect→TR-027 进 S8（须发现证据）；reject_acceptance→TR-028 回验收（**带对应旧证据定向失效 action**）；reject_release_audit→TR-029 回审计（同）；abort→TR-030 终态 | loop-definition.json:1717-1793 |
| squash 拦截 | HOOK_SQUASH_MERGE：PreToolUse/Bash 识别 `git merge --squash`/`gh pr merge --squash` → block，恢复=改用普通 merge | hook-policy.json:18；policy/engine.go:236-260 |
| 不可自动发布的其余边界 | catalog forbidden（automated_squash_merge/formal_release）+ INV-015（human_gate）+ skill 程序性约束（"Hook Policy does not block these commands"——hook 只硬拦 squash，deploy/publish 靠 catalog+程序纪律） | loop-definition.json:1969-1984,2057-2060；SKILL.md:54 |
| rollover 前提 | 归档重启仅允许从 `release_authorized`/`aborted`——从 awaiting_human_release rollover 被无副作用拒绝（有测试钉死） | runtime/store.go:589-600；s11_migration_test.go |
| 终止保障联动 | GTR-003（不可逆动作无人工批准→paused）覆盖生产数据/安全/合规动作 | loop-definition.json:1859-1885 |

## 3. 选了什么、为什么（含否决）

| 子问题 | 选用 | 否决与否决理由（减法） |
|:--|:--|:--|
| 怎么防"默认通过" | 结构排除（无自动候选+固定 switch 拒自由文本）——不靠提示词自觉 | 否决"超时提醒+升级"机制——任何超时语义都是变相默认 |
| 发布动作怎么禁 | 三层：hook 硬拦 squash；catalog strong_block+INV-015 human_gate；skill 程序纪律 | 如实记录：hook 只拦 squash——deploy/publish 的禁令是程序性的（SKILL 自认），靠负空间+审计留痕兜底 |
| 驳回怎么回滚 | 六处置固定映射；reject_acceptance/audit **定向失效对应旧证据**（不带污染地重做） | 否决"驳回=整体作废"——定向失效让重做只烧该烧的 |
| defer 语义 | paused+S11 检查点——无超时升级，等人再议 | 否决"defer 计时自动 abort"——时间压力不该替人做价值判断 |
| 授权记录 | TR-025 只记 human authorization，不执行任何发布副作用 | 把"授权"与"执行"永久分离 |

## 4. 怎么编排（时间线讲完一件事）

1. **提交**：交接包落盘（五要素+automation stops 声明）——此后 agent 对发布无任何动作可做（候选为空）。
2. **等待**：人读包（30 秒可抓要害：做了什么/剩什么没决/影响/建议/恢复点）。
3. **处置**：人执行 human-decision 命令——引擎校验（枚举/身份参数/CAS revision/证据齐全）→ 固定 TR 迁移 → 审计留痕。
4. **路由兑现**：approve→授权终态（人在 harness 外执行发布，之后 rollover 开新周期）；reject_defect→S8（附发现，修复后**全新完整轮**再回闸）；reject_acceptance/audit→定向回 S10 重做（旧证据定向失效）；defer→挂起；abort→终态归档。
5. **周期闭环**：终态+人审批证据 → rollover 归档 → 新 inactive runtime → 下一 REQ 从 S0 重新开始。

## 5. 期望效果

走完 S11：

- **发布权威 100% 在人**：授权与执行永久分离；每次闸门决策可审计（人/时间/证据/去向）；
- **结构性防住**：默认批准（无候选+固定枚举）、agent 代批（命令身份参数+forbidden events）、驳回跳轮回闸（TR-027→S8→新轮）、搁置被静默转换（defer 无超时）；
- **交给下一周期**：归档 runtime+journal——完整的可审计历史。

**如实记录**：①deploy/publication/formal release 无 hook 硬拦（程序性约束+catalog 禁令），若要升级为硬拦需扩 hook-policy（目前刻意只留两条 block 规则——防告警疲劳的设计选择）；②六处置的 `--actor` 同样只是字符串（与 S1 approved-by 同级的人身份约定）。

## 6. 注意力预算与渐进披露

总评：**无错配**。S11 是全系统"最小机制、最强控制"的样板：空候选结构性停机（cycle.go:753-756）+ 六处置固定 switch（s11_migration.go:23-40）+ 授权与执行分离——agent 零动作、零阅读，人只读交接包五要素。判定尺见 L3-README「注意力分配原则」。

### 6.1 分配评估（为什么它是对的）

- 没有任何"请 agent 自觉不要发布"的叙述——不发布由**结构**保证（候选为空，想做也没有动作可做）；这是"机制承载规范"的极限形态：规范不存在于任何文档，只存在于结构里；
- 程序性边界（deploy/publish 无 hook 硬拦）是防告警疲劳的刻意取舍（刻意只留两条 block 规则）——正确的克制，不是缺口；
- 其他 stage 的整改方向都应向 S11 看齐：每加一个机制先问"能否用更少的叙述达成同等的结构保证"。

### 6.2 阅读预算（谁在何时读什么）

| 角色 | 必读 | 永不需要读 |
|:--|:--|:--|
| 人 | 交接包五要素（做了什么/唯一未决/影响/建议/恢复点——30 秒可抓要害） | 全部过程证据——汇总已由 S10 完成 |
| agent | 无（迁移候选为空） | 一切——结构保证不需要任何 agent 知晓（它知晓与否不影响结果，这才是真保证） |

### 6.3 整改方向

不动。唯一记录项：`--actor` 字符串身份与 S1 `--approved-by` 同级的人身份约定——威胁模型是"防 agent 越权"而非"防人冒充人"，如实记录即可，不加机制。

## 变更记录

| 日期 | 版本 | 变更 |
|:--|:--|:--|
| 2026-08-14 | v1/v2 | 前两版（被判空洞→叙事不清） |
| 2026-08-14 | v3.0.0 | 叙事版；机制事实经调查核实（无候选结构保证/固定 switch/定向失效回滚/程序性边界如实入档） | owner 复核 |
| 2026-08-15 | v3.1.0 | 新增 §6 注意力预算与渐进披露（分配评估/阅读预算/整改方向=不动），判定尺引 L3-README | owner 指示：渐进披露、机制承载规范、削减平白叙述 |
