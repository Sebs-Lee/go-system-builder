---
name: specification-planning
description: Use when designing architecture, final UI design packages, contracts, and task decomposition in the planning phase
category: methodology
version: 2.1.0
---
# Specification Planning

## Authority
The locked REQ is the baseline. Design, contracts, and tasks must trace back to it. Runtime authority lives in `docs/loop-definition.json`; stage contracts live in `docs/agent-protocol.md`; the method summary is inlined below.

## Entry Conditions
- The Loop is in `planning` (phases advance design → contracts → tasks via PTR-PLAN-01/02 and TR-002).
- The locked REQ is readable and its fingerprint matches the runtime baseline.

## Required Inputs
| Input | Path / field | Why |
|:---|:---|:--|
| Locked REQ | runtime `bound_req.path` | source of acceptance criteria and scope |
| Existing design | `docs/design/**` | reuse and conflict detection |
| Module current truth | `docs/design/prototypes/<module>/{index.html, stories.md, flows.md, scenario-model.json, cross-matrix.json, cases.json, scenario-coverage.json, fixture-contract.json, *.html}` | current module package and prototype gate input |
| Rules | `docs/rules/*.md` | naming, security, api-design, state-machine constraints |
| Loop Definition | `docs/loop-definition.json` | planning exit transition and executable guards |

## Procedure — dual-track convergence (v2.0.0)

**Step 0 — route by UI impact before doing anything else**: read the bound
REQ's top `UI impact` field. `unknown` → stop, resolve it in the REQ's §D
first (the `ui_impact_resolved` guard blocks PTR-PLAN-01). `none` → skip
the entire S2 package work (steps 1–9) and go straight to contracts
(step 10). `changed` → run the full flow below.

The S2 portion runs as two tracks with free ordering within one agent (not
subagent parallelism), converging twice. "Free ordering" means either track
may start first — inside Convergence 1 there is a hard order: **stories.md
must exist before rules/branches** (every branch's `story_refs` cites an
S-NNN; writing rules first means citing stories that do not exist yet).
Architecture constrains journeys; when a story challenges the architecture,
escalate that trade-off to the ADR human gate — never silently let one side
win.

**Track system (architecture → facts):**
1. Draft the architecture: components, data model, state machines, data
   flow. Record decisions as ADRs (decision/risk/consequence/alternative)
   under `docs/design/decisions/` — the single human gate of S2 is the
   ADR direction sign-off; the N/A list of the AC bridge joins the same
   sign-off package.
2. Land facts partitions from the architecture vocabulary.

**Track user (REQ §A → stories):**
3. Write stories (S-NNN citing REQ-id) from the REQ's §A intent only.
   When evolving an existing module, read its complete current package
   first — new stories must not duplicate or contradict existing S-NNN.
   Stories seed behavior completeness; they do not depend on the
   architecture track.

**Convergence 1 (behavioral completeness):**
4. Draft rules/branches (oracle written with the branch — a branch's
   polarity forces its outcome thinking) by crossing facts × REQ demand
   points (FR) × stories, rejection paths included. `source_refs` cite
   `REQ-<id>/FR-<id>` so the AC bridge can resolve. The hunt's carrier is
   `cross-matrix.json`: every meaningful cell names its covering branch or
   records a no-branch reason — silence is not N/A. Machine floors (enforced
   by `scenario validate` / generate):
   - every declared fact and every story must appear in at least one cell
     (per-fact and per-story floors; the fact×story combinations themselves
     are your hunting judgment, **not** a cartesian-product requirement);
   - a branch cell must be a branch whose rule actually cites the cell's
     `REQ-<id>/FR-<id>` in `source_refs` (the matrix joins the model, it
     does not assert alongside it); `req_ref` may only reference the bound
     REQ;
   - a `no_branch_reason` is a rationale: name the why — at least 8
     characters including a letter (a bare "不需要"/"." is rejected;
     free-word escapes are not endorsed N/A).
5. Immediately run `go run ./cmd/loop-harness scenario bridge --root .` — the
   AC source check (AC→FR→BR) needs no generated outputs; fix gaps now, not
   at close.

**Fixtures:**
6. After branches settle, write `fixture-contract.json` (synthetic data +
   cleanup) — data needs are branch-certain by now.

**Convergence 2 (walkable):**
7. Write flows (F-NNN/PATH-*) and prototype pages (4-field header) as the
   convergence of stories' journeys and branches' behavior;
   `browser_required` branches bind their PATH here.

**Depth self-review, then close:**
8. Before closing, attack your own package in three roles — implementer:
   "which oracle can't I build or can't distinguish a wrong implementation
   by?"; e2e-tester: "which negative case can't I evidence across the
   seven oracle dimensions? (visible, terminal_state, persisted_effects,
   forbidden_side_effects — plus, for negative branches, rejection,
   expected_state, recovery)"; maintainer: "which rule will clash with
   module evolution?" Record the conclusion as one paragraph under a
   `## Depth Self-Review` heading in the ADR package.
9. Close: `go run ./cmd/loop-harness scenario generate --module <module>
   --root .` then `scenario validate --module <module> --root .` — ratio
   gates, reference existence, byte-frozen outputs, cross-matrix
   references, and the full AC bridge (every AC reaches a CASE or carries
   an endorsed N/A: a declared NFR id or an explicit §A4 negative-space
   pointer — §A4 is the REQ's "明确不做" table; free text is rejected as a
   silent removal from the verification denominator).

**S3/S4 steps:**
10. Draft contracts in order: FE-contract → BE-contract → SYNC-contract,
    from the four templates under `docs/contracts/` (CONTRACTS / BE / FE /
    SYNC). Each must link to the REQ source ref and the module
    current-truth package. The CONTRACTS index 需求覆盖矩阵 is the clause
    universe — one `{id} §{n}` cell per clause, and each `§n` must match
    the clause number the target contract itself declares in any of its
    「本合同条款」 columns. On finalization set each contract's top `Status`
    field to `locked` (PTR-PLAN-02 registers only locked contracts), and
    run `go run ./cmd/loop-harness contracts check --root .` — token
    references, clause cells, and fingerprint columns are machine-checked
    there and again at PTR-PLAN-02.
11. Decompose into TASKs: each TASK binds one primary contract, declares
    its Delivered Clauses (§3) and Module Impact (§3.1), has a Closing
    Contract (§7 four assert lines), and obeys single-responsibility.
    Never hand-copy fingerprints or versions — runtime documents[] owns
    them; §2 keeps read order only.
    **拆分纪律（builder 视角）**：一句话说不成交付物、或出现"以及/然后"→ 拆；FE+BE+SYNC 不混进同一任务；类型/schema/迁移是地基，下游任务必须声明对它的依赖。**compact 警示：尽量避免 subagent 中途 compact——builder 丢任务信息是灾难性表现**。宁可多拆一个任务，也不要让 builder 读着读着上下文被压缩；每个任务的 §2 清单只引用它真正需要的条款切片，不是整份文档。规模直觉锚：必读合计 ~30KB / 触碰 ~8 文件 / 改动 ~400 行——超了先想想能不能拆。
12. Run `loop-harness tasks check` before requesting `TR-002`: batch
    completeness, clause coverage against the index, DAG acyclicity, and
    closing contracts are machine-checked there. Request `TR-002` (its
    `planning_complete` + `tasks_checked` guards re-run the same checks)
    only when the self-check is green.
13. If document verification returns `document_fix_required` (`TR-004`),
    repair the affected documents. Re-open an architecture or UI decision
    only when verification evidence shows that the decision itself is
    invalid; otherwise keep rework bounded to the flagged contract or TASK.

## Outputs
- Architecture and ADR records (with the depth self-review paragraph and
  the endorsed-N/A list) under `docs/design/`.
- Locked current module UI/scenario package (when UI impact or behavior is
  changed) with fingerprints for the scenario package (nine files incl.
  cross-matrix.json), HTML prototype, stories, flows, and module spec.
- FE/BE/SYNC contracts with REQ and design traceability.
- TASK batch with Closing Contracts and single-responsibility assignments.
- Current planning evidence required by `TR-002`.

## Exit Conditions
- The planning checkpoint is committed and the Loop transitioned to `document_verification`.

## Stop Conditions
Stop immediately and surface to the human if any of:
- The REQ is ambiguous about a core acceptance criterion.
- A design decision conflicts with an immutable rule and cannot be resolved without REQ clarification.
- A story challenges the architecture in a way that changes the ADR direction (escalate to the human gate).
- An AC cannot reach a CASE and has no endorseable N/A (category + pointer) — surface the criterion, do not free-text it away.
- User story, user flow, and prototype contradict each other in a way that changes product semantics.
- A contract cannot trace to a REQ clause.

## Non-Goals
- Do not implement code (that is the Builder Agent's job).
- Do not verify contracts (that is `document-verification`).
- Do not create the Agent Team (that is `team-planning`).

## Inlined Methodology

Planning is an executable Loop phase machine (design → contracts →
tasks via PTR-PLAN-01/02), not a document-production state machine.
Architecture, optional UI design, contracts, and TASKs are work products
within those phases; they do not each require a runtime transition.
`TR-002 planning_ready` is the only planning exit and evaluates the
current planning package (registered contracts + the complete TASK batch). `TR-004
document_fix_required` returns failed documents to planning for
evidence-bounded rework. The TASK lifecycle remains `candidate ->
reviewed -> locked -> in_progress -> review -> done`; contracts and TASKs
must trace back to the locked REQ. Invalid or guard-failing events do not
change state, execute no side effect, and record a rejected event.
Idempotency uses CAS revision checks; one committed transition per runtime
revision.
