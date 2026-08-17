# Canonical BUG: BUG-CX-08

> Status: reported
> Severity: P1
> Runtime ref: N/A（模板仓库自审——第二轮复杂度审查；结论经主会话亲证）
> Found in review round: complexity-review-2
> Finding evidence: 本文件 §2/§3 所引 file:line（HEAD=f2c23a4 实测）
> Original responsibility: controller/projection 文案 / AGENTS-template / REQ-template
> Owner: PM / Architect

## 1. Fingerprinted Specification Chain

| Order | Kind | ID | Path | Version | SHA-256 | Relevant clauses |
|:--|:--|:--|:--|:--|:--|:--|
| 1 | BUG | BUG-CX-08 | `docs/reports/bugs/BUG-CX-08.md` | v1 | n/a | all |
| 2 | template | AGENTS | `AGENTS-template.md` | current | n/a | fresh 分支/Control boundaries |
| 3 | template | REQ | `docs/requirements/REQ-template.md` | v2.0.1 | n/a | §C 回显/§E/变更记录 |
| 4 | skill | requirement-funnel | `skills/requirement-funnel/SKILL.md` | v1.1 | n/a | Procedure 0 |

## 2. Observed Contradiction

**症状：BUG-CX-02/05 修复引入或暴露的 S0/S1 信任链文案矛盾，以及一处假安全网机检。**

| Field | Value |
|:--|:--|
| expected | "谁批准、谁执行"在 packet/AGENTS/skill/模板四层口径一致；机检声明与机检行为一致 |
| observed | ① **`<you>` 占位符冲突**（高）：packet 与建议命令把审批人写成 `--approved-by <you>`——读者是 agent，"<you>" 直译即"agent 填自己的名字"，与 AGENTS 白名单 "never infer the approver name from context"（AGENTS-template.md:81）正面冲突（req_discovery.go:146、controller.go:662）。② **AGENTS 同文件自相矛盾**（高）：第 22 行 "Get one REQ human-locked … (REQ locking is human-only)"、第 76 行 "Humans lock REQs / Loop automation cannot lock" vs 第 78-83 行白名单 "the file edit that flips 状态：locked is executed by the main session"——第一次读的 agent 在 22 行处会停下来干等人类编辑文件。③ **§C 回显一致性校验是假安全网**（中，机器项）：模板 §C 回显是表行 `| UI impact（引自顶部） | none / changed / unknown（…） |`（REQ-template.md:107，无冒号分隔），而 `sectionCUIImpact`（engine.go:779-798）只认 "UI impact：value" 冒号形式——按模板填写的 REQ **从不触发**校验；模板注释"不一致会被拒绝"名不副实（f2c23a4 修复的缺口，pin 测试用的是冒号形式故未暴露）。④ **拍板表名/列序错位**（中）：skill Procedure 0（SKILL.md:23）说"写进 REQ 变更记录表（日期 / 人 / 层）"——文末「变更记录」表列是 日期/版本/变更内容/申请人/审批人（无"层"列）；带"层"的表叫「§E 逐层拍板记录」（层/日期/拍板人/方式）。表名与列序双双对不上。⑤ **§E 标题与内容矛盾**（低）：§E 标题「（人-only）」（REQ-template.md:120）但其下拍板表与自审段由 agent 填写。⑥ 变更记录首行示例日期是模板自身历史（低） |
| user/data/system impact | ①是信任链直接漏洞（假人类签核的最短路径）；②在 S0 进口阅读流制造停顿；③使 changed-需求被当 none 放行的真值风险重新敞开（CX-05 的修复目标未达成） |
| reproduction | ①：新 agent 收 freshStart packet → 按 Action 填 `--approved-by <自己的 agent 名>`。③：照模板 §C 表行写 changed、顶部留 none → bind 通过（校验静默不触发） |

## 3. Root Cause Investigation

| Hypothesis | Evidence | Result |
|:--|:--|:--|
| H1: `<you>` 文案写给"终端前的人类"的旧假设残留，packet 化后读者变成 agent 但人称未换 | req_discovery 的命令建议早于 packet 化；freshStartGuidance 复用了同一短语 | confirmed |
| H2: AGENTS 22/76 行是 v4 之前"human-only"的强表述，白名单轮只加了新段没回改旧段 | 22 行括号注释与 76 行 Control boundaries 属不同轮次文案；白名单（78-83）是 f2c23a4 新增 | confirmed |
| H3: §C 假安全网是修复实现与模板格式的假设错配——实现按冒号键值写、模板用表行 | pin 测试用冒号形式自证通过；模板表行实测无冒号 | confirmed |
| H4: 拍板表名错位是 skill 与模板两轮分别落盘、无对照 | skill 写"变更记录表"时 §E 表名为"逐层拍板记录"已定 | confirmed |

Accepted root cause: 与 BUG-CX-INDEX 系统性根因 2（词汇无单一 normative 居所）同族——**手势/审批词汇新增了多居所（packet/AGENTS/skill/模板）但同步纪律仍未建制**；③则是修复自测用例与模板真实格式脱节（测试钉的是实现的自嗨格式，不是用户会写的格式）。

## 4. Closing Contract

### 4.1 Repair scope

- ①：`internal/cli/req_discovery.go:146`、`internal/cli/controller.go:662`（及 grep 全部 `<approved-by <you>`）改为 `--approved-by <the human who locked it>`
- ②：AGENTS-template.md:22 改为 "Get one REQ locked via the human lock gesture (you execute the file flip on their approval — see the lifecycle-verb whitelist below)"；:76 的 "cannot lock" 限定为 "cannot lock without the human's lock gesture / cannot modify the **bound** REQ"
- ③（机器项）：`sectionCUIImpact` 增加表行识别（`| UI impact（引自顶部） | <value> …` 格式：行含 "UI impact" 且为表行时取第二列为值，取值须为三值之一否则忽略）；pin 测试改用模板真实表行格式
- ④：SKILL.md:23 改为 "写进 §E 逐层拍板记录表（层/日期/拍板人/方式）"
- ⑤：§E 标题改「locked 行人-only；拍板行与自审由 agent 依凭证填写」
- ⑥：变更记录示例行标注 {示例}

### 4.2 Forbidden scope

- 不改白名单语义（f2c23a4 已裁）
- ③不改模板格式迁就实现（改实现认模板，模板是对外契约）

### 4.3 Before-fix evidence

§2 所引 file:line（含 §C 表行实测无冒号的复现）。

### 4.4 Retest contract

```text
assert 全仓无 "--approved-by <you>"（grep）
assert AGENTS-template 无互相矛盾的 lock 表述（人工核对 22/76/78-83 三处）
assert pin 测试改用模板表行格式后：表行回显漂移被拒、一致通过
assert skill 拍板落盘指引指向 §E 逐层拍板记录表且列序一致
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
| 新 agent 冷启动至落锁无矛盾/无假签核路径 | 待派 | pending | — |

## 7. Deduplication And History

Canonical BUG: BUG-CX-08（S0/S1 信任链文案与假安全网族；BUG-CX-02/05 修复的残留与缺口）

| Date | Event | Actor | Runtime revision | Evidence |
|:--|:--|:--|:--|:--|
| 2026-08-17 | reported（二轮复杂度审查，主会话亲证全部条目） | 主会话 | n/a | 本文件 |
