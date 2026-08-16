# 需求：REQ-{id}

> 名称：{需求名称}
> 状态：draft / locked / changed / archived
> Owner：{产品/用户}
> PM / Architect：{姓名}
> 版本：v0.1.0
> 创建日期：YYYY-MM-DD
> UI impact：none / changed / unknown

> `UI impact` 是 `状态：locked` 的强制顶部字段，由 `loop-harness req bind` 的 `parseUIImpact`（`internal/transition/engine.go`）读取。值必须三选一：`none` / `changed` / `unknown`。`unknown` 会触发规划暂停门禁，需在 §D 澄清后才能推进 S2。

<!-- 给 agent（固定阅读，2 行）：
人给预期效果，你全权接管设计，人只逐层拍板；按 §A→§B→§C 漏斗推进，上一层拍板前不设计下一层。
提案纪律与自审见 requirement-funnel skill——上交唯一合格形态是完整方案，开放问题禁止上交。 -->

## §A 理念（Why）

> A1 不是人的原话：你接收并理解人的表述后，过滤口语歧义、补全隐含前提，整理成结构化复述，交人确认"这就是我想说的"。原话会歧义；对话不入册，整理稿才是权威。

| 字段 | 内容 |
|:--|:--|
| A1 预期效果（agent 整理稿，人确认） | {做什么、为什么值得做——两三句} |
| A2 背景与痛点 | {问题和上下文}{现在没有它时用户怎么凑合} |
| A3 成功长什么样 | {完成后用户可感知的变化——验收的精神锚} |
| A4 明确不做 | {范围外事项——与 A1 同等重要；每条一句不做的原因} |

用户和利益相关者（决策权"高"的诉求冲突 = 拍板点候选）：

| 角色 | 诉求 | 决策权 |
|:--|:--|:--|
| {角色} | {诉求} | 高 / 中 / 低 |

术语（分歧当场定名；统一叫法供全部下游文档使用）：

| 术语 | 含义 | 禁止叫法 |
|:--|:--|:--|
| {术语} | {含义} | {容易混淆的叫法} |

## §B 方向与约束（How 的边界——需求级选型，架构归 S2）

方向取舍（每个被否方向必须有一句否决理由——ADR 雏形，S2 承接）：

| 方向 | 一句话方案 | 采纳/否决 | 理由 |
|:--|:--|:--|:--|
| {方向 1（推荐标 ★）} | | 采纳 | {为什么} |
| {方向 2} | | 否决 | {一句否决理由} |

硬约束（说不出打破的代价 = 偏好，移去下方假设与风险表）：

| 硬约束 | 为什么是必须 |
|:--|:--|
| {约束} | {代价/理由} |

关键假设与风险：

| 假设 / 风险 | 影响 | 缓解措施 / 验证方式 |
|:--|:--|:--|
| {假设或风险} | 高 / 中 / 低 | {缓解或验证} |

## §C 具体需求（What——逐条可追溯）

> 正向范围 = 本节需求点全集（FR + NFR）；§A4 只登记负向。

需求点（"服务于"指不回任何 §A 条目 = 范围蔓延，当场质疑）：

| 编号 | 模块 / 流程 | 需求 | 服务于 §A_ | 优先级 |
|:--|:--|:--|:--|:--|
| FR-001 | {模块} | {需求} | {A1/A2/…} | Must / Should / Could / Won't |

流程和异常：

| 流程 | 输入 | 输出 | 状态 | 异常 | 权限 | 验收标准 |
|:--|:--|:--|:--|:--|:--|:--|
| {流程} | {输入} | {输出} | {状态} | {异常} | {谁能做/谁被拒} | {证据} |

验收标准（换个人来判，结果一样；**指向列是 AC↔CASE 桥的机器读口**——`FR-{id}` 须经场景包 rule.source_refs 达到 CASE，`NFR-{id}`/`§A4-条目` 是经背书的 N/A，自由文本会被 `scenario bridge` 拒绝）：

| 编号 | 标准 | 指向 |
|:--|:--|:--|
| AC-001 | {可判定的验收表述，含衡量指标} | FR-{id} / NFR-{id} / §A4-{条目} |

非功能需求：

| 编号 | 类型 | 要求 | 验收标准 |
|:--|:--|:--|:--|
| NFR-001 | 性能 | {要求} | {指标} |
| NFR-002 | 安全 | {要求} | {证据} |
| NFR-003 | 可靠性 | {要求} | {证据} |
| NFR-004 | 合规 | {要求} | {依据} |

数据迁移范围清单（轻量模式、迁移类、重构类需求即使跳过独立设计文档也必须填写本区；本区是 task 拆分边界依据）：

| 路径 / 文件 | 类型 | 处理动作 | 所属任务 | 引用 / 依赖 | 说明 |
|:--|:--|:--|:--|:--|:--|
| `{path}` | source / config / script / generated / test | migrate / keep / delete / review | TASK-{id} | `{import/script}` | {说明} |

填写纪律：删除项在"说明"列注引用清理标准；每个文件映射到唯一 task 或明确标注共享边界；任务拆分前检查重复 import、脚本引用、生成物入口和别名路径。

UI 影响（任何涉及前端页面、组件、交互、可见状态、错误展示、权限展示或响应式布局变化的需求必须填写本区；`UI impact = changed` 时必须先完成 UI Design Package Gate 才能锁定 FE/BE/SYNC 合同）：

| 字段 | 内容 |
|:--|:--|
| UI impact（引自顶部） | none / changed / unknown（顶部 blockquote 是唯一被 `parseUIImpact` 解析的位置，本节只回显，不独立声明） |
| 影响页面 / 模块 | {页面、路由、组件或 N/A} |
| 模块当前真相包 | `docs/design/prototypes/<module>/`（index.html + stories.md + flows.md + scenario-model.json + cross-matrix.json + cases.json + scenario-coverage.json + fixture-contract.json + *.html）/ N/A |
| 原型门禁状态 | N/A / pending / ready |

> 真相包齐备性、S/F-NNN 完整性与 branch 覆盖率由 S2 的 `loop-harness scenario validate` 校验，不在本节自查。

## §D 待澄清问题

| 编号 | 问题 | 影响范围 | 阻塞范围? | 状态 | 结论 |
|:--|:--|:--|:--|:--|:--|
| Q-001 | {问题} | 范围 / 流程 / 状态 / API / 权限 / 验收 | 是 / 否 | open / closed | {结论} |

## §E 锁定记录（人-only）

| 日期 | 操作 | 操作人 | 说明 |
|:--|:--|:--|:--|
| YYYY-MM-DD | locked | {PM / Architect} | 用户确认范围；UI impact={none/changed/unknown} 已随顶部字段固化 |

{agent 自审结论——提议锁定前全审的输出，一段话：审过什么、修了什么、留了什么旁注}

## §F 派生文档与覆盖矩阵（骨架——**活矩阵唯一居所在 CONTRACTS-{id} 索引**，本表仅锁定时快照；REQ 是锁定基线，agent 不填写本表——L3-S3 v4.0.1）

### 派生 UI 设计包

| 模块原型集 | 路径 | 状态 | 覆盖需求 |
|:--|:--|:--|:--|
| <module> | `docs/design/prototypes/<module>/` | N/A / pending / ready | FR-{id} |

### 派生合同

| 合同 | 路径 | 状态 | 覆盖需求 |
|:--|:--|:--|:--|
| CONTRACTS-{id} | `docs/contracts/CONTRACTS-{id}.md` | draft / reviewed / locked | 当前 REQ 全量索引 |
| FE-{id} | `docs/contracts/FE-{id}.md` | draft / reviewed / locked | FR-{id} |

### 条款覆盖

| REQ 条款 / source_ref | Rule → CASE → Story → PATH → Spec | 合同条款 | TASK | 验证证据 | 状态 |
|:--|:--|:--|:--|:--|:--|
| REQ-{id}/FR-{id} | BR-{id} → CASE-{id} → S-{id} → F-{id} → PATH-{id} → `web/e2e/{module}/*.spec.ts` | FE/BE/SYNC-{id} §{n} | TASK-{id} | REV/QA/E2E/BUG/ACC-{id} | planned / covered / verified |

### 模块当前真相映射

REQ 只能作为来源，不能创建需求私有设计或测试副本。任何触及模块的 REQ 都必须维护模块全集并触发 full-module regression。

| 模块 | Rule / CASE | Story | PATH | Spec | Round evidence |
|:--|:--|:--|:--|:--|:--|
| {module} | `scenario-model.json` / `cases.json` | `stories.md` / S-{id} | `flows.md` / F-{id} / PATH-{id} | `web/e2e/{module}/` | REV/QA/E2E round {n} |

### 批次下游索引

| 关系 | 文档 | 路径 | 状态 |
|:--|:--|:--|:--|
| downstream | 任务看板 | `docs/tasks/index.md` | draft / executing / completed |
| evidence | Review round manifest | `{team-manifest-path}` | pending / complete |
| evidence | Delivery verification | `docs/reports/review/REV-{id}.md` | pending / PASS |
| evidence | QA evidence | `docs/reports/qa/QA-{id}.md` | pending / PASS |
| evidence | E2E Tester evidence | `docs/reports/e2e/E2E-{id}.md` | pending / PASS |
| evidence | Clean round | `{review-evidence-path}` | pending / valid |
| evidence | 验收 | `docs/reports/acceptance/ACC-{id}.md` | pending / passed |
| downstream | 发布审计 | `docs/release_audits/{file}.md` | pending / approved |

## 变更记录

| 日期 | 版本 | 变更内容 | 申请人 | 审批人 |
|:--|:--|:--|:--|:--|
| 2026-08-15 | v2.0.0 | 重构为漏斗结构（§A 理念→§B 方向与约束→§C 具体需求→§D 澄清→§E 锁定→§F 覆盖矩阵）；删除 PM Todo、21 项自查 checklist 与设计公理五问（实质分别转入字段/S2 引擎/blueprint）；状态枚举精简为 draft/locked/changed/archived；过程方法论由 requirement-funnel skill 承载；流程表补权限列、非功能表补 NFR 编号列 | owner | owner |
