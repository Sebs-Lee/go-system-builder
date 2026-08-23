# Blueprint — 分层设计蓝图（项目定调文档）

> 本文件夹是整个模板项目的**分层设计文档永久居所**：自顶向下、逐层推导，每一层可对照上层验证。安装时随模板分发，放到 `.claude/bin/` 旁边，作为设计理念与规划的双向查询备用。

## 层结构与文件索引

| 层 | 文件 | 内容 | 状态 |
|:--|:--|:--|:--|
| 第一层 | `L1-design-principles.md` | 工程哲学与设计蓝图：场景、物理约束 C1-C5、使命与六项把控、主干设计决策 D1-D7、五公理、失效模式目录、演化协议、权威层结构定义 | ✅ v2.4.0 |
| 第二层 | `L2-lifecycle-plan.md` | 生命周期实战目标：三条铁律 + 能量函数、S0-S11 每阶段（任务/把控/风险对应/理论根据/入口/出口/失败路由）、全局规则、层间校验 | ✅ v1.4.4 |
| 第三层 | `L3-README.md` + 12 份 `L3-Sn-<stage名>.md`（S0-requirement-design … S11-release-gate） | 各 Stage 详细落地设计：角色/产物/过程/门禁判定/工具承载（T1-T7）/失败处置/反作弊/度量 | ✅ |
| 第四层 | `L4-agent-dispatch-governance.md`（首份；后续按机制域扩展） | 横跨多个 L3 的工具机制与治理设计：统一对象模型、状态、消息、Hook、Harness、恢复和验收契约 | ✅ v0.3.0 |
| 第五层 | `L5-*.md`（待建） | 实现规格与实现：Schema、Skill、Agent Definition、模板、代码、测试和迁移任务 | ⏳ |
| 第六层 | `L6-*.md`（待建） | 实战运营记录与回灌（失效数据 → 第一层演化协议） | ⏳ |

## 阅读纪律

1. **从上往下读**：先 L1 后 L2；任何下层困惑先回上层找根据。
2. **映射纪律**（L1 §六）：下层每条设计必须能指认它承载的 D1-D7 与通过的公理；指认不出 = 越层设计，退回。
3. **修订走 L1 第五部分演化协议**：任何层的修订须引用实战证据、过准入五问、带版本与 changelog。

## 安装位置

目标项目应用本模板时，把本文件夹整体复制到 Manual 旁边，供 agent 与人随时查证设计意图：

```bash
cp -R blueprint <target-project>/.claude/bin/blueprint
```

## 变更记录

| 日期 | 变更 |
|:--|:--|
| 2026-08-14 | 建立 `blueprint/`；L1/L2 自 `docs/rules/` 迁入并按层级命名；接入发布白名单与安装流程 | owner 指示：层级文档永久保留、安装时置于 `.claude/bin/loop-harness` 旁备用 |
| 2026-08-20 | 建立首份 L4 `L4-agent-dispatch-governance.md`；L4 定位调整为横跨多个 L3 的工具机制与治理层，具体实现下沉 L5 | owner 指示：统一梳理 Sub-agent / Agent Team 调度与治理 |
| 2026-08-20 | L4 终审收敛为机制准入规则、单一事实链与可验证平台控制；Claims/Assignment/Result 为事实，coverage/board/ledger 为视图；普通计划回执连续执行，高风险才审批 | 删除重复控制面，避免把未验证的 Hook 唤醒能力写成既成事实 |
| 2026-08-22 | L4 P0 平台接线落地（官方 payload 字段、TeammateIdle/SubagentStop exit 2、TaskUpdate self-claim 门）；S7 侧落地 FindingSupplement、运营指标子集、`s7 manifest-draft`、oneOf 错误剪枝、SessionStart/PreCompact S7 恢复投影与兼容别名审计 | S7 §13 / L4 §15 各补 2026-08-22 审计记录；真实平台 doctor 与 Controller 假唤醒收敛仍待后续批次 |
| 2026-08-22 | 批次二：Controller idle 语义收敛（假唤醒删除、首写屏障入 wire 路径）、blocked_by_confirmed_finding 投影、site_lost BLOCKER 复用、overlap/cold-start overload validator、resource lock 排队、worktree 共享控制面核验、capture exec 自动采集、supplement 并入主 loop-state + discriminator 判别门 | S7 §13 审计记录；仍待：真实平台 doctor、产品侧浏览器 wrapper、regression_available 指纹复用校验 |
| 2026-08-22 | 五视角 E2E 代入测试（Planner/Reviewer/cold-start/对抗/恢复+复杂度）：对抗测试 12 类攻击 11 类 D3 合规零误伤；净复杂度判定持平（操作更简、词汇更重、无镀金）；修复 wire 路径 workspace 投影 P0 缺陷、`s7 workspace-digest` 落地、agent 状态错误文案诚实化 | S7 §13 审计记录；高优先未修：next/ready 矛盾、reading→working 断点、oneOf 剪枝生产路径休眠、恢复包三缺陷、S7 前置状态无引导 |
| 2026-08-22 | 高优先缺口根因修复批次：TR-008/009 文案对齐自动提交语义、恢复包 drain/来源/噪音三修复、剪枝推广到嵌套判别节点、plan_checkpoint 自动激活链（register-workgroup 预生成信封 + PostToolUse 链式推进 + `runtime agent-begin` 兜底）、s7 draft 前置门与 sandbox recipe | S7 §13 审计记录；剩余缺口收敛为环境/产品侧与低优先 UX 项 |
| 2026-08-23 | 验证轮（盲测五视角）确认第一轮修复生效，且发现两处回归：R1 资产未 go:embed、R2 链中途失败后不可恢复——均已修；R3 文案残留、R4 恢复包矛盾、R5 ready 非确定性、R9 文档/schema 矛盾——均已修。R6 CLI 路径解析、R7 controller 抢停 stop-idle 门、R8 错误指引弱、R10 capture exec 终端泄漏——待修 | S7 §13 审计记录；剩余缺口以 CLI/UX 为主，机制层已稳定 |
| 2026-08-23 | 验证轮后续修复批次：R6 `--root` 路径一致（register-workgroup + fingerprint + reconcile-policy-ref 同类 bug 同步）、R7 stop-idle 门不再被 controller 抢拍（lifecycle 各阶段回归）、R8 三处错误信息 next-action 加强（stale revision / verdict=fail / supplement 缺字段）、R10 capture exec 终端秘密不泄漏 | S7 §13 审计记录；剩余收敛为命令面拆分与产品侧 wrapper |
