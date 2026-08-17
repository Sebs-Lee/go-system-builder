# Canonical BUG: BUG-CX-05

> Status: reported
> Severity: P2
> Runtime ref: N/A（模板仓库自审——S0/S1 agent 视角复杂度审查）
> Found in review round: complexity-review-1
> Finding evidence: 本文件 §2/§3 所引 file:line（HEAD=9cd52fa 实测）
> Original responsibility: REQ-template / scenario 模板 / cross-matrix 模板
> Owner: PM / Architect

## 1. Fingerprinted Specification Chain

| Order | Kind | ID | Path | Version | SHA-256 | Relevant clauses |
|:--|:--|:--|:--|:--|:--|:--|
| 1 | BUG | BUG-CX-05 | `docs/reports/bugs/BUG-CX-05.md` | v1 | n/a | all |
| 2 | template | REQ | `docs/requirements/REQ-template.md` | v2.0 | n/a | 头部锚点/§AC/§A4/§B/§F |
| 3 | template | cross-matrix | `docs/design/prototypes/cross-matrix-template.json` | v1 | n/a | 占位符 |

## 2. Observed Contradiction

**症状：S0 模板存在前向引用谜语、静默不一致与语义残句——第一次填表的 agent 需要猜。**

| Field | Value |
|:--|:--|
| expected | REQ-template 的每个字段在填写时就地可懂；不一致的声明在 bind 时被机器抓住 |
| observed | ① **AC 指向列前向引用 5 个未定义机制**（REQ-template.md:77：AC↔CASE 桥/场景包/rule.source_refs/经背书的 N/A/scenario bridge——命令名还不完整，实际是 `scenario validate`；解释在 blueprint/L3-S2 与 S2 skill，均不在 S0 阅读单）且无完整填写示例行。② **UI impact 双声明位不一致静默**：顶部 blockquote 是唯一解析位（模板:9-11），§C 回显"只回映"（:104）——agent 改 §C 不改顶部，bind 读旧值无报错（engine.go:734-752 只拒非法三值，不拒两处不一致）；模板注释还向需求文档读者泄漏 Go 函数名 `parseUIImpact`。③ **§B"挑战每个 must"是残句**（:49「说不出打破的代价 = 偏好，移去下方假设与风险表」无主语无对象，完整规则在 SKILL.md:25）——两处措辞不同。④ **"负空间"词汇漂移**：§A4 叫"明确不做"，L1/L3-S2 叫"负空间"，AC 指向列又写 "§A4-条目"（自造指针语法无示例）。⑤ **§F 40 行骨架对 S0 是噪音**（:125-165，模板自称 self-describing 但内嵌 S3 机制与版本号）。⑥ cross-matrix 模板占位符 `"<一句话：…>"`（中文短句）与 ≥8 字符含字母的机器下限冲突（详见 BUG-CX-03⑥，此处只记模板侧） |
| user/data/system impact | S0 是漏斗第一层；谜语字段直接拉长来回（agent 填错→机检红→翻 S2 文档→回填）；UI impact 静默不一致可把 changed 需求当 none 放行（真值风险） |
| reproduction | ②：复制 REQ-template → 改 §C 回显为 changed、顶部留 none → `req bind` 通过且 ui_impact=none → PTR-PLAN-01 不做设计包直进——错误静默放大 |

## 3. Root Cause Investigation

| Hypothesis | Evidence | Result |
|:--|:--|:--|
| H1: REQ-template v2.0 的减法轮（删 PM Todo/checkbox/§11/§14）只做删除，新增机器读口（指向列）时把 S2/S3 机制术语直接搬进模板注释，未配 S0 视角的示例与锚点 | L3-S0 v3.6.0 变更记录："被删内容新居映射全量入档"——删除有纪律，新增字段无同等纪律 | confirmed |
| H2: 双声明位是历史遗留（顶部锚点为机器、§C 为人），一致性校验从未被要求 | engine.go 只解析顶部；无任何 issue/REQ 提过两处一致 | confirmed |
| H3: 残句是 skill 重写后模板未同步清理 | SKILL.md:25 有完整规则且措辞不同 | confirmed |

Accepted root cause: **模板的新增字段缺少"第一次填写者测试"**——减法轮建立了"删除必有新居映射"的纪律，加法方向（新机器读口）没有对称纪律（就地示例+机制锚点+一致性校验）。UI impact 静默不一致属于 D2 缺口：控制点（bind）没有覆盖全部声明位。

## 4. Closing Contract

### 4.1 Repair scope

- `docs/requirements/REQ-template.md`：指向列补 2-3 个完整示例行（好/坏各一）+ 一行机制锚点（「详见 #s2 与 skills/scenario-model-design；S0 只需保证 FR- 指向真实存在的 FR 行」）；顶部 UI impact 注释删函数名改「顶部字段是机器锚点」；§B 残句补主语（「agent 对每条人给的'必须'追问打破代价；答不出则降级为偏好」）；§A4 标题补别名（「明确不做（负空间/negative space；AC N/A 背书的合法指向）」）；§F 收敛为一句指路（「S3 起由 CONTRACTS 索引维护，S0 不填」，骨架移 CONTRACTS-template 或保留但加"S0 跳过"首行）
- `internal/cli/run.go`（或 engine parseUIImpact 调用侧）：bind/amend preflight 校验顶部与 §C 回映一致，不一致报错指路（附测试）
- `docs/design/prototypes/cross-matrix-template.json`：占位符改含下限提示的双语示例（与 BUG-CX-03 联动）

### 4.2 Forbidden scope

- 不删双声明位（顶部为机器锚点已定，§C 为人读回显——加校验而非删字段）
- 不把 S2 机制解释全文搬进 REQ-template（公理五要"锚点"不要"重复居所"）

### 4.3 Before-fix evidence

本文件 §2 所引 file:line + 复现②的步骤。

### 4.4 Retest contract

```text
assert REQ-template 指向列有完整示例行与 S0 可懂的锚点
assert bind 对顶部/§C UI impact 不一致报错（新增测试 fails_before_and_passes_after）
assert 模板内无 Go 函数名
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
| 新会话仅依 REQ-template 填完 §AC 指向列无回头翻阅 | 待派 | pending | — |

## 7. Deduplication And History

Canonical BUG: BUG-CX-05（S0 模板前向引用与静默不一致族）

| Date | Event | Actor | Runtime revision | Evidence |
|:--|:--|:--|:--|:--|
| 2026-08-17 | reported（复杂度审查 B3-B7） | 主会话+sub-agent | n/a | 本文件 |
