# Canonical BUG: BUG-CX-01

> Status: fixed
> Severity: P1
> Runtime ref: N/A（模板仓库自审——S0-S4 agent 视角复杂度审查，未绑定 runtime）
> Found in review round: complexity-review-1（三路 sub-agent 模拟行走 + 主会话复核）
> Finding evidence: 本文件 §2/§3 所引 file:line（HEAD=9cd52fa 实测）
> Original responsibility: controller/projection（loop-harness 控制面）
> Owner: PM / Architect

## 1. Fingerprinted Specification Chain

Repair read order is BUG -> TASK -> contracts -> REQ -> UI/design -> rules.

| Order | Kind | ID | Path | Version | SHA-256 | Relevant clauses |
|:--|:--|:--|:--|:--|:--|:--|
| 1 | BUG | BUG-CX-01 | `docs/reports/bugs/BUG-CX-01.md` | v1 | n/a（未锁定） | all |
| 2 | design | L3-S0/L3-S1 | `blueprint/L3-S0-requirement-design.md`, `blueprint/L3-S1-bind.md` | v4.0.3/v4.6.1 | n/a | §投影/引导 |
| 3 | rule | agent-protocol | `docs/agent-protocol.md` | current | n/a | #s0 #s1 |

## 2. Observed Contradiction

**症状：全新项目的第一次 SessionStart 给 agent 一条必然失败的指令，且 S0 引导三处互相矛盾。**

| Field | Value |
|:--|:--|
| expected | fresh checkout（无 `.claude/loop-state.json`）时，SessionStart 投影应给出非 BLOCKED 的 S0 起步引导（起草 REQ → req bind） |
| observed | ① `reconcileGuidance`（internal/cli/controller.go:304-315）对 `store.Snapshot()` 失败统一返回 err → `fallbackGuidance`（controller.go:638-660）输出 **BLOCKED=true** + Action=「run loop-harness runtime reconcile --root .」——该命令对不存在的 state 文件必然失败（run.go:922 起 `read runtime: no such file`）。正确路径（`req bind` 自动 init，run.go:219-230）只写在 AGENTS-template.md:20-24，且该段预设 runtime 已存在。② primary_skill 三处分歧：stage list 说 S0 无 skill（agent-protocol.md:84 为 `—`）；#s0 说 requirement-funnel（:183）；投影实际返回 loop-orchestration（run.go:454-459 `projectNext` inactive 分支）。③ S0 投影 objective/missing 把 S1 的 bind 混入 S0（projection.go:23 objective="bind one human-locked requirement"，missing=`locked_req_binding` 全仓无解释），与 #s0 的产出定义（human-locked REQ，bind 归 S1）冲突。④ S1 在投影层不存在（projectionContracts 无 "S1" 键），#s1 done_when 含「cursor = S1」——该状态永远观测不到（bind 后直接进 planning.design） |
| user/data/system impact | 首次真实使用即卡死或被误导读错误 skill；agent 对系统第一印象是"报错自相矛盾"，损耗后续遵从度 |
| reproduction | 全新目录 clone 模板 → 触发 SessionStart hook → 观察 packet 为 BLOCKED + reconcile 指令 → 执行 `runtime reconcile` → 失败 |

## 3. Root Cause Investigation

| Hypothesis | Evidence | Result |
|:--|:--|:--|
| H1: 投影/引导代码只覆盖"runtime 已存在"的主干路径，冷启动边界从未被任何测试行走 | `projectNext` 的 inactive 分支假设 loop-state.json 已存在（inactive 态文件由谁创建？只有 `req bind` 自动 init——但 SessionStart 在 bind 之前就会触发）；`fallbackGuidance` 把「文件不存在」与「文件损坏」混为同一个 BLOCKED 分支；grep 测试：无任何测试以"无 state 文件"为起点驱动 SessionStart | confirmed |
| H2: stage 表、#s0 段、projection 三处由不同轮次修改，无单一权威导致漂移 | agent-protocol stage list 是 v3 期产物；#s0 在 requirement-funnel 落地轮加了 primary_skill；projectNext inactive 分支返回 loop-orchestration 是编排视角（bind 命令投影）——三者各自"局部合理"，没有 cross-check 机制 | confirmed |
| H3: S1 无投影键是设计（S1 是瞬时动作非驻留阶段） | projectionContracts 无 S1；但 #s1 done_when 仍写「cursor = S1」，文案未随设计清理 | confirmed（文案残留，非设计缺陷） |

Accepted root cause: **控制面引导（projection/protocol/fallback）把"runtime 尚未初始化"当作异常态处理，而它恰恰是每个新项目的第一个常态**；同时 stage 引导信息分散在三处（stage list / #sn 段 / projectNext）无单一权威，随版本演进各自漂移。这是 L1-D1（权威外置）在"引导层"的违例——引导信息本身没有单一居所。

## 4. Closing Contract

### 4.1 Repair scope

- `internal/cli/controller.go`（reconcileGuidance/fallbackGuidance：区分 os.IsNotExist 与损坏，前者返回非 BLOCKED 的 S0 起步 packet：起草 REQ + req bind 指路）
- `internal/cli/run.go` projectNext inactive 分支（primary_skill 按是否存在 bindable REQ 分流：无 → requirement-funnel；有 → loop-orchestration）
- `internal/cli/projection.go`（S0 contract 去 S1 化：objective="produce one human-locked REQ"，missing=human_locked_req）
- `docs/agent-protocol.md`（stage list S0 行补 requirement-funnel；#s1 done_when 删「cursor = S1」改可观测事实）
- 新增测试：以空目录为起点的 SessionStart 投影测试（非 BLOCKED、指路 S0）

### 4.2 Forbidden scope

- 不改 `req bind` 的自动 init 语义（已正确）
- 不引入"初始化 runtime 的新命令"（YAGNI——bind 已承载）
- 不动 AGENTS-template 的口头授权链（v4.2.0 已裁决）

### 4.3 Before-fix evidence

本文件 §2 所列 file:line（controller.go:638-660 / run.go:451-459 / projection.go:23 / agent-protocol.md:84,183,196）+ 冷启动复现步骤。

### 4.4 Retest contract

```text
assert SessionStart(empty_root).Blocked == false
assert SessionStart(empty_root).Action 指向 "draft REQ from template / req bind"
assert projectNext(inactive, 无 bindable).PrimarySkill == "requirement-funnel"
assert projectNext(inactive, 有 bindable).PrimarySkill == "loop-orchestration"
assert agent-protocol stage list #s0 primary_skill == #s0 段 == 投影（三处一致）
assert 新增空根投影测试 fails_before_and_passes_after
```

## 5. Acceptance And Repair

| Field | Reference |
|:--|:--|
| BUG acceptance evidence | owner 全部接受（2026-08-17） |
| repair assignment | same batch（本轮修复） |
| Builder activation | same batch（本轮修复） |
| repair fingerprint | repair commit（本轮） |
| impact analysis | same batch（本轮修复） |
| invalidated evidence | n/a |

## 6. Verification

| Verification | Owner | Result | Evidence |
|:--|:--|:--|:--|
| 冷启动投影测试 + 三处一致性检查 | 待派 | pass | TestFreshCheckoutSessionStartIsNotBlocked；S0 contract/inactiveRuntimeState/projection 三处一致 |

## 7. Deduplication And History

Canonical BUG: BUG-CX-01（冷启动/引导权威分裂族）

| Date | Event | Actor | Runtime revision | Evidence |
|:--|:--|:--|:--|:--|
| 2026-08-17 | reported（复杂度审查发现 A1/A2/A3/A5） | 主会话+sub-agent×3 | n/a | 本文件 |
| 2026-08-17 | fixed+verified（修复落地，全量测试/validate/doctor 绿） | 主会话 | n/a | TestFreshCheckoutSessionStartIsNotBlocked；S0 contract/inactiveRuntimeState/projection 三处一致 |
