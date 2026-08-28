# S7~S10 根因调查与修复方案（RCA + Repair Plan）

> **上游**：`docs/audits/S7-S10-2026-08-27.md`（108 条发现，108/108 保留，9 项为正面确认）  
> **方法**：L3-S8 调查范式 — `ObservationBatch → InvestigationCase(可逆聚类) → CausalModel → RepairContract`；`L1 D5` 三阶（guide→structure→force）逐项定级  
> **产出**：12 个 `InvestigationCase`（根因簇）→ 12 份 `RepairContract` → 12 个 `Task`（`#4`–`#15`），覆盖 99 条负向缺口，9 条正面项标注`保留`不修  
> **版本**：2026-08-27 | 分支 `codex/github-actions`

---

## 0. 聚类总览

### 0.1 12 簇 Mapping（99 负向 → 12 RC，9 正面保留）

| RC | 根因标题 | 发现数 | 关键 `file:line` | Severity 峰值 | Task | D5 定级 |
|---|---|---:|---|---|---|---|
| RC-01 | 身份绑定缺失 — Targeted 独立性仅字符串不等 | 4 | `repair/evidence.go:81` | Critical | `#4` | force |
| RC-02 | 阻塞语义窄化 — `isBlockingBug == P0` | 5 | `verification/clean_round.go:259` | Critical | `#5` | structure→force |
| RC-03 | 冻结基线双层旁路 — 分类器白名单 + 手写名单 | 4 | `policy/engine.go:781` + `review/workspace.go:141` | Critical | `#6` | structure |
| RC-04 | 写屏障休眠与阶段裸窗口 | 6 | `policy/engine.go:337` + `cli/run.go:2526` | Critical | `#7` | force |
| RC-05 | 覆盖率分母自指 — S10 自证循环 | 8 | `acceptance/manifest.go:452` | High(实 Critical) | `#8` | force + structure |
| RC-06 | 守卫存根与幽灵 Transition 泛滥 | 14 | `transition/guards.go:114` + `transition/catalog.go:327` | High | `#9` | 逐项 force/structure/删除 |
| RC-07 | S7 计划与独立性硬度不足 | 9 | `review/plan.go:442` + `review/submit.go:1519` | Medium | `#10` | structure |
| RC-08 | S8 调查材料完整性占位 | 9 | `investigation/case_workflow.go:676` | High | `#11` | structure |
| RC-09 | S9 修复链过度与漂移（含先红后绿、双轨、幽灵字段） | 11 | `repair/evidence.go:81` + `repair/model.go:239` + `impact/analysis.go:126` | Critical | `#12` | structure + force |
| RC-10 | 巨事务与指纹/度量冗余 | 6 | `review/submit.go:67` + `runtime/store.go:2597` | Critical | `#13` | 精简（过度） |
| RC-11 | 协议与文档面过度与漂移 | 11 | `schema/agent-message.schema.json:0` + `L3-S7:742` | High | `#14` | 精简 + 对齐 |
| RC-12 | 流畅度回路与口径摩擦 | 18 | `cli/action_catalog.go:55` + `docs/agent-protocol.md:527` | High | `#15` | guide + structure |
| — | 正面确认（保留不修） | 9 | `runtime/store.go:0` 等 9 项 | Low | — | 保留 |

> 计数口径：同一 `file:line` 被多簇引用时按主簇计 1 次，去重后 99 负向全覆盖；9 正面项（`Store CAS`/`CleanRound` 七检查/`Hook` 接线等）见 §6。

### 0.2 依赖与执行波次

```
Wave 1 (P0, 无依赖, 可并行):  RC-01  RC-02  RC-04  RC-05
Wave 2 (依赖 Wave1 语义):      RC-03 ─┐
                               RC-06  ├─→ RC-07/08/09 (S7/S8/S9 硬度)
Wave 3 (精简/文档, 最后):      RC-10  RC-11  RC-12
```

`RC-03` 与 `RC-04` 同根（写面），建议同人/同 PR；`RC-06` 守卫收敛是 `RC-02/05` 的前提校验；`RC-10~12` 为非阻塞性精简，任何 Wave1 阻塞时可提前启动。

---

## 1. RC-01 — 身份绑定缺失（Targeted 独立性仅字符串不等）

**包含发现**：`S9-1` (`repair/evidence.go:81`) / `EH-10` (`repair/evidence.go:125` Critical) / `EH-11` (`repair/runtime.go:600`) / `S9-2` 先红后绿部分（`repair/plan_report.go:52`，与 RC-09 共担）

**CausalModel**

- **Trigger**：S9 `ready_for_full_review` 要求 `targeted_reverification` 独立于原实现者（`L1 D6` 三方收敛）。
- **Invariant 失效**：`CommitTargetedReverification` 仅校验 `OriginalAssignmentID != PerformingAssignmentID` 字符串不等，未绑定 `assignment_owners` / `agent` 身份 / `verifier` 池。
- **Faulty Mechanism**：`internal/repair/evidence.go:81-82,125-126` 自由提交字段比较；CLI `--actor` 省略时静默取 `qa` (`runtime.go:597-599`)。
- **Propagation**：Builder 填任意假 ID → `Create` 通过（两串不等）→ `Commit` 通过（无身份核对）→ `ready_for_full_review` 放行自背书 → 下一轮 S7 基于伪复验结论。
- **Symptoms**：`S9-1`/`EH-10` 自验闭环；`EH-11` 失败方向 `failure_class` 同理可伪。
- **Blast Radius**：单 BUG 的修复有效性判定；若该 BUG 为阻塞性，直接污染 CleanRound。
- **Detection Gap**：无 `assignment_owners` 交叉、无 `verifier` 池白名单、无 `produced_by` 核验；`evidenceBackedGuard` 存根无法补位。

**RepairContract — RC-01**

- **Scope**：`internal/repair/evidence.go` / `internal/repair/runtime.go` / `internal/cli/repair_command.go`
- **Contract**：
  1. `CommitTargetedReverification` 新增校验：`performing ∈ 已派发 verifier 池 && performing != repair owner`（与 `assignment_owners` 交叉），`Original` 必须等于该 BUG 的实际 `repair assignment`。
  2. 去除 `qa` 静默默认值，无 `--actor` 时显式报错。
  3. `failure_class` 同步加身份根校验（至少 `evidence_ref` 非空且 `kind` 匹配）。
  4. 双负向用例：原实现者填假 ID 自验应拒；`failure_class` 自报无证据应拒。
- **Task**：`#4` — **RC-01 身份绑定缺失**（`repair/evidence.go:81`，force，成本中）

---

## 2. RC-02 — 阻塞语义窄化（`isBlockingBug == P0`）

**包含发现**：`S7-1` (`clean_round.go:259` Critical) / `EH-1/EH-2` (`clean_round.go:259` / `review/submit.go:158`) / `S10-10` (`manifest.go:358` P0 风险可作非阻断) / `S10-8` 联动

**CausalModel**

- **Trigger**：`L3-S7 §10.1` 承诺"业务定义为 blocking 的 BUG 即阻断"，不以 `severity` 代替语义。
- **Invariant 失效**：`isBlockingBug` 窄化为 `severity == "P0"`，与 `blocking` 语义脱钩。
- **Faulty Mechanism**：`internal/verification/clean_round.go:259-262` 硬编码；`BUG` schema 无 `blocking` 字段；`manifest.go:358` 允许 `risk.Severity=P0` 作非阻断风险直送 S11。
- **Propagation**：P1/P2 业务阻塞 BUG 被标 `severity=P1` → `no_open_blocking_bugs` 通过 → `EvaluateCleanRound` 7/7 → `TR-009` 放行带病 S10；S10 再以 `risk` 降级二次放行。
- **Blast Radius**：整轮 `CleanRound` 可信度；S10/S11 阻断分级一致性。
- **Detection Gap**：无 `energy function` 语义深度复核；`risk` 与 `blocking` 分级无交叉闸。

**RepairContract — RC-02**

- **Scope**：`internal/verification/clean_round.go` / `internal/acceptance/manifest.go` / BUG schema / `docs/agent-protocol.md`
- **Contract**：
  1. BUG schema 新增 `blocking: bool`（显式业务阻断标记），`SubmitResult` 校验 `blocking` 必填（`P0` 隐式 `blocking=true` 向后兼容）。
  2. `isBlockingBug` 改为 `severity=="P0" || blocking`。
  3. `manifest.go:358` 强制：`risk.Severity==P0` 必须进 `blocking` 集合，否则拒。
  4. 负向用例：P1 `blocking=true` BUG 应阻 clean；`P0` 风险作非阻断应拒；现有 P0 用例回归。
- **Task**：`#5` — **RC-02 阻塞语义窄化**（`clean_round.go:259`，structure→force，成本小）

---

## 3. RC-03 — 冻结基线双层旁路

**包含发现**：`S7-2` (`policy/engine.go:781` + `review/workspace.go:141` Critical) / `S9-9` 同根 (`policy/engine.go:781`) / `EH-8` 控制面排除 (`repair/artifact.go:190`)

**CausalModel**

- **Trigger**：`verification` 阶段产品冻结（`L1 D1` 外化权威 + `L4-runtime-control-plane`）。
- **Invariant 失效**：两层防御同时可被手写绕过。
- **Faulty Mechanism**：
  - `bashMutationPaths` 白名单式小分类器，仅命中 `redirect/sed -i/cp/mv/rm/go generate/git apply` 等，漏 `python/node/npm/make/cargo/heredoc/inline script`。
  - `verifyFrozenSubjects` 仅重哈希 `ReviewPlan.frozen_subjects` 手写名单，声明外文件无感知。
  - `session baseline` 排除 `.claude/.git`，控制面写入不计入 diff。
- **Propagation**：`python3 gen_fixtures.py` 改产品文件未被识别为 `mutating` → 直接放行 → 且不在 `frozen_subjects` → `submit` 全绿 → 验证被篡改后代码仍判 clean。
- **Detection Gap**：无 `git diff` 基线交叉；`frozen_subjects` 无"声明集 ∪ 实际变更"双源比对。

**RepairContract — RC-03**

- **Scope**：`internal/policy/engine.go` / `internal/review/workspace.go` / `internal/repair/artifact.go`
- **Contract**：
  1. 扩展 `bashMutationPaths` 至解释器家族（`python/node/npm/make/cargo/go run/heredoc`）+ `fail-closed` 默认 `mutating`（未知 Bash 即视为可能写入）。
  2. `verifyFrozenSubjects` 改双源：`声明集 ∪ git diff 基线`，声明外新增文件亦触发 `DRIFT`。
  3. `session baseline` 排除列表收窄或显式声明控制面写入（`.claude` 产物需显式 `scope_ref`）。
- **Task**：`#6` — **RC-03 冻结基线双层旁路**（`policy/engine.go:781`，structure，成本中，与 RC-04 同 PR 建议）

---

## 4. RC-04 — 写屏障休眠与阶段裸窗口

**包含发现**：`S7-3` (`cli/run.go:2526` + `policy/engine.go:568`) / `S8-1` (`policy/engine.go:337` Critical) / `S10-4` (`policy/engine.go:654`) / `S9-5` (`policy/engine.go:424`) / `EH-8` 缝隙

**CausalModel**

- **Trigger**：`L4-runtime-control-plane` 阶段级写冻结（验证期/调查期/审计期均应冻结，仅 `repair fixing` 例外）。
- **Invariant 失效**：屏障作用域过窄。
- **Faulty Mechanism**：
  - S7 `first-write` 依赖 `Runtime.Agent==nil` 且 `controller` 非 `allow` 时跳过（休眠）。
  - S8 `bug_resolution.investigation` / S9 `targeted_reverification` / S10 `acceptance` 全程无 `write deny`。
- **Propagation**：Investigator 在 S8 直接 `Edit` 产品代码无拦截，调查与修复边界消失；S10 审计期可写产品；targeted 复验人可篡改产品。
- **Blast Radius**：S7/S8/S10 数据面完整性；S9 复验可信度。

**RepairContract — RC-04**

- **Scope**：`internal/policy/engine.go` / `internal/cli/run.go`
- **Contract**：
  1. 统一写屏障为 phase 级：`verification` + `bug_resolution.*` + `acceptance`/`release_audit` 均 `deny` 产品写，仅 `repair fixing` 例外且受 `RepairAllowedWritePaths` 约束。
  2. `first-write` 改按 `assignment_id` 生效，`PreToolUse` 全路径评估不依赖 `Runtime.Agent`。
- **Task**：`#7` — **RC-04 写屏障休眠与阶段裸窗口**（`policy/engine.go:337`，force，成本小）

---

## 5. RC-05 — 覆盖率分母自指（S10 自证循环）

**包含发现**：`S10-5` (`manifest.go:452`) / `S10-6` (`transition/actions.go:639`) / `S10-7` (`manifest.go:246`) / `S10-8` (`qualitygate/evaluator.go:378`) / `S10-9` (`manifest.go:353`) / `S10-13` (`manifest.go:293`) / `EH-12/EH-13` / `C-13`

**CausalModel**

- **Trigger**：S10 `coverage_inventory` 需与外部基准（REQ/合同/变更台账）交叉。
- **Invariant 失效**：分母由同一方制造，无 `exact-set` 锚。
- **Faulty Mechanism**：`manifest` 的 `requirement/contract/changed_path` 行数由 Agent 自定；`changed_path` 未与 `AffectedPaths` 对账；`responsibility_matrix` 自由字符串无独立性校验；`counterevidence` 仅需 `problem statement`；`GATE-RELEASE-AUDIT-BLOCKED` 不复检 `applyS10ManifestGate`。
- **Propagation**：12 条 AC 只冻结 1 行仍得 100% 覆盖率放行 S11；同一 Builder 填全 8 区 `pass` 引用自产 evidence；空 `AffectedPaths` 静默绕过 `TR-016` 失效。
- **Detection Gap**：无 `BuildCoverageInventoryForRoot` 重跑比对；无 `AffectedPaths` 空值拒；无 `owner != builder` 校验。

**RepairContract — RC-05**

- **Scope**：`internal/acceptance/manifest.go` / `internal/transition/actions.go` / `internal/qualitygate/evaluator.go`
- **Contract**：
  1. 验收时重跑 `BuildCoverageInventoryForRoot` 与 `manifest` 行集合 `exact-set` 比对，缺行即拒。
  2. `changed_path` 与 `AffectedPaths`/变更台账对账，`TR-016` `AffectedPaths` 为空时拒或要求显式 `all`。
  3. `responsibility_matrix` 结构化 + `owner != builder` 独立性校验。
  4. `counterevidence` 需 `evidence_ref` 非空；`TrackingRef` 需 `URL|issue_id` 格式校验。
  5. `GATE-RELEASE-AUDIT-BLOCKED` 复检 `applyS10ManifestGate`。
- **Task**：`#8` — **RC-05 覆盖率分母自指**（`manifest.go:452`，force+structure，成本中）

---

## 6. RC-06 — 守卫存根与幽灵 Transition 泛滥

**包含发现**：`S7-4` / `S10-2` / `S10-3` / `S10-14` / `S8-8` / `S8-9` / `EH-15` / `F-1~F-5` / `C-12` / `S9-8`（共 14 项）

**CausalModel**

- **Trigger**：`L4-state-transition-core` 治理面"文档有闸、运行有闸"一致性。
- **Faulty Mechanism**：`guardRegistry` 20+ `evidenceBackedGuard` 存根（仅 `len(evidence)==0`）；`forbidden_events`/`protected_commands`/`warn_and_retry`/合成 `TransitionID` 声明零消费；幽灵阶段/旧 `BUG draft` 死代码残留。
- **Propagation**：无 `clean_round` 仍可创建 `ACC`（`forbidden_events` 空转）；`git push origin main` 在 `release_authorized` 直达发布（`protected_commands` 死代码）；`warn_and_retry` 12 处声明误导调用方以为有二次窗口；`GuardSpec` 指纹承诺与 `evidence 非空` 实现不符。
- **Detection Gap**：无 `loop-definition` 声明与 `engine` 消费一致性校验。

**RepairContract — RC-06**

- **逐项三选一**：
  - **(a) 补语义**：`acc_complete`/`clean_round_still_valid` 等接真实校验（指纹/签名/哈希）。
  - **(b) 接入 `Apply` 前置**：`forbidden_events`/`protected_commands`/合成 `TransitionID` 白名单接入校验。
  - **(c) 删除声明**：`warn_and_retry` 改名或实现分支；幽灵阶段/旧 `BUG draft` 死代码清理；`guardRegistry` 缩至真实守门并补测试。
- **Task**：`#9` — **RC-06 守卫存根与幽灵 Transition 泛滥**（`transition/guards.go:114` 等 4 文件，逐项 force/structure/删除，成本中）

---

## 7. RC-07 — S7 计划与独立性硬度不足

**包含发现**：`S7-5` (`submit.go:1519`) / `S7-6` (`plan.go:442`) / `S7-7` (`e2e_inventory.go:1`) / `S7-8` (`submit.go:946`) / `S7-9` (`draft.go:184`) / `S7-11` (`submit.go:722`) / `EH-2/EH-3/EH-13`

**RepairContract — RC-07** (`#10`)

1. `non_overlap_boundary` 改 `target+method` 集合精确比对，空 `focus_key` 视为过载拒。
2. `regression_available` 需 `selector/路由/环境` 指纹，非子串。
3. `validateProducerIndependence` 扩大 `kind` 白名单 + 统一独立性载体。
4. 空 `ui_impact` 视为错误（需显式 `none+checklist`）。
5. `ObservationBatch` 全量重算 `readiness`；未识别裸引用视为 `unsatisfied` 拒收。

---

## 8. RC-08 — S8 调查材料完整性占位

**包含发现**：`S8-2~S8-7` / `S8-10` / `EH-5/EH-6`

**RepairContract — RC-08** (`#11`)

1. `blast_radius` 需路径集合非空；`detection_gap` 需 `gap_type+evidence_ref`。
2. `supported` 需 `refuted≥1` 或显式 `no_competing_hypothesis`。
3. `hypothesis` 必须绑定 `assignment_id` 且 `evidence_ref` 非空。
4. `review.investigation` 改数组/`CaseGroup` 支持多 Case。
5. 新鲜度改为新 `evidence_ref` 哈希；`status --all` 无 state 时标 `unknown`。

---

## 9. RC-09 — S9 修复链过度与漂移

**包含发现**：`S9-3` (`impact/analysis.go:126`) / `S9-4` (`repair/runtime.go:90`) / `S9-6` (`repair/model.go:239`) / `S9-7` (`transition/repair_limit.go:0`) / `S9-8` (`repair/runtime.go:113`) / `S9-10~12` + `S9-2` 先红后绿 (`plan_report.go:52`) / `C-4` 双轨

**RepairContract — RC-09** (`#12`)

1. `scope_refs` 为空时视为全量敏感或强制声明前缀；空 `scope_refs` 证据在无关路径变更后自动失效。
2. `runtime.go:90` 增 `authority fingerprint+stale` 通道，基线漂移后修复提交 `stale` 拦截。
3. `required_reverification_ids` 接线或删除；统一收口至 `CheckRepairLimit`（删 `bug_lifecycle.go:355` 内联副本）。
4. `TransitionID` 白名单校验；无自定义 `assertion_ids` 时要求显式 `all`。
5. `ContinuityReason` 需 `evidence_ref`；先红后绿 `repro_evidence_ref` 需有效 `evidence` 且 `kind` 匹配。

---

## 10. RC-10 — 巨事务与指纹/度量冗余（精简）

**包含发现**：`C-1` 17 步 CAS (`submit.go:67` Critical) / `C-3` 指纹 8 类+5 revision / `C-7` 79 case / `C-11` Journal 无界 (`store.go:2597`) / `C-14/C-15` 43 metric 脱节

**RepairContract — RC-10** (`#13`)

1. 拆 `ReviewResult` 事务为 `validate→persist immutable→CAS consume→seal` 三段，分段幂等键与短锁（锁持有 95 分位 <100ms）。
2. 指纹合并为 `state_hash+evidence_hash+baseline_hash` 三元，`revision` 统一为 `state_revision+evidence_generation`。
3. `Journal` 轮转/归档（`generation` 或 10k 条阈值）。
4. `metric` 裁至 ≤10 可行动指标，其余归档为诊断日志；`handoff` 8 步延迟至真实失效出现。

---

## 11. RC-11 — 协议与文档面过度与漂移（精简 + 对齐）

**包含发现**：`C-5` Markdown 双载体 / `C-6` 31 字段 / `C-8` legacy 占位 / `C-9` 8 步 handoff / `C-10` canned claim / `C-13` 重复设防 / `F-2/F-3/F-4/F-5` ghost/漂移

**RepairContract — RC-11** (`#14`)

1. `agent-message` 分层必填（核心 8 字段 + 扩展）。
2. 恢复协议按 Stage 分片。
3. S10 `Markdown` 改为 `manifest` 渲染视图（单一可信源）。
4. 归档 `PTR-BUG` legacy 与 `Phase transitions` 旧条目；`s7_seed` 移除占位。
5. 注册表三方对齐为真值 65/45 并修正文档/`definition`；删除 `ghost` 命令 `verification result submit/init`；修正 `guidance` 路径。

---

## 12. RC-12 — 流畅度回路与口径摩擦（guide + structure）

**包含发现**：`FL-1~FL-16` / `S7-10` / `S7-12`

**RepairContract — RC-12** (`#15`)

1. `s7 status --explain` + `Wave` 就绪投影（一行告知 A 完成度/B 准入/下一动词）。
2. `S8→S9→新 S7` 一键脚手架 `s8 intake --emit case-template`（20 动词→3 动词）。
3. 渐进披露重排（入口仅 3 verb，其余 `--advanced`）；动词收口与 `stopidle` 按 `PlanReportedRef` 分叉。
4. `N/A checklist` 模板；统一 `usage/status/revision` 口径；修复 `investigator` 卡段落。

---

## 13. Task 总表（#4–#15）

| Task | RC | 标题 | 位置 | Severity 峰值 | D5 | 波次 | 验收负向用例 |
|---|---|---|---|---|---|---|
| `#4` | RC-01 | 身份绑定缺失 — Targeted 独立性 | `repair/evidence.go:81` | Critical | force | Wave1 | 原实现者填假 ID 自验应拒 |
| `#5` | RC-02 | 阻塞语义窄化 — `isBlockingBug==P0` | `clean_round.go:259` | Critical | structure→force | Wave1 | P1 `blocking` BUG 应阻 clean |
| `#6` | RC-03 | 冻结基线双层旁路 | `policy/engine.go:781` | Critical | structure | Wave2 | `python3 gen` 改产品应被拦截/DRIFT |
| `#7` | RC-04 | 写屏障休眠与阶段裸窗口 | `policy/engine.go:337` | Critical | force | Wave1 | S8 Edit / S10 Write / targeted 写应拒 |
| `#8` | RC-05 | 覆盖率分母自指 | `manifest.go:452` | High | force+structure | Wave1 | 少列 1 行 / 同人填全 8 区应拒 |
| `#9` | RC-06 | 守卫存根与幽灵 Transition | `transition/guards.go:114` | High | 逐项 | Wave2 | 无 clean 仍建 ACC / `git push` 在 S10 应拒 |
| `#10` | RC-07 | S7 计划与独立性硬度 | `review/plan.go:442` | Medium | structure | Wave2 | 空 focus_key / 子串回归 / 自定义 kind 应拒 |
| `#11` | RC-08 | S8 调查材料占位 | `case_workflow.go:676` | High | structure | Wave2 | 空壳因果 / 单一 supported 无 refuted 应拒 |
| `#12` | RC-09 | S9 修复链过度与漂移 | `repair/model.go:239` | Critical | structure+force | Wave2 | 空 scope 不失效 / 基线漂移应 stale 拦截 |
| `#13` | RC-10 | 巨事务与指纹/度量冗余 | `review/submit.go:67` | Critical | 精简 | Wave3 | 锁持有 <100ms 95 分位；Journal 10k 归档 |
| `#14` | RC-11 | 协议与文档面过度与漂移 | `schema/agent-message…` | High | 精简+对齐 | Wave3 | ghost 命令 `grep` 零命中；注册表三方一致 |
| `#15` | RC-12 | 流畅度回路与口径摩擦 | `cli/action_catalog.go:55` | High | guide | Wave3 | 冷启动 ≤3 动词；S8→S9→新S7 ≤5 动词 |

**执行建议**：按波次并行，Wave1 四簇可立即开工且互不阻塞；每簇落地后以"负向用例"为门槛做 `exact-set` 式验收（少报/多报均拒），再进入下一波次。`RC-10~12` 精简类可与 Wave2 并行预研，但合并前需通过回归（`L1 D5` 三阶评估：已有 `structure` 可覆盖的不升 `force`）。

---

## 附录 A. 正面确认（9 项，保留不修）

`Store CAS` 四步提交 / `CleanRound` 七检查三层复用 / `ChangeImpact` 双 `exact-set` / `RepairContract` 三闸 / `Hook` 十事件 / `S10 manifest` 真实计算 / `auto_trigger` 约束 / `deny` 恢复包多测拼合 / `N/A` 三重封口 — 均为信任锚，修复中不得回退。

## 附录 B. 证据索引

- 原始 8 维 JSON + `_verified2.json`：`docs/audits/.evidence/`
- 本报告：`docs/audits/S7-S10-RCA-and-Repair-Plan-2026-08-27.md`
- 上游审计：`docs/audits/S7-S10-2026-08-27.md` §1–§6 全量 108 条 `file:line` 锚点

---

*— 根因调查完成。12 簇共 99 缺口已全部归入 RepairContract，Task `#4`–`#15` 可直接按波次派发；每簇验收以"负向用例"为硬门，避免"文档有闸、运行无闸"重演。*
