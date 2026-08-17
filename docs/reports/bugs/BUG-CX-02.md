# Canonical BUG: BUG-CX-02

> Status: reported
> Severity: P1
> Runtime ref: N/A（模板仓库自审——S0/S1 agent 视角复杂度审查）
> Found in review round: complexity-review-1
> Finding evidence: 本文件 §2/§3 所引 file:line（HEAD=9cd52fa 实测）
> Original responsibility: requirement-funnel skill / REQ-template / AGENTS-template
> Owner: PM / Architect

## 1. Fingerprinted Specification Chain

| Order | Kind | ID | Path | Version | SHA-256 | Relevant clauses |
|:--|:--|:--|:--|:--|:--|:--|
| 1 | BUG | BUG-CX-02 | `docs/reports/bugs/BUG-CX-02.md` | v1 | n/a | all |
| 2 | skill | requirement-funnel | `skills/requirement-funnel/SKILL.md` | v1.0 | n/a | Exit Conditions |
| 3 | template | REQ | `docs/requirements/REQ-template.md` | v2.0 | n/a | §A/§E 头部 |
| 4 | design | L3-S0 | `blueprint/L3-S0-requirement-design.md` | v4.0.3 | n/a | 漏斗/拍板 |

## 2. Observed Contradiction

**症状：S0 漏斗的每一处"人类拍板"都没有可执行的手势定义，agent 不知道谁在何时写下 `状态：locked`，也不知道每层"approved"凭什么算数。**

| Field | Value |
|:--|:--|
| expected | 漏斗的每一层停顿点（§A/§B/§C 拍板、最终落锁）都有明文的"人做什么、agent 据此做什么"手势；compaction 后可凭持久痕迹证明拍板发生过 |
| observed | ① 落锁：REQ 全程 agent 写文件；到落锁时刻 requirement-funnel SKILL.md:39 只说「human confirms … and signs the §E lock record; `status: locked`」——把顶部 `状态：` 从 draft 改成 locked 这一步**谁做**无规定。自己改违反 "REQ locking is human-only"（AGENTS-template.md:22,76）字面；请人改则人不知道改哪。`req bind`（run.go:257-261）只校验文件写着 locked，无法区分谁写。② 层拍板：SKILL.md:24-27「have the human confirm」「Never design the next layer before the current one is approved」——"approved" 无持久化形式（§A-§C 无逐层确认字段），对话里说"可以"算不算、compaction 后如何证明 §B 曾被拍板，均无答案；Stop Conditions（:41-44）不覆盖"正常等待人类拍板的姿态"，agent 可能永久停住或越层推进。③ 人闸动词白名单：agent 可代执行哪些人闸命令口径不一——agent-protocol.md:191 允许代跑 bind；AGENTS-template.md:209 读起来像 agent 去跑 pause；loop-orchestration SKILL.md:418 说 Driver 不碰 bound REQ；而 pause/resume/amend/unbind 只靠 `--approved-by <string>`（lifecycle_commands.go），agent 完全能自己填一个人名（与 v4.5.1/v4.6.1 已如实入档的信任边界同源，但**文档层连"意图"都没说清） |
| user/data/system impact | S0 是系统的进口；落锁手势缺失是最可能卡住真实使用的单点。层拍板无痕 → compaction 后"你怎么就设计 §C 了"无据可依；人闸白名单缺失 → agent 要么越权要么过度请示 |
| reproduction | 按 requirement-funnel 走到 Exit Conditions：尝试确定谁执行 draft→locked 翻转 → 文档无答案 |

## 3. Root Cause Investigation

| Hypothesis | Evidence | Result |
|:--|:--|:--|
| H1: 漏斗设计以"人类在对话中持续在场"为隐含前提，从未定义"人在对话里、文件由 agent 写"这一实际交互形态的手势映射 | L3-S0 v3.5.0 定义 A1 为"agent 整理稿交人确认"，但整理→确认→落盘三步中"确认如何变成文件状态"没有任何一轮设计覆盖；REQ-template §E 只有终局签字位 | confirmed |
| H2: 拍板痕迹被认为由对话历史承载 | 对话不是权威（L1-D1 推论：未经持久化的决策等于没发生）——skill 设计违反了自己蓝图的第一条决策 | confirmed |
| H3: 人闸白名单缺失是 v4 六轮 lifecycle 改造的累积漂移 | bind（P0 放开代跑）/pause/resume（P1 封装）/unbind/amend（P2/P3）各轮分别陈述权限意图，从未汇成一张"谁可执行什么"的总表 | confirmed |

Accepted root cause: **人机交互手势（human gesture）从未被当作设计对象**——漏斗层拍板、落锁、人闸动词执行权三处都停留在"描述意图"（human-only / approved），没有落到"具体动作序列 + 持久痕迹"。这是 L1-D1 在交互层的投影缺口：D1 只外置了机器状态，没外置"人的决定"。

## 4. Closing Contract

### 4.1 Repair scope

- `skills/requirement-funnel/SKILL.md`：Exit Conditions 明文落锁手势——「人类在对话中明确说"锁定"即构成拍板；agent 据此将 `状态：` 置为 locked 并在 §E 记录操作人与时间；`req bind --approved-by <同一人>` 为二次确认」；Procedure 每层补一行「层拍板 = 人类在对话中的明确肯定，agent 记入 REQ 变更记录（日期/人/层）」；Stop Conditions 补「等待人类拍板是正常姿态：明确告知人需要确认什么，然后停下」
- `docs/requirements/REQ-template.md`：§E 增"逐层拍板记录"小表（层/日期/人/方式）；头部状态行注释补「draft→locked 由 agent 依据人类对话拍板执行」
- `AGENTS-template.md` Control boundaries：人闸动词白名单表——agent 可代执行：`req bind`（口头授权链）；必须人类亲手（或逐字提供完整命令行）：`runtime pause/resume`、`req amend/unbind`、`runtime rollover/human-decision`
- `docs/agent-protocol.md` #s1 与白名单对齐

### 4.2 Forbidden scope

- 不引入新的机器强制（如 locked 字段的 actor 签名校验——仓库层无法区分，v4.5.1 已裁决属威胁模型外）
- 不改 REQ-template 的层结构（§A-§F 划分 v3.6.0 已定）

### 4.3 Before-fix evidence

本文件 §2 所引 file:line（SKILL.md:24-27,39-44 / AGENTS-template.md:22,76,209 / REQ-template.md:4 / run.go:257-261）。

### 4.4 Retest contract

```text
assert requirement-funnel Exit Conditions 含"谁把 状态： 置为 locked"的明文手势
assert REQ-template 存在逐层拍板记录位（机器不校验，供 compaction 后审计）
assert AGENTS-template 存在人闸动词白名单表，且与 agent-protocol #s1 无矛盾
assert 三份文档对 "locking is human-only" 的表述一致（意图=决策权在人，执行=agent 依拍板落盘）
```

## 5. Acceptance And Repair

| Field | Reference |
|:--|:--|
| BUG acceptance evidence | pending（owner 拍板手势方案） |
| repair assignment | pending |
| Builder activation | pending |
| repair fingerprint | pending |
| impact analysis | pending |
| invalidated evidence | n/a |

## 6. Verification

| Verification | Owner | Result | Evidence |
|:--|:--|:--|:--|
| 以新会话走漏斗至落锁，无手势歧义 | 待派 | pending | — |

## 7. Deduplication And History

Canonical BUG: BUG-CX-02（人工手势协议缺失族；含 v4.5.1/v4.6.1 已入档信任边界的文档侧补全）

| Date | Event | Actor | Runtime revision | Evidence |
|:--|:--|:--|:--|:--|
| 2026-08-17 | reported（复杂度审查 B1/B2/C3） | 主会话+sub-agent | n/a | 本文件 |
