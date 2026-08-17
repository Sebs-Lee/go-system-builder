# Canonical BUG: BUG-CX-10

> Status: fixed
> Severity: P2
> Runtime ref: N/A（模板仓库自审——第二轮复杂度审查；结论经主会话亲证）
> Found in review round: complexity-review-2
> Finding evidence: 本文件 §2/§3 所引 file:line（HEAD=f2c23a4 实测）
> Original responsibility: semantic 机检 / 契约模板
> Owner: PM / Architect

## 1. Fingerprinted Specification Chain

| Order | Kind | ID | Path | Version | SHA-256 | Relevant clauses |
|:--|:--|:--|:--|:--|:--|:--|
| 1 | BUG | BUG-CX-10 | `docs/reports/bugs/BUG-CX-10.md` | v1 | n/a | all |
| 2 | code | contracts check | `internal/semantic/contracts.go` | current | n/a | §n 机检 |
| 3 | template | SYNC | `docs/contracts/SYNC-contract-template.md` | current | n/a | 全文 |
| 4 | template | CONTRACTS | `docs/contracts/CONTRACTS-template.md` | current | n/a | 状态行/矩阵 |

## 2. Observed Contradiction

**症状：BUG-CX-04 新机检的精度缺口与模板配套缺位——十位条款查不出、SYNC 条款无声明居所、reviewed 是文档谎言。**

| Field | Value |
|:--|:--|
| expected | §n 对齐机检对任意条款号精确；被要求编号的契约都有声明条款号的列；词表与机器行为一致 |
| observed | ① **`§1` 子串匹配 `§10`**：contracts.go:170 用 `strings.Contains(targetData, "§"+n)`——索引引 §1 而契约只有 §10 时误通过（引错十位条款永远查不出）。② **SYNC 无家可归**：CONTRACTS 索引要求填 `SYNC-{id} §{n}`（CONTRACTS-template.md:52），但 SYNC-contract-template.md 没有"本合同条款"列——它仅有的 §n（:28）指向 FE/BE 的条款号；agent 不知在哪声明 SYNC 条款号，且 Contains 检查被 SYNC 文件里 FE/BE 的 §n 引用**伪满足**（SYNC 条款漂移基本检不出）。③ **`reviewed` 是文档谎言**：四契约模板状态行 `draft / reviewed / locked`（CONTRACTS-template.md:4 等），机器登记只认 `locked`（actions.go:458-462：非 locked 非 cancelled 即 "register refuses partial batches" 硬失败）——"reviewed 是合法中间态"文档与机器不一致。④ **CONTRACTS 模板的 §n 对齐注记漏落地**：BUG-CX-04 §4.1 承诺在 CONTRACTS-template.md:48 加"§n 必须与目标契约正文一致"，f2c23a4 只改了 SKILL step 11（且只点名"需求条款映射 table"，未覆盖 BE:49 的 UI 反推表与 SYNC——SYNC 本无此表）。⑤ protocol #s3 步骤 5/6 与 SKILL step 11 近逐字重复且顺序相反（protocol 先 check 后 flip、SKILL 先 flip 后 check），双源必漂移。⑥ TASK-template:3 的 Status 行是长枚举句，agent 忘改时解析出整句进报错（可读但刺眼） |
| user/data/system impact | ①②使 §n 对齐机检在多条款与 SYNC 场景形同虚设；③诱导 agent 留 reviewed 后被硬拒；④只读模板的 agent 被拦前不知对齐义务 |
| reproduction | ①：索引写 `BE-001 §1`、契约正文只有 §10 → contracts check 绿。②：索引写 `SYNC-001 §3`、SYNC 文件引用 `FE-001 §3` → 绿（伪满足），SYNC 实际无第 3 条 |
| reproduction③ | 契约留 `Status: reviewed` → PTR-PLAN-02 登记 "register refuses partial batches"，与词表承诺矛盾 |

## 3. Root Cause Investigation

| Hypothesis | Evidence | Result |
|:--|:--|:--|
| H1: §n 机检是廉价实现的初版，未考虑多位数与跨文件伪满足 | Contains 子串匹配；SYNC 模板从未进入机检的格式假设 | confirmed |
| H2: SYNC 模板早于"索引条款宇宙"设计（v4.0.2 单一居所轮只改了索引与 BE/FE） | SYNC 模板无条款列而索引有 SYNC 列——两侧设计不同轮 | confirmed |
| H3: reviewed 词表是早期三态设计，机器收敛为二值后词表未回改 | actions.go 只认 locked/cancelled；模板仍三词 | confirmed |
| H4: ④是修复执行遗漏（BUG-CX-04 §4.1 明列 CONTRACTS-template 却未改） | f2c23a4 文件清单无 CONTRACTS-template.md | confirmed |

Accepted root cause: **机检落地时未对"全部被检对象"做格式假设核对**（§n 文法 × 四个模板 × 多位数）；词表与机器各自演化无对账（同 INDEX 根因 2）。④是纯执行遗漏。

## 4. Closing Contract

### 4.1 Repair scope

- ①：contracts.go §n 存在性改数字边界匹配（`§(\d+)` 提取目标契约全部条款号集合，索引 cell 的 n 必须在集合内）；新增十位数与伪满足两个 pin 测试
- ②：SYNC-contract-template.md 增"本合同条款"列（对齐 BE/FE 的需求条款映射表）
- ③：四契约模板状态行注明"机器登记只认 locked（cancelled 出批）；reviewed 仅人审中间态，推进前须翻 locked"（或删 reviewed，owner 裁）
- ④：CONTRACTS-template.md:48 补一句"§n 必须与目标契约正文的『本合同条款』列一致（contracts check 机检）"；SKILL step 11 的"需求条款映射 table"改"任一『本合同条款』列"
- ⑤：protocol #s3 步骤 5/6 压缩为一句 + 指向 SKILL step 11（单一居所）
- ⑥：TASK-template Status 行拆为 `> Status: draft` + 注释行

### 4.2 Forbidden scope

- 不把 §n 对齐升级为全结构解析（保持廉价正则，只修精度）
- 不改索引/条款宇宙的单一居所裁决

### 4.3 Before-fix evidence

§2 所引 file:line + 两条复现（主会话亲证 Contains 子串与 SYNC 模板无条款列）。

### 4.4 Retest contract

```text
assert §1 vs §10 十位数场景：引错报 problem（新增测试 fails_before_and_passes_after）
assert SYNC 条款漂移（伪满足场景）报 problem
assert reviewed 词在四处模板的表述与机器行为一致
assert CONTRACTS-template 含 §n 对齐义务一句
```

## 5. Acceptance And Repair

| Field | Reference |
|:--|:--|
| BUG acceptance evidence | owner 接受（本轮直接修复） |
| repair assignment | same batch |
| Builder activation | same batch |
| repair fingerprint | 本轮修复 commit |
| impact analysis | same batch |
| invalidated evidence | n/a |

## 6. Verification

| Verification | Owner | Result | Evidence |
|:--|:--|:--|:--|
| §n 机检精度测试 + 四模板词表核对 | 待派 | pass | §n 数字集合精确匹配（pin：TestContractsCheckClauseNumberPrecision §1-vs-§10 假阴性修复）；SYNC 增「本合同条款」列（唯一声明居所）；reviewed 词表四处注明机器只认 locked；CONTRACTS 模板补 §n 对齐注记；protocol #s3 与 SKILL 去重单一居所；TASK Status 行拆分 |

## 7. Deduplication And History

Canonical BUG: BUG-CX-10（S3/S4 机检精度与模板配套族；BUG-CX-04 修复的精度缺口与执行遗漏）

| Date | Event | Actor | Runtime revision | Evidence |
|:--|:--|:--|:--|:--|
| 2026-08-17 | reported（二轮复杂度审查，主会话亲证） | 主会话 | n/a | 本文件 |
| 2026-08-17 | fixed+verified | 主会话 | n/a | §n 数字集合精确匹配（pin：TestContractsCheckClauseNumberPrecision §1-vs-§10 假阴性修复）；SYNC 增「本合同条款」列（唯一声明居所）；reviewed 词表四处注明机器只认 lock |
