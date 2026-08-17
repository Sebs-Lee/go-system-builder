# Canonical BUG: BUG-CX-09

> Status: reported
> Severity: P2
> Runtime ref: N/A（模板仓库自审——第二轮复杂度审查；结论经主会话亲证）
> Found in review round: complexity-review-2
> Finding evidence: 本文件 §2/§3 所引 file:line（HEAD=f2c23a4 实测）
> Original responsibility: scenario 模板 / 两 skill / protocol #s2 / ADR-template
> Owner: PM / Architect

## 1. Fingerprinted Specification Chain

| Order | Kind | ID | Path | Version | SHA-256 | Relevant clauses |
|:--|:--|:--|:--|:--|:--|:--|
| 1 | BUG | BUG-CX-09 | `docs/reports/bugs/BUG-CX-09.md` | v1 | n/a | all |
| 2 | template | scenario-model | `docs/design/prototypes/scenario-model-template.json` | current | n/a | risk 字段 |
| 3 | skill | scenario-model-design | `skills/scenario-model-design/SKILL.md` | v1.3 | n/a | 包清单/字段表 |
| 4 | template | ADR | `docs/design/decisions/ADR-template.md` | current | n/a | 全文 |

## 2. Observed Contradiction

**症状：BUG-CX-03 修复后的 S2 剩余口径债——模板示例违背自家指导、包清单三处口径不一、"ADR package" 无定义。**

| Field | Value |
|:--|:--|
| expected | 模板示例与 skill 指导一致；模块包文件清单各处一致；被要求写入的产物（ADR 包）有格式与居所定义 |
| observed | ① **模板 risk 示例就是指导明令禁止的写法**：scenario-model-template.json:25 `"risk": "rule-dense"`——把 coverage_profile 枚举名抄进 risk，恰是 scenario-model-design SKILL:82-85 "do not copy profile names into it" 要防的错误；引擎不查，静默存活到 S5。② **模块包清单三处口径矛盾**：scenario-model-design SKILL:35-46 的包清单（8 槽位）**无 cross-matrix.json**，同文件 step 3 又说它是人工输入三件之一；rules/scenario-model.md §2 目录树与 §3"四个文件"同样缺它、§7 却称"模块包第八文件"；protocol #s2:215 说 "eight-file package" 实列 9 类。③ **"ADR package" 全库无定义**：specification-planning step 8 要求结论写在 "ADR package" 的 `## Depth Self-Review` 标题下，protocol #s2:219 说 N/A 清单 "joins the same sign-off package"——但 ADR-template.md 无 Depth Self-Review 段、无 Endorsed N/A 段（grep 亲证 0 命中）；"包"是单份 ADR 还是目录级聚合、放哪、谁消费（人闸拍板时人看哪里）均无答案。④ **步骤号断档**：specification-planning step 9 之后直接 11-14，无 step 10（旧 step 10 跳过条件移为 Step 0 后编号未重排；Step 0 的 "go to step 11" 依赖这个断档编号）。⑤ recovery N/A 伴随字段（recovery_source_refs + recovery_reason，rules §3.1:85-87 与模板 :80-84 有）未进 SKILL 的 oracle 字段表。⑥ SKILL:100 oracle 表 recovery 行与后续段落粘连在一行（markdown 渲染会把整段吞进单元格） |
| user/data/system impact | ①固化 risk/profile 混淆；②建包清单误导（幸有 fail-closed 兜底）；③自审与 N/A 清单两个 S2 关键产物落盘位置自由发挥、人闸无处看；④对照阅读者困惑 |
| reproduction | ①：新 agent 抄模板起笔 scenario-model → risk 字段原样 "rule-dense"。③：close 时找 "ADR package" 该写哪 → 无处可查 |

## 3. Root Cause Investigation

| Hypothesis | Evidence | Result |
|:--|:--|:--|
| H1: 模板是 v1 期产物，后续 skill 指导轮（risk 正交化是 f2c23a4）没回改模板示例 | 模板 risk 值早于指导存在；f2c23a4 只改了 SKILL | confirmed |
| H2: 包清单是 cross-matrix 引入（第八文件轮）时只改了部分居所 | SKILL 清单、rules §2/§3、protocol 各轮分别改，无对账 | confirmed |
| H3: "ADR package" 是设计讨论中的口头聚合概念，从未落成结构 | ADR-template 无对应段；两个消费点（自审结论/N/A 清单）都指向这个悬空概念 | confirmed |
| H4: 步骤断档是 Step 0 前置重构时删旧 step 10 未重排 | 旧 step 10（ui_impact 跳过）内容已成 Step 0，编号未回收 | confirmed |

Accepted root cause: 与 INDEX 系统性根因 1/2 同族——**指导层多处居所各自演化无对账**；模板作为"可抄的最强引导"（D4）反而缺一条"模板示例必须与指导同审"的纪律。ADR package 则是概念被两个消费点引用却从未被定义（公理三违例：产物无固定消费者阅读位）。

## 4. Closing Contract

### 4.1 Repair scope

- ①：scenario-model-template.json risk 示例改自由短语（如 `"money is involved"`）
- ②：三处清单统一（SKILL 包清单 + rules §2 目录树/§3 表加 cross-matrix.json；protocol/文档口径统一为"九文件包"或注明计数口径）
- ③：ADR-template.md 增 `## Depth Self-Review` 与 `## Endorsed N/A` 两个固定段；protocol #s2 一句话定义 sign-off package = 本 REQ 的 `docs/design/decisions/ADR-<id>.md`（含上述两段）
- ④：specification-planning S3/S4 段重排为 10-13，Step 0 的跳转号同步；blueprint L3-S2 §6.3 的旧编号引用同步
- ⑤⑥：oracle 表 recovery 行补伴随字段半句；粘连处换行

### 4.2 Forbidden scope

- 不给 risk 加机器枚举（f2c23a4 已裁自由短语方向）
- 不把 ADR 包做成新机检（先定义居所；机器化待实战数据）

### 4.3 Before-fix evidence

§2 所引 file:line（模板 risk 值/三处清单/ADR-template grep 0 命中/步骤断档均亲证）。

### 4.4 Retest contract

```text
assert scenario-model-template risk 值不是 coverage_profile 枚举名
assert SKILL/rules/protocol 三处包清单文件集合一致
assert ADR-template 含 Depth Self-Review 与 Endorsed N/A 段
assert specification-planning 步骤号连续（0-13）
```

## 5. Acceptance And Repair

| Field | Reference |
|:--|:--|
| BUG acceptance evidence | pending |
| repair assignment | pending |
| Builder activation | pending |
| repair fingerprint | pending |
| impact analysis | pending |
| invalidated evidence | n/a |

## 6. Verification

| Verification | Owner | Result | Evidence |
|:--|:--|:--|:--|
| 新会话只读 protocol+skill+模板走通 S2 零考古 | 待派 | pending | — |

## 7. Deduplication And History

Canonical BUG: BUG-CX-09（S2 口径与产物居所族；BUG-CX-03 修复残留）

| Date | Event | Actor | Runtime revision | Evidence |
|:--|:--|:--|:--|:--|
| 2026-08-17 | reported（二轮复杂度审查，主会话亲证） | 主会话 | n/a | 本文件 |
