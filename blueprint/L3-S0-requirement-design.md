# L3-S0 — 需求设计（Requirement Design）

> 层：第三层 ｜ 上游：L2 §S0 ｜ 版本 v3.0.0（叙事版；机制事实经调查核实，含 file:line）

## 1. 要实现什么

把人的意图固化成一份**人锁定的需求基线**——它是全链指纹的根：锁定后不可变，后续一切设计/契约/任务/证据都挂在它的指纹上。

- 进入时：人的意图陈述（对话），非结构化。
- 出去时：`docs/requirements/REQ-{id}.md` 状态=locked，含锁定记录（日期/操作人/版本）与可计算的 SHA-256——即 S1 `req bind` 的直接输入。
- 衡量：**锁定即冻结**——之后任何人（含 agent）改动它都必须走人的修订流程，且旧版本保持不可变（换 generation，不覆盖）。

## 2. 手头有什么（真实机制）

| 机制 | 它能干什么 | 载体 |
|:--|:--|:--|
| REQ 模板 | 17 节结构承载逼问：§7 优先级 Must/Should/Could/Won't、§9 流程与异常、§10 迁移范围清单（带 4 项自查）、§12 待澄清问题、§14 完整性自检 8 项（含设计公理五问）、§6 术语表带"禁止叫法"列 | `docs/requirements/REQ-template.md` |
| 模板自描述 | 模板直接告诉作者"顶部 UI impact 是唯一被 parseUIImpact 解析的位置"——文档即接口契约 | REQ-template.md:11,101 |
| harness 绑定校验 | 状态=locked、版本非空、文件名 REQ- 前缀、SHA-256 复核、UI impact 三值解析 | `req bind`（run.go:177-232；engine.go:462-541） |
| UI impact 三值机制 | none/changed/unknown；unknown 可锁定但触发 `ui_impact_resolved` guard，规划不放行（"先干着再说"无路） | engine.go:524-541；guards.go:289-304 |
| 锁定写入拦截 | 锁定产物写入→硬阻断，恢复提示指向新 generation 目录 `docs/req/versions/{REQ-ID}/g{N+1}/`——旧版不可变 | HOOK_LOCKED_ARTIFACT_WRITE（hook-policy.json:17；policy/engine.go:264-312） |
| 变更控制规则 | "Chat never changes a baseline"；REQ 目标/范围/优先级/验收的变更暂停整个 loop；锁定与修订 human-only | `docs/rules/change-control.md`（R-CHANGE-01） |
| 协议契约 | S0 的 done_when（locked+锁定记录+SHA-256 可算）与 human_gateway（req_amendment） | `docs/agent-protocol.md` #s0 |

注意两个**如实记录的事实**：①S0 无 primary skill——协议明确它是 human-driven、主会话仅辅助（agent-protocol.md:182）；user-story/flow/scenario 等 skill 全部在 S2 消费 REQ，不在 S0。②`validate --all` 对 requirements 目录**没有**内容检查（semantic/validator.go:69-165）——REQ 的质量把关在模板自检+人，机器校验集中在 bind 时。

## 3. 选了什么、为什么（含否决）

| 子问题 | 选用 | 否决与否决理由（减法） |
|:--|:--|:--|
| 怎么逼出完整需求 | 模板字段（§7 优先级/§9 异常/§10 迁移/§12 未知项）——填写即思考 | 否决"为 S0 配专职 skill"——协议定位 human-driven；模板逼问已够，加 skill 只加阅读量 |
| 怎么管 UI 影响 | 顶部三值字段 + `parseUIImpact` 机械解析 + unknown 触发 guard | 否决"让 agent 自行判断何时补 UI 分析"——三值+guard 是零成本的硬地板 |
| 怎么防 agent 改需求 | human-only 锁定（bind 必须 `--approved-by`）+ 锁定写入拦截 + generation 不可变 | 否决"目录级静态保护"——hook-policy 的 protected_paths 不含 requirements/（事实），运行时锁定清单才是对的粒度：锁前可自由草拟、锁后才冻结 |
| 怎么管锁定后修订 | change-control 规则（暂停 loop + 人批 + 新 generation） | 否决"就地编辑+版本号"——会破坏下游指纹链；换 generation 旧版永久可审计 |
| 怎么保证 REQ 质量 | §14 八项自检（含公理五问）+ 人 review | 否决"validate --all 扩展 REQ 检查"——bind 时已校验机器可核项；内容质量属人判断，机检收益低 |

## 4. 怎么编排（时间线讲完一件事）

1. **草拟**：主会话按模板逐节写；填不出的节生成澄清问题交人（不臆造）——§12 待澄清问题就是这些问题的正式居所，open 状态显式可见。
2. **逼问自检**：§10 迁移自查 4 项 → §11 UI 门禁 9 项 → §14 完整性 8 项逐项过；每项都是模板预埋的问题，不是事后审查。
3. **UI 影响判定**：顶部三值选一（unknown 允许——意图未明不必硬猜，但 S2 的 guard 会拦住直到 §11 澄清）。
4. **人锁定**：人 review → 修改 → 在 §15 签锁定记录（日期/操作人/版本）。到此文件转为锁定产物，进入 documents 清单。
5. **冻结生效**：此后任何写尝试触发 hook 硬阻断，恢复提示指向新 generation 目录；修订=暂停 loop（change-control）→ 人批 → g{N+1} 重锁。
6. **交接**：S1 `req bind` 复核 locked/版本/指纹 → 入册。§16 覆盖矩阵骨架（REQ→BR→CASE→…→证据）此时埋下，是全链追溯的起点。

## 5. 期望效果

走完 S0：

- **意图有了不可变的锚**：锁定 REQ 是全链指纹的根（bind 后 generation+1，"never auto-promotable"）；
- **结构性防住**：agent 自创/自改需求（human-only 锁定）、锁定后漂移（写入拦截+generation 不可变）、"先实现后补文档"（change-control 禁令）、UI 影响含糊进设计（unknown guard）；
- **未知是安全的**：unknown 不会混过去，也不会阻塞发现——它显式登记在 §12，逼人在 S2 前澄清；
- **交给 S1**：locked 文件+锁定记录+SHA-256——绑定命令的三个直接输入。

**如实记录的已知不一致**（供第四层修复清单）：①unknown 澄清位置两处文案不一致（engine.go 注释说 §12，guard 报错说 §11）；②agent-protocol.md:173 的 S0 actions 只列两值未含 unknown；③锁定拦截特判的 kind 字符串 "requirement" 与实际写入的 "req" 不一致（engine.go:274 vs :508），导致 REQ 的 human_required 计算偏差。

## 变更记录

| 日期 | 版本 | 变更 |
|:--|:--|:--|
| 2026-08-14 | v1/v2 | 前两版（被判空洞→叙事不清） |
| 2026-08-14 | v3.0.0 | 叙事版；机制事实经 sub-agent 调查核实（含三处代码/文档不一致如实入档） | owner 复核 |
